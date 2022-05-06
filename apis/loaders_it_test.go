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
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	ldrCfgPath   string
	ldrCfg       *config.CGRConfig
	ldrRPC       *birpc.Client
	ldrConfigDIR string //run tests for specific configuration

	sTestsLdr = []func(t *testing.T){
		testLoadersRemoveFolders,
		testLoadersCreateFolders,

		testLoadersInitCfg,
		testLoadersInitDataDB,
		testLoadersStartEngine,
		testLoadersRPCConn,

		testLoadersWriteCSVs,
		testLoadersLoad,

		testLoadersGetAccounts,
		testLoadersGetActionProfiles,
		testLoadersGetAttributeProfiles,
		testLoadersGetChargerProfiles,
		testLoadersGetDispatcherProfiles,
		testLoadersGetDispatcherHosts,
		testLoadersGetFilters,
		testLoadersGetRateProfiles,
		testLoadersGetResourceProfiles,
		testLoadersGetRouteProfiles,
		testLoadersGetStatQueueProfiles,
		testLoadersGetThresholdProfiles,

		testLoadersRemove,
		testLoadersGetAccountAfterRemove,
		testLoadersGetActionProfileAfterRemove,
		testLoadersGetAttributeProfileAfterRemove,
		testLoadersGetChargerProfileAfterRemove,
		testLoadersGetDispatcherProfileAfterRemove,
		testLoadersGetDispatcherHostAfterRemove,
		testLoadersGetFilterAfterRemove,
		testLoadersGetRateProfileAfterRemove,
		testLoadersGetResourceProfileAfterRemove,
		testLoadersGetRouteProfileAfterRemove,
		testLoadersGetStatQueueProfileAfterRemove,
		testLoadersGetThresholdProfileAfterRemove,

		testLoadersRemoveFolders,

		testLoadersPing,
		testLoadersKillEngine,
	}
)

func TestLoadersIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		ldrConfigDIR = "apis_loaders_internal"
	case utils.MetaMongo:
		ldrConfigDIR = "apis_loaders_mongo"
	case utils.MetaMySQL:
		ldrConfigDIR = "apis_loaders_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsLdr {
		t.Run(ldrConfigDIR, stest)
	}
}

func testLoadersInitCfg(t *testing.T) {
	var err error
	ldrCfgPath = path.Join(*dataDir, "conf", "samples", ldrConfigDIR)
	ldrCfg, err = config.NewCGRConfigFromPath(context.Background(), ldrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testLoadersInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(ldrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testLoadersStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ldrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testLoadersRPCConn(t *testing.T) {
	var err error
	ldrRPC, err = newRPCClient(ldrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testLoadersKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testLoadersPing(t *testing.T) {
	var reply string
	if err := ldrRPC.Call(context.Background(), utils.LoaderSv1Ping,
		new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testLoadersCreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestLoadersIT/in", 0755); err != nil {
		t.Error(err)
	}
}

func testLoadersRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestLoadersIT/in"); err != nil {
		t.Error(err)
	}
}

func testLoadersWriteCSVs(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvAccounts, err := os.Create(path.Join(ldrCfg.LoaderCfg()[0].TpInDir, fileName))
		if err != nil {
			return err
		}
		defer csvAccounts.Close()
		_, err = csvAccounts.WriteString(data)
		if err != nil {
			return err

		}
		return csvAccounts.Sync()
	}
	// Create and populate Accounts.csv
	if err := writeFile(utils.AccountsCsv, `
#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,,,,,,,,,,,,,,,,
cgrates.org,1001,,;20,;false,,,,,,,,,,,,,
cgrates.org,1001,,,,,MonetaryBalance,,;10,*string:~*req.Account:1002;true;;false,*monetary,14,fltr1&fltr2;100;fltr3;200,,fltr1&fltr2;1.3;2.3;3.3,attr1;attr2,,*none
cgrates.org,1001,,,,,,,,,,,,,,,,
cgrates.org,1001,,,,,VoiceBalance,,;10,,*voice,1h,,,,,,
cgrates.org,1002,,,,,MonetaryBalance,,;20,,*monetary,1h,,,,,,
cgrates.org,1002,,;30,;false,,VoiceBalance,,;10,,*voice,14,fltr3&fltr4;150;fltr5;250,,fltr3&fltr4;1.3;2.3;3.3,attr3;attr4,,*none
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate ActionProfiles.csv
	if err := writeFile(utils.ActionsCsv, `
#Tenant,ID,FilterIDs,Weights,Blockers,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ONE_TIME_ACT,,,,,,,,,,,,,
cgrates.org,ONE_TIME_ACT,,;10,;true,*asap,*accounts,1001;1002,,,,,,,
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP,,0s,*add_balance,,,
cgrates.org,ONE_TIME_ACT,,,,*asap,*accounts,1001;1002,,,,,,,
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP,,,,,*balance.TestBalance.Value,10
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_DATA,,0s,*set_balance,,*balance.TestDataBalance.Type,*data
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_DATA,,0s,*add_balance,,*balance.TestDataBalance.Value,1024
cgrates.org,ONE_TIME_ACT,,,,,,,SET_BALANCE_TEST_VOICE,,0s,*set_balance,,*balance.TestVoiceBalance.Type,*voice
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,0s,*add_balance,,*balance.TestVoiceBalance.Value,15m15s
cgrates.org,ONE_TIME_ACT,,,,,,,TOPUP_TEST_VOICE,,0s,*add_balance,,*balance.TestVoiceBalance2.Value,15m15s
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,FilterIDs,Weights,Blockers,AttributeFilterIDs,AttributeBlockers,Path,Type,Value
cgrates.org,ALS1,*string:~*req.Account:1001;*string:~*opts.*context:con1,;20,;true,,,,,
cgrates.org,ALS1,,,,*string:~*req.Field1:Initial,,*req.Field1,*variable,Sub1
cgrates.org,ALS1,*string:~*opts.*context:con2|con3,,,,,*req.Field2,*variable,Sub2
cgrates.org,ALS2,*string:~*opts.*context:con2|con3,,;false,,,*req.Field2,*variable,Sub2
cgrates.org,ALS2,*string:~*req.Account:1002;*string:~*opts.*context:con1,;20,,*string:~*req.Field1:Initial,,*req.Field1,*variable,Sub1
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Chargers.csv
	if err := writeFile(utils.ChargersCsv, `
#Tenant,ID,FilterIDs,Weights,Blockers,RunID,AttributeIDs
cgrates.org,Charger1,*string:~*req.Account:1001,;20,;true,,
cgrates.org,Charger1,,,,*rated,ATTR_1001_SIMPLEAUTH
cgrates.org,Charger2,,,,*rated,ATTR_1002_SIMPLEAUTH
cgrates.org,Charger2,*string:~*req.Account:1002,;15,;false,,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DispatcherProfiles.csv
	if err := writeFile(utils.DispatcherProfilesCsv, `
#Tenant,ID,FilterIDs,Weight,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters
cgrates.org,D1,,,,,,,,,
cgrates.org,D1,*string:~*req.Account:1001,20,*first,,,,,,
cgrates.org,D1,,,,,C1,fltr1,10,true,*ratio:1;param1:value1
cgrates.org,D1,,,,,,,,,
cgrates.org,D1,,,,,C1,fltr2;fltr4,,false,param2:value2
cgrates.org,D2,,,*first,,C3,fltr2,20,true,
cgrates.org,D2,*string:~*req.Account:1002,20,,,C2,fltr3,10,,param3:value3
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DispatcherHosts.csv
	if err := writeFile(utils.DispatcherHostsCsv, `
#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],ConnectTimeout[6],ReplyTimeout[7],Tls[8],ClientKey[9],ClientCertificate[10],CaCertificate[11]
cgrates.org,ALL,,,,,,,,,,
cgrates.org,ALL,127.0.0.1:6012,,1,3,1m,2m,true,,,
cgrates.org,ALL,,*json,1,3,1m,2m,false,,,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Filters.csv
	if err := writeFile(utils.FiltersCsv, `
#Tenant[0],ID[1],Type[2],Element[3],Values[4]
cgrates.org,FLTR_ACCOUNT_1001,,,
cgrates.org,FLTR_ACCOUNT_1001,*string,~*req.Account,1001;1002
cgrates.org,FLTR_ACCOUNT_1001,,,
cgrates.org,FLTR_ACCOUNT_1001,*prefix,~*req.Destination,10;20
cgrates.org,FLTR_ACCOUNT_1001,*rsr,~*req.Subject,~^1.*1$
cgrates.org,FLTR_ACCOUNT_1001,*rsr,~*req.Destination,1002
cgrates.org,FLTR_ACNT_dan,*string,~*req.Account,dan
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RateProfiles.csv
	if err := writeFile(utils.RatesCsv, `
#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,,,,,,,,,,,,,,,
cgrates.org,RP1,*string:~*req.Subject:1001,;0,0.1,0.6,*free,,,,,,,,,,
cgrates.org,RP1,,,,,,RT_WEEK,,"* * * * 1-5",;0,false,0s,0,0.12,1m,1m
cgrates.org,RP1,,,,,,RT_WEEK,,,,,1m,1.234,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_WEEKEND,,,,true,,0.067,0.03,,
cgrates.org,RP1,,,,,,RT_WEEKEND,,"* * * * 0,6",;10,false,0s,0.089,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,* * 24 12 *,;30,false,0s,0.0564,0.06,1m,1s
cgrates.org,RP1,,,,,,RT_CHRISTMAS,,,,true,,,,,
cgrates.org,RP2,,,,,,RT_WEEK,,,,,1m,1.234,0.06,1m,1s
cgrates.org,RP2,*string:~*req.Subject:1002,;10,0.2,0.4,*free,RT_WEEK,,"* * * * 1-5",fltr1;20,false,0s,0,0.24,2m,30s
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Resources.csv
	if err := writeFile(utils.ResourcesCsv, `
#Tenant[0],Id[1],FilterIDs[2],Weights[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],Thresholds[9]
cgrates.org,ResGroup21,*string:~*req.Account:1001,;10,1s,2,call,true,true,
cgrates.org,ResGroup21,,,,,,,,
cgrates.org,ResGroup22,,,,,,,,
cgrates.org,ResGroup22,*string:~*req.Account:dan,;10,3600s,2,premium_call,true,true,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Routes.csv
	if err := writeFile(utils.RoutesCsv, `
#Tenant[0],ID[1],FilterIDs[2],Weights[3],Blockers[4],Sorting[5],SortingParameters[6],RouteID[7],RouteFilterIDs[8],RouteAccountIDs[9],RouteRateProfileIDs[10],RouteResourceIDs[11],RouteStatIDs[12],RouteWeights[13],RouteBlockers[14],RouteParameters[15]
cgrates.org,RoutePrf1,,,,,,,,,,,,,,
cgrates.org,RoutePrf1,*string:~*req.Account:1001,;20,;true,*lc,,,,,,,,,,
cgrates.org,RoutePrf1,,,,,,route1,fltr1,Account1;Account2,RPL_1,ResGroup1,Stat1,,;true,param1
cgrates.org,RoutePrf1,,,,,,,,,,,,,,
cgrates.org,RoutePrf1,,,,,,route1,,,RPL_2,ResGroup2,,;10,,
cgrates.org,RoutePrf1,,,,,,route1,fltr2,,RPL_3,ResGroup3,Stat2,,,param2
cgrates.org,RoutePrf1,,,,,,route1,,,,ResGroup4,Stat3,,,
cgrates.org,RoutePrf1,,,,,,route2,fltr5,Account1,RPL_1,ResGroup1,Stat1,fltr1;10,;true,param1
cgrates.org,RoutePrf2,,,,,,,,,,,,,,
cgrates.org,RoutePrf2,*string:~*req.Account:1002,;20,,*lc,,route1,fltr3,Account3;Account4,RPL_2,ResGroup2,Stat2,;10,;true,param1
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Stats.csv
	if err := writeFile(utils.StatsCsv, `
#Tenant[0],ID[1],FilterIDs[2],Weights[3],Blockers[4],QueueLength[5],TTL[6],MinItems[7],Stored[8],ThresholdIDs[9],MetricIDs[10],MetricFilterIDs[11],MetricBlockers[12]
cgrates.org,TestStats,*string:~*req.Account:1001,;20,;true,100,1s,2,true,Th1;Th2,*sum#~*req.Value;*average#~*req.Value,fltr1;fltr2,
cgrates.org,TestStats,,,,,,2,,,*sum#~*req.Usage,,*string:~*req.Account:1003&fltr2;true;;false
cgrates.org,TestStats2,*string:~*req.Account:1002,,;true,100,1s,2,true,Th,*sum#~*req.Value;*sum#~*req.Usage;*average#~*req.Value;*average#~*req.Usage,,
cgrates.org,TestStats2,,;20,,,,2,true,,*sum#~*req.Cost;*average#~*req.Cost,,;false
cgrates.org,TestStats3,,,,,,,,,,,
cgrates.org,TestStats3,*string:~*req.Account:1003,;20,;true,100,1s,2,true,Th1;Th2,*sum#~*req.Value;*average#~*req.Value,,fltr_for_stats;false
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Thresholds.csv
	if err := writeFile(utils.ThresholdsCsv, `
#Tenant[0],Id[1],FilterIDs[2],Weights[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],ActionProfileIDs[8],Async[9]
cgrates.org,TH1,*string:~*req.Account:1001;*string:~*req.RunID:*default,;10,12,10,1s,true,ACT_PRF1,true
cgrates.org,TH1,,,,,,,,
cgrates.org,TH2,,,,,,,,
cgrates.org,TH2,*string:~*req.Account:1002,,,,,true,,true
cgrates.org,TH2,,;5,10,8,1s,true,ACT_PRF2,true
`); err != nil {
		t.Fatal(err)
	}
}

func testLoadersLoad(t *testing.T) {
	var reply string
	if err := ldrRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]interface{}{
				utils.MetaCache:       utils.MetaReload,
				utils.MetaStopOnError: true,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testLoadersGetAccounts(t *testing.T) {
	expAccs := []*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "1001",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: false,
				},
			},
			Opts: make(map[string]interface{}),
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Blockers: utils.Blockers{
						{
							FilterIDs: []string{"*string:~*req.Account:1002"},
							Blocker:   true,
						},
						{
							Blocker: false,
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
					Opts: map[string]interface{}{},
					CostIncrements: []*utils.CostIncrement{
						{
							FilterIDs:    []string{"fltr1", "fltr2"},
							Increment:    utils.NewDecimal(13, 1),
							FixedFee:     utils.NewDecimal(23, 1),
							RecurrentFee: utils.NewDecimal(33, 1),
						},
					},
					AttributeIDs: []string{"attr1", "attr2"},
				},
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Type:  utils.MetaVoice,
					Units: utils.NewDecimal(int64(time.Hour), 0),
					Opts:  make(map[string]interface{}),
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
		{
			Tenant: "cgrates.org",
			ID:     "1002",
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: false,
				},
			},
			Opts: make(map[string]interface{}),
			Balances: map[string]*utils.Balance{
				"MonetaryBalance": {
					ID: "MonetaryBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},

					Type:  utils.MetaMonetary,
					Units: utils.NewDecimal(int64(time.Hour), 0),
					Opts:  map[string]interface{}{},
				},
				"VoiceBalance": {
					ID: "VoiceBalance",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Type:  utils.MetaVoice,
					Units: utils.NewDecimal(14, 0),
					Opts:  make(map[string]interface{}),
					UnitFactors: []*utils.UnitFactor{
						{
							FilterIDs: []string{"fltr3", "fltr4"},
							Factor:    utils.NewDecimal(150, 0),
						},
						{
							FilterIDs: []string{"fltr5"},
							Factor:    utils.NewDecimal(250, 0),
						},
					},
					CostIncrements: []*utils.CostIncrement{
						{
							FilterIDs:    []string{"fltr3", "fltr4"},
							Increment:    utils.NewDecimal(13, 1),
							FixedFee:     utils.NewDecimal(23, 1),
							RecurrentFee: utils.NewDecimal(33, 1),
						},
					},
					AttributeIDs: []string{"attr3", "attr4"},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var accs []*utils.Account
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAccounts,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &accs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(accs, func(i, j int) bool {
			return accs[i].ID < accs[j].ID
		})
		if !reflect.DeepEqual(accs, expAccs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAccs), utils.ToJSON(accs))
		}
	}
}

func testLoadersGetActionProfiles(t *testing.T) {
	expActs := []*engine.ActionProfile{
		{
			Tenant: "cgrates.org",
			ID:     "ONE_TIME_ACT",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: true,
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
					Opts: map[string]interface{}{},
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
					Opts: map[string]interface{}{},
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
					Opts: map[string]interface{}{},
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
					Opts: map[string]interface{}{},
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
					Opts: map[string]interface{}{},
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
		},
	}
	var acts []*engine.ActionProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetActionProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &acts); err != nil {
		t.Error(err)
	} else {
		sort.Slice(acts, func(i, j int) bool {
			return acts[i].ID < acts[j].ID
		})
		if !reflect.DeepEqual(acts, expActs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expActs, acts)
		}
	}
}

func testLoadersGetAttributeProfiles(t *testing.T) {
	expAttrs := []*engine.APIAttributeProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "ALS1",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*opts.*context:con1", "*string:~*opts.*context:con2|con3"},
			Attributes: []*engine.ExternalAttribute{
				{
					FilterIDs: []string{"*string:~*req.Field1:Initial"},
					Path:      utils.MetaReq + utils.NestingSep + "Field1",
					Type:      utils.MetaVariable,
					Value:     "Sub1",
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Type:  utils.MetaVariable,
					Value: "Sub2",
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: true,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "ALS2",
			FilterIDs: []string{"*string:~*opts.*context:con2|con3", "*string:~*req.Account:1002", "*string:~*opts.*context:con1"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Type:  utils.MetaVariable,
					Value: "Sub2",
				},
				{
					FilterIDs: []string{"*string:~*req.Field1:Initial"},
					Path:      utils.MetaReq + utils.NestingSep + "Field1",
					Type:      utils.MetaVariable,
					Value:     "Sub1",
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var attrs []*engine.APIAttributeProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &attrs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(attrs, func(i, j int) bool {
			return attrs[i].ID < attrs[j].ID
		})
		if !reflect.DeepEqual(attrs, expAttrs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAttrs), utils.ToJSON(attrs))
		}
	}
}

func testLoadersGetChargerProfiles(t *testing.T) {
	expChrgs := []*engine.ChargerProfile{
		{
			Tenant:       "cgrates.org",
			ID:           "Charger1",
			FilterIDs:    []string{"*string:~*req.Account:1001"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1001_SIMPLEAUTH"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: true,
				},
			},
		},
		{
			Tenant:       "cgrates.org",
			ID:           "Charger2",
			FilterIDs:    []string{"*string:~*req.Account:1002"},
			RunID:        "*rated",
			AttributeIDs: []string{"ATTR_1002_SIMPLEAUTH"},
			Weights: utils.DynamicWeights{
				{
					Weight: 15,
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: false,
				},
			},
		},
	}
	var chrgs []*engine.ChargerProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetChargerProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &chrgs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(chrgs, func(i, j int) bool {
			return chrgs[i].ID < chrgs[j].ID
		})
		if !reflect.DeepEqual(chrgs, expChrgs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expChrgs), utils.ToJSON(chrgs))
		}
	}
}

func testLoadersGetDispatcherProfiles(t *testing.T) {
	expDspPrfs := []*engine.DispatcherProfile{
		{
			Tenant:         "cgrates.org",
			ID:             "D1",
			FilterIDs:      []string{"*string:~*req.Account:1001"},
			Strategy:       utils.MetaFirst,
			StrategyParams: map[string]interface{}{},
			Weight:         20,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:        "C1",
					FilterIDs: []string{"fltr1", "fltr2", "fltr4"},
					Weight:    10,
					Params: map[string]interface{}{
						utils.MetaRatio: "1",
						"param1":        "value1",
						"param2":        "value2",
					},
					Blocker: true,
				},
			},
		},
		{
			Tenant:         "cgrates.org",
			ID:             "D2",
			FilterIDs:      []string{"*string:~*req.Account:1002"},
			Strategy:       utils.MetaFirst,
			StrategyParams: map[string]interface{}{},
			Weight:         20,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:        "C3",
					FilterIDs: []string{"fltr2"},
					Weight:    20,
					Params:    map[string]interface{}{},
					Blocker:   true,
				},
				{
					ID:        "C2",
					FilterIDs: []string{"fltr3"},
					Weight:    10,
					Params: map[string]interface{}{
						"param3": "value3",
					},
					Blocker: false,
				},
			},
		},
	}
	var dspPrfs []*engine.DispatcherProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &dspPrfs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(dspPrfs, func(i, j int) bool {
			return dspPrfs[i].ID < dspPrfs[j].ID
		})
		if !reflect.DeepEqual(dspPrfs, expDspPrfs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expDspPrfs), utils.ToJSON(dspPrfs))
		}
	}
}

func testLoadersGetDispatcherHosts(t *testing.T) {
	expDspHosts := []*engine.DispatcherHost{
		{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "ALL",
				Address:         "127.0.0.1:6012",
				Transport:       utils.MetaJSON,
				ConnectAttempts: 1,
				Reconnects:      3,
				ConnectTimeout:  time.Minute,
				ReplyTimeout:    2 * time.Minute,
				TLS:             true,
			},
		},
	}
	var dspHosts []*engine.DispatcherHost
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHosts,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &dspHosts); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dspHosts, expDspHosts) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expDspHosts), utils.ToJSON(dspHosts))
	}
}

func testLoadersGetFilters(t *testing.T) {
	expFltrs := []*engine.Filter{
		{
			Tenant: "cgrates.org",
			ID:     "FLTR_ACCOUNT_1001",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001", "1002"},
				},
				{
					Element: "~*req.Destination",
					Type:    utils.MetaPrefix,
					Values:  []string{"10", "20"},
				},
				{
					Element: "~*req.Subject",
					Type:    utils.MetaRSR,
					Values:  []string{"~^1.*1$"},
				},
				{
					Element: "~*req.Destination",
					Type:    utils.MetaRSR,
					Values:  []string{"1002"},
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "FLTR_ACNT_dan",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"dan"},
				},
			},
		},
	}
	var fltrs []*engine.Filter
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetFilters,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &fltrs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(fltrs, func(i, j int) bool {
			return fltrs[i].ID < fltrs[j].ID
		})
		if !reflect.DeepEqual(fltrs, expFltrs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expFltrs), utils.ToJSON(fltrs))
		}
	}
}

func testLoadersGetRateProfiles(t *testing.T) {
	expRatePrfs := []*utils.RateProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "RP1",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			MinCost:         utils.NewDecimal(1, 1),
			MaxCost:         utils.NewDecimal(6, 1),
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: utils.DynamicWeights{
						{
							Weight: 0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
							FixedFee:      utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(12, 2),
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
					ID: "RT_WEEKEND",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					ActivationTimes: "* * * * 0,6",
					Blocker:         true,
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
							FixedFee:      utils.NewDecimal(89, 3),
							RecurrentFee:  utils.NewDecimal(6, 2),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
				"RT_CHRISTMAS": {
					ID: "RT_CHRISTMAS",
					Weights: utils.DynamicWeights{
						{
							Weight: 30,
						},
					},
					Blocker:         true,
					ActivationTimes: "* * 24 12 *",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
							FixedFee:      utils.NewDecimal(564, 4),
							RecurrentFee:  utils.NewDecimal(6, 2),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "RP2",
			FilterIDs: []string{"*string:~*req.Subject:1002"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			MinCost:         utils.NewDecimal(2, 1),
			MaxCost:         utils.NewDecimal(4, 1),
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: utils.DynamicWeights{
						{
							FilterIDs: []string{"fltr1"},
							Weight:    20,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
							FixedFee:      utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(24, 2),
							Unit:          utils.NewDecimal(int64(2*time.Minute), 0),
							Increment:     utils.NewDecimal(int64(30*time.Second), 0),
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
			},
		},
	}
	var ratePrfs []*utils.RateProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRateProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &ratePrfs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(ratePrfs, func(i, j int) bool {
			return ratePrfs[i].ID < ratePrfs[j].ID
		})
		if !reflect.DeepEqual(ratePrfs, expRatePrfs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expRatePrfs), utils.ToJSON(ratePrfs))
		}
	}
}

func testLoadersGetResourceProfiles(t *testing.T) {
	expRsPrfs := []*engine.ResourceProfile{
		{
			Tenant:            "cgrates.org",
			ID:                "ResGroup21",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			UsageTTL:          time.Second,
			AllocationMessage: "call",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit:   2,
			Blocker: true,
			Stored:  true,
		},
		{
			Tenant:            "cgrates.org",
			ID:                "ResGroup22",
			FilterIDs:         []string{"*string:~*req.Account:dan"},
			UsageTTL:          time.Hour,
			AllocationMessage: "premium_call",
			Blocker:           true,
			Stored:            true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			Limit: 2,
		},
	}
	var rsPrfs []*engine.ResourceProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetResourceProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rsPrfs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rsPrfs, func(i, j int) bool {
			return rsPrfs[i].ID < rsPrfs[j].ID
		})
		if !reflect.DeepEqual(rsPrfs, expRsPrfs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expRsPrfs), utils.ToJSON(rsPrfs))
		}
	}
}

func testLoadersGetRouteProfiles(t *testing.T) {
	expRouPrfs := []*engine.RouteProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "RoutePrf1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Sorting:   utils.MetaLC,
			Blockers: utils.Blockers{
				{
					Blocker: true,
				},
			},
			Routes: []*engine.Route{
				{
					ID:             "route1",
					FilterIDs:      []string{"fltr1", "fltr2"},
					AccountIDs:     []string{"Account1", "Account2"},
					RateProfileIDs: []string{"RPL_1", "RPL_2", "RPL_3"},
					ResourceIDs:    []string{"ResGroup1", "ResGroup2", "ResGroup3", "ResGroup4"},
					StatIDs:        []string{"Stat1", "Stat2", "Stat3"},
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Blockers: utils.Blockers{
						{
							Blocker: true,
						},
					},
					RouteParameters: "param2",
				},
				{
					ID:             "route2",
					FilterIDs:      []string{"fltr5"},
					AccountIDs:     []string{"Account1"},
					RateProfileIDs: []string{"RPL_1"},
					ResourceIDs:    []string{"ResGroup1"},
					StatIDs:        []string{"Stat1"},
					Weights: utils.DynamicWeights{
						{
							FilterIDs: []string{"fltr1"},
							Weight:    10,
						},
					},
					Blockers: utils.Blockers{
						{
							Blocker: true,
						},
					},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		{
			Tenant:    "cgrates.org",
			ID:        "RoutePrf2",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Sorting:   utils.MetaLC,
			Routes: []*engine.Route{
				{
					ID:             "route1",
					FilterIDs:      []string{"fltr3"},
					AccountIDs:     []string{"Account3", "Account4"},
					RateProfileIDs: []string{"RPL_2"},
					ResourceIDs:    []string{"ResGroup2"},
					StatIDs:        []string{"Stat2"},
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Blockers: utils.Blockers{
						{
							Blocker: true,
						},
					},
					RouteParameters: "param1",
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var rouPrfs []*engine.RouteProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rouPrfs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rouPrfs, func(i, j int) bool {
			return rouPrfs[i].ID < rouPrfs[j].ID
		})
		if !reflect.DeepEqual(rouPrfs, expRouPrfs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expRouPrfs), utils.ToJSON(rouPrfs))
		}
	}
}

func testLoadersGetStatQueueProfiles(t *testing.T) {
	expSqPrfs := []*engine.StatQueueProfile{
		{
			Tenant:      "cgrates.org",
			ID:          "TestStats",
			FilterIDs:   []string{"*string:~*req.Account:1001"},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: "*sum#~*req.Value",
				},
				{
					FilterIDs: []string{"fltr1", "fltr2"},
					MetricID:  "*average#~*req.Value",
				},
				{
					MetricID: "*sum#~*req.Usage",
					Blockers: utils.Blockers{
						{
							FilterIDs: []string{"*string:~*req.Account:1003", "fltr2"},
							Blocker:   true,
						},
						{
							Blocker: false,
						},
					},
				},
			},
			ThresholdIDs: []string{"Th1", "Th2"},
			Blockers:     utils.Blockers{{Blocker: true}},
			Stored:       true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 2,
		},
		{
			Tenant:      "cgrates.org",
			ID:          "TestStats2",
			FilterIDs:   []string{"*string:~*req.Account:1002"},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: "*sum#~*req.Value",
				},
				{
					MetricID: "*sum#~*req.Usage",
				},
				{
					MetricID: "*average#~*req.Value",
				},
				{
					MetricID: "*average#~*req.Usage",
				},
				{
					MetricID: "*sum#~*req.Cost",
				},
				{
					MetricID: "*average#~*req.Cost",
					Blockers: utils.Blockers{{Blocker: false}},
				},
			},
			ThresholdIDs: []string{"Th"},
			Blockers:     utils.Blockers{{Blocker: true}},
			Stored:       true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 2,
		},
		{
			Tenant:      "cgrates.org",
			ID:          "TestStats3",
			FilterIDs:   []string{"*string:~*req.Account:1003"},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: "*sum#~*req.Value",
				},
				{
					MetricID: "*average#~*req.Value",
					Blockers: utils.Blockers{
						{
							FilterIDs: []string{"fltr_for_stats"},
							Blocker:   false,
						},
					},
				},
			},
			ThresholdIDs: []string{"Th1", "Th2"},
			Blockers:     utils.Blockers{{Blocker: true}},
			Stored:       true,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinItems: 2,
		},
	}
	var sqPrfs []*engine.StatQueueProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &sqPrfs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(sqPrfs, func(i, j int) bool {
			return sqPrfs[i].ID < sqPrfs[j].ID
		})
		if !reflect.DeepEqual(sqPrfs, expSqPrfs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expSqPrfs), utils.ToJSON(sqPrfs))
		}
	}
}

func testLoadersGetThresholdProfiles(t *testing.T) {
	expThPrfs := []*engine.ThresholdProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "TH1",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.RunID:*default"},
			MaxHits:   12,
			MinHits:   10,
			MinSleep:  time.Second,
			Blocker:   true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			ActionProfileIDs: []string{"ACT_PRF1"},
			Async:            true,
		},
		{
			Tenant:    "cgrates.org",
			ID:        "TH2",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			MaxHits:   10,
			MinHits:   8,
			MinSleep:  time.Second,
			Blocker:   true,
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			ActionProfileIDs: []string{"ACT_PRF2"},
			Async:            true,
		},
	}
	var thPrfs []*engine.ThresholdProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &thPrfs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(thPrfs, func(i, j int) bool {
			return thPrfs[i].ID < thPrfs[j].ID
		})
		if !reflect.DeepEqual(thPrfs, expThPrfs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expThPrfs), utils.ToJSON(thPrfs))
		}
	}
}

func testLoadersRemove(t *testing.T) {
	var reply string
	if err := ldrRPC.Call(context.Background(), utils.LoaderSv1Run, //Remove,
		&loaders.ArgsProcessFolder{
			LoaderID: "remove",
			APIOpts: map[string]interface{}{
				utils.MetaCache:       utils.MetaReload,
				utils.MetaStopOnError: true,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testLoadersGetAccountAfterRemove(t *testing.T) {
	var accIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAccountIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &accIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyAccPrf utils.Account
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ACC_PRF",
		}, &rplyAccPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetActionProfileAfterRemove(t *testing.T) {
	var actIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &actIDs); err == nil || utils.ErrNotFound.Error() != err.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyActPrf engine.ActionProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ACT_PRF",
		}, &rplyActPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetAttributeProfileAfterRemove(t *testing.T) {
	var attrIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &attrIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyAttrPrf engine.AttributeProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ATTR_ACNT_1001",
		}, &rplyAttrPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetChargerProfileAfterRemove(t *testing.T) {
	var chgIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &chgIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyChgPrf engine.ChargerProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Raw",
		}, &rplyChgPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetDispatcherProfileAfterRemove(t *testing.T) {
	var dspPrfIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &dspPrfIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyDspPrf engine.ChargerProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "DSP1",
		}, &rplyDspPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetDispatcherHostAfterRemove(t *testing.T) {
	var dspHostIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHostIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &dspHostIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyDspHost engine.DispatcherHost
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHost,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "DSPHOST1",
		}, &rplyDspHost); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetFilterAfterRemove(t *testing.T) {
	var fltrIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &fltrIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyFltrPrf engine.Filter
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "FLTR_ACCOUNT_1001",
		}, &rplyFltrPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetRateProfileAfterRemove(t *testing.T) {
	var rateIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRateProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rateIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyRatePrf utils.RateProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RP1",
		}, &rplyRatePrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetResourceProfileAfterRemove(t *testing.T) {
	var rsIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetResourceProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rsIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyRsPrf engine.ResourceProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "RES_ACNT_1001",
		}, &rplyRsPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetRouteProfileAfterRemove(t *testing.T) {
	var rtIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rtIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyRtPrf engine.RouteProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ROUTE_ACNT_1001",
		}, &rplyRtPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetStatQueueProfileAfterRemove(t *testing.T) {
	var sqIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &sqIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplySqPrf engine.StatQueueProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "Stat_1",
		}, &rplySqPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testLoadersGetThresholdProfileAfterRemove(t *testing.T) {
	var thIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &thIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyThPrf engine.ThresholdProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1001",
		}, &rplyThPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestLoadersLoad(t *testing.T) {
	dirPath := "/tmp/TestLoadersLoad"
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dirPath)

	dirPathIn := "/tmp/TestLoadersLoad/in"
	err = os.Mkdir(dirPathIn, 0755)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dirPathIn)

	f1, err := os.Create(dirPathIn + "/emptyfile.txt")
	if err != nil {
		t.Error(err)
	}
	defer f1.Close()

	dirPathOut := "/tmp/TestLoadersLoad/out"
	err = os.Mkdir(dirPathOut, 0755)
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dirPathOut)

	f2, err := os.Create(dirPathOut + "/emptyfile.txt")
	if err != nil {
		t.Error(err)
	}
	defer f2.Close()

	cfg := config.NewDefaultCGRConfig()
	loaderCfg := &config.LoaderSCfg{
		ID:             "LoaderID",
		Enabled:        true,
		Tenant:         "cgrates.org",
		RunDelay:       1 * time.Millisecond,
		LockFilePath:   "lockFileName",
		FieldSeparator: ";",
		TpInDir:        dirPathIn,
		TpOutDir:       dirPathOut,
		Opts:           &config.LoaderSOptsCfg{},
	}

	cfg.LoaderCfg()[0] = loaderCfg
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	fltrs := engine.NewFilterS(cfg, nil, dm)
	ldrS := loaders.NewLoaderS(cfg, dm, fltrs, nil)
	lSv1 := NewLoaderSv1(ldrS)

	args := &loaders.ArgsProcessFolder{
		LoaderID: "LoaderID",
	}
	var reply string
	if err := lSv1.Run(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}
