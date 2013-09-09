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

package engine

import (
	"encoding/json"
	"github.com/cgrates/cgrates/cache2go"
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

func TestDestinationStorageStore(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	err := storageGetter.SetDestination(nationale)
	if err != nil {
		t.Error("Error storing destination: ", err)
	}
	result, err := storageGetter.GetDestination(nationale.Id)
	if !reflect.DeepEqual(nationale, result) {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationContainsPrefix(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision, ok := nationale.containsPrefix("0256")
	if !ok || precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}

}

func TestDestinationContainsPrefixLong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision, ok := nationale.containsPrefix("0256723045")
	if !ok || precision != len("0256") {
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
	if _, err := cache2go.GetCached("NAT"); err != nil {
		t.Error("Destination not cached!")
	}
}

func TestDestinationGetNotExists(t *testing.T) {
	d, err := GetDestination("not existing")
	if d != nil {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationGetNotExistsCache(t *testing.T) {
	GetDestination("not existing")
	if d, err := cache2go.GetCached("not existing"); err == nil {
		t.Error("Bad destination cached: ", d)
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkDestinationStorageStoreRestore(b *testing.B) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		storageGetter.SetDestination(nationale)
		storageGetter.GetDestination(nationale.Id)
	}
}
