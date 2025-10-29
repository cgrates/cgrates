/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ers

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func testCreateDirs(t *testing.T) {
	for _, dir := range []string{"/tmp/ers/in", "/tmp/ers/out",
		"/tmp/ers2/in", "/tmp/ers2/out",
		"/tmp/init_session/in", "/tmp/init_session/out",
		"/tmp/terminate_session/in", "/tmp/terminate_session/out",
		"/tmp/cdrs/in", "/tmp/cdrs/out",
		"/tmp/ers_with_filters/in", "/tmp/ers_with_filters/out",
		"/tmp/xmlErs/in", "/tmp/xmlErs/out",
		"/tmp/xmlErs2/in", "/tmp/xmlErs2/out",
		"/tmp/fwvErs/in", "/tmp/fwvErs/out",
		"/tmp/partErs1/in", "/tmp/partErs1/out",
		"/tmp/partErs2/in", "/tmp/partErs2/out",
		"/tmp/flatstoreErs/in", "/tmp/flatstoreErs/out",
		"/tmp/ErsJSON/in", "/tmp/ErsJSON/out",
		"/tmp/readerWithTemplate/in", "/tmp/readerWithTemplate/out",
		"/tmp/flatstoreACKErs/in", "/tmp/flatstoreACKErs/out",
		"/tmp/flatstoreMMErs/in", "/tmp/flatstoreMMErs/out"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testCleanupFiles(t *testing.T) {
	for _, dir := range []string{"/tmp/ers",
		"/tmp/ers2", "/tmp/init_session", "/tmp/terminate_session",
		"/tmp/cdrs", "/tmp/ers_with_filters", "/tmp/xmlErs", "/tmp/fwvErs",
		"/tmp/partErs1", "/tmp/partErs2", "tmp/flatstoreErs", "/tmp/ErsJSON",
		"/tmp/readerWithTemplate", "/tmp/flatstoreACKErs", "/tmp/flatstoreMMErs",
		"/tmp/xmlErs2"} {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func TestProcessReaderDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "testProcessReaderDir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	file1 := filepath.Join(dir, "file1.csv")
	file2 := filepath.Join(dir, "file2.csv")
	file3 := filepath.Join(dir, "file3.txt")

	if err := os.WriteFile(file1, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}
	if err := os.WriteFile(file3, []byte("data"), 0644); err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	var processedFiles []string
	var mu sync.Mutex
	mockFunc := func(fn string) error {
		mu.Lock()
		defer mu.Unlock()
		processedFiles = append(processedFiles, fn)
		return nil
	}

	processReaderDir(dir, ".csv", mockFunc)

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(processedFiles) != 2 {
		t.Errorf("Expected 2 files to be processed, got %d", len(processedFiles))
	}

	expectedFiles := []string{"file1.csv", "file2.csv"}
	for _, expected := range expectedFiles {
		found := false
		for _, processed := range processedFiles {
			if strings.HasSuffix(processed, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %s to be processed, but it was not", expected)
		}
	}
}
