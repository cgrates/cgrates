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
	"time"

	"github.com/cgrates/cgrates/utils"
)

// ERsCfg the config for ERs
type ERsCfg struct {
	Enabled       bool
	SessionSConns []string
	Readers       []*EventReaderCfg
}

func (erS *ERsCfg) loadFromJSONCfg(jsnCfg *ERsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string, dfltRdrCfg *EventReaderCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		erS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Sessions_conns != nil {
		erS.SessionSConns = updateInternalConns(*jsnCfg.Sessions_conns, utils.MetaSessionS)
	}
	return erS.appendERsReaders(jsnCfg.Readers, msgTemplates, sep, dfltRdrCfg)
}

func (erS *ERsCfg) appendERsReaders(jsnReaders *[]*EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string,
	dfltRdrCfg *EventReaderCfg) (err error) {
	if jsnReaders == nil {
		return
	}
	for _, jsnReader := range *jsnReaders {
		var rdr *EventReaderCfg
		if jsnReader.Id != nil {
			for _, reader := range erS.Readers {
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
			erS.Readers = append(erS.Readers, rdr)
		}
		if err := rdr.loadFromJSONCfg(jsnReader, msgTemplates, sep); err != nil {
			return err
		}

	}
	return nil
}

// Clone returns a deep copy of ERsCfg
func (erS *ERsCfg) Clone() (cln *ERsCfg) {
	cln = &ERsCfg{
		Enabled: erS.Enabled,
		Readers: make([]*EventReaderCfg, len(erS.Readers)),
	}
	if erS.SessionSConns != nil {
		cln.SessionSConns = utils.CloneStringSlice(erS.SessionSConns)
	}
	for idx, rdr := range erS.Readers {
		cln.Readers[idx] = rdr.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (erS *ERsCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg: erS.Enabled,
	}
	if erS.SessionSConns != nil {
		initialMP[utils.SessionSConnsCfg] = getInternalJSONConns(erS.SessionSConns)
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

// EventReaderCfg the event for the Event Reader
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
	XMLRootPath              utils.HierarchyPath
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

func (er *EventReaderCfg) loadFromJSONCfg(jsnCfg *EventReaderJsonCfg, msgTemplates map[string][]*FCTemplate, sep string) (err error) {
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
		er.XMLRootPath = utils.ParseHierarchyPath(*jsnCfg.Xml_root_path, utils.EmptyString)
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
		if er.Fields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Fields, sep); err != nil {
			return err
		}
		if tpls, err := InflateTemplates(er.Fields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			er.Fields = tpls
		}
	}
	if jsnCfg.Cache_dump_fields != nil {
		if er.CacheDumpFields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Cache_dump_fields, sep); err != nil {
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

// Clone returns a deep copy of EventReaderCfg
func (er EventReaderCfg) Clone() (cln *EventReaderCfg) {
	cln = &EventReaderCfg{
		ID:                       er.ID,
		Type:                     er.Type,
		FieldSep:                 er.FieldSep,
		HeaderDefineChar:         er.HeaderDefineChar,
		RunDelay:                 er.RunDelay,
		ConcurrentReqs:           er.ConcurrentReqs,
		SourcePath:               er.SourcePath,
		ProcessedPath:            er.ProcessedPath,
		XMLRootPath:              er.XMLRootPath.Clone(),
		Tenant:                   er.Tenant.Clone(),
		Timezone:                 er.Timezone,
		Flags:                    er.Flags.Clone(),
		FailedCallsPrefix:        er.FailedCallsPrefix,
		PartialCacheExpiryAction: er.PartialCacheExpiryAction,
		PartialRecordCache:       er.PartialRecordCache,
		Opts:                     make(map[string]interface{}),
	}
	if er.Filters != nil {
		cln.Filters = utils.CloneStringSlice(er.Filters)
	}
	if er.Fields != nil {
		cln.Fields = make([]*FCTemplate, len(er.Fields))
		for idx, fld := range er.Fields {
			cln.Fields[idx] = fld.Clone()
		}
	}
	if er.CacheDumpFields != nil {
		cln.CacheDumpFields = make([]*FCTemplate, len(er.CacheDumpFields))
		for idx, fld := range er.CacheDumpFields {
			cln.CacheDumpFields[idx] = fld.Clone()
		}
	}
	for k, v := range er.Opts {
		cln.Opts[k] = v
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (er *EventReaderCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.IDCfg:                       er.ID,
		utils.TypeCfg:                     er.Type,
		utils.RowLengthCfg:                er.RowLength,
		utils.FieldSepCfg:                 er.FieldSep,
		utils.HeaderDefCharCfg:            er.HeaderDefineChar,
		utils.ConcurrentRequestsCfg:       er.ConcurrentReqs,
		utils.SourcePathCfg:               er.SourcePath,
		utils.ProcessedPathCfg:            er.ProcessedPath,
		utils.TenantCfg:                   er.Tenant.GetRule(separator),
		utils.XMLRootPathCfg:              er.XMLRootPath.AsString("/", len(er.XMLRootPath) != 0 && len(er.XMLRootPath[0]) != 0),
		utils.TimezoneCfg:                 er.Timezone,
		utils.FiltersCfg:                  er.Filters,
		utils.FlagsCfg:                    []string{},
		utils.FailedCallsPrefixCfg:        er.FailedCallsPrefix,
		utils.PartialCacheExpiryActionCfg: er.PartialCacheExpiryAction,
		utils.PartialRecordCacheCfg:       "0",
		utils.RunDelayCfg:                 "0",
	}

	opts := make(map[string]interface{})
	for k, v := range er.Opts {
		opts[k] = v
	}
	initialMP[utils.OptsCfg] = opts

	if flags := er.Flags.SliceFlags(); flags != nil {
		initialMP[utils.FlagsCfg] = flags
	}

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

	if er.RunDelay > 0 {
		initialMP[utils.RunDelayCfg] = er.RunDelay.String()
	} else if er.RunDelay < 0 {
		initialMP[utils.RunDelayCfg] = "-1"
	}

	if er.PartialRecordCache != 0 {
		initialMP[utils.PartialRecordCacheCfg] = er.PartialRecordCache.String()
	}
	return
}

// EventReaderSJsonCfg is the configuration of a single EventReader
type EventReaderJsonCfg struct {
	Id                          *string
	Type                        *string
	Row_length                  *int
	Field_separator             *string
	Header_define_character     *string
	Run_delay                   *string
	Concurrent_requests         *int
	Source_path                 *string
	Processed_path              *string
	Opts                        map[string]interface{}
	Xml_root_path               *string
	Tenant                      *string
	Timezone                    *string
	Filters                     *[]string
	Flags                       *[]string
	Failed_calls_prefix         *string
	Partial_record_cache        *string
	Partial_cache_expiry_action *string
	Fields                      *[]*FcTemplateJsonCfg
	Cache_dump_fields           *[]*FcTemplateJsonCfg
}

func diffEventReaderJsonCfg(d *EventReaderJsonCfg, v1, v2 *EventReaderCfg, separator string) *EventReaderJsonCfg {
	if d == nil {
		d = new(EventReaderJsonCfg)
	}
	if v1.ID != v2.ID {
		d.Id = utils.StringPointer(v2.ID)
	}
	if v1.Type != v2.Type {
		d.Type = utils.StringPointer(v2.Type)
	}
	if v1.RowLength != v2.RowLength {
		d.Row_length = utils.IntPointer(v2.RowLength)
	}
	if v1.FieldSep != v2.FieldSep {
		d.Field_separator = utils.StringPointer(v2.FieldSep)
	}
	if v1.HeaderDefineChar != v2.HeaderDefineChar {
		d.Header_define_character = utils.StringPointer(v2.HeaderDefineChar)
	}
	if v1.RunDelay != v2.RunDelay {
		d.Run_delay = utils.StringPointer(v2.RunDelay.String())
	}
	if v1.ConcurrentReqs != v2.ConcurrentReqs {
		d.Concurrent_requests = utils.IntPointer(v2.ConcurrentReqs)
	}
	if v1.SourcePath != v2.SourcePath {
		d.Source_path = utils.StringPointer(v2.SourcePath)
	}
	if v1.ProcessedPath != v2.ProcessedPath {
		d.Processed_path = utils.StringPointer(v2.ProcessedPath)
	}
	d.Opts = diffMap(d.Opts, v1.Opts, v2.Opts)
	xml1 := v1.XMLRootPath.AsString("/", len(v1.XMLRootPath) != 0 && len(v1.XMLRootPath[0]) != 0)
	xml2 := v2.XMLRootPath.AsString("/", len(v2.XMLRootPath) != 0 && len(v2.XMLRootPath[0]) != 0)
	if xml1 != xml2 {
		d.Xml_root_path = utils.StringPointer(xml2)
	}
	tnt1 := v1.Tenant.GetRule(separator)
	tnt2 := v2.Tenant.GetRule(separator)
	if tnt1 != tnt2 {
		d.Tenant = utils.StringPointer(tnt2)
	}
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	if !utils.SliceStringEqual(v1.Filters, v2.Filters) {
		d.Filters = &v2.Filters
	}
	flgs1 := v1.Flags.SliceFlags()
	flgs2 := v2.Flags.SliceFlags()
	if !utils.SliceStringEqual(flgs1, flgs2) {
		d.Flags = &flgs2
	}
	if v1.FailedCallsPrefix != v2.FailedCallsPrefix {
		d.Failed_calls_prefix = utils.StringPointer(v2.FailedCallsPrefix)
	}
	if v1.PartialRecordCache != v2.PartialRecordCache {
		d.Partial_record_cache = utils.StringPointer(v2.PartialRecordCache.String())
	}
	if v1.PartialCacheExpiryAction != v2.PartialCacheExpiryAction {
		d.Partial_cache_expiry_action = utils.StringPointer(v2.PartialCacheExpiryAction)
	}
	var flds []*FcTemplateJsonCfg
	if d.Fields != nil {
		flds = *d.Fields
	}
	flds = diffFcTemplateJsonCfg(flds, v1.Fields, v2.Fields, separator)
	d.Fields = &flds

	var cdf []*FcTemplateJsonCfg
	if d.Cache_dump_fields != nil {
		cdf = *d.Cache_dump_fields
	}
	cdf = diffFcTemplateJsonCfg(cdf, v1.CacheDumpFields, v2.CacheDumpFields, separator)
	d.Cache_dump_fields = &cdf
	return d
}

func getEventReaderJsonCfg(d []*EventReaderJsonCfg, id string) (*EventReaderJsonCfg, int) {
	for i, v := range d {
		if v.Id != nil && *v.Id == id {
			return v, i
		}
	}
	return nil, -1
}

func getEventReaderCfg(d []*EventReaderCfg, id string) *EventReaderCfg {
	for _, v := range d {
		if v.ID == id {
			return v
		}
	}
	return new(EventReaderCfg)
}

func diffEventReadersJsonCfg(d *[]*EventReaderJsonCfg, v1, v2 []*EventReaderCfg, separator string) *[]*EventReaderJsonCfg {
	if d == nil || *d == nil {
		d = &[]*EventReaderJsonCfg{}
	}
	for _, val := range v2 {
		dv, i := getEventReaderJsonCfg(*d, val.ID)
		dv = diffEventReaderJsonCfg(dv, getEventReaderCfg(v1, val.ID), val, separator)
		if i == -1 {
			*d = append(*d, dv)
		} else {
			(*d)[i] = dv
		}
	}

	return d
}

// EventReaderSJsonCfg contains the configuration of EventReaderService
type ERsJsonCfg struct {
	Enabled        *bool
	Sessions_conns *[]string
	Readers        *[]*EventReaderJsonCfg
}

func diffERsJsonCfg(d *ERsJsonCfg, v1, v2 *ERsCfg, separator string) *ERsJsonCfg {
	if d == nil {
		d = new(ERsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.SessionSConns, v2.SessionSConns) {
		d.Sessions_conns = utils.SliceStringPointer(getInternalJSONConns(v2.SessionSConns))
	}
	d.Readers = diffEventReadersJsonCfg(d.Readers, v1.Readers, v2.Readers, separator)
	return d
}
