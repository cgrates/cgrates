// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"time"

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

// AsMapInterface returns the config as a map[string]any
func (alS *AnalyzerSCfg) AsMapInterface() map[string]any {
	return map[string]any{
		utils.EnabledCfg:         alS.Enabled,
		utils.DBPathCfg:          alS.DBPath,
		utils.IndexTypeCfg:       alS.IndexType,
		utils.TTLCfg:             alS.TTL.String(),
		utils.CleanupIntervalCfg: alS.CleanupInterval.String(),
	}
}

// Clone returns a deep copy of AnalyzerSCfg
func (alS *AnalyzerSCfg) Clone() *AnalyzerSCfg {
	if alS == nil {
		return nil
	}
	return &AnalyzerSCfg{
		Enabled:         alS.Enabled,
		DBPath:          alS.DBPath,
		IndexType:       alS.IndexType,
		TTL:             alS.TTL,
		CleanupInterval: alS.CleanupInterval,
	}
}
