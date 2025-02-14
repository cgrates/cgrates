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
package config

import (
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

type DynamicStringSliceOpt struct {
	FilterIDs []string
	Tenant    string
	Values    []string
}

type DynamicStringOpt struct {
	FilterIDs []string
	Tenant    string
	value     string
	rsVal     RSRParsers
}

type DynamicIntOpt struct {
	FilterIDs []string
	Tenant    string
	value     int
	rsVal     RSRParsers
}

type DynamicFloat64Opt struct {
	FilterIDs []string
	Tenant    string
	value     float64
	rsVal     RSRParsers
}

type DynamicBoolOpt struct {
	FilterIDs []string
	Tenant    string
	value     bool
	rsVal     RSRParsers
}

type DynamicDurationOpt struct {
	FilterIDs []string
	Tenant    string
	value     time.Duration
	rsVal     RSRParsers
}

type DynamicDecimalOpt struct {
	FilterIDs []string
	Tenant    string
	value     *decimal.Big
	rsVal     RSRParsers
}

type DynamicInterfaceOpt struct {
	FilterIDs []string `json:",omitempty"`
	Tenant    string
	Value     any
}

type DynamicIntPointerOpt struct {
	FilterIDs []string
	Tenant    string
	value     *int
	rsVal     RSRParsers
}

type DynamicDurationPointerOpt struct {
	FilterIDs []string
	Tenant    string
	value     *time.Duration
	rsVal     RSRParsers
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
			value:     val.value,
			rsVal:     val.rsVal.Clone(),
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
			value:     val.value,
			rsVal:     val.rsVal.Clone(),
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
			value:     val.value,
			rsVal:     val.rsVal,
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
			value:     val.value,
			rsVal:     val.rsVal.Clone(),
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
			value:     val.value,
			rsVal:     val.rsVal.Clone(),
		}
	}
	return
}

func CloneDynamicDecimalOpt(in []*DynamicDecimalOpt) (cl []*DynamicDecimalOpt) {
	cl = make([]*DynamicDecimalOpt, len(in))
	for i, val := range in {
		cl[i] = &DynamicDecimalOpt{
			Tenant:    val.Tenant,
			FilterIDs: slices.Clone(val.FilterIDs),
			value:     utils.CloneDecimalBig(val.value),
			rsVal:     val.rsVal.Clone(),
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
			value:     val.value,
			rsVal:     val.rsVal,
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
			value:     utils.DurationPointer(*val.value),
			rsVal:     val.rsVal.Clone(),
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
		if v1[i].value != v2[i].value {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
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
		if v1[i].value != v2[i].value {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
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
		if v1[i].value != v2[i].value {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
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
		if v1[i].value != v2[i].value {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
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
		if v1[i].value != v2[i].value {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
			return false
		}
	}
	return true
}

func DynamicDecimalOptEqual(v1, v2 []*DynamicDecimalOpt) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v1 {
		if v1[i].Tenant != v2[i].Tenant {
			return false
		}
		if !slices.Equal(v1[i].FilterIDs, v2[i].FilterIDs) ||
			v1[i].value.Cmp(v2[i].value) != 0 {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
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
		if *v1[i].value != *v2[i].value {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
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
		if *v1[i].value != *v2[i].value {
			return false
		}
		if v1[i].rsVal.GetRule() != v2[i].rsVal.GetRule() {
			return false
		}
	}
	return true
}

func InterfaceToDynamicStringOpts(strJsOpts []*DynamicInterfaceOpt) (strOpts []*DynamicStringOpt, err error) {
	strOpts = make([]*DynamicStringOpt, len(strJsOpts))
	for indx, opt := range strJsOpts {
		strOpts[indx] = &DynamicStringOpt{
			FilterIDs: opt.FilterIDs,
			Tenant:    opt.Tenant,
		}
		strval := utils.IfaceAsString(opt.Value)
		if strings.HasPrefix(strval, utils.DynamicDataPrefix) {
			strOpts[indx].rsVal, err = NewRSRParsers(strval, utils.RSRSep)
			if err != nil {
				return
			}
			continue
		}
		strOpts[indx].value = strval
	}
	return
}

func DynamicStringToInterfaceOpts(strOpts []*DynamicStringOpt) (strJsOpts []*DynamicInterfaceOpt) {
	strJsOpts = make([]*DynamicInterfaceOpt, len(strOpts))
	for index, opt := range strOpts {
		strJsOpts[index] = &DynamicInterfaceOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		if opt.rsVal != nil {
			strOpts[index].value = opt.rsVal.GetRule()
			continue
		}
		strJsOpts[index].Value = opt.value
	}
	return
}

func InterfaceToFloat64DynamicOpts(strJsOpts []*DynamicInterfaceOpt) (flOpts []*DynamicFloat64Opt, err error) {
	flOpts = make([]*DynamicFloat64Opt, len(strJsOpts))
	for index, opt := range strJsOpts {
		flOpts[index] = &DynamicFloat64Opt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		strVal := utils.IfaceAsString(opt.Value)
		if strings.HasPrefix(strVal, utils.DynamicDataPrefix) {
			flOpts[index].rsVal, err = NewRSRParsers(strVal, utils.RSRSep)
			if err != nil {
				return nil, err
			}
			continue
		}
		var flVal float64
		flVal, err = strconv.ParseFloat(strVal, 64)
		if err != nil {
			return
		}
		flOpts[index].value = flVal
	}
	return
}

func Float64ToInterfaceDynamicOpts(flOpts []*DynamicFloat64Opt) (strJsOpts []*DynamicInterfaceOpt) {
	strJsOpts = make([]*DynamicInterfaceOpt, len(flOpts))
	for index, opt := range flOpts {
		strJsOpts[index] = &DynamicInterfaceOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     opt.value,
		}
	}
	return
}

func IfaceToIntDynamicOpts(strJsOpts []*DynamicInterfaceOpt) (intPtOpts []*DynamicIntOpt, err error) {
	intPtOpts = make([]*DynamicIntOpt, len(strJsOpts))
	for index, opt := range strJsOpts {
		intPtOpts[index] = &DynamicIntOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		strval := utils.IfaceAsString(opt.Value)
		if strings.HasPrefix(strval, utils.DynamicDataPrefix) {
			intPtOpts[index].rsVal, err = NewRSRParsers(strval, utils.RSRSep)
			if err != nil {
				return nil, err
			}
			continue
		}
		var intVal int
		intVal, err = utils.IfaceAsInt(opt.Value)
		if err != nil {
			return
		}
		intPtOpts[index].value = intVal
	}
	return
}

func IntToIfaceDynamicOpts(intPtOpts []*DynamicIntOpt) (strJsOpts []*DynamicInterfaceOpt) {
	strJsOpts = make([]*DynamicInterfaceOpt, len(intPtOpts))
	for index, opt := range intPtOpts {
		strJsOpts[index] = &DynamicInterfaceOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     opt.value,
		}
	}
	return
}

func IfaceToDecimalBigDynamicOpts(strOpts []*DynamicInterfaceOpt) (decOpts []*DynamicDecimalOpt, err error) {
	decOpts = make([]*DynamicDecimalOpt, len(strOpts))
	for index, opt := range strOpts {
		decOpts[index] = &DynamicDecimalOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		strVal := utils.IfaceAsString(opt.Value)
		if strings.HasPrefix(strVal, utils.DynamicDataPrefix) {
			decOpts[index].rsVal, err = NewRSRParsers(strVal, utils.RSRSep)
			if err != nil {
				return nil, err
			}
			continue
		}
		if decOpts[index].value, err = utils.IfaceAsBig(opt.Value); err != nil {
			return
		}
	}
	return
}

func DecimalToIfaceDynamicOpts(decOpts []*DynamicDecimalOpt) (strOpts []*DynamicInterfaceOpt) {
	strOpts = make([]*DynamicInterfaceOpt, len(decOpts))
	for index, opt := range decOpts {
		strOpts[index] = &DynamicInterfaceOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		if opt.value == nil {
			strOpts[index].Value = opt.rsVal.GetRule()
			continue
		}
		strOpts[index].Value = opt.value.String()
	}
	return
}

func IfaceToDurationDynamicOpts(ifOpts []*DynamicInterfaceOpt) (durOpts []*DynamicDurationOpt, err error) {
	durOpts = make([]*DynamicDurationOpt, len(ifOpts))
	for index, opt := range ifOpts {
		durOpts[index] = &DynamicDurationOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		strVal := utils.IfaceAsString(opt.Value)
		if strings.HasPrefix(strVal, utils.DynamicDataPrefix) {
			durOpts[index].rsVal, err = NewRSRParsers(strVal, utils.RSRSep)
			if err != nil {
				return nil, err
			}
			continue
		}
		if durOpts[index].value, err = utils.ParseDurationWithNanosecs(strVal); err != nil {
			return
		}
	}
	return
}

func DurationToIfaceDynamicOpts(durOpts []*DynamicDurationOpt) (strOpts []*DynamicInterfaceOpt) {
	strOpts = make([]*DynamicInterfaceOpt, len(durOpts))
	for index, opt := range durOpts {
		strOpts[index] = &DynamicInterfaceOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		if opt.rsVal != nil {
			strOpts[index].Value = opt.rsVal.GetRule()
			continue
		}
		strOpts[index].Value = opt.value
	}
	return
}

func IfaceToIntPointerDynamicOpts(ifOpts []*DynamicInterfaceOpt) (intPtOpts []*DynamicIntPointerOpt, err error) {
	intPtOpts = make([]*DynamicIntPointerOpt, len(ifOpts))
	for index, opt := range ifOpts {
		intPtOpts[index] = &DynamicIntPointerOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		strval := utils.IfaceAsString(opt.Value)
		if strings.HasPrefix(strval, utils.DynamicDataPrefix) {
			intPtOpts[index].rsVal, err = NewRSRParsers(strval, utils.RSRSep)
			if err != nil {
				return nil, err
			}
			continue
		}
		var intVal int
		intVal, err = strconv.Atoi(strval)
		if err != nil {
			return
		}
		intPtOpts[index].value = utils.IntPointer(intVal)
	}
	return
}

func IntPointerToIfaceDynamicOpts(intPtOpts []*DynamicIntPointerOpt) (strJsOpts []*DynamicInterfaceOpt) {
	strJsOpts = make([]*DynamicInterfaceOpt, len(intPtOpts))
	for index, opt := range intPtOpts {
		strJsOpts[index] = &DynamicInterfaceOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
			Value:     opt.value,
		}
	}
	return
}

func IfaceToDurationPointerDynamicOpts(ifOpts []*DynamicInterfaceOpt) (durPtOpts []*DynamicDurationPointerOpt, err error) {
	durPtOpts = make([]*DynamicDurationPointerOpt, len(ifOpts))
	for index, opt := range ifOpts {
		durPtOpts[index] = &DynamicDurationPointerOpt{
			Tenant:    opt.Tenant,
			FilterIDs: opt.FilterIDs,
		}
		strVal := utils.IfaceAsString(opt.Value)
		if strings.HasPrefix(strVal, utils.DynamicDataPrefix) {
			durPtOpts[index].rsVal, err = NewRSRParsers(strVal, utils.RSRSep)
			if err != nil {
				return nil, err
			}
			continue
		}
		var durOpt time.Duration
		if durOpt, err = utils.ParseDurationWithNanosecs(strVal); err != nil {
			return
		}
		durPtOpts[index].value = utils.DurationPointer(durOpt)
	}
	return
}

func DurationPointerToIfaceDynamicOpts(durPtOpts []*DynamicDurationPointerOpt) (strOpts []*DynamicInterfaceOpt) {
	strOpts = make([]*DynamicInterfaceOpt, len(durPtOpts))
	for index, opt := range durPtOpts {
		strOpts[index] = &DynamicInterfaceOpt{
			FilterIDs: opt.FilterIDs,
			Tenant:    opt.Tenant,
			Value:     opt.value,
		}
	}
	return
}

func IfaceToBoolDynamicOpts(strOpts []*DynamicInterfaceOpt) (boolOpts []*DynamicBoolOpt, err error) {
	boolOpts = make([]*DynamicBoolOpt, len(strOpts))
	for index, opt := range strOpts {
		boolOpts[index] = &DynamicBoolOpt{
			FilterIDs: opt.FilterIDs,
			Tenant:    opt.Tenant,
		}
		if dynVal := utils.IfaceAsString(opt.Value); strings.HasPrefix(dynVal, utils.DynamicDataPrefix) {
			boolOpts[index].rsVal, err = NewRSRParsers(dynVal, utils.RSRSep)
			if err != nil {
				return
			}
			continue
		}
		var boolOpt bool
		boolOpt, err = utils.IfaceAsBool(opt.Value)
		if err != nil {
			return
		}
		boolOpts[index].value = boolOpt
	}
	return
}

func BoolToIfaceDynamicOpts(boolOpts []*DynamicBoolOpt) (strOpts []*DynamicInterfaceOpt) {
	strOpts = make([]*DynamicInterfaceOpt, len(boolOpts))
	for index, opt := range boolOpts {
		strOpts[index] = &DynamicInterfaceOpt{
			FilterIDs: opt.FilterIDs,
			Tenant:    opt.Tenant,
			Value:     opt.value,
		}
	}
	return
}

func NewDynamicStringOpt(filterIDs []string, tenant string, value string, dynValue RSRParsers) *DynamicStringOpt {
	return &DynamicStringOpt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func NewDynamicIntOpt(filterIDs []string, tenant string, value int, dynValue RSRParsers) *DynamicIntOpt {
	return &DynamicIntOpt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func NewDynamicFloat64Opt(filterIDs []string, tenant string, value float64, dynValue RSRParsers) *DynamicFloat64Opt {
	return &DynamicFloat64Opt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func NewDynamicDecimalOpt(filterIDs []string, tenant string, value *decimal.Big, dynValue RSRParsers) *DynamicDecimalOpt {
	return &DynamicDecimalOpt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func NewDynamicDurationOpt(filterIDs []string, tenant string, value time.Duration, dynValue RSRParsers) *DynamicDurationOpt {
	return &DynamicDurationOpt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func NewDynamicBoolOpt(filterIDs []string, tenant string, value bool, dynValue RSRParsers) *DynamicBoolOpt {
	return &DynamicBoolOpt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func NewDynamicIntPointerOpt(filterIDs []string, tenant string, value *int, dynValue RSRParsers) *DynamicIntPointerOpt {
	return &DynamicIntPointerOpt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func NewDynamicDurationPointerOpt(filterIDs []string, tenant string, value *time.Duration, dynValue RSRParsers) *DynamicDurationPointerOpt {
	return &DynamicDurationPointerOpt{
		FilterIDs: filterIDs,
		Tenant:    tenant,
		value:     value,
		rsVal:     dynValue,
	}
}

func (dynStr *DynamicStringOpt) Value(dP utils.DataProvider) (string, error) {
	if dynStr.rsVal != nil {
		out, err := dynStr.rsVal.ParseDataProvider(dP)
		if err != nil {
			return "", err
		}
		return out, nil
	}
	return dynStr.value, nil
}

func (dynInt *DynamicIntOpt) Value(dP utils.DataProvider) (int, error) {
	if dynInt.rsVal != nil {
		out, err := dynInt.rsVal.ParseDataProvider(dP)
		if err != nil {
			return 0, err
		}
		return utils.IfaceAsInt(out)
	}
	return dynInt.value, nil
}

func (dynFlt *DynamicFloat64Opt) Value(dP utils.DataProvider) (float64, error) {
	if dynFlt.rsVal != nil {
		out, err := dynFlt.rsVal.ParseDataProvider(dP)
		if err != nil {
			return 0, err
		}
		return utils.IfaceAsFloat64(out)
	}
	return dynFlt.value, nil
}

func (dynFlt *DynamicBoolOpt) Value(dP utils.DataProvider) (bool, error) {
	if dynFlt.rsVal != nil {
		out, err := dynFlt.rsVal.ParseDataProvider(dP)
		if err != nil {
			return false, err
		}
		return utils.IfaceAsBool(out)
	}
	return dynFlt.value, nil
}

func (dynDec *DynamicDecimalOpt) Value(dP utils.DataProvider) (*decimal.Big, error) {
	if dynDec.rsVal != nil {
		out, err := dynDec.rsVal.ParseDataProviderWithInterfaces2(dP)
		if err != nil {
			return nil, err
		}
		return utils.IfaceAsBig(out)
	}
	return dynDec.value, nil
}

func (dynDur *DynamicDurationOpt) Value(dP utils.DataProvider) (time.Duration, error) {
	if dynDur.rsVal != nil {
		out, err := dynDur.rsVal.ParseDataProvider(dP)
		if err != nil {
			return 0, err
		}
		return utils.ParseDurationWithNanosecs(out)
	}
	return dynDur.value, nil
}

func (dynFlt *DynamicIntPointerOpt) Value(dP utils.DataProvider) (*int, error) {
	if dynFlt.rsVal != nil {
		out, err := dynFlt.rsVal.ParseDataProvider(dP)
		if err != nil {
			return nil, err
		}
		intPnt, err := utils.IfaceAsInt(out)
		if err != nil {
			return nil, err
		}
		return utils.IntPointer(intPnt), nil
	}
	return dynFlt.value, nil
}

func (dynFlt *DynamicDurationPointerOpt) Value(dP utils.DataProvider) (*time.Duration, error) {
	if dynFlt.rsVal != nil {
		out, err := dynFlt.rsVal.ParseDataProvider(dP)
		if err != nil {
			return nil, err
		}
		durPnt, err := utils.IfaceAsDuration(out)
		if err != nil {
			return nil, err
		}
		return utils.DurationPointer(durPnt), nil
	}
	return dynFlt.value, nil
}
