// Package caddygitea is a Caddy plugin that runs Gitea, the version control server,
// INSIDE Caddy: the needed parts of Gitea's source are in this repository (gitea/),
// copied from upstream Gitea with wekan/gitea-patches and this repository's patches
// applied (releases/sync-gitea.sh), so one Caddy binary serves Gitea itself.
//
//   - the app "gitea" starts Gitea in Caddy's process, from its app.ini;
//   - the handler "gitea" (http.handlers.gitea) serves a site from it - no proxy, no
//     second process, no port of its own.
//
// Caddyfile:
//
//	{
//		gitea {
//			work_path /var/lib/gitea
//		}
//	}
//	git.example.org {
//		gitea
//	}
package caddygitea

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/graceful"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers"
	"github.com/wekan/caddy-gitea-plugin/gitea/routers/install"
)

func init() {
	caddy.RegisterModule(App{})
	caddy.RegisterModule(Handler{})
}

// App runs Gitea inside Caddy.
type App struct {
	// WorkPath is Gitea's WORK_PATH: repositories, data and custom/ are below it.
	WorkPath string `json:"work_path,omitempty"`
	// CustomPath is Gitea's custom directory. Default: <work_path>/custom.
	CustomPath string `json:"custom_path,omitempty"`
	// Config is Gitea's app.ini. Default: <custom_path>/conf/app.ini.
	Config string `json:"config,omitempty"`
	// StopTimeout is how long Gitea gets to finish its work when Caddy exits.
	// Default 30s.
	StopTimeout caddy.Duration `json:"stop_timeout,omitempty"`

	log *zap.Logger
}

// Gitea's state - settings, database engine, queues, cron - is global to the
// process, so there is one Gitea per Caddy process. It starts with the first
// configuration that has the gitea app, stays running across configuration reloads,
// and stops when Caddy exits.
var running struct {
	sync.Mutex
	started bool
	key     string
	handler atomic.Pointer[http.Handler]
	ready   atomic.Bool
}

// CaddyModule returns the Caddy module information.
func (App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{ID: "gitea", New: func() caddy.Module { return new(App) }}
}

// Provision keeps the logger.
func (a *App) Provision(ctx caddy.Context) error {
	a.log = ctx.Logger()
	if a.StopTimeout == 0 {
		a.StopTimeout = caddy.Duration(30 * time.Second)
	}
	return nil
}

func (a *App) key() string { return a.WorkPath + "\x00" + a.CustomPath + "\x00" + a.Config }

// Start starts Gitea in the background: Caddy does not wait for its database
// migrations; the handler answers 503 until Gitea is ready.
func (a *App) Start() error {
	running.Lock()
	defer running.Unlock()
	if running.started {
		if running.key != a.key() {
			a.log.Warn("Gitea's work_path, custom_path or config changed: restart Caddy to use them; Gitea keeps running with the old ones")
		}
		return nil
	}
	running.started, running.key = true, a.key()

	setting.InitWorkPathAndCommonConfig(os.Getenv, setting.ArgWorkPathAndCustomConf{
		WorkPath: a.WorkPath, CustomPath: a.CustomPath, CustomConf: a.Config,
	})
	graceful.Embedded = true // Caddy owns the process: signals, systemd, its HTTP server
	graceful.InitManager(context.Background())
	_ = templates.PageRenderer()

	go func() {
		ctx := graceful.GetManager().HammerContext()
		if !setting.InstallLock {
			a.log.Info("Gitea is not installed yet: serving its install page", zap.String("config", setting.CustomConf))
			routers.InitWebInstallPage(ctx)
			install.OnInstalled = func() { a.serveInstalled(ctx) }
			var h http.Handler = install.Routes()
			running.handler.Store(&h)
			running.ready.Store(true)
			return
		}
		a.serveInstalled(ctx)
	}()
	return nil
}

func (a *App) serveInstalled(ctx context.Context) {
	running.ready.Store(false)
	routers.InitWebInstalled(ctx)
	if _, err := os.Stat(setting.AppDataPath); err != nil {
		a.log.Error("Gitea's APP_DATA_PATH does not exist", zap.String("path", setting.AppDataPath))
	}
	setting.AppDataTempDir("").RemoveOutdated(3 * 24 * time.Hour)
	var h http.Handler = routers.NormalRoutes()
	running.handler.Store(&h)
	running.ready.Store(true)
	a.log.Info("Gitea is ready", zap.String("version", setting.AppVer), zap.String("root_url", setting.AppURL))
}

// Stop stops Gitea when Caddy exits; on a configuration reload Gitea keeps running.
func (a *App) Stop() error {
	if !caddy.Exiting() {
		return nil
	}
	running.Lock()
	defer running.Unlock()
	if !running.started {
		return nil
	}
	running.ready.Store(false)
	graceful.GetManager().DoGracefulShutdown()
	select {
	case <-graceful.GetManager().Done():
		return nil
	case <-time.After(time.Duration(a.StopTimeout)):
		return fmt.Errorf("gitea: did not stop within %v", time.Duration(a.StopTimeout))
	}
}

// Ready reports whether Gitea serves requests.
func (a *App) Ready() bool { return running.ready.Load() }

// Handler serves a site from the Gitea running in Caddy.
type Handler struct{}

// CaddyModule returns the Caddy module information.
func (Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{ID: "http.handlers.gitea", New: func() caddy.Module { return new(Handler) }}
}

// Provision checks that the gitea app is configured.
func (h *Handler) Provision(ctx caddy.Context) error {
	_, err := ctx.AppIfConfigured("gitea")
	return err // "app module gitea: not configured": add the gitea global option
}

// ServeHTTP lets Gitea answer, or answers 503 while it starts.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, _ caddyhttp.Handler) error {
	p := running.handler.Load()
	if p == nil || !running.ready.Load() {
		w.Header().Set("Retry-After", "5")
		return caddyhttp.Error(http.StatusServiceUnavailable, fmt.Errorf("Gitea is starting"))
	}
	(*p).ServeHTTP(w, r)
	return nil
}

var (
	_ caddy.App                   = (*App)(nil)
	_ caddy.Provisioner           = (*App)(nil)
	_ caddy.Provisioner           = (*Handler)(nil)
	_ caddyhttp.MiddlewareHandler = (*Handler)(nil)
)
