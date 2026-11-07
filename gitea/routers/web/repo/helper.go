// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/git"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

func HandleGitError(ctx *context.Context, msg string, err error) {
	if git.IsErrNotExist(err) {
		ctx.Data["NotFoundPrompt"] = ctx.Locale.Tr("repo.tree_path_not_found", ctx.Repo.TreePath, ctx.Repo.RefTypeNameSubURL())
		ctx.Data["NotFoundGoBackURL"] = ctx.Repo.RepoLink + "/src/" + ctx.Repo.RefTypeNameSubURL()
		ctx.NotFound(err)
	} else {
		ctx.ServerError(msg, err)
	}
}
