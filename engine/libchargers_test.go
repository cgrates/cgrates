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
	"reflect"
	"testing"
)

func TestChargerProfileTenantID(t *testing.T) {
	profile1 := ChargerProfile{
		Tenant: "cgrates1",
		ID:     "2012",
	}
	expectedID1 := "cgrates1:2012"
	got1 := profile1.TenantID()
	if got1 != expectedID1 {
		t.Errorf("TenantID() = %v, want %v", got1, expectedID1)
	}
	profile2 := ChargerProfile{
		Tenant: "cgrates2",
		ID:     "2012",
	}
	expectedID2 := "cgrates2:2012"
	got2 := profile2.TenantID()
	if got2 != expectedID2 {
		t.Errorf("TenantID() = %v, want %v", got2, expectedID2)
	}
}
func TestChargerProfilesSort(t *testing.T) {
	tests := []struct {
		name     string
		profiles ChargerProfiles
		want     ChargerProfiles
	}{
		{
			name: "Sort descending",
			profiles: ChargerProfiles{
				{Weight: 5},
				{Weight: 10},
				{Weight: 1},
			},
			want: ChargerProfiles{
				{Weight: 10},
				{Weight: 5},
				{Weight: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.profiles.Sort()
			if !reflect.DeepEqual(tt.profiles, tt.want) {
				t.Errorf("Sort() = %v, want %v", tt.profiles, tt.want)
			}
		})
	}
}
