// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_11

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
)

func AddCanCreateOrgRepoColumnForTeam(_ context.Context, x base.EngineMigration) error {
	type Team struct {
		CanCreateOrgRepo bool `xorm:"NOT NULL DEFAULT false"`
	}

	return x.Sync(new(Team))
}
