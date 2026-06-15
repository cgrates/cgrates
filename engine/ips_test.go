// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"testing"
	"time"
)

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
