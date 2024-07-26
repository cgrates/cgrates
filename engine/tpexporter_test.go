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
	"archive/zip"
	"bytes"
	"encoding/csv"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestTpExporterNewTPExporter(t *testing.T) {
	str := "test"
	bl := false
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr := NewInternalDB(nil, nil, true, sSCfg.DataDbCfg().Items)
	rcv, err := NewTPExporter(mpStr, str, str, str, str, false)

	if err != nil {
		if err.Error() != "Unsupported file format" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, "", str, utils.CSV, str, false)

	if err != nil {
		if err.Error() != "Missing TPid" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, str, utils.CSV, str, bl)
	if err != nil {
		t.Error(err)
	}

	exp := &TPExporter{
		storDb:     mpStr,
		tpID:       str,
		exportPath: str,
		fileFormat: utils.CSV,
		sep:        't',
		compress:   bl,
		cacheBuff:  new(bytes.Buffer),
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected %v\nreceived %v\n", exp, rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, str, utils.CSV, "", bl)

	if err != nil {
		if err.Error() != "Invalid field separator: " {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, str, utils.CSV, str, true)

	if err != nil {
		if err.Error() != "open test/tpexport.zip: no such file or directory" {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}

	rcv, err = NewTPExporter(mpStr, str, "", utils.CSV, str, true)
	if err != nil {
		t.Error(err)
	}

	exp2 := &TPExporter{
		storDb:     mpStr,
		tpID:       str,
		exportPath: "",
		fileFormat: utils.CSV,
		sep:        't',
		compress:   true,
		cacheBuff:  exp.cacheBuff,
		zipWritter: zip.NewWriter(exp.cacheBuff),
	}

	if !reflect.DeepEqual(rcv, exp2) {
		t.Errorf("\nexpected %v\nreceived %v\n", exp2, rcv)
	}
}

func TestTpExporterExportStats(t *testing.T) {
	str := "test"
	bl := false
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr := NewInternalDB(nil, nil, true, sSCfg.DataDbCfg().Items)
	self, err := NewTPExporter(mpStr, str, str, utils.CSV, str, bl)
	if err != nil {
		t.Error(err)
	}

	rcv := self.ExportStats()
	exp := &utils.ExportedTPStats{
		ExportPath: str,
		Compressed: bl,
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected %v\nreceived %v\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestTpExporterGetCacheBuffer(t *testing.T) {
	str := "test"
	bl := false
	sSCfg, _ := config.NewDefaultCGRConfig()
	mpStr := NewInternalDB(nil, nil, true, sSCfg.DataDbCfg().Items)
	self, err := NewTPExporter(mpStr, str, str, utils.CSV, str, bl)
	if err != nil {
		t.Error(err)
	}

	rcv := self.GetCacheBuffer()
	exp := new(bytes.Buffer)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected %v\nreceived %v\n", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestTPExporterRemoveFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tpExporterTest")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, file := range files {
		tempFile := path.Join(tempDir, file)
		_, err := os.Create(tempFile)
		if err != nil {
			t.Fatalf("Failed to create temp file %s: %v", file, err)
		}
	}
	tpExporter := &TPExporter{
		exportPath:    tempDir,
		exportedFiles: files,
	}
	err = tpExporter.removeFiles()
	if err != nil {
		t.Fatalf("removeFiles() returned an error: %v", err)
	}
	for _, file := range files {
		_, err := os.Stat(path.Join(tempDir, file))
		if !os.IsNotExist(err) {
			t.Errorf("File %s was not removed", file)
		}
	}
}

func TestTPExporterWriteOut(t *testing.T) {
	type TestStruct struct {
		Header1 string
		Header2 string
		Header3 string
	}
	tempDir, err := os.MkdirTemp("", "tpExporterTest")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	tpExporter := &TPExporter{
		exportPath: tempDir,
		sep:        ',',
		fileFormat: utils.CSV,
	}
	tpData := []any{
		TestStruct{"header1", "header2", "header3"},
		TestStruct{"value1", "value2", "value3"},
	}
	fileName := "test.csv"
	err = tpExporter.writeOut(fileName, tpData)
	if err != nil {
		t.Fatalf("writeOut() returned an error: %v", err)
	}
	writtenFilePath := path.Join(tempDir, fileName)
	file, err := os.Open(writtenFilePath)
	if err != nil {
		t.Fatalf("Failed to open written file: %v", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV content: %v", err)
	}
	expected := [][]string{
		{"header1", "header2", "header3"},
		{"value1", "value2", "value3"},
	}
	if len(records) == len(expected) {
		t.Fatalf("Expected %d records, got %d", len(expected), len(records))
	}
	for i := range records {
		for j := range records[i] {
			if records[i][j] != expected[i][j] {
				t.Errorf("Expected %s, got %s at record %d, field %d", expected[i][j], records[i][j], i, j)
			}
		}
	}
}
