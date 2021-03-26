// +build integration

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
		module: "module",
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal("Error removing folder: ", dir, err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal("Error creating folder: ", dir, err)
	}
	config.CgrConfig().GeneralCfg().FailedPostsDir = dir
	writeFailedPosts("itmID", exportEvent)

	if filename, err := filepath.Glob(filepath.Join(dir, "module|*.gob")); err != nil {
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
		Events: []interface{}{"something1", "something2"},
		Path:   "path",
		Format: "test",
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
