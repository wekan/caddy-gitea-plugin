package caddygitea

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

// The test binary is also the stand-in Gitea: GITEA_TEST_HELPER=serve listens on
// GITEA_TEST_ADDR after GITEA_TEST_DELAY and answers "from gitea <path>".
func TestMain(m *testing.M) {
	if os.Getenv("GITEA_TEST_HELPER") == "serve" {
		d, _ := time.ParseDuration(os.Getenv("GITEA_TEST_DELAY"))
		time.Sleep(d)
		l, err := net.Listen("tcp", os.Getenv("GITEA_TEST_ADDR"))
		if err != nil {
			os.Exit(3)
		}
		_ = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "from gitea "+r.URL.Path)
		}))
		return
	}
	os.Exit(m.Run())
}

func freeAddr(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func httpcaddyfileServerType() caddyfile.ServerType { return httpcaddyfile.ServerType{} }
