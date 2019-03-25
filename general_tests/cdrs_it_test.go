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
	"net/rpc/jsonrpc"
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

// subtests to be executed for each confDIR
var sTestsCDRsIT = []func(t *testing.T){
	testV2CDRsInitConfig,
	testV2CDRsInitDataDb,
	testV2CDRsInitCdrDb,
	testV2CDRsStartEngine,
	testV2CDRsRpcConn,
	testV2CDRsLoadTariffPlanFromFolder,
	//default process
	testV2CDRsProcessCDR,
	testV2CDRsGetCdrs,
	//custom process
	testV2CDRsProcessCDR2,
	testV2CDRsGetCdrs2,
	testV2CDRsProcessCDR3,
	testV2CDRsGetCdrs3,

	testV2CDRsKillEngine,
}

// Tests starting here
func TestCDRsITMySQL(t *testing.T) {
	cdrsConfDIR = "cdrsv2mysql"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITMongo(t *testing.T) {
	cdrsConfDIR = "cdrsv2mongo"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func testV2CDRsInitConfig(t *testing.T) {
	var err error
	cdrsCfgPath = path.Join(*dataDir, "conf", "samples", cdrsConfDIR)
	if cdrsCfg, err = config.NewCGRConfigFromPath(cdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV2CDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testV2CDRsInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV2CDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testV2CDRsRpcConn(t *testing.T) {
	cdrsRpc, err = jsonrpc.Dial("tcp", cdrsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV2CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := cdrsRpc.Call("ApierV2.LoadTariffPlanFromFolder",
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
	var resp string
	if err := cdrsRpc.Call("ApierV1.RemoveChargerProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.AttributeProfile
	if err := cdrsRpc.Call("ApierV1.GetChargerProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV2CDRsProcessCDR(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsProcessCDR1",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDRsProcessCDR",
				utils.RequestType: utils.META_RATED,
				utils.Category:    "call",
				utils.Account:     "testV2CDRsProcessCDR",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCDRs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 2 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
}

//Disable Attributes process
func testV2CDRsProcessCDR2(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		AttributeS: utils.BoolPointer(false),
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsProcessCDR2",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDRsProcessCDR2",
				utils.RequestType: utils.META_RATED,
				utils.Category:    "call",
				utils.Account:     "testV2CDRsProcessCDR2",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrs2(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{Accounts: []string{"testV2CDRsProcessCDR2"}}
	if err := cdrsRpc.Call("ApierV2.CountCDRs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 2 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, OriginIDs: []string{"testV2CDRsProcessCDR2"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		//we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := cdrs[0].ExtraFields["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"testV2CDRsProcessCDR2"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		//we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := cdrs[0].ExtraFields["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
}

//Disable Attributes and Charger process
func testV2CDRsProcessCDR3(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		AttributeS: utils.BoolPointer(false),
		ChargerS:   utils.BoolPointer(false),
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:    "testV2CDRsProcessCDR3",
				utils.OriginHost:  "192.168.1.1",
				utils.Source:      "testV2CDRsProcessCDR3",
				utils.RequestType: utils.META_RATED,
				utils.Category:    "call",
				utils.Account:     "testV2CDRsProcessCDR3",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "+4986517174963",
				utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:       time.Duration(1) * time.Minute,
				"field_extr1":     "val_extr1",
				"fieldextr2":      "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrs3(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{Accounts: []string{"testV2CDRsProcessCDR3"}}
	if err := cdrsRpc.Call("ApierV2.CountCDRs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 1 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, OriginIDs: []string{"testV2CDRsProcessCDR3"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		//we disable the connection to AttributeS and PayPalAccount shouldn't be present
		if _, has := cdrs[0].ExtraFields["PayPalAccount"]; has {
			t.Errorf("PayPalAccount should NOT be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"testV2CDRsProcessCDR3"}}
	if err := cdrsRpc.Call(utils.ApierV2GetCDRs, args, &cdrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error("Unexpected error: ", err.Error())
	}
}

func testV2CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
