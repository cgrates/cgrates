package timespans

import (
	"testing"
)

func TestTariffPlanStoreRestore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100,  MinuteBuckets: []*MinuteBucket{b1, b2}}
	s := seara.store()
	tp1 := &TariffPlan{Id: "seara"}
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
	seara := &TariffPlan{Id: "seara", SmsCredit: 100,  MinuteBuckets: []*MinuteBucket{b1, b2}}
	getter.SetTariffPlan(seara)
	result, _ := getter.GetTariffPlan(seara.Id)
	if result.SmsCredit != seara.SmsCredit || len(result.MinuteBuckets) != len(seara.MinuteBuckets) {
		t.Errorf("Expected %q was %q", seara, result)
	}
}

func TestTariffPlanRedisStore(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100,  MinuteBuckets: []*MinuteBucket{b1, b2}}
	getter.SetTariffPlan(seara)
	result, _ := getter.GetTariffPlan(seara.Id)
	if result.SmsCredit != seara.SmsCredit || len(result.MinuteBuckets) != len(seara.MinuteBuckets) {
		t.Errorf("Expected %q was %q", seara, result)
	}
}

func TestTariffPlanMongoStore(t *testing.T) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100,  MinuteBuckets: []*MinuteBucket{b1, b2}}
	getter.SetTariffPlan(seara)
	result, _ := getter.GetTariffPlan(seara.Id)
	if result.SmsCredit != seara.SmsCredit || len(result.MinuteBuckets) != len(seara.MinuteBuckets) {
		t.Errorf("Expected %q was %q", seara, result)
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkTariffPlanKyotoStoreRestore(b *testing.B) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100,  MinuteBuckets: []*MinuteBucket{b1, b2}}
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
	seara := &TariffPlan{Id: "seara", SmsCredit: 100,  MinuteBuckets: []*MinuteBucket{b1, b2}}
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
	seara := &TariffPlan{Id: "seara", SmsCredit: 100,  MinuteBuckets: []*MinuteBucket{b1, b2}}
	for i := 0; i < b.N; i++ {
		getter.SetTariffPlan(seara)
		getter.GetTariffPlan(seara.Id)
	}
}
