//go:build call

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package calltest

import (
	"testing"
	"time"
)

func TestVoiceBlenderUACBasicCall(t *testing.T) {
	VoiceBlenderUAS{Port: 5092}.Start(t)
	client := VoiceBlenderServer{Port: 5090}.Start(t)
	VoiceBlenderUAC{Client: client, Addr: "127.0.0.1:5092"}.Call(t, CallParams{
		To:       "test",
		From:     "1001",
		HoldTime: 2 * time.Second,
	})
}
