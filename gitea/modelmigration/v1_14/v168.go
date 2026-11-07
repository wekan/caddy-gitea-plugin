// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_14

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
)

func RecreateUserTableToFixDefaultValues(_ context.Context, _ base.EngineMigration) error {
	return nil
}
