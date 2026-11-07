// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"context"
	"fmt"

	git_model "github.com/wekan/caddy-gitea-plugin/gitea/models/git"
	repo_model "github.com/wekan/caddy-gitea-plugin/gitea/models/repo"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/git"
)

// UpdateRepoSize updates the repository size, calculating it using getDirectorySize
func UpdateRepoSize(ctx context.Context, repo *repo_model.Repository) error {
	size, err := git.CalcRepositorySize(repo)
	if err != nil {
		return fmt.Errorf("updateSize: %w", err)
	}

	lfsSize, err := git_model.GetRepoLFSSize(ctx, repo.ID)
	if err != nil {
		return fmt.Errorf("updateSize: GetLFSMetaObjects: %w", err)
	}

	return repo_model.UpdateRepoSize(ctx, repo.ID, size, lfsSize)
}
