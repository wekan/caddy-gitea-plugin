// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package explore

import (
	"net/http"

	"github.com/wekan/caddy-gitea-plugin/gitea/models/db"
	repo_model "github.com/wekan/caddy-gitea-plugin/gitea/models/repo"
	api "github.com/wekan/caddy-gitea-plugin/gitea/modules/structs"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/convert"
)

// TopicSearch search for creating topic
func TopicSearch(ctx *context.Context) {
	opts := &repo_model.FindTopicOptions{
		Keyword: ctx.FormString("q"),
		ListOptions: db.ListOptions{
			Page:     ctx.FormInt("page"),
			PageSize: convert.ToCorrectPageSize(ctx.FormInt("limit")),
		},
	}

	topics, total, err := db.FindAndCount[repo_model.Topic](ctx, opts)
	if err != nil {
		ctx.HTTPError(http.StatusInternalServerError)
		return
	}

	topicResponses := make([]*api.TopicResponse, len(topics))
	for i, topic := range topics {
		topicResponses[i] = convert.ToTopicResponse(topic)
	}

	ctx.SetTotalCountHeader(total)
	ctx.JSON(http.StatusOK, map[string]any{
		"topics": topicResponses,
	})
}
