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
