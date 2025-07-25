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

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestIPProfileClone(t *testing.T) {
	activation := time.Date(2025, 7, 21, 10, 0, 0, 0, time.UTC)
	expiry := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	ip := &IPProfile{
		Tenant:    "cgrates.org",
		ID:        "ip_profile_1",
		FilterIDs: []string{"flt1", "flt2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: activation,
			ExpiryTime:     expiry,
		},
		TTL:         10 * time.Minute,
		Type:        "ipv4",
		AddressPool: "main_pool",
		Allocation:  "dynamic",
		Stored:      true,
		Weight:      2.5,
	}

	cloned := ip.Clone()

	if cloned == ip {
		t.Error("Clone returned the same reference as original")
	}
	if cloned.Tenant != ip.Tenant {
		t.Error("Tenant mismatch")
	}
	if cloned.ID != ip.ID {
		t.Error("ID mismatch")
	}
	if cloned.Type != ip.Type {
		t.Error("Type mismatch")
	}
	if cloned.AddressPool != ip.AddressPool {
		t.Error("AddressPool mismatch")
	}
	if cloned.Allocation != ip.Allocation {
		t.Error("Allocation mismatch")
	}
	if cloned.Stored != ip.Stored {
		t.Error("Stored mismatch")
	}
	if cloned.Weight != ip.Weight {
		t.Error("Weight mismatch")
	}
	if cloned.TTL != ip.TTL {
		t.Error("TTL mismatch")
	}

	if cloned.ActivationInterval == ip.ActivationInterval {
		t.Error("ActivationInterval not deeply cloned")
	}
	if cloned.ActivationInterval.ActivationTime != ip.ActivationInterval.ActivationTime {
		t.Error("ActivationTime mismatch")
	}
	if cloned.ActivationInterval.ExpiryTime != ip.ActivationInterval.ExpiryTime {
		t.Error("ExpiryTime mismatch")
	}

	if len(cloned.FilterIDs) != len(ip.FilterIDs) {
		t.Error("FilterIDs length mismatch")
	}
	for i := range cloned.FilterIDs {
		if cloned.FilterIDs[i] != ip.FilterIDs[i] {
			t.Errorf("FilterIDs[%d] mismatch", i)
		}
	}

	cloned.FilterIDs[0] = "changed"
	if ip.FilterIDs[0] == "changed" {
		t.Error("Original FilterIDs changed after modifying clone")
	}

	var nilIP *IPProfile
	clonedNil := nilIP.Clone()
	if clonedNil != nil {
		t.Error("Clone of nil IPProfile should return nil")
	}
}

func TestIPProfileCacheClone(t *testing.T) {
	ip := &IPProfile{
		Tenant:    "cgrates.org",
		ID:        "ip_cache_test",
		FilterIDs: []string{"fltA"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2025, 7, 21, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		},
		Type:        "ipv6",
		TTL:         5 * time.Minute,
		AddressPool: "test_pool",
		Allocation:  "static",
		Stored:      true,
		Weight:      1.0,
	}

	clonedAny := ip.CacheClone()
	cloned, ok := clonedAny.(*IPProfile)
	if !ok {
		t.Error("CacheClone did not return *IPProfile")
	}
	if cloned == ip {
		t.Error("CacheClone returned the same reference instead of a clone")
	}
	if cloned.ID != ip.ID || cloned.Tenant != ip.Tenant {
		t.Error("Cloned fields do not match original")
	}
	cloned.FilterIDs[0] = "modified"
	if ip.FilterIDs[0] == "modified" {
		t.Error("Original FilterIDs modified")
	}

	var nilIP *IPProfile
	clonedNil := nilIP.CacheClone()
	if clonedNil != nil {
		if _, ok := clonedNil.(*IPProfile); !ok {
			t.Error("CacheClone on nil receiver should return nil")
		}
	}
}

func TestIPProfileLockKey(t *testing.T) {
	tenant := "cgrates.org"
	id := "1001"

	got := ipProfileLockKey(tenant, id)
	want := utils.CacheIPProfiles + ":" + tenant + ":" + id
	if got != want {
		t.Errorf("ipProfileLockKey() = %q; want %q", got, want)
	}
	got = ipProfileLockKey("", "")
	want = utils.CacheIPProfiles + "::"
	if got != want {
		t.Errorf("ipProfileLockKey() with empty strings = %q; want %q", got, want)
	}
}

func TestIPProfilelock(t *testing.T) {
	ip := &IPProfile{
		Tenant: "cgrates.org",
		ID:     "profile123",
	}

	ip.lock("IDlock")
	if ip.lkID != "IDlock" {
		t.Errorf("expected lkID 'IDlock', got %q", ip.lkID)
	}

	ip.lock(utils.EmptyString)
	if ip.lkID == "" {
		t.Error("expected lkID to be set by guardian but got empty string")
	}
}

func TestIPProfileUnlock(t *testing.T) {
	ip := &IPProfile{}
	ip.lkID = utils.EmptyString
	ip.unlock()
	if ip.lkID != utils.EmptyString {
		t.Errorf("expected lkID to remain empty, got %q", ip.lkID)
	}
	ip.lkID = "Id"
	ip.unlock()
	if ip.lkID != utils.EmptyString {
		t.Errorf("expected lkID to be cleared, got %q", ip.lkID)
	}
}

func TestIPUsageTenantID(t *testing.T) {
	u := &IPUsage{
		Tenant: "cgrates.org",
		ID:     "usage01",
	}
	got := u.TenantID()
	want := "cgrates.org:usage01"
	if got != want {
		t.Errorf("TenantID() = %q; want %q", got, want)
	}

	u.Tenant = ""
	u.ID = ""
	got = u.TenantID()
	want = ":"
	if got != want {
		t.Errorf("TenantID() with empty = %q; want %q", got, want)
	}
}

func TestIPUsageIsActive(t *testing.T) {
	now := time.Now()

	u := &IPUsage{
		ExpiryTime: now.Add(1 * time.Hour),
	}
	if !u.isActive(now) {
		t.Errorf("Expected active usage, got inactive")
	}

	u.ExpiryTime = now.Add(-1 * time.Hour)
	if u.isActive(now) {
		t.Errorf("Expected inactive usage, got active")
	}

	u.ExpiryTime = time.Time{}
	if !u.isActive(now) {
		t.Errorf("Expected active usage for zero expiry time, got inactive")
	}
}

func TestIPTotalUsage(t *testing.T) {
	ip := &IP{
		Usages: map[string]*IPUsage{
			"u1": {Units: 1.5},
			"u2": {Units: 2.0},
			"u3": {Units: 0.5},
		},
	}

	expected := 4.0
	got := ip.TotalUsage()
	if got != expected {
		t.Errorf("TotalUsage() = %v, want %v", got, expected)
	}

	gotCached := ip.TotalUsage()
	if gotCached != expected {
		t.Errorf("TotalUsage() after cache = %v, want %v", gotCached, expected)
	}
}

func TestIPUsageClone(t *testing.T) {
	original := &IPUsage{
		Tenant:     "cgrates.org",
		ID:         "ID1001",
		ExpiryTime: time.Now().Add(24 * time.Hour),
	}

	cloned := original.Clone()

	if cloned == nil {
		t.Fatal("expected clone not to be nil")
	}
	if cloned == original {
		t.Error("expected clone to be a different instance")
	}
	if *cloned != *original {
		t.Errorf("expected clone to have same content, got %+v, want %+v", cloned, original)
	}

	var nilUsage *IPUsage
	nilClone := nilUsage.Clone()
	if nilClone != nil {
		t.Error("expected nil clone for nil input")
	}
}

func TestIPClone(t *testing.T) {
	ttl := 5 * time.Minute
	tUsage := 123.45
	dirty := true
	expTime := time.Now().Add(time.Hour)

	original := &IP{
		Tenant: "cgrates.org",
		ID:     "ip01",
		TTLIdx: []string{"idx1", "idx2"},
		cfg: &IPProfile{
			Tenant:    "cgrates.org",
			ID:        "profile1",
			FilterIDs: []string{"f1", "f2"},
			TTL:       time.Minute * 10,
			Type:      "dynamic",
		},
		Usages: map[string]*IPUsage{
			"u1": {
				Tenant:     "cgrates.org",
				ID:         "u1",
				ExpiryTime: expTime,
				Units:      50.0,
			},
		},
		ttl:    &ttl,
		tUsage: &tUsage,
		dirty:  &dirty,
	}

	cloned := original.Clone()

	t.Run("nil IP returns nil", func(t *testing.T) {
		var ip *IP = nil
		cloned := ip.Clone()
		if cloned != nil {
			t.Errorf("expected nil, got %+v", cloned)
		}
	})

	if !reflect.DeepEqual(original, cloned) {
		t.Errorf("Expected clone to be deeply equal to original, got difference:\noriginal: %+v\nclone: %+v", original, cloned)
	}

	if cloned == original {
		t.Error("Clone should return a different pointer than the original")
	}
	if cloned.cfg == original.cfg {
		t.Error("cfg field was not deeply cloned")
	}
	if cloned.Usages["u1"] == original.Usages["u1"] {
		t.Error("Usages map content was not deeply cloned")
	}
	if cloned.ttl == original.ttl {
		t.Error("ttl pointer was not deeply cloned")
	}
	if cloned.tUsage == original.tUsage {
		t.Error("tUsage pointer was not deeply cloned")
	}
	if cloned.dirty == original.dirty {
		t.Error("dirty pointer was not deeply cloned")
	}
}

func TestIpLockKey(t *testing.T) {
	tnt := "cgrates.org"
	id := "192.168.0.1"
	expected := utils.ConcatenatedKey(utils.CacheIPs, tnt, id)

	got := ipLockKey(tnt, id)
	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestIPlock(t *testing.T) {
	ip := &IP{Tenant: "cgrates.org", ID: "1001"}

	ip.lock("customLockID")
	if ip.lkID != "customLockID" {
		t.Errorf("Expected lkID to be 'customLockID', got %s", ip.lkID)
	}

	ip2 := &IP{Tenant: "cgrates.org2", ID: "1002"}
	ip2.lock(utils.EmptyString)
	if ip2.lkID == utils.EmptyString {
		t.Error("Expected lkID to be set by Guardian.GuardIDs, got empty string")
	}
}

func TestIPUunlock(t *testing.T) {
	ip := &IP{lkID: "LockID"}

	ip.unlock()

	if ip.lkID != utils.EmptyString {
		t.Errorf("Expected lkID to be cleared, got %s", ip.lkID)
	}
	ip.unlock()
}

func TestIPRemoveExpiredUnits(t *testing.T) {
	now := time.Now()
	expiredID := "expired"
	activeID := "active"

	ip := &IP{
		ID:     "ip-test",
		Usages: map[string]*IPUsage{},
		TTLIdx: []string{expiredID, activeID},
		tUsage: utils.Float64Pointer(30.0),
	}

	ip.Usages[expiredID] = &IPUsage{
		ID:         expiredID,
		ExpiryTime: now.Add(-10 * time.Minute),
		Units:      10.0,
	}

	ip.Usages[activeID] = &IPUsage{
		ID:         activeID,
		ExpiryTime: now.Add(10 * time.Minute),
		Units:      20.0,
	}

	ip.removeExpiredUnits()

	if _, ok := ip.Usages[expiredID]; ok {
		t.Errorf("Expected expired usage to be removed")
	}
	if _, ok := ip.Usages[activeID]; !ok {
		t.Errorf("Expected active usage to be retained")
	}
	if ip.tUsage != nil {
		t.Errorf("Expected tUsage to be set to nil after recalculation")
	}
	if len(ip.TTLIdx) != 1 || ip.TTLIdx[0] != activeID {
		t.Errorf("Expected TTLIdx to only contain activeID")
	}
}

func TestIPRecordUsage(t *testing.T) {
	t.Run("record new usage with ttl set", func(t *testing.T) {
		ttl := 10 * time.Minute
		ip := &IP{
			Usages: make(map[string]*IPUsage),
			TTLIdx: []string{},
			ttl:    &ttl,
			tUsage: new(float64),
		}
		usage := &IPUsage{Tenant: "cgrates.org", ID: "usage1", Units: 5}
		err := ip.recordUsage(usage)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got, want := len(ip.Usages), 1; got != want {
			t.Fatalf("unexpected number of usages: got %d, want %d", got, want)
		}
		if _, ok := ip.Usages[usage.ID]; !ok {
			t.Fatal("usage not recorded")
		}
		if len(ip.TTLIdx) != 1 || ip.TTLIdx[0] != usage.ID {
			t.Fatal("TTLIdx not updated properly")
		}
		if ip.tUsage == nil || *ip.tUsage != usage.Units {
			t.Fatalf("tUsage not updated properly, got %v", ip.tUsage)
		}
	})

	t.Run("duplicate usage id", func(t *testing.T) {
		ip := &IP{
			Usages: map[string]*IPUsage{
				"usage1": {Tenant: "cgrates.org", ID: "usage1", Units: 5},
			},
		}
		usage := &IPUsage{Tenant: "cgrates.org", ID: "usage1", Units: 10}
		err := ip.recordUsage(usage)
		if err == nil {
			t.Fatal("expected error on duplicate usage id, got nil")
		}
	})

	t.Run("ttl zero disables expiry setting", func(t *testing.T) {
		zeroTTL := time.Duration(0)
		ip := &IP{
			Usages: make(map[string]*IPUsage),
			ttl:    &zeroTTL,
		}
		usage := &IPUsage{Tenant: "cgrates.org", ID: "noexpiry", Units: 2}
		err := ip.recordUsage(usage)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(ip.Usages) != 0 {
			t.Fatal("usage should NOT be recorded when ttl is 0")
		}
	})
}
