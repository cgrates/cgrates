// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
