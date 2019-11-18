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
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accExist   bool
	accCfgPath string
	accCfg     *config.CGRConfig
	accRPC     *rpc.Client
	accAcount  = "refundAcc"
	accTenant  = "cgrates.org"
	accBallID  = "Balance1"

	accTests = []func(t *testing.T){
		testAccITLoadConfig,
		testAccITResetDataDB,
		testAccITResetStorDb,
		testAccITStartEngine,
		testAccITRPCConn,
		testAccITAddVoiceBalance,
		testAccITDebitBalance,
		testAccITAddBalance,
		testAccITSetBalance,
		testAccITSetBalanceWithExtraData,
		testAccITSetBalanceWithExtraData2,
		testAccITSetBalanceTimingIds,
		testAccITAddBalanceWithNegative,
		testAccITGetDisabledAccounts,
		testAccITStopCgrEngine,
	}
)

func TestAccITWithRemove(t *testing.T) {
	accCfgPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	for _, test := range accTests {
		t.Run("TestAccIT", test)
	}
}

func TestAccITWithoutRemove(t *testing.T) {
	accCfgPath = path.Join(*dataDir, "conf", "samples", "acc_balance_keep")
	accExist = true
	for _, test := range accTests {
		t.Run("TestAccIT", test)
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
	accRPC, err = jsonrpc.Dial("tcp", accCfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func testAccountBalance(t *testing.T, sracc, srten, balType string, expected float64) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  srten,
		Account: sracc,
	}
	if err := accRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
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
	if err := accRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
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
		Tenant:        accTenant,
		Account:       accAcount,
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer(accBallID),
		Value:         utils.Float64Pointer(2 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero5ms"),
		ExpiryTime:    utils.StringPointer(time.Now().Add(5 * time.Second).Format("2006-01-02 15:04:05")),
	}
	var reply string
	if err := accRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance(t, accAcount, accTenant, utils.VOICE, 2*float64(time.Second)) })
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
	if err := accRPC.Call("ApierV1.SetTPTiming", tpTiming, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	if err := accRPC.Call("ApierV1.LoadTariffPlanFromStorDb",
		AttrLoadTpFromStorDb{TPid: "TEST_TPID1"}, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromStorDb: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.LoadTariffPlanFromStorDb got reply: ", reply)
	}

	args := &utils.AttrSetBalance{
		Tenant:      accTenant,
		Account:     accAcount,
		TimingIds:   utils.StringPointer("Timing"),
		BalanceType: utils.VOICE,
		BalanceID:   utils.StringPointer("testBalanceID"),
	}
	if err := accRPC.Call(utils.ApierV1SetBalance, args, &reply); err != nil {
		t.Error("Got error on SetBalance: ", err.Error())
	} else if reply != "OK" {
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
			Years:     utils.Years{2017},
			Months:    utils.Months{05},
			MonthDays: utils.MonthDays{1},
			WeekDays:  utils.WeekDays{1},
			StartTime: "15:00:00Z",
			EndTime:   "",
		},
	}
	if err := accRPC.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Fatal(err)
	}

	for _, value := range acnt.BalanceMap[utils.VOICE] {
		// check only where balance ID is testBalanceID (SetBalance function call was made with this Balance ID)
		if value.ID == "testBalanceID" {
			if !reflect.DeepEqual(eOut, value.Timings) {
				t.Errorf("\nExpecting %+v, \nreceived: %+v", utils.ToJSON(eOut), utils.ToJSON(value.Timings))
			}
			break
		}
	}
}

func testAccITDebitBalance(t *testing.T) {
	time.Sleep(5 * time.Second)
	var reply string
	if err := accRPC.Call("ApierV1.DebitBalance", &AttrAddBalance{
		Tenant:      accTenant,
		Account:     accAcount,
		BalanceType: utils.VOICE,
		Value:       0,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if has := testBalanceIfExists(t, accAcount, accTenant, utils.VOICE, accBallID); accExist != has {
		var exstr string
		if !accExist {
			exstr = "not "
		}
		t.Fatalf("Balance with ID %s should %sexist", accBallID, exstr)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance(t, accAcount, accTenant, utils.VOICE, 0) })

}

func testAccITAddBalance(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccAddBalance",
		BalanceType: utils.MONETARY,
		Value:       1.5,
		Cdrlog:      utils.BoolPointer(true),
	}
	if err := accRPC.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLOG}}
	if err := accRPC.Call(utils.ApierV2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testAccITSetBalance(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{Tenant: "cgrates.org", Account: "testAccSetBalance",
		BalanceId:   utils.StringPointer("testAccSetBalance"),
		BalanceType: "*monetary", Value: 1.5, Cdrlog: utils.BoolPointer(true)}
	if err := accRPC.Call("ApierV1.SetBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.SetBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.SetBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLOG}}
	if err := accRPC.Call(utils.ApierV2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testAccITSetBalanceWithExtraData(t *testing.T) {
	extraDataMap := map[string]interface{}{
		"ExtraField":  "ExtraValue",
		"ExtraField2": "RandomValue",
	}
	var reply string
	attrs := &AttrAddBalance{Tenant: "cgrates.org", Account: "testAccITSetBalanceWithExtraData",
		BalanceId:   utils.StringPointer("testAccITSetBalanceWithExtraData"),
		BalanceType: "*monetary", Value: 1.5, Cdrlog: utils.BoolPointer(true),
		ExtraData: &extraDataMap}
	if err := accRPC.Call("ApierV1.SetBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.SetBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.SetBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLOG}, Accounts: []string{"testAccITSetBalanceWithExtraData"}}
	if err := accRPC.Call(utils.ApierV2GetCDRs, req, &cdrs); err != nil {
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
		"ActionVal":  "~ActionValue",
	}
	var reply string
	attrs := &AttrAddBalance{Tenant: "cgrates.org", Account: "testAccITSetBalanceWithExtraData2",
		BalanceId:   utils.StringPointer("testAccITSetBalanceWithExtraData2"),
		BalanceType: "*monetary", Value: 1.5, Cdrlog: utils.BoolPointer(true),
		ExtraData: &extraDataMap}
	if err := accRPC.Call("ApierV1.SetBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.SetBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.SetBalance received: %s", reply)
	}
	time.Sleep(50 * time.Millisecond)
	// verify the cdr from CdrLog
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Sources: []string{utils.CDRLOG}, Accounts: []string{"testAccITSetBalanceWithExtraData2"}}
	if err := accRPC.Call(utils.ApierV2GetCDRs, req, &cdrs); err != nil {
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

	if err := accRPC.Call("ApierV2.GetAccount", attrAcc, &acnt); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//topup with a negative value
	var reply string
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "AddBalanceWithNegative",
		BalanceType: "*monetary",
		Value:       -3.5,
	}
	if err := accRPC.Call("ApierV1.AddBalance", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.AddBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.AddBalance received: %s", reply)
	}
	//give time to create the account and execute the action
	time.Sleep(50 * time.Millisecond)

	attrAcc = &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "AddBalanceWithNegative",
	}
	if err := accRPC.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 3.5 {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	if err := accRPC.Call("ApierV1.DebitBalance", &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "AddBalanceWithNegative",
		BalanceType: utils.MONETARY,
		Value:       2,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	if err := accRPC.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 1.5 {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	if err := accRPC.Call("ApierV1.DebitBalance", &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "AddBalanceWithNegative",
		BalanceType: utils.MONETARY,
		Value:       -1,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	if err := accRPC.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 0.5 {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func testAccITGetDisabledAccounts(t *testing.T) {
	var reply string
	acnt1 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account1", Disabled: utils.BoolPointer(true)}
	acnt2 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account2", Disabled: utils.BoolPointer(false)}
	acnt3 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account3", Disabled: utils.BoolPointer(true)}
	acnt4 := utils.AttrSetAccount{Tenant: "cgrates.org", Account: "account4", Disabled: utils.BoolPointer(true)}

	for _, account := range []utils.AttrSetAccount{acnt1, acnt2, acnt3, acnt4} {
		if err := accRPC.Call("ApierV1.SetAccount", account, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("Calling ApierV1.SetAccount received: %s", reply)
		}
	}

	var acnts []*engine.Account
	if err := accRPC.Call("ApierV2.GetAccounts", utils.AttrGetAccounts{Tenant: "cgrates.org", Disabled: utils.BoolPointer(true)},
		&acnts); err != nil {
		t.Error(err)
	} else if len(acnts) != 3 {
		t.Errorf("Accounts received: %+v", acnts)
	}
}

func testAccITStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
