// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package audit

import (
	"io"

	audit_model "github.com/wekan/caddy-gitea-plugin/gitea/models/audit"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/json"
)

// WriteEventsAsJSON writes one JSON object per line.
func WriteEventsAsJSON(w io.Writer, events []*audit_model.Event) error {
	for _, event := range events {
		b, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	return nil
}
