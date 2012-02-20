package timespans

import (
	"testing"
)

func TestDestinationStoreRestore(t *testing.T) {
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	s := nationale.store()
	d1 := &Destination{Id: "nationale"}
	d1.restore(s)
	if d1.store() != s {
		t.Errorf("Expected %q was %q", s, d1.store())
	}
}

func TestKyotoStore(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	getter.SetDestination(nationale)
	result, _ := getter.GetDestination(nationale.Id)
	if result.Id != nationale.Id || result.Prefixes[2] != nationale.Prefixes[2] {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestRedisStore(t *testing.T) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	getter.SetDestination(nationale)
	result, _ := getter.GetDestination(nationale.Id)
	if result.Id != nationale.Id || result.Prefixes[2] != nationale.Prefixes[2] {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestMongoStore(t *testing.T) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	getter.SetDestination(nationale)
	result, _ := getter.GetDestination(nationale.Id)
	if result.Id != nationale.Id || result.Prefixes[2] != nationale.Prefixes[2] {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkDestinationKyotoStoreRestore(b *testing.B) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		getter.SetDestination(nationale)
		getter.GetDestination(nationale.Id)
	}
}

func BenchmarkDestinationRedisStoreRestore(b *testing.B) {
	getter, _ := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		getter.SetDestination(nationale)
		getter.GetDestination(nationale.Id)
	}
}

func BenchmarkDestinationMongoStoreRestore(b *testing.B) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		getter.SetDestination(nationale)
		getter.GetDestination(nationale.Id)
	}
}
