// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestFAsSessionSClientIface(t *testing.T) {
	_ = sessions.BiRPCClient(new(FSsessions))
}

func TestMinAccountUsageProcessEventReply(t *testing.T) {
	rply := sessions.V1ProcessEventReply{
		AccountsUsage: map[string]time.Duration{
			utils.MetaPrimary: 2 * time.Minute,
			utils.MetaDefault: time.Minute,
		},
	}
	got, found := minAccountUsage(&utils.DataNode{
		Type: utils.NMMapType,
		Map:  rply.AsNavigableMap(),
	})
	if !found || got != time.Minute {
		t.Fatalf("minAccountUsage() = %v, %t, want %v, true", got, found, time.Minute)
	}
}
