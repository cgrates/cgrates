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
package general_tests

import (
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAccountNewObjectDPFieldAsInterface(t *testing.T) {
	acc := &engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MetaMonetary: []*engine.Balance{
				{
					Value:  20,
					Weight: 10,
				},
			},
		},
	}
	accDP := config.NewObjectDP(acc)
	if data, err := accDP.FieldAsInterface([]string{"BalanceMap", "*monetary[0]", "Value"}); err != nil {
		t.Error(err)
	} else if data != 20. {
		t.Errorf("Expected: %+v ,received: %+v", 20., data)
	}
	if _, err := accDP.FieldAsInterface([]string{"BalanceMap", "*monetary[1]", "Value"}); err == nil ||
		err.Error() != "index out of range" {
		t.Error(err)
	}
	if _, err := accDP.FieldAsInterface([]string{"BalanceMap", "*monetary[0]", "InexistentField"}); err == nil ||
		err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestAccountNewObjectDPFieldAsInterfaceFromCache(t *testing.T) {
	acc := &engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MetaMonetary: []*engine.Balance{
				{
					Value:  20,
					Weight: 10,
				},
			},
		},
	}
	accDP := config.NewObjectDP(acc)

	if data, err := accDP.FieldAsInterface([]string{"BalanceMap", "*monetary[0]", "Value"}); err != nil {
		t.Error(err)
	} else if data != 20. {
		t.Errorf("Expected: %+v ,received: %+v", 20., data)
	}
	// the value should be taken from cache
	if data, err := accDP.FieldAsInterface([]string{"BalanceMap", "*monetary[0]", "Value"}); err != nil {
		t.Error(err)
	} else if data != 20. {
		t.Errorf("Expected: %+v ,received: %+v", 20., data)
	}
	if data, err := accDP.FieldAsInterface([]string{"BalanceMap", "*monetary[0]"}); err != nil {
		t.Error(err)
	} else if data != acc.BalanceMap[utils.MetaMonetary][0] {
		t.Errorf("Expected: %+v ,received: %+v", acc.BalanceMap[utils.MetaMonetary][0], data)
	}
}
