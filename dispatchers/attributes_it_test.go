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

package dispatchers

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspAttr = []func(t *testing.T){
	testDspAttrPingFailover,
	testDspAttrPingFailover2,
	testDspAttrPingFailoverNotFoundHost,
	testDspAttrGetAttrFailover,
	testDspAttrGetAttrRoundRobin,

	testDspAttrPing,
	testDspAttrPingEmptyCGREventWIthArgDispatcher,
	testDspAttrTestMissingArgDispatcher,
	testDspAttrTestMissingApiKey,
	testDspAttrTestUnknownApiKey,
	testDspAttrTestAuthKey,
	testDspAttrTestAuthKey2,
	testDspAttrTestAuthKey3,

	testDspAttrGetAttrInternal,
}

//Test start here
func TestDspAttributeS(t *testing.T) {
	var config1, config2, config3 string
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspAttr, "TestDspAttributeS", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func TestDspAttributeSNoConn(t *testing.T) {
	if *dbType != utils.MetaMySQL {
		t.SkipNow()
	}
	testDsp(t, []func(t *testing.T){
		testDspAttrPingFailover,
		testDspAttrPing,
		testDspAttrPingNoArgDispatcher,
	}, "TestDspAttributeS", "all_mysql", "all2_mysql", "dispatchers_no_attributes", "tutorial", "oldtutorial", "dispatchers")
}

func testDspAttrPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	if err := allEngine2.RPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	reply = ""
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	reply = ""
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
	reply = ""
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspAttrPingFailoverNotFoundHost(t *testing.T) {
	var reply string
	if err := allEngine2.RPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"EventName": "UnexistedHost",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}

	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t) // stop the engine and we expect to get error
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine2.startEngine(t)
	reply = ""
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspAttrPingFailover2(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	if err := allEngine2.RPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	reply = ""
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}
	allEngine.stopEngine(t) // stop the engine and the call should go to the second engine
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	reply = ""
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
	reply = ""
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspAttrGetAttrFailover(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.EVENT_NAME: "Event1",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    args.Tenant,
		ID:        "ATTR_1002_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20.0,
	}
	eAttrPrf.Compile()
	if *encoding == utils.MetaGOB {
		eAttrPrf.Attributes[0].FilterIDs = nil // empty slice are nil in gob
	}

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1002_SIMPLEAUTH"},
		AlteredFields:   []string{"*req.Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.EVENT_NAME: "Event1",
				"Password":       "CGRateS.org",
			},
		},
	}

	var attrReply *engine.AttributeProfile
	var rplyEv engine.AttrSProcessEventReply
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := dispEngine.RPC.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	} else if reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}

	allEngine2.stopEngine(t)

	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply != nil {
		attrReply.Compile()
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	if err := dispEngine.RPC.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}

	allEngine2.startEngine(t)
}

func testDspAttrPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if dispEngine.RPC == nil {
		t.Fatal(dispEngine.RPC)
	}
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspAttrPingEmptyCGREventWIthArgDispatcher(t *testing.T) {
	expected := "MANDATORY_IE_MISSING: [APIKey]"
	var reply string
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping,
		&utils.CGREventWithArgDispatcher{}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testDspAttrTestMissingArgDispatcher(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField).Error() {
		t.Errorf("Error:%v rply=%s", err, utils.ToJSON(attrReply))
	}
}

func testDspAttrTestMissingApiKey(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.NewErrMandatoryIeMissing(utils.APIKey).Error() {
		t.Errorf("Error:%v rply=%s", err, utils.ToJSON(attrReply))
	}
}

func testDspAttrTestUnknownApiKey(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("1234"),
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrUnknownApiKey.Error() {
		t.Error(err)
	}
}

func testDspAttrTestAuthKey(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("12345"),
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspAttrTestAuthKey2(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account: "1001",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    args.Tenant,
		ID:        "ATTR_1001_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20.0,
	}
	eAttrPrf.Compile()
	if *encoding == utils.MetaGOB {
		eAttrPrf.Attributes[0].FilterIDs = nil // empty slice are nil in gob
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply != nil {
		attrReply.Compile()
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
		AlteredFields:   []string{"*req.Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account: "1001",
				"Password":    "CGRateS.org",
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := dispEngine.RPC.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testDspAttrTestAuthKey3(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account:    "1001",
				utils.EVENT_NAME: "Event1",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}
	var attrReply *engine.AttributeProfile
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDspAttrGetAttrRoundRobin(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.EVENT_NAME: "RoundRobin",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}
	eAttrPrf := &engine.AttributeProfile{
		Tenant:    args.Tenant,
		ID:        "ATTR_1002_SIMPLEAUTH",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Contexts:  []string{"simpleauth"},
		Attributes: []*engine.Attribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "Password",
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("CGRateS.org", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20.0,
	}
	eAttrPrf.Compile()
	if *encoding == utils.MetaGOB {
		eAttrPrf.Attributes[0].FilterIDs = nil // empty slice are nil in gob
	}

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1002_SIMPLEAUTH"},
		AlteredFields:   []string{"*req.Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.EVENT_NAME: "RoundRobin",
				"Password":       "CGRateS.org",
			},
		},
	}

	var attrReply *engine.AttributeProfile
	var rplyEv engine.AttrSProcessEventReply
	// To ALL2
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// To ALL
	if err := dispEngine.RPC.Call(utils.AttributeSv1GetAttributeForEvent,
		args, &attrReply); err != nil {
		t.Error(err)
	}
	if attrReply != nil {
		attrReply.Compile()
	}
	if !reflect.DeepEqual(eAttrPrf, attrReply) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eAttrPrf), utils.ToJSON(attrReply))
	}

	// To ALL2
	if err := dispEngine.RPC.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	} else if reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}

	// To ALL
	if err := dispEngine.RPC.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testDspAttrGetAttrInternal(t *testing.T) {
	args := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer("simpleauth"),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.EVENT_NAME: "Internal",
				utils.Account:    "1003",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("attr12345"),
		},
	}

	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1003_SIMPLEAUTH"},
		AlteredFields:   []string{"*req.Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSGetAttributeForEvent",
			Event: map[string]interface{}{
				utils.Account:    "1003",
				utils.EVENT_NAME: "Internal",
				"Password":       "CGRateS.com",
			},
		},
	}

	var rplyEv engine.AttrSProcessEventReply
	if err := dispEngine.RPC.Call(utils.AttributeSv1ProcessEvent,
		args, &rplyEv); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRply, &rplyEv) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eRply), utils.ToJSON(rplyEv))
	}
}

func testDspAttrPingNoArgDispatcher(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.AttributeSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if dispEngine.RPC == nil {
		t.Fatal(dispEngine.RPC)
	}
	if err := dispEngine.RPC.Call(utils.AttributeSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{Tenant: "cgrates.org"},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}
