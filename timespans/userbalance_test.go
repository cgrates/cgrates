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

/*import (
	// "log"
	"reflect"
	"testing"
)

var (
	nationale = &Destination{Id: "nationale", Prefixes: []string{"0257", "0256", "0723"}}
	retea     = &Destination{Id: "retea", Prefixes: []string{"0723", "0724"}}
)

func TestUserBalanceStoreRestore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	s := rifsBalance.store()
	ub1 := &UserBalance{Id: "other"}
	ub1.restore(s)
	if ub1.store() != s {
		t.Errorf("Expected %q was %q", s, ub1.store())
	}
}

func TestGetSecondsForPrefix(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBalance{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 200, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	seconds, bucketList := ub1.getSecondsForPrefix(nil, "0723")
	expected := 110.0
	if seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetPricedSeconds(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Weight: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Weight: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBalance{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	seconds, bucketList := ub1.getSecondsForPrefix(nil, "0723")
	expected := 21.0
	if seconds != expected || bucketList[0].Weight < bucketList[1].Weight {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestUserBalanceKyotoStore(t *testing.T) {
	getter, _ := NewKyotoStorage("../data/test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	getter.SetUserBalance(rifsBalance)
	result, _ := getter.GetUserBalance(rifsBalance.Id)
	if !reflect.DeepEqual(rifsBalance, result) {
		t.Log(rifsBalance)
		t.Log(result)
		t.Errorf("Expected %q was %q", rifsBalance, result)
	}
}

func TestUserBalanceRedisStore(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	getter.SetUserBalance(rifsBalance)
	result, _ := getter.GetUserBalance(rifsBalance.Id)
	if !reflect.DeepEqual(rifsBalance, result) {
		t.Errorf("Expected %q was %q", rifsBalance, result)
	}
}

func TestUserBalanceMongoStore(t *testing.T) {
	getter, err := NewMongoStorage("127.0.0.1", "test")
	if err != nil {
		return
	}
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	getter.SetUserBalance(rifsBalance)
	result, _ := getter.GetUserBalance(rifsBalance.Id)
	if !reflect.DeepEqual(rifsBalance, result) {
		t.Errorf("Expected %q was %q", rifsBalance, result)
	}
}

func TestDebitMoneyBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "o4her", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	result := rifsBalance.debitMoneyBalance(getter, 6)
	if rifsBalance.Credit != 15 || result != rifsBalance.Credit {
		t.Errorf("Expected %v was %v", 15, rifsBalance.Credit)
	}
}

func TestDebitAllMoneyBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	rifsBalance.debitMoneyBalance(getter, 21)
	result := rifsBalance.debitMoneyBalance(getter, 0)
	if rifsBalance.Credit != 0 || result != rifsBalance.Credit {
		t.Errorf("Expected %v was %v", 0, rifsBalance.Credit)
	}
}

func TestDebitMoreMoneyBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	result := rifsBalance.debitMoneyBalance(getter, 22)
	if rifsBalance.Credit != -1 || result != rifsBalance.Credit {
		t.Errorf("Expected %v was %v", -1, rifsBalance.Credit)
	}
}

func TestDebitNegativeMoneyBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	result := rifsBalance.debitMoneyBalance(getter, -15)
	if rifsBalance.Credit != 36 || result != rifsBalance.Credit {
		t.Errorf("Expected %v was %v", 36, rifsBalance.Credit)
	}
}

func TestDebitMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, 6, "0723")
	if b2.Seconds != 94 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 94, b2.Seconds)
	}
}

func TestDebitMultipleBucketsMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, 105, "0723")
	if b2.Seconds != 0 || b1.Seconds != 5 || err != nil {
		t.Log(err)
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitAllMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, 110, "0723")
	if b2.Seconds != 0 || b1.Seconds != 0 || err != nil {
		t.Errorf("Expected %v was %v", 0, b2.Seconds)
	}
}

func TestDebitMoreMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, 115, "0723")
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil {
		t.Errorf("Expected %v was %v", 1000, b2.Seconds)
	}
}

func TestDebitPriceMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, 5, "0723")
	if b2.Seconds != 95 || b1.Seconds != 10 || err != nil || rifsBalance.Credit != 16 {
		t.Log(rifsBalance.Credit)
		t.Errorf("Expected %v was %v", 16, rifsBalance.Credit)
	}
}

func TestDebitPriceAllMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, 21, "0723")
	if b2.Seconds != 79 || b1.Seconds != 10 || err != nil || rifsBalance.Credit != 0 {
		t.Errorf("Expected %v was %v", 0, rifsBalance.Credit)
	}
}

func TestDebitPriceMoreMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, 25, "0723")
	if b2.Seconds != 100 || b1.Seconds != 10 || err == nil || rifsBalance.Credit != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBalance.Credit)
	}
}

func TestDebitPriceNegativeMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 1.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, -15, "0723")
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBalance.Credit != 36 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 36, rifsBalance.Credit)
	}
}

func TestDebitNegativeMinuteBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, ResetDayOfTheMonth: 10}
	err := rifsBalance.debitMinutesBalance(getter, -15, "0723")
	if b2.Seconds != 115 || b1.Seconds != 10 || err != nil || rifsBalance.Credit != 21 {
		t.Log(b1, b2, err)
		t.Errorf("Expected %v was %v", 21, rifsBalance.Credit)
	}
}

func TestDebitSMSBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBalance.debitSMSBuget(getter, 12)
	if rifsBalance.SmsCredit != 88 || result != rifsBalance.SmsCredit || err != nil {
		t.Errorf("Expected %v was %v", 88, rifsBalance.SmsCredit)
	}
}

func TestDebitAllSMSBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBalance.debitSMSBuget(getter, 100)
	if rifsBalance.SmsCredit != 0 || result != rifsBalance.SmsCredit || err != nil {
		t.Errorf("Expected %v was %v", 0, rifsBalance.SmsCredit)
	}
}

func TestDebitMoreSMSBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBalance.debitSMSBuget(getter, 110)
	if rifsBalance.SmsCredit != 100 || result != rifsBalance.SmsCredit || err == nil {
		t.Errorf("Expected %v was %v", 100, rifsBalance.SmsCredit)
	}
}

func TestDebitNegativeSMSBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.0, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, SmsCredit: 100, ResetDayOfTheMonth: 10}
	result, err := rifsBalance.debitSMSBuget(getter, -15)
	if rifsBalance.SmsCredit != 115 || result != rifsBalance.SmsCredit || err != nil {
		t.Errorf("Expected %v was %v", 115, rifsBalance.SmsCredit)
	}
}

func TestResetUserBalance(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	rifsBalance.MinuteBuckets[0].Seconds, rifsBalance.MinuteBuckets[1].Seconds = 0.0, 0.0
	err := rifsBalance.resetUserBalance(getter)
	if err != nil ||
		rifsBalance.MinuteBuckets[0] == b1 ||
		rifsBalance.MinuteBuckets[0].Seconds != seara.MinuteBuckets[0].Seconds ||
		rifsBalance.MinuteBuckets[1].Seconds != seara.MinuteBuckets[1].Seconds ||
		rifsBalance.SmsCredit != seara.SmsCredit {
		t.Log(rifsBalance.MinuteBuckets[0])
		t.Log(rifsBalance.MinuteBuckets[1])
		t.Log(rifsBalance.SmsCredit)
		t.Log(rifsBalance.Traffic)
		t.Errorf("Expected %v was %v", seara, rifsBalance)
	}

}

func TestGetVolumeDiscountHaving(t *testing.T) {
	vd := &VolumeDiscount{100, 11}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	rifsBalance := &UserBalance{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, VolumeDiscountSeconds: 100}
	result, err := rifsBalance.getVolumeDiscount(nil)
	if err != nil || result != 11 {
		t.Errorf("Expected %v was %v", 11, result)
	}
}

func TestGetVolumeDiscountNotHaving(t *testing.T) {
	vd := &VolumeDiscount{100, 11}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	rifsBalance := &UserBalance{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, VolumeDiscountSeconds: 99}
	result, err := rifsBalance.getVolumeDiscount(nil)
	if err != nil || result != 0 {
		t.Errorf("Expected %v was %v", 0, result)
	}
}

func TestGetVolumeDiscountSteps(t *testing.T) {
	vd1 := &VolumeDiscount{100, 11}
	vd2 := &VolumeDiscount{500, 20}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, VolumeDiscountThresholds: []*VolumeDiscount{vd1, vd2}}
	rifsBalance := &UserBalance{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, VolumeDiscountSeconds: 551}
	result, err := rifsBalance.getVolumeDiscount(nil)
	if err != nil || result != 20 {
		t.Errorf("Expected %v was %v", 20, result)
	}
}

func TestRecivedCallsBonus(t *testing.T) {
	getter, _ := NewKyotoStorage("../data/test.kch")
	defer getter.Close()
	rcb := &RecivedCallBonus{Credit: 100}
	seara := &TariffPlan{Id: "seara_voo", SmsCredit: 100, ReceivedCallSecondsLimit: 10, RecivedCallBonus: rcb}
	rifsBalance := &UserBalance{Id: "other", Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10, ReceivedCallSeconds: 1}
	err := rifsBalance.addReceivedCallSeconds(getter, 12)
	if err != nil || rifsBalance.Credit != 121 || rifsBalance.ReceivedCallSeconds != 3 {
		t.Error("Wrong Received call bonus procedure: ", rifsBalance)
	}
}*/

/*********************************** Benchmarks *******************************/

/*func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Weight: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Weight: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix(nil, "0723")
	}
}

func BenchmarkUserBalanceKyotoStoreRestore(b *testing.B) {
	getter, _ := NewKyotoStorage("../data/test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		getter.SetUserBalance(rifsBalance)
		getter.GetUserBalance(rifsBalance.Id)
	}
}

func BenchmarkUserBalanceRedisStoreRestore(b *testing.B) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		getter.SetUserBalance(rifsBalance)
		getter.GetUserBalance(rifsBalance.Id)
	}
}

func BenchmarkUserBalanceMongoStoreRestore(b *testing.B) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBalance := &UserBalance{Id: "other", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		getter.SetUserBalance(rifsBalance)
		getter.GetUserBalance(rifsBalance.Id)
	}
}

func BenchmarkGetSecondsForPrefix(b *testing.B) {
	b1 := &MinuteBucket{Seconds: 10, Weight: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Weight: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBalance{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 200, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		ub1.getSecondsForPrefix(nil, "0723")
	}
}
*/
