/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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

package timespans

import (
	// "log"
	"reflect"
	"testing"
)

var (
	NAT = &Destination{Id: "NAT", Prefixes: []string{"0257", "0256", "0723"}}
	RET = &Destination{Id: "RET", Prefixes: []string{"0723", "0724"}}
)

func init() {
	getter, _ = NewRedisStorage("tcp:127.0.0.1:6379", 10)
	SetStorageGetter(getter)
}

func TestUserBalanceStoreRestore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	getter.SetUserBalance(rifsBalance)
	ub1, err := getter.GetUserBalance("other")
	if err != nil || ub1.BalanceMap[CREDIT] != rifsBalance.BalanceMap[CREDIT] {
		t.Errorf("Expected %v was %v", rifsBalance.BalanceMap[CREDIT], ub1.BalanceMap[CREDIT])
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 200}}
	seconds, bucketList := ub1.getSecondsForPrefix("0723")
	expected := 110.0
	if seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetPricedSeconds(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Weight: 20, DestinationId: "RET"}

	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	seconds, bucketList := ub1.getSecondsForPrefix("0723")
	expected := 21.0
	if seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestUserBalanceRedisStore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	getter.SetUserBalance(rifsBalance)
	result, _ := getter.GetUserBalance(rifsBalance.Id)
	if !reflect.DeepEqual(rifsBalance, result) {
		t.Errorf("Expected %v was %v", rifsBalance, result)
	}
}

func TestDebitMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "o4her", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	result := rifsBalance.debitMoneyBalance(6)
	if rifsBalance.BalanceMap[CREDIT] != 15 || result != rifsBalance.BalanceMap[CREDIT] {
		t.Errorf("Expected %v was %v", 15, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitAllMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	rifsBalance.debitMoneyBalance(21)
	result := rifsBalance.debitMoneyBalance(0)
	if rifsBalance.BalanceMap[CREDIT] != 0 || result != rifsBalance.BalanceMap[CREDIT] {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitMoreMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	result := rifsBalance.debitMoneyBalance(22)
	if rifsBalance.BalanceMap[CREDIT] != -1 || result != rifsBalance.BalanceMap[CREDIT] {
		t.Errorf("Expected %v was %v", -1, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitNegativeMoneyBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	result := rifsBalance.debitMoneyBalance(-15)
	if rifsBalance.BalanceMap[CREDIT] != 36 || result != rifsBalance.BalanceMap[CREDIT] {
		t.Errorf("Expected %v was %v", 36, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(6, "0723")
	if b2.Seconds != 94 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 94, b2.Seconds)
	}
}

func TestDebitMultipleBucketsMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(105, "0723")
	if b2.Seconds != 0 || b1.Seconds != 5 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitAllMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(110, "0723")
	if b2.Seconds != 0 || b1.Seconds != 0 || err != nil {
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitMoreMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(115, "0723")
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil {
		t.Errorf("Expected %v was %v", 1000, b2.Seconds)
	}
}

func TestDebitPriceMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(5, "0723")
	if b2.Seconds != 95 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT] != 16 {
		t.Log(rifsBalance.BalanceMap[CREDIT])
		t.Errorf("Expected %v was %v", 16, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitPriceAllMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(21, "0723")
	if b2.Seconds != 79 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT] != 0 {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitPriceMoreMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(25, "0723")
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil || rifsBalance.BalanceMap[CREDIT] != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitPriceNegativeMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(-15, "0723")
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT] != 36 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 36, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitNegativeMinuteBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	err := rifsBalance.debitMinutesBalance(-15, "0723")
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBalance.BalanceMap[CREDIT] != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBalance.BalanceMap[CREDIT])
	}
}

func TestDebitSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21, SMS: 100}}
	result, err := rifsBalance.debitSMSBuget(12)
	if rifsBalance.BalanceMap[SMS] != 88 || result != rifsBalance.BalanceMap[SMS] || err != nil {
		t.Errorf("Expected %v was %v", 88, rifsBalance.BalanceMap[SMS])
	}
}

func TestDebitAllSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21, SMS: 100}}
	result, err := rifsBalance.debitSMSBuget(100)
	if rifsBalance.BalanceMap[SMS] != 0 || result != rifsBalance.BalanceMap[SMS] || err != nil {
		t.Errorf("Expected %v was %v", 0, rifsBalance.BalanceMap[SMS])
	}
}

func TestDebitMoreSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21, SMS: 100}}
	result, err := rifsBalance.debitSMSBuget(110)
	if rifsBalance.BalanceMap[SMS] != 100 || result != rifsBalance.BalanceMap[SMS] || err == nil {
		t.Errorf("Expected %v was %v", 100, rifsBalance.BalanceMap[SMS])
	}
}

func TestDebitNegativeSMSBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21, SMS: 100}}
	result, err := rifsBalance.debitSMSBuget(-15)
	if rifsBalance.BalanceMap[SMS] != 115 || result != rifsBalance.BalanceMap[SMS] || err != nil {
		t.Errorf("Expected %v was %v", 115, rifsBalance.BalanceMap[SMS])
	}
}

/*func TestResetUserBalance(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	rifsBalance.MinuteBuckets[0].Seconds, rifsBalance.MinuteBuckets[1].Seconds = 0.0, 0.0
	err := rifsBalance.resetUserBalance(getter)
	if err != nil ||
		rifsBalance.MinuteBuckets[0] == b1 ||
		rifsBalance.BalanceMap[SMS] != seara.SmsCredit {
		t.Log(rifsBalance.MinuteBuckets[0])
		t.Log(rifsBalance.MinuteBuckets[1])
		t.Log(rifsBalance.BalanceMap[SMS])
		t.Log(rifsBalance.BalanceMap[TRAFFIC])
		t.Errorf("Expected %v was %v", "xxx", rifsBalance)
	}

}*/

/*func TestGetVolumeDiscountHaving(t *testing.T) {
	vd := &VolumeDiscount{100, 11}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]float64{CREDIT: 21}, tariffPlan: seara, VolumeDiscountSeconds: 100}
	result, err := rifsBalance.getVolumeDiscount(nil)
	if err != nil || result != 11 {
		t.Errorf("Expected %v was %v", 11, result)
	}
}

func TestGetVolumeDiscountNotHaving(t *testing.T) {
	vd := &VolumeDiscount{100, 11}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]float64{CREDIT: 21}, tariffPlan: seara, VolumeDiscountSeconds: 99}
	result, err := rifsBalance.getVolumeDiscount(nil)
	if err != nil || result != 0 {
		t.Errorf("Expected %v was %v", 0, result)
	}
}

func TestGetVolumeDiscountSteps(t *testing.T) {
	vd1 := &VolumeDiscount{100, 11}
	vd2 := &VolumeDiscount{500, 20}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd1, vd2}}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]float64{CREDIT: 21}, tariffPlan: seara, VolumeDiscountSeconds: 551}
	result, err := rifsBalance.getVolumeDiscount(nil)
	if err != nil || result != 20 {
		t.Errorf("Expected %v was %v", 20, result)
	}
}

func TestRecivedCallsBonus(t *testing.T) {
	_ := NewKyotoStorage("../data/test.kch")
	defer getter.Close()
	rcb := &RecivedCallBonus{Credit: 100}
	seara := &TariffPlan{Id: "seara_voo", SmsCredit: 100, ReceivedCallSecondsLimit: 10, RecivedCallBonus: rcb}
	rifsBalance := &UserBalance{Id: "other", BalanceMap: map[string]float64{CREDIT: 21}, tariffPlan: seara, ReceivedCallSeconds: 1}
	err := rifsBalance.addReceivedCallSeconds(12)
	if err != nil || rifsBalance.BalanceMap[CREDIT] != 121 || rifsBalance.ReceivedCallSeconds != 3 {
		t.Error("Wrong Received call bonus procedure: ", rifsBalance)
	}
}*/

func TestUBAddMinutBucket(t *testing.T) {
	mb1 := &MinuteBucket{Seconds: 10, DestinationId: "NAT"}
	mb2 := &MinuteBucket{Seconds: 10, DestinationId: "NAT"}
	mb3 := &MinuteBucket{Seconds: 10, DestinationId: "OTHER"}
	ub := &UserBalance{}
	ub.addMinuteBucket(mb1)
	if len(ub.MinuteBuckets) != 1 {
		t.Error("Error adding minute bucket: ", ub.MinuteBuckets)
	}
	ub.addMinuteBucket(mb2)
	if len(ub.MinuteBuckets) != 1 || ub.MinuteBuckets[0].Seconds != 20 {
		t.Error("Error adding minute bucket: ", ub.MinuteBuckets)
	}
	ub.addMinuteBucket(mb3)
	if len(ub.MinuteBuckets) != 2 {
		t.Error("Error adding minute bucket: ", ub.MinuteBuckets)
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Weight: 20, DestinationId: "RET"}

	ub1 := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix("0723")
	}
}

func BenchmarkUserBalanceRedisStoreRestore(b *testing.B) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "RET"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	for i := 0; i < b.N; i++ {
		getter.SetUserBalance(rifsBalance)
		getter.GetUserBalance(rifsBalance.Id)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, DestinationId: "NAT"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, DestinationId: "RET"}
	ub1 := &UserBalance{Id: "OUT:CUSTOMER_1:rif", MinuteBuckets: []*MinuteBucket{b1, b2}, BalanceMap: map[string]float64{CREDIT: 21}}
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix("0723")
	}
}
