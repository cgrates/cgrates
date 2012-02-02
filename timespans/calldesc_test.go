package timespans

import (
		"testing"
		"time"
		"log"
)

func TestSplitSpans(t *testing.T){
	getter, _ := NewKyotoStorage("test.kch")
	
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	key := cd.GetKey()
	values, _ := getter.Get(key)
	
	cd.decodeValues([]byte(values))
	
	intervals := cd.getActiveIntervals()
	timespans := cd.splitInTimeSpans(intervals)
	for _, ts := range timespans{
		log.Print(ts)
	}
}

func TestGetCost(t *testing.T){
	getter, _ := NewKyotoStorage("test.kch")
	
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	cc := cd.GetCost(getter)
	log.Print(ts)
}

func BenchmarkGetCost(b *testing.B) {	
	b.StopTime()
	getter, _ := NewKyotoStorage("test.kch")
	
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := &CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	b.StartTimeTime()
    for i := 0; i < b.N; i++ {		
    	cd.GetCost(getter)
    }
}
