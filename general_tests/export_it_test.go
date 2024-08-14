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
package general_tests

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

var (
	expCfgDir  string
	replyBts   []byte
	expCfgPath string
	expCfg     *config.CGRConfig
	expRpc     *birpc.Client

	sTestsExp = []func(t *testing.T){
		testExpCreateFiles,
		testExpLoadConfig,
		testExpFlushDBs,
		testExpStartEngine,
		testExpRPCConn,
		testExpLoadTPFromFolder,
		//testExpExportToFolder,
		/* 	testExpCreatDirectoryWithTariffplan,
		testExpStopCgrEngine, // restart the engine and reset the database
		testExpResetDataDB,
		testExpStartEngineChangedLoderDirectory,
		testExpRPCConn,
		testExpLoadTPFromExported,
		testExpVerifyAttributes,
		testExpVerifyFilters,
		testExpVerifyThresholds,
		testExpVerifyResources,
		testExpVerifyStats,
		testExpVerifyRoutes,
		testExpVerifyRateProfiles,
		testExpVerifyActionProfiles,
		testExpVerifyAccounts, */
		testExpCleanFiles,
		testExpStopCgrEngine,
	}
)

func TestExport(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		expCfgDir = "export_it_test_internal"
	case utils.MetaMySQL:
		expCfgDir = "export_it_test_mysql"
	case utils.MetaMongo:
		expCfgDir = "export_it_test_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsExp {
		t.Run(expCfgDir, stest)
	}
}
func testExpCreateFiles(t *testing.T) {
	for _, dir := range eeSBlockerFiles {
		if err := os.RemoveAll("/tmp/archivesTP"); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll("/tmp/archivesTP", os.ModePerm); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testExpLoadConfig(t *testing.T) {
	expCfgPath = path.Join(*utils.DataDir, "conf", "samples", expCfgDir)
	if expCfg, err = config.NewCGRConfigFromPath(context.Background(), expCfgPath); err != nil {
		t.Error(err)
	}
}

func testExpFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(expCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(expCfg); err != nil {
		t.Fatal(err)
	}
}

func testExpStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(expCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testExpStartEngineChangedLoderDirectory(t *testing.T) {
	if _, err := engine.StopStartEngine(expCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testExpRPCConn(t *testing.T) {
	var err error
	expRpc, err = engine.NewRPCClient(expCfg.ListenCfg(), *utils.Encoding)
	if err != nil {
		t.Fatal(err)
	}
}

func testExpLoadTPFromFolder(t *testing.T) {
	caching := utils.MetaReload
	if expCfg.DataDbCfg().Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := expRpc.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaCache:       caching,
				utils.MetaStopOnError: true,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testExpExportToFolder(t *testing.T) {
	arg := &tpes.ArgsExportTP{
		Tenant:      "cgrates.org",
		ExportItems: map[string][]string{},
	}
	if err := expRpc.Call(context.Background(), utils.TPeSv1ExportTariffPlan, arg, &replyBts); err != nil {
		t.Error(err)
	}
}

func testExpCreatDirectoryWithTariffplan(t *testing.T) {
	filePath := filepath.Join("/tmp", "archivesTP")
	if err := os.RemoveAll(filePath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		t.Error(err)
	}

	rdr, err := zip.NewReader(bytes.NewReader(replyBts), int64(len(replyBts)))
	if err != nil {
		t.Error(err)
	}
	for _, f := range rdr.File {
		// 1
		newFilePath := filepath.Join(filePath, f.Name)
		w1, err := os.Create(newFilePath)
		if err != nil {
			t.Error(err)
		}
		defer w1.Close()
		wrtr, err := os.OpenFile(newFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			t.Error(err)
		}
		defer wrtr.Close()

		// 2
		f2, err := rdr.Open(f.Name)
		if err != nil {
			t.Error(err)
		}
		defer f2.Close()
		info := csv.NewReader(f2)
		csvFile, err := info.ReadAll()
		if err != nil {
			t.Error(err)
		}

		// 3
		csvWrtr := csv.NewWriter(wrtr)
		csvWrtr.Comma = utils.CSVSep
		if err := csvWrtr.WriteAll(csvFile); err != nil {
			t.Error(err)
		}
		csvWrtr.Flush()
	}
}

func testExpLoadTPFromExported(t *testing.T) {
	caching := utils.MetaReload
	if expCfg.DataDbCfg().Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := expRpc.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			LoaderID: "exported_ldr",
			APIOpts: map[string]any{
				utils.MetaCache: caching,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testExpVerifyAttributes(t *testing.T) {
	exp := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001", "*string:~*opts.*context:*sessions"},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + "OfficeGroup",
				FilterIDs: []string{},
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("Marketing", utils.InfieldSep),
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10.0,
			},
		},
	}
	var reply *engine.AttributeProfile
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACNT_1001"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if *utils.Encoding == utils.MetaGOB {
		for _, v := range exp.Attributes {
			v.FilterIDs = nil
		}
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testExpVerifyFilters(t *testing.T) {
	exp := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ACCOUNT_1001",
		Rules: []*engine.FilterRule{
			{
				Element: utils.MetaDynReq + utils.NestingSep + "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			},
		},
	}
	var reply *engine.Filter
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACCOUNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}

}

func testExpVerifyThresholds(t *testing.T) {
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_ACNT_1001",
			FilterIDs: []string{"FLTR_ACCOUNT_1001", "*ai:~*opts.*startTime:2014-07-29T15:00:00Z"},
			MaxHits:   -1,
			MinHits:   0,
			MinSleep:  0,
			Blocker:   false,
			Weights: utils.DynamicWeights{
				{
					Weight: 10.0,
				},
			},
			ActionProfileIDs: []string{"TOPUP_MONETARY_10"},
			Async:            false,
		},
	}
	var reply *engine.ThresholdProfile
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(tPrfl.ThresholdProfile), utils.ToJSON(reply))
	}
}

func testExpVerifyResources(t *testing.T) {
	rPrf := &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RES_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		UsageTTL:  time.Hour,
		Limit:     1,
		Blocker:   false,
		Stored:    false,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			}},
		ThresholdIDs: []string{},
	}
	if *utils.Encoding == utils.MetaGOB {
		rPrf.ThresholdIDs = nil
	}
	var reply *engine.ResourceProfile
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rPrf), utils.ToJSON(reply))
	}
}

func testExpVerifyStats(t *testing.T) {
	sPrf := &engine.StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "Stat_1",
		FilterIDs:   []string{"FLTR_STAT_1", "*ai:~*opts.*startTime:2014-07-29T15:00:00Z"},
		QueueLength: 100,
		TTL:         10 * time.Second,
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
		Blockers: utils.DynamicBlockers{{Blocker: true}},
		Stored:   false,
		Weights: utils.DynamicWeights{
			{
				Weight: 40,
			},
		},
		MinItems:     0,
		ThresholdIDs: []string{utils.MetaNone},
	}
	var reply *engine.StatQueueProfile
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}, &reply); err != nil {
		t.Error(err)
	}
	sort.Slice(reply.Metrics, func(i, j int) bool {
		return reply.Metrics[i].MetricID < reply.Metrics[j].MetricID
	})
	if !reflect.DeepEqual(sPrf, reply) {
		t.Errorf("Expecting: %+v \n  ,\n received: %+v",
			utils.ToJSON(sPrf), utils.ToJSON(reply))
	}
}

func testExpVerifyRoutes(t *testing.T) {
	var reply *engine.RouteProfile
	splPrf := &engine.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ROUTE_ACNT_1001",
		FilterIDs:         []string{"FLTR_ACCOUNT_1001"},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				RouteParameters: utils.EmptyString,
			},
			{
				ID: "route2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				RouteParameters: utils.EmptyString,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	splPrf2 := &engine.RouteProfile{
		Tenant:            "cgrates.org",
		ID:                "ROUTE_ACNT_1001",
		FilterIDs:         []string{"FLTR_ACCOUNT_1001"},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID: "route2",

				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				RouteParameters: utils.EmptyString,
			},
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 210,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				RouteParameters: utils.EmptyString,
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}, &reply); err != nil {
		t.Fatal(err)
	}
	if *utils.Encoding == utils.MetaGOB {
		splPrf.SortingParameters = nil
		splPrf2.SortingParameters = nil
	}
	if !reflect.DeepEqual(splPrf, reply) && !reflect.DeepEqual(splPrf2, reply) {
		t.Errorf("Expecting: %+v \n or %+v \n,\n received: %+v",
			utils.ToJSON(splPrf), utils.ToJSON(splPrf2), utils.ToJSON(reply))
	}
}

func testExpVerifyRateProfiles(t *testing.T) {
	var reply *utils.RateProfile
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}

	splPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RT_SPECIAL_1002",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		MinCost:         utils.NewDecimal(0, 0),
		MaxCost:         utils.NewDecimal(0, 0),
		MaxCostStrategy: utils.MetaMaxCostFree,
		Rates: map[string]*utils.Rate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				FilterIDs:       nil,
				ActivationTimes: "* * * * *",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Blocker: false,
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(int64(0*time.Second), 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
						FixedFee:      utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}

	if *utils.Encoding == utils.MetaGOB {
		splPrf.FilterIDs = nil
	}
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetRateProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RT_SPECIAL_1002"}}, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(splPrf, reply) {
		t.Errorf("Expecting: %+v,\n received: %+v",
			utils.ToJSON(splPrf), utils.ToJSON(reply))
	}
}

func testExpVerifyActionProfiles(t *testing.T) {
	var reply *engine.ActionProfile
	actPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ONE_TIME_ACT",
		FilterIDs: []string{},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Schedule: utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: {"1001": {}, "1002": {}},
		},
		Actions: []*engine.APAction{
			{
				ID:   "TOPUP",
				Type: utils.MetaAddBalance,
				Diktats: []*engine.APDiktat{{
					Path:  utils.MetaBalance + utils.NestingSep + "TestBalance" + utils.NestingSep + utils.Units,
					Value: "10",
				}},
			},

			{
				ID:   "SET_BALANCE_TEST_DATA",
				Type: utils.MetaSetBalance,
				Diktats: []*engine.APDiktat{{
					Path:  utils.MetaBalance + utils.NestingSep + "TestDataBalance" + utils.NestingSep + utils.Type,
					Value: utils.MetaData,
				}},
			},
			{
				ID:   "TOPUP_TEST_DATA",
				Type: utils.MetaAddBalance,
				Diktats: []*engine.APDiktat{{
					Path:  utils.MetaBalance + utils.NestingSep + "TestDataBalance" + utils.NestingSep + utils.Units,
					Value: "1024",
				}},
			},
			{
				ID:   "SET_BALANCE_TEST_VOICE",
				Type: utils.MetaSetBalance,
				Diktats: []*engine.APDiktat{{
					Path:  utils.MetaBalance + utils.NestingSep + "TestVoiceBalance" + utils.NestingSep + utils.Type,
					Value: utils.MetaVoice,
				}},
			},
			{
				ID:   "TOPUP_TEST_VOICE",
				Type: utils.MetaAddBalance,
				Diktats: []*engine.APDiktat{{
					Path:  utils.MetaBalance + utils.NestingSep + "TestVoiceBalance" + utils.NestingSep + utils.Units,
					Value: "15m15s",
				}},
			},
			{
				ID:   "SET_BALANCE_TEST_FILTERS",
				Type: utils.MetaSetBalance,
				Diktats: []*engine.APDiktat{{
					Path:  utils.MetaBalance + utils.NestingSep + "TestVoiceBalance" + utils.NestingSep + utils.Filters,
					Value: "*string:~*req.CustomField:500",
				}},
			},
			{
				ID:   "TOPUP_REM_VOICE",
				Type: utils.MetaRemBalance,
				Diktats: []*engine.APDiktat{{
					Path: "TestVoiceBalance2",
				}},
			},
		},
	}
	if *utils.Encoding == utils.MetaGOB {
		actPrf.FilterIDs = nil
		for _, act := range actPrf.Actions {
			act.FilterIDs = nil
		}
	}
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetActionProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(actPrf, reply) {
		t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(actPrf), utils.ToJSON(reply))
	}
}

func testExpVerifyAccounts(t *testing.T) {
	var reply *utils.Account
	acctPrf := &utils.Account{
		Tenant:    "cgrates.org",
		ID:        "ACC_PRF_1",
		FilterIDs: []string{},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Balances: map[string]*utils.Balance{
			"MonetaryBalance": {
				ID: "MonetaryBalance",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type: "*monetary",
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"fltr1", "fltr2"},
						Increment:    utils.NewDecimal(13, 1),
						FixedFee:     utils.NewDecimal(23, 1),
						RecurrentFee: utils.NewDecimal(33, 1),
					},
				},
				AttributeIDs: []string{"attr1", "attr2"},
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
				Units: utils.NewDecimal(14, 0),
			},
		},
		ThresholdIDs: []string{"*none"},
	}
	sort.Strings(acctPrf.Balances["MonetaryBalance"].CostIncrements[0].FilterIDs)
	sort.Strings(acctPrf.Balances["MonetaryBalance"].UnitFactors[0].FilterIDs)
	sort.Strings(acctPrf.Balances["MonetaryBalance"].AttributeIDs)
	if err := expRpc.Call(context.Background(), utils.AdminSv1GetAccount, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_PRF_1"}}, &reply); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(acctPrf.Balances["MonetaryBalance"].CostIncrements[0].FilterIDs)
		sort.Strings(acctPrf.Balances["MonetaryBalance"].UnitFactors[0].FilterIDs)
		sort.Strings(acctPrf.Balances["MonetaryBalance"].AttributeIDs)
		if !reflect.DeepEqual(acctPrf, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(acctPrf), utils.ToJSON(reply))
		}
	}
}

func testExpCleanFiles(t *testing.T) {
	if err := os.RemoveAll("/tmp/archivesTP"); err != nil {
		t.Error(err)
	}
}

func testExpStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
