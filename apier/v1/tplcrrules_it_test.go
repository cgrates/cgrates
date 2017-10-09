// +build offline_tp

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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
)

var (
	tpLcrRuleCfgPath   string
	tpLcrRuleCfg       *config.CGRConfig
	tpLcrRuleRPC       *rpc.Client
	tpLcrRuleDataDir   = "/usr/share/cgrates"
	tpLcrRules         *utils.TPLcrRules
	tpLcrRuleDelay     int
	tpLcrRuleConfigDIR string //run tests for specific configuration
	tpLcrRuleID        = "*out:cgrates.org:call:1001:*any"
)

var sTestsTPLcrRules = []func(t *testing.T){
	testTPLcrRulesInitCfg,
	testTPLcrRulesResetStorDb,
	testTPLcrRulesStartEngine,
	testTPLcrRulesRpcConn,
	testTPLcrRulesGetTPLcrRulesBeforeSet,
	testTPLcrRulesSetTPLcrRules,
	testTPLcrRulesGetTPLcrRulesAfterSet,
	testTPLcrRulesGetTPLcrRuleIds,
	testTPLcrRulesUpdateTPLcrRules,
	testTPLcrRulesGetTPLcrRulesAfterUpdate,
	testTPLcrRulesRemTPLcrRules,
	testTPLcrRulesGetTPLcrRulesAfterRemove,
	testTPLcrRulesKillEngine,
}

//Test start here
func TestTPLcrRulesITMySql(t *testing.T) {
	tpLcrRuleConfigDIR = "tutmysql"
	for _, stest := range sTestsTPLcrRules {
		t.Run(tpLcrRuleConfigDIR, stest)
	}
}

func TestTPLcrRulesITMongo(t *testing.T) {
	tpLcrRuleConfigDIR = "tutmongo"
	for _, stest := range sTestsTPLcrRules {
		t.Run(tpLcrRuleConfigDIR, stest)
	}
}

func TestTPLcrRulesITPG(t *testing.T) {
	tpLcrRuleConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPLcrRules {
		t.Run(tpLcrRuleConfigDIR, stest)
	}
}

func testTPLcrRulesInitCfg(t *testing.T) {
	var err error
	tpLcrRuleCfgPath = path.Join(tpLcrRuleDataDir, "conf", "samples", tpLcrRuleConfigDIR)
	tpLcrRuleCfg, err = config.NewCGRConfigFromFolder(tpLcrRuleCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpLcrRuleCfg.DataFolderPath = tpLcrRuleDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpLcrRuleCfg)
	switch tpLcrRuleConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpLcrRuleDelay = 2000
	default:
		tpLcrRuleDelay = 1000
	}
}

// Wipe out the cdr database
func testTPLcrRulesResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpLcrRuleCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPLcrRulesStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpLcrRuleCfgPath, tpLcrRuleDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPLcrRulesRpcConn(t *testing.T) {
	var err error
	tpLcrRuleRPC, err = jsonrpc.Dial("tcp", tpLcrRuleCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPLcrRulesGetTPLcrRulesBeforeSet(t *testing.T) {
	var reply *utils.TPRatingPlan
	if err := tpLcrRuleRPC.Call("ApierV1.GetTPLcrRule", &AttrGetTPLcrRules{TPid: "TPLRC1", LcrRuleId: tpLcrRuleID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPLcrRulesSetTPLcrRules(t *testing.T) {
	tpLcrRules = &utils.TPLcrRules{
		TPid:      "TPLRC1",
		Direction: "*out",
		Tenant:    "cgrates.org",
		Category:  "call",
		Account:   "1001",
		Subject:   "*any",
		Rules: []*utils.TPLcrRule{
			&utils.TPLcrRule{
				DestinationId:  "DST_1002",
				RpCategory:     "lcr_profile1",
				Strategy:       "*static",
				StrategyParams: "suppl2;suppl1",
				ActivationTime: "05:00:00",
				Weight:         10,
			},
			&utils.TPLcrRule{
				DestinationId:  "*any",
				RpCategory:     "lcr_profile1",
				Strategy:       "*highest_cost",
				StrategyParams: "",
				ActivationTime: "05:00:00",
				Weight:         10,
			},
		},
	}
	var result string
	if err := tpLcrRuleRPC.Call("ApierV1.SetTPLcrRule", tpLcrRules, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPLcrRulesGetTPLcrRulesAfterSet(t *testing.T) {
	var reply *utils.TPLcrRules
	if err := tpLcrRuleRPC.Call("ApierV1.GetTPLcrRule", &AttrGetTPLcrRules{TPid: "TPLRC1", LcrRuleId: tpLcrRuleID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpLcrRules.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpLcrRules.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpLcrRules.GetLcrRuleId(), reply.GetLcrRuleId()) {
		t.Errorf("Expecting : %+v, received: %+v", tpLcrRules.GetLcrRuleId(), reply.GetLcrRuleId())
	} else if !reflect.DeepEqual(len(tpLcrRules.Rules), len(reply.Rules)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpLcrRules.Rules), len(reply.Rules))
	}

}

func testTPLcrRulesGetTPLcrRuleIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"*out:cgrates.org:call:1001:*any"}
	if err := tpLcrRuleRPC.Call("ApierV1.GetTPLcrRuleIds", &AttrGetTPLcrIds{TPid: "TPLRC1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPLcrRulesUpdateTPLcrRules(t *testing.T) {
	tpLcrRules.Rules = []*utils.TPLcrRule{
		&utils.TPLcrRule{
			DestinationId:  "DST_1002",
			RpCategory:     "lcr_profile1",
			Strategy:       "*static",
			StrategyParams: "suppl2;suppl1",
			ActivationTime: "03:00:00",
			Weight:         10,
		},
		&utils.TPLcrRule{
			DestinationId:  "*any",
			RpCategory:     "lcr_profile1",
			Strategy:       "*highest_cost",
			StrategyParams: "",
			ActivationTime: "05:00:00",
			Weight:         10,
		},
		&utils.TPLcrRule{
			DestinationId:  "*any",
			RpCategory:     "lcr_profile2",
			Strategy:       "*load_distribution",
			StrategyParams: "supplier1:5",
			ActivationTime: "01:00:00",
			Weight:         10,
		},
	}
	var result string
	if err := tpLcrRuleRPC.Call("ApierV1.SetTPLcrRule", tpLcrRules, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPLcrRulesGetTPLcrRulesAfterUpdate(t *testing.T) {
	var reply *utils.TPLcrRules
	if err := tpLcrRuleRPC.Call("ApierV1.GetTPLcrRule", &AttrGetTPLcrRules{TPid: "TPLRC1", LcrRuleId: tpLcrRuleID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpLcrRules.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpLcrRules.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpLcrRules.GetLcrRuleId(), reply.GetLcrRuleId()) {
		t.Errorf("Expecting : %+v, received: %+v", tpLcrRules.GetLcrRuleId(), reply.GetLcrRuleId())
	} else if !reflect.DeepEqual(len(tpLcrRules.Rules), len(reply.Rules)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpLcrRules.Rules), len(reply.Rules))
	}
}

func testTPLcrRulesRemTPLcrRules(t *testing.T) {
	var resp string
	if err := tpLcrRuleRPC.Call("ApierV1.RemTPLcrRule", &AttrRemTPLcrRules{TPid: tpLcrRules.TPid, Direction: tpLcrRules.Direction, Tenant: tpLcrRules.Tenant, Category: tpLcrRules.Category, Account: tpLcrRules.Account, Subject: tpLcrRules.Subject}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPLcrRulesGetTPLcrRulesAfterRemove(t *testing.T) {
	var reply *utils.TPRatingPlan
	if err := tpLcrRuleRPC.Call("ApierV1.GetTPLcrRule", &AttrGetTPLcrRules{TPid: "TPLRC1", LcrRuleId: tpLcrRuleID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testTPLcrRulesKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpLcrRuleDelay); err != nil {
		t.Error(err)
	}
}
