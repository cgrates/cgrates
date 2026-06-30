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
	"os/exec"
	"slices"
	"strconv"
	"testing"
	"time"
)

const sippReadyDelay = 500 * time.Millisecond

// SippUAC places calls through Addr with sipp. Calls, Rate and Limit map to
// -m, -r and -l.
type SippUAC struct {
	Addr     string            // proxy address, host:port
	Scenario string            // scenario file (-sf); built in uac when empty
	Keys     map[string]string // -key NAME VALUE substitutions
	Calls    int               // total calls (-m), defaults to 1
	Rate     int               // calls per second (-r)
	Limit    int               // max simultaneous calls (-l)
}

func (u SippUAC) args(c CallParams) []string {
	args := scenarioArgs(u.Scenario, "uac")
	if c.To != "" {
		args = append(args, "-s", c.To)
	}
	args = append(args, "-m", strconv.Itoa(max(u.Calls, 1)))
	if u.Rate > 0 {
		args = append(args, "-r", strconv.Itoa(u.Rate))
	}
	if u.Limit > 0 {
		args = append(args, "-l", strconv.Itoa(u.Limit))
	}
	args = append(args, "-d", strconv.FormatInt(c.HoldTime.Milliseconds(), 10))
	args = append(args, keyArgs(u.callKeys(c))...)
	return append(args, u.Addr)
}

func (u SippUAC) callKeys(c CallParams) map[string]string {
	keys := make(map[string]string, len(u.Keys)+1)
	maps.Copy(keys, u.Keys)
	keys["fromUser"] = c.From
	return keys
}

// Call blocks until sipp exits, failing t if any call did not complete its
// scenario.
func (u SippUAC) Call(t testing.TB, c CallParams) {
	t.Helper()
	checkCallParams(t, "sipp uac", c)
	checkAddr(t, "sipp uac", u.Addr)
	path := needBinary(t, "sipp")
	out, err := exec.Command(path, u.args(c)...).CombinedOutput()
	if err != nil {
		t.Fatalf("sipp uac %s->%s: %v\n%s", u.Addr, c.To, err, out)
	}
}

// SippUAS answers calls on Port until the test ends.
type SippUAS struct {
	Port     int
	Scenario string            // scenario file (-sf); built in uas when empty
	Keys     map[string]string // -key NAME VALUE substitutions
	Calls    int               // exit after this many calls (-m); 0 runs until killed
}

func (u SippUAS) args() []string {
	args := scenarioArgs(u.Scenario, "uas")
	args = append(args, "-p", strconv.Itoa(u.Port))
	if u.Calls > 0 {
		args = append(args, "-m", strconv.Itoa(u.Calls))
	}
	return append(args, keyArgs(u.Keys)...)
}

func (u SippUAS) Start(t testing.TB) {
	t.Helper()
	if u.Port == 0 {
		t.Fatal("sipp uas: port not set")
	}
	p := startProcess(t, "sipp", u.args()...)
	// sipp has no readiness socket; delay gives time to exit in case of errors.
	readyAt := time.Now().Add(sippReadyDelay)
	p.waitReady(t, time.Second, fmt.Sprintf("sipp uas on :%d", u.Port), func() bool {
		return time.Now().After(readyAt)
	})
}

func scenarioArgs(scenario, preset string) []string {
	if scenario != "" {
		return []string{"-sf", scenario}
	}
	return []string{"-sn", preset}
}

func keyArgs(keys map[string]string) []string {
	args := make([]string, 0, 3*len(keys))
	for _, name := range slices.Sorted(maps.Keys(keys)) {
		args = append(args, "-key", name, keys[name])
	}
	return args
}
