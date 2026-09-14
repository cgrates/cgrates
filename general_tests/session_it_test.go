//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

// import (
// 	"net/rpc"
// 	"path"
// 	"testing"
// 	"time"

// 	"github.com/cgrates/cgrates/config"
// 	"github.com/cgrates/cgrates/engine"
// 	"github.com/cgrates/cgrates/sessions"
// 	"github.com/cgrates/cgrates/utils"
// )

// var (
// 	sesCfgPath string
// 	sesCfgDIR  string
// 	sesCfg     *config.CGRConfig
// 	sesRPC     *rpc.Client
// 	sesAccount = "refundAcc"
// 	sesTenant  = "cgrates.org"

// 	sTestSesIt = []func(t *testing.T){
// 		testSesItLoadConfig,
// 		testSesItResetDataDB,
// 		testSesItStartEngine,
// 		testSesItRPCConn,
// 		testSesItLoadFromFolder,
// 		testSesItAddVoiceBalance,
// 		testSesItInitSession,
// 		testSesItTerminateSession,
// 		testSesItStopCgrEngine,
// 	}
// )

// func TestSesIt(t *testing.T) {
// 	switch *utils.DBType {
// 	case utils.MetaInternal:
// 		sesCfgDIR = "tutinternal"
// 	case utils.MetaRedis:
//     t.SkipNow()
// case utils.MetaMySQL:
// 		sesCfgDIR = "tutmysql_internal"
// 	case utils.MetaMongo:
// 		sesCfgDIR = "tutmongo"
// 	case utils.MetaPostgres:
// 		t.SkipNow()
// 	default:
// 		t.Fatal("Unknown Database type")
// 	}
// 	for _, stest := range sTestSesIt {
// 		t.Run(sesCfgDIR, stest)
// 	}
// }

// // test for 0 balance with session terminate with 1s usage
// func testSesItLoadConfig(t *testing.T) {
// 	sesCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesCfgDIR)
// 	if sesCfg, err = config.NewCGRConfigFromPath(sesCfgPath); err != nil {
// 		t.Error(err)
// 	}
// }

// func testSesItResetDataDB(t *testing.T) {
// 	if err := engine.InitDataDB(sesCfg); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testSesItStartEngine(t *testing.T) {
// 	if _, err := engine.StopStartEngine(sesCfgPath, *utils.WaitRater); err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testSesItRPCConn(t *testing.T) {
// 	var err error
// 	sesRPC, err = engine.NewRPCClient(sesCfg.ListenCfg(), *utils.Encoding)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func testSesItLoadFromFolder(t *testing.T) {
// 	var reply string
// 	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
// 	if err := sesRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
// 		t.Error(err)
// 	}
// 	time.Sleep(100 * time.Millisecond)
// }

// func testAccountBalance2(t *testing.T, sracc, srten, balType string, expected float64) {
// 	var acnt engine.Account
// 	attrs := &utils.AttrGetAccount{
// 		Tenant:  srten,
// 		Account: sracc,
// 	}
// 	if err := sesRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
// 		t.Error(err)
// 	} else if rply := acnt.BalanceMap[balType].GetTotalValue(); rply != expected {
// 		t.Errorf("Expecting: %v, received: %v",
// 			expected, rply)
// 	}
// }

// func testSesItAddVoiceBalance(t *testing.T) {
// 	attrSetBalance := utils.AttrSetBalance{
// 		Tenant:      sesTenant,
// 		Account:     sesAccount,
// 		BalanceType: utils.MetaMonetary,
// 		Value:       0,
// 		Balance: map[string]any{
// 			utils.ID:            "TestDynamicDebitBalance",
// 			utils.RatingSubject: "*zero1s",
// 		},
// 	}
// 	var reply string
// 	if err := sesRPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != utils.OK {
// 		t.Errorf("Received: %s", reply)
// 	}
// 	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance2(t, sesAccount, sesTenant, utils.MetaMonetary, 0) })
// }

// func testSesItStopCgrEngine(t *testing.T) {
// 	if err := engine.KillEngine(100); err != nil {
// 		t.Error(err)
// 	}
// }
