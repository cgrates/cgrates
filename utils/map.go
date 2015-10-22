/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

import "strings"

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
		if val, hasKey := inMap[reqKey]; !hasKey || val == "" {
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

type StringMap map[string]bool

func NewStringMap(s ...string) StringMap {
	result := make(StringMap)
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != "" {
			result[v] = true
		}
	}
	return result
}

func ParseStringMap(s string) StringMap {
	slice := strings.Split(s, INFIELD_SEP)
	result := make(StringMap)
	for _, v := range slice {
		v = strings.TrimSpace(v)
		if v != "" {
			result[v] = true
		}
	}
	return result
}

func (sm StringMap) Equal(om StringMap) bool {
	if len(sm) != len(om) {
		return false
	}
	for key := range sm {
		if !om[key] {
			return false
		}
	}
	return true
}

func (sm StringMap) Includes(om StringMap) bool {
	if len(sm) < len(om) {
		return false
	}
	for key := range om {
		if !sm[key] {
			return false
		}
	}
	return true
}

func (sm StringMap) Clone() StringMap {
	result := make(StringMap, len(sm))
	for k := range sm {
		result[k] = true
	}
	return result
}

func (sm StringMap) Slice() []string {
	result := make([]string, len(sm))
	i := 0
	for k := range sm {
		result[i] = k
		i++
	}
	return result
}

func (sm StringMap) String() string {
	return strings.Join(sm.Slice(), INFIELD_SEP)
}
