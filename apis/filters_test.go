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

func TestApisSetGetGetIDsCountFilters(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `"OK"`, utils.ToJSON(&reply))
	}

	var replyGet engine.Filter
	argsGet := &utils.TenantID{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_attr",
	}

	err = admS.GetFilter(context.Background(), argsGet, &replyGet)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&replyGet, fltr.Filter) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", fltr.Filter, &replyGet)
	}

	var replyCnt int
	argsCnt := &utils.TenantWithAPIOpts{
		Tenant: utils.CGRateSorg,
	}
	err = admS.GetFilterIDsCount(context.Background(), argsCnt, &replyCnt)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(&replyCnt), utils.ToJSON(1)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(1), utils.ToJSON(&replyCnt))
	}
	fltr2 := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr2",
		},
	}
	var reply2 string
	err = admS.SetFilter(context.Background(), fltr2, &reply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	var replyCnt2 int
	argsCnt2 := &utils.TenantWithAPIOpts{
		Tenant: utils.CGRateSorg,
	}
	err = admS.GetFilterIDsCount(context.Background(), argsCnt2, &replyCnt2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(&replyCnt2), utils.ToJSON(2)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(2), utils.ToJSON(&replyCnt2))
	}
	argRmv := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr2",
		},
		APIOpts: nil,
	}
	var replyRmv string
	err = admS.RemoveFilter(context.Background(), argRmv, &replyRmv)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(&replyRmv)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `"OK"`, utils.ToJSON(&replyRmv))
	}
	var replyCnt3 int
	argsCnt3 := &utils.TenantWithAPIOpts{
		Tenant: utils.CGRateSorg,
	}
	err = admS.GetFilterIDsCount(context.Background(), argsCnt3, &replyCnt3)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(&replyCnt3), utils.ToJSON(1)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(1), utils.ToJSON(&replyCnt3))
	}
	argRmv2 := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
		},
		APIOpts: nil,
	}
	var replyRmv2 string
	err = admS.RemoveFilter(context.Background(), argRmv2, &replyRmv2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(&replyRmv2)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `"OK"`, utils.ToJSON(&replyRmv2))
	}

	var replyCnt4 int
	argsCnt4 := &utils.TenantWithAPIOpts{}
	err = admS.GetFilterIDsCount(context.Background(), argsCnt4, &replyCnt4)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
	engine.Cache = cacheInit
}

func TestApisSetFiltersMissingField(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}
	if !reflect.DeepEqual(`""`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `""`, utils.ToJSON(&reply))
	}

	engine.Cache = cacheInit
}

func TestApisSetFiltersTenantEmpty(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "fltr_for_attr",
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `"OK"`, utils.ToJSON(&reply))
	}
	var replyGet engine.Filter
	argsGet := &utils.TenantID{
		Tenant: utils.CGRateSorg,
		ID:     "fltr_for_attr",
	}

	err = admS.GetFilter(context.Background(), argsGet, &replyGet)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&replyGet, fltr.Filter) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", fltr.Filter, &replyGet)
	}

	engine.Cache = cacheInit
}

func TestApisSetFiltersGetFilterError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(`""`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `""`, utils.ToJSON(&reply))
	}

	engine.Cache = cacheInit
}

func TestApisSetFiltersError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return &engine.Filter{
				Tenant: utils.CGRateSorg,
				ID:     "fltr_for_attr",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement",
						Values:  []string{"testValue1", "testValue2"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement2",
						Values:  []string{"testValue3", "testValue4"},
					},
				},
			}, nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    "testType",
					Element: "~testElement",
					Values:  []string{"testValue1", "testValue2"},
				},
				{
					Type:    "testType2",
					Element: "~testElement2",
					Values:  []string{"testValue3", "testValue4"},
				},
			},
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(`""`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `""`, utils.ToJSON(&reply))
	}

	engine.Cache = cacheInit
}

func TestApisSetFiltersSetFilterError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotFound
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    "testType",
					Element: "~testElement",
					Values:  []string{"testValue1", "testValue2"},
				},
				{
					Type:    "testType2",
					Element: "~testElement2",
					Values:  []string{"testValue3", "testValue4"},
				},
			},
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(`""`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `""`, utils.ToJSON(&reply))
	}

	engine.Cache = cacheInit
}

func TestApisSetFiltersComposeCacheArgsForFilterError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		SetFilterDrvF: func(ctx *context.Context, fltr *engine.Filter) error {

			return nil
		},
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return &engine.Filter{
				Tenant: utils.CGRateSorg,
				ID:     "fltr_for_attr",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement",
						Values:  []string{"testValue1", "testValue2"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement2",
						Values:  []string{"testValue3", "testValue4"},
					},
				},
			}, utils.ErrNotFound
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~testElement",
					Values:  []string{"testValue1", "testValue2"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~testElement2",
					Values:  []string{"testValue3", "testValue4"},
				},
			},
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(`""`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `""`, utils.ToJSON(&reply))
	}

	engine.Cache = cacheInit
}

func TestApisSetFiltersSetLoadIDsError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		SetFilterDrvF: func(ctx *context.Context, fltr *engine.Filter) error {
			return nil
		},
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return &engine.Filter{
				Tenant: utils.CGRateSorg,
				ID:     "fltr_for_attr",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement",
						Values:  []string{"testValue1", "testValue2"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement2",
						Values:  []string{"testValue3", "testValue4"},
					},
				},
			}, nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~testElement",
					Values:  []string{"testValue1", "testValue2"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~testElement2",
					Values:  []string{"testValue3", "testValue4"},
				},
			},
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}
	if !reflect.DeepEqual(`""`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `""`, utils.ToJSON(&reply))
	}

	engine.Cache = cacheInit
}
func TestApisSetFiltersCacheForFilterError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		SetFilterDrvF: func(ctx *context.Context, fltr *engine.Filter) error {
			return nil
		},
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return &engine.Filter{
				Tenant: utils.CGRateSorg,
				ID:     "fltr_for_attr",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement",
						Values:  []string{"testValue1", "testValue2"},
					},
					{
						Type:    utils.MetaPrefix,
						Element: "~testElement2",
						Values:  []string{"testValue3", "testValue4"},
					},
				},
			}, nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return nil, nil
		},
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~testElement",
					Values:  []string{"testValue1", "testValue2"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~testElement2",
					Values:  []string{"testValue3", "testValue4"},
				},
			},
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err == nil || err.Error() != "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]", err)
	}
	if !reflect.DeepEqual(`""`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `""`, utils.ToJSON(&reply))
	}

	engine.Cache = cacheInit
}

func TestApisGetFilterNoTenant(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "fltr_for_attr",
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `"OK"`, utils.ToJSON(&reply))
	}
	var replyGet engine.Filter
	argsGet := &utils.TenantID{
		ID: "fltr_for_attr",
	}

	err = admS.GetFilter(context.Background(), argsGet, &replyGet)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(&replyGet, fltr.Filter) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", fltr.Filter, &replyGet)
	}

	engine.Cache = cacheInit
}

func TestApisGetFilterMissingField(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	fltr := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "fltr_for_attr",
		},
	}
	var reply string
	err := admS.SetFilter(context.Background(), fltr, &reply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(`"OK"`, utils.ToJSON(&reply)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", `"OK"`, utils.ToJSON(&reply))
	}
	var replyGet engine.Filter
	argsGet := &utils.TenantID{}

	err = admS.GetFilter(context.Background(), argsGet, &replyGet)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}

	engine.Cache = cacheInit
}

func TestApisGetFilterGetFilterError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	var replyGet engine.Filter
	argsGet := &utils.TenantID{
		ID: "fltr_for_attr",
	}

	err := admS.GetFilter(context.Background(), argsGet, &replyGet)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}

	engine.Cache = cacheInit
}

func TestApisGetFilterIDsCountError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	var reply int
	args := &utils.TenantWithAPIOpts{}

	err := admS.GetFilterIDsCount(context.Background(), args, &reply)
	if err == nil || err.Error() != "NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_IMPLEMENTED", err)
	}

	engine.Cache = cacheInit
}

func TestApisRemoveFilterMissingStructFieldError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
		APIOpts:  nil,
	}

	err := admS.RemoveFilter(context.Background(), args, &reply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [ID]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "MANDATORY_IE_MISSING: [ID]", err)
	}

	engine.Cache = cacheInit
}

func TestApisRemoveFilterRemoveFilterError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return nil, utils.ErrNotImplemented
		},
	}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "",
			ID:     "testID",
		},
		APIOpts: nil,
	}

	err := admS.RemoveFilter(context.Background(), args, &reply)
	if err == nil || err.Error() != "SERVER_ERROR: NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: NOT_IMPLEMENTED", err)
	}

	engine.Cache = cacheInit
}
