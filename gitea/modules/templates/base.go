// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package templates

import (
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/assetfs"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
)

func AssetFS() *assetfs.LayeredFS {
	return assetfs.Layered(CustomAssets(), BuiltinAssets())
}

func CustomAssets() *assetfs.Layer {
	return assetfs.Local("custom", setting.CustomPath, "templates")
}
