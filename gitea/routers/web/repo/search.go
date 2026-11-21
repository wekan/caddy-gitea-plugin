// Copyright 2017 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/git"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/common"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/codesearch"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

const tplSearch templates.TplName = "repo/search"

// Search render repository search page
func Search(ctx *context.Context) {
	ctx.Data["PageIsViewCode"] = true
	ctx.Data["Repo"] = ctx.Repo.Repository
	engine := codesearch.GitGrep
	if setting.Indexer.RepoIndexerEnabled {
		engine = codesearch.Indexer
	}
	common.SearchCode(ctx, &codesearch.Options{
		Engine: engine, RepoID: ctx.Repo.Repository.ID, GitRepo: ctx.Repo.GitRepo, Ref: git.RefNameFromBranch(ctx.Repo.Repository.DefaultBranch),
	})
	if !ctx.Written() {
		ctx.HTML(http.StatusOK, tplSearch)
	}
}
