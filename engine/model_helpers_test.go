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
	l, err := csvLoad(TpDestinationMdl{}, []string{"TEST_DEST", "+492"})
	tpd, ok := l.(TpDestinationMdl)
	if err != nil || !ok || tpd.Tag != "TEST_DEST" || tpd.Prefix != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestModelHelperCsvDump(t *testing.T) {
	tpd := TpDestinationMdl{
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
	in := &TpDestinations{}
	eOut := map[string]*Destination{}

	if rcv, err := in.AsMapDestinations(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
	in = &TpDestinations{
		TpDestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST1", Prefix: "+491"},
		TpDestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST2", Prefix: "+492"},
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
	in = &TpDestinations{
		TpDestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST1", Prefix: "+491"},
		TpDestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST2", Prefix: "+492"},
		TpDestinationMdl{Tpid: "TEST_ID", Tag: "", Prefix: ""},
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
	eOut := TpDestinations{
		TpDestinationMdl{},
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
	eOut = TpDestinations{
		TpDestinationMdl{
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
	tpd1 := TpDestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+491"}
	tpd2 := TpDestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+492"}
	tpd3 := TpDestinationMdl{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+493"}
	eTPDestinations := []*utils.TPDestination{{TPid: "TEST_TPID", ID: "TEST_DEST",
		Prefixes: []string{"+491", "+492", "+493"}}}
	if tpDst := TpDestinations([]TpDestinationMdl{tpd1, tpd2, tpd3}).AsTPDestinations(); !reflect.DeepEqual(eTPDestinations, tpDst) {
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
	eOut := TpTimings{}
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
	eOut = TpTimings{
		TpTimingMdl{
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
	eOut = TpTimings{
		TpTimingMdl{
			Tpid:   "TPid1",
			Months: "1;2;3;4",
			Tag:    "ID1",
		},
		TpTimingMdl{
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
	eOut := TpRates{}
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
	eOut = TpRates{
		TpRateMdl{
			Tpid:       "TPid",
			Tag:        "SomeID",
			ConnectFee: 0.7,
			Rate:       0.8,
		},
		TpRateMdl{
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
	eOut = TpRates{
		TpRateMdl{
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
	eOut := TpDestinationRates{
		TpDestinationRateMdl{
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
	eOut := TpDestinationRates{
		TpDestinationRateMdl{
			Tpid:             "TEST_TPID",
			Tag:              "TEST_DSTRATE",
			DestinationsTag:  "TEST_DEST1",
			RatesTag:         "TEST_RATE1",
			RoundingMethod:   "*up",
			RoundingDecimals: 4,
		},
		TpDestinationRateMdl{
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
	pts := TpDestinationRates{}
	eOut := []*utils.TPDestinationRate{}
	if rcv, err := pts.AsTPDestinationRates(); err != nil {
		t.Error(err)
	} else if rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", utils.ToJSON(rcv))
	}

	pts = TpDestinationRates{
		TpDestinationRateMdl{
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
	if rcv, err := pts.AsTPDestinationRates(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
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
	eOut := TpRatingPlans{TpRatingPlanMdl{}}
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

	eOut = TpRatingPlans{
		TpRatingPlanMdl{
			Tpid:         "TEST_TPID",
			Tag:          "TEST_RPLAN",
			DestratesTag: "TEST_DSTRATE1",
			TimingTag:    "TEST_TIMING1",
			Weight:       10.0,
		},
		TpRatingPlanMdl{
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
	eOut = TpRatingPlans{
		TpRatingPlanMdl{
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
	eOut := TpRatingPlans{
		TpRatingPlanMdl{
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
	eOut := TpRatingProfiles{
		TpRatingProfileMdl{
			Tpid:             "TEST_TPID",
			Loadid:           "TEST_LOADID",
			Tenant:           "cgrates.org",
			Category:         "call",
			Subject:          "*any",
			RatingPlanTag:    "TEST_RPLAN1",
			FallbackSubjects: "subj1;subj2",
			ActivationTime:   "2014-01-14T00:00:00Z",
		},
		TpRatingProfileMdl{
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
	eOut = TpRatingProfiles{
		TpRatingProfileMdl{
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
	eOut := TpRatingProfiles{
		TpRatingProfileMdl{
			Tpid:     "TEST_TPID",
			Loadid:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "*any",
		},
		TpRatingProfileMdl{
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
	eOut := TpSharedGroups{
		TpSharedGroupMdl{
			Tpid:          "TEST_TPID",
			Tag:           "SHARED_GROUP_TEST",
			Account:       "*any",
			Strategy:      "*highest",
			RatingSubject: "special1",
		},
		TpSharedGroupMdl{
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
	eOut = TpSharedGroups{
		TpSharedGroupMdl{
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
	eOut = TpSharedGroups{
		TpSharedGroupMdl{
			Tpid:          "TEST_TPID",
			Tag:           "SHARED_GROUP_TEST",
			Account:       "*any",
			Strategy:      "*highest",
			RatingSubject: "special1",
		},
		TpSharedGroupMdl{
			Tpid:          "TEST_TPID",
			Tag:           "SHARED_GROUP_TEST",
			Account:       "*second",
			Strategy:      "*highest",
			RatingSubject: "special2",
		},
		TpSharedGroupMdl{
			Tpid:          "TEST_TPID2",
			Tag:           "SHARED_GROUP_TEST2",
			Account:       "*any",
			Strategy:      "*highest",
			RatingSubject: "special1",
		},
		TpSharedGroupMdl{
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

	eOut := TpActionPlans{
		TpActionPlanMdl{
			Tpid:       "TEST_TPID",
			Tag:        "PACKAGE_10",
			ActionsTag: "TOPUP_RST_10",
			TimingTag:  "ASAP",
			Weight:     10,
		},
		TpActionPlanMdl{
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
	eOut = TpActionPlans{
		TpActionPlanMdl{
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
	eOut := TpActionPlans{
		TpActionPlanMdl{
			Tpid:       "TEST_TPID",
			Tag:        "PACKAGE_10",
			ActionsTag: "TOPUP_RST_10",
			TimingTag:  "ASAP",
			Weight:     10,
		},
		TpActionPlanMdl{
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
	eOut = TpActionPlans{
		TpActionPlanMdl{
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
	eOut := TpActionTriggers{
		TpActionTriggerMdl{
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
		TpActionTriggerMdl{
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
	eOut = TpActionTriggers{
		TpActionTriggerMdl{
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
	eOut := TpActionTriggers{
		TpActionTriggerMdl{
			Tpid:          "TEST_TPID",
			Tag:           "STANDARD_TRIGGERS",
			UniqueId:      "1",
			ThresholdType: "*min_balance",
			Weight:        0.7,
		},
		TpActionTriggerMdl{
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
	eOut := TpActions{
		TpActionMdl{
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
	eOut = TpActions{
		TpActionMdl{
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
		TpActionMdl{
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
	eOut := TpActions{
		TpActionMdl{
			Tpid: "TEST_TPID",
			Tag:  "TEST_ACTIONS",
		},
		TpActionMdl{
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
	eOut = TpActions{
		TpActionMdl{
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
		TpActionMdl{
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
	tps := []*TpResourceMdl{
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
	rcvTPs := TpResources(tps).AsTPResources()
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
	expModel := &TpResourceMdl{
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
	tps := TpStats{
		&TpStatMdl{
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
		&TpStatMdl{
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
		&TpStatMdl{
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
	eRcv := TpStats{
		&TpStatMdl{
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
		&TpStatMdl{
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
	tps := []*TpThresholdMdl{
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
	rcvTPs := TpThresholds(tps).AsTPThreshold()
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
	eOut := TpAccountActions{
		TpAccountActionMdl{
			Tpid:              "TEST_TPID",
			Loadid:            "TEST_LOADID",
			Tenant:            "cgrates.org",
			Account:           "1001",
			ActionPlanTag:     "PACKAGE_10_SHARED_A_5",
			ActionTriggersTag: "STANDARD_TRIGGERS",
		},
		TpAccountActionMdl{
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
	var tps TpResources
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
	models := TpThresholds{
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
	models := TpThresholds{
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
	models := TpThresholds{
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
	models := TpThresholds{
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
	tps := []*TpFilterMdl{
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

	rcvTPs := TpFilterS(tps).AsTPFilter()
	if !(reflect.DeepEqual(eTPs, rcvTPs) || reflect.DeepEqual(eTPs[0], rcvTPs[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilterWithDynValues(t *testing.T) {
	tps := []*TpFilterMdl{
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

	rcvTPs := TpFilterS(tps).AsTPFilter()
	if !(reflect.DeepEqual(eTPs, rcvTPs) || reflect.DeepEqual(eTPs[0], rcvTPs[0])) {
		t.Errorf("Expecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilter2(t *testing.T) {
	tps := []*TpFilterMdl{
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

	rcvTPs := TpFilterS(tps).AsTPFilter()
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
	eOut := TpFilterS{
		&TpFilterMdl{
			ID:      "someID",
			Type:    "*prefix",
			Element: "Account",
			Values:  "1010",
		},
		&TpFilterMdl{
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
	eOut = TpFilterS{
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
	eOut = TpFilterS{
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
	var tps TPRoutes
	eOut := []string{
		"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.Sorting, utils.SortingParameters, utils.RouteID, utils.RouteFilterIDs,
		utils.RouteAccountIDs, utils.RouteRatingplanIDs, utils.RouteRateProfileIDs, utils.RouteResourceIDs,
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
				Value: config.NewRSRParsersMustCompile("Al1", utils.INFIELD_SEP),
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
				Value: config.NewRSRParsersMustCompile("Al1", utils.INFIELD_SEP),
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
				Value: config.NewRSRParsersMustCompile("Al1", utils.INFIELD_SEP),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Test",
				Value: config.NewRSRParsersMustCompile("~*req.Account", utils.INFIELD_SEP),
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
	expected := TPAttributes{
		&TPAttributeMdl{
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
	expected := TPAttributes{
		&TPAttributeMdl{
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
		&TPAttributeMdl{
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

func TestModelAsTPAttribute(t *testing.T) {
	models := TPAttributes{
		&TPAttributeMdl{
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
			ExpiryTime:     "",
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

//Number of FilterIDs and AttributeIDs are equal
func TestAPItoModelTPCharger(t *testing.T) {
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
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	expected := TPChargers{
		&TPChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&TPChargerMdl{
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

//Number of FilterIDs is smaller than AttributeIDs
func TestAPItoModelTPCharger2(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	expected := TPChargers{
		&TPChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&TPChargerMdl{
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

//Number of FilterIDs is greater than AttributeIDs
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
	expected := TPChargers{
		&TPChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&TPChargerMdl{
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

//len(AttributeIDs) is 0
func TestAPItoModelTPCharger4(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Weight: 20,
	}
	expected := TPChargers{
		&TPChargerMdl{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
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

//len(FilterIDs) is 0
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
	expected := TPChargers{
		&TPChargerMdl{
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

//both len(AttributeIDs) and len(FilterIDs) are 0
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
	expected := TPChargers{
		&TPChargerMdl{
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
	models := TPChargers{
		&TPChargerMdl{
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
	models := TPChargers{
		&TPChargerMdl{
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
		StrategyParams: []interface{}{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []interface{}{"192.168.54.203", "*ratio:2"},
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
		StrategyParams: map[string]interface{}{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.54.203", utils.MetaRatio: "2"},
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
		StrategyParams: []interface{}{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []interface{}{"192.168.54.203", "*ratio:2"},
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
		StrategyParams: []interface{}{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []interface{}{"*ratio:2", "192.168.54.203"},
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
		StrategyParams: map[string]interface{}{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]interface{}{"0": "192.168.54.203", utils.MetaRatio: "2"},
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
		StrategyParams: []interface{}{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []interface{}{"192.168.54.203"},
				Blocker:   false,
			},
			{
				ID:        "C2",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []interface{}{"192.168.54.204"},
				Blocker:   false,
			},
		},
	}
	expected := TPDispatcherProfiles{
		&TPDispatcherProfileMdl{
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
		&TPDispatcherProfileMdl{
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
	tps := &TPDispatcherHosts{}
	eOut := []string{"#" + utils.Tenant, utils.ID, utils.Address, utils.Transport, utils.TLS}
	if rcv := tps.CSVHeader(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}
}

func TestTPDispatcherHostsAsTPDispatcherHosts(t *testing.T) {
	tps := &TPDispatcherHosts{}
	if rcv := tps.AsTPDispatcherHosts(); rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tps = &TPDispatcherHosts{
		&TPDispatcherHostMdl{
			ID:     "ID1",
			Tenant: "Tenant1",
		}}
	if rcv := tps.AsTPDispatcherHosts(); rcv != nil {
		t.Errorf("Expecting: nil,\nReceived: %+v", utils.ToJSON(rcv))
	}

	tps = &TPDispatcherHosts{
		&TPDispatcherHostMdl{
			Address:   "Address1",
			ID:        "ID1",
			Tenant:    "Tenant1",
			Transport: utils.EmptyString,
		}}
	eOut := []*utils.TPDispatcherHost{
		{
			Tenant: "Tenant1",
			ID:     "ID1",
			Conn: &utils.TPDispatcherHostConn{
				Address:   "Address1",
				Transport: "*json",
			},
		},
	}
	if rcv := tps.AsTPDispatcherHosts(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &TPDispatcherHosts{
		&TPDispatcherHostMdl{
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
	if rcv := tps.AsTPDispatcherHosts(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &TPDispatcherHosts{
		&TPDispatcherHostMdl{
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
	if rcv := tps.AsTPDispatcherHosts(); !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

	tps = &TPDispatcherHosts{
		&TPDispatcherHostMdl{
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
	rcv := tps.AsTPDispatcherHosts()
	sort.Slice(rcv, func(i, j int) bool { return strings.Compare(rcv[i].ID, rcv[j].ID) < 0 })
	if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
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
			Address:   "Address1",
			Transport: "*json",
		},
	}
	eOut := &TPDispatcherHostMdl{
		Address:   "Address1",
		Transport: "*json",
		Tenant:    "Tenant",
		ID:        "ID",
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
			Address:   "Address1",
			Transport: "*json",
		},
	}

	eOut := &DispatcherHost{
		Tenant: "Tenant1",
		ID:     "ID1",
		Conn: &config.RemoteHost{
			Address:   "Address1",
			Transport: "*json",
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
		ID:     "ID2",
		Conn: &config.RemoteHost{
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
		ID:     "ID1",
		Conn: &config.RemoteHost{
			Address:   "Address1",
			Transport: "*json",
			TLS:       true,
		},
	}
	eOut := &utils.TPDispatcherHost{
		Tenant: "Tenant1",
		ID:     "ID1",
		Conn: &utils.TPDispatcherHostConn{
			Address:   "Address1",
			Transport: "*json",
			TLS:       true,
		},
	}
	if rcv := DispatcherHostToAPI(dph); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}

func TestTPRoutesAsTPRouteProfile(t *testing.T) {
	mdl := TPRoutes{
		&TpRouteMdl{
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
		&TpRouteMdl{
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

	mdlReverse := TPRoutes{
		&TpRouteMdl{
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
		&TpRouteMdl{
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

func TestRateProfileToAPI(t *testing.T) {
	rPrf := &RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						FixedFee:      0.00234,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	eTPRatePrf := &utils.TPRateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weight:    0,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "",
			ExpiryTime:     "",
		},
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.12,
						Unit:          "1m0s",
						Increment:     "1m0s",
					},
					{
						IntervalStart: "1m0s",
						FixedFee:      0.00234,
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
			"RT_WEEKEND": {
				ID:             "RT_WEEKEND",
				Weight:         10,
				ActivationTime: "* * * * 0,6",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:             "RT_CHRISTMAS",
				Weight:         30,
				ActivationTime: "* * 24 12 *",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
		},
	}
	if rcv := RateProfileToAPI(rPrf); !reflect.DeepEqual(rcv, eTPRatePrf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eTPRatePrf), utils.ToJSON(rcv))
	}
}

func TestAPIToRateProfile(t *testing.T) {
	eRprf := &RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						FixedFee:      2.3451,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	tpRprf := &utils.TPRateProfile{
		TPid:             "",
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						FixedFee:      2.3451,
						RecurrentFee:  0.12,
						Unit:          "1m0s",
						Increment:     "1m0s",
					},
					{
						IntervalStart: "1m0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
			"RT_WEEKEND": {
				ID:             "RT_WEEKEND",
				Weight:         10,
				ActivationTime: "* * * * 0,6",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:             "RT_CHRISTMAS",
				Weight:         30,
				ActivationTime: "* * 24 12 *",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
		},
	}
	if rcv, err := APItoRateProfile(tpRprf, utils.EmptyString); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, eRprf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eRprf), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPRateProfile(t *testing.T) {
	tpRprf := &utils.TPRateProfile{
		TPid:             "",
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.12,
						Unit:          "1m0s",
						Increment:     "1m0s",
					},
					{
						IntervalStart: "1m",
						RecurrentFee:  0.06,
						Unit:          "1m0s",
						Increment:     "1s",
					},
				},
			},
		},
	}

	expModels := RateProfileMdls{
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "*string:~*req.Subject:1001",
			ActivationInterval:  "",
			Weight:              0,
			RoundingMethod:      "*up",
			RoundingDecimals:    4,
			MinCost:             0.1,
			MaxCost:             0.6,
			MaxCostStrategy:     "*free",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationStart: "* * * * 1-5",
			RateWeight:          0,
			RateBlocker:         false,
			RateIntervalStart:   "1m",
			RateRecurrentFee:    0.06,
			RateUnit:            "1m0s",
			RateIncrement:       "1s",
			CreatedAt:           time.Time{},
		},
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "",
			ActivationInterval:  "",
			Weight:              0,
			RoundingMethod:      "",
			RoundingDecimals:    0,
			MinCost:             0,
			MaxCost:             0,
			MaxCostStrategy:     "",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationStart: "",
			RateWeight:          0,
			RateBlocker:         false,
			RateIntervalStart:   "0s",
			RateRecurrentFee:    0.12,
			RateUnit:            "1m0s",
			RateIncrement:       "1m0s",
			CreatedAt:           time.Time{},
		},
	}
	expModelsRev := RateProfileMdls{
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "*string:~*req.Subject:1001",
			ActivationInterval:  "",
			Weight:              0,
			RoundingMethod:      "*up",
			RoundingDecimals:    4,
			MinCost:             0.1,
			MaxCost:             0.6,
			MaxCostStrategy:     "*free",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationStart: "* * * * 1-5",
			RateWeight:          0,
			RateBlocker:         false,
			RateIntervalStart:   "0s",
			RateRecurrentFee:    0.12,
			RateUnit:            "1m0s",
			RateIncrement:       "1m0s",
			CreatedAt:           time.Time{},
		},
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "",
			ActivationInterval:  "",
			Weight:              0,
			RoundingMethod:      "",
			RoundingDecimals:    0,
			MinCost:             0,
			MaxCost:             0,
			MaxCostStrategy:     "",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationStart: "",
			RateWeight:          0,
			RateBlocker:         false,
			RateIntervalStart:   "1m",
			RateRecurrentFee:    0.06,
			RateUnit:            "1m0s",
			RateIncrement:       "1s",
			CreatedAt:           time.Time{},
		},
	}
	rcv := APItoModelTPRateProfile(tpRprf)
	if !reflect.DeepEqual(rcv, expModels) && !reflect.DeepEqual(rcv, expModelsRev) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expModels), utils.ToJSON(rcv))
	}
}

func TestAsTPRateProfile(t *testing.T) {
	rtMdl := RateProfileMdls{
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "*string:~*req.Subject:1001",
			ActivationInterval:  "",
			Weight:              0,
			RoundingMethod:      "*up",
			RoundingDecimals:    4,
			MinCost:             0.1,
			MaxCost:             0.6,
			MaxCostStrategy:     "*free",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationStart: "* * * * 1-5",
			RateWeight:          0,
			RateBlocker:         false,
			RateIntervalStart:   "1m",
			RateRecurrentFee:    0.06,
			RateUnit:            "1m",
			RateIncrement:       "1s",
			CreatedAt:           time.Time{},
		},
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "",
			ActivationInterval:  "",
			Weight:              0,
			RoundingMethod:      "",
			RoundingDecimals:    0,
			MinCost:             0,
			MaxCost:             0,
			MaxCostStrategy:     "",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationStart: "",
			RateWeight:          0,
			RateBlocker:         false,
			RateIntervalStart:   "0s",
			RateRecurrentFee:    0.12,
			RateUnit:            "1m",
			RateIncrement:       "1m",
			CreatedAt:           time.Time{},
		},
	}

	eRprf := &utils.TPRateProfile{
		TPid:             utils.EmptyString,
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "1m",
						RecurrentFee:  0.06,
						Unit:          "1m",
						Increment:     "1s",
					},
					{
						IntervalStart: "0s",
						RecurrentFee:  0.12,
						Unit:          "1m",
						Increment:     "1m",
					},
				},
			},
		},
	}
	rcv := rtMdl.AsTPRateProfile()
	if len(rcv) != 1 {
		t.Errorf("Expecting: %+v,\nReceived: %+v", 1, len(rcv))
	} else if !reflect.DeepEqual(rcv[0], eRprf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(eRprf), utils.ToJSON(rcv[0]))
	}
}

func TestModelHelperCsvLoadError(t *testing.T) {

	type testStruct struct {
		Id        int64
		Tpid      string
		Tag       string `index:"cat" re:"\w+\s*,\s*"`
		Prefix    string `index:"1" re:"\+?\d+.?\d*"`
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
		Prefix    string `index:"1" re:"\+?\d+.?\d*"`
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
		Tag       string `index:"cat" re:"\w+\s*,\s*"`
		Prefix    string `index:"1" re:"\+?\d+.?\d*"`
		CreatedAt time.Time
	}
	var testStruct1 testStruct
	_, err := CsvDump(testStruct1)
	if err == nil || err.Error() != "invalid testStruct.Tag index cat" {
		t.Errorf("\nExpecting: <invalid testStruct.Tag index cat>,\n  Received: <%+v>", err)
	}
}

func TestModelHelpersAsMapRatesError(t *testing.T) {
	tps := TpRates{{RateUnit: "true"}}
	_, err := tps.AsMapRates()
	if err == nil || err.Error() != "time: invalid duration \"true\"" {
		t.Errorf("Expecting: <time: invalid duration \"true\">,\n  Received: <%+v>", err)
	}
}

func TestModelHelpersAsTPRatesError(t *testing.T) {
	tps := TpRates{{RateUnit: "true"}}
	_, err := tps.AsTPRates()
	if err == nil || err.Error() != "time: invalid duration \"true\"" {
		t.Errorf("Expecting: <time: invalid duration \"true\">,\n  Received: <%+v>", err)
	}

}

func TestAPItoModelTPRoutesCase1(t *testing.T) {
	structTest := &utils.TPRouteProfile{}
	structTest2 := TPRoutes{}
	structTest2 = nil
	result := APItoModelTPRoutes(structTest)
	if !reflect.DeepEqual(structTest2, result) {
		t.Errorf("Expecting: <%+v>,\n  Received: <%+v>", structTest2, result)
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
			&utils.TPRoute{
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

	expected := "[{\"PK\":0,\"Tpid\":\"TP1\",\"Tenant\":\"cgrates.org\",\"ID\":\"RoutePrf\",\"FilterIDs\":\"FLTR_ACNT_dan;FLTR_DST_DE\",\"ActivationInterval\":\"2014-07-29T15:00:00Z;2014-08-29T15:00:00Z\",\"Sorting\":\"*lc\",\"SortingParameters\":\"PARAM1;PARAM2\",\"RouteID\":\"route1\",\"RouteFilterIDs\":\"FLTR_1;FLTR_2\",\"RouteAccountIDs\":\"Acc1;Acc2\",\"RouteRatingplanIDs\":\"RPL_1;RPL_2\",\"RouteRateProfileIDs\":\"\",\"RouteResourceIDs\":\"ResGroup1;ResGroup2\",\"RouteStatIDs\":\"Stat1;Stat2\",\"RouteWeight\":10,\"RouteBlocker\":false,\"RouteParameters\":\"SortingParam1\",\"Weight\":20,\"CreatedAt\":\"0001-01-01T00:00:00Z\"}]"
	result := APItoModelTPRoutes(structTest)
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("Expecting: <%+v>,\n  Received: <%+v>", utils.ToJSON(expected), utils.ToJSON(result))
	}

}

func TestAPItoModelResourceCase1(t *testing.T) {
	var testStruct *utils.TPResourceProfile
	var testStruct2 TpResources
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

	result := APItoModelResource(testStruct)
	expected := "[{\"PK\":0,\"Tpid\":\"LoaderCSVTests\",\"Tenant\":\"cgrates.org\",\"ID\":\"ResGroup1\",\"FilterIDs\":\"\",\"ActivationInterval\":\"2014-07-29T15:00:00Z;2015-07-29T15:00:00Z\",\"UsageTTL\":\"Test_TTL\",\"Limit\":\"2\",\"AllocationMessage\":\"test\",\"Blocker\":false,\"Stored\":false,\"Weight\":10,\"ThresholdIDs\":\"TRes1;TRes2\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"}]"
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestAPItoModelResourceCase3(t *testing.T) {
	testStruct := &utils.TPResourceProfile{
		Tenant:    "cgrates.org",
		TPid:      testTPID,
		ID:        "ResGroup1",
		FilterIDs: []string{"FilterID1", "FilterID2"},
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
	result := APItoModelResource(testStruct)
	expected := "[{\"PK\":0,\"Tpid\":\"LoaderCSVTests\",\"Tenant\":\"cgrates.org\",\"ID\":\"ResGroup1\",\"FilterIDs\":\"FilterID1\",\"ActivationInterval\":\"2014-07-29T15:00:00Z;2015-07-29T15:00:00Z\",\"UsageTTL\":\"Test_TTL\",\"Limit\":\"2\",\"AllocationMessage\":\"test\",\"Blocker\":false,\"Stored\":false,\"Weight\":10,\"ThresholdIDs\":\"TRes1;TRes2\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"},{\"PK\":0,\"Tpid\":\"LoaderCSVTests\",\"Tenant\":\"cgrates.org\",\"ID\":\"ResGroup1\",\"FilterIDs\":\"FilterID2\",\"ActivationInterval\":\"\",\"UsageTTL\":\"\",\"Limit\":\"\",\"AllocationMessage\":\"\",\"Blocker\":false,\"Stored\":false,\"Weight\":0,\"ThresholdIDs\":\"\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"}]"
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
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
			&Route{ID: "ResGroup2"},
		},
	}
	expected := "{\"TPid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"FilterIDs\":[\"FilterID1\",\"FilterID2\"],\"ActivationInterval\":{\"ActivationTime\":\"2020-04-11T21:34:01Z\",\"ExpiryTime\":\"2020-04-12T21:34:01Z\"},\"Sorting\":\"\",\"SortingParameters\":[\"Param1\",\"Param2\"],\"Routes\":[{\"ID\":\"ResGroup2\",\"FilterIDs\":null,\"AccountIDs\":null,\"RatingPlanIDs\":null,\"RateProfileIDs\":null,\"ResourceIDs\":null,\"StatIDs\":null,\"Weight\":0,\"Blocker\":false,\"RouteParameters\":\"\"}],\"Weight\":0}"
	result := RouteProfileToAPI(structTest)
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestRateProfileToAPIWithActInterval(t *testing.T) {
	testProfile := &RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2020, time.April,
				11, 21, 34, 01, 0, time.UTC),
			ExpiryTime: time.Date(2020, time.April,
				12, 21, 34, 01, 0, time.UTC),
		},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates:            map[string]*Rate{},
	}
	expected := "{\"TPid\":\"\",\"Tenant\":\"cgrates.org\",\"ID\":\"RP1\",\"FilterIDs\":[\"*string:~*req.Subject:1001\"],\"ActivationInterval\":{\"ActivationTime\":\"2020-04-11T21:34:01Z\",\"ExpiryTime\":\"2020-04-12T21:34:01Z\"},\"Weight\":0,\"RoundingDecimals\":4,\"RoundingMethod\":\"*up\",\"MinCost\":0.1,\"MaxCost\":0.6,\"MaxCostStrategy\":\"*free\",\"Rates\":{}}"
	result := RateProfileToAPI(testProfile)
	if !reflect.DeepEqual(utils.ToJSON(*result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}

}

func TestModelHelpersAPItoRateProfileErrorTime(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		TPid:      "",
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat",
			ExpiryTime:     "cat2",
		},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates:            map[string]*utils.TPRate{},
	}
	_, err := APItoRateProfile(testStruct, utils.EmptyString)
	if err == nil || err.Error() != "Unsupported time format" {
		t.Errorf("\nExpecting <Unsupported time format>,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoRateProfileError1(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		TPid:             "",
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "cat",
						RecurrentFee:  0.12,
						Unit:          "1m0s",
						Increment:     "1m0s",
					},
				},
			},
		},
	}

	_, err := APItoRateProfile(testStruct, utils.EmptyString)
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\n<time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoRateProfileError2(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		TPid:             "",
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.12,
						Unit:          "cat",
						Increment:     "1m0s",
					},
				},
			},
		},
	}

	_, err := APItoRateProfile(testStruct, utils.EmptyString)
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\n<time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestModelHelpersAPItoRateProfileError3(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		TPid:             "",
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*utils.TPRate{
			"RT_WEEK": {
				ID:             "RT_WEEK",
				Weight:         0,
				ActivationTime: "* * * * 1-5",
				IntervalRates: []*utils.TPIntervalRate{
					{
						IntervalStart: "0s",
						RecurrentFee:  0.12,
						Unit:          "1s",
						Increment:     "cat",
					},
				},
			},
		},
	}

	_, err := APItoRateProfile(testStruct, utils.EmptyString)
	if err == nil || err.Error() != "time: invalid duration \"cat\"" {
		t.Errorf("\n<time: invalid duration \"cat\">,\n Received <%+v>", err)
	}
}

func TestAPItoModelTPRateProfileNil(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		Rates: map[string]*utils.TPRate{},
	}
	result := APItoModelTPRateProfile(testStruct)
	if !reflect.DeepEqual(utils.ToJSON(result), "null") {
		t.Errorf("\nExpecting <null>,\n Received <%+v>", utils.ToJSON(result))
	}
}

func TestAPItoModelTPRateProfileCase2(t *testing.T) {
	testStruct := &utils.TPRateProfile{
		FilterIDs: []string{"test_string1", "test_string2"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2014-08-29T15:00:00Z",
		},
		Rates: map[string]*utils.TPRate{"RT_CHRISTMAS": {
			ID:             "RT_CHRISTMAS",
			FilterIDs:      []string{"test_string1", "test_string2"},
			Weight:         30,
			ActivationTime: "* * 24 12 *",
			IntervalRates: []*utils.TPIntervalRate{
				{
					IntervalStart: "0s",
					RecurrentFee:  0.06,
					Unit:          "1m0s",
					Increment:     "1s",
				},
				{
					IntervalStart: "0s",
					RecurrentFee:  0.06,
					Unit:          "1m0s",
					Increment:     "1s",
				},
			},
		},
		},
	}
	expected := "[{\"PK\":0,\"Tpid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"FilterIDs\":\"test_string1;test_string2\",\"ActivationInterval\":\"2014-07-29T15:00:00Z;2014-08-29T15:00:00Z\",\"Weight\":0,\"RoundingMethod\":\"\",\"RoundingDecimals\":0,\"MinCost\":0,\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"RateID\":\"RT_CHRISTMAS\",\"RateFilterIDs\":\"test_string1;test_string2\",\"RateActivationStart\":\"* * 24 12 *\",\"RateWeight\":30,\"RateBlocker\":false,\"RateIntervalStart\":\"0s\",\"RateFixedFee\":0,\"RateRecurrentFee\":0.06,\"RateUnit\":\"1m0s\",\"RateIncrement\":\"1s\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"},{\"PK\":0,\"Tpid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"FilterIDs\":\"\",\"ActivationInterval\":\"\",\"Weight\":0,\"RoundingMethod\":\"\",\"RoundingDecimals\":0,\"MinCost\":0,\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"RateID\":\"RT_CHRISTMAS\",\"RateFilterIDs\":\"\",\"RateActivationStart\":\"\",\"RateWeight\":0,\"RateBlocker\":false,\"RateIntervalStart\":\"0s\",\"RateFixedFee\":0,\"RateRecurrentFee\":0.06,\"RateUnit\":\"1m0s\",\"RateIncrement\":\"1s\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"}]"
	result := APItoModelTPRateProfile(testStruct)
	if !reflect.DeepEqual(expected, utils.ToJSON(result)) {
		t.Errorf("\nExpecting <[{\"PK\":0,\"Tpid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"FilterIDs\":\"test_string1;test_string2\",\"ActivationInterval\":\"2014-07-29T15:00:00Z;2014-08-29T15:00:00Z\",\"Weight\":0,\"RoundingMethod\":\"\",\"RoundingDecimals\":0,\"MinCost\":0,\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"RateID\":\"RT_CHRISTMAS\",\"RateFilterIDs\":\"test_string1;test_string2\",\"RateActivationStart\":\"* * 24 12 *\",\"RateWeight\":30,\"RateBlocker\":false,\"RateIntervalStart\":\"0s\",\"RateFixedFee\":0,\"RateRecurrentFee\":0.06,\"RateUnit\":\"1m0s\",\"RateIncrement\":\"1s\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"},{\"PK\":0,\"Tpid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"FilterIDs\":\"\",\"ActivationInterval\":\"\",\"Weight\":0,\"RoundingMethod\":\"\",\"RoundingDecimals\":0,\"MinCost\":0,\"MaxCost\":0,\"MaxCostStrategy\":\"\",\"RateID\":\"RT_CHRISTMAS\",\"RateFilterIDs\":\"\",\"RateActivationStart\":\"\",\"RateWeight\":0,\"RateBlocker\":false,\"RateIntervalStart\":\"0s\",\"RateFixedFee\":0,\"RateRecurrentFee\":0.06,\"RateUnit\":\"1m0s\",\"RateIncrement\":\"1s\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"}]>,\n Received <%+v>", utils.ToJSON(result))
	}
}

func TestRateProfileMdlsAsTPRateProfileCase2(t *testing.T) {
	testRPMdls := RateProfileMdls{&RateProfileMdl{
		PK:                  0,
		Tpid:                "",
		Tenant:              "cgrates.org",
		ID:                  "RP1",
		FilterIDs:           "*string:~*req.Subject:1001",
		ActivationInterval:  "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z",
		Weight:              1.2,
		RoundingMethod:      "*up",
		RoundingDecimals:    4,
		MinCost:             0.1,
		MaxCost:             0.6,
		MaxCostStrategy:     "*free",
		RateID:              "RT_WEEK",
		RateFilterIDs:       "TEST_RateFilterIDs",
		RateActivationStart: "* * * * 1-5",
		RateWeight:          1.2,
		RateBlocker:         false,
		RateIntervalStart:   "1m",
		RateRecurrentFee:    0.06,
		RateUnit:            "1m0s",
		RateIncrement:       "1s",
		CreatedAt:           time.Time{},
	},
		&RateProfileMdl{
			PK:                  0,
			Tpid:                "",
			Tenant:              "cgrates.org",
			ID:                  "RP1",
			FilterIDs:           "TEST_RateFilterIDs",
			ActivationInterval:  "2014-07-29T15:00:00Z",
			Weight:              1.2,
			RoundingMethod:      "",
			RoundingDecimals:    0,
			MinCost:             0,
			MaxCost:             0,
			MaxCostStrategy:     "",
			RateID:              "RT_WEEK",
			RateFilterIDs:       "",
			RateActivationStart: "",
			RateWeight:          1.3,
			RateBlocker:         false,
			RateIntervalStart:   "0s",
			RateRecurrentFee:    0.12,
			RateUnit:            "1m0s",
			RateIncrement:       "1m0s",
			CreatedAt:           time.Time{},
		},
	}
	result := testRPMdls.AsTPRateProfile()
	expected := "[{\"TPid\":\"\",\"Tenant\":\"cgrates.org\",\"ID\":\"RP1\",\"FilterIDs\":[\"*string:~*req.Subject:1001\",\"TEST_RateFilterIDs\"],\"ActivationInterval\":{\"ActivationTime\":\"2014-07-29T15:00:00Z\",\"ExpiryTime\":\"\"},\"Weight\":1.2,\"RoundingMethod\":\"*up\",\"RoundingDecimals\":4,\"MinCost\":0.1,\"MaxCost\":0.6,\"MaxCostStrategy\":\"*free\",\"Rates\":{\"RT_WEEK\":{\"ID\":\"RT_WEEK\",\"FilterIDs\":[\"TEST_RateFilterIDs\"],\"ActivationTime\":\"* * * * 1-5\",\"Weight\":1.3,\"Blocker\":false,\"IntervalRates\":[{\"IntervalStart\":\"1m\",\"FixedFee\":0,\"Unit\":\"1m0s\",\"Increment\":\"1s\",\"RecurrentFee\":0.06},{\"IntervalStart\":\"0s\",\"FixedFee\":0,\"Unit\":\"1m0s\",\"Increment\":\"1m0s\",\"RecurrentFee\":0.12}]}}}]"
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <[{\"TPid\":\"\",\"Tenant\":\"cgrates.org\",\"ID\":\"RP1\",\"FilterIDs\":[\"*string:~*req.Subject:1001\",\"TEST_RateFilterIDs\"],\"ActivationInterval\":{\"ActivationTime\":\"2014-07-29T15:00:00Z\",\"ExpiryTime\":\"\"},\"Weight\":1.2,\"RoundingMethod\":\"*up\",\"RoundingDecimals\":4,\"MinCost\":0.1,\"MaxCost\":0.6,\"MaxCostStrategy\":\"*free\",\"Rates\":{\"RT_WEEK\":{\"ID\":\"RT_WEEK\",\"FilterIDs\":[\"TEST_RateFilterIDs\"],\"ActivationTime\":\"* * * * 1-5\",\"Weight\":1.3,\"Blocker\":false,\"IntervalRates\":[{\"IntervalStart\":\"1m\",\"FixedFee\":0,\"Unit\":\"1m0s\",\"Increment\":\"1s\",\"RecurrentFee\":0.06},{\"IntervalStart\":\"0s\",\"FixedFee\":0,\"Unit\":\"1m0s\",\"Increment\":\"1m0s\",\"RecurrentFee\":0.12}]}}}]>,\n Received <%+v>", utils.ToJSON(result))
	}
}

func TestRateProfileMdlsCSVHeader(t *testing.T) {
	testRPMdls := RateProfileMdls{}
	result := testRPMdls.CSVHeader()
	expected := []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.ActivationIntervalString, utils.Weight, utils.ConnectFee, utils.RoundingMethod,
		utils.RoundingDecimals, utils.MinCost, utils.MaxCost, utils.MaxCostStrategy, utils.RateID,
		utils.RateFilterIDs, utils.RateActivationStart, utils.RateWeight, utils.RateBlocker,
		utils.RateIntervalStart, utils.RateFixedFee, utils.RateRecurrentFee, utils.RateUnit, utils.RateIncrement}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, result)
	}
}

func TestDispatcherProfileToAPICase2(t *testing.T) {
	structTest := &DispatcherProfile{
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 15, 14, 35, 0, 0, time.UTC),
		},
		FilterIDs: []string{"field1", "field2"},
		StrategyParams: map[string]interface{}{
			"Field1": "Params1",
		},
		Hosts: []*DispatcherHostProfile{
			{FilterIDs: []string{"fieldA", "fieldB"}},
		},
	}
	expected := "{\"TPid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"Subsystems\":[],\"FilterIDs\":[\"field1\",\"field2\"],\"ActivationInterval\":{\"ActivationTime\":\"2014-07-14T14:35:00Z\",\"ExpiryTime\":\"2014-07-15T14:35:00Z\"},\"Strategy\":\"\",\"StrategyParams\":[\"Params1\"],\"Weight\":0,\"Hosts\":[{\"ID\":\"\",\"FilterIDs\":[\"fieldA\",\"fieldB\"],\"Weight\":0,\"Params\":[],\"Blocker\":false}]}"
	result := DispatcherProfileToAPI(structTest)
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestAPItoDispatcherProfileCase2(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		StrategyParams: []interface{}{"Param1"},
		Hosts: []*utils.TPDispatcherHostProfile{{
			Params: []interface{}{""},
		}},
	}
	expected := "{\"Tenant\":\"\",\"ID\":\"\",\"Subsystems\":[],\"FilterIDs\":[],\"ActivationInterval\":null,\"Strategy\":\"\",\"StrategyParams\":{\"0\":\"Param1\"},\"Weight\":0,\"Hosts\":[{\"ID\":\"\",\"FilterIDs\":[],\"Weight\":0,\"Params\":{},\"Blocker\":false}]}"
	result, _ := APItoDispatcherProfile(structTest, "")
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestAPItoDispatcherProfileError(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "cat1",
			ExpiryTime:     "cat2",
		},
		StrategyParams: []interface{}{"Param1"},
		Hosts: []*utils.TPDispatcherHostProfile{{
			Params: []interface{}{""},
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

func TestAPItoModelTPDispatcherProfileCase2(t *testing.T) {
	structTest := &utils.TPDispatcherProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "2014-07-30T15:00:00Z",
		},
	}
	expected := "[{\"PK\":0,\"Tpid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"Subsystems\":\"\",\"FilterIDs\":\"\",\"ActivationInterval\":\"2014-07-29T15:00:00Z;2014-07-30T15:00:00Z\",\"Strategy\":\"\",\"StrategyParameters\":\"\",\"ConnID\":\"\",\"ConnFilterIDs\":\"\",\"ConnWeight\":0,\"ConnBlocker\":false,\"ConnParameters\":\"\",\"Weight\":0,\"CreatedAt\":\"0001-01-01T00:00:00Z\"}]"
	result := APItoModelTPDispatcherProfile(structTest)
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestModelHelpersParamsToString(t *testing.T) {
	testInterface := []interface{}{"Param1", "Param2"}
	result := paramsToString(testInterface)
	if !reflect.DeepEqual(result, "Param1;Param2") {
		t.Errorf("\nExpecting <Param1;Param2>,\n Received <%+v>", result)
	}
}

func TestModelHelpersAsTPDispatcherProfiles(t *testing.T) {
	structTest := TPDispatcherProfiles{
		&TPDispatcherProfileMdl{
			ActivationInterval: "2014-07-29T15:00:00Z;2014-08-29T15:00:00Z",
			StrategyParameters: "Param1",
		},
	}
	result := structTest.AsTPDispatcherProfiles()
	expected := "[{\"TPid\":\"\",\"Tenant\":\"\",\"ID\":\"\",\"Subsystems\":null,\"FilterIDs\":null,\"ActivationInterval\":{\"ActivationTime\":\"2014-07-29T15:00:00Z\",\"ExpiryTime\":\"2014-08-29T15:00:00Z\"},\"Strategy\":\"\",\"StrategyParams\":[\"Param1\"],\"Weight\":0,\"Hosts\":null}]"
	if !reflect.DeepEqual(utils.ToJSON(result), expected) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", expected, utils.ToJSON(result))
	}
}

func TestTPDispatcherProfilesCSVHeader(t *testing.T) {
	structTest := TPDispatcherProfiles{
		&TPDispatcherProfileMdl{
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
		&TPDispatcherProfileMdl{
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
