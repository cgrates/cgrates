// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package loaders

import (
	"os"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

type loaderLocker struct {
	path     string
	loaderID string
	memory   *guardian.Locker
}

func newLoaderLocker(path, loaderID string, memory *guardian.Locker) loaderLocker {
	locker := loaderLocker{path: path}
	if path == utils.MetaMemory {
		locker.loaderID = loaderID
		locker.memory = memory
	}
	return locker
}

func (l loaderLocker) lock() (func(), error) {
	switch l.path {
	case "":
		return func() {}, nil
	case utils.MetaMemory:
		return l.memory.Lock(utils.ConcatenatedKey(utils.LoaderS, l.loaderID)), nil
	}
	file, err := os.OpenFile(l.path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	_ = file.Close()
	return func() { _ = os.Remove(l.path) }, nil
}

func (l loaderLocker) forceUnlock() error {
	if l.path == "" || l.path == utils.MetaMemory {
		return nil
	}
	return os.Remove(l.path)
}

func (l loaderLocker) locked() (bool, error) {
	if l.path == "" || l.path == utils.MetaMemory {
		return false, nil
	}
	if _, err := os.Stat(l.path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l loaderLocker) isLockFile(path string) bool {
	return l.path != "" && l.path != utils.MetaMemory && path == l.path
}
