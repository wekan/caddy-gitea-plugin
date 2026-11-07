// Copyright 2014 The Gogs Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package misc

import (
	"io"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	api "github.com/wekan/caddy-gitea-plugin/gitea/modules/structs"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/util"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/common"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

// Markup render markup document to HTML
func Markup(ctx *context.APIContext) {
	// swagger:operation POST /markup miscellaneous renderMarkup
	// ---
	// summary: Render a markup document as HTML
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/MarkupOption"
	// consumes:
	// - application/json
	// produces:
	//     - text/html
	// responses:
	//   "200":
	//     "$ref": "#/responses/MarkupRender"
	//   default:
	//     "$ref": "#/responses/error"

	form := web.GetForm[*api.MarkupOption](ctx)
	mode := util.Iif(form.Wiki, "wiki", form.Mode) //nolint:staticcheck // form.Wiki is deprecated
	common.RenderMarkup(ctx.Base, ctx.Repo, mode, form.Text, form.Context, form.FilePath)
}

// Markdown render markdown document to HTML
func Markdown(ctx *context.APIContext) {
	// swagger:operation POST /markdown miscellaneous renderMarkdown
	// ---
	// summary: Render a markdown document as HTML
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/MarkdownOption"
	// consumes:
	// - application/json
	// produces:
	//     - text/html
	// responses:
	//   "200":
	//     "$ref": "#/responses/MarkdownRender"
	//   default:
	//     "$ref": "#/responses/error"

	form := web.GetForm[*api.MarkdownOption](ctx)
	mode := util.Iif(form.Wiki, "wiki", form.Mode) //nolint:staticcheck // form.Wiki is deprecated
	common.RenderMarkup(ctx.Base, ctx.Repo, mode, form.Text, form.Context, "")
}

// MarkdownRaw render raw markdown HTML
func MarkdownRaw(ctx *context.APIContext) {
	// swagger:operation POST /markdown/raw miscellaneous renderMarkdownRaw
	// ---
	// summary: Render raw markdown as HTML
	// parameters:
	//     - name: body
	//       in: body
	//       description: Request body to render
	//       required: true
	//       schema:
	//         type: string
	// consumes:
	//     - text/plain
	// produces:
	//     - text/html
	// responses:
	//   "200":
	//     "$ref": "#/responses/MarkdownRender"
	//   default:
	//     "$ref": "#/responses/error"
	textBytes, _ := io.ReadAll(io.LimitReader(ctx.Req.Body, setting.UI.MaxDisplayFileSize))
	common.RenderMarkup(ctx.Base, ctx.Repo, "markdown", util.UnsafeBytesToString(textBytes), "", "")
}
