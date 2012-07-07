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
	"testing"
	"time"
	//"log"
)

func init() {
	storageGetter, _ = NewRedisStorage("127.0.0.1:6379", 10)
	SetStorageGetter(storageGetter)
	populateDB()
}

func populateDB() {
	minu := &UserBalance{
		Id:   "OUT:vdf:minu",
		Type: UB_TYPE_PREPAID,
		BalanceMap: map[string]float64{
			CREDIT: 0,
		},
		MinuteBuckets: []*MinuteBucket{
			&MinuteBucket{Seconds: 200, DestinationId: "NAT", Weight: 10},
			&MinuteBucket{Seconds: 100, DestinationId: "RET", Weight: 20},
		},
	}
	broker := &UserBalance{
		Id:   "OUT:vdf:broker",
		Type: UB_TYPE_PREPAID,
		MinuteBuckets: []*MinuteBucket{
			&MinuteBucket{Seconds: 20, DestinationId: "NAT", Weight: 10, Price: 1},
			&MinuteBucket{Seconds: 100, DestinationId: "RET", Weight: 20},
		},
	}
	storageGetter.SetUserBalance(broker)
	storageGetter.SetUserBalance(minu)
}

func TestSplitSpans(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}

	cd.SearchStorageForPrefix()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.ActivationPeriods)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestRedisSplitSpans(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0257", TimeStart: t1, TimeEnd: t2}

	cd.SearchStorageForPrefix()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.ActivationPeriods)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestGetCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0256", Cost: 2700, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestFullDestNotFound(t *testing.T) {
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0256", Cost: 2700, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Log(cd.ActivationPeriods)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMultipleActivationPeriods(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0257", Cost: 2700, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Log(result.Timespans)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSpansMultipleActivationPeriods(t *testing.T) {
	t1 := time.Date(2012, time.February, 7, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 0, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0257", Cost: 1200, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestLessThanAMinute(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 30, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0257308200", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0257", Cost: 15, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestUniquePrice(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 21, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0723045326", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "rif", Destination: "0723", Cost: 1810.5, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMinutesCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 22, 51, 50, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0723", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost()
	expected := &CallCost{Tenant: "vdf", Subject: "minutosu", Destination: "0723", Cost: 55, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Log(result.Timespans[0])
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeNoUserBalance(t *testing.T) {
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0723", Amount: 1000}
	result, err := cd.GetMaxSessionTime()
	if result != 1000 || err != nil {
		t.Errorf("Expected %v was %v", 1000, result)
	}
}

func TestMaxSessionTimeWithUserBalance(t *testing.T) {
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "minu", Destination: "0723", Amount: 1000}
	result, err := cd.GetMaxSessionTime()
	expected := 300.0
	if result != expected || err != nil {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeNoCredit(t *testing.T) {
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "broker", Destination: "0723", Amount: 5400}
	result, err := cd.GetMaxSessionTime()
	if result != 100 || err != nil {
		t.Errorf("Expected %v was %v", 100, result)
	}
}

func TestApAddAPIfNotPresent(t *testing.T) {
	ap1 := &ActivationPeriod{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 0, time.UTC)}
	ap2 := &ActivationPeriod{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 0, time.UTC)}
	ap3 := &ActivationPeriod{ActivationTime: time.Date(2012, time.July, 2, 14, 24, 30, 1, time.UTC)}
	cd := &CallDescriptor{}
	cd.AddActivationPeriodIfNotPresent(ap1)
	cd.AddActivationPeriodIfNotPresent(ap2)
	if len(cd.ActivationPeriods) != 1 {
		t.Error("Wronfully appended activation period ;)", len(cd.ActivationPeriods))
	}
	cd.AddActivationPeriodIfNotPresent(ap3)
	if len(cd.ActivationPeriods) != 2 {
		t.Error("Wronfully not appended activation period ;)", len(cd.ActivationPeriods))
	}
}

/*********************************** BENCHMARKS ***************************************/
func BenchmarkRedisGetting(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		storageGetter.GetActivationPeriodsOrFallback(cd.GetKey())
	}
}

func BenchmarkRedisRestoring(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.SearchStorageForPrefix()
	}
}

func BenchmarkRedisGetCost(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost()
	}
}

func BenchmarkSplitting(b *testing.B) {
	b.StopTimer()
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	cd.SearchStorageForPrefix()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.splitInTimeSpans()
	}
}

func BenchmarkRedisSingleGetSessionTime(b *testing.B) {
	b.StopTimer()
	cd := &CallDescriptor{Tenant: "vdf", Subject: "minutosu", Destination: "0723", Amount: 100}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}

func BenchmarkRedisMultipleGetSessionTime(b *testing.B) {
	b.StopTimer()
	cd := &CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "minutosu", Destination: "0723", Amount: 5400}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}
