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
	"testing"

	"github.com/cgrates/cgrates/config"
)

func TestWriteFailedPosts(t *testing.T) {
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

	if filename, err := filepath.Glob(filepath.Join(dir, "*.gob")); err != nil {
		t.Error(err)
	} else if len(filename) == 0 {
		t.Error("Expecting one file")
	} else if len(filename) > 1 {
		t.Error("Expecting only one file")
	}
}
