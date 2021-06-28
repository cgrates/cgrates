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
		l.CacheSConns = updateInternalConns(*jsnCfg.Caches_conns, utils.MetaCaches)
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
		for _, jsnLoCfg := range *jsnCfg.Data {
			var ldrDataType *LoaderDataType
			if jsnLoCfg.Type != nil {
				for _, ldrDT := range l.Data {
					if ldrDT.Type == *jsnLoCfg.Type {
						ldrDataType = ldrDT
						break
					}
				}
			}
			if ldrDataType == nil {
				ldrDataType = new(LoaderDataType)
				l.Data = append(l.Data, ldrDataType) // use append so the loaderS profile to be loaded from multiple files
			}
			if err := ldrDataType.loadFromJSONCfg(jsnLoCfg, msgTemplates, separator); err != nil {
				return err
			}
		}
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
		initialMP[utils.CachesConnsCfg] = getInternalJSONConns(l.CacheSConns)
	}
	return
}

type LoaderJsonDataType struct {
	Type      *string
	File_name *string
	Flags     *[]string
	Fields    *[]*FcTemplateJsonCfg
}

type LoaderJsonCfg struct {
	ID              *string
	Enabled         *bool
	Tenant          *string
	Dry_run         *bool
	Run_delay       *string
	Lock_filename   *string
	Caches_conns    *[]string
	Field_separator *string
	Tp_in_dir       *string
	Tp_out_dir      *string
	Data            *[]*LoaderJsonDataType
}

func equalsLoaderDatasType(v1, v2 []*LoaderDataType) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].Type != v2[i].Type ||
			v1[i].Filename != v2[i].Filename ||
			!utils.SliceStringEqual(v1[i].Flags.SliceFlags(), v2[i].Flags.SliceFlags()) ||
			!fcTemplatesEqual(v1[i].Fields, v2[i].Fields) {
			return false
		}
	}
	return true
}

func diffLoaderJsonCfg(v1, v2 *LoaderSCfg, separator string) (d *LoaderJsonCfg) {
	d = new(LoaderJsonCfg)
	if v1.ID != v2.ID {
		d.ID = utils.StringPointer(v2.ID)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	tnt1 := v1.Tenant.GetRule(separator)
	tnt2 := v2.Tenant.GetRule(separator)
	if tnt1 != tnt2 {
		d.Tenant = utils.StringPointer(tnt2)
	}
	if v1.DryRun != v2.DryRun {
		d.Dry_run = utils.BoolPointer(v2.DryRun)
	}
	if v1.RunDelay != v2.RunDelay {
		d.Run_delay = utils.StringPointer(v2.RunDelay.String())
	}
	if v1.LockFileName != v2.LockFileName {
		d.Lock_filename = utils.StringPointer(v2.LockFileName)
	}
	if !utils.SliceStringEqual(v1.CacheSConns, v2.CacheSConns) {
		d.Caches_conns = utils.SliceStringPointer(getInternalJSONConns(v2.CacheSConns))
	}
	if v1.FieldSeparator != v2.FieldSeparator {
		d.Field_separator = utils.StringPointer(v2.FieldSeparator)
	}
	if v1.TpInDir != v2.TpInDir {
		d.Tp_in_dir = utils.StringPointer(v2.TpInDir)
	}
	if v1.TpOutDir != v2.TpOutDir {
		d.Tp_out_dir = utils.StringPointer(v2.TpOutDir)
	}
	if !equalsLoaderDatasType(v1.Data, v2.Data) {
		data := make([]*LoaderJsonDataType, len(v2.Data))
		for i, val2 := range v2.Data {
			var req []*FcTemplateJsonCfg
			req = diffFcTemplateJsonCfg(req, nil, val2.Fields, separator)
			data[i] = &LoaderJsonDataType{
				Type:      utils.StringPointer(val2.Type),
				File_name: utils.StringPointer(val2.Filename),
				Flags:     utils.SliceStringPointer(val2.Flags.SliceFlags()),
				Fields:    &req,
			}
		}
		d.Data = &data
	}
	return
}

func equalsLoadersJsonCfg(v1, v2 LoaderSCfgs) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].ID != v2[i].ID ||
			v1[i].Enabled != v2[i].Enabled ||
			!utils.SliceStringEqual(v1[i].Tenant.AsStringSlice(), v2[i].Tenant.AsStringSlice()) ||
			v1[i].DryRun != v2[i].DryRun ||
			v1[i].RunDelay != v2[i].RunDelay ||
			v1[i].LockFileName != v2[i].LockFileName ||
			!utils.SliceStringEqual(v1[i].CacheSConns, v2[i].CacheSConns) ||
			v1[i].FieldSeparator != v2[i].FieldSeparator ||
			v1[i].TpInDir != v2[i].TpInDir ||
			v1[i].TpOutDir != v2[i].TpOutDir ||
			!equalsLoaderDatasType(v1[i].Data, v2[i].Data) {
			return false
		}
	}
	return true
}
func diffLoadersJsonCfg(d []*LoaderJsonCfg, v1, v2 LoaderSCfgs, separator string) []*LoaderJsonCfg {
	if equalsLoadersJsonCfg(v1, v2) {
		return d
	}
	d = make([]*LoaderJsonCfg, len(v2))
	dft := NewDfltLoaderSCfg()
	for i, val2 := range v2 {
		d[i] = diffLoaderJsonCfg(dft, val2, separator)
	}
	return d
}
