// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
