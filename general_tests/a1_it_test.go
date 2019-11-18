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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"sync"
	"testing"
	"time"

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
	a1rpc       *rpc.Client
)

var sTestsA1it = []func(t *testing.T){
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

func TestA1ItMongo(t *testing.T) {
	a1ConfigDir = "tutmongo"
	for _, stest := range sTestsA1it {
		t.Run(a1ConfigDir, stest)
	}
}

func TestA1ItRedis(t *testing.T) {
	a1ConfigDir = "tutmysql"
	for _, stest := range sTestsA1it {
		t.Run(a1ConfigDir, stest)
	}
}

func TestA1ItInternal(t *testing.T) {
	a1ConfigDir = "tutinternal"
	for _, stest := range sTestsA1it {
		t.Run(a1ConfigDir, stest)
	}
}

func testA1itLoadConfig(t *testing.T) {
	a1CfgPath = path.Join(*dataDir, "conf", "samples", a1ConfigDir)
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
	if _, err := engine.StopStartEngine(a1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testA1itRPCConn(t *testing.T) {
	var err error
	a1rpc, err = jsonrpc.Dial("tcp", a1Cfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func testA1itLoadTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "test", "a1")}
	if err := a1rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	time.Sleep(time.Duration(100 * time.Millisecond))
	tStart := time.Date(2017, 3, 3, 10, 39, 33, 0, time.UTC)
	tEnd := time.Date(2017, 3, 3, 10, 39, 33, 10240, time.UTC)
	cd := engine.CallDescriptor{
		Category:    "data1",
		Tenant:      "cgrates.org",
		Subject:     "rpdata1",
		Destination: "data",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	var cc engine.CallCost
	if err := a1rpc.Call(utils.ResponderGetCost, cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.0 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}

	//add a default charger
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result string
	if err := a1rpc.Call(utils.ApierV1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testA1itAddBalance1(t *testing.T) {
	var reply string
	argAdd := &v1.AttrAddBalance{Tenant: "cgrates.org", Account: "rpdata1",
		BalanceType: utils.DATA, BalanceId: utils.StringPointer("rpdata1_test"),
		Value: 10000000000}
	if err := a1rpc.Call(utils.ApierV1AddBalance, argAdd, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf(reply)
	}
	argGet := &utils.AttrGetAccount{Tenant: argAdd.Tenant, Account: argAdd.Account}
	var acnt *engine.Account
	if err := a1rpc.Call("ApierV2.GetAccount", argGet, &acnt); err != nil {
		t.Error(err)
	} else {
		if acnt.BalanceMap[utils.DATA].GetTotalValue() != argAdd.Value { // We expect 11.5 since we have added in the previous test 1.5
			t.Errorf("Received account value: %f", acnt.BalanceMap[utils.DATA].GetTotalValue())
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
			Event: map[string]interface{}{
				utils.EVENT_NAME:         "INITIATE_SESSION",
				utils.ToR:                utils.DATA,
				utils.OriginID:           "504966119",
				utils.Account:            "rpdata1",
				utils.Subject:            "rpdata1",
				utils.Destination:        "data",
				utils.Category:           "data1",
				utils.Tenant:             "cgrates.org",
				utils.RequestType:        utils.META_PREPAID,
				utils.SetupTime:          "2017-03-03 11:39:32 +0100 CET",
				utils.AnswerTime:         "2017-03-03 11:39:32 +0100 CET",
				utils.Usage:              "10240",
				utils.SessionTTL:         "28800s",
				utils.SessionTTLLastUsed: "0s",
				utils.SessionTTLUsage:    "0s",
			},
		},
	}

	var initRpl *sessions.V1InitSessionReply
	if err := a1rpc.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, *initRpl.MaxUsage)
	}

	updateArgs := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]interface{}{
				utils.EVENT_NAME:         "UPDATE_SESSION",
				utils.Account:            "rpdata1",
				utils.Category:           "data1",
				utils.Destination:        "data",
				utils.InitialOriginID:    "504966119",
				utils.LastUsed:           "0s",
				utils.OriginID:           "504966119-1",
				utils.RequestType:        utils.META_PREPAID,
				utils.SessionTTL:         "28800s",
				utils.SessionTTLLastUsed: "2097152s",
				utils.SessionTTLUsage:    "0s",
				utils.Subject:            "rpdata1",
				utils.Tenant:             "cgrates.org",
				utils.ToR:                utils.DATA,
				utils.SetupTime:          "2017-03-03 11:39:32 +0100 CET",
				utils.AnswerTime:         "2017-03-03 11:39:32 +0100 CET",
				utils.Usage:              "2097152",
			},
		},
	}

	usage = time.Duration(2097152)
	var updateRpl *sessions.V1UpdateSessionReply
	if err := a1rpc.Call(utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, *updateRpl.MaxUsage)
	}

	usage = time.Duration(1 * time.Minute)
	termArgs := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]interface{}{
				utils.EVENT_NAME:     "TERMINATE_SESSION",
				utils.Account:        "rpdata1",
				utils.Category:       "data1",
				utils.Destination:    "data",
				utils.LastUsed:       "2202800",
				utils.OriginID:       "504966119-1",
				utils.OriginIDPrefix: "504966119-1",
				utils.RequestType:    utils.META_PREPAID,
				utils.SetupTime:      "2017-03-03 11:39:32 +0100 CET",
				utils.AnswerTime:     "2017-03-03 11:39:32 +0100 CET",
				utils.Subject:        "rpdata1",
				utils.Tenant:         "cgrates.org",
				utils.ToR:            utils.DATA,
			},
		},
	}

	var rpl string
	if err := a1rpc.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	if err := a1rpc.Call(utils.SessionSv1ProcessCDR, termArgs.CGREvent, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.OK {
		t.Errorf("Received reply: %s", rpl)
	}

	time.Sleep(time.Duration(20) * time.Millisecond)

	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}}
	if err := a1rpc.Call(utils.ApierV2GetCDRs, req, &cdrs); err != nil {
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
		cc = *ec.AsCallCost()
		if len(cc.Timespans) != 1 {
			t.Errorf("Unexpected number of timespans: %+v, for %+v\n from:%+v", len(cc.Timespans), utils.ToJSON(cc.Timespans), utils.ToJSON(ec))
		}
		if cc.RatedUsage != 2202800 {
			t.Errorf("RatingUsage expected: %f received %f, callcost: %+v ", 2202800.0, cc.RatedUsage, cc)
		}
	}
	expBalance := float64(10000000000 - 2202800) // initial - total usage
	var acnt *engine.Account
	if err := a1rpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "rpdata1"}, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != expBalance { // We expect 11.5 since we have added in the previous test 1.5
		t.Errorf("Expecting: %f, received: %f", expBalance, acnt.BalanceMap[utils.DATA].GetTotalValue())
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
				ActionPlanIDs: &[]string{"PACKAGE_1"},
			}
			var reply string
			if err := a1rpc.Call("ApierV2.SetAccount", attrSetAcnt, &reply); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(acnt)
	}
	wg.Wait()
	// Make sure action plan was properly set
	var aps []*engine.ActionPlan
	if err := a1rpc.Call("ApierV1.GetActionPlan", v1.AttrGetActionPlan{ID: "PACKAGE_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps[0].AccountIDs.Slice()) != len(acnts) {
		t.Errorf("Received: %+v", aps[0])
	}
	// Change offer
	for _, acnt := range acnts {
		wg.Add(3)
		go func(acnt string) {
			var atms []*v1.AccountActionTiming
			if err := a1rpc.Call(utils.ApierV1GetAccountActionPlan,
				utils.TenantAccount{Tenant: "cgrates.org", Account: acnt}, &atms); err != nil {
				t.Error(err)
				//} else if len(atms) != 2 || atms[0].ActionPlanId != "PACKAGE_1" {
				//	t.Errorf("Received: %+v", atms)
			}
			wg.Done()
		}(acnt)
		go func(acnt string) {
			var reply string
			if err := a1rpc.Call("ApierV1.RemoveActionTiming",
				v1.AttrRemoveActionTiming{Tenant: "cgrates.org", Account: acnt, ActionPlanId: "PACKAGE_1"}, &reply); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(acnt)
		go func(acnt string) {
			attrSetAcnt := v2.AttrSetAccount{
				Tenant:        "cgrates.org",
				Account:       acnt,
				ActionPlanIDs: &[]string{"PACKAGE_2"},
			}
			var reply string
			if err := a1rpc.Call("ApierV2.SetAccount", attrSetAcnt, &reply); err != nil {
				t.Error(err)
			}
			wg.Done()
		}(acnt)
	}
	wg.Wait()
	// Make sure action plan was properly rem/set
	aps = []*engine.ActionPlan{}
	if err := a1rpc.Call("ApierV1.GetActionPlan", v1.AttrGetActionPlan{ID: "PACKAGE_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps[0].AccountIDs.Slice()) != 0 {
		t.Errorf("Received: %+v", aps[0])
	}
	aps = []*engine.ActionPlan{}
	if err := a1rpc.Call("ApierV1.GetActionPlan", v1.AttrGetActionPlan{ID: "PACKAGE_2"}, &aps); err != nil {
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
