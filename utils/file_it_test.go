//go:build integration
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
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/fsnotify/fsnotify"
)

var (
	testsWatchDir = []func(t *testing.T){
		testWatchWatcherError,
		testWatchWatcherEvents,
		testWatchDirValidPath,
		testWatchDirInvalidPath,
		testWatchNewWatcherError,
	}
)

func TestFileIT(t *testing.T) {
	for _, test := range testsWatchDir {
		t.Run("Watch_Dir_Tests", test)
	}
}

func testWatchWatcherError(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	watcher.Events = make(chan fsnotify.Event, 1)
	watcher.Errors = make(chan error, 1)

	chanErr := fmt.Errorf("")
	watcher.Errors <- chanErr
	stopWatching := make(chan struct{}, 1)
	if err := watch(EmptyString, EmptyString, nil, watcher, stopWatching); err != chanErr {
		t.Error(err)
	}
}

func testWatchWatcherEvents(t *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	watcher.Events = make(chan fsnotify.Event, 1)
	watcher.Errors = make(chan error, 1)

	watcher.Events <- fsnotify.Event{
		Name: "/tmp/file.txt",
		Op:   fsnotify.Create,
	}
	stopWatching := make(chan struct{}, 1)
	f := func(itmPath, itmID string) error {
		close(stopWatching)
		if itmPath != "/tmp" || itmID != "file.txt" {
			t.Errorf("Invalid directory or file")
		}
		return fmt.Errorf("Can't match path")
	}
	expected := "Can't match path"
	if err := watch(EmptyString, EmptyString, f, watcher, stopWatching); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testWatchDirValidPath(t *testing.T) {
	flPath := "/tmp/testWatchDirValidPath/"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}

	newFile, err := os.Create(path.Join(flPath, "file.txt"))
	if err != nil {
		t.Error(err)
	}
	newFile.Close()

	stopWatching := make(chan struct{}, 1)
	if err := WatchDir(path.Join(flPath, "file.txt"), nil, EmptyString, stopWatching); err != nil {
		t.Error(err)
	}

	if err := os.Remove(path.Join(flPath, "file.txt")); err != nil {
		t.Error(err)
	}
	if err := os.Remove(flPath); err != nil {
		t.Error(err)
	}
}

func testWatchDirInvalidPath(t *testing.T) {
	flPath := "tmp/testWatchDirInvalidPath"
	stopWatching := make(chan struct{}, 1)
	expected := "no such file or directory"
	if err := WatchDir(flPath, nil, EmptyString, stopWatching); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testWatchNewWatcherError(t *testing.T) {
	newWatcher := func() (*fsnotify.Watcher, error) {
		return nil, fmt.Errorf("Invalid watcher")
	}
	stopWatching := make(chan struct{}, 1)
	expected := "Invalid watcher"
	if err := watchDir(EmptyString, nil, EmptyString, stopWatching, newWatcher); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}
