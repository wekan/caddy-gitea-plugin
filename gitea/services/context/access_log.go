// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"net/http"
	"strings"
	"text/template"
	"time"
	"unicode"

	user_model "github.com/wekan/caddy-gitea-plugin/gitea/models/user"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/httplib"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/log"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/util"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web/middleware"
)

type accessLoggerTmplData struct {
	Identity       *string
	Start          *time.Time
	ResponseWriter struct {
		Status, Size int
	}
	Ctx       map[string]any
	RequestID *string
}

const keyOfRequestIDInTemplate = ".RequestID"

// According to:
// TraceId: A valid trace identifier is a 16-byte array with at least one non-zero byte
// MD5 output is 16 or 32 bytes: md5-bytes is 16, md5-hex is 32
// SHA1: similar, SHA1-bytes is 20, SHA1-hex is 40.
// UUID is 128-bit, 32 hex chars, 36 ASCII chars with 4 dashes
// So, we accept a Request ID with a maximum character length of 40
const maxRequestIDByteLength = 40

func isSafeRequestID(id string) bool {
	for _, r := range id {
		safe := unicode.IsPrint(r)
		if !safe {
			return false
		}
	}
	return true
}

func parseRequestIDFromRequestHeader(req *http.Request) string {
	requestID := "-"
	for _, key := range setting.Log.RequestIDHeaders {
		if req.Header.Get(key) != "" {
			requestID = req.Header.Get(key)
			break
		}
	}
	if !isSafeRequestID(requestID) {
		return "-"
	}
	if len(requestID) > maxRequestIDByteLength {
		requestID = util.TruncateStringBytes(requestID, maxRequestIDByteLength) + "..."
	}
	return requestID
}

// redactedQueryKeys and redactedHeaders carry credentials: a request with any of
// them is logged with "[REDACTED]" in their place, never their value.
var (
	redactedQueryKeys = []string{"token", "access_token", "private_token", "password"}
	redactedHeaders   = []string{"Authorization", "Cookie", "X-Gitea-Otp", "Proxy-Authorization"}
)

// redactedRequest is req as the access log may show it: a shallow copy whose URL
// and headers have the credentials replaced. req itself is not changed.
func redactedRequest(req *http.Request) *http.Request {
	changed := false
	u := *req.URL
	if q := u.Query(); len(q) > 0 {
		for _, k := range redactedQueryKeys {
			if q.Has(k) {
				q.Set(k, "[REDACTED]")
				changed = true
			}
		}
		if changed {
			u.RawQuery = q.Encode()
		}
	}
	h, cloned := req.Header, false
	for _, k := range redactedHeaders {
		if h.Get(k) != "" {
			if !cloned {
				h, cloned = h.Clone(), true // req's own headers stay as they are
			}
			h.Set(k, "[REDACTED]")
			changed = true
		}
	}
	if !changed {
		return req
	}
	r := new(http.Request)
	*r = *req
	r.URL, r.Header, r.RequestURI = &u, h, u.RequestURI()
	return r
}

type accessLogRecorder struct {
	logger        log.BaseLogger
	logTemplate   *template.Template
	needRequestID bool
}

func (lr *accessLogRecorder) record(start time.Time, respWriter ResponseWriter, req *http.Request) {
	var requestID string
	if lr.needRequestID {
		requestID = parseRequestIDFromRequestHeader(req)
	}

	reqHost := httplib.RemoteHost(req)

	identity := "-"
	data := middleware.GetContextData(req.Context())
	if signedUser, ok := data[middleware.ContextDataKeySignedUser].(*user_model.User); ok {
		identity = signedUser.Name
	}
	buf := bytes.NewBuffer([]byte{})
	tmplData := accessLoggerTmplData{
		Identity: &identity,
		Start:    &start,
		Ctx: map[string]any{
			"RemoteAddr": req.RemoteAddr,
			"RemoteHost": reqHost,
			"Req":        redactedRequest(req),
		},
		RequestID: &requestID,
	}
	tmplData.ResponseWriter.Status = respWriter.WrittenStatus()
	tmplData.ResponseWriter.Size = respWriter.WrittenSize()
	err := lr.logTemplate.Execute(buf, tmplData)
	if err != nil {
		log.Error("Could not execute access logger template: %v", err.Error())
	}

	lr.logger.Log(1, &log.Event{Level: log.INFO}, "%s", buf.String())
}

func newAccessLogRecorder() *accessLogRecorder {
	return &accessLogRecorder{
		logger:        log.GetLogger("access"),
		logTemplate:   template.Must(template.New("log").Parse(setting.Log.AccessLogTemplate)),
		needRequestID: len(setting.Log.RequestIDHeaders) > 0 && strings.Contains(setting.Log.AccessLogTemplate, keyOfRequestIDInTemplate),
	}
}

// AccessLogger returns a middleware to log access logger
func AccessLogger() func(http.Handler) http.Handler {
	recorder := newAccessLogRecorder()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			respWriter := WrapResponseWriter(w)
			next.ServeHTTP(respWriter, req)
			recorder.record(start, respWriter, req)
		})
	}
}
