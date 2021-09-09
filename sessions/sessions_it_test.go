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
package sessions

import (
	"net/rpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sItCfgPath string
	sItCfgDIR  string
	sItCfg     *config.CGRConfig
	sItRPC     *rpc.Client

	sessionsITtests = []func(t *testing.T){
		testSessionsItInitCfg,
		testSessionsItResetDataDb,
		testSessionsItResetStorDb,
		/*
			testSessionsItStartEngine,
			testSessionsItApierRpcConn,
			testSessionsItTPFromFolder,
			testSessionsItTerminatNonexist,
			testSessionsItUpdateNonexist,
			testSessionsItTerminatePassive,
			testSessionsItEventCostCompressing,

		*/
		testSessionsItStopCgrEngine,
	}
)

func TestSessionsIt(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		sItCfgDIR = "sessions_internal"
	case utils.MetaMySQL:
		sItCfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		sItCfgDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sessionsITtests {
		t.Run(sItCfgDIR, stest)
	}
}

// Init config firs
func testSessionsItInitCfg(t *testing.T) {
	sItCfgPath = path.Join(*dataDir, "conf", "samples", sItCfgDIR)
	var err error
	sItCfg, err = config.NewCGRConfigFromPath(sItCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSessionsItResetDataDb(t *testing.T) {
	if err := engine.InitDataDB(sItCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSessionsItResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(sItCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionsItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sItCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionsItApierRpcConn(t *testing.T) {
	var err error
	sItRPC, err = newRPCClient(sItCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

/*

// Load the tariff plan, creating accounts and their balances
func testSessionsItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := sItRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSessionsItTerminatNonexist(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 10.0
	if err := sItRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage := 2 * time.Minute
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSessionsItTerminatNonexist",
			Event: map[string]interface{}{
				utils.EventName:    "TerminateEvent",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123451",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
			APIOpts: map[string]interface{}{
				utils.OptsDebitInterval: "10s",
			},
		},
	}

	var rpl string
	if err := sItRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)

	eAcntVal = 9.299800
	if err := sItRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
	time.Sleep(100 * time.Millisecond)
	if err := sItRPC.Call(utils.SessionSv1ProcessCDR, termArgs.CGREvent, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.OK {
		t.Errorf("Received reply: %s", rpl)
	}

	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{
		DestinationPrefixes: []string{"1002"},
		RunIDs:              []string{utils.MetaDefault},
	}
	if err := sItRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Unexpected number of CDRs returned: %v \n cdrs=%s", len(cdrs), utils.ToJSON(cdrs))
	} else {
		if cdrs[0].Usage != "2m0s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
		if cdrs[0].Cost != 0.7002 {
			t.Errorf("Unexpected CDR Cost received, cdr: %v %+v ", cdrs[0].Cost, cdrs[0])
		}
	}

}

func testSessionsItUpdateNonexist(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 9.299800
	if err := sItRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage := 2 * time.Minute
	updtArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSessionsItUpdateNonexist",
			Event: map[string]interface{}{
				utils.EventName:    "UpdateEvent",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123789",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
			APIOpts: map[string]interface{}{
				utils.OptsDebitInterval: "0s",
			},
		},
	}

	var updtRpl V1UpdateSessionReply
	if err := sItRPC.Call(utils.SessionSv1UpdateSession, updtArgs, &updtRpl); err != nil {
		t.Error(err)
	}
	if updtRpl.MaxUsage == nil || *updtRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, updtRpl.MaxUsage)
	}

	time.Sleep(10 * time.Millisecond)

	eAcntVal = 8.599600
	if err := sItRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
	var rpl string
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSessionsItTerminatNonexist",
			Event: map[string]interface{}{
				utils.EventName:    "TerminateEvent",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123789",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	if err := sItRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
}

func testSessionsItTerminatePassive(t *testing.T) {
	//create the event for session
	sEv := engine.NewMapEvent(map[string]interface{}{
		utils.EventName:    "UpdateEvent",
		utils.ToR:          utils.MetaVoice,
		utils.OriginID:     "123789",
		utils.AccountField: "1001",
		utils.Subject:      "1001",
		utils.Destination:  "1002",
		utils.Category:     "call",
		utils.Tenant:       "cgrates.org",
		utils.RequestType:  utils.MetaPrepaid,
		utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
		utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
		utils.Usage:        time.Minute,
	})

	cgrID := GetSetCGRID(sEv)
	s := &Session{
		CGRID:      cgrID,
		EventStart: sEv,
		SRuns: []*SRun{
			{
				Event:      sEv,
				TotalUsage: time.Minute,
				CD:         &engine.CallDescriptor{},
			},
		},
	}

	var rply string
	//transfer the session from active to pasive
	if err := sItRPC.Call(utils.SessionSv1SetPassiveSession,
		s, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, rply)
	}
	var pSessions []*ExternalSession
	//check if the passive session was created
	if err := sItRPC.Call(utils.SessionSv1GetPassiveSessions,
		&utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "123789"),
			},
		}, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", pSessions)
	} else if pSessions[0].Usage != time.Minute {
		t.Errorf("Expecting 1m, received usage: %v", pSessions[0].Usage)
	}

	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSessionsItTerminatNonexist",
			Event: map[string]interface{}{
				utils.EventName:    "TerminateEvent",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123789",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        time.Minute,
			},
		},
	}

	var rpl string
	if err := sItRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)

	//check if the passive session was terminate
	if err := sItRPC.Call(utils.SessionSv1GetPassiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "123789"),
			},
		}, &pSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(aSessions)=%v , session : %+v", err, len(pSessions), utils.ToJSON(pSessions))
	}

}

func testSessionsItEventCostCompressing(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "TestSessionsItEventCostCompressing",
		Value:       float64(5) * float64(time.Second),
		BalanceType: utils.MetaVoice,
		Balance: map[string]interface{}{
			utils.ID:            "TestSessionsItEventCostCompressing",
			utils.RatingSubject: "*zero50ms",
		},
	}
	var reply string
	if err := sItRPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	// Init the session
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsItEventCostCompressing",
			Event: map[string]interface{}{
				utils.OriginID:     "TestSessionsItEventCostCompressing",
				utils.AccountField: "TestSessionsItEventCostCompressing",
				utils.Destination:  "1002",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AnswerTime:   time.Date(2019, time.March, 1, 13, 57, 05, 0, time.UTC),
				utils.Usage:        "1s",
			},
		},
	}
	var initRpl *V1InitSessionReply
	if err := sItRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != time.Second {
		t.Errorf("received: %+v", initRpl.MaxUsage)
	}
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsItEventCostCompressing",
			Event: map[string]interface{}{
				utils.OriginID: "TestSessionsItEventCostCompressing",
				utils.Usage:    "1s",
			},
		},
	}
	var updateRpl *V1UpdateSessionReply
	if err := sItRPC.Call(utils.SessionSv1UpdateSession,
		updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if err := sItRPC.Call(utils.SessionSv1UpdateSession,
		updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if err := sItRPC.Call(utils.SessionSv1UpdateSession,
		updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsDataLastUsedData",
			Event: map[string]interface{}{
				utils.OriginID:     "TestSessionsItEventCostCompressing",
				utils.AccountField: "TestSessionsItEventCostCompressing",
				utils.Destination:  "1002",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AnswerTime:   time.Date(2019, time.March, 1, 13, 57, 05, 0, time.UTC),
				utils.Usage:        "4s",
			},
		},
	}
	var rpl string
	if err := sItRPC.Call(utils.SessionSv1TerminateSession,
		termArgs, &rpl); err != nil ||
		rpl != utils.OK {
		t.Error(err)
	}
	if err := sItRPC.Call(utils.SessionSv1ProcessCDR,
		termArgs.CGREvent, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(20 * time.Millisecond)
	cgrID := utils.Sha1("TestSessionsItEventCostCompressing", "")
	var ec *engine.EventCost
	if err := sItRPC.Call(utils.APIerSv1GetEventCost,
		&utils.AttrGetCallCost{CgrId: cgrID, RunId: utils.MetaDefault},
		&ec); err != nil {
		t.Fatal(err)
	}
	// make sure we only have one aggregated Charge
	if len(ec.Charges) != 1 ||
		ec.Charges[0].CompressFactor != 4 ||
		len(ec.Rating) != 1 ||
		len(ec.Accounting) != 1 ||
		len(ec.RatingFilters) != 1 ||
		len(ec.Rates) != 1 {
		t.Errorf("unexpected EC returned: %s", utils.ToIJSON(ec))
	}

}

*/

func testSessionsItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
