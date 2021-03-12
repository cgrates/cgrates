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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

//Use from sessionsv1_it_test.go
//functions insted of duplicate them here
// eg: initCfg,ResetDB,StopEngine,etc...
var sTestSessionSv1ProcessEvent = []func(t *testing.T){
	testSSv1ItInitCfg,
	testSSv1ItResetDataDb,
	testSSv1ItResetStorDb,
	testSSv1ItStartEngine,
	testSSv1ItRpcConn,
	testSSv1ItPing,
	testSSv1ItTPFromFolder,
	testSSv1ItProcessEventAuth,
	testSSv1ItProcessEventInitiateSession,
	testSSv1ItProcessEventUpdateSession,
	testSSv1ItProcessEventTerminateSession,
	testSSv1ItProcessCDRForSessionFromProcessEvent,
	testSSv1ItGetCDRs,
	testSSv1ItProcessEventWithGetCost,
	testSSv1ItProcessEventWithGetCost2,
	testSSv1ItProcessEventWithGetCost3,
	testSSv1ItProcessEventWithGetCost4,
	testSSv1ItGetCost,
	testSSv1ItProcessEventWithCDR,
	testSSv1ItGetCDRsFromProcessEvent,
	testSSv1ItProcessEventWithCDRResourceError,
	testSSv1ItGetCDRsFromProcessEventResourceError,
	testSSv1ItProcessEventWithCDRResourceErrorBlockError,
	testSSv1ItGetCDRsFromProcessEventResourceErrorBlockError,
	testSSv1ItStopCgrEngine,
}

func TestSSv1ItProcessEventWithPrepaid(t *testing.T) {
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

	sSV1RequestType = utils.MetaPrepaid
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sessionsConfDIR+utils.EmptyString+sSV1RequestType, stest)
	}
}

func TestSSv1ItProcessEventWithPostPaid(t *testing.T) {
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

	sSV1RequestType = utils.MetaPostpaid
	sTestSessionSv1ProcessEvent = append(sTestSessionSv1ProcessEvent[:len(sTestSessionSv1ProcessEvent)-7], testSSv1ItStopCgrEngine)
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sessionsConfDIR+utils.EmptyString+sSV1RequestType, stest)
	}
}

func TestSSv1ItProcessEventWithRated(t *testing.T) {
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

	sSV1RequestType = utils.MetaRated
	sTestSessionSv1ProcessEvent = append(sTestSessionSv1ProcessEvent[:len(sTestSessionSv1ProcessEvent)-7], testSSv1ItStopCgrEngine)
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sessionsConfDIR+utils.EmptyString+sSV1RequestType, stest)
	}
}

func TestSSv1ItProcessEventWithPseudoPrepaid(t *testing.T) {
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

	sSV1RequestType = utils.MetaPseudoPrepaid
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sessionsConfDIR+utils.EmptyString+sSV1RequestType, stest)
	}
}

func testSSv1ItInitCfg(t *testing.T) {
	var err error
	sSv1CfgPath = path.Join(*dataDir, "conf", "samples", sessionsConfDIR)
	// Init config first
	sSv1Cfg, err = config.NewCGRConfigFromPath(sSv1CfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testSSv1ItProcessEventAuth(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.ConcatenatedKey(utils.MetaResources, utils.MetaAuthorize),
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaDerivedReply),
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaAuthorize),
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
			utils.MetaRoutes, utils.MetaAttributes, utils.MetaChargers},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventAuth",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSSv1ItProcessEvent",
				utils.RequestType:  sSV1RequestType,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:        authUsage,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	expMaxUsage := map[string]time.Duration{
		"CustomerCharges": authUsage,
		"SupplierCharges": authUsage,
		"raw":             authUsage,
		utils.MetaRaw:     authUsage,
	}
	if !reflect.DeepEqual(expMaxUsage, rply.MaxUsage) {
		t.Errorf("Expected %s received %s", expMaxUsage, rply.MaxUsage)
	}
	if rply.ResourceAllocation == nil || rply.ResourceAllocation["CustomerCharges"] == utils.EmptyString {
		t.Errorf("Unexpected ResourceAllocation: %s", rply.ResourceAllocation)
	}
	eSplrs := &engine.SortedRoutes{
		ProfileID: "ROUTE_ACNT_1001",
		Sorting:   utils.MetaWeight,
		Count:     2,
		Routes: []*engine.SortedRoute{
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
	if !reflect.DeepEqual(eSplrs, rply.Routes[utils.MetaRaw]) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eSplrs), utils.ToJSON(rply.Routes[utils.MetaRaw]))
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventAuth",
			Event: map[string]interface{}{
				utils.CGRID:        "4be779c004d9f784e836db9ffd41b50319d71fe8",
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				"OfficeGroup":      "Marketing",
				utils.OriginID:     "testSSv1ItProcessEvent",
				utils.RequestType:  sSV1RequestType,
				utils.SetupTime:    "2018-01-07T17:00:00Z",
				utils.Usage:        300000000000.0,
			},
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes[utils.MetaRaw]) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes[utils.MetaRaw]))
	}
}

func testSSv1ItProcessEventInitiateSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.ConcatenatedKey(utils.MetaRALs, utils.MetaInitiate),
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaAllocate),
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaDerivedReply),
			utils.MetaAttributes, utils.MetaChargers},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSSv1ItProcessEvent",
				utils.RequestType:  sSV1RequestType,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        initUsage,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect the value of Usage field
	// if this was missing the MaxUsage should be equal to MaxCallDuration from config
	expMaxUsage := map[string]time.Duration{
		"CustomerCharges": initUsage,
		"SupplierCharges": initUsage,
		// "raw":             initUsage,
		utils.MetaRaw: initUsage,
	}
	if !reflect.DeepEqual(expMaxUsage, rply.MaxUsage) {
		t.Errorf("Expected %s received %s", expMaxUsage, rply.MaxUsage)
	}
	if rply.ResourceAllocation == nil || rply.ResourceAllocation["CustomerCharges"] != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", rply.ResourceAllocation)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventInitiateSession",
			Event: map[string]interface{}{
				utils.CGRID:        "4be779c004d9f784e836db9ffd41b50319d71fe8",
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				"OfficeGroup":      "Marketing",
				utils.OriginID:     "testSSv1ItProcessEvent",
				utils.RequestType:  sSV1RequestType,
				utils.SetupTime:    "2018-01-07T17:00:00Z",
				utils.AnswerTime:   "2018-01-07T17:00:10Z",
				utils.Usage:        300000000000.0,
			},
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes[utils.MetaRaw]) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes[utils.MetaRaw]))
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	}
}

func testSSv1ItProcessEventUpdateSession(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.ConcatenatedKey(utils.MetaRALs, utils.MetaUpdate),
			utils.ConcatenatedKey(utils.MetaRALs, utils.MetaDerivedReply),
			utils.MetaAttributes},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSSv1ItProcessEvent",
				utils.RequestType:  sSV1RequestType,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        reqUsage,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"*req.OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventUpdateSession",
			Event: map[string]interface{}{
				utils.CGRID:        "4be779c004d9f784e836db9ffd41b50319d71fe8",
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				"OfficeGroup":      "Marketing",
				utils.OriginID:     "testSSv1ItProcessEvent",
				utils.RequestType:  sSV1RequestType,
				utils.SetupTime:    "2018-01-07T17:00:00Z",
				utils.AnswerTime:   "2018-01-07T17:00:10Z",
				utils.Usage:        300000000000.0,
			},
			Opts: map[string]interface{}{utils.Subsys: utils.MetaSessionS},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes[utils.MetaRaw]) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes[utils.MetaRaw]))
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect the value of Usage field
	// if this was missing the MaxUsage should be equal to MaxCallDuration from config
	expMaxUsage := map[string]time.Duration{
		"CustomerCharges": reqUsage,
		"SupplierCharges": reqUsage,
		// "raw":             reqUsage,
		utils.MetaRaw: reqUsage,
	}
	if !reflect.DeepEqual(expMaxUsage, rply.MaxUsage) {
		t.Errorf("Expected %s received %s", expMaxUsage, rply.MaxUsage)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
}

func testSSv1ItProcessEventTerminateSession(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.ConcatenatedKey(utils.MetaRALs, utils.MetaTerminate),
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaRelease)},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventTerminateSession",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSSv1ItProcessEvent",
				utils.RequestType:  sSV1RequestType,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSSv1ItProcessCDRForSessionFromProcessEvent(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSSv1ItProcessCDRForSessionFromProcessEvent",
		Event: map[string]interface{}{
			utils.Tenant:       "cgrates.org",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "testSSv1ItProcessEvent",
			utils.RequestType:  sSV1RequestType,
			utils.AccountField: "1001",
			utils.Subject:      "ANY2CNT",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        10 * time.Minute,
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

func testSSv1ItGetCDRs(t *testing.T) {
	var cdrCnt int64
	req := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 { // 3 for each CDR
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{"raw"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"},
		OriginIDs: []string{"testSSv1ItProcessEvent"}}}
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
		OriginIDs: []string{"testSSv1ItProcessEvent"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testSSv1ItProcessEventWithGetCost(t *testing.T) {
	// GetCost for ANY2CNT Subject
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaAttributes, utils.ConcatenatedKey(utils.MetaRALs, utils.MetaCost)},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventWithGetCost",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.MetaMonetary,
				utils.OriginID:    "testSSv1ItProcessEventWithGetCost",
				utils.RequestType: sSV1RequestType,
				utils.Subject:     "*attributes",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.Attributes == nil {
		t.Error("Received nil Attributes")
	} else if !reflect.DeepEqual(rply.Attributes[utils.MetaRaw].MatchedProfiles, []string{"ATTR_SUBJECT_CASE1"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"ATTR_SUBJECT_CASE1"}, rply.Attributes[utils.MetaRaw].MatchedProfiles)
	} else if !reflect.DeepEqual(rply.Attributes[utils.MetaRaw].AlteredFields, []string{"*req.Subject"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"*req.Subject"}, rply.Attributes[utils.MetaRaw].AlteredFields)
	}
	if rply.Cost == nil {
		t.Errorf("Received nil Cost")
	} else if rply.Cost[utils.MetaRaw] != 0.198 { // same cost as in CDR
		t.Errorf("Expected: %+v,received: %+v", 0.198, rply.Cost[utils.MetaRaw])
	}
}

func testSSv1ItProcessEventWithGetCost2(t *testing.T) {
	// GetCost for SPECIAL_1002 Subject
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaAttributes, utils.ConcatenatedKey(utils.MetaRALs, utils.MetaCost)},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventWithGetCost2",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.MetaMonetary,
				utils.OriginID:    "testSSv1ItProcessEventWithGetCost2",
				utils.RequestType: sSV1RequestType,
				utils.Subject:     "*attributes",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.Attributes == nil {
		t.Error("Received nil Attributes")
	} else if !reflect.DeepEqual(rply.Attributes[utils.MetaRaw].MatchedProfiles, []string{"ATTR_SUBJECT_CASE2"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"ATTR_SUBJECT_CASE2"}, rply.Attributes[utils.MetaRaw].MatchedProfiles)
	} else if !reflect.DeepEqual(rply.Attributes[utils.MetaRaw].AlteredFields, []string{"*req.Subject"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"*req.Subject"}, rply.Attributes[utils.MetaRaw].AlteredFields)
	}
	if rply.Cost == nil {
		t.Errorf("Received nil Cost")
	} else if rply.Cost[utils.MetaRaw] != 0.102 { // same cost as in CDR
		t.Errorf("Expected: %+v,received: %+v", 0.102, rply.Cost[utils.MetaRaw])
	}
}

func testSSv1ItProcessEventWithGetCost3(t *testing.T) {
	// GetCost for RP_RETAIL Subject
	// 0.8 connect fee + 0.4 for first minute
	// for the 9 minutes remaining apply
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaAttributes, utils.ConcatenatedKey(utils.MetaRALs, utils.MetaCost)},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventWithGetCost3",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.MetaMonetary,
				utils.OriginID:    "testSSv1ItProcessEventWithGetCost3",
				utils.RequestType: sSV1RequestType,
				utils.Subject:     "*attributes",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.Attributes == nil {
		t.Error("Received nil Attributes")
	} else if !reflect.DeepEqual(rply.Attributes[utils.MetaRaw].MatchedProfiles, []string{"ATTR_SUBJECT_CASE3"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"ATTR_SUBJECT_CASE3"}, rply.Attributes[utils.MetaRaw].MatchedProfiles)
	} else if !reflect.DeepEqual(rply.Attributes[utils.MetaRaw].AlteredFields, []string{"*req.Subject"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"*req.Subject"}, rply.Attributes[utils.MetaRaw].AlteredFields)
	}
	if rply.Cost == nil {
		t.Errorf("Received nil Cost")
	} else if rply.Cost[utils.MetaRaw] != 2.9999 {
		t.Errorf("Expected: %+v,received: %+v", 2.9999, rply.Cost[utils.MetaRaw])
	}
}

func testSSv1ItProcessEventWithGetCost4(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaAttributes, utils.ConcatenatedKey(utils.MetaRALs, utils.MetaCost)},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventWithGetCost4",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.MetaMonetary,
				utils.OriginID:    "testSSv1ItProcessEventWithGetCost4",
				utils.RequestType: sSV1RequestType,
				utils.Subject:     "*attributes",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err == nil || err.Error() != utils.ErrRatingPlanNotFound.Error() {
		t.Error(err)
	}

}

func testSSv1ItGetCost(t *testing.T) {
	// GetCost for ANY2CNT Subject
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaAttributes},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItGetCost",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.MetaMonetary,
				utils.OriginID:    "testSSv1ItProcessEventWithGetCost",
				utils.RequestType: sSV1RequestType,
				utils.Subject:     "*attributes",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply sessions.V1GetCostReply
	if err := sSv1BiRpc.Call(utils.SessionSv1GetCost,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.Attributes == nil {
		t.Error("Received nil Attributes")
	} else if !reflect.DeepEqual(rply.Attributes.MatchedProfiles, []string{"ATTR_SUBJECT_CASE1"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"ATTR_SUBJECT_CASE1"}, rply.Attributes.MatchedProfiles)
	} else if !reflect.DeepEqual(rply.Attributes.AlteredFields, []string{"*req.Subject"}) {
		t.Errorf("Expected: %+v,received: %+v", []string{"*req.Subject"}, rply.Attributes.AlteredFields)
	}
	if rply.EventCost == nil {
		t.Errorf("Received nil EventCost")
	} else if *rply.EventCost.Cost != 0.198 { // same cost as in CDR
		t.Errorf("Expected: %+v,received: %+v", 0.198, *rply.EventCost.Cost)
	} else if *rply.EventCost.Usage != 10*time.Minute {
		t.Errorf("Expected: %+v,received: %+v", 10*time.Minute, *rply.EventCost.Usage)
	}
}

func testSSv1ItProcessEventWithCDR(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaCDRs + utils.InInFieldSep + utils.MetaRALs}, // *cdrs:*rals
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventWithCDR",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSSv1ItProcessEventWithCDR",
				utils.RequestType:  sSV1RequestType,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
}

func testSSv1ItGetCDRsFromProcessEvent(t *testing.T) {
	var cdrCnt int64
	req := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		OriginIDs: []string{"testSSv1ItProcessEventWithCDR"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 { // 3 for each CDR
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		OriginIDs: []string{"testSSv1ItProcessEventWithCDR"},
		RunIDs:    []string{"raw"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		RunIDs:    []string{"CustomerCharges"},
		OriginIDs: []string{"testSSv1ItProcessEventWithCDR"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		RunIDs:    []string{"SupplierCharges"},
		OriginIDs: []string{"testSSv1ItProcessEventWithCDR"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testSSv1ItProcessEventWithCDRResourceError(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaCDRs + utils.InInFieldSep + utils.MetaRALs,
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaRelease)}, // force a resource error and expect that the cdr to be written
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventWithCDRResourceError",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSSv1ItProcessEventWithCDRResourceError",
				utils.RequestType:  sSV1RequestType,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Error(err)
	}
}

func testSSv1ItGetCDRsFromProcessEventResourceError(t *testing.T) {
	var cdrCnt int64
	req := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		OriginIDs: []string{"testSSv1ItProcessEventWithCDRResourceError"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 { // 3 for each CDR
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}

	var cdrs []*engine.CDR
	args := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		OriginIDs: []string{"testSSv1ItProcessEventWithCDRResourceError"},
		RunIDs:    []string{"raw"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		RunIDs:    []string{"CustomerCharges"},
		OriginIDs: []string{"testSSv1ItProcessEventWithCDRResourceError"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		RunIDs:    []string{"SupplierCharges"},
		OriginIDs: []string{"testSSv1ItProcessEventWithCDRResourceError"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func testSSv1ItProcessEventWithCDRResourceErrorBlockError(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaCDRs + utils.InInFieldSep + utils.MetaRALs,
			utils.ConcatenatedKey(utils.MetaResources, utils.MetaRelease),
			utils.MetaBlockerError}, // expended to stop the processing because we have error at resource
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventWithCDRResourceErrorBlockError",
			Event: map[string]interface{}{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSSv1ItProcessEventWithCDRResourceErrorBlockError",
				utils.RequestType:  sSV1RequestType,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err == nil || err.Error() != "RESOURCES_ERROR:cannot find usage record with id: testSSv1ItProcessEventWithCDRResourceErrorBlockError" {
		t.Error(err)
	}
}

func testSSv1ItGetCDRsFromProcessEventResourceErrorBlockError(t *testing.T) {
	var cdrCnt int64
	req := &utils.RPCCDRsFilterWithOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		OriginIDs: []string{"testSSv1ItProcessEventWithCDRResourceErrorBlockError"}}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRsCount, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 0 {
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}

}
