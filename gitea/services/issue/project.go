// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package issue

import (
	"context"

	issues_model "github.com/wekan/caddy-gitea-plugin/gitea/models/issues"
	user_model "github.com/wekan/caddy-gitea-plugin/gitea/models/user"
	notify_service "github.com/wekan/caddy-gitea-plugin/gitea/services/notify"
)

func AssignOrRemoveProjects(ctx context.Context, issue *issues_model.Issue, doer *user_model.User, newProjectIDs []int64) error {
	added, removed, err := issues_model.IssueAssignOrRemoveProject(ctx, issue, doer, newProjectIDs)
	if err != nil {
		return err
	}
	notify_service.IssueChangeProjects(ctx, doer, issue, added, removed)
	return nil
}
