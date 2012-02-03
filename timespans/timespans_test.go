package timespans

import (
	"time"
	"testing"
)

func TestTimespanGetCost(t *testing.T) {
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 17, 55, 0, 0, time.UTC)
	ts1 := TimeSpan{TimeStart: t1, TimeEnd: t2}
	if ts1.GetCost() != 0 {
		t.Error("No interval and still kicking")
	}
	ts1.Interval = &Interval{Price: 1}
	if ts1.GetCost() != 600 {
		t.Error("Expected 10 got ", ts1.GetCost())
	}
	ts1.Interval.BillingUnit = .1
	if ts1.GetCost() != 6000 {
		t.Error("Expected 6000 got ", ts1.GetCost())
	}
}

func TestSetInterval(t *testing.T) {
	i1 := &Interval{Price: 1}
	ts1 := TimeSpan{Interval: i1}
	i2 := &Interval{Price: 2}
	ts1.SetInterval(i2)
	if ts1.Interval != i1 {
		t.Error("Smaller price interval should win")
	}
	i2.Ponder = 1
	ts1.SetInterval(i2)
	if ts1.Interval != i2 {
		t.Error("Bigger ponder interval should win")
	}
}