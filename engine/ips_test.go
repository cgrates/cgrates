// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"net/netip"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestIPsReload(t *testing.T) {
	synctest.Test(t, func(*testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 5 * time.Millisecond
		s := NewIPService(nil, cfg, nil, nil)
		s.StartLoop()
		s.Reload()
		s.Shutdown()
		s.Shutdown()
		s.Reload()
	})
}

func TestIPsReloadShutdownConcurrent(t *testing.T) {
	synctest.Test(t, func(*testing.T) {
		cfg := config.NewDefaultCGRConfig()
		cfg.IPsCfg().StoreInterval = 5 * time.Millisecond
		s := NewIPService(nil, cfg, nil, nil)
		s.StartLoop()
		var wg sync.WaitGroup
		wg.Go(func() { s.Reload() })
		wg.Go(func() { s.Shutdown() })
		wg.Wait()
	})
}

func TestIPsStartLoop(t *testing.T) {
	synctest.Test(t, func(*testing.T) {
		cfg := config.NewDefaultCGRConfig()
		s := NewIPService(nil, cfg, nil, nil)
		s.StartLoop()
		s.backupLoop.Wait()
	})
}

func TestStoreIPAllocationsList(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	data, _ := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	s := NewIPService(dm, cfg, nil, nil)

	exp := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "alloc1",
		Allocations: map[string]*PoolAllocation{
			"alloc1": {
				PoolID:  "pool1",
				Address: netip.MustParseAddr("192.168.1.10"),
				Time:    time.Now(),
			},
		},
	}
	Cache.SetWithoutReplicate(utils.CacheIPAllocations, "cgrates.org:alloc1", exp, nil, true,
		utils.NonTransactional)
	s.storedIPs.Add("cgrates.org:alloc1")
	s.storeIPAllocationsList()

	rcv, err := s.dm.GetIPAllocations("cgrates.org", "alloc1", true, false,
		utils.NonTransactional, nil)
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

	Cache.Remove(utils.CacheIPAllocations, "cgrates.org:alloc1", true, utils.NonTransactional)
}

func newTestIPAllocations(t *testing.T) *IPAllocations {
	t.Helper()
	allocs := &IPAllocations{
		Tenant:      "cgrates.org",
		ID:          "IP1",
		Allocations: map[string]*PoolAllocation{},
	}
	profile := &IPProfile{
		Tenant: "cgrates.org",
		ID:     "IP1",
		Pools: []*IPPool{
			{ID: "pool1", Range: "10.0.0.1/32", Message: "ok"},
			{ID: "pool2", Range: "10.0.0.2/32"},
		},
	}
	if err := allocs.computeUnexported(profile); err != nil {
		t.Fatalf("computeUnexported: %v", err)
	}
	return allocs
}

func TestIPAllocationsAllocateIPOnPoolTTLIndex(t *testing.T) {
	profile := &IPProfile{
		ID:  "IP1",
		TTL: time.Minute,
		Pools: []*IPPool{
			{ID: "pool1", Range: "10.0.0.1/32"},
			{ID: "pool2", Range: "10.0.0.2/32"},
		},
	}
	allocs := &IPAllocations{ID: "IP1", Allocations: map[string]*PoolAllocation{}}
	if err := allocs.computeUnexported(profile); err != nil {
		t.Fatal(err)
	}
	if _, err := allocs.allocateIPOnPool("a1", profile.Pools[0], false); err != nil {
		t.Fatal(err)
	}
	if _, err := allocs.allocateIPOnPool("a2", profile.Pools[1], false); err != nil {
		t.Fatal(err)
	}
	if len(allocs.TTLIndex) != 2 || allocs.TTLIndex[0] != "a1" || allocs.TTLIndex[1] != "a2" {
		t.Fatalf("new allocations not indexed: TTLIndex = %v, want [a1 a2]", allocs.TTLIndex)
	}

	// refreshing a1 moves it to the back of the index
	if _, err := allocs.allocateIPOnPool("a1", profile.Pools[0], false); err != nil {
		t.Fatal(err)
	}
	if len(allocs.TTLIndex) != 2 || allocs.TTLIndex[0] != "a2" || allocs.TTLIndex[1] != "a1" {
		t.Errorf("refresh did not reorder TTLIndex: %v, want [a2 a1]", allocs.TTLIndex)
	}
}

func TestIPAllocationsAllocateIPOnPoolNoTTL(t *testing.T) {
	allocs := newTestIPAllocations(t) // profile has no TTL
	pool := allocs.prfl.Pools[0]
	if _, err := allocs.allocateIPOnPool("a1", pool, false); err != nil {
		t.Fatal(err)
	}
	if _, err := allocs.allocateIPOnPool("a1", pool, false); err != nil { // refresh
		t.Fatal(err)
	}
	if len(allocs.TTLIndex) != 0 {
		t.Fatalf("TTLIndex should stay empty without a TTL, got %v", allocs.TTLIndex)
	}

	// a stray index entry would make the next allocate wrongly expire a1
	if _, err := allocs.allocateIPOnPool("a1", pool, false); err != nil {
		t.Fatal(err)
	}
	if _, has := allocs.Allocations["a1"]; !has {
		t.Error("a1 wrongly expired without a TTL")
	}
}
