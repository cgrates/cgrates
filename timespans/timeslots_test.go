package timeslots

import (
	"time"
	"testing"
)

func TestStorageEncoding(t *testing.T){
	i1 := &Interval{Month: time.December, MonthDay: 1, StartHour: "09:00"}
	i2 := &Interval{WeekDays: []time.Weekday{time.Sunday}}
	i3 := &Interval{Month: time.February, MonthDay: 1, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}, StartHour: "14:30", EndHour: "15:00"}
	c := &Customer{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256"}
	ap := &ActivationPeriod{ActivationTime: time.Now()}
	ap.AddInterval(i1, i2, i3)
	c.addActivationPeriod(ap)
	received := c.encodeValue()	
	c.decodeValue(received)
	f1 := c.ActivationPeriods[0].Intervals[0]	
	if f1.Month != i1.Month || f1.MonthDay != i1.MonthDay || f1.StartHour != i1.StartHour {
		t.Errorf("Decode values are not the same: %v vs %v", f1, i1)
	}
	f2 := c.ActivationPeriods[0].Intervals[1]
	for i,v := range f2.WeekDays {
		if v != i2.WeekDays[i] {
			t.Errorf("Decode values are not the same: %v vs %v", f2, i2)
		}
	}
}

func BenchmarkDecoding(b *testing.B) {
	b.StopTimer()
	i1 := &Interval{Month: time.December, MonthDay: 1, StartHour: "09:00"}
	i2 := &Interval{WeekDays: []time.Weekday{time.Sunday}}
	i3 := &Interval{Month: time.February, MonthDay: 1, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}, StartHour: "14:30", EndHour: "15:00"}
	c := &Customer{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256"}
	ap := &ActivationPeriod{ActivationTime: time.Now()}
	ap.AddInterval(i1, i2, i3)
	c.addActivationPeriod(ap)
	received := c.encodeValue()	
	b.StartTimer()
    for i := 0; i < b.N; i++ {
		c.decodeValue(received)
    }
}