// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package files

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/models/asymkey"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/git"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/structs"
	asymkey_service "github.com/wekan/caddy-gitea-plugin/gitea/services/asymkey"
)

// GetPayloadCommitVerification returns the verification information of a commit
func GetPayloadCommitVerification(ctx context.Context, commit *git.Commit) *structs.PayloadCommitVerification {
	verification := &structs.PayloadCommitVerification{}
	commitVerification := asymkey_service.ParseCommitWithSignature(ctx, commit)
	if commit.Signature != nil {
		verification.Signature = commit.Signature.Signature
		verification.Payload = commit.Signature.Payload
	}
	if commitVerification.SigningUser != nil {
		verification.Signer = &structs.PayloadUser{
			Name:  commitVerification.SigningUser.Name,
			Email: commitVerification.SigningUser.Email,
		}
	}
	verification.Verified = commitVerification.Verified
	verification.Reason = commitVerification.Reason
	if verification.Reason == "" && !verification.Verified {
		verification.Reason = asymkey.VerificationReasonCommitNotSigned
	}
	return verification
}
