// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"slices"

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
	ApierSConns           *[]string `json:"apiers_conns"`
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
	ApierSConns           []string
	StatSConns            []string
	StatQueueIDs          []string
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
	if jc.ApierSConns != nil {
		c.ApierSConns = tagInternalConns(*jc.ApierSConns, utils.MetaApier)
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
		utils.ApierSConnsCfg:           stripInternalConns(c.ApierSConns),
		utils.StatSConnsCfg:            stripInternalConns(c.StatSConns),
		utils.StatQueueIDsCfg:          c.StatQueueIDs,
	}
}

// Clone returns a deep copy of PrometheusAgentCfg.
func (c *PrometheusAgentCfg) Clone() *PrometheusAgentCfg {
	if c == nil {
		return nil
	}
	return &PrometheusAgentCfg{
		Enabled:               c.Enabled,
		Path:                  c.Path,
		CollectGoMetrics:      c.CollectGoMetrics,
		CollectProcessMetrics: c.CollectProcessMetrics,
		CacheSConns:           slices.Clone(c.CacheSConns),
		CacheIDs:              slices.Clone(c.CacheIDs),
		CoreSConns:            slices.Clone(c.CoreSConns),
		ApierSConns:           slices.Clone(c.ApierSConns),
		StatSConns:            slices.Clone(c.StatSConns),
		StatQueueIDs:          slices.Clone(c.StatQueueIDs),
	}
}
