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

package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
)

var (
	tests = []func(t *testing.T){
		testUnzip,
		testUnzipADirectory,
		testUnzipOpenFileError,
	}
)

func TestCoreUtilsIT(t *testing.T) {
	for _, tests := range tests {
		t.Run("Core_utils", tests)
	}
}

func testUnzip(t *testing.T) {
	flPath := "/tmp/testUnzip"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	newFile, err := os.Create(path.Join(flPath, "random.zip"))
	if err != nil {
		t.Error(err)
	}

	expectedErr := "zip: not a valid zip file"
	if err := Unzip(path.Join(flPath, "random.zip"), EmptyString); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	w := zip.NewWriter(newFile)
	for _, file := range []string{"file.txt"} {
		f, err := w.Create(file)
		if err != nil {
			t.Error(err)
		}
		f.Write([]byte("noMessage"))
	}

	w.Close()

	newFile.Close()

	expectedErr = "open /tmp/randomMessage/file.txt: no such file or directory"
	if err := Unzip(path.Join(flPath, "random.zip"), "/tmp/randomMessage"); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}

	if err = os.Remove(path.Join(flPath, "random.zip")); err != nil {
		t.Fatal(err)
	}

	if err = os.RemoveAll(flPath); err != nil {
		t.Fatal(err)
	}
}

func testUnzipADirectory(t *testing.T) {
	flPath := "/tmp/testUnzip"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	newFile, err := os.Create(path.Join(flPath, "random.zip"))
	if err != nil {
		t.Error(err)
	}

	w := zip.NewWriter(newFile)

	for _, file := range []string{"file/", "file.txt"} {
		f, err := w.Create(file)
		if err != nil {
			t.Error(err)
		}
		f.Write([]byte(`noMessage`))
	}

	w.Close()

	newFile.Close()

	if err := Unzip(path.Join(flPath, "random.zip"), flPath); err != nil {
		t.Error(err)
	}
	if err = os.Remove(path.Join(flPath, "random.zip")); err != nil {
		t.Fatal(err)
	}

	if err = os.RemoveAll(flPath); err != nil {
		t.Fatal(err)
	}
}

type zipFileTest struct{}

func (zipFileTest) Open() (io.ReadCloser, error) {
	return nil, fmt.Errorf("Cannot open the file")
}

func testUnzipOpenFileError(t *testing.T) {
	expectdErr := "Cannot open the file"
	if err := unzipFile(new(zipFileTest), EmptyString, 0); err == nil || err.Error() != expectdErr {
		t.Errorf("Expected %+v, received %+v", expectdErr, err)
	}
}
