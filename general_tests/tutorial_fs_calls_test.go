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
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var tutFsCallsCfgPath string
var tutFsCallsCfg *config.CGRConfig
var tutFsCallsRpc *rpc.Client
var tutFsCallsPjSuaListener *os.File

func TestTutFsCallsInitCfg(t *testing.T) {
	if !*testCalls {
		return
	}
	// Init config first
	tutFsCallsCfgPath = path.Join(*dataDir, "tutorials", "fs_evsock", "cgrates", "etc", "cgrates")
	var err error
	tutFsCallsCfg, err = config.NewCGRConfigFromFolder(tutFsCallsCfgPath)
	if err != nil {
		t.Error(err)
	}
	tutFsCallsCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutFsCallsCfg)
}

func TestTutFsCallsResetDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitCdrDb(tutFsCallsCfg); err != nil {
		t.Fatal(err)
	}
}

func TestTutFsCallsResetDataDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitDataDb(tutFsCallsCfg); err != nil {
		t.Fatal(err)
	}
}

func TestTutFsCallsStartFS(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.StartFreeSWITCH(path.Join(*dataDir, "tutorials", "fs_evsock", "freeswitch", "etc", "init.d", "freeswitch"), 3000); err != nil {
		t.Fatal(err)
	}
}

func TestTutFsCallsStartEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	if _, err := engine.StartEngine(tutFsCallsCfgPath, 3000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutFsCallsRpcConn(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	tutFsCallsRpc, err = jsonrpc.Dial("tcp", tutFsCallsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestTutFsCallsLoadTariffPlanFromFolder(t *testing.T) {
	if !*testCalls {
		return
	}
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutFsCallsRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestTutFsCallsStartPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	acnts := []*engine.PjsuaAccount{
		&engine.PjsuaAccount{Id: "sip:1004@10.10.10.102", Username: "1001", Password: "1234", Realm: "*", Registrar: "sip:10.10.10.102:5060"},
		&engine.PjsuaAccount{Id: "sip:1007@10.10.10.102", Username: "1002", Password: "1234", Realm: "*", Registrar: "sip:10.10.10.102:5060"}}
	if tutFsCallsPjSuaListener, err = engine.StartPjsuaListener(acnts, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func TestTutFsCallsCall1001To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1001@10.10.10.102", Username: "1001", Password: "1234", Realm: "*"}, "sip:1002@10.10.10.102", time.Duration(35)*time.Second, 5071); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Duration(40) * time.Second)
}

func TestTutFsCallsStopPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}
	tutFsCallsPjSuaListener.Write([]byte("q\n")) // Close pjsua
	time.Sleep(time.Duration(1) * time.Second)   // Allow pjsua to finish it's tasks, eg un-REGISTER
}

func TestTutFsCallsStopCgrEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func TestTutFsCallsStopFS(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.KillFreeSWITCH(1000); err != nil {
		t.Fatal(err)
	}
}
