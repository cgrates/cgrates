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

// EEsCfg the config for Event Exporters
type EEsCfg struct {
	Enabled         bool
	AttributeSConns []string
	Cache           map[string]*CacheParamCfg
	Exporters       []*EventExporterCfg
}

// GetDefaultExporter returns the exporter with the *default id
func (eeS *EEsCfg) GetDefaultExporter() *EventExporterCfg {
	for _, es := range eeS.Exporters {
		if es.ID == utils.MetaDefault {
			return es
		}
	}
	return nil
}

func (eeS *EEsCfg) loadFromJSONCfg(jsnCfg *EEsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string, dfltExpCfg *EventExporterCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		eeS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cache != nil {
		for kJsn, vJsn := range *jsnCfg.Cache {
			val := new(CacheParamCfg)
			if err := val.loadFromJSONCfg(vJsn); err != nil {
				return err
			}
			eeS.Cache[kJsn] = val
		}
	}
	if jsnCfg.Attributes_conns != nil {
		eeS.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for i, fID := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			eeS.AttributeSConns[i] = fID
			if fID == utils.MetaInternal {
				eeS.AttributeSConns[i] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	return eeS.appendEEsExporters(jsnCfg.Exporters, msgTemplates, sep, dfltExpCfg)
}

func (eeS *EEsCfg) appendEEsExporters(exporters *[]*EventExporterJsonCfg, msgTemplates map[string][]*FCTemplate, separator string, dfltExpCfg *EventExporterCfg) (err error) {
	if exporters == nil {
		return
	}
	for _, jsnExp := range *exporters {
		var exp *EventExporterCfg
		if jsnExp.Id != nil {
			for _, exporter := range eeS.Exporters {
				if exporter.ID == *jsnExp.Id {
					exp = exporter
					break
				}
			}
		}
		if exp == nil {
			if dfltExpCfg != nil {
				exp = dfltExpCfg.Clone()
			} else {
				exp = new(EventExporterCfg)
				exp.Opts = make(map[string]interface{})
			}
			eeS.Exporters = append(eeS.Exporters, exp)
		}
		if err = exp.loadFromJSONCfg(jsnExp, msgTemplates, separator); err != nil {
			return
		}
	}
	return
}

// Clone returns a deep copy of EEsCfg
func (eeS *EEsCfg) Clone() (cln *EEsCfg) {
	cln = &EEsCfg{
		Enabled:         eeS.Enabled,
		AttributeSConns: make([]string, len(eeS.AttributeSConns)),
		Cache:           make(map[string]*CacheParamCfg),
		Exporters:       make([]*EventExporterCfg, len(eeS.Exporters)),
	}
	for idx, sConn := range eeS.AttributeSConns {
		cln.AttributeSConns[idx] = sConn
	}
	for key, value := range eeS.Cache {
		cln.Cache[key] = value.Clone()
	}
	for idx, exp := range eeS.Exporters {
		cln.Exporters[idx] = exp.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (eeS *EEsCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg: eeS.Enabled,
	}
	if eeS.AttributeSConns != nil {
		attributeSConns := make([]string, len(eeS.AttributeSConns))
		for i, item := range eeS.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	if eeS.Cache != nil {
		cache := make(map[string]interface{}, len(eeS.Cache))
		for key, value := range eeS.Cache {
			cache[key] = value.AsMapInterface()
		}
		initialMP[utils.CacheCfg] = cache
	}
	if eeS.Exporters != nil {
		exporters := make([]map[string]interface{}, len(eeS.Exporters))
		for i, item := range eeS.Exporters {
			exporters[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ExportersCfg] = exporters
	}
	return
}

// EventExporterCfg the config for a Event Exporter
type EventExporterCfg struct {
	ID                 string
	Type               string
	ExportPath         string
	Opts               map[string]interface{}
	Timezone           string
	Filters            []string
	Flags              utils.FlagsWithParams
	AttributeSIDs      []string // selective AttributeS profiles
	AttributeSCtx      string   // context to use when querying AttributeS
	Synchronous        bool
	Attempts           int
	ConcurrentRequests int
	Fields             []*FCTemplate
	headerFields       []*FCTemplate
	contentFields      []*FCTemplate
	trailerFields      []*FCTemplate
}

func (eeC *EventExporterCfg) loadFromJSONCfg(jsnEec *EventExporterJsonCfg, msgTemplates map[string][]*FCTemplate, separator string) (err error) {
	if jsnEec == nil {
		return
	}
	if jsnEec.Id != nil {
		eeC.ID = *jsnEec.Id
	}
	if jsnEec.Type != nil {
		eeC.Type = *jsnEec.Type
	}
	if jsnEec.Export_path != nil {
		eeC.ExportPath = *jsnEec.Export_path
	}
	if jsnEec.Timezone != nil {
		eeC.Timezone = *jsnEec.Timezone
	}
	if jsnEec.Filters != nil {
		eeC.Filters = make([]string, len(*jsnEec.Filters))
		for i, fltr := range *jsnEec.Filters {
			eeC.Filters[i] = fltr
		}
	}
	if jsnEec.Flags != nil {
		eeC.Flags = utils.FlagsWithParamsFromSlice(*jsnEec.Flags)
	}
	if jsnEec.Attribute_context != nil {
		eeC.AttributeSCtx = *jsnEec.Attribute_context
	}
	if jsnEec.Attribute_ids != nil {
		eeC.AttributeSIDs = make([]string, len(*jsnEec.Attribute_ids))
		for i, fltr := range *jsnEec.Attribute_ids {
			eeC.AttributeSIDs[i] = fltr
		}
	}
	if jsnEec.Synchronous != nil {
		eeC.Synchronous = *jsnEec.Synchronous
	}
	if jsnEec.Attempts != nil {
		eeC.Attempts = *jsnEec.Attempts
	}
	if jsnEec.Concurrent_requests != nil {
		eeC.ConcurrentRequests = *jsnEec.Concurrent_requests
	}
	if jsnEec.Fields != nil {
		eeC.Fields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnEec.Fields, separator)
		if err != nil {
			return
		}
		if tpls, err := InflateTemplates(eeC.Fields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			eeC.Fields = tpls
		}
		eeC.ComputeFields()
	}
	if jsnEec.Opts != nil {
		for k, v := range jsnEec.Opts {
			eeC.Opts[k] = v
		}
	}
	return
}

// ComputeFields will split the fields in header trailer or content
// exported for ees testing
func (eeC *EventExporterCfg) ComputeFields() {
	eeC.headerFields = make([]*FCTemplate, 0)
	eeC.contentFields = make([]*FCTemplate, 0)
	eeC.trailerFields = make([]*FCTemplate, 0)
	for _, field := range eeC.Fields {
		switch field.GetPathSlice()[0] {
		case utils.MetaHdr:
			eeC.headerFields = append(eeC.headerFields, field)
		case utils.MetaExp, utils.MetaUCH:
			eeC.contentFields = append(eeC.contentFields, field)
		case utils.MetaTrl:
			eeC.trailerFields = append(eeC.trailerFields, field)
		}
	}
}

// HeaderFields returns the fields that have *hdr prefix
func (eeC *EventExporterCfg) HeaderFields() []*FCTemplate {
	return eeC.headerFields
}

// ContentFields returns the fields that do not have *hdr or *trl prefix
func (eeC *EventExporterCfg) ContentFields() []*FCTemplate {
	return eeC.contentFields
}

// TrailerFields returns the fields that have *trl prefix
func (eeC *EventExporterCfg) TrailerFields() []*FCTemplate {
	return eeC.trailerFields
}

// Clone returns a deep copy of EventExporterCfg
func (eeC EventExporterCfg) Clone() (cln *EventExporterCfg) {
	cln = &EventExporterCfg{
		ID:                 eeC.ID,
		Type:               eeC.Type,
		ExportPath:         eeC.ExportPath,
		Timezone:           eeC.Timezone,
		Flags:              eeC.Flags.Clone(),
		AttributeSCtx:      eeC.AttributeSCtx,
		Synchronous:        eeC.Synchronous,
		Attempts:           eeC.Attempts,
		ConcurrentRequests: eeC.ConcurrentRequests,
		Fields:             make([]*FCTemplate, len(eeC.Fields)),
		headerFields:       make([]*FCTemplate, len(eeC.headerFields)),
		contentFields:      make([]*FCTemplate, len(eeC.contentFields)),
		trailerFields:      make([]*FCTemplate, len(eeC.trailerFields)),
		Opts:               make(map[string]interface{}),
	}

	if eeC.Filters != nil {
		cln.Filters = make([]string, len(eeC.Filters))
		for idx, val := range eeC.Filters {
			cln.Filters[idx] = val
		}
	}
	if eeC.AttributeSIDs != nil {
		cln.AttributeSIDs = make([]string, len(eeC.AttributeSIDs))
		for idx, val := range eeC.AttributeSIDs {
			cln.AttributeSIDs[idx] = val
		}
	}

	for idx, fld := range eeC.Fields {
		cln.Fields[idx] = fld.Clone()
	}
	for idx, fld := range eeC.headerFields {
		cln.headerFields[idx] = fld.Clone()
	}
	for idx, fld := range eeC.contentFields {
		cln.contentFields[idx] = fld.Clone()
	}
	for idx, fld := range eeC.trailerFields {
		cln.trailerFields[idx] = fld.Clone()
	}
	for k, v := range eeC.Opts {
		cln.Opts[k] = v
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (eeC *EventExporterCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	flgs := eeC.Flags.SliceFlags()
	if flgs == nil {
		flgs = []string{}
	}
	initialMP = map[string]interface{}{
		utils.IDCfg:                 eeC.ID,
		utils.TypeCfg:               eeC.Type,
		utils.ExportPathCfg:         eeC.ExportPath,
		utils.TimezoneCfg:           eeC.Timezone,
		utils.FiltersCfg:            eeC.Filters,
		utils.FlagsCfg:              flgs,
		utils.AttributeContextCfg:   eeC.AttributeSCtx,
		utils.AttributeIDsCfg:       eeC.AttributeSIDs,
		utils.SynchronousCfg:        eeC.Synchronous,
		utils.AttemptsCfg:           eeC.Attempts,
		utils.ConcurrentRequestsCfg: eeC.ConcurrentRequests,
	}
	opts := make(map[string]interface{})
	for k, v := range eeC.Opts {
		opts[k] = v
	}
	initialMP[utils.OptsCfg] = opts

	if eeC.Fields != nil {
		fields := make([]map[string]interface{}, 0, len(eeC.Fields))
		for _, fld := range eeC.Fields {
			fields = append(fields, fld.AsMapInterface(separator))
		}
		initialMP[utils.FieldsCfg] = fields
	}
	return
}
