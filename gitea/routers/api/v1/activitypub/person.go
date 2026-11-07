// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package activitypub

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

func NotImplemented(ctx *context.APIContext) {
	http.Error(ctx.Resp, "Not implemented", http.StatusNotImplemented)
}
