/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2015 ITsysCOM GmbH

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
	"time"

	"github.com/cgrates/cgrates/utils"
)

type CdrcConfig struct {
	Enabled                 bool            // Enable/Disable the profile
	Cdrs                    string          // The address where CDRs can be reached
	CdrFormat               string          // The type of CDR file to process <csv|opensips_flatstore>
	FieldSeparator          rune            // The separator to use when reading csvs
	DataUsageMultiplyFactor float64         // Conversion factor for data usage
	RunDelay                time.Duration   // Delay between runs, 0 for inotify driven requests
	MaxOpenFiles            int             // Maximum number of files opened simultaneously
	CdrInDir                string          // Folder to process CDRs from
	CdrOutDir               string          // Folder to move processed CDRs to
	FailedCallsPrefix       string          // Used in case of flatstore CDRs to avoid searching for BYE records
	CdrSourceId             string          // Source identifier for the processed CDRs
	CdrFilter               utils.RSRFields // Filter CDR records to import
	PartialRecordCache      time.Duration   // Duration to cache partial records when not pairing
	HeaderFields            []*CfgCdrField
	ContentFields           []*CfgCdrField
	TrailerFields           []*CfgCdrField
}

func (self *CdrcConfig) loadFromJsonCfg(jsnCfg *CdrcJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cdrs != nil {
		self.Cdrs = *jsnCfg.Cdrs
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
	if jsnCfg.Max_open_files != nil {
		self.MaxOpenFiles = *jsnCfg.Max_open_files
	}
	if jsnCfg.Cdr_in_dir != nil {
		self.CdrInDir = *jsnCfg.Cdr_in_dir
	}
	if jsnCfg.Cdr_out_dir != nil {
		self.CdrOutDir = *jsnCfg.Cdr_out_dir
	}
	if jsnCfg.Failed_calls_prefix != nil {
		self.FailedCallsPrefix = *jsnCfg.Failed_calls_prefix
	}
	if jsnCfg.Cdr_source_id != nil {
		self.CdrSourceId = *jsnCfg.Cdr_source_id
	}
	if jsnCfg.Cdr_filter != nil {
		if self.CdrFilter, err = utils.ParseRSRFields(*jsnCfg.Cdr_filter, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Partial_record_cache != nil {
		if self.PartialRecordCache, err = utils.ParseDurationWithSecs(*jsnCfg.Partial_record_cache); err != nil {
			return err
		}
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

// Clone itself into a new CdrcConfig
func (self *CdrcConfig) Clone() *CdrcConfig {
	clnCdrc := new(CdrcConfig)
	clnCdrc.Enabled = self.Enabled
	clnCdrc.Cdrs = self.Cdrs
	clnCdrc.CdrFormat = self.CdrFormat
	clnCdrc.FieldSeparator = self.FieldSeparator
	clnCdrc.DataUsageMultiplyFactor = self.DataUsageMultiplyFactor
	clnCdrc.RunDelay = self.RunDelay
	clnCdrc.MaxOpenFiles = self.MaxOpenFiles
	clnCdrc.CdrInDir = self.CdrInDir
	clnCdrc.CdrOutDir = self.CdrOutDir
	clnCdrc.CdrSourceId = self.CdrSourceId
	clnCdrc.HeaderFields = make([]*CfgCdrField, len(self.HeaderFields))
	clnCdrc.ContentFields = make([]*CfgCdrField, len(self.ContentFields))
	clnCdrc.TrailerFields = make([]*CfgCdrField, len(self.TrailerFields))
	for idx, fld := range self.HeaderFields {
		clonedVal := *fld
		clnCdrc.HeaderFields[idx] = &clonedVal
	}
	for idx, fld := range self.ContentFields {
		clonedVal := *fld
		clnCdrc.ContentFields[idx] = &clonedVal
	}
	for idx, fld := range self.TrailerFields {
		clonedVal := *fld
		clnCdrc.TrailerFields[idx] = &clonedVal
	}
	return clnCdrc
}
