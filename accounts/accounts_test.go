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
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/rates"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestMatchingAccountsForEventMockingErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	fltr := engine.NewFilterS(cfg, nil, nil)

	accPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "1004",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{Big: decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
					},
				},
			},
		},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
	}

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	/*
		mockDataDB := &dataDBMockErrorNotFound{}
		//if the error is NOT_FOUND, continue to match the
		newDm := engine.NewDataManager(mockDataDB, cfg.CacheCfg(), nil)
		accnts = NewAccountS(cfg, fltr, nil, newDm)
		if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
			[]string{}, false,true); err == nil || err != utils.ErrNotFound {
			t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
		}

	*/

	//mocking error in order to get from data base
	dataDB := &dataDBMockError{}
	newDm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	accnts = NewAccountS(cfg, fltr, nil, newDm)
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, false, true); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestMatchingAccountsForEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1004",
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"invalid_filter_format"},
				Weight:    20,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004", "*ai:~*req.AnswerTime:2020-07-21T00:00:00Z|2020-07-21T10:00:00Z"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",

				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{Big: decimal.New(0, 0)},
			},
		},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
	}

	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, false, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	cgrEvent.Event[utils.AccountField] = "1004"
	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	cgrEvent.APIOpts = make(map[string]any)
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, false, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	accPrf.FilterIDs = []string{"invalid_filter_format"}
	expected := "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, false, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.FilterIDs = []string{"*string:~*req.Account:1003"}

	expected = "NOT_FOUND"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, false, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.FilterIDs = []string{"*string:~*req.Account:1004"}

	expected = "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, false, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expectedAccPrfWeght := utils.AccountsWithWeight{
		{
			Account: accPrf,
			Weight:  20,
		},
	}
	if rcv, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccPrfWeght, rcv) {
		t.Errorf("Expected %+v, received %+v", expectedAccPrfWeght, utils.ToJSON(rcv))
	}
}

func TestAccountDebit(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "TestAccountDebit",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"invalid_filter_format"},
						Weight:    20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{Big: decimal.New(150, 0)},
			},
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{Big: decimal.New(200, 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
					},
				},
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TEST_EVENT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
	}

	usage := &utils.Decimal{Big: decimal.New(190, 0)}
	expected := "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].Weights[0].FilterIDs = []string{}
	accPrf.Balances["ConcreteBalance1"].Type = utils.MetaConcrete

	usage = &utils.Decimal{Big: decimal.New(0, 0)}
	if _, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err != nil {
		t.Error(err)
	}
	usage = &utils.Decimal{Big: decimal.New(190, 0)}

	accPrf.Balances["ConcreteBalance1"].UnitFactors = []*utils.UnitFactor{
		{
			FilterIDs: []string{"invalid_format_type"},
			Factor:    &utils.Decimal{Big: decimal.New(1, 0)},
		},
	}
	expected = "NOT_FOUND:invalid_format_type"
	if _, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].UnitFactors[0].FilterIDs = []string{}

	expectedUsage := &utils.Decimal{Big: decimal.New(150, 0)}
	if evCh, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(evCh.Concretes.Big, expectedUsage.Big) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedUsage.Big), utils.ToJSON(evCh.Concretes.Big))
	}
}

func TestAccountsDebitGetUsage(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accntsPrf := []*utils.AccountWithWeight{
		{
			Account: &utils.Account{
				Tenant:    "cgrates.org",
				ID:        "TestAccountsDebitGetUsage",
				FilterIDs: []string{"*prefix:~*req.Destination:+44"},
				Balances: map[string]*utils.Balance{
					"ConcreteBal1": {
						ID: "ConcreteBal1",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
						},
						Type:  utils.MetaConcrete,
						Units: &utils.Decimal{Big: decimal.New(90, 0)},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
								FixedFee:     &utils.Decimal{Big: decimal.New(2, 1)}, // 0.2
								RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
							},
						},
					},
				},
			},
		},
	}

	evChExp := &utils.EventCharges{
		Abstracts: utils.NewDecimal(89, 0),
		Concretes: utils.NewDecimal(892, 1),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "CHARGING1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"CHARGING1": {
				AccountID:       "TestAccountsDebitGetUsage",
				BalanceID:       "*mockabstract",
				Units:           utils.NewDecimal(89, 0),
				RatingID:        "RATING1",
				JoinedChargeIDs: []string{"CHARGING1_JOINEDCHARGE"},
			},
			"CHARGING1_JOINEDCHARGE": {
				AccountID:    "TestAccountsDebitGetUsage",
				BalanceID:    "ConcreteBal1",
				Units:        utils.NewDecimal(892, 1), // 89.2
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		Rating: map[string]*utils.RateSInterval{
			"RATING1": {
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rates: map[string]*utils.IntervalRate{
			"RATE1": {
				FixedFee:     utils.NewDecimal(2, 1),
				RecurrentFee: utils.NewDecimal(1, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TestAccountsDebitGetUsage": accntsPrf[0].Account,
		},
	}

	// get usage from *ratesUsage
	cgrEvent := &utils.CGREvent{
		ID:     "TEST_EVENT_get_usage",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Destination: "+445643",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsUsage: "2s",
		},
	}
	if rcv, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, false); err != nil { // debit abstract
		t.Error(err)
	} else if !rcv.Equals(evChExp) {
		t.Errorf("Expected %v, \n received %v", utils.ToJSON(evChExp), utils.ToJSON(rcv))
	}

	// get usage from *usage
	//firstly reset the account
	accntsPrf[0].Account.Balances["ConcreteBal1"].Units = utils.NewDecimal(90, 0)
	accnts = NewAccountS(cfg, fltr, nil, dm)
	cgrEvent = &utils.CGREvent{
		ID:     "TEST_EVENT_get_usage",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Destination: "+445643",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "2s",
		},
	}
	if rcv, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, false); err != nil {
		t.Error(err)
	} else if !rcv.Equals(evChExp) {
		t.Errorf("Expected %v, \n received %v", utils.ToJSON(evChExp), utils.ToJSON(rcv))
	}
}

func TestAccountsDebit(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accntsPrf := []*utils.AccountWithWeight{
		{
			Account: &utils.Account{
				Tenant:    "cgrates.org",
				ID:        "TestAccountsDebit",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Balances: map[string]*utils.Balance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type:  utils.MetaAbstract,
						Units: &utils.Decimal{Big: decimal.New(40, 0)},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
								FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
								RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
							},
						},
					},
					"ConcreteBalance2": {
						ID: "ConcreteBalance2",
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
						Type:  utils.MetaConcrete,
						Units: &utils.Decimal{Big: decimal.New(213, 0)},
					},
				},
			},
		},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TEST_EVENT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsUsage: "not_time_format",
		},
	}

	expected := "can't convert <not_time_format> to decimal"
	if _, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(cgrEvent.APIOpts, utils.OptsAccountsUsage)

	cgrEvent.APIOpts[utils.MetaUsage] = "not_time_format"
	if _, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(cgrEvent.APIOpts, utils.MetaUsage)

	if _, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, true, false); err != nil {
		t.Error(err)
	}
	cgrEvent.APIOpts[utils.MetaUsage] = "0"

	if _, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, false); err != nil {
		t.Error(err)
	}

	cgrEvent.APIOpts[utils.MetaUsage] = "55s"

	accntsPrf[0].Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{"invalid_filter_format"}
	expected = "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accntsPrf[0].Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	cgrEvent.Event[utils.Usage] = "300ns"
	if evCh, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, true, true); err != nil {
		t.Error(err)
	} else if evCh != nil {
		t.Errorf("received %+v", utils.ToJSON(evCh))
	}

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	accntsPrf[0].Balances["ConcreteBalance2"].Units = &utils.Decimal{Big: decimal.New(213, 0)}
	accnts.dm = nil
	expected = utils.ErrNoDatabaseConn.Error()
	if _, err := accnts.accountsDebit(context.Background(), accntsPrf, cgrEvent, true, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	subString := "<AccountS> error <NO_DATABASE_CONNECTION> restoring account <cgrates.org:TestAccountsDebit>"
	if rcv := buf.String(); !strings.Contains(rcv, subString) {
		t.Errorf("Expected %+q, received %+q", subString, rcv)
	}
}

func TestV1AccountsForEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1004",
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"invalid_filter_format"},
				Weight:    20,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",

				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{Big: decimal.New(0, 0)},
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
	}
	rply := make([]*utils.Account, 0)

	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter_format"
	if err := accnts.V1AccountsForEvent(context.Background(), ev, &rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	accPrf.Weights[0].FilterIDs = []string{}
	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	} else if err := accnts.V1AccountsForEvent(context.Background(), ev, &rply); err != nil {
		t.Errorf("Expected %+v, received %+v", expected, err)
	} else if !reflect.DeepEqual(rply[0], accPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rply[0]), utils.ToJSON(accPrf))
	}
}

func TestV1MaxAbstracts(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1MaxAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight:    20,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(213, 0),
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "210ns",
		},
	}
	reply := utils.EventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(accPrf.Balances, "ConcreteBalance2")

	exEvCh := utils.EventCharges{
		Abstracts: utils.NewDecimal(210, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1MaxAbstracts",
				BalanceID:    "AbstractBalance1",
				Units:        utils.NewDecimal(210, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
				RatingID:     "GENUUID_RATING",
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating: map[string]*utils.RateSInterval{
			"GENUUID_RATING": {
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"RATE1": {
				FixedFee:     utils.NewDecimal(0, 0),
				RecurrentFee: utils.NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TestV1MaxAbstracts": accPrf,
		},
	}
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1MaxAbstracts"].Balances["AbstractBalance1"].Units = utils.NewDecimal(int64(40*time.Second-210), 0)
		if !reply.Equals(&exEvCh) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}
}

func TestV1DebitAbstracts1(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1MaxAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight:    25,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(213, 0), // 213 - 27
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "27s",
		},
	}
	reply := utils.EventCharges{}

	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.EventCharges{
		Abstracts: utils.NewDecimal(int64(27*time.Second), 0),
		Concretes: utils.NewDecimal(27, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "CHARGE1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"CHARGE1": {
				AccountID:       "TestV1MaxAbstracts",
				BalanceID:       "AbstractBalance1",
				Units:           utils.NewDecimal(int64(27*time.Second), 0),
				BalanceLimit:    utils.NewDecimal(0, 0),
				RatingID:        "RATING1",
				JoinedChargeIDs: []string{"JoinedCh1"},
			},
			"JoinedCh1": {
				AccountID:    "TestV1MaxAbstracts",
				BalanceID:    "ConcreteBalance2",
				BalanceLimit: utils.NewDecimal(0, 0),
				Units:        utils.NewDecimal(27, 0),
			},
		},
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating: map[string]*utils.RateSInterval{
			"RATING1": {
				Increments: []*utils.RateSIncrement{
					{
						RateID:         "rate1",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"rate1": {
				RecurrentFee: utils.NewDecimal(1, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TestV1MaxAbstracts": accPrf,
		},
	}
	if err := accnts.V1DebitAbstracts(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1MaxAbstracts"].Balances["AbstractBalance1"].Units = utils.NewDecimal(int64(13*time.Second), 0)
		exEvCh.Accounts["TestV1MaxAbstracts"].Balances["ConcreteBalance2"].Units = utils.NewDecimal(186, 0)
		if !exEvCh.Equals(&reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}

	//now we'll check the debited account
	accPrf.Balances["AbstractBalance1"].Units = utils.NewDecimal(int64(13*time.Second), 0)
	accPrf.Balances["ConcreteBalance2"].Units = utils.NewDecimal(186, 0)
	if debitedAcc, err := accnts.dm.GetAccount(context.Background(), accPrf.Tenant, accPrf.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accPrf, debitedAcc) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(accPrf), utils.ToJSON(debitedAcc))
	}
}

func TestV1MaxConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight:    15,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(30*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "3m",
		},
	}
	reply := utils.EventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.EventCharges{
		Concretes: utils.NewDecimal(int64(31*time.Second), 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance1",
				Units:        utils.NewDecimal(int64(time.Second), 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"GENUUID2": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.NewDecimal(int64(30*time.Second), 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"TestV1DebitAbstracts": accPrf,
		},
	}
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance1"].Units = utils.NewDecimal(0, 0)
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance2"].Units = utils.NewDecimal(0, 0)
		if !exEvCh.Equals(&reply) {
			//if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}
}

func TestV1DebitConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight:    15,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(time.Minute), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(30*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: "3m",
		},
	}
	reply := utils.EventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.EventCharges{
		Concretes: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance1",
				Units:        utils.NewDecimal(60000000000, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"GENUUID2": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.NewDecimal(30000000000, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"TestV1DebitAbstracts": accPrf,
		},
	}
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance1"].Units = utils.NewDecimal(0, 0)
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance2"].Units = utils.NewDecimal(0, 0)
		if !exEvCh.Equals(&reply) {
			//if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}

	//now we will check the debited account
	rcv, err := accnts.dm.GetAccount(context.Background(), "cgrates.org", "TestV1DebitAbstracts")
	if err != nil {
		t.Error(err)
	}
	accPrf.Balances["ConcreteBalance1"].Units = &utils.Decimal{Big: decimal.New(0, 0)}
	accPrf.Balances["ConcreteBalance2"].Units = &utils.Decimal{Big: decimal.New(0, 0)}
	if !reflect.DeepEqual(rcv, accPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(accPrf), utils.ToJSON(rcv))
	}
}

func TestMultipleAccountsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := []*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "TestV1MaxAbstracts",
			Weights: []*utils.DynamicWeight{
				{
					Weight: 20,
				},
			},
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"ConcreteBalance2": {
					ID: "ConcreteBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaConcrete,
					Units: &utils.Decimal{Big: decimal.New(213, 0)},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
							FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
						},
					},
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "TestV1MaxAbstracts2",
			Weights: []*utils.DynamicWeight{
				{
					Weight: 10,
				},
			},
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"ConcreteBalance2": {
					ID: "ConcreteBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: &utils.Decimal{Big: decimal.New(213, 0)},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
							FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
						},
					},
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "TestV1MaxAbstracts3",
			Weights: []*utils.DynamicWeight{
				{
					FilterIDs: []string{"invalid_format"},
					Weight:    5,
				},
			},
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"ConcreteBalance2": {
					ID: "ConcreteBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaConcrete,
					Units: &utils.Decimal{Big: decimal.New(213, 0)},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    &utils.Decimal{Big: decimal.New(1, 0)},
							FixedFee:     &utils.Decimal{Big: decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{Big: decimal.New(1, 0)},
						},
					},
				},
			},
		},
	}

	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Usage:        "210ns",
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf[0], true); err != nil {
		t.Error(err)
	}
	if err := accnts.dm.SetAccount(context.Background(), accPrf[1], true); err != nil {
		t.Error(err)
	}
	if err := accnts.dm.SetAccount(context.Background(), accPrf[2], true); err != nil {
		t.Error(err)
	}

	expected := "NOT_FOUND:invalid_format"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", args,
		[]string{"TestV1MaxAbstracts", "TestV1MaxAbstracts2", "TestV1MaxAbstracts3"}, false, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "NOT_FOUND:invalid_format"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", args,
		[]string{"TestV1MaxAbstracts", "TestV1MaxAbstracts2", "TestV1MaxAbstracts3"}, false, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1ActionSetBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	args := &utils.ArgsActSetBalance{
		Reset: false,
	}
	var reply string

	args.AccountID = ""
	expected := "MANDATORY_IE_MISSING: [AccountID]"
	if err := accnts.V1ActionSetBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.AccountID = "TestV1ActionSetBalance"

	expected = "MANDATORY_IE_MISSING: [Diktats]"
	if err := accnts.V1ActionSetBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.Diktats = []*utils.BalDiktat{
		{
			Path:  "*balance.AbstractBalance1",
			Value: "10",
		},
	}

	expected = "WRONG_PATH"
	if err := accnts.V1ActionSetBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	args.Diktats = []*utils.BalDiktat{
		{
			Path:  "*balance.AbstractBalance1.Units",
			Value: "10",
		},
	}
	args.Tenant = "cgrates.org"
	if err := accnts.V1ActionSetBalance(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected status reply", reply)
	}

	expectedAcc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1ActionSetBalance",
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID:    "AbstractBalance1",
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{Big: decimal.New(10, 0)},
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
	if rcv, err := accnts.dm.GetAccount(context.Background(), args.Tenant, args.AccountID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAcc, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedAcc), utils.ToJSON(rcv))
	}
}

func TestV1ActionRemoveBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	//firstly we will set a balance in order to remove it
	argsSet := &utils.ArgsActSetBalance{
		AccountID: "TestV1ActionRemoveBalance",
		Tenant:    "cgrates.org",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*balance.AbstractBalance1.Units",
				Value: "10",
			},
		},
		Reset: false,
	}
	var reply string

	if err := accnts.V1ActionSetBalance(context.Background(), argsSet, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected status reply", reply)
	}

	//remove it
	args := &utils.ArgsActRemoveBalances{}

	expected := "MANDATORY_IE_MISSING: [AccountID]"
	if err := accnts.V1ActionRemoveBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.AccountID = "TestV1ActionRemoveBalance"

	expected = "MANDATORY_IE_MISSING: [BalanceIDs]"
	if err := accnts.V1ActionRemoveBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.BalanceIDs = []string{"AbstractBalance1"}

	expected = utils.ErrNoDatabaseConn.Error()
	accnts.dm = nil
	if err := accnts.V1ActionRemoveBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accnts.dm = engine.NewDataManager(data, cfg.CacheCfg(), nil)

	if err := accnts.V1ActionRemoveBalance(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected status reply", reply)
	}
}

// TestV1DebitAbstractsEventCharges is designed to cover multiple EventCharges merges
func TestV1DebitAbstractsEventCharges(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)

	// set the internal AttributeS within connMngr
	attrSConn := make(chan birpc.ClientConnector, 1)
	attrSrv, _ := birpc.NewServiceWithMethodsRename(engine.NewAttributeService(dm, fltrS, cfg), utils.AttributeSv1, true, func(key string) (newKey string) { return strings.TrimPrefix(key, utils.V1Prfx) }) // update the name of the functions
	attrSConn <- attrSrv
	cfg.AccountSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	// Set the internal rateS within connMngr
	rateSConn := make(chan birpc.ClientConnector, 1)
	rateSrv, _ := birpc.NewServiceWithMethodsRename(rates.NewRateS(cfg, fltrS, dm), utils.RateSv1, true, func(key string) (newKey string) {
		return strings.TrimPrefix(key, utils.V1Prfx)
	}) // update the name of the functions
	rateSConn <- rateSrv

	cfg.AccountSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}

	connMngr := engine.NewConnManager(cfg)
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, attrSConn)
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, rateSConn)

	// provision the data
	atrPrfl := &engine.AttributeProfile{
		Tenant: utils.CGRateSorg,
		ID:     "ATTR_ATTACH_RATES_PROFILE_RP_2",
		Attributes: []*engine.Attribute{
			{
				Path:  "*opts.RateSProfile",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("RP_2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), atrPrfl, true); err != nil {
		t.Error(err)
	}

	rtPflls := []*utils.RateProfile{
		{
			Tenant: utils.CGRateSorg,
			ID:     "RP_1",
			Rates: map[string]*utils.Rate{
				"RT_1": {
					ID: "RT_1", // lower preference, matching always
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(1, 2), // 0.01 per second
							Unit:          utils.NewDecimal(int64(time.Second), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			}},
		{
			Tenant:    utils.CGRateSorg,
			ID:        "RP_2", // higher preference, only matching with filter
			FilterIDs: []string{"*string:~*opts.RateSProfile:RP_2"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Rates: map[string]*utils.Rate{
				"RT_2": {
					ID: "RT_2",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(2, 2), // 0.02 per second
							Unit:          utils.NewDecimal(int64(time.Second), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
							FixedFee:      utils.NewDecimal(1, 1), // 0.1 for the rest of call
							Unit:          utils.NewDecimal(int64(time.Second), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			}},
	}
	for _, rtPfl := range rtPflls {
		if err := dm.SetRateProfile(context.Background(), rtPfl, false, true); err != nil {
			t.Error(err)
		}
	}

	accnts := NewAccountS(cfg, fltrS, connMngr, dm)

	ab1ID := "AB1"
	ab2ID := "AB2"
	ab3ID := "AB3"
	cb1ID := "CB1"
	cb2ID := "CB2"
	// populate the Account
	acnt1 := &utils.Account{
		Tenant: utils.CGRateSorg,
		ID:     "TestV1DebitAbstractsEventCharges1",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Balances: map[string]*utils.Balance{
			ab1ID: { // cost: 0.4 connectFee plus 0.2 per minute, available 2 minutes, should remain  10s
				ID:   ab1ID,
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 40,
					},
				},
				// RecurrentFee/Increment
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(4, 1), // 0.4
						RecurrentFee: utils.NewDecimal(2, 1), // 0.2 per minute
					},
				},
				Units: utils.NewDecimal(int64(130*time.Second), 0), // 2 Minute 10s, rest 10s
			},
			// total 0.8 needs to be debited from CB1, with UnitFactor will be 0.8 * 100
			cb1ID: { // paying the AB1 plus own debit of 0.1 per second, limit of -200 cents
				ID:   cb1ID,
				Type: utils.MetaConcrete,
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Opts: map[string]any{
					utils.MetaBalanceLimit: -200.0,
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 1), // 0.1 per second
					},
				},
				UnitFactors: []*utils.UnitFactor{
					{
						Factor: utils.NewDecimal(100, 0), // EuroCents
					},
				},
				Units: utils.NewDecimal(80, 0), // 80 EuroCents for the debit from AB1, rest for 20 seconds of limit
			},
			// 2m20s ABSTR, 2.8 CONCR
			ab2ID: { // continues debitting after CB1, 0 cost in increments of seconds, maximum of 1 minute
				ID:   ab2ID,
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(0, 0)},
				},
				Units: utils.NewDecimal(int64(1*time.Minute), 0), // 1 Minute, no cost
			},
			// 3m20s ABSTR, 2.8 CONCR
			ab3ID: { // not matching due to filter
				ID:        ab3ID,
				Type:      utils.MetaAbstract,
				FilterIDs: []string{"*string:*~req.Account:AnotherAccount"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 0)},
				},
				Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			},
			//3m20s ABSTR 2.8 CONCR
			cb2ID: { //125s with rating from RateS (1.25/0.01 from rates)
				ID:   cb2ID,
				Type: utils.MetaConcrete,
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				AttributeIDs: []string{utils.MetaNone},
				Units:        utils.NewDecimal(125, 2), // 1.25
			},
			//5m25s ABSTR, 4.05 CONCR
		},
	}
	if err := dm.SetAccount(context.Background(), acnt1, true); err != nil {
		t.Error(err)
	}

	acnt2 := &utils.Account{
		Tenant: utils.CGRateSorg,
		ID:     "TestV1DebitAbstractsEventCharges2",
		Balances: map[string]*utils.Balance{
			ab1ID: { // cost: 0.4 connectFee plus 0.2 per minute, available 2 minutes, should remain  10 units
				ID:   ab1ID,
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(4, 1),  // 0.4
						RecurrentFee: utils.NewDecimal(2, 1)}, // 0.2 per minute
				},
				UnitFactors: []*utils.UnitFactor{
					{
						Factor: utils.NewDecimal(1, 9), // Nanoseconds
					},
				},
				Units: utils.NewDecimal(130, 0),
			},
			//7m25s ABSTR, 4.05 CONCR
			cb1ID: { // absorb all costs, standard rating used when primary debiting
				ID:   cb1ID,
				Type: utils.MetaConcrete,
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				AttributeIDs: []string{utils.MetaNone},
				Units:        utils.NewDecimal(5, 1), //0.5 covering partially the AB1
			},
			// 7m25s ABSTR, 4.55 CONCR
			cb2ID: { // absorb all costs, standard rating used when primary debiting
				ID:   cb2ID,
				Type: utils.MetaConcrete,
				Opts: map[string]any{
					utils.MetaBalanceUnlimited: true,
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				AttributeIDs: []string{"ATTR_ATTACH_RATES_PROFILE_RP_2"},
				Units:        utils.NewDecimal(35, 2), //0.3 + 0.05 covering 5s  it's own with RateS
				// ToDo: change rating with a lower preference here, should have multiple groups also // RateProfileIDs: ["TWOCENTS"]
			},
			// 7m25s ABSTR, 4.85 CONCR to cover the AB1

			// 7m26S ABSTR, 4.95 CONCR with remaining flat covered by CB2 with RateS, RP_2, -0.05 on CB2

		},
	}
	if err := dm.SetAccount(context.Background(), acnt2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID:     "TestV1DebitAbstractsEventCharges",
		Tenant: utils.CGRateSorg,
		APIOpts: map[string]any{
			utils.MetaUsage: "7m26s",
		},
	}
	var rcvEC utils.EventCharges
	if err := accnts.V1DebitAbstracts(context.Background(), args, &rcvEC); err != nil {
		t.Error(err)
	}

	// expected EventCharges
	eEvChgs := &utils.EventCharges{
		Abstracts: utils.NewDecimal(446000000000, 0),
		Concretes: utils.NewDecimal(495, 2),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "ACCOUNT5",
				CompressFactor: 1,
			},
			{
				ChargingID:     "ACCOUNT6",
				CompressFactor: 1,
			},
			{
				ChargingID:     "ACCOUNT3",
				CompressFactor: 1,
			},
			{
				ChargingID:     "ACCOUNT4",
				CompressFactor: 1,
			},
			{
				ChargingID:     "ACCOUNT2",
				CompressFactor: 1,
			},
			{
				ChargingID:     "ACCOUNT1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"account1_joinedcharges": {
				AccountID: "TestV1DebitAbstractsEventCharges2",
				BalanceID: "CB2",
				Units:     utils.NewDecimal(1, 1),
			},
			"ACCOUNT2": {
				AccountID:       "TestV1DebitAbstractsEventCharges2",
				BalanceID:       "AB1",
				Units:           utils.NewDecimal(int64(120*time.Second), 0),
				BalanceLimit:    utils.NewDecimal(0, 0),
				UnitFactorID:    "UF3",
				RatingID:        "account2_rating",
				JoinedChargeIDs: []string{"account2_joinedcharges", "account2_joinedcharges2"},
			},
			"ACCOUNT3": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    "AB2",
				Units:        utils.NewDecimal(int64(time.Minute), 0), //6.000000000e+10,
				BalanceLimit: utils.NewDecimal(0, 0),
				RatingID:     "account3_rating",
			},
			"account4_joinedcharges": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    "CB2",
				Units:        utils.NewDecimal(125, 2),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"ACCOUNT5": {
				AccountID:       "TestV1DebitAbstractsEventCharges1",
				BalanceID:       "AB1",
				Units:           utils.NewDecimal(int64(120*time.Second), 0),
				BalanceLimit:    utils.NewDecimal(0, 0),
				RatingID:        "account5_rating",
				JoinedChargeIDs: []string{"account5_joinedcharges"},
			},
			"account6_joinedcharges": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(2, 0),
				BalanceLimit: utils.SubstractDecimal(utils.NewDecimal(200, 0), utils.NewDecimal(400, 0)), // this should be -200
				UnitFactorID: "UF1",
			},
			"ACCOUNT4": {
				AccountID:       "TestV1DebitAbstractsEventCharges1",
				BalanceID:       utils.MetaMockAbstract,
				Units:           utils.NewDecimal(int64(125*time.Second), 0), // 125s
				RatingID:        "account4_rating",
				JoinedChargeIDs: []string{"account4_joinedcharges"},
			},
			"account2_joinedcharges2": {
				AccountID: "TestV1DebitAbstractsEventCharges2",
				BalanceID: "CB2",
				Units:     utils.NewDecimal(3, 1),
			},
			"account2_joinedcharges": {
				AccountID:    "TestV1DebitAbstractsEventCharges2",
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(5, 1),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"account5_joinedcharges": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(8, 1),
				BalanceLimit: utils.SubstractDecimal(utils.NewDecimal(200, 0), utils.NewDecimal(400, 0)), // this should be -200
				UnitFactorID: "UF2",
			},
			"ACCOUNT6": {
				AccountID:       "TestV1DebitAbstractsEventCharges1",
				BalanceID:       utils.MetaMockAbstract,
				Units:           utils.NewDecimal(int64(20*time.Second), 0), // 2.000000000e+10,
				RatingID:        "account6_rating",
				JoinedChargeIDs: []string{"account6_joinedcharges"},
			},
			"ACCOUNT1": {
				AccountID:       "TestV1DebitAbstractsEventCharges2",
				BalanceID:       utils.MetaMockAbstract,
				Units:           utils.NewDecimal(int64(time.Second), 0),
				RatingID:        "account1_rating",
				JoinedChargeIDs: []string{"account1_joinedcharges"},
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{
			"UF1": {
				Factor: utils.NewDecimal(100, 0),
			},
			"UF2": {
				Factor: utils.NewDecimal(100, 0),
			},
			"UF3": {
				Factor: utils.NewDecimal(1, 9), // Nanoseconds
			},
		},
		Rating: map[string]*utils.RateSInterval{
			"account3_rating": {
				Increments: []*utils.RateSIncrement{{
					RateIntervalIndex: 0,
					RateID:            "rate1",
					CompressFactor:    1,
				}},
				CompressFactor: 1,
			},
			"account2_rating": {
				Increments: []*utils.RateSIncrement{{
					RateIntervalIndex: 0,
					RateID:            "rate4",
					CompressFactor:    1,
				}},
				CompressFactor: 1,
			},
			"account1_rating": {
				Increments: []*utils.RateSIncrement{{
					RateIntervalIndex: 0,
					RateID:            "rate5",
					CompressFactor:    1,
				}},
				CompressFactor: 1,
			},
			"account5_rating": {
				Increments: []*utils.RateSIncrement{{
					RateIntervalIndex: 0,
					RateID:            "rate6",
					CompressFactor:    1,
				}},
				CompressFactor: 1,
			},
			"account6_rating": {
				Increments: []*utils.RateSIncrement{{
					RateIntervalIndex: 0,
					RateID:            "rate3",
					CompressFactor:    1,
				}},
				CompressFactor: 1,
			},
			"account4_rating": {
				Increments: []*utils.RateSIncrement{{
					RateIntervalIndex: 0,
					RateID:            "rate2",
					CompressFactor:    1,
				}},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"rate1": {
				RecurrentFee: utils.NewDecimal(0, 0),
			},
			"rate3": {
				RecurrentFee: utils.NewDecimal(1, 1),
			},
			"rate4": {
				FixedFee:     utils.NewDecimal(4, 1),
				RecurrentFee: utils.NewDecimal(2, 1),
			},
			"rate2": {},
			"rate5": {},
			"rate6": {
				FixedFee:     utils.NewDecimal(4, 1),
				RecurrentFee: utils.NewDecimal(2, 1),
			},
		},
		Accounts: map[string]*utils.Account{
			"TestV1DebitAbstractsEventCharges1": acnt1,
			"TestV1DebitAbstractsEventCharges2": acnt2,
		},
	}

	acnt1.Balances[ab1ID].Units = utils.NewDecimal(int64(10*time.Second), 0)
	acnt1.Balances[cb1ID].Units = utils.NewDecimal(-200, 0)
	acnt1.Balances[ab2ID].Units = utils.SumDecimal(&utils.Decimal{Big: utils.NewDecimal(0, 0).Neg(utils.NewDecimal(1, 0).Big)}, utils.NewDecimal(1, 0)) // negative 0
	acnt1.Balances[cb2ID].Units = utils.NewDecimal(0, 0)
	if rcv, err := dm.GetAccount(context.Background(), acnt1.Tenant, acnt1.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acnt1) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acnt1), utils.ToJSON(rcv))
	}

	acnt2.Balances[ab1ID].Units = utils.NewDecimal(10000000000, 9)
	acnt2.Balances[cb1ID].Units = utils.NewDecimal(0, 0)
	acnt2.Balances[cb2ID].Units = utils.SubstractDecimal(utils.NewDecimal(1, 0), utils.NewDecimal(105, 2)) // -0.05
	if rcv, err := dm.GetAccount(context.Background(), acnt2.Tenant, acnt2.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acnt2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acnt2), utils.ToJSON(rcv))
	}

	// now we will change the units because both accounts were debited
	// compare these 2 eventCHarges
	eEvChgs.Accounts["TestV1DebitAbstractsEventCharges1"].Balances[ab1ID].Units = utils.NewDecimal(int64(10*time.Second), 0)
	eEvChgs.Accounts["TestV1DebitAbstractsEventCharges1"].Balances[cb1ID].Units = utils.NewDecimal(-200, 0)
	eEvChgs.Accounts["TestV1DebitAbstractsEventCharges1"].Balances[ab2ID].Units = utils.SumDecimal(&utils.Decimal{Big: utils.NewDecimal(0, 0).Neg(utils.NewDecimal(1, 0).Big)}, utils.NewDecimal(1, 0)) // negative 0
	eEvChgs.Accounts["TestV1DebitAbstractsEventCharges1"].Balances[cb2ID].Units = utils.NewDecimal(0, 0)
	eEvChgs.Accounts["TestV1DebitAbstractsEventCharges2"].Balances[ab1ID].Units = utils.NewDecimal(10000000000, 9)
	eEvChgs.Accounts["TestV1DebitAbstractsEventCharges2"].Balances[cb1ID].Units = utils.NewDecimal(0, 0)
	eEvChgs.Accounts["TestV1DebitAbstractsEventCharges2"].Balances[cb2ID].Units = utils.SubstractDecimal(utils.NewDecimal(1, 0), utils.NewDecimal(105, 2)) // -0.05
	if !eEvChgs.Equals(&rcvEC) {
		t.Errorf("expecting: %s, \nreceived: %s\n", utils.ToJSON(eEvChgs), utils.ToJSON(rcvEC))
	}
}

func TestV1DebitAbstractsEventChargesWithRefundCharges(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)

	// set the internal AttributeS within connMngr
	attrSConn := make(chan birpc.ClientConnector, 1)
	attrSrv, _ := birpc.NewServiceWithMethodsRename(engine.NewAttributeService(dm, fltrS, cfg), utils.AttributeSv1, true, func(key string) (newKey string) { return strings.TrimPrefix(key, utils.V1Prfx) }) // update the name of the functions
	attrSConn <- attrSrv
	cfg.AccountSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	// Set the internal rateS within connMngr
	rateSConn := make(chan birpc.ClientConnector, 1)
	rateSrv, _ := birpc.NewServiceWithMethodsRename(rates.NewRateS(cfg, fltrS, dm), utils.RateSv1, true, func(key string) (newKey string) {
		return strings.TrimPrefix(key, utils.V1Prfx)
	}) // update the name of the functions
	rateSConn <- rateSrv

	cfg.AccountSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}

	connMngr := engine.NewConnManager(cfg)
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, attrSConn)
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, rateSConn)

	// provision the data
	atrPrfl := &engine.AttributeProfile{
		Tenant: utils.CGRateSorg,
		ID:     "ATTR_ATTACH_RATES_PROFILE_RP_2",
		Attributes: []*engine.Attribute{
			{
				Path:  "*opts.RateSProfile",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("RP_2", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), atrPrfl, true); err != nil {
		t.Error(err)
	}

	rtPflls := []*utils.RateProfile{
		{
			Tenant: utils.CGRateSorg,
			ID:     "RP_1",
			Rates: map[string]*utils.Rate{
				"RT_1": {
					ID: "RT_1", // lower preference, matching always
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(1, 2), // 0.01 per second
							Unit:          utils.NewDecimal(int64(time.Second), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			}},
		{
			Tenant:    utils.CGRateSorg,
			ID:        "RP_2", // higher preference, only matching with filter
			FilterIDs: []string{"*string:~*opts.RateSProfile:RP_2"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Rates: map[string]*utils.Rate{
				"RT_2": {
					ID: "RT_2",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(2, 2), // 0.02 per second
							Unit:          utils.NewDecimal(int64(time.Second), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
							FixedFee:      utils.NewDecimal(1, 1), // 0.1 for the rest of call
							Unit:          utils.NewDecimal(int64(time.Second), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			}},
	}
	for _, rtPfl := range rtPflls {
		if err := dm.SetRateProfile(context.Background(), rtPfl, false, true); err != nil {
			t.Error(err)
		}
	}

	accnts := NewAccountS(cfg, fltrS, connMngr, dm)

	ab1ID := "AB1"
	ab2ID := "AB2"
	ab3ID := "AB3"
	cb1ID := "CB1"
	cb2ID := "CB2"
	// populate the Account
	acnt1 := &utils.Account{
		Tenant: utils.CGRateSorg,
		ID:     "TestV1DebitAbstractsEventCharges1",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Balances: map[string]*utils.Balance{
			ab1ID: { // cost: 0.4 connectFee plus 0.2 per minute, available 2 minutes, should remain  10s
				ID:   ab1ID,
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 40,
					},
				},
				// RecurrentFee/Increment
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(4, 1), // 0.4
						RecurrentFee: utils.NewDecimal(2, 1), // 0.2 per minute
					},
				},
				Units: utils.NewDecimal(int64(130*time.Second), 0), // 2 Minute 10s, rest 10s
			},
			// total 0.8 needs to be debited from CB1, with UnitFactor will be 0.8 * 100
			cb1ID: { // paying the AB1 plus own debit of 0.1 per second, limit of -200 cents
				ID:   cb1ID,
				Type: utils.MetaConcrete,
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Opts: map[string]any{
					utils.MetaBalanceLimit: -200.0,
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 1), // 0.1 per second
					},
				},
				UnitFactors: []*utils.UnitFactor{
					{
						Factor: utils.NewDecimal(100, 0), // EuroCents
					},
				},
				Units: utils.NewDecimal(800, 1), // 80 EuroCents for the debit from AB1, rest for 20 seconds of limit
			},
			// 2m20s ABSTR, 2.8 CONCR
			ab2ID: { // continues debitting after CB1, 0 cost in increments of seconds, maximum of 1 minute
				ID:   ab2ID,
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(0, 0)},
				},
				Units: utils.NewDecimal(int64(1*time.Minute), 0), // 1 Minute, no cost
			},
			// 3m20s ABSTR, 2.8 CONCR
			ab3ID: { // not matching due to filter
				ID:        ab3ID,
				Type:      utils.MetaAbstract,
				FilterIDs: []string{"*string:*~req.Account:AnotherAccount"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 0)},
				},
				Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
			},
			//3m20s ABSTR 2.8 CONCR
			cb2ID: { //125s with rating from RateS (1.25/0.01 from rates)
				ID:   cb2ID,
				Type: utils.MetaConcrete,
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				AttributeIDs: []string{utils.MetaNone},
				Units:        utils.NewDecimal(125, 2), // 1.25
			},
			//5m25s ABSTR, 4.05 CONCR
		},
	}
	if err := dm.SetAccount(context.Background(), acnt1, true); err != nil {
		t.Error(err)
	}

	acnt2 := &utils.Account{
		Tenant: utils.CGRateSorg,
		ID:     "TestV1DebitAbstractsEventCharges2",
		Balances: map[string]*utils.Balance{
			ab1ID: { // cost: 0.4 connectFee plus 0.2 per minute, available 2 minutes, should remain  10 units
				ID:   ab1ID,
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(4, 1),  // 0.4
						RecurrentFee: utils.NewDecimal(2, 1)}, // 0.2 per minute
				},
				UnitFactors: []*utils.UnitFactor{
					{
						Factor: utils.NewDecimal(1, 9), // Nanoseconds
					},
				},
				Units: utils.NewDecimal(130000000000, 9), // this is 130 units (130.000000000) for comparing this test
			},
			//7m25s ABSTR, 4.05 CONCR
			cb1ID: { // absorb all costs, standard rating used when primary debiting
				ID:   cb1ID,
				Type: utils.MetaConcrete,
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				AttributeIDs: []string{utils.MetaNone},
				Units:        utils.NewDecimal(5, 1), //0.5 covering partially the AB1
			},
			// 7m25s ABSTR, 4.55 CONCR
			cb2ID: { // absorb all costs, standard rating used when primary debiting
				ID:   cb2ID,
				Type: utils.MetaConcrete,
				Opts: map[string]any{
					utils.MetaBalanceUnlimited: true,
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				AttributeIDs: []string{"ATTR_ATTACH_RATES_PROFILE_RP_2"},
				Units:        utils.NewDecimal(35, 2), //0.3 + 0.05 covering 5s  it's own with RateS
				// ToDo: change rating with a lower preference here, should have multiple groups also // RateProfileIDs: ["TWOCENTS"]
			},
			// 7m25s ABSTR, 4.85 CONCR to cover the AB1

			// 7m26S ABSTR, 4.95 CONCR with remaining flat covered by CB2 with RateS, RP_2, -0.05 on CB2

		},
	}
	if err := dm.SetAccount(context.Background(), acnt2, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID:     "TestV1DebitAbstractsEventCharges",
		Tenant: utils.CGRateSorg,
		APIOpts: map[string]any{
			utils.MetaUsage: "7m26s",
		},
	}
	var rcvEC utils.EventCharges
	if err := accnts.V1DebitAbstracts(context.Background(), args, &rcvEC); err != nil {
		t.Error(err)
	}

	// as we debited our accounts, we should compare our costs (abstracts and concretes)

	abstracts := utils.NewDecimal(int64(7*time.Minute+26*time.Second), 0)
	concretes := utils.NewDecimal(495, 2)
	if !reflect.DeepEqual(rcvEC.Abstracts, abstracts) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(abstracts), utils.ToJSON(rcvEC.Abstracts))
	}
	if !reflect.DeepEqual(rcvEC.Concretes, concretes) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(concretes), utils.ToJSON(rcvEC.Concretes))
	}

	// our debit was correct, now we will refund all those costs by those charges
	var replyRefund string
	argsRefund := &utils.APIEventCharges{
		Tenant:       "cgrates.org",
		EventCharges: &rcvEC,
	}
	if err := accnts.V1RefundCharges(context.Background(), argsRefund, &replyRefund); err != nil {
		t.Error(err)
	} else if replyRefund != utils.OK {
		t.Errorf("Unexpected reply returned: %v", replyRefund)
	}

	// get both accounts and compare if those units are added back properly]
	if rcv, err := dm.GetAccount(context.Background(), acnt1.Tenant, acnt1.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acnt1) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acnt1), utils.ToJSON(rcv))
	}

	if rcv, err := dm.GetAccount(context.Background(), acnt2.Tenant, acnt2.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acnt2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acnt2), utils.ToJSON(rcv))
	}
}

func TestV1DebitAbstractsWithRecurrentFeeNegative(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltrS, nil, dm)

	acnt := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstractsWithRecurrentFeeNegative",
		Balances: map[string]*utils.Balance{
			"ab1": {
				ID:   "ab1",
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 0)}, // 1.0 per minute
				},
				Units: utils.NewDecimal(int64(40*time.Second), 0),
			},
			"cb1": {
				ID:   "cb1",
				Type: utils.MetaConcrete,
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(-1, 0),
					},
				},
				Units: utils.NewDecimal(1, 0),
			},
		},
	}
	if err := dm.SetAccount(context.Background(), acnt, true); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		ID:     "TestV1DebitAbstractsWithRecurrentFeeNegative",
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaUsage: "30s",
		},
	}
	expEvCh := &utils.EventCharges{
		Abstracts: utils.NewDecimal(int64(30*time.Second), 0),
		Concretes: utils.NewDecimal(-28, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "CHARGE1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "CHARGE2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"CHARGE1": {
				AccountID:       "TestV1DebitAbstractsWithRecurrentFeeNegative",
				BalanceID:       "ab1",
				Units:           utils.NewDecimal(int64(time.Second), 0),
				BalanceLimit:    utils.NewDecimal(0, 0),
				RatingID:        "RATING1",
				JoinedChargeIDs: []string{"JOINED1"},
			},
			"JOINED1": {
				AccountID:    "TestV1DebitAbstractsWithRecurrentFeeNegative",
				BalanceID:    "cb1",
				Units:        utils.NewDecimal(1, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"CHARGE2": {
				AccountID:       "TestV1DebitAbstractsWithRecurrentFeeNegative",
				BalanceID:       utils.MetaMockAbstract,
				Units:           utils.NewDecimal(int64(29*time.Second), 0),
				RatingID:        "RATING2",
				JoinedChargeIDs: []string{"JOINED2"},
			},
			"JOINED2": {
				AccountID:    "TestV1DebitAbstractsWithRecurrentFeeNegative",
				BalanceID:    "cb1",
				Units:        utils.NewDecimal(-29, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating: map[string]*utils.RateSInterval{
			"RATING1": {
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "rate1",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			"RATING2": {
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "rate2",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"rate1": {
				RecurrentFee: utils.NewDecimal(1, 0),
			},
			"rate2": {
				RecurrentFee: utils.NewDecimal(-1, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TestV1DebitAbstractsWithRecurrentFeeNegative": acnt,
		},
	}
	ev := &utils.EventCharges{}
	if err := accnts.V1DebitAbstracts(context.Background(), args, ev); err != nil {
		t.Error(err)
	} else {
		expEvCh.Accounts["TestV1DebitAbstractsWithRecurrentFeeNegative"].Balances["ab1"].Units = utils.NewDecimal(int64(39*time.Second), 0)
		expEvCh.Accounts["TestV1DebitAbstractsWithRecurrentFeeNegative"].Balances["cb1"].Units = utils.NewDecimal(29, 0)
		if !ev.Equals(expEvCh) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expEvCh), utils.ToJSON(ev))
		}
	}

	acnt.Balances["ab1"].Units = utils.NewDecimal(int64(39*time.Second), 0)
	acnt.Balances["cb1"].Units = utils.NewDecimal(29, 0)
	if rcv, err := dm.GetAccount(context.Background(), acnt.Tenant, acnt.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acnt) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acnt), utils.ToJSON(rcv))
	}
}

func TestDebitAbstractsMaxDebitAbstractFromConcreteNoConcrBal(t *testing.T) {
	// this test will call maxDebitAbstractsFromConcretes but without any concreteBal and not calling rates
	cache := engine.Cache
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)

	acnts := NewAccountS(cfg, filterS, nil, dm)
	acnt := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstractsWithRecurrentFeeNegative",
		Balances: map[string]*utils.Balance{
			"ab1": {
				ID:    "ab1",
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(60*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						FixedFee:  utils.NewDecimal(1, 1),
						Increment: utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}
	if err := dm.SetAccount(context.Background(), acnt, true); err != nil {
		t.Error(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EV",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:    2 * time.Minute,
			utils.MetaAccounts: true,
		},
	}

	expEcCh := utils.EventCharges{
		Abstracts:   utils.NewDecimal(0, 0),
		Accounting:  make(map[string]*utils.AccountCharge),
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating:      make(map[string]*utils.RateSInterval),
		Rates:       make(map[string]*utils.IntervalRate),
		Accounts: map[string]*utils.Account{
			"TestV1DebitAbstractsWithRecurrentFeeNegative": acnt,
		},
	}

	// not having concrBal and connection to rates, this will not perform a debit, so the EventChargers abstract will be empty
	var eEc utils.EventCharges
	if err := acnts.V1DebitAbstracts(context.Background(), cgrEv, &eEc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eEc, expEcCh) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expEcCh), utils.ToJSON(eEc))
	}

	engine.Cache = cache
}

func TestDebitAbstractUsingRatesWithRoundByIncrement(t *testing.T) {
	// get the config
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Enabled = true
	cfg.AccountSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}
	//cfg.AccountSCfg().IndexedSelects = false
	cfg.RateSCfg().Enabled = true

	// get the connMngr
	connMngr := engine.NewConnManager(cfg)

	// data manager
	dm := engine.NewDataManager(engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), connMngr)

	// configure filters
	fltrs := engine.NewFilterS(cfg, connMngr, dm)

	// add the internal connection between accounts and rates
	ratesConns := make(chan birpc.ClientConnector, 1)
	rateSrv, err := birpc.NewServiceWithMethodsRename(rates.NewRateS(cfg, fltrs, dm), utils.RateSv1, true, func(key string) (newKey string) {
		return strings.TrimPrefix(key, utils.V1Prfx)
	})
	if err != nil {
		t.Error(err)
	}
	ratesConns <- rateSrv
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, ratesConns)

	//create the accounts obj
	acnts := NewAccountS(cfg, fltrs, connMngr, dm)

	// set an AccountProfile with balances that contains connection with RateProfiles
	accPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "ACNT1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"ABS1": {
				ID: "ABS1",
				Weights: utils.DynamicWeights{
					{
						Weight: 15,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(30*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
						FixedFee:  utils.NewDecimal(0, 0),
					},
				},
			},
			"ABS2": {
				ID: "ABS2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(15*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				RateProfileIDs: []string{"RP1"},
			},
			"CNCRT1": {
				ID: "CNCRT1",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(100, 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				RateProfileIDs: []string{"RP1"},
			},
		},
	}
	if err := acnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	// set the rate profile which is used in accounts
	rtPrf := &utils.RateProfile{
		ID:        "RP1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Destination:1234"},
		Rates: map[string]*utils.Rate{
			"RT1": {
				ID: "RT1",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(int64(time.Second), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rtPrf, false, true); err != nil {
		t.Error(err)
	}

	// now we will try to debit account using rates instead for the second balance
	cgrEv := &utils.CGREvent{
		ID:     "TestDebitAbstractUsingRatesWithRoundByIncrement",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1234",
		},
		APIOpts: map[string]any{
			utils.StartTime: time.Date(2020, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.MetaUsage: "44825100us",
		},
	}
	var reply utils.EventCharges
	if err := acnts.V1DebitAbstracts(context.Background(), cgrEv, &reply); err != nil {
		t.Error(err)
	}

	// verify the EvChargers
	expEvCHargers := &utils.EventCharges{
		Abstracts: utils.NewDecimal(44825100000, 0),
		Concretes: utils.NewDecimal(15, 2),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "CHARGER1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "CHARGER2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"CHARGER1": {
				AccountID:    "ACNT1",
				BalanceID:    "ABS1",
				Units:        &utils.Decimal{Big: utils.NewDecimal(int64(30*time.Second), 0).Reduce()}, // 30 seconds from *abstract ABS1
				BalanceLimit: utils.NewDecimal(0, 0),
				RatingID:     "Rating1",
			},
			"CHARGER2": {
				AccountID:       "ACNT1",
				BalanceID:       "ABS2",
				Units:           utils.NewDecimal(14825100000, 0), // 14.8251 seconds from *abstract ABS2 ((it should be debited all 15s from the entire balance))
				BalanceLimit:    utils.NewDecimal(0, 0),
				RatingID:        "Rating2",
				JoinedChargeIDs: []string{"Joined1"},
			},
			"Joined1": {
				AccountID:    "ACNT1",
				BalanceID:    "CNCRT1",
				Units:        utils.NewDecimal(15, 2), // 15s seconds from *concrete
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating: map[string]*utils.RateSInterval{
			"Rating1": {
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "RateID1",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"RateID1": {
				FixedFee: utils.NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"ACNT1": {
				Tenant:    "cgrates.org",
				ID:        "ACNT1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Balances: map[string]*utils.Balance{
					"ABS1": {
						ID: "ABS1",
						Weights: utils.DynamicWeights{
							{
								Weight: 15,
							},
						},
						Type:  utils.MetaAbstract,
						Units: utils.SumDecimal(&utils.Decimal{Big: utils.NewDecimal(0, 0).Neg(utils.NewDecimal(1, 0).Big)}, utils.NewDecimal(1, 0)), // negative 0
						CostIncrements: []*utils.CostIncrement{
							{
								Increment: utils.NewDecimal(int64(time.Second), 0),
								FixedFee:  utils.NewDecimal(0, 0),
							},
						},
					},
					"ABS2": {
						ID: "ABS2",
						Weights: utils.DynamicWeights{
							{
								Weight: 10,
							},
						},
						Type:  utils.MetaAbstract,
						Units: utils.NewDecimal(174900000, 0), // 0.17.. s available
						CostIncrements: []*utils.CostIncrement{
							{
								Increment: utils.NewDecimal(int64(time.Second), 0),
							},
						},
						RateProfileIDs: []string{"RP1"},
					},
					"CNCRT1": {
						ID: "CNCRT1",
						Weights: utils.DynamicWeights{
							{
								Weight: 5,
							},
						},
						Type:  utils.MetaConcrete,
						Units: utils.NewDecimal(9985, 2), // 99.85 available
						CostIncrements: []*utils.CostIncrement{
							{
								Increment: utils.NewDecimal(int64(time.Second), 0),
							},
						},
						RateProfileIDs: []string{"RP1"},
					},
				},
			},
		},
	}
	if expEvCHargers.Equals(&reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expEvCHargers), utils.ToJSON(reply))
	}

	// as we checked the EventCharges, we should check if the debit in our account was correct
	expAcc := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "ACNT1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"ABS1": {
				ID: "ABS1",
				Weights: utils.DynamicWeights{
					{
						Weight: 15,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.SumDecimal(&utils.Decimal{Big: utils.NewDecimal(0, 0).Neg(utils.NewDecimal(1, 0).Big)}, utils.NewDecimal(1, 0)), // this should be -0
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
						FixedFee:  utils.NewDecimal(0, 0),
					},
				},
			},
			"ABS2": {
				ID: "ABS2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(174900000, 0), // 0.17.. s available (should be 0 because of Increment of 1s)
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				RateProfileIDs: []string{"RP1"},
			},
			"CNCRT1": {
				ID: "CNCRT1",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(9985, 2), // 99.85 available
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				RateProfileIDs: []string{"RP1"},
			},
		},
	}
	if accRPly, err := acnts.dm.GetAccount(context.Background(), "cgrates.org", "ACNT1"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accRPly, expAcc) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expAcc), utils.ToJSON(accRPly))
	}

}

func TestV1AccountsForEventProfileIgnoreFilters(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	//inactive MetaProfileIgnoreFilters opt but correct filter
	accPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Account:1004"},
	}
	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs:   []string{"AC1"},
			utils.MetaProfileIgnoreFilters: false,
		},
	}
	rply := make([]*utils.Account, 0)
	if err := dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	} else if err := accnts.V1AccountsForEvent(context.Background(), ev, &rply); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	} else if !reflect.DeepEqual(rply[0], accPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rply[0]), utils.ToJSON(accPrf))
	}
	//active MetaProfileIgnoreFilters opt
	ev2 := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs:   []string{"AC1"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	rply2 := make([]*utils.Account, 0)
	if err := dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	} else if err := accnts.V1AccountsForEvent(context.Background(), ev2, &rply2); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	} else if !reflect.DeepEqual(rply2[0], accPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rply2[0]), utils.ToJSON(accPrf))
	}
	//for error case
	ev3 := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs:   []string{"AC1"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}
	rply3 := make([]*utils.Account, 0)
	if err := dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	} else if err := accnts.V1AccountsForEvent(context.Background(), ev3, &rply3); err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("Expected %+v, received %+v", "cannot convert field: 1s to bool", err)
	}
}

func TestV1MaxAbstractsMetaProfileIgnoreFilters(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1MaxAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight:    20,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(213, 0),
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "210ns",
			utils.OptsAccountsProfileIDs:   []string{"TestV1MaxAbstracts"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	reply := utils.EventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(accPrf.Balances, "ConcreteBalance2")

	exEvCh := utils.EventCharges{
		Abstracts: utils.NewDecimal(210, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1MaxAbstracts",
				BalanceID:    "AbstractBalance1",
				Units:        utils.NewDecimal(210, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
				RatingID:     "GENUUID_RATING",
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating: map[string]*utils.RateSInterval{
			"GENUUID_RATING": {
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "rate1",
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"rate1": {
				FixedFee:     utils.NewDecimal(0, 0),
				RecurrentFee: utils.NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TestV1MaxAbstracts": accPrf,
		},
	}
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1MaxAbstracts"].Balances["AbstractBalance1"].Units = utils.NewDecimal(int64(40*time.Second-210), 0)
		if !reply.Equals(&exEvCh) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}
}

func TestV1MaxAbstractsMetaProfileIgnoreFiltersError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1MaxAbstracts",
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "210ns",
			utils.OptsAccountsProfileIDs:   []string{"TestV1MaxAbstracts"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}
	reply := utils.EventCharges{}
	expected := "cannot convert field: 1s to bool"
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1DebitAbstractsMetaProfileIgnoreFilters(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1MaxAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight:    25,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(213, 0), // 213 - 27
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "27s",
			utils.OptsAccountsProfileIDs:   []string{"TestV1MaxAbstracts"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	reply := utils.EventCharges{}

	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.EventCharges{
		Abstracts: utils.NewDecimal(int64(27*time.Second), 0),
		Concretes: utils.NewDecimal(27, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "CHARGE1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"CHARGE1": {
				AccountID:       "TestV1MaxAbstracts",
				BalanceID:       "AbstractBalance1",
				Units:           utils.NewDecimal(int64(27*time.Second), 0),
				BalanceLimit:    utils.NewDecimal(0, 0),
				RatingID:        "RATING1",
				JoinedChargeIDs: []string{"JoinedCh1"},
			},
			"JoinedCh1": {
				AccountID:    "TestV1MaxAbstracts",
				BalanceID:    "ConcreteBalance2",
				BalanceLimit: utils.NewDecimal(0, 0),
				Units:        utils.NewDecimal(27, 0),
			},
		},
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating: map[string]*utils.RateSInterval{
			"RATING1": {
				Increments: []*utils.RateSIncrement{
					{
						RateID:         "rate1",
						CompressFactor: 1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"rate1": {
				RecurrentFee: utils.NewDecimal(1, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TestV1MaxAbstracts": accPrf,
		},
	}
	if err := accnts.V1DebitAbstracts(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1MaxAbstracts"].Balances["AbstractBalance1"].Units = utils.NewDecimal(int64(13*time.Second), 0)
		exEvCh.Accounts["TestV1MaxAbstracts"].Balances["ConcreteBalance2"].Units = utils.NewDecimal(186, 0)
		if !exEvCh.Equals(&reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}

	//now we'll check the debited account
	accPrf.Balances["AbstractBalance1"].Units = utils.NewDecimal(int64(13*time.Second), 0)
	accPrf.Balances["ConcreteBalance2"].Units = utils.NewDecimal(186, 0)
	if debitedAcc, err := accnts.dm.GetAccount(context.Background(), accPrf.Tenant, accPrf.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accPrf, debitedAcc) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(accPrf), utils.ToJSON(debitedAcc))
	}
}

func TestV1DebitAbstractsMetaProfileIgnoreFiltersError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1MaxAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight:    25,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(213, 0), // 213 - 27
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "27s",
			utils.OptsAccountsProfileIDs:   []string{"TestV1MaxAbstracts"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}
	reply := utils.EventCharges{}

	expected := "cannot convert field: 1s to bool"
	if err := accnts.V1DebitAbstracts(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1MaxConcretesProfileIgnoreFilters(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.TestFieldAcccount:testValue"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight:    15,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(30*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField:  "1004",
			"TestFieldAcccount": "testValue1",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "3m",
			utils.OptsAccountsProfileIDs:   []string{"TestV1DebitAbstracts"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	reply := utils.EventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.EventCharges{
		Concretes: utils.NewDecimal(int64(31*time.Second), 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance1",
				Units:        utils.NewDecimal(int64(time.Second), 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"GENUUID2": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.NewDecimal(int64(30*time.Second), 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"TestV1DebitAbstracts": accPrf,
		},
	}
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance1"].Units = utils.NewDecimal(0, 0)
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance2"].Units = utils.NewDecimal(0, 0)
		if !exEvCh.Equals(&reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}
}

func TestV1MaxConcretesProfileIgnoreFiltersError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstracts",
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField:  "1004",
			"TestFieldAcccount": "testValue1",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "3m",
			utils.OptsAccountsProfileIDs:   []string{"TestV1DebitAbstracts"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}
	reply := utils.EventCharges{}
	expected := "cannot convert field: 1s to bool"
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1DebitConcretesProfileIgnoreFilters(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstracts",
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: []string{"invalid_filter"},
				Weight:    0,
			},
		},
		FilterIDs: []string{"*string:~*req.TestFieldAcccount: testValue"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight:    15,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(int64(40*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance1": {
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(time.Minute), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
			"ConcreteBalance2": {
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(int64(30*time.Second), 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField:  "1004",
			"TestFieldAcccount": "testValue1",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "3m",
			utils.OptsAccountsProfileIDs:   []string{"TestV1DebitAbstracts"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	reply := utils.EventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.EventCharges{
		Concretes: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance1",
				Units:        utils.NewDecimal(60000000000, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"GENUUID2": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.NewDecimal(30000000000, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"TestV1DebitAbstracts": accPrf,
		},
	}
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance1"].Units = utils.NewDecimal(0, 0)
		exEvCh.Accounts["TestV1DebitAbstracts"].Balances["ConcreteBalance2"].Units = utils.NewDecimal(0, 0)
		if !exEvCh.Equals(&reply) {
			//if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}

	//now we will check the debited account
	rcv, err := accnts.dm.GetAccount(context.Background(), "cgrates.org", "TestV1DebitAbstracts")
	if err != nil {
		t.Error(err)
	}
	accPrf.Balances["ConcreteBalance1"].Units = &utils.Decimal{Big: decimal.New(0, 0)}
	accPrf.Balances["ConcreteBalance2"].Units = &utils.Decimal{Big: decimal.New(0, 0)}
	if !reflect.DeepEqual(rcv, accPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(accPrf), utils.ToJSON(rcv))
	}
}

func TestV1DebitConcretesProfileIgnoreFiltersError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "TestV1DebitAbstracts",
		FilterIDs: []string{"*string:~*req.TestFieldAcccount: testValue"},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField:  "1004",
			"TestFieldAcccount": "testValue1",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                "3m",
			utils.OptsAccountsProfileIDs:   []string{"TestV1DebitAbstracts"},
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}
	reply := utils.EventCharges{}
	expected := "cannot convert field: 1s to bool"
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

}
