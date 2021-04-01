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

package registrarc

import (
	"net/rpc"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	rpcDir     string
	rpcCMD     *exec.Cmd
	rpcCfgPath string

	rpcsDir     string
	rpcsCfgPath string
	rpcsCfg     *config.CGRConfig
	rpcsRPC     *rpc.Client

	rpchTest = []func(t *testing.T){
		testRPCInitCfg,
		testRPCInitDB,
		testRPCStartEngine,
		testRPCLoadData,
		testRPCChargerSNoAttr,
		testRPCStartRegc,
		testRPCChargerSWithAttr,
		testRPCStopEngines,
		testRPCChargerSNoAttr,
		testRPCStopRegs,
	}
)

func TestRPCHosts(t *testing.T) {
	switch *dbType {
	case utils.MetaMySQL:
		rpcDir = "registrarc_rpc_mysql"
		rpcsDir = "registrars_rpc_mysql"
	case utils.MetaMongo:
		rpcDir = "registrarc_rpc_mongo"
		rpcsDir = "registrars_rpc_mongo"
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range rpchTest {
		t.Run(rpcDir, stest)
	}
}

func testRPCInitCfg(t *testing.T) {
	rpcCfgPath = path.Join(*dataDir, "conf", "samples", "registrarc", rpcDir)
	rpcsCfgPath = path.Join(*dataDir, "conf", "samples", "registrarc", rpcsDir)
	var err error
	if rpcsCfg, err = config.NewCGRConfigFromPath(rpcsCfgPath); err != nil {
		t.Error(err)
	}
}

func testRPCInitDB(t *testing.T) {
	if err := engine.InitDataDB(rpcsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(rpcsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRPCStartEngine(t *testing.T) {
	var err error
	if _, err = engine.StopStartEngine(rpcsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	rpcsRPC, err = newRPCClient(rpcsCfg.ListenCfg())
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testRPCLoadData(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := rpcsRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testRPCChargerSNoAttr(t *testing.T) {
	cgrEv := &utils.CGREvent{ // matching Charger1
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1010",
		},
		APIOpts: map[string]interface{}{utils.OptsAttributesProcessRuns: 1.},
	}
	expErr := utils.NewErrServerError(rpcclient.ErrDisconnected).Error()
	var rply []*engine.ChrgSProcessEventReply
	if err := rpcsRPC.Call(utils.ChargerSv1ProcessEvent, cgrEv, &rply); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s,received: %v", expErr, err)
	}
}

func testRPCStartRegc(t *testing.T) {
	var err error
	if rpcCMD, err = engine.StartEngine(rpcCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
}

func testRPCChargerSWithAttr(t *testing.T) {
	cgrEv := &utils.CGREvent{ // matching Charger1
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.AccountField: "1010",
		},
		APIOpts: map[string]interface{}{utils.OptsAttributesProcessRuns: 1.},
	}

	processedEv := []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile: "CustomerCharges",
			AlteredFields:   []string{"*req.RunID"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					"Account": "1010",
					"RunID":   "CustomerCharges",
				},
				APIOpts: map[string]interface{}{
					"*processRuns": 1.,
					"*subsys":      "*chargers",
				},
			},
		}, {
			ChargerSProfile:    "Raw",
			AttributeSProfiles: []string{"*constant:*req.RequestType:*none"},
			AlteredFields:      []string{"*req.RunID", "*req.RequestType"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					"Account":     "1010",
					"RequestType": "*none",
					"RunID":       "raw",
				},
				APIOpts: map[string]interface{}{
					"*processRuns": 1.,
					"*subsys":      "*chargers",
				},
			},
		}, {
			ChargerSProfile:    "SupplierCharges",
			AttributeSProfiles: []string{"ATTR_SUPPLIER1"},
			AlteredFields:      []string{"*req.RunID", "*req.Subject"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					"Account": "1010",
					"RunID":   "SupplierCharges",
					"Subject": "SUPPLIER1",
				},
				APIOpts: map[string]interface{}{
					"*processRuns": 1.,
					"*subsys":      "*chargers",
				},
			},
		},
	}
	var rply []*engine.ChrgSProcessEventReply
	if err := rpcsRPC.Call(utils.ChargerSv1ProcessEvent, cgrEv, &rply); err != nil {
		t.Fatal(err)
	}
	sort.Slice(rply, func(i, j int) bool {
		return rply[i].ChargerSProfile < rply[j].ChargerSProfile
	})
	if !reflect.DeepEqual(rply, processedEv) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(processedEv), utils.ToJSON(rply))
	}
}

func testRPCStopEngines(t *testing.T) {
	if err := rpcCMD.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
}

func testRPCStopRegs(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
