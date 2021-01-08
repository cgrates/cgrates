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
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestValueFormulaDayWeek(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "week",
		"Increment": "day",
	}
	if x := incrementalFormula(params); x != 10/7.0 {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaDayMonth(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "month",
		"Increment": "day",
	}
	now := time.Now()
	if x := incrementalFormula(params); x != 10/DaysInMonth(now.Year(), now.Month()) {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaDayYear(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "year",
		"Increment": "day",
	}

	now := time.Now()
	if x := incrementalFormula(params); x != 10/DaysInYear(now.Year()) {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaParseBalanceFilterValue(t *testing.T) {
	eVF := &ValueFormula{Static: 10000000000.0}
	if vf, err := ParseBalanceFilterValue(MetaVoice, "10s"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVF, vf) {
		t.Errorf("Expecting: %+v, received: %+v", eVF, vf)
	}
	eVF = &ValueFormula{Static: 1024.0}
	if vf, err := ParseBalanceFilterValue(MetaData, "1024"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVF, vf) {
		t.Errorf("Expecting: %+v, received: %+v", eVF, vf)
	}
	eVF = &ValueFormula{Static: 10.0}
	if vf, err := ParseBalanceFilterValue(MetaMonetary, "10"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVF, vf) {
		t.Errorf("Expecting: %+v, received: %+v", eVF, vf)
	}
}

func TestValueFormulaString(t *testing.T) {
	vf := &ValueFormula{}
	eOut := `{"Method":"","Params":null,"Static":0}`
	if rcv := vf.String(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	vf = &ValueFormula{Static: 10000000000.0}
	eOut = `{"Method":"","Params":null,"Static":10000000000}`
	if rcv := vf.String(); !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestParseBalanceFilterValue(t *testing.T) {
	d, err := ParseDurationWithNanosecs("18")
	if err != nil {
		t.Error("error parsing time: ", err)
	}
	eOut := &ValueFormula{
		Static: float64(d.Nanoseconds()),
	}
	if rcv, err := ParseBalanceFilterValue(MetaVoice, "18"); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	eOut = &ValueFormula{
		Static: float64(19),
	}
	if rcv, err := ParseBalanceFilterValue(EmptyString, "19"); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	eOut = &ValueFormula{}
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"week", "Increment":"day"}`), &eOut); err != nil {
		t.Error("error unmarshalling params: ", err)
	}

	if rcv, err := ParseBalanceFilterValue(EmptyString, `{"Units":10, "Interval":"week", "Increment":"day"}`); err != nil {
		t.Errorf("Expecting: nil, received: %+v", err)
	} else if !reflect.DeepEqual(rcv, eOut) {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}

	if rcv, err := ParseBalanceFilterValue(EmptyString, `not really a json`); rcv != nil {
		t.Errorf("Expecting: nil, received: %+v", rcv)
	} else if err.Error() != "Invalid value: not really a json" {
		t.Errorf("Expecting: %+v, received: %+v", eOut, rcv)
	}
}

func TestValueFormulaEmptyFields(t *testing.T) {
	params := map[string]interface{}{}
	expected := 0.0
	received := incrementalFormula(params)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaConvertFloat64(t *testing.T) {
	params := map[string]interface{}{
		"Units":     50,
		"Interval":  "day",
		"Increment": "hour",
	}
	expected := 0.0
	received := incrementalFormula(params)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIntervalByte(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  []byte("week"),
		"Increment": "day",
	}

	expected := 10.0 / 7.0
	received := incrementalFormula(params)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIntervalDefault(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  5,
		"Increment": "day",
	}

	expected := 0.0
	received := incrementalFormula(params)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIncrementDefault(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "week",
		"Increment": 5,
	}

	expected := 0.0
	received := incrementalFormula(params)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIncrementByte(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "week",
		"Increment": []byte("day"),
	}

	expected := 10.0 / 7.0
	received := incrementalFormula(params)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIncrementHourDay(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "day",
		"Increment": "hour",
	}

	expected := 10.0 / 24.0
	received := incrementalFormula(params)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIncrementHourMonth(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "month",
		"Increment": "hour",
	}
	now := time.Now()
	expected := 10.0 / (DaysInMonth(now.Year(), now.Month()) * 24)
	received := incrementalFormula(params)

	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIncrementHourYear(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "year",
		"Increment": "hour",
	}
	now := time.Now()
	expected := 10.0 / (DaysInYear(now.Year()) * 24)
	received := incrementalFormula(params)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaIncrementMinute(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "hour",
		"Increment": "minute",
	}

	expected := 10.0 / 60
	received := incrementalFormula(params)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}

func TestValueFormulaCover(t *testing.T) {
	params := map[string]interface{}{
		"Units":     10.0,
		"Interval":  "cat",
		"Increment": "cat",
	}

	expected := 0.0
	received := incrementalFormula(params)
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Expecting: %+v, received: %+v", expected, received)
	}
}
