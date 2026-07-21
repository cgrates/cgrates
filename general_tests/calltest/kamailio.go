// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package calltest

import (
	"fmt"
	"maps"
	"net"
	"slices"
	"testing"
	"time"
)

// Kamailio runs kamailio from ConfigFile as a foreground process. Start
// blocks until ReadyAddr accepts a connection and kills the process when the
// test ends.
type Kamailio struct {
	ConfigFile string            // -f
	Defines    map[string]string // -A NAME="value", overrides for #!define
	ReadyAddr  string            // address polled for readiness, e.g. the evapi port
	RuntimeDir string            // -Y; defaults to a test temp dir
}

func (k Kamailio) args() []string {
	args := []string{"-f", k.ConfigFile, "-DD", "-E"}
	if k.RuntimeDir != "" {
		args = append(args, "-Y", k.RuntimeDir)
	}
	for _, name := range slices.Sorted(maps.Keys(k.Defines)) {
		args = append(args, "-A", fmt.Sprintf("%s=%q", name, k.Defines[name]))
	}
	return args
}

func (k Kamailio) Start(t testing.TB) {
	t.Helper()
	if k.RuntimeDir == "" {
		k.RuntimeDir = t.TempDir()
	}
	p := startProcess(t, "kamailio", k.args()...)
	p.waitReady(t, 15*time.Second, "kamailio at "+k.ReadyAddr, func() bool {
		c, err := net.DialTimeout("tcp", k.ReadyAddr, 200*time.Millisecond)
		if err != nil {
			return false
		}
		_ = c.Close()
		return true
	})
}
