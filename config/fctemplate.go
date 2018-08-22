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

func NewFCTemplateFromFCTemplateJsonCfg(jsnCfg *FcTemplateJsonCfg) *FCTemplate {
	fcTmp := new(FCTemplate)
	if jsnCfg.Id != nil {
		fcTmp.ID = *jsnCfg.Id
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
		fcTmp.Value = utils.NewRSRParsersMustCompile(*jsnCfg.Value, true)
	}
	return fcTmp
}

type FCTemplate struct {
	ID      string
	Type    string   // Type of field
	FieldId string   // Field identifier
	Filters []string // list of filter profiles
	Value   utils.RSRParsers
}

func FCTemplatesFromFCTemapltesJsonCfg(jsnCfgFlds []*FcTemplateJsonCfg) []*FCTemplate {
	retFields := make([]*FCTemplate, len(jsnCfgFlds))
	for i, jsnFld := range jsnCfgFlds {
		retFields[i] = NewFCTemplateFromFCTemplateJsonCfg(jsnFld)
	}
	return retFields
}
