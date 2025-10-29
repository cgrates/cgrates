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
	"fmt"
	"strconv"
	"time"
)

// ValueFormula for computing a dynamic value for Value field
type ValueFormula struct {
	Method string
	Params map[string]any
	Static float64
}

func ParseBalanceFilterValue(tor string, val string) (*ValueFormula, error) {
	if tor == MetaVoice { // Voice balance is parsed as nanoseconds with support for time duration strings
		if d, err := ParseDurationWithNanosecs(val); err == nil {
			return &ValueFormula{Static: float64(d.Nanoseconds())}, err
		}
	} else if u, err := strconv.ParseFloat(val, 64); err == nil {
		return &ValueFormula{Static: u}, err
	}
	var vf ValueFormula
	if err := json.Unmarshal([]byte(val), &vf); err == nil {
		return &vf, err
	}
	return nil, errors.New("Invalid value: " + val)
}

type valueFormula func(map[string]any) float64

const (
	INCREMENTAL = "*incremental"
)

var ValueFormulas = map[string]valueFormula{
	INCREMENTAL: incrementalFormula,
}

func (vf *ValueFormula) String() string {
	return ToJSON(vf)
}

func incrementalFormula(params map[string]any) float64 {
	// check parameters
	unitsInterface, unitsFound := params["Units"]
	intervalInterface, intervalFound := params["Interval"]
	incrementInterface, incrementFound := params["Increment"]

	if !unitsFound || !intervalFound || !incrementFound {
		return 0.0
	}
	units, ok := unitsInterface.(float64)
	if !ok {
		return 0.0
	}
	var interval string
	switch intr := intervalInterface.(type) {
	case string:
		interval = intr
	case []byte:
		interval = string(intr)
	default:
		return 0.0
	}
	var increment string
	switch incr := incrementInterface.(type) {
	case string:
		increment = incr
	case []byte:
		increment = string(incr)
	default:
		return 0.0
	}
	now := time.Now()
	if increment == "day" {
		if interval == "week" {
			return units / 7
		}
		if interval == "month" {
			return units / DaysInMonth(now.Year(), now.Month())
		}
		if interval == "year" {
			return units / DaysInYear(now.Year())
		}
	}
	if increment == "hour" {
		if interval == "day" {
			return units / 24
		}
		if interval == "month" {
			return units / (DaysInMonth(now.Year(), now.Month()) * 24)
		}
		if interval == "year" {
			return units / (DaysInYear(now.Year()) * 24)
		}
	}
	if increment == "minute" {
		if interval == "hour" {
			return units / 60
		}
	}
	return 0.0
}

func (vf *ValueFormula) FieldAsInterface(fldPath []string) (val any, err error) {
	if vf == nil || len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		opath, indx := GetPathIndexString(fldPath[0])
		if indx != nil && opath == Params {
			return MapStorage(vf.Params).FieldAsInterface(append([]string{*indx}, fldPath[1:]...))
		}
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case Method:
		if len(fldPath) != 1 {
			return nil, ErrNotFound
		}
		return vf.Method, nil
	case Static:
		if len(fldPath) != 1 {
			return nil, ErrNotFound
		}
		return vf.Static, nil
	case Params:
		if len(fldPath) == 1 {
			return vf.Params, nil
		}
		return MapStorage(vf.Params).FieldAsInterface(fldPath[1:])
	}
}

func (vf *ValueFormula) FieldAsString(fldPath []string) (val string, err error) {
	var iface any
	iface, err = vf.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return IfaceAsString(iface), nil
}
