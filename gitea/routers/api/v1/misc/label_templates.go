// Copyright 2023 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package misc

import (
	"net/http"

	repo_module "github.com/wekan/caddy-gitea-plugin/gitea/modules/repository"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/util"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/convert"
)

// Shows a list of all Label templates
func ListLabelTemplates(ctx *context.APIContext) {
	// swagger:operation GET /label/templates miscellaneous listLabelTemplates
	// ---
	// summary: Returns a list of all label templates
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/LabelTemplateList"
	result := make([]string, len(repo_module.LabelTemplateFiles))
	for i := range repo_module.LabelTemplateFiles {
		result[i] = repo_module.LabelTemplateFiles[i].DisplayName
	}

	ctx.JSON(http.StatusOK, result)
}

// Shows all labels in a template
func GetLabelTemplate(ctx *context.APIContext) {
	// swagger:operation GET /label/templates/{name} miscellaneous getLabelTemplateInfo
	// ---
	// summary: Returns all labels in a template
	// produces:
	// - application/json
	// parameters:
	// - name: name
	//   in: path
	//   description: name of the template
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/LabelTemplateInfo"
	//   default:
	//     "$ref": "#/responses/error"
	name := util.PathJoinRelX(ctx.PathParam("name"))

	labels, err := repo_module.LoadTemplateLabelsByDisplayName(name)
	if err != nil {
		ctx.APIErrorNotFound()
		return
	}

	ctx.JSON(http.StatusOK, convert.ToLabelTemplateList(labels))
}
