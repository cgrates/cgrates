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

package dispatchers

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dispEngine *testDispatcher
	allEngine  *testDispatcher
	allEngine2 *testDispatcher
)

type testDispatcher struct {
	CfgParh string
	Cfg     *config.CGRConfig
	RCP     *rpc.Client
	cmd     *exec.Cmd
}

func newTestEngine(t *testing.T, cfgPath string, initDataDB, intitStoreDB bool) (d *testDispatcher) {
	d = new(testDispatcher)
	d.CfgParh = cfgPath
	var err error
	d.Cfg, err = config.NewCGRConfigFromPath(d.CfgParh)
	if err != nil {
		t.Fatalf("Error at config init :%v\n", err)
	}
	d.Cfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()

	if initDataDB {
		d.initDataDb(t)
	}

	if intitStoreDB {
		d.resetStorDb(t)
	}
	d.startEngine(t)
	return d
}

func (d *testDispatcher) startEngine(t *testing.T) {
	var err error
	if d.cmd, err = engine.StartEngine(d.CfgParh, dspDelay); err != nil {
		t.Fatalf("Error at engine start:%v\n", err)
	}

	if d.RCP, err = jsonrpc.Dial("tcp", d.Cfg.ListenCfg().RPCJSONListen); err != nil {
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
	if err := engine.InitDataDb(d.Cfg); err != nil {
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
	if err := d.RCP.Call(utils.ApierV1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Errorf("<%s>Error at loading data from folder :%v", d.CfgParh, err)
	}
}

func (d *testDispatcher) loadData2(t *testing.T, path string) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", d.CfgParh, "-path", path)

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
	allEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", all), true, true)
	allEngine2 = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", all2), true, true)
	dispEngine = newTestEngine(t, path.Join(dspDataDir, "conf", "samples", "dispatchers", disp), true, true)
	dispEngine.loadData2(t, path.Join(dspDataDir, "tariffplans", attrTF))
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", allTF))
	allEngine2.loadData(t, path.Join(dspDataDir, "tariffplans", all2TF))
	time.Sleep(500 * time.Millisecond)
	for _, stest := range tests {
		t.Run(testName, stest)
	}
	dispEngine.stopEngine(t)
	allEngine.stopEngine(t)
	allEngine2.stopEngine(t)
	engine.KillEngine(0)
}
