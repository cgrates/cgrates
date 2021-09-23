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
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrCfgPath     string
	attrCfg         *config.CGRConfig
	attrRPC         *rpc.Client
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
	switch *dbType {
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
	attrCfgPath = path.Join(*dataDir, "conf", "samples", alsPrfConfigDIR)
	attrCfg, err = config.NewCGRConfigFromPath(attrCfgPath)
	if err != nil {
		t.Error(err)
	}
	attrCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(attrCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(attrCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(attrCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrRPC, err = newRPCClient(attrCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAttributeSLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := attrRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testAttributeSProcessEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testAttributeSProcessEvent",
		Event: map[string]interface{}{
			utils.EventName: "VariableTest",
			utils.ToR:       utils.MetaVoice,
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_VARIABLE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + utils.Category},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]interface{}{
				utils.EventName: "VariableTest",
				utils.Category:  utils.MetaVoice,
				utils.ToR:       utils.MetaVoice,
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
			FilterIDs: []string{"*string:~*req.EventName:AddAccountInfo", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName": "AddAccountInfo",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_ACCOUNT"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Balance"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithAccount",
			Event: map[string]interface{}{
				"EventName": "AddAccountInfo",
				"Balance":   "10",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
			FilterIDs: []string{"*string:~*req.EventName:AddFullAccount", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName": "AddFullAccount",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_ACCOUNT2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "FullAccount"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithAccount2",
			Event: map[string]interface{}{
				"EventName":   "AddFullAccount",
				"FullAccount": "{\"ID\":\"cgrates.org:1001\",\"BalanceMap\":{\"*monetary\":[{\"Uuid\":\"18160631-a4ae-4078-8048-b4c6b87a36c6\",\"ID\":\"\",\"Value\":10,\"ExpirationDate\":\"0001-01-01T00:00:00Z\",\"Weight\":10,\"DestinationIDs\":{},\"RatingSubject\":\"\",\"Categories\":{},\"SharedGroups\":{},\"Disabled\":false,\"Factor\":null,\"Blocker\":false}]},\"UnitCounters\":null,\"ActionTriggers\":null,\"AllowNegative\":false,\"Disabled\":false,\"UpdateTime\":\"2020-10-06T12:43:51.805Z\"}",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
	ev1 := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        11 * time.Second,
				utils.Cost:         10.0,
			},
		},
	}
	if err := attrRPC.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1"}
	ev1.CGREvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        11 * time.Second,
			utils.Cost:         10.5,
		},
	}
	if err := attrRPC.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	// add new attribute profile
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_STATS",
			FilterIDs: []string{"*string:~*req.EventName:AddStatEvent", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName": "AddStatEvent",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_STATS"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "AcdMetric"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithStat",
			Event: map[string]interface{}{
				"EventName": "AddStatEvent",
				"AcdMetric": "11000000000",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
			FilterIDs: []string{"*string:~*req.EventName:AddFullStats", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName": "AddFullStats",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_STATS2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "AllMetrics"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithStat",
			Event: map[string]interface{}{
				"EventName":  "AddFullStats",
				"AllMetrics": "{\"*acd\":11000000000,\"*asr\":100,\"*tcd\":22000000000}",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
	if err := attrRPC.Call(utils.APIerSv1SetResourceProfile, &engine.ResourceProfileWithAPIOpts{ResourceProfile: rlsConfig}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *engine.ResourceProfile
	if err := attrRPC.Call(utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig), utils.ToJSON(reply))
	}

	// Allocate 3 units for resource ResTest
	argsRU := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]interface{}{
			"Account":     "3001",
			"Destination": "3002"},
		APIOpts: map[string]interface{}{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
			utils.OptsResourcesUnits:   3,
		},
	}
	if err := attrRPC.Call(utils.ResourceSv1AllocateResources,
		argsRU, &result); err != nil {
		t.Error(err)
	}
	// Allocate 2 units for resource ResTest
	argsRU2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]interface{}{
			"Account":     "3001",
			"Destination": "3002"},
		APIOpts: map[string]interface{}{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e22",
			utils.OptsResourcesUnits:   2,
		},
	}
	if err := attrRPC.Call(utils.ResourceSv1AllocateResources,
		argsRU2, &result); err != nil {
		t.Error(err)
	}

	// add new attribute profile
	alsPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_RESOURCE",
			FilterIDs: []string{"*string:~*req.EventName:AddResourceUsages", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName": "AddResourceUsages",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_RESOURCE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "ResourceTotalUsages"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithResource",
			Event: map[string]interface{}{
				"EventName":           "AddResourceUsages",
				"ResourceTotalUsages": "5",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
			FilterIDs: []string{"*string:~*req.EventName:AddFullResource", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName": "AddFullResource",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_RESOURCE2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "FullResource"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithResource2",
			Event: map[string]interface{}{
				"EventName":    "AddFullResource",
				"FullResource": "{\"Tenant\":\"cgrates.org\",\"ID\":\"ResTest\",\"Usages\":{\"651a8db2-4f67-4cf8-b622-169e8a482e21\":{\"Tenant\":\"cgrates.org\",\"ID\":\"651a8db2-4f67-4cf8-b622-169e8a482e21\",\"ExpiryTime\":\"2020-10-06T16:12:52.450804203+03:00\",\"Units\":3},\"651a8db2-4f67-4cf8-b622-169e8a482e22\":{\"Tenant\":\"cgrates.org\",\"ID\":\"651a8db2-4f67-4cf8-b622-169e8a482e22\",\"ExpiryTime\":\"2020-10-06T16:12:52.451034151+03:00\",\"Units\":2}},\"TTLIdx\":[\"651a8db2-4f67-4cf8-b622-169e8a482e21\",\"651a8db2-4f67-4cf8-b622-169e8a482e22\"]}",
			},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
			FilterIDs: []string{"*string:~*req.EventName:AddDestinationCarrier", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName":   "AddDestinationCarrier",
			"Destination": "+447779330921",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_LIBPHONENUMBER2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "DestinationCarrier"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithLibPhoneNumber2",
			Event: map[string]interface{}{
				"EventName":          "AddDestinationCarrier",
				"Destination":        "+447779330921",
				"DestinationCarrier": "Orange",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
			FilterIDs: []string{"*string:~*req.EventName:AddComposedInfo", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName":   "AddComposedInfo",
			"Destination": "+447779330921",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_LIBPHONENUMBER_COMPOSED"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "DestinationCarrier"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithLibPhoneNumberComposed",
			Event: map[string]interface{}{
				"EventName":          "AddComposedInfo",
				"Destination":        "+447779330921",
				"DestinationCarrier": "Orange;44",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
			FilterIDs: []string{"*string:~*req.EventName:AddDestinationDetails", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z"},
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
	if err := attrRPC.Call(utils.AdminSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.AdminSv1GetAttributeProfile,
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
		Event: map[string]interface{}{
			"EventName":   "AddDestinationDetails",
			"Destination": "+447779330921",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_LIBPHONENUMBER"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "DestinationDetails"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEventWithLibPhoneNumber",
			Event: map[string]interface{}{
				"EventName":          "AddDestinationDetails",
				"Destination":        "+447779330921",
				"DestinationDetails": "{\"Carrier\":\"Orange\",\"CountryCode\":44,\"CountryCodeSource\":1,\"Extension\":\"\",\"GeoLocation\":\"\",\"ItalianLeadingZero\":false,\"LengthOfNationalDestinationCode\":0,\"NationalNumber\":7779330921,\"NumberOfLeadingZeros\":1,\"NumberType\":1,\"PreferredDomesticCarrierCode\":\"\",\"RawInput\":\"+447779330921\",\"Region\":\"GB\"}",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	sort.Strings(eRply.AlteredFields)
	var rplyEv engine.AttrSProcessEventReply
	if err := attrRPC.Call(utils.AttributeSv1ProcessEvent,
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
