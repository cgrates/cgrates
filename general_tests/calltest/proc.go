/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package calltest

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const processStopTimeout = 5 * time.Second

func needBinary(t testing.TB, name string) string {
	t.Helper()
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	for _, dir := range []string{"/usr/sbin", "/sbin"} {
		path := filepath.Join(dir, name)
		if st, err := os.Stat(path); err == nil && !st.IsDir() && st.Mode()&0111 != 0 {
			return path
		}
	}
	t.Skipf("%s not found in PATH, /usr/sbin, or /sbin; install it to run this test", name)
	return ""
}

// proc is a foreground process whose lifetime is tied to the test.
type proc struct {
	stdout *strings.Builder
	stderr *strings.Builder
	done   chan struct{} // closed once the process has exited and been reaped
}

// startProcess starts name in its own process group. Children die with the
// parent, so old workers do not keep ports open after cleanup.
func startProcess(t testing.TB, name string, args ...string) *proc {
	t.Helper()
	path := needBinary(t, name)
	return startCmd(t, name, exec.Command(path, args...))
}

// startCmd starts cmd in its own process group.
func startCmd(t testing.TB, name string, cmd *exec.Cmd) *proc {
	t.Helper()
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", name, err)
	}
	p := &proc{stdout: &stdout, stderr: &stderr, done: make(chan struct{})}
	go func() {
		_ = cmd.Wait()
		close(p.done)
	}()
	t.Cleanup(func() {
		defer logProcessOutput(t, name, &stdout, &stderr)
		select {
		case <-p.done:
			return
		default:
		}
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
			t.Logf("stop %s: %v", name, err)
		}
		if waitDone(p.done, processStopTimeout) {
			return
		}
		t.Logf("%s did not exit after SIGTERM; sending SIGKILL", name)
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			t.Logf("kill %s: %v", name, err)
		}
		if waitDone(p.done, processStopTimeout) {
			return
		}
		t.Errorf("%s did not exit after SIGKILL", name)
	})
	return p
}

func waitDone(done <-chan struct{}, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return true
	case <-timer.C:
		return false
	}
}

func logProcessOutput(t testing.TB, name string, stdout, stderr *strings.Builder) {
	t.Helper()
	if t.Failed() {
		t.Logf("%s stdout:\n%s", name, stdout.String())
		t.Logf("%s stderr:\n%s", name, stderr.String())
	}
}

// waitReady fails early if the process exits before ok returns true.
func (p *proc) waitReady(t testing.TB, timeout time.Duration, label string, ok func() bool) {
	t.Helper()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-p.done:
			t.Fatalf("%s exited before ready\nstdout:\n%s\nstderr:\n%s",
				label, p.stdout.String(), p.stderr.String())
		case <-deadline.C:
			t.Fatalf("%s not ready after %s", label, timeout)
		case <-tick.C:
			if ok() {
				return
			}
		}
	}
}
