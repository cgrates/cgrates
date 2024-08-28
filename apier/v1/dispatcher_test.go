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

package v1

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/utils"
)

func TestPing(t *testing.T) {
	dispatcher := &DispatcherSv1{}
	ctx := context.Background()
	var reply string
	err := dispatcher.Ping(ctx, nil, &reply)
	if err != nil {
		t.Errorf("Ping method returned an error: %v", err)
	}
	if reply != utils.Pong {
		t.Errorf("Expected reply %s, got %s", utils.Pong, reply)
	}
}

func TestNewDispatcherEeSv1(t *testing.T) {
	dispatcherService := &dispatchers.DispatcherService{}
	dispatcher := NewDispatcherEeSv1(dispatcherService)
	if dispatcher == nil {
		t.Fatal("Expected NewDispatcherEeSv1 to return a non-nil DispatcherEeSv1")
	}
	if dispatcher.dS != dispatcherService {
		t.Errorf("Expected dS to be %v, got %v", dispatcherService, dispatcher.dS)
	}
}

func TestNewDispatcherCoreSv1(t *testing.T) {
	mockService := &dispatchers.DispatcherService{}
	dispatcher := NewDispatcherCoreSv1(mockService)
	if dispatcher == nil {
		t.Fatal("Expected dispatcher to be non-nil")
	}
	if dispatcher.dS != mockService {
		t.Errorf("Expected dispatcher.dS to be %v, got %v", mockService, dispatcher.dS)
	}
}

func TestNewDispatcherSv1(t *testing.T) {
	mockDispatcherService := &dispatchers.DispatcherService{}
	dispatcher := NewDispatcherSv1(mockDispatcherService)
	if dispatcher == nil {
		t.Fatal("Expected a non-nil DispatcherSv1, got nil")
	}
	if dispatcher.dS != mockDispatcherService {
		t.Errorf("Expected DispatcherService to be %v, got %v", mockDispatcherService, dispatcher.dS)
	}
}

func TestNewDispatcherErSv1(t *testing.T) {
	mockDispatcherService := &dispatchers.DispatcherService{}
	dispatcherErSv1 := NewDispatcherErSv1(mockDispatcherService)
	if dispatcherErSv1 == nil {
		t.Fatal("Expected a non-nil DispatcherErSv1, got nil")
	}
	if dispatcherErSv1.dS != mockDispatcherService {
		t.Errorf("Expected DispatcherService to be %v, got %v", mockDispatcherService, dispatcherErSv1.dS)
	}
}
