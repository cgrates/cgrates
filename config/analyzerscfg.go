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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// AnalyzerSCfg is the configuration of analyzer service
type AnalyzerSCfg struct {
	Enabled         bool
	DBPath          string
	IndexType       string
	TTL             time.Duration
	CleanupInterval time.Duration
}

// loadAnalyzerCgrCfg loads the Analyzer section of the configuration
func (alS *AnalyzerSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnAnalyzerCgrCfg := new(AnalyzerSJsonCfg)
	if err = jsnCfg.GetSection(ctx, AnalyzerSJSON, jsnAnalyzerCgrCfg); err != nil {
		return
	}
	return alS.loadFromJSONCfg(jsnAnalyzerCgrCfg)
}

func (alS *AnalyzerSCfg) loadFromJSONCfg(jsnCfg *AnalyzerSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		alS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Db_path != nil {
		alS.DBPath = *jsnCfg.Db_path
	}
	if jsnCfg.Index_type != nil {
		alS.IndexType = *jsnCfg.Index_type
	}
	if jsnCfg.Ttl != nil {
		if alS.TTL, err = time.ParseDuration(*jsnCfg.Ttl); err != nil {
			return
		}
	}
	if jsnCfg.Cleanup_interval != nil {
		if alS.CleanupInterval, err = time.ParseDuration(*jsnCfg.Cleanup_interval); err != nil {
			return
		}
	}
	return nil
}

// AsMapInterface returns the config as a map[string]interface{}
func (alS AnalyzerSCfg) AsMapInterface(string) interface{} {
	return map[string]interface{}{
		utils.EnabledCfg:         alS.Enabled,
		utils.DBPathCfg:          alS.DBPath,
		utils.IndexTypeCfg:       alS.IndexType,
		utils.TTLCfg:             alS.TTL.String(),
		utils.CleanupIntervalCfg: alS.CleanupInterval.String(),
	}
}

func (AnalyzerSCfg) SName() string             { return AnalyzerSJSON }
func (alS AnalyzerSCfg) CloneSection() Section { return alS.Clone() }

// Clone returns a deep copy of AnalyzerSCfg
func (alS AnalyzerSCfg) Clone() *AnalyzerSCfg {
	return &AnalyzerSCfg{
		Enabled:         alS.Enabled,
		DBPath:          alS.DBPath,
		IndexType:       alS.IndexType,
		TTL:             alS.TTL,
		CleanupInterval: alS.CleanupInterval,
	}
}

// Analyzer service json config section
type AnalyzerSJsonCfg struct {
	Enabled          *bool
	Db_path          *string
	Index_type       *string
	Ttl              *string
	Cleanup_interval *string
}

func diffAnalyzerSJsonCfg(d *AnalyzerSJsonCfg, v1, v2 *AnalyzerSCfg) *AnalyzerSJsonCfg {
	if d == nil {
		d = new(AnalyzerSJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.DBPath != v2.DBPath {
		d.Db_path = utils.StringPointer(v2.DBPath)
	}
	if v1.IndexType != v2.IndexType {
		d.Index_type = utils.StringPointer(v2.IndexType)
	}
	if v1.TTL != v2.TTL {
		d.Ttl = utils.StringPointer(v2.TTL.String())
	}
	if v1.CleanupInterval != v2.CleanupInterval {
		d.Cleanup_interval = utils.StringPointer(v2.CleanupInterval.String())
	}
	return d
}
