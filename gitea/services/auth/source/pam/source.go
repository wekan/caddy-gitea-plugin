// Copyright 2021 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package pam

import (
	"github.com/wekan/caddy-gitea-plugin/gitea/models/auth"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/json"
)

// Source holds configuration for the PAM login source.
type Source struct {
	auth.ConfigBase `json:"-"`

	ServiceName string // pam service (e.g. system-auth)
	EmailDomain string
}

func (source *Source) FromDB(bs []byte) error {
	auth.UnmarshalAuthSourceConfig(bs, source)
	return nil
}

func (source *Source) ToDB() ([]byte, error) {
	return json.Marshal(source)
}

func init() {
	auth.RegisterTypeConfig(auth.PAM, &Source{})
}
