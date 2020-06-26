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

	"github.com/cgrates/cgrates/utils"
)

type EEsCfg struct {
	Enabled         bool
	AttributeSConns []string
	Cache           map[string]*CacheParamCfg
	Exporters       []*EventExporterCfg
}

func (eeS *EEsCfg) loadFromJsonCfg(jsnCfg *EEsJsonCfg, sep string, dfltExpCfg *EventExporterCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		eeS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cache != nil {
		for kJsn, vJsn := range *jsnCfg.Cache {
			val := new(CacheParamCfg)
			if err := val.loadFromJsonCfg(vJsn); err != nil {
				return err
			}
			eeS.Cache[kJsn] = val
		}
	}
	if jsnCfg.Attributes_conns != nil {
		eeS.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for i, fID := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if fID == utils.MetaInternal {
				eeS.AttributeSConns[i] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			} else {
				eeS.AttributeSConns[i] = fID
			}
		}
	}
	return eeS.appendEEsExporters(jsnCfg.Exporters, sep, dfltExpCfg)
}

func (eeS *EEsCfg) appendEEsExporters(exporters *[]*EventExporterJsonCfg, separator string, dfltExpCfg *EventExporterCfg) (err error) {
	if exporters == nil {
		return
	}
	for _, jsnExp := range *exporters {
		exp := new(EventExporterCfg)
		if dfltExpCfg != nil {
			exp = dfltExpCfg.Clone()
		}
		var haveID bool
		if jsnExp.Id != nil {
			for _, exporter := range eeS.Exporters {
				if exporter.ID == *jsnExp.Id {
					exp = exporter
					haveID = true
					break
				}
			}
		}

		if err := exp.loadFromJsonCfg(jsnExp, separator); err != nil {
			return err
		}
		if !haveID {
			eeS.Exporters = append(eeS.Exporters, exp)
		}

	}
	return
}

// Clone itself into a new EEsCfg
func (eeS *EEsCfg) Clone() (cln *EEsCfg) {
	cln = new(EEsCfg)
	cln.Enabled = eeS.Enabled
	cln.AttributeSConns = make([]string, len(eeS.AttributeSConns))
	for idx, sConn := range eeS.AttributeSConns {
		cln.AttributeSConns[idx] = sConn
	}
	cln.Exporters = make([]*EventExporterCfg, len(eeS.Exporters))
	for idx, exp := range eeS.Exporters {
		cln.Exporters[idx] = exp.Clone()
	}
	return
}

func (eeS *EEsCfg) AsMapInterface(separator string) map[string]interface{} {
	exporters := make([]map[string]interface{}, len(eeS.Exporters))
	for i, item := range eeS.Exporters {
		exporters[i] = item.AsMapInterface(separator)
	}
	return map[string]interface{}{
		utils.EnabledCfg:         eeS.Enabled,
		utils.AttributeSConnsCfg: eeS.AttributeSConns,
		utils.ExportersCfg:       exporters,
	}
}

type EventExporterCfg struct {
	ID            string
	Type          string
	ExportPath    string
	Tenant        RSRParsers
	Timezone      string
	Filters       []string
	Flags         utils.FlagsWithParams
	AttributeSIDs []string // selective AttributeS profiles
	AttributeSCtx string   // context to use when querying AttributeS
	Synchronous   bool
	Attempts      int
	FieldSep      string
	Fields        []*FCTemplate
	headerFields  []*FCTemplate
	contentFields []*FCTemplate
	trailerFields []*FCTemplate
}

func (eeC *EventExporterCfg) loadFromJsonCfg(jsnEec *EventExporterJsonCfg, separator string) (err error) {
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
	if jsnEec.Tenant != nil {
		if eeC.Tenant, err = NewRSRParsers(*jsnEec.Tenant, true, separator); err != nil {
			return err
		}
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
		if eeC.Flags, err = utils.FlagsWithParamsFromSlice(*jsnEec.Flags); err != nil {
			return
		}
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
	if jsnEec.Field_separator != nil {
		eeC.FieldSep = *jsnEec.Field_separator
	}
	if jsnEec.Fields != nil {
		eeC.headerFields = make([]*FCTemplate, 0)
		eeC.contentFields = make([]*FCTemplate, 0)
		eeC.trailerFields = make([]*FCTemplate, 0)
		eeC.Fields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnEec.Fields, separator)
		if err != nil {
			return
		}
		for _, field := range eeC.Fields {
			switch field.GetPathSlice()[0] {
			case utils.MetaHdr:
				eeC.headerFields = append(eeC.headerFields, field)
			case utils.MetaExp:
				eeC.contentFields = append(eeC.contentFields, field)
			case utils.MetaTrl:
				eeC.trailerFields = append(eeC.trailerFields, field)
			}
			if strings.HasPrefix(field.GetPathSlice()[0], utils.MetaCache) { // special cache when loading fields that contains *cache in path
				eeC.contentFields = append(eeC.contentFields, field)
			}
		}
	}
	return
}

func (eeC *EventExporterCfg) HeaderFields() []*FCTemplate {
	return eeC.headerFields
}

func (eeC *EventExporterCfg) ContentFields() []*FCTemplate {
	return eeC.contentFields
}

func (eeC *EventExporterCfg) TrailerFields() []*FCTemplate {
	return eeC.trailerFields
}

func (eeC *EventExporterCfg) Clone() (cln *EventExporterCfg) {
	cln = new(EventExporterCfg)
	cln.ID = eeC.ID
	cln.Type = eeC.Type
	cln.ExportPath = eeC.ExportPath
	if len(eeC.Tenant) != 0 {
		cln.Tenant = make(RSRParsers, len(eeC.Tenant))
		for idx, val := range eeC.Tenant {
			clnVal := *val
			cln.Tenant[idx] = &clnVal
		}
	}
	cln.Timezone = eeC.Timezone
	if len(eeC.Filters) != 0 {
		cln.Filters = make([]string, len(eeC.Filters))
		for idx, val := range eeC.Filters {
			cln.Filters[idx] = val
		}
	}
	cln.Flags = eeC.Flags
	cln.AttributeSCtx = eeC.AttributeSCtx
	if len(eeC.AttributeSIDs) != 0 {
		cln.AttributeSIDs = make([]string, len(eeC.AttributeSIDs))
		for idx, val := range eeC.AttributeSIDs {
			cln.AttributeSIDs[idx] = val
		}
	}
	cln.Synchronous = eeC.Synchronous
	cln.Attempts = eeC.Attempts
	cln.FieldSep = eeC.FieldSep

	cln.Fields = make([]*FCTemplate, len(eeC.Fields))
	for idx, fld := range eeC.Fields {
		cln.Fields[idx] = fld.Clone()
	}
	cln.headerFields = make([]*FCTemplate, len(eeC.headerFields))
	for idx, fld := range eeC.headerFields {
		cln.headerFields[idx] = fld.Clone()
	}
	cln.contentFields = make([]*FCTemplate, len(eeC.contentFields))
	for idx, fld := range eeC.contentFields {
		cln.contentFields[idx] = fld.Clone()
	}
	cln.trailerFields = make([]*FCTemplate, len(eeC.trailerFields))
	for idx, fld := range eeC.trailerFields {
		cln.trailerFields[idx] = fld.Clone()
	}
	return
}

func (eeC *EventExporterCfg) AsMapInterface(separator string) map[string]interface{} {
	var tenant string
	if eeC.Tenant != nil {
		values := make([]string, len(eeC.Tenant))
		for i, item := range eeC.Tenant {
			values[i] = item.Rules
		}
		tenant = strings.Join(values, separator)
	}

	flags := make(map[string][]interface{}, len(eeC.Flags))
	for key, val := range eeC.Flags {
		buf := make([]interface{}, len(val))
		for i, item := range val {
			buf[i] = item
		}
		flags[key] = buf
	}
	fields := make([]map[string]interface{}, 0, len(eeC.Fields))
	for _, fld := range eeC.Fields {
		fields = append(fields, fld.AsMapInterface(separator))
	}

	return map[string]interface{}{
		utils.IDCfg:               eeC.ID,
		utils.TypeCfg:             eeC.Type,
		utils.ExportPath:          eeC.ExportPath,
		utils.ExportPathCfg:       eeC.FieldSep,
		utils.TenantCfg:           tenant,
		utils.TimezoneCfg:         eeC.Timezone,
		utils.FiltersCfg:          eeC.Filters,
		utils.FlagsCfg:            flags,
		utils.AttributeContextCfg: eeC.AttributeSCtx,
		utils.AttributeIDsCfg:     eeC.AttributeSIDs,
		utils.SynchronousCfg:      eeC.Synchronous,
		utils.AttemptsCfg:         eeC.Attempts,
		utils.FieldSeparatorCfg:   eeC.FieldSep,
		utils.FieldsCfg:           fields,
	}
}
