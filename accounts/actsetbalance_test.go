/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/ericlagergren/decimal"
)

func TestActSetAccountBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)

	acntID := "TestActSetAccount"
	diktats := []*utils.BalDiktat{
		{
			Path:  "*balance.Concrete1",
			Value: ";10",
		},
	}

	expected := utils.ErrNoDatabaseConn.Error()
	if err := actSetAccount(nil, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	expected = "WRONG_PATH"
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*balance.Concrete1.NOT_A_FIELD"

	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*balance.Concrete1.Weights"

	expectedAcc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     acntID,
		Balances: map[string]*utils.Balance{
			"Concrete1": {
				ID:    "Concrete1",
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				Weights: []*utils.DynamicWeight{
					{
						Weight: 10,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
					{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    utils.NewDecimal(1024*1024, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
					{
						FilterIDs:    []string{"*string:~*req.ToR:*sms"},
						Increment:    utils.NewDecimal(1, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err != nil {
		t.Error(err)
	} else if rcv, err := dm.GetAccount("cgrates.org", acntID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expectedAcc) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedAcc), utils.ToJSON(rcv))
	}
}

func TestActSetAccount(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)

	acntID := "TestActSetAccount"
	diktats := []*utils.BalDiktat{
		{
			Path:  "*accountFilterIDs",
			Value: "10",
		},
	}

	expected := "WRONG_PATH"
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*account"

	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Path = "*account.Weights"

	expected = "invalid DynamicWeight format for string <10>"
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	diktats[0].Value = ";10"

	expectedAcc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     acntID,
		Weights: []*utils.DynamicWeight{
			{
				Weight: 10,
			},
		},
	}
	if err := actSetAccount(dm, "cgrates.org", acntID, diktats, false); err != nil {
		t.Error(err)
	} else if rcv, err := dm.GetAccount("cgrates.org", acntID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expectedAcc) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedAcc), utils.ToJSON(rcv))
	}
}

func TestActSetAccountFields(t *testing.T) {
	accPrf := &utils.Account{}

	expectedAccprf := &utils.Account{
		FilterIDs: []string{"*string:~*req.ToR:*sms", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z|2014-08-29T15:00:00Z"},
		Weights: []*utils.DynamicWeight{
			{
				Weight: 10,
			},
		},
		Opts: map[string]interface{}{
			utils.AccountField: "1004",
		},
		ThresholdIDs: []string{"TH_ID1"},
	}
	if err := actSetAccountFields(accPrf, []string{utils.FilterIDs}, "*string:~*req.ToR:*sms"); err != nil {
		t.Error(err)
	} else if err := actSetAccountFields(accPrf, []string{utils.ActivationIntervalString}, "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z"); err != nil {
		t.Error(err)
	} else if err := actSetAccountFields(accPrf, []string{utils.Weights}, ";10"); err != nil {
		t.Error(err)
	} else if err := actSetAccountFields(accPrf, []string{utils.Opts, utils.AccountField}, "1004"); err != nil {
		t.Error(err)
	} else if err := actSetAccountFields(accPrf, []string{utils.ThresholdIDs}, "TH_ID1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccprf, accPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedAccprf), utils.ToJSON(accPrf))
	}

	expected := "Unsupported time format"
	if err := actSetAccountFields(accPrf, []string{utils.ActivationIntervalString}, "not_a_time"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	expected = "WRONG_PATH"
	if err := actSetAccountFields(accPrf, []string{"not_an_account_field"}, utils.EmptyString); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestActSetBalanceFields(t *testing.T) {
	bal := &utils.Balance{}

	expectedBal := &utils.Balance{
		ID:        "TestActSetBalanceFields",
		FilterIDs: []string{"*string:~*req.ToR:*sms"},
		Weights: []*utils.DynamicWeight{
			{
				Weight: 10,
			},
		},
		Type:  utils.MetaAbstract,
		Units: &utils.Decimal{decimal.New(20, 0)},
		UnitFactors: []*utils.UnitFactor{
			{
				FilterIDs: []string{"fltr1"},
				Factor:    &utils.Decimal{decimal.New(100, 0)},
			},
		},
		Opts: map[string]interface{}{
			utils.AccountField: "1004",
		},
		CostIncrements: []*utils.CostIncrement{
			{
				FilterIDs:    []string{"fltr1"},
				Increment:    &utils.Decimal{decimal.New(1, 0)},
				FixedFee:     &utils.Decimal{decimal.New(0, 0)},
				RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
			},
		},
		AttributeIDs:   []string{"ATTR_ID"},
		RateProfileIDs: []string{"RATE_ID"},
	}
	if err := actSetBalance(bal, []string{utils.ID}, "TestActSetBalanceFields", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.FilterIDs}, "*string:~*req.ToR:*sms", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.Weights}, ";10", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.Type}, utils.MetaAbstract, true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.Units}, "20", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.UnitFactors}, "fltr1;100", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.Opts, utils.AccountField}, "1004", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.CostIncrements}, "fltr1;1;0;1", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.AttributeIDs}, "ATTR_ID", true); err != nil {
		t.Error(err)
	} else if err := actSetBalance(bal, []string{utils.RateProfileIDs}, "RATE_ID", true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(bal, expectedBal) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedBal), utils.ToJSON(bal))
	}

	expected := "WRONG_PATH"
	if err := actSetBalance(bal, []string{}, utils.EmptyString, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := actSetBalance(bal, []string{"not_a_balance_field"}, utils.EmptyString, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := actSetBalance(bal, []string{"UnitFactors[0].Factor"}, "200", true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if err := actSetBalance(bal, []string{"CostIncrements[0].Increment"}, "2", true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "can't convert <not_converting_decimal> to decimal"
	if err := actSetBalance(bal, []string{utils.Units}, "not_converting_decimal", true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestActNewUnitFactorsFromString(t *testing.T) {
	value := "OneValue"
	expected := "invalid key: <OneValue> for BalanceUnitFactors"
	if _, err := actNewUnitFactorsFromString(value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "can't convert <not_decimal> to decimal"
	value = ";not_decimal"
	if _, err := actNewUnitFactorsFromString(value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestActNewCostIncrementsFromString(t *testing.T) {
	value := "OneValue"
	expected := "invalid key: <OneValue> for BalanceCostIncrements"
	if _, err := actNewCostIncrementsFromString(value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	value = ";invalid_decimal;;"
	expected = "can't convert <invalid_decimal> to decimal"
	if _, err := actNewCostIncrementsFromString(value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	value = ";;invalid_decimal;"
	if _, err := actNewCostIncrementsFromString(value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	value = ";;;invalid_decimal"
	if _, err := actNewCostIncrementsFromString(value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestActSetUnitFactor(t *testing.T) {
	unitFctr := []*utils.UnitFactor{
		{
			Factor: &utils.Decimal{decimal.New(100, 0)},
		},
	}
	path := []string{"UnitFactors[0]", utils.Factor}
	value := "200"

	expected := []*utils.UnitFactor{
		{
			Factor: &utils.Decimal{decimal.New(200, 0)},
		},
	}
	if rcv, err := actSetUnitFactor(unitFctr, path, value); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	//change the filters
	unitFctr = []*utils.UnitFactor{
		{
			FilterIDs: []string{"fltr1"},
		},
	}
	path = []string{"UnitFactors[1]", utils.FilterIDs}
	value = "fltr2"
	expected = []*utils.UnitFactor{
		{
			FilterIDs: []string{"fltr1"},
		},
		{
			FilterIDs: []string{"fltr2"},
		},
	}
	if rcv, err := actSetUnitFactor(unitFctr, path, value); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestActSetUnitFactorErrors(t *testing.T) {
	unitFctr := []*utils.UnitFactor{
		{
			FilterIDs: []string{"fltr1"},
		},
	}
	path := []string{"UnitFactors[a]", utils.FilterIDs}
	value := "fltr2"
	expected := "strconv.Atoi: parsing \"a\": invalid syntax"
	if _, err := actSetUnitFactor(unitFctr, path, value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	path = []string{"UnitFactors[7]", utils.FilterIDs}
	expected = "WRONG_PATH"
	if _, err := actSetUnitFactor(unitFctr, path, value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	path = []string{"UnitFactors[0]", "not_a_field"}
	expected = "WRONG_PATH"
	if _, err := actSetUnitFactor(unitFctr, path, value); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestActSetCostIncrement(t *testing.T) {
	costIncr := []*utils.CostIncrement{
		{
			FilterIDs:    []string{"fltr1"},
			Increment:    &utils.Decimal{decimal.New(1, 0)},
			FixedFee:     &utils.Decimal{decimal.New(0, 0)},
			RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
		},
	}

	expected := []*utils.CostIncrement{
		{
			FilterIDs:    []string{"fltr1"},
			Increment:    &utils.Decimal{decimal.New(1, 0)},
			FixedFee:     &utils.Decimal{decimal.New(0, 0)},
			RecurrentFee: &utils.Decimal{decimal.New(2, 0)},
		},
	}
	if rcv, err := actSetCostIncrement(costIncr, []string{"CostIncrements[0]", utils.RecurrentFee}, "2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	expected[0].FilterIDs = []string{"fltr2"}
	if rcv, err := actSetCostIncrement(costIncr, []string{"CostIncrements[0]", utils.FilterIDs}, "fltr2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	expected[0].FixedFee = &utils.Decimal{decimal.New(1, 0)}
	if rcv, err := actSetCostIncrement(costIncr, []string{"CostIncrements[0]", utils.FixedFee}, "1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	expected[0].Increment = &utils.Decimal{decimal.New(2, 0)}
	if rcv, err := actSetCostIncrement(costIncr, []string{"CostIncrements[0]", utils.Increment}, "2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	expected = []*utils.CostIncrement{
		{
			FilterIDs:    []string{"fltr2"},
			Increment:    &utils.Decimal{decimal.New(2, 0)},
			FixedFee:     &utils.Decimal{decimal.New(1, 0)},
			RecurrentFee: &utils.Decimal{decimal.New(2, 0)},
		},
		{
			Increment: &utils.Decimal{decimal.New(2, 0)},
		},
	}
	if rcv, err := actSetCostIncrement(costIncr, []string{"CostIncrements[1]", utils.Increment}, "2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestActSetCostIncrementErrors(t *testing.T) {
	costIncr := make([]*utils.CostIncrement, 0)
	expected := "strconv.Atoi: parsing \"a\": invalid syntax"
	if _, err := actSetCostIncrement(costIncr, []string{"CostIncrements[a]", utils.Increment}, "2"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "WRONG_PATH"
	if _, err := actSetCostIncrement(costIncr, []string{"CostIncrements[8]", utils.Increment}, "2"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if _, err := actSetCostIncrement(costIncr, []string{"CostIncrements[0]", "not_a_field"}, "2"); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
