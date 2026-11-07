# embedded-in-a-host

Run Gitea inside another program - here, inside Caddy.

Gitea assumes it owns its process and its HTTP server. Inside Caddy both are
Caddy's, and two things would go wrong:

- Gitea's graceful manager handles SIGHUP, SIGTERM, SIGINT and SIGUSR1/2 itself:
  a SIGHUP makes it "restart" by re-executing the current binary - which is
  Caddy. It also notifies systemd, runs a Windows service, and shuts Gitea down
  when its own HTTP servers are not created within STARTUP_TIMEOUT; embedded,
  they never are.
- When the install page is done, Gitea shuts down `ctx.Value(http.ServerContextKey)`
  - the server that served the request, which is Caddy's - to start the normal
  routes.

The patch adds `graceful.Embedded`, which the host sets before
`graceful.InitManager` to skip all of the manager's process handling, and
`install.OnInstalled`, which the host sets to be called instead of the server
shutdown. With neither set, Gitea behaves exactly as upstream.

**Files:** `modules/graceful/manager.go`, `modules/graceful/manager_unix.go`,
`modules/graceful/manager_windows.go`, `routers/install/install.go`
**Applies to:** upstream Gitea main (verified against e09813f472), after
wekan/gitea-patches.
