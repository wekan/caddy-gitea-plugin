package caddygitea

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func freeAddr(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func TestCaddyfile(t *testing.T) {
	adapter := caddyfile.Adapter{ServerType: httpcaddyfile.ServerType{}}
	out, _, err := adapter.Adapt([]byte("{\n\tgitea {\n\t\twork_path /var/lib/gitea\n\t\tcustom_path /etc/gitea\n\t\tconfig /etc/gitea/app.ini\n\t\tstop_timeout 45s\n\t}\n}\ngit.example.org {\n\tgitea\n}\n"), nil)
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct{ Apps map[string]json.RawMessage }
	_ = json.Unmarshal(out, &cfg)
	var app App
	_ = json.Unmarshal(cfg.Apps["gitea"], &app)
	if app.WorkPath != "/var/lib/gitea" || app.CustomPath != "/etc/gitea" || app.Config != "/etc/gitea/app.ini" || app.StopTimeout != caddy.Duration(45e9) {
		t.Fatalf("the gitea option: %+v", app)
	}
	if !strings.Contains(string(cfg.Apps["http"]), `"handler":"gitea"`) {
		t.Fatal("the site has no gitea handler")
	}
	for name, bad := range map[string]string{
		"an unknown option":        "{\n\tgitea {\n\t\tbinary /usr/bin/gitea\n\t}\n}\n",
		"arguments to the handler": "x.example.org {\n\tgitea extra\n}\n",
	} {
		if _, _, err := adapter.Adapt([]byte(bad), nil); err == nil {
			t.Errorf("negative: %s must be refused", name)
		}
	}
}

// A real Caddy runs the embedded Gitea - its migrations on a new SQLite database,
// its templates and assets from gitea/, wekan/gitea-patches' Finnish - and serves it.
func TestCaddyRunsGiteaInside(t *testing.T) {
	work := t.TempDir()
	site := freeAddr(t)
	// Gitea maintains ~/.ssh/authorized_keys of the user it runs as by default: the
	// test points it, and HOME, into its own directory.
	t.Setenv("HOME", work)
	ini := fmt.Sprintf(`APP_NAME = Weddy test
RUN_MODE = prod
WORK_PATH = %[1]s

[database]
DB_TYPE = sqlite3
PATH = %[1]s/data/gitea.db

[repository]
ROOT = %[1]s/repositories

[server]
ROOT_URL = http://%[2]s/
APP_DATA_PATH = %[1]s/data
DISABLE_SSH = true
START_SSH_SERVER = false
SSH_ROOT_PATH = %[1]s/ssh
SSH_CREATE_AUTHORIZED_KEYS_FILE = false
OFFLINE_MODE = true

[security]
INSTALL_LOCK = true
SECRET_KEY = weddy-test-secret-key-weddy-test-secret-key
INTERNAL_TOKEN = weddy-test-internal-token-weddy-test-internal-token-0123456789

[log]
MODE = console
LEVEL = warn
`, filepath.ToSlash(work), site)
	conf := filepath.Join(work, "custom", "conf", "app.ini")
	if err := os.MkdirAll(filepath.Dir(conf), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(conf, []byte(ini), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := map[string]any{
		"admin": map[string]any{"disabled": true},
		"apps": map[string]any{
			"gitea": map[string]any{"work_path": work},
			"http": map[string]any{"servers": map[string]any{"s": map[string]any{
				"listen": []string{site}, "automatic_https": map[string]any{"disable": true},
				"routes": []any{map[string]any{"handle": []any{map[string]any{"handler": "gitea"}}}},
			}}},
		},
	}
	raw, _ := json.Marshal(cfg)
	if err := caddy.Load(raw, true); err != nil {
		t.Fatal(err)
	}
	defer caddy.Stop()

	get := func(path, lang string) (int, string) {
		req, _ := http.NewRequest("GET", "http://"+site+path, nil)
		if lang != "" {
			req.Header.Set("Accept-Language", lang)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err.Error()
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(b)
	}
	code, body := get("/user/login", "fi-FI")
	if code != http.StatusServiceUnavailable && code != http.StatusOK {
		t.Fatalf("first answer %d: %.300s", code, body)
	}
	deadline := time.Now().Add(2 * time.Minute)
	for code != http.StatusOK && time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		code, body = get("/user/login", "fi-FI")
	}
	if code != http.StatusOK {
		t.Fatalf("Gitea did not become ready: %d %.300s", code, body)
	}
	if !strings.Contains(body, "Weddy test") || !strings.Contains(body, "Kirjaudu sisään") {
		t.Fatalf("the login page is not Gitea's in Finnish (Kirjaudu sisään): %.500s", body)
	}
	if code, body := get("/assets/img/logo.svg", ""); code != http.StatusOK || !strings.Contains(body, "<svg") {
		t.Fatalf("negative: embedded public assets are not served: %d %.200s", code, body)
	}
	if code, _ := get("/no/such/user/or/repo", ""); code != http.StatusNotFound {
		t.Fatalf("Gitea's own 404 expected, got %d", code)
	}
	if _, err := os.Stat(filepath.Join(work, ".ssh")); err == nil {
		t.Fatal("negative: Gitea wrote ~/.ssh although the test's app.ini turns that off")
	}
}
