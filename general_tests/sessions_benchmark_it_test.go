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
	"fmt"
	"net/rpc"
	"path"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesPCfgDir  string
	sesPCfgPath string
	sesPCfg     *config.CGRConfig
	sesPRPC     *rpc.Client

	sesPTests = []func(t *testing.T){
		testSesPItLoadConfig,
		testSesPItResetDataDB,
		testSesPItResetStorDB,
		testSesPItStartEngine,
		testSesPItRPCConn,
		testSesPItBenchmark,
		// testSesPItCheckAccounts,
		testSesPItStopCgrEngine,
	}
)

func TestSesPIt(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		sesPCfgDir = "tutinternal"
	case utils.MetaMySQL:
		sesPCfgDir = "tutmysql"
	case utils.MetaMongo:
		sesPCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sesPTests {
		t.Run(sesPCfgDir, stest)
	}
}

func testSesPItLoadConfig(t *testing.T) {
	sesPCfgPath = path.Join(*dataDir, "conf", "samples", sesPCfgDir)
	if sesPCfg, err = config.NewCGRConfigFromPath(sesPCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesPItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesPCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesPItResetStorDB(t *testing.T) {
	if err := engine.InitStorDb(sesPCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesPItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesPCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesPItRPCConn(t *testing.T) {
	var err error
	sesPRPC, err = newRPCClient(sesPCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func setAccounts(t *testing.T, n int) (err error) {
	for i := 0; i < n; i++ {
		fmt.Println(fmt.Sprintf("acc%d", i))
		if err := setAccBalance(fmt.Sprintf("acc%d", i)); err != nil {
			t.Error(err)
		}
	}
	return nil
}

func getAccounts(ids []string) (accounts *[]interface{}, err error) {
	var reply *[]interface{}

	attr := &utils.AttrGetAccounts{
		Tenant:     "cgrates.org",
		AccountIDs: ids,
	}
	err = sesPRPC.Call("APIerSv1.GetAccounts", attr, &reply)
	if err != nil {
		return
	}
	accounts = reply
	return
}

func setAccBalance(acc string) (err error) {
	args := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     acc,
		BalanceType: utils.MetaVoice,
		Value:       float64(24 * time.Hour),
		Balance: map[string]interface{}{
			utils.ID:            "TestBalance",
			utils.RatingSubject: "*zero1ms",
		},
	}
	var reply string

	err = sesPRPC.Call(utils.APIerSv1SetBalance, args, &reply)
	return err
}

func testSesPItBenchmark(t *testing.T) {
	// add 2 charger profiles
	var reply string
	args := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        "runID1",
			AttributeIDs: []string{"*none"},
		},
		APIOpts: map[string]interface{}{},
	}
	if err := sesPRPC.Call(utils.APIerSv1SetChargerProfile, args, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error("Expected OK")
	}

	args2 := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        "runID2",
			AttributeIDs: []string{"*none"},
		},
		APIOpts: map[string]interface{}{},
	}

	if err := sesPRPC.Call(utils.APIerSv1SetChargerProfile, args2, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error("Expected OK")
	}

	////////// charger profiles set

	if err := setAccounts(t, 100); err != nil {
		t.Error(err)
	}

	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
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
			APIOpts: map[string]interface{}{
				utils.OptsDebitInterval: 2 * time.Second,
			},
		},
	}

	var initRpl *sessions.V1InitSessionReply
	if err := sesPRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Fatal(err)
	}

	var accIDs []string
	for i := 0; i < 100; i++ {
		accIDs = append(accIDs, fmt.Sprintf("acc%d", i))
	}
	_, err := getAccounts(accIDs)
	if err != nil {
		t.Error(err)
	}
}

func testSesPItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
