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
	alsPrfCfgPath   string
	alsPrfCfg       *config.CGRConfig
	attrSRPC        *rpc.Client
	alsPrfDataDir   = "/usr/share/cgrates"
	alsPrf          *engine.AttributeProfile
	alsPrfDelay     int
	alsPrfConfigDIR string //run tests for specific configuration
)

var sTestsAlsPrf = []func(t *testing.T){
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
	testAttributeSProcessEventWithNoneSubstitute,
	testAttributeSProcessEventWithNoneSubstitute2,
	testAttributeSProcessEventWithNoneSubstitute3,
	testAttributeSGetAlsPrfBeforeSet,
	testAttributeSSetAlsPrf,
	testAttributeSUpdateAlsPrf,
	testAttributeSRemAlsPrf,
	testAttributeSKillEngine,
}

//Test start here
func TestAttributeSITMySql(t *testing.T) {
	alsPrfConfigDIR = "tutmysql"
	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func TestAttributeSITMongo(t *testing.T) {
	alsPrfConfigDIR = "tutmongo"
	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAttributeSInitCfg(t *testing.T) {
	var err error
	alsPrfCfgPath = path.Join(alsPrfDataDir, "conf", "samples", alsPrfConfigDIR)
	alsPrfCfg, err = config.NewCGRConfigFromFolder(alsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	alsPrfCfg.DataFolderPath = alsPrfDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(alsPrfCfg)
	switch alsPrfConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		alsPrfDelay = 2000
	default:
		alsPrfDelay = 1000
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
	if _, err := engine.StopStartEngine(alsPrfCfgPath, alsPrfDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAttributeSRPCConn(t *testing.T) {
	var err error
	attrSRPC, err = jsonrpc.Dial("tcp", alsPrfCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAttributeSGetAlsPrfBeforeSet(t *testing.T) {
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := attrSRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testAttributeSGetAttributeForEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSGetAttributeForEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1007",
			utils.Destination: "+491511231234",
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:Account:1007"},
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Account,
				Initial:    utils.ANY,
				Substitute: "1001",
				Append:     false,
			},
			&engine.Attribute{
				FieldName:  utils.Subject,
				Initial:    utils.ANY,
				Substitute: "1001",
				Append:     true,
			},
		},
		Weight: 10.0,
	}

	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}
}

func testAttributeSGetAttributeForEventNotFound(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSGetAttributeForEventWihMetaAnyContext",
		Context: utils.StringPointer(utils.MetaCDRs),
		Event: map[string]interface{}{
			utils.Account:     "dan",
			utils.Destination: "+491511231234",
		},
	}
	eAttrPrf2 := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_3",
		FilterIDs: []string{"*string:Account:dan"},
		Contexts:  []string{utils.MetaSessionS},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Account,
				Initial:    utils.ANY,
				Substitute: "1001",
				Append:     false,
			},
		},
		Weight: 10.0,
	}
	var result string
	if err := attrSRPC.Call("ApierV1.SetAttributeProfile", eAttrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_3"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAttrPrf2, reply) {
		t.Errorf("Expecting : %+v, received: %+v", eAttrPrf2, reply)
	}
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSGetAttributeForEventWithMetaAnyContext(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSGetAttributeForEventWihMetaAnyContext",
		Context: utils.StringPointer(utils.MetaCDRs),
		Event: map[string]interface{}{
			utils.Account:     "dan",
			utils.Destination: "+491511231234",
		},
	}
	eAttrPrf2 := &engine.AttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:Account:dan"},
		Contexts:  []string{utils.META_ANY},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC)},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Account,
				Initial:    utils.ANY,
				Substitute: "1001",
				Append:     false,
			},
		},
		Weight: 10.0,
	}
	var result string
	if err := attrSRPC.Call("ApierV1.SetAttributeProfile", eAttrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAttrPrf2, reply) {
		t.Errorf("Expecting : %+v, received: %+v", eAttrPrf2, reply)
	}
	var attrReply *engine.AttributeProfile
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAttrPrf2, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf2), utils.ToJSON(attrReply))
	}
}

func testAttributeSProcessEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1007",
			utils.Destination: "+491511231234",
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{utils.Subject, utils.Account},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "+491511231234",
			},
		},
	}
	eRply2 := &engine.AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{utils.Account, utils.Subject},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account: "1001",
				utils.Subject: "1001",
				"Destination": "+491511231234",
			},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := attrSRPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) &&
		!reflect.DeepEqual(eRply2, &rplyEv) { // second for reversed order of attributes
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testAttributeSProcessEventWithNoneSubstitute(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSWithNoneSubstitute",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1008",
			utils.Destination: "+491511231234",
		},
	}
	alsPrf = &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttributeWithNonSubstitute",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"*string:Account:1008"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Account,
				Initial:    "1008",
				Substitute: "1001",
				Append:     false,
			},
			&engine.Attribute{
				FieldName:  utils.Subject,
				Initial:    utils.ANY,
				Substitute: "*none",
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := attrSRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfile: "AttributeWithNonSubstitute",
		AlteredFields:  []string{utils.Account, utils.Subject},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSWithNoneSubstitute",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "+491511231234",
			},
		},
	}
	eRply2 := &engine.AttrSProcessEventReply{
		MatchedProfile: "AttributeWithNonSubstitute",
		AlteredFields:  []string{utils.Subject, utils.Account},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSWithNoneSubstitute",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "+491511231234",
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

func testAttributeSProcessEventWithNoneSubstitute2(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSWithNoneSubstitute",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1008",
			utils.Subject:     "1008",
			utils.Destination: "+491511231234",
		},
	}
	alsPrf = &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttributeWithNonSubstitute",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:Account:1008"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Account,
				Initial:    "1008",
				Substitute: "1001",
				Append:     false,
			},
			&engine.Attribute{
				FieldName:  utils.Subject,
				Initial:    utils.ANY,
				Substitute: utils.META_NONE,
				Append:     false,
			},
		},
		Weight: 20,
	}
	var result string
	if err := attrSRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfile: "AttributeWithNonSubstitute",
		AlteredFields:  []string{"Account", "Subject"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSWithNoneSubstitute",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "+491511231234",
			},
		},
	}
	eRply2 := &engine.AttrSProcessEventReply{
		MatchedProfile: "AttributeWithNonSubstitute",
		AlteredFields:  []string{utils.Subject, utils.Account},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSWithNoneSubstitute",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "+491511231234",
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
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSWithNoneSubstitute",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1008",
			utils.Destination: "+491511231234",
		},
	}
	alsPrf = &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttributeWithNonSubstitute",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:Account:1008"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Account,
				Initial:    "1008",
				Substitute: "1001",
				Append:     false,
			},
			&engine.Attribute{
				FieldName:  utils.Subject,
				Initial:    "1008",
				Substitute: utils.META_NONE,
				Append:     false,
			},
		},
		Weight: 20,
	}
	var result string
	if err := attrSRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfile: "AttributeWithNonSubstitute",
		AlteredFields:  []string{"Account", "Subject"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSWithNoneSubstitute",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "+491511231234",
			},
		},
	}
	eRply2 := &engine.AttrSProcessEventReply{
		MatchedProfile: "AttributeWithNonSubstitute",
		AlteredFields:  []string{utils.Subject, utils.Account},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSWithNoneSubstitute",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Destination: "+491511231234",
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

func testAttributeSSetAlsPrf(t *testing.T) {
	alsPrf = &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := attrSRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testAttributeSUpdateAlsPrf(t *testing.T) {
	alsPrf.Attributes = []*engine.Attribute{
		&engine.Attribute{
			FieldName:  "FL1",
			Initial:    "In1",
			Substitute: "Al1",
			Append:     true,
		},
		&engine.Attribute{
			FieldName:  "FL2",
			Initial:    "In2",
			Substitute: "Al2",
			Append:     false,
		},
	}
	var result string
	if err := attrSRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testAttributeSRemAlsPrf(t *testing.T) {
	var resp string
	if err := attrSRPC.Call("ApierV1.RemoveAttributeProfile",
		&ArgRemoveAttrProfile{Tenant: alsPrf.Tenant, ID: alsPrf.ID,
			Contexts: alsPrf.Contexts}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.AttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(alsPrfDelay); err != nil {
		t.Error(err)
	}
}
