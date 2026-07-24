// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
