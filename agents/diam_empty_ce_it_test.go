//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	lstnr := daCfgND.DiameterAgentCfg().Listeners[0]
	diamClntND, err = NewDiameterClient(lstnr.Address,
		"INTEGRATION_TESTS",
		daCfgND.DiameterAgentCfg().OriginRealm, daCfgND.DiameterAgentCfg().VendorID,
		daCfgND.DiameterAgentCfg().ProductName, utils.DiameterFirmwareRevision,
		daCfgND.DiameterAgentCfg().DictionariesPath, daCfgND.DiameterAgentCfg().DictionariesAppendDefaults, lstnr.Network)
	if err.Error() != "missing application" {
		t.Fatal(err)
	}
}

func testDiamEmptyCEItKillEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
