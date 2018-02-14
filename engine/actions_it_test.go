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
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var actsLclCfg *config.CGRConfig
var actsLclRpc *rpc.Client
var actsLclCfgPath = path.Join(*dataDir, "conf", "samples", "actions")

var waitRater = flag.Int("wait_rater", 100, "Number of miliseconds to wait for rater to start and cache")

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
	if err := InitStorDb(actsLclCfg); err != nil {
		t.Fatal(err)
	}
}

// Finds cgr-engine executable and starts it with default configuration
func TestActionsitStartEngine(t *testing.T) {
	if _, err := StartEngine(actsLclCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestActionsitRpcConn(t *testing.T) {
	var err error
	time.Sleep(500 * time.Millisecond)
	actsLclRpc, err = jsonrpc.Dial("tcp", actsLclCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
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
		&utils.TPAction{Identifier: DEBIT, BalanceType: utils.MONETARY, Units: "5", ExpiryTime: UNLIMITED, Weight: 20.0},
		&utils.TPAction{Identifier: CDRLOG},
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
		&utils.TPAction{Identifier: TOPUP, BalanceType: utils.MONETARY, Units: "5", ExpiryTime: UNLIMITED, Weight: 20.0},
		&utils.TPAction{Identifier: CDRLOG},
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

func TestActionsitStopCgrEngine(t *testing.T) {

	if err := KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
