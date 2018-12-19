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

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sSv1CfgPath      string
	sSv1Cfg          *config.CGRConfig
	sSv1BiRpc        *rpc2.Client
	sSApierRpc       *rpc.Client
	disconnectEvChan = make(chan *utils.AttrDisconnectSession)
)

func handleDisconnectSession(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	disconnectEvChan <- args
	*reply = utils.OK
	return nil
}

func handleGetSessionIDs(clnt *rpc2.Client,
	ignParam string, sessionIDs *[]*sessions.SessionID) error {
	return nil
}

func TestSSv1ItInitCfg(t *testing.T) {
	var err error
	sSv1CfgPath = path.Join(*dataDir, "conf", "samples", "sessions")
	// Init config first
	sSv1Cfg, err = config.NewCGRConfigFromFolder(sSv1CfgPath)
	if err != nil {
		t.Error(err)
	}
	sSv1Cfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(sSv1Cfg)
}

func TestSSv1ItResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(sSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestSSv1ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func TestSSv1ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sSv1CfgPath, 100); err != nil {
		t.Fatal(err)
	}
}

func TestSSv1ItRpcConn(t *testing.T) {
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
	if sSApierRpc, err = jsonrpc.Dial("tcp", sSv1Cfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	dummyClnt.Close() // close so we don't get EOF error when disconnecting server
}

func TestSSv1ItPing(t *testing.T) {
	var resp string
	if err := sSv1BiRpc.Call(utils.SessionSv1Ping, "", &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSSv1ItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	var loadInst utils.LoadInstance
	if err := sSApierRpc.Call(utils.ApierV2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func TestSSv1ItAuth(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
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
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := sSv1BiRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Error(err)
	}
	if *rply.MaxUsage != authUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation == "" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eSplrs := &engine.SortedSuppliers{
		ProfileID: "SPL_ACNT_1001",
		Sorting:   utils.MetaWeight,
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
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eSplrs), utils.ToJSON(rply.Suppliers))
	}
	eAttrs := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "TestSSv1ItAuth",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
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

func TestSSv1ItAuthWithDigest(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
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
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
	}
	var rply sessions.V1AuthorizeReplyWithDigest
	if err := sSv1BiRpc.Call(utils.SessionSv1AuthorizeEventWithDigest, args, &rply); err != nil {
		t.Error(err)
	}
	if *rply.MaxUsage != authUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation == "" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	eSplrs := utils.StringPointer("supplier1,supplier2")
	if *eSplrs != *rply.SuppliersDigest {
		t.Errorf("expecting: %v, received: %v", *eSplrs, *rply.SuppliersDigest)
	}
	eAttrs := utils.StringPointer("OfficeGroup:Marketing")
	if *eAttrs != *rply.AttributesDigest {
		t.Errorf("expecting: %v, received: %v", *eAttrs, *rply.AttributesDigest)
	}
}

func TestSSv1ItInitiateSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
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
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
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
			ID:      "TestSSv1ItInitiateSession",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "TestSSv1It1",
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
	aSessions := make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
}

func TestSSv1ItInitiateSessionWithDigest(t *testing.T) {
	initUsage := time.Duration(5 * time.Minute)
	args := &sessions.V1InitSessionArgs{
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
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
	}
	var rply sessions.V1InitReplyWithDigest
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSessionWithDigest,
		args, &rply); err != nil {
		t.Error(err)
	}
	if *rply.MaxUsage != initUsage.Seconds() {
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
	aSessions := make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 4 { // the digest has increased the number of sessions
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
}

func TestSSv1ItUpdateSession(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
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
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
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
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "TestSSv1ItUpdateSession",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.CGRID:       "5668666d6b8e44eb949042f25ce0796ec3592ff9",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				"OfficeGroup":     "Marketing",
				utils.OriginID:    "TestSSv1It1",
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
	if *rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	aSessions := make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 4 { // the digest has increased the number of sessions
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
}

func TestSSv1ItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
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
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
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
	aSessions := make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestSSv1ItProcessCDR(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestSSv1ItProcessCDR",
		Event: map[string]interface{}{
			utils.Tenant:      "cgrates.org",
			utils.Category:    "call",
			utils.ToR:         utils.VOICE,
			utils.OriginID:    "TestSSv1It1",
			utils.RequestType: utils.META_PREPAID,
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

// TestSSv1ItProcessEvent processes individual event and also checks it's CDRs
func TestSSv1ItProcessEvent(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1ProcessEventArgs{
		AllocateResources: true,
		Debit:             true,
		GetAttributes:     true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItProcessEvent",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It2",
				utils.RequestType: utils.META_PREPAID,
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
				utils.CGRID:       "c87609aa1cb6e9529ab1836cfeeeb0ab7aa7ebaf",
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
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
	aSessions := make([]*sessions.ActiveSession, 0)
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
	time.Sleep(100 * time.Millisecond)
}

func TestSSv1ItCDRsGetCdrs(t *testing.T) {
	var cdrCnt int64
	req := utils.AttrGetCdrs{}
	if err := sSApierRpc.Call(utils.CdrsV1CountCDRs, req, &cdrCnt); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if cdrCnt != 6 { // 3 for each CDR
		t.Error("Unexpected number of CDRs returned: ", cdrCnt)
	}

	var cdrs []*engine.CDR
	args := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaRaw}}
	if err := sSApierRpc.Call(utils.CdrsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != -1.0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"},
		OriginIDs: []string{"TestSSv1It1"}}
	if err := sSApierRpc.Call(utils.CdrsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.198 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
		OriginIDs: []string{"TestSSv1It1"}}
	if err := sSApierRpc.Call(utils.CdrsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.102 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"},
		OriginIDs: []string{"TestSSv1It2"}}
	if err := sSApierRpc.Call(utils.CdrsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.099 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	args = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
		OriginIDs: []string{"TestSSv1It2"}}
	if err := sSApierRpc.Call(utils.CdrsV1GetCDRs, args, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.051 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func TestSSv1ItForceUpdateSession(t *testing.T) {
	aSessions := make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(asessions)=%v", err, len(aSessions))
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.25
	if err := sSApierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
		GetAttributes: true,
		UpdateSession: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
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
		AlteredFields:   []string{"OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "TestSSv1ItUpdateSession",
			Context: utils.StringPointer(utils.MetaSessionS),
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
	if *rply.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	aSessions = make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("wrong active ssesions: %s", utils.ToJSON(aSessions))
	}

	eAcntVal = 9.10
	if err := sSApierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
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
	aSessions = make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := sSApierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal { // no monetary change bacause the sessin was terminated
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	time.Sleep(100 * time.Millisecond)
	var cdrs []*engine.CDR
	argsCDR := utils.RPCCDRsFilter{RunIDs: []string{"CustomerCharges"}, OriginIDs: []string{"TestSSv1It"}}
	if err := sSApierRpc.Call(utils.CdrsV1GetCDRs, argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs), "\n", utils.ToJSON(cdrs))
	} else {
		if cdrs[0].Cost != 0.099 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	argsCDR = utils.RPCCDRsFilter{RunIDs: []string{"SupplierCharges"},
		OriginIDs: []string{"TestSSv1It"}}
	if err := sSApierRpc.Call(utils.CdrsV1GetCDRs, argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.051 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
}

func TestSSv1ItDynamicDebit(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "TestDynamicDebit",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestDynamicDebitBalance"),
		Value:         utils.Float64Pointer(2 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero5ms")}
	var reply string
	if err := sSApierRpc.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
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
	if err := sSApierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}

	args1 := &sessions.V1InitSessionArgs{
		InitSession:   true,
		GetAttributes: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:           "cgrates.org",
				utils.Category:         "call",
				utils.ToR:              utils.VOICE,
				utils.OriginID:         "TestDynamicTDebit",
				utils.RequestType:      utils.META_PREPAID,
				utils.Account:          "TestDynamicDebit",
				utils.Subject:          "TEST",
				utils.Destination:      "TEST",
				utils.SetupTime:        time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:       time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:            0,
				utils.CGRDebitInterval: 30 * time.Millisecond,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := sSv1BiRpc.Call(utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if *rply1.MaxUsage != -1 {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}

	aSessions := make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
	time.Sleep(time.Millisecond)
	eAcntVal -= float64(time.Millisecond) * 30 * 2 // 2 session
	if err := sSApierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}

	time.Sleep(10 * time.Millisecond)
	if err := sSApierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}
	time.Sleep(20 * time.Millisecond)
	eAcntVal -= float64(time.Millisecond) * 30 * 2 // 2 session
	if err := sSApierRpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %v, received: %v",
			time.Duration(eAcntVal), time.Duration(acnt.BalanceMap[utils.VOICE].GetTotalValue()))
	}

	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("wrong active sessions: %s", utils.ToJSON(aSessions))
	}
	rplyt := ""
	if err := sSv1BiRpc.Call(utils.SessionSv1ForceDisconnect,
		nil, &rplyt); err != nil {
		t.Error(err)
	} else if rplyt != utils.OK {
		t.Errorf("Unexpected reply: %s", rplyt)
	}
	aSessions = make([]*sessions.ActiveSession, 0)
	if err := sSv1BiRpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestSSv1ItStopCgrEngine(t *testing.T) {
	if err := sSv1BiRpc.Close(); err != nil { // Close the connection so we don't get EOF warnings from client
		t.Error(err)
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
