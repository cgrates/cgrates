package timespans

import (
	"testing"
	"time"
	//"log"
)

func TestApStoreRestore(t *testing.T) {
	d := time.Date(2012, time.February, 1, 14, 30, 1, 0, time.UTC)
	i := &Interval{Month: time.February,
		MonthDay:  1,
		WeekDays:  []time.Weekday{time.Wednesday, time.Thursday},
		StartTime: "14:30:00",
		EndTime:   "15:00:00"}
	ap := &ActivationPeriod{ActivationTime: d}
	ap.AddInterval(i)
	storage, _ := NewKyotoStorage("test.kch")
	result := storage.store(ap)
	expected := "1328106601000000000;2|1|3,4|14:30:00|15:00:00|0|0|0|0;"
	if result != expected {
		t.Errorf("Expected %q was %q", expected, result)
	}
	ap1 := storage.restore(result)
	if ap1.ActivationTime != ap.ActivationTime {
		t.Errorf("Expected %v was %v", ap.ActivationTime, ap1.ActivationTime)
	}
	i1 := ap1.Intervals[0]
	if i1.Month != i.Month {
		t.Errorf("Expected %q was %q", i.Month, i1.Month)
	}
	if i1.MonthDay != i.MonthDay {
		t.Errorf("Expected %q was %q", i.MonthDay, i1.MonthDay)
	}
	for j, wd := range i1.WeekDays {
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

func BenchmarkActivationPeriodRestore(b *testing.B) {
	storage, _ := NewKyotoStorage("test.kch")
	for i := 0; i < b.N; i++ {
		storage.restore("1328106601;2|1|3,4|14:30:00|15:00:00|0|0|0|0;")
	}
}
