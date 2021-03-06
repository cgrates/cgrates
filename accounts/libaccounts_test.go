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

	rates2 "github.com/cgrates/cgrates/rates"

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
)

func TestNewAccountBalanceOperators(t *testing.T) {
	acntPrf := &utils.AccountProfile{
		ID:     "TEST_ID",
		Tenant: "cgrates.org",
		Balances: map[string]*utils.Balance{
			"BL0": {
				ID:   "BALANCE1",
				Type: utils.MetaAbstract,
			},
			"BL1": {
				ID:   "BALANCE1",
				Type: utils.MetaConcrete,
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	filters := engine.NewFilterS(config.NewDefaultCGRConfig(), nil, nil)

	concrete, err := newBalanceOperator(acntPrf.ID, acntPrf.Balances["BL1"], nil, filters, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
	var cncrtBlncs []*concreteBalance
	cncrtBlncs = append(cncrtBlncs, concrete.(*concreteBalance))

	expected := &abstractBalance{
		acntID:     acntPrf.ID,
		blnCfg:     acntPrf.Balances["BL0"],
		fltrS:      filters,
		cncrtBlncs: cncrtBlncs,
	}
	blnCfgs := []*utils.Balance{acntPrf.Balances["BL0"], acntPrf.Balances["BL1"]}
	if blcOp, err := newBalanceOperators(acntPrf.ID, blnCfgs, filters, nil,
		nil, nil); err != nil {
		t.Error(err)
	} else {
		for _, bal := range blcOp {
			if rcv, canCast := bal.(*abstractBalance); canCast {
				if !reflect.DeepEqual(expected, rcv) {
					t.Errorf("Expected %+v, received %+v", expected, rcv)
				}
			}
		}
	}

	acntPrf.Balances["BL1"].Type = "INVALID_TYPE"
	expectedErr := "unsupported balance type: <INVALID_TYPE>"
	if _, err := newBalanceOperators(acntPrf.ID, blnCfgs, filters, nil,
		nil, nil); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

type testMockCall struct {
	calls map[string]func(args interface{}, reply interface{}) error
}

func (tS *testMockCall) Call(method string, args interface{}, rply interface{}) error {
	if call, has := tS.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(args, rply)
	}
}

func TestProcessAttributeS(t *testing.T) {
	engine.Cache.Clear(nil)

	config := config.NewDefaultCGRConfig()
	sTestMock := &testMockCall{ // coverage purpose
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.AttributeSv1ProcessEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(config, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): chanInternal,
	})
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID1",
		Opts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: "20",
		},
	}

	attrsConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	if _, err := processAttributeS(connMgr, cgrEvent, attrsConns, nil); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestRateSCostForEvent(t *testing.T) { // coverage purpose
	engine.Cache.Clear(nil)

	config := config.NewDefaultCGRConfig()
	sTestMock := &testMockCall{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.RateSv1CostForEvent: func(args interface{}, reply interface{}) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(config, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS): chanInternal,
	})
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID1",
	}

	rateSConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)}

	if _, err := rateSCostForEvent(connMgr, cgrEvent, rateSConns, nil); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestRateSCostForEvent2(t *testing.T) { // coverage purpose
	engine.Cache.Clear(nil)

	config := config.NewDefaultCGRConfig()
	sTestMock := &testMockCall{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.RateSv1CostForEvent: func(args interface{}, reply interface{}) error {
				rplCast, canCast := reply.(*engine.RateProfileCost)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				customRply := &engine.RateProfileCost{
					ID:   "test",
					Cost: 1,
				}
				*rplCast = *customRply
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(config, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS): chanInternal,
	})
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID1",
	}

	rateSConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)}

	if _, err := rateSCostForEvent(connMgr, cgrEvent, rateSConns, nil); err != nil {
		t.Error(err)
	}
}

func TestDebitUsageFromConcretes(t *testing.T) {
	engine.Cache.Clear(nil)

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()

	filterS := engine.NewFilterS(cfg, nil, dm)
	cb1 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB1",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	cb2 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB2",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	expectedEvCh := &utils.EventCharges{
		Concretes:   utils.NewDecimal(700, 0),
		Accounting:  make(map[string]*utils.AccountCharge),
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating:      make(map[string]*utils.RateSInterval),
	}

	if evCh, err := debitConcreteUnits(decimal.New(700, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	} else if !reflect.DeepEqual(expectedEvCh, evCh) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCh), utils.ToJSON(evCh))
	}

	cb1.blnCfg.Units = utils.NewDecimal(500, 0)
	cb2.blnCfg.Units = utils.NewDecimal(500, 0)

	if _, err := debitConcreteUnits(decimal.New(1100, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err == nil || err != utils.ErrInsufficientCredit {
		t.Errorf("Expected %+v, received %+v", utils.ErrInsufficientCredit, err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}
}

func TestDebitUsageFromConcretesFromRateS(t *testing.T) {
	engine.Cache.Clear(nil)

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	sTestMock := &testMockCall{
		calls: map[string]func(args interface{}, reply interface{}) error{
			utils.RateSv1CostForEvent: func(args interface{}, reply interface{}) error {
				rplCast, canCast := reply.(*engine.RateProfileCost)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				customRply := &engine.RateProfileCost{
					ID:   "test",
					Cost: 100,
				}
				*rplCast = *customRply
				return nil
			},
		},
	}
	chanInternal := make(chan rpcclient.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS): chanInternal,
	})
	filterS := engine.NewFilterS(cfg, nil, dm)
	cb1 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB1",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS:   filterS,
		connMgr: connMgr,
	}
	cb2 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB2",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS:   filterS,
		connMgr: connMgr,
	}

	expectedEvCh := &utils.EventCharges{
		Concretes:   utils.NewDecimal(700, 0),
		Accounting:  make(map[string]*utils.AccountCharge),
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating:      make(map[string]*utils.RateSInterval),
	}

	if evCh, err := debitConcreteUnits(decimal.New(700, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	} else if !reflect.DeepEqual(expectedEvCh, evCh) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCh), utils.ToJSON(evCh))
	}

	// debit all the units from balances
	cb1.blnCfg.Units = utils.NewDecimal(500, 0)
	cb2.blnCfg.Units = utils.NewDecimal(500, 0)

	if _, err := debitConcreteUnits(decimal.New(1000, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}
}

func TestDebitUsageFromConcretesRestore(t *testing.T) {
	engine.Cache.Clear(nil)

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()

	filterS := engine.NewFilterS(cfg, nil, dm)
	cb1 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:        "CB1",
			Type:      utils.MetaConcrete,
			FilterIDs: []string{"*string"},
			Units:     utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	cb2 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB2",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}

	if _, err := debitConcreteUnits(decimal.New(200, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2},
		new(utils.CGREvent)); err == nil || err.Error() != "inline parse error for string: <*string>" {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}
}

func TestMaxDebitUsageFromConcretes(t *testing.T) {
	engine.Cache.Clear(nil)

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AccountSCfg().MaxIterations = 100
	config.SetCgrConfig(cfg)
	filterS := engine.NewFilterS(cfg, nil, dm)
	cb1 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB1",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	cb2 := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "CB2",
			Type:  utils.MetaConcrete,
			Units: utils.NewDecimal(500, 0), // 500 Units
		},
		fltrS: filterS,
	}
	if _, err := maxDebitAbstractsFromConcretes(decimal.New(1100, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, nil, new(utils.CGREvent),
		nil, nil, nil, nil, &utils.CostIncrement{
			Increment:    utils.NewDecimal(1, 0),
			RecurrentFee: utils.NewDecimal(1, 0),
		}); err == nil || err != utils.ErrMaxIncrementsExceeded {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}
}

func TestRestoreAccount(t *testing.T) { //coverage purpose
	engine.Cache.Clear(nil)

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	acntPrf := &utils.AccountProfile{
		Tenant: "cgrates.org",
		ID:     "1001",
		Balances: map[string]*utils.Balance{
			"CB1": {
				ID:    "CB1",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(500, 0),
			},
			"CB2": {
				ID:    "CB2",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(300, 0),
			},
		},
	}
	if err := dm.SetAccountProfile(acntPrf, false); err != nil {
		t.Error(err)
	}

	restoreAccounts(dm, []*utils.AccountProfileWithWeight{
		{acntPrf, 0, utils.EmptyString},
	}, []utils.AccountBalancesBackup{
		map[string]*decimal.Big{"CB2": decimal.New(100, 0)},
	})

	if rcv, err := dm.GetAccountProfile("cgrates.org", "1001"); err != nil {
		t.Error(err)
	} else if len(rcv.Balances) != 2 {
		t.Errorf("Unexpected number of balances received")
	} else if rcv.Balances["CB2"].Units.Cmp(decimal.New(100, 0)) != 0 {
		t.Errorf("Unexpected balance received after restore")
	}
}

type dataDBMockError struct {
	*engine.DataDBMock
}

func TestRestoreAccount2(t *testing.T) { //coverage purpose
	engine.Cache.Clear(nil)

	dm := engine.NewDataManager(&dataDBMockError{}, config.CgrConfig().CacheCfg(), nil)
	acntPrf := &utils.AccountProfile{
		Tenant: "cgrates.org",
		ID:     "1001",
		Balances: map[string]*utils.Balance{
			"CB1": {
				ID:    "CB1",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(500, 0), // 500 Units
			},
			"CB2": {
				ID:    "CB2",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(100, 0), // 500 Units
			},
		},
	}
	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)
	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	restoreAccounts(dm, []*utils.AccountProfileWithWeight{
		{acntPrf, 0, utils.EmptyString},
	}, []utils.AccountBalancesBackup{
		map[string]*decimal.Big{"CB1": decimal.New(100, 0)},
	})

	subString := "<AccountS> error <NOT_IMPLEMENTED> restoring account <cgrates.org:1001>"
	if rcv := buff.String(); !strings.Contains(rcv, subString) {
		t.Errorf("Expected %+q, received %+q", subString, rcv)
	}

	log.SetOutput(os.Stderr)
}

func TestRestoreAccount3(t *testing.T) { //coverage purpose
	engine.Cache.Clear(nil)

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	acntPrf := &utils.AccountProfile{
		Tenant: "cgrates.org",
		ID:     "1001",
		Balances: map[string]*utils.Balance{
			"CB1": {
				ID:    "CB1",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(500, 0), // 500 Units
			},
			"CB2": {
				ID:    "CB2",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(100, 0), // 500 Units
			},
		},
	}
	if err := dm.SetAccountProfile(acntPrf, false); err != nil {
		t.Error(err)
	}

	restoreAccounts(dm, []*utils.AccountProfileWithWeight{
		{acntPrf, 0, utils.EmptyString},
	}, []utils.AccountBalancesBackup{
		nil,
	})
}

func TestDebitFromBothBalances(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	rates := rates2.NewRateS(cfg, fltr, dm)
	//RateS
	rpcClientConn := make(chan rpcclient.ClientConnector, 1)
	rpcClientConn <- rates
	connMngr := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS): rpcClientConn,
	})
	cfg.AccountSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)}
	//AccountS
	accnts := NewAccountS(cfg, fltr, connMngr, dm)

	accPrf := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "1002",
		FilterIDs: []string{"*string:~*req.Account:2003"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance": {
				ID:    "AbstractBalance",
				Type:  utils.MetaAbstract,
				Units: utils.NewDecimal(1500, 0),
				Weights: []*utils.DynamicWeight{
					{
						Weight: 50,
					},
				},
			},
			"ConcreteBalance1": {
				ID:   "ConcreteCalance1",
				Type: utils.MetaConcrete,
				Weights: []*utils.DynamicWeight{
					{
						Weight: 10,
					},
				},
				Units: utils.NewDecimal(20, 0),
			},
			"ConcreteBalance2": {
				ID:   "ConcreteCalance2",
				Type: utils.MetaConcrete,
				Weights: []*utils.DynamicWeight{
					{
						Weight: 20,
					},
				},
				Units: utils.NewDecimal(50, 0),
			},
		},
	}

	if err := dm.SetAccountProfile(accPrf, true); err != nil {
		t.Error(err)
	}

	minDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1ns")
	if err != nil {
		t.Error(err)
	}
	rtPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:2003"},
		Rates: map[string]*engine.Rate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				FilterIDs:       nil,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0 * time.Second,
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
						FixedFee:      utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(rtPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestV1DebitID",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "2003",
				utils.Usage:        "300",
			},
		},
	}

	var reply utils.ExtEventCharges
	exEvCh := utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(300),
	}
	if err := accnts.V1DebitAbstracts(args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exEvCh, reply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
	}

	accPrf.Balances["AbstractBalance"].Units = utils.NewDecimal(1200, 0)
	accPrf.Balances["ConcreteBalance2"].Units = utils.NewDecimal(49999999997, 9)
	//as we debited, the account is changed
	if rcvAcc, err := dm.GetAccountProfile(accPrf.Tenant, accPrf.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvAcc, accPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(accPrf), utils.ToJSON(rcvAcc))
	}
}
