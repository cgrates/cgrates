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

package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

var (
	engine1Cfg     *config.CGRConfig
	engine1RPC     *rpc.Client
	engine1CfgPath string
	engine2Cfg     *config.CGRConfig
	engine2RPC     *rpc.Client
	engine2CfgPath string

	sTestsCacheSReplicate = []func(t *testing.T){
		testCacheSReplicateLoadConfig,
		testCacheSReplicateInitDataDb,
		testCacheSReplicateInitStorDb,
		testCacheSReplicateStartEngine,
		testCacheSReplicateRpcConn,
		testCacheSReplicateLoadTariffPlanFromFolder,
		testCacheSReplicateProcessAttributes,
		testCacheSReplicateProcessRateProfile,
		testCacheSReplicateStopEngine,
	}
)

func TestCacheSv1ReplicateIT(t *testing.T) {
	for _, stest := range sTestsCacheSReplicate {
		t.Run("TestCacheSv1ReplicateIT", stest)
	}
}

func testCacheSReplicateLoadConfig(t *testing.T) {
	var err error
	engine1CfgPath = path.Join(*dataDir, "conf", "samples", "replication_cache", "engine1")
	if engine1Cfg, err = config.NewCGRConfigFromPath(engine1CfgPath); err != nil {
		t.Error(err)
	}
	engine2CfgPath = path.Join(*dataDir, "conf", "samples", "replication_cache", "engine2")
	if engine2Cfg, err = config.NewCGRConfigFromPath(engine2CfgPath); err != nil {
		t.Error(err)
	}
}

func testCacheSReplicateInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(engine1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(engine2Cfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func testCacheSReplicateInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(engine1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(engine2Cfg); err != nil {
		t.Fatal(err)
	}
}

// Start engine
func testCacheSReplicateStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(engine1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(engine2CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCacheSReplicateRpcConn(t *testing.T) {
	var err error
	engine1RPC, err = newRPCClient(engine1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to RPC: ", err.Error())
	}
	engine2RPC, err = newRPCClient(engine2Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to RPC: ", err.Error())
	}
}

func testCacheSReplicateLoadTariffPlanFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := engine2RPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testCacheSReplicateProcessAttributes(t *testing.T) {
	ev := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.MetaSessionS),
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testCacheSReplicateProcessAttributes",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_ACNT_1001"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "OfficeGroup"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testCacheSReplicateProcessAttributes",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				"OfficeGroup":      "Marketing",
			},
			Opts: map[string]interface{}{},
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := engine1RPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	} else {
		sort.Strings(eRply.AlteredFields)
		sort.Strings(rplyEv.AlteredFields)
		if !reflect.DeepEqual(eRply, &rplyEv) { // second for reversed order of attributes
			t.Errorf("Expecting: %s, received: %s",
				utils.ToJSON(eRply), utils.ToJSON(rplyEv))
		}
	}
	if err := engine2RPC.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	} else {
		sort.Strings(eRply.AlteredFields)
		sort.Strings(rplyEv.AlteredFields)
		if !reflect.DeepEqual(eRply, &rplyEv) { // second for reversed order of attributes
			t.Errorf("Expecting: %s, received: %s",
				utils.ToJSON(eRply), utils.ToJSON(rplyEv))
		}
	}
}

func testCacheSReplicateProcessRateProfile(t *testing.T) {
	var rply *utils.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.AccountField: "1002",
			},
		},
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rate1 := &utils.Rate{
		ID: "RT_ALWAYS",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		ActivationTimes: "* * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				FixedFee:      utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(1, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	exp := &utils.RateProfileCost{
		ID:   "RT_SPECIAL_1002",
		Cost: 0.01,
		RateSIntervals: []*utils.RateSInterval{{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{{
				IncrementStart:    utils.NewDecimal(0, 0),
				Usage:             utils.NewDecimal(int64(time.Minute), 0),
				Rate:              rate1,
				IntervalRateIndex: 0,
				CompressFactor:    60,
			}},
			CompressFactor: 1,
		}},
	}
	if err := engine1RPC.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
	if err := engine2RPC.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))

	}
}

func testCacheSReplicateStopEngine(t *testing.T) {
	if err := engine.KillEngine(300); err != nil {
		t.Error(err)
	}
}
