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