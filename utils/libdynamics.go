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
	"time"

	"github.com/ericlagergren/decimal"
)

type DynamicStringSliceOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Values    []string
}

type DynamicStringOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     string
}

type DynamicIntOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     int
}

type DynamicFloat64Opt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     float64
}

type DynamicBoolOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     bool
}

type DynamicDurationOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     time.Duration
}

type DynamicDecimalBigOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     *decimal.Big
}

type DynamicInterfaceOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     any
}

type DynamicIntPointerOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     *int
}

type DynamicDurationPointerOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     *time.Duration
}

func CloneDynamicStringSliceOpt(in []*DynamicStringSliceOpt) (cl []*DynamicStringSliceOpt) {
	cl = make([]*DynamicStringSliceOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicStringSliceOpt{
			Tenant:    val.Tenant,
			FilterIDs: slices.Clone(val.FilterIDs),
			Values:    slices.Clone(val.Values),
		}
	}
	return
}

func CloneDynamicStringOpt(in []*DynamicStringOpt) (cl []*DynamicStringOpt) {
	cl = make([]*DynamicStringOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicStringOpt{
			Tenant:    val.Tenant,
			FilterIDs: slices.Clone(val.FilterIDs),
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
			FilterIDs: slices.Clone(val.FilterIDs),
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
			FilterIDs: slices.Clone(val.FilterIDs),
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
			FilterIDs: slices.Clone(val.FilterIDs),
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
			FilterIDs: slices.Clone(val.FilterIDs),
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
			FilterIDs: slices.Clone(val.FilterIDs),
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
			FilterIDs: slices.Clone(val.FilterIDs),
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
			FilterIDs: slices.Clone(val.FilterIDs),
			Value:     IntPointer(*val.Value),
		}
	}
	return
}

func CloneDynamicDurationPointerOpt(in []*DynamicDurationPointerOpt) (cl []*DynamicDurationPointerOpt) {
	cl = make([]*DynamicDurationPointerOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicDurationPointerOpt{
			Tenant:    val.Tenant,
			FilterIDs: slices.Clone(val.FilterIDs),
			Value:     DurationPointer(*val.Value),
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if !slices.Equal(v1[i].Values, v2[i].Values) {
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) ||
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
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
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if *v1[i].Value != *v2[i].Value {
			return false
		}
	}
	return true
}

func DynamicDurationPointerOptEqual(v1, v2 []*DynamicDurationPointerOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) {
			return false
		}
		if *v1[i].Value != *v2[i].Value {
			return false
		}
	}
	return true
}

func StringToDecimalBigDynamicOpts(strOpts []*DynamicStringOpt) (decOpts []*DynamicDecimalBigOpt, err error) {
	decOpts = make([]*DynamicDecimalBigOpt, len(strOpts))
	for index, opt := range strOpts {
		decOpts[index] = &DynamicDecimalBigOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		if decOpts[index].Value, err = StringAsBig(opt.Value); err != nil {
			return
		}
	}
	return
}

func DecimalBigToStringDynamicOpts(decOpts []*DynamicDecimalBigOpt) (strOpts []*DynamicStringOpt) {
	strOpts = make([]*DynamicStringOpt, len(decOpts))
	for index, opt := range decOpts {
		strOpts[index] = &DynamicStringOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     opt.Value.String(),
		}
	}
	return
}

func StringToDurationDynamicOpts(strOpts []*DynamicStringOpt) (durOpts []*DynamicDurationOpt, err error) {
	durOpts = make([]*DynamicDurationOpt, len(strOpts))
	for index, opt := range strOpts {
		durOpts[index] = &DynamicDurationOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		if durOpts[index].Value, err = ParseDurationWithNanosecs(opt.Value); err != nil {
			return
		}
	}
	return
}

func DurationToStringDynamicOpts(durOpts []*DynamicDurationOpt) (strOpts []*DynamicStringOpt) {
	strOpts = make([]*DynamicStringOpt, len(durOpts))
	for index, opt := range durOpts {
		strOpts[index] = &DynamicStringOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     opt.Value.String(),
		}
	}
	return
}

func IntToIntPointerDynamicOpts(intOpts []*DynamicIntOpt) (intPtOpts []*DynamicIntPointerOpt) {
	intPtOpts = make([]*DynamicIntPointerOpt, len(intOpts))
	for index, opt := range intOpts {
		intPtOpts[index] = &DynamicIntPointerOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     IntPointer(opt.Value),
		}
	}
	return
}

func IntPointerToIntDynamicOpts(intPtOpts []*DynamicIntPointerOpt) (intOpts []*DynamicIntOpt) {
	intOpts = make([]*DynamicIntOpt, len(intPtOpts))
	for index, opt := range intPtOpts {
		intOpts[index] = &DynamicIntOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     *opt.Value,
		}
	}
	return
}

func StringToDurationPointerDynamicOpts(strOpts []*DynamicStringOpt) (durPtOpts []*DynamicDurationPointerOpt, err error) {
	durPtOpts = make([]*DynamicDurationPointerOpt, len(strOpts))
	for index, opt := range strOpts {
		var durOpt time.Duration
		if durOpt, err = ParseDurationWithNanosecs(opt.Value); err != nil {
			return
		}
		durPtOpts[index] = &DynamicDurationPointerOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     DurationPointer(durOpt),
		}
	}
	return
}

func DurationPointerToStringDynamicOpts(durPtOpts []*DynamicDurationPointerOpt) (strOpts []*DynamicStringOpt) {
	strOpts = make([]*DynamicStringOpt, len(durPtOpts))
	for index, opt := range durPtOpts {
		strOpts[index] = &DynamicStringOpt{
			FilterIDs: opt.FilterIDs,
			Tenant:    opt.Tenant,
			Value:     (*opt.Value).String(),
		}
	}
	return
}
