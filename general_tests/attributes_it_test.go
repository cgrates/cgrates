//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package general_tests

import (
	"path"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrCfgPath     string
	attrCfg         *config.CGRConfig
	attrRPC         *birpc.Client
	alsPrfConfigDIR string
	sTestsAlsPrf    = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSFlushDBs,

		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testAttributeSLoadFromFolder,
		testAttributeSProcessEvent,
		testAttributeSProcessEventWithAccount,
		testAttributeSProcessEventWithAccountFull,
		testAttributeSProcessEventWithStat,
		testAttributeSProcessEventWithStatFull,
		testAttributeSProcessEventWithResource,
		testAttributeSProcessEventWithResourceFull,
		testAttributeSProcessEventWithLibPhoneNumber,
		testAttributeSProcessEventWithLibPhoneNumberComposed,
		testAttributeSProcessEventWithLibPhoneNumberFull,
		testAttributeSStopEngine,
	}
)

func TestAttributeSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		alsPrfConfigDIR = "attr_test_internal"
	case utils.MetaMySQL:
		alsPrfConfigDIR = "attr_test_mysql"
	case utils.MetaMongo:
		alsPrfConfigDIR = "attr_test_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAttributeSInitCfg(t *testing.T) {
	var err error
	attrCfgPath = path.Join(*utils.DataDir, "conf", "samples", alsPrfConfigDIR)
	attrCfg, err = config.NewCGRConfigFromPath(context.Background(), attrCfgPath)
	if err != nil {
		t.Error(err)
	}
	attrCfg.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

func testAttributeSFlushDBs(t *testing.T) {
	if err := engine.InitDB(attrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(attrCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeSRPCConn(t *testing.T) {
	attrRPC = engine.NewRPCClient(t, attrCfg.ListenCfg(), *utils.Encoding)
}

func testAttributeSLoadFromFolder(t *testing.T) {
	caching := utils.MetaReload
	if attrCfg.DbCfg().DBConns[utils.MetaDefault].Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var replyLD string
	if err := attrRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaCache:       caching,
				utils.MetaStopOnError: true,
			},
		}, &replyLD); err != nil {
		t.Error(err)
	} else if replyLD != utils.OK {
		t.Error("Unexpected reply returned:", replyLD)
	}
	time.Sleep(200 * time.Millisecond)
}

func testAttributeSProcessEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEvent",
		Event: map[string]any{
			utils.EventName: "VariableTest",
			utils.ToR:       utils.MetaVoice,
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_VARIABLE",
				Fields:           []string{utils.MetaReq + utils.NestingSep + utils.Category},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.EventName: "VariableTest",
				utils.Category:  utils.MetaVoice,
				utils.ToR:       utils.MetaVoice,
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithAccount(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_ACCOUNT",
			FilterIDs: []string{"*string:~*req.EventName:AddAccountInfo"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Balance",
					Type:  utils.MetaVariable,
					Value: "10",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACCOUNT"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithAccount",
		Event: map[string]any{
			"EventName": "AddAccountInfo",
		},
		APIOpts: map[string]any{},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_ACCOUNT",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "Balance"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithAccount",
			Event: map[string]any{
				"EventName": "AddAccountInfo",
				"Balance":   "10",
			},
			APIOpts: map[string]any{},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithAccountFull(t *testing.T) {
	var result string
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_ACCOUNT2",
			FilterIDs: []string{"*string:~*req.EventName:AddFullAccount"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "FullAccount",
					Type:  utils.MetaVariable,
					Value: "~*accounts.1001",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACCOUNT2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithAccount2",
		Event: map[string]any{
			"EventName": "AddFullAccount",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_ACCOUNT2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "FullAccount"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithAccount2",
			Event: map[string]any{
				"EventName":   "AddFullAccount",
				"FullAccount": "{\"ID\":\"cgrates.org:1001\",\"BalanceMap\":{\"*monetary\":[{\"Uuid\":\"18160631-a4ae-4078-8048-b4c6b87a36c6\",\"ID\":\"\",\"Value\":10,\"ExpirationDate\":\"0001-01-01T00:00:00Z\",\"Weight\":10,\"DestinationIDs\":{},\"RatingSubject\":\"\",\"Categories\":{},\"SharedGroups\":{},\"Disabled\":false,\"Factor\":null,\"Blocker\":false}]},\"UnitCounters\":null,\"ActionTriggers\":null,\"AllowNegative\":false,\"Disabled\":false,\"UpdateTime\":\"2020-10-06T12:43:51.805Z\"}",
			},
			APIOpts: map[string]any{},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rplyEv.AlteredFields, eRply.AlteredFields) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply.AlteredFields), utils.ToJSON(rplyEv.AlteredFields))
	}
}

func testAttributeSProcessEventWithStat(t *testing.T) {
	// simulate some stat event
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     11 * time.Second,
			utils.MetaCost:      10,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaStartTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaUsage:     11 * time.Second,
			utils.MetaCost:      10.5,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	// add new attribute profile
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_STATS",
			FilterIDs: []string{"*string:~*req.EventName:AddStatEvent"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "AcdMetric",
					Type:  utils.MetaVariable,
					Value: "~*stats.Stat_1.*acd",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_STATS"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithStat",
		Event: map[string]any{
			"EventName": "AddStatEvent",
		},
		APIOpts: map[string]any{},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_STATS",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "AcdMetric"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithStat",
			Event: map[string]any{
				"EventName": "AddStatEvent",
				"AcdMetric": "1.1E+10",
			},
			APIOpts: map[string]any{},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithStatFull(t *testing.T) {
	// add new attribute profile
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_STATS2",
			FilterIDs: []string{"*string:~*req.EventName:AddFullStats"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "AllMetrics",
					Type:  utils.MetaVariable,
					Value: "~*stats.Stat_1",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_STATS2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithStat",
		Event: map[string]any{
			"EventName": "AddFullStats",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_STATS2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "AllMetrics"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithStat",
			Event: map[string]any{
				"EventName":  "AddFullStats",
				"AllMetrics": "{\"*acd\":1.1E+10,\"*asr\":100,\"*tcd\":22000000000}",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithResource(t *testing.T) {
	//create a resourceProfile
	rlsConfig := &utils.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResTest",
		UsageTTL:          time.Minute,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Stored:            true,
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			}},
		ThresholdIDs: []string{utils.MetaNone},
	}

	var result string
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile, &utils.ResourceProfileWithAPIOpts{ResourceProfile: rlsConfig}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *utils.ResourceProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig), utils.ToJSON(reply))
	}

	// Allocate 3 units for resource ResTest
	argsRU := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "3001",
			"Destination": "3002"},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
			utils.OptsResourcesUnits:   3,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.ResourceSv1AllocateResources,
		argsRU, &result); err != nil {
		t.Error(err)
	}
	// Allocate 2 units for resource ResTest
	argsRU2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "3001",
			"Destination": "3002"},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e22",
			utils.OptsResourcesUnits:   2,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.ResourceSv1AllocateResources,
		argsRU2, &result); err != nil {
		t.Error(err)
	}

	// add new attribute profile
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_RESOURCE",
			FilterIDs: []string{"*string:~*req.EventName:AddResourceUsages"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "ResourceTotalUsages",
					Type:  utils.MetaVariable,
					Value: "~*resources.ResTest.TotalUsage",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_RESOURCE"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithResource",
		Event: map[string]any{
			"EventName": "AddResourceUsages",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_RESOURCE",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "ResourceTotalUsages"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithResource",
			Event: map[string]any{
				"EventName":           "AddResourceUsages",
				"ResourceTotalUsages": "5",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithResourceFull(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_RESOURCE2",
			FilterIDs: []string{"*string:~*req.EventName:AddFullResource"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "FullResource",
					Type:  utils.MetaVariable,
					Value: "~*resources.ResTest",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_RESOURCE2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithResource2",
		Event: map[string]any{
			"EventName": "AddFullResource",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_RESOURCE2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "FullResource"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithResource2",
			Event: map[string]any{
				"EventName":    "AddFullResource",
				"FullResource": "{\"Tenant\":\"cgrates.org\",\"ID\":\"ResTest\",\"Usages\":{\"651a8db2-4f67-4cf8-b622-169e8a482e21\":{\"Tenant\":\"cgrates.org\",\"ID\":\"651a8db2-4f67-4cf8-b622-169e8a482e21\",\"ExpiryTime\":\"2020-10-06T16:12:52.450804203+03:00\",\"Units\":3},\"651a8db2-4f67-4cf8-b622-169e8a482e22\":{\"Tenant\":\"cgrates.org\",\"ID\":\"651a8db2-4f67-4cf8-b622-169e8a482e22\",\"ExpiryTime\":\"2020-10-06T16:12:52.451034151+03:00\",\"Units\":2}},\"TTLIdx\":[\"651a8db2-4f67-4cf8-b622-169e8a482e21\",\"651a8db2-4f67-4cf8-b622-169e8a482e22\"]}",
			},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rplyEv.AlteredFields, eRply.AlteredFields) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply.AlteredFields), utils.ToJSON(rplyEv.AlteredFields))
	}
	// some fields are generated(e.g. Time) and compare only some part of the string
	strRes := utils.IfaceAsString(rplyEv.CGREvent.Event["FullResource"])
	if !strings.Contains(strRes, "{\"Tenant\":\"cgrates.org\",\"ID\":\"ResTest\",\"Usages\":{") {
		t.Errorf("Expecting: %s, received: %s",
			"{\"Tenant\":\"cgrates.org\",\"ID\":\"ResTest\",\"Usages\":{", strRes)
	} else if !strings.Contains(strRes, ",\"TTLIdx\":[") {
		t.Errorf("Expecting: %s, received: %s",
			",\"TTLIdx\":[", strRes)
	}
}

func testAttributeSProcessEventWithLibPhoneNumber(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_LIBPHONENUMBER2",
			FilterIDs: []string{"*string:~*req.EventName:AddDestinationCarrier"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "DestinationCarrier",
					Type:  utils.MetaVariable,
					Value: "~*libphonenumber.<~*req.Destination>.Carrier",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_LIBPHONENUMBER2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithLibPhoneNumber2",
		Event: map[string]any{
			"EventName":   "AddDestinationCarrier",
			"Destination": "+447779330921",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_LIBPHONENUMBER2",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "DestinationCarrier"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithLibPhoneNumber2",
			Event: map[string]any{
				"EventName":          "AddDestinationCarrier",
				"Destination":        "+447779330921",
				"DestinationCarrier": "Orange",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithLibPhoneNumberComposed(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_LIBPHONENUMBER_COMPOSED",
			FilterIDs: []string{"*string:~*req.EventName:AddComposedInfo"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "DestinationCarrier",
					Type:  utils.MetaComposed,
					Value: "~*libphonenumber.<~*req.Destination>.Carrier",
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "DestinationCarrier",
					Type:  utils.MetaComposed,
					Value: "~*libphonenumber.<~*req.Destination>.CountryCode",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_LIBPHONENUMBER_COMPOSED"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithLibPhoneNumberComposed",
		Event: map[string]any{
			"EventName":   "AddComposedInfo",
			"Destination": "+447779330921",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_LIBPHONENUMBER_COMPOSED",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "DestinationCarrier"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithLibPhoneNumberComposed",
			Event: map[string]any{
				"EventName":          "AddComposedInfo",
				"Destination":        "+447779330921",
				"DestinationCarrier": "Orange44",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithLibPhoneNumberFull(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &utils.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &utils.APIAttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_LIBPHONENUMBER",
			FilterIDs: []string{"*string:~*req.EventName:AddDestinationDetails"},
			Attributes: []*utils.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "DestinationDetails",
					Type:  utils.MetaVariable,
					Value: "~*libphonenumber.<~*req.Destination>",
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}
	if err := attrRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *utils.APIAttributeProfile
	if err := attrRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_LIBPHONENUMBER"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(alsPrf.APIAttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.APIAttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithLibPhoneNumber",
		Event: map[string]any{
			"EventName":   "AddDestinationDetails",
			"Destination": "+447779330921",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := attributes.AttrSProcessEventReply{
		AlteredFields: []*attributes.FieldsAltered{
			{
				MatchedProfileID: "cgrates.org:ATTR_LIBPHONENUMBER",
				Fields:           []string{utils.MetaReq + utils.NestingSep + "DestinationDetails"},
			},
		},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithLibPhoneNumber",
			Event: map[string]any{
				"EventName":          "AddDestinationDetails",
				"Destination":        "+447779330921",
				"DestinationDetails": "{\"Carrier\":\"Orange\",\"CountryCode\":44,\"CountryCodeSource\":1,\"Extension\":\"\",\"GeoLocation\":\"United Kingdom\",\"ItalianLeadingZero\":false,\"LengthOfNationalDestinationCode\":4,\"NationalNumber\":7779330921,\"NumberOfLeadingZeros\":1,\"NumberType\":1,\"PreferredDomesticCarrierCode\":\"\",\"RawInput\":\"+447779330921\",\"Region\":\"GB\"}",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var rplyEv attributes.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
