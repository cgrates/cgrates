//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestWriteFldPosts(t *testing.T) {
	// can't convert
	var notanExportEvent string
	writeFailedPosts("somestring", notanExportEvent)
	// can convert & write
	dir := "/tmp/engine/libcdre_test/"
	exportEvent := &ExportEvents{
		failedPostsDir: dir,
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal("Error removing folder: ", dir, err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal("Error creating folder: ", dir, err)
	}
	config.CgrConfig().EEsCfg().FailedPosts.Dir = dir
	writeFailedPosts("itmID", exportEvent)

	if filename, err := filepath.Glob(filepath.Join(dir, "EEs|*.gob")); err != nil {
		t.Error(err)
	} else if len(filename) == 0 {
		t.Error("Expecting one file")
	} else if len(filename) > 1 {
		t.Error("Expecting only one file")
	}
}

func TestWriteToFile(t *testing.T) {
	filePath := "/tmp/engine/libcdre_test/writeToFile.txt"
	exportEvent := &ExportEvents{}
	//call WriteToFile function
	if err := exportEvent.WriteToFile(filePath); err != nil {
		t.Error(err)
	}
	// check if the file exists / throw error if the file doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("File doesn't exists")
	}
	//check if the file was written correctly
	rcv, err := NewExportEventsFromFile(filePath)
	if err != nil {
		t.Errorf("Error deconding the file content: %+v", err)
	}
	if !reflect.DeepEqual(rcv, exportEvent) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(exportEvent), utils.ToJSON(rcv))
	}
	//populate the exportEvent struct
	exportEvent = &ExportEvents{
		Events: []any{"something1", "something2"},
		Path:   "path",
		Type:   "test",
	}
	filePath = "/tmp/engine/libcdre_test/writeToFile2.txt"
	if err := exportEvent.WriteToFile(filePath); err != nil {
		t.Error(err)
	}
	// check if the file exists / throw error if the file doesn't exist
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("File doesn't exists")
	}
	//check if the file was written correctly
	rcv, err = NewExportEventsFromFile(filePath)
	if err != nil {
		t.Errorf("Error deconding the file content: %+v", err)
	}
	if !reflect.DeepEqual(rcv, exportEvent) {
		t.Errorf("Expected: %+v,\nReceived: %+v", utils.ToJSON(exportEvent), utils.ToJSON(rcv))
	}
	//wrong path *reading
	exportEvent = &ExportEvents{}
	filePath = "/tmp/engine/libcdre_test/wrongpath.txt"
	if _, err = NewExportEventsFromFile(filePath); err == nil || err.Error() != "open /tmp/engine/libcdre_test/wrongpath.txt: no such file or directory" {
		t.Errorf("Expecting: 'open /tmp/engine/libcdre_test/wrongpath.txt: no such file or directory',\nReceived: '%+v'", err)
	}
	//wrong path *writing
	filePath = utils.EmptyString
	if err := exportEvent.WriteToFile(filePath); err == nil || err.Error() != "open : no such file or directory" {
		t.Error(err)
	}
}
