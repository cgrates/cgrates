/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ips

import (
	"errors"
	"net/netip"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestStoreMatchedIPAllocations(t *testing.T) {
	t.Run("StoreInterval is zero", func(t *testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 0

		s := &IPs{
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

		s := &IPs{
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

		s := &IPs{
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

		s := &IPs{
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
		dm := engine.NewDataManager(engine.NewDBConnManager(map[string]engine.DataDB{}, &config.DbCfg{}), cfg, nil)
		dm.SetCache(engine.Cache)

		s := &IPs{
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

		db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
		if err != nil {
			t.Fatal(err)
		}
		dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
		dm := engine.NewDataManager(dbCM, cfg, nil)
		dm.SetCache(engine.Cache)

		s := &IPs{
			cfg:       cfg,
			dm:        dm,
			cache:     engine.Cache,
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

		s := &IPs{
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
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(engine.Cache)
	filters := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(engine.Cache)

	svc := NewIPService(cfg, dm, engine.Cache, filters, connMgr)

	if svc == nil {
		t.Fatalf("expected non-nil IPs")
	}
	if svc.dm != dm {
		t.Errorf("expected dm to be set, got %+v", svc.dm)
	}
	if svc.cache != engine.Cache {
		t.Errorf("expected cache to be set, got %+v", svc.cache)
	}
	if svc.cfg != cfg {
		t.Errorf("expected cfg to be set, got %+v", svc.cfg)
	}
	if svc.filters != filters {
		t.Errorf("expected filters to be set, got %+v", svc.filters)
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
}

func TestFilterAndSortPools(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	db, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: db}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(engine.Cache)
	filters := engine.NewFilterS(cfg, nil, dm)
	ctx := context.Background()
	tenant := "cgrates.org"

	t.Run("EmptyPools", func(t *testing.T) {
		pools := []*utils.IPPool{}
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

		if err != utils.ErrNotFound {
			t.Errorf("Expected ErrNotFound, got: %v", err)
		}
		if result != nil {
			t.Errorf("Expected nil result, got: %v", result)
		}
	})

	t.Run("NilPools", func(t *testing.T) {
		ev := utils.MapStorage{}

		result, err := filterAndSortPools(ctx, tenant, nil, filters, ev)

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

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

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

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

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

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

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

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

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

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

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

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

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

		result, err := filterAndSortPools(ctx, tenant, pools, filters, ev)

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

func TestIPsReload(t *testing.T) {
	synctest.Test(t, func(*testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 5 * time.Millisecond
		s := NewIPService(cfg, nil, nil, nil, nil)
		s.StartLoop(context.Background())
		s.Reload(context.Background())
		s.Shutdown(context.Background())
		s.Shutdown(context.Background())
		s.Reload(context.Background())
	})
}

func TestIPsReloadShutdownConcurrent(t *testing.T) {
	synctest.Test(t, func(*testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 5 * time.Millisecond
		s := NewIPService(cfg, nil, nil, nil, nil)
		s.StartLoop(context.Background())
		var wg sync.WaitGroup
		wg.Go(func() { s.Reload(context.Background()) })
		wg.Go(func() { s.Shutdown(context.Background()) })
		wg.Wait()
	})
}

func TestIPsStartLoop(t *testing.T) {
	synctest.Test(t, func(*testing.T) {
		cfg := config.NewDefaultCGRConfig()
		s := NewIPService(cfg, nil, nil, nil, nil)
		s.StartLoop(context.Background())
		s.backupLoop.Wait()
	})
}

func TestStoreIPAllocationsList(t *testing.T) {
	tmp := engine.Cache
	defer func() {
		engine.Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	dm.SetCache(engine.Cache)
	s := NewIPService(cfg, dm, engine.Cache, nil, nil)

	exp := &utils.IPAllocations{
		Tenant: "cgrates.org",
		ID:     "alloc1",
		Allocations: map[string]*utils.PoolAllocation{
			"alloc1": {
				PoolID:  "pool1",
				Address: netip.MustParseAddr("192.168.1.10"),
				Time:    time.Now(),
			},
		},
	}
	engine.Cache.SetWithoutReplicate(utils.CacheIPAllocations, "cgrates.org:alloc1", exp, nil, true,
		utils.NonTransactional)
	s.storedIPs.Add("cgrates.org:alloc1")
	s.storeIPAllocationsList(context.Background())

	rcv, err := s.dm.GetIPAllocations(context.Background(), "cgrates.org", "alloc1", true, false,
		utils.NonTransactional)
	if err != nil {
		t.Fatal(err)
	}
	if rcv.ID != exp.ID {
		t.Errorf("expected IPAllocations ID %q, got %q", exp.ID, rcv.ID)
	}
	if pa, has := rcv.Allocations["alloc1"]; !has {
		t.Errorf("expected PoolAllocation %q to exist", "alloc1")
	} else if pa.Address.String() != "192.168.1.10" {
		t.Errorf("expected address 192.168.1.10, got %s", pa.Address.String())
	}

	engine.Cache.Remove(context.Background(), utils.CacheIPAllocations, "cgrates.org:alloc1", true, utils.NonTransactional)
}

func newTestMatchedIPAllocs(t *testing.T) *matchedIPAllocs {
	t.Helper()
	allocs := &utils.IPAllocations{
		Tenant:      "cgrates.org",
		ID:          "IP1",
		Allocations: map[string]*utils.PoolAllocation{},
	}
	profile := &utils.IPProfile{
		Tenant: "cgrates.org",
		ID:     "IP1",
		Pools: []*utils.IPPool{
			{ID: "pool1", Range: "10.0.0.1/32", Message: "ok"},
			{ID: "pool2", Range: "10.0.0.2/32"},
		},
	}
	m, err := newMatchedIPAllocs(allocs, profile)
	if err != nil {
		t.Fatalf("newMatchedIPAllocs: %v", err)
	}
	return m
}

func TestNewMatchedIPAllocs(t *testing.T) {
	allocs := &utils.IPAllocations{
		Tenant: "cgrates.org",
		ID:     "IP1",
		Allocations: map[string]*utils.PoolAllocation{
			"a1": {PoolID: "pool1", Address: netip.MustParseAddr("10.0.0.1")},
		},
	}
	profile := &utils.IPProfile{
		ID:    "IP1",
		Pools: []*utils.IPPool{{ID: "pool1", Range: "10.0.0.1/32"}},
	}
	m, err := newMatchedIPAllocs(allocs, profile)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.poolRanges["pool1"]; got.String() != "10.0.0.1/32" {
		t.Errorf("poolRanges[pool1] = %v", got)
	}
	if got := m.poolAllocs["pool1"][netip.MustParseAddr("10.0.0.1")]; got != "a1" {
		t.Errorf("poolAllocs reverse index = %q, want a1", got)
	}

	bad := &utils.IPProfile{ID: "IP1", Pools: []*utils.IPPool{{ID: "pool1", Range: "not-a-cidr"}}}
	if _, err := newMatchedIPAllocs(allocs, bad); err == nil {
		t.Error("expected error for invalid pool range")
	}
}

func TestMatchedIPAllocsAllocateIPOnPool(t *testing.T) {
	m := newTestMatchedIPAllocs(t)
	pool := m.profile.Pools[0] // pool1, 10.0.0.1/32

	ip, err := m.allocateIPOnPool("a1", pool, true) // dry run must not mutate
	if err != nil {
		t.Fatal(err)
	}
	if ip.Address.String() != "10.0.0.1" {
		t.Errorf("address = %s", ip.Address)
	}
	if len(m.allocs.Allocations) != 0 {
		t.Errorf("dry run mutated allocations: %v", m.allocs.Allocations)
	}

	if _, err = m.allocateIPOnPool("a1", pool, false); err != nil {
		t.Fatal(err)
	}
	if _, has := m.allocs.Allocations["a1"]; !has {
		t.Error("allocation a1 not recorded")
	}
	if m.poolAllocs["pool1"][netip.MustParseAddr("10.0.0.1")] != "a1" {
		t.Error("reverse index not updated")
	}

	if _, err = m.allocateIPOnPool("a2", pool, false); !errors.Is(err, utils.ErrIPAlreadyAllocated) {
		t.Errorf("expected ErrIPAlreadyAllocated, got %v", err)
	}

	// refreshing an existing allocation reuses its record and IP
	prev := m.allocs.Allocations["a1"]
	if ip, err = m.allocateIPOnPool("a1", pool, false); err != nil {
		t.Fatal(err)
	}
	if ip.Address.String() != "10.0.0.1" {
		t.Errorf("refresh address = %s", ip.Address)
	}
	if m.allocs.Allocations["a1"] != prev {
		t.Error("refresh should reuse the existing allocation record")
	}
	if len(m.allocs.Allocations) != 1 {
		t.Errorf("refresh created a new record, got %d allocations", len(m.allocs.Allocations))
	}
}

func TestMatchedIPAllocsAllocateIPOnPoolNonSingleIP(t *testing.T) {
	allocs := &utils.IPAllocations{ID: "IP1", Allocations: map[string]*utils.PoolAllocation{}}
	profile := &utils.IPProfile{ID: "IP1", Pools: []*utils.IPPool{{ID: "pool1", Range: "10.0.0.0/24"}}}
	m, err := newMatchedIPAllocs(allocs, profile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.allocateIPOnPool("a1", profile.Pools[0], false); err == nil {
		t.Error("expected error for non single IP pool")
	}
}

func TestMatchedIPAllocsReleaseAllocation(t *testing.T) {
	m := newTestMatchedIPAllocs(t)
	if _, err := m.allocateIPOnPool("a1", m.profile.Pools[0], false); err != nil {
		t.Fatal(err)
	}
	if err := m.releaseAllocation("a1"); err != nil {
		t.Fatal(err)
	}
	if _, has := m.allocs.Allocations["a1"]; has {
		t.Error("allocation not released")
	}
	if err := m.releaseAllocation("missing"); err == nil {
		t.Error("expected error releasing unknown allocation")
	}
}

func TestMatchedIPAllocsClearAllocations(t *testing.T) {
	m := newTestMatchedIPAllocs(t)
	if _, err := m.allocateIPOnPool("a1", m.profile.Pools[0], false); err != nil {
		t.Fatal(err)
	}
	if _, err := m.allocateIPOnPool("a2", m.profile.Pools[1], false); err != nil {
		t.Fatal(err)
	}

	if err := m.clearAllocations([]string{"a1", "missing"}); err == nil {
		t.Error("expected error for unknown id")
	}
	if len(m.allocs.Allocations) != 2 {
		t.Errorf("nothing should have been cleared, got %d", len(m.allocs.Allocations))
	}

	if err := m.clearAllocations([]string{"a1"}); err != nil {
		t.Fatal(err)
	}
	if _, has := m.allocs.Allocations["a1"]; has {
		t.Error("a1 not cleared")
	}

	if err := m.clearAllocations(nil); err != nil {
		t.Fatal(err)
	}
	if len(m.allocs.Allocations) != 0 {
		t.Errorf("clear all left %d allocations", len(m.allocs.Allocations))
	}
}

func TestMatchedIPAllocsRemoveExpiredUnits(t *testing.T) {
	profile := &utils.IPProfile{
		ID:    "IP1",
		TTL:   time.Minute,
		Pools: []*utils.IPPool{{ID: "pool1", Range: "10.0.0.1/32"}},
	}
	allocs := &utils.IPAllocations{
		ID: "IP1",
		Allocations: map[string]*utils.PoolAllocation{
			"expired": {PoolID: "pool1", Address: netip.MustParseAddr("10.0.0.1"), Time: time.Now().Add(-2 * time.Minute)},
			"active":  {PoolID: "pool1", Address: netip.MustParseAddr("10.0.0.2"), Time: time.Now()},
		},
		TTLIndex: []string{"expired", "active"},
	}
	m, err := newMatchedIPAllocs(allocs, profile)
	if err != nil {
		t.Fatal(err)
	}
	m.removeExpiredUnits()
	if _, has := m.allocs.Allocations["expired"]; has {
		t.Error("expired allocation not removed")
	}
	if _, has := m.allocs.Allocations["active"]; !has {
		t.Error("active allocation wrongly removed")
	}
	if len(m.allocs.TTLIndex) != 1 || m.allocs.TTLIndex[0] != "active" {
		t.Errorf("TTLIndex = %v, want [active]", m.allocs.TTLIndex)
	}
}

func TestMatchedIPAllocsAllocateIPOnPoolTTLIndex(t *testing.T) {
	profile := &utils.IPProfile{
		ID:  "IP1",
		TTL: time.Minute,
		Pools: []*utils.IPPool{
			{ID: "pool1", Range: "10.0.0.1/32"},
			{ID: "pool2", Range: "10.0.0.2/32"},
		},
	}
	allocs := &utils.IPAllocations{ID: "IP1", Allocations: map[string]*utils.PoolAllocation{}}
	m, err := newMatchedIPAllocs(allocs, profile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.allocateIPOnPool("a1", profile.Pools[0], false); err != nil {
		t.Fatal(err)
	}
	if _, err := m.allocateIPOnPool("a2", profile.Pools[1], false); err != nil {
		t.Fatal(err)
	}
	if len(m.allocs.TTLIndex) != 2 || m.allocs.TTLIndex[0] != "a1" || m.allocs.TTLIndex[1] != "a2" {
		t.Fatalf("new allocations not indexed: TTLIndex = %v, want [a1 a2]", m.allocs.TTLIndex)
	}

	// refreshing a1 moves it to the back of the index
	if _, err := m.allocateIPOnPool("a1", profile.Pools[0], false); err != nil {
		t.Fatal(err)
	}
	if len(m.allocs.TTLIndex) != 2 || m.allocs.TTLIndex[0] != "a2" || m.allocs.TTLIndex[1] != "a1" {
		t.Errorf("refresh did not reorder TTLIndex: %v, want [a2 a1]", m.allocs.TTLIndex)
	}
}

func TestMatchedIPAllocsAllocateIPOnPoolNoTTL(t *testing.T) {
	m := newTestMatchedIPAllocs(t) // profile has no TTL
	pool := m.profile.Pools[0]
	if _, err := m.allocateIPOnPool("a1", pool, false); err != nil {
		t.Fatal(err)
	}
	if _, err := m.allocateIPOnPool("a1", pool, false); err != nil { // refresh
		t.Fatal(err)
	}
	if len(m.allocs.TTLIndex) != 0 {
		t.Fatalf("TTLIndex should stay empty without a TTL, got %v", m.allocs.TTLIndex)
	}

	// a stray index entry would make the next allocate wrongly expire a1
	if _, err := m.allocateIPOnPool("a1", pool, false); err != nil {
		t.Fatal(err)
	}
	if _, has := m.allocs.Allocations["a1"]; !has {
		t.Error("a1 wrongly expired without a TTL")
	}
}

func TestIPsV1ReleaseIPNotFound(t *testing.T) {
	tmp := engine.Cache
	t.Cleanup(func() { engine.Cache = tmp })
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	engine.Cache = engine.NewCacheS(cfg, dm, nil, nil)
	dm.SetCache(engine.Cache)
	filters := engine.NewFilterS(cfg, nil, dm)
	s := NewIPService(cfg, dm, engine.Cache, filters, nil)

	profile := &utils.IPProfile{
		Tenant:    "cgrates.org",
		ID:        "IP1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights:   utils.DynamicWeights{{Weight: 10}},
		Pools:     []*utils.IPPool{{ID: "pool1", Range: "10.0.0.1/32"}},
	}
	if err := dm.SetIPProfile(context.Background(), profile, true); err != nil {
		t.Fatal(err)
	}

	allocArgs := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EventAllocateIP",
		Event:   map[string]any{utils.AccountField: "1001"},
		APIOpts: map[string]any{utils.OptsIPsAllocationID: "alloc1"},
	}
	var allocReply utils.AllocatedIP
	if err := s.V1AllocateIP(context.Background(), allocArgs, &allocReply); err != nil {
		t.Fatal(err)
	}

	releaseArgs := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EventReleaseIP",
		Event:   map[string]any{utils.AccountField: "1001"},
		APIOpts: map[string]any{utils.OptsIPsAllocationID: "alloc2"},
	}
	var reply string
	experr := "cannot find allocation record with id: alloc2"
	if err := s.V1ReleaseIP(context.Background(), releaseArgs, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%v>, received: <%v>", experr, err)
	}
}
