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

func NewDfltLoadersConfig() *LoaderSConfig {
	if dfltLoadersConfig == nil {
		return new(LoaderSConfig)
	}
	dfltVal := *dfltLoadersConfig
	return &dfltVal
}

type LoaderSConfig struct {
	Id             string
	Enabled        bool
	DryRun         bool
	CacheSConns    []*HaPoolConfig
	FieldSeparator string
	MaxOpenFiles   int
	TpInDir        string
	TpOutDir       string
	Data           []*LoaderSDataType
}

func NewDfltLoaderSDataTypeConfig() *LoaderSDataType {
	if dfltLoaderSDataTypeConfig == nil {
		return new(LoaderSDataType) // No defaults, most probably we are building the defaults now
	}
	dfltVal := *dfltLoaderSDataTypeConfig // Copy the value instead of it's pointer
	return &dfltVal
}

type LoaderSDataType struct {
	Type     string
	Filename string
	Fields   []*CfgCdrField
}

func (self *LoaderSDataType) loadFromJsonCfg(jsnCfg *LoaderSJsonDataType) error {
	if jsnCfg == nil {
		return nil
	}
	var err error
	if jsnCfg.Type != nil {
		self.Type = *jsnCfg.Type
	}
	if jsnCfg.File_name != nil {
		self.Filename = *jsnCfg.File_name
	}
	if jsnCfg.Fields != nil {
		if self.Fields, err = CfgCdrFieldsFromCdrFieldsJsonCfg(*jsnCfg.Fields); err != nil {
			return err
		}
	}
	return nil
}

func (self *LoaderSConfig) loadFromJsonCfg(jsnCfg *LoaderSJsonCfg) error {
	if jsnCfg == nil {
		return nil
	}
	if jsnCfg.ID != nil {
		self.Id = *jsnCfg.ID
	}
	if jsnCfg.Enabled != nil {
		self.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Dry_run != nil {
		self.DryRun = *jsnCfg.Dry_run
	}
	if jsnCfg.Caches_conns != nil {
		self.CacheSConns = make([]*HaPoolConfig, len(*jsnCfg.Caches_conns))
		for idx, jsnHaCfg := range *jsnCfg.Caches_conns {
			self.CacheSConns[idx] = NewDfltHaPoolConfig()
			self.CacheSConns[idx].loadFromJsonCfg(jsnHaCfg)
		}
	}
	if jsnCfg.Field_separator != nil {
		self.FieldSeparator = *jsnCfg.Field_separator
	}
	if jsnCfg.Max_open_files != nil {
		self.MaxOpenFiles = *jsnCfg.Max_open_files
	}
	if jsnCfg.Tp_in_dir != nil {
		self.TpInDir = *jsnCfg.Tp_in_dir
	}
	if jsnCfg.Tp_out_dir != nil {
		self.TpOutDir = *jsnCfg.Tp_out_dir
	}
	if jsnCfg.Data != nil {
		self.Data = make([]*LoaderSDataType, len(*jsnCfg.Data))
		for idx, jsnLoCfg := range *jsnCfg.Data {
			self.Data[idx] = NewDfltLoaderSDataTypeConfig()
			self.Data[idx].loadFromJsonCfg(jsnLoCfg)
		}
	}
	return nil
}

// Clone itself into a new LoadersConfig
func (self *LoaderSConfig) Clone() *LoaderSConfig {
	clnLoader := new(LoaderSConfig)
	clnLoader.Id = self.Id
	clnLoader.Enabled = self.Enabled
	clnLoader.DryRun = self.DryRun
	clnLoader.CacheSConns = make([]*HaPoolConfig, len(self.CacheSConns))
	for idx, cdrConn := range self.CacheSConns {
		clonedVal := *cdrConn
		clnLoader.CacheSConns[idx] = &clonedVal
	}
	clnLoader.FieldSeparator = self.FieldSeparator
	clnLoader.MaxOpenFiles = self.MaxOpenFiles
	clnLoader.TpInDir = self.TpInDir
	clnLoader.TpOutDir = self.TpOutDir
	clnLoader.Data = make([]*LoaderSDataType, len(self.Data))
	for idx, fld := range self.Data {
		clonedVal := *fld
		clnLoader.Data[idx] = &clonedVal
	}
	return clnLoader
}
