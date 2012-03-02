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

func TestTariffPlanStoreRestore(t *testing.T) {
        b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
        b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
        rcb := &RecivedCallBonus{Credit: 100}
        vd := &VolumeDiscount{100, 10}
        seara := &TariffPlan{Id: "seara_voo",
                SmsCredit:                100,
                ReceivedCallSecondsLimit: 0,
                RecivedCallBonus:         rcb,
                MinuteBuckets:            []*MinuteBucket{b1, b2},
                VolumeDiscountThresholds: []*VolumeDiscount{vd}}
        s := seara.store()
        tp1 := &TariffPlan{Id: "seara_voo"}
        tp1.restore(s)
        if tp1.store() != s {
                t.Errorf("Expected %q was %q", s, tp1.store())
        }
}



func TestTariffPlanKyotoStore(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	vd := &VolumeDiscount{100, 10}
	seara := &TariffPlan{Id: "seara_voo", SmsCredit: 100, ReceivedCallSecondsLimit: 0,
		MinuteBuckets: []*MinuteBucket{b1, b2}, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	getter.SetTariffPlan(seara)
	result, _ := getter.GetTariffPlan(seara.Id)
	if !reflect.DeepEqual(seara, result) {
		t.Errorf("Expected %q was %q", seara, result)
	}
}

func TestTariffPlanRedisStore(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	vd := &VolumeDiscount{100, 10}
	seara := &TariffPlan{Id: "seara_voo", SmsCredit: 100, ReceivedCallSecondsLimit: 0,
		MinuteBuckets: []*MinuteBucket{b1, b2}, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	getter.SetTariffPlan(seara)
	result, _ := getter.GetTariffPlan(seara.Id)
	if !reflect.DeepEqual(seara, result) {
		t.Errorf("Expected %q was %q", seara, result)
	}
}

func TestTariffPlanMongoStore(t *testing.T) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	vd := &VolumeDiscount{100, 10}
	seara := &TariffPlan{Id: "seara_voo", SmsCredit: 100, ReceivedCallSecondsLimit: 0,
		MinuteBuckets: []*MinuteBucket{b1, b2}, VolumeDiscountThresholds: []*VolumeDiscount{vd}}
	getter.SetTariffPlan(seara)
	result, _ := getter.GetTariffPlan(seara.Id)
	if reflect.DeepEqual(seara, result) {
		t.Log(seara)
		t.Log(result)
		t.Errorf("Expected %q was %q", seara, result)
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkTariffPlanKyotoStoreRestore(b *testing.B) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara_other", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	for i := 0; i < b.N; i++ {
		getter.SetTariffPlan(seara)
		getter.GetTariffPlan(seara.Id)
	}
}

func BenchmarkTariffPlanRedisStoreRestore(b *testing.B) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara_other", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	for i := 0; i < b.N; i++ {
		getter.SetTariffPlan(seara)
		getter.GetTariffPlan(seara.Id)
	}
}

func BenchmarkTariffPlanMongoStoreRestore(b *testing.B) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara_other", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	for i := 0; i < b.N; i++ {
		getter.SetTariffPlan(seara)
		getter.GetTariffPlan(seara.Id)
	}
}
