//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package registrarc

import (
	"bytes"
	"errors"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dspNodeDir     string
	dspNodeCfgPath string
	dspNodeCfg     *config.CGRConfig
	dspNodeCmd     *exec.Cmd
	dspNodeRPC     *birpc.Client

	node1Dir     string
	node1CfgPath string
	node1Cmd     *exec.Cmd

	dsphNodeTest = []func(t *testing.T){
		testDsphNodeInitCfg,
		testDsphNodeInitDB,
		testDsphNodeStartEngine,
		testDsphNodeLoadData,
		testDsphNodeBeforeDsphStart,
		testDsphNodeStartAll,
		testDsphNodeStopEngines,
		testDsphNodeStopDispatcher,
	}
)

func TestDspNodeHosts(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
		node1Dir = "registrarc_node_id"
		dspNodeDir = "registrars_node_id"
	case utils.MetaInternal, utils.MetaPostgres, utils.MetaMongo:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range dsphNodeTest {
		t.Run(dspNodeDir, stest)
	}
}

func testDsphNodeInitCfg(t *testing.T) {
	dspNodeCfgPath = path.Join(*utils.DataDir, "conf", "samples", "registrarc", dspNodeDir)
	node1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "registrarc", node1Dir)
	var err error
	if dspNodeCfg, err = config.NewCGRConfigFromPath(dspNodeCfgPath); err != nil {
		t.Error(err)
	}
}

func testDsphNodeInitDB(t *testing.T) {
	if err := engine.InitDataDB(dspNodeCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(dspNodeCfg); err != nil {
		t.Fatal(err)
	}
}

func testDsphNodeStartEngine(t *testing.T) {
	var err error
	if dspNodeCmd, err = engine.StopStartEngine(dspNodeCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	dspNodeRPC, err = newRPCClient(dspNodeCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testDsphNodeLoadData(t *testing.T) {
	loader := exec.Command("cgr-loader", "-config_path", dspNodeCfgPath, "-path", path.Join(*utils.DataDir, "tariffplans", "registrarc2"), "-caches_address=")
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

func testDsphNodeGetNodeID() (id string, err error) {
	var status map[string]any
	if err = dspNodeRPC.Call(context.Background(), utils.DispatcherSv1RemoteStatus, utils.TenantWithAPIOpts{
		Tenant:  "cgrates.org",
		APIOpts: map[string]any{},
	}, &status); err != nil {
		return
	}
	return utils.IfaceAsString(status[utils.NodeID]), nil
}

func testDsphNodeBeforeDsphStart(t *testing.T) {
	if _, err := testDsphNodeGetNodeID(); err == nil || err.Error() != utils.ErrDSPHostNotFound.Error() {
		t.Errorf("Expected error: %s received: %v", utils.ErrDSPHostNotFound, err)
	}
}

func testDsphNodeStartAll(t *testing.T) {
	var err error
	if node1Cmd, err = engine.StartEngine(node1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if nodeID, err := testDsphNodeGetNodeID(); err != nil {
		t.Fatal(err)
	} else if nodeID != "NODE1" {
		t.Errorf("Expected nodeID: %q ,received: %q", "NODE1", nodeID)
	}
}

func testDsphNodeStopEngines(t *testing.T) {
	if err := node1Cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	if _, err := testDsphNodeGetNodeID(); err == nil || err.Error() != utils.ErrDSPHostNotFound.Error() {
		t.Errorf("Expected error: %s received: %v", utils.ErrDSPHostNotFound, err)
	}
}

func testDsphNodeStopDispatcher(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *utils.Encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}
