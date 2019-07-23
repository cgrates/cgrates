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
	"reflect"
	"testing"
	"time"

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
	testSSv1ItStopCgrEngine,
}

func TestSSv1ItProcessEventWithPrepaid(t *testing.T) {
	sSV1RequestType = utils.META_PREPAID
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sSV1RequestType, stest)
	}
}

func TestSSv1ItProcessEventWithPostPaid(t *testing.T) {
	sSV1RequestType = utils.META_POSTPAID
	sTestSessionSv1ProcessEvent = append(sTestSessionSv1ProcessEvent[:len(sTestSessionSv1ProcessEvent)-3], testSSv1ItStopCgrEngine)
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sSV1RequestType, stest)
	}
}

func TestSSv1ItProcessEventWithRated(t *testing.T) {
	sSV1RequestType = utils.META_RATED
	sTestSessionSv1ProcessEvent = append(sTestSessionSv1ProcessEvent[:len(sTestSessionSv1ProcessEvent)-3], testSSv1ItStopCgrEngine)
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sSV1RequestType, stest)
	}
}

func TestSSv1ItProcessEventWithPseudoPrepaid(t *testing.T) {
	sSV1RequestType = utils.META_PSEUDOPREPAID
	for _, stest := range sTestSessionSv1ProcessEvent {
		t.Run(sSV1RequestType, stest)
	}
}

func testSSv1ItProcessEventAuth(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{"*resources:*authorize", "*rals:*auth", "*suppliers", "*attributes"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventAuth",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testSSv1ItProcessEvent",
				utils.RequestType: sSV1RequestType,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	if *rply.MaxUsage != authUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceMessage == "" {
		t.Errorf("Unexpected ResourceMessage: %s", *rply.ResourceMessage)
	}
	eSplrs := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1001",
		Sorting:   utils.MetaWeight,
		Count:     2,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
			},
		},
	}
	if !reflect.DeepEqual(eSplrs, rply.Suppliers) {
		t.Errorf("expecting: %+v,\n received: %+v", utils.ToJSON(eSplrs), utils.ToJSON(rply.Suppliers))
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventAuth",
			Event: map[string]interface{}{
				utils.CGRID:       "4be779c004d9f784e836db9ffd41b50319d71fe8",
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "testSSv1ItProcessEvent",
				utils.RequestType: sSV1RequestType,
				utils.SetupTime:   "2018-01-07T17:00:00Z",
				utils.Usage:       300000000000.0,
			},
		},
	}
	if !reflect.DeepEqual(eAttrs, rply.Attributes) {
		t.Errorf("expecting: %+v, received: %+v",
			utils.ToJSON(eAttrs), utils.ToJSON(rply.Attributes))
	}
}

func testSSv1ItProcessEventInitiateSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{"*rals:*init", "*resources:*allocate", "*attributes"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testSSv1ItProcessEvent",
				utils.RequestType: sSV1RequestType,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := sSv1BiRpc.Call(utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect -1
	if ((sSV1RequestType == utils.META_PREPAID ||
		sSV1RequestType == utils.META_PSEUDOPREPAID) && *rply.MaxUsage != initUsage) ||
		((sSV1RequestType == utils.META_POSTPAID ||
			sSV1RequestType == utils.META_RATED) && *rply.MaxUsage != -1) {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceMessage != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceMessage: %s", *rply.ResourceMessage)
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "testSSv1ItProcessEvent",
				utils.RequestType: sSV1RequestType,
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
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	}
}

func testSSv1ItProcessEventUpdateSession(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{"*rals:*update", "*attributes"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testSSv1ItProcessEvent",
				utils.RequestType: sSV1RequestType,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
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
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventUpdateSession",
			Event: map[string]interface{}{
				utils.CGRID:       "4be779c004d9f784e836db9ffd41b50319d71fe8",
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "testSSv1ItProcessEvent",
				utils.RequestType: sSV1RequestType,
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
	// in case of prepaid and pseudoprepade we expect a MaxUsage of 5min
	// and in case of postpaid and rated we expect -1
	if ((sSV1RequestType == utils.META_PREPAID ||
		sSV1RequestType == utils.META_PSEUDOPREPAID) && *rply.MaxUsage != reqUsage) ||
		((sSV1RequestType == utils.META_POSTPAID ||
			sSV1RequestType == utils.META_RATED) && *rply.MaxUsage != -1) {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
}

func testSSv1ItProcessEventTerminateSession(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{"*rals:*terminate", "*resources:*release"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSSv1ItProcessEventTerminateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testSSv1ItProcessEvent",
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
			utils.Tenant:      "cgrates.org",
			utils.ToR:         utils.VOICE,
			utils.OriginID:    "testSSv1ItProcessEvent",
			utils.RequestType: sSV1RequestType,
			utils.Account:     "1001",
			utils.Subject:     "ANY2CNT",
			utils.Destination: "1002",
			utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:       10 * time.Minute,
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
	time.Sleep(100 * time.Millisecond)
}

func testSSv1ItGetCDRs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := sSApierRpc.Call(utils.CDRsV1CountCDRs, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 3 { // 3 for each CDR
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}

	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"},
		OriginIDs: []string{"testSSv1ItProcessEvent"}}
	if err := sSApierRpc.Call(utils.CDRsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
		OriginIDs: []string{"testSSv1ItProcessEvent"}}
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
