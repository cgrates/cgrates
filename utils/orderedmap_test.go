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
package utils

import (
	"crypto/rand"
	"reflect"
	"testing"
)

func TestOrderedMapSetGetDelete(t *testing.T) {
	testCases := []struct {
		key   string
		value string
	}{
		{"shortKey", "shortValue"},
		{"longKeylongKeylongKeylongKeylongKeylongKey", "longValuelongValuelongValuelongValuelongValue"},
		{"keyWithSpecialCharacters!@#$%^&*()", "valueWithSpecialCharacters!@#$%^&*()"},
		{"", ""},
	}

	for _, tc := range testCases {
		// Initialize a new OrderedMap for each test case
		om := NewOrderedMap[string, string]()

		// Perform a set and a get for the initial key-value pair
		om.Set(tc.key, tc.value)
		val, ok := om.Get(tc.key)
		if !ok || val != tc.value {
			t.Errorf("Set key-value pair did not match Get result: expected %s, got %s", tc.value, val)
		}

		// Update the value
		newValue := tc.value + "updated"
		om.Set(tc.key, newValue)
		val, ok = om.Get(tc.key)
		if !ok || val != newValue {
			t.Errorf("Updated key-value pair did not match Get result: expected %s, got %s", newValue, val)
		}

		// Delete the key
		om.Delete(tc.key)
		_, ok = om.Get(tc.key)
		if ok {
			t.Errorf("Deleted key was still found in map")
		}

		// Try to get a non-existent key
		_, ok = om.Get("non-existent key")
		if ok {
			t.Errorf("Non-existent key was found in map")
		}
	}
}

func TestOrderedMapKeysValues(t *testing.T) {
	// Initialize a new OrderedMap
	om := NewOrderedMap[string, string]()

	// Set multiple key-value pairs
	om.Set("key1", "value1")
	om.Set("key2", "value2")
	om.Set("key3", "value3")

	// Get the keys and values
	keys := om.Keys()
	values := om.Values()

	// Check the keys
	expectedKeys := []string{"key1", "key2", "key3"}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("Keys do not match expected keys: expected %v, received %v", expectedKeys, keys)
	}

	// Check the values
	expectedValues := []string{"value1", "value2", "value3"}
	if !reflect.DeepEqual(values, expectedValues) {
		t.Errorf("Values do not match expected values: expected %v, received %v", expectedValues, values)
	}
}

// Benchmark that emphasizes the difference in performance for simple values between
// OrderedNavigableMap and the generic OrderedMap.
// Sample usage: go test -bench=. -run=Benchmark_NavigableMaps -benchtime=5s -count 3 -benchmem
func BenchmarkOrderedMaps(b *testing.B) {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	// Function to generate a random string of `strLen` length
	randomString := func(strLen int) string {
		byt := make([]byte, strLen)
		rand.Read(byt)

		for i := range byt {
			byt[i] = letterBytes[byt[i]%byte(len(letterBytes))]
		}
		return string(byt)
	}

	// Function to generate a slice of n random strings
	generateStringSlice := func(n int) []string {
		randStrings := make([]string, n)
		for i := range randStrings {
			randStrings[i] = randomString(10)
		}
		return randStrings
	}

	keys := generateStringSlice(100000)
	values := generateStringSlice(100000)

	b.Run("Generic ordered map - Set", func(b *testing.B) {
		genericOm := NewOrderedMap[string, string]()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			genericOm.Set(keys[i%len(keys)], values[i%len(values)])
		}
	})
	b.Run("Generic ordered map - Get", func(b *testing.B) {
		genericOm := NewOrderedMap[string, string]()
		for i := 0; i < len(keys); i++ {
			genericOm.Set(keys[i], values[i])
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = genericOm.Get(keys[i%len(keys)])
		}
	})

	b.Run("OrderedNavigableMap - Set", func(b *testing.B) {
		onm := NewOrderedNavigableMap()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			onm.Set(&FullPath{
				Path:      keys[i%len(keys)],
				PathItems: NewPathItems([]string{keys[i%len(keys)]}),
			}, NewNMData(values[i%len(values)]))
		}
	})
	b.Run("OrderedNavigableMap - Field+String", func(b *testing.B) {
		onm := NewOrderedNavigableMap()
		for i := 0; i < len(keys); i++ {
			onm.Set(&FullPath{
				Path:      keys[i],
				PathItems: NewPathItems([]string{keys[i]}),
			}, NewNMData(values[i%len(values)]))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			val, _ := onm.Field(NewPathItems([]string{keys[i%len(keys)]}))
			_ = val.String()
		}
	})

}
