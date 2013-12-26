/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package apier

import (
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"
	"testing"
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
  * Execute remote Apis and test their replies(follow prepaid1cent scenario so we can test load in dataDb also).
*/

var cfg *config.CGRConfig
var rater *rpc.Client

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates/data", "CGR data dir path here")
var storDbType = flag.String("stordb_type", "mysql", "The type of the storDb database <mysql>")
var waitRater = flag.Int("wait_rater", 300, "Number of miliseconds to wait for rater to start and cache")

func init() {
	cfgPath := path.Join(*dataDir, "conf", "cgrates.cfg")
	cfg, _ = config.NewCGRConfig(&cfgPath)
}

// Empty tables before using them
func TestCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	var mysql *engine.MySQLStorage
	if d, err := engine.NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass); err != nil {
		t.Fatal("Error on opening database connection: ", err)
	} else {
		mysql = d.(*engine.MySQLStorage)
	}
	for _, scriptName := range []string{engine.CREATE_CDRS_TABLES_SQL, engine.CREATE_COSTDETAILS_TABLES_SQL, engine.CREATE_MEDIATOR_TABLES_SQL, engine.CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := mysql.CreateTablesFromScript(path.Join(*dataDir, "storage", *storDbType, scriptName)); err != nil {
			t.Fatal("Error on mysql creation: ", err.Error())
			return // No point in going further
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := mysql.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			t.Fatal(err.Error())
		}
	}
}

func TestInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	ratingDb, err := engine.ConfigureRatingStorage(cfg.RatingDBType, cfg.RatingDBHost, cfg.RatingDBPort, cfg.RatingDBName, cfg.RatingDBUser, cfg.RatingDBPass, cfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	accountDb, err := engine.ConfigureAccountingStorage(cfg.AccountDBType, cfg.AccountDBHost, cfg.AccountDBPort, cfg.AccountDBName,
		cfg.AccountDBUser, cfg.AccountDBPass, cfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	for _, db := range []engine.Storage{ratingDb, accountDb} {
		if err := db.Flush(); err != nil {
			t.Fatal("Cannot reset dataDb", err)
		}
	}
}

// Finds cgr-engine executable and starts it with default configuration
func TestStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("Cannot find cgr-engine executable")
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	engine := exec.Command(enginePath, "-rater", "-scheduler", "-config", path.Join(*dataDir, "conf", "cgrates.cfg"))
	if err := engine.Start(); err != nil {
		t.Fatal("Cannot start cgr-engine: ", err.Error())
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time to rater to fire up
}

// Connect rpc client to rater
func TestRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	if cfg.RPCEncoding == utils.JSON {
		rater, err = jsonrpc.Dial("tcp", cfg.MediatorRater)
	} else {
		rater, err = rpc.Dial("tcp", cfg.MediatorRater)
	}
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Test here TPTiming APIs
func TestApierTPTiming(t *testing.T) {
	if !*testLocal {
		return
	}
	// ALWAYS,*any,*any,*any,*any,00:00:00
	tmAlways := &utils.ApierTPTiming{TPid: engine.TEST_SQL,
		TimingId:  "ALWAYS",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "*any",
		Time:      "00:00:00",
	}
	tmAlways2 := new(utils.ApierTPTiming)
	*tmAlways2 = *tmAlways
	tmAlways2.TimingId = "ALWAYS2"
	tmAsap := &utils.ApierTPTiming{TPid: engine.TEST_SQL,
		TimingId:  "ASAP",
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid TimingId Years Months MonthDays WeekDays Time]" {
		t.Error("Calling ApierV1.SetTPTiming got unexpected error: ", err.Error())
	}
	// Test get
	var rplyTmAlways2 *utils.ApierTPTiming
	if err := rater.Call("ApierV1.GetTPTiming", AttrGetTPTiming{tmAlways2.TPid, tmAlways2.TimingId}, &rplyTmAlways2); err != nil {
		t.Error("Calling ApierV1.GetTPTiming, got error: ", err.Error())
	} else if !reflect.DeepEqual(tmAlways2, rplyTmAlways2) {
		t.Errorf("Calling ApierV1.GetTPTiming expected: %v, received: %v", tmAlways, rplyTmAlways2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPTiming", AttrGetTPTiming{tmAlways2.TPid, tmAlways2.TimingId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPTiming, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPTiming received: ", reply)
	}
	// Test getIds
	var rplyTmIds []string
	expectedTmIds := []string{"ALWAYS", "ASAP"}
	if err := rater.Call("ApierV1.GetTPTimingIds", AttrGetTPTimingIds{tmAlways.TPid}, &rplyTmIds); err != nil {
		t.Error("Calling ApierV1.GetTPTimingIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedTmIds, rplyTmIds) {
		t.Errorf("Calling ApierV1.GetTPTimingIds expected: %v, received: %v", expectedTmIds, rplyTmIds)
	}
}

// Test here TPTiming APIs
func TestApierTPDestination(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	dstDe := &utils.TPDestination{TPid: engine.TEST_SQL, DestinationId: "GERMANY", Prefixes: []string{"+49"}}
	dstDeMobile := &utils.TPDestination{TPid: engine.TEST_SQL, DestinationId: "GERMANY_MOBILE", Prefixes: []string{"+4915", "+4916", "+4917"}}
	dstFs := &utils.TPDestination{TPid: engine.TEST_SQL, DestinationId: "FS_USERS", Prefixes: []string{"10"}}
	dstDe2 := new(utils.TPDestination)
	*dstDe2 = *dstDe // Data which we use for remove, still keeping the sample data to check proper loading
	dstDe2.DestinationId = "GERMANY2"
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid DestinationId Prefixes]" {
		t.Error("Calling ApierV1.SetTPDestination got unexpected error: ", err.Error())
	}
	// Test get
	var rplyDstDe2 *utils.TPDestination
	if err := rater.Call("ApierV1.GetTPDestination", AttrGetTPDestination{dstDe2.TPid, dstDe2.DestinationId}, &rplyDstDe2); err != nil {
		t.Error("Calling ApierV1.GetTPDestination, got error: ", err.Error())
	} else if !reflect.DeepEqual(dstDe2, rplyDstDe2) {
		t.Errorf("Calling ApierV1.GetTPDestination expected: %v, received: %v", dstDe2, rplyDstDe2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPDestination", AttrGetTPDestination{dstDe2.TPid, dstDe2.DestinationId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPTiming, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPTiming received: ", reply)
	}
	// Test getIds
	var rplyDstIds []string
	expectedDstIds := []string{"FS_USERS", "GERMANY", "GERMANY_MOBILE"}
	if err := rater.Call("ApierV1.GetTPDestinationIds", AttrGetTPDestinationIds{dstDe.TPid}, &rplyDstIds); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedDstIds, rplyDstIds) {
		t.Errorf("Calling ApierV1.GetTPDestinationIds expected: %v, received: %v", expectedDstIds, rplyDstIds)
	}
}

// Test here TPRate APIs
func TestApierTPRate(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	rt := &utils.TPRate{TPid: engine.TEST_SQL, RateId: "RT_FS_USERS", RateSlots: []*utils.RateSlot{
		&utils.RateSlot{ConnectFee: 0, Rate: 0, RateUnit: "60s", RateIncrement: "60s", GroupIntervalStart: "0s", RoundingMethod: "*up", RoundingDecimals: 0},
	}}
	rt2 := new(utils.TPRate)
	*rt2 = *rt
	rt2.RateId = "RT_FS_USERS2"
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid RateId RateSlots]" {
		t.Error("Calling ApierV1.SetTPRate got unexpected error: ", err.Error())
	}
	// Test get
	var rplyRt2 *utils.TPRate
	if err := rater.Call("ApierV1.GetTPRate", AttrGetTPRate{rt2.TPid, rt2.RateId}, &rplyRt2); err != nil {
		t.Error("Calling ApierV1.GetTPRate, got error: ", err.Error())
	} else if !reflect.DeepEqual(rt2, rplyRt2) {
		t.Errorf("Calling ApierV1.GetTPRate expected: %v, received: %v", rt2, rplyRt2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPRate", AttrGetTPRate{rt2.TPid, rt2.RateId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRate, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRate received: ", reply)
	}
	// Test getIds
	var rplyRtIds []string
	expectedRtIds := []string{"RT_FS_USERS"}
	if err := rater.Call("ApierV1.GetTPRateIds", AttrGetTPRateIds{rt.TPid}, &rplyRtIds); err != nil {
		t.Error("Calling ApierV1.GetTPRateIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRtIds, rplyRtIds) {
		t.Errorf("Calling ApierV1.GetTPDestinationIds expected: %v, received: %v", expectedRtIds, rplyRtIds)
	}
}

// Test here TPDestinationRate APIs
func TestApierTPDestinationRate(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	dr := &utils.TPDestinationRate{TPid: engine.TEST_SQL, DestinationRateId: "DR_FREESWITCH_USERS", DestinationRates: []*utils.DestinationRate{
		&utils.DestinationRate{DestinationId: "FS_USERS", RateId: "RT_FS_USERS"},
	}}
	drDe := &utils.TPDestinationRate{TPid: engine.TEST_SQL, DestinationRateId: "DR_FREESWITCH_USERS", DestinationRates: []*utils.DestinationRate{
		&utils.DestinationRate{DestinationId: "GERMANY_MOBILE", RateId: "RT_FS_USERS"},
	}}
	dr2 := new(utils.TPDestinationRate)
	*dr2 = *dr
	dr2.DestinationRateId = engine.TEST_SQL
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid DestinationRateId DestinationRates]" {
		t.Error("Calling ApierV1.SetTPDestinationRate got unexpected error: ", err.Error())
	}
	// Test get
	var rplyDr2 *utils.TPDestinationRate
	if err := rater.Call("ApierV1.GetTPDestinationRate", AttrGetTPDestinationRate{dr2.TPid, dr2.DestinationRateId}, &rplyDr2); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationRate, got error: ", err.Error())
	} else if !reflect.DeepEqual(dr2, rplyDr2) {
		t.Errorf("Calling ApierV1.GetTPDestinationRate expected: %v, received: %v", dr2, rplyDr2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPDestinationRate", AttrGetTPDestinationRate{dr2.TPid, dr2.DestinationRateId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRate, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRate received: ", reply)
	}
	// Test getIds
	var rplyDrIds []string
	expectedDrIds := []string{"DR_FREESWITCH_USERS"}
	if err := rater.Call("ApierV1.GetTPDestinationRateIds", AttrTPDestinationRateIds{dr.TPid}, &rplyDrIds); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationRateIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedDrIds, rplyDrIds) {
		t.Errorf("Calling ApierV1.GetTPDestinationRateIds expected: %v, received: %v", expectedDrIds, rplyDrIds)
	}
}

// Test here TPRatingPlan APIs
func TestApierTPRatingPlan(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	rp := &utils.TPRatingPlan{TPid: engine.TEST_SQL, RatingPlanId: "RETAIL1", RatingPlanBindings: []*utils.TPRatingPlanBinding{
		&utils.TPRatingPlanBinding{DestinationRatesId: "DR_FREESWITCH_USERS", TimingId: "ALWAYS", Weight: 10},
	}}
	rpTst := new(utils.TPRatingPlan)
	*rpTst = *rp
	rpTst.RatingPlanId = engine.TEST_SQL
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid RatingPlanId RatingPlanBindings]" {
		t.Error("Calling ApierV1.SetTPRatingPlan got unexpected error: ", err.Error())
	}
	// Test get
	var rplyRpTst *utils.TPRatingPlan
	if err := rater.Call("ApierV1.GetTPRatingPlan", AttrGetTPRatingPlan{rpTst.TPid, rpTst.RatingPlanId}, &rplyRpTst); err != nil {
		t.Error("Calling ApierV1.GetTPRatingPlan, got error: ", err.Error())
	} else if !reflect.DeepEqual(rpTst, rplyRpTst) {
		t.Errorf("Calling ApierV1.GetTPRatingPlan expected: %v, received: %v", rpTst, rplyRpTst)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPRatingPlan", AttrGetTPRatingPlan{rpTst.TPid, rpTst.RatingPlanId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRatingPlan, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRatingPlan received: ", reply)
	}
	// Test getIds
	var rplyRpIds []string
	expectedRpIds := []string{"RETAIL1"}
	if err := rater.Call("ApierV1.GetTPRatingPlanIds", AttrGetTPRatingPlanIds{rp.TPid}, &rplyRpIds); err != nil {
		t.Error("Calling ApierV1.GetTPRatingPlanIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRpIds, rplyRpIds) {
		t.Errorf("Calling ApierV1.GetTPRatingPlanIds expected: %v, received: %v", expectedRpIds, rplyRpIds)
	}
}

// Test here TPRatingPlan APIs
func TestApierTPRatingProfile(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	rpf := &utils.TPRatingProfile{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org", TOR: "call", Direction: "*out", Subject: "*any",
		RatingPlanActivations: []*utils.TPRatingActivation{
			&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1", FallbackSubjects: ""},
		}}
	rpfTst := new(utils.TPRatingProfile)
	*rpfTst = *rpf
	rpfTst.Subject = engine.TEST_SQL
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid LoadId Tenant TOR Direction Subject RatingPlanActivations]" {
		t.Error("Calling ApierV1.SetTPRatingProfile got unexpected error: ", err.Error())
	}
	// Test get
	var rplyRpfs []*utils.TPRatingProfile
	if err := rater.Call("ApierV1.GetTPRatingProfiles", rpfTst, &rplyRpfs); err != nil {
		t.Error("Calling ApierV1.GetTPRatingProfiles, got error: ", err.Error())
	} else if !reflect.DeepEqual(rpfTst, rplyRpfs[0]) {
		t.Errorf("Calling ApierV1.GetTPRatingProfiles expected: %v, received: %v", rpfTst, rplyRpfs[0])
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPRatingProfile", rpfTst, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPRatingProfile, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPRatingProfile received: ", reply)
	}
	// Test getLoadIds
	var rplyRpIds []string
	expectedRpIds := []string{engine.TEST_SQL}
	if err := rater.Call("ApierV1.GetTPRatingProfileLoadIds", utils.AttrTPRatingProfileIds{TPid: rpf.TPid}, &rplyRpIds); err != nil {
		t.Error("Calling ApierV1.GetTPRatingProfileLoadIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRpIds, rplyRpIds) {
		t.Errorf("Calling ApierV1.GetTPRatingProfileLoadIds expected: %v, received: %v", expectedRpIds, rplyRpIds)
	}
}

func TestApierTPActions(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	act := &utils.TPActions{TPid: engine.TEST_SQL, ActionsId: "PREPAID_10", Actions: []*utils.TPAction{
		&utils.TPAction{Identifier: "*topup_reset", BalanceType: "*monetary", Direction: "*out", Units: 10, ExpiryTime: "*unlimited",
			DestinationId: "*any", BalanceWeight: 10, Weight: 10},
	}}
	actWarn := &utils.TPActions{TPid: engine.TEST_SQL, ActionsId: "WARN_VIA_HTTP", Actions: []*utils.TPAction{
		&utils.TPAction{Identifier: "*call_url", ExtraParameters: "http://localhost:8000", Weight: 10},
	}}
	actTst := new(utils.TPActions)
	*actTst = *act
	actTst.ActionsId = engine.TEST_SQL
	for _, ac := range []*utils.TPActions{act, actWarn, actTst} {
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid ActionsId Actions]" {
		t.Error("Calling ApierV1.SetTPActions got unexpected error: ", err.Error())
	}
	// Test get
	var rplyActs *utils.TPActions
	if err := rater.Call("ApierV1.GetTPActions", AttrGetTPActions{TPid: actTst.TPid, ActionsId: actTst.ActionsId}, &rplyActs); err != nil {
		t.Error("Calling ApierV1.GetTPActions, got error: ", err.Error())
	} else if !reflect.DeepEqual(actTst, rplyActs) {
		t.Errorf("Calling ApierV1.GetTPActions expected: %v, received: %v", actTst, rplyActs)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPActions", AttrGetTPActions{TPid: actTst.TPid, ActionsId: actTst.ActionsId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPActions, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPActions received: ", reply)
	}
	// Test getIds
	var rplyIds []string
	expectedIds := []string{"PREPAID_10", "WARN_VIA_HTTP"}
	if err := rater.Call("ApierV1.GetTPActionIds", AttrGetTPActionIds{TPid: actTst.TPid}, &rplyIds); err != nil {
		t.Error("Calling ApierV1.GetTPActionIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedIds, rplyIds) {
		t.Errorf("Calling ApierV1.GetTPActionIds expected: %v, received: %v", expectedIds, rplyIds)
	}
}

func TestApierTPActionTimings(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	at := &utils.TPActionTimings{TPid: engine.TEST_SQL, ActionTimingsId: "PREPAID_10", ActionTimings: []*utils.TPActionTiming{
		&utils.TPActionTiming{ActionsId: "PREPAID_10", TimingId: "ASAP", Weight: 10},
	}}
	atTst := new(utils.TPActionTimings)
	*atTst = *at
	atTst.ActionTimingsId = engine.TEST_SQL
	for _, act := range []*utils.TPActionTimings{at, atTst} {
		if err := rater.Call("ApierV1.SetTPActionTimings", act, &reply); err != nil {
			t.Error("Got error on ApierV1.SetTPActionTimings: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPActionTimings: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPActionTimings", atTst, &reply); err != nil {
		t.Error("Got error on second ApierV1.SetTPActionTimings: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPActionTimings got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPActionTimings", new(utils.TPActionTimings), &reply); err == nil {
		t.Error("Calling ApierV1.SetTPActionTimings, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid ActionTimingsId ActionTimings]" {
		t.Error("Calling ApierV1.SetTPActionTimings got unexpected error: ", err.Error())
	}
	// Test get
	var rplyActs *utils.TPActionTimings
	if err := rater.Call("ApierV1.GetTPActionTimings", AttrGetTPActionTimings{TPid: atTst.TPid, ActionTimingsId: atTst.ActionTimingsId}, &rplyActs); err != nil {
		t.Error("Calling ApierV1.GetTPActionTimings, got error: ", err.Error())
	} else if !reflect.DeepEqual(atTst, rplyActs) {
		t.Errorf("Calling ApierV1.GetTPActionTimings expected: %v, received: %v", atTst, rplyActs)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPActionTimings", AttrGetTPActionTimings{TPid: atTst.TPid, ActionTimingsId: atTst.ActionTimingsId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPActionTimings, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPActionTimings received: ", reply)
	}
	// Test getIds
	var rplyIds []string
	expectedIds := []string{"PREPAID_10"}
	if err := rater.Call("ApierV1.GetTPActionTimingIds", AttrGetTPActionTimingIds{TPid: atTst.TPid}, &rplyIds); err != nil {
		t.Error("Calling ApierV1.GetTPActionTimingIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedIds, rplyIds) {
		t.Errorf("Calling ApierV1.GetTPActionTimingIds expected: %v, received: %v", expectedIds, rplyIds)
	}
}

func TestApierTPActionTriggers(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	at := &utils.TPActionTriggers{TPid: engine.TEST_SQL, ActionTriggersId: "STANDARD_TRIGGERS", ActionTriggers: []*utils.TPActionTrigger{
		&utils.TPActionTrigger{BalanceType: "*monetary", Direction: "*out", ThresholdType: "*min_balance", ThresholdValue: 2, ActionsId: "LOG_BALANCE", Weight: 10},
	}}
	atTst := new(utils.TPActionTriggers)
	*atTst = *at
	atTst.ActionTriggersId = engine.TEST_SQL
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid ActionTriggersId]" {
		t.Error("Calling ApierV1.SetTPActionTriggers got unexpected error: ", err.Error())
	}
	// Test get
	var rplyActs *utils.TPActionTriggers
	if err := rater.Call("ApierV1.GetTPActionTriggers", AttrGetTPActionTriggers{TPid: atTst.TPid, ActionTriggersId: atTst.ActionTriggersId}, &rplyActs); err != nil {
		t.Error("Calling ApierV1.GetTPActionTriggers, got error: ", err.Error())
	} else if !reflect.DeepEqual(atTst, rplyActs) {
		t.Errorf("Calling ApierV1.GetTPActionTriggers expected: %v, received: %v", atTst, rplyActs)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPActionTriggers", AttrGetTPActionTriggers{TPid: atTst.TPid, ActionTriggersId: atTst.ActionTriggersId}, &reply); err != nil {
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
	if !*testLocal {
		return
	}
	reply := ""
	aa1 := &utils.TPAccountActions{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org",
		Account: "1001", Direction: "*out", ActionTimingsId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa2 := &utils.TPAccountActions{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org",
		Account: "1002", Direction: "*out", ActionTimingsId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa3 := &utils.TPAccountActions{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org",
		Account: "1003", Direction: "*out", ActionTimingsId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa4 := &utils.TPAccountActions{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org",
		Account: "1004", Direction: "*out", ActionTimingsId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aa5 := &utils.TPAccountActions{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org",
		Account: "1005", Direction: "*out", ActionTimingsId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	aaTst := new(utils.TPAccountActions)
	*aaTst = *aa1
	aaTst.Account = engine.TEST_SQL
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
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid LoadId Tenant Account Direction ActionTimingsId ActionTriggersId]" {
		t.Error("Calling ApierV1.SetTPAccountActions got unexpected error: ", err.Error())
	}
	// Test get
	var rplyaas []*utils.TPAccountActions
	if err := rater.Call("ApierV1.GetTPAccountActions", aaTst, &rplyaas); err != nil {
		t.Error("Calling ApierV1.GetTPAccountActions, got error: ", err.Error())
	} else if !reflect.DeepEqual(aaTst, rplyaas[0]) {
		t.Errorf("Calling ApierV1.GetTPAccountActions expected: %v, received: %v", aaTst, rplyaas[0])
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPAccountActions", aaTst, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPAccountActions, got error: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.RemTPAccountActions received: ", reply)
	}
	// Test getLoadIds
	var rplyRpIds []string
	expectedRpIds := []string{engine.TEST_SQL}
	if err := rater.Call("ApierV1.GetTPAccountActionLoadIds", AttrGetTPAccountActionIds{TPid: aaTst.TPid}, &rplyRpIds); err != nil {
		t.Error("Calling ApierV1.GetTPAccountActionLoadIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedRpIds, rplyRpIds) {
		t.Errorf("Calling ApierV1.GetTPAccountActionLoadIds expected: %v, received: %v", expectedRpIds, rplyRpIds)
	}
}

// Test here LoadRatingPlan
func TestApierLoadRatingPlan(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	if err := rater.Call("ApierV1.LoadRatingPlan", AttrLoadRatingPlan{TPid: engine.TEST_SQL, RatingPlanId: "RETAIL1"}, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadRatingPlan: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadRatingPlan got reply: ", reply)
	}
}

// Test here LoadRatingProfile
func TestApierLoadRatingProfile(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	rpf := &utils.TPRatingProfile{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org", TOR: "call", Direction: "*out", Subject: "*any"}
	if err := rater.Call("ApierV1.LoadRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadRatingProfile got reply: ", reply)
	}
}

// Test here SetRatingProfile
func TestApierSetRatingProfile(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	rpa := &utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1", FallbackSubjects: "dan2;*any"}
	rpf := &AttrSetRatingProfile{Tenant: "cgrates.org", TOR: "call", Direction: "*out", Subject: "dan", RatingPlanActivations: []*utils.TPRatingActivation{rpa}}
	if err := rater.Call("ApierV1.SetRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.SetRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetRatingProfile got reply: ", reply)
	}
	// Calling the second time should raise EXISTS
	if err := rater.Call("ApierV1.SetRatingProfile", rpf, &reply); err == nil || err.Error() != "EXISTS"{
		t.Error("Unexpected result on duplication: ", err.Error())
	}
}

// Test here SetAccountActions
func TestApierSetAccountActions(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	aa1 := &utils.TPAccountActions{TPid: engine.TEST_SQL, LoadId: engine.TEST_SQL, Tenant: "cgrates.org", Account: "1001", Direction: "*out"}
	if err := rater.Call("ApierV1.SetAccountActions", aa1, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccountActions: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetAccountActions got reply: ", reply)
	}
}

// Test here ReloadScheduler
func TestApierReloadScheduler(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.ReloadScheduler", reply, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadScheduler: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.ReloadScheduler got reply: ", reply)
	}
}

// Test here ReloadCache
func TestApierReloadCache(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	arc := new(utils.ApiReloadCache)
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.ReloadCache", arc, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadCache: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.ReloadCache got reply: ", reply)
	}
}

func TestApierGetCacheStats(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations:4, RatingPlans: 1, RatingProfiles: 2, Actions: 1}
	var args utils.AttrCacheStats
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

func TestApierGetCachedItemAge(t *testing.T) {
	if !*testLocal {
		return
	}
	var rcvAge *time.Duration
	qryData := &utils.AttrCachedItemAge{Category: strings.TrimSuffix(utils.DESTINATIONS_CSV, ".csv"), ItemId: "+4917"} // Destinations are cached per prefix not id
	if err := rater.Call("ApierV1.GetCachedItemAge", qryData, &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if *rcvAge > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
	qryData = &utils.AttrCachedItemAge{Category: strings.TrimSuffix(utils.RATING_PLANS_CSV, ".csv"), ItemId: "RETAIL1"}
	if err := rater.Call("ApierV1.GetCachedItemAge", qryData, &rcvAge); err != nil {
		t.Error("Got error on ApierV1.GetCachedItemAge: ", err.Error())
	} else if *rcvAge > time.Duration(2)*time.Second {
		t.Errorf("Cache too old: %d", rcvAge)
	}
}

// Test here GetDestination
func TestApierGetDestination(t *testing.T) {
	if !*testLocal {
		return
	}
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
	if !*testLocal {
		return
	}
	reply := new(engine.RatingPlan)
	rplnId := "RETAIL1"
	//{"Id":"RETAIL1","Timings":{"96c78ff5":{"Years":[],"Months":[],"MonthDays":[],"WeekDays":[],"StartTime":"00:00:00","EndTime":""}},"Ratings":{"e41ffcf2":{"ConnectFee":0,"Rates":[{"GroupIntervalStart":0,"Value":0,"RateIncrement":60000000000,"RateUnit":60000000000}],"RoundingMethod":"*up","RoundingDecimals":0}},"DestinationRates":{"FS_USERS":[{"Timing":"96c78ff5","Rating":"e41ffcf2","Weight":10}],"GERMANY_MOBILE":[{"Timing":"96c78ff5","Rating":"e41ffcf2","Weight":10}]}
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
	/*
		riTiming := &engine.RITiming{StartTime: "00:00:00"}
		for _, tm := range reply.Timings { // We only get one loop
			if  !reflect.DeepEqual(tm, riTiming) {
				t.Errorf("Unexpected timings value: %v, expecting: %v", tm, riTiming)
			}
		}
	*/
	riRate := &engine.RIRate{ConnectFee: 0, RoundingMethod: "*up", RoundingDecimals: 0, Rates: []*engine.Rate{
		&engine.Rate{GroupIntervalStart: 0, Value: 0, RateIncrement: time.Duration(60) * time.Second, RateUnit: time.Duration(60) * time.Second},
	}}
	for _, rating := range reply.Ratings {
		if !reflect.DeepEqual(rating, riRate) {
			t.Errorf("Unexpected riRate received: %v", rating)
		}
	}
}

// Test here AddBalance
func TestApierAddBalance(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	attrs := &AttrAddBalance{Tenant: "cgrates.org", Account: "1001", BalanceId: "*monetary", Direction: "*out", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan", BalanceId: "*monetary", Direction: "*out", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	attrs = &AttrAddBalance{Tenant: "cgrates.org", Account: "dan2", BalanceId: "*monetary", Direction: "*out", Value: 1.5}
	if err := rater.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
}

// Test here ExecuteAction
func TestApierExecuteAction(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Add balance to a previously known account
	attrs := AttrExecuteAction{Direction: "*out", Tenant: "cgrates.org", Account: "dan2", ActionsId: "PREPAID_10"}
	if err := rater.Call("ApierV1.ExecuteAction", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}
	reply2 := ""
	// Add balance to an account which does n exist
	attrs = AttrExecuteAction{Direction: "*out", Tenant: "cgrates.org", Account: "dan2", ActionsId: "DUMMY_ACTION"}
	if err := rater.Call("ApierV1.ExecuteAction", attrs, &reply2); err == nil || reply2 == "OK" {
		t.Error("Expecting error on ApierV1.ExecuteAction.", err, reply2)
	}
}

func TestApierSetActions(t *testing.T) {
	if !*testLocal {
		return
	}
	act1 := &utils.TPAction {Identifier: engine.TOPUP_RESET, BalanceType: engine.CREDIT, Direction: engine.OUTBOUND, Units: 75.0, ExpiryTime: engine.UNLIMITED, Weight: 20.0}
	attrs1 := &AttrSetActions{ActionsId: "ACTS_1", Actions : []*utils.TPAction{act1}}
	reply1 := ""
	if err := rater.Call("ApierV1.SetActions", attrs1, &reply1); err != nil {
		t.Error("Got error on ApierV1.SetActions: ", err.Error())
	} else if reply1 != "OK" {
		t.Errorf("Calling ApierV1.SetActions received: %s", reply1)
	}
	// Calling the second time should raise EXISTS
	if err := rater.Call("ApierV1.SetActions", attrs1, &reply1); err == nil || err.Error() != "EXISTS"{
		t.Error("Unexpected result on duplication: ", err.Error())
	}
}

func TestApierSetActionTimings(t *testing.T) {
	if !*testLocal {
		return
	}
	atm1 := &ApiActionTiming{ActionsId: "ACTS_1", MonthDays: "1", Time: "00:00:00", Weight: 20.0}
	atms1 := &AttrSetActionTimings{ ActionTimingsId: "ATMS_1", ActionTimings: []*ApiActionTiming{atm1} }
	reply1 := ""
	if err := rater.Call("ApierV1.SetActionTimings", atms1, &reply1); err != nil {
		t.Error("Got error on ApierV1.SetActionTimings: ", err.Error())
	} else if reply1 != "OK" {
		t.Errorf("Calling ApierV1.SetActionTimings received: %s", reply1)
	}
	// Calling the second time should raise EXISTS
	if err := rater.Call("ApierV1.SetActionTimings", atms1, &reply1); err == nil || err.Error() != "EXISTS"{
		t.Error("Unexpected result on duplication: ", err.Error())
	}
}

// Test here AddTriggeredAction
func TestApierAddTriggeredAction(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Add balance to a previously known account
	attrs := &AttrAddActionTrigger{Tenant: "cgrates.org", Account: "dan2", Direction: "*out", BalanceId: "*monetary",
		ThresholdType: "*min_balance", ThresholdValue: 2, DestinationId: "*any", Weight: 10, ActionsId: "WARN_VIA_HTTP"}
	if err := rater.Call("ApierV1.AddTriggeredAction", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddTriggeredAction: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddTriggeredAction received: %s", reply)
	}
	reply2 := ""
	attrs2 := new(AttrAddActionTrigger)
	*attrs2 = *attrs
	attrs2.Account = "dan3" // Does not exist so it should error when adding triggers on it
	// Add trigger to an account which does n exist
	if err := rater.Call("ApierV1.ExecuteAction", attrs2, &reply2); err == nil || reply2 == "OK" {
		t.Error("Expecting error on ApierV1.AddTriggeredAction.", err, reply2)
	}
}

// Test here AddAccount
func TestApierAddAccount(t *testing.T) {
	if !*testLocal {
		return
	}
	//reply := ""
	attrs := &AttrAddAccount{Tenant: "cgrates.org", Direction: "*out", Account: "dan4", Type: "prepaid", ActionTimingsId: "PREPAID_10"}
	//if err := rater.Call("ApierV1.AddAccount", attrs, &reply); err != nil {
	//	t.Error("Got error on ApierV1.AddAccount: ", err.Error())
	//} else if reply != "OK" {
	//	t.Errorf("Calling ApierV1.AddAccount received: %s", reply)
	//}
	reply2 := ""
	attrs2 := new(AttrAddAccount)
	*attrs2 = *attrs
	attrs2.ActionTimingsId = "DUMMY_DATA" // Does not exist so it should error when adding triggers on it
	// Add account with actions timing which does not exist
	if err := rater.Call("ApierV1.AddAccount", attrs2, &reply2); err == nil || reply2 == "OK" {
		t.Error("Expecting error on ApierV1.AddAccount.", err, reply2)
	}
}

// Test here GetBalance
func TestApierGetBalance(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply float64
	attrs := &AttrGetBalance{Tenant: "cgrates.org", Account: "1001", BalanceId: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetBalance: ", err.Error())
	} else if reply != 11.5 { // We expect 11.5 since we have added in the previous test 1.5
		t.Errorf("Calling ApierV1.GetBalance expected: 11.5, received: %f", reply)
	}
	attrs = &AttrGetBalance{Tenant: "cgrates.org", Account: "dan", BalanceId: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetBalance: ", err.Error())
	} else if reply != 1.5 {
		t.Errorf("Calling ApierV1.GetBalance expected: 1.5, received: %f", reply)
	}
	// The one we have topped up though executeAction
	attrs = &AttrGetBalance{Tenant: "cgrates.org", Account: "dan2", BalanceId: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetBalance: ", err.Error())
	} else if reply != 10 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10, received: %f", reply)
	}
}

// Test here LoadTariffPlanFromFolder
func TestApierLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Simple test that command is executed without errors
	attrs := &AttrLoadTPFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "prepaid1centpsec")}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
}

// Test here ResponderGetCost
func TestResponderGetCost(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart, _ := utils.ParseDate("2013-08-07T17:30:00Z")
	tEnd, _ := utils.ParseDate("2013-08-07T17:31:30Z")
	cd := engine.CallDescriptor{
		Direction:    "*out",
		TOR:          "call",
		Tenant:       "cgrates.org",
		Subject:      "1001",
		Account:      "1001",
		Destination:  "+4917621621391",
		CallDuration: 90,
		TimeStart:    tStart,
		TimeEnd:      tEnd,
	}
	var cc engine.CallCost
	// Simple test that command is executed without errors
	if err := rater.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 90.0 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}
}

// Simply kill the engine after we are done with tests within this file
func TestStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run()
}
