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

package apis

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/accounts"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAccountSetGetAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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

func TestAccountGetAccountErrorMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
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

func TestAccountGetAccountErrorNotFound(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestAccountGetAccountErrorGetAccount(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestAccountSetGetAccountNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			ID:   "test_ID1",
			Opts: map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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

func TestAccountSetGetAccountErrorMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Opts: map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
}

func TestAccountSetGetAccountErrorAsAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			ID:   "test_ID1",
			Opts: map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   "12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: "10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "invalid DynamicWeight format for string <10>" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "invalid DynamicWeight format for string <10>", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
}

func TestAccountSetGetAccountErrorBadFilter(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "",
			ID:        "test_ID1",
			FilterIDs: []string{"*string*req.Account1001"},
			Weights:   ";10",
			Opts:      map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string*req.Account1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			ThresholdIDs: nil,
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: broken reference to filter: *string*req.Account1001 for item with ID: cgrates.org:test_ID1" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: broken reference to filter: *string*req.Account1001 for item with ID: cgrates.org:test_ID1", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
}

func TestAccountSetGetAccountErrorSetLoadIDs(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
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
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:  "",
			ID:      "test_ID1",
			Weights: ";10",
			Opts:    map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:      "VoiceBalance",
					Weights: ";12",
					Type:    "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			ThresholdIDs: nil,
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
	engine.Cache = cacheInit
}

func TestAccountSetGetAccountErrorCallCache(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg, nil)
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
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:  "",
			ID:      "test_ID1",
			Weights: ";10",
			Opts:    map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:      "VoiceBalance",
					Weights: ";12",
					Type:    "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			ThresholdIDs: nil,
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
	engine.Cache = cacheInit
}

func TestAccountSetGetAccountIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "testTenant",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant:    "testTenant",
			Paginator: utils.Paginator{},
			APIOpts:   nil,
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := []string{"test_ID1"}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}

	var getRplyCount int
	err = admS.GetAccountIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant:  "testTenant",
			APIOpts: nil,
		}, &getRplyCount)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(getRplyCount, 1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "1", utils.ToJSON(getRplyCount))
	}

	ext2 := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "testTenant",
			ID:     "test_ID2",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply2 string
	err = admS.SetAccount(context.Background(), ext2, &setRply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply2, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply2))
	}
	var getRply3 []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant:    "testTenant",
			Paginator: utils.Paginator{},
			APIOpts:   nil,
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
	err = admS.GetAccountIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant:  "testTenant",
			APIOpts: nil,
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
	err = admS.GetAccountIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant:  "testTenant",
			APIOpts: nil,
		}, &getRplyCount2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(getRplyCount2, 1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "1", utils.ToJSON(getRplyCount2))
	}

	var getRplyID []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant:    "testTenant",
			Paginator: utils.Paginator{},
			APIOpts:   nil,
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
	err = admS.GetAccountIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant:  "testTenant",
			APIOpts: nil,
		}, &getRplyCount3)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}

	var getRplyID2 []string
	err = admS.GetAccountIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant:    "testTenant",
			Paginator: utils.Paginator{},
			APIOpts:   nil,
		}, &getRplyID2)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
}

func TestAccountGetAccountIDSCountError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	var getRplyCount3 int
	err := admS.GetAccountIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			APIOpts: nil,
		}, &getRplyCount3)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestAccountGetAccountIDSError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	var getRplyCount3 []string
	err := admS.GetAccountIDs(context.Background(),
		&utils.PaginatorWithTenant{
			APIOpts: nil,
		}, &getRplyCount3)
	if err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestAccountRemoveAccountErrorMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestAccountRemoveAccountErrorRmvAccount(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestAccountRemoveAccountErrorSetLoadIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
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
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestAccountRemoveAccountErrorCallCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg, nil)
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
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
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

func TestAccountNewAccountSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	result1 := accounts.NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	expected := &AccountSv1{
		aS: result1,
	}
	result2 := NewAccountSv1(result1)
	if !reflect.DeepEqual(expected, result2) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(result2))
	}
}

func TestAccountAccountsForEvent(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:      "VoiceBalance",
					Weights: ";12",
					Type:    "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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

	accS := accounts.NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	expected := &AccountSv1{
		aS: accS,
	}
	accSv1 := NewAccountSv1(accS)
	if !reflect.DeepEqual(expected, accSv1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(accSv1))
	}
	rpEv := make([]*utils.Account, 0)
	accArg := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestMatchingAccountsForEvent",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	err = accSv1.AccountsForEvent(context.Background(), accArg, &rpEv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expEvAcc := []*utils.Account{
		{

			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
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
					Opts: map[string]interface{}{
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

func TestAccountMaxAbstracts(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:      "VoiceBalance",
					Weights: ";12",
					Type:    "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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

	accS := accounts.NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	expected := &AccountSv1{
		aS: accS,
	}
	accSv1 := NewAccountSv1(accS)
	if !reflect.DeepEqual(expected, accSv1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(accSv1))
	}
	var rpEv utils.ExtEventCharges
	accArg := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestMatchingAccountsForEvent",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	err = accSv1.MaxAbstracts(context.Background(), accArg, &rpEv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	var accKEy, rtID string
	for key, val := range rpEv.Accounting {
		accKEy = key
		rtID = val.RatingID
	}
	var crgID string
	for _, val := range rpEv.Charges {
		crgID = val.ChargingID
	}
	expRating := &utils.ExtRateSInterval{
		IntervalStart: nil,
		Increments: []*utils.ExtRateSIncrement{
			{
				IncrementStart:    nil,
				IntervalRateIndex: 0,
				RateID:            "",
				CompressFactor:    0,
				Usage:             nil,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range rpEv.Rating {
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	rpEv.Rating = map[string]*utils.ExtRateSInterval{}
	expEvAcc := &utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			accKEy: &utils.ExtAccountCharge{
				AccountID:       "test_ID1",
				BalanceID:       "VoiceBalance",
				Units:           utils.Float64Pointer(0),
				BalanceLimit:    utils.Float64Pointer(0),
				UnitFactorID:    "",
				RatingID:        rtID,
				JoinedChargeIDs: nil,
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"test_ID1": {
				Tenant: "cgrates.org",
				ID:     "test_ID1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Balances: map[string]*utils.ExtBalance{
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: nil,
								Weight:    12,
							},
						},
						Type: "*abstract",
						Opts: map[string]interface{}{
							"Destination": 10,
						},
						Units: utils.Float64Pointer(0),
					},
				},
				Opts: map[string]interface{}{},
			},
		},
	}
	eql := rpEv.Equals(expEvAcc)
	if eql != true {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", true, eql)
	}
	engine.Cache = cacheInit
}

func TestAccountDebitAbstracts(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:      "VoiceBalance",
					Weights: ";12",
					Type:    "*abstract",
					Opts: map[string]interface{}{
						"Destination": 10,
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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

	accS := accounts.NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	expected := &AccountSv1{
		aS: accS,
	}
	accSv1 := NewAccountSv1(accS)
	if !reflect.DeepEqual(expected, accSv1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(accSv1))
	}
	var rpEv utils.ExtEventCharges
	accArg := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestMatchingAccountsForEvent",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	err = accSv1.DebitAbstracts(context.Background(), accArg, &rpEv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	var accKEy, rtID string
	for key, val := range rpEv.Accounting {
		accKEy = key
		rtID = val.RatingID
	}
	var crgID string
	for _, val := range rpEv.Charges {
		crgID = val.ChargingID
	}
	expRating := &utils.ExtRateSInterval{
		IntervalStart: nil,
		Increments: []*utils.ExtRateSIncrement{
			{
				IncrementStart:    nil,
				IntervalRateIndex: 0,
				RateID:            "",
				CompressFactor:    0,
				Usage:             nil,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range rpEv.Rating {
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	rpEv.Rating = map[string]*utils.ExtRateSInterval{}
	expEvAcc := &utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			accKEy: &utils.ExtAccountCharge{
				AccountID:       "test_ID1",
				BalanceID:       "VoiceBalance",
				Units:           utils.Float64Pointer(0),
				BalanceLimit:    utils.Float64Pointer(0),
				UnitFactorID:    "",
				RatingID:        rtID,
				JoinedChargeIDs: nil,
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"test_ID1": {
				Tenant: "cgrates.org",
				ID:     "test_ID1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Balances: map[string]*utils.ExtBalance{
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: nil,
								Weight:    12,
							},
						},
						Type: "*abstract",
						Opts: map[string]interface{}{
							"Destination": 10,
						},
						Units: utils.Float64Pointer(0),
					},
				},
				Opts: map[string]interface{}{},
			},
		},
	}
	eql := rpEv.Equals(expEvAcc)
	if eql != true {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", true, eql)
	}
	engine.Cache = cacheInit
}

/*
func TestAccountMaxConcretes(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	accS := accounts.NewAccountS(cfg, &engine.FilterS{}, connMgr, dm)
	engine.Cache = newCache
	ext := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:      "VoiceBalance",
					Weights: ";12",
					Type:    "*abstract",
					Opts:    map[string]interface{}{},
					Units:   1,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetAccount(context.Background(), ext, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	expected := &AccountSv1{
		aS: accS,
	}

	accSv1 := NewAccountSv1(accS)
	if !reflect.DeepEqual(expected, accSv1) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(accSv1))
	}
	var rpEv utils.ExtEventCharges
	accArg := &utils.ArgsAccountsForEvent{
		CGREvent: &utils.CGREvent{
			ID:     "TestMaxConcretes",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Usage: 6,
			},
		},
	}
	err = accSv1.MaxConcretes(context.Background(), accArg, &rpEv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	var accKEy, rtID string
	for key, val := range rpEv.Accounting {
		accKEy = key
		rtID = val.RatingID
	}
	var crgID string
	for _, val := range rpEv.Charges {
		crgID = val.ChargingID
	}
	expRating := &utils.ExtRateSInterval{
		IntervalStart: nil,
		Increments: []*utils.ExtRateSIncrement{
			{
				IncrementStart:    nil,
				IntervalRateIndex: 0,
				RateID:            "",
				CompressFactor:    0,
				Usage:             nil,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range rpEv.Rating {
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	rpEv.Rating = map[string]*utils.ExtRateSInterval{}
	expEvAcc := &utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			accKEy: &utils.ExtAccountCharge{
				AccountID:       "test_ID1",
				BalanceID:       "VoiceBalance",
				Units:           utils.Float64Pointer(0),
				BalanceLimit:    utils.Float64Pointer(0),
				UnitFactorID:    "",
				RatingID:        rtID,
				JoinedChargeIDs: nil,
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"test_ID1": {
				Tenant: "cgrates.org",
				ID:     "test_ID1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Balances: map[string]*utils.ExtBalance{
					"VoiceBalance": {
						ID: "VoiceBalance",
						Weights: utils.DynamicWeights{
							{
								FilterIDs: nil,
								Weight:    12,
							},
						},
						Type: "*abstract",
						Opts: map[string]interface{}{
							"Destination": 10,
						},
						Units: utils.Float64Pointer(0),
					},
				},
				Opts: map[string]interface{}{},
			},
		},
	}
	eql := rpEv.Equals(expEvAcc)
	if eql != true {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", true, eql)
	}
	engine.Cache = cacheInit
}
*/
