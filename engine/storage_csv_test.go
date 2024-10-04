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
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewFileCSVStorageSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	subDirs := []string{"folder1", "folder2"}
	for _, dir := range subDirs {
		err := os.Mkdir(filepath.Join(tmpDir, dir), 0755)
		if err != nil {
			t.Fatalf("Could not create test directory %q: %v", dir, err)
		}
	}

	storage, err := NewFileCSVStorage(',', tmpDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if storage == nil {
		t.Fatalf("Expected CSVStorage, got nil")
	}

	var filteredResProfilesFn []string
	for _, path := range storage.resProfilesFn {
		if strings.Contains(path, "folder1") || strings.Contains(path, "folder2") {
			filteredResProfilesFn = append(filteredResProfilesFn, path)
		}
	}

	if len(filteredResProfilesFn) != 2 {
		t.Fatalf("Expected 2 resource profile paths, got %d", len(filteredResProfilesFn))
	}

	expectedPaths := []string{
		filepath.Join(tmpDir, "folder1", "Resources.csv"),
		filepath.Join(tmpDir, "folder2", "Resources.csv"),
	}

	for i, path := range filteredResProfilesFn {
		if !strings.EqualFold(path, expectedPaths[i]) {
			t.Errorf("Expected %v, got %v", expectedPaths[i], path)
		}
	}
}

func TestNewCsvFile(t *testing.T) {
	csv := NewCsvFile()

	if csv == nil {
		t.Fatalf("Expected a non-nil csvReaderCloser, got nil")
	}

	_, ok := csv.(*csvFile)
	if !ok {
		t.Fatalf("Expected type *csvFile, got %T", csv)
	}
}

func TestGetTpTableIds(t *testing.T) {
	csvStorage := &CSVStorage{}

	tpid := "TPID"
	table := "Table"
	distinct := []string{"col1", "col2"}
	filters := map[string]string{"key": "value"}
	paginator := &utils.PaginatorWithSearch{}

	ids, err := csvStorage.GetTpTableIds(tpid, table, distinct, filters, paginator)

	if ids != nil {
		t.Errorf("Expected ids to be nil, got %v", ids)
	}

	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestGetTpIds(t *testing.T) {
	csvStorage := &CSVStorage{}

	colName := "Column"

	ids, err := csvStorage.GetTpIds(colName)

	if ids != nil {
		t.Errorf("Expected ids to be nil, got %v", ids)
	}

	if err != utils.ErrNotImplemented {
		t.Errorf("Expected error to be %v, got %v", utils.ErrNotImplemented, err)
	}
}

func TestJoinURL(t *testing.T) {
	tests := []struct {
		baseURL string
		fn      string
		want    string
	}{
		{"http://cgrates.org", "path/to/resource", "http://cgrates.org/path/to/resource"},
		{"http://cgrates.org/", "path/to/resource", "http://cgrates.org/path/to/resource"},
		{"http://cgrates.org/path/", "to/resource", "http://cgrates.org/path/to/resource"},
		{"invalid-url", "path/to/resource", "invalid-url/path/to/resource"},
		{"http://cgrates.org", "", "http://cgrates.org"},
		{"http://cgrates.org/", "", "http://cgrates.org/"},
		{"http://cgrates.org/path", "", "http://cgrates.org/path"},
	}

	for _, tt := range tests {
		t.Run(tt.baseURL+tt.fn, func(t *testing.T) {
			got := joinURL(tt.baseURL, tt.fn)
			if got != tt.want {
				t.Errorf("joinURL(%q, %q) = %q; want %q", tt.baseURL, tt.fn, got, tt.want)
			}
		})
	}
}

func TestNewURLCSVStorage(t *testing.T) {
	sep := ','
	dataPath := "http://cgrates.org/data1,http://cgrates.org/data2"

	csvStorage := NewURLCSVStorage(sep, dataPath)

	if csvStorage == nil {
		t.Fatalf("Expected CSVStorage to be initialized, got nil")
	}

	if csvStorage.generator == nil {
		t.Fatal("Expected generator to be set, got nil")
	}
}
