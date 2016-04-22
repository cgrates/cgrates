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
	if x := periodicFormula(params); x != 10/7.0 {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaDayMonth(t *testing.T) {
	params := make(map[string]interface{})
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"month", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	now := time.Now()
	if x := periodicFormula(params); x != 10/DaysInMonth(now.Year(), now.Month()) {
		t.Error("error caclulating value using formula: ", x)
	}
}

func TestValueFormulaDayYear(t *testing.T) {
	params := make(map[string]interface{})
	if err := json.Unmarshal([]byte(`{"Units":10, "Interval":"year", "Increment":"day"}`), &params); err != nil {
		t.Error("error unmarshalling params: ", err)
	}
	now := time.Now()
	if x := periodicFormula(params); x != 10/DaysInYear(now.Year()) {
		t.Error("error caclulating value using formula: ", x)
	}
}
