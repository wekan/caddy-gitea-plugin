// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/tailmsg"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

func PerfTrace(ctx *context.Context) {
	monitorTraceCommon(ctx)
	ctx.Data["PageIsAdminMonitorPerfTrace"] = true
	ctx.Data["PerfTraceRecords"] = tailmsg.GetManager().GetTraceRecorder().GetRecords()
	ctx.HTML(http.StatusOK, tplPerfTrace)
}
