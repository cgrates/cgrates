// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
