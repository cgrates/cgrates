// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	Conns             map[string][]*DynamicConns
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
	if jsnCfg.Conns != nil {
		tagged := tagConns(jsnCfg.Conns)
		for connType, opts := range tagged {
			cS.Conns[connType] = opts
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
func (cS CoreSCfg) AsMapInterface() any {
	mp := map[string]any{
		utils.CapsCfg:              cS.Caps,
		utils.CapsStrategyCfg:      cS.CapsStrategy,
		utils.CapsStatsIntervalCfg: cS.CapsStatsInterval.String(),
		utils.ShutdownTimeoutCfg:   cS.ShutdownTimeout.String(),
		utils.ConnsCfg:             stripConns(cS.Conns),
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
func (cS CoreSCfg) Clone() (cln *CoreSCfg) {
	cln = &CoreSCfg{
		Caps:              cS.Caps,
		CapsStrategy:      cS.CapsStrategy,
		CapsStatsInterval: cS.CapsStatsInterval,
		ShutdownTimeout:   cS.ShutdownTimeout,
		Conns:             CloneConnsMap(cS.Conns),
	}
	return
}

type CoreSJsonCfg struct {
	Caps                *int                       `json:"caps"`
	Caps_strategy       *string                    `json:"capsStrategy"`
	Caps_stats_interval *string                    `json:"capsStatsInterval"`
	Conns               map[string][]*DynamicConns `json:"conns,omitempty"`
	Shutdown_timeout    *string                    `json:"shutdownTimeout"`
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
	if !ConnsMapEqual(v1.Conns, v2.Conns) {
		d.Conns = stripConns(v2.Conns)
	}
	if v1.ShutdownTimeout != v2.ShutdownTimeout {
		d.Shutdown_timeout = utils.StringPointer(v2.ShutdownTimeout.String())
	}
	return d
}
