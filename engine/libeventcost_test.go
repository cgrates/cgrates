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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

//Start tests for ChargingInterval
func TestChargingIntervalPartiallyEquals(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
		},
		CompressFactor: 3,
	}
	ci2 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
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
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
			},
		},
		CompressFactor: 3,
	}
	tCi1 := ci1.Usage()
	eTCi1 := time.Duration(14*time.Second + 12*time.Millisecond)
	if *tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, *tCi1)
	}
}

func TestChargingIntervalTotalUsage(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
			},
		},
		CompressFactor: 3,
	}
	tCi1 := ci1.TotalUsage()
	eTCi1 := 3 * time.Duration(14*time.Second+12*time.Millisecond)
	if *tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, *tCi1)
	}
}

func TestChargingIntervalEventCostUsageIndex(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
			},
		},
		CompressFactor: 3,
	}
	ci1.ecUsageIdx = ci1.TotalUsage()
	tCi1 := ci1.EventCostUsageIndex()
	eTCi1 := 3 * time.Duration(14*time.Second+12*time.Millisecond)
	if *tCi1 != eTCi1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCi1, *tCi1)
	}
}

func TestChargingIntervalStartTime(t *testing.T) {
	ci1 := &ChargingInterval{
		RatingID: "Rating1",
		Increments: []*ChargingIncrement{
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
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
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
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
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
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
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
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
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           2.345,
				Usage:          time.Duration(2 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 2,
				Cost:           1.23,
				Usage:          time.Duration(5 * time.Second),
			},
			&ChargingIncrement{
				AccountingID:   "Acc1",
				CompressFactor: 3,
				Cost:           1.23,
				Usage:          time.Duration(4 * time.Millisecond),
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

//Start tests for ChargingIncrement
func TestChargingIncrementEquals(t *testing.T) {
	ch1 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          time.Duration(2 * time.Second),
	}
	ch2 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          time.Duration(2 * time.Second),
	}
	ch3 := &ChargingIncrement{
		AccountingID:   "Acc2",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          time.Duration(2 * time.Second),
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
		Usage:          time.Duration(2 * time.Second),
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
		Usage:          time.Duration(2 * time.Second),
	}
	tCh1 := ch1.TotalUsage()
	eTCh1 := time.Duration(4 * time.Second)
	if tCh1 != eTCh1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCh1, tCh1)
	}
}

func TestChargingIncrementTotalCost(t *testing.T) {
	ch1 := &ChargingIncrement{
		AccountingID:   "Acc1",
		CompressFactor: 2,
		Cost:           2.345,
		Usage:          time.Duration(2 * time.Second),
	}
	tCh1 := ch1.TotalCost()
	eTCh1 := 4.69
	if tCh1 != eTCh1 {
		t.Errorf("Expecting: %+v, received: %+v", eTCh1, tCh1)
	}
}

//Start tests for BalanceCharge
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

//Start tests for RatingMatchedFilters
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

//Start tests for ChargedTiming
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

//Start tests for RatingUnit
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

//Start tests for RatingFilters
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

//Start tests for Rating
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

//Start tests for ChargedRates
func TestChargedRatesGetIDWithSet(t *testing.T) {
	cr1 := ChargedRates{
		"Key1": RateGroups{
			&Rate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&Rate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
		"Key2": RateGroups{
			&Rate{
				GroupIntervalStart: time.Hour,
				Value:              1.12,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&Rate{
				GroupIntervalStart: 0,
				Value:              2,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
	}
	if id1 := cr1.GetIDWithSet(RateGroups{
		&Rate{
			GroupIntervalStart: time.Hour,
			Value:              0.17,
			RateIncrement:      time.Second,
			RateUnit:           time.Minute,
		},
		&Rate{
			GroupIntervalStart: time.Hour,
			Value:              0.17,
			RateIncrement:      time.Second,
			RateUnit:           time.Minute,
		},
	}); id1 != "Key1" {
		t.Errorf("Expecting: Key1, received: %+v", id1)
	}

	id2 := cr1.GetIDWithSet(RateGroups{
		&Rate{
			GroupIntervalStart: time.Hour,
			Value:              1,
			RateIncrement:      time.Second,
			RateUnit:           time.Minute,
		},
		&Rate{
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
			&Rate{
				GroupIntervalStart: time.Hour,
				Value:              0.17,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&Rate{
				GroupIntervalStart: 0,
				Value:              0.7,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
		},
		"Key2": RateGroups{
			&Rate{
				GroupIntervalStart: time.Hour,
				Value:              1.12,
				RateIncrement:      time.Second,
				RateUnit:           time.Minute,
			},
			&Rate{
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

//Start tests for ChargedTimings
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

//Start tests for Accounting
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
