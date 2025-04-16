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

package loaders

import (
	"io"
	"os"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

type Locker interface {
	Lock() error
	Unlock() error
	Locked() (bool, error)
	IsLockFile(string) bool
}

func newLocker(path, loaderID string) Locker {
	switch path {
	case utils.EmptyString:
		return new(nopLock)
	case utils.MetaMemory:
		return &memoryLock{loaderID: loaderID}
	default:
		return folderLock(path)
	}
}

type folderLock string

// lockFolder will attempt to lock the folder by creating the lock file
func (fl folderLock) Lock() (err error) {
	var file io.Closer
	file, err = os.OpenFile(string(fl),
		os.O_RDONLY|os.O_CREATE, 0644)
	file.Close()
	return
}

func (fl folderLock) Unlock() (err error) {
	return os.Remove(string(fl))
}

func (fl folderLock) Locked() (lk bool, err error) {
	if _, err = os.Stat(string(fl)); err != nil {
		if os.IsNotExist(err) {
			lk, err = false, nil
		}
		return
	}
	lk = true
	return
}
func (fl folderLock) IsLockFile(path string) bool { return path == string(fl) }

type nopLock struct{}

func (nopLock) Lock() (_ error)            { return }
func (nopLock) Unlock() (_ error)          { return }
func (nopLock) Locked() (_ bool, _ error)  { return }
func (nopLock) IsLockFile(string) (_ bool) { return }

type memoryLock struct {
	loaderID string
	refID    string
}

func (m *memoryLock) Lock() (_ error) {
	m.refID = guardian.Guardian.GuardIDs(m.refID, 0, utils.ConcatenatedKey(utils.LoaderS, m.loaderID))
	return
}
func (m *memoryLock) Unlock() (_ error) {
	guardian.Guardian.UnguardIDs(m.refID)
	m.refID = utils.EmptyString
	return
}
func (m memoryLock) Locked() (bool, error)      { return len(m.refID) != 0, nil }
func (m memoryLock) IsLockFile(string) (_ bool) { return }
