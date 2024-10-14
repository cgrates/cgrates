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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesRndCfgPath string
	sesRndCfgDIR  string
	sesRndCfg     *config.CGRConfig
	sesRndRPC     *birpc.Client
	sesRndAccount = "testAccount"
	sesRndTenant  = "cgrates.org"

	sesRndExpCost         float64
	sesRndExpMaxUsage     time.Duration
	sesRndExpBalanceValue float64

	sesRndCgrEv = &utils.CGREvent{
		Tenant: sesRndTenant,
		Event: map[string]any{
			utils.Tenant:       sesRndTenant,
			utils.Category:     utils.Call,
			utils.ToR:          utils.MetaVoice,
			utils.AccountField: sesRndAccount,
			utils.Destination:  "TEST",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        10 * time.Second,
		},
		APIOpts: map[string]any{
			utils.OptsSessionsTTL:   0,
			utils.OptsDebitInterval: time.Second,
		},
	}

	sTestSesRndIt = []func(t *testing.T){
		testSesRndItLoadConfig,
		testSesRndItResetDataDB,
		testSesRndItResetStorDb,
		testSesRndItStartEngine,
		testSesRndItRPCConn,
		testSesRndItLoadRating,
		testSesRndItAddCharger,

		testSesRndItPreparePostpaidUP,
		testSesRndItAddVoiceBalance,
		testSesRndItPrepareCDRs,
		testSesRndItCheckCdrs,

		testSesRndItPreparePostpaidDOWN,
		testSesRndItAddVoiceBalance,
		testSesRndItPrepareCDRs,
		testSesRndItCheckCdrs,

		testSesRndItPreparePrepaidUP,
		testSesRndItAddVoiceBalance,
		testSesRndItPrepareCDRs,
		testSesRndItCheckCdrs,

		testSesRndItPreparePrepaidDOWN,
		testSesRndItAddVoiceBalance,
		testSesRndItPrepareCDRs,
		testSesRndItCheckCdrs,

		testSesRndItStopCgrEngine,
	}
)

func TestSesRndIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesRndCfgDIR = "sessions_internal"
	case utils.MetaMySQL:
		sesRndCfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		sesRndCfgDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSesRndIt {
		t.Run(sesRndCfgDIR, stest)
	}
}

func testSesRndItPreparePostpaidUP(t *testing.T) {
	sesRndCgrEv.Event[utils.Subject] = "up"
	sesRndCgrEv.Event[utils.RequestType] = utils.MetaPostpaid
	sesRndCgrEv.Event[utils.OriginID] = "RndupMetaPostpaid"
	sesRndExpMaxUsage = 10 * time.Second
	sesRndExpCost = 0.4
	sesRndExpBalanceValue = 3599999999999.5977
}

func testSesRndItPreparePostpaidDOWN(t *testing.T) {
	sesRndCgrEv.Event[utils.Subject] = "down"
	sesRndCgrEv.Event[utils.RequestType] = utils.MetaPostpaid
	sesRndCgrEv.Event[utils.OriginID] = "RnddownMetaPostpaid"
	sesRndExpMaxUsage = 10 * time.Second
	sesRndExpCost = 0.3
	sesRndExpBalanceValue = 3599999999999.697
}

func testSesRndItPreparePrepaidUP(t *testing.T) {
	sesRndCgrEv.Event[utils.Subject] = "up"
	sesRndCgrEv.Event[utils.RequestType] = utils.MetaPrepaid
	sesRndCgrEv.Event[utils.OriginID] = "RndupMetaPrepaid"
	sesRndExpMaxUsage = 3 * time.Hour
	sesRndExpCost = 0.4
	sesRndExpBalanceValue = 3599999999999.5977
}

func testSesRndItPreparePrepaidDOWN(t *testing.T) {
	sesRndCgrEv.Event[utils.Subject] = "down"
	sesRndCgrEv.Event[utils.RequestType] = utils.MetaPrepaid
	sesRndCgrEv.Event[utils.OriginID] = "RnddownMetaPrepaid"
	sesRndExpMaxUsage = 3 * time.Hour
	sesRndExpCost = 0.3
	sesRndExpBalanceValue = 3599999999999.697
}

// test for 0 balance with sesRndsion terminate with 1s usage
func testSesRndItLoadConfig(t *testing.T) {
	var err error
	sesRndCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesRndCfgDIR)
	if sesRndCfg, err = config.NewCGRConfigFromPath(sesRndCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesRndItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesRndCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesRndItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesRndCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesRndItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesRndCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesRndItRPCConn(t *testing.T) {
	sesRndRPC = engine.NewRPCClient(t, sesRndCfg.ListenCfg())
}

func testSesRndItLoadRating(t *testing.T) {
	var reply string
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetTPRate, &utils.TPRateRALs{
		TPid: utils.TestSQL,
		ID:   "RT1",
		RateSlots: []*utils.RateSlot{
			{ConnectFee: 0, Rate: 0.033, RateUnit: "1s", RateIncrement: "1s", GroupIntervalStart: "0s"},
		},
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRate: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRate: ", reply)
	}

	dr := &utils.TPDestinationRate{
		TPid: utils.TestSQL,
		ID:   "DR_UP",
		DestinationRates: []*utils.DestinationRate{
			{DestinationId: utils.MetaAny, RateId: "RT1", RoundingMethod: utils.MetaRoundingUp, RoundingDecimals: 1},
		},
	}
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetTPDestinationRate, dr, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPDestinationRate: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPDestinationRate: ", reply)
	}
	dr.ID = "DR_DOWN"
	dr.DestinationRates[0].RoundingMethod = utils.MetaRoundingDown
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetTPDestinationRate, dr, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPDestinationRate: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPDestinationRate: ", reply)
	}

	rp := &utils.TPRatingPlan{
		TPid: utils.TestSQL,
		ID:   "RP_UP",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{DestinationRatesId: "DR_UP", TimingId: utils.MetaAny, Weight: 10},
		},
	}
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetTPRatingPlan, rp, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRatingPlan: ", reply)
	}
	rp.ID = "RP_DOWN"
	rp.RatingPlanBindings[0].DestinationRatesId = "DR_DOWN"
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetTPRatingPlan, rp, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRatingPlan: ", reply)
	}

	rpf := &utils.TPRatingProfile{
		TPid:     utils.TestSQL,
		LoadId:   utils.TestSQL,
		Tenant:   sesRndTenant,
		Category: utils.Call,
		Subject:  "up",
		RatingPlanActivations: []*utils.TPRatingActivation{{
			RatingPlanId:     "RP_UP",
			FallbackSubjects: utils.EmptyString,
		}},
	}
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetTPRatingProfile, rpf, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRatingProfile: ", reply)
	}
	rpf.Subject = "down"
	rpf.RatingPlanActivations[0].RatingPlanId = "RP_DOWN"
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetTPRatingProfile, rpf, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetTPRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received when calling APIerSv1.SetTPRatingProfile: ", reply)
	}

	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1LoadRatingPlan, &v1.AttrLoadRatingPlan{TPid: utils.TestSQL}, &reply); err != nil {
		t.Error("Got error on APIerSv1.LoadRatingPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.LoadRatingPlan got reply: ", reply)
	}

	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1LoadRatingProfile, &utils.TPRatingProfile{
		TPid: utils.TestSQL, LoadId: utils.TestSQL,
		Tenant: sesRndTenant, Category: utils.Call}, &reply); err != nil {
		t.Error("Got error on APIerSv1.LoadRatingProfile: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.LoadRatingProfile got reply: ", reply)
	}

}

func testSesRndItAddCharger(t *testing.T) {
	//add a default charger
	var result string
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, &engine.ChargerProfile{
		Tenant:       sesRndTenant,
		ID:           "default",
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{utils.MetaNone},
		Weight:       20,
	}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testSesRndItAddVoiceBalance(t *testing.T) {
	var reply string
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      sesRndTenant,
		Account:     sesRndAccount,
		BalanceType: utils.MetaMonetary,
		Value:       float64(time.Hour),
		Balance: map[string]any{
			utils.ID: "TestSesBal1",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}

	var acnt engine.Account
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  sesRndTenant,
			Account: sesRndAccount,
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	expected := float64(time.Hour)
	if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != expected {
		t.Errorf("Expected: %v, received: %v", expected, rply)
	}
}

func testSesRndItPrepareCDRs(t *testing.T) {
	var reply sessions.V1InitSessionReply
	if err := sesRndRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent:    sesRndCgrEv,
		}, &reply); err != nil {
		t.Error(err)
		return
	} else if *reply.MaxUsage != sesRndExpMaxUsage {
		t.Errorf("Unexpected MaxUsage: %v", reply.MaxUsage)
	}
	time.Sleep(50 * time.Millisecond)

	var rply string
	if err := sesRndRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent:         sesRndCgrEv,
		}, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}

	if err := sesRndRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		sesRndCgrEv, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Received reply: %s", rply)
	}
	time.Sleep(20 * time.Millisecond)
}

func testSesRndItCheckCdrs(t *testing.T) {
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{Accounts: []string{sesRndAccount}, OriginIDs: []string{utils.IfaceAsString(sesRndCgrEv.Event[utils.OriginID])}}
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Fatal("Wrong number of CDRs")
	} else if cd, err := engine.IfaceAsEventCost(cdrs[0].CostDetails); err != nil {
		t.Fatal(err)
	} else if cd.Cost == nil ||
		*cd.Cost != sesRndExpCost ||
		*cd.Cost != cdrs[0].Cost {
		t.Errorf("CDR cost= %v", utils.ToJSON(cdrs[0].Cost))
		t.Errorf("CostDetails cost= %v", utils.ToJSON(cd.Cost))
		t.Errorf("Expected cost=%v", utils.ToJSON(sesRndExpCost))
		t.Log(cdrs[0].CostDetails)
	} else if len(cd.AccountSummary.BalanceSummaries) != 1 ||
		cd.AccountSummary.BalanceSummaries[0].Value != sesRndExpBalanceValue {
		t.Errorf("Unexpected AccountSummary: %v", utils.ToJSON(cd.AccountSummary))
	}
	var acnt engine.Account
	if err := sesRndRPC.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{
			Tenant:  sesRndTenant,
			Account: sesRndAccount,
		}, &acnt); err != nil {
		t.Fatal(err)
	}
	if rply := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rply != sesRndExpBalanceValue {
		t.Errorf("Expected: %+v, received: %v", utils.ToJSON(sesRndExpBalanceValue), utils.ToJSON(rply))
	}
}

func testSesRndItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
