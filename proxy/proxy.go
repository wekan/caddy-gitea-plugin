// Package proxy sends HTTP requests to a supervised child process with Caddy's own
// reverse proxy - WebSockets, streaming and the X-Forwarded-* headers included - and
// answers 503 with Retry-After while the child is not accepting connections yet, so
// a client sees "starting" rather than a gateway error. caddy-gitea-plugin and
// caddy-wekan-plugin both use it.
package proxy

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
)

// Proxy is a reverse proxy to one child process.
type Proxy struct {
	name  string
	ready func() bool
	rp    *reverseproxy.Handler
}

// New provisions a reverse proxy to dial ("host:port" or "unix//path"). ready
// reports whether the child accepts connections.
func New(ctx caddy.Context, name, dial string, ready func() bool) (*Proxy, error) {
	rp := &reverseproxy.Handler{Upstreams: reverseproxy.UpstreamPool{{Dial: dial}}}
	if err := rp.Provision(ctx); err != nil {
		return nil, fmt.Errorf("%s: reverse proxy to %s: %w", name, dial, err)
	}
	return &Proxy{name: name, ready: ready, rp: rp}, nil
}

// ServeHTTP proxies the request, or answers 503 while the child is starting.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if !p.ready() {
		w.Header().Set("Retry-After", "5")
		return caddyhttp.Error(http.StatusServiceUnavailable, fmt.Errorf("%s is starting", p.name))
	}
	return p.rp.ServeHTTP(w, r, next)
}

// Cleanup releases the reverse proxy.
func (p *Proxy) Cleanup() error {
	if p == nil || p.rp == nil {
		return nil
	}
	return p.rp.Cleanup()
}
