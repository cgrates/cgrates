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
	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestActionsSetGetRemActionProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
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
			Tenant:  "cgrates.org",
			ID:      "actID",
			Weight:  10,
			Actions: make([]*engine.APAction, 1),
		},
	}

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetActionProfile(context.Background(), arg, &result); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(result, *actPrf.ActionProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(actPrf.ActionProfile), utils.ToJSON(result))
	}

	var actPrfIDs []string
	expactPrfIDs := []string{"actID"}

	if err := adms.GetActionProfileIDs(context.Background(), &utils.PaginatorWithTenant{},
		&actPrfIDs); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(actPrfIDs, expactPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expactPrfIDs, actPrfIDs)
	}

	var rplyCount int

	if err := adms.GetActionProfileCount(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	},
		&rplyCount); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if rplyCount != len(actPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(actPrfIDs), rplyCount)
	}

	if err := adms.RemoveActionProfile(context.Background(), arg, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if err := adms.GetActionProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
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
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.GetActionProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestActionsGetActionProfileCheckErrors",
		},
	}, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsSetActionProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
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
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	actPrf.ID = "TestActionsSetActionProfileCheckErrors"
	actPrf.Actions = make([]*engine.APAction, 1)
	actPrf.FilterIDs = []string{"invalid_filter_format"}
	experr = "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TestActionsSetActionProfileCheckErrors"

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	actPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
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
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			ID:      "TestActionsRemoveActionProfileCheckErrors",
			Tenant:  "cgrates.org",
			Weight:  10,
			Actions: make([]*engine.APAction, 1),
		},
	}
	var reply string

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
	experr := "SERVER_ERROR: context deadline exceeded"

	if err := adms.RemoveActionProfile(ctx, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestActionsRemoveActionProfileCheckErrors",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	cancel()

	adms.cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	var rcv engine.ActionProfile

	if err := adms.GetActionProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestActionsRemoveActionProfileCheckErrors",
		},
	}, &rcv); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
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
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
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
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveActionProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestActionsRemoveActionProfileCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsErrMock(t *testing.T) {
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
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsErrKeys(t *testing.T) {
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
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsCountErrMock(t *testing.T) {
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

	if err := adms.GetActionProfileCount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
			},
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestActionsGetActionProfileIDsCountErrKeys(t *testing.T) {
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

	if err := adms.GetActionProfileCount(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
			},
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestActionsNewActionSv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	aS := actions.NewActionS(cfg, nil, dm, nil)

	exp := &ActionSv1{
		aS: aS,
	}
	rcv := NewActionSv1(aS)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestActionsSv1Ping(t *testing.T) {
	actSv1 := new(ActionSv1)
	var reply string
	if err := actSv1.Ping(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply error")
	}
}
