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
	sesExpCfgPath string
	sesExpCfgDIR  string
	sesExpCfg     *config.CGRConfig
	sesExpRPC     *birpc.Client
	sesExpAccount = "refundAcc"
	sesExpTenant  = "cgrates.org"

	sesExpCgrEv = &utils.CGREvent{
		Tenant: sesExpTenant,
		Event: map[string]any{
			utils.Tenant:       sesExpTenant,
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestRefund",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: sesExpAccount,
			utils.Subject:      "*zero1s",
			utils.Destination:  "TEST",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		},
		APIOpts: map[string]any{
			utils.OptsSessionsTTL:   0,
			utils.OptsDebitInterval: time.Second,
		},
	}

	sTestSesExpIt = []func(t *testing.T){
		testSesExpItLoadConfig,
		testSesExpItResetDataDB,
		testSesExpItResetStorDb,
		testSesExpItStartEngine,
		testSesExpItRPCConn,
		testSesExpItLoadFromFolder,
		testSesExpItAddVoiceBalance,
		testSesExpItInitSession,
		testSesExpItTerminateSession,
		testSesExpItProcessCDR,
		testSesExpItRerate,
		testSesExpItStopCgrEngine,
	}
)

func TestSesExpIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesExpCfgDIR = "sessions_internal"
	case utils.MetaMySQL:
		sesExpCfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		sesExpCfgDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSesExpIt {
		t.Run(sesExpCfgDIR, stest)
	}
}

// test for 0 balance with sesExpsion terminate with 1s usage
func testSesExpItLoadConfig(t *testing.T) {
	var err error
	sesExpCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesExpCfgDIR)
	if sesExpCfg, err = config.NewCGRConfigFromPath(sesExpCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesExpItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesExpCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesExpItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesExpCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesExpItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesExpCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesExpItRPCConn(t *testing.T) {
	sesExpRPC = engine.NewRPCClient(t, sesExpCfg.ListenCfg())
}

func testSesExpItLoadFromFolder(t *testing.T) {
	//add a default charger
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result string
	if err := sesExpRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testSesExpItAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      sesExpTenant,
		Account:     sesExpAccount,
		BalanceType: utils.MetaVoice,
		Value:       float64(time.Hour),
		Balance: map[string]any{
			utils.ID:            "TestSesBal1",
			utils.RatingSubject: "*zero1s",
		},
	}
	var reply string
	if err := sesExpRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	attrSetBalance = utils.AttrSetBalance{
		Tenant:      sesExpTenant,
		Account:     sesExpAccount,
		BalanceType: utils.MetaVoice,
		Value:       float64(time.Second),
		Balance: map[string]any{
			utils.ID:            "TestSesBalExpire",
			utils.RatingSubject: "*zero1s",
			utils.ExpiryTime:    time.Now().Add(50 * time.Millisecond),
			utils.Weight:        100,
		},
	}
	if err := sesExpRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	var acnt engine.Account
	if err := sesExpRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  sesExpTenant,
			Account: sesExpAccount,
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := float64(time.Hour + time.Second)
	if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}

func testSesExpItInitSession(t *testing.T) {
	sesExpCgrEv.Event[utils.Usage] = time.Second
	var rply sessions.V1InitSessionReply
	if err := sesExpRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent:    sesExpCgrEv,
		}, &rply); err != nil {
		t.Error(err)
		return
	} else if *rply.MaxUsage != 3*time.Hour {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	time.Sleep(50 * time.Millisecond)
}

func testSesExpItTerminateSession(t *testing.T) {
	sesExpCgrEv.Event[utils.Usage] = 10 * time.Second
	var rply string
	if err := sesExpRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent:         sesExpCgrEv,
		}, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sesExpRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	var acnt engine.Account
	if err := sesExpRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  sesExpTenant,
			Account: sesExpAccount,
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := float64(time.Hour - 9*time.Second)
	if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}

func testSesExpItProcessCDR(t *testing.T) {
	var reply string
	if err := sesExpRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		sesExpCgrEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(20 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Accounts: []string{sesExpAccount}}
	if err := sesExpRPC.Call(context.Background(), utils.APIerSv2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Fatal("Wrong number of CDRs")
	} else if cd, err := engine.IfaceAsEventCost(cdrs[0].CostDetails); err != nil {
		t.Fatal(err)
	} else if len(cd.AccountSummary.BalanceSummaries) != 2 {
		t.Fatal("Expected two BalanceSummaries but received:", len(cd.AccountSummary.BalanceSummaries))
	}
}
func testSesExpItRerate(t *testing.T) {
	var reply string
	sesExpCgrEv.Event[utils.Usage] = time.Second
	sesExpCgrEv.Event[utils.RequestType] = utils.MetaPostpaid // change the request type in order to not wait 12s to check the cost for a closed session
	if err := sesExpRPC.Call(context.Background(), utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags:    []string{utils.MetaRerate},
			CGREvent: *sesExpCgrEv,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var acnt engine.Account
	if err := sesExpRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  sesExpTenant,
			Account: sesExpAccount,
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := float64(time.Hour - time.Second)
	if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}

func testSesExpItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
