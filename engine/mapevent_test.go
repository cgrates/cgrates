/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var mapEv = MapEvent(map[string]interface{}{
	"test1": nil,
	"test2": 42,
	"test3": 42.3,
	"test4": true,
	"test5": "test",
	"test6": 10 * time.Second,
	"test7": "42s",
	"test8": time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
	"test9": "2009-11-10T23:00:00Z",
})

func TestMapEventNewMapEvent(t *testing.T) {
	if rply, expected := NewMapEvent(nil), MapEvent(make(map[string]interface{})); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	mp := map[string]interface{}{
		"test1": nil,
		"test2": 42,
		"test3": 42.3,
		"test4": true,
		"test5": "test",
	}
	if rply, expected := NewMapEvent(mp), MapEvent(mp); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventFieldAsInterface(t *testing.T) {
	data := utils.DataProvider(mapEv)
	if _, err := data.FieldAsInterface([]string{"first", "second"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := data.FieldAsInterface([]string{"first"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if rply, err := data.FieldAsInterface([]string{"test1"}); err != nil {
		t.Error(err)
	} else if rply != nil {
		t.Errorf("Expecting %+v, received: %+v", nil, rply)
	}
	if rply, err := data.FieldAsInterface([]string{"test4"}); err != nil {
		t.Error(err)
	} else if expected := true; rply != expected {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventFieldAsString(t *testing.T) {
	data := utils.DataProvider(mapEv)
	if _, err := data.FieldAsString([]string{"first", "second"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := data.FieldAsString([]string{"first"}); err != utils.ErrNotFound {
		t.Error(err)
	}
	if rply, err := data.FieldAsString([]string{"test1"}); err != nil {
		t.Error(err)
	} else if rply != "" {
		t.Errorf("Expecting %+v, received: %+v", "", rply)
	}
	if rply, err := data.FieldAsString([]string{"test4"}); err != nil {
		t.Error(err)
	} else if expected := "true"; rply != expected {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventRemoteHost(t *testing.T) {
	data := utils.DataProvider(mapEv)
	if rply, expected := data.RemoteHost(), utils.LocalAddr(); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventString(t *testing.T) {
	me := NewMapEvent(nil)
	if rply, expected := me.String(), utils.ToJSON(me); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	if rply, expected := mapEv.String(), utils.ToJSON(mapEv); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventHasField(t *testing.T) {
	me := NewMapEvent(nil)
	if rply := me.HasField("test1"); rply {
		t.Errorf("Expecting false, received: %+v", rply)
	}
	if rply := mapEv.HasField("test2"); !rply {
		t.Errorf("Expecting true, received: %+v", rply)
	}
	if rply := mapEv.HasField("test"); rply {
		t.Errorf("Expecting false, received: %+v", rply)
	}
}

func TestMapEventGetString(t *testing.T) {
	if rply, err := mapEv.GetString("test"); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rply != utils.EmptyString {
		t.Errorf("Expected error: %+v , received string: %+v", utils.ErrNotFound, rply)
	}
	if rply, err := mapEv.GetString("test2"); err != nil {
		t.Error(err)
	} else if rply != "42" {
		t.Errorf("Expecting %+v, received: %+v", "42", rply)
	}
	if rply, err := mapEv.GetString("test1"); err != nil {
		t.Error(err)
	} else if rply != utils.EmptyString {
		t.Errorf("Expecting , received: %+v", rply)
	}
}

func TestMapEventGetStringIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetStringIgnoreErrors("test"); rply != utils.EmptyString {
		t.Errorf("Expected: , received: %+v", rply)
	}
	if rply := mapEv.GetStringIgnoreErrors("test2"); rply != "42" {
		t.Errorf("Expecting 42, received: %+v", rply)
	}
	if rply := mapEv.GetStringIgnoreErrors("test1"); rply != utils.EmptyString {
		t.Errorf("Expecting , received: %+v", rply)
	}
}

func TestMapEventGetDuration(t *testing.T) {
	if rply, err := mapEv.GetDuration("test"); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rply != 0 {
		t.Errorf("Expected: %+v , received duration: %+v", 0, rply)
	}
	expected := 10 * time.Second
	if rply, err := mapEv.GetDuration("test6"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = 42 * time.Second
	if rply, err := mapEv.GetDuration("test7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = 42
	if rply, err := mapEv.GetDuration("test2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetDurationIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetDurationIgnoreErrors("test"); rply != 0 {
		t.Errorf("Expected: %+v, received: %+v", 0, rply)
	}
	expected := 10 * time.Second
	if rply := mapEv.GetDurationIgnoreErrors("test6"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = 42 * time.Second
	if rply := mapEv.GetDurationIgnoreErrors("test7"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = 42
	if rply := mapEv.GetDurationIgnoreErrors("test2"); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetTime(t *testing.T) {
	if rply, err := mapEv.GetTime("test", utils.EmptyString); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if !rply.IsZero() {
		t.Errorf("Expected: January 1, year 1, 00:00:00.000000000 UTC, received: %+v", rply)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	if rply, err := mapEv.GetTime("test8", utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	if rply, err := mapEv.GetTime("test9", utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetTimeIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetTimeIgnoreErrors("test", utils.EmptyString); !rply.IsZero() {
		t.Errorf("Expected: January 1, year 1, 00:00:00.000000000 UTC, received: %+v", rply)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	if rply := mapEv.GetTimeIgnoreErrors("test8", utils.EmptyString); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	if rply := mapEv.GetTimeIgnoreErrors("test9", utils.EmptyString); rply != expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
}

func TestGetTimePtr(t *testing.T) {
	rcv1, err := mapEv.GetTimePtr("test", utils.EmptyString)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rcv1 != nil {
		t.Errorf("Expected: nil, received: %+v", rcv1)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	rcv2, err := mapEv.GetTimePtr("test8", utils.EmptyString)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, *rcv2) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv2)
	}
	rcv3, err := mapEv.GetTimePtr("test9", utils.EmptyString)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, *rcv3) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv3)
	}
	if rcv1 == rcv2 || rcv2 == rcv3 || rcv1 == rcv3 {
		t.Errorf("Expecting to be different adresses")
	}
}

func TestGetTimePtrIgnoreErrors(t *testing.T) {
	rcv1 := mapEv.GetTimePtrIgnoreErrors("test", utils.EmptyString)
	if rcv1 != nil {
		t.Errorf("Expected: nil, received: %+v", rcv1)
	}
	expected := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
	rcv2 := mapEv.GetTimePtrIgnoreErrors("test8", utils.EmptyString)
	if rcv2 != nil && !reflect.DeepEqual(expected, *rcv2) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv2)
	}
	rcv3 := mapEv.GetTimePtrIgnoreErrors("test9", utils.EmptyString)
	if rcv3 != nil && !reflect.DeepEqual(expected, *rcv3) {
		t.Errorf("Expecting %+v, received: %+v", expected, rcv3)
	}
	if rcv1 == rcv2 || rcv2 == rcv3 || rcv1 == rcv3 {
		t.Errorf("Expecting to be different adresses")
	}
}

func TestMapEventClone(t *testing.T) {
	rply := mapEv.Clone()
	if !reflect.DeepEqual(mapEv, rply) {
		t.Errorf("Expecting %+v, received: %+v", mapEv, rply)
	}
	rply["test1"] = "testTest"
	if reflect.DeepEqual(mapEv, rply) {
		t.Errorf("Expecting different from: %+v, received: %+v", mapEv, rply)
	}
}

func TestMapEventAsMapString(t *testing.T) {
	expected := map[string]string{
		"test1": utils.EmptyString,
		"test2": "42",
		"test3": "42.3",
		"test4": "true",
		"test5": "test",
	}
	mpIgnore := utils.NewStringSet([]string{"test6", "test7", "test8", "test9"})

	if rply := mapEv.AsMapString(mpIgnore); !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	var mp MapEvent
	mp = nil
	if rply := mp.AsMapString(nil); !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
	if rply := mp.AsMapString(mpIgnore); !reflect.DeepEqual(map[string]string{}, rply) {
		t.Errorf("Expecting %+v, received: %+v", map[string]string{}, rply)
	}
}

func TestMapEventAsCDR(t *testing.T) {
	me := NewMapEvent(nil)
	expected := &CDR{Cost: -1.0, ExtraFields: make(map[string]string)}
	if rply, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	cfg := config.NewDefaultCGRConfig()

	expected = &CDR{
		CGRID:       "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Cost:        -1.0,
		RunID:       utils.MetaDefault,
		ToR:         utils.MetaVoice,
		RequestType: cfg.GeneralCfg().DefaultReqType,
		Tenant:      cfg.GeneralCfg().DefaultTenant,
		Category:    cfg.GeneralCfg().DefaultCategory,
		ExtraFields: make(map[string]string),
	}
	if rply, err := me.AsCDR(cfg, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	me = MapEvent{"SetupTime": "clearly not time string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"AnswerTime": "clearly not time string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"Usage": "clearly not duration string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"Partial": "clearly not bool string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"PreRated": "clearly not bool string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"Cost": "clearly not float64 string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"CostDetails": "clearly not CostDetails string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}
	me = MapEvent{"OrderID": "clearly not int64 string"}
	if _, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err == nil {
		t.Errorf("Expecting not null error, received: null error")
	}

	me = MapEvent{"ExtraField1": 5, "ExtraField2": "extra"}
	expected = &CDR{
		Cost: -1.0,
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		}}
	if rply, err := me.AsCDR(nil, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	me = MapEvent{
		"ExtraField1": 5,
		"Source":      1001,
		"CostSource":  "1002",
		"ExtraField2": "extra",
		"ExtraInfo":   "ACCOUNT_NOT_FOUND",
	}
	expected = &CDR{
		CGRID:      "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Cost:       -1.0,
		Source:     "1001",
		CostSource: "1002",
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		},
		RunID:       utils.MetaDefault,
		ToR:         utils.MetaVoice,
		RequestType: cfg.GeneralCfg().DefaultReqType,
		Tenant:      cfg.GeneralCfg().DefaultTenant,
		Category:    cfg.GeneralCfg().DefaultCategory,
		ExtraInfo:   "ACCOUNT_NOT_FOUND",
	}
	if rply, err := me.AsCDR(cfg, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	me = MapEvent{
		utils.CGRID:        "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		utils.RunID:        utils.MetaDefault,
		utils.OriginHost:   utils.FreeSWITCHAgent,
		utils.OriginID:     "127.0.0.1",
		utils.ToR:          utils.MetaVoice,
		utils.RequestType:  utils.MetaPrepaid,
		utils.Tenant:       "cgrates.org",
		utils.Category:     utils.Call,
		utils.AccountField: "10010",
		utils.Subject:      "10010",
		utils.Destination:  "10012",
		"ExtraField1":      5,
		"Source":           1001,
		"CostSource":       "1002",
		"ExtraField2":      "extra",
		"ExtraInfo":        "ACCOUNT_NOT_FOUND",
	}
	expected = &CDR{
		CGRID:       "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		RunID:       utils.MetaDefault,
		OriginHost:  utils.FreeSWITCHAgent,
		OriginID:    "127.0.0.1",
		ToR:         utils.MetaVoice,
		RequestType: utils.MetaPrepaid,
		Tenant:      "cgrates.org",
		Category:    utils.Call,
		Account:     "10010",
		Subject:     "10010",
		Destination: "10012",
		Cost:        -1.0,
		Source:      "1001",
		CostSource:  "1002",
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		},
		ExtraInfo: "ACCOUNT_NOT_FOUND",
	}
	if rply, err := me.AsCDR(cfg, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	ec1 := &EventCost{
		CGRID:     "164b0422fdc6a5117031b427439482c6a4f90e41",
		RunID:     utils.MetaDefault,
		StartTime: time.Date(2017, 1, 9, 16, 18, 21, 0, time.UTC),
		Charges: []*ChargingInterval{
			{
				RatingID: "c1a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          0,
						Cost:           0.1,
						AccountingID:   "9bdad10",
						CompressFactor: 1,
					},
					{
						Usage:          time.Second,
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          10 * time.Second,
						Cost:           0.01,
						AccountingID:   "a012888",
						CompressFactor: 2,
					},
					{
						Usage:          time.Second,
						Cost:           0.005,
						AccountingID:   "44d6c02",
						CompressFactor: 30,
					},
				},
				CompressFactor: 1,
			},
			{
				RatingID: "c1a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Second,
						Cost:           0.01,
						AccountingID:   "a012888",
						CompressFactor: 60,
					},
				},
				CompressFactor: 4,
			},
			{
				RatingID: "c1a5ab9",
				Increments: []*ChargingIncrement{
					{
						Usage:          time.Second,
						Cost:           0,
						AccountingID:   "3455b83",
						CompressFactor: 10,
					},
					{
						Usage:          10 * time.Second,
						Cost:           0.01,
						AccountingID:   "a012888",
						CompressFactor: 2,
					},
					{
						Usage:          time.Second,
						Cost:           0.005,
						AccountingID:   "44d6c02",
						CompressFactor: 30,
					},
				},
				CompressFactor: 5,
			},
		},
		AccountSummary: &AccountSummary{
			Tenant: "cgrates.org",
			ID:     "dan",
			BalanceSummaries: []*BalanceSummary{
				{
					Type:     "*monetary",
					Value:    50,
					Disabled: false},
				{
					ID:       "4b8b53d7-c1a1-4159-b845-4623a00a0165",
					Type:     "*monetary",
					Value:    25,
					Disabled: false},
				{
					Type:     "*voice",
					Value:    200,
					Disabled: false,
				},
			},
			AllowNegative: false,
			Disabled:      false,
		},
		Rating: Rating{
			"3cd6425": &RatingUnit{
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "7f324ab",
				RatesID:          "4910ecf",
				RatingFiltersID:  "43e77dc",
			},
			"c1a5ab9": &RatingUnit{
				ConnectFee:       0.1,
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				TimingID:         "7f324ab",
				RatesID:          "ec1a177",
				RatingFiltersID:  "43e77dc",
			},
		},
		Accounting: Accounting{
			"a012888": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.01,
			},
			"188bfa6": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.005,
			},
			"9bdad10": &BalanceCharge{
				AccountID:   "cgrates.org:dan",
				BalanceUUID: "8c54a9e9-d610-4c82-bcb5-a315b9a65010",
				Units:       0.1,
			},
			"44d6c02": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "7a54a9e9-d610-4c82-bcb5-a315b9a65010",
				RatingID:      "3cd6425",
				Units:         1,
				ExtraChargeID: "188bfa6",
			},
			"3455b83": &BalanceCharge{
				AccountID:     "cgrates.org:dan",
				BalanceUUID:   "9d54a9e9-d610-4c82-bcb5-a315b9a65089",
				Units:         1,
				ExtraChargeID: "*none",
			},
		},
		RatingFilters: RatingFilters{
			"43e77dc": RatingMatchedFilters{
				"DestinationID":     "GERMANY",
				"DestinationPrefix": "+49",
				"RatingPlanID":      "RPL_RETAIL1",
				"Subject":           "*out:cgrates.org:call:*any",
			},
		},
		Rates: ChargedRates{
			"ec1a177": RateGroups{
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.01,
					RateIncrement:      time.Minute,
					RateUnit:           time.Second},
			},
			"4910ecf": RateGroups{
				&RGRate{
					GroupIntervalStart: 0,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
				&RGRate{
					GroupIntervalStart: 60 * time.Second,
					Value:              0.005,
					RateIncrement:      time.Second,
					RateUnit:           time.Second},
			},
		},
		Timings: ChargedTimings{
			"7f324ab": &ChargedTiming{
				StartTime: "00:00:00",
			},
		},
	}
	ec1.initCache()
	me = MapEvent{
		"ExtraField1": 5,
		"Source":      1001,
		"CostSource":  "1002",
		"ExtraField2": "extra",
		"SetupTime":   "2009-11-10T23:00:00Z",
		"Usage":       "42s",
		"PreRated":    "True",
		"Cost":        "42.3",
		"CostDetails": utils.ToJSON(ec1),
	}
	expected = &CDR{
		CGRID:      "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		Tenant:     "itsyscom.com",
		Cost:       42.3,
		Source:     "1001",
		CostSource: "1002",
		PreRated:   true,
		Usage:      42 * time.Second,
		SetupTime:  time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC),
		ExtraFields: map[string]string{
			"ExtraField1": "5",
			"ExtraField2": "extra",
		},
		RunID:       utils.MetaDefault,
		ToR:         utils.MetaVoice,
		RequestType: cfg.GeneralCfg().DefaultReqType,
		Category:    cfg.GeneralCfg().DefaultCategory,
		CostDetails: ec1,
	}
	if rply, err := me.AsCDR(cfg, "itsyscom.com", utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetTInt64(t *testing.T) {
	if rply, err := mapEv.GetTInt64("test2"); err != nil {
		t.Error(err)
	} else if rply != int64(42) {
		t.Errorf("Expecting %+v, received: %+v", int64(42), rply)
	}

	if rply, err := mapEv.GetTInt64("test3"); err != nil {
		t.Error(err)
	} else if rply != int64(42) {
		t.Errorf("Expecting %+v, received: %+v", int64(42), rply)
	}

	if rply, err := mapEv.GetTInt64("test4"); err == nil {
		t.Errorf("Expecting error, received: %+v with error %v", rply, err)
	}

	if rply, err := mapEv.GetTInt64("0test"); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting error: %v, received: %+v with error %v", utils.ErrNotFound, rply, err)
	}
}

func TestMapEventGetFloat64(t *testing.T) {
	if rply, err := mapEv.GetFloat64("test2"); err != nil {
		t.Error(err)
	} else if rply != float64(42) {
		t.Errorf("Expecting %+v, received: %+v", float64(42), rply)
	}

	if rply, err := mapEv.GetFloat64("test3"); err != nil {
		t.Error(err)
	} else if rply != float64(42.3) {
		t.Errorf("Expecting %+v, received: %+v", float64(42.3), rply)
	}

	if rply, err := mapEv.GetFloat64("test4"); err == nil {
		t.Errorf("Expecting error, received: %+v with error %v", rply, err)
	}

	if rply, err := mapEv.GetFloat64("0test"); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expecting error: %v, received: %+v with error %v", utils.ErrNotFound, rply, err)
	}
}

func TestMapEventGetDurationPtr(t *testing.T) {
	if rply, err := mapEv.GetDurationPtr("test4"); err == nil {
		t.Errorf("Expecting error, received: %+v with error %v", rply, err)
	}
	if rply, err := mapEv.GetDurationPtr("test"); err != utils.ErrNotFound {
		t.Errorf("Expected: %+v, received: %+v", utils.ErrNotFound, err)
	} else if rply != nil {
		t.Errorf("Expected: %+v , received duration: %+v", nil, rply)
	}
	expected := utils.DurationPointer(10 * time.Second)
	if rply, err := mapEv.GetDurationPtr("test6"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42 * time.Second)
	if rply, err := mapEv.GetDurationPtr("test7"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42)
	if rply, err := mapEv.GetDurationPtr("test2"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetDurationPtrIgnoreErrors(t *testing.T) {
	if rply := mapEv.GetDurationPtrIgnoreErrors("test"); rply != nil {
		t.Errorf("Expected: %+v, received: %+v", nil, rply)
	}
	expected := utils.DurationPointer(10 * time.Second)
	if rply := mapEv.GetDurationPtrIgnoreErrors("test6"); *rply != *expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42 * time.Second)
	if rply := mapEv.GetDurationPtrIgnoreErrors("test7"); *rply != *expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
	expected = utils.DurationPointer(42)
	if rply := mapEv.GetDurationPtrIgnoreErrors("test2"); *rply != *expected {
		t.Errorf("Expected: %+v, received: %+v", expected, rply)
	}
}

func TestMapEventGetDurationPtrOrDefault(t *testing.T) {
	mapEv := NewMapEvent(nil)
	dflt := time.Nanosecond
	if ptr, _ := mapEv.GetDurationPtrOrDefault("test7", &dflt); dflt.String() != ptr.String() {
		t.Errorf("Expected: %+v, received: %+v", dflt, ptr)
	}
	newVal := 2 * time.Nanosecond
	mapEv["test7"] = newVal
	if ptr, _ := mapEv.GetDurationPtrOrDefault("test7", &dflt); newVal.String() != ptr.String() {
		t.Errorf("Expected: %+v, received: %+v", newVal, ptr)
	}
}

func TestMapEventCloneError(t *testing.T) {
	testStruct := MapEvent{}
	testStruct = nil
	exp := MapEvent{}
	exp = nil
	result := testStruct.Clone()
	if !reflect.DeepEqual(result, exp) {
		t.Errorf("Expected: %+v, received: %+v", exp, result)
	}
}

func TestMapEventData(t *testing.T) {
	testStruct := MapEvent{
		"key1": "val1",
	}
	expStruct := map[string]interface{}{
		"key1": "val1",
	}
	result := testStruct.Data()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("Expected: %+v, received: %+v", expStruct, result)
	}
}
