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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestListenAndServe(t *testing.T) {
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
	accnts.ListenAndServe(stopChan, cfgRld)

	if err := accnts.Shutdown(); err != nil {
		t.Error(err)
	}
}

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

type dataDBMockErrorNotFound struct {
	*engine.DataDBMock
}

func (dB *dataDBMockErrorNotFound) GetAccountProfileDrv(string, string) (*utils.AccountProfile, error) {
	return nil, utils.ErrNotFound
}

func TestMatchingAccountsForEventMockingErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fltr := engine.NewFilterS(cfg, nil, nil)

	accPrf := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "1004",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": &utils.Balance{
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
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
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.AccountProfile{
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
			"AbstractBalance1": &utils.Balance{
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

	cgrEvent.Opts = make(map[string]interface{})
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

	expectedAccPrfWeght := utils.AccountProfilesWithWeight{
		{
			AccountProfile: accPrf,
			Weight:         20,
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
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accPrf := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "TestAccountDebit",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": &utils.Balance{
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"invalid_filter_format"},
						Weight:    20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(200, 0)},
			},
			"AbstractBalance1": &utils.Balance{
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{decimal.New(20, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
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
	usage := &utils.Decimal{decimal.New(210, 0)}

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
	usage = &utils.Decimal{decimal.New(210, 0)}

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

	expectedUsage := &utils.Decimal{decimal.New(200, 0)}
	if evCh, err := accnts.accountDebit(accPrf, usage.Big, cgrEvent, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(evCh.Usage.Big, expectedUsage.Big) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedUsage), utils.ToJSON(evCh.Usage))
	}
}

func TestAccountsDebit(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	accntsPrf := []*utils.AccountProfileWithWeight{
		{
			AccountProfile: &utils.AccountProfile{
				Tenant:    "cgrates.org",
				ID:        "TestAccountsDebit",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Balances: map[string]*utils.Balance{
					"AbstractBalance1": &utils.Balance{
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 5,
							},
						},
						Type:  utils.MetaAbstract,
						Units: &utils.Decimal{decimal.New(int64(40*time.Second), 0)},
						CostIncrements: []*utils.CostIncrement{
							&utils.CostIncrement{
								Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
								FixedFee:     &utils.Decimal{decimal.New(0, 0)},
								RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
							},
						},
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
		Opts: map[string]interface{}{},
	}

	expected := "time: invalid duration \"not_time_format\""
	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, false); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	delete(cgrEvent.Event, utils.Usage)

	cgrEvent.Opts[utils.MetaUsage] = "not_time_format"
	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, false); err != nil {
		t.Error(err)
	}
	cgrEvent.Opts[utils.MetaUsage] = "55s"

	if _, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, false); err != nil {
		t.Error(err)
	}

	cgrEvent.Event[utils.Usage] = "55s"
	expectedUsage := &utils.Decimal{decimal.New(0, 0)}
	if evCh, err := accnts.accountsDebit(accntsPrf, cgrEvent, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedUsage, evCh.Usage) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedUsage), utils.ToJSON(evCh.Usage))
	}
}
