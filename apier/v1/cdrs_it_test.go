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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdrsCfgPath string
	cdrsCfg     *config.CGRConfig
	cdrsRpc     *rpc.Client
	cdrsConfDIR string // run the tests for specific configuration

	sTestsCDRsIT = []func(t *testing.T){
		testV1CDRsInitConfig,
		testV1CDRsInitDataDb,
		testV1CDRsInitCdrDb,
		testV1CDRsStartEngine,
		testV1CDRsRpcConn,
		testV1CDRsLoadTariffPlanFromFolder,
		testV1CDRsProcessEventWithRefund,
		testV1CDRsRefundOutOfSessionCost,
		testV1CDRsRefundCDR,
		testV1CDRsKillEngine,

		testV1CDRsInitConfig,
		testV1CDRsInitDataDb,
		testV1CDRsInitCdrDb,
		testV1CDRsStartEngine,
		testV1CDRsRpcConn,
		testV1CDRsLoadTariffPlanFromFolderSMS,
		testV1CDRsAddBalanceForSMS,
		testV1CDRsKillEngine,
	}
)

func TestCDRsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		cdrsConfDIR = "cdrsv1internal"
	case utils.MetaMySQL:
		cdrsConfDIR = "cdrsv1mysql"
	case utils.MetaMongo:
		cdrsConfDIR = "cdrsv1mongo"
	case utils.MetaPostgres:
		cdrsConfDIR = "cdrsv1postgres"
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsCDRsIT {
		t.Run(cdrsConfDIR, stest)
	}
}

func testV1CDRsInitConfig(t *testing.T) {
	var err error
	cdrsCfgPath = path.Join(*dataDir, "conf", "samples", cdrsConfDIR)
	if cdrsCfg, err = config.NewCGRConfigFromPath(cdrsCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testV1CDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testV1CDRsInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1CDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testV1CDRsRpcConn(t *testing.T) {
	var err error
	cdrsRpc, err = newRPCClient(cdrsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1CDRsLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst string
	if err := cdrsRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testV1CDRsProcessEventWithRefund(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testV1CDRsProcessEventWithRefund"}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.VOICE,
		Value:       120000000000,
		Balance: map[string]interface{}{
			utils.ID:     "BALANCE1",
			utils.Weight: 20,
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}
	attrSetBalance = utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.VOICE,
		Value:       180000000000,
		Balance: map[string]interface{}{
			utils.ID:     "BALANCE2",
			utils.Weight: 10,
		},
	}
	if err := cdrsRpc.Call(utils.APIerSv1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: <%s>", reply)
	}
	expectedVoice := 300000000000.0
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply != expectedVoice {
		t.Errorf("Expecting: %v, received: %v", expectedVoice, rply)
	}
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.RunID:        "testv1",
					utils.OriginID:     "testV1CDRsProcessEventWithRefund",
					utils.RequestType:  utils.META_PSEUDOPREPAID,
					utils.AccountField: "testV1CDRsProcessEventWithRefund",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
					utils.Usage:        3 * time.Minute,
				},
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	if err := cdrsRpc.Call(utils.APIerSv1GetCDRs, &utils.AttrGetCdrs{}, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0 {
			t.Errorf("Unexpected cost for CDR: %f", cdrs[0].Cost)
		}
	}
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if blc1 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE1"); blc1.Value != 0 {
		t.Errorf("Balance1 is: %s", utils.ToIJSON(blc1))
	} else if blc2 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE2"); blc2.Value != 120000000000 {
		t.Errorf("Balance2 is: %s", utils.ToIJSON(blc2))
	}
	// without re-rate we should be denied
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err == nil {
		t.Error("should receive error here")
	}
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if blc1 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE1"); blc1.Value != 0 {
		t.Errorf("Balance1 is: %s", utils.ToIJSON(blc1))
	} else if blc2 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE2"); blc2.Value != 120000000000 {
		t.Errorf("Balance2 is: %s", utils.ToIJSON(blc2))
	}
	argsEv = &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, utils.MetaRerate},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.RunID:        "testv1",
					utils.OriginID:     "testV1CDRsProcessEventWithRefund",
					utils.RequestType:  utils.META_PSEUDOPREPAID,
					utils.AccountField: "testV1CDRsProcessEventWithRefund",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
					utils.Usage:        time.Minute,
				},
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if blc1 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE1"); blc1.Value != 120000000000 { // refund is done after debit
		t.Errorf("Balance1 is: %s", utils.ToIJSON(blc1))
	} else if blc2 := acnt.GetBalanceWithID(utils.VOICE, "BALANCE2"); blc2.Value != 120000000000 {
		t.Errorf("Balance2 is: %s", utils.ToIJSON(blc2))
	}
	return
}

func testV1CDRsRefundOutOfSessionCost(t *testing.T) {
	//create a sessionCost and store it into storDB
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testV1CDRsRefundOutOfSessionCost"}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.MONETARY,
		Value:       123,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 20,
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}

	exp := 123.0
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v", exp, rply)
	}
	balanceUuid := acnt.BalanceMap[utils.MONETARY][0].Uuid

	attr := &engine.AttrCDRSStoreSMCost{
		Cost: &engine.SMCost{
			CGRID:      "test1",
			RunID:      utils.MetaDefault,
			OriginID:   "testV1CDRsRefundOutOfSessionCost",
			CostSource: utils.MetaSessionS,
			Usage:      3 * time.Minute,
			CostDetails: &engine.EventCost{
				CGRID:     "test1",
				RunID:     utils.MetaDefault,
				StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
				Usage:     utils.DurationPointer(3 * time.Minute),
				Cost:      utils.Float64Pointer(2.3),
				Charges: []*engine.ChargingInterval{
					{
						RatingID: "c1a5ab9",
						Increments: []*engine.ChargingIncrement{
							{
								Usage:          2 * time.Minute,
								Cost:           2.0,
								AccountingID:   "a012888",
								CompressFactor: 1,
							},
							{
								Usage:          time.Second,
								Cost:           0.005,
								AccountingID:   "44d6c02",
								CompressFactor: 60,
							},
						},
						CompressFactor: 1,
					},
				},
				AccountSummary: &engine.AccountSummary{
					Tenant: "cgrates.org",
					ID:     "testV1CDRsRefundOutOfSessionCost",
					BalanceSummaries: []*engine.BalanceSummary{
						{
							UUID:  balanceUuid,
							Type:  utils.MONETARY,
							Value: 50,
						},
					},
					AllowNegative: false,
					Disabled:      false,
				},
				Rating: engine.Rating{
					"c1a5ab9": &engine.RatingUnit{
						ConnectFee:       0.1,
						RoundingMethod:   "*up",
						RoundingDecimals: 5,
						RatesID:          "ec1a177",
						RatingFiltersID:  "43e77dc",
					},
				},
				Accounting: engine.Accounting{
					"a012888": &engine.BalanceCharge{
						AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
						BalanceUUID: balanceUuid,
						Units:       120.7,
					},
					"44d6c02": &engine.BalanceCharge{
						AccountID:   "cgrates.org:testV1CDRsRefundOutOfSessionCost",
						BalanceUUID: balanceUuid,
						Units:       120.7,
					},
				},
				Rates: engine.ChargedRates{
					"ec1a177": engine.RateGroups{
						&engine.RGRate{
							GroupIntervalStart: 0,
							Value:              0.01,
							RateIncrement:      time.Minute,
							RateUnit:           time.Second},
					},
				},
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1StoreSessionCost,
		attr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}

	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.CGRID:        "test1",
					utils.RunID:        utils.MetaDefault,
					utils.OriginID:     "testV1CDRsRefundOutOfSessionCost",
					utils.RequestType:  utils.META_PREPAID,
					utils.AccountField: "testV1CDRsRefundOutOfSessionCost",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
					utils.Usage:        123 * time.Minute,
				},
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	// Initial the balance was 123.0
	// after refunc the balance become 123.0+2.3=125.3
	exp = 124.0454
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v", exp, rply)
	}
}

func testV1CDRsRefundCDR(t *testing.T) {
	//create a sessionCost and store it into storDB
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testV1CDRsRefundCDR"}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      acntAttrs.Tenant,
		Account:     acntAttrs.Account,
		BalanceType: utils.MONETARY,
		Value:       123,
		Balance: map[string]interface{}{
			utils.ID:     utils.MetaDefault,
			utils.Weight: 20,
		},
	}
	var reply string
	if err := cdrsRpc.Call(utils.APIerSv1SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("received: %s", reply)
	}

	exp := 123.0
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v", exp, rply)
	}

	balanceUuid := acnt.BalanceMap[utils.MONETARY][0].Uuid

	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRefund},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.RunID:        utils.MetaDefault,
					utils.OriginID:     "testV1CDRsRefundCDR",
					utils.RequestType:  utils.META_PSEUDOPREPAID,
					utils.AccountField: "testV1CDRsRefundCDR",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
					utils.Usage:        10 * time.Minute,
					utils.CostDetails: &engine.EventCost{
						CGRID:     "test1",
						RunID:     utils.MetaDefault,
						StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
						Usage:     utils.DurationPointer(3 * time.Minute),
						Cost:      utils.Float64Pointer(2.3),
						Charges: []*engine.ChargingInterval{
							{
								RatingID: "c1a5ab9",
								Increments: []*engine.ChargingIncrement{
									{
										Usage:          2 * time.Minute,
										Cost:           2.0,
										AccountingID:   "a012888",
										CompressFactor: 1,
									},
									{
										Usage:          time.Second,
										Cost:           0.005,
										AccountingID:   "44d6c02",
										CompressFactor: 60,
									},
								},
								CompressFactor: 1,
							},
						},
						AccountSummary: &engine.AccountSummary{
							Tenant: "cgrates.org",
							ID:     "testV1CDRsRefundCDR",
							BalanceSummaries: []*engine.BalanceSummary{
								{
									UUID:  balanceUuid,
									Type:  utils.MONETARY,
									Value: 50,
								},
							},
							AllowNegative: false,
							Disabled:      false,
						},
						Rating: engine.Rating{
							"c1a5ab9": &engine.RatingUnit{
								ConnectFee:       0.1,
								RoundingMethod:   "*up",
								RoundingDecimals: 5,
								RatesID:          "ec1a177",
								RatingFiltersID:  "43e77dc",
							},
						},
						Accounting: engine.Accounting{
							"a012888": &engine.BalanceCharge{
								AccountID:   "cgrates.org:testV1CDRsRefundCDR",
								BalanceUUID: balanceUuid,
								Units:       120.7,
							},
							"44d6c02": &engine.BalanceCharge{
								AccountID:   "cgrates.org:testV1CDRsRefundCDR",
								BalanceUUID: balanceUuid,
								Units:       120.7,
							},
						},
						Rates: engine.ChargedRates{
							"ec1a177": engine.RateGroups{
								&engine.RGRate{
									GroupIntervalStart: 0,
									Value:              0.01,
									RateIncrement:      time.Minute,
									RateUnit:           time.Second},
							},
						},
					},
				},
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	// Initial the balance was 123.0
	// after refund the balance become 123.0 + 2.3 = 125.3
	exp = 125.3
	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MONETARY].GetTotalValue(); rply != exp {
		t.Errorf("Expecting: %v, received: %v", exp, rply)
	}
}

func testV1CDRsLoadTariffPlanFromFolderSMS(t *testing.T) {
	var loadInst string
	if err := cdrsRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*dataDir, "tariffplans", "tutorial")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testV1CDRsAddBalanceForSMS(t *testing.T) {
	var reply string
	attrs := &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testV1CDRsAddBalanceForSMS",
		BalanceType: utils.SMS,
		Value:       1000,
	}
	if err := cdrsRpc.Call(utils.APIerSv1AddBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}

	attrs = &AttrAddBalance{
		Tenant:      "cgrates.org",
		Account:     "testV1CDRsAddBalanceForSMS",
		BalanceType: utils.MONETARY,
		Value:       10,
	}
	if err := cdrsRpc.Call(utils.APIerSv1AddBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.AddBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.AddBalance received: %s", reply)
	}

	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testV1CDRsAddBalanceForSMS",
	}

	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.SMS]) != 1 {
		t.Errorf("Expecting: %v, received: %v", 1, len(acnt.BalanceMap[utils.SMS]))
	} else if acnt.BalanceMap[utils.SMS].GetTotalValue() != 1000 {
		t.Errorf("Expecting: %v, received: %v", 1000, acnt.BalanceMap[utils.SMS].GetTotalValue())
	} else if len(acnt.BalanceMap[utils.MONETARY]) != 1 {
		t.Errorf("Expecting: %v, received: %v", 1, len(acnt.BalanceMap[utils.MONETARY]))
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 10 {
		t.Errorf("Expecting: %v, received: %v", 10, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREventWithOpts: utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Event: map[string]interface{}{
					utils.CGRID:        "asdfas",
					utils.ToR:          utils.SMS,
					utils.Category:     "sms",
					utils.RunID:        utils.MetaDefault,
					utils.OriginID:     "testV1CDRsAddBalanceForSMS",
					utils.RequestType:  utils.META_PREPAID,
					utils.AccountField: "testV1CDRsAddBalanceForSMS",
					utils.Destination:  "+4986517174963",
					utils.AnswerTime:   time.Date(2019, 11, 27, 12, 21, 26, 0, time.UTC),
					utils.Usage:        1,
				},
			},
		},
	}
	if err := cdrsRpc.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	if err := cdrsRpc.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if len(acnt.BalanceMap[utils.SMS]) != 1 {
		t.Errorf("Expecting: %v, received: %v", 1, len(acnt.BalanceMap[utils.SMS]))
	} else if acnt.BalanceMap[utils.SMS].GetTotalValue() != 999 {
		t.Errorf("Expecting: %v, received: %v", 999, acnt.BalanceMap[utils.SMS].GetTotalValue())
	} else if len(acnt.BalanceMap[utils.MONETARY]) != 1 {
		t.Errorf("Expecting: %v, received: %v", 1, len(acnt.BalanceMap[utils.MONETARY]))
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 10 {
		t.Errorf("Expecting: %v, received: %v", 10, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}

}

func testV1CDRsKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
