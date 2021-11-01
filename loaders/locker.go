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
)

type Locker interface {
	Lock() error
	Unlock() error
	Locked() (bool, error)
}

func newLocker(path string) Locker {
	if path != utils.EmptyString {
		return folderLock(path)
	}
	return new(nopLock)
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

type nopLock struct{}

// lockFolder will attempt to lock the folder by creating the lock file
func (nopLock) Lock() (_ error)           { return }
func (nopLock) Unlock() (_ error)         { return }
func (nopLock) Locked() (_ bool, _ error) { return }
