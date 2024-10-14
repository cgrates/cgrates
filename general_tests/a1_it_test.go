//go:build integration
// +build integration

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
package general_tests

import (
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	a1ConfigDir string
	a1CfgPath   string
	a1Cfg       *config.CGRConfig
	a1rpc       *birpc.Client

	sTestsA1it = []func(t *testing.T){
		testA1itLoadConfig,
		testA1itResetDataDB,
		testA1itResetStorDb,
		testA1itStartEngine,
		testA1itRPCConn,
		testA1itLoadTPFromFolder,
		testA1itAddBalance1,
		testA1itDataSession1,
		testA1itConcurrentAPs,
		testA1itStopCgrEngine,
	}
)

func TestA1It(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		a1ConfigDir = "tutinternal"
	case utils.MetaMySQL:
		a1ConfigDir = "tutmysql"
	case utils.MetaMongo:
		a1ConfigDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsA1it {
		t.Run(a1ConfigDir, stest)
	}
}

func testA1itLoadConfig(t *testing.T) {
	var err error
	a1CfgPath = path.Join(*utils.DataDir, "conf", "samples", a1ConfigDir)
	if a1Cfg, err = config.NewCGRConfigFromPath(a1CfgPath); err != nil {
		t.Error(err)
	}
}

func testA1itResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(a1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testA1itResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(a1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testA1itStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(a1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testA1itRPCConn(t *testing.T) {
	a1rpc = engine.NewRPCClient(t, a1Cfg.ListenCfg())
}

func testA1itLoadTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "test", "a1")}
	if err := a1rpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	time.Sleep(100 * time.Millisecond)
	tStart := time.Date(2017, 3, 3, 10, 39, 33, 0, time.UTC)
	tEnd := time.Date(2017, 3, 3, 10, 39, 33, 10240, time.UTC)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:    "data1",
			Tenant:      "cgrates.org",
			Subject:     "rpdata1",
			Destination: "data",
			TimeStart:   tStart,
			TimeEnd:     tEnd,
		},
	}
	var cc engine.CallCost
	if err := a1rpc.Call(context.Background(), utils.ResponderGetCost, cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.0 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}

	//add a default charger
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result string
	if err := a1rpc.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testA1itAddBalance1(t *testing.T) {
	var reply string
	argAdd := &v1.AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "rpdata1",
		BalanceType: utils.MetaData,
		Value:       10000000000,
		Balance: map[string]any{
			utils.ID: "rpdata1_test",
		},
	}
	if err := a1rpc.Call(context.Background(), utils.APIerSv1AddBalance, argAdd, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf(reply)
	}
	argGet := &utils.AttrGetAccount{Tenant: argAdd.Tenant, Account: argAdd.Account}
	var acnt *engine.Account
	if err := a1rpc.Call(context.Background(), utils.APIerSv2GetAccount, argGet, &acnt); err != nil {
		t.Error(err)
	} else {
		if acnt.BalanceMap[utils.MetaData].GetTotalValue() != argAdd.Value { // We expect 11.5 since we have added in the previous test 1.5
			t.Errorf("Received account value: %f", acnt.BalanceMap[utils.MetaData].GetTotalValue())
		}
	}
}

func testA1itDataSession1(t *testing.T) {
	usage := time.Duration(10240)
	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestA1itDataSession1",
			Event: map[string]any{
				utils.EventName:    "INITIATE_SESSION",
				utils.ToR:          utils.MetaData,
				utils.OriginID:     "504966119",
				utils.AccountField: "rpdata1",
				utils.Subject:      "rpdata1",
				utils.Destination:  "data",
				utils.Category:     "data1",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    "2017-03-03 11:39:32 +0100 CET",
				utils.AnswerTime:   "2017-03-03 11:39:32 +0100 CET",
				utils.Usage:        "10240",
			},
			APIOpts: map[string]any{
				utils.OptsSessionsTTL:         "28800s",
				utils.OptsSessionsTTLLastUsed: "0s",
				utils.OptsSessionsTTLUsage:    "0s",
			},
		},
	}

	var initRpl *sessions.V1InitSessionReply
	if err := a1rpc.Call(context.Background(), utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Fatal(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	updateArgs := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:       "UPDATE_SESSION",
				utils.AccountField:    "rpdata1",
				utils.Category:        "data1",
				utils.Destination:     "data",
				utils.InitialOriginID: "504966119",
				utils.LastUsed:        "0s",
				utils.OriginID:        "504966119-1",
				utils.RequestType:     utils.MetaPrepaid,
				utils.Subject:         "rpdata1",
				utils.Tenant:          "cgrates.org",
				utils.ToR:             utils.MetaData,
				utils.SetupTime:       "2017-03-03 11:39:32 +0100 CET",
				utils.AnswerTime:      "2017-03-03 11:39:32 +0100 CET",
				utils.Usage:           "2097152",
			},
			APIOpts: map[string]any{
				utils.OptsSessionsTTL:         "28800s",
				utils.OptsSessionsTTLLastUsed: "2097152s",
				utils.OptsSessionsTTLUsage:    "0s",
			},
		},
	}

	usage = 2097152
	var updateRpl *sessions.V1UpdateSessionReply
	if err := a1rpc.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	usage = time.Minute
	termArgs := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:      "TERMINATE_SESSION",
				utils.AccountField:   "rpdata1",
				utils.Category:       "data1",
				utils.Destination:    "data",
				utils.LastUsed:       "2202800",
				utils.OriginID:       "504966119-1",
				utils.OriginIDPrefix: "504966119-1",
				utils.RequestType:    utils.MetaPrepaid,
				utils.SetupTime:      "2017-03-03 11:39:32 +0100 CET",
				utils.AnswerTime:     "2017-03-03 11:39:32 +0100 CET",
				utils.Subject:        "rpdata1",
				utils.Tenant:         "cgrates.org",
				utils.ToR:            utils.MetaData,
			},
		},
	}

	var rpl string
	if err := a1rpc.Call(context.Background(), utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	if err := a1rpc.Call(context.Background(), utils.SessionSv1ProcessCDR, termArgs.CGREvent, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.OK {
		t.Errorf("Received reply: %s", rpl)
	}

	time.Sleep(20 * time.Millisecond)

	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}
	if err := a1rpc.Call(context.Background(), utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "2202800" {
			t.Errorf("Unexpected CDR Usage received, cdr: %+v ", cdrs[0])
		}
		var cc engine.CallCost
		var ec engine.EventCost
		if err := json.Unmarshal([]byte(cdrs[0].CostDetails), &ec); err != nil {
			t.Error(err)
		}
		cc = *ec.AsCallCost(utils.EmptyString)
		if len(cc.Timespans) != 1 {
			t.Errorf("Unexpected number of timespans: %+v, for %+v\n from:%+v", len(cc.Timespans), utils.ToJSON(cc.Timespans), utils.ToJSON(ec))
		}
		if cc.RatedUsage != 2202800 {
			t.Errorf("RatingUsage expected: %f received %f, callcost: %+v ", 2202800.0, cc.RatedUsage, cc)
		}
	}
	expBalance := float64(10000000000 - 2202800) // initial - total usage
	var acnt *engine.Account
	if err := a1rpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "rpdata1"}, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaData].GetTotalValue() != expBalance { // We expect 11.5 since we have added in the previous test 1.5
		t.Errorf("Expecting: %f, received: %f", expBalance, acnt.BalanceMap[utils.MetaData].GetTotalValue())
	}
}

func testA1itConcurrentAPs(t *testing.T) {
	var wg sync.WaitGroup
	var acnts []string
	for i := 0; i < 1000; i++ {
		acnts = append(acnts, fmt.Sprintf("acnt_%d", i))
	}
	// Set initial action plans
	for _, acnt := range acnts {
		wg.Add(1)
		go func(acnt string) {
			attrSetAcnt := v2.AttrSetAccount{
				Tenant:        "cgrates.org",
				Account:       acnt,
				ActionPlanIDs: []string{"PACKAGE_1"},
			}
			var reply string
			if err := a1rpc.Call(context.Background(), utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(acnt)
	}
	wg.Wait()
	// Make sure action plan was properly set
	var aps []*engine.ActionPlan
	if err := a1rpc.Call(context.Background(), utils.APIerSv1GetActionPlan, &v1.AttrGetActionPlan{ID: "PACKAGE_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps[0].AccountIDs.Slice()) != len(acnts) {
		t.Errorf("Received: %+v", aps[0])
	}
	// Change offer
	for _, acnt := range acnts {
		wg.Add(3)
		go func(acnt string) {
			var atms []*v1.AccountActionTiming
			if err := a1rpc.Call(context.Background(), utils.APIerSv1GetAccountActionPlan,
				&utils.TenantAccount{Tenant: "cgrates.org", Account: acnt}, &atms); err != nil {
				t.Error(err)
				//} else if len(atms) != 2 || atms[0].ActionPlanId != "PACKAGE_1" {
				//	t.Errorf("Received: %+v", atms)
			}
			wg.Done()
		}(acnt)
		go func(acnt string) {
			var reply string
			if err := a1rpc.Call(context.Background(), utils.APIerSv1RemoveActionTiming,
				&v1.AttrRemoveActionTiming{Tenant: "cgrates.org", Account: acnt, ActionPlanId: "PACKAGE_1"}, &reply); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(acnt)
		go func(acnt string) {
			attrSetAcnt := v2.AttrSetAccount{
				Tenant:        "cgrates.org",
				Account:       acnt,
				ActionPlanIDs: []string{"PACKAGE_2"},
			}
			var reply string
			if err := a1rpc.Call(context.Background(), utils.APIerSv2SetAccount, &attrSetAcnt, &reply); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(acnt)
	}
	wg.Wait()
	// Make sure action plan was properly rem/set
	aps = []*engine.ActionPlan{}
	if err := a1rpc.Call(context.Background(), utils.APIerSv1GetActionPlan, &v1.AttrGetActionPlan{ID: "PACKAGE_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps[0].AccountIDs.Slice()) != 0 {
		t.Errorf("Received: %+v", aps[0])
	}
	aps = []*engine.ActionPlan{}
	if err := a1rpc.Call(context.Background(), utils.APIerSv1GetActionPlan, &v1.AttrGetActionPlan{ID: "PACKAGE_2"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps[0].AccountIDs.Slice()) != len(acnts) {
		t.Errorf("Received: %+v", aps[0])
	}
}

func testA1itStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
