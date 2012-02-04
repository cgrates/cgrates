package main

import (
	"flag"
	"fmt"
	"github.com/simonz05/godis"
	"github.com/fsouza/gokabinet/kc"
	"github.com/rif/cgrates/timespans"
	"time"
)

var (	
	storage = flag.String("storage", "kyoto", "kyoto | redis")
	filename = flag.String("filename", "storage.kch", "kyoto storage file (storage.kch)")
	redisserver = flag.String("server", "tcp:127.0.0.1:6379", "redis server address (tcp:127.0.0.1:6379)")
	redisdb = flag.Int("db", 10, "redis database number (10)")
	redispass = flag.String("pass", "", "redis database password")
)

func main() {
	flag.Parse()

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

	if *storage == "kyoto" {
		db, _ := kc.Open(*filename, kc.WRITE)
		db.Set(key, string(value))
		db.Close()
	} else {		
		db := godis.New(*redisserver, *redisdb, *redispass)
		db.Set(key, value)
		db.Quit()
	}	
	fmt.Println("Done!")
}
