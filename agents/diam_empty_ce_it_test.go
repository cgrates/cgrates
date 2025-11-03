//go:build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package agents

import (
	"path"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	daCfgPathND, diamConfigDIRND string
	daCfgND                      *config.CGRConfig
	diamClntND                   *DiameterClient

	sTestsDiamND = []func(t *testing.T){
		testDiamEmptyCEItInitCfg,
		testDiamEmptyCEItDataDb,
		testDiamEmptyCEItStartEngine,
		testDiamEmptyCEItConnectDiameterClient,
		testDiamEmptyCEItKillEngine,
	}
)

// Test start here
func TestDiamEmptyCEItTcp(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		diamConfigDIRND = "diamagent_internal_empty_apps"
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsDiamND {
		t.Run(diamConfigDIRND, stest)
	}
}

func testDiamEmptyCEItInitCfg(t *testing.T) {
	daCfgPathND = path.Join(*utils.DataDir, "conf", "samples", diamConfigDIRND)
	// Init config first
	var err error
	daCfgND, err = config.NewCGRConfigFromPath(context.Background(), daCfgPathND)
	if err != nil {
		t.Fatal(err)
	}
	daCfgND.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

// Remove data in both rating and accounting db
func testDiamEmptyCEItDataDb(t *testing.T) {
	if err := engine.InitDB(daCfgND); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDiamEmptyCEItStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(daCfgPathND, 500); err != nil {
		t.Fatal(err)
	}
}

func testDiamEmptyCEItConnectDiameterClient(t *testing.T) {
	var err error
	diamClntND, err = NewDiameterClient(daCfgND.DiameterAgentCfg().Listen,
		"INTEGRATION_TESTS",
		daCfgND.DiameterAgentCfg().OriginRealm, daCfgND.DiameterAgentCfg().VendorID,
		daCfgND.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		daCfgND.DiameterAgentCfg().DictionariesPath, daCfgND.DiameterAgentCfg().ListenNet)
	if err.Error() != "missing application" {
		t.Fatal(err)
	}
}

func testDiamEmptyCEItKillEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
