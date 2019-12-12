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
package general_tests

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var pecdrsCfgPath string
var pecdrsCfg *config.CGRConfig
var pecdrsRpc *rpc.Client
var pecdrsConfDIR string

var sTestsCDRsIT_ProcessEvent = []func(t *testing.T){
	testV1CDRsInitConfig,
	testV1CDRsInitDataDb,
	testV1CDRsInitCdrDb,
	testV1CDRsStartEngine,
	testV1CDRsRpcConn,
	testV1CDRsLoadTariffPlanFromFolder,
	testV1CDRsProcessEvent,
	testV1CDRsKillEngine,
}

func TestCDRsITPEInternal(t *testing.T) {
	pecdrsConfDIR = "cdrsv1preocessevent"
	for _, stest := range sTestsCDRsIT_ProcessEvent {
		t.Run(pecdrsConfDIR, stest)
	}
}

// func TestCDRsITPEMongo(t *testing.T) {
// 	pecdrsConfDIR = "cdrsv1mongo"
// 	for _, stest := range sTestsCDRsIT_ProcessEvent {
// 		t.Run(pecdrsConfDIR, stest)
// 	}
// }

// func TestCDRsITPEMySql(t *testing.T) {
// 	pecdrsConfDIR = "cdrsv1mysql"
// 	for _, stest := range sTestsCDRsIT_ProcessEvent {
// 		t.Run(pecdrsConfDIR, stest)
// 	}
// }

func testV1CDRsInitConfig(t *testing.T) {
	var err error
	pecdrsCfgPath = path.Join(*dataDir, "conf", "samples", pecdrsConfDIR)
	if pecdrsCfg, err = config.NewCGRConfigFromPath(pecdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV1CDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(pecdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(pecdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(pecdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsRpcConn(t *testing.T) {
	var err error
	pecdrsRpc, err = newRPCClient(pecdrsCfg.ListenCfg())
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst string
	if err := pecdrsRpc.Call(utils.ApierV1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testV1CDRsProcessEvent(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "test1_processEvent"}
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
	if err := pecdrsRpc.Call(utils.ApierV1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}
	expectedVoice := 120000000000.0
	if err := pecdrsRpc.Call(utils.ApierV2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply != expectedVoice {
		t.Errorf("Expecting: %v, received: %v", expectedVoice, rply)
	}
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaAttributes, utils.MetaStore},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.RunID:       "testv1",
				utils.OriginID:    "test1_processEvent",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var cdrs []*engine.ExternalCDR
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.META_ANY},
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*engine.Attribute{
			{
				FieldName: utils.Subject,
				Value:     config.NewRSRParsersMustCompile("1011", true, utils.INFIELD_SEP),
			},
		},
		Weight: 200,
	}
	alsPrf.Compile()
	var result string
	if err := pecdrsRpc.Call(utils.ApierV1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAt *engine.AttributeProfile
	if err := pecdrsRpc.Call(utils.ApierV1GetAttributeProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &replyAt); err != nil {
		t.Error(err)
	}
	replyAt.Compile()
	if !reflect.DeepEqual(alsPrf, replyAt) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
	if err := pecdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	if err := pecdrsRpc.Call(utils.ApierV1GetCDRs, utils.AttrGetCdrs{}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	}
	return
}

func testV1CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
