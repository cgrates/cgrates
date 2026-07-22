//go:build integration || flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"errors"
	"os/exec"
	"path"
	"strconv"
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
	dispEngine *testDispatcher
	allEngine  *testDispatcher
	allEngine2 *testDispatcher
)

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

type testDispatcher struct {
	CfgPath string
	Cfg     *config.CGRConfig
	RPC     *birpc.Client
	cmd     *exec.Cmd
}

func newTestEngine(t *testing.T, cfgPath string, initDataDB, initStoreDB bool) (d *testDispatcher) {
	d = new(testDispatcher)
	d.CfgPath = cfgPath
	var err error
	d.Cfg, err = config.NewCGRConfigFromPath(d.CfgPath)
	if err != nil {
		t.Fatalf("Error at config init :%v\n", err)
	}
	d.Cfg.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()

	if initDataDB {
		d.initDataDb(t)
	}

	if initStoreDB {
		d.resetStorDb(t)
	}
	d.startEngine(t)
	return d
}

func (d *testDispatcher) startEngine(t *testing.T) {
	var err error
	if d.cmd, err = engine.StartEngine(d.CfgPath, *utils.WaitRater); err != nil {
		t.Fatalf("Error at engine start:%v\n", err)
	}

	if d.RPC, err = newRPCClient(d.Cfg.ListenCfg()); err != nil {
		t.Fatalf("Error at dialing rcp client:%v\n", err)
	}
}

func (d *testDispatcher) stopEngine(t *testing.T) {
	pid := strconv.Itoa(d.cmd.Process.Pid)
	if err := exec.Command("kill", "-9", pid).Run(); err != nil {
		t.Fatalf("Error at stop engine:%v\n", err)
	}
	// // if err := d.cmd.Process.Kill(); err != nil {
	// // 	t.Fatalf("Error at stop engine:%v\n", err)
	// }
}

func (d *testDispatcher) initDataDb(t *testing.T) {
	if err := engine.InitDataDB(d.Cfg); err != nil {
		t.Fatalf("Error at DataDB init:%v\n", err)
	}
}

// Wipe out the cdr database
func (d *testDispatcher) resetStorDb(t *testing.T) {
	if err := engine.InitStorDb(d.Cfg); err != nil {
		t.Fatalf("Error at DataDB init:%v\n", err)
	}
}
func (d *testDispatcher) loadData(t *testing.T, path string) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path}
	if err := d.RPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Errorf("<%s>Error at loading data from folder :%v", d.CfgPath, err)
	}
}

func (d *testDispatcher) loadData2(t *testing.T, path string) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", d.CfgPath, "-path", path)

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(5 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}
}

func testDsp(t *testing.T, tests []func(t *testing.T), testName, all, all2, disp, allTF, all2TF, attrTF string) {
	engine.KillEngine(0)
	allEngine = newTestEngine(t, path.Join(*utils.DataDir, "conf", "samples", "dispatchers", all), true, true)
	allEngine2 = newTestEngine(t, path.Join(*utils.DataDir, "conf", "samples", "dispatchers", all2), true, true)
	dispEngine = newTestEngine(t, path.Join(*utils.DataDir, "conf", "samples", "dispatchers", disp), true, true)
	dispEngine.loadData2(t, path.Join(*utils.DataDir, "tariffplans", attrTF))
	allEngine.loadData(t, path.Join(*utils.DataDir, "tariffplans", allTF))
	allEngine2.loadData(t, path.Join(*utils.DataDir, "tariffplans", all2TF))
	time.Sleep(200 * time.Millisecond)
	for _, stest := range tests {
		t.Run(testName, stest)
	}
	dispEngine.stopEngine(t)
	allEngine.stopEngine(t)
	allEngine2.stopEngine(t)
	engine.KillEngine(0)
}
