// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v29

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
)

func DropUserOpenIDTable(_ctx context.Context, x base.EngineMigration) error {
	return x.DropTables("user_open_id")
}
