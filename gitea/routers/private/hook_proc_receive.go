// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"net/http"
	"slices"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/git"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/private"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/agit"
	gitea_context "github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

// HookProcReceive proc-receive hook - only handles agit Proc-Receive requests at present
func HookProcReceive(ctx *gitea_context.PrivateContext) {
	opts := web.GetForm[*private.HookOptions](ctx)
	if !loadContextDoerPermission(ctx, opts.UserID, opts.UserExtDoerData) ||
		(slices.ContainsFunc(opts.RefFullNames, git.RefName.IsFor) && !assertCanCreatePullRequest(ctx, opts)) {
		return
	}

	results, err := agit.ProcReceive(ctx, ctx.Repo.Repository, ctx.Repo.GitRepo, ctx.Repo.Permission, &agit.ProcReceiveOptions{
		OldCommitIDs:   opts.OldCommitIDs,
		NewCommitIDs:   opts.NewCommitIDs,
		RefFullNames:   opts.RefFullNames,
		GitPushOptions: opts.GitPushOptions,
		Doer:           ctx.Doer,
	})
	if err != nil {
		ctx.PrivateUserErrorAuto(err)
		return
	}

	ctx.JSON(http.StatusOK, private.HookProcReceiveResult{
		Results: results,
	})
}
