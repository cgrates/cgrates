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
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type CdrcCfg struct {
	ID                       string              // free-form text identifying this CDRC instance
	Enabled                  bool                // Enable/Disable the profile
	DryRun                   bool                // Do not post CDRs to the server
	CdrsConns                []*HaPoolConfig     // The address where CDRs can be reached
	CdrFormat                string              // The type of CDR file to process <*csv|*opensips_flatstore>
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
	Filters                  []string
	Tenant                   RSRParsers
	ContinueOnSuccess        bool          // Continue after execution
	PartialRecordCache       time.Duration // Duration to cache partial records when not pairing
	PartialCacheExpiryAction string
	HeaderFields             []*FCTemplate
	ContentFields            []*FCTemplate
	TrailerFields            []*FCTemplate
	CacheDumpFields          []*FCTemplate
}

func (self *CdrcCfg) loadFromJsonCfg(jsnCfg *CdrcJsonCfg, separator string) error {
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
		self.CdrFormat = strings.TrimPrefix(*jsnCfg.Cdr_format, "*")
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
	if jsnCfg.Filters != nil {
		self.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			self.Filters[i] = fltr
		}
	}
	if jsnCfg.Tenant != nil {
		if self.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, true, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Continue_on_success != nil {
		self.ContinueOnSuccess = *jsnCfg.Continue_on_success
	}
	if jsnCfg.Partial_record_cache != nil {
		if self.PartialRecordCache, err = utils.ParseDurationWithNanosecs(*jsnCfg.Partial_record_cache); err != nil {
			return err
		}
	}
	if jsnCfg.Partial_cache_expiry_action != nil {
		self.PartialCacheExpiryAction = *jsnCfg.Partial_cache_expiry_action
	}
	if jsnCfg.Header_fields != nil {
		if self.HeaderFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Header_fields, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Content_fields != nil {
		if self.ContentFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Content_fields, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Trailer_fields != nil {
		if self.TrailerFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Trailer_fields, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Cache_dump_fields != nil {
		if self.CacheDumpFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Cache_dump_fields, separator); err != nil {
			return err
		}
	}
	return nil
}

// Clone itself into a new CdrcCfg
func (self *CdrcCfg) Clone() *CdrcCfg {
	clnCdrc := new(CdrcCfg)
	clnCdrc.ID = self.ID
	clnCdrc.Enabled = self.Enabled
	clnCdrc.DryRun = self.DryRun
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
	clnCdrc.CDRPath = make(utils.HierarchyPath, len(self.CDRPath))
	for i, path := range self.CDRPath {
		clnCdrc.CDRPath[i] = path
	}
	clnCdrc.FailedCallsPrefix = self.FailedCallsPrefix
	clnCdrc.Filters = make([]string, len(self.Filters))
	for i, fltr := range self.Filters {
		clnCdrc.Filters[i] = fltr
	}
	clnCdrc.Tenant = self.Tenant
	clnCdrc.CdrSourceId = self.CdrSourceId
	clnCdrc.ContinueOnSuccess = self.ContinueOnSuccess
	clnCdrc.PartialRecordCache = self.PartialRecordCache
	clnCdrc.PartialCacheExpiryAction = self.PartialCacheExpiryAction
	clnCdrc.HeaderFields = make([]*FCTemplate, len(self.HeaderFields))
	clnCdrc.ContentFields = make([]*FCTemplate, len(self.ContentFields))
	clnCdrc.TrailerFields = make([]*FCTemplate, len(self.TrailerFields))
	clnCdrc.CacheDumpFields = make([]*FCTemplate, len(self.CacheDumpFields))
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
