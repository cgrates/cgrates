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
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	expCfgDir  string
	expCfgPath string
	expCfg     *config.CGRConfig
	expRpc     *rpc.Client

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
	switch *dbType {
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
	expCfgPath = path.Join(*dataDir, "conf", "samples", expCfgDir)
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
	if _, err := engine.StopStartEngine(expCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testExpRPCConn(t *testing.T) {
	var err error
	expRpc, err = newRPCClient(expCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testExpLoadTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := expRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
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
	if err := expRpc.Call(utils.APIerSv1ExportToFolder, arg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testExpLoadTPFromExported(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: "/tmp/tp/"}
	if err := expRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testExpVerifyAttributes(t *testing.T) {
	exp := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1003_SESSIONAUTH",
		FilterIDs: []string{"*string:~*req.Account:1003"},
		Contexts:  []string{utils.MetaSessionS},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				FilterIDs: []string{},
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", utils.INFIELD_SEP),
			},
			{
				Path:      utils.MetaReq + utils.NestingSep + utils.RequestType,
				FilterIDs: []string{},
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("*prepaid", utils.INFIELD_SEP),
			},
			{
				Path:      utils.MetaReq + utils.NestingSep + "PaypalAccount",
				FilterIDs: []string{},
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("cgrates@paypal.com", utils.INFIELD_SEP),
			},
			{
				Path:      utils.MetaReq + utils.NestingSep + "LCRProfile",
				FilterIDs: []string{},
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("premium_cli", utils.INFIELD_SEP),
			},
		},
		Weight: 10.0,
	}
	var reply *engine.AttributeProfile
	if err := expRpc.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1003_SESSIONAUTH"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if *encoding == utils.MetaGOB {
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
		ID:     "FLTR_ACNT_1001_1002",
		Rules: []*engine.FilterRule{
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.RunID,
				Type:    utils.MetaString,
				Values:  []string{utils.MetaDefault},
			},
			{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destination,
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002", "1003"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	var reply *engine.Filter
	if err := expRpc.Call(utils.APIerSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACNT_1001_1002"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}

}

func testExpVerifyThresholds(t *testing.T) {
	tPrfl := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_ACNT_1001",
			FilterIDs: []string{"FLTR_ACNT_1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			MaxHits:   1,
			MinHits:   1,
			MinSleep:  time.Second,
			Blocker:   false,
			Weight:    10.0,
			ActionIDs: []string{"ACT_LOG_WARNING"},
			Async:     true,
		},
	}
	var reply *engine.ThresholdProfile
	if err := expRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(tPrfl.ThresholdProfile), utils.ToJSON(reply))
	}
}

func testExpVerifyResources(t *testing.T) {
	rPrf := &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_RES"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		UsageTTL:     -1,
		Limit:        7,
		Blocker:      false,
		Stored:       true,
		Weight:       10,
		ThresholdIDs: []string{utils.META_NONE},
	}
	var reply *engine.ResourceProfile
	if err := expRpc.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rPrf), utils.ToJSON(reply))
	}
}

func testExpVerifyStats(t *testing.T) {
	sPrf := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "Stats2",
		FilterIDs: []string{"FLTR_ACNT_1001_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		QueueLength: 100,
		TTL:         -1,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		Blocker:      true,
		Stored:       false,
		Weight:       30,
		MinItems:     0,
		ThresholdIDs: []string{utils.META_NONE},
	}

	sPrf2 := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "Stats2",
		FilterIDs: []string{"FLTR_ACNT_1001_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		QueueLength: 100,
		TTL:         -1,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaTCC,
			},
		},
		Blocker:      true,
		Stored:       false,
		Weight:       30,
		MinItems:     0,
		ThresholdIDs: []string{utils.META_NONE},
	}

	var reply *engine.StatQueueProfile
	if err := expRpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(sPrf, reply) && !reflect.DeepEqual(sPrf2, reply) {
		t.Errorf("Expecting: %+v \n or %+v \n ,\n received: %+v",
			utils.ToJSON(sPrf), utils.ToJSON(sPrf2), utils.ToJSON(reply))
	}
}

func testExpVerifyRoutes(t *testing.T) {
	var reply *engine.RouteProfile
	splPrf := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1002",
		FilterIDs: []string{"FLTR_ACNT_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:              "route1",
				RatingPlanIDs:   []string{"RP_1002_LOW"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			{
				ID:              "route2",
				RatingPlanIDs:   []string{"RP_1002"},
				Weight:          20,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 10,
	}

	splPrf2 := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1002",
		FilterIDs: []string{"FLTR_ACNT_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:              "route2",
				RatingPlanIDs:   []string{"RP_1002"},
				Weight:          20,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
			{
				ID:              "route1",
				RatingPlanIDs:   []string{"RP_1002_LOW"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: utils.EmptyString,
			},
		},
		Weight: 10,
	}
	if err := expRpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1002"}, &reply); err != nil {
		t.Fatal(err)
	}
	if *encoding == utils.MetaGOB {
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
