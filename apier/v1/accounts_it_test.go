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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accExist       bool
	accCfgPath     string
	accCfg         *config.CGRConfig
	accRPC         *rpc.Client
	accAcount      = "refundAcc"
	accTenant      = "cgrates.org"
	accBallID      = "Balance1"
	accTestsConfig string //run tests for specific configuration

	accTests = []func(t *testing.T){
		testAccITLoadConfig,
		testAccITResetDataDB,
		testAccITResetStorDb,
		testAccITStartEngine,
		testAccITRPCConn,
		testAccITAddVoiceBalance,
		testAccITDebitBalance,
		testAccITDebitBalanceWithoutTenant,
		testAccITAddBalance,
		testAccITAddBalanceWithoutTenant,
		testAccITAddBalanceWithValue0,
		testAccITAddBalanceWithValueInMap,
		testAccITSetBalance,
		testAccITSetBalanceWithoutTenant,
		testAccITSetBalanceWithVaslue0,
		testAccITSetBalanceWithVaslueInMap,
		testAccITSetBalanceWithExtraData,
		testAccITSetBalanceWithExtraData2,
		testAccITSetBalanceTimingIds,
		testAccITAddBalanceWithNegative,
		testAccITGetDisabledAccounts,
		testAccITGetDisabledAccountsWithoutTenant,
		testAccITCountAccounts,
		testAccITCountAccountsWithoutTenant,
		testAccITRemoveAccountWithoutTenant,
		testAccITTPFromFolder,
		testAccITAddBalanceWithDestinations,
		testAccITAccountWithTriggers,
		testAccITAccountMonthlyEstimated,
		testAccITMultipleBalance,
		testAccITMultipleBalanceWithoutTenant,
		testAccITRemoveBalances,
		testAccITAddVoiceBalanceWithDestinations,
		testAccITStopCgrEngine,
	}
)

func TestAccITWithRemove(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		accTestsConfig = "tutinternal"
	case utils.MetaMySQL:
		accTestsConfig = "tutmysql"
	case utils.MetaMongo:
		accTestsConfig = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	accCfgPath = path.Join(*dataDir, "conf", "samples", accTestsConfig)

	for _, stest := range accTests {
		t.Run(accTestsConfig, stest)
	}
}

func TestAccITWithoutRemove(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		accTestsConfig = "acc_balance_keep_internal"
	case utils.MetaMySQL:
		accTestsConfig = "acc_balance_keep_mysql"
	case utils.MetaMongo:
		accTestsConfig = "acc_balance_keep_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatalf("Unknown Database type <%s>", *dbType)
	}
	if *encoding == utils.MetaGOB {
		accTestsConfig += "_gob"
	}
	accCfgPath = path.Join(*dataDir, "conf", "samples", accTestsConfig)

	accExist = true
	for _, stest := range accTests {
		t.Run(accTestsConfig, stest)
	}
}

func testAccITLoadConfig(t *testing.T) {
	var err error
	if accCfg, err = config.NewCGRConfigFromPath(accCfgPath); err != nil {
		t.Error(err)
	}
}

func testAccITResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(accCfg); err != nil {
		t.Fatal(err)
	}
}

func testAccITResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(accCfg); err != nil {
		t.Fatal(err)
	}
}

func testAccITStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testAccITRPCConn(t *testing.T) {
	var err error
	if accRPC, err = newRPCClient(accCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testAccountBalance(t *testing.T, sracc, srten, balType string, expected float64) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  srten,
		Account: sracc,
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[balType].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testBalanceIfExists(t *testing.T, acc, ten, balType, balID string) (has bool) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  ten,
		Account: acc,
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
		return false
	}
	for _, bal := range acnt.BalanceMap[balType] {
		if bal.ID == balID {
			return true
		}
	}
	return false
}

func testAccITAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      accTenant,
		Account:     accAcount,
		BalanceType: utils.MetaVoice,
		Value:       2 * float64(time.Second),
		Balance: map[string]interface{}{
			utils.ID:            accBallID,
			utils.RatingSubject: "*zero5ms",
			utils.ExpiryTime:    time.Now().Add(5 * time.Second).Format("2006-01-02 15:04:05"),
		},
	}
	var reply string
	if err := accRPC.Call(utils.APIerSv2SetBalance, &attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) {
		testAccountBalance(t, accAcount, accTenant, utils.MetaVoice, 2*float64(time.Second))
	})
}

func testAccITSetBalanceTimingIds(t *testing.T) {
	tpTiming := &utils.ApierTPTiming{
		TPid:      "TEST_TPID1",
		ID:        "Timing",
		Years:     "2017",
		Months:    "05",
		MonthDays: "01",
		WeekDays:  "1",
		Time:      "15:00:00Z",
	}
	var reply string
	if err := accRPC.Call(utils.APIerSv1SetTPTiming, &tpTiming, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	if err := accRPC.Call(utils.APIerSv1LoadTariffPlanFromStorDb,
		&AttrLoadTpFromStorDb{TPid: "TEST_TPID1"}, &reply); err != nil {
		t.Errorf("Got error on %s: %+v", utils.APIerSv1LoadTariffPlanFromStorDb, err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling %s got reply: %+v", utils.APIerSv1LoadTariffPlanFromStorDb, reply)
	}

	args := &utils.AttrSetBalance{
		Tenant:      accTenant,
		Account:     accAcount,
		BalanceType: utils.MetaVoice,
		Balance: map[string]interface{}{
			utils.ID:        "testBalanceID",
			utils.TimingIDs: "Timing",
		},
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}

	// verify if Timing IDs is populated
	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  accTenant,
		Account: accAcount,
	}
	eOut := []*engine.RITiming{
		{
			ID:        "Timing",
			Years:     utils.Years{2017},
			Months:    utils.Months{05},
			MonthDays: utils.MonthDays{1},
			WeekDays:  utils.WeekDays{1},
			StartTime: "15:00:00Z",
			EndTime:   "",
		},
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.MetaVoice] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testBalanceID" {
			if !reflect.DeepEqual(eOut, value.Timings) {
				t.Errorf("Expecting %+v, \nreceived: %+v", utils.ToJSON(eOut), utils.ToJSON(value.Timings))
			}
			break
		}
	}
}

func testAccITDebitBalance(t *testing.T) {
	time.Sleep(5 * time.Second)
	var reply string
	if err := accRPC.Call(utils.APIerSv1DebitBalance, &AttrAddBalance{
		Tenant:      accTenant,
		Account:     accAcount,
		BalanceType: utils.MetaVoice,
		Value:       0,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if has := testBalanceIfExists(t, accAcount, accTenant, utils.MetaVoice, accBallID); accExist != has {
		var exstr string
		if !accExist {
			exstr = "not "
		}
		t.Fatalf("Balance with ID %s should %s exist", accBallID, exstr)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance(t, accAcount, accTenant, utils.MetaVoice, 0) })
}

func testAccITDebitBalanceWithoutTenant(t *testing.T) {
	var reply string
	if err := accRPC.Call(utils.APIerSv1DebitBalance, &AttrAddBalance{
		Account:     accAcount,
		BalanceType: utils.MetaVoice,
		Value:       0,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if has := testBalanceIfExists(t, accAcount, accTenant, utils.MetaVoice, accBallID); accExist != has {
		var exstr string
		if !accExist {
			exstr = "not "
		}
		t.Fatalf("Balance with ID %s should %s exist", accBallID, exstr)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance(t, accAcount, accTenant, utils.MetaVoice, 0) })
}

func testAccITAddBalance(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccAddBalance",
		BalanceType: utils.MetaMonetary,
		Value:       1.5,
		Cdrlog:      true,
	}
	if err := accRPC.Call(utils.APIerSv1AddBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLog}}
	if err := accRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testAccITAddBalanceWithoutTenant(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{
		Account:     "testAccAddBalance",
		BalanceType: utils.MetaMonetary,
		Value:       1.5,
		Cdrlog:      true,
	}
	if err := accRPC.Call(utils.APIerSv1AddBalance, &attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLog}}
	if err := accRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testAccITSetBalance(t *testing.T) {
	var reply string
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccSetBalance",
		BalanceType: "*monetary",
		Value:       1.5,
		Balance: map[string]interface{}{
			utils.ID: "testAccSetBalance",
		},
		Cdrlog: true,
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLog}}
	if err := accRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 3 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testAccITSetBalanceWithoutTenant(t *testing.T) {
	var reply string
	attrs := &utils.AttrSetBalance{
		Account:     "testrandomAccoutSetBalance",
		BalanceType: "*monetary",
		Value:       1.5,
		Balance: map[string]interface{}{
			utils.ID: "testAccSetBalance",
		},
		Cdrlog: true,
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, &attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLog}}
	if err := accRPC.Call(utils.APIerSv1GetCDRs, &req, &cdrs); err != nil {
		t.Error(err)
	} else if len(cdrs) != 4 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}

}

func testAccITSetBalanceWithExtraData(t *testing.T) {
	extraDataMap := map[string]interface{}{
		"ExtraField":  "ExtraValue",
		"ExtraField2": "RandomValue",
	}
	var reply string
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccITSetBalanceWithExtraData",
		BalanceType: utils.MetaMonetary,
		Value:       1.5,
		Balance: map[string]interface{}{
			utils.ID: "testAccITSetBalanceWithExtraData",
		},
		Cdrlog:          true,
		ActionExtraData: &extraDataMap,
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLog}, Accounts: []string{"testAccITSetBalanceWithExtraData"}}
	if err := accRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else if len(cdrs[0].ExtraFields) != 2 {
		t.Error("Unexpected number of ExtraFields returned: ", len(cdrs[0].ExtraFields))
	}
}

func testAccITSetBalanceWithExtraData2(t *testing.T) {
	extraDataMap := map[string]interface{}{
		"ExtraField": "ExtraValue",
		"ActionVal":  "~*act.ActionValue",
	}
	var reply string
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccITSetBalanceWithExtraData2",
		BalanceType: utils.MetaMonetary,
		Value:       1.5,
		Balance: map[string]interface{}{
			utils.ID: "testAccITSetBalanceWithExtraData2",
		},
		Cdrlog:          true,
		ActionExtraData: &extraDataMap,
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLog}, Accounts: []string{"testAccITSetBalanceWithExtraData2"}}
	if err := accRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else if len(cdrs[0].ExtraFields) != 2 {
		t.Error("Unexpected number of ExtraFields returned: ", len(cdrs[0].ExtraFields))
	} else if cdrs[0].ExtraFields["ActionVal"] != "1.5" {
		t.Error("Unexpected value of ExtraFields[ActionVal] returned: ", cdrs[0].ExtraFields["ActionVal"])
	}
}

func testAccITAddBalanceWithNegative(t *testing.T) {
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "AddBalanceWithNegative",
	}

	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//topup with a negative value
	var reply string
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "AddBalanceWithNegative",
		BalanceType: utils.MetaMonetary,
		Value:       -3.5,
	}
	if err := accRPC.Call(utils.APIerSv1AddBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}
	//give time to create the account and execute the action
	time.Sleep(50 * time.Millisecond)

	attrAcc = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "AddBalanceWithNegative",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != 3.5 {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	if err := accRPC.Call(utils.APIerSv1DebitBalance, &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "AddBalanceWithNegative",
		BalanceType: utils.MetaMonetary,
		Value:       2,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != 1.5 {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	if err := accRPC.Call(utils.APIerSv1DebitBalance, &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "AddBalanceWithNegative",
		BalanceType: utils.MetaMonetary,
		Value:       -1,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != 0.5 {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testAccITGetDisabledAccounts(t *testing.T) {
	var reply string
	acnt1 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account1", ExtraOptions: map[string]bool{utils.Disabled: true}}
	acnt2 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account2", ExtraOptions: map[string]bool{utils.Disabled: false}}
	acnt3 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account3", ExtraOptions: map[string]bool{utils.Disabled: true}}
	acnt4 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account4", ExtraOptions: map[string]bool{utils.Disabled: true}}

	for _, account := range []utils.AttrSetAccount{acnt1, acnt2, acnt3, acnt4} {
		if err := accRPC.Call(utils.APIerSv1SetAccount, account, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
		}
	}

	var acnts []*engine.Account
	if err := accRPC.Call(utils.APIerSv2GetAccounts, &utils.AttrGetAccounts{Tenant: "cgrates.org", Filter: map[string]bool{utils.Disabled: true}},
		&acnts); err != nil {
		t.Error(err)
	} else if len(acnts) != 3 {
		t.Errorf("Accounts received: %+v", acnts)
	}
}

func testAccITGetDisabledAccountsWithoutTenant(t *testing.T) {
	var reply string
	acnt1 := utils.AttrSetAccount{Account: "account1", ExtraOptions: map[string]bool{utils.Disabled: true}}
	acnt2 := utils.AttrSetAccount{Account: "account2", ExtraOptions: map[string]bool{utils.Disabled: false}}
	acnt3 := utils.AttrSetAccount{Account: "account3", ExtraOptions: map[string]bool{utils.Disabled: true}}
	acnt4 := utils.AttrSetAccount{Account: "account4", ExtraOptions: map[string]bool{utils.Disabled: true}}

	for _, account := range []utils.AttrSetAccount{acnt1, acnt2, acnt3, acnt4} {
		if err := accRPC.Call(utils.APIerSv1SetAccount, account, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
		}
	}

	var acnts []*engine.Account
	if err := accRPC.Call(utils.APIerSv2GetAccounts, &utils.AttrGetAccounts{Filter: map[string]bool{utils.Disabled: true}},
		&acnts); err != nil {
		t.Error(err)
	} else if len(acnts) != 3 {
		t.Errorf("Accounts received: %+v", acnts)
	}
}

func testAccITRemoveAccountWithoutTenant(t *testing.T) {
	var reply string
	acnt1 := utils.AttrSetAccount{Account: "randomAccount"}
	if err := accRPC.Call(utils.APIerSv1SetAccount, acnt1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
	var result *engine.Account
	if err := accRPC.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Account: "randomAccount"},
		&result); err != nil {
		t.Fatal(err)
	} else if result.ID != "cgrates.org:randomAccount" {
		t.Errorf("Expected %+v, received %+v", "cgrates.org:randomAccount", result.ID)
	}

	if err := accRPC.Call(utils.APIerSv1RemoveAccount,
		&utils.AttrRemoveAccount{Account: "randomAccount"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.RemoveAccount received: %s", reply)
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Account: "randomAccount"},
		&result); err == nil || utils.ErrNotFound.Error() != err.Error() {
		t.Error(err)
	}
}

func testAccITCountAccounts(t *testing.T) {
	var reply int
	args := &utils.TenantWithOpts{
		Tenant: "cgrates.org",
	}
	if err := accRPC.Call(utils.APIerSv1GetAccountsCount, &args, &reply); err != nil {
		t.Error(err)
	} else if reply != 11 {
		t.Errorf("Expecting: %v, received: %v", 11, reply)
	}
}
func testAccITCountAccountsWithoutTenant(t *testing.T) {
	var reply int
	if err := accRPC.Call(utils.APIerSv1GetAccountsCount,
		&utils.TenantIDWithOpts{},
		&reply); err != nil {
		t.Error(err)
	} else if reply != 11 {
		t.Errorf("Expecting: %v, received: %v", 11, reply)
	}
}

func testAccITStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testAccITSetBalanceWithVaslue0(t *testing.T) {
	var reply string
	args := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccSetBalance",
		BalanceType: "*monetary",
		Balance: map[string]interface{}{
			utils.ID:     "testAccSetBalance",
			utils.Weight: 10,
		},
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}

	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  accTenant,
		Account: "testAccSetBalance",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testAccSetBalance" {
			if value.GetValue() != 1.5 {
				t.Errorf("Expecting %+v, received: %+v", 1.5, value.GetValue())
			}
			if value.Weight != 10 {
				t.Errorf("Expecting %+v, received: %+v", 10, value.Weight)
			}
			break
		}
	}
}

func testAccITSetBalanceWithVaslueInMap(t *testing.T) {
	var reply string
	args := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccSetBalance",
		BalanceType: "*monetary",
		Balance: map[string]interface{}{
			utils.ID:     "testAccSetBalance",
			utils.Weight: 10,
			utils.Value:  2,
		},
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}

	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  accTenant,
		Account: "testAccSetBalance",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testAccSetBalance" {
			if value.GetValue() != 2 {
				t.Errorf("Expecting %+v, received: %+v", 2, value.GetValue())
			}
			if value.Weight != 10 {
				t.Errorf("Expecting %+v, received: %+v", 10, value.Weight)
			}
			break
		}
	}
}

func testAccITAddBalanceWithValue0(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccAddBalance",
		BalanceType: utils.MetaMonetary,
	}
	if err := accRPC.Call(utils.APIerSv1AddBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}

	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  accTenant,
		Account: "testAccAddBalance",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testAccSetBalance" {
			if value.GetValue() != 1.5 {
				t.Errorf("Expecting %+v, received: %+v", 1.5, value.GetValue())
			}
			break
		}
	}
}

func testAccITAddBalanceWithValueInMap(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccAddBalance",
		BalanceType: utils.MetaMonetary,
		Balance: map[string]interface{}{
			utils.Value: 1.5,
		},
	}
	if err := accRPC.Call(utils.APIerSv1AddBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}

	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  accTenant,
		Account: "testAccAddBalance",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testAccSetBalance" {
			if value.GetValue() != 3 {
				t.Errorf("Expecting %+v, received: %+v", 3, value.GetValue())
			}
			break
		}
	}
}

// Load the tariff plan, creating accounts and their balances
func testAccITTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := accRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAccITAddBalanceWithDestinations(t *testing.T) {
	var reply string
	args := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccITAddBalanceWithDestinations",
		BalanceType: utils.MetaMonetary,
		Balance: map[string]interface{}{
			utils.ID:             "testAccITAddBalanceWithDestinations",
			utils.DestinationIDs: "DST_1002;!DST_1001;!DST_1003",
			utils.Weight:         10,
			utils.Value:          2,
		},
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}

	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  accTenant,
		Account: "testAccITAddBalanceWithDestinations",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testAccITAddBalanceWithDestinations" {
			if value.GetValue() != 2 {
				t.Errorf("Expecting %+v, received: %+v", 2, value.GetValue())
			}
			if value.Weight != 10 {
				t.Errorf("Expecting %+v, received: %+v", 10, value.Weight)
			}
			dstMp := utils.StringMap{
				"DST_1002": true, "DST_1001": false, "DST_1003": false,
			}
			if !reflect.DeepEqual(value.DestinationIDs, dstMp) {
				t.Errorf("Expecting %+v, received: %+v", dstMp, value.DestinationIDs)
			}

			break
		}
	}

	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptorWithOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "sms",
			Tenant:        "cgrates.org",
			Subject:       "testAccITAddBalanceWithDestinations",
			Account:       "testAccITAddBalanceWithDestinations",
			Destination:   "1003",
			DurationIndex: 0,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(time.Nanosecond),
		},
	}
	var cc engine.CallCost
	if err := accRPC.Call(utils.ResponderMaxDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.Debit: ", err.Error())
	} else if cc.GetDuration() != 0 {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}

	tStart = time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd = &engine.CallDescriptorWithOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "sms",
			Tenant:        "cgrates.org",
			Subject:       "testAccITAddBalanceWithDestinations",
			Account:       "testAccITAddBalanceWithDestinations",
			Destination:   "1002",
			DurationIndex: 0,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(time.Nanosecond),
		},
	}
	if err := accRPC.Call(utils.ResponderMaxDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.Debit: ", err.Error())
	} else if cc.GetDuration() != 1 {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}
}

func testAccITAccountWithTriggers(t *testing.T) {
	var reply string
	args := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccITAccountWithTriggers",
		BalanceType: utils.MetaMonetary,
		Balance: map[string]interface{}{
			utils.ID:     "testAccITAccountWithTriggers",
			utils.Weight: 10,
			utils.Value:  5,
		},
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}

	// add an action that contains topup and reset_triggers
	topupAction := &utils.AttrSetActions{ActionsId: "TOPUP_WITH_RESET_TRIGGER", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUpReset, BalanceId: "testAccITAccountWithTriggers",
			BalanceType: utils.MetaMonetary, Units: "5", Weight: 10.0},
		{Identifier: utils.MetaResetTriggers},
	}}

	if err := accRPC.Call(utils.APIerSv2SetActions, topupAction, &reply); err != nil {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	// add an action to be executed when the trigger is triggered
	actTrigger := &utils.AttrSetActions{ActionsId: "ACT_TRIGGER", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUp, BalanceId: "CustomBanalce",
			BalanceType: utils.MetaMonetary, Units: "5", Weight: 10.0},
	}}

	if err := accRPC.Call(utils.APIerSv2SetActions, actTrigger, &reply); err != nil {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	attrsAddTrigger := &AttrAddActionTrigger{Tenant: "cgrates.org", Account: "testAccITAccountWithTriggers", BalanceType: utils.MetaMonetary,
		ThresholdType: "*min_balance", ThresholdValue: 2, Weight: 10, ActionsId: "ACT_TRIGGER"}
	if err := accRPC.Call(utils.APIerSv1AddTriggeredAction, attrsAddTrigger, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddTriggeredAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddTriggeredAction received: %s", reply)
	}

	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  accTenant,
		Account: "testAccITAccountWithTriggers",
	}

	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	} else {
		for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
			if value.ID == "testAccITAccountWithTriggers" {
				if value.GetValue() != 5 {
					t.Errorf("Expecting %+v, received: %+v", 5, value.GetValue())
				}
				if value.Weight != 10 {
					t.Errorf("Expecting %+v, received: %+v", 10, value.Weight)
				}
				break
			}
		}
		if len(acnt.ActionTriggers) != 1 {
			t.Errorf("Expected 1, received: %+v", len(acnt.ActionTriggers))
		} else {
			if acnt.ActionTriggers[0].Executed {
				t.Errorf("Expected false, received: %+v", acnt.ActionTriggers[0].Executed)
			}
		}
	}

	// Debit balance will trigger the Trigger from the account
	if err := accRPC.Call(utils.APIerSv1DebitBalance, &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccITAccountWithTriggers",
		BalanceType: utils.MetaMonetary,
		Value:       3,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	} else {
		for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
			if value.ID == "testAccITAccountWithTriggers" {
				if value.GetValue() != 2 {
					t.Errorf("Expecting %+v, received: %+v", 2, value.GetValue())
				}
			} else if value.ID == "CustomBanalce" {
				if value.GetValue() != 5 {
					t.Errorf("Expecting %+v, received: %+v", 5, value.GetValue())
				}
			}
		}
		if len(acnt.ActionTriggers) != 1 {
			t.Errorf("Expected 1, received: %+v", len(acnt.ActionTriggers))
		} else {
			if !acnt.ActionTriggers[0].Executed {
				t.Errorf("Expected true, received: %+v", acnt.ActionTriggers[0].Executed)
			}
		}
	}

	// execute the action that topup_reset the balance and reset the trigger
	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "testAccITAccountWithTriggers",
		ActionsId: "TOPUP_WITH_RESET_TRIGGER"}
	if err := accRPC.Call(utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	acnt = engine.Account{}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	} else {
		for _, value := range acnt.BalanceMap[utils.MetaMonetary] {
			if value.ID == "testAccITAccountWithTriggers" {
				if value.GetValue() != 5 {
					t.Errorf("Expecting %+v, received: %+v", 5, value.GetValue())
				}
			} else if value.ID == "CustomBanalce" {
				if value.GetValue() != 5 {
					t.Errorf("Expecting %+v, received: %+v", 5, value.GetValue())
				}
			}
		}
		if len(acnt.ActionTriggers) != 1 {
			t.Errorf("Expected 1, received: %+v", len(acnt.ActionTriggers))
		} else {
			if acnt.ActionTriggers[0].Executed {
				t.Errorf("Expected false, received: %+v", acnt.ActionTriggers[0].Executed)
			}
		}
	}

}

func testAccITAccountMonthlyEstimated(t *testing.T) {
	var reply string
	// add an action that contains topup
	topupAction := &utils.AttrSetActions{ActionsId: "TOPUP_ACTION", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUpReset, BalanceId: "testAccITAccountMonthlyEstimated",
			BalanceType: utils.MetaMonetary, Units: "5", Weight: 10.0},
	}}

	if err := accRPC.Call(utils.APIerSv2SetActions, topupAction, &reply); err != nil {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	atms1 := &AttrSetActionPlan{
		Id:              "ATMS_1",
		ReloadScheduler: true,
		ActionPlan: []*AttrActionPlan{{
			ActionsId: "TOPUP_ACTION",
			MonthDays: "31",
			Time:      "00:00:00",
			Weight:    20.0,
			TimingID:  utils.MetaMonthlyEstimated,
		}},
	}
	if err := accRPC.Call(utils.APIerSv1SetActionPlan, &atms1, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply)
	}

	acnt1 := utils.AttrSetAccount{Tenant: "cgrates.org",
		Account:         "testAccITAccountMonthlyEstimated",
		ReloadScheduler: true,
		ActionPlanID:    "ATMS_1",
	}
	if err := accRPC.Call(utils.APIerSv1SetAccount, acnt1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}

	var aps []*engine.ActionPlan
	accIDsStrMp := utils.StringMap{
		"cgrates.org:testAccITAccountMonthlyEstimated": true,
	}
	if err := accRPC.Call(utils.APIerSv1GetActionPlan,
		&AttrGetActionPlan{ID: "ATMS_1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "ATMS_1" {
		t.Errorf("Expected: %v,\n received: %v", "AP_PACKAGE_10", aps[0].Id)
	} else if !reflect.DeepEqual(aps[0].AccountIDs, accIDsStrMp) {
		t.Errorf("Expected: %v,\n received: %v", accIDsStrMp, aps[0].AccountIDs)
	} else if len(aps[0].ActionTimings) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else {
		// verify the GetNextTimeStart
		endOfMonth := utils.GetEndOfMonth(time.Now())
		if execDay := aps[0].ActionTimings[0].GetNextStartTime(time.Now()).Day(); execDay != endOfMonth.Day() {
			t.Errorf("Expected: %v,\n received: %v", endOfMonth.Day(), execDay)
		}
	}
}

func testAccITMultipleBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalances{
		Tenant:  "cgrates.org",
		Account: "testAccITMultipleBalance",
		Balances: []*utils.AttrBalance{
			{
				BalanceType: utils.MetaVoice,
				Value:       2 * float64(time.Second),
				Balance: map[string]interface{}{
					utils.ID:            "Balance1",
					utils.RatingSubject: "*zero5ms",
				},
			},
			{
				BalanceType: utils.MetaVoice,
				Value:       10 * float64(time.Second),
				Balance: map[string]interface{}{
					utils.ID:            "Balance2",
					utils.RatingSubject: "*zero5ms",
				},
			},
			{
				BalanceType: utils.MetaMonetary,
				Value:       10,
				Balance: map[string]interface{}{
					utils.ID: "MonBalance",
				},
			},
		},
	}
	var reply string
	if err := accRPC.Call(utils.APIerSv1SetBalances, &attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testAccITMultipleBalance",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaVoice]) != 2 {
		t.Errorf("Expected %+v, received: %+v", 2, len(acnt.BalanceMap[utils.MetaVoice]))
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != float64(12*time.Second) {
		t.Errorf("Expected %+v, received: %+v", float64(12*time.Second), acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10.0 {
		t.Errorf("Expected %+v, received: %+v", 10.0, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

}

func testAccITMultipleBalanceWithoutTenant(t *testing.T) {
	attrSetBalance := utils.AttrSetBalances{
		Account: "testAccITMultipleBalance",
		Balances: []*utils.AttrBalance{
			{
				BalanceType: utils.MetaVoice,
				Value:       2 * float64(time.Second),
				Balance: map[string]interface{}{
					utils.ID:            "Balance1",
					utils.RatingSubject: "*zero5ms",
				},
			},
			{
				BalanceType: utils.MetaVoice,
				Value:       10 * float64(time.Second),
				Balance: map[string]interface{}{
					utils.ID:            "Balance2",
					utils.RatingSubject: "*zero5ms",
				},
			},
			{
				BalanceType: utils.MetaMonetary,
				Value:       10,
				Balance: map[string]interface{}{
					utils.ID: "MonBalance",
				},
			},
		},
	}
	var reply string
	if err := accRPC.Call(utils.APIerSv1SetBalances, &attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Account: "testAccITMultipleBalance",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.MetaVoice]) != 2 {
		t.Errorf("Expected %+v, received: %+v", 2, len(acnt.BalanceMap[utils.MetaVoice]))
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != float64(12*time.Second) {
		t.Errorf("Expected %+v, received: %+v", float64(12*time.Second), acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(acnt.BalanceMap[utils.MetaMonetary]))
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != 10.0 {
		t.Errorf("Expected %+v, received: %+v", 10.0, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testAccITRemoveBalances(t *testing.T) {
	var reply string
	if err := accRPC.Call(utils.APIerSv1RemoveBalances,
		&utils.AttrSetBalance{Account: "testAccITMultipleBalance", BalanceType: utils.MetaVoice},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	attrs := &AttrAddBalance{
		Account:     "testAccITMultipleBalance",
		BalanceType: utils.MetaMonetary,
		Value:       1.5,
		Cdrlog:      true,
	}
	if err := accRPC.Call(utils.APIerSv1AddBalance, &attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}
}

func testAccITAddVoiceBalanceWithDestinations(t *testing.T) {
	var reply string
	args := &utils.AttrSetBalance{
		Tenant:      "cgrates.com",
		Account:     "testAccITAddVoiceBalanceWithDestinations",
		BalanceType: utils.MetaVoice,
		Balance: map[string]interface{}{
			utils.ID:             "testAccITAddVoiceBalanceWithDestinations",
			utils.DestinationIDs: "DST_1002",
			utils.RatingSubject:  "RP_1001",
			utils.Weight:         10,
			utils.Value:          time.Hour,
		},
	}
	if err := accRPC.Call(utils.APIerSv1SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling SetBalance received: %s", reply)
	}

	var acnt engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.com",
		Account: "testAccITAddVoiceBalanceWithDestinations",
	}
	if err := accRPC.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.MetaVoice] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testAccITAddVoiceBalanceWithDestinations" {
			if value.GetValue() != 3.6e+12 {
				t.Errorf("Expecting %+v, received: %+v", 3.6e+12, value.GetValue())
			}
			if value.Weight != 10 {
				t.Errorf("Expecting %+v, received: %+v", 10, value.Weight)
			}
			if value.RatingSubject != "RP_1001" {
				t.Errorf("Expecting %+v, received: %+v", "RP_1001", value.Weight)
			}
			dstMp := utils.StringMap{
				"DST_1002": true,
			}
			if !reflect.DeepEqual(value.DestinationIDs, dstMp) {
				t.Errorf("Expecting %+v, received: %+v", dstMp, value.DestinationIDs)
			}
			break
		}
	}

	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptorWithOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.com",
			Subject:       "testAccITAddVoiceBalanceWithDestinations",
			Account:       "testAccITAddVoiceBalanceWithDestinations",
			Destination:   "1002",
			DurationIndex: time.Minute,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(time.Minute),
		},
	}
	var rply time.Duration
	if err := accRPC.Call(utils.ResponderGetMaxSessionTime, cd, &rply); err != nil {
		t.Error("Got error on Responder.Debit: ", err.Error())
	} else if rply != 0 {
		t.Errorf("Expecting %+v, received: %+v", 0, rply)
	}
}
