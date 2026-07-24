// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
