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
	"os"
	"testing"
)

var (
	testsWatchDir = []func(t *testing.T){
		testWatchDir,
		testWatchDirNoError,
	}
)

func TestFileIT(t *testing.T) {
	for _, test := range testsWatchDir {
		t.Run("Watch_Dir_Tests", test)
	}
}

func testWatchDir(t *testing.T) {
	stopWatching := make(chan struct{}, 1)
	close(stopWatching)
	flPath := "/tmp/testWatchDir"
	if err := os.MkdirAll(flPath, 0777); err != nil {
		t.Error(err)
	}
	if err := WatchDir(flPath, nil, "randomID", stopWatching); err != nil {
		t.Error(err)
	}

	if err := os.RemoveAll(flPath); err != nil {
		t.Fatal(err)
	}
}

func testWatchDirNoError(t *testing.T) {
	stopWatching := make(chan struct{}, 1)
	flPath := "/tmp/inexistentDir"
	expectedErr := "no such file or directory"
	if err := WatchDir(flPath, nil, "randomID", stopWatching); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}
