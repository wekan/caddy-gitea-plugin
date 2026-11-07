// Copyright 2022 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package migration

import (
	"io/fs"
	"path"
	"sync"

	"embed"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Copied from modules/migration/schemas by releases/copy-gitea.py.
//
//go:embed all:assets
var bindata embed.FS

var BuiltinAssets = sync.OnceValue(func() fs.FS {
	return embeddedAssets()
})

func openSchema(filename string) (any, error) {
	f, err := BuiltinAssets().Open(path.Base(filename))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return jsonschema.UnmarshalJSON(f)
}

// embeddedAssets is assets/ as the root of a file system.
func embeddedAssets() fs.FS {
	sub, err := fs.Sub(bindata, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}
