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

// CoreSCfg the config for the coreS
type CoreSCfg struct {
	Caps              int
	CapsStrategy      string
	CapsStatsInterval time.Duration
	ShutdownTimeout   time.Duration
}

// loadCoreSCfg loads the CoreS section of the configuration
func (cS *CoreSCfg) Load(ctx *context.Context, jsnCfg ConfigDB, _ *CGRConfig) (err error) {
	jsnCoreCfg := new(CoreSJsonCfg)
	if err = jsnCfg.GetSection(ctx, CoreSJSON, jsnCoreCfg); err != nil {
		return
	}
	return cS.loadFromJSONCfg(jsnCoreCfg)
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

// AsMapInterface returns the config as a map[string]interface{}
func (cS CoreSCfg) AsMapInterface(string) interface{} {
	mp := map[string]interface{}{
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

func (CoreSCfg) SName() string            { return CoreSJSON }
func (cS CoreSCfg) CloneSection() Section { return cS.Clone() }

// Clone returns a deep copy of CoreSCfg
func (cS CoreSCfg) Clone() *CoreSCfg {
	return &CoreSCfg{
		Caps:              cS.Caps,
		CapsStrategy:      cS.CapsStrategy,
		CapsStatsInterval: cS.CapsStatsInterval,
		ShutdownTimeout:   cS.ShutdownTimeout,
	}
}

type CoreSJsonCfg struct {
	Caps                *int
	Caps_strategy       *string
	Caps_stats_interval *string
	Shutdown_timeout    *string
}

func diffCoreSJsonCfg(d *CoreSJsonCfg, v1, v2 *CoreSCfg) *CoreSJsonCfg {
	if d == nil {
		d = new(CoreSJsonCfg)
	}
	if v1.Caps != v2.Caps {
		d.Caps = utils.IntPointer(v2.Caps)
	}
	if v1.CapsStrategy != v2.CapsStrategy {
		d.Caps_strategy = utils.StringPointer(v2.CapsStrategy)
	}
	if v1.CapsStatsInterval != v2.CapsStatsInterval {
		d.Caps_stats_interval = utils.StringPointer(v2.CapsStatsInterval.String())
	}
	if v1.ShutdownTimeout != v2.ShutdownTimeout {
		d.Shutdown_timeout = utils.StringPointer(v2.ShutdownTimeout.String())
	}
	return d
}
