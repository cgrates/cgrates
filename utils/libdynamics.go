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

type DynamicBoolOpt struct {
	Value     bool
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

type DynamicInterfaceOpt struct {
	Value     interface{}
	FilterIDs []string
}

func CloneDynamicStringSliceOpt(in []*DynamicStringSliceOpt) (cl []*DynamicStringSliceOpt) {
	cl = make([]*DynamicStringSliceOpt, len(in))
	copy(cl, in)
	return
}

func CloneDynamicStringOpt(in []*DynamicStringOpt) (cl []*DynamicStringOpt) {
	cl = make([]*DynamicStringOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicStringOpt{
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     val.Value,
		}
	}
	return
}

func CloneDynamicInterfaceOpt(in []*DynamicInterfaceOpt) (cl []*DynamicInterfaceOpt) {
	cl = make([]*DynamicInterfaceOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicInterfaceOpt{
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     val.Value,
		}
	}
	return
}

func CloneDynamicBoolOpt(in []*DynamicBoolOpt) (cl []*DynamicBoolOpt) {
	cl = make([]*DynamicBoolOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicBoolOpt{
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     val.Value,
		}
	}
	return
}

func CloneDynamicIntOpt(in []*DynamicIntOpt) (cl []*DynamicIntOpt) {
	cl = make([]*DynamicIntOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicIntOpt{
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     val.Value,
		}
	}
	return
}

func CloneDynamicFloat64Opt(in []*DynamicFloat64Opt) (cl []*DynamicFloat64Opt) {
	cl = make([]*DynamicFloat64Opt, len(in))
	for i, val := range in {
		cl[i] = &DynamicFloat64Opt{
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     val.Value,
		}
	}
	return
}

func CloneDynamicDurationOpt(in []*DynamicDurationOpt) (cl []*DynamicDurationOpt) {
	cl = make([]*DynamicDurationOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicDurationOpt{
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     val.Value,
		}
	}
	return
}

func CloneDynamicDecimalBigOpt(in []*DynamicDecimalBigOpt) (cl []*DynamicDecimalBigOpt) {
	cl = make([]*DynamicDecimalBigOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicDecimalBigOpt{
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     CloneDecimalBig(val.Value),
		}
	}
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

func DynamicBoolOptEqual(v1, v2 []*DynamicBoolOpt) bool {
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
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) ||
			v1[i].Value.Cmp(v2[i].Value) != 0 {
			return false
		}
	}
	return true
}

func DynamicInterfaceOptEqual(v1, v2 []*DynamicInterfaceOpt) bool {
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

func DynamicStringSliceOptsToMap(dynOpts []*DynamicStringSliceOpt) map[string][]string {
	optMap := make(map[string][]string)
	for _, opt := range dynOpts {
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicStringOptsToMap(dynOpts []*DynamicStringOpt) map[string]string {
	optMap := make(map[string]string)
	for _, opt := range dynOpts {
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicBoolOptsToMap(dynOpts []*DynamicBoolOpt) map[string]bool {
	optMap := make(map[string]bool)
	for _, opt := range dynOpts {
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicIntOptsToMap(dynOpts []*DynamicIntOpt) map[string]int {
	optMap := make(map[string]int)
	for _, opt := range dynOpts {
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicFloat64OptsToMap(dynOpts []*DynamicFloat64Opt) map[string]float64 {
	optMap := make(map[string]float64)
	for _, opt := range dynOpts {
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func DynamicDurationOptsToMap(dynOpts []*DynamicDurationOpt) map[string]string {
	optMap := make(map[string]string)
	for _, opt := range dynOpts {
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = opt.Value.String()
	}
	return optMap
}

func DynamicDecimalBigOptsToMap(dynOpts []*DynamicDecimalBigOpt) map[string]string {
	optMap := make(map[string]string)
	for _, opt := range dynOpts {
		value := IfaceAsString(opt.Value)
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = value
	}
	return optMap
}

func DynamicInterfaceOptsToMap(dynOpts []*DynamicInterfaceOpt) map[string]interface{} {
	optMap := make(map[string]interface{})
	for _, opt := range dynOpts {
		optMap[strings.Join(opt.FilterIDs, InfieldSep)] = opt.Value
	}
	return optMap
}

func MapToDynamicStringSliceOpts(optsMap map[string][]string) (dynOpts []*DynamicStringSliceOpt) {
	dynOpts = make([]*DynamicStringSliceOpt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicStringSliceOpt{
			FilterIDs: filterIDs,
			Value:     opt,
		})
	}
	return
}

func MapToDynamicStringOpts(optsMap map[string]string) (dynOpts []*DynamicStringOpt) {
	dynOpts = make([]*DynamicStringOpt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicStringOpt{
			FilterIDs: filterIDs,
			Value:     opt,
		})
	}
	return
}

func MapToDynamicBoolOpts(optsMap map[string]bool) (dynOpts []*DynamicBoolOpt) {
	dynOpts = make([]*DynamicBoolOpt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicBoolOpt{
			FilterIDs: filterIDs,
			Value:     opt,
		})
	}
	return
}

func MapToDynamicIntOpts(optsMap map[string]int) (dynOpts []*DynamicIntOpt) {
	dynOpts = make([]*DynamicIntOpt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicIntOpt{
			FilterIDs: filterIDs,
			Value:     opt,
		})
	}
	return
}

func MapToDynamicDecimalBigOpts(optsMap map[string]string) (dynOpts []*DynamicDecimalBigOpt, err error) {
	dynOpts = make([]*DynamicDecimalBigOpt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var bigVal *decimal.Big
		if bigVal, err = StringAsBig(opt); err != nil {
			return
		}
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicDecimalBigOpt{
			FilterIDs: filterIDs,
			Value:     bigVal,
		})
	}
	return
}

func MapToDynamicDurationOpts(optsMap map[string]string) (dynOpts []*DynamicDurationOpt, err error) {
	dynOpts = make([]*DynamicDurationOpt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var usageTTL time.Duration
		if usageTTL, err = ParseDurationWithNanosecs(opt); err != nil {
			return
		}
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicDurationOpt{
			FilterIDs: filterIDs,
			Value:     usageTTL,
		})
	}
	return
}

func MapToDynamicFloat64Opts(optsMap map[string]float64) (dynOpts []*DynamicFloat64Opt) {
	dynOpts = make([]*DynamicFloat64Opt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicFloat64Opt{
			FilterIDs: filterIDs,
			Value:     opt,
		})
	}
	return
}

func MapToDynamicInterfaceOpts(optsMap map[string]interface{}) (dynOpts []*DynamicInterfaceOpt) {
	dynOpts = make([]*DynamicInterfaceOpt, 0, len(optsMap))
	for filters, opt := range optsMap {
		var filterIDs []string
		if filters != EmptyString {
			filterIDs = strings.Split(filters, InfieldSep)
		}
		dynOpts = append(dynOpts, &DynamicInterfaceOpt{
			FilterIDs: filterIDs,
			Value:     opt,
		})
	}
	return
}

func CloneDecimalBig(in *decimal.Big) (cln *decimal.Big) {
	cln = new(decimal.Big)
	cln.Copy(in)
	return
}
