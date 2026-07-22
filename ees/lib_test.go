// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
