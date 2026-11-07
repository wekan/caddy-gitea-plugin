// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

//go:build windows

package private

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/graceful"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

// Restart is not implemented for Windows based servers as they can't fork
func Restart(ctx *context.PrivateContext) {
	ctx.PrivateUserErrorf(http.StatusNotImplemented, "windows servers cannot be gracefully restarted - shutdown and restart manually")
}

// Shutdown causes the server to perform a graceful shutdown
func Shutdown(ctx *context.PrivateContext) {
	graceful.GetManager().DoGracefulShutdown()
	ctx.PlainText(http.StatusOK, "success")
}
