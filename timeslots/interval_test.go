package timeslots

import (
	"time"
	"testing"	
)

func TestMonth(t *testing.T){
	i := &Interval{Month: time.February}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}	
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}	
}

func TestMonthDay(t *testing.T){
	i := &Interval{MonthDay: 10}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 11, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}	
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}	
}

func TestMonthAndMonthDay(t *testing.T){
	i := &Interval{Month: time.February, MonthDay: 10}
	d := time.Date(2012, time.February, 10, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 11, 23, 0, 0, 0, time.UTC)
	d2 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}	
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}	
	if i.Contains(d2) {
		t.Errorf("Date %v shoud not be in interval %v", d2, i)
	}	
}

func TestWeekDays(t *testing.T){
	i := &Interval{WeekDays: []time.Weekday{time.Wednesday}}
	i2 := &Interval{WeekDays: []time.Weekday{time.Wednesday, time.Thursday}}
	d := time.Date(2012, time.February, 1, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 2, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}	
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}	
	if !i2.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i2)
	}	
	if !i2.Contains(d1) {
		t.Errorf("Date %v shoud be in interval %v", d1, i2)
	}	
}

func TestMonthAndMonthDayAndWeekDays(t *testing.T){
	i := &Interval{Month: time.February, MonthDay: 1, WeekDays: []time.Weekday{time.Wednesday}}
	i2 := &Interval{Month: time.February, MonthDay: 2, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}}
	d := time.Date(2012, time.February, 1, 23, 0, 0, 0, time.UTC)
	d1 := time.Date(2012, time.February, 2, 23, 0, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}	
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}	
	if i2.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i2)
	}	
	if !i2.Contains(d1) {
		t.Errorf("Date %v shoud be in interval %v", d1, i2)
	}	
}

func TestHours(t *testing.T){
	i := &Interval{StartHour: "14:30", EndHour: "15:00"}
	d := time.Date(2012, time.February, 10, 14, 30, 0, 0, time.UTC)
	d1 := time.Date(2012, time.January, 10, 14, 29, 0, 0, time.UTC)
	d2 := time.Date(2012, time.January, 10, 14, 59, 0, 0, time.UTC)
	d3 := time.Date(2012, time.January, 10, 15, 01, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}	
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}	
	if !i.Contains(d2) {
		t.Errorf("Date %v shoud be in interval %v", d2, i)
	}	
	if i.Contains(d3) {
		t.Errorf("Date %v shoud not be in interval %v", d3, i)
	}	
}

func TestEverything(t *testing.T){
	i := &Interval{Month: time.February, MonthDay: 1, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}, StartHour: "14:30", EndHour: "15:00"}
	d := time.Date(2012, time.February, 1, 14, 30, 0, 0, time.UTC)
	d1 := time.Date(2012, time.January, 1, 14, 29, 0, 0, time.UTC)
	if !i.Contains(d) {
		t.Errorf("Date %v shoud be in interval %v", d, i)
	}	
	if i.Contains(d1) {
		t.Errorf("Date %v shoud not be in interval %v", d1, i)
	}	
}

func TestRightMargin(t *testing.T){
	i := &Interval{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}
	t1 := time.Date(2012, time.February, 3, 23, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 4, 0, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	result := i.Split(ts)
	ts1, ts2 := result[0], result[1]
	if ts1.TimeStart != t1 || ts1.TimeEnd != time.Date(2012, time.February, 3, 23, 59, 59, 0, time.UTC) {
		t.Error("Incorrect first half", ts1)
	}
	if ts2.TimeStart != time.Date(2012, time.February, 3, 23, 59, 59, 0, time.UTC) || ts2.TimeEnd != t2 {
		t.Error("Incorrect second half", ts2)
	}
	if ts1.Interval != i {
		t.Error("Interval not attached correctly")
	}
}

func TestLeftMargin(t *testing.T){
	i := &Interval{WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}}
	t1 := time.Date(2012, time.February, 5, 23, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 6, 0, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	result := i.Split(ts)
	ts1, ts2 := result[0], result[1]
	if ts1.TimeStart != t1 || ts1.TimeEnd != time.Date(2012, time.February, 6, 0, 0, 0, 0, time.UTC) {
		t.Error("Incorrect first half", ts1)
	}
	if ts2.TimeStart != time.Date(2012, time.February, 6, 0, 0, 0, 0, time.UTC) || ts2.TimeEnd != t2 {
		t.Error("Incorrect second half", ts2)
	}
	if ts2.Interval != i {
		t.Error("Interval not attached correctly")
	}
}

func TestEnclosingMargin(t *testing.T){
	i := &Interval{WeekDays: []time.Weekday{time.Sunday}}
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 18, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	result := i.Split(ts)
	ts1 := result[0]
	if ts1.TimeStart != t1 || ts1.TimeEnd != t2 {
		t.Error("Incorrect enclosing", ts1)
	}	
	if ts1.Interval != i {
		t.Error("Interval not attached correctly")
	}
}

func TestOutsideMargin(t *testing.T){
	i := &Interval{WeekDays: []time.Weekday{time.Monday}}
	t1 := time.Date(2012, time.February, 5, 17, 45, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 5, 18, 10, 0, 0, time.UTC)
	ts := &TimeSpan{TimeStart: t1, TimeEnd: t2}
	result := i.Split(ts)	
	if result != nil {
		t.Error("Interval not split correctly")
	}
}

func BenchmarkIntervalFull(b *testing.B) {
	i := &Interval{Month: time.February, MonthDay: 1, WeekDays: []time.Weekday{time.Wednesday, time.Thursday}, StartHour: "14:30", EndHour: "15:00"}
	d := time.Date(2012, time.February, 1, 14, 30, 0, 0, time.UTC)
    for x := 0; x < b.N; x++ {
    	i.Contains(d)
    }
}