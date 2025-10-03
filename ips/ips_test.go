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

package ips

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestStoreMatchedIPAllocations(t *testing.T) {
	t.Run("StoreInterval is zero", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 0

		s := &IPService{
			cfg:       cfg,
			storedIPs: make(utils.StringSet),
		}

		matched := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC1",
		}

		ctx := context.Background()
		err := s.storeMatchedIPAllocations(ctx, matched)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(s.storedIPs) != 0 {
			t.Errorf("Expected storedIPs to be empty, got length: %d", len(s.storedIPs))
		}
	})

	t.Run("StoreInterval is positive, single allocation", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 10

		s := &IPService{
			cfg:       cfg,
			storedIPs: make(utils.StringSet),
		}

		matched := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC1",
		}

		ctx := context.Background()
		err := s.storeMatchedIPAllocations(ctx, matched)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		expectedTenantID := matched.TenantID()
		if !s.storedIPs.Has(expectedTenantID) {
			t.Errorf("Expected storedIPs to contain %q", expectedTenantID)
		}

		if len(s.storedIPs) != 1 {
			t.Errorf("Expected storedIPs length to be 1, got: %d", len(s.storedIPs))
		}
	})

	t.Run("StoreInterval is positive, multiple allocations", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 10

		s := &IPService{
			cfg:       cfg,
			storedIPs: make(utils.StringSet),
		}

		matched1 := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC1",
		}

		matched2 := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC2",
		}

		ctx := context.Background()

		if err := s.storeMatchedIPAllocations(ctx, matched1); err != nil {
			t.Errorf("Expected no error for first allocation, got: %v", err)
		}

		if err := s.storeMatchedIPAllocations(ctx, matched2); err != nil {
			t.Errorf("Expected no error for second allocation, got: %v", err)
		}

		expectedTenantID := matched1.TenantID()
		if !s.storedIPs.Has(expectedTenantID) {
			t.Errorf("Expected storedIPs to contain %q", expectedTenantID)
		}
	})

	t.Run("StoreInterval is negative, no DataManager", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = -1

		s := &IPService{
			cfg:       cfg,
			dm:        nil,
			storedIPs: make(utils.StringSet),
		}

		matched := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC1",
		}

		ctx := context.Background()
		err := s.storeMatchedIPAllocations(ctx, matched)

		if err == nil {
			t.Error("Expected error when DataManager is nil, got nil")
		}

		if err != utils.ErrNoDatabaseConn {
			t.Errorf("Expected error %v, got: %v", utils.ErrNoDatabaseConn, err)
		}
	})

	t.Run("StoreInterval is negative, with DataManager no DB", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = -1

		dm := engine.NewDataManager(nil, cfg, nil)

		s := &IPService{
			cfg:       cfg,
			dm:        dm,
			storedIPs: make(utils.StringSet),
		}

		matched := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC1",
		}

		ctx := context.Background()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic due to nil DataDB, got none")
			}
		}()

		_ = s.storeMatchedIPAllocations(ctx, matched)
	})

	t.Run("StoreInterval is negative, with working DataManager", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = -1

		db, err := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
		if err != nil {
			t.Fatal(err)
		}
		dm := engine.NewDataManager(db, cfg, nil)

		s := &IPService{
			cfg:       cfg,
			dm:        dm,
			storedIPs: make(utils.StringSet),
		}

		matched := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC1",
		}

		ctx := context.Background()
		err = s.storeMatchedIPAllocations(ctx, matched)

		if err != nil {
			t.Errorf("Expected no error with working DataManager, got: %v", err)
		}
	})

	t.Run("Nil context with StoreInterval zero", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 0

		s := &IPService{
			cfg:       cfg,
			storedIPs: make(utils.StringSet),
		}

		matched := &utils.IPAllocations{
			Tenant: "cgrates.org",
			ID:     "ALLOC1",
		}

		err := s.storeMatchedIPAllocations(nil, matched)

		if err != nil {
			t.Errorf("Expected no error with nil context for StoreInterval=0, got: %v", err)
		}
	})
}

func TestNewIPService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)

	svc := NewIPService(dm, cfg, fltrs, connMgr)

	if svc == nil {
		t.Fatalf("expected non-nil IPService")
	}
	if svc.dm != dm {
		t.Errorf("expected dm to be set, got %+v", svc.dm)
	}
	if svc.cfg != cfg {
		t.Errorf("expected cfg to be set, got %+v", svc.cfg)
	}
	if svc.fltrs != fltrs {
		t.Errorf("expected fltrs to be set, got %+v", svc.fltrs)
	}
	if svc.cm != connMgr {
		t.Errorf("expected connMgr to be set, got %+v", svc.cm)
	}
	if svc.storedIPs == nil {
		t.Errorf("expected storedIPs initialized")
	}
	if svc.stopBackup == nil {
		t.Errorf("expected stopBackup channel initialized")
	}
	if svc.loopStopped == nil {
		t.Errorf("expected loopStopped channel initialized")
	}
}
