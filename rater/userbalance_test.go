/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package rater

import (
	//"log"
	"reflect"
	"testing"
)

var (
	NAT = &Destination{Id: "NAT", Prefixes: []string{"0257", "0256", "0723"}}
	RET = &Destination{Id: "RET", Prefixes: []string{"0723", "0724"}}
)

func init() {
	populateTestActionsForTriggers()
}

func populateTestActionsForTriggers() {
	ats := []*Action{
		&Action{ActionType: "TOPUP", BalanceId: CREDIT, Direction: OUTBOUND, Units: 10},
		&Action{ActionType: "TOPUP", BalanceId: MINUTES, Direction: OUTBOUND, MinuteBucket: &MinuteBucket{Weight: 20, Price: 1, Seconds: 10, DestinationId: "NAT"}},
	}
	storageGetter.SetActions("TEST_ACTIONS", ats)
	ats1 := []*Action{
		&Action{ActionType: "TOPUP", BalanceId: CREDIT, Direction: OUTBOUND, Units: 10, Weight: 20},
		&Action{ActionType: "RESET_PREPAID", Weight: 10},
	}
	storageGetter.SetActions("TEST_ACTIONS_ORDER", ats1)
}

func TestUserBalanceStoreRestore(t *testing.T) {
	uc := &UnitsCounter{
		Direction:     OUTBOUND,
		BalanceId:     SMS,
		Units:         100,
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, Percent: 0, DestinationId: "RET"}},
	}
	at := &ActionTrigger{
		Id:             "some_uuid",
		BalanceId:      CREDIT,
		Direction:      OUTBOUND,
		ThresholdValue: 100.0,
		DestinationId:  "NAT",
		Weight:         10.0,
		ActionsId:      "Commando",
	}
	ub := &UserBalance{
		Id:             "rif",
		Type:           UB_TYPE_POSTPAID,
		BalanceMap:     map[string]float64{SMS + OUTBOUND: 14, TRAFFIC + OUTBOUND: 1024},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, Percent: 0, DestinationId: "RET"}},
		UnitCounters:   []*UnitsCounter{uc, uc},
		ActionTriggers: ActionTriggerPriotityList{at, at, at},
	}
	r := ub.store()
	if string(r) != "rif|postpaid|SMSOUT:14#INTERNETOUT:1024|0;20;1;0;NAT#0;10;10;0;RET|OUT/SMS/100/0;20;1;0;NAT,0;10;10;0;RET#OUT/SMS/100/0;20;1;0;NAT,0;10;10;0;RET|some_uuid;MONETARY;OUT;NAT;Commando;100;10;false#some_uuid;MONETARY;OUT;NAT;Commando;100;10;false#some_uuid;MONETARY;OUT;NAT;Commando;100;10;false" &&
		string(r) != "rif|postpaid|INTERNETOUT:1024#SMSOUT:14|0;20;1;0;NAT#0;10;10;0;RET|OUT/SMS/100/0;20;1;0;NAT,0;10;10;0;RET#OUT/SMS/100/0;20;1;0;NAT,0;10;10;0;RET|some_uuid;MONETARY;OUT;NAT;Commando;100;10;false#some_uuid;MONETARY;OUT;NAT;Commando;100;10;false#some_uuid;MONETARY;OUT;NAT;Commando;100;10;false" {
		t.Errorf("Error serializing action timing: %v", string(r))
	}
	o := &UserBalance{}
	o.restore(r)
	if !reflect.DeepEqual(o, ub) {
		t.Errorf("Expected %v was  %v", ub, o)
	}
}

func TestUserBalanceStorageStoreRestore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	storageGetter.SetUserBalance(rifsBalance)
	ub1, err := storageGetter.GetUserBalance("other")
	if err != nil || ub1.BalanceMap[CREDIT+OUTBOUND] != rifsBalance.BalanceMap[CREDIT+OUTBOUND] {
		t.Errorf("Expected %v was something else", rifsBalance.BalanceMap[CREDIT+OUTBOUND]) //, ub1.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 200}}
	seconds, credit, bucketList := ub1.getSecondsForPrefix("0723")
	expected := 110.0
	if credit != 200 || seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetPricedSeconds(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Weight: 20, DestinationId: "RET"}

	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	seconds, credit, bucketList := ub1.getSecondsForPrefix("0723")
	expected := 21.0
	if credit != 0 || seconds != expected || len(bucketList) < 2 || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestUserBalanceStorageStore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	storageGetter.SetUserBalance(rifsBalance)
	result, err := storageGetter.GetUserBalance(rifsBalance.Id)
	if err != nil || rifsBalance.Id != result.Id ||
		len(rifsBalance.MinuteBuckets) < 2 || len(result.MinuteBuckets) < 2 ||
		!(rifsBalance.MinuteBuckets[0].Equal(result.MinuteBuckets[0])) ||
		!(rifsBalance.MinuteBuckets[1].Equal(result.MinuteBuckets[1])) ||
		(rifsBalance.BalanceMap[CREDIT+OUTBOUND] != result.BalanceMap[CREDIT+OUTBOUND]) {
		t.Errorf("Expected %v was %v", rifsBalance.MinuteBuckets, result)
	}
}

func TestDebitMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	result := rifsBalance.debitBalance(CREDIT, 6, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 15 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND] {
		t.Errorf("Expected %v was %v", 15, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitAllMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	rifsBalance.debitBalance(CREDIT, 21, false)
	result := rifsBalance.debitBalance(CREDIT, 0, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 0 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND] {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitMoreMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	result := rifsBalance.debitBalance(CREDIT, 22, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND] != -1 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND] {
		t.Errorf("Expected %v was %v", -1, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitNegativeMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	result := rifsBalance.debitBalance(CREDIT, -15, false)
	if rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 36 || result != rifsBalance.BalanceMap[CREDIT+OUTBOUND] {
		t.Errorf("Expected %v was %v", 36, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(6, "0723", false)
	if b2.Seconds != 94 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 94, b2.Seconds)
	}
}

func TestDebitMultipleBucketsMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(105, "0723", false)
	if b2.Seconds != 0 || b1.Seconds != 5 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitAllMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(110, "0723", false)
	if b2.Seconds != 0 || b1.Seconds != 0 || err != nil {
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitMoreMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(115, "0723", false)
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil {
		t.Errorf("Expected %v was %v", 1000, b2.Seconds)
	}
}

func TestDebitPriceMinuteBalance0(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(5, "0723", false)
	if b2.Seconds != 95 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 16 {
		t.Errorf("Expected %v was %v", 16, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitPriceAllMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(21, "0723", false)
	if b2.Seconds != 79 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 0 {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitPriceMoreMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(25, "0723", false)
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil || rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitPriceNegativeMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(-15, "0723", false)
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 36 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 36, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitNegativeMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	err := rifsBalance.debitMinutesBalance(-15, "0723", false)
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT+OUTBOUND] != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBalance.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestDebitSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21, SMS + OUTBOUND: 100}}
	result := rifsBalance.debitBalance(SMS, 12, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND] != 88 || result != rifsBalance.BalanceMap[SMS+OUTBOUND] {
		t.Errorf("Expected %v was %v", 88, rifsBalance.BalanceMap[SMS+OUTBOUND])
	}
}

func TestDebitAllSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21, SMS + OUTBOUND: 100}}
	result := rifsBalance.debitBalance(SMS, 100, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND] != 0 || result != rifsBalance.BalanceMap[SMS+OUTBOUND] {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[SMS+OUTBOUND])
	}
}

func TestDebitMoreSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21, SMS + OUTBOUND: 100}}
	result := rifsBalance.debitBalance(SMS, 110, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND] != -10 || result != rifsBalance.BalanceMap[SMS+OUTBOUND] {
		t.Errorf("Expected %v was %v", -10, rifsBalance.BalanceMap[SMS+OUTBOUND])
	}
}

func TestDebitNegativeSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21, SMS + OUTBOUND: 100}}
	result := rifsBalance.debitBalance(SMS, -15, false)
	if rifsBalance.BalanceMap[SMS+OUTBOUND] != 115 || result != rifsBalance.BalanceMap[SMS+OUTBOUND] {
		t.Errorf("Expected %v was %v", 115, rifsBalance.BalanceMap[SMS+OUTBOUND])
	}
}

func TestUserBalancedebitMinuteBucket(t *testing.T) {
	ub := &UserBalance{
		Id:            "rif",
		Type:          UB_TYPE_POSTPAID,
		BalanceMap:    map[string]float64{SMS: 14, TRAFFIC: 1024},
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, Percent: 0, DestinationId: "RET"}},
	}
	newMb := &MinuteBucket{Weight: 20, Price: 1, DestinationId: "NEW"}
	ub.debitMinuteBucket(newMb)
	if len(ub.MinuteBuckets) != 3 || ub.MinuteBuckets[2] != newMb {
		t.Error("Error adding minute bucket!", len(ub.MinuteBuckets), ub.MinuteBuckets)
	}
}

func TestUserBalancedebitMinuteBucketExists(t *testing.T) {

	ub := &UserBalance{
		Id:            "rif",
		Type:          UB_TYPE_POSTPAID,
		BalanceMap:    map[string]float64{SMS + OUTBOUND: 14, TRAFFIC + OUTBOUND: 1024},
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Seconds: 15, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, Percent: 0, DestinationId: "RET"}},
	}
	newMb := &MinuteBucket{Seconds: -10, Weight: 20, Price: 1, DestinationId: "NAT"}
	ub.debitMinuteBucket(newMb)
	if len(ub.MinuteBuckets) != 2 || ub.MinuteBuckets[0].Seconds != 25 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUserBalanceAddMinuteNil(t *testing.T) {
	ub := &UserBalance{
		Id:            "rif",
		Type:          UB_TYPE_POSTPAID,
		BalanceMap:    map[string]float64{SMS + OUTBOUND: 14, TRAFFIC + OUTBOUND: 1024},
		MinuteBuckets: []*MinuteBucket{&MinuteBucket{Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, Percent: 0, DestinationId: "RET"}},
	}
	ub.debitMinuteBucket(nil)
	if len(ub.MinuteBuckets) != 2 {
		t.Error("Error adding minute bucket!")
	}
}

func TestUserBalanceAddMinutBucketEmpty(t *testing.T) {
	mb1 := &MinuteBucket{Seconds: -10, DestinationId: "NAT"}
	mb2 := &MinuteBucket{Seconds: -10, DestinationId: "NAT"}
	mb3 := &MinuteBucket{Seconds: -10, DestinationId: "OTHER"}
	ub := &UserBalance{}
	ub.debitMinuteBucket(mb1)
	if len(ub.MinuteBuckets) != 1 {
		t.Error("Error adding minute bucket: ", ub.MinuteBuckets)
	}
	ub.debitMinuteBucket(mb2)
	if len(ub.MinuteBuckets) != 1 || ub.MinuteBuckets[0].Seconds != 20 {
		t.Error("Error adding minute bucket: ", ub.MinuteBuckets)
	}
	ub.debitMinuteBucket(mb3)
	if len(ub.MinuteBuckets) != 2 {
		t.Error("Error adding minute bucket: ", ub.MinuteBuckets)
	}
}

/*
func TestUserBalanceExecuteTriggeredActions(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB",
		BalanceMap:     map[string]float64{CREDIT + OUTBOUND: 100},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		MinuteBuckets:  []*MinuteBucket{&MinuteBucket{Seconds: 10, Weight: 20, Price: 1, DestinationId: "NAT"}, &MinuteBucket{Weight: 10, Price: 10, Percent: 0, DestinationId: "RET"}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, Direction: OUTBOUND, ThresholdValue: 2, ActionsId: "TEST_ACTIONS"}},
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Units: 1})
	if ub.BalanceMap[CREDIT+OUTBOUND] != 110 || ub.MinuteBuckets[0].Seconds != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND], ub.MinuteBuckets[0].Seconds)
	}
	// are set to executed
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1})
	if ub.BalanceMap[CREDIT+OUTBOUND] != 110 || ub.MinuteBuckets[0].Seconds != 20 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND], ub.MinuteBuckets[0].Seconds)
	}
	// we can reset them
	ub.resetActionTriggers()
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1})
	if ub.BalanceMap[CREDIT+OUTBOUND] != 120 || ub.MinuteBuckets[0].Seconds != 30 {
		t.Error("Error executing triggered actions", ub.BalanceMap[CREDIT+OUTBOUND], ub.MinuteBuckets[0].Seconds)
	}
}*/

func TestUserBalanceExecuteTriggeredActionsOrder(t *testing.T) {
	ub := &UserBalance{
		Id:             "TEST_UB_OREDER",
		BalanceMap:     map[string]float64{CREDIT + OUTBOUND: 100},
		UnitCounters:   []*UnitsCounter{&UnitsCounter{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1}},
		ActionTriggers: ActionTriggerPriotityList{&ActionTrigger{BalanceId: CREDIT, ThresholdValue: 2, ActionsId: "TEST_ACTIONS_ORDER"}},
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 1})
	if ub.BalanceMap[CREDIT+OUTBOUND] != 10 {
		t.Error("Error executing triggered actions in order", ub.BalanceMap[CREDIT+OUTBOUND])
	}
}

func TestUserBalanceUnitCounting(t *testing.T) {
	ub := &UserBalance{}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 {
		t.Error("Error counting units")
	}
}

func TestUserBalanceUnitCountingOutbound(t *testing.T) {
	ub := &UserBalance{}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 30 {
		t.Error("Error counting units")
	}
}

func TestUserBalanceUnitCountingOutboundInbound(t *testing.T) {
	ub := &UserBalance{}
	ub.countUnits(&Action{BalanceId: CREDIT, Units: 10})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 10 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: OUTBOUND, Units: 10})
	if len(ub.UnitCounters) != 1 && ub.UnitCounters[0].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 {
		t.Error("Error counting units")
	}
	ub.countUnits(&Action{BalanceId: CREDIT, Direction: INBOUND, Units: 10})
	if len(ub.UnitCounters) != 2 && ub.UnitCounters[1].BalanceId != CREDIT || ub.UnitCounters[0].Units != 20 || ub.UnitCounters[1].Units != 10 {
		t.Error("Error counting units")
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Weight: 20, DestinationId: "RET"}

	ub1 := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix("0723")
	}
}

func BenchmarkUserBalanceStorageStoreRestore(b *testing.B) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	for i := 0; i < b.N; i++ {
		storageGetter.SetUserBalance(rifsBalance)
		storageGetter.GetUserBalance(rifsBalance.Id)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT + OUTBOUND: 21}}
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix("0723")
	}
}
