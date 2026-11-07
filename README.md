# Caddy plugin: Gitea

A [Caddy](https://caddyserver.com) plugin that runs [Gitea](https://about.gitea.com),
the version control server, **inside Caddy**: one Caddy binary is the web server and
Gitea, with no second process and no port of its own. Caddy does TLS, HTTP/2 and
HTTP/3 in front of it. Part of [Weddy](https://github.com/wekan/weddy), Caddy with the
Gitea and [WeKan](https://github.com/wekan/caddy-wekan-plugin) plugins.

```caddyfile
{
	gitea {
		work_path /var/lib/gitea
	}
}

git.example.org {
	gitea
}
```

```sh
xcaddy build --with github.com/wekan/caddy-gitea-plugin
```

## How: a patches repository with the needed parts of Gitea in it

Like [wekan/gitea-patches](https://github.com/wekan/gitea-patches), this repository
follows upstream Gitea with patches - but because a Caddy plugin is a Go module that
`go` and xcaddy fetch from its repository, the Gitea code it builds is **committed
here**, in [`gitea/`](gitea/SOURCE.md). [`releases/sync-gitea.sh`](releases/sync-gitea.sh)
makes it:

1. upstream Gitea, with wekan/gitea-patches - its complete Finnish translation and its
   patches - applied as commits by gitea-patches' own `apply-patches.sh`;
2. this repository's [`dist/gitea/`](dist/gitea) patches on top, as commits;
3. Gitea's frontend built;
4. [`releases/copy-gitea.py`](releases/copy-gitea.py) copies **only the needed parts**
   into `gitea/`: the packages this plugin's code imports and everything they import,
   computed with `go list -deps` for every operating system Weddy is built for - no
   tests, `cmd/`, frontend sources, docs or tools - with imports moved into this module,
   and the templates, translations and built frontend embedded from `assets/`
   directories, so a plain `go build` with CGO off and no build tags has all of Gitea's
   web server. `gitea/SOURCE.md` records the exact commits.

`gitea/` is generated: change Gitea upstream, in gitea-patches or in `dist/gitea/`,
and sync again.

| Patch | Why |
|-------|-----|
| [`embedded-in-a-host`](dist/gitea/embedded-in-a-host.md) | Inside Caddy, Gitea must not handle Caddy's signals - a SIGHUP would make it "restart" by re-executing Caddy's binary - nor shut down the HTTP server after its install page, which is Caddy's |

## Modules

| Module | What it does |
|--------|--------------|
| `gitea` (app) | starts Gitea in Caddy's process from its app.ini, in the background - its database migrations do not hold up Caddy. Gitea's state is process-wide, so it starts once, keeps running across Caddy configuration reloads, and stops - gracefully, within `stop_timeout` - when Caddy exits |
| `http.handlers.gitea` | serves a site from it: Gitea answers itself, including its 404s; `503` with `Retry-After` while Gitea starts. Before Gitea is installed, it serves Gitea's install page, and switches to Gitea itself when the installation is done |

| Option | Meaning | Default |
|--------|---------|---------|
| `work_path` | Gitea's WORK_PATH | Gitea's default |
| `custom_path` | Gitea's custom directory | `<work_path>/custom` |
| `config` | Gitea's app.ini | `<custom_path>/conf/app.ini` |
| `stop_timeout` | how long Gitea gets to stop when Caddy exits | `30s` |

The Gitea settings are Gitea's own, in app.ini - `ROOT_URL` should be the site's
address. Changing `work_path`, `custom_path` or `config` needs a restart of Caddy, not
only a reload. Gitea maintains `~/.ssh/authorized_keys` of the user it runs as - the
user Caddy runs as - unless app.ini says `SSH_CREATE_AUTHORIZED_KEYS_FILE = false` or
another `SSH_ROOT_PATH`.

## For other plugins

[`supervisor`](supervisor/supervisor.go) runs a program as a supervised child process
and [`proxy`](proxy/proxy.go) serves a site from it - for programs that stay separate
processes. [caddy-wekan-plugin](https://github.com/wekan/caddy-wekan-plugin) runs WeKan
with them.

## Releases

[Releases](https://github.com/wekan/caddy-gitea-plugin/releases) are tags of this
module, `v0.<YYYYMMDD>.<HHMM>` - `xcaddy build --with github.com/wekan/caddy-gitea-plugin@<tag>`.
Go accepts only v0 and v1 versions for this module path, so Gitea's version is in each
release's title and notes rather than in its tag. The
source attached as `caddy-gitea-plugin-source-with-history.zip` (the repository with its
git history) and `caddy-gitea-plugin-source-without-history.zip`.

## Tests

```sh
CGO_ENABLED=0 go test ./...
```

They run a real Caddy with the embedded Gitea on a new SQLite database: `503` while it
starts, then its login page - in Finnish for a Finnish browser - its embedded assets
and its own 404; and the supervisor with real child processes.

Gitea's own tests are not in `gitea/`, so `go get` of the plugin does not download
them. [`releases/copy-gitea-tests.py`](releases/copy-gitea-tests.py) copies them -
with the packages only they need and their fixtures - into a COPY of this
repository, from the Gitea source `gitea/` was made from; Weddy's
`tests/gitea/run.sh` makes that copy and runs them, CGO off and no build tags,
the way Weddy builds Gitea in.

## License

This repository's own files are MIT - see [LICENSE](LICENSE). Gitea is MIT too:
[gitea/LICENSE](gitea/LICENSE).

Maintainer and contributor rules are in
[WeKan's CLAUDE.md and AGENTS.md](https://github.com/wekan/wekan), which cover every
repository WeKan clones into its `.tools/` directory.
