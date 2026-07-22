//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rpLateCfGPath   string
	rpLateCfg       *config.CGRConfig
	rpLateRPC       *birpc.Client
	rpLateConfigDIR string //run tests for specific configuration

	rpLateAPIer = []func(t *testing.T){
		testRpLateInitCfg,
		testRpLateInitDataDb,
		testRpLateResetStorDb,
		testRpLateStartEngine,
		testRpLateRPCConn,
		testRpLateLoadFromFolder,
		testRpLateCDRProcessEvent,
		testRpLateKillEngine,
	}
)

// Test start here
func TestRPLateIT2(t *testing.T) {
	// no need for a new config with *gob transport in this case
	switch *utils.DBType {
	case utils.MetaInternal:
		rpLateConfigDIR = "processcdrs_late_internal"
	case utils.MetaMySQL:
		rpLateConfigDIR = "processcdrs_late_mysql"
	case utils.MetaMongo:
		rpLateConfigDIR = "processcdrs_late_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range rpLateAPIer {
		t.Run(rpLateConfigDIR, stest)
	}
}

func testRpLateInitCfg(t *testing.T) {
	var err error
	rpLateCfGPath = path.Join(*utils.DataDir, "conf", "samples", rpLateConfigDIR)
	rpLateCfg, err = config.NewCGRConfigFromPath(rpLateCfGPath)
	if err != nil {
		t.Error(err)
	}
}

func testRpLateInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(rpLateCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testRpLateResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rpLateCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRpLateStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rpLateCfGPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testRpLateRPCConn(t *testing.T) {
	var err error
	rpLateRPC, err = newRPCClient(rpLateCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testRpLateLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := rpLateRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testRpLateCDRProcessEvent(t *testing.T) {
	args := &engine.CDRWithAPIOpts{
		CDR: &engine.CDR{
			CGRID:       "RandomCgrId",
			Tenant:      "cgrates.org",
			OriginID:    "testRpLateCDRProcessEvent",
			Source:      "testRpLateCDRProcessEvent",
			RequestType: utils.MetaPrepaid,
			Account:     "1001",
			Category:    "call",
			Subject:     "1001",
			Destination: "1001",
			SetupTime:   time.Date(2010, 8, 24, 16, 00, 26, 0, time.UTC),
			AnswerTime:  time.Date(2010, 8, 24, 16, 00, 26, 0, time.UTC),
			Usage:       2 * time.Minute,
		},
	}
	var reply string
	if err := rpLateRPC.Call(context.Background(), utils.CDRsV1ProcessCDR, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	var replyy []*engine.CDR
	req := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{
		/*
			CGRIDs: []string{"RandomCgrId"},
			//RunIDs: []string{utils.MetaRaw, utils.MetaDefault},
			Tenants: []string{"cgrates.org"},
			Categories: []string{"call"},
			Subjects: []string{"1001"},

		*/
	}}
	if err := rpLateRPC.Call(context.Background(), utils.CDRsV1GetCDRs, &req, &replyy); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(replyy) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(replyy), "and CDRS: ", utils.ToJSON(replyy))
	}
}

func testRpLateKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
