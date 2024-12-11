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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestActionsSetGetRemActionProfile(t *testing.T) {
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
			ID: "actID",
		},
	}
	var result engine.ActionProfile
	var reply string

	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actID",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Actions: make([]*engine.APAction, 1),
		},
	}

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetActionProfile(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *actPrf.ActionProfile) {
		t.Errorf("expected: <%+v>, received: <%+v>",
			utils.ToJSON(actPrf.ActionProfile), utils.ToJSON(result))
	}

	var actPrfIDs []string
	expactPrfIDs := []string{"actID"}

	if err := adms.GetActionProfileIDs(context.Background(), &utils.ArgsItemIDs{},
		&actPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrfIDs, expactPrfIDs) {
		t.Errorf("expected: <%+v>, received: <%+v>", expactPrfIDs, actPrfIDs)
	}

	var rplyCount int

	if err := adms.GetActionProfilesCount(context.Background(), &utils.ArgsItemIDs{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(actPrfIDs) {
		t.Errorf("expected: <%+v>, received: <%+v>", len(actPrfIDs), rplyCount)
	}

	if err := adms.RemoveActionProfile(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)
	result = engine.ActionProfile{}
	if err := adms.GetActionProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.ActionProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetActionProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.GetActionProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestActionsGetActionProfileCheckErrors",
		},
	}, &rcv); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsSetActionProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{},
	}

	var reply string
	experr := "MANDATORY_IE_MISSING: [ID Actions]"

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	actPrf.ID = "TestActionsSetActionProfileCheckErrors"
	actPrf.Actions = make([]*engine.APAction, 1)
	actPrf.FilterIDs = []string{"invalid_filter_format"}
	experr = "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TestActionsSetActionProfileCheckErrors"

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	actPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	experr = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := adms.SetActionProfile(ctx, actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>,\nreceived <%+v>", experr, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*engine.ActionProfile, error) {
			actPrf := &engine.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *engine.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsRemoveActionProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			ID:     "TestActionsRemoveActionProfileCheckErrors",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Actions: make([]*engine.APAction, 1),
		},
	}
	var reply string

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
	experr := "SERVER_ERROR: context deadline exceeded"

	if err := adms.RemoveActionProfile(ctx, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestActionsRemoveActionProfileCheckErrors",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}
	cancel()

	adms.cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	var rcv engine.ActionProfile

	if err := adms.GetActionProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestActionsRemoveActionProfileCheckErrors",
		},
	}, &rcv); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}

	experr = "MANDATORY_IE_MISSING: [ID]"

	if err := adms.RemoveActionProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.RemoveActionProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestActionsRemoveActionProfileCheckErrors",
		}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*engine.ActionProfile, error) {
			actPrf := &engine.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *engine.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, nil
		},
	}

	engine.Cache.Clear(nil)
	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveActionProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestActionsRemoveActionProfileCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*engine.ActionProfile, error) {
			actPrf := &engine.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *engine.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
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

	if err := adms.GetActionProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsErrKeys(t *testing.T) {
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

	var reply []string

	if err := adms.GetActionProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfilesCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*engine.ActionProfile, error) {
			actPrf := &engine.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *engine.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetActionProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestActionsGetActionProfilesCountErrKeys(t *testing.T) {
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

	if err := adms.GetActionProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}
}

func TestActionsGetActionProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args1 := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Actions: []*engine.APAction{
				{
					ID: "Action1",
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
	if err := admS.SetActionProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Actions: []*engine.APAction{
				{
					ID: "Action2",
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

	if err := admS.SetActionProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test2_ID1",
			Actions: []*engine.APAction{
				{
					ID: "Action1",
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

	if err := admS.SetActionProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*engine.ActionProfile{
		{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Actions: []*engine.APAction{
				{
					ID: "Action1",
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
			Actions: []*engine.APAction{
				{
					ID: "Action2",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	var getReply []*engine.ActionProfile
	if err := admS.GetActionProfiles(context.Background(), argsGet, &getReply); err != nil {
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

func TestActionsGetActionProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, nil, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Actions: []*engine.APAction{
				{
					ID: "Action1",
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
	if err := admS.SetActionProfile(context.Background(), args, &setReply); err != nil {
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
	var getReply []*engine.ActionProfile
	if err := admS.GetActionProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestActionsGetActionProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetActionProfileDrvF: func(*context.Context, *engine.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"acp_cgrates.org:TEST"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*engine.ActionProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetActionProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*engine.ActionProfile, error) {
			actionPrf := &engine.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actionPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *engine.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"acp_cgrates.org:key1", "acp_cgrates.org:key2", "acp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetActionProfileIDs(context.Background(),
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

func TestActionsGetActionProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*engine.ActionProfile, error) {
			actionPrf := &engine.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actionPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *engine.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"acp_cgrates.org:key1", "acp_cgrates.org:key2", "acp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetActionProfileIDs(context.Background(),
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

	dm.DataDB().Flush(utils.EmptyString)
}
