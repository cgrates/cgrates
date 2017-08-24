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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
)

var loaderCfgPath string
var loaderCfg *config.CGRConfig
var loaderRPC *rpc.Client
var loaderDataDir = "/usr/share/cgrates"

func TestLoaderInitCfg(t *testing.T) {
	var err error
	loaderCfgPath = path.Join(loaderDataDir, "conf", "samples", "tutmysql")
	loaderCfg, err = config.NewCGRConfigFromFolder(loaderCfgPath)
	if err != nil {
		t.Error(err)
	}
	loaderCfg.DataFolderPath = loaderDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(loaderCfg)
}

func TestLoaderInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(loaderCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestLoaderStorDb(t *testing.T) {
	if err := engine.InitStorDb(loaderCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestLoaderStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(loaderCfgPath, 1000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestLoaderRpcConn(t *testing.T) {
	var err error
	loaderRPC, err = jsonrpc.Dial("tcp", loaderCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoaderImportTPFromFolderPath(t *testing.T) {
	var reply string
	if err := loaderRPC.Call("ApierV1.ImportTariffPlanFromFolder", utils.AttrImportTPFromFolder{TPid: "TEST_LOADER", FolderPath: path.Join(loaderDataDir, "tariffplans", "tutorial")}, &reply); err != nil {
		t.Error("Got error on ApierV1.ImportTarrifPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ImportTarrifPlanFromFolder got reply: ", reply)
	}
}

func TestLoaderLoadTariffPlanFromStorDbDryRun(t *testing.T) {
	var reply string
	if err := loaderRPC.Call("ApierV1.LoadTariffPlanFromStorDb", AttrLoadTpFromStorDb{TPid: "TEST_LOADER", DryRun: true}, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromStorDb: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.LoadTariffPlanFromStorDb got reply: ", reply)
	}
}

func TestLoaderKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
