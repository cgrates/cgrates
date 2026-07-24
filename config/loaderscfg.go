// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func NewDfltLoaderSCfg() *LoaderSCfg {
	if dfltLoaderConfig == nil {
		return new(LoaderSCfg)
	}
	dfltVal := *dfltLoaderConfig
	return &dfltVal
}

// LoaderSCfgs to export some methods for LoaderS profiles
type LoaderSCfgs []*LoaderSCfg

// Enabled returns true if Loader Service is enabled
func (ldrs LoaderSCfgs) Enabled() bool {
	for _, ldr := range ldrs {
		if ldr.Enabled {
			return true
		}
	}
	return false
}

type LoaderSCfg struct {
	Id             string
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

func NewDfltLoaderDataTypeConfig() *LoaderDataType {
	if dfltLoaderDataTypeConfig == nil {
		return new(LoaderDataType) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltLoaderDataTypeConfig // Copy the value instead of it's pointer
	return &dfltVal
}

type LoaderDataType struct { //rename to LoaderDataType
	Type     string
	Filename string
	Fields   []*FCTemplate
}

func (self *LoaderDataType) loadFromJsonCfg(jsnCfg *LoaderJsonDataType, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.Type != nil {
		self.Type = *jsnCfg.Type
	}
	if jsnCfg.File_name != nil {
		self.Filename = *jsnCfg.File_name
	}
	if jsnCfg.Fields != nil {
		if self.Fields, err = FCTemplatesFromFCTemplatesJsonCfg(*jsnCfg.Fields, separator); err != nil {
			return
		}
	}
	return nil
}

func (self *LoaderSCfg) loadFromJsonCfg(jsnCfg *LoaderJsonCfg, separator string) (err error) {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ID != nil {
		self.Id = *jsnCfg.ID
	}
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Tenant != nil {
		if self.Tenant, err = NewRSRParsers(*jsnCfg.Tenant, true, separator); err != nil {
			return err
		}
	}
	if jsnCfg.Dry_run != nil {
		self.DryRun = *jsnCfg.Dry_run
	}
	if jsnCfg.Run_delay != nil {
		self.RunDelay = time.Duration(*jsnCfg.Run_delay) * time.Second
	}
	if jsnCfg.Lock_filename != nil {
		self.LockFileName = *jsnCfg.Lock_filename
	}
	if jsnCfg.Caches_conns != nil {
		self.CacheSConns = make([]string, len(*jsnCfg.Caches_conns))
		for idx, connID := range *jsnCfg.Caches_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			if connID == utils.MetaInternal {
				self.CacheSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
			} else {
				self.CacheSConns[idx] = connID
			}
		}
	}
	if jsnCfg.Field_separator != nil {
		self.FieldSeparator = *jsnCfg.Field_separator
	}
	if jsnCfg.Tp_in_dir != nil {
		self.TpInDir = *jsnCfg.Tp_in_dir
	}
	if jsnCfg.Tp_out_dir != nil {
		self.TpOutDir = *jsnCfg.Tp_out_dir
	}
	if jsnCfg.Data != nil {
		data := make([]*LoaderDataType, len(*jsnCfg.Data))
		for idx, jsnLoCfg := range *jsnCfg.Data {
			data[idx] = NewDfltLoaderDataTypeConfig()
			data[idx].loadFromJsonCfg(jsnLoCfg, separator)
		}
		self.Data = data
	}

	return nil
}

// Clone itself into a new LoaderDataType
func (self *LoaderDataType) Clone() *LoaderDataType {
	cln := new(LoaderDataType)
	cln.Type = self.Type
	cln.Filename = self.Filename
	cln.Fields = make([]*FCTemplate, len(self.Fields))
	for idx, val := range self.Fields {
		cln.Fields[idx] = val.Clone()
	}
	return cln
}

// Clone itself into a new LoadersConfig
func (self *LoaderSCfg) Clone() *LoaderSCfg {
	clnLoader := new(LoaderSCfg)
	clnLoader.Id = self.Id
	clnLoader.Enabled = self.Enabled
	clnLoader.Tenant = self.Tenant
	clnLoader.DryRun = self.DryRun
	clnLoader.RunDelay = self.RunDelay
	clnLoader.LockFileName = self.LockFileName
	clnLoader.CacheSConns = make([]string, len(self.CacheSConns))
	for idx, connID := range self.CacheSConns {
		clnLoader.CacheSConns[idx] = connID
	}
	clnLoader.FieldSeparator = self.FieldSeparator
	clnLoader.TpInDir = self.TpInDir
	clnLoader.TpOutDir = self.TpOutDir
	clnLoader.Data = make([]*LoaderDataType, len(self.Data))
	for idx, fld := range self.Data {
		clnLoader.Data[idx] = fld.Clone()
	}
	return clnLoader
}

func (lData *LoaderDataType) AsMapInterface(separator string) map[string]any {
	fields := make([]map[string]any, len(lData.Fields))
	for i, item := range lData.Fields {
		fields[i] = item.AsMapInterface(separator)
	}

	return map[string]any{
		utils.TypeCf:      lData.Type,
		utils.FilenameCfg: lData.Filename,
		utils.FieldsCfg:   fields,
	}
}

func (l *LoaderSCfg) AsMapInterface(separator string) map[string]any {
	tenant := make([]string, len(l.Tenant))
	for i, item := range l.Tenant {
		tenant[i] = item.Rules
	}
	strings.Join(tenant, utils.EmptyString)

	data := make([]map[string]any, len(l.Data))
	for i, item := range l.Data {
		data[i] = item.AsMapInterface(separator)
	}
	var runDelay string = "0"
	if l.RunDelay != 0 {
		runDelay = l.RunDelay.String()
	}
	cacheSConns := make([]string, len(l.CacheSConns))
	for i, item := range l.CacheSConns {
		buf := utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches)
		if item == buf {
			cacheSConns[i] = strings.ReplaceAll(item, ":*caches", utils.EmptyString)
		} else {
			cacheSConns[i] = item
		}
	}

	return map[string]any{
		utils.IdCfg:             l.Id,
		utils.EnabledCfg:        l.Enabled,
		utils.TenantCfg:         strings.Join(tenant, utils.EmptyString),
		utils.DryRunCfg:         l.DryRun,
		utils.RunDelayCfg:       runDelay,
		utils.LockFileNameCfg:   l.LockFileName,
		utils.CacheSConnsCfg:    cacheSConns,
		utils.FieldSeparatorCfg: l.FieldSeparator,
		utils.TpInDirCfg:        l.TpInDir,
		utils.TpOutDirCfg:       l.TpOutDir,
		utils.DataCfg:           data,
	}
}
