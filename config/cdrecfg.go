// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"github.com/cgrates/cgrates/utils"
)

// One instance of CdrExporter
type CdreCfg struct {
	ExportFormat      string
	ExportPath        string
	Filters           []string
	Tenant            string
	AttributeSContext string
	Synchronous       bool
	Attempts          int
	FieldSeparator    rune
	Fields            []*FCTemplate
}

func (self *CdreCfg) loadFromJsonCfg(jsnCfg *CdreJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Export_format != nil {
		self.ExportFormat = *jsnCfg.Export_format
	}
	if jsnCfg.Export_path != nil {
		self.ExportPath = *jsnCfg.Export_path
	}
	if jsnCfg.Filters != nil {
		self.Filters = make([]string, len(*jsnCfg.Filters))
		for i, fltr := range *jsnCfg.Filters {
			self.Filters[i] = fltr
		}
	}
	if jsnCfg.Tenant != nil {
		self.Tenant = *jsnCfg.Tenant
	}
	if jsnCfg.Synchronous != nil {
		self.Synchronous = *jsnCfg.Synchronous
	}
	if jsnCfg.Attempts != nil {
		self.Attempts = *jsnCfg.Attempts
	}
	if jsnCfg.Attributes_context != nil {
		self.AttributeSContext = *jsnCfg.Attributes_context
	}
	if jsnCfg.Field_separator != nil && len(*jsnCfg.Field_separator) > 0 { // Make sure we got at least one character so we don't get panic here
		sepStr := *jsnCfg.Field_separator
		self.FieldSeparator = rune(sepStr[0])
	}
	if jsnCfg.Fields != nil {
		if self.Fields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Fields, separator); err != nil {
			return err
		}
	}
	return nil
}

// Clone itself into a new CdreCfg
func (self *CdreCfg) Clone() *CdreCfg {
	clnCdre := new(CdreCfg)
	clnCdre.ExportFormat = self.ExportFormat
	clnCdre.ExportPath = self.ExportPath
	clnCdre.Synchronous = self.Synchronous
	clnCdre.Attempts = self.Attempts
	clnCdre.FieldSeparator = self.FieldSeparator
	clnCdre.Tenant = self.Tenant
	clnCdre.Filters = make([]string, len(self.Filters))
	for i, fltr := range self.Filters {
		clnCdre.Filters[i] = fltr
	}
	clnCdre.Fields = make([]*FCTemplate, len(self.Fields))
	for idx, fld := range self.Fields {
		clnCdre.Fields[idx] = fld.Clone()
	}
	return clnCdre
}

func (cdre *CdreCfg) AsMapInterface(separator string) map[string]any {
	fields := make([]map[string]any, len(cdre.Fields))
	for i, item := range cdre.Fields {
		fields[i] = item.AsMapInterface(separator)
	}

	return map[string]any{
		utils.ExportFormatCfg:      cdre.ExportFormat,
		utils.ExportPathCfg:        cdre.ExportPath,
		utils.FiltersCfg:           cdre.Filters,
		utils.TenantCfg:            cdre.Tenant,
		utils.AttributeSContextCfg: cdre.AttributeSContext,
		utils.SynchronousCfg:       cdre.Synchronous,
		utils.AttemptsCfg:          cdre.Attempts,
		utils.FieldSeparatorCfg:    string(cdre.FieldSeparator),
		utils.FieldsCfg:            fields,
	}
}
