// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package user

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/common"
	shared_user "github.com/wekan/caddy-gitea-plugin/gitea/routers/web/shared/user"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/codesearch"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

const (
	tplUserCode templates.TplName = "user/code"
)

// CodeSearch render user/organization code search page
func CodeSearch(ctx *context.Context) {
	if !setting.Indexer.RepoIndexerEnabled {
		ctx.Redirect(ctx.ContextUser.HomeLink())
		return
	}
	if _, err := shared_user.RenderUserOrgHeader(ctx); err != nil {
		ctx.ServerErrorAuto(err)
		return
	}

	ctx.Data["IsPackageEnabled"] = setting.Packages.Enabled
	ctx.Data["Title"] = ctx.Tr("explore.code")
	ctx.Data["IsCodePage"] = true

	common.SearchCode(ctx, &codesearch.Options{Engine: codesearch.Indexer, OwnerName: ctx.ContextUser.Name})
	if !ctx.Written() {
		ctx.HTML(http.StatusOK, tplUserCode)
	}
}
