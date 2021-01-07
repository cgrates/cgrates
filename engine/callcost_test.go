/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
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
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestSingleResultMerge(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 0, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 17, 1, 0, 0, time.UTC)
	cd := &CallDescriptor{Category: "0",
		Tenant: "vdf", Subject: "rif",
		Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.getCost()
	if cc1.Cost != 61 {
		t.Errorf("expected 61 was %v", cc1.Cost)
	}
	t1 = time.Date(2012, time.February, 2, 17, 1, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 17, 2, 0, 0, time.UTC)
	cd = &CallDescriptor{Category: "0", Tenant: "vdf",
		Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.GetCost()
	if cc2.Cost != 61 {
		t.Errorf("expected 60 was %v", cc2.Cost)
	}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 2 || cc1.Timespans[0].GetDuration().Seconds() != 60 || cc1.Timespans[1].GetDuration().Seconds() != 60 {
		for _, ts := range cc1.Timespans {
			t.Logf("TS: %+v", ts)
		}
		t.Error("wrong resulted timespan: ", len(cc1.Timespans), cc1.Timespans[0].GetDuration().Seconds())
	}
	if cc1.Cost != 122 {
		t.Errorf("Exdpected 120 was %v", cc1.Cost)
	}
	d := cc1.UpdateRatedUsage()
	if d != 2*time.Minute || cc1.RatedUsage != 120.0*float64(time.Second) {
		t.Errorf("error updating rating usage: %v, %v", d, cc1.RatedUsage)
	}
}

func TestMultipleResultMerge(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 0, 0, 0, time.UTC)
	cd := &CallDescriptor{Category: "0", Tenant: "vdf",
		Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.getCost()
	if cc1.Cost != 61 {
		//ils.LogFull(cc1)
		t.Errorf("expected 61 was %v", cc1.Cost)
		for _, ts := range cc1.Timespans {
			t.Log(ts.RateInterval)
		}
	}
	t1 = time.Date(2012, time.February, 2, 18, 00, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	cd = &CallDescriptor{Category: "0", Tenant: "vdf", Subject: "rif",
		Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.getCost()
	if cc2.Cost != 30 {
		t.Errorf("expected 30 was %v", cc2.Cost)
		for _, ts := range cc1.Timespans {
			t.Log(ts.RateInterval)
		}
	}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 2 || cc1.Timespans[0].GetDuration().Seconds() != 60 {
		t.Error("wrong resulted timespans: ", len(cc1.Timespans))
	}
	if cc1.Cost != 91 {
		t.Errorf("Exdpected 91 was %v", cc1.Cost)
	}
}

func TestMultipleInputLeftMerge(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	cd := &CallDescriptor{Category: "0", Tenant: "vdf",
		Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.getCost()
	//log.Printf("Timing: %+v", cc1.Timespans[1].RateInterval.Timing)
	//log.Printf("Rating: %+v", cc1.Timespans[1].RateInterval.Rating)
	if cc1.Cost != 91 {
		t.Errorf("expected 91 was %v", cc1.Cost)
	}
	/*t1 = time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 18, 02, 0, 0, time.UTC)
	cd = &CallDescriptor{ToR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.getCost()
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
	cd := &CallDescriptor{Category: "0", Tenant: "vdf", Subject: "rif",
		Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.getCost()
	if cc1.Cost != 61 {
		t.Errorf("expected 61 was %v", cc1.Cost)
	}
	t1 = time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	t2 = time.Date(2012, time.February, 2, 18, 01, 0, 0, time.UTC)
	cd = &CallDescriptor{Category: "0", Tenant: "vdf",
		Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc2, _ := cd.getCost()
	if cc2.Cost != 91 {
		t.Errorf("expected 91 was %v", cc2.Cost)
	}
	cc1.Merge(cc2)
	if len(cc1.Timespans) != 3 || cc1.Timespans[0].GetDuration().Seconds() != 60 {
		t.Error("wrong resulted timespan: ", len(cc1.Timespans), cc1.Timespans[0].GetDuration().Seconds())
	}
	if cc1.Cost != 152 {
		t.Errorf("Exdpected 152 was %v", cc1.Cost)
	}
}

func TestCallCostMergeEmpty(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 58, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 17, 59, 0, 0, time.UTC)
	cd := &CallDescriptor{Category: "0", Tenant: "vdf",
		Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cc1, _ := cd.getCost()
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
			{
				TimeStart: time.Date(2013, 9, 10, 13, 40, 0, 0, time.UTC),
				TimeEnd:   time.Date(2013, 9, 10, 13, 41, 0, 0, time.UTC),
			},
			{
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
		Category:    "data",
		Tenant:      "cgrates.org",
		Subject:     "rif",
		Destination: utils.MetaAny,
		TimeStart:   time.Date(2014, 3, 4, 6, 0, 0, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 4, 6, 1, 5, 0, time.UTC),
		ToR:         utils.VOICE,
	}
	cc, _ := cd.getCost()
	_, err := cc.ToDataCost()
	if err == nil {
		t.Error("Failed to throw error on call to datacost!")
	}
}
