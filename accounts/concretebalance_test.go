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

package accounts

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestCBDebitUnits(t *testing.T) {
	// with limit and unit factor
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: utils.NewDecimal(-200, 0),
			},
			UnitFactors: []*utils.UnitFactor{
				{
					Factor: utils.NewDecimal(100, 0), // EuroCents
				},
			},
			Units: utils.NewDecimal(500, 0), // 500 EURcents
		},
		fltrS: new(engine.FilterS),
	}
	toDebit := utils.NewDecimal(6, 0)
	if dbted, uFctr, err := cb.debitUnits(toDebit,
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cb.blnCfg.UnitFactors[0], uFctr) {
		t.Errorf("received unit factor: %+v", uFctr)
	} else if dbted.Compare(toDebit) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(-100, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}

	//with increment and not enough balance
	cb = &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: utils.NewDecimal(-1, 0),
			},
			Units: utils.NewDecimal(125, 2), // 1.25
		},
		fltrS: new(engine.FilterS),
	}
	toDebit = utils.NewDecimal(25, 1) //2.5
	if dbted, _, err := cb.debitUnits(toDebit,
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Cmp(decimal.New(225, 2)) != 0 { // 2.25 debited
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(-1, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}

	//with increment and unlimited balance
	cb = &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			Units: utils.NewDecimal(125, 2), // 1.25
		},
		fltrS: new(engine.FilterS),
	}
	toDebit = utils.NewDecimal(25, 1) // 2.5
	if dbted, _, err := cb.debitUnits(toDebit,
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Cmp(decimal.New(25, 1)) != 0 { // debit more than available since we have unlimited
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(-125, 2)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}

	//with increment and positive limit
	cb = &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: utils.NewDecimal(5, 1), // 0.5 as limit
			},
			Units: utils.NewDecimal(125, 2), // 1.25
		},
		fltrS: new(engine.FilterS),
	}
	toDebit = utils.NewDecimal(25, 1) //2.5
	if dbted, _, err := cb.debitUnits(toDebit,
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Cmp(decimal.New(75, 2)) != 0 { // limit is 0.5
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(5, 1)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}

}
