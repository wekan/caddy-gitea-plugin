// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package v1_16

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/modelmigration/base"
)

func UseBase32HexForCredIDInWebAuthnCredential(_ context.Context, x base.EngineMigration) error {
	// noop
	return nil
}
