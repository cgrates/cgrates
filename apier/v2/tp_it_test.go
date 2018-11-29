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
package v2

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tpCfgPath string
var tpCfg *config.CGRConfig
var tpRPC *rpc.Client
var err error
var delay int
var configDIR string // relative path towards a config directory under samples prefix

var (
	testTPid = "V2TestTPit"
)

// subtests to be executed for each confDIR
// FixMe#alin104n: add tests for rest of TP methods in loader interface
var sTestsTutIT = []func(t *testing.T){
	testTPitLoadConfig,
	testTPitResetDataDb,
	testTPitResetStorDb,
	testTPitStartEngine,
	testTPitRpcConn,
	testTPitTimings,
	testTPitDestinations,
	testTPitKillEngine,
}

// Tests starting here

func TestITMySQLTutorial(t *testing.T) {
	configDIR = "tutmysql"
	for _, stest := range sTestsTutIT {
		t.Run(configDIR, stest)
	}
}

func TestITpgTutorial(t *testing.T) {
	configDIR = "tutpostgres"
	for _, stest := range sTestsTutIT {
		t.Run(configDIR, stest)
	}
}

func TestITMongoTutorial(t *testing.T) {
	configDIR = "tutmongo"
	for _, stest := range sTestsTutIT {
		t.Run(configDIR, stest)
	}
}

func testTPitLoadConfig(t *testing.T) {
	tpCfgPath = path.Join(*dataDir, "conf", "samples", configDIR)
	if tpCfg, err = config.NewCGRConfigFromFolder(tpCfgPath); err != nil {
		t.Error(err)
	}
	switch configDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		delay = 4000
	default:
		delay = *waitRater
	}
}

// Remove data in both rating and accounting db
func testTPitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testTPitResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpCfgPath, delay); err != nil { // Mongo requires more time to start
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPitRpcConn(t *testing.T) {
	var err error
	tpRPC, err = jsonrpc.Dial("tcp", tpCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPitTimings(t *testing.T) {
	// PEAK,*any,*any,*any,1;2;3;4;5,08:00:00
	tmPeak := &utils.ApierTPTiming{
		TPid:      testTPid,
		ID:        "PEAK",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "08:00:00",
	}
	// OFFPEAK_MORNING,*any,*any,*any,1;2;3;4;5,00:00:00
	tmOffPeakMorning := &utils.ApierTPTiming{
		TPid:      testTPid,
		ID:        "OFFPEAK_MORNING",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "00:00:00",
	}
	// OFFPEAK_EVENING,*any,*any,*any,1;2;3;4;5,19:00:00
	tmOffPeakEvening := &utils.ApierTPTiming{
		TPid:      testTPid,
		ID:        "OFFPEAK_EVENING",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "19:00:00",
	}
	// OFFPEAK_WEEKEND,*any,*any,*any,6;7,00:00:00
	tmOffPeakWeekend := &utils.ApierTPTiming{
		TPid:      testTPid,
		ID:        "OFFPEAK_WEEKEND",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "6;7",
		Time:      "00:00:00",
	}
	// DUMMY, only used for the purpose of testing remove function
	tmDummyRemove := &utils.ApierTPTiming{
		TPid:      testTPid,
		ID:        "DUMMY_REMOVE",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "*any",
		Time:      "01:00:00",
	}
	// Test set
	var reply string
	for _, tm := range []*utils.ApierTPTiming{tmPeak, tmOffPeakMorning, tmOffPeakEvening, tmOffPeakWeekend, tmDummyRemove} {
		if err := tpRPC.Call("ApierV2.SetTPTiming", tm, &reply); err != nil {
			t.Error("Got error on ApierV2.SetTPTiming: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received when calling ApierV2.SetTPTiming: ", reply)
		}
	}
	// Test get
	var rplyTmDummy *utils.ApierTPTiming
	if err := tpRPC.Call("ApierV2.GetTPTiming", v1.AttrGetTPTiming{tmDummyRemove.TPid, tmDummyRemove.ID}, &rplyTmDummy); err != nil {
		t.Error("Calling ApierV2.GetTPTiming, got error: ", err.Error())
	} else if !reflect.DeepEqual(tmDummyRemove, rplyTmDummy) {
		t.Errorf("Calling ApierV2.GetTPTiming expected: %v, received: %v", tmDummyRemove, rplyTmDummy)
	}
	var rplyTmIDs []string
	expectedTmIDs := []string{"OFFPEAK_EVENING", "OFFPEAK_MORNING", "OFFPEAK_WEEKEND", "PEAK", tmDummyRemove.ID}
	if err := tpRPC.Call("ApierV1.GetTPTimingIds", v1.AttrGetTPTimingIds{testTPid, utils.Paginator{}}, &rplyTmIDs); err != nil {
		t.Error("Calling ApierV1.GetTPTimingIds, got error: ", err.Error())
	} else if len(expectedTmIDs) != len(rplyTmIDs) {
		t.Errorf("Calling ApierV1.GetTPTimingIds expected: %v, received: %v", expectedTmIDs, rplyTmIDs)
	}
	// Test remove
	if err := tpRPC.Call("ApierV2.RemTPTiming", v1.AttrGetTPTiming{tmDummyRemove.TPid, tmDummyRemove.ID}, &reply); err != nil {
		t.Error("Calling ApierV2.RemTPTiming, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV2.RemTPTiming received: ", reply)
	}
	// Test getIds
	rplyTmIDs = []string{}
	expectedTmIDs = []string{"OFFPEAK_EVENING", "OFFPEAK_MORNING", "OFFPEAK_WEEKEND", "PEAK"}
	if err := tpRPC.Call("ApierV1.GetTPTimingIds", v1.AttrGetTPTimingIds{testTPid, utils.Paginator{}}, &rplyTmIDs); err != nil {
		t.Error("Calling ApierV1.GetTPTimingIds, got error: ", err.Error())
	} else if len(expectedTmIDs) != len(rplyTmIDs) {
		t.Errorf("Calling ApierV1.GetTPTimingIds expected: %v, received: %v", expectedTmIDs, rplyTmIDs)
	}
}

func testTPitDestinations(t *testing.T) {
	var reply string
	// DST_1002,1002
	dst1002 := &utils.TPDestination{TPid: testTPid, ID: "DST_1002", Prefixes: []string{"1002"}}
	// DST_1003,1003
	dst1003 := &utils.TPDestination{TPid: testTPid, ID: "DST_1003", Prefixes: []string{"1003"}}
	// DST_1007,1007
	dst1007 := &utils.TPDestination{TPid: testTPid, ID: "DST_1007", Prefixes: []string{"1007"}}
	// DST_FS,10
	dstFS := &utils.TPDestination{TPid: testTPid, ID: "DST_FS", Prefixes: []string{"10"}}
	// DST_DE_MOBILE,+49151
	// DST_DE_MOBILE,+49161
	// DST_DE_MOBILE,+49171
	dstDEMobile := &utils.TPDestination{TPid: testTPid, ID: "DST_DE_MOBILE", Prefixes: []string{"+49151", "+49161", "+49171"}}
	dstDUMMY := &utils.TPDestination{TPid: testTPid, ID: "DUMMY_REMOVE", Prefixes: []string{"999"}}
	for _, dst := range []*utils.TPDestination{dst1002, dst1003, dst1007, dstFS, dstDEMobile, dstDUMMY} {
		if err := tpRPC.Call("ApierV2.SetTPDestination", dst, &reply); err != nil {
			t.Error("Got error on ApierV2.SetTPDestination: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received when calling ApierV2.SetTPDestination: ", reply)
		}
	}
	// Test get
	var rplyDst *utils.TPDestination
	if err := tpRPC.Call("ApierV2.GetTPDestination", AttrGetTPDestination{testTPid, dstDEMobile.ID}, &rplyDst); err != nil {
		t.Error("Calling ApierV2.GetTPDestination, got error: ", err.Error())
	} else if len(dstDEMobile.Prefixes) != len(rplyDst.Prefixes) {
		t.Errorf("Calling ApierV2.GetTPDestination expected: %v, received: %v", dstDEMobile, rplyDst)
	}
	// Test remove
	if err := tpRPC.Call("ApierV2.RemTPDestination", AttrGetTPDestination{testTPid, dstDUMMY.ID}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Received: ", reply)
	}
	// Test getIds
	var rplyDstIds []string
	expectedDstIds := []string{"DST_1002", "DST_1003", "DST_1007", "DST_DE_MOBILE", "DST_FS"}
	if err := tpRPC.Call("ApierV2.GetTPDestinationIDs", v1.AttrGetTPDestinationIds{TPid: testTPid}, &rplyDstIds); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationIDs, got error: ", err.Error())
	} else if len(expectedDstIds) != len(rplyDstIds) {
		t.Errorf("Calling ApierV2.GetTPDestinationIDs expected: %v, received: %v", expectedDstIds, rplyDstIds)
	}
}

func testTPitKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
