/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cron"

	"github.com/cgrates/rpcclient"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestMatchingActionProfilesForEvent(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)

	evNM := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  1002,
		},
		utils.MetaOpts: map[string]interface{}{},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Diktats: []*engine.APDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}

	if err := acts.dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}

	expActionPrf := engine.ActionProfiles{actPrf}

	if rcv, err := acts.matchingActionProfilesForEvent(context.Background(), "cgrates.org",
		evNM, []string{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expActionPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expActionPrf), utils.ToJSON(rcv))
	}

	evNM = utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "10",
		},
		utils.MetaOpts: map[string]interface{}{},
	}
	//This Event is not matching with our filter
	if _, err := acts.matchingActionProfilesForEvent(context.Background(), "cgrates.org",
		evNM, []string{}); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	evNM = utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "1001",
		},
		utils.MetaOpts: map[string]interface{}{},
	}
	actPrfIDs := []string{"inexisting_id"}
	//Unable to get from database an ActionProfile if the ID won't match
	if _, err := acts.matchingActionProfilesForEvent(context.Background(), "cgrates.org",
		evNM, actPrfIDs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	actPrfIDs = []string{"test_id1"}
	if _, err := acts.matchingActionProfilesForEvent(context.Background(), "cgrates.org",
		evNM, actPrfIDs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
	actPrf.FilterIDs = append(actPrf.FilterIDs, "*ai:~*req.AnswerTime:2012-07-21T00:00:00Z|2012-08-21T00:00:00Z")
	//this event is not active in this interval time
	if _, err := acts.matchingActionProfilesForEvent(context.Background(), "cgrates.org",
		evNM, actPrfIDs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//when dataManager is nil, it won't be able to get ActionsProfile from database
	acts.dm = nil
	if _, err := acts.matchingActionProfilesForEvent(context.Background(), "INVALID_TENANT",
		evNM, actPrfIDs); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}

	acts.dm = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	actPrf.FilterIDs = []string{"invalid_filters"}
	//Set in database and invalid filter, so it won t pass
	if err := acts.dm.SetActionProfile(context.Background(), actPrf, false); err != nil {
		t.Error(err)
	}
	expected := "NOT_FOUND:invalid_filters"
	if _, err := acts.matchingActionProfilesForEvent(context.Background(), "cgrates.org",
		evNM, actPrfIDs); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if err := acts.dm.RemoveActionProfile(context.Background(), actPrf.Tenant,
		actPrf.ID, utils.NonTransactional, false); err != nil {
		t.Error(err)
	}
}

func TestScheduledActions(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ACTIONS1",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.Destination:  1002,
		},
	}
	evNM := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: map[string]interface{}{},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Diktats: []*engine.APDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}

	if err := acts.dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}

	if rcv, err := acts.scheduledActions(context.Background(), cgrEv.Tenant, cgrEv, []string{}, false); err != nil {
		t.Error(err)
	} else {
		expSchedActs := newScheduledActs(context.Background(), cgrEv.Tenant, cgrEv.ID, utils.MetaNone, utils.EmptyString,
			utils.EmptyString, evNM, rcv[0].acts)
		if reflect.DeepEqual(expSchedActs, rcv) {
			t.Errorf("Expected %+v, received %+v", expSchedActs, rcv)
		}
	}

	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TEST_ACTIONS1",
		Event: map[string]interface{}{
			utils.Accounts: "10",
		},
	}
	if _, err := acts.scheduledActions(context.Background(), cgrEv.Tenant, cgrEv, []string{}, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestScheduleAction(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)

	cgrEv := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACTIONS1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  1002,
			},
		},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Schedule:  "* * * * *",
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Diktats: []*engine.APDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}
	if err := acts.dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}

	if err := acts.scheduleActions(context.Background(), cgrEv, []string{}, true); err != nil {
		t.Error(err)
	}

	//Cannot schedule an action if the ID is invalid
	if err := acts.scheduleActions(context.Background(), cgrEv, []string{"INVALID_ID1"}, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

	//When schedule is "*asap", the action will execute immediately
	actPrf.Schedule = utils.MetaASAP
	if err := acts.dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}
	if err := acts.scheduleActions(context.Background(), cgrEv, []string{}, true); err != nil {
		t.Error(err)
	}

	//Cannot execute the action if the cron is invalid
	actPrf.Schedule = "* * * *"
	if err := acts.dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}
	if err := acts.scheduleActions(context.Background(), cgrEv, []string{}, true); err == nil || err != utils.ErrPartiallyExecuted {
		t.Error(err)
	}
}

func TestAsapExecuteActions(t *testing.T) {
	newData := &dataDBMockError{}
	dm := engine.NewDataManager(newData, config.CgrConfig().CacheCfg(), nil)
	defaultCfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)

	cgrEv := []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "CHANGED_ID",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  1002,
			},
		},
	}

	evNM := utils.MapStorage{
		utils.MetaReq:  cgrEv[0].Event,
		utils.MetaOpts: map[string]interface{}{},
	}

	expSchedActs := newScheduledActs(context.Background(), cgrEv[0].Tenant, cgrEv[0].ID, utils.MetaNone, utils.EmptyString,
		utils.EmptyString, evNM, nil)

	if err := acts.asapExecuteActions(context.Background(), expSchedActs); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("Expected %+v, received %+v", utils.ErrNoDatabaseConn, err)
	}

	data := engine.NewInternalDB(nil, nil, true)
	acts.dm = engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	expSchedActs = newScheduledActs(context.Background(), cgrEv[0].Tenant, "another_id", utils.MetaNone, utils.EmptyString,
		utils.EmptyString, evNM, nil)
	if err := acts.asapExecuteActions(context.Background(), expSchedActs); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestActionSListenAndServe(t *testing.T) {
	newData := &dataDBMockError{}
	dm := engine.NewDataManager(newData, config.CgrConfig().CacheCfg(), nil)
	defaultCfg := config.NewDefaultCGRConfig()
	defaultCfg.ActionSCfg().Tenants = &[]string{"cgrates1.org", "cgrates.org2"}
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)

	stopChan := make(chan struct{}, 1)
	cfgRld := make(chan struct{}, 1)
	cfgRld <- struct{}{}
	go func() {
		time.Sleep(10)
		stopChan <- struct{}{}
	}()
	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)
	buff := new(bytes.Buffer)
	log.SetOutput(buff)
	acts.ListenAndServe(stopChan, cfgRld)
	expString := "CGRateS <> [INFO] <CoreS> starting <ActionS>"
	if rcv := buff.String(); !strings.Contains(rcv, expString) {
		t.Errorf("Expected %+v, received %+v", expString, rcv)
	}
	buff.Reset()
}

func TestV1ScheduleActions(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)

	var reply string
	newArgs := &utils.ArgActionSv1ScheduleActions{
		ActionProfileIDs: []string{},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test_id1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  1002,
			},
		},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Schedule:  utils.MetaASAP,
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Diktats: []*engine.APDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}

	if err := acts.dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}

	if err := acts.V1ScheduleActions(context.Background(), newArgs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %+v", reply)
	}

	newArgs.ActionProfileIDs = []string{"invalid_id"}
	if err := acts.V1ScheduleActions(context.Background(), newArgs, &reply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

	if err := acts.dm.RemoveActionProfile(context.Background(), actPrf.Tenant, actPrf.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestV1ExecuteActions(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)

	var reply string
	newArgs := &utils.ArgActionSv1ScheduleActions{
		ActionProfileIDs: []string{},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "test_id1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  1002,
			},
		},
	}

	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Schedule:  utils.MetaASAP,
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Diktats: []*engine.APDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}
	if err := acts.dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}

	if err := acts.V1ExecuteActions(context.Background(), newArgs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %+v", reply)
	}

	newArgs.ActionProfileIDs = []string{"invalid_id"}
	if err := acts.V1ExecuteActions(context.Background(), newArgs, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	newData := &dataDBMockError{}
	newDm := engine.NewDataManager(newData, config.CgrConfig().CacheCfg(), nil)
	newActs := NewActionS(defaultCfg, filters, newDm, nil)
	newArgs.ActionProfileIDs = []string{}
	if err := newActs.V1ExecuteActions(context.Background(), newArgs, &reply); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

	if err := acts.dm.RemoveActionProfile(context.Background(), actPrf.Tenant, actPrf.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestActionShutDown(t *testing.T) {
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	defaultCfg := config.NewDefaultCGRConfig()
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	acts := NewActionS(defaultCfg, filters, dm, nil)
	acts.crn = &cron.Cron{}

	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)
	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	acts.Shutdown()
	expBuff := "CGRateS <> [INFO] <CoreS> shutdown <ActionS>"
	if rcv := buff.String(); !strings.Contains(rcv, expBuff) {
		t.Errorf("Expected %+v, received %+v", expBuff, rcv)
	}

	buff.Reset()
}

type dataDBMockError struct {
	*engine.DataDBMock
}

func (dbM *dataDBMockError) GetActionProfileDrv(*context.Context, string, string) (*engine.ActionProfile, error) {
	return &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "test_id1",
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      utils.MetaLog,
				Diktats: []*engine.APDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}, nil
}

func (dbM *dataDBMockError) SetActionProfileDrv(*context.Context, *engine.ActionProfile) error {
	return utils.ErrNoDatabaseConn
}

func TestLogActionExecute(t *testing.T) {
	evNM := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "10",
		},
		utils.MetaOpts: map[string]interface{}{},
	}

	if newLogger, err := utils.Newlogger(utils.MetaStdLog, "Engine1"); err != nil {
		t.Error(err)
	} else {
		newLogger.SetLogLevel(7)
		utils.Logger = newLogger
	}

	output := new(bytes.Buffer)
	log.SetOutput(output)

	logAction := actLog{}
	if err := logAction.execute(context.Background(), evNM, utils.MetaNone); err != nil {
		t.Error(err)
	}

	expected := "CGRateS <Engine1> [INFO] LOG Event: {\"*opts\":{},\"*req\":{\"Account\":\"10\"}}"
	if rcv := output.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
	output.Reset()

	log.SetOutput(os.Stderr)
}

type testMockCDRsConn struct {
	calls map[string]func(_ *context.Context, _, _ interface{}) error
}

func (s *testMockCDRsConn) Call(ctx *context.Context, method string, arg, rply interface{}) error {
	call, has := s.calls[method]
	if !has {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	return call(ctx, arg, rply)
}

func TestCDRLogActionExecute(t *testing.T) {
	sMock := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.CDRsV1ProcessEvent: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*engine.ArgV1ProcessEvent)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if !reflect.DeepEqual(argConv.Flags, []string{utils.ConcatenatedKey(utils.MetaChargers, "false")}) {
					return fmt.Errorf("Expected %+v, received %+v", []string{utils.ConcatenatedKey(utils.MetaChargers, "false")}, argConv.Flags)
				}
				if val, has := argConv.CGREvent.Event[utils.Subject]; !has {
					return fmt.Errorf("missing Subject")
				} else if strVal := utils.IfaceAsString(val); strVal != "10" {
					return fmt.Errorf("Expected %+v, received %+v", "10", strVal)
				}
				if val, has := argConv.CGREvent.Event[utils.Cost]; !has {
					return fmt.Errorf("missing Cost")
				} else if strVal := utils.IfaceAsString(val); strVal != "0.15" {
					return fmt.Errorf("Expected %+v, received %+v", "0.15", strVal)
				}
				if val, has := argConv.CGREvent.Event[utils.RequestType]; !has {
					return fmt.Errorf("missing RequestType")
				} else if strVal := utils.IfaceAsString(val); strVal != utils.MetaNone {
					return fmt.Errorf("Expected %+v, received %+v", utils.MetaNone, strVal)
				}
				if val, has := argConv.CGREvent.Event[utils.RunID]; !has {
					return fmt.Errorf("missing RunID")
				} else if strVal := utils.IfaceAsString(val); strVal != utils.MetaTopUp {
					return fmt.Errorf("Expected %+v, received %+v", utils.MetaNone, strVal)
				}
				return nil
			},
		},
	}
	internalCDRsChann := make(chan birpc.ClientConnector, 1)
	internalCDRsChann <- sMock
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): internalCDRsChann,
	})
	apA := &engine.APAction{
		ID:   "ACT_CDRLOG",
		Type: utils.MetaCdrLog,
	}
	cdrLogAction := &actCDRLog{
		config:  cfg,
		filterS: filterS,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "10",
			utils.Tenant:       "cgrates.org",
			utils.BalanceType:  utils.MetaConcrete,
			utils.Cost:         0.15,
			utils.ActionType:   utils.MetaTopUp,
		},
		utils.MetaOpts: map[string]interface{}{},
	}
	if err := cdrLogAction.execute(context.Background(), evNM, utils.MetaNone); err != nil {
		t.Error(err)
	}
}

func TestCDRLogActionWithOpts(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})
	sMock2 := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.CDRsV1ProcessEvent: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*engine.ArgV1ProcessEvent)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if !reflect.DeepEqual(argConv.Flags, []string{utils.ConcatenatedKey(utils.MetaChargers, "false")}) {
					return fmt.Errorf("Expected %+v, received %+v", []string{utils.ConcatenatedKey(utils.MetaChargers, "false")}, argConv.Flags)
				}
				if val, has := argConv.CGREvent.Event[utils.Tenant]; !has {
					return fmt.Errorf("missing Tenant")
				} else if strVal := utils.IfaceAsString(val); strVal != "cgrates.org" {
					return fmt.Errorf("Expected %+v, received %+v", "cgrates.org", strVal)
				}
				if val, has := argConv.CGREvent.APIOpts["EventFieldOpt"]; !has {
					return fmt.Errorf("missing EventFieldOpt from Opts")
				} else if strVal := utils.IfaceAsString(val); strVal != "eventValue" {
					return fmt.Errorf("Expected %+v, received %+v", "eventValue", strVal)
				}
				if val, has := argConv.CGREvent.APIOpts["Option1"]; !has {
					return fmt.Errorf("missing Option1 from Opts")
				} else if strVal := utils.IfaceAsString(val); strVal != "Value1" {
					return fmt.Errorf("Expected %+v, received %+v", "Value1", strVal)
				}
				if val, has := argConv.CGREvent.APIOpts["Option3"]; !has {
					return fmt.Errorf("missing Option3 from Opts")
				} else if strVal := utils.IfaceAsString(val); strVal != "eventValue" {
					return fmt.Errorf("Expected %+v, received %+v", "eventValue", strVal)
				}
				return nil
			},
		},
	}
	internalCDRsChann := make(chan birpc.ClientConnector, 1)
	internalCDRsChann <- sMock2
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().CDRsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs)}
	cfg.TemplatesCfg()["CustomTemplate"] = []*config.FCTemplate{
		{
			Tag:    "Tenant",
			Type:   "*constant",
			Path:   "*cdr.Tenant",
			Value:  config.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
			Layout: time.RFC3339,
		},
		{
			Tag:    "Opt1",
			Type:   "*constant",
			Path:   "*opts.Option1",
			Value:  config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			Layout: time.RFC3339,
		},
		{
			Tag:    "Opt2",
			Type:   "*constant",
			Path:   "*opts.Option2",
			Value:  config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			Layout: time.RFC3339,
		},
		{
			Tag:    "Opt3",
			Type:   "*variable",
			Path:   "*opts.Option3",
			Value:  config.NewRSRParsersMustCompile("~*opts.EventFieldOpt", utils.InfieldSep),
			Layout: time.RFC3339,
		},
	}
	for _, tpl := range cfg.TemplatesCfg()["CustomTemplate"] {
		tpl.ComputePath()
	}

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs): internalCDRsChann,
	})
	apA := &engine.APAction{
		ID:   "ACT_CDRLOG2",
		Type: utils.MetaCdrLog,
		Opts: map[string]interface{}{
			utils.MetaTemplateID: "CustomTemplate",
		},
	}
	cdrLogAction := &actCDRLog{
		config:  cfg,
		filterS: filterS,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "10",
			utils.Tenant:       "cgrates.org",
			utils.BalanceType:  utils.MetaConcrete,
			utils.Cost:         0.15,
			utils.ActionType:   utils.MetaTopUp,
		},
		utils.MetaOpts: map[string]interface{}{
			"EventFieldOpt": "eventValue",
		},
	}
	if err := cdrLogAction.execute(context.Background(), evNM, utils.MetaNone); err != nil {
		t.Error(err)
	}
}

func TestExportAction(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})
	sMock2 := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.EeSv1ProcessEvent: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*utils.CGREventWithEeIDs)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if argConv.CGREvent.Tenant != "cgrates.org" {
					return fmt.Errorf("Expected %+v, received %+v", "cgrates.org", argConv.CGREvent.Tenant)
				}
				return nil
			},
		},
	}
	internalCDRsChann := make(chan birpc.ClientConnector, 1)
	internalCDRsChann <- sMock2
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)}

	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs): internalCDRsChann,
	})
	apA := &engine.APAction{
		ID:   "ACT_CDRLOG2",
		Type: utils.MetaExport,
	}
	exportAction := &actExport{
		tnt:     "cgrates.org",
		config:  cfg,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "10",
			utils.Tenant:       "cgrates.org",
			utils.BalanceType:  utils.MetaConcrete,
			utils.Cost:         0.15,
			utils.ActionType:   utils.MetaTopUp,
		},
		utils.MetaOpts: map[string]interface{}{
			"EventFieldOpt": "eventValue",
		},
	}
	if err := exportAction.execute(context.Background(), evNM, utils.MetaNone); err != nil {
		t.Error(err)
	}
}

func TestExportActionWithEeIDs(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})
	sMock2 := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.EeSv1ProcessEvent: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*utils.CGREventWithEeIDs)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if argConv.CGREvent.Tenant != "cgrates.org" {
					return fmt.Errorf("Expected %+v, received %+v", "cgrates.org", argConv.CGREvent.Tenant)
				}
				if !reflect.DeepEqual(argConv.EeIDs, []string{"Exporter1", "Exporter2", "Exporter3"}) {
					return fmt.Errorf("Expected %+v, received %+v", []string{"Exporter1", "Exporter2", "Exporter3"}, argConv.EeIDs)
				}
				return nil
			},
		},
	}
	internalCDRsChann := make(chan birpc.ClientConnector, 1)
	internalCDRsChann <- sMock2
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().EEsConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)}

	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs): internalCDRsChann,
	})
	apA := &engine.APAction{
		ID:   "ACT_CDRLOG2",
		Type: utils.MetaExport,
		Opts: map[string]interface{}{
			utils.MetaExporterIDs: "Exporter1;Exporter2;Exporter3",
		},
	}
	exportAction := &actExport{
		tnt:     "cgrates.org",
		config:  cfg,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaReq: map[string]interface{}{
			utils.AccountField: "10",
			utils.Tenant:       "cgrates.org",
			utils.BalanceType:  utils.MetaConcrete,
			utils.Cost:         0.15,
			utils.ActionType:   utils.MetaTopUp,
		},
		utils.MetaOpts: map[string]interface{}{
			"EventFieldOpt": "eventValue",
		},
	}
	if err := exportAction.execute(context.Background(), evNM, utils.MetaNone); err != nil {
		t.Error(err)
	}
}

func TestExportActionResetThresholdStaticTenantID(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})
	sMock2 := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.ThresholdSv1ResetThreshold: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*utils.TenantIDWithAPIOpts)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if argConv.Tenant != "cgrates.org" {
					return fmt.Errorf("Expected %+v, received %+v", "cgrates.org", argConv.Tenant)
				}
				if argConv.ID != "TH1" {
					return fmt.Errorf("Expected %+v, received %+v", "TH1", argConv.ID)
				}
				return nil
			},
		},
	}
	internalChann := make(chan birpc.ClientConnector, 1)
	internalChann <- sMock2
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}

	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): internalChann,
	})
	apA := &engine.APAction{
		ID:      "ACT_RESET_TH",
		Type:    utils.MetaResetThreshold,
		Diktats: []*engine.APDiktat{{}},
	}
	exportAction := &actResetThreshold{
		tnt:     "cgrates.org",
		config:  cfg,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaOpts: map[string]interface{}{},
	}
	if err := exportAction.execute(context.Background(), evNM, "cgrates.org:TH1"); err != nil {
		t.Error(err)
	}
}

func TestExportActionResetThresholdStaticID(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})
	sMock2 := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.ThresholdSv1ResetThreshold: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*utils.TenantIDWithAPIOpts)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if argConv.Tenant != "cgrates.org" {
					return fmt.Errorf("Expected %+v, received %+v", "cgrates.org", argConv.Tenant)
				}
				if argConv.ID != "TH1" {
					return fmt.Errorf("Expected %+v, received %+v", "TH1", argConv.ID)
				}
				return nil
			},
		},
	}
	internalChann := make(chan birpc.ClientConnector, 1)
	internalChann <- sMock2
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().ThresholdSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)}

	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds): internalChann,
	})
	apA := &engine.APAction{
		ID:      "ACT_RESET_TH",
		Type:    utils.MetaResetThreshold,
		Diktats: []*engine.APDiktat{{}},
	}
	exportAction := &actResetThreshold{
		tnt:     "cgrates.org",
		config:  cfg,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaOpts: map[string]interface{}{},
	}
	if err := exportAction.execute(context.Background(), evNM, "TH1"); err != nil {
		t.Error(err)
	}
}

func TestExportActionResetStatStaticTenantID(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})
	sMock2 := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.StatSv1ResetStatQueue: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*utils.TenantIDWithAPIOpts)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if argConv.Tenant != "cgrates.org" {
					return fmt.Errorf("Expected %+v, received %+v", "cgrates.org", argConv.Tenant)
				}
				if argConv.ID != "ST1" {
					return fmt.Errorf("Expected %+v, received %+v", "TH1", argConv.ID)
				}
				return nil
			},
		},
	}
	internalChann := make(chan birpc.ClientConnector, 1)
	internalChann <- sMock2
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}

	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats): internalChann,
	})
	apA := &engine.APAction{
		ID:      "ACT_RESET_ST",
		Type:    utils.MetaResetStatQueue,
		Diktats: []*engine.APDiktat{{}},
	}
	exportAction := &actResetStat{
		tnt:     "cgrates.org",
		config:  cfg,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaOpts: map[string]interface{}{},
	}
	if err := exportAction.execute(context.Background(), evNM, "cgrates.org:ST1"); err != nil {
		t.Error(err)
	}
}

func TestExportActionResetStatStaticID(t *testing.T) {
	// Clear cache because connManager sets the internal connection in cache
	engine.Cache.Clear([]string{utils.CacheRPCConnections})
	sMock2 := &testMockCDRsConn{
		calls: map[string]func(_ *context.Context, _, _ interface{}) error{
			utils.StatSv1ResetStatQueue: func(_ *context.Context, arg, rply interface{}) error {
				argConv, can := arg.(*utils.TenantIDWithAPIOpts)
				if !can {
					return fmt.Errorf("Wrong argument type: %T", arg)
				}
				if argConv.Tenant != "cgrates.org" {
					return fmt.Errorf("Expected %+v, received %+v", "cgrates.org", argConv.Tenant)
				}
				if argConv.ID != "ST1" {
					return fmt.Errorf("Expected %+v, received %+v", "TH1", argConv.ID)
				}
				return nil
			},
		},
	}
	internalChann := make(chan birpc.ClientConnector, 1)
	internalChann <- sMock2
	cfg := config.NewDefaultCGRConfig()
	cfg.ActionSCfg().StatSConns = []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}

	connMgr := engine.NewConnManager(config.CgrConfig(), map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats): internalChann,
	})
	apA := &engine.APAction{
		ID:   "ACT_RESET_ST",
		Type: utils.MetaResetStatQueue,
		Diktats: []*engine.APDiktat{{
			Value: "ST1",
		}},
	}
	exportAction := &actResetStat{
		tnt:     "cgrates.org",
		config:  cfg,
		connMgr: connMgr,
		aCfg:    apA,
	}
	evNM := utils.MapStorage{
		utils.MetaOpts: map[string]interface{}{},
	}
	if err := exportAction.execute(context.Background(), evNM, "ST1"); err != nil {
		t.Error(err)
	}
}

func TestACScheduledActions(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "TestACScheduledActions",
		FilterIDs: []string{"*string:~*req.Destination:1005"},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "inexistent_type",
				Diktats: []*engine.APDiktat{{
					Path:  "~*balance.TestBalance.Value",
					Value: "10",
				}},
			},
		},
	}

	if err := dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.Destination: "1005",
		},
	}

	var err error
	utils.Logger, err = utils.Newlogger(utils.MetaStdLog, utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	utils.Logger.SetLogLevel(7)

	buff := new(bytes.Buffer)
	log.SetOutput(buff)

	acts := NewActionS(cfg, fltrs, dm, nil)
	expected := "WARNING] <ActionS> ignoring ActionProfile with id: <cgrates.org:TestACScheduledActions> creating action: <TOPUP>, error: <unsupported action type: <inexistent_type>>"
	if _, err := acts.scheduledActions(context.Background(), "cgrates.org", cgrEv, []string{}, true); err != nil {
		t.Error(err)
	} else if rcv := buff.String(); !strings.Contains(rcv, expected) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
	buff.Reset()

	actPrf.Actions[0].Type = utils.MetaResetStatQueue
	actPrf.Targets = map[string]utils.StringSet{
		utils.MetaStats: map[string]struct{}{
			"ID_TEST": {},
		},
	}
	if err := dm.SetActionProfile(context.Background(), actPrf, true); err != nil {
		t.Error(err)
	}

	mapStorage := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.APIOpts,
	}
	expectedSChed := []*scheduledActs{
		{
			tenant:   "cgrates.org",
			apID:     "TestACScheduledActions",
			trgTyp:   utils.MetaStats,
			trgID:    "ID_TEST",
			schedule: utils.EmptyString,
			ctx:      context.Background(),
			data:     mapStorage,
		},
	}
	var schedActs []*scheduledActs
	if schedActs, err = acts.scheduledActions(context.Background(), "cgrates.org", cgrEv, []string{}, true); err != nil {
		t.Error(err)
	} else {

	}
	//execute asap the actions
	schedActs[0].trgID = "invalid_type"
	if err := acts.asapExecuteActions(context.Background(), schedActs[0]); err == nil || err != utils.ErrPartiallyExecuted {
		t.Errorf("Expected %+v, received %+v", utils.ErrPartiallyExecuted, err)
	}

	schedActs[0].trgID = "ID_TEST"
	schedActs[0].acts = nil
	schedActs[0].cch = nil
	if !reflect.DeepEqual(schedActs, expectedSChed) {
		t.Errorf("Expected %+v, received %+v", expectedSChed, schedActs)
	}
}
