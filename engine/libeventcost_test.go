/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package engine

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Start tests for ChargingInterval
func TestChargingIntervalPartiallyEquals(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
		},
		CompressFactor: 3,
	}
	ci2 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
		},
		CompressFactor: 3,
	}
	if eq := ci1.PartiallyEquals(ci2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	ci2.RatingID = "Rating2"
	if eq := ci1.PartiallyEquals(ci2); eq {
		t.Errorf("Expecting: false, received: %+v", eq)
	}
	ci2.RatingID = "Rating1"
	ci2.Increments[0].AccountingID = "Acc2"
	if eq := ci1.PartiallyEquals(ci2); eq {
		t.Errorf("Expecting: false, received: %+v", eq)
	}
}

func TestChargingIntervalUsage(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	tCi1 := ci1.Usage()
	eTCi1 := 14*time.Second + 12*time.Millisecond
	if *tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, *tCi1)
	}
}

func TestChargingIntervalTotalUsage(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	tCi1 := ci1.TotalUsage()
	eTCi1 := 3 * (14*time.Second + 12*time.Millisecond)
	if *tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, *tCi1)
	}
}

func TestChargingIntervalEventCostUsageIndex(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	ci1.ecUsageIdx = ci1.TotalUsage()
	tCi1 := ci1.EventCostUsageIndex()
	eTCi1 := 3 * (14*time.Second + 12*time.Millisecond)
	if *tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, *tCi1)
	}
}

func TestChargingIntervalStartTime(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	ci1.ecUsageIdx = ci1.TotalUsage()
	tCi1 := ci1.StartTime(time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC))
	eTCi1 := time.Date(2013, 11, 7, 8, 43, 8, 36000000, time.UTC)
	if tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, tCi1)
	}
}

func TestChargingIntervalEndTime(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	ci1.ecUsageIdx = ci1.TotalUsage()
	tCi1 := ci1.EndTime(time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC))
	eTCi1 := time.Date(2013, 11, 7, 8, 43, 8, 36000000, time.UTC)
	if tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, tCi1)
	}
}

func TestChargingIntervalCost(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	tCi1 := ci1.Cost()
	eTCi1 := 10.84
	if tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, tCi1)
	}
}

func TestChargingIntervalTotalCost(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	tCi1 := ci1.TotalCost()
	eTCi1 := 32.52
	if tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, tCi1)
	}
}

func TestChargingIntervalClone(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          2 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          5 * time.Second,
			},
			{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          4 * time.Millisecond,
			},
		},
		CompressFactor: 3,
	}
	ci2 := ci1.Clone()
	if !reflect.DeepEqual(ci1, ci2) {
		t.Errorf("Expecting: %+v, received: %+v", ci1, ci2)
	}
	ci1.RatingID = "Rating2"
	if ci2.RatingID != "Rating1" {
		t.Errorf("Expecting: Acc1, received: %+v", ci2)
	}

}

// Start tests for ChargingIncrement
func TestChargingIncrementEquals(t *testing.T) {
	ch1 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          2 * time.Second,
	}
	ch2 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          2 * time.Second,
	}
	ch3 := &ChargingIncrement{
		AccountingID:   "Acc2",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          2 * time.Second,
	}
	if eq := ch1.Equals(ch2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	if eq := ch1.Equals(ch3); eq {
		t.Errorf("Expecting: false, received: %+v", eq)
	}
}

func TestChargingIncrementClone(t *testing.T) {
	ch1 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          2 * time.Second,
	}
	ch2 := ch1.Clone()
	if !reflect.DeepEqual(ch1, ch2) {
		t.Errorf("Expecting: %+v, received: %+v", ch2, ch1)
	}
	ch1.AccountingID = "Acc2"
	if ch2.AccountingID != "Acc1" {
		t.Errorf("Expecting: Acc1, received: %+v", ch1)
	}
}

func TestChargingIncrementTotalUsage(t *testing.T) {
	ch1 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          2 * time.Second,
	}
	tCh1 := ch1.TotalUsage()
	eTCh1 := 4 * time.Second
	if tCh1 != eTCh1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCh1, tCh1)
	}
}

func TestChargingIncrementTotalCost(t *testing.T) {
	ch1 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          2 * time.Second,
	}
	tCh1 := ch1.TotalCost()
	eTCh1 := 4.69
	if tCh1 != eTCh1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCh1, tCh1)
	}
}

// Start tests for BalanceCharge
func TestBalanceChargeEquals(t *testing.T) {
	bc1 := &BalanceCharge{
		AccountID:     "1001",
		BalanceUUID:   "ASD_FGH",
		RatingID:      "Rating1001",
		Units:         2.34,
		ExtraChargeID: "Extra1",
	}
	bc2 := &BalanceCharge{
		AccountID:     "1001",
		BalanceUUID:   "ASD_FGH",
		RatingID:      "Rating1001",
		Units:         2.34,
		ExtraChargeID: "Extra1",
	}
	bc3 := &BalanceCharge{
		AccountID:     "1002",
		BalanceUUID:   "ASD_FGH",
		RatingID:      "Rating1001",
		Units:         2.34,
		ExtraChargeID: "Extra1",
	}
	if eq := bc1.Equals(bc2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	if eq := bc1.Equals(bc3); eq {
		t.Errorf("Expecting: false, received: %+v", eq)
	}
}

func TestBalanceChargeClone(t *testing.T) {
	bc1 := &BalanceCharge{
		AccountID:     "1001",
		BalanceUUID:   "ASD_FGH",
		RatingID:      "Rating1001",
		Units:         2.34,
		ExtraChargeID: "Extra1",
	}
	bc2 := bc1.Clone()
	if !reflect.DeepEqual(bc1, bc2) {
		t.Errorf("Expecting: %+v, received: %+v", bc1, bc2)
	}
	bc1.AccountID = "1002"
	if bc2.AccountID != "1001" {
		t.Errorf("Expecting: 1001, received: %+v", bc2)
	}
}

// Start tests for RatingMatchedFilters
func TestRatingMatchedFiltersEquals(t *testing.T) {
	rmf1 := RatingMatchedFilters{
		"AccountID":     "1001",
		"Units":         2.34,
		"ExtraChargeID": "Extra1",
	}
	rmf2 := RatingMatchedFilters{
		"AccountID":     "1001",
		"Units":         2.34,
		"ExtraChargeID": "Extra1",
	}
	rmf3 := RatingMatchedFilters{
		"AccountID":     "1002",
		"Units":         2.34,
		"ExtraChargeID": "Extra1",
	}
	if eq := rmf1.Equals(rmf2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	if eq := rmf1.Equals(rmf3); eq {
		t.Errorf("Expecting: false, received: %+v", eq)
	}
}

func TestRatingMatchedFiltersClone(t *testing.T) {
	rmf1 := RatingMatchedFilters{
		"AccountID":     "1001",
		"Units":         2.34,
		"ExtraChargeID": "Extra1",
	}
	rmf2 := rmf1.Clone()
	if eq := rmf1.Equals(rmf2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	rmf1["AccountID"] = "1002"
	if rmf2["AccountID"] != "1001" {
		t.Errorf("Expecting: 1001, received: %+v", rmf2)
	}
}

// Start tests for ChargedTiming
func TestChargedTimingEquals(t *testing.T) {
	ct1 := &ChargedTiming{
		Years:     utils.Years{1, 2},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{4, 5},
		WeekDays:  utils.WeekDays{2, 3},
		StartTime: "Time",
	}
	ct2 := &ChargedTiming{
		Years:     utils.Years{1, 2},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{4, 5},
		WeekDays:  utils.WeekDays{2, 3},
		StartTime: "Time",
	}
	ct3 := &ChargedTiming{
		Years:     utils.Years{2, 2},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{4, 5},
		WeekDays:  utils.WeekDays{2, 3},
		StartTime: "Time2",
	}
	if eq := ct1.Equals(ct2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	if eq := ct1.Equals(ct3); eq {
		t.Errorf("Expecting: false, received: %+v", eq)
	}
}

func TestChargedTimingClone(t *testing.T) {
	ct1 := &ChargedTiming{
		Years:     utils.Years{1, 2},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{4, 5},
		WeekDays:  utils.WeekDays{2, 3},
		StartTime: "Time",
	}
	ct2 := ct1.Clone()
	if eq := ct1.Equals(ct2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	ct1.StartTime = "Time2"
	if ct2.StartTime != "Time" {
		t.Errorf("Expecting: Time, received: %+v", ct2)
	}
}

// Start tests for RatingUnit
func TestRatingUnitEquals(t *testing.T) {
	ru1 := &RatingUnit{
		ConnectFee:       1.23,
		RoundingMethod:   "Meth1",
		RoundingDecimals: 4,
		MaxCost:          3.45,
		MaxCostStrategy:  "MaxMeth",
		TimingID:         "TimingID1",
		RatesID:          "RatesID1",
		RatingFiltersID:  "RatingFltrID1",
	}
	ru2 := &RatingUnit{
		ConnectFee:       1.23,
		RoundingMethod:   "Meth1",
		RoundingDecimals: 4,
		MaxCost:          3.45,
		MaxCostStrategy:  "MaxMeth",
		TimingID:         "TimingID1",
		RatesID:          "RatesID1",
		RatingFiltersID:  "RatingFltrID1",
	}
	ru3 := &RatingUnit{
		ConnectFee:       1.24,
		RoundingMethod:   "Meth2",
		RoundingDecimals: 4,
		MaxCost:          3.45,
		MaxCostStrategy:  "MaxMeth",
		TimingID:         "TimingID1",
		RatesID:          "RatesID1",
		RatingFiltersID:  "RatingFltrID1",
	}
	if eq := ru1.Equals(ru2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	if eq := ru1.Equals(ru3); eq {
		t.Errorf("Expecting: false, received: %+v", eq)
	}
}

func TestRatingUnitClone(t *testing.T) {
	ru1 := &RatingUnit{
		ConnectFee:       1.23,
		RoundingMethod:   "Meth1",
		RoundingDecimals: 4,
		MaxCost:          3.45,
		MaxCostStrategy:  "MaxMeth",
		TimingID:         "TimingID1",
		RatesID:          "RatesID1",
		RatingFiltersID:  "RatingFltrID1",
	}
	ru2 := ru1.Clone()
	if eq := ru1.Equals(ru2); !eq {
		t.Errorf("Expecting: true, received: %+v", eq)
	}
	ru1.ConnectFee = 2.34
	if ru2.ConnectFee != 1.23 {
		t.Errorf("Expecting: 1.23, received: %+v", ru2)
	}
}

// Start tests for RatingFilters
func TestRatingFiltersGetIDWithSet(t *testing.T) {
	rf1 := RatingFilters{
		"Key1": RatingMatchedFilters{
			"AccountID":     "1001",
			"Units":         2.34,
			"ExtraChargeID": "Extra1",
		},
		"Key2": RatingMatchedFilters{
			"AccountID":     "1002",
			"Units":         1.23,
			"ExtraChargeID": "Extra2",
		},
	}

	if id1 := rf1.GetIDWithSet(RatingMatchedFilters{
		"AccountID":     "1001",
		"Units":         2.34,
		"ExtraChargeID": "Extra1",
	}); id1 != "Key1" {
		t.Errorf("Expecting: Key1, received: %+v", id1)
	}

	if id2 := rf1.GetIDWithSet(RatingMatchedFilters{
		"AccountID":     "1004",
		"Units":         2.34,
		"ExtraChargeID": "Extra3",
	}); id2 == "" {
		t.Errorf("Expecting id , received: %+v", id2)
	}

	if id3 := rf1.GetIDWithSet(nil); id3 != "" {
		t.Errorf("Expecting , received: %+v", id3)
	}
}

func TestRatingFiltersClone(t *testing.T) {
	rf1 := RatingFilters{
		"Key1": RatingMatchedFilters{
			"AccountID":     "1001",
			"Units":         2.34,
			"ExtraChargeID": "Extra1",
		},
		"Key2": RatingMatchedFilters{
			"AccountID":     "1002",
			"Units":         1.23,
			"ExtraChargeID": "Extra2",
		},
	}
	rf2 := rf1.Clone()
	if !reflect.DeepEqual(rf1, rf2) {
		t.Errorf("Expecting: %+v, received: %+v", rf1, rf2)
	}
	rf1["Key1"]["AccountID"] = "1003"
	if rf2["Key1"]["AccountID"] != "1001" {
		t.Errorf("Expecting 1001 , received: %+v", rf2)
	}
}

// Start tests for Rating
func TestRatingGetIDWithSet(t *testing.T) {
	r1 := Rating{
		"Key1": &RatingUnit{
			ConnectFee:       1.23,
			RoundingMethod:   "Meth1",
			RoundingDecimals: 4,
			MaxCost:          3.45,
			MaxCostStrategy:  "MaxMeth",
			TimingID:         "TimingID1",
			RatesID:          "RatesID1",
			RatingFiltersID:  "RatingFltrID1",
		},
		"Key2": &RatingUnit{
			ConnectFee:       0.2,
			RoundingMethod:   "Meth1",
			RoundingDecimals: 4,
			MaxCost:          2.12,
			MaxCostStrategy:  "MaxMeth",
			TimingID:         "TimingID1",
			RatesID:          "RatesID1",
			RatingFiltersID:  "RatingFltrID1",
		},
	}

	if id1 := r1.GetIDWithSet(&RatingUnit{
		ConnectFee:       1.23,
		RoundingMethod:   "Meth1",
		RoundingDecimals: 4,
		MaxCost:          3.45,
		MaxCostStrategy:  "MaxMeth",
		TimingID:         "TimingID1",
		RatesID:          "RatesID1",
		RatingFiltersID:  "RatingFltrID1",
	}); id1 != "Key1" {
		t.Errorf("Expecting: Key1, received: %+v", id1)
	}

	if id2 := r1.GetIDWithSet(&RatingUnit{
		ConnectFee:       0.23,
		RoundingMethod:   "Meth1",
		RoundingDecimals: 4,
		MaxCost:          3.45,
		MaxCostStrategy:  "MaxMeth",
		TimingID:         "TimingID1",
		RatesID:          "RatesID1",
		RatingFiltersID:  "RatingFltrID1",
	}); id2 == "" {
		t.Errorf("Expecting id , received: %+v", id2)
	}

	if id3 := r1.GetIDWithSet(nil); id3 != "" {
		t.Errorf("Expecting , received: %+v", id3)
	}
}

func TestRatingClone(t *testing.T) {
	rf1 := Rating{
		"Key1": &RatingUnit{
			ConnectFee:       1.23,
			RoundingMethod:   "Meth1",
			RoundingDecimals: 4,
			MaxCost:          3.45,
			MaxCostStrategy:  "MaxMeth",
			TimingID:         "TimingID1",
			RatesID:          "RatesID1",
			RatingFiltersID:  "RatingFltrID1",
		},
		"Key2": &RatingUnit{
			ConnectFee:       0.2,
			RoundingMethod:   "Meth1",
			RoundingDecimals: 4,
			MaxCost:          2.12,
			MaxCostStrategy:  "MaxMeth",
			TimingID:         "TimingID1",
			RatesID:          "RatesID1",
			RatingFiltersID:  "RatingFltrID1",
		},
	}
	rf2 := rf1.Clone()
	if !reflect.DeepEqual(rf1, rf2) {
		t.Errorf("Expecting: %+v, received: %+v", rf1, rf2)
	}
	rf1["Key1"].RatesID = "RatesID2"
	if rf2["Key1"].RatesID != "RatesID1" {
		t.Errorf("Expecting RatesID1 , received: %+v", rf2)
	}
}

// Start tests for ChargedRates
func TestChargedRatesGetIDWithSet(t *testing.T) {
	cr1 := ChargedRates{
		"Key1": RateGroups{
			&RGRate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&RGRate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
		"Key2": RateGroups{
			&RGRate{
				GroupIntervalStart: time.Hour,
				Value:              1.12,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&RGRate{
				GroupIntervalStart: 0,
				Value:              2,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
	}
	if id1 := cr1.GetIDWithSet(RateGroups{
		&RGRate{
			GroupIntervalStart: time.Hour,
			Value:              0.17,
			RateIncrement:      time.Second,
			RateUnit:           time.Minute,
		},
		&RGRate{
			GroupIntervalStart: time.Hour,
			Value:              0.17,
			RateIncrement:      time.Second,
			RateUnit:           time.Minute,
		},
	}); id1 != "Key1" {
		t.Errorf("Expecting: Key1, received: %+v", id1)
	}

	id2 := cr1.GetIDWithSet(RateGroups{
		&RGRate{
			GroupIntervalStart: time.Hour,
			Value:              1,
			RateIncrement:      time.Second,
			RateUnit:           time.Minute,
		},
		&RGRate{
			GroupIntervalStart: 0,
			Value:              2,
			RateIncrement:      time.Second,
			RateUnit:           time.Minute,
		},
	})
	if id2 == "" {
		t.Errorf("Expecting id , received: %+v", id2)
	}

	if id3 := cr1.GetIDWithSet(nil); id3 != "" {
		t.Errorf("Expecting , received: %+v", id3)
	}
}

func TestChargedRatesClone(t *testing.T) {
	cr1 := ChargedRates{
		"Key1": RateGroups{
			&RGRate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&RGRate{
				GroupIntervalStart: 0,
				Value:              0.7,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
		"Key2": RateGroups{
			&RGRate{
				GroupIntervalStart: time.Hour,
				Value:              1.12,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&RGRate{
				GroupIntervalStart: 0,
				Value:              2,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
	}
	cr2 := cr1.Clone()
	if !reflect.DeepEqual(cr1, cr2) {
		t.Errorf("Expecting: %+v, received: %+v", cr1, cr2)
	}
	cr1["Key1"][0].Value = 12.2
	if cr2["Key1"][0].Value != 0.17 {
		t.Errorf("Expecting 0.17 , received: %+v", cr2)
	}
}

// Start tests for ChargedTimings
func TestChargedTimingsGetIDWithSet(t *testing.T) {
	ct1 := ChargedTimings{
		"Key1": &ChargedTiming{
			Years:     utils.Years{2, 2},
			Months:    utils.Months{2, 3},
			MonthDays: utils.MonthDays{1, 2, 3, 5},
			WeekDays:  utils.WeekDays{2, 3},
			StartTime: "Time",
		},
		"Key2": &ChargedTiming{
			Years:     utils.Years{1, 2},
			Months:    utils.Months{2, 3},
			MonthDays: utils.MonthDays{4, 5},
			WeekDays:  utils.WeekDays{2, 3},
			StartTime: "Time",
		},
	}

	if id1 := ct1.GetIDWithSet(&ChargedTiming{
		Years:     utils.Years{2, 2},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{1, 2, 3, 5},
		WeekDays:  utils.WeekDays{2, 3},
		StartTime: "Time",
	}); id1 != "Key1" {
		t.Errorf("Expecting: Key1, received: %+v", id1)
	}

	if id2 := ct1.GetIDWithSet(&ChargedTiming{
		Years:     utils.Years{1, 2, 3},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{1, 2, 3, 5},
		WeekDays:  utils.WeekDays{2, 4, 3},
		StartTime: "Time",
	}); id2 == "" {
		t.Errorf("Expecting id , received: %+v", id2)
	}

	if id3 := ct1.GetIDWithSet(nil); id3 != "" {
		t.Errorf("Expecting , received: %+v", id3)
	}
}

func TestChargedTimingsClone(t *testing.T) {
	ct1 := ChargedTimings{
		"Key1": &ChargedTiming{
			Years:     utils.Years{2, 2},
			Months:    utils.Months{2, 3},
			MonthDays: utils.MonthDays{1, 2, 3, 5},
			WeekDays:  utils.WeekDays{2, 3},
			StartTime: "Time",
		},
		"Key2": &ChargedTiming{
			Years:     utils.Years{1, 2},
			Months:    utils.Months{2, 3},
			MonthDays: utils.MonthDays{4, 5},
			WeekDays:  utils.WeekDays{2, 3},
			StartTime: "Time",
		},
	}
	ct2 := ct1.Clone()
	if !reflect.DeepEqual(ct1, ct2) {
		t.Errorf("Expecting: %+v, received: %+v", ct1, ct2)
	}
	ct1["Key1"].StartTime = "Time2"
	if ct2["Key1"].StartTime != "Time" {
		t.Errorf("Expecting Time , received: %+v", ct2)
	}
}

// Start tests for Accounting
func TestAccountingGetIDWithSet(t *testing.T) {
	a1 := Accounting{
		"Key1": &BalanceCharge{
			AccountID:     "1001",
			BalanceUUID:   "ASD_FGH",
			RatingID:      "Rating1001",
			Units:         2.34,
			ExtraChargeID: "Extra1",
		},
		"Key2": &BalanceCharge{
			AccountID:     "1002",
			BalanceUUID:   "ASD_FGH",
			RatingID:      "Rating1001",
			Units:         1.23,
			ExtraChargeID: "Extra1",
		},
	}

	if id1 := a1.GetIDWithSet(&BalanceCharge{
		AccountID:     "1001",
		BalanceUUID:   "ASD_FGH",
		RatingID:      "Rating1001",
		Units:         2.34,
		ExtraChargeID: "Extra1",
	}); id1 != "Key1" {
		t.Errorf("Expecting: Key1, received: %+v", id1)
	}

	if id2 := a1.GetIDWithSet(&BalanceCharge{
		AccountID:     "1002",
		BalanceUUID:   "ASD_FGH",
		RatingID:      "Rating1001",
		Units:         2.34,
		ExtraChargeID: "Extra1",
	}); id2 == "" {
		t.Errorf("Expecting id , received: %+v", id2)
	}

	if id3 := a1.GetIDWithSet(nil); id3 != "" {
		t.Errorf("Expecting , received: %+v", id3)
	}
}

func TestAccountingClone(t *testing.T) {
	a1 := Accounting{
		"Key1": &BalanceCharge{
			AccountID:     "1001",
			BalanceUUID:   "ASD_FGH",
			RatingID:      "Rating1001",
			Units:         2.34,
			ExtraChargeID: "Extra1",
		},
		"Key2": &BalanceCharge{
			AccountID:     "1002",
			BalanceUUID:   "ASD_FGH",
			RatingID:      "Rating1001",
			Units:         1.23,
			ExtraChargeID: "Extra1",
		},
	}
	a2 := a1.Clone()
	if !reflect.DeepEqual(a1, a2) {
		t.Errorf("Expecting: %+v, received: %+v", a1, a2)
	}
	a1["Key1"].AccountID = "1004"
	if a2["Key1"].AccountID != "1001" {
		t.Errorf("Expecting 1001 , received: %+v", a2)
	}
}

func TestChargingIncrementFieldAsInterface(t *testing.T) {
	cIt := &ChargingIncrement{
		Usage:          1 * time.Minute,
		Cost:           19,
		AccountingID:   "account_id",
		CompressFactor: 1,
	}
	if _, err := cIt.FieldAsInterface([]string{utils.Usage, utils.Cost}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = cIt.FieldAsInterface([]string{"default"}); err == nil {
		t.Error(err)
	} else if val, err := cIt.FieldAsInterface([]string{utils.Usage}); err != nil && val != cIt.Usage {
		t.Error(err)
	} else if val, err := cIt.FieldAsInterface([]string{utils.Cost}); err != nil && val != cIt.Cost {
		t.Error(err)
	} else if val, err := cIt.FieldAsInterface([]string{utils.AccountingID}); err != nil && val != cIt.AccountingID {
		t.Error(err)
	} else if val, err := cIt.FieldAsInterface([]string{utils.CompressFactor}); err != nil && val != cIt.CompressFactor {
		t.Error(err)
	}
}

func TestBalanceChargeFieldAsInterface(t *testing.T) {
	bc := &BalanceCharge{
		AccountID:     "ACC_ID",
		BalanceUUID:   "BAL_UUID",
		RatingID:      "Rating_ID",
		Units:         10.0,
		ExtraChargeID: "extra",
	}
	if _, err = bc.FieldAsInterface([]string{utils.AccountID, utils.BalanceUUID}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	} else if _, err = bc.FieldAsInterface([]string{"default"}); err == nil {
		t.Error(err)
	} else if val, err := bc.FieldAsInterface([]string{utils.AccountID}); err != nil && val != bc.AccountID {
		t.Error(err)
	} else if val, err := bc.FieldAsInterface([]string{utils.BalanceUUID}); err != nil && val != bc.BalanceUUID {
		t.Error(err)
	} else if val, err := bc.FieldAsInterface([]string{utils.RatingID}); err != nil && val != bc.RatingID {
		t.Error(err)
	} else if val, err := bc.FieldAsInterface([]string{utils.Units}); err != nil && val != bc.Units {
		t.Error(err)
	} else if val, err := bc.FieldAsInterface([]string{utils.ExtraChargeID}); err != nil && val != bc.Units {
		t.Error(err)
	}

}

func TestNewFreeEventCost(t *testing.T) {
	cgrID := "testCGRID"
	runID := "testRunID"
	account := "testAccount"
	tStart := time.Now()
	usage := 5 * time.Second
	eventCost := NewFreeEventCost(cgrID, runID, account, tStart, usage)
	if eventCost == nil {
		t.Errorf("Expected non-nil EventCost, got nil")
	}
	if eventCost.CGRID != cgrID {
		t.Errorf("Expected CGRID %v, got %v", cgrID, eventCost.CGRID)
	}
	if eventCost.RunID != runID {
		t.Errorf("Expected RunID %v, got %v", runID, eventCost.RunID)
	}
	if !eventCost.StartTime.Equal(tStart) {
		t.Errorf("Expected StartTime %v, got %v", tStart, eventCost.StartTime)
	}
	if eventCost.Cost == nil || *eventCost.Cost != 0 {
		t.Errorf("Expected Cost 0, got %v", *eventCost.Cost)
	}
	if len(eventCost.Charges) != 1 {
		t.Errorf("Expected 1 charge, got %d", len(eventCost.Charges))
	}
	charge := eventCost.Charges[0]
	if charge.RatingID != utils.MetaPause {
		t.Errorf("Expected RatingID %v, got %v", utils.MetaPause, charge.RatingID)
	}
	if len(charge.Increments) != 1 {
		t.Errorf("Expected 1 increment, got %d", len(charge.Increments))
	}
	increment := charge.Increments[0]
	if increment.Usage != usage {
		t.Errorf("Expected Usage %v, got %v", usage, increment.Usage)
	}
	if increment.AccountingID != utils.MetaPause {
		t.Errorf("Expected AccountingID %v, got %v", utils.MetaPause, increment.AccountingID)
	}
	if increment.CompressFactor != 1 {
		t.Errorf("Expected CompressFactor 1, got %v", increment.CompressFactor)
	}
	rating := eventCost.Rating[utils.MetaPause]
	if rating.RoundingMethod != "*up" {
		t.Errorf("Expected RoundingMethod *up, got %v", rating.RoundingMethod)
	}
	if rating.RoundingDecimals != 5 {
		t.Errorf("Expected RoundingDecimals 5, got %v", rating.RoundingDecimals)
	}
	if rating.RatesID != utils.MetaPause {
		t.Errorf("Expected RatesID %v, got %v", utils.MetaPause, rating.RatesID)
	}
	if rating.RatingFiltersID != utils.MetaPause {
		t.Errorf("Expected RatingFiltersID %v, got %v", utils.MetaPause, rating.RatingFiltersID)
	}
	if rating.TimingID != utils.MetaPause {
		t.Errorf("Expected TimingID %v, got %v", utils.MetaPause, rating.TimingID)
	}
	accounting := eventCost.Accounting[utils.MetaPause]
	if accounting.AccountID != account {
		t.Errorf("Expected AccountID %v, got %v", account, accounting.AccountID)
	}
	if accounting.RatingID != utils.MetaPause {
		t.Errorf("Expected Accounting RatingID %v, got %v", utils.MetaPause, accounting.RatingID)
	}
	ratingFilters := eventCost.RatingFilters[utils.MetaPause]
	if ratingFilters[utils.Subject] != "" {
		t.Errorf("Expected empty Subject, got %v", ratingFilters[utils.Subject])
	}
	if ratingFilters[utils.DestinationPrefixName] != "" {
		t.Errorf("Expected empty DestinationPrefixName, got %v", ratingFilters[utils.DestinationPrefixName])
	}
	if ratingFilters[utils.DestinationID] != "" {
		t.Errorf("Expected empty DestinationID, got %v", ratingFilters[utils.DestinationID])
	}
	if ratingFilters[utils.RatingPlanID] != utils.MetaPause {
		t.Errorf("Expected RatingPlanID %v, got %v", utils.MetaPause, ratingFilters[utils.RatingPlanID])
	}
	rates := eventCost.Rates[utils.MetaPause]
	if len(rates) != 1 {
		t.Errorf("Expected 1 rate, got %d", len(rates))
	}
	if rates[0].RateIncrement != 1 {
		t.Errorf("Expected RateIncrement 1, got %v", rates[0].RateIncrement)
	}
	if rates[0].RateUnit != 1 {
		t.Errorf("Expected RateUnit 1, got %v", rates[0].RateUnit)
	}
	timings := eventCost.Timings[utils.MetaPause]
	if timings.StartTime != "00:00:00" {
		t.Errorf("Expected StartTime 00:00:00, got %v", timings.StartTime)
	}
	if eventCost.cache == nil {
		t.Errorf("Expected non-nil cache, got nil")
	}
}

func TestIfaceAsEventCostMapStringAny(t *testing.T) {
	input := map[string]any{
		"CGRID":     "testCGRID",
		"RunID":     "testRunID",
		"StartTime": "2024-08-09T12:34:56Z",
		"Cost":      0.0,
	}
	expectedCGRID := "testCGRID"
	expectedRunID := "testRunID"
	ec, err := IfaceAsEventCost(input)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ec == nil {
		t.Errorf("Expected non-nil EventCost, got nil")
	}
	if ec.CGRID != expectedCGRID {
		t.Errorf("Expected CGRID %v, got %v", expectedCGRID, ec.CGRID)
	}
	if ec.RunID != expectedRunID {
		t.Errorf("Expected RunID %v, got %v", expectedRunID, ec.RunID)
	}
}

func TestIfaceAsEventCostDefault(t *testing.T) {
	input := 42
	_, err := IfaceAsEventCost(input)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expectedErrMsg := fmt.Sprintf("cannot convert type %v to *EventCost", reflect.TypeOf(input).String())
	if err.Error() == expectedErrMsg {
		t.Errorf("Expected error message %v, got %v", expectedErrMsg, err.Error())
	}
}

func TestFieldAsInterface(t *testing.T) {

	cts := ChargedTimings{}
	tests := []struct {
		name     string
		cts      ChargedTimings
		fldPath  []string
		expected any
		err      error
	}{
		{
			name:     "nil ChargedTimings",
			cts:      nil,
			fldPath:  []string{"field1"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "empty field path",
			cts:      cts,
			fldPath:  []string{},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "field not found",
			cts:      cts,
			fldPath:  []string{"nonexistentField"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.cts.FieldAsInterface(tt.fldPath)
			if err != tt.err {
				t.Errorf("Expected error: %v, got: %v", tt.err, err)
			}
			if val != tt.expected {
				t.Errorf("Expected value: %v, got: %v", tt.expected, val)
			}
		})
	}
}

func TestChargedRatesFieldAsInterface(t *testing.T) {
	rateGroup := []*RGRate{
		{
			GroupIntervalStart: 10 * time.Second,
			Value:              1.5,
			RateIncrement:      5 * time.Second,
			RateUnit:           60 * time.Second,
		},
		{
			GroupIntervalStart: 20 * time.Second,
			Value:              2.0,
			RateIncrement:      10 * time.Second,
			RateUnit:           60 * time.Second,
		},
	}

	crs := ChargedRates{
		"rateGroup[0]": rateGroup,
		"rateGroup":    rateGroup,
	}

	tests := []struct {
		name     string
		crs      ChargedRates
		fldPath  []string
		expected any
		err      error
	}{
		{
			name:     "nil ChargedRates",
			crs:      nil,
			fldPath:  []string{"rateGroup[0]"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "empty field path",
			crs:      crs,
			fldPath:  []string{},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "rate group with valid index",
			crs:      crs,
			fldPath:  []string{"rateGroup[0]"},
			expected: rateGroup[0],
			err:      nil,
		},
		{
			name:     "rate group with invalid index",
			crs:      crs,
			fldPath:  []string{"rateGroup[2]"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "non-existent rate group",
			crs:      crs,
			fldPath:  []string{"nonExistentGroup"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.crs.FieldAsInterface(tt.fldPath)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("Expected error: %v, got: %v", tt.err, err)
			}
			if val != tt.expected {
				t.Errorf("Expected value: %v, got: %v", tt.expected, val)
			}
		})
	}
}

func TestRatingFiltersFieldAsInterface(t *testing.T) {
	ratingFilters := RatingFilters{
		"filterA": RatingMatchedFilters{
			"subField1": "ValueA",
			"subField2": 100,
		},
		"filterB": RatingMatchedFilters{
			"subField1": "ValueB",
		},
	}

	tests := []struct {
		name     string
		rfs      RatingFilters
		fldPath  []string
		expected any
		err      error
	}{
		{
			name:     "nil RatingFilters",
			rfs:      nil,
			fldPath:  []string{"filterA"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "empty field path",
			rfs:      ratingFilters,
			fldPath:  []string{},
			expected: nil,
			err:      utils.ErrNotFound,
		},

		{
			name:     "field found with nested path",
			rfs:      ratingFilters,
			fldPath:  []string{"filterA", "subField1"},
			expected: "ValueA",
			err:      nil,
		},
		{
			name:     "field not found",
			rfs:      ratingFilters,
			fldPath:  []string{"nonExistent"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.rfs.FieldAsInterface(tt.fldPath)
			if !errors.Is(err, tt.err) {
				t.Errorf("Expected error: %v, got: %v", tt.err, err)
			}
			if val != tt.expected {
				t.Errorf("Expected value: %v, got: %v", tt.expected, val)
			}
		})
	}
}

func TestRatingMatchedFiltersFieldAsInterface(t *testing.T) {
	rmf := RatingMatchedFilters{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}

	tests := []struct {
		name     string
		rmf      RatingMatchedFilters
		fldPath  []string
		expected any
		err      error
	}{
		{
			name:     "Field exists",
			rmf:      rmf,
			fldPath:  []string{"field1"},
			expected: "value1",
			err:      nil,
		},
		{
			name:     "Field exists with integer value",
			rmf:      rmf,
			fldPath:  []string{"field2"},
			expected: 42,
			err:      nil,
		},
		{
			name:     "Field exists with boolean value",
			rmf:      rmf,
			fldPath:  []string{"field3"},
			expected: true,
			err:      nil,
		},
		{
			name:     "Field does not exist",
			rmf:      rmf,
			fldPath:  []string{"fieldNotFound"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "Invalid field path length",
			rmf:      rmf,
			fldPath:  []string{"field1", "extra"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
		{
			name:     "Nil RatingMatchedFilters",
			rmf:      nil,
			fldPath:  []string{"field1"},
			expected: nil,
			err:      utils.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.rmf.FieldAsInterface(tt.fldPath)
			if !errors.Is(err, tt.err) && (err == nil || tt.err == nil || err.Error() != tt.err.Error()) {
				t.Errorf("Expected error: %v, got: %v", tt.err, err)
			}

		})
	}
}

func TestFieldAsInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		accounting Accounting
		fldPath    []string
		expectErr  bool
	}{
		{
			name:       "Non-existent path",
			accounting: Accounting{"balance1": &BalanceCharge{AccountID: "1"}},
			fldPath:    []string{"balance2"},
			expectErr:  true,
		},
		{
			name:       "Nil value in map",
			accounting: Accounting{"balance1": nil},
			fldPath:    []string{"balance1"},
			expectErr:  true,
		},
		{
			name:       "Valid key with non-nil value",
			accounting: Accounting{"balance1": &BalanceCharge{AccountID: "1"}},
			fldPath:    []string{"balance1"},
			expectErr:  false,
		},
		{
			name:       "Nil Accounting map",
			accounting: nil,
			fldPath:    []string{"balance1"},
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.accounting.FieldAsInterface(tt.fldPath)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestBalanceCharge(t *testing.T) {
	balanceCharge := BalanceCharge{
		AccountID:     "account123",
		BalanceUUID:   "uuid",
		RatingID:      "ID1001",
		Units:         10.0,
		BalanceFactor: 0.5,
		ExtraChargeID: "1",
	}

	tests := []struct {
		path   []string
		expect interface{}
	}{
		{[]string{"AccountID"}, "account123"},
		{[]string{"BalanceUUID"}, "uuid"},
		{[]string{"RatingID"}, "ID1001"},
		{[]string{"Units"}, 10.0},
		{[]string{"BalanceFactor"}, 0.5},
		{[]string{"ExtraChargeID"}, "1"},
	}

	for _, test := range tests {
		val, err := balanceCharge.FieldAsInterface(test.path)
		if err != nil {
			t.Errorf("Unexpected error for path %v: %v", test.path, err)
			continue
		}
		if val != test.expect {
			t.Errorf("Expected %v for path %v, got %v", test.expect, test.path, val)
		}
	}

	_, err := balanceCharge.FieldAsInterface([]string{"invalid_field"})
	if err == nil {
		t.Error("Expected error for invalid field path")
	}

	_, err = balanceCharge.FieldAsInterface([]string{})
	if err == nil {
		t.Error("Expected error for empty field path")
	}

	balanceCharge.BalanceFactor = 0
	val, err := balanceCharge.FieldAsInterface([]string{"BalanceFactor"})
	if err != nil {
		t.Errorf("Unexpected error for BalanceFactor: %v", err)
	}
	if val != 1.0 {
		t.Errorf("Expected 1.0 for BalanceFactor, got %v", val)
	}
}

func TestRatingFieldAsInterface(t *testing.T) {
	rating := Rating{
		"rating1": &RatingUnit{
			ConnectFee:       1.23,
			RoundingMethod:   "ROUNDING",
			RoundingDecimals: 2,
			MaxCost:          100,
			MaxCostStrategy:  "CAP",
			TimingID:         "timing1",
			RatesID:          "rates1",
			RatingFiltersID:  "filters1",
		},
	}

	tests := []struct {
		path   []string
		expect interface{}
	}{
		{[]string{"rating1"}, rating["rating1"]},
		{[]string{"rating1", "ConnectFee"}, 1.23},
		{[]string{"rating1", "RoundingMethod"}, "ROUNDING"},
		{[]string{"rating1", "RoundingDecimals"}, 2},
		{[]string{"rating1", "MaxCost"}, 100.0},
		{[]string{"rating1", "MaxCostStrategy"}, "CAP"},
		{[]string{"rating1", "TimingID"}, "timing1"},
		{[]string{"rating1", "RatesID"}, "rates1"},
		{[]string{"rating1", "RatingFiltersID"}, "filters1"},
	}

	for _, test := range tests {
		val, err := rating.FieldAsInterface(test.path)
		if err != nil {
			t.Errorf("Unexpected error for path %v: %v", test.path, err)
			continue
		}
		if !reflect.DeepEqual(val, test.expect) {
			t.Errorf("Expected %v for path %v, got %v", test.expect, test.path, val)
		}
	}

	_, err := rating.FieldAsInterface([]string{"invalid_rating"})
	if err == nil {
		t.Error("Expected error for invalid rating ID")
	}

	_, err = rating.FieldAsInterface([]string{})
	if err == nil {
		t.Error("Expected error for empty field path")
	}

	_, err = rating.FieldAsInterface([]string{"rating1", "invalid_field"})
	if err == nil {
		t.Error("Expected error for invalid field within rating")
	}
}

func TestBalanceChargesClone(t *testing.T) {
	t.Run("Normal", func(t *testing.T) {
		original := &BalanceCharge{
			AccountID:     "account1",
			BalanceUUID:   "uuid1",
			RatingID:      "rating1",
			Units:         100.5,
			BalanceFactor: 1.25,
			ExtraChargeID: "extraCharge01",
		}
		cloned := original.Clone()
		if cloned == nil {
			t.Errorf("Expected cloned BalanceCharge to be non-nil, but got nil")
		}
		if cloned.AccountID != original.AccountID || cloned.BalanceUUID != original.BalanceUUID ||
			cloned.RatingID != original.RatingID || cloned.Units != original.Units ||
			cloned.BalanceFactor != original.BalanceFactor || cloned.ExtraChargeID != original.ExtraChargeID {
			t.Errorf("Cloned BalanceCharge does not match original. Expected %+v, but got %+v", original, cloned)
		}
		if cloned == original {
			t.Errorf("Expected original and cloned BalanceCharge to be different instances, but they are the same")
		}
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var original *BalanceCharge
		cloned := original.Clone()
		if cloned != nil {
			t.Errorf("Expected cloned BalanceCharge to be nil, but got %+v", cloned)
		}
	})
}

func TestFieldsAsInterfaces(t *testing.T) {
	ct := &ChargedTiming{
		Years:     utils.Years{1, 2},
		Months:    utils.Months{2, 3},
		MonthDays: utils.MonthDays{4, 5},
		WeekDays:  utils.WeekDays{2, 3},
		StartTime: "Time",
	}
	tests := []struct {
		name      string
		fldPath   []string
		expected  any
		expectErr bool
	}{
		{
			name:      "Years Field",
			fldPath:   []string{utils.YearsFieldName},
			expected:  ct.Years,
			expectErr: false,
		},
		{
			name:      "Months Field",
			fldPath:   []string{utils.MonthsFieldName},
			expected:  ct.Months,
			expectErr: false,
		},
		{
			name:      "MonthDays Field",
			fldPath:   []string{utils.MonthDaysFieldName},
			expected:  ct.MonthDays,
			expectErr: false,
		},
		{
			name:      "WeekDays Field",
			fldPath:   []string{utils.WeekDaysFieldName},
			expected:  ct.WeekDays,
			expectErr: false,
		},
		{
			name:      "StartTime Field",
			fldPath:   []string{utils.StartTime},
			expected:  ct.StartTime,
			expectErr: false,
		},
		{
			name:      "Invalid Field",
			fldPath:   []string{"invalidField"},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := ct.FieldAsInterface(tt.fldPath)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if !reflect.DeepEqual(val, tt.expected) {
					t.Errorf("Expected value %v, got %v", tt.expected, val)
				}
			}
		})
	}
}

func TestChargingIntervalTotalUsages(t *testing.T) {
	cIl := &ChargingInterval{
		RatingID: "RtID1",
		Increments: []*ChargingIncrement{
			{
				Usage:          2 * time.Second,
				Cost:           0.2,
				AccountingID:   "AccID1",
				CompressFactor: 1,
			},
			{
				Usage:          3 * time.Second,
				Cost:           0.3,
				AccountingID:   "AccID2",
				CompressFactor: 1,
			},
		},
		CompressFactor: 2,
		usage:          nil,
	}

	expected := 10 * time.Second

	result := cIl.TotalUsage()

	if result == nil {
		t.Errorf("Expected %v, but got nil", expected)
	} else if *result != expected {
		t.Errorf("Expected %v, but got %v", expected, *result)
	}

	manualUsage := 7 * time.Second
	cIl.usage = &manualUsage

	expected2 := 14 * time.Second
	result2 := cIl.TotalUsage()

	if result2 == nil {
		t.Errorf("Expected %v, but got nil", expected2)
	} else if *result2 != expected2 {
		t.Errorf("Expected %v, but got %v", expected2, *result2)
	}

	zeroUsage := time.Duration(0)
	cIl4 := &ChargingInterval{
		RatingID:       "RtID3",
		CompressFactor: 5,
		usage:          &zeroUsage,
	}

	expected4 := time.Duration(0)
	result3 := cIl4.TotalUsage()

	if result3 == nil {
		t.Errorf("Expected %v, but got nil", expected4)
	} else if *result3 != expected4 {
		t.Errorf("Expected %v, but got %v", expected4, *result3)
	}
}
