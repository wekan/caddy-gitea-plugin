// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package label

import (
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/label"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/forms"
)

func GetLabelEditForm(ctx *context.Context) *forms.CreateLabelForm {
	form := web.GetForm[*forms.CreateLabelForm](ctx)
	if ctx.HasError() {
		ctx.JSONError(ctx.GetErrMsg())
		return nil
	}
	var err error
	form.Color, err = label.NormalizeColor(form.Color)
	if err != nil {
		ctx.JSONError(ctx.Tr("repo.issues.label_color_invalid"))
		return nil
	}
	return form
}
