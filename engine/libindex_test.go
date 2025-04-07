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
	dataDB, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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
	dataDB, dErr := NewInternalDB(nil, nil, true, false, cfg.DataDbCfg().Items)
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
