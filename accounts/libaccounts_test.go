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

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
)

func TestNewAccountBalanceOperators(t *testing.T) {
	acntPrf := &utils.Account{
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

	concrete, err := newBalanceOperator(context.Background(), acntPrf.ID, acntPrf.Balances["BL1"], nil, filters, nil, nil, nil)
	if err != nil {
		t.Error(err)
	}
	var cncrtBlncs []*concreteBalance
	cncrtBlncs = append(cncrtBlncs, concrete.(*concreteBalance))

	expected := &abstractBalance{
		acntID:     acntPrf.ID,
		blnCfg:     acntPrf.Balances["BL0"],
		fltrS:      filters,
		ctx:        context.Background(),
		cncrtBlncs: cncrtBlncs,
	}
	blnCfgs := []*utils.Balance{acntPrf.Balances["BL0"], acntPrf.Balances["BL1"]}
	if blcOp, err := newBalanceOperators(context.Background(), acntPrf.ID, blnCfgs, filters, nil,
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
}

type testMockCall struct {
	calls map[string]func(_ *context.Context, _, _ any) error
}

func (tS *testMockCall) Call(ctx *context.Context, method string, args, rply any) error {
	if call, has := tS.calls[method]; !has {
		return rpcclient.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, rply)
	}
}

func TestProcessAttributeS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	engine.Cache.Clear(nil)

	sTestMock := &testMockCall{ // coverage purpose
		calls: map[string]func(_ *context.Context, _, _ any) error{
			utils.AttributeSv1ProcessEvent: func(_ *context.Context, _, _ any) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, chanInternal)
	cgrEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ID1",
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: "20",
		},
	}

	attrsConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}

	if _, err := processAttributeS(context.Background(), connMgr, cgrEvent, attrsConns, nil); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestRateSCostForEvent(t *testing.T) { // coverage purpose
	cfg := config.NewDefaultCGRConfig()
	engine.Cache.Clear(nil)

	sTestMock := &testMockCall{
		calls: map[string]func(_ *context.Context, _, _ any) error{
			utils.RateSv1CostForEvent: func(_ *context.Context, _, _ any) error {
				return utils.ErrNotImplemented
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, chanInternal)
	cgrEvent := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "TEST_ID1",
		APIOpts: make(map[string]any),
	}

	rateSConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}

	if _, err := rateSCostForEvent(context.Background(), connMgr, cgrEvent, rateSConns, nil); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotImplemented, err)
	}
}

func TestRateSCostForEvent2(t *testing.T) { // coverage purpose
	cfg := config.NewDefaultCGRConfig()
	engine.Cache.Clear(nil)

	sTestMock := &testMockCall{
		calls: map[string]func(_ *context.Context, _, _ any) error{
			utils.RateSv1CostForEvent: func(_ *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.RateProfileCost)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				customRply := &utils.RateProfileCost{
					ID:   "test",
					Cost: utils.NewDecimal(1, 0),
				}
				*rplCast = *customRply
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, chanInternal)
	cgrEvent := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "TEST_ID1",
		APIOpts: make(map[string]any),
	}

	rateSConns := []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates)}

	if _, err := rateSCostForEvent(context.Background(), connMgr, cgrEvent, rateSConns, nil); err != nil {
		t.Error(err)
	}
}

func TestDebitUsageFromConcretes(t *testing.T) {
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
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
		Concretes: utils.NewDecimal(700, 0),
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
			"GENUUID2": {
				BalanceID:    "CB2",
				Units:        utils.NewDecimal(200, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"GENUUID1": {
				BalanceID:    "CB2",
				Units:        utils.NewDecimal(500, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating:      make(map[string]*utils.RateSInterval),
		Rates:       make(map[string]*utils.IntervalRate),
		Accounts:    make(map[string]*utils.Account),
	}

	if evCh, err := debitConcreteUnits(context.Background(), decimal.New(700, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	} else {
		//as the names of accounting, charges, UF are GENUUIDs generator, we will change their names for comparing
		expectedEvCh.Charges = evCh.Charges
		expectedEvCh.Accounting = evCh.Accounting
		if !reflect.DeepEqual(evCh, expectedEvCh) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCh), utils.ToJSON(evCh))
		}
	}

	cb1.blnCfg.Units = utils.NewDecimal(500, 0)
	cb2.blnCfg.Units = utils.NewDecimal(500, 0)

	if _, err := debitConcreteUnits(context.Background(), decimal.New(1100, 0), utils.EmptyString,
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

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	sTestMock := &testMockCall{
		calls: map[string]func(_ *context.Context, _, _ any) error{
			utils.RateSv1CostForEvent: func(_ *context.Context, args, reply any) error {
				rplCast, canCast := reply.(*utils.RateProfileCost)
				if !canCast {
					t.Errorf("Wrong argument type : %T", reply)
					return nil
				}
				customRply := &utils.RateProfileCost{
					ID:   "test",
					Cost: utils.NewDecimal(100, 0),
				}
				*rplCast = *customRply
				return nil
			},
		},
	}
	chanInternal := make(chan birpc.ClientConnector, 1)
	chanInternal <- sTestMock
	connMgr := engine.NewConnManager(cfg)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, chanInternal)
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
		Concretes: utils.NewDecimal(700, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "GENUUID1", // will be changed
				CompressFactor: 1,
			},
			{
				ChargingID:     "GENUUID2", // will be changed
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"GENUUID2": {
				BalanceID:    "CB2",
				Units:        utils.NewDecimal(200, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
			"GENUUID1": {
				BalanceID:    "CB1",
				Units:        utils.NewDecimal(500, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
			},
		},
		UnitFactors: make(map[string]*utils.UnitFactor),
		Rating:      make(map[string]*utils.RateSInterval),
		Rates:       make(map[string]*utils.IntervalRate),
		Accounts:    make(map[string]*utils.Account),
	}

	if evCh, err := debitConcreteUnits(context.Background(), decimal.New(700, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(300, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	} else {
		//as the names of accounting, charges, UF are GENUUIDs generator, we will change their names for comparing
		expectedEvCh.Charges = evCh.Charges
		expectedEvCh.Accounting = evCh.Accounting
		if !reflect.DeepEqual(evCh, expectedEvCh) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedEvCh), utils.ToJSON(evCh))
		}
	}

	// debit all the units from balances
	cb1.blnCfg.Units = utils.NewDecimal(500, 0)
	cb2.blnCfg.Units = utils.NewDecimal(500, 0)

	if _, err := debitConcreteUnits(context.Background(), decimal.New(1000, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err != nil {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}

	// debit more than we have in units
	cb1.blnCfg.Units = utils.NewDecimal(500, 0)
	cb2.blnCfg.Units = utils.NewDecimal(500, 0)

	if _, err := debitConcreteUnits(context.Background(), decimal.New(1100, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, new(utils.CGREvent)); err == nil || err != utils.ErrInsufficientCredit {
		t.Errorf("Expected %+v, received %+v", utils.ErrInsufficientCredit, err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}
}

func TestDebitUsageFromConcretesRestore(t *testing.T) {
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)

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

	if _, err := debitConcreteUnits(context.Background(), decimal.New(200, 0), utils.EmptyString,
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

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
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

	if _, err := maxDebitAbstractsFromConcretes(context.Background(), decimal.New(900, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, nil, new(utils.CGREvent),
		nil, nil, nil, nil, &utils.CostIncrement{
			Increment:    utils.NewDecimal(1, 0),
			RecurrentFee: utils.NewDecimal(1, 0),
		}, decimal.New(0, 0)); err != nil {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(100, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}

	//debit more than we have in balances with the restored units
	cb1.blnCfg.Units = utils.NewDecimal(500, 0)
	cb2.blnCfg.Units = utils.NewDecimal(500, 0)
	if _, err := maxDebitAbstractsFromConcretes(context.Background(), decimal.New(1100, 0), utils.EmptyString,
		[]*concreteBalance{cb1, cb2}, nil, &utils.CGREvent{
			ID: "Unique_id",
		},
		nil, nil, nil, nil, &utils.CostIncrement{
			Increment:    utils.NewDecimal(1, 0),
			RecurrentFee: utils.NewDecimal(1, 0),
		}, decimal.New(0, 0)); err == nil || err != utils.ErrMaxIncrementsExceeded {
		t.Error(err)
	} else if cb1.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb1.blnCfg.Units)
	} else if cb2.blnCfg.Units.Cmp(decimal.New(500, 0)) != 0 {
		t.Errorf("balance remaining: %s", cb2.blnCfg.Units)
	}
}

func TestRestoreAccount(t *testing.T) { //coverage purpose
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	acntPrf := &utils.Account{
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
	if err := dm.SetAccount(context.Background(), acntPrf, false); err != nil {
		t.Error(err)
	}

	restoreAccounts(context.Background(), dm, []*utils.AccountWithLock{
		{Account: acntPrf, LockID: utils.EmptyString},
	}, []utils.AccountBalancesBackup{
		map[string]*decimal.Big{"CB2": decimal.New(100, 0)},
	})

	if rcv, err := dm.GetAccount(context.Background(), "cgrates.org", "1001"); err != nil {
		t.Error(err)
	} else if len(rcv.Balances) != 2 {
		t.Errorf("Unexpected number of balances received")
	} else if rcv.Balances["CB2"].Units.Cmp(decimal.New(100, 0)) != 0 {
		t.Errorf("Unexpected balance received after restore")
	}
}

type dataDBMockError struct {
	engine.DataDBMock
}

func TestRestoreAccount2(t *testing.T) { //coverage purpose
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	dm := engine.NewDataManager(&dataDBMockError{}, cfg, nil)
	acntPrf := &utils.Account{
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

	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 7)

	restoreAccounts(context.Background(), dm, []*utils.AccountWithLock{
		{Account: acntPrf, LockID: utils.EmptyString},
	}, []utils.AccountBalancesBackup{
		map[string]*decimal.Big{"CB1": decimal.New(100, 0)},
	})

	subString := "<AccountS> error <NOT_IMPLEMENTED> restoring account <cgrates.org:1001>"
	if rcv := buf.String(); !strings.Contains(rcv, subString) {
		t.Errorf("Expected %+q, received %+q", subString, rcv)
	}
}

func TestRestoreAccount3(t *testing.T) { //coverage purpose
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	acntPrf := &utils.Account{
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
	if err := dm.SetAccount(context.Background(), acntPrf, false); err != nil {
		t.Error(err)
	}

	restoreAccounts(context.Background(), dm, []*utils.AccountWithLock{
		{Account: acntPrf, LockID: utils.EmptyString},
	}, []utils.AccountBalancesBackup{
		nil,
	})
}

/*
cfg := config.NewDefaultCGRConfig()
func TestDebitFromBothBalances(t *testing.T) {
	engine.Cache.Clear(nil)
	data , _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	rates := rates2.NewRateS(cfg, fltr, dm)
	//RateS
	rpcClientConn := make(chan birpc.ClientConnector, 1)
	rpcClientConn <- rates
	connMngr := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS): rpcClientConn,
	})
	cfg.AccountSCfg().RateSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS)}
	//AccountS
	accnts := NewAccountS(cfg, fltr, connMngr, dm)

	accPrf := &utils.Account{
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

	if err := dm.SetAccount(context.Background(), accPrf, true); err != nil {
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
	rtPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:2003"},
		Rates: map[string]*utils.Rate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				FilterIDs:       nil,
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
						FixedFee:      utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rtPrf, true); err != nil {
		t.Error(err)
	}

	args := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestV1DebitID",
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "2003",
				utils.Usage:        "300",
			},
		},
	}

	var reply utils.ExtEventCharges
	exEvCh := utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(300),
	}
	if err := accnts.V1DebitAbstracts(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exEvCh, reply) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
	}

	accPrf.Balances["AbstractBalance"].Units = utils.NewDecimal(1200, 0)
	accPrf.Balances["ConcreteBalance2"].Units = utils.NewDecimal(49999999997, 9)
	//as we debited, the account is changed
	if rcvAcc, err := dm.GetAccount(context.Background(), accPrf.Tenant, accPrf.ID); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvAcc, accPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(accPrf), utils.ToJSON(rcvAcc))
	}

	if err := dm.RemoveAccount(context.Background(), accPrf.Tenant, accPrf.ID,
		utils.NonTransactional, true); err != nil {
		t.Error(err)
	} else if err := dm.RemoveRateProfile(context.Background(), rtPrf.Tenant, rtPrf.ID,
		utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

*/

func TestMaxDebitAbstractFromConcretesInsufficientCredit(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	cncrtBlncs := []*concreteBalance{
		{
			blnCfg: &utils.Balance{
				ID:   "CB1",
				Type: utils.MetaAbstract,
				UnitFactors: []*utils.UnitFactor{
					{
						FilterIDs: []string{"*test"},
						Factor:    utils.NewDecimal(10, 0), // EuroCents
					},
				},
				Units: utils.NewDecimal(80, 0), // 500 EuroCents
			},
			fltrS: filters,
		},
		{
			blnCfg: &utils.Balance{
				ID:    "CB2",
				Type:  utils.MetaConcrete,
				Units: utils.NewDecimal(1, 0),
			},
			fltrS: filters,
		},
	}

	expectedErr := "inline parse error for string: <*test>"
	if _, err := maxDebitAbstractsFromConcretes(context.Background(), decimal.New(110, 0), utils.EmptyString,
		cncrtBlncs, nil, new(utils.CGREvent),
		nil, nil, nil, nil, &utils.CostIncrement{
			Increment:    utils.NewDecimal(2, 0),
			RecurrentFee: utils.NewDecimal(1, 0),
		}, decimal.New(0, 0)); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

}
