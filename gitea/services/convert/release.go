// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	"context"

	repo_model "github.com/wekan/caddy-gitea-plugin/gitea/models/repo"
	api "github.com/wekan/caddy-gitea-plugin/gitea/modules/structs"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/util"
)

// ToAPIRelease convert a repo_model.Release to api.Release
func ToAPIRelease(ctx context.Context, repo *repo_model.Repository, r *repo_model.Release) *api.Release {
	return &api.Release{
		ID:           r.ID,
		TagName:      r.TagName,
		Target:       r.Target,
		Title:        r.Title,
		Note:         r.Note,
		URL:          r.APIURL(),
		HTMLURL:      r.HTMLURL(),
		TarURL:       r.TarURL(),
		ZipURL:       r.ZipURL(),
		UploadURL:    r.APIUploadURL(),
		IsDraft:      r.IsDraft,
		IsPrerelease: r.IsPrerelease,
		CreatedAt:    r.CreatedUnix.AsTime(),
		PublishedAt:  util.Iif(r.IsDraft, nil, r.PublishedUnix.AsTimePtr()),
		Publisher:    ToAuthor(ctx, r.Publisher, r.OriginalAuthor, nil),
		Attachments:  ToAPIAttachments(ctx, repo, r.Attachments),
	}
}
