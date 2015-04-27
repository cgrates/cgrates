/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestSingleResultMerge(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 0, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 17, 1, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: OUTBOUND, Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.GetCost()
	if cc1.Cost != 61 {
		t.Errorf("expected 61 was %v", cc1.Cost)
	}
	/*t1 = time.Date(2012, time.February, 2, 17, 1, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 17, 2, 0, 0, time.UTC)
	cd = &CallDescriptor{Direction: OUTBOUND, TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.GetCost()
	if cc2.Cost != 60 {
		t.Errorf("expected 60 was %v", cc2.Cost)
	}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 1 || cc1.Timespans[0].GetDuration().Seconds() != 120 {
		t.Error("wrong resulted timespan: ", len(cc1.Timespans))
	}
	if cc1.Cost != 120 {
		t.Errorf("Exdpected 120 was %v", cc1.Cost)
	}*/
}

func TestMultipleResultMerge(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 0, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: OUTBOUND, Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.GetCost()
	if cc1.Cost != 61 {
		t.Errorf("expected 61 was %v", cc1.Cost)
		for _, ts := range cc1.Timespans {
			t.Log(ts.RateInterval)
		}
	}
	t1 = time.Date(2012, time.February, 2, 18, 00, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	cd = &CallDescriptor{Direction: OUTBOUND, Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.GetCost()
	if cc2.Cost != 30 {
		t.Errorf("expected 30 was %v", cc2.Cost)
		for _, ts := range cc1.Timespans {
			t.Log(ts.RateInterval)
		}
	}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 2 || cc1.Timespans[0].GetDuration().Seconds() != 60 {
		t.Error("wrong resulted timespan: ", len(cc1.Timespans))
	}
	if cc1.Cost != 91 {
		t.Errorf("Exdpected 91 was %v", cc1.Cost)
	}
}

func TestMultipleInputLeftMerge(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: OUTBOUND, Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.GetCost()
	//log.Printf("Timing: %+v", cc1.Timespans[1].RateInterval.Timing)
	//log.Printf("Rating: %+v", cc1.Timespans[1].RateInterval.Rating)
	if cc1.Cost != 91 {
		t.Errorf("expected 91 was %v", cc1.Cost)
	}
	/*t1 = time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 18, 02, 0, 0, time.UTC)
	cd = &CallDescriptor{Direction: OUTBOUND, TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.GetCost()
	if cc2.Cost != 30 {
		t.Errorf("expected 30 was %v", cc2.Cost)
	}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 2 || cc1.Timespans[1].GetDuration().Seconds() != 120 {
		t.Error("wrong resulted timespan: ", len(cc1.Timespans))
	}
	if cc1.Cost != 120 {
		t.Errorf("Exdpected 120 was %v", cc1.Cost)
	}*/
}

func TestMultipleInputRightMerge(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 58, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: OUTBOUND, Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.GetCost()
	if cc1.Cost != 61 {
		t.Errorf("expected 61 was %v", cc1.Cost)
	}
	t1 = time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	cd = &CallDescriptor{Direction: OUTBOUND, Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.GetCost()
	if cc2.Cost != 91 {
		t.Errorf("expected 91 was %v", cc2.Cost)
	}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 2 || cc1.Timespans[0].GetDuration().Seconds() != 120 {
		t.Error("wrong resulted timespan: ", len(cc1.Timespans))
	}
	if cc1.Cost != 152 {
		t.Errorf("Exdpected 152 was %v", cc1.Cost)
	}
}

func TestCallCostMergeEmpty(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 58, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: OUTBOUND, Category: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.GetCost()
	cc2 := &CallCost{}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 1 {
		t.Error("Error mergin empty call cost: ", len(cc1.Timespans))
	}
}

func TestCallCostGetDurationZero(t *testing.T) {
	cc := &CallCost{}
	if cc.GetDuration().Seconds() != 0 {
		t.Error("Wrong call cost duration for zero timespans: ", cc.GetDuration())
	}
}

func TestCallCostGetDuration(t *testing.T) {
	cc := &CallCost{
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart: time.Date(2013, 9, 10, 13, 40, 0, 0, time.UTC),
				TimeEnd:   time.Date(2013, 9, 10, 13, 41, 0, 0, time.UTC),
			},
			&TimeSpan{
				TimeStart: time.Date(2013, 9, 10, 13, 41, 0, 0, time.UTC),
				TimeEnd:   time.Date(2013, 9, 10, 13, 41, 30, 0, time.UTC),
			},
		},
	}
	if cc.GetDuration().Seconds() != 90 {
		t.Error("Wrong call cost duration: ", cc.GetDuration())
	}
}

func TestCallCostToDataCostError(t *testing.T) {
	cd := &CallDescriptor{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "cgrates.org",
		Subject:     "rif",
		Destination: utils.ANY,
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 1, 5, 0, time.UTC),
		TOR:         utils.VOICE,
	}
	cc, _ := cd.GetCost()
	_, err := cc.ToDataCost()
	if err == nil {
		t.Error("Failed to throw error on call to datacost!")
	}
}

/*func TestCallCostToDataCost(t *testing.T) {
	cd := &CallDescriptor{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "cgrates.org",
		Subject:     "rif",
		Destination: utils.ANY,
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 1, 5, 0, time.UTC),
		TOR:         DATA,
	}
	cc, _ := cd.GetCost()
	dc, err := cc.ToDataCost()
	if err != nil {
		t.Error("Error convertiong to data cost: ", err)
	}
	expected := &DataCost{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "cgrates.org",
		Subject:     "rif",
		Account:     "",
		Destination: "*any",
		TOR:         "*data",
		Cost:        65,
		DataSpans: []*DataSpan{
			&DataSpan{
				DataStart: 0,
				DataEnd:   60,
				Cost:      60,
				RateInterval: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: "00:00:00",
						EndTime:   "",
					},
					Rating: &RIRate{
						ConnectFee:       0,
						RoundingMethod:   "*middle",
						RoundingDecimals: 4,
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 60000000000,
								RateUnit:      1000000000},
							&Rate{GroupIntervalStart: 60000000000,
								Value:         1,
								RateIncrement: 1000000000,
								RateUnit:      1000000000},
						},
					},
					Weight: 10},
				DataIndex:      60,
				Increments:     []*DataIncrement{},
				MatchedSubject: "",
				MatchedPrefix:  "",
				MatchedDestId:  ""},
			&DataSpan{
				DataStart: 60,
				DataEnd:   65,
				Cost:      5,
				RateInterval: &RateInterval{
					Timing: &RITiming{
						Years:     utils.Years{},
						Months:    utils.Months{},
						MonthDays: utils.MonthDays{},
						WeekDays:  utils.WeekDays{},
						StartTime: "00:00:00",
						EndTime:   "",
					},
					Rating: &RIRate{
						ConnectFee:       0,
						RoundingMethod:   "*middle",
						RoundingDecimals: 4,
						Rates: RateGroups{
							&Rate{GroupIntervalStart: 0,
								Value:         1,
								RateIncrement: 60000000000,
								RateUnit:      1000000000},
							&Rate{GroupIntervalStart: 60000000000,
								Value:         1,
								RateIncrement: 1000000000,
								RateUnit:      1000000000},
						},
					},
					Weight: 10},
				DataIndex:      65,
				Increments:     []*DataIncrement{},
				MatchedSubject: "*out:cgrates.org:data:rif",
				MatchedPrefix:  "*any",
				MatchedDestId:  "*any"},
		},
	}

}*/
