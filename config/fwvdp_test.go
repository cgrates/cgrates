// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewFWVProvider(t *testing.T) {
	record := `"cgrates.org", "ATTR_1", "*sessions;*cdrs", "*string:~*req.Account:1007", "2014-01-14T00:00:00Z", "*req.Account", "*constant", "1001", "false", "10"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	if value := NewFWVProvider(record); !reflect.DeepEqual(value, dp) {
		t.Errorf("Expected %+v, received %+v", dp, value)
	}
}

func TestStringReqFWV(t *testing.T) {
	record := `"cgrates.org", "ATTR_1", "*sessions;*cdrs", "*string:~*req.Account:1007", "2014-01-14T00:00:00Z", "*req.Account", "*constant", "1001", "false", "10"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := `"\"cgrates.org\", \"ATTR_1\", \"*sessions;*cdrs\", \"*string:~*req.Account:1007\", \"2014-01-14T00:00:00Z\", \"*req.Account\", \"*constant\", \"1001\", \"false\", \"10\""`
	if received := dp.String(); !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestFieldAsInterfaceFWV(t *testing.T) {
	pth := []string{"1-12"}
	record := `"cgrates.org", "ATTR_1"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "cgrates.org"
	if received, err := dp.FieldAsInterface(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}

	if received, err := dp.FieldAsInterface(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestFieldAsInterfaceFWVEmptyPath(t *testing.T) {
	dp := new(FWVProvider)
	var expected any
	if received, err := dp.FieldAsInterface([]string{}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestFieldAsInterfaceFWVInvalidPath(t *testing.T) {
	pth := []string{"112"}
	record := `"cgrates.org", "ATTR_1"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "Invalid format for index : [112] "
	if _, err := dp.FieldAsInterface(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceFWVInvalidConvert1(t *testing.T) {
	pth := []string{"1s-12"}
	record := `"cgrates.org", "ATTR_1"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "strconv.Atoi: parsing \"1s\": invalid syntax"
	if _, err := dp.FieldAsInterface(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceFWVInvalidConvert2(t *testing.T) {
	pth := []string{"1-1s"}
	record := `"cgrates.org", "ATTR_1"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "strconv.Atoi: parsing \"1s\": invalid syntax"
	if _, err := dp.FieldAsInterface(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceFWVLowerLength(t *testing.T) {
	pth := []string{"10-14"}
	record := `"cgra"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "StartIndex : 10 is greater than : 6"
	if _, err := dp.FieldAsInterface(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsInterfaceFWVHigherLength(t *testing.T) {
	pth := []string{"3-14"}
	record := `"cgra"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "FinalIndex : 14 is greater than : 6"
	if _, err := dp.FieldAsInterface(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestFieldAsStringFWV(t *testing.T) {
	pth := []string{"1-12"}
	record := `"cgrates.org", "ATTR_1"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "cgrates.org"
	if received, err := dp.FieldAsString(pth); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expected %+v, received %+v", expected, received)
	}
}

func TestFieldAsStringFWVError(t *testing.T) {
	pth := []string{"112"}
	record := `"cgrates.org", "ATTR_1"`
	dp := &FWVProvider{
		req:   record,
		cache: utils.MapStorage{},
	}
	expected := "Invalid format for index : [112] "
	if _, err := dp.FieldAsString(pth); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
