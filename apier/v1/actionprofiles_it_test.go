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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	actPrfCfgPath   string
	actPrfCfg       *config.CGRConfig
	actSRPC         *rpc.Client
	actPrfDataDir   = "/usr/share/cgrates"
	actPrf          *AttributeWithCache
	actPrfConfigDIR string //run tests for specific configuration

	sTestsActPrf = []func(t *testing.T){
		testActionSInitCfg,
		testActionSInitDataDb,
		testActionSResetStorDb,
		testActionSStartEngine,
		testActionSRPCConn,
		testActionSLoadFromFolder,
		testActionSGetActionProfile,
		testActionSPing,
		testActionSKillEngine,
	}
)

//Test start here
func TestActionSIT(t *testing.T) {
	actsTests := sTestsActPrf
	switch *dbType {
	case utils.MetaInternal:
		actPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		actPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		actPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range actsTests {
		t.Run(actPrfConfigDIR, stest)
	}
}

func testActionSInitCfg(t *testing.T) {
	var err error
	actPrfCfgPath = path.Join(actPrfDataDir, "conf", "samples", actPrfConfigDIR)
	actPrfCfg, err = config.NewCGRConfigFromPath(actPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(actPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testActionSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(actPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testActionSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(actPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testActionSRPCConn(t *testing.T) {
	var err error
	actSRPC, err = newRPCClient(actPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testActionSLoadFromFolder(t *testing.T) {
	var reply string
	acts := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutactions")}
	if err := actSRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, acts, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testActionSGetActionProfile(t *testing.T) {
	expected := &engine.ActionProfile{
		Tenant:     "cgrates.org",
		ID:         "ONE_TIME_ACT",
		FilterIDs:  []string{},
		Weight:     10,
		Schedule:   utils.ASAP,
		AccountIDs: utils.StringSet{"1001": {}, "1002": {}},
		Actions: []*engine.APAction{
			{
				ID:        "TOPUP",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestBalance.Value",
				Value:     config.NewRSRParsersMustCompile("10", actPrfCfg.GeneralCfg().RSRSep),
			},
			{
				ID:        "SET_BALANCE_TEST_DATA",
				FilterIDs: []string{},
				Type:      "*set_balance",
				Path:      "~*balance.TestDataBalance.Type",
				Value:     config.NewRSRParsersMustCompile("*data", actPrfCfg.GeneralCfg().RSRSep),
			},
			{
				ID:        "TOPUP_TEST_DATA",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestDataBalance.Value",
				Value:     config.NewRSRParsersMustCompile("1024", actPrfCfg.GeneralCfg().RSRSep),
			},
			{
				ID:        "SET_BALANCE_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*set_balance",
				Path:      "~*balance.TestVoiceBalance.Type",
				Value:     config.NewRSRParsersMustCompile("*voice", actPrfCfg.GeneralCfg().RSRSep),
			},
			{
				ID:        "TOPUP_TEST_VOICE",
				FilterIDs: []string{},
				Type:      "*topup",
				Path:      "~*balance.TestVoiceBalance.Value",
				Value:     config.NewRSRParsersMustCompile("15m15s", actPrfCfg.GeneralCfg().RSRSep),
			},
		},
	}
	if *encoding == utils.MetaGOB {
		expected.FilterIDs = nil
		for i := range expected.Actions {
			expected.Actions[i].FilterIDs = nil
		}
	}
	var reply *engine.ActionProfile
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}}, &reply); err != nil {
		t.Fatal(err)
	} else {
		for _, act := range reply.Actions { // the path variable from RSRParsers is with lower letter and need to be compiled manually in tests to pass reflect.DeepEqual
			act.Value.Compile()
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}
}

func testActionSPing(t *testing.T) {
	var resp string
	if err := actSRPC.Call(utils.ActionSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testActionSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
