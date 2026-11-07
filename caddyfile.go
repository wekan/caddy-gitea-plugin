package caddygitea

import (
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"github.com/wekan/caddy-gitea-plugin/proxy"
)

func init() {
	httpcaddyfile.RegisterGlobalOption("gitea", parseGlobal)
	httpcaddyfile.RegisterHandlerDirective("gitea", parseHandler)
	httpcaddyfile.RegisterDirectiveOrder("gitea", httpcaddyfile.Before, "reverse_proxy")
}

// Handler serves a site from the Gitea that the gitea app runs.
type Handler struct {
	app   *App
	proxy *proxy.Proxy
}

// CaddyModule returns the Caddy module information.
func (Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{ID: "http.handlers.gitea", New: func() caddy.Module { return new(Handler) }}
}

// Provision connects the handler to the gitea app.
func (h *Handler) Provision(ctx caddy.Context) error {
	app, err := ctx.AppIfConfigured("gitea")
	if err != nil {
		return err // "app module gitea: not configured": add the gitea global option
	}
	h.app = app.(*App)
	h.proxy, err = proxy.New(ctx, "Gitea", h.app.Dial, h.app.Ready)
	return err
}

// ServeHTTP proxies the request to Gitea.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return h.proxy.ServeHTTP(w, r, next)
}

// Cleanup releases the reverse proxy.
func (h *Handler) Cleanup() error { return h.proxy.Cleanup() }

// UnmarshalCaddyfile: the directive takes no arguments, "gitea".
func (h *Handler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	if d.NextArg() || d.NextBlock(0) {
		return d.Err("gitea takes no arguments: its settings are in the global gitea option")
	}
	return nil
}

func parseHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var handler Handler
	return &handler, handler.UnmarshalCaddyfile(h.Dispenser)
}

// UnmarshalCaddyfile reads the global option:
//
//	gitea {
//		binary       <path>
//		work_path    <dir>
//		config       <app.ini>
//		dial         <host:port | unix//path>
//		args         <arg>...
//		env          <KEY> <value>
//		stop_timeout <duration>
//	}
func (a *App) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // the option name
	if d.NextArg() {
		return d.ArgErr()
	}
	for d.NextBlock(0) {
		key := d.Val()
		switch key {
		case "binary", "work_path", "config", "dial":
			if !d.NextArg() {
				return d.ArgErr()
			}
			v := d.Val()
			switch key {
			case "binary":
				a.Binary = v
			case "work_path":
				a.WorkPath = v
			case "config":
				a.Config = v
			case "dial":
				a.Dial = v
			}
			if d.NextArg() {
				return d.ArgErr()
			}
		case "args":
			a.Args = append(a.Args, d.RemainingArgs()...)
		case "env":
			args := d.RemainingArgs()
			if len(args) < 2 {
				return d.Err("env needs a name and a value")
			}
			if a.Env == nil {
				a.Env = map[string]string{}
			}
			a.Env[args[0]] = strings.Join(args[1:], " ")
		case "stop_timeout":
			if !d.NextArg() {
				return d.ArgErr()
			}
			dur, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("stop_timeout: %v", err)
			}
			a.StopTimeout = caddy.Duration(dur)
		default:
			return d.Errf("unknown gitea option %q", key)
		}
	}
	return nil
}

func parseGlobal(d *caddyfile.Dispenser, _ any) (any, error) {
	var app App
	if err := app.UnmarshalCaddyfile(d); err != nil {
		return nil, err
	}
	return httpcaddyfile.App{Name: "gitea", Value: caddyconfig.JSON(app, nil)}, nil
}

var (
	_ caddy.Provisioner           = (*Handler)(nil)
	_ caddy.CleanerUpper          = (*Handler)(nil)
	_ caddyhttp.MiddlewareHandler = (*Handler)(nil)
	_ caddyfile.Unmarshaler       = (*Handler)(nil)
	_ caddyfile.Unmarshaler       = (*App)(nil)
)
