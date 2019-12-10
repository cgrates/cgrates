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
package v1

import (
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var cdrsCfgPath string
var cdrsCfg *config.CGRConfig
var cdrsRpc *rpc.Client
var cdrsConfDIR string // run the tests for specific configuration

// subtests to be executed
var sTestsCDRsIT = []func(t *testing.T){
	testV1CDRsInitConfig,
	testV1CDRsInitDataDb,
	testV1CDRsInitCdrDb,
	testV1CDRsStartEngine,
	testV1CDRsRpcConn,
	testV1CDRsLoadTariffPlanFromFolder,
	testV1CDRsProcessEventWithRefund,
	testV1CDRsKillEngine,
}

// Tests starting here
func TestCDRsITInternal(t *testing.T) {
	cdrsConfDIR = "cdrsv1internal"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITMongo(t *testing.T) {
	cdrsConfDIR = "cdrsv1mongo"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITMySql(t *testing.T) {
	cdrsConfDIR = "cdrsv1mysql"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITPostgres(t *testing.T) {
	cdrsConfDIR = "cdrsv1postgres"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func testV1CDRsInitConfig(t *testing.T) {
	var err error
	cdrsCfgPath = path.Join(*dataDir, "conf", "samples", cdrsConfDIR)
	if cdrsCfg, err = config.NewCGRConfigFromPath(cdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV1CDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testV1CDRsInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testV1CDRsRpcConn(t *testing.T) {
	var err error
	cdrsRpc, err = newRPCClient(cdrsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst string
	if err := cdrsRpc.Call(utils.ApierV1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testV1CDRsProcessEventWithRefund(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testV1CDRsProcessEventWithRefund"}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.VOICE,
		Balance: map[string]interface{}{
			utils.ID:     "BALANCE1",
			utils.Value:  120000000000,
			utils.Weight: 20,
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.ApierV1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}
	attrSetBalance = utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.VOICE,
		Balance: map[string]interface{}{
			utils.ID:     "BALANCE2",
			utils.Value:  180000000000,
			utils.Weight: 10,
		},
	}
	if err := cdrsRpc.Call(utils.ApierV1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: <%s>", reply)
	}
	expectedVoice := 300000000000.0
	if err := cdrsRpc.Call(utils.ApierV2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply != expectedVoice {
		t.Errorf("Expecting: %v, received: %v", expectedVoice, rply)
	}
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "testV1CDRsProcessEventWithRefund",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "testV1CDRsProcessEventWithRefund",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       time.Duration(3) * time.Minute,
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	if err := cdrsRpc.Call(utils.ApierV1GetCDRs, utils.AttrGetCdrs{}, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	if err := cdrsRpc.Call(utils.ApierV2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if blc1 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE1"); blc1.Value != 0 {
		t.Errorf("Balance1 is: %s", utils.ToIJSON(blc1))
	} else if blc2 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE2"); blc2.Value != 120000000000 {
		t.Errorf("Balance2 is: %s", utils.ToIJSON(blc2))
	}
	// without re-rate we should be denied
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err == nil {
		t.Error("should receive error here")
	}
	if err := cdrsRpc.Call(utils.ApierV2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if blc1 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE1"); blc1.Value != 0 {
		t.Errorf("Balance1 is: %s", utils.ToIJSON(blc1))
	} else if blc2 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE2"); blc2.Value != 120000000000 {
		t.Errorf("Balance2 is: %s", utils.ToIJSON(blc2))
	}
	argsEv = &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, utils.MetaRerate},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "testV1CDRsProcessEventWithRefund",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "testV1CDRsProcessEventWithRefund",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Minute,
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	if err := cdrsRpc.Call(utils.ApierV2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if blc1 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE1"); blc1.Value != 120000000000 { // refund is done after debit
		t.Errorf("Balance1 is: %s", utils.ToIJSON(blc1))
	} else if blc2 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE2"); blc2.Value != 120000000000 {
		t.Errorf("Balance2 is: %s", utils.ToIJSON(blc2))
	}
	return
}

func testV1CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
