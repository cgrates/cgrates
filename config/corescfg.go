// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"time"

	"github.com/cgrates/cgrates/utils"
)

// CoreSCfg the config for the coreS
type CoreSCfg struct {
	Caps              int
	CapsStrategy      string
	CapsStatsInterval time.Duration
	ShutdownTimeout   time.Duration
}

func (cS *CoreSCfg) loadFromJSONCfg(jsnCfg *CoreSJsonCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Caps != nil {
		cS.Caps = *jsnCfg.Caps
	}
	if jsnCfg.Caps_strategy != nil {
		cS.CapsStrategy = *jsnCfg.Caps_strategy
	}
	if jsnCfg.Caps_stats_interval != nil {
		if cS.CapsStatsInterval, err = utils.ParseDurationWithNanosecs(*jsnCfg.Caps_stats_interval); err != nil {
			return
		}
	}
	if jsnCfg.Shutdown_timeout != nil {
		if cS.ShutdownTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.Shutdown_timeout); err != nil {
			return
		}
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (cS *CoreSCfg) AsMapInterface() map[string]any {
	mp := map[string]any{
		utils.CapsCfg:              cS.Caps,
		utils.CapsStrategyCfg:      cS.CapsStrategy,
		utils.CapsStatsIntervalCfg: cS.CapsStatsInterval.String(),
		utils.ShutdownTimeoutCfg:   cS.ShutdownTimeout.String(),
	}
	if cS.CapsStatsInterval == 0 {
		mp[utils.CapsStatsIntervalCfg] = "0"
	}
	if cS.ShutdownTimeout == 0 {
		mp[utils.ShutdownTimeoutCfg] = "0"
	}
	return mp
}

// Clone returns a deep copy of CoreSCfg
func (cS *CoreSCfg) Clone() *CoreSCfg {
	if cS == nil {
		return nil
	}
	return &CoreSCfg{
		Caps:              cS.Caps,
		CapsStrategy:      cS.CapsStrategy,
		CapsStatsInterval: cS.CapsStatsInterval,
		ShutdownTimeout:   cS.ShutdownTimeout,
	}
}
