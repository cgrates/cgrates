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
	params := make(map[string]interface{})
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"week", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	if x := incrementalFormula(params); x != 10/7.0 {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaDayMonth(t *testing.T) {
	params := make(map[string]interface{})
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"month", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	now := time.Now()
	if x := incrementalFormula(params); x != 10/DaysInMonth(now.Year(), now.Month()) {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaDayYear(t *testing.T) {
	params := make(map[string]interface{})
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"year", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	now := time.Now()
	if x := incrementalFormula(params); x != 10/DaysInYear(now.Year()) {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaParseBalanceFilterValue(t *testing.T) {
	eVF := &ValueFormula{Static: 10000000000.0}
	if vf, err := ParseBalanceFilterValue(VOICE, "10s"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVF, vf) {
		t.Errorf("Expecting: %+v, received: %+v", eVF, vf)
	}
	eVF = &ValueFormula{Static: 1024.0}
	if vf, err := ParseBalanceFilterValue(DATA, "1024"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eVF, vf) {
		t.Errorf("Expecting: %+v, received: %+v", eVF, vf)
	}
	eVF = &ValueFormula{Static: 10.0}
	if vf, err := ParseBalanceFilterValue(MONETARY, "10"); err != nil {
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
	if rcv, err := ParseBalanceFilterValue(VOICE, "18"); err != nil {
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
