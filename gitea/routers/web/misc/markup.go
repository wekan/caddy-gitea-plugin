// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package misc

import (
	api "github.com/wekan/caddy-gitea-plugin/gitea/modules/structs"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/util"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/common"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

// Markup render markup document to HTML
func Markup(ctx *context.Context) {
	form := web.GetForm[*api.MarkupOption](ctx)
	mode := util.Iif(form.Wiki, "wiki", form.Mode) //nolint:staticcheck // form.Wiki is deprecated
	common.RenderMarkup(ctx.Base, ctx.Repo, mode, form.Text, form.Context, form.FilePath)
}
