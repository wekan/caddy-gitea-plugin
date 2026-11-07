// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package session

import (
	"context"

	"github.com/wekan/caddy-gitea-plugin/gitea/models/auth"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/log"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/timeutil"
)

type dbBackend struct {
	maxLifetime int64
}

func dbContext() context.Context {
	return context.Background()
}

func (b *dbBackend) load(sid string) ([]byte, error) {
	sess, exist, err := auth.GetSession(dbContext(), sid)
	if err != nil || !exist || sess.LastAccessTime.Add(b.maxLifetime) <= timeutil.TimeStampNow() {
		return nil, err
	}
	return sess.Data, auth.UpdateSessionLastAccessTime(dbContext(), sid)
}

func (b *dbBackend) save(sid string, data []byte, create bool) error {
	return auth.UpdateSession(dbContext(), sid, data, create)
}

func (b *dbBackend) destroy(sid string) error {
	return auth.DestroySession(dbContext(), sid)
}

func (b *dbBackend) gc() {
	if err := auth.CleanupSessions(dbContext(), b.maxLifetime); err != nil {
		log.Error("Unable to garbage collect sessions: %v", err)
	}
}
