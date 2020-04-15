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

import "github.com/cgrates/cgrates/utils"

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

func (cdre *CdreCfg) AsMapInterface() map[string]interface{} {
	fields := make([]map[string]interface{}, len(cdre.Fields))
	for i, item := range cdre.Fields {
		fields[i] = item.AsMapInterface()
	}

	return map[string]interface{}{
		utils.ExportFormatCfg:      cdre.ExportFormat,
		utils.ExportPathCfg:        cdre.ExportPath,
		utils.FiltersCfg:           cdre.Filters,
		utils.TenantCfg:            cdre.Tenant,
		utils.AttributeSContextCfg: cdre.AttributeSContext,
		utils.SynchronousCfg:       cdre.Synchronous,
		utils.AttemptsCfg:          cdre.Attempts,
		utils.FieldSeparatorCfg:    cdre.FieldSeparator,
		utils.FieldsCfg:            fields,
	}
}
