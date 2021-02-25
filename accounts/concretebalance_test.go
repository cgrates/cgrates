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

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

/*
func TestCBDebitUnits(t *testing.T) {
	// with limit and unit factor
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: -200.0,
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
				utils.MetaBalanceLimit: -1.0,
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
				utils.MetaBalanceLimit: 0.5, // 0.5 as limit
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
*/
/*
func TestCBSimpleDebit(t *testing.T) {
	// debit 10 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	toDebit := utils.NewDecimal(10, 0)
	if dbted, _, err := cb.debitUnits(toDebit,
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(toDebit) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(490, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitExceed(t *testing.T) {
	// debit 510 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(510, 0),
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(500, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitUnlimited(t *testing.T) {
	// debit 510 units from an unlimited concrete balance with 100 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			Units: utils.NewDecimal(100, 0),
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(510, 0),
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(510, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(-410, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitLimit(t *testing.T) {
	// debit 190 units from a concrete balance with 500 units and limit of 300
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 300.0, // 300 as limit
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	toDebit := utils.NewDecimal(190, 0)
	if dbted, _, err := cb.debitUnits(toDebit,
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(toDebit) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(310, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitLimitExceed(t *testing.T) {
	// debit 210 units from a concrete balance with 500 units and limit of 300
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 300.0, // 300 as limit
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(210, 0),
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(200, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitLimitExceed2(t *testing.T) {
	// debit 510 units from a concrete balance with 500 units but because of limit it will debit only 200
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 300.0, // 300 as limit
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(510, 0),
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(200, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithUnitFactor(t *testing.T) {
	// debit 1 unit from balance but because of unit factor it will debit 100
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			UnitFactors: []*utils.UnitFactor{
				&utils.UnitFactor{
					Factor: utils.NewDecimal(100, 0),
				},
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	toDebit := utils.NewDecimal(1, 0)
	if dbted, _, err := cb.debitUnits(toDebit,
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(toDebit) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(400, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithUnitFactorWithLimit(t *testing.T) {
	// debit 3 units from balance but because of unit factor and limit it will debit 200
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			UnitFactors: []*utils.UnitFactor{
				&utils.UnitFactor{
					Factor: utils.NewDecimal(100, 0),
				},
			},
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 300.0, // 300 as limit
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(2, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithUnitFactorWithUnlimited(t *testing.T) {
	// debit 3 units from balance but because of unit factor and limit it will debit 200
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			UnitFactors: []*utils.UnitFactor{
				&utils.UnitFactor{
					Factor: utils.NewDecimal(100, 0),
				},
			},
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(7, 0),
		"cgrates.org", utils.MapStorage{}); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(7, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(-200, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithUnitFactorWithFilters1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			UnitFactors: []*utils.UnitFactor{
				&utils.UnitFactor{
					FilterIDs: []string{"*string:~*req.CustomField:CustomValue"},
					Factor:    utils.NewDecimal(100, 0),
				},
			},
			Opts: map[string]interface{}{
				utils.MetaBalanceUnlimited: true,
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"NotMatching": "NotMatching",
	})
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(100, 0),
		"cgrates.org", mp); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(100, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(400, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithUnitFactorWithFiltersWithLimit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			UnitFactors: []*utils.UnitFactor{
				&utils.UnitFactor{
					FilterIDs: []string{"*string:~*req.CustomField:CustomValue"},
					Factor:    utils.NewDecimal(100, 0),
				},
			},
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: 300.0, // 300 as limit
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"CustomField": "CustomValue",
	})
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", mp); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(2, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithMultipleUnitFactor(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			UnitFactors: []*utils.UnitFactor{
				&utils.UnitFactor{
					FilterIDs: []string{"*string:~*req.CustomField:CustomValue"},
					Factor:    utils.NewDecimal(100, 0),
				},
				&utils.UnitFactor{
					FilterIDs: []string{"*string:~*req.CustomField2:CustomValue2"},
					Factor:    utils.NewDecimal(50, 0),
				},
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"CustomField2": "CustomValue2",
	})
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", mp); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(3, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(350, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithBalanceFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:        "CB",
			FilterIDs: []string{"*string:~*req.CustomField:CustomValue"},
			Type:      utils.MetaConcrete,
			Units:     utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"CustomField": "CustomValue",
	})
	if dbted, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", mp); err != nil {
		t.Error(err)
	} else if dbted.Compare(utils.NewDecimal(3, 0)) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(497, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}
*/
/*
func TestCBDebitWithBalanceFilterNotPassing(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:        "CB",
			FilterIDs: []string{"*string:~*req.CustomField2:CustomValue2"},
			Type:      utils.MetaConcrete,
			Units:     utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"CustomField": "CustomValue",
	})
	if _, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", mp); err == nil || err != utils.ErrFilterNotPassingNoCaps {
		t.Error(err)
	}
}
*/
/*
func TestCBDebitWithBalanceInvalidFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:        "CB",
			FilterIDs: []string{"*string"},
			Type:      utils.MetaConcrete,
			Units:     utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"CustomField": "CustomValue",
	})
	if _, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", mp); err == nil || err.Error() != "inline parse error for string: <*string>" {
		t.Error(err)
	}
}
*/
/*
func TestCBDebitWithInvalidUnitFactorFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			UnitFactors: []*utils.UnitFactor{
				&utils.UnitFactor{
					FilterIDs: []string{"*string"},
					Factor:    utils.NewDecimal(100, 0),
				},
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"CustomField": "CustomValue",
	})
	if _, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", mp); err == nil || err.Error() != "inline parse error for string: <*string>" {
		t.Error(err)
	}
}
*/
/*
func TestCBDebitWithInvalidLimit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 100 units from a balance ( the unit factor doesn't match )
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "CB",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: "invalid",
			},
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	mp := utils.MapStorage{}
	mp.Set([]string{utils.MetaReq}, map[string]interface{}{
		"CustomField": "CustomValue",
	})
	if _, _, err := cb.debitUnits(utils.NewDecimal(3, 0),
		"cgrates.org", mp); err == nil || err.Error() != "unsupported *balanceLimit format" {
		t.Error(err)
	}
}
*/
func TestCBSDebitUsage(t *testing.T) {
	// debit 10 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(5, 0),
					RecurrentFee: utils.NewDecimal(1, 0),
				},
			},
		},
		fltrS: new(engine.FilterS),
	}
	toDebit := decimal.New(10, 0)
	if dbted, err := cb.debitAbstracts(toDebit,
		new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if dbted.Usage.Big.Cmp(toDebit) != 0 {
		t.Errorf("debited: %+v", dbted)
	} else if cb.blnCfg.Units.Cmp(decimal.New(498, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb.blnCfg.Units)
	}
}

func TestCBSDebitUsageInvalidFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 10 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:        "CB",
			Type:      utils.MetaConcrete,
			Units:     utils.NewDecimal(500, 0), // 500 Units
			FilterIDs: []string{"*string"},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(5, 0),
					RecurrentFee: utils.NewDecimal(1, 0),
				},
			},
		},
		fltrS: filterS,
	}
	toDebit := decimal.New(10, 0)
	if _, err := cb.debitAbstracts(toDebit, new(utils.CGREvent)); err == nil ||
		err.Error() != "inline parse error for string: <*string>" {
		t.Error(err)
	}
}

func TestCBSDebitUsageNoMatchFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 10 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:        "CB",
			Type:      utils.MetaConcrete,
			Units:     utils.NewDecimal(500, 0), // 500 Units
			FilterIDs: []string{"*string:~*req.CustomField:CustomValue"},
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(5, 0),
					RecurrentFee: utils.NewDecimal(1, 0),
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
	toDebit := decimal.New(10, 0)
	if _, err := cb.debitAbstracts(toDebit, cgrEv); err == nil ||
		err != utils.ErrFilterNotPassingNoCaps {
		t.Error(err)
	}
}

func TestCBSDebitUsageInvalidCostIncrementFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	// debit 10 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
			CostIncrements: []*utils.CostIncrement{
				{
					FilterIDs:    []string{"*string"},
					Increment:    utils.NewDecimal(5, 0),
					RecurrentFee: utils.NewDecimal(1, 0),
				},
			},
		},
		fltrS: filterS,
	}
	toDebit := decimal.New(10, 0)
	if _, err := cb.debitAbstracts(toDebit, new(utils.CGREvent)); err == nil ||
		err.Error() != "inline parse error for string: <*string>" {
		t.Error(err)
	}
}

func TestCBSDebitUsageCoverProcessAttributes(t *testing.T) { // coverage purpose
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	engine.Cache.Clear(nil)

	sTestMock := &testMockCall{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})

	// debit 10 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(5, 0),
					RecurrentFee: utils.NewDecimal(-1, 0),
				},
			},
			AttributeIDs: []string{"CustomAttr"},
		},
		fltrS:   filterS,
		connMgr: connMgr,
	}
	toDebit := decimal.New(10, 0)
	if _, err := cb.debitAbstracts(toDebit, new(utils.CGREvent)); err == nil ||
		err.Error() != "NOT_CONNECTED: AttributeS" {
		t.Error(err)
	}
}

func TestCBSDebitUsageCoverProcessAttributes2(t *testing.T) { // coverage purpose
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	engine.Cache.Clear(nil)

	sTestMock := &testMockCall{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				rplCast, canCast := reply.(*engine.AttrSProcessEventReply)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				customEv := &engine.AttrSProcessEventReply{
					MatchedProfiles: nil,
					AlteredFields:   []string{"CustomField2"},
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "EV",
						Event: map[string]interface{}{
							"CustomField2": "CustomValue2",
						},
					},
				}
				*rplCast = *customEv
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})

	// debit 10 units from a concrete balance with 500 units
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
			CostIncrements: []*utils.CostIncrement{
				{
					Increment:    utils.NewDecimal(5, 0),
					RecurrentFee: utils.NewDecimal(-1, 0),
				},
			},
			AttributeIDs: []string{"CustomAttr"},
		},
		fltrS:      filterS,
		connMgr:    connMgr,
		attrSConns: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)},
	}
	toDebit := decimal.New(10, 0)
	if _, err := cb.debitAbstracts(toDebit, new(utils.CGREvent)); err == nil ||
		err.Error() != "RATES_ERROR:NOT_CONNECTED: RateS" {
		t.Error(err)
	}
}
