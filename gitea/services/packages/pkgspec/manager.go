// Copyright 2026 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package pkgspec

import (
	packages_model "github.com/wekan/caddy-gitea-plugin/gitea/models/packages"
	packages_service "github.com/wekan/caddy-gitea-plugin/gitea/services/packages"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/alpine"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/arch"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/conan"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/conda"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/debian"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/nuget"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/rpm"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/terraform"
	"github.com/wekan/caddy-gitea-plugin/gitea/services/packages/vagrant"
)

func InitManager() error {
	mgr := packages_service.GetSpecManager()
	mgr.Add(packages_model.TypeAlpine, &alpine.Specialization{})
	mgr.Add(packages_model.TypeArch, &arch.Specialization{})
	mgr.Add(packages_model.TypeConan, &conan.Specialization{})
	mgr.Add(packages_model.TypeConda, &conda.Specialization{})
	mgr.Add(packages_model.TypeDebian, &debian.Specialization{})
	mgr.Add(packages_model.TypeNuGet, &nuget.Specialization{})
	mgr.Add(packages_model.TypeRpm, &rpm.Specialization{})
	mgr.Add(packages_model.TypeTerraformState, &terraform.Specialization{})
	mgr.Add(packages_model.TypeVagrant, &vagrant.Specialization{})
	// TODO: add more in the future, refactor the existing code to use this approach
	return nil
}
