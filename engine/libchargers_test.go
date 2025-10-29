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
	"testing"
)

func TestLibChargersSort(t *testing.T) {
	tests := []struct {
		name   string
		input  ChargerProfiles
		output ChargerProfiles
	}{
		{
			"AlreadySorted",
			ChargerProfiles{
				{Weight: 5},
				{Weight: 3},
				{Weight: 1},
			},
			ChargerProfiles{
				{Weight: 5},
				{Weight: 3},
				{Weight: 1},
			},
		},
		{
			"Unsorted",
			ChargerProfiles{
				{Weight: 1},
				{Weight: 5},
				{Weight: 3},
			},
			ChargerProfiles{
				{Weight: 5},
				{Weight: 3},
				{Weight: 1},
			},
		},
		{
			"AllSameWeight",
			ChargerProfiles{
				{Weight: 2},
				{Weight: 2},
				{Weight: 2},
			},
			ChargerProfiles{
				{Weight: 2},
				{Weight: 2},
				{Weight: 2},
			},
		},
		{
			"SingleElement",
			ChargerProfiles{
				{Weight: 4},
			},
			ChargerProfiles{
				{Weight: 4},
			},
		},
		{
			"Empty",
			ChargerProfiles{},
			ChargerProfiles{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.input.Sort()
			if len(tt.input) != len(tt.output) {
				t.Errorf("expected length %d, got %d", len(tt.output), len(tt.input))
			}
			for i := range tt.input {
				if tt.input[i].Weight != tt.output[i].Weight {
					t.Errorf("at index %d, expected Weight %f, got %f", i, tt.output[i].Weight, tt.input[i].Weight)
				}
			}
		})
	}
}

func TestLibChargersTenantID(t *testing.T) {
	cp := &ChargerProfile{
		Tenant: "cgrates.org",
		ID:     "2012",
	}
	result := cp.TenantID()
	expected := "cgrates.org:2012"
	if result != expected {
		t.Errorf("TenantID() = %v, want %v", result, expected)
	}
}
