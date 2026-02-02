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

package apis

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAccountsSetGetAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
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
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
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
}

func TestAccountsGetAccountErrorMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
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
			},
			APIOpts: nil,
		}, &getRply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
}

func TestAccountsGetAccountErrorNotFound(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRply utils.Account
	err := admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "test_ID1",
			},
			APIOpts: nil,
		}, &getRply)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
	engine.Cache = cacheInit
}

func TestAccountsGetAccountErrorGetAccount(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRply utils.Account
	err := admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "test_ID1",
			},
			APIOpts: nil,
		}, &getRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	engine.Cache = cacheInit
}

func TestAccountsSetGetAccountNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			ID:   "test_ID1",
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
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
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
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
}

func TestAccountsSetGetAccountErrorMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
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
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
}

func TestAccountsSetGetAccountErrorBadFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant:    "",
			ID:        "test_ID1",
			FilterIDs: []string{"*string*req.Account1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string*req.Account1001"},
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
			ThresholdIDs: nil,
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), args, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: broken reference to filter: <*string*req.Account1001> for item with ID: cgrates.org:test_ID1" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: broken reference to filter: *string*req.Account1001 for item with ID: cgrates.org:test_ID1", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
}

func TestAccountsSetGetAccountErrorSetLoadIDs(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return nil
		},
		GetAccountDrvF: func(ctx *context.Context, str1 string, str2 string) (*utils.Account, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "",
			ID:     "test_ID1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
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
			ThresholdIDs: nil,
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), args, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
	engine.Cache = cacheInit
}

func TestAccountsSetGetAccountErrorCallCache(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return nil
		},
		GetAccountDrvF: func(ctx *context.Context, str1 string, str2 string) (*utils.Account, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "",
			ID:     "test_ID1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Opts: map[string]any{},
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
			ThresholdIDs: nil,
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), args, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
	engine.Cache = cacheInit
}

func TestAccountsSetGetAccountIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "testTenant",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
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
	var getRply []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := []string{"test_ID1"}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}

	var getRplyCount int
	err = admS.GetAccountsCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRplyCount)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(getRplyCount, 1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "1", utils.ToJSON(getRplyCount))
	}

	args2 := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "testTenant",
			ID:     "test_ID2",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
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

	var setRply2 string
	err = admS.SetAccount(context.Background(), args2, &setRply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply2, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply2))
	}
	var getRply3 []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRply3)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet3 := []string{"test_ID1", "test_ID2"}
	sort.Strings(getRply3)
	if !reflect.DeepEqual(getRply3, expectedGet3) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet3), utils.ToJSON(getRply3))
	}
	var getRply4 int
	err = admS.GetAccountsCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRply4)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(getRply4, 2) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "2", utils.ToJSON(getRply4))
	}

	var getRplyRmv string
	err = admS.RemoveAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "testTenant",
				ID:     "test_ID2",
			},
			APIOpts: nil,
		}, &getRplyRmv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(getRplyRmv, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(getRplyRmv))
	}

	var getRplyCount2 int
	err = admS.GetAccountsCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRplyCount2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(getRplyCount2, 1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "1", utils.ToJSON(getRplyCount2))
	}

	var getRplyID []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRplyID)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedID := []string{"test_ID1"}
	if !reflect.DeepEqual(getRplyID, expectedID) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedID), utils.ToJSON(getRplyID))
	}
	var getNtf utils.Account
	err = admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "test_ID2",
			},
			APIOpts: nil,
		}, &getNtf)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}

	var getRplyRmv2 string
	err = admS.RemoveAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "testTenant",
				ID:     "test_ID1",
			},
			APIOpts: nil,
		}, &getRplyRmv2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(getRplyRmv2, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(getRplyRmv2))
	}

	var getRplyCount3 int
	err = admS.GetAccountsCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRplyCount3)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}

	var getRplyID2 []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "testTenant",
		}, &getRplyID2)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestAccountsGetAccountsCountError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRplyCount3 int
	err := admS.GetAccountsCount(context.Background(),
		&utils.ArgsItemIDs{}, &getRplyCount3)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestAccountsGetAccountIDSError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRplyCount3 []string
	err := admS.GetAccountIDs(context.Background(),
		&utils.ArgsItemIDs{}, &getRplyCount3)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestAccountsRemoveAccountErrorMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRplyRmv string
	err := admS.RemoveAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "",
			},
			APIOpts: nil,
		}, &getRplyRmv)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
}

func TestAccountsRemoveAccountErrorRmvAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRplyRmv string
	err := admS.RemoveAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "id",
			},
			APIOpts: nil,
		}, &getRplyRmv)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
}

func TestAccountsRemoveAccountErrorSetLoadIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1 string, str2 string) (*utils.Account, error) {
			return &utils.Account{}, nil
		},
		RemoveAccountDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRplyRmv string
	err := admS.RemoveAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "id",
			},
			APIOpts: nil,
		}, &getRplyRmv)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
}

func TestAccountsRemoveAccountErrorCallCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetAccountDrvF: func(ctx *context.Context, str1 string, str2 string) (*utils.Account, error) {
			return &utils.Account{}, nil
		},
		RemoveAccountDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
		SetAccountDrvF: func(ctx *context.Context, profile *utils.Account) error {
			return nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	var getRplyRmv string
	err := admS.RemoveAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "id",
			},
			APIOpts: nil,
		}, &getRplyRmv)
	if err == nil || err.Error() != "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]", err)
	}
}

func TestAccountsGetAccountsOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args1 := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
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

	var setReply string
	if err := admS.SetAccount(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetAccount(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test2_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
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

	if err := admS.SetAccount(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsSearch: "test_ID",
	}
	exp := []*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}

	var getReply []*utils.Account
	if err := admS.GetAccounts(context.Background(), argsGet, &getReply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(getReply, func(i, j int) bool {
			return getReply[i].ID < getReply[j].ID
		})
		if !reflect.DeepEqual(getReply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(getReply))
		}
	}
}

func TestAccountsGetAccountsGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID: "VoiceBalance",
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

	var setReply string
	if err := admS.SetAccount(context.Background(), args, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsSearch: "test_ID",
		APIOpts: map[string]any{
			utils.PageLimitOpt:    2,
			utils.PageOffsetOpt:   4,
			utils.PageMaxItemsOpt: 5,
		},
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	var getReply []*utils.Account
	if err := admS.GetAccounts(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAccountsGetAccountsGetErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetAccountDrvF: func(*context.Context, *utils.Account) error {
			return nil
		},
		RemoveAccountDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"acn_cgrates.org:TEST"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*utils.Account
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetAccounts(context.Background(),
		&utils.ArgsItemIDs{
			ItemsSearch: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestAccountsGetAccountIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetAccountDrvF: func(*context.Context, string, string) (*utils.Account, error) {
			accPrf := &utils.Account{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return accPrf, nil
		},
		SetAccountDrvF: func(*context.Context, *utils.Account) error {
			return nil
		},
		RemoveAccountDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"acn_cgrates.org:key1", "acn_cgrates.org:key2", "acn_cgrates.org:key3"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetAccountIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt: true,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestAccountsGetAccountIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetAccountDrvF: func(*context.Context, string, string) (*utils.Account, error) {
			accPrf := &utils.Account{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return accPrf, nil
		},
		SetAccountDrvF: func(*context.Context, *utils.Account) error {
			return nil
		},
		RemoveAccountDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"acn_cgrates.org:key1", "acn_cgrates.org:key2", "acn_cgrates.org:key3"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetAccountIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt:    2,
				utils.PageOffsetOpt:   4,
				utils.PageMaxItemsOpt: 5,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}
