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
	"strings"
)

// DataStorage is the new DataProvider
type DataStorage interface {
	String() string // printable version of data
	Get(fldPath []string) (interface{}, error)
	GetString(fldPath []string) (string, error)
	Set(fldPath []string, val interface{}) error
	Remove(fldPath []string) error
	GetKeys(nesteed bool) []string
	// RemoteHost() net.Addr
}

// MapStorage is the basic DataStorage
type MapStorage map[string]interface{}

// String returns the map as json string
func (ms *MapStorage) String() string { return ToJSON(ms) }

// Get returns the value from the path
func (ms *MapStorage) Get(fldPath []string) (val interface{}, err error) {
	if len(fldPath) == 0 {
		err = ErrNotFound
		return
	}
	var has bool
	if val, has = (*ms)[fldPath[0]]; !has {
		err = ErrNotFound
		return
	}
	if len(fldPath) == 1 {
		return
	}
	ds, ok := val.(DataStorage)
	if !ok {
		err = fmt.Errorf("Wrong type")
		return
	}
	return ds.Get(fldPath[1:])
}

// GetString returns thevalue from path as string
func (ms *MapStorage) GetString(fldPath []string) (str string, err error) {
	var val interface{}
	if val, err = ms.Get(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// Set sets the value at the given path
func (ms *MapStorage) Set(fldPath []string, val interface{}) (err error) {
	if len(fldPath) == 0 {
		return fmt.Errorf("Wrong path")
	}
	if len(fldPath) == 1 {
		(*ms)[fldPath[0]] = val
		return
	}
	nMap := &MapStorage{}
	(*ms)[fldPath[0]] = nMap
	return nMap.Set(fldPath[1:], val)
}

// GetKeys returns all the keys from map
func (ms *MapStorage) GetKeys(nesteed bool) (keys []string) {
	for k, v := range *ms {
		keys = append(keys, k)
		if !nesteed {
			continue
		}
		ds, ok := v.(DataStorage)
		if !ok {
			continue
		}
		for _, dsKey := range ds.GetKeys(nesteed) {
			keys = append(keys, k+NestingSep+dsKey)
		}
	}
	return
}

// Remove removes the item at path
func (ms *MapStorage) Remove(fldPath []string) (err error) {
	if len(fldPath) == 0 {
		return fmt.Errorf("Wrong path")
	}
	var val interface{}
	var has bool
	if val, has = (*ms)[fldPath[0]]; !has {
		return // ignore (already removed)
	}
	if len(fldPath) == 1 {
		delete(*ms, fldPath[0])
		return
	}
	ds, ok := val.(DataStorage)
	if !ok {
		err = fmt.Errorf("Wrong type")
		return
	}
	return ds.Remove(fldPath[1:])
}

// NavigableMap is a DataStorage
type NavigableMap map[string]DataStorage

// String returns the map as json string
func (nm *NavigableMap) String() string { return ToJSON(nm) }

// Get returns the value from the path
func (nm *NavigableMap) Get(fldPath []string) (val interface{}, err error) {
	if len(fldPath) == 0 {
		err = ErrNotFound
		return
	}
	ds, has := (*nm)[fldPath[0]]
	if !has {
		err = ErrNotFound
		return
	}
	if len(fldPath) == 1 {
		val = ds
		return
	}
	return ds.Get(fldPath[1:])
}

// GetString returns thevalue from path as string
func (nm *NavigableMap) GetString(fldPath []string) (str string, err error) {
	var val interface{}
	if val, err = nm.Get(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// Set sets the value at the given path
func (nm *NavigableMap) Set(fldPath []string, val interface{}) (err error) {
	if len(fldPath) == 0 {
		return fmt.Errorf("Wrong path")
	}
	if len(fldPath) == 1 {
		ds, ok := val.(DataStorage)
		if !ok {
			return fmt.Errorf("Wrong type")
		}
		(*nm)[fldPath[0]] = ds
		return
	}
	if _, has := (*nm)[fldPath[0]]; !has {
		(*nm)[fldPath[0]] = &MapStorage{}
	}
	return (*nm)[fldPath[0]].Set(fldPath[1:], val)
}

// GetKeys returns all the keys from map
func (nm *NavigableMap) GetKeys(nesteed bool) (keys []string) {
	for k, v := range *nm {
		keys = append(keys, k)
		if !nesteed {
			continue
		}
		for _, dsKey := range v.GetKeys(nesteed) {
			keys = append(keys, k+NestingSep+dsKey)
		}
	}
	return
}

// Remove removes the item at path
func (nm *NavigableMap) Remove(fldPath []string) (err error) {
	if len(fldPath) == 0 {
		return fmt.Errorf("Wrong path")
	}
	var val DataStorage
	var has bool
	if val, has = (*nm)[fldPath[0]]; !has {
		return // ignore (already removed)
	}
	if len(fldPath) == 1 {
		delete(*nm, fldPath[0])
		return
	}
	return val.Remove(fldPath[1:])
}

// NewOrderedNavigableMap initializates a structure of OrderedNavigableMap with a NavigableMap
func NewOrderedNavigableMap(nm *NavigableMap) *OrderedNavigableMap {
	if nm == nil {
		return &OrderedNavigableMap{
			nm:    &NavigableMap{},
			order: [][]string{},
		}
	}
	keys := nm.GetKeys(true)
	order := make([][]string, len(keys))
	for i, k := range keys {
		order[i] = strings.Split(k, NestingSep)
	}
	return &OrderedNavigableMap{
		nm:    nm,
		order: order,
	}
}

// OrderedNavigableMap is the same as NavigableMap but keeps the order of fields
type OrderedNavigableMap struct {
	nm    *NavigableMap
	order [][]string
}

// String returns the map as json string
func (onm *OrderedNavigableMap) String() string { return ToJSON(onm.nm) }

// Get returns the value from the path
func (onm *OrderedNavigableMap) Get(fldPath []string) (val interface{}, err error) {
	return onm.nm.Get(fldPath)
}

// GetString returns thevalue from path as string
func (onm *OrderedNavigableMap) GetString(fldPath []string) (str string, err error) {
	return onm.nm.GetString(fldPath)
}

// Set sets the value at the given path
func (onm *OrderedNavigableMap) Set(fldPath []string, val interface{}) (err error) {
	if err = onm.nm.Set(fldPath, val); err == nil {
		onm.order = append(onm.order, fldPath)
	}
	return
}

// GetKeys returns all the keys from map
func (onm *OrderedNavigableMap) GetKeys(nesteed bool) (keys []string) {
	keys = make([]string, len(onm.order))
	for i, k := range onm.order {
		keys[i] = strings.Join(k, NestingSep)
	}
	return
}

// Remove removes the item at path
func (onm *OrderedNavigableMap) Remove(fldPath []string) (err error) {
	if len(fldPath) == 0 {
		return fmt.Errorf("Wrong path")
	}
	return onm.nm.Remove(fldPath)
}
