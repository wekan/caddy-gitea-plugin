// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v28

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"

	"xorm.io/xorm"
)

type pullAutoMerge struct {
	MergedCommitID string `xorm:"VARCHAR(64)"`
}

func (pullAutoMerge) TableName() string {
	return "pull_auto_merge"
}

func AddAutoMergeMergedCommitID(_ context.Context, x base.EngineMigration) error {
	_, err := x.SyncWithOptions(xorm.SyncOptions{
		IgnoreConstrains:  true,
		IgnoreDropIndices: true,
	}, new(pullAutoMerge))
	return err
}
