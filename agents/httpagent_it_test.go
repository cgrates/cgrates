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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	haCfgPath string
	haCfg     *config.CGRConfig
	haRPC     *rpc.Client
	httpC     *http.Client // so we can cache the connection
)

func TestHAitInitCfg(t *testing.T) {
	haCfgPath = path.Join(*dataDir, "conf", "samples", "httpagent")
	// Init config first
	var err error
	haCfg, err = config.NewCGRConfigFromFolder(haCfgPath)
	if err != nil {
		t.Error(err)
	}
	haCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(haCfg)
	httpC = new(http.Client)
}

// Remove data in both rating and accounting db
func TestHAitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(haCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestHAitResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(haCfg); err != nil {
		t.Fatal(err)
	}
}

/*
// Start CGR Engine
func TestHAitStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(haCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}
*/

// Connect rpc client to rater
func TestHAitApierRpcConn(t *testing.T) {
	var err error
	haRPC, err = jsonrpc.Dial("tcp", haCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestHAitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := haRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestHAitAuth(t *testing.T) {
	reqUrl := fmt.Sprintf("http://%s%s?request_type=OutboundAUTH&CallID=123456&Msisdn=497700056231&Imsi=2343000000000123&Destination=491239440004&MSRN=0102220233444488999&ProfileID=1&AgentID=176&GlobalMSISDN=497700056129&GlobalIMSI=214180000175129&ICCID=8923418450000089629&MCC=234&MNC=10&calltype=callback",
		haCfg.HTTPListen, haCfg.HttpAgentCfg()[0].Url)
	rply, err := httpC.Get(reqUrl)
	if err != nil {
		t.Error(err)
	}
	if body, err := ioutil.ReadAll(rply.Body); err != nil {
		t.Error(err)
	} else {
		fmt.Printf("Got reply: %s\n", string(body))
	}
	rply.Body.Close()
}
