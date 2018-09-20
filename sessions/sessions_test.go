/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package sessions

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestSessionsNewV1AuthorizeArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1AuthorizeArgs{
		AuthorizeResources: true,
		GetAttributes:      true,
		CGREvent:           cgrEv,
	}
	rply := NewV1AuthorizeArgs(true, true, false, false, false, false, false, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1AuthorizeArgs{
		GetAttributes:         true,
		AuthorizeResources:    false,
		GetMaxUsage:           true,
		ProcessThresholds:     false,
		ProcessStats:          true,
		GetSuppliers:          false,
		SuppliersIgnoreErrors: true,
		SuppliersMaxCost:      utils.MetaSuppliersEventCost,
		CGREvent:              cgrEv,
	}
	rply = NewV1AuthorizeArgs(true, false, true, false, true, false, true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v,\n received: %+v", expected, rply)
	}
}

func TestSessionsNewV1UpdateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply := NewV1UpdateSessionArgs(true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1UpdateSessionArgs{
		GetAttributes: false,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply = NewV1UpdateSessionArgs(false, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1TerminateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1TerminateSessionArgs{
		TerminateSession:  true,
		ProcessThresholds: true,
		CGREvent:          cgrEv,
	}
	rply := NewV1TerminateSessionArgs(true, false, true, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1TerminateSessionArgs{
		CGREvent: cgrEv,
	}
	rply = NewV1TerminateSessionArgs(false, false, false, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1ProcessEventArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1ProcessEventArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
	}
	rply := NewV1ProcessEventArgs(true, true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1ProcessEventArgs{
		AllocateResources: true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
	}
	rply = NewV1ProcessEventArgs(true, false, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1InitSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		ProcessThresholds: true,
		ProcessStats:      true,
		CGREvent:          cgrEv,
	}
	rply := NewV1InitSessionArgs(true, true, true, true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: false,
		InitSession:       true,
		ProcessThresholds: false,
		ProcessStats:      true,
		CGREvent:          cgrEv,
	}
	rply = NewV1InitSessionArgs(true, false, true, false, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1UpdateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}
	expected := &V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent:      cgrEv,
	}
	rply := NewV1UpdateSessionArgs(true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1TerminateSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}

	expected := &V1TerminateSessionArgs{
		TerminateSession:  true,
		ProcessThresholds: true,
		CGREvent:          cgrEv,
	}
	rply := NewV1TerminateSessionArgs(true, false, true, false, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestSessionsNewV1ProcessEventArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}

	expected := &V1ProcessEventArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent:          cgrEv,
	}
	rply := NewV1ProcessEventArgs(true, true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}

}

func TestSessionsNewV1InitSessionArgs(t *testing.T) {
	cgrEv := utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "Event",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.Destination: "1002",
		},
	}

	expected := &V1InitSessionArgs{
		GetAttributes:     true,
		AllocateResources: true,
		InitSession:       true,
		ProcessThresholds: true,
		ProcessStats:      true,
		CGREvent:          cgrEv,
	}
	rply := NewV1InitSessionArgs(true, true, true, true, true, cgrEv)
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}

}
