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
	"sort"
	"strings"
	"testing"
	"time"

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
	testDspSessionUpdate,
	testDspSessionTerminate,
	testDspSessionProcessCDR,
	testDspSessionProcessEvent,
}

//Test start here
func TestDspSessionSTMySQL(t *testing.T) {
	testDsp(t, sTestsDspSession, "TestDspSessionS", "all", "all2", "attributes", "dispatchers", "testit", "oldtutorial", "dispatchers")
}

func TestDspSessionSMongo(t *testing.T) {
	testDsp(t, sTestsDspSession, "TestDspSessionS", "all", "all2", "attributes_mongo", "dispatchers_mongo", "testit", "oldtutorial", "dispatchers")
}

func testDspSessionAddBalacne(t *testing.T) {
	initUsage := 15 * time.Minute
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "1001",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("SessionBalance"),
		Value:         utils.Float64Pointer(float64(initUsage)),
		RatingSubject: utils.StringPointer("*zero5ms"),
	}
	var reply string
	if err := allEngine.RCP.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
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
}

func testDspSessionPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.SessionSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RCP.Call(utils.SessionSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
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
	ev := CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
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
	args := AuthorizeArgsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "12345",
		},
		V1AuthorizeArgs: sessions.V1AuthorizeArgs{
			GetMaxUsage:        true,
			AuthorizeResources: true,
			GetSuppliers:       true,
			GetAttributes:      true,
			CGREvent: utils.CGREvent{
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
	argsAuth := &AuthorizeArgsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
		},
		V1AuthorizeArgs: sessions.V1AuthorizeArgs{
			GetMaxUsage:        true,
			AuthorizeResources: true,
			GetSuppliers:       true,
			GetAttributes:      true,
			CGREvent: utils.CGREvent{
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
	argsInit := &InitArgsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
		},
		V1InitSessionArgs: sessions.V1InitSessionArgs{
			InitSession:       true,
			AllocateResources: true,
			GetAttributes:     true,
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession",
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
					utils.Usage:       initUsage,
				},
			},
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

func testDspSessionUpdate(t *testing.T) {
	reqUsage := 5 * time.Minute
	argsUpdate := &UpdateSessionWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
		},
		V1UpdateSessionArgs: sessions.V1UpdateSessionArgs{
			GetAttributes: true,
			UpdateSession: true,
			CGREvent: utils.CGREvent{
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
			Tenant:  "cgrates.org",
			ID:      "TestSSv1ItUpdateSession",
			Context: utils.StringPointer(utils.MetaSessionS),
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
				utils.Usage:       300000000000.0,
				"CGRID":           "5668666d6b8e44eb949042f25ce0796ec3592ff9",
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
	if *rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", utils.ToJSON(rply))
	}
}

func testDspSessionTerminate(t *testing.T) {
	args := &TerminateSessionWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
		},
		V1TerminateSessionArgs: sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			ReleaseResources: true,
			CGREvent: utils.CGREvent{
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
	args := CGREvWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
		},
		CGREvent: utils.CGREvent{
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
	args := ProcessEventWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "ses12345",
		},
		V1ProcessEventArgs: sessions.V1ProcessEventArgs{
			AllocateResources: true,
			Debit:             true,
			GetAttributes:     true,
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItProcessEvent",
				Event: map[string]interface{}{
					"CGRID":           "c87609aa1cb6e9529ab1836cfeeeb0ab7aa7ebaf",
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
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := dispEngine.RCP.Call(utils.SessionSv1ProcessEvent,
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
			Tenant:  "cgrates.org",
			ID:      "TestSSv1ItProcessEvent",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				"CGRID":           "c87609aa1cb6e9529ab1836cfeeeb0ab7aa7ebaf",
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
