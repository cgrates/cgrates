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
		Tenant:             "",
		ID:                 "",
		FilterIDs:          nil,
		ActivationInterval: nil,
		Weight:             0,
		Schedule:           "",
		AccountIDs:         nil,
		Actions:            nil,
	}
	var reply *engine.ActionProfile
	if err := actSRPC.Call(utils.APIerSv1GetActionProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_3"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testActionSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
