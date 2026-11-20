// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package accounts

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/utils"
)

func TestABDebitUsageFromConcretes1(t *testing.T) {
	aB := &abstractBalance{
		cncrtBlncs: []*concreteBalance{
			{
				blnCfg: &utils.Balance{
					ID:   "CB1",
					Type: utils.MetaConcrete,
					Opts: map[string]any{
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
					Opts: map[string]any{
						utils.MetaBalanceLimit: -1.0,
					},
					Units: utils.NewDecimal(125, 2),
				},
			},
		},
	}

	// consume only from first balance
	expectedEvCharg := &utils.EventCharges{
		Concretes: utils.NewDecimal(5, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID", //will be changed
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID": {
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(5, 0),
				BalanceLimit: utils.NewDecimal(-200, 0),
				UnitFactorID: "GENUUID",
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{
			"GENUUID": {
				Factor: utils.NewDecimal(100, 0),
			},
		},
		Rating:   make(map[string]*utils.RateSInterval),
		Rates:    make(map[string]*utils.IntervalRate),
		Accounts: make(map[string]*utils.Account),
	}
	if evCh, err := debitConcreteUnits(context.Background(), utils.NewDecimal(5, 0),
		utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(125, 2)) != 0 {
		t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	} else {
		//as the names of accounting, charges, UF are GENUUIDs generator, we will change their names for comparing
		expectedEvCharg.Charges[0].ChargingID = evCh.Charges[0].ChargingID
		expectedEvCharg.Accounting = evCh.Accounting
		expectedEvCharg.UnitFactors = evCh.UnitFactors
		if !reflect.DeepEqual(evCh, expectedEvCharg) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCharg), utils.ToJSON(evCh))
		}
	}

	//back with the main units
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(500, 0)
	expectedEvCharg = &utils.EventCharges{
		Concretes: utils.NewDecimal(9, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID", //will be changed
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2", //will be changed
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID2": {
				BalanceID:    "CB2",
				Units:        utils.NewDecimal(2, 0),
				BalanceLimit: utils.NewDecimal(-1, 0),
			},
			"GENUUID": {
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(7, 0),
				BalanceLimit: utils.NewDecimal(-200, 0),
				UnitFactorID: "GENNUUID_FACTOR",
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{
			"GENNUUID_FACTOR": {
				Factor: utils.NewDecimal(100, 0),
			},
		},
		Rating:   make(map[string]*utils.RateSInterval),
		Rates:    make(map[string]*utils.IntervalRate),
		Accounts: make(map[string]*utils.Account),
	}

	if evCh, err := debitConcreteUnits(context.Background(), utils.NewDecimal(9, 0),
		utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(-200, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(-75, 2)) != 0 {
		t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	} else {
		//as the names of accounting, charges, UF are GENUUIDs generator, we will change their names for comparing
		expectedEvCharg.Charges = evCh.Charges
		expectedEvCharg.Accounting = evCh.Accounting
		expectedEvCharg.UnitFactors = evCh.UnitFactors
		if !reflect.DeepEqual(evCh, expectedEvCharg) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCharg), utils.ToJSON(evCh))
		}
	}

	//back with the main units
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(500, 0)
	aB.cncrtBlncs[1].blnCfg.Units = utils.NewDecimal(125, 2)

	if _, err := debitConcreteUnits(context.Background(), utils.NewDecimal(int64(10*time.Minute), 0),
		utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err == nil || err != utils.ErrInsufficientCredit {
		t.Error(err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(500, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(125, 2)) != 0 {
		t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}

	expectedEvCharg = &utils.EventCharges{
		Concretes: utils.NewDecimal(925, 2),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID", //will be changed
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2", //will be changed
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID": {
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(7, 0),
				BalanceLimit: utils.NewDecimal(-200, 0),
				UnitFactorID: "GENNUUID_FACTOR",
			},
			"GENUUID2": {
				BalanceID:    "CB2",
				Units:        utils.NewDecimal(225, 2),
				BalanceLimit: utils.NewDecimal(-1, 0),
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{
			"GENNUUID_FACTOR": {
				Factor: utils.NewDecimal(100, 0),
			},
		},
		Rating:   make(map[string]*utils.RateSInterval),
		Rates:    make(map[string]*utils.IntervalRate),
		Accounts: make(map[string]*utils.Account),
	}
	if evCh, err := debitConcreteUnits(context.Background(), utils.NewDecimal(925, 2),
		utils.EmptyString, aB.cncrtBlncs, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(-200, 0)) != 0 {
		t.Errorf("Unexpected units in first balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(-1, 0)) != 0 {
		t.Errorf("Unexpected units in second balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	} else {
		//as the names of accounting, charges, UF are GENUUIDs generator, we will change their names for comparing
		expectedEvCharg.Charges = evCh.Charges
		expectedEvCharg.Accounting = evCh.Accounting
		expectedEvCharg.UnitFactors = evCh.UnitFactors
		if !reflect.DeepEqual(evCh, expectedEvCharg) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCharg), utils.ToJSON(evCh))
		}
	}

}

/*
func TestABDebitAbstracts(t *testing.T) {
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:   "AB1",
			Type: utils.MetaAbstract,
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
					RecurrentFee: utils.NewDecimal(1, 0)},
			},
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
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
	aB.blnCfg.Units = utils.NewDecimal(int64(60*time.Second), 0)
	aB.cncrtBlncs[0].blnCfg.Units = utils.NewDecimal(29, 0) // not enough concrete

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(29*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(31*time.Second), 0)) != 0 { // used 29 units
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
	aB.blnCfg.Units = utils.NewDecimal(int64(29*time.Second), 0) // not enough abstract
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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

func TestABCost0Exceed1(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceUnlimited: true,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(-int64(20*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			UnitFactors: []*utils.UnitFactor{
				{
					Factor: utils.NewDecimal(int64(2*time.Second), 0),
				},
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
					RecurrentFee: utils.NewDecimal(0, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(30*time.Second), 0),
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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

func TestABCost0WithLimitExceed(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	}
}

func TestABCost0WithLimitExceedWithConcrete(t *testing.T) {
	// consume more units that has an abstract balance
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
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
		Event: map[string]any{
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
			Opts: map[string]any{
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
		Event: map[string]any{
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
		Event: map[string]any{
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
					FilterIDs: []string{"INVALID_FILTER_FORMAT"},
					Increment: utils.NewDecimal(int64(1*time.Second), 0),
				},
			},
			Units: utils.NewDecimal(int64(60*time.Second), 0),
		},
		fltrS: filters,
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Usage: "10s",
		},
	}

	if _, err := aB.debitAbstracts(decimal.New(int64(20*time.Second), 0), cgrEv); err == nil ||
		err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}

	//Will check the error by making the event charge
	//the cost is unknown, will use attributes to query from rates
	aB.blnCfg.CostIncrements[0].FilterIDs = []string{}
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(50*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(9, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestABCostWithFiltersNotMatch(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data , _ := engine.NewInternalDB(nil, true)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
dm := engine.NewDataManager(dbCM, config.CgrConfig().CacheCfg(), nil)
dm.SetCache(engine.Cache)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// we expect to receive an error because it will try calculate the cost from rates
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					FilterIDs:    []string{"*string:~*req.CustomField:CustomValue"},
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
		Event: map[string]any{
			"CustomField2": "CustomValue2",
		},
		APIOpts: map[string]any{},
	}
	if _, err := aB.debitAbstracts(decimal.New(int64(10*time.Second), 0),
		cgrEv); err == nil || err.Error() != "RATES_ERROR:NOT_CONNECTED: RateS" {
		t.Error(err)
	}
}

func TestABCostWithFilters(t *testing.T) {
	// debit 10 seconds with cost of 0.1 per second
	cfg := config.NewDefaultCGRConfig()
	data , _ := engine.NewInternalDB(nil, true)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
dm := engine.NewDataManager(dbCM, config.CgrConfig().CacheCfg(), nil)
dm.SetCache(engine.Cache)
	filterS := engine.NewFilterS(cfg, nil, dm)
	aB := &abstractBalance{
		blnCfg: &utils.Balance{
			ID:    "AB_COST_0",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					FilterIDs:    []string{"*string:~*req.CustomField:CustomValue"},
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
		Event: map[string]any{
			"CustomField": "CustomValue",
		},
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(10*time.Second), 0),
		cgrEv); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(10*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(50*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceUnlimited: true,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(-int64(10*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30000000000.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(30*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(10*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(5*time.Second), 0)) != 0 {
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
			Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			Opts: map[string]any{
				utils.MetaBalanceUnlimited: true,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(int64(time.Second), 0),
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
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(-int64(10*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(0, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	} else if aB.cncrtBlncs[1].blnCfg.Units.Compare(utils.NewDecimal(3, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[1].blnCfg.Units)
	}
}

func TestAMCostWithUnitFactor(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data , _ := engine.NewInternalDB(nil, true)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
dm := engine.NewDataManager(dbCM, config.CgrConfig().CacheCfg(), nil)
dm.SetCache(engine.Cache)
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
		Event: map[string]any{
			"CustomField": "CustomValue",
		},
	}

	if ec, err := aB.debitAbstracts(decimal.New(int64(10*time.Second), 0),
		cgrEv); err != nil {
		t.Error(err)
	} else if ec.Abstracts.Cmp(decimal.New(int64(20*time.Second), 0)) != 0 {
		t.Errorf("Unexpected debited units: %s", ec.Abstracts)
	} else if aB.blnCfg.Units.Compare(utils.NewDecimal(int64(40*time.Second), 0)) != 0 {
		t.Errorf("Unexpected units in abstract balance: %s", aB.blnCfg.Units)
	} else if aB.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(30, 0)) != 0 {
		t.Errorf("Unexpected units in concrete balance: %s", aB.cncrtBlncs[0].blnCfg.Units)
	}
}
*/

func TestAbstractBalanceId(t *testing.T) {
	balance := &utils.Balance{
		ID: "1001",
	}

	abstractBalanceInstance := &abstractBalance{
		blnCfg: balance,
	}

	result := abstractBalanceInstance.id()

	if result != "1001" {
		t.Errorf("Expected id to be '1001', but got '%s'", result)
	}
}

func TestAbstractBalanceCostNoMatchingRates(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	connMgr := engine.NewConnManager(cfg)
	data, errdb := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	if errdb != nil {
		t.Fatal(errdb)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	cacheS := engine.NewCacheS(cfg, dm, connMgr, nil, locker)
	connMgr.SetCache(cacheS)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, connMgr, dm)

	rateSConn := make(chan birpc.ClientConnector, 1)
	rateSrv, cErr := birpc.NewServiceWithMethodsRename(rates.NewRateS(cfg, filterS, dm), utils.RateSv1, true, func(key string) (newKey string) {
		return strings.TrimPrefix(key, utils.V1Prfx)
	})
	if cErr != nil {
		t.Fatal(cErr)
	}
	rateSConn <- rateSrv
	rateConnID := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)
	connMgr.AddInternalConn(rateConnID, utils.RateSv1, rateSConn)

	rtPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.Rate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				FilterIDs:       []string{"*string:~*req.ToR:*voice"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Increment:     utils.NewDecimal(int64(time.Second), 1),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	if errSet := dm.SetRateProfile(context.Background(), rtPrf, false, true); errSet != nil {
		t.Fatal(errSet)
	}

	concrete := &concreteBalance{
		cfg:     cfg,
		acntID:  "1001",
		fltrS:   filterS,
		connMgr: connMgr,
		blnCfg: &utils.Balance{
			ID:    "ConcreteBalance1",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(100, 0),
		},
	}

	abstractBalance := &abstractBalance{
		cfg:    cfg,
		acntID: "1001",
		blnCfg: &utils.Balance{
			ID:    "AbstractBalance1",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(int64(60*time.Second), 0),
			CostIncrements: []*utils.CostIncrement{
				{
					Increment: utils.NewDecimal(int64(time.Second), 0),
				},
			},
			RateProfileIDs: []string{"RATE_1"},
		},
		cncrtBlncs: []*concreteBalance{concrete},
		fltrS:      filterS,
		connMgr:    connMgr,
		rateSConns: []string{rateConnID},
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EV1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}

	expErr := "RATES_ERROR:NOT_FOUND"
	_, err := abstractBalance.debitAbstracts(context.Background(), utils.NewDecimal(int64(10*time.Second), 0), cgrEv, utils.NewDecimal(0, 0))
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected %v \n received %v", expErr, err)
	}
}

func TestAbstractBalanceCostLimitExceed(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	connMgr := engine.NewConnManager(cfg)
	data, err := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	cacheS := engine.NewCacheS(cfg, dm, connMgr, nil, locker)
	connMgr.SetCache(cacheS)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, connMgr, dm)

	abstractBalance := &abstractBalance{
		cfg:     cfg,
		connMgr: connMgr,
		blnCfg: &utils.Balance{
			ID:    "AbstractBalance",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(60, 0),
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30.0,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(1, 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				cfg:     cfg,
				connMgr: connMgr,
				blnCfg: &utils.Balance{
					ID:    "ConcreteBalance",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
				fltrS: filterS,
			},
		},
		fltrS: filterS,
	}

	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EV1",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	ec, err := abstractBalance.debitAbstracts(context.Background(), utils.NewDecimal(40, 0), cgrEv, utils.NewDecimal(0, 0))
	if err != nil {
		t.Error(err)
	}
	if ec.Abstracts.Compare(utils.NewDecimal(30, 0)) != 0 {
		t.Errorf("Expected 30, recieved %v", ec.Abstracts)
	}
	if ec.Concretes.Compare(utils.NewDecimal(3, 0)) != 0 {
		t.Errorf("Expected 3, recieved %v", ec.Concretes)
	}
	if abstractBalance.blnCfg.Units.Compare(utils.NewDecimal(30, 0)) != 0 {
		t.Errorf("Expected 30, recieved %v", abstractBalance.blnCfg.Units)
	}
	if abstractBalance.cncrtBlncs[0].blnCfg.Units.Compare(utils.NewDecimal(7, 0)) != 0 {
		t.Errorf("Expected 7, recieved %v", abstractBalance.cncrtBlncs[0].blnCfg.Units)
	}
}

func TestAbstractBalanceCostLimitError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	connMgr := engine.NewConnManager(cfg)
	data, errDB := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	if errDB != nil {
		t.Fatal(errDB)
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	cacheS := engine.NewCacheS(cfg, dm, connMgr, nil, locker)
	connMgr.SetCache(cacheS)
	dm.SetCache(cacheS)
	filterS := engine.NewFilterS(cfg, connMgr, dm)

	abstractBalance := &abstractBalance{
		cfg:     cfg,
		connMgr: connMgr,
		blnCfg: &utils.Balance{
			ID:    "AbstractBalance",
			Type:  utils.MetaAbstract,
			Units: utils.NewDecimal(60, 0),
			Opts: map[string]any{
				utils.MetaBalanceLimit: 30,
			},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(1, 0),
					RecurrentFee: utils.NewDecimal(1, 1),
				},
			},
		},
		cncrtBlncs: []*concreteBalance{
			{
				cfg:     cfg,
				connMgr: connMgr,
				blnCfg: &utils.Balance{
					ID:    "ConcreteBalance",
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(10, 0),
				},
				fltrS: filterS,
			},
		},
		fltrS: filterS,
	}

	cgrEv := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "EV1",
		Event:   map[string]any{},
		APIOpts: map[string]any{},
	}

	expErr := "unsupported *balanceLimit format"
	_, err := abstractBalance.debitAbstracts(context.Background(), utils.NewDecimal(40, 0), cgrEv, utils.NewDecimal(0, 0))
	if err == nil || err.Error() != expErr {
		t.Errorf("Expected %v \n received %v", expErr, err)
	}
}
