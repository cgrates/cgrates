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
	tests := []struct {
		name         string
		tntCtxIdxKey string
		wantErr      string
	}{
		{
			name:         "invalid index key format with less than 4 parts",
			tntCtxIdxKey: "invalid:key",
			wantErr:      "WRONG_IDX_KEY_FORMAT<invalid:key>",
		},
		{
			name:         "another invalid index key format with less than 4 parts",
			tntCtxIdxKey: "another:invalid:key",
			wantErr:      "WRONG_IDX_KEY_FORMAT<another:invalid:key>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := splitFilterIndex(tt.tntCtxIdxKey)
			if err == nil {
				t.Fatalf("Expected error but got nil")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("Expected error %v, but got %v", tt.wantErr, err.Error())
			}
		})
	}
}
