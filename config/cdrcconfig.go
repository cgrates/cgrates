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
	ID                       string              // free-form text identifying this CDRC instance
	Enabled                  bool                // Enable/Disable the profile
	DryRun                   bool                // Do not post CDRs to the server
	CdrsConns                []*HaPoolConfig     // The address where CDRs can be reached
	CdrFormat                string              // The type of CDR file to process <csv|opensips_flatstore>
	FieldSeparator           rune                // The separator to use when reading csvs
	DataUsageMultiplyFactor  float64             // Conversion factor for data usage
	Timezone                 string              // timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	RunDelay                 time.Duration       // Delay between runs, 0 for inotify driven requests
	MaxOpenFiles             int                 // Maximum number of files opened simultaneously
	CdrInDir                 string              // Folder to process CDRs from
	CdrOutDir                string              // Folder to move processed CDRs to
	FailedCallsPrefix        string              // Used in case of flatstore CDRs to avoid searching for BYE records
	CDRPath                  utils.HierarchyPath // used for XML CDRs to specify the path towards CDR elements
	CdrSourceId              string              // Source identifier for the processed CDRs
	CdrFilter                utils.RSRFields     // Filter CDR records to import
	ContinueOnSuccess        bool                // Continue after execution
	PartialRecordCache       time.Duration       // Duration to cache partial records when not pairing
	PartialCacheExpiryAction string
	HeaderFields             []*CfgCdrField
	ContentFields            []*CfgCdrField
	TrailerFields            []*CfgCdrField
	CacheDumpFields          []*CfgCdrField
}

func (self *CdrcConfig) loadFromJsonCfg(jsnCfg *CdrcJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Id != nil {
		self.ID = *jsnCfg.Id
	}
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Dry_run != nil {
		self.DryRun = *jsnCfg.Dry_run
	}
	if jsnCfg.Cdrs_conns != nil {
		self.CdrsConns = make([]*HaPoolConfig, len(*jsnCfg.Cdrs_conns))
		for idx, jsnHaCfg := range *jsnCfg.Cdrs_conns {
			self.CdrsConns[idx] = NewDfltHaPoolConfig()
			self.CdrsConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
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
	if jsnCfg.Timezone != nil {
		self.Timezone = *jsnCfg.Timezone
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
	if jsnCfg.Cdr_path != nil {
		self.CDRPath = utils.ParseHierarchyPath(*jsnCfg.Cdr_path, "")
	}
	if jsnCfg.Cdr_source_id != nil {
		self.CdrSourceId = *jsnCfg.Cdr_source_id
	}
	if jsnCfg.Cdr_filter != nil {
		if self.CdrFilter, err = utils.ParseRSRFields(*jsnCfg.Cdr_filter, utils.INFIELD_SEP); err != nil {
			return err
		}
	}
	if jsnCfg.Continue_on_success != nil {
		self.ContinueOnSuccess = *jsnCfg.Continue_on_success
	}
	if jsnCfg.Partial_record_cache != nil {
		if self.PartialRecordCache, err = utils.ParseDurationWithSecs(*jsnCfg.Partial_record_cache); err != nil {
			return err
		}
	}
	if jsnCfg.Partial_cache_expiry_action != nil {
		self.PartialCacheExpiryAction = *jsnCfg.Partial_cache_expiry_action
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
	if jsnCfg.Cache_dump_fields != nil {
		if self.CacheDumpFields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Cache_dump_fields); err != nil {
			return err
		}
	}
	return nil
}

// Clone itself into a new CdrcConfig
func (self *CdrcConfig) Clone() *CdrcConfig {
	clnCdrc := new(CdrcConfig)
	clnCdrc.ID = self.ID
	clnCdrc.Enabled = self.Enabled
	clnCdrc.CdrsConns = make([]*HaPoolConfig, len(self.CdrsConns))
	for idx, cdrConn := range self.CdrsConns {
		clonedVal := *cdrConn
		clnCdrc.CdrsConns[idx] = &clonedVal
	}
	clnCdrc.CdrFormat = self.CdrFormat
	clnCdrc.FieldSeparator = self.FieldSeparator
	clnCdrc.DataUsageMultiplyFactor = self.DataUsageMultiplyFactor
	clnCdrc.Timezone = self.Timezone
	clnCdrc.RunDelay = self.RunDelay
	clnCdrc.MaxOpenFiles = self.MaxOpenFiles
	clnCdrc.CdrInDir = self.CdrInDir
	clnCdrc.CdrOutDir = self.CdrOutDir
	clnCdrc.CdrSourceId = self.CdrSourceId
	clnCdrc.PartialRecordCache = self.PartialRecordCache
	clnCdrc.PartialCacheExpiryAction = self.PartialCacheExpiryAction
	clnCdrc.HeaderFields = make([]*CfgCdrField, len(self.HeaderFields))
	clnCdrc.ContentFields = make([]*CfgCdrField, len(self.ContentFields))
	clnCdrc.TrailerFields = make([]*CfgCdrField, len(self.TrailerFields))
	clnCdrc.CacheDumpFields = make([]*CfgCdrField, len(self.CacheDumpFields))
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
	for idx, fld := range self.CacheDumpFields {
		clonedVal := *fld
		clnCdrc.CacheDumpFields[idx] = &clonedVal
	}
	return clnCdrc
}
