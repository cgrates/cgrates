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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestChargerSSetGetChargerProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
			APIOpts: map[string]interface{}{},
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
		Weight:       20,
	}
	if !reflect.DeepEqual(getRply, expectedGet) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expectedGet), utils.ToJSON(getRply))
	}
}

func TestChargerSSetGetChargerProfileErrMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
			APIOpts: map[string]interface{}{},
		}, &getRply)
	if err != nil {
		t.Error(err)
	}
}

func TestChargerSSetGetChargerProfileErrNotFound(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
			APIOpts: map[string]interface{}{},
		}, &getRply)
	if err != nil {
		t.Error(err)
	}
}

func TestChargerSSetChargerProfileErrMissingID(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	var setRply string
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
		&utils.PaginatorWithTenant{}, &getRply)
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
			Weight:       20,
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
		&utils.PaginatorWithTenant{}, &getRply2)
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
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	var getRply []string
	err := admS.GetChargerProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{}, &getRply)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND", err)
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSSetGetChargerProfileIDsErr2(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	var getRply []string
	err := admS.GetChargerProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{}, &getRply)
	if err == nil || err.Error() != "NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_IMPLEMENTED", err)
	}
	dm.DataDB().Flush(utils.EmptyString)
	engine.Cache = cacheInit
}

func TestChargerSSetGetRmvGetChargerProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
			APIOpts: map[string]interface{}{},
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
		Weight:       20,
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

	var getRply2 engine.ChargerProfile
	err = admS.GetChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
			APIOpts: map[string]interface{}{},
		}, &getRply2)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND", err)
	}
}

func TestChargerSSetGetRmvGetChargerProfileNoTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
	ext := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:           "1001",
			RunID:        utils.MetaDefault,
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
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
			APIOpts: map[string]interface{}{},
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
		Weight:       20,
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

	var getRply2 engine.ChargerProfile
	err = admS.GetChargerProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "1001",
			},
			APIOpts: map[string]interface{}{},
		}, &getRply2)
	if err == nil || err.Error() != "NOT_FOUND" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_FOUND", err)
	}
}

func TestChargerSRmvChargerProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr)
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
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
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
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
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
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
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
