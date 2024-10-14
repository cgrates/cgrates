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
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrCfgPath     string
	attrCfg         *config.CGRConfig
	attrRPC         *birpc.Client
	alsPrfConfigDIR string
	sTestsAlsPrf    = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
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
		alsPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		alsPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		alsPrfConfigDIR = "tutmongo"
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
	attrCfg, err = config.NewCGRConfigFromPath(attrCfgPath)
	if err != nil {
		t.Error(err)
	}
	attrCfg.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(attrCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(attrCfg); err != nil {
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
	attrRPC = engine.NewRPCClient(t, attrCfg.ListenCfg())
}

func testAttributeSLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := attrRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
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
	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_VARIABLE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + utils.Category},
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
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithAccount(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_ACCOUNT",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddAccountInfo"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "Balance",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*accounts.1001.BalanceMap.*monetary[0].Value",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACCOUNT"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithAccount",
		Event: map[string]any{
			"EventName": "AddAccountInfo",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_ACCOUNT"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Balance"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithAccount",
			Event: map[string]any{
				"EventName": "AddAccountInfo",
				"Balance":   "10",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithAccountFull(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_ACCOUNT2",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddFullAccount"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "FullAccount",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*accounts.1001",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACCOUNT2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
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

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_ACCOUNT2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "FullAccount"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithAccount2",
			Event: map[string]any{
				"EventName":   "AddFullAccount",
				"FullAccount": "{\"ID\":\"cgrates.org:1001\",\"BalanceMap\":{\"*monetary\":[{\"Uuid\":\"18160631-a4ae-4078-8048-b4c6b87a36c6\",\"ID\":\"\",\"Value\":10,\"ExpirationDate\":\"0001-01-01T00:00:00Z\",\"Weight\":10,\"DestinationIDs\":{},\"RatingSubject\":\"\",\"Categories\":{},\"SharedGroups\":{},\"Timings\":null,\"TimingIDs\":{},\"Disabled\":false,\"Factor\":null,\"Blocker\":false}]},\"UnitCounters\":null,\"ActionTriggers\":null,\"AllowNegative\":false,\"Disabled\":false,\"UpdateTime\":\"2020-10-06T12:43:51.805Z\"}",
			},
			APIOpts: map[string]any{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rplyEv.MatchedProfiles, eRply.MatchedProfiles) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply.MatchedProfiles), utils.ToJSON(rplyEv.MatchedProfiles))
	} else if !reflect.DeepEqual(rplyEv.AlteredFields, eRply.AlteredFields) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply.AlteredFields), utils.ToJSON(rplyEv.AlteredFields))
	}
	// some fields are generated(e.g. BalanceID) and compare only some part of the string
	strAcc := utils.IfaceAsString(rplyEv.CGREvent.Event["FullAccount"])
	if !strings.Contains(strAcc, "\"ID\":\"cgrates.org:1001\"") {
		t.Errorf("Expecting: %s, received: %s",
			"\"ID\":\"cgrates.org:1001\"", strAcc)
	} else if !strings.Contains(strAcc, "\"UnitCounters\":null,\"ActionTriggers\":null,\"AllowNegative\":false,\"Disabled\":false") {
		t.Errorf("Expecting: %s, received: %s",
			"\"UnitCounters\":null,\"ActionTriggers\":null,\"AllowNegative\":false,\"Disabled\":false", strAcc)
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
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        11 * time.Second,
			utils.Cost:         10.0,
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
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        11 * time.Second,
			utils.Cost:         10.5,
		},
	}
	if err := attrRPC.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	// add new attribute profile
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_STATS",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddStatEvent"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "AcdMetric",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*stats.Stat_1.*acd",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	var result string
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_STATS"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEventWithStat",
		Event: map[string]any{
			"EventName": "AddStatEvent",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_STATS"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "AcdMetric"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithStat",
			Event: map[string]any{
				"EventName": "AddStatEvent",
				"AcdMetric": "11000000000",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithStatFull(t *testing.T) {
	// add new attribute profile
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_STATS2",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddFullStats"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "AllMetrics",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*stats.Stat_1",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	var result string
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_STATS2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
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

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_STATS2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "AllMetrics"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithStat",
			Event: map[string]any{
				"EventName":  "AddFullStats",
				"AllMetrics": "{\"*acd\":11000000000,\"*asr\":100,\"*tcd\":22000000000}",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithResource(t *testing.T) {
	//create a resourceProfile
	rlsConfig := &engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResTest",
		UsageTTL:          time.Minute,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{utils.MetaNone},
	}

	var result string
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetResourceProfile, &engine.ResourceProfileWithAPIOpts{ResourceProfile: rlsConfig}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *engine.ResourceProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetResourceProfile,
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
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_RESOURCE",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddResourceUsages"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "ResourceTotalUsages",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*resources.ResTest.TotalUsage",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_RESOURCE"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
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

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_RESOURCE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "ResourceTotalUsages"},
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
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithResourceFull(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_RESOURCE2",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddFullResource"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "FullResource",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*resources.ResTest",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_RESOURCE2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
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

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_RESOURCE2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "FullResource"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithResource2",
			Event: map[string]any{
				"EventName":    "AddFullResource",
				"FullResource": "{\"Tenant\":\"cgrates.org\",\"ID\":\"ResTest\",\"Usages\":{\"651a8db2-4f67-4cf8-b622-169e8a482e21\":{\"Tenant\":\"cgrates.org\",\"ID\":\"651a8db2-4f67-4cf8-b622-169e8a482e21\",\"ExpiryTime\":\"2020-10-06T16:12:52.450804203+03:00\",\"Units\":3},\"651a8db2-4f67-4cf8-b622-169e8a482e22\":{\"Tenant\":\"cgrates.org\",\"ID\":\"651a8db2-4f67-4cf8-b622-169e8a482e22\",\"ExpiryTime\":\"2020-10-06T16:12:52.451034151+03:00\",\"Units\":2}},\"TTLIdx\":[\"651a8db2-4f67-4cf8-b622-169e8a482e21\",\"651a8db2-4f67-4cf8-b622-169e8a482e22\"]}",
			},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rplyEv.MatchedProfiles, eRply.MatchedProfiles) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply.MatchedProfiles), utils.ToJSON(rplyEv.MatchedProfiles))
	} else if !reflect.DeepEqual(rplyEv.AlteredFields, eRply.AlteredFields) {
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
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_LIBPHONENUMBER2",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddDestinationCarrier"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "DestinationCarrier",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*libphonenumber.<~*req.Destination>.Carrier",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_LIBPHONENUMBER2"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
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

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_LIBPHONENUMBER2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "DestinationCarrier"},
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
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithLibPhoneNumberComposed(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_LIBPHONENUMBER_COMPOSED",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddComposedInfo"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "DestinationCarrier",
					Type: utils.MetaComposed,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*libphonenumber.<~*req.Destination>.Carrier",
						},
					},
				},
				{
					Path: utils.MetaReq + utils.NestingSep + "DestinationCarrier",
					Type: utils.MetaComposed,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: ";",
						},
					},
				},
				{
					Path: utils.MetaReq + utils.NestingSep + "DestinationCarrier",
					Type: utils.MetaComposed,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*libphonenumber.<~*req.Destination>.CountryCode",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_LIBPHONENUMBER_COMPOSED"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
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

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_LIBPHONENUMBER_COMPOSED"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "DestinationCarrier"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithLibPhoneNumberComposed",
			Event: map[string]any{
				"EventName":          "AddComposedInfo",
				"Destination":        "+447779330921",
				"DestinationCarrier": "Orange;44",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, rplyEv) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithLibPhoneNumberFull(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_LIBPHONENUMBER",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"*string:~*req.EventName:AddDestinationDetails"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + "DestinationDetails",
					Type: utils.MetaVariable,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "~*libphonenumber.<~*req.Destination>",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	alsPrf.Compile()
	if err := attrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_LIBPHONENUMBER"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
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

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_LIBPHONENUMBER"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "DestinationDetails"},
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
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	}
	sort.Strings(rplyEv.AlteredFields)
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

func TestAttributesDestinationFilters(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"attributes": {
	"enabled": true
},

"filters": {			
	"apiers_conns": ["*localhost"]
},

"apiers": {
	"enabled": true,
}

}`

	tpFiles := map[string]string{
		utils.AttributesCsv: `#Tenant,ID,Context,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_INLINE_FILTER,,*string:~*req.Account:1001,,,,,,,30
cgrates.org,ATTR_INLINE_FILTER,,,,*destinations:~*req.Destination:1002,*req.InlinePrefixCase,*constant,shouldnotmatch,,
cgrates.org,ATTR_INLINE_FILTER,,,,*destinations:~*req.Destination:DST_20,*req.InlineWrongDestination,*constant,shouldnotmatch,,
cgrates.org,ATTR_INLINE_FILTER,,,,*destinations:~*req.Destination:DST_20|DST_10,*req.InlineOrDestinationMatch,*constant,shouldmatch,,
cgrates.org,ATTR_INLINE_FILTER,,,,*destinations:~*req.Destination:DST_10,*req.InlineDestinationMatch,*constant,shouldmatch,,
cgrates.org,ATTR_PREDEFINED_FILTER,,*string:~*req.Account:2001,,,,,,,
cgrates.org,ATTR_PREDEFINED_FILTER,,,,FLTR_DESTINATION_DIRECT,*req.PredefinedPrefixCase,*constant,shouldnotmatch,,
cgrates.org,ATTR_PREDEFINED_FILTER,,,,FLTR_WRONG_DESTINATION,*req.PredefinedWrongDestination,*constant,shouldnotmatch,,
cgrates.org,ATTR_PREDEFINED_FILTER,,,,FLTR_OR_DESTINATION_MATCH,*req.PredefinedOrDestinationMatch,*constant,shouldmatch,,
cgrates.org,ATTR_PREDEFINED_FILTER,,,,FLTR_DESTINATION_MATCH,*req.PredefinedDestinationMatch,*constant,shouldmatch,,`,
		utils.DestinationsCsv: `#Id,Prefix
DST_10,10
DST_20,20`,
		utils.FiltersCsv: `#Tenant[0],ID[1],Type[2],Path[3],Values[4],ActivationInterval[5]
cgrates.org,FLTR_DESTINATION_DIRECT,*destinations,~*req.Destination,1002,
cgrates.org,FLTR_WRONG_DESTINATION,*destinations,~*req.Destination,DST_20,
cgrates.org,FLTR_OR_DESTINATION_MATCH,*destinations,~*req.Destination,DST_20;DST_10,
cgrates.org,FLTR_DESTINATION_MATCH,*destinations,~*req.Destination,DST_10,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	t.Run("SetAttributesThroughAPI", func(t *testing.T) {
		var reply string
		attrPrf := &engine.AttributeProfileWithAPIOpts{
			AttributeProfile: &engine.AttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_API",
				FilterIDs: []string{"*string:~*req.Account:3001"},
				Contexts:  []string{utils.MetaAny},
				Attributes: []*engine.Attribute{
					{
						FilterIDs: []string{"FLTR_DESTINATION_DIRECT", "*destinations:~*req.Destination:1002"},
						Path:      "*req.PrefixCase",
						Type:      utils.MetaConstant,
						Value:     config.NewRSRParsersMustCompile("shouldnotmatch", utils.InfieldSep),
					},
					{
						FilterIDs: []string{"FLTR_WRONG_DESTINATION", "*destinations:~*req.Destination:DST_20"},
						Path:      "*req.WrongDestination",
						Type:      utils.MetaConstant,
						Value:     config.NewRSRParsersMustCompile("shouldnotmatch", utils.InfieldSep),
					},
					{
						FilterIDs: []string{"FLTR_OR_DESTINATION_MATCH", "*destinations:~*req.Destination:DST_20|DST_10"},
						Path:      "*req.OrDestinationMatch",
						Type:      utils.MetaConstant,
						Value:     config.NewRSRParsersMustCompile("shouldmatch", utils.InfieldSep),
					},
					{
						FilterIDs: []string{"FLTR_DESTINATION_MATCH", "*destinations:~*req.Destination:DST_10"},
						Path:      "*req.DestinationMatch",
						Type:      utils.MetaConstant,
						Value:     config.NewRSRParsersMustCompile("shouldmatch", utils.InfieldSep),
					},
				},
				Weight: 10,
			},
		}
		attrPrf.Compile()
		if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckFieldsAlteredByAttributeS", func(t *testing.T) {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event_test",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
			},
		}
		expected := engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"cgrates.org:ATTR_INLINE_FILTER"},
			AlteredFields:   []string{"*req.InlineDestinationMatch", "*req.InlineOrDestinationMatch"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event_test",
				Event: map[string]any{
					utils.AccountField:         "1001",
					utils.Destination:          "1002",
					"InlineDestinationMatch":   "shouldmatch",
					"InlineOrDestinationMatch": "shouldmatch",
				},
				APIOpts: make(map[string]any),
			},
		}

		var reply engine.AttrSProcessEventReply
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &reply); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expected.AlteredFields)
			sort.Strings(reply.AlteredFields)
			if !reflect.DeepEqual(expected, reply) {
				t.Errorf("\nexpected: %s, \nreceived: %s",
					utils.ToJSON(expected), utils.ToJSON(reply))
			}
		}

		ev.Event = map[string]any{
			utils.AccountField: "2001",
			utils.Destination:  "1002",
		}
		expected.MatchedProfiles = []string{"cgrates.org:ATTR_PREDEFINED_FILTER"}
		expected.AlteredFields = []string{"*req.PredefinedDestinationMatch", "*req.PredefinedOrDestinationMatch"}
		expected.CGREvent.Event = map[string]any{
			utils.AccountField:             "2001",
			utils.Destination:              "1002",
			"PredefinedDestinationMatch":   "shouldmatch",
			"PredefinedOrDestinationMatch": "shouldmatch",
		}

		reply = engine.AttrSProcessEventReply{}
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &reply); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expected.AlteredFields)
			sort.Strings(reply.AlteredFields)
			if !reflect.DeepEqual(expected, reply) {
				t.Errorf("\nexpected: %s, \nreceived: %s",
					utils.ToJSON(expected), utils.ToJSON(reply))
			}
		}

		ev.Event = map[string]any{
			utils.AccountField: "3001",
			utils.Destination:  "1002",
		}
		expected.MatchedProfiles = []string{"cgrates.org:ATTR_API"}
		expected.AlteredFields = []string{"*req.DestinationMatch", "*req.OrDestinationMatch"}
		expected.CGREvent.Event = map[string]any{
			utils.AccountField:   "3001",
			utils.Destination:    "1002",
			"DestinationMatch":   "shouldmatch",
			"OrDestinationMatch": "shouldmatch",
		}

		reply = engine.AttrSProcessEventReply{}
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &reply); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expected.AlteredFields)
			sort.Strings(reply.AlteredFields)
			if !reflect.DeepEqual(expected, reply) {
				t.Errorf("\nexpected: %s, \nreceived: %s",
					utils.ToJSON(expected), utils.ToJSON(reply))
			}
		}
	})
}

func TestAttributesArith(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := `{

"general": {
	"log_level": 7,
},

"data_db": {								
	"db_type": "*internal"
},

"stor_db": {
	"db_type": "*internal"
},

"attributes": {
	"enabled": true
},

"apiers": {
	"enabled": true,
}

}`

	tpFiles := map[string]string{
		utils.AttributesCsv: `#Tenant,ID,Context,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_ARITH,,*string:~*req.AttrSource:csv,,,,,,,
cgrates.org,ATTR_ARITH,,,,,*req.3*4,*multiply,3;4,,
cgrates.org,ATTR_ARITH,,,,,*req.12/4,*divide,12;4,,
cgrates.org,ATTR_ARITH,,,,,*req.3+4,*sum,3;4,,
cgrates.org,ATTR_ARITH,,,,,*req.3-4,*difference,3;4,,
cgrates.org,ATTR_ARITH,,,,,*req.MultiplyBetweenVariables,*multiply,~*req.Elem1;~*req.Elem2,,`,
	}

	ng := engine.TestEngine{
		ConfigJSON: content,
		TpFiles:    tpFiles,
	}
	client, _ := ng.Run(t)

	t.Run("SetAttributesThroughAPI", func(t *testing.T) {
		var reply string
		attrPrf := &engine.AttributeProfileWithAPIOpts{
			AttributeProfile: &engine.AttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_API",
				FilterIDs: []string{"*string:~*req.AttrSource:api"},
				Contexts:  []string{utils.MetaAny},
				Attributes: []*engine.Attribute{
					{
						Path:  "*req.3*4",
						Type:  utils.MetaMultiply,
						Value: config.NewRSRParsersMustCompile("3;4", utils.InfieldSep),
					},
					{
						Path:  "*req.12/4",
						Type:  utils.MetaDivide,
						Value: config.NewRSRParsersMustCompile("12;4", utils.InfieldSep),
					},
					{
						Path:  "*req.3+4",
						Type:  utils.MetaSum,
						Value: config.NewRSRParsersMustCompile("3;4", utils.InfieldSep),
					},
					{
						Path:  "*req.3-4",
						Type:  utils.MetaDifference,
						Value: config.NewRSRParsersMustCompile("3;4", utils.InfieldSep),
					},
					{
						Path:  "*req.MultiplyBetweenVariables",
						Type:  utils.MetaMultiply,
						Value: config.NewRSRParsersMustCompile("~*req.Elem1;~*req.Elem2", utils.InfieldSep),
					},
				},
				Weight: 10,
			},
		}
		attrPrf.Compile()
		if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &reply); err != nil {
			t.Error(err)
		}
	})

	t.Run("CheckFieldsAlteredByAttributeS", func(t *testing.T) {
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event_test",
			Event: map[string]any{
				"AttrSource": "csv",
				"Elem1":      "3",
				"Elem2":      "4",
			},
		}
		expected := engine.AttrSProcessEventReply{
			MatchedProfiles: []string{"cgrates.org:ATTR_ARITH"},
			AlteredFields:   []string{"*req.12/4", "*req.3*4", "*req.3+4", "*req.3-4", "*req.MultiplyBetweenVariables"},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event_test",
				Event: map[string]any{
					"AttrSource":               "csv",
					"Elem1":                    "3",
					"Elem2":                    "4",
					"12/4":                     "3",
					"3*4":                      "12",
					"3+4":                      "7",
					"3-4":                      "-1",
					"MultiplyBetweenVariables": "12",
				},
				APIOpts: make(map[string]any),
			},
		}

		var reply engine.AttrSProcessEventReply
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &reply); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expected.AlteredFields)
			sort.Strings(reply.AlteredFields)
			if !reflect.DeepEqual(expected, reply) {
				t.Errorf("\nexpected: %s, \nreceived: %s",
					utils.ToJSON(expected), utils.ToJSON(reply))
			}
		}

		ev.Event["AttrSource"] = "api"
		expected.MatchedProfiles = []string{"cgrates.org:ATTR_API"}
		expected.AlteredFields = []string{"*req.12/4", "*req.3*4", "*req.3+4", "*req.3-4", "*req.MultiplyBetweenVariables"}
		expected.CGREvent.Event = map[string]any{
			"AttrSource":               "api",
			"Elem1":                    "3",
			"Elem2":                    "4",
			"12/4":                     "3",
			"3*4":                      "12",
			"3+4":                      "7",
			"3-4":                      "-1",
			"MultiplyBetweenVariables": "12",
		}

		reply = engine.AttrSProcessEventReply{}
		if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
			ev, &reply); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expected.AlteredFields)
			sort.Strings(reply.AlteredFields)
			if !reflect.DeepEqual(expected, reply) {
				t.Errorf("\nexpected: %s, \nreceived: %s",
					utils.ToJSON(expected), utils.ToJSON(reply))
			}
		}
	})
}
