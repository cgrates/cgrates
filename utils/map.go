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
func MirrorMap(mapIn map[string]string) map[string]string {
	mapOut := make(map[string]string, len(mapIn))
	for key, val := range mapIn {
		mapOut[val] = key
	}
	return mapOut
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
			if strings.HasPrefix(v, "!") {
				result[v[1:]] = false
			} else {
				result[v] = true
			}
		}
	}
	return result
}

func ParseStringMap(s string) StringMap {
	if s == ZERO {
		return make(StringMap)
	}
	return StringMapFromSlice(strings.Split(s, INFIELD_SEP))
}

func (sm StringMap) Equal(om StringMap) bool {
	if sm == nil && om != nil {
		return false
	}
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

func (sm StringMap) Slice() []string {
	result := make([]string, len(sm))
	i := 0
	for k := range sm {
		result[i] = k
		i++
	}
	return result
}

func (sm StringMap) IsEmpty() bool {
	return sm == nil ||
		len(sm) == 0 ||
		sm[ANY] == true
}

func StringMapFromSlice(s []string) StringMap {
	result := make(StringMap, len(s))
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != "" {
			if strings.HasPrefix(v, "!") {
				result[v[1:]] = false
			} else {
				result[v] = true
			}
		}
	}
	return result
}

func (sm StringMap) Copy(o StringMap) {
	for k, v := range o {
		sm[k] = v
	}
}

func (sm StringMap) Clone() StringMap {
	result := make(StringMap, len(sm))
	result.Copy(sm)
	return result
}

func (sm StringMap) String() string {
	return strings.Join(sm.Slice(), INFIELD_SEP)
}

func (sm StringMap) GetOne() string {
	for key := range sm {
		return key
	}
	return ""
}

/*
func NoDots(m map[string]struct{}) map[string]struct{} {
	return MapKeysReplace(m, ".", "．")
}

func YesDots(m map[string]struct{}) map[string]struct{} {
	return MapKeysReplace(m, "．", ".")
}

func MapKeysReplace(m map[string]struct{}, old, new string) map[string]struct{} {
	for key, val := range m {
		delete(m, key)
		key = strings.Replace(key, old, new, -1)
		m[key] = val
	}
	return m
}
*/

// Used to merge multiple maps (eg: output of struct having ExtraFields)
func MergeMapsStringIface(mps ...map[string]interface{}) (outMp map[string]interface{}) {
	outMp = make(map[string]interface{})
	for i, mp := range mps {
		if i == 0 {
			outMp = mp
			continue
		}
		for k, v := range mp {
			outMp[k] = v
		}
	}
	return
}
