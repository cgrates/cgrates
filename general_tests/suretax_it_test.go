//go:build suretax
// +build suretax

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

package general_tests

import (
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

/*
Integration tests with SureTax platform.
Configuration file is kept outside of CGRateS repository since it contains sensitive customer information
*/

var (
	configDir = flag.String("config_path", "", "CGR config dir path here")
	tpDir     = flag.String("tp_dir", "", "CGR config dir path here")

	stiCfg      *config.CGRConfig
	stiRpc      *rpc.Client
	stiLoadInst utils.LoadInstance

	sTestSTI = []func(t *testing.T){
		testSTIInitCfg,
		testSTIResetDataDb,
		testSTIResetStorDb,
		testSTIStartEngine,
		testSTIRpcConn,
		testSTILoadTariffPlanFromFolder,
		testSTICacheStats,
		testSTIProcessExternalCdr,
		testSTIGetCdrs,
		testSTIStopCgrEngine,
	}
)

func TestSTI(t *testing.T) {
	for _, stest := range sTestSTI {
		t.Run("TestSTI", stest)
	}
}

func testSTIInitCfg(t *testing.T) {
	// Init config first
	var err error
	stiCfg, err = config.NewCGRConfigFromPath(*configDir)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSTIResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(stiCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSTIResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(stiCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSTIStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(*configDir, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSTIRpcConn(t *testing.T) {
	var err error
	stiRpc, err = newRPCClient(stiCfg) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSTILoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: *tpDir}
	if err := stiRpc.Call(utils.APIerSv2LoadTariffPlanFromFolder, attrs, &stiLoadInst); err != nil {
		t.Error(err)
	} else if stiLoadInst.RatingLoadID == "" || stiLoadInst.AccountingLoadID == "" {
		t.Error("Empty loadId received, loadInstance: ", stiLoadInst)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func testSTICacheStats(t *testing.T) {
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 1, RatingPlans: 1, RatingProfiles: 1}
	var args utils.AttrCacheStats
	if err := stiRpc.Call(utils.APIerSv2GetCacheStats, args, &rcvStats); err != nil {
		t.Error("Got error on APIerSv2.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling APIerSv2.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
}

// Test CDR from external sources
func testSTIProcessExternalCdr(t *testing.T) {
	cdr := &engine.ExternalCDR{ToR: utils.VOICE,
		OriginID: "teststicdr1", OriginHost: "192.168.1.1", Source: "STI_TEST", RequestType: utils.META_RATED,
		Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "+14082342500", Destination: "+16268412300", Supplier: "SUPPL1",
		SetupTime: "2015-10-18T13:00:00Z", AnswerTime: "2015-10-18T13:00:00Z",
		Usage: "15s", PDD: "7.0", ExtraFields: map[string]string{"CustomerNumber": "000000534", "ZipCode": ""},
	}
	var reply string
	if err := stiRpc.Call(utils.CdrsV2ProcessExternalCdr, cdr, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(time.Duration(2) * time.Second)
}

func testSTIGetCdrs(t *testing.T) {
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}, Accounts: []string{"1001"}}
	if err := stiRpc.Call(utils.APIerSv2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.012 {
			t.Errorf("Unexpected Cost for CDR: %+v", cdrs[0])
		}
	}
	req = utils.RPCCDRsFilter{RunIDs: []string{utils.META_SURETAX}, Accounts: []string{"1001"}}
	if err := stiRpc.Call(utils.APIerSv2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Cost != 0.0027 {
			t.Errorf("Unexpected Cost for CDR: %+v", cdrs[0])
		}
	}
}

func testSTIStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
