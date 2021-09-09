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
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
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
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("\nexpected: <%+v>, received: <%+v>", utils.OK, reply)
	}

	if err := adms.GetActionProfile(context.Background(), arg, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, *actPrf.ActionProfile) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(actPrf.ActionProfile), utils.ToJSON(result))
	}

	var actPrfIDs []string
	expactPrfIDs := []string{"actID"}

	if err := adms.GetActionProfileIDs(context.Background(), &utils.PaginatorWithTenant{},
		&actPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrfIDs, expactPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expactPrfIDs, actPrfIDs)
	}

	var rplyCount int

	if err := adms.GetActionProfileCount(context.Background(), &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{},
	},
		&rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != len(actPrfIDs) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", len(actPrfIDs), rplyCount)
	}

	if err := adms.RemoveActionProfile(context.Background(), arg, &reply); err != nil {
		t.Error(err)
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
	adms.connMgr = engine.NewConnManager(cfg)
	adms.connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, make(chan birpc.ClientConnector))
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

func TestActionsGetActionProfileCountErrMock(t *testing.T) {
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

func TestActionsGetActionProfileCountErrKeys(t *testing.T) {
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

func TestActionsAPIs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	adms := &AdminSv1{
		dm:  dm,
		cfg: cfg,
	}
	aS := actions.NewActionS(cfg, fltrs, dm, nil)
	aSv1 := NewActionSv1(aS)

	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "actPrfID",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Actions: []*engine.APAction{
				{
					ID: "actID",
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf, &reply); err != nil {
		t.Error(err)
	}

	args := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
			ID: "EventTest",
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ScheduleActions(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}

	if err := aSv1.ExecuteActions(context.Background(), args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}
}

func TestActionsExecuteActionsResetTH(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ID",
		},
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.ThresholdSv1ResetThreshold: func(ctx *context.Context, args, reply interface{}) error {
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>",
						utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				executed = true
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, rpcInternal)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	aS := actions.NewActionS(cfg, fltrs, dm, cM)
	aSv1 := NewActionSv1(aS)

	// Set ActionProfile
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaResetThreshold,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaThresholds: {
					"THD_ID": struct{}{},
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// ExecuteActions with ResetThreshold
	argsAct := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ExecuteActions(context.Background(), argsAct,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	if !executed {
		t.Errorf("ResetThreshold hasn't been executed")
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsExecuteActionsResetSQ(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "SQ_ID",
		},
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.StatSv1ResetStatQueue: func(ctx *context.Context, args, reply interface{}) error {
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>",
						utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				executed = true
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, rpcInternal)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	aS := actions.NewActionS(cfg, fltrs, dm, cM)
	aSv1 := NewActionSv1(aS)

	// Set ActionProfile
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaResetStatQueue,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaStats: {
					"SQ_ID": struct{}{},
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// ExecuteActions with ResetStatQueue
	argsAct := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ExecuteActions(context.Background(), argsAct,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// Check if ResetStatQueue has been executed
	if !executed {
		t.Errorf("ResetStatQueue hasn't been executed")
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsExecuteActionsLog(t *testing.T) {
	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	aS := actions.NewActionS(cfg, fltrs, dm, nil)
	aSv1 := NewActionSv1(aS)

	// Set ActionProfile
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaLog,
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// ExecuteActions with Log
	argsAct := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ExecuteActions(context.Background(), argsAct,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// Check if the log action was executed
	expected := `CGRateS <> [INFO] LOG Event: {"*opts":null,"*req":{"Account":"1001"}}`
	if rcv := buf.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected log: %q", expected)
	}

	utils.Logger.SetLogLevel(0)
	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsExecuteActionsLogCDRs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.CDRs)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	// expArgs := &engine.ArgV1ProcessEvent{
	// 	Flags: []string{utils.ConcatenatedKey(utils.MetaChargers, utils.FalseStr)}, // do not try to get the chargers for cdrlog
	// 	CGREvent: *utils.NMAsCGREvent(utils.NewOrderedNavigableMap(), "cgrates.org",
	// 		utils.NestingSep, utils.MapStorage{}),
	// }
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.CDRsV1ProcessEvent: func(ctx *context.Context, args, reply interface{}) error {
				// if !reflect.DeepEqual(args, expArgs) {
				// 	return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>",
				// 		utils.ToJSON(expArgs), utils.ToJSON(args))
				// }
				executed = true
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.CDRs), utils.CDRsV1, rpcInternal)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	aS := actions.NewActionS(cfg, fltrs, dm, cM)
	aSv1 := NewActionSv1(aS)

	// Set ActionProfile
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.CDRLog,
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// ExecuteActions with CDRLog
	argsAct := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Tenant:       "cgrates.org",
				utils.BalanceType:  utils.MetaConcrete,
				utils.Cost:         0.15,
				utils.ActionType:   utils.MetaTopUp,
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ExecuteActions(context.Background(), argsAct,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// Check if CDRs ProcessEvent has been executed
	if !executed {
		t.Errorf("CDRLog hasn't been executed")
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsExecuteActionsSetBalance(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.ArgsActSetBalance{
		Tenant:    "cgrates.org",
		AccountID: "ACC_ID",
		Diktats:   []*utils.BalDiktat{},
		Reset:     true,
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1ActionSetBalance: func(ctx *context.Context, args, reply interface{}) error {
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>",
						utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				executed = true
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, rpcInternal)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	aS := actions.NewActionS(cfg, fltrs, dm, cM)
	aSv1 := NewActionSv1(aS)

	// Set ActionProfile
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaSetBalance,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaAccounts: {
					"ACC_ID": struct{}{},
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// ExecuteActions with SetBalance
	argsAct := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ExecuteActions(context.Background(), argsAct,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// Check if SetBalance has been executed
	if !executed {
		t.Errorf("SetBalance hasn't been executed")
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsExecuteActionsRemBalance(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.ArgsActRemoveBalances{
		Tenant:     "cgrates.org",
		AccountID:  "ACC_ID",
		BalanceIDs: []string{},
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1ActionRemoveBalance: func(ctx *context.Context, args, reply interface{}) error {
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>",
						utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				executed = true
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, rpcInternal)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	aS := actions.NewActionS(cfg, fltrs, dm, cM)
	aSv1 := NewActionSv1(aS)

	// Set ActionProfile
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaRemBalance,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaAccounts: {
					"ACC_ID": struct{}{},
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// ExecuteActions with RemBalance
	argsAct := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ExecuteActions(context.Background(), argsAct,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// Check if RemoveBalance has been executed
	if !executed {
		t.Errorf("RemoveBalance hasn't been executed")
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsExecuteActionsAddBalance(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	dataDB := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(dataDB, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.ArgsActSetBalance{
		Tenant:    "cgrates.org",
		AccountID: "ACC_ID",
		Diktats:   []*utils.BalDiktat{},
		Reset:     false,
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args interface{}, reply interface{}) error{
			utils.AccountSv1ActionSetBalance: func(ctx *context.Context, args, reply interface{}) error {
				if !reflect.DeepEqual(args, expArgs) {
					return fmt.Errorf("expected: <%+v>,\nreceived: <%+v>",
						utils.ToJSON(expArgs), utils.ToJSON(args))
				}
				executed = true
				return nil
			},
		},
	}
	rpcInternal := make(chan birpc.ClientConnector, 1)
	rpcInternal <- cc
	cM := engine.NewConnManager(cfg)
	cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, rpcInternal)
	adms := &AdminSv1{
		cfg: cfg,
		dm:  dm,
	}

	aS := actions.NewActionS(cfg, fltrs, dm, cM)
	aSv1 := NewActionSv1(aS)

	// Set ActionProfile
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaAddBalance,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaAccounts: {
					"ACC_ID": struct{}{},
				},
			},
		},
	}

	var reply string
	if err := adms.SetActionProfile(context.Background(), actPrf,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// ExecuteActions with AddBalance
	argsAct := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	if err := aSv1.ExecuteActions(context.Background(), argsAct,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// Check if AddBalance has been executed
	if !executed {
		t.Errorf("AddBalance hasn't been executed")
	}

	dm.DataDB().Flush(utils.EmptyString)
}
