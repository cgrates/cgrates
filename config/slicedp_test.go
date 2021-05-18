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

package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewSliceDp(t *testing.T) {
	record := []string{"cgrates.org", "ATTR_1", "*string:~*opts.*context:*sessions|*cdrs;*string:~*req.Account:1007", "10", "*req.Account", "*constant", "1001", "false"}
	index := map[string]int{
		"Tenant":             1,
		"ID":                 2,
		"FilterIDs":          3,
		"Weight":             4,
		"AttributeFilterIDs": 5,
		"Path":               6,
		"Type":               7,
		"Value":              8,
		"Blocker":            9,
	}
	expected := &SliceDP{
		req:    record,
		cache:  utils.MapStorage{utils.Length: len(record)},
		idxAls: index,
	}
	if newSliceDP := NewSliceDP(record, index); !reflect.DeepEqual(expected, newSliceDP) {
		t.Errorf("Expected %+v, received %+v", expected, newSliceDP)
	}
}

func TestGetIndexValue(t *testing.T) {
	index := map[string]int{
		"Tenant":             1,
		"ID":                 2,
		"FilterIDs":          3,
		"Weight":             4,
		"AttributeFilterIDs": 5,
		"Path":               6,
		"Type":               7,
		"Value":              8,
		"Blocker":            9,
	}
	sliceDp := SliceDP{
		idxAls: index,
	}
	expected := 5
	if idx, err := sliceDp.getIndex("5"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, idx) {
		t.Errorf("Expected %+v, received %+v", expected, idx)
	}
}

func TestGetIndexValueKey(t *testing.T) {
	index := map[string]int{
		"Tenant":             1,
		"ID":                 2,
		"FilterIDs":          3,
		"Weight":             4,
		"AttributeFilterIDs": 5,
		"Path":               6,
		"Type":               7,
		"Value":              8,
		"Blocker":            9,
	}
	sliceDp := SliceDP{
		idxAls: index,
	}
	for key, value := range index {
		expected := value
		if idx, err := sliceDp.getIndex(fmt.Sprintf("%v", key)); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(expected, idx) {
			t.Errorf("Expected %+v, received %+v", expected, idx)
		}
	}
}

func TestRemoteHostSliceDP(t *testing.T) {
	expected := utils.LocalAddr()
	sliceDP := new(SliceDP)
	if received := sliceDP.RemoteHost(); !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestStringReqSliceDP(t *testing.T) {
	record := []string{"cgrates.org", "ATTR_1", "*sessions;*cdrs", "*string:~*req.Account:1007", "2014-01-14T00:00:00Z", "*req.Account", "*constant", "1001", "false", "10"}
	sliceDP := &SliceDP{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := `["cgrates.org","ATTR_1","*sessions;*cdrs","*string:~*req.Account:1007","2014-01-14T00:00:00Z","*req.Account","*constant","1001","false","10"]`
	if received := sliceDP.String(); !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestFieldAsInterfaceSliceDP(t *testing.T) {
	pth := []string{"Tenant"}
	slicedp := SliceDP{
		req:   []string{"cgrates.org"},
		cache: utils.MapStorage{},
		idxAls: map[string]int{
			"Tenant": 0,
		},
	}
	expected := slicedp.req[0]
	if value, err := slicedp.FieldAsInterface(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(value, expected) {
		t.Errorf("Expected %+v, received %+v", expected, value)
	}
	if value, err := slicedp.FieldAsInterface(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(value, expected) {
		t.Errorf("Expected %+v, received %+v", expected, value)
	}
}

func TestFieldAsInterfaceMultiplePaths(t *testing.T) {
	pth := []string{"Tenant", "ID"}
	sliceDp := new(SliceDP)
	expected := "Invalid fieldPath [Tenant ID] "
	if _, err := sliceDp.FieldAsInterface(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceEmptyPath(t *testing.T) {
	sliceDp := new(SliceDP)
	var expected interface{}
	if value, err := sliceDp.FieldAsInterface([]string{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, value) {
		t.Errorf("Expected %+v, received %+v", expected, value)
	}
}

func TestFieldAsInterfaceIgnoringError(t *testing.T) {
	pth := []string{"Tenant"}
	slicedp := SliceDP{
		req:   []string{"cgrates.org"},
		cache: utils.MapStorage{},
		idxAls: map[string]int{
			"NotFound": 0,
		},
	}
	expected := "Ignoring record: [cgrates.org] with error : strconv.Atoi: parsing \"Tenant\": invalid syntax "
	if _, err := slicedp.FieldAsInterface(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceErrNotFound(t *testing.T) {
	pth := []string{"Tenant"}
	slicedp := SliceDP{
		req:   []string{"cgrates.org"},
		cache: utils.MapStorage{},
		idxAls: map[string]int{
			"Tenant": 2,
		},
	}
	if _, err := slicedp.FieldAsInterface(pth); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func TestFieldAsString(t *testing.T) {
	pth := []string{"Tenant"}
	slicedp := SliceDP{
		req:   []string{"cgrates.org"},
		cache: utils.MapStorage{},
		idxAls: map[string]int{
			"Tenant": 0,
		},
	}
	expected := "cgrates.org"
	if value, err := slicedp.FieldAsString(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(value, expected) {
		t.Errorf("Expected %+v, received %+v", expected, value)
	}
}

func TestFieldAsStringErr(t *testing.T) {
	pth := []string{"Tenant"}
	slicedp := SliceDP{
		req:   []string{"cgrates.org"},
		cache: utils.MapStorage{},
		idxAls: map[string]int{
			"Tenant": 1,
		},
	}
	if _, err := slicedp.FieldAsString(pth); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}
