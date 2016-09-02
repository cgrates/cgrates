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
