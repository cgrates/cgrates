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

package v2

import (
	"flag"
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

var testTP = flag.Bool("tp", false, "Perform the tests for TariffPlans, not by default.") // Separate from integration so we can run on multiple DBs without involving all other tests on each run
var configDIR = flag.String("config_dir", "tutmysql", "Relative path towards a config directory under samples prefix")

var tpCfgPath string
var tpCfg *config.CGRConfig
var tpRPC *rpc.Client
var err error

var testTPid = "V2TestTPit"
var delay int

func TestTPitLoadConfig(t *testing.T) {
	if !*testTP {
		return
	}
	tpCfgPath = path.Join(*dataDir, "conf", "samples", *configDIR)
	if tpCfg, err = config.NewCGRConfigFromFolder(tpCfgPath); err != nil {
		t.Error(err)
	}
	switch *configDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		delay = 4000
	default:
		delay = *waitRater
	}
}

// Remove data in both rating and accounting db
func TestTPitResetDataDb(t *testing.T) {
	if !*testTP {
		return
	}
	if err := engine.InitDataDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTPitResetStorDb(t *testing.T) {
	if !*testTP {
		return
	}
	if err := engine.InitStorDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTPitStartEngine(t *testing.T) {
	if !*testTP {
		return
	}
	if _, err := engine.StopStartEngine(tpCfgPath, delay); err != nil { // Mongo requires more time to start
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTPitRpcConn(t *testing.T) {
	if !*testTP {
		return
	}
	var err error
	tpRPC, err = jsonrpc.Dial("tcp", tpCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestTPitTimings(t *testing.T) {
	if !*testTP {
		return
	}
	// PEAK,*any,*any,*any,1;2;3;4;5,08:00:00
	tmPeak := &utils.ApierTPTiming{
		TPid:      testTPid,
		TimingId:  "PEAK",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "08:00:00",
	}
	// OFFPEAK_MORNING,*any,*any,*any,1;2;3;4;5,00:00:00
	tmOffPeakMorning := &utils.ApierTPTiming{
		TPid:      testTPid,
		TimingId:  "OFFPEAK_MORNING",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "00:00:00",
	}
	// OFFPEAK_EVENING,*any,*any,*any,1;2;3;4;5,19:00:00
	tmOffPeakEvening := &utils.ApierTPTiming{
		TPid:      testTPid,
		TimingId:  "OFFPEAK_EVENING",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;3;4;5",
		Time:      "19:00:00",
	}
	// OFFPEAK_WEEKEND,*any,*any,*any,6;7,00:00:00
	tmOffPeakWeekend := &utils.ApierTPTiming{
		TPid:      testTPid,
		TimingId:  "OFFPEAK_WEEKEND",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "6;7",
		Time:      "00:00:00",
	}
	// DUMMY, only used for the purpose of testing remove function
	tmDummyRemove := &utils.ApierTPTiming{
		TPid:      testTPid,
		TimingId:  "DUMMY_REMOVE",
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
	if err := tpRPC.Call("ApierV2.GetTPTiming", v1.AttrGetTPTiming{tmDummyRemove.TPid, tmDummyRemove.TimingId}, &rplyTmDummy); err != nil {
		t.Error("Calling ApierV2.GetTPTiming, got error: ", err.Error())
	} else if !reflect.DeepEqual(tmDummyRemove, rplyTmDummy) {
		t.Errorf("Calling ApierV2.GetTPTiming expected: %v, received: %v", tmDummyRemove, rplyTmDummy)
	}
	// Test remove
	if err := tpRPC.Call("ApierV2.RemTPTiming", v1.AttrGetTPTiming{tmDummyRemove.TPid, tmDummyRemove.TimingId}, &reply); err != nil {
		t.Error("Calling ApierV2.RemTPTiming, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV2.RemTPTiming received: ", reply)
	}
	// Test getIds
	var rplyTmIDs []string
	expectedTmIDs := []string{"OFFPEAK_EVENING", "OFFPEAK_MORNING", "OFFPEAK_WEEKEND", "PEAK"}
	if err := tpRPC.Call("ApierV1.GetTPTimingIds", v1.AttrGetTPTimingIds{testTPid, utils.Paginator{}}, &rplyTmIDs); err != nil {
		t.Error("Calling ApierV1.GetTPTimingIds, got error: ", err.Error())
	} else if len(expectedTmIDs) != len(rplyTmIDs) {
		t.Errorf("Calling ApierV1.GetTPTimingIds expected: %v, received: %v", expectedTmIDs, rplyTmIDs)
	}
}

func TestTPitDestinations(t *testing.T) {
	if !*testTP {
		return
	}
	var reply string
	// DST_1002,1002
	dst1002 := &utils.TPDestination{TPid: testTPid, DestinationId: "DST_1002", Prefixes: []string{"1002"}}
	// DST_1003,1003
	dst1003 := &utils.TPDestination{TPid: testTPid, DestinationId: "DST_1003", Prefixes: []string{"1003"}}
	// DST_1007,1007
	dst1007 := &utils.TPDestination{TPid: testTPid, DestinationId: "DST_1007", Prefixes: []string{"1007"}}
	// DST_FS,10
	dstFS := &utils.TPDestination{TPid: testTPid, DestinationId: "DST_FS", Prefixes: []string{"10"}}
	// DST_DE_MOBILE,+49151
	// DST_DE_MOBILE,+49161
	// DST_DE_MOBILE,+49171
	dstDEMobile := &utils.TPDestination{TPid: testTPid, DestinationId: "DST_DE_MOBILE", Prefixes: []string{"+49151", "+49161", "+49171"}}
	dstDUMMY := &utils.TPDestination{TPid: testTPid, DestinationId: "DUMMY_REMOVE", Prefixes: []string{"999"}}
	for _, dst := range []*utils.TPDestination{dst1002, dst1003, dst1007, dstFS, dstDEMobile, dstDUMMY} {
		if err := tpRPC.Call("ApierV2.SetTPDestination", dst, &reply); err != nil {
			t.Error("Got error on ApierV2.SetTPDestination: ", err.Error())
		} else if reply != utils.OK {
			t.Error("Unexpected reply received when calling ApierV2.SetTPDestination: ", reply)
		}
	}
	// Test get
	/*
		 FixMe for mongodb
			var rplyDst *utils.TPDestination
			if err := tpRPC.Call("ApierV2.GetTPDestination", v1.AttrGetTPDestination{testTPid, dstDEMobile.DestinationId}, &rplyDst); err != nil {
				t.Error("Calling ApierV2.GetTPDestination, got error: ", err.Error())
			} else if len(dstDEMobile.Prefixes) != len(rplyDst.Prefixes) {
				t.Errorf("Calling ApierV2.GetTPDestination expected: %v, received: %v", dstDEMobile, rplyDst)
			}
	*/
	// Test remove
	if err := tpRPC.Call("ApierV2.RemTPDestination", v1.AttrGetTPDestination{testTPid, dstDUMMY.DestinationId}, &reply); err != nil {
		t.Error("Calling ApierV1.RemTPTiming, got error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV2.RemTPTiming received: ", reply)
	}
	// Test getIds
	var rplyDstIds []string
	expectedDstIds := []string{"DST_1002", "DST_1003", "DST_1007", "DST_DE_MOBILE", "DST_FS"}
	if err := tpRPC.Call("ApierV2.GetTPDestinationIds", v1.AttrGetTPDestinationIds{TPid: testTPid}, &rplyDstIds); err != nil {
		t.Error("Calling ApierV1.GetTPDestinationIds, got error: ", err.Error())
	} else if len(expectedDstIds) != len(rplyDstIds) {
		t.Errorf("Calling ApierV2.GetTPDestinationIds expected: %v, received: %v", expectedDstIds, rplyDstIds)
	}

}

func TestTPitKillEngine(t *testing.T) {
	if !*testTP {
		return
	}
	if err := engine.KillEngine(delay); err != nil {
		t.Error(err)
	}
}
