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

func TestApStoreRestore(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Month: time.February,
		MonthDay:  1,
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}	
	ap := ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	result := ap.Store()
	expected := "1328106601,1328106601000000000;2|1||14:30:00|15:00:00|0|0|0|0;"
	if result != expected {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := ActivationPeriod{}
	ap1.Restore(result)
	if ap1.ActivationTime != ap.ActivationTime {
		//t.Errorf("Expected %v was %v", ap.ActivationTime, ap1.ActivationTime)
	}
	i1 := ap1.Intervals[0]
	if i1.Month != i.Month {
		t.Errorf("Expected %q was %q", i.Month, i1.Month)	
	}
	if i1.MonthDay != i.MonthDay {
		t.Errorf("Expected %q was %q", i.MonthDay, i1.MonthDay)	
	}
	for j,wd := range i1.WeekDays {
		if wd != i1.WeekDays[j] {
			t.Errorf("Expected %q was %q", i.StartTime, i1.StartTime)	
		}	
	}
	if i1.StartTime != i.StartTime {
		t.Errorf("Expected %q was %q", i.StartTime, i1.StartTime)	
	}
	if i1.EndTime != i.EndTime {
		t.Errorf("Expected %q was %q", i.EndTime, i1.EndTime)	
	}
	if i1.Ponder != i.Ponder {
		t.Errorf("Expected %q was %q", i.Ponder, i1.Ponder)	
	}
	if i1.ConnectFee != i.ConnectFee {
		t.Errorf("Expected %q was %q", i.ConnectFee, i1.ConnectFee)	
	}
	if i1.Price != i.Price {
		t.Errorf("Expected %q was %q", i.Price, i1.Price)	
	}
	if i1.BillingUnit != i.BillingUnit {
		t.Errorf("Expected %q was %q", i.BillingUnit, i1.BillingUnit)	
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

func BenchmarkSplitting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}	
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		intervals := cd.getActiveIntervals()
		cd.splitInTimeSpans(intervals)
	}
}

func BenchmarkKyotoGetting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}	
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		getter.Get(cd.GetKey())
	}
}

func BenchmarkRedisGetting(b *testing.B) {
	b.StopTimer()
	getter, _ := NewRedisStorage("", 10)
	defer getter.Close()

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}	
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		getter.Get(cd.GetKey())
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

func BenchmarkRestore(b *testing.B) {	
	ap1 := ActivationPeriod{}
	for i := 0; i < b.N; i++ {
		ap1.Restore("1328106601,1328106601000000000;2|1|3,4|14:30:00|15:00:00|0|0|0|0;")
	}
}
