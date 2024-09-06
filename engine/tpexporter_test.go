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
	"bytes"
	"os"
	"path"
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/cgrates/cgrates/utils"
)

func TestTPExporterGetCacheBuffer(t *testing.T) {
	expectedBuffer := bytes.NewBufferString("test data")
	tpExp := &TPExporter{
		cacheBuff: expectedBuffer,
	}
	actualBuffer := tpExp.GetCacheBuffer()
	if actualBuffer != expectedBuffer {
		t.Errorf("GetCacheBuffer() returned the wrong buffer. Expected %v, got %v", expectedBuffer, actualBuffer)
	}
	if actualBuffer.String() != "test data" {
		t.Errorf("GetCacheBuffer() returned a buffer with unexpected content. Expected %s, got %s", "test data", actualBuffer.String())
	}
}

func TestTPExporterExportStats(t *testing.T) {
	expectedExportPath := "test/path"
	expectedExportedFiles := []string{"file1.csv", "file2.csv"}
	expectedCompressed := true
	tpExp := &TPExporter{
		exportPath:    expectedExportPath,
		exportedFiles: expectedExportedFiles,
		compress:      expectedCompressed,
	}
	actualStats := tpExp.ExportStats()
	if actualStats == nil {
		t.Fatalf("ExportStats() returned nil, expected non-nil *ExportedTPStats")
	}
	expectedStats := &utils.ExportedTPStats{
		ExportPath:    expectedExportPath,
		ExportedFiles: expectedExportedFiles,
		Compressed:    expectedCompressed,
	}
	if !reflect.DeepEqual(actualStats, expectedStats) {
		t.Errorf("ExportStats() returned incorrect stats. Expected %v, got %v", expectedStats, actualStats)
	}
}

func TestTPExporterRemoveFiles(t *testing.T) {
	t.Run("NoExportPath", func(t *testing.T) {
		tpExp := &TPExporter{
			exportPath:    "",
			exportedFiles: []string{"file1.csv", "file2.csv"},
		}
		err := tpExp.removeFiles()
		if err != nil {
			t.Fatalf("removeFiles() returned an unexpected error: %v", err)
		}

	})

	t.Run("RemoveFiles", func(t *testing.T) {
		testDir := t.TempDir()
		file1 := path.Join(testDir, "file1.csv")
		file2 := path.Join(testDir, "file2.csv")
		os.WriteFile(file1, []byte("test content"), 0644)
		os.WriteFile(file2, []byte("test content"), 0644)
		tpExp := &TPExporter{
			exportPath:    testDir,
			exportedFiles: []string{"file1.csv", "file2.csv"},
		}
		err := tpExp.removeFiles()
		if err != nil {
			t.Fatalf("removeFiles() returned an unexpected error: %v", err)
		}
		if _, err := os.Stat(file1); !os.IsNotExist(err) {
			t.Errorf("removeFiles() did not remove %s", file1)
		}
		if _, err := os.Stat(file2); !os.IsNotExist(err) {
			t.Errorf("removeFiles() did not remove %s", file2)
		}
	})
}

func TestNewTPExporter(t *testing.T) {
	var testStorDb LoadStorage
	t.Run("MissingTPid", func(t *testing.T) {
		_, err := NewTPExporter(testStorDb, "", "export/path", utils.CSV, ",", false)
		if err == nil || err.Error() != "Missing TPid" {
			t.Fatalf("Expected error 'Missing TPid', got %v", err)
		}
	})
	t.Run("UnsupportedFileFormat", func(t *testing.T) {
		_, err := NewTPExporter(testStorDb, "testTPid", "export/path", "txt", ",", false)
		if err == nil || err.Error() != "Unsupported file format" {
			t.Fatalf("Expected error 'Unsupported file format', got %v", err)
		}
	})
	t.Run("InvalidFieldSeparator", func(t *testing.T) {
		_, err := NewTPExporter(testStorDb, "testTPid", "export/path", utils.CSV, string(utf8.RuneError), false)
		if err == nil || err.Error() != "Invalid field separator: \uFFFD" {
			t.Fatalf("Expected error 'Invalid field separator', got %v", err)
		}
	})
	t.Run("SuccessWithCompressionAndNoExportPath", func(t *testing.T) {
		tpExp, err := NewTPExporter(testStorDb, "testTPid", "", utils.CSV, ",", true)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if tpExp == nil {
			t.Fatal("Expected non-nil TPExporter, got nil")
		}
		if tpExp.zipWritter == nil {
			t.Fatal("Expected non-nil zip.Writer, got nil")
		}
	})
	t.Run("SuccessWithCompressionAndExportPath", func(t *testing.T) {
		testDir := t.TempDir()
		tpExp, err := NewTPExporter(testStorDb, "testTPid", testDir, utils.CSV, ",", true)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if tpExp == nil {
			t.Fatal("Expected non-nil TPExporter, got nil")
		}
		if tpExp.zipWritter == nil {
			t.Fatal("Expected non-nil zip.Writer, got nil")
		}
		zipFilePath := path.Join(testDir, "tpexport.zip")
		if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
			t.Fatalf("Expected ZIP file to be created at %s, but it was not", zipFilePath)
		}
	})
	t.Run("SuccessWithoutCompression", func(t *testing.T) {
		tpExp, err := NewTPExporter(testStorDb, "testTPid", "export/path", utils.CSV, ",", false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if tpExp == nil {
			t.Fatal("Expected non-nil TPExporter, got nil")
		}
		if tpExp.zipWritter != nil {
			t.Fatal("Expected nil zip.Writer, got non-nil")
		}
	})
}

func TestTPExporterWriteOut(t *testing.T) {
	type Data struct {
		Field1 string
		Field2 string
	}
	tpExp := &TPExporter{
		compress:   false,
		exportPath: "",
		fileFormat: utils.CSV,
		sep:        ',',
	}
	t.Run("empty tpData", func(t *testing.T) {
		err := tpExp.writeOut("testfile.csv", []any{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("exportPath is set", func(t *testing.T) {
		tmpDir := t.TempDir()
		tpExp.exportPath = tmpDir
		tpData := []any{
			Data{"ID1", "ID2"},
			Data{"ID3", "ID4"},
		}
		err := tpExp.writeOut("testfile.csv", tpData)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		expectedFilePath := path.Join(tmpDir, "testfile.csv")
		if _, err := os.Stat(expectedFilePath); err != nil {
			t.Errorf("Expected file to be created at %v, but got error: %v", expectedFilePath, err)
		}
		defer os.Remove(expectedFilePath)
	})
	t.Run("write to buffer", func(t *testing.T) {
		tpExp.exportPath = ""
		tpData := []any{
			Data{"ID1", "ID2"},
			Data{"ID3", "ID4"},
		}
		err := tpExp.writeOut("testfile.csv", tpData)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

	})
}
