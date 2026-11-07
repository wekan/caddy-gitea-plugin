// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_22

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/timeutil"
)

func CreateAuthTokenTable(_ context.Context, x base.EngineMigration) error {
	type AuthToken struct {
		ID          string `xorm:"pk"`
		TokenHash   string
		UserID      int64              `xorm:"INDEX"`
		ExpiresUnix timeutil.TimeStamp `xorm:"INDEX"`
	}

	return x.Sync(new(AuthToken))
}
