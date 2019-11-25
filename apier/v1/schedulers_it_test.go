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
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
)

var (
	schedCfgPath string
	schedCfg     *config.CGRConfig
	schedRpc     *rpc.Client
	schedConfDIR string //run tests for specific configuration
)
var sTestsSchedFiltered = []func(t *testing.T){
	testSchedLoadConfig,
	testSchedInitDataDb,
	testSchedResetStorDb,
	testSchedStartEngine,
	testSchedRpcConn,
	testSchedFromFolder,
	testSchedVeifyAllAccounts,
	testSchedVeifyAccount1001,
	testSchedVeifyAccount1002and1003,
	testSchedStopEngine,
}

//TestSchedWithoutFilters will execute action for all accounts
func TestSchedWithoutFilters(t *testing.T) {
	schedConfDIR = "tutmysql"
	for _, stest := range sTestsSchedFiltered {
		t.Run(schedConfDIR, stest)
	}
}

//TestSchedWithFiltersSingleAccount will execute actions only for account 1001
func TestSchedWithFiltersSingleAccount(t *testing.T) {
	schedConfDIR = "filtered_scheduler"
	for _, stest := range sTestsSchedFiltered {
		t.Run(schedConfDIR, stest)
	}
}

//TestSchedWithFilters2 will execute actions for accounts 1002 and 1003 ( 1001 will be skiped )
func TestSchedWithFilters2(t *testing.T) {
	schedConfDIR = "filtered_scheduler2"
	for _, stest := range sTestsSchedFiltered {
		t.Run(schedConfDIR, stest)
	}
}

func testSchedLoadConfig(t *testing.T) {
	var err error
	schedCfgPath = path.Join(*dataDir, "conf", "samples", schedConfDIR)
	if schedCfg, err = config.NewCGRConfigFromPath(schedCfgPath); err != nil {
		t.Error(err)
	}
}

func testSchedInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(schedCfg); err != nil {
		t.Fatal(err)
	}
}

func testSchedResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(schedCfg); err != nil {
		t.Fatal(err)
	}
}

func testSchedStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(schedCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSchedRpcConn(t *testing.T) {
	var err error
	schedRpc, err = newRPCClient(schedCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testSchedFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := schedRpc.Call(utils.ApierV1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testSchedVeifyAllAccounts(t *testing.T) {
	if schedConfDIR == "filtered_scheduler" || schedConfDIR == "filtered_scheduler2" {
		t.SkipNow()
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}
	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}
	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1003",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}
}

func testSchedVeifyAccount1001(t *testing.T) {
	if schedConfDIR == "tutmysql" || schedConfDIR == "filtered_scheduler2" {
		t.SkipNow()
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}

	acnt = nil // in case of gob ( it doesn't update the empty fields)
	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if lenBal := len(acnt.BalanceMap[utils.MONETARY]); lenBal != 0 {
		t.Errorf("Expecting: %v, received: %v",
			0, lenBal)
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1003",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if lenBal := len(acnt.BalanceMap[utils.MONETARY]); lenBal != 0 {
		t.Errorf("Expecting: %v, received: %v",
			0, lenBal)
	}

}

func testSchedVeifyAccount1002and1003(t *testing.T) {
	if schedConfDIR == "tutmysql" || schedConfDIR == "filtered_scheduler" {
		t.SkipNow()
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if lenBal := len(acnt.BalanceMap[utils.MONETARY]); lenBal != 0 {
		t.Errorf("Expecting: %v, received: %v",
			0, lenBal)
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1002",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}

	attrs = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1003",
	}
	if err := schedRpc.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != 10 {
		t.Errorf("Expecting: %v, received: %v",
			10, rply)
	}

}

func testSchedStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
