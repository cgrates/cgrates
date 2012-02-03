package main

import (
	"fmt"
	"github.com/simonz05/godis"
	"github.com/rif/cgrates/timespans"
	"time"
)

func main() {
	r := godis.New("", 10, "")
	t1 := time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)
	cd1 := &timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256"}
	ap1 := &timespans.ActivationPeriod{ActivationTime: t1}
	ap1.AddInterval(&timespans.Interval{
		WeekDays:    []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		EndTime:     "18:00:00",
		ConnectFee:  0,
		Price:       0.2,
		BillingUnit: 1.0})
	ap1.AddInterval(&timespans.Interval{
		WeekDays:    []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		StartTime:   "18:00:00",
		ConnectFee:  0,
		Price:       0.1,
		BillingUnit: 1.0})
	ap1.AddInterval(&timespans.Interval{
		WeekDays:    []time.Weekday{time.Saturday, time.Sunday},
		ConnectFee:  0,
		Price:       0.1,
		BillingUnit: 1.0})
	cd1.AddActivationPeriod(ap1)
	key := cd1.GetKey()

	value := cd1.EncodeValues()
	r.Set(key, string(value))
	fmt.Println("Done!")
}
