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
package v2

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
	testV2CDRsProcessCDR,
	testV2CDRsGetCdrs,
	testV2CDRsRateCDRs,
	testV2CDRsGetCdrs2,
	testV2CDRsUsageNegative,
	testV2CDRsKillEngine,
}

// Tests starting here
func TestCDRsITMySQL(t *testing.T) {
	cdrsConfDIR = "cdrsv2mysql"
	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func TestCDRsITpg(t *testing.T) {
	cdrsConfDIR = "cdrsv2psql"
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
	if cdrsCfg, err = config.NewCGRConfigFromFolder(cdrsCfgPath); err != nil {
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
}

func testV2CDRsProcessCDR(t *testing.T) {
	cgrEv := &utils.CGREvent{
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
	}
	var reply string
	if err := cdrsRpc.Call(utils.CdrsV2ProcessCDR, cgrEv, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCdrs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
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
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
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
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].ExtraFields["PayPalAccount"] != "paypal@cgrates.org" {
			t.Errorf("PayPalAccount should be added by AttributeS, have: %s",
				cdrs[0].ExtraFields["PayPalAccount"])
		}
	}
}

// Should re-rate the supplier1 cost with RP_ANY2CNT
func testV2CDRsRateCDRs(t *testing.T) {
	rpf := &utils.AttrSetRatingProfile{
		Tenant:    "cgrates.org",
		Category:  "call",
		Direction: "*out",
		Subject:   "SUPPLIER1",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime: "2018-01-01T00:00:00Z",
				RatingPlanId:   "RP_ANY2CNT"}},
		Overwrite: true}
	var reply string
	if err := cdrsRpc.Call("ApierV1.SetRatingProfile", rpf, &reply); err != nil {
		t.Error("Got error on ApierV1.SetRatingProfile: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetRatingProfile got reply: ", reply)
	}
	if err := cdrsRpc.Call(utils.CdrsV2RateCDRs,
		&utils.RPCCDRsFilter{NotRunIDs: []string{utils.MetaRaw}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated
}

func testV2CDRsGetCdrs2(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := cdrsRpc.Call("ApierV2.CountCdrs", req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}
	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testV2CDRsUsageNegative(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]interface{}{
			utils.OriginID:    "testV2CDRsUsageNegative",
			utils.OriginHost:  "192.168.1.1",
			utils.Source:      "testV2CDRsUsageNegative",
			utils.RequestType: utils.META_RATED,
			utils.Category:    "call",
			utils.Account:     "testV2CDRsUsageNegative",
			utils.Subject:     "ANY2CNT",
			utils.Destination: "+4986517174963",
			utils.AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.Usage:       -time.Duration(1) * time.Minute,
			"field_extr1":     "val_extr1",
			"fieldextr2":      "valextr2",
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.CdrsV2ProcessCDR, cgrEv, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(150) * time.Millisecond) // Give time for CDR to be rated

	var cdrs []*engine.ExternalCDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].Usage != "-1m0s" {
			t.Errorf("Unexpected usage for CDR: %s", cdrs[0].Usage)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].Usage != "0s" {
			t.Errorf("Unexpected usage for CDR: %s", cdrs[0].Usage)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"}, OriginIDs: []string{"testV2CDRsUsageNegative"}}
	if err := cdrsRpc.Call("ApierV2.GetCdrs", args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
		if cdrs[0].Usage != "0s" {
			t.Errorf("Unexpected usage for CDR: %s", cdrs[0].Usage)
		}
	}
}

func testV2CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
