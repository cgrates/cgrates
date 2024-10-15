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
	"slices"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestRankingSCfgLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name     string
		jsonCfg  *RankingSJsonCfg
		expected RankingSCfg
	}{
		{
			name: "successful load, enabled true",
			jsonCfg: &RankingSJsonCfg{
				Enabled:     utils.BoolPointer(true),
				Stats_conns: &[]string{"conn1", "conn2"},
			},
			expected: RankingSCfg{Enabled: true, StatSConns: []string{"conn1", "conn2"}},
		},

		{
			name:     "nil jsonCfg",
			jsonCfg:  nil,
			expected: RankingSCfg{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rankingCfg RankingSCfg
			err := rankingCfg.loadFromJSONCfg(tt.jsonCfg)

			if err != nil {
				t.Errorf("loadFromJSONCfg() error = %v", err)
			}

			if len(rankingCfg.StatSConns) != len(tt.expected.StatSConns) {
				t.Errorf("StatSConns length = %d, want %d", len(rankingCfg.StatSConns), len(tt.expected.StatSConns))
			}

			for i := range rankingCfg.StatSConns {
				if rankingCfg.StatSConns[i] != tt.expected.StatSConns[i] {
					t.Errorf("StatSConns[%d] = %v, want %v", i, rankingCfg.StatSConns[i], tt.expected.StatSConns[i])
				}
			}

			if rankingCfg.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled = %v, want %v", rankingCfg.Enabled, tt.expected.Enabled)
			}
		})
	}
}

func TestDiffRankingsJsonCfg(t *testing.T) {
	tests := []struct {
		name     string
		v1       *RankingSCfg
		v2       *RankingSCfg
		expected *RankingSJsonCfg
	}{
		{
			name: "enabled diff",
			v1:   &RankingSCfg{Enabled: false, StatSConns: []string{"conn1"}},
			v2:   &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			expected: &RankingSJsonCfg{
				Enabled: utils.BoolPointer(true),
			},
		},

		{
			name: "statSConns diff",
			v1:   &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			v2:   &RankingSCfg{Enabled: true, StatSConns: []string{"conn2"}},
			expected: &RankingSJsonCfg{
				Stats_conns: utils.SliceStringPointer([]string{"conn2"}),
			},
		},

		{
			name:     "no diff",
			v1:       &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			v2:       &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			expected: &RankingSJsonCfg{},
		},

		{
			name:     "no diff",
			v1:       &RankingSCfg{Enabled: false, StatSConns: []string{"conn1", "conn2"}},
			v2:       &RankingSCfg{Enabled: false, StatSConns: []string{"conn1", "conn2"}},
			expected: &RankingSJsonCfg{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diffRankingsJsonCfg(nil, tt.v1, tt.v2)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("diffRankingsJsonCfg() = %v, want %v", result, tt.expected)
			}

			if (result.Enabled == nil && tt.expected.Enabled != nil) || (result.Enabled != nil && *result.Enabled != *tt.expected.Enabled) {
				t.Errorf("diffRankingsJsonCfg() Enabled = %v, want %v", result.Enabled, tt.expected.Enabled)
			}

		})
	}
}

func TestRankingSCfgClone(t *testing.T) {
	tests := []struct {
		name     string
		input    *RankingSCfg
		expected *RankingSCfg
	}{

		{
			name:     "clone with non-nil StatSConns",
			input:    &RankingSCfg{Enabled: true, StatSConns: []string{"conn1", "conn2"}},
			expected: &RankingSCfg{Enabled: true, StatSConns: []string{"conn1", "conn2"}},
		},

		{
			name:     "clone with nil StatSConns",
			input:    &RankingSCfg{Enabled: false, StatSConns: nil},
			expected: &RankingSCfg{Enabled: false, StatSConns: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.input.Clone()

			if clone.Enabled != tt.expected.Enabled {
				t.Errorf("Clone() Enabled = %v, want %v", clone.Enabled, tt.expected.Enabled)
			}

			if !slices.Equal(clone.StatSConns, tt.expected.StatSConns) {
				t.Errorf("Clone() StatSConns = %v, want %v", clone.StatSConns, tt.expected.StatSConns)
			}

			if tt.input.StatSConns != nil && clone.StatSConns != nil && &tt.input.StatSConns[0] == &clone.StatSConns[0] {
				t.Error("Clone() StatSConns has the same reference, expected a deep copy")
			}
		})
	}
}

func TestRankingSCfgCloneSection(t *testing.T) {
	tests := []struct {
		name     string
		input    RankingSCfg
		expected *RankingSCfg
	}{

		{
			name:     "clone with non-nil StatSConns",
			input:    RankingSCfg{Enabled: true, StatSConns: []string{"conn1", "conn2"}},
			expected: &RankingSCfg{Enabled: true, StatSConns: []string{"conn1", "conn2"}},
		},

		{
			name:     "clone with nil StatSConns",
			input:    RankingSCfg{Enabled: false, StatSConns: nil},
			expected: &RankingSCfg{Enabled: false, StatSConns: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clonedSection := tt.input.CloneSection()

			clonedRankingSCfg, ok := clonedSection.(*RankingSCfg)
			if !ok {
				t.Errorf("CloneSection() returned wrong type, got %T, want *RankingSCfg", clonedSection)
			}

			if !reflect.DeepEqual(clonedRankingSCfg, tt.expected) {
				t.Errorf("CloneSection() = %v, want %v", clonedRankingSCfg, tt.expected)
			}

			if tt.input.StatSConns != nil && clonedRankingSCfg.StatSConns != nil && &tt.input.StatSConns[0] == &clonedRankingSCfg.StatSConns[0] {
				t.Error("CloneSection() StatSConns has the same reference, expected a deep copy")
			}
		})
	}
}
