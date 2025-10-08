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
)

func TestTableName(t *testing.T) {
	cdrSQLTable := CDRSQLTable{}

	tableName := cdrSQLTable.TableName()

	if tableName != CDRsTBL {
		t.Errorf("TableName() = %s, expected %s", tableName, CDRsTBL)
	}
}

func TestJSONB_GormDataType(t *testing.T) {
	j := JSONB{"key": "value"}

	dataType := j.GormDataType()

	if dataType != "JSONB" {
		t.Errorf("GormDataType() returned %v, expected 'JSONB'", dataType)
	}
}

func TestJSONBScan(t *testing.T) {
	{
		j := &JSONB{}
		value := []byte(`{"key": "value"}`)

		err := j.Scan(value)

		if err != nil {
			t.Errorf("Scan([]byte) returned error: %v, expected no error", err)
		}

		expected := JSONB{"key": "value"}
		if !reflect.DeepEqual(*j, expected) {
			t.Errorf("Scan([]byte) resulted in %v, expected %v", *j, expected)
		}
	}

	{
		j := &JSONB{}
		value := `{"key": "value"}`

		err := j.Scan(value)

		if err != nil {
			t.Errorf("Scan(string) returned error: %v, expected no error", err)
		}

		expected := JSONB{"key": "value"}
		if !reflect.DeepEqual(*j, expected) {
			t.Errorf("Scan(string) resulted in %v, expected %v", *j, expected)
		}
	}

	{
		j := &JSONB{}
		value := 123

		err := j.Scan(value)

		if err == nil {
			t.Errorf("Scan(int) did not return an error, expected an error")
		} else {
			expectedError := errors.New("Failed to unmarshal JSONB value:123") // Removed space after ":"
			if err.Error() != expectedError.Error() {
				t.Errorf("Scan(int) returned error: %v, expected %v", err, expectedError)
			}
		}
	}
}

func TestJSONBValue(t *testing.T) {
	{
		j := JSONB{"key": "value"}

		value, err := j.Value()

		if err != nil {
			t.Errorf("Value() returned error: %v, expected no error", err)
		}

		expected, _ := json.Marshal(j)
		if string(value.([]byte)) != string(expected) {
			t.Errorf("Value() resulted in %s, expected %s", value, expected)
		}
	}

	{
		j := JSONB{}

		value, err := j.Value()

		if err != nil {
			t.Errorf("Value() returned error: %v, expected no error", err)
		}

		expected, _ := json.Marshal(j)
		if string(value.([]byte)) != string(expected) {
			t.Errorf("Value() resulted in %s, expected %s", value, expected)
		}
	}
}

func TestGetUniqueCDRID(t *testing.T) {
	{
		cgrEv := &CGREvent{
			APIOpts: map[string]interface{}{
				MetaChargeID: "charge_id_123",
			},
		}

		result := GetUniqueCDRID(cgrEv)

		expected := "charge_id_123"
		if result != expected {
			t.Errorf("GetUniqueCDRID() returned %s, expected %s", result, expected)
		}
	}

	{
		cgrEv := &CGREvent{
			APIOpts: map[string]interface{}{
				MetaOriginID: "origin_id_456",
			},
		}

		result := GetUniqueCDRID(cgrEv)

		expected := "origin_id_456"
		if result != expected {
			t.Errorf("GetUniqueCDRID() returned %s, expected %s", result, expected)
		}
	}

	{
		cgrEv := &CGREvent{
			APIOpts: map[string]interface{}{},
		}

		result := GetUniqueCDRID(cgrEv)

		if len(result) != 7 {
			t.Errorf("GetUniqueCDRID() returned %s, expected a 7-character UUID prefix", result)
		}
	}
}

func TestCDR_CGREvent(t *testing.T) {
	{
		cdr := &CDR{
			Tenant: "test_tenant",
			Event:  map[string]interface{}{"key": "value"},
			Opts:   map[string]interface{}{"opt_key": "opt_value"},
		}

		cgrEvent := cdr.CGREvent()

		if cgrEvent.Tenant != cdr.Tenant {
			t.Errorf("CGREvent().Tenant = %s, expected %s", cgrEvent.Tenant, cdr.Tenant)
		}

		if cgrEvent.Event["key"] != cdr.Event["key"] {
			t.Errorf("CGREvent().Event = %v, expected %v", cgrEvent.Event, cdr.Event)
		}

		if cgrEvent.APIOpts["opt_key"] != cdr.Opts["opt_key"] {
			t.Errorf("CGREvent().APIOpts = %v, expected %v", cgrEvent.APIOpts, cdr.Opts)
		}

		if cgrEvent.ID == "" {
			t.Errorf("CGREvent().ID is empty, expected a generated ID")
		}
	}
}

func TestCDRsToCGREvents(t *testing.T) {
	{
		cdrs := []*CDR{
			{
				Tenant: "tenant1",
				Event:  map[string]interface{}{"event_key1": "event_value1"},
				Opts:   map[string]interface{}{"opt_key1": "opt_value1"},
			},
			{
				Tenant: "tenant2",
				Event:  map[string]interface{}{"event_key2": "event_value2"},
				Opts:   map[string]interface{}{"opt_key2": "opt_value2"},
			},
		}

		cgrEvents := CDRsToCGREvents(cdrs)

		if len(cgrEvents) != len(cdrs) {
			t.Errorf("CDRsToCGREvents() returned %d events, expected %d", len(cgrEvents), len(cdrs))
		}

		for i, cgrEvent := range cgrEvents {
			cdr := cdrs[i]

			if cgrEvent.Tenant != cdr.Tenant {
				t.Errorf("CGREvent[%d].Tenant = %s, expected %s", i, cgrEvent.Tenant, cdr.Tenant)
			}

			if cgrEvent.Event["event_key1"] != cdr.Event["event_key1"] {
				t.Errorf("CGREvent[%d].Event = %v, expected %v", i, cgrEvent.Event, cdr.Event)
			}

			if cgrEvent.APIOpts["opt_key1"] != cdr.Opts["opt_key1"] {
				t.Errorf("CGREvent[%d].APIOpts = %v, expected %v", i, cgrEvent.APIOpts, cdr.Opts)
			}

			if cgrEvent.ID == "" {
				t.Errorf("CGREvent[%d].ID is empty, expected a generated ID", i)
			}
		}
	}

	{
		cdrs := []*CDR{}

		cgrEvents := CDRsToCGREvents(cdrs)

		if len(cgrEvents) != 0 {
			t.Errorf("CDRsToCGREvents() returned %d events, expected 0", len(cgrEvents))
		}
	}
}
