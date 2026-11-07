// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"net/http"
	"time"

	"github.com/wekan/caddy-gitea-plugin/gitea/models/db"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/graceful"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/graceful/releasereopen"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/log"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/private"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/queue"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

// ReloadTemplates reloads all the templates
func ReloadTemplates(ctx *context.PrivateContext) {
	err := templates.ReloadAllTemplates()
	if err != nil {
		ctx.PrivateInternalErrorf("Template error: %v", err)
		return
	}
	ctx.PlainText(http.StatusOK, "success")
}

// FlushQueues flushes all the Queues
func FlushQueues(ctx *context.PrivateContext) {
	opts := web.GetForm[*private.FlushOptions](ctx)
	if opts.NonBlocking {
		// Save the hammer ctx here - as a new one is created each time you call this.
		baseCtx := graceful.GetManager().HammerContext()
		go func() {
			err := queue.GetManager().FlushAll(baseCtx, time.Duration(opts.Timeout))
			if err != nil {
				log.Error("Flushing request timed-out with error: %v", err)
			}
		}()
		ctx.JSON(http.StatusAccepted, private.Response{
			UserMsg: "Flushing",
		})
		return
	}
	err := queue.GetManager().FlushAll(ctx, time.Duration(opts.Timeout))
	if err != nil {
		ctx.PrivateUserErrorf(http.StatusRequestTimeout, "%v", err)
		return
	}
	ctx.PlainText(http.StatusOK, "success")
}

// PauseLogging pauses logging
func PauseLogging(ctx *context.PrivateContext) {
	log.GetManager().PauseAll()
	ctx.PlainText(http.StatusOK, "success")
}

// ResumeLogging resumes logging
func ResumeLogging(ctx *context.PrivateContext) {
	log.GetManager().ResumeAll()
	ctx.PlainText(http.StatusOK, "success")
}

// ReleaseReopenLogging releases and reopens logging files
func ReleaseReopenLogging(ctx *context.PrivateContext) {
	if err := releasereopen.GetManager().ReleaseReopen(); err != nil {
		ctx.PrivateInternalErrorf("Error during release and reopen: %v", err)
		return
	}
	ctx.PlainText(http.StatusOK, "success")
}

// SetLogSQL re-sets database SQL logging
func SetLogSQL(ctx *context.PrivateContext) {
	db.SetLogSQL(ctx, ctx.FormBool("on"))
	ctx.PlainText(http.StatusOK, "success")
}
