package timeslots

import (
	"time"
	"testing"	
)

func setUp() {
	c1 := &Customer{CstmId:"rif",DestinationPrefix: "40256"}
	t1 := time.Date(2012, time.January, 10, 23, 0, 0, 0, time.UTC) 
	ap1 := &ActivationPeriod{ActivationTime: t1}
	c1.AddActivationPeriod(ap1)
	d1,_ := time.ParseDuration("1m")
	d2,_ := time.ParseDuration("2m")
	r1 := &RatingProfile{StartTime: d1, ConnectFee: 1.1, Price: 0.1, BillingUnit: SECOND}
	r2 := &RatingProfile{StartTime: d2, ConnectFee: 2.2, Price: 0.2, BillingUnit: SECOND}
	ap1.AddRatingProfile(r1, r2)
}

func TestSimple(t *testing.T){
	setUp()
	cc, err := GetCost(nil, nil)
	if err != nil {
		t.Error("Got error on getting cost")
	}
	expected:= &CallCost{TOR: 1, CstmId:"",Subject:"",Prefix:"", Cost:1, ConnectFee:1}
	if *cc != *expected {
		t.Errorf("Expected %v got %v", expected, cc)
	}
}

func BenchmarkSimple(b *testing.B) {
    for i := 0; i < b.N; i++ {
		GetCost(nil, nil)
    }
}