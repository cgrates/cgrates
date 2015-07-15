package utils

import (
	"reflect"
	"testing"
)

func TestStructMapStruct(t *testing.T) {
	type TestStruct struct {
		Name    string
		Surname string
		Address string
		Other   string
	}
	ts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	m, err := ToMapStringString(ts)
	if err != nil {
		t.Error("Error converting to map: ", err)
	}
	out, err := FromMapStringString(m, ts)
	if err != nil {
		t.Error("Error converting to struct: ", err)
	}
	nts := out.(TestStruct)
	if !reflect.DeepEqual(ts, &nts) {
		t.Log(m)
		t.Errorf("Expected: %+v got: %+v", ts, nts)
	}
}

func TestMapStructAddStructs(t *testing.T) {
	type TestStruct struct {
		Name    string
		Surname string
		Address string
		Other   string
	}
	ts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	m, err := ToMapStringString(ts)
	if err != nil {
		t.Error("Error converting to map: ", err)
	}
	m["Test"] = "4"
	out, err := FromMapStringString(m, ts)
	if err != nil {
		t.Error("Error converting to struct: ", err)
	}
	nts := out.(TestStruct)
	if !reflect.DeepEqual(ts, &nts) {
		t.Log(m)
		t.Errorf("Expected: %+v got: %+v", ts, nts)
	}
}

func TestStructExtraFields(t *testing.T) {
	ts := struct {
		Name        string
		Surname     string
		Address     string
		ExtraFields map[string]string
	}{
		Name:    "1",
		Surname: "2",
		Address: "3",
		ExtraFields: map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		},
	}
	efMap, err := GetMapExtraFields(ts, "ExtraFields")
	if err != nil {
		t.Error("Error getting extra fields: ", err)
	}

	if !reflect.DeepEqual(efMap, ts.ExtraFields) {
		t.Errorf("expected: %v got: %v", ts.ExtraFields, efMap)
	}
}
