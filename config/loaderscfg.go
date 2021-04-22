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

// NewDfltLoaderSCfg returns the first cached default value for a LoaderSCfg connection
func NewDfltLoaderSCfg() *LoaderSCfg {
	if dfltLoaderConfig == nil {
		return new(LoaderSCfg)
	}
	return dfltLoaderConfig.Clone()
}

// LoaderSCfgs to export some methods for LoaderS profiles
type LoaderSCfgs []*LoaderSCfg

// AsMapInterface returns the config as a map[string]interface{}
func (ldrs LoaderSCfgs) AsMapInterface(separator string) (loaderCfg []map[string]interface{}) {
	loaderCfg = make([]map[string]interface{}, len(ldrs))
	for i, item := range ldrs {
		loaderCfg[i] = item.AsMapInterface(separator)
	}
	return
}

// Enabled returns true if Loader Service is enabled
func (ldrs LoaderSCfgs) Enabled() bool {
	for _, ldr := range ldrs {
		if ldr.Enabled {
			return true
		}
	}
	return false
}

// Clone itself into a new LoaderSCfgs
func (ldrs LoaderSCfgs) Clone() (cln LoaderSCfgs) {
	cln = make(LoaderSCfgs, len(ldrs))
	for i, ldr := range ldrs {
		cln[i] = ldr.Clone()
	}
	return
}

// LoaderSCfg the config for a loader
type LoaderSCfg struct {
	ID             string
	Enabled        bool
	Tenant         RSRParsers
	DryRun         bool
	RunDelay       time.Duration
	LockFileName   string
	CacheSConns    []string
	FieldSeparator string
	TpInDir        string
	TpOutDir       string
	Data           []*LoaderDataType
}

// LoaderDataType the template for profile loading
type LoaderDataType struct {
	Type     string
	Filename string
	Flags    utils.FlagsWithParams
	Fields   []*FCTemplate
}

func (lData *LoaderDataType) loadFromJSONCfg(jsnCfg *LoaderJsonDataType, msgTemplates map[string][]*FCTemplate, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Type != nil {
		lData.Type = *jsnCfg.Type
	}
	if jsnCfg.File_name != nil {
		lData.Filename = *jsnCfg.File_name
	}
	if jsnCfg.Flags != nil {
		lData.Flags = utils.FlagsWithParamsFromSlice(*jsnCfg.Flags)
	}
	if jsnCfg.Fields != nil {
		if lData.Fields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Fields, separator); err != nil {
			return
		}
		if tpls, err := InflateTemplates(lData.Fields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			lData.Fields = tpls
		}
	}
	return nil
}

func (l *LoaderSCfg) loadFromJSONCfg(jsnCfg *LoaderJsonCfg, msgTemplates map[string][]*FCTemplate, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ID != nil {
		l.ID = *jsnCfg.ID
	}
	if jsnCfg.Enabled != nil {
		l.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Tenant != nil {
		if l.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Dry_run != nil {
		l.DryRun = *jsnCfg.Dry_run
	}
	if jsnCfg.Run_delay != nil {
		if l.RunDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.Run_delay); err != nil {
			return
		}
	}
	if jsnCfg.Lock_filename != nil {
		l.LockFileName = *jsnCfg.Lock_filename
	}
	if jsnCfg.Caches_conns != nil {
		l.CacheSConns = make([]string, len(*jsnCfg.Caches_conns))
		for idx, connID := range *jsnCfg.Caches_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				l.CacheSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			} else {
				l.CacheSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Field_separator != nil {
		l.FieldSeparator = *jsnCfg.Field_separator
	}
	if jsnCfg.Tp_in_dir != nil {
		l.TpInDir = *jsnCfg.Tp_in_dir
	}
	if jsnCfg.Tp_out_dir != nil {
		l.TpOutDir = *jsnCfg.Tp_out_dir
	}
	if jsnCfg.Data != nil {
		data := make([]*LoaderDataType, len(*jsnCfg.Data))
		for idx, jsnLoCfg := range *jsnCfg.Data {
			data[idx] = new(LoaderDataType)
			if err := data[idx].loadFromJSONCfg(jsnLoCfg, msgTemplates, separator); err != nil {
				return err
			}
		}
		l.Data = data
	}

	return nil
}

// Clone itself into a new LoaderDataType
func (lData LoaderDataType) Clone() (cln *LoaderDataType) {
	cln = &LoaderDataType{
		Type:     lData.Type,
		Filename: lData.Filename,
		Flags:    lData.Flags.Clone(),
		Fields:   make([]*FCTemplate, len(lData.Fields)),
	}
	for idx, val := range lData.Fields {
		cln.Fields[idx] = val.Clone()
	}
	return
}

// Clone itself into a new LoadersConfig
func (l LoaderSCfg) Clone() (cln *LoaderSCfg) {
	cln = &LoaderSCfg{
		ID:             l.ID,
		Enabled:        l.Enabled,
		Tenant:         l.Tenant,
		DryRun:         l.DryRun,
		RunDelay:       l.RunDelay,
		LockFileName:   l.LockFileName,
		CacheSConns:    utils.CloneStringSlice(l.CacheSConns),
		FieldSeparator: l.FieldSeparator,
		TpInDir:        l.TpInDir,
		TpOutDir:       l.TpOutDir,
		Data:           make([]*LoaderDataType, len(l.Data)),
	}
	for idx, connID := range l.CacheSConns {
		cln.CacheSConns[idx] = connID
	}
	for idx, fld := range l.Data {
		cln.Data[idx] = fld.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (lData *LoaderDataType) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.TypeCf:      lData.Type,
		utils.FilenameCfg: lData.Filename,
		utils.FlagsCfg:    lData.Flags.SliceFlags(),
	}

	fields := make([]map[string]interface{}, len(lData.Fields))
	for i, item := range lData.Fields {
		fields[i] = item.AsMapInterface(separator)
	}
	initialMP[utils.FieldsCfg] = fields
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (l *LoaderSCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.IDCfg:           l.ID,
		utils.TenantCfg:       l.Tenant.GetRule(separator),
		utils.EnabledCfg:      l.Enabled,
		utils.DryRunCfg:       l.DryRun,
		utils.LockFileNameCfg: l.LockFileName,
		utils.FieldSepCfg:     l.FieldSeparator,
		utils.TpInDirCfg:      l.TpInDir,
		utils.TpOutDirCfg:     l.TpOutDir,
		utils.RunDelayCfg:     "0",
	}
	if l.Data != nil {
		data := make([]map[string]interface{}, len(l.Data))
		for i, item := range l.Data {
			data[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.DataCfg] = data
	}
	if l.RunDelay != 0 {
		initialMP[utils.RunDelayCfg] = l.RunDelay.String()
	}
	if l.CacheSConns != nil {
		cacheSConns := make([]string, len(l.CacheSConns))
		for i, item := range l.CacheSConns {
			cacheSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches) {
				cacheSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.CachesConnsCfg] = cacheSConns
	}
	return
}
