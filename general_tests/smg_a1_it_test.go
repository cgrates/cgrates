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

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
)

var (
	smgA1CfgPath string
	smgA1Cfg     *config.CGRConfig
	smgA1rpc     *rpc.Client
)

func TestSMGa1ITLoadConfig(t *testing.T) {
	smgA1CfgPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	if smgA1Cfg, err = config.NewCGRConfigFromFolder(smgA1CfgPath); err != nil {
		t.Error(err)
	}
}

func TestSMGa1ITResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(smgA1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(smgA1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(smgA1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITRPCConn(t *testing.T) {
	time.Sleep(1500 * time.Millisecond) // flushdb takes time in mongo
	var err error
	smgA1rpc, err = jsonrpc.Dial("tcp", smgA1Cfg.RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSMGa1ITLoadTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "test", "smg_a1")}
	if err := smgA1rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	time.Sleep(time.Duration(100 * time.Millisecond))
	tStart, _ := utils.ParseDate("2017-03-03T10:39:33Z")
	tEnd, _ := utils.ParseDate("2017-03-03T12:30:13Z") // Equivalent of 10240 which is a chunk of data charged
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "data1",
		Tenant:      "cgrates.org",
		Subject:     "rpdata1",
		Destination: "data",
		TimeStart:   tStart,
		TimeEnd:     tEnd,
	}
	var cc engine.CallCost
	if err := smgA1rpc.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 0.0 || cc.RatedUsage != 10240 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}
}

func TestSMGa1ITAddBalance1(t *testing.T) {
	var reply string
	argAdd := &v1.AttrAddBalance{Tenant: "cgrates.org", Account: "rpdata1",
		BalanceType: utils.DATA, BalanceId: utils.StringPointer("rpdata1_test"),
		Value: 10000000000}
	if err := smgA1rpc.Call("ApierV1.AddBalance", argAdd, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf(reply)
	}
	argGet := &utils.AttrGetAccount{Tenant: argAdd.Tenant, Account: argAdd.Account}
	var acnt *engine.Account
	if err := smgA1rpc.Call("ApierV2.GetAccount", argGet, &acnt); err != nil {
		t.Error(err)
	} else {
		if acnt.BalanceMap[utils.DATA].GetTotalValue() != argAdd.Value { // We expect 11.5 since we have added in the previous test 1.5
			t.Errorf("Received account value: %f", acnt.BalanceMap[utils.DATA].GetTotalValue())
		}
	}
}

func TestSMGa1ITDataSession1(t *testing.T) {
	smgEv := sessionmanager.SMGenericEvent{
		utils.EVENT_NAME:         "INITIATE_SESSION",
		utils.TOR:                utils.DATA,
		utils.ACCID:              "504966119",
		utils.DIRECTION:          utils.OUT,
		utils.ACCOUNT:            "rpdata1",
		utils.SUBJECT:            "rpdata1",
		utils.DESTINATION:        "data",
		utils.CATEGORY:           "data1",
		utils.TENANT:             "cgrates.org",
		utils.REQTYPE:            utils.META_PREPAID,
		utils.SETUP_TIME:         "2017-03-03 11:39:32 +0100 CET",
		utils.ANSWER_TIME:        "2017-03-03 11:39:32 +0100 CET",
		utils.USAGE:              "10240",
		utils.SessionTTL:         "28800s",
		utils.SessionTTLLastUsed: "0s",
		utils.SessionTTLUsage:    "0s",
	}
	var maxUsage float64
	if err := smgA1rpc.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	} else if maxUsage != 10240 {
		t.Error("Received: ", maxUsage)
	}

	smgEv = sessionmanager.SMGenericEvent{
		utils.EVENT_NAME:         "UPDATE_SESSION",
		utils.ACCOUNT:            "rpdata1",
		utils.CATEGORY:           "data1",
		utils.DESTINATION:        "data",
		utils.DIRECTION:          utils.OUT,
		utils.InitialOriginID:    "504966119",
		utils.LastUsed:           "0s",
		utils.ACCID:              "504966119-1",
		utils.REQTYPE:            utils.META_PREPAID,
		utils.SessionTTL:         "28800s",
		utils.SessionTTLLastUsed: "2097152s",
		utils.SessionTTLUsage:    "0s",
		utils.SUBJECT:            "rpdata1",
		utils.TENANT:             "cgrates.org",
		utils.TOR:                utils.DATA,
		utils.SETUP_TIME:         "2017-03-03 11:39:32 +0100 CET",
		utils.ANSWER_TIME:        "2017-03-03 11:39:32 +0100 CET",
		utils.USAGE:              "2097152",
	}
	if err := smgA1rpc.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	} else if maxUsage != 2097152 {
		t.Error("Bad max usage: ", maxUsage)
	}
	smgEv = sessionmanager.SMGenericEvent{
		utils.EVENT_NAME:     "TERMINATE_SESSION",
		utils.ACCOUNT:        "rpdata1",
		utils.CATEGORY:       "data1",
		utils.DESTINATION:    "data",
		utils.DIRECTION:      utils.OUT,
		utils.LastUsed:       "2202800",
		utils.ACCID:          "504966119-1",
		utils.OriginIDPrefix: "504966119-1",
		utils.REQTYPE:        utils.META_PREPAID,
		utils.SETUP_TIME:     "2017-03-03 11:39:32 +0100 CET",
		utils.ANSWER_TIME:    "2017-03-03 11:39:32 +0100 CET",
		utils.SUBJECT:        "rpdata1",
		utils.TENANT:         "cgrates.org",
		utils.TOR:            utils.DATA,
	}
	var rpl string
	if err = smgA1rpc.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	/*
		if err := smgA1rpc.Call("SMGenericV1.ProcessCDR", smgEv, &rpl); err != nil {
			t.Error(err)
		} else if rpl != utils.OK {
			t.Errorf("Received reply: %s", rpl)
		}
		var cdrs []*engine.ExternalCDR
		req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}}
		if err := smgA1rpc.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
			t.Error("Unexpected error: ", err.Error())
		} else if len(cdrs) != 1 {
			t.Error("Unexpected number of CDRs returned: ", len(cdrs))
		} else {
			if cdrs[0].Usage != "60" {
				t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
			}
		}
	*/
}
