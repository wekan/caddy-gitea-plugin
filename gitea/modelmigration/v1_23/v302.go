// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_23

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/timeutil"

	"xorm.io/xorm"
)

func AddIndexToActionTaskStoppedLogExpired(_ context.Context, x base.EngineMigration) error {
	type ActionTask struct {
		Stopped    timeutil.TimeStamp `xorm:"index(stopped_log_expired)"`
		LogExpired bool               `xorm:"index(stopped_log_expired)"`
	}
	_, err := x.SyncWithOptions(xorm.SyncOptions{
		IgnoreDropIndices: true,
	}, new(ActionTask))
	return err
}
