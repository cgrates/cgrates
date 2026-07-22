//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
)

var (
	cdrsIntCfgPath string
	cdrsIntCfgDIR  string
	cdrsIntCfg     *config.CGRConfig
	cdrsIntRPC     *birpc.Client

	sTestsCdrsInt = []func(t *testing.T){
		testCdrsIntInitCfg,
		testCdrsIntStartEngine,
		testCdrsIntRpcConn,
		testCdrsIntTestTTL,
		testCdrsIntStopEngine,
	}
)

// This test is valid only for internal
// to test the ttl for cdrs
func TestCdrsIntIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		cdrsIntCfgDIR = "internal_ttl_internal"
	case utils.MetaMySQL:
		t.SkipNow()
	case utils.MetaMongo:
		t.SkipNow()
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCdrsInt {
		t.Run(cdrsIntCfgDIR, stest)
	}
}

func testCdrsIntInitCfg(t *testing.T) {
	var err error
	cdrsIntCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsIntCfgDIR)
	cdrsIntCfg, err = config.NewCGRConfigFromPath(cdrsIntCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCdrsIntStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsIntCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testCdrsIntRpcConn(t *testing.T) {
	cdrsIntRPC = engine.NewRPCClient(t, cdrsIntCfg.ListenCfg())
}

func testCdrsIntTestTTL(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{"*store:true"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testCdrsIntTestTTL",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testCdrsIntTestTTL",
				utils.RequestType:  utils.MetaNone,
				utils.Category:     "call",
				utils.AccountField: "testCdrsIntTestTTL",
				utils.Subject:      "ANY2CNT2",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
			},
		},
	}

	var reply string
	if err := cdrsIntRPC.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.ExternalCDR
	if err := cdrsIntRPC.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{}, &cdrs); err != nil {
		t.Fatal("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Errorf("Expected 1 result received %v ", len(cdrs))
	}
	time.Sleep(time.Second + 50*time.Millisecond)
	if err := cdrsIntRPC.Call(context.Background(), utils.APIerSv2GetCDRs, &utils.RPCCDRsFilter{}, &cdrs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal("Unexpected error: ", err)
	}
}

func testCdrsIntStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
