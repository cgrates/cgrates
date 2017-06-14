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
package agents

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

var raCfgPath string
var raCfg *config.CGRConfig
var raSMGrpc *rpc.Client
var raAuthClnt, raAcctClnt *radigo.Client

func TestRAitInitCfg(t *testing.T) {
	raCfgPath = path.Join(*dataDir, "conf", "samples", "radagent")
	// Init config first
	var err error
	raCfg, err = config.NewCGRConfigFromFolder(raCfgPath)
	if err != nil {
		t.Error(err)
	}
	raCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(raCfg)
}

// Remove data in both rating and accounting db
func TestRAitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(raCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestRAitResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(raCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestRAitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(raCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestRAitApierRpcConn(t *testing.T) {
	var err error
	raSMGrpc, err = jsonrpc.Dial("tcp", raCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestRAitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := raSMGrpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestRAitAuth(t *testing.T) {
	if raAuthClnt, err = radigo.NewClient("udp", "127.0.0.1:1812", "CGRateS.org", dictRad, 1, nil); err != nil {
		t.Fatal(err)
	}
	authReq := raAuthClnt.NewRequest(radigo.AccessRequest, 1) // emulates Kamailio packet out of radius_load_caller_avps()
	if err := authReq.AddAVPWithName("User-Name", "1001", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("Service-Type", "SIP-Caller-AVPs", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("Acct-Session-Id", "2ca13afce9b2d76e15de7e1ec6568fc8@0:0:0:0:0:0:0:0", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("Sip-From-Tag", "7f30055f", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("Event-Timestamp", "1497106115", ""); err != nil {
		t.Error(err)
	}

	reply, err := raAuthClnt.SendRequest(authReq)
	if err != nil {
		t.Error(err)
	}
	if reply.Code != radigo.AccessAccept {
		t.Errorf("Received reply: %+v", reply)
	}
	if len(reply.AVPs) != 1 { // make sure max duration is received
		t.Errorf("Received AVPs: %+v", reply.AVPs)
	} else if !reflect.DeepEqual([]byte("session_max_time#10800"), reply.AVPs[0].RawValue) {
		t.Errorf("Received: %s", string(reply.AVPs[0].RawValue))
	}
}

func TestRAitStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
