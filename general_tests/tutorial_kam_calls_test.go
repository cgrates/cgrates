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

var tutKamCallsCfg *config.CGRConfig
var tutKamCallsRpc *rpc.Client
var tutKamCallsPjSuaListener *os.File

func TestTutKamCallsInitCfg(t *testing.T) {
	if !*testCalls {
		return
	}
	// Init config first
	var err error
	tutKamCallsCfg, err = config.NewCGRConfigFromFolder(path.Join(*dataDir, "tutorials", "kamevapi", "cgrates", "etc", "cgrates"))
	if err != nil {
		t.Error(err)
	}
	tutKamCallsCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutKamCallsCfg)
}

// Wipe out the cdr database
func TestTutKamCallsResetDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitCdrDb(tutKamCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func TestTutKamCallsResetDataDb(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.InitDataDb(tutKamCallsCfg); err != nil {
		t.Fatal(err)
	}
}

// start FS server
func TestTutKamCallsStartKam(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("kamailio", *waitRater)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "kamevapi", "kamailio", "etc", "init.d", "kamailio"), "start", *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutKamCallsStartEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	engine.KillProcName("cgr-engine", *waitRater)
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "kamevapi", "cgrates", "etc", "init.d", "cgrates"), "start", *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Restart FS so we make sure reconnects are working

func TestTutKamCallsRestartKam(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "kamevapi", "kamailio", "etc", "init.d", "kamailio"), "restart", 1000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutKamCallsRpcConn(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	tutKamCallsRpc, err = jsonrpc.Dial("tcp", tutKamCallsCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutKamCallsLoadTariffPlanFromFolder(t *testing.T) {
	if !*testCalls {
		return
	}
	reply := ""
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutKamCallsRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Start Pjsua as listener and register it to receive calls
func TestTutKamCallsStartPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}
	var err error
	acnts := []*engine.PjsuaAccount{
		&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"},
		&engine.PjsuaAccount{Id: "sip:1002@127.0.0.1", Username: "1002", Password: "CGRateS.org", Realm: "*", Registrar: "sip:127.0.0.1:5060"}}
	if tutKamCallsPjSuaListener, err = engine.StartPjsuaListener(acnts, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Call from 1001 (prepaid) to 1002
func TestTutKamCallsCall1001To1002(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.PjsuaCallUri(&engine.PjsuaAccount{Id: "sip:1001@127.0.0.1", Username: "1001", Password: "CGRateS.org", Realm: "*"}, "sip:1002@127.0.0.1",
		"sip:127.0.0.1:5060", time.Duration(67)*time.Second, 5071); err != nil {
		t.Fatal(err)
	}
}

// Make sure account was debited properly
func TestTutKamCallsAccount1001(t *testing.T) {
	if !*testCalls {
		return
	}
	time.Sleep(time.Duration(70) * time.Second) // Allow calls to finish before start querying the results
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001", Direction: "*out"}
	if err := tutKamCallsRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue() == 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue())
	}
}

// Make sure account was debited properly
func TestTutKamCallsCdrs1001(t *testing.T) {
	if !*testCalls {
		return
	}
	var reply []*engine.ExternalCdr
	req := utils.RpcCdrsFilter{Accounts: []string{"1001"}, RunIds: []string{utils.META_DEFAULT}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].CdrSource != "KAMAILIO_CGR_CALL_END" {
			t.Errorf("Unexpected CdrSource for CDR: %s", reply[0].CdrSource)
		}
		if reply[0].ReqType != utils.META_PREPAID {
			t.Errorf("Unexpected ReqType for CDR: %s", reply[0].ReqType)
		}
		if reply[0].Usage != "67" { // Usage as seconds
			t.Errorf("Unexpected Usage for CDR: %+v", reply[0].Usage)
		}
		if reply[0].Cost != 0.0159 {
			t.Errorf("Unexpected Cost for CDR: %f", reply[0].Cost)
		}
	}
	req = utils.RpcCdrsFilter{Accounts: []string{"1001"}, RunIds: []string{"derived_run1"}}
	if err := tutKamCallsRpc.Call("ApierV2.GetCdrs", req, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(reply))
	} else {
		if reply[0].ReqType != utils.META_RATED {
			t.Errorf("Unexpected ReqType for CDR: %+v", reply[0])
		}
		if reply[0].Subject != "1002" {
			t.Errorf("Unexpected Subject for CDR: %+v", reply[0])
		}
		if reply[0].Cost != 1.2059 {
			t.Errorf("Unexpected Cost for CDR: %+v", reply[0].Cost)
		}
	}
}

func TestTutKamCallsStopPjsuaListener(t *testing.T) {
	if !*testCalls {
		return
	}

	tutKamCallsPjSuaListener.Write([]byte("q\n")) // Close pjsua
	time.Sleep(time.Duration(1) * time.Second)    // Allow pjsua to finish it's tasks, eg un-REGISTER
}

func TestTutKamCallsStopCgrEngine(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func TestTutKamCallsStopKamailio(t *testing.T) {
	if !*testCalls {
		return
	}
	if err := engine.CallScript(path.Join(*dataDir, "tutorials", "kamevapi", "kamailio", "etc", "init.d", "kamailio"), "stop", *waitRater); err != nil {
		t.Fatal(err)
	}
}
