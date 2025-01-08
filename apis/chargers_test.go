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
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestChargerSSetChargerProfileErrMissingID(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	err := admS.SetChargerProfile(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
}

func TestChargerSDmSetChargerProfileErr(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	err := admS.SetChargerProfile(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSSetChargerProfileSetLoadIDsErr(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		SetChargerProfileDrvF: func(ctx *context.Context, chr *engine.ChargerProfile) (err error) {
			return nil
		},
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*engine.ChargerProfile, error) {
			return nil, nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	err := admS.SetChargerProfile(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSSetChargerProfileCallCacheErr(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		SetChargerProfileDrvF: func(ctx *context.Context, chr *engine.ChargerProfile) (err error) {
			return nil
		},
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*engine.ChargerProfile, error) {
			return nil, utils.ErrNotFound
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	err := admS.SetChargerProfile(context.Background(), ext, &setRply)
	if err == nil || err.Error() != "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]", err)
	}
	if !reflect.DeepEqual(setRply, "") {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "", utils.ToJSON(setRply))
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSSetGetChargerProfileIDs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}

	var setRply string
	err := admS.SetChargerProfile(context.Background(), ext, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply []string
	err = admS.GetChargerProfileIDs(context.Background(),
		&utils.ArgsItemIDs{}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := []string{"1001"}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}
	var setRply2 string
	extSet2 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "1002",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}

	err2 := admS.SetChargerProfile(context.Background(), extSet2, &setRply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err2)
	}
	if !reflect.DeepEqual(setRply2, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply2))
	}

	var getRply2 []string
	err = admS.GetChargerProfileIDs(context.Background(),
		&utils.ArgsItemIDs{}, &getRply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	sort.Strings(getRply2)
	expectedGet2 := []string{"1001", "1002"}
	if !reflect.DeepEqual(getRply2, expectedGet2) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet2), utils.ToJSON(getRply2))
	}
}

func TestChargerSSetGetChargerProfileIDsErr(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	var getRply []string
	err := admS.GetChargerProfileIDs(context.Background(),
		&utils.ArgsItemIDs{}, &getRply)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND", err)
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSSetGetChargerProfileIDsErr2(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	var getRply []string
	err := admS.GetChargerProfileIDs(context.Background(),
		&utils.ArgsItemIDs{}, &getRply)
	if err == nil || err.Error() != "NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_IMPLEMENTED", err)
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSSetGetRmvGetChargerProfile(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	var setRply string
	err := admS.SetChargerProfile(context.Background(), ext, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply engine.ChargerProfile
	err = admS.GetChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
			APIOpts: map[string]any{},
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "1001",
		RunID:        utils.MetaDefault,
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}
	var rmvRply string
	err = admS.RemoveChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
			APIOpts: nil,
		}, &rmvRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedRmv := "OK"
	if !reflect.DeepEqual(rmvRply, expectedRmv) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedRmv), utils.ToJSON(rmvRply))
	}
	engine.Cache.Clear(nil)

	var getRply2 engine.ChargerProfile
	err = admS.GetChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
			APIOpts: map[string]any{},
		}, &getRply2)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND", err)
	}
}

func TestChargerSSetGetRmvGetChargerProfileNoTenant(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	var setRply string
	err := admS.SetChargerProfile(context.Background(), ext, &setRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(setRply, `OK`) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `OK`, utils.ToJSON(setRply))
	}
	var getRply engine.ChargerProfile
	err = admS.GetChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
			APIOpts: map[string]any{},
		}, &getRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedGet := engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "1001",
		RunID:        utils.MetaDefault,
		FilterIDs:    []string{"*string:~*req.Account:1001"},
		AttributeIDs: []string{"*none"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
	}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}
	var rmvRply string
	err = admS.RemoveChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "1001",
			},
			APIOpts: nil,
		}, &rmvRply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	expectedRmv := "OK"
	if !reflect.DeepEqual(rmvRply, expectedRmv) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedRmv), utils.ToJSON(rmvRply))
	}
	engine.Cache.Clear(nil)

	var getRply2 engine.ChargerProfile
	err = admS.GetChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
			APIOpts: map[string]any{},
		}, &getRply2)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND", err)
	}
}

func TestChargerSRmvChargerProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var rmvRply string
	err := admS.RemoveChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{},
			APIOpts:  nil,
		}, &rmvRply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}

}

func TestChargerSRmvChargerProfileErrRemoveChargerProfile(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var rmvRply string
	err := admS.RemoveChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "1001",
			},
			APIOpts: nil,
		}, &rmvRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSRmvChargerProfileErrSetLoadIDs(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		RemoveChargerProfileDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*engine.ChargerProfile, error) {
			return &engine.ChargerProfile{
				Tenant: "cgrates.org",
			}, nil
		},
		SetChargerProfileDrvF: func(ctx *context.Context, chr *engine.ChargerProfile) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var rmvRply string
	err := admS.RemoveChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "1001",
			},
			APIOpts: nil,
		}, &rmvRply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSRmvChargerProfileErrRemoveCallCache(t *testing.T) {
	cacheInit := engine.Cache
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		RemoveChargerProfileDrvF: func(ctx *context.Context, str1 string, str2 string) error {
			return nil
		},
		GetChargerProfileDrvF: func(ctx *context.Context, tnt, id string) (*engine.ChargerProfile, error) {
			return &engine.ChargerProfile{
				Tenant: "cgrates.org",
			}, nil
		},
		SetChargerProfileDrvF: func(ctx *context.Context, chr *engine.ChargerProfile) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, connMgr, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	var rmvRply string
	err := admS.RemoveChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "1001",
			},
			APIOpts: nil,
		}, &rmvRply)
	if err == nil || err.Error() != "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]", err)
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargersGetChargerProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args1 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetChargerProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetChargerProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "test2_ID1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetChargerProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*engine.ChargerProfile{
		{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	var getReply []*engine.ChargerProfile
	if err := admS.GetChargerProfiles(context.Background(), argsGet, &getReply); err != nil {
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

func TestChargersGetChargerProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetChargerProfile(context.Background(), args, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
		APIOpts: map[string]any{
			utils.PageLimitOpt:    2,
			utils.PageOffsetOpt:   4,
			utils.PageMaxItemsOpt: 5,
		},
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	var getReply []*engine.ChargerProfile
	if err := admS.GetChargerProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestChargersGetChargerProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetChargerProfileDrvF: func(*context.Context, *engine.ChargerProfile) error {
			return nil
		},
		RemoveChargerProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"chp_cgrates.org:TEST"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*engine.ChargerProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetChargerProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestChargersGetChargerProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetChargerProfileDrvF: func(*context.Context, string, string) (*engine.ChargerProfile, error) {
			chrgPrf := &engine.ChargerProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return chrgPrf, nil
		},
		SetChargerProfileDrvF: func(*context.Context, *engine.ChargerProfile) error {
			return nil
		},
		RemoveChargerProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"cpp_cgrates.org:key1", "cpp_cgrates.org:key2", "cpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetChargerProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt: true,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestChargersGetChargerProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetChargerProfileDrvF: func(*context.Context, string, string) (*engine.ChargerProfile, error) {
			chrgPrf := &engine.ChargerProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return chrgPrf, nil
		},
		SetChargerProfileDrvF: func(*context.Context, *engine.ChargerProfile) error {
			return nil
		},
		RemoveChargerProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"cpp_cgrates.org:key1", "cpp_cgrates.org:key2", "cpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetChargerProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
			APIOpts: map[string]any{
				utils.PageLimitOpt:    2,
				utils.PageOffsetOpt:   4,
				utils.PageMaxItemsOpt: 5,
			},
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestChargersGetChargerProfilesCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetChargerProfileDrvF: func(*context.Context, string, string) (*engine.ChargerProfile, error) {
			chrgPrf := &engine.ChargerProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return chrgPrf, nil
		},
		SetChargerProfileDrvF: func(*context.Context, *engine.ChargerProfile) error {
			return nil
		},
		RemoveChargerProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetChargerProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestChargersGetChargerProfilesCountErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetChargerProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestChargersSetGetRemChargerProfile(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "chrgPrf",
		},
	}
	var result engine.ChargerProfile
	var reply string

	chrgPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "chrgPrf",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	if err := adms.SetChargerProfile(context.Background(), chrgPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetChargerProfile(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *chrgPrf.ChargerProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(chrgPrf.ChargerProfile), utils.ToJSON(result))
	}

	var thPrfIDs []string
	expThPrfIDs := []string{"chrgPrf"}

	if err := adms.GetChargerProfileIDs(context.Background(), &utils.ArgsItemIDs{},
		&thPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thPrfIDs, expThPrfIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expThPrfIDs, thPrfIDs)
	}

	var rplyCount int

	if err := adms.GetChargerProfilesCount(context.Background(), &utils.ArgsItemIDs{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(thPrfIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", len(thPrfIDs), rplyCount)
	}

	if err := adms.RemoveChargerProfile(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	if err := adms.GetChargerProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestChargersGetChargerProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.ChargerProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetChargerProfile(context.Background(), &utils.TenantIDWithAPIOpts{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestChargersGetChargerProfileCheckErrors",
		},
	}

	if err := adms.GetChargerProfile(context.Background(), arg, &rcv); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestChargersNewChargerSv1(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	chS := engine.NewChargerService(dm, nil, cfg, nil)

	exp := &ChargerSv1{
		cS: chS,
	}
	rcv := NewChargerSv1(chS)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestChargersAPIs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ChargerSCfg().AttributeSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)}
	cfg.AttributeSCfg().Opts.ProcessRuns = []*config.DynamicIntOpt{
		{
			Value: 2,
		},
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileIDs: []string{"ATTR1", "ATTR2"},
			utils.MetaChargeID:             "",
			utils.OptsContext:              utils.MetaChargers,
			utils.MetaSubsys:               utils.MetaChargers,
			utils.MetaRunID:                "run_1",
		},
	}

	mCC := &mockClientConn{
		calls: map[string]func(*context.Context, any, any) error{
			utils.AttributeSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				expEv.APIOpts[utils.MetaChargeID] = args.(*utils.CGREvent).APIOpts[utils.MetaChargeID]
				if !reflect.DeepEqual(args, expEv) {
					return fmt.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expEv), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- mCC
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, rpcInternal)

	adms := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}

	argsCharger1 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "CHARGER1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			RunID:        "run_1",
			AttributeIDs: []string{"ATTR1", "ATTR2"},
			FilterIDs:    []string{"*string:~*req.Account:1001"},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := adms.SetChargerProfile(context.Background(), argsCharger1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsCharger2 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "CHARGER2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			RunID:        "run_2",
			AttributeIDs: []string{"ATTR3"},
			FilterIDs:    []string{"*string:~*req.Account:1001"},
		},
		APIOpts: nil,
	}

	if err := adms.SetChargerProfile(context.Background(), argsCharger2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	cS := engine.NewChargerService(dm, fltrs, cfg, cM)
	cSv1 := NewChargerSv1(cS)

	argsGetForEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	exp := engine.ChargerProfiles{
		{

			Tenant: "cgrates.org",
			ID:     "CHARGER2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			RunID:        "run_2",
			AttributeIDs: []string{"ATTR3"},
			FilterIDs:    []string{"*string:~*req.Account:1001"},
		},
		{
			Tenant: "cgrates.org",
			ID:     "CHARGER1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			RunID:        "run_1",
			AttributeIDs: []string{"ATTR1", "ATTR2"},
			FilterIDs:    []string{"*string:~*req.Account:1001"},
		},
	}
	var reply engine.ChargerProfiles
	if err := cSv1.GetChargersForEvent(context.Background(), argsGetForEvent, &reply); err != nil {
		t.Error(err)
	} else {
		if utils.ToJSON(reply) != utils.ToJSON(exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(reply))
		}
	}

	argsCharger2 = &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "CHARGER2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			RunID:        "run_2",
			AttributeIDs: []string{"ATTR3"},
			FilterIDs:    []string{"*string:~*req.Account:1002"},
		},
		APIOpts: nil,
	}

	if err := adms.SetChargerProfile(context.Background(), argsCharger2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsProcessEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{},
	}
	expProcessEv := []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile: "CHARGER1",
			AlteredFields: []*engine.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "EventTest",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.OptsAttributesProfileIDs: []string{"ATTR1", "ATTR2"},
					utils.MetaChargeID:             "",
					utils.OptsContext:              utils.MetaChargers,
					utils.MetaSubsys:               utils.MetaChargers,
					utils.MetaRunID:                "run_1",
				},
			},
		},
	}
	var replyProcessEv []*engine.ChrgSProcessEventReply
	if err := cSv1.ProcessEvent(context.Background(), argsProcessEv, &replyProcessEv); err != nil {
		t.Error(err)
	} else {
		expProcessEv[0].CGREvent.APIOpts[utils.MetaChargeID] = replyProcessEv[0].CGREvent.APIOpts[utils.MetaChargeID]
		if !reflect.DeepEqual(replyProcessEv, expProcessEv) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expProcessEv), utils.ToJSON(replyProcessEv))
		}
	}
}

func TestChargersSv1Ping(t *testing.T) {
	cSv1 := new(ChargerSv1)
	var reply string
	if err := cSv1.Ping(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply error")
	}
}
