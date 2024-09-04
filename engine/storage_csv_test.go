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
	"os"
	"path/filepath"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestAppendName(t *testing.T) {
	paths := []string{"/path/to/dir1", "/path/to/dir2"}
	fileName := "file.csv"
	expected := []string{"/path/to/dir1/file.csv", "/path/to/dir2/file.csv"}
	result := appendName(paths, fileName)
	if len(result) != len(expected) {
		t.Errorf("TestAppendName (Multiple paths): Length mismatch. Expected %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("TestAppendName (Multiple paths): Element %d mismatch. Expected %s, got %s", i, expected[i], result[i])
		}
	}
	paths = []string{"/single/path"}
	fileName = "single.csv"
	expected = []string{"/single/path/single.csv"}
	result = appendName(paths, fileName)
	if len(result) != len(expected) {
		t.Errorf("TestAppendName (Single path): Length mismatch. Expected %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("TestAppendName (Single path): Element %d mismatch. Expected %s, got %s", i, expected[i], result[i])
		}
	}
	paths = []string{}
	fileName = "cdr.csv"
	expected = []string{}
	result = appendName(paths, fileName)
	if len(result) != len(expected) {
		t.Errorf("TestAppendName (Empty paths list): Length mismatch. Expected %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("TestAppendName (Empty paths list): Element %d mismatch. Expected %s, got %s", i, expected[i], result[i])
		}
	}
	paths = []string{"/path/to/dir1", "/path/to/dir2"}
	fileName = "cdr.txt"
	expected = []string{"/path/to/dir1/cdr.txt", "/path/to/dir2/cdr.txt"}
	result = appendName(paths, fileName)
	if len(result) != len(expected) {
		t.Errorf("TestAppendName (Different file name): Length mismatch. Expected %d, got %d", len(expected), len(result))
	}
	for i := range result {
		if result[i] != expected[i] {
			t.Errorf("TestAppendName (Different file name): Element %d mismatch. Expected %s, got %s", i, expected[i], result[i])
		}
	}
}

func TestJoinURL(t *testing.T) {
	baseURL := "http://cgrates.org/path"
	fn := "file.txt"
	expected := "http://cgrates.org/path/file.txt"
	result := joinURL(baseURL, fn)
	if result != expected {
		t.Errorf("TestJoinURL (Valid URL): Expected %s, got %s", expected, result)
	}
	baseURL = "http://cgrates.org/path/"
	fn = "file.txt"
	expected = "http://cgrates.org/path/file.txt"
	result = joinURL(baseURL, fn)
	if result != expected {
		t.Errorf("TestJoinURL (Trailing Slash): Expected %s, got %s", expected, result)
	}
	baseURL = "http://cgrates.org/path"
	fn = "/file.txt"
	expected = "http://cgrates.org/path/file.txt"
	result = joinURL(baseURL, fn)
	if result != expected {
		t.Errorf("TestJoinURL (Leading Slash): Expected %s, got %s", expected, result)
	}
	baseURL = "http://cgrates.org"
	fn = "file.txt"
	expected = "http://cgrates.org/file.txt"
	result = joinURL(baseURL, fn)
	if result != expected {
		t.Errorf("TestJoinURL (No Path): Expected %s, got %s", expected, result)
	}
	baseURL = "http://cgrates.org/path"
	fn = "dir1/dir2/file.txt"
	expected = "http://cgrates.org/path/dir1/dir2/file.txt"
	result = joinURL(baseURL, fn)
	if result != expected {
		t.Errorf("TestJoinURL (Multiple Segments): Expected %s, got %s", expected, result)
	}
}

func TestGetAllFolders(t *testing.T) {
	baseDir := t.TempDir()
	subDir1 := filepath.Join(baseDir, "subdir1")
	subDir2 := filepath.Join(baseDir, "subdir2")
	subSubDir := filepath.Join(subDir1, "subsubdir")
	err := os.MkdirAll(subSubDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectories: %v", err)
	}
	file, err := os.Create(filepath.Join(baseDir, "file.txt"))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()
	dirs, err := getAllFolders(baseDir)
	if err != nil {
		t.Fatalf("getAllFolders returned an error: %v", err)
	}
	for _, dir := range dirs {
		if dir != baseDir && dir != subDir1 && dir != subDir2 && dir != subSubDir {
			t.Errorf("Unexpected directory found in result: %s", dir)
		}
	}
}

func TestGetTpIds(t *testing.T) {
	csvStorage := &CSVStorage{}
	colName := "colName"
	tpIds, err := csvStorage.GetTpIds(colName)
	if err != utils.ErrNotImplemented {
		t.Errorf("GetTpIds() returned error = %v, want %v", err, utils.ErrNotImplemented)
	}
	if tpIds != nil {
		t.Errorf("GetTpIds() returned slice = %v, want nil", tpIds)
	}
}

func TestStorageGetTpTableIds(t *testing.T) {
	csvStorage := &CSVStorage{}
	tpid := "Tpid"
	table := "Table"
	distinct := utils.TPDistinctIds{}
	filters := map[string]string{"filterKey": "filterValue"}
	paginator := &utils.PaginatorWithSearch{}
	result, err := csvStorage.GetTpTableIds(tpid, table, distinct, filters, paginator)
	if err != utils.ErrNotImplemented {
		t.Errorf("GetTpTableIds() returned error = %v, want %v", err, utils.ErrNotImplemented)
	}
	if result != nil {
		t.Errorf("GetTpTableIds() returned slice = %v, want nil", result)
	}
}

func TestNewCsvFile(t *testing.T) {
	csvReader := NewCsvFile()
	_, ok := csvReader.(csvReaderCloser)
	if !ok {
		t.Errorf("NewCsvFile() did not return a type that implements csvReaderCloser interface")
	}
}

func TestOpen(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	testData := "col1,col2\nval1,val2"
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}
	csvFile := &csvFile{}
	if err := csvFile.Open(tmpFile.Name(), ',', 2); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	if csvFile.fp == nil {
		t.Errorf("Expected non-nil file pointer, got nil")
	}
	if csvFile.csvReader == nil {
		t.Errorf("Expected non-nil csvReader, got nil")
	}
	if csvFile.csvReader.Comma != ',' {
		t.Errorf("Expected comma delimiter to be ',', got '%c'", csvFile.csvReader.Comma)
	}
	if csvFile.csvReader.FieldsPerRecord != 2 {
		t.Errorf("Expected FieldsPerRecord to be 2, got %d", csvFile.csvReader.FieldsPerRecord)
	}
	if err := csvFile.fp.Close(); err != nil {
		t.Errorf("Failed to close file: %v", err)
	}
}
