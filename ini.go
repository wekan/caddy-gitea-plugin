package caddygitea

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// DialFromINI reads where Gitea listens from app.ini's [server] section:
//
//	PROTOCOL = http        HTTP_ADDR = 127.0.0.1   HTTP_PORT = 3000  ->  127.0.0.1:3000
//	PROTOCOL = http+unix   HTTP_ADDR = /run/gitea.sock                ->  unix//run/gitea.sock
//
// Gitea's defaults apply to what is not set (http, 0.0.0.0, 3000), and a missing
// app.ini - before Gitea's first start - is Gitea's defaults too. 0.0.0.0 and :: are
// dialled on the loopback address, and returned with a warning: Gitea then also
// answers outside Caddy, without Caddy's TLS. https and fcgi are not proxied.
func DialFromINI(path string) (dial, warning string, err error) {
	vals := map[string]string{}
	if path != "" {
		f, err := os.Open(path)
		if err != nil && !os.IsNotExist(err) {
			return "", "", err
		}
		if err == nil {
			defer f.Close()
			section := ""
			s := bufio.NewScanner(f)
			for s.Scan() {
				line := strings.TrimSpace(s.Text())
				if line == "" || line[0] == ';' || line[0] == '#' {
					continue
				}
				if line[0] == '[' && strings.HasSuffix(line, "]") {
					section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
					continue
				}
				if section != "server" {
					continue
				}
				k, v, ok := strings.Cut(line, "=")
				if !ok {
					continue
				}
				v = strings.TrimSpace(v)
				if len(v) >= 2 && (v[0] == '"' || v[0] == '`') && v[len(v)-1] == v[0] {
					v = v[1 : len(v)-1]
				}
				vals[strings.ToUpper(strings.TrimSpace(k))] = v
			}
			if err := s.Err(); err != nil {
				return "", "", err
			}
		}
	}
	protocol := strings.ToLower(vals["PROTOCOL"])
	if protocol == "" {
		protocol = "http"
	}
	addr, port := vals["HTTP_ADDR"], vals["HTTP_PORT"]
	switch protocol {
	case "http+unix":
		if addr == "" {
			return "", "", fmt.Errorf("app.ini: PROTOCOL = http+unix needs HTTP_ADDR, the socket path")
		}
		return "unix/" + addr, "", nil
	case "http":
		if port == "" {
			port = "3000"
		}
		if addr == "" || addr == "0.0.0.0" || addr == "::" || addr == "[::]" {
			return net.JoinHostPort("127.0.0.1", port),
				"Gitea listens on all addresses (app.ini [server] HTTP_ADDR); set HTTP_ADDR = 127.0.0.1, or " +
					"PROTOCOL = http+unix, so it is reached only through Caddy", nil
		}
		return net.JoinHostPort(strings.Trim(addr, "[]"), port), "", nil
	default:
		return "", "", fmt.Errorf("app.ini: PROTOCOL = %s cannot be proxied; use http or http+unix behind Caddy", protocol)
	}
}
