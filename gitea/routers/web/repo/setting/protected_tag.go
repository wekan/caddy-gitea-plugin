// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package setting

import (
	"fmt"
	"net/http"
	"strings"

	audit_model "github.com/wekan/caddy-gitea-plugin/gitea/models/audit"
	git_model "github.com/wekan/caddy-gitea-plugin/gitea/models/git"
	"github.com/wekan/caddy-gitea-plugin/gitea/models/organization"
	"github.com/wekan/caddy-gitea-plugin/gitea/models/perm"
	access_model "github.com/wekan/caddy-gitea-plugin/gitea/models/perm/access"
	"github.com/wekan/caddy-gitea-plugin/gitea/models/unit"
	user_model "github.com/wekan/caddy-gitea-plugin/gitea/models/user"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/base"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/web"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/audit"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/forms"
)

const (
	tplTags templates.TplName = "repo/settings/tags"
)

// Tags render the page to protect tags
func ProtectedTags(ctx *context.Context) {
	if setTagsContext(ctx) != nil {
		return
	}

	ctx.HTML(http.StatusOK, tplTags)
}

// NewProtectedTagPost handles creation of a protect tag
func NewProtectedTagPost(ctx *context.Context) {
	if setTagsContext(ctx) != nil {
		return
	}

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, tplTags)
		return
	}

	repo := ctx.Repo.Repository
	form := web.GetForm[*forms.ProtectTagForm](ctx)

	pt := &git_model.ProtectedTag{
		RepoID:      repo.ID,
		NamePattern: strings.TrimSpace(form.NamePattern),
	}

	if strings.TrimSpace(form.AllowlistUsers) != "" {
		pt.AllowlistUserIDs, _ = base.StringsToInt64s(strings.Split(form.AllowlistUsers, ","))
	}
	if strings.TrimSpace(form.AllowlistTeams) != "" {
		pt.AllowlistTeamIDs, _ = base.StringsToInt64s(strings.Split(form.AllowlistTeams, ","))
	}

	if err := git_model.InsertProtectedTag(ctx, pt); err != nil {
		ctx.ServerError("InsertProtectedTag", err)
		return
	}

	audit.Record(ctx, audit_model.RepositoryTagProtectionAdd, ctx.Repo.Repository, "pattern", pt.NamePattern)

	ctx.Flash.Success(ctx.Tr("repo.settings.update_settings_success"))
	ctx.Redirect(setting.AppSubURL + ctx.Req.URL.EscapedPath())
}

// EditProtectedTag render the page to edit a protect tag
func EditProtectedTag(ctx *context.Context) {
	if setTagsContext(ctx) != nil {
		return
	}

	ctx.Data["PageIsEditProtectedTag"] = true

	pt := selectProtectedTagByContext(ctx)
	if pt == nil {
		return
	}

	ctx.Data["name_pattern"] = pt.NamePattern
	ctx.Data["allowlist_users"] = strings.Join(base.Int64sToStrings(pt.AllowlistUserIDs), ",")
	ctx.Data["allowlist_teams"] = strings.Join(base.Int64sToStrings(pt.AllowlistTeamIDs), ",")

	ctx.HTML(http.StatusOK, tplTags)
}

// EditProtectedTagPost handles creation of a protect tag
func EditProtectedTagPost(ctx *context.Context) {
	if setTagsContext(ctx) != nil {
		return
	}

	ctx.Data["PageIsEditProtectedTag"] = true

	if ctx.HasError() {
		ctx.HTML(http.StatusOK, tplTags)
		return
	}

	pt := selectProtectedTagByContext(ctx)
	if pt == nil {
		return
	}

	form := web.GetForm[*forms.ProtectTagForm](ctx)

	pt.NamePattern = strings.TrimSpace(form.NamePattern)
	pt.AllowlistUserIDs, _ = base.StringsToInt64s(strings.Split(form.AllowlistUsers, ","))
	pt.AllowlistTeamIDs, _ = base.StringsToInt64s(strings.Split(form.AllowlistTeams, ","))

	if err := git_model.UpdateProtectedTag(ctx, pt); err != nil {
		ctx.ServerError("UpdateProtectedTag", err)
		return
	}

	audit.Record(ctx, audit_model.RepositoryTagProtectionUpdate, ctx.Repo.Repository, "pattern", pt.NamePattern)

	ctx.Flash.Success(ctx.Tr("repo.settings.update_settings_success"))
	ctx.Redirect(ctx.Repo.Repository.Link() + "/settings/tags")
}

// DeleteProtectedTagPost handles deletion of a protected tag
func DeleteProtectedTagPost(ctx *context.Context) {
	pt := selectProtectedTagByContext(ctx)
	if pt == nil {
		return
	}

	if err := git_model.DeleteProtectedTag(ctx, pt); err != nil {
		ctx.ServerError("DeleteProtectedTag", err)
		return
	}

	audit.Record(ctx, audit_model.RepositoryTagProtectionRemove, ctx.Repo.Repository, "pattern", pt.NamePattern)

	ctx.Flash.Success(ctx.Tr("repo.settings.update_settings_success"))
	ctx.Redirect(ctx.Repo.Repository.Link() + "/settings/tags")
}

func setTagsContext(ctx *context.Context) error {
	ctx.Data["Title"] = ctx.Tr("repo.settings.tags")
	ctx.Data["PageIsSettingsTags"] = true

	protectedTags, err := git_model.GetProtectedTags(ctx, ctx.Repo.Repository.ID)
	if err != nil {
		ctx.ServerError("GetProtectedTags", err)
		return err
	}
	ctx.Data["ProtectedTags"] = protectedTags

	repoUsers, err := access_model.GetUsersWithAnyUnitAccess(ctx, ctx.Repo.Repository, perm.AccessModeRead, unit.TypeCode, unit.TypePullRequests)
	if err != nil {
		ctx.ServerError("GetUsersWithUnitAccess", err)
		return err
	}
	users := []*user_model.User{user_model.NewActionsUser(), user_model.NewDeployKeyUser()}
	ctx.Data["PushUsers"] = append(users, repoUsers...)

	if ctx.Repo.Owner.IsOrganization() {
		teams, err := organization.GetTeamsWithAccessToAnyRepoUnit(ctx, ctx.Repo.Owner.ID, ctx.Repo.Repository.ID, perm.AccessModeRead, unit.TypeCode, unit.TypePullRequests)
		if err != nil {
			ctx.ServerError("Repo.Owner.TeamsWithAccessToRepo", err)
			return err
		}
		ctx.Data["PushTeams"] = teams
	}

	return nil
}

func selectProtectedTagByContext(ctx *context.Context) *git_model.ProtectedTag {
	id := ctx.FormInt64("id")
	if id == 0 {
		id = ctx.PathParamInt64("id")
	}

	tag, err := git_model.GetProtectedTagByID(ctx, id)
	if err != nil {
		ctx.ServerError("GetProtectedTagByID", err)
		return nil
	}

	if tag != nil && tag.RepoID == ctx.Repo.Repository.ID {
		return tag
	}

	ctx.NotFound(fmt.Errorf("ProtectedTag[%v] not associated to repository %v", id, ctx.Repo.Repository))

	return nil
}
