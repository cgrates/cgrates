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
package console

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestToJSON(t *testing.T) {
	jsn := ToJSON(`TimeStart="Test"     Crazy = 1 Mama=true coco Test=1`)
	expected := `{"TimeStart":"Test","Crazy":1,"Mama":true,"Test":1}`
	if string(jsn) != expected {
		t.Errorf("Expected: %s got: %s", expected, jsn)
	}
}

func TestToJSONValid(t *testing.T) {
	jsn := ToJSON(`TimeStart="Test"     Crazy = 1 Mama=true coco Test=1`)
	a := make(map[string]interface{})
	if err := json.Unmarshal(jsn, &a); err != nil {
		t.Error("Error unmarshaling generated json: ", err)
	}
}

func TestToJSONEmpty(t *testing.T) {
	jsn := ToJSON("")
	if string(jsn) != `{}` {
		t.Error("Error empty: ", string(jsn))
	}
}

func TestToJSONString(t *testing.T) {
	jsn := ToJSON("1002")
	if string(jsn) != `{"Item":"1002"}` {
		t.Error("Error string: ", string(jsn))
	}
}

func TestToJSONArrayNoSpace(t *testing.T) {
	jsn := ToJSON(`Param=["id1","id2","id3"] Another="Patram"`)
	if string(jsn) != `{"Param":["id1","id2","id3"],"Another":"Patram"}` {
		t.Error("Error string: ", string(jsn))
	}
}

func TestToJSONArraySpace(t *testing.T) {
	jsn := ToJSON(`Param=["id1", "id2", "id3"]  Another="Patram"`)
	if string(jsn) != `{"Param":["id1", "id2", "id3"],"Another":"Patram"}` {
		t.Error("Error string: ", string(jsn))
	}
}

func TestFromJSON(t *testing.T) {
	line := FromJSON([]byte(`{"TimeStart":"Test","Crazy":1,"Mama":true,"Test":1}`), []string{"TimeStart", "Crazy", "Mama", "Test"})
	expected := `TimeStart="Test" Crazy=1 Mama=true Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONInterestingFields(t *testing.T) {
	line := FromJSON([]byte(`{"TimeStart":"Test","Crazy":1,"Mama":true,"Test":1}`), []string{"TimeStart", "Test"})
	expected := `TimeStart="Test" Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONString(t *testing.T) {
	line := FromJSON([]byte(`1002`), []string{"string"})
	expected := `"1002"`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONArrayNoSpace(t *testing.T) {
	line := FromJSON([]byte(`{"Param":["id1","id2","id3"], "TimeStart":"Test", "Test":1}`), []string{"Param", "TimeStart", "Test"})
	expected := `Param=["id1","id2","id3"] TimeStart="Test" Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestFromJSONArraySpace(t *testing.T) {
	line := FromJSON([]byte(`{"Param":["id1", "id2", "id3"], "TimeStart":"Test", "Test":1}`), []string{"Param", "TimeStart", "Test"})
	expected := `Param=["id1", "id2", "id3"] TimeStart="Test" Test=1`
	if line != expected {
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestGetStringValue(t *testing.T) {
	dflt := utils.StringSet{}
	expected := "10"
	if rply := getStringValue(int64(10), dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = "true"
	if rply := getStringValue(true, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = "null"
	if rply := getStringValue(nil, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = "10.5"
	if rply := getStringValue(10.5, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = "[10,5]"
	if rply := getStringValue([]float32{10, 5}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = `{"ID":"id1","TimeValue":10000}`
	if rply := getStringValue(struct {
		ID        string
		TimeValue int64
	}{ID: "id1", TimeValue: 10000}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	if rply := getStringValue(map[string]interface{}{
		"ID":        "id1",
		"TimeValue": 10000}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = `{"ID":"id1","TimeValue":"1s"}`
	if rply := getStringValue(map[string]interface{}{
		"ID":        "id1",
		"TimeValue": int64(time.Second)}, utils.StringSet{"TimeValue": {}}); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = "[10,20,30]"
	if rply := getSliceAsString([]interface{}{10, 20, 30}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}
}

func TestGetSliceAsString(t *testing.T) {
	dflt := utils.StringSet{}
	expected := "[10,20,30]"
	if rply := getSliceAsString([]interface{}{10, 20, 30}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = `["test1","test2","test3"]`
	if rply := getSliceAsString([]interface{}{"test1", "test2", "test3"}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}
}

func TestGetMapAsString(t *testing.T) {
	dflt := utils.StringSet{}
	expected := `{"ID":"id1","TimeValue":10000}`
	if rply := getStringValue(map[string]interface{}{
		"ID":        "id1",
		"TimeValue": 10000}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = `{"ID":"id1","TimeValue":"1s"}`
	if rply := getStringValue(map[string]interface{}{
		"ID":        "id1",
		"TimeValue": int64(time.Second)}, utils.StringSet{"TimeValue": {}}); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}
}

func TestGetFormatedResult(t *testing.T) {
	dflt := utils.StringSet{}
	expected := `{
 "ID": "id1",
 "TimeValue": 10000
}`
	if rply := GetFormatedResult(map[string]interface{}{
		"ID":        "id1",
		"TimeValue": 10000}, dflt); rply != expected {
		t.Errorf("Expecting: %q , received: %q", expected, rply)
	}

	expected = `{
 "ID": "id1",
 "TimeValue": "1s"
}`
	if rply := GetFormatedResult(map[string]interface{}{
		"ID":        "id1",
		"TimeValue": int64(time.Second)}, utils.StringSet{"TimeValue": {}}); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = `{
 "ID": "id1",
 "TimeValue": 10000
}`
	if rply := GetFormatedResult(struct {
		ID        string
		TimeValue int64
	}{ID: "id1", TimeValue: 10000}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}
}

func TestGetFormatedSliceResult(t *testing.T) {
	dflt := utils.StringSet{}
	expected := "[10,20,30]"
	if rply := getSliceAsString([]interface{}{10, 20, 30}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}

	expected = `["test1","test2","test3"]`
	if rply := getSliceAsString([]interface{}{"test1", "test2", "test3"}, dflt); rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}
}

func TestFromJSONInterestingFields2(t *testing.T) {
	jsn := utils.ToJSON(&utils.TenantIDWithOpts{
		TenantID: new(utils.TenantID),
		Opts:     make(map[string]interface{}),
	})

	line := FromJSON([]byte(jsn), []string{"Tenant", "ID", "Opts"})
	expected := `Tenant="" ID="" Opts={}`
	if line != expected {
		t.Log(jsn)
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}
