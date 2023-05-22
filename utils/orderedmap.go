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

// OrderedMap is a map that maintains the order of keys.
type OrderedMap struct {
	keys   []string          // keys stores the order of keys.
	values map[string]string // values is a map from keys to their associated values.
}

// NewOrderedMap creates a new, empty OrderedMap.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		keys:   make([]string, 0),       // Initialize keys as an empty slice.
		values: make(map[string]string), // Initialize values as an empty map.
	}
}

// Set adds a key-value pair to the OrderedMap. If the key already exists, its value is updated.
func (om *OrderedMap) Set(key string, value string) {
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key) // Add the key to keys if it's new.
	}
	om.values[key] = value // Add or update the value in values.
}

// Get retrieves the value associated with a key in the OrderedMap.
// It returns the value and a boolean indicating whether the key was found.
func (om *OrderedMap) Get(key string) (string, bool) {
	value, exists := om.values[key]
	return value, exists
}

// Delete removes a key-value pair from the OrderedMap.
func (om *OrderedMap) Delete(key string) {
	delete(om.values, key) // Remove the key-value pair from values.
	for i, k := range om.keys {
		if k == key {
			om.keys = append(om.keys[:i], om.keys[i+1:]...) // Remove the key from keys.
			break
		}
	}
}

// Keys returns a slice of all keys in the OrderedMap, in their original order.
func (om *OrderedMap) Keys() []string {
	return om.keys
}

// Slice returns a slice of all values in the OrderedMap, in the order of their keys.
func (om *OrderedMap) Slice() []string {
	out := make([]string, 0, len(om.keys))
	for _, key := range om.keys {
		out = append(out, om.values[key]) // Add each value to out in order of its key.
	}
	return out
}
