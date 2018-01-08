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
	alsPrf          *engine.ExternalAttributeProfile
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
	testAttributeSProcessEvent,
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
	var reply *engine.ExternalAttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
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
		Context: utils.StringPointer(utils.MetaRating),
		Event: map[string]interface{}{
			utils.Account:     "1007",
			utils.Destination: "+491511231234",
		},
	}
	eAttrPrf := &engine.ExternalAttributeProfile{
		Tenant:    ev.Tenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"FLTR_ACNT_1007"},
		Context:   utils.MetaRating,
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
	reverseSubstitute := []*engine.Attribute{
		&engine.Attribute{
			FieldName:  utils.Subject,
			Initial:    utils.ANY,
			Substitute: "1001",
			Append:     true,
		},
		&engine.Attribute{
			FieldName:  utils.Account,
			Initial:    utils.ANY,
			Substitute: "1001",
			Append:     false,
		},
	}

	var attrReply *engine.ExternalAttributeProfile
	if err := attrSRPC.Call(utils.AttributeSv1GetAttributeForEvent,
		ev, &attrReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAttrPrf.Tenant, attrReply.Tenant) {
		t.Errorf("Expecting: %s, received: %s", eAttrPrf.Tenant, attrReply.Tenant)
	} else if !reflect.DeepEqual(eAttrPrf.ID, attrReply.ID) {
		t.Errorf("Expecting: %s, received: %s", eAttrPrf.ID, attrReply.ID)
	} else if !reflect.DeepEqual(eAttrPrf.Context, attrReply.Context) {
		t.Errorf("Expecting: %s, received: %s", eAttrPrf.Tenant, attrReply.Tenant)
	} else if !reflect.DeepEqual(eAttrPrf.FilterIDs, attrReply.FilterIDs) {
		t.Errorf("Expecting: %s, received: %s", eAttrPrf.FilterIDs, attrReply.FilterIDs)
	} else if !reflect.DeepEqual(eAttrPrf.ActivationInterval.ActivationTime.Local(), attrReply.ActivationInterval.ActivationTime.Local()) {
		t.Errorf("Expecting: %s, received: %s",
			eAttrPrf.ActivationInterval.ActivationTime.Local(), attrReply.ActivationInterval.ActivationTime.Local())
	} else if !reflect.DeepEqual(eAttrPrf.ActivationInterval.ExpiryTime.Local(), attrReply.ActivationInterval.ExpiryTime.Local()) {
		t.Errorf("Expecting: %s, received: %s",
			eAttrPrf.ActivationInterval.ExpiryTime.Local(), attrReply.ActivationInterval.ExpiryTime.Local())
	} else if !reflect.DeepEqual(eAttrPrf.Attributes, attrReply.Attributes) && !reflect.DeepEqual(reverseSubstitute, attrReply.Attributes) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf.Attributes), utils.ToJSON(attrReply.Attributes))
	} else if !reflect.DeepEqual(eAttrPrf.Weight, attrReply.Weight) {
		t.Errorf("Expecting: %s, received: %s", eAttrPrf.Weight, attrReply.Weight)
	}
}

func testAttributeSProcessEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaRating),
		Event: map[string]interface{}{
			"Account":     "1007",
			"Destination": "+491511231234",
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{"Subject", "Account"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaRating),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Subject":     "1001",
				"Destination": "+491511231234",
			},
		},
	}
	eRply2 := &engine.AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{"Account", "Subject"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaRating),
			Event: map[string]interface{}{
				"Account":     "1001",
				"Subject":     "1001",
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

func testAttributeSSetAlsPrf(t *testing.T) {
	alsPrf = &engine.ExternalAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Context:   "*rating",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
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
	var reply *engine.ExternalAttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf.FilterIDs, reply.FilterIDs) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.FilterIDs, reply.FilterIDs)
	} else if !reflect.DeepEqual(alsPrf.ActivationInterval, reply.ActivationInterval) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ActivationInterval, reply.ActivationInterval)
	} else if !reflect.DeepEqual(len(alsPrf.Attributes), len(reply.Attributes)) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.Attributes), utils.ToJSON(reply.Attributes))
	} else if !reflect.DeepEqual(alsPrf.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ID, reply.ID)
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
	var reply *engine.ExternalAttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf.FilterIDs, reply.FilterIDs) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.FilterIDs, reply.FilterIDs)
	} else if !reflect.DeepEqual(alsPrf.ActivationInterval, reply.ActivationInterval) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ActivationInterval, reply.ActivationInterval)
	} else if !reflect.DeepEqual(len(alsPrf.Attributes), len(reply.Attributes)) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.Attributes), utils.ToJSON(reply.Attributes))
	} else if !reflect.DeepEqual(alsPrf.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ID, reply.ID)
	}
}

func testAttributeSRemAlsPrf(t *testing.T) {
	var resp string
	if err := attrSRPC.Call("ApierV1.RemAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.ExternalAttributeProfile
	if err := attrSRPC.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAttributeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(alsPrfDelay); err != nil {
		t.Error(err)
	}
}
