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
func NewCdreCdrFieldsFromIds(withFixedWith bool, fldsIds ...string) ([]*CdreCdrField, error) {
	cdrFields := make([]*CdreCdrField, len(fldsIds))
	for idx, fldId := range fldsIds {
		if parsedRsr, err := utils.NewRSRField(fldId); err != nil {
			return nil, err
		} else {
			cdrFld := &CdreCdrField{Name: fldId, Type: utils.CDRFIELD, Value: fldId, valueAsRsrField: parsedRsr}
			if err := cdrFld.setDefaultFieldProperties(withFixedWith); err != nil { // Set default fixed width properties to be used later if needed
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
	if flds, err := NewCdreCdrFieldsFromIds(false, utils.CGRID, utils.MEDI_RUNID, utils.TOR, utils.ACCID, utils.REQTYPE, utils.DIRECTION, utils.TENANT,
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
	Filter          *utils.RSRField
	Mandatory       bool
	valueAsRsrField *utils.RSRField // Cached if the need arrises
}

func (cdrField *CdreCdrField) ValueAsRSRField() *utils.RSRField {
	return cdrField.valueAsRsrField
}

// Should be called on .fwv configuration without providing default values for fixed with parameters
func (cdrField *CdreCdrField) setDefaultFieldProperties(fixedWidth bool) error {
	if cdrField.valueAsRsrField == nil {
		return errors.New("Missing valueAsRsrField")
	}
	switch cdrField.valueAsRsrField.Id {
	case utils.CGRID:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 40
		}
	case utils.ORDERID:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 11
			cdrField.Padding = "left"
		}
	case utils.TOR:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 6
			cdrField.Padding = "left"
		}
	case utils.ACCID:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 36
			cdrField.Strip = "left"
			cdrField.Padding = "left"
		}
	case utils.CDRHOST:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 15
			cdrField.Strip = "left"
			cdrField.Padding = "left"
		}
	case utils.CDRSOURCE:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 15
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.REQTYPE:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 13
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.DIRECTION:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 4
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.TENANT:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 24
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.CATEGORY:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 10
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.ACCOUNT:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 24
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.SUBJECT:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 24
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.DESTINATION:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 24
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.SETUP_TIME:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 30
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
			cdrField.Layout = "2006-01-02T15:04:05Z07:00"
		}
	case utils.ANSWER_TIME:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 30
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
			cdrField.Layout = "2006-01-02T15:04:05Z07:00"
		}
	case utils.USAGE:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 30
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.MEDI_RUNID:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 20
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	case utils.COST:
		cdrField.Mandatory = true
		if fixedWidth {
			cdrField.Width = 24
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	default:
		cdrField.Mandatory = false
		if fixedWidth {
			cdrField.Width = 30
			cdrField.Strip = "xright"
			cdrField.Padding = "left"
		}
	}
	return nil
}
