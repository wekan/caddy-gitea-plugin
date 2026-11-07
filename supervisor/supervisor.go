// Package supervisor runs one program as a child process of Caddy: it starts it,
// restarts it when it exits on its own, knows when it accepts connections, and stops
// it - and every process it started - when Caddy stops.
//
// It is the design of WeKan's docs/Design/Multiverse/Go.md: a separate program stays
// a separate, versioned process that one command supervises, rather than being
// imported into the same binary. caddy-gitea-plugin runs Gitea with it, and
// caddy-wekan-plugin runs WeKan (with its FerretDB) with it.
package supervisor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Config is what to run and how.
type Config struct {
	// Name is used in log lines.
	Name string
	// Command and Args are the program and its arguments.
	Command string
	Args    []string
	// Dir is the working directory; empty = Caddy's.
	Dir string
	// Env is added to Caddy's environment, KEY=VALUE.
	Env []string
	// Dial is where the program accepts connections, as Caddy's reverse proxy
	// writes it: "host:port" or "unix//path". Ready() is true once it does.
	Dial string
	// StopTimeout is how long the program gets to exit after SIGTERM before it
	// and its children are killed. 0 = 30 s.
	StopTimeout time.Duration
	// MinRestartDelay and MaxRestartDelay bound the wait before a restart, which
	// doubles after each exit that came soon after a start. 0 = 1 s and 1 min.
	MinRestartDelay time.Duration
	MaxRestartDelay time.Duration
}

// Process is one supervised program.
type Process struct {
	cfg   Config
	log   *zap.Logger
	ready atomic.Bool

	mu      sync.Mutex
	cmd     *exec.Cmd
	exited  chan struct{} // closed when the current run has exited
	cancel  context.CancelFunc
	done    chan struct{} // closed when the supervising loop has ended
	started bool
	runs    atomic.Int64
}

// New returns a Process that is not running yet.
func New(cfg Config, log *zap.Logger) *Process {
	if cfg.StopTimeout <= 0 {
		cfg.StopTimeout = 30 * time.Second
	}
	if cfg.MinRestartDelay <= 0 {
		cfg.MinRestartDelay = time.Second
	}
	if cfg.MaxRestartDelay < cfg.MinRestartDelay {
		cfg.MaxRestartDelay = time.Minute
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &Process{cfg: cfg, log: log.With(zap.String("process", cfg.Name))}
}

// Ready reports whether the program is running and accepts connections at Dial.
func (p *Process) Ready() bool { return p.ready.Load() }

// Runs is how many times the program has been started.
func (p *Process) Runs() int64 { return p.runs.Load() }

// Start starts the program and the loop that restarts it. It returns an error only
// when the program cannot be started at all (for example, it does not exist).
func (p *Process) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return fmt.Errorf("%s: already started", p.cfg.Name)
	}
	if _, err := exec.LookPath(p.cfg.Command); err != nil {
		return fmt.Errorf("%s: %w", p.cfg.Name, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel, p.done, p.started = cancel, make(chan struct{}), true
	if err := p.spawnLocked(); err != nil {
		cancel()
		close(p.done)
		return err
	}
	go p.loop(ctx)
	return nil
}

func (p *Process) spawnLocked() error {
	cmd := exec.Command(p.cfg.Command, p.cfg.Args...)
	cmd.Dir = p.cfg.Dir
	cmd.Env = append(os.Environ(), p.cfg.Env...)
	newProcessGroup(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%s: starting %s: %w", p.cfg.Name, p.cfg.Command, err)
	}
	p.runs.Add(1)
	p.log.Info("started", zap.String("command", p.cfg.Command), zap.Strings("args", p.cfg.Args), zap.Int("pid", cmd.Process.Pid))
	var lines sync.WaitGroup
	lines.Add(2)
	go p.pipe(stdout, false, &lines)
	go p.pipe(stderr, true, &lines)
	exited := make(chan struct{})
	p.cmd, p.exited = cmd, exited
	go func() {
		lines.Wait()
		err := cmd.Wait()
		p.ready.Store(false)
		if err != nil {
			p.log.Warn("exited", zap.Error(err))
		} else {
			p.log.Info("exited")
		}
		close(exited)
	}()
	go p.watchReady(exited)
	return nil
}

// pipe passes the program's output to Caddy's log, a line at a time.
func (p *Process) pipe(r io.Reader, isErr bool, wg *sync.WaitGroup) {
	defer wg.Done()
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	for s.Scan() {
		if isErr {
			p.log.Warn(s.Text())
		} else {
			p.log.Info(s.Text())
		}
	}
}

func (p *Process) watchReady(exited chan struct{}) {
	if p.cfg.Dial == "" {
		p.ready.Store(true)
		return
	}
	network, address := DialNetwork(p.cfg.Dial)
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-exited:
			return
		case <-tick.C:
			c, err := net.DialTimeout(network, address, time.Second)
			if err == nil {
				c.Close()
				p.ready.Store(true)
				p.log.Info("accepts connections", zap.String("dial", p.cfg.Dial))
				return
			}
		}
	}
}

// loop restarts the program whenever it exits on its own.
func (p *Process) loop(ctx context.Context) {
	defer close(p.done)
	delay := p.cfg.MinRestartDelay
	for {
		p.mu.Lock()
		exited, startedAt := p.exited, time.Now()
		p.mu.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-exited:
		}
		if time.Since(startedAt) > 5*time.Minute {
			delay = p.cfg.MinRestartDelay // it ran well; start over with a short wait
		}
		p.log.Warn("restarting", zap.Duration("after", delay))
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, p.cfg.MaxRestartDelay)
		p.mu.Lock()
		if ctx.Err() == nil {
			if err := p.spawnLocked(); err != nil {
				p.log.Error("restart failed", zap.Error(err))
				// The previous run's channel is closed: the loop waits the delay again.
			}
		}
		p.mu.Unlock()
	}
}

// Stop stops the program: SIGTERM to it and every process it started, then, after
// StopTimeout, SIGKILL. It returns when they have exited.
func (p *Process) Stop() error {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return nil
	}
	p.cancel()
	cmd, exited := p.cmd, p.exited
	p.started = false
	p.mu.Unlock()
	<-p.done
	p.ready.Store(false)
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	select {
	case <-exited:
		return nil
	default:
	}
	p.log.Info("stopping", zap.Duration("timeout", p.cfg.StopTimeout))
	if err := terminate(cmd); err != nil && !errors.Is(err, os.ErrProcessDone) {
		p.log.Warn("terminate", zap.Error(err))
	}
	select {
	case <-exited:
		return nil
	case <-time.After(p.cfg.StopTimeout):
	}
	p.log.Warn("did not stop in time; killing it")
	if err := kill(cmd); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	<-exited
	return nil
}

// DialNetwork splits Caddy's dial address: "unix//path" -> ("unix", "/path"),
// "host:port" -> ("tcp", "host:port").
func DialNetwork(dial string) (network, address string) {
	if rest, ok := strings.CutPrefix(dial, "unix/"); ok {
		return "unix", rest
	}
	return "tcp", dial
}
