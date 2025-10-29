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

package engine

import (
	"errors"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestLibIndexRemoveFilterIndexesForFilter(t *testing.T) {
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatalf("failed to init default cfg: %v", err)
	}
	dataDB := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	tntCtx := "cgrates.org:*sessions"

	tests := []struct {
		name    string
		idx     map[string]utils.StringMap // initial indexes map
		keys    []string                   // that will be removed from the index
		itemIDs utils.StringMap
		want    map[string]utils.StringMap // indexes map after remove
	}{
		{
			name: "remove one filter index from all profiles",
			idx: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
			keys: []string{"*string:~*req.Account:1001"},
			itemIDs: utils.StringMap{
				"AP1": true,
				"AP2": true,
			},
			want: map[string]utils.StringMap{
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
		},
		{
			name: "remove one filter index from one profile",
			idx: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
			keys: []string{"*string:~*req.Account:1001"},
			itemIDs: utils.StringMap{
				"AP1": true,
			},
			want: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
		},
		{
			name: "attempt remove non-existent filter index",
			idx: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
			keys: []string{"*string:~*req.Account:notfound"},
			itemIDs: utils.StringMap{
				"AP1": true,
				"AP2": true,
			},
			want: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
		},
		{
			name: "remove all filter indexes from one profile",
			idx: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
			keys: []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
			itemIDs: utils.StringMap{
				"AP1": true,
			},
			want: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP2": true,
				},
			},
		},
		{
			name: "remove all filter indexes from all profiles",
			idx: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
				},
				"*string:~*req.Account:1002": {
					"AP1": true,
					"AP2": true,
				},
			},
			keys: []string{"*string:~*req.Account:1001", "*string:~*req.Account:1002"},
			itemIDs: utils.StringMap{
				"AP1": true,
				"AP2": true,
			},
			want: nil,
		},
		{
			name: "remove multiple filter indexes from mixed profiles",
			idx: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
					"AP3": true,
				},
				"*string:~*req.Destination:1010": {
					"AP2": true,
					"AP3": true,
				},
				"*string:~*req.Destination:1011": {
					"AP1": true,
					"AP3": true,
					"AP4": true,
				},
				"*string:~*req.Destination:1012": {
					"AP2": true,
				},
			},
			keys: []string{"*string:~*req.Destination:1010", "*string:~*req.Destination:1012"},
			itemIDs: utils.StringMap{
				"AP2": true,
				"AP3": true,
			},
			want: map[string]utils.StringMap{
				"*string:~*req.Account:1001": {
					"AP1": true,
					"AP2": true,
					"AP3": true,
				},
				"*string:~*req.Destination:1011": {
					"AP1": true,
					"AP3": true,
					"AP4": true,
				},
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
			if err := dm.SetFilterIndexes(utils.CacheAttributeFilterIndexes, tntCtx, test.idx, true, ""); err != nil {
				t.Fatalf("dm.SetFilterIndexes() returned unexpected error: %v", err)
			}
			if err := removeFilterIndexesForFilter(dm, utils.CacheAttributeFilterIndexes, utils.CacheAttributeProfiles,
				tntCtx, test.keys, test.itemIDs); err != nil {
				t.Fatalf("removeFilterIndexesForFilter() returned unexpected error: %v", err)
			}
			got, err := dm.GetFilterIndexes(utils.CacheAttributeFilterIndexes, tntCtx, "", nil)
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
