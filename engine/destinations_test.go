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
	nationale := &Destination{Id: "nat", Prefixes: map[string]interface{}{"0257": nil, "0256": nil, "0723": nil}}
	s, _ := json.Marshal(nationale)
	d1 := &Destination{Id: "nat"}
	json.Unmarshal(s, d1)
	s1, _ := json.Marshal(d1)
	if string(s1) != string(s) {
		t.Errorf("Expected %q was %q", s, s1)
	}
}

func TestDestinationStorageStore(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: map[string]interface{}{"0257": nil, "0256": nil, "0723": nil}}
	err := storageGetter.SetDestination(nationale)
	if err != nil {
		t.Error("Error storing destination: ", err)
	}
	result, err := storageGetter.GetDestination(nationale.Id, false)
	_, a := nationale.Prefixes["0257"]
	_, b := nationale.Prefixes["0256"]
	_, c := nationale.Prefixes["0723"]
	if !a || !b || !c {
		t.Errorf("Expected %q was %q", nationale, result)
	}
}

func TestDestinationContainsPrefix(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: map[string]interface{}{"0257": nil, "0256": nil, "0723": nil}}
	precision := nationale.containsPrefix("0256")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixLong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: map[string]interface{}{"0257": nil, "0256": nil, "0723": nil}}
	precision := nationale.containsPrefix("0256723045")
	if precision != len("0256") {
		t.Error("Should contain prefix: ", nationale)
	}
}

func TestDestinationContainsPrefixWrong(t *testing.T) {
	nationale := &Destination{Id: "nat", Prefixes: map[string]interface{}{"0257": nil, "0256": nil, "0723": nil}}
	precision := nationale.containsPrefix("01234567")
	if precision != 0 {
		t.Error("Should not contain prefix: ", nationale)
	}
}

func TestDestinationGetExists(t *testing.T) {
	d, err := storageGetter.GetDestination("NAT", false)
	if err != nil || d == nil {
		t.Error("Could not get destination: ", d)
	}
}

func TestDestinationGetExistsCache(t *testing.T) {
	storageGetter.GetDestination("NAT", false)
	if _, err := cache2go.GetCached(DESTINATION_PREFIX + "NAT"); err != nil {
		t.Error("Destination not cached:", err)
	}
}

func TestDestinationGetNotExists(t *testing.T) {
	d, err := storageGetter.GetDestination("not existing", false)
	if d != nil {
		t.Error("Got false destination: ", d, err)
	}
}

func TestDestinationGetNotExistsCache(t *testing.T) {
	storageGetter.GetDestination("not existing", false)
	if d, err := cache2go.GetCached("not existing"); err == nil {
		t.Error("Bad destination cached: ", d)
	}
}

/*
func TestConcurrentDestReadWrite(t *testing.T) {
	dst1 := &Destination{Id: "TST_1", Prefixes: []string{"1"}}
	err := storageGetter.SetDestination(dst1)
	if err != nil {
		t.Error("Error setting  destination: ", err)
	}
	rec := 500
	go func() {
		for i := 0; i < rec; i++ {
			storageGetter.SetDestination(&Destination{Id: fmt.Sprintf("TST_%d", i), Prefixes: []string{"1"}})
		}
	}()

	for i := 0; i < rec; i++ {
		dst2, err := storageGetter.GetDestination(dst1.Id)
		if err != nil {
			t.Error("Error retrieving destination: ", err)
		}
		if !reflect.DeepEqual(dst1, dst2) {
			t.Error("Cannot retrieve properly the destination 1", dst1, dst2)
		}
	}
}

func TestNonConcurrentDestReadWrite(t *testing.T) {
	dst1 := &Destination{Id: "TST_1", Prefixes: []string{"1"}}
	err := storageGetter.SetDestination(dst1)
	if err != nil {
		t.Error("Error setting destination: ", err)
	}
	rec := 10000
	//go func(){
	for i := 0; i < rec; i++ {
		storageGetter.SetDestination(&Destination{Id: fmt.Sprintf("TST_%d", i), Prefixes: []string{"1"}})
	}
	//}()

	for i := 0; i < rec; i++ {
		dst2, err := storageGetter.GetDestination(dst1.Id)
		if err != nil {
			t.Error("Error retrieving destination: ", err)
		}
		if !reflect.DeepEqual(dst1, dst2) {
			t.Error("Cannot retrieve properly the destination 1", dst1, dst2)
		}
	}
}
*/
/********************************* Benchmarks **********************************/

func BenchmarkDestinationStorageStoreRestore(b *testing.B) {
	nationale := &Destination{Id: "nat", Prefixes: map[string]interface{}{"0257": nil, "0256": nil, "0723": nil}}
	for i := 0; i < b.N; i++ {
		storageGetter.SetDestination(nationale)
		storageGetter.GetDestination(nationale.Id, true)
	}
}
