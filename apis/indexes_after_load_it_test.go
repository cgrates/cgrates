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
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	idxLoadCfgPath   string
	idxLoadCfg       *config.CGRConfig
	idxLoadBiRPC     *birpc.Client
	idxLoadConfigDIR string //run tests for specific configuration

	idxLoadAccPrf = []func(t *testing.T){
		testIdxLoadInitCfg,
		testIdxLoadInitDataDb,
		testIdxLoadResetStorDb,
		testIdxLoadStartEngine,
		testIdxLoadRPCConn,
		testIdxLoadTariffPlan,
		// testIdxLoadCheckRateProfileIndexes,
		testIdxLoadKillEngine,
	}
)

func TestIdxCheckAfterLoad(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		idxLoadConfigDIR = "session_volume_discount_internal"
	case utils.MetaMySQL:
		idxLoadConfigDIR = "session_volume_discount_mysql"
	case utils.MetaMongo:
		idxLoadConfigDIR = "session_volume_discount_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range idxLoadAccPrf {
		t.Run(idxLoadConfigDIR, stest)
	}
}

func testIdxLoadInitCfg(t *testing.T) {
	var err error
	idxLoadCfgPath = path.Join(*dataDir, "conf", "samples", idxLoadConfigDIR)
	idxLoadCfg, err = config.NewCGRConfigFromPath(context.Background(), idxLoadCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testIdxLoadInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(idxLoadCfg); err != nil {
		t.Fatal(err)
	}
}

func testIdxLoadResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(idxLoadCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testIdxLoadStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(idxLoadCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testIdxLoadRPCConn(t *testing.T) {
	var err error
	idxLoadBiRPC, err = newRPCClient(idxLoadCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testIdxLoadTariffPlan(t *testing.T) {
	var reply string
	if err := idxLoadBiRPC.Call(context.Background(), utils.LoaderSv1Load,
		&loaders.ArgsProcessFolder{
			// StopOnError: true,
			Caching: utils.StringPointer(utils.MetaReload),
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testIdxLoadCheckRateProfileIndexes(t *testing.T) {
	//get indexes *rate_profiles
	var reply []string
	if err := idxLoadBiRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			ItemType: utils.MetaRateProfiles,
		}, &reply); err != nil {
		t.Error(err)
	}

	//get indexes
	if err := idxLoadBiRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			ItemType: utils.MetaChargers,
		}, &reply); err != nil {
		t.Error(err)
	} else {
		t.Error(reply)
	}

	//get indexes
	if err := idxLoadBiRPC.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			ItemType: utils.MetaAttributes,
		}, &reply); err != nil {
		t.Error(err)
	} else {
		t.Error(reply)
	}
}

//Kill the engine when it is about to be finished
func testIdxLoadKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
