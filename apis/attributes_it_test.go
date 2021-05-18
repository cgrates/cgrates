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

package apis

import (
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	alsPrfCfgPath   string
	alsPrfCfg       *config.CGRConfig
	attrSRPC        *birpc.Client
	alsPrfConfigDIR string //run tests for specific configuration

	sTestsAlsPrf = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testGetAttributeProfileBeforeSet,
		//testAttributeSLoadFromFolder,
		testAttributeSetAttributeProfile,
		testAttributeGetAttributeIDs,
		testAttributeGetAttributeIDsCount,
		testGetAttributeProfileBeforeSet2,
		testAttributeSetAttributeProfile2,
		testAttributeGetAttributeIDs2,
		testAttributeGetAttributeIDsCount2,
		testAttributeRemoveAttributeProfile,
		testAttributeGetAttributeIDs,
		testAttributeGetAttributeIDsCount,
		testAttributeSetAttributeProfileBrokenReference,
		testAttributeSGetAttributeForEventMissingEvent,
		testAttributeSGetAttributeForEventAnyContext,
		testAttributeSGetAttributeForEventSameAnyContext,
		testAttributeSGetAttributeForEventNotFound,
		testAttributeSGetAttributeForEvent,
		testAttributeProcessEvent,
		testAttributeProcessEventWithSearchAndReplace,
		testAttributeSProcessWithMultipleRuns,
		testAttributeSProcessWithMultipleRuns2,
		testAttributeGetAttributeProfileAllIDs,
		testAttributeGetAttributeProfileAllIDsCount,
		testAttributeRemoveRemainAttributeProfiles,
		testAttributeGetAttributeProfileAfterRemove,
		testAttributeSKillEngine,
	}
)

func TestAttributeSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		alsPrfConfigDIR = "attributes_internal"
	case utils.MetaMongo:
		alsPrfConfigDIR = "attributes_mongo"
	case utils.MetaMySQL:
		alsPrfConfigDIR = "attributes_mysql"
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
	alsPrfCfgPath = path.Join(*dataDir, "conf", "samples", alsPrfConfigDIR)
	alsPrfCfg, err = config.NewCGRConfigFromPath(alsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(alsPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrSRPC, err = newRPCClient(alsPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testGetAttributeProfileBeforeSet(t *testing.T) {
	var reply *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := attrSRPC.Call(context.Background(),
		utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAttributeSetAttributeProfile(t *testing.T) {
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	var reply string
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAttr := &engine.APIAttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_IT_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: "1002",
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: "cgrates.itsyscom",
			},
		},
	}
	var result *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAttr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedAttr), utils.ToJSON(result))
	}
}

func testAttributeGetAttributeIDs(t *testing.T) {
	var reply []string
	args := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_ATTRIBUTES_IT_TEST"}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testAttributeGetAttributeIDsCount(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDsCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected %+v \n, received %+v", 1, reply)
	}
}

func testGetAttributeProfileBeforeSet2(t *testing.T) {
	var reply *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST_SECOND",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSetAttributeProfile2(t *testing.T) {
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST_SECOND",
			FilterIDs: []string{"*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	var reply string
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAttr := &engine.APIAttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_IT_TEST_SECOND",
		FilterIDs: []string{"*string:~*opts.*context:*sessions"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: "cgrates.itsyscom",
			},
		},
	}
	var result *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST_SECOND",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAttr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedAttr), utils.ToJSON(result))
	}
}

func testAttributeGetAttributeIDs2(t *testing.T) {
	var reply []string
	args := &utils.PaginatorWithTenant{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_ATTRIBUTES_IT_TEST", "TEST_ATTRIBUTES_IT_TEST_SECOND"}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testAttributeGetAttributeIDsCount2(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDsCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("Expected %+v \n, received %+v", 2, reply)
	}
}

func testAttributeRemoveAttributeProfile(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "TEST_ATTRIBUTES_IT_TEST_SECOND",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	//nothing to get from db
	var result *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST_SECOND",
			},
		}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v \n, received %+v", utils.ErrNotFound, err)
	}
}

func testAttributeSetAttributeProfileBrokenReference(t *testing.T) {
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST_SECOND",
			FilterIDs: []string{"invalid_filter_format", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	var reply string
	expectedErr := "SERVER_ERROR: broken reference to filter: invalid_filter_format for item with ID: cgrates.org:TEST_ATTRIBUTES_IT_TEST_SECOND"
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v \n, received %+v", expectedErr, err)
	}
}

func testAttributeSGetAttributeForEventMissingEvent(t *testing.T) {
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		&engine.AttrArgsProcessEvent{}, &rplyEv); err == nil ||
		err.Error() != "MANDATORY_IE_MISSING: [CGREvent]" {
		t.Error(err)
	}
}

func testAttributeSGetAttributeForEventAnyContext(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
			Event: map[string]interface{}{
				utils.AccountField: "dan",
				utils.Destination:  "+491511231234",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaCDRs,
			},
		},
	}
	eAttrPrf2 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Account:dan", "*ai:~*req.AnswerTime:2014-01-14T00:00:00Z"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1001",
				},
			},
			Weight: 10.0,
		},
	}
	var result string
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		eAttrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_2"}}, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eAttrPrf2.APIAttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", eAttrPrf2.APIAttributeProfile, reply)
	}
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}

	expAttrFromEv := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:dan", "*ai:~*req.AnswerTime:2014-01-14T00:00:00Z"},
		Attributes: []*engine.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weight: 10.0,
	}
	expAttrFromEv.Compile()
	attrReply.Compile()
	if !reflect.DeepEqual(expAttrFromEv, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(expAttrFromEv), utils.ToJSON(attrReply))
	}
}

func testAttributeSGetAttributeForEventSameAnyContext(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
			Event: map[string]interface{}{
				utils.AccountField: "dan",
				utils.Destination:  "+491511231234",
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}

	expAttrFromEv := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:dan", "*ai:~*req.AnswerTime:2014-01-14T00:00:00Z"},
		Attributes: []*engine.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weight: 10.0,
	}
	expAttrFromEv.Compile()
	attrReply.Compile()
	if !reflect.DeepEqual(expAttrFromEv, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(expAttrFromEv), utils.ToJSON(attrReply))
	}
}

func testAttributeSGetAttributeForEventNotFound(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
			Event: map[string]interface{}{
				utils.AccountField: "dann",
				utils.Destination:  "+491511231234",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaCDRs,
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSGetAttributeForEvent(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.AccountField: "1007",
				utils.Destination:  "+491511231234",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}

	eAttrPrf := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1007", "*ai:~*req.AnswerTime:2014-01-14T00:00:00Z", "*string:~*opts.*context:*cdrs|*sessions"},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value:     config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
				Type:      utils.MetaConstant,
				FilterIDs: []string{},
			},
			{
				Path:      utils.MetaReq + utils.NestingSep + utils.Subject,
				Value:     config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
				Type:      utils.MetaConstant,
				FilterIDs: []string{},
			},
		},
		Weight: 10.0,
	}
	if *encoding == utils.MetaGOB {
		eAttrPrf.Attributes[0].FilterIDs = nil
		eAttrPrf.Attributes[1].FilterIDs = nil
	}
	var result string
	eAttrPrfApi := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_1",
			FilterIDs: []string{"*string:~*req.Account:1007", "*ai:~*req.AnswerTime:2014-01-14T00:00:00Z", "*string:~*opts.*context:*cdrs|*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:      utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value:     "1001",
					Type:      utils.MetaConstant,
					FilterIDs: []string{},
				},
				{
					Path:      utils.MetaReq + utils.NestingSep + utils.Subject,
					Value:     "1001",
					Type:      utils.MetaConstant,
					FilterIDs: []string{},
				},
			},
			Weight: 10.0,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		eAttrPrfApi, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}
	attrReply.Compile()
	eAttrPrf.Compile()
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	ev.Tenant = utils.EmptyString
	ev.ID = "randomID"
	var attrPrf *engine.AttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1GetAttributeForEvent,
		ev, &attrPrf); err != nil {
		t.Fatal(err)
	}

	attrPrf.Compile()
	// Populate private variables in RSRParsers
	if !reflect.DeepEqual(eAttrPrf, attrPrf) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attrPrf))
	}
}

func testAttributeProcessEvent(t *testing.T) {
	var reply *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}}, &reply); err != nil {
		t.Fatal(err)
	}
	args := &engine.AttrArgsProcessEvent{
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Event: map[string]interface{}{
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1002",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaCDRs,
			},
		},
	}
	expEvReply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"TEST_ATTRIBUTES_IT_TEST"},
		AlteredFields:   []string{"*tenant", utils.AccountField},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.itsyscom",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.ToR:          utils.MetaVoice,
			},
			APIOpts: map[string]interface{}{},
		},
	}
	evRply := &engine.AttrSProcessEventReply{}
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		args, &evRply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expEvReply.AlteredFields)
		sort.Strings(evRply.AlteredFields)
		if !reflect.DeepEqual(evRply, expEvReply) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expEvReply), utils.ToJSON(evRply))
		}
	}
}

func testAttributeProcessEventWithSearchAndReplace(t *testing.T) {
	attrPrf1 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_Search_and_replace",
			FilterIDs: []string{"*string:~*req.Category:call", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Category",
					Value: "~*req.Category:s/(.*)/${1}_suffix/",
				},
			},
			Blocker: true,
			Weight:  10,
		},
	}
	var result string
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "HeaderEventForAttribute",
			Event: map[string]interface{}{
				"Category": "call",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_Search_and_replace"},
		AlteredFields:   []string{"*req.Category"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "HeaderEventForAttribute",
			Event: map[string]interface{}{
				"Category": "call_suffix",
			},
			APIOpts: map[string]interface{}{},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessWithMultipleRuns(t *testing.T) {
	attrPrf1 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: "Value1",
				},
			},
			Weight: 10,
		},
	}
	attrPrf2 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: "Value2",
				},
			},
			Weight: 20,
		},
	}
	attrPrf3 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_3",
			FilterIDs: []string{"*string:~*req.NotFound:NotFound", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Value: "Value3",
				},
			},
			Weight: 30,
		},
	}
	// Add attribute in DM
	var result string
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		ProcessRuns: utils.IntPointer(4),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2", "ATTR_1", "ATTR_2"},
		AlteredFields:   []string{"*req.Field1", "*req.Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(eRply.MatchedProfiles, rplyEv.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, rplyEv.MatchedProfiles)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, rplyEv.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, rplyEv.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent.Event, rplyEv.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, rplyEv.CGREvent.Event)
	}
}

func testAttributeSProcessWithMultipleRuns2(t *testing.T) {
	attrPrf1 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: "Value1",
				},
			},
			Weight: 10,
		},
	}
	attrPrf2 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: "Value2",
				},
			},
			Weight: 20,
		},
	}
	attrPrf3 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_3",
			FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Value: "Value3",
				},
			},
			Weight: 30,
		},
	}
	// Add attributeProfiles
	var result string
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		ProcessRuns: utils.IntPointer(4),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2", "ATTR_3", "ATTR_2"},
		AlteredFields:   []string{"*req.Field1", "*req.Field2", "*req.Field3"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
				"Field3":       "Value3",
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(eRply.MatchedProfiles, rplyEv.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, rplyEv.MatchedProfiles)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, rplyEv.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, rplyEv.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent.Event, rplyEv.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, rplyEv.CGREvent.Event)
	}
}

func testAttributeGetAttributeProfileAllIDs(t *testing.T) {
	var rply []string
	expectedIds := []string{"ATTR_1", "ATTR_2", "ATTR_3", "ATTR_Search_and_replace",
		"TEST_ATTRIBUTES_IT_TEST"} //"TEST_ATTRIBUTES_IT_TEST_SECOND"}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant: "cgrates.org",
		}, &rply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(rply)
		sort.Strings(expectedIds)
		if !reflect.DeepEqual(expectedIds, rply) {
			t.Errorf("Expected %+v, received %+v", expectedIds, rply)
		}
	}
}

func testAttributeGetAttributeProfileAllIDsCount(t *testing.T) {
	var rply int
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDsCount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
			},
		}, &rply); err != nil {
		t.Error(err)
	} else if rply != 5 { //6 for internal
		t.Errorf("Expected %+v, received %+v", 5, rply)
	}
}

func testAttributeRemoveRemainAttributeProfiles(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_1",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_2",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_3",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	args = &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "ATTR_Search_and_replace",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
}

func testAttributeGetAttributeProfileAfterRemove(t *testing.T) {
	var reply *engine.APIAttributeProfile
	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_Search_and_replace",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_1",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_2",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := attrSRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ATTR_3",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//Kill the engine when it is about to be finished
func testAttributeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
