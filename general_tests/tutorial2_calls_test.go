//go:build call
// +build call

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

// import (
// 	"flag"
// 	"net/rpc"
// 	"net/rpc/jsonrpc"
// 	"os"
// 	"path"
// 	"reflect"
// 	"sort"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/sessions"
// 	"github.com/cgrates/cgrates/utils"
// )

// var tutorial2CallsCfg *config.CGRConfig
// var tutorial2CallsRpc *rpc.Client
// var tutorial2CallsPjSuaListener *os.File
// var tutorial2FSConfig = flag.String("tutorial2FSConfig", "/usr/share/cgrates/tutorial_tests/fs_evsock", "FreeSwitch tutorial folder")
// var tutorial2OptConf string

// var sTestsTutorial2Calls = []func(t *testing.T){
// 	testTutorial2CallInitCfg,
// 	testTutorial2CallResetDataDb,
// 	testTutorial2CallResetStorDb,
// 	testTutorial2CallStartFS,
// 	testTutorial2CallStartEngine,
// 	testTutorial2CallRestartFS,
// 	testTutorial2CallRpcConn,
// 	testTutorial2CallLoadTariffPlanFromFolder,
// 	testTutorial2CallAccountsBefore,
// 	testTutorial2CallStatMetricsBefore,
// 	testTutorial2CallCheckResourceBeforeAllocation,
// 	testTutorial2CallCheckThreshold1001Before,
// 	testTutorial2CallCheckThreshold1002Before,
// 	testTutorial2CallStartPjsuaListener,
// 	testTutorial2CallCall1001To1002,
// 	testTutorial2CallGetActiveSessions,
// 	testTutorial2CallCheckResourceAllocation,
// 	testTutorial2CallAccount1001,
// 	testTutorial2Call1001Cdrs,
// 	testTutorial2CallStatMetrics,
// 	testTutorial2CallCheckResourceRelease,
// 	testTutorial2CallCheckThreshold1001After,
// 	testTutorial2CallCheckThreshold1002After,
// 	testTutorial2CallSyncSessions,
// 	testTutorial2CallStopPjsuaListener,
// 	testTutorial2CallStopCgrEngine,
// 	testTutorial2CallStopFS,
// }

// //Test start here
// func TestTutorial2FreeswitchCalls(t *testing.T) {
// 	tutorial2OptConf = utils.Freeswitch
// 	for _, stest := range sTestsTutorial2Calls {
// 		t.Run("", stest)
// 	}
// }

// func testTutorial2CallInitCfg(t *testing.T) {
// 	// Init config first
// 	var err error
// 	switch tutorial2OptConf {
// 	case utils.Freeswitch:
// 		tutorial2CallsCfg, err = config.NewCGRConfigFromPath(path.Join(*tutorial2FSConfig, "cgrates", "etc", "cgrates"))
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	default:
// 		t.Error("Invalid option")
// 	}

// 	tutorial2CallsCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
// 	config.SetCgrConfig(tutorial2CallsCfg)
// }

// // Remove data in both rating and accounting db
// func testTutorial2CallResetDataDb(t *testing.T) {
// 	if err := engine.InitDataDb(tutorial2CallsCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Wipe out the cdr database
// func testTutorial2CallResetStorDb(t *testing.T) {
// 	if err := engine.InitStorDb(tutorial2CallsCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // start FS server
// func testTutorial2CallStartFS(t *testing.T) {
// 	switch tutorial2OptConf {
// 	case utils.Freeswitch:
// 		engine.KillProcName(utils.Freeswitch, 5000)
// 		if err := engine.CallScript(path.Join(*tutorial2FSConfig, "freeswitch", "etc", "init.d", "freeswitch"), "start", 3000); err != nil {
// 			t.Fatal(err)
// 		}
// 	default:
// 		t.Error("Invalid option")
// 	}
// }

// // Start CGR Engine
// func testTutorial2CallStartEngine(t *testing.T) {
// 	engine.KillProcName("cgr-engine", *waitRater)
// 	switch tutorial2OptConf {
// 	case utils.Freeswitch:
// 		if err := engine.CallScript(path.Join(*tutorial2FSConfig, "cgrates", "etc", "init.d", "cgrates"), "start", 100); err != nil {
// 			t.Fatal(err)
// 		}
// 	default:
// 		t.Error("invalid option")
// 	}
// }

// // Restart FS so we make sure reconnects are working
// func testTutorial2CallRestartFS(t *testing.T) {
// 	switch tutorial2OptConf {
// 	case utils.Freeswitch:
// 		engine.KillProcName(utils.Freeswitch, 5000)
// 		if err := engine.CallScript(path.Join(*tutorial2FSConfig, "freeswitch", "etc", "init.d", "freeswitch"), "restart", 3000); err != nil {
// 			t.Fatal(err)
// 		}
// 	default:
// 		t.Error("Invalid option")
// 	}
// }

// // Connect rpc client to rater
// func testTutorial2CallRpcConn(t *testing.T) {
// 	var err error
// 	tutorial2CallsRpc, err = jsonrpc.Dial(utils.TCP, tutorial2CallsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Load the tariff plan, creating accounts and their balances
// func testTutorial2CallLoadTariffPlanFromFolder(t *testing.T) {
// 	var reply string
// 	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
// 	if err := tutorial2CallsRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
// }

// // Make sure account was debited properly
// func testTutorial2CallAccountsBefore(t *testing.T) {
// 	var reply *engine.Account
// 	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
// 	if err := tutorial2CallsRpc.Call(utils.APIerSv2GetAccount, attrs, &reply); err != nil {
// 		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
// 	} else if reply.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10.0 {
// 		t.Errorf("Calling APIerSv1.GetBalance received: %f", reply.BalanceMap[utils.MetaMonetary].GetTotalValue())
// 	}
// 	var reply2 *engine.Account
// 	attrs2 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}
// 	if err := tutorial2CallsRpc.Call(utils.APIerSv2GetAccount, attrs2, &reply2); err != nil {
// 		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
// 	} else if reply2.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10.0 {
// 		t.Errorf("Calling APIerSv1.GetBalance received: %f", reply2.BalanceMap[utils.MetaMonetary].GetTotalValue())
// 	}
// }

// func testTutorial2CallStatMetricsBefore(t *testing.T) {
// 	var metrics map[string]string
// 	expectedMetrics := map[string]string{
// 		utils.MetaTCC: utils.NotAvailable,
// 		utils.MetaTCD: utils.NotAvailable,
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.StatSv1GetQueueStringMetrics,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &metrics); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
// 		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.StatSv1GetQueueStringMetrics,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2_1"}, &metrics); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
// 		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
// 	}
// }

// func testTutorial2CallCheckResourceBeforeAllocation(t *testing.T) {
// 	var rs *engine.Resources
// 	args := &utils.ArgRSv1ResourceUsage{
// 		UsageID: "OriginID",
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "ResourceEvent",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 				utils.Subject:      "1001",
// 				utils.Destination:  "1002",
// 			},
// 		},
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
// 		t.Fatal(err)
// 	} else if len(*rs) != 1 {
// 		t.Fatalf("Resources: %+v", utils.ToJSON(rs))
// 	}
// 	for _, r := range *rs {
// 		if r.ID == "ResGroup1" &&
// 			(len(r.Usages) != 0 || len(r.TTLIdx) != 0) {
// 			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
// 		}
// 	}
// }

// func testTutorial2CallCheckThreshold1001Before(t *testing.T) {
// 	var td engine.Threshold
// 	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1001", Hits: 0}
// 	if err := tutorial2CallsRpc.Call(utils.ThresholdSv1GetThreshold,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eTd, td) {
// 		t.Errorf("expecting: %+v, received: %+v", eTd, td)
// 	}
// }

// func testTutorial2CallCheckThreshold1002Before(t *testing.T) {
// 	var td engine.Threshold
// 	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1002", Hits: 0}
// 	if err := tutorial2CallsRpc.Call(utils.ThresholdSv1GetThreshold,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}, &td); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eTd, td) {
// 		t.Errorf("expecting: %+v, received: %+v", eTd, td)
// 	}
// }

// // Start Pjsua as listener and register it to receive calls
// func testTutorial2CallStartPjsuaListener(t *testing.T) {
// 	var err error
// 	acnts := []*engine.PjsuaAccount{
// 		{Id: "sip:1001@127.0.0.1",
// 			Username: "1001", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5080"},
// 		{Id: "sip:1002@127.0.0.1",
// 			Username: "1002", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5080"},
// 	}
// 	if tutorial2CallsPjSuaListener, err = engine.StartPjsuaListener(
// 		acnts, 5070, time.Duration(*waitRater)*time.Millisecond); err != nil {
// 		t.Fatal(err)
// 	}
// 	time.Sleep(3 * time.Second)
// }

// // Call from 1001 (prepaid) to 1002
// func testTutorial2CallCall1001To1002(t *testing.T) {
// 	if err := engine.PjsuaCallUri(
// 		&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*"},
// 		"sip:1002@127.0.0.1", "sip:127.0.0.1:5080", 67*time.Second, 5071); err != nil {
// 		t.Fatal(err)
// 	}
// 	// give time to session to start so we can check it
// 	time.Sleep(time.Second)
// }

// // GetActiveSessions
// func testTutorial2CallGetActiveSessions(t *testing.T) {
// 	var reply *[]*sessions.ExternalSession
// 	expected := &[]*sessions.ExternalSession{
// 		{
// 			RequestType: "*prepaid",
// 			Tenant:      "cgrates.org",
// 			Category:    "call",
// 			Account:     "1001",
// 			Subject:     "1001",
// 			Destination: "1002",
// 		},
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.SessionSv1GetActiveSessions,
// 		nil, &reply); err != nil {
// 		t.Error("Got error on SessionSv1.GetActiveSessions: ", err.Error())
// 	} else {
// 		if len(*reply) == 2 {
// 			sort.Slice(*reply, func(i, j int) bool {
// 				return strings.Compare((*reply)[i].RequestType, (*reply)[j].RequestType) > 0
// 			})
// 		}
// 		// compare some fields (eg. CGRId is generated)
// 		if !reflect.DeepEqual((*expected)[0].RequestType, (*reply)[0].RequestType) {
// 			t.Errorf("Expected: %s, received: %s", (*expected)[0].RequestType, (*reply)[0].RequestType)
// 		} else if !reflect.DeepEqual((*expected)[0].Account, (*reply)[0].Account) {
// 			t.Errorf("Expected: %s, received: %s", (*expected)[0].Account, (*reply)[0].Account)
// 		} else if !reflect.DeepEqual((*expected)[0].Destination, (*reply)[0].Destination) {
// 			t.Errorf("Expected: %s, received: %s", (*expected)[0].Destination, (*reply)[0].Destination)
// 		}
// 	}
// }

// // Check if the resource was Allocated
// func testTutorial2CallCheckResourceAllocation(t *testing.T) {
// 	var rs *engine.Resources
// 	args := &utils.ArgRSv1ResourceUsage{
// 		UsageID: "OriginID1",
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "ResourceAllocation",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 				utils.Subject:      "1001",
// 				utils.Destination:  "1002",
// 			},
// 		},
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
// 		t.Fatal(err)
// 	} else if len(*rs) != 1 {
// 		t.Fatalf("Resources: %+v", utils.ToJSON(rs))
// 	}
// 	for _, r := range *rs {
// 		if r.ID == "ResGroup1" && len(r.Usages) != 3 {
// 			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
// 		}
// 	}
// 	// Allow calls to finish before start querying the results
// 	time.Sleep(50 * time.Second)
// }

// // Make sure account was debited properly
// func testTutorial2CallAccount1001(t *testing.T) {
// 	var reply *engine.Account
// 	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
// 	if err := tutorial2CallsRpc.Call(utils.APIerSv2GetAccount, attrs, &reply); err != nil {
// 		t.Error(err.Error())
// 	} else if reply.BalanceMap[utils.MetaMonetary].GetTotalValue() == 10.0 { // Make sure we debitted
// 		t.Errorf("Expected: 10, received: %+v", reply.BalanceMap[utils.MetaMonetary].GetTotalValue())
// 	} else if reply.Disabled == true {
// 		t.Error("Account disabled")
// 	}
// }

// // Make sure account was debited properly
// func testTutorial2Call1001Cdrs(t *testing.T) {
// 	var reply []*engine.ExternalCDR
// 	req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}, Accounts: []string{"1001"}}
// 	if err := tutorial2CallsRpc.Call(utils.APIerSv2GetCDRs, &req, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(reply) != 2 {
// 		t.Error("Unexpected number of CDRs returned: ", len(reply))
// 	} else {
// 		for _, cdr := range reply {
// 			if cdr.RequestType != utils.MetaPrepaid {
// 				t.Errorf("Unexpected RequestType for CDR: %+v", cdr.RequestType)
// 			}
// 			if cdr.Destination == "1002" {
// 				if cdr.Usage != "1m7s" && cdr.Usage != "1m8s" { // Usage as seconds
// 					t.Errorf("Unexpected Usage for CDR: %+v", cdr.Usage)
// 				}
// 				if cdr.CostSource != utils.MetaSessionS {
// 					t.Errorf("Unexpected CostSource for CDR: %+v", cdr.CostSource)
// 				}
// 			}
// 		}
// 	}
// }

// func testTutorial2CallStatMetrics(t *testing.T) {
// 	var metrics map[string]string
// 	firstStatMetrics1 := map[string]string{
// 		utils.MetaTCC: "1.35346",
// 		utils.MetaTCD: "2m27s",
// 	}
// 	firstStatMetrics2 := map[string]string{
// 		utils.MetaTCC: "1.35009",
// 		utils.MetaTCD: "2m25s",
// 	}
// 	firstStatMetrics3 := map[string]string{
// 		utils.MetaTCC: "1.34009",
// 		utils.MetaTCD: "2m24s",
// 	}
// 	firstStatMetrics4 := map[string]string{
// 		utils.MetaTCC: "1.35346",
// 		utils.MetaTCD: "2m24s",
// 	}
// 	secondStatMetrics1 := map[string]string{
// 		utils.MetaTCC: "0.6",
// 		utils.MetaTCD: "35s",
// 	}
// 	secondStatMetrics2 := map[string]string{
// 		utils.MetaTCC: "0.6",
// 		utils.MetaTCD: "37s",
// 	}

// 	if err := tutorial2CallsRpc.Call(utils.StatSv1GetQueueStringMetrics,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats"}, &metrics); err != nil {
// 		t.Fatal(err)
// 	}
// 	if !reflect.DeepEqual(firstStatMetrics1, metrics) &&
// 		!reflect.DeepEqual(firstStatMetrics2, metrics) &&
// 		!reflect.DeepEqual(firstStatMetrics3, metrics) &&
// 		!reflect.DeepEqual(firstStatMetrics4, metrics) {
// 		t.Errorf("expecting: %+v, received reply: %s", firstStatMetrics1, metrics)
// 	}
// }

// func testTutorial2CallCheckResourceRelease(t *testing.T) {
// 	var rs *engine.Resources
// 	args := &utils.ArgRSv1ResourceUsage{
// 		UsageID: "OriginID2",
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "ResourceRelease",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 				utils.Subject:      "1001",
// 				utils.Destination:  "1002",
// 			},
// 		},
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
// 		t.Fatal(err)
// 	} else if len(*rs) != 1 {
// 		t.Fatalf("Resources: %+v", rs)
// 	}
// 	for _, r := range *rs {
// 		if r.ID == "ResGroup1" && len(r.Usages) != 0 {
// 			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
// 		}
// 	}
// }

// func testTutorial2CallCheckThreshold1001After(t *testing.T) {
// 	var td engine.Threshold
// 	if err := tutorial2CallsRpc.Call(utils.ThresholdSv1GetThreshold,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil &&
// 		err.Error() != utils.ErrNotFound.Error() {
// 		t.Error(err)
// 	}
// }

// func testTutorial2CallCheckThreshold1002After(t *testing.T) {
// 	var td engine.Threshold
// 	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1002", Hits: 4}
// 	if err := tutorial2CallsRpc.Call(utils.ThresholdSv1GetThreshold,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}, &td); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eTd.Tenant, td.Tenant) {
// 		t.Errorf("expecting: %+v, received: %+v", eTd.Tenant, td.Tenant)
// 	} else if !reflect.DeepEqual(eTd.ID, td.ID) {
// 		t.Errorf("expecting: %+v, received: %+v", eTd.ID, td.ID)
// 	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
// 		t.Errorf("expecting: %+v, received: %+v", eTd.Hits, td.Hits)
// 	}
// }

// func testTutorial2CallSyncSessions(t *testing.T) {
// 	var reply *[]*sessions.ExternalSession
// 	// activeSessions shouldn't be active
// 	if err := tutorial2CallsRpc.Call(utils.SessionSv1GetActiveSessions,
// 		nil, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
// 		t.Error("Got error on SessionSv1.GetActiveSessions: ", err)
// 	}
// 	// 1001 call 1002 stop the call after 12 seconds
// 	if err := engine.PjsuaCallUri(
// 		&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*"},
// 		"sip:1002@127.0.0.1", "sip:127.0.0.1:5080", 120*time.Second, 5076); err != nil {
// 		t.Fatal(err)
// 	}
// 	// 1001 call 1003 stop the call after 11 seconds
// 	if err := engine.PjsuaCallUri(
// 		&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*"},
// 		"sip:1003@127.0.0.1", "sip:127.0.0.1:5080", 120*time.Second, 5077); err != nil {
// 		t.Fatal(err)
// 	}
// 	time.Sleep(time.Second)
// 	// get active sessions
// 	if err := tutorial2CallsRpc.Call(utils.SessionSv1GetActiveSessions,
// 		nil, &reply); err != nil {
// 		t.Error("Got error on SessionSv1.GetActiveSessions: ", err.Error())
// 	} else if len(*reply) != 4 { // expect to have 4 sessions ( two for 1001 to 1003 *raw and *default and two from 1001 to 1002 *raw and *default)
// 		t.Errorf("expecting 4 active sessions, received: %+v", utils.ToJSON(reply))
// 	}
// 	//check if resource was allocated for 2 calls(1001->1002;1001->1003)
// 	var rs *engine.Resources
// 	args := &utils.ArgRSv1ResourceUsage{
// 		UsageID: "OriginID3",
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "AllocateResource",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 				utils.Subject:      "1001",
// 				utils.Destination:  "1002",
// 			},
// 		},
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
// 		t.Fatal(err)
// 	} else if len(*rs) != 1 {
// 		t.Fatalf("Resources: %+v", utils.ToJSON(rs))
// 	}
// 	for _, r := range *rs {
// 		if r.ID == "ResGroup1" && len(r.Usages) != 2 {
// 			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
// 		}
// 	}

// 	time.Sleep(3 * time.Second)
// 	//stop the FS
// 	switch tutorial2OptConf {
// 	case utils.Freeswitch:
// 		engine.ForceKillProcName(utils.Freeswitch,
// 			int(tutorial2CallsCfg.SessionSCfg().ChannelSyncInterval.Nanoseconds()/1e6))
// 	default:
// 		t.Errorf("unsupported format")
// 	}

// 	time.Sleep(2 * time.Second)

// 	// activeSessions shouldn't be active
// 	if err := tutorial2CallsRpc.Call(utils.SessionSv1GetActiveSessions,
// 		nil, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
// 		t.Errorf("Got error on SessionSv1.GetActiveSessions: %v and reply: %s", err, utils.ToJSON(reply))
// 	}

// 	var sourceForCDR string
// 	numberOfCDR := 3
// 	switch tutorial2OptConf {
// 	case utils.Freeswitch:
// 		sourceForCDR = "FS_CHANNEL_ANSWER"
// 	}
// 	// verify cdr
// 	var rplCdrs []*engine.ExternalCDR
// 	req := utils.RPCCDRsFilter{
// 		Sources:  []string{sourceForCDR},
// 		MaxUsage: "20s",
// 		RunIDs:   []string{utils.MetaDefault},
// 		Accounts: []string{"1001"},
// 	}
// 	if err := tutorial2CallsRpc.Call(utils.APIerSv2GetCDRs, &req, &rplCdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(rplCdrs) != numberOfCDR { // cdr from sync session + cdr from before
// 		t.Fatal("Unexpected number of CDRs returned: ", len(rplCdrs), utils.ToJSON(rplCdrs))
// 	} else if time1, err := utils.ParseDurationWithSecs(rplCdrs[0].Usage); err != nil {
// 		t.Error(err)
// 	} else if time1 > 15*time.Second {
// 		t.Error("Unexpected time duration : ", time1)
// 	} else if time1, err := utils.ParseDurationWithSecs(rplCdrs[1].Usage); err != nil {
// 		t.Error(err)
// 	} else if time1 > 15*time.Second {
// 		t.Error("Unexpected time duration : ", time1)
// 	} else if numberOfCDR == 3 {
// 		if time1, err := utils.ParseDurationWithSecs(rplCdrs[2].Usage); err != nil {
// 			t.Error(err)
// 		} else if time1 > 15*time.Second {
// 			t.Error("Unexpected time duration : ", time1)
// 		}
// 	}

// 	//check if resource was released
// 	var rsAfter *engine.Resources
// 	if err := tutorial2CallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rsAfter); err != nil {
// 		t.Fatal(err)
// 	} else if len(*rsAfter) != 1 {
// 		t.Fatalf("Resources: %+v", rsAfter)
// 	}
// 	for _, r := range *rsAfter {
// 		if r.ID == "ResGroup1" && len(r.Usages) != 0 {
// 			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
// 		}
// 	}
// }

// func testTutorial2CallStopPjsuaListener(t *testing.T) {
// 	tutorial2CallsPjSuaListener.Write([]byte("q\n")) // Close pjsua
// 	time.Sleep(time.Second)                          // Allow pjsua to finish it's tasks, eg un-REGISTER
// }

// func testTutorial2CallStopCgrEngine(t *testing.T) {
// 	if err := engine.KillEngine(100); err != nil {
// 		t.Error(err)
// 	}
// }

// func testTutorial2CallStopFS(t *testing.T) {
// 	switch tutorial2OptConf {
// 	case utils.Freeswitch:
// 		engine.ForceKillProcName(utils.Freeswitch, 1000)
// 	default:
// 		t.Errorf("unsupported format")
// 	}
// }
