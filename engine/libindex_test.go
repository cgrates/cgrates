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
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestLibIndexIsDynamicDPPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"Dynamic Path (stats)", "~*stats/", true},
		{"Static Path", "/static/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDynamicDPPath(tt.path)
			if got != tt.want {
				t.Errorf("IsDynamicDPPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestLibIndexRemoveFilterIndexesForFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	tntCtx := "cgrates.org:*sessions"

	tests := []struct {
		name    string
		idx     map[string]utils.StringSet // initial indexes map
		keys    []string                   // that will be removed from the index
		itemIDs utils.StringSet
		want    map[string]utils.StringSet // indexes map after remove
	}{
		{
			name: "remove one filter index from all profiles",
			idx: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP1", "AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
			keys:    []string{"*string:~*req.Account:1001"},
			itemIDs: utils.NewStringSet([]string{"AP1", "AP2"}),
			want: map[string]utils.StringSet{
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
		},
		{
			name: "remove one filter index from one profile",
			idx: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP1", "AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
			keys:    []string{"*string:~*req.Account:1001"},
			itemIDs: utils.NewStringSet([]string{"AP1"}),
			want: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
		},
		{
			name: "attempt remove non-existent filter index",
			idx: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP1", "AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
			keys:    []string{"*string:~*req.Account:notfound"},
			itemIDs: utils.NewStringSet([]string{"AP1", "AP2"}),
			want: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP1", "AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
		},
		{
			name: "remove all filter indexes from one profile",
			idx: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP1", "AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
			keys:    []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
			itemIDs: utils.NewStringSet([]string{"AP1"}),
			want: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP2"}),
			},
		},
		{
			name: "remove all filter indexes from all profiles",
			idx: map[string]utils.StringSet{
				"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP1", "AP2"}),
				"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
			},
			keys:    []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
			itemIDs: utils.NewStringSet([]string{"AP1", "AP2"}),
			want:    nil,
		},
		{
			name: "remove multiple filter indexes from mixed profiles",
			idx: map[string]utils.StringSet{
				"*string:~*req.Account:1001":     utils.NewStringSet([]string{"AP1", "AP2", "AP3"}),
				"*string:~*req.Destination:1010": utils.NewStringSet([]string{"AP2", "AP3"}),
				"*string:~*req.Destination:1011": utils.NewStringSet([]string{"AP1", "AP3", "AP4"}),
				"*string:~*req.Destination:1012": utils.NewStringSet([]string{"AP2"}),
			},
			keys:    []string{"*string:~*req.Destination:1010", "*string:~*req.Destination:1012"},
			itemIDs: utils.NewStringSet([]string{"AP2", "AP3"}),
			want: map[string]utils.StringSet{
				"*string:~*req.Account:1001":     utils.NewStringSet([]string{"AP1", "AP2", "AP3"}),
				"*string:~*req.Destination:1011": utils.NewStringSet([]string{"AP1", "AP3", "AP4"}),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(func() {
				if err := dataDB.Flush(""); err != nil {
					t.Logf("failed to flush dataDB: %v", err)
				}
			})
			if err := dm.SetIndexes(utils.CacheAttributeFilterIndexes, tntCtx, test.idx, true, ""); err != nil {
				t.Fatalf("dm.SetFilterIndexes() returned unexpected error: %v", err)
			}
			if err := removeFilterIndexesForFilter(dm, utils.CacheAttributeFilterIndexes,
				tntCtx, test.keys, test.itemIDs); err != nil {
				t.Fatalf("removeFilterIndexesForFilter() returned unexpected error: %v", err)
			}
			got, err := dm.GetIndexes(utils.CacheAttributeFilterIndexes, tntCtx, "", true, false)
			switch len(test.want) {
			case 0:
				if !errors.Is(err, utils.ErrNotFound) {
					t.Fatalf("dm.GetFilterIndexes(%q,%q) err = %v, want %v",
						utils.CacheAttributeFilterIndexes, tntCtx, err, utils.ErrNotFound)
				}
			default:
				if err != nil {
					t.Fatalf("dm.GetFilterIndexes(%q,%q) returned unexpected error: %v",
						utils.CacheAttributeFilterIndexes, tntCtx, err)
				}
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("dm.GetFilterIndexes(%q,%q) returned unexpected indexes (-want +got): \n%s",
					utils.CacheAttributeFilterIndexes, tntCtx, diff)
			}
		})
	}
}

func TestLibIndexSplitFilterIndexErrWrongIdxKeyFormat(t *testing.T) {
	tntCtxIdxKey := "invalid:key"
	expectedErr := "WRONG_IDX_KEY_FORMAT<invalid:key>"
	_, _, err := splitFilterIndex(tntCtxIdxKey)
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestLibIndexNewFilterIndexGetFilterErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, dErr := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if dErr != nil {
		t.Error(dErr)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	tnt := "tenant"
	ctx := "context"
	itemID := "item1"
	filterIDs := []string{"nonexistent_filter"}
	idxItmType := "indexItemType"
	newFlt := &Filter{
		Tenant: tnt,
		ID:     "filter1",
		Rules:  []*FilterRule{},
	}
	_, err := newFilterIndex(dm, idxItmType, tnt, ctx, itemID, filterIDs, newFlt)
	expectedErr := "broken reference to filter: nonexistent_filter for itemType: indexItemType and ID: item1"
	if err == nil || err.Error() != expectedErr {
		t.Fatalf("Expected error %v, got %v", expectedErr, err)
	}
}
func TestLibIndex_newFilterIndex(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB, err := NewInternalDB(nil, nil, true, nil, cfg.DataDbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)

	for i := range 2 {
		idx := strconv.Itoa(i + 1)
		flt := &Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_" + idx,
			Rules: []*FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Field" + idx,
					Values:  []string{"val" + idx},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Field" + idx,
					Values:  []string{"val"},
				},
				{
					Type:    utils.MetaSuffix,
					Element: "~*req.Field" + idx,
					Values:  []string{idx},
				},
				{
					Type:    utils.MetaExists,
					Element: "~*req.Field" + idx,
				},
			},
		}
		if err := dm.SetFilter(flt, true); err != nil {
			t.Fatal(err)
		}
	}

	filterIDsList := [][]string{
		{"*prefix:~*req.Field2:val", "*empty:~*req.Field3:", "FLTR_1"},
		{"*suffix:~*req.Field1:1", "*notstring:~*req.Field2:val1", "FLTR_2"},
		{"*exists:~*req.Field2:", "*suffix:~*req.Field1:1", "FLTR_1"},
	}

	for i, filterIDs := range filterIDsList {
		idx := strconv.Itoa(i + 1)
		if err := dm.SetAttributeProfile(&AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_" + idx,
			FilterIDs: filterIDs,
			Contexts:  []string{utils.MetaCDRs},
		}, true); err != nil {
			t.Fatal(err)
		}
	}

	wantIndexes := map[string]utils.StringSet{
		"*exists:*req.Field1": {
			"ATTR_1": {},
			"ATTR_3": {},
		},
		"*exists:*req.Field2": {
			"ATTR_2": {},
			"ATTR_3": {},
		},
		"*prefix:*req.Field1:val": {
			"ATTR_1": {},
			"ATTR_3": {},
		},
		"*prefix:*req.Field2:val": {
			"ATTR_1": {},
			"ATTR_2": {},
		},
		"*string:*req.Field1:val1": {
			"ATTR_1": {},
			"ATTR_3": {},
		},
		"*string:*req.Field2:val2": {
			"ATTR_2": {},
		},
		"*suffix:*req.Field1:1": {
			"ATTR_1": {},
			"ATTR_2": {},
			"ATTR_3": {},
		},
		"*suffix:*req.Field2:2": {
			"ATTR_2": {},
		},
	}
	tntCtx := utils.ConcatenatedKey("cgrates.org", utils.MetaCDRs)
	gotIndexes, err := dm.GetIndexes(utils.CacheAttributeFilterIndexes, tntCtx, "", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotIndexes, wantIndexes) {
		t.Errorf("dm.GetIndexes() = %s, want %s", utils.ToJSON(gotIndexes), utils.ToJSON(wantIndexes))
	}

	tests := []struct {
		name       string
		idxItmType string
		tnt        string
		ctx        string
		itemID     string
		filterIDs  []string
		newFlt     *Filter
		want       map[string]utils.StringSet
		wantErr    bool
	}{
		{
			name:       "no filterIDs",
			idxItmType: utils.CacheAttributeFilterIndexes,
			want: map[string]utils.StringSet{
				"*none:*any:*any": {},
			},
		},
		{
			name:       "newFlt with no filterIDs",
			idxItmType: utils.CacheAttributeFilterIndexes,
			newFlt: &Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_TEST",
				Rules: []*FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Field",
						Values:  []string{"val"},
					},
				},
			},
			want: map[string]utils.StringSet{
				"*none:*any:*any": {},
			},
		},
		{
			name:       "broken reference",
			idxItmType: utils.CacheAttributeFilterIndexes,
			itemID:     "ATTR_1",
			filterIDs:  []string{"FLTR_1"},
			wantErr:    true,
		},
		{
			name:       "unindexable filter",
			idxItmType: utils.CacheAttributeFilterIndexes,
			filterIDs:  []string{"*notstring:~*req.Field1:val2"},
			want:       make(map[string]utils.StringSet),
		},
		{
			name:       "dynamic element, constant value",
			idxItmType: utils.CacheAttributeFilterIndexes,
			filterIDs:  []string{"*string:~*req.Field1:val1"},
			want: map[string]utils.StringSet{
				"*string:*req.Field1:val1": {},
			},
		},
		{
			name:       "constant element, dynamic value",
			idxItmType: utils.CacheAttributeFilterIndexes,
			filterIDs:  []string{"*string:val1:~*req.Field1"},
			want: map[string]utils.StringSet{
				"*string:*req.Field1:val1": {},
			},
		},
		{
			name:       "dynamic element, dynamic value",
			idxItmType: utils.CacheAttributeFilterIndexes,
			filterIDs:  []string{"*string:~*req.Field1:~*req.Field1"},
			want:       make(map[string]utils.StringSet),
		},
		{
			name:       "constant element, constant value",
			idxItmType: utils.CacheAttributeFilterIndexes,
			filterIDs:  []string{"*string:val1:val1"},
			want:       make(map[string]utils.StringSet),
		},
		{
			name:       "filter and tenant without context",
			idxItmType: utils.CacheAttributeFilterIndexes,
			tnt:        "cgrates.org",
			// ctx:        utils.MetaCDRs,
			filterIDs: []string{"*string:~*req.Field1:val1"},
			want: map[string]utils.StringSet{
				"*string:*req.Field1:val1": {},
			},
		},
		{
			name:       "filter and tenant with context (without references)",
			idxItmType: utils.CacheAttributeFilterIndexes,
			tnt:        "cgrates.org",
			ctx:        utils.MetaCDRs,
			filterIDs:  []string{"*string:~*req.Random:val"},
			want: map[string]utils.StringSet{
				"*string:*req.Random:val": {},
			},
		},
		{
			name:       "filter and tenant with context (with references)",
			idxItmType: utils.CacheAttributeFilterIndexes,
			tnt:        "cgrates.org",
			ctx:        utils.MetaCDRs,
			filterIDs:  []string{"*string:~*req.Field1:val1"},
			want: map[string]utils.StringSet{
				"*string:*req.Field1:val1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
			},
		},
		{
			name:       "filterID of newFlt",
			idxItmType: utils.CacheAttributeFilterIndexes,
			tnt:        "cgrates.org",
			ctx:        utils.MetaCDRs,
			filterIDs:  []string{"FLTR_TEST"},
			newFlt: &Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_TEST",
				Rules: []*FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Random",
						Values:  []string{"val"},
					},
					{
						Type:    utils.MetaString,
						Element: "~*req.Field1",
						Values:  []string{"val1"},
					},
				},
			},
			want: map[string]utils.StringSet{
				"*string:*req.Field1:val1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*string:*req.Random:val": {},
			},
		},
		{
			name:       "all indexable filters with references",
			idxItmType: utils.CacheAttributeFilterIndexes,
			tnt:        "cgrates.org",
			ctx:        "*cdrs",
			filterIDs: []string{
				"FLTR_1",
				"FLTR_2",
				"*prefix:~*req.Field2:val",
				"*suffix:~*req.Field1:1",
				"*suffix:~*req.Field1:1",
			},
			want: map[string]utils.StringSet{
				"*exists:*req.Field1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*exists:*req.Field2": {
					"ATTR_2": {},
					"ATTR_3": {},
				},
				"*prefix:*req.Field1:val": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*prefix:*req.Field2:val": {
					"ATTR_1": {},
					"ATTR_2": {},
				},
				"*string:*req.Field1:val1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*string:*req.Field2:val2": {
					"ATTR_2": {},
				},
				"*suffix:*req.Field1:1": {
					"ATTR_1": {},
					"ATTR_2": {},
					"ATTR_3": {},
				},
				"*suffix:*req.Field2:2": {
					"ATTR_2": {},
				},
			},
		},
		{
			name:       "all filters",
			idxItmType: utils.CacheAttributeFilterIndexes,
			tnt:        "cgrates.org",
			ctx:        "*cdrs",
			filterIDs: []string{
				"FLTR_1",
				"FLTR_2",
				"*prefix:~*req.Field2:val",
				"*suffix:~*req.Field1:1",
				"*exists:~*req.Field2:",
				"*empty:~*req.Field3:",
				"*notstring:~*req.Field2:val1",
			},
			want: map[string]utils.StringSet{
				"*exists:*req.Field1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*exists:*req.Field2": {
					"ATTR_2": {},
					"ATTR_3": {},
				},
				"*prefix:*req.Field1:val": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*prefix:*req.Field2:val": {
					"ATTR_1": {},
					"ATTR_2": {},
				},
				"*string:*req.Field1:val1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*string:*req.Field2:val2": {
					"ATTR_2": {},
				},
				"*suffix:*req.Field1:1": {
					"ATTR_1": {},
					"ATTR_2": {},
					"ATTR_3": {},
				},
				"*suffix:*req.Field2:2": {
					"ATTR_2": {},
				},
			},
		},
		{
			name:       "all filters (with extra unreferenced filters)",
			idxItmType: utils.CacheAttributeFilterIndexes,
			tnt:        "cgrates.org",
			ctx:        "*cdrs",
			filterIDs: []string{
				"FLTR_1",
				"FLTR_2",
				"*prefix:~*req.Field2:val",
				"*suffix:~*req.Field1:1",
				"*exists:~*req.Field2:",
				"*empty:~*req.Field3:",
				"*notstring:~*req.Field2:val1",
				"*prefix:~*req.Field4:val",
				"*string:~*req.Field5:val5",
			},
			want: map[string]utils.StringSet{
				"*exists:*req.Field1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*exists:*req.Field2": {
					"ATTR_2": {},
					"ATTR_3": {},
				},
				"*prefix:*req.Field1:val": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*prefix:*req.Field2:val": {
					"ATTR_1": {},
					"ATTR_2": {},
				},
				"*string:*req.Field1:val1": {
					"ATTR_1": {},
					"ATTR_3": {},
				},
				"*string:*req.Field2:val2": {
					"ATTR_2": {},
				},
				"*suffix:*req.Field1:1": {
					"ATTR_1": {},
					"ATTR_2": {},
					"ATTR_3": {},
				},
				"*suffix:*req.Field2:2": {
					"ATTR_2": {},
				},
				"*prefix:*req.Field4:val":  {},
				"*string:*req.Field5:val5": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := newFilterIndex(dm, tt.idxItmType, tt.tnt, tt.ctx, tt.itemID, tt.filterIDs, tt.newFlt)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("newFilterIndex() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("newFilterIndex() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newFilterIndex() = %s, want %s", utils.ToJSON(got), utils.ToJSON(tt.want))
			}
		})
	}
}

// func TestLibIndex_prepareFilterIndexMap(t *testing.T) {
// 	cfg := config.NewDefaultCGRConfig()
// 	dataDB := NewInternalDB(nil, nil, true, true, cfg.DataDbCfg().Items)
// 	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
//
// 	var flt *Filter // to be used as newFlt
// 	for i := range 2 {
// 		idx := strconv.Itoa(i + 1)
// 		flt = &Filter{
// 			Tenant: "cgrates.org",
// 			ID:     "FLTR_" + idx,
// 			Rules: []*FilterRule{
// 				{
// 					Type:    utils.MetaString,
// 					Element: "~*req.Field" + idx,
// 					Values:  []string{"val" + idx},
// 				},
// 				{
// 					Type:    utils.MetaPrefix,
// 					Element: "~*req.Field" + idx,
// 					Values:  []string{"val"},
// 				},
// 				{
// 					Type:    utils.MetaSuffix,
// 					Element: "~*req.Field" + idx,
// 					Values:  []string{idx},
// 				},
// 				{
// 					Type:    utils.MetaExists,
// 					Element: "~*req.Field" + idx,
// 				},
// 			},
// 		}
// 		if err := dm.SetFilter(flt, true); err != nil {
// 			t.Fatal(err)
// 		}
// 	}
//
// 	filterIDsList := [][]string{
// 		{"*prefix:~*req.Field2:val", "FLTR_1"},
// 		{"*suffix:~*req.Field1:1", "FLTR_2"},
// 		{"*exists:~*req.Field2:", "*suffix:~*req.Field1:1", "FLTR_1"},
// 	}
//
// 	for i, filterIDs := range filterIDsList {
// 		idx := strconv.Itoa(i + 1)
// 		if err := dm.SetChargerProfile(&ChargerProfile{
// 			Tenant:       "cgrates.org",
// 			ID:           "CP_" + idx,
// 			FilterIDs:    filterIDs,
// 			RunID:        "DEFAULT" + idx,
// 			AttributeIDs: []string{"*none"},
// 		}, true); err != nil {
// 			t.Fatal(err)
// 		}
// 		if err := dm.SetAttributeProfile(&AttributeProfile{
// 			Tenant:    "cgrates.org",
// 			ID:        "ATTR_CDRS_" + idx,
// 			FilterIDs: filterIDs,
// 			Contexts:  []string{utils.MetaCDRs},
// 		}, true); err != nil {
// 			t.Fatal(err)
// 		}
// 		if err := dm.SetAttributeProfile(&AttributeProfile{
// 			Tenant:    "cgrates.org",
// 			ID:        "ATTR_SESSIONS_" + idx,
// 			FilterIDs: filterIDs,
// 			Contexts:  []string{utils.MetaSessionS},
// 		}, true); err != nil {
// 			t.Fatal(err)
// 		}
// 	}
//
// 	wantTemplate := func(prefix string) map[string]utils.StringSet {
// 		return map[string]utils.StringSet{
// 			"*prefix:*req.Field1:val": {
// 				prefix + "_1": {},
// 				prefix + "_3": {},
// 			},
// 			"*prefix:*req.Field2:val": {
// 				prefix + "_1": {},
// 				prefix + "_2": {},
// 			},
// 			"*string:*req.Field1:val1": {
// 				prefix + "_1": {},
// 				prefix + "_3": {},
// 			},
// 			"*string:*req.Field2:val2": {
// 				prefix + "_2": {},
// 			},
// 			"*suffix:*req.Field1:1": {
// 				prefix + "_1": {},
// 				prefix + "_2": {},
// 				prefix + "_3": {},
// 			},
// 			"*suffix:*req.Field2:2": {
// 				prefix + "_2": {},
// 			},
// 		}
// 	}
//
// 	argsList := []struct {
// 		itemType string
// 		ctx      string
// 		prefix   string // of the profile (CP, ATTR_CDRS, ATTR_SESSIONS)
// 	}{
// 		{
// 			itemType: utils.CacheChargerFilterIndexes,
// 			ctx:      "",
// 			prefix:   "CP",
// 		},
// 		{
// 			itemType: utils.CacheAttributeFilterIndexes,
// 			ctx:      utils.MetaCDRs,
// 			prefix:   "ATTR_CDRS",
// 		},
// 		{
// 			itemType: utils.CacheAttributeFilterIndexes,
// 			ctx:      utils.MetaSessionS,
// 			prefix:   "ATTR_SESSIONS",
// 		},
// 	}
//
// 	for _, args := range argsList {
// 		wantIndexes := wantTemplate(args.prefix)
// 		tntCtx := "cgrates.org"
// 		if args.ctx != "" {
// 			tntCtx = utils.ConcatenatedKey(tntCtx, args.ctx)
// 		}
// 		gotIndexes, err := dm.GetIndexes(args.itemType, tntCtx, "", false, false)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		if !reflect.DeepEqual(gotIndexes, wantIndexes) {
// 			t.Errorf("dm.GetIndexes() = %s, want %s", utils.ToJSON(gotIndexes), utils.ToJSON(wantIndexes))
// 		}
//
// 	}
//
// 	tests := []struct {
// 		name       string
// 		idxItmType string
// 		tnt        string
// 		ctx        string
// 		itemID     string
// 		filterIDs  []string
// 		newFlt     *Filter
// 		want       map[string]utils.StringSet
// 		wantErr    bool
// 	}{
// 		{
// 			name:       "test1",
// 			idxItmType: utils.CacheAttributeFilterIndexes,
// 			tnt:        "cgrates.org",
// 			ctx:        "*cdrs",
// 			itemID:     "ATTR_CDRS",
// 			filterIDs:  []string{"FLTR_1"},
// 			newFlt:     flt,
// 			want:       map[string]utils.StringSet{},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, gotErr := prepareFilterIndexMap(dm, tt.idxItmType, tt.tnt, tt.ctx, tt.itemID, tt.filterIDs, tt.newFlt)
// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					t.Errorf("prepareFilterIndexMap() failed: %v", gotErr)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				t.Fatal("prepareFilterIndexMap() succeeded unexpectedly")
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("prepareFilterIndexMap() = %s, want %s", utils.ToJSON(got), utils.ToJSON(tt.want))
// 			}
// 		})
// 	}
// }
