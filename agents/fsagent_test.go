// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package agents

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
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

func TestSetFSSetCGRRequest(t *testing.T) {
	uid := utils.GenUUID()
	fsev := map[string]string{
		EventName:              FsAnswerEvent,
		UUID:                   uid,
		"variable_cgr_account": "1001",
		"variable_*originID":   uid,
		"variable_*sessions":   "true",
		"variable_*totalUsage": "100",
		"variable_*sesTTL":     "540",
	}
	nm := utils.NewOrderedNavigableMap()
	opts := make(utils.MapStorage)
	if err := setFSCGRRequest(fsev, nm, opts); err != nil {
		t.Fatal(err)
	}
	eEv := map[string]any{
		EventName:              FsAnswerEvent,
		UUID:                   uid,
		"variable_cgr_account": "1001",
	}
	rcv := nm.AsMap()
	if diff := cmp.Diff(rcv, eEv); diff != "" {
		t.Errorf("(-expected +got):\n%s", diff)
	}
	eopts := utils.MapStorage{
		"*originID":   uid,
		"*totalUsage": "100",
		"*sessions":   "true",
		"*sesTTL":     "540",
	}
	if diff := cmp.Diff(opts, eopts); diff != "" {
		t.Errorf("(-expected +got):\n%s", diff)
	}
}
