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
	nts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	m := ToMapStringString(ts)

	FromMapStringString(m, ts)
	if !reflect.DeepEqual(ts, nts) {
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
	nts := &TestStruct{
		Name:    "1",
		Surname: "2",
		Address: "3",
		Other:   "",
	}
	m := ToMapStringString(ts)
	m["Test"] = "4"
	FromMapStringString(m, ts)

	if !reflect.DeepEqual(ts, nts) {
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
	efMap := GetMapExtraFields(ts, "ExtraFields")

	if !reflect.DeepEqual(efMap, ts.ExtraFields) {
		t.Errorf("expected: %v got: %v", ts.ExtraFields, efMap)
	}
}
