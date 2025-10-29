/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"reflect"
	"testing"
	"time"

	"github.com/ericlagergren/decimal"
)

func TestOptionsGetFloat64Opts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetFloat64Opts(ev, 1.2, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 1.2 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 1.2, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 0.11
	if rcv, err := GetFloat64Opts(ev, 1.2, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 0.11 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 0.11, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseFloat: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetFloat64Opts(ev, 1.2, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetDurationOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", time.Minute, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 2 * time.Second
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 2*time.Second {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Second, rcv)
	}
	ev.APIOpts["optionName"] = 600
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 600 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 600, rcv)
	}
	ev.APIOpts["optionName"] = "2m0s"
	if rcv, err := GetDurationOpts(ev, time.Minute, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 2*time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Minute, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `time: invalid duration "invalid"`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetDurationOpts(ev, time.Minute, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetStringOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "default" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "default", rcv)
	}

	// option key populated, get its value
	ev.APIOpts["optionName"] = "optionValue"
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "optionValue" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "optionValue", rcv)
	}
	ev.APIOpts["optionName"] = false
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "false" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "false", rcv)
	}
	ev.APIOpts["optionName"] = 5 * time.Minute
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "5m0s" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "5m0s", rcv)
	}
	ev.APIOpts["optionName"] = 12.34
	if rcv := GetStringOpts(ev, "default", "optionName"); rcv != "12.34" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "12.34", rcv)
	}
}

func TestOptionsGetStringSliceOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	defaultValue := []string{"default"}
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, defaultValue) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	optValue := []string{"optValue"}
	ev.APIOpts["optionName"] = optValue
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, optValue) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue, rcv)
	}
	optValue2 := []string{"true", "false"}
	ev.APIOpts["optionName"] = []bool{true, false}
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, optValue2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue2, rcv)
	}
	optValue3 := []string{"12.34", "3"}
	ev.APIOpts["optionName"] = []float64{12.34, 3}
	if rcv, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, optValue3) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue3, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `cannot convert field: string to []string`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetStringSliceOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetIntOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetIntOpts(ev, 5, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 5 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 5, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 12
	if rcv, err := GetIntOpts(ev, 5, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}
	ev.APIOpts["optionName"] = 12.7
	if rcv, err := GetIntOpts(ev, 5, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseInt: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetIntOpts(ev, 5, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetBoolOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != false {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", false, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = true
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}
	ev.APIOpts["optionName"] = 5
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}
	ev.APIOpts["optionName"] = "true"
	if rcv, err := GetBoolOpts(ev, false, "optionName"); err != nil {
		t.Error(err)
	} else if rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseBool: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetBoolOpts(ev, false, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetInterfaceOpts(t *testing.T) {

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != "default" {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", "default", rcv)
	}

	// option key populated, get its value
	ev.APIOpts["optionName"] = 0.11
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != 0.11 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 0.11, rcv)
	}
	ev.APIOpts["optionName"] = true
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != true {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", true, rcv)
	}
	ev.APIOpts["optionName"] = 5 * time.Minute
	if rcv := GetInterfaceOpts(ev, "default", "optionName"); rcv != 5*time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 5*time.Minute, rcv)
	}
}

func TestOptionsGetIntPointerOpts(t *testing.T) {
	defaultValue := IntPointer(5)

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || rcv != defaultValue {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 12
	if rcv, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}
	ev.APIOpts["optionName"] = 12.7
	if rcv, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 12 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 12, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `strconv.ParseInt: parsing "invalid": invalid syntax`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetIntPointerOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetDurationPointerOpts(t *testing.T) {
	defaultValue := DurationPointer(time.Minute)
	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || rcv != defaultValue {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	ev.APIOpts["optionName"] = 2 * time.Second
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 2*time.Second {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Second, rcv)
	}
	ev.APIOpts["optionName"] = 600
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 600 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 600, rcv)
	}
	ev.APIOpts["optionName"] = "2m0s"
	if rcv, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv == nil || *rcv != 2*time.Minute {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2*time.Minute, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `time: invalid duration "invalid"`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetDurationPointerOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}

func TestOptionsGetDecimalBigOpts(t *testing.T) {
	defaultValue := decimal.New(1234, 3)

	// option key not populated, retrieve default value
	ev := &CGREvent{
		APIOpts: make(map[string]any),
	}
	if rcv, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv.Cmp(defaultValue) != 0 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", defaultValue, rcv)
	}

	// option key populated with valid input, get its value
	optValue := decimal.New(15, 1)
	ev.APIOpts["optionName"] = 1.5
	if rcv, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv.Cmp(optValue) != 0 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue, rcv)
	}
	ev.APIOpts["optionName"] = "1.5"
	if rcv, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err != nil {
		t.Error(err)
	} else if rcv.Cmp(optValue) != 0 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", optValue, rcv)
	}

	// option key populated with invalid input, receive error
	expectedErr := `can't convert <invalid> to decimal`
	ev.APIOpts["optionName"] = "invalid"
	if _, err := GetDecimalBigOpts(ev, defaultValue, "optionName"); err == nil ||
		err.Error() != expectedErr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedErr, err)
	}
}
