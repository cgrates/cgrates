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

import "testing"

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
