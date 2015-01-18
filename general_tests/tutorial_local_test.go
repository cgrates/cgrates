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

package general_tests

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tutCfgPath string
var tutCfg *config.CGRConfig
var tutRpc *rpc.Client

func TestInitCfg(t *testing.T) {
	if !*testLocal {
		return
	}
	// Init config first
	tutCfgPath = path.Join(*dataDir, "conf", "samples", "tutorial")
	var err error
	tutCfg, err = config.NewCGRConfigFromFolder(tutCfgPath)
	if err != nil {
		t.Error(err)
	}
	tutCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutCfg)
}

func TestTutLclResetDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitCdrDb(tutCfg); err != nil {
		t.Fatal(err)
	}
}

func TestTutLclResetDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := engine.InitDataDb(tutCfg); err != nil {
		t.Fatal(err)
	}
}

func TestTutLclStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("Cannot find cgr-engine executable")
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	engine := exec.Command(enginePath, "-config_dir", tutCfgPath)
	if err := engine.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time to rater to fire up
}

// Connect rpc client to rater
func TestTutLclRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	tutRpc, err = jsonrpc.Dial("tcp", tutCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestTutLclLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups

}

func TestShutdown(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}
