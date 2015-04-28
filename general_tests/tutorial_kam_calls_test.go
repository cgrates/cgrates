/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
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
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tutKamCallsCfg *config.CGRConfig
var tutKamCallsRpc *rpc.Client
var tutKamCallsPjSuaListener *os.File

func TestTutKamCallsInitCfg(t *testing.T) {
	if !*testCalls {
		return
	}
	// Init config first
	var err error
	tutKamCallsCfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "tutorials", "kamevapi", "cgrates", "etc", "cgrates"))
	if err != nil {
		t.Error(err)
	}
	tutKamCallsCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutKamCallsCfg)
}

// Remove data in both rating and accounting db
func TestTutKamCallsResetDataDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitDataDb(tutKamCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTutKamCallsResetStorDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitStorDb(tutKamCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// start Kam server
func TestTutKamCallsStartKam(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("kamailio", *waitRater)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "kamevapi", "kamailio", "etc", "init.d", "kamailio"), "start", 3000); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutKamCallsStartEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("cgr-engine", *waitRater)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "kamevapi", "cgrates", "etc", "init.d", "cgrates"), "start", 100); err != nil {
		t.Fatal(err)
	}
}

// Restart Kam so we make sure reconnects are working
func TestTutKamCallsRestartKam(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "kamevapi", "kamailio", "etc", "init.d", "kamailio"), "restart", 1000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutKamCallsRpcConn(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	tutKamCallsRpc, err = jsonrpc.Dial("tcp", tutKamCallsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutKamCallsLoadTariffPlanFromFolder(t *testing.T) {
	if !*testCalls {
		return
	}
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutKamCallsRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestTutKamCallsCacheStats(t *testing.T) {
	if !*testCalls {
		return
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 3, RatingPlans: 3, RatingProfiles: 3, Actions: 6, SharedGroups: 1, RatingAliases: 1, AccountAliases: 1, DerivedChargers: 1}
	var args utils.AttrCacheStats
	if err := tutKamCallsRpc.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

// Check items age
func TestTutKamCallsGetCachedItemAge(t *testing.T) {
	if !*testCalls {
		return
	}
	var rcvAge *utils.CachedItemAge
	if err := tutKamCallsRpc.Call("ApierV1.GetCachedItemAge", "1002", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.Destination > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutKamCallsRpc.Call("ApierV1.GetCachedItemAge", "RP_RETAIL1", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.RatingPlan > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutKamCallsRpc.Call("ApierV1.GetCachedItemAge", "*out:cgrates.org:call:*any", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.RatingProfile > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutKamCallsRpc.Call("ApierV1.GetCachedItemAge", "LOG_WARNING", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.Action > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutKamCallsRpc.Call("ApierV1.GetCachedItemAge", "SHARED_A", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.SharedGroup > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	/*
		if err := tutKamCallsRpc.Call("ApierV1.GetCachedItemAge", "1006", &rcvAge); err != nil {
			t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
		} else if rcvAge.RatingAlias > time.Duration(2)*time.Second {
			t.Errorf("Cache too old: %d", rcvAge)
		}
		if err := tutKamCallsRpc.Call("ApierV1.GetCachedItemAge", "1006", &rcvAge); err != nil {
			t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
		} else if rcvAge.RatingAlias > time.Duration(2)*time.Second || rcvAge.AccountAlias > time.Duration(2)*time.Second {
			t.Errorf("Cache too old: %d", rcvAge)
		}
	*/
}

// Make sure account was debited properly
func TestTutKamCallsAccountsBefore(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1003", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1004", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1007", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue() != 0.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1005", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err == nil || !strings.HasSuffix(err.Error(), "does not exist") {
		t.Error("Got error on ApierV1.GetAccount: %v", err)
	}
}

// Check call costs
func TestTutKamCallsGetCosts(t *testing.T) {
	if !*testCalls {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:00:20Z")
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	var cc engine.CallCost
	if err := tutKamCallsRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.6 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutKamCallsRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.6417 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:00:20Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1003",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutKamCallsRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1003",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutKamCallsRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1.3 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:00:20Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1004",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutKamCallsRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1004",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	if err := tutKamCallsRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1.3 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
}

func TestTutKamCallsCdrStats(t *testing.T) {
	if !*testCalls {
		return
	}
	var queueIds []string
	eQueueIds := []string{"*default", "CDRST1", "CDRST_1001", "CDRST_1002", "CDRST_1003"}
	if err := tutKamCallsRpc.Call("CDRStatsV1.GetQueueIds", "", &queueIds); err != nil {
		t.Error("Calling CDRStatsV1.GetQueueIds, got error: ", err.Error())
	} else if len(eQueueIds) != len(queueIds) {
		t.Errorf("Expecting: %v, received: %v", eQueueIds, queueIds)
	}
}

// Start Pjsua as listener and register it to receive calls
func TestTutKamCallsStartPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	acnts := []*engine.PjsuaAccount{
		&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1002@127.0.0.1", Username: "1002", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1003@127.0.0.1", Username: "1003", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1004@127.0.0.1", Username: "1004", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1006@127.0.0.1", Username: "1006", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1007@127.0.0.1", Username: "1007", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"}}
	if tutKamCallsPjSuaListener, err = engine.StartPjsuaListener(acnts, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Call from 1001 (prepaid) to 1002
func TestTutKamCallsCall1001To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*"}, "sip:1002@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(67)*time.Second, 5071); err != nil {
		t.Fatal(err)
	}
}

func TestTutKamCallsCall1002To1001(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1002@127.0.0.1", Username: "1002", Password: "CGRateS.org", Realm: "*"}, "sip:1001@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(61)*time.Second, 5072); err != nil {
		t.Fatal(err)
	}
}

func TestTutKamCallsCall1003To1001(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1003@127.0.0.1", Username: "1003", Password: "CGRateS.org", Realm: "*"}, "sip:1001@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(63)*time.Second, 5073); err != nil {
		t.Fatal(err)
	}
}

func TestTutKamCallsCall1004To1001(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1004@127.0.0.1", Username: "1004", Password: "CGRateS.org", Realm: "*"}, "sip:1001@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(62)*time.Second, 5074); err != nil {
		t.Fatal(err)
	}
}

func TestTutKamCallsCall1006To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1006@127.0.0.1", Username: "1006", Password: "CGRateS.org", Realm: "*"}, "sip:1002@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(64)*time.Second, 5075); err != nil {
		t.Fatal(err)
	}
}

func TestTutKamCallsCall1007To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1007@127.0.0.1", Username: "1007", Password: "CGRateS.org", Realm: "*"}, "sip:1002@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(66)*time.Second, 5076); err != nil {
		t.Fatal(err)
	}
}

// Make sure account was debited properly
func TestTutKamCallsAccount1001(t *testing.T) {
	if !*testCalls {
		return
	}
	time.Sleep(time.Duration(70) * time.Second) // Allow calls to finish before start querying the results
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue() == 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY+attrs.Direction].GetTotalValue())
	} else if reply.Disabled == true {
		t.Error("Account disabled")
	}
}

// Make sure account was debited properly
func TestTutKamCallsCdrs(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCdr
	req := utils.RpcCdrsFilter{Accounts: []string{"1001"}, RunIds: []string{utils.META_DEFAULT}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].CdrSource != "KAMAILIO_CGR_CALL_END" {
			t.Errorf("Unexpected CdrSource for CDR: %+v", reply[0])
		}
		if reply[0].ReqType != utils.META_PREPAID {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "67" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}
	req = utils.RpcCdrsFilter{Accounts: []string{"1001"}, RunIds: []string{"derived_run1"}, FilterOnDerived: true}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].ReqType != utils.META_RATED {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Subject != "1002" {
			t.Errorf("Unexpected Subject for CDR: %+v", reply[0])
		}
	}
	req = utils.RpcCdrsFilter{Accounts: []string{"1002"}, RunIds: []string{utils.META_DEFAULT}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].CdrSource != "KAMAILIO_CGR_CALL_END" {
			t.Errorf("Unexpected CdrSource for CDR: %+v", reply[0])
		}
		if reply[0].ReqType != utils.META_POSTPAID {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1001" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "61" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}
	req = utils.RpcCdrsFilter{Accounts: []string{"1003"}, RunIds: []string{utils.META_DEFAULT}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].CdrSource != "KAMAILIO_CGR_CALL_END" {
			t.Errorf("Unexpected CdrSource for CDR: %+v", reply[0])
		}
		if reply[0].ReqType != utils.META_PSEUDOPREPAID {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1001" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "63" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}
	req = utils.RpcCdrsFilter{Accounts: []string{"1004"}, RunIds: []string{utils.META_DEFAULT}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].CdrSource != "KAMAILIO_CGR_CALL_END" {
			t.Errorf("Unexpected CdrSource for CDR: %+v", reply[0])
		}
		if reply[0].ReqType != utils.META_RATED {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1001" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "62" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}
	req = utils.RpcCdrsFilter{Accounts: []string{"1006"}, RunIds: []string{utils.META_DEFAULT}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].CdrSource != "KAMAILIO_CGR_CALL_END" {
			t.Errorf("Unexpected CdrSource for CDR: %+v", reply[0])
		}
		if reply[0].ReqType != utils.META_PREPAID {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1002" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "64" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}
	req = utils.RpcCdrsFilter{Accounts: []string{"1007"}, RunIds: []string{utils.META_DEFAULT}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].CdrSource != "KAMAILIO_CGR_CALL_END" {
			t.Errorf("Unexpected CdrSource for CDR: %+v", reply[0])
		}
		if reply[0].ReqType != utils.META_PREPAID {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Destination != "1002" {
			t.Errorf("Unexpected Destination for CDR: %+v", reply[0])
		}
		if reply[0].Usage != "66" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0])
		}
	}
}

// Make sure account was debited properly
func TestTutKamCallsAccountFraud1001(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply string
	attrAddBlnc := &v1.AttrAddBalance{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Direction: "*out", Value: 101}
	if err := tutKamCallsRpc.Call("ApierV1.AddBalance", attrAddBlnc, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
}

// Based on Fraud automatic mitigation, our account should be disabled
func TestTutKamCallsAccountDisabled1001(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.Disabled == false {
		t.Error("Account should be disabled per fraud detection rules.")
	}
}

func TestTutKamCallsStopPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}

	tutKamCallsPjSuaListener.Write([]byte("q\n")) // Close pjsua
	time.Sleep(time.Duration(1) * time.Second)    // Allow pjsua to finish it's tasks, eg un-REGISTER
}

func TestTutKamCallsStopCgrEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func TestTutKamCallsStopKam(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("kamailio", 1000)
}
