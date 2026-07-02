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
	"testing"
	"time"
)

// CallParams configures one originated call.
type CallParams struct {
	From     string
	To       string
	HoldTime time.Duration
}

// UAC places a call through a SIP endpoint and blocks until it completes.
type UAC interface {
	Call(t testing.TB, c CallParams)
}

// UAS answers calls in the background until the test ends.
type UAS interface {
	Start(t testing.TB)
}

var (
	_ UAC = SipgoUAC{}
	_ UAC = SippUAC{}
	_ UAC = VoiceBlenderUAC{}
	_ UAS = SipgoUAS{}
	_ UAS = SippUAS{}
	_ UAS = VoiceBlenderUAS{}
)

func checkCallParams(t testing.TB, backend string, c CallParams) {
	t.Helper()
	if c.From == "" {
		t.Fatalf("%s: from not set", backend)
	}
	if c.To == "" {
		t.Fatalf("%s: to not set", backend)
	}
	if c.HoldTime <= 0 {
		t.Fatalf("%s: hold time not set", backend)
	}
}

func checkAddr(t testing.TB, backend, addr string) {
	t.Helper()
	if addr == "" {
		t.Fatalf("%s: addr not set", backend)
	}
}
