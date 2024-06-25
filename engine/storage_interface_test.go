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

package engine

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ugorji/go/codec"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bson.D
	}{
		{
			name:     "Marshal struct",
			input:    struct{ Name string }{Name: "Cgrates"},
			expected: bson.D{{Key: "name", Value: "Cgrates"}},
		},
		{
			name:     "Marshal map",
			input:    map[string]interface{}{"port": int32(2012)},
			expected: bson.D{{Key: "port", Value: int32(2012)}},
		},
	}

	marshaler := BSONMarshaler{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := marshaler.Marshal(tt.input)
			if err != nil {
				t.Errorf("Marshal() error = %v, wantErr false", err)
				return
			}
			var gotDoc bson.D
			err = bson.Unmarshal(got, &gotDoc)
			if err != nil {
				t.Errorf("Failed to unmarshal BSON bytes: %v", err)
				return
			}
			if !reflect.DeepEqual(gotDoc, tt.expected) {
				t.Errorf("Marshal() got = %v, want %v", gotDoc, tt.expected)
			}
		})
	}
}

func TestBSONMarshalUnmarshal(t *testing.T) {
	marshaler := BSONMarshaler{}

	type TestStruct struct {
		Name string
		Port int
	}
	originalStruct := TestStruct{Name: "Cgrates", Port: 2012}
	data, err := marshaler.Marshal(originalStruct)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var resultStruct TestStruct
	err = marshaler.Unmarshal(data, &resultStruct)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(originalStruct, resultStruct) {
		t.Fatalf("Unmarshal() got = %v, want %v", resultStruct, originalStruct)
	}

	originalMap := map[string]interface{}{"name": "Cgrates", "port": int32(2012)}
	data, err = marshaler.Marshal(originalMap)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var resultMap map[string]interface{}
	err = marshaler.Unmarshal(data, &resultMap)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !reflect.DeepEqual(originalMap, resultMap) {
		t.Fatalf("Unmarshal() got = %v, want %v", resultMap, originalMap)
	}
}

func TestGOBUnmarshal(t *testing.T) {
	marshaler := GOBMarshaler{}

	type TestStruct struct {
		Name string
		Port int
	}
	originalStruct := TestStruct{Name: "Cgrates", Port: 2012}

	data, err := marshaler.Marshal(originalStruct)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var resultStruct TestStruct
	err = marshaler.Unmarshal(data, &resultStruct)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(originalStruct, resultStruct) {
		t.Fatalf("Unmarshal() got = %v, want %v", resultStruct, originalStruct)
	}

	originalMap := map[string]interface{}{"name": "Cgrates", "port": 2012}

	data, err = marshaler.Marshal(originalMap)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var resultMap map[string]interface{}
	err = marshaler.Unmarshal(data, &resultMap)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(originalMap, resultMap) {
		t.Fatalf("Unmarshal() got = %v, want %v", resultMap, originalMap)
	}
}

func TestBincUnmarshal(t *testing.T) {

	bh := &codec.BincHandle{}
	marshaler := BincMarshaler{bh: bh}

	type TestData struct {
		Name string
		Port int
	}
	original := TestData{Name: "Cgrates", Port: 2012}
	data, err := marshaler.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var result TestData
	err = marshaler.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(original, result) {
		t.Errorf("Unmarshal() got = %v, want %v", result, original)
	}
}

func TestNewBincMarshaler(t *testing.T) {
	marshaler := NewBincMarshaler()
	type TestData struct {
		Name string
		Port int
	}
	original := TestData{Name: "Cgrates", Port: 2012}

	data, err := marshaler.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var result TestData
	err = marshaler.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(original, result) {
		t.Errorf("Unmarshal() got = %v, want %v", result, original)
	}
}

func TestJSONBufUnmarshal(t *testing.T) {

	marshaler := JSONBufMarshaler{}

	type TestData struct {
		Name string
		Port int
	}
	original := TestData{Name: "Cgrates", Port: 2012}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result TestData
	err = marshaler.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(original, result) {
		t.Errorf("Unmarshal() got = %v, want %v", result, original)
	}
}

func TestJSONBufMarshal(t *testing.T) {

	marshaler := JSONBufMarshaler{}

	type TestData struct {
		Name string
		Port int
	}
	original := TestData{Name: "Cgrates", Port: 2012}

	data, err := marshaler.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var result TestData
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(original, result) {
		t.Errorf("Marshal() and Unmarshal() mismatch. Got = %v, want %v", result, original)
	}
}

func TestNewMarshalerUnsupported(t *testing.T) {

	m, err := NewMarshaler("Unknown")

	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if m != nil {
		t.Errorf("Expected Marshaler to be nil, got %+v", m)
	}

	expectedErrMsg := "Unsupported marshaler: Unknown"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
	}
}
