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
