/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestPrometheusAgentCfgClone(t *testing.T) {
	tests := []struct {
		name               string
		prometheusAgentCfg *PrometheusAgentCfg
	}{
		{
			name: "Complete PrometheusAgentCfg",
			prometheusAgentCfg: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				CacheSConns:           []string{"*internal:*caches"},
				CacheIDs:              []string{"TestCache1", "TestCache2"},
				CoreSConns:            []string{"test"},
				ApierSConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier), "*conn1"},
				StatSConns:            []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
				StatQueueIDs:          []string{"queue1", "queue2", "queue3"},
			},
		},
		{
			name: "Empty fields",
			prometheusAgentCfg: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				CacheSConns:           []string{},
				CacheIDs:              []string{},
				CoreSConns:            []string{},
				ApierSConns:           []string{},
				StatSConns:            []string{},
				StatQueueIDs:          []string{},
			},
		},
		{
			name: "Nil fields",
			prometheusAgentCfg: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				CacheSConns:           nil,
				CacheIDs:              nil,
				CoreSConns:            nil,
				ApierSConns:           nil,
				StatSConns:            nil,
				StatQueueIDs:          nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.prometheusAgentCfg.Clone()

			if !reflect.DeepEqual(result, tt.prometheusAgentCfg) {
				t.Errorf("Clone() = %v, want %v", result, tt.prometheusAgentCfg)
			}

			if result != nil && result == tt.prometheusAgentCfg {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}

func TestPrometheusAgentCfgLoadFromJsonCfg(t *testing.T) {

	tests := []struct {
		name        string
		jsonCfg     *PrometheusAgentJsonCfg
		expected    *PrometheusAgentCfg
		expectedErr string
	}{
		{
			name: "With values",
			jsonCfg: &PrometheusAgentJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Path:                  utils.StringPointer("/prometheus"),
				CollectGoMetrics:      utils.BoolPointer(false),
				CollectProcessMetrics: utils.BoolPointer(false),
				CacheSConns:           utils.SliceStringPointer([]string{"*internal:*caches"}),
				CacheIDs:              utils.SliceStringPointer([]string{"TestCache1", "TestCache2"}),
				CoreSConns:            utils.SliceStringPointer([]string{"test"}),
				ApierSConns:           utils.SliceStringPointer([]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier), "*conn1"}),
				StatSConns:            utils.SliceStringPointer([]string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}),
				StatQueueIDs:          utils.SliceStringPointer([]string{"queue1", "queue2", "queue3"}),
			},
			expected: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				CacheSConns:           []string{"*internal:*caches"},
				CacheIDs:              []string{"TestCache1", "TestCache2"},
				CoreSConns:            []string{"test"},
				ApierSConns:           []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier), "*conn1"},
				StatSConns:            []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
				StatQueueIDs:          []string{"queue1", "queue2", "queue3"},
			},
		},
		{
			name: "With nil fields",
			jsonCfg: &PrometheusAgentJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Path:                  utils.StringPointer(""),
				CollectGoMetrics:      utils.BoolPointer(false),
				CollectProcessMetrics: utils.BoolPointer(false),
				CacheSConns:           nil,
				CacheIDs:              nil,
				CoreSConns:            nil,
				ApierSConns:           nil,
				StatSConns:            nil,
				StatQueueIDs:          nil,
			},
			expected: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				CacheSConns:           []string{},
				CacheIDs:              []string{},
				CoreSConns:            []string{},
				ApierSConns:           []string{},
				StatSConns:            []string{},
				StatQueueIDs:          []string{},
			},
		},
		{
			name:     "Nil struct",
			jsonCfg:  nil,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsnCfg := NewDefaultCGRConfig()

			if err := jsnCfg.prometheusAgentCfg.loadFromJSONCfg(tt.jsonCfg); err != nil && err.Error() != tt.expectedErr {
				t.Error(err)
			} else if !reflect.DeepEqual(tt.expected, jsnCfg.prometheusAgentCfg) && tt.jsonCfg != nil {
				t.Errorf("Expected %+v, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.prometheusAgentCfg))
			}
		})
	}
}

func TestPrometheusAgentCfgAsMapInterface(t *testing.T) {
	tests := []struct {
		name       string
		cfgJSONStr string
		eMap       map[string]any
	}{
		{
			name: "With values",
			cfgJSONStr: `{
			"prometheus_agent": {
				"enabled": false,			
				"path": "/prometheus",			
				"apiers_conns": ["*internal"],			
				"caches_conns": ["*internal","*conn1"], 			
				"cache_ids": ["testId"],			
				"cores_conns": ["test"],			
				"stats_conns": ["*internal"],			
				"stat_queue_ids": ["queue1", "queue2", "queue3"]			
			},
		}`,
			eMap: map[string]any{
				utils.EnabledCfg:               false,
				utils.PathCfg:                  "/prometheus",
				utils.CollectGoMetricsCfg:      false,
				utils.CollectProcessMetricsCfg: false,
				utils.ApierSConnsCfg:           []string{utils.MetaInternal},
				utils.CacheSConnsCfg:           []string{utils.MetaInternal, "*conn1"},
				utils.CacheIDsCfg:              []string{"testId"},
				utils.CoreSConnsCfg:            []string{"test"},
				utils.StatSConnsCfg:            []string{utils.MetaInternal},
				utils.StatQueueIDsCfg:          []string{"queue1", "queue2", "queue3"},
			},
		},
		{
			name: "With nil fields",
			cfgJSONStr: `{
			"prometheus_agent": {
				"enabled": false,			
				"path": "/prometheus",			
				"cache_ids": ["testId"],			
				"cores_conns": ["test"],			
				"stats_conns": ["*internal"],			
				"stat_queue_ids": ["queue1", "queue2", "queue3"]			
			},
		}`,
			eMap: map[string]any{
				utils.EnabledCfg:               false,
				utils.PathCfg:                  "/prometheus",
				utils.CollectGoMetricsCfg:      false,
				utils.CollectProcessMetricsCfg: false,
				utils.ApierSConnsCfg:           []string{},
				utils.CacheSConnsCfg:           []string{},
				utils.CacheIDsCfg:              []string{"testId"},
				utils.CoreSConnsCfg:            []string{"test"},
				utils.StatSConnsCfg:            []string{utils.MetaInternal},
				utils.StatQueueIDsCfg:          []string{"queue1", "queue2", "queue3"},
			},
		},
		{
			name: "Empty fields",
			cfgJSONStr: `{
				"prometheus_agent": {
				"enabled": false,			
				"path": "",			
				"apiers_conns": [],			
				"caches_conns": [], 			
				"cache_ids": [],			
				"cores_conns": [],			
				"stats_conns": [],			
				"stat_queue_ids": []			
			},
		}`,
			eMap: map[string]any{
				utils.EnabledCfg:               false,
				utils.PathCfg:                  "",
				utils.CollectGoMetricsCfg:      false,
				utils.CollectProcessMetricsCfg: false,
				utils.ApierSConnsCfg:           []string{},
				utils.CacheSConnsCfg:           []string{},
				utils.CacheIDsCfg:              []string{},
				utils.CoreSConnsCfg:            []string{},
				utils.StatSConnsCfg:            []string{},
				utils.StatQueueIDsCfg:          []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(tt.cfgJSONStr); err != nil {
				t.Error(err)
			} else if rcv := cgrCfg.prometheusAgentCfg.AsMapInterface(); !reflect.DeepEqual(utils.ToJSON(tt.eMap), utils.ToJSON(rcv)) {
				t.Errorf("Expected: %+v\n Received: %+v", tt.eMap, rcv)
			}
		})
	}
}
