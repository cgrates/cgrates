//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
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
	Logger.SetLogLevel(4)
	Logger.SetSyslog(nil)
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	t.Cleanup(func() {
		Logger.SetLogLevel(0)
		log.SetOutput(os.Stderr)
	})
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
	f := func(itmID string) error {
		close(stopWatching)
		if itmID != "file.txt" {
			t.Errorf("Invalid directory or file")
		}
		return fmt.Errorf("Can't match path")
	}
	if err := watch(EmptyString, EmptyString, f, watcher, stopWatching); err != nil {
		t.Error(err)
	}
	expLog := `[WARNING] <> processing path </tmp/file.txt>, error: <Can't match path>`
	if !strings.Contains(buf.String(), expLog) {
		t.Errorf("expected %q to be present, received:\n%s", expLog, buf.String())
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
