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
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tutCfgPath string
	tutCfg     *config.CGRConfig
	tutRpc     *birpc.Client
	tutCfgDir  string //run tests for specific configuration
	tutDelay   int
)

var sTutTests = []func(t *testing.T){
	testTutLoadConfig,
	testTutResetDB,
	testTutStartEngine,
	testTutRpcConn,
	testTutFromFolder,
	testTutGetCost,
	testTutAccounts,
	testTutStopEngine,
}

// Test start here
func TestTutorial2(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tutCfgDir = "tutinternal"
	case utils.MetaMySQL:
		tutCfgDir = "tutmysql2"
	case utils.MetaMongo:
		tutCfgDir = "tutmongo2"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *utils.Encoding == utils.MetaGOB {
		tutCfgDir += "_gob"
	}

	for _, stest := range sTutTests {
		t.Run(tutCfgDir, stest)
	}
}

func testTutLoadConfig(t *testing.T) {
	var err error
	tutCfgPath = path.Join(*utils.DataDir, "conf", "samples", tutCfgDir)
	if tutCfg, err = config.NewCGRConfigFromPath(tutCfgPath); err != nil {
		t.Error(err)
	}
	tutDelay = 2000
}

func testTutResetDB(t *testing.T) {
	if err := engine.InitDataDb(tutCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(tutCfg); err != nil {
		t.Fatal(err)
	}
}

func testTutStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutCfgPath, tutDelay); err != nil {
		t.Fatal(err)
	}
}

func testTutStopEngine(t *testing.T) {
	if err := engine.KillEngine(tutDelay); err != nil {
		t.Error(err)
	}
}

func testTutRpcConn(t *testing.T) {
	tutRpc = engine.NewRPCClient(t, tutCfg.ListenCfg())
}

func testTutFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial2")}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder,
		attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testTutGetCost(t *testing.T) {
	// Standard pricing for 1001->1002
	attrs := v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "45s",
	}
	var rply *engine.EventCost
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.550000 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Fallback pricing from *any, Usage will be rounded to 60s
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "2019-03-11T09:00:00Z",
		Usage:       "45s",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 1.4 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Fallback pricing from *any, Usage will be rounded to 60s
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "2019-03-11T21:00:00Z",
		Usage:       "45s",
	}
	// *any to 2001
	attrs = v1.AttrGetCost{
		Subject:     "1002",
		Destination: "2001",
		AnswerTime:  "*now",
		Usage:       "45s",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 1.4 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// *any to 2001 on NEW_YEAR
	attrs = v1.AttrGetCost{
		Subject:     "1002",
		Destination: "2001",
		AnswerTime:  "2020-01-01T21:00:00Z",
		Usage:       "45s",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.55 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Fallback pricing from *any, Usage will be rounded to 60s
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "2019-03-11T21:00:00Z",
		Usage:       "45s",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.55 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Unauthorized destination
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "4003",
		AnswerTime:  "2019-03-11T09:00:00Z",
		Usage:       "1m",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: UNAUTHORIZED_DESTINATION" {
		t.Error("Unexpected nil error received: ", err)
	}
	// Data charging
	attrs = v1.AttrGetCost{
		Category:   "data",
		Subject:    "1001",
		AnswerTime: "*now",
		Usage:      "2048",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 2.0 { // FixMe: missing ConnectFee out of Cost
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// SMS charging 1002
	attrs = v1.AttrGetCost{
		Category:    "sms",
		Subject:     "1003",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "1",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.1 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// SMS charging 10
	attrs = v1.AttrGetCost{
		Category:    "sms",
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "*now",
		Usage:       "1",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.2 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// SMS charging UNAUTHORIZED
	attrs = v1.AttrGetCost{
		Category:    "sms",
		Subject:     "1001",
		Destination: "2001",
		AnswerTime:  "*now",
		Usage:       "1",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: UNAUTHORIZED_DESTINATION" {
		t.Error("Unexpected nil error received: ", err)
	}
	// Per call charges
	attrs = v1.AttrGetCost{
		Category:    "call",
		Subject:     "RPF_SPECIAL_BLC",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "5m",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.1 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// reseller1 pricing for 1001->1002
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "45s",
		Category:    "reseller1",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.1 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// reseller1 pricing for 1001->1002 duration independent
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "10m45s",
		Category:    "reseller1",
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.1 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

func testTutAccounts(t *testing.T) {
	// make sure Account was created
	var acnt *engine.Account
	if err := tutRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"},
		&acnt); err != nil {
		t.Fatal(err)
	}
	if len(acnt.BalanceMap) != 4 ||
		len(acnt.BalanceMap[utils.MetaMonetary]) != 1 ||
		acnt.BalanceMap[utils.MetaMonetary][0].Value != 10 ||
		len(acnt.BalanceMap[utils.MetaVoice]) != 2 ||
		len(acnt.BalanceMap[utils.MetaSMS]) != 1 ||
		acnt.BalanceMap[utils.MetaSMS][0].Value != 100 ||
		len(acnt.BalanceMap[utils.MetaData]) != 1 ||
		acnt.BalanceMap[utils.MetaData][0].Value != 1024 ||
		len(acnt.ActionTriggers) != 2 ||
		acnt.Disabled {
		t.Errorf("received account: %s", utils.ToIJSON(acnt))
	}

	// test ActionTriggers
	attrBlc := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MetaMonetary,
		Value:       1,
		Balance: map[string]any{
			utils.ID: utils.MetaDefault,
		},
	}
	var rplySetBlc string
	if err := tutRpc.Call(context.Background(), utils.APIerSv1SetBalance, attrBlc, &rplySetBlc); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"},
		&acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaSMS]) != 2 ||
		acnt.GetBalanceWithID(utils.MetaSMS, "BONUS_SMSes").Value != 10 {
		t.Errorf("account: %s", utils.ToIJSON(acnt))
	}
	attrBlc = utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MetaMonetary,
		Value:       101,
		Balance: map[string]any{
			utils.ID: utils.MetaDefault,
		},
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv1SetBalance, attrBlc, &rplySetBlc); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	}
	if err := tutRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"},
		&acnt); err != nil {
		t.Error(err)
	} else if !acnt.Disabled {
		t.Errorf("account: %s", utils.ToIJSON(acnt))
	}
	// enable the account again
	var rplySetAcnt string
	if err := tutRpc.Call(context.Background(), utils.APIerSv2SetAccount,
		&v2.AttrSetAccount{
			Tenant:  "cgrates.org",
			Account: "1001",
			ExtraOptions: map[string]bool{
				utils.Disabled: false,
			},
		}, &rplySetAcnt); err != nil {
		t.Error(err)
	}
	acnt = new(engine.Account)
	if err := tutRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"},
		&acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled {
		t.Errorf("account: %s", utils.ToJSON(acnt))
	}
}
