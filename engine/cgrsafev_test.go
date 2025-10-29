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
