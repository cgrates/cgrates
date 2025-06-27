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

package utils

import (
	"fmt"
	"net/netip"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestIPProfileTenantID(t *testing.T) {
	p := &IPProfile{
		Tenant: "cgrates.org",
		ID:     "1001",
	}

	expected := "cgrates.org:1001"
	got := p.TenantID()
	if got != expected {
		t.Errorf("TenantID() = %q; want %q", got, expected)
	}
}

func TestIPProfileClone(t *testing.T) {
	var p *IPProfile = nil
	clone := p.Clone()
	if clone != nil {
		t.Errorf("Clone() with nil receiver: got %v, want nil", clone)
	}

	orig := &IPProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		FilterIDs: []string{"f1", "f2"},
		Weights: DynamicWeights{
			&DynamicWeight{
				FilterIDs: []string{"w1"},
				Weight:    10.5,
			},
		},
		TTL:    5 * time.Minute,
		Stored: true,
		Pools: []*IPPool{
			{
				ID:        "pool1",
				FilterIDs: []string{"pf1"},
				Type:      "type1",
				Range:     "range1",
				Strategy:  "strat1",
				Message:   "msg1",
				Weights: DynamicWeights{
					&DynamicWeight{
						FilterIDs: []string{"pw1"},
						Weight:    7.5,
					},
				},
				Blockers: DynamicBlockers{
					&DynamicBlocker{
						FilterIDs: []string{"pb1"},
						Blocker:   true,
					},
				},
			},
		},
	}

	clone = orig.Clone()
	if clone == nil {
		t.Errorf("Clone() returned nil, want non-nil")
		return
	}

	if clone.Tenant != orig.Tenant {
		t.Errorf("Tenant mismatch: got %q, want %q", clone.Tenant, orig.Tenant)
	}
	if clone.ID != orig.ID {
		t.Errorf("ID mismatch: got %q, want %q", clone.ID, orig.ID)
	}
	if clone.Stored != orig.Stored {
		t.Errorf("Stored mismatch: got %v, want %v", clone.Stored, orig.Stored)
	}
	if clone.TTL != orig.TTL {
		t.Errorf("TTL mismatch: got %v, want %v", clone.TTL, orig.TTL)
	}

	if len(clone.FilterIDs) != len(orig.FilterIDs) {
		t.Errorf("FilterIDs length mismatch: got %d, want %d", len(clone.FilterIDs), len(orig.FilterIDs))
	}
	if len(clone.Weights) != len(orig.Weights) {
		t.Errorf("Weights length mismatch: got %d, want %d", len(clone.Weights), len(orig.Weights))
	}
	if len(clone.Pools) != len(orig.Pools) {
		t.Errorf("Pools length mismatch: got %d, want %d", len(clone.Pools), len(orig.Pools))
	}

	if len(clone.Pools) > 0 && len(clone.Pools[0].Weights) > 0 {
		if clone.Pools[0].Weights[0].Weight != orig.Pools[0].Weights[0].Weight {
			t.Errorf("Pool[0] Weight[0] mismatch: got %v, want %v", clone.Pools[0].Weights[0].Weight, orig.Pools[0].Weights[0].Weight)
		}
	}

	if len(clone.Pools) > 0 && len(clone.Pools[0].Blockers) > 0 {
		if clone.Pools[0].Blockers[0].Blocker != orig.Pools[0].Blockers[0].Blocker {
			t.Errorf("Pool[0] Blockers[0] mismatch: got %v, want %v", clone.Pools[0].Blockers[0].Blocker, orig.Pools[0].Blockers[0].Blocker)
		}
	}

	if &clone.Pools[0] == &orig.Pools[0] {
		t.Errorf("Pools[0] was not cloned, got same pointer")
	}
}

func TestIPProfileCacheClone(t *testing.T) {
	var p *IPProfile = nil
	res := p.CacheClone()
	if res != nil {
		if ip, ok := res.(*IPProfile); ok && ip == nil {
		} else {
			t.Errorf("CacheClone() with nil receiver: got %v, want nil or typed nil", res)
		}
	}
	orig := &IPProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		FilterIDs: []string{"f1", "f2"},
		Weights: DynamicWeights{
			&DynamicWeight{
				FilterIDs: []string{"w1"},
				Weight:    10.5,
			},
		},
		TTL:    5 * time.Minute,
		Stored: true,
		Pools: []*IPPool{
			{
				ID:        "pool1",
				FilterIDs: []string{"pf1"},
				Type:      "type1",
				Range:     "range1",
				Strategy:  "strat1",
				Message:   "msg1",
				Weights: DynamicWeights{
					&DynamicWeight{
						FilterIDs: []string{"pw1"},
						Weight:    7.5,
					},
				},
				Blockers: DynamicBlockers{
					&DynamicBlocker{
						FilterIDs: []string{"pb1"},
						Blocker:   true,
					},
				},
			},
		},
	}

	res = orig.CacheClone()
	clone, ok := res.(*IPProfile)
	if !ok {
		t.Errorf("CacheClone() returned type %T, want *IPProfile", res)
		return
	}

	if clone == nil {
		t.Errorf("CacheClone() returned nil, want non-nil")
		return
	}

	if clone.Tenant != orig.Tenant {
		t.Errorf("Tenant mismatch: got %q, want %q", clone.Tenant, orig.Tenant)
	}
}

func TestIPProfileSet(t *testing.T) {
	p := &IPProfile{}

	err := p.Set([]string{}, "val", false)
	if err != ErrWrongPath {
		t.Errorf("Set with empty path: got %v, want ErrWrongPath", err)
	}

	err = p.Set([]string{"unknown"}, "val", false)
	if err != ErrWrongPath {
		t.Errorf("Set with unknown key: got %v, want ErrWrongPath", err)
	}

	err = p.Set([]string{Tenant}, "cgrates.org", false)
	if err != nil || p.Tenant != "cgrates.org" {
		t.Errorf("Set Tenant: err=%v, value=%q", err, p.Tenant)
	}

	err = p.Set([]string{ID}, "1001", false)
	if err != nil || p.ID != "1001" {
		t.Errorf("Set ID: err=%v, value=%q", err, p.ID)
	}

	err = p.Set([]string{FilterIDs}, []string{"f1", "f2"}, false)
	if err != nil || len(p.FilterIDs) != 2 {
		t.Errorf("Set FilterIDs: err=%v, value=%v", err, p.FilterIDs)
	}

	err = p.Set([]string{TTL}, "1m", false)
	if err != nil || p.TTL != time.Minute {
		t.Errorf("Set TTL: err=%v, value=%v", err, p.TTL)
	}

	err = p.Set([]string{Stored}, true, false)
	if err != nil || !p.Stored {
		t.Errorf("Set Stored: err=%v, value=%v", err, p.Stored)
	}

	err = p.Set([]string{Weights}, "f1&f2;1.5;f3;2.5", false)
	if err != nil {
		t.Errorf("Set Weights: %v", err)
	}
	if len(p.Weights) != 2 || p.Weights[0].Weight != 1.5 || p.Weights[1].Weight != 2.5 {
		t.Errorf("Weights: got %+v", p.Weights)
	}

	err = p.Set([]string{Pools}, "val", false)
	if err != ErrWrongPath {
		t.Errorf("Set Pools missing subpath: got %v, want ErrWrongPath", err)
	}

	err = p.Set([]string{Pools, ID}, "", true)
	if err != nil {
		t.Errorf("Set Pools with empty val: %v", err)
	}
	if p.Pools != nil {
		t.Errorf("Pools should be nil, got: %+v", p.Pools)
	}

	err = p.Set([]string{Pools, ID}, "pool1", true)
	if err != nil {
		t.Errorf("Set Pools actual: %v", err)
	}
	if len(p.Pools) != 1 || p.Pools[0].ID != "pool1" {
		t.Errorf("Pools content: %+v", p.Pools)
	}
}

func TestIPPoolMerge(t *testing.T) {
	original := &IPPool{
		ID:        "pool1",
		FilterIDs: []string{"filterA"},
		Type:      "primary",
		Range:     "rangeA",
		Strategy:  "strategyA",
		Message:   "initial message",
		Weights:   DynamicWeights{{FilterIDs: []string{"wa"}, Weight: 1.0}},
		Blockers:  DynamicBlockers{{FilterIDs: []string{"ba"}, Blocker: true}},
	}

	mergeFrom := &IPPool{
		ID:        "pool1",
		FilterIDs: []string{"filterB"},
		Type:      "secondary",
		Range:     "rangeB",
		Strategy:  "strategyB",
		Message:   "updated message",
		Weights:   DynamicWeights{{FilterIDs: []string{"wb"}, Weight: 2.0}},
		Blockers:  DynamicBlockers{{FilterIDs: []string{"bb"}, Blocker: false}},
	}

	original.Merge(mergeFrom)

	if original.ID != "pool1" {
		t.Errorf("expected ID 'pool1', got %s", original.ID)
	}
	if len(original.FilterIDs) != 2 {
		t.Errorf("expected 2 FilterIDs, got %+v", original.FilterIDs)
	}
	if original.Type != "secondary" {
		t.Errorf("expected Type 'secondary', got %s", original.Type)
	}
	if original.Range != "rangeB" {
		t.Errorf("expected Range 'rangeB', got %s", original.Range)
	}
	if original.Strategy != "strategyB" {
		t.Errorf("expected Strategy 'strategyB', got %s", original.Strategy)
	}
	if original.Message != "updated message" {
		t.Errorf("expected Message 'updated message', got %s", original.Message)
	}
	if len(original.Weights) != 2 {
		t.Errorf("expected 2 Weights, got %+v", original.Weights)
	}
	if len(original.Blockers) != 2 {
		t.Errorf("expected 2 Blockers, got %+v", original.Blockers)
	}
}

func TestIPProfileString(t *testing.T) {
	p := &IPProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		FilterIDs: []string{"filter1", "filter2"},
		TTL:       5 * time.Minute,
		Stored:    true,
	}

	jsonStr := p.String()

	if !strings.Contains(jsonStr, `"Tenant":"cgrates.org"`) {
		t.Errorf("String() output missing Tenant: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"ID":"1001"`) {
		t.Errorf("String() output missing ID: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"FilterIDs":["filter1","filter2"]`) {
		t.Errorf("String() output missing FilterIDs: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"Stored":true`) {
		t.Errorf("String() output missing Stored: %s", jsonStr)
	}
}

func TestIPProfileFieldAsInterface(t *testing.T) {
	ip := &IPProfile{
		Tenant:    "cgrates.org",
		ID:        "id1",
		FilterIDs: []string{"filter1", "filter2"},
		Weights: DynamicWeights{
			{FilterIDs: []string{"wfilter1"}, Weight: 0.5},
			{FilterIDs: []string{"wfilter2"}, Weight: 0.7},
		},
		TTL:    10,
		Stored: true,
		Pools: []*IPPool{
			{
				ID:       "pool1",
				Range:    "192.168.0.0/24",
				Strategy: "strategy1",
				Message:  "test",
				Weights: DynamicWeights{
					{FilterIDs: []string{"fw1"}, Weight: 0.1},
					{FilterIDs: []string{"fw2"}, Weight: 0.9},
				},
				Blockers: []*DynamicBlocker{
					{FilterIDs: []string{"block1"}, Blocker: true},
					{FilterIDs: []string{"block2"}, Blocker: false},
				},
			},
		},
	}

	tests := []struct {
		name    string
		fldPath []string
		exp     any
		expErr  bool
	}{
		{
			name:    "Tenant",
			fldPath: []string{"Tenant"},
			exp:     "cgrates.org",
		},
		{
			name:    "ID",
			fldPath: []string{"ID"},
			exp:     "id1",
		},
		{
			name:    "FilterIDs whole slice",
			fldPath: []string{"FilterIDs"},
			exp:     []string{"filter1", "filter2"},
		},
		{
			name:    "FilterIDs first element",
			fldPath: []string{"FilterIDs[0]"},
			exp:     "filter1",
		},
		{
			name:    "Stored",
			fldPath: []string{"Stored"},
			exp:     true,
		},
		{
			name:    "TTL",
			fldPath: []string{"TTL"},
			exp:     time.Duration(10),
		},
		{
			name:    "Weights whole slice",
			fldPath: []string{"Weights"},
			exp:     ip.Weights,
		},
		// {
		// 	name:    "Weights first Weight",
		// 	fldPath: []string{"Weights[0]", "Weight"},
		// 	exp:     0.5,
		// },
		// {
		// 	name:    "Weights second FilterID first element",
		// 	fldPath: []string{"Weights[1]", "FilterIDs[0]"},
		// 	exp:     "wfilter2",
		// },
		{
			name:    "Pools whole slice",
			fldPath: []string{"Pools"},
			exp:     ip.Pools,
		},
		// {
		// 	name:    "Pools first pool ID",
		// 	fldPath: []string{"Pools[0]", "ID"},
		// 	exp:     "pool1",
		// },
		// {
		// 	name:    "Pools first pool Strategy",
		// 	fldPath: []string{"Pools[0]", "Strategy"},
		// 	exp:     "strategy1",
		// },
		// {
		// 	name:    "Pools first pool Message",
		// 	fldPath: []string{"Pools[0]", "Message"},
		// 	exp:     "test",
		// },
		// {
		// 	name:    "Pools first pool Weights first Weight",
		// 	fldPath: []string{"Pools[0]", "Weights[0]", "Weight"},
		// 	exp:     0.1,
		// },
		// {
		// 	name:    "Pools first pool Weights second Weight",
		// 	fldPath: []string{"Pools[0]", "Weights[1]", "Weight"},
		// 	exp:     0.9,
		// },
		// {
		// 	name:    "Pools first pool Blockers first FilterID first element",
		// 	fldPath: []string{"Pools[0]", "Blockers[0]", "FilterIDs[0]"},
		// 	exp:     "block1",
		// },
		// {
		// 	name:    "Pools first pool Blockers first Blocker",
		// 	fldPath: []string{"Pools[0]", "Blockers[0]", "Blocker"},
		// 	exp:     true,
		// },
		// {
		// 	name:    "Pools first pool Blockers second Blocker",
		// 	fldPath: []string{"Pools[0]", "Blockers[1]", "Blocker"},
		// 	exp:     false,
		// },
		{
			name:    "Unknown field",
			fldPath: []string{"Unknown"},
			expErr:  true,
		},
		{
			name:    "Empty path",
			fldPath: []string{},
			expErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ip.FieldAsInterface(tt.fldPath)
			if (err != nil) != tt.expErr {
				t.Fatalf("FieldAsInterface() error = %v, wantErr %v", err, tt.expErr)
			}
			if !tt.expErr {
				if !reflect.DeepEqual(got, tt.exp) {
					t.Errorf("FieldAsInterface() = %v, want %v", got, tt.exp)
				}
			}
		})
	}
}

func TestIPAllocationsLockKey(t *testing.T) {
	tnt := "cgrates.org"
	id := "1001"
	expected := "*ip_allocations:cgrates.org:1001"

	got := IPAllocationsLockKey(tnt, id)
	if got != expected {
		t.Errorf("Expected '%s', got '%s'", expected, got)
	}
}

func TestIPAllocationsTenantID(t *testing.T) {
	ipAlloc := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "1001",
	}
	expected := "cgrates.org:1001"

	got := ipAlloc.TenantID()
	if got != expected {
		t.Errorf("Expected '%s', got '%s'", expected, got)
	}
}

func TestIPAllocationsCacheClone(t *testing.T) {
	orig := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "1001",
	}

	clonedAny := orig.CacheClone()
	cloned, ok := clonedAny.(*IPAllocations)
	if !ok {
		t.Errorf("Expected type *IPAllocations, got %T", clonedAny)
	}

	if !reflect.DeepEqual(orig, cloned) {
		t.Errorf("Expected cloned object to equal original.\nOriginal: %#v\nCloned: %#v", orig, cloned)
	}

	if orig == cloned {
		t.Errorf("Expected different pointer for clone, got the same")
	}
}

func TestIPProfileLockKey(t *testing.T) {
	tnt := "cgrates.org"
	id := "profile123"
	expected := "*ip_profiles:cgrates.org:profile123"

	got := IPProfileLockKey(tnt, id)
	if got != expected {
		t.Errorf("Expected '%s', got '%s'", expected, got)
	}
}

func TestIPPoolString(t *testing.T) {
	pool := &IPPool{
		ID:        "FIRST_POOL",
		FilterIDs: []string{},
		Type:      "*ipv4",
		Range:     "192.168.122.1/24",
		Strategy:  "*ascending",
		Message:   "Some message",
		Weights: DynamicWeights{
			&DynamicWeight{
				FilterIDs: nil,
				Weight:    15,
			},
		},
		Blockers: DynamicBlockers{
			&DynamicBlocker{
				FilterIDs: nil,
				Blocker:   false,
			},
		},
	}

	jsonStr := pool.String()

	if !strings.Contains(jsonStr, `"ID":"FIRST_POOL"`) {
		t.Errorf("Expected JSON to contain ID 'FIRST_POOL', got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"Type":"*ipv4"`) {
		t.Errorf("Expected JSON to contain Type '*ipv4', got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"Range":"192.168.122.1/24"`) {
		t.Errorf("Expected JSON to contain Range, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"Weight":15`) {
		t.Errorf("Expected JSON to contain Weight 15, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"Blocker":false`) {
		t.Errorf("Expected JSON to contain Blocker false, got: %s", jsonStr)
	}
}

func TestIPPoolClone(t *testing.T) {
	t.Run("Clone valid IPPool", func(t *testing.T) {
		original := &IPPool{
			ID:        "FIRST_POOL",
			FilterIDs: []string{"flt1", "flt2"},
			Type:      "*ipv4",
			Range:     "192.168.122.1/24",
			Strategy:  "*ascending",
			Message:   "Some message",
			Weights: DynamicWeights{
				&DynamicWeight{
					FilterIDs: nil,
					Weight:    15,
				},
			},
			Blockers: DynamicBlockers{
				&DynamicBlocker{
					FilterIDs: nil,
					Blocker:   false,
				},
			},
		}

		clone := original.Clone()

		if clone == nil {
			t.Fatal("Expected clone to be non-nil")
		}
		if clone == original {
			t.Error("Clone should not be the same pointer as original")
		}

		if clone.ID != original.ID || clone.Type != original.Type || clone.Range != original.Range ||
			clone.Strategy != original.Strategy || clone.Message != original.Message {
			t.Error("Basic fields not cloned correctly")
		}

		if &clone.FilterIDs == &original.FilterIDs {
			t.Error("FilterIDs slice was not deeply copied")
		}
		if len(clone.FilterIDs) != 2 || clone.FilterIDs[0] != "flt1" {
			t.Errorf("Unexpected FilterIDs in clone: %+v", clone.FilterIDs)
		}

		if &clone.Weights == &original.Weights {
			t.Error("Weights slice was not deeply copied")
		}
		if len(clone.Weights) != 1 || clone.Weights[0].Weight != 15 {
			t.Errorf("Unexpected Weights in clone: %+v", clone.Weights)
		}
		if clone.Weights[0] == original.Weights[0] {
			t.Error("Weight pointer not deeply cloned")
		}

		if &clone.Blockers == &original.Blockers {
			t.Error("Blockers slice was not deeply copied")
		}
		if len(clone.Blockers) != 1 || clone.Blockers[0].Blocker != false {
			t.Errorf("Unexpected Blockers in clone: %+v", clone.Blockers)
		}
		if clone.Blockers[0] == original.Blockers[0] {
			t.Error("Blocker pointer not deeply cloned")
		}
	})

	t.Run("Clone nil IPPool", func(t *testing.T) {
		var p *IPPool
		if p.Clone() != nil {
			t.Error("Expected nil from Clone() on nil receiver")
		}
	})
}

func TestIPPoolFieldAsString(t *testing.T) {
	pool := &IPPool{
		ID:        "FIRST_POOL",
		FilterIDs: []string{"flt1", "flt2"},
		Type:      "*ipv4",
		Range:     "192.168.122.1/24",
		Strategy:  "*ascending",
		Message:   "Some message",
	}

	tests := []struct {
		name    string
		fldPath []string
		want    string
		wantErr bool
	}{
		{
			name:    "ID field",
			fldPath: []string{"ID"},
			want:    "FIRST_POOL",
			wantErr: false,
		},
		{
			name:    "Type field",
			fldPath: []string{"Type"},
			want:    "*ipv4",
			wantErr: false,
		},
		{
			name:    "Range field",
			fldPath: []string{"Range"},
			want:    "192.168.122.1/24",
			wantErr: false,
		},
		{
			name:    "Strategy field",
			fldPath: []string{"Strategy"},
			want:    "*ascending",
			wantErr: false,
		},
		{
			name:    "Message field",
			fldPath: []string{"Message"},
			want:    "Some message",
			wantErr: false,
		},
		{
			name:    "Invalid field",
			fldPath: []string{"Unknown"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pool.FieldAsString(tt.fldPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldAsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FieldAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPPoolFieldAsInterface(t *testing.T) {
	pool := &IPPool{
		ID:        "FIRST_POOL",
		FilterIDs: []string{"flt1", "flt2"},
		Type:      "*ipv4",
		Range:     "192.168.122.1/24",
		Strategy:  "*ascending",
		Message:   "message",
		Weights: DynamicWeights{
			&DynamicWeight{FilterIDs: nil, Weight: 15},
		},
		Blockers: DynamicBlockers{
			&DynamicBlocker{FilterIDs: nil, Blocker: false},
		},
	}

	tests := []struct {
		name    string
		fldPath []string
		want    any
		wantErr bool
	}{
		{
			name:    "ID field",
			fldPath: []string{"ID"},
			want:    "FIRST_POOL",
			wantErr: false,
		},
		{
			name:    "Type field",
			fldPath: []string{"Type"},
			want:    "*ipv4",
			wantErr: false,
		},
		{
			name:    "Range field",
			fldPath: []string{"Range"},
			want:    "192.168.122.1/24",
			wantErr: false,
		},
		{
			name:    "Strategy field",
			fldPath: []string{"Strategy"},
			want:    "*ascending",
			wantErr: false,
		},
		{
			name:    "Message field",
			fldPath: []string{"Message"},
			want:    "message",
			wantErr: false,
		},
		{
			name:    "FilterIDs full",
			fldPath: []string{"FilterIDs"},
			want:    []string{"flt1", "flt2"},
			wantErr: false,
		},
		{
			name:    "FilterIDs index out of range",
			fldPath: []string{"FilterIDs:5"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Weights field",
			fldPath: []string{"Weights"},
			want:    pool.Weights,
			wantErr: false,
		},
		{
			name:    "Blockers field",
			fldPath: []string{"Blockers"},
			want:    pool.Blockers,
			wantErr: false,
		},
		{
			name:    "Invalid field",
			fldPath: []string{"Unknown"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Too deep path",
			fldPath: []string{"ID", "extra"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty path (whole object)",
			fldPath: []string{},
			want:    pool,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pool.FieldAsInterface(tt.fldPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldAsInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			switch expected := tt.want.(type) {
			case string:
				gotStr, ok := got.(string)
				if !ok || gotStr != expected {
					t.Errorf("FieldAsInterface() = %v, want %v", got, expected)
				}
			case []string:
				gotSlice, ok := got.([]string)
				if !ok || !reflect.DeepEqual(gotSlice, expected) {
					t.Errorf("FieldAsInterface() = %v, want %v", got, expected)
				}
			default:
				if !reflect.DeepEqual(got, expected) {
					t.Errorf("FieldAsInterface() = %v, want %v", got, expected)
				}
			}
		})
	}
}

func TestIPPoolSet(t *testing.T) {

	pool := &IPPool{}

	tests := []struct {
		name    string
		path    []string
		val     any
		wantErr bool
		check   func(t *testing.T, p *IPPool)
	}{
		{
			name:    "Set ID field",
			path:    []string{"ID"},
			val:     "FIRST_POOL",
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if p.ID != "FIRST_POOL" {
					t.Errorf("ID = %v, want %v", p.ID, "FIRST_POOL")
				}
			},
		},
		{
			name:    "Set FilterIDs field",
			path:    []string{"FilterIDs"},
			val:     []string{"flt1", "flt2"},
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if len(p.FilterIDs) != 2 || p.FilterIDs[0] != "flt1" || p.FilterIDs[1] != "flt2" {
					t.Errorf("FilterIDs = %v, want [flt1 flt2]", p.FilterIDs)
				}
			},
		},
		{
			name:    "Set Type field",
			path:    []string{"Type"},
			val:     "*ipv4",
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if p.Type != "*ipv4" {
					t.Errorf("Type = %v, want %v", p.Type, "*ipv4")
				}
			},
		},
		{
			name:    "Set Range field",
			path:    []string{"Range"},
			val:     "192.168.122.1/24",
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if p.Range != "192.168.122.1/24" {
					t.Errorf("Range = %v, want %v", p.Range, "192.168.122.1/24")
				}
			},
		},
		{
			name:    "Set Strategy field",
			path:    []string{"Strategy"},
			val:     "*ascending",
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if p.Strategy != "*ascending" {
					t.Errorf("Strategy = %v, want %v", p.Strategy, "*ascending")
				}
			},
		},
		{
			name:    "Set Message field",
			path:    []string{"Message"},
			val:     "Some message",
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if p.Message != "Some message" {
					t.Errorf("Message = %v, want %v", p.Message, "Some message")
				}
			},
		},
		{
			name:    "Set Weights with valid string",
			path:    []string{"Weights"},
			val:     "flt1&flt2;15;flt3;20",
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if len(p.Weights) != 2 {
					t.Errorf("Weights count = %d, want 2", len(p.Weights))
					return
				}
				if p.Weights[0].Weight != 15 {
					t.Errorf("Weights[0].Weight = %v, want 15", p.Weights[0].Weight)
				}
				if p.Weights[1].Weight != 20 {
					t.Errorf("Weights[1].Weight = %v, want 20", p.Weights[1].Weight)
				}
				if len(p.Weights[0].FilterIDs) != 2 || p.Weights[0].FilterIDs[0] != "flt1" {
					t.Errorf("Weights[0].FilterIDs = %v, want [flt1 flt2]", p.Weights[0].FilterIDs)
				}
			},
		},
		{
			name:    "Set Blockers with valid string",
			path:    []string{"Blockers"},
			val:     "flt1&flt2;false;flt3;true",
			wantErr: false,
			check: func(t *testing.T, p *IPPool) {
				if len(p.Blockers) != 2 {
					t.Errorf("Blockers count = %d, want 2", len(p.Blockers))
					return
				}
				if p.Blockers[0].Blocker != false {
					t.Errorf("Blockers[0].Blocker = %v, want false", p.Blockers[0].Blocker)
				}
				if p.Blockers[1].Blocker != true {
					t.Errorf("Blockers[1].Blocker = %v, want true", p.Blockers[1].Blocker)
				}
				if len(p.Blockers[0].FilterIDs) != 2 || p.Blockers[0].FilterIDs[0] != "flt1" {
					t.Errorf("Blockers[0].FilterIDs = %v, want [flt1 flt2]", p.Blockers[0].FilterIDs)
				}
			},
		},
		{
			name:    "Set with wrong path length",
			path:    []string{"ID", "extra"},
			val:     "bad",
			wantErr: true,
		},
		{
			name:    "Set unknown field",
			path:    []string{"Unknown"},
			val:     "value",
			wantErr: true,
		},
		{
			name:    "Set Weights with invalid string",
			path:    []string{"Weights"},
			val:     "flt1;badweight",
			wantErr: true,
		},
		{
			name:    "Set Blockers with invalid string",
			path:    []string{"Blockers"},
			val:     "flt1;notabool",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pool.Set(tt.path, tt.val, false)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, pool)
			}
		})
	}
}

func TestIPProfileMerge(t *testing.T) {
	tests := []struct {
		name     string
		original *IPProfile
		other    *IPProfile
		expected *IPProfile
	}{
		{
			name: "Merge non-empty fields and merge pools by ID",
			original: &IPProfile{
				Tenant:    "cgrates.org",
				ID:        "origID",
				FilterIDs: []string{"f1"},
				Weights: DynamicWeights{
					&DynamicWeight{FilterIDs: []string{"w1"}, Weight: 10},
				},
				TTL:    0,
				Stored: false,
				Pools: []*IPPool{
					{ID: "pool1", Message: "original"},
				},
			},
			other: &IPProfile{
				Tenant:    "newTenant",
				ID:        "newID",
				FilterIDs: []string{"f2"},
				Weights: DynamicWeights{
					&DynamicWeight{FilterIDs: []string{"w2"}, Weight: 20},
				},
				TTL:    1 * time.Hour,
				Stored: true,
				Pools: []*IPPool{
					{ID: "pool1", Message: "merged"},
					{ID: "pool2", Message: "new"},
				},
			},
			expected: &IPProfile{
				Tenant:    "newTenant",
				ID:        "newID",
				FilterIDs: []string{"f1", "f2"},
				Weights: DynamicWeights{
					&DynamicWeight{FilterIDs: []string{"w1"}, Weight: 10},
					&DynamicWeight{FilterIDs: []string{"w2"}, Weight: 20},
				},
				TTL:    1 * time.Hour,
				Stored: true,
				Pools: []*IPPool{
					{ID: "pool1", Message: "merged"},
					{ID: "pool2", Message: "new"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.original.Merge(tt.other)

			if tt.original.Tenant != tt.expected.Tenant {
				t.Errorf("Tenant = %v, want %v", tt.original.Tenant, tt.expected.Tenant)
			}
			if tt.original.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", tt.original.ID, tt.expected.ID)
			}
			if !reflect.DeepEqual(tt.original.FilterIDs, tt.expected.FilterIDs) {
				t.Errorf("FilterIDs = %v, want %v", tt.original.FilterIDs, tt.expected.FilterIDs)
			}
			if !reflect.DeepEqual(tt.original.Weights, tt.expected.Weights) {
				t.Errorf("Weights = %v, want %v", tt.original.Weights, tt.expected.Weights)
			}
			if tt.original.TTL != tt.expected.TTL {
				t.Errorf("TTL = %v, want %v", tt.original.TTL, tt.expected.TTL)
			}
			if tt.original.Stored != tt.expected.Stored {
				t.Errorf("Stored = %v, want %v", tt.original.Stored, tt.expected.Stored)
			}
			if len(tt.original.Pools) != len(tt.expected.Pools) {
				t.Fatalf("Pools length = %v, want %v", len(tt.original.Pools), len(tt.expected.Pools))
			}
			for i := range tt.original.Pools {
				if tt.original.Pools[i].ID != tt.expected.Pools[i].ID ||
					tt.original.Pools[i].Message != tt.expected.Pools[i].Message {
					t.Errorf("Pools[%d] = %+v, want %+v", i, tt.original.Pools[i], tt.expected.Pools[i])
				}
			}
		})
	}
}

func TestIPProfileFieldAsString(t *testing.T) {
	profile := &IPProfile{
		Tenant:    "cgrates.org",
		ID:        "prof1",
		FilterIDs: []string{"flt1", "flt2"},
		Weights: DynamicWeights{
			&DynamicWeight{FilterIDs: []string{"fltW"}, Weight: 10},
		},
		TTL:    3600 * time.Second,
		Stored: true,
		Pools: []*IPPool{
			{
				ID:        "pool1",
				FilterIDs: []string{"poolflt"},
				Type:      "*ipv4",
			},
		},
	}

	tests := []struct {
		name    string
		fldPath []string
		want    string
		wantErr bool
	}{
		{
			name:    "Tenant field",
			fldPath: []string{"Tenant"},
			want:    "cgrates.org",
			wantErr: false,
		},
		{
			name:    "ID field",
			fldPath: []string{"ID"},
			want:    "prof1",
			wantErr: false,
		},
		{
			name:    "FilterIDs full slice",
			fldPath: []string{"FilterIDs"},
			want:    `["flt1","flt2"]`,
			wantErr: false,
		},
		{
			name:    "TTL field",
			fldPath: []string{"TTL"},
			want:    "1h0m0s",
			wantErr: false,
		},
		{
			name:    "Stored field",
			fldPath: []string{"Stored"},
			want:    "true",
			wantErr: false,
		},
		{
			name:    "Weights field",
			fldPath: []string{"Weights"},
			want:    `[{"FilterIDs":["fltW"],"Weight":10}]`,
			wantErr: false,
		},
		{
			name:    "Pools field",
			fldPath: []string{"Pools"},
			want:    fmt.Sprintf("[%v]", profile.Pools[0]),
			wantErr: false,
		},

		{
			name:    "Invalid field",
			fldPath: []string{"Invalid"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Too deep path",
			fldPath: []string{"Tenant", "extra"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid index in FilterIDs",
			fldPath: []string{"FilterIDs:10"},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := profile.FieldAsString(tt.fldPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("FieldAsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FieldAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPAllocationsAllocateIP(t *testing.T) {
	prfl := &IPProfile{
		Tenant: "cgrates.org",
		ID:     "TestIPAllocationsAllocateIP",
		FilterIDs: []string{
			"*string:~*req.IMSI:12345678",
		},
		TTL:    time.Duration(1) * time.Minute,
		Stored: true,
		Pools: []*IPPool{
			{
				ID:        "FIRST_POOL",
				FilterIDs: []string{},
				Type:      MetaIPv4,
				Range:     "10.10.10.10/32",
				Strategy:  MetaAscending,
				Message:   "FIRST_POOL_ALLOCATION",
				Weights: DynamicWeights{&DynamicWeight{
					Weight: 10.0,
				}},
				Blockers: DynamicBlockers{},
			},
			{
				ID:        "SECOND_POOL",
				FilterIDs: []string{},
				Type:      MetaIPv4,
				Range:     "10.10.10.20/32",
				Strategy:  MetaAscending,
				Message:   "SECOND_POOL_ALLOCATION",
				Weights: DynamicWeights{&DynamicWeight{
					Weight: 5.0,
				}},
				Blockers: DynamicBlockers{},
			},
		},
	}
	alcIP := netip.MustParseAddr("10.10.10.10")
	ipa := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "TestIPAllocationsAllocateIP",
		Allocations: map[string]*PoolAllocation{
			"alloc1": {
				PoolID:  "FIRST_POOL",
				Address: alcIP,
				Time:    time.Date(2025, time.June, 06, 14, 00, 00, 0, time.UTC),
			},
		},
		TTLIndex: []string{},
	}

	if err := ipa.ComputeUnexported(prfl); err != nil {
		t.Error(err)
	}
	now := time.Now()
	if ipAddr, err := ipa.AllocateIPOnPool("alloc1", prfl.Pools[0], false); err != nil {
		t.Error(err)
	} else if ipAddr.Address != ipa.Allocations["alloc1"].Address {
		t.Errorf("Expecting: %s, received: %s", ipa.Allocations["alloc1"].Address, ipAddr)
	} else if ipa.Allocations["alloc1"].Time.Sub(now) > time.Duration(1)*time.Second {
		t.Errorf("Allocation time is: %v", ipa.Allocations["alloc1"].Time)
	}
}
