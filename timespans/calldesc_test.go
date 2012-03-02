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

/*
json
BenchmarkRedisGetCost	    5000	    462787 ns/op
BenchmarkKyotoGetCost	   10000	    203543 ns/op
BenchmarkMongoGetCost	   10000	    320457 ns/op

gob
BenchmarkRedisGetCost	   10000	    258751 ns/op
BenchmarkKyotoGetCost	   50000	     38449 ns/op
BenchmarkMongoGetCost	   10000	    323262 ns/op
*/

func TestKyotoSplitSpans(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}

	cd.SearchStorageForPrefix()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.ActivationPeriods)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestRedisSplitSpans(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", TimeStart: t1, TimeEnd: t2, storageGetter: getter}

	cd.SearchStorageForPrefix()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Log(cd.ActivationPeriods)
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestKyotoGetCost(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 540, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
	cd = &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ = cd.GetCost()
	expected = &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", Cost: 540, ConnectFee: 0}
}

func TestRedisGetCost(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 540, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMongoGetCost(t *testing.T) {
	getter, err := NewMongoStorage("127.0.0.1", "test")
	if err != nil {
		return
	}
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 540, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestFullDestNotFound(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256308200", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 540, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Log(cd.ActivationPeriods)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMultipleActivationPeriods(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 8, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257308200", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", Cost: 330, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Log(result.Timespans)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSpansMultipleActivationPeriods(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 7, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 0, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257308200", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", Cost: 360, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestLessThanAMinute(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 8, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 30, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257308200", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", Cost: 0.5, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestUniquePrice(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 21, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723045326", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723", Cost: 60.35, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestPresentSecodCost(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 23, 50, 21, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.getPresentSecondCost()
	expected := 0.016
	if result != expected {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMinutesCost(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 8, 22, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 22, 51, 50, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", Cost: 0.1, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Log(result.Timespans[0])
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMaxSessionTimeNoUserBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723", storageGetter: getter, Amount: 1000}
	result, err := cd.GetMaxSessionTime()
	if result != 1000 || err != nil {
		t.Errorf("Expected %v was %v", 1000, result)
	}
}

func TestMaxSessionTimeWithUserBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", storageGetter: getter, Amount: 5400}
	result, err := cd.GetMaxSessionTime()
	if result != 1080 || err != nil {
		t.Errorf("Expected %v was %v", 1080, result)
	}
}

func TestMaxSessionTimeNoCredit(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "broker", DestinationPrefix: "0723", storageGetter: getter, Amount: 5400}
	result, err := cd.GetMaxSessionTime()
	if result != 100 || err != nil {
		t.Errorf("Expected %v was %v", 100, result)
	}
}

func TestGetCostWithVolumeDiscount(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	vd1 := &VolumeDiscount{100, 10}
	vd2 := &VolumeDiscount{500, 20}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd1, vd2}}
	rifsBudget := &UserBudget{Id: "rif", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, VolumeDiscountSeconds: 105}
	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723", TimeStart: t1, TimeEnd: t2, storageGetter: getter, userBudget: rifsBudget}
	callCost, err := cd.GetCost()
	if callCost.Cost != 54.0 || err != nil {
		t.Errorf("Expected %v was %v", 54.0, callCost)
	}
}

/*********************************** BENCHMARKS ***************************************/
func BenchmarkRedisGetting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		getter.GetActivationPeriods(cd.GetKey())
	}
}

func BenchmarkRedisRestoring(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.SearchStorageForPrefix()
	}
}

func BenchmarkRedisGetCost(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost()
	}
}

func BenchmarkKyotoGetting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		key := cd.GetKey()
		getter.GetActivationPeriods(key)
	}
}

func BenchmarkKyotoRestoring(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.SearchStorageForPrefix()
	}
}

func BenchmarkSplitting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	cd.SearchStorageForPrefix()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.splitInTimeSpans()
	}
}

func BenchmarkKyotoGetCost(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost()
	}
}

func BenchmarkMongoGetting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		getter.GetActivationPeriods(cd.GetKey())
	}
}

func BenchmarkMongoGetCost(b *testing.B) {
	b.StopTimer()
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, storageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost()
	}
}

func BenchmarkKyotoSingleGetSessionTime(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", storageGetter: getter, Amount: 100}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}

func BenchmarkKyotoMultipleGetSessionTime(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", storageGetter: getter, Amount: 5400}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}

func BenchmarkRedisSingleGetSessionTime(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", storageGetter: getter, Amount: 100}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}

func BenchmarkRedisMultipleGetSessionTime(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", storageGetter: getter, Amount: 5400}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}

func BenchmarkMongoSingleGetSessionTime(b *testing.B) {
	b.StopTimer()
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", storageGetter: getter, Amount: 100}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}

func BenchmarkMongoMultipleGetSessionTime(b *testing.B) {
	b.StopTimer()
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", storageGetter: getter, Amount: 5400}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetMaxSessionTime()
	}
}
