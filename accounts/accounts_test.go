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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestRPCCall(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(nil, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	method := "ApierSv1Ping"
	expected := "UNSUPPORTED_SERVICE_METHOD"
	if err := accnts.Call(method, nil, nil); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

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
}

type dataDBMockErrorNotFound struct {
	*engine.DataDBMock
}

func (dB *dataDBMockErrorNotFound) GetAccountProfileDrv(string, string) (*utils.Account, error) {
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

	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	mockDataDB := &dataDBMockErrorNotFound{}
	//if the error is NOT_FOUND, continue to match the
	newDm := engine.NewDataManager(mockDataDB, cfg.CacheCfg(), nil)
	accnts = NewAccountS(cfg, fltr, nil, newDm)
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//mocking error in order to get from data base
	dataDB := &dataDBMockError{}
	newDm = engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	accnts = NewAccountS(cfg, fltr, nil, newDm)
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
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
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2020, 7, 21, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC),
		},
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

	cgrEvent := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1003",
		},
	}

	if _, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	cgrEvent.Event[utils.AccountField] = "1004"
	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	cgrEvent.APIOpts = make(map[string]interface{})
	cgrEvent.Time = utils.TimePointer(time.Date(2020, 8, 21, 0, 0, 0, 0, time.UTC))
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
	cgrEvent.Time = utils.TimePointer(time.Date(2020, 7, 21, 5, 0, 0, 0, time.UTC))

	accPrf.FilterIDs = []string{"invalid_filter_format"}
	expected := "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.FilterIDs = []string{"*string:~*req.Account:1003"}

	expected = "NOT_FOUND"
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
		[]string{}, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.FilterIDs = []string{"*string:~*req.Account:1004"}

	expected = "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
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
	if rcv, err := accnts.matchingAccountsForEvent("cgrates.org", cgrEvent,
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
	if _, err := accnts.accountDebit(accPrf, usage.Big, cgrEvent, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].Weights[0].FilterIDs = []string{}

	accPrf.Balances["ConcreteBalance1"].Type = "not_a_type"
	expected = "unsupported balance type: <not_a_type>"
	if _, err := accnts.accountDebit(accPrf, usage.Big, cgrEvent, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].Type = utils.MetaConcrete

	usage = &utils.Decimal{decimal.New(0, 0)}
	if _, err := accnts.accountDebit(accPrf, usage.Big, cgrEvent, true); err != nil {
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
	if _, err := accnts.accountDebit(accPrf, usage.Big, cgrEvent, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["ConcreteBalance1"].UnitFactors[0].FilterIDs = []string{}

	expectedUsage := &utils.Decimal{decimal.New(150, 0)}
	if evCh, err := accnts.accountDebit(accPrf, usage.Big, cgrEvent, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(evCh.Concretes.Big, expectedUsage.Big) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedUsage.Big), utils.ToJSON(evCh.Concretes.Big))
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
			utils.Usage:        "not_time_format",
		},
		APIOpts: map[string]interface{}{},
	}

	expected := "time: invalid duration \"not_time_format\""
	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(cgrEvent.Event, utils.Usage)

	cgrEvent.Event[utils.MetaUsage] = "not_time_format"
	expected = "time: invalid duration \"not_time_format\""
	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(cgrEvent.Event, utils.MetaUsage)

	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, true, false); err != nil {
		t.Error(err)
	}
	cgrEvent.Event[utils.MetaUsage] = "0"

	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, false); err != nil {
		t.Error(err)
	}

	cgrEvent.Event[utils.MetaUsage] = "55s"

	accntsPrf[0].Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{"invalid_filter_format"}
	expected = "NOT_FOUND:invalid_filter_format"
	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accntsPrf[0].Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	cgrEvent.Event[utils.Usage] = "300ns"
	if evCh, err := accnts.accountsDebit(accntsPrf, cgrEvent, true, true); err != nil {
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
	expected = "NO_DATA_BASE_CONNECTION"
	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, true, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	subString := "<AccountS> error <NO_DATA_BASE_CONNECTION> restoring account <cgrates.org:TestAccountsDebit>"
	if rcv := buff.String(); !strings.Contains(rcv, subString) {
		t.Errorf("Expected %+q, received %+q", subString, rcv)
	}

	log.SetOutput(os.Stderr)
}

func TestV1AccountProfilesForEvent(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "1004",
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2020, 7, 21, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC),
		},
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

	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestMatchingAccountsForEvent",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1004",
			},
		},
	}
	rply := make([]*utils.Account, 0)

	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter_format"
	if err := accnts.V1AccountProfilesForEvent(args, &rply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	accPrf.Weights[0].FilterIDs = []string{}
	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	} else if err := accnts.V1AccountProfilesForEvent(args, &rply); err != nil {
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

	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestMatchingAccountsForEvent",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1004",
				utils.Usage:        "210ns",
			},
		},
	}
	reply := utils.ExtEventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxAbstracts(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxAbstracts(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(accPrf.Balances, "ConcreteBalance2")

	exEvCh := utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(210),
	}
	if err := accnts.V1MaxAbstracts(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exEvCh, reply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
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

	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestV1DebitID",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1004",
				utils.Usage:        "27s",
			},
		},
	}
	reply := utils.ExtEventCharges{}

	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitAbstracts(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitAbstracts(args, &reply); err == nil || err.Error() != expected {
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
	if debitedAcc, err := accnts.dm.GetAccountProfile(accPrf.Tenant, accPrf.ID); err != nil {
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

	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestV1DebitID",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1004",
				utils.Usage:        "3m",
			},
		},
	}
	reply := utils.ExtEventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxConcretes(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1MaxConcretes(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.ExtEventCharges{
		Concretes: utils.Float64Pointer(float64(time.Minute + 30*time.Second)),
	}
	if err := accnts.V1MaxConcretes(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exEvCh, reply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
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

	if err := accnts.dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestV1DebitID",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1004",
				utils.Usage:        "3m",
			},
		},
	}
	reply := utils.ExtEventCharges{}
	expected := "SERVER_ERROR: NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitConcretes(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Weights[0].FilterIDs = []string{}

	expected = "NOT_FOUND:invalid_filter"
	if err := accnts.V1DebitConcretes(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights[0].FilterIDs = []string{}

	exEvCh := utils.ExtEventCharges{
		Concretes: utils.Float64Pointer(float64(time.Minute + 30*time.Second)),
	}
	if err := accnts.V1DebitConcretes(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exEvCh, reply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
	}

	//now we will check the debited account
	rcv, err := accnts.dm.GetAccountProfile("cgrates.org", "TestV1DebitAbstracts")
	if err != nil {
		t.Error(err)
	}
	accPrf.Balances["ConcreteBalance1"].Units = &utils.Decimal{decimal.New(0, 0)}
	accPrf.Balances["ConcreteBalance2"].Units = &utils.Decimal{decimal.New(0, 0)}
	if !reflect.DeepEqual(rcv, accPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(accPrf), utils.ToJSON(rcv))
	}

}

func TestMultipleAccountsFail(t *testing.T) {
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

	if err := accnts.dm.SetAccountProfile(accPrf[0], true); err != nil {
		t.Error(err)
	}
	if err := accnts.dm.SetAccountProfile(accPrf[1], true); err != nil {
		t.Error(err)
	}
	if err := accnts.dm.SetAccountProfile(accPrf[2], true); err != nil {
		t.Error(err)
	}

	expected := "NOT_FOUND:invalid_format"
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", args,
		[]string{"TestV1MaxAbstracts", "TestV1MaxAbstracts2", "TestV1MaxAbstracts3"}, true); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "NOT_FOUND:invalid_format"
	if _, err := accnts.matchingAccountsForEvent("cgrates.org", args,
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
	if err := accnts.V1ActionSetBalance(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.AccountID = "TestV1ActionSetBalance"

	expected = "MANDATORY_IE_MISSING: [Diktats]"
	if err := accnts.V1ActionSetBalance(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.Diktats = []*utils.BalDiktat{
		{
			Path:  "*balance.AbstractBalance1",
			Value: "10",
		},
	}

	expected = "WRONG_PATH"
	if err := accnts.V1ActionSetBalance(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	args.Diktats = []*utils.BalDiktat{
		{
			Path:  "*balance.AbstractBalance1.Units",
			Value: "10",
		},
	}
	args.Tenant = "cgrates.org"
	if err := accnts.V1ActionSetBalance(args, &reply); err != nil {
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
	if rcv, err := accnts.dm.GetAccountProfile(args.Tenant, args.AccountID); err != nil {
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

	if err := accnts.V1ActionSetBalance(argsSet, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected status reply", reply)
	}

	//remove it
	args := &utils.ArgsActRemoveBalances{}

	expected := "MANDATORY_IE_MISSING: [AccountID]"
	if err := accnts.V1ActionRemoveBalance(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.AccountID = "TestV1ActionRemoveBalance"

	expected = "MANDATORY_IE_MISSING: [BalanceIDs]"
	if err := accnts.V1ActionRemoveBalance(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.BalanceIDs = []string{"AbstractBalance1"}

	expected = "NO_DATA_BASE_CONNECTION"
	accnts.dm = nil
	if err := accnts.V1ActionRemoveBalance(args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	accnts.dm = engine.NewDataManager(data, cfg.CacheCfg(), nil)

	if err := accnts.V1ActionRemoveBalance(args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected status reply", reply)
	}
}
