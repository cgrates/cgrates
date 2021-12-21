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
		testLoadersResetStorDB,
		testLoadersStartEngine,
		testLoadersRPCConn,

		testLoadersWriteCSVs,
		testLoadersLoad,

		testLoadersGetAccount,
		testLoadersGetActionProfile,
		testLoadersGetAttributeProfile,
		testLoadersGetChargerProfile,
		testLoadersGetDispatcherProfile,
		testLoadersGetDispatcherHost,
		testLoadersGetFilter,
		testLoadersGetRateProfile,
		testLoadersGetResourceProfile,
		testLoadersGetRouteProfile,
		testLoadersGetStatQueueProfile,
		testLoadersGetThresholdProfile,

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

func testLoadersResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(ldrCfg); err != nil {
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
	if err := writeFile(utils.AccountsCsv, `#Tenant,ID,FilterIDs,Weights,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,ACC_PRF,,;20,,MonetaryBalance,,;10,*concrete,14,fltr1&fltr2;100;fltr3;200,,fltr1&fltr2;1.3;2.3;3.3,attr1;attr2,,*none`); err != nil {
		t.Fatal(err)
	}

	// Create and populate ActionProfiles.csv
	if err := writeFile(utils.ActionsCsv, `#Tenant,ID,FilterIDs,Weight,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionBlocker,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,ACT_PRF,,10,*asap,*accounts,1001,TOPUP,,false,0s,*add_balance,,*balance.TestBalance.Units,10`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `#Tenant,ID,FilterIDs,Weight,AttributeFilterIDs,Path,Type,Value,Blocker
cgrates.org,ATTR_ACNT_1001,FLTR_ACCOUNT_1001,10,,*req.OfficeGroup,*constant,Marketing,false`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Chargers.csv
	if err := writeFile(utils.ChargersCsv, `#Tenant,ID,FilterIDs,Weight,RunID,AttributeIDs
cgrates.org,Raw,FLTR_ACCOUNT_1001,20,*raw,*constant:*req.RequestType:*none`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DispatcherProfiles.csv
	if err := writeFile(utils.DispatcherProfilesCsv, `#Tenant,ID,FilterIDs,Weight,Strategy,StrategyParameters,ConnID,ConnFilterIDs,ConnWeight,ConnBlocker,ConnParameters
cgrates.org,DSP1,FLTR_ACCOUNT_1001,10,*weight,,ALL,,20,false,`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DispatcherHosts.csv
	if err := writeFile(utils.DispatcherHostsCsv, `#Tenant[0],ID[1],Address[2],Transport[3],ConnectAttempts[4],Reconnects[5],ConnectTimeout[6],ReplyTimeout[7],Tls[8],ClientKey[9],ClientCertificate[10],CaCertificate[11]
cgrates.org,DSPHOST1,*internal,,1,3,"1m","2m",false,,,`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Filters.csv
	if err := writeFile(utils.FiltersCsv, `#Tenant[0],ID[1],Type[2],Path[3],Values[4]
cgrates.org,FLTR_ACCOUNT_1001,*string,~*req.Account,1001`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RateProfiles.csv
	if err := writeFile(utils.RatesCsv, `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,FLTR_ACCOUNT_1001,;0,0.1,0.6,*free,RT_WEEK,FLTR_ACCOUNT_1001,"* * * * 1-5",;0,false,0s,,0.12,1m,1m`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Resources.csv
	if err := writeFile(utils.ResourcesCsv, `#Tenant[0],Id[1],FilterIDs[2],Weight[3],TTL[4],Limit[5],AllocationMessage[6],Blocker[7],Stored[8],ThresholdIDs[9]
cgrates.org,RES_ACNT_1001,FLTR_ACCOUNT_1001,10,1h,1,,false,false,`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Routes.csv
	if err := writeFile(utils.RoutesCsv, `#Tenant,ID,FilterIDs,Weights,Sorting,SortingParameters,RouteID,RouteFilterIDs,RouteAccountIDs,RouteRateProfileIDs,RouteResourceIDs,RouteStatIDs,RouteWeights,RouteBlocker,RouteParameters
cgrates.org,ROUTE_ACNT_1001,FLTR_ACCOUNT_1001,;10,*weight,,route1,,,,,,;20,,`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Stats.csv
	if err := writeFile(utils.StatsCsv, `#Tenant[0],Id[1],FilterIDs[2],Weight[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],ThresholdIDs[11]
cgrates.org,Stat_1,FLTR_ACCOUNT_1001,30,100,10s,0,*acd;*tcd;*asr,,false,true,*none`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Thresholds.csv
	if err := writeFile(utils.ThresholdsCsv, `#Tenant[0],Id[1],FilterIDs[2],Weight[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],ActionProfileIDs[8],Async[9]
cgrates.org,THD_ACNT_1001,FLTR_ACCOUNT_1001,10,-1,0,0,false,ACT_PRF,false`); err != nil {
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

func testLoadersGetAccount(t *testing.T) {
	expIDs := []string{"ACC_PRF"}
	var accIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAccountIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &accIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(accIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, accIDs)
	}

	expAccPrf := utils.Account{
		Tenant: "cgrates.org",
		ID:     "ACC_PRF",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Opts: make(map[string]interface{}),
		Balances: map[string]*utils.Balance{
			"MonetaryBalance": {
				ID:   "MonetaryBalance",
				Opts: make(map[string]interface{}),
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  "*concrete",
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
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	var rplyAccPrf utils.Account
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyAccPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyAccPrf, expAccPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expAccPrf), utils.ToJSON(rplyAccPrf))
	}
}

func testLoadersGetActionProfile(t *testing.T) {
	expIDs := []string{"ACT_PRF"}
	var actIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &actIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(actIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, actIDs)
	}

	expActPrf := engine.ActionProfile{
		Tenant:   "cgrates.org",
		ID:       "ACT_PRF",
		Schedule: utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: {
				"1001": {},
			},
		},
		Weight: 10,
		Actions: []*engine.APAction{
			{
				ID:   "TOPUP",
				Type: utils.MetaAddBalance,
				Opts: make(map[string]interface{}),
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestBalance.Units",
						Value: "10",
					},
				},
			},
		},
	}

	var rplyActPrf engine.ActionProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyActPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActPrf, expActPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expActPrf), utils.ToJSON(rplyActPrf))
	}
}

func testLoadersGetAttributeProfile(t *testing.T) {
	expIDs := []string{"ATTR_ACNT_1001"}
	var attrIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &attrIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attrIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, attrIDs)
	}

	expAttrPrf := engine.APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		Weight:    10,
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  "*req.OfficeGroup",
				Type:  utils.MetaConstant,
				Value: "Marketing",
			},
		},
	}

	var rplyAttrPrf engine.APIAttributeProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyAttrPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyAttrPrf, expAttrPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expAttrPrf), utils.ToJSON(rplyAttrPrf))
	}
}

func testLoadersGetChargerProfile(t *testing.T) {
	expIDs := []string{"Raw"}
	var chgIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetChargerProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &chgIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chgIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, chgIDs)
	}

	expChgPrf := engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Raw",
		FilterIDs:    []string{"FLTR_ACCOUNT_1001"},
		RunID:        utils.MetaRaw,
		AttributeIDs: []string{"*constant:*req.RequestType:*none"},
		Weight:       20,
	}

	var rplyChgPrf engine.ChargerProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyChgPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyChgPrf, expChgPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expChgPrf), utils.ToJSON(rplyChgPrf))
	}
}

func testLoadersGetDispatcherProfile(t *testing.T) {
	expIDs := []string{"DSP1"}
	var dspPrfIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &dspPrfIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dspPrfIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, dspPrfIDs)
	}

	expDspPrf := engine.DispatcherProfile{
		Tenant:         "cgrates.org",
		ID:             "DSP1",
		FilterIDs:      []string{"FLTR_ACCOUNT_1001"},
		Strategy:       utils.MetaWeight,
		StrategyParams: make(map[string]interface{}),
		Hosts: engine.DispatcherHostProfiles{
			{
				ID:     "ALL",
				Weight: 20,
				Params: make(map[string]interface{}),
			},
		},
		Weight: 10,
	}

	var rplyDspPrf engine.DispatcherProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyDspPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyDspPrf, expDspPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expDspPrf), utils.ToJSON(rplyDspPrf))
	}
}

func testLoadersGetDispatcherHost(t *testing.T) {
	expIDs := []string{"DSPHOST1"}
	var dspHostIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHostIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &dspHostIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dspHostIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, dspHostIDs)
	}

	expDspHost := engine.DispatcherHost{
		Tenant: "cgrates.org",
		RemoteHost: &config.RemoteHost{
			ID:              expIDs[0],
			Address:         utils.MetaInternal,
			ConnectAttempts: 1,
			Reconnects:      3,
			ConnectTimeout:  time.Minute,
			ReplyTimeout:    2 * time.Minute,
		},
	}

	var rplyDspHost engine.DispatcherHost
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetDispatcherHost,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyDspHost); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyDspHost, expDspHost) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expDspHost), utils.ToJSON(rplyDspHost))
	}
}

func testLoadersGetFilter(t *testing.T) {
	expIDs := []string{"FLTR_ACCOUNT_1001"}
	var fltrIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetFilterIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &fltrIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(fltrIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, fltrIDs)
	}

	expFltrPrf := engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ACCOUNT_1001",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
		},
	}

	var rplyFltrPrf engine.Filter
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyFltrPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFltrPrf, expFltrPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expFltrPrf), utils.ToJSON(rplyFltrPrf))
	}
}

func testLoadersGetRateProfile(t *testing.T) {
	expIDs := []string{"RP1"}
	var rateIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRateProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rateIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rateIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rateIDs)
	}

	expRatePrf := utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
		MaxCostStrategy: utils.MetaMaxCostFree,
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:        "RT_WEEK",
				FilterIDs: []string{"FLTR_ACCOUNT_1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          utils.NewDecimal(60000000000, 0),
						Increment:     utils.NewDecimal(60000000000, 0),
					},
				},
			},
		},
	}

	var rplyRatePrf utils.RateProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyRatePrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyRatePrf, expRatePrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expRatePrf), utils.ToJSON(rplyRatePrf))
	}
}

func testLoadersGetResourceProfile(t *testing.T) {
	expIDs := []string{"RES_ACNT_1001"}
	var rsIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetResourceProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rsIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rsIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rsIDs)
	}

	expRsPrf := engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		Weight:    10,
		UsageTTL:  3600000000000,
		Limit:     1,
	}

	var rplyRsPrf engine.ResourceProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyRsPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyRsPrf, expRsPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expRsPrf), utils.ToJSON(rplyRsPrf))
	}
}

func testLoadersGetRouteProfile(t *testing.T) {
	expIDs := []string{"ROUTE_ACNT_1001"}
	var rtIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rtIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, rtIDs)
	}

	expRtPrf := engine.APIRouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		Sorting:   utils.MetaWeight,
		Routes: []*engine.ExternalRoute{
			{
				ID:      "route1",
				Weights: ";20",
			},
		},
		Weights: ";10",
	}

	var rplyRtPrf engine.APIRouteProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyRtPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyRtPrf, expRtPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expRtPrf), utils.ToJSON(rplyRtPrf))
	}
}

func testLoadersGetStatQueueProfile(t *testing.T) {
	expIDs := []string{"Stat_1"}
	var sqIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &sqIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sqIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, sqIDs)
	}

	expSqPrf := engine.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "Stat_1",
		FilterIDs:   []string{"FLTR_ACCOUNT_1001"},
		QueueLength: 100,
		TTL:         10000000000,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaACD,
			},
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		Stored:       true,
		Weight:       30,
		ThresholdIDs: []string{utils.MetaNone},
	}

	var rplySqPrf engine.StatQueueProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplySqPrf); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rplySqPrf.Metrics, func(i, j int) bool { return rplySqPrf.Metrics[i].MetricID < rplySqPrf.Metrics[j].MetricID })
		if !reflect.DeepEqual(rplySqPrf, expSqPrf) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expSqPrf), utils.ToJSON(rplySqPrf))
		}
	}
}

func testLoadersGetThresholdProfile(t *testing.T) {
	expIDs := []string{"THD_ACNT_1001"}
	var thIDs []string
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &thIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thIDs, expIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, thIDs)
	}

	expThPrf := engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_ACNT_1001",
		FilterIDs:        []string{"FLTR_ACCOUNT_1001"},
		MaxHits:          -1,
		Weight:           10,
		ActionProfileIDs: []string{"ACT_PRF"},
	}

	var rplyThPrf engine.ThresholdProfile
	if err := ldrRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     expIDs[0],
		}, &rplyThPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyThPrf, expThPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expThPrf), utils.ToJSON(rplyThPrf))
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &attrIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyAttrPrf engine.APIAttributeProfile
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rtIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}

	var rplyRtPrf engine.APIRouteProfile
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
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
	ldrS := loaders.NewLoaderService(cfg, dm, fltrs, nil)
	lSv1 := NewLoaderSv1(ldrS)

	args := &loaders.ArgsProcessFolder{
		LoaderID: "LoaderID",
	}
	var reply string
	if err := lSv1.Run(context.Background(), args, &reply); err != nil {
		t.Error(err)
	}
}
