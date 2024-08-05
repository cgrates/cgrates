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

package loaders

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func TestLoaderServiceListenAndServe(t *testing.T) {
	stopChan := make(chan struct{})
	defer close(stopChan)
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{
			"loader1": {ldrID: "loader1"},
			"loader2": {ldrID: "loader2"},
			"loader3": {ldrID: "error"},
		},
	}

	loaderService.ldrs["loader3"].ldrID = "loader3"
	t.Run("All loaders succeed", func(t *testing.T) {
		err := loaderService.ListenAndServe(stopChan)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestLoaderServiceV1Load(t *testing.T) {
	ctx := context.Background()
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{},
	}

	t.Run("Unknown loader", func(t *testing.T) {
		args := &ArgsProcessFolder{LoaderID: "unknown"}
		var rply string
		err := loaderService.V1Load(ctx, args, &rply)
		if err == nil || err.Error() != "UNKNOWN_LOADER: unknown" {
			t.Errorf("expected error 'UNKNOWN_LOADER: unknown', got %v", err)
		}
	})

	t.Run("Another loader running without force", func(t *testing.T) {
		args := &ArgsProcessFolder{LoaderID: "loader1"}
		var rply string
		err := loaderService.V1Load(ctx, args, &rply)
		if err == nil || err.Error() == "ANOTHER_LOADER_RUNNING" {
			t.Errorf("expected error 'ANOTHER_LOADER_RUNNING', got %v", err)
		}
	})

	t.Run("Unlock folder and process", func(t *testing.T) {
		args := &ArgsProcessFolder{LoaderID: "loader1", ForceLock: true}
		var rply string
		err := loaderService.V1Load(ctx, args, &rply)
		if err == nil {
			t.Errorf("expected no error, got %v", err)
		}
		if rply == "OK" {
			t.Errorf("expected reply 'OK', got %s", rply)
		}

	})

	t.Run("Process with no locking issues", func(t *testing.T) {
		args := &ArgsProcessFolder{LoaderID: "loader2"}
		var rply string
		err := loaderService.V1Load(ctx, args, &rply)
		if err == nil {
			t.Errorf("expected no error, got %v", err)
		}
		if rply == "OK" {
			t.Errorf("expected reply 'OK', got %s", rply)
		}
	})
}

func TestLoaderServiceReload(t *testing.T) {
	loaderService := &LoaderService{}
	dm := &engine.DataManager{}
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}
	timezone := "UTC"
	cachingDelay := time.Duration(5 * time.Second)
	ldrsCfg := []*config.LoaderSCfg{
		{ID: "loader1", Enabled: true},
		{ID: "loader2", Enabled: false},
		{ID: "loader3", Enabled: true},
	}
	loaderService.Reload(dm, ldrsCfg, timezone, cachingDelay, filterS, connMgr)
	if len(loaderService.ldrs) != 2 {
		t.Errorf("expected 2 enabled loaders, got %d", len(loaderService.ldrs))
	}

	if _, exists := loaderService.ldrs["loader1"]; !exists {
		t.Error("expected loader1 to be in the loaders map")
	}

	if _, exists := loaderService.ldrs["loader3"]; !exists {
		t.Error("expected loader3 to be in the loaders map")
	}

	if _, exists := loaderService.ldrs["loader2"]; exists {
		t.Error("did not expect loader2 to be in the loaders map")
	}
}

func TestLoaderServiceV1Remove(t *testing.T) {
	ctx := context.Background()
	loaderService := &LoaderService{
		ldrs: map[string]*Loader{},
	}

	args := &ArgsProcessFolder{LoaderID: "unknown"}
	var rply string
	err := loaderService.V1Remove(ctx, args, &rply)
	if err == nil || err.Error() != "UNKNOWN_LOADER: unknown" {
		t.Errorf("expected error 'UNKNOWN_LOADER: unknown', got %v", err)
	}
}

func TestNewLoaderService(t *testing.T) {
	dm := &engine.DataManager{}
	timezone := "UTC"
	cachingDlay := time.Second
	filterS := &engine.FilterS{}
	connMgr := &engine.ConnManager{}

	ldrsCfg := []*config.LoaderSCfg{
		{ID: "loader1", Enabled: true},
		{ID: "loader2", Enabled: false},
		{ID: "loader3", Enabled: true},
	}

	ldrService := NewLoaderService(dm, ldrsCfg, timezone, cachingDlay, filterS, connMgr)

	if len(ldrService.ldrs) != 2 {
		t.Errorf("expected 2 loaders, got %d", len(ldrService.ldrs))
	}

	if _, exists := ldrService.ldrs["loader1"]; !exists {
		t.Errorf("expected loader1 to be present")
	}

	if _, exists := ldrService.ldrs["loader2"]; exists {
		t.Errorf("expected loader2 to not be present")
	}

	if _, exists := ldrService.ldrs["loader3"]; !exists {
		t.Errorf("expected loader3 to be present")
	}
}
