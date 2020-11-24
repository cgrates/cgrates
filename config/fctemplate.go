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
	"fmt"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func NewFCTemplateFromFCTemplateJsonCfg(jsnCfg *FcTemplateJsonCfg, separator string) (*FCTemplate, error) {
	fcTmp := new(FCTemplate)
	var err error
	if jsnCfg.Type != nil {
		fcTmp.Type = *jsnCfg.Type
	}
	if jsnCfg.Path != nil {
		fcTmp.Path = *jsnCfg.Path
		fcTmp.pathSlice = strings.Split(*jsnCfg.Path, utils.NestingSep)
		fcTmp.pathItems = utils.NewPathItems(fcTmp.pathSlice)
		fcTmp.Tag = fcTmp.Path
	}
	fcTmp.Tag = fcTmp.Path
	if jsnCfg.Tag != nil {
		fcTmp.Tag = *jsnCfg.Tag
	}
	if jsnCfg.Filters != nil {
		fcTmp.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			fcTmp.Filters[i] = fltr
		}
	}
	if jsnCfg.Value != nil {
		if fcTmp.Value, err = NewRSRParsers(*jsnCfg.Value, separator); err != nil {
			return nil, err
		}
	}
	if jsnCfg.Width != nil {
		fcTmp.Width = *jsnCfg.Width
	}
	if jsnCfg.Strip != nil {
		fcTmp.Strip = *jsnCfg.Strip
	}
	if jsnCfg.Padding != nil {
		fcTmp.Padding = *jsnCfg.Padding
	}
	if jsnCfg.Mandatory != nil {
		fcTmp.Mandatory = *jsnCfg.Mandatory
	}
	if jsnCfg.Attribute_id != nil {
		fcTmp.AttributeID = *jsnCfg.Attribute_id
	}
	if jsnCfg.New_branch != nil {
		fcTmp.NewBranch = *jsnCfg.New_branch
	}
	if jsnCfg.Timezone != nil {
		fcTmp.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Blocker != nil {
		fcTmp.Blocker = *jsnCfg.Blocker
	}
	fcTmp.Layout = time.RFC3339
	if jsnCfg.Layout != nil {
		fcTmp.Layout = *jsnCfg.Layout
	}
	if jsnCfg.Cost_shift_digits != nil {
		fcTmp.CostShiftDigits = *jsnCfg.Cost_shift_digits
	}
	if jsnCfg.Rounding_decimals != nil {
		fcTmp.RoundingDecimals = new(int)
		*fcTmp.RoundingDecimals = *jsnCfg.Rounding_decimals
	}
	if jsnCfg.Mask_destinationd_id != nil {
		fcTmp.MaskDestID = *jsnCfg.Mask_destinationd_id
	}
	if jsnCfg.Mask_length != nil {
		fcTmp.MaskLen = *jsnCfg.Mask_length
	}
	return fcTmp, nil
}

type FCTemplate struct {
	Tag              string
	Type             string   // Type of field
	Path             string   // Field identifier
	Filters          []string // list of filter profiles
	Value            RSRParsers
	Width            int
	Strip            string
	Padding          string
	Mandatory        bool
	AttributeID      string // Used by NavigableMap when creating CGREvent/XMLElements
	NewBranch        bool   // Used by NavigableMap when creating XMLElements
	Timezone         string
	Blocker          bool
	Layout           string // time format
	CostShiftDigits  int    // Used for CDR
	RoundingDecimals *int
	MaskDestID       string
	MaskLen          int
	pathItems        utils.PathItems // Field identifier
	pathSlice        []string        // Used when we set a NMItem to not recreate this slice for every itemsc
}

func FCTemplatesFromFCTemplatesJsonCfg(jsnCfgFlds []*FcTemplateJsonCfg, separator string) (retFields []*FCTemplate, err error) {
	retFields = make([]*FCTemplate, len(jsnCfgFlds))
	for i, jsnFld := range jsnCfgFlds {
		if retFields[i], err = NewFCTemplateFromFCTemplateJsonCfg(jsnFld, separator); err != nil {
			return nil, err
		}
	}
	return
}

// InflateTemplates will replace the *template fields with template content out msgTpls
func InflateTemplates(fcts []*FCTemplate, msgTpls map[string][]*FCTemplate) ([]*FCTemplate, error) {
	var hasTpl bool
	for i := 0; i < len(fcts); {
		if fcts[i].Type == utils.MetaTemplate {
			hasTpl = true
			tplID, err := fcts[i].Value.ParseValue(nil)
			if err != nil {
				return nil, err
			}
			refTpl, has := msgTpls[tplID]
			if !has {
				return nil, fmt.Errorf("no template with id: <%s>", tplID)
			} else if len(refTpl) == 0 {
				return nil, fmt.Errorf("empty template with id: <%s>", tplID)
			}
			wrkSlice := make([]*FCTemplate, len(refTpl)+len(fcts[i:])-1) // so we can cover tpls[i+1:]
			copy(wrkSlice[:len(refTpl)], refTpl)                         // copy fields out of referenced template
			if len(fcts[i:]) > 1 {                                       // copy the rest of the fields after MetaTemplate
				copy(wrkSlice[len(refTpl):], fcts[i+1:])
			}
			fcts = append(fcts[:i], wrkSlice...) // append the work
			continue                             // don't increase index so we can recheck
		}
		i++
	}
	if !hasTpl {
		return nil, nil
	}
	return fcts, nil
}

// Clone returns a deep copy of FCTemplate
func (fc FCTemplate) Clone() (cln *FCTemplate) {
	cln = &FCTemplate{
		Tag:             fc.Tag,
		Type:            fc.Type,
		Path:            fc.Path,
		pathItems:       fc.pathItems.Clone(),
		Value:           fc.Value.Clone(),
		Width:           fc.Width,
		Strip:           fc.Strip,
		Padding:         fc.Padding,
		Mandatory:       fc.Mandatory,
		AttributeID:     fc.AttributeID,
		NewBranch:       fc.NewBranch,
		Timezone:        fc.Timezone,
		Blocker:         fc.Blocker,
		Layout:          fc.Layout,
		CostShiftDigits: fc.CostShiftDigits,
		MaskDestID:      fc.MaskDestID,
		MaskLen:         fc.MaskLen,
	}
	if fc.RoundingDecimals != nil {
		cln.RoundingDecimals = utils.IntPointer(*fc.RoundingDecimals)
	}
	if fc.pathSlice != nil {
		cln.pathSlice = make([]string, len(fc.pathSlice))
		for i, v := range fc.pathSlice {
			cln.pathSlice[i] = v
		}
	}
	if fc.Filters != nil {
		cln.Filters = make([]string, len(fc.Filters))
		for idx, val := range fc.Filters {
			cln.Filters[idx] = val
		}
	}
	return
}

type FcTemplates map[string][]*FCTemplate

// AsMapInterface returns the config as a map[string]interface{}
func (sCft FcTemplates) AsMapInterface(separator string) (initialMP map[string][]map[string]interface{}) {
	initialMP = make(map[string][]map[string]interface{})
	for key, value := range sCft {
		initialMP[key] = make([]map[string]interface{}, len(value))
		for i, item := range value {
			initialMP[key][i] = item.AsMapInterface(separator)
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (fc *FCTemplate) AsMapInterface(separator string) (mp map[string]interface{}) {
	mp = make(map[string]interface{})
	if fc.Tag != utils.EmptyString {
		mp[utils.TagCfg] = fc.Tag
	}
	if fc.Type != utils.EmptyString {
		mp[utils.TypeCf] = fc.Type
	}
	if fc.Path != utils.EmptyString {
		mp[utils.PathCfg] = fc.Path
	}
	if fc.Filters != nil {
		mp[utils.FiltersCfg] = fc.Filters
	}
	if fc.Value != nil {
		mp[utils.ValueCfg] = fc.Value.GetRule(separator)
	}
	if fc.Width != 0 {
		mp[utils.WidthCfg] = fc.Width
	}
	if fc.Strip != utils.EmptyString {
		mp[utils.StripCfg] = fc.Strip
	}
	if fc.Padding != utils.EmptyString {
		mp[utils.PaddingCfg] = fc.Padding
	}
	if fc.Mandatory != false {
		mp[utils.MandatoryCfg] = fc.Mandatory
	}
	if fc.AttributeID != utils.EmptyString {
		mp[utils.AttributeIDCfg] = fc.AttributeID
	}
	if fc.NewBranch != false {
		mp[utils.NewBranchCfg] = fc.NewBranch
	}
	if fc.Timezone != utils.EmptyString {
		mp[utils.TimezoneCfg] = fc.Timezone
	}
	if fc.Blocker != false {
		mp[utils.BlockerCfg] = fc.Blocker
	}
	if fc.Layout != time.RFC3339 {
		mp[utils.LayoutCfg] = fc.Layout
	}
	if fc.CostShiftDigits != 0 {
		mp[utils.CostShiftDigitsCfg] = fc.CostShiftDigits
	}
	if fc.RoundingDecimals != nil {
		mp[utils.RoundingDecimalsCfg] = *fc.RoundingDecimals
	}
	if fc.MaskDestID != utils.EmptyString {
		mp[utils.MaskDestIDCfg] = fc.MaskDestID
	}
	if fc.MaskLen != 0 {
		mp[utils.MaskLenCfg] = fc.MaskLen
	}
	return
}

// GetPathSlice returns the cached split of the path
func (fc *FCTemplate) GetPathSlice() []string {
	return fc.pathSlice
}

// GetPathItems returns the cached path as PathItems
func (fc *FCTemplate) GetPathItems() utils.PathItems {
	return fc.pathItems
}

// ComputePath used in test to populate private fields used to store the path
func (fc *FCTemplate) ComputePath() {
	fc.pathSlice = strings.Split(fc.Path, utils.NestingSep)
	fc.pathItems = utils.NewPathItems(fc.pathSlice)
}

// Clone returns a deep copy of FcTemplates
func (sCft FcTemplates) Clone() (cln FcTemplates) {
	if sCft == nil {
		return
	}
	cln = make(FcTemplates)
	for k, fcs := range sCft {
		fcln := make([]*FCTemplate, len(fcs))
		for i, fc := range fcs {
			fcln[i] = fc.Clone()
		}
		cln[k] = fcln
	}
	return
}
