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

func TestActionErrors(t *testing.T) {
	tests := []struct {
		name string
		rcv  string
		exp  string
	}{
		{
			name: "resetTriggersAction",
			rcv:  resetTriggersAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  setRecurrentAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  unsetRecurrentAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  allowNegativeAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  denyNegativeAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  resetAccountAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  topupResetAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  debitResetAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  debitAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  resetCountersAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  genericDebit(nil, nil, false).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  enableAccountAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  disableAccountAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  topupZeroNegativeAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  setExpiryAction(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  publishAccount(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
		{
			name: "resetTriggersAction",
			rcv:  publishBalance(nil, nil, nil, nil).Error(),
			exp:  "nil account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rcv != tt.exp {
				t.Errorf("expected %s, receives %s", tt.exp, tt.rcv)
			}
		})
	}
}
