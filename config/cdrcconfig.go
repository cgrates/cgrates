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
	"time"
)

type CdrcConfig struct {
	Enabled                 bool            // Enable/Disable the profile
	CdrsAddress             string          // The address where CDRs can be reached
	CdrFormat               string          // The type of CDR file to process <csv>
	FieldSeparator          rune            // The separator to use when reading csvs
	DataUsageMultiplyFactor float64         // Conversion factor for data usage
	RunDelay                time.Duration   // Delay between runs, 0 for inotify driven requests
	CdrInDir                string          // Folder to process CDRs from
	CdrOutDir               string          // Folder to move processed CDRs to
	CdrSourceId             string          // Source identifier for the processed CDRs
	CdrFilter               utils.RSRFields // Filter CDR records to import
	CdrFields               []*CfgCdrField  // List of fields to be processed
}

func (self *CdrcConfig) loadFromJsonCfg(jsnCfg *CdrcJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cdrs_address != nil {
		self.CdrsAddress = *jsnCfg.Cdrs_address
	}
	if jsnCfg.Cdrs_address != nil {
		self.CdrsAddress = *jsnCfg.Cdrs_address
	}
	if jsnCfg.Cdr_format != nil {
		self.CdrFormat = *jsnCfg.Cdr_format
	}
	if jsnCfg.Field_separator != nil && len(*jsnCfg.Field_separator) > 0 {
		sepStr := *jsnCfg.Field_separator
		self.FieldSeparator = rune(sepStr[0])
	}
	if jsnCfg.Data_usage_multiply_factor != nil {
		self.DataUsageMultiplyFactor = *jsnCfg.Data_usage_multiply_factor
	}
	if jsnCfg.Run_delay != nil {
		self.RunDelay = time.Duration(*jsnCfg.Run_delay) * time.Second
	}
	if jsnCfg.Cdr_in_dir != nil {
		self.CdrInDir = *jsnCfg.Cdr_in_dir
	}
	if jsnCfg.Cdr_out_dir != nil {
		self.CdrOutDir = *jsnCfg.Cdr_out_dir
	}
	if jsnCfg.Cdr_source_id != nil {
		self.CdrSourceId = *jsnCfg.Cdr_source_id
	}
	if jsnCfg.Cdr_filter != nil {
		if self.CdrFilter, err = utils.ParseRSRFields(*jsnCfg.Cdr_filter, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Cdr_fields != nil {
		if self.CdrFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Cdr_fields); err != nil {
			return err
		}
	}
	return nil
}

// Clone itself into a new CdrcConfig
func (self *CdrcConfig) Clone() *CdrcConfig {
	clnCdrc := new(CdrcConfig)
	clnCdrc.Enabled = self.Enabled
	clnCdrc.CdrsAddress = self.CdrsAddress
	clnCdrc.CdrFormat = self.CdrFormat
	clnCdrc.FieldSeparator = self.FieldSeparator
	clnCdrc.DataUsageMultiplyFactor = self.DataUsageMultiplyFactor
	clnCdrc.RunDelay = self.RunDelay
	clnCdrc.CdrInDir = self.CdrInDir
	clnCdrc.CdrOutDir = self.CdrOutDir
	clnCdrc.CdrSourceId = self.CdrSourceId
	clnCdrc.CdrFields = make([]*CfgCdrField, len(self.CdrFields))
	for idx, fld := range self.CdrFields {
		clonedVal := *fld
		clnCdrc.CdrFields[idx] = &clonedVal
	}
	return clnCdrc
}
