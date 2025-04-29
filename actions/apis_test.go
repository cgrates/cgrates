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

package actions

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type mockClientConn struct {
	calls map[string]func(*context.Context, any, any) error
}

func (mCC *mockClientConn) Call(ctx *context.Context, serviceMethod string, args any, reply any) (err error) {
	if call, has := mCC.calls[serviceMethod]; has {
		return call(ctx, args, reply)
	}
	return utils.ErrUnsupporteServiceMethod
}
func TestActionsAPIs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)
	aS := NewActionS(cfg, fltrs, dm, nil)
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "actPrfID",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Actions: []*utils.APAction{
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

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		ID: "EventTest",
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ScheduleActions(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}

	if err := aS.V1ExecuteActions(context.Background(), ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.OK, reply)
	}
}

func TestActionsExecuteActionsResetTH(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ID",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}
	var executed bool

	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.ThresholdSv1ResetThreshold: func(ctx *context.Context, args, reply any) error {
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
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)

	aS := NewActionS(cfg, fltrs, dm, cM)

	// Set ActionProfile
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
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
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ExecuteActions(context.Background(), ev,
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
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "SQ_ID",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.StatSv1ResetStatQueue: func(ctx *context.Context, args, reply any) error {
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
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)

	aS := NewActionS(cfg, fltrs, dm, cM)

	// Set ActionProfile
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
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
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ExecuteActions(context.Background(), ev,
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

func TestActionsExecuteActionsSetBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.ArgsActSetBalance{
		Tenant:    "cgrates.org",
		AccountID: "ACC_ID",
		Diktats:   []*utils.BalDiktat{},
		Reset:     true,
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AccountSv1ActionSetBalance: func(ctx *context.Context, args, reply any) error {
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
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)

	aS := NewActionS(cfg, fltrs, dm, cM)

	// Set ActionProfile
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
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
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ExecuteActions(context.Background(), ev,
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
func TestActionsExecuteActionsAddBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.ArgsActSetBalance{
		Tenant:    "cgrates.org",
		AccountID: "ACC_ID",
		Diktats:   []*utils.BalDiktat{},
		Reset:     false,
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AccountSv1ActionSetBalance: func(ctx *context.Context, args, reply any) error {
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
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)

	aS := NewActionS(cfg, fltrs, dm, cM)

	// Set ActionProfile
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
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
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ExecuteActions(context.Background(), ev,
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

func TestActionsExecuteActionsLog(t *testing.T) {
	engine.Cache.Clear(nil)
	tmpLogger := utils.Logger
	defer func() {
		utils.Logger = tmpLogger
	}()
	var buf bytes.Buffer
	utils.Logger = utils.NewStdLoggerWithWriter(&buf, "", 6)

	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)

	aS := NewActionS(cfg, fltrs, dm, nil)

	// Set ActionProfile
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
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
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ExecuteActions(context.Background(), ev,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	// Check if the log action was executed
	expected := `{"*opts":{"*actProfileIDs":["actPrfID"]},"*req":{"Account":"1001"}}`
	if rcv := buf.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected log: %s to be included in %s", expected, rcv)
	}

	dm.DataDB().Flush(utils.EmptyString)
}

func TestActionsExecuteActionsLogCDRs(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.CDRs)}
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	// expArgs := &utils.CGREvent{
	// 	Flags: []string{utils.ConcatenatedKey(utils.MetaChargers, utils.FalseStr)}, // do not try to get the chargers for cdrlog
	// 	CGREvent: *utils.NMAsCGREvent(utils.NewOrderedNavigableMap(), "cgrates.org",
	// 		utils.NestingSep, utils.MapStorage{}),
	// }
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.CDRsV1ProcessEvent: func(ctx *context.Context, args, reply any) error {
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
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)

	aS := NewActionS(cfg, fltrs, dm, cM)

	// Set ActionProfile
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
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
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Tenant:       "cgrates.org",
			utils.BalanceType:  utils.MetaConcrete,
			utils.Cost:         0.15,
			utils.ActionType:   utils.MetaTopUp,
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ExecuteActions(context.Background(), ev,
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

func TestActionsExecuteActionsRemBalance(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.GeneralCfg().DefaultCaching = utils.MetaNone
	cfg.ActionSCfg().AccountSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts)}
	dataDB := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(dataDB, cfg, nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)

	expArgs := &utils.ArgsActRemoveBalances{
		Tenant:     "cgrates.org",
		AccountID:  "ACC_ID",
		BalanceIDs: []string{},
	}
	var executed bool
	cc := &mockClientConn{
		calls: map[string]func(ctx *context.Context, args any, reply any) error{
			utils.AccountSv1ActionRemoveBalance: func(ctx *context.Context, args, reply any) error {
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
	adms := apis.NewAdminSv1(cfg, dm, nil, nil, nil)

	aS := NewActionS(cfg, fltrs, dm, cM)

	// Set ActionProfile
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
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
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	if err := aS.V1ExecuteActions(context.Background(), ev,
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
