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
	"reflect"
	"testing"
)

func TestDestinationStoreRestore(t *testing.T) {
        nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
        s := nationale.store()
        d1 := &Destination{Id: "nat"}
        d1.restore(s)
        if d1.store() != s {
                t.Errorf("Expected %q was %q", s, d1.store())
        }
}

func TestDestinationKyotoStore(t *testing.T) {
	getter, _ := NewKyotoStorage("test.kch")
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	getter.SetDestination(nationale)
	result, _ := getter.GetDestination(nationale.Id)
	if !reflect.DeepEqual(nationale, result) {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationRedisStore(t *testing.T) {
	getter, err := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	if err != nil {
		return
	}
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	getter.SetDestination(nationale)
	result, _ := getter.GetDestination(nationale.Id)
	if !reflect.DeepEqual(nationale, result) {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationMongoStore(t *testing.T) {
	getter, err := NewMongoStorage("127.0.0.1", "test")
	if err != nil {
		return
	}
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	getter.SetDestination(nationale)
	result, _ := getter.GetDestination(nationale.Id)
	if !reflect.DeepEqual(nationale, result) {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationContainsPrefix(t *testing.T) {
	getter, _ := NewMongoStorage("127.0.0.1", "test")
	defer getter.Close()
	nationale = &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	contains, precision := nationale.containsPrefix("0256")
	if !contains || precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
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
