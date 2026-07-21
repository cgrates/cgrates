//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"os"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sProcEvCfgPath string
	sProcEvCfgDIR  string
	sProcEvCfg     *config.CGRConfig
	sProcEvRPC     *birpc.Client

	SessionsBkupTests = []func(t *testing.T){
		testSessionSProcEvInitCfg,
		testSessionSProcEvResetDB,
		testSessionSProcEvStartEngine,
		testSessionSProcEvApierRpcConn,

		testSessionSProcEvStopCgrEngine,
	}
)

func TestSessionSProcEv(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sProcEvCfgDIR = "sessions_procev_internal"
		defer func() {
			if err := os.RemoveAll("/tmp/internal_db"); err != nil {
				t.Error(err)
			}
		}()
	case utils.MetaRedis, utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres:
		return
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range SessionsBkupTests {
		t.Run(*utils.DBType, stest)
	}
}

func testSessionSProcEvInitCfg(t *testing.T) {
	var err error
	sProcEvCfgPath = path.Join(*utils.DataDir, "conf", "samples", sProcEvCfgDIR)
	if sProcEvCfg, err = config.NewCGRConfigFromPath(context.Background(), sProcEvCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSessionSProcEvResetDB(t *testing.T) {
	engine.FlushDBs(t, sProcEvCfg, true)
}

// Start CGR Engine
func testSessionSProcEvStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(sProcEvCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionSProcEvApierRpcConn(t *testing.T) {
	sProcEvRPC = engine.NewRPCClient(t, sProcEvCfg.ListenCfg(), *utils.Encoding)
}

func testSessionSProcEvStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
