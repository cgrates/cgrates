/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
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
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestModelHelperCsvLoad(t *testing.T) {
	l, err := csvLoad(DestinationMdl{}, []string{"TEST_DEST", "+492"})
	tpd, ok := l.(DestinationMdl)
	if err != nil || !ok || tpd.Tag != "TEST_DEST" || tpd.Prefix != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestModelHelperCsvDump(t *testing.T) {
	tpd := DestinationMdl{
		Tag:    "TEST_DEST",
		Prefix: "+492"}
	csv, err := CsvDump(tpd)
	if err != nil || csv[0] != "TEST_DEST" || csv[1] != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestTPDestinationAsExportSlice(t *testing.T) {
	tpDst := &utils.TPDestination{
		TPid:     "TEST_TPID",
		ID:       "TEST_DEST",
		Prefixes: []string{"49", "49176", "49151"},
	}
	expectedSlc := [][]string{
		{"TEST_DEST", "49"},
		{"TEST_DEST", "49176"},
		{"TEST_DEST", "49151"},
	}
	mdst := APItoModelDestination(tpDst)
	var slc [][]string
	for _, md := range mdst {
		lc, err := CsvDump(md)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTpDestinationsAsMapDestinations(t *testing.T) {
	in := &DestinationMdls{}
	eOut := map[string]*Destination{}

	if rcv, err := in.AsMapDestinations(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	in = &DestinationMdls{
		DestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST1", Prefix: "+491"},
		DestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST2", Prefix: "+492"},
	}
	eOut = map[string]*Destination{
		"TEST_DEST1": {
			Id:       "TEST_DEST1",
			Prefixes: []string{"+491"},
		},
		"TEST_DEST2": {
			Id:       "TEST_DEST2",
			Prefixes: []string{"+492"},
		},
	}
	var rcv map[string]*Destination
	if rcv, err = in.AsMapDestinations(); err != nil {
		t.Error(err)
	}
	for key := range rcv {
		if !reflect.DeepEqual(rcv[key], eOut[key]) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut[key]), utils.ToJSON(rcv[key]))
		}
	}
	in = &DestinationMdls{
		DestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST1", Prefix: "+491"},
		DestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST2", Prefix: "+492"},
		DestinationMdl{Tpid: "TEST_ID", Tag: "", Prefix: ""},
	}
	eOut = map[string]*Destination{
		"TEST_DEST1": {
			Id:       "TEST_DEST1",
			Prefixes: []string{"+491"},
		},
		"TEST_DEST2": {
			Id:       "TEST_DEST2",
			Prefixes: []string{"+492"},
		},
		"": {
			Id:       "",
			Prefixes: []string{""},
		},
	}
	if rcv, err = in.AsMapDestinations(); err != nil {
		t.Error(err)
	}
	for key := range rcv {
		if !reflect.DeepEqual(rcv[key], eOut[key]) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut[key]), utils.ToJSON(rcv[key]))
		}
	}
}

func TestTpDestinationsAPItoModelDestination(t *testing.T) {
	d := &utils.TPDestination{}
	eOut := DestinationMdls{
		DestinationMdl{},
	}
	if rcv := APItoModelDestination(d); rcv != nil {
		if !reflect.DeepEqual(rcv, eOut) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
		}

	}
	d = &utils.TPDestination{
		TPid:     "TEST_TPID",
		ID:       "TEST_ID",
		Prefixes: []string{"+491"},
	}
	eOut = DestinationMdls{
		DestinationMdl{
			Tpid:   "TEST_TPID",
			Tag:    "TEST_ID",
			Prefix: "+491",
		},
	}
	if rcv := APItoModelDestination(d); rcv != nil {
		if !reflect.DeepEqual(rcv, eOut) {
			t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
		}
	}
}

func TestTpDestinationsAsTPDestinations(t *testing.T) {
	tpd1 := DestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+491"}
	tpd2 := DestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+492"}
	tpd3 := DestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+493"}
	eTPDestinations := []*utils.TPDestination{{TPid: "TEST_TPID", ID: "TEST_DEST",
		Prefixes: []string{"+491", "+492", "+493"}}}
	if tpDst := DestinationMdls([]DestinationMdl{tpd1, tpd2, tpd3}).AsTPDestinations(); !reflect.DeepEqual(eTPDestinations, tpDst) {
		t.Errorf("Expecting: %+v, received: %+v", eTPDestinations, tpDst)
	}

}

func TestMapTPTimings(t *testing.T) {
	var tps []*utils.ApierTPTiming
	eOut := map[string]*utils.TPTiming{}
	if rcv, err := MapTPTimings(tps); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	tps = []*utils.ApierTPTiming{
		{
			TPid: "TPid1",
			ID:   "ID1",
		},
	}
	eOut = map[string]*utils.TPTiming{
		"ID1": {
			ID:        "ID1",
			Years:     utils.Years{},
			Months:    utils.Months{},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
		},
	}
	if rcv, err := MapTPTimings(tps); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	tps = []*utils.ApierTPTiming{
		{
			TPid:   "TPid1",
			ID:     "ID1",
			Months: "1;2;3;4",
		},
	}
	eOut = map[string]*utils.TPTiming{
		"ID1": {
			ID:        "ID1",
			Years:     utils.Years{},
			Months:    utils.Months{1, 2, 3, 4},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
		},
	}
	if rcv, err := MapTPTimings(tps); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	//same id error
	tps = []*utils.ApierTPTiming{
		{
			TPid:   "TPid1",
			ID:     "ID1",
			Months: "1;2;3;4",
		},
		{
			TPid:   "TPid1",
			ID:     "ID1",
			Months: "1;2;3;4",
		},
	}
	eOut = map[string]*utils.TPTiming{
		"ID1": {
			ID:        "ID1",
			Years:     utils.Years{},
			Months:    utils.Months{1, 2, 3, 4},
			MonthDays: utils.MonthDays{},
			WeekDays:  utils.WeekDays{},
		},
	}
	if _, err := MapTPTimings(tps); err == nil || err.Error() != "duplicate timing tag: ID1" {
		t.Errorf("Expecting: nil, received: %+v", err)
	}
}

func TestMapTPRates(t *testing.T) {
	s := []*utils.TPRateRALs{}
	eOut := map[string]*utils.TPRateRALs{}
	if rcv, err := MapTPRates(s); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	s = []*utils.TPRateRALs{
		{
			ID:   "ID",
			TPid: "TPid",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.100,
					Rate:               0.200,
					RateUnit:           "60",
					RateIncrement:      "60",
					GroupIntervalStart: "0"},
				{
					ConnectFee:         0.0,
					Rate:               0.1,
					RateUnit:           "1",
					RateIncrement:      "60",
					GroupIntervalStart: "60"},
			},
		},
	}
	eOut = map[string]*utils.TPRateRALs{
		"ID": {
			TPid: "TPid",
			ID:   "ID",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee:         0.1,
					Rate:               0.2,
					RateUnit:           "60",
					RateIncrement:      "60",
					GroupIntervalStart: "0",
				}, {
					ConnectFee:         0,
					Rate:               0.1,
					RateUnit:           "1",
					RateIncrement:      "60",
					GroupIntervalStart: "60",
				},
			},
		},
	}
	if rcv, err := MapTPRates(s); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	s = []*utils.TPRateRALs{
		{
			ID:   "",
			TPid: "",
			RateSlots: []*utils.RateSlot{
				{ConnectFee: 0.8},
				{ConnectFee: 0.7},
			},
		},
	}
	eOut = map[string]*utils.TPRateRALs{
		"": {
			TPid: "",
			ID:   "",
			RateSlots: []*utils.RateSlot{
				{ConnectFee: 0.8},
				{ConnectFee: 0.7},
			},
		},
	}
	if rcv, err := MapTPRates(s); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	s = []*utils.TPRateRALs{
		{
			ID:   "SameID",
			TPid: "",
			RateSlots: []*utils.RateSlot{
				{ConnectFee: 0.8},
				{ConnectFee: 0.7},
			},
		},
		{
			ID:   "SameID",
			TPid: "",
			RateSlots: []*utils.RateSlot{
				{ConnectFee: 0.9},
				{ConnectFee: 0.1},
			},
		},
	}
	if _, err := MapTPRates(s); err == nil || err.Error() != "Non unique ID SameID" {
		t.Error(err)
	}
}

func TestAPItoModelTimings(t *testing.T) {
	ts := []*utils.ApierTPTiming{}
	eOut := TimingMdls{}
	if rcv := APItoModelTimings(ts); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", utils.ToJSON(rcv))
	}

	ts = []*utils.ApierTPTiming{
		{
			TPid:   "TPid1",
			ID:     "ID1",
			Months: "1;2;3;4",
		},
	}
	eOut = TimingMdls{
		TimingMdl{
			Tpid:   "TPid1",
			Months: "1;2;3;4",
			Tag:    "ID1",
		},
	}
	if rcv := APItoModelTimings(ts); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	ts = []*utils.ApierTPTiming{
		{
			TPid:   "TPid1",
			ID:     "ID1",
			Months: "1;2;3;4",
		},
		{
			TPid:      "TPid2",
			ID:        "ID2",
			Months:    "1;2;3;4",
			MonthDays: "1;2;3;4;28",
			Years:     "2020;2019",
			WeekDays:  "4;5",
		},
	}
	eOut = TimingMdls{
		TimingMdl{
			Tpid:   "TPid1",
			Months: "1;2;3;4",
			Tag:    "ID1",
		},
		TimingMdl{
			Tpid:      "TPid2",
			Tag:       "ID2",
			Months:    "1;2;3;4",
			MonthDays: "1;2;3;4;28",
			Years:     "2020;2019",
			WeekDays:  "4;5",
		},
	}
	if rcv := APItoModelTimings(ts); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPRateAsExportSlice(t *testing.T) {
	tpRate := &utils.TPRateRALs{
		TPid: "TEST_TPID",
		ID:   "TEST_RATEID",
		RateSlots: []*utils.RateSlot{
			{
				ConnectFee:         0.100,
				Rate:               0.200,
				RateUnit:           "60",
				RateIncrement:      "60",
				GroupIntervalStart: "0"},
			{
				ConnectFee:         0.0,
				Rate:               0.1,
				RateUnit:           "1",
				RateIncrement:      "60",
				GroupIntervalStart: "60"},
		},
	}
	expectedSlc := [][]string{
		{"TEST_RATEID", "0.1", "0.2", "60", "60", "0"},
		{"TEST_RATEID", "0", "0.1", "1", "60", "60"},
	}

	ms := APItoModelRate(tpRate)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc[0], slc[0])
	}
}

func TestAPItoModelRates(t *testing.T) {
	rs := []*utils.TPRateRALs{}
	eOut := RateMdls{}
	if rcv := APItoModelRates(rs); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", utils.ToJSON(rcv))
	}

	rs = []*utils.TPRateRALs{
		{
			ID:   "SomeID",
			TPid: "TPid",
			RateSlots: []*utils.RateSlot{
				{
					ConnectFee: 0.7,
					Rate:       0.8,
				},
				{
					ConnectFee: 0.77,
					Rate:       0.88,
				},
			},
		},
	}
	eOut = RateMdls{
		RateMdl{
			Tpid:       "TPid",
			Tag:        "SomeID",
			ConnectFee: 0.7,
			Rate:       0.8,
		},
		RateMdl{
			Tpid:       "TPid",
			Tag:        "SomeID",
			ConnectFee: 0.77,
			Rate:       0.88,
		},
	}
	if rcv := APItoModelRates(rs); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	rs = []*utils.TPRateRALs{
		{
			ID:        "SomeID",
			TPid:      "TPid",
			RateSlots: []*utils.RateSlot{},
		},
	}
	eOut = RateMdls{
		RateMdl{
			Tpid: "TPid",
			Tag:  "SomeID",
		},
	}
	if rcv := APItoModelRates(rs); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPDestinationRateAsExportSlice(t *testing.T) {
	tpDstRate := &utils.TPDestinationRate{
		TPid: "TEST_TPID",
		ID:   "TEST_DSTRATE",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "TEST_DEST1",
				RateId:           "TEST_RATE1",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
			{
				DestinationId:    "TEST_DEST2",
				RateId:           "TEST_RATE2",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
		},
	}
	expectedSlc := [][]string{
		{"TEST_DSTRATE", "TEST_DEST1", "TEST_RATE1", "*up", "4", "0", ""},
		{"TEST_DSTRATE", "TEST_DEST2", "TEST_RATE2", "*up", "4", "0", ""},
	}
	ms := APItoModelDestinationRate(tpDstRate)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}

	tpDstRate = &utils.TPDestinationRate{
		TPid:             "TEST_TPID",
		ID:               "TEST_DSTRATE",
		DestinationRates: []*utils.DestinationRate{},
	}
	eOut := DestinationRateMdls{
		DestinationRateMdl{
			Tpid: "TEST_TPID",
			Tag:  "TEST_DSTRATE",
		},
	}
	if rcv := APItoModelDestinationRate(tpDstRate); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)

	}
}

func TestAPItoModelDestinationRates(t *testing.T) {
	var drs []*utils.TPDestinationRate
	if rcv := APItoModelDestinationRates(drs); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	drs = []*utils.TPDestinationRate{
		{
			TPid: "TEST_TPID",
			ID:   "TEST_DSTRATE",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "TEST_DEST1",
					RateId:           "TEST_RATE1",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
				{
					DestinationId:    "TEST_DEST2",
					RateId:           "TEST_RATE2",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
			},
		},
	}
	eOut := DestinationRateMdls{
		DestinationRateMdl{
			Tpid:             "TEST_TPID",
			Tag:              "TEST_DSTRATE",
			DestinationsTag:  "TEST_DEST1",
			RatesTag:         "TEST_RATE1",
			RoundingMethod:   "*up",
			RoundingDecimals: 4,
		},
		DestinationRateMdl{
			Tpid:             "TEST_TPID",
			Tag:              "TEST_DSTRATE",
			DestinationsTag:  "TEST_DEST2",
			RatesTag:         "TEST_RATE2",
			RoundingMethod:   "*up",
			RoundingDecimals: 4,
		},
	}
	if rcv := APItoModelDestinationRates(drs); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestTpDestinationRatesAsTPDestinationRates(t *testing.T) {
	pts := DestinationRateMdls{}
	eOut := []*utils.TPDestinationRate{}
	rcv := pts.AsTPDestinationRates()
	if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", utils.ToJSON(rcv))
	}

	pts = DestinationRateMdls{
		DestinationRateMdl{
			Id:               66,
			Tpid:             "Tpid",
			Tag:              "Tag",
			DestinationsTag:  "DestinationsTag",
			RatesTag:         "RatesTag",
			RoundingMethod:   "*up",
			RoundingDecimals: 2,
			MaxCost:          0.7,
			MaxCostStrategy:  "*free",
		},
	}
	eOut = []*utils.TPDestinationRate{
		{
			TPid: "Tpid",
			ID:   "Tag",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "DestinationsTag",
					RateId:           "RatesTag",
					RoundingMethod:   "*up",
					RoundingDecimals: 2,
					MaxCost:          0.7,
					MaxCostStrategy:  "*free",
				},
			},
		},
	}
	rcv = pts.AsTPDestinationRates()
	if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestMapTPDestinationRates(t *testing.T) {
	var s []*utils.TPDestinationRate
	eOut := map[string]*utils.TPDestinationRate{}
	if rcv, err := MapTPDestinationRates(s); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	s = []*utils.TPDestinationRate{
		{
			TPid: "TEST_TPID",
			ID:   "TEST_DSTRATE",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "TEST_DEST1",
					RateId:           "TEST_RATE1",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
				{
					DestinationId:    "TEST_DEST2",
					RateId:           "TEST_RATE2",
					RoundingMethod:   "*up",
					RoundingDecimals: 4},
			},
		},
	}
	eOut = map[string]*utils.TPDestinationRate{
		"TEST_DSTRATE": {
			TPid: "TEST_TPID",
			ID:   "TEST_DSTRATE",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "TEST_DEST1",
					RateId:           "TEST_RATE1",
					RoundingMethod:   "*up",
					RoundingDecimals: 4,
				},
				{
					DestinationId:    "TEST_DEST2",
					RateId:           "TEST_RATE2",
					RoundingMethod:   "*up",
					RoundingDecimals: 4,
				},
			},
		},
	}
	if rcv, err := MapTPDestinationRates(s); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	s = []*utils.TPDestinationRate{
		{
			TPid:             "TEST_TPID",
			ID:               "TEST_DSTRATE",
			DestinationRates: []*utils.DestinationRate{},
		},
		{
			TPid:             "TEST_TPID",
			ID:               "TEST_DSTRATE",
			DestinationRates: []*utils.DestinationRate{},
		},
	}
	if rcv, err := MapTPDestinationRates(s); err == nil || err.Error() != "Non unique ID TEST_DSTRATE" {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}

}

func TestApierTPTimingAsExportSlice(t *testing.T) {
	tpTiming := &utils.ApierTPTiming{
		TPid:      "TEST_TPID",
		ID:        "TEST_TIMING",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;4",
		Time:      "00:00:01"}
	expectedSlc := [][]string{
		{"TEST_TIMING", "*any", "*any", "*any", "1;2;4", "00:00:01"},
	}
	ms := APItoModelTiming(tpTiming)
	var slc [][]string

	lc, err := CsvDump(ms)
	if err != nil {
		t.Error("Error dumping to csv: ", err)
	}
	slc = append(slc, lc)

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPRatingPlanAsExportSlice(t *testing.T) {
	tpRpln := &utils.TPRatingPlan{
		TPid: "TEST_TPID",
		ID:   "TEST_RPLAN",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{
				DestinationRatesId: "TEST_DSTRATE1",
				TimingId:           "TEST_TIMING1",
				Weight:             10.0},
			{
				DestinationRatesId: "TEST_DSTRATE2",
				TimingId:           "TEST_TIMING2",
				Weight:             20.0},
		}}
	expectedSlc := [][]string{
		{"TEST_RPLAN", "TEST_DSTRATE1", "TEST_TIMING1", "10"},
		{"TEST_RPLAN", "TEST_DSTRATE2", "TEST_TIMING2", "20"},
	}

	ms := APItoModelRatingPlan(tpRpln)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestAPItoModelRatingPlan(t *testing.T) {
	rp := &utils.TPRatingPlan{}
	eOut := RatingPlanMdls{RatingPlanMdl{}}
	if rcv := APItoModelRatingPlan(rp); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	rp = &utils.TPRatingPlan{
		TPid: "TEST_TPID",
		ID:   "TEST_RPLAN",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{
				DestinationRatesId: "TEST_DSTRATE1",
				TimingId:           "TEST_TIMING1",
				Weight:             10.0},
			{
				DestinationRatesId: "TEST_DSTRATE2",
				TimingId:           "TEST_TIMING2",
				Weight:             20.0},
		}}

	eOut = RatingPlanMdls{
		RatingPlanMdl{
			Tpid:         "TEST_TPID",
			Tag:          "TEST_RPLAN",
			DestratesTag: "TEST_DSTRATE1",
			TimingTag:    "TEST_TIMING1",
			Weight:       10.0,
		},
		RatingPlanMdl{
			Tpid:         "TEST_TPID",
			Tag:          "TEST_RPLAN",
			DestratesTag: "TEST_DSTRATE2",
			TimingTag:    "TEST_TIMING2",
			Weight:       20.0,
		},
	}
	if rcv := APItoModelRatingPlan(rp); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	rp = &utils.TPRatingPlan{
		TPid: "TEST_TPID",
		ID:   "TEST_RPLAN",
	}
	eOut = RatingPlanMdls{
		RatingPlanMdl{
			Tpid: "TEST_TPID",
			Tag:  "TEST_RPLAN",
		},
	}
	if rcv := APItoModelRatingPlan(rp); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoModelRatingPlans(t *testing.T) {
	var rps []*utils.TPRatingPlan
	if rcv := APItoModelRatingPlans(rps); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", utils.ToJSON(rcv))
	}
	rps = []*utils.TPRatingPlan{
		{
			ID:   "ID",
			TPid: "TPid",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "DestinationRatesId",
					TimingId:           "TimingId",
					Weight:             0.7,
				},
			},
		},
	}
	eOut := RatingPlanMdls{
		RatingPlanMdl{
			Tag:          "ID",
			Tpid:         "TPid",
			DestratesTag: "DestinationRatesId",
			TimingTag:    "TimingId",
			Weight:       0.7,
		},
	}
	if rcv := APItoModelRatingPlans(rps); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestTPRatingProfileAsExportSlice(t *testing.T) {
	tpRpf := &utils.TPRatingProfile{
		TPid:     "TEST_TPID",
		LoadId:   "TEST_LOADID",
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "*any",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime:   "2014-01-14T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN1",
				FallbackSubjects: "subj1;subj2"},
			{
				ActivationTime:   "2014-01-15T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN2",
				FallbackSubjects: "subj1;subj2"},
		},
	}
	expectedSlc := [][]string{
		{"cgrates.org", "call", "*any", "2014-01-14T00:00:00Z", "TEST_RPLAN1", "subj1;subj2"},
		{"cgrates.org", "call", "*any", "2014-01-15T00:00:00Z", "TEST_RPLAN2", "subj1;subj2"},
	}

	ms := APItoModelRatingProfile(tpRpf)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestAPItoModelRatingProfile(t *testing.T) {
	var rp *utils.TPRatingProfile
	if rcv := APItoModelRatingProfile(rp); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	rp = &utils.TPRatingProfile{
		TPid:     "TEST_TPID",
		LoadId:   "TEST_LOADID",
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "*any",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime:   "2014-01-14T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN1",
				FallbackSubjects: "subj1;subj2"},
			{
				ActivationTime:   "2014-01-15T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN2",
				FallbackSubjects: "subj1;subj2"},
		},
	}
	eOut := RatingProfileMdls{
		RatingProfileMdl{
			Tpid:             "TEST_TPID",
			Loadid:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Category:         "call",
			Subject:          "*any",
			RatingPlanTag:    "TEST_RPLAN1",
			FallbackSubjects: "subj1;subj2",
			ActivationTime:   "2014-01-14T00:00:00Z",
		},
		RatingProfileMdl{
			Tpid:             "TEST_TPID",
			Loadid:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Category:         "call",
			Subject:          "*any",
			RatingPlanTag:    "TEST_RPLAN2",
			FallbackSubjects: "subj1;subj2",
			ActivationTime:   "2014-01-15T00:00:00Z",
		},
	}
	if rcv := APItoModelRatingProfile(rp); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	rp = &utils.TPRatingProfile{
		TPid:     "TEST_TPID",
		LoadId:   "TEST_LOADID",
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "*any",
	}
	eOut = RatingProfileMdls{
		RatingProfileMdl{
			Tpid:     "TEST_TPID",
			Loadid:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "*any",
		},
	}
	if rcv := APItoModelRatingProfile(rp); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoModelRatingProfiles(t *testing.T) {
	var rps []*utils.TPRatingProfile
	if rcv := APItoModelRatingProfiles(rps); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	rps = []*utils.TPRatingProfile{
		{
			TPid:     "TEST_TPID",
			LoadId:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "*any",
		},
		{
			TPid:     "TEST_TPID2",
			LoadId:   "TEST_LOADID2",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "*any",
		},
	}
	eOut := RatingProfileMdls{
		RatingProfileMdl{
			Tpid:     "TEST_TPID",
			Loadid:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "*any",
		},
		RatingProfileMdl{
			Tpid:     "TEST_TPID2",
			Loadid:   "TEST_LOADID2",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "*any",
		},
	}
	if rcv := APItoModelRatingProfiles(rps); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPActionsAsExportSlice(t *testing.T) {
	tpActs := &utils.TPActions{
		TPid: "TEST_TPID",
		ID:   "TEST_ACTIONS",
		Actions: []*utils.TPAction{
			{
				Identifier:      "*topup_reset",
				BalanceType:     "*monetary",
				Units:           "5.0",
				ExpiryTime:      "*never",
				DestinationIds:  "*any",
				RatingSubject:   "special1",
				Categories:      "call",
				SharedGroups:    "GROUP1",
				BalanceWeight:   "10.0",
				ExtraParameters: "",
				Weight:          10.0},
			{
				Identifier:      "*http_post",
				BalanceType:     "",
				Units:           "0.0",
				ExpiryTime:      "",
				DestinationIds:  "",
				RatingSubject:   "",
				Categories:      "",
				SharedGroups:    "",
				BalanceWeight:   "0.0",
				ExtraParameters: "http://localhost/&param1=value1",
				Weight:          20.0},
		},
	}
	expectedSlc := [][]string{
		{"TEST_ACTIONS", "*topup_reset", "", "", "", "*monetary", "call", "*any", "special1", "GROUP1", "*never", "", "5.0", "10.0", "", "", "10"},
		{"TEST_ACTIONS", "*http_post", "http://localhost/&param1=value1", "", "", "", "", "", "", "", "", "", "0.0", "0.0", "", "", "20"},
	}

	ms := APItoModelAction(tpActs)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: \n%+v, received: \n%+v", expectedSlc, slc)
	}
}

// SHARED_A,*any,*highest,
func TestTPSharedGroupsAsExportSlice(t *testing.T) {
	tpSGs := &utils.TPSharedGroups{
		TPid: "TEST_TPID",
		ID:   "SHARED_GROUP_TEST",
		SharedGroups: []*utils.TPSharedGroup{
			{
				Account:       "*any",
				Strategy:      "*highest",
				RatingSubject: "special1"},
			{
				Account:       "second",
				Strategy:      "*highest",
				RatingSubject: "special2"},
		},
	}
	expectedSlc := [][]string{
		{"SHARED_GROUP_TEST", "*any", "*highest", "special1"},
		{"SHARED_GROUP_TEST", "second", "*highest", "special2"},
	}

	ms := APItoModelSharedGroup(tpSGs)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestAPItoModelSharedGroups(t *testing.T) {
	sgs := []*utils.TPSharedGroups{}
	if rcv := APItoModelSharedGroups(sgs); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	sgs = []*utils.TPSharedGroups{
		{
			TPid: "TEST_TPID",
			ID:   "SHARED_GROUP_TEST",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "*any",
					Strategy:      "*highest",
					RatingSubject: "special1"},
				{
					Account:       "*second",
					Strategy:      "*highest",
					RatingSubject: "special2"},
			},
		},
	}
	eOut := SharedGroupMdls{
		SharedGroupMdl{
			Tpid:          "TEST_TPID",
			Tag:           "SHARED_GROUP_TEST",
			Account:       "*any",
			Strategy:      "*highest",
			RatingSubject: "special1",
		},
		SharedGroupMdl{
			Tpid:          "TEST_TPID",
			Tag:           "SHARED_GROUP_TEST",
			Account:       "*second",
			Strategy:      "*highest",
			RatingSubject: "special2",
		},
	}
	if rcv := APItoModelSharedGroups(sgs); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	sgs = []*utils.TPSharedGroups{
		{
			TPid: "TEST_TPID",
			ID:   "SHARED_GROUP_TEST",
		},
	}
	eOut = SharedGroupMdls{
		SharedGroupMdl{
			Tpid: "TEST_TPID",
			Tag:  "SHARED_GROUP_TEST",
		},
	}
	if rcv := APItoModelSharedGroups(sgs); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	sgs = []*utils.TPSharedGroups{
		{
			TPid: "TEST_TPID",
			ID:   "SHARED_GROUP_TEST",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "*any",
					Strategy:      "*highest",
					RatingSubject: "special1"},
				{
					Account:       "*second",
					Strategy:      "*highest",
					RatingSubject: "special2"},
			},
		},
		{
			TPid: "TEST_TPID2",
			ID:   "SHARED_GROUP_TEST2",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "*any",
					Strategy:      "*highest",
					RatingSubject: "special1"},
				{
					Account:       "second",
					Strategy:      "*highest",
					RatingSubject: "special2"},
			},
		},
	}
	eOut = SharedGroupMdls{
		SharedGroupMdl{
			Tpid:          "TEST_TPID",
			Tag:           "SHARED_GROUP_TEST",
			Account:       "*any",
			Strategy:      "*highest",
			RatingSubject: "special1",
		},
		SharedGroupMdl{
			Tpid:          "TEST_TPID",
			Tag:           "SHARED_GROUP_TEST",
			Account:       "*second",
			Strategy:      "*highest",
			RatingSubject: "special2",
		},
		SharedGroupMdl{
			Tpid:          "TEST_TPID2",
			Tag:           "SHARED_GROUP_TEST2",
			Account:       "*any",
			Strategy:      "*highest",
			RatingSubject: "special1",
		},
		SharedGroupMdl{
			Tpid:          "TEST_TPID2",
			Tag:           "SHARED_GROUP_TEST2",
			Account:       "second",
			Strategy:      "*highest",
			RatingSubject: "special2",
		},
	}
	if rcv := APItoModelSharedGroups(sgs); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPActionTriggersAsExportSlice(t *testing.T) {
	ap := &utils.TPActionPlan{
		TPid: "TEST_TPID",
		ID:   "PACKAGE_10",
		ActionPlan: []*utils.TPActionTiming{
			{
				ActionsId: "TOPUP_RST_10",
				TimingId:  "ASAP",
				Weight:    10.0},
			{
				ActionsId: "TOPUP_RST_5",
				TimingId:  "ASAP",
				Weight:    20.0},
		},
	}
	expectedSlc := [][]string{
		{"PACKAGE_10", "TOPUP_RST_10", "ASAP", "10"},
		{"PACKAGE_10", "TOPUP_RST_5", "ASAP", "20"},
	}
	ms := APItoModelActionPlan(ap)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestAPItoModelActionPlan(t *testing.T) {
	var a *utils.TPActionPlan
	if rcv := APItoModelActionPlan(a); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	a = &utils.TPActionPlan{
		TPid: "TEST_TPID",
		ID:   "PACKAGE_10",
		ActionPlan: []*utils.TPActionTiming{
			{
				ActionsId: "TOPUP_RST_10",
				TimingId:  "ASAP",
				Weight:    10.0},
			{
				ActionsId: "TOPUP_RST_5",
				TimingId:  "ASAP",
				Weight:    20.0},
		},
	}

	eOut := ActionPlanMdls{
		ActionPlanMdl{
			Tpid:       "TEST_TPID",
			Tag:        "PACKAGE_10",
			ActionsTag: "TOPUP_RST_10",
			TimingTag:  "ASAP",
			Weight:     10,
		},
		ActionPlanMdl{
			Tpid:       "TEST_TPID",
			Tag:        "PACKAGE_10",
			ActionsTag: "TOPUP_RST_5",
			TimingTag:  "ASAP",
			Weight:     20,
		},
	}
	if rcv := APItoModelActionPlan(a); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	a = &utils.TPActionPlan{
		TPid: "TEST_TPID",
		ID:   "PACKAGE_10",
	}
	eOut = ActionPlanMdls{
		ActionPlanMdl{
			Tpid: "TEST_TPID",
			Tag:  "PACKAGE_10",
		},
	}
	if rcv := APItoModelActionPlan(a); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoModelActionPlans(t *testing.T) {
	var a []*utils.TPActionPlan
	if rcv := APItoModelActionPlans(a); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	a = []*utils.TPActionPlan{
		{
			TPid: "TEST_TPID",
			ID:   "PACKAGE_10",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "TOPUP_RST_10",
					TimingId:  "ASAP",
					Weight:    10.0},
				{
					ActionsId: "TOPUP_RST_5",
					TimingId:  "ASAP",
					Weight:    20.0},
			},
		},
	}
	eOut := ActionPlanMdls{
		ActionPlanMdl{
			Tpid:       "TEST_TPID",
			Tag:        "PACKAGE_10",
			ActionsTag: "TOPUP_RST_10",
			TimingTag:  "ASAP",
			Weight:     10,
		},
		ActionPlanMdl{
			Tpid:       "TEST_TPID",
			Tag:        "PACKAGE_10",
			ActionsTag: "TOPUP_RST_5",
			TimingTag:  "ASAP",
			Weight:     20,
		},
	}
	if rcv := APItoModelActionPlans(a); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	a = []*utils.TPActionPlan{
		{
			TPid: "TEST_TPID",
			ID:   "PACKAGE_10",
		},
	}
	eOut = ActionPlanMdls{
		ActionPlanMdl{
			Tpid: "TEST_TPID",
			Tag:  "PACKAGE_10",
		},
	}
	if rcv := APItoModelActionPlans(a); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPActionPlanAsExportSlice(t *testing.T) {
	at := &utils.TPActionTriggers{
		TPid: "TEST_TPID",
		ID:   "STANDARD_TRIGGERS",
		ActionTriggers: []*utils.TPActionTrigger{
			{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "1",
				ThresholdType:         "*min_balance",
				ThresholdValue:        2.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b1",
				BalanceType:           "*monetary",
				BalanceDestinationIds: "",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				ActionsId:             "LOG_WARNING",
				Weight:                10},
			{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "2",
				ThresholdType:         "*max_event_counter",
				ThresholdValue:        5.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b2",
				BalanceType:           "*monetary",
				BalanceDestinationIds: "FS_USERS",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				ActionsId:             "LOG_WARNING",
				Weight:                10},
		},
	}
	expectedSlc := [][]string{
		{"STANDARD_TRIGGERS", "1", "*min_balance", "2", "false", "0", "", "", "b1", "*monetary", "call", "", "special1", "SHARED_1", "*never", "T1", "0.0", "false", "false", "LOG_WARNING", "10"},
		{"STANDARD_TRIGGERS", "2", "*max_event_counter", "5", "false", "0", "", "", "b2", "*monetary", "call", "FS_USERS", "special1", "SHARED_1", "*never", "T1", "0.0", "false", "false", "LOG_WARNING", "10"},
	}
	ms := APItoModelActionTrigger(at)
	var slc [][]string
	for _, m := range ms {
		lc, err := CsvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestAPItoModelActionTrigger(t *testing.T) {
	var at *utils.TPActionTriggers
	if rcv := APItoModelActionTrigger(at); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}

	at = &utils.TPActionTriggers{
		TPid: "TEST_TPID",
		ID:   "STANDARD_TRIGGERS",
		ActionTriggers: []*utils.TPActionTrigger{
			{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "1",
				ThresholdType:         "*min_balance",
				ThresholdValue:        2.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b1",
				BalanceType:           "*monetary",
				BalanceDestinationIds: "",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				ActionsId:             "LOG_WARNING",
				Weight:                10},
			{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "2",
				ThresholdType:         "*max_event_counter",
				ThresholdValue:        5.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b2",
				BalanceType:           "*monetary",
				BalanceDestinationIds: "FS_USERS",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				ActionsId:             "LOG_WARNING",
				Weight:                10},
		},
	}
	eOut := ActionTriggerMdls{
		ActionTriggerMdl{
			Tpid:                 "TEST_TPID",
			Tag:                  "STANDARD_TRIGGERS",
			UniqueId:             "1",
			ThresholdType:        "*min_balance",
			ThresholdValue:       2,
			MinSleep:             "0",
			BalanceTag:           "b1",
			BalanceType:          "*monetary",
			BalanceCategories:    "call",
			BalanceRatingSubject: "special1",
			BalanceSharedGroups:  "SHARED_1",
			BalanceExpiryTime:    "*never",
			BalanceTimingTags:    "T1",
			BalanceWeight:        "0.0",
			BalanceBlocker:       "false",
			BalanceDisabled:      "false",
			ActionsTag:           "LOG_WARNING",
			Weight:               10,
		},
		ActionTriggerMdl{
			Tpid:                   "TEST_TPID",
			Tag:                    "STANDARD_TRIGGERS",
			UniqueId:               "2",
			ThresholdType:          "*max_event_counter",
			ThresholdValue:         5,
			MinSleep:               "0",
			BalanceTag:             "b2",
			BalanceType:            "*monetary",
			BalanceCategories:      "call",
			BalanceDestinationTags: "FS_USERS",
			BalanceRatingSubject:   "special1",
			BalanceSharedGroups:    "SHARED_1",
			BalanceExpiryTime:      "*never",
			BalanceTimingTags:      "T1",
			BalanceWeight:          "0.0",
			BalanceBlocker:         "false",
			BalanceDisabled:        "false",
			ActionsTag:             "LOG_WARNING",
			Weight:                 10,
		},
	}
	if rcv := APItoModelActionTrigger(at); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	at = &utils.TPActionTriggers{
		TPid: "TEST_TPID",
		ID:   "STANDARD_TRIGGERS",
	}
	eOut = ActionTriggerMdls{
		ActionTriggerMdl{
			Tpid: "TEST_TPID",
			Tag:  "STANDARD_TRIGGERS",
		},
	}
	if rcv := APItoModelActionTrigger(at); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPAccountActionsAsExportSlice(t *testing.T) {
	aa := &utils.TPAccountActions{
		TPid:             "TEST_TPID",
		LoadId:           "TEST_LOADID",
		Tenant:           "cgrates.org",
		Account:          "1001",
		ActionPlanId:     "PACKAGE_10_SHARED_A_5",
		ActionTriggersId: "STANDARD_TRIGGERS",
	}
	expectedSlc := [][]string{
		{"cgrates.org", "1001", "PACKAGE_10_SHARED_A_5", "STANDARD_TRIGGERS", "false", "false"},
	}
	ms := APItoModelAccountAction(aa)
	var slc [][]string
	lc, err := CsvDump(*ms)
	if err != nil {
		t.Error("Error dumping to csv: ", err)
	}
	slc = append(slc, lc)
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}
func TestAPItoModelActionTriggers(t *testing.T) {
	var ts []*utils.TPActionTriggers
	if rcv := APItoModelActionTriggers(ts); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	ts = []*utils.TPActionTriggers{
		{
			TPid: "TEST_TPID",
			ID:   "STANDARD_TRIGGERS",
			ActionTriggers: []*utils.TPActionTrigger{
				{
					Id:            "STANDARD_TRIGGERS",
					UniqueID:      "1",
					ThresholdType: "*min_balance",
					Weight:        0.7},
				{
					Id:            "STANDARD_TRIGGERS",
					UniqueID:      "2",
					ThresholdType: "*max_event_counter",
					Weight:        0.8},
			},
		},
	}
	eOut := ActionTriggerMdls{
		ActionTriggerMdl{
			Tpid:          "TEST_TPID",
			Tag:           "STANDARD_TRIGGERS",
			UniqueId:      "1",
			ThresholdType: "*min_balance",
			Weight:        0.7,
		},
		ActionTriggerMdl{
			Tpid:          "TEST_TPID",
			Tag:           "STANDARD_TRIGGERS",
			UniqueId:      "2",
			ThresholdType: "*max_event_counter",
			Weight:        0.8,
		},
	}
	if rcv := APItoModelActionTriggers(ts); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoModelAction(t *testing.T) {
	var as *utils.TPActions
	if rcv := APItoModelAction(as); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	as = &utils.TPActions{
		TPid: "TEST_TPID",
		ID:   "TEST_ACTIONS",
	}
	eOut := ActionMdls{
		ActionMdl{
			Tpid: "TEST_TPID",
			Tag:  "TEST_ACTIONS",
		},
	}
	if rcv := APItoModelAction(as); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	as = &utils.TPActions{
		TPid: "TEST_TPID",
		ID:   "TEST_ACTIONS",
		Actions: []*utils.TPAction{
			{
				Identifier:      "*topup_reset",
				BalanceType:     "*monetary",
				Units:           "5.0",
				ExpiryTime:      "*never",
				DestinationIds:  "*any",
				RatingSubject:   "special1",
				Categories:      "call",
				SharedGroups:    "GROUP1",
				BalanceWeight:   "10.0",
				ExtraParameters: "",
				Weight:          10.0},
			{
				Identifier:      "*http_post",
				BalanceType:     "",
				Units:           "0.0",
				ExpiryTime:      "",
				DestinationIds:  "",
				RatingSubject:   "",
				Categories:      "",
				SharedGroups:    "",
				BalanceWeight:   "0.0",
				ExtraParameters: "http://localhost/&param1=value1",
				Weight:          20.0},
		},
	}
	eOut = ActionMdls{
		ActionMdl{
			Tpid:            "TEST_TPID",
			Tag:             "TEST_ACTIONS",
			BalanceType:     "*monetary",
			Categories:      "call",
			DestinationTags: "*any",
			RatingSubject:   "special1",
			SharedGroups:    "GROUP1",
			ExpiryTime:      "*never",
			Units:           "5.0",
			BalanceWeight:   "10.0",
			Weight:          10,
			Action:          "*topup_reset",
		},
		ActionMdl{
			Tpid:            "TEST_TPID",
			Tag:             "TEST_ACTIONS",
			Action:          "*http_post",
			ExtraParameters: "http://localhost/\u0026param1=value1",
			Units:           "0.0",
			BalanceWeight:   "0.0",
			Weight:          20,
		},
	}
	if rcv := APItoModelAction(as); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoModelActions(t *testing.T) {
	var as []*utils.TPActions
	if rcv := APItoModelActions(as); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	}
	as = []*utils.TPActions{
		{
			TPid: "TEST_TPID",
			ID:   "TEST_ACTIONS",
		},
		{
			TPid: "TEST_TPID2",
			ID:   "TEST_ACTIONS2",
		},
	}
	eOut := ActionMdls{
		ActionMdl{
			Tpid: "TEST_TPID",
			Tag:  "TEST_ACTIONS",
		},
		ActionMdl{
			Tpid: "TEST_TPID2",
			Tag:  "TEST_ACTIONS2",
		},
	}
	if rcv := APItoModelActions(as); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	as = []*utils.TPActions{
		{
			TPid: "TEST_TPID",
			ID:   "TEST_ACTIONS",
			Actions: []*utils.TPAction{
				{
					Identifier:      "*topup_reset",
					BalanceType:     "*monetary",
					Units:           "5.0",
					ExpiryTime:      "*never",
					DestinationIds:  "*any",
					RatingSubject:   "special1",
					Categories:      "call",
					SharedGroups:    "GROUP1",
					BalanceWeight:   "10.0",
					ExtraParameters: "",
					Weight:          10.0},
				{
					Identifier:      "*http_post",
					BalanceType:     "",
					Units:           "0.0",
					BalanceWeight:   "0.0",
					ExtraParameters: "http://localhost/&param1=value1",
					Weight:          20.0},
			},
		},
	}
	eOut = ActionMdls{
		ActionMdl{
			Tpid:            "TEST_TPID",
			Tag:             "TEST_ACTIONS",
			BalanceType:     "*monetary",
			Categories:      "call",
			DestinationTags: "*any",
			RatingSubject:   "special1",
			SharedGroups:    "GROUP1",
			ExpiryTime:      "*never",
			Units:           "5.0",
			BalanceWeight:   "10.0",
			Weight:          10,
			Action:          "*topup_reset",
		},
		ActionMdl{
			Tpid:            "TEST_TPID",
			Tag:             "TEST_ACTIONS",
			Action:          "*http_post",
			ExtraParameters: "http://localhost/\u0026param1=value1",
			Units:           "0.0",
			BalanceWeight:   "0.0",
			Weight:          20,
		},
	}
	if rcv := APItoModelActions(as); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nreceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTpResourcesAsTpResources(t *testing.T) {
	tps := []*ResourceMdl{
		{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "ResGroup1",
			FilterIDs:          "FLTR_RES_GR1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			Stored:             false,
			Blocker:            false,
			Weight:             10.0,
			Limit:              "45",
			ThresholdIDs:       "WARN_RES1;WARN_RES1"},
		{
			Tpid:         "TEST_TPID",
			ID:           "ResGroup1",
			Tenant:       "cgrates.org",
			FilterIDs:    "FLTR_RES_GR1",
			ThresholdIDs: "WARN3"},
		{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "ResGroup2",
			FilterIDs:          "FLTR_RES_GR2",
			ActivationInterval: "2014-07-29T15:00:00Z",
			Stored:             false,
			Blocker:            false,
			Weight:             10.0,
			Limit:              "20"},
	}
	eTPs := []*utils.TPResourceProfile{
		{
			TPid:      tps[0].Tpid,
			Tenant:    tps[0].Tenant,
			ID:        tps[0].ID,
			FilterIDs: []string{"FLTR_RES_GR1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			Stored:       tps[0].Stored,
			Blocker:      tps[0].Blocker,
			Weight:       tps[0].Weight,
			Limit:        tps[0].Limit,
			ThresholdIDs: []string{"WARN_RES1", "WARN3"},
		},
		{
			TPid:      tps[2].Tpid,
			Tenant:    tps[2].Tenant,
			ID:        tps[2].ID,
			FilterIDs: []string{"FLTR_RES_GR2"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[2].ActivationInterval,
			},
			Stored:  tps[2].Stored,
			Blocker: tps[2].Blocker,
			Weight:  tps[2].Weight,
			Limit:   tps[2].Limit,
		},
	}
	rcvTPs := ResourceMdls(tps).AsTPResources()
	if len(rcvTPs) != len(eTPs) {
		t.Errorf("Expecting: %+v Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoResource(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:             "cgrates.org",
		TPid:               testTPID,
		ID:                 "ResGroup1",
		FilterIDs:          []string{"FLTR_RES_GR_1"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		Stored:             false,
		Blocker:            false,
		Weight:             10,
		Limit:              "2",
		ThresholdIDs:       []string{"TRes1"},
		AllocationMessage:  "asd",
	}
	eRL := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                tpRL.ID,
		Stored:            tpRL.Stored,
		Blocker:           tpRL.Blocker,
		Weight:            tpRL.Weight,
		FilterIDs:         []string{"FLTR_RES_GR_1"},
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: tpRL.AllocationMessage,
		Limit:             2,
	}
	at, _ := utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", "UTC")
	eRL.ActivationInterval = &utils.ActivationInterval{ActivationTime: at}
	if rl, err := APItoResource(tpRL, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRL, rl) {
		t.Errorf("Expecting: %+v, received: %+v", eRL, rl)
	}
}

func TestResourceProfileToAPI(t *testing.T) {
	expected := &utils.TPResourceProfile{
		Tenant:             "cgrates.org",
		ID:                 "ResGroup1",
		FilterIDs:          []string{"FLTR_RES_GR_1"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		Weight:             10,
		Limit:              "2",
		ThresholdIDs:       []string{"TRes1"},
		AllocationMessage:  "asd",
	}
	rp := &ResourceProfile{
		Tenant: "cgrates.org",
		ID:     "ResGroup1",
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Weight:            10,
		FilterIDs:         []string{"FLTR_RES_GR_1"},
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: "asd",
		Limit:             2,
	}

	if rcv := ResourceProfileToAPI(rp); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoModelResource(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:             "cgrates.org",
		TPid:               testTPID,
		ID:                 "ResGroup1",
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		Weight:             10,
		Limit:              "2",
		ThresholdIDs:       []string{"TRes1"},
		AllocationMessage:  "test",
	}
	expModel := &ResourceMdl{
		Tpid:               testTPID,
		Tenant:             "cgrates.org",
		ID:                 "ResGroup1",
		ActivationInterval: "2014-07-29T15:00:00Z",
		Weight:             10.0,
		Limit:              "2",
		ThresholdIDs:       "TRes1",
		AllocationMessage:  "test",
	}
	rcv := APItoModelResource(tpRL)
	if len(rcv) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], expModel) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expModel), utils.ToJSON(rcv[0]))
	}
}

func TestTPStatsAsTPStats(t *testing.T) {
	tps := StatMdls{
		&StatMdl{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*asr;*acc;*tcc;*acd;*tcd;*pdd",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
		&StatMdl{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*sum#BalanceValue;*average#BalanceValue;*tcc",
			ThresholdIDs:       "THRESH3",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
		&StatMdl{
			Tpid:               "TEST_TPID",
			Tenant:             "itsyscom.com",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*sum#BalanceValue;*average#BalanceValue;*tcc",
			ThresholdIDs:       "THRESH4",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
	}
	rcvTPs := tps.AsTPStats()
	if len(rcvTPs) != 2 {
		t.Errorf("Expecting: 2, received: %+v", len(rcvTPs))
	}
	for _, rcvTP := range rcvTPs {
		if rcvTP.Tenant == "cgrates.org" {
			if len(rcvTP.Metrics) != 8 {
				t.Errorf("Expecting: 8, received: %+v", len(rcvTP.Metrics))
			}
		} else {
			if len(rcvTP.Metrics) != 3 {
				t.Errorf("Expecting: 3, received: %+v", len(rcvTP.Metrics))
			}
		}
	}
}

func TestAPItoTPStats(t *testing.T) {
	tps := &utils.TPStatProfile{
		TPid:               testTPID,
		ID:                 "Stats1",
		FilterIDs:          []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		QueueLength:        100,
		TTL:                "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Stored:       false,
		Blocker:      false,
		Weight:       20.0,
	}
	eTPs := &StatQueueProfile{ID: tps.ID,
		QueueLength: tps.QueueLength,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		FilterIDs:    []string{"FLTR_1"},
		Stored:       tps.Stored,
		Blocker:      tps.Blocker,
		Weight:       20.0,
		MinItems:     tps.MinItems,
	}
	if eTPs.TTL, err = utils.ParseDurationWithNanosecs(tps.TTL); err != nil {
		t.Errorf("Got error: %+v", err)
	}
	at, _ := utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", "UTC")
	eTPs.ActivationInterval = &utils.ActivationInterval{ActivationTime: at}

	if st, err := APItoStats(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestStatQueueProfileToAPI(t *testing.T) {
	expected := &utils.TPStatProfile{
		Tenant:             "cgrates.org",
		ID:                 "Stats1",
		FilterIDs:          []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		QueueLength:        100,
		TTL:                "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Weight:       20.0,
	}
	sqPrf := &StatQueueProfile{
		Tenant:      "cgrates.org",
		ID:          "Stats1",
		QueueLength: 100,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		TTL:          time.Second,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		FilterIDs:    []string{"FLTR_1"},
		Weight:       20.0,
		MinItems:     1,
	}

	if rcv := StatQueueProfileToAPI(sqPrf); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoModelStats(t *testing.T) {
	tpS := &utils.TPStatProfile{
		TPid:      "TPS1",
		Tenant:    "cgrates.org",
		ID:        "Stat1",
		FilterIDs: []string{"*string:Account:1002"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*tcc",
			},
			{
				MetricID: "*average#Usage",
			},
		},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     2,
		ThresholdIDs: []string{"Th1"},
	}
	rcv := APItoModelStats(tpS)
	eRcv := StatMdls{
		&StatMdl{
			Tpid:               "TPS1",
			Tenant:             "cgrates.org",
			ID:                 "Stat1",
			FilterIDs:          "*string:Account:1002",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*tcc",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
			ThresholdIDs:       "Th1",
		},
		&StatMdl{
			Tpid:      "TPS1",
			Tenant:    "cgrates.org",
			ID:        "Stat1",
			MetricIDs: "*average#Usage",
		},
	}
	if len(rcv) != len(eRcv) {
		t.Errorf("Expecting: %+v, received: %+v", len(eRcv), len(rcv))
	} else if !reflect.DeepEqual(eRcv, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(eRcv), utils.ToJSON(rcv))
	}
}

func TestTPThresholdsAsTPThreshold(t *testing.T) {
	tps := []*ThresholdMdl{
		{
			Tpid:               "TEST_TPID",
			ID:                 "Threhold",
			FilterIDs:          "FilterID1;FilterID2;FilterID1;FilterID2;FilterID2",
			ActivationInterval: "2014-07-29T15:00:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
	}
	eTPs := []*utils.TPThresholdProfile{
		{
			TPid:      tps[0].Tpid,
			ID:        tps[0].ID,
			FilterIDs: []string{"FilterID1", "FilterID2"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			MinSleep:  tps[0].MinSleep,
			MaxHits:   tps[0].MaxHits,
			MinHits:   tps[0].MinHits,
			Blocker:   tps[0].Blocker,
			Weight:    tps[0].Weight,
			ActionIDs: []string{"WARN3"},
		},
		{
			TPid:      tps[0].Tpid,
			ID:        tps[0].ID,
			FilterIDs: []string{"FilterID2", "FilterID1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			MinSleep:  tps[0].MinSleep,
			MaxHits:   tps[0].MaxHits,
			MinHits:   tps[0].MinHits,
			Blocker:   tps[0].Blocker,
			Weight:    tps[0].Weight,
			ActionIDs: []string{"WARN3"},
		},
	}
	rcvTPs := ThresholdMdls(tps).AsTPThreshold()
	if !reflect.DeepEqual(eTPs[0], rcvTPs[0]) && !reflect.DeepEqual(eTPs[1], rcvTPs[0]) {
		t.Errorf("Expecting: %+v , Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoModelAccountActions(t *testing.T) {
	var aas []*utils.TPAccountActions
	if rcv := APItoModelAccountActions(aas); rcv != nil {
		t.Errorf("Expecting: nil , Received: %+v", utils.ToIJSON(rcv))
	}
	aas = []*utils.TPAccountActions{
		{
			TPid:             "TEST_TPID",
			LoadId:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Account:          "1001",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
		},
		{
			TPid:             "TEST_TPID2",
			LoadId:           "TEST_LOADID2",
			Tenant:           "cgrates.org",
			Account:          "1001",
			ActionPlanId:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersId: "STANDARD_TRIGGERS",
		},
	}
	eOut := AccountActionMdls{
		AccountActionMdl{
			Tpid:              "TEST_TPID",
			Loadid:            "TEST_LOADID",
			Tenant:            "cgrates.org",
			Account:           "1001",
			ActionPlanTag:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersTag: "STANDARD_TRIGGERS",
		},
		AccountActionMdl{
			Tpid:              "TEST_TPID2",
			Loadid:            "TEST_LOADID2",
			Tenant:            "cgrates.org",
			Account:           "1001",
			ActionPlanTag:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersTag: "STANDARD_TRIGGERS",
		},
	}
	if rcv := APItoModelAccountActions(aas); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestCSVHeader(t *testing.T) {
	var tps ResourceMdls
	eOut := []string{
		"#Tenant", "ID", "FilterIDs", "ActivationInterval", "UsageTTL", "Limit", "AllocationMessage", "Blocker", "Stored", "Weight", "ThresholdIDs",
	}
	if rcv := tps.CSVHeader(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestAPItoModelTPThreshold(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FilterID1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3"},
	}
	models := ThresholdMdls{
		{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			FilterIDs:          "FilterID1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold2(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_1", "FLTR_2"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3"},
	}
	models := ThresholdMdls{
		{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
		{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: "FLTR_2",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold3(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3", "LOG"},
	}
	models := ThresholdMdls{
		{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
		{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			ActionIDs: "LOG",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold4(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3", "LOG"},
	}
	models := ThresholdMdls{
		{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
		{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			ActionIDs: "LOG",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold5(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{},
	}
	rcv := APItoModelTPThreshold(th)
	if rcv != nil {
		t.Errorf("Expecting : nil, received: %+v", utils.ToJSON(rcv))
	}
}

func TestAPItoTPThreshold(t *testing.T) {
	tps := &utils.TPThresholdProfile{
		TPid:               testTPID,
		ID:                 "TH1",
		FilterIDs:          []string{"FilterID1", "FilterID2"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		MaxHits:            12,
		MinHits:            10,
		MinSleep:           "1s",
		Blocker:            false,
		Weight:             20.0,
		ActionIDs:          []string{"WARN3"},
	}

	eTPs := &ThresholdProfile{
		ID:        tps.ID,
		MaxHits:   tps.MaxHits,
		Blocker:   tps.Blocker,
		MinHits:   tps.MinHits,
		Weight:    tps.Weight,
		FilterIDs: tps.FilterIDs,
		ActionIDs: []string{"WARN3"},
	}
	if eTPs.MinSleep, err = utils.ParseDurationWithNanosecs(tps.MinSleep); err != nil {
		t.Errorf("Got error: %+v", err)
	}
	at, _ := utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", "UTC")
	eTPs.ActivationInterval = &utils.ActivationInterval{ActivationTime: at}
	if st, err := APItoThresholdProfile(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestThresholdProfileToAPI(t *testing.T) {
	expected := &utils.TPThresholdProfile{
		Tenant:             "cgrates.org",
		ID:                 "TH1",
		FilterIDs:          []string{"FilterID1", "FilterID2"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		MaxHits:            12,
		MinHits:            10,
		MinSleep:           "1s",
		Weight:             20.0,
		ActionIDs:          []string{"WARN3"},
	}

	thPrf := &ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TH1",
		FilterIDs: []string{"FilterID1", "FilterID2"},

		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  time.Second,
		Weight:    20.0,
		ActionIDs: []string{"WARN3"},
	}

	if rcv := ThresholdProfileToAPI(thPrf); !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestTPFilterAsTPFilter(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:    "TEST_TPID",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001;1002",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid: tps[0].Tpid,
			ID:   tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	if !(reflect.DeepEqual(eTPs, rcvTPs) || reflect.DeepEqual(eTPs[0], rcvTPs[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilterWithDynValues(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:               "TEST_TPID",
			ID:                 "Filter1",
			ActivationInterval: "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z",
			Type:               utils.MetaString,
			Element:            "CustomField",
			Values:             "1001;~*uch.<~*rep.CGRID;~*rep.RunID;-Cost>;1002;~*uch.<~*rep.CGRID;~*rep.RunID>",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid: tps[0].Tpid,
			ID:   tps[0].ID,
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "2014-08-29T15:00:00Z",
			},
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaString,
					Element: "CustomField",
					Values:  []string{"1001", "~*uch.<~*rep.CGRID;~*rep.RunID;-Cost>", "1002", "~*uch.<~*rep.CGRID;~*rep.RunID>"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	if !(reflect.DeepEqual(eTPs, rcvTPs) || reflect.DeepEqual(eTPs[0], rcvTPs[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilter2(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:    "TEST_TPID",
			Tenant:  "cgrates.org",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001;1002",
		},
		{
			Tpid:    "TEST_TPID",
			Tenant:  "anotherTenant",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1010",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid:   tps[0].Tpid,
			Tenant: "cgrates.org",
			ID:     tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
		{
			TPid:   tps[1].Tpid,
			Tenant: "anotherTenant",
			ID:     tps[1].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1010"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	if len(eTPs) != len(rcvTPs) {
		t.Errorf("Expecting: %+v ,Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilter3(t *testing.T) {
	tps := []*FilterMdl{
		{
			Tpid:    "TEST_TPID",
			Tenant:  "cgrates.org",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001",
		},
		{
			Tpid:    "TEST_TPID",
			Tenant:  "cgrates.org",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001",
		},
		{
			Tpid:    "TEST_TPID",
			Tenant:  "anotherTenant",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1010",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid:   tps[0].Tpid,
			Tenant: "cgrates.org",
			ID:     tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
		{
			TPid:   tps[1].Tpid,
			Tenant: "anotherTenant",
			ID:     tps[1].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1010"},
				},
			},
		},
	}

	rcvTPs := FilterMdls(tps).AsTPFilter()
	sort.Strings(rcvTPs[0].Filters[0].Values)
	if len(eTPs) != len(rcvTPs) {
		t.Errorf("Expecting: %+v ,Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoModelTPFilter(t *testing.T) {
	var th *utils.TPFilterProfile
	if rcv := APItoModelTPFilter(th); rcv != nil {
		t.Errorf("Expecting: nil ,Received: %+v", utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		ID: "someID",
	}
	if rcv := APItoModelTPFilter(th); rcv != nil {
		t.Errorf("Expecting: nil ,Received: %+v", utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		ID: "someID",
		Filters: []*utils.TPFilter{
			{
				Type:    utils.MetaPrefix,
				Element: "Account",
				Values:  []string{"1010"},
			},

			{
				Type:    utils.MetaPrefix,
				Element: "Account",
				Values:  []string{"0708"},
			},
		},
	}
	eOut := FilterMdls{
		&FilterMdl{
			ID:      "someID",
			Type:    "*prefix",
			Element: "Account",
			Values:  "1010",
		},
		&FilterMdl{
			ID:      "someID",
			Type:    "*prefix",
			Element: "Account",
			Values:  "0708",
		},
	}
	if rcv := APItoModelTPFilter(th); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		TPid:   "TPid",
		Tenant: "cgrates.org",
		ID:     "someID",
		Filters: []*utils.TPFilter{
			{
				Type:    utils.MetaPrefix,
				Element: "Account",
				Values:  []string{"1001", "1002"},
			},
		},
	}
	eOut = FilterMdls{
		{
			Tpid:    "TPid",
			Tenant:  "cgrates.org",
			ID:      "someID",
			Type:    "*prefix",
			Element: "Account",
			Values:  "1001;1002",
		},
	}
	if rcv := APItoModelTPFilter(th); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	th = &utils.TPFilterProfile{
		TPid:   "TPid",
		ID:     "testID",
		Tenant: "cgrates.org",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2014-08-29T15:00:00Z",
		},
		Filters: []*utils.TPFilter{
			{
				Type:    utils.MetaString,
				Element: "CustomField",
				Values:  []string{"1001", "~*uch.<~*rep.CGRID;~*rep.RunID;-Cost>", "1002", "~*uch.<~*rep.CGRID;~*rep.RunID>"},
			},
		},
	}
	eOut = FilterMdls{
		{
			Tpid:               "TPid",
			Tenant:             "cgrates.org",
			ID:                 "testID",
			Type:               "*string",
			Element:            "CustomField",
			Values:             "1001;~*uch.\u003c~*rep.CGRID;~*rep.RunID;-Cost\u003e;1002;~*uch.\u003c~*rep.CGRID;~*rep.RunID\u003e",
			ActivationInterval: "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z",
		},
	}
	if rcv := APItoModelTPFilter(th); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoTPFilter(t *testing.T) {
	tps := &utils.TPFilterProfile{
		TPid:   testTPID,
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Filters: []*utils.TPFilter{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}

	eTPs := &Filter{
		Tenant: "cgrates.org",
		ID:     tps.ID,
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	if err := eTPs.Compile(); err != nil {
		t.Fatal(err)
	}
	if st, err := APItoFilter(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestFilterToTPFilter(t *testing.T) {
	filter := &Filter{
		Tenant: "cgrates.org",
		ID:     "Fltr1",
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
		},
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	tpfilter := &utils.TPFilterProfile{
		ID:     "Fltr1",
		Tenant: "cgrates.org",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-01-14T00:00:00Z",
			ExpiryTime:     "2014-01-14T00:00:00Z",
		},
		Filters: []*utils.TPFilter{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	eTPFilter := FilterToTPFilter(filter)
	if !reflect.DeepEqual(tpfilter, eTPFilter) {
		t.Errorf("Expecting: %+v, received: %+v", tpfilter, eTPFilter)
	}
}

func TestCsvHeader(t *testing.T) {
	var tps RouteMdls
	eOut := []string{
		"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.Sorting, utils.SortingParameters, utils.RouteID, utils.RouteFilterIDs,
		utils.RouteAccountIDs, utils.RouteRatingplanIDs, utils.RouteResourceIDs,
		utils.RouteStatIDs, utils.RouteWeight, utils.RouteBlocker,
		utils.RouteParameters, utils.Weight,
	}
	if rcv := tps.CSVHeader(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestAPItoAttributeProfile(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weight: 20,
	}
	expected := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if rcv, err := APItoAttributeProfile(tpAlsPrf, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAttributeProfileToAPI(t *testing.T) {
	exp := &utils.TPAttributeProfile{
		TPid:      utils.EmptyString,
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weight: 20,
	}
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if rcv := AttributeProfileToAPI(attr); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestAttributeProfileToAPI2(t *testing.T) {
	exp := &utils.TPAttributeProfile{
		TPid:      utils.EmptyString,
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Test",
				Value: "~*req.Account",
			},
		},
		Weight: 20,
	}
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Test",
				Value: config.NewRSRParsersMustCompile("~*req.Account", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if rcv := AttributeProfileToAPI(attr); !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPAttribute(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1", "con2"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		Attributes: []*utils.TPAttribute{
			{FilterIDs: []string{"filter_id1", "filter_id2"},
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weight: 20,
	}
	expected := AttributeMdls{
		&AttributeMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "ALS1",
			Contexts:           "con1;con2",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			AttributeFilterIDs: "filter_id1;filter_id2",
			Path:               utils.MetaReq + utils.NestingSep + "FL1",
			Value:              "Al1",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPAttribute(tpAlsPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestCsvDumpForAttributeModels(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL2",
				Value: "Al2",
			},
		},
		Weight: 20,
	}
	expected := AttributeMdls{
		&AttributeMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "ALS1",
			Contexts:           "con1",
			FilterIDs:          "FLTR_ACNT_dan",
			Path:               utils.MetaReq + utils.NestingSep + "FL1",
			Value:              "Al1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&AttributeMdl{
			Tpid:   "TP1",
			Tenant: "cgrates.org",
			ID:     "ALS1",
			Path:   utils.MetaReq + utils.NestingSep + "FL2",
			Value:  "Al2",
		},
	}
	rcv := APItoModelTPAttribute(tpAlsPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
	expRecord := []string{"cgrates.org", "ALS1", "con1", "FLTR_ACNT_dan", "2014-07-14T14:35:00Z", "", "*req.FL1", "", "Al1", "false", "20"}
	for i, model := range rcv {
		if i == 1 {
			expRecord = []string{"cgrates.org", "ALS1", "", "", "", "", "*req.FL2", "", "Al2", "false", "0"}
		}
		if csvRecordRcv, _ := CsvDump(model); !reflect.DeepEqual(expRecord, csvRecordRcv) {
			t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRecord), utils.ToJSON(csvRecordRcv))
		}
	}

}

func TestModelAsTPAttribute2(t *testing.T) {
	models := AttributeMdls{
		&AttributeMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "ALS1",
			Contexts:           "con1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			Path:               utils.MetaReq + utils.NestingSep + "FL1",
			Value:              "Al1",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			Weight:             20,
		},
	}
	expected := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weight: 20,
	}
	expected2 := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_DST_DE", "FLTR_ACNT_dan"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weight: 20,
	}
	rcv := models.AsTPAttributes()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) && !reflect.DeepEqual(expected2, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestModelAsTPAttribute(t *testing.T) {
	models := AttributeMdls{
		&AttributeMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "ALS1",
			Contexts:           "con1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			Path:               utils.MetaReq + utils.NestingSep + "FL1",
			Value:              "Al1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	expected := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weight: 20,
	}
	expected2 := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_DST_DE", "FLTR_ACNT_dan"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weight: 20,
	}
	rcv := models.AsTPAttributes()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) && !reflect.DeepEqual(expected2, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoChargerProfile(t *testing.T) {
	tpCPP := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}

	expected := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	if rcv, err := APItoChargerProfile(tpCPP, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestChargerProfileToAPI(t *testing.T) {
	exp := &utils.TPChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}

	chargerPrf := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 35, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	if rcv := ChargerProfileToAPI(chargerPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs and AttributeIDs are equal
func TestAPItoModelTPCharger(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			Weight:             20,
		},
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_DST_DE",
			AttributeIDs:       "ATTR2",
			ActivationInterval: "",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs is smaller than AttributeIDs
func TestAPItoModelTPCharger2(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			Weight:             20,
		},
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			AttributeIDs:       "ATTR2",
			ActivationInterval: "",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs is greater than AttributeIDs
func TestAPItoModelTPCharger3(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&ChargerMdl{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: "FLTR_DST_DE",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// len(AttributeIDs) is 0
func TestAPItoModelTPCharger4(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		Weight: 20,
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// len(FilterIDs) is 0
func TestAPItoModelTPCharger5(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:   "TP1",
		Tenant: "cgrates.org",
		ID:     "Charger1",
		RunID:  "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// both len(AttributeIDs) and len(FilterIDs) are 0
func TestAPItoModelTPCharger6(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:   "TP1",
		Tenant: "cgrates.org",
		ID:     "Charger1",
		RunID:  "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Weight: 20,
	}
	expected := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			RunID:              "*rated",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestModelAsTPChargers(t *testing.T) {
	models := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	expected2 := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_DST_DE", "FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	rcv := models.AsTPChargers()
	if !reflect.DeepEqual(expected, rcv[0]) && !reflect.DeepEqual(expected2, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestModelAsTPChargers2(t *testing.T) {
	models := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:              "*rated",
			AttributeIDs:       "*constant:*req.RequestType:*rated;*constant:*req.Category:call;ATTR1;*constant:*req.Category:call",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"*constant:*req.RequestType:*rated;*constant:*req.Category:call", "ATTR1", "*constant:*req.Category:call"},
		Weight:       20,
	}
	rcv := models.AsTPChargers()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestModelAsTPChargers3(t *testing.T) {
	models := ChargerMdls{
		&ChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:              "*rated",
			AttributeIDs:       "*constant:*req.RequestType:*rated;*constant:*req.Category:call;ATTR1;*constant:*req.Category:call&<~*req.OriginID;_suf>",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			Weight:             20,
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		AttributeIDs: []string{"*constant:*req.RequestType:*rated;*constant:*req.Category:call", "ATTR1", "*constant:*req.Category:call&<~*req.OriginID;_suf>"},
		Weight:       20,
	}
	rcv := models.AsTPChargers()
	sort.Strings(rcv[0].FilterIDs)
	if !reflect.DeepEqual(expected, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoDispatcherProfile(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:       "TP1",
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203", "*ratio:2"},
				Blocker:   false,
			},
		},
	}

	expected := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		StrategyParams: map[string]any{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.54.203", utils.MetaRatio: "2"},
				Blocker:   false,
			},
		},
	}
	if rcv, err := APItoDispatcherProfile(tpDPP, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestDispatcherProfileToAPI(t *testing.T) {
	exp := &utils.TPDispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203", "*ratio:2"},
				Blocker:   false,
			},
		},
	}
	exp2 := &utils.TPDispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"*ratio:2", "192.168.54.203"},
				Blocker:   false,
			},
		},
	}

	dspPrf := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		StrategyParams: map[string]any{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.54.203", utils.MetaRatio: "2"},
				Blocker:   false,
			},
		},
	}
	if rcv := DispatcherProfileToAPI(dspPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rcv) && !reflect.DeepEqual(exp2, rcv) {
		t.Errorf("Expecting : \n %+v \n  or \n %+v \n ,\n received: %+v", utils.ToJSON(exp), utils.ToJSON(exp2), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPDispatcher(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:       "TP1",
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203"},
				Blocker:   false,
			},
			{
				ID:        "C2",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.204"},
				Blocker:   false,
			},
		},
	}
	expected := DispatcherProfileMdls{
		&DispatcherProfileMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Dsp",
			Subsystems:         "*any",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			Strategy:           utils.MetaFirst,
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
			ConnID:             "C1",
			ConnWeight:         10,
			ConnBlocker:        false,
			ConnParameters:     "192.168.54.203",
		},
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			ConnID:         "C2",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.204",
		},
	}
	rcv := APItoModelTPDispatcherProfile(tpDPP)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestTPDispatcherHostsCSVHeader(t *testing.T) {
	tps := &DispatcherHostMdls{}
	eOut := []string{"#" + utils.Tenant, utils.ID, utils.Address, utils.Transport, utils.SynchronousCfg, utils.ConnectAttemptsCfg, utils.ReconnectsCfg, utils.MaxReconnectIntervalCfg, utils.ConnectTimeoutCfg, utils.ReplyTimeoutCfg, utils.TLS, utils.ClientKeyCfg, utils.ClientCerificateCfg, utils.CaCertificateCfg}
	if rcv := tps.CSVHeader(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPDispatcherHostsAsTPDispatcherHosts(t *testing.T) {
	tps := &DispatcherHostMdls{}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:     "ID1",
			Tenant: "Tenant1",
		}}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			ID:                "ID1",
			Tenant:            "Tenant1",
			Address:           "localhost:6012",
			ConnectAttempts:   2,
			Reconnects:        5,
			ConnectTimeout:    "2m",
			ReplyTimeout:      "1m",
			TLS:               true,
			ClientKey:         "client_key",
			ClientCertificate: "client_certificate",
			CaCertificate:     "ca_certificate",
		}}
	eOut := []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant1",
			ID:     "ID1",
			Conn: &utils.TPDispatcherHostConn{
				Address:           "localhost:6012",
				Transport:         "*json",
				ConnectAttempts:   2,
				Reconnects:        5,
				ConnectTimeout:    2 * time.Minute,
				ReplyTimeout:      1 * time.Minute,
				TLS:               true,
				ClientKey:         "client_key",
				ClientCertificate: "client_certificate",
				CaCertificate:     "ca_certificate",
			},
		},
	}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:   "Address2",
			ID:        "ID2",
			Tenant:    "Tenant2",
			Transport: "*gob",
		}}
	eOut = []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant2",
			ID:     "ID2",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "Address2",
				Transport: "*gob",
			},
		},
	}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:   "Address3",
			ID:        "ID3",
			Tenant:    "Tenant3",
			Transport: "*gob",
		},
	}
	eOut = []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant3",
			ID:     "ID3",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "Address3",
				Transport: "*gob",
			},
		},
	}
	if rcv, err := tps.AsTPDispatcherHosts(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:   "Address4",
			ID:        "ID4",
			Tenant:    "Tenant4",
			Transport: "*gob",
		},
	}
	eOut = []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant4",
			ID:     "ID4",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "Address4",
				Transport: "*gob",
			},
		},
	}
	rcv, err := tps.AsTPDispatcherHosts()
	if err != nil {
		t.Error(err)
	}
	sort.Slice(rcv, func(i, j int) bool { return strings.Compare(rcv[i].ID, rcv[j].ID) < 0 })
	if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:              "Address4",
			ID:                   "ID4",
			Tenant:               "Tenant4",
			Transport:            "*gob",
			MaxReconnectInterval: "val",
		},
	}
	if _, err = tps.AsTPDispatcherHosts(); err == nil {
		t.Error("expected <error>")
	}
	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:   "Address4",
			ID:        "ID4",
			Tenant:    "Tenant4",
			Transport: "*gob",

			ConnectTimeout: "timeout",
		},
	}
	if _, err = tps.AsTPDispatcherHosts(); err == nil {
		t.Error("expected <error>")
	}
	tps = &DispatcherHostMdls{
		&DispatcherHostMdl{
			Address:      "Address4",
			ID:           "ID4",
			Tenant:       "Tenant4",
			Transport:    "*gob",
			ReplyTimeout: "reply",
		},
	}
	if _, err = tps.AsTPDispatcherHosts(); err == nil {
		t.Error("expected <error>")
	}

}

func TestAPItoModelTPDispatcherHost(t *testing.T) {
	var tpDPH *utils.TPDispatcherHost
	if rcv := APItoModelTPDispatcherHost(tpDPH); rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tpDPH = &utils.TPDispatcherHost{
		Tenant: "Tenant",
		ID:     "ID",
		Conn: &utils.TPDispatcherHostConn{
			Address:           "Address1",
			Transport:         "*json",
			ConnectAttempts:   3,
			Reconnects:        5,
			ConnectTimeout:    1 * time.Minute,
			ReplyTimeout:      2 * time.Minute,
			TLS:               true,
			ClientKey:         "client_key",
			ClientCertificate: "client_certificate",
			CaCertificate:     "ca_certificate",
		},
	}
	eOut := &DispatcherHostMdl{
		Address:              "Address1",
		Transport:            "*json",
		Tenant:               "Tenant",
		ID:                   "ID",
		ConnectAttempts:      3,
		Reconnects:           5,
		MaxReconnectInterval: "0s",
		ConnectTimeout:       "1m0s",
		ReplyTimeout:         "2m0s",
		TLS:                  true,
		ClientKey:            "client_key",
		ClientCertificate:    "client_certificate",
		CaCertificate:        "ca_certificate",
	}
	if rcv := APItoModelTPDispatcherHost(tpDPH); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestAPItoDispatcherHost(t *testing.T) {
	var tpDPH *utils.TPDispatcherHost
	if rcv := APItoDispatcherHost(tpDPH); rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tpDPH = &utils.TPDispatcherHost{
		Tenant: "Tenant1",
		ID:     "ID1",
		Conn: &utils.TPDispatcherHostConn{
			Address:           "localhost:6012",
			Transport:         "*json",
			ConnectAttempts:   3,
			Reconnects:        5,
			ConnectTimeout:    1 * time.Minute,
			ReplyTimeout:      2 * time.Minute,
			TLS:               true,
			ClientKey:         "client_key",
			ClientCertificate: "client_certificate",
			CaCertificate:     "ca_certificate",
		},
	}

	eOut := &DispatcherHost{
		Tenant: "Tenant1",
		RemoteHost: &config.RemoteHost{
			ID:                "ID1",
			Address:           "localhost:6012",
			Transport:         "*json",
			Reconnects:        5,
			ConnectTimeout:    1 * time.Minute,
			ReplyTimeout:      2 * time.Minute,
			TLS:               true,
			ClientKey:         "client_key",
			ClientCertificate: "client_certificate",
			CaCertificate:     "ca_certificate",
			ConnectAttempts:   3,
		},
	}
	if rcv := APItoDispatcherHost(tpDPH); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tpDPH = &utils.TPDispatcherHost{
		Tenant: "Tenant2",
		ID:     "ID2",
		Conn: &utils.TPDispatcherHostConn{
			Address:   "Address1",
			Transport: "*json",
			TLS:       true,
		},
	}
	eOut = &DispatcherHost{
		Tenant: "Tenant2",
		RemoteHost: &config.RemoteHost{
			ID:        "ID2",
			Address:   "Address1",
			Transport: "*json",
			TLS:       true,
		},
	}
	if rcv := APItoDispatcherHost(tpDPH); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestDispatcherHostToAPI(t *testing.T) {
	dph := &DispatcherHost{
		Tenant: "Tenant1",
		RemoteHost: &config.RemoteHost{
			Address:           "127.0.0.1:2012",
			Transport:         "*json",
			ConnectAttempts:   0,
			Reconnects:        0,
			ConnectTimeout:    1 * time.Minute,
			ReplyTimeout:      1 * time.Minute,
			TLS:               false,
			ClientKey:         "",
			ClientCertificate: "",
			CaCertificate:     "",
		},
	}
	eOut := &utils.TPDispatcherHost{
		Tenant: "Tenant1",
		Conn: &utils.TPDispatcherHostConn{
			Address:           "127.0.0.1:2012",
			Transport:         "*json",
			ConnectAttempts:   0,
			Reconnects:        0,
			ConnectTimeout:    1 * time.Minute,
			ReplyTimeout:      1 * time.Minute,
			TLS:               false,
			ClientKey:         "",
			ClientCertificate: "",
			CaCertificate:     "",
		},
	}
	if rcv := DispatcherHostToAPI(dph); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestTPRoutesAsTPRouteProfile(t *testing.T) {
	mdl := RouteMdls{
		&RouteMdl{
			PK:                 1,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "FltrRoute",
			ActivationInterval: "2017-11-27T00:00:00Z",
			Sorting:            "*weight",
			SortingParameters:  "srtPrm1",
			RouteID:            "route1",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        10.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             10.0,
			CreatedAt:          time.Time{},
		},
		&RouteMdl{
			PK:                 2,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "",
			ActivationInterval: "",
			Sorting:            "",
			SortingParameters:  "",
			RouteID:            "route2",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        20.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             0,
			CreatedAt:          time.Time{},
		},
	}
	expPrf := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"FltrRoute"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2017-11-27T00:00:00Z",
				ExpiryTime:     "",
			},
			Routes: []*utils.TPRoute{
				{
					ID:     "route1",
					Weight: 10.0,
				},
				{
					ID:     "route2",
					Weight: 20.0,
				},
			},
			Weight: 10,
		},
	}
	rcv := mdl.AsTPRouteProfile()
	sort.Slice(rcv[0].Routes, func(i, j int) bool {
		return strings.Compare(rcv[0].Routes[i].ID, rcv[0].Routes[j].ID) < 0
	})
	if !reflect.DeepEqual(rcv, expPrf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrf), utils.ToJSON(rcv))
	}

	mdlReverse := RouteMdls{
		&RouteMdl{
			PK:                 2,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "",
			ActivationInterval: "",
			Sorting:            "",
			SortingParameters:  "",
			RouteID:            "route2",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        20.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             0,
			CreatedAt:          time.Time{},
		},
		&RouteMdl{
			PK:                 1,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "FltrRoute",
			ActivationInterval: "2017-11-27T00:00:00Z",
			Sorting:            "*weight",
			SortingParameters:  "srtPrm1",
			RouteID:            "route1",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        10.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             10.0,
			CreatedAt:          time.Time{},
		},
	}
	expPrfRev := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"FltrRoute"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2017-11-27T00:00:00Z",
				ExpiryTime:     "",
			},
			Routes: []*utils.TPRoute{
				{
					ID:     "route1",
					Weight: 10.0,
				},
				{
					ID:     "route2",
					Weight: 20.0,
				},
			},
			Weight: 10,
		},
	}
	rcvRev := mdlReverse.AsTPRouteProfile()
	sort.Slice(rcvRev[0].Routes, func(i, j int) bool {
		return strings.Compare(rcvRev[0].Routes[i].ID, rcvRev[0].Routes[j].ID) < 0
	})
	sort.Strings(rcvRev[0].SortingParameters)
	if !reflect.DeepEqual(rcvRev, expPrfRev) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrfRev), utils.ToJSON(rcvRev))
	}
}

func TestTPRoutesAsTPRouteProfile2(t *testing.T) {
	mdl := RouteMdls{
		&RouteMdl{
			PK:                 1,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "FltrRoute",
			ActivationInterval: "2017-11-27T00:00:00Z;2017-11-28T00:00:00Z",
			Sorting:            "*weight",
			SortingParameters:  "srtPrm1",
			RouteID:            "route1",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        10.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             10.0,
			CreatedAt:          time.Time{},
		},
		&RouteMdl{
			PK:                 2,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "",
			ActivationInterval: "2017-11-27T00:00:00Z;2017-11-28T00:00:00Z",
			Sorting:            "",
			SortingParameters:  "",
			RouteID:            "route2",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        20.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             0,
			CreatedAt:          time.Time{},
		},
	}
	expPrf := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"FltrRoute"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2017-11-27T00:00:00Z",
				ExpiryTime:     "2017-11-28T00:00:00Z",
			},
			Routes: []*utils.TPRoute{
				{
					ID:     "route1",
					Weight: 10.0,
				},
				{
					ID:     "route2",
					Weight: 20.0,
				},
			},
			Weight: 10,
		},
	}
	rcv := mdl.AsTPRouteProfile()
	sort.Slice(rcv[0].Routes, func(i, j int) bool {
		return strings.Compare(rcv[0].Routes[i].ID, rcv[0].Routes[j].ID) < 0
	})
	if !reflect.DeepEqual(rcv, expPrf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrf), utils.ToJSON(rcv))
	}

	mdlReverse := RouteMdls{
		&RouteMdl{
			PK:                 2,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "",
			ActivationInterval: "",
			Sorting:            "",
			SortingParameters:  "",
			RouteID:            "route2",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        20.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             0,
			CreatedAt:          time.Time{},
		},
		&RouteMdl{
			PK:                 1,
			Tpid:               "TP",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "FltrRoute",
			ActivationInterval: "2017-11-27T00:00:00Z;2017-11-28T00:00:00Z",
			Sorting:            "*weight",
			SortingParameters:  "srtPrm1",
			RouteID:            "route1",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "",
			RouteWeight:        10.0,
			RouteBlocker:       false,
			RouteParameters:    "",
			Weight:             10.0,
			CreatedAt:          time.Time{},
		},
	}
	expPrfRev := []*utils.TPRouteProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "RoutePrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1"},
			FilterIDs:         []string{"FltrRoute"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2017-11-27T00:00:00Z",
				ExpiryTime:     "2017-11-28T00:00:00Z",
			},
			Routes: []*utils.TPRoute{
				{
					ID:     "route1",
					Weight: 10.0,
				},
				{
					ID:     "route2",
					Weight: 20.0,
				},
			},
			Weight: 10,
		},
	}
	rcvRev := mdlReverse.AsTPRouteProfile()
	sort.Slice(rcvRev[0].Routes, func(i, j int) bool {
		return strings.Compare(rcvRev[0].Routes[i].ID, rcvRev[0].Routes[j].ID) < 0
	})
	sort.Strings(rcvRev[0].SortingParameters)
	if !reflect.DeepEqual(rcvRev, expPrfRev) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrfRev), utils.ToJSON(rcvRev))
	}
}

func TestModelHelperCsvLoadError(t *testing.T) {
	type testStruct struct {
		Id        int64
		Tpid      string
		Tag       string `index:"cat" re:".*"`
		Prefix    string `index:"1" re:".*"`
		CreatedAt time.Time
	}
	var testStruct1 testStruct
	_, err := csvLoad(testStruct1, []string{"TEST_DEST", "+492"})
	if err == nil || err.Error() != "invalid testStruct.Tag index cat" {
		t.Errorf("Expecting: <invalid testStruct.Tag index cat>,\nReceived: <%+v>", err)
	}
}

func TestModelHelperCsvLoadError2(t *testing.T) {
	type testStruct struct {
		Id        int64
		Tpid      string
		Tag       string `index:"0" re:"cat"`
		Prefix    string `index:"1" re:".*"`
		CreatedAt time.Time
	}
	var testStruct1 testStruct
	_, err := csvLoad(testStruct1, []string{"TEST_DEST", "+492"})

	if err == nil || err.Error() != "invalid testStruct.Tag value TEST_DEST" {
		t.Errorf("Expecting: <invalid testStruct.Tag value TEST_DEST>,\nReceived: <%+v>", err)
	}
}

func TestModelHelpersCsvDumpError(t *testing.T) {
	type testStruct struct {
		Id        int64
		Tpid      string
		Tag       string `index:"cat"  re:".*"`
		Prefix    string `index:"1"  re:".*"`
		CreatedAt time.Time
	}
	var testStruct1 testStruct
	_, err := CsvDump(testStruct1)
	if err == nil || err.Error() != "invalid testStruct.Tag index cat" {
		t.Errorf("\nExpecting: <invalid testStruct.Tag index cat>,\n  Received: <%+v>", err)
	}
}

func TestModelHelpersAsMapRatesError(t *testing.T) {
	tps := RateMdls{{RateUnit: "true"}}
	_, err := tps.AsMapRates()
	if err == nil || err.Error() != "time: invalid duration \"true\"" {
		t.Errorf("Expecting: <time: invalid duration \"true\">,\n  Received: <%+v>", err)
	}
}

func TestModelHelpersAsTPRatesError(t *testing.T) {
	tps := RateMdls{{RateUnit: "true"}}
	_, err := tps.AsTPRates()
	if err == nil || err.Error() != "time: invalid duration \"true\"" {
		t.Errorf("Expecting: <time: invalid duration \"true\">,\n  Received: <%+v>", err)
	}

}

func TestAPItoModelTPRoutesCase1(t *testing.T) {
	structTest := &utils.TPRouteProfile{}
	structTest2 := RouteMdls{}
	structTest2 = nil
	result := APItoModelTPRoutes(structTest)
	if !reflect.DeepEqual(structTest2, result) {
		t.Errorf("Expecting: <%+v>,\n  Received: <%+v>", structTest2, result)
	}
}

func TestAPItoModelTPRoutesEmptySlice(t *testing.T) {
	tpRoute := []*utils.TPRouteProfile{
		{
			TPid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "RoutePrf",
			FilterIDs: []string{},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-29T15:00:00Z",
				ExpiryTime:     "2014-08-29T15:00:00Z",
			},
			Sorting:           "*lc",
			SortingParameters: []string{},
			Routes: []*utils.TPRoute{
				{
					ID:              "route1",
					FilterIDs:       []string{},
					AccountIDs:      []string{},
					RatingPlanIDs:   []string{},
					ResourceIDs:     []string{},
					StatIDs:         []string{"Stat1", "Stat2"},
					Weight:          10,
					Blocker:         false,
					RouteParameters: "SortingParam1",
				},
			},
			Weight: 20,
		},
	}
	expMdl := RouteMdls{
		{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "RoutePrf",
			FilterIDs:          "",
			ActivationInterval: "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z",
			Sorting:            "*lc",
			SortingParameters:  "",
			RouteID:            "route1",
			RouteFilterIDs:     "",
			RouteAccountIDs:    "",
			RouteRatingplanIDs: "",
			RouteResourceIDs:   "",
			RouteStatIDs:       "Stat1;Stat2",
			RouteWeight:        10,
			RouteBlocker:       false,
			RouteParameters:    "SortingParam1",
			Weight:             20,
		},
	}
	var mdl RouteMdls
	if mdl = APItoModelTPRoutes(tpRoute[0]); !reflect.DeepEqual(mdl, expMdl) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expMdl), utils.ToJSON(mdl))
	}

	//back to route profile
	//all the empty slices will be nil because of converting back an empty string into a slice
	tpRoute[0].FilterIDs = nil
	tpRoute[0].SortingParameters = nil
	tpRoute[0].Routes[0].FilterIDs = nil
	tpRoute[0].Routes[0].AccountIDs = nil
	tpRoute[0].Routes[0].RatingPlanIDs = nil
	tpRoute[0].Routes[0].ResourceIDs = nil
	if newRcv := mdl.AsTPRouteProfile(); !reflect.DeepEqual(newRcv, tpRoute) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(tpRoute), utils.ToJSON(newRcv))
	}
}

func TestAPItoModelTPRoutesCase2(t *testing.T) {
	structTest := &utils.TPRouteProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "RoutePrf",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2014-08-29T15:00:00Z",
		},
		Sorting:           "*lc",
		SortingParameters: []string{"PARAM1", "PARAM2"},
		Routes: []*utils.TPRoute{
			{
				ID:              "route1",
				FilterIDs:       []string{"FLTR_1", "FLTR_2"},
				AccountIDs:      []string{"Acc1", "Acc2"},
				RatingPlanIDs:   []string{"RPL_1", "RPL_2"},
				ResourceIDs:     []string{"ResGroup1", "ResGroup2"},
				StatIDs:         []string{"Stat1", "Stat2"},
				Weight:          10,
				Blocker:         false,
				RouteParameters: "SortingParam1",
			},
		},
		Weight: 20,
	}
	expStructTest := RouteMdls{{
		Tpid:               "TP1",
		Tenant:             "cgrates.org",
		ID:                 "RoutePrf",
		FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
		ActivationInterval: "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z",
		Sorting:            "*lc",
		SortingParameters:  "PARAM1;PARAM2",
		RouteID:            "route1",
		RouteFilterIDs:     "FLTR_1;FLTR_2",
		RouteAccountIDs:    "Acc1;Acc2",
		RouteRatingplanIDs: "RPL_1;RPL_2",
		RouteResourceIDs:   "ResGroup1;ResGroup2",
		RouteStatIDs:       "Stat1;Stat2",
		RouteWeight:        10,
		RouteBlocker:       false,
		RouteParameters:    "SortingParam1",
		Weight:             20,
	},
	}
	sort.Strings(structTest.FilterIDs)
	sort.Strings(structTest.SortingParameters)
	result := APItoModelTPRoutes(structTest)
	if !reflect.DeepEqual(result, expStructTest) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStructTest), utils.ToJSON(result))
	}
}

func TestAPItoModelResourceCase1(t *testing.T) {
	var testStruct *utils.TPResourceProfile
	var testStruct2 ResourceMdls
	testStruct = nil
	result := APItoModelResource(testStruct)
	if !reflect.DeepEqual(result, testStruct2) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", testStruct2, result)
	}
}

func TestAPItoModelResourceCase2(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		Tenant:    "cgrates.org",
		TPid:      testTPID,
		ID:        "ResGroup1",
		FilterIDs: []string{},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2015-07-29T15:00:00Z",
		},
		UsageTTL:          "Test_TTL",
		Weight:            10,
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1", "TRes2"},
		AllocationMessage: "test",
	}
	expectedStruct := ResourceMdls{
		{
			Tenant:             "cgrates.org",
			Tpid:               "LoaderCSVTests",
			ID:                 "ResGroup1",
			FilterIDs:          "",
			ActivationInterval: "2014-07-29T15:00:00Z;2015-07-29T15:00:00Z",
			UsageTTL:           "Test_TTL",
			Weight:             10,
			Limit:              "2",
			ThresholdIDs:       "TRes1;TRes2",
			AllocationMessage:  "test",
		},
	}
	result := APItoModelResource(testStruct)
	if !reflect.DeepEqual(result, expectedStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expectedStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelResourceCase3(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		Tenant:    "cgrates.org",
		TPid:      testTPID,
		ID:        "ResGroup1",
		FilterIDs: []string{"FilterID1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2015-07-29T15:00:00Z",
		},
		UsageTTL:          "Test_TTL",
		Weight:            10,
		Limit:             "2",
		ThresholdIDs:      []string{"TRes1", "TRes2"},
		AllocationMessage: "test",
	}
	expStruct := ResourceMdls{{
		Tenant:             "cgrates.org",
		Tpid:               testTPID,
		ID:                 "ResGroup1",
		FilterIDs:          "FilterID1",
		ActivationInterval: "2014-07-29T15:00:00Z;2015-07-29T15:00:00Z",
		UsageTTL:           "Test_TTL",
		Weight:             10,
		Limit:              "2",
		ThresholdIDs:       "TRes1;TRes2",
		AllocationMessage:  "test",
	}}
	result := APItoModelResource(testStruct)
	if !reflect.DeepEqual(expStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestRouteProfileToAPICase1(t *testing.T) {
	structTest := &RouteProfile{
		FilterIDs: []string{"FilterID1", "FilterID2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2020, time.April,
				11, 21, 34, 01, 0, time.UTC),
			ExpiryTime: time.Date(2020, time.April,
				12, 21, 34, 01, 0, time.UTC),
		},
		SortingParameters: []string{"Param1", "Param2"},
		Routes: []*Route{
			{ID: "ResGroup2"},
		},
	}

	expStruct := &utils.TPRouteProfile{
		FilterIDs: []string{"FilterID1", "FilterID2"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2020-04-11T21:34:01Z",
			ExpiryTime:     "2020-04-12T21:34:01Z",
		},
		SortingParameters: []string{"Param1", "Param2"},
		Routes: []*utils.TPRoute{{
			ID: "ResGroup2",
		}},
	}

	result := RouteProfileToAPI(structTest)
	sort.Strings(result.FilterIDs)
	sort.Strings(result.SortingParameters)
	if !reflect.DeepEqual(expStruct, result) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestDispatcherProfileToAPICase2(t *testing.T) {
	structTest := &DispatcherProfile{
		Subsystems: []string{},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 35, 0, 0, time.UTC),
		},
		FilterIDs: []string{"field1", "field2"},
		StrategyParams: map[string]any{
			"Field1": "Params1",
		},
		Hosts: []*DispatcherHostProfile{
			{
				FilterIDs: []string{"fieldA", "fieldB"},
				Params:    map[string]any{},
			},
		},
	}

	expStruct := &utils.TPDispatcherProfile{
		Subsystems: []string{},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		FilterIDs:      []string{"field1", "field2"},
		StrategyParams: []any{"Params1"},
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				FilterIDs: []string{"fieldA", "fieldB"},
				Params:    []any{},
			},
		},
	}

	result := DispatcherProfileToAPI(structTest)
	sort.Strings(result.FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoDispatcherProfileCase2(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		Subsystems:     []string{},
		FilterIDs:      []string{},
		StrategyParams: []any{"Param1"},
		Hosts: []*utils.TPDispatcherHostProfile{{
			Params: []any{"Param1"},
		}},
	}
	expStruct := &DispatcherProfile{
		Subsystems: []string{},
		FilterIDs:  []string{},
		StrategyParams: map[string]any{
			"0": "Param1",
		},
		Hosts: DispatcherHostProfiles{{
			FilterIDs: []string{},
			Params: map[string]any{
				"0": "Param1",
			},
		},
		},
	}
	result, _ := APItoDispatcherProfile(structTest, "")
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoDispatcherProfileError(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat1",
			ExpiryTime:     "cat2",
		},
		StrategyParams: []any{"Param1"},
		Hosts: []*utils.TPDispatcherHostProfile{{
			Params: []any{""},
		}},
	}

	_, err := APItoDispatcherProfile(structTest, "")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpected <Unsupported time format>,\n Received <%+v>", err)
	}
}

func TestAPItoModelTPDispatcherProfileNil(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{}
	structTest = nil
	expected := "null"
	result := APItoModelTPDispatcherProfile(structTest)
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestModelHelpersParamsToString(t *testing.T) {
	testInterface := []any{"Param1", "Param2"}
	result := paramsToString(testInterface)
	if !reflect.DeepEqual(result, "Param1;Param2") {
		t.Errorf("\nExpecting <Param1;Param2>,\n Received <%+v>", result)
	}
}

func TestModelHelpersAsTPDispatcherProfiles(t *testing.T) {
	structTest := DispatcherProfileMdls{
		&DispatcherProfileMdl{
			ActivationInterval: "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z",
			StrategyParameters: "Param1",
		},
	}
	expStruct := []*utils.TPDispatcherProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2014-08-29T15:00:00Z",
		},
		StrategyParams: []any{"Param1"},
	},
	}
	result := structTest.AsTPDispatcherProfiles()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestTPDispatcherProfilesCSVHeader(t *testing.T) {
	structTest := DispatcherProfileMdls{
		&DispatcherProfileMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Dsp",
			Subsystems:         "*any",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			Strategy:           utils.MetaFirst,
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
			ConnID:             "C1",
			ConnWeight:         10,
			ConnBlocker:        false,
			ConnParameters:     "192.168.54.203",
		},
		&DispatcherProfileMdl{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			ConnID:         "C2",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.204",
		},
	}
	expected := []string{"#" + utils.Tenant, utils.ID, utils.Subsystems, utils.FilterIDs, utils.ActivationIntervalString,
		utils.Strategy, utils.StrategyParameters, utils.ConnID, utils.ConnFilterIDs,
		utils.ConnWeight, utils.ConnBlocker, utils.ConnParameters, utils.Weight}
	result := structTest.CSVHeader()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestModelHelpersAPItoChargerProfilError(t *testing.T) {
	structTest := &utils.TPChargerProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat1",
			ExpiryTime:     "cat2",
		},
	}
	_, err := APItoChargerProfile(structTest, "")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpected <Unsupported time format>,\n Received <%+v>", err)
	}
}

func TestChargerProfileToAPILastCase(t *testing.T) {
	testStruct := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CPP_1",
		FilterIDs: []string{"*string:~*opts.*subsys:*chargers", "FLTR_CP_1", "FLTR_CP_4"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		RunID:        "TestRunID",
		AttributeIDs: []string{"*none"},
		Weight:       20,
	}

	expStruct := &utils.TPChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "CPP_1",
		FilterIDs: []string{"*string:~*opts.*subsys:*chargers", "FLTR_CP_1", "FLTR_CP_4"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:25:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"*none"},
		RunID:        "TestRunID",
		Weight:       20,
	}

	result := ChargerProfileToAPI(testStruct)
	sort.Strings(result.FilterIDs)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoModelTPDispatcherProfileCase2(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2014-07-30T15:00:00Z",
		},
	}
	expStruct := DispatcherProfileMdls{{
		ActivationInterval: "2014-07-29T15:00:00Z;2014-07-30T15:00:00Z",
	},
	}
	result := APItoModelTPDispatcherProfile(structTest)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func ModelHelpersTestStatMdlsCSVHeader(t *testing.T) {
	testStruct := ResourceMdls{
		{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "ResGroup1",
			FilterIDs:          "FLTR_RES_GR1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			Stored:             false,
			Blocker:            false,
			Weight:             10.0,
			Limit:              "45",
			ThresholdIDs:       "WARN_RES1;WARN_RES1",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.UsageTTL, utils.Limit, utils.AllocationMessage, utils.Blocker, utils.Stored,
		utils.Weight, utils.ThresholdIDs}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestThresholdMdlsCSVHeader(t *testing.T) {
	testStruct := ThresholdMdls{
		{
			Tpid:   "test_tpid",
			Tenant: "test_tenant",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.MaxHits, utils.MinHits, utils.MinSleep,
		utils.Blocker, utils.Weight, utils.ActionIDs, utils.Async}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestChargerMdlsCSVHeader(t *testing.T) {

	testStruct := ChargerMdls{
		{
			Tenant: "cgrates.org",
			ID:     "RP1",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.RunID, utils.AttributeIDs, utils.Weight}

	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAPItoAttributeProfileError1(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  "",
				Value: "Al1",
			},
		},
		Weight: 20,
	}

	_, err := APItoAttributeProfile(tpAlsPrf, "UTC")
	if err == nil || err.Error() != "empty path in AttributeProfile <cgrates.org:ALS1>" {
		t.Errorf("\nExpecting <empty path in AttributeProfile <cgrates.org:ALS1>>,\n Received <%+v>", err)
	}

}

func TestAPItoAttributeProfileError2(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "\"constant;`>;q=0.7;expires=3600constant\"",
			},
		},
		Weight: 20,
	}

	_, err := APItoAttributeProfile(tpAlsPrf, "UTC")
	if err == nil || err.Error() != "Unclosed unspilit syntax" {
		t.Errorf("\nExpecting <Unclosed unspilit syntax>,\n Received <%+v>", err)
	}

}

func TestAPItoAttributeProfileError3(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weight: 20,
	}

	_, err := APItoAttributeProfile(tpAlsPrf, "UTC")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Received <%+v>", err)
	}
}

func TestAPItoModelTPAttributeNoAttributes(t *testing.T) {
	testStruct := &utils.TPAttributeProfile{}
	expStruct := AttributeMdls{}
	expStruct = nil
	result := APItoModelTPAttribute(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestAttributeMdlsCSVHeader(t *testing.T) {
	testStruct := AttributeMdls{
		{
			Tenant: "cgrates.org",
			ID:     "ALS1",
		},
	}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.AttributeFilterIDs, utils.Path, utils.Type, utils.Value, utils.Blocker, utils.Weight}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersTestAPItoRouteProfile(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*utils.TPRoute{},
	}
	expStruct := &RouteProfile{
		FilterIDs:         []string{},
		SortingParameters: []string{"param1"},
		Routes:            []*Route{},
	}
	result, err := APItoRouteProfile(testStruct, "")
	if err != nil {
		t.Errorf("\nExpecting <nil>,\n Received <%+v>", err)
	}
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}
func TestModelHelpersTestAPItoRouteProfileErr(t *testing.T) {
	testStruct := &utils.TPRouteProfile{
		FilterIDs: []string{},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat",
		},
		SortingParameters: []string{"param1"},
		Routes:            []*utils.TPRoute{},
	}
	_, err := APItoRouteProfile(testStruct, "")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Received <%+v>", err)
	}

}

func TestModelHelperAPItoFilterError(t *testing.T) {
	testStruct := &utils.TPFilterProfile{
		Filters: []*utils.TPFilter{{
			Type:    "test_type",
			Element: "",
			Values:  []string{"val1"},
		},
		},
	}

	_, err := APItoFilter(testStruct, "")
	if err == nil || err.Error() != "empty RSRParser in rule: <>" {
		t.Errorf("\nExpecting <empty RSRParser in rule: <>>,\n Received <%+v>", err)
	}

}

func TestModelHelperAPItoFilterError2(t *testing.T) {
	testStruct := &utils.TPFilterProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat",
		},
	}

	_, err := APItoFilter(testStruct, "")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Received <%+v>", err)
	}

}

func TestFilterMdlsCSVHeader(t *testing.T) {
	testStruct := FilterMdls{{
		Tpid:   "test_tpid",
		Tenant: "test_tenant",
	}}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.Type, utils.Element,
		utils.Values, utils.ActivationIntervalString}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestModelHelpersThresholdProfileToAPIExpTime(t *testing.T) {
	testStruct := &ThresholdProfile{
		FilterIDs: []string{"test_filter_id"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 25, 0, 0, time.UTC),
		},
		ActionIDs: []string{"test_action_id"},
	}
	expStruct := &utils.TPThresholdProfile{
		FilterIDs: []string{"test_filter_id"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:25:00Z",
			ExpiryTime:     "2014-07-15T14:25:00Z",
		},
		ActionIDs: []string{"test_action_id"},
	}
	result := ThresholdProfileToAPI(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoThresholdProfileError1(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:               "",
		Tenant:             "",
		ID:                 "",
		FilterIDs:          nil,
		ActivationInterval: nil,
		MaxHits:            0,
		MinHits:            0,
		MinSleep:           "cat",
		Blocker:            false,
		Weight:             0,
		ActionIDs:          nil,
		Async:              false,
	}
	_, err := APItoThresholdProfile(testStruct, "")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\nExpecting <time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoThresholdProfileError2(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:      "",
		Tenant:    "",
		ID:        "",
		FilterIDs: nil,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat",
		},
		MaxHits:   0,
		MinHits:   0,
		MinSleep:  "",
		Blocker:   false,
		Weight:    0,
		ActionIDs: nil,
		Async:     false,
	}
	_, err := APItoThresholdProfile(testStruct, "")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoModelTPThresholdExpTime1(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3", "LOG"},
	}
	expStruct := ThresholdMdls{
		{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
		{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			ActionIDs: "LOG",
		},
	}

	result := APItoModelTPThreshold(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoModelTPThresholdExpTime2(t *testing.T) {
	testStruct := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FilterID1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "2014-07-15T14:35:00Z",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3"},
	}
	expStruct := ThresholdMdls{
		{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			FilterIDs:          "FilterID1",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
	}

	result := APItoModelTPThreshold(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestThresholdMdlsAsTPThresholdActivationTime(t *testing.T) {
	testStruct := ThresholdMdls{
		{
			Tpid:               "",
			Tenant:             "",
			ID:                 "",
			FilterIDs:          "",
			ActivationInterval: "2014-07-14T14:35:00Z;2014-07-15T14:35:00Z",
			MaxHits:            0,
			MinHits:            0,
			MinSleep:           "",
			Blocker:            false,
			Weight:             0,
			ActionIDs:          "",
			Async:              false,
		},
	}
	expStruct := []*utils.TPThresholdProfile{
		{
			TPid:   "",
			Tenant: "",
			ID:     "",
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-14T14:35:00Z",
				ExpiryTime:     "2014-07-15T14:35:00Z",
			},
			MaxHits:  0,
			MinHits:  0,
			MinSleep: "",
			Blocker:  false,
			Weight:   0,
			Async:    false,
		},
	}
	result := testStruct.AsTPThreshold()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersStatQueueProfileToAPIFilterIds(t *testing.T) {
	testStruct := &StatQueueProfile{
		Tenant:    "",
		ID:        "",
		FilterIDs: []string{"test_filter_Id"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 0,
		MinItems:    0,
		Metrics: []*MetricWithFilters{{
			FilterIDs: []string{"test_id"},
		},
		},
		Stored:       false,
		Blocker:      false,
		Weight:       0,
		ThresholdIDs: []string{"threshold_id"},
	}
	expStruct := &utils.TPStatProfile{
		Tenant:    "",
		ID:        "",
		FilterIDs: []string{"test_filter_Id"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:25:00Z",
			ExpiryTime:     "2014-07-15T14:25:00Z",
		},
		QueueLength: 0,
		MinItems:    0,
		Metrics: []*utils.MetricWithFilters{
			{
				FilterIDs: []string{"test_id"},
			},
		},
		Blocker:      false,
		Stored:       false,
		Weight:       0,
		ThresholdIDs: []string{"threshold_id"},
	}
	result := StatQueueProfileToAPI(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoStatsError1(t *testing.T) {
	testStruct := &utils.TPStatProfile{
		TPid:               "",
		Tenant:             "",
		ID:                 "",
		FilterIDs:          nil,
		ActivationInterval: nil,
		QueueLength:        0,
		TTL:                "cat",
		Metrics:            nil,
		Blocker:            false,
		Stored:             false,
		Weight:             0,
		MinItems:           0,
		ThresholdIDs:       nil,
	}
	_, err := APItoStats(testStruct, "")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\nExpecting <time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoStatsError2(t *testing.T) {
	testStruct := &utils.TPStatProfile{
		TPid:      "",
		Tenant:    "",
		ID:        "",
		FilterIDs: nil,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat",
			ExpiryTime:     "cat",
		},
		QueueLength:  0,
		TTL:          "",
		Metrics:      nil,
		Blocker:      false,
		Stored:       false,
		Weight:       0,
		MinItems:     0,
		ThresholdIDs: nil,
	}
	_, err := APItoStats(testStruct, "")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoModelStatsCase2(t *testing.T) {
	testStruct := &utils.TPStatProfile{
		TPid:      "TPS1",
		Tenant:    "cgrates.org",
		ID:        "Stat1",
		FilterIDs: []string{"*string:Account:1002", "*string:Account:1003"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-25T15:00:00Z",
			ExpiryTime:     "2014-07-26T15:00:00Z",
		},
		QueueLength: 100,
		TTL:         "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				FilterIDs: []string{"test_filter_id1", "test_filter_id2"},
				MetricID:  "*tcc",
			},
		},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     2,
		ThresholdIDs: []string{"Th1", "Th2"},
	}
	expStruct := StatMdls{
		&StatMdl{
			Tpid:               "TPS1",
			Tenant:             "cgrates.org",
			ID:                 "Stat1",
			FilterIDs:          "*string:Account:1002;*string:Account:1003",
			ActivationInterval: "2014-07-25T15:00:00Z;2014-07-26T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*tcc",
			MetricFilterIDs:    "test_filter_id1;test_filter_id2",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
			ThresholdIDs:       "Th1;Th2",
		},
	}
	result := APItoModelStats(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestStatMdlsAsTPStatsCase2(t *testing.T) {
	testStruct := StatMdls{{
		ActivationInterval: "2014-07-25T15:00:00Z;2014-07-26T15:00:00Z",
		MetricIDs:          "test_id",
		MetricFilterIDs:    "test_filter_id",
	}}
	expStruct := []*utils.TPStatProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-25T15:00:00Z",
			ExpiryTime:     "2014-07-26T15:00:00Z",
		},
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID:  "test_id",
				FilterIDs: []string{"test_filter_id"},
			},
		},
	}}
	result := testStruct.AsTPStats()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestStatMdlsCSVHeader(t *testing.T) {
	testStruct := StatMdls{{
		PK:                 0,
		Tpid:               "",
		Tenant:             "test_tenant",
		ID:                 "test_id",
		FilterIDs:          "test_filter_id",
		ActivationInterval: "test_interval",
		QueueLength:        0,
		TTL:                "",
		MinItems:           0,
		MetricIDs:          "",
		MetricFilterIDs:    "",
		Stored:             false,
		Blocker:            false,
		Weight:             0,
		ThresholdIDs:       "",
		CreatedAt:          time.Time{},
	}}
	expStruct := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.QueueLength, utils.TTL, utils.MinItems, utils.MetricIDs, utils.MetricFilterIDs,
		utils.Stored, utils.Blocker, utils.Weight, utils.ThresholdIDs}
	result := testStruct.CSVHeader()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersResourceProfileToAPICase2(t *testing.T) {
	testStruct := &ResourceProfile{
		Tenant:    "",
		ID:        "",
		FilterIDs: []string{"test_filter_id"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Second,
		Limit:             0,
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weight:            0,
		ThresholdIDs:      []string{"test_threshold_id"},
	}
	expStruct := &utils.TPResourceProfile{
		TPid:      "",
		Tenant:    "",
		ID:        "",
		FilterIDs: []string{"test_filter_id"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:25:00Z",
			ExpiryTime:     "2014-07-15T14:25:00Z",
		},
		UsageTTL:          "1s",
		Limit:             "0",
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weight:            0,
		ThresholdIDs:      []string{"test_threshold_id"},
	}
	result := ResourceProfileToAPI(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersAPItoResourceError1(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		TPid:               "",
		Tenant:             "",
		ID:                 "",
		FilterIDs:          nil,
		ActivationInterval: nil,
		UsageTTL:           "cat",
		Limit:              "",
		AllocationMessage:  "",
		Blocker:            false,
		Stored:             false,
		Weight:             0,
		ThresholdIDs:       nil,
	}
	_, err := APItoResource(testStruct, "")
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\nExpecting <time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoResourceError2(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		TPid:      "",
		Tenant:    "",
		ID:        "",
		FilterIDs: nil,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat",
		},
		UsageTTL:          "",
		Limit:             "",
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weight:            0,
		ThresholdIDs:      nil,
	}
	_, err := APItoResource(testStruct, "")
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoResourceError3(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		TPid:              "",
		Tenant:            "",
		ID:                "",
		FilterIDs:         nil,
		UsageTTL:          "",
		Limit:             "cat",
		AllocationMessage: "",
		Blocker:           false,
		Stored:            false,
		Weight:            0,
		ThresholdIDs:      nil,
	}
	_, err := APItoResource(testStruct, "")
	if err == nil || err.Error() != "strconv.ParseFloat: parsing \"cat\": invalid syntax" {
		t.Errorf("\nExpecting <strconv.ParseFloat: parsing \"cat\": invalid syntax>,\n Received <%+v>", err)
	}
}

func TestTpResourcesAsTpResources2(t *testing.T) {
	testStruct := []*ResourceMdl{
		{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "ResGroup1",
			FilterIDs:          "FLTR_RES_GR1",
			ActivationInterval: "2014-07-27T15:00:00Z;2014-07-28T15:00:00Z",
			ThresholdIDs:       "WARN_RES1",
		},
	}
	expStruct := []*utils.TPResourceProfile{
		{
			TPid:      "TEST_TPID",
			Tenant:    "cgrates.org",
			ID:        "ResGroup1",
			FilterIDs: []string{"FLTR_RES_GR1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-27T15:00:00Z",
				ExpiryTime:     "2014-07-28T15:00:00Z",
			},
			ThresholdIDs: []string{"WARN_RES1"},
		},
	}
	result := ResourceMdls(testStruct).AsTPResources()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersMapTPAccountActionsError(t *testing.T) {
	testStruct := []*utils.TPAccountActions{
		{
			TPid:             "ee",
			LoadId:           "ee",
			Tenant:           "",
			Account:          "",
			ActionPlanId:     "",
			ActionTriggersId: "",
			AllowNegative:    false,
			Disabled:         false,
		},
		{
			TPid:             "ee",
			LoadId:           "ee",
			Tenant:           "",
			Account:          "",
			ActionPlanId:     "",
			ActionTriggersId: "",
			AllowNegative:    false,
			Disabled:         false,
		},
	}

	_, err := MapTPAccountActions(testStruct)
	if err == nil || err.Error() != "Non unique ID :" {
		t.Errorf("\nExpecting <Non unique ID :>,\n Received <%+v>", err)
	}
}

func TestModelHelpersMapTPSharedGroup2(t *testing.T) {
	testStruct := []*utils.TPSharedGroups{
		{
			TPid: "",
			ID:   "2",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "",
					Strategy:      "",
					RatingSubject: "",
				},
			},
		},
		{
			TPid: "",
			ID:   "2",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "",
					Strategy:      "",
					RatingSubject: "",
				},
			},
		},
	}
	expStruct := map[string][]*utils.TPSharedGroup{
		"2": {
			{
				Account:       "",
				Strategy:      "",
				RatingSubject: "",
			},
			{
				Account:       "",
				Strategy:      "",
				RatingSubject: "",
			},
		},
	}
	result := MapTPSharedGroup(testStruct)
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}

}

func TestSharedGroupMdlsAsMapTPSharedGroups2(t *testing.T) {
	testStruct := SharedGroupMdls{
		{
			Id:            2,
			Tpid:          "2",
			Tag:           "",
			Account:       "",
			Strategy:      "",
			RatingSubject: "",
		},
		{
			Id:            2,
			Tpid:          "2",
			Tag:           "",
			Account:       "",
			Strategy:      "",
			RatingSubject: "",
		},
	}
	expStruct := map[string]*utils.TPSharedGroups{
		"": {
			TPid: "2",
			ID:   "",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "",
					Strategy:      "",
					RatingSubject: "",
				},
				{
					Account:       "",
					Strategy:      "",
					RatingSubject: "",
				},
			},
		},
	}
	result := testStruct.AsMapTPSharedGroups()
	if !reflect.DeepEqual(result, expStruct) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expStruct), utils.ToJSON(result))
	}
}

func TestModelHelpersMapTPRatingProfilesError(t *testing.T) {
	testStruct := []*utils.TPRatingProfile{
		{
			TPid:                  "2",
			LoadId:                "",
			Tenant:                "",
			Category:              "",
			Subject:               "",
			RatingPlanActivations: nil,
		},
		{
			TPid:                  "2",
			LoadId:                "",
			Tenant:                "",
			Category:              "",
			Subject:               "",
			RatingPlanActivations: nil,
		},
	}
	_, err := MapTPRatingProfiles(testStruct)
	if err == nil || err.Error() != "Non unique id :*out:::" {
		t.Errorf("\nExpecting <Non unique id :*out:::>,\n Received <%+v>", err)
	}
}

func TestModelHelpersCSVLoadErrorInt(t *testing.T) {
	type testStruct struct {
		Id        int64
		Tpid      string
		Tag       int `index:"0" re:".*"`
		CreatedAt time.Time
	}

	_, err := csvLoad(testStruct{}, []string{"TEST_DEST"})
	if err == nil || err.Error() != "invalid value \"TEST_DEST\" for field testStruct.Tag" {
		t.Errorf("\nExpecting <invalid value \"TEST_DEST\" for field testStruct.Tag>,\n Received <%+v>", err)
	}
}

func TestModelHelpersCSVLoadErrorFloat64(t *testing.T) {
	type testStruct struct {
		Id        int64
		Tpid      string
		Tag       float64 `index:"0" re:".*"`
		CreatedAt time.Time
	}

	_, err := csvLoad(testStruct{}, []string{"TEST_DEST"})
	if err == nil || err.Error() != "invalid value \"TEST_DEST\" for field testStruct.Tag" {
		t.Errorf("\nExpecting <invalid value \"TEST_DEST\" for field testStruct.Tag>,\n Received <%+v>", err)
	}
}

func TestModelHelpersCSVLoadErrorBool(t *testing.T) {
	type testStruct struct {
		Id        int64
		Tpid      string
		Tag       bool `index:"0" re:".*"`
		CreatedAt time.Time
	}

	_, err := csvLoad(testStruct{}, []string{"TEST_DEST"})
	if err == nil || err.Error() != "invalid value \"TEST_DEST\" for field testStruct.Tag" {
		t.Errorf("\nExpecting <invalid value \"TEST_DEST\" for field testStruct.Tag>,\n Received <%+v>", err)
	}
}

func TestCSVHeaders(t *testing.T) {
	expected := []string{
		"#" + utils.Tenant,
		utils.ID,
		utils.Schedule,
		utils.StatIDs,
		utils.MetricIDs,
		utils.Sorting,
		utils.SortingParameters,
		utils.Stored,
		utils.ThresholdIDs,
	}
	var tps RankingsMdls
	result := tps.CSVHeader()
	if len(result) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(result))
		return
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected value at index %d to be %s, got %s", i, v, result[i])
		}
	}
}

func TestTrendsMdlCSVHeader(t *testing.T) {
	expected := []string{
		"#" + utils.Tenant,
		utils.ID,
		utils.Schedule,
		utils.StatID,
		utils.Metrics,
		utils.TTL,
		utils.QueueLength,
		utils.MinItems,
		utils.CorrelationType,
		utils.Tolerance,
		utils.Stored,
		utils.ThresholdIDs,
	}
	var tps TrendsMdls
	result := tps.CSVHeader()
	if len(result) != len(expected) {
		t.Errorf("Expected %d elements, got %d", len(expected), len(result))
		return
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("Expected value at index %d to be %s, got %s", i, v, result[i])
		}
	}
}

func TestRankingProfileToAPI(t *testing.T) {
	sg := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "1001",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{"metric1", "metric2"},
		SortingParameters: []string{"sort1", "sort2"},
		ThresholdIDs:      []string{"threshold1", "threshold2"},
		Schedule:          "@every 30m",
	}

	expected := &utils.TPRankingProfile{
		Tenant:            "cgrates.org",
		ID:                "1001",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{"metric1", "metric2"},
		SortingParameters: []string{"sort1", "sort2"},
		ThresholdIDs:      []string{"threshold1", "threshold2"},
		Schedule:          "@every 30m",
	}
	result := RankingProfileToAPI(sg)
	if result.Tenant != expected.Tenant {
		t.Errorf("Expected Tenant %s, got %s", expected.Tenant, result.Tenant)
	}
	if result.ID != expected.ID {
		t.Errorf("Expected ID %s, got %s", expected.ID, result.ID)
	}
	if len(result.StatIDs) != len(expected.StatIDs) {
		t.Errorf("Expected %d StatIDs, got %d", len(expected.StatIDs), len(result.StatIDs))
	} else {
		for i, v := range expected.StatIDs {
			if result.StatIDs[i] != v {
				t.Errorf("Expected StatID at index %d to be %s, got %s", i, v, result.StatIDs[i])
			}
		}
	}
	if len(result.MetricIDs) != len(expected.MetricIDs) {
		t.Errorf("Expected %d MetricIDs, got %d", len(expected.MetricIDs), len(result.MetricIDs))
	} else {
		for i, v := range expected.MetricIDs {
			if result.MetricIDs[i] != v {
				t.Errorf("Expected MetricID at index %d to be %s, got %s", i, v, result.MetricIDs[i])
			}
		}
	}
	if len(result.SortingParameters) != len(expected.SortingParameters) {
		t.Errorf("Expected %d SortingParameters, got %d", len(expected.SortingParameters), len(result.SortingParameters))
	} else {
		for i, v := range expected.SortingParameters {
			if result.SortingParameters[i] != v {
				t.Errorf("Expected SortingParameter at index %d to be %s, got %s", i, v, result.SortingParameters[i])
			}
		}
	}
	if len(result.ThresholdIDs) != len(expected.ThresholdIDs) {
		t.Errorf("Expected %d ThresholdIDs, got %d", len(expected.ThresholdIDs), len(result.ThresholdIDs))
	} else {
		for i, v := range expected.ThresholdIDs {
			if result.ThresholdIDs[i] != v {
				t.Errorf("Expected ThresholdID at index %d to be %s, got %s", i, v, result.ThresholdIDs[i])
			}
		}
	}
	if result.Schedule != expected.Schedule {
		t.Errorf("Expected QueryInterval %s, got %s", expected.Schedule, result.Schedule)
	}
}

func TestAPItoModelTPRanking(t *testing.T) {
	tests := []struct {
		name     string
		input    *utils.TPRankingProfile
		expected RankingsMdls
	}{
		{
			name:     "Nil Input",
			input:    nil,
			expected: RankingsMdls{},
		},
		{
			name: "No StatIDs",
			input: &utils.TPRankingProfile{
				TPid:              "tpid1",
				Tenant:            "cgrates.org",
				ID:                "id1",
				Schedule:          "1h",
				Sorting:           "asc",
				ThresholdIDs:      []string{"threshold1", "threshold2"},
				MetricIDs:         []string{"metric1", "metric2"},
				SortingParameters: []string{"param1", "param2"},
				StatIDs:           []string{},
			},
			expected: RankingsMdls{
				&RankingsMdl{
					Tpid:              "tpid1",
					Tenant:            "cgrates.org",
					ID:                "id1",
					Schedule:          "1h",
					Sorting:           "asc",
					StatIDs:           "",
					ThresholdIDs:      "threshold1" + utils.InfieldSep + "threshold2",
					MetricIDs:         "metric1" + utils.InfieldSep + "metric2",
					SortingParameters: "param1" + utils.InfieldSep + "param2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := APItoModelTPRanking(tt.input)

			if len(actual) != len(tt.expected) {
				t.Errorf("Expected %d models, got %d", len(tt.expected), len(actual))
				return
			}

			for i := range tt.expected {
				if *tt.expected[i] != *actual[i] {
					t.Errorf("Expected model %+v, got %+v", *tt.expected[i], *actual[i])
				}
			}
		})
	}
}

func TestAPItoModelTrends(t *testing.T) {
	tests := []struct {
		name     string
		input    *utils.TPTrendsProfile
		expected TrendsMdls
	}{
		{
			name:     "Nil Input",
			input:    nil,
			expected: TrendsMdls{},
		},
		{
			name: "Valid Input",
			input: &utils.TPTrendsProfile{
				TPid:            "tpid1",
				Tenant:          "cgrates.org",
				ID:              "id1",
				Schedule:        "daily",
				QueueLength:     10,
				StatID:          "stat1",
				TTL:             "3600",
				MinItems:        5,
				CorrelationType: "type1",
				Tolerance:       0.1,
				Stored:          true,
				ThresholdIDs:    []string{"threshold1", "threshold2"},
				Metrics:         []string{"metric1", "metric2"},
			},
			expected: TrendsMdls{
				&TrendsMdl{
					Tpid:            "tpid1",
					Tenant:          "cgrates.org",
					ID:              "id1",
					Schedule:        "daily",
					QueueLength:     10,
					StatID:          "stat1",
					TTL:             "3600",
					MinItems:        5,
					CorrelationType: "type1",
					Tolerance:       0.1,
					Stored:          true,
					ThresholdIDs:    "threshold1" + utils.InfieldSep + "threshold2",
					Metrics:         "metric1" + utils.InfieldSep + "metric2",
					CreatedAt:       time.Time{},
				},
			},
		},
		{
			name: "Empty ThresholdIDs and Metrics",
			input: &utils.TPTrendsProfile{
				TPid:            "tpid2",
				Tenant:          "tenant2",
				ID:              "id2",
				Schedule:        "weekly",
				QueueLength:     15,
				StatID:          "stat2",
				TTL:             "7200",
				MinItems:        10,
				CorrelationType: "type2",
				Tolerance:       0.2,
				Stored:          false,
				ThresholdIDs:    []string{},
				Metrics:         []string{},
			},
			expected: TrendsMdls{
				&TrendsMdl{
					Tpid:            "tpid2",
					Tenant:          "tenant2",
					ID:              "id2",
					Schedule:        "weekly",
					QueueLength:     15,
					StatID:          "stat2",
					TTL:             "7200",
					MinItems:        10,
					CorrelationType: "type2",
					Tolerance:       0.2,
					Stored:          false,
					ThresholdIDs:    "",
					Metrics:         "",
					CreatedAt:       time.Time{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := APItoModelTrends(tt.input)

			if len(actual) != len(tt.expected) {
				t.Errorf("Expected %d models, got %d", len(tt.expected), len(actual))
				return
			}

			for i := range tt.expected {
				if *tt.expected[i] != *actual[i] {
					t.Errorf("Expected model %+v, got %+v", *tt.expected[i], *actual[i])
				}
			}
		})
	}
}
