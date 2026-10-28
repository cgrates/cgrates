// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package rates

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestNewRateS(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)

	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dataManager := engine.NewDataManager(dbCM, cfg, nil, locker)
	dataManager.SetCache(cacheS)
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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
	cacheS.Clear(nil)
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
	cacheS.Clear(nil)
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
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	cacheS := engine.NewCacheS(cfg, nil, nil, nil, locker)
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(cacheS)
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

	expectedErr := "SERVER_ERROR: Unsupported filter Type: *notAType"
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
	cacheS.Clear(nil)
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
	locker := engine.NewLocker(cfg)

	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)

	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)

	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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

func TestRateProfileCostForEventProfileIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	cfg.RateSCfg().Opts.ProfileIgnoreFilters = []*config.DynamicBoolOpt{
		config.NewDynamicBoolOpt(nil, "", true, nil),
	}
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
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

func TestRateProfileCostForEventMinMaxCost(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		MinCost:   utils.NewDecimal(5, 0),
		MaxCost:   utils.NewDecimal(10, 0),
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
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
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name    string
		usage   *utils.Decimal
		expCost *utils.Decimal
	}{
		{
			name:    "Below MinCost",
			usage:   utils.NewDecimal(10, 0),
			expCost: utils.NewDecimal(2, 0),
		},
		{
			name:    "Between MinCost And MaxCost",
			usage:   utils.NewDecimal(30, 0),
			expCost: utils.NewDecimal(6, 0),
		},
		{
			name:    "Above MaxCost",
			usage:   utils.NewDecimal(60, 0),
			expCost: utils.NewDecimal(12, 0),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE1",
				Event: map[string]any{
					utils.AccountField: "1001",
				},
				APIOpts: map[string]any{
					utils.OptsRatesStartTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
					utils.OptsRatesUsage:     tc.usage,
				},
			}
			rpCost, err := rateS.rateProfileCostForEvent(context.Background(), rPrf, ev, rateS.cfg.RateSCfg().Verbosity)
			if err != nil {
				t.Fatal(err)
			}
			if rpCost.MinCost == nil || rpCost.MinCost.Compare(rPrf.MinCost) != 0 {
				t.Errorf("Expected MinCost %v, received %v", rPrf.MinCost, rpCost.MinCost)
			}
			if rpCost.MaxCost == nil || rpCost.MaxCost.Compare(rPrf.MaxCost) != 0 {
				t.Errorf("Expected MaxCost %v, received %v", rPrf.MaxCost, rpCost.MaxCost)
			}
			if rpCost.Cost.Compare(tc.expCost) != 0 {
				t.Errorf("Expected cost %v, received %v", tc.expCost, rpCost.Cost)
			}
		})
	}
}

func TestV1CostForEventRateProfileSelection(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)
	ctx := context.Background()

	rp1 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
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
						RecurrentFee:  utils.NewDecimal(5, 1),
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	rp2 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
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
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rp1, false, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetRateProfile(context.Background(), rp2, false, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "rateEV",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsRatesStartTime: time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC),
			utils.OptsRatesUsage:     utils.NewDecimal(3, 0),
		},
	}
	var rpCost utils.RateProfileCost
	if err := rateS.V1CostForEvent(ctx, ev, &rpCost); err != nil {
		t.Error(err)
	}
	if rpCost.ID != rp2.ID {
		t.Errorf("Expected %v, received %v", rp2.ID, rpCost.ID)
	}
	if rpCost.Cost.Compare(utils.NewDecimal(6, 1)) != 0 {
		t.Errorf("Expected cost 0.6, received %v", rpCost.Cost)
	}
}

func TestOrderRatesOnIntervalsRateChangeDuringCall(t *testing.T) {
	startTime := time.Date(2026, time.October, 22, 12, 30, 0, 0, time.UTC)
	rate1 := &utils.Rate{
		ID: "rate1",
	}
	rate2 := &utils.Rate{
		ID: "rate2",
	}
	for _, tc := range []struct {
		name         string
		r2ActiveTime string
		weights      []float64
		usage        time.Duration
		expIDs       []string
		expOrdered   []*orderedRate
	}{
		{
			name:         "Higher weight starts during call",
			r2ActiveTime: "* 13 * * *",
			weights:      []float64{10, 20},
			usage:        time.Hour,
			expIDs:       []string{"rate1", "rate2"},
			expOrdered: []*orderedRate{
				{
					utils.NewDecimal(0, 0),
					rate1,
				},
				{
					utils.NewDecimal(int64(30*time.Minute), 0),
					rate2,
				},
			},
		},
		{
			name:         "Higher weight expires during call",
			r2ActiveTime: "* 12 * * *",
			weights:      []float64{10, 20},
			usage:        time.Hour,
			expIDs:       []string{"rate2", "rate1"},
			expOrdered: []*orderedRate{
				{
					utils.NewDecimal(0, 0),
					rate2,
				},
				{
					utils.NewDecimal(int64(30*time.Minute), 0),
					rate1,
				},
			},
		},
		{
			name:         "Lower weight starts during call",
			r2ActiveTime: "* 13 * * *",
			weights:      []float64{20, 10},
			usage:        time.Hour,
			expIDs:       []string{"rate1"},
			expOrdered: []*orderedRate{
				{
					utils.NewDecimal(0, 0),
					rate1,
				},
			},
		},
		{
			name:         "Higher weight starts and expires during call",
			r2ActiveTime: "* 13 * * *",
			weights:      []float64{10, 20},
			usage:        2 * time.Hour,
			expIDs:       []string{"rate1", "rate2", "rate1"},
			expOrdered: []*orderedRate{
				{
					utils.NewDecimal(0, 0),
					rate1,
				},
				{
					utils.NewDecimal(int64(30*time.Minute), 0),
					rate2,
				},
				{
					utils.NewDecimal(int64(90*time.Minute), 0),
					rate1,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rate1.ActivationTimes = "* * * * *"
			rate2.ActivationTimes = tc.r2ActiveTime
			rate1.Compile()
			rate2.Compile()
			if ordRts, err := orderRatesOnIntervals([]*utils.Rate{rate1, rate2}, tc.weights, startTime, utils.NewDecimal(int64(tc.usage), 0), true, 10); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(tc.expOrdered, ordRts) {
				t.Log(ordRts)
				t.Errorf("expecting: %v\n, received: %v", tc.expOrdered, ordRts)
			}
		})
	}
}

func TestMatchingRateProfileWeightsWithFilterIDs(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)

	rp1 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rateProfileID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*prefix:~*req.Destination:+49"},
				Weight:    50,
			},
			{
				FilterIDs: []string{"*prefix:~*req.Destination:+44"},
				Weight:    40,
			},
		},
	}
	rp2 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rateProfileID2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rp1, false, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetRateProfile(context.Background(), rp2, false, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name        string
		destination string
		expID       string
	}{
		{
			name:        "first filter with higher weight than rp2",
			destination: "+4930123456",
			expID:       "rateProfileID1",
		},
		{
			name:        "second filter with higher weight than rp2",
			destination: "+4430123456",
			expID:       "rateProfileID1",
		},
		{
			name:        "destination that doesn't belong to any filter",
			destination: "+3312345678",
			expID:       "rateProfileID2",
		},
		{
			name:        "empty destination",
			destination: "",
			expID:       "rateProfileID2",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "rpID",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  tc.destination,
				},
			}
			rp, err := rateS.matchingRateProfileForEvent(context.Background(), "cgrates.org", nil, ev, false, nil)
			if err != nil {
				t.Error(err)
			}
			if rp.ID != tc.expID {
				t.Errorf("Expected %v, received %v", tc.expID, rp.ID)
			}
		})
	}
}

func TestMatchingRateProfileWeightsFilterIdsAreIgnored(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)

	rp1 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rateProfileID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*prefix:~*req.Destination:+49"},
				Weight:    50,
			},
			{
				FilterIDs: []string{"*prefix:~*req.Destination:+44"},
				Weight:    40,
			},
			{
				FilterIDs: []string{"*prefix:~*req.Destination:+40"},
				Weight:    70,
			},
		},
	}
	rp2 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rateProfileID2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 60,
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rp1, false, true); err != nil {
		t.Fatal(err)
	}
	if err := dm.SetRateProfile(context.Background(), rp2, false, true); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name        string
		destination string
		expID       string
	}{
		{
			name:        "first filter with lower weight than the rp2",
			destination: "+49220344445",
			expID:       "rateProfileID2",
		},
		{
			name:        "second filter with lower weight than the rp2",
			destination: "+44334564777",
			expID:       "rateProfileID2",
		},
		{
			name:        "third filter with higher weight than the rp2",
			destination: "+40445632229",
			expID:       "rateProfileID1",
		},
		{
			name:        "destination that doesn't have a filter",
			destination: "+30556798955",
			expID:       "rateProfileID2",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "rpID",
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  tc.destination,
				},
			}
			rp, err := rateS.matchingRateProfileForEvent(context.Background(), "cgrates.org", nil, ev, false, nil)
			if err != nil {
				t.Error(err)
			}
			if rp.ID != tc.expID {
				t.Errorf("Expected %v, received %v", tc.expID, rp.ID)
			}
		})
	}
}

func TestMatchingRateProfileWeightsUnsupportedType(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)

	rp := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rateProfileID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: []string{"*inv:~*req.Destination:+49"},
				Weight:    50,
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rp, false, true); err != nil {
		t.Fatal(err)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "rpID",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "+4930123456",
		},
	}
	expErr := "Unsupported filter Type: *inv"
	if _, err := rateS.matchingRateProfileForEvent(context.Background(), "cgrates.org", nil, ev, false, nil); err == nil || err.Error() != expErr {
		t.Errorf("Expected error %v, recieved %v", expErr, err)
	}
}

func TestMatchingRateProfileForEventNilEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	locker := engine.NewLocker(cfg)
	data, _ := engine.NewInternalDB(nil, cfg.DbCfg().Items)
	dbCM := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: data}, cfg.DbCfg())
	dm := engine.NewDataManager(dbCM, cfg, nil, locker)
	dm.SetCache(engine.NewCacheS(cfg, nil, nil, nil, locker))
	filters := engine.NewFilterS(cfg, nil, dm)
	rateS := NewRateS(cfg, filters, dm)

	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "rpID1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Rates: map[string]*utils.Rate{
			"RATE1": {
				ID: "RATE1",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          utils.NewDecimal(1, 0),
						Increment:     utils.NewDecimal(1, 0),
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(context.Background(), rPrf, false, true); err != nil {
		t.Fatal(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "evID1",
		Event:  nil,
	}
	_, err := rateS.matchingRateProfileForEvent(context.Background(), "cgrates.org", []string{"rpID1"}, ev, false, nil)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %v, received %v", utils.ErrNotFound, err)
	}
}
