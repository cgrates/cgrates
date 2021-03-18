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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestABDebitUsageFromConcretes1(t *testing.T) {
	aB := &abstractBalance{
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:   "CB1",
					Type: utils.MetaConcrete,
					Opts: map[string]interface{}{
						utils.MetaBalanceLimit: -200.0,
					},
					UnitFactors: []*utils.UnitFactor{
						{
							Factor: utils.NewDecimal(100, 0), // EuroCents
						},
					},
					Units: utils.NewDecimal(500, 0), // 500 EuroCents
				},
			},
			{
				blnCfg: &utils.Balance{
					ID:   "CB2",
					Type: utils.MetaConcrete,
					Opts: map[string]interface{}{
						utils.MetaBalanceLimit: -1.0,
					},
					Units: utils.NewDecimal(125, 2),
				},
			},
		},
	}

	// consume only from first balance
	if _, err := debitConcreteUnits(decimal.New(int64(time.Duration(5*time.Minute)), 0),
		utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err == nil || err != utils.ErrInsufficientCredit {
		t.Errorf("Expected %+v, received %+v", utils.ErrInsufficientCredit, err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(500, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(125, 2)) != 0 {
		t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}
	/*
		if _, err := debitConcreteUnits(decimal.New(int64(time.Duration(9*time.Minute)), 0),
			utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err == nil || err != utils.ErrInsufficientCredit {
			t.Errorf("Expected %+v, received %+v", utils.ErrInsufficientCredit, err)
		} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(500, 0)) != 0 {
			t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
		} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(125, 2)) != 0 {
			t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
		}

		if _, err := debitConcreteUnits(decimal.New(int64(time.Duration(10*time.Minute)), 0),
			utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err == nil || err != utils.ErrInsufficientCredit {
			t.Error(err)
		} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(500, 0)) != 0 {
			t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
		} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(125, 2)) != 0 {
			t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
		}

		expectedEvCharg := &utils.EventCharges{
			Concretes:   utils.NewDecimal(925, 2),
			Accounting:  make(map[string]*utils.AccountCharge),
			UnitFactors: make(map[string]*utils.UnitFactor),
			Rating:      make(map[string]*utils.RateSInterval),
		}
		if evCh, err := debitConcreteUnits(decimal.New(925, 2),
			utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err != nil {
			t.Error(err)
		} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(-200, 0)) != 0 {
			t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
		} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(-1, 0)) != 0 {
			t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
		} else if !reflect.DeepEqual(evCh, expectedEvCharg) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCharg), utils.ToJSON(evCh))
		}

	*/
}

func TestABDebitAbstracts(t *testing.T) {
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

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(20, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}

	// limited by concrete
	aB.blnCfg.Units = utils.NewDecimal(int64(time.Duration(60*time.Second)), 0)
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(29, 0) // not enough concrete

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(29*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(31*time.Second)), 0)) != 0 { // used 29 units
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}

	// limited by concrete
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(0, 0) // not enough concrete

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(31*time.Second), 0)) != 0 { // same as above
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 { // same as above
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}

	// limited by abstract
	aB.blnCfg.Units = utils.NewDecimal(int64(time.Duration(29*time.Second)), 0) // not enough abstract
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(60, 0)

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(29*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 { // should be all used
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(31, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCost0WithConcrete(t *testing.T) {
	// consume units only from abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(10, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCost0WithoutConcrete(t *testing.T) {
	// consume units only from abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	}
}

func TestABCost0Exceed(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(70*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(60*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(10, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCost0ExceedWithoutConcrete(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(70*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(60*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	}
}

func TestABCost0WithUnlimitedWithConcrete(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(80*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(80*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(-int64(time.Duration(20*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(10, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCost0WithLimit(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			UnitFactors: []*utils.UnitFactor{
				{
					Factor: utils.NewDecimal(int64(2*time.Second), 0),
				},
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(30*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Expected %+v, received %+v", decimal.New(int64(30*time.Second), 0), ec.Abstracts)
	}
}

func TestABCost0WithLimitWithConcrete(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(30*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(10, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCost0WithLimitExceed(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(50*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(30*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	}
}

func TestABCost0WithLimitExceedWithConcrete(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(50*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(30*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(10, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestDebitUsageFiltersError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(cfg, nil, nil)
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:        "ID_TEST",
			Type:      utils.MetaAbstract,
			FilterIDs: []string{"*string:*~req.Usage:10s"},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(1*time.Second), 0),
					RecurrentFee: utils.NewDecimal(2, 0),
				},
			},
			Units: utils.NewDecimal(int64(50*time.Second), 0),
		},
		fltrS: filters,
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}
	_, err := aB.debitAbstracts(decimal.New(int64(40*time.Second), 0),
		cgrEv)
	if err == nil || err != utils.ErrFilterNotPassingNoCaps {
		t.Errorf("Expected %+v, received %+v", utils.ErrFilterNotPassingNoCaps, err)
	}

	aB.blnCfg.FilterIDs = []string{"invalid_filter_format"}
	_, err = aB.debitAbstracts(decimal.New(int64(40*time.Second), 0),
		cgrEv)
	if err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}
}

func TestDebitUsageBalanceLimitErrors(t *testing.T) {
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:   "ID_TEST",
			Type: utils.MetaAbstract,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: "not_FLOAT64",
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(1*time.Second), 0),
					RecurrentFee: utils.NewDecimal(2, 0),
				},
			},
			Units: utils.NewDecimal(int64(60*time.Second), 0),
		},
		fltrS: new(engine.FilterS),
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}

	expectedErr := "unsupported *balanceLimit format"
	_, err := aB.debitAbstracts(decimal.New(int64(40*time.Second), 0),
		cgrEv)
	if err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	aB.blnCfg.Opts[utils.MetaBalanceLimit] = float64(16 * time.Second)
	if _, err = aB.debitAbstracts(decimal.New(int64(40*time.Second), 0),
		cgrEv); err != nil {
		t.Error(err)
	}
	if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(60*time.Second), 0)) != 0 {
		t.Errorf("Expected %+v, received %+v", aB.blnCfg.Units.Big, utils.NewDecimal(int64(50*time.Second), 0))
	}
}

func TestDebitUsageUnitFactorsErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(cfg, nil, nil)
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:   "ID_TEST",
			Type: utils.MetaAbstract,
			UnitFactors: []*utils.UnitFactor{
				{
					FilterIDs: []string{"invalid_filter_fromat"},
					Factor:    utils.NewDecimal(2, 0),
				},
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(1*time.Second), 0),
					RecurrentFee: utils.NewDecimal(2, 0),
				},
			},
			Units: utils.NewDecimal(int64(60*time.Second), 0),
		},
		fltrS: filters,
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}

	if _, err := aB.debitAbstracts(decimal.New(int64(20*time.Second), 0), cgrEv); err == nil ||
		err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}

	aB.blnCfg.UnitFactors[0].FilterIDs = []string{"*string:*~req.Usage:10s"}
	if ec, err := aB.debitAbstracts(decimal.New(int64(20*time.Second), 0), cgrEv); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(0, 0)) != 0 {
		t.Error(err)
	}
}

func TestDebitUsageCostIncrementError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(cfg, nil, nil)
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:   "ID_TEST",
			Type: utils.MetaAbstract,
			CostIncrements: []*utils.CostIncrement{
				{
					FilterIDs:    []string{"INVALID_FILTER_FORMAT"},
					Increment:    utils.NewDecimal(int64(1*time.Second), 0),
					RecurrentFee: utils.NewDecimal(2, 0),
				},
			},
			Units: utils.NewDecimal(int64(60*time.Second), 0),
		},
		fltrS: filters,
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Usage: "10s",
		},
	}

	if _, err := aB.debitAbstracts(decimal.New(int64(20*time.Second), 0), cgrEv); err == nil ||
		err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}

	//Will check the error by making the event charge
	//the cost is unknown, will use attributes to query from rates
	aB.blnCfg.CostIncrements = nil
	aB.blnCfg.AttributeIDs = []string{"attr11"}
	expected := "NOT_CONNECTED: AttributeS"
	if _, err := aB.debitAbstracts(decimal.New(int64(20*time.Second), 0), cgrEv); err == nil ||
		err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestABCost(t *testing.T) {
	// debit 10 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(10*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(10*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(50*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(9, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostWithFiltersNotMatch(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// we expect to receive an error because it will try calculate the cost from rates
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					FilterIDs:    []string{"*string:~*req.CustomField:CustomValue"},
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: filterS,
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EV",
		Event: map[string]interface{}{
			"CustomField2": "CustomValue2",
		},
	}
	if _, err := aB.debitAbstracts(decimal.New(int64(10*time.Second), 0),
		cgrEv); err == nil || err.Error() != "RATES_ERROR:NOT_CONNECTED: RateS" {
		t.Error(err)
	}
}

func TestABCostWithFilters(t *testing.T) {
	// debit 10 seconds with cost of 0.1 per second
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					FilterIDs:    []string{"*string:~*req.CustomField:CustomValue"},
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: filterS,
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EV",
		Event: map[string]interface{}{
			"CustomField": "CustomValue",
		},
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(10*time.Second), 0),
		cgrEv); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(10*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(50*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(9, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostExceed(t *testing.T) {
	// debit 70 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(70*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(60*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(4, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostUnlimitedExceed(t *testing.T) {
	// debit 70 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(70*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(70*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(-int64(time.Duration(10*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(3, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostLimit(t *testing.T) {
	// debit 70 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(30*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(7, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostLimitExceed(t *testing.T) {
	// debit 70 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(70*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(30*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(7, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostNotEnoughConcrete(t *testing.T) {
	// debit 55 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(5, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(55*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(50*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(10*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostMultipleConcrete(t *testing.T) {
	// debit 55 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB1",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(5, 0),
				},
			},
			{
				blnCfg: &utils.Balance{
					ID:    "CB2",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(5, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(55*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(55*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(5*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(45, 1)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}
}

func TestABCostMultipleConcreteUnlimited(t *testing.T) {
	// debit 55 seconds with cost of 0.1 per second
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(time.Duration(60*time.Second)), 0), // 1 Minute
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Duration(time.Second)), 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB1",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(5, 0),
				},
			},
			{
				blnCfg: &utils.Balance{
					ID:    "CB2",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(5, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(70*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(70*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(-int64(time.Duration(10*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(3, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}
}

func TestAMCostWithUnitFactor(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "ID_TEST",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(60*time.Second), 0),
			UnitFactors: []*utils.UnitFactor{
				{
					Factor: utils.NewDecimal(2, 0),
				},
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(1*time.Second), 0),
					RecurrentFee: utils.NewDecimal(1, 0),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:    "CB1",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(50, 0),
				},
			},
		},
		fltrS: filterS,
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EV",
		Event: map[string]interface{}{
			"CustomField": "CustomValue",
		},
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(10*time.Second), 0),
		cgrEv); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(20*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(time.Duration(40*time.Second)), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(30, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}
