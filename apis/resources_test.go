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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestResourcesSetGetRemResourceProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantID{
		ID: "RES_1",
	}
	var result engine.ResourceProfile
	var reply string

	resPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES_1",
			AllocationMessage: "Approved",
			Limit:             5,
			Weight:            10,
		},
	}
	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}

	if err := adms.GetResourceProfile(context.Background(), arg, &result); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(result, *resPrf.ResourceProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(resPrf.ResourceProfile), utils.ToJSON(result))
	}

	var rsPrfIDs []string
	expRsPrfIDs := []string{"RES_1"}

	if err := adms.GetResourceProfileIDs(context.Background(), &utils.PaginatorWithTenant{},
		&rsPrfIDs); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(rsPrfIDs, expRsPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expRsPrfIDs, rsPrfIDs)
	}

	var rplyCount int

	if err := adms.GetResourceProfileCount(context.Background(), &utils.TenantWithAPIOpts{},
		&rplyCount); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if rplyCount != len(rsPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(rsPrfIDs), rplyCount)
	}

	argRem := &utils.TenantIDWithAPIOpts{
		TenantID: arg,
	}

	if err := adms.RemoveResourceProfile(context.Background(), argRem, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if err := adms.GetResourceProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.ResourceProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetResourceProfile(context.Background(), &utils.TenantID{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.GetResourceProfile(context.Background(), &utils.TenantID{
		ID: "TestResourcesGetResourceProfileCheckErrors",
	}, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesSetResourceProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	resPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{},
	}

	var reply string
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	resPrf.ID = "TestResourcesSetResourceProfileCheckErrors"
	resPrf.FilterIDs = []string{"invalid_filter_format"}
	experr = "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TestResourcesSetResourceProfileCheckErrors"

	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	resPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	experr = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := adms.SetResourceProfile(ctx, resPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>,\nreceived <%+v>", experr, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*engine.ResourceProfile, error) {
			resPrf := &engine.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *engine.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesRemoveResourceProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	resPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			ID:                "TestResourcesRemoveResourceProfileCheckErrors",
			Tenant:            "cgrates.org",
			Limit:             5,
			Weight:            10,
			AllocationMessage: "Approved",
		},
	}
	var reply string

	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
	experr := "SERVER_ERROR: context deadline exceeded"

	if err := adms.RemoveResourceProfile(ctx, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestResourcesRemoveResourceProfileCheckErrors",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	cancel()

	adms.cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	var rcv engine.ResourceProfile

	if err := adms.GetResourceProfile(context.Background(), &utils.TenantID{
		ID: "TestResourcesRemoveResourceProfileCheckErrors",
	}, &rcv); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	experr = "MANDATORY_IE_MISSING: [ID]"

	if err := adms.RemoveResourceProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.RemoveResourceProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestResourcesRemoveResourceProfileCheckErrors",
		}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*engine.ResourceProfile, error) {
			resPrf := &engine.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *engine.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		RemoveResourceDrvF: func(ctx *context.Context, tnt, id string) error {
			return nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveResourceProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestResourcesRemoveResourceProfileCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileIDsErrMock(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*engine.ResourceProfile, error) {
			resPrf := &engine.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *engine.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "NOT_IMPLEMENTED"

	if err := adms.GetResourceProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileIDsErrKeys(t *testing.T) {
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

	var reply []string

	if err := adms.GetResourceProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileCountErrMock(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*engine.ResourceProfile, error) {
			resPrf := &engine.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *engine.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetResourceProfileCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestResourcesGetResourceProfileCountErrKeys(t *testing.T) {
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

	if err := adms.GetResourceProfileCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesNewResourceSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	rls := engine.NewResourceService(dm, cfg, nil, nil)

	exp := &ResourceSv1{
		rls: rls,
	}
	rcv := NewResourceSv1(rls)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestResourcesSv1Ping(t *testing.T) {
	resSv1 := new(ResourceSv1)
	var reply string
	if err := resSv1.Ping(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply error")
	}
}

func TestResourcesGetResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rls := engine.NewResourceService(dm, cfg, fltrs, nil)
	adms := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}

	resPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "rsID",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             5,
			AllocationMessage: "Approved",
			Weight:            10,
		},
	}

	var reply string
	if err := adms.SetResourceProfile(context.Background(), resPrf,
		&reply); err != nil {
		t.Error(err)
	}

	rsv1 := NewResourceSv1(rls)
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		UsageID: "RU_Test",
	}

	expResources := engine.Resources{
		{
			Tenant: "cgrates.org",
			ID:     "rsID",
			Usages: make(map[string]*engine.ResourceUsage),
		},
	}

	var rplyResources engine.Resources
	if err := rsv1.GetResourcesForEvent(context.Background(), args, &rplyResources); err != nil {
		t.Error(err)
	} else {
		// We compare JSONs because the received Resources have unexported fields
		if utils.ToJSON(expResources) != utils.ToJSON(rplyResources) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expResources), utils.ToJSON(rplyResources))
		}
	}

	expResource := engine.Resource{
		Tenant: "cgrates.org",
		ID:     "rsID",
		Usages: make(map[string]*engine.ResourceUsage),
	}

	var rplyResource engine.Resource
	if err := rsv1.GetResource(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "rsID",
	}}, &rplyResource); err != nil {
		t.Error(err)
	} else {
		// We compare JSONs because the received Resource has unexported fields
		if utils.ToJSON(rplyResource) != utils.ToJSON(expResource) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expResource), utils.ToJSON(rplyResource))
		}
	}

	expResourceWithCfg := engine.ResourceWithConfig{
		Resource: &engine.Resource{
			Tenant: "cgrates.org",
			ID:     "rsID",
			Usages: make(map[string]*engine.ResourceUsage),
		},
		Config: resPrf.ResourceProfile,
	}

	var rplyResourceWithCfg engine.ResourceWithConfig
	if err := rsv1.GetResourceWithConfig(context.Background(), &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "rsID",
	}}, &rplyResourceWithCfg); err != nil {
		t.Error(err)
	} else {
		// We compare JSONs because the received Resource has unexported fields
		if utils.ToJSON(expResourceWithCfg) != utils.ToJSON(rplyResourceWithCfg) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expResourceWithCfg), utils.ToJSON(rplyResourceWithCfg))
		}
	}
	dm.DataDB().Flush(utils.EmptyString)
}

func TestResourcesAuthorizeAllocateReleaseResource(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	rls := engine.NewResourceService(dm, cfg, fltrs, nil)
	adms := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}

	resPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "rsID",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             5,
			AllocationMessage: "Approved",
			Weight:            10,
		},
	}

	var reply string
	if err := adms.SetResourceProfile(context.Background(), resPrf,
		&reply); err != nil {
		t.Error(err)
	}

	rsv1 := NewResourceSv1(rls)
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		UsageID: "RU_Test",
	}

	if err := rsv1.AuthorizeResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "Approved", reply)
	}

	if err := rsv1.AllocateResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "Approved", reply)
	}

	if err := rsv1.ReleaseResources(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}
}
