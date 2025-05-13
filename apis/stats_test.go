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
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestStatsSetGetRemStatQueueProfile(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "sqID",
		},
	}
	var result engine.StatQueueProfile
	var reply string

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "sqID",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetStatQueueProfile(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *sqPrf.StatQueueProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(sqPrf.StatQueueProfile), utils.ToJSON(result))
	}

	var sqPrfIDs []string
	expsqPrfIDs := []string{"sqID"}

	if err := adms.GetStatQueueProfileIDs(context.Background(), &utils.ArgsItemIDs{},
		&sqPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sqPrfIDs, expsqPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expsqPrfIDs, sqPrfIDs)
	}

	var rplyCount int

	if err := adms.GetStatQueueProfilesCount(context.Background(), &utils.ArgsItemIDs{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(sqPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(sqPrfIDs), rplyCount)
	}

	if err := adms.RemoveStatQueueProfile(context.Background(), arg, &reply); err != nil {
		t.Error(err)
	}
	engine.Cache.Clear(nil)

	if err := adms.GetStatQueueProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsGetStatQueueProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv engine.StatQueueProfile
	experr := "MANDATORY_IE_MISSING: [ID]"

	if err := adms.GetStatQueueProfile(context.Background(), &utils.TenantIDWithAPIOpts{}, &rcv); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	adms.dm = nil
	experr = "SERVER_ERROR: NO_DATABASE_CONNECTION"

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "sqID",
		},
	}

	if err := adms.GetStatQueueProfile(context.Background(), arg, &rcv); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsSetStatQueueProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
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
	experr = "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TestStatsSetStatQueueProfileCheckErrors"

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	sqPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
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

	engine.Cache.Clear(nil)
	adms.dm = engine.NewDataManager(dbMock, cfg, nil)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsRemoveStatQueueProfileCheckErrors(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			ID:     "TestStatsRemoveStatQueueProfileCheckErrors",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	var reply string

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
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

	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "sqID",
		},
	}

	if err := adms.GetStatQueueProfile(context.Background(), arg, &rcv); err == nil || err != utils.ErrNotFound {
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
		RemStatQueueDrvF: func(ctx *context.Context, tenant, id string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, idxKey, transactionID string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, nil
		},
	}

	adms.dm = engine.NewDataManager(dbMock, cfg, nil)
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
	engine.Cache.Clear(nil)
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

	dm := engine.NewDataManager(dbMock, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "NOT_IMPLEMENTED"

	if err := adms.GetStatQueueProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatsGetStatQueueProfileIDsErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string

	if err := adms.GetStatQueueProfileIDs(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestStatQueuesGetStatQueueProfileIDsGetOptsErr(t *testing.T) {
	engine.Cache.Clear(nil)
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
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"thp_cgrates.org:key1", "thp_cgrates.org:key2", "thp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := "cannot convert field<bool>: true to int"

	if err := adms.GetStatQueueProfileIDs(context.Background(),
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

func TestStatQueuesGetStatQueueProfileIDsPaginateErr(t *testing.T) {
	engine.Cache.Clear(nil)
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
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"dpp_cgrates.org:key1", "dpp_cgrates.org:key2", "dpp_cgrates.org:key3"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []string
	experr := `SERVER_ERROR: maximum number of items exceeded`

	if err := adms.GetStatQueueProfileIDs(context.Background(),
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

func TestStatsGetStatQueueProfilesCountErrMock(t *testing.T) {
	engine.Cache.Clear(nil)
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
	dm := engine.NewDataManager(dbMock, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetStatQueueProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestStatsGetStatQueueProfilesCountErrKeys(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{}, nil
		},
	}
	dm := engine.NewDataManager(dbMock, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply int

	if err := adms.GetStatQueueProfilesCount(context.Background(),
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestStatsNewStatsv1(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
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

func TestStatsAPIs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.FilterSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
	cfg.FilterSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.FilterSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.StatSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)

	expThEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.MetaACD:   utils.NewDecimalFromStringIgnoreError("3E+3"),
			utils.MetaASR:   utils.NewDecimal(0, 0),
			utils.MetaTCD:   utils.NewDecimal(3000, 0),
			utils.EventType: utils.StatUpdate,
			utils.StatID:    "sq2",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                3000,
			utils.MetaEventType:            utils.StatUpdate,
			utils.OptsThresholdsProfileIDs: []string{"thdID"},
			utils.OptsStatsProfileIDs:      []string{"sq1", "sq2"},
		},
	}
	mCC := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply any) error {
				expThEv.ID = args.(*utils.CGREvent).ID
				if !reflect.DeepEqual(args.(*utils.CGREvent), expThEv) {
					t.Errorf("expected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(expThEv), utils.ToJSON(args))
					return fmt.Errorf("expected: <%+v>, \nreceived: <%+v>",
						utils.ToJSON(expThEv), utils.ToJSON(args))
				}
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- mCC
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)

	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	adms := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}
	sS := engine.NewStatService(dm, cfg, fltrs, cM)
	stV1 := NewStatSv1(sS)
	var reply string

	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "actPrfID",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Actions: []*utils.APAction{
				{
					ID: "actID",
				},
			},
		},
	}

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "thdID",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			MaxHits:   10,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			ActionProfileIDs: []string{"actPrfID"},
		},
	}

	if err := adms.SetThresholdProfile(context.Background(), thPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	sqPrf1 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "sq1",
			FilterIDs:   []string{"*string:~*req.Account:1001"},
			QueueLength: 100,
			TTL:         10 * time.Second,
			MinItems:    0,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			Blockers:     utils.DynamicBlockers{{Blocker: true}},
			ThresholdIDs: []string{utils.MetaNone},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}

	sqPrf2 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "sq2",
			FilterIDs:   []string{"*string:~*req.Account:1002"},
			QueueLength: 100,
			TTL:         1 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaASR,
				},
			},
			Blockers:     utils.DynamicBlockers{{Blocker: true}},
			ThresholdIDs: []string{"thdID"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}

	if err := adms.SetStatQueueProfile(context.Background(), sqPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}

	expIDs := []string{"sq1", "sq2"}
	var qIDs []string
	if err := stV1.GetQueueIDs(context.Background(), &utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
	}, &qIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(qIDs)
		if !reflect.DeepEqual(qIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, qIDs)
		}
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "StatsEventTest",
		Event: map[string]any{
			utils.AccountField: "1002",
			//utils.Usage:        3000,
		},
		APIOpts: map[string]any{
			utils.MetaUsage:           3000,
			utils.OptsStatsProfileIDs: []string{"sq1", "sq2"},
		},
	}

	expIDs = []string{"sq2"}
	if err := stV1.ProcessEvent(context.Background(), args, &qIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(qIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, qIDs)
	}

	expIDs = []string{"sq2"}
	if err := stV1.GetStatQueuesForEvent(context.Background(), args, &qIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(qIDs, expIDs) {
		t.Errorf("expected: <%+v>, received: <%+v>", expIDs, qIDs)
	}

	expStatQueue := &engine.StatQueue{
		Tenant: "cgrates.org",
		ID:     "sq1",
		SQMetrics: map[string]engine.StatMetric{
			utils.MetaACD: engine.NewACD(0, "", nil),
			utils.MetaTCD: engine.NewTCD(0, "", nil),
		},
	}

	var rplyStatQueue engine.StatQueue
	if err := stV1.GetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "sq1",
		},
	}, &rplyStatQueue); err != nil {
		t.Error(err)
	} else {
		// We compare JSONs because the received StatQueue has unexported fields
		if utils.ToJSON(rplyStatQueue) != utils.ToJSON(expStatQueue) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expStatQueue), utils.ToJSON(rplyStatQueue))
		}
	}

	expStrMetrics := map[string]string{
		utils.MetaACD: "3µs",
		utils.MetaASR: "0%",
		utils.MetaTCD: "3µs",
	}
	rplyStrMetrics := make(map[string]string)
	if err := stV1.GetQueueStringMetrics(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "sq2",
		},
	}, &rplyStrMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyStrMetrics, expStrMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expStrMetrics, rplyStrMetrics)
	}

	expFloatMetrics := map[string]float64{
		utils.MetaACD: 3000,
		utils.MetaASR: 0,
		utils.MetaTCD: 3000,
	}
	rplyFloatMetrics := make(map[string]float64)
	if err := stV1.GetQueueFloatMetrics(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "sq2",
		},
	}, &rplyFloatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFloatMetrics, expFloatMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expFloatMetrics, rplyFloatMetrics)
	}

	if err := stV1.ResetStatQueue(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "sq2",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}
}

func TestStatQueuesGetStatQueueProfilesOK(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args1 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:       "cgrates.org",
			ID:           "test_ID1",
			QueueLength:  10,
			MinItems:     2,
			ThresholdIDs: []string{utils.MetaNone},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
			},
			Stored: true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetStatQueueProfile(context.Background(), args1, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	args2 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:       "cgrates.org",
			ID:           "test_ID2",
			QueueLength:  15,
			MinItems:     3,
			ThresholdIDs: []string{utils.MetaNone},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
			Stored: false,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetStatQueueProfile(context.Background(), args2, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	// this profile will not match
	args3 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:       "cgrates.org",
			ID:           "test2_ID1",
			QueueLength:  10,
			MinItems:     2,
			ThresholdIDs: []string{utils.MetaNone},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			Stored: false,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	if err := admS.SetStatQueueProfile(context.Background(), args3, &setReply); err != nil {
		t.Error(err)
	} else if setReply != "OK" {
		t.Error("Unexpected reply returned:", setReply)
	}

	argsGet := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "test_ID",
	}
	exp := []*engine.StatQueueProfile{
		{
			Tenant:       "cgrates.org",
			ID:           "test_ID1",
			QueueLength:  10,
			MinItems:     2,
			ThresholdIDs: []string{utils.MetaNone},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
			},
			Stored: true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		{
			Tenant:       "cgrates.org",
			ID:           "test_ID2",
			QueueLength:  15,
			MinItems:     3,
			ThresholdIDs: []string{utils.MetaNone},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
			Stored: false,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	var getReply []*engine.StatQueueProfile
	if err := admS.GetStatQueueProfiles(context.Background(), argsGet, &getReply); err != nil {
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

func TestStatQueuesGetStatQueueProfilesGetIDsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	connMgr := engine.NewConnManager(cfg)
	dataDB, _ := engine.NewInternalDB(nil, nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, connMgr)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, nil)
	args := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:       "cgrates.org",
			ID:           "test_ID1",
			QueueLength:  10,
			MinItems:     2,
			ThresholdIDs: []string{utils.MetaNone},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
			},
			Stored: true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}

	var setReply string
	if err := admS.SetStatQueueProfile(context.Background(), args, &setReply); err != nil {
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
	var getReply []*engine.StatQueueProfile
	if err := admS.GetStatQueueProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestStatQueuesGetStatQueueProfilesGetProfileErr(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dbMock := &engine.DataDBMock{
		SetStatQueueProfileDrvF: func(*context.Context, *engine.StatQueueProfile) error {
			return nil
		},
		RemStatQueueProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string) ([]string, error) {
			return []string{"thp_cgrates.org:TEST"}, nil
		},
	}

	dm := engine.NewDataManager(dbMock, cfg, nil)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*engine.StatQueueProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetStatQueueProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsPrefix: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DataDB().Flush(utils.EmptyString)
}
