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
package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	internalCfgPath    string
	internalCfgDirPath string
	internalCfg        *config.CGRConfig
	internalRPC        *rpc.Client

	engineOneCfgPath    string
	engineOneCfgDirPath string
	engineOneCfg        *config.CGRConfig
	engineOneRPC        *rpc.Client

	engineTwoCfgPath    string
	engineTwoCfgDirPath string
	engineTwoCfg        *config.CGRConfig
	engineTwoRPC        *rpc.Client

	sTestsInternalRemoteIT = []func(t *testing.T){
		testInternalRemoteITInitCfg,
		testInternalRemoteITDataFlush,
		testInternalRemoteITStartEngine,
		testInternalRemoteITRPCConn,
		testInternalRemoteLoadDataInEngineTwo,

		testInternalRemoteITGetAccount,
		testInternalRemoteITGetAttribute,
		testInternalRemoteITGetThreshold,
		testInternalRemoteITGetThresholdProfile,
		testInternalRemoteITGetResource,
		testInternalRemoteITGetResourceProfile,
		testInternalRemoteITGetStatQueueProfile,
		testInternalRemoteITGetRoute,
		testInternalRemoteITGetFilter,
		testInternalRemoteITGetRatingPlan,
		testInternalRemoteITGetRatingProfile,
		testInternalRemoteITGetAction,
		testInternalRemoteITGetActionPlan,
		testInternalRemoteITGetDestination,
		testInternalRemoteITGetReverseDestination,
		testInternalRemoteITGetChargerProfile,
		testInternalRemoteITGetDispatcherProfile,
		testInternalRemoteITGetDispatcherHost,

		testInternalReplicationSetThreshold,
		testInternalMatchThreshold,
		testInternalAccountBalanceOperations,
		testInternalSetAccount,
		testInternalReplicateStats,

		testInternalRemoteITKillEngine,
	}
)

func TestInternalRemoteIT(t *testing.T) {
	internalCfgDirPath = "internal"
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		engineOneCfgDirPath = "engine1_redis"
		engineTwoCfgDirPath = "engine2_redis"
	case utils.MetaMongo:
		engineOneCfgDirPath = "engine1_mongo"
		engineTwoCfgDirPath = "engine2_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *encoding == utils.MetaGOB {
		internalCfgDirPath += "_gob"
	}
	for _, stest := range sTestsInternalRemoteIT {
		t.Run(*dbType, stest)
	}
}

func testInternalRemoteITInitCfg(t *testing.T) {
	var err error
	internalCfgPath = path.Join(*dataDir, "conf", "samples", "remote_replication", internalCfgDirPath)
	internalCfg, err = config.NewCGRConfigFromPath(internalCfgPath)
	if err != nil {
		t.Error(err)
	}

	// prepare config for engine1
	engineOneCfgPath = path.Join(*dataDir, "conf", "samples",
		"remote_replication", engineOneCfgDirPath)
	engineOneCfg, err = config.NewCGRConfigFromPath(engineOneCfgPath)
	if err != nil {
		t.Error(err)
	}
	engineOneCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()

	// prepare config for engine2
	engineTwoCfgPath = path.Join(*dataDir, "conf", "samples",
		"remote_replication", engineTwoCfgDirPath)
	engineTwoCfg, err = config.NewCGRConfigFromPath(engineTwoCfgPath)
	if err != nil {
		t.Error(err)
	}
	engineTwoCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()

}

func testInternalRemoteITDataFlush(t *testing.T) {
	if err := engine.InitDataDb(engineOneCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(engineTwoCfg); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testInternalRemoteITStartEngine(t *testing.T) {
	engine.KillEngine(100)
	if _, err := engine.StartEngine(engineOneCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(engineTwoCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(internalCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testInternalRemoteITRPCConn(t *testing.T) {
	var err error
	engineOneRPC, err = newRPCClient(engineOneCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	engineTwoRPC, err = newRPCClient(engineTwoCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	internalRPC, err = newRPCClient(internalCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testInternalRemoteLoadDataInEngineTwo(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := engineTwoRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testInternalRemoteITGetAccount(t *testing.T) {
	var acnt *engine.Account
	expAcc := &engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MetaMonetary: []*engine.Balance{
				{
					ID:     "testAccount",
					Value:  10,
					Weight: 10,
				},
			},
		},
	}
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	// make sure account exist in engine2
	if err := engineTwoRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.ID != expAcc.ID {
		t.Errorf("expecting: %+v, received: %+v", expAcc.ID, acnt.ID)
	} else if len(acnt.BalanceMap) != 1 {
		t.Errorf("unexpected number of balances received: %+v", utils.ToJSON(acnt))
	}
	// check the account in internal
	if err := internalRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.ID != expAcc.ID {
		t.Errorf("expecting: %+v, received: %+v", expAcc.ID, acnt.ID)
	} else if len(acnt.BalanceMap) != 1 {
		t.Errorf("unexpected number of balances received: %+v", utils.ToJSON(acnt))
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "nonexistAccount",
	}
	if err := internalRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expecting: %+v, received: %+v", utils.ErrNotFound, err)
	}
}

func testInternalRemoteITGetAttribute(t *testing.T) {
	alsPrf = &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_1001_SIMPLEAUTH",
			Contexts:  []string{"simpleauth"},
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Attributes: []*engine.Attribute{{
				Path:  utils.MetaReq + utils.NestingSep + "Password",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			}},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := internalRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}}, &reply); err != nil {
		t.Fatal(err)
	}
	if *encoding == utils.MetaGOB { // in gob empty slice is encoded as nil
		alsPrf.AttributeProfile.Attributes[0].FilterIDs = nil
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetThreshold(t *testing.T) {
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}
	if err := internalRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testInternalRemoteITGetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
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
			Weight:    10.0,
			ActionIDs: []string{"ACT_LOG_WARNING"},
			Async:     true,
		},
	}
	if err := internalRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(tPrfl.ThresholdProfile), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetResource(t *testing.T) {
	var reply *engine.Resource
	expectedResources := &engine.Resource{
		Tenant: "cgrates.org",
		ID:     "ResGroup1",
		Usages: map[string]*engine.ResourceUsage{},
	}
	if err := internalRPC.Call(utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expectedResources) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedResources), utils.ToJSON(reply))
	}

	if err := internalRPC.Call(utils.ResourceSv1GetResource,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "ResGroup1"}}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expectedResources) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedResources), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetResourceProfile(t *testing.T) {
	rlsPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResGroup1",
			FilterIDs: []string{"FLTR_RES"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			UsageTTL:          -1,
			Limit:             7,
			AllocationMessage: "",
			Stored:            true,
			Weight:            10,
			ThresholdIDs:      []string{utils.MetaNone},
		},
	}
	var reply *engine.ResourceProfile
	if err := internalRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsPrf.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsPrf.ResourceProfile) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsPrf.ResourceProfile), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetStatQueueProfile(t *testing.T) {
	expStq := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "Stats2",
		FilterIDs: []string{"FLTR_ACNT_1001_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		QueueLength: 100,
		TTL:         -1,
		MinItems:    0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		Stored:       false,
		Blocker:      true,
		Weight:       30,
		ThresholdIDs: []string{utils.MetaNone},
	}
	//reverse metric order
	expStq2 := &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "Stats2",
		FilterIDs: []string{"FLTR_ACNT_1001_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		QueueLength: 100,
		TTL:         -1,
		MinItems:    0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaTCC,
			},
		},
		Stored:       false,
		Blocker:      true,
		Weight:       30,
		ThresholdIDs: []string{utils.MetaNone},
	}
	var reply *engine.StatQueueProfile
	if err := internalRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStq, reply) && !reflect.DeepEqual(reply, expStq2) {
		t.Errorf("Expecting: %+v or %+v, received: %+v", utils.ToJSON(expStq),
			utils.ToJSON(expStq2), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetRoute(t *testing.T) {
	var reply *engine.RouteProfile
	routePrf := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1001",
		FilterIDs: []string{"FLTR_ACNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting: utils.MetaWeight,
		Routes: []*engine.Route{
			{
				ID:     "route1",
				Weight: 10,
			},
			{
				ID:     "route2",
				Weight: 20,
			},
		},
		Weight: 20,
	}
	// routeProfile in reverse order
	routePrf2 := &engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1001",
		FilterIDs: []string{"FLTR_ACNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting: utils.MetaWeight,
		Routes: []*engine.Route{{
			ID:     "route2",
			Weight: 20,
		}, {
			ID:     "route1",
			Weight: 10,
		}},
		Weight: 20,
	}
	if *encoding == utils.MetaGOB { // in gob emtpty slice is encoded as nil
		routePrf.SortingParameters = nil
		routePrf2.SortingParameters = nil
	}

	if err := internalRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(routePrf, reply) && !reflect.DeepEqual(routePrf2, reply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(routePrf), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetFilter(t *testing.T) {
	expFltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ACNT_1001",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Values:  []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	var reply *engine.Filter
	if err := internalRPC.Call(utils.APIerSv1GetFilter,
		&utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFltr, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expFltr), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetRatingPlan(t *testing.T) {
	var reply engine.RatingPlan
	if err := internalRPC.Call(utils.APIerSv1GetRatingPlan, utils.StringPointer("RP_1001"), &reply); err != nil {
		t.Error(err.Error())
	} else if reply.Id != "RP_1001" {
		t.Errorf("Expected: %+v, received: %+v", "RP_1001", reply.Id)
	}
}

func testInternalRemoteITGetRatingProfile(t *testing.T) {
	var rpl engine.RatingProfile
	attrGetRatingPlan := &utils.AttrGetRatingProfile{
		Tenant: "cgrates.org", Category: "call", Subject: "1001"}
	actTime, err := utils.ParseTimeDetectLayout("2014-01-14T00:00:00Z", "")
	if err != nil {
		t.Error(err)
	}
	actTime1, err := utils.ParseTimeDetectLayout("2010-01-14T00:00:00Z", "")
	if err != nil {
		t.Error(err)
	}
	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:call:1001",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RP_1001",
			},
			{
				ActivationTime: actTime1,
				RatingPlanId:   "RP_1002",
			},
		},
	}
	if err := internalRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingPlan, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}
}

func testInternalRemoteITGetAction(t *testing.T) {
	expectActs := []*utils.TPAction{
		{Identifier: utils.MetaTopUpReset, BalanceId: "test", BalanceType: utils.MetaMonetary,
			Units: "10", BalanceWeight: "10", BalanceBlocker: "false",
			BalanceDisabled: "false", ExpiryTime: utils.MetaUnlimited, Weight: 10.0}}

	var reply []*utils.TPAction
	if err := internalRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACT_TOPUP_RST_10"), &reply); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(expectActs, reply) {
		t.Errorf("Expected: %v,\n received: %v", utils.ToJSON(expectActs), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetActionPlan(t *testing.T) {
	var aps []*engine.ActionPlan
	if err := internalRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "AP_PACKAGE_10"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "AP_PACKAGE_10" {
		t.Errorf("Expected: %v,\n received: %v", "AP_PACKAGE_10", aps[0].Id)
	}
	if err := internalRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: utils.EmptyString}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "AP_PACKAGE_10" {
		t.Errorf("Expected: %v,\n received: %v", "AP_PACKAGE_10", aps[0].Id)
	}
}

func testInternalRemoteITGetDestination(t *testing.T) {
	var dst *engine.Destination
	eDst := &engine.Destination{
		Id:       "DST_1002",
		Prefixes: []string{"1002"},
	}
	if err := internalRPC.Call(utils.APIerSv1GetDestination, utils.StringPointer("DST_1002"), &dst); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDst, dst) {
		t.Errorf("Expected: %v,\n received: %v", eDst, dst)
	}
}

func testInternalRemoteITGetReverseDestination(t *testing.T) {
	var ids []string
	eIDs := []string{"DST_1002"}
	if err := internalRPC.Call(utils.APIerSv1GetReverseDestination, utils.StringPointer("1002"), &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIDs, ids) {
		t.Errorf("Expected: %v,\n received: %v", eIDs, ids)
	}
}

func testInternalRemoteITGetChargerProfile(t *testing.T) {
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "DEFAULT",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{utils.MetaNone},
		Weight:       0,
	}
	if *encoding == utils.MetaGOB {
		chargerProfile.FilterIDs = nil
	}
	var reply *engine.ChargerProfile
	if err := internalRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DEFAULT"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(chargerProfile), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetDispatcherProfile(t *testing.T) {
	var reply string
	if err := internalRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp1",
			FilterIDs:  []string{"*string:~*req.Account:1001"},
			Subsystems: []string{utils.MetaAny},
			Strategy:   utils.MetaFirst,
			Weight:     20,
		},
	}

	if err := engineTwoRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	var dsp *engine.DispatcherProfile
	if err := internalRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherProfile.DispatcherProfile, dsp)
	}
}

func testInternalRemoteITGetDispatcherHost(t *testing.T) {
	var reply string
	if err := internalRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	dispatcherHost = &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:      "DspHst1",
				Address: "*internal",
			},
		},
	}

	if err := engineTwoRPC.Call(utils.APIerSv1SetDispatcherHost,
		dispatcherHost,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	var dsp *engine.DispatcherHost
	if err := internalRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherHost.DispatcherHost, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherHost.DispatcherHost, dsp)
	}
}

func testInternalReplicationSetThreshold(t *testing.T) {
	var reply *engine.ThresholdProfile
	var result string
	if err := internalRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Replication"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//verify indexes on engine2 before adding new threshold profile
	var indexes []string
	expectedIDX := []string{"*string:*req.Account:1001:THD_ACNT_1001",
		"*string:*req.Account:1002:THD_ACNT_1002"}
	if err := engineTwoRPC.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: "cgrates.org", FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	sort.Strings(indexes)
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
	//verify indexes on internal before adding new threshold profile
	if err := internalRPC.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: "cgrates.org", FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	sort.Strings(indexes)
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}

	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_Replication",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.CustomField:CustomValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_LOG_WARNING"},
			Async:     true,
		},
	}
	if err := internalRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	time.Sleep(50 * time.Millisecond)
	if err := internalRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Replication"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	expectedIDX = []string{
		"*string:*req.Account:1001:THD_ACNT_1001",
		"*string:*req.Account:1001:THD_Replication",
		"*string:*req.Account:1002:THD_ACNT_1002",
		"*string:*req.CustomField:CustomValue:THD_Replication",
	}
	// verify index on internal
	sort.Strings(expectedIDX)
	if err := internalRPC.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: "cgrates.org", FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	sort.Strings(indexes)
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
	// verify data on engine1
	if err := engineOneRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Replication"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	expectedIDX2 := []string{
		"*string:*req.Account:1001:THD_ACNT_1001",
		"*string:*req.Account:1001:THD_Replication",
		"*string:*req.CustomField:CustomValue:THD_Replication",
	}
	// verify indexes on engine1 (should be the same as internal)
	if err := engineOneRPC.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: "cgrates.org", FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	sort.Strings(indexes)
	if !reflect.DeepEqual(expectedIDX2, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX2, utils.ToJSON(indexes))
	}
	// verify data on engine2
	if err := engineTwoRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Replication"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	expectedIDX = []string{"*string:*req.Account:1001:THD_ACNT_1001",
		"*string:*req.Account:1001:THD_Replication",
		"*string:*req.Account:1002:THD_ACNT_1002",
		"*string:*req.CustomField:CustomValue:THD_Replication",
	}
	// check if indexes was created correctly on engine2
	if err := engineTwoRPC.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: "cgrates.org", FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	sort.Strings(indexes)
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v,\n received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}

}

func testInternalMatchThreshold(t *testing.T) {
	ev := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
	}
	var ids []string
	eIDs := []string{"THD_ACNT_1002"}
	if err := internalRPC.Call(utils.ThresholdSv1ProcessEvent, ev, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	ev = &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	eIDs = []string{"THD_ACNT_1001"}
	if err := internalRPC.Call(utils.ThresholdSv1ProcessEvent, ev, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	ev2 := &engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	if err := internalRPC.Call(utils.ThresholdSv1ProcessEvent, ev2, &ids); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testInternalAccountBalanceOperations(t *testing.T) {
	var reply string
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccount1",
		BalanceType: utils.MetaMonetary,
		Value:       17.4,
		Balance: map[string]interface{}{
			utils.ID: "testAccSetBalance",
		},
	}
	if err := internalRPC.Call(utils.APIerSv1SetBalance, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testAccount1",
	}
	// verify account on engineOne
	if err := engineOneRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 17.4 {
		t.Errorf("Expecting: %+v, received: %+v",
			17.4, val)
	}

	// verify account on engineTwo
	if err := engineTwoRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 17.4 {
		t.Errorf("Expecting: %+v, received: %+v",
			17.4, val)
	}

	// debit balance on internal and the account should be replicated to other engines
	if err := internalRPC.Call(utils.APIerSv1DebitBalance, &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccount1",
		BalanceType: utils.MetaMonetary,
		Value:       3.62,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify debited account on engineOne
	if err := engineOneRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 13.78 {
		t.Errorf("Expecting: %+v, received: %+v",
			13.78, val)
	}

	// verify debited account on engineTwo
	if err := engineTwoRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 13.78 {
		t.Errorf("Expecting: %+v, received: %+v",
			13.78, val)
	}

	addBal := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccount1",
		BalanceType: utils.MetaMonetary,
		Value:       12.765,
	}
	// add balance for the account on internal and this should be replicated to other engines
	if err := internalRPC.Call(utils.APIerSv1AddBalance, addBal, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
	// verify account on engineOne
	if err := engineOneRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 26.545 {
		t.Errorf("Expecting: %+v, received: %+v",
			26.545, val)
	}

	// verify account on engineTwo
	if err := engineTwoRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 26.545 {
		t.Errorf("Expecting: %+v, received: %+v",
			26.545, val)
	}

}

func testInternalSetAccount(t *testing.T) {
	var reply string

	if err := internalRPC.Call(utils.APIerSv1SetAccount,
		utils.AttrSetAccount{
			Tenant:          "cgrates.org",
			Account:         "testSetAccount",
			ActionPlanID:    "AP_PACKAGE_10",
			ReloadScheduler: true,
		}, &reply); err != nil {
		t.Error(err)
	}
	// give some time to scheduler to execute the action
	time.Sleep(100 * time.Millisecond)

	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testSetAccount",
	}
	// verify account on engineOne
	if err := engineOneRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 10 {
		t.Errorf("Expecting: %+v, received: %+v",
			10, val)
	}

	// verify account on engineTwo
	if err := engineTwoRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expecting: %+v, received: %+v",
			1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if val := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); val != 10 {
		t.Errorf("Expecting: %+v, received: %+v",
			10, val)
	}
}

func testInternalReplicateStats(t *testing.T) {
	var reply string

	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "StatsToReplicate",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}

	if err := internalRPC.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var rcv *engine.StatQueueProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "StatsToReplicate"}, &rcv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(rcv))
	}

	if err := engineTwoRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "StatsToReplicate"}, &rcv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(rcv))
	}

}

func testInternalRemoteITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
