// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_19

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
)

func AddIndexForAccessToken(_ context.Context, x base.EngineMigration) error {
	type AccessToken struct {
		TokenLastEight string `xorm:"INDEX token_last_eight"`
	}

	return x.Sync(new(AccessToken))
}
