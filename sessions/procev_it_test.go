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
