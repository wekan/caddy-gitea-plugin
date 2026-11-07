# Caddy plugin: Gitea

A [Caddy](https://caddyserver.com) plugin that runs [Gitea](https://about.gitea.com),
the version control server, inside Caddy's lifetime and serves it through Caddy - one
command starts both, Caddy does TLS, HTTP/2 and HTTP/3 in front of Gitea, and stopping
Caddy stops Gitea. Part of [Weddy](https://github.com/wekan/weddy), Caddy with the Gitea
and [WeKan](https://github.com/wekan/caddy-wekan-plugin) plugins.

| Module | What it does |
|--------|--------------|
| `gitea` (app) | starts the Gitea program as a supervised child process: passes its output to Caddy's log, knows when it accepts connections, restarts it when it exits on its own (waiting 1 s, doubling to 1 min), and stops it - SIGTERM, then SIGKILL after `stop_timeout` - when Caddy stops |
| `http.handlers.gitea` | serves a site from that Gitea with Caddy's own reverse proxy; answers `503` with `Retry-After` while Gitea is starting |

Gitea stays its own program: no Gitea code is compiled into Caddy, so each is
upgraded on its own - the design of WeKan's
[docs/Design/Multiverse/Go.md](https://github.com/wekan/wekan/blob/main/docs/Design/Multiverse/Go.md),
a separate, versioned process that one command supervises. Use a Gitea release of
[wekan/gitea-patches](https://github.com/wekan/gitea-patches/releases) - upstream Gitea
with its Finnish translation and the mirroring tools - or any Gitea.

## Caddyfile

```caddyfile
{
	gitea {
		binary     /usr/local/bin/gitea
		work_path  /var/lib/gitea
		config     /var/lib/gitea/custom/conf/app.ini
	}
}

git.example.org {
	gitea
}
```

| Option | Meaning | Default |
|--------|---------|---------|
| `binary` | the Gitea program | `gitea` from `PATH` |
| `work_path` | Gitea's `--work-path`, also its working directory | Gitea's default |
| `config` | Gitea's `--config`, app.ini | `<work_path>/custom/conf/app.ini` |
| `dial` | where Gitea listens: `host:port` or `unix//path` | read from app.ini |
| `args` | more arguments after `web` | - |
| `env <NAME> <value>` | added to Gitea's environment | - |
| `stop_timeout` | how long Gitea gets to stop | `30s` |

The `gitea` directive takes no arguments; it is ordered before `reverse_proxy`.

## Gitea's app.ini

The plugin reads where Gitea listens from `[server]` - it never writes app.ini (Gitea's
`--port` would, so it is not used):

```ini
[server]
PROTOCOL  = http+unix               ; reached only through Caddy - or:
HTTP_ADDR = /run/gitea/gitea.sock   ; PROTOCOL = http, HTTP_ADDR = 127.0.0.1, HTTP_PORT = 3000
ROOT_URL  = https://git.example.org/
```

`HTTP_ADDR = 0.0.0.0` (Gitea's default) works, with a warning in the log: Gitea then
also answers outside Caddy, without its TLS. `PROTOCOL = https` and `fcgi` cannot be
proxied and are refused.

## For other plugins

[`supervisor`](supervisor/supervisor.go) runs a program as a supervised child process,
and [`proxy`](proxy/proxy.go) serves a site from it. They do not depend on Gitea:
[caddy-wekan-plugin](https://github.com/wekan/caddy-wekan-plugin) uses both to run
WeKan.

## Building

With [Weddy](https://github.com/wekan/weddy), whose releases have Caddy with both
plugins for every platform, or with [xcaddy](https://github.com/caddyserver/xcaddy):

```sh
xcaddy build --with github.com/wekan/caddy-gitea-plugin
```

## Tests

```sh
go test ./...
```

They run a real Caddy with a stand-in Gitea: `503` while it starts, the proxied
answer once it listens, and Gitea gone after Caddy stops; and the supervisor with a
real child process - restart after a crash, a program that ignores SIGTERM is killed.

## License

MIT - see [LICENSE](LICENSE).

Maintainer and contributor rules are in
[WeKan's CLAUDE.md and AGENTS.md](https://github.com/wekan/wekan), which cover every
repository WeKan clones into its `.tools/` directory.
