// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v28

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"

	"xorm.io/xorm"
)

func AddMaxParallelToActionRunJob(_ context.Context, x base.EngineMigration) error {
	type ActionRunJob struct {
		MaxParallel int `xorm:"NOT NULL DEFAULT 0"`
	}

	_, err := x.SyncWithOptions(xorm.SyncOptions{
		IgnoreConstrains:  true,
		IgnoreDropIndices: true,
	}, new(ActionRunJob))
	return err
}
