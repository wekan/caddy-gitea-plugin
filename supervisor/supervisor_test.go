package supervisor

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The test binary is also the supervised program: SUPERVISOR_HELPER says what to do.
func TestMain(m *testing.M) {
	switch os.Getenv("SUPERVISOR_HELPER") {
	case "serve": // listen on SUPERVISOR_ADDR and serve until stopped
		l, err := net.Listen("tcp", os.Getenv("SUPERVISOR_ADDR"))
		if err != nil {
			os.Exit(3)
		}
		fmt.Println("serving")
		_ = http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, os.Getenv("SUPERVISOR_GREETING"))
		}))
	case "crash-once": // exit at once the first time, serve the next time
		marker := os.Getenv("SUPERVISOR_MARKER")
		if _, err := os.Stat(marker); err != nil {
			_ = os.WriteFile(marker, []byte("x"), 0o600)
			os.Exit(1)
		}
		os.Setenv("SUPERVISOR_HELPER", "serve")
		l, _ := net.Listen("tcp", os.Getenv("SUPERVISOR_ADDR"))
		_ = http.Serve(l, nil)
	case "stubborn": // ignore SIGTERM
		signal.Ignore(syscall.SIGTERM)
		fmt.Println("ignoring")
		time.Sleep(time.Hour)
	default:
		os.Exit(m.Run())
	}
}

func freeAddr(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func self(t *testing.T) string {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return exe
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestStartReadyStop(t *testing.T) {
	addr := freeAddr(t)
	p := New(Config{Name: "t", Command: self(t), Dial: addr,
		Env: []string{"SUPERVISOR_HELPER=serve", "SUPERVISOR_ADDR=" + addr, "SUPERVISOR_GREETING=hello"}}, nil)
	if p.Ready() {
		t.Fatal("ready before start")
	}
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "ready", p.Ready)
	res, err := http.Get("http://" + addr)
	if err != nil {
		t.Fatal(err)
	}
	var b [5]byte
	_, _ = res.Body.Read(b[:])
	res.Body.Close()
	if string(b[:]) != "hello" {
		t.Fatalf("the environment did not reach the program: %q", b)
	}
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
	if p.Ready() {
		t.Fatal("ready after stop")
	}
	if c, err := net.DialTimeout("tcp", addr, time.Second); err == nil {
		c.Close()
		t.Fatal("negative: the program still listens after Stop")
	}
}

func TestRestartAfterCrash(t *testing.T) {
	addr := freeAddr(t)
	p := New(Config{Name: "t", Command: self(t), Dial: addr, MinRestartDelay: 100 * time.Millisecond,
		Env: []string{"SUPERVISOR_HELPER=crash-once", "SUPERVISOR_ADDR=" + addr,
			"SUPERVISOR_MARKER=" + filepath.Join(t.TempDir(), "crashed")}}, nil)
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	defer p.Stop()
	waitFor(t, "ready after the restart", p.Ready)
	if p.Runs() != 2 {
		t.Fatalf("runs = %d, want 2 (one crash, one restart)", p.Runs())
	}
}

func TestStopKillsAProgramThatIgnoresSIGTERM(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no SIGTERM on Windows")
	}
	p := New(Config{Name: "t", Command: self(t), StopTimeout: 500 * time.Millisecond,
		Env: []string{"SUPERVISOR_HELPER=stubborn"}}, nil)
	if err := p.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)
	start := time.Now()
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
	if d := time.Since(start); d < 400*time.Millisecond || d > 10*time.Second {
		t.Fatalf("Stop took %v: it should wait StopTimeout, then kill", d)
	}
}

func TestStartRefusesAMissingProgram(t *testing.T) {
	p := New(Config{Name: "t", Command: filepath.Join(t.TempDir(), "no-such-program")}, nil)
	if err := p.Start(); err == nil {
		t.Fatal("negative: a missing program must be an error at Start")
	}
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestDialNetwork(t *testing.T) {
	for in, want := range map[string]string{"127.0.0.1:3000": "tcp 127.0.0.1:3000", "unix//run/g.sock": "unix /run/g.sock"} {
		n, a := DialNetwork(in)
		if got := n + " " + a; got != want {
			t.Errorf("%s: %s, want %s", in, got, want)
		}
	}
	if !strings.HasPrefix(fmt.Sprint(DialNetwork("[::1]:80")), "tcp") {
		t.Error("IPv6 is tcp")
	}
}
