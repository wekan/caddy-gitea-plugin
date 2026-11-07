package caddygitea

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	httpcaddyfile.RegisterGlobalOption("gitea", parseGlobal)
	httpcaddyfile.RegisterHandlerDirective("gitea", parseHandler)
	httpcaddyfile.RegisterDirectiveOrder("gitea", httpcaddyfile.Before, "reverse_proxy")
}

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
//		work_path    <dir>
//		custom_path  <dir>
//		config       <app.ini>
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
		case "work_path", "custom_path", "config":
			if !d.NextArg() {
				return d.ArgErr()
			}
			switch key {
			case "work_path":
				a.WorkPath = d.Val()
			case "custom_path":
				a.CustomPath = d.Val()
			case "config":
				a.Config = d.Val()
			}
			if d.NextArg() {
				return d.ArgErr()
			}
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
	_ caddyfile.Unmarshaler = (*Handler)(nil)
	_ caddyfile.Unmarshaler = (*App)(nil)
)
