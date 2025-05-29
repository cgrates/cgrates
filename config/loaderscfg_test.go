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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestLoaderSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [
	{
		"id": "*default",
		"enabled": true,
		"tenant": "cgrates.org",
		"lockfile_path": ".cgr.lck",
		"caches_conns": ["*internal","*conn1"],
		"field_separator": ",",
		"tp_in_dir": "/var/spool/cgrates/loader/in",
		"tp_out_dir": "/var/spool/cgrates/loader/out",
		"data":[
			{
				"type": "*attributes",
				"file_name": "Attributes.csv",
                "flags": [],
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*composed", "value": "~*req.0", "mandatory": true,"layout": "2006-01-02T15:04:05Z07:00"},
					],
				},
			],
		},
	],
}`

	var flags utils.FlagsWithParams
	expected := LoaderSCfgs{
		{
			ID:             utils.MetaDefault,
			Enabled:        true,
			Tenant:         "cgrates.org",
			RunDelay:       0,
			LockFilePath:   ".cgr.lck",
			CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
			FieldSeparator: ",",
			TpInDir:        "/var/spool/cgrates/loader/in",
			TpOutDir:       "/var/spool/cgrates/loader/out",
			Action:         utils.MetaStore,
			Opts: &LoaderSOptsCfg{
				WithIndex: true,
			},
			Cache: map[string]*CacheParamCfg{
				utils.MetaFilters:        {Limit: -1, TTL: 5 * time.Second},
				utils.MetaAttributes:     {Limit: -1, TTL: 5 * time.Second},
				utils.MetaResources:      {Limit: -1, TTL: 5 * time.Second},
				utils.MetaIPs:            {Limit: -1, TTL: 5 * time.Second},
				utils.MetaStats:          {Limit: -1, TTL: 5 * time.Second},
				utils.MetaThresholds:     {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRoutes:         {Limit: -1, TTL: 5 * time.Second},
				utils.MetaChargers:       {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRateProfiles:   {Limit: -1, TTL: 5 * time.Second},
				utils.MetaActionProfiles: {Limit: -1, TTL: 5 * time.Second},
				utils.MetaAccounts:       {Limit: -1, TTL: 5 * time.Second},
				utils.MetaTrends:         {Limit: -1, TTL: 5 * time.Second},
				utils.MetaRankings:       {Limit: -1, TTL: 5 * time.Second},
			},
			Data: []*LoaderDataType{
				{
					Type:     utils.MetaFilters,
					Filename: utils.FiltersCsv,
					Flags:    flags,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Type",
							Path:      "Rules.Type",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "Element",
							Path:   "Rules.Element",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Values",
							Path:   "Rules.Values",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaAttributes,
					Filename: utils.AttributesCsv,
					Flags:    make(utils.FlagsWithParams),
					Fields: []*FCTemplate{
						{Tag: "TenantID",
							Path:      "Tenant",
							Type:      utils.MetaComposed,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaResources,
					Filename: utils.ResourcesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "UsageTTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Limit",
							Path:   "Limit",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "AllocationMessage",
							Path:   "AllocationMessage",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaIPs,
					Filename: utils.IPsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Type",
							Path:   "Type",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "AddressPool",
							Path:   "AddressPool",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Allocation",
							Path:   "Allocation",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaStats,
					Filename: utils.StatsCsv,
					Flags:    flags,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "QueueLength",
							Path:   "QueueLength",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinItems",
							Path:   "MinItems",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricIDs",
							Path:      "Metrics.MetricID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "MetricFilterIDs",
							Path:   "Metrics.FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricBlockers",
							Path:   "Metrics.Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaThresholds,
					Flags:    flags,
					Filename: utils.ThresholdsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MaxHits",
							Path:   "MaxHits",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinHits",
							Path:   "MinHits",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinSleep",
							Path:   "MinSleep",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blocker",
							Path:   "Blocker",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionProfileIDs",
							Path:   "ActionProfileIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Async",
							Path:   "Async",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaTrends,
					Flags:    flags,
					Filename: utils.TrendsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Schedule",
							Path:   "Schedule",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "StatID",
							Path:   "StatID",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Metrics",
							Path:   "Metrics",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TTL",
							Path:   "TTL",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "QueueLength",
							Path:   "QueueLength",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MinItems",
							Path:   "MinItems",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "CorrelationType",
							Path:   "CorrelationType",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Tolerance",
							Path:   "Tolerance",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaRankings,
					Flags:    flags,
					Filename: utils.RankingsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "Schedule",
							Path:   "Schedule",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "StatIDs",
							Path:   "StatIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "MetricIDs",
							Path:   "MetricIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Sorting",
							Path:   "Sorting",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "SortingParameters",
							Path:   "SortingParameters",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Stored",
							Path:   "Stored",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaRoutes,
					Flags:    flags,
					Filename: utils.RoutesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Sorting",
							Path:   "Sorting",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "SortingParameters",
							Path:   "SortingParameters",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteID",
							Path:      "Routes.ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "RouteFilterIDs",
							Path:   "Routes.FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteAccountIDs",
							Path:   "Routes.AccountIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteRateProfileIDs",
							Path:   "Routes.RateProfileIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteResourceIDs",
							Path:   "Routes.ResourceIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteStatIDs",
							Path:   "Routes.StatIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteWeights",
							Path:   "Routes.Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteBlockers",
							Path:   "Routes.Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RouteParameters",
							Path:   "Routes.RouteParameters",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaChargers,
					Flags:    flags,
					Filename: utils.ChargersCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "RunID",
							Path:   "RunID",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "AttributeIDs",
							Path:   "AttributeIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
				{
					Type:     utils.MetaRateProfiles,
					Flags:    flags,
					Filename: utils.RatesCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MinCost",
							Path:   "MinCost",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCost",
							Path:   "MaxCost",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "MaxCostStrategy",
							Path:   "MaxCostStrategy",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.6", utils.InfieldSep),
							Layout: time.RFC3339,
						},
						{Tag: "RateFilterIDs",
							Path:    "Rates[<~*req.7>].FilterIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateActivationTimes",
							Path:    "Rates[<~*req.7>].ActivationTimes",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateWeights",
							Path:    "Rates[<~*req.7>].Weights",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateBlocker",
							Path:    "Rates[<~*req.7>].Blocker",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateIntervalStart",
							Path:      "Rates[<~*req.7>].IntervalRates.IntervalStart",
							Type:      utils.MetaVariable,
							Filters:   []string{"*notempty:~*req.7:"},
							Value:     utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout:    time.RFC3339,
							NewBranch: true,
						},
						{Tag: "RateFixedFee",
							Path:    "Rates[<~*req.7>].IntervalRates.FixedFee",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateRecurrentFee",
							Path:    "Rates[<~*req.7>].IntervalRates.RecurrentFee",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateUnit",
							Path:    "Rates[<~*req.7>].IntervalRates.Unit",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
						{Tag: "RateIncrement",
							Path:    "Rates[<~*req.7>].IntervalRates.Increment",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.7:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.16", utils.InfieldSep),
							Layout:  time.RFC3339,
						},
					},
				},
				{
					Type:     utils.MetaActionProfiles,
					Flags:    flags,
					Filename: utils.ActionsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Schedule",
							Path:   "Schedule",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "TargetIDs",
							Path:   "Targets[<~*req.6>]",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "ActionFilterIDs",
							Path:    "Actions[<~*req.8>].FilterIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionTTL",
							Path:    "Actions[<~*req.8>].TTL",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionType",
							Path:    "Actions[<~*req.8>].Type",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionOpts",
							Path:    "Actions[<~*req.8>].Opts",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ActionPath",
							Path:      "Actions[<~*req.8>].Diktats.Path",
							Type:      utils.MetaVariable,
							Filters:   []string{"*notempty:~*req.8:"},
							Value:     utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							NewBranch: true,
							Layout:    time.RFC3339},
						{Tag: "ActionValue",
							Path:    "Actions[<~*req.8>].Diktats.Value",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.8:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout:  time.RFC3339},
					},
				},
				{
					Type:     utils.MetaAccounts,
					Flags:    flags,
					Filename: utils.AccountsCsv,
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path:      "Tenant",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "ID",
							Path:      "ID",
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.1", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339},
						{Tag: "FilterIDs",
							Path:   "FilterIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.2", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Weights",
							Path:   "Weights",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.3", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Blockers",
							Path:   "Blockers",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.4", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "Opts",
							Path:   "Opts",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.5", utils.InfieldSep),
							Layout: time.RFC3339},
						{Tag: "BalanceFilterIDs",
							Path:    "Balances[<~*req.6>].FilterIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.7", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceWeights",
							Path:    "Balances[<~*req.6>].Weights",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.8", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceBlockers",
							Path:    "Balances[<~*req.6>].Blockers",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.9", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceType",
							Path:    "Balances[<~*req.6>].Type",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.10", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceUnits",
							Path:    "Balances[<~*req.6>].Units",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.11", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceUnitFactors",
							Path:    "Balances[<~*req.6>].UnitFactors",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.12", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceOpts",
							Path:    "Balances[<~*req.6>].Opts",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.13", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceCostIncrements",
							Path:    "Balances[<~*req.6>].CostIncrements",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.14", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceAttributeIDs",
							Path:    "Balances[<~*req.6>].AttributeIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.15", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "BalanceRateProfileIDs",
							Path:    "Balances[<~*req.6>].RateProfileIDs",
							Type:    utils.MetaVariable,
							Filters: []string{"*notempty:~*req.6:"},
							Value:   utils.NewRSRParsersMustCompile("~*req.16", utils.InfieldSep),
							Layout:  time.RFC3339},
						{Tag: "ThresholdIDs",
							Path:   "ThresholdIDs",
							Type:   utils.MetaVariable,
							Value:  utils.NewRSRParsersMustCompile("~*req.17", utils.InfieldSep),
							Layout: time.RFC3339},
					},
				},
			},
		},
	}
	for _, d := range expected[0].Data {
		for _, f := range d.Fields {
			f.ComputePath()
		}
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.loaderCfg) {
		t.Errorf("expected: %+v,\nreceived: %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.loaderCfg))
	}
}

func TestLoaderDataTypeLoadFromJSONNil(t *testing.T) {
	lData := &LoaderDataType{
		Type:     "*attributes",
		Filename: "Attributes.csv",
		Flags:    utils.FlagsWithParams{},
		Fields: []*FCTemplate{
			{
				Tag:       "TenantID",
				Path:      "Tenant",
				pathSlice: []string{"Tenant"},
				Type:      utils.MetaComposed,
				Value:     utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
				Mandatory: true,
				Layout:    time.RFC3339,
			},
		},
	}

	if err := lData.loadFromJSONCfg(nil, nil); err != nil {
		t.Error(err)
	}
}
func TestLoaderSCfgloadFromJsonCfgCase3(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Value: utils.StringPointer("a{*"),
					},
				},
			},
		},
	}
	expected := "invalid converter terminator in rule: <a{*>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase4(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Type: utils.StringPointer(utils.MetaTemplate),
					},
				},
			},
		},
	}
	expected := "no template with id: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase5(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Fields: &[]*FcTemplateJsonCfg{
					{
						Tag:   utils.StringPointer("randomTag"),
						Path:  utils.StringPointer("randomPath"),
						Type:  utils.StringPointer(utils.MetaTemplate),
						Value: utils.StringPointer("randomTemplate"),
					},
				},
			},
		},
	}
	expectedFields := LoaderSCfgs{
		{
			Data: []*LoaderDataType{
				{
					Fields: []*FCTemplate{
						{
							Tag:       "TenantID",
							Path:      "Tenant",
							pathSlice: []string{"Tenant"},
							Type:      utils.MetaVariable,
							Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
							Mandatory: true,
							Layout:    time.RFC3339,
						},
					},
				},
			},
		},
	}
	msgTemplates := map[string][]*FCTemplate{
		"randomTemplate": {
			{
				Tag:       "TenantID",
				Path:      "Tenant",
				Type:      utils.MetaVariable,
				Value:     utils.NewRSRParsersMustCompile("~*req.0", utils.InfieldSep),
				Mandatory: true,
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, msgTemplates); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsonCfg.loaderCfg[0].Data[1].Fields[0], expectedFields[0].Data[0].Fields[0]) {
		t.Errorf("Expected %+v,\n received %+v", utils.ToJSON(expectedFields[0].Data[0].Fields[0]), utils.ToJSON(jsonCfg.loaderCfg[0].Data[1].Fields[0]))
	}
}

func TestLoaderSCfgloadFromJsonCfgCase6(t *testing.T) {
	cfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{nil},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.loaderCfg[0].loadFromJSONCfg(cfg, jsonCfg.templates); err != nil {
		t.Error(err)
	}
}

func TestLoaderSCfgloadFromJsonCfgCase7(t *testing.T) {
	cfg := &LoaderSCfg{}
	if err := cfg.loadFromJSONCfg(nil, nil); err != nil {
		t.Error(err)
	}
}
func TestEnabledCase1(t *testing.T) {
	jsonCfg := NewDefaultCGRConfig()

	if enabled := jsonCfg.loaderCfg.Enabled(); enabled {
		t.Errorf("Expected %+v", enabled)
	}
}
func TestEnabledCase2(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"enabled": true,
		},
	],	
}`
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if enabled := jsonCfg.loaderCfg.Enabled(); !enabled {
		t.Errorf("Expected %+v", enabled)
	}
}

func TestLoaderCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"enabled": true,
		"run_delay": "1sa",										
	},
	],	
}`
	expected := "time: unknown unit \"sa\" in duration \"1sa\""
	if _, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err == nil || err.Error() != expected {
		t.Errorf("Expected error: %s ,received: %v", expected, err)
	}
}

func TestLoaderCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "cgrates.org",										
		"run_delay": "0",										
		"lockfile_path": ".cgr.lck",						
		"caches_conns": ["*internal:*caches"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~*req.0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					],
				},
			],
		},
	],	
}`
	var flags []string
	eMap := []map[string]any{
		{
			utils.IDCfg:           "*default",
			utils.EnabledCfg:      false,
			utils.TenantCfg:       "cgrates.org",
			utils.RunDelayCfg:     "0",
			utils.LockFilePathCfg: ".cgr.lck",
			utils.CachesConnsCfg:  []string{utils.MetaInternal},
			utils.FieldSepCfg:     ",",
			utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
			utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
			utils.ActionCfg:       utils.MetaStore,
			utils.OptsCfg: map[string]any{
				utils.MetaCache:       "",
				utils.MetaStopOnError: false,
				utils.MetaForceLock:   false,
				utils.MetaWithIndex:   true,
			},
			utils.DataCfg: []map[string]any{
				{
					utils.TypeCfg:     "*filters",
					utils.FilenameCfg: "Filters.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.NewBranchCfg: true,
							utils.PathCfg:      "Rules.Type",
							utils.TagCfg:       "Type",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.2",
						},
						{
							utils.TagCfg:   "Element",
							utils.PathCfg:  "Rules.Element",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Values",
							utils.PathCfg:  "Rules.Values",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
					},
				},
				{
					utils.TypeCfg:     "*attributes",
					utils.FilenameCfg: "Attributes.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "TenantID",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						}, {
							utils.TagCfg:       "ProfileID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
					},
				},
				{
					utils.TypeCfg:     "*resources",
					utils.FilenameCfg: "Resources.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "TTL",
							utils.PathCfg:  "UsageTTL",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "Limit",
							utils.PathCfg:  "Limit",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "AllocationMessage",
							utils.PathCfg:  "AllocationMessage",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "Blocker",
							utils.PathCfg:  "Blocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "Stored",
							utils.PathCfg:  "Stored",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
					},
				},
				{
					utils.TypeCfg:     "*ips",
					utils.FilenameCfg: "IPs.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "TTL",
							utils.PathCfg:  "TTL",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "Type",
							utils.PathCfg:  "Type",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "AddressPool",
							utils.PathCfg:  "AddressPool",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "Allocation",
							utils.PathCfg:  "Allocation",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "Stored",
							utils.PathCfg:  "Stored",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
					},
				},
				{
					utils.TypeCfg:     "*stats",
					utils.FilenameCfg: "Stats.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Blockers",
							utils.PathCfg:  "Blockers",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "QueueLength",
							utils.PathCfg:  "QueueLength",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "TTL",
							utils.PathCfg:  "TTL",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "MinItems",
							utils.PathCfg:  "MinItems",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "Stored",
							utils.PathCfg:  "Stored",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.NewBranchCfg: true,
							utils.TagCfg:       "MetricIDs",
							utils.PathCfg:      "Metrics.MetricID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.10",
						},
						{
							utils.TagCfg:   "MetricFilterIDs",
							utils.PathCfg:  "Metrics.FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
						{
							utils.TagCfg:   "MetricBlockers",
							utils.PathCfg:  "Metrics.Blockers",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.12",
						},
					},
				},
				{
					utils.TypeCfg:     "*thresholds",
					utils.FilenameCfg: "Thresholds.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "MaxHits",
							utils.PathCfg:  "MaxHits",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "MinHits",
							utils.PathCfg:  "MinHits",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "MinSleep",
							utils.PathCfg:  "MinSleep",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "Blocker",
							utils.PathCfg:  "Blocker",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "ActionProfileIDs",
							utils.PathCfg:  "ActionProfileIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "Async",
							utils.PathCfg:  "Async",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
					},
				},
				{
					utils.TypeCfg:     "*trends",
					utils.FilenameCfg: "Trends.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "Schedule",
							utils.PathCfg:  "Schedule",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "StatID",
							utils.PathCfg:  "StatID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Metrics",
							utils.PathCfg:  "Metrics",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "TTL",
							utils.PathCfg:  "TTL",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "QueueLength",
							utils.PathCfg:  "QueueLength",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "MinItems",
							utils.PathCfg:  "MinItems",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "CorrelationType",
							utils.PathCfg:  "CorrelationType",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "Tolerance",
							utils.PathCfg:  "Tolerance",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "Stored",
							utils.PathCfg:  "Stored",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						}, {
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
					},
				},
				{
					utils.TypeCfg:     "*rankings",
					utils.FilenameCfg: "Rankings.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "Schedule",
							utils.PathCfg:  "Schedule",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "StatIDs",
							utils.PathCfg:  "StatIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "MetricIDs",
							utils.PathCfg:  "MetricIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "Sorting",
							utils.PathCfg:  "Sorting",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "SortingParameters",
							utils.PathCfg:  "SortingParameters",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:   "Stored",
							utils.PathCfg:  "Stored",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
					},
				},
				{
					utils.TypeCfg:     "*routes",
					utils.FilenameCfg: "Routes.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Blockers",
							utils.PathCfg:  "Blockers",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "Sorting",
							utils.PathCfg:  "Sorting",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "SortingParameters",
							utils.PathCfg:  "SortingParameters",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.NewBranchCfg: true,
							utils.TagCfg:       "RouteID",
							utils.PathCfg:      "Routes.ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.7",
						},
						{
							utils.TagCfg:   "RouteFilterIDs",
							utils.PathCfg:  "Routes.FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.8",
						},
						{
							utils.TagCfg:   "RouteAccountIDs",
							utils.PathCfg:  "Routes.AccountIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.9",
						},
						{
							utils.TagCfg:   "RouteRateProfileIDs",
							utils.PathCfg:  "Routes.RateProfileIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.10",
						},
						{
							utils.TagCfg:   "RouteResourceIDs",
							utils.PathCfg:  "Routes.ResourceIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.11",
						},
						{
							utils.TagCfg:   "RouteStatIDs",
							utils.PathCfg:  "Routes.StatIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.12",
						},
						{
							utils.TagCfg:   "RouteWeights",
							utils.PathCfg:  "Routes.Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.13",
						},
						{
							utils.TagCfg:   "RouteBlockers",
							utils.PathCfg:  "Routes.Blockers",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.14",
						},
						{
							utils.TagCfg:   "RouteParameters",
							utils.PathCfg:  "Routes.RouteParameters",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.15",
						},
					},
				},
				{
					utils.TypeCfg:     "*chargers",
					utils.FilenameCfg: "Chargers.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Blockers",
							utils.PathCfg:  "Blockers",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "RunID",
							utils.PathCfg:  "RunID",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "AttributeIDs",
							utils.PathCfg:  "AttributeIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
					},
				},
				{
					utils.TypeCfg:     "*rate_profiles",
					utils.FilenameCfg: "Rates.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "MinCost",
							utils.PathCfg:  "MinCost",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "MaxCost",
							utils.PathCfg:  "MaxCost",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "MaxCostStrategy",
							utils.PathCfg:  "MaxCostStrategy",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.6",
						},
						{
							utils.TagCfg:     "RateFilterIDs",
							utils.PathCfg:    "Rates[<~*req.7>].FilterIDs",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.8",
						},
						{
							utils.TagCfg:     "RateActivationTimes",
							utils.PathCfg:    "Rates[<~*req.7>].ActivationTimes",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.9",
						},
						{
							utils.TagCfg:     "RateWeights",
							utils.PathCfg:    "Rates[<~*req.7>].Weights",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.10",
						},
						{
							utils.TagCfg:     "RateBlocker",
							utils.PathCfg:    "Rates[<~*req.7>].Blocker",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.11",
						},
						{
							utils.NewBranchCfg: true,
							utils.TagCfg:       "RateIntervalStart",
							utils.PathCfg:      "Rates[<~*req.7>].IntervalRates.IntervalStart",
							utils.TypeCfg:      "*variable",
							utils.FiltersCfg:   []string{"*notempty:~*req.7:"},
							utils.ValueCfg:     "~*req.12",
						},
						{
							utils.TagCfg:     "RateFixedFee",
							utils.PathCfg:    "Rates[<~*req.7>].IntervalRates.FixedFee",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.13",
						},
						{
							utils.TagCfg:     "RateRecurrentFee",
							utils.PathCfg:    "Rates[<~*req.7>].IntervalRates.RecurrentFee",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.14",
						},
						{
							utils.TagCfg:     "RateUnit",
							utils.PathCfg:    "Rates[<~*req.7>].IntervalRates.Unit",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.15",
						},
						{
							utils.TagCfg:     "RateIncrement",
							utils.PathCfg:    "Rates[<~*req.7>].IntervalRates.Increment",
							utils.TypeCfg:    "*variable",
							utils.FiltersCfg: []string{"*notempty:~*req.7:"},
							utils.ValueCfg:   "~*req.16",
						},
					},
				},
				{
					utils.TypeCfg:     "*action_profiles",
					utils.FilenameCfg: "Actions.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Blockers",
							utils.PathCfg:  "Blockers",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "Schedule",
							utils.PathCfg:  "Schedule",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:   "TargetIDs",
							utils.PathCfg:  "Targets[<~*req.6>]",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.7",
						},
						{
							utils.TagCfg:     "ActionFilterIDs",
							utils.PathCfg:    "Actions[<~*req.8>].FilterIDs",
							utils.FiltersCfg: []string{"*notempty:~*req.8:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.9",
						},
						{
							utils.TagCfg:     "ActionTTL",
							utils.PathCfg:    "Actions[<~*req.8>].TTL",
							utils.FiltersCfg: []string{"*notempty:~*req.8:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.10",
						},
						{
							utils.TagCfg:     "ActionType",
							utils.PathCfg:    "Actions[<~*req.8>].Type",
							utils.FiltersCfg: []string{"*notempty:~*req.8:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.11",
						},
						{
							utils.TagCfg:     "ActionOpts",
							utils.PathCfg:    "Actions[<~*req.8>].Opts",
							utils.FiltersCfg: []string{"*notempty:~*req.8:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.12",
						},
						{
							utils.NewBranchCfg: true,
							utils.TagCfg:       "ActionPath",
							utils.PathCfg:      "Actions[<~*req.8>].Diktats.Path",
							utils.FiltersCfg:   []string{"*notempty:~*req.8:"},
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.13",
						},
						{
							utils.TagCfg:     "ActionValue",
							utils.PathCfg:    "Actions[<~*req.8>].Diktats.Value",
							utils.FiltersCfg: []string{"*notempty:~*req.8:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.14",
						},
					},
				},
				{
					utils.TypeCfg:     "*accounts",
					utils.FilenameCfg: "Accounts.csv",
					utils.FlagsCfg:    flags,
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "Tenant",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:       "ID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
						{
							utils.TagCfg:   "FilterIDs",
							utils.PathCfg:  "FilterIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.2",
						},
						{
							utils.TagCfg:   "Weights",
							utils.PathCfg:  "Weights",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.3",
						},
						{
							utils.TagCfg:   "Blockers",
							utils.PathCfg:  "Blockers",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.4",
						},
						{
							utils.TagCfg:   "Opts",
							utils.PathCfg:  "Opts",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.5",
						},
						{
							utils.TagCfg:     "BalanceFilterIDs",
							utils.PathCfg:    "Balances[<~*req.6>].FilterIDs",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.7",
						},
						{
							utils.TagCfg:     "BalanceWeights",
							utils.PathCfg:    "Balances[<~*req.6>].Weights",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.8",
						},
						{
							utils.TagCfg:     "BalanceBlockers",
							utils.PathCfg:    "Balances[<~*req.6>].Blockers",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.9",
						},
						{
							utils.TagCfg:     "BalanceType",
							utils.PathCfg:    "Balances[<~*req.6>].Type",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.10",
						},
						{
							utils.TagCfg:     "BalanceUnits",
							utils.PathCfg:    "Balances[<~*req.6>].Units",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.11",
						},
						{
							utils.TagCfg:     "BalanceUnitFactors",
							utils.PathCfg:    "Balances[<~*req.6>].UnitFactors",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.12",
						},
						{
							utils.TagCfg:     "BalanceOpts",
							utils.PathCfg:    "Balances[<~*req.6>].Opts",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.13",
						},
						{
							utils.TagCfg:     "BalanceCostIncrements",
							utils.PathCfg:    "Balances[<~*req.6>].CostIncrements",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.14",
						},
						{
							utils.TagCfg:     "BalanceAttributeIDs",
							utils.PathCfg:    "Balances[<~*req.6>].AttributeIDs",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.15",
						},
						{
							utils.TagCfg:     "BalanceRateProfileIDs",
							utils.PathCfg:    "Balances[<~*req.6>].RateProfileIDs",
							utils.FiltersCfg: []string{"*notempty:~*req.6:"},
							utils.TypeCfg:    "*variable",
							utils.ValueCfg:   "~*req.16",
						},
						{
							utils.TagCfg:   "ThresholdIDs",
							utils.PathCfg:  "ThresholdIDs",
							utils.TypeCfg:  "*variable",
							utils.ValueCfg: "~*req.17",
						},
					},
				},
			},
			utils.CacheCfg: map[string]any{
				utils.MetaFilters:        map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaAttributes:     map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaResources:      map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaIPs:            map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaStats:          map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaThresholds:     map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaRoutes:         map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaChargers:       map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaRateProfiles:   map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaActionProfiles: map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaAccounts:       map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaRankings:       map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
				utils.MetaTrends:         map[string]any{utils.LimitCfg: -1, utils.TTLCfg: "5s", utils.PrecacheCfg: false, utils.RemoteCfg: false, utils.ReplicateCfg: false, utils.StaticTTLCfg: false},
			},
		},
	}
	if cfgCgr, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else {
		rcv := cfgCgr.loaderCfg.AsMapInterface()
		if len(cfgCgr.loaderCfg) != 1 {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", 1, len(cfgCgr.loaderCfg))
		} else if !reflect.DeepEqual(rcv, eMap) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(eMap), utils.ToJSON(rcv))
		}
	}
}

func TestLoaderCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
			"loaders": [												
	{
		"id": "*default",									
		"enabled": false,									
		"tenant": "~*req.Destination1",										
		"run_delay": "1",										
		"lockfile_path": ".cgr.lck",						
		"caches_conns": ["*conn1"],
		"field_separator": ",",								
		"tp_in_dir": "/var/spool/cgrates/loader/in",		
		"tp_out_dir": "/var/spool/cgrates/loader/out",		
		"data":[											
			{
				"type": "*attributes",						
				"file_name": "Attributes.csv",				
				"fields": [
					{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~req.0", "mandatory": true},
					{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
					],
				},
			],
		},
	],	
}`
	eMap := []map[string]any{
		{
			utils.IDCfg:           "*default",
			utils.EnabledCfg:      false,
			utils.TenantCfg:       "~*req.Destination1",
			utils.RunDelayCfg:     "0",
			utils.LockFilePathCfg: ".cgr.lck",
			utils.CachesConnsCfg:  []string{"*conn1"},
			utils.FieldSepCfg:     ",",
			utils.TpInDirCfg:      "/var/spool/cgrates/loader/in",
			utils.TpOutDirCfg:     "/var/spool/cgrates/loader/out",
			utils.DataCfg: []map[string]any{
				{
					utils.TypeCfg:     "*attributes",
					utils.FilenameCfg: "Attributes.csv",
					utils.FieldsCfg: []map[string]any{
						{
							utils.TagCfg:       "TenantID",
							utils.PathCfg:      "Tenant",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.0",
							utils.MandatoryCfg: true,
						}, {
							utils.TagCfg:       "ProfileID",
							utils.PathCfg:      "ID",
							utils.TypeCfg:      "*variable",
							utils.ValueCfg:     "~*req.1",
							utils.MandatoryCfg: true,
						},
					},
				},
			},
		},
	}
	if jsonCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := jsonCfg.loaderCfg.AsMapInterface().([]map[string]any); !reflect.DeepEqual(rcv[0][utils.Tenant], eMap[0][utils.Tenant]) {
		t.Errorf("Expected %+v, received %+v", rcv[0][utils.Tenant], eMap[0][utils.Tenant])
	}
}

func TestLoaderSCfgsClone(t *testing.T) {
	ban := LoaderSCfgs{{
		Enabled:        true,
		ID:             utils.MetaDefault,
		Tenant:         "cgrates.org",
		LockFilePath:   ".cgr.lck",
		CacheSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), "*conn1"},
		FieldSeparator: ",",
		TpInDir:        "/var/spool/cgrates/loader/in",
		TpOutDir:       "/var/spool/cgrates/loader/out",
		Data: []*LoaderDataType{{
			Type:     "*attributes",
			Filename: "Attributes.csv",
			Flags:    utils.FlagsWithParams{},
			Fields: []*FCTemplate{
				{
					Tag:       "TenantID",
					Path:      "Tenant",
					pathSlice: []string{"Tenant"},
					Type:      utils.MetaComposed,
					Value:     utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep),
					Mandatory: true,
					Layout:    time.RFC3339,
				},
			}},
		},
		Opts: &LoaderSOptsCfg{},
		Cache: map[string]*CacheParamCfg{
			utils.MetaFilters:        {Limit: -1, TTL: 5 * time.Second},
			utils.MetaAttributes:     {Limit: -1, TTL: 5 * time.Second},
			utils.MetaResources:      {Limit: -1, TTL: 5 * time.Second},
			utils.MetaStats:          {Limit: -1, TTL: 5 * time.Second},
			utils.MetaThresholds:     {Limit: -1, TTL: 5 * time.Second},
			utils.MetaRoutes:         {Limit: -1, TTL: 5 * time.Second},
			utils.MetaChargers:       {Limit: -1, TTL: 5 * time.Second},
			utils.MetaRateProfiles:   {Limit: -1, TTL: 5 * time.Second},
			utils.MetaActionProfiles: {Limit: -1, TTL: 5 * time.Second},
			utils.MetaAccounts:       {Limit: -1, TTL: 5 * time.Second},
		},
	}}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, *rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(*rcv))
	}
	if (*rcv)[0].CacheSConns[1] = ""; ban[0].CacheSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv)[0].Data[0].Type = ""; ban[0].Data[0].Type != "*attributes" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestEqualsLoaderDatasType(t *testing.T) {
	v1 := []*LoaderDataType{
		{
			Type:     "*json",
			Filename: "file.json",
			Flags: utils.FlagsWithParams{
				"FLAG_1": {
					"PARAM_1": []string{"param1"},
				},
			},
			Fields: []*FCTemplate{
				{
					Type: "Type",
					Tag:  "Tag",
				},
			},
		},
	}

	v2 := []*LoaderDataType{
		{
			Type:     "*xml",
			Filename: "file.xml",
			Flags: utils.FlagsWithParams{
				"FLAG_2": {
					"PARAM_2": []string{"param2"},
				},
			},
			Fields: []*FCTemplate{
				{
					Type: "Type2",
					Tag:  "Tag2",
				},
			},
		},
	}

	if equalsLoaderDatasType(v1, v2) {
		t.Error("Loaders should not match")
	}

	v1 = v2
	if !equalsLoaderDatasType(v1, v2) {
		t.Error("Loaders should match")
	}

	v2 = []*LoaderDataType{}
	if equalsLoaderDatasType(v1, v2) {
		t.Error("Loaders should not match")
	}
}

func TestDiffLoaderJsonCfg(t *testing.T) {

	v1 := &LoaderSCfg{
		ID:             "LoaderID",
		Enabled:        true,
		Tenant:         "cgrates.org",
		RunDelay:       1 * time.Millisecond,
		LockFilePath:   "lockFileName",
		CacheSConns:    []string{"*localhost"},
		FieldSeparator: ";",
		TpInDir:        "/tp/in/dir",
		TpOutDir:       "/tp/out/dir",
		Data:           nil,
		Opts:           &LoaderSOptsCfg{},
	}

	v2 := &LoaderSCfg{
		ID:             "LoaderID2",
		Enabled:        false,
		Tenant:         "itsyscom.com",
		RunDelay:       2 * time.Millisecond,
		LockFilePath:   "lockFileName2",
		CacheSConns:    []string{"*birpc"},
		FieldSeparator: ":",
		TpInDir:        "/tp/in/dir/2",
		TpOutDir:       "/tp/out/dir/2",
		Data: []*LoaderDataType{
			{
				Type:     "*xml",
				Filename: "file.xml",
				Flags: utils.FlagsWithParams{
					"FLAG_2": {
						"PARAM_2": []string{"param2"},
					},
				},
				Fields: []*FCTemplate{
					{
						Type: "Type2",
						Tag:  "Tag2",
					},
				},
			},
		},
		Opts: &LoaderSOptsCfg{},
	}

	expected := &LoaderJsonCfg{
		ID:              utils.StringPointer("LoaderID2"),
		Enabled:         utils.BoolPointer(false),
		Tenant:          utils.StringPointer("itsyscom.com"),
		Run_delay:       utils.StringPointer("2ms"),
		Lockfile_path:   utils.StringPointer("lockFileName2"),
		Caches_conns:    &[]string{"*birpc"},
		Field_separator: utils.StringPointer(":"),
		Tp_in_dir:       utils.StringPointer("/tp/in/dir/2"),
		Tp_out_dir:      utils.StringPointer("/tp/out/dir/2"),
		Data: &[]*LoaderJsonDataType{
			{
				Id:        utils.StringPointer(""),
				Type:      utils.StringPointer("*xml"),
				File_name: utils.StringPointer("file.xml"),
				Flags:     &[]string{"FLAG_2:PARAM_2:param2"},
				Fields: &[]*FcTemplateJsonCfg{
					{
						Type:   utils.StringPointer("Type2"),
						Tag:    utils.StringPointer("Tag2"),
						Layout: utils.StringPointer(""),
					},
				},
			},
		},
		Opts:  &LoaderJsonOptsCfg{},
		Cache: map[string]*CacheParamJsonCfg{},
	}

	rcv := diffLoaderJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &LoaderJsonCfg{Opts: &LoaderJsonOptsCfg{}, Cache: make(map[string]*CacheParamJsonCfg)}
	rcv = diffLoaderJsonCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

}

func TestEqualsLoadersJsonCfg(t *testing.T) {
	v1 := LoaderSCfgs{
		{
			ID:             "LoaderID",
			Enabled:        true,
			Tenant:         "cgrates.org",
			RunDelay:       1 * time.Millisecond,
			LockFilePath:   "lockFileName",
			CacheSConns:    []string{"*localhost"},
			FieldSeparator: ";",
			TpInDir:        "/tp/in/dir",
			TpOutDir:       "/tp/out/dir",
			Data:           nil,
			Opts:           &LoaderSOptsCfg{},
		},
	}

	v2 := LoaderSCfgs{
		{
			ID:             "LoaderID2",
			Enabled:        false,
			Tenant:         "cgrates.org",
			RunDelay:       2 * time.Millisecond,
			LockFilePath:   "lockFileName2",
			CacheSConns:    []string{"*birpc"},
			FieldSeparator: ":",
			TpInDir:        "/tp/in/dir/2",
			TpOutDir:       "/tp/out/dir/2",
			Data: []*LoaderDataType{
				{
					Type:     "*xml",
					Filename: "file.xml",
					Flags: utils.FlagsWithParams{
						"FLAG_2": {
							"PARAM_2": []string{"param2"},
						},
					},
					Fields: []*FCTemplate{
						{
							Type: "Type2",
							Tag:  "Tag2",
						},
					},
				},
			},
			Opts: &LoaderSOptsCfg{},
		},
	}

	if equalsLoadersJsonCfg(v1, v2) {
		t.Error("Loaders shouldn't match")
	}

	v2 = v1
	if !equalsLoadersJsonCfg(v1, v2) {
		t.Error("Loaders shouldn't match")
	}

	v2 = LoaderSCfgs{}
	if equalsLoadersJsonCfg(v1, v2) {
		t.Error("Loaders shouldn't match")
	}
}

func TestDiffLoadersJsonCfg(t *testing.T) {
	var d []*LoaderJsonCfg

	v1 := LoaderSCfgs{
		{
			ID:             "LoaderID",
			Enabled:        false,
			Tenant:         "cgrates.org",
			RunDelay:       1 * time.Millisecond,
			LockFilePath:   "lockFileName",
			CacheSConns:    []string{"*localhost"},
			FieldSeparator: ";",
			TpInDir:        "/tp/in/dir",
			TpOutDir:       "/tp/out/dir",
			Data:           nil,
			Opts:           &LoaderSOptsCfg{},
		},
	}

	v2 := LoaderSCfgs{
		{
			ID:             "LoaderID2",
			Enabled:        true,
			Tenant:         "itsyscom.com",
			RunDelay:       2 * time.Millisecond,
			LockFilePath:   "lockFileName2",
			CacheSConns:    []string{"*birpc"},
			FieldSeparator: ":",
			TpInDir:        "/tp/in/dir/2",
			TpOutDir:       "/tp/out/dir/2",
			Data: []*LoaderDataType{
				{
					Type:     "*xml",
					Filename: "file.xml",
					Flags: utils.FlagsWithParams{
						"FLAG_2": {
							"PARAM_2": []string{"param2"},
						},
					},
					Fields: []*FCTemplate{
						{
							Type: "Type2",
							Tag:  "Tag2",
						},
					},
				},
			},
			Opts: &LoaderSOptsCfg{},
		},
	}

	expected := []*LoaderJsonCfg{
		{
			ID:              utils.StringPointer("LoaderID2"),
			Enabled:         utils.BoolPointer(true),
			Tenant:          utils.StringPointer("itsyscom.com"),
			Run_delay:       utils.StringPointer("2ms"),
			Lockfile_path:   utils.StringPointer("lockFileName2"),
			Caches_conns:    &[]string{"*birpc"},
			Field_separator: utils.StringPointer(":"),
			Tp_in_dir:       utils.StringPointer("/tp/in/dir/2"),
			Tp_out_dir:      utils.StringPointer("/tp/out/dir/2"),
			Data: &[]*LoaderJsonDataType{
				{
					Id:        utils.StringPointer(""),
					Type:      utils.StringPointer("*xml"),
					File_name: utils.StringPointer("file.xml"),
					Flags:     &[]string{"FLAG_2:PARAM_2:param2"},
					Fields: &[]*FcTemplateJsonCfg{
						{
							Type:   utils.StringPointer("Type2"),
							Tag:    utils.StringPointer("Tag2"),
							Layout: utils.StringPointer(""),
						},
					},
				},
			},
			Action: utils.StringPointer(""),
			Opts:   &LoaderJsonOptsCfg{WithIndex: utils.BoolPointer(false)},
			Cache:  map[string]*CacheParamJsonCfg{},
		},
	}

	rcv := diffLoadersJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = nil
	rcv = diffLoadersJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestLockFolderRelativePath(t *testing.T) {
	ldr := &LoaderSCfg{
		TpInDir:      "/var/spool/cgrates/loader/in/",
		TpOutDir:     "/var/spool/cgrates/loader/out/",
		LockFilePath: utils.ResourcesCsv,
	}

	jsonCfg := &LoaderJsonCfg{
		ID:              utils.StringPointer("loaderid"),
		Enabled:         utils.BoolPointer(true),
		Tenant:          utils.StringPointer("cgrates.org"),
		Lockfile_path:   utils.StringPointer(utils.ResourcesCsv),
		Field_separator: utils.StringPointer(utils.InfieldSep),
		Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in/"),
		Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out/"),
	}
	expPath := path.Join(ldr.LockFilePath)
	if err := ldr.loadFromJSONCfg(jsonCfg, map[string][]*FCTemplate{}); err != nil {
		t.Error(err)
	} else if ldr.LockFilePath != expPath {
		t.Errorf("Expected %v \n but received \n %v", expPath, ldr.LockFilePath)
	}
}
func TestLockFolderNonRelativePath(t *testing.T) {
	ldr := &LoaderSCfg{
		TpInDir:      "/var/spool/cgrates/loader/in/",
		TpOutDir:     "/var/spool/cgrates/loader/out/",
		LockFilePath: utils.ResourcesCsv,
	}

	jsonCfg := &LoaderJsonCfg{
		ID:              utils.StringPointer("loaderid"),
		Enabled:         utils.BoolPointer(true),
		Tenant:          utils.StringPointer("cgrates.org"),
		Lockfile_path:   utils.StringPointer(path.Join("/tmp/", utils.ResourcesCsv)),
		Field_separator: utils.StringPointer(utils.InfieldSep),
		Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in/"),
		Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out/"),
	}
	expPath := path.Join("/tmp/", utils.ResourcesCsv)
	if err := ldr.loadFromJSONCfg(jsonCfg, map[string][]*FCTemplate{}); err != nil {
		t.Error(err)
	} else if ldr.LockFilePath != expPath {
		t.Errorf("Expected %v \n but received \n %v", expPath, ldr.LockFilePath)
	}
}

func TestLockFolderIsDir(t *testing.T) {
	ldr := &LoaderSCfg{
		LockFilePath: "test",
	}

	jsonCfg := &LoaderJsonCfg{
		ID:              utils.StringPointer("loaderid"),
		Enabled:         utils.BoolPointer(true),
		Tenant:          utils.StringPointer("cgrates.org"),
		Lockfile_path:   utils.StringPointer("/tmp"),
		Field_separator: utils.StringPointer(utils.InfieldSep),
		Tp_in_dir:       utils.StringPointer("/var/spool/cgrates/loader/in/"),
		Tp_out_dir:      utils.StringPointer("/var/spool/cgrates/loader/out/"),
	}
	expPath := path.Join("/tmp")

	if err := ldr.loadFromJSONCfg(jsonCfg, map[string][]*FCTemplate{}); err != nil {
		t.Error(err)
	} else if ldr.LockFilePath != expPath {
		t.Errorf("Expected %v \n but received \n %v", expPath, ldr.LockFilePath)
	}
}

func TestLoaderSCloneSection(t *testing.T) {
	ldrsCfg := LoaderSCfgs{
		{
			ID:             "LoaderID",
			Enabled:        false,
			Tenant:         "cgrates.org",
			RunDelay:       1 * time.Millisecond,
			LockFilePath:   "lockFileName",
			CacheSConns:    []string{"*localhost"},
			FieldSeparator: ";",
			TpInDir:        "/tp/in/dir",
			TpOutDir:       "/tp/out/dir",
			Data: []*LoaderDataType{
				{
					Type:     "*csv",
					Filename: "test",
					Flags: utils.FlagsWithParams{
						"k1": map[string][]string{
							"k2": {"f1", "f2"},
						},
					},
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path: utils.MetaRep + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
							Value: utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
					},
				},
			},
			Opts: &LoaderSOptsCfg{},
		},
	}

	exp := LoaderSCfgs{
		{
			ID:             "LoaderID",
			Enabled:        false,
			Tenant:         "cgrates.org",
			RunDelay:       1 * time.Millisecond,
			LockFilePath:   "lockFileName",
			CacheSConns:    []string{"*localhost"},
			FieldSeparator: ";",
			TpInDir:        "/tp/in/dir",
			TpOutDir:       "/tp/out/dir",
			Data: []*LoaderDataType{
				{
					Type:     "*csv",
					Filename: "test",
					Flags: utils.FlagsWithParams{
						"k1": map[string][]string{
							"k2": {"f1", "f2"},
						},
					},
					Fields: []*FCTemplate{
						{Tag: "Tenant",
							Path: utils.MetaRep + utils.NestingSep + utils.Tenant, Type: utils.MetaVariable,
							Value: utils.NewRSRParsersMustCompile("cgrates.org", utils.InfieldSep)},
					},
				},
			},
			Opts:  &LoaderSOptsCfg{},
			Cache: make(map[string]*CacheParamCfg),
		},
	}

	rcv := ldrsCfg.CloneSection()
	if !reflect.DeepEqual((*rcv.(*LoaderSCfgs))[0], exp[0]) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp[0]), utils.ToJSON((*rcv.(*LoaderSCfgs))[0]))
	}
}
func TestLoaderDataTypeLoadFromJSONId(t *testing.T) {
	lData := &LoaderDataType{}
	jsnCfg := &LoaderJsonDataType{
		Id: utils.StringPointer("IdTest"),
	}

	if err := lData.loadFromJSONCfg(jsnCfg, nil); err != nil {
		t.Errorf("Expected error <%v>, Received error <%v>", nil, err)
	}
}

func TestLoaderSCfgloadFromJSONCfgDataId(t *testing.T) {
	l := &LoaderSCfg{}
	jsnCfg := &LoaderJsonCfg{
		Data: &[]*LoaderJsonDataType{
			{
				Id: utils.StringPointer(""),
			},
		},
	}
	if err := l.loadFromJSONCfg(jsnCfg, nil); err != nil {

		t.Errorf("Expected error <%v>, Received error <%v>", nil, err)
	}
}
func TestLoaderSCfgloadFromJSONCacheErr(t *testing.T) {
	l := &LoaderSCfg{}
	jsnCfg := &LoaderJsonCfg{
		Cache: map[string]*CacheParamJsonCfg{
			"": {
				Ttl: utils.StringPointer("invalid"),
			},
		},
	}

	expErr := `time: invalid duration "invalid"`
	if err := l.loadFromJSONCfg(jsnCfg, nil); err.Error() != expErr {
		t.Errorf("Expected error <%v>, Received error <%v>", expErr, err.Error())
	}
}

func TestDiffLoaderJsonOptsCfg(t *testing.T) {

	v1 := &LoaderSOptsCfg{
		Cache:       "cacheFiller",
		WithIndex:   true,
		ForceLock:   true,
		StopOnError: true,
	}

	v2 := &LoaderSOptsCfg{
		Cache:       "cacheFillerV2",
		WithIndex:   false,
		ForceLock:   false,
		StopOnError: false,
	}

	expected := &LoaderJsonOptsCfg{
		Cache:       utils.StringPointer("cacheFillerV2"),
		WithIndex:   utils.BoolPointer(false),
		ForceLock:   utils.BoolPointer(false),
		StopOnError: utils.BoolPointer(false),
	}

	rcv := diffLoaderJsonOptsCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &LoaderJsonOptsCfg{}
	rcv = diffLoaderJsonOptsCfg(v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
