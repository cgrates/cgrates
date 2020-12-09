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
}

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
