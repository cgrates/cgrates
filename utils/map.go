/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

// Converts map[string]string into map[string]interface{}
func ConvertMapValStrIf(inMap map[string]string) map[string]interface{} {
	outMap := make(map[string]interface{})
	for field, val := range inMap {
		outMap[field] = val
	}
	return outMap
}

// Mirrors key/val
func MirrorMap(mapIn map[string]string) (map[string]string, error) {
	mapOut := make(map[string]string)
	for key, val := range mapIn {
		mapOut[val] = key
	}
	return mapOut, nil
}

// Returns mising keys in a map
func MissingMapKeys(inMap map[string]string, requiredKeys []string) []string {
	missingKeys := []string{}
	for _, reqKey := range requiredKeys {
		if val, hasKey := inMap[reqKey]; !hasKey {
			missingKeys = append(missingKeys, reqKey)
		} else if val == "" {
			missingKeys = append(missingKeys, reqKey)
		}
	}
	return missingKeys
}

// Return map keys
func MapKeys(m map[string]string) []string {
	n := make([]string, len(m))
	i := 0
	for k := range m {
		n[i] = k
		i++
	}
	return n
}
