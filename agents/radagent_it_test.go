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
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

var (
	raonfigDIR             string
	raCfgPath              string
	raCfg                  *config.CGRConfig
	raAuthClnt, raAcctClnt *radigo.Client
	raRPC                  *rpc.Client

	sTestsRadius = []func(t *testing.T){
		testRAitInitCfg,
		testRAitResetDataDb,
		testRAitResetStorDb,
		testRAitStartEngine,
		testRAitApierRpcConn,
		testRAitTPFromFolder,
		testRAitAuth,
		testRAitAcctStart,
		testRAitAcctStop,
		testRAitStopCgrEngine,
	}
)

// Test start here
func TestRAit(t *testing.T) {
	engine.KillEngine(0)
	raonfigDIR = "radagent"
	for _, stest := range sTestsRadius {
		t.Run(diamConfigDIR, stest)
	}
}

func TestRAitDispatcher(t *testing.T) {
	isDispatcherActive = true
	engine.StartEngine(path.Join(*dataDir, "conf", "samples", "dispatchers", "all"), 200)
	engine.StartEngine(path.Join(*dataDir, "conf", "samples", "dispatchers", "all2"), 200)
	raonfigDIR = "dispatchers/radagent"
	testRadiusitResetAllDB(t)
	for _, stest := range sTestsRadius {
		t.Run(diamConfigDIR, stest)
	}
	engine.KillEngine(100)
	isDispatcherActive = false
}

func testRAitInitCfg(t *testing.T) {
	raCfgPath = path.Join(*dataDir, "conf", "samples", raonfigDIR)
	// Init config first
	var err error
	raCfg, err = config.NewCGRConfigFromPath(raCfgPath)
	if err != nil {
		t.Error(err)
	}
	raCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(raCfg)
	if isDispatcherActive {
		raCfg.ListenCfg().RPCJSONListen = ":6012"
	}
}

func testRadiusitResetAllDB(t *testing.T) {
	cfgPath1 := path.Join(*dataDir, "conf", "samples", "dispatchers", "all")
	allCfg, err := config.NewCGRConfigFromPath(cfgPath1)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(allCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(allCfg); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testRAitResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(raCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testRAitResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(raCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRAitStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(raCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testRAitApierRpcConn(t *testing.T) {
	var err error
	raRPC, err = jsonrpc.Dial("tcp", raCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testRAitTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := raRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	if isDispatcherActive {
		testRadiusitTPLoadData(t)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testRadiusitTPLoadData(t *testing.T) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", raCfgPath, "-path", path.Join(*dataDir, "tariffplans", "dispatchers"))

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

func testRAitAuth(t *testing.T) {
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
	if err := authReq.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
		t.Error(err)
	}
	if err := authReq.AddAVPWithName("Sip-From-Tag", "51585361", ""); err != nil {
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

func testRAitAcctStart(t *testing.T) {
	if raAcctClnt, err = radigo.NewClient("udp", "127.0.0.1:1813", "CGRateS.org", dictRad, 1, nil); err != nil {
		t.Fatal(err)
	}
	req := raAcctClnt.NewRequest(radigo.AccountingRequest, 2) // emulates Kamailio packet for accounting start
	if err := req.AddAVPWithName("Acct-Status-Type", "Start", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Service-Type", "Sip-Session", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-Response-Code", "200", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-Method", "Invite", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Event-Timestamp", "1497106115", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-From-Tag", "51585361", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-To-Tag", "75c2f57b", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("User-Name", "1001", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Ascend-User-Acct-Time", "1497106115", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("NAS-Port-Id", "5060", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Acct-Delay-Time", "0", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
		t.Error(err)
	}
	reply, err := raAcctClnt.SendRequest(req)
	if err != nil {
		t.Error(err)
	}
	if reply.Code != radigo.AccountingResponse {
		t.Errorf("Received reply: %+v", reply)
	}
	if len(reply.AVPs) != 0 { // we don't expect AVPs to be populated
		t.Errorf("Received AVPs: %+v", reply.AVPs)
	}
	// Make sure the sessin is managed by SMG
	time.Sleep(10 * time.Millisecond)
	expUsage := 10 * time.Second
	if isDispatcherActive { // no debit interval set on dispatched session
		expUsage = 3 * time.Hour
	}
	var aSessions []*sessions.ActiveSession
	if err := raRPC.Call(utils.SessionSv1GetActiveSessions,
		map[string]string{utils.RunID: utils.META_DEFAULT,
			utils.OriginID: "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-51585361-75c2f57b"},
		&aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != expUsage {
		t.Errorf("Expecting %v, received usage: %v\nAnd Session: %s ", expUsage, aSessions[0].Usage, utils.ToJSON(aSessions))
	}
}

func testRAitAcctStop(t *testing.T) {
	req := raAcctClnt.NewRequest(radigo.AccountingRequest, 3) // emulates Kamailio packet for accounting start
	if err := req.AddAVPWithName("Acct-Status-Type", "Stop", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Service-Type", "Sip-Session", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-Response-Code", "200", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-Method", "Bye", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Event-Timestamp", "1497106119", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-From-Tag", "51585361", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Sip-To-Tag", "75c2f57b", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("User-Name", "1001", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Ascend-User-Acct-Time", "1497106115", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("NAS-Port-Id", "5060", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("Acct-Delay-Time", "0", ""); err != nil {
		t.Error(err)
	}
	if err := req.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
		t.Error(err)
	}
	reply, err := raAcctClnt.SendRequest(req)
	if err != nil {
		t.Error(err)
	}
	if reply.Code != radigo.AccountingResponse {
		t.Errorf("Received reply: %+v", reply)
	}
	if len(reply.AVPs) != 0 { // we don't expect AVPs to be populated
		t.Errorf("Received AVPs: %+v", reply.AVPs)
	}
	// Make sure the sessin was disconnected from SMG
	var aSessions []*sessions.ActiveSession
	if err := raRPC.Call(utils.SessionSv1GetActiveSessions,
		map[string]string{utils.RunID: utils.META_DEFAULT, utils.OriginID: "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-51585361-75c2f57b"},
		&aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	time.Sleep(150 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, DestinationPrefixes: []string{"1002"}}
	if err := raRPC.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "4s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v ", cdrs[0])
		}
		if cdrs[0].CostSource != utils.MetaSessionS {
			t.Errorf("Unexpected CDR CostSource received for CDR: %v", cdrs[0])
		}
		if cdrs[0].Cost != 0.01 {
			t.Errorf("Unexpected CDR Cost received for CDR: %v", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["RemoteAddr"] != "127.0.0.1" {
			t.Errorf("Unexpected CDR RemoteAddr received for CDR: %+v", utils.ToJSON(cdrs[0]))
		}
	}
}

func testRAitStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
