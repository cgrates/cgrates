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
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspSession = []func(t *testing.T){
	testDspSessionAddBalacne,

	testDspSessionPingFailover,

	testDspSessionPing,
	testDspSessionTestAuthKey,
	testDspSessionAuthorize,
	testDspSessionInit,
	testDspGetSessions,
	testDspSessionUpdate,
	testDspSessionTerminate,
	testDspSessionProcessCDR,
	testDspSessionProcessEvent,
	testDspSessionProcessEvent2,

	testDspSessionProcessEvent3,

	testDspSessionReplicate,
	testDspSessionPassive,
	testDspSessionForceDisconect,
}

//Test start here
func TestDspSessionSTMySQL(t *testing.T) {
	testDsp(t, sTestsDspSession, "TestDspSessionS", "all", "all2", "dispatchers", "testit", "tutorial", "dispatchers")
}

func TestDspSessionSMongo(t *testing.T) {
	testDsp(t, sTestsDspSession, "TestDspSessionS", "all", "all2", "dispatchers_mongo", "testit", "tutorial", "dispatchers")
}

func testDspSessionAddBalacne(t *testing.T) {
	initUsage := 40 * time.Minute
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "1001",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("SessionBalance"),
		Value:         utils.Float64Pointer(float64(initUsage)),
		RatingSubject: utils.StringPointer("*zero5ms"),
	}
	var reply string
	if err := allEngine.RCP.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  attrSetBalance.Tenant,
		Account: attrSetBalance.Account,
	}
	eAcntVal := float64(initUsage)
	if err := allEngine.RCP.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}
	if err := allEngine2.RCP.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if err := allEngine2.RCP.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}
}

func testDspSessionPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.SessionSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.SessionSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspSessionPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.SessionSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	if err := dispEngine.RCP.Call(utils.SessionSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.SessionSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RCP.Call(utils.SessionSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but recived %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspSessionTestAuthKey(t *testing.T) {
	authUsage := 5 * time.Minute
	args := sessions.V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetSuppliers:       true,
		GetAttributes:      true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("12345"),
		},
	}
	var rply sessions.V1AuthorizeReplyWithDigest
	if err := dispEngine.RCP.Call(utils.SessionSv1AuthorizeEventWithDigest,
		args, &rply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspSessionAuthorize(t *testing.T) {
	authUsage := 5 * time.Minute
	argsAuth := &sessions.V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetSuppliers:       true,
		GetAttributes:      true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply sessions.V1AuthorizeReplyWithDigest
	if err := dispEngine.RCP.Call(utils.SessionSv1AuthorizeEventWithDigest,
		argsAuth, &rply); err != nil {
		t.Error(err)
		return
	}
	if *rply.MaxUsage != authUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", *rply.MaxUsage)
	}
	if *rply.ResourceAllocation == "" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eSplrs := "supplier1,supplier2"
	tp := strings.Split(*rply.SuppliersDigest, ",")
	sort.Strings(tp)
	*rply.SuppliersDigest = strings.Join(tp, ",")
	if eSplrs != *rply.SuppliersDigest {
		t.Errorf("expecting: %v, received: %v", eSplrs, *rply.SuppliersDigest)
	}
	eAttrs := "OfficeGroup:Marketing"
	if eAttrs != *rply.AttributesDigest {
		t.Errorf("expecting: %v, received: %v", eAttrs, *rply.AttributesDigest)
	}
}

func testDspSessionInit(t *testing.T) {
	initUsage := time.Duration(5 * time.Minute)
	argsInit := &sessions.V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		GetAttributes:     true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItInitiateSession",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply sessions.V1InitReplyWithDigest
	if err := dispEngine.RCP.Call(utils.SessionSv1InitiateSessionWithDigest,
		argsInit, &rply); err != nil {
		t.Fatal(err)
	}
	if *rply.MaxUsage != initUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", *rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
}

func testDspGetSessions(t *testing.T) {
	filtr := utils.SessionFilter{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
		Tenant:  "cgrates.org",
		Filters: []string{},
	}
	var reply int
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessionsCount,
		&filtr, &reply); err != nil {
		t.Fatal(err)
	} else if reply != 2 {
		t.Errorf("Expected 2 active sessions recived %v", reply)
	}
	var rply []*sessions.ExternalSession
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessions,
		&filtr, &rply); err != nil {
		t.Fatal(err)
	} else if len(rply) != 2 {
		t.Errorf("Unexpected number of sessions returned %v :%s", len(rply), utils.ToJSON(rply))
	}

	if err := dispEngine.RCP.Call(utils.SessionSv1GetPassiveSessionsCount,
		&filtr, &reply); err != nil {
		t.Fatal(err)
	} else if reply != 0 {
		t.Errorf("Expected no pasive sessions recived %v", reply)
	}
	rply = nil
	if err := dispEngine.RCP.Call(utils.SessionSv1GetPassiveSessions,
		&filtr, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected %v recived %v with reply %s", utils.ErrNotFound, err, utils.ToJSON(rply))
	}
}

func testDspSessionUpdate(t *testing.T) {
	reqUsage := 5 * time.Minute
	argsUpdate := &sessions.V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := dispEngine.RCP.Call(utils.SessionSv1UpdateSession,
		argsUpdate, &rply); err != nil {
		t.Error(err)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   "2018-01-07T17:00:00Z",
				utils.AnswerTime:  "2018-01-07T17:00:10Z",
				utils.Usage:       float64(reqUsage),
				"CGRID":           "5668666d6b8e44eb949042f25ce0796ec3592ff9",
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", utils.ToJSON(rply))
	}
}

func testDspSessionUpdate2(t *testing.T) {
	reqUsage := 5 * time.Minute
	argsUpdate := &sessions.V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := dispEngine.RCP.Call(utils.SessionSv1UpdateSession,
		argsUpdate, &rply); err != nil {
		t.Fatal(err)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1001_SESSIONAUTH"},
		AlteredFields:   []string{"LCRProfile", "Password", "RequestType", "PaypalAccount"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Destination: "1002",
				"LCRProfile":      "premium_cli",
				"Password":        "CGRateS.org",
				"PaypalAccount":   "cgrates@paypal.com",
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   "2018-01-07T17:00:00Z",
				utils.AnswerTime:  "2018-01-07T17:00:10Z",
				utils.Usage:       float64(reqUsage),
			},
		},
	}
	sort.Strings(eAttrs.AlteredFields)
	if rply.Attributes != nil && rply.Attributes.AlteredFields != nil {
		sort.Strings(rply.Attributes.AlteredFields)
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", utils.ToJSON(rply))
	}
}

func testDspSessionTerminate(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		ReleaseResources: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply string
	if err := dispEngine.RCP.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testDspSessionProcessCDR(t *testing.T) {
	args := utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessCDR",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}

	var rply string
	if err := dispEngine.RCP.Call(utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testDspSessionProcessEvent(t *testing.T) {
	initUsage := 5 * time.Minute
	args := sessions.V1ProcessMessageArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It2",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply sessions.V1ProcessMessageReply
	if err := dispEngine.RCP.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	if *rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "TestSSv1It2",
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   "2018-01-07T17:00:00Z",
				utils.AnswerTime:  "2018-01-07T17:00:10Z",
				utils.Usage:       300000000000.0,
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
}

func testDspSessionProcessEvent2(t *testing.T) {
	initUsage := 5 * time.Minute
	args := sessions.V1ProcessMessageArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It2",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
				utils.EVENT_NAME:  "Internal",
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("pse12345"),
		},
	}
	var rply sessions.V1ProcessMessageReply
	if err := dispEngine.RCP.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	if *rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
		AlteredFields:   []string{"Password", "EventName"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Destination: "1002",
				"Password":        "CGRateS.org",
				utils.OriginID:    "TestSSv1It2",
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   "2018-01-07T17:00:00Z",
				utils.AnswerTime:  "2018-01-07T17:00:10Z",
				utils.Usage:       300000000000.0,
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
}

func testDspSessionReplicate(t *testing.T) {
	allEngine.initDataDb(t)
	allEngine.resetStorDb(t)
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "testit"))
	testDspSessionAddBalacne(t)
	testDspSessionAuthorize(t)
	testDspSessionInit(t)

	var reply string
	if err := dispEngine.RCP.Call(utils.SessionSv1ReplicateSessions, ArgsReplicateSessionsWithApiKey{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
		ArgsReplicateSessions: sessions.ArgsReplicateSessions{
			CGRID:   "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
			Passive: false,
			Connections: []*config.RemoteHost{
				&config.RemoteHost{
					Address:   "127.0.0.1:7012",
					Transport: utils.MetaJSONrpc,
				},
			},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}

	var repl int
	time.Sleep(10 * time.Millisecond)
	if err := allEngine2.RCP.Call(utils.SessionSv1GetPassiveSessionsCount,
		nil, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 2 {
		t.Errorf("Expected 1 sessions recived %v", repl)
	}
}

func testDspSessionPassive(t *testing.T) {
	allEngine.stopEngine(t)
	testDspSessionUpdate2(t)
	var repl int
	filtr := utils.SessionFilter{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
		Tenant:  "cgrates.org",
		Filters: []string{},
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RCP.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no passive sessions recived %v", repl)
	}
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 2 {
		t.Errorf("Expected 1 active sessions recived %v", repl)
	}

	var rply []*sessions.ExternalSession
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessions,
		&filtr, &rply); err != nil {
		t.Fatal(err)
	} else if len(rply) != 2 {
		t.Errorf("Unexpected number of sessions returned %v :%s", len(rply), utils.ToJSON(rply))
	}

	var reply string
	if err := dispEngine.RCP.Call(utils.SessionSv1SetPassiveSession, sessions.Session{
		CGRID:      rply[0].CGRID,
		Tenant:     rply[0].Tenant,
		ResourceID: "TestSSv1It1",
		EventStart: engine.NewMapEvent(map[string]interface{}{
			utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
			utils.Tenant:      "cgrates.org",
			utils.Category:    "call",
			utils.ToR:         utils.VOICE,
			utils.OriginID:    "TestSSv1It1",
			utils.RequestType: utils.META_PREPAID,
			utils.Account:     "1001",
			utils.Destination: "1002",
			utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:       5 * time.Minute,
		}),
		SRuns: []*sessions.SRun{
			&sessions.SRun{
				Event: engine.NewMapEvent(map[string]interface{}{
					"RunID":           "CustomerCharges",
					utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: utils.META_PREPAID,
					utils.Account:     "1001",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       5 * time.Minute,
				}),
				CD:        &engine.CallDescriptor{},
				EventCost: &engine.EventCost{},

				LastUsage:  5 * time.Minute,
				TotalUsage: 10 * time.Minute,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RCP.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 1 {
		t.Errorf("Expected 1 passive sessions recived %v", repl)
	}
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no active sessions recived %v", repl)
	}
}

func testDspSessionForceDisconect(t *testing.T) {
	allEngine.startEngine(t)
	allEngine.initDataDb(t)
	allEngine.resetStorDb(t)
	allEngine.loadData(t, path.Join(dspDataDir, "tariffplans", "testit"))
	testDspSessionAddBalacne(t)
	testDspSessionAuthorize(t)
	testDspSessionInit(t)
	var repl int
	filtr := utils.SessionFilter{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
		Tenant:  "cgrates.org",
		Filters: []string{},
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RCP.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no passive sessions recived %v", repl)
	}
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 2 {
		t.Errorf("Expected 1 active sessions recived %v", repl)
	}

	var rply []*sessions.ExternalSession
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessions,
		&filtr, &rply); err != nil {
		t.Fatal(err)
	} else if len(rply) != 2 {
		t.Errorf("Unexpected number of sessions returned %v :%s", len(rply), utils.ToJSON(rply))
	}

	var reply string
	if err := dispEngine.RCP.Call(utils.SessionSv1ForceDisconnect, &filtr, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RCP.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected 1 passive sessions recived %v", repl)
	}
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no active sessions recived %v", repl)
	}
}

func testDspSessionProcessEvent3(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{"*rals:*terminate", "*resources:*release"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventTerminateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testSSv1ItProcessEvent",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("ses12345"),
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := dispEngine.RCP.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}

	var repl int
	if err := dispEngine.RCP.Call(utils.SessionSv1GetActiveSessionsCount,
		utils.SessionFilter{
			ArgDispatcher: &utils.ArgDispatcher{
				APIKey: utils.StringPointer("ses12345"),
			},
			Tenant:  "cgrates.org",
			Filters: []string{},
		}, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no active sessions recived %v", repl)
	}
}
