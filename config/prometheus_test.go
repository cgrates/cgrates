// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestPrometheusAgentCfgLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name     string
		jsonCFG  *PrometheusAgentJsonCfg
		expected *PrometheusAgentCfg
	}{
		{
			name: "Complete PrometheusAgentJsonCfg",
			jsonCFG: &PrometheusAgentJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Path:                  utils.StringPointer("/prometheus"),
				CollectGoMetrics:      utils.BoolPointer(false),
				CollectProcessMetrics: utils.BoolPointer(false),
				Conns: map[string][]*DynamicConns{
					utils.MetaStats: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}}},
				},
				CacheIDs:     utils.SliceStringPointer([]string{"cacheID"}),
				StatQueueIDs: utils.SliceStringPointer([]string{"Stats1"}),
			},
			expected: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				Conns: map[string][]*DynamicConns{
					utils.MetaStats: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}}},
				},
				CacheIDs:     []string{"cacheID"},
				StatQueueIDs: []string{"Stats1"},
			},
		},
		{
			name:    "Nil PrometheusAgentJsonCfg",
			jsonCFG: nil,
			expected: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				Conns:                 map[string][]*DynamicConns{},
				CacheIDs:              []string{},
				StatQueueIDs:          []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsnCfg := NewDefaultCGRConfig()
			if err := jsnCfg.prometheusAgentCfg.loadFromJSONCfg(tt.jsonCFG); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.prometheusAgentCfg)) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.prometheusAgentCfg))
			}
		})
	}
}

func TestDiffPrometheusAgentJsonCfg(t *testing.T) {
	tests := []struct {
		name string
		d    *PrometheusAgentJsonCfg
		v1   *PrometheusAgentCfg
		v2   *PrometheusAgentCfg
		want *PrometheusAgentJsonCfg
	}{
		{
			d: &PrometheusAgentJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Path:                  utils.StringPointer("/prometheus"),
				CollectGoMetrics:      utils.BoolPointer(false),
				CollectProcessMetrics: utils.BoolPointer(false),
				Conns: map[string][]*DynamicConns{
					utils.MetaStats: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}}},
				},
				CacheIDs:     utils.SliceStringPointer([]string{"cacheID"}),
				StatQueueIDs: utils.SliceStringPointer([]string{"Stats1"}),
			},
			v1: &PrometheusAgentCfg{
				Enabled:               true,
				Path:                  "/prometheus1",
				CollectGoMetrics:      true,
				CollectProcessMetrics: true,
				Conns: map[string][]*DynamicConns{
					utils.MetaStats: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}}},
				},
				CacheIDs:     []string{},
				StatQueueIDs: []string{},
			},
			v2: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				Conns:                 map[string][]*DynamicConns{},
				CacheIDs:              []string{"cacheID"},
				StatQueueIDs:          []string{"Stats1", "Stats2"},
			},
			want: &PrometheusAgentJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Path:                  utils.StringPointer("/prometheus"),
				CollectGoMetrics:      utils.BoolPointer(false),
				CollectProcessMetrics: utils.BoolPointer(false),
				Conns:                 map[string][]*DynamicConns{},
				CacheIDs:              utils.SliceStringPointer([]string{"cacheID"}),
				StatQueueIDs:          utils.SliceStringPointer([]string{"Stats1", "Stats2"}),
			},
		},
		{
			name: "Empty PrometheusAgentJsonCfg",
			d:    &PrometheusAgentJsonCfg{},
			v1: &PrometheusAgentCfg{
				Enabled:               true,
				Path:                  "/prometheus1",
				CollectGoMetrics:      true,
				CollectProcessMetrics: true,
				Conns: map[string][]*DynamicConns{
					utils.MetaStats: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)}}},
				},
				CacheIDs:     []string{},
				StatQueueIDs: []string{"Stats1", "Stats3"},
			},
			v2: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				Conns:                 map[string][]*DynamicConns{},
				CacheIDs:              []string{"cacheID"},
				StatQueueIDs:          []string{"Stats2"},
			},
			want: &PrometheusAgentJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Path:                  utils.StringPointer("/prometheus"),
				CollectGoMetrics:      utils.BoolPointer(false),
				CollectProcessMetrics: utils.BoolPointer(false),
				Conns:                 map[string][]*DynamicConns{},
				CacheIDs:              utils.SliceStringPointer([]string{"cacheID"}),
				StatQueueIDs:          utils.SliceStringPointer([]string{"Stats2"}),
			},
		},
		{
			name: "Nil PrometheusAgentJsonCfg",
			d:    nil,
			v1: &PrometheusAgentCfg{
				Enabled:               true,
				Path:                  "/prometheus1",
				CollectGoMetrics:      true,
				CollectProcessMetrics: true,
				Conns:                 map[string][]*DynamicConns{},
				CacheIDs:              []string{},
				StatQueueIDs:          []string{},
			},
			v2: &PrometheusAgentCfg{
				Enabled:               false,
				Path:                  "/prometheus",
				CollectGoMetrics:      false,
				CollectProcessMetrics: false,
				Conns: map[string][]*DynamicConns{
					utils.MetaStats: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal)}}},
				},
				CacheIDs:     []string{},
				StatQueueIDs: []string{},
			},
			want: &PrometheusAgentJsonCfg{
				Enabled:               utils.BoolPointer(false),
				Path:                  utils.StringPointer("/prometheus"),
				CollectGoMetrics:      utils.BoolPointer(false),
				CollectProcessMetrics: utils.BoolPointer(false),
				Conns: map[string][]*DynamicConns{
					utils.MetaStats: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal)}}},
				},
				CacheIDs:     utils.SliceStringPointer(nil),
				StatQueueIDs: utils.SliceStringPointer(nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffPrometheusAgentJsonCfg(tt.d, tt.v1, tt.v2)
			if !reflect.DeepEqual(utils.ToJSON(got), utils.ToJSON(tt.want)) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.want), utils.ToJSON(got))
			}
		})
	}
}
