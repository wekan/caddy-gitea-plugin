// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package vagrant

import (
	vagrant_module "github.com/wekan/caddy-gitea-plugin/gitea/modules/packages/vagrant"
	packages_service "github.com/wekan/caddy-gitea-plugin/gitea/services/packages"
)

type Specialization struct {
	packages_service.SpecGetViewPackageVersionData
	packages_service.SpecOnBeforeRemovePackageAll
	packages_service.SpecOnBeforeRemovePackageVersion
	packages_service.SpecRenderUsageManual
}

var _ packages_service.Specialization = (*Specialization)(nil)

func (s Specialization) PublicPropertyNames() []string {
	return []string{vagrant_module.PropertyProvider}
}
