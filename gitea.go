// Package caddygitea is a Caddy plugin that runs Gitea, the version control server,
// inside Caddy's lifetime and serves it through Caddy:
//
//   - the app "gitea" starts the Gitea program - a release of
//     https://github.com/wekan/gitea-patches - as a supervised child process,
//     restarts it when it exits, and stops it when Caddy stops;
//   - the handler "gitea" (http.handlers.gitea) proxies requests to it with Caddy's
//     own reverse proxy, so Caddy does TLS, HTTP/2 and HTTP/3 in front of it.
//
// Caddyfile:
//
//	{
//		gitea {
//			binary     /usr/local/bin/gitea
//			work_path  /var/lib/gitea
//			config     /var/lib/gitea/custom/conf/app.ini
//		}
//	}
//	git.example.org {
//		gitea
//	}
//
// Gitea stays its own program: no Gitea code is compiled into Caddy, so Gitea and
// Caddy are upgraded independently (WeKan's docs/Design/Multiverse/Go.md).
package caddygitea

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"

	"github.com/wekan/caddy-gitea-plugin/supervisor"
)

func init() {
	caddy.RegisterModule(App{})
	caddy.RegisterModule(Handler{})
}

// App runs Gitea.
type App struct {
	// Binary is the Gitea program. Default: "gitea" from PATH.
	Binary string `json:"binary,omitempty"`
	// WorkPath is Gitea's working directory (--work-path).
	WorkPath string `json:"work_path,omitempty"`
	// Config is Gitea's app.ini (--config). Default: <work_path>/custom/conf/app.ini.
	Config string `json:"config,omitempty"`
	// Dial is where Gitea listens, as Caddy's reverse proxy writes it: "host:port"
	// or "unix//path". Default: read from app.ini's [server] PROTOCOL, HTTP_ADDR and
	// HTTP_PORT - this plugin never writes app.ini.
	Dial string `json:"dial,omitempty"`
	// Args are added after "web".
	Args []string `json:"args,omitempty"`
	// Env is added to Gitea's environment.
	Env map[string]string `json:"env,omitempty"`
	// StopTimeout is how long Gitea gets to stop before it is killed. Default 30s.
	StopTimeout caddy.Duration `json:"stop_timeout,omitempty"`

	proc *supervisor.Process
	log  *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{ID: "gitea", New: func() caddy.Module { return new(App) }}
}

// Provision works out the command line and where Gitea listens.
func (a *App) Provision(ctx caddy.Context) error {
	a.log = ctx.Logger()
	if a.Binary == "" {
		a.Binary = "gitea"
	}
	if a.Config == "" && a.WorkPath != "" {
		a.Config = filepath.Join(a.WorkPath, "custom", "conf", "app.ini")
	}
	if a.Dial == "" {
		dial, warn, err := DialFromINI(a.Config)
		if err != nil {
			return fmt.Errorf("gitea: %w; set dial in the gitea options", err)
		}
		if warn != "" {
			a.log.Warn(warn)
		}
		a.Dial = dial
	}
	a.proc = supervisor.New(supervisor.Config{
		Name:        "gitea",
		Command:     a.Binary,
		Args:        a.CommandArgs(),
		Dir:         a.WorkPath,
		Env:         envList(a.Env),
		Dial:        a.Dial,
		StopTimeout: time.Duration(a.StopTimeout),
	}, a.log)
	return nil
}

// CommandArgs is Gitea's command line after the program: the global options, then
// "web". No --port: Gitea would write LOCAL_ROOT_URL into app.ini on every start.
func (a *App) CommandArgs() []string {
	var args []string
	if a.WorkPath != "" {
		args = append(args, "--work-path", a.WorkPath)
	}
	if a.Config != "" {
		args = append(args, "--config", a.Config)
	}
	return append(append(args, "web"), a.Args...)
}

// Validate checks what can be checked before starting.
func (a *App) Validate() error {
	if a.Config != "" {
		if _, err := os.Stat(a.Config); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("gitea: config %s: %w", a.Config, err)
		}
	}
	return nil
}

// Start starts Gitea.
func (a *App) Start() error { return a.proc.Start() }

// Stop stops Gitea and waits until it has exited.
func (a *App) Stop() error { return a.proc.Stop() }

// Ready reports whether Gitea accepts connections.
func (a *App) Ready() bool { return a.proc != nil && a.proc.Ready() }

func envList(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

var (
	_ caddy.App         = (*App)(nil)
	_ caddy.Provisioner = (*App)(nil)
	_ caddy.Validator   = (*App)(nil)
)
