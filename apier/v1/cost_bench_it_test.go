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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	costBenchCfgPath   string
	costBenchCfg       *config.CGRConfig
	costBenchRPC       *rpc.Client
	costBenchConfigDIR string //run tests for specific configuration
)

func testCostBenchInitCfg(b *testing.B) {
	var err error
	costBenchCfgPath = path.Join(costDataDir, "conf", "samples", costBenchConfigDIR)
	costBenchCfg, err = config.NewCGRConfigFromPath(costBenchCfgPath)
	if err != nil {
		b.Error(err)
	}
}

func testCostBenchInitDataDb(b *testing.B) {
	if err := engine.InitDataDb(costBenchCfg); err != nil {
		b.Fatal(err)
	}
}

// Wipe out the cdr database
func testCostBenchResetStorDb(b *testing.B) {
	if err := engine.InitStorDb(costBenchCfg); err != nil {
		b.Fatal(err)
	}
}

// Start CGR Engine
func testCostBenchStartEngine(b *testing.B) {
	if _, err := engine.StopStartEngine(costBenchCfgPath, 500); err != nil {
		b.Fatal(err)
	}
}

// Connect rpc client to rater
func testCostBenchRPCConn(b *testing.B) {
	var err error
	costBenchRPC, err = newRPCClient(costBenchCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		b.Fatal(err)
	}
}

func testCostBenchLoadFromFolder(b *testing.B) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := costBenchRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		b.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testCostBenchSetRateProfile(b *testing.B) {
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				FixedFee:      0.4,
				RecurrentFee:  0.2,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.1,
				Unit:          time.Minute,
				Increment:     time.Second,
			},
		},
	}
	rPrf := &RateProfileWithCache{
		RateProfileWithOpts: &engine.RateProfileWithOpts{
			RateProfile: &engine.RateProfile{
				ID:        "DefaultRate",
				FilterIDs: []string{"*string:~*req.Subject:1001"},
				Weight:    10,
				Rates: map[string]*engine.Rate{
					"RATE1": rate1,
				},
			},
		},
	}
	var reply string
	if err := costBenchRPC.Call(utils.APIerSv1SetRateProfile, rPrf, &reply); err != nil {
		b.Error(err)
	} else if reply != utils.OK {
		b.Error("Unexpected reply returned", reply)
	}
}

func BenchmarkCostWithRALs(b *testing.B) {
	costBenchConfigDIR = "tutinternal"
	testCostBenchInitCfg(b)
	testCostBenchInitDataDb(b)
	testCostBenchResetStorDb(b)
	testCostBenchStartEngine(b)
	testCostBenchRPCConn(b)
	testCostBenchLoadFromFolder(b)

	tNow := time.Now()
	cd := &engine.CallDescriptorWithOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			Account:       "1001",
			Destination:   "1002",
			DurationIndex: 120000000000,
			TimeStart:     tNow,
			TimeEnd:       tNow.Add(120000000000),
		},
	}
	var cc engine.CallCost
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := costBenchRPC.Call(utils.ResponderGetCost, cd, &cc); err != nil {
			b.Error("Got error on Responder.GetCost: ", err.Error())
		}
	}
	b.StopTimer()
	testCostBenchKillEngine(b)
}

func BenchmarkCostWithRateS(b *testing.B) {
	costBenchConfigDIR = "tutinternal"
	testCostBenchInitCfg(b)
	testCostBenchInitDataDb(b)
	testCostBenchResetStorDb(b)
	testCostBenchStartEngine(b)
	testCostBenchRPCConn(b)
	testCostBenchSetRateProfile(b)
	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "2m",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := costBenchRPC.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
	testCostBenchKillEngine(b)
}

func testCostBenchKillEngine(b *testing.B) {
	if err := engine.KillEngine(500); err != nil {
		b.Error(err)
	}
}
