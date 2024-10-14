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
	"github.com/cgrates/cgrates/utils"
)

var (
	expCfgDir  string
	expCfgPath string
	expCfg     *config.CGRConfig
	expRpc     *birpc.Client

	sTestsExp = []func(t *testing.T){
		testExpLoadConfig,
		testExpResetDataDB,
		testExpResetStorDb,
		testExpStartEngine,
		testExpRPCConn,
		testExpLoadTPFromFolder,
		testExpExportToFolder,
		testExpStopCgrEngine, // restart the engine and reset the database
		testExpResetDataDB,
		testExpResetStorDb,
		testExpStartEngine,
		testExpRPCConn,
		testExpLoadTPFromExported,
		testExpVerifyAttributes,
		testExpVerifyFilters,
		testExpVerifyThresholds,
		testExpVerifyResources,
		testExpVerifyStats,
		testExpVerifyRoutes,
		testExpCleanFiles,
		testExpStopCgrEngine,
	}
)

func TestExport(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		expCfgDir = "tutinternal"
	case utils.MetaMySQL:
		expCfgDir = "tutmysql"
	case utils.MetaMongo:
		expCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsExp {
		t.Run(expCfgDir, stest)
	}
}

func testExpLoadConfig(t *testing.T) {
	var err error
	expCfgPath = path.Join(*utils.DataDir, "conf", "samples", expCfgDir)
	if expCfg, err = config.NewCGRConfigFromPath(expCfgPath); err != nil {
		t.Error(err)
	}
}

func testExpResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(expCfg); err != nil {
		t.Fatal(err)
	}
}

func testExpResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(expCfg); err != nil {
		t.Fatal(err)
	}
}

func testExpStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(expCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testExpRPCConn(t *testing.T) {
	expRpc = engine.NewRPCClient(t, expCfg.ListenCfg())
}

func testExpLoadTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := expRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testExpExportToFolder(t *testing.T) {
	var reply string
	arg := &utils.ArgExportToFolder{
		Path: "/tmp/tp/",
	}
	if err := expRpc.Call(context.Background(), utils.APIerSv1ExportToFolder, arg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testExpLoadTPFromExported(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: "/tmp/tp/"}
	if err := expRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testExpVerifyAttributes(t *testing.T) {
	exp := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + "OfficeGroup",
				FilterIDs: []string{},
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("Marketing", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  10.0,
	}
	var reply *engine.AttributeProfile
	if err := expRpc.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
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
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	var reply *engine.Filter
	if err := expRpc.Call(context.Background(), utils.APIerSv1GetFilter,
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
			FilterIDs: []string{"FLTR_ACCOUNT_1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinHits:   0,
			MinSleep:  0,
			Blocker:   false,
			Weight:    10.0,
			ActionIDs: []string{"TOPUP_MONETARY_10"},
			Async:     false,
		},
	}
	var reply *engine.ThresholdProfile
	if err := expRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(tPrfl.ThresholdProfile), utils.ToJSON(reply))
	}
}

func testExpVerifyResources(t *testing.T) {
	rPrf := &engine.ResourceProfile{
		Tenant:       "cgrates.org",
		ID:           "RES_ACNT_1001",
		FilterIDs:    []string{"FLTR_ACCOUNT_1001"},
		UsageTTL:     time.Hour,
		Limit:        1,
		Blocker:      false,
		Stored:       false,
		Weight:       10,
		ThresholdIDs: []string{},
	}
	if *utils.Encoding == utils.MetaGOB {
		rPrf.ThresholdIDs = nil
	}
	var reply *engine.ResourceProfile
	if err := expRpc.Call(context.Background(), utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rPrf), utils.ToJSON(reply))
	}
}

func testExpVerifyStats(t *testing.T) {
	sPrf := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "Stat_1",
		FilterIDs: []string{"FLTR_STAT_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
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
		Blocker:      true,
		Stored:       false,
		Weight:       30,
		MinItems:     0,
		ThresholdIDs: []string{utils.MetaNone},
	}
	var reply *engine.StatQueueProfile
	if err := expRpc.Call(context.Background(), utils.APIerSv1GetStatQueueProfile,
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
				ID:              "route1",
				Weight:          20,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			{
				ID:              "route2",
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 10,
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

				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			{
				ID:              "route1",
				Weight:          20,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 10,
	}
	if err := expRpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
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

func testExpCleanFiles(t *testing.T) {
	if err := os.RemoveAll("/tmp/tp/"); err != nil {
		t.Error(err)
	}
}

func testExpStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
