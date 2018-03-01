/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
	"testing"

	"github.com/cgrates/cgrates/utils"
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
	nationale := &Destination{Id: "nat",
		Prefixes: []string{"0257", "0256", "0723"}}
	err := dm.DataDB().SetDestination(nationale, utils.NonTransactional)
	if err != nil {
		t.Error("Error storing destination: ", err)
	}
	result, err := dm.DataDB().GetDestination(nationale.Id,
		false, utils.NonTransactional)
	if nationale.containsPrefix("0257") == 0 ||
		nationale.containsPrefix("0256") == 0 ||
		nationale.containsPrefix("0723") == 0 {
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
	d, err := dm.DataDB().GetDestination("NAT", false, utils.NonTransactional)
	if err != nil || d == nil {
		t.Error("Could not get destination: ", d)
	}
}

func TestDestinationReverseGetExistsCache(t *testing.T) {
	dm.DataDB().GetReverseDestination("0256", false, utils.NonTransactional)
	if _, ok := Cache.Get(utils.CacheReverseDestinations, "0256"); !ok {
		t.Error("Destination not cached:", err)
	}
}

func TestDestinationGetNotExists(t *testing.T) {
	if d, ok := Cache.Get(utils.CacheDestinations, "not existing"); ok {
		t.Error("Bad destination cached: ", d)
	}
	d, err := dm.DataDB().GetDestination("not existing", false, utils.NonTransactional)
	if d != nil {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationCachedDestHasPrefix(t *testing.T) {
	if !CachedDestHasPrefix("NAT", "0256") {
		t.Error("Could not find prefix in destination")
	}
}

func TestDestinationCachedDestHasWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("NAT", "771") {
		t.Error("Prefix should not belong to destination")
	}
}

func TestDestinationNonCachedDestRightPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "0256") {
		t.Error("Destination should not belong to prefix")
	}
}

func TestDestinationNonCachedDestWrongPrefix(t *testing.T) {
	if CachedDestHasPrefix("FAKE", "771") {
		t.Error("Both arguments should be fake")
	}
}

/*
func TestCleanStalePrefixes(t *testing.T) {
	x := struct{}{}
	cache.Set(utils.DESTINATION_PREFIX+"1", map[string]struct{}{"D1": x, "D2": x})
	cache.Set(utils.DESTINATION_PREFIX+"2", map[string]struct{}{"D1": x})
	cache.Set(utils.DESTINATION_PREFIX+"3", map[string]struct{}{"D2": x})
	CleanStalePrefixes([]string{"D1"})
	if r, ok := cache.Get(utils.DESTINATION_PREFIX + "1"); !ok || len(r.(map[string]struct{})) != 1 {
		t.Error("Error cleaning stale destination ids", r)
	}
	if r, ok := cache.Get(utils.DESTINATION_PREFIX + "2"); ok {
		t.Error("Error removing stale prefix: ", r)
	}
	if r, ok := cache.Get(utils.DESTINATION_PREFIX + "3"); !ok || len(r.(map[string]struct{})) != 1 {
		t.Error("Error performing stale cleaning: ", r)
	}
}*/

/********************************* Benchmarks **********************************/

func BenchmarkDestinationStorageStoreRestore(b *testing.B) {
	nationale := &Destination{Id: "nat", Prefixes: []string{"0257", "0256", "0723"}}
	for i := 0; i < b.N; i++ {
		dm.DataDB().SetDestination(nationale, utils.NonTransactional)
		dm.DataDB().GetDestination(nationale.Id, true, utils.NonTransactional)
	}
}
