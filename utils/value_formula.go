// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"
)

// for computing a dynamic value for Value field
type ValueFormula struct {
	Method string
	Params map[string]any
	Static float64
}

func ParseBalanceFilterValue(tor string, val string) (*ValueFormula, error) {
	if tor == VOICE { // VOICE balance is parsed as nanoseconds with support for time duration strings
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
		log.Print("units")
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
