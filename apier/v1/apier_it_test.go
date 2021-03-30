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
package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/streadway/amqp"
)

// ToDo: Replace rpc.Client with internal rpc server and Apier using internal map as both data and stor so we can run the tests non-local

/*
README:

 Enable local tests by passing '-local' to the go test command
 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data or passed via command arguments.
 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/data/storage/mysql/create_db_with_users.sql
 What these tests do:
  * Flush tables in storDb to start clean.
  * Start engine with default configuration and give it some time to listen (here caching can slow down, hence the command argument parameter).
  * Connect rpc client depending on encoding defined in configuration.
  * Execute remote Apis and test their replies(follow testtp scenario so we can test load in dataDb also).
*/
var (
	cfgPath           string
	cfg               *config.CGRConfig
	rater             *rpc.Client
	APIerSv1ConfigDIR string

	apierTests = []func(t *testing.T){
		testApierLoadConfig,
		testApierCreateDirs,
		testApierInitDataDb,
		testApierInitStorDb,
		testApierStartEngine,
		testApierRpcConn,
		testApierTPTiming,
		testApierTPDestination,
		testApierTPActions,

		testApierSetRatingProfileWithoutTenant,
		testApierRemoveRatingProfilesWithoutTenant,
		testApierReloadCache,
		testApierGetDestination,
		testApierExecuteAction,
		testApierExecuteActionWithoutTenant,
		testApierSetActions,
		testApierGetActions,
		testApierSetActionPlan,
		testApierResetDataBeforeLoadFromFolder,
		testApierLoadTariffPlanFromFolder,
		testApierComputeReverse,
		testApierResetDataAfterLoadFromFolder,
		testApierSetChargerS,
		testApierGetAccountAfterLoad,
		testApierResponderGetCost,
		testApierMaxDebitInexistentAcnt,
		testApierCdrServer,
		testApierITGetCdrs,
		testApierITProcessCdr,
		testApierGetCallCostLog,
		testApierITSetDestination,
		testApierITGetScheduledActions,
		testApierITGetDataCost,
		testApierITGetCost,
		testApierInitDataDb2,
		testApierInitStorDb2,
		testApierReloadCache2,
		testApierReloadScheduler2,
		testApierImportTPFromFolderPath,
		testApierLoadTariffPlanFromStorDbDryRun,
		testApierGetCacheStats2,
		testApierLoadTariffPlanFromStorDb,
		testApierStartStopServiceStatus,
		testApierReplayFldPosts,
		testApierGetDataDBVesions,
		testApierGetStorDBVesions,
		testApierBackwardsCompatible,
		testApierStopEngine,
	}
)

func TestApierIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow() // need tests redesign
	case utils.MetaMySQL:
		APIerSv1ConfigDIR = "apier_mysql"
	case utils.MetaMongo:
		APIerSv1ConfigDIR = "apier_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range apierTests {
		t.Run(APIerSv1ConfigDIR, stest)
	}
}

func testApierLoadConfig(t *testing.T) {
	var err error
	cfgPath = path.Join(*dataDir, "conf", "samples", APIerSv1ConfigDIR) // no need for a new config with *gob transport in this case
	if cfg, err = config.NewCGRConfigFromPath(cfgPath); err != nil {
		t.Error(err)
	}
}

func testApierCreateDirs(t *testing.T) {
	for _, pathDir := range []string{"/var/log/cgrates/ers/in", "/var/log/cgrates/ers/out"} {
		if err := os.RemoveAll(pathDir); err != nil {
			t.Fatal("Error removing folder: ", pathDir, err)
		}
		if err := os.MkdirAll(pathDir, 0755); err != nil {
			t.Fatal("Error creating folder: ", pathDir, err)
		}
	}
}

func testApierInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func testApierInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}
}

// Start engine
func testApierStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testApierRpcConn(t *testing.T) {
	var err error
	rater, err = newRPCClient(cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Test here TPTiming APIs
func testApierTPTiming(t *testing.T) {
	// ALWAYS,*any,*any,*any,*any,00:00:00
	tmAlways := &utils.ApierTPTiming{TPid: utils.TestSQL,
		ID:        "ALWAYS",
		Years:     utils.MetaAny,
		Months:    utils.MetaAny,
		MonthDays: utils.MetaAny,
		WeekDays:  utils.MetaAny,
		Time:      "00:00:00",
	}
	tmAlways2 := new(utils.ApierTPTiming)
	*tmAlways2 = *tmAlways
	tmAlways2.ID = "ALWAYS2"
	tmAsap := &utils.ApierTPTiming{
		TPid:      utils.TestSQL,
		ID:        "ASAP",
		Years:     utils.MetaAny,
		Months:    utils.MetaAny,
		MonthDays: utils.MetaAny,
		WeekDays:  utils.MetaAny,
		Time:      "*asap",
	}
	var reply string
	for _, tm := range []*utils.ApierTPTiming{tmAlways, tmAsap, tmAlways2} {
		if err := rater.Call(utils.APIerSv1SetTPTiming, &tm, &reply); err != nil {
			t.Error("Got error on APIerSv1.SetTPTiming: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received when calling APIerSv1.SetTPTiming: ", reply)
		}
	}
	// Check second set
	if err := rater.Call(utils.APIerSv1SetTPTiming, &tmAlways, &reply); err != nil {
		t.Error("Got error on second APIerSv1.SetTPTiming: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetTPTiming got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call(utils.APIerSv1SetTPTiming, new(utils.ApierTPTiming), &reply); err == nil {
		t.Error("Calling APIerSv1.SetTPTiming, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID Years Months MonthDays WeekDays Time]" {
		t.Error("Calling APIerSv1.SetTPTiming got unexpected error: ", err.Error())
	}
	// Test get
	var rplyTmAlways2 *utils.ApierTPTiming
	if err := rater.Call(utils.APIerSv1GetTPTiming, &AttrGetTPTiming{tmAlways2.TPid, tmAlways2.ID}, &rplyTmAlways2); err != nil {
		t.Error("Calling APIerSv1.GetTPTiming, got error: ", err.Error())
	} else if !reflect.DeepEqual(tmAlways2, rplyTmAlways2) {
		t.Errorf("Calling APIerSv1.GetTPTiming expected: %v, received: %v", tmAlways, rplyTmAlways2)
	}
	// Test remove
	if err := rater.Call(utils.APIerSv1RemoveTPTiming, AttrGetTPTiming{tmAlways2.TPid, tmAlways2.ID}, &reply); err != nil {
		t.Error("Calling APIerSv1.RemoveTPTiming, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.RemoveTPTiming received: ", reply)
	}
	// Test getIds
	var rplyTmIds []string
	expectedTmIds := []string{"ALWAYS", "ASAP"}
	if err := rater.Call(utils.APIerSv1GetTPTimingIds, &AttrGetTPTimingIds{tmAlways.TPid, utils.PaginatorWithSearch{}}, &rplyTmIds); err != nil {
		t.Error("Calling APIerSv1.GetTPTimingIds, got error: ", err.Error())
	}
	sort.Strings(expectedTmIds)
	sort.Strings(rplyTmIds)
	if !reflect.DeepEqual(expectedTmIds, rplyTmIds) {
		t.Errorf("Calling APIerSv1.GetTPTimingIds expected: %v, received: %v", expectedTmIds, rplyTmIds)
	}
}

// Test here TPTiming APIs
func testApierTPDestination(t *testing.T) {
	var reply string
	dstDe := &utils.TPDestination{TPid: utils.TestSQL, ID: "GERMANY", Prefixes: []string{"+49"}}
	dstDeMobile := &utils.TPDestination{TPid: utils.TestSQL, ID: "GERMANY_MOBILE", Prefixes: []string{"+4915", "+4916", "+4917"}}
	dstFs := &utils.TPDestination{TPid: utils.TestSQL, ID: "FS_USERS", Prefixes: []string{"10"}}
	dstDe2 := new(utils.TPDestination)
	*dstDe2 = *dstDe // Data which we use for remove, still keeping the sample data to check proper loading
	dstDe2.ID = "GERMANY2"
	for _, dst := range []*utils.TPDestination{dstDe, dstDeMobile, dstFs, dstDe2} {
		if err := rater.Call(utils.APIerSv1SetTPDestination, dst, &reply); err != nil {
			t.Error("Got error on APIerSv1.SetTPDestination: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received when calling APIerSv1.SetTPDestination: ", reply)
		}
	}
	// Check second set
	if err := rater.Call(utils.APIerSv1SetTPDestination, dstDe2, &reply); err != nil {
		t.Error("Got error on second APIerSv1.SetTPDestination: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetTPDestination got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call(utils.APIerSv1SetTPDestination, new(utils.TPDestination), &reply); err == nil {
		t.Error("Calling APIerSv1.SetTPDestination, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID Prefixes]" {
		t.Error("Calling APIerSv1.SetTPDestination got unexpected error: ", err.Error())
	}
	// Test get
	var rplyDstDe2 *utils.TPDestination
	if err := rater.Call(utils.APIerSv1GetTPDestination, &AttrGetTPDestination{dstDe2.TPid, dstDe2.ID}, &rplyDstDe2); err != nil {
		t.Error("Calling APIerSv1.GetTPDestination, got error: ", err.Error())
	} else if !reflect.DeepEqual(dstDe2, rplyDstDe2) {
		t.Errorf("Calling APIerSv1.GetTPDestination expected: %v, received: %v", dstDe2, rplyDstDe2)
	}
	// Test remove
	if err := rater.Call(utils.APIerSv1RemoveTPDestination, &AttrGetTPDestination{dstDe2.TPid, dstDe2.ID}, &reply); err != nil {
		t.Error("Calling APIerSv1.RemoveTPTiming, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.RemoveTPTiming received: ", reply)
	}
	// Test getIds
	var rplyDstIds []string
	expectedDstIds := []string{"FS_USERS", "GERMANY", "GERMANY_MOBILE"}
	if err := rater.Call(utils.APIerSv1GetTPDestinationIDs, &AttrGetTPDestinationIds{TPid: dstDe.TPid}, &rplyDstIds); err != nil {
		t.Error("Calling APIerSv1.GetTPDestinationIDs, got error: ", err.Error())
	}
	sort.Strings(expectedDstIds)
	sort.Strings(rplyDstIds)
	if !reflect.DeepEqual(expectedDstIds, rplyDstIds) {
		t.Errorf("Calling APIerSv1.GetTPDestinationIDs expected: %v, received: %v", expectedDstIds, rplyDstIds)
	}
}

func testApierTPActions(t *testing.T) {
	var reply string
	act := &utils.TPActions{TPid: utils.TestSQL,
		ID: "PREPAID_10", Actions: []*utils.TPAction{
			{Identifier: "*topup_reset", BalanceType: utils.MetaMonetary,
				Units: "10", ExpiryTime: "*unlimited",
				DestinationIds: utils.MetaAny, BalanceWeight: "10", Weight: 10},
		}}
	actWarn := &utils.TPActions{TPid: utils.TestSQL, ID: "WARN_VIA_HTTP", Actions: []*utils.TPAction{
		{Identifier: "*http_post", ExtraParameters: "http://localhost:8000", Weight: 10},
	}}
	actLog := &utils.TPActions{TPid: utils.TestSQL, ID: "LOG_BALANCE", Actions: []*utils.TPAction{
		{Identifier: "*log", Weight: 10},
	}}
	actTst := new(utils.TPActions)
	*actTst = *act
	actTst.ID = utils.TestSQL
	for _, ac := range []*utils.TPActions{act, actWarn, actTst, actLog} {
		if err := rater.Call(utils.APIerSv1SetTPActions, ac, &reply); err != nil {
			t.Error("Got error on APIerSv1.SetTPActions: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received when calling APIerSv1.SetTPActions: ", reply)
		}
	}
	// Check second set
	if err := rater.Call(utils.APIerSv1SetTPActions, actTst, &reply); err != nil {
		t.Error("Got error on second APIerSv1.SetTPActions: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetTPActions got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call(utils.APIerSv1SetTPActions, new(utils.TPActions), &reply); err == nil {
		t.Error("Calling APIerSv1.SetTPActions, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID Actions]" {
		t.Error("Calling APIerSv1.SetTPActions got unexpected error: ", err.Error())
	}
	// Test get
	var rplyActs *utils.TPActions
	if err := rater.Call(utils.APIerSv1GetTPActions, &AttrGetTPActions{TPid: actTst.TPid, ID: actTst.ID}, &rplyActs); err != nil {
		t.Error("Calling APIerSv1.GetTPActions, got error: ", err.Error())
	} else if !reflect.DeepEqual(actTst, rplyActs) {
		t.Errorf("Calling APIerSv1.GetTPActions expected: %v, received: %v", actTst, rplyActs)
	}
	// Test remove
	if err := rater.Call(utils.APIerSv1RemoveTPActions, &AttrGetTPActions{TPid: actTst.TPid, ID: actTst.ID}, &reply); err != nil {
		t.Error("Calling APIerSv1.RemoveTPActions, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.RemoveTPActions received: ", reply)
	}
	// Test getIds
	var rplyIds []string
	expectedIds := []string{"LOG_BALANCE", "PREPAID_10", "WARN_VIA_HTTP"}
	if err := rater.Call(utils.APIerSv1GetTPActionIds, &AttrGetTPActionIds{TPid: actTst.TPid}, &rplyIds); err != nil {
		t.Error("Calling APIerSv1.GetTPActionIds, got error: ", err.Error())
	}
	sort.Strings(expectedIds)
	sort.Strings(rplyIds)
	if !reflect.DeepEqual(expectedIds, rplyIds) {
		t.Errorf("Calling APIerSv1.GetTPActionIds expected: %v, received: %v", expectedIds, rplyIds)
	}
}

// Test here ReloadCache
func testApierReloadCache(t *testing.T) {
	var reply string
	arc := new(utils.AttrReloadCacheWithAPIOpts)
	// Simple test that command is executed without errors
	if err := rater.Call(utils.CacheSv1ReloadCache, arc, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expectedStats := engine.GetDefaultEmptyCacheStats()
	expectedStats[utils.CacheActions].Items = 1
	expectedStats[utils.CacheReverseDestinations].Items = 10
	expectedStats[utils.CacheLoadIDs].Items = 8
	expectedStats[utils.CacheRPCConnections].Items = 1
	if err := rater.Call(utils.CacheSv1GetCacheStats, new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling CacheSv1.GetCacheStats expected: %+v,\n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

// Test here GetDestination
func testApierGetDestination(t *testing.T) {
	reply := new(engine.Destination)
	dstId := "GERMANY_MOBILE"
	expectedReply := &engine.Destination{Id: dstId, Prefixes: []string{"+4915", "+4916", "+4917"}}
	if err := rater.Call(utils.APIerSv1GetDestination, &dstId, reply); err != nil {
		t.Error("Got error on APIerSv1.GetDestination: ", err.Error())
	} else if !reflect.DeepEqual(expectedReply, reply) {
		t.Errorf("Calling APIerSv1.GetDestination expected: %v, received: %v", expectedReply, reply)
	}
}

// Test here ExecuteAction
func testApierExecuteAction(t *testing.T) {
	var reply string
	// Add balance to a previously known account
	attrs := utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "dan2", ActionsId: "PREPAID_10"}
	if err := rater.Call(utils.APIerSv1ExecuteAction, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	reply2 := utils.EmptyString
	// Add balance to an account which does n exist
	attrs = utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "dan2", ActionsId: "DUMMY_ACTION"}
	if err := rater.Call(utils.APIerSv1ExecuteAction, attrs, &reply2); err == nil || reply2 == utils.OK {
		t.Error("Expecting error on APIerSv1.ExecuteAction.", err, reply2)
	}
}

func testApierExecuteActionWithoutTenant(t *testing.T) {
	var reply string
	// Add balance to a previously known account
	attrs := utils.AttrExecuteAction{Account: "dan2", ActionsId: "PREPAID_10"}
	if err := rater.Call(utils.APIerSv1ExecuteAction, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	reply2 := utils.EmptyString
	// Add balance to an account which does n exist
	attrs = utils.AttrExecuteAction{Account: "dan2", ActionsId: "DUMMY_ACTION"}
	if err := rater.Call(utils.APIerSv1ExecuteAction, attrs, &reply2); err == nil || reply2 == utils.OK {
		t.Error("Expecting error on APIerSv1.ExecuteAction.", err, reply2)
	}
}

func testApierSetActions(t *testing.T) {
	act1 := &V1TPAction{Identifier: utils.MetaTopUpReset, BalanceType: utils.MetaMonetary, Units: 75.0, ExpiryTime: utils.MetaUnlimited, Weight: 20.0}
	attrs1 := &V1AttrSetActions{ActionsId: "ACTS_1", Actions: []*V1TPAction{act1}}
	reply1 := utils.EmptyString
	if err := rater.Call(utils.APIerSv1SetActions, &attrs1, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActions: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Calling APIerSv1.SetActions received: %s", reply1)
	}
	// Calling the second time should raise EXISTS
	if err := rater.Call(utils.APIerSv1SetActions, &attrs1, &reply1); err == nil || err.Error() != "EXISTS" {
		t.Error("Unexpected result on duplication: ", err.Error())
	}
}

func testApierGetActions(t *testing.T) {
	expectActs := []*utils.TPAction{
		{Identifier: utils.MetaTopUpReset, BalanceType: utils.MetaMonetary,
			Units: "75", BalanceWeight: "0", BalanceBlocker: "false",
			BalanceDisabled: "false", ExpiryTime: utils.MetaUnlimited, Weight: 20.0}}

	var reply []*utils.TPAction
	if err := rater.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(expectActs, reply) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(expectActs), utils.ToJSON(reply))
	}
}

func testApierSetActionPlan(t *testing.T) {
	atm1 := &AttrActionPlan{ActionsId: "ACTS_1", MonthDays: "1", Time: "00:00:00", Weight: 20.0}
	atms1 := &AttrSetActionPlan{Id: "ATMS_1", ActionPlan: []*AttrActionPlan{atm1}}
	reply1 := utils.EmptyString
	if err := rater.Call(utils.APIerSv1SetActionPlan, &atms1, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply1)
	}
	// Calling the second time should raise EXISTS
	if err := rater.Call(utils.APIerSv1SetActionPlan, &atms1, &reply1); err == nil || err.Error() != "EXISTS" {
		t.Error("Unexpected result on duplication: ", err.Error())
	}
}

// Start fresh before loading from folder
func testApierResetDataBeforeLoadFromFolder(t *testing.T) {
	testApierInitDataDb(t)
	var reply string
	// Simple test that command is executed without errors
	if err := rater.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expectedStats := engine.GetDefaultEmptyCacheStats()
	err := rater.Call(utils.CacheSv1GetCacheStats, new(utils.AttrCacheIDsWithAPIOpts), &rcvStats)
	if err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(rcvStats, expectedStats) {
		t.Errorf("Calling CacheSv1.GetCacheStats expected: %v,  received: %v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

// Test here LoadTariffPlanFromFolder
func testApierLoadTariffPlanFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: utils.EmptyString}
	if err := rater.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err == nil || !strings.HasPrefix(err.Error(), utils.ErrMandatoryIeMissing.Error()) {
		t.Error(err)
	}
	attrs = &utils.AttrLoadTpFromFolder{FolderPath: "/INVALID/"}
	if err := rater.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err == nil || err.Error() != utils.ErrInvalidPath.Error() {
		t.Error(err)
	}
	// Simple test that command is executed without errors
	attrs = &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := rater.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

// For now just test that they execute without errors
func testApierComputeReverse(t *testing.T) {
	var reply string
	if err := rater.Call(utils.APIerSv1ComputeReverseDestinations, utils.EmptyString, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Received: ", reply)
	}
	if err := rater.Call(utils.APIerSv1ComputeAccountActionPlans, utils.EmptyString, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Received: ", reply)
	}
}

func testApierResetDataAfterLoadFromFolder(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	expStats[utils.CacheAccountActionPlans].Items = 3
	expStats[utils.CacheActionPlans].Items = 7
	expStats[utils.CacheActions].Items = 5
	expStats[utils.CacheDestinations].Items = 3
	expStats[utils.CacheLoadIDs].Items = 17
	expStats[utils.CacheRPCConnections].Items = 2
	if err := rater.Call(utils.CacheSv1GetCacheStats, new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
	var reply string
	// Simple test that command is executed without errors
	if err := rater.Call(utils.CacheSv1LoadCache, utils.NewAttrReloadCacheWithOpts(), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	expStats[utils.CacheActionTriggers].Items = 1
	expStats[utils.CacheActions].Items = 13
	expStats[utils.CacheAttributeProfiles].Items = 1
	expStats[utils.CacheFilters].Items = 15
	expStats[utils.CacheRatingPlans].Items = 5
	expStats[utils.CacheRatingProfiles].Items = 5
	expStats[utils.CacheResourceProfiles].Items = 3
	expStats[utils.CacheResources].Items = 3
	expStats[utils.CacheReverseDestinations].Items = 5
	expStats[utils.CacheStatQueueProfiles].Items = 1
	expStats[utils.CacheStatQueues].Items = 1
	expStats[utils.CacheRouteProfiles].Items = 2
	expStats[utils.CacheThresholdProfiles].Items = 1
	expStats[utils.CacheThresholds].Items = 1
	expStats[utils.CacheLoadIDs].Items = 33
	expStats[utils.CacheTimings].Items = 12
	expStats[utils.CacheThresholdFilterIndexes].Items = 5
	expStats[utils.CacheThresholdFilterIndexes].Groups = 1
	expStats[utils.CacheStatFilterIndexes].Items = 2
	expStats[utils.CacheStatFilterIndexes].Groups = 1
	expStats[utils.CacheRouteFilterIndexes].Items = 2
	expStats[utils.CacheRouteFilterIndexes].Groups = 1
	expStats[utils.CacheResourceFilterIndexes].Items = 5
	expStats[utils.CacheResourceFilterIndexes].Groups = 1
	expStats[utils.CacheAttributeFilterIndexes].Items = 4
	expStats[utils.CacheAttributeFilterIndexes].Groups = 1
	expStats[utils.CacheReverseFilterIndexes].Items = 10
	expStats[utils.CacheReverseFilterIndexes].Groups = 7

	if err := rater.Call(utils.CacheSv1GetCacheStats, new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testApierSetChargerS(t *testing.T) {
	//add a default charger
	chargerProfile := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result string
	if err := rater.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

// Make sure balance was topped-up
// Bug reported by DigiDaz over IRC
func testApierGetAccountAfterLoad(t *testing.T) {
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rater.Call(utils.APIerSv2GetAccount, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MetaMonetary].GetTotalValue() != 13 {
		t.Errorf("Calling APIerSv1.GetBalance expected: 13, received: %v \n\n for:%s", reply.BalanceMap[utils.MetaMonetary].GetTotalValue(), utils.ToJSON(reply))
	}
}

// Test here ResponderGetCost
func testApierResponderGetCost(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", utils.EmptyString)
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:30Z", utils.EmptyString)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			Account:       "1001",
			Destination:   "+4917621621391",
			DurationIndex: 90,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var cc engine.CallCost
	// Simple test that command is executed without errors
	if err := rater.Call(utils.ResponderGetCost, cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 90.0 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}
}

func testApierMaxDebitInexistentAcnt(t *testing.T) {
	cc := &engine.CallCost{}
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Tenant:      "cgrates.org",
			Category:    "call",
			Subject:     "INVALID",
			Account:     "INVALID",
			Destination: "1002",
			TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
			TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(10 * time.Second),
		},
	}
	if err := rater.Call(utils.ResponderMaxDebit, cd, cc); err == nil {
		t.Error(err.Error())
	}
	if err := rater.Call(utils.ResponderDebit, cd, cc); err == nil {
		t.Error(err.Error())
	}
}

func testApierCdrServer(t *testing.T) {
	httpClient := new(http.Client)
	cdrForm1 := url.Values{utils.OriginID: []string{"dsafdsaf"}, utils.OriginHost: []string{"192.168.1.1"}, utils.RequestType: []string{utils.MetaRated},
		utils.Tenant: []string{"cgrates.org"}, utils.Category: []string{"call"}, utils.AccountField: []string{"1001"}, utils.Subject: []string{"1001"}, utils.Destination: []string{"1002"},
		utils.SetupTime:  []string{"2013-11-07T08:42:22Z"},
		utils.AnswerTime: []string{"2013-11-07T08:42:26Z"}, utils.Usage: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	cdrForm2 := url.Values{utils.OriginID: []string{"adsafdsaf"}, utils.OriginHost: []string{"192.168.1.1"}, utils.RequestType: []string{utils.MetaRated},
		utils.Tenant: []string{"cgrates.org"}, utils.Category: []string{"call"}, utils.AccountField: []string{"1001"}, utils.Subject: []string{"1001"}, utils.Destination: []string{"1002"},
		utils.SetupTime:  []string{"2013-11-07T08:42:23Z"},
		utils.AnswerTime: []string{"2013-11-07T08:42:26Z"}, utils.Usage: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	for _, cdrForm := range []url.Values{cdrForm1, cdrForm2} {
		cdrForm.Set(utils.Source, utils.TestSQL)
		if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cdr_http", "127.0.0.1:2080"), cdrForm); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testApierITGetCdrs(t *testing.T) {
	var reply []*engine.ExternalCDR
	req := utils.AttrGetCdrs{MediationRunIds: []string{utils.MetaDefault}}
	if err := rater.Call(utils.APIerSv1GetCDRs, &req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func testApierITProcessCdr(t *testing.T) {
	var reply string
	cdr := &engine.CDRWithAPIOpts{
		CDR: &engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.MetaVoice, OriginID: "dsafdsaf",
			OriginHost: "192.168.1.1", Source: "test", RequestType: utils.MetaRated, Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001",
			Destination: "1002",
			SetupTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.MetaDefault,
			Usage: 10 * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
		},
	}
	if err := rater.Call(utils.CDRsV1ProcessCDR, cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	req := utils.AttrGetCdrs{MediationRunIds: []string{utils.MetaDefault}}
	if err := rater.Call(utils.APIerSv1GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 3 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

// Test here ResponderGetCost
func testApierGetCallCostLog(t *testing.T) {
	var cc engine.EventCost
	var attrs utils.AttrGetCallCost
	// Simple test that command is executed without errors
	if err := rater.Call(utils.APIerSv1GetEventCost, &attrs, &cc); err == nil {
		t.Error("Failed to detect missing fields in APIerSv1.GetCallCostLog")
	}
	attrs.CgrId = "dummyid"
	attrs.RunId = "default"
	if err := rater.Call(utils.APIerSv1GetEventCost, &attrs, &cc); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error("APIerSv1.GetCallCostLog: should return NOT_FOUND, got:", err)
	}
	tm := time.Now().Truncate(time.Millisecond).UTC()
	cdr := &engine.CDRWithAPIOpts{
		CDR: &engine.CDR{
			CGRID:       "Cdr1",
			OrderID:     123,
			ToR:         utils.MetaVoice,
			OriginID:    "OriginCDR1",
			OriginHost:  "192.168.1.1",
			Source:      "test",
			RequestType: utils.MetaRated,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1001",
			Subject:     "1001",
			Destination: "+4986517174963",
			SetupTime:   tm,
			AnswerTime:  tm,
			RunID:       utils.MetaDefault,
			Usage:       0,
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			Cost:        1.01,
		},
	}
	var reply string
	if err := rater.Call(utils.CDRsV1ProcessCDR, cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(100 * time.Millisecond)
	expected := engine.EventCost{
		CGRID:     "Cdr1",
		RunID:     utils.MetaDefault,
		StartTime: tm,
		Usage:     utils.DurationPointer(0),
		Cost:      utils.Float64Pointer(0),
		Charges: []*engine.ChargingInterval{{
			RatingID:       utils.EmptyString,
			Increments:     nil,
			CompressFactor: 0,
		}},
		AccountSummary: nil,
		Rating:         engine.Rating{},
		Accounting:     engine.Accounting{},
		RatingFilters:  engine.RatingFilters{},
		Rates:          engine.ChargedRates{},
		Timings:        engine.ChargedTimings{},
	}
	if *encoding == utils.MetaGOB {
		expected.Usage = nil // 0 value are encoded as nil in gob
		expected.Cost = nil
	}
	attrs.CgrId = "Cdr1"
	attrs.RunId = utils.EmptyString
	if err := rater.Call(utils.APIerSv1GetEventCost, &attrs, &cc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cc) {
		t.Errorf("Expecting %s ,received %s", utils.ToJSON(expected), utils.ToJSON(cc))
	}
}

func testApierITSetDestination(t *testing.T) {
	attrs := utils.AttrSetDestination{Id: "TEST_SET_DESTINATION", Prefixes: []string{"+4986517174963", "+4986517174960"}}
	var reply string
	if err := rater.Call(utils.APIerSv1SetDestination, &attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := rater.Call(utils.APIerSv1SetDestination, &attrs, &reply); err == nil || err.Error() != "EXISTS" { // Second time without overwrite should generate error
		t.Error("Unexpected error", err.Error())
	}
	attrs = utils.AttrSetDestination{Id: "TEST_SET_DESTINATION", Prefixes: []string{"+4986517174963", "+4986517174964"}, Overwrite: true}
	if err := rater.Call(utils.APIerSv1SetDestination, &attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	eDestination := engine.Destination{Id: attrs.Id, Prefixes: attrs.Prefixes}
	var rcvDestination engine.Destination
	if err := rater.Call(utils.APIerSv1GetDestination, &attrs.Id, &rcvDestination); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(eDestination, rcvDestination) {
		t.Errorf("Expecting: %+v, received: %+v", eDestination, rcvDestination)
	}
	eRcvIDs := []string{attrs.Id}
	var rcvIDs []string
	if err := rater.Call(utils.APIerSv1GetReverseDestination, &attrs.Prefixes[0], &rcvIDs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(eRcvIDs, rcvIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eRcvIDs, rcvIDs)
	}
}

func testApierITGetScheduledActions(t *testing.T) {
	var rply []*scheduler.ScheduledAction
	if err := rater.Call(utils.APIerSv1GetScheduledActions, scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
		t.Error("Unexpected error: ", err)
	}
}

func testApierITGetDataCost(t *testing.T) {
	attrs := AttrGetDataCost{Category: "data", Tenant: "cgrates.org",
		Subject: "1001", AnswerTime: utils.MetaNow, Usage: 640113}
	var rply *engine.DataCost
	if err := rater.Call(utils.APIerSv1GetDataCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if rply.Cost != 128.0240 {
		t.Errorf("Unexpected cost received: %f", rply.Cost)
	}
}

func testApierITGetCost(t *testing.T) {
	attrs := AttrGetCost{Category: "data", Tenant: "cgrates.org",
		Subject: "1001", AnswerTime: utils.MetaNow, Usage: "640113"}
	var rply *engine.EventCost
	if err := rater.Call(utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 128.0240 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

// Test LoadTPFromStorDb
func testApierInitDataDb2(t *testing.T) {
	if err := engine.InitDataDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func testApierInitStorDb2(t *testing.T) {
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func testApierReloadCache2(t *testing.T) {
	var reply string
	// Simple test that command is executed without errors
	if err := rater.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}

func testApierReloadScheduler2(t *testing.T) {
	var reply string
	// Simple test that command is executed without errors
	if err := rater.Call(utils.SchedulerSv1Reload, utils.StringWithAPIOpts{}, &reply); err != nil {
		t.Error("Got error on SchedulerSv1.Reload: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling SchedulerSv1.Reload got reply: ", reply)
	}
}

func testApierImportTPFromFolderPath(t *testing.T) {
	var reply string
	if err := rater.Call(utils.APIerSv1ImportTariffPlanFromFolder,
		utils.AttrImportTPFromFolder{TPid: "TEST_TPID2",
			FolderPath: "/usr/share/cgrates/tariffplans/oldtutorial"}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ImportTarrifPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.ImportTarrifPlanFromFolder got reply: ", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testApierLoadTariffPlanFromStorDbDryRun(t *testing.T) {
	var reply string
	if err := rater.Call(utils.APIerSv1LoadTariffPlanFromStorDb,
		&AttrLoadTpFromStorDb{TPid: "TEST_TPID2", DryRun: true}, &reply); err != nil {
		t.Error("Got error on APIerSv1.LoadTariffPlanFromStorDb: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.LoadTariffPlanFromStorDb got reply: ", reply)
	}
}

func testApierGetCacheStats2(t *testing.T) {
	var rcvStats map[string]*ltcache.CacheStats
	expectedStats := engine.GetDefaultEmptyCacheStats()
	err := rater.Call(utils.CacheSv1GetCacheStats, new(utils.AttrCacheIDsWithAPIOpts), &rcvStats)
	if err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling CacheSv1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

func testApierLoadTariffPlanFromStorDb(t *testing.T) {
	var reply string
	if err := rater.Call(utils.APIerSv1LoadTariffPlanFromStorDb,
		&AttrLoadTpFromStorDb{TPid: "TEST_TPID2"}, &reply); err != nil {
		t.Error("Got error on APIerSv1.LoadTariffPlanFromStorDb: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.LoadTariffPlanFromStorDb got reply: ", reply)
	}
}

func testApierStartStopServiceStatus(t *testing.T) {
	var reply string
	if err := rater.Call(utils.ServiceManagerV1ServiceStatus, dispatchers.ArgStartServiceWithAPIOpts{ArgStartService: servmanager.ArgStartService{ServiceID: utils.MetaScheduler}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.RunningCaps {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call(utils.ServiceManagerV1StopService, dispatchers.ArgStartServiceWithAPIOpts{ArgStartService: servmanager.ArgStartService{ServiceID: "INVALID"}},
		&reply); err == nil || err.Error() != utils.UnsupportedServiceIDCaps {
		t.Error(err)
	}
	if err := rater.Call(utils.ServiceManagerV1StopService, dispatchers.ArgStartServiceWithAPIOpts{ArgStartService: servmanager.ArgStartService{ServiceID: utils.MetaScheduler}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call(utils.ServiceManagerV1ServiceStatus, dispatchers.ArgStartServiceWithAPIOpts{ArgStartService: servmanager.ArgStartService{ServiceID: utils.MetaScheduler}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.StoppedCaps {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call(utils.ServiceManagerV1StartService, &dispatchers.ArgStartServiceWithAPIOpts{ArgStartService: servmanager.ArgStartService{ServiceID: utils.MetaScheduler}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call(utils.ServiceManagerV1ServiceStatus, dispatchers.ArgStartServiceWithAPIOpts{ArgStartService: servmanager.ArgStartService{ServiceID: utils.MetaScheduler}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.RunningCaps {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call(utils.SchedulerSv1Reload, utils.StringWithAPIOpts{}, &reply); err != nil {
		t.Error("Got error on SchedulerSv1.Reload: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling SchedulerSv1.Reload got reply: ", reply)
	}
}

func testApierSetRatingProfileWithoutTenant(t *testing.T) {
	var reply string
	rpa := &utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1", FallbackSubjects: "dan4"}
	rpf := &utils.AttrSetRatingProfile{Category: utils.Call, Subject: "dan3", RatingPlanActivations: []*utils.TPRatingActivation{rpa}}
	if err := rater.Call(utils.APIerSv1SetRatingProfile, &rpf, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.SetRatingProfile got reply: ", reply)
	}
	expectedID := utils.ConcatenatedKey(utils.MetaOut, "cgrates.org", utils.Call, "dan3")
	var result *engine.RatingProfile
	if err := rater.Call(utils.APIerSv1GetRatingProfile,
		&utils.AttrGetRatingProfile{Category: utils.Call, Subject: "dan3"},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedID, result.Id) {
		t.Errorf("Expected %+v, received %+v", expectedID, result.Id)
	}
}

func testApierRemoveRatingProfilesWithoutTenant(t *testing.T) {
	var reply string
	if err := rater.Call(utils.APIerSv1RemoveRatingProfile, &AttrRemoveRatingProfile{
		Category: utils.Call,
		Subject:  "dan3",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected: %s, received: %s ", utils.OK, reply)
	}
	var result *engine.RatingProfile
	if err := rater.Call(utils.APIerSv1GetRatingProfile,
		&utils.AttrGetRatingProfile{Category: utils.Call, Subject: "dan3"},
		&result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testApierReplayFldPosts(t *testing.T) {
	bev := []byte(`{"ID":"cgrates.org:1007","BalanceMap":{"*monetary":[{"Uuid":"367be35a-96ee-40a5-b609-9130661f5f12","ID":"","Value":0,"ExpirationDate":"0001-01-01T00:00:00Z","Weight":10,"DestinationIDs":{},"RatingSubject":"","Categories":{},"SharedGroups":{"SHARED_A":true},"Timings":null,"TimingIDs":{},"Disabled":false,"Factor":null,"Blocker":false}]},"UnitCounters":{"*monetary":[{"CounterType":"*event","Counters":[{"Value":0,"Filter":{"Uuid":null,"ID":"b8531413-10d5-47ad-81ad-2bc272e8f0ca","Type":"*monetary","Value":null,"ExpirationDate":null,"Weight":null,"DestinationIDs":{"FS_USERS":true},"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null}}]}]},"ActionTriggers":[{"ID":"STANDARD_TRIGGERS","UniqueID":"46ac7b8c-685d-4555-bf73-fa6cfbc2fa21","ThresholdType":"*min_balance","ThresholdValue":2,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"ExpirationDate":null,"Weight":null,"DestinationIDs":null,"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"LOG_WARNING","MinQueuedItems":0,"Executed":true,"LastExecutionTime":"2017-01-31T14:03:57.961651647+01:00"},{"ID":"STANDARD_TRIGGERS","UniqueID":"b8531413-10d5-47ad-81ad-2bc272e8f0ca","ThresholdType":"*max_event_counter","ThresholdValue":5,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"ExpirationDate":null,"Weight":null,"DestinationIDs":{"FS_USERS":true},"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"LOG_WARNING","MinQueuedItems":0,"Executed":false,"LastExecutionTime":"0001-01-01T00:00:00Z"},{"ID":"STANDARD_TRIGGERS","UniqueID":"8b424186-7a31-4aef-99c5-35e12e6fed41","ThresholdType":"*max_balance","ThresholdValue":20,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"ExpirationDate":null,"Weight":null,"DestinationIDs":null,"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"LOG_WARNING","MinQueuedItems":0,"Executed":false,"LastExecutionTime":"0001-01-01T00:00:00Z"},{"ID":"STANDARD_TRIGGERS","UniqueID":"28557f3b-139c-4a27-9d17-bda1f54b7c19","ThresholdType":"*max_balance","ThresholdValue":100,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"ExpirationDate":null,"Weight":null,"DestinationIDs":null,"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"DISABLE_AND_LOG","MinQueuedItems":0,"Executed":false,"LastExecutionTime":"0001-01-01T00:00:00Z"}],"AllowNegative":false,"Disabled":false}"`)
	ev := &engine.ExportEvents{
		Path:   "http://localhost:2081",
		Format: utils.MetaHTTPjson,
		Events: []interface{}{&engine.HTTPPosterRequest{Body: bev, Header: http.Header{"Content-Type": []string{"application/json"}}}},
	}
	fileName := "act>*http_post|63bed4ea-615e-4096-b1f4-499f64f29b28.json"

	args := ArgsReplyFailedPosts{
		FailedRequestsInDir:  utils.StringPointer("/tmp/TestsAPIerSv1/in"),
		FailedRequestsOutDir: utils.StringPointer("/tmp/TestsAPIerSv1/out"),
	}
	for _, dir := range []string{*args.FailedRequestsInDir, *args.FailedRequestsOutDir} {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("Error %s removing folder: %s", err, dir)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Errorf("Error %s creating folder: %s", err, dir)
		}
	}
	err := ev.WriteToFile(path.Join(*args.FailedRequestsInDir, fileName))
	if err != nil {
		t.Error(err)
	}
	var reply string
	if err := rater.Call(utils.APIerSv1ReplayFailedPosts, &args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply: ", reply)
	}
	outPath := path.Join(*args.FailedRequestsOutDir, fileName)
	outEv, err := engine.NewExportEventsFromFile(outPath)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ev, outEv) {
		t.Errorf("Expecting: %q, received: %q", utils.ToJSON(ev), utils.ToJSON(outEv))
	}
	fileName = "cdr|ae8cc4b3-5e60-4396-b82a-64b96a72a03c.json"
	bev = []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)
	fileInPath := path.Join(*args.FailedRequestsInDir, fileName)
	ev = &engine.ExportEvents{
		Path: "amqp://guest:guest@localhost:5672/",
		Opts: map[string]interface{}{
			"queueID": "cgrates_cdrs",
		},
		Format: utils.MetaAMQPjsonMap,
		Events: []interface{}{bev},
	}
	err = ev.WriteToFile(path.Join(*args.FailedRequestsInDir, fileName))
	if err != nil {
		t.Error(err)
	}
	if err := rater.Call(utils.APIerSv1ReplayFailedPosts, &args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply: ", reply)
	}
	if _, err := os.Stat(fileInPath); !os.IsNotExist(err) {
		t.Error("InFile still exists")
	}
	if _, err := os.Stat(path.Join(*args.FailedRequestsOutDir, fileName)); !os.IsNotExist(err) {
		t.Error("OutFile created")
	}
	// connect to RabbitMQ server and check if the content was posted there
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()
	q, err := ch.QueueDeclare("cgrates_cdrs", true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, utils.EmptyString, true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case d := <-msgs:
		var rcvCDR map[string]string
		if err := json.Unmarshal(d.Body, &rcvCDR); err != nil {
			t.Error(err)
		}
		if rcvCDR[utils.CGRID] != "88ed9c38005f07576a1e1af293063833b60edcc6" {
			t.Errorf("Unexpected CDR received: %+v", rcvCDR)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("No message received from RabbitMQ")
	}
	for _, dir := range []string{*args.FailedRequestsInDir, *args.FailedRequestsOutDir} {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("Error %s removing folder: %s", err, dir)
		}
	}
}

func testApierGetDataDBVesions(t *testing.T) {
	var reply *engine.Versions
	if err := rater.Call(utils.APIerSv1GetDataDBVersions, utils.StringPointer(utils.EmptyString), &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(engine.CurrentDataDBVersions(), *reply) {
		t.Errorf("Expecting : %+v, received: %+v", engine.CurrentDataDBVersions(), *reply)
	}
}

func testApierGetStorDBVesions(t *testing.T) {
	var reply *engine.Versions
	if err := rater.Call(utils.APIerSv1GetStorDBVersions, utils.StringPointer(utils.EmptyString), &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(engine.CurrentStorDBVersions(), *reply) {
		t.Errorf("Expecting : %+v, received: %+v", engine.CurrentStorDBVersions(), *reply)
	}
}

func testApierBackwardsCompatible(t *testing.T) {
	var reply string
	if err := rater.Call("ApierV1.Ping", new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Expecting : %+v, received: %+v", utils.Pong, reply)
	}
}

// Simply kill the engine after we are done with tests within this file
func testApierStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
