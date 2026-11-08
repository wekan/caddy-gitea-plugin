# gitea/ - the parts of Gitea this plugin builds into Caddy

Copied by `releases/sync-gitea.sh` - do not edit here: change upstream Gitea,
wekan/gitea-patches, or this repository's `dist/gitea/`, and sync again.

| | |
|---|---|
| Upstream | [go-gitea/gitea](https://github.com/go-gitea/gitea) `main`, version 29.0.0, commit [`e09813f47233577e8fca3d7a1890aed4fea195ca`](https://github.com/go-gitea/gitea/commit/e09813f47233577e8fca3d7a1890aed4fea195ca) |
| wekan/gitea-patches | commit `31d800886418a4ccbce38a115a621e768cf7c31a` |
| This repository's patches | `embedded-in-a-host.patch`, `security-hardening.patch` |
| Resulting Gitea commit | `3e64b3593f95b4baa99be3d908873cc91b31843f` |

Only the packages the plugin needs are here - `go list -deps` of what it imports,
for every operating system Weddy is built for - with Gitea's assets embedded from
`assets/` directories. Gitea is MIT licensed: [LICENSE](LICENSE).
