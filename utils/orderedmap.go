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

// OrderedMap is a map that maintains the order of its key-value pairs.
type OrderedMap[K comparable, V any] struct {
	keys   []K     // keys holds the keys in order of their insertion.
	values map[K]V // values is a map of key-value pairs.
}

// NewOrderedMap creates a new ordered map and returns a pointer to it.
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		keys:   make([]K, 0),  // Initialize an empty slice for keys.
		values: make(map[K]V), // Initialize an empty map for key-value pairs.
	}
}

// Set adds a new key-value pair to the ordered map. If the key already exists, it updates the value.
func (om *OrderedMap[K, V]) Set(key K, value V) {
	// If the key does not exist in the map, append it to the keys slice.
	if _, exists := om.values[key]; !exists {
		om.keys = append(om.keys, key)
	}
	// Add or update the value for the key in the map.
	om.values[key] = value
}

// Get retrieves the value associated with the given key from the ordered map.
// It returns the value and a boolean indicating whether the key exists in the map.
func (om *OrderedMap[K, V]) Get(key K) (V, bool) {
	// Retrieve the value for the key from the map.
	val, ok := om.values[key]
	return val, ok
}

// Delete removes the key-value pair associated with the given key from the ordered map.
func (om *OrderedMap[K, V]) Delete(key K) {
	// Iterate over the keys slice to find the key to delete.
	for i, k := range om.keys {
		// When the key is found, remove it from the slice.
		if k == key {
			om.keys = append(om.keys[:i], om.keys[i+1:]...)
			break
		}
	}
	// Remove the key-value pair from the map.
	delete(om.values, key)
}

// Keys returns all keys of the ordered map in order of their insertion.
func (om *OrderedMap[K, V]) Keys() []K {
	return om.keys
}

// Values returns all values of the ordered map in the order of their corresponding keys' insertion.
func (om *OrderedMap[K, V]) Values() []V {
	// Initialize an empty slice to hold the values.
	vals := make([]V, 0, len(om.values))

	// Iterate over the keys in order and append the corresponding value to the values slice.
	for _, key := range om.keys {
		vals = append(vals, om.values[key])
	}
	return vals
}

// Map returns a deep copy of the ordered map's key-value pairs.
func (om *OrderedMap[K, V]) Map() map[K]V {
	mp := make(map[K]V, len(om.values))
	for key, value := range om.values {
		mp[key] = value
	}
	return mp
}

// GetByIndex returns the key-value pair at the specified index within the ordered map.
// If the index is out of bounds, the zero values for K and V are returned with a false flag.
func (om *OrderedMap[K, V]) GetByIndex(index int) (K, V, bool) {
	if index < 0 || index >= len(om.keys) {
		var zeroK K
		var zeroV V
		return zeroK, zeroV, false
	}
	key := om.keys[index]
	value := om.values[key]
	return key, value, true
}
