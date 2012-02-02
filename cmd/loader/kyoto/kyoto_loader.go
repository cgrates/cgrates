package main

import (
	"fmt"
	"github.com/fsouza/gokabinet/kc"
	"flag"
	"time"
	"github.com/rif/cgrates/timespans"
)

var (
	filename = flag.String("filename", "storage.kch", "kyoto storage file")
)

func main() {
	flag.Parse()
	db, _ := kc.Open(*filename, kc.WRITE)
	defer db.Close()			

	t1 := time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC)
	cd1 := &timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256"}
	cd1.AddInterval(t1, &timespans.Interval{
		WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}, 
		EndHour:"18:00",
		ConnectFee: 0,
		Price: 0.2,
		BillingUnit: 1.0})
	cd1.AddInterval(t1, &timespans.Interval{
		WeekDays: []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}, 
		StartHour:"18:00",
		ConnectFee: 0,
		Price: 0.1,
		BillingUnit: 1.0})
	cd1.AddInterval(t1, &timespans.Interval{
		WeekDays: []time.Weekday{time.Saturday, time.Sunday}, 		
		ConnectFee: 0,
		Price: 0.1,
		BillingUnit: 1.0})
	key :=  cd1.GetKey()

	value := cd1.EncodeValues()
   	
   	db.Set(key, string(value))
   	fmt.Println("Done!")
}
