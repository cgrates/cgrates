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
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	attrCfgPath     string
	attrCfg         *config.CGRConfig
	attrRPC         *rpc.Client
	attrDataDir     = "/usr/share/cgrates"
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
		testAttributeSProcessEventWithStat,
		testAttributeSProcessEventWithResource,
		testAttributeSProcessEventWithLibPhoneNumber,
		testAttributeSStopEngine,
	}
)

func TestAttributeSIT(t *testing.T) {
	attrsTests := sTestsAlsPrf
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
	for _, stest := range attrsTests {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAttributeSInitCfg(t *testing.T) {
	var err error
	attrCfgPath = path.Join(attrDataDir, "conf", "samples", alsPrfConfigDIR)
	attrCfg, err = config.NewCGRConfigFromPath(attrCfgPath)
	if err != nil {
		t.Error(err)
	}
	attrCfg.DataFolderPath = attrDataDir // Share DataFolderPath through config towards StoreDb for Flush()
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
	time.Sleep(500 * time.Millisecond)
}
func testAttributeSProcessEvent(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]interface{}{
					utils.EVENT_NAME: "VariableTest",
					utils.ToR:        utils.VOICE,
				},
			},
		},
	}
	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_VARIABLE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + utils.Category},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]interface{}{
					utils.EVENT_NAME: "VariableTest",
					utils.Category:   utils.VOICE,
					utils.ToR:        utils.VOICE,
				},
			},
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
	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_ACCOUNT",
			Contexts:  []string{utils.META_ANY},
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
	if err := attrRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_ACCOUNT"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
	}

	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithAccount",
				Event: map[string]interface{}{
					"EventName": "AddAccountInfo",
				},
			},
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACCOUNT"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Balance"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithAccount",
				Event: map[string]interface{}{
					"EventName": "AddAccountInfo",
					"Balance":   "10",
				},
			},
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

func testAttributeSProcessEventWithStat(t *testing.T) {
	// simulate some stat event
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := &engine.StatsArgsProcessEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]interface{}{
					utils.Account:    "1001",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:      time.Duration(11 * time.Second),
					utils.COST:       10.0,
				},
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
			utils.Account:    "1001",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       10.5,
		},
	}
	if err := attrRPC.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	// add new attribute profile
	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_STATS",
			Contexts:  []string{utils.META_ANY},
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
	if err := attrRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_STATS"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
	}

	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithStat",
				Event: map[string]interface{}{
					"EventName": "AddStatEvent",
				},
			},
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_STATS"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "AcdMetric"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithStat",
				Event: map[string]interface{}{
					"EventName": "AddStatEvent",
					"AcdMetric": "11",
				},
			},
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
		UsageTTL:          time.Duration(1) * time.Minute,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{utils.META_NONE},
	}

	var result string
	if err := attrRPC.Call(utils.APIerSv1SetResourceProfile, &v1.ResourceWithCache{ResourceProfile: rlsConfig}, &result); err != nil {
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
	argsRU := utils.ArgRSv1ResourceUsage{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					"Account":     "3001",
					"Destination": "3002"},
			},
		},
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
		Units:   3,
	}
	if err := attrRPC.Call(utils.ResourceSv1AllocateResources,
		argsRU, &result); err != nil {
		t.Error(err)
	}
	// Allocate 2 units for resource ResTest
	argsRU2 := utils.ArgRSv1ResourceUsage{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					"Account":     "3001",
					"Destination": "3002"},
			},
		},
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e22",
		Units:   2,
	}
	if err := attrRPC.Call(utils.ResourceSv1AllocateResources,
		argsRU2, &result); err != nil {
		t.Error(err)
	}

	// add new attribute profile
	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_RESOURCE",
			Contexts:  []string{utils.META_ANY},
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
	if err := attrRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_RESOURCE"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
	}

	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithResource",
				Event: map[string]interface{}{
					"EventName": "AddResourceUsages",
				},
			},
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_RESOURCE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "ResourceTotalUsages"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithResource",
				Event: map[string]interface{}{
					"EventName":           "AddResourceUsages",
					"ResourceTotalUsages": "5",
				},
			},
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

func testAttributeSProcessEventWithLibPhoneNumber(t *testing.T) {
	// add new attribute profile
	var result string
	alsPrf := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_LIBPHONENUMBER",
			Contexts:  []string{utils.META_ANY},
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
	if err := attrRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var replyAttr *engine.AttributeProfile
	if err := attrRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_LIBPHONENUMBER"}}, &replyAttr); err != nil {
		t.Fatal(err)
	}
	replyAttr.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, replyAttr) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, replyAttr)
	}

	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithLibPhoneNumber",
				Event: map[string]interface{}{
					"EventName":   "AddDestinationDetails",
					"Destination": "+447779330921",
				},
			},
		},
	}

	eRply := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_LIBPHONENUMBER"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "DestinationDetails"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventWithLibPhoneNumber",
				Event: map[string]interface{}{
					"EventName":          "AddDestinationDetails",
					"Destination":        "+447779330921",
					"DestinationDetails": "{\"Carrier\":\"Orange\",\"CountryCode\":44,\"GeoLocation\":\"\",\"NationalNumber\":7779330921,\"NumberType\":1,\"Region\":\"GB\"}",
				},
			},
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
