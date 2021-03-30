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
	"os"
	"reflect"
	"sort"
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
	jsn := utils.ToJSON(&utils.TenantIDWithAPIOpts{
		TenantID: new(utils.TenantID),
		APIOpts:  make(map[string]interface{}),
	})

	line := FromJSON([]byte(jsn), []string{"Tenant", "ID", "APIOpts"})
	expected := `Tenant="" ID="" APIOpts={}`
	if line != expected {
		t.Log(jsn)
		t.Errorf("Expected: %s got: '%s'", expected, line)
	}
}

func TestGetStringValueInterface(t *testing.T) {
	dflt := utils.StringSet{}
	expected := getSliceAsString([]interface{}{}, dflt)
	rply := getStringValue([]interface{}{}, dflt)
	if rply != expected {
		t.Errorf("Expecting: %s , received: %s", expected, rply)
	}
}

func TestGetFormatedSliceResultCase2(t *testing.T) {
	dflt := utils.StringSet{}
	rply := GetFormatedSliceResult(true, dflt)
	expected := true
	if reflect.DeepEqual(expected, rply) {
		t.Errorf("Expecting: %+v , received: %+v", expected, rply)
	}
}

func TestCommandExecuterUsage(t *testing.T) {
	testStruct := &CommandExecuter{}
	testStruct.command = commands["accounts_profile"]
	result := testStruct.Usage()
	expected := "\n\tUsage: accounts_profile APIOpts=null \n"
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected <%+q>, Received <%+q>", expected, result)
	}

}

func TestCommandExecuterLocalExecute(t *testing.T) {
	testStruct := &CommandExecuter{}
	testStruct.command = commands["accounts_profile"]
	result := testStruct.LocalExecute()
	expected := utils.EmptyString
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, result)
	}
}

func TestCommandExecuterLocalFromArgs(t *testing.T) {
	null, _ := os.Open(os.DevNull)
	stdout := os.Stdout
	os.Stdout = null
	testStruct := &CommandExecuter{}
	testStruct.command = commands["status"]
	cmdArgs := "argument_test"
	result := testStruct.FromArgs(cmdArgs, true)
	if !reflect.DeepEqual(nil, result) {
		t.Errorf("Expected <%+v>, Received <%+v>", nil, result)
	}
	os.Stdout = stdout
}

type mockCommandExecuter struct {
	Commander
}

func (*mockCommandExecuter) Name() string {
	return utils.EmptyString
}

func (*mockCommandExecuter) RpcMethod() string {
	return utils.EmptyString
}

func (*mockCommandExecuter) RpcParams(reset bool) interface{} {
	return struct{}{}
}

func (*mockCommandExecuter) PostprocessRpcParams() error {
	return nil
}

func (*mockCommandExecuter) RpcResult() interface{} {
	return nil
}

func (*mockCommandExecuter) ClientArgs() (args []string) {
	return
}

func TestCommandExecuterLocalFromArgsCase2(t *testing.T) {
	testStruct := &CommandExecuter{new(mockCommandExecuter)}
	cmdArgs := "argument_test"
	err := testStruct.FromArgs(cmdArgs, true)
	expected := "json: Unmarshal(non-pointer struct {})"
	if err == nil || err.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nRecevied <%+v>", expected, err)
	}
}

func TestCommandExecuterClientArgs(t *testing.T) {
	testStruct := &CommandExecuter{}
	testStruct.command = commands["accounts_profile"]
	result := testStruct.clientArgs(testStruct.command.RpcParams(true))
	expected := []string{"APIOpts", "ID", "Tenant"}
	sort.Strings(result)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, result)
	}
}

type mockTest struct {
	a int
	B *bool
	C struct {
	}
	D time.Time
}

func TestCommandExecuterClientArgsCase(t *testing.T) {
	testStruct := &CommandExecuter{}
	testStruct.command = commands["accounts"]
	result := testStruct.clientArgs(new(mockTest))
	expected := []string{"B", "D"}
	sort.Strings(result)
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected <%+v>, Received <%+v>", expected, result)
	}
}
