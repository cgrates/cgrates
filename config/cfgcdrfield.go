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
	"github.com/cgrates/cgrates/utils"
)

func NewCfgCdrFieldFromCdrFieldJsonCfg(jsnCfgFld *CdrFieldJsonCfg) (*CfgCdrField, error) {
	var err error
	cfgFld := new(CfgCdrField)
	if jsnCfgFld.Tag != nil {
		cfgFld.Tag = *jsnCfgFld.Tag
	}
	if jsnCfgFld.Type != nil {
		cfgFld.Type = *jsnCfgFld.Type
	}
	if jsnCfgFld.Field_id != nil {
		cfgFld.FieldId = *jsnCfgFld.Field_id
	}
	if jsnCfgFld.Attribute_id != nil {
		cfgFld.AttributeID = *jsnCfgFld.Attribute_id
	}
	if jsnCfgFld.Handler_id != nil {
		cfgFld.HandlerId = *jsnCfgFld.Handler_id
	}
	if jsnCfgFld.Value != nil {
		if cfgFld.Value, err = utils.ParseRSRFields(*jsnCfgFld.Value, utils.INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	if jsnCfgFld.Append != nil {
		cfgFld.Append = *jsnCfgFld.Append
	}
	if jsnCfgFld.Field_filter != nil {
		if cfgFld.FieldFilter, err = utils.ParseRSRFields(*jsnCfgFld.Field_filter, utils.INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	if jsnCfgFld.Filters != nil {
		cfgFld.Filters = make([]string, len(*jsnCfgFld.Filters))
		for i, fltr := range *jsnCfgFld.Filters {
			cfgFld.Filters[i] = fltr
		}
	}
	if jsnCfgFld.Width != nil {
		cfgFld.Width = *jsnCfgFld.Width
	}
	if jsnCfgFld.Strip != nil {
		cfgFld.Strip = *jsnCfgFld.Strip
	}
	if jsnCfgFld.Padding != nil {
		cfgFld.Padding = *jsnCfgFld.Padding
	}
	if jsnCfgFld.Layout != nil {
		cfgFld.Layout = *jsnCfgFld.Layout
	}
	if jsnCfgFld.Mandatory != nil {
		cfgFld.Mandatory = *jsnCfgFld.Mandatory
	}
	if jsnCfgFld.Cost_shift_digits != nil {
		cfgFld.CostShiftDigits = *jsnCfgFld.Cost_shift_digits
	}
	if jsnCfgFld.Rounding_decimals != nil {
		cfgFld.RoundingDecimals = *jsnCfgFld.Rounding_decimals
	}
	if jsnCfgFld.Timezone != nil {
		cfgFld.Timezone = *jsnCfgFld.Timezone
	}
	if jsnCfgFld.Mask_destinationd_id != nil {
		cfgFld.MaskDestID = *jsnCfgFld.Mask_destinationd_id
	}
	if jsnCfgFld.Mask_length != nil {
		cfgFld.MaskLen = *jsnCfgFld.Mask_length
	}
	if jsnCfgFld.Break_on_success != nil {
		cfgFld.BreakOnSuccess = *jsnCfgFld.Break_on_success
	}
	if jsnCfgFld.New_branch != nil {
		cfgFld.NewBranch = *jsnCfgFld.New_branch
	}
	if jsnCfgFld.Blocker != nil {
		cfgFld.Blocker = *jsnCfgFld.Blocker
	}
	return cfgFld, nil
}

type CfgCdrField struct {
	Tag              string // Identifier for the administrator
	Type             string // Type of field
	FieldId          string // Field identifier
	AttributeID      string
	Filters          []string // list of filter profiles
	HandlerId        string
	Value            utils.RSRFields
	Append           bool
	FieldFilter      utils.RSRFields
	Width            int
	Strip            string
	Padding          string
	Layout           string
	Mandatory        bool
	CostShiftDigits  int // Used in exports
	RoundingDecimals int
	Timezone         string
	MaskDestID       string
	MaskLen          int
	BreakOnSuccess   bool
	NewBranch        bool
	Blocker          bool
}

func CfgCdrFieldsFromCdrFieldsJsonCfg(jsnCfgFldss []*CdrFieldJsonCfg) ([]*CfgCdrField, error) {
	retFields := make([]*CfgCdrField, len(jsnCfgFldss))
	for idx, jsnFld := range jsnCfgFldss {
		if cfgFld, err := NewCfgCdrFieldFromCdrFieldJsonCfg(jsnFld); err != nil {
			return nil, err
		} else {
			retFields[idx] = cfgFld
		}
	}
	return retFields, nil
}
