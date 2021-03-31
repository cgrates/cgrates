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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accPrfCfgPath   string
	accPrfCfg       *config.CGRConfig
	accSRPC         *rpc.Client
	accPrfConfigDIR string //run tests for specific configuration

	sTestsAccPrf = []func(t *testing.T){
		testAccActionsInitCfg,
		testAccActionsInitDataDb,
		testAccActionsResetStorDb,
		testAccActionsStartEngine,
		testAccActionsRPCConn,
		testAccActionsExecuteAction,
		testAccActionsExecuteAction2,
		testAccActionsGetAccountAfterActions,
		testAccActionsExecuteAction3,
		testAccActionsGetAccountAfterRemActions,
		testAccActionsKillEngine,
	}
)

//Test start here
func TestAccActionsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		accPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		accPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		accPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAccPrf {
		t.Run(accPrfConfigDIR, stest)
	}
}

func testAccActionsInitCfg(t *testing.T) {
	var err error
	accPrfCfgPath = path.Join(*dataDir, "conf", "samples", accPrfConfigDIR)
	if accPrfCfg, err = config.NewCGRConfigFromPath(accPrfCfgPath); err != nil {
		t.Error(err)
	}
}

func testAccActionsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAccActionsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAccActionsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAccActionsRPCConn(t *testing.T) {
	var err error
	accSRPC, err = newRPCClient(accPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAccActionsExecuteAction(t *testing.T) {
	var reply string
	if err := accSRPC.Call(utils.ActionSv1ExecuteActions, &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account": 1001,
			},
			APIOpts: map[string]interface{}{
				"BAL_NEW": true,
				"BAL_ADD": true,
			},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testAccActionsExecuteAction2(t *testing.T) {
	var reply string
	if err := accSRPC.Call(utils.ActionSv1ExecuteActions, &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account": 1001,
			},
			APIOpts: map[string]interface{}{
				"BAL_NEW": true,
				"BAL_ADD": true,
			},
		},
		ActionProfileIDs: []string{"CREATE_ACC"},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testAccActionsGetAccountAfterActions(t *testing.T) {
	accPrf := &utils.AccountProfile{
		Tenant: "cgrates.org",
		ID:     "1001",
		Balances: map[string]*utils.Balance{
			"MONETARY": {
				ID:      "MONETARY",
				Weights: utils.DynamicWeights{{}},
				Type:    utils.MetaConcrete,
				Units:   utils.NewDecimalFromFloat64(1048576),
				CostIncrements: []*utils.CostIncrement{{
					FilterIDs:    []string{"*string:~*req.ToR:*data"},
					Increment:    utils.NewDecimalFromFloat64(1024),
					FixedFee:     utils.NewDecimalFromFloat64(0),
					RecurrentFee: utils.NewDecimalFromFloat64(0.01),
				}},
			},
			"VOICE": {
				ID:        "VOICE",
				FilterIDs: []string{"*string:~*req.ToR:*voice"},
				Weights:   utils.DynamicWeights{{Weight: 2}},
				Type:      utils.MetaAbstract,
				Units:     utils.NewDecimalFromFloat64(2 * 10800000000000),
				CostIncrements: []*utils.CostIncrement{{
					FilterIDs:    []string{"*string:~*req.ToR:*voice"},
					Increment:    utils.NewDecimalFromFloat64(1000000000),
					FixedFee:     utils.NewDecimalFromFloat64(0),
					RecurrentFee: utils.NewDecimalFromFloat64(0.01),
				}},
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	var result *utils.AccountProfile
	if err := accSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accPrf, result) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(accPrf), utils.ToJSON(result))
	}
}

func testAccActionsExecuteAction3(t *testing.T) {
	var reply string
	if err := accSRPC.Call(utils.ActionSv1ExecuteActions, &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account": 1001,
			},
		},
		ActionProfileIDs: []string{"REM_ACC"},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testAccActionsGetAccountAfterRemActions(t *testing.T) {
	accPrf := &utils.AccountProfile{
		Tenant:       "cgrates.org",
		ID:           "1001",
		Balances:     map[string]*utils.Balance{},
		ThresholdIDs: []string{utils.MetaNone},
	}
	var result *utils.AccountProfile
	if err := accSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accPrf, result) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(accPrf), utils.ToJSON(result))
	}
}

func testAccActionsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
