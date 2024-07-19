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

package engine

import (
	"testing"
)

func TestNewCoreService(t *testing.T) {
	coreService := NewCoreService()
	if coreService == nil {
		t.Fatal("Expected non-nil *CoreService, got nil")
	}
	if _, ok := interface{}(coreService).(*CoreService); !ok {
		t.Fatalf("Expected type *CoreService, got %T", coreService)
	}
}

func TestListenAndServe(t *testing.T) {
	coreService := &CoreService{}
	exitChan := make(chan bool, 1)
	go func() {
		err := coreService.ListenAndServe(exitChan)
		if err != nil {
			t.Errorf("ListenAndServe returned an error: %v", err)
		}
	}()
	exitChan <- true
}

func TestShutdown(t *testing.T) {
	coreService := &CoreService{}
	err := coreService.Shutdown()
	if err != nil {
		t.Errorf("Shutdown returned an error: %v", err)
	}
}
