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

func TestABDebitUsage(t *testing.T) {
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:   "AB1",
			Type: utils.MetaAbstract,
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 0)},
			},
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:   "CB1",
					Type: utils.MetaConcrete,
					UnitFactors: []*utils.UnitFactor{
						{
							Factor: utils.NewDecimal(1, 0), // EuroCents
						},
					},
					Units: utils.NewDecimal(50, 0), // 50 EURcents
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if dbted, _, err := aB.debitUsage(utils.NewDecimal(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", dbted)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(20, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}

	// limited by concrete
	aB.blnCfg.Units = utils.NewDecimal(int64(time.Duration(60*time.Second)), 0)
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(29, 0) // not enough concrete

	if dbted, _, err := aB.debitUsage(utils.NewDecimal(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(int64(29*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", dbted)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(31*time.Second)), 0)) != 0 { // used 29 units
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}

	// limited by concrete
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(0, 0) // not enough concrete

	if dbted, _, err := aB.debitUsage(utils.NewDecimal(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", dbted)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(31*time.Second), 0)) != 0 { // same as above
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 { // same as above
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}

	// limited by abstract
	aB.blnCfg.Units = utils.NewDecimal(int64(time.Duration(29*time.Second)), 0) // not enough abstract
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(60, 0)

	if dbted, _, err := aB.debitUsage(utils.NewDecimal(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(int64(29*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", dbted)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 { // should be all used
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(31, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}

}
