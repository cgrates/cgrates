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
	"testing"

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

func TestAccountSetGetAccountError(t *testing.T) {
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
