//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
