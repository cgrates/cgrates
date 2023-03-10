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
package engine

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var rsponder *Responder

func init() {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
	rsponder = &Responder{MaxComputedUsage: cfg.RalsCfg().MaxComputedUsage}
}

func TestResponderGobSMCost(t *testing.T) {
	cc := &CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "data",
		ToR:         "*data",
		Cost:        0,
		Timespans: TimeSpans{&TimeSpan{
			TimeStart: time.Date(2016, 1, 5, 12, 30, 10, 0, time.UTC),
			TimeEnd:   time.Date(2016, 1, 5, 12, 55, 46, 0, time.UTC),
			Cost:      0,
			RateInterval: &RateInterval{
				Timing: nil,
				Rating: &RIRate{
					ConnectFee:       0,
					RoundingMethod:   "",
					RoundingDecimals: 0,
					MaxCost:          0,
					MaxCostStrategy:  "",
					Rates: RateGroups{&Rate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      1 * time.Second,
						RateUnit:           1 * time.Second,
					},
					},
				},
				Weight: 0,
			},
			DurationIndex: 0,
			Increments: Increments{&Increment{
				Duration: 1 * time.Second,
				Cost:     0,
				BalanceInfo: &DebitInfo{
					Unit: &UnitInfo{
						UUID:          "fa0aa280-2b76-4b5b-bb06-174f84b8c321",
						ID:            "",
						Value:         100864,
						DestinationID: "data",
						Consumed:      1,
						ToR:           "*data",
						RateInterval:  nil,
					},
					Monetary:  nil,
					AccountID: "cgrates.org:1001",
				},
				CompressFactor: 1536,
			},
			},
			RoundIncrement: nil,
			MatchedSubject: "fa0aa280-2b76-4b5b-bb06-174f84b8c321",
			MatchedPrefix:  "data",
			MatchedDestId:  "*any",
			RatingPlanId:   "*none",
			CompressFactor: 1,
		},
		},
		RatedUsage: 1536,
	}
	attr := AttrCDRSStoreSMCost{
		Cost: &SMCost{
			CGRID:       "b783a8bcaa356570436983cd8a0e6de4993f9ba6",
			RunID:       utils.MetaDefault,
			OriginHost:  "",
			OriginID:    "testdatagrp_grp1",
			CostSource:  "SMR",
			Usage:       1536,
			CostDetails: NewEventCostFromCallCost(cc, "b783a8bcaa356570436983cd8a0e6de4993f9ba6", utils.MetaDefault),
		},
		CheckDuplicate: false,
	}

	var network bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&network) // Will write to network.
	dec := gob.NewDecoder(&network) // Will read from network.
	err := enc.Encode(attr)
	if err != nil {
		t.Error("encode error: ", err)
	}

	// Decode (receive) and print the values.
	var q AttrCDRSStoreSMCost
	err = dec.Decode(&q)
	if err != nil {
		t.Error("decode error: ", err)
	}
	q.Cost.CostDetails.initCache()
	if !reflect.DeepEqual(attr, q) {
		t.Error("wrong transmission")
	}
}

func TestResponderUsageAllow(t *testing.T) {
	rsp := &Responder{
		MaxComputedUsage: map[string]time.Duration{
			utils.ANY:   time.Duration(10 * time.Second),
			utils.VOICE: time.Duration(20 * time.Second),
		},
	}
	if allow := rsp.usageAllowed(utils.VOICE, time.Duration(17*time.Second)); !allow {
		t.Errorf("Expected true, received : %+v", allow)
	}
	if allow := rsp.usageAllowed(utils.VOICE, time.Duration(22*time.Second)); allow {
		t.Errorf("Expected false, received : %+v", allow)
	}
	if allow := rsp.usageAllowed(utils.DATA, time.Duration(7*time.Second)); !allow {
		t.Errorf("Expected true, received : %+v", allow)
	}
	if allow := rsp.usageAllowed(utils.DATA, time.Duration(12*time.Second)); allow {
		t.Errorf("Expected false, received : %+v", allow)
	}
}

func TestResponderGetCostMaxUsageANY(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.ANY,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var cc CallCost
	if err := rsponder.GetCost(cd, &cc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderGetCostMaxUsageVOICE(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.VOICE,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var cc CallCost
	if err := rsponder.GetCost(cd, &cc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderDebitMaxUsageANY(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.ANY,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var cc CallCost
	if err := rsponder.Debit(cd, &cc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderDebitMaxUsageVOICE(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.VOICE,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var cc CallCost
	if err := rsponder.Debit(cd, &cc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderMaxDebitMaxUsageANY(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.ANY,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var cc CallCost
	if err := rsponder.MaxDebit(cd, &cc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderMaxDebitMaxUsageVOICE(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.VOICE,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var cc CallCost
	if err := rsponder.MaxDebit(cd, &cc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderRefundIncrementsMaxUsageANY(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.ANY,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var acc Account
	if err := rsponder.RefundIncrements(cd, &acc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderRefundIncrementsMaxUsageVOICE(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.VOICE,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var acc Account
	if err := rsponder.RefundIncrements(cd, &acc); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderRefundRoundingMaxUsageANY(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.ANY,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var reply Account
	if err := rsponder.RefundRounding(cd, &reply); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderRefundRoundingMaxUsageVOICE(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.VOICE,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var reply Account
	if err := rsponder.RefundRounding(cd, &reply); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderGetMaxSessionTimeMaxUsageANY(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.ANY,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var reply time.Duration
	if err := rsponder.GetMaxSessionTime(cd, &reply); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderGetMaxSessionTimeMaxUsageVOICE(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.ANY:   time.Duration(10 * time.Second),
		utils.VOICE: time.Duration(20 * time.Second),
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.VOICE,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
			TimeStart:     tStart,
			TimeEnd:       tEnd,
		},
	}
	var reply time.Duration
	if err := rsponder.GetMaxSessionTime(cd, &reply); err == nil ||
		err.Error() != utils.ErrMaxUsageExceeded.Error() {
		t.Errorf("Expected %+v, received : %+v", utils.ErrMaxUsageExceeded, err)
	}
}

func TestResponderGetCostOnRatingPlans(t *testing.T) {
	tmpCache := Cache
	defer func() {
		Cache = tmpCache
	}()
	arg := &utils.GetCostOnRatingPlansArgs{
		Tenant:        "cgrates.org",
		Account:       "test",
		Subject:       "1001",
		Destination:   "*any",
		SetupTime:     time.Date(2023, 3, 10, 8, 0, 0, 0, time.UTC),
		Usage:         20 * time.Second,
		RatingPlanIDs: []string{"RPL1"},
	}
	rp := &RatingPlan{
		Id: "RPL1",
		Timings: map[string]*RITiming{
			"59a981b9": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"ebefae11": {
				ConnectFee: 0,
				Rates: []*Rate{
					{
						GroupIntervalStart: 0,
						Value:              0.2,
						RateIncrement:      2 * time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 1,
			},
		},
		DestinationRates: map[string]RPRateList{
			"*any": []*RPRate{
				{
					Timing: "59a981b9",
					Rating: "ebefae11",
					Weight: 10,
				},
			},
		},
	}
	Cache.Set(utils.CacheRatingPlans, "RPL1", rp, []string{}, true, utils.NonTransactional)
	var reply map[string]interface{}
	exp := map[string]interface{}{
		utils.Cost:         4.0,
		utils.RatingPlanID: "RPL1",
	}
	if err := rsponder.GetCostOnRatingPlans(arg, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expected %v,Received %v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func TestResponderGetCost(t *testing.T) {
	tmpCache := Cache
	cfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Error(err)
	}
	tmpDm := dm
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)

	defer func() {
		Cache = tmpCache
		SetDataStorage(tmpDm)
	}()
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.VOICE: time.Duration(20 * time.Second),
	}
	cd := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			ToR:           utils.VOICE,
			Account:       "1001",
			Destination:   "1002",
			DurationIndex: 90,
			TimeStart:     time.Date(2023, 3, 10, 9, 30, 0, 0, time.UTC),
			TimeEnd:       time.Date(2023, 3, 10, 9, 30, 2, 0, time.UTC),
		},
	}
	rpDflt := &RatingPlan{
		Id: "RP_DFLT",
		Timings: map[string]*RITiming{
			"30eab301": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f861": {
				Rates: []*Rate{
					{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"DEST": []*RPRate{
				{
					Timing: "30eab301",
					Rating: "b457f861",
					Weight: 10,
				},
			},
		},
	}

	rpfTCDDBSWF := &RatingProfile{Id: utils.ConcatenatedKey(utils.META_OUT, cd.Tenant, "call", cd.Subject),
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:   rpDflt.Id,
		}},
	}
	dest := &Destination{
		Id:       "DEST",
		Prefixes: []string{"1001", "1002", "1003"},
	}
	if err := dm.SetReverseDestination(dest, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	Cache.Set(utils.CacheRatingPlans, rpDflt.Id, rpDflt, []string{}, true, utils.NonTransactional)
	Cache.Set(utils.CacheRatingProfilesTmp, rpfTCDDBSWF.Id, rpfTCDDBSWF, []string{}, true, utils.NonTransactional)
	var reply CallCost
	SetDataStorage(dm)
	if err := rsponder.GetCost(cd, &reply); err != nil {
		t.Error(err)
	} else if reply.Cost != 0.02 {
		t.Errorf("Received %v", reply.Cost)
	}
}

func TestResponderDebit11(t *testing.T) {
	tmpCache := Cache
	tmpDm := *dm
	defer func() {
		Cache = tmpCache
		SetDataStorage(&tmpDm)
	}()
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.VOICE: time.Duration(60 * time.Second),
	}
	arg := &CallDescriptorWithArgDispatcher{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			ToR:           utils.VOICE,
			Account:       "1001",
			Destination:   "1002",
			DurationIndex: 90,
			TimeStart:     time.Date(2023, 3, 10, 9, 30, 0, 0, time.UTC),
			TimeEnd:       time.Date(2023, 3, 10, 9, 30, 16, 0, time.UTC),
		},
	}
	acc := &Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]Balances{
			utils.VOICE: {
				&Balance{Value: 20 * float64(time.Second),
					DestinationIDs: utils.NewStringMap("1002"),
					Weight:         10, RatingSubject: "rif"},
			}},
	}
	dm.SetAccount(acc)
	var reply CallCost
	if err := rsponder.Debit(arg, &reply); err != nil {
		t.Error(err)
	}

}
