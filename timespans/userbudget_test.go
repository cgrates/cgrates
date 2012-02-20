package timespans

import (
	"testing"
	//"log"
)

var (
	nationale = &Destination{Id: "nationale", Prefixes: []string{"0257", "0256", "0723"}}
	retea     = &Destination{Id: "retea", Prefixes: []string{"0723", "0724"}}
)

func TestGetSeconds(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 200, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	seconds := ub1.GetSecondsForPrefix(nil, "0723")
	expected := 100
	if seconds != expected {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestGetPricedSeconds(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Priority: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Priority: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	seconds := ub1.GetSecondsForPrefix(nil, "0723")
	expected := 21
	if seconds != expected {
		t.Errorf("Expected %v was %v", expected, seconds)
	}
}

func TestUserBudgetStoreRestore(t *testing.T) {
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	s := rifsBudget.store()
	ub1 := &UserBudget{Id: "rif"}
	ub1.restore(s)
	if ub1.store() != s {
		t.Errorf("Expected %q was %q", s, ub1.store())
	}
}

func TestUserBudgetKyotoStore(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	getter.SetUserBudget(rifsBudget)
	result, _ := getter.GetUserBudget(rifsBudget.Id)
	if result.SmsCredit != rifsBudget.SmsCredit || len(result.MinuteBuckets) != len(rifsBudget.MinuteBuckets) {
		t.Errorf("Expected %q was %q", rifsBudget, result)
	}
}

func TestUserBudgetRedisStore(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	getter.SetUserBudget(rifsBudget)
	result, _ := getter.GetUserBudget(rifsBudget.Id)
	if result.SmsCredit != rifsBudget.SmsCredit || len(result.MinuteBuckets) != len(rifsBudget.MinuteBuckets) {
		t.Errorf("Expected %q was %q", rifsBudget, result)
	}
}

func TestUserBudgetMongoStore(t *testing.T) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	getter.SetUserBudget(rifsBudget)
	result, _ := getter.GetUserBudget(rifsBudget.Id)
	if result.SmsCredit != rifsBudget.SmsCredit || len(result.MinuteBuckets) != len(rifsBudget.MinuteBuckets) {
		t.Errorf("Expected %q was %q", rifsBudget, result)
	}
}

/*********************************** Benchmarks *******************************/

func BenchmarkGetSecondForPrefix(b *testing.B) {
	b.StopTimer()
	b1 := &MinuteBucket{Seconds: 10, Price: 10, Priority: 10, destination: nationale}
	b2 := &MinuteBucket{Seconds: 100, Price: 1, Priority: 20, destination: retea}
	tf1 := &TariffPlan{MinuteBuckets: []*MinuteBucket{b1, b2}}

	ub1 := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: tf1, ResetDayOfTheMonth: 10}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ub1.GetSecondsForPrefix(nil, "0723")
	}
}

func BenchmarkUserBudgetKyotoStoreRestore(b *testing.B) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	b1 := &MinuteBucket{Seconds: 10, Priority: 10, Price: 0.01, DestinationId: "nationale"}
	b2 := &MinuteBucket{Seconds: 100, Priority: 20, Price: 0.0, DestinationId: "retea"}
	seara := &TariffPlan{Id: "seara", SmsCredit: 100, MinuteBuckets: []*MinuteBucket{b1, b2}}
	rifsBudget := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
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
	rifsBudget := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
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
	rifsBudget := &UserBudget{Id: "rif", MinuteBuckets: []*MinuteBucket{b1, b2}, Credit: 21, tariffPlan: seara, ResetDayOfTheMonth: 10}
	for i := 0; i < b.N; i++ {
		getter.SetUserBudget(rifsBudget)
		getter.GetUserBudget(rifsBudget.Id)
	}
}
