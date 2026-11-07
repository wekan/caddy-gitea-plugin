// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package private

import (
	"io"
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/json"
	myCtx "github.com/wekan/caddy-gitea-plugin/gitea/services/context"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/migrations"
)

// RestoreRepo restore a repository from data
func RestoreRepo(ctx *myCtx.PrivateContext) {
	bs, err := io.ReadAll(ctx.Req.Body)
	if err != nil {
		ctx.PrivateInternalErrorf("%v", err)
		return
	}
	params := struct {
		RepoDir    string
		OwnerName  string
		RepoName   string
		Units      []string
		Validation bool
	}{}
	if err = json.Unmarshal(bs, &params); err != nil {
		ctx.PrivateInternalErrorf("%v", err)
		return
	}

	if err := migrations.RestoreRepository(
		ctx,
		params.RepoDir,
		params.OwnerName,
		params.RepoName,
		params.Units,
		params.Validation,
	); err != nil {
		ctx.PrivateInternalErrorf("%v", err)
	} else {
		ctx.PlainText(http.StatusOK, "success")
	}
}
