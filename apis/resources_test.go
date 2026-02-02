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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestResourcesSetGetRemResourceProfile(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES_1",
		},
	}
	var result utils.ResourceProfile
	var reply string

	resPrf := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "RES_1",
			AllocationMessage: "Approved",
			Limit:             5,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
		},
	}
	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}

	if err := adms.GetResourceProfile(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *resPrf.ResourceProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(resPrf.ResourceProfile), utils.ToJSON(result))
	}

	var rsPrfIDs []string
	expRsPrfIDs := []string{"RES_1"}

	if err := adms.GetResourceProfileIDs(context.Background(), &utils.ArgsItemIDs{},
		&rsPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsPrfIDs, expRsPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expRsPrfIDs, rsPrfIDs)
	}

	var rplyCount int

	if err := adms.GetResourceProfilesCount(context.Background(), &utils.ArgsItemIDs{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(rsPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(rsPrfIDs), rplyCount)
	}

	argRem := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES_1",
		},
	}

	if err := adms.RemoveResourceProfile(context.Background(), argRem, &reply); err != nil {
		t.Error(err)
	}

	engine.Cache.Clear(nil)
	if err := adms.GetResourceProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv utils.ResourceProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetResourceProfile(context.Background(), &utils.TenantIDWithAPIOpts{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES_1",
		},
	}

	if err := adms.GetResourceProfile(context.Background(), arg, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesSetResourceProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	resPrf := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{},
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
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	experr = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := adms.SetResourceProfile(ctx, resPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>,\nreceived <%+v>", experr, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*utils.ResourceProfile, error) {
			resPrf := &utils.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *utils.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return nil, nil
		},
	}

	dbCm := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	adms.dm = engine.NewDataManager(dbCm, cfg, nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"
	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesRemoveResourceProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	resPrf := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
			ID:     "TestResourcesRemoveResourceProfileCheckErrors",
			Tenant: "cgrates.org",
			Limit:  5,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			AllocationMessage: "Approved",
		},
	}
	var reply string

	if err := adms.SetResourceProfile(context.Background(), resPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
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
	var rcv utils.ResourceProfile

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "RES_1",
		},
	}

	if err := adms.GetResourceProfile(context.Background(), arg, &rcv); err == nil || err != utils.ErrNotFound {
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
		GetResourceProfileDrvF: func(*context.Context, string, string) (*utils.ResourceProfile, error) {
			resPrf := &utils.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *utils.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		RemoveResourceDrvF: func(ctx *context.Context, tnt, id string) error {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, nil
		},
	}

	dbCm := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	adms.dm = engine.NewDataManager(dbCm, cfg, nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveResourceProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestResourcesRemoveResourceProfileCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileIDsErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*utils.ResourceProfile, error) {
			resPrf := &utils.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *utils.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "NOT_IMPLEMENTED"

	if err := adms.GetResourceProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileIDsErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{}, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string

	if err := adms.GetResourceProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfilesCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*utils.ResourceProfile, error) {
			resPrf := &utils.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return resPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *utils.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetResourceProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestResourcesGetResourceProfilesCountErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{}, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetResourceProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestResourcesGetResourceProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args1 := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "test_ID1",
			Limit:             10,
			AllocationMessage: "Approved",
			Stored:            true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetResourceProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "test_ID2",
			Limit:             15,
			AllocationMessage: "Approved",
			Stored:            false,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetResourceProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "test2_ID1",
			Limit:             10,
			AllocationMessage: "Approved",
			Stored:            false,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetResourceProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsSearch: "test_ID",
	}
	exp := []*utils.ResourceProfile{
		{
			Tenant:            "cgrates.org",
			ID:                "test_ID1",
			Limit:             10,
			AllocationMessage: "Approved",
			Stored:            true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		{
			Tenant:            "cgrates.org",
			ID:                "test_ID2",
			Limit:             15,
			AllocationMessage: "Approved",
			Stored:            false,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	var getReply []*utils.ResourceProfile
	if err := admS.GetResourceProfiles(context.Background(), argsGet, &getReply); err != nil {
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

func TestResourcesGetResourceProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil)
	args := &utils.ResourceProfileWithAPIOpts{
		ResourceProfile: &utils.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "test_ID1",
			Limit:             10,
			AllocationMessage: "Approved",
			Stored:            true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetResourceProfile(context.Background(), args, &setReply); err != nil {
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
	var getReply []*utils.ResourceProfile
	if err := admS.GetResourceProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestResourcesGetResourceProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetResourceProfileDrvF: func(*context.Context, *utils.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"rsp_cgrates.org:TEST"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*utils.ResourceProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetResourceProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsSearch: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestResourcesGetResourceProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*utils.ResourceProfile, error) {
			rsPrf := &utils.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return rsPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *utils.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"rsp_cgrates.org:key1", "rsp_cgrates.org:key2", "rsp_cgrates.org:key3"}, nil
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

	if err := adms.GetResourceProfileIDs(context.Background(),
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

func TestResourcesGetResourceProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetResourceProfileDrvF: func(*context.Context, string, string) (*utils.ResourceProfile, error) {
			rsPrf := &utils.ResourceProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return rsPrf, nil
		},
		SetResourceProfileDrvF: func(*context.Context, *utils.ResourceProfile) error {
			return nil
		},
		RemoveResourceProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"rsp_cgrates.org:key1", "rsp_cgrates.org:key2", "rsp_cgrates.org:key3"}, nil
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

	if err := adms.GetResourceProfileIDs(context.Background(),
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
