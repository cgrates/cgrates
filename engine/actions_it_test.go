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
package engine

import (
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var actsLclCfg *config.CGRConfig
var actsLclRpc *rpc.Client
var actsLclCfgPath = path.Join(*dataDir, "conf", "samples", "actions")

var waitRater = flag.Int("wait_rater", 500, "Number of miliseconds to wait for rater to start and cache")

func TestActionsitInitCfg(t *testing.T) {
	// Init config first
	var err error
	actsLclCfg, err = config.NewCGRConfigFromFolder(actsLclCfgPath)
	if err != nil {
		t.Error(err)
	}
	actsLclCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(actsLclCfg)
}

func TestActionsitInitCdrDb(t *testing.T) {
	if err := InitDataDb(actsLclCfg); err != nil { // need it for versions
		t.Fatal(err)
	}
	if err := InitStorDb(actsLclCfg); err != nil {
		t.Fatal(err)
	}
}

// Finds cgr-engine executable and starts it with default configuration
func TestActionsitStartEngine(t *testing.T) {
	if _, err := StopStartEngine(actsLclCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestActionsitRpcConn(t *testing.T) {
	var err error
	// time.Sleep(500 * time.Millisecond)
	actsLclRpc, err = jsonrpc.Dial("tcp", actsLclCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestActionsitSetCdrlogDebit(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
	if err := actsLclRpc.Call("ApierV1.SetAccount", attrsSetAccount, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_1", Actions: []*utils.TPAction{
		{Identifier: DEBIT, BalanceType: utils.MONETARY, Units: "5", ExpiryTime: UNLIMITED, Weight: 20.0},
		{Identifier: CDRLOG},
	}}
	if err := actsLclRpc.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call("ApierV1.ExecuteAction", attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{Sources: []string{CDRLOG},
		Accounts: []string{attrsSetAccount.Account}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].ToR != utils.MONETARY ||
		rcvedCdrs[0].OriginHost != "127.0.0.1" ||
		rcvedCdrs[0].Source != CDRLOG ||
		rcvedCdrs[0].RequestType != utils.META_PREPAID ||
		rcvedCdrs[0].Tenant != "cgrates.org" ||
		rcvedCdrs[0].Account != "dan2904" ||
		rcvedCdrs[0].Subject != "dan2904" ||
		rcvedCdrs[0].Usage != "1" ||
		rcvedCdrs[0].RunID != DEBIT ||
		strconv.FormatFloat(rcvedCdrs[0].Cost, 'f', -1, 64) != attrsAA.Actions[0].Units {
		t.Errorf("Received: %+v", rcvedCdrs[0])
	}
}

func TestActionsitSetCdrlogTopup(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2905"}
	if err := actsLclRpc.Call("ApierV1.SetAccount", attrsSetAccount, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_2", Actions: []*utils.TPAction{
		{Identifier: TOPUP, BalanceType: utils.MONETARY, Units: "5", ExpiryTime: UNLIMITED, Weight: 20.0},
		{Identifier: CDRLOG},
	}}
	if err := actsLclRpc.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call("ApierV1.ExecuteAction", attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{Sources: []string{CDRLOG},
		Accounts: []string{attrsSetAccount.Account}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].ToR != utils.MONETARY ||
		rcvedCdrs[0].OriginHost != "127.0.0.1" ||
		rcvedCdrs[0].Source != CDRLOG ||
		rcvedCdrs[0].RequestType != utils.META_PREPAID ||
		rcvedCdrs[0].Tenant != "cgrates.org" ||
		rcvedCdrs[0].Account != "dan2905" ||
		rcvedCdrs[0].Subject != "dan2905" ||
		rcvedCdrs[0].Usage != "1" ||
		rcvedCdrs[0].RunID != TOPUP ||
		strconv.FormatFloat(rcvedCdrs[0].Cost, 'f', -1, 64) != attrsAA.Actions[0].Units {
		t.Errorf("Received: %+v", rcvedCdrs[0])
	}
}

func TestActionsitCdrlogEmpty(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_3", Actions: []*utils.TPAction{
		{Identifier: DEBIT, BalanceType: utils.MONETARY, DestinationIds: "RET",
			Units: "5", ExpiryTime: UNLIMITED, Weight: 20.0},
		{Identifier: CDRLOG},
	}}
	if err := actsLclRpc.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call("ApierV1.ExecuteAction", attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{Sources: []string{CDRLOG},
		Accounts: []string{attrsSetAccount.Account}, RunIDs: []string{DEBIT}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else {
		for _, cdr := range rcvedCdrs {
			if cdr.RunID != DEBIT {
				t.Errorf("Expecting : DEBIT, received: %+v", cdr.RunID)
			}
		}
	}
}

func TestActionsitCdrlogWithParams(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan2904"}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_4",
		Actions: []*utils.TPAction{
			{Identifier: DEBIT, BalanceType: utils.MONETARY,
				DestinationIds: "RET", Units: "25", ExpiryTime: UNLIMITED, Weight: 20.0},
			{Identifier: CDRLOG,
				ExtraParameters: `{"RequestType":"*pseudoprepaid","Subject":"DifferentThanAccount", "ToR":"~ActionType:s/^\\*(.*)$/did_$1/"}`},
			{Identifier: DEBIT_RESET, BalanceType: utils.MONETARY,
				DestinationIds: "RET", Units: "25", ExpiryTime: UNLIMITED, Weight: 20.0},
		},
	}
	if err := actsLclRpc.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call("ApierV1.ExecuteAction", attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{Sources: []string{CDRLOG},
		Accounts: []string{attrsSetAccount.Account}, RunIDs: []string{DEBIT}, RequestTypes: []string{"*pseudoprepaid"}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	}
	if err := actsLclRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{Sources: []string{CDRLOG},
		Accounts: []string{attrsSetAccount.Account}, RunIDs: []string{DEBIT_RESET}, RequestTypes: []string{"*pseudoprepaid"}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	}
}

func TestActionsitThresholdCDrLog(t *testing.T) {
	var thReply *ThresholdProfile
	var result string
	var reply string

	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "th_acc"}
	if err := actsLclRpc.Call("ApierV1.SetAccount", attrsSetAccount, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACT_TH_CDRLOG", Actions: []*utils.TPAction{
		{Identifier: TOPUP, BalanceType: utils.MONETARY, Units: "5", ExpiryTime: UNLIMITED, Weight: 20.0},
		{Identifier: CDRLOG},
	}}
	if err := actsLclRpc.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	//make sure that the threshold don't exit
	if err := actsLclRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &thReply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_Test",
		FilterIDs: []string{"*string:Account:th_acc"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:   -1,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_TH_CDRLOG"},
		Async:     false,
	}
	if err := actsLclRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := actsLclRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &thReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, thReply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, thReply)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "cdrev1",
		Event: map[string]interface{}{
			utils.EventType:   utils.CDR,
			"field_extr1":     "val_extr1",
			"fieldextr2":      "valextr2",
			utils.CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.RunID:       utils.MetaRaw,
			utils.OrderID:     123,
			utils.OriginHost:  "192.168.1.1",
			utils.Source:      utils.UNIT_TEST,
			utils.OriginID:    "dsafdsaf",
			utils.ToR:         utils.VOICE,
			utils.RequestType: utils.META_RATED,
			utils.Direction:   "*out",
			utils.Tenant:      "cgrates.org",
			utils.Category:    "call",
			utils.Account:     "th_acc",
			utils.Subject:     "th_acc",
			utils.Destination: "+4986517174963",
			utils.SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			utils.PDD:         time.Duration(0) * time.Second,
			utils.AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:       time.Duration(10) * time.Second,
			utils.SUPPLIER:    "SUPPL1",
			utils.COST:        -1.0,
		},
	}
	var ids []string
	eIDs := []string{"THD_Test"}
	if err := actsLclRpc.Call(utils.ThresholdSv1ProcessEvent, ev, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	var rcvedCdrs []*ExternalCDR
	if err := actsLclRpc.Call("ApierV2.GetCdrs", utils.RPCCDRsFilter{Sources: []string{CDRLOG},
		Accounts: []string{attrsSetAccount.Account}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].ToR != utils.MONETARY ||
		rcvedCdrs[0].OriginHost != "127.0.0.1" ||
		rcvedCdrs[0].Source != CDRLOG ||
		rcvedCdrs[0].RequestType != utils.META_PREPAID ||
		rcvedCdrs[0].Tenant != "cgrates.org" ||
		rcvedCdrs[0].Account != "th_acc" ||
		rcvedCdrs[0].Subject != "th_acc" ||
		rcvedCdrs[0].Usage != "1" ||
		rcvedCdrs[0].RunID != TOPUP ||
		strconv.FormatFloat(rcvedCdrs[0].Cost, 'f', -1, 64) != attrsAA.Actions[0].Units {
		t.Errorf("Received: %+v", rcvedCdrs[0])
	}

}

func TestActionsitStopCgrEngine(t *testing.T) {
	if err := KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
