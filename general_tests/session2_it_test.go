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

package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	ses2CfgDir  string
	ses2CfgPath string
	ses2Cfg     *config.CGRConfig
	ses2RPC     *birpc.Client

	ses2Tests = []func(t *testing.T){
		testSes2ItLoadConfig,
		testSes2ItResetDataDB,
		testSes2ItResetStorDb,
		testSes2ItStartEngine,
		testSes2ItRPCConn,
		testSes2ItLoadFromFolder,
		testSes2ItInitSession,
		testSes2ItAsActiveSessions,
		testSes2StirAuthenticate,
		testSes2StirInit,
		testSes2STIRAuthenticate,
		testSes2STIRIdentity,
		testSes2ItStopCgrEngine,
	}
)

func TestSes2It(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		ses2CfgDir = "tutinternal"
	case utils.MetaMySQL:
		ses2CfgDir = "tutmysql"
	case utils.MetaMongo:
		ses2CfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range ses2Tests {
		t.Run(ses2CfgDir, stest)
	}
}

func testSes2ItLoadConfig(t *testing.T) {
	var err error
	ses2CfgPath = path.Join(*utils.DataDir, "conf", "samples", ses2CfgDir)
	if ses2Cfg, err = config.NewCGRConfigFromPath(ses2CfgPath); err != nil {
		t.Error(err)
	}
}

func testSes2ItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(ses2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes2ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ses2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes2ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ses2CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSes2ItRPCConn(t *testing.T) {
	ses2RPC = engine.NewRPCClient(t, ses2Cfg.ListenCfg())
}

func testSes2ItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := ses2RPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSes2ItInitSession(t *testing.T) {
	// Set balance
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MetaVoice,
		Value:       float64(time.Hour),
		Balance: map[string]any{
			utils.ID: "TestDynamicDebitBalance",
		},
	}
	var reply string
	if err := ses2RPC.Call(context.Background(), utils.APIerSv2SetBalance,
		attrSetBalance, &reply); err != nil {
		t.Fatal(err)
	}

	// Init session
	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.OriginID:     utils.UUIDSha1Prefix(),
				utils.ToR:          utils.MetaVoice,
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			},
		},
	}
	var initRpl *sessions.V1InitSessionReply
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Fatal(err)
	}

}

func testSes2ItAsActiveSessions(t *testing.T) {
	var count int
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1GetActiveSessionsCount, utils.SessionFilter{
		Filters: []string{"*string:~*req.Account:1001"},
	}, &count); err != nil {
		t.Fatal(err)
	} else if count != 2 { // 2 chargers
		t.Errorf("Expected 2 session received %v session(s)", count)
	}
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1GetActiveSessionsCount, utils.SessionFilter{
		Filters: []string{"*string:~*req.Account:1002"},
	}, &count); err != nil {
		t.Fatal(err)
	} else if count != 0 {
		t.Errorf("Expected 0 session received %v session(s)", count)
	}
}

func testSes2ItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testSes2StirAuthenticate(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaSTIRAuthenticate},

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSes2StirAuthorize",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSes2StirAuthorize",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.Usage:        10 * time.Minute,
			},
			APIOpts: map[string]any{
				utils.OptsStirIdentity: "eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiL3Vzci9zaGFyZS9jZ3JhdGVzL3N0aXIvc3Rpcl9wdWJrZXkucGVtIn0.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMzg4MDIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.cMEMlFnfyTu8uxfeU4RoZTamA7ifFT9Ibwrvi1_LKwL2xAU6fZ_CSIxKbtyOpNhM_sV03x7CfA_v0T4sHkifzg;info=</usr/share/cgrates/stir/stir_pubkey.pem>;ppt=shaken",
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1ProcessEvent,
		args, &rply); err != nil { // no error verificated with success
		t.Error(err)
	}
	// altered originator
	args.APIOpts[utils.OptsStirOriginatorTn] = "1005"
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1ProcessEvent,
		args, &rply); err == nil || err.Error() != "*stir_authenticate: wrong originatorTn" {
		t.Errorf("Expected error :%q ,receved: %v", "*stir_authenticate: wrong originatorTn", err)
	}

	// altered identity
	args.APIOpts[utils.OptsStirIdentity] = "eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiL3Vzci9zaGFyZS9jZ3JhdGVzL3N0aXIvc3Rpcl9wdWJrZXkucGVtIn0.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMzg4MDIsIm9yaWciOnsidG4iOiIxMDA1In0sIm9yaWdpZCI6IjEyMzQ1NiJ9.cMEMlFnfyTu8uxfeU4RoZTamA7ifFT9Ibwrvi1_LKwL2xAU6fZ_CSIxKbtyOpNhM_sV03x7CfA_v0T4sHkifzg;info=</usr/share/cgrates/stir/stir_pubkey.pem>;ppt=shaken"
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1ProcessEvent,
		args, &rply); err == nil || err.Error() != "*stir_authenticate: crypto/ecdsa: verification error" {
		t.Errorf("Expected error :%q ,receved: %v", "*stir_authenticate: crypto/ecdsa: verification error", err)
	}
}

func testSes2StirInit(t *testing.T) {
	args := &sessions.V1ProcessEventArgs{
		Flags: []string{utils.MetaSTIRInitiate},

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSes2StirInit",
			Event: map[string]any{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSes2StirInit",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.Usage:        10 * time.Minute,
			},
			APIOpts: map[string]any{
				utils.OptsStirPublicKeyPath:  "/usr/share/cgrates/stir/stir_pubkey.pem",
				utils.OptsStirPrivateKeyPath: "/usr/share/cgrates/stir/stir_privatekey.pem",
			},
		},
	}
	var rply sessions.V1ProcessEventReply
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1ProcessEvent,
		args, &rply); err != nil { // no error verificated with success
		t.Error(err)
	}
	if err := sessions.AuthStirShaken(rply.STIRIdentity[utils.MetaRaw], "1001", "", "1002", "", utils.NewStringSet([]string{"A"}), 10*time.Minute); err != nil {
		t.Fatal(err)
	}
}

func testSes2STIRAuthenticate(t *testing.T) {
	var rply string
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1STIRAuthenticate,
		&sessions.V1STIRAuthenticateArgs{
			Attest:             []string{"A"},
			PayloadMaxDuration: "-1",
			DestinationTn:      "1002",
			Identity:           "eyJhbGciOiJFUzI1NiIsInBwdCI6InNoYWtlbiIsInR5cCI6InBhc3Nwb3J0IiwieDV1IjoiL3Vzci9zaGFyZS9jZ3JhdGVzL3N0aXIvc3Rpcl9wdWJrZXkucGVtIn0.eyJhdHRlc3QiOiJBIiwiZGVzdCI6eyJ0biI6WyIxMDAyIl19LCJpYXQiOjE1ODcwMzg4MDIsIm9yaWciOnsidG4iOiIxMDAxIn0sIm9yaWdpZCI6IjEyMzQ1NiJ9.cMEMlFnfyTu8uxfeU4RoZTamA7ifFT9Ibwrvi1_LKwL2xAU6fZ_CSIxKbtyOpNhM_sV03x7CfA_v0T4sHkifzg;info=</usr/share/cgrates/stir/stir_pubkey.pem>;ppt=shaken",
			OriginatorTn:       "1001",
		}, &rply); err != nil {
		t.Fatal(err)
	} else if rply != utils.OK {
		t.Errorf("Expected: %s ,received: %s", utils.OK, rply)
	}
}

func testSes2STIRIdentity(t *testing.T) {
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
	}
	var rply string
	if err := ses2RPC.Call(context.Background(), utils.SessionSv1STIRIdentity,
		args, &rply); err != nil {
		t.Error(err)
	}
	if err := sessions.AuthStirShaken(rply, "1001", "", "1002", "", utils.NewStringSet([]string{"A"}), 10*time.Minute); err != nil {
		t.Fatal(err)
	}
}
