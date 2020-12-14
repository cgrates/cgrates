/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

func TestListenAndServe(t *testing.T) {
	newRates := &RateS{}
	cfgRld := make(chan struct{}, 1)
	stopChan := make(chan struct{}, 1)
	cfgRld <- struct{}{}
	go func() {
		time.Sleep(10)
		stopChan <- struct{}{}
	}()
	newRates.ListenAndServe(stopChan, cfgRld)
}

func TestNewRateS(t *testing.T) {
	config := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, true)
	dataManager := engine.NewDataManager(data, config.CacheCfg(), nil)
	filters := engine.NewFilterS(config, nil, dataManager)
	expected := &RateS{
		cfg:     config,
		filterS: filters,
		dm:      dataManager,
	}
	if newRateS := NewRateS(config, filters, dataManager); !reflect.DeepEqual(newRateS, expected) {
		t.Errorf("Expected %+v, received %+v", expected, newRateS)
	}
}

func TestCallRates(t *testing.T) {
	newRates := &RateS{}
	var reply *string
	expectedErr := "UNSUPPORTED_SERVICE_METHOD"
	if err := newRates.Call("inexistentMethodCall", nil, reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestMatchingRateProfileForEventActivationInterval(t *testing.T) {
	dftCfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, dftCfg.CacheCfg(), nil)
	filterS := engine.NewFilterS(dftCfg, nil, dm)
	rateS := RateS{
		cfg:     dftCfg,
		filterS: filterS,
		dm:      dm,
	}

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		Weight:    10,
		FilterIDs: []string{"*string:~*req.Account:1001;1002;1003", "*prefix:~*req.Destination:10"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2020, 7, 21, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC),
		},
	}

	err := dm.SetRateProfile(rPrf, true)
	if err != nil {
		t.Error(err)
	}
	if _, err := rateS.matchingRateProfileForEvent("cgrates.org", []string{},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Time:   utils.TimePointer(time.Date(2020, 7, 21, 11, 0, 0, 0, time.UTC)),
					Event: map[string]interface{}{
						utils.Account:     "1001",
						utils.Destination: 1002,
						utils.AnswerTime:  rPrf.ActivationInterval.ExpiryTime.Add(-10 * time.Second),
					},
				},
			},
		}); err != utils.ErrNotFound {
		t.Error(err)
	}

	err = dm.RemoveRateProfile(rPrf.Tenant, rPrf.ID, utils.NonTransactional, true)
	if err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEvent(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	rateS := NewRateS(defaultCfg, filters, dm)

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": {
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
		},
	}

	//MatchItmID before setting
	if _, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}}}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}

	if err := rateS.dm.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}

	expectedRPCost := &engine.RateProfileCost{
		ID:   "RATE_1",
		Cost: 0.20,
		RateSIntervals: []*engine.RateSInterval{
			{
				UsageStart: 0,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        0,
						Rate:              rPrf.Rates["RATE1"],
						IntervalRateIndex: 0,
						CompressFactor:    1,
						Usage:             time.Minute,
					},
				},
				CompressFactor: 1,
			},
		},
	}
	expectedRPCost.RateSIntervals[0].Cost()

	if rcv, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}}}}, rateS.cfg.RateSCfg().Verbosity); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expectedRPCost) {
		t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(expectedRPCost), utils.ToJSON(rcv))
	}

	if err := rateS.V1CostForEvent(&utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}}}}, expectedRPCost); err != nil {
		t.Error(err)
	}
	err := dm.RemoveRateProfile(rPrf.Tenant, rPrf.ID, utils.NonTransactional, true)
	if err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEventUnmatchEvent(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	rateS := NewRateS(defaultCfg, filters, dm)

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_PRF1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": {
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "* * * * *",
				FilterIDs:       []string{"*string:~*req.Destination:10"},
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
			"RATE2": {
				ID:              "RATE2",
				Weight:          0,
				ActivationTimes: "* * * * *",
				FilterIDs:       []string{"*string:~*req.Destination:10"},
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}

	expectedErr := "invalid converter terminator in rule: <~*req.Cost{*>"
	rPrf.Rates["RATE2"].FilterIDs = []string{"*gt:~*req.Cost{*:10"}
	if _, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "10",
					utils.Cost:        1002,
				}}}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, receied %+v", expectedErr, err)
	}
	rPrf.Rates["RATE2"].FilterIDs = []string{"*prefix:~*req.CustomValue:randomValue"}

	if _, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account:     "1001",
					utils.Destination: "10",
					utils.Cost:        1002,
				}}}}, rateS.cfg.RateSCfg().Verbosity); err != nil {
		t.Error(err)
	}

	if err := rateS.dm.RemoveRateProfile(rPrf.Tenant, rPrf.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

func TestMatchingRateProfileEvent(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	rate := RateS{
		cfg:     defaultCfg,
		filterS: filters,
		dm:      dm,
	}
	t1 := time.Date(2020, 7, 21, 10, 0, 0, 0, time.UTC)
	rpp := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		Weight:    7,
		FilterIDs: []string{"*string:~*req.Account:1001;1002;1003", "*prefix:~*req.Destination:10"},
		ActivationInterval: &utils.ActivationInterval{
			ExpiryTime: t1,
		},
	}
	err := dm.SetRateProfile(rpp, true)
	if err != nil {
		t.Error(err)
	}

	if rtPRf, err := rate.matchingRateProfileForEvent("cgrates.org", []string{},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event: map[string]interface{}{
						utils.Account:     "1001",
						utils.Destination: 1002,
						utils.AnswerTime:  t1.Add(-10 * time.Second),
					},
				},
			},
		}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtPRf, rpp) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rpp), utils.ToJSON(rtPRf))
	}

	if _, err := rate.matchingRateProfileForEvent("cgrates.org", []string{},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event: map[string]interface{}{
						utils.Account:     "1001",
						utils.Destination: 2002,
						utils.AnswerTime:  t1.Add(-10 * time.Second),
					},
				},
			},
		}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := rate.matchingRateProfileForEvent("cgrates.org", []string{},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event: map[string]interface{}{
						utils.Account:     "1001",
						utils.Destination: 2002,
						utils.AnswerTime:  t1.Add(10 * time.Second),
					},
				},
			},
		}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := rate.matchingRateProfileForEvent("cgrates.org", []string{},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event: map[string]interface{}{
						utils.Account:     "1007",
						utils.Destination: 1002,
						utils.AnswerTime:  t1.Add(-10 * time.Second),
					},
				},
			},
		}); err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := rate.matchingRateProfileForEvent("cgrates.org", []string{"rp2"},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event: map[string]interface{}{
						utils.Account:     "1007",
						utils.Destination: 1002,
						utils.AnswerTime:  t1.Add(-10 * time.Second),
					},
				},
			},
		}); err != utils.ErrNotFound {
		t.Error(err)
	}
	rpp.FilterIDs = []string{"*string:~*req.Account:1001;1002;1003", "*gt:~*req.Cost{*:10"}
	if _, err := rate.matchingRateProfileForEvent("cgrates.org", []string{},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event: map[string]interface{}{
						utils.Account:    "1001",
						utils.Cost:       1002,
						utils.AnswerTime: t1.Add(-10 * time.Second),
					},
				},
			},
		}); err.Error() != "invalid converter terminator in rule: <~*req.Cost{*>" {
		t.Error(err)
	}
	rpp.FilterIDs = []string{"*string:~*req.Account:1001;1002;1003"}

	rate.dm = nil
	if _, err := rate.matchingRateProfileForEvent("cgrates.org", []string{"rp3"},
		&utils.ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event: map[string]interface{}{
						utils.Account:     "1007",
						utils.Destination: 1002,
						utils.AnswerTime:  t1.Add(-10 * time.Second),
					},
				},
			},
		}); err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}

	err = dm.RemoveRateProfile(rpp.Tenant, rpp.ID, utils.NonTransactional, true)
	if err != nil {
		t.Error(err)
	}
}

func TestV1CostForEventError(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	rateS := NewRateS(defaultCfg, filters, dm)

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": {
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}
	rcv, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}}}}, rateS.cfg.RateSCfg().Verbosity)
	if err != nil {
		t.Error(err)
	}

	expectedErr := "SERVER_ERROR: NOT_IMPLEMENTED:*notAType"
	rPrf.FilterIDs = []string{"*notAType:~*req.Account:1001"}
	if err := rateS.V1CostForEvent(&utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}}},
		RateProfileIDs: []string{"RATE_1"}}, rcv); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
	rPrf.FilterIDs = []string{"*string:~*req.Destination:10"}

	expectedErr = "SERVER_ERROR: zero increment to be charged within rate: <cgrates.org:RATE_1:RATE1>"
	rPrf.Rates["RATE1"].IntervalRates[0].Increment = 0
	if err := rateS.V1CostForEvent(&utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Destination: "10"}}},
		RateProfileIDs: []string{"RATE_1"}}, rcv); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	if err := rateS.dm.RemoveRateProfile(rPrf.Tenant, rPrf.ID, utils.NonTransactional, true); err != nil {
		t.Error(err)
	}
}

// go test -run=^$ -v -bench=BenchmarkRateS_V1CostForEvent -benchtime=5s
func BenchmarkRateS_V1CostForEvent(b *testing.B) {
	defaultCfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	rateS := RateS{
		cfg:     defaultCfg,
		filterS: filters,
		dm:      dm,
	}

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RateChristmas",
		FilterIDs: []string{"*string:~*req.Subject:1010"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": &engine.Rate{
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.10,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RATE_CHRISTMAS": &engine.Rate{
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{{
					IntervalStart: 0,
					RecurrentFee:  0.06,
					Unit:          time.Minute,
					Increment:     time.Second,
				}},
			},
		},
	}
	if err := dm.SetRateProfile(rPrf, true); err != nil {
		b.Error(err)
	}
	if err := rPrf.Compile(); err != nil {
		b.Fatal(err)
	}
	if rcv, err := dm.GetRateProfile("cgrates.org", "RateChristmas",
		true, true, utils.NonTransactional); err != nil {
		b.Error(err)
	} else if !reflect.DeepEqual(rPrf, rcv) {
		b.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	rply := new(engine.RateProfileCost)
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2020, 12, 23, 59, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:     "2h",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1010",
				},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rateS.V1CostForEvent(argsRt, rply); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()

}

// go test -run=^$ -v -bench=BenchmarkRateS_V1CostForEventSingleRate -benchtime=5s
func BenchmarkRateS_V1CostForEventSingleRate(b *testing.B) {
	defaultCfg := config.NewDefaultCGRConfig()

	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)
	rateS := RateS{
		cfg:     defaultCfg,
		filterS: filters,
		dm:      dm,
	}

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RateAlways",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": &engine.Rate{
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.10,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	if err := dm.SetRateProfile(rPrf, true); err != nil {
		b.Error(err)
	}
	if err := rPrf.Compile(); err != nil {
		b.Fatal(err)
	}
	if rcv, err := dm.GetRateProfile("cgrates.org", "RateAlways",
		true, true, utils.NonTransactional); err != nil {
		b.Error(err)
	} else if !reflect.DeepEqual(rPrf, rcv) {
		b.Errorf("Expecting: %v, received: %v", rPrf, rcv)
	}
	rply := new(engine.RateProfileCost)
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2020, 12, 23, 59, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:     "2h",
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
		if err := rateS.V1CostForEvent(argsRt, rply); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
}

func TestRatesShutDown(t *testing.T) {
	rateS := new(RateS)
	if err := rateS.Shutdown(); err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEventInvalidUsage(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)

	rateS := NewRateS(defaultCfg, filters, dm)

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": {
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}

	expected := "time: invalid duration \"invalidUsageFormat\""
	if _, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}},
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "invalidUsageFormat",
			}}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	expected = "Unsupported time format"
	if _, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}},
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: "invalidStartTimeFormat",
			}}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	err := dm.RemoveRateProfile(rPrf.Tenant, rPrf.ID, utils.NonTransactional, true)
	if err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEventZeroIncrement(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)

	rateS := NewRateS(defaultCfg, filters, dm)

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": {
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "1 * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     0,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}

	expected := "zero increment to be charged within rate: <cgrates.org:RATE_1:RATE1>"
	if _, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}},
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "100m",
			}}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
	err := dm.RemoveRateProfile(rPrf.Tenant, rPrf.ID, utils.NonTransactional, true)
	if err != nil {
		t.Error(err)
	}
}

func TestRateProfileCostForEventMaximumIterations(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	data := engine.NewInternalDB(nil, nil, true)
	dm := engine.NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	filters := engine.NewFilterS(defaultCfg, nil, dm)

	rateS := NewRateS(defaultCfg, filters, dm)

	rPrf := &engine.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Weight:    50,
		Rates: map[string]*engine.Rate{
			"RATE1": {
				ID:              "RATE1",
				Weight:          0,
				ActivationTimes: "1 * * * *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.20,
						Unit:          time.Minute,
						Increment:     0,
					},
				},
			},
		},
	}

	if err := rateS.dm.SetRateProfile(rPrf, true); err != nil {
		t.Error(err)
	}
	rateS.cfg.RateSCfg().Verbosity = 10

	expected := "maximum iterations reached"
	if _, err := rateS.rateProfileCostForEvent(rPrf, &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "RATE_1",
				Event: map[string]interface{}{
					utils.Account: "1001"}},
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "10000m",
			}}}, rateS.cfg.RateSCfg().Verbosity); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}

	err := dm.RemoveRateProfile(rPrf.Tenant, rPrf.ID, utils.NonTransactional, true)
	if err != nil {
		t.Error(err)
	}
}
