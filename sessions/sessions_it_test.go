//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sItCfgPath string
	sItCfgDIR  string
	sItCfg     *config.CGRConfig
	sItRPC     *birpc.Client

	sessionsITtests = []func(t *testing.T){
		testSessionsItInitCfg,
		testSessionsItFlushDBs,
		/*
			testSessionsItStartEngine,
			testSessionsItApierRpcConn,
			testSessionsItTPFromFolder,
			testSessionsItTerminatNonexist,
			testSessionsItUpdateNonexist,
			testSessionsItTerminatePassive,
			testSessionsItEventCostCompressing,

			testSessionsItStopCgrEngine,
		*/
	}
)

func TestSessionsIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sItCfgDIR = "sessions_internal"
	case utils.MetaRedis:
		t.SkipNow()
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
	sItCfgPath = path.Join(*utils.DataDir, "conf", "samples", sItCfgDIR)
	var err error
	sItCfg, err = config.NewCGRConfigFromPath(context.Background(), sItCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSessionsItFlushDBs(t *testing.T) {
	if err := engine.InitDB(sItCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionsItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sItCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionsItApierRpcConn(t *testing.T) {
	sItRPC = engine.NewRPCClient(t, sItCfg.ListenCfg(), *utils.Encoding)
}

/*

// Load the tariff plan, creating accounts and their balances
func testSessionsItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := sItRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}


func testSessionsItEventCostCompressing(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "TestSessionsItEventCostCompressing",
		Value:       float64(5) * float64(time.Second),
		BalanceType: utils.MetaVoice,
		Balance: map[string]any{
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
			Event: map[string]any{
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
			Event: map[string]any{
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
			Event: map[string]any{
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
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
