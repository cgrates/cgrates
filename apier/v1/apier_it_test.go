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
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
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

var cfgPath string
var cfg *config.CGRConfig
var rater *rpc.Client

var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var storDbType = flag.String("stordb_type", "mysql", "The type of the storDb database <mysql>")
var waitRater = flag.Int("wait_rater", 500, "Number of miliseconds to wait for rater to start and cache")

func TestApierLoadConfig(t *testing.T) {
	var err error
	cfgPath = path.Join(*dataDir, "conf", "samples", "apier")
	if cfg, err = config.NewCGRConfigFromFolder(cfgPath); err != nil {
		t.Error(err)
	}
}

func TestApierCreateDirs(t *testing.T) {
	for _, pathDir := range []string{cfg.CdreProfiles[utils.META_DEFAULT].ExportPath, "/var/log/cgrates/cdrc/in", "/var/log/cgrates/cdrc/out"} {
		if err := os.RemoveAll(pathDir); err != nil {
			t.Fatal("Error removing folder: ", pathDir, err)
		}
		if err := os.MkdirAll(pathDir, 0755); err != nil {
			t.Fatal("Error creating folder: ", pathDir, err)
		}
	}
}

func TestApierInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func TestApierInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}
}

// Start engine
func TestApierStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestApierRpcConn(t *testing.T) {
	var err error
	rater, err = jsonrpc.Dial("tcp", cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Test here TPTiming APIs
func TestApierTPTiming(t *testing.T) {
	// ALWAYS,*any,*any,*any,*any,00:00:00
	tmAlways := &utils.ApierTPTiming{TPid: utils.TEST_SQL,
		ID:        "ALWAYS",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "*any",
		Time:      "00:00:00",
	}
	tmAlways2 := new(utils.ApierTPTiming)
	*tmAlways2 = *tmAlways
	tmAlways2.ID = "ALWAYS2"
	tmAsap := &utils.ApierTPTiming{TPid: utils.TEST_SQL,
		ID:        "ASAP",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "*any",
		Time:      "*asap",
	}
	reply := ""
	for _, tm := range []*utils.ApierTPTiming{tmAlways, tmAsap, tmAlways2} {
		if err := rater.Call("ApierV1.SetTPTiming", tm, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPTiming: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPTiming: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPTiming", tmAlways, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPTiming: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPTiming got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPTiming", new(utils.ApierTPTiming), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPTiming, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID Years Months MonthDays WeekDays Time]" {
		t.Error("Calling ApierV1.SetTPTiming got unexpected error: ", err.Error())
	}
	// Test get
	var rplyTmAlways2 *utils.ApierTPTiming
	if err := rater.Call("ApierV1.GetTPTiming", AttrGetTPTiming{tmAlways2.TPid, tmAlways2.ID}, &rplyTmAlways2); err != nil {
		t.Error("Calling ApierV1.GetTPTiming, got error: ", err.Error())
	} else if !reflect.DeepEqual(tmAlways2, rplyTmAlways2) {
		t.Errorf("Calling ApierV1.GetTPTiming expected: %v, received: %v", tmAlways, rplyTmAlways2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPTiming", AttrGetTPTiming{tmAlways2.TPid, tmAlways2.ID}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPTiming, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPTiming received: ", reply)
	}
	// Test getIds
	var rplyTmIds []string
	expectedTmIds := []string{"ALWAYS", "ASAP"}
	if err := rater.Call("ApierV1.GetTPTimingIds", AttrGetTPTimingIds{tmAlways.TPid, utils.Paginator{}}, &rplyTmIds); err != nil {
		t.Error("Calling ApierV1.GetTPTimingIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedTmIds, rplyTmIds) {
		t.Errorf("Calling ApierV1.GetTPTimingIds expected: %v, received: %v", expectedTmIds, rplyTmIds)
	}
}

// Test here TPTiming APIs
func TestApierTPDestination(t *testing.T) {
	reply := ""
	dstDe := &utils.TPDestination{TPid: utils.TEST_SQL, ID: "GERMANY", Prefixes: []string{"+49"}}
	dstDeMobile := &utils.TPDestination{TPid: utils.TEST_SQL, ID: "GERMANY_MOBILE", Prefixes: []string{"+4915", "+4916", "+4917"}}
	dstFs := &utils.TPDestination{TPid: utils.TEST_SQL, ID: "FS_USERS", Prefixes: []string{"10"}}
	dstDe2 := new(utils.TPDestination)
	*dstDe2 = *dstDe // Data which we use for remove, still keeping the sample data to check proper loading
	dstDe2.ID = "GERMANY2"
	for _, dst := range []*utils.TPDestination{dstDe, dstDeMobile, dstFs, dstDe2} {
		if err := rater.Call("ApierV1.SetTPDestination", dst, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPDestination: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPDestination: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPDestination", dstDe2, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPDestination: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPDestination got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPDestination", new(utils.TPDestination), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPDestination, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID Prefixes]" {
		t.Error("Calling ApierV1.SetTPDestination got unexpected error: ", err.Error())
	}
	// Test get
	var rplyDstDe2 *utils.TPDestination
	if err := rater.Call("ApierV1.GetTPDestination", AttrGetTPDestination{dstDe2.TPid, dstDe2.ID}, &rplyDstDe2); err != nil {
		t.Error("Calling ApierV1.GetTPDestination, got error: ", err.Error())
	} else if !reflect.DeepEqual(dstDe2, rplyDstDe2) {
		t.Errorf("Calling ApierV1.GetTPDestination expected: %v, received: %v", dstDe2, rplyDstDe2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPDestination", AttrGetTPDestination{dstDe2.TPid, dstDe2.ID}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPTiming, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPTiming received: ", reply)
	}
	// Test getIds
	var rplyDstIds []string
	expectedDstIds := []string{"FS_USERS", "GERMANY", "GERMANY_MOBILE"}
	if err := rater.Call("ApierV1.GetTPDestinationIDs", AttrGetTPDestinationIds{TPid: dstDe.TPid}, &rplyDstIds); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationIDs, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedDstIds, rplyDstIds) {
		t.Errorf("Calling ApierV1.GetTPDestinationIDs expected: %v, received: %v", expectedDstIds, rplyDstIds)
	}
}

// Test here TPRate APIs
func TestApierTPRate(t *testing.T) {
	reply := ""
	rt := &utils.TPRate{TPid: utils.TEST_SQL, ID: "RT_FS_USERS", RateSlots: []*utils.RateSlot{
		{ConnectFee: 0, Rate: 0, RateUnit: "60s", RateIncrement: "60s", GroupIntervalStart: "0s"},
	}}
	rt2 := new(utils.TPRate)
	*rt2 = *rt
	rt2.ID = "RT_FS_USERS2"
	for _, r := range []*utils.TPRate{rt, rt2} {
		if err := rater.Call("ApierV1.SetTPRate", r, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPRate: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPRate: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPRate", rt2, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPRate: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPRate got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPRate", new(utils.TPRate), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPDestination, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID RateSlots]" {
		t.Error("Calling ApierV1.SetTPRate got unexpected error: ", err.Error())
	}
	// Test get
	var rplyRt2 *utils.TPRate
	if err := rater.Call("ApierV1.GetTPRate", AttrGetTPRate{rt2.TPid, rt2.ID}, &rplyRt2); err != nil {
		t.Error("Calling ApierV1.GetTPRate, got error: ", err.Error())
	} else if !reflect.DeepEqual(rt2, rplyRt2) {
		t.Errorf("Calling ApierV1.GetTPRate expected: %+v, received: %+v", rt2, rplyRt2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPRate", AttrGetTPRate{rt2.TPid, rt2.ID}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRate, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRate received: ", reply)
	}
	// Test getIds
	var rplyRtIds []string
	expectedRtIds := []string{"RT_FS_USERS"}
	if err := rater.Call("ApierV1.GetTPRateIds", AttrGetTPRateIds{rt.TPid, utils.Paginator{}}, &rplyRtIds); err != nil {
		t.Error("Calling ApierV1.GetTPRateIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRtIds, rplyRtIds) {
		t.Errorf("Calling ApierV1.GetTPDestinationIDs expected: %v, received: %v", expectedRtIds, rplyRtIds)
	}
}

// Test here TPDestinationRate APIs
func TestApierTPDestinationRate(t *testing.T) {
	reply := ""
	dr := &utils.TPDestinationRate{TPid: utils.TEST_SQL, ID: "DR_FREESWITCH_USERS", DestinationRates: []*utils.DestinationRate{
		{DestinationId: "FS_USERS", RateId: "RT_FS_USERS", RoundingMethod: "*up", RoundingDecimals: 2},
	}}
	drDe := &utils.TPDestinationRate{TPid: utils.TEST_SQL, ID: "DR_FREESWITCH_USERS", DestinationRates: []*utils.DestinationRate{
		{DestinationId: "GERMANY_MOBILE", RateId: "RT_FS_USERS", RoundingMethod: "*up", RoundingDecimals: 2},
	}}
	dr2 := new(utils.TPDestinationRate)
	*dr2 = *dr
	dr2.ID = utils.TEST_SQL
	for _, d := range []*utils.TPDestinationRate{dr, dr2, drDe} {
		if err := rater.Call("ApierV1.SetTPDestinationRate", d, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPDestinationRate: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPDestinationRate: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPDestinationRate", dr2, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPDestinationRate: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPDestinationRate got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPDestinationRate", new(utils.TPDestinationRate), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPDestination, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID DestinationRates]" {
		t.Error("Calling ApierV1.SetTPDestinationRate got unexpected error: ", err.Error())
	}
	// Test get
	var rplyDr2 *utils.TPDestinationRate
	if err := rater.Call("ApierV1.GetTPDestinationRate", AttrGetTPDestinationRate{dr2.TPid, dr2.ID, utils.Paginator{}}, &rplyDr2); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationRate, got error: ", err.Error())
	} else if !reflect.DeepEqual(dr2, rplyDr2) {
		t.Errorf("Calling ApierV1.GetTPDestinationRate expected: %v, received: %v", dr2, rplyDr2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPDestinationRate", AttrGetTPDestinationRate{dr2.TPid, dr2.ID, utils.Paginator{}}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRate, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRate received: ", reply)
	}
	// Test getIds
	var rplyDrIds []string
	expectedDrIds := []string{"DR_FREESWITCH_USERS"}
	if err := rater.Call("ApierV1.GetTPDestinationRateIds", AttrTPDestinationRateIds{dr.TPid, utils.Paginator{}}, &rplyDrIds); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationRateIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedDrIds, rplyDrIds) {
		t.Errorf("Calling ApierV1.GetTPDestinationRateIds expected: %v, received: %v", expectedDrIds, rplyDrIds)
	}
}

// Test here TPRatingPlan APIs
func TestApierTPRatingPlan(t *testing.T) {
	reply := ""
	rp := &utils.TPRatingPlan{TPid: utils.TEST_SQL, ID: "RETAIL1", RatingPlanBindings: []*utils.TPRatingPlanBinding{
		{DestinationRatesId: "DR_FREESWITCH_USERS", TimingId: "ALWAYS", Weight: 10},
	}}
	rpTst := new(utils.TPRatingPlan)
	*rpTst = *rp
	rpTst.ID = utils.TEST_SQL
	for _, rpl := range []*utils.TPRatingPlan{rp, rpTst} {
		if err := rater.Call("ApierV1.SetTPRatingPlan", rpl, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPRatingPlan: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPRatingPlan: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPRatingPlan", rpTst, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPRatingPlan: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPRatingPlan got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPRatingPlan", new(utils.TPRatingPlan), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPRatingPlan, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID RatingPlanBindings]" {
		t.Error("Calling ApierV1.SetTPRatingPlan got unexpected error: ", err.Error())
	}
	// Test get
	var rplyRpTst *utils.TPRatingPlan
	if err := rater.Call("ApierV1.GetTPRatingPlan", AttrGetTPRatingPlan{TPid: rpTst.TPid, ID: rpTst.ID}, &rplyRpTst); err != nil {
		t.Error("Calling ApierV1.GetTPRatingPlan, got error: ", err.Error())
	} else if !reflect.DeepEqual(rpTst, rplyRpTst) {
		t.Errorf("Calling ApierV1.GetTPRatingPlan expected: %v, received: %v", rpTst, rplyRpTst)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPRatingPlan", AttrGetTPRatingPlan{TPid: rpTst.TPid, ID: rpTst.ID}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRatingPlan, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRatingPlan received: ", reply)
	}
	// Test getIds
	var rplyRpIds []string
	expectedRpIds := []string{"RETAIL1"}
	if err := rater.Call("ApierV1.GetTPRatingPlanIds", AttrGetTPRatingPlanIds{rp.TPid, utils.Paginator{}}, &rplyRpIds); err != nil {
		t.Error("Calling ApierV1.GetTPRatingPlanIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRpIds, rplyRpIds) {
		t.Errorf("Calling ApierV1.GetTPRatingPlanIds expected: %v, received: %v", expectedRpIds, rplyRpIds)
	}
}

// Test here TPRatingPlan APIs
func TestApierTPRatingProfile(t *testing.T) {
	reply := ""
	rpf := &utils.TPRatingProfile{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1", FallbackSubjects: ""},
		}}
	rpfTst := new(utils.TPRatingProfile)
	*rpfTst = *rpf
	rpfTst.Subject = utils.TEST_SQL
	for _, rp := range []*utils.TPRatingProfile{rpf, rpfTst} {
		if err := rater.Call("ApierV1.SetTPRatingProfile", rp, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPRatingProfile: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPRatingProfile: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPRatingProfile", rpfTst, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPRatingProfile got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPRatingProfile", new(utils.TPRatingProfile), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPRatingProfile, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid LoadId Tenant Category Direction Subject RatingPlanActivations]" {
		t.Error("Calling ApierV1.SetTPRatingProfile got unexpected error: ", err.Error())
	}
	// Test get
	var rplyRpf *utils.TPRatingProfile
	if err := rater.Call("ApierV1.GetTPRatingProfile", AttrGetTPRatingProfile{TPid: rpfTst.TPid, RatingProfileId: rpfTst.GetRatingProfilesId()}, &rplyRpf); err != nil {
		t.Error("Calling ApierV1.GetTPRatingProfiles, got error: ", err.Error())
	} else if !reflect.DeepEqual(rpfTst, rplyRpf) {
		t.Errorf("Calling ApierV1.GetTPRatingProfiles expected: %v, received: %v", rpfTst, rplyRpf)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPRatingProfile", AttrGetTPRatingProfile{TPid: rpfTst.TPid, RatingProfileId: rpfTst.GetRatingProfilesId()}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRatingProfile, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRatingProfile received: ", reply)
	}
	// Test getLoadIds
	var rplyRpIds []string
	expectedRpIds := []string{utils.TEST_SQL}
	if err := rater.Call("ApierV1.GetTPRatingProfileLoadIds", utils.AttrTPRatingProfileIds{TPid: rpf.TPid}, &rplyRpIds); err != nil {
		t.Error("Calling ApierV1.GetTPRatingProfileLoadIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRpIds, rplyRpIds) {
		t.Errorf("Calling ApierV1.GetTPRatingProfileLoadIds expected: %v, received: %v", expectedRpIds, rplyRpIds)
	}
}

func TestApierTPActions(t *testing.T) {
	reply := ""
	act := &utils.TPActions{TPid: utils.TEST_SQL, ID: "PREPAID_10", Actions: []*utils.TPAction{
		{Identifier: "*topup_reset", BalanceType: "*monetary", Directions: "*out", Units: "10", ExpiryTime: "*unlimited",
			DestinationIds: "*any", BalanceWeight: "10", Weight: 10},
	}}
	actWarn := &utils.TPActions{TPid: utils.TEST_SQL, ID: "WARN_VIA_HTTP", Actions: []*utils.TPAction{
		{Identifier: "*call_url", ExtraParameters: "http://localhost:8000", Weight: 10},
	}}
	actLog := &utils.TPActions{TPid: utils.TEST_SQL, ID: "LOG_BALANCE", Actions: []*utils.TPAction{
		{Identifier: "*log", Weight: 10},
	}}
	actTst := new(utils.TPActions)
	*actTst = *act
	actTst.ID = utils.TEST_SQL
	for _, ac := range []*utils.TPActions{act, actWarn, actTst, actLog} {
		if err := rater.Call("ApierV1.SetTPActions", ac, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPActions: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPActions: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPActions", actTst, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPActions: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPActions got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPActions", new(utils.TPActions), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPActions, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID Actions]" {
		t.Error("Calling ApierV1.SetTPActions got unexpected error: ", err.Error())
	}
	// Test get
	var rplyActs *utils.TPActions
	if err := rater.Call("ApierV1.GetTPActions", AttrGetTPActions{TPid: actTst.TPid, ID: actTst.ID}, &rplyActs); err != nil {
		t.Error("Calling ApierV1.GetTPActions, got error: ", err.Error())
	} else if !reflect.DeepEqual(actTst, rplyActs) {
		t.Errorf("Calling ApierV1.GetTPActions expected: %v, received: %v", actTst, rplyActs)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPActions", AttrGetTPActions{TPid: actTst.TPid, ID: actTst.ID}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPActions, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPActions received: ", reply)
	}
	// Test getIds
	var rplyIds []string
	expectedIds := []string{"LOG_BALANCE", "PREPAID_10", "WARN_VIA_HTTP"}
	if err := rater.Call("ApierV1.GetTPActionIds", AttrGetTPActionIds{TPid: actTst.TPid}, &rplyIds); err != nil {
		t.Error("Calling ApierV1.GetTPActionIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedIds, rplyIds) {
		t.Errorf("Calling ApierV1.GetTPActionIds expected: %v, received: %v", expectedIds, rplyIds)
	}
}

func TestApierTPActionPlan(t *testing.T) {
	reply := ""
	at := &utils.TPActionPlan{TPid: utils.TEST_SQL, ID: "PREPAID_10", ActionPlan: []*utils.TPActionTiming{
		{ActionsId: "PREPAID_10", TimingId: "ASAP", Weight: 10},
	}}
	atTst := new(utils.TPActionPlan)
	*atTst = *at
	atTst.ID = utils.TEST_SQL
	for _, act := range []*utils.TPActionPlan{at, atTst} {
		if err := rater.Call("ApierV1.SetTPActionPlan", act, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPActionPlan: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPActionPlan: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPActionPlan", atTst, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPActionPlan: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPActionPlan got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPActionPlan", new(utils.TPActionPlan), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPActionPlan, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID ActionPlan]" {
		t.Error("Calling ApierV1.SetTPActionPlan got unexpected error: ", err.Error())
	}
	// Test get
	var rplyActs *utils.TPActionPlan
	if err := rater.Call("ApierV1.GetTPActionPlan", AttrGetTPActionPlan{TPid: atTst.TPid, ID: atTst.ID}, &rplyActs); err != nil {
		t.Error("Calling ApierV1.GetTPActionPlan, got error: ", err.Error())
	} else if !reflect.DeepEqual(atTst, rplyActs) {
		t.Errorf("Calling ApierV1.GetTPActionPlan expected: %v, received: %v", atTst, rplyActs)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPActionPlan", AttrGetTPActionPlan{TPid: atTst.TPid, ID: atTst.ID}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPActionPlan, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPActionPlan received: ", reply)
	}
	// Test getIds
	var rplyIds []string
	expectedIds := []string{"PREPAID_10"}
	if err := rater.Call("ApierV1.GetTPActionPlanIds", AttrGetTPActionPlanIds{TPid: atTst.TPid}, &rplyIds); err != nil {
		t.Error("Calling ApierV1.GetTPActionPlanIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedIds, rplyIds) {
		t.Errorf("Calling ApierV1.GetTPActionPlanIds expected: %v, received: %v", expectedIds, rplyIds)
	}
}

func TestApierTPActionTriggers(t *testing.T) {
	reply := ""
	at := &utils.TPActionTriggers{TPid: utils.TEST_SQL, ID: "STANDARD_TRIGGERS", ActionTriggers: []*utils.TPActionTrigger{
		{Id: "STANDARD_TRIGGERS", UniqueID: "MYFIRSTTRIGGER", BalanceType: "*monetary", BalanceDirections: "*out", ThresholdType: "*min_balance", ThresholdValue: 2, ActionsId: "LOG_BALANCE", Weight: 10},
	}}
	atTst := new(utils.TPActionTriggers)
	*atTst = *at
	atTst.ID = utils.TEST_SQL
	for _, act := range []*utils.TPActionTriggers{at, atTst} {
		if err := rater.Call("ApierV1.SetTPActionTriggers", act, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPActionTriggers: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPActionTriggers: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPActionTriggers", atTst, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPActionTriggers: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPActionTriggers got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPActionTriggers", new(utils.TPActionTriggers), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPActionTriggers, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid ID]" {
		t.Error("Calling ApierV1.SetTPActionTriggers got unexpected error: ", err.Error())
	}
	atTst.ActionTriggers[0].Id = utils.TEST_SQL
	// Test get
	var rplyActs *utils.TPActionTriggers
	if err := rater.Call("ApierV1.GetTPActionTriggers", AttrGetTPActionTriggers{TPid: atTst.TPid, ID: atTst.ID}, &rplyActs); err != nil {
		t.Errorf("Calling ApierV1.GetTPActionTriggers %s, got error: %s", atTst.ID, err.Error())
	} else if !reflect.DeepEqual(atTst, rplyActs) {
		t.Errorf("Calling ApierV1.GetTPActionTriggers expected: %+v, received: %+v", atTst.ActionTriggers[0], rplyActs.ActionTriggers[0])
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPActionTriggers", AttrGetTPActionTriggers{TPid: atTst.TPid, ID: atTst.ID}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPActionTriggers, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPActionTriggers received: ", reply)
	}
	// Test getIds
	var rplyIds []string
	expectedIds := []string{"STANDARD_TRIGGERS"}
	if err := rater.Call("ApierV1.GetTPActionTriggerIds", AttrGetTPActionTriggerIds{TPid: atTst.TPid}, &rplyIds); err != nil {
		t.Error("Calling ApierV1.GetTPActionTriggerIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedIds, rplyIds) {
		t.Errorf("Calling ApierV1.GetTPActionTriggerIds expected: %v, received: %v", expectedIds, rplyIds)
	}
}

// Test here TPAccountActions APIs
func TestApierTPAccountActions(t *testing.T) {
	reply := ""
	aa1 := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org",
		Account: "1001", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa2 := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org",
		Account: "1002", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa3 := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org",
		Account: "1003", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa4 := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org",
		Account: "1004", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa5 := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org",
		Account: "1005", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aaTst := new(utils.TPAccountActions)
	*aaTst = *aa1
	aaTst.Account = utils.TEST_SQL
	for _, aact := range []*utils.TPAccountActions{aa1, aa2, aa3, aa4, aa5, aaTst} {
		if err := rater.Call("ApierV1.SetTPAccountActions", aact, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPAccountActions: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPAccountActions: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPAccountActions", aaTst, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPAccountActions: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPAccountActions got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPAccountActions", new(utils.TPAccountActions), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPAccountActions, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING: [TPid LoadId Tenant Account ActionPlanId]" {
		t.Error("Calling ApierV1.SetTPAccountActions got unexpected error: ", err.Error())
	}
	// Test get
	var rplyaa *utils.TPAccountActions
	if err := rater.Call("ApierV1.GetTPAccountActions", AttrGetTPAccountActions{TPid: aaTst.TPid, AccountActionsId: aaTst.GetId()}, &rplyaa); err != nil {
		t.Error("Calling ApierV1.GetTPAccountActions, got error: ", err.Error())
	} else if !reflect.DeepEqual(aaTst, rplyaa) {
		t.Errorf("Calling ApierV1.GetTPAccountActions expected: %v, received: %v", aaTst, rplyaa)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPAccountActions", AttrGetTPAccountActions{TPid: aaTst.TPid, AccountActionsId: aaTst.GetId()}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPAccountActions, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPAccountActions received: ", reply)
	}
	// Test getLoadIds
	var rplyRpIds []string
	expectedRpIds := []string{utils.TEST_SQL}
	if err := rater.Call("ApierV1.GetTPAccountActionLoadIds", AttrGetTPAccountActionIds{TPid: aaTst.TPid}, &rplyRpIds); err != nil {
		t.Error("Calling ApierV1.GetTPAccountActionLoadIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRpIds, rplyRpIds) {
		t.Errorf("Calling ApierV1.GetTPAccountActionLoadIds expected: %v, received: %v", expectedRpIds, rplyRpIds)
	}
}

// Test here LoadRatingPlan
func TestApierLoadRatingPlan(t *testing.T) {
	reply := ""
	if err := rater.Call("ApierV1.LoadRatingPlan", AttrLoadRatingPlan{TPid: utils.TEST_SQL, RatingPlanId: "RETAIL1"}, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadRatingPlan: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadRatingPlan got reply: ", reply)
	}
}

// Test here LoadRatingProfile
func TestApierLoadRatingProfile(t *testing.T) {
	reply := ""
	rpf := &utils.TPRatingProfile{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any"}
	if err := rater.Call("ApierV1.LoadRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadRatingProfile got reply: ", reply)
	}
}

// Test here LoadAccountActions
func TestApierLoadAccountActions(t *testing.T) {
	var rcvStats *utils.CacheStats
	var args utils.AttrCacheStats
	expectedStats := new(utils.CacheStats) // Make sure nothing in cache so far
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
	reply := ""
	aa1 := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Account: "1001"}
	if err := rater.Call("ApierV1.LoadAccountActions", aa1, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadAccountActions: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadAccountActions got reply: ", reply)
	}
	time.Sleep(10 * time.Millisecond)
	expectedStats = &utils.CacheStats{Actions: 1, ActionPlans: 1, AccountActionPlans: 1}
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

// Test here ReloadScheduler
func TestApierReloadScheduler(t *testing.T) {
	reply := ""
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.ReloadScheduler", reply, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadScheduler: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.ReloadScheduler got reply: ", reply)
	}
}

// Test here SetRatingProfile
func TestApierSetRatingProfile(t *testing.T) {
	reply := ""
	rpa := &utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1", FallbackSubjects: "dan2"}
	rpf := &utils.AttrSetRatingProfile{Tenant: "cgrates.org", Category: "call",
		Direction: "*out", Subject: "dan", RatingPlanActivations: []*utils.TPRatingActivation{rpa}}
	if err := rater.Call("ApierV1.SetRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.SetRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetRatingProfile got reply: ", reply)
	}
	var rcvStats *utils.CacheStats
	var args utils.AttrCacheStats
	expectedStats := &utils.CacheStats{RatingProfiles: 1, Actions: 1, ActionPlans: 1, AccountActionPlans: 1}
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
	// Calling the second time should not raise EXISTS
	if err := rater.Call("ApierV1.SetRatingProfile", rpf, &reply); err != nil {
		t.Error("Unexpected result on duplication: ", err.Error())
	}
	// Make sure rates were loaded for account dan
	// Test here ResponderGetCost
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:30Z", "")
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "dan",
		Account:       "dan",
		Destination:   "+4917621621391",
		DurationIndex: 90,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	var cc engine.CallCost
	// Simple test that command is executed without errors
	if err := rater.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	expectedStats = &utils.CacheStats{
		ReverseDestinations: 10,
		RatingPlans:         1,
		RatingProfiles:      1,
		Actions:             1,
		ActionPlans:         1,
		AccountActionPlans:  1,
		Aliases:             6,
	}
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

func TestApierV1GetRatingProfile(t *testing.T) {
	var rpl engine.RatingProfile
	attrGetRatingPlan := &utils.AttrGetRatingProfile{Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "dan"}
	actTime, err := utils.ParseTimeDetectLayout("2012-01-01T00:00:00Z", "")
	if err != nil {
		t.Error(err)
	}
	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:call:dan",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime:  actTime,
				RatingPlanId:    "RETAIL1",
				FallbackKeys:    []string{"*out:cgrates.org:call:dan2"},
				CdrStatQueueIds: nil,
			},
			{
				ActivationTime:  actTime,
				RatingPlanId:    "RETAIL1",
				FallbackKeys:    []string{"*out:cgrates.org:call:dan2"},
				CdrStatQueueIds: nil,
			},
		},
	}
	if err := rater.Call("ApierV1.GetRatingProfile", attrGetRatingPlan, &rpl); err != nil {
		t.Errorf("Got error on ApierV1.GetRatingProfile: %+v", err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Calling ApierV1.GetRatingProfile expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}
	attrGetRatingPlan.Subject = ""
	if err := rater.Call("ApierV1.GetRatingProfile", attrGetRatingPlan, &rpl); err == nil {
		t.Errorf("Expected error on ApierV1.GetRatingProfile, recived : %+v", rpl)
	}
	attrGetRatingPlan.Subject = "dan"
	attrGetRatingPlan.Tenant = "other_tenant"
	if err := rater.Call("ApierV1.GetRatingProfile", attrGetRatingPlan, &rpl); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error on ApierV1.GetRatingProfile, recived : %+v", err)
	}
}

// Test here ReloadCache
func TestApierReloadCache(t *testing.T) {
	reply := ""
	arc := new(utils.AttrReloadCache)
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.ReloadCache", arc, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadCache: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.ReloadCache got reply: ", reply)
	}
	var rcvStats *utils.CacheStats
	var args utils.AttrCacheStats
	expectedStats := &utils.CacheStats{
		ReverseDestinations: 10,
		RatingPlans:         1,
		RatingProfiles:      2,
		Actions:             1,
		ActionPlans:         1,
		AccountActionPlans:  1,
		Aliases:             6,
	}
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

// Test here GetDestination
func TestApierGetDestination(t *testing.T) {
	reply := new(engine.Destination)
	dstId := "GERMANY_MOBILE"
	expectedReply := &engine.Destination{Id: dstId, Prefixes: []string{"+4915", "+4916", "+4917"}}
	if err := rater.Call("ApierV1.GetDestination", dstId, reply); err != nil {
		t.Error("Got error on ApierV1.GetDestination: ", err.Error())
	} else if !reflect.DeepEqual(expectedReply, reply) {
		t.Errorf("Calling ApierV1.GetDestination expected: %v, received: %v", expectedReply, reply)
	}
}

// Test here GetRatingPlan
func TestApierGetRatingPlan(t *testing.T) {
	reply := new(engine.RatingPlan)
	rplnId := "RETAIL1"
	if err := rater.Call("ApierV1.GetRatingPlan", rplnId, reply); err != nil {
		t.Error("Got error on ApierV1.GetRatingPlan: ", err.Error())
	}
	// Check parts of info received since a full one is not possible due to unique map keys inside reply
	if reply.Id != rplnId {
		t.Error("Unexpected id received", reply.Id)
	}
	if len(reply.Timings) != 1 || len(reply.Ratings) != 1 {
		t.Error("Unexpected number of items received")
	}
	riRate := &engine.RIRate{ConnectFee: 0, RoundingMethod: "*up", RoundingDecimals: 2, Rates: []*engine.Rate{
		{GroupIntervalStart: 0, Value: 0, RateIncrement: time.Duration(60) * time.Second, RateUnit: time.Duration(60) * time.Second},
	}}
	for _, rating := range reply.Ratings {
		riRateJson, _ := json.Marshal(rating)
		if !reflect.DeepEqual(rating, riRate) {
			t.Errorf("Unexpected riRate received: %s", riRateJson)
			// {"Id":"RT_FS_USERS","ConnectFee":0,"Rates":[{"GroupIntervalStart":0,"Value":0,"RateIncrement":60000000000,"RateUnit":60000000000}],"RoundingMethod":"*up","RoundingDecimals":0}
		}
	}
}

// Test here AddBalance
func TestApierAddBalance(t *testing.T) {
	reply := ""
	attrs := &AttrAddBalance{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan", BalanceType: "*monetary", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan2", BalanceType: "*monetary", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan3", BalanceType: "*monetary", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan3", BalanceType: "*monetary", Value: 2.1}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan6", BalanceType: "*monetary", Value: 2.1}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan6", BalanceType: "*monetary", Value: 1, Overwrite: true}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}

}

// Test here ExecuteAction
func TestApierExecuteAction(t *testing.T) {
	reply := ""
	// Add balance to a previously known account
	attrs := utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "dan2", ActionsId: "PREPAID_10"}
	if err := rater.Call("ApierV1.ExecuteAction", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}
	reply2 := ""
	// Add balance to an account which does n exist
	attrs = utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "dan2", ActionsId: "DUMMY_ACTION"}
	if err := rater.Call("ApierV1.ExecuteAction", attrs, &reply2); err == nil || reply2 == "OK" {
		t.Error("Expecting error on ApierV1.ExecuteAction.", err, reply2)
	}
}

func TestApierSetActions(t *testing.T) {
	act1 := &V1TPAction{Identifier: engine.TOPUP_RESET, BalanceType: utils.MONETARY, Directions: utils.OUT, Units: 75.0, ExpiryTime: engine.UNLIMITED, Weight: 20.0}
	attrs1 := &V1AttrSetActions{ActionsId: "ACTS_1", Actions: []*V1TPAction{act1}}
	reply1 := ""
	if err := rater.Call("ApierV1.SetActions", attrs1, &reply1); err != nil {
		t.Error("Got error on ApierV1.SetActions: ", err.Error())
	} else if reply1 != "OK" {
		t.Errorf("Calling ApierV1.SetActions received: %s", reply1)
	}
	// Calling the second time should raise EXISTS
	if err := rater.Call("ApierV1.SetActions", attrs1, &reply1); err == nil || err.Error() != "EXISTS" {
		t.Error("Unexpected result on duplication: ", err.Error())
	}
}

func TestApierGetActions(t *testing.T) {
	expectActs := []*utils.TPAction{
		{Identifier: engine.TOPUP_RESET, BalanceType: utils.MONETARY, Directions: utils.OUT, Units: "75", BalanceWeight: "0", BalanceBlocker: "false", BalanceDisabled: "false", ExpiryTime: engine.UNLIMITED, Weight: 20.0}}

	var reply []*utils.TPAction
	if err := rater.Call("ApierV1.GetActions", "ACTS_1", &reply); err != nil {
		t.Error("Got error on ApierV1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(expectActs, reply) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(expectActs), utils.ToJSON(reply))
	}
}

func TestApierSetActionPlan(t *testing.T) {
	atm1 := &AttrActionPlan{ActionsId: "ACTS_1", MonthDays: "1", Time: "00:00:00", Weight: 20.0}
	atms1 := &AttrSetActionPlan{Id: "ATMS_1", ActionPlan: []*AttrActionPlan{atm1}}
	reply1 := ""
	if err := rater.Call("ApierV1.SetActionPlan", atms1, &reply1); err != nil {
		t.Error("Got error on ApierV1.SetActionPlan: ", err.Error())
	} else if reply1 != "OK" {
		t.Errorf("Calling ApierV1.SetActionPlan received: %s", reply1)
	}
	// Calling the second time should raise EXISTS
	if err := rater.Call("ApierV1.SetActionPlan", atms1, &reply1); err == nil || err.Error() != "EXISTS" {
		t.Error("Unexpected result on duplication: ", err.Error())
	}
}

// Test here AddTriggeredAction
func TestApierAddTriggeredAction(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{Tenant: "cgrates.org", Account: "dan32", BalanceType: "*monetary", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	// Add balance to a previously known account
	attrsAddTrigger := &AttrAddActionTrigger{Tenant: "cgrates.org", Account: "dan32", BalanceDirection: "*out", BalanceType: "*monetary",
		ThresholdType: "*min_balance", ThresholdValue: 2, BalanceDestinationIds: "*any", Weight: 10, ActionsId: "WARN_VIA_HTTP"}
	if err := rater.Call("ApierV1.AddTriggeredAction", attrsAddTrigger, &reply); err != nil {
		t.Error("Got error on ApierV1.AddTriggeredAction: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddTriggeredAction received: %s", reply)
	}
	reply2 := ""
	attrs2 := new(AttrAddActionTrigger)
	*attrs2 = *attrsAddTrigger
	attrs2.Account = "dan10" // Does not exist so it should error when adding triggers on it
	// Add trigger to an account which does n exist
	if err := rater.Call("ApierV1.AddTriggeredAction", attrs2, &reply2); err == nil {
		t.Error("Expecting error on ApierV1.AddTriggeredAction.", err, reply2)
	}
}

// Test here GetAccountActionTriggers
func TestApierGetAccountActionTriggers(t *testing.T) {
	var reply engine.ActionTriggers
	req := AttrAcntAction{Tenant: "cgrates.org", Account: "dan32"}
	if err := rater.Call("ApierV1.GetAccountActionTriggers", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionTimings: ", err.Error())
	} else if len(reply) != 1 || reply[0].ActionsID != "WARN_VIA_HTTP" {
		t.Errorf("Unexpected action triggers received %v", reply)
	}
}

func TestApierAddTriggeredAction2(t *testing.T) {
	reply := ""
	// Add balance to a previously known account
	attrs := &AttrAddAccountActionTriggers{ActionTriggerIDs: &[]string{"STANDARD_TRIGGERS"}, Tenant: "cgrates.org", Account: "dan2"}
	if err := rater.Call("ApierV1.AddAccountActionTriggers", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddAccountActionTriggers: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddAccountActionTriggers received: %s", reply)
	}
	reply2 := ""
	attrs2 := new(AttrAddAccountActionTriggers)
	*attrs2 = *attrs
	attrs2.Account = "dan10" // Does not exist so it should error when adding triggers on it
	// Add trigger to an account which does n exist
	if err := rater.Call("ApierV1.AddAccountActionTriggers", attrs2, &reply2); err == nil || reply2 == "OK" {
		t.Error("Expecting error on ApierV1.AddAccountActionTriggers.", err, reply2)
	}
}

// Test here GetAccountActionTriggers
func TestApierGetAccountActionTriggers2(t *testing.T) {
	var reply engine.ActionTriggers
	req := AttrAcntAction{Tenant: "cgrates.org", Account: "dan2"}
	if err := rater.Call("ApierV1.GetAccountActionTriggers", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionTimings: ", err.Error())
	} else if len(reply) != 1 || reply[0].ActionsID != "LOG_BALANCE" {
		t.Errorf("Unexpected action triggers received %v", reply)
	}
}

// Test here SetAccountActionTriggers
func TestApierSetAccountActionTriggers(t *testing.T) {
	// Test first get so we can steal the id which we need to remove
	var reply engine.ActionTriggers
	req := AttrAcntAction{Tenant: "cgrates.org", Account: "dan2"}
	if err := rater.Call("ApierV1.GetAccountActionTriggers", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionTimings: ", err.Error())
	} else if len(reply) != 1 || reply[0].ActionsID != "LOG_BALANCE" {
		for _, atr := range reply {
			t.Logf("ATR: %+v", atr)
		}
		t.Errorf("Unexpected action triggers received %v", reply)
	}
	var setReply string
	setReq := AttrSetAccountActionTriggers{Tenant: "cgrates.org", Account: "dan2", UniqueID: reply[0].UniqueID, ActivationDate: utils.StringPointer("2016-02-05T18:00:00Z")}
	if err := rater.Call("ApierV1.ResetAccountActionTriggers", setReq, &setReply); err != nil {
		t.Error("Got error on ApierV1.ResetActionTiming: ", err.Error())
	} else if setReply != OK {
		t.Error("Unexpected answer received", setReply)
	}
	if err := rater.Call("ApierV1.SetAccountActionTriggers", setReq, &setReply); err != nil {
		t.Error("Got error on ApierV1.RemoveActionTiming: ", err.Error())
	} else if setReply != OK {
		t.Error("Unexpected answer received", setReply)
	}
	if err := rater.Call("ApierV1.GetAccountActionTriggers", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionTriggers: ", err.Error())
	} else if len(reply) != 1 || reply[0].ActivationDate != time.Date(2016, 2, 5, 18, 0, 0, 0, time.UTC) {
		t.Errorf("Unexpected action triggers received %+v", reply[0])
	}
}

// Test here RemAccountActionTriggers
func TestApierRemAccountActionTriggers(t *testing.T) {
	// Test first get so we can steal the id which we need to remove
	var reply engine.ActionTriggers
	req := AttrAcntAction{Tenant: "cgrates.org", Account: "dan2"}
	if err := rater.Call("ApierV1.GetAccountActionTriggers", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionTimings: ", err.Error())
	} else if len(reply) != 1 || reply[0].ActionsID != "LOG_BALANCE" {
		for _, atr := range reply {
			t.Logf("ATR: %+v", atr)
		}
		t.Errorf("Unexpected action triggers received %v", reply)
	}
	var rmReply string
	rmReq := AttrRemoveAccountActionTriggers{Tenant: "cgrates.org", Account: "dan2", UniqueID: reply[0].UniqueID}
	if err := rater.Call("ApierV1.ResetAccountActionTriggers", rmReq, &rmReply); err != nil {
		t.Error("Got error on ApierV1.ResetActionTiming: ", err.Error())
	} else if rmReply != OK {
		t.Error("Unexpected answer received", rmReply)
	}
	if err := rater.Call("ApierV1.RemoveAccountActionTriggers", rmReq, &rmReply); err != nil {
		t.Error("Got error on ApierV1.RemoveActionTiming: ", err.Error())
	} else if rmReply != OK {
		t.Error("Unexpected answer received", rmReply)
	}
	if err := rater.Call("ApierV1.GetAccountActionTriggers", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionTriggers: ", err.Error())
	} else if len(reply) != 0 {
		t.Errorf("Unexpected action triggers received %+v", reply[0])
	}
}

// Test here SetAccount
func TestApierSetAccount(t *testing.T) {
	reply := ""
	attrs := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan7", ActionPlanId: "ATMS_1", ReloadScheduler: true}
	if err := rater.Call("ApierV1.SetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccount: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
	}
	reply2 := ""
	attrs2 := new(utils.AttrSetAccount)
	*attrs2 = *attrs
	attrs2.ActionPlanId = "DUMMY_DATA" // Does not exist so it should error when adding triggers on it
	// Add account with actions timing which does not exist
	if err := rater.Call("ApierV1.SetAccount", attrs2, &reply2); err == nil || reply2 == "OK" { // OK is not welcomed
		t.Error("Expecting error on ApierV1.SetAccount.", err, reply2)
	}
}

// Test here GetAccountActionTimings
func TestApierGetAccountActionPlan(t *testing.T) {
	var reply []*AccountActionTiming
	req := AttrAcntAction{Tenant: "cgrates.org", Account: "dan7"}
	if err := rater.Call("ApierV1.GetAccountActionPlan", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionPlan: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected action plan received: ", utils.ToJSON(reply))
	} else {
		if reply[0].ActionPlanId != "ATMS_1" {
			t.Errorf("Unexpected ActionoveAccountPlanId received")
		}
	}
}

// Make sure we have scheduled actions
func TestApierITGetScheduledActionsForAccount(t *testing.T) {
	var rply []*scheduler.ScheduledAction
	if err := rater.Call("ApierV1.GetScheduledActions",
		scheduler.ArgsGetScheduledActions{
			Tenant:  utils.StringPointer("cgrates.org"),
			Account: utils.StringPointer("dan7")}, &rply); err != nil {
		t.Error("Unexpected error: ", err)
	} else if len(rply) == 0 {
		t.Errorf("ScheduledActions: %+v", rply)
	}
}

// Test here RemoveActionTiming
func TestApierRemUniqueIDActionTiming(t *testing.T) {
	var rmReply string
	rmReq := AttrRemActionTiming{ActionPlanId: "ATMS_1", Tenant: "cgrates.org", Account: "dan4"}
	if err := rater.Call("ApierV1.RemActionTiming", rmReq, &rmReply); err != nil {
		t.Error("Got error on ApierV1.RemActionTiming: ", err.Error())
	} else if rmReply != OK {
		t.Error("Unexpected answer received", rmReply)
	}
	var reply []*AccountActionTiming
	req := AttrAcntAction{Tenant: "cgrates.org", Account: "dan4"}
	if err := rater.Call("ApierV1.GetAccountActionPlan", req, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccountActionPlan: ", err.Error())
	} else if len(reply) != 0 {
		t.Error("Action timings was not removed")
	}
}

// Test here GetAccount
func TestApierGetAccount(t *testing.T) {
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rater.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 11.5 { // We expect 11.5 since we have added in the previous test 1.5
		t.Errorf("Calling ApierV1.GetBalance expected: 11.5, received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan"}
	if err := rater.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 1.5 {
		t.Errorf("Calling ApierV1.GetAccount expected: 1.5, received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	// The one we have topped up though executeAction
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan2"}
	if err := rater.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 11.5 {
		t.Errorf("Calling ApierV1.GetAccount expected: 10, received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan3"}
	if err := rater.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 3.6 {
		t.Errorf("Calling ApierV1.GetAccount expected: 3.6, received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "dan6"}
	if err := rater.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 1 {
		t.Errorf("Calling ApierV1.GetAccount expected: 1, received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

// Start with initial balance, top-up to test max_balance
func TestApierTriggersExecute(t *testing.T) {
	reply := ""
	attrs := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan8", ReloadScheduler: true}
	if err := rater.Call("ApierV1.SetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccount: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
	}
	attrAddBlnc := &AttrAddBalance{Tenant: "cgrates.org", Account: "1008", BalanceType: "*monetary", Value: 2}
	if err := rater.Call("ApierV1.AddBalance", attrAddBlnc, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
}

// Start fresh before loading from folder
func TestApierResetDataBeforeLoadFromFolder(t *testing.T) {
	TestApierInitDataDb(t)
	var reply string
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.FlushCache", utils.AttrReloadCache{FlushAll: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error("Reply: ", reply)
	}
	var rcvStats *utils.CacheStats
	var args utils.AttrCacheStats
	err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats)
	expectedStats := new(utils.CacheStats)
	if err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(rcvStats, expectedStats) {
		t.Errorf("Calling ApierV1.GetCacheStats received: %v, expected: %v", rcvStats, expectedStats)
	}
}

// Test here LoadTariffPlanFromFolder
func TestApierLoadTariffPlanFromFolder(t *testing.T) {
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: ""}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err == nil || !strings.HasPrefix(err.Error(), utils.ErrMandatoryIeMissing.Error()) {
		t.Error(err)
	}
	attrs = &utils.AttrLoadTpFromFolder{FolderPath: "/INVALID/"}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err == nil || err.Error() != utils.ErrInvalidPath.Error() {
		t.Error(err)
	}
	// Simple test that command is executed without errors
	attrs = &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

// For now just test that they execute without errors
func TestApierComputeReverse(t *testing.T) {
	var reply string
	if err := rater.Call("ApierV1.ComputeReverseDestinations", "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Received: ", reply)
	}
	if err := rater.Call("ApierV1.ComputeReverseAliases", "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Received: ", reply)
	}
	if err := rater.Call("ApierV1.ComputeAccountActionPlans", "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Received: ", reply)
	}
}

func TestApierResetDataAfterLoadFromFolder(t *testing.T) {
	time.Sleep(10 * time.Millisecond)
	expStats := &utils.CacheStats{
		Destinations:       3,
		Actions:            6,
		ActionPlans:        7,
		AccountActionPlans: 13,
		Aliases:            1,
		AttributeProfiles:  0} // Did not cache because it wasn't previously cached
	var rcvStats *utils.CacheStats
	if err := rater.Call("ApierV1.GetCacheStats", utils.AttrCacheStats{}, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
	reply := ""
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.LoadCache", utils.AttrReloadCache{}, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	//expStats = &utils.CacheStats{Destinations: 3, ReverseDestinations: 5}
	if err := rater.Call("ApierV1.GetCacheStats", utils.AttrCacheStats{}, &rcvStats); err != nil {
		t.Error(err)
		//} else if !reflect.DeepEqual(expStats, rcvStats) {
	} else {
		if rcvStats.Destinations != 3 ||
			rcvStats.ReverseDestinations != 5 ||
			rcvStats.RatingPlans != 5 ||
			rcvStats.RatingProfiles != 5 ||
			rcvStats.Actions != 13 ||
			rcvStats.ActionPlans != 7 ||
			rcvStats.SharedGroups != 0 ||
			rcvStats.DerivedChargers != 3 ||
			rcvStats.Aliases != 1 ||
			rcvStats.ReverseAliases != 2 ||
			rcvStats.ResourceProfiles != 3 ||
			rcvStats.Resources != 3 {
			t.Errorf("Expecting: %+v, received: %+v", expStats, rcvStats)
		}
	}
}

// Make sure balance was topped-up
// Bug reported by DigiDaz over IRC
func TestApierGetAccountAfterLoad(t *testing.T) {
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rater.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 11 {
		t.Errorf("Calling ApierV1.GetBalance expected: 11, received: %v \n\n for:%s", reply.BalanceMap[utils.MONETARY].GetTotalValue(), utils.ToJSON(reply))
	}
}

// Test here ResponderGetCost
func TestApierResponderGetCost(t *testing.T) {
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:30Z", "")
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "+4917621621391",
		DurationIndex: 90,
		TimeStart:     tStart,
		TimeEnd:       tEnd,
	}
	var cc engine.CallCost
	// Simple test that command is executed without errors
	if err := rater.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 90.0 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}
}

func TestApierMaxDebitInexistentAcnt(t *testing.T) {

	cc := &engine.CallCost{}
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		Category:    "call",
		Subject:     "INVALID",
		Account:     "INVALID",
		Destination: "1002",
		TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(time.Duration(10) * time.Second),
	}
	if err := rater.Call("Responder.MaxDebit", cd, cc); err == nil {
		t.Error(err.Error())
	}
	if err := rater.Call("Responder.Debit", cd, cc); err == nil {
		t.Error(err.Error())
	}

}

func TestApierCdrServer(t *testing.T) {
	httpClient := new(http.Client)
	cdrForm1 := url.Values{utils.OriginID: []string{"dsafdsaf"}, utils.OriginHost: []string{"192.168.1.1"}, utils.RequestType: []string{utils.META_RATED},
		utils.Tenant: []string{"cgrates.org"}, utils.Category: []string{"call"}, utils.Account: []string{"1001"}, utils.Subject: []string{"1001"}, utils.Destination: []string{"1002"},
		utils.SetupTime:  []string{"2013-11-07T08:42:22Z"},
		utils.AnswerTime: []string{"2013-11-07T08:42:26Z"}, utils.Usage: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	cdrForm2 := url.Values{utils.OriginID: []string{"adsafdsaf"}, utils.OriginHost: []string{"192.168.1.1"}, utils.RequestType: []string{utils.META_RATED},
		utils.Tenant: []string{"cgrates.org"}, utils.Category: []string{"call"}, utils.Account: []string{"1001"}, utils.Subject: []string{"1001"}, utils.Destination: []string{"1002"},
		utils.SetupTime:  []string{"2013-11-07T08:42:23Z"},
		utils.AnswerTime: []string{"2013-11-07T08:42:26Z"}, utils.Usage: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	for _, cdrForm := range []url.Values{cdrForm1, cdrForm2} {
		cdrForm.Set(utils.Source, utils.TEST_SQL)
		if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cdr_http", "127.0.0.1:2080"), cdrForm); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func TestApierITGetCdrs(t *testing.T) {
	var reply []*engine.ExternalCDR
	req := utils.AttrGetCdrs{MediationRunIds: []string{utils.MetaRaw}}
	if err := rater.Call("ApierV1.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	}
}

func TestApierITProcessCdr(t *testing.T) {
	var reply string
	cdr := engine.CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "test", RequestType: utils.META_RATED, Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001",
		Destination: "1002",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), RunID: utils.DEFAULT_RUNID,
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01,
	}
	if err := rater.Call("CdrsV1.ProcessCDR", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	req := utils.AttrGetCdrs{MediationRunIds: []string{utils.MetaRaw}}
	if err := rater.Call("ApierV1.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 3 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

// Test here ResponderGetCost
func TestApierGetCallCostLog(t *testing.T) {
	var cc engine.EventCost
	var attrs utils.AttrGetCallCost
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.GetEventCost", attrs, &cc); err == nil {
		t.Error("Failed to detect missing fields in ApierV1.GetCallCostLog")
	}
	attrs.CgrId = "dummyid"
	attrs.RunId = "default"
	if err := rater.Call("ApierV1.GetEventCost", attrs, &cc); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error("ApierV1.GetCallCostLog: should return NOT_FOUND, got:", err)
	}
	tm := time.Now().Truncate(time.Millisecond)
	cdr := &engine.CDR{
		CGRID:       "Cdr1",
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "OriginCDR1",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.META_RATED,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   tm,
		AnswerTime:  tm,
		RunID:       utils.DEFAULT_RUNID,
		Usage:       time.Duration(0),
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
	}
	var reply string
	if err := rater.Call("CdrsV1.ProcessCDR", cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(100 * time.Millisecond)
	expected := engine.EventCost{
		CGRID:     "Cdr1",
		RunID:     "*default",
		StartTime: tm,
		Usage:     utils.DurationPointer(0),
		Cost:      utils.Float64Pointer(0),
		Charges: []*engine.ChargingInterval{{
			RatingID:       "",
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
	attrs.CgrId = "Cdr1"
	attrs.RunId = ""
	if err := rater.Call("ApierV1.GetEventCost", attrs, &cc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cc) {
		t.Errorf("Expecting %s ,recived %s", utils.ToJSON(expected), utils.ToJSON(cc))
	}
}

func TestApierITSetDC(t *testing.T) {
	dcs1 := []*utils.DerivedCharger{
		{RunID: "extra1", RequestTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	attrs := AttrSetDerivedChargers{Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan", DerivedChargers: dcs1, Overwrite: true}
	var reply string
	if err := rater.Call("ApierV1.SetDerivedChargers", attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func TestApierITGetDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	eDcs := utils.DerivedChargers{DestinationIDs: utils.NewStringMap(),
		Chargers: []*utils.DerivedCharger{
			{RunID: "extra1", RequestTypeField: "^prepaid", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
				AccountField: "rif", SubjectField: "rif", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
			{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
				AccountField: "ivo", SubjectField: "ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		}}
	var dcs utils.DerivedChargers
	if err := rater.Call("ApierV1.GetDerivedChargers", attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, eDcs) {
		t.Errorf("Expecting: %v, received: %v", eDcs, dcs)
	}
}

func TestApierITRemDC(t *testing.T) {
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	var reply string
	if err := rater.Call("ApierV1.RemDerivedChargers", attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func TestApierITSetDestination(t *testing.T) {
	attrs := utils.AttrSetDestination{Id: "TEST_SET_DESTINATION", Prefixes: []string{"+4986517174963", "+4986517174960"}}
	var reply string
	if err := rater.Call("ApierV1.SetDestination", attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := rater.Call("ApierV1.SetDestination", attrs, &reply); err == nil || err.Error() != "EXISTS" { // Second time without overwrite should generate error
		t.Error("Unexpected error", err.Error())
	}
	attrs = utils.AttrSetDestination{Id: "TEST_SET_DESTINATION", Prefixes: []string{"+4986517174963", "+4986517174964"}, Overwrite: true}
	if err := rater.Call("ApierV1.SetDestination", attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	eDestination := engine.Destination{Id: attrs.Id, Prefixes: attrs.Prefixes}
	var rcvDestination engine.Destination
	if err := rater.Call("ApierV1.GetDestination", attrs.Id, &rcvDestination); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(eDestination, rcvDestination) {
		t.Errorf("Expecting: %+v, received: %+v", eDestination, rcvDestination)
	}
	eRcvIDs := []string{attrs.Id}
	var rcvIDs []string
	if err := rater.Call("ApierV1.GetReverseDestination", attrs.Prefixes[0], &rcvIDs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(eRcvIDs, rcvIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eRcvIDs, rcvIDs)
	}
}

func TestApierITGetAliases(t *testing.T) {
	var alias engine.Alias
	if err := rater.Call("AliasesV1.GetAlias", engine.Alias{Context: utils.MetaRating, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "2001", Subject: "2001"}, &alias); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestApierITAddRatingSubjectAliases(t *testing.T) {
	/*var reply string
	if err := rater.Call("ApierV1.FlushCache", utils.AttrReloadCache{FlushAll: true}, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ReloadCache got reply: ", reply)
	}
	*/
	addRtSubjAliases := &AttrAddRatingSubjectAliases{Tenant: "cgrates.org", Category: "call", Subject: "1001", Aliases: []string{"2001", "2002", "2003"}}
	var rply string
	if err := rater.Call("ApierV1.AddRatingSubjectAliases", addRtSubjAliases, &rply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if rply != utils.OK {
		t.Error("Unexpected reply: ", rply)
	}
	var alias engine.Alias
	for _, als := range addRtSubjAliases.Aliases {
		if err := rater.Call("AliasesV1.GetAlias", engine.Alias{Context: utils.MetaRating, Direction: "*out",
			Tenant: addRtSubjAliases.Tenant, Category: addRtSubjAliases.Category,
			Account: als, Subject: als}, &alias); err != nil {
			t.Error("Unexpected error", err.Error())
		}
	}
}

func TestApierITRemRatingSubjectAliases(t *testing.T) {
	tenantRatingSubj := engine.TenantRatingSubject{Tenant: "cgrates.org", Subject: "1001"}
	var rply string
	if err := rater.Call("ApierV1.RemRatingSubjectAliases", tenantRatingSubj, &rply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if rply != utils.OK {
		t.Error("Unexpected reply: ", rply)
	}
	var alias engine.Alias
	//al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Group
	if err := rater.Call("AliasesV1.GetAlias", engine.Alias{Context: utils.MetaRating, Direction: "*out",
		Tenant: "cgrates.org", Category: "call", Account: "2001", Subject: "2001"}, &alias); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error %v, alias: %+v, values: %+v", err, alias, alias.Values[0])
	}
}

func TestApierITAddAccountAliases(t *testing.T) {
	addAcntAliases := &AttrAddAccountAliases{Tenant: "cgrates.org", Category: "call", Account: "1001", Aliases: []string{"2001", "2002", "2003"}}
	var rply string
	if err := rater.Call("ApierV1.AddAccountAliases", addAcntAliases, &rply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if rply != utils.OK {
		t.Error("Unexpected reply: ", rply)
	}
	var alias engine.Alias
	for _, als := range addAcntAliases.Aliases {
		if err := rater.Call("AliasesV1.GetAlias", engine.Alias{Context: utils.MetaRating, Direction: "*out", Tenant: addAcntAliases.Tenant, Category: addAcntAliases.Category,
			Account: als, Subject: als}, &alias); err != nil {
			t.Error("Unexpected error", err.Error())
		}
	}
}

func TestApierITRemAccountAliases(t *testing.T) {
	tenantAcnt := engine.TenantAccount{Tenant: "cgrates.org", Account: "1001"}
	var rply string
	if err := rater.Call("ApierV1.RemAccountAliases", tenantAcnt, &rply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if rply != utils.OK {
		t.Error("Unexpected reply: ", rply)
	}
	var alias engine.Alias
	//al.Direction, al.Tenant, al.Category, al.Account, al.Subject, al.Group
	if err := rater.Call("AliasesV1.GetAlias", engine.Alias{Context: utils.MetaRating, Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "2001", Subject: "2001"}, &alias); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error %v, alias: %+v", err, alias)
	}
}

func TestApierITGetScheduledActions(t *testing.T) {
	var rply []*scheduler.ScheduledAction
	if err := rater.Call("ApierV1.GetScheduledActions", scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
		t.Error("Unexpected error: ", err)
	}
}

func TestApierITGetDataCost(t *testing.T) {
	attrs := AttrGetDataCost{Category: "data", Tenant: "cgrates.org",
		Subject: "1001", AnswerTime: "*now", Usage: 640113}
	var rply *engine.DataCost
	if err := rater.Call("ApierV1.GetDataCost", attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if rply.Cost != 128.0240 {
		t.Errorf("Unexpected cost received: %f", rply.Cost)
	}
}

func TestApierITGetCost(t *testing.T) {
	attrs := AttrGetCost{Category: "data", Tenant: "cgrates.org",
		Subject: "1001", AnswerTime: "*now", Usage: "640113"}
	var rply *engine.EventCost
	if err := rater.Call("ApierV1.GetCost", attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 128.0240 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

// Test LoadTPFromStorDb
func TestApierInitDataDb2(t *testing.T) {
	if err := engine.InitDataDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestApierInitStorDb2(t *testing.T) {
	if err := engine.InitStorDb(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestApierReloadCache2(t *testing.T) {
	reply := ""
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.FlushCache", utils.AttrReloadCache{FlushAll: true}, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ReloadCache got reply: ", reply)
	}
}

func TestApierReloadScheduler2(t *testing.T) {
	reply := ""
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.ReloadScheduler", reply, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadScheduler: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ReloadScheduler got reply: ", reply)
	}
}

func TestApierImportTPFromFolderPath(t *testing.T) {
	var reply string
	if err := rater.Call("ApierV1.ImportTariffPlanFromFolder",
		utils.AttrImportTPFromFolder{TPid: "TEST_TPID2",
			FolderPath: "/usr/share/cgrates/tariffplans/oldtutorial"}, &reply); err != nil {
		t.Error("Got error on ApierV1.ImportTarrifPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ImportTarrifPlanFromFolder got reply: ", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func TestApierLoadTariffPlanFromStorDbDryRun(t *testing.T) {
	var reply string
	if err := rater.Call("ApierV1.LoadTariffPlanFromStorDb",
		AttrLoadTpFromStorDb{TPid: "TEST_TPID2", DryRun: true}, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromStorDb: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.LoadTariffPlanFromStorDb got reply: ", reply)
	}
}

func TestApierGetCacheStats2(t *testing.T) {
	var rcvStats *utils.CacheStats
	var args utils.AttrCacheStats
	err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats)
	expectedStats := new(utils.CacheStats)
	if err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

func TestApierLoadTariffPlanFromStorDb(t *testing.T) {
	var reply string
	if err := rater.Call("ApierV1.LoadTariffPlanFromStorDb",
		AttrLoadTpFromStorDb{TPid: "TEST_TPID2"}, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromStorDb: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.LoadTariffPlanFromStorDb got reply: ", reply)
	}
}

func TestApierStartStopServiceStatus(t *testing.T) {
	var reply string
	if err := rater.Call("ApierV1.ServiceStatus", servmanager.ArgStartService{ServiceID: utils.MetaScheduler},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.RunningCaps {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call("ApierV1.StopService", servmanager.ArgStartService{ServiceID: "INVALID"},
		&reply); err == nil || err.Error() != utils.UnsupportedServiceIDCaps {
		t.Error(err)
	}
	if err := rater.Call("ApierV1.StopService", servmanager.ArgStartService{ServiceID: utils.MetaScheduler},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call("ApierV1.ServiceStatus", servmanager.ArgStartService{ServiceID: utils.MetaScheduler},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.StoppedCaps {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call("ApierV1.StartService", servmanager.ArgStartService{ServiceID: utils.MetaScheduler},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call("ApierV1.ServiceStatus", servmanager.ArgStartService{ServiceID: utils.MetaScheduler},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.RunningCaps {
		t.Errorf("Received: <%s>", reply)
	}
	if err := rater.Call("ApierV1.ReloadScheduler", reply, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadScheduler: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ReloadScheduler got reply: ", reply)
	}
}

func TestApierReplayFailedPosts(t *testing.T) {
	fileName := "act>*call_url|*http_json|http%3A%2F%2Flocalhost%3A2081|63bed4ea-615e-4096-b1f4-499f64f29b28.json"
	fileContent := []byte(`{"ID":"cgrates.org:1007","BalanceMap":{"*monetary":[{"Uuid":"367be35a-96ee-40a5-b609-9130661f5f12","ID":"","Value":0,"Directions":{"*out":true},"ExpirationDate":"0001-01-01T00:00:00Z","Weight":10,"DestinationIDs":{},"RatingSubject":"","Categories":{},"SharedGroups":{"SHARED_A":true},"Timings":null,"TimingIDs":{},"Disabled":false,"Factor":null,"Blocker":false}]},"UnitCounters":{"*monetary":[{"CounterType":"*event","Counters":[{"Value":0,"Filter":{"Uuid":null,"ID":"b8531413-10d5-47ad-81ad-2bc272e8f0ca","Type":"*monetary","Value":null,"Directions":{"*out":true},"ExpirationDate":null,"Weight":null,"DestinationIDs":{"FS_USERS":true},"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null}}]}]},"ActionTriggers":[{"ID":"STANDARD_TRIGGERS","UniqueID":"46ac7b8c-685d-4555-bf73-fa6cfbc2fa21","ThresholdType":"*min_balance","ThresholdValue":2,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"Directions":{"*out":true},"ExpirationDate":null,"Weight":null,"DestinationIDs":null,"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"LOG_WARNING","MinQueuedItems":0,"Executed":true,"LastExecutionTime":"2017-01-31T14:03:57.961651647+01:00"},{"ID":"STANDARD_TRIGGERS","UniqueID":"b8531413-10d5-47ad-81ad-2bc272e8f0ca","ThresholdType":"*max_event_counter","ThresholdValue":5,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"Directions":{"*out":true},"ExpirationDate":null,"Weight":null,"DestinationIDs":{"FS_USERS":true},"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"LOG_WARNING","MinQueuedItems":0,"Executed":false,"LastExecutionTime":"0001-01-01T00:00:00Z"},{"ID":"STANDARD_TRIGGERS","UniqueID":"8b424186-7a31-4aef-99c5-35e12e6fed41","ThresholdType":"*max_balance","ThresholdValue":20,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"Directions":{"*out":true},"ExpirationDate":null,"Weight":null,"DestinationIDs":null,"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"LOG_WARNING","MinQueuedItems":0,"Executed":false,"LastExecutionTime":"0001-01-01T00:00:00Z"},{"ID":"STANDARD_TRIGGERS","UniqueID":"28557f3b-139c-4a27-9d17-bda1f54b7c19","ThresholdType":"*max_balance","ThresholdValue":100,"Recurrent":false,"MinSleep":0,"ExpirationDate":"0001-01-01T00:00:00Z","ActivationDate":"0001-01-01T00:00:00Z","Balance":{"Uuid":null,"ID":null,"Type":"*monetary","Value":null,"Directions":{"*out":true},"ExpirationDate":null,"Weight":null,"DestinationIDs":null,"RatingSubject":null,"Categories":null,"SharedGroups":null,"TimingIDs":null,"Timings":null,"Disabled":null,"Factor":null,"Blocker":null},"Weight":10,"ActionsID":"DISABLE_AND_LOG","MinQueuedItems":0,"Executed":false,"LastExecutionTime":"0001-01-01T00:00:00Z"}],"AllowNegative":false,"Disabled":false}"`)
	args := ArgsReplyFailedPosts{
		FailedRequestsInDir:  utils.StringPointer("/tmp/TestsApierV1/in"),
		FailedRequestsOutDir: utils.StringPointer("/tmp/TestsApierV1/out"),
	}
	for _, dir := range []string{*args.FailedRequestsInDir, *args.FailedRequestsOutDir} {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("Error %s removing folder: %s", err, dir)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Errorf("Error %s creating folder: %s", err, dir)
		}
	}
	fileOut, err := os.Create(path.Join(*args.FailedRequestsInDir, fileName))
	if err != nil {
		t.Error(err)
	}
	if _, err := fileOut.Write(fileContent); err != nil {
		t.Error(err)
	}
	fileOut.Close()
	var reply string
	if err := rater.Call("ApierV1.ReplayFailedPosts", args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply: ", reply)
	}
	outPath := path.Join(*args.FailedRequestsOutDir, fileName)
	if outContent, err := ioutil.ReadFile(outPath); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fileContent, outContent) {
		t.Errorf("Expecting: %q, received: %q", string(fileContent), string(outContent))
	}
	fileName = "cdr|*amqp_json_map|amqp%3A%2F%2Fguest%3Aguest%40localhost%3A5672%2F%3Fqueue_id%3Dcgrates_cdrs|ae8cc4b3-5e60-4396-b82a-64b96a72a03c.json"
	fileContent = []byte(`{"CGRID":"88ed9c38005f07576a1e1af293063833b60edcc6"}`)
	fileInPath := path.Join(*args.FailedRequestsInDir, fileName)
	fileOut, err = os.Create(fileInPath)
	if err != nil {
		t.Error(err)
	}
	if _, err := fileOut.Write(fileContent); err != nil {
		t.Error(err)
	}
	fileOut.Close()
	if err := rater.Call("ApierV1.ReplayFailedPosts", args, &reply); err != nil {
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
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
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
	case <-time.After(time.Duration(100 * time.Millisecond)):
		t.Error("No message received from RabbitMQ")
	}
	for _, dir := range []string{*args.FailedRequestsInDir, *args.FailedRequestsOutDir} {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("Error %s removing folder: %s", err, dir)
		}
	}
}

func TestApierGetDataDBVesions(t *testing.T) {
	var reply *engine.Versions
	if err := rater.Call("ApierV1.GetDataDBVersions", "", &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(engine.CurrentDataDBVersions(), *reply) {
		t.Errorf("Expecting : %+v, received: %+v", engine.CurrentDataDBVersions(), *reply)
	}
}

func TestApierGetStorDBVesions(t *testing.T) {
	var reply *engine.Versions
	if err := rater.Call("ApierV1.GetStorDBVersions", "", &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(engine.CurrentStorDBVersions(), *reply) {
		t.Errorf("Expecting : %+v, received: %+v", engine.CurrentStorDBVersions(), *reply)
	}
}

func TestApierPing(t *testing.T) {
	var reply string
	for _, method := range []string{utils.StatSv1Ping, utils.ResourceSv1Ping,
		utils.SupplierSv1Ping, utils.ThresholdSv1Ping, utils.AttributeSv1Ping} {
		if err := rater.Call(method, "", &reply); err != nil {
			t.Error(err)
		} else if reply != utils.Pong {
			t.Errorf("Received: %s", reply)
		}
	}
}

// Simply kill the engine after we are done with tests within this file
func TestApierStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
