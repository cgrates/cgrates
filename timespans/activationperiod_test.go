/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	"reflect"
	"testing"
	"time"
	//"log"
)

func TestApRestoreKyoto(t *testing.T) {
	getter, _ := NewKyotoStorage("../data/test.kch")
	defer getter.Close()

	cd := &CallDescriptor{Tenant: "vdf",
		Subject:           "rif",
		DestinationPrefix: "0257",
		storageGetter:     getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 2 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

func TestApRestoreRedis(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	cd := &CallDescriptor{Tenant: "vdf",
		Subject:           "rif",
		DestinationPrefix: "0257",
		storageGetter:     getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 2 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

func TestApStoreRestore(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result := ap.store()
	expected := "1328106601000000000;2|1|3,4|14:30:00|15:00:00|0|0|0|0;"
	if result != expected {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := ActivationPeriod{}
	ap1.restore(result)
	if reflect.DeepEqual(ap, ap1) {
		t.Errorf("Expected %v was %v", ap, ap1)
	}
}

func TestFallbackDirect(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	cd := &CallDescriptor{Tenant: "vdf", Subject: "rif", DestinationPrefix: "0745", storageGetter: getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

func TestFallbackWithBackTrace(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	cd := &CallDescriptor{Tenant: "vdf", Subject: "rif", DestinationPrefix: "0745121", storageGetter: getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

func TestFallbackDefault(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	cd := &CallDescriptor{Tenant: "vdf", Subject: "rif", DestinationPrefix: "00000", storageGetter: getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

func TestFallbackNoInfiniteLoop(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	cd := &CallDescriptor{Tenant: "vdf", Subject: "rif", DestinationPrefix: "0721", storageGetter: getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 0 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

/**************************** Benchmarks *************************************/

func BenchmarkActivationPeriodRestore(b *testing.B) {
	ap := ActivationPeriod{}
	for i := 0; i < b.N; i++ {
		ap.restore("1328106601;2|1|3,4|14:30:00|15:00:00|0|0|0|0;")
	}
}

func BenchmarkActivationPeriodStoreRestore(b *testing.B) {
	b.StopTimer()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Months: []time.Month{time.February},
		MonthDays: []int{1},
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)

	ap1 := ActivationPeriod{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		result := ap.store()
		ap1.restore(result)
	}
}
