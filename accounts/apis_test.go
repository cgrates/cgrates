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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAccountsRefundCharges(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	acc := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	var reply string

	args := &utils.APIEventCharges{
		Tenant: "cgrates.org",
		EventCharges: &utils.EventCharges{
			Abstracts: nil,
		},
	}

	if err := acc.V1RefundCharges(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if reply != utils.OK {
		t.Errorf("Expected %v\n but received %v", utils.OK, reply)
	}
}

func TestAccountsActionRemoveBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
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

	if err := accnts.V1ActionRemoveBalance(context.Background(), args, &reply); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	}
}

func TestAccountsGetAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	acc := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	fltrs := engine.NewFilterS(cfg, connMgr, dm)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	admS := apis.NewAdminSv1(cfg, dm, nil, fltrs, nil)
	acc_args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": 10,
					},
					CostIncrements: []*utils.CostIncrement{
						{
							RecurrentFee: utils.NewDecimal(1, 0),
							Increment:    utils.NewDecimal(1, 1),
						},
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), acc_args, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	var reply utils.Account

	args := &utils.TenantIDWithAPIOpts{
		TenantID: utils.NewTenantID("cgrates.org:test_ID1"),
	}

	if err := acc.V1GetAccount(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(&reply, acc_args.Account) {
		t.Errorf("Expected %v\n but received %v", reply, acc_args.Account)
	}
}

func TestAccountsDebitConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	admS := apis.NewAdminSv1(cfg, dm, nil, nil, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{

			Tenant:    "cgrates.org",
			ID:        "TestV1DebitAbstracts",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance1": {
					ID: "AbstractBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 15,
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
		},
		APIOpts: nil,
	}
	var setRpl string
	if err := admS.SetAccount(context.Background(), args, &setRpl); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		ID:     "TestV1DebitID",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Usage:        "3m",
		},
	}
	reply := utils.EventCharges{}
	if err := accnts.V1DebitConcretes(context.Background(), ev, &reply); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	}

}

func TestAccountsMaxConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	admS := apis.NewAdminSv1(cfg, dm, nil, nil, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant:    "cgrates.org",
			ID:        "TestV1DebitAbstracts",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance1": {
					ID: "AbstractBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 15,
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
		},
		APIOpts: nil,
	}
	var setRpl string
	if err := admS.SetAccount(context.Background(), args, &setRpl); err != nil {
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
				Units:        utils.NewDecimal(int64(time.Minute), 0),
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
			"TestV1DebitAbstracts": {
				Tenant:    "cgrates.org",
				ID:        "TestV1DebitAbstracts",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Balances: map[string]*utils.Balance{
					"AbstractBalance1": {
						ID:   "AbstractBalance1",
						Type: utils.MetaAbstract,
						Weights: []*utils.DynamicWeight{
							{
								Weight: 15,
							},
						},
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
						Weights: []*utils.DynamicWeight{
							{
								Weight: 25,
							},
						},
						Type:  utils.MetaConcrete,
						Units: utils.NewDecimal(0, 0),
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
						Weights: []*utils.DynamicWeight{
							{
								Weight: 5,
							},
						},
						Type:  utils.MetaConcrete,
						Units: utils.NewDecimal(0, 0),
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(int64(time.Second), 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(1, 0),
							},
						},
					},
				},
			},
		},
	}
	if err := accnts.V1MaxConcretes(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else {
		if !exEvCh.Equals(&reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}

	// check the account was not debited
	extAccPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "TestV1DebitAbstracts",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance1": {
				ID:   "AbstractBalance1",
				Type: utils.MetaAbstract,
				Weights: utils.DynamicWeights{
					{
						Weight: 15,
					},
				},
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
	if rplyAcc, err := dm.GetAccount(context.Background(), "cgrates.org", "TestV1DebitAbstracts"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyAcc, extAccPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(extAccPrf), utils.ToJSON(rplyAcc))
	}
}

func TestAccountsActionSetBalance(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.AccountSCfg().RateSConns = []string{"*internal"}
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := apis.NewAdminSv1(cfg, dm, connMgr, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": 10,
					},
					CostIncrements: []*utils.CostIncrement{
						{
							RecurrentFee: utils.NewDecimal(1, 1),
							Increment:    utils.NewDecimal(1, 1),
						},
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), args, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply utils.Account
	err = admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "test_ID1",
			},
			APIOpts: nil,
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := utils.Account{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID: "VoiceBalance",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": 10,
				},
				CostIncrements: []*utils.CostIncrement{
					{
						RecurrentFee: utils.NewDecimal(1, 1),
						Increment:    utils.NewDecimal(1, 1),
					},
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}

	accS := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)

	var rpEv utils.EventCharges
	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	err = accS.V1DebitAbstracts(context.Background(), ev, &rpEv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	expRating := &utils.RateSInterval{
		IntervalStart: nil,
		Increments: []*utils.RateSIncrement{
			{
				RateIntervalIndex: 0,
				RateID:            "id_for_test",
				CompressFactor:    1,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range rpEv.Rating {
		val.Increments[0].RateID = "id_for_test"
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	rpEv.Rating = map[string]*utils.RateSInterval{}
	expEvAcc := &utils.EventCharges{
		Abstracts:   utils.NewDecimal(0, 0),
		Accounting:  map[string]*utils.AccountCharge{},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"test_ID1": {
				Tenant: "cgrates.org",
				ID:     "test_ID1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Balances: map[string]*utils.Balance{
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: nil,
								Weight:    12,
							},
						},
						Type: "*abstract",
						Opts: map[string]any{
							"Destination": 10,
						},
						CostIncrements: []*utils.CostIncrement{
							{
								RecurrentFee: utils.NewDecimal(1, 1),
								Increment:    utils.NewDecimal(1, 1),
							},
						},
						Units: utils.NewDecimal(0, 0),
					},
				},
				Opts: map[string]any{},
			},
		},
	}
	eql := rpEv.Equals(expEvAcc)
	if eql != true {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", true, eql)
	}
	engine.Cache = cacheInit
}

func TestAccountsDebitAbstracts(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.AccountSCfg().RateSConns = []string{"*internal"}
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := apis.NewAdminSv1(cfg, dm, connMgr, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": 10,
					},
					CostIncrements: []*utils.CostIncrement{
						{
							RecurrentFee: utils.NewDecimal(1, 0),
							Increment:    utils.NewDecimal(1, 1),
						},
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), args, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply utils.Account
	err = admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "test_ID1",
			},
			APIOpts: nil,
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := utils.Account{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID: "VoiceBalance",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": 10,
				},
				Units: utils.NewDecimal(0, 0),
				CostIncrements: []*utils.CostIncrement{
					{
						RecurrentFee: utils.NewDecimal(1, 0),
						Increment:    utils.NewDecimal(1, 1),
					},
				},
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}

	accS := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)

	var rpEv utils.EventCharges
	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	err = accS.V1DebitAbstracts(context.Background(), ev, &rpEv)
	if err != nil {
		t.Error(err)

	}

	expRating := &utils.RateSInterval{
		Increments: []*utils.RateSIncrement{
			{
				RateIntervalIndex: 0,
				RateID:            "id_for_test",
				CompressFactor:    1,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range rpEv.Rating {
		val.Increments[0].RateID = "id_for_test"
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expRating), utils.ToJSON(val))
		}
	}
	rpEv.Rating = map[string]*utils.RateSInterval{}
	expEvAcc := &utils.EventCharges{
		Abstracts:   utils.NewDecimal(0, 0),
		Accounting:  map[string]*utils.AccountCharge{},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"test_ID1": {
				Tenant: "cgrates.org",
				ID:     "test_ID1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Balances: map[string]*utils.Balance{
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: nil,
								Weight:    12,
							},
						},
						Type: "*abstract",
						Opts: map[string]any{
							"Destination": 10,
						},
						CostIncrements: []*utils.CostIncrement{
							{
								RecurrentFee: utils.NewDecimal(1, 0),
								Increment:    utils.NewDecimal(1, 1),
							},
						},
						Units: utils.NewDecimal(0, 0),
					},
				},
				Opts: map[string]any{},
			},
		},
	}
	eql := rpEv.Equals(expEvAcc)
	if eql != true {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", true, eql)
	}
	engine.Cache = cacheInit
}

func TestAccountsMaxAbstracts(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.AccountSCfg().RateSConns = []string{"*internal"}
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := apis.NewAdminSv1(cfg, dm, connMgr, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": 10,
					},
					CostIncrements: []*utils.CostIncrement{
						{
							RecurrentFee: utils.NewDecimal(1, 1),
							Increment:    utils.NewDecimal(1, 1),
						},
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), args, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	} else if setRply != utils.OK {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply utils.Account
	err = admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "test_ID1",
			},
			APIOpts: map[string]any{},
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := utils.Account{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID: "VoiceBalance",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": 10,
				},
				Units: utils.NewDecimal(0, 0),
				CostIncrements: []*utils.CostIncrement{
					{
						RecurrentFee: utils.NewDecimal(1, 1),
						Increment:    utils.NewDecimal(1, 1),
					},
				},
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}

	cfg.AccountSCfg().RateSConns = []string{"*internal"}
	accS := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)

	var rpEv utils.EventCharges
	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	err = accS.V1MaxAbstracts(context.Background(), ev, &rpEv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	expRating := &utils.RateSInterval{
		IntervalStart: nil,
		Increments: []*utils.RateSIncrement{
			{
				IncrementStart:    nil,
				RateIntervalIndex: 0,
				RateID:            "id_for_Test",
				CompressFactor:    1,
				Usage:             nil,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range rpEv.Rating {
		val.Increments[0].RateID = "id_for_Test"
		if !reflect.DeepEqual(expRating, val) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expRating), utils.ToJSON(val))
		}
	}
	rpEv.Rating = map[string]*utils.RateSInterval{}
	expEvAcc := &utils.EventCharges{
		Abstracts:   utils.NewDecimal(0, 0),
		Accounting:  map[string]*utils.AccountCharge{},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"test_ID1": {
				Tenant: "cgrates.org",
				ID:     "test_ID1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Balances: map[string]*utils.Balance{
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: nil,
								Weight:    12,
							},
						},
						Type: "*abstract",
						Opts: map[string]any{
							"Destination": 10,
						},
						CostIncrements: []*utils.CostIncrement{
							{
								RecurrentFee: utils.NewDecimal(1, 1),
								Increment:    utils.NewDecimal(1, 1),
							},
						},
						Units: utils.NewDecimal(0, 0),
					},
				},
				Opts: map[string]any{},
			},
		},
	}
	eql := rpEv.Equals(expEvAcc)
	if eql != true {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expEvAcc), utils.ToJSON(rpEv))
	}
	engine.Cache = cacheInit
}

func TestAccountsAccountsForEvent(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := apis.NewAdminSv1(cfg, dm, connMgr, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": 10,
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), args, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply utils.Account
	err = admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "test_ID1",
			},
			APIOpts: nil,
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := utils.Account{
		Tenant: "cgrates.org",
		ID:     "test_ID1",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID: "VoiceBalance",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": 10,
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}

	accS := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)

	rpEv := make([]*utils.Account, 0)
	ev := &utils.CGREvent{
		ID:     "TestMatchingAccountsForEvent",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	err = accS.V1AccountsForEvent(context.Background(), ev, &rpEv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expEvAcc := []*utils.Account{
		{

			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							FilterIDs: nil,
							Weight:    12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": 10,
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: nil,
					Weight:    10,
				},
			},
		},
	}
	if !reflect.DeepEqual(rpEv, expEvAcc) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expEvAcc), utils.ToJSON(rpEv))
	}
	engine.Cache = cacheInit
}
