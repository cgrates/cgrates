// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/actions"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestActionSv1GetScheduledActions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	filters := engine.NewFilterS(cfg, nil, dm)
	acts := actions.NewActionS(cfg, cacheS, filters, dm, nil)
	t.Cleanup(acts.Shutdown)
	api := NewActionSv1(acts)
	ctx := context.Background()
	profile := &utils.ActionProfile{
		Tenant:   "cgrates.org",
		ID:       "renew",
		Schedule: "@every 1h",
		Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"account1": {}, "account2": {}}},
		Actions:  []*utils.APAction{{ID: "balance", Type: utils.MetaSetBalance}},
	}
	if err := dm.SetActionProfile(ctx, profile, true); err != nil {
		t.Fatal(err)
	}
	if err := api.ScheduleActions(ctx, &utils.CGREvent{
		APIOpts: map[string]any{utils.OptsActionsProfileIDs: []string{profile.ID}},
	}, new(string)); err != nil {
		t.Fatal(err)
	}
	profile.Schedule = "@every 2h"
	if err := dm.SetActionProfile(ctx, profile, true); err != nil {
		t.Fatal(err)
	}
	var batches []utils.ScheduledActionBatch
	if err := api.GetScheduledActions(ctx, &utils.GetScheduledActionsArgs{
		ActionProfileIDs: []string{profile.ID},
		Targets:          map[string][]string{utils.MetaAccounts: {"account1"}},
	}, &batches); err != nil {
		t.Fatal(err)
	}
	if len(batches) != 1 {
		t.Fatalf("expected one registered batch, received %s", utils.ToJSON(batches))
	}
	batch := batches[0]
	if batch.Tenant != profile.Tenant || batch.ActionProfileID != profile.ID || batch.Schedule != "@every 1h" || batch.NextRun.IsZero() {
		t.Fatalf("unexpected registration: %s", utils.ToJSON(batch))
	}
	if want := map[string][]string{utils.MetaAccounts: {"account1", "account2"}}; !reflect.DeepEqual(batch.Targets, want) {
		t.Fatalf("expected targets %v, received %v", want, batch.Targets)
	}
	if err := api.GetScheduledActions(ctx, &utils.GetScheduledActionsArgs{
		Targets: map[string][]string{utils.MetaAccounts: {"missing"}},
	}, &batches); err != utils.ErrNotFound {
		t.Fatalf("expected NOT_FOUND, received %v", err)
	}
}

func TestSetActionProfileReloadScheduler(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.AdminSCfg().Conns[utils.MetaActions] = []*config.DynamicConns{{ConnIDs: []string{"*internal:*actions"}}}
	cfg.ActionSCfg().Conns[utils.MetaAccounts] = []*config.DynamicConns{{ConnIDs: []string{"*internal:*accounts"}}}
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm.SetCache(cacheS)
	filters := engine.NewFilterS(cfg, nil, dm)
	cm := engine.NewConnManager(cfg)
	cm.SetCache(cacheS)
	acts := actions.NewActionS(cfg, cacheS, filters, dm, cm)
	acts.Shutdown()
	api := NewActionSv1(acts)
	executed := make(chan struct{}, 1)
	rpc := make(chan birpc.ClientConnector, 1)
	rpc <- &mockClientConn{calls: map[string]func(*context.Context, any, any) error{
		utils.ActionSv1ReloadScheduler: func(ctx *context.Context, args, reply any) error {
			return api.ReloadScheduler(ctx, args.(*utils.TenantIDWithAPIOpts), reply.(*string))
		},
		utils.AccountSv1ActionSetBalance: func(*context.Context, any, any) error {
			executed <- struct{}{}
			return nil
		},
	}}
	cm.AddInternalConn("*internal:*actions", utils.ActionSv1, rpc)
	cm.AddInternalConn("*internal:*accounts", utils.AccountSv1, rpc)
	admin := NewAdminSv1(cfg, dm, cm, filters, locker)
	ctx := context.Background()
	var reply string
	profile := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			ID:       "renew",
			Schedule: "@monthly",
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"account1": {}}},
			Actions:  []*utils.APAction{{ID: "balance", Type: utils.MetaSetBalance}},
		},
		ReloadScheduler: true,
	}
	if err := admin.SetActionProfile(ctx, profile, &reply); err != nil {
		t.Fatal(err)
	}
	checkSchedule := func(schedule string, targets []string) {
		t.Helper()
		var batches []utils.ScheduledActionBatch
		if err := api.GetScheduledActions(ctx, &utils.GetScheduledActionsArgs{ActionProfileIDs: []string{"renew"}}, &batches); err != nil {
			t.Fatal(err)
		}
		if len(batches) != 1 || batches[0].Schedule != schedule || !reflect.DeepEqual(batches[0].Targets[utils.MetaAccounts], targets) {
			t.Fatalf("unexpected registration: %s", utils.ToJSON(batches))
		}
	}
	checkSchedule("@monthly", []string{"account1"})

	unscheduled := &utils.ActionProfileWithAPIOpts{ActionProfile: profile.ActionProfile.Clone()}
	unscheduled.ID = "unscheduled"
	if err := admin.SetActionProfile(ctx, unscheduled, &reply); err != nil {
		t.Fatal(err)
	}
	var batches []utils.ScheduledActionBatch
	if err := api.GetScheduledActions(ctx, &utils.GetScheduledActionsArgs{ActionProfileIDs: []string{unscheduled.ID}}, &batches); err != utils.ErrNotFound {
		t.Fatalf("save-only profile: expected NOT_FOUND, received %v", err)
	}
	unscheduled.ReloadScheduler = true
	if err := admin.SetActionProfile(ctx, unscheduled, &reply); err != nil {
		t.Fatal(err)
	}

	if err := admin.SetActionProfile(ctx, &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			ID:       profile.ID,
			Schedule: "@weekly",
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"account2": {}}},
		},
		ReloadScheduler: true,
	}, &reply); err != nil {
		t.Fatal(err)
	}
	checkSchedule("@weekly", []string{"account1", "account2"})

	profile.Overwrite = true
	profile.Schedule = "@yearly"
	profile.Targets[utils.MetaAccounts] = utils.StringSet{"account3": {}}
	if err := dm.Guard(ctx, func(ctx *context.Context) error {
		return admin.SetActionProfile(ctx, profile, &reply)
	}, utils.ActionProfilePrefix+profile.TenantID()); err != nil {
		t.Fatal(err)
	}
	checkSchedule("@yearly", []string{"account3"})

	update := &utils.ActionProfileWithAPIOpts{ActionProfile: &utils.ActionProfile{ID: profile.ID, Schedule: "@daily"}}
	if err := admin.SetActionProfile(ctx, update, &reply); err != nil {
		t.Fatal(err)
	}
	checkSchedule("@yearly", []string{"account3"})
	update.Schedule = "invalid"
	update.ReloadScheduler = true
	if err := admin.SetActionProfile(ctx, update, &reply); err == nil {
		t.Fatal("expected reload to reject the invalid schedule")
	}
	checkSchedule("@yearly", []string{"account3"})
	stored, err := dm.GetActionProfile(ctx, profile.Tenant, profile.ID, false, false, utils.NonTransactional)
	if err != nil || stored.Schedule != "invalid" {
		t.Fatalf("reload failure should not roll back the save: %v, %v", stored, err)
	}

	profile.Schedule = utils.MetaASAP
	if err := admin.SetActionProfile(ctx, profile, &reply); err != nil {
		t.Fatal(err)
	}
	if err := api.GetScheduledActions(ctx, &utils.GetScheduledActionsArgs{ActionProfileIDs: []string{profile.ID}}, &batches); err != utils.ErrNotFound {
		t.Fatalf("ASAP reload: expected NOT_FOUND, received %v", err)
	}
	select {
	case <-executed:
		t.Fatal("reload executed an ASAP action")
	case <-time.After(20 * time.Millisecond):
	}
	if err := api.GetScheduledActions(ctx, &utils.GetScheduledActionsArgs{ActionProfileIDs: []string{unscheduled.ID}}, &batches); err != nil {
		t.Fatalf("reload changed the other profile: %v", err)
	}
	if err := api.ScheduleActions(ctx, &utils.CGREvent{
		APIOpts: map[string]any{utils.OptsActionsProfileIDs: []string{profile.ID}},
	}, &reply); err != nil {
		t.Fatal(err)
	}
	select {
	case <-executed:
	case <-time.After(time.Second):
		t.Fatal("ScheduleActions did not execute the ASAP action")
	}
}

func TestActionsSetGetRemActionProfile(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dataDB, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	arg := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID: "actID",
		},
	}
	var result utils.ActionProfile
	var reply string

	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actID",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Actions: []*utils.APAction{{ID: "log", Type: utils.MetaLog, Diktats: []*utils.APDiktat{{ID: "d"}}}},
			Targets: map[string]utils.StringSet{
				utils.MetaAccounts: utils.NewStringSet([]string{"account1", "account2"}),
			},
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

	update := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			ID: " actID ",
			Targets: map[string]utils.StringSet{
				utils.MetaAccounts: utils.NewStringSet([]string{"account2", "account3"}),
			},
		},
	}
	if err := adms.SetActionProfile(context.Background(), update, &reply); err != nil {
		t.Fatal(err)
	}
	if err := adms.SetActionProfile(context.Background(), &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			ID:      "actID",
			Actions: []*utils.APAction{{ID: "log", Diktats: []*utils.APDiktat{{ID: " d "}}}},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}
	actPrf.Targets[utils.MetaAccounts].Add("account3")
	cacheS.Clear(nil)
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
	cacheS.Clear(nil)
	result = utils.ActionProfile{}
	if err := adms.GetActionProfile(context.Background(), arg, &result); err == nil ||
		err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, received: <%+v>", utils.ErrNotFound, err)
	}

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsGetActionProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dataDB, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}
	var rcv utils.ActionProfile
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

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsSetActionProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dataDB, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{},
	}

	var reply string
	experr := "MANDATORY_IE_MISSING: [ID Actions]"

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	actPrf.ID = "TestActionsSetActionProfileCheckErrors"
	actPrf.Actions = make([]*utils.APAction, 1)
	actPrf.FilterIDs = []string{"invalid_filter_format"}
	experr = "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TestActionsSetActionProfileCheckErrors"

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	actPrf.FilterIDs = []string{}
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.SetCache(cacheS)
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
		GetActionProfileDrvF: func(*context.Context, string, string) (*utils.ActionProfile, error) {
			actPrf := &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *utils.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return nil, nil
		},
	}

	dbCm := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	adms.dm = engine.NewDataManager(dbCm, cfg, nil, locker)
	adms.dm.SetCache(cacheS)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err == nil ||
		err.Error() != experr {
		t.Errorf("\nexpected <%+v>, \nreceived <%+v>", experr, err)
	}

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsRemoveActionProfileCheckErrors(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dataDB, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			ID:     "TestActionsRemoveActionProfileCheckErrors",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Actions: make([]*utils.APAction, 1),
		},
	}
	var reply string

	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10)
	adms.cfg.GeneralCfg().DefaultCaching = "not_a_caching_type"
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.SetCache(cacheS)
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
	var rcv utils.ActionProfile

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
		GetActionProfileDrvF: func(*context.Context, string, string) (*utils.ActionProfile, error) {
			actPrf := &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *utils.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return nil, nil
		},
		SetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx string, indexes map[string]utils.StringSet, commit bool, transactionID string) (err error) {
			return nil
		},
		GetIndexesDrvF: func(ctx *context.Context, idxItmType, tntCtx, transactionID string, idxKeys ...string) (indexes map[string]utils.StringSet, err error) {
			return map[string]utils.StringSet{}, nil
		},
	}

	cacheS.Clear(nil)
	dbCm := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	adms.dm = engine.NewDataManager(dbCm, cfg, nil, locker)
	adms.dm.SetCache(cacheS)
	experr = "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.RemoveActionProfile(context.Background(),
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				ID: "TestActionsRemoveActionProfileCheckErrors",
			}}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, received: <%+v>", experr, err)
	}

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsErrMock(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*utils.ActionProfile, error) {
			actPrf := &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *utils.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsErrKeys(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{}, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsGetActionProfilesCountErrMock(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*utils.ActionProfile, error) {
			actPrf := &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *utils.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dbMock := &engine.DataDBMock{
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{}, nil
		},
	}
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	dataDB, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, locker)
	args1 := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Actions: []*utils.APAction{
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

	args2 := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID2",
			Actions: []*utils.APAction{
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
	args3 := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test2_ID1",
			Actions: []*utils.APAction{
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
		ItemsSearch: "test_ID",
	}
	exp := []*utils.ActionProfile{
		{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Actions: []*utils.APAction{
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
			Actions: []*utils.APAction{
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

	var getReply []*utils.ActionProfile
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
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cacheS)
	dataDB, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dataDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, connMgr, locker)
	dm.SetCache(cacheS)
	admS := NewAdminSv1(cfg, dm, connMgr, nil, locker)
	args := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "test_ID1",
			Actions: []*utils.APAction{
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
		ItemsSearch: "test_ID",
		APIOpts: map[string]any{
			utils.PageLimitOpt:    2,
			utils.PageOffsetOpt:   4,
			utils.PageMaxItemsOpt: 5,
		},
	}

	experr := `SERVER_ERROR: maximum number of items exceeded`
	var getReply []*utils.ActionProfile
	if err := admS.GetActionProfiles(context.Background(), argsGet, &getReply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestActionsGetActionProfilesGetProfileErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dbMock := &engine.DataDBMock{
		SetActionProfileDrvF: func(*context.Context, *utils.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"acp_cgrates.org:TEST"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	var reply []*utils.ActionProfile
	experr := "SERVER_ERROR: NOT_IMPLEMENTED"

	if err := adms.GetActionProfiles(context.Background(),
		&utils.ArgsItemIDs{
			ItemsSearch: "TEST",
		}, &reply); err == nil || err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsGetOptsErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*utils.ActionProfile, error) {
			actionPrf := &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actionPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *utils.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"acp_cgrates.org:key1", "acp_cgrates.org:key2", "acp_cgrates.org:key3"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}

func TestActionsGetActionProfileIDsPaginateErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dbMock := &engine.DataDBMock{
		GetActionProfileDrvF: func(*context.Context, string, string) (*utils.ActionProfile, error) {
			actionPrf := &utils.ActionProfile{
				Tenant: "cgrates.org",
				ID:     "TEST",
			}
			return actionPrf, nil
		},
		SetActionProfileDrvF: func(*context.Context, *utils.ActionProfile) error {
			return nil
		},
		RemoveActionProfileDrvF: func(*context.Context, string, string) error {
			return nil
		},
		GetKeysForPrefixF: func(c *context.Context, s string, srch string) ([]string, error) {
			return []string{"acp_cgrates.org:key1", "acp_cgrates.org:key2", "acp_cgrates.org:key3"}, nil
		},
	}

	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: dbMock}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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

	dm.DB()[utils.MetaDefault].Flush(utils.EmptyString)
}
