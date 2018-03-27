// +build newcall

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
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var tutFsCallsCfg *config.CGRConfig
var tutFsCallsRpc *rpc.Client
var tutFsCallsPjSuaListener *os.File
var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var fsConfig = flag.String("fsConfig", "/usr/share/cgrates/tutorials/fs_evsock", "FreeSwitch tutorial folder")

var sTestsCalls = []func(t *testing.T){
	testCallInitCfg,
	testCallResetDataDb,
	testCallResetStorDb,
	testCallStartFS,
	testCallStartEngine,
	testCallRestartFS,
	testCallRpcConn,
	testCallLoadTariffPlanFromFolder,
	testCallAccountsBefore,
	testCallStatMetricsBefore,
	testCallCheckResourceBeforeAllocation,
	testCallCheckThreshold1001Before,
	testCallCheckThreshold1002Before,
	testCallStartPjsuaListener,
	testCallCall1001To1002,
	testCallGetActiveSessions,
	testCallCall1002To1001,
	testCallCall1001To1003,
	testCallCheckResourceAllocation,
	testCallAccount1001,
	testCall1001Cdrs,
	testCall1002Cdrs,
	testCallStatMetrics,
	testCallCheckResourceRelease,
	testCallCheckThreshold1001After,
	testCallCheckThreshold1002After,
	testCallStopPjsuaListener,
	testCallStopCgrEngine,
	testCallStopFS,
}

//Test start here
func TestFSCall(t *testing.T) {
	for _, stest := range sTestsCalls {
		t.Run("", stest)
	}
}

func testCallInitCfg(t *testing.T) {
	// Init config first
	var err error
	tutFsCallsCfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "tutorials", "fs_evsock", "cgrates", "etc", "cgrates"))
	if err != nil {
		t.Error(err)
	}
	tutFsCallsCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutFsCallsCfg)
}

// Remove data in both rating and accounting db
func testCallResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(tutFsCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testCallResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tutFsCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// start FS server
func testCallStartFS(t *testing.T) {
	engine.KillProcName("freeswitch", 5000)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "fs_evsock", "freeswitch", "etc", "init.d", "freeswitch"), "start", 3000); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testCallStartEngine(t *testing.T) {
	engine.KillProcName("cgr-engine", *waitRater)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "fs_evsock", "cgrates", "etc", "init.d", "cgrates"), "start", 100); err != nil {
		t.Fatal(err)
	}
}

// Restart FS so we make sure reconnects are working
func testCallRestartFS(t *testing.T) {
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "fs_evsock", "freeswitch", "etc", "init.d", "freeswitch"), "restart", 5000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCallRpcConn(t *testing.T) {
	var err error
	tutFsCallsRpc, err = jsonrpc.Dial("tcp", tutFsCallsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Load the tariff plan, creating accounts and their balances
func testCallLoadTariffPlanFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial2")}
	if err := tutFsCallsRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Make sure account was debited properly
func testCallAccountsBefore(t *testing.T) {
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tutFsCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	var reply2 *engine.Account
	attrs2 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}
	if err := tutFsCallsRpc.Call("ApierV2.GetAccount", attrs2, &reply2); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply2.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply2.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func testCallStatMetricsBefore(t *testing.T) {
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaTCC: utils.NOT_AVAILABLE,
		utils.MetaTCD: utils.NOT_AVAILABLE,
	}
	if err := tutFsCallsRpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testCallCheckResourceBeforeAllocation(t *testing.T) {
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002"},
		}}
	if err := tutFsCallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", utils.ToJSON(rs))
	}
	for _, r := range *rs {
		if r.ID == "ResGroup1" &&
			(len(r.Usages) != 0 || len(r.TTLIdx) != 0) {
			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
		}
	}
}

func testCallCheckThreshold1001Before(t *testing.T) {
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1001", Hits: 0}
	if err := tutFsCallsRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testCallCheckThreshold1002Before(t *testing.T) {
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1002", Hits: 0}
	if err := tutFsCallsRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

// Start Pjsua as listener and register it to receive calls
func testCallStartPjsuaListener(t *testing.T) {
	var err error
	acnts := []*engine.PjsuaAccount{
		&engine.PjsuaAccount{Id: "sip:1001@192.168.56.202",
			Username: "1001", Password: "CGRateS.org", Realm: "*", Registrar: "sip:192.168.56.202:5060"},
		&engine.PjsuaAccount{Id: "sip:1002@192.168.56.202",
			Username: "1002", Password: "CGRateS.org", Realm: "*", Registrar: "sip:192.168.56.202:5060"}}
	if tutFsCallsPjSuaListener, err = engine.StartPjsuaListener(
		acnts, 5070, time.Duration(*waitRater)*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

// Call from 1001 (prepaid) to 1002
func testCallCall1001To1002(t *testing.T) {
	if err := engine.PjsuaCallUri(
		&engine.PjsuaAccount{Id: "sip:1001@192.168.56.202", Username: "1001", Password: "CGRateS.org", Realm: "*"},
		"sip:1002@192.168.56.202", "sip:192.168.56.202:5060", time.Duration(67)*time.Second, 5071); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
}

// GetActiveSessions
func testCallGetActiveSessions(t *testing.T) {
	var reply *[]*sessions.ActiveSession
	expected := &[]*sessions.ActiveSession{
		&sessions.ActiveSession{
			TOR:         "*voice",
			ReqType:     "*prepaid",
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
		},
	}
	if err := tutFsCallsRpc.Call("SessionSv1.GetActiveSessions",
		&map[string]string{}, &reply); err != nil {
		t.Error("Got error on SessionSv1.GetActiveSessions: ", err.Error())
	} else {
		// compare some fields (eg. CGRId is generated)
		if !reflect.DeepEqual((*expected)[0].TOR, (*reply)[0].TOR) {
			t.Errorf("Expected: %s, received: %s", (*expected)[0].TOR, (*reply)[0].TOR)
		} else if !reflect.DeepEqual((*expected)[0].ReqType, (*reply)[0].ReqType) {
			t.Errorf("Expected: %s, received: %s", (*expected)[0].ReqType, (*reply)[0].ReqType)
		} else if !reflect.DeepEqual((*expected)[0].Account, (*reply)[0].Account) {
			t.Errorf("Expected: %s, received: %s", (*expected)[0].Account, (*reply)[0].Account)
		} else if !reflect.DeepEqual((*expected)[0].Destination, (*reply)[0].Destination) {
			t.Errorf("Expected: %s, received: %s", (*expected)[0].Destination, (*reply)[0].Destination)
		}
	}
}

// Call from 1002 (postpaid) to 1001
func testCallCall1002To1001(t *testing.T) {
	if err := engine.PjsuaCallUri(
		&engine.PjsuaAccount{Id: "sip:1002@192.168.56.202", Username: "1002", Password: "CGRateS.org", Realm: "*"},
		"sip:1001@192.168.56.202", "sip:192.168.56.202:5060", time.Duration(65)*time.Second, 5072); err != nil {
		t.Fatal(err)
	}
}

// Call from 1001 (prepaid) to 1003 limit to 12 seconds
func testCallCall1001To1003(t *testing.T) {
	if err := engine.PjsuaCallUri(
		&engine.PjsuaAccount{Id: "sip:1001@192.168.56.202", Username: "1001", Password: "CGRateS.org", Realm: "*"},
		"sip:1003@192.168.56.202", "sip:192.168.56.202:5060", time.Duration(60)*time.Second, 5073); err != nil {
		t.Fatal(err)
	}
}

// Check if the resource was Allocated
func testCallCheckResourceAllocation(t *testing.T) {
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002"},
		}}
	if err := tutFsCallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", utils.ToJSON(rs))
	}
	for _, r := range *rs {
		if r.ID == "ResGroup1" &&
			(len(r.Usages) != 1 || len(r.TTLIdx) != 1) {
			t.Errorf("Unexpected resource: %+v", utils.ToJSON(r))
		}
	}
}

// Make sure account was debited properly
func testCallAccount1001(t *testing.T) {
	time.Sleep(time.Duration(80) * time.Second) // Allow calls to finish before start querying the results
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tutFsCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error(err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() == 10.0 { // Make sure we debitted
		t.Errorf("Expected: 10, received: %+v", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	} else if reply.Disabled == true {
		t.Error("Account disabled")
	}
}

// Make sure account was debited properly
func testCall1001Cdrs(t *testing.T) {
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{"1001"}}
	if err := tutFsCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		for _, cdr := range reply {
			if cdr.Source != "freeswitch_json" {
				t.Errorf("Unexpected Source for CDR: %+v", cdr.Source)
			}
			if cdr.RequestType != utils.META_PREPAID {
				t.Errorf("Unexpected RequestType for CDR: %+v", cdr.RequestType)
			}
			if cdr.Destination == "1002" {
				if cdr.Usage != "1m7s" && cdr.Usage != "1m8s" { // Usage as seconds
					t.Errorf("Unexpected Usage for CDR: %+v", cdr.Usage)
				}
				if cdr.CostSource != utils.MetaSessionS {
					t.Errorf("Unexpected CostSource for CDR: %+v", cdr.CostSource)
				}
			} else {
				if cdr.Usage != "12s" && cdr.Usage != "13s" { // Usage as seconds
					t.Errorf("Unexpected Usage for CDR: %+v", cdr.Usage)
				}
				if cdr.CostSource != utils.MetaSessionS {
					t.Errorf("Unexpected CostSource for CDR: %+v", cdr.CostSource)
				}
			}

		}

	}
}

// Make sure account was debited properly
func testCall1002Cdrs(t *testing.T) {
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{"1002"}, DestinationPrefixes: []string{"1001"}}
	if err := tutFsCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].Source != "freeswitch_json" {
			t.Errorf("Unexpected Source for CDR: %+v", reply[0].Source)
		}
		if reply[0].RequestType != utils.META_POSTPAID {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0].RequestType)
		}
		if reply[0].Usage != "1m5s" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0].Usage)
		}
		if reply[0].CostSource != utils.MetaCDRs {
			t.Errorf("Unexpected CostSource for CDR: %+v", reply[0].CostSource)
		}
	}
}

func testCallStatMetrics(t *testing.T) {
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaTCC: "0.91589",
		utils.MetaTCD: "2m12s",
	}
	if err := tutFsCallsRpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testCallCheckResourceRelease(t *testing.T) {
	var rs *engine.Resources
	args := &utils.ArgRSv1ResourceUsage{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002"},
		}}
	if err := tutFsCallsRpc.Call(utils.ResourceSv1GetResourcesForEvent, args, &rs); err != nil {
		t.Error(err)
	} else if len(*rs) != 2 {
		t.Errorf("Resources: %+v", rs)
	}
	for _, r := range *rs {
		if r.ID == "ResGroup1" &&
			(len(r.Usages) != 0 || len(r.TTLIdx) != 0) {
			t.Errorf("Unexpected resource: %+v", r)
		}
	}
}

func testCallCheckThreshold1001After(t *testing.T) {
	var td engine.Threshold
	if err := tutFsCallsRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testCallCheckThreshold1002After(t *testing.T) {
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1002", Hits: 4}
	if err := tutFsCallsRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1002"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Tenant, td.Tenant) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Tenant, td.Tenant)
	} else if !reflect.DeepEqual(eTd.ID, td.ID) {
		t.Errorf("expecting: %+v, received: %+v", eTd.ID, td.ID)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd.Hits, td.Hits)
	}
}

func testCallStopPjsuaListener(t *testing.T) {
	tutFsCallsPjSuaListener.Write([]byte("q\n")) // Close pjsua
	time.Sleep(time.Duration(1) * time.Second)   // Allow pjsua to finish it's tasks, eg un-REGISTER
}

func testCallStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testCallStopFS(t *testing.T) {
	engine.KillProcName("freeswitch", 1000)
}
