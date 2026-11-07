// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	audit_model "github.com/wekan/caddy-gitea-plugin/gitea/models/audit"
	shared_audit "github.com/wekan/caddy-gitea-plugin/gitea/routers/web/shared/audit"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

func ViewAuditLogs(ctx *context.Context) {
	shared_audit.View(ctx, shared_audit.ViewOptions{
		Template:  "org/settings/audit_logs",
		ScopeType: audit_model.ScopeOrganization,
		ScopeID:   ctx.Org.Organization.ID,
		PageData: map[string]any{
			"PageIsOrgSettings":   true,
			"PageIsSettingsAudit": true,
		},
	})
}
