// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_10

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
)

func AddRepoAdminChangeTeamAccessColumnForUser(_ context.Context, x base.EngineMigration) error {
	type User struct {
		RepoAdminChangeTeamAccess bool `xorm:"NOT NULL DEFAULT false"`
	}

	return x.Sync(new(User))
}
