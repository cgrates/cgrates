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

func TestStatsSetGetRemStatQueueProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantID{
		ID: "sqID",
	}
	var result engine.StatQueueProfile
	var reply string

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "sqID",
			Weight: 10,
		},
	}

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetStatQueueProfile(context.Background(), arg, &result); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(result, *sqPrf.StatQueueProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(sqPrf.StatQueueProfile), utils.ToJSON(result))
	}

	var sqPrfIDs []string
	expsqPrfIDs := []string{"sqID"}

	if err := adms.GetStatQueueProfileIDs(context.Background(), &utils.PaginatorWithTenant{},
		&sqPrfIDs); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if !reflect.DeepEqual(sqPrfIDs, expsqPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expsqPrfIDs, sqPrfIDs)
	}

	var rplyCount int

	if err := adms.GetStatQueueProfileIDsCount(context.Background(), &utils.TenantWithAPIOpts{},
		&rplyCount); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	} else if rplyCount != len(sqPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(sqPrfIDs), rplyCount)
	}

	argRem := &utils.TenantIDWithAPIOpts{
		TenantID: arg,
	}

	if err := adms.RemoveStatQueueProfile(context.Background(), argRem, &reply); err != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if err := adms.GetStatQueueProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsGetStatQueueProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.StatQueueProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetStatQueueProfile(context.Background(), &utils.TenantID{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.GetStatQueueProfile(context.Background(), &utils.TenantID{
		ID: "TestStatsGetStatQueueProfileCheckErrors",
	}, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsSetStatQueueProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{},
	}

	var reply string
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	sqPrf.ID = "TestStatsSetStatQueueProfileCheckErrors"
	sqPrf.FilterIDs = []string{"invalid_filter_format"}
	experr = "SERVER_ERROR: broken reference to filter: invalid_filter_format for item with ID: cgrates.org:TestStatsSetStatQueueProfileCheckErrors"

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	sqPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10)
	experr = "SERVER_ERROR: context deadline exceeded"
	cfg.GeneralCfg().DefaultCaching = utils.MetaRemove
	if err := adms.SetStatQueueProfile(ctx, sqPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>,\nreceived <%+v>", experr, err)
	}
	cancel()

	dbMock := &engine.DataDBMock{
		GetStatQueueProfileDrvF: func(*context.Context, string, string) (*engine.StatQueueProfile, error) {
			sqPrf := &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return sqPrf, nil
		},
		SetStatQueueProfileDrvF: func(*context.Context, *engine.StatQueueProfile) error {
			return nil
		},
		RemStatQueueProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return nil, nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsRemoveStatQueueProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			ID:     "TestStatsRemoveStatQueueProfileCheckErrors",
			Tenant: "cgrates.org",
			Weight: 10,
		},
	}
	var reply string

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches): make(chan birpc.ClientConnector),
	})
	experr := "SERVER_ERROR: context deadline exceeded"

	if err := adms.RemoveStatQueueProfile(ctx, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestStatsRemoveStatQueueProfileCheckErrors",
		},
	}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
	cancel()

	adms.cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	var rcv engine.StatQueueProfile

	if err := adms.GetStatQueueProfile(context.Background(), &utils.TenantID{
		ID: "TestStatsRemoveStatQueueProfileCheckErrors",
	}, &rcv); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	experr = "MANDATORY_IE_MISSING: [ID]"

	if err := adms.RemoveStatQueueProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	if err := adms.RemoveStatQueueProfile(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "TestStatsRemoveStatQueueProfileCheckErrors",
		}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dbMock := &engine.DataDBMock{
		GetStatQueueProfileDrvF: func(*context.Context, string, string) (*engine.StatQueueProfile, error) {
			sqPrf := &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return sqPrf, nil
		},
		SetStatQueueProfileDrvF: func(*context.Context, *engine.StatQueueProfile) error {
			return nil
		},
		RemStatQueueProfileDrvF: func(*context.Context, string, string) error {
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

	if err := adms.RemoveStatQueueProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestStatsRemoveStatQueueProfileCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsGetStatQueueProfileIDsErrMock(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetStatQueueProfileDrvF: func(*context.Context, string, string) (*engine.StatQueueProfile, error) {
			sqPrf := &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return sqPrf, nil
		},
		SetStatQueueProfileDrvF: func(*context.Context, *engine.StatQueueProfile) error {
			return nil
		},
		RemStatQueueProfileDrvF: func(*context.Context, string, string) error {
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

	if err := adms.GetStatQueueProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsGetStatQueueProfileIDsErrKeys(t *testing.T) {
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

	if err := adms.GetStatQueueProfileIDs(context.Background(),
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsGetStatQueueProfileIDsCountErrMock(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetStatQueueProfileDrvF: func(*context.Context, string, string) (*engine.StatQueueProfile, error) {
			sqPrf := &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return sqPrf, nil
		},
		SetStatQueueProfileDrvF: func(*context.Context, *engine.StatQueueProfile) error {
			return nil
		},
		RemStatQueueProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg.CacheCfg(), nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetStatQueueProfileIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestStatsGetStatQueueProfileIDsCountErrKeys(t *testing.T) {
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

	if err := adms.GetStatQueueProfileIDsCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatsNewStatsv1(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	sS := engine.NewStatService(dm, cfg, nil, nil)

	exp := &StatSv1{
		sS: sS,
	}
	rcv := NewStatSv1(sS)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestStatsSv1Ping(t *testing.T) {
	statSv1 := new(StatSv1)
	var reply string
	if err := statSv1.Ping(nil, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply error")
	}
}
