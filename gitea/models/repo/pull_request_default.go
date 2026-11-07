// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/models/unit"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/util"
)

func (repo *Repository) GetPullRequestTargetBranch(ctx context.Context) string {
	unitPRConfig := repo.MustGetUnit(ctx, unit.TypePullRequests).PullRequestsConfig()
	return util.IfZero(unitPRConfig.DefaultTargetBranch, repo.DefaultBranch)
}
