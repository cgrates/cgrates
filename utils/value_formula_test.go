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
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestValueFormulaDayWeek(t *testing.T) {
	params := make(map[string]any)
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"week", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	if x := incrementalFormula(params); x != 10/7.0 {
		t.Error("error calculating value using formula: ", x)
	}
}

func TestValueFormulaDayMonth(t *testing.T) {
	params := make(map[string]any)
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"month", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	now := time.Now()
	if x := incrementalFormula(params); x != 10/DaysInMonth(now.Year(), now.Month()) {
		t.Error("error calculating value using formula: ", x)
	}
}

func TestValueFormulaDayYear(t *testing.T) {
	params := make(map[string]any)
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"year", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	now := time.Now()
	if x := incrementalFormula(params); x != 10/DaysInYear(now.Year()) {
		t.Error("error calculating value using formula: ", x)
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
	vf := ValueFormula{
		Method: "test",
		Params: map[string]any{"test": "val1"},
		Static: 1.2,
	}

	rcv := vf.String()
	exp := `{"Method":"test","Params":{"test":"val1"},"Static":1.2}`

	if rcv != exp {
		t.Errorf("recived %s, expected %s", rcv, exp)
	}
}

func TestValueFormulaParseBalanceFilterValue2(t *testing.T) {
	vf := ValueFormula{
		Method: "test",
		Params: map[string]any{"test": "val1"},
		Static: 1.2,
	}

	type args struct {
		tor string
		val string
	}

	type exp struct {
		val *ValueFormula
		err error
	}
	tests := []struct {
		name string
		args args
		exp  exp
	}{
		{
			name: "json unmarshal error case",
			args: args{tor: "*voice", val: "test"},
			exp:  exp{val: nil, err: errors.New("Invalid value: " + "test")},
		},
		{
			name: "json unmarshal case",
			args: args{tor: "*voice", val: `{"Method":"test","Params":{"test":"val1"},"Static":1.2}`},
			exp:  exp{val: &vf, err: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv, err := ParseBalanceFilterValue(tt.args.tor, tt.args.val)

			if err != nil {
				if err.Error() != tt.exp.err.Error() {
					t.Fatalf("recived %s, expected %s", err, tt.exp.err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp.val) {
				t.Errorf("recived %s, expected %s", rcv, tt.exp.val)
			}
		})
	}
}

func TestValueFormulaIncremetalFormula(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		arg  map[string]any
		exp  float64
	}{
		{
			name: "keys not found",
			arg:  map[string]any{},
			exp:  0.0,
		},
		{
			name: "units is not a float64",
			arg:  map[string]any{"Units": "test", "Interval": "test", "Increment": "test"},
			exp:  0.0,
		},
		{
			name: "type []byte",
			arg:  map[string]any{"Units": 1.2, "Interval": []byte{1, 2}, "Increment": []byte{1, 2}},
			exp:  0.0,
		},
		{
			name: "default interval",
			arg:  map[string]any{"Units": 1.2, "Interval": 1, "Increment": 1},
			exp:  0.0,
		},
		{
			name: "default increment",
			arg:  map[string]any{"Units": 1.2, "Interval": []byte{1, 2}, "Increment": 1},
			exp:  0.0,
		},
		{
			name: "hour day",
			arg:  map[string]any{"Units": 1.5, "Interval": "day", "Increment": "hour"},
			exp:  1.5 / 24,
		},
		{
			name: "hour month",
			arg:  map[string]any{"Units": 1.5, "Interval": "month", "Increment": "hour"},
			exp:  1.5 / (DaysInMonth(now.Year(), now.Month()) * 24),
		},
		{
			name: "hour year",
			arg:  map[string]any{"Units": 1.5, "Interval": "year", "Increment": "hour"},
			exp:  1.5 / (DaysInYear(now.Year()) * 24),
		},
		{
			name: "minute hour",
			arg:  map[string]any{"Units": 1.5, "Interval": "hour", "Increment": "minute"},
			exp:  1.5 / 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := incrementalFormula(tt.arg)

			if rcv != tt.exp {
				t.Errorf("recived %v, expected %v", rcv, tt.exp)
			}
		})
	}
}
