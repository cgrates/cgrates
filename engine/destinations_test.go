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
	err := dataStorage.SetDestination(nationale)
	if err != nil {
		t.Error("Error storing destination: ", err)
	}
	result, err := dataStorage.GetDestination(nationale.Id)
	if nationale.containsPrefix("0257") == 0 || nationale.containsPrefix("0256") == 0 || nationale.containsPrefix("0723") == 0 {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationContainsPrefix(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("0256")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixLong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("0256723045")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixWrong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	precision := nationale.containsPrefix("01234567")
	if precision != 0 {
		t.Error("Should not contain prefix: ", nationale)
	}
}

func TestDestinationGetExists(t *testing.T) {
	d, err := dataStorage.GetDestination("NAT")
	if err != nil || d == nil {
		t.Error("Could not get destination: ", d)
	}
}

func TestDestinationGetExistsCache(t *testing.T) {
	dataStorage.GetDestination("NAT")
	if _, err := cache2go.GetCached(DESTINATION_PREFIX + "0256"); err != nil {
		t.Error("Destination not cached:", err)
	}
}

func TestDestinationGetNotExists(t *testing.T) {
	d, err := dataStorage.GetDestination("not existing")
	if d != nil {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationGetNotExistsCache(t *testing.T) {
	dataStorage.GetDestination("not existing")
	if d, err := cache2go.GetCached("not existing"); err == nil {
		t.Error("Bad destination cached: ", d)
	}
}

func TestCachedDestHasPrefix(t *testing.T) {
	if !CachedDestHasPrefix("NAT", "0256") {
		t.Error("Could not find prefix in destination")
	}
}

func TestCachedDestHasWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("NAT", "771") {
		t.Error("Prefix should not belong to destination")
	}
}

func TestNonCachedDestRightPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "0256") {
		t.Error("Destination should not belong to prefix")
	}
}

func TestNonCachedDestWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "771") {
		t.Error("Both arguments should be fake")
	}
}

func TestCleanStalePrefixes(t *testing.T) {
	x := struct{}{}
	cache2go.Cache(DESTINATION_PREFIX+"1", map[interface{}]struct{}{"D1": x, "D2": x})
	cache2go.Cache(DESTINATION_PREFIX+"2", map[interface{}]struct{}{"D1": x})
	cache2go.Cache(DESTINATION_PREFIX+"3", map[interface{}]struct{}{"D2": x})
	CleanStalePrefixes([]string{"D1"})
	if r, err := cache2go.GetCached(DESTINATION_PREFIX + "1"); err != nil || len(r.(map[interface{}]struct{})) != 1 {
		t.Error("Error cleaning stale destination ids", r)
	}
	if r, err := cache2go.GetCached(DESTINATION_PREFIX + "2"); err == nil {
		t.Error("Error removing stale prefix: ", r)
	}
	if r, err := cache2go.GetCached(DESTINATION_PREFIX + "3"); err != nil || len(r.(map[interface{}]struct{})) != 1 {
		t.Error("Error performing stale cleaning: ", r)
	}
}

/********************************* Benchmarks **********************************/

func BenchmarkDestinationStorageStoreRestore(b *testing.B) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		dataStorage.SetDestination(nationale)
		dataStorage.GetDestination(nationale.Id)
	}
}
