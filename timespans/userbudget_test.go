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
	nationale = &Destination{Id: "nationale", Prefixes: []string{"0257", "0256", "0723"}}
	retea     = &Destination{Id: "retea", Prefixes: []string{"0723", "0724"}}
)

func TestUserBudgetStoreRestore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	s := rifsBudget.store()
	ub1 := &UserBudget{Id: "other"}
	ub1.restore(s)
	if ub1.store() != s {
		t.Errorf("Expected %q was %q", s, ub1.store())
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 200, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	seconds, bucketList := ub1.getSecondsForPrefix(nil, "0723")
	expected := 110.0
	if seconds != expected || bucketList[0].Priority < bucketList[1].Priority {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetPricedSeconds(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Priority: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Priority: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	seconds, bucketList := ub1.getSecondsForPrefix(nil, "0723")
	expected := 21.0
	if seconds != expected || bucketList[0].Priority < bucketList[1].Priority {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestUserBudgetKyotoStore(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	getter.SetUserBudget(rifsBudget)
	result, _ := getter.GetUserBudget(rifsBudget.Id)
	if !reflect.DeepEqual(rifsBudget, result) {
		t.Log(rifsBudget)
		t.Log(result)
		t.Errorf("Expected %q was %q", rifsBudget, result)
	}
}

func TestUserBudgetRedisStore(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	getter.SetUserBudget(rifsBudget)
	result, _ := getter.GetUserBudget(rifsBudget.Id)
	if !reflect.DeepEqual(rifsBudget, result) {
		t.Errorf("Expected %q was %q", rifsBudget, result)
	}
}

func TestUserBudgetMongoStore(t *testing.T) {
	getter, err := NewMongoStorage("127.0.0.1", "test")
	if err != nil {
		return
	}
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	getter.SetUserBudget(rifsBudget)
	result, _ := getter.GetUserBudget(rifsBudget.Id)
	if !reflect.DeepEqual(rifsBudget, result) {
		t.Errorf("Expected %q was %q", rifsBudget, result)
	}
}

func TestDebitMoneyBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "o4her", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	result := rifsBudget.debitMoneyBudget(getter, 6)
	if rifsBudget.Credit != 15 || result != rifsBudget.Credit {
		t.Errorf("Expected %v was %v", 15, rifsBudget.Credit)
	}
}

func TestDebitAllMoneyBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	rifsBudget.debitMoneyBudget(getter, 21)
	result := rifsBudget.debitMoneyBudget(getter, 0)
	if rifsBudget.Credit != 0 || result != rifsBudget.Credit {
		t.Errorf("Expected %v was %v", 0, rifsBudget.Credit)
	}
}

func TestDebitMoreMoneyBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	result := rifsBudget.debitMoneyBudget(getter, 22)
	if rifsBudget.Credit != -1 || result != rifsBudget.Credit {
		t.Errorf("Expected %v was %v", -1, rifsBudget.Credit)
	}
}

func TestDebitNegativeMoneyBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	result := rifsBudget.debitMoneyBudget(getter, -15)
	if rifsBudget.Credit != 36 || result != rifsBudget.Credit {
		t.Errorf("Expected %v was %v", 36, rifsBudget.Credit)
	}
}

func TestDebitMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, 6, "0723")
	if b2.Seconds != 94 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 94, b2.Seconds)
	}
}

func TestDebitMultipleBucketsMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, 105, "0723")
	if b2.Seconds != 0 || b1.Seconds != 5 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitAllMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, 110, "0723")
	if b2.Seconds != 0 || b1.Seconds != 0 || err != nil {
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitMoreMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, 115, "0723")
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil {
		t.Errorf("Expected %v was %v", 1000, b2.Seconds)
	}
}

func TestDebitPriceMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 1.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, 5, "0723")
	if b2.Seconds != 95 || b1.Seconds != 10 || err != nil || rifsBudget.Credit != 16 {
		t.Log(rifsBudget.Credit)
		t.Errorf("Expected %v was %v", 16, rifsBudget.Credit)
	}
}

func TestDebitPriceAllMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 1.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, 21, "0723")
	if b2.Seconds != 79 || b1.Seconds != 10 || err != nil || rifsBudget.Credit != 0 {
		t.Errorf("Expected %v was %v", 0, rifsBudget.Credit)
	}
}

func TestDebitPriceMoreMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 1.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, 25, "0723")
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil || rifsBudget.Credit != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBudget.Credit)
	}
}

func TestDebitPriceNegativeMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 1.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, -15, "0723")
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBudget.Credit != 36 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 36, rifsBudget.Credit)
	}
}

func TestDebitNegativeMinuteBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBudget.debitMinutesBudget(getter, -15, "0723")
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBudget.Credit != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBudget.Credit)
	}
}

func TestDebitSMSBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBudget.debitSMSBuget(getter, 12)
	if rifsBudget.SmsCredit != 88 || result != rifsBudget.SmsCredit || err != nil {
		t.Errorf("Expected %v was %v", 88, rifsBudget.SmsCredit)
	}
}

func TestDebitAllSMSBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBudget.debitSMSBuget(getter, 100)
	if rifsBudget.SmsCredit != 0 || result != rifsBudget.SmsCredit || err != nil {
		t.Errorf("Expected %v was %v", 0, rifsBudget.SmsCredit)
	}
}

func TestDebitMoreSMSBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBudget.debitSMSBuget(getter, 110)
	if rifsBudget.SmsCredit != 100 || result != rifsBudget.SmsCredit || err == nil {
		t.Errorf("Expected %v was %v", 100, rifsBudget.SmsCredit)
	}
}

func TestDebitNegativeSMSBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBudget.debitSMSBuget(getter, -15)
	if rifsBudget.SmsCredit != 115 || result != rifsBudget.SmsCredit || err != nil {
		t.Errorf("Expected %v was %v", 115, rifsBudget.SmsCredit)
	}
}

func TestResetUserBudget(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	rifsBudget.MinuteBuckets[0].Seconds, rifsBudget.MinuteBuckets[1].Seconds = 0.0, 0.0
	err := rifsBudget.resetUserBudget(getter)
	if err != nil ||
		rifsBudget.MinuteBuckets[0] == b1 ||
		rifsBudget.MinuteBuckets[0].Seconds != seara.MinuteBuckets[0].Seconds ||
		rifsBudget.MinuteBuckets[1].Seconds != seara.MinuteBuckets[1].Seconds ||
		rifsBudget.SmsCredit != seara.SmsCredit {
		t.Log(rifsBudget.MinuteBuckets[0])
		t.Log(rifsBudget.MinuteBuckets[1])
		t.Log(rifsBudget.SmsCredit)
		t.Log(rifsBudget.Traffic)
		t.Errorf("Expected %v was %v", seara, rifsBudget)
	}

}

func TestGetVolumeDiscountHaving(t *testing.T) {
	vd := &VolumeDiscount{100, 11}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	rifsBudget := &UserBudget{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, VolumeDiscountSeconds: 100}
	result, err := rifsBudget.getVolumeDiscount(nil)
	if err != nil || result != 11 {
		t.Errorf("Expected %v was %v", 11, result)
	}
}

func TestGetVolumeDiscountNotHaving(t *testing.T) {
	vd := &VolumeDiscount{100, 11}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	rifsBudget := &UserBudget{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, VolumeDiscountSeconds: 99}
	result, err := rifsBudget.getVolumeDiscount(nil)
	if err != nil || result != 0 {
		t.Errorf("Expected %v was %v", 0, result)
	}
}

func TestGetVolumeDiscountSteps(t *testing.T) {
	vd1 := &VolumeDiscount{100, 11}
	vd2 := &VolumeDiscount{500, 20}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd1, vd2}}
	rifsBudget := &UserBudget{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, VolumeDiscountSeconds: 551}
	result, err := rifsBudget.getVolumeDiscount(nil)
	if err != nil || result != 20 {
		t.Errorf("Expected %v was %v", 20, result)
	}
}

func TestRecivedCallsBonus(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	rcb := &RecivedCallBonus{Credit: 100}
	seara := &TariffPlan{Id: "seara_voo", SmsCredit: 100, ReceivedCallSecondsLimit: 10, RecivedCallBonus: rcb}
	rifsBudget := &UserBudget{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, ReceivedCallSeconds: 1}
	err := rifsBudget.addReceivedCallSeconds(getter, 12)
	if err != nil || rifsBudget.Credit != 121 || rifsBudget.ReceivedCallSeconds != 3 {
		t.Error("Wrong Received call bonus procedure: ", rifsBudget)
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Priority: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Priority: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix(nil, "0723")
	}
}

func BenchmarkUserBudgetKyotoStoreRestore(b *testing.B) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		getter.SetUserBudget(rifsBudget)
		getter.GetUserBudget(rifsBudget.Id)
	}
}

func BenchmarkUserBudgetRedisStoreRestore(b *testing.B) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		getter.SetUserBudget(rifsBudget)
		getter.GetUserBudget(rifsBudget.Id)
	}
}

func BenchmarkUserBudgetMongoStoreRestore(b *testing.B) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		getter.SetUserBudget(rifsBudget)
		getter.GetUserBudget(rifsBudget.Id)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 200, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix(nil, "0723")
	}
}
