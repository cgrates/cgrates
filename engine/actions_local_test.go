/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM

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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var actsLclCfg *config.CGRConfig
var actsLclRpc *rpc.Client
var actsLclCfgPath = path.Join(*dataDir, "conf", "samples", "cgradmin")

func TestActionsLocalInitCfg(t *testing.T) {
	if !*testLocal {
		return
	}
	// Init config first
	var err error
	actsLclCfg, err = config.NewCGRConfigFromFolder(actsLclCfgPath)
	if err != nil {
		t.Error(err)
	}
	actsLclCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(actsLclCfg)
}

func TestActionsLocalInitCdrDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := InitStorDb(actsLclCfg); err != nil {
		t.Fatal(err)
	}
}

// Finds cgr-engine executable and starts it with default configuration
func TestActionsLocalStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if _, err := StartEngine(actsLclCfgPath, waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestActionsLocalRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	actsLclRpc, err = jsonrpc.Dial("tcp", actsLclCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func TestActionsLocalSetCdrlogActions(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Direction: utils.OUT, Account: "dan2904"}
	if err := actsLclRpc.Call("ApierV1.SetAccount", attrsSetAccount, &reply); err != nil {
		t.Error("Got error on ApierV1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
	}
	attrsAA := &utils.AttrSetActions{ActionsId: "ACTS_1", Actions: []*utils.TPAction{
		&utils.TPAction{Identifier: DEBIT, BalanceType: utils.MONETARY, Direction: attrsSetAccount.Direction, Units: 5.0, ExpiryTime: UNLIMITED, Weight: 20.0},
		&utils.TPAction{Identifier: CDRLOG},
	}}
	if err := actsLclRpc.Call("ApierV1.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ERR_EXISTS {
		t.Error("Got error on ApierV1.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Direction: attrsSetAccount.Direction, Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: attrsAA.ActionsId}
	if err := actsLclRpc.Call("ApierV1.ExecuteAction", attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}
	var rcvedCdrs []*ExternalCdr
	if err := actsLclRpc.Call("ApierV2.GetCdrs", utils.RpcCdrsFilter{CdrSources: []string{CDRLOG}}, &rcvedCdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(rcvedCdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(rcvedCdrs))
	} else if rcvedCdrs[0].TOR != utils.MONETARY ||
		rcvedCdrs[0].CdrHost != "127.0.0.1" ||
		rcvedCdrs[0].CdrSource != CDRLOG ||
		rcvedCdrs[0].ReqType != utils.META_PREPAID ||
		rcvedCdrs[0].Tenant != "cgrates.org" ||
		rcvedCdrs[0].Account != "dan2904" ||
		rcvedCdrs[0].Subject != "dan2904" ||
		rcvedCdrs[0].Usage != "1" ||
		rcvedCdrs[0].MediationRunId != utils.META_DEFAULT ||
		rcvedCdrs[0].Cost != attrsAA.Actions[0].Units {
		t.Errorf("Received: %+v", rcvedCdrs[0])
	}

}

func TestActionsLocalStopCgrEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	if err := KillEngine(waitRater); err != nil {
		t.Error(err)
	}
}
