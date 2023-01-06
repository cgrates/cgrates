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
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var rsponder = &Responder{MaxComputedUsage: config.CgrConfig().RalsCfg().MaxComputedUsage}

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
					Rates: RateGroups{&RGRate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
					},
				},
				Weight: 0,
			},
			DurationIndex: 0,
			Increments: Increments{&Increment{
				Duration: time.Second,
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
			utils.MetaAny:   10 * time.Second,
			utils.MetaVoice: 20 * time.Second,
		},
	}
	if allow := rsp.usageAllowed(utils.MetaVoice, 17*time.Second); !allow {
		t.Errorf("Expected true, received : %+v", allow)
	}
	if allow := rsp.usageAllowed(utils.MetaVoice, 22*time.Second); allow {
		t.Errorf("Expected false, received : %+v", allow)
	}
	if allow := rsp.usageAllowed(utils.MetaData, 7*time.Second); !allow {
		t.Errorf("Expected true, received : %+v", allow)
	}
	if allow := rsp.usageAllowed(utils.MetaData, 12*time.Second); allow {
		t.Errorf("Expected false, received : %+v", allow)
	}
}

func TestResponderGetCostMaxUsageANY(t *testing.T) {
	rsponder.MaxComputedUsage = map[string]time.Duration{
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaAny,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaVoice,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaAny,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaVoice,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaAny,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaVoice,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaAny,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaVoice,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaAny,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaVoice,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:11Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaAny,
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
		utils.MetaAny:   10 * time.Second,
		utils.MetaVoice: 20 * time.Second,
	}
	tStart, _ := utils.ParseTimeDetectLayout("2013-08-07T17:30:00Z", "")
	tEnd, _ := utils.ParseTimeDetectLayout("2013-08-07T17:31:21Z", "")
	cd := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaVoice,
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

func TestResponderGetCost(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
		},
		APIOpts: map[string]interface{}{},
	}
	reply := &CallCost{

		Category:    "category",
		Tenant:      "tenant",
		Subject:     "subject",
		Account:     "acount",
		Destination: "uk",
	}

	if err = rs.GetCost(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderGetCost, arg.CgrID))

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(rcv), utils.ToJSON(exp))
	}

}

func TestResponderGetCostSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
		},
		APIOpts: map[string]interface{}{},
	}
	reply := &CallCost{

		Category:    "category",
		Tenant:      "tenant",
		Subject:     "subject",
		Account:     "acount",
		Destination: "uk",
	}
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderGetCost, arg.CgrID),
		&utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)

	if err = rs.GetCost(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderGetCost, arg.CgrID))

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %+v,received %+v", utils.ToJSON(rcv), utils.ToJSON(exp))
	}

}

func TestResponderDebit(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &CallCost{

		Category:    "category",
		Tenant:      "tenant",
		Subject:     "subject",
		Account:     "acount",
		Destination: "uk",
	}
	if err := rs.Debit(arg, reply); err == nil || err != utils.ErrAccountNotFound {
		t.Errorf("expected %+v ,received %+v", utils.ErrAccountNotFound, err)
	}
}

func TestGetCostOnRatingPlansErr(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	arg := &utils.GetCostOnRatingPlansArgs{
		Account:       "account",
		Subject:       "subj",
		Destination:   "destination",
		Tenant:        "cgrates.org",
		SetupTime:     time.Date(2021, 12, 24, 8, 0, 0, 0, time.UTC),
		Usage:         10 * time.Minute,
		RatingPlanIDs: []string{"rplan1", "rplan2", "rplan3"},
		APIOpts: map[string]interface{}{
			"apiopts": "opt",
		},
	}
	reply := &map[string]interface{}{}
	rs := &Responder{
		FilterS: &FilterS{
			cfg: cfg,
			dm:  dm,
		},
	}
	if err := rs.GetCostOnRatingPlans(arg, reply); err == nil || err != utils.ErrUnauthorizedDestination {
		t.Errorf("expected %+v ,received %+v", utils.ErrUnauthorizedDestination, err)
	}
}

func TestSetMaxComputedUsage(t *testing.T) {

	rs := &Responder{
		Timeout:  10 * time.Minute,
		Timezone: "UTC",
	}

	mx := map[string]time.Duration{
		"usage1": 2 * time.Minute,
		"usage2": 4 * time.Minute,
	}
	rs.SetMaxComputedUsage(mx)
	if !reflect.DeepEqual(rs.MaxComputedUsage, mx) {
		t.Errorf("expected %v,received %v", mx, rs.MaxComputedUsage)
	}
}

func TestResponderDebitSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &CallCost{

		Category:    "category",
		Tenant:      "tenant",
		Subject:     "subject",
		Account:     "acount",
		Destination: "uk",
	}
	key := utils.ConcatenatedKey(utils.ResponderDebit, arg.CgrID)
	Cache.Set(utils.CacheRPCResponses, key,
		&utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)

	if err := rs.Debit(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{Result: reply, Error: nil}
	rcv, has := Cache.Get(utils.CacheRPCResponses, key)

	if !has {
		t.Error("has no values")
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %+v,received %+v", exp, rcv)
	}
}

func TestResponderMaxDebit(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &CallCost{

		Category:    "category",
		Tenant:      "tenant",
		Subject:     "subject",
		Account:     "acount",
		Destination: "uk",
	}
	if err := rs.MaxDebit(arg, reply); err == nil || err != utils.ErrAccountNotFound {
		t.Errorf("expected %+v ,received %+v", utils.ErrAccountNotFound, err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderMaxDebit, arg.CgrID))

	if !has {
		t.Error("has no value")
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestResponderMaxDebitSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &CallCost{

		Category:    "category",
		Tenant:      "tenant",
		Subject:     "subject",
		Account:     "acount",
		Destination: "uk",
	}
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderMaxDebit, arg.CgrID),
		&utils.CachedRPCResponse{Result: reply, Error: nil},
		nil, true, utils.NonTransactional)
	if err := rs.MaxDebit(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderMaxDebit, arg.CgrID))

	if !has {
		t.Error("has no value")
	}
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestResponderRefundIncrements(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &Account{
		ID:                "acc_id",
		BalanceMap:        map[string]Balances{},
		UnitCounters:      UnitCounters{},
		ActionTriggers:    ActionTriggers{},
		AllowNegative:     false,
		Disabled:          false,
		UpdateTime:        time.Date(2021, 12, 1, 12, 0, 0, 0, time.UTC),
		executingTriggers: false,
	}
	if err := rs.RefundIncrements(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderRefundIncrements, arg.CgrID))

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}
func TestResponderRefundIncrementsSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &Account{
		ID:                "acc_id",
		BalanceMap:        map[string]Balances{},
		UnitCounters:      UnitCounters{},
		ActionTriggers:    ActionTriggers{},
		AllowNegative:     false,
		Disabled:          false,
		UpdateTime:        time.Date(2021, 12, 1, 12, 0, 0, 0, time.UTC),
		executingTriggers: false,
	}

	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderRefundIncrements, arg.CgrID), &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}, nil, true, utils.NonTransactional)
	if err := rs.RefundIncrements(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderRefundIncrements, arg.CgrID))

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestResponderRefundRounding(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &Account{
		ID:                "acc_id",
		BalanceMap:        map[string]Balances{},
		UnitCounters:      UnitCounters{},
		ActionTriggers:    ActionTriggers{},
		AllowNegative:     false,
		Disabled:          false,
		UpdateTime:        time.Date(2021, 12, 1, 12, 0, 0, 0, time.UTC),
		executingTriggers: false,
	}
	if err := rs.RefundRounding(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderRefundRounding, arg.CgrID))

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}
func TestResponderRefundRoundingSet(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	arg := &CallDescriptorWithAPIOpts{

		CallDescriptor: &CallDescriptor{
			CgrID:       "cgrid",
			Category:    "category",
			Tenant:      "tenant",
			Subject:     "subject",
			Account:     "acount",
			Destination: "uk",
			ToR:         "tor",
			TimeStart:   time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
			TimeEnd:     time.Date(2022, 12, 1, 12, 0, 0, 0, time.UTC),
		},
		APIOpts: map[string]interface{}{
			"tor": 30 * time.Minute,
		},
	}
	reply := &Account{
		ID:                "acc_id",
		BalanceMap:        map[string]Balances{},
		UnitCounters:      UnitCounters{},
		ActionTriggers:    ActionTriggers{},
		AllowNegative:     false,
		Disabled:          false,
		UpdateTime:        time.Date(2021, 12, 1, 12, 0, 0, 0, time.UTC),
		executingTriggers: false,
	}
	Cache.Set(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderRefundRounding, arg.CgrID),
		&utils.CachedRPCResponse{Result: reply, Error: err},
		nil, true, utils.NonTransactional)

	if err := rs.RefundRounding(arg, reply); err != nil {
		t.Error(err)
	}
	exp := &utils.CachedRPCResponse{
		Result: reply,
		Error:  nil,
	}
	rcv, has := Cache.Get(utils.CacheRPCResponses, utils.ConcatenatedKey(utils.ResponderRefundRounding, arg.CgrID))

	if !has {
		t.Error("has no value")
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v,received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

}

func TestGetMaxSessionTimeOnAccountsErr(t *testing.T) {
	utils.Logger.SetLogLevel(4)
	utils.Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	defer func() {
		utils.Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	}()
	arg := &utils.GetMaxSessionTimeOnAccountsArgs{
		Subject:     "subject",
		Tenant:      "",
		Destination: "destination",
		AccountIDs:  []string{"acc_id1", "acc_id2"},
		Usage:       10 * time.Minute,
		SetupTime:   time.Date(2022, 12, 1, 1, 0, 0, 0, time.UTC),
		APIOpts:     map[string]interface{}{},
	}

	reply := &map[string]interface{}{}
	rs := &Responder{
		FilterS: &FilterS{
			cfg:     cfg,
			dm:      dm,
			connMgr: nil,
		},
	}
	expLog := ` ignoring cost for account: `
	if err := rs.GetMaxSessionTimeOnAccounts(arg, reply); err == nil || err != utils.ErrAccountNotFound {
		t.Error(err)
	}
	if rcvLog := buf.String(); !strings.Contains(rcvLog, expLog) {
		t.Errorf("logger %v doesn't contain %v", utils.ToJSON(rcvLog), utils.ToJSON(expLog))
	}

}

func TestResponderCall(t *testing.T) {
	tmpConn := connMgr
	tmp := Cache
	defer func() {
		Cache = tmp
		connMgr = tmpConn
	}()
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.CacheCfg().Partitions[utils.CacheRPCResponses].Limit = 1
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	config.SetCgrConfig(cfg)
	rs := &Responder{
		Timezone: "UTC",
		FilterS: &FilterS{
			cfg: cfg,
			dm:  dm,
		},
		MaxComputedUsage: map[string]time.Duration{},
	}
	clientConn := make(chan rpcclient.ClientConnector, 1)
	clientConn <- rs
	connMgr := NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{})
	config.SetCgrConfig(cfg)
	SetConnManager(connMgr)
	var reply CallCost
	attr := &CallDescriptorWithAPIOpts{
		CallDescriptor: &CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "dan",
			ToR:           utils.MetaAny,
			Account:       "dan",
			Destination:   "+4917621621391",
			DurationIndex: 9,
		},
	}
	if err := rs.Call(utils.ResponderGetCost, attr, &reply); err != nil {
		t.Error(err)
	}
}
