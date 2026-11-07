// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"errors"
	"net/http"

	issues_model "github.com/wekan/caddy-gitea-plugin/gitea/models/issues"
	user_model "github.com/wekan/caddy-gitea-plugin/gitea/models/user"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/private"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/agit"
	gitea_context "github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

// HookProcReceive proc-receive hook - only handles agit Proc-Receive requests at present
func HookProcReceive(ctx *gitea_context.PrivateContext) {
	opts := web.GetForm[*private.HookOptions](ctx)
	if !loadContextDoerPermission(ctx, opts.UserID, opts.UserExtDoerData) {
		return
	}

	results, err := agit.ProcReceive(ctx, ctx.Repo.Repository, ctx.Repo.GitRepo, &agit.ProcReceiveOptions{
		OldCommitIDs:   opts.OldCommitIDs,
		NewCommitIDs:   opts.NewCommitIDs,
		RefFullNames:   opts.RefFullNames,
		GitPushOptions: opts.GitPushOptions,
		Doer:           ctx.Doer,
	})
	if err != nil {
		if errors.Is(err, issues_model.ErrMustCollaborator) {
			ctx.PrivateUserErrorf(http.StatusUnauthorized, "You must be a collaborator to create pull request.")
		} else if errors.Is(err, user_model.ErrBlockedUser) {
			ctx.PrivateUserErrorf(http.StatusUnauthorized, "Cannot create pull request because you are blocked by the repository owner.")
		} else {
			ctx.PrivateInternalErrorf("agit.ProcReceive failed: %v", err)
		}

		return
	}

	ctx.JSON(http.StatusOK, private.HookProcReceiveResult{
		Results: results,
	})
}
