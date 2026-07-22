//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	sesMFDCfgDir  string
	sesMFDCfgPath string
	sesMFDCfg     *config.CGRConfig
	sesMFDRPC     *birpc.Client

	sesMFDTests = []func(t *testing.T){
		testSesMFDItLoadConfig,
		testSesMFDItResetDataDB,
		testSesMFDItResetStorDb,
		testSesMFDItStartEngine,
		testSesMFDItRPCConn,
		testSesMFDItSetChargers,
		testSesMFDItAddVoiceBalance,
		testSesMFDItProcessMessage,
		testSesMFDItGetAccountAfter,
		testSesMFDItStopCgrEngine,
	}
)

func TestSesMFDIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesMFDCfgDir = "sessions_internal"
	case utils.MetaMySQL:
		sesMFDCfgDir = "sessions_mysql"
	case utils.MetaMongo:
		sesMFDCfgDir = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sesMFDTests {
		t.Run(sesMFDCfgDir, stest)
	}
}

func testSesMFDItLoadConfig(t *testing.T) {
	var err error
	sesMFDCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesMFDCfgDir)
	if sesMFDCfg, err = config.NewCGRConfigFromPath(sesMFDCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesMFDItResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(sesMFDCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesMFDItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesMFDCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesMFDItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesMFDCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesMFDItRPCConn(t *testing.T) {
	sesMFDRPC = engine.NewRPCClient(t, sesMFDCfg.ListenCfg())
}

func testSesMFDItSetChargers(t *testing.T) {
	//add a default charger
	var result string
	if err := sesMFDRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := sesMFDRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "default2",
		RunID:        "default2",
		AttributeIDs: []string{"*none"},
		Weight:       10,
	}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testSesMFDItAddVoiceBalance(t *testing.T) {
	var reply string
	if err := sesMFDRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MetaSMS,
		Value:       1,
		Balance: map[string]any{
			utils.ID:            "TestSesBal1",
			utils.RatingSubject: "*zero1",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	var acnt engine.Account
	if err := sesMFDRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := 1.
	if rply := acnt.BalanceMap[utils.MetaSMS].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}
func testSesMFDItProcessMessage(t *testing.T) {
	var initRpl *sessions.V1ProcessMessageReply
	if err := sesMFDRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		&sessions.V1ProcessMessageArgs{
			Debit:         true,
			ForceDuration: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]any{
					utils.OriginID:     utils.UUIDSha1Prefix(),
					utils.ToR:          utils.MetaSMS,
					utils.Category:     utils.MetaSMS,
					utils.Tenant:       "cgrates.org",
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.RequestType:  utils.MetaPrepaid,
					utils.Usage:        1,
					utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				},
			},
		}, &initRpl); err == nil || err.Error() != utils.NewErrRALs(utils.ErrRatingPlanNotFound).Error() {
		t.Fatal(err)
	}

}

func testSesMFDItGetAccountAfter(t *testing.T) {
	var acnt engine.Account
	if err := sesMFDRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := 1.
	if rply := acnt.BalanceMap[utils.MetaSMS].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}

func testSesMFDItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
