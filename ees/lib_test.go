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

package ees

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/cgrates/cgrates/engine"
)

var exportPath = []string{"/tmp/testCSV", "/tmp/testComposedCSV", "/tmp/testFWV", "/tmp/testCSVMasked",
	"/tmp/testCSVfromVirt", "/tmp/testCSVExpTemp"}

func testCreateDirectory(t *testing.T) {
	for _, dir := range exportPath {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testCleanDirectory(t *testing.T) {
	for _, dir := range exportPath {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func TestGetOneData(t *testing.T) {
	type testData struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}
	ub := &engine.Account{}
	expected1, _ := json.Marshal(ub)
	got1, err1 := getOneData(ub, nil)
	if err1 != nil {
		t.Errorf("getOneData() error = %v", err1)
	}
	if string(got1) != string(expected1) {
		t.Errorf("getOneData() got = %v, want %v", string(got1), string(expected1))
	}
	extraData := testData{Field1: "test", Field2: 123}
	expected2, _ := json.Marshal(extraData)
	got2, err2 := getOneData(nil, extraData)
	if err2 != nil {
		t.Errorf("getOneData() error = %v", err2)
	}
	if string(got2) != string(expected2) {
		t.Errorf("getOneData() got = %v, want %v", string(got2), string(expected2))
	}
	expected3 := []byte(nil)
	got3, err3 := getOneData(nil, nil)
	if err3 != nil {
		t.Errorf("getOneData() error = %v", err3)
	}
	if (got3 != nil && string(got3) != string(expected3)) || (got3 == nil && expected3 != nil) {
		t.Errorf("getOneData() got = %v, want %v", got3, expected3)
	}
}
