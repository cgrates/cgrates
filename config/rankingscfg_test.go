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

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestRankingSCfgLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name        string
		jsnCfg      *RankingsJsonCfg
		expectedCfg RankingSCfg
		expectErr   bool
	}{
		{
			name:        "nil input",
			jsnCfg:      nil,
			expectedCfg: RankingSCfg{},
			expectErr:   false,
		},

		{
			name: "enabled true, no stats conns",
			jsnCfg: &RankingsJsonCfg{
				Enabled:     utils.BoolPointer(true),
				Stats_conns: nil,
			},
			expectedCfg: RankingSCfg{
				Enabled: true,
			},
			expectErr: false,
		},

		{
			name: "enabled false with stats conns",
			jsnCfg: &RankingsJsonCfg{
				Enabled:     utils.BoolPointer(false),
				Stats_conns: &[]string{"conn1", utils.MetaInternal},
			},
			expectedCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sgsCfg RankingSCfg
			err := sgsCfg.loadFromJSONCfg(tt.jsnCfg)

			if (err != nil) != tt.expectErr {
				t.Errorf("loadFromJSONCfg() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !reflect.DeepEqual(sgsCfg, tt.expectedCfg) {
				t.Errorf("loadFromJSONCfg() = %v, want %v", sgsCfg, tt.expectedCfg)
			}
		})
	}
}

func TestRankingSCfgClone(t *testing.T) {
	tests := []struct {
		name          string
		originalCfg   RankingSCfg
		expectedClone RankingSCfg
	}{
		{
			name: "enabled true, no stat conns",
			originalCfg: RankingSCfg{
				Enabled:    true,
				StatSConns: nil,
			},
			expectedClone: RankingSCfg{
				Enabled:    true,
				StatSConns: nil,
			},
		},

		{
			name: "enabled false, stat conns present",
			originalCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", "conn2"},
			},
			expectedClone: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", "conn2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.originalCfg.Clone()

			if !reflect.DeepEqual(clone, &tt.expectedClone) {
				t.Errorf("Clone() = %v, want %v", clone, &tt.expectedClone)
			}

			if clone.StatSConns != nil && tt.originalCfg.StatSConns != nil && &clone.StatSConns[0] == &tt.originalCfg.StatSConns[0] {
				t.Errorf("StatSConns points to the same slice, expected a deep copy")
			}
		})
	}
}

func TestRankingSCfgAsMapInterface(t *testing.T) {

	tests := []struct {
		name        string
		rankingSCfg RankingSCfg
		expectedMap map[string]any
	}{
		{
			name: "enabled true, no stat conns",
			rankingSCfg: RankingSCfg{
				Enabled:    true,
				StatSConns: nil,
			},
			expectedMap: map[string]any{
				utils.EnabledCfg: true,
			},
		},

		{
			name: "enabled false, stat conns with MetaInternal",
			rankingSCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:    false,
				utils.StatSConnsCfg: []string{"conn1", utils.MetaInternal},
			},
		},

		{
			name: "stat conns without MetaInternal",
			rankingSCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", "conn2"},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:    false,
				utils.StatSConnsCfg: []string{"conn1", "conn2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rankingSCfg.AsMapInterface()

			if !reflect.DeepEqual(result, tt.expectedMap) {
				t.Errorf("AsMapInterface() = %v, want %v", result, tt.expectedMap)
			}
		})
	}
}
