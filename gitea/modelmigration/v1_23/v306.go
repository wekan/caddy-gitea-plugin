// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_23

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"

	"xorm.io/xorm"
)

func AddBlockAdminMergeOverrideBranchProtection(_ context.Context, x base.EngineMigration) error {
	type ProtectedBranch struct {
		BlockAdminMergeOverride bool `xorm:"NOT NULL DEFAULT false"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{
		IgnoreConstrains: true,
		IgnoreIndices:    true,
	}, new(ProtectedBranch))
	return err
}
