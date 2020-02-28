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

		// testInternalReplicateITDestination,
		testInternalReplicateITAttributeProfile,
		// testInternalReplicateITRatingProfile,
		testInternalReplicateITSupplierProfile,
		testInternalReplicateITStatQueueProfile,
		testInternalReplicateITDispatcherProfile,
		testInternalReplicateITChargerProfile,
		testInternalReplicateITDispatcherHost,
		testInternalReplicateITFilter,
		testInternalReplicateITResourceProfile,
		// testInternalReplicateITActions,
		testInternalReplicateITActionPlan,
		testInternalReplicateITThresholdProfile,

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
	internalCfgPath = path.Join(*dataDir, "conf", "samples", "remote_replication", internalCfgDirPath)
	internalCfg, err = config.NewCGRConfigFromPath(internalCfgPath)
	if err != nil {
		t.Error(err)
	}
	internalCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(internalCfg)

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

func testInternalReplicateITDataFlush(t *testing.T) {
	if err := engine.InitDataDb(engineOneCfg); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if err := engine.InitDataDb(engineTwoCfg); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
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
	time.Sleep(200 * time.Millisecond)
	engineTwoRPC, err = newRPCClient(engineTwoCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	internalRPC, err = newRPCClient(internalCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testInternalReplicateITDestination(t *testing.T) {
	//set
	attrs := utils.AttrSetDestination{Id: "TEST_SET_DESTINATION3", Prefixes: []string{"004", "005"}}
	var reply string
	if err := internalRPC.Call(utils.APIerSv1SetDestination, attrs, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	eDst := &engine.Destination{
		Id:       "TEST_SET_DESTINATION3",
		Prefixes: []string{"004", "005"},
	}
	// check
	rpl := &engine.Destination{}
	if err := engineOneRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION3", &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDst, rpl) {
		t.Errorf("Expected: %v,\n received: %v", eDst, rpl)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION3", &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eDst, rpl) {
		t.Errorf("Expected: %v,\n received: %v", eDst, rpl)
	}

	// remove
	attr := &AttrRemoveDestination{DestinationIDs: []string{"TEST_SET_DESTINATION"}, Prefixes: []string{"004", "005"}}
	if err := internalRPC.Call(utils.APIerSv1RemoveDestination, attr, &reply); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION", &rpl); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetDestination, "TEST_SET_DESTINATION", &rpl); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testInternalReplicateITAttributeProfile(t *testing.T) {
	//set
	alsPrf := &AttributeWithCache{
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
					Value: config.NewRSRParsersMustCompile("ATTR_SUBJECT", true, utils.INFIELD_SEP),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Category,
					Value: config.NewRSRParsersMustCompile("ATTR_CATEGORY", true, utils.INFIELD_SEP),
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
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
	reply = &engine.AttributeProfile{}
	//remove
	if err := internalRPC.Call(utils.APIerSv1RemoveAttributeProfile, &utils.TenantIDWithCache{
		Tenant: alsPrf.Tenant, ID: alsPrf.ID}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//check again
	if err := engineOneRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v recived: %+v", utils.ErrNotFound, err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: alsPrf.Tenant, ID: alsPrf.ID}}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting: %+v recived: %+v", utils.ErrNotFound, err)
	}
}

func testInternalReplicateITRatingProfile(t *testing.T) {
	// set
	var reply string
	attrSetRatingProfile := &utils.AttrSetRatingProfile{
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "Subject",
		RatingPlanActivations: []*utils.TPRatingActivation{
			&utils.TPRatingActivation{
				ActivationTime:   "2012-01-01T00:00:00Z",
				RatingPlanId:     "RETAIL1",
				FallbackSubjects: "FallbackSubjects"},
		}}
	if err := internalRPC.Call(utils.APIerSv1SetRatingProfile, attrSetRatingProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	// Calling the second time should not raise EXISTS
	if err := internalRPC.Call(utils.APIerSv1SetRatingProfile, attrSetRatingProfile, &reply); err != nil {
		t.Error(err)
	}
	//check
	var rpl engine.RatingProfile
	attrGetRatingProfile := &utils.AttrGetRatingProfile{
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "Subject"}
	actTime, err := utils.ParseTimeDetectLayout("2012-01-01T00:00:00Z", utils.EmptyString)
	if err != nil {
		t.Error(err)
	}
	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:call:Subject",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RETAIL1",
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

func testInternalReplicateITSupplierProfile(t *testing.T) {
	// check
	var reply *engine.SupplierProfile
	if err := engineOneRPC.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:            "cgrates.org",
			ID:                "TEST_PROFILE1",
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Suppliers: []*engine.Supplier{
				{
					ID:                 "SPL1",
					RatingPlanIDs:      []string{"RP1"},
					AccountIDs:         []string{"Acc"},
					ResourceIDs:        []string{"Res1", "ResGroup2"},
					StatIDs:            []string{"Stat1"},
					Weight:             20,
					Blocker:            false,
					SupplierParameters: "SortingParameter1",
				},
			},
			Weight: 10,
		},
	}
	// set
	var result string
	if err := internalRPC.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
	// remove
	var resp string
	if err := internalRPC.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetSupplierProfile,
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
	statConfig = &StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: tenant,
			ID:     "TEST_PROFILE1",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				&engine.MetricWithFilters{
					MetricID: "*sum",
				},
				&engine.MetricWithFilters{
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
	dispatcherProfile = &DispatcherWithCache{
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
	time.Sleep(20 * time.Millisecond)
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
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Dsp1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}
	// remove again
	if err := internalRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Dsp1"}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
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
	chargerProfile = &ChargerWithCache{
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
	// if err := internalRPC.Call(utils.APIerSv1GetChargerProfile,
	// 	&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"},
	// 	&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
	// 	t.Errorf("Error: %v, rcv: %+v", err, utils.ToJSON(reply))
	// }
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
	dispatcherHost = &DispatcherHostWithCache{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			ID:     "DspHst1",
			Conns: []*config.RemoteHost{
				&config.RemoteHost{
					Address: "*internal",
				},
				&config.RemoteHost{
					Address:   ":2012",
					Transport: utils.MetaJSON,
					TLS:       true,
				},
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
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "DspHst1"},
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
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "Filter1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.MetaString,
					Type:    "~Account",
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
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Filter1"}, &resp); err != nil {
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
	rlsConfig = &ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RES_GR_TEST",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(1) * time.Nanosecond,
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
	if err := engineOneRPC.Call(utils.APIerSv1GetActions, "ACTS_1", &reply1); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActions, "ACTS_1", &reply1); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	attrs1 := &V1AttrSetActions{
		ActionsId: "ACTS_1",
		Actions: []*V1TPAction{
			&V1TPAction{
				Identifier:  utils.TOPUP_RESET,
				BalanceType: utils.MONETARY,
				Units:       75.0,
				ExpiryTime:  utils.UNLIMITED,
				Weight:      20.0}}}
	var reply string
	if err := internalRPC.Call(utils.APIerSv1SetActions, attrs1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
	if err := internalRPC.Call(utils.APIerSv1SetActions, attrs1, &reply); err == nil || err.Error() != "EXISTS" {
		t.Error("Unexpected result on duplication: ", err.Error())
	}
	//check
	if err := engineOneRPC.Call(utils.APIerSv1GetActions, "ACTS_1", &reply1); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(attrs1.Actions, reply1) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(attrs1.Actions), utils.ToJSON(reply1))
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActions, "ACTS_1", &reply1); err != nil {
		t.Error("Got error on APIerSv1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(attrs1.Actions, reply1) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(attrs1.Actions), utils.ToJSON(reply1))
	}
}

func testInternalReplicateITActionPlan(t *testing.T) {
	var reply string
	if err := internalRPC.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACTS_1",
		Actions:   []*utils.TPAction{{Identifier: utils.LOG}},
	}, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	// check
	var aps []*engine.ActionPlan
	if err := engineOneRPC.Call(utils.APIerSv1GetActionPlan,
		AttrGetActionPlan{ID: utils.EmptyString}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v, rcv: %+v", err, utils.ToJSON(aps))
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionPlan,
		AttrGetActionPlan{ID: utils.EmptyString}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v, rcv: %+v", err, utils.ToJSON(aps))
	}
	// set
	atms1 := &AttrSetActionPlan{
		Id: "ATMS_1",
		ActionPlan: []*AttrActionPlan{
			&AttrActionPlan{
				ActionsId: "ACTS_1",
				MonthDays: "1",
				Time:      "00:00:00",
				Weight:    20.0},
		},
	}
	var reply1 string
	if err := internalRPC.Call(utils.APIerSv1SetActionPlan, atms1, &reply1); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply1 != utils.OK {
		t.Errorf("Unexpected reply returned: %s", reply1)
	}
	// check
	if err := engineOneRPC.Call(utils.APIerSv1GetActionPlan,
		AttrGetActionPlan{ID: utils.EmptyString}, &aps); err != nil {
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
		AttrGetActionPlan{ID: utils.EmptyString}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %+v, rcv: %+v", err, utils.ToJSON(aps))
	}
	if err := engineTwoRPC.Call(utils.APIerSv1GetActionPlan,
		AttrGetActionPlan{ID: utils.EmptyString}, &aps); err == nil || err.Error() != utils.ErrNotFound.Error() {
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
	filter = &FilterWithCache{
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
	tPrfl = &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    tenant,
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"TestFilter"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   1,
			MinSleep:  time.Duration(5 * time.Minute),
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

func testInternalReplicateITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
