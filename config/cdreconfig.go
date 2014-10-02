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
)

func NewDefaultCdreConfig() *CdreConfig {
	cdreCfg := new(CdreConfig)
	cdreCfg.setDefaults()
	return cdreCfg
}

func NewCdreConfigFromXmlCdreCfg(xmlCdreCfg *CgrXmlCdreCfg) (*CdreConfig, error) {
	var err error
	cdreCfg := NewDefaultCdreConfig()
	if xmlCdreCfg.CdrFormat != nil {
		cdreCfg.CdrFormat = *xmlCdreCfg.CdrFormat
	}
	if xmlCdreCfg.FieldSeparator != nil && len(*xmlCdreCfg.FieldSeparator) == 1 {
		sepStr := *xmlCdreCfg.FieldSeparator
		cdreCfg.FieldSeparator = rune(sepStr[0])
	}
	if xmlCdreCfg.DataUsageMultiplyFactor != nil {
		cdreCfg.DataUsageMultiplyFactor = *xmlCdreCfg.DataUsageMultiplyFactor
	}
	if xmlCdreCfg.CostMultiplyFactor != nil {
		cdreCfg.CostMultiplyFactor = *xmlCdreCfg.CostMultiplyFactor
	}
	if xmlCdreCfg.CostRoundingDecimals != nil {
		cdreCfg.CostRoundingDecimals = *xmlCdreCfg.CostRoundingDecimals
	}
	if xmlCdreCfg.CostShiftDigits != nil {
		cdreCfg.CostShiftDigits = *xmlCdreCfg.CostShiftDigits
	}
	if xmlCdreCfg.MaskDestId != nil {
		cdreCfg.MaskDestId = *xmlCdreCfg.MaskDestId
	}
	if xmlCdreCfg.MaskLength != nil {
		cdreCfg.MaskLength = *xmlCdreCfg.MaskLength
	}
	if xmlCdreCfg.ExportDir != nil {
		cdreCfg.ExportDir = *xmlCdreCfg.ExportDir
	}
	if xmlCdreCfg.Header != nil {
		cdreCfg.HeaderFields = make([]*CfgCdrField, len(xmlCdreCfg.Header.Fields))
		for idx, xmlFld := range xmlCdreCfg.Header.Fields {
			cdreCfg.HeaderFields[idx], err = NewCfgCdrFieldFromCgrXmlCfgCdrField(xmlFld, cdreCfg.CdrFormat == utils.CDRE_FIXED_WIDTH)
			if err != nil {
				return nil, err
			}
		}
	}
	if xmlCdreCfg.Content != nil {
		cdreCfg.ContentFields = make([]*CfgCdrField, len(xmlCdreCfg.Content.Fields))
		for idx, xmlFld := range xmlCdreCfg.Content.Fields {
			cdreCfg.ContentFields[idx], err = NewCfgCdrFieldFromCgrXmlCfgCdrField(xmlFld, cdreCfg.CdrFormat == utils.CDRE_FIXED_WIDTH)
			if err != nil {
				return nil, err
			}
		}
	}
	if xmlCdreCfg.Trailer != nil {
		cdreCfg.TrailerFields = make([]*CfgCdrField, len(xmlCdreCfg.Trailer.Fields))
		for idx, xmlFld := range xmlCdreCfg.Trailer.Fields {
			cdreCfg.TrailerFields[idx], err = NewCfgCdrFieldFromCgrXmlCfgCdrField(xmlFld, cdreCfg.CdrFormat == utils.CDRE_FIXED_WIDTH)
			if err != nil {
				return nil, err
			}
		}
	}
	return cdreCfg, nil
}

// One instance of CdrExporter
type CdreConfig struct {
	CdrFormat               string
	FieldSeparator          rune
	DataUsageMultiplyFactor float64
	CostMultiplyFactor      float64
	CostRoundingDecimals    int
	CostShiftDigits         int
	MaskDestId              string
	MaskLength              int
	ExportDir               string
	HeaderFields            []*CfgCdrField
	ContentFields           []*CfgCdrField
	TrailerFields           []*CfgCdrField
}

// Set here defaults
func (cdreCfg *CdreConfig) setDefaults() error {
	cdreCfg.CdrFormat = utils.CSV
	cdreCfg.FieldSeparator = utils.CSV_SEP
	cdreCfg.DataUsageMultiplyFactor = 0.0
	cdreCfg.CostMultiplyFactor = 0.0
	cdreCfg.CostRoundingDecimals = -1
	cdreCfg.CostShiftDigits = 0
	cdreCfg.MaskDestId = ""
	cdreCfg.MaskLength = 0
	cdreCfg.ExportDir = "/var/log/cgrates/cdre"
	if flds, err := NewCfgCdrFieldsFromIds(false, utils.CGRID, utils.MEDI_RUNID, utils.TOR, utils.ACCID, utils.REQTYPE, utils.DIRECTION, utils.TENANT,
		utils.CATEGORY, utils.ACCOUNT, utils.SUBJECT, utils.DESTINATION, utils.SETUP_TIME, utils.ANSWER_TIME, utils.USAGE, utils.COST); err != nil {
		return err
	} else {
		cdreCfg.ContentFields = flds
	}
	return nil
}
