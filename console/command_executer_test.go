/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	if string(jsn) != `{"Item":""}` {
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
