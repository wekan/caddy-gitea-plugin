// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_14

import (
	"context"
	"fmt"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
)

func AddDismissedReviewColumn(_ context.Context, x base.EngineMigration) error {
	type Review struct {
		Dismissed bool `xorm:"NOT NULL DEFAULT false"`
	}

	if err := x.Sync(new(Review)); err != nil {
		return fmt.Errorf("Sync: %w", err)
	}
	return nil
}
