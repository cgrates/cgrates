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

type ERsCfg struct {
	Enabled       bool
	SessionSConns []string
	Readers       []*EventReaderCfg
}

func (erS *ERsCfg) loadFromJsonCfg(jsnCfg *ERsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string, dfltRdrCfg *EventReaderCfg, separator string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		erS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		erS.SessionSConns = make([]string, len(*jsnCfg.Sessions_conns))
		for i, fID := range *jsnCfg.Sessions_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if fID == utils.MetaInternal {
				erS.SessionSConns[i] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)
			} else {
				erS.SessionSConns[i] = fID
			}
		}
	}
	return erS.appendERsReaders(jsnCfg.Readers, msgTemplates, sep, dfltRdrCfg)
}

func (ers *ERsCfg) appendERsReaders(jsnReaders *[]*EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string,
	dfltRdrCfg *EventReaderCfg) (err error) {
	if jsnReaders == nil {
		return
	}
	for _, jsnReader := range *jsnReaders {
		var rdr *EventReaderCfg
		if jsnReader.Id != nil {
			for _, reader := range ers.Readers {
				if reader.ID == *jsnReader.Id {
					rdr = reader
					break
				}
			}
		}
		if rdr == nil {
			if dfltRdrCfg != nil {
				rdr = dfltRdrCfg.Clone()
			} else {
				rdr = new(EventReaderCfg)
				rdr.Opts = make(map[string]interface{})
			}
			ers.Readers = append(ers.Readers, rdr)
		}
		if err := rdr.loadFromJsonCfg(jsnReader, msgTemplates, sep); err != nil {
			return err
		}

	}
	return nil
}

// Clone itself into a new ERsCfg
func (erS *ERsCfg) Clone() (cln *ERsCfg) {
	cln = new(ERsCfg)
	cln.Enabled = erS.Enabled
	cln.SessionSConns = make([]string, len(erS.SessionSConns))
	for idx, sConn := range erS.SessionSConns {
		cln.SessionSConns[idx] = sConn
	}
	cln.Readers = make([]*EventReaderCfg, len(erS.Readers))
	for idx, rdr := range erS.Readers {
		cln.Readers[idx] = rdr.Clone()
	}
	return
}

func (erS *ERsCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg:       erS.Enabled,
		utils.SessionSConnsCfg: erS.SessionSConns,
	}
	if erS.Readers != nil {
		readers := make([]map[string]interface{}, len(erS.Readers))
		for i, item := range erS.Readers {
			readers[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ReadersCfg] = readers
	}
	return
}

type EventReaderCfg struct {
	ID                       string
	Type                     string
	RowLength                int
	FieldSep                 string
	HeaderDefineChar         string
	RunDelay                 time.Duration
	ConcurrentReqs           int
	SourcePath               string
	ProcessedPath            string
	Opts                     map[string]interface{}
	XmlRootPath              utils.HierarchyPath
	Tenant                   RSRParsers
	Timezone                 string
	Filters                  []string
	Flags                    utils.FlagsWithParams
	FailedCallsPrefix        string        // Used in case of flatstore CDRs to avoid searching for BYE records
	PartialRecordCache       time.Duration // Duration to cache partial records when not pairing
	PartialCacheExpiryAction string
	Fields                   []*FCTemplate
	CacheDumpFields          []*FCTemplate
}

func (er *EventReaderCfg) loadFromJsonCfg(jsnCfg *EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Id != nil {
		er.ID = *jsnCfg.Id
	}
	if jsnCfg.Type != nil {
		er.Type = *jsnCfg.Type
	}
	if jsnCfg.Row_length != nil {
		er.RowLength = *jsnCfg.Row_length
	}
	if jsnCfg.Field_separator != nil {
		er.FieldSep = *jsnCfg.Field_separator
	}
	if jsnCfg.Header_define_character != nil {
		er.HeaderDefineChar = *jsnCfg.Header_define_character
	}
	if jsnCfg.Run_delay != nil {
		if er.RunDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.Run_delay); err != nil {
			return
		}
	}
	if jsnCfg.Concurrent_requests != nil {
		er.ConcurrentReqs = *jsnCfg.Concurrent_requests
	}
	if jsnCfg.Source_path != nil {
		er.SourcePath = *jsnCfg.Source_path
	}
	if jsnCfg.Processed_path != nil {
		er.ProcessedPath = *jsnCfg.Processed_path
	}
	if jsnCfg.Xml_root_path != nil {
		er.XmlRootPath = utils.ParseHierarchyPath(*jsnCfg.Xml_root_path, utils.EmptyString)
	}
	if jsnCfg.Tenant != nil {
		if er.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, sep); err != nil {
			return err
		}
	}
	if jsnCfg.Timezone != nil {
		er.Timezone = *jsnCfg.Timezone
	}
	if jsnCfg.Filters != nil {
		er.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			er.Filters[i] = fltr
		}
	}
	if jsnCfg.Flags != nil {
		er.Flags = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Failed_calls_prefix != nil {
		er.FailedCallsPrefix = *jsnCfg.Failed_calls_prefix
	}
	if jsnCfg.Partial_record_cache != nil {
		if er.PartialRecordCache, err = utils.ParseDurationWithNanosecs(*jsnCfg.Partial_record_cache); err != nil {
			return err
		}
	}
	if jsnCfg.Partial_cache_expiry_action != nil {
		er.PartialCacheExpiryAction = *jsnCfg.Partial_cache_expiry_action
	}
	if jsnCfg.Fields != nil {
		if er.Fields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Fields, sep); err != nil {
			return err
		}
		if tpls, err := InflateTemplates(er.Fields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			er.Fields = tpls
		}
	}
	if jsnCfg.Cache_dump_fields != nil {
		if er.CacheDumpFields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Cache_dump_fields, sep); err != nil {
			return err
		}
		if tpls, err := InflateTemplates(er.CacheDumpFields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			er.CacheDumpFields = tpls
		}
	}
	if jsnCfg.Opts != nil {
		for k, v := range jsnCfg.Opts {
			er.Opts[k] = v
		}
	}
	return
}

//Clone itself into a new EventReaderCfg
func (er *EventReaderCfg) Clone() (cln *EventReaderCfg) {
	cln = &EventReaderCfg{
		ID:                       er.ID,
		Type:                     er.Type,
		FieldSep:                 er.FieldSep,
		HeaderDefineChar:         er.HeaderDefineChar,
		RunDelay:                 er.RunDelay,
		ConcurrentReqs:           er.ConcurrentReqs,
		SourcePath:               er.SourcePath,
		ProcessedPath:            er.ProcessedPath,
		XmlRootPath:              er.XmlRootPath,
		PartialCacheExpiryAction: er.PartialCacheExpiryAction,
		PartialRecordCache:       er.PartialRecordCache,
		Opts:                     make(map[string]interface{}),
	}
	if len(er.Tenant) != 0 {
		cln.Tenant = make(RSRParsers, len(er.Tenant))
		for idx, val := range er.Tenant {
			clnVal := *val
			cln.Tenant[idx] = &clnVal
		}
	}
	cln.Timezone = er.Timezone
	if len(er.Filters) != 0 {
		cln.Filters = make([]string, len(er.Filters))
		for idx, val := range er.Filters {
			cln.Filters[idx] = val
		}
	}
	cln.Flags = er.Flags
	cln.FailedCallsPrefix = er.FailedCallsPrefix
	cln.Fields = make([]*FCTemplate, len(er.Fields))
	for idx, fld := range er.Fields {
		cln.Fields[idx] = fld.Clone()
	}
	cln.CacheDumpFields = make([]*FCTemplate, len(er.CacheDumpFields))
	for idx, fld := range er.CacheDumpFields {
		cln.CacheDumpFields[idx] = fld.Clone()
	}
	for k, v := range er.Opts {
		cln.Opts[k] = v
	}
	return
}

func (er *EventReaderCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.IDCfg:                       er.ID,
		utils.TypeCfg:                     er.Type,
		utils.RowLengthCfg:                er.RowLength,
		utils.FieldSepCfg:                 er.FieldSep,
		utils.HeaderDefCharCfg:            er.HeaderDefineChar,
		utils.ConcurrentReqsCfg:           er.ConcurrentReqs,
		utils.SourcePathCfg:               er.SourcePath,
		utils.ProcessedPathCfg:            er.ProcessedPath,
		utils.TimezoneCfg:                 er.Timezone,
		utils.FiltersCfg:                  []string{},
		utils.FlagsCfg:                    []string{},
		utils.FailedCallsPrefixCfg:        er.FailedCallsPrefix,
		utils.PartialCacheExpiryActionCfg: er.PartialCacheExpiryAction,
		utils.OptsCfg:                     er.Opts,
	}
	if flags := er.Flags.SliceFlags(); len(flags) != 0 {
		initialMP[utils.FlagsCfg] = flags
	}
	if len(er.Filters) != 0 {
		initialMP[utils.FiltersCfg] = er.Filters
	}

	if er.XmlRootPath != nil {
		xmlRootPath := make([]string, len(er.XmlRootPath))
		for i, item := range er.XmlRootPath {
			xmlRootPath[i] = item
		}
		initialMP[utils.XMLRootPathCfg] = xmlRootPath
	}
	var tenant string
	values := make([]string, len(er.Tenant))
	for i, item := range er.Tenant {
		values[i] = item.Rules
	}
	tenant = strings.Join(values, separator)
	initialMP[utils.TenantCfg] = tenant

	if er.Fields != nil {
		fields := make([]map[string]interface{}, len(er.Fields))
		for i, item := range er.Fields {
			fields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.FieldsCfg] = fields
	}
	if er.CacheDumpFields != nil {
		cacheDumpFields := make([]map[string]interface{}, len(er.CacheDumpFields))
		for i, item := range er.CacheDumpFields {
			cacheDumpFields[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.CacheDumpFieldsCfg] = cacheDumpFields
	}

	var runDelay string
	if er.RunDelay > 0 {
		runDelay = er.RunDelay.String()
	} else if er.RunDelay == 0 {
		runDelay = "0"
	} else {
		runDelay = "-1"
	}
	initialMP[utils.RunDelayCfg] = runDelay

	if er.PartialRecordCache != 0 {
		initialMP[utils.PartialRecordCacheCfg] = er.PartialRecordCache.String()
	} else {
		initialMP[utils.PartialRecordCacheCfg] = "0"
	}
	return
}
