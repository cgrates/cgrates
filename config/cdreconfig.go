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
	"errors"
	"github.com/cgrates/cgrates/utils"
)

// Converts a list of field identifiers into proper CDR field content
func NewCdreCdrFieldsFromIds(fldsIds ...string) ([]*CdreCdrField, error) {
	cdrFields := make([]*CdreCdrField, len(fldsIds))
	for idx, fldId := range fldsIds {
		if parsedRsr, err := utils.NewRSRField(fldId); err != nil {
			return nil, err
		} else {
			cdrFld := &CdreCdrField{Name: fldId, Type: utils.CDRFIELD, Value: fldId, valueAsRsrField: parsedRsr}
			if err := cdrFld.setDefaultFixedWidthProperties(); err != nil { // Set default fixed width properties to be used later if needed
				return nil, err
			}
			cdrFields[idx] = cdrFld
		}
	}
	return cdrFields, nil
}

func NewDefaultCdreConfig() (*CdreConfig, error) {
	cdreCfg := new(CdreConfig)
	if err := cdreCfg.setDefaults(); err != nil {
		return nil, err
	}
	return cdreCfg, nil
}

// One instance of CdrExporter
type CdreConfig struct {
	CdrFormat               string
	DataUsageMultiplyFactor float64
	CostMultiplyFactor      float64
	CostRoundingDecimals    int
	CostShiftDigits         int
	MaskDestId              string
	MaskLength              int
	ExportDir               string
	HeaderFields            []*CdreCdrField
	ContentFields           []*CdreCdrField
	TrailerFields           []*CdreCdrField
}

// Set here defaults
func (cdreCfg *CdreConfig) setDefaults() error {
	cdreCfg.CdrFormat = utils.CSV
	cdreCfg.DataUsageMultiplyFactor = 0.0
	cdreCfg.CostMultiplyFactor = 0.0
	cdreCfg.CostRoundingDecimals = -1
	cdreCfg.CostShiftDigits = 0
	cdreCfg.MaskDestId = ""
	cdreCfg.MaskLength = 0
	cdreCfg.ExportDir = "/var/log/cgrates/cdre"
	if flds, err := NewCdreCdrFieldsFromIds(utils.CGRID, utils.MEDI_RUNID, utils.TOR, utils.ACCID, utils.REQTYPE, utils.DIRECTION, utils.TENANT,
		utils.CATEGORY, utils.ACCOUNT, utils.SUBJECT, utils.DESTINATION, utils.SETUP_TIME, utils.ANSWER_TIME, utils.USAGE, utils.COST); err != nil {
		return err
	} else {
		cdreCfg.ContentFields = flds
	}
	return nil
}

type CdreCdrField struct {
	Name            string
	Type            string
	Value           string
	Width           int
	Strip           string
	Padding         string
	Layout          string
	Mandatory       bool
	valueAsRsrField *utils.RSRField // Cached if the need arrises
}

func (cdrField *CdreCdrField) ValueAsRSRField() *utils.RSRField {
	return cdrField.valueAsRsrField
}

// Should be called on .fwv configuration without providing default values for fixed with parameters
func (cdrField *CdreCdrField) setDefaultFixedWidthProperties() error {
	if cdrField.valueAsRsrField == nil {
		return errors.New("Missing valueAsRsrField")
	}
	switch cdrField.valueAsRsrField.Id {
	case utils.CGRID:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.ORDERID:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.TOR:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.ACCID:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.CDRHOST:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.CDRSOURCE:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.REQTYPE:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.DIRECTION:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.TENANT:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.CATEGORY:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.ACCOUNT:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.SUBJECT:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.DESTINATION:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.SETUP_TIME:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = "2006-01-02T15:04:05Z07:00"
		cdrField.Mandatory = true
	case utils.ANSWER_TIME:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = "2006-01-02T15:04:05Z07:00"
		cdrField.Mandatory = true
	case utils.USAGE:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.MEDI_RUNID:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	case utils.COST:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	default:
		cdrField.Width = 10
		cdrField.Strip = "xright"
		cdrField.Padding = ""
		cdrField.Layout = ""
		cdrField.Mandatory = true
	}
	return nil
}
