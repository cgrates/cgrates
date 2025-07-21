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
