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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	dspSessionCfgPath  string
	dspSessionCfg      *config.CGRConfig
	dspSessionRPC      *rpc.Client
	instSessionCfgPath string
	instSessionCfg     *config.CGRConfig
	instSessionRPC     *rpc.Client
)

var sTestsDspSession = []func(t *testing.T){
	testDspSessionInitCfg,
	testDspSessionInitDataDb,
	testDspSessionResetStorDb,
	testDspSessionStartEngine,
	testDspSessionRPCConn,
	testDspSessionPing,
	testDspSessionLoadData,
	testDspSessionAddAttributesWithPermision,
	testDspSessionTestAuthKey,
	testDspSessionAddAttributesWithPermision2,
	testDspSessionAuthorize,
	testDspSessionInit,
	testDspSessionUpdate,
	testDspSessionTerminate,
	testDspSessionProcessCDR,
	testDspSessionKillEngine,
}

//Test start here
func TestDspSessionS(t *testing.T) {
	for _, stest := range sTestsDspSession {
		t.Run("", stest)
	}
}

func testDspSessionInitCfg(t *testing.T) {
	var err error
	dspSessionCfgPath = path.Join(dspDataDir, "conf", "samples", "dispatcher")
	dspSessionCfg, err = config.NewCGRConfigFromFolder(dspSessionCfgPath)
	if err != nil {
		t.Error(err)
	}
	dspSessionCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(dspSessionCfg)
	instSessionCfgPath = path.Join(dspDataDir, "conf", "samples", "sessions")
	instSessionCfg, err = config.NewCGRConfigFromFolder(instSessionCfgPath)
	if err != nil {
		t.Error(err)
	}
	instSessionCfg.DataFolderPath = dspDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(instSessionCfg)
}

func testDspSessionInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(instSessionCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDspSessionResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(instSessionCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDspSessionStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(instSessionCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspSessionCfgPath, dspDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDspSessionRPCConn(t *testing.T) {
	var err error

	instSessionRPC, err = jsonrpc.Dial("tcp", instSessionCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
	dspSessionRPC, err = jsonrpc.Dial("tcp", dspSessionCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}

}

func testDspSessionPing(t *testing.T) {
	var reply string
	if err := instSessionRPC.Call(utils.SessionSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dspSessionRPC.Call(utils.SessionSv1Ping, "", &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspSessionLoadData(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(dspDataDir, "tariffplans", "testit")}
	if err := instSessionRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testDspSessionAddAttributesWithPermision(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AuthKey",
		Contexts:  []string{utils.MetaAuth},
		FilterIDs: []string{"*string:APIKey:12345"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.APIMethods,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("ThresholdSv1.GetThSessionholdsForEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instSessionRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instSessionRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspSessionTestAuthKey(t *testing.T) {
	authUsage := 5 * time.Minute
	args := AuthorizeArgsWithApiKey{
		APIKey: "12345",
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
	if err := dspSessionRPC.Call(utils.SessionSv1AuthorizeEventWithDigest,
		args, &rply); err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspSessionAddAttributesWithPermision2(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AuthKey",
		Contexts:  []string{utils.MetaAuth},
		FilterIDs: []string{"*string:APIKey:12345"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.APIMethods,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("SessionSv1.AuthorizeEventWithDigest&SessionSv1.InitiateSessionWithDigest&SessionSv1.UpdateSession&SessionSv1.TerminateSession&SessionSv1.ProcessCDR&SessionSv1.ProcessEvent", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	var result string
	if err := instSessionRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := instSessionRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "AuthKey"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testDspSessionAuthorize(t *testing.T) {
	authUsage := 5 * time.Minute
	argsAuth := &AuthorizeArgsWithApiKey{
		APIKey: "12345",
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
	if err := dspSessionRPC.Call(utils.SessionSv1AuthorizeEventWithDigest,
		argsAuth, &rply); err != nil {
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

func testDspSessionInit(t *testing.T) {
	initUsage := time.Duration(5 * time.Minute)
	argsInit := &InitArgsWithApiKey{
		APIKey: "12345",
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
	if err := dspSessionRPC.Call(utils.SessionSv1InitiateSessionWithDigest,
		argsInit, &rply); err != nil {
		t.Error(err)
	}
	if *rply.MaxUsage != initUsage.Seconds() {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	if *rply.ResourceAllocation != "RES_ACNT_1001" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
}

func testDspSessionUpdate(t *testing.T) {
	reqUsage := 5 * time.Minute
	argsUpdate := &UpdateSessionWithApiKey{
		APIKey: "12345",
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
	if err := dspSessionRPC.Call(utils.SessionSv1UpdateSession,
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
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
}

func testDspSessionTerminate(t *testing.T) {
	args := &TerminateSessionWithApiKey{
		APIKey: "12345",
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
	if err := dspSessionRPC.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testDspSessionProcessCDR(t *testing.T) {
	args := CGREvWithApiKey{
		APIKey: "12345",
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
	if err := dspSessionRPC.Call(utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testDspSessionProcessEvent(t *testing.T) {
	initUsage := 5 * time.Minute
	args := ProcessEventWithApiKey{
		APIKey: "12345",
		V1ProcessEventArgs: sessions.V1ProcessEventArgs{
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
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:       initUsage,
				},
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := dspSessionRPC.Call(utils.SessionSv1ProcessEvent,
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

func testDspSessionKillEngine(t *testing.T) {
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(dspDelay); err != nil {
		t.Error(err)
	}
}
