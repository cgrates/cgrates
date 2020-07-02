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
package utils

import (
	"strings"
	"testing"
)

func TestDynamicDataProviderFieldAsInterface(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": MapStorage{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": MapStorage{
				"SortedRoutes": []MapStorage{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"BestRoute": 0,
		},
	}
	ddp := NewDynamicDataProvider(dp)
	path := "*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END]"
	out, err := ddp.FieldAsInterface(strings.Split(path, NestingSep))
	if err != nil {
		t.Fatal(err)
	}
	expected := "Identity1"
	if out != expected {
		t.Errorf("Expected: %q,received %q", expected, out)
	}
	path = "*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END"
	_, err = ddp.FieldAsInterface(strings.Split(path, NestingSep))
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}
}

func TestDynamicDataProviderFieldAsInterface2(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": map[string]interface{}{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": map[string]interface{}{
				"SortedRoutes": []map[string]interface{}{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"BestRoute": 0,
		},
	}
	ddp := NewDynamicDataProvider(dp)
	path := "*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END]"
	out, err := ddp.FieldAsInterface(strings.Split(path, NestingSep))
	if err != nil {
		t.Fatal(err)
	}
	expected := "Identity1"
	if out != expected {
		t.Errorf("Expected: %q,received %q", expected, out)
	}
	path = "*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END"
	_, err = ddp.FieldAsInterface(strings.Split(path, NestingSep))
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}
}

func TestDynamicDataProviderProccesFieldPath2(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": MapStorage{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": MapStorage{
				"SortedRoutes": []MapStorage{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"BestRoute": 0,
		},
	}
	ddp := NewDynamicDataProvider(dp)
	newpath, err := ddp.processFieldPathForSet("~*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[1].ID|_END].Something[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END]")
	if err != nil {
		t.Fatal(err)
	}
	expectedPath := "~*cgrep.Stir.CHRG_ROUTE2_END.Something.CHRG_ROUTE1_END"
	if newpath != expectedPath {
		t.Errorf("Expected: %q,received %q", expectedPath, newpath)
	}
	_, err = ddp.processFieldPathForSet("~*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[1].ID|_END].Something[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END")
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = ddp.processFieldPathForSet("~*cgrep.Stir[CHRG_")
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = ddp.processFieldPathForSet("~*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[1].ID2|_END]")
	if err != ErrNotFound {
		t.Errorf("Expected error %s received %v", ErrNotFound, err)
	}
	newpath, err = ddp.processFieldPathForSet("~*cgrep.Stir[1]")
	if err != nil {
		t.Fatal(err)
	}
	if newpath != EmptyString {
		t.Errorf("Expected: %q,received %q", EmptyString, newpath)
	}
}

func TestDynamicDataProviderFieldAsString(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": map[string]interface{}{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": map[string]interface{}{
				"SortedRoutes": []map[string]interface{}{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"BestRoute": 0,
		},
	}
	ddp := NewDynamicDataProvider(dp)
	path := "*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END]"
	out, err := ddp.FieldAsString(strings.Split(path, NestingSep))
	if err != nil {
		t.Fatal(err)
	}
	expected := "Identity1"
	if out != expected {
		t.Errorf("Expected: %q,received %q", expected, out)
	}
	path = "*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END"
	_, err = ddp.FieldAsString(strings.Split(path, NestingSep))
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}
}

func TestDynamicDataProviderGetFullFieldPath(t *testing.T) {
	dp := MapStorage{
		MetaCgrep: MapStorage{
			"Stir": MapStorage{
				"CHRG_ROUTE1_END": "Identity1",
				"CHRG_ROUTE2_END": "Identity2",
				"CHRG_ROUTE3_END": "Identity3",
				"CHRG_ROUTE4_END": "Identity4",
			},
			"Routes": MapStorage{
				"SortedRoutes": []MapStorage{
					{"ID": "ROUTE1"},
					{"ID": "ROUTE2"},
					{"ID": "ROUTE3"},
					{"ID": "ROUTE4"},
				},
			},
			"BestRoute": 0,
		},
	}
	ddp := NewDynamicDataProvider(dp)
	newpath, err := ddp.GetFullFieldPath("~*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[1].ID|_END].Something[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END]")
	if err != nil {
		t.Fatal(err)
	}
	expectedPath := "~*cgrep.Stir.CHRG_ROUTE2_END.Something.CHRG_ROUTE1_END"
	if newpath.Path != expectedPath {
		t.Errorf("Expected: %q,received %q", expectedPath, newpath)
	}
	_, err = ddp.GetFullFieldPath("~*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[1].ID|_END].Something[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END")
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = ddp.GetFullFieldPath("~*cgrep.Stir[CHRG_")
	if err != ErrWrongPath {
		t.Errorf("Expected error %s received %v", ErrWrongPath, err)
	}

	_, err = ddp.GetFullFieldPath("~*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[1].ID2|_END]")
	if err != ErrNotFound {
		t.Errorf("Expected error %s received %v", ErrNotFound, err)
	}
	newpath, err = ddp.GetFullFieldPath("~*cgrep.Stir[1]")
	if err != nil {
		t.Fatal(err)
	}
	if newpath != nil {
		t.Errorf("Expected: %v,received %q", nil, newpath)
	}

	newpath, err = ddp.GetFullFieldPath("~*cgrep.Stir")
	if err != nil {
		t.Fatal(err)
	}
	if newpath != nil {
		t.Errorf("Expected: %v,received %q", nil, newpath)
	}

	newpath, err = ddp.GetFullFieldPath("*cgrep.Stir[CHRG_|~*cgrep.Routes.SortedRoutes[~*cgrep.BestRoute].ID|_END]")
	if err != nil {
		t.Fatal(err)
	}
	if newpath == nil {
		t.Errorf("Expected: %v,received %q", nil, newpath)
	}
}
