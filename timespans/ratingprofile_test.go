/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can Storagetribute it and/or modify
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

package timespans

import (
	"reflect"
	"testing"
	"time"
)

func TestRpStoreRestore(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{
		Months:    Months{time.February},
		MonthDays: MonthDays{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	rp := &RatingProfile{FallbackKey: "test"}
	rp.AddActivationPeriodIfNotPresent("0723", ap)
	result := rp.store()
	expected := "test>0723=1328106601000000000|;2;1;3,4;14:30:00;15:00:00;0;0;0;0;0"
	if result != expected {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := ActivationPeriod{}
	ap1.restore(result)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}
func TestRpAddAPIfNotPresent(t *testing.T) {
	ap1 := &ActivationPeriod{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 0, time.UTC)}
	ap2 := &ActivationPeriod{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 0, time.UTC)}
	ap3 := &ActivationPeriod{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 1, time.UTC)}
	rp := &RatingProfile{}
	rp.AddActivationPeriodIfNotPresent("test", ap1)
	rp.AddActivationPeriodIfNotPresent("test", ap2)
	if len(rp.DestinationMap["test"]) != 1 {
		t.Error("Wronfully appended activation period ;)", len(rp.DestinationMap["test"]))
	}
	rp.AddActivationPeriodIfNotPresent("test", ap3)
	if len(rp.DestinationMap["test"]) != 2 {
		t.Error("Wronfully not appended activation period ;)", len(rp.DestinationMap["test"]))
	}
}
