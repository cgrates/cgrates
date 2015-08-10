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
	//"os"
	"path"
	"reflect"
	//"strings"
	"testing"
	"time"

	//"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tutLocalCfgPath string
var tutFsLocalCfg *config.CGRConfig
var tutLocalRpc *rpc.Client

func TestTutLocalInitCfg(t *testing.T) {
	if !*testLocal {
		return
	}
	tutLocalCfgPath = path.Join(*dataDir, "conf", "samples", "tutlocal")
	// Init config first
	var err error
	tutFsLocalCfg, err = config.NewCGRConfigFromFolder(tutLocalCfgPath)
	if err != nil {
		t.Error(err)
	}
	tutFsLocalCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutFsLocalCfg)
}

// Remove data in both rating and accounting db
func TestTutLocalResetDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(tutFsLocalCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTutLocalResetStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitStorDb(tutFsLocalCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutLocalStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := engine.StopStartEngine(tutLocalCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutLocalRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	tutLocalRpc, err = jsonrpc.Dial("tcp", tutFsLocalCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutLocalLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutLocalRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestTutLocalCacheStats(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 4, RatingPlans: 3, RatingProfiles: 8, Actions: 7, SharedGroups: 1, RatingAliases: 1, AccountAliases: 1,
		DerivedChargers: 1, LcrProfiles: 5, CdrStats: 6, Users: 2}
	var args utils.AttrCacheStats
	if err := tutLocalRpc.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

// Check items age
func TestTutLocalGetCachedItemAge(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvAge *utils.CachedItemAge
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "1002", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.Destination > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "RP_RETAIL1", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.RatingPlan > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "*out:cgrates.org:call:*any", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.RatingProfile > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "LOG_WARNING", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.Action > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "SHARED_A", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.SharedGroup > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "*out:cgrates.org:call:1001:*any", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.SharedGroup > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "*out:cgrates.org:call:*any:*any", &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if rcvAge.SharedGroup > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}

	/*
		if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "1006", &rcvAge); err != nil {
			t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
		} else if rcvAge.RatingAlias > time.Duration(2)*time.Second {
			t.Errorf("Cache too old: %d", rcvAge)
		}
		if err := tutLocalRpc.Call("ApierV1.GetCachedItemAge", "1006", &rcvAge); err != nil {
			t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
		} else if rcvAge.RatingAlias > time.Duration(2)*time.Second || rcvAge.AccountAlias > time.Duration(2)*time.Second {
			t.Errorf("Cache too old: %d", rcvAge)
		}
	*/
}

func TestTutLocalGetUsers(t *testing.T) {
	if !*testLocal {
		return
	}
	var users engine.UserProfiles
	if err := tutLocalRpc.Call("UsersV1.GetUsers", engine.UserProfile{}, &users); err != nil {
		t.Error("Got error on UsersV1.GetUsers: ", err.Error())
	} else if len(users) != 2 {
		t.Error("Calling UsersV1.GetUsers got users:", len(users))
	}
}

// Check call costs
func TestTutLocalGetCosts(t *testing.T) {
	if !*testLocal {
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
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.6 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	// Make sure that the same cost is returned via users aliasing
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        utils.USERS,
		Subject:       utils.USERS,
		Account:       utils.USERS,
		Destination:   "1002",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
		ExtraFields:   map[string]string{"Uuid": "388539dfd4f5cefee8f488b78c6c244b9e19138e"},
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
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
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.6417 { // 0.01 first minute, 0.04 25 seconds with RT_20CNT
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
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
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
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
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
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
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
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1.3 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	tStart = time.Date(2014, 8, 4, 13, 0, 0, 0, time.UTC)
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1007",
		TimeStart:   tStart,
		TimeEnd:     tStart.Add(time.Duration(50) * time.Second),
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.5 {
		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1007",
		TimeStart:   tStart,
		TimeEnd:     tStart.Add(time.Duration(70) * time.Second),
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.62 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1002",
		Account:     "1002",
		Destination: "1007",
		TimeStart:   tStart,
		TimeEnd:     tStart.Add(time.Duration(50) * time.Second),
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.5 {
		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1002",
		Account:     "1002",
		Destination: "1007",
		TimeStart:   tStart,
		TimeEnd:     tStart.Add(time.Duration(70) * time.Second),
	}
	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.7 { // In case of *disconnect strategy, it will not be applied so we can go on negative costs
		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
	}
}

// Check call costs
func TestTutLocalMaxDebit(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart := time.Date(2014, 8, 4, 13, 0, 0, 0, time.UTC)
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
	}
	var cc engine.CallCost
	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() == 20 {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
	}
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1003",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(200) * time.Second),
	}
	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
		t.Error("Got error on Responder.MaxDebit: ", err.Error())
	} else if cc.GetDuration() == 200 {
		t.Errorf("Calling Responder.MaxDebit got duration: %v", cc.GetDuration())
	}
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1007",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(120) * time.Second),
	}
	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() == 120 {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
	}
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1004",
		Account:       "1004",
		Destination:   "1007",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(120) * time.Second),
	}
	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() != time.Duration(62)*time.Second { // We have as strategy *dsconnect
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
	}
	var maxTime float64
	if err := tutLocalRpc.Call("Responder.GetMaxSessionTime", cd, &maxTime); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if maxTime != 62000000000 { // We have as strategy *dsconnect
		t.Errorf("Calling Responder.GetMaxSessionTime got maxTime: %f", maxTime)
	}
}

// Check call costs
func TestTutLocalDerivedMaxSessionTime(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart := time.Date(2014, 8, 4, 13, 0, 0, 0, time.UTC)
	ev := engine.StoredCdr{
		CgrId:       utils.Sha1("testevent1", tStart.String()),
		TOR:         utils.VOICE,
		AccId:       "testevent1",
		CdrHost:     "127.0.0.1",
		ReqType:     utils.META_PREPAID,
		Direction:   utils.OUT,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1004",
		Subject:     "1004",
		Destination: "1007",
		SetupTime:   tStart,
		AnswerTime:  tStart,
		Usage:       time.Duration(120) * time.Second,
		Supplier:    "suppl1",
		Cost:        -1,
	}
	var maxTime float64
	if err := tutLocalRpc.Call("Responder.GetDerivedMaxSessionTime", ev, &maxTime); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if maxTime != 62000000000 { // We have as strategy *dsconnect
		t.Errorf("Calling Responder.GetMaxSessionTime got maxTime: %f", maxTime)
	}
}

// Check MaxUsage
func TestTutLocalMaxUsage(t *testing.T) {
	if !*testLocal {
		return
	}
	setupReq := &engine.UsageRecord{TOR: utils.VOICE, ReqType: utils.META_PREPAID, Direction: utils.OUT, Tenant: "cgrates.org", Category: "call",
		Account: "1003", Subject: "1003", Destination: "1001",
		SetupTime: "2014-08-04T13:00:00Z", Usage: "1",
	}
	var maxTime float64
	if err := tutLocalRpc.Call("ApierV2.GetMaxUsage", setupReq, &maxTime); err != nil {
		t.Error(err)
	} else if maxTime != 1 {
		t.Errorf("Calling ApierV2.MaxUsage got maxTime: %f", maxTime)
	}
}

// Check MaxUsage
func TestTutLocalDebitUsage(t *testing.T) {
	if !*testLocal {
		return
	}
	setupReq := &engine.UsageRecord{TOR: utils.VOICE, ReqType: utils.META_PREPAID, Direction: utils.OUT, Tenant: "cgrates.org", Category: "call",
		Account: "1003", Subject: "1003", Destination: "1001",
		AnswerTime: "2014-08-04T13:00:00Z", Usage: "1",
	}
	var reply string
	if err := tutLocalRpc.Call("ApierV2.DebitUsage", setupReq, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Calling ApierV2.DebitUsage reply: ", reply)
	}
}

// Test CDR from external sources
func TestTutLocalProcessExternalCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	cdr := &engine.ExternalCdr{TOR: utils.VOICE,
		AccId: "testextcdr1", CdrHost: "192.168.1.1", CdrSource: utils.UNIT_TEST, ReqType: utils.META_RATED, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1001", Supplier: "SUPPL1",
		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
		Usage: "1", Pdd: "7.0", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
	}
	var reply string
	if err := tutLocalRpc.Call("CdrsV2.ProcessExternalCdr", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

// Make sure queueids were created
func TestTutFsCallsCdrStats(t *testing.T) {
	if !*testLocal {
		return
	}
	var queueIds []string
	eQueueIds := []string{"CDRST1", "CDRST_1001", "CDRST_1002", "CDRST_1003", "STATS_SUPPL1", "STATS_SUPPL2"}
	if err := tutLocalRpc.Call("CDRStatsV1.GetQueueIds", "", &queueIds); err != nil {
		t.Error("Calling CDRStatsV1.GetQueueIds, got error: ", err.Error())
	} else if len(eQueueIds) != len(queueIds) {
		t.Errorf("Expecting: %v, received: %v", eQueueIds, queueIds)
	}
}

// Check LCR
func TestTutLocalLcrStatic(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_STATIC, StrategyParams: "suppl2;suppl1", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 0.6, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	var lcr engine.LCRCost
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1003",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_STATIC, StrategyParams: "suppl1;suppl2", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
}

func TestTutLocalLcrHighestCost(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1002",
		Account:     "1002",
		Destination: "1002",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_HIGHEST, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 0.6, Duration: 60 * time.Second},
		},
	}
	var lcr engine.LCRCost
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
}

func TestTutLocalLcrQos(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1002",
		Account:     "1002",
		Destination: "1003",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1, engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1, engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
		},
	}
	eStLcr2 := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1, engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1, engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
		},
	}
	var lcr engine.LCRCost
	// Since there is no real quality difference, the suppliers will come in random order here
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
	// Post some CDRs to influence stats
	testCdr1 := &engine.StoredCdr{CgrId: utils.Sha1("testcdr1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		TOR: utils.VOICE, AccId: "testcdr1", CdrHost: "192.168.1.1", CdrSource: "TEST_QOS_LCR", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(2) * time.Minute, Supplier: "suppl1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	testCdr2 := &engine.StoredCdr{CgrId: utils.Sha1("testcdr2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		TOR: utils.VOICE, AccId: "testcdr2", CdrHost: "192.168.1.1", CdrSource: "TEST_QOS_LCR", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1002", Subject: "1002", Destination: "1003",
		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(90) * time.Second, Supplier: "suppl2",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
	var reply string
	for _, cdr := range []*engine.StoredCdr{testCdr1, testCdr2} {
		if err := tutLocalRpc.Call("CdrsV2.ProcessCdr", cdr, &reply); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received: ", reply)
		}
	}
	// Based on stats, supplier1 should always be better since he has a higer ACD
	eStLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 90, engine.ACC: 0.325, engine.TCC: 0.325, engine.ASR: 100, engine.ACD: 90}},
		},
	}
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
		//} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
		//	t.Errorf("Expecting: %+v, %+v, received: %+v, %+v", eStLcr.SupplierCosts[0], eStLcr.SupplierCosts[1], lcr.SupplierCosts[0], lcr.SupplierCosts[1])
	}
	testCdr3 := &engine.StoredCdr{CgrId: utils.Sha1("testcdr3", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		TOR: utils.VOICE, AccId: "testcdr3", CdrHost: "192.168.1.1", CdrSource: "TEST_QOS_LCR", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1004",
		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(180) * time.Second, Supplier: "suppl2"}
	if err := tutLocalRpc.Call("CdrsV2.ProcessCdr", testCdr3, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	// Since ACD has considerably increased for supplier2, we should have it as first prio now
	eStLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 270, engine.ACC: 0.3625, engine.TCC: 0.725, engine.ASR: 100, engine.ACD: 135}},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
		},
	}
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
		//} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
		//	t.Errorf("Expecting: %+v, %+v, received: %+v, %+v", eStLcr.SupplierCosts[0], eStLcr.SupplierCosts[1], lcr.SupplierCosts[0], lcr.SupplierCosts[1])
	}
}

func TestTutLocalLcrQosThreshold(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1003",
		Account:     "1003",
		Destination: "1002",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eLcr := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "20;;;;2m;;;;;;;", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 0.6, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 270, engine.ACC: 0.3625, engine.TCC: 0.725, engine.ASR: 100, engine.ACD: 135}},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
		},
	}
	var lcr engine.LCRCost
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) {
		//	t.Errorf("Expecting: %+v, %+v received: %+v, %+v", eLcr.SupplierCosts[0], eLcr.SupplierCosts[1], lcr.SupplierCosts[0], lcr.SupplierCosts[1])
	}
	testCdr4 := &engine.StoredCdr{CgrId: utils.Sha1("testcdr4", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		TOR: utils.VOICE, AccId: "testcdr4", CdrHost: "192.168.1.1", CdrSource: "TEST_QOS_LCR", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1004",
		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(60) * time.Second, Supplier: "suppl2"}
	var reply string
	if err := tutLocalRpc.Call("CdrsV2.ProcessCdr", testCdr4, &reply); err != nil { // Should drop ACD under the 2m required by threshold,  removing suppl2 from lcr
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	eLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "20;;;;2m;;;;;;;", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
		},
	}
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) {
		//	t.Errorf("Expecting: %+v, received: %+v", eLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1003",
		Account:     "1003",
		Destination: "1004",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "40;;;;90s;;;;;;;", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 330, engine.ACC: 0.3416666667, engine.TCC: 1.025, engine.ASR: 100, engine.ACD: 110}},
		},
	}
	/*eLcr2 := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "40;;90s;;;;;;;", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 330, engine.ACC: 0.3416666667, engine.TCC: 1.025, engine.ASR: 100, engine.ACD: 110}},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
		},
	}
	*/
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eLcr2.SupplierCosts, lcr.SupplierCosts) {
		//	t.Errorf("Expecting: %+v, received: %+v", eLcr.SupplierCosts[1], lcr.SupplierCosts[1])
	}
	testCdr5 := &engine.StoredCdr{CgrId: utils.Sha1("testcdr5", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		TOR: utils.VOICE, AccId: "testcdr5", CdrHost: "192.168.1.1", CdrSource: "TEST_QOS_LCR", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1004",
		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(1) * time.Second, Supplier: "suppl2"}
	if err := tutLocalRpc.Call("CdrsV2.ProcessCdr", testCdr5, &reply); err != nil { // Should drop ACD under the 1m required by threshold,  removing suppl2 from lcr
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	eLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "40;;;;90s;;;;;;;", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
		},
	}
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) {
		//	t.Errorf("Expecting: %+v, received: %+v", eLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
}

func TestTutLocalLeastCost(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1005",
		Account:     "1005",
		Destination: "1002",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile2", Strategy: engine.LCR_STRATEGY_LOWEST, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile2:suppl3", Cost: 0.01, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile2:suppl1", Cost: 0.6, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile2:suppl2", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	var lcr engine.LCRCost
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[2], lcr.SupplierCosts[2])
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "1005",
		Account:     "1005",
		Destination: "1003",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	eStLcr = &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_LOWEST, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	eStLcr2 := &engine.LCRCost{
		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_LOWEST, StrategyParams: "", Weight: 10.0},
		SupplierCosts: []*engine.LCRSupplierCost{
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second},
			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
	}
}

/*
// Make sure all stats queues were updated
func TestTutLocalCdrStatsAfter(t *testing.T) {
	if !*testLocal {
		return
	}
	var statMetrics map[string]float64
	eMetrics := map[string]float64{engine.ACD: 90.2, engine.ASR: 100, engine.TCC: 1.675, engine.TCD: 451, engine.ACC: 0.335}
	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST1"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.ACC: 0.35, engine.ACD: 120, engine.ASR: 100, engine.TCC: 1.675, engine.TCD: 451}
	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1001"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.TCD: 451, engine.ACC: 0.325, engine.ACD: 90, engine.ASR: 100, engine.TCC: 1.675}
	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1002"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.TCC: 1.675, engine.TCD: 451, engine.ACC: 0.325, engine.ACD: 90, engine.ASR: 100}
	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1003"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.TCC: 0.7, engine.TCD: 240, engine.ACC: 0.35, engine.ACD: 120, engine.ASR: 100}
	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "STATS_SUPPL1"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
		//} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		//	t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
	eMetrics = map[string]float64{engine.TCD: 331, engine.ACC: 0.33125, engine.ACD: 82.75, engine.ASR: 100, engine.TCC: 1.325}
	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "STATS_SUPPL2"}, &statMetrics); err != nil {
		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
	}
}
*/

/*
func TestTutLocalStopCgrEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
*/
