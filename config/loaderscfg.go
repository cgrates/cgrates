/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package config

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// LoaderSCfgs to export some methods for LoaderS profiles
type LoaderSCfgs []*LoaderSCfg

// loadLoaderSCfg loads the LoaderS section of the configuration
func (ldrs *LoaderSCfgs) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnLoaderCfg := make([]*LoaderJsonCfg, 0)
	if err = jsnCfg.GetSection(ctx, LoaderSJSON, &jsnLoaderCfg); err != nil {
		return
	}
	// cfg.loaderCfg = make(LoaderSCfgs, len(jsnLoaderCfg))
	for _, profile := range jsnLoaderCfg {
		var ldr *LoaderSCfg
		if profile.ID != nil {
			for _, loader := range cfg.loaderCfg {
				if loader.ID == *profile.ID {
					ldr = loader
					break
				}
			}
		}
		if ldr == nil {
			ldr = getDftLoaderCfg()
			ldr.Data = nil
			*ldrs = append(*ldrs, ldr) // use append so the loaderS profile to be loaded from multiple files
		}

		if err = ldr.loadFromJSONCfg(profile, cfg.templates); err != nil {
			return
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (ldrs LoaderSCfgs) AsMapInterface() any {
	mp := make([]map[string]any, len(ldrs))
	for i, item := range ldrs {
		mp[i] = item.AsMapInterface()
	}
	return mp
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

func (LoaderSCfgs) SName() string              { return LoaderSJSON }
func (ldrs LoaderSCfgs) CloneSection() Section { return ldrs.Clone() }

// Clone itself into a new LoaderSCfgs
func (ldrs LoaderSCfgs) Clone() *LoaderSCfgs {
	cln := make(LoaderSCfgs, len(ldrs))
	for i, ldr := range ldrs {
		cln[i] = ldr.Clone()
	}
	return &cln
}

type LoaderSOptsCfg struct {
	Cache       string
	WithIndex   bool
	ForceLock   bool
	StopOnError bool
}

// LoaderSCfg the config for a loader
type LoaderSCfg struct {
	ID             string
	Enabled        bool
	Tenant         string
	RunDelay       time.Duration
	LockFilePath   string
	CacheSConns    []string
	FieldSeparator string
	TpInDir        string
	TpOutDir       string
	Data           []*LoaderDataType

	Action string
	Opts   *LoaderSOptsCfg
	Cache  map[string]*CacheParamCfg
}

// LoaderDataType the template for profile loading
type LoaderDataType struct {
	ID       string
	Type     string
	Filename string
	Flags    utils.FlagsWithParams
	Fields   []*FCTemplate
}

func (l *LoaderSOptsCfg) loadFromJSONCfg(jsnCfg *LoaderJsonOptsCfg) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Cache != nil {
		l.Cache = *jsnCfg.Cache
	}
	if jsnCfg.WithIndex != nil {
		l.WithIndex = *jsnCfg.WithIndex
	}
	if jsnCfg.ForceLock != nil {
		l.ForceLock = *jsnCfg.ForceLock
	}
	if jsnCfg.StopOnError != nil {
		l.StopOnError = *jsnCfg.StopOnError
	}
}
func (lData *LoaderDataType) loadFromJSONCfg(jsnCfg *LoaderJsonDataType, msgTemplates map[string][]*FCTemplate) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Id != nil {
		lData.ID = *jsnCfg.Id
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
		if lData.Fields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnCfg.Fields); err != nil {
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

func (l *LoaderSCfg) loadFromJSONCfg(jsnCfg *LoaderJsonCfg, msgTemplates map[string][]*FCTemplate) (err error) {
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
		l.Tenant = *jsnCfg.Tenant
	}
	if jsnCfg.Run_delay != nil {
		if l.RunDelay, err = utils.ParseDurationWithNanosecs(*jsnCfg.Run_delay); err != nil {
			return
		}
	}
	if jsnCfg.Caches_conns != nil {
		l.CacheSConns = tagInternalConns(*jsnCfg.Caches_conns, utils.MetaCaches)
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
	if jsnCfg.Lockfile_path != nil {
		// Check if path is relative, case in which "tpIn" folder should be prepended
		l.LockFilePath = *jsnCfg.Lockfile_path
	}
	if jsnCfg.Data != nil {
		for _, jsnLoCfg := range *jsnCfg.Data {
			if jsnLoCfg == nil {
				continue
			}
			var ldrDataType *LoaderDataType
			var lType, id string
			if jsnLoCfg.Type != nil {
				lType = *jsnLoCfg.Type
			}
			if jsnLoCfg.Id != nil {
				id = *jsnLoCfg.Id
			}
			for _, ldrDT := range l.Data {
				if ldrDT.Type == lType && id == ldrDT.ID {
					ldrDataType = ldrDT
					break
				}
			}
			if ldrDataType == nil {
				ldrDataType = new(LoaderDataType)
				l.Data = append(l.Data, ldrDataType) // use append so the loaderS profile to be loaded from multiple files
			}
			if err := ldrDataType.loadFromJSONCfg(jsnLoCfg, msgTemplates); err != nil {
				return err
			}
		}
	}
	if jsnCfg.Action != nil {
		l.Action = *jsnCfg.Action
	}
	for kJsn, vJsn := range jsnCfg.Cache {
		val := new(CacheParamCfg)
		if err := val.loadFromJSONCfg(vJsn); err != nil {
			return err
		}
		l.Cache[kJsn] = val
	}
	l.Opts.loadFromJSONCfg(jsnCfg.Opts)
	return nil
}

func (l LoaderSCfg) GetLockFilePath() (pathL string) {
	pathL = l.LockFilePath
	if !filepath.IsAbs(pathL) {
		pathL = path.Join(l.TpInDir, pathL)
	}

	if file, err := os.Stat(pathL); err == nil && file.IsDir() {
		pathL = path.Join(pathL, l.ID+".lck")
	}
	return
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
	opts := *l.Opts
	cln = &LoaderSCfg{
		ID:             l.ID,
		Enabled:        l.Enabled,
		Tenant:         l.Tenant,
		RunDelay:       l.RunDelay,
		LockFilePath:   l.LockFilePath,
		CacheSConns:    slices.Clone(l.CacheSConns),
		FieldSeparator: l.FieldSeparator,
		TpInDir:        l.TpInDir,
		TpOutDir:       l.TpOutDir,
		Data:           make([]*LoaderDataType, len(l.Data)),
		Action:         l.Action,
		Opts:           &opts,
		Cache:          make(map[string]*CacheParamCfg),
	}
	for idx, fld := range l.Data {
		cln.Data[idx] = fld.Clone()
	}
	for key, value := range l.Cache {
		cln.Cache[key] = value.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (lData LoaderDataType) AsMapInterface() (initialMP map[string]any) {
	initialMP = map[string]any{
		utils.TypeCf:      lData.Type,
		utils.FilenameCfg: lData.Filename,
		utils.FlagsCfg:    lData.Flags.SliceFlags(),
	}

	fields := make([]map[string]any, len(lData.Fields))
	for i, item := range lData.Fields {
		fields[i] = item.AsMapInterface()
	}
	initialMP[utils.FieldsCfg] = fields
	return
}

// AsMapInterface returns the config as a map[string]any
func (l LoaderSCfg) AsMapInterface() (mp map[string]any) {
	mp = map[string]any{
		utils.IDCfg:           l.ID,
		utils.TenantCfg:       l.Tenant,
		utils.EnabledCfg:      l.Enabled,
		utils.LockFilePathCfg: l.LockFilePath,
		utils.FieldSepCfg:     l.FieldSeparator,
		utils.TpInDirCfg:      l.TpInDir,
		utils.TpOutDirCfg:     l.TpOutDir,
		utils.RunDelayCfg:     "0",
		utils.ActionCfg:       l.Action,
		utils.OptsCfg: map[string]any{
			utils.MetaCache:       l.Opts.Cache,
			utils.MetaWithIndex:   l.Opts.WithIndex,
			utils.MetaForceLock:   l.Opts.ForceLock,
			utils.MetaStopOnError: l.Opts.StopOnError,
		},
	}
	if l.Data != nil {
		data := make([]map[string]any, len(l.Data))
		for i, item := range l.Data {
			data[i] = item.AsMapInterface()
		}
		mp[utils.DataCfg] = data
	}
	if l.RunDelay != 0 {
		mp[utils.RunDelayCfg] = l.RunDelay.String()
	}
	if l.CacheSConns != nil {
		mp[utils.CachesConnsCfg] = stripInternalConns(l.CacheSConns)
	}
	if l.Cache != nil {
		cache := make(map[string]any, len(l.Cache))
		for key, value := range l.Cache {
			cache[key] = value.AsMapInterface()
		}
		mp[utils.CacheCfg] = cache
	}
	return
}

type LoaderJsonDataType struct {
	Id        *string
	Type      *string
	File_name *string
	Flags     *[]string
	Fields    *[]*FcTemplateJsonCfg
}

type LoaderJsonOptsCfg struct {
	Cache       *string `json:"*cache"`
	WithIndex   *bool   `json:"*withIndex"`
	ForceLock   *bool   `json:"*forceLock"`
	StopOnError *bool   `json:"*stopOnError"`
}

type LoaderJsonCfg struct {
	ID              *string
	Enabled         *bool
	Tenant          *string
	Run_delay       *string
	Lockfile_path   *string
	Caches_conns    *[]string
	Field_separator *string
	Tp_in_dir       *string
	Tp_out_dir      *string
	Data            *[]*LoaderJsonDataType

	Action *string
	Opts   *LoaderJsonOptsCfg
	Cache  map[string]*CacheParamJsonCfg
}

func equalsLoaderDatasType(v1, v2 []*LoaderDataType) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].ID != v2[i].ID ||
			v1[i].Type != v2[i].Type ||
			v1[i].Filename != v2[i].Filename ||
			!slices.Equal(v1[i].Flags.SliceFlags(), v2[i].Flags.SliceFlags()) ||
			!fcTemplatesEqual(v1[i].Fields, v2[i].Fields) {
			return false
		}
	}
	return true
}

func diffLoaderJsonOptsCfg(v1, v2 *LoaderSOptsCfg) (d *LoaderJsonOptsCfg) {
	d = new(LoaderJsonOptsCfg)
	if v1.Cache != v2.Cache {
		d.Cache = utils.StringPointer(v2.Cache)
	}
	if v1.WithIndex != v2.WithIndex {
		d.WithIndex = utils.BoolPointer(v2.WithIndex)
	}
	if v1.ForceLock != v2.ForceLock {
		d.ForceLock = utils.BoolPointer(v2.ForceLock)
	}
	if v1.StopOnError != v2.StopOnError {
		d.StopOnError = utils.BoolPointer(v2.StopOnError)
	}
	return
}
func diffLoaderJsonCfg(v1, v2 *LoaderSCfg) (d *LoaderJsonCfg) {
	d = new(LoaderJsonCfg)
	if v1.ID != v2.ID {
		d.ID = utils.StringPointer(v2.ID)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.Tenant != v2.Tenant {
		d.Tenant = utils.StringPointer(v2.Tenant)
	}
	if v1.RunDelay != v2.RunDelay {
		d.Run_delay = utils.StringPointer(v2.RunDelay.String())
	}
	if v1.LockFilePath != v2.LockFilePath {
		d.Lockfile_path = utils.StringPointer(v2.LockFilePath)
	}
	if !slices.Equal(v1.CacheSConns, v2.CacheSConns) {
		d.Caches_conns = utils.SliceStringPointer(stripInternalConns(v2.CacheSConns))
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
			req = diffFcTemplateJsonCfg(req, nil, val2.Fields)
			data[i] = &LoaderJsonDataType{
				Id:        utils.StringPointer(val2.ID),
				Type:      utils.StringPointer(val2.Type),
				File_name: utils.StringPointer(val2.Filename),
				Flags:     utils.SliceStringPointer(val2.Flags.SliceFlags()),
				Fields:    &req,
			}
		}
		d.Data = &data
	}
	if v1.Action != v2.Action {
		d.Action = utils.StringPointer(v2.Action)
	}
	d.Opts = diffLoaderJsonOptsCfg(v1.Opts, v2.Opts)
	d.Cache = diffCacheParamsJsonCfg(d.Cache, v2.Cache)
	return
}

func equalsLoadersJsonCfg(v1, v2 LoaderSCfgs) bool {
	if len(v1) != len(v2) {
		return false
	}
	for i := range v2 {
		if v1[i].ID != v2[i].ID ||
			v1[i].Enabled != v2[i].Enabled ||
			v1[i].Tenant != v2[i].Tenant ||
			v1[i].RunDelay != v2[i].RunDelay ||
			v1[i].LockFilePath != v2[i].LockFilePath ||
			!slices.Equal(v1[i].CacheSConns, v2[i].CacheSConns) ||
			v1[i].FieldSeparator != v2[i].FieldSeparator ||
			v1[i].TpInDir != v2[i].TpInDir ||
			v1[i].TpOutDir != v2[i].TpOutDir ||
			v1[i].Action != v2[i].Action ||
			!equalsLoaderDatasType(v1[i].Data, v2[i].Data) ||
			v1[i].Opts.Cache != v2[i].Opts.Cache ||
			v1[i].Opts.WithIndex != v2[i].Opts.WithIndex ||
			v1[i].Opts.ForceLock != v2[i].Opts.ForceLock ||
			v1[i].Opts.StopOnError != v2[i].Opts.StopOnError {
			return false
		}
	}
	return true
}
func diffLoadersJsonCfg(d []*LoaderJsonCfg, v1, v2 LoaderSCfgs) []*LoaderJsonCfg {
	if equalsLoadersJsonCfg(v1, v2) {
		return d
	}
	d = make([]*LoaderJsonCfg, len(v2))
	dft := getDftLoaderCfg()
	for i, val2 := range v2 {
		d[i] = diffLoaderJsonCfg(dft, val2)
	}
	return d
}
