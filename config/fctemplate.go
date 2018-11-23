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

	"github.com/cgrates/cgrates/utils"
)

func NewFCTemplateFromFCTemplateJsonCfg(jsnCfg *FcTemplateJsonCfg, separator string) (*FCTemplate, error) {
	fcTmp := new(FCTemplate)
	var err error
	if jsnCfg.Tag != nil {
		fcTmp.Tag = *jsnCfg.Tag
	}
	if jsnCfg.Type != nil {
		fcTmp.Type = *jsnCfg.Type
	}
	if jsnCfg.Field_id != nil {
		fcTmp.FieldId = *jsnCfg.Field_id
	}
	if jsnCfg.Filters != nil {
		fcTmp.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			fcTmp.Filters[i] = fltr
		}
	}
	if jsnCfg.Value != nil {
		if fcTmp.Value, err = NewRSRParsers(*jsnCfg.Value, true, separator); err != nil {
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
	if jsnCfg.Break_on_success != nil {
		fcTmp.BreakOnSuccess = *jsnCfg.Break_on_success
	}
	if jsnCfg.Handler_id != nil {
		fcTmp.HandlerId = *jsnCfg.Handler_id
	}
	if jsnCfg.Layout != nil {
		fcTmp.Layout = *jsnCfg.Layout
	}
	if jsnCfg.Cost_shift_digits != nil {
		fcTmp.CostShiftDigits = *jsnCfg.Cost_shift_digits
	}
	if jsnCfg.Rounding_decimals != nil {
		fcTmp.RoundingDecimals = *jsnCfg.Rounding_decimals
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
	FieldId          string   // Field identifier
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
	BreakOnSuccess   bool
	HandlerId        string // used by XML in CDRC
	Layout           string // time format
	CostShiftDigits  int    // Used for CDR
	RoundingDecimals int
	MaskDestID       string
	MaskLen          int
}

func FCTemplatesFromFCTemplatesJsonCfg(jsnCfgFlds []*FcTemplateJsonCfg, separator string) ([]*FCTemplate, error) {
	retFields := make([]*FCTemplate, len(jsnCfgFlds))
	var err error
	for i, jsnFld := range jsnCfgFlds {
		if retFields[i], err = NewFCTemplateFromFCTemplateJsonCfg(jsnFld, separator); err != nil {
			return nil, err
		}
	}
	return retFields, nil
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
