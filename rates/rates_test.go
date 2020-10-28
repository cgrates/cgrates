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
	exitChan := make(chan bool, 1)
	cfgRld <- struct{}{}
	go func() {
		time.Sleep(10)
		exitChan <- true
	}()
	if err := newRates.ListenAndServe(exitChan, cfgRld); err != nil {
		t.Error(err)
	}
}

func TestNewRateS(t *testing.T) {
	config, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
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

func TestStartTime(t *testing.T) {
	cgrEvent := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "randomID",
			Time:   utils.TimePointer(time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC)),
			Event: map[string]interface{}{
				utils.ToR: utils.SMS,
			},
		},
	}
	rateProfileIDs := []string{"randomIDs"}
	argsCost := &ArgsCostForEvent{
		rateProfileIDs,
		cgrEvent,
	}
	expectedTime := time.Date(2020, 12, 24, 0, 0, 0, 0, time.UTC)
	if sTime, err := argsCost.StartTime(utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTime, sTime) {
		t.Errorf("Expected %+v, received %+v", expectedTime, sTime)
	}
}

func TestStartTimeError(t *testing.T) {
	cgrEvent := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "randomID",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.ToR: utils.SMS,
			},
		},
		Opts: map[string]interface{}{
			utils.OptsRatesStartTime: "invalidStartTime",
		},
	}
	rateProfileIDs := []string{"randomIDs"}
	argsCost := &ArgsCostForEvent{
		rateProfileIDs,
		cgrEvent,
	}
	expectedErr := "Unsupported time format"
	if _, err := argsCost.StartTime(utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestStartTimeFieldAsTimeErrorAnswerTime(t *testing.T) {
	cgrEvent := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "randomID",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.AnswerTime: "0s",
			},
		},
	}
	rateProfileIDs := []string{"randomIDs"}
	argsCost := &ArgsCostForEvent{
		rateProfileIDs,
		cgrEvent,
	}
	expectedErr := "Unsupported time format"
	if _, err := argsCost.StartTime(utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestStartTimeFieldAsTimeErrorSetupTime(t *testing.T) {
	cgrEvent := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "randomID",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.SetupTime: "1h0m0s",
			},
		},
	}
	rateProfileIDs := []string{"randomIDs"}
	argsCost := &ArgsCostForEvent{
		rateProfileIDs,
		cgrEvent,
	}
	expectedErr := "Unsupported time format"
	if _, err := argsCost.StartTime(utils.EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func TestStartTimeFieldAsTimeNilTime(t *testing.T) {
	cgrEvent := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "randomID",
			Event: map[string]interface{}{
				utils.AnswerTime: time.Time{},
			},
		},
	}
	rateProfileIDs := []string{"randomIDs"}
	argsCost := &ArgsCostForEvent{
		rateProfileIDs,
		cgrEvent,
	}
	if _, err := argsCost.StartTime(utils.EmptyString); err != nil {
		t.Error(err)
	}
}

func TestStartTimeFieldAsTimeNow(t *testing.T) {
	cgrEvent := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "randomID",
			Event:  map[string]interface{}{},
		},
	}
	rateProfileIDs := []string{"randomIDs"}
	argsCost := &ArgsCostForEvent{
		rateProfileIDs,
		cgrEvent,
	}
	if _, err := argsCost.StartTime(utils.EmptyString); err != nil {
		t.Error(err)
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

func TestV1CostForEvent(t *testing.T) {
	rts := &RateS{}
	cgrEvent := &utils.CGREventWithOpts{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "randomID",
			Event:  map[string]interface{}{},
		},
	}
	rateProfileIDs := []string{"randomIDs"}
	argsCost := &ArgsCostForEvent{
		rateProfileIDs,
		cgrEvent,
	}
	if err := rts.V1CostForEvent(argsCost, nil); err != nil {
		t.Error(err)
	}
}

func TestMatchingRateProfileEvent(t *testing.T) {
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Fatal(err)
	}
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
	err = dm.SetRateProfile(rpp, true)
	if err != nil {
		t.Error(err)
	}
	if rtPRf, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
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
		},
		[]string{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rtPRf, rpp) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rpp), utils.ToJSON(rtPRf))
	}

	if _, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
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
		},
		[]string{}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
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
		},
		[]string{}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
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
		},
		[]string{}); err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
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
		},
		[]string{"rp2"}); err != utils.ErrNotFound {
		t.Error(err)
	}

	if _, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
			CGREventWithOpts: &utils.CGREventWithOpts{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "CACHE1",
					Event:  map[string]interface{}{},
				},
				Opts: map[string]interface{}{
					utils.OptsRatesStartTime: 0,
				},
			},
		},
		[]string{"randomProfileID"}); err.Error() != "cannot convert field: 0 to time.Time" {
		t.Error(err)
	}

	rpp.FilterIDs = []string{"*string:~*req.Account:1001;1002;1003", "*gt:~*req.Cost{*:10"}
	if _, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
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
		},
		[]string{}); err.Error() != "invalid converter terminator in rule: <~*req.Cost{*>" {
		t.Error(err)
	}

	rate.dm = nil
	if _, err := rate.matchingRateProfileForEvent("cgrates.org",
		&ArgsCostForEvent{
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
		},
		[]string{"rp3"}); err != utils.ErrNoDatabaseConn {
		t.Error(err)
	}
}
