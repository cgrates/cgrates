// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"sync"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestNewGlobalVarS(t *testing.T) {
	cfg := &config.CGRConfig{}
	srvDep := make(map[string]*sync.WaitGroup)
	srvDep["service1"] = &sync.WaitGroup{}
	result := NewGlobalVarS(cfg, srvDep)
	if result == nil {
		t.Fatalf("Expected non-nil result, got nil")
	}
	globalVarS, ok := result.(*GlobalVarS)
	if !ok {
		t.Fatalf("Expected result to be of type *GlobalVarS, got %T", result)
	}
	if globalVarS.cfg != cfg {
		t.Errorf("Expected cfg to be %v, got %v", cfg, globalVarS.cfg)
	}

}

func TestGlobalVarSshouldRun(t *testing.T) {
	gv := &GlobalVarS{}
	result := gv.ShouldRun()
	if !result {
		t.Errorf("Expected ShouldRun to return true, but got %v", result)
	}
}

func TestGlobalVarSServiceName(t *testing.T) {
	gv := &GlobalVarS{}
	result := gv.ServiceName()
	expected := utils.GlobalVarS
	if result != expected {
		t.Errorf("Expected ServiceName to return %s, but got %s", expected, result)
	}
}

func TestGlobalVarSIsRunning(t *testing.T) {
	gv := &GlobalVarS{}
	result := gv.IsRunning()
	expected := true
	if result != expected {
		t.Errorf("Expected IsRunning to return %v, but got %v", expected, result)
	}
}

func TestGlobalVarSShutdown(t *testing.T) {
	gv := &GlobalVarS{}
	err := gv.Shutdown()
	if err != nil {
		t.Errorf("Expected Shutdown to return nil, but got %v", err)
	}
}
