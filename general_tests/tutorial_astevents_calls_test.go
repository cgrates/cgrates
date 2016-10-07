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
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

const astPassword = "CGRateS.org"

var tutAstCallsCfg *config.CGRConfig
var tutAstCallsRpc *rpc.Client
var tutAstCallsPjSuaListener *os.File

func TestTutAstCallsInitCfg(t *testing.T) {
	if !*testCalls {
		return
	}
	// Init config first
	var err error
	tutAstCallsCfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "tutorials", "asterisk_ari", "cgrates", "etc", "cgrates"))
	if err != nil {
		t.Error(err)
	}
	tutAstCallsCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutAstCallsCfg)
}

// Remove data in both rating and accounting db
func TestTutAstCallsResetDataDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitDataDb(tutAstCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTutAstCallsResetStorDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitStorDb(tutAstCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// start Asterisk server
func TestTutAstCallsStartAsterisk(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("asterisk", 1000)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "asterisk_ari", "asterisk", "etc", "init.d", "asterisk"), "start", 2000); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutAstCallsStartEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("cgr-engine", *waitRater)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "asterisk_ari", "cgrates", "etc", "init.d", "cgrates"), "start", 100); err != nil {
		t.Fatal(err)
	}
}

/*
// Restart FS so we make sure reconnects are working
func TestTutAstCallsRestartAsterisk(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "asterisk_ari", "asterisk", "etc", "init.d", "asterisk"), "restart", 2000); err != nil {
		t.Fatal(err)
	}
}
*/
// Connect rpc client to rater
func TestTutAstCallsRpcConn(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	tutAstCallsRpc, err = jsonrpc.Dial("tcp", tutAstCallsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutAstCallsLoadTariffPlanFromFolder(t *testing.T) {
	if !*testCalls {
		return
	}
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutAstCallsRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Make sure account was debited properly
func TestTutAstCallsAccountsBefore(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1003"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1004"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1007"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 0.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1005"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err == nil || err.Error() != engine.ErrRedisNotFound.Error() {
		t.Error("Got error on ApierV2.GetAccount: %v", err)
	}
}

// Make sure all stats queues are in place
func TestTutAstCallsCdrStatsBefore(t *testing.T) {
	if !*testCalls {
		return
	}
	//eQueueIds := []string{"*default", "CDRST1", "CDRST_1001", "CDRST_1002", "CDRST_1003", "STATS_SUPPL1", "STATS_SUPPL2"}
	var statMetrics map[string]float64
	eMetrics := map[string]float64{engine.ACD: -1, engine.ASR: -1, engine.TCC: -1, engine.TCD: -1, engine.ACC: -1}
	if err := tutAstCallsRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST1"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.ACC: -1, engine.ACD: -1, engine.ASR: -1, engine.TCC: -1, engine.TCD: -1}
	if err := tutAstCallsRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1001"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.ACD: -1, engine.ASR: -1, engine.TCC: -1, engine.TCD: -1, engine.ACC: -1}
	if err := tutAstCallsRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1002"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.ACD: -1, engine.ASR: -1, engine.TCC: -1, engine.TCD: -1, engine.ACC: -1}
	if err := tutAstCallsRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1003"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.ACD: -1, engine.ASR: -1, engine.TCC: -1, engine.TCD: -1, engine.ACC: -1}
	if err := tutAstCallsRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "STATS_SUPPL1"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.ACD: -1, engine.ASR: -1, engine.TCC: -1, engine.TCD: -1, engine.ACC: -1}
	if err := tutAstCallsRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "STATS_SUPPL2"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
}

// Start Pjsua as listener and register it to receive calls
func TestTutAstCallsStartPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	acnts := []*engine.PjsuaAccount{
		&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: astPassword, Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1002@127.0.0.1", Username: "1002", Password: astPassword, Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1003@127.0.0.1", Username: "1003", Password: astPassword, Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1004@127.0.0.1", Username: "1004", Password: astPassword, Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1006@127.0.0.1", Username: "1006", Password: astPassword, Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1007@127.0.0.1", Username: "1007", Password: astPassword, Realm: "*", Registrar: "sip:127.0.0.1:5060"}}
	if tutAstCallsPjSuaListener, err = engine.StartPjsuaListener(acnts, 5070, time.Duration(500*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
}

// Call from 1001 (prepaid) to 1002
func TestTutAstCallsCall1001To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: astPassword, Realm: "*"}, "sip:1002@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(67)*time.Second, 5071); err != nil {
		t.Fatal(err)
	}
}

// Call from 1001 (prepaid) to 1003
func TestTutAstCallsCall1001To1003(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: astPassword, Realm: "*"}, "sip:1003@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(65)*time.Second, 5072); err != nil {
		t.Fatal(err)
	}
}

func TestTutAstCallsCall1002To1001(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1002@127.0.0.1", Username: "1002", Password: astPassword, Realm: "*"}, "sip:1001@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(61)*time.Second, 5073); err != nil {
		t.Fatal(err)
	}
}

func TestTutAstCallsCall1003To1001(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1003@127.0.0.1", Username: "1003", Password: astPassword, Realm: "*"}, "sip:1001@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(63)*time.Second, 5074); err != nil {
		t.Fatal(err)
	}
}

func TestTutAstCallsCall1004To1001(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1004@127.0.0.1", Username: "1004", Password: astPassword, Realm: "*"}, "sip:1001@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(62)*time.Second, 5075); err != nil {
		t.Fatal(err)
	}
}

func TestTutAstCallsCall1006To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1006@127.0.0.1", Username: "1006", Password: astPassword, Realm: "*"}, "sip:1002@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(64)*time.Second, 5076); err != nil {
		t.Fatal(err)
	}
}

func TestTutAstCallsCall1007To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1007@127.0.0.1", Username: "1007", Password: astPassword, Realm: "*"}, "sip:1002@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(66)*time.Second, 5077); err != nil {
		t.Fatal(err)
	}
}

// Make sure account was debited properly
func TestTutAstCallsAccount1001(t *testing.T) {
	if !*testCalls {
		return
	}
	time.Sleep(time.Duration(70) * time.Second) // Allow calls to finish before start querying the results
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() == 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	} else if reply.Disabled == true {
		t.Error("Account disabled")
	}
}

// Make sure account was debited properly
func TestTutAstCalls1001Cdrs(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCDR
	//var cgrId string // Share  with getCostDetails
	//var cCost engine.CallCost
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{"1001"}, DestinationPrefixes: []string{"1002"}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		//cgrId = reply[0].CGRID
		if reply[0].Source != "OSIPS_E_ACC_EVENT" {
			t.Errorf("Unexpected Source for CDR: %+v", reply[0])
		}
		if reply[0].RequestType != utils.META_PREPAID {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "67" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
		if reply[0].Cost == -1.0 { // Cost was not calculated
			t.Errorf("Unexpected Cost for CDR: %+v", reply[0])
		}
		//if reply[0].Supplier != "suppl2" { // Usage as seconds
		//	t.Errorf("Unexpected Supplier for CDR: %+v", reply[0])
		//}
	}
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{"1001"}, DestinationPrefixes: []string{"1003"}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		//cgrId = reply[0].CGRID
		if reply[0].RequestType != utils.META_PREPAID {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "65" && reply[0].Usage != "66" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
		if reply[0].Cost != 0 { // Cost was not calculated
			t.Errorf("Unexpected Cost for CDR: %+v", reply[0])
		}
	}
	req = utils.RPCCDRsFilter{Accounts: []string{"1001"}, RunIDs: []string{"derived_run1"}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].RequestType != utils.META_RATED {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0])
		}
		if reply[0].Subject != "1002" {
			t.Errorf("Unexpected Subject for CDR: %+v", reply[0])
		}
	}

}

// Make sure account was debited properly
func TestTutAstCalls1002Cdrs(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Accounts: []string{"1002"}, RunIDs: []string{utils.META_DEFAULT}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].Source != "OSIPS_E_ACC_EVENT" {
			t.Errorf("Unexpected Source for CDR: %+v", reply[0])
		}
		if reply[0].RequestType != utils.META_POSTPAID {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1001" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "61" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}
}

// Make sure account was debited properly
func TestTutAstCalls1003Cdrs(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Accounts: []string{"1003"}, RunIDs: []string{utils.META_DEFAULT}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].Source != "OSIPS_E_ACC_EVENT" {
			t.Errorf("Unexpected Source for CDR: %+v", reply[0])
		}
		if reply[0].RequestType != utils.META_PSEUDOPREPAID {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1001" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "63" && reply[0].Usage != "64" { // Usage as seconds, sometimes takes a second longer to disconnect
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}

}

// Make sure account was debited properly
func TestTutAstCalls1004Cdrs(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Accounts: []string{"1004"}, RunIDs: []string{utils.META_DEFAULT}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].Source != "OSIPS_E_ACC_EVENT" {
			t.Errorf("Unexpected Source for CDR: %+v", reply[0])
		}
		if reply[0].RequestType != utils.META_RATED {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1001" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "62" && reply[0].Usage != "63" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}

}

// Make sure aliasing was done for 1006 and we have no CDRs for it
func TestTutAstCalls1006Cdrs(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Accounts: []string{"1006"}, RunIDs: []string{utils.META_DEFAULT}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 0 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

// Make sure account was debited properly
func TestTutAstCalls1007Cdrs(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Accounts: []string{"1007"}, RunIDs: []string{utils.META_DEFAULT}}
	if err := tutAstCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].Source != "OSIPS_E_ACC_EVENT" {
			t.Errorf("Unexpected Source for CDR: %+v", reply[0])
		}
		if reply[0].RequestType != utils.META_PREPAID {
			t.Errorf("Unexpected RequestType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1002" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "66" && reply[0].Usage != "67" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
		if reply[0].Cost == -1.0 { // Cost was not calculated
			t.Errorf("Unexpected Cost for CDR: %+v", reply[0])
		}
	}
}

// Make sure account was debited properly
func TestTutAstCallsAccountFraud1001(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply string
	attrAddBlnc := &v1.AttrAddBalance{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Value: 101}
	if err := tutAstCallsRpc.Call("ApierV1.AddBalance", attrAddBlnc, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
}

// Based on Fraud automatic mitigation, our account should be disabled
func TestTutAstCallsAccountDisabled1001(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tutAstCallsRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.Disabled == false {
		t.Error("Account should be disabled per fraud detection rules.")
	}
}

func TestTutAstCallsStopPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}
	time.Sleep(70 * time.Second)                  // Give time for calls to go through before unregistering
	tutAstCallsPjSuaListener.Write([]byte("q\n")) // Close pjsua
	time.Sleep(time.Duration(1) * time.Second)    // Allow pjsua to finish it's tasks, eg un-REGISTER
}

func TestTutAstCallsStopCgrEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func TestTutAstCallsStopOpensips(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("opensips", 100)
}
