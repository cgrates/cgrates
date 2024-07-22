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
	"testing"
)

func TestNewRPCClientSet(t *testing.T) {
	tests := []struct {
		name      string
		clientSet *RPCClientSet
	}{
		{
			name:      "default case",
			clientSet: NewRPCClientSet(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.clientSet == nil {
				t.Errorf("Expected RPCClientSet to be non-nil")
			}
			if tt.clientSet.set == nil {
				t.Errorf("Expected 'set' map to be initialized, got nil")
			}
			if len(tt.clientSet.set) != 0 {
				t.Errorf("Expected 'set' map to be empty, got %d items", len(tt.clientSet.set))
			}
		})
	}
}
