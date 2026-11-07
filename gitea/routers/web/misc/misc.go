// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package misc

import (
	"net/http"
	"path"
	"strconv"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/httpcache"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/log"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/util"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web/middleware"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/agit"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

func SSHInfo(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("content-type", "text/json;charset=UTF-8")
	_, err := rw.Write([]byte(agit.SshInfoJson))
	if err != nil {
		log.Error("fail to write result: err: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func DummyOK(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func RobotsTxt(w http.ResponseWriter, req *http.Request) {
	robotsTxt := util.FilePathJoinAbs(setting.CustomPath, "public/robots.txt")
	if ok, _ := util.IsExist(robotsTxt); !ok {
		robotsTxt = util.FilePathJoinAbs(setting.CustomPath, "robots.txt") // the legacy "robots.txt"
	}
	httpcache.SetCacheControlInHeader(w.Header(), httpcache.CacheControlForPublicStatic())
	http.ServeFile(w, req, robotsTxt)
}

func StaticRedirect(target string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, path.Join(setting.StaticURLPrefix, target), http.StatusMovedPermanently)
	}
}

func LocationRedirect(target string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, target, http.StatusSeeOther)
	}
}

func WebBannerDismiss(ctx *context.Context) {
	_, rev, _ := setting.Config().Instance.WebBanner.ValueRevision(ctx)
	middleware.SetSiteCookie(ctx.Resp, middleware.CookieWebBannerDismissed, strconv.Itoa(rev), 48*3600)
	ctx.JSONOK()
}
