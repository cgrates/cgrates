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
	"bytes"
	"net/rpc"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dspDir     string
	dspCfgPath string
	dspCfg     *config.CGRConfig
	dspCmd     *exec.Cmd
	dspRPC     *rpc.Client

	allDir     string
	allCfgPath string
	allCmd     *exec.Cmd

	all2Dir     string
	all2CfgPath string
	all2Cmd     *exec.Cmd

	dsphTest = []func(t *testing.T){
		testDsphInitCfg,
		testDsphInitDB,
		testDsphStartEngine,
		testDsphLoadData,
		testDsphBeforeDsphStart,
		testDsphStartAll2,
		testDsphStartAll,
		testDsphStopEngines,
		testDsphStopDispatcher,
	}
)

func TestDspHosts(t *testing.T) {
	switch *dbType {
	case utils.MetaMySQL:
		allDir = "all_mysql"
		all2Dir = "all2_mysql"
		dspDir = "dispatchers_mysql"
	case utils.MetaMongo:
		allDir = "all_mongo"
		all2Dir = "all2_mongo"
		dspDir = "dispatchers_mongo"
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range dsphTest {
		t.Run(dspDir, stest)
	}
}

func testDsphInitCfg(t *testing.T) {
	dspCfgPath = path.Join(*dataDir, "conf", "samples", "registrarc", dspDir)
	allCfgPath = path.Join(*dataDir, "conf", "samples", "registrarc", allDir)
	all2CfgPath = path.Join(*dataDir, "conf", "samples", "registrarc", all2Dir)
	var err error
	if dspCfg, err = config.NewCGRConfigFromPath(dspCfgPath); err != nil {
		t.Error(err)
	}
}

func testDsphInitDB(t *testing.T) {
	if err := engine.InitDataDb(dspCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(dspCfg); err != nil {
		t.Fatal(err)
	}
}

func testDsphStartEngine(t *testing.T) {
	var err error
	if dspCmd, err = engine.StopStartEngine(dspCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	dspRPC, err = newRPCClient(dspCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testDsphLoadData(t *testing.T) {
	loader := exec.Command("cgr-loader", "-config_path", dspCfgPath, "-path", path.Join(*dataDir, "tariffplans", "registrarc"), "-caches_address=")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	loader.Stdout = output
	loader.Stderr = outerr
	if err := loader.Run(); err != nil {
		t.Log(loader.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

func testDsphGetNodeID() (id string, err error) {
	var status map[string]interface{}
	if err = dspRPC.Call(utils.CoreSv1Status, utils.TenantWithOpts{
		Tenant: "cgrates.org",
		Opts:   map[string]interface{}{},
	}, &status); err != nil {
		return
	}
	return utils.IfaceAsString(status[utils.NodeID]), nil
}

func testDsphBeforeDsphStart(t *testing.T) {
	if _, err := testDsphGetNodeID(); err == nil || err.Error() != utils.ErrHostNotFound.Error() {
		t.Errorf("Expected error: %s received: %v", utils.ErrHostNotFound, err)
	}
}

func testDsphStartAll2(t *testing.T) {
	var err error
	if all2Cmd, err = engine.StartEngine(all2CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if nodeID, err := testDsphGetNodeID(); err != nil {
		t.Fatal(err)
	} else if nodeID != "ALL2" {
		t.Errorf("Expected nodeID: %q ,received: %q", "ALL2", nodeID)
	}
}

func testDsphStartAll(t *testing.T) {
	var err error
	if allCmd, err = engine.StartEngine(allCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if nodeID, err := testDsphGetNodeID(); err != nil {
		t.Fatal(err)
	} else if nodeID != "ALL" {
		t.Errorf("Expected nodeID: %q ,received: %q", "ALL", nodeID)
	}
}

func testDsphStopEngines(t *testing.T) {
	if err := allCmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err := all2Cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	if _, err := testDsphGetNodeID(); err == nil || err.Error() != utils.ErrHostNotFound.Error() {
		t.Errorf("Expected error: %s received: %v", utils.ErrHostNotFound, err)
	}
}

func testDsphStopDispatcher(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
