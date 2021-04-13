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
	"testing"
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

/*
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

	testDspSessionGetCost,
	testDspSessionReplicate,
	testDspSessionPassive,

	testDspSessionSTIRAuthenticate,
	testDspSessionSTIRIdentity,
	testDspSessionForceDisconect,
}

//Test start here
func TestDspSessionS(t *testing.T) {
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
		config3 += "_gob"
	}
	testDsp(t, sTestsDspSession, "TestDspSessionS", config1, config2, config3, "testit", "tutorial", dispDIR)
}

func testDspSessionAddBalacne(t *testing.T) {
	initUsage := 40 * time.Minute
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MetaVoice,
		Value:       float64(initUsage),
		Balance: map[string]interface{}{
			utils.ID:            "SessionBalance",
			utils.RatingSubject: "*zero5ms",
		},
	}
	var reply string
	if err := allEngine.RPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
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
	if err := allEngine.RPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.MetaVoice].GetTotalValue()))
	}
	if err := allEngine2.RPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if err := allEngine2.RPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.MetaVoice].GetTotalValue()))
	}
}

func testDspSessionPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.SessionSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.SessionSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspSessionPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.SessionSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.SessionSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.SessionSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(utils.SessionSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspSessionTestAuthKey(t *testing.T) {
	authUsage := 5 * time.Minute
	args := sessions.V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetRoutes:          true,
		GetAttributes:      true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:        authUsage,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "12345",
			},
		},
	}
	var rply sessions.V1AuthorizeReplyWithDigest
	if err := dispEngine.RPC.Call(utils.SessionSv1AuthorizeEventWithDigest,
		args, &rply); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspSessionAuthorize(t *testing.T) {
	authUsage := 5 * time.Minute
	argsAuth := &sessions.V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetRoutes:          true,
		GetAttributes:      true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth",
			Event: map[string]interface{}{
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:        authUsage,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey:             "ses12345",
				utils.OptsRouteProfilesCount: 1.,
			},
		},
	}
	var rply sessions.V1AuthorizeReplyWithDigest
	if err := dispEngine.RPC.Call(utils.SessionSv1AuthorizeEventWithDigest,
		argsAuth, &rply); err != nil {
		t.Error(err)
		return
	}
	if rply.MaxUsage != authUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation == "" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eSplrs := "route1,route2"
	tp := strings.Split(*rply.RoutesDigest, ",")
	sort.Strings(tp)
	*rply.RoutesDigest = strings.Join(tp, ",")
	if eSplrs != *rply.RoutesDigest {
		t.Errorf("expecting: %v, received: %v", eSplrs, *rply.RoutesDigest)
	}
	eAttrs := "OfficeGroup:Marketing"
	if eAttrs != *rply.AttributesDigest {
		t.Errorf("expecting: %v, received: %v", eAttrs, *rply.AttributesDigest)
	}
}

func testDspSessionInit(t *testing.T) {
	initUsage := 5 * time.Minute
	argsInit := &sessions.V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		GetAttributes:     true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItInitiateSession",
			Event: map[string]interface{}{
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        initUsage,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
		},
	}
	var rply sessions.V1InitReplyWithDigest
	if err := dispEngine.RPC.Call(utils.SessionSv1InitiateSessionWithDigest,
		argsInit, &rply); err != nil {
		t.Fatal(err)
	}
	if rply.MaxUsage != initUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
}

func testDspGetSessions(t *testing.T) {
	filtr := utils.SessionFilter{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
		Tenant:  "cgrates.org",
		Filters: []string{},
	}
	var reply int
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessionsCount,
		&filtr, &reply); err != nil {
		t.Fatal(err)
	} else if reply != 3 {
		t.Errorf("Expected 3 active sessions received %v", reply)
	}
	var rply []*sessions.ExternalSession
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessions,
		&filtr, &rply); err != nil {
		t.Fatal(err)
	} else if len(rply) != 3 {
		t.Errorf("Unexpected number of sessions returned %v :%s", len(rply), utils.ToJSON(rply))
	}

	if err := dispEngine.RPC.Call(utils.SessionSv1GetPassiveSessionsCount,
		&filtr, &reply); err != nil {
		t.Fatal(err)
	} else if reply != 0 {
		t.Errorf("Expected no pasive sessions received %v", reply)
	}
	rply = nil
	if err := dispEngine.RPC.Call(utils.SessionSv1GetPassiveSessions,
		&filtr, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected %v received %v with reply %s", utils.ErrNotFound, err, utils.ToJSON(rply))
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
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        reqUsage,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := dispEngine.RPC.Call(utils.SessionSv1UpdateSession,
		argsUpdate, &rply); err != nil {
		t.Error(err)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				"OfficeGroup":      "Marketing",
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    "2018-01-07T17:00:00Z",
				utils.AnswerTime:   "2018-01-07T17:00:10Z",
				utils.Usage:        float64(reqUsage),
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
				utils.Subsys:     utils.MetaSessionS,
			},
		},
	}
	if *encoding == utils.MetaGOB { // gob maintains the variable type
		eAttrs.CGREvent.Event[utils.Usage] = reqUsage
		eAttrs.CGREvent.Event[utils.SetupTime] = argsUpdate.CGREvent.Event[utils.SetupTime]
		eAttrs.CGREvent.Event[utils.AnswerTime] = argsUpdate.CGREvent.Event[utils.AnswerTime]
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
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        reqUsage,
			},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := dispEngine.RPC.Call(utils.SessionSv1UpdateSession,
		argsUpdate, &rply); err != nil {
		t.Fatal(err)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1001_SESSIONAUTH"},
		AlteredFields:   []string{"*req.LCRProfile", "*req.Password", "*req.RequestType", "*req.PaypalAccount"},

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				"LCRProfile":       "premium_cli",
				"Password":         "CGRateS.org",
				"PaypalAccount":    "cgrates@paypal.com",
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    "2018-01-07T17:00:00Z",
				utils.AnswerTime:   "2018-01-07T17:00:10Z",
				utils.Usage:        float64(reqUsage),
			},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
				utils.Subsys:     utils.MetaSessionS,
			},
		},
	}
	sort.Strings(eAttrs.AlteredFields)
	if *encoding == utils.MetaGOB { // gob maintains the variable type
		eAttrs.CGREvent.Event[utils.Usage] = reqUsage
		eAttrs.CGREvent.Event[utils.SetupTime] = argsUpdate.CGREvent.Event[utils.SetupTime]
		eAttrs.CGREvent.Event[utils.AnswerTime] = argsUpdate.CGREvent.Event[utils.AnswerTime]
	}
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
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
		},
	}
	var rply string
	if err := dispEngine.RPC.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testDspSessionProcessCDR(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestSSv1ItProcessCDR",
		Event: map[string]interface{}{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestSSv1It1",
			utils.RequestType:  utils.MetaPostpaid,
			utils.AccountField: "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        10 * time.Minute,
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
	}

	var rply string
	if err := dispEngine.RPC.Call(utils.SessionSv1ProcessCDR,
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
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebac",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginHost:   "disp",
				utils.OriginID:     "TestSSv1It2",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        initUsage,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
		},
	}
	var rply sessions.V1ProcessMessageReply
	if err := dispEngine.RPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebac",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				"OfficeGroup":      "Marketing",
				utils.OriginHost:   "disp",
				utils.OriginID:     "TestSSv1It2",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    "2018-01-07T17:00:00Z",
				utils.AnswerTime:   "2018-01-07T17:00:10Z",
				utils.Usage:        300000000000.0,
			},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
				utils.Subsys:     utils.MetaSessionS,
			},
		},
	}
	if *encoding == utils.MetaGOB { // gob maintains the variable type
		eAttrs.CGREvent.Event[utils.Usage] = initUsage
		eAttrs.CGREvent.Event[utils.SetupTime] = args.CGREvent.Event[utils.SetupTime]
		eAttrs.CGREvent.Event[utils.AnswerTime] = args.CGREvent.Event[utils.AnswerTime]
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
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It2",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        initUsage,
				utils.EventName:    "Internal",
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "pse12345",
			},
		},
	}
	var rply sessions.V1ProcessMessageReply
	if err := dispEngine.RPC.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1001_SIMPLEAUTH"},
		AlteredFields:   []string{"*req.EventName", "*req.Password"},

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				"Password":         "CGRateS.org",
				utils.OriginID:     "TestSSv1It2",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    "2018-01-07T17:00:00Z",
				utils.AnswerTime:   "2018-01-07T17:00:10Z",
				utils.Usage:        300000000000.0,
			},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "pse12345",
				utils.Subsys:     utils.MetaSessionS,
			},
		},
	}
	if *encoding == utils.MetaGOB { // gob maintains the variable type
		eAttrs.CGREvent.Event[utils.Usage] = initUsage
		eAttrs.CGREvent.Event[utils.SetupTime] = args.CGREvent.Event[utils.SetupTime]
		eAttrs.CGREvent.Event[utils.AnswerTime] = args.CGREvent.Event[utils.AnswerTime]
	}
	sort.Strings(rply.Attributes.AlteredFields)
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
}

func testDspSessionReplicate(t *testing.T) {
	allEngine.initDataDb(t)
	allEngine.resetStorDb(t)
	var reply string
	// reload cache  in order to corectly cahce the indexes
	if err := allEngine.RPC.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	allEngine.loadData(t, path.Join(*dataDir, "tariffplans", "testit"))
	testDspSessionAddBalacne(t)
	testDspSessionAuthorize(t)
	testDspSessionInit(t)

	if err := dispEngine.RPC.Call(utils.SessionSv1ReplicateSessions, &ArgsReplicateSessionsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
		Tenant: "cgrates.org",
		ArgsReplicateSessions: sessions.ArgsReplicateSessions{
			CGRID:   "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
			Passive: false,
			ConnIDs: []string{"rplConn"},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}

	var repl int
	time.Sleep(10 * time.Millisecond)
	if err := allEngine2.RPC.Call(utils.SessionSv1GetPassiveSessionsCount,
		new(utils.SessionFilter), &repl); err != nil {
		t.Fatal(err)
	} else if repl != 3 {
		t.Errorf("Expected 3 sessions received %v", repl)
	}
}

func testDspSessionPassive(t *testing.T) {
	allEngine.stopEngine(t)
	testDspSessionUpdate2(t)
	var repl int
	filtr := utils.SessionFilter{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
		Tenant:  "cgrates.org",
		Filters: []string{},
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RPC.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no passive sessions received %v", repl)
	}
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 3 {
		t.Errorf("Expected 3 active sessions received %v", repl)
	}

	var rply []*sessions.ExternalSession
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessions,
		&filtr, &rply); err != nil {
		t.Fatal(err)
	} else if len(rply) != 3 {
		t.Errorf("Unexpected number of sessions returned %v :%s", len(rply), utils.ToJSON(rply))
	}

	var reply string
	if err := dispEngine.RPC.Call(utils.SessionSv1SetPassiveSession, sessions.Session{
		CGRID:      rply[0].CGRID,
		Tenant:     rply[0].Tenant,
		ResourceID: "TestSSv1It1",
		EventStart: engine.NewMapEvent(map[string]interface{}{
			utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestSSv1It1",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        5 * time.Minute,
		}),
		SRuns: []*sessions.SRun{
			{
				Event: engine.NewMapEvent(map[string]interface{}{
					"RunID":            "CustomerCharges",
					utils.CGRID:        "c87609aa1cb6e9529ab1836cfeeebaab7aa7ebaf",
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestSSv1It1",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:        5 * time.Minute,
				}),
				CD:        &engine.CallDescriptor{},
				EventCost: &engine.EventCost{},

				LastUsage:  5 * time.Minute,
				TotalUsage: 10 * time.Minute,
			},
		},
		OptsStart: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RPC.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 1 {
		t.Errorf("Expected 1 passive sessions received %v", repl)
	}
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no active sessions received %v", repl)
	}
}

func testDspSessionForceDisconect(t *testing.T) {
	allEngine.startEngine(t)
	allEngine.initDataDb(t)
	allEngine.resetStorDb(t)
	allEngine.loadData(t, path.Join(*dataDir, "tariffplans", "testit"))
	testDspSessionAddBalacne(t)
	testDspSessionAuthorize(t)
	testDspSessionInit(t)
	var repl int
	filtr := utils.SessionFilter{
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
		Tenant:  "cgrates.org",
		Filters: []string{},
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RPC.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no passive sessions received %v", repl)
	}
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 3 {
		t.Errorf("Expected 3 active sessions received %v", repl)
	}

	var rply []*sessions.ExternalSession
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessions,
		&filtr, &rply); err != nil {
		t.Fatal(err)
	} else if len(rply) != 3 {
		t.Errorf("Unexpected number of sessions returned %v :%s", len(rply), utils.ToJSON(rply))
	}

	var reply string
	if err := dispEngine.RPC.Call(utils.SessionSv1ForceDisconnect, &filtr, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	if err := dispEngine.RPC.Call(utils.SessionSv1GetPassiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected 1 passive sessions received %v", repl)
	}
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessionsCount,
		filtr, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no active sessions received %v", repl)
	}
}

func testDspSessionProcessEvent3(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{"*rals:*terminate", "*resources:*release"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventTerminateSession",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestSSv1It2",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "pse12345",
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := dispEngine.RPC.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}

	var repl int
	if err := dispEngine.RPC.Call(utils.SessionSv1GetActiveSessionsCount,
		utils.SessionFilter{
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
			Tenant:  "cgrates.org",
			Filters: []string{},
		}, &repl); err != nil {
		t.Fatal(err)
	} else if repl != 0 {
		t.Errorf("Expected no active sessions received %v", repl)
	}
}

func testDspSessionGetCost(t *testing.T) {

	args := &sessions.V1ProcessEventArgs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItGetCost",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.MetaMonetary,
				utils.OriginID:    "testSSv1ItProcessEventWithGetCost",
				utils.RequestType: utils.MetaPrepaid,
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
		},
	}

	var rply sessions.V1GetCostReply
	if err := dispEngine.RPC.Call(utils.SessionSv1GetCost,
		args, &rply); err != nil {
		t.Error(err)
	}

	if rply.EventCost == nil {
		t.Errorf("Received nil EventCost")
	} else if *rply.EventCost.Cost != 0.198 { // same cost as in CDR
		t.Errorf("Expected: %+v,received: %+v", 0.198, *rply.EventCost.Cost)
	} else if *rply.EventCost.Usage != 10*time.Minute {
		t.Errorf("Expected: %+v,received: %+v", 10*time.Minute, *rply.EventCost.Usage)
	}

}

func testDspSessionSTIRAuthenticate(t *testing.T) {
	var rply string
	if err := dispEngine.RPC.Call(utils.SessionSv1STIRAuthenticate,
		&sessions.V1STIRAuthenticateArgs{
			Attest:             []string{"A"},
			PayloadMaxDuration: "-1",
			DestinationTn:      "1002",
			Identity:           "eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiL3Vzci9zaGFyZS9jZ3JhdGVzL3N0aXIvc3Rpcl9wdWJrZXkucGVtIn0.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMzg4MDIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.cMEMlFnfyTu8uxfeU4RoZTamA7ifFT9Ibwrvi1_LKwL2xAU6fZ_CSIxKbtyOpNhM_sV03x7CfA_v0T4sHkifzg;info=</usr/share/cgrates/stir/stir_pubkey.pem>;ppt=shaken",
			OriginatorTn:       "1001",
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "ses12345",
			},
		}, &rply); err != nil {
		t.Fatal(err)
	} else if rply != utils.OK {
		t.Errorf("Expected: %s ,received: %s", utils.OK, rply)
	}
}
*/
func testDspSessionSTIRIdentity(t *testing.T) {
	payload := &utils.PASSporTPayload{
		Dest:   utils.PASSporTDestinationsIdentity{Tn: []string{"1002"}},
		IAT:    1587019822,
		Orig:   utils.PASSporTOriginsIdentity{Tn: "1001"},
		OrigID: "123456",
	}
	args := &sessions.V1STIRIdentityArgs{
		Payload:        payload,
		PublicKeyPath:  "/usr/share/cgrates/stir/stir_pubkey.pem",
		PrivateKeyPath: "/usr/share/cgrates/stir/stir_privatekey.pem",
		OverwriteIAT:   true,
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "ses12345",
		},
	}
	var rply string
	if err := dispEngine.RPC.Call(utils.SessionSv1STIRIdentity,
		args, &rply); err != nil {
		t.Error(err)
	}
	if err := sessions.AuthStirShaken(rply, "1001", "", "1002", "", utils.NewStringSet([]string{"A"}), 10*time.Minute); err != nil {
		t.Fatal(err)
	}
}
