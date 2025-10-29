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
	"testing"
)

func TestSetModule(t *testing.T) {
	expEv := &ExportEvents{}
	testModule := "newModuleValue"
	expEv.SetModule(testModule)
	if expEv.module != testModule {
		t.Errorf("Expected module %s, got %s", testModule, expEv.module)
	}
}

func TestAddEvent(t *testing.T) {
	expEv := &ExportEvents{
		Events: []any{},
	}
	event := "testEvent"
	expEv.AddEvent(event)
	if len(expEv.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(expEv.Events))
	}
	if expEv.Events[0] != event {
		t.Errorf("Expected event %v, got %v", event, expEv.Events[0])
	}
}
