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
	"testing"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sSv1CfgPath     string
	sSv1Cfg         *config.CGRConfig
	sSv1BiRpc       *rpc2.Client
	sSApierRpc      *rpc.Client
	discEvChan      = make(chan *utils.AttrDisconnectSession, 1)
	sSV1RequestType string

	sTestSessionSv1 = []func(t *testing.T){
		testSSv1ItInitCfgDir,
		testSSv1ItInitCfg,
		testSSv1ItResetDataDb,
		testSSv1ItResetStorDb,
		testSSv1ItStartEngine,
		testSSv1ItRpcConn,
		testSSv1ItPing,
		testSSv1ItTPFromFolder,
		testSSv1ItAuth,
		testSSv1ItAuthWithDigest,
		testSSv1ItInitiateSession,
		testSSv1ItUpdateSession,
		testSSv1ItTerminateSession,
		testSSv1ItProcessCDR,
		testSSv1ItProcessEvent,
		testSSv1ItCDRsGetCdrs,
		testSSv1ItForceUpdateSession,
		testSSv1ItDynamicDebit,
		testSSv1ItDeactivateSessions,
		testSSv1ItAuthNotFoundCharger,
		testSSv1ItInitiateSessionNotFoundCharger,

		testSSv1ItInitiateSessionWithDigest, // no need for session terminate because is the last test

		testSSv1ItStopCgrEngine,
	}
)

func testSSv1ItInitCfgDir(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		sessionsConfDIR = "sessions_internal"
	case utils.MetaMySQL:
		sessionsConfDIR = "sessions_mysql"
	case utils.MetaMongo:
		sessionsConfDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
}

func handleDisconnectSession(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	discEvChan <- args
	// free the channel
	<-discEvChan
	*reply = utils.OK
	return nil
}

func handleGetSessionIDs(clnt *rpc2.Client,
	ignParam string, sessionIDs *[]*sessions.SessionID) error {
	return nil
}

func TestSSv1ItWithPrepaid(t *testing.T) {
	if *dbType == utils.MetaPostgres {
		t.SkipNow()
	}
	sSV1RequestType = utils.META_PREPAID
	for _, stest := range sTestSessionSv1 {
		t.Run(sSV1RequestType, stest)
	}
}

func TestSSv1ItWithPostPaid(t *testing.T) {
	if *dbType == utils.MetaPostgres {
		t.SkipNow()
	}
	sSV1RequestType = utils.META_POSTPAID
	for _, stest := range sTestSessionSv1 {
		t.Run(sSV1RequestType, stest)
	}
}

func TestSSv1ItWithRated(t *testing.T) {
	if *dbType == utils.MetaPostgres {
		t.SkipNow()
	}
	sSV1RequestType = utils.META_RATED
	for _, stest := range sTestSessionSv1 {
		t.Run(sSV1RequestType, stest)
	}
}

func TestSSv1ItWithPseudoPrepaid(t *testing.T) {
	if *dbType == utils.MetaPostgres {
		t.SkipNow()
	}
	sSV1RequestType = utils.META_PSEUDOPREPAID
	for _, stest := range sTestSessionSv1 {
		t.Run(sSV1RequestType, stest)
	}
}

func testSSv1ItResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(sSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSSv1ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSSv1ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sSv1CfgPath, 1000); err != nil {
		t.Fatal(err)
	}
}

func testSSv1ItRpcConn(t *testing.T) {
	dummyClnt, err := utils.NewBiJSONrpcClient(sSv1Cfg.SessionSCfg().ListenBijson,
		nil)
	if err != nil {
		t.Fatal(err)
	}
	clntHandlers := map[string]interface{}{
		utils.SessionSv1DisconnectSession:   handleDisconnectSession,
		utils.SessionSv1GetActiveSessionIDs: handleGetSessionIDs,
	}
	if sSv1BiRpc, err = utils.NewBiJSONrpcClient(sSv1Cfg.SessionSCfg().ListenBijson,
		clntHandlers); err != nil {
		t.Fatal(err)
	}
	if sSApierRpc, err = newRPCClient(sSv1Cfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	dummyClnt.Close() // close so we don't get EOF error when disconnecting server
}

func testSSv1ItPing(t *testing.T) {
	var resp string
	if err := sSv1BiRpc.Call(utils.SessionSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSSv1ItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	var loadInst utils.LoadInstance
	if err := sSApierRpc.Call(utils.APIerSv2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSSv1ItAuth(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetRoutes:          true,
		GetAttributes:      true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItAuth",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.Usage:       authUsage,
				},
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := sSv1BiRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != authUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation == "" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eSplrs := &engine.SortedRoutes{
		ProfileID: "ROUTE_ACNT_1001",
		Sorting:   utils.MetaWeight,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
			},
			{
				RouteID: "route2",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
			},
		},
	}
	if !reflect.DeepEqual(eSplrs, rply.Routes) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eSplrs), utils.ToJSON(rply.Routes))
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItAuth",
				Event: map[string]interface{}{
					utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					"OfficeGroup":     "Marketing",
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.SetupTime:   "2018-01-07T17:00:00Z",
					utils.Usage:       300000000000.0,
				},
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
}

func testSSv1ItAuthWithDigest(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
		GetMaxUsage:        true,
		AuthorizeResources: true,
		GetRoutes:          true,
		GetAttributes:      true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItAuth",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.Usage:       authUsage,
				},
			},
		},
	}
	var rply sessions.V1AuthorizeReplyWithDigest
	if err := sSv1BiRpc.Call(utils.SessionSv1AuthorizeEventWithDigest, args, &rply); err != nil {
		t.Fatal(err)
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect -1
	if rply.MaxUsage != authUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation == "" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eSplrs := utils.StringPointer("route1,route2")
	if *eSplrs != *rply.RoutesDigest {
		t.Errorf("expecting: %v, received: %v", *eSplrs, *rply.RoutesDigest)
	}
	eAttrs := utils.StringPointer("OfficeGroup:Marketing")
	if *eAttrs != *rply.AttributesDigest {
		t.Errorf("expecting: %v, received: %v", *eAttrs, *rply.AttributesDigest)
	}
}

func testSSv1ItInitiateSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		GetAttributes:     true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       initUsage,
				},
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect the value of Usage field
	// if this was missing the MaxUsage should be equal to MaxCallDuration from config
	if rply.MaxUsage == nil || *rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession",
				Event: map[string]interface{}{
					utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					"OfficeGroup":     "Marketing",
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.SetupTime:   "2018-01-07T17:00:00Z",
					utils.AnswerTime:  "2018-01-07T17:00:10Z",
					utils.Usage:       300000000000.0,
				},
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	}
}

func testSSv1ItInitiateSessionWithDigest(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		GetAttributes:     true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It2",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       initUsage,
				},
			},
		},
	}
	var rply sessions.V1InitReplyWithDigest
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSessionWithDigest,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect the value of Usage field
	// if this was missing the MaxUsage should be equal to MaxCallDuration from config
	if rply.MaxUsage != initUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := utils.StringPointer("OfficeGroup:Marketing")
	if !reflect.DeepEqual(eAttrs, rply.AttributesDigest) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.AttributesDigest))
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
}

func testSSv1ItUpdateSession(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItUpdateSession",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       reqUsage,
				},
			},
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1UpdateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItUpdateSession",
				Event: map[string]interface{}{
					utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					"OfficeGroup":     "Marketing",
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.SetupTime:   "2018-01-07T17:00:00Z",
					utils.AnswerTime:  "2018-01-07T17:00:10Z",
					utils.Usage:       300000000000.0,
				},
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Fatalf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect the value of Usage field
	// if this was missing the MaxUsage should be equal to MaxCallDuration from config
	if rply.MaxUsage == nil || *rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
}

func testSSv1ItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		ReleaseResources: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItUpdateSession",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       10 * time.Minute,
				},
			},
		},
	}
	var rply string
	if err := sSv1BiRpc.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error %s received error %v and reply %s", utils.ErrNotFound, err, utils.ToJSON(aSessions))
	}
}

func testSSv1ItProcessCDR(t *testing.T) {
	args := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessCDR",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: sSV1RequestType,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply string
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

// TestSSv1ItProcessEvent processes individual event and also checks it's CDRs
func testSSv1ItProcessEvent(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1ProcessMessageArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItProcessEvent",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It2",
					utils.OriginHost:  "TestSSv1It3",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       initUsage,
				},
			},
		},
	}
	var rply sessions.V1ProcessMessageReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect the value of Usage field
	// if this was missing the MaxUsage should be equal to MaxCallDuration from config
	if rply.MaxUsage == nil || *rply.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItProcessEvent",
				Event: map[string]interface{}{
					utils.CGRID:       "f7f5cf1029905f9b98be1a608e4bd975b8e51413",
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					"OfficeGroup":     "Marketing",
					utils.OriginID:    "TestSSv1It2",
					utils.OriginHost:  "TestSSv1It3",
					utils.RequestType: sSV1RequestType,
					utils.SetupTime:   "2018-01-07T17:00:00Z",
					utils.AnswerTime:  "2018-01-07T17:00:10Z",
					utils.Usage:       300000000000.0,
				},
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var rplyCDR string
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessCDR,
		args.CGREvent, &rplyCDR); err != nil {
		t.Error(err)
	}
	if rplyCDR != utils.OK {
		t.Errorf("Unexpected reply: %s", rplyCDR)
	}
}

func testSSv1ItCDRsGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 6 { // 3 for each CDR
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{"raw"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"},
		OriginIDs: []string{"TestSSv1It1"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
		OriginIDs: []string{"TestSSv1It1"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}

	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"},
		OriginIDs: []string{"TestSSv1It2"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.099 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
		OriginIDs: []string{"TestSSv1It2"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.051 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testSSv1ItForceUpdateSession(t *testing.T) {
	if sSV1RequestType != utils.META_PREPAID {
		t.SkipNow()
		return
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(asessions)=%v", err, len(aSessions))
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.55
	if err := sSApierRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItUpdateSession",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       reqUsage,
				},
			},
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1UpdateSession,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItUpdateSession",
				Event: map[string]interface{}{
					utils.CGRID:       "70876773b294f0e1476065f8d18bb9ec6bcb3d5f",
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					"OfficeGroup":     "Marketing",
					utils.OriginID:    "TestSSv1It",
					utils.RequestType: sSV1RequestType,
					utils.SetupTime:   "2018-01-07T17:00:00Z",
					utils.AnswerTime:  "2018-01-07T17:00:10Z",
					utils.Usage:       300000000000.0,
				},
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Fatalf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	aSessions = make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active ssesions: %s", utils.ToJSON(aSessions))
	}

	eAcntVal = 9.4
	if err := sSApierRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	rplyt := ""
	if err := sSv1BiRpc.Call(utils.SessionSv1ForceDisconnect,
		map[string]string{utils.OriginID: "TestSSv1It"}, &rplyt); err != nil {
		t.Error(err)
	} else if rplyt != utils.OK {
		t.Errorf("Unexpected reply: %s", rplyt)
	}
	aSessions = make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := sSApierRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal { // no monetary change bacause the sessin was terminated
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	var cdrs []*engine.CDR
	argsCDR := &utils.RPCCDRsFilterWithOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs:    []string{"CustomerCharges"},
			OriginIDs: []string{"TestSSv1It"},
		},
	}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs), "\n", utils.ToJSON(cdrs))
	} else {
		if cdrs[0].Cost != 0.099 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	argsCDR = &utils.RPCCDRsFilterWithOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs:    []string{"SupplierCharges"},
			OriginIDs: []string{"TestSSv1It"},
		},
	}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.051 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testSSv1ItDynamicDebit(t *testing.T) {
	if sSV1RequestType != utils.META_PREPAID {
		t.SkipNow()
		return
	}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "TestDynamicDebit",
		BalanceType: utils.VOICE,
		Value:       2 * float64(time.Second),
		Balance: map[string]interface{}{
			utils.ID:            "TestDynamicDebitBalance",
			utils.RatingSubject: "*zero5ms",
		},
	}
	var reply string
	if err := sSApierRpc.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  attrSetBalance.Tenant,
		Account: attrSetBalance.Account,
	}
	eAcntVal := 2 * float64(time.Second)
	if err := sSApierRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}

	args1 := &sessions.V1InitSessionArgs{
		InitSession:   true,
		GetAttributes: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsDebitInterval: 30 * time.Millisecond,
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession2",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.Category:    "call",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestDynamicTDebit",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "TestDynamicDebit",
					utils.Subject:     "TEST",
					utils.Destination: "TEST",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       0,
				},
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if rply1.MaxUsage == nil || *rply1.MaxUsage != 3*time.Hour /* MaxCallDuration from config*/ {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %+v , %s ", len(aSessions), utils.ToJSON(aSessions))
	}
	time.Sleep(time.Millisecond)
	eAcntVal -= float64(time.Millisecond) * 30 * 2 // 2 session
	if err := sSApierRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}

	time.Sleep(10 * time.Millisecond)
	if err := sSApierRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}
	time.Sleep(20 * time.Millisecond)
	eAcntVal -= float64(time.Millisecond) * 30 * 2 // 2 session
	if err := sSApierRpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}

	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %+v , %s ", len(aSessions), utils.ToJSON(aSessions))
	}

	var rplyt string
	if err := sSv1BiRpc.Call(utils.SessionSv1ForceDisconnect,
		nil, &rplyt); err != nil {
		t.Error(err)
	} else if rplyt != utils.OK {
		t.Errorf("Unexpected reply: %s", rplyt)
	}

	time.Sleep(50 * time.Millisecond)

	aSessions = make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSSv1ItDeactivateSessions(t *testing.T) {
	aSessions := make([]*sessions.ExternalSession, 0)
	pSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error(err)
	}
	if err := sSv1BiRpc.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &pSessions); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error(err)
	}
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession:   true,
		GetAttributes: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession",
				Event: map[string]interface{}{
					utils.Tenant:      "cgrates.org",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "TestSSv1It1",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       initUsage,
				},
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSession, args, &rply); err != nil {
		t.Fatal(err)
	}
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	}
	if err := sSv1BiRpc.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &pSessions); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error(err)
	}
	var reply string
	err := sSv1BiRpc.Call(utils.SessionSv1DeactivateSessions, &utils.SessionIDsWithArgsDispatcher{}, &reply)
	if err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: OK, received : %+v", reply)
	}
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil && err.Error() != utils.NotFoundCaps {
		t.Error(err)
	}
	if err := sSv1BiRpc.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 3 {
		t.Errorf("Expecting: 2, received: %+v", len(pSessions))
	}
}

func testSSv1ItAuthNotFoundCharger(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "Unexist",
				ID:     "testSSv1ItAuthNotFoundCharger",
				Event: map[string]interface{}{
					utils.Tenant:      "Unexist",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "testSSv1ItAuthNotFoundCharger",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.Usage:       authUsage,
				},
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := sSv1BiRpc.Call(utils.SessionSv1AuthorizeEvent, args,
		&rply); err == nil || err.Error() != utils.NewErrChargerS(utils.ErrNotFound).Error() {
		t.Errorf("Expecting: %+v, received: %+v", utils.NewErrChargerS(utils.ErrNotFound), err)
	}
}

func testSSv1ItInitiateSessionNotFoundCharger(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "Unexist",
				ID:     "testSSv1ItInitiateSessionNotFoundCharger",
				Event: map[string]interface{}{
					utils.Tenant:      "Unexist",
					utils.ToR:         utils.VOICE,
					utils.OriginID:    "testSSv1ItInitiateSessionNotFoundCharger",
					utils.RequestType: sSV1RequestType,
					utils.Account:     "1001",
					utils.Subject:     "ANY2CNT",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       initUsage,
				},
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSession,
		args, &rply); err == nil || err.Error() != utils.NewErrChargerS(utils.ErrNotFound).Error() {
		t.Errorf("Expecting: %+v, received: %+v", utils.NewErrChargerS(utils.ErrNotFound), err)
	}
}

func testSSv1ItStopCgrEngine(t *testing.T) {
	if err := sSv1BiRpc.Close(); err != nil { // Close the connection so we don't get EOF warnings from client
		t.Error(err)
	}
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
