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
	"strconv"
	"strings"
)

type StringMap map[string]bool

func NewStringMap(s ...string) StringMap {
	result := make(StringMap)
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != EmptyString {
			if strings.HasPrefix(v, NegativePrefix) {
				result[v[1:]] = false
			} else {
				result[v] = true
			}
		}
	}
	return result
}

func ParseStringMap(s string) StringMap {
	if s == MetaZero {
		return make(StringMap)
	}
	return StringMapFromSlice(strings.Split(s, InfieldSep))
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
		sm[MetaAny]
}

func StringMapFromSlice(s []string) StringMap {
	result := make(StringMap, len(s))
	for _, v := range s {
		v = strings.TrimSpace(v)
		if v != EmptyString {
			if strings.HasPrefix(v, NegativePrefix) {
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
	return strings.Join(sm.Slice(), InfieldSep)
}

func (sm StringMap) GetOne() string {
	for key := range sm {
		return key
	}
	return EmptyString
}

func (sm StringMap) HasKey(key string) (has bool) {
	_, has = sm[key]
	return
}

func MapStringToInt64(in map[string]string) (out map[string]int64, err error) {
	mapout := make(map[string]int64, len(in))
	for key, val := range in {
		x, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		mapout[key] = int64(x)
	}
	return mapout, nil
}

// FlagsWithParamsFromSlice construct a  FlagsWithParams from the given slice
func FlagsWithParamsFromSlice(s []string) (flags FlagsWithParams) {
	flags = make(FlagsWithParams)
	for _, v := range s {
		flag := strings.SplitN(v, InInFieldSep, 3)
		if !flags.Has(flag[0]) {
			flags[flag[0]] = make(FlagParams)
		}
		flags[flag[0]].Add(flag[1:])
	}
	return
}

// FlagParams stores the parameters for a flag
type FlagParams map[string][]string

// Has returns if the key was mentioned in flags
func (fWp FlagParams) Has(opt string) (has bool) {
	_, has = fWp[opt]
	return
}

// ParamValue returns the value of the flag
func (fWp FlagParams) ParamValue(opt string) (ps string) {
	for _, ps = range fWp[opt] {
		if ps != EmptyString {
			return
		}
	}
	return
}

// Add adds the options to the flag
func (fWp FlagParams) Add(opts []string) {
	switch len(opts) {
	default: // just in case we call this function with more elements than needed
		fallthrough
	case 2:
		fWp[opts[0]] = strings.Split(opts[1], ANDSep)
	case 0:
	case 1:
		fWp[opts[0]] = []string{}
	}
}

// ParamsSlice returns the list of profiles for the subsystem
func (fWp FlagParams) ParamsSlice(opt string) (ps []string) {
	return fWp[opt] // if it doesn't have the option it will return an empty slice
}

// SliceFlags converts from FlagsParams to []string
func (fWp FlagParams) SliceFlags() (sls []string) {
	for key, sub := range fWp {
		if len(sub) == 0 { // no option for these subsystem
			sls = append(sls, key)
			continue
		}
		sls = append(sls, ConcatenatedKey(key, strings.Join(sub, InfieldSep)))
	}
	return
}

// Clone returns a deep copy of FlagParams
func (fWp FlagParams) Clone() (cln FlagParams) {
	if fWp == nil {
		return
	}
	cln = make(FlagParams)
	for flg, params := range fWp {
		var cprm []string
		if params != nil {
			cprm = CloneStringSlice(params)
		}
		cln[flg] = cprm
	}
	return
}

// FlagsWithParams should store a list of flags for each subsystem
type FlagsWithParams map[string]FlagParams

// Has returns if the key was mentioned in flags
func (fWp FlagsWithParams) Has(flag string) (has bool) {
	_, has = fWp[flag]
	return
}

// ParamsSlice returns the list of profiles for the subsystem
func (fWp FlagsWithParams) ParamsSlice(subs, opt string) (ps []string) {
	if psIfc, has := fWp[subs]; has {
		ps = psIfc.ParamsSlice(opt)
	}
	return
}

// ParamValue returns the value of the flag
func (fWp FlagsWithParams) ParamValue(subs string) (ps string) {
	for ps = range fWp[subs] {
		return
	}
	return
}

// SliceFlags converts from FlagsWithParams back to []string
func (fWp FlagsWithParams) SliceFlags() (sls []string) {
	for key, sub := range fWp {
		if len(sub) == 0 { // no option for these subsystem
			sls = append(sls, key)
			continue
		}
		for opt, values := range sub {
			if len(values) == 0 { // it's an option without values(e.g *derived_reply)
				sls = append(sls, ConcatenatedKey(key, opt))
				continue
			}
			sls = append(sls, ConcatenatedKey(key, opt, strings.Join(values, ANDSep)))
		}
	}
	return
}

// GetBool returns the flag as boolean
func (fWp FlagsWithParams) GetBool(key string) (b bool) {
	var v FlagParams
	if v, b = fWp[key]; !b {
		return // not present means false
	}
	if v == nil || len(v) == 0 {
		return true // empty map
	}
	return v.Has(TrueStr) || !v.Has(FalseStr)
}

// Clone returns a deep copy of FlagsWithParams
func (fWp FlagsWithParams) Clone() (cln FlagsWithParams) {
	if fWp == nil {
		return
	}
	cln = make(FlagsWithParams)
	for flg, p := range fWp {
		cln[flg] = p.Clone()
	}
	return
}

func (sm StringMap) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if sm == nil || len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	bl, has := sm[fldPath[0]]
	if !has {
		return nil, ErrNotFound
	}
	return bl, nil
}

func (sm StringMap) FieldAsString(fldPath []string) (val string, err error) {
	var iface interface{}
	iface, err = sm.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	return IfaceAsString(iface), nil
}
