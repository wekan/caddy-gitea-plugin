// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repo

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/routers/common"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
	issue_service "github.com/wekan/caddy-gitea-plugin/gitea/services/issue"
)

// IssueSuggestions returns a list of issue suggestions
func IssueSuggestions(ctx *context.Context) {
	keyword := ctx.Req.FormValue("q")

	isPull, ok := common.ParseIssueFilterTypeIsPullWithPermission(ctx.Repo.Permission, "")
	if !ok {
		ctx.HTTPError(http.StatusForbidden, "no permission to read issues or pulls")
		return
	}

	suggestions, err := issue_service.GetSuggestion(ctx, ctx.Repo.Repository, isPull, keyword)
	if err != nil {
		ctx.ServerError("GetSuggestion", err)
		return
	}

	ctx.JSON(http.StatusOK, suggestions)
}
