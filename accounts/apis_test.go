/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package accounts

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAccountsRefundCharges(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
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
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
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

func TestAccountsDebitConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	if err := dm.SetAccount(context.Background(),
		&utils.Account{

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
		}, true); err != nil {
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
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)
	if err := dm.SetAccount(context.Background(),
		&utils.Account{
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
		}, true); err != nil {
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
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	if err := dm.SetAccount(context.Background(),
		&utils.Account{
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
		}, true); err != nil {
		t.Errorf("expected <%+v>,\nreceived <%+v>", nil, err)
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
	if err := accS.V1DebitAbstracts(context.Background(), ev, &rpEv); err != nil {
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
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	accS := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	engine.Cache = newCache
	err := dm.SetAccount(context.Background(),
		&utils.Account{
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
		}, true)
	if err != nil {
		t.Fatal(err)
	}

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
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	err := dm.SetAccount(context.Background(),
		&utils.Account{
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
		}, true)
	if err != nil {
		t.Fatal(err)
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
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	err := dm.SetAccount(context.Background(),
		&utils.Account{
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
		}, true)
	if err != nil {
		t.Fatal(err)
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

func TestV1GetAccount(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := NewAccountS(cfg, fltr, nil, dm)

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org"},
	}
	var reply utils.Account

	expected := "MANDATORY_IE_MISSING: [ID]"
	err := accnts.V1GetAccount(context.Background(), arg, &reply)
	if err == nil || err.Error() != expected {
		t.Errorf("Expected %v, got %v", expected, err)
	}

	arg.TenantID.ID = "unknown_acc"
	err = accnts.V1GetAccount(context.Background(), arg, &reply)
	if err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}

	acc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "acc1",
	}
	if err := dm.SetAccount(context.Background(), acc, false); err != nil {
		t.Fatalf("failed to create test account: %v", err)
	}

	arg.TenantID.ID = "acc1"
	err = accnts.V1GetAccount(context.Background(), arg, &reply)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if reply.ID != acc.ID || reply.Tenant != acc.Tenant {
		t.Errorf("expected %+v, got %+v", acc, reply)
	}
}

func TestV1RefundCharges(t *testing.T) {
	engine.Cache.Clear(nil)

	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	accnts := NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)

	ctx := context.Background()
	var reply string

	acc1 := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "acc1",
		Balances: map[string]*utils.Balance{
			"balance1": {ID: "balance1", Units: utils.NewDecimal(100, 0)},
		},
	}
	acc2 := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "acc2",
		Balances: map[string]*utils.Balance{
			"balance2": {ID: "balance2", Units: utils.NewDecimal(50, 0)},
		},
	}

	if err := dm.SetAccount(ctx, acc1, false); err != nil {
		t.Fatalf("failed to create account1: %v", err)
	}
	if err := dm.SetAccount(ctx, acc2, false); err != nil {
		t.Fatalf("failed to create account2: %v", err)
	}

	eventCharges := &utils.EventCharges{
		Accounts: map[string]*utils.Account{
			acc1.ID: acc1,
			acc2.ID: acc2,
		},
		Charges: []*utils.ChargeEntry{
			{ChargingID: "charge1"},
		},
		Accounting: map[string]*utils.AccountCharge{
			"charge1": {
				AccountID: "acc1",
				BalanceID: "balance1",
				Units:     utils.NewDecimal(10, 0),
			},
		},
	}

	args := &utils.APIEventCharges{
		Tenant:       "cgrates.org",
		EventCharges: eventCharges,
	}

	if err := accnts.V1RefundCharges(ctx, args, &reply); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != utils.OK {
		t.Fatalf("expected reply OK, got %v", reply)
	}

	acc1After, _ := dm.GetAccount(ctx, acc1.Tenant, acc1.ID)
	expected := utils.NewDecimal(110, 0)
	if acc1After.Balances["balance1"].Units.Big.Cmp(expected.Big) != 0 {
		t.Errorf("expected balance1 %v, got %v", expected, acc1After.Balances["balance1"].Units)
	}

	acc2After, _ := dm.GetAccount(ctx, acc2.Tenant, acc2.ID)
	expected2 := utils.NewDecimal(50, 0)
	if acc2After.Balances["balance2"].Units.Big.Cmp(expected2.Big) != 0 {
		t.Errorf("expected balance2 %v, got %v", expected2, acc2After.Balances["balance2"].Units)
	}
}
