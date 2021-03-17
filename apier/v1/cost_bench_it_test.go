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
	costBenchCfgPath = path.Join(*dataDir, "conf", "samples", costBenchConfigDIR)
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

func testCostBenchLoadFromFolder2(b *testing.B) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial2")}
	if err := costBenchRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		b.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testCostBenchSetRateProfile(b *testing.B) {
	rPrf := &engine.APIRateProfileWithOpts{
		APIRateProfile: &engine.APIRateProfile{
			ID:        "DefaultRate",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Weights:   ";10",
			Rates: map[string]*engine.APIRate{
				"RATE1": &engine.APIRate{
					ID:              "RATE1",
					Weights:         ";0",
					ActivationTimes: "* * * * *",
					IntervalRates: []*engine.APIIntervalRate{
						{
							IntervalStart: "0",
							FixedFee:      utils.Float64Pointer(0.4),
							RecurrentFee:  utils.Float64Pointer(0.2),
							Unit:          utils.Float64Pointer(60000000000),
							Increment:     utils.Float64Pointer(60000000000),
						},
						{
							IntervalStart: "1m",
							RecurrentFee:  utils.Float64Pointer(0.1),
							Unit:          utils.Float64Pointer(60000000000),
							Increment:     utils.Float64Pointer(1000000000),
						},
					},
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

func testCostBenchSetRateProfile2(b *testing.B) {
	rate1 := &engine.APIRate{
		ID:              "RATE1",
		Weights:         ";0",
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.APIIntervalRate{
			{
				IntervalStart: "0",
				RecurrentFee:  utils.Float64Pointer(0.2),
				Unit:          utils.Float64Pointer(60000000000),
				Increment:     utils.Float64Pointer(60000000000),
			},
			{
				IntervalStart: "1m",
				RecurrentFee:  utils.Float64Pointer(0.1),
				Unit:          utils.Float64Pointer(60000000000),
				Increment:     utils.Float64Pointer(1000000000),
			},
		},
	}
	rtChristmas := &engine.APIRate{
		ID:              "RT_CHRISTMAS",
		Weights:         ";30",
		ActivationTimes: "* * 24 12 *",
		IntervalRates: []*engine.APIIntervalRate{{
			IntervalStart: "0",
			RecurrentFee:  utils.Float64Pointer(0.6),
			Unit:          utils.Float64Pointer(60000000000),
			Increment:     utils.Float64Pointer(1000000000),
		}},
	}
	rPrf := &engine.APIRateProfileWithOpts{
		APIRateProfile: &engine.APIRateProfile{
			ID:        "RateChristmas",
			FilterIDs: []string{"*string:~*req.Subject:1010"},
			Weights:   ";50",
			Rates: map[string]*engine.APIRate{
				"RATE1":          rate1,
				"RATE_CHRISTMAS": rtChristmas,
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

// go test -run=^$ -tags=integration -v -bench=BenchmarkCostWithRALs -benchtime=5s
func BenchmarkCostWithRALs(b *testing.B) {
	costBenchConfigDIR = "tutinternal"
	testCostBenchInitCfg(b)
	testCostBenchInitDataDb(b)
	testCostBenchResetStorDb(b)
	testCostBenchStartEngine(b)
	testCostBenchRPCConn(b)
	testCostBenchLoadFromFolder(b)

	tNow := time.Now()
	cd := &engine.CallDescriptorWithAPIOpts{
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

// go test -run=^$ -tags=integration -v -bench=BenchmarkCostDiffPeriodWithRALs -benchtime=5s
func BenchmarkCostDiffPeriodWithRALs(b *testing.B) {
	costBenchConfigDIR = "tutinternal"
	testCostBenchInitCfg(b)
	testCostBenchInitDataDb(b)
	testCostBenchResetStorDb(b)
	testCostBenchStartEngine(b)
	testCostBenchRPCConn(b)
	testCostBenchLoadFromFolder2(b)

	tStart, _ := utils.ParseTimeDetectLayout("2020-12-09T07:00:00Z", utils.EmptyString)
	tEnd, _ := utils.ParseTimeDetectLayout("2020-12-09T09:00:00Z", utils.EmptyString)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1010",
			Destination:   "1012",
			DurationIndex: tEnd.Sub(tStart),
			TimeStart:     tStart,
			TimeEnd:       tEnd,
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

// go test -run=^$ -tags=integration -v -bench=BenchmarkCostWithRateS -benchtime=5s
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
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "2m",
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

// go test -run=^$ -tags=integration -v -bench=BenchmarkCostDiffPeriodWithRateS -benchtime=5s
func BenchmarkCostDiffPeriodWithRateS(b *testing.B) {
	costBenchConfigDIR = "tutinternal"
	testCostBenchInitCfg(b)
	testCostBenchInitDataDb(b)
	testCostBenchResetStorDb(b)
	testCostBenchStartEngine(b)
	testCostBenchRPCConn(b)
	testCostBenchSetRateProfile2(b)
	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1010",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2020, 12, 23, 59, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:     "2h",
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
