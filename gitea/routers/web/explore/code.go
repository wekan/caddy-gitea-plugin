// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package explore

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/common"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/codesearch"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

const (
	// tplExploreCode explore code page template
	tplExploreCode templates.TplName = "explore/code"
)

// Code render explore code page
func Code(ctx *context.Context) {
	if !setting.Indexer.RepoIndexerEnabled || setting.Service.Explore.DisableCodePage {
		ctx.Redirect(setting.AppSubURL + "/explore")
		return
	}

	ctx.Data["UsersPageIsDisabled"] = setting.Service.Explore.DisableUsersPage
	ctx.Data["OrganizationsPageIsDisabled"] = setting.Service.Explore.DisableOrganizationsPage
	ctx.Data["IsRepoIndexerEnabled"] = setting.Indexer.RepoIndexerEnabled
	ctx.Data["Title"] = ctx.Tr("explore_title")
	ctx.Data["PageIsExplore"] = true
	ctx.Data["PageIsExploreCode"] = true
	ctx.Data["PageIsViewCode"] = true

	common.SearchCode(ctx, &codesearch.Options{Engine: codesearch.Indexer})
	if !ctx.Written() {
		ctx.HTML(http.StatusOK, tplExploreCode)
	}
}
