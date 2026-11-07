// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package options

import (
	"sync"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/assetfs"

	"embed"
	"io/fs"
)

// Copied from options by releases/copy-gitea.py.
//
//go:embed all:assets
var bindata embed.FS

var BuiltinAssets = sync.OnceValue(func() *assetfs.Layer {
	return assetfs.Bindata("builtin(bindata)", embeddedAssets())
})

// embeddedAssets is assets/ as the root of a file system.
func embeddedAssets() fs.FS {
	sub, err := fs.Sub(bindata, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}
