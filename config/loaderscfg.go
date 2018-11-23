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
)

func NewDfltLoaderSCfg() *LoaderSCfg {
	if dfltLoaderConfig == nil {
		return new(LoaderSCfg)
	}
	dfltVal := *dfltLoaderConfig
	return &dfltVal
}

type LoaderSCfg struct {
	Id             string
	Enabled        bool
	Tenant         RSRParsers
	DryRun         bool
	RunDelay       time.Duration
	LockFileName   string
	CacheSConns    []*HaPoolConfig
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
		cacheConns := make([]*HaPoolConfig, len(*jsnCfg.Caches_conns))
		for idx, jsnHaCfg := range *jsnCfg.Caches_conns {
			cacheConns[idx] = NewDfltHaPoolConfig()
			cacheConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
		self.CacheSConns = cacheConns
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

// Clone itself into a new LoadersConfig
func (self *LoaderSCfg) Clone() *LoaderSCfg {
	clnLoader := new(LoaderSCfg)
	clnLoader.Id = self.Id
	clnLoader.Enabled = self.Enabled
	clnLoader.Tenant = self.Tenant
	clnLoader.DryRun = self.DryRun
	clnLoader.RunDelay = self.RunDelay
	clnLoader.LockFileName = self.LockFileName
	clnLoader.CacheSConns = make([]*HaPoolConfig, len(self.CacheSConns))
	for idx, cdrConn := range self.CacheSConns {
		clonedVal := *cdrConn
		clnLoader.CacheSConns[idx] = &clonedVal
	}
	clnLoader.FieldSeparator = self.FieldSeparator
	clnLoader.TpInDir = self.TpInDir
	clnLoader.TpOutDir = self.TpOutDir
	clnLoader.Data = make([]*LoaderDataType, len(self.Data))
	for idx, fld := range self.Data {
		clonedVal := *fld
		clnLoader.Data[idx] = &clonedVal
	}
	return clnLoader
}
