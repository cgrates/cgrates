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
	"net/rpc"
	"net/rpc/jsonrpc"
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
	dataCfgPath string
	dataCfg     *config.CGRConfig
	dataRpc     *rpc.Client
	dataConfDIR string //run tests for specific configuration
	dataDelay   int

	sTestsData = []func(t *testing.T){
		testV1DataLoadConfig,
		testV1DataInitDataDb,
		testV1DataResetStorDb,
		testV1DataStartEngine,
		testV1DataRpcConn,
		testV1DataLoadTarrifPlans,
		// testV1DataDataDebitUsageWith10Kilo,
		// testV1DataGetCostWith10Kilo,
		// testV1DataDebitBalanceWith10Kilo,
		// testV1DataDataDebitUsage1G0,
		// testV1DataGetCost1G0,
		// testV1DataDebitBalance1G0,
		testV1DataInitSession,
		testV1DataUpdateWith1Mo,
		testV1DataUpdateWith1Go,
		testV1DataStopEngine,
	}
)

// Test start here
func TestDataITMongo(t *testing.T) {
	dataConfDIR = "tutmongo"
	for _, stest := range sTestsData {
		t.Run(dataConfDIR, stest)
	}
}

func testV1DataLoadConfig(t *testing.T) {
	var err error
	dataCfgPath = path.Join(*dataDir, "conf", "samples", dataConfDIR)
	if dataCfg, err = config.NewCGRConfigFromPath(dataCfgPath); err != nil {
		t.Error(err)
	}
	dataDelay = 1000
}

func testV1DataInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(dataCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1DataResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(dataCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1DataStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dataCfgPath, dataDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1DataRpcConn(t *testing.T) {
	var err error
	dataRpc, err = jsonrpc.Dial("tcp", dataCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1DataGetAccountBeforeSet(t *testing.T) {
	var reply *engine.Account
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1DataLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testData")}
	if err := dataRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1DataDataDebitUsageWith10Kilo(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "testV1DataDataCost",
		BalanceType:   utils.DATA,
		Categories:    utils.StringPointer("data"),
		BalanceID:     utils.StringPointer("testV1DataDataCost"),
		Value:         utils.Float64Pointer(356000000),
		RatingSubject: utils.StringPointer("*zero10000ns"),
	}
	var reply string
	if err := dataRpc.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	expected := 356000000.0
	var acc *engine.Account
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDataCost"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}

	usageRecord := &engine.UsageRecord{
		Tenant:      "cgrates.org",
		Account:     "testV1DataDataCost",
		Destination: "*any",
		Usage:       "256000000",
		ToR:         utils.DATA,
		Category:    "data",
		SetupTime:   time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String(),
		AnswerTime:  time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String(),
	}
	tStart := time.Now()
	if err := dataRpc.Call("ApierV1.DebitUsage",
		engine.UsageRecordWithArgDispatcher{UsageRecord: usageRecord}, &reply); err != nil {
		t.Error(err)
	}
	if time.Now().Sub(tStart) > time.Duration(50*time.Millisecond) {
		t.Error("Take's too long for GetDataCost")
	}

	expected = 100000000.0
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDataCost"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testV1DataGetCostWith10Kilo(t *testing.T) {
	attrs := v1.AttrGetDataCost{Category: "data", Tenant: "cgrates.org",
		Subject: "10kilo", AnswerTime: "*now", Usage: 256000000}
	var rply *engine.DataCost
	tStart := time.Now()
	if err := dataRpc.Call("ApierV1.GetDataCost", attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if rply.Cost != 25600.000000 {
		t.Errorf("Unexpected cost received: %f", rply.Cost)
	}
	if time.Now().Sub(tStart) > time.Duration(50*time.Millisecond) {
		t.Error("Take's too long for GetDataCost")
	}
}

func testV1DataDebitBalanceWith10Kilo(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "testV1DataDebitBalance",
		BalanceType:   utils.DATA,
		Categories:    utils.StringPointer("data"),
		BalanceID:     utils.StringPointer("testV1DataDebitBalance"),
		Value:         utils.Float64Pointer(356000000),
		RatingSubject: utils.StringPointer("*zero10000ns"),
	}
	var reply string
	if err := dataRpc.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	expected := 356000000.0
	var acc *engine.Account
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDebitBalance"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
	tStart := time.Now()
	if err := dataRpc.Call(utils.ApierV1DebitBalance, &v1.AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testV1DataDebitBalance",
		BalanceType: utils.DATA,
		Value:       256000000,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if time.Now().Sub(tStart) > time.Duration(50*time.Millisecond) {
		t.Error("Take's too long for GetDataCost")
	}

	expected = 100000000.0
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDebitBalance"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testV1DataDataDebitUsage1G0(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "testV1DataDataDebitUsage1G0",
		BalanceType:   utils.DATA,
		Categories:    utils.StringPointer("data"),
		BalanceID:     utils.StringPointer("testV1DataDataDebitUsage1G0"),
		Value:         utils.Float64Pointer(1100000000),
		RatingSubject: utils.StringPointer("*zero10000ns"),
	}
	var reply string
	if err := dataRpc.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	expected := 1100000000.0
	var acc *engine.Account
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDataDebitUsage1G0"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}

	usageRecord := &engine.UsageRecord{
		Tenant:      "cgrates.org",
		Account:     "testV1DataDataDebitUsage1G0",
		Destination: "*any",
		Usage:       "1000000000",
		ToR:         utils.DATA,
		Category:    "data",
		SetupTime:   time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String(),
		AnswerTime:  time.Date(2013, 11, 7, 7, 42, 20, 0, time.UTC).String(),
	}
	tStart := time.Now()
	if err := dataRpc.Call("ApierV1.DebitUsage",
		engine.UsageRecordWithArgDispatcher{UsageRecord: usageRecord}, &reply); err != nil {
		t.Error(err)
	}
	if time.Now().Sub(tStart) > time.Duration(50*time.Millisecond) {
		t.Errorf("Take's too long for GetDataCost : %+v", time.Now().Sub(tStart))
	}

	expected = 100000000.0
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDataDebitUsage1G0"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testV1DataGetCost1G0(t *testing.T) {
	attrs := v1.AttrGetDataCost{Category: "data", Tenant: "cgrates.org",
		Subject: "10kilo", AnswerTime: "*now", Usage: 1000000000}
	var rply *engine.DataCost
	tStart := time.Now()
	if err := dataRpc.Call("ApierV1.GetDataCost", attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if rply.Cost != 100000.000000 {
		t.Errorf("Unexpected cost received: %f", rply.Cost)
	}
	if time.Now().Sub(tStart) > time.Duration(50*time.Millisecond) {
		t.Error("Take's too long for GetDataCost")
	}
}

func testV1DataDebitBalance1G0(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "testV1DataDebitBalance1G0",
		BalanceType:   utils.DATA,
		Categories:    utils.StringPointer("data"),
		BalanceID:     utils.StringPointer("testV1DataDebitBalance1G0"),
		Value:         utils.Float64Pointer(1100000000),
		RatingSubject: utils.StringPointer("*zero10000ns"),
	}
	var reply string
	if err := dataRpc.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	expected := 1100000000.0
	var acc *engine.Account
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDebitBalance1G0"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
	tStart := time.Now()
	if err := dataRpc.Call(utils.ApierV1DebitBalance, &v1.AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testV1DataDebitBalance1G0",
		BalanceType: utils.DATA,
		Value:       1000000000,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if time.Now().Sub(tStart) > time.Duration(50*time.Millisecond) {
		t.Error("Take's too long for GetDataCost")
	}

	expected = 100000000.0
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataDebitBalance1G0"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testV1DataInitSession(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "testV1DataInitSession",
		BalanceType:   utils.DATA,
		Categories:    utils.StringPointer("data"),
		BalanceID:     utils.StringPointer("testV1DataInitSession"),
		Value:         utils.Float64Pointer(1100000000),
		RatingSubject: utils.StringPointer("*zero10000ns"),
	}
	var reply string
	if err := dataRpc.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	expected := 1100000000.0
	var acc *engine.Account
	if err := dataRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testV1DataInitSession"},
		&acc); err != nil {
		t.Error(err)
	} else if _, has := acc.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(acc.BalanceMap[utils.DATA]))
	} else if rply := acc.BalanceMap[utils.DATA].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}

	initUsage := 10000
	args := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1DataInitSession",
			Event: map[string]interface{}{
				utils.Tenant:         "cgrates.org",
				utils.ToR:            utils.DATA,
				utils.Category:       "data",
				utils.OriginID:       "781512335",
				utils.RequestType:    utils.META_PREPAID,
				utils.Account:        "testV1DataInitSession",
				utils.Subject:        "10kilo",
				utils.Destination:    "*any",
				utils.SetupTime:      time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:     time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:          initUsage,
				"SessionTTL":         "300s",
				"SessionTTLLastUsed": "0s",
				"SessionTTLMaxDelay": "1800s",
				"SessionTTLUsage":    "0s",
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := dataRpc.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := dataRpc.Call(utils.SessionSv1GetActiveSessions,
		&utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v",
			utils.ToJSON(aSessions), len(aSessions))
	}
}

func testV1DataUpdateWith1Mo(t *testing.T) {
	reqUsage := 1000000
	args := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1DataUpdateWith1Mo",
			Event: map[string]interface{}{
				utils.Tenant:         "cgrates.org",
				utils.ToR:            utils.DATA,
				utils.Category:       "data",
				"InitialOriginID":    "781512335",
				"LastUsed":           0,
				"OriginID":           "781512335-11",
				utils.RequestType:    utils.META_PREPAID,
				utils.Account:        "testV1DataInitSession",
				utils.Subject:        "10kilo",
				utils.Destination:    "*any",
				utils.SetupTime:      time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:     time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:          reqUsage,
				"SessionTTL":         "28807s",
				"SessionTTLLastUsed": "0s",
				"SessionTTLMaxDelay": "1800s",
				"SessionTTLUsage":    "0s",
			},
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := dataRpc.Call(utils.SessionSv1UpdateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := dataRpc.Call(utils.SessionSv1GetActiveSessions,
		&utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v",
			utils.ToJSON(aSessions), len(aSessions))
	}
}

func testV1DataUpdateWith1Go(t *testing.T) {
	reqUsage := 1000000000
	args := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1DataUpdateWith1Go",
			Event: map[string]interface{}{
				utils.Tenant:         "cgrates.org",
				utils.ToR:            utils.DATA,
				utils.Category:       "data",
				"InitialOriginID":    "781512335",
				"LastUsed":           946405,
				"OriginID":           "781512335-11",
				utils.RequestType:    utils.META_PREPAID,
				utils.Account:        "testV1DataInitSession",
				utils.Subject:        "10kilo",
				utils.Destination:    "*any",
				utils.SetupTime:      time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:     time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:          reqUsage,
				"SessionTTL":         "28807s",
				"SessionTTLLastUsed": "0s",
				"SessionTTLMaxDelay": "1800s",
				"SessionTTLUsage":    "0s",
			},
		},
	}
	var rply sessions.V1UpdateSessionReply
	if err := dataRpc.Call(utils.SessionSv1UpdateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := dataRpc.Call(utils.SessionSv1GetActiveSessions,
		&utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v",
			utils.ToJSON(aSessions), len(aSessions))
	}

}

func testV1DataStopEngine(t *testing.T) {
	if err := engine.KillEngine(dataDelay); err != nil {
		t.Error(err)
	}
}
