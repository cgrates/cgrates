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
	"time"

	"github.com/ericlagergren/decimal"
)

type DynamicStringSliceOpt struct {
	Tenant    string
	FilterIDs []string
	Value     []string
}

type DynamicStringOpt struct {
	Tenant    string
	FilterIDs []string
	Value     string
}

type DynamicIntOpt struct {
	Tenant    string
	FilterIDs []string
	Value     int
}

type DynamicFloat64Opt struct {
	Tenant    string
	FilterIDs []string
	Value     float64
}

type DynamicBoolOpt struct {
	Tenant    string
	FilterIDs []string
	Value     bool
}

type DynamicDurationOpt struct {
	Tenant    string
	FilterIDs []string
	Value     time.Duration
}

type DynamicDecimalBigOpt struct {
	Tenant    string
	FilterIDs []string
	Value     *decimal.Big
}

type DynamicInterfaceOpt struct {
	Tenant    string
	FilterIDs []string
	Value     interface{}
}

type DynamicIntPointerOpt struct {
	Tenant    string
	FilterIDs []string
	Value     *int
}

func CloneDynamicStringSliceOpt(in []*DynamicStringSliceOpt) (cl []*DynamicStringSliceOpt) {
	cl = make([]*DynamicStringSliceOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicStringSliceOpt{
			Tenant:    val.Tenant,
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     val.Value,
		}
	}
	return
}

func CloneDynamicStringOpt(in []*DynamicStringOpt) (cl []*DynamicStringOpt) {
	cl = make([]*DynamicStringOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicStringOpt{
			Tenant:    val.Tenant,
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
			Tenant:    val.Tenant,
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
			Tenant:    val.Tenant,
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
			Tenant:    val.Tenant,
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
			Tenant:    val.Tenant,
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
			Tenant:    val.Tenant,
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
			Tenant:    val.Tenant,
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     CloneDecimalBig(val.Value),
		}
	}
	return
}

func CloneDynamicIntPointerOpt(in []*DynamicIntPointerOpt) (cl []*DynamicIntPointerOpt) {
	cl = make([]*DynamicIntPointerOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicIntPointerOpt{
			Tenant:    val.Tenant,
			FilterIDs: CloneStringSlice(val.FilterIDs),
			Value:     IntPointer(*val.Value),
		}
	}
	return
}

func DynamicStringSliceOptEqual(v1, v2 []*DynamicStringSliceOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
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
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
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
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
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
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
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
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
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
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
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
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
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
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if v1[i].Value != v2[i].Value {
			return false
		}
	}
	return true
}

func DynamicIntPointerOptEqual(v1, v2 []*DynamicIntPointerOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
		if !SliceStringEqual(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if *v1[i].Value != *v2[i].Value {
			return false
		}
	}
	return true
}

func CloneDecimalBig(in *decimal.Big) (cln *decimal.Big) {
	cln = new(decimal.Big)
	cln.Copy(in)
	return
}

func StringToDecimalBigDynamicOpts(strOpts []*DynamicStringOpt) (decOpts []*DynamicDecimalBigOpt, err error) {
	decOpts = make([]*DynamicDecimalBigOpt, len(strOpts))
	for index, opt := range strOpts {
		decOpts[index].Tenant = opt.Tenant
		decOpts[index].FilterIDs = opt.FilterIDs
		if decOpts[index].Value, err = StringAsBig(opt.Value); err != nil {
			return
		}
	}
	return
}

func DecimalBigToStringDynamicOpts(decOpts []*DynamicDecimalBigOpt) (strOpts []*DynamicStringOpt) {
	strOpts = make([]*DynamicStringOpt, len(decOpts))
	for index, opt := range decOpts {
		strOpts[index].Tenant = opt.Tenant
		strOpts[index].FilterIDs = opt.FilterIDs
		strOpts[index].Value = opt.Value.String()
	}
	return
}

func StringToDurationDynamicOpts(strOpts []*DynamicStringOpt) (durOpts []*DynamicDurationOpt, err error) {
	durOpts = make([]*DynamicDurationOpt, len(strOpts))
	for index, opt := range strOpts {
		durOpts[index].Tenant = opt.Tenant
		durOpts[index].FilterIDs = opt.FilterIDs
		if durOpts[index].Value, err = ParseDurationWithNanosecs(opt.Value); err != nil {
			return
		}
	}
	return
}

func DurationToStringDynamicOpts(durOpts []*DynamicDurationOpt) (strOpts []*DynamicStringOpt) {
	strOpts = make([]*DynamicStringOpt, len(durOpts))
	for index, opt := range durOpts {
		strOpts[index].Tenant = opt.Tenant
		strOpts[index].FilterIDs = opt.FilterIDs
		strOpts[index].Value = opt.Value.String()
	}
	return
}

func IntToIntPointerDynamicOpts(intOpts []*DynamicIntOpt) (intPtOpts []*DynamicIntPointerOpt) {
	intPtOpts = make([]*DynamicIntPointerOpt, len(intOpts))
	for index, opt := range intOpts {
		intPtOpts[index].Tenant = opt.Tenant
		intPtOpts[index].FilterIDs = opt.FilterIDs
		intPtOpts[index].Value = IntPointer(opt.Value)
	}
	return
}

func IntPointerToIntDynamicOpts(intPtOpts []*DynamicIntPointerOpt) (intOpts []*DynamicIntOpt) {
	intOpts = make([]*DynamicIntOpt, len(intPtOpts))
	for index, opt := range intPtOpts {
		intOpts[index].Tenant = opt.Tenant
		intOpts[index].FilterIDs = opt.FilterIDs
		intOpts[index].Value = *opt.Value
	}
	return
}
