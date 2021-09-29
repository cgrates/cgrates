/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

type DynamicStringSliceOpt struct {
	Value     []string
	FilterIDs []string
}

type DynamicStringOpt struct {
	Value     string
	FilterIDs []string
}

type DynamicIntOpt struct {
	Value     int
	FilterIDs []string
}

type DynamicFloat64Opt struct {
	Value     float64
	FilterIDs []string
}

type DynamicDurationOpt struct {
	Value     time.Duration
	FilterIDs []string
}

type DynamicDecimalBigOpt struct {
	Value     *decimal.Big
	FilterIDs []string
}

func CloneDynamicStringSliceOpt(in []*DynamicStringSliceOpt) (cl []*DynamicStringSliceOpt) {
	cl = make([]*DynamicStringSliceOpt, len(in))
	copy(cl, in)
	return
}

func CloneDynamicStringOpt(in []*DynamicStringOpt) (cl []*DynamicStringOpt) {
	cl = make([]*DynamicStringOpt, len(in))
	copy(cl, in)
	return
}

func CloneDynamicIntOpt(in []*DynamicIntOpt) (cl []*DynamicIntOpt) {
	cl = make([]*DynamicIntOpt, len(in))
	copy(cl, in)
	return
}

func CloneDynamicFloat64Opt(in []*DynamicFloat64Opt) (cl []*DynamicFloat64Opt) {
	cl = make([]*DynamicFloat64Opt, len(in))
	copy(cl, in)
	return
}

func CloneDynamicDurationOpt(in []*DynamicDurationOpt) (cl []*DynamicDurationOpt) {
	cl = make([]*DynamicDurationOpt, len(in))
	copy(cl, in)
	return
}

func CloneDynamicDecimalBigOpt(in []*DynamicDecimalBigOpt) (cl []*DynamicDecimalBigOpt) {
	cl = make([]*DynamicDecimalBigOpt, len(in))
	copy(cl, in)
	return
}

func DynamicStringSliceOptEqual(v1, v2 []*DynamicStringSliceOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if !SliceStringEqual(v1[i].Value, v2[i].Value) {
			return false
		}
	}
	return true
}

func DynamicStringOptEqual(v1, v2 []*DynamicStringOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if v1[i].Value != v2[i].Value {
			return false
		}
	}
	return true
}

func DynamicIntOptEqual(v1, v2 []*DynamicIntOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if v1[i].Value != v2[i].Value {
			return false
		}
	}
	return true
}

func DynamicFloat64OptEqual(v1, v2 []*DynamicFloat64Opt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if v1[i].Value != v2[i].Value {
			return false
		}
	}
	return true
}

func DynamicDurationOptEqual(v1, v2 []*DynamicDurationOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if v1[i].Value != v2[i].Value {
			return false
		}
	}
	return true
}

func DynamicDecimalBigOptEqual(v1, v2 []*DynamicDecimalBigOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if v1[i].Value.Cmp(v2[i].Value) != 0 {
			return false
		}
	}
	return true
}

func DynamicStringSliceOptsToMap(dynOpts []*DynamicStringSliceOpt) map[string][]string {
	optMap := make(map[string][]string)
	for _, opt := range dynOpts {
		optMap[StringSliceToStringWithSep(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicStringOptsToMap(dynOpts []*DynamicStringOpt) map[string]string {
	optMap := make(map[string]string)
	for _, opt := range dynOpts {
		optMap[StringSliceToStringWithSep(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicIntOptsToMap(dynOpts []*DynamicIntOpt) map[string]int {
	optMap := make(map[string]int)
	for _, opt := range dynOpts {
		optMap[StringSliceToStringWithSep(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicFloat64OptsToMap(dynOpts []*DynamicFloat64Opt) map[string]float64 {
	optMap := make(map[string]float64)
	for _, opt := range dynOpts {
		optMap[StringSliceToStringWithSep(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicDurationOptsToMap(dynOpts []*DynamicDurationOpt) map[string]string {
	optMap := make(map[string]string)
	for _, opt := range dynOpts {
		optMap[StringSliceToStringWithSep(opt.FilterIDs, InfieldSep)] = opt.Value.String()
	}
	return optMap
}

func DynamicDecimalBigOptsToMap(dynOpts []*DynamicDecimalBigOpt) map[string]string {
	optMap := make(map[string]string)
	for _, opt := range dynOpts {
		optMap[StringSliceToStringWithSep(opt.FilterIDs, InfieldSep)] = opt.Value.String()
	}
	return optMap
}

func StringSliceToStringWithSep(ss []string, sep string) string {
	var str strings.Builder
	for i := range ss {
		str.WriteString(ss[i])
		if i != len(ss)-1 {
			str.WriteString(sep)
		}
	}
	return str.String()
}
