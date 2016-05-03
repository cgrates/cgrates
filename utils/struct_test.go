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

func TestStructFromMapStringInterface(t *testing.T) {
	ts := &struct {
		Name     string
		Class    *string
		List     []string
		Elements struct {
			Type  string
			Value float64
		}
	}{}
	s := "test2"
	m := map[string]interface{}{
		"Name":  "test1",
		"Class": &s,
		"List":  []string{"test3", "test4"},
		"Elements": struct {
			Type  string
			Value float64
		}{
			Type:  "test5",
			Value: 9.8,
		},
	}
	if err := FromMapStringInterface(m, ts); err != nil {
		t.Logf("ts: %+v", ToJSON(ts))
		t.Error("Error converting map to struct: ", err)
	}
}
