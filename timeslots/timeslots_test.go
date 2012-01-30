package timeslots

import (
	"testing"
)

func TestSimple(t *testing.T){
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