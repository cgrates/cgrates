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
			ID: "sqID",
		},
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

	if err := adms.GetStatQueueProfileIDs(context.Background(), &utils.PaginatorWithTenant{},
		&sqPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sqPrfIDs, expsqPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expsqPrfIDs, sqPrfIDs)
	}

	var rplyCount int

	if err := adms.GetStatQueueProfileCount(context.Background(), &utils.TenantWithAPIOpts{},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(sqPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(sqPrfIDs), rplyCount)
	}

	if err := adms.RemoveStatQueueProfile(context.Background(), arg, &reply); err != nil {
		t.Error(err)
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

func TestStatsGetStatQueueProfileCountErrMock(t *testing.T) {
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

	if err := adms.GetStatQueueProfileCount(context.Background(),
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotImplemented, err)
	}
}

func TestStatsGetStatQueueProfileCountErrKeys(t *testing.T) {
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

	if err := adms.GetStatQueueProfileCount(context.Background(),
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

func TestStatsAPIs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.FilterSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS)}
	cfg.FilterSCfg().ResourceSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources)}
	cfg.FilterSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	cfg.StatSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	data := engine.NewInternalDB(nil, nil, true)

	expThEv := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.MetaACD:   time.Duration(0),
				utils.MetaASR:   float64(0),
				utils.MetaTCD:   time.Duration(0),
				utils.EventType: utils.StatUpdate,
				utils.StatID:    "sq2",
			},
			APIOpts: map[string]interface{}{
				utils.MetaEventType:              utils.StatUpdate,
				utils.OptsThresholdsThresholdIDs: []string{"thdID"},
				utils.OptsStatsStatIDs:           []string{"sq1", "sq2"},
			},
		},
	}
	mCC := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				expThEv.ID = args.(*engine.ThresholdsArgsProcessEvent).ID
				if !reflect.DeepEqual(args.(*engine.ThresholdsArgsProcessEvent), expThEv) {
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

	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	adms := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}
	sS := engine.NewStatService(dm, cfg, fltrs, cM)
	stV1 := NewStatSv1(sS)
	var reply string

	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "actPrfID",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Actions: []*engine.APAction{
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
			Tenant:           "cgrates.org",
			ID:               "thdID",
			FilterIDs:        []string{"*string:~*req.Account:1002"},
			MaxHits:          10,
			Weight:           10,
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
			Blocker:      true,
			ThresholdIDs: []string{utils.MetaNone},
			Weight:       20,
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
			Blocker:      true,
			ThresholdIDs: []string{"thdID"},
			Weight:       20,
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

	args := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "StatsEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.Usage:        3000,
			},
			APIOpts: map[string]interface{}{
				utils.OptsStatsStatIDs: []string{"sq1", "sq2"},
			},
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
			utils.MetaACD: &engine.StatACD{
				Events: make(map[string]*engine.DurationWithCompress),
			},
			utils.MetaTCD: &engine.StatTCD{
				Events: make(map[string]*engine.DurationWithCompress),
			},
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
		utils.MetaACD: "0s",
		utils.MetaASR: "0%",
		utils.MetaTCD: "0s",
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
		utils.MetaACD: 0,
		utils.MetaASR: 0,
		utils.MetaTCD: 0,
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
