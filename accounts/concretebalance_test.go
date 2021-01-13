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
	if dbted, uFctr, err := cb.debitUnits(toDebit, utils.NewDecimal(1, 0),
		&utils.CGREvent{Tenant: "cgrates.org"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cb.blnCfg.UnitFactors[0], uFctr) {
		t.Errorf("received unit factor: %+v", uFctr)
	} else if dbted.Compare(toDebit) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(-100, 0)) != 0 {
		t.Errorf("balance remaining: %f", cb.blnCfg.Units)
	}
	//with increment and not enough balance
	cb = &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: utils.NewDecimal(-1, 0),
			},
			Units: utils.NewDecimal(125, 2),
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(
		utils.NewDecimal(25, 1), //2.5
		utils.NewDecimal(1, 1),  //0.1
		&utils.CGREvent{Tenant: "cgrates.org"}); err != nil {
		t.Error(err)
	} else if dbted.Cmp(decimal.New(22, 1)) != 0 { // only 1.2 is possible due to increment
		t.Errorf("debited: %s, cmp: %v", dbted, dbted.Cmp(new(decimal.Big).SetFloat64(1.2)))
	} else if cb.blnCfg.Units.Cmp(decimal.New(-95, 2)) != 0 {
		t.Errorf("balance remaining: %f", cb.blnCfg.Units)
	}
	//with increment and unlimited balance
	cb = &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			Units: &utils.Decimal{decimal.New(125, 2)},
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(
		utils.NewDecimal(25, 1), //2.5
		utils.NewDecimal(1, 1),  //0.1
		&utils.CGREvent{Tenant: "cgrates.org"}); err != nil {
		t.Error(err)
	} else if dbted.Cmp(decimal.New(25, 1)) != 0 { // only 1.2 is possible due to increment
		t.Errorf("debited: %s, cmp: %v", dbted, dbted.Cmp(new(decimal.Big).SetFloat64(1.2)))
	} else if cb.blnCfg.Units.Cmp(decimal.New(-125, 2)) != 0 {
		t.Errorf("balance remaining: %f", cb.blnCfg.Units)
	}
	//with increment and positive limit
	cb = &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: utils.NewDecimal(5, 1), // 0.5 as limit
			},
			Units: &utils.Decimal{decimal.New(125, 2)},
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(
		utils.NewDecimal(25, 1), //2.5
		utils.NewDecimal(1, 1),  //0.1
		&utils.CGREvent{Tenant: "cgrates.org"}); err != nil {
		t.Error(err)
	} else if dbted.Cmp(decimal.New(7, 1)) != 0 { // only 1.2 is possible due to increment
		t.Errorf("debited: %s, cmp: %v", dbted, dbted.Cmp(new(decimal.Big).SetFloat64(1.2)))
	} else if cb.blnCfg.Units.Cmp(decimal.New(55, 2)) != 0 {
		t.Errorf("balance remaining: %f", cb.blnCfg.Units)
	}
}
