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

package utils

import (
	"errors"
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
			{FilterIDs: []string{"fltW"}, Weight: 10},
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
			name:    "Invalid index in FilterIDs",
			fldPath: []string{"FilterIDs[10]"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "FilterIDs indexed access",
			fldPath: []string{"FilterIDs[1]"},
			want:    "flt2",
			wantErr: false,
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

func TestIPAllocationsConfig(t *testing.T) {
	prfl := &IPProfile{
		Tenant: "cgrates.org",
		ID:     "ID1",
		FilterIDs: []string{
			"*string:~*vars.radReqCode:StatusServer",
		},
		TTL:    time.Minute,
		Stored: true,
		Pools: []*IPPool{
			{
				ID:        "ID1001",
				FilterIDs: []string{},
				Type:      "*ipv4",
				Range:     "10.0.0.11/32",
				Strategy:  "*ascending",
				Message:   "OK",
				Weights: DynamicWeights{&DynamicWeight{
					Weight: 15.0,
				}},
				Blockers: DynamicBlockers{},
			},
			{
				ID:        "ID1002",
				FilterIDs: []string{},
				Type:      "*ipv4",
				Range:     "10.0.0.12/32",
				Strategy:  "*ascending",
				Message:   "OK",
				Weights: DynamicWeights{&DynamicWeight{
					Weight: 7.5,
				}},
				Blockers: DynamicBlockers{},
			},
		},
	}

	ipAllocs := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "statusServerAlloc",
		prfl:   prfl,
	}

	got := ipAllocs.Config()
	if got != prfl {
		t.Errorf("Config() returned wrong profile:\ngot:  %+v\nwant: %+v", got, prfl)
	}
}

func TestIPAllocationsLock(t *testing.T) {
	ipAlloc := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "IDLock",
	}

	givenLockID := "ID2012"
	ipAlloc.Lock(givenLockID)
	if ipAlloc.lockID != givenLockID {
		t.Errorf("Expected lockID %s, got %s", givenLockID, ipAlloc.lockID)
	}

	ipAlloc2 := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "testAutoLock",
	}
	ipAlloc2.Lock("")
	if ipAlloc2.lockID == "" {
		t.Errorf("Expected auto-generated lockID, got empty string")
	}
}

func TestIPAllocationsUnlock(t *testing.T) {
	ipa1 := &IPAllocations{}
	ipa1.Unlock()
	if ipa1.lockID != "" {
		t.Errorf("Expected empty lockID, got %q", ipa1.lockID)
	}

	ipa2 := &IPAllocations{
		lockID: "LockID",
	}
	ipa2.Unlock()
	if ipa2.lockID != "" {
		t.Errorf("Expected lockID to be cleared, got %q", ipa2.lockID)
	}
}

func TestIPAllocationsRremoveAllocFromTTLIndex(t *testing.T) {
	ipa := &IPAllocations{
		TTLIndex: []string{"allocID1", "allocID2", "allocID3"},
	}

	ipa.removeAllocFromTTLIndex("allocID2")
	expected := []string{"allocID1", "allocID3"}
	if !reflect.DeepEqual(ipa.TTLIndex, expected) {
		t.Errorf("After removing 'alloc2', TTLIndex = %v, want %v", ipa.TTLIndex, expected)
	}

	ipa.removeAllocFromTTLIndex("nonexistent")
	if !reflect.DeepEqual(ipa.TTLIndex, expected) {
		t.Errorf("After removing 'nonexistent', TTLIndex = %v, want %v", ipa.TTLIndex, expected)
	}
}

func TestIPAllocationsReleaseAllocation(t *testing.T) {
	allocID1 := "allocID1"
	allocID2 := "allocID2"

	addr1 := netip.MustParseAddr("10.20.30.40")
	addr2 := netip.MustParseAddr("10.20.30.41")

	poolID1 := "firstPool"

	ipa := &IPAllocations{
		Allocations: map[string]*PoolAllocation{
			allocID1: {PoolID: poolID1, Address: addr1, Time: time.Now()},
			allocID2: {PoolID: poolID1, Address: addr2, Time: time.Now()},
		},
		poolAllocs: map[string]map[netip.Addr]string{
			poolID1: {
				addr1: allocID1,
				addr2: allocID2,
			},
		},
		TTLIndex: []string{allocID1, allocID2},
		prfl: &IPProfile{
			TTL: 5 * time.Minute,
		},
	}

	err := ipa.ReleaseAllocation(allocID1)
	if err != nil {
		t.Fatalf("ReleaseAllocation returned error: %v", err)
	}

	if _, exists := ipa.Allocations[allocID1]; exists {
		t.Errorf("Allocation %s was not removed from Allocations", allocID1)
	}

	if _, exists := ipa.poolAllocs[poolID1][addr1]; exists {
		t.Errorf("Allocation %s was not removed from poolAllocs", allocID1)
	}

	for _, id := range ipa.TTLIndex {
		if id == allocID1 {
			t.Errorf("Allocation %s was not removed from TTLIndex", allocID1)
		}
	}

	if _, exists := ipa.Allocations[allocID2]; !exists {
		t.Errorf("Allocation %s should still exist", allocID2)
	}

	err = ipa.ReleaseAllocation("nonExistingAlloc")
	if err == nil {
		t.Error("Expected error when releasing non-existing allocation, got nil")
	}
}
func TestAllocatedIPAsNavigableMap(t *testing.T) {
	ip := &AllocatedIP{
		ProfileID: "IDprofile",
		PoolID:    "ID1",
		Message:   "OK",
		Address:   netip.MustParseAddr("192.168.1.100"),
	}

	nmap := ip.AsNavigableMap()

	tests := []struct {
		key     string
		wantVal string
	}{
		{"ProfileID", "IDprofile"},
		{"PoolID", "ID1"},
		{"Message", "OK"},
		{"Address", "192.168.1.100"},
	}

	for _, tt := range tests {
		node, ok := nmap[tt.key]
		if !ok {
			t.Errorf("AsNavigableMap() missing key %q", tt.key)
			continue
		}
		if node == nil || node.Value == nil {
			t.Errorf("AsNavigableMap() key %q has nil DataNode or Value", tt.key)
			continue
		}

		strVal, ok := node.Value.Data.(string)
		if !ok {
			t.Errorf("AsNavigableMap() key %q data is not a string", tt.key)
			continue
		}

		if strVal != tt.wantVal {
			t.Errorf("AsNavigableMap() key %q = %q; want %q", tt.key, strVal, tt.wantVal)
		}
	}
}

func TestIPAllocationsRemoveExpiredUnits(t *testing.T) {
	now := time.Now()

	expiredAlloc := &PoolAllocation{
		PoolID:  "pool1",
		Address: netip.MustParseAddr("192.168.100.119"),
		Time:    now.Add(-2 * time.Hour),
	}
	activeAlloc := &PoolAllocation{
		PoolID:  "pool1",
		Address: netip.MustParseAddr("192.168.56.1"),
		Time:    now,
	}

	ipAlloc := &IPAllocations{
		Allocations: map[string]*PoolAllocation{
			"expired": expiredAlloc,
			"active":  activeAlloc,
		},
		TTLIndex: []string{"expired", "active"},
		prfl:     &IPProfile{TTL: time.Hour},
		poolAllocs: map[string]map[netip.Addr]string{
			"pool1": {
				expiredAlloc.Address: "expired",
				activeAlloc.Address:  "active",
			},
		},
	}

	ipAlloc.removeExpiredUnits()

	if len(ipAlloc.Allocations) != 1 {
		t.Fatalf("expected 1 allocation after cleanup, got %d", len(ipAlloc.Allocations))
	}
	if _, exists := ipAlloc.Allocations["expired"]; exists {
		t.Error("expired allocation was not removed")
	}
	if _, exists := ipAlloc.Allocations["active"]; !exists {
		t.Error("active allocation was incorrectly removed")
	}
	if len(ipAlloc.TTLIndex) != 1 || ipAlloc.TTLIndex[0] != "active" {
		t.Errorf("TTLIndex not updated correctly, got %v", ipAlloc.TTLIndex)
	}
	if _, exists := ipAlloc.poolAllocs["pool1"][expiredAlloc.Address]; exists {
		t.Error("expired address not removed from poolAllocs")
	}
}
func TestIPProfileLock(t *testing.T) {
	ipProf := &IPProfile{
		Tenant: "cgrates.org",
		ID:     "ID1",
	}

	customLockID := "LockID"
	ipProf.Lock(customLockID)

	if ipProf.lockID != customLockID {
		t.Errorf("expected lockID %q, got %q", customLockID, ipProf.lockID)
	}

	ipProf = &IPProfile{
		Tenant: "cgrates.org",
		ID:     "ID2",
	}

	ipProf.Lock("")

	if ipProf.lockID == "" {
		t.Error("expected non-empty lockID from Guardian, got empty string")
	}
}

func TestIPProfileUnlock(t *testing.T) {
	lockID := "IDlock1"
	ipProf := &IPProfile{
		Tenant: "cgrates.org",
		ID:     "ID1",
		lockID: lockID,
	}

	ipProf.Unlock()

	if ipProf.lockID != "" {
		t.Errorf("expected lockID to be cleared after Unlock, got %q", ipProf.lockID)
	}

	ipProf = &IPProfile{
		Tenant: "cgrates.org",
		ID:     "ID2",
		lockID: "",
	}

	ipProf.Unlock()
}

func TestPoolAllocationIsActive(t *testing.T) {
	ttl := 5 * time.Minute

	alloc := &PoolAllocation{
		PoolID:  "ID1",
		Address: netip.MustParseAddr("192.168.100.42"),
		Time:    time.Now().Add(-3 * time.Minute),
	}

	if !alloc.IsActive(ttl) {
		t.Errorf("expected allocation to be active (within TTL), but got inactive")
	}

	allocExpired := &PoolAllocation{
		PoolID:  "ID2",
		Address: netip.MustParseAddr("10.10.10.25"),
		Time:    time.Now().Add(-10 * time.Minute),
	}

	if allocExpired.IsActive(ttl) {
		t.Errorf("expected allocation to be inactive (TTL expired), but got active")
	}
}

func TestPoolAllocationClone(t *testing.T) {
	original := &PoolAllocation{
		PoolID:  "ID1",
		Address: netip.MustParseAddr("172.16.0.10"),
		Time:    time.Now(),
	}
	cloned := original.Clone()

	if cloned == nil {
		t.Fatal("expected cloned object, got nil")
	}
	if cloned == original {
		t.Errorf("expected a different instance, got the same")
	}
	if *cloned != *original {
		t.Errorf("expected clone to be equal to original, got different values")
	}

	var nilOriginal *PoolAllocation
	nilCloned := nilOriginal.Clone()

	if nilCloned != nil {
		t.Errorf("expected nil clone from nil original, got non-nil")
	}
}

func TestIPAllocationsClone(t *testing.T) {
	now := time.Now()

	orig := &IPAllocations{
		Tenant:   "cgrates.org",
		ID:       "alloc1",
		TTLIndex: []string{"allocA", "allocB"},
		prfl: &IPProfile{
			Tenant: "cgrates.org",
			ID:     "profile1",
			TTL:    time.Hour * 24,
		},
		poolRanges: map[string]netip.Prefix{
			"pool1": netip.MustParsePrefix("192.168.100.0/24"),
		},
		poolAllocs: map[string]map[netip.Addr]string{
			"pool1": {
				netip.MustParseAddr("192.168.100.119"): "allocA",
			},
		},
		Allocations: map[string]*PoolAllocation{
			"allocA": {
				PoolID:  "pool1",
				Address: netip.MustParseAddr("192.168.100.119"),
				Time:    now.Add(-time.Hour),
			},
			"allocB": {
				PoolID:  "pool1",
				Address: netip.MustParseAddr("192.168.56.1"),
				Time:    now,
			},
		},
	}

	clone := orig.Clone()

	if clone.Tenant != orig.Tenant {
		t.Errorf("Tenant mismatch: got %s, want %s", clone.Tenant, orig.Tenant)
	}
	if clone.ID != orig.ID {
		t.Errorf("ID mismatch: got %s, want %s", clone.ID, orig.ID)
	}

	if &clone.TTLIndex == &orig.TTLIndex {
		t.Error("TTLIndex slice was not cloned deeply")
	}
	if len(clone.TTLIndex) != len(orig.TTLIndex) {
		t.Errorf("TTLIndex length mismatch: got %d, want %d", len(clone.TTLIndex), len(orig.TTLIndex))
	}

	if clone.prfl == orig.prfl {
		t.Error("IPProfile was not cloned deeply")
	}
	if clone.prfl.ID != orig.prfl.ID {
		t.Errorf("Profile ID mismatch: got %s, want %s", clone.prfl.ID, orig.prfl.ID)
	}

	if &clone.poolRanges == &orig.poolRanges {
		t.Error("poolRanges map was not cloned deeply")
	}
	if _, ok := clone.poolRanges["pool1"]; !ok {
		t.Error("poolRanges missing pool1")
	}

	if &clone.poolAllocs == &orig.poolAllocs {
		t.Error("poolAllocs map was not cloned deeply")
	}
	if poolMap, ok := clone.poolAllocs["pool1"]; !ok || poolMap[netip.MustParseAddr("192.168.100.119")] != "allocA" {
		t.Error("poolAllocs missing expected allocation")
	}

	if &clone.Allocations == &orig.Allocations {
		t.Error("Allocations map was not cloned deeply")
	}
	if alloc, ok := clone.Allocations["allocA"]; !ok {
		t.Error("Allocations missing allocA")
	} else {
		if alloc.Address != orig.Allocations["allocA"].Address {
			t.Error("Allocation address mismatch")
		}
		if alloc == orig.Allocations["allocA"] {
			t.Error("Allocation PoolAllocation was not cloned deeply")
		}
	}
}

func TestIPAllocationsCloneNil(t *testing.T) {
	var a *IPAllocations = nil
	clone := a.Clone()
	if clone != nil {
		t.Error("Clone() of nil IPAllocations should return nil")
	}
}

func TestIPAllocationsAllocateIPOnPool(t *testing.T) {
	pool := &IPPool{
		ID:      "poolA",
		Message: "OK",
	}
	prefix, _ := netip.ParsePrefix("192.168.100.119/32")
	allocs := &IPAllocations{
		Tenant:   "cgrTenant",
		ID:       "profile1",
		TTLIndex: []string{},
		prfl:     &IPProfile{TTL: time.Minute * 5},
		poolRanges: map[string]netip.Prefix{
			pool.ID: prefix,
		},
		poolAllocs:  make(map[string]map[netip.Addr]string),
		Allocations: make(map[string]*PoolAllocation),
	}

	allocID := "alloc1"
	allocatedIP, err := allocs.AllocateIPOnPool(allocID, pool, true)
	if err != nil {
		t.Fatalf("DryRun allocation failed: %v", err)
	}
	if allocatedIP.Address != prefix.Addr() {
		t.Errorf("Expected address %v, got %v", prefix.Addr(), allocatedIP.Address)
	}
	if len(allocs.Allocations) != 0 {
		t.Errorf("DryRun should not create allocation, but allocations exist")
	}

	allocatedIP, err = allocs.AllocateIPOnPool(allocID, pool, false)
	if err != nil {
		t.Fatalf("Allocation failed: %v", err)
	}
	if allocatedIP.Address != prefix.Addr() {
		t.Errorf("Expected address %v, got %v", prefix.Addr(), allocatedIP.Address)
	}
	if len(allocs.Allocations) != 1 {
		t.Errorf("Expected 1 allocation, got %d", len(allocs.Allocations))
	}

	oldTime := allocs.Allocations[allocID].Time
	time.Sleep(10 * time.Millisecond)
	allocatedIP, err = allocs.AllocateIPOnPool(allocID, pool, false)
	if err != nil {
		t.Fatalf("Re-allocation failed: %v", err)
	}
	if !allocs.Allocations[allocID].Time.After(oldTime) {
		t.Errorf("Allocation time was not refreshed")
	}
	if len(allocs.TTLIndex) == 0 || allocs.TTLIndex[len(allocs.TTLIndex)-1] != allocID {
		t.Errorf("TTLIndex was not updated correctly")
	}

	allocID2 := "alloc2"
	_, err = allocs.AllocateIPOnPool(allocID2, pool, false)
	if err == nil {
		t.Fatal("Expected allocation conflict error, got none")
	}
	if !errors.Is(err, ErrIPAlreadyAllocated) {
		t.Fatalf("Expected ErrIPAlreadyAllocated, got %v", err)
	}

	multiPrefix, _ := netip.ParsePrefix("192.168.100.0/24")
	allocs.poolRanges["multiPool"] = multiPrefix
	multiPool := &IPPool{ID: "multiPool", Message: "multi pool"}
	_, err = allocs.AllocateIPOnPool("alloc3", multiPool, false)
	if err == nil || err.Error() != "only single IP Pools are supported for now" {
		t.Fatalf("Expected error for multi IP pool, got %v", err)
	}
}

func TestIPProfileFieldsAsInterface(t *testing.T) {
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
				Strategy: MetaAscending,
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
			{
				ID:       "pool2",
				Range:    "192.168.0.0/24",
				Strategy: MetaAscending,
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
		{
			name:    "Pools first pool ID",
			fldPath: []string{"Pools[0]", "ID"},
			exp:     "pool1",
		},
		{
			name:    "Pools first pool Strategy",
			fldPath: []string{"Pools[0]", "Strategy"},
			exp:     "*ascending",
		},
		{
			name:    "Pools first pool Message",
			fldPath: []string{"Pools[0]", "Message"},
			exp:     "test",
		},
		{
			name:    "Pools first pool Message",
			fldPath: []string{"Pools[1]", "Message"},
			exp:     "test",
		},
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
					t.Errorf("FieldAsInterface() = %v, want %v", ToJSON(got), tt.exp)
				}
			}
		})
	}
}

func TestIPAllocationUnlock(t *testing.T) {
	ipAlloc := &IPAllocations{
		lockID: "",
		prfl:   nil,
	}
	ipAlloc.Unlock()
	if ipAlloc.lockID != "" {
		t.Errorf("expected lockID to remain empty, got %q", ipAlloc.lockID)
	}

	ipAlloc.lockID = "lock1"
	ipAlloc.prfl = nil
	ipAlloc.Unlock()
	if ipAlloc.lockID != "" {
		t.Errorf("expected lockID cleared, got %q", ipAlloc.lockID)
	}

	profile := &IPProfile{
		lockID: "profileLock2",
	}
	ipAlloc.lockID = "lock2"
	ipAlloc.prfl = profile

	ipAlloc.Unlock()

	if ipAlloc.lockID != "" {
		t.Errorf("expected ipAlloc lockID cleared, got %q", ipAlloc.lockID)
	}
	if profile.lockID != "" {
		t.Errorf("expected profile lockID cleared after Unlock, got %q", profile.lockID)
	}
}

func TestIPAllocationsComputeUnexported(t *testing.T) {
	tests := []struct {
		name           string
		allocations    *IPAllocations
		profile        *IPProfile
		expectedError  bool
		validateResult func(t *testing.T, a *IPAllocations)
	}{
		{
			name: "successful computation",
			allocations: &IPAllocations{
				Tenant: "cgrates.org",
				ID:     "ID1",
				Allocations: map[string]*PoolAllocation{
					"alloc1": {
						PoolID:  "pool1",
						Address: netip.MustParseAddr("192.168.1.10"),
						Time:    time.Now(),
					},
					"alloc2": {
						PoolID:  "pool1",
						Address: netip.MustParseAddr("192.168.1.11"),
						Time:    time.Now(),
					},
					"alloc3": {
						PoolID:  "pool2",
						Address: netip.MustParseAddr("10.0.0.5"),
						Time:    time.Now(),
					},
				},
			},
			profile: &IPProfile{
				Tenant: "cgrates.org",
				ID:     "IPProfileID1",
				Pools: []*IPPool{
					{
						ID:    "pool1",
						Range: "192.168.1.0/24",
						Type:  "static",
					},
					{
						ID:    "pool2",
						Range: "10.0.0.0/16",
						Type:  "dynamic",
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, a *IPAllocations) {
				if a.prfl == nil {
					t.Error("Expected profile to be set")
				}
				if a.prfl.ID != "IPProfileID1" {
					t.Errorf("Expected profile ID 'IPProfileID1', got '%s'", a.prfl.ID)
				}

				if len(a.poolAllocs) != 2 {
					t.Errorf("Expected 2 pools in poolAllocs, got %d", len(a.poolAllocs))
				}

				pool1Allocs := a.poolAllocs["pool1"]
				if len(pool1Allocs) != 2 {
					t.Errorf("Expected 2 allocations in pool1, got %d", len(pool1Allocs))
				}
				if allocID := pool1Allocs[netip.MustParseAddr("192.168.1.10")]; allocID != "alloc1" {
					t.Errorf("Expected alloc1 for IP 192.168.1.10, got %s", allocID)
				}
				if allocID := pool1Allocs[netip.MustParseAddr("192.168.1.11")]; allocID != "alloc2" {
					t.Errorf("Expected alloc2 for IP 192.168.1.11, got %s", allocID)
				}

				pool2Allocs := a.poolAllocs["pool2"]
				if len(pool2Allocs) != 1 {
					t.Errorf("Expected 1 allocation in pool2, got %d", len(pool2Allocs))
				}
				if allocID := pool2Allocs[netip.MustParseAddr("10.0.0.5")]; allocID != "alloc3" {
					t.Errorf("Expected alloc3 for IP 10.0.0.5, got %s", allocID)
				}

				if len(a.poolRanges) != 2 {
					t.Errorf("Expected 2 pool ranges, got %d", len(a.poolRanges))
				}

				expectedPrefix1 := netip.MustParsePrefix("192.168.1.0/24")
				if a.poolRanges["pool1"] != expectedPrefix1 {
					t.Errorf("Expected prefix %s for pool1, got %s", expectedPrefix1, a.poolRanges["pool1"])
				}

				expectedPrefix2 := netip.MustParsePrefix("10.0.0.0/16")
				if a.poolRanges["pool2"] != expectedPrefix2 {
					t.Errorf("Expected prefix %s for pool2, got %s", expectedPrefix2, a.poolRanges["pool2"])
				}
			},
		},
		{
			name: "empty allocations",
			allocations: &IPAllocations{
				Tenant:      "cgrates.org",
				ID:          "ID1",
				Allocations: map[string]*PoolAllocation{},
			},
			profile: &IPProfile{
				Tenant: "cgrates.org",
				ID:     "IPProfileID1",
				Pools: []*IPPool{
					{
						ID:    "pool1",
						Range: "192.168.1.0/24",
						Type:  "static",
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, a *IPAllocations) {
				if a.prfl == nil {
					t.Error("Expected profile to be set")
				}
				if len(a.poolAllocs) != 0 {
					t.Errorf("Expected 0 pools in poolAllocs, got %d", len(a.poolAllocs))
				}
				if len(a.poolRanges) != 1 {
					t.Errorf("Expected 1 pool range, got %d", len(a.poolRanges))
				}
			},
		},
		{
			name: "nil allocations map",
			allocations: &IPAllocations{
				Tenant:      "cgrates.org",
				ID:          "ID1",
				Allocations: nil,
			},
			profile: &IPProfile{
				Tenant: "cgrates.org",
				ID:     "IPProfileID1",
				Pools: []*IPPool{
					{
						ID:    "pool1",
						Range: "192.168.1.0/24",
						Type:  "static",
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, a *IPAllocations) {
				if a.prfl == nil {
					t.Error("Expected profile to be set")
				}
				if len(a.poolAllocs) != 0 {
					t.Errorf("Expected 0 pools in poolAllocs, got %d", len(a.poolAllocs))
				}
				if len(a.poolRanges) != 1 {
					t.Errorf("Expected 1 pool range, got %d", len(a.poolRanges))
				}
			},
		},
		{
			name: "invalid CIDR range",
			allocations: &IPAllocations{
				Tenant:      "cgrates.org",
				ID:          "ID1",
				Allocations: map[string]*PoolAllocation{},
			},
			profile: &IPProfile{
				Tenant: "cgrates.org",
				ID:     "IPProfileID1",
				Pools: []*IPPool{
					{
						ID:    "pool1",
						Range: "invalid-cidr",
						Type:  "static",
					},
				},
			},
			expectedError: true,
			validateResult: func(t *testing.T, a *IPAllocations) {
				if a.prfl == nil {
					t.Error("Expected profile to be set even on error")
				}
			},
		},
		{
			name: "profile with no pools",
			allocations: &IPAllocations{
				Tenant: "cgrates.org",
				ID:     "ID1",
				Allocations: map[string]*PoolAllocation{
					"alloc1": {
						PoolID:  "pool1",
						Address: netip.MustParseAddr("192.168.1.10"),
						Time:    time.Now(),
					},
				},
			},
			profile: &IPProfile{
				Tenant: "cgrates.org",
				ID:     "IPProfileID1",
				Pools:  []*IPPool{},
			},
			expectedError: false,
			validateResult: func(t *testing.T, a *IPAllocations) {
				if a.prfl == nil {
					t.Error("Expected profile to be set")
				}
				if len(a.poolAllocs) != 1 {
					t.Errorf("Expected 1 pool in poolAllocs, got %d", len(a.poolAllocs))
				}
				if len(a.poolRanges) != 0 {
					t.Errorf("Expected 0 pool ranges, got %d", len(a.poolRanges))
				}
			},
		},
		{
			name: "multiple IPs in same pool",
			allocations: &IPAllocations{
				Tenant: "cgrates.org",
				ID:     "ID1",
				Allocations: map[string]*PoolAllocation{
					"alloc1": {
						PoolID:  "pool1",
						Address: netip.MustParseAddr("192.168.1.10"),
						Time:    time.Now(),
					},
					"alloc2": {
						PoolID:  "pool1",
						Address: netip.MustParseAddr("192.168.1.11"),
						Time:    time.Now(),
					},
					"alloc3": {
						PoolID:  "pool1",
						Address: netip.MustParseAddr("192.168.1.12"),
						Time:    time.Now(),
					},
				},
			},
			profile: &IPProfile{
				Tenant: "cgrates.org",
				ID:     "IPProfileID1",
				Pools: []*IPPool{
					{
						ID:    "pool1",
						Range: "192.168.1.0/24",
						Type:  "static",
					},
				},
			},
			expectedError: false,
			validateResult: func(t *testing.T, a *IPAllocations) {
				if len(a.poolAllocs) != 1 {
					t.Errorf("Expected 1 pool in poolAllocs, got %d", len(a.poolAllocs))
				}
				pool1Allocs := a.poolAllocs["pool1"]
				if len(pool1Allocs) != 3 {
					t.Errorf("Expected 3 allocations in pool1, got %d", len(pool1Allocs))
				}

				expectedMappings := map[string]string{
					"192.168.1.10": "alloc1",
					"192.168.1.11": "alloc2",
					"192.168.1.12": "alloc3",
				}

				for ipStr, expectedAllocID := range expectedMappings {
					ip := netip.MustParseAddr(ipStr)
					if actualAllocID := pool1Allocs[ip]; actualAllocID != expectedAllocID {
						t.Errorf("Expected %s for IP %s, got %s", expectedAllocID, ipStr, actualAllocID)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.allocations.ComputeUnexported(tt.profile)

			if tt.expectedError {
				if err == nil {
					t.Error("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if tt.validateResult != nil {
				tt.validateResult(t, tt.allocations)
			}
		})
	}
}

func TestIPAllocationsClearAllocations(t *testing.T) {
	addr1 := netip.MustParseAddr("192.168.1.10")
	addr2 := netip.MustParseAddr("192.168.1.11")

	ip1 := netip.MustParseAddr("192.168.1.10")
	ip2 := netip.MustParseAddr("192.168.1.11")

	alloc := &IPAllocations{
		Tenant: "cgrates.org",
		ID:     "profile1",
		Allocations: map[string]*PoolAllocation{
			"a1": {
				PoolID:  "pool1",
				Address: ip1,
				Time:    time.Now().Add(-2 * time.Minute),
			},
			"a2": {
				PoolID:  "pool1",
				Address: ip2,
				Time:    time.Now().Add(-1 * time.Minute),
			},
		},
		TTLIndex: []string{"a1", "a2"},
		prfl: &IPProfile{
			Tenant: "cgrates.org",
			ID:     "profile1",
			TTL:    time.Minute,
			Stored: true,
			Pools: []*IPPool{
				{
					ID:       "pool1",
					Type:     "*ipv4",
					Range:    "192.168.1.0/24",
					Strategy: "*ascending",
				},
			},
		},
		poolRanges: map[string]netip.Prefix{
			"pool1": netip.MustParsePrefix("192.168.1.0/24"),
		},
		poolAllocs: map[string]map[netip.Addr]string{
			"pool1": {
				ip1: "a1",
				ip2: "a2",
			},
		},
	}

	alloc.Allocations["a1"] = &PoolAllocation{PoolID: "p1", Address: addr1, Time: time.Now()}
	alloc.Allocations["a2"] = &PoolAllocation{PoolID: "p1", Address: addr2, Time: time.Now()}
	alloc.TTLIndex = []string{"a1", "a2"}
	alloc.poolAllocs["p1"] = map[netip.Addr]string{
		addr1: "a1",
		addr2: "a2",
	}

	err := alloc.ClearAllocations([]string{"doesNotExist"})
	if err == nil {
		t.Errorf("expected error for missing allocation ID")
	}
	if len(alloc.Allocations) != 2 {
		t.Errorf("expected allocations unchanged, got %d", len(alloc.Allocations))
	}

	err = alloc.ClearAllocations([]string{"a1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := alloc.Allocations["a1"]; ok {
		t.Errorf("expected a1 to be cleared")
	}
	if _, ok := alloc.Allocations["a2"]; !ok {
		t.Errorf("expected a2 to remain")
	}
	if _, ok := alloc.poolAllocs["p1"][addr1]; ok {
		t.Errorf("expected addr1 to be removed from poolAllocs")
	}
	if len(alloc.TTLIndex) != 1 || alloc.TTLIndex[0] != "a2" {
		t.Errorf("expected TTLIndex to contain only a2, got %v", alloc.TTLIndex)
	}

	err = alloc.ClearAllocations(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alloc.Allocations) != 0 {
		t.Errorf("expected all allocations cleared, got %d", len(alloc.Allocations))
	}
	if len(alloc.poolAllocs["p1"]) != 0 {
		t.Errorf("expected all poolAllocs cleared, got %d", len(alloc.poolAllocs["p1"]))
	}
	if len(alloc.TTLIndex) != 0 {
		t.Errorf("expected TTLIndex cleared, got %v", alloc.TTLIndex)
	}
}
