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
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestABDebitUsageFromConcrete(t *testing.T) {
	aB := &abstractBalance{
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:   "CB1",
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
			},
			{
				blnCfg: &utils.Balance{
					ID:   "CB2",
					Type: utils.MetaConcrete,
					Opts: map[string]interface{}{
						utils.MetaBalanceLimit: utils.NewDecimal(-1, 0),
					},
					Units: utils.NewDecimal(125, 2),
				},
				fltrS: new(engine.FilterS),
			},
		}}
	// consume only from first balance
	if err := aB.debitUsageFromConcrete(
		utils.NewDecimal(int64(time.Duration(5*time.Minute)), 0),
		&utils.CostIncrement{
			Increment:    utils.NewDecimal(int64(time.Duration(time.Minute)), 0),
			RecurrentFee: utils.NewDecimal(1, 0)},
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(125, 2)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}

	// consume from second also, remaining in second
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(500, 0)
	aB.cncrtBlncs[1].blnCfg.Units = utils.NewDecimal(125, 2)

	if err := aB.debitUsageFromConcrete(
		utils.NewDecimal(int64(time.Duration(9*time.Minute)), 0),
		&utils.CostIncrement{
			Increment:    utils.NewDecimal(int64(time.Duration(time.Minute)), 0),
			RecurrentFee: utils.NewDecimal(1, 0)},
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(-200, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(-75, 2)) != 0 {
		t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}

	// not enough balance
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(500, 0)
	aB.cncrtBlncs[1].blnCfg.Units = utils.NewDecimal(125, 2)

	if err := aB.debitUsageFromConcrete(
		utils.NewDecimal(int64(time.Duration(10*time.Minute)), 0),
		&utils.CostIncrement{
			Increment:    utils.NewDecimal(int64(time.Duration(time.Minute)), 0),
			RecurrentFee: utils.NewDecimal(1, 0)},
		new(utils.CGREvent)); err == nil || err != utils.ErrInsufficientCredit {
		t.Error(err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(500, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(125, 2)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}
}
