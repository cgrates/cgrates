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
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", storageGetter: getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 2 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

func TestApRestoreRedis(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", storageGetter: getter}
	cd.SearchStorageForPrefix()
	if len(cd.ActivationPeriods) != 2 {
		t.Error("Error restoring activation periods: ", cd.ActivationPeriods)
	}
}

func TestApStoreRestore(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Month: time.February,
		MonthDay:  1,
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)

	getter.SetActivationPeriods("storerestore", []*ActivationPeriod{ap})
	aps, err := getter.GetActivationPeriods("storerestore")
	if err != nil || len(aps) != 1 || !reflect.DeepEqual(ap, aps[0]) {
		t.Log(aps)
		t.Errorf("Expected %v was %v ", ap, aps)
	}

}

/**************************** Benchmarks *************************************/
