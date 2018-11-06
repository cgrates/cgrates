// +build integration

// /*
// Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
// Copyright (C) ITsysCOM GmbH

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>
// */

package general_tests

// import (
// 	"io/ioutil"
// 	"net/rpc"
// 	"net/rpc/jsonrpc"
// 	"os"
// 	"path"
// 	"reflect"
// 	"testing"
// 	"time"

// 	"github.com/cgrates/cgrates/apier/v1"
// 	"github.com/cgrates/cgrates/apier/v2"
// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/utils"
// )

// var tutLocalCfgPath string
// var tutFsLocalCfg *config.CGRConfig
// var tutLocalRpc *rpc.Client
// var loadInst utils.LoadInstance // Share load information between tests

// func TestTutITInitCfg(t *testing.T) {
// 	tutLocalCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
// 	// Init config first
// 	var err error
// 	tutFsLocalCfg, err = config.NewCGRConfigFromFolder(tutLocalCfgPath)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	tutFsLocalCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
// 	config.SetCgrConfig(tutFsLocalCfg)
// }

// // Remove data in dataDB
// func TestTutITResetDataDb(t *testing.T) {
// 	if err := engine.InitDataDb(tutFsLocalCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Wipe out the cdr database
// func TestTutITResetStorDb(t *testing.T) {
// 	if err := engine.InitStorDb(tutFsLocalCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Start CGR Engine
// func TestTutITStartEngine(t *testing.T) {
// 	if _, err := engine.StopStartEngine(tutLocalCfgPath, *waitRater); err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Connect rpc client to rater
// func TestTutITRpcConn(t *testing.T) {
// 	var err error
// 	tutLocalRpc, err = jsonrpc.Dial("tcp", tutFsLocalCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// // Load the tariff plan, creating accounts and their balances
// func TestTutITLoadTariffPlanFromFolder(t *testing.T) {
// 	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
// 	if err := tutLocalRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(100*time.Millisecond + time.Duration(*waitRater)*time.Millisecond) // Give time for scheduler to execute topups
// }

// // Check loaded stats
// func TestTutITCacheStats(t *testing.T) {
// 	var reply string
// 	if err := tutLocalRpc.Call("ApierV1.LoadCache", utils.AttrReloadCache{}, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != "OK" {
// 		t.Error(reply)
// 	}
// 	var rcvStats *utils.CacheStats
// 	expectedStats := &utils.CacheStats{Destinations: 5, ReverseDestinations: 7, RatingPlans: 4, RatingProfiles: 10,
// 		Actions: 9, ActionPlans: 4, AccountActionPlans: 5, SharedGroups: 1, DerivedChargers: 1, LcrProfiles: 5,
// 		Aliases: 1, ReverseAliases: 2, ResourceProfiles: 3, Resources: 3, StatQueues: 1, StatQueueProfiles: 1, Thresholds: 7,
// 		ThresholdProfiles: 7, Filters: 16, SupplierProfiles: 3, AttributeProfiles: 1,
// 		CdrStats: 0, Users: 0} // CdrStats and Users are 0 because deprecated. To be removed
// 	var args utils.AttrCacheStats
// 	if err := tutLocalRpc.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
// 		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
// 	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
// 		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
// 	}
// 	expKeys := utils.ArgsCache{
// 		DestinationIDs: &[]string{"DST_1003", "DST_1002", "DST_DE_MOBILE", "DST_1007", "DST_FS"},
// 		RatingPlanIDs:  &[]string{"RP_RETAIL1", "RP_GENERIC"},
// 	}
// 	var rcvKeys utils.ArgsCache
// 	argsAPI := utils.ArgsCacheKeys{ArgsCache: utils.ArgsCache{
// 		DestinationIDs: &[]string{}, RatingPlanIDs: &[]string{"RP_RETAIL1", "RP_GENERIC", "NONEXISTENT"}}}
// 	if err := tutLocalRpc.Call("ApierV1.GetCacheKeys", argsAPI, &rcvKeys); err != nil {
// 		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
// 	} else {
// 		if rcvKeys.DestinationIDs == nil {
// 			t.Errorf("Expecting rcvKeys.DestinationIDs to not be nil")
// 			// rcvKeys.DestinationIDs shoud not be nil so exit function
// 			// to avoid nil segmentation fault;
// 			// if this happens try to run this test manualy
// 			return
// 		}
// 		if len(*expKeys.DestinationIDs) != len(*rcvKeys.DestinationIDs) {
// 			t.Errorf("Expected: %+v, received: %+v", expKeys.DestinationIDs, rcvKeys.DestinationIDs)
// 		}
// 		if !reflect.DeepEqual(*expKeys.RatingPlanIDs, *rcvKeys.RatingPlanIDs) {
// 			t.Errorf("Expected: %+v, received: %+v", expKeys.RatingPlanIDs, rcvKeys.RatingPlanIDs)
// 		}
// 	}
// 	if _, err := engine.StopStartEngine(tutLocalCfgPath, 1500); err != nil {
// 		t.Fatal(err)
// 	}
// 	var err error
// 	tutLocalRpc, err = jsonrpc.Dial("tcp", tutFsLocalCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := tutLocalRpc.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
// 		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
// 	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
// 		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
// 	}
// }

// // Deprecated
// // func TestTutITGetUsers(t *testing.T) {
// // 	var users engine.UserProfiles
// // 	if err := tutLocalRpc.Call("UsersV1.GetUsers", engine.UserProfile{}, &users); err != nil {
// // 		t.Error("Got error on UsersV1.GetUsers: ", err.Error())
// // 	} else if len(users) != 3 {
// // 		t.Error("Calling UsersV1.GetUsers got users:", len(users))
// // 	}
// // }

// // func TestTutITGetMatchingAlias(t *testing.T) {
// // 	args := engine.AttrMatchingAlias{
// // 		Destination: "1005",
// // 		Direction:   "*out",
// // 		Tenant:      "cgrates.org",
// // 		Category:    "call",
// // 		Account:     "1006",
// // 		Subject:     "1006",
// // 		Context:     utils.MetaRating,
// // 		Target:      "Account",
// // 		Original:    "1006",
// // 	}
// // 	eMatchingAlias := "1002"
// // 	var matchingAlias string
// // 	if err := tutLocalRpc.Call("AliasesV1.GetMatchingAlias", args, &matchingAlias); err != nil {
// // 		t.Error(err)
// // 	} else if matchingAlias != eMatchingAlias {
// // 		t.Errorf("Expecting: %s, received: %+v", eMatchingAlias, matchingAlias)
// // 	}
// // }

// // Check call costs
// func TestTutITGetCosts(t *testing.T) {
// 	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ := utils.ParseDate("2014-08-04T13:00:20Z")
// 	cd := engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1002",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tEnd,
// 	}
// 	var cc engine.CallCost
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.6 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	// Make sure that the same cost is returned via users aliasing
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        utils.USERS,
// 		Subject:       utils.USERS,
// 		Account:       utils.USERS,
// 		Destination:   "1002",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tEnd,
// 		ExtraFields:   map[string]string{"Uuid": "388539dfd4f5cefee8f488b78c6c244b9e19138e"},
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.6 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1002",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tEnd,
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.6418 { // 0.01 first minute, 0.04 25 seconds with RT_20CNT
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ = utils.ParseDate("2014-08-04T13:00:20Z")
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1003",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tEnd,
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 1 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1003",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tEnd,
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 1.3 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ = utils.ParseDate("2014-08-04T13:00:20Z")
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1004",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tEnd,
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 1 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	tStart, _ = utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ = utils.ParseDate("2014-08-04T13:01:25Z")
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1004",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tEnd,
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 1.3 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	tStart = time.Date(2014, 8, 4, 13, 0, 0, 0, time.UTC)
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "1007",
// 		TimeStart:   tStart,
// 		TimeEnd:     tStart.Add(time.Duration(50) * time.Second),
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.5 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "1007",
// 		TimeStart:   tStart,
// 		TimeEnd:     tStart.Add(time.Duration(70) * time.Second),
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.62 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1002",
// 		Account:     "1002",
// 		Destination: "1007",
// 		TimeStart:   tStart,
// 		TimeEnd:     tStart.Add(time.Duration(50) * time.Second),
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.5 {
// 		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1002",
// 		Account:     "1002",
// 		Destination: "1007",
// 		TimeStart:   tStart,
// 		TimeEnd:     tStart.Add(time.Duration(70) * time.Second),
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.7 { // In case of *disconnect strategy, it will not be applied so we can go on negative costs
// 		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "1004",
// 		TimeStart:   time.Date(2016, 1, 6, 19, 0, 0, 0, time.UTC),
// 		TimeEnd:     time.Date(2016, 1, 6, 19, 1, 30, 0, time.UTC),
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.3249 { //
// 		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "1004",
// 		TimeStart:   time.Date(2016, 1, 6, 18, 31, 5, 0, time.UTC),
// 		TimeEnd:     time.Date(2016, 1, 6, 18, 32, 35, 0, time.UTC),
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 1.3 { //
// 		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "1002",
// 		TimeStart:   time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
// 		TimeEnd:     time.Date(2014, 12, 7, 8, 44, 26, 0, time.UTC),
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetCost", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.Cost != 0.3498 { //
// 		t.Errorf("Calling Responder.GetCost got callcost: %s", cc.AsJSON())
// 	}
// }

// // Check call costs
// func TestTutITMaxDebit(t *testing.T) {
// 	tStart := time.Date(2014, 8, 4, 13, 0, 0, 0, time.UTC)
// 	cd := engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1002",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
// 	}
// 	var cc engine.CallCost
// 	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.GetDuration() == 20 {
// 		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1003",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tStart.Add(time.Duration(200) * time.Second),
// 	}
// 	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.MaxDebit: ", err.Error())
// 	} else if cc.GetDuration() == 200 {
// 		t.Errorf("Calling Responder.MaxDebit got duration: %v", cc.GetDuration())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1001",
// 		Account:       "1001",
// 		Destination:   "1007",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tStart.Add(time.Duration(120) * time.Second),
// 	}
// 	cd.CgrID = "1"
// 	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.GetDuration() == 120 {
// 		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:     "*out",
// 		Category:      "call",
// 		Tenant:        "cgrates.org",
// 		Subject:       "1004",
// 		Account:       "1004",
// 		Destination:   "1007",
// 		DurationIndex: 0,
// 		TimeStart:     tStart,
// 		TimeEnd:       tStart.Add(time.Duration(120) * time.Second),
// 	}
// 	cd.CgrID = "2"
// 	if err := tutLocalRpc.Call("Responder.MaxDebit", cd, &cc); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if cc.GetDuration() != time.Duration(62)*time.Second { // We have as strategy *dsconnect
// 		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
// 	}
// 	var maxTime float64
// 	if err := tutLocalRpc.Call("Responder.GetMaxSessionTime", cd, &maxTime); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if maxTime != 62000000000 { // We have as strategy *dsconnect
// 		t.Errorf("Calling Responder.GetMaxSessionTime got maxTime: %f", maxTime)
// 	}
// }

// // Check call costs
// func TestTutITDerivedMaxSessionTime(t *testing.T) {
// 	tStart := time.Date(2014, 8, 4, 13, 0, 0, 0, time.UTC)
// 	ev := engine.CDR{
// 		CGRID:       utils.Sha1("testevent1", tStart.String()),
// 		ToR:         utils.VOICE,
// 		OriginID:    "testevent1",
// 		OriginHost:  "127.0.0.1",
// 		RequestType: utils.META_PREPAID,
// 		Tenant:      "cgrates.org",
// 		Category:    "call",
// 		Account:     "1004",
// 		Subject:     "1004",
// 		Destination: "1007",
// 		SetupTime:   tStart,
// 		AnswerTime:  tStart,
// 		Usage:       time.Duration(120) * time.Second,
// 		Cost:        -1,
// 	}
// 	var maxTime float64
// 	if err := tutLocalRpc.Call("Responder.GetDerivedMaxSessionTime", ev, &maxTime); err != nil {
// 		t.Error("Got error on Responder.GetCost: ", err.Error())
// 	} else if maxTime != 62000000000 { // We have as strategy *dsconnect
// 		t.Errorf("Calling Responder.GetMaxSessionTime got maxTime: %f", maxTime)
// 	}
// }

// // Check MaxUsage
// func TestTutITMaxUsage(t *testing.T) {
// 	setupReq := &engine.UsageRecord{ToR: utils.VOICE,
// 		RequestType: utils.META_PREPAID, Tenant: "cgrates.org", Category: "call",
// 		Account: "1003", Subject: "1003", Destination: "1001",
// 		SetupTime: "2014-08-04T13:00:00Z", Usage: "1s",
// 	}
// 	var maxTime float64
// 	if err := tutLocalRpc.Call("ApierV2.GetMaxUsage", setupReq, &maxTime); err != nil {
// 		t.Error(err)
// 	} else if maxTime != 1 {
// 		t.Errorf("Calling ApierV2.MaxUsage got maxTime: %f", maxTime)
// 	}
// 	setupReq = &engine.UsageRecord{ToR: utils.VOICE, RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call",
// 		Account: "test_max_usage", Destination: "1001",
// 		SetupTime: "2014-08-04T13:00:00Z",
// 	}
// 	if err := tutLocalRpc.Call("ApierV2.GetMaxUsage", setupReq, &maxTime); err != nil {
// 		t.Error(err)
// 	} else if maxTime != -1 {
// 		t.Errorf("Calling ApierV2.MaxUsage got maxTime: %f", maxTime)
// 	}
// }

// // Check DebitUsage
// func TestTutITDebitUsage(t *testing.T) {
// 	setupReq := &engine.UsageRecord{ToR: utils.VOICE, RequestType: utils.META_PREPAID, Tenant: "cgrates.org", Category: "call",
// 		Account: "1003", Subject: "1003", Destination: "1001",
// 		AnswerTime: "2014-08-04T13:00:00Z", Usage: "1",
// 	}
// 	var reply string
// 	if err := tutLocalRpc.Call("ApierV2.DebitUsage", setupReq, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != utils.OK {
// 		t.Error("Calling ApierV2.DebitUsage reply: ", reply)
// 	}
// }

// // Test CDR from external sources
// func TestTutITProcessExternalCdr(t *testing.T) {
// 	cdr := &engine.ExternalCDR{ToR: utils.VOICE,
// 		OriginID: "testextcdr1", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1001",
// 		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
// 		Usage: "1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
// 	}
// 	var reply string
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessExternalCdr", cdr, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// }

// // Test CDR involving UserProfile
// func TestTutITProcessExternalCdrUP(t *testing.T) {
// 	cdr := &engine.ExternalCDR{ToR: utils.VOICE,
// 		OriginID: "testextcdr2", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST,
// 		RequestType: utils.USERS, Tenant: utils.USERS, Account: utils.USERS, Destination: "1001",
// 		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z", Usage: "2s",
// 		ExtraFields: map[string]string{"Cli": "+4986517174964", "fieldextr2": "valextr2", "SysUserName": utils.USERS},
// 	}
// 	var reply string
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessExternalCdr", cdr, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
// 	eCdr := &engine.ExternalCDR{CGRID: "63a8d2bfeca2cfb790826c3ec461696d6574cfde", OrderID: 2,
// 		ToR:      utils.VOICE,
// 		OriginID: "testextcdr2", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1004", Subject: "1004", Destination: "1001",
// 		SetupTime:  time.Date(2014, 8, 4, 13, 0, 0, 0, time.UTC).Local().Format(time.RFC3339),
// 		AnswerTime: time.Date(2014, 8, 4, 13, 0, 7, 0, time.UTC).Local().Format(time.RFC3339), Usage: "2s",
// 		ExtraFields: map[string]string{"Cli": "+4986517174964", "fieldextr2": "valextr2", "SysUserName": "danb4"},
// 		RunID:       utils.DEFAULT_RUNID, Cost: 1}
// 	var cdrs []*engine.ExternalCDR
// 	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT},
// 		Accounts: []string{"1004"}, DestinationPrefixes: []string{"1001"}}
// 	if err := tutLocalRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].CGRID != eCdr.CGRID {
// 			t.Errorf("Unexpected CGRID for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].ToR != eCdr.ToR {
// 			t.Errorf("Unexpected TOR for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Source != eCdr.Source {
// 			t.Errorf("Unexpected Source for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].RequestType != eCdr.RequestType {
// 			t.Errorf("Unexpected RequestType for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Tenant != eCdr.Tenant {
// 			t.Errorf("Unexpected Tenant for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Category != eCdr.Category {
// 			t.Errorf("Unexpected Category for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Account != eCdr.Account {
// 			t.Errorf("Unexpected Account for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Subject != eCdr.Subject {
// 			t.Errorf("Unexpected Subject for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Destination != eCdr.Destination {
// 			t.Errorf("Unexpected Destination for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].SetupTime != eCdr.SetupTime {
// 			t.Errorf("Unexpected SetupTime for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].AnswerTime != eCdr.AnswerTime {
// 			t.Errorf("Unexpected AnswerTime for CDR: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Usage != eCdr.Usage {
// 			t.Errorf("Unexpected Usage for CDR: %+v", cdrs[0])
// 		}
// 	}
// }

// func TestTutITCostErrors(t *testing.T) {
// 	cdr := &engine.ExternalCDR{ToR: utils.VOICE,
// 		OriginID: "TestTutIT_1", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "fake", Account: "2001", Subject: "2001", Destination: "1001",
// 		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
// 		Usage: "1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
// 	}
// 	var reply string
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessExternalCdr", cdr, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
// 	var cdrs []*engine.ExternalCDR
// 	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{cdr.Account}, DestinationPrefixes: []string{cdr.Destination}}
// 	if err := tutLocalRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].OriginID != cdr.OriginID {
// 			t.Errorf("Unexpected OriginID for Cdr received: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Cost != -1 {
// 			t.Errorf("Unexpected Cost for Cdr received: %+v", cdrs[0])
// 		}
// 	}
// 	cdr2 := &engine.ExternalCDR{ToR: utils.VOICE,
// 		OriginID: "TestTutIT_2", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_POSTPAID,
// 		Tenant: "cgrates.org", Category: "fake", Account: "2002", Subject: "2002", Destination: "1001",
// 		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
// 		Usage: "1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
// 	}
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessExternalCdr", cdr2, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for CDR to be processed
// 	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{cdr2.Account}, DestinationPrefixes: []string{cdr2.Destination}}
// 	if err := tutLocalRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].OriginID != cdr2.OriginID {
// 			t.Errorf("Unexpected OriginID for Cdr received: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Cost != -1 {
// 			t.Errorf("Unexpected Cost for Cdr received: %+v", cdrs[0])
// 		}
// 	}
// 	cdr3 := &engine.ExternalCDR{ToR: utils.VOICE,
// 		OriginID: "TestTutIT_3", OriginHost: "192.168.1.1", Source: utils.UNIT_TEST, RequestType: utils.META_POSTPAID,
// 		Tenant: "cgrates.org", Category: "fake", Account: "1001", Subject: "1001", Destination: "2002",
// 		SetupTime: "2014-08-04T13:00:00Z", AnswerTime: "2014-08-04T13:00:07Z",
// 		Usage: "1", ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
// 	}
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessExternalCdr", cdr3, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for CDR to be processed
// 	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, Accounts: []string{cdr3.Account}, DestinationPrefixes: []string{cdr3.Destination}}
// 	if err := tutLocalRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].OriginID != cdr3.OriginID {
// 			t.Errorf("Unexpected OriginID for Cdr received: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Cost != -1 {
// 			t.Errorf("Unexpected Cost for Cdr received: %+v", cdrs[0])
// 		}
// 	}
// }

// // Make sure queueids were created
// func TestTutITCdrStats(t *testing.T) {
// 	var queueIds []string
// 	eQueueIds := []string{"CDRST1", "CDRST_1001", "CDRST_1002", "CDRST_1003", "STATS_SUPPL1", "STATS_SUPPL2"}
// 	if err := tutLocalRpc.Call("CDRStatsV1.GetQueueIds", "", &queueIds); err != nil {
// 		t.Error("Calling CDRStatsV1.GetQueueIds, got error: ", err.Error())
// 	} else if len(eQueueIds) != len(queueIds) {
// 		t.Errorf("Expecting: %v, received: %v", eQueueIds, queueIds)
// 	}
// }

// func TestTutITLeastCost(t *testing.T) {
// 	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
// 	cd := engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1005",
// 		Account:     "1005",
// 		Destination: "1002",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eStLcr := &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile2", Strategy: engine.LCR_STRATEGY_LOWEST, StrategyParams: "", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile2:suppl3", Cost: 0.01, Duration: 60 * time.Second},
// 			{Supplier: "*out:cgrates.org:lcr_profile2:suppl1", Cost: 0.6, Duration: 60 * time.Second},
// 			{Supplier: "*out:cgrates.org:lcr_profile2:suppl2", Cost: 1.2, Duration: 60 * time.Second},
// 		},
// 	}
// 	var lcr engine.LCRCost
// 	cd.CgrID = "10"
// 	cd.RunID = "10"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts, lcr.SupplierCosts)
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1005",
// 		Account:     "1005",
// 		Destination: "1003",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eStLcr = &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_LOWEST, StrategyParams: "", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second},
// 		},
// 	}
// 	eStLcr2 := &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_LOWEST, StrategyParams: "", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
// 		},
// 	}
// 	cd.CgrID = "11"
// 	cd.RunID = "11"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 	}
// }

// // Check LCR
// func TestTutITLcrStatic(t *testing.T) {
// 	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
// 	cd := engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "1002",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eStLcr := &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_STATIC, StrategyParams: "suppl2;suppl1", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 0.6, Duration: 60 * time.Second},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
// 		},
// 	}
// 	var lcr engine.LCRCost
// 	cd.CgrID = "1"
// 	cd.RunID = "1"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "1003",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eStLcr = &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_STATIC, StrategyParams: "suppl1;suppl2", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second},
// 		},
// 	}
// 	cd.CgrID = "2"
// 	cd.RunID = "2"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 	}
// }

// func TestTutITLcrHighestCost(t *testing.T) {
// 	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
// 	cd := engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1002",
// 		Account:     "1002",
// 		Destination: "1002",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eStLcr := &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_HIGHEST, StrategyParams: "", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 0.6, Duration: 60 * time.Second},
// 		},
// 	}
// 	var lcr engine.LCRCost
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 	}
// 	// LCR with Alias
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1006",
// 		Account:     "1006",
// 		Destination: "1002",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eStLcr.Entry), utils.ToJSON(lcr.Entry))
// 	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 	}
// }

// func TestTutITLcrQos(t *testing.T) {
// 	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
// 	cd := engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1002",
// 		Account:     "1002",
// 		Destination: "1003",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eStLcr := &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1",
// 			Strategy: engine.LCR_STRATEGY_QOS, StrategyParams: "", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1",
// 				Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1,
// 					engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2",
// 				Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1,
// 					engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
// 		},
// 	}
// 	/*
// 		eStLcr2 := &engine.LCRCost{
// 			Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1",
// 				Strategy: engine.LCR_STRATEGY_QOS, StrategyParams: "", Weight: 10.0},
// 			SupplierCosts: []*engine.LCRSupplierCost{
// 				&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2",
// 					Cost: 1.2, Duration: 60 * time.Second,
// 					QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1,
// 						engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
// 				&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1",
// 					Cost: 1.2, Duration: 60 * time.Second,
// 					QOS: map[string]float64{engine.TCD: -1, engine.ACC: -1, engine.TCC: -1,
// 						engine.ASR: -1, engine.ACD: -1, engine.DDC: -1}},
// 			},
// 		}
// 	*/
// 	var lcr engine.LCRCost
// 	// Since there is no real quality difference, the suppliers will come in random order here
// 	cd.CgrID = "3"
// 	cd.RunID = "3"
// 	/*
// 		if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 			t.Error(err)
// 		} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 			t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 		} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) &&
// 			!reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
// 			t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 		}
// 	*/
// 	// Post some CDRs to influence stats
// 	testCdr1 := &engine.CDR{CGRID: utils.Sha1("testcdr1",
// 		time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
// 		ToR: utils.VOICE, OriginID: "testcdr1", OriginHost: "192.168.1.1",
// 		Source: "TEST_QOS_LCR", RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1001",
// 		Subject: "1001", Destination: "1002",
// 		SetupTime:   time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC),
// 		AnswerTime:  time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
// 		Usage:       time.Duration(2) * time.Minute,
// 		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
// 	testCdr2 := &engine.CDR{CGRID: utils.Sha1("testcdr2",
// 		time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
// 		ToR: utils.VOICE, OriginID: "testcdr2", OriginHost: "192.168.1.1",
// 		Source: "TEST_QOS_LCR", RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1002",
// 		Subject: "1002", Destination: "1003",
// 		SetupTime:   time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC),
// 		AnswerTime:  time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
// 		Usage:       time.Duration(90) * time.Second,
// 		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
// 	var reply string
// 	for _, cdr := range []*engine.CDR{testCdr1, testCdr2} {
// 		if err := tutLocalRpc.Call("CdrsV1.ProcessCdr", cdr, &reply); err != nil {
// 			t.Error("Unexpected error: ", err.Error())
// 		} else if reply != utils.OK {
// 			t.Error("Unexpected reply received: ", reply)
// 		}
// 	}
// 	// Based on stats, supplier1 should always be better since he has a higer ACD
// 	eStLcr = &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY,
// 			RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS,
// 			StrategyParams: "", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{
// 				Supplier: "*out:cgrates.org:lcr_profile1:suppl1",
// 				Cost:     1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35,
// 					engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
// 			{
// 				Supplier: "*out:cgrates.org:lcr_profile1:suppl2",
// 				Cost:     1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 90, engine.ACC: 0.325,
// 					engine.TCC: 0.325, engine.ASR: 100, engine.ACD: 90}},
// 		},
// 	}
// 	cd.CgrID = "4"
// 	cd.RunID = "4"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 		//} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
// 		//	t.Errorf("Expecting: %+v, %+v, received: %+v, %+v", eStLcr.SupplierCosts[0], eStLcr.SupplierCosts[1], lcr.SupplierCosts[0], lcr.SupplierCosts[1])
// 	}
// 	testCdr3 := &engine.CDR{CGRID: utils.Sha1("testcdr3", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
// 		ToR: utils.VOICE, OriginID: "testcdr3", OriginHost: "192.168.1.1", Source: "TEST_QOS_LCR", RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1004",
// 		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
// 		Usage: time.Duration(180) * time.Second}
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessCdr", testCdr3, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	// Since ACD has considerably increased for supplier2, we should have it as first prio now
// 	eStLcr = &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS, StrategyParams: "", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 270, engine.ACC: 0.3625, engine.TCC: 0.725, engine.ASR: 100, engine.ACD: 135}},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
// 		},
// 	}
// 	cd.CgrID = "5"
// 	cd.RunID = "5"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
// 		//} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eStLcr2.SupplierCosts, lcr.SupplierCosts) {
// 		//	t.Errorf("Expecting: %+v, %+v, received: %+v, %+v", eStLcr.SupplierCosts[0], eStLcr.SupplierCosts[1], lcr.SupplierCosts[0], lcr.SupplierCosts[1])
// 	}
// }

// func TestTutITLcrQosThreshold(t *testing.T) {
// 	tStart, _ := utils.ParseDate("2014-08-04T13:00:00Z")
// 	tEnd, _ := utils.ParseDate("2014-08-04T13:01:00Z")
// 	cd := engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1003",
// 		Account:     "1003",
// 		Destination: "1002",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eLcr := &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "20;;;;2m;;;;;;;", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 0.6, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 270, engine.ACC: 0.3625, engine.TCC: 0.725, engine.ASR: 100, engine.ACD: 135}},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
// 		},
// 	}
// 	var lcr engine.LCRCost
// 	cd.CgrID = "6"
// 	cd.RunID = "6"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
// 		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) {
// 		//	t.Errorf("Expecting: %+v, %+v received: %+v, %+v", eLcr.SupplierCosts[0], eLcr.SupplierCosts[1], lcr.SupplierCosts[0], lcr.SupplierCosts[1])
// 	}
// 	testCdr4 := &engine.CDR{CGRID: utils.Sha1("testcdr4", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
// 		ToR: utils.VOICE, OriginID: "testcdr4", OriginHost: "192.168.1.1", Source: "TEST_QOS_LCR", RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1004",
// 		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
// 		Usage: time.Duration(60) * time.Second}
// 	var reply string
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessCdr", testCdr4, &reply); err != nil { // Should drop ACD under the 2m required by threshold,  removing suppl2 from lcr
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	eLcr = &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: "DST_1002", RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "20;;;;2m;;;;;;;", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
// 		},
// 	}
// 	cd.CgrID = "7"
// 	cd.RunID = "7"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
// 		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) {
// 		//	t.Errorf("Expecting: %+v, received: %+v", eLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 	}
// 	cd = engine.CallDescriptor{
// 		Direction:   "*out",
// 		Category:    "call",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1003",
// 		Account:     "1003",
// 		Destination: "1004",
// 		TimeStart:   tStart,
// 		TimeEnd:     tEnd,
// 	}
// 	eLcr = &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "40;;;;90s;;;;;;;", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 330, engine.ACC: 0.3416666667, engine.TCC: 1.025, engine.ASR: 100, engine.ACD: 110}},
// 		},
// 	}
// 	/*eLcr2 := &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "40;;90s;;;;;;;", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl2", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 330, engine.ACC: 0.3416666667, engine.TCC: 1.025, engine.ASR: 100, engine.ACD: 110}},
// 			&engine.LCRSupplierCost{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
// 		},
// 	}
// 	*/
// 	cd.CgrID = "8"
// 	cd.RunID = "8"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
// 		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) && !reflect.DeepEqual(eLcr2.SupplierCosts, lcr.SupplierCosts) {
// 		//	t.Errorf("Expecting: %+v, received: %+v", eLcr.SupplierCosts[1], lcr.SupplierCosts[1])
// 	}
// 	testCdr5 := &engine.CDR{CGRID: utils.Sha1("testcdr5", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
// 		ToR: utils.VOICE, OriginID: "testcdr5", OriginHost: "192.168.1.1", Source: "TEST_QOS_LCR", RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "1004",
// 		SetupTime: time.Date(2014, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2014, 12, 7, 8, 42, 26, 0, time.UTC),
// 		Usage: time.Duration(1) * time.Second}
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessCdr", testCdr5, &reply); err != nil { // Should drop ACD under the 1m required by threshold,  removing suppl2 from lcr
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	eLcr = &engine.LCRCost{
// 		Entry: &engine.LCREntry{DestinationId: utils.ANY, RPCategory: "lcr_profile1", Strategy: engine.LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "40;;;;90s;;;;;;;", Weight: 10.0},
// 		SupplierCosts: []*engine.LCRSupplierCost{
// 			{Supplier: "*out:cgrates.org:lcr_profile1:suppl1", Cost: 1.2, Duration: 60 * time.Second,
// 				QOS: map[string]float64{engine.TCD: 240, engine.ACC: 0.35, engine.TCC: 0.7, engine.ASR: 100, engine.ACD: 120}},
// 		},
// 	}
// 	cd.CgrID = "9"
// 	cd.RunID = "9"
// 	if err := tutLocalRpc.Call("Responder.GetLCR", cd, &lcr); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(eLcr.Entry, lcr.Entry) {
// 		t.Errorf("Expecting: %+v, received: %+v", eLcr.Entry, lcr.Entry)
// 		//} else if !reflect.DeepEqual(eLcr.SupplierCosts, lcr.SupplierCosts) {
// 		//	t.Errorf("Expecting: %+v, received: %+v", eLcr.SupplierCosts[0], lcr.SupplierCosts[0])
// 	}
// }

// // Test adding the account via API, using the data previously devined in .csv
// func TestTutITSetAccount(t *testing.T) {
// 	var reply string
// 	attrs := &v2.AttrSetAccount{Tenant: "cgrates.org", Account: "tutacnt1", ActionPlanIDs: &[]string{"PACKAGE_10"}, ActionTriggerIDs: &[]string{"STANDARD_TRIGGERS"}, ReloadScheduler: true}
// 	if err := tutLocalRpc.Call("ApierV2.SetAccount", attrs, &reply); err != nil {
// 		t.Error("Got error on ApierV2.SetAccount: ", err.Error())
// 	} else if reply != "OK" {
// 		t.Errorf("Calling ApierV2.SetAccount received: %s", reply)
// 	}
// 	type AttrGetAccounts struct {
// 		Tenant     string
// 		Direction  string
// 		AccountIds []string
// 		Offset     int // Set the item offset
// 		Limit      int // Limit number of items retrieved
// 	}
// 	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
// 	var acnts []*engine.Account
// 	if err := tutLocalRpc.Call("ApierV2.GetAccounts", utils.AttrGetAccounts{Tenant: attrs.Tenant, AccountIds: []string{attrs.Account}}, &acnts); err != nil {
// 		t.Error(err)
// 	} else if len(acnts) != 1 {
// 		t.Errorf("Accounts received: %+v", acnts)
// 	} else {
// 		acnt := acnts[0]
// 		dta, _ := utils.NewTAFromAccountKey(acnt.ID)
// 		if dta.Tenant != attrs.Tenant || dta.Account != attrs.Account {
// 			t.Error("Unexpected account id received: ", acnt.ID)
// 		}
// 		if balances := acnt.BalanceMap["*monetary"]; len(balances) != 1 {
// 			t.Errorf("Unexpected balances found: %+v", balances)
// 		}
// 		if len(acnt.ActionTriggers) != 4 {
// 			t.Errorf("Unexpected action triggers for account: %+v", acnt.ActionTriggers)
// 		}
// 		if acnt.AllowNegative {
// 			t.Error("AllowNegative should not be set")
// 		}
// 		if acnt.Disabled {
// 			t.Error("Disabled should not be set")
// 		}
// 	}
// 	attrs = &v2.AttrSetAccount{Tenant: "cgrates.org", Account: "tutacnt1", ActionPlanIDs: &[]string{"PACKAGE_10"}, ActionTriggerIDs: &[]string{"STANDARD_TRIGGERS"}, AllowNegative: utils.BoolPointer(true), Disabled: utils.BoolPointer(true), ReloadScheduler: true}

// 	if err := tutLocalRpc.Call("ApierV2.SetAccount", attrs, &reply); err != nil {
// 		t.Error("Got error on ApierV2.SetAccount: ", err.Error())
// 	} else if reply != "OK" {
// 		t.Errorf("Calling ApierV2.SetAccount received: %s", reply)
// 	}
// 	if err := tutLocalRpc.Call("ApierV2.GetAccounts", utils.AttrGetAccounts{Tenant: attrs.Tenant, AccountIds: []string{attrs.Account}}, &acnts); err != nil {
// 		t.Error(err)
// 	} else if len(acnts) != 1 {
// 		t.Errorf("Accounts received: %+v", acnts)
// 	} else {
// 		acnt := acnts[0]
// 		dta, _ := utils.NewTAFromAccountKey(acnt.ID)
// 		if dta.Tenant != attrs.Tenant || dta.Account != attrs.Account {
// 			t.Error("Unexpected account id received: ", acnt.ID)
// 		}
// 		if balances := acnt.BalanceMap["*monetary"]; len(balances) != 1 {
// 			t.Errorf("Unexpected balances found: %+v", balances)
// 		}
// 		if len(acnt.ActionTriggers) != 4 {
// 			t.Errorf("Unexpected action triggers for account: %+v", acnt.ActionTriggers)
// 		}
// 		if !acnt.AllowNegative {
// 			t.Error("AllowNegative should be set")
// 		}
// 		if !acnt.Disabled {
// 			t.Error("Disabled should be set")
// 		}
// 	}
// 	attrs = &v2.AttrSetAccount{Tenant: "cgrates.org", Account: "tutacnt1", ActionPlanIDs: &[]string{"PACKAGE_1001"}, ActionTriggerIDs: &[]string{"CDRST1_WARN"}, AllowNegative: utils.BoolPointer(true), Disabled: utils.BoolPointer(true), ReloadScheduler: true}

// 	if err := tutLocalRpc.Call("ApierV2.SetAccount", attrs, &reply); err != nil {
// 		t.Error("Got error on ApierV2.SetAccount: ", err.Error())
// 	} else if reply != "OK" {
// 		t.Errorf("Calling ApierV2.SetAccount received: %s", reply)
// 	}
// 	time.Sleep(100*time.Millisecond + time.Duration(*waitRater)*time.Millisecond) // Give time for scheduler to execute topups
// 	if err := tutLocalRpc.Call("ApierV2.GetAccounts", utils.AttrGetAccounts{Tenant: attrs.Tenant, AccountIds: []string{attrs.Account}}, &acnts); err != nil {
// 		t.Error(err)
// 	} else if len(acnts) != 1 {
// 		t.Errorf("Accounts received: %+v", acnts)
// 	} else {
// 		acnt := acnts[0]
// 		dta, _ := utils.NewTAFromAccountKey(acnt.ID)
// 		if dta.Tenant != attrs.Tenant || dta.Account != attrs.Account {
// 			t.Error("Unexpected account id received: ", acnt.ID)
// 		}
// 		if balances := acnt.BalanceMap["*monetary"]; len(balances) != 3 {
// 			t.Errorf("Unexpected balances found: %+v", balances)
// 		}
// 		if len(acnt.ActionTriggers) != 7 {
// 			t.Errorf("Unexpected action triggers for account: %+v", acnt.ActionTriggers)
// 		}
// 		if !acnt.AllowNegative {
// 			t.Error("AllowNegative should be set")
// 		}
// 		if !acnt.Disabled {
// 			t.Error("Disabled should be set")
// 		}
// 	}
// }

// /*
// // Make sure all stats queues were updated
// func TestTutITCdrStatsAfter(t *testing.T) {
// 	var statMetrics map[string]float64
// 	eMetrics := map[string]float64{engine.ACD: 90.2, engine.ASR: 100, engine.TCC: 1.675, engine.TCD: 451, engine.ACC: 0.335}
// 	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST1"}, &statMetrics); err != nil {
// 		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
// 	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
// 		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
// 	}
// 	eMetrics = map[string]float64{engine.ACC: 0.35, engine.ACD: 120, engine.ASR: 100, engine.TCC: 1.675, engine.TCD: 451}
// 	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1001"}, &statMetrics); err != nil {
// 		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
// 	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
// 		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
// 	}
// 	eMetrics = map[string]float64{engine.TCD: 451, engine.ACC: 0.325, engine.ACD: 90, engine.ASR: 100, engine.TCC: 1.675}
// 	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1002"}, &statMetrics); err != nil {
// 		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
// 	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
// 		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
// 	}
// 	eMetrics = map[string]float64{engine.TCC: 1.675, engine.TCD: 451, engine.ACC: 0.325, engine.ACD: 90, engine.ASR: 100}
// 	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "CDRST_1003"}, &statMetrics); err != nil {
// 		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
// 	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
// 		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
// 	}
// 	eMetrics = map[string]float64{engine.TCC: 0.7, engine.TCD: 240, engine.ACC: 0.35, engine.ACD: 120, engine.ASR: 100}
// 	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "STATS_SUPPL1"}, &statMetrics); err != nil {
// 		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
// 		//} else if !reflect.DeepEqual(eMetrics, statMetrics) {
// 		//	t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
// 	}
// 	eMetrics = map[string]float64{engine.TCD: 331, engine.ACC: 0.33125, engine.ACD: 82.75, engine.ASR: 100, engine.TCC: 1.325}
// 	if err := tutLocalRpc.Call("CDRStatsV1.GetMetrics", v1.AttrGetMetrics{StatsQueueId: "STATS_SUPPL2"}, &statMetrics); err != nil {
// 		t.Error("Calling CDRStatsV1.GetMetrics, got error: ", err.Error())
// 	} else if !reflect.DeepEqual(eMetrics, statMetrics) {
// 		t.Errorf("Expecting: %v, received: %v", eMetrics, statMetrics)
// 	}
// }
// */

// /* FixMe : In CallCost (Timespans) Increments is not populated so does not convert properly CallCost to Event

// func TestTutITPrepaidCDRWithSMCost(t *testing.T) {
// 	cdr := &engine.CDR{CGRID: utils.Sha1("testprepaid1", time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC).String()),
// 		ToR: utils.VOICE, OriginID: "testprepaid1", OriginHost: "192.168.1.1",
// 		Source: "TEST_PREPAID_CDR_SMCOST1", RequestType: utils.META_PREPAID, Tenant: "cgrates.org",
// 		RunID:    utils.META_DEFAULT,
// 		Category: "call", Account: "1001", Subject: "1001", Destination: "1003",
// 		SetupTime:   time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC),
// 		AnswerTime:  time.Date(2016, 4, 6, 13, 30, 0, 0, time.UTC),
// 		Usage:       time.Duration(90) * time.Second,
// 		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}

// 	cc := &engine.CallCost{
// 		Category:    "call",
// 		Account:     "1001",
// 		Subject:     "1001",
// 		Tenant:      "cgrates.org",
// 		Direction:   utils.OUT,
// 		Destination: "1003",
// 		Timespans: []*engine.TimeSpan{
// 			&engine.TimeSpan{
// 				TimeStart:     time.Date(2016, 4, 6, 13, 30, 0, 0, time.UTC),
// 				TimeEnd:       time.Date(2016, 4, 6, 13, 31, 30, 0, time.UTC),
// 				DurationIndex: 0,
// 				RateInterval: &engine.RateInterval{
// 					Rating: &engine.RIRate{
// 						Rates: engine.RateGroups{
// 							&engine.Rate{
// 								GroupIntervalStart: 0,
// 								Value:              0.01,
// 								RateIncrement:      10 * time.Second,
// 								RateUnit:           time.Second}}}},
// 			},
// 		},
// 		TOR: utils.VOICE}
// 	smCost := &engine.SMCost{CGRID: cdr.CGRID,
// 		RunID:       utils.META_DEFAULT,
// 		OriginHost:  cdr.OriginHost,
// 		OriginID:    cdr.OriginID,
// 		CostSource:  "TestTutITPrepaidCDRWithSMCost",
// 		Usage:       cdr.Usage,
// 		CostDetails: engine.NewEventCostFromCallCost(cc, cdr.CGRID, utils.META_DEFAULT),
// 	}
// 	var reply string
// 	if err := tutLocalRpc.Call("CdrsV1.StoreSMCost", &engine.AttrCDRSStoreSMCost{Cost: smCost}, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessCdr", cdr, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for CDR to be processed
// 	var cdrs []*engine.ExternalCDR
// 	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, CGRIDs: []string{cdr.CGRID}}
// 	if err := tutLocalRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].OriginID != cdr.OriginID {
// 			t.Errorf("Unexpected OriginID for Cdr received: %+v", cdrs[0].OriginID)
// 		}
// 		if cdrs[0].Cost != 0.9 {
// 			t.Errorf("Unexpected Cost for Cdr received: %+v", utils.ToJSON(cdrs[0].Cost))
// 		}
// 	}
// }
// */

// func TestTutITPrepaidCDRWithoutSMCost(t *testing.T) {
// 	cdr := &engine.CDR{CGRID: utils.Sha1("testprepaid2", time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC).String()),
// 		ToR: utils.VOICE, OriginID: "testprepaid2", OriginHost: "192.168.1.1", Source: "TEST_PREPAID_CDR_NO_SMCOST1", RequestType: utils.META_PREPAID,
// 		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1003",
// 		SetupTime: time.Date(2016, 4, 6, 13, 29, 24, 0, time.UTC), AnswerTime: time.Date(2016, 4, 6, 13, 30, 0, 0, time.UTC),
// 		Usage:       time.Duration(90) * time.Second,
// 		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
// 	var reply string
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessCdr", cdr, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	/*
// 		time.Sleep(time.Duration(7000) * time.Millisecond) // Give time for CDR to be processed
// 		var cdrs []*engine.ExternalCDR
// 		req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, CGRIDs: []string{cdr.CGRID}}
// 		if err := tutLocalRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
// 			t.Error("Unexpected error: ", err.Error())
// 		} else if len(cdrs) != 1 {
// 			t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 		} else {
// 			if cdrs[0].OriginID != cdr.OriginID {
// 				t.Errorf("Unexpected OriginID for Cdr received: %+v", cdrs[0])
// 			}
// 			if cdrs[0].Cost != 0.9 {
// 				t.Errorf("Unexpected Cost for Cdr received: %+v", cdrs[0])
// 			}
// 		}
// 	*/
// }

// func TestTutITExportCDR(t *testing.T) {
// 	cdr := &engine.CDR{ToR: utils.VOICE, OriginID: "testexportcdr1", OriginHost: "192.168.1.1", Source: "TestTutITExportCDR", RequestType: utils.META_RATED,
// 		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1003",
// 		SetupTime: time.Date(2016, 11, 30, 17, 5, 24, 0, time.UTC), AnswerTime: time.Date(2016, 11, 30, 17, 6, 4, 0, time.UTC),
// 		Usage:       time.Duration(98) * time.Second,
// 		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}
// 	cdr.ComputeCGRID()
// 	var reply string
// 	if err := tutLocalRpc.Call("CdrsV1.ProcessCdr", cdr, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	time.Sleep(time.Duration(50) * time.Millisecond) // Give time for CDR to be processed
// 	var cdrs []*engine.ExternalCDR
// 	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, CGRIDs: []string{cdr.CGRID}}
// 	if err := tutLocalRpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if len(cdrs) != 1 {
// 		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
// 	} else {
// 		if cdrs[0].OriginID != cdr.OriginID {
// 			t.Errorf("Unexpected OriginID for Cdr received: %+v", cdrs[0])
// 		}
// 		if cdrs[0].Cost != 1.3334 {
// 			t.Errorf("Unexpected Cost for Cdr received: %+v", cdrs[0])
// 		}
// 	}
// 	var replyExport v1.RplExportedCDRs
// 	exportArgs := v1.ArgExportCDRs{
// 		ExportPath:     utils.StringPointer("/tmp"),
// 		ExportFileName: utils.StringPointer("TestTutITExportCDR.csv"),
// 		ExportTemplate: utils.StringPointer("TestTutITExportCDR"),
// 		RPCCDRsFilter:  utils.RPCCDRsFilter{CGRIDs: []string{cdr.CGRID}, NotRunIDs: []string{utils.MetaRaw}}}
// 	if err := tutLocalRpc.Call("ApierV1.ExportCDRs", exportArgs, &replyExport); err != nil {
// 		t.Error(err)
// 	}
// 	eExportContent := `f0a92222a7d21b4d9f72744aabe82daef52e20d8,*default,testexportcdr1,*rated,cgrates.org,call,1001,1003,2016-11-30T18:06:04+01:00,98,1.33340,RETA
// f0a92222a7d21b4d9f72744aabe82daef52e20d8,derived_run1,testexportcdr1,*rated,cgrates.org,call,1001,1003,2016-11-30T18:06:04+01:00,98,1.33340,RETA
// `
// 	eExportContent2 := `f0a92222a7d21b4d9f72744aabe82daef52e20d8,derived_run1,testexportcdr1,*rated,cgrates.org,call,1001,1003,2016-11-30T18:06:04+01:00,98,1.33340,RETA
// f0a92222a7d21b4d9f72744aabe82daef52e20d8,*default,testexportcdr1,*rated,cgrates.org,call,1001,1003,2016-11-30T18:06:04+01:00,98,1.33340,RETA
// `
// 	expFilePath := path.Join(*exportArgs.ExportPath, *exportArgs.ExportFileName)
// 	if expContent, err := ioutil.ReadFile(expFilePath); err != nil {
// 		t.Error(err)
// 	} else if eExportContent != string(expContent) && eExportContent2 != string(expContent) { // CDRs are showing up randomly so we cannot predict order of export
// 		t.Errorf("Expecting: <%q> or <%q> received: <%q>", eExportContent, eExportContent2, string(expContent))
// 	}
// 	if err := os.Remove(expFilePath); err != nil {
// 		t.Error(err)
// 	}

// }

// func TestTutITStopCgrEngine(t *testing.T) {
// 	if err := engine.KillEngine(1000); err != nil {
// 		t.Error(err)
// 	}
// }
