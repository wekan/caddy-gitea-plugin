package caddygitea

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func ini(t *testing.T, text string) string {
	p := filepath.Join(t.TempDir(), "app.ini")
	if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDialFromINI(t *testing.T) {
	for _, c := range []struct {
		name, ini, dial string
		warn, fail      bool
	}{
		{"loopback", "[server]\nHTTP_ADDR = 127.0.0.1\nHTTP_PORT = 3001\n", "127.0.0.1:3001", false, false},
		{"unix socket", "[server]\nPROTOCOL = http+unix\nHTTP_ADDR = /run/gitea/gitea.sock\n", "unix//run/gitea/gitea.sock", false, false},
		{"defaults: all addresses, warned", "[server]\nDOMAIN = x\n", "127.0.0.1:3000", true, false},
		{"other sections do not count", "[database]\nHTTP_PORT = 9\n[server]\nHTTP_ADDR = \"127.0.0.1\"\n", "127.0.0.1:3000", false, false},
		{"IPv6", "[server]\nHTTP_ADDR = ::1\nHTTP_PORT = 3000\n", "[::1]:3000", false, false},
		{"negative: https cannot be proxied", "[server]\nPROTOCOL = https\n", "", false, true},
		{"negative: a socket needs its path", "[server]\nPROTOCOL = http+unix\n", "", false, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			dial, warn, err := DialFromINI(ini(t, c.ini))
			if (err != nil) != c.fail || dial != c.dial || (warn != "") != c.warn {
				t.Fatalf("got %q, warning %q, error %v", dial, warn, err)
			}
		})
	}
	if dial, _, err := DialFromINI(filepath.Join(t.TempDir(), "missing.ini")); err != nil || dial != "127.0.0.1:3000" {
		t.Fatalf("a missing app.ini is Gitea's defaults: %q %v", dial, err)
	}
}

func TestCommandArgs(t *testing.T) {
	a := App{WorkPath: "/var/lib/gitea", Config: "/etc/gitea/app.ini", Args: []string{"--verbose"}}
	want := []string{"--work-path", "/var/lib/gitea", "--config", "/etc/gitea/app.ini", "web", "--verbose"}
	if got := a.CommandArgs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
	for _, arg := range a.CommandArgs() {
		if strings.HasPrefix(arg, "--port") {
			t.Fatal("negative: --port makes Gitea write LOCAL_ROOT_URL into app.ini")
		}
	}
}

func TestCaddyfile(t *testing.T) {
	const input = `{
	gitea {
		binary /usr/local/bin/gitea
		work_path /var/lib/gitea
		dial unix//run/gitea.sock
		env GITEA_CUSTOM /etc/gitea custom
		args --quiet
		stop_timeout 45s
	}
}
git.example.org {
	gitea
}`
	adapter := caddyfile.Adapter{ServerType: httpcaddyfileServerType()}
	out, warnings, err := adapter.Adapt([]byte(input), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range warnings {
		if !strings.Contains(w.Message, "not formatted") {
			t.Fatalf("warning: %v", w)
		}
	}
	var cfg struct {
		Apps map[string]json.RawMessage `json:"apps"`
	}
	if err := json.Unmarshal(out, &cfg); err != nil {
		t.Fatal(err)
	}
	var app App
	if err := json.Unmarshal(cfg.Apps["gitea"], &app); err != nil {
		t.Fatal(err)
	}
	if app.Binary != "/usr/local/bin/gitea" || app.WorkPath != "/var/lib/gitea" || app.Dial != "unix//run/gitea.sock" ||
		app.Env["GITEA_CUSTOM"] != "/etc/gitea custom" || !reflect.DeepEqual(app.Args, []string{"--quiet"}) ||
		app.StopTimeout != caddy.Duration(45e9) {
		t.Fatalf("the gitea option: %+v", app)
	}
	if !strings.Contains(string(cfg.Apps["http"]), `"handler":"gitea"`) {
		t.Fatalf("the site has no gitea handler: %s", cfg.Apps["http"])
	}
	if _, _, err := adapter.Adapt([]byte("{\n\tgitea {\n\t\tcolour red\n\t}\n}\n"), nil); err == nil {
		t.Fatal("negative: an unknown option must be an error")
	}
	if _, _, err := adapter.Adapt([]byte("x.example.org {\n\tgitea extra\n}\n"), nil); err == nil {
		t.Fatal("negative: the gitea directive takes no arguments")
	}
}

// A real Caddy: the gitea app supervises a stand-in Gitea (this test binary, which
// starts listening after a delay), and the gitea handler serves a site from it.
func TestCaddyRunsAndProxiesGitea(t *testing.T) {
	exe, _ := os.Executable()
	backend, site := freeAddr(t), freeAddr(t)
	cfg := map[string]any{
		"admin": map[string]any{"disabled": true},
		"apps": map[string]any{
			"gitea": map[string]any{"binary": exe, "dial": backend, "stop_timeout": "5s",
				"env": map[string]string{"GITEA_TEST_HELPER": "serve", "GITEA_TEST_ADDR": backend, "GITEA_TEST_DELAY": "1s"}},
			"http": map[string]any{
				"http_port": 0, "https_port": 0,
				"servers": map[string]any{"s": map[string]any{
					"listen":          []string{site},
					"automatic_https": map[string]any{"disable": true},
					"routes":          []any{map[string]any{"handle": []any{map[string]any{"handler": "gitea"}}}},
				}},
			},
		},
	}
	raw, _ := json.Marshal(cfg)
	if err := caddy.Load(raw, true); err != nil {
		t.Fatal(err)
	}
	stopped := false
	defer func() {
		if !stopped {
			_ = caddy.Stop()
		}
	}()

	res, err := http.Get("http://" + site + "/wekan/wekan")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable || res.Header.Get("Retry-After") == "" {
		t.Fatalf("negative: while Gitea starts the answer must be 503 with Retry-After, got %d", res.StatusCode)
	}

	var body string
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		res, err := http.Get("http://" + site + "/wekan/wekan")
		if err == nil {
			b, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode == http.StatusOK {
				body = string(b)
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if body != "from gitea /wekan/wekan" {
		t.Fatalf("proxied answer %q", body)
	}

	if err := caddy.Stop(); err != nil {
		t.Fatal(err)
	}
	stopped = true
	if c, err := net.DialTimeout("tcp", backend, time.Second); err == nil {
		c.Close()
		t.Fatal("negative: Gitea still runs after Caddy stopped")
	}
}
