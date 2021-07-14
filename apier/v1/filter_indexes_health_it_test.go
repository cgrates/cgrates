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
MERCHANTABILITY or FIdxTNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tFIdxHRpc *rpc.Client

	sTestsFilterIndexesSHealth = []func(t *testing.T){
		testV1FIdxHLoadConfig,
		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,
		testV1FIdxHStartEngine,
		testV1FIdxHRpcConn,
		testV1FIdxHLoadFromFolderTutorial2,
		testV1FIdxHAccountActionPlansHealth,
		testV1FIdxHReverseDestinationHealth,
		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,
		testV1FIdxHLoadFromFolderTutorial,
		testV1FIdxGetThresholdsIndexesHealth,

		testV1FIdxHStopEngine,
	}
)

// Test start here
func TestFIdxHealthIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		tSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		tSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFilterIndexesSHealth {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxHLoadConfig(t *testing.T) {
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	var err error
	if tSv1Cfg, err = config.NewCGRConfigFromPath(tSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1FIdxHdxInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxHResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHRpcConn(t *testing.T) {
	var err error
	tFIdxHRpc, err = newRPCClient(tSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FIdxHLoadFromFolderTutorial2(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial2")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxHAccountActionPlansHealth(t *testing.T) {
	var reply engine.AccountActionPlanIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetAccountActionPlansIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxHReverseDestinationHealth(t *testing.T) {
	var reply engine.ReverseDestinationsIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetReverseDestinationsIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxHLoadFromFolderTutorial(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxGetThresholdsIndexesHealth(t *testing.T) {
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    tenant,
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"*string:~*req.Account:1004",
				"*prefix:~*opts.Destination:+442|+554"},
			MaxHits:   1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
	}

	var rplyok string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &rplyok); err != nil {
		t.Error(err)
	} else if rplyok != utils.OK {
		t.Error("Unexpected reply returned", rplyok)
	}

	expiIdx := []string{
		"*string:*req.Account:1002:THD_ACNT_1002",
		"*string:*req.Account:1001:THD_ACNT_1001",
		"*string:*req.Account:1004:TEST_PROFILE1",
		"*prefix:*opts.Destination:+442:TEST_PROFILE1",
		"*prefix:*opts.Destination:+554:TEST_PROFILE1",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds,
		Tenant: "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expiIdx)
		if !reflect.DeepEqual(expiIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expiIdx, result)
		}
	}

	args := &engine.IndexHealthArgsWith3Ch{}
	expRPly := &engine.FilterIHReply{
		MissingIndexes: map[string][]string{
			"cgrates.org:*prefix:*opts.Destination:+442": {"TEST_PROFILE1"},
			"cgrates.org:*prefix:*opts.Destination:+554": {"TEST_PROFILE1"},
			"cgrates.org:*string:*req.Account:1001": {"THD_ACNT_1001"},
			"cgrates.org:*string:*req.Account:1002": {"THD_ACNT_1002"},
			"cgrates.org:*string:*req.Account:1004": {"TEST_PROFILE1"},
		},
		BrokenIndexes: map[string][]string{},
		MissingFilters: map[string][]string{},
	}
		var rply *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetThresholdsIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}
}

func testV1FIdxHStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
