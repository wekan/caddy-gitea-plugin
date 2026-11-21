# gitea/ - the parts of Gitea this plugin builds into Caddy

Copied by `releases/sync-gitea.sh` - do not edit here: change upstream Gitea,
wekan/gitea-patches, or this repository's `dist/gitea/`, and sync again.

| | |
|---|---|
| Upstream | [go-gitea/gitea](https://github.com/go-gitea/gitea) `main`, version 29.0.0, commit [`6f8ebc0b2fb0eb247273f7da4059c5e5999de691`](https://github.com/go-gitea/gitea/commit/6f8ebc0b2fb0eb247273f7da4059c5e5999de691) |
| wekan/gitea-patches | commit `fda5da26ea5a419f6af09dcd31cab58bb05a351c` |
| This repository's patches | `embedded-in-a-host.patch` |
| Resulting Gitea commit | `d3bac4a04b93631cda89db89fef6d91bbb4cb02e` |

Only the packages the plugin needs are here - `go list -deps` of what it imports,
for every operating system Weddy is built for - with Gitea's assets embedded from
`assets/` directories. Gitea is MIT licensed: [LICENSE](LICENSE).
