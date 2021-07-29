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

package main

import (
	"bytes"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cgrLdrCfgPath string
	cgrLdrCfgDir  string
	cgrLdrCfg     *config.CGRConfig
	cgrLdrBIRPC   *birpc.Client
	cgrLdrTests   = []func(t *testing.T){
		testCgrLdrInitCfg,
		testCgrLdrInitDataDB,
		testCgrLdrInitStorDB,
		testCgrLdrStartEngine,
		testCgrLdrRPCConn,
		testCgrLdrGetSubsystemsNotLoadedLoad,
		testCgrLdrLoadData,
		testCgrLdrGetAccountAfterLoad,
		testCgrLdrGetActionProfileAfterLoad,
		testCgrLdrGetAttributeProfileAfterLoad,
		testCgrLdrGetFilterAfterLoad,
		testCgrLdrGetRateProfileAfterLoad,
		testCgrLdrGetResourceProfileAfterLoad,
		testCgrLdrGetResourceAfterLoad,
		testCgrLdrGetRouteProfileAfterLoad,
		testCgrLdrGetStatsProfileAfterLoad,
		testCgrLdrGetStatQueueAfterLoad,
		testCgrLdrGetThresholdProfileAfterLoad,
		testCgrLdrGetThresholdAfterLoad,

		//remove all data with cgr-loader and remove flag
		testCgrLdrRemoveData,
		testCgrLdrGetSubsystemsNotLoadedLoad,
		testCgrLdrKillEngine,
	}
)

func TestCGRLoaderRemove(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		cgrLdrCfgDir = "tutinternal"
	case utils.MetaMongo:
		cgrLdrCfgDir = "tutmongo"
	case utils.MetaMySQL:
		cgrLdrCfgDir = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, test := range cgrLdrTests {
		t.Run("cgr-loader remove tests", test)
	}
}

func testCgrLdrInitCfg(t *testing.T) {
	var err error
	cgrLdrCfgPath = path.Join(*dataDir, "conf", "samples", cgrLdrCfgDir)
	cgrLdrCfg, err = config.NewCGRConfigFromPath(cgrLdrCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCgrLdrInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(cgrLdrCfg); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrInitStorDB(t *testing.T) {
	if err := engine.InitStorDB(cgrLdrCfg); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cgrLdrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrRPCConn(t *testing.T) {
	var err error
	cgrLdrBIRPC, err = newRPCClient(cgrLdrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCgrLdrGetSubsystemsNotLoadedLoad(t *testing.T) {
	//account
	var replyAcc *utils.Account
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_PRF_1"}},
		&replyAcc); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//actionsPrf
	var replyAct *engine.ActionProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}},
		&replyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	//attributesPrf
	var replyAttr *engine.APIAttributeProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACNT_1001"}},
		&replyAttr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	//filtersPrf
	var replyFltr *engine.Filter
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}},
		&replyFltr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	//ratesPrf
	var replyRates *utils.RateProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RT_SPECIAL_1002"}},
		&replyRates); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %+q", utils.ErrNotFound.Error(), err.Error())
	}

	// resourcesPrf
	var replyResPrf *engine.ResourceProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyResPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}
	var replyRes *engine.Resource
	if err := cgrLdrBIRPC.Call(context.Background(), utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyRes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	// routesPrf
	var replyRts *engine.RouteProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}},
		&replyRts); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	// statsPrf
	var replySts *engine.StatQueueProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replySts); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	// statQueue
	var replyStQue *engine.StatQueue
	if err := cgrLdrBIRPC.Call(context.Background(), utils.StatSv1GetStatQueue,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replyStQue); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	// thresholdPrf
	var replyThdPrf *engine.ThresholdProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1"}},
		&replyThdPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	// threshold
	var rplyThd *engine.Threshold
	if err := cgrLdrBIRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1"}},
		&rplyThd); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+q, received %v", utils.ErrNotFound.Error(), err)
	}

	//chargers
	var replyChrgr *engine.ChargerProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Raw"},
		&replyChrgr); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %v", utils.ErrNotFound.Error(), err)
	}

}

func testCgrLdrLoadData(t *testing.T) {
	// *cacheSAddress = "127.0.0.1:2012"
	cmd := exec.Command("cgr-loader", "-config_path="+cgrLdrCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "testit"))
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

func testCgrLdrGetAccountAfterLoad(t *testing.T) {
	expAcc := &utils.Account{
		Tenant: "cgrates.org",
		ID:     "ACC_PRF_1",
		Weights: []*utils.DynamicWeight{
			{
				Weight: 20,
			},
		},
		FilterIDs: []string{},
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
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	var replyAcc *utils.Account
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_PRF_1"}},
		&replyAcc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyAcc, expAcc) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expAcc), utils.ToJSON(replyAcc))
	}
}

func testCgrLdrGetActionProfileAfterLoad(t *testing.T) {
	expActPrf := &engine.ActionProfile{
		Tenant:    "cgrates.org",
		ID:        "ONE_TIME_ACT",
		FilterIDs: []string{},
		Weight:    10,
		Schedule:  utils.MetaASAP,
		Targets: map[string]utils.StringSet{
			utils.MetaAccounts: {
				"1001": {},
				"1002": {},
			},
		},
		Actions: []*engine.APAction{
			{
				ID:   "TOPUP",
				Type: utils.MetaAddBalance,
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestBalance.Units",
						Value: "10",
					},
				},
			},
			{
				ID:   "SET_BALANCE_TEST_DATA",
				Type: utils.MetaSetBalance,
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestDataBalance.Type",
						Value: utils.MetaData,
					},
				},
			},
			{
				ID:   "TOPUP_TEST_DATA",
				Type: utils.MetaAddBalance,
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestDataBalance.Units",
						Value: "1024",
					},
				},
			},
			{
				ID:   "SET_BALANCE_TEST_VOICE",
				Type: utils.MetaSetBalance,
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestVoiceBalance.Type",
						Value: utils.MetaVoice,
					},
				},
			},
			{
				ID:   "TOPUP_TEST_VOICE",
				Type: utils.MetaAddBalance,
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestVoiceBalance.Units",
						Value: "15m15s",
					},
				},
			},
			{
				ID:   "SET_BALANCE_TEST_FILTERS",
				Type: utils.MetaSetBalance,
				Diktats: []*engine.APDiktat{
					{
						Path:  "*balance.TestVoiceBalance.Filters",
						Value: "*string:~*req.CustomField:500",
					},
				},
			},
			{
				ID:   "TOPUP_REM_VOICE",
				Type: utils.MetaRemBalance,
				Diktats: []*engine.APDiktat{
					{
						Path: "TestVoiceBalance2",
					},
				},
			},
		},
	}
	var replyAct *engine.ActionProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ONE_TIME_ACT"}},
		&replyAct); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyAct, expActPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expActPrf), utils.ToJSON(replyAct))
	}
}

func testCgrLdrGetAttributeProfileAfterLoad(t *testing.T) {
	extAttrPrf := &engine.APIAttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "ATTR_ACNT_1001",
		FilterIDs: []string{"*string:~*opts.*context:*sessions", "FLTR_ACCOUNT_1001"},
		Weight:    10,
		Attributes: []*engine.ExternalAttribute{
			{
				FilterIDs: []string{},
				Path:      "*req.OfficeGroup",
				Type:      utils.MetaConstant,
				Value:     "Marketing",
			},
		},
	}
	var replyAttr *engine.APIAttributeProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACNT_1001"}},
		&replyAttr); err != nil {
		t.Error(err)
	} else {
		sort.Strings(extAttrPrf.FilterIDs)
		sort.Strings(replyAttr.FilterIDs)
		if !reflect.DeepEqual(extAttrPrf, replyAttr) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(extAttrPrf), utils.ToJSON(replyAttr))
		}
	}
}

func testCgrLdrGetFilterAfterLoad(t *testing.T) {
	expFilter := &engine.Filter{
		Tenant: utils.CGRateSorg,
		ID:     "FLTR_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1003", "1002"},
			},
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Destination",
				Values:  []string{"10", "20"},
			},
			{
				Type:    utils.MetaRSR,
				Element: "~*req.Destination",
				Values:  []string{"1002"},
			},
		},
	}
	var replyFltr *engine.Filter
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}},
		&replyFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFilter, replyFltr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expFilter), utils.ToJSON(replyFltr))
	}
}

func testCgrLdrGetRateProfileAfterLoad(t *testing.T) {
	expRatePrf := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "RT_SPECIAL_1002",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Weights: []*utils.DynamicWeight{
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
				ActivationTimes: "* * * * *",
				Weights: []*utils.DynamicWeight{
					{
						Weight: 0,
					},
				},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						FixedFee:      utils.NewDecimal(0, 0),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
					},
				},
			},
		},
	}
	var replyRates *utils.RateProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RT_SPECIAL_1002"}},
		&replyRates); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyRates, expRatePrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRatePrf), utils.ToJSON(replyRates))
	}
}

func testCgrLdrGetResourceProfileAfterLoad(t *testing.T) {
	expREsPrf := &engine.ResourceProfile{
		Tenant:       utils.CGRateSorg,
		ID:           "RES_ACNT_1001",
		FilterIDs:    []string{"FLTR_ACCOUNT_1001"},
		Weight:       10,
		UsageTTL:     time.Hour,
		Limit:        1,
		ThresholdIDs: []string{},
	}
	var replyRes *engine.ResourceProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyRes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expREsPrf, replyRes) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expREsPrf), utils.ToJSON(replyRes))
	}
}

func testCgrLdrGetResourceAfterLoad(t *testing.T) {
	expREsPrf := &engine.Resource{
		Tenant: "cgrates.org",
		ID:     "RES_ACNT_1001",
		Usages: map[string]*engine.ResourceUsage{},
	}
	var replyRes *engine.Resource
	if err := cgrLdrBIRPC.Call(context.Background(), utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}},
		&replyRes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expREsPrf, replyRes) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expREsPrf), utils.ToJSON(replyRes))
	}
}

func testCgrLdrGetRouteProfileAfterLoad(t *testing.T) {
	expRoutePrf := &engine.RouteProfile{
		ID:                "ROUTE_ACNT_1001",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"FLTR_ACCOUNT_1001"},
		Weight:            10,
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:     "route1",
				Weight: 20,
			},
			{
				ID:     "route2",
				Weight: 10,
			},
		},
	}
	var replyRts *engine.RouteProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}},
		&replyRts); err != nil {
		t.Error(err)
	} else {
		sort.Slice(expRoutePrf.Routes, func(i, j int) bool {
			return expRoutePrf.Routes[i].ID < expRoutePrf.Routes[j].ID
		})
		sort.Slice(replyRts.Routes, func(i, j int) bool {
			return replyRts.Routes[i].ID < replyRts.Routes[j].ID
		})
		if !reflect.DeepEqual(expRoutePrf, replyRts) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRoutePrf), utils.ToJSON(replyRts))
		}
	}
}

func testCgrLdrGetStatsProfileAfterLoad(t *testing.T) {
	expStatsprf := &engine.StatQueueProfile{
		Tenant:      utils.CGRateSorg,
		ID:          "Stat_1",
		FilterIDs:   []string{"FLTR_STAT_1"},
		Weight:      30,
		QueueLength: 100,
		TTL:         10 * time.Second,
		MinItems:    0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: "*tcd",
			},
			{
				MetricID: "*asr",
			},
			{
				MetricID: "*acd",
			},
		},
		Blocker:      true,
		ThresholdIDs: []string{utils.MetaNone},
	}
	var replySts *engine.StatQueueProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replySts); err != nil {
		t.Error(err)
	} else {
		sort.Slice(expStatsprf.Metrics, func(i, j int) bool {
			return expStatsprf.Metrics[i].MetricID < expStatsprf.Metrics[j].MetricID
		})
		sort.Slice(replySts.Metrics, func(i, j int) bool {
			return replySts.Metrics[i].MetricID < replySts.Metrics[j].MetricID
		})
		if !reflect.DeepEqual(expStatsprf, replySts) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expStatsprf), utils.ToJSON(replySts))
		}
	}
}

func testCgrLdrGetStatQueueAfterLoad(t *testing.T) {
	expStatQueue := map[string]string{
		"*acd": "N/A",
		"*tcd": "N/A",
		"*asr": "N/A",
	}
	replyStQue := make(map[string]string)
	if err := cgrLdrBIRPC.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_1"}},
		&replyStQue); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStatQueue, replyStQue) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expStatQueue), utils.ToJSON(replyStQue))
	}
}

func testCgrLdrGetThresholdProfileAfterLoad(t *testing.T) {
	expThPrf := &engine.ThresholdProfile{
		Tenant:           utils.CGRateSorg,
		ID:               "THD_ACNT_1001",
		FilterIDs:        []string{"FLTR_ACCOUNT_1001"},
		Weight:           10,
		MaxHits:          -1,
		MinHits:          0,
		ActionProfileIDs: []string{"TOPUP_MONETARY_10"},
	}
	var replyThdPrf *engine.ThresholdProfile
	if err := cgrLdrBIRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}},
		&replyThdPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expThPrf, replyThdPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expThPrf), utils.ToJSON(replyThdPrf))
	}
}

func testCgrLdrGetThresholdAfterLoad(t *testing.T) {
	expThPrf := &engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ACNT_1001",
		Hits:   0,
	}
	var replyThdPrf *engine.Threshold
	if err := cgrLdrBIRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}},
		&replyThdPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expThPrf, replyThdPrf) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expThPrf), utils.ToJSON(replyThdPrf))
	}
}

func testCgrLdrRemoveData(t *testing.T) {
	// *cacheSAddress = "127.0.0.1:2012"
	cmd := exec.Command("cgr-loader", "-config_path="+cgrLdrCfgPath, "-path="+path.Join(*dataDir, "tariffplans", "testit"), "-remove")
	output := bytes.NewBuffer(nil)
	outerr := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = outerr
	if err := cmd.Run(); err != nil {
		t.Log(cmd.Args)
		t.Log(output.String())
		t.Log(outerr.String())
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testCgrLdrKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
