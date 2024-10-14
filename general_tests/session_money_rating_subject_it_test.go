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
	sesMRSCfgPath string
	sesMRSCfgDIR  string
	sesMRSCfg     *config.CGRConfig
	sesMRSRPC     *birpc.Client
	sesMRSAccount = "refundAcc"
	sesMRSTenant  = "cgrates.org"

	sesMRSCgrEv = &utils.CGREvent{
		Tenant: sesMRSTenant,
		Event: map[string]any{
			utils.Tenant:       sesMRSTenant,
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestRefund",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: sesMRSAccount,
			utils.Destination:  "TEST",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		},
		APIOpts: map[string]any{
			utils.OptsSessionsTTL:   0,
			utils.OptsDebitInterval: time.Second,
		},
	}

	sTestSesMRSIt = []func(t *testing.T){
		testSesMRSItLoadConfig,
		testSesMRSItResetDataDB,
		testSesMRSItResetStorDb,
		testSesMRSItStartEngine,
		testSesMRSItRPCConn,
		testSesMRSItLoadFromFolder,
		testSesMRSItAddVoiceBalance,
		testSesMRSItInitSession,
		testSesMRSItTerminateSession,
		testSesMRSItStopCgrEngine,
	}
)

func TestSesMRSItMoney(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesMRSCfgDIR = "sessions_internal"
	case utils.MetaMySQL:
		sesMRSCfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		sesMRSCfgDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSesMRSIt {
		t.Run(sesMRSCfgDIR, stest)
	}
}

// test for 0 balance with sesMRSsion terminate with 1s usage
func testSesMRSItLoadConfig(t *testing.T) {
	var err error
	sesMRSCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesMRSCfgDIR)
	if sesMRSCfg, err = config.NewCGRConfigFromPath(sesMRSCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesMRSItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesMRSCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesMRSItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesMRSCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesMRSItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesMRSCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesMRSItRPCConn(t *testing.T) {
	sesMRSRPC = engine.NewRPCClient(t, sesMRSCfg.ListenCfg())
}

func testSesMRSItLoadFromFolder(t *testing.T) {
	//add a default charger
	chargerProfile := &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}
	var result string
	if err := sesMRSRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testSesMRSItAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      sesMRSTenant,
		Account:     sesMRSAccount,
		BalanceType: utils.MetaMonetary,
		Value:       float64(time.Hour - 5*time.Second),
		Balance: map[string]any{
			utils.ID:            "TestSesBal1",
			utils.RatingSubject: "*zero1s",
		},
	}
	var reply string
	if err := sesMRSRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	attrSetBalance = utils.AttrSetBalance{
		Tenant:      sesMRSTenant,
		Account:     sesMRSAccount,
		BalanceType: utils.MetaVoice,
		Value:       float64(5 * time.Second),
		Balance: map[string]any{
			utils.ID:            "TestSesBal2",
			utils.RatingSubject: "*zero1s",
			utils.Weight:        10,
		},
	}
	if err := sesMRSRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	var acnt engine.Account
	if err := sesMRSRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  sesMRSTenant,
			Account: sesMRSAccount,
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := float64(time.Hour - 5*time.Second)
	if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
	expected = float64(5 * time.Second)
	if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}

func testSesMRSItInitSession(t *testing.T) {
	sesMRSCgrEv.Event[utils.Usage] = time.Second
	var rply sessions.V1InitSessionReply
	if err := sesMRSRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent:    sesMRSCgrEv,
		}, &rply); err != nil {
		t.Error(err)
		return
	} else if *rply.MaxUsage != 3*time.Hour {
		t.Errorf("Unexpected MaxUsage: %v", rply.MaxUsage)
	}
	time.Sleep(50 * time.Millisecond)
}

func testSesMRSItTerminateSession(t *testing.T) {
	sesMRSCgrEv.Event[utils.Usage] = 10 * time.Second
	var rply string
	if err := sesMRSRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent:         sesMRSCgrEv,
		}, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sesMRSRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	var acnt engine.Account
	if err := sesMRSRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  sesMRSTenant,
			Account: sesMRSAccount,
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := float64(time.Hour - 10*time.Second)
	if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	expected = 0
	if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}

func testSesMRSItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
