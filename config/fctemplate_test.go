/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestNewFCTemplateFromFCTemplateJsonCfg(t *testing.T) {

	tag := "tagTest"
	typ := "typeTest"
	path := "pathTest"
	att := "attTest"
	slc := []string{"val1", "val2"}
	val := "valTest"
	width := 10
	strip := "stripTest"
	pad := "padTest"
	mand := false
	newBr := false
	time := "timeTest"
	block := false
	brk := false
	lay := "layTest"
	cost := 50
	round := 1
	maskD := "maskDTest"
	maskL := 5

	fcJs := FcTemplateJsonCfg{
		Tag:                  &tag,
		Type:                 &typ,
		Path:                 &path,
		Attribute_id:         &att,
		Filters:              &slc,
		Value:                &val,
		Width:                &width,
		Strip:                &strip,
		Padding:              &pad,
		Mandatory:            &mand,
		New_branch:           &newBr,
		Timezone:             &time,
		Blocker:              &block,
		Break_on_success:     &brk,
		Layout:               &lay,
		Cost_shift_digits:    &cost,
		Rounding_decimals:    &round,
		Mask_destinationd_id: &maskD,
		Mask_length:          &maskL,
	}

	valRp, _ := NewRSRParsers(val, true, "")
	pSlc := strings.Split(path, utils.NestingSep)
	pItm := utils.NewPathItems(pSlc)

	fcT := FCTemplate{
		Tag:              tag,
		Type:             typ,  // Type of field
		Path:             path, // Field identifier
		Filters:          slc,  // list of filter profiles
		Value:            valRp,
		Width:            width,
		Strip:            strip,
		Padding:          pad,
		Mandatory:        mand,
		AttributeID:      att,   // Used by NavigableMap when creating CGREvent/XMLElements
		NewBranch:        newBr, // Used by NavigableMap when creating XMLElements
		Timezone:         time,
		Blocker:          block,
		BreakOnSuccess:   brk,
		Layout:           lay,  // time format
		CostShiftDigits:  cost, // Used for CDR
		RoundingDecimals: &round,
		MaskDestID:       maskD,
		MaskLen:          maskL,
		pathItems:        pItm, // Field identifier
		pathSlice:        pSlc,
	}

	valErr := "test`test"

	fcJs2 := FcTemplateJsonCfg{
		Value: &valErr,
	}

	type args struct {
		jsnCfg    *FcTemplateJsonCfg
		separator string
	}

	type exp struct {
		val *FCTemplate
		err error
	}

	tests := []struct {
		name string
		args args
		exp  exp
	}{
		{
			name: "cover most if statements",
			args: args{jsnCfg: &fcJs, separator: ""},
			exp:  exp{val: &fcT, err: nil},
		},
		{
			name: "cover most if statements",
			args: args{jsnCfg: &fcJs2, separator: ""},
			exp:  exp{val: nil, err: fmt.Errorf("Unclosed unspilit syntax")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rcv, err := NewFCTemplateFromFCTemplateJsonCfg(tt.args.jsnCfg, tt.args.separator)

			if err != nil {
				if err.Error() != tt.exp.err.Error() {
					t.Fatalf("expected %s, recived %s", tt.exp.err, err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp.val) {
				t.Errorf("expected %v, recived %v", tt.exp.val, rcv)
			}
		})
	}
}

func TestFCTemplatesFromFCTemplatesJsonCfg(t *testing.T) {
	jsnCfgs := []*FcTemplateJsonCfg{
		{
			Tag:     utils.StringPointer("Tenant"),
			Type:    utils.StringPointer("*composed"),
			Path:    utils.StringPointer("Tenant"),
			Filters: &[]string{"Filter1", "Filter2"},
			Value:   utils.StringPointer("cgrates.org"),
		},
		{
			Tag:     utils.StringPointer("RunID"),
			Type:    utils.StringPointer("*composed"),
			Path:    utils.StringPointer("RunID"),
			Filters: &[]string{"Filter1_1", "Filter2_2"},
			Value:   utils.StringPointer("SampleValue"),
		},
	}
	expected := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
	}
	for _, v := range expected {
		v.ComputePath()
	}
	if rcv, err := FCTemplatesFromFCTemplatesJsonCfg(jsnCfgs, utils.INFIELD_SEP); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	t.Run("recive error", func(t *testing.T) {

		fcJs2 := FcTemplateJsonCfg{
			Value: utils.StringPointer("test`test"),
		}

		slcFc := []*FcTemplateJsonCfg{
			&fcJs2,
		}

		_, err := FCTemplatesFromFCTemplatesJsonCfg(slcFc, "")

		if err.Error() != fmt.Errorf("Unclosed unspilit syntax").Error() {
			t.Error("didn't recive an error or wrong error message")
		}
	})
}

func TestFCTemplateInflateTemplates(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", true, utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {
			{
				Tag:     "Elem1",
				Type:    "*composed",
				Path:    "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2",
				Type:    "*composed",
				Path:    "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", true, utils.INFIELD_SEP),
			},
		},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", true, utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", true, utils.INFIELD_SEP),
			},
		},
	}
	expFC := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "Elem1",
			Type:    "*composed",
			Path:    "Elem1",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "Elem2",
			Type:    "*composed",
			Path:    "Elem2",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("Elem2", true, utils.INFIELD_SEP),
		},
	}
	if rcv, err := InflateTemplates(fcTmp1, fcTmpMp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expFC, rcv) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(expFC), utils.ToJSON(rcv))
	}
}

func TestFCTemplateInflate2(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "TmpMap3",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap3", true, utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {
			{
				Tag:     "Elem1",
				Type:    "*composed",
				Path:    "Elem1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2",
				Type:    "*composed",
				Path:    "Elem2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2", true, utils.INFIELD_SEP),
			},
		},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", true, utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", true, utils.INFIELD_SEP),
			},
		},
	}
	if _, err := InflateTemplates(fcTmp1, fcTmpMp); err.Error() != "no template with id: <TmpMap3>" {
		t.Error(err)
	}
}

func TestFCTemplateInflate3(t *testing.T) {
	fcTmp1 := []*FCTemplate{
		{
			Tag:     "Tenant",
			Type:    "*composed",
			Path:    "Tenant",
			Filters: []string{"Filter1", "Filter2"},
			Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "RunID",
			Type:    "*composed",
			Path:    "RunID",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("SampleValue", true, utils.INFIELD_SEP),
		},
		{
			Tag:     "TmpMap",
			Type:    "*template",
			Filters: []string{"Filter1_1", "Filter2_2"},
			Value:   NewRSRParsersMustCompile("TmpMap", true, utils.INFIELD_SEP),
		},
	}
	fcTmpMp := map[string][]*FCTemplate{
		"TmpMap": {},
		"TmpMap2": {
			{
				Tag:     "Elem2.1",
				Type:    "*composed",
				Path:    "Elem2.1",
				Filters: []string{"Filter1", "Filter2"},
				Value:   NewRSRParsersMustCompile("Elem2.1", true, utils.INFIELD_SEP),
			},
			{
				Tag:     "Elem2.2",
				Type:    "*composed",
				Path:    "Elem2.2",
				Filters: []string{"Filter1_1", "Filter2_2"},
				Value:   NewRSRParsersMustCompile("Elem2.2", true, utils.INFIELD_SEP),
			},
		},
	}
	for _, v := range fcTmp1 {
		v.ComputePath()
	}
	for _, tmpl := range fcTmpMp {
		for _, v := range tmpl {
			v.ComputePath()
		}
	}
	if _, err := InflateTemplates(fcTmp1, fcTmpMp); err == nil ||
		err.Error() != "empty template with id: <TmpMap>" {
		t.Error(err)
	}
}

func TestFCTemplateInflate4(t *testing.T) {

	fcT := FCTemplate{
		Tag:     "TmpMap",
		Type:    "*template",
		Filters: []string{"Filter1_1", "Filter2_2"},
		Value:   NewRSRParsersMustCompile("TmpMap", true, utils.INFIELD_SEP),
	}
	fcT2 := FCTemplate{
		Tag:     "TmpMap2",
		Type:    "*template",
		Filters: []string{"Filter1_1", "Filter2_2"},
		Value:   NewRSRParsersMustCompile("TmpMap", true, utils.INFIELD_SEP),
	}
	fcT3 := FCTemplate{
		Tag:     "TmpMap3",
		Type:    "*template",
		Filters: []string{"Filter1_1", "Filter2_2"},
		Value:   NewRSRParsersMustCompile("TmpMap", true, utils.INFIELD_SEP),
	}

	fcTm := FCTemplate{}

	type args struct {
		fcts    []*FCTemplate
		msgTpls map[string][]*FCTemplate
	}

	type exp struct {
		out []*FCTemplate
		err error
	}

	tests := []struct {
		name string
		args args
		exp  exp
	}{
		{
			name: "does not have templates",
			args: args{[]*FCTemplate{}, map[string][]*FCTemplate{}},
			exp:  exp{out: nil, err: nil},
		},
		{
			name: "expecting error",
			args: args{fcts: []*FCTemplate{&fcT, &fcT2, &fcT3}, msgTpls: map[string][]*FCTemplate{"TmpMap": {&fcTm}}},
			exp:  exp{out: []*FCTemplate{&fcTm, &fcTm, &fcTm}, err: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rcv, err := InflateTemplates(tt.args.fcts, tt.args.msgTpls)

			if err != nil {
				if err.Error() != tt.exp.err.Error() {
					t.Error("wrong error message:", err)
				}
			}

			if !reflect.DeepEqual(rcv, tt.exp.out) {
				t.Errorf("expecting %v, recived %v", tt.exp.out, rcv)
			}
		})
	}
}

func TestFCTemplateClone(t *testing.T) {
	smpl := &FCTemplate{
		Tag:     "Tenant",
		Type:    "*composed",
		Path:    "Tenant",
		Filters: []string{"Filter1", "Filter2"},
		Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
	}
	smpl.ComputePath()
	cloned := smpl.Clone()
	if !reflect.DeepEqual(cloned, smpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(smpl), utils.ToJSON(cloned))
	}
	initialSmpl := &FCTemplate{
		Tag:     "Tenant",
		Type:    "*composed",
		Path:    "Tenant",
		Filters: []string{"Filter1", "Filter2"},
		Value:   NewRSRParsersMustCompile("cgrates.org", true, utils.INFIELD_SEP),
	}
	initialSmpl.ComputePath()
	smpl.Filters = []string{"SingleFilter"}
	smpl.Value = NewRSRParsersMustCompile("cgrates.com", true, utils.INFIELD_SEP)
	if !reflect.DeepEqual(cloned, initialSmpl) {
		t.Errorf("expected: %s ,received: %s", utils.ToJSON(initialSmpl), utils.ToJSON(cloned))
	}
}

func TestFCTemplateGetPathSlice(t *testing.T) {
	fc := FCTemplate{
		Tag:       "Elem1",
		Type:      "*composed",
		Path:      "Elem1",
		Filters:   []string{"Filter1", "Filter2"},
		Value:     NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
		pathItems: utils.PathItems{utils.PathItem{Field: "test"}},
		pathSlice: []string{"val1", "val2"},
	}

	rcv := fc.GetPathSlice()
	exp := []string{"val1", "val2"}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}

}

func TestFCTemplateGetPathItems(t *testing.T) {
	fc := FCTemplate{
		Tag:       "Elem1",
		Type:      "*composed",
		Path:      "Elem1",
		Filters:   []string{"Filter1", "Filter2"},
		Value:     NewRSRParsersMustCompile("Elem1", true, utils.INFIELD_SEP),
		pathItems: utils.PathItems{utils.PathItem{Field: "test"}},
		pathSlice: []string{"val1", "val2"},
	}

	rcv := fc.GetPathItems()
	exp := utils.PathItems{utils.PathItem{Field: "test"}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}

}

func TestFCTemplateAsMapInterface(t *testing.T) {

	tag := "tagTest"
	typ := "typeTest"
	path := "pathTest"
	att := "attTest"
	slc := []string{"val1", "val2"}
	val := "valTest"
	width := 10
	strip := "stripTest"
	pad := "padTest"
	mand := true
	newBr := true
	time := "timeTest"
	block := true
	brk := true
	lay := "layTest"
	cost := 50
	round := 1
	maskD := "maskDTest"
	maskL := 5

	valRp, _ := NewRSRParsers(val, true, "")
	pSlc := strings.Split(path, utils.NestingSep)
	pItm := utils.NewPathItems(pSlc)

	fcT := FCTemplate{
		Tag:              tag,
		Type:             typ,  // Type of field
		Path:             path, // Field identifier
		Filters:          slc,  // list of filter profiles
		Value:            valRp,
		Width:            width,
		Strip:            strip,
		Padding:          pad,
		Mandatory:        mand,
		AttributeID:      att,   // Used by NavigableMap when creating CGREvent/XMLElements
		NewBranch:        newBr, // Used by NavigableMap when creating XMLElements
		Timezone:         time,
		Blocker:          block,
		BreakOnSuccess:   brk,
		Layout:           lay,  // time format
		CostShiftDigits:  cost, // Used for CDR
		RoundingDecimals: &round,
		MaskDestID:       maskD,
		MaskLen:          maskL,
		pathItems:        pItm, // Field identifier
		pathSlice:        pSlc,
	}

	fcMp := map[string]any{
		utils.TagCfg:              tag,
		utils.TypeCf:              typ,
		utils.PathCfg:             path,
		utils.FiltersCfg:          slc,
		utils.ValueCfg:            val,
		utils.WidthCfg:            width,
		utils.StripCfg:            strip,
		utils.PaddingCfg:          pad,
		utils.MandatoryCfg:        mand,
		utils.AttributeIDCfg:      att,
		utils.NewBranchCfg:        newBr,
		utils.TimezoneCfg:         time,
		utils.BlockerCfg:          block,
		utils.BreakOnSuccessCfg:   brk,
		utils.LayoutCfg:           lay,
		utils.CostShiftDigitsCfg:  cost,
		utils.RoundingDecimalsCfg: round,
		utils.MaskDestIDCfg:       maskD,
		utils.MaskLenCfg:          maskL,
	}

	tests := []struct {
		name string
		arg  string
		exp  map[string]any
	}{
		{
			name: "cover most returns",
			arg:  "",
			exp:  fcMp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rcv := fcT.AsMapInterface(tt.arg)

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("recived %v, expected %v", rcv, tt.exp)
			}
		})
	}
}

func TestFCTemplateInflateTemplatesError(t *testing.T) {
	mc := utils.MultiplyConverter{
		Value: 1.2,
	}
	nm := 1
	bl := false
	str := "test)"
	slc := []string{str}
	fcts := []*FCTemplate{{
		Tag:     str,
		Type:    utils.MetaTemplate,
		Path:    str,
		Filters: slc,
		Value: RSRParsers{{
			Rules:           str,
			AllFiltersMatch: bl,
			path:            str,
			converters:      utils.DataConverters{&mc},
		}},
		Width:            nm,
		Strip:            str,
		Padding:          str,
		Mandatory:        bl,
		AttributeID:      str,
		NewBranch:        bl,
		Timezone:         str,
		Blocker:          bl,
		BreakOnSuccess:   bl,
		Layout:           str,
		CostShiftDigits:  nm,
		RoundingDecimals: &nm,
		MaskDestID:       str,
		MaskLen:          nm,
		pathItems:        utils.PathItems{},
		pathSlice:        slc,
	}}
	msgTpls := map[string][]*FCTemplate{str: fcts}
	rcv, err := InflateTemplates(fcts, msgTpls)

	if err != nil {
		if err.Error() != `strconv.ParseFloat: parsing "": invalid syntax` {
			t.Error(err)
		}
	} else {
		t.Error("was expecting an error")
	}

	if rcv != nil {
		t.Error(rcv)
	}
}
