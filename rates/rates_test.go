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

package rates

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestListenAndServe(t *testing.T) {
	newRates := &RateS{}
	cfgRld := make(chan struct{}, 1)
	stopChan := make(chan struct{}, 1)
	cfgRld <- struct{}{}
	go func() {
		time.Sleep(10 * time.Nanosecond)
		stopChan <- struct{}{}
	}()
	newRates.ListenAndServe(stopChan, cfgRld)
}

func TestNewRateS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dataManager := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dataManager)
	expected := &RateS{
		cfg:   cfg,
		fltrS: filters,
		dm:    dataManager,
	}
	if newRateS := NewRateS(cfg, filters, dataManager); !reflect.DeepEqual(newRateS, expected) {
		t.Errorf("Expected %+v, received %+v", expected, newRateS)
	}
}

func TestRateProfileCostForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}

	//MatchItmID before setting
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}

	expectedRPCost := &utils.RateProfileCost{
		ID:   "RATE_1",
		Cost: utils.NewDecimal(2, 1),
		CostIntervals: []*utils.RateSIntervalCost{
			{
				Increments: []*utils.RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"RATE1": {
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}

	if rcv, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity); err != nil {
		t.Error(err)
	} else {
		rtsIntrvl := []*utils.RateSInterval{
			{
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		}
		rtsIntrvl[0].Cost(expectedRPCost.Rates)
		expectedRPCost.CostIntervals[0] = rtsIntrvl[0].AsRatesIntervalsCost()
		if !rcv.Equals(expectedRPCost) {
			t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(expectedRPCost), utils.ToJSON(rcv))
		}
	}

	expRpCostAfterV1 := expectedRPCost
	if err := rateS.V1CostForEvent(context.Background(), &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, expectedRPCost); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedRPCost, expRpCostAfterV1) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRpCostAfterV1), utils.ToJSON(expectedRPCost))
	}

	//fmt.Printf("received costV1: \n%s\n", utils.ToIJSON(expectedRPCost))

	if err := dm.RemoveRateProfile(context.Background(), rPrf.Tenant, rPrf.ID, true); err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEventUnmatchEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_PRF1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				FilterIDs:       []string{"*string:~*req.Destination:10"},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
			"RATE2": {
				ID: "RATE2",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				FilterIDs:       []string{"*string:~*req.Destination:10"},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}

	expectedErr := "invalid converter terminator in rule: <~*req.Cost{*>"
	rPrf.Rates["RATE2"].FilterIDs = []string{"*gt:~*req.Cost{*:10"}
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "10",
			utils.Cost:         1002,
		}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, receied %+v", expectedErr, err)
	}
	rPrf.Rates["RATE2"].FilterIDs = []string{"*prefix:~*req.CustomValue:randomValue"}

	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "10",
			utils.Cost:         1002,
		}}, rateS.cfg.RateSCfg().Verbosity); err != nil {
		t.Error(err)
	}

	if err := rateS.dm.RemoveRateProfile(context.Background(), rPrf.Tenant, rPrf.ID, true); err != nil {
		t.Error(err)
	}
}

func TestMatchingRateProfileEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rate := RateS{
		cfg:   cfg,
		fltrS: filters,
		dm:    dm,
	}
	t1 := time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC)
	rpp := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Weights: utils.DynamicWeights{
			{
				Weight: 7,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
	}
	err := dm.SetRateProfile(context.Background(), rpp, false, true)
	if err != nil {
		t.Error(err)
	}

	var ignoredRPIDs utils.StringSet
	if rtPRf, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtPRf, rpp) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rpp), utils.ToJSON(rtPRf))
	}

	if _, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  2002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  2002,
				utils.AnswerTime:   t1.Add(10 * time.Second),
			},
		}, false, ignoredRPIDs); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1007",
				utils.Destination:  1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{"rp2"},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1007",
				utils.Destination:  1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != utils.ErrNotFound {
		t.Error(err)
	}
	rpp.FilterIDs = []string{"*string:~*req.Account:1001|1002|1003", "*gt:~*req.Cost{*:10"}
	if _, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Cost:         1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err.Error() != "invalid converter terminator in rule: <~*req.Cost{*>" {
		t.Error(err)
	}
	rpp.FilterIDs = []string{"*string:~*req.Account:1001|1002|1003"}

	rate.dm = nil
	if _, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{"rp3"},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1007",
				utils.Destination:  1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}

	err = dm.RemoveRateProfile(context.Background(), rpp.Tenant, rpp.ID, true)
	if err != nil {
		t.Error(err)
	}
}

func TestV1CostForEventError(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}
	rcv, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity)
	if err != nil {
		t.Error(err)
	}

	expectedErr := "SERVER_ERROR: NOT_IMPLEMENTED:*notAType"
	rPrf.FilterIDs = []string{"*notAType:~*req.Account:1001"}
	if err := rateS.V1CostForEvent(context.Background(), &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.OptsRatesProfileIDs: []string{"RATE_1"},
		},
	}, rcv); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
	rPrf.FilterIDs = []string{"*string:~*req.Destination:10"}

	expectedErr = "SERVER_ERROR: zero increment to be charged within rate: <cgrates.org:RATE_1:RATE1>"
	rPrf.Rates["RATE1"].IntervalRates[0].Increment = utils.NewDecimal(0, 0)
	if err := rateS.V1CostForEvent(context.Background(), &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.Destination: "10"},
		APIOpts: map[string]any{
			utils.OptsRatesProfileIDs: []string{"RATE_1"},
		}}, rcv); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	if err := rateS.dm.RemoveRateProfile(context.Background(), rPrf.Tenant, rPrf.ID, true); err != nil {
		t.Error(err)
	}
}

// go test -run=^$ -v -bench=BenchmarkRateS_V1CostForEvent -benchtime=5s
func BenchmarkRateS_V1CostForEvent(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := RateS{
		cfg:   cfg,
		fltrS: filters,
		dm:    dm,
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		b.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		b.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RateChristmas",
		FilterIDs: []string{"*string:~*req.Subject:1010"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 50,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  utils.NewDecimal(1, 1),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RATE_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.IntervalRate{{
					IntervalStart: utils.NewDecimal(0, 0),
					RecurrentFee:  utils.NewDecimal(6, 3),
					Unit:          minDecimal,
					Increment:     secDecimal,
				}},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		b.Error(err)
	}
	if err := rPrf.Compile(); err != nil {
		b.Fatal(err)
	}
	if rcv, err := dm.GetRateProfile(context.TODO(), "cgrates.org", "RateChristmas",
		true, true, utils.NonTransactional); err != nil {
		b.Error(err)
	} else if !reflect.DeepEqual(rPrf, rcv) {
		b.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	rply := new(utils.RateProfileCost)
	argsRt := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			utils.Subject: "1010",
		},
		APIOpts: map[string]any{
			utils.OptsRatesStartTime: time.Date(2020, 12, 23, 59, 0, 0, 0, time.UTC),
			utils.OptsRatesUsage:     "2h",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rateS.V1CostForEvent(context.Background(), argsRt, rply); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()

}

// go test -run=^$ -v -bench=BenchmarkRateS_V1CostForEventSingleRate -benchtime=5s
func BenchmarkRateS_V1CostForEventSingleRate(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := RateS{
		cfg:   cfg,
		fltrS: filters,
		dm:    dm,
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		b.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		b.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RateAlways",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  utils.NewDecimal(1, 1),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		b.Error(err)
	}
	if err := rPrf.Compile(); err != nil {
		b.Fatal(err)
	}
	if rcv, err := dm.GetRateProfile(context.TODO(), "cgrates.org", "RateAlways",
		true, true, utils.NonTransactional); err != nil {
		b.Error(err)
	} else if !reflect.DeepEqual(rPrf, rcv) {
		b.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	rply := new(utils.RateProfileCost)
	argsRt := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			utils.Subject: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsRatesStartTime: time.Date(2020, 12, 23, 59, 0, 0, 0, time.UTC),
			utils.OptsRatesUsage:     "2h",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rateS.V1CostForEvent(context.Background(), argsRt, rply); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
}

func TestRateProfileCostForEventInvalidUsage(t *testing.T) {
	engine.Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)

	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}

	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}

	expected := "can't convert <invalidUsageFormat> to decimal"
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.OptsRatesUsage: "invalidUsageFormat",
		}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "Unsupported time format"
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.OptsRatesStartTime: "invalidStartTimeFormat",
		}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if err := dm.RemoveRateProfile(context.Background(), rPrf.Tenant, rPrf.ID, true); err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEventZeroIncrement(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)

	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "1 * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}

	expected := "zero increment to be charged within rate: <cgrates.org:RATE_1:RATE1>"
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.OptsRatesUsage: "100m",
		}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	if err := dm.RemoveRateProfile(context.Background(), rPrf.Tenant, rPrf.ID, true); err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEventMaximumIterations(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)

	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "1 * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}
	rateS.cfg.RateSCfg().Verbosity = 10

	expected := "maximum iterations reached"
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"},
		APIOpts: map[string]any{
			utils.OptsRatesUsage: "10000m",
		}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	if err := dm.RemoveRateProfile(context.Background(), rPrf.Tenant, rPrf.ID, true); err != nil {
		t.Error(err)
	}
}

func TestRateSMatchingRateProfileForEventErrFltr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(cfg, nil, dm)
	rateS := RateS{
		cfg:   cfg,
		fltrS: filterS,
		dm:    dm,
	}

	rPrf := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Weights: utils.DynamicWeights{
			{
				Weight:    10,
				FilterIDs: []string{"fi"},
			},
		},
		FilterIDs: []string{"*ai:~*req.AnswerTime:2020-07-21T00:00:00Z|9999-07-21T10:00:00Z"},
	}

	err := dm.SetRateProfile(context.Background(), rPrf, false, true)
	if err != nil {
		t.Error(err)
	}
	var ignoredRPfIDs utils.StringSet
	_, err = rateS.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "CACHE1",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  1002,
				utils.AnswerTime:   time.Date(9999, 7, 21, 10, 0, 0, 0, time.UTC).Add(-10 * time.Second),
			},
		}, false, ignoredRPfIDs)
	expectedErr := "NOT_FOUND"
	if err == nil || err.Error() != expectedErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expectedErr, err)
	}
}

func TestRateSRateProfileCostForEventErrFltr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{

				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight:    0,
						FilterIDs: []string{"fi"},
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}
	//MatchItmID before setting
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}

	expected := "NOT_FOUND:fi"
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v\n, received %+v", expected, err)
	}
}

func TestRateSRateProfileCostForEventErrInterval(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}
	expected := "can't convert <wrongValue> to decimal"
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		APIOpts: map[string]any{
			utils.OptsRatesIntervalStart: "wrongValue",
		},
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Error(err)
	}
}

func TestCDRProcessRatesCostForEvent(t *testing.T) {
	cache := engine.Cache
	engine.Cache.Clear(nil)
	jsonCfg := `{

		"cdrs": {
			"enabled": true,
			"rates_conns": ["*internal"],
			"opts": {
				"*rates": [
					{
						"Tenant": "cgrates.org",
						"FilterIDs": [],
						"Value": true,
					},			
				],					
			},	
		},

		"rates": {
			"enabled": true,
		},
	}
`
	cfg, err := config.NewCGRConfigFromJSONStringWithDefaults(jsonCfg)
	if err != nil {
		t.Error(err)
	}

	connMgr := engine.NewConnManager(cfg)
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, cfg.CacheCfg(), connMgr)
	filters := engine.NewFilterS(cfg, connMgr, dm)
	storDB := engine.NewInternalDB(nil, nil, nil)
	cdrs := cdrs.NewCDRServer(cfg, dm, filters, connMgr, storDB)
	ratesConns := make(chan birpc.ClientConnector, 1)
	rateSrv, err := birpc.NewServiceWithMethodsRename(NewRateS(cfg, filters, dm), utils.RateSv1, true, func(key string) (newKey string) {
		return strings.TrimPrefix(key, utils.V1Prfx)
	})
	if err != nil {
		t.Error(err)
	}
	ratesConns <- rateSrv

	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, ratesConns)
	// set a RateProfile for usage
	ratePrf := &utils.RateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_PROCESS_CDR",
		FilterIDs:       []string{"*string:~*req.Account:1001"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 1),
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(int64(time.Second), 1),
					},
				},
			},
		},
	}
	err = dm.SetRateProfile(context.Background(), ratePrf, false, true)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}

	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestCDRProcessRatesCostForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: 15 * time.Second,
		},
	}
	var reply string
	// here we will test that cdrs can communicate with rates correctly
	if err := cdrs.V1ProcessEvent(context.Background(), cgrEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+q, received %+q", utils.OK, reply)
	}

	//compare the new event that we got from RateSv1.CostForEvent called from cdrs
	expCgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestCDRProcessRatesCostForEvent",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaRateSCost: utils.RateProfileCost{
				ID:   "TEST_RATE_PROCESS_CDR",
				Cost: utils.NewDecimal(int64(15*time.Second)/10, 0),
				CostIntervals: []*utils.RateSIntervalCost{
					{
						Increments: []*utils.RateSIncrementCost{
							{
								Usage:             utils.NewDecimal(int64(15*time.Second), 0),
								RateID:            cgrEv.APIOpts[utils.MetaRateSCost].(utils.RateProfileCost).CostIntervals[0].Increments[0].RateID,
								RateIntervalIndex: 0,
								CompressFactor:    150,
							},
						},
						CompressFactor: 1,
					},
				},
				Rates: map[string]*utils.IntervalRate{
					cgrEv.APIOpts[utils.MetaRateSCost].(utils.RateProfileCost).CostIntervals[0].Increments[0].RateID: {
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 1),
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(int64(time.Second), 1),
					},
				},
			},
			utils.MetaCost:  utils.NewDecimal(int64(15*time.Second)/10, 0),
			utils.MetaUsage: 15 * time.Second,
		},
	}
	delete(cgrEv.APIOpts, utils.MetaCDRID) // ignore autogenerated *cdr field when comparing
	if !reflect.DeepEqual(expCgrEv, cgrEv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expCgrEv), utils.ToJSON(cgrEv))
	}

	engine.Cache = cache
}

func TestRateProfileCostForEventProfileIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.RateSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001",
			"*string:~*req.TestField:testValue"},
		Weights: utils.DynamicWeights{
			{
				Weight: 50,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
				},
			},
		},
	}

	//MatchItmID before setting
	if _, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

	if err := rateS.dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Error(err)
	}

	expectedRPCost := &utils.RateProfileCost{
		ID:   "RATE_1",
		Cost: utils.NewDecimal(2, 1),
		CostIntervals: []*utils.RateSIntervalCost{
			{
				Increments: []*utils.RateSIncrementCost{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"RATE1": {
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
		},
	}

	if rcv, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001"}}, rateS.cfg.RateSCfg().Verbosity); err != nil {
		t.Error(err)
	} else {
		rtsIntrvl := []*utils.RateSInterval{
			{
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "RATE1",
						CompressFactor:    1,
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
					},
				},
				CompressFactor: 1,
			},
		}
		rtsIntrvl[0].Cost(expectedRPCost.Rates)
		expectedRPCost.CostIntervals[0] = rtsIntrvl[0].AsRatesIntervalsCost()
		if !rcv.Equals(expectedRPCost) {
			t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(expectedRPCost), utils.ToJSON(rcv))
		}
	}
	event := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		Event: map[string]any{
			utils.AccountField: "1001",
			"TestField":        "testValue1",
		},
		APIOpts: map[string]any{
			utils.OptsRatesProfileIDs:      []string{"RATE_1"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	expRpCostAfterV1 := expectedRPCost
	if err := rateS.V1CostForEvent(context.Background(), event, expectedRPCost); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedRPCost, expRpCostAfterV1) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRpCostAfterV1), utils.ToJSON(expectedRPCost))
	}

	//fmt.Printf("received costV1: \n%s\n", utils.ToIJSON(expectedRPCost))

	if err := dm.RemoveRateProfile(context.Background(), rPrf.Tenant, rPrf.ID, true); err != nil {
		t.Error(err)
	}
}

func TestMatchingRateProfileFallbacks(t *testing.T) {

	cfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(cfg, nil, dm)
	rate := RateS{
		cfg:   cfg,
		fltrS: filters,
		dm:    dm,
	}
	t1 := time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC)

	rpp1 := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
	}

	rpp2 := &utils.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP2",
		Weights: utils.DynamicWeights{
			{
				Weight: 5,
			},
		},
		FilterIDs: []string{"*string:~*req.Account:1001|1002|1003", "*prefix:~*req.Destination:10"},
	}

	// Set both rate profiles
	err := dm.SetRateProfile(context.Background(), rpp1, false, true)
	if err != nil {
		t.Error(err)
	}
	err = dm.SetRateProfile(context.Background(), rpp2, false, true)
	if err != nil {
		t.Error(err)
	}

	// Initialize ignoredRPIDs as a new empty map
	ignoredRPIDs := make(utils.StringSet)

	// 1. Test without any ignored profiles (should match RP1 with higher weight)
	if rtPRf, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ID",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtPRf, rpp1) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rpp1), utils.ToJSON(rtPRf))
	}

	// 2. Test with ignoredRPIDs containing RP1 (should match RP2  )
	ignoredRPIDs.Add(rpp1.ID)
	if rtPRf, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ID",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtPRf, rpp2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rpp2), utils.ToJSON(rtPRf))
	}

	// 3. Test with both RP1 and RP2 in ignoredRPIDs (should return ErrNotFound)
	ignoredRPIDs.Add(rpp2.ID)
	if _, err := rate.matchingRateProfileForEvent(context.TODO(), "cgrates.org", []string{},
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ID",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  1002,
				utils.AnswerTime:   t1.Add(-10 * time.Second),
			},
		}, false, ignoredRPIDs); err != utils.ErrNotFound {
		t.Errorf("Expected error %v, got %v", utils.ErrNotFound, err)
	}

	err = dm.RemoveRateProfile(context.Background(), rpp1.Tenant, rpp1.ID, true)
	if err != nil {
		t.Error(err)
	}
	err = dm.RemoveRateProfile(context.Background(), rpp2.Tenant, rpp2.ID, true)
	if err != nil {
		t.Error(err)
	}
}
