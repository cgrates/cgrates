//go:build integration
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

package apis

import (
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
)

var (
	preloadCFG     *config.CGRConfig
	preloadRPC     *birpc.Client
	preloadCfgPath string
	preloadCfgDir  string

	preloadTests = []func(t *testing.T){
		testPreloadITCreateDirectories,
		testPreloadITInitCfg,
		testPreloadITStartEngine,
		testPreloadITRPCConn,
		testPreloadITVerifyRateProfile,
		testPreloadITVerifyAccounts,
		testPreloadITVerifyActionProfiles,
		testCleanupFiles,
		testPreloadITKillEngine,
	}
)

func TestPreloadIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		preloadCfgDir = "preload_internal"
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatalf("Unsuported dbtype: %q", *dbType)
	}
	for _, test := range preloadTests {
		t.Run("Running TestPreloadIT:", test)
	}
}

func testPreloadITCreateDirectories(t *testing.T) {
	// creating the directories
	for _, dir := range []string{"/tmp/RatesIn", "/tmp/RatesOut",
		"/tmp/AccountsIn", "/tmp/AccountsOut",
		"/tmp/ActionsIn", "/tmp/ActionsOut"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("Error when removing the directory: %s because of %v", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Cannot create directory %s because of %v", dir, err)
		}
	}
	// writing in files the csv containing the profile for RateProfile
	if err := os.WriteFile(path.Join("/tmp/RatesIn", utils.RatesCsv), []byte(`
#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,*string:~*req.Subject:1001,;0,0.1,0.6,*free,RT_WEEK,,"* * * * 1-5",;0,false,0s,0,0.12,1m,1m
cgrates.org,RP1,,,,,,RT_WEEK,,,,,1m,1.234,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,0.089,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s
`), 0644); err != nil {
		t.Fatalf("Err %v when writing in file %s", err, utils.RatesCsv)
	}
	// writing in files the csv containing the profile for Accounts
	if err := os.WriteFile(path.Join("/tmp/AccountsIn", utils.AccountsCsv), []byte(`
#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,,;20,,,MonetaryBalance,,;10,,*monetary,14,fltr1&fltr2;100;fltr3;200,,fltr1&fltr2;1.3;2.3;3.3,attr1;attr2,,*none
cgrates.org,1001,,,,,VoiceBalance,,;10,*string:~*req.Destination:1002;true;;false,*voice,1h,,,,,,
`), 0644); err != nil {
		t.Fatalf("Err %v when writing in file %s", err, utils.AccountsCsv)
	}
	// writing in files the csv containing the profile for ActionProfile
	if err := os.WriteFile(path.Join("/tmp/ActionsIn", utils.ActionsCsv), []byte(`
#Tenant,ID,FilterIDs,Weights,Blockers,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ONE_TIME_ACT,,;10,,*asap,*accounts,1001;1002,TOPUP,,0s,*add_balance,,*balance.TestBalance.Value,10
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_DATA,,0s,*set_balance,,*balance.TestDataBalance.Type,*data
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_DATA,,0s,*add_balance,,*balance.TestDataBalance.Value,1024
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_VOICE,,0s,*set_balance,,*balance.TestVoiceBalance.Type,*voice
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,0s,*add_balance,,*balance.TestVoiceBalance.Value,15m15s
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,0s,*add_balance,,*balance.TestVoiceBalance2.Value,15m15s
`), 0644); err != nil {
		t.Fatalf("Err %v when writing in file %s", err, utils.ActionsCsv)
	}
}

func testPreloadITInitCfg(t *testing.T) {
	var err error
	preloadCfgPath = path.Join(*dataDir, "conf", "samples", preloadCfgDir)
	if preloadCFG, err = config.NewCGRConfigFromPath(context.Background(), preloadCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testPreloadITStartEngine(t *testing.T) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Error(err)
	}
	engine := exec.Command(enginePath, "-config_path", preloadCfgPath, "-preload",
		"Rates_Loader,Accounts_Loader,Actions_Loader")
	if err := engine.Start(); err != nil {
		t.Error(err)
	}

	fib := utils.Fib()
	var connected bool
	for i := 0; i < 25; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err := jsonrpc.Dial(utils.TCP, preloadCFG.ListenCfg().RPCJSONListen); err != nil {
			t.Logf("Error <%s> when opening test connection to: <%s>",
				err.Error(), preloadCFG.ListenCfg().RPCJSONListen)
		} else {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("Engine did not open at port %v", preloadCFG.ListenCfg().RPCJSONListen)
	}
	time.Sleep(100 * time.Millisecond)
}

func testPreloadITRPCConn(t *testing.T) {
	var err error
	if preloadRPC, err = newRPCClient(preloadCFG.ListenCfg()); err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testPreloadITVerifyRateProfile(t *testing.T) {
	var reply *utils.RateProfile
	expected := &utils.RateProfile{
		ID:        "RP1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				ActivationTimes: "* * 24 12 *",
				Weights: []*utils.DynamicWeight{{
					Weight: 30,
				}},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(564, 4),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * 1-5",
				Weights: []*utils.DynamicWeight{{
					Weight: 0,
				}},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						FixedFee:      utils.NewDecimal(0, 0),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Minute), 0),
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						FixedFee:      utils.NewDecimal(1234, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				ActivationTimes: "* * * * 0,6",
				Weights: []*utils.DynamicWeight{{
					Weight: 10,
				}},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(89, 3),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	if err := preloadRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RP1", Tenant: "cgrates.org"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testPreloadITVerifyAccounts(t *testing.T) {
	expAcc := &utils.Account{
		Tenant: utils.CGRateSorg,
		ID:     "1001",
		Opts:   map[string]any{},
		Weights: []*utils.DynamicWeight{
			{
				Weight: 20,
			},
		},
		Balances: map[string]*utils.Balance{
			"MonetaryBalance": {
				ID: "MonetaryBalance",
				Weights: []*utils.DynamicWeight{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaMonetary,
				Units: utils.NewDecimal(14, 0),
				UnitFactors: []*utils.UnitFactor{
					{
						FilterIDs: []string{"fltr1", "fltr2"},
						Factor:    utils.NewDecimal(100, 0),
					},
					{
						FilterIDs: []string{"fltr3"},
						Factor:    utils.NewDecimal(200, 0),
					},
				},
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"fltr1", "fltr2"},
						Increment:    utils.NewDecimal(13, 1),
						FixedFee:     utils.NewDecimal(23, 1),
						RecurrentFee: utils.NewDecimal(33, 1),
					},
				},
				AttributeIDs: []string{"attr1", "attr2"},
				Opts:         map[string]any{},
			},
			"VoiceBalance": {
				ID: "VoiceBalance",
				Weights: []*utils.DynamicWeight{
					{
						Weight: 10,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						FilterIDs: []string{"*string:~*req.Destination:1002"},
						Blocker:   true,
					},
					{
						Blocker: false,
					},
				},
				Type:  utils.MetaVoice,
				Units: utils.NewDecimal(int64(time.Hour), 0),
				Opts:  map[string]any{},
			},
		},
		ThresholdIDs: []string{"*none"},
	}
	var reply *utils.Account
	if err := preloadRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "1001", Tenant: "cgrates.org"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAcc, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expAcc), utils.ToJSON(reply))
	}
}

func testPreloadITVerifyActionProfiles(t *testing.T) {
	var reply *engine.ActionProfile
	expected := &engine.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "ONE_TIME_ACT",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Schedule: utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			"*accounts": {
				"1001": {},
				"1002": {},
			},
		},
		Actions: []*engine.APAction{
			{
				ID:   "TOPUP",
				TTL:  0,
				Type: utils.MetaAddBalance,
				Opts: map[string]any{},
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestBalance.Value",
						Value: "10",
					},
				},
			},
			{
				ID:   "SET_BALANCE_TEST_DATA",
				TTL:  0,
				Type: utils.MetaSetBalance,
				Opts: map[string]any{},
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestDataBalance.Type",
						Value: utils.MetaData,
					},
				},
			},
			{
				ID:   "TOPUP_TEST_DATA",
				TTL:  0,
				Type: utils.MetaAddBalance,
				Opts: map[string]any{},
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestDataBalance.Value",
						Value: "1024",
					},
				},
			},
			{
				ID:   "SET_BALANCE_TEST_VOICE",
				TTL:  0,
				Type: utils.MetaSetBalance,
				Opts: map[string]any{},
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestVoiceBalance.Type",
						Value: utils.MetaVoice,
					},
				},
			},
			{
				ID:   "TOPUP_TEST_VOICE",
				TTL:  0,
				Type: utils.MetaAddBalance,
				Opts: map[string]any{},
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestVoiceBalance.Value",
						Value: "15m15s",
					},
					{
						Path:  "*balance.TestVoiceBalance2.Value",
						Value: "15m15s",
					},
				},
			},
		},
	}
	if err := preloadRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "ONE_TIME_ACT", Tenant: "cgrates.org"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}

}

func testCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/RatesIn", "/tmp/RatesOut",
		"/tmp/AccountsIn", "/tmp/AccountsOut",
		"/tmp/ActionsIn", "/tmp/ActionsOut"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func testPreloadITKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
