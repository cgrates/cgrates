//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package general_tests

import (
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accPrfCfgPath   string
	accPrfCfg       *config.CGRConfig
	accSRPC         *birpc.Client
	accPrfConfigDIR string //run tests for specific configuration

	sTestsAccPrf = []func(t *testing.T){
		testAccActionsInitCfg,
		testAccActionsResetDBs,
		testAccActionsStartEngine,
		testAccActionsRPCConn,
		testAccActionsSetActionProfile,
		testAccActionsExecuteAction,
		testAccActionsSetActionProfile,
		testAccActionsExecuteAction2,
		testAccActionsGetAccountAfterActions,
		testAccActionsSetActionProfile2,
		testAccActionsExecuteAction3,
		testAccActionsGetAccountAfterRemActions,
		testAccActionsKillEngine,
	}
)

// Test start here
func TestAccActionsIT(t *testing.T) {
	switch *utils.DBType {
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
	accPrfCfgPath = path.Join(*utils.DataDir, "conf", "samples", accPrfConfigDIR)
	if accPrfCfg, err = config.NewCGRConfigFromPath(context.Background(), accPrfCfgPath); err != nil {
		t.Error(err)
	}
}

func testAccActionsResetDBs(t *testing.T) {
	if err := engine.InitDataDB(accPrfCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAccActionsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accPrfCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAccActionsRPCConn(t *testing.T) {
	accSRPC = engine.NewRPCClient(t, accPrfCfg.ListenCfg(), *utils.Encoding)
}

func testAccActionsSetActionProfile(t *testing.T) {
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "CREATE_ACC",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
			Schedule: utils.MetaASAP,
			Actions: []*utils.APAction{
				{
					ID:        "SET_NEW_BAL",
					FilterIDs: []string{"*exists:*opts.BAL_NEW:"},
					Type:      utils.MetaSetBalance,
					Diktats: []*utils.APDiktat{
						{
							ID: "SetAccThresholds",
							Opts: map[string]any{
								"*balancePath":  "*account.ThresholdIDs",
								"*balanceValue": utils.MetaNone,
							},
						},
						{
							ID: "SetBalMonetaryType",
							Opts: map[string]any{
								"*balancePath":  "*balance.MONETARY.Type",
								"*balanceValue": utils.MetaConcrete,
							},
						},
						{
							ID: "SetBalMonetaryUnits",
							Opts: map[string]any{
								"*balancePath":  "*balance.MONETARY.Units",
								"*balanceValue": "1048576",
							},
						},
						{
							ID: "SetBalMonetaryWeights",
							Opts: map[string]any{
								"*balancePath":  "*balance.MONETARY.Weights",
								"*balanceValue": "`;0`",
							},
						},
						{
							ID: "SetBalMonetaryCostIncrements",
							Opts: map[string]any{
								"*balancePath":  "*balance.MONETARY.CostIncrements",
								"*balanceValue": "`*string:~*req.ToR:*data;1024;0;0.01`",
							},
						},
					},
				},
				{
					ID:        "SET_ADD_BAL",
					FilterIDs: []string{"*exists:*opts.BAL_ADD:"},
					Type:      utils.MetaAddBalance,
					Diktats: []*utils.APDiktat{
						{
							ID: "AddBalVoiceType",
							Opts: map[string]any{
								"*balancePath":  "*balance.VOICE.Type",
								"*balanceValue": utils.MetaAbstract,
							},
						},
						{
							ID: "AddBalVoiceUnits",
							Opts: map[string]any{
								"*balancePath":  "*balance.VOICE.Units",
								"*balanceValue": strconv.FormatInt((3 * time.Hour).Nanoseconds(), 10),
							},
						},
						{
							ID: "AddBalVoiceFilters",
							Opts: map[string]any{
								"*balancePath":  "*balance.VOICE.FilterIDs",
								"*balanceValue": "`*string:~*req.ToR:*voice`",
							},
						},
						{
							ID: "AddBalVoiceWeights",
							Opts: map[string]any{
								"*balancePath":  "*balance.VOICE.Weights",
								"*balanceValue": "`;2`",
							},
						},
						{
							ID: "AddBalVoiceCostIncrements",
							Opts: map[string]any{
								"*balancePath":  "*balance.VOICE.CostIncrements",
								"*balanceValue": "`*string:~*req.ToR:*voice;1000000000;0;0.01`",
							},
						},
					},
				},
			},
		},
		APIOpts: map[string]any{},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetActionProfile, actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *utils.ActionProfile
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: actPrf.Tenant, ID: actPrf.ID}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrf.ActionProfile, result) {
		t.Errorf("Expecting : %+v, received: %+v", actPrf.ActionProfile, result)
	}
}

func testAccActionsExecuteAction(t *testing.T) {
	var reply string
	if err := accSRPC.Call(context.Background(), utils.ActionSv1ExecuteActions, &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Account": 1001,
		},
		APIOpts: map[string]any{
			"BAL_NEW": true,
			"BAL_ADD": true,
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testAccActionsExecuteAction2(t *testing.T) {
	var reply string
	if err := accSRPC.Call(context.Background(), utils.ActionSv1ExecuteActions, &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Account": 1001,
		},
		APIOpts: map[string]any{
			"BAL_NEW":                   true,
			"BAL_ADD":                   true,
			utils.OptsActionsProfileIDs: []string{"CREATE_ACC"},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testAccActionsGetAccountAfterActions(t *testing.T) {
	accPrf := &utils.Account{
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
	var result *utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accPrf, result) {
		t.Errorf("Expecting : %s, received: %s", utils.ToJSON(accPrf), utils.ToJSON(result))
	}
}

func testAccActionsSetActionProfile2(t *testing.T) {
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "REM_ACC",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
			Schedule: utils.MetaASAP,
			Actions: []*utils.APAction{
				{
					ID:   "REM_BAL",
					Type: utils.MetaRemBalance,
					Diktats: []*utils.APDiktat{
						{
							ID: "RemBalMonetary",
							Opts: map[string]any{
								"*balancePath": "MONETARY",
							},
						},
						{
							ID: "RemBalVoice",
							Opts: map[string]any{
								"*balancePath": "VOICE",
							},
						},
					},
				},
			},
		},
		APIOpts: map[string]any{},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetActionProfile, actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *utils.ActionProfile
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: actPrf.Tenant, ID: actPrf.ID}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actPrf.ActionProfile, result) {
		t.Errorf("Expecting : %+v, received: %+v", actPrf.ActionProfile, result)
	}
}

func testAccActionsExecuteAction3(t *testing.T) {
	var reply string
	if err := accSRPC.Call(context.Background(), utils.ActionSv1ExecuteActions, &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Account": 1001,
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"REM_ACC"},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testAccActionsGetAccountAfterRemActions(t *testing.T) {
	accPrf := &utils.Account{
		Tenant:       "cgrates.org",
		ID:           "1001",
		Balances:     map[string]*utils.Balance{},
		ThresholdIDs: []string{utils.MetaNone},
	}
	var result *utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount, &utils.TenantIDWithAPIOpts{
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
