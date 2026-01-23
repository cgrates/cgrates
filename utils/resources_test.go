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
	"reflect"
	"testing"
	"time"
)

func TestResourceProfileTenantID(t *testing.T) {
	testStruct := ResourceProfile{
		Tenant: "test_tenant",
		ID:     "test_id",
	}
	result := testStruct.TenantID()
	expected := ConcatenatedKey(testStruct.Tenant, testStruct.ID)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestResourceUsageTenantID(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant: "test_tenant",
		ID:     "test_id",
	}
	result := testStruct.TenantID()
	expected := ConcatenatedKey(testStruct.Tenant, testStruct.ID)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestResourceUsageIsActive(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.IsActive(time.Date(2014, 1, 13, 0, 0, 0, 0, time.UTC))
	if !reflect.DeepEqual(true, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", true, result)
	}
}

func TestResourceUsageIsActiveFalse(t *testing.T) {
	testStruct := ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.IsActive(time.Date(2014, 1, 15, 0, 0, 0, 0, time.UTC))
	if !reflect.DeepEqual(false, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", false, result)
	}
}

func TestResourceUsageClone(t *testing.T) {
	testStruct := &ResourceUsage{
		Tenant:     "test_tenant",
		ID:         "test_id",
		ExpiryTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
	}
	result := testStruct.Clone()
	if !reflect.DeepEqual(testStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct, result)
	}
	testStruct.Tenant = "test_tenant2"
	testStruct.ID = "test_id2"
	testStruct.ExpiryTime = time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC)
	if reflect.DeepEqual(testStruct.Tenant, result.Tenant) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.Tenant, result.Tenant)
	}
	if reflect.DeepEqual(testStruct.ID, result.ID) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.ID, result.ID)
	}
	if reflect.DeepEqual(testStruct.ExpiryTime, result.ExpiryTime) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.ExpiryTime, result.ExpiryTime)
	}
}

func TestResourceTenantID(t *testing.T) {
	testStruct := Resource{
		Tenant: "test_tenant",
	}
	result := testStruct.TenantID()
	if reflect.DeepEqual(testStruct.Tenant, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct.Tenant, result)
	}
}

func TestResourceTotalUsage1(t *testing.T) {
	testStruct := Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
			"0": {
				Tenant:     "test_tenant2",
				ID:         "test_id2",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      1,
			},
			"1": {
				Tenant:     "test_tenant3",
				ID:         "test_id3",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      2,
			},
		},
	}
	result := testStruct.TotalUsage()
	if reflect.DeepEqual(3, result) {
		t.Errorf("\nExpecting <3>,\n Received <%+v>", result)
	}
}

func TestResourceTotalUsage2(t *testing.T) {
	testStruct := Resource{
		Tenant: "test_tenant",
		ID:     "test_id",
		Usages: map[string]*ResourceUsage{
			"0": {
				Tenant:     "test_tenant2",
				ID:         "test_id2",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      1,
			},
			"1": {
				Tenant:     "test_tenant3",
				ID:         "test_id3",
				ExpiryTime: time.Date(2015, 1, 14, 0, 0, 0, 0, time.UTC),
				Units:      2,
			},
		},
	}
	result := testStruct.TotalUsage()
	if reflect.DeepEqual(3, result) {
		t.Errorf("\nExpecting <3>,\n Received <%+v>", result)
	}
}

func TestResourcesAvailable(t *testing.T) {
	r := ResourceWithConfig{
		Resource: &Resource{
			Usages: map[string]*ResourceUsage{
				"RU_1": {
					Units: 4,
				},
				"RU_2": {
					Units: 7,
				},
			},
		},
		Config: &ResourceProfile{
			Limit: 10,
		},
	}

	exp := -1.0
	rcv := r.Available()

	if rcv != exp {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

type mockWriter struct {
	WriteF func(p []byte) (n int, err error)
}

func (mW *mockWriter) Write(p []byte) (n int, err error) {
	if mW.WriteF != nil {
		return mW.WriteF(p)
	}
	return 0, nil
}

func TestResourceProfileSet(t *testing.T) {
	cp := ResourceProfile{}
	exp := ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			}},
		UsageTTL:          10,
		Limit:             10,
		AllocationMessage: "new",
		Blocker:           true,
		Stored:            true,
		ThresholdIDs:      []string{"TH1"},
	}
	if err := cp.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := cp.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := cp.Set([]string{"NotAField", "1"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := cp.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{UsageTTL}, 10, false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{Limit}, 10, false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{AllocationMessage}, "new", false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{Blocker}, true, false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{Stored}, true, false); err != nil {
		t.Error(err)
	}
	if err := cp.Set([]string{ThresholdIDs}, "TH1", false); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, cp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(cp))
	}
}

func TestResourceProfileAsInterface(t *testing.T) {
	rp := ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			}},
		UsageTTL:          10,
		Limit:             10,
		AllocationMessage: "new",
		Blocker:           true,
		Stored:            true,
		ThresholdIDs:      []string{"TH1"},
	}
	if _, err := rp.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := rp.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{ThresholdIDs}); err != nil {
		t.Fatal(err)
	} else if exp := rp.ThresholdIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{ThresholdIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := rp.ThresholdIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := rp.FieldAsInterface([]string{UsageTTL}); err != nil {
		t.Fatal(err)
	} else if exp := rp.UsageTTL; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Limit}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Limit; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{AllocationMessage}); err != nil {
		t.Fatal(err)
	} else if exp := rp.AllocationMessage; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Blocker}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Blocker; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := rp.FieldAsInterface([]string{Stored}); err != nil {
		t.Fatal(err)
	} else if exp := rp.Stored; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := rp.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := rp.FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := rp.String(), ToJSON(rp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestResourceProfileMerge(t *testing.T) {
	dp := &ResourceProfile{}
	exp := &ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			}},
		UsageTTL:          10,
		Limit:             10,
		AllocationMessage: "new",
		Blocker:           true,
		Stored:            true,
		ThresholdIDs:      []string{"TH1"},
	}
	if dp.Merge(&ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10,
			}},
		UsageTTL:          10,
		Limit:             10,
		AllocationMessage: "new",
		Blocker:           true,
		Stored:            true,
		ThresholdIDs:      []string{"TH1"},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(dp))
	}
}

func TestResourceProfileCloneCacheClone(t *testing.T) {
	var nilRP *ResourceProfile
	if nilRP.Clone() != nil {
		t.Fatal("Expected nil clone for nil ResourceProfile")
	}

	rp := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RP1",
		FilterIDs:         []string{"fltr1", "*string:~*req.Account:1001"},
		ThresholdIDs:      []string{"TH1"},
		UsageTTL:          10,
		Limit:             100,
		AllocationMessage: "ok",
		Blocker:           true,
		Stored:            true,
		Weights: DynamicWeights{
			{Weight: 10},
		},
	}

	tests := []struct {
		name  string
		clone func(*ResourceProfile) *ResourceProfile
	}{
		{
			name: "Clone",
			clone: func(r *ResourceProfile) *ResourceProfile {
				return r.Clone()
			},
		},
		{
			name: "CacheClone",
			clone: func(r *ResourceProfile) *ResourceProfile {
				cc := r.CacheClone()
				cl, ok := cc.(*ResourceProfile)
				if !ok {
					t.Fatalf("CacheClone returned %T, expected *ResourceProfile", cc)
				}
				return cl
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := tt.clone(rp)

			if !reflect.DeepEqual(rp, cl) {
				t.Fatalf("%s not equal\nexp=%v\ngot=%v",
					tt.name, ToJSON(rp), ToJSON(cl))
			}

			cl.FilterIDs[0] = "changed"
			cl.ThresholdIDs[0] = "changed"
			cl.Weights[0].Weight = 99

			if rp.FilterIDs[0] == "changed" {
				t.Errorf("%s did not deep-copy FilterIDs", tt.name)
			}
			if rp.ThresholdIDs[0] == "changed" {
				t.Errorf("%s did not deep-copy ThresholdIDs", tt.name)
			}
			if rp.Weights[0].Weight == 99 {
				t.Errorf("%s did not deep-copy Weights", tt.name)
			}
		})
	}
}

func TestResourceCloneAndCacheClone(t *testing.T) {
	var nilR *Resource
	if nilR.Clone() != nil {
		t.Fatal("Expected nil clone for nil Resource")
	}

	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"u1": {
				Tenant:     "cgrates.org",
				ID:         "U1",
				Units:      10,
				ExpiryTime: time.Now().Add(time.Hour),
			},
			"u2": {
				Tenant: "cgrates.org",
				ID:     "U2",
				Units:  20,
			},
		},
		TTLIdx: []string{"u1", "u2"},
	}

	tests := []struct {
		name  string
		clone func(*Resource) *Resource
	}{
		{
			name: "Clone",
			clone: func(res *Resource) *Resource {
				return res.Clone()
			},
		},
		{
			name: "CacheClone",
			clone: func(res *Resource) *Resource {
				cc := res.CacheClone()
				cl, ok := cc.(*Resource)
				if !ok {
					t.Fatalf("CacheClone returned %T, expected *Resource", cc)
				}
				return cl
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := tt.clone(r)

			if !reflect.DeepEqual(r, cl) {
				t.Fatalf("%s not equal\nexp=%v\ngot=%v",
					tt.name, ToJSON(r), ToJSON(cl))
			}

			delete(cl.Usages, "u1")
			if _, ok := r.Usages["u1"]; !ok {
				t.Errorf("%s did not deep-copy Usages map", tt.name)
			}

			cl.Usages["u2"].Units = 999
			if r.Usages["u2"].Units == 999 {
				t.Errorf("%s did not deep-copy ResourceUsage", tt.name)
			}

			cl.TTLIdx[0] = "changed"
			if r.TTLIdx[0] == "changed" {
				t.Errorf("%s did not deep-copy TTLIdx", tt.name)
			}
		})
	}
}

func TestResourceAsMapStringInterface(t *testing.T) {
	var nilR *Resource
	if nilR.AsMapStringInterface() != nil {
		t.Fatal("Expected nil map for nil Resource")
	}

	r := &Resource{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]*ResourceUsage{
			"u1": {
				Tenant: "cgrates.org",
				ID:     "U1",
				Units:  10,
			},
		},
		TTLIdx: []string{"u1"},
	}

	m := r.AsMapStringInterface()

	if m == nil {
		t.Fatal("Expected non-nil map")
	}

	if v, ok := m[Tenant].(string); !ok || v != r.Tenant {
		t.Errorf("Tenant mismatch, expected %v got %v", r.Tenant, m[Tenant])
	}
	if v, ok := m[ID].(string); !ok || v != r.ID {
		t.Errorf("ID mismatch, expected %v got %v", r.ID, m[ID])
	}

	usg, ok := m[Usages].(map[string]*ResourceUsage)
	if !ok {
		t.Fatalf("Expected Usages as map[string]*ResourceUsage, got %T", m[Usages])
	}
	if !reflect.DeepEqual(usg, r.Usages) {
		t.Errorf("Usages mismatch\nexp=%v\ngot=%v", ToJSON(r.Usages), ToJSON(usg))
	}

	ttl, ok := m[TTLIdx].([]string)
	if !ok {
		t.Fatalf("Expected TTLIdx as []string, got %T", m[TTLIdx])
	}
	if !reflect.DeepEqual(ttl, r.TTLIdx) {
		t.Errorf("TTLIdx mismatch\nexp=%v\ngot=%v", ToJSON(r.TTLIdx), ToJSON(ttl))
	}
}

func TestMapStringInterfaceToResource(t *testing.T) {
	m := map[string]any{
		Tenant: "cgrates.org",
		ID:     "RES1",
		Usages: map[string]any{
			"u1": map[string]any{
				Tenant: "cgrates.org",
				ID:     "U1",
				Units:  10.0,
			},
			"u2": &ResourceUsage{
				Tenant: "cgrates.org",
				ID:     "U2",
				Units:  20,
			},
		},
		TTLIdx: []any{"u1", "u2"},
	}

	r := MapStringInterfaceToResource(m)

	if r == nil {
		t.Fatal("Expected non-nil Resource")
	}

	if r.Tenant != "cgrates.org" {
		t.Errorf("Tenant mismatch, got %v", r.Tenant)
	}
	if r.ID != "RES1" {
		t.Errorf("ID mismatch, got %v", r.ID)
	}

	if len(r.Usages) != 2 {
		t.Fatalf("Expected 2 usages, got %d", len(r.Usages))
	}

	if u1 := r.Usages["u1"]; u1 == nil || u1.ID != "U1" || u1.Units != 10 {
		t.Errorf("Usage u1 not correctly mapped: %+v", u1)
	}
	if u2 := r.Usages["u2"]; u2 == nil || u2.ID != "U2" || u2.Units != 20 {
		t.Errorf("Usage u2 not correctly mapped: %+v", u2)
	}

	if !reflect.DeepEqual(r.TTLIdx, []string{"u1", "u2"}) {
		t.Errorf("TTLIdx mismatch, got %v", r.TTLIdx)
	}
}

func TestResourceProfileLockKey(t *testing.T) {
	tnt := "cgrates.org"
	id := "RP1"

	exp := ConcatenatedKey(CacheResourceProfiles, tnt, id)
	got := ResourceProfileLockKey(tnt, id)

	if exp != got {
		t.Errorf("ResourceProfileLockKey mismatch\nexp=%v\ngot=%v", exp, got)
	}
}

func TestResourceLockKey(t *testing.T) {
	tnt := "cgrates.org"
	id := "RES1"

	exp := ConcatenatedKey(CacheResources, tnt, id)
	got := ResourceLockKey(tnt, id)

	if exp != got {
		t.Errorf("ResourceLockKey mismatch\nexp=%v\ngot=%v", exp, got)
	}
}
