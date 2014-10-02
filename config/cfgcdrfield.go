/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"strings"
)

func NewCfgCdrFieldFromCgrXmlCfgCdrField(xmlCdrFld *XmlCfgCdrField, fixedWidth bool) (*CfgCdrField, error) {
	var err error
	var val, fltr utils.RSRFields
	if xmlCdrFld.Value != nil {
		if xmlCdrFld.Type != nil && utils.IsSliceMember([]string{utils.CONSTANT, utils.FILLER, utils.METATAG, utils.HTTP_POST}, *xmlCdrFld.Type) && !strings.HasPrefix(*xmlCdrFld.Value, utils.STATIC_VALUE_PREFIX) { // Enforce static values for fields which do not support RSR rules
			*xmlCdrFld.Value = utils.STATIC_VALUE_PREFIX + *xmlCdrFld.Value
		}
		if val, err = utils.ParseRSRFields(*xmlCdrFld.Value, utils.INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	if xmlCdrFld.Filter != nil {
		if fltr, err = utils.ParseRSRFields(*xmlCdrFld.Filter, utils.INFIELD_SEP); err != nil {
			return nil, err
		}
	}
	if cdrFld, err := NewCfgCdrFieldWithDefaults(fixedWidth, val, fltr, xmlCdrFld.Type, xmlCdrFld.CdrFieldId, xmlCdrFld.Tag,
		xmlCdrFld.Mandatory, xmlCdrFld.Layout, xmlCdrFld.Width, xmlCdrFld.Strip, xmlCdrFld.Padding); err != nil {
		return nil, err
	} else {
		return cdrFld, nil
	}
}

func NewCfgCdrFieldWithDefaults(fixedWidth bool, val, filter utils.RSRFields, typ, cdrFieldId, tag *string, mandatory *bool, layout *string, width *int, strip, padding *string) (*CfgCdrField, error) {
	cdrField := &CfgCdrField{Value: val, Filter: filter}
	if typ != nil {
		cdrField.Type = *typ
	} else {
		cdrField.Type = utils.CDRFIELD
	}
	if cdrFieldId != nil {
		cdrField.CdrFieldId = *cdrFieldId
	} else if utils.IsSliceMember([]string{utils.CDRFIELD, utils.COMBIMED, utils.METATAG}, cdrField.Type) && len(cdrField.Value) != 0 {
		cdrField.CdrFieldId = cdrField.Value[0].Id
	}
	if tag != nil {
		cdrField.Tag = *tag
	} else {
		cdrField.Tag = cdrField.CdrFieldId
	}
	mandatoryFields := append(utils.PrimaryCdrFields, utils.CGRID, utils.COST, utils.MEDI_RUNID, utils.ORDERID)
	if mandatory != nil {
		cdrField.Mandatory = *mandatory
	} else if utils.IsSliceMember(mandatoryFields, cdrField.CdrFieldId) {
		cdrField.Mandatory = true
	}
	if layout != nil {
		cdrField.Layout = *layout
	} else if utils.IsSliceMember([]string{utils.SETUP_TIME, utils.ANSWER_TIME}, cdrField.CdrFieldId) {
		cdrField.Layout = "2006-01-02T15:04:05Z07:00"
	}
	if width != nil {
		cdrField.Width = *width
	} else if fixedWidth && cdrField.Type != utils.CONSTANT {
		switch cdrField.CdrFieldId { // First value element is used as field reference, giving the default properties out, good enough for default configs which do not have more than one value anyway
		case utils.CGRID:
			cdrField.Width = 40
		case utils.ORDERID:
			cdrField.Width = 11
		case utils.TOR:
			cdrField.Width = 6
		case utils.ACCID:
			cdrField.Width = 36
		case utils.CDRHOST:
			cdrField.Width = 15
		case utils.CDRSOURCE:
			cdrField.Width = 15
		case utils.REQTYPE:
			cdrField.Width = 13
		case utils.DIRECTION:
			cdrField.Width = 4
		case utils.TENANT:
			cdrField.Width = 24
		case utils.CATEGORY:
			cdrField.Width = 10
		case utils.ACCOUNT:
			cdrField.Width = 24
		case utils.SUBJECT:
			cdrField.Width = 24
		case utils.DESTINATION:
			cdrField.Width = 24
		case utils.SETUP_TIME:
			cdrField.Width = 30
		case utils.ANSWER_TIME:
			cdrField.Width = 30
		case utils.USAGE:
			cdrField.Width = 30
		case utils.MEDI_RUNID:
			cdrField.Width = 20
		case utils.COST:
			cdrField.Width = 24
		default:
			cdrField.Width = 30
		}
	}
	if strip != nil {
		cdrField.Strip = *strip
	} else if fixedWidth && cdrField.Type != utils.CONSTANT {
		switch cdrField.CdrFieldId { // First value element is used as field reference, giving the default properties out, good enough for default configs which do not have more than one value anyway
		case utils.CGRID, utils.ORDERID, utils.TOR:
		case utils.ACCID, utils.CDRHOST:
			cdrField.Strip = "left"
		case utils.CDRSOURCE, utils.REQTYPE, utils.DIRECTION, utils.TENANT, utils.CATEGORY, utils.ACCOUNT, utils.SUBJECT, utils.DESTINATION, utils.SETUP_TIME, utils.ANSWER_TIME, utils.USAGE, utils.MEDI_RUNID, utils.COST:
			cdrField.Strip = "xright"
		default:
			cdrField.Strip = "xright"
		}
	}
	if padding != nil {
		cdrField.Padding = *padding
	} else if fixedWidth && cdrField.Type != utils.CONSTANT {
		switch cdrField.CdrFieldId { // First value element is used as field reference, giving the default properties out, good enough for default configs which do not have more than one value anyway
		case utils.CGRID:
		default:
			cdrField.Padding = "left"
		}
	}
	return cdrField, nil
}

type CfgCdrField struct {
	Tag        string // Identifier for the administrator
	Type       string // Type of field
	CdrFieldId string // StoredCdr field name
	Value      utils.RSRFields
	Filter     utils.RSRFields
	Width      int
	Strip      string
	Padding    string
	Layout     string
	Mandatory  bool
}

// Converts a list of field identifiers into proper CDR field content
func NewCfgCdrFieldsFromIds(withFixedWith bool, fldsIds ...string) ([]*CfgCdrField, error) {
	cdrFields := make([]*CfgCdrField, len(fldsIds))
	for idx, fldId := range fldsIds {
		if parsedRsr, err := utils.NewRSRField(fldId); err != nil {
			return nil, err
		} else {
			if cdrFld, err := NewCfgCdrFieldWithDefaults(withFixedWith, utils.RSRFields{parsedRsr}, nil, nil, nil, nil, nil, nil, nil, nil, nil); err != nil {
				return nil, err
			} else {
				cdrFields[idx] = cdrFld
			}
		}
	}
	return cdrFields, nil
}
