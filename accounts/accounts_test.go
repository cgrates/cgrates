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
	"bytes"
	"log"
	"os"
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

func TestShutDownCoverage(t *testing.T) {
	//this is called in order to cover the ListenAndServe method
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	stopChan := make(chan struct{}, 1)
	cfgRld := make(chan struct{}, 1)
	cfgRld <- struct{}{}
	go func() {
		time.Sleep(10)
		stopChan <- struct{}{}
	}()

	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)
	buff := new(bytes.Buffer)
	log.SetOutput(buff)
	accnts.ListenAndServe(stopChan, cfgRld)

	//this is called in order to cover the ShutDown method
	accnts.Shutdown()
	expected := "CGRateS <> [INFO] <CoreS> shutdown <AccountS>"
	if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}

	log.SetOutput(os.Stderr)
	utils.Logger.SetLogLevel(6)
}

type dataDBMockErrorNotFound struct {
	engine.DataDBMock
}

func (dB *dataDBMockErrorNotFound) GetAccountDrv(*context.Context, string, string) (*utils.Account, error) {
	return nil, utils.ErrNotFound
}

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
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
		},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
	}

	data := engine.NewInternalDB(nil, nil, true)
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
			[]string{}, true); err == nil || err != utils.ErrNotFound {
			t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
		}

	*/

	//mocking error in order to get from data base
	dataDB := &dataDBMockError{}
	newDm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	accnts = NewAccountS(cfg, fltr, nil, newDm)
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestMatchingAccountsForEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Units: &utils.Decimal{decimal.New(0, 0)},
			},
		},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1003",
		},
	}

	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	cgrEvent.Event[utils.AccountField] = "1004"
	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	cgrEvent.APIOpts = make(map[string]interface{})
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	accPrf.FilterIDs = []string{"invalid_filter_format"}
	expected := "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.FilterIDs = []string{"*string:~*req.Account:1003"}

	expected = "NOT_FOUND"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.FilterIDs = []string{"*string:~*req.Account:1004"}

	expected = "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err.Error() != expected {
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
		[]string{}, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccPrfWeght, rcv) {
		t.Errorf("Expected %+v, received %+v", expectedAccPrfWeght, utils.ToJSON(rcv))
	}
}

func TestAccountDebit(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Units: &utils.Decimal{decimal.New(150, 0)},
			},
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{decimal.New(200, 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TEST_EVENT",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
	}

	usage := &utils.Decimal{decimal.New(190, 0)}
	expected := "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].Weights[0].FilterIDs = []string{}

	accPrf.Balances["ConcreteBalance1"].Type = "not_a_type"
	expected = "unsupported balance type: <not_a_type>"
	if _, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].Type = utils.MetaConcrete

	usage = &utils.Decimal{decimal.New(0, 0)}
	if _, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err != nil {
		t.Error(err)
	}
	usage = &utils.Decimal{decimal.New(190, 0)}

	accPrf.Balances["ConcreteBalance1"].UnitFactors = []*utils.UnitFactor{
		{
			FilterIDs: []string{"invalid_format_type"},
			Factor:    &utils.Decimal{decimal.New(1, 0)},
		},
	}
	expected = "NOT_FOUND:invalid_format_type"
	if _, err := accnts.accountDebit(context.Background(), accPrf, usage.Big,
		cgrEvent, true, decimal.New(0, 0)); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].UnitFactors[0].FilterIDs = []string{}

	expectedUsage := &utils.Decimal{decimal.New(150, 0)}
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
	data := engine.NewInternalDB(nil, nil, true)
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
						Units: &utils.Decimal{decimal.New(90, 0)},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    &utils.Decimal{decimal.New(1, 0)},
								FixedFee:     &utils.Decimal{decimal.New(2, 1)},
								RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
							},
						},
					},
				},
			},
		},
	}

	evChExp := &utils.EventCharges{
		Abstracts: utils.NewDecimal(89, 0),
		Concretes: utils.NewDecimal(1484, 1),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "CHARGING1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"CHARGING1": {
				AccountID:       "TestAccountsDebitGetUsage",
				BalanceID:       "*transabstract",
				Units:           utils.NewDecimal(89, 0),
				RatingID:        "RATING1",
				JoinedChargeIDs: []string{"JND_CHRG1", "JND_CHRG2"},
			},
			"JND_CHRG1": {
				AccountID:    "TestAccountsDebitGetUsage",
				BalanceID:    "ConcreteBal1",
				BalanceLimit: utils.NewDecimal(0, 0),
				Units:        utils.NewDecimal(592, 1),
			},
			"JND_CHRG2": {
				AccountID:    "TestAccountsDebitGetUsage",
				BalanceID:    "ConcreteBal1",
				BalanceLimit: utils.NewDecimal(0, 0),
				Units:        utils.NewDecimal(892, 1),
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
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"TestAccountsDebitGetUsage": accntsPrf[0].Account,
		},
	}

	// get usage from *ratesUsage
	cgrEvent := &utils.CGREvent{
		ID:     "TEST_EVENT_get_usage",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Destination: "+445643",
		},
		APIOpts: map[string]interface{}{
			utils.OptsRatesUsage: "2s",
		},
	}
	if rcv, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, false); err != nil {
		t.Error(err)
	} else if !rcv.Equals(evChExp) {
		t.Errorf("Expected %v, \n received %v", utils.ToJSON(evChExp), utils.ToJSON(rcv))
	}

	// get usage from *usage
	//firstly reset the account
	accntsPrf[0].Account.Balances["ConcreteBal1"].Units = &utils.Decimal{decimal.New(90, 0)}
	accnts = NewAccountS(cfg, fltr, nil, dm)
	cgrEvent = &utils.CGREvent{
		ID:     "TEST_EVENT_get_usage",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Destination: "+445643",
		},
		APIOpts: map[string]interface{}{
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
	data := engine.NewInternalDB(nil, nil, true)
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
						Units: &utils.Decimal{decimal.New(40, 0)},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    &utils.Decimal{decimal.New(1, 0)},
								FixedFee:     &utils.Decimal{decimal.New(0, 0)},
								RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
						Units: &utils.Decimal{decimal.New(213, 0)},
					},
				},
			},
		},
	}

	cgrEvent := &utils.CGREvent{
		ID:     "TEST_EVENT",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]interface{}{
			utils.OptsRatesUsage: "not_time_format",
		},
	}

	expected := "can't convert <not_time_format> to decimal"
	if _, err := accnts.accountsDebit(context.Background(), accntsPrf,
		cgrEvent, false, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(cgrEvent.APIOpts, utils.OptsRatesUsage)

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

	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)
	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	accntsPrf[0].Balances["ConcreteBalance2"].Units = &utils.Decimal{decimal.New(213, 0)}
	accnts.dm = nil
	expected = utils.ErrNoDatabaseConn.Error()
	if _, err := accnts.accountsDebit(context.Background(), accntsPrf, cgrEvent, true, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	subString := "<AccountS> error <NO_DATABASE_CONNECTION> restoring account <cgrates.org:TestAccountsDebit>"
	if rcv := buff.String(); !strings.Contains(rcv, subString) {
		t.Errorf("Expected %+q, received %+q", subString, rcv)
	}

	log.SetOutput(os.Stderr)
	utils.Logger.SetLogLevel(6)
}

func TestV1AccountsForEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Units: &utils.Decimal{decimal.New(0, 0)},
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
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
	data := engine.NewInternalDB(nil, nil, true)
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
				Units: &utils.Decimal{decimal.New(int64(40*time.Second), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(0, 0)},
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
				Units: &utils.Decimal{decimal.New(213, 0)},
			},
		},
	}

	if err := accnts.dm.SetAccount(context.Background(), accPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage: "210ns",
		},
	}
	reply := utils.ExtEventCharges{}
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

	extAccPrf, err := accPrf.AsExtAccount()
	if err != nil {
		t.Error(err)
	}
	extAccPrf.Balances["AbstractBalance1"].Units = utils.Float64Pointer(float64(40*time.Second - 210*time.Nanosecond))

	exEvCh := utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(210),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1MaxAbstracts",
				BalanceID:    "AbstractBalance1",
				BalanceLimit: utils.Float64Pointer(0),
				RatingID:     "GENUUID_RATING",
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating: map[string]*utils.ExtRateSInterval{
			"GENUUID_RATING": {
				Increments: []*utils.ExtRateSIncrement{
					{
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: make(map[string]*utils.ExtIntervalRate),
		Accounts: map[string]*utils.ExtAccount{
			"TestV1MaxAbstracts": extAccPrf,
		},
	}
	if err := accnts.V1MaxAbstracts(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Charges = reply.Charges
		exEvCh.Rating = reply.Rating
		exEvCh.Accounting = reply.Accounting
		if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}
}

func TestV1DebitAbstracts(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
						Weight:    25,
						FilterIDs: []string{"invalid_filter"},
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(int64(40*time.Second), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
		Event: map[string]interface{}{
			utils.AccountField: "1004",
			utils.Usage:        "27s",
		},
	}
	reply := utils.ExtEventCharges{}

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
	/*
		exEvCh := utils.ExtEventCharges{
			Abstracts: utils.Float64Pointer(float64(27 * time.Second)),
		}
		if err := accnts.V1DebitAbstracts(args, &reply); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	*/

	//now we'll check the debited account
	accPrf.Balances["AbstractBalance1"].Units = &utils.Decimal{decimal.New(39999999973, 0)}
	if debitedAcc, err := accnts.dm.GetAccount(context.Background(), accPrf.Tenant, accPrf.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accPrf, debitedAcc) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(accPrf), utils.ToJSON(debitedAcc))
	}
}

func TestV1MaxConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Units: &utils.Decimal{decimal.New(int64(40*time.Second), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
				Units: &utils.Decimal{decimal.New(int64(time.Minute), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
				Units: &utils.Decimal{decimal.New(int64(30*time.Second), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
		Event: map[string]interface{}{
			utils.AccountField: "1004",
			utils.Usage:        "3m",
		},
	}
	reply := utils.ExtEventCharges{}
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

	extAccPrf, err := accPrf.AsExtAccount()
	if err != nil {
		t.Error(err)
	}
	extAccPrf.Balances["ConcreteBalance1"].Units = utils.Float64Pointer(0)
	extAccPrf.Balances["ConcreteBalance2"].Units = utils.Float64Pointer(0)

	exEvCh := utils.ExtEventCharges{
		Concretes: utils.Float64Pointer(float64(time.Minute + 30*time.Second)),
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
		Accounting: map[string]*utils.ExtAccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance1",
				Units:        utils.Float64Pointer(float64(time.Minute)),
				BalanceLimit: utils.Float64Pointer(0),
			},
			"GENUUID2": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.Float64Pointer(float64(30 * time.Second)),
				BalanceLimit: utils.Float64Pointer(0),
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"TestV1DebitAbstracts": extAccPrf,
		},
	}
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Charges = reply.Charges
		exEvCh.Accounting = reply.Accounting
		if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}
}

func TestV1DebitConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Units: &utils.Decimal{decimal.New(int64(40*time.Second), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
				Units: &utils.Decimal{decimal.New(int64(time.Minute), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
				Units: &utils.Decimal{decimal.New(int64(30*time.Second), 0)},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
		Event: map[string]interface{}{
			utils.AccountField: "1004",
			utils.Usage:        "3m",
		},
	}
	reply := utils.ExtEventCharges{}
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

	extAccPrf, err := accPrf.AsExtAccount()
	if err != nil {
		t.Error(err)
	}
	extAccPrf.Balances["ConcreteBalance1"].Units = utils.Float64Pointer(0)
	extAccPrf.Balances["ConcreteBalance2"].Units = utils.Float64Pointer(0)
	exEvCh := utils.ExtEventCharges{
		Concretes: utils.Float64Pointer(float64(time.Minute + 30*time.Second)),
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
		Accounting: map[string]*utils.ExtAccountCharge{
			"GENUUID1": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance1",
				Units:        utils.Float64Pointer(60000000000),
				BalanceLimit: utils.Float64Pointer(0),
			},
			"GENUUID2": {
				AccountID:    "TestV1DebitAbstracts",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.Float64Pointer(30000000000),
				BalanceLimit: utils.Float64Pointer(0),
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"TestV1DebitAbstracts": extAccPrf,
		},
	}
	if err := accnts.V1DebitConcretes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Accounting = reply.Accounting
		exEvCh.Charges = reply.Charges
		if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}

	//now we will check the debited account
	rcv, err := accnts.dm.GetAccount(context.Background(), "cgrates.org", "TestV1DebitAbstracts")
	if err != nil {
		t.Error(err)
	}
	accPrf.Balances["ConcreteBalance1"].Units = &utils.Decimal{decimal.New(0, 0)}
	accPrf.Balances["ConcreteBalance2"].Units = &utils.Decimal{decimal.New(0, 0)}
	if !reflect.DeepEqual(rcv, accPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(accPrf), utils.ToJSON(rcv))
	}

}

func TestMultipleAccountsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
					Units: &utils.Decimal{decimal.New(213, 0)},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    &utils.Decimal{decimal.New(1, 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
					Units: &utils.Decimal{decimal.New(213, 0)},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    &utils.Decimal{decimal.New(1, 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
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
					Units: &utils.Decimal{decimal.New(213, 0)},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    &utils.Decimal{decimal.New(1, 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
						},
					},
				},
			},
		},
	}

	args := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
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
		[]string{"TestV1MaxAbstracts", "TestV1MaxAbstracts2", "TestV1MaxAbstracts3"}, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "NOT_FOUND:invalid_format"
	if _, err := accnts.matchingAccountsForEvent(context.Background(), "cgrates.org", args,
		[]string{"TestV1MaxAbstracts", "TestV1MaxAbstracts2", "TestV1MaxAbstracts3"}, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestV1ActionSetBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
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
				Units: &utils.Decimal{decimal.New(10, 0)},
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
	data := engine.NewInternalDB(nil, nil, true)
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
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)

	// set the internal AttributeS within connMngr
	attrSConn := make(chan birpc.ClientConnector, 1)
	attrSrv, _ := birpc.NewServiceWithMethodsRename(engine.NewAttributeService(dm, fltrS, cfg), utils.AttributeSv1, true, func(key string) (newKey string) { return strings.TrimPrefix(key, utils.V1Prfx) }) // update the name of the functions
	attrSConn <- attrSrv
	cfg.AccountSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	// Set the internal rateS within connMngr
	rateSConn := make(chan birpc.ClientConnector, 1)
	rateSrv, _ := birpc.NewServiceWithMethodsRename(rates.NewRateS(cfg, fltrS, dm), utils.RateSv1, true, func(key string) (newKey string) { return strings.TrimPrefix(key, utils.V1Prfx) }) // update the name of the functions
	rateSConn <- rateSrv

	cfg.AccountSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)}

	connMngr := engine.NewConnManager(cfg)
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, attrSConn)
	connMngr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS), utils.RateSv1, rateSConn)

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
		Blocker: false,
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
		if err := dm.SetRateProfile(context.Background(), rtPfl, true); err != nil {
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
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:     utils.NewDecimal(4, 1),  // 0.4
						RecurrentFee: utils.NewDecimal(2, 1)}, // 0.2 per minute
				},
				Units: utils.NewDecimal(int64(130*time.Second), 0), // 2 Minute 10s, rest 10s
			},
			cb1ID: { // paying the AB1 plus own debit of 0.1 per second, limit of -200 cents
				ID:   cb1ID,
				Type: utils.MetaConcrete,
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Opts: map[string]interface{}{
					utils.MetaBalanceLimit: -200.0,
				},
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee: utils.NewDecimal(1, 1)}, // 0.1 per second
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
			cb2ID: &utils.Balance{ //125s with rating from RateS
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
			ab1ID: &utils.Balance{ // cost: 0.4 connectFee plus 0.2 per minute, available 2 minutes, should remain  10 units
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
			cb1ID: &utils.Balance{ // absorb all costs, standard rating used when primary debiting
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
			cb2ID: &utils.Balance{ // absorb all costs, standard rating used when primary debiting
				ID:   cb2ID,
				Type: utils.MetaConcrete,
				Opts: map[string]interface{}{
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

			// 7m26 ABSTR, 4.95 CONCR with remaining flat covered by CB2 with RateS, RP_2, -0.05 on CB2

		},
	}
	if err := dm.SetAccount(context.Background(), acnt2, true); err != nil {
		t.Error(err)
	}

	eEvChgs := &utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(446000000000),
		Concretes: utils.Float64Pointer(4.95),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID3",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID4",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID5",
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID6",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			"GENUUID_GHOST1": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    cb1ID,
				Units:        utils.Float64Pointer(0.8),
				BalanceLimit: utils.Float64Pointer(-200),
				UnitFactorID: "GENUUID_FACTOR1",
			},
			"GENUUID3": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    ab2ID,
				BalanceLimit: utils.Float64Pointer(0),
				RatingID:     "GENUUID_RATING1",
			},
			"GENUUID2": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    cb1ID,
				Units:        utils.Float64Pointer(2),
				BalanceLimit: utils.Float64Pointer(-200),
				UnitFactorID: "GENUUID_FACTOR2",
			},
			"GENUUID5": {
				AccountID:       "TestV1DebitAbstractsEventCharges2",
				BalanceID:       ab1ID,
				BalanceLimit:    utils.Float64Pointer(0),
				RatingID:        "GENUUID_RATING2",
				JoinedChargeIDs: []string{"GENUUID_GHOST2"},
			},
			"GENUUID6": {
				AccountID: "TestV1DebitAbstractsEventCharges2",
				BalanceID: cb1ID,
				Units:     utils.Float64Pointer(0.3),
			},
			"GENUUID4": {
				AccountID:    "TestV1DebitAbstractsEventCharges1",
				BalanceID:    cb2ID,
				Units:        utils.Float64Pointer(1.25),
				BalanceLimit: utils.Float64Pointer(0),
			},
			"GENUUID_GHOST2": {
				AccountID: "TestV1DebitAbstractsEventCharges2",
				BalanceID: cb1ID,
				Units:     utils.Float64Pointer(0.6),
			},
			"GENUUID1": {
				AccountID:       "TestV1DebitAbstractsEventCharges1",
				BalanceID:       ab1ID,
				BalanceLimit:    utils.Float64Pointer(0),
				RatingID:        "GENUUID_RATING3",
				JoinedChargeIDs: []string{"GENUUID_GHOST1"},
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{
			"GENUUID_FACTOR1": {
				Factor: utils.Float64Pointer(100),
			},
			"GENUUID_FACTOR2": {
				Factor: utils.Float64Pointer(100),
			},
		},
		Rating: map[string]*utils.ExtRateSInterval{
			"GENUUID_RATING1": {
				Increments: []*utils.ExtRateSIncrement{
					{
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			"GENUUID_RATING2": {
				Increments: []*utils.ExtRateSIncrement{
					{
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
			"GENUUID_RATING3": {
				Increments: []*utils.ExtRateSIncrement{
					{
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
		Rates:    map[string]*utils.ExtIntervalRate{},
		Accounts: make(map[string]*utils.ExtAccount),
	}
	args := &utils.CGREvent{
		ID:     "TestV1DebitAbstractsEventCharges",
		Tenant: utils.CGRateSorg,
		APIOpts: map[string]interface{}{
			utils.MetaUsage: "7m26s",
		},
	}
	var rcvEC utils.ExtEventCharges
	if err := accnts.V1DebitAbstracts(context.Background(), args, &rcvEC); err != nil {
		t.Error(err)
		//} else if eEvChgs.Equals(&rcvEC) {
	} else if !reflect.DeepEqual(eEvChgs, rcvEC) {
		t.Errorf("expecting: %s, \nreceived: %s\n", utils.ToJSON(eEvChgs), utils.ToJSON(rcvEC))
	}

	/*
		acnt1.Balances[ab1ID].Units = utils.NewDecimal(int64(10*time.Second), 0)
		acnt1.Balances[cb1ID].Units = utils.NewDecimal(-200, 0)
		acnt1.Balances[ab2ID].Units = &utils.Decimal{new(decimal.Big).CopySign(decimal.New(0, 0), decimal.New(-1, 0))} // negative 0
		acnt1.Balances[cb2ID].Units = utils.NewDecimal(0, 0)
		if rcv, err := dm.GetAccount(acnt1.Tenant, acnt1.ID); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rcv, acnt1) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acnt1), utils.ToJSON(rcv))
		}


		acnt2.Balances[ab1ID].Units = utils.NewDecimal(int64(10*time.Second), 0)
		acnt2.Balances[cb1ID].Units = utils.NewDecimal(-1, 1)
		if rcv, err := dm.GetAccount(acnt2.Tenant, acnt2.ID); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rcv, acnt2) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acnt2), utils.ToJSON(rcv))
		}


		extAcnt1, err := acnt1.AsExtAccount()
		if err != nil {
			t.Error(err)
		}
		extAcnt2, err := acnt2.AsExtAccount()
		if err != nil {
			t.Error(err)
		}

		//as the names of accounting, charges, UF are GENUUIDs generator, we will change their names for comparing
		eEvChgs.Accounts = map[string]*utils.ExtAccount{
			"TestV1DebitAbstractsEventCharges1": extAcnt1,
			"TestV1DebitAbstractsEventCharges2": extAcnt2,
		}
		eEvChgs.Charges = rply.Charges
		eEvChgs.Accounting = rply.Accounting
		eEvChgs.UnitFactors = rply.UnitFactors
		eEvChgs.Accounts = rply.Accounts
		eEvChgs.Rating = rply.Rating
		if !reflect.DeepEqual(eEvChgs, rply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(eEvChgs), utils.ToJSON(rply))
		}
	*/
}

/*
func TestV1DebitAbstractsWithRecurrentFeeNegative(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrS := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltrS, nil, dm)

	acnt := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "TestV1DebitAbstractsWithRecurrentFeeNegative",
		Balances: map[string]*utils.Balance{
			"ab1": &utils.Balance{
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
			"cb1": &utils.Balance{
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
	if err := dm.SetAccount(acnt, true); err != nil {
		t.Error(err)
	}
	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestV1DebitAbstractsWithRecurrentFeeNegative",
			Tenant: "cgrates.org",
			APIOpts: map[string]interface{}{
				utils.MetaUsage: "72h",
			},
		},
	}
	expEvCh := &utils.ExtEventCharges{
		Abstracts:   utils.Float64Pointer(259200000000000),
		Concretes:   utils.Float64Pointer(-259198),
		Accounting:  map[string]*utils.ExtAccountCharge{},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{}}
	ev := &utils.ExtEventCharges{}
	if err := accnts.V1DebitAbstracts(args, ev); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ev, expEvCh) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expEvCh), utils.ToJSON(ev))
	}

	acnt.Balances["ab1"].Units = utils.NewDecimal(int64(39*time.Second), 0)
	acnt.Balances["cb1"].Units = utils.NewDecimal(259199, 0)
	if rcv, err := dm.GetAccount(acnt.Tenant, acnt.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, acnt) {
		t.Errorf("Expected %+v,received %+v", utils.ToJSON(acnt), utils.ToJSON(rcv))
	}
}

*/
