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
	"time"

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
	if err == nil || err.Error() != "SERVER_ERROR: broken reference to filter: <*string*req.Account1001> for item with ID: cgrates.org:test_ID1" {
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
	err = admS.GetAccountCount(context.Background(),
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
	err = admS.GetAccountCount(context.Background(),
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
	err = admS.GetAccountCount(context.Background(),
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
	err = admS.GetAccountCount(context.Background(),
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

func TestAccountGetAccountCountError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	var getRplyCount3 int
	err := admS.GetAccountCount(context.Background(),
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
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment: utils.Float64Pointer(0.1),
						},
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
	} else if setRply != utils.OK {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply utils.Account
	err = admS.GetAccount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "test_ID1",
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
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(1, 1),
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
			accKEy: {
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
						CostIncrements: []*utils.ExtCostIncrement{
							{
								Increment: utils.Float64Pointer(0.1),
							},
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
			accKEy: {
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

func TestAccountActionSetBalance(t *testing.T) {
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
			accKEy: {
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

func TestAccountActionRemoveBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := accounts.NewAccountS(cfg, fltr, nil, dm)
	accSv1 := NewAccountSv1(accnts)
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

	if err := accSv1.ActionSetBalance(context.Background(), argsSet, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected status reply", reply)
	}

	//remove it
	args := &utils.ArgsActRemoveBalances{}

	expected := "MANDATORY_IE_MISSING: [AccountID]"
	if err := accSv1.ActionRemoveBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.AccountID = "TestV1ActionRemoveBalance"

	expected = "MANDATORY_IE_MISSING: [BalanceIDs]"
	if err := accSv1.ActionRemoveBalance(context.Background(), args, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	args.BalanceIDs = []string{"AbstractBalance1"}

	if err := accSv1.ActionRemoveBalance(context.Background(), args, &reply); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	}

}

func TestAccountMaxConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := accounts.NewAccountS(cfg, fltr, nil, dm)
	accSv1 := NewAccountSv1(accnts)
	admS := NewAdminSv1(cfg, dm, nil)
	accPrf := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{

			Tenant:    "cgrates.org",
			ID:        "TestV1DebitAbstracts",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";15",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
				"ConcreteBalance1": {
					ID:      "ConcreteBalance1",
					Weights: ";25",
					Type:    utils.MetaConcrete,
					Units:   float64(time.Minute),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
				"ConcreteBalance2": {
					ID:      "ConcreteBalance2",
					Weights: ";5",
					Type:    utils.MetaConcrete,
					Units:   float64(30 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
			},
		},
		APIOpts: nil,
	}
	var setRpl string
	if err := admS.SetAccount(context.Background(), accPrf, &setRpl); err != nil {
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
	if err := accSv1.MaxConcretes(context.Background(), args, &reply); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	}

	if err := accSv1.MaxConcretes(context.Background(), args, &reply); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	}
	accPrf.Balances["AbstractBalance1"].Weights = ""

	extAccPrf := &utils.ExtAccount{
		Tenant:    "cgrates.org",
		ID:        "TestV1DebitAbstracts",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.ExtBalance{
			"AbstractBalance1": {
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 15,
					},
				},
				Type:  utils.MetaAbstract,
				Units: utils.Float64Pointer(float64(40 * time.Second)),
				CostIncrements: []*utils.ExtCostIncrement{
					{
						Increment:    utils.Float64Pointer(float64(time.Second)),
						FixedFee:     utils.Float64Pointer(float64(0)),
						RecurrentFee: utils.Float64Pointer(float64(1)),
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
				Units: utils.Float64Pointer(float64(time.Minute)),
				CostIncrements: []*utils.ExtCostIncrement{
					{
						Increment:    utils.Float64Pointer(float64(time.Second)),
						FixedFee:     utils.Float64Pointer(float64(0)),
						RecurrentFee: utils.Float64Pointer(float64(1)),
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
				Units: utils.Float64Pointer(float64(30 * time.Second)),
				CostIncrements: []*utils.ExtCostIncrement{
					{
						Increment:    utils.Float64Pointer(float64(time.Second)),
						FixedFee:     utils.Float64Pointer(float64(0)),
						RecurrentFee: utils.Float64Pointer(float64(1)),
					},
				},
			},
		},
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
	if err := accSv1.MaxConcretes(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else {
		exEvCh.Charges = reply.Charges
		exEvCh.Accounting = reply.Accounting
		if !reflect.DeepEqual(exEvCh, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exEvCh), utils.ToJSON(reply))
		}
	}
}

func TestAccountDebitConcretes(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltr := engine.NewFilterS(cfg, nil, dm)
	accnts := accounts.NewAccountS(cfg, fltr, nil, dm)
	accSv1 := NewAccountSv1(accnts)
	admS := NewAdminSv1(cfg, dm, nil)
	accPrf := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{

			Tenant:    "cgrates.org",
			ID:        "TestV1DebitAbstracts",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";15",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
				"ConcreteBalance1": {
					ID:      "ConcreteBalance1",
					Weights: ";25",
					Type:    utils.MetaConcrete,
					Units:   float64(time.Minute),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
				"ConcreteBalance2": {
					ID:      "ConcreteBalance2",
					Weights: ";5",
					Type:    utils.MetaConcrete,
					Units:   float64(30 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
			},
		},
		APIOpts: nil,
	}
	var setRpl string
	if err := admS.SetAccount(context.Background(), accPrf, &setRpl); err != nil {
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
	if err := accSv1.DebitConcretes(context.Background(), args, &reply); err != nil {
		t.Errorf("Expected %+v, received %+v", nil, err)
	}

}
