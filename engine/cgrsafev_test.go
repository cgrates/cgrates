// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestCGRSafEventNewCGRSafEventFromCGREventAsCGREvent(t *testing.T) {
	tm := time.Date(2021, 8, 15, 14, 30, 45, 100, time.Local)
	cgrEv := &utils.CGREvent{
		Tenant: "test",
		ID:     "test",
		Time:   &tm,
		Event:  map[string]any{"test": 1},
	}

	exp := &CGRSafEvent{
		Tenant: cgrEv.Tenant,
		ID:     cgrEv.ID,
		Time:   cgrEv.Time,
		Event:  NewSafEvent(cgrEv.Event),
	}
	rcv := NewCGRSafEventFromCGREvent(cgrEv)

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	rcv2 := rcv.AsCGREvent()

	if !reflect.DeepEqual(cgrEv, rcv2) {
		t.Errorf("\nexpected: %s\nreceived: %s\n", utils.ToJSON(cgrEv), utils.ToJSON(rcv2))
	}
}
