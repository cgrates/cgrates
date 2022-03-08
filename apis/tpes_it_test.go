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

package apis

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
	tpesCfgPath   string
	tpesCfg       *config.CGRConfig
	tpeSRPC       *birpc.Client
	tpeSConfigDIR string //run tests for specific configuration

	sTestTpes = []func(t *testing.T){
		testTpeSInitCfg,
		testTpeSInitDataDb,

		testTpeSStartEngine,
		testTpeSRPCConn,
		testTpeSPing,
		testTpeSKillEngine,
	}
)

func TestTpeSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpeSConfigDIR = "tutinternal"
	case utils.MetaMongo:
		tpeSConfigDIR = "tutmongo"
	case utils.MetaMySQL:
		tpeSConfigDIR = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestTpes {
		t.Run(tpeSConfigDIR, stest)
	}
}

func testTpeSInitCfg(t *testing.T) {
	var err error
	tpesCfgPath = path.Join(*dataDir, "conf", "samples", tpeSConfigDIR)
	tpesCfg, err = config.NewCGRConfigFromPath(context.Background(), tpesCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testTpeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(tpesCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTpeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpesCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testTpeSPing(t *testing.T) {
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.TpeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
}

func testTpeSRPCConn(t *testing.T) {
	var err error
	tpeSRPC, err = newRPCClient(tpesCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testTpeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
