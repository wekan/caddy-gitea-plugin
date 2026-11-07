// Copyright 2014 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"strconv"

	"github.com/wekan/caddy-gitea-plugin/gitea/models/db"
	repo_model "github.com/wekan/caddy-gitea-plugin/gitea/models/repo"
	user_model "github.com/wekan/caddy-gitea-plugin/gitea/models/user"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/log"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/optional"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/sitemap"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/structs"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/web/feed"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/web/user"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/context"
)

const tplHome templates.TplName = "home"

func Home(ctx *context.Context) {
	if _, feedType := feed.GetFeedType(ctx.Req.URL.Path, ctx.Req); feedType != "" {
		ctx.Resp.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}
	if ctx.IsSigned {
		user.Dashboard(ctx)
		return
	} else if setting.LandingPageURL != setting.LandingPageHome {
		// Check non-logged users landing page
		ctx.Redirect(setting.AppSubURL + string(setting.LandingPageURL))
		return
	}

	// Check auto-login.
	if ctx.GetSiteCookie(setting.CookieRememberName) != "" {
		ctx.Redirect(setting.AppSubURL + "/user/login")
		return
	}

	ctx.Data["PageIsHome"] = true
	ctx.Data["IsRepoIndexerEnabled"] = setting.Indexer.RepoIndexerEnabled
	ctx.HTML(http.StatusOK, tplHome)
}

// HomeSitemap renders the main sitemap
func HomeSitemap(ctx *context.Context) {
	m := sitemap.NewSitemapIndex()
	if !setting.Service.Explore.DisableUsersPage {
		_, cnt, err := user_model.SearchUsers(ctx, user_model.SearchUserOptions{
			Types:       []user_model.UserType{user_model.UserTypeIndividual},
			ListOptions: db.ListOptions{PageSize: 1},
			IsActive:    optional.Some(true),
			Visible:     []structs.VisibleType{structs.VisibleTypePublic},
		})
		if err != nil {
			ctx.ServerError("SearchUsers", err)
			return
		}
		count := int(cnt)
		idx := 1
		for i := 0; i < count; i += setting.UI.SitemapPagingNum {
			m.Add(sitemap.URL{URL: setting.AppURL + "explore/users/sitemap-" + strconv.Itoa(idx) + ".xml"})
			idx++
		}
	}

	_, cnt, err := repo_model.SearchRepository(ctx, repo_model.SearchRepoOptions{
		ListOptions: db.ListOptions{
			PageSize: 1,
		},
		Actor:     ctx.Doer,
		AllPublic: true,
	})
	if err != nil {
		ctx.ServerError("SearchRepository", err)
		return
	}
	count := int(cnt)
	idx := 1
	for i := 0; i < count; i += setting.UI.SitemapPagingNum {
		m.Add(sitemap.URL{URL: setting.AppURL + "explore/repos/sitemap-" + strconv.Itoa(idx) + ".xml"})
		idx++
	}

	ctx.Resp.Header().Set("Content-Type", "text/xml")
	if _, err := m.WriteTo(ctx.Resp); err != nil {
		log.Error("Failed writing sitemap: %v", err)
	}
}
