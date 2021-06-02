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

func TestFiltersSetGetGetCountFilters(t *testing.T) {
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
	err = admS.GetFilterCount(context.Background(), argsCnt, &replyCnt)
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
	err = admS.GetFilterCount(context.Background(), argsCnt2, &replyCnt2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(&replyCnt2), utils.ToJSON(2)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(2), utils.ToJSON(&replyCnt2))
	}

	args5 := &utils.PaginatorWithTenant{
		Tenant:    "",
		Paginator: utils.Paginator{},
		APIOpts:   nil,
	}
	var reply5 []string
	err = admS.GetFilterIDs(context.Background(), args5, &reply5)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	sort.Strings(reply5)
	if !reflect.DeepEqual(reply5, []string{"fltr_for_attr", "fltr_for_attr2"}) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", []string{"fltr_for_attr", "fltr_for_attr2"}, reply5)
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
	err = admS.GetFilterCount(context.Background(), argsCnt3, &replyCnt3)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(utils.ToJSON(&replyCnt3), utils.ToJSON(1)) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(1), utils.ToJSON(&replyCnt3))
	}

	args6 := &utils.PaginatorWithTenant{
		Tenant:    "",
		Paginator: utils.Paginator{},
		APIOpts:   nil,
	}
	var reply6 []string
	err = admS.GetFilterIDs(context.Background(), args6, &reply6)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(reply6, []string{"fltr_for_attr"}) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", []string{"fltr_for_attr"}, reply6)
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
	err = admS.GetFilterCount(context.Background(), argsCnt4, &replyCnt4)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}

	args7 := &utils.PaginatorWithTenant{
		Tenant:    "",
		Paginator: utils.Paginator{},
		APIOpts:   nil,
	}
	var reply7 []string
	err = admS.GetFilterIDs(context.Background(), args7, &reply7)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}
	engine.Cache = cacheInit
}

func TestFiltersSetFiltersMissingField(t *testing.T) {
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

func TestFiltersSetFiltersTenantEmpty(t *testing.T) {
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

func TestFiltersSetFiltersGetFilterError(t *testing.T) {
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

func TestFiltersSetFiltersError(t *testing.T) {
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

func TestFiltersSetFiltersSetFilterError(t *testing.T) {
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

func TestFiltersSetFiltersComposeCacheArgsForFilterError(t *testing.T) {
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

func TestFiltersSetFiltersSetLoadIDsError(t *testing.T) {
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
func TestFiltersSetFiltersCacheForFilterError(t *testing.T) {
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

func TestFiltersGetFilterNoTenant(t *testing.T) {
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

func TestFiltersGetFilterMissingField(t *testing.T) {
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

func TestFiltersGetFilterGetFilterError(t *testing.T) {
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

func TestFiltersGetFilterCountError(t *testing.T) {
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

	err := admS.GetFilterCount(context.Background(), args, &reply)
	if err == nil || err.Error() != "NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_IMPLEMENTED", err)
	}

	engine.Cache = cacheInit
}

func TestFiltersRemoveFilterMissingStructFieldError(t *testing.T) {
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

func TestFiltersRemoveFilterRemoveFilterError(t *testing.T) {
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

type mockClientConn struct {
	calls map[string]func(ctx *context.Context, args interface{}, reply interface{}) error
}

func (mCC *mockClientConn) Call(ctx *context.Context, serviceMethod string, args interface{}, reply interface{}) (err error) {
	if call, has := mCC.calls[serviceMethod]; !has {
		return utils.ErrUnsupporteServiceMethod
	} else {
		return call(ctx, args, reply)
	}
}

func TestFiltersSetFilterReloadCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	expArgs := &utils.AttrReloadCacheWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
		Tenant: "cgrates.org",
		ArgsCache: map[string][]string{
			utils.FilterIDs: {"cgrates.org:FLTR_ID"},
		},
	}
	ccM := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1ReloadCache: func(ctx *context.Context, args, reply interface{}) error {
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	adms := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: cM,
	}
	arg := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
	}
	var reply string

	if err := adms.SetFilter(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			FilterIDs: []string{"FLTR_ID"},
			ID:        "ATTR_ID",
			Weight:    10,
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.Account",
					Value: "1003",
					Type:  utils.MetaConstant,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetAttributeProfile(context.Background(), attrPrf, &reply); err != nil {
		t.Error(err)
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			ID:        "THD_ID",
			FilterIDs: []string{"FLTR_ID"},
			MaxHits:   10,
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err != nil {
		t.Error(err)
	}

	rsPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			ID:        "RES_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetResourceProfile(context.Background(), rsPrf, &reply); err != nil {
		t.Error(err)
	}

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			ID:        "SQ_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err != nil {
		t.Error(err)
	}

	arg = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1002"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
	}
	expArgs = &utils.AttrReloadCacheWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaReload,
		},
		Tenant: "cgrates.org",
		ArgsCache: map[string][]string{
			utils.FilterIDs:               {"cgrates.org:FLTR_ID"},
			utils.AttributeFilterIndexIDs: {"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
			utils.ResourceFilterIndexIDs:  {"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
			utils.StatFilterIndexIDs:      {"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
			utils.ThresholdFilterIndexIDs: {"cgrates.org:*string:*req.Account:1001", "cgrates.org:*string:*req.Account:1002"},
		},
	}

	if err := adms.SetFilter(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestFiltersSetFilterClearCache(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.AdminSCfg().CachesConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	expArgs := &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
		Tenant:   "cgrates.org",
		CacheIDs: []string{utils.CacheFilters},
	}
	ccM := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.CacheSv1Clear: func(ctx *context.Context, args, reply interface{}) error {
				sort.Strings(args.(*utils.AttrCacheIDsWithAPIOpts).CacheIDs)
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- ccM
	cM := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): rpcInternal,
	})
	adms := &AdminSv1{
		cfg:     cfg,
		dm:      dm,
		connMgr: cM,
	}
	arg := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
	}
	var reply string

	if err := adms.SetFilter(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			FilterIDs: []string{"FLTR_ID"},
			ID:        "ATTR_ID",
			Weight:    10,
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.Account",
					Value: "1003",
					Type:  utils.MetaConstant,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetAttributeProfile(context.Background(), attrPrf, &reply); err != nil {
		t.Error(err)
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			ID:        "THD_ID",
			FilterIDs: []string{"FLTR_ID"},
			MaxHits:   10,
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err != nil {
		t.Error(err)
	}

	rsPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			ID:        "RES_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetResourceProfile(context.Background(), rsPrf, &reply); err != nil {
		t.Error(err)
	}

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			ID:        "SQ_ID",
			FilterIDs: []string{"FLTR_ID"},
			Weight:    10,
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err != nil {
		t.Error(err)
	}

	arg = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			ID: "FLTR_ID",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1002"},
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
	}
	expArgs = &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		},
		Tenant: "cgrates.org",
		CacheIDs: []string{utils.CacheAttributeFilterIndexes, utils.CacheThresholdFilterIndexes,
			utils.CacheResourceFilterIndexes, utils.CacheStatFilterIndexes,
			utils.CacheFilters},
	}
	sort.Strings(expArgs.CacheIDs)

	if err := adms.SetFilter(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}
func TestFiltersRemoveFilterSetLoadIDsError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.CacheCfg().ReplicationConns = []string{"rep"}
	cfg.CacheCfg().Partitions[utils.CacheReverseFilterIndexes].Replicate = false
	cfg.RPCConns()["connID"] = &config.RPCConn{}
	config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes].Remote = true
	config.CgrConfig().DataDbCfg().RmtConns = []string{"connID"}
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return &engine.Filter{}, nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, utils.ErrNotFound
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotFound
		},
		RemoveFilterDrvF: func(str1 string, str2 string) error {
			return nil
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
	config.CgrConfig().DataDbCfg().RmtConns = []string{}
	config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes].Remote = false
	engine.Cache = cacheInit
}

func TestFiltersRemoveFilterCallCacheForFilterError(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = "123"
	cfg.CacheCfg().ReplicationConns = []string{"rep"}
	cfg.CacheCfg().Partitions[utils.CacheReverseFilterIndexes].Replicate = false
	cfg.RPCConns()["connID"] = &config.RPCConn{}
	config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes].Remote = true
	config.CgrConfig().DataDbCfg().RmtConns = []string{"connID"}
	cfg.AdminSCfg().CachesConns = []string{}
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{
		SetLoadIDsDrvF: func(ctx *context.Context, loadIDs map[string]int64) error {
			return nil
		},
		GetFilterDrvF: func(ctx *context.Context, str1 string, str2 string) (*engine.Filter, error) {
			return &engine.Filter{}, nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, utils.ErrNotFound
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return utils.ErrNotFound
		},
		RemoveFilterDrvF: func(str1 string, str2 string) error {
			return nil
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
	if err == nil || err.Error() != "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "SERVER_ERROR: MANDATORY_IE_MISSING: [connIDs]", err)
	}
	config.CgrConfig().DataDbCfg().RmtConns = []string{}
	config.CgrConfig().DataDbCfg().Items[utils.MetaIndexes].Remote = false
	engine.Cache = cacheInit
}

func TestFiltersGetFilterIDs(t *testing.T) {
	cacheInit := engine.Cache
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg, nil)
	dataDB := &engine.DataDBMock{}
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	newCache := engine.NewCacheS(cfg, dm, nil)
	engine.Cache = newCache
	admS := NewAdminSv1(cfg, dm, connMgr)

	args6 := &utils.PaginatorWithTenant{
		Tenant:    "",
		Paginator: utils.Paginator{},
		APIOpts:   nil,
	}
	var reply6 []string
	err := admS.GetFilterIDs(context.Background(), args6, &reply6)
	if err == nil || err.Error() != "NOT_IMPLEMENTED" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "NOT_IMPLEMENTED", err)
	}

	engine.Cache = cacheInit
}
