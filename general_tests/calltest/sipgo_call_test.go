//go:build call

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package calltest

import (
	"testing"
	"time"
)

func TestSipgoUACBasicCall(t *testing.T) {
	SipgoUAS{Port: 5094}.Start(t)
	SipgoUAC{Addr: "127.0.0.1:5094"}.Call(t, CallParams{
		To:       "test",
		From:     "1001",
		HoldTime: time.Second,
	})
}
