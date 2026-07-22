// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
