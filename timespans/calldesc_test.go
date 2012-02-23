package timespans

import (
	"testing"
	"time"
	//"log"
)

func TestKyotoSplitSpans(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}

	cd.RestoreFromStorage()
	timespans := cd.splitInTimeSpans()
	if len(timespans) != 2 {
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestRedisSplitSpans(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	cd.RestoreFromStorage()

	timespans := cd.splitInTimeSpans()

	if len(timespans) != 2 {
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestKyotoGetCost(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 540, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
	cd = &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	result, _ = cd.GetCost()
	expected = &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", Cost: 540, ConnectFee: 0}
}

func TestRedisGetCost(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 540, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMongoGetCost(t *testing.T) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256308200", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 540, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestMultipleActivationPeriods(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 8, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257308200", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	result, _ := cd.GetCost()
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257", Cost: 330, ConnectFee: 0}
	if result.Cost != expected.Cost || result.ConnectFee != expected.ConnectFee {
		t.Log(result.Timespans[0].ActivationPeriod)
		t.Log(result.Timespans[1].ActivationPeriod)
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestSpansMultipleActivationPeriods(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 7, 23, 50, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 8, 0, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257308200", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0257308200", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723045326", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0723", StorageGetter: getter}
	result, err := cd.GetMaxSessionTime(1000)
	if result != 1000 || err != nil {
		t.Errorf("Expected %v was %v", 1000, result)
	}
}

func TestMaxSessionTimeWithUserBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	cd := &CallDescriptor{CstmId: "vdf", Subject: "minutosu", DestinationPrefix: "0723", StorageGetter: getter}
	result, err := cd.GetMaxSessionTime(5400)
	if result != 1080 || err != nil {
		t.Errorf("Expected %v was %v", 1080, result)
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.RestoreFromStorage()
	}
}

func BenchmarkRedisGetCost(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.RestoreFromStorage()
	}
}

func BenchmarkSplitting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 2, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 2, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	cd.RestoreFromStorage()
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
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
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2, StorageGetter: getter}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost()
	}
}
