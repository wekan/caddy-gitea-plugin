//go:build !unix

package supervisor

import "os/exec"

// Without process groups (Windows, Plan 9) the program itself is stopped; it is
// the program's job to stop what it started.
func newProcessGroup(cmd *exec.Cmd) {}

func terminate(cmd *exec.Cmd) error { return cmd.Process.Kill() }

func kill(cmd *exec.Cmd) error { return cmd.Process.Kill() }
