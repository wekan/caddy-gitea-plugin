//go:build unix

package supervisor

import (
	"os/exec"
	"syscall"
)

// The program gets a process group of its own, so that stopping it reaches the
// processes it starts too - WeKan's start-wekan.sh starts node and FerretDB.
func newProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminate(cmd *exec.Cmd) error { return signalGroup(cmd, syscall.SIGTERM) }

func kill(cmd *exec.Cmd) error { return signalGroup(cmd, syscall.SIGKILL) }

func signalGroup(cmd *exec.Cmd, sig syscall.Signal) error {
	if err := syscall.Kill(-cmd.Process.Pid, sig); err != nil {
		return cmd.Process.Signal(sig)
	}
	return nil
}
