// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
			utils.MONETARY: []*engine.Balance{
				{
					Value:  20,
					Weight: 10,
				},
			},
		},
	}
	accDP := config.NewObjectDP(acc, nil)
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
			utils.MONETARY: []*engine.Balance{
				{
					Value:  20,
					Weight: 10,
				},
			},
		},
	}
	accDP := config.NewObjectDP(acc, nil)

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
	} else if data != acc.BalanceMap[utils.MONETARY][0] {
		t.Errorf("Expected: %+v ,received: %+v", acc.BalanceMap[utils.MONETARY][0], data)
	}
}
