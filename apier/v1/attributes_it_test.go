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
	alsPrfCfgPath   string
	alsPrfCfg       *config.CGRConfig
	attrSRPC        *rpc.Client
	alsPrfDataDir   = "/usr/share/cgrates"
	alsPrf          *AttributeWithCache
	alsPrfConfigDIR string //run tests for specific configuration

	sTestsAlsPrf = []func(t *testing.T){
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testAttributeSLoadFromFolder,
		testAttributeSGetAttributeForEvent,
		testAttributeSGetAttributeForEventNotFound,
		testAttributeSGetAttributeForEventWithMetaAnyContext,
		testAttributeSProcessEvent,
		testAttributeSProcessEventNotFound,
		testAttributeSProcessEventMissing,
		testAttributeSProcessEventWithNoneSubstitute,
		testAttributeSProcessEventWithNoneSubstitute2,
		testAttributeSProcessEventWithNoneSubstitute3,
		testAttributeSProcessEventWithHeader,
		testAttributeSGetAttPrfIDs,
		testAttributeSSetAlsPrfBrokenReference,
		testAttributeSGetAlsPrfBeforeSet,
		testAttributeSSetAlsPrf,
		testAttributeSUpdateAlsPrf,
		testAttributeSRemAlsPrf,
		testAttributeSSetAlsPrf2,
		testAttributeSSetAlsPrf3,
		testAttributeSSetAlsPrf4,
		testAttributeSPing,
		testAttributeSProcessEventWithSearchAndReplace,
		testAttributeSProcessWithMultipleRuns,
		testAttributeSProcessWithMultipleRuns2,
		testAttributeSGetAttributeProfileIDsCount,
		testAttributeSSetAttributeWithEmptyPath,
		testAttributeSSetAlsPrfWithoutTenant,
		testAttributeSRmvAlsPrfWithoutTenant,
		testAttributeSKillEngine,
		//start test for cache options
		testAttributeSInitCfg,
		testAttributeSInitDataDb,
		testAttributeSResetStorDb,
		testAttributeSStartEngine,
		testAttributeSRPCConn,
		testAttributeSCachingMetaNone,
		testAttributeSCachingMetaLoad,
		testAttributeSCachingMetaReload1,
		testAttributeSCachingMetaReload2,
		testAttributeSCachingMetaRemove,
		testAttributeSCacheOpts,
		testAttributeSKillEngine,
	}
)

//Test start here
func TestAttributeSIT(t *testing.T) {
	attrsTests := sTestsAlsPrf
	switch *dbType {
	case utils.MetaInternal:
		attrsTests = sTestsAlsPrf[:29]
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
	alsPrfCfgPath = path.Join(alsPrfDataDir, "conf", "samples", alsPrfConfigDIR)
	alsPrfCfg, err = config.NewCGRConfigFromPath(alsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAttributeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAttributeSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAttributeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(alsPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrSRPC, err = newRPCClient(alsPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAttributeSGetAlsPrfBeforeSet(t *testing.T) {
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := attrSRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAttributeSGetAttributeForEvent(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEvent",
				Event: map[string]interface{}{
					utils.Account:     "1007",
					utils.Destination: "+491511231234",
				},
			},
		},
	}

	eAttrPrf := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1007"},
		Contexts:  []string{utils.MetaCDRs, utils.MetaSessionS},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			{
				Path:      utils.MetaReq + utils.NestingSep + utils.Account,
				Value:     config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				Type:      utils.META_CONSTANT,
				FilterIDs: []string{},
			},
			{
				Path:      utils.MetaReq + utils.NestingSep + utils.Subject,
				Value:     config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				Type:      utils.META_CONSTANT,
				FilterIDs: []string{},
			},
		},
		Weight: 10.0,
	}
	if *encoding == utils.MetaGOB {
		eAttrPrf.Attributes[0].FilterIDs = nil
		eAttrPrf.Attributes[1].FilterIDs = nil
	}
	eAttrPrf.Compile()
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}
	attrReply.Compile() // Populate private variables in RSRParsers
	sort.Strings(attrReply.Contexts)
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	ev.Tenant = utils.EmptyString
	ev.ID = "randomID"
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}
	attrReply.Compile() // Populate private variables in RSRParsers
	sort.Strings(attrReply.Contexts)
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}
}

func testAttributeSGetAttributeForEventNotFound(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaCDRs),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
				Event: map[string]interface{}{
					utils.Account:     "dan",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	eAttrPrf2 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_3",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Contexts:  []string{utils.MetaSessionS},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Account,
					Value: config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				},
			},
			Weight: 10.0,
		},
	}
	eAttrPrf2.Compile()
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, eAttrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_3"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(eAttrPrf2.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", eAttrPrf2, reply)
	}
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSGetAttributeForEventWithMetaAnyContext(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaCDRs),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSGetAttributeForEventWihMetaAnyContext",
				Event: map[string]interface{}{
					utils.Account:     "dan",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	eAttrPrf2 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    ev.Tenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Contexts:  []string{utils.META_ANY},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Account,
					Value: config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				},
			},
			Weight: 10.0,
		},
	}
	eAttrPrf2.Compile()
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, eAttrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_2"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(eAttrPrf2.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", eAttrPrf2.AttributeProfile, reply)
	}
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Fatal(err)
	}
	attrReply.Compile()
	if !reflect.DeepEqual(eAttrPrf2.AttributeProfile, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf2.AttributeProfile), utils.ToJSON(attrReply))
	}
}

func testAttributeSProcessEvent(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]interface{}{
					utils.Account:     "1007",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + utils.Account,
			utils.MetaReq + utils.NestingSep + utils.Subject},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Subject:     "1001",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	} else {
		sort.Strings(eRply.AlteredFields)
		sort.Strings(rplyEv.AlteredFields)
		if !reflect.DeepEqual(eRply, &rplyEv) { // second for reversed order of attributes
			t.Errorf("Expecting: %s, received: %s",
				utils.ToJSON(eRply), utils.ToJSON(rplyEv))
		}
	}

	ev.Tenant = ""
	ev.ID = "randomID"
	eRply.ID = "randomID"
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	} else {
		sort.Strings(eRply.AlteredFields)
		sort.Strings(rplyEv.AlteredFields)
		if !reflect.DeepEqual(eRply, &rplyEv) { // second for reversed order of attributes
			t.Errorf("Expecting: %s, received: %s",
				utils.ToJSON(eRply), utils.ToJSON(rplyEv))
		}
	}
}

func testAttributeSProcessEventNotFound(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEventNotFound",
				Event: map[string]interface{}{
					utils.Account:     "Inexistent",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSProcessEventMissing(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSProcessEvent",
				Event: map[string]interface{}{
					utils.Account:     "NonExist",
					utils.Category:    "*attributes",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err == nil ||
		err.Error() != "MANDATORY_IE_MISSING: [Category]" {
		t.Error(err)
	}
}

func testAttributeSProcessEventWithNoneSubstitute(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSWithNoneSubstitute",
				Event: map[string]interface{}{
					utils.Account:     "1008",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "AttributeWithNonSubstitute",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"*string:~*req.Account:1008"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.Account:1008"},
					Path:      utils.MetaReq + utils.NestingSep + utils.Account,
					Value:     config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.NewRSRParsersMustCompile(utils.MetaRemove, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"AttributeWithNonSubstitute"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + utils.Account,
			utils.MetaReq + utils.NestingSep + utils.Subject},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSWithNoneSubstitute",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "+491511231234",
				},
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	}
	sort.Strings(rplyEv.AlteredFields)
	if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithNoneSubstitute2(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSWithNoneSubstitute",
				Event: map[string]interface{}{
					utils.Account:     "1008",
					utils.Subject:     "1008",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "AttributeWithNonSubstitute",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Account:1008"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.Account:1008"},
					Path:      utils.MetaReq + utils.NestingSep + utils.Account,
					Value:     config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.NewRSRParsersMustCompile(utils.MetaRemove, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"AttributeWithNonSubstitute"},
		AlteredFields:   []string{"*req.Account", "*req.Subject"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSWithNoneSubstitute",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	eRply2 := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"AttributeWithNonSubstitute"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + utils.Subject,
			utils.MetaReq + utils.NestingSep + utils.Account},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSWithNoneSubstitute",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) &&
		!reflect.DeepEqual(eRply2, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithNoneSubstitute3(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSWithNoneSubstitute",
				Event: map[string]interface{}{
					utils.Account:     "1008",
					utils.Subject:     "1001",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "AttributeWithNonSubstitute",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Account:1008"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.Account:1008"},
					Path:      utils.MetaReq + utils.NestingSep + utils.Account,
					Value:     config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				},
				{
					FilterIDs: []string{"*string:~*req.Subject:1008"},
					Path:      utils.MetaReq + utils.NestingSep + utils.Subject,
					Value:     config.NewRSRParsersMustCompile(utils.MetaRemove, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"AttributeWithNonSubstitute"},
		AlteredFields:   []string{"*req.Account"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAttributeSWithNoneSubstitute",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Subject:     "1001",
					utils.Destination: "+491511231234",
				},
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithHeader(t *testing.T) {
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_Header",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Field1:Value1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: config.NewRSRParsersMustCompile("~*req.Field1", utils.INFIELD_SEP),
				},
			},
			Blocker: true,
			Weight:  5,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		ProcessRuns: utils.IntPointer(1),
		Context:     utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     "HeaderEventForAttribute",
				Event: map[string]interface{}{
					"Field1": "Value1",
				},
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_Header"},
		AlteredFields:   []string{"*req.Field2"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     "HeaderEventForAttribute",
				Event: map[string]interface{}{
					"Field1": "Value1",
					"Field2": "Value1",
				},
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSGetAttPrfIDs(t *testing.T) {
	expected := []string{"ATTR_2", "ATTR_PASS", "ATTR_1", "ATTR_3", "ATTR_Header", "AttributeWithNonSubstitute"}
	var result []string
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{
		Tenant:    "cgrates.org",
		Paginator: utils.Paginator{Limit: utils.IntPointer(10)},
	}, &result); err != nil {
		t.Error(err)
	} else if 10 < len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testAttributeSSetAlsPrfBrokenReference(t *testing.T) {
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"FLTR_ACNT_danBroken", "FLTR_DST_DEBroken"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "FL1",
					Value: config.NewRSRParsersMustCompile("Al1", utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	var result string
	expErr := "SERVER_ERROR: broken reference to filter: FLTR_ACNT_danBroken for item with ID: cgrates.org:ApierTest"
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
}

func testAttributeSSetAlsPrf(t *testing.T) {
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "FL1",
					Value: config.NewRSRParsersMustCompile("Al1", utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
}

func testAttributeSUpdateAlsPrf(t *testing.T) {
	alsPrf.Attributes = []*engine.Attribute{
		{
			Path:  utils.MetaReq + utils.NestingSep + "FL1",
			Value: config.NewRSRParsersMustCompile("Al1", utils.INFIELD_SEP),
		},
		{
			Path:  utils.MetaReq + utils.NestingSep + "FL2",
			Value: config.NewRSRParsersMustCompile("Al2", utils.INFIELD_SEP),
		},
	}
	alsPrf.Compile()
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	sort.Strings(reply.FilterIDs)
	sort.Strings(alsPrf.AttributeProfile.FilterIDs)
	sort.Strings(reply.Contexts)
	sort.Strings(alsPrf.AttributeProfile.Contexts)
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
}

func testAttributeSRemAlsPrf(t *testing.T) {
	var resp string
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{Tenant: alsPrf.Tenant, ID: alsPrf.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if alsPrfConfigDIR == "tutinternal" { // do not double remove the profile
		return
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// remove twice shoud return not found
	resp = ""
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{Tenant: alsPrf.Tenant, ID: alsPrf.ID}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v received: %v", utils.ErrNotFound, err)
	}
}

func testAttributeSSetAlsPrf2(t *testing.T) {
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "golant",
			ID:        "ATTR_972587832508_SESSIONAUTH",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Account:972587832508"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "roam",
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
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "golant", ID: "ATTR_972587832508_SESSIONAUTH"}}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.AttributeProfile, reply)
	}
}

func testAttributeSSetAlsPrf3(t *testing.T) {
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "golant",
			ID:        "ATTR_972587832508_SESSIONAUTH",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:Account:972587832508"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.RSRParsers{
						&config.RSRParser{
							Rules: "",
						},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err == nil {
		t.Error(err)
	}
}

func testAttributeSSetAlsPrf4(t *testing.T) {
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "golant",
			ID:        "ATTR_972587832508_SESSIONAUTH",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Account:972587832508"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path: utils.MetaReq + utils.NestingSep + utils.Subject,
					Value: config.RSRParsers{
						&config.RSRParser{},
					},
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err == nil {
		t.Error(err)
	}
}

func testAttributeSPing(t *testing.T) {
	var resp string
	if err := attrSRPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testAttributeSProcessEventWithSearchAndReplace(t *testing.T) {
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_Search_and_replace",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Category:call"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Category",
					Value: config.NewRSRParsersMustCompile("~*req.Category:s/(.*)/${1}_suffix/", utils.INFIELD_SEP),
				},
			},
			Blocker: true,
			Weight:  10,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		ProcessRuns: utils.IntPointer(1),
		Context:     utils.StringPointer(utils.MetaSessionS),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     "HeaderEventForAttribute",
				Event: map[string]interface{}{
					"Category": "call",
				},
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_Search_and_replace"},
		AlteredFields:   []string{"*req.Category"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     "HeaderEventForAttribute",
				Event: map[string]interface{}{
					"Category": "call_suffix",
				},
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		attrArgs, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessWithMultipleRuns(t *testing.T) {
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
	}
	attrPrf2 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_2",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Field1:Value1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: config.NewRSRParsersMustCompile("Value2", utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	attrPrf3 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_3",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.NotFound:NotFound"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Value: config.NewRSRParsersMustCompile("Value3", utils.INFIELD_SEP),
				},
			},
			Weight: 30,
		},
	}
	// Add attribute in DM
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(4),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					"InitialField": "InitialValue",
				},
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2", "ATTR_1", "ATTR_2"},
		AlteredFields:   []string{"*req.Field1", "*req.Field2"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					"InitialField": "InitialValue",
					"Field1":       "Value1",
					"Field2":       "Value2",
				},
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
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
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
	}
	attrPrf2 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_2",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Field1:Value1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: config.NewRSRParsersMustCompile("Value2", utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	attrPrf3 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_3",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Field2:Value2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Value: config.NewRSRParsersMustCompile("Value3", utils.INFIELD_SEP),
				},
			},
			Weight: 30,
		},
	}
	// Add attributeProfiles
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(4),
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					"InitialField": "InitialValue",
				},
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2", "ATTR_3", "ATTR_2"},
		AlteredFields:   []string{"*req.Field1", "*req.Field2", "*req.Field3"},
		CGREventWithOpts: &utils.CGREventWithOpts{
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
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
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

func testAttributeSGetAttributeProfileIDsCount(t *testing.T) {
	var reply int
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{}, &reply); err != nil {
		t.Error(err)
	} else if reply != 7 {
		t.Errorf("Expecting: 7, received: %+v", reply)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 7 {
		t.Errorf("Expecting: 7, received: %+v", reply)
	}
	var resp string
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ATTR_1",
			Cache:  utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 6 {
		t.Errorf("Expecting: 6, received: %+v", reply)
	}
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ATTR_2",
			Cache:  utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 5 {
		t.Errorf("Expecting: 5, received: %+v", reply)
	}
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ATTR_3",
			Cache:  utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 4 {
		t.Errorf("Expecting: 4, received: %+v", reply)
	}
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ATTR_Header",
			Cache:  utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 3 {
		t.Errorf("Expecting: 3, received: %+v", reply)
	}
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ATTR_PASS",
			Cache:  utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("Expecting: 2, received: %+v", reply)
	}
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ATTR_Search_and_replace",
			Cache:  utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expecting: 1, received: %+v", reply)
	}
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "AttributeWithNonSubstitute",
			Cache:  utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

//Start tests for caching
func testAttributeSCachingMetaNone(t *testing.T) {
	//*none option should not add attribute in cache only in Datamanager
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Cache: utils.StringPointer(utils.META_NONE),
	}
	// set the profile
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply bool
	argsCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  "cgrates.org:ATTR_1",
	}
	if err := attrSRPC.Call(utils.CacheSv1HasItem, argsCache, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: false, received:%v", reply)
	}

	var rcvKeys []string
	argsCache2 := utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheAttributeProfiles,
	}
	if err := attrSRPC.Call(utils.CacheSv1GetItemIDs, argsCache2, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ",
			utils.ErrNotFound, err.Error(), rcvKeys)
	}

	//check in dataManager
	expected := []string{"ATTR_1"}
	var rcvIDs []string
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &rcvIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcvIDs) {
		t.Errorf("Expecting : %+v, received: %+v", expected, rcvIDs)
	}
}

func testAttributeSCachingMetaLoad(t *testing.T) {
	//*load option should add attribute in cache and in Datamanager
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Cache: utils.StringPointer(utils.MetaLoad),
	}
	// set the profile
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply bool
	argsCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  "cgrates.org:ATTR_1",
	}
	if err := attrSRPC.Call(utils.CacheSv1HasItem, argsCache, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: true, received:%v", reply)
	}

	var rcvKeys []string
	expectedIDs := []string{"cgrates.org:ATTR_1"}
	argsCache2 := utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheAttributeProfiles,
	}
	if err := attrSRPC.Call(utils.CacheSv1GetItemIDs, argsCache2, &rcvKeys); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rcvKeys, expectedIDs) {
		t.Errorf("Expecting : %+v, received: %+v", expectedIDs, rcvKeys)
	}

	rcvKeys = nil
	argsCache2 = utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheAttributeFilterIndexes,
	}
	expectedIDs = []string{"cgrates.org:*sessions:*string:*req.InitialField:InitialValue"}
	if err := attrSRPC.Call(utils.CacheSv1GetItemIDs, argsCache2, &rcvKeys); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rcvKeys, expectedIDs) {
		t.Errorf("Expecting : %+v, received: %+v", expectedIDs, rcvKeys)
	}

	//check in dataManager
	expected := []string{"ATTR_1"}
	var rcvIDs []string
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &rcvIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcvIDs) {
		t.Errorf("Expecting : %+v, received: %+v", expected, rcvIDs)
	}
	//remove from cache and DataManager the profile
	var resp string
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{Tenant: attrPrf1.Tenant, ID: attrPrf1.ID,
			Cache: utils.StringPointer(utils.MetaRemove)}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

	argsCache = utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  "cgrates.org:ATTR_1",
	}
	if err := attrSRPC.Call(utils.CacheSv1HasItem, argsCache, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: false, received:%v", reply)
	}

	if err := attrSRPC.Call(utils.CacheSv1GetItemIDs, argsCache2, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ",
			utils.ErrNotFound, err, rcvKeys)
	}

	//check in dataManager
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &rcvIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ",
			utils.ErrNotFound, err, rcvIDs)
	}
}

func testAttributeSCachingMetaReload1(t *testing.T) {
	//*reload add the attributes in cache if was there before
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Cache: utils.StringPointer(utils.MetaReload),
	}
	// set the profile
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply bool
	argsCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  "cgrates.org:ATTR_1",
	}
	if err := attrSRPC.Call(utils.CacheSv1HasItem, argsCache, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: false, received:%v", reply)
	}

	var rcvKeys []string
	argsCache2 := utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheAttributeProfiles,
	}
	if err := attrSRPC.Call(utils.CacheSv1GetItemIDs, argsCache2, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ",
			utils.ErrNotFound, err, rcvKeys)
	}

	//check in dataManager
	expected := []string{"ATTR_1"}
	var rcvIDs []string
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &rcvIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcvIDs) {
		t.Errorf("Expecting : %+v, received: %+v", expected, rcvIDs)
	}
}

func testAttributeSCachingMetaReload2(t *testing.T) {
	//add cache with *load option
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Cache: utils.StringPointer(utils.MetaLoad),
	}
	// set the profile
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}}, &reply); err != nil {
		t.Fatal(err)
	}
	attrPrf1.Compile()
	reply.Compile()
	if !reflect.DeepEqual(attrPrf1.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", attrPrf1.AttributeProfile, reply)
	}

	//add cache with *reload option
	// should overwrite the first
	attrPrf2 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Test:Test"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Cache: utils.StringPointer(utils.MetaReload),
	}
	// set the profile
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}}, &reply); err != nil {
		t.Fatal(err)
	}
	attrPrf2.Compile()
	reply.Compile()
	if !reflect.DeepEqual(attrPrf2.AttributeProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", attrPrf2.AttributeProfile, reply)
	}
}

func testAttributeSCachingMetaRemove(t *testing.T) {
	//add cache with *load option
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Cache: utils.StringPointer(utils.MetaLoad),
	}
	// set the profile
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply bool
	argsCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheAttributeProfiles,
		ItemID:  "cgrates.org:ATTR_1",
	}
	if err := attrSRPC.Call(utils.CacheSv1HasItem, argsCache, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: true, received:%v", reply)
	}

	var rcvKeys []string
	expectedIDs := []string{"cgrates.org:ATTR_1"}
	argsCache2 := utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheAttributeProfiles,
	}
	if err := attrSRPC.Call(utils.CacheSv1GetItemIDs, argsCache2, &rcvKeys); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rcvKeys, expectedIDs) {
		t.Errorf("Expecting : %+v, received: %+v", expectedIDs, rcvKeys)
	}

	// add with *remove cache option
	// should delete it from cache
	attrPrf2 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.Test:Test"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Cache: utils.StringPointer(utils.MetaRemove),
	}
	// set the profile
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := attrSRPC.Call(utils.CacheSv1HasItem, argsCache, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: false, received:%v", reply)
	}

	if err := attrSRPC.Call(utils.CacheSv1GetItemIDs, argsCache2, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ",
			utils.ErrNotFound, err.Error(), rcvKeys)
	}

	//check in dataManager
	expected := []string{"ATTR_1"}
	var rcvIDs []string
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &rcvIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcvIDs) {
		t.Errorf("Expecting : %+v, received: %+v", expected, rcvIDs)
	}

}

func testAttributeSSetAttributeWithEmptyPath(t *testing.T) {
	eAttrPrf2 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_3",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Contexts:  []string{utils.MetaSessionS},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
			Attributes: []*engine.Attribute{
				{
					Value: config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
				},
			},
			Weight: 10.0,
		},
	}
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, eAttrPrf2, &result); err == nil {
		t.Errorf("Expected error received nil")
	}
}

func testAttributeSCacheOpts(t *testing.T) {
	attrPrf1 := &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_WITH_OPTS",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: config.NewRSRParsersMustCompile("Value1", utils.INFIELD_SEP),
				},
			},
			Weight: 10,
		},
		Opts: map[string]interface{}{
			"Method":      "SetAttributeProfile",
			"CustomField": "somethingCustom",
		},
	}
	// set the profile
	var result string
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testAttributeSSetAlsPrfWithoutTenant(t *testing.T) {
	var reply string
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			ID:        "ApierTest1",
			Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
			FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "FL1",
					Value: config.NewRSRParsersMustCompile("Al1", utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	alsPrf.Compile()
	if err := attrSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	alsPrf.AttributeProfile.Tenant = "cgrates.org"
	var result *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{ID: "ApierTest1"}},
		&result); err != nil {
		t.Error(err)
	} else if result.Compile(); !reflect.DeepEqual(alsPrf.AttributeProfile, result) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(alsPrf.AttributeProfile), utils.ToJSON(result))
	}
}

func testAttributeSRmvAlsPrfWithoutTenant(t *testing.T) {
	var reply string
	if err := attrSRPC.Call(utils.APIerSv1RemoveAttributeProfile,
		&utils.TenantIDWithCache{ID: "ApierTest1"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.AttributeProfile
	if err := attrSRPC.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{ID: "ApierTest1"}},
		&result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
