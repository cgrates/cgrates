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
		testInternalReplicateLoadDataInEngineTwo,

		// testInternalReplicateSetDestination,
		testInternalReplicateITSetAttributeProfile,
		// testInternalReplicateITSetRatingProfile,
		testInternalReplicateITSetSupplierProfile,
		testInternalReplicateITSetStatQueueProfile,

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

func testInternalReplicateLoadDataInEngineTwo(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := engineTwoRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testInternalReplicateSetDestination(t *testing.T) {
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

func testInternalReplicateITSetAttributeProfile(t *testing.T) {
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
	time.Sleep(30 * time.Millisecond)
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
	time.Sleep(50 * time.Millisecond)
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

func testInternalReplicateITSetRatingProfile(t *testing.T) {
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
	//remove

}

func testInternalReplicateITSetSupplierProfile(t *testing.T) {
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
	time.Sleep(30 * time.Millisecond)
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
	time.Sleep(30 * time.Millisecond)
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

func testInternalReplicateITSetStatQueueProfile(t *testing.T) {
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
	time.Sleep(50 * time.Millisecond)
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
	time.Sleep(50 * time.Millisecond)
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

func testInternalReplicateITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
