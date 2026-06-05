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

func TestThresholdProfileSet(t *testing.T) {
	th := ThresholdProfile{}
	exp := ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}
	if err := th.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := th.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := th.Set([]string{"NotAField", "1"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := th.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{MaxHits}, 10, false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{MinHits}, 10, false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{MinSleep}, 10, false); err != nil {
		t.Error(err)
	}

	if err := th.Set([]string{Blocker}, true, false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{Async}, true, false); err != nil {
		t.Error(err)
	}
	if err := th.Set([]string{ActionProfileIDs}, "acc1", false); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(exp, th) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(th))
	}
}

func TestThresholdProfileAsInterface(t *testing.T) {
	tp := ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}
	if _, err := tp.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := tp.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := tp.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := tp.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := tp.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := tp.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := tp.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{ActionProfileIDs}); err != nil {
		t.Fatal(err)
	} else if exp := tp.ActionProfileIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{ActionProfileIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := tp.ActionProfileIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if val, err := tp.FieldAsInterface([]string{MaxHits}); err != nil {
		t.Fatal(err)
	} else if exp := tp.MaxHits; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{MinHits}); err != nil {
		t.Fatal(err)
	} else if exp := tp.MinHits; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{MinSleep}); err != nil {
		t.Fatal(err)
	} else if exp := tp.MinSleep; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{Blocker}); err != nil {
		t.Fatal(err)
	} else if exp := tp.Blocker; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := tp.FieldAsInterface([]string{Async}); err != nil {
		t.Fatal(err)
	} else if exp := tp.Async; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := tp.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := tp.FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := tp.String(), ToJSON(tp); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestThresholdProfileMerge(t *testing.T) {
	dp := &ThresholdProfile{}
	exp := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}
	if dp.Merge(&ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights: DynamicWeights{
			{
				Weight: 10.0,
			},
		},
		MaxHits:          10,
		MinHits:          10,
		MinSleep:         10,
		Blocker:          true,
		Async:            true,
		ActionProfileIDs: []string{"acc1"},
	}); !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(dp))
	}
}

func TestThresholdsProfileSet(t *testing.T) {
	tests := []struct {
		name      string
		path      []string
		val       any
		expectErr bool
	}{
		{
			name:      "Set Tenant",
			path:      []string{Tenant},
			val:       "cgrates.org",
			expectErr: false,
		},
		{
			name:      "Set ID",
			path:      []string{ID},
			val:       "TH001",
			expectErr: false,
		},
		{
			name:      "Set Blocker true",
			path:      []string{Blocker},
			val:       true,
			expectErr: false,
		},
		{
			name:      "Set Async true",
			path:      []string{Async},
			val:       "true",
			expectErr: false,
		},
		{
			name:      "Set MaxHits",
			path:      []string{MaxHits},
			val:       99,
			expectErr: false,
		},
		{
			name:      "Set MinHits",
			path:      []string{MinHits},
			val:       "10",
			expectErr: false,
		},
		{
			name:      "Set MinSleep",
			path:      []string{MinSleep},
			val:       "2s",
			expectErr: false,
		},
		{
			name:      "Append FilterIDs",
			path:      []string{FilterIDs},
			val:       []string{"flt1", "flt2"},
			expectErr: false,
		},
		{
			name:      "Append ActionProfileIDs",
			path:      []string{ActionProfileIDs},
			val:       []string{"act1"},
			expectErr: false,
		},
		{
			name:      "Append EeIDs",
			path:      []string{EeIDs},
			val:       []string{"ee1"},
			expectErr: false,
		},
		{
			name:      "Invalid path",
			path:      []string{"Invalid"},
			val:       "test",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tp := &ThresholdProfile{}
			err := tp.Set(tc.path, tc.val, false)
			if (err != nil) != tc.expectErr {
				t.Errorf("expected error=%v, got %v", tc.expectErr, err)
			}

		})
	}
}

func TestThresholdProfileClone(t *testing.T) {
	orig := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THRESHOLD_TEST",
		FilterIDs:        []string{"*string:~*req.Account:1001", "*string:~*req.Subject:1001"},
		MaxHits:          100,
		MinHits:          5,
		MinSleep:         10 * time.Second,
		Blocker:          true,
		Weights:          DynamicWeights{&DynamicWeight{Weight: 20.0}},
		ActionProfileIDs: []string{"ACT_LOG", "ACT_MAIL", "ACT_CDR"},
		Async:            true,
		EeIDs:            []string{"EE1", "EE2"},
	}

	cloned := orig.Clone()

	if cloned == nil {
		t.Errorf("Expected cloned ThresholdProfile, got nil")
		return
	}

	if cloned.Tenant != orig.Tenant {
		t.Errorf("Tenant mismatch: got %s, want %s", cloned.Tenant, orig.Tenant)
	}
	if cloned.ID != orig.ID {
		t.Errorf("ID mismatch: got %s, want %s", cloned.ID, orig.ID)
	}
	if cloned.MaxHits != orig.MaxHits {
		t.Errorf("MaxHits mismatch: got %d, want %d", cloned.MaxHits, orig.MaxHits)
	}
	if cloned.MinHits != orig.MinHits {
		t.Errorf("MinHits mismatch: got %d, want %d", cloned.MinHits, orig.MinHits)
	}
	if cloned.MinSleep != orig.MinSleep {
		t.Errorf("MinSleep mismatch: got %v, want %v", cloned.MinSleep, orig.MinSleep)
	}
	if cloned.Blocker != orig.Blocker {
		t.Errorf("Blocker mismatch: got %v, want %v", cloned.Blocker, orig.Blocker)
	}
	if cloned.Async != orig.Async {
		t.Errorf("Async mismatch: got %v, want %v", cloned.Async, orig.Async)
	}

	if !reflect.DeepEqual(cloned.FilterIDs, orig.FilterIDs) {
		t.Errorf("FilterIDs mismatch: got %v, want %v", cloned.FilterIDs, orig.FilterIDs)
	}
	if !reflect.DeepEqual(cloned.ActionProfileIDs, orig.ActionProfileIDs) {
		t.Errorf("ActionProfileIDs mismatch: got %v, want %v", cloned.ActionProfileIDs, orig.ActionProfileIDs)
	}
	if !reflect.DeepEqual(cloned.Weights, orig.Weights) {
		t.Errorf("Weights mismatch: got %v, want %v", cloned.Weights, orig.Weights)
	}
	if len(cloned.Weights) > 0 && cloned.Weights[0] == orig.Weights[0] {
		t.Errorf("Weights not deep copied")
	}
	if !reflect.DeepEqual(cloned.EeIDs, orig.EeIDs) {
		t.Errorf("EeIDs mismatch: got %v, want %v", cloned.EeIDs, orig.EeIDs)
	}

	var nilTP *ThresholdProfile
	nilClone := nilTP.Clone()
	if nilClone != nil {
		t.Errorf("Expected nil from Clone on nil receiver, got: %+v", nilClone)
	}
}

func TestThresholdClone(t *testing.T) {
	orig := &Threshold{
		Tenant: "cgrates.org",
		ID:     "TH001",
		Hits:   42,
		Snooze: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	cloned := orig.Clone()

	if cloned == nil {
		t.Errorf("Expected non-nil clone")
		return
	}

	if cloned.Tenant != orig.Tenant {
		t.Errorf("Tenant mismatch: got %s, want %s", cloned.Tenant, orig.Tenant)
	}
	if cloned.ID != orig.ID {
		t.Errorf("ID mismatch: got %s, want %s", cloned.ID, orig.ID)
	}
	if cloned.Hits != orig.Hits {
		t.Errorf("Hits mismatch: got %d, want %d", cloned.Hits, orig.Hits)
	}
	if cloned.Snooze != orig.Snooze {
		t.Errorf("Snooze mismatch: got %v, want %v", cloned.Snooze, orig.Snooze)
	}

	var nilThreshold *Threshold
	nilClone := nilThreshold.Clone()
	if nilClone != nil {
		t.Errorf("Expected nil from Clone on nil receiver, got: %+v", nilClone)
	}
}

func TestThresholdProfileFieldAsInterface(t *testing.T) {
	tp := &ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THRESHOLD_TEST",
		FilterIDs:        []string{"flt1", "flt2"},
		MaxHits:          100,
		MinHits:          10,
		MinSleep:         5 * time.Second,
		Blocker:          true,
		Weights:          DynamicWeights{&DynamicWeight{Weight: 1.23}},
		ActionProfileIDs: []string{"act1", "act2"},
		Async:            true,
		EeIDs:            []string{"ee1", "ee2"},
	}

	tests := []struct {
		name    string
		path    []string
		want    any
		wantErr bool
	}{
		{
			name:    "Tenant",
			path:    []string{Tenant},
			want:    "cgrates.org",
			wantErr: false,
		},
		{
			name:    "ID",
			path:    []string{ID},
			want:    "THRESHOLD_TEST",
			wantErr: false,
		},
		{
			name:    "FilterIDs",
			path:    []string{FilterIDs},
			want:    []string{"flt1", "flt2"},
			wantErr: false,
		},
		{
			name:    "FilterIDs[1]",
			path:    []string{"FilterIDs[1]"},
			want:    "flt2",
			wantErr: false,
		},
		{
			name:    "ActionProfileIDs",
			path:    []string{ActionProfileIDs},
			want:    []string{"act1", "act2"},
			wantErr: false,
		},
		{
			name:    "ActionProfileIDs[0]",
			path:    []string{"ActionProfileIDs[0]"},
			want:    "act1",
			wantErr: false,
		},
		{
			name:    "EeIDs",
			path:    []string{EeIDs},
			want:    []string{"ee1", "ee2"},
			wantErr: false,
		},
		{
			name:    "EeIDs[1]",
			path:    []string{"EeIDs[1]"},
			want:    "ee2",
			wantErr: false,
		},
		{
			name:    "MaxHits",
			path:    []string{MaxHits},
			want:    100,
			wantErr: false,
		},
		{
			name:    "MinHits",
			path:    []string{MinHits},
			want:    10,
			wantErr: false,
		},
		{
			name:    "MinSleep",
			path:    []string{MinSleep},
			want:    5 * time.Second,
			wantErr: false,
		},
		{
			name:    "Blocker",
			path:    []string{Blocker},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Async",
			path:    []string{Async},
			want:    true,
			wantErr: false,
		},
		{
			name:    "Weights",
			path:    []string{Weights},
			want:    tp.Weights,
			wantErr: false,
		},
		{
			name:    "Unknown field",
			path:    []string{"UnknownField"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Out of range index",
			path:    []string{"EeIDs[10]"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid path length",
			path:    []string{"FilterIDs", "Extra"},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tp.FieldAsInterface(tc.path)
			if (err != nil) != tc.wantErr {
				t.Errorf("Expected error: %v, got: %v", tc.wantErr, err)
				return
			}
			if !tc.wantErr && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Got value %v (%T), expected %v (%T)", got, got, tc.want, tc.want)
			}
		})
	}
}
