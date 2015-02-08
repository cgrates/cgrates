/*
Real-time Charging System for Telecom & ISP environments
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
	if jsnCfgFld.Cdr_field_id != nil {
		cfgFld.CdrFieldId = *jsnCfgFld.Cdr_field_id
	}
	if jsnCfgFld.Value != nil {
		if cfgFld.Value, err = utils.ParseRSRFields(*jsnCfgFld.Value, utils.INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	if jsnCfgFld.Field_filter != nil {
		if cfgFld.FieldFilter, err = utils.ParseRSRFields(*jsnCfgFld.Field_filter, utils.INFIELD_SEP); err != nil {
			return nil, err
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
	return cfgFld, nil
}

type CfgCdrField struct {
	Tag         string // Identifier for the administrator
	Type        string // Type of field
	CdrFieldId  string // StoredCdr field name
	Value       utils.RSRFields
	FieldFilter utils.RSRFields
	Width       int
	Strip       string
	Padding     string
	Layout      string
	Mandatory   bool
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
