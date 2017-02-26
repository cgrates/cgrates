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

// One instance of CdrExporter
type CdreConfig struct {
	ExportFormat        string
	ExportPath          string
	FallbackPath        string
	CDRFilter           utils.RSRFields
	Synchronous         bool
	Attempts            int
	FieldSeparator      rune
	UsageMultiplyFactor utils.FieldMultiplyFactor
	CostMultiplyFactor  float64
	HeaderFields        []*CfgCdrField
	ContentFields       []*CfgCdrField
	TrailerFields       []*CfgCdrField
}

func (self *CdreConfig) loadFromJsonCfg(jsnCfg *CdreJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Export_format != nil {
		self.ExportFormat = *jsnCfg.Export_format
	}
	if jsnCfg.Export_path != nil {
		self.ExportPath = *jsnCfg.Export_path
	}
	if jsnCfg.Cdr_filter != nil {
		if self.CDRFilter, err = utils.ParseRSRFields(*jsnCfg.Cdr_filter, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Synchronous != nil {
		self.Synchronous = *jsnCfg.Synchronous
	}
	if jsnCfg.Attempts != nil {
		self.Attempts = *jsnCfg.Attempts
	}
	if jsnCfg.Field_separator != nil && len(*jsnCfg.Field_separator) > 0 { // Make sure we got at least one character so we don't get panic here
		sepStr := *jsnCfg.Field_separator
		self.FieldSeparator = rune(sepStr[0])
	}
	if jsnCfg.Usage_multiply_factor != nil {
		if self.UsageMultiplyFactor == nil { // not yet initialized
			self.UsageMultiplyFactor = make(map[string]float64, len(*jsnCfg.Usage_multiply_factor))
		}
		for k, v := range *jsnCfg.Usage_multiply_factor {
			self.UsageMultiplyFactor[k] = v
		}
	}
	if jsnCfg.Cost_multiply_factor != nil {
		self.CostMultiplyFactor = *jsnCfg.Cost_multiply_factor
	}
	if jsnCfg.Header_fields != nil {
		if self.HeaderFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Header_fields); err != nil {
			return err
		}
	}
	if jsnCfg.Content_fields != nil {
		if self.ContentFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Content_fields); err != nil {
			return err
		}
	}
	if jsnCfg.Trailer_fields != nil {
		if self.TrailerFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Trailer_fields); err != nil {
			return err
		}
	}
	return nil
}

// Clone itself into a new CdreConfig
func (self *CdreConfig) Clone() *CdreConfig {
	clnCdre := new(CdreConfig)
	clnCdre.ExportFormat = self.ExportFormat
	clnCdre.ExportPath = self.ExportPath
	clnCdre.Synchronous = self.Synchronous
	clnCdre.Attempts = self.Attempts
	clnCdre.FieldSeparator = self.FieldSeparator
	clnCdre.UsageMultiplyFactor = make(map[string]float64, len(self.UsageMultiplyFactor))
	for k, v := range self.UsageMultiplyFactor {
		clnCdre.UsageMultiplyFactor[k] = v
	}
	clnCdre.CostMultiplyFactor = self.CostMultiplyFactor
	clnCdre.HeaderFields = make([]*CfgCdrField, len(self.HeaderFields))
	for idx, fld := range self.HeaderFields {
		clonedVal := *fld
		clnCdre.HeaderFields[idx] = &clonedVal
	}
	clnCdre.ContentFields = make([]*CfgCdrField, len(self.ContentFields))
	for idx, fld := range self.ContentFields {
		clonedVal := *fld
		clnCdre.ContentFields[idx] = &clonedVal
	}
	clnCdre.TrailerFields = make([]*CfgCdrField, len(self.TrailerFields))
	for idx, fld := range self.TrailerFields {
		clonedVal := *fld
		clnCdre.TrailerFields[idx] = &clonedVal
	}
	return clnCdre
}
