package timespans

import (
	"testing"
	"time"
)

func TestKyotoSplitSpans(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	key := cd.GetKey()
	values, _ := getter.Get(key)

	cd.decodeValues([]byte(values))

	intervals := cd.getActiveIntervals()
	timespans := cd.splitInTimeSpans(intervals)
	if len(timespans) != 2 {
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}

func TestRedisSplitSpans(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	key := cd.GetKey()
	values, _ := getter.Get(key)

	cd.decodeValues([]byte(values))

	intervals := cd.getActiveIntervals()
	timespans := cd.splitInTimeSpans(intervals)
	if len(timespans) != 2 {
		t.Error("Wrong number of timespans: ", len(timespans))
	}
}


func TestKyotoGetCost(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost(getter)
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 360, ConnectFee: 0}
	if *result != *expected {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func TestRedisGetCost(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	result, _ := cd.GetCost(getter)
	expected := &CallCost{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", Cost: 360, ConnectFee: 0}
	if *result != *expected {
		t.Errorf("Expected %v was %v", expected, result)
	}
}

func BenchmarkKyotoGetCost(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost(getter)
	}
}

func BenchmarkRedisGetCost(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.GetCost(getter)
	}
}

func BenchmarkDecoding(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256"}
	key := cd.GetKey()
	values, _ := getter.Get(key)
	
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cd.decodeValues([]byte(values))
	}
}