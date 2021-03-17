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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sTestsInternalReplicateIT = []func(t *testing.T){
		testInternalReplicateITInitCfg,
		testInternalReplicateITDataFlush,
		testInternalReplicateITStartEngine,
		testInternalReplicateITRPCConn,
		testInternalReplicateLoadDataInInternalEngine,

		testInternalReplicateITDestination,
		testInternalReplicateITAttributeProfile,
		testInternalReplicateITRatingProfile,
		testInternalReplicateITRouteProfile,
		testInternalReplicateITStatQueueProfile,
		testInternalReplicateITDispatcherProfile,
		testInternalReplicateITChargerProfile,
		testInternalReplicateITDispatcherHost,
		testInternalReplicateITFilter,
		testInternalReplicateITResourceProfile,
		testInternalReplicateITActions,
		testInternalReplicateITActionPlan,
		testInternalReplicateITThresholdProfile,
		testInternalReplicateITSetAccount,
		testInternalReplicateITActionTrigger,
		testInternalReplicateITThreshold,
		testInternalReplicateITRateProfile,
		testInternalReplicateITLoadIds,

		testInternalReplicateITKillEngine,
	}
)

func TestInternalReplicateIT(t *testing.T) {
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
	for _, stest := range sTestsInternalReplicateIT {
		t.Run(*dbType, stest)
	}
}

func testInternalReplicateITInitCfg(t *testing.T) {
	var err error
	internalCfgPath = path.Join(*dataDir, "conf", "samples", "replication", internalCfgDirPath)
	internalCfg, err = config.NewCGRConfigFromPath(internalCfgPath)
	if err != nil {
		t.Error(err)
	}

	// prepare config for engine1
	engineOneCfgPath = path.Join(*dataDir, "conf", "samples", "replication", engineOneCfgDirPath)
	engineOneCfg, err = config.NewCGRConfigFromPath(engineOneCfgPath)
	if err != nil {
		t.Error(err)
	}

	// prepare config for engine2
	engineTwoCfgPath = path.Join(*dataDir, "conf", "samples", "replication", engineTwoCfgDirPath)
	engineTwoCfg, err = config.NewCGRConfigFromPath(engineTwoCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testInternalReplicateITDataFlush(t *testing.T) {
	if err := engine.InitDataDb(engineOneCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(engineTwoCfg); err != nil {
		t.Fatal(err)
	}
}

func testInternalReplicateITStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(engineOneCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(engineTwoCfgPath, 500); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(internalCfgPath, 500); err != nil {
		t.Fatal(err)
	}
}

func testInternalReplicateITRPCConn(t *testing.T) {
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

func testInternalReplicateLoadDataInInternalEngine(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := internalRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testInternalReplicateITDestination(t *testing.T) {
	//check
	rpl := &engine.Destination{}
	if err := engineOneRPC.Call(utils.APIerSv1GetDestination, utils.StringPointer("testDestination"), &rpl); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDestination, utils.StringPointer("testDestination"), &rpl); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//set
	attrs := utils.AttrSetDestination{Id: "testDestination", Prefixes: []string{"004", "005"}}
	var reply string
	if err := internalRPC.Call(utils.APIerSv1SetDestination, &attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	eDst := &engine.Destination{
		Id:       "testDestination",
		Prefixes: []string{"004", "005"},
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetDestination, utils.StringPointer("testDestination"), &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDst, rpl) {
		t.Errorf("Expected: %v,\n received: %v", eDst, rpl)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDestination, utils.StringPointer("testDestination"), &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDst, rpl) {
		t.Errorf("Expected: %v,\n received: %v", eDst, rpl)
	}

	// remove
	attr := &AttrRemoveDestination{DestinationIDs: []string{"testDestination"}, Prefixes: []string{"004", "005"}}
	if err := internalRPC.Call(utils.APIerSv1RemoveDestination, &attr, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %+v", reply)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetDestination, utils.StringPointer("testDestination"), &rpl); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDestination, utils.StringPointer("testDestination"), &rpl); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITAttributeProfile(t *testing.T) {
	//set
	alsPrf := &engine.AttributeProfileWithOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_CDRE",
			Contexts:  []string{"*cdre"},
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.NewRSRParsersMustCompile("ATTR_SUBJECT", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Category,
					Value: config.NewRSRParsersMustCompile("ATTR_CATEGORY", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	var reply *engine.AttributeProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
	reply = &engine.AttributeProfile{}
	//remove
	if err := internalRPC.Call(utils.APIerSv1RemoveAttributeProfile, &utils.TenantIDWithOpts{TenantID: &utils.TenantID{
		Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check again
	if err := engineOneRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v received: %+v", utils.ErrNotFound, err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v received: %+v", utils.ErrNotFound, err)
	}
}

func testInternalReplicateITRatingProfile(t *testing.T) {
	//check
	var rpl engine.RatingProfile
	attrGetRatingProfile := &utils.AttrGetRatingProfile{
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "Subject"}
	if err := engineOneRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingProfile, &rpl); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v received: %+v", utils.ErrNotFound, err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingProfile, &rpl); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v received: %+v", utils.ErrNotFound, err)
	}
	// set
	var reply string
	attrSetRatingProfile := &utils.AttrSetRatingProfile{
		Overwrite: true,
		Tenant:    "cgrates.org",
		Category:  "call",
		Subject:   "Subject",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime:   "2012-01-01T00:00:00Z",
				RatingPlanId:     "RP_1001",
				FallbackSubjects: "FallbackSubjects"},
		}}
	if err := internalRPC.Call(utils.APIerSv1SetRatingProfile, &attrSetRatingProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	// Calling the second time should not raise EXISTS
	if err := internalRPC.Call(utils.APIerSv1SetRatingProfile, &attrSetRatingProfile, &reply); err != nil {
		t.Error(err)
	}
	//check
	actTime, err := utils.ParseTimeDetectLayout("2012-01-01T00:00:00Z", utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:call:Subject",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RP_1001",
				FallbackKeys:   []string{"*out:cgrates.org:call:FallbackSubjects"},
			},
		},
	}
	if err := engineOneRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingProfile, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetRatingProfile, attrGetRatingProfile, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}
}

func testInternalReplicateITRouteProfile(t *testing.T) {
	// check
	var reply *engine.RouteProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	rPrf := &RouteWithOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "TEST_PROFILE1",
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Routes: []*engine.Route{
				{
					ID:              "SPL1",
					RatingPlanIDs:   []string{"RP1"},
					AccountIDs:      []string{"Acc"},
					ResourceIDs:     []string{"Res1", "ResGroup2"},
					StatIDs:         []string{"Stat1"},
					Weight:          20,
					Blocker:         false,
					RouteParameters: "SortingParameter1",
				},
			},
			Weight: 10,
		},
	}
	// set
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetRouteProfile, rPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", rPrf.RouteProfile, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", rPrf.RouteProfile, reply)
	}
	// remove
	var resp string
	if err := internalRPC.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITStatQueueProfile(t *testing.T) {
	// check
	var reply *engine.StatQueueProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	statConfig = &engine.StatQueueProfileWithOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: tenant,
			ID:     "TEST_PROFILE1",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: "*sum",
				},
				{
					MetricID: "*acd",
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := engineOneRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}
	//remove
	if err := internalRPC.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITDispatcherProfile(t *testing.T) {
	// check
	var reply string
	if err := engineOneRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:    "cgrates.org",
			ID:        "Dsp1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Strategy:  utils.MetaFirst,
			Weight:    20,
		},
	}
	if err := internalRPC.Call(utils.APIerSv1SetDispatcherProfile, dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
	// check
	var dsp *engine.DispatcherProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherProfile.DispatcherProfile, dsp)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherProfile.DispatcherProfile, dsp)
	}
	// remove
	var result string
	if err := internalRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}
	// remove again
	if err := internalRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// check again
	if err := engineOneRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}, &dsp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}, &dsp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITChargerProfile(t *testing.T) {
	// check
	var reply *engine.ChargerProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	chargerProfile = &ChargerWithOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~Account:1002"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"Attr1", "Attr2"},
			Weight:       20,
		},
	}
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chargerProfile.ChargerProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", chargerProfile.ChargerProfile, reply)
	}
	// remove
	if err := internalRPC.Call(utils.APIerSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check
	if err := engineOneRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITDispatcherHost(t *testing.T) {
	// check
	var reply string
	if err := engineOneRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	dispatcherHost = &engine.DispatcherHostWithOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:      "DspHst1",
				Address: "*internal",
			},
		},
	}
	//set
	if err := internalRPC.Call(utils.APIerSv1SetDispatcherHost,
		dispatcherHost,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
	// check
	var dsp *engine.DispatcherHost
	if err := engineOneRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherHost.DispatcherHost, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherHost.DispatcherHost, dsp)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherHost.DispatcherHost, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherHost.DispatcherHost, dsp)
	}
	// remove
	if err := internalRPC.Call(utils.APIerSv1RemoveDispatcherHost,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
	//check
	if err := engineOneRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITFilter(t *testing.T) {
	// check
	var reply *engine.Filter
	if err := engineOneRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//set
	filter = &engine.FilterWithOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "Filter1",
			Rules: []*engine.FilterRule{
				{
					Element: "~*req.Account",
					Type:    utils.MetaString,
					Values:  []string{"1001", "1002"},
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var rcv string
	if err := internalRPC.Call(utils.APIerSv1SetFilter, filter, &rcv); err != nil {
		t.Error(err)
	} else if rcv != utils.OK {
		t.Error("Unexpected reply returned", rcv)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(filter.Filter, reply) {
		t.Errorf("Expecting : %+v, received: %+v", filter.Filter, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(filter.Filter, reply) {
		t.Errorf("Expecting : %+v, received: %+v", filter.Filter, reply)
	}
	// remove
	var resp string
	if err := internalRPC.Call(utils.APIerSv1RemoveFilter,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	// check again
	if err := engineOneRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetFilter, &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITResourceProfile(t *testing.T) {
	// check
	var reply *engine.ResourceProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_GR_TEST"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_GR_TEST"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	rlsConfig = &engine.ResourceProfileWithOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RES_GR_TEST",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Nanosecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            20,
			ThresholdIDs:      []string{"Val1"},
		},
	}

	var result string
	if err := internalRPC.Call(utils.APIerSv1SetResourceProfile, rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_GR_TEST"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig.ResourceProfile) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig.ResourceProfile), utils.ToJSON(reply))
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_GR_TEST"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig.ResourceProfile) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig.ResourceProfile), utils.ToJSON(reply))
	}
	// remove
	if err := internalRPC.Call(utils.APIerSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_GR_TEST"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check again
	if err := engineOneRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_GR_TEST"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RES_GR_TEST"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITActions(t *testing.T) {
	// check
	var reply1 []*utils.TPAction
	if err := engineOneRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err == nil || err.Error() != "SERVER_ERROR: NOT_FOUND" {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err == nil || err.Error() != "SERVER_ERROR: NOT_FOUND" {
		t.Error(err)
	}
	// set
	attrs1 := &V1AttrSetActions{
		ActionsId: "ACTS_1",
		Actions: []*V1TPAction{{
			Identifier:  utils.MetaTopUpReset,
			BalanceType: utils.MetaMonetary,
			Units:       75.0,
			ExpiryTime:  utils.MetaUnlimited,
			Weight:      20.0}}}
	var reply string
	if err := internalRPC.Call(utils.APIerSv1SetActions, &attrs1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
	if err := internalRPC.Call(utils.APIerSv1SetActions, &attrs1, &reply); err == nil || err.Error() != "EXISTS" {
		t.Error("Unexpected result on duplication: ", err)
	}
	// check
	eOut := []*utils.TPAction{{
		Identifier:      utils.MetaTopUpReset,
		BalanceType:     utils.MetaMonetary,
		Units:           "75",
		BalanceWeight:   "0",
		BalanceBlocker:  "false",
		BalanceDisabled: "false",
		ExpiryTime:      utils.MetaUnlimited,
		Weight:          20.0,
	}}
	if err := internalRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(eOut, reply1) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(eOut), utils.ToJSON(reply1))
	}
	if err := engineOneRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(eOut, reply1) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(eOut), utils.ToJSON(reply1))
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(eOut, reply1) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(eOut), utils.ToJSON(reply1))
	}
	// remove
	if err := internalRPC.Call(utils.APIerSv1RemoveActions,
		&AttrRemoveActions{
			ActionIDs: []string{"ACTS_1"}}, &reply); err != nil {
		t.Error("Got error on APIerSv1.RemoveActions: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply when calling APIerSv1.RemoveActions: ", err.Error())
	}
	// check again
	if err := engineOneRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err == nil || err.Error() != "SERVER_ERROR: NOT_FOUND" {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActions, utils.StringPointer("ACTS_1"), &reply1); err == nil || err.Error() != "SERVER_ERROR: NOT_FOUND" {
		t.Error(err)
	}

}

func testInternalReplicateITActionPlan(t *testing.T) {
	var reply string
	if err := internalRPC.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACTS_1",
		Actions:   []*utils.TPAction{{Identifier: utils.MetaLog}},
	}, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	// check
	var aps []*engine.ActionPlan
	if err := engineOneRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "ATMS_1"}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error at APIerSv1.GetActionPlan: %+v", err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "ATMS_1"}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error at APIerSv1.GetActionPlan: %+v", err)
	}
	// set
	atms1 := &AttrSetActionPlan{
		Id: "ATMS_1",
		ActionPlan: []*AttrActionPlan{
			{
				ActionsId: "ACTS_1",
				Time:      utils.MetaASAP,
				Weight:    20.0},
		},
	}
	var reply1 string
	if err := internalRPC.Call(utils.APIerSv1SetActionPlan, &atms1, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply1)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "ATMS_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "ATMS_1" {
		t.Errorf("Expected: ATMS_1,\n received: %v", aps[0].Id)
	} else if aps[0].ActionTimings[0].ActionsID != "ACTS_1" {
		t.Errorf("Expected: ACTS_1,\n received: %v", aps[0].ActionTimings[0].ActionsID)
	} else if aps[0].ActionTimings[0].Weight != 20.0 {
		t.Errorf("Expected: 20.0,\n received: %v", aps[0].ActionTimings[0].Weight)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "ATMS_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "ATMS_1" {
		t.Errorf("Expected: ATMS_1,\n received: %v", aps[0].Id)
	} else if aps[0].ActionTimings[0].ActionsID != "ACTS_1" {
		t.Errorf("Expected: ACTS_1,\n received: %v", aps[0].ActionTimings[0].ActionsID)
	} else if aps[0].ActionTimings[0].Weight != 20.0 {
		t.Errorf("Expected: 20.0,\n received: %v", aps[0].ActionTimings[0].Weight)
	}
	// remove
	if err := internalRPC.Call(utils.APIerSv1RemoveActionPlan, &AttrGetActionPlan{
		ID: "ATMS_1"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//check again
	if err := engineOneRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "ATMS_1"}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v, rcv: %+v", err, utils.ToJSON(aps))
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "ATMS_1"}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v, rcv: %+v", err, utils.ToJSON(aps))
	}
}

func testInternalReplicateITThresholdProfile(t *testing.T) {
	// check
	var reply *engine.ThresholdProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	filter = &engine.FilterWithOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "TestFilter",
			Rules: []*engine.FilterRule{{
				Element: "~*req.Account",
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tPrfl = &engine.ThresholdProfileWithOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    tenant,
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"TestFilter"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
	}
	if err := internalRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	// remove
	if err := internalRPC.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check again
	if err := engineOneRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITSetAccount(t *testing.T) {
	if *encoding == utils.MetaGOB {
		t.SkipNow() // skip this function because
		// APIerSv1GetAccount returns the old format of Account
		// and it can not register that interface because is duplicate
		// of the real Account
	}
	//check
	var reply string
	if err := engineOneRPC.Call(utils.APIerSv1GetAccount,
		&utils.AttrGetAccount{Account: "AccountTest", Tenant: tenant}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetAccount,
		&utils.AttrGetAccount{Account: "AccountTest", Tenant: tenant}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//set
	attrSetAccount := &utils.AttrSetAccount{
		Account: "AccountTest",
		Tenant:  tenant}
	if err := internalRPC.Call(utils.APIerSv1SetAccount, attrSetAccount, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//check
	tmp := engine.Account{}
	rcvAccount := tmp.AsOldStructure()
	if err := engineOneRPC.Call(utils.APIerSv1GetAccount,
		&utils.AttrGetAccount{Account: "AccountTest", Tenant: tenant}, &rcvAccount); err != nil {
		t.Errorf("Unexpected error : %+v\nRCV: %+v", err, rcvAccount)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetAccount,
		&utils.AttrGetAccount{Account: "AccountTest", Tenant: tenant}, &rcvAccount); err != nil {
		t.Errorf("Unexpected error : %+v", err)
	}
	//remove
	if err := internalRPC.Call(utils.APIerSv1RemoveAccount,
		&utils.AttrRemoveAccount{
			Account: "AccountTest",
			Tenant:  tenant}, &reply); err != nil {
		t.Errorf("Unexpected error : %+v", err)
	}
	//check
	if err := engineOneRPC.Call(utils.APIerSv1GetAccount,
		&utils.AttrGetAccount{Account: "AccountTest", Tenant: tenant}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetAccount,
		&utils.AttrGetAccount{Account: "AccountTest", Tenant: tenant}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testInternalReplicateITActionTrigger(t *testing.T) {
	// check
	var atrs engine.ActionTriggers
	if err := engineOneRPC.Call(utils.APIerSv1GetActionTriggers,
		&AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error("Got error on APIerSv1.GetActionTriggers: ", err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionTriggers,
		&AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error("Got error on APIerSv1.GetActionTriggers: ", err)
	}
	// set
	var reply string
	attrSet := AttrSetActionTrigger{
		GroupID:  "TestATR",
		UniqueID: "UniqueID",
		ActionTrigger: map[string]interface{}{
			utils.BalanceID: utils.StringPointer("BalanceIDtest1"),
		}}

	if err := internalRPC.Call(utils.APIerSv1SetActionTrigger, attrSet, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling v1.SetActionTrigger got: %v", reply)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetActionTriggers, &AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err != nil {
		t.Error("Got error on APIerSv1.GetActionTriggers: ", err)
	} else if len(atrs) != 1 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	} else if atrs[0].ID != "TestATR" {
		t.Errorf("Expecting: TestATR, received: %+v", atrs[0].ID)
	} else if atrs[0].UniqueID != "UniqueID" {
		t.Errorf("Expecting UniqueID, received: %+v", atrs[0].UniqueID)
	} else if *atrs[0].Balance.ID != "BalanceIDtest1" {
		t.Errorf("Expecting BalanceIDtest1, received: %+v", atrs[0].Balance.ID)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionTriggers, &AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err != nil {
		t.Error("Got error on APIerSv1.GetActionTriggers: ", err)
	} else if len(atrs) != 1 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	} else if atrs[0].ID != "TestATR" {
		t.Errorf("Expecting: TestATR, received: %+v", atrs[0].ID)
	} else if atrs[0].UniqueID != "UniqueID" {
		t.Errorf("Expecting UniqueID, received: %+v", atrs[0].UniqueID)
	} else if *atrs[0].Balance.ID != "BalanceIDtest1" {
		t.Errorf("Expecting BalanceIDtest1, received: %+v", atrs[0].Balance.ID)
	}
	//remove
	asttrRemove := &AttrRemoveActionTrigger{
		GroupID:  "TestATR",
		UniqueID: "UniqueID",
	}
	if err := internalRPC.Call(utils.APIerSv1RemoveActionTrigger, asttrRemove, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling v1.RemoveActionTrigger got: %v", reply)
	}
	//check
	if err := engineOneRPC.Call(utils.APIerSv1GetActionTriggers,
		&AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Got error on APIerSv1.GetActionTriggers: %+v", err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionTriggers,
		&AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error("Got error on APIerSv1.GetActionTriggers: ", err)
	}
}

func testInternalReplicateITThreshold(t *testing.T) {
	// get threshold
	var td engine.Threshold
	if err := engineOneRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{
				Tenant: tenant,
				ID:     "THD_Test"},
		}, &td); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{
				Tenant: tenant,
				ID:     "THD_Test"},
		}, &td); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tEvs := engine.ThresholdsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EventType:     utils.AccountUpdate,
				utils.AccountField:  "1005",
				utils.AllowNegative: true,
				utils.Disabled:      false,
				utils.Units:         12.3},
			Opts: map[string]interface{}{
				utils.MetaEventType: utils.AccountUpdate,
			},
		},
	}
	//set Actions
	var reply string
	if err := internalRPC.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACT_LOG",
		Actions:   []*utils.TPAction{{Identifier: utils.MetaLog}},
	}, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	tPrfl := engine.ThresholdProfileWithOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    tenant,
			ID:        "THD_Test",
			FilterIDs: []string{},
			MaxHits:   -1,
			Weight:    30,
			ActionIDs: []string{"ACT_LOG"},
		},
	}
	// set Threshold
	if err := internalRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//get
	if err := internalRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{
				Tenant: tenant,
				ID:     "THD_Test"},
		}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 0 { //still not processed
		t.Errorf("Expecting threshold to be hit once received: %v", td.Hits)
	}
	//set account
	attrSetAccount := &utils.AttrSetAccount{
		Account: "1005",
		Tenant:  tenant,
		ExtraOptions: map[string]bool{
			utils.AllowNegative: true}}
	if err := internalRPC.Call(utils.APIerSv1SetAccount, attrSetAccount, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//set balance
	attrs := &utils.AttrSetBalance{
		Tenant:      tenant,
		Account:     "1005",
		BalanceType: utils.MetaMonetary,
		Value:       1,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 10.0,
		},
	}
	if err := internalRPC.Call(utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Fatal(err)
	}
	// processEvent
	var ids []string
	//eIDs := []string{}
	if err := internalRPC.Call(utils.ThresholdSv1ProcessEvent, &tEvs, &ids); err != nil {
		t.Error(err)
	} else if len(ids) != 1 {
		t.Errorf("Expecting 1: ,received %+v", len(ids))
	} else if ids[0] != "THD_Test" {
		t.Errorf("Expecting: THD_Test, received %q", ids[0])
	}
	//get
	if err := internalRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{
				Tenant: tenant,
				ID:     "THD_Test"},
		}, &td); err != nil {
		t.Error(err)
	} else if td.Hits != 1 { //processed
		t.Errorf("Expecting threshold to be hit once received: %v", td.Hits)
	}

	// remove
	var result string
	if err := internalRPC.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "THD_Test"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := engineOneRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{
				Tenant: tenant,
				ID:     "THD_Test"},
		}, &td); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithOpts{
			TenantID: &utils.TenantID{
				Tenant: tenant,
				ID:     "THD_Test"},
		}, &td); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testInternalReplicateITRateProfile(t *testing.T) {
	//set
	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * 1-5",
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
			},
		},
	}

	apiRPrf := &engine.APIRateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Subject:1001"},
		Weights:         ";0",
		MaxCostStrategy: "*free",
		Rates: map[string]*engine.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weights:         ";10",
				ActivationTimes: "* * * * 0,6",
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weights:         ";30",
				ActivationTimes: "* * 24 12 *",
			},
		},
	}
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetRateProfile, apiRPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	var reply *engine.RateProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: rPrf.Tenant, ID: rPrf.ID}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", rPrf, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: rPrf.Tenant, ID: rPrf.ID}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", rPrf, reply)
	}
	//remove
	if err := internalRPC.Call(utils.APIerSv1RemoveRateProfile, &utils.TenantIDWithOpts{TenantID: &utils.TenantID{
		Tenant: rPrf.Tenant, ID: rPrf.ID}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check again
	if err := engineOneRPC.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: rPrf.Tenant, ID: rPrf.ID}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v received: %+v", utils.ErrNotFound, err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: rPrf.Tenant, ID: rPrf.ID}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v received: %+v", utils.ErrNotFound, err)
	}
}

func testInternalReplicateITLoadIds(t *testing.T) {
	// get LoadIDs
	var rcv1e1 map[string]int64
	if err := engineOneRPC.Call(utils.APIerSv1GetLoadIDs, utils.StringPointer(utils.EmptyString), &rcv1e1); err != nil {
		t.Error(err)
	}
	var rcv1e2 map[string]int64
	if err := engineTwoRPC.Call(utils.APIerSv1GetLoadIDs, utils.StringPointer(utils.EmptyString), &rcv1e2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rcv1e1, rcv1e2) {
		t.Errorf("Expecting same LoadIDs for both engines")
	}
	// set AttributeProfile
	alsPrf = &engine.AttributeProfileWithOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "AttributeWithNonSubstitute",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Account:1008"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2020, 4, 18, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2020, 4, 18, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.Account:1008"},
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value:     config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.NewRSRParsersMustCompile(utils.MetaRemove, utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check AttributeProfile
	var reply *engine.AttributeProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
	// check again the LoadIDs
	var rcv2e1 map[string]int64
	if err := engineOneRPC.Call(utils.APIerSv1GetLoadIDs, utils.StringPointer(utils.EmptyString), &rcv2e1); err != nil {
		t.Error(err)
	}
	var rcv2e2 map[string]int64
	if err := engineTwoRPC.Call(utils.APIerSv1GetLoadIDs, utils.StringPointer(utils.EmptyString), &rcv2e2); err != nil {
		t.Error(err)
	}

	// needs to be different LoadIds after the APIerSv1SetAttributeProfile call
	if reflect.DeepEqual(rcv1e1, rcv2e1) {
		t.Errorf("Expecting same LoadIDs for both engines")
	}
	// needs to be same LoadIds in both engines
	if !reflect.DeepEqual(rcv2e1, rcv2e2) {
		t.Errorf("Expecting same LoadIDs for both engines")
	}
	// check if the data was corectly modified after the APIerSv1SetAttributeProfile call
	// only CacheAttributeProfiles should differ
	if rcv1e1[utils.CacheAttributeProfiles] == rcv2e1[utils.CacheAttributeProfiles] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheAttributeProfiles], rcv2e1[utils.CacheAttributeProfiles])
	} else if rcv1e1[utils.CacheAccountActionPlans] != rcv2e1[utils.CacheAccountActionPlans] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheAccountActionPlans], rcv2e1[utils.CacheAccountActionPlans])
	} else if rcv1e1[utils.CacheActionPlans] != rcv2e1[utils.CacheActionPlans] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheActionPlans], rcv2e1[utils.CacheActionPlans])
	} else if rcv1e1[utils.CacheActions] != rcv2e1[utils.CacheActions] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheActions], rcv2e1[utils.CacheActions])
	} else if rcv1e1[utils.CacheChargerProfiles] != rcv2e1[utils.CacheChargerProfiles] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheChargerProfiles], rcv2e1[utils.CacheChargerProfiles])
	} else if rcv1e1[utils.CacheDestinations] != rcv2e1[utils.CacheDestinations] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheDestinations], rcv2e1[utils.CacheDestinations])
	} else if rcv1e1[utils.CacheFilters] != rcv2e1[utils.CacheFilters] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheFilters], rcv2e1[utils.CacheFilters])
	} else if rcv1e1[utils.CacheRatingPlans] != rcv2e1[utils.CacheRatingPlans] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheRatingPlans], rcv2e1[utils.CacheRatingPlans])
	} else if rcv1e1[utils.CacheRatingProfiles] != rcv2e1[utils.CacheRatingProfiles] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheRatingProfiles], rcv2e1[utils.CacheRatingProfiles])
	} else if rcv1e1[utils.CacheResourceProfiles] != rcv2e1[utils.CacheResourceProfiles] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheResourceProfiles], rcv2e1[utils.CacheResourceProfiles])
	} else if rcv1e1[utils.CacheResources] != rcv2e1[utils.CacheResources] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheResources], rcv2e1[utils.CacheResources])
	} else if rcv1e1[utils.CacheReverseDestinations] != rcv2e1[utils.CacheReverseDestinations] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheReverseDestinations], rcv2e1[utils.CacheReverseDestinations])
	} else if rcv1e1[utils.CacheStatQueueProfiles] != rcv2e1[utils.CacheStatQueueProfiles] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheStatQueueProfiles], rcv2e1[utils.CacheStatQueueProfiles])
	} else if rcv1e1[utils.CacheRouteProfiles] != rcv2e1[utils.CacheRouteProfiles] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheRouteProfiles], rcv2e1[utils.CacheRouteProfiles])
	} else if rcv1e1[utils.CacheThresholdProfiles] != rcv2e1[utils.CacheThresholdProfiles] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheThresholdProfiles], rcv2e1[utils.CacheThresholdProfiles])
	} else if rcv1e1[utils.CacheThresholds] != rcv2e1[utils.CacheThresholds] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheThresholds], rcv2e1[utils.CacheThresholds])
	} else if rcv1e1[utils.CacheTimings] != rcv2e1[utils.CacheTimings] {
		t.Errorf("Expecting: %+v, received: %+v", rcv1e1[utils.CacheTimings], rcv2e1[utils.CacheTimings])
	}
}

func testInternalReplicateITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
