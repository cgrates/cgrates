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

func TestFilterAndSortPools(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dm := engine.NewDataManager(db, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	ctx := context.Background()
	tenant := "cgrates.org"

	t.Run("EmptyPools", func(t *testing.T) {
		pools := []*utils.IPPool{}
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err != utils.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got: %v", result)
		}
	})

	t.Run("NilPools", func(t *testing.T) {
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, nil, fltrs, ev)

		if err != utils.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got: %v", result)
		}
	})

	t.Run("SinglePoolNoFilters", func(t *testing.T) {
		pools := []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{},
				Weights:   utils.DynamicWeights{},
				Blockers:  utils.DynamicBlockers{},
			},
		}
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("Expected 1 pool, got: %d", len(result))
		}
		if result[0] != "POOL1" {
			t.Errorf("Expected POOL1, got: %s", result[0])
		}
	})

	t.Run("MultiplePoolsSortedByWeight", func(t *testing.T) {
		pools := []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{},
				Weights:   utils.DynamicWeights{{Weight: 10.0}},
				Blockers:  utils.DynamicBlockers{},
			},
			{
				ID:        "POOL2",
				FilterIDs: []string{},
				Weights:   utils.DynamicWeights{{Weight: 20.0}},
				Blockers:  utils.DynamicBlockers{},
			},
			{
				ID:        "POOL3",
				FilterIDs: []string{},
				Weights:   utils.DynamicWeights{{Weight: 15.0}},
				Blockers:  utils.DynamicBlockers{},
			},
		}
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("Expected 3 pools, got: %d", len(result))
		}
		if result[0] != "POOL2" {
			t.Errorf("Expected POOL2 first, got: %s", result[0])
		}
		if result[1] != "POOL3" {
			t.Errorf("Expected POOL3 second, got: %s", result[1])
		}
		if result[2] != "POOL1" {
			t.Errorf("Expected POOL1 third, got: %s", result[2])
		}
	})

	t.Run("PoolsFilteredOut", func(t *testing.T) {
		filter := &engine.Filter{
			Tenant: tenant,
			ID:     "FltrNoMatch",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		}
		if err := dm.SetFilter(ctx, filter, true); err != nil {
			t.Fatal(err)
		}

		pools := []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{"FltrNoMatch"},
				Weights:   utils.DynamicWeights{},
				Blockers:  utils.DynamicBlockers{},
			},
		}
		ev := utils.MapStorage{
			utils.MetaReq: map[string]interface{}{
				"Account": "1002",
			},
		}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err != utils.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got: %v", result)
		}
	})

	t.Run("PoolsPass", func(t *testing.T) {
		filter := &engine.Filter{
			Tenant: tenant,
			ID:     "FltrMatch",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		}
		if err := dm.SetFilter(ctx, filter, true); err != nil {
			t.Fatal(err)
		}

		pools := []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{"FltrMatch"},
				Weights:   utils.DynamicWeights{{Weight: 10.0}},
				Blockers:  utils.DynamicBlockers{},
			},
			{
				ID:        "POOL2",
				FilterIDs: []string{},
				Weights:   utils.DynamicWeights{{Weight: 5.0}},
				Blockers:  utils.DynamicBlockers{},
			},
		}
		ev := utils.MapStorage{
			utils.MetaReq: map[string]interface{}{
				"Account": "1001",
			},
		}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 pools, got: %d", len(result))
		}
		if result[0] != "POOL1" {
			t.Errorf("Expected POOL1 first, got: %s", result[0])
		}
		if result[1] != "POOL2" {
			t.Errorf("Expected POOL2 second, got: %s", result[1])
		}
	})

	t.Run("FilterPassError", func(t *testing.T) {
		pools := []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{"NonExistentFilter"},
				Weights:   utils.DynamicWeights{},
				Blockers:  utils.DynamicBlockers{},
			},
		}
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err == nil {
			t.Error("Expected error for non-existent filter, got nil")
		}
		if result != nil {
			t.Errorf("Expected nil result on error, got: %v", result)
		}
	})

	t.Run("WeightFromDynamicsError", func(t *testing.T) {
		pools := []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{},
				Weights: utils.DynamicWeights{
					{FilterIDs: []string{"NonExistentWeightFilter"}, Weight: 10.0},
				},
				Blockers: utils.DynamicBlockers{},
			},
		}
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err == nil {
			t.Error("Expected error for weight calculation, got nil")
		}
		if result != nil {
			t.Errorf("Expected nil result on error, got: %v", result)
		}
	})

	t.Run("BlockerFromDynamicsError", func(t *testing.T) {
		pools := []*utils.IPPool{
			{
				ID:        "POOL1",
				FilterIDs: []string{},
				Weights:   utils.DynamicWeights{{Weight: 10.0}},
				Blockers: utils.DynamicBlockers{
					{FilterIDs: []string{"NonExistentBlockerFilter"}, Blocker: true},
				},
			},
		}
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, pools, fltrs, ev)

		if err == nil {
			t.Error("Expected error for blocker, got nil")
		}
		if result != nil {
			t.Errorf("Expected nil result on error, got: %v", result)
		}
	})
}

func TestFindPoolByID(t *testing.T) {
	pools := []*utils.IPPool{
		{ID: "pool1", Type: "*ipv4"},
		{ID: "pool2", Type: "*ipv4"},
		{ID: "pool3", Type: "*ipv4"},
	}

	result := findPoolByID(pools, "pool2")
	if result == nil {
		t.Errorf("expected non-nil result for existing ID")
	} else if result.ID != "pool2" {
		t.Errorf("expected ID 'pool2', got '%s'", result.ID)
	}

	result = findPoolByID(pools, "notfound")
	if result != nil {
		t.Errorf("expected nil for non-existing ID, got %+v", result)
	}

	result = findPoolByID([]*utils.IPPool{}, "pool1")
	if result != nil {
		t.Errorf("expected nil for empty slice, got %+v", result)
	}

	var nilPools []*utils.IPPool
	result = findPoolByID(nilPools, "pool1")
	if result != nil {
		t.Errorf("expected nil for nil slice, got %+v", result)
	}
}
