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
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
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
	rmtDM              *engine.DataManager
)

var sTestsInternalRemoteIT = []func(t *testing.T){
	testInternalRemoteITDataFlush,
	testInternalRemoteITCheckEmpty,
	testInternalRemoteITLoadData,
	testInternalRemoteITVerifyLoadedDataInRemote,
	testInternalRemoteITInitCfg,
	testInternalRemoteITStartEngine,
	testInternalRemoteITRPCConn,
	testInternalRemoteITGetAccount,
	testInternalRemoteITGetAttribute,
	testInternalRemoteITGetThreshold,
	testInternalRemoteITGetThresholdProfile,
	testInternalRemoteITGetResource,
	testInternalRemoteITGetResourceProfile,
	testInternalRemoteITGetStatQueueProfile,
	testInternalRemoteITGetSupplier,
	testInternalRemoteITGetFilter,
	testInternalRemoteITGetRatingPlan,
	testInternalRemoteITGetRatingProfile,
	testInternalRemoteITGetAction,
	testInternalRemoteITGetActionPlan,
	testInternalRemoteITKillEngine,
}

func TestInternalRemoteITRedis(t *testing.T) {
	internalCfgDirPath = "remote_redis"
	cfg, _ := config.NewDefaultCGRConfig()
	dataDB, err := engine.NewRedisStorage(
		fmt.Sprintf("%s:%s", cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort),
		10, cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
		utils.REDIS_MAX_CONNS, "")
	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
	rmtDM = engine.NewDataManager(dataDB, nil, nil, nil)
	for _, stest := range sTestsInternalRemoteIT {
		t.Run("TestInternalRemoteITRedis", stest)
	}
}

func TestInternalRemoteITMongo(t *testing.T) {
	internalCfgDirPath = "remote_mongo"
	mgoITCfg, err := config.NewCGRConfigFromPath(path.Join(*dataDir, "conf", "samples", "tutmongo"))
	if err != nil {
		t.Fatal(err)
	}
	dataDB, err := engine.NewMongoStorage(mgoITCfg.DataDbCfg().DataDbHost,
		mgoITCfg.DataDbCfg().DataDbPort, mgoITCfg.DataDbCfg().DataDbName,
		mgoITCfg.DataDbCfg().DataDbUser, mgoITCfg.DataDbCfg().DataDbPass,
		utils.DataDB, nil, false)
	if err != nil {
		t.Fatal("Could not connect to Mongo", err.Error())
	}
	rmtDM = engine.NewDataManager(dataDB, nil, nil, nil)
	for _, stest := range sTestsInternalRemoteIT {
		t.Run("TestInternalRemoteITMongo", stest)
	}
}

func testInternalRemoteITDataFlush(t *testing.T) {
	if err := rmtDM.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
}

func testInternalRemoteITCheckEmpty(t *testing.T) {
	test, err := rmtDM.DataDB().IsDBEmpty()
	if err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
}

func testInternalRemoteITLoadData(t *testing.T) {
	loader, err := engine.NewTpReader(rmtDM.DataDB(),
		engine.NewFileCSVStorage(utils.CSV_SEP, path.Join(*dataDir, "tariffplans", "tutorial"), false),
		"", "", nil, nil)
	if err != nil {
		t.Error(err)
	}
	if err := loader.LoadAll(); err != nil {
		t.Error(err)
	}
	if err := loader.WriteToDatabase(false, false); err != nil {
		t.Error(err)
	}
	acc := &engine.Account{
		ID: "cgrates.org:testAccount",
		BalanceMap: map[string]engine.Balances{
			"utils.MONETARY": []*engine.Balance{
				{
					ID:     "testAccount",
					Value:  10,
					Weight: 10,
				},
			},
		},
	}
	if err := rmtDM.DataDB().SetAccount(acc); err != nil {
		t.Error(err)
	}
}

func testInternalRemoteITVerifyLoadedDataInRemote(t *testing.T) {
	exp := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1001_SIMPLEAUTH",
		Contexts:  []string{"simpleauth"},
		FilterIDs: []string{"*string:~Account:1001"},
		Attributes: []*engine.Attribute{
			{
				FieldName: "Password",
				FilterIDs: []string{},
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	if tempAttr, err := rmtDM.GetAttributeProfile("cgrates.org", "ATTR_1001_SIMPLEAUTH",
		false, false, utils.NonTransactional); err != nil {
		t.Errorf("Error: %+v", err)
	} else {
		exp.Compile()
		tempAttr.Compile()
		if !reflect.DeepEqual(exp, tempAttr) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(tempAttr))
		}
	}
}

func testInternalRemoteITInitCfg(t *testing.T) {
	var err error
	internalCfgPath = path.Join(*dataDir, "conf", "samples", internalCfgDirPath)
	internalCfg, err = config.NewCGRConfigFromPath(internalCfgPath)
	if err != nil {
		t.Error(err)
	}
	internalCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(internalCfg)
}

func testInternalRemoteITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(internalCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testInternalRemoteITRPCConn(t *testing.T) {
	var err error
	internalRPC, err = jsonrpc.Dial("tcp", internalCfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func testInternalRemoteITGetAccount(t *testing.T) {
	var acnt *engine.Account
	expAcc := &engine.Account{
		ID: "cgrates.org:testAccount",
		BalanceMap: map[string]engine.Balances{
			"utils.MONETARY": []*engine.Balance{
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
		Account: "testAccount",
	}
	if err := internalRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAcc, acnt) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(expAcc), utils.ToJSON(acnt))
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "nonexistAccount",
	}
	if err := internalRPC.Call("ApierV2.GetAccount", attrs, &acnt); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expecting: %+v, received: %+v", utils.ErrNotFound, err)
	}
}

func testInternalRemoteITGetAttribute(t *testing.T) {
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_1001_SIMPLEAUTH",
			Contexts:  []string{"simpleauth"},
			FilterIDs: []string{"*string:~Account:1001"},

			Attributes: []*engine.Attribute{
				{
					FieldName: "Password",
					FilterIDs: []string{},
					Type:      utils.META_CONSTANT,
					Value:     config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := internalRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1001_SIMPLEAUTH"}, &reply); err != nil {
		t.Fatal(err)
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
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testInternalRemoteITGetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	tPrfl = &ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_ACNT_1001",
			FilterIDs: []string{"FLTR_ACNT_1001"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			MaxHits:   1,
			MinHits:   1,
			MinSleep:  time.Duration(1 * time.Second),
			Weight:    10.0,
			ActionIDs: []string{"ACT_LOG_WARNING"},
			Async:     true,
		},
	}
	if err := internalRPC.Call("ApierV1.GetThresholdProfile",
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
		&utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expectedResources) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedResources), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetResourceProfile(t *testing.T) {
	rlsPrf := &ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResGroup1",
			FilterIDs: []string{"FLTR_RES"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(-1),
			Limit:             7,
			AllocationMessage: "",
			Stored:            true,
			Weight:            10,
			ThresholdIDs:      []string{utils.META_NONE},
		},
	}
	var reply *engine.ResourceProfile
	if err := internalRPC.Call("ApierV1.GetResourceProfile",
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
			&engine.MetricWithFilters{
				MetricID: utils.MetaTCC,
			},
			&engine.MetricWithFilters{
				MetricID: utils.MetaTCD,
			},
		},
		Stored:       false,
		Blocker:      true,
		Weight:       30,
		ThresholdIDs: []string{utils.META_NONE},
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
			&engine.MetricWithFilters{
				MetricID: utils.MetaTCD,
			},
			&engine.MetricWithFilters{
				MetricID: utils.MetaTCC,
			},
		},
		Stored:       false,
		Blocker:      true,
		Weight:       30,
		ThresholdIDs: []string{utils.META_NONE},
	}
	var reply *engine.StatQueueProfile
	if err := internalRPC.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStq, reply) && !reflect.DeepEqual(expStq, expStq2) {
		t.Errorf("Expecting: %+v or %+v, received: %+v", utils.ToJSON(expStq),
			utils.ToJSON(expStq2), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetSupplier(t *testing.T) {
	var reply *engine.SupplierProfile
	splPrf := &engine.SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_ACNT_1001",
		FilterIDs: []string{"FLTR_ACNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			{
				ID:     "supplier1",
				Weight: 10,
			},
			{
				ID:     "supplier2",
				Weight: 20,
			},
		},
		Weight: 20,
	}
	// supplier in reverse order
	splPrf2 := &engine.SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_ACNT_1001",
		FilterIDs: []string{"FLTR_ACNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 0, 0, 0, 0, time.UTC),
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			{
				ID:     "supplier2",
				Weight: 20,
			},
			{
				ID:     "supplier1",
				Weight: 10,
			},
		},
		Weight: 20,
	}

	if err := internalRPC.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "SPL_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf, reply) && !reflect.DeepEqual(splPrf2, reply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(splPrf), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetFilter(t *testing.T) {
	expFltr := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_ACNT_1001",
		Rules: []*engine.FilterRule{
			{
				Type:      "*string",
				FieldName: "~Account",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
	}
	var reply *engine.Filter
	if err := internalRPC.Call("ApierV1.GetFilter",
		&utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFltr, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expFltr), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetRatingPlan(t *testing.T) {
	var reply engine.RatingPlan
	if err := internalRPC.Call("ApierV1.GetRatingPlan", "RP_1001", &reply); err != nil {
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
	expected := engine.RatingProfile{
		Id: "*out:cgrates.org:call:1001",
		RatingPlanActivations: engine.RatingPlanActivations{
			{
				ActivationTime: actTime,
				RatingPlanId:   "RP_1001",
			},
		},
	}
	if err := internalRPC.Call("ApierV1.GetRatingProfile", attrGetRatingPlan, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rpl) {
		t.Errorf("Expected: %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rpl))
	}
}

func testInternalRemoteITGetAction(t *testing.T) {
	expectActs := []*utils.TPAction{
		{Identifier: utils.TOPUP_RESET, BalanceId: "test", BalanceType: utils.MONETARY,
			Units: "10", BalanceWeight: "10", BalanceBlocker: "false",
			BalanceDisabled: "false", ExpiryTime: utils.UNLIMITED, Weight: 10.0}}

	var reply []*utils.TPAction
	if err := internalRPC.Call("ApierV1.GetActions", "ACT_TOPUP_RST_10", &reply); err != nil {
		t.Error("Got error on ApierV1.GetActions: ", err.Error())
	} else if !reflect.DeepEqual(expectActs, reply) {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(expectActs), utils.ToJSON(reply))
	}
}

func testInternalRemoteITGetActionPlan(t *testing.T) {
	var aps *[]*engine.ActionPlan
	//if err := internalRPC.Call("ApierV1.GetActionPlan",
	//	AttrGetActionPlan{ID: "AP_PACKAGE_10"}, &aps); err != nil {
	//	t.Error(err)
	//} else {
	//	fmt.Println(utils.ToJSON(aps))
	//}
	if err := internalRPC.Call("ApierV1.GetActionPlan",
		AttrGetActionPlan{ID: utils.EmptyString}, &aps); err != nil {
		t.Error(err)
	} else {
		fmt.Println(utils.ToJSON(aps))
	}
}

func testInternalRemoteITKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
