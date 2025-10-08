/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package loaders

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNopLocker(t *testing.T) {
	np := newLocker(utils.EmptyString, utils.EmptyString)
	if err := np.Lock(); err != nil {
		t.Error(err)
	}
	exp := new(nopLock)
	if !reflect.DeepEqual(np, exp) {
		t.Errorf("Expeceted: %+v, received: %+v", exp, np)
	}
	if lk, err := np.Locked(); err != nil {
		t.Error(err)
	} else if lk {
		t.Error("Expected no lock")
	}
	if err := np.Unlock(); err != nil {
		t.Error(err)
	}
	if np.IsLockFile(utils.EmptyString) {
		t.Error("Expected to not be lock file")
	}
}

func TestFolderLocker(t *testing.T) {
	dir, err := os.MkdirTemp(utils.EmptyString, "TestFolderLocker")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fp := path.Join(dir, ".lkr")
	np := newLocker(fp, utils.EmptyString)
	exp := folderLock(fp)
	if !reflect.DeepEqual(np, exp) {
		t.Errorf("Expeceted: %+v, received: %+v", exp, np)
	}
	if err := np.Lock(); err != nil {
		t.Error(err)
	}
	if lk, err := np.Locked(); err != nil {
		t.Error(err)
	} else if !lk {
		t.Error("Expected lock")
	}
	if err := np.Unlock(); err != nil {
		t.Error(err)
	}
	if np.IsLockFile(utils.EmptyString) {
		t.Error("Expected to not be lock file")
	}
	if !np.IsLockFile(fp) {
		t.Error("Expected to be lock file")
	}
	if lk, err := np.Locked(); err != nil {
		t.Error(err)
	} else if lk {
		t.Error("Expected no lock")
	}
}

func TestMemoryLocker(t *testing.T) {
	np := newLocker(utils.MetaMemory, "ID")
	exp := &memoryLock{loaderID: "ID"}
	if !reflect.DeepEqual(np, exp) {
		t.Errorf("Expeceted: %+v, received: %+v", exp, np)
	}
	if err := np.Lock(); err != nil {
		t.Error(err)
	}
	if lk, err := np.Locked(); err != nil {
		t.Error(err)
	} else if !lk {
		t.Error("Expected lock")
	}
	if err := np.Unlock(); err != nil {
		t.Error(err)
	}
	if np.IsLockFile(utils.EmptyString) {
		t.Error("Expected to not be lock file")
	}
	if lk, err := np.Locked(); err != nil {
		t.Error(err)
	} else if lk {
		t.Error("Expected no lock")
	}
}
