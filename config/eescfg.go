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
	for kJsn, vJsn := range jsnCfg.Cache {
		val := new(CacheParamCfg)
		if err := val.loadFromJSONCfg(vJsn); err != nil {
			return err
		}
		eeS.Cache[kJsn] = val
	}
	if jsnCfg.Attributes_conns != nil {
		eeS.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
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
		Enabled:   eeS.Enabled,
		Cache:     make(map[string]*CacheParamCfg),
		Exporters: make([]*EventExporterCfg, len(eeS.Exporters)),
	}
	if eeS.AttributeSConns != nil {
		cln.AttributeSConns = utils.CloneStringSlice(eeS.AttributeSConns)
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
		initialMP[utils.AttributeSConnsCfg] = getInternalJSONConns(eeS.AttributeSConns)
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
	ID            string
	Type          string
	ExportPath    string
	Opts          map[string]interface{}
	Tenant        RSRParsers
	Timezone      string
	Filters       []string
	Flags         utils.FlagsWithParams
	AttributeSIDs []string // selective AttributeS profiles
	AttributeSCtx string   // context to use when querying AttributeS
	Synchronous   bool
	Attempts      int
	Fields        []*FCTemplate
	headerFields  []*FCTemplate
	contentFields []*FCTemplate
	trailerFields []*FCTemplate
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
	if jsnEec.Tenant != nil {
		if eeC.Tenant, err = NewRSRParsers(*jsnEec.Tenant, separator); err != nil {
			return err
		}
	}
	if jsnEec.Timezone != nil {
		eeC.Timezone = *jsnEec.Timezone
	}
	if jsnEec.Filters != nil {
		eeC.Filters = utils.CloneStringSlice(*jsnEec.Filters)
	}
	if jsnEec.Flags != nil {
		eeC.Flags = utils.FlagsWithParamsFromSlice(*jsnEec.Flags)
	}
	if jsnEec.Attribute_context != nil {
		eeC.AttributeSCtx = *jsnEec.Attribute_context
	}
	if jsnEec.Attribute_ids != nil {
		eeC.AttributeSIDs = utils.CloneStringSlice(*jsnEec.Attribute_ids)
	}
	if jsnEec.Synchronous != nil {
		eeC.Synchronous = *jsnEec.Synchronous
	}
	if jsnEec.Attempts != nil {
		eeC.Attempts = *jsnEec.Attempts
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
		ID:            eeC.ID,
		Type:          eeC.Type,
		ExportPath:    eeC.ExportPath,
		Tenant:        eeC.Tenant.Clone(),
		Timezone:      eeC.Timezone,
		Flags:         eeC.Flags.Clone(),
		AttributeSCtx: eeC.AttributeSCtx,
		Synchronous:   eeC.Synchronous,
		Attempts:      eeC.Attempts,
		Fields:        make([]*FCTemplate, len(eeC.Fields)),
		headerFields:  make([]*FCTemplate, len(eeC.headerFields)),
		contentFields: make([]*FCTemplate, len(eeC.contentFields)),
		trailerFields: make([]*FCTemplate, len(eeC.trailerFields)),
		Opts:          make(map[string]interface{}),
	}

	if eeC.Filters != nil {
		cln.Filters = utils.CloneStringSlice(eeC.Filters)
	}
	if eeC.AttributeSIDs != nil {
		cln.AttributeSIDs = utils.CloneStringSlice(eeC.AttributeSIDs)
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
		utils.IDCfg:               eeC.ID,
		utils.TypeCfg:             eeC.Type,
		utils.ExportPathCfg:       eeC.ExportPath,
		utils.TenantCfg:           eeC.Tenant.GetRule(separator),
		utils.TimezoneCfg:         eeC.Timezone,
		utils.FiltersCfg:          eeC.Filters,
		utils.FlagsCfg:            flgs,
		utils.AttributeContextCfg: eeC.AttributeSCtx,
		utils.AttributeIDsCfg:     eeC.AttributeSIDs,
		utils.SynchronousCfg:      eeC.Synchronous,
		utils.AttemptsCfg:         eeC.Attempts,
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

// EventExporterJsonCfg is the configuration of a single EventExporter
type EventExporterJsonCfg struct {
	Id                *string
	Type              *string
	Export_path       *string
	Opts              map[string]interface{}
	Tenant            *string
	Timezone          *string
	Filters           *[]string
	Flags             *[]string
	Attribute_ids     *[]string
	Attribute_context *string
	Synchronous       *bool
	Attempts          *int
	Fields            *[]*FcTemplateJsonCfg
}

func diffEventExporterJsonCfg(d *EventExporterJsonCfg, v1, v2 *EventExporterCfg, separator string) *EventExporterJsonCfg {
	if d == nil {
		d = new(EventExporterJsonCfg)
	}
	if v1.ID != v2.ID {
		d.Id = utils.StringPointer(v2.ID)
	}
	if v1.Type != v2.Type {
		d.Type = utils.StringPointer(v2.Type)
	}
	if v1.ExportPath != v2.ExportPath {
		d.Export_path = utils.StringPointer(v2.ExportPath)
	}
	d.Opts = diffMap(d.Opts, v1.Opts, v2.Opts)
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
	if !utils.SliceStringEqual(v1.AttributeSIDs, v2.AttributeSIDs) {
		d.Attribute_ids = &v2.AttributeSIDs
	}
	if v1.AttributeSCtx != v2.AttributeSCtx {
		d.Attribute_context = utils.StringPointer(v2.AttributeSCtx)
	}
	if v1.Synchronous != v2.Synchronous {
		d.Synchronous = utils.BoolPointer(v2.Synchronous)
	}
	if v1.Attempts != v2.Attempts {
		d.Attempts = utils.IntPointer(v2.Attempts)
	}
	var flds []*FcTemplateJsonCfg
	if d.Fields != nil {
		flds = *d.Fields
	}
	flds = diffFcTemplateJsonCfg(flds, v1.Fields, v2.Fields, separator)
	if flds != nil {
		d.Fields = &flds
	}
	return d
}

func getEventExporterJsonCfg(d []*EventExporterJsonCfg, id string) (*EventExporterJsonCfg, int) {
	for i, v := range d {
		if v.Id != nil && *v.Id == id {
			return v, i
		}
	}
	return nil, -1
}

func getEventExporterCfg(d []*EventExporterCfg, id string) *EventExporterCfg {
	for _, v := range d {
		if v.ID == id {
			return v
		}
	}
	return new(EventExporterCfg)
}

func diffEventExportersJsonCfg(d *[]*EventExporterJsonCfg, v1, v2 []*EventExporterCfg, separator string) *[]*EventExporterJsonCfg {
	if d == nil || *d == nil {
		d = &[]*EventExporterJsonCfg{}
	}
	for _, val := range v2 {
		dv, i := getEventExporterJsonCfg(*d, val.ID)
		dv = diffEventExporterJsonCfg(dv, getEventExporterCfg(v1, val.ID), val, separator)
		if i == -1 {
			*d = append(*d, dv)
		} else {
			(*d)[i] = dv
		}
	}

	return d
}

// EEsJsonCfg contains the configuration of EventExporterService
type EEsJsonCfg struct {
	Enabled          *bool
	Attributes_conns *[]string
	Cache            map[string]*CacheParamJsonCfg
	Exporters        *[]*EventExporterJsonCfg
}

func diffEEsJsonCfg(d *EEsJsonCfg, v1, v2 *EEsCfg, separator string) *EEsJsonCfg {
	if d == nil {
		d = new(EEsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	d.Cache = diffCacheParamsJsonCfg(d.Cache, v2.Cache)
	d.Exporters = diffEventExportersJsonCfg(d.Exporters, v1.Exporters, v2.Exporters, separator)
	return d
}
