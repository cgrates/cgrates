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
	"slices"
	"strconv"
	"strings"
)

func MapStringToInt64(in map[string]string) (mapout map[string]int64, err error) {
	mapout = make(map[string]int64, len(in))
	for key, val := range in {
		mapout[key], err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return
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
func (fp FlagParams) Has(opt string) bool {
	_, has := fp[opt]
	return has
}

// ParamValue returns the value of the flag
func (fp FlagParams) ParamValue(opt string) string {
	for _, v := range fp[opt] {
		if v != EmptyString {
			return v
		}
	}
	return ""
}

// Add adds the options to the flag
func (fp FlagParams) Add(opts []string) {
	switch len(opts) {
	default: // just in case we call this function with more elements than needed
		fallthrough
	case 2:
		fp[opts[0]] = strings.Split(opts[1], ANDSep)
	case 0:
	case 1:
		fp[opts[0]] = []string{}
	}
}

// ParamsSlice returns the list of profiles for the subsystem
func (fp FlagParams) ParamsSlice(opt string) []string {
	return fp[opt] // if it doesn't have the option it will return an empty slice
}

// SliceFlags converts from FlagsParams to []string
func (fp FlagParams) SliceFlags() (sls []string) {
	for key, sub := range fp {
		if len(sub) == 0 { // no option for these subsystem
			sls = append(sls, key)
			continue
		}
		sls = append(sls, ConcatenatedKey(key, strings.Join(sub, InfieldSep)))
	}
	return
}

// Clone returns a deep copy of FlagParams
func (fp FlagParams) Clone() FlagParams {
	if fp == nil {
		return nil
	}
	cln := make(FlagParams)
	for flg, params := range fp {
		var cprm []string
		if params != nil {
			cprm = slices.Clone(params)
		}
		cln[flg] = cprm
	}
	return cln
}

func (fp FlagParams) Equal(other FlagParams) bool {
	if len(fp) != len(other) {
		return false
	}
	for k, v := range fp {
		if otherV, exists := other[k]; !exists || !slices.Equal(v, otherV) {
			return false
		}
	}
	return true
}

// FlagsWithParams should store a list of flags for each subsystem
type FlagsWithParams map[string]FlagParams

// Has returns if the key was mentioned in flags
func (fwp FlagsWithParams) Has(flag string) bool {
	_, has := fwp[flag]
	return has
}

// ParamsSlice returns the list of profiles for the subsystem
func (fwp FlagsWithParams) ParamsSlice(subs, opt string) []string {
	if psIfc, has := fwp[subs]; has {
		return psIfc.ParamsSlice(opt)
	}
	return nil
}

// ParamValue returns the value of the flag
func (fwp FlagsWithParams) ParamValue(subs string) string {
	for v := range fwp[subs] {
		return v
	}
	return ""
}

// SliceFlags converts from FlagsWithParams back to []string
func (fwp FlagsWithParams) SliceFlags() (sls []string) {
	for key, sub := range fwp {
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
func (fwp FlagsWithParams) GetBool(key string) bool {
	v, exists := fwp[key]
	if !exists {
		return false // not present means false
	}
	if len(v) == 0 {
		return true // empty map
	}
	return v.Has(TrueStr) || !v.Has(FalseStr)
}

// Clone returns a deep copy of FlagsWithParams
func (fwp FlagsWithParams) Clone() FlagsWithParams {
	if fwp == nil {
		return nil
	}
	cln := make(FlagsWithParams)
	for flg, p := range fwp {
		cln[flg] = p.Clone()
	}
	return cln
}

func (fwp FlagsWithParams) Equal(other FlagsWithParams) bool {
	if len(fwp) != len(other) {
		return false
	}
	for k, v := range fwp {
		if otherV, exists := other[k]; !exists || !v.Equal(otherV) {
			return false
		}
	}
	return true
}

func MapStringStringEqual(v1, v2 map[string]string) bool {
	if len(v1) != len(v2) {
		return false
	}
	for k, val2 := range v2 {
		if val1, has := v1[k]; !has || val1 != val2 {
			return false
		}
	}
	return true
}
