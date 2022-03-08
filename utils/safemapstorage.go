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
	"sync"
)

// SafeMapStorage is a Mapstorage with mutex
type SafeMapStorage struct {
	MapStorage
	sync.RWMutex
}

// String returns the map as json string
func (ms *SafeMapStorage) String() string {
	ms.RLock()
	defer ms.RUnlock()
	return ms.MapStorage.String()
}

// FieldAsInterface returns the value from the path
func (ms *SafeMapStorage) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	ms.RLock()
	defer ms.RUnlock()
	return ms.MapStorage.FieldAsInterface(fldPath)
}

// FieldAsString returns the value from path as string
func (ms *SafeMapStorage) FieldAsString(fldPath []string) (str string, err error) {
	ms.RLock()
	defer ms.RUnlock()
	return ms.MapStorage.FieldAsString(fldPath)
}

// Set sets the value at the given path
func (ms *SafeMapStorage) Set(fldPath []string, val interface{}) (err error) {
	ms.Lock()
	defer ms.Unlock()
	return ms.MapStorage.Set(fldPath, val)
}

// GetKeys returns all the keys from map
func (ms *SafeMapStorage) GetKeys(nested bool, nestedLimit int, prefix string) (keys []string) {
	ms.RLock()
	defer ms.RUnlock()
	return ms.MapStorage.GetKeys(nested, nestedLimit, prefix)
}

// Remove removes the item at path
func (ms *SafeMapStorage) Remove(fldPath []string) (err error) {
	ms.Lock()
	defer ms.Unlock()
	return ms.MapStorage.Remove(fldPath)
}

func (ms *SafeMapStorage) Clone() (msClone *SafeMapStorage) {
	ms.RLock()
	defer ms.RUnlock()
	return &SafeMapStorage{MapStorage: ms.MapStorage.Clone()}
}

func (ms *SafeMapStorage) ClonedMapStorage() (msClone MapStorage) {
	ms.RLock()
	defer ms.RUnlock()
	return ms.MapStorage.Clone()
}
