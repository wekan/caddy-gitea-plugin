// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package versioned_migration

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration"
	"github.com/wekan/caddy-gitea-plugin/gitea/models/db"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/globallock"
)

func Migrate(ctx context.Context, x db.EngineMigration) error {
	// only one instance can do the migration at the same time if there are multiple instances
	release, err := globallock.Lock(ctx, "gitea_versioned_migration")
	if err != nil {
		return err
	}
	defer release()

	return modelmigration.Migrate(ctx, x)
}
