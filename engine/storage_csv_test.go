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
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestCsvFileRead(t *testing.T) {
	input := "subject,duration,destination\n101,60,1001\n102,45,1002\n"
	reader := strings.NewReader(input)
	csvReader := csv.NewReader(reader)
	csvFileInstance := &csvFile{
		csvReader: csvReader,
	}
	expectedRecord1 := []string{"subject", "duration", "destination"}
	record1, err := csvFileInstance.Read()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(record1) != len(expectedRecord1) {
		t.Fatalf("expected %d fields, got %d", len(expectedRecord1), len(record1))
	}
	for i, field := range record1 {
		if field != expectedRecord1[i] {
			t.Errorf("expected field %d to be %s, got %s", i, expectedRecord1[i], field)
		}
	}
	expectedRecord2 := []string{"101", "60", "1001"}
	record2, err := csvFileInstance.Read()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(record2) != len(expectedRecord2) {
		t.Fatalf("expected %d fields, got %d", len(expectedRecord2), len(record2))
	}
	for i, field := range record2 {
		if field != expectedRecord2[i] {
			t.Errorf("expected field %d to be %s, got %s", i, expectedRecord2[i], field)
		}
	}
	expectedRecord3 := []string{"102", "45", "1002"}
	record3, err := csvFileInstance.Read()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(record3) != len(expectedRecord3) {
		t.Fatalf("expected %d fields, got %d", len(expectedRecord3), len(record3))
	}
	for i, field := range record3 {
		if field != expectedRecord3[i] {
			t.Errorf("expected field %d to be %s, got %s", i, expectedRecord3[i], field)
		}
	}
	_, err = csvFileInstance.Read()
	if err == nil {
		t.Fatalf("expected EOF error, got nil")
	}
}

func TestCsvURLRead(t *testing.T) {
	input := "field1,field2,field3\nvalue1,value2,value3\n"
	reader := strings.NewReader(input)
	csvReader := csv.NewReader(reader)
	csvURLInstance := &csvURL{
		csvReader: csvReader,
	}
	expectedRecord := []string{"field1", "field2", "field3"}
	record, err := csvURLInstance.Read()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(record) != len(expectedRecord) {
		t.Fatalf("expected %d fields, got %d", len(expectedRecord), len(record))
	}
	for i, field := range record {
		if field != expectedRecord[i] {
			t.Errorf("expected field %d to be %s, got %s", i, expectedRecord[i], field)
		}
	}
	expectedRecord2 := []string{"value1", "value2", "value3"}
	record2, err := csvURLInstance.Read()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(record2) != len(expectedRecord2) {
		t.Fatalf("expected %d fields, got %d", len(expectedRecord2), len(record2))
	}
	for i, field := range record2 {
		if field != expectedRecord2[i] {
			t.Errorf("expected field %d to be %s, got %s", i, expectedRecord2[i], field)
		}
	}
}

func TestCsvURLOpenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("field1,field2,field3\nvalue1,value2,value3\n"))
	}))
	defer server.Close()
	c := &csvURL{}
	err := c.Open(server.URL, ',', 3)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.csvReader == nil {
		t.Fatalf("expected csvReader to be initialized")
	}

	record, err := c.csvReader.Read()
	if err != nil {
		t.Fatalf("expected no error while reading CSV, got: %v", err)
	}

	expectedRecord := []string{"field1", "field2", "field3"}
	for i, field := range expectedRecord {
		if record[i] != field {
			t.Errorf("expected field %s, got: %s", field, record[i])
		}
	}
}

func TestCsvURLOpenInvalidURL(t *testing.T) {
	c := &csvURL{}
	err := c.Open("invalid-url", ',', 3)
	if err == nil {
		t.Fatalf("expected an error for invalid URL, got none")
	}
}

func TestCsvURLOpenNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	c := &csvURL{}
	err := c.Open(server.URL, ',', 3)
	if err == nil {
		t.Fatalf("expected ErrNotFound, got none")
	}
}

func TestCsvURLOpenPathNotReachable(t *testing.T) {
	c := &csvURL{}
	err := c.Open("http://invalid.localhost", ',', 3)
	if err == nil {
		t.Fatalf("expected path not reachable error, got none")
	}
}

func TestCsvURLClosePageNotNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("CsvUrlClose"))
	}))
	defer server.Close()
	c := &csvURL{}
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("expected no error while getting mock URL, got: %v", err)
	}
	c.page = resp.Body
	c.Close()
}

func TestGetTPTrends(t *testing.T) {
	csvStorage := &CSVStorage{}
	tpid := "tpid"
	tenant := "cgrates.org"
	id := "1002"
	result, err := csvStorage.GetTPTrends(tpid, tenant, id)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
	expectedResults := []*utils.TPTrendsProfile{
		{ID: "tpid"},
		{ID: "tpid"},
	}

	for i, res := range result {
		if res.ID != expectedResults[i].ID {
			t.Errorf("Expected result ID %s but got %s", expectedResults[i].ID, res.ID)
		}
	}
}

func TestNewURLCSVStorage(t *testing.T) {
	sep := ','
	dataPath := "http://cgrates.com/destinations.csv;http://cgrates.com/timings.csv;http://cgrates.com/invalidpath"

	csvStorage := NewURLCSVStorage(sep, dataPath)

	if csvStorage == nil {
		t.Fatal("Expected csvStorage to be initialized, but got nil")
	}

	if csvStorage.generator == nil {
		t.Fatal("Expected csvStorage.generator to be initialized, but got nil")
	}

}

func TestNewGoogleCSVStorageInvalidSpreadsheetID(t *testing.T) {
	sep := ','
	spreadsheetID := "invalid-id"
	storage, err := NewGoogleCSVStorage(sep, spreadsheetID)
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	if storage != nil {
		t.Error("Expected nil storage, but got a non-nil value")
	}
}

func TestGetAllFoldersInvalidPath(t *testing.T) {
	nonExistentPath := "/non/existent/path"
	paths, err := getAllFolders(nonExistentPath)
	if err == nil {
		t.Error("Expected error, but got nil")
	}
	if len(paths) != 0 {
		t.Errorf("Expected no paths, but got: %v", paths)
	}
}

func TestJoinURLValidBaseURL(t *testing.T) {
	baseURL := "https://cgrates.com/path"
	fn := "subpath"
	expected := "https://cgrates.com/path/subpath"
	result := joinURL(baseURL, fn)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestGetCfgJSONData(t *testing.T) {
	t.Run("Empty raw JSON", func(t *testing.T) {
		raw := json.RawMessage{}
		data, err := getCfgJSONData(raw)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if len(data) != 0 {
			t.Errorf("Expected no data, but got: %v", data)
		}
	})
	t.Run("Valid raw JSON", func(t *testing.T) {
		raw := json.RawMessage(`{"key": "value"}`)
		data, err := getCfgJSONData(raw)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if string(data) != string(raw) {
			t.Errorf("Expected data %s, but got %s", string(raw), string(data))
		}
	})
	t.Run("Valid file path", func(t *testing.T) {
		path := "/tmp/mockConfigPath.json"
		expectedContent := []byte(`{"key": "value"}`)
		err := os.WriteFile(path, expectedContent, 0600)
		if err != nil {
			t.Fatalf("Failed to create mock file: %v", err)
		}
		defer os.Remove(path)
		raw := json.RawMessage(`"` + path + `"`)
		data, err := getCfgJSONData(raw)
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
		if string(data) != string(expectedContent) {
			t.Errorf("Expected data %s, but got %s", string(expectedContent), string(data))
		}
	})

	t.Run("Invalid file path", func(t *testing.T) {
		raw := json.RawMessage(`"/nonexistent/path/to/file.json"`)
		data, err := getCfgJSONData(raw)
		if err == nil {
			t.Errorf("Expected error, but got none")
		}
		if data != nil {
			t.Errorf("Expected no data, but got: %v", data)
		}
	})
}
