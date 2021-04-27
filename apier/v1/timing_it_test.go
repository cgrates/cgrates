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
	"net/rpc"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	timCfgPath   string
	timCfg       *config.CGRConfig
	timSRPC      *rpc.Client
	timConfigDIR string //run tests for specific configuration

	sTestsTiming = []func(t *testing.T){
		testTimingInitCfg,
		testTimingInitDataDb,
		testTimingResetStorDb,
		testTimingStartEngine,
		testTimingRPCConn,
		testTimingGetTimingNotFound,
		testTimingSetTiming,
		testTimingGetTimingAfterSet,
		testTimingRemoveTiming,
		testTimingGetTimingNotFound,
		testTimingKillEngine,
	}
)

//Tests start here
func TestTimingIT(t *testing.T) {
	timingTests := sTestsTiming
	switch *dbType {
	case utils.MetaInternal:
		timConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		timConfigDIR = "tutmysql"
	case utils.MetaMongo:
		timConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range timingTests {
		t.Run(timConfigDIR, stest)
	}
}

func testTimingInitCfg(t *testing.T) {
	var err error
	timCfgPath = path.Join(*dataDir, "conf", "samples", timConfigDIR)
	timCfg, err = config.NewCGRConfigFromPath(timCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testTimingInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(timCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testTimingResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(timCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTimingStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(timCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTimingRPCConn(t *testing.T) {
	var err error
	timSRPC, err = newRPCClient(timCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTimingKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testTimingGetTimingNotFound(t *testing.T) {
	var reply *utils.TPTiming
	if err := timSRPC.Call(utils.APIerSv1GetTiming, &utils.ArgsGetTimingID{ID: "MIDNIGHT"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTimingSetTiming(t *testing.T) {
	timing := &utils.TPTimingWithAPIOpts{
		TPTiming: &utils.TPTiming{
			ID:        "MIDNIGHT",
			Years:     utils.Years{2020, 2019},
			Months:    utils.Months{1, 2, 3, 4},
			MonthDays: utils.MonthDays{5, 6, 7, 8},
			WeekDays:  utils.WeekDays{0, 1, 2, 3},
			StartTime: "00:00:00",
			EndTime:   "00:00:01",
		},
		APIOpts: map[string]interface{}{},
	}

	var reply string

	if err := timSRPC.Call(utils.APIerSv1SetTiming, timing, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testTimingGetTimingAfterSet(t *testing.T) {
	var reply *utils.TPTiming
	if err := timSRPC.Call(utils.APIerSv1GetTiming, &utils.ArgsGetTimingID{ID: "MIDNIGHT"},
		&reply); err != nil {
		t.Fatal(err)
	}
	exp := &utils.TPTiming{
		ID:        "MIDNIGHT",
		Years:     utils.Years{2020, 2019},
		Months:    utils.Months{1, 2, 3, 4},
		MonthDays: utils.MonthDays{5, 6, 7, 8},
		WeekDays:  utils.WeekDays{0, 1, 2, 3},
		StartTime: "00:00:00",
		EndTime:   "00:00:01",
	}
	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, reply)
	}
}

func testTimingRemoveTiming(t *testing.T) {
	var reply string
	if err := timSRPC.Call(utils.APIerSv1RemoveTiming,
		&utils.ArgsGetTimingID{ID: "MIDNIGHT"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}
