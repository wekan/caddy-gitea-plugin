// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	issues_model "github.com/wekan/caddy-gitea-plugin/gitea/models/issues"
	"github.com/wekan/caddy-gitea-plugin/gitea/models/perm"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/forms"
)

// LockIssue locks an issue. This would limit commenting abilities to
// users with write access to the repo.
func LockIssue(ctx *context.Context) {
	form := web.GetForm[*forms.IssueLockForm](ctx)
	issue := GetActionIssue(ctx, perm.AccessModeWrite)
	if issue == nil {
		return
	}

	if issue.IsLocked {
		ctx.JSONError(ctx.Tr("repo.issues.lock_duplicate"))
		return
	}

	if err := issues_model.LockIssue(ctx, &issues_model.IssueLockOptions{
		Doer:   ctx.Doer,
		Issue:  issue,
		Reason: form.Reason,
	}); err != nil {
		ctx.ServerError("LockIssue", err)
		return
	}

	ctx.JSONRedirect(issue.Link())
}

// UnlockIssue unlocks a previously locked issue.
func UnlockIssue(ctx *context.Context) {
	issue := GetActionIssue(ctx, perm.AccessModeWrite)
	if issue == nil {
		return
	}

	if !issue.IsLocked {
		ctx.JSONError(ctx.Tr("repo.issues.unlock_error"))
		return
	}

	if err := issues_model.UnlockIssue(ctx, &issues_model.IssueLockOptions{
		Doer:  ctx.Doer,
		Issue: issue,
	}); err != nil {
		ctx.ServerError("UnlockIssue", err)
		return
	}

	ctx.JSONRedirect(issue.Link())
}
