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
