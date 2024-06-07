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

func TestLibIndexRemoveFilterIndexesForFilterErrnotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	tntCtx := "cgrates.org:*sessions"
	test := struct {
		name    string
		idx     map[string]utils.StringSet
		keys    []string
		itemIDs utils.StringSet
	}{
		name: "testing for ErrNotFound",
		idx: map[string]utils.StringSet{
			"*string:~*req.Account:1001": utils.NewStringSet([]string{"AP1", "AP2"}),
			"*string:~*req.Account:1002": utils.NewStringSet([]string{"AP1", "AP2"}),
		},
		keys:    []string{"*string:~*req.Account:1001"},
		itemIDs: utils.NewStringSet([]string{"AP1", "AP2"}),
	}
	t.Run(test.name, func(t *testing.T) {
		t.Cleanup(func() {
			if err := dataDB.Flush(""); err != nil {
				t.Logf("failed to flush dataDB: %v", err)
			}
		})
		if err := dm.SetIndexes(utils.CacheAttributeFilterIndexes, tntCtx, test.idx, true, ""); err != nil {
			t.Fatalf("dm.SetIndexes() returned unexpected error: %v", err)
		}
		err := removeFilterIndexesForFilter(dm, utils.CacheAttributeFilterIndexes, tntCtx, test.keys, test.itemIDs)
		if err != nil && !errors.Is(err, utils.ErrNotFound) {
			t.Fatalf("Expected error %v, got %v", utils.ErrNotFound, err)
		}
	})
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
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
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
