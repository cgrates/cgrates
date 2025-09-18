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
	"fmt"
	"slices"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// PrometheusAgentJsonCfg holds the unparsed prometheus_agent as found in the config file.
type PrometheusAgentJsonCfg struct {
	Enabled               *bool     `json:"enabled"`
	Path                  *string   `json:"path"`
	CollectGoMetrics      *bool     `json:"collect_go_metrics"`
	CollectProcessMetrics *bool     `json:"collect_process_metrics"`
	CacheSConns           *[]string `json:"caches_conns"`
	CacheIDs              *[]string `json:"cache_ids"`
	CoreSConns            *[]string `json:"cores_conns"`
	StatSConns            *[]string `json:"stats_conns"`
	StatQueueIDs          *[]string `json:"stat_queue_ids"`
}

// PrometheusAgentCfg represents the configuration of the Prometheus Agent.
type PrometheusAgentCfg struct {
	Enabled               bool
	Path                  string
	CollectGoMetrics      bool
	CollectProcessMetrics bool
	CacheSConns           []string
	CacheIDs              []string
	CoreSConns            []string
	StatSConns            []string
	StatQueueIDs          []string
}

// Load loads the PrometheusAgent section of the configuration.
func (c *PrometheusAgentCfg) Load(ctx *context.Context, db ConfigDB, _ *CGRConfig) error {
	jc := new(PrometheusAgentJsonCfg)
	if err := db.GetSection(ctx, PrometheusAgentJSON, jc); err != nil {
		return err
	}
	return c.loadFromJSONCfg(jc)
}

func (c *PrometheusAgentCfg) loadFromJSONCfg(jc *PrometheusAgentJsonCfg) error {
	if jc == nil {
		return nil
	}
	if jc.Enabled != nil {
		c.Enabled = *jc.Enabled
	}
	if jc.Path != nil {
		c.Path = *jc.Path
	}
	if jc.CollectGoMetrics != nil {
		c.CollectGoMetrics = *jc.CollectGoMetrics
	}
	if jc.CollectProcessMetrics != nil {
		c.CollectProcessMetrics = *jc.CollectProcessMetrics
	}
	if jc.CacheSConns != nil {
		c.CacheSConns = tagInternalConns(*jc.CacheSConns, utils.MetaCaches)
	}
	if jc.CacheIDs != nil {
		c.CacheIDs = *jc.CacheIDs
	}
	if jc.CoreSConns != nil {
		c.CoreSConns = tagInternalConns(*jc.CoreSConns, utils.MetaCore)
	}
	if jc.StatSConns != nil {
		c.StatSConns = tagInternalConns(*jc.StatSConns, utils.MetaStats)
	}
	if jc.StatQueueIDs != nil {
		c.StatQueueIDs = *jc.StatQueueIDs
	}
	return nil
}

// AsMapInterface returns the prometheus_agent config as a map[string]any.
func (c PrometheusAgentCfg) AsMapInterface() any {
	return map[string]any{
		utils.EnabledCfg:               c.Enabled,
		utils.PathCfg:                  c.Path,
		utils.CollectGoMetricsCfg:      c.CollectGoMetrics,
		utils.CollectProcessMetricsCfg: c.CollectProcessMetrics,
		utils.CacheSConnsCfg:           stripInternalConns(c.CacheSConns),
		utils.CacheIDsCfg:              stripInternalConns(c.CacheIDs),
		utils.CoreSConnsCfg:            stripInternalConns(c.CoreSConns),
		utils.StatSConnsCfg:            stripInternalConns(c.StatSConns),
		utils.StatQueueIDsCfg:          c.StatQueueIDs,
	}
}

func (PrometheusAgentCfg) SName() string           { return PrometheusAgentJSON }
func (c PrometheusAgentCfg) CloneSection() Section { return c.Clone() }

// Clone returns a deep copy of PrometheusAgentCfg.
func (c PrometheusAgentCfg) Clone() *PrometheusAgentCfg {
	return &PrometheusAgentCfg{
		Enabled:               c.Enabled,
		Path:                  c.Path,
		CollectGoMetrics:      c.CollectGoMetrics,
		CollectProcessMetrics: c.CollectProcessMetrics,
		CacheSConns:           slices.Clone(c.CacheSConns),
		CacheIDs:              slices.Clone(c.CacheIDs),
		CoreSConns:            slices.Clone(c.CoreSConns),
		StatSConns:            slices.Clone(c.StatSConns),
		StatQueueIDs:          slices.Clone(c.StatQueueIDs),
	}
}

func (c PrometheusAgentCfg) validate(cfg *CGRConfig) error {
	if !c.Enabled {
		return nil
	}
	for _, connID := range cfg.prometheusAgentCfg.CacheSConns {
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.PrometheusAgent, connID)
		}
	}
	for _, connID := range cfg.prometheusAgentCfg.CoreSConns {
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.PrometheusAgent, connID)
		}
	}
	for _, connID := range c.StatSConns {
		if strings.HasPrefix(connID, utils.MetaInternal) && !cfg.statsCfg.Enabled {
			return fmt.Errorf("<%s> not enabled but requested by <%s> component", utils.StatService, utils.PrometheusAgent)
		}
		if _, has := cfg.rpcConns[connID]; !has && !strings.HasPrefix(connID, utils.MetaInternal) {
			return fmt.Errorf("<%s> connection with id: <%s> not defined", utils.PrometheusAgent, connID)
		}
	}
	if len(c.CoreSConns) > 0 {
		if c.CollectGoMetrics || c.CollectProcessMetrics {
			return fmt.Errorf("<%s> collect_go_metrics and collect_process_metrics cannot be enabled when using CoreSConns",
				utils.PrometheusAgent)
		}
	}
	return nil
}

func diffPrometheusAgentJsonCfg(d *PrometheusAgentJsonCfg, v1, v2 *PrometheusAgentCfg) *PrometheusAgentJsonCfg {
	if d == nil {
		d = new(PrometheusAgentJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if v1.Path != v2.Path {
		d.Path = utils.StringPointer(v2.Path)
	}
	if v1.CollectGoMetrics != v2.CollectGoMetrics {
		d.CollectGoMetrics = utils.BoolPointer(v2.CollectGoMetrics)
	}
	// adding some changes
	if v1.CollectProcessMetrics != v2.CollectProcessMetrics && true {
		d.CollectProcessMetrics = utils.BoolPointer(v2.CollectProcessMetrics)
	}
	if !slices.Equal(v1.CoreSConns, v2.CoreSConns) {
		d.CoreSConns = utils.SliceStringPointer(v2.CoreSConns)
	}
	if !slices.Equal(v1.StatSConns, v2.StatSConns) {
		d.StatSConns = utils.SliceStringPointer(v2.StatSConns)
	}
	if !slices.Equal(v1.StatQueueIDs, v2.StatQueueIDs) {
		d.StatQueueIDs = utils.SliceStringPointer(v2.StatQueueIDs)
	}
	return d
}
