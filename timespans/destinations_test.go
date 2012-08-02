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
	"encoding/json"
	"reflect"
	"testing"
)

func TestDestinationStoreRestore(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	s, _ := json.Marshal(nationale)
	d1 := &Destination{Id: "nat"}
	json.Unmarshal(s, d1)
	s1, _ := json.Marshal(d1)
	if string(s1) != string(s) {
		t.Errorf("Expected %q was %q", s, s1)
	}
}

func TestDestinationRedisStore(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	storageGetter.SetDestination(nationale)
	result, _ := storageGetter.GetDestination(nationale.Id)
	if !reflect.DeepEqual(nationale, result) {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationContainsPrefix(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	contains, precision := nationale.containsPrefix("0256")
	if !contains || precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}

}

func TestDestinationGetExists(t *testing.T) {
	d, err := GetDestination("NAT")
	if err != nil || d == nil {
		t.Error("Could not get destination: ", d)
	}
}

func TestDestinationGetExistsCache(t *testing.T) {
	GetDestination("NAT")
	if _, exists := DestinationCacheMap["NAT"]; !exists {
		t.Error("Destination not cached!")
	}
}

func TestDestinationGetNotExists(t *testing.T) {
	d, err := GetDestination("not existing")
	if d != nil {
		t.Error("Got false destination: ", err)
	}
}

func TestDestinationGetNotExistsCache(t *testing.T) {
	GetDestination("not existing")
	if _, exists := DestinationCacheMap["not existing"]; exists {
		t.Error("Bad destination cached")
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkDestinationRedisStoreRestore(b *testing.B) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		storageGetter.SetDestination(nationale)
		storageGetter.GetDestination(nationale.Id)
	}
}
