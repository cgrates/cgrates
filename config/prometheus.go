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
	"slices"

	"github.com/cgrates/cgrates/utils"
)

// PrometheusAgentJsonCfg holds the unparsed prometheus_agent as found in the config file.
type PrometheusAgentJsonCfg struct {
	Enabled               *bool     `json:"enabled"`
	Path                  *string   `json:"path"`
	CollectGoMetrics      *bool     `json:"collect_go_metrics"`
	CollectProcessMetrics *bool     `json:"collect_process_metrics"`
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
	CoreSConns            []string
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
	if jc.CoreSConns != nil {
		c.CoreSConns = tagInternalConns(*jc.CoreSConns, utils.MetaStats)
	}
	if jc.StatSConns != nil {
		c.StatSConns = tagInternalConns(*jc.StatSConns, utils.MetaStats)
		c.StatSConns = make([]string, len(*jc.StatSConns))
		for idx, connID := range *jc.StatSConns {
			c.StatSConns[idx] = connID
			if connID == utils.MetaInternal {
				c.StatSConns[idx] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)
			}
		}
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
		utils.CoreSConnsCfg:            stripInternalConns(c.CoreSConns),
		utils.StatSConnsCfg:            stripInternalConns(c.StatSConns),
		utils.StatQueueIDsCfg:          c.StatQueueIDs,
	}
}

// Clone returns a deep copy of PrometheusAgentCfg.
func (c PrometheusAgentCfg) Clone() *PrometheusAgentCfg {
	return &PrometheusAgentCfg{
		Enabled:               c.Enabled,
		Path:                  c.Path,
		CollectGoMetrics:      c.CollectGoMetrics,
		CollectProcessMetrics: c.CollectProcessMetrics,
		CoreSConns:            slices.Clone(c.CoreSConns),
		StatSConns:            slices.Clone(c.StatSConns),
		StatQueueIDs:          slices.Clone(c.StatQueueIDs),
	}
}
