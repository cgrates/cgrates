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

package tpes

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"io"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var itemIDsByExport = map[string][]string{
	utils.MetaAccounts:   {"acc1", "acc2"},
	utils.MetaActions:    {"ap1", "ap2"},
	utils.MetaRates:      {"rp1", "rp2"},
	utils.MetaFilters:    {"fltr1", "fltr2"},
	utils.MetaAttributes: {"attr1", "attr2"},
	utils.MetaResources:  {"res1", "res2"},
	utils.MetaStats:      {"stat1", "stat2"},
	utils.MetaThresholds: {"th1", "th2"},
	utils.MetaTrends:     {"trend1", "trend2"},
	utils.MetaRankings:   {"rank1", "rank2"},
	utils.MetaRoutes:     {"routePrf1", "routePrf2"},
	utils.MetaChargers:   {"charger1", "charger2"},
}

func TestCSVImportExport(t *testing.T) {
	files := map[string][][]string{
		utils.AccountsCsv: {
			accountHeader,
			csvRow(t, accountHeader, map[string]string{
				utils.Tenant:                utils.CGRateSorg,
				utils.ID:                    "acc1",
				utils.FilterIDs:             "*string:~*req.Account:acc1;*prefix:~*req.Destination:+1",
				utils.Weights:               "accWeightFltr;10",
				utils.Blockers:              "accBlockFltr;false",
				utils.Opts:                  "accountOpt:accountVal;zOpt:zVal",
				utils.BalanceID:             "bal1",
				utils.BalanceFilterIDs:      "*string:~*req.Balance:bal1;*exists:~*opts.*usage:",
				utils.BalanceWeights:        "balWeightFltr;20",
				utils.BalanceBlockers:       "balBlockFltr;true",
				utils.BalanceType:           utils.MetaMonetary,
				utils.BalanceUnits:          "10",
				utils.BalanceUnitFactors:    "uf1&uf2;1.5;uf3;2",
				utils.BalanceOpts:           "balanceOpt:balanceVal;zOpt:zVal",
				utils.BalanceCostIncrements: "ci1&ci2;1;0.1;0.01;ci3;2;0.2;0.02",
				utils.BalanceAttributeIDs:   "attr1;attr2",
				utils.BalanceRateProfileIDs: "rp1;rp2",
				utils.ThresholdIDs:          "th1;th2",
			}),
			csvRow(t, accountHeader, map[string]string{
				utils.Tenant:              utils.CGRateSorg,
				utils.ID:                  "acc1",
				utils.BalanceID:           "bal2",
				utils.BalanceFilterIDs:    "*string:~*req.Balance:bal2",
				utils.BalanceWeights:      "bal2WeightFltr;30",
				utils.BalanceBlockers:     "bal2BlockFltr;false",
				utils.BalanceType:         utils.MetaVoice,
				utils.BalanceUnits:        "20",
				utils.BalanceOpts:         "voiceOpt:voiceVal",
				utils.BalanceAttributeIDs: "attr3",
			}),
			csvRow(t, accountHeader, map[string]string{
				utils.Tenant:                utils.CGRateSorg,
				utils.ID:                    "acc2",
				utils.FilterIDs:             "*string:~*req.Account:acc2",
				utils.Weights:               "acc2WeightFltr;15",
				utils.Blockers:              "acc2BlockFltr;true",
				utils.Opts:                  "acc2Opt:acc2Val",
				utils.BalanceID:             "dataBal",
				utils.BalanceFilterIDs:      "*string:~*req.Balance:data",
				utils.BalanceWeights:        "dataWeightFltr;25",
				utils.BalanceBlockers:       "dataBlockFltr;false",
				utils.BalanceType:           utils.MetaData,
				utils.BalanceUnits:          "1024",
				utils.BalanceUnitFactors:    "dataUf;0.5",
				utils.BalanceOpts:           "dataOpt:dataVal",
				utils.BalanceRateProfileIDs: "rp2",
				utils.ThresholdIDs:          "th3",
			}),
		},
		utils.ActionsCsv: {
			actionHeader,
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:     utils.CGRateSorg,
				utils.ID:         "ap1",
				utils.FilterIDs:  "*string:~*req.Account:1001;*prefix:~*req.Destination:+1",
				utils.Weights:    "apWeightFltr;10",
				utils.Blockers:   "apBlockFltr;true",
				utils.Schedule:   utils.MetaASAP,
				utils.TargetType: utils.MetaAccounts,
				utils.TargetIDs:  "acc1;acc2",
			}),
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:     utils.CGRateSorg,
				utils.ID:         "ap1",
				utils.TargetType: utils.MetaResources,
				utils.TargetIDs:  "res1",
			}),
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:                 utils.CGRateSorg,
				utils.ID:                     "ap1",
				utils.ActionID:               "act1",
				utils.ActionFilterIDs:        "*string:~*req.Action:act1;*exists:~*opts.*usage:",
				utils.ActionTTL:              "1s",
				utils.ActionType:             utils.MetaLog,
				utils.ActionOpts:             "msg:hello;scope:test",
				utils.ActionWeights:          "actWeightFltr;20",
				utils.ActionBlockers:         "actBlockFltr;true",
				utils.ActionDiktatsID:        "d1",
				utils.ActionDiktatsFilterIDs: "*string:~*req.Diktat:d1",
				utils.ActionDiktatsOpts:      "k1:v1;k2:v2",
				utils.ActionDiktatsWeights:   "dweight;30",
				utils.ActionDiktatsBlockers:  "dblock;false",
			}),
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:          utils.CGRateSorg,
				utils.ID:              "ap1",
				utils.ActionID:        "act2",
				utils.ActionTTL:       "2s",
				utils.ActionType:      utils.MetaLog,
				utils.ActionDiktatsID: "empty1",
			}),
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:          utils.CGRateSorg,
				utils.ID:              "ap1",
				utils.ActionID:        "act2",
				utils.ActionDiktatsID: "empty2",
			}),
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:     utils.CGRateSorg,
				utils.ID:         "ap2",
				utils.Schedule:   utils.MetaASAP,
				utils.TargetType: utils.MetaAccounts,
				utils.TargetIDs:  "acc3",
			}),
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:     utils.CGRateSorg,
				utils.ID:         "ap2",
				utils.ActionID:   "act3",
				utils.ActionTTL:  "3s",
				utils.ActionType: utils.MetaLog,
				utils.ActionOpts: "note:noDiktat",
			}),
			csvRow(t, actionHeader, map[string]string{
				utils.Tenant:     utils.CGRateSorg,
				utils.ID:         "ap2",
				utils.ActionID:   "act4",
				utils.ActionType: utils.MetaLog,
			}),
		},
		utils.RatesCsv: {
			rateHeader,
			csvRow(t, rateHeader, map[string]string{
				utils.Tenant:              utils.CGRateSorg,
				utils.ID:                  "rp1",
				utils.FilterIDs:           "*string:~*req.Account:1001;*prefix:~*req.Destination:+1",
				utils.Weights:             "rpWeightFltr;10",
				utils.MinCost:             "1",
				utils.MaxCost:             "10",
				utils.MaxCostStrategy:     utils.MetaMaxCostFree,
				utils.RateID:              "rt1",
				utils.RateFilterIDs:       "*string:~*req.Subject:1001;*exists:~*opts.*usage:",
				utils.RateActivationStart: "* * * * *",
				utils.RateWeights:         "rtWeightFltr;20",
				utils.RateBlocker:         "true",
				utils.RateIntervalStart:   "0",
				utils.RateFixedFee:        "0.1",
				utils.RateRecurrentFee:    "0.2",
				utils.RateUnit:            "60",
				utils.RateIncrement:       "1",
			}),
			csvRow(t, rateHeader, map[string]string{
				utils.Tenant:            utils.CGRateSorg,
				utils.ID:                "rp1",
				utils.RateID:            "rt1",
				utils.RateIntervalStart: "60",
				utils.RateFixedFee:      "0.3",
				utils.RateRecurrentFee:  "0.4",
				utils.RateUnit:          "60",
				utils.RateIncrement:     "1",
			}),
			csvRow(t, rateHeader, map[string]string{
				utils.Tenant:              utils.CGRateSorg,
				utils.ID:                  "rp1",
				utils.RateID:              "rt2",
				utils.RateFilterIDs:       "*suffix:~*req.Destination:99",
				utils.RateActivationStart: "0 0 * * *",
				utils.RateWeights:         "rt2WeightFltr;30",
				utils.RateBlocker:         "false",
				utils.RateIntervalStart:   "0",
				utils.RateFixedFee:        "0.5",
				utils.RateRecurrentFee:    "0.6",
				utils.RateUnit:            "30",
				utils.RateIncrement:       "1",
			}),
			csvRow(t, rateHeader, map[string]string{
				utils.Tenant:              utils.CGRateSorg,
				utils.ID:                  "rp2",
				utils.FilterIDs:           "*string:~*req.Account:2001",
				utils.Weights:             "rp2WeightFltr;15",
				utils.MinCost:             "2",
				utils.MaxCost:             "20",
				utils.MaxCostStrategy:     utils.MetaMaxCostDisconnect,
				utils.RateID:              "rt3",
				utils.RateFilterIDs:       "*exists:~*opts.*cost:",
				utils.RateActivationStart: "30 6 * * 1",
				utils.RateWeights:         "rt3WeightFltr;40",
				utils.RateBlocker:         "false",
				utils.RateIntervalStart:   "0",
				utils.RateFixedFee:        "0.7",
				utils.RateRecurrentFee:    "0.8",
				utils.RateUnit:            "120",
				utils.RateIncrement:       "2",
			}),
		},
		utils.FiltersCsv: {
			filterHeader,
			csvRow(t, filterHeader, map[string]string{
				utils.Tenant: utils.CGRateSorg,
				utils.ID:     "fltr1",
				utils.Type:   utils.MetaString,
				utils.Path:   "~*req.Account",
				utils.Values: "1001;1002",
			}),
			csvRow(t, filterHeader, map[string]string{
				utils.Tenant: utils.CGRateSorg,
				utils.ID:     "fltr1",
				utils.Type:   utils.MetaPrefix,
				utils.Path:   "~*req.Destination",
				utils.Values: "+1;+44",
			}),
			csvRow(t, filterHeader, map[string]string{
				utils.Tenant: utils.CGRateSorg,
				utils.ID:     "fltr2",
				utils.Type:   utils.MetaExists,
				utils.Path:   "~*opts.*usage",
			}),
		},
		utils.AttributesCsv: {
			attributeHeader,
			csvRow(t, attributeHeader, map[string]string{
				utils.Tenant:             utils.CGRateSorg,
				utils.ID:                 "attr1",
				utils.FilterIDs:          "*string:~*req.Account:1001;*exists:~*opts.*usage:",
				utils.Weights:            "*string:~*req.Weight:attr1;10",
				utils.Blockers:           "*string:~*req.Blocker:attr1;false",
				utils.AttributeFilterIDs: "*string:~*req.Attribute:first",
				utils.AttributeBlockers:  "*string:~*req.AttrBlocker:first;true",
				utils.Path:               "*req.Subject",
				utils.Type:               utils.MetaConstant,
				utils.Value:              "1001",
			}),
			csvRow(t, attributeHeader, map[string]string{
				utils.Tenant: utils.CGRateSorg,
				utils.ID:     "attr1",
				utils.Path:   "*opts.*usage",
				utils.Type:   utils.MetaVariable,
				utils.Value:  "~*req.Usage",
			}),
			csvRow(t, attributeHeader, map[string]string{
				utils.Tenant:             utils.CGRateSorg,
				utils.ID:                 "attr2",
				utils.FilterIDs:          "*string:~*req.Category:call",
				utils.Weights:            "*string:~*req.Weight:attr2;20",
				utils.Blockers:           "*string:~*req.Blocker:attr2;false",
				utils.AttributeFilterIDs: "*string:~*req.Attribute:second",
				utils.AttributeBlockers:  "*string:~*req.AttrBlocker:second;false",
				utils.Path:               "*req.Category",
				utils.Type:               utils.MetaConstant,
				utils.Value:              "call",
			}),
		},
		utils.ResourcesCsv: {
			resourceHeader,
			csvRow(t, resourceHeader, map[string]string{
				utils.Tenant:            utils.CGRateSorg,
				utils.ID:                "res1",
				utils.FilterIDs:         "*string:~*req.Account:1001;*prefix:~*req.Destination:+1",
				utils.Weights:           "*string:~*req.Weight:res1;10",
				utils.TTL:               "1m0s",
				utils.Limit:             "2",
				utils.AllocationMessage: "allowed",
				utils.Blocker:           "true",
				utils.Stored:            "true",
				utils.ThresholdIDs:      "th1;th2",
			}),
			csvRow(t, resourceHeader, map[string]string{
				utils.Tenant:            utils.CGRateSorg,
				utils.ID:                "res2",
				utils.FilterIDs:         "*string:~*req.Account:2001",
				utils.Weights:           "*string:~*req.Weight:res2;20",
				utils.TTL:               "2m0s",
				utils.Limit:             "5",
				utils.AllocationMessage: "queued",
				utils.Blocker:           "false",
				utils.Stored:            "false",
				utils.ThresholdIDs:      "th2",
			}),
		},
		utils.StatsCsv: {
			statHeader,
			csvRow(t, statHeader, map[string]string{
				utils.Tenant:          utils.CGRateSorg,
				utils.ID:              "stat1",
				utils.FilterIDs:       "*string:~*req.Account:1001;*exists:~*opts.*usage:",
				utils.Weights:         "*string:~*req.Weight:stat1;10",
				utils.Blockers:        "*string:~*req.Blocker:stat1;false",
				utils.QueueLength:     "10",
				utils.TTL:             "1m0s",
				utils.MinItems:        "2",
				utils.Stored:          "true",
				utils.ThresholdIDs:    "th1;th2",
				utils.MetricIDs:       utils.MetaASR,
				utils.MetricFilterIDs: "*string:~*req.Metric:asr",
				utils.MetricBlockers:  "*string:~*req.MetricBlocker:asr;false",
			}),
			csvRow(t, statHeader, map[string]string{
				utils.Tenant:          utils.CGRateSorg,
				utils.ID:              "stat1",
				utils.MetricIDs:       utils.MetaTCD,
				utils.MetricFilterIDs: "*string:~*req.Metric:tcd",
				utils.MetricBlockers:  "*string:~*req.MetricBlocker:tcd;true",
			}),
			csvRow(t, statHeader, map[string]string{
				utils.Tenant:       utils.CGRateSorg,
				utils.ID:           "stat2",
				utils.QueueLength:  "5",
				utils.TTL:          "2m0s",
				utils.MinItems:     "1",
				utils.Stored:       "false",
				utils.ThresholdIDs: "th2",
				utils.MetricIDs:    utils.MetaTCC,
			}),
		},
		utils.ThresholdsCsv: {
			thresholdHeader,
			csvRow(t, thresholdHeader, map[string]string{
				utils.Tenant:           utils.CGRateSorg,
				utils.ID:               "th1",
				utils.FilterIDs:        "*string:~*req.Account:1001;*prefix:~*req.Destination:+1",
				utils.Weights:          "*string:~*req.Weight:th1;10",
				utils.MaxHits:          "10",
				utils.MinHits:          "1",
				utils.MinSleep:         "1s",
				utils.Blocker:          "true",
				utils.AttributeIDs:     "attr1;attr2",
				utils.ActionProfileIDs: "ap1",
				utils.Async:            "true",
				utils.EeIDs:            "ee1;ee2",
			}),
			csvRow(t, thresholdHeader, map[string]string{
				utils.Tenant:           utils.CGRateSorg,
				utils.ID:               "th2",
				utils.FilterIDs:        "*string:~*req.Account:2001",
				utils.Weights:          "*string:~*req.Weight:th2;20",
				utils.MaxHits:          "20",
				utils.MinHits:          "2",
				utils.MinSleep:         "2s",
				utils.Blocker:          "false",
				utils.AttributeIDs:     "attr2",
				utils.ActionProfileIDs: "ap2",
				utils.Async:            "false",
				utils.EeIDs:            "ee3",
			}),
		},
		utils.TrendsCsv: {
			trendHeader,
			csvRow(t, trendHeader, map[string]string{
				utils.Tenant:          utils.CGRateSorg,
				utils.ID:              "trend1",
				utils.Schedule:        utils.MetaASAP,
				utils.StatID:          "stat1",
				utils.Metrics:         utils.MetaASR + ";" + utils.MetaTCD,
				utils.TTL:             "1m0s",
				utils.QueueLength:     "10",
				utils.MinItems:        "2",
				utils.CorrelationType: utils.MetaLast,
				utils.Tolerance:       "0.1",
				utils.Stored:          "true",
				utils.ThresholdIDs:    "th1;th2",
			}),
			csvRow(t, trendHeader, map[string]string{
				utils.Tenant:          utils.CGRateSorg,
				utils.ID:              "trend2",
				utils.Schedule:        utils.MetaASAP,
				utils.StatID:          "stat2",
				utils.Metrics:         utils.MetaTCC,
				utils.TTL:             "2m0s",
				utils.QueueLength:     "5",
				utils.MinItems:        "1",
				utils.CorrelationType: utils.MetaLast,
				utils.Tolerance:       "0.2",
				utils.Stored:          "false",
				utils.ThresholdIDs:    "th2",
			}),
		},
		utils.RankingsCsv: {
			rankingHeader,
			csvRow(t, rankingHeader, map[string]string{
				utils.Tenant:            utils.CGRateSorg,
				utils.ID:                "rank1",
				utils.Schedule:          utils.MetaASAP,
				utils.StatIDs:           "stat1;stat2",
				utils.MetricIDs:         utils.MetaASR + ";" + utils.MetaTCD,
				utils.Sorting:           utils.MetaDesc,
				utils.SortingParameters: "param1;param2",
				utils.Stored:            "true",
				utils.ThresholdIDs:      "th1;th2",
			}),
			csvRow(t, rankingHeader, map[string]string{
				utils.Tenant:            utils.CGRateSorg,
				utils.ID:                "rank2",
				utils.Schedule:          utils.MetaASAP,
				utils.StatIDs:           "stat2",
				utils.MetricIDs:         utils.MetaTCC,
				utils.Sorting:           utils.MetaAsc,
				utils.SortingParameters: "param3",
				utils.Stored:            "false",
				utils.ThresholdIDs:      "th2",
			}),
		},
		utils.RoutesCsv: {
			routeHeader,
			csvRow(t, routeHeader, map[string]string{
				utils.Tenant:              utils.CGRateSorg,
				utils.ID:                  "routePrf1",
				utils.FilterIDs:           "*string:~*req.Account:1001;*prefix:~*req.Destination:+1",
				utils.Weights:             "*string:~*req.Weight:routePrf1;10",
				utils.Blockers:            "*string:~*req.Blocker:routePrf1;false",
				utils.Sorting:             utils.MetaWeight,
				utils.RouteID:             "route1",
				utils.RouteFilterIDs:      "*string:~*req.Route:route1",
				utils.RouteAccountIDs:     "acc1;acc2",
				utils.RouteRateProfileIDs: "rp1;rp2",
				utils.RouteResourceIDs:    "res1",
				utils.RouteStatIDs:        "stat1",
				utils.RouteWeights:        "*string:~*req.RouteWeight:route1;10",
				utils.RouteBlockers:       "*string:~*req.RouteBlocker:route1;false",
				utils.RouteParameters:     "p1:v1",
			}),
			csvRow(t, routeHeader, map[string]string{
				utils.Tenant:              utils.CGRateSorg,
				utils.ID:                  "routePrf1",
				utils.RouteID:             "route2",
				utils.RouteFilterIDs:      "*string:~*req.Route:route2",
				utils.RouteAccountIDs:     "acc2",
				utils.RouteRateProfileIDs: "rp2",
				utils.RouteResourceIDs:    "res2",
				utils.RouteStatIDs:        "stat2",
				utils.RouteWeights:        "*string:~*req.RouteWeight:route2;20",
				utils.RouteBlockers:       "*string:~*req.RouteBlocker:route2;true",
				utils.RouteParameters:     "p2:v2",
			}),
			csvRow(t, routeHeader, map[string]string{
				utils.Tenant:              utils.CGRateSorg,
				utils.ID:                  "routePrf2",
				utils.FilterIDs:           "*string:~*req.Account:2001",
				utils.Weights:             "*string:~*req.Weight:routePrf2;20",
				utils.Blockers:            "*string:~*req.Blocker:routePrf2;false",
				utils.Sorting:             utils.MetaWeight,
				utils.RouteID:             "route3",
				utils.RouteFilterIDs:      "*string:~*req.Route:route3",
				utils.RouteAccountIDs:     "acc2",
				utils.RouteRateProfileIDs: "rp2",
				utils.RouteResourceIDs:    "res2",
				utils.RouteStatIDs:        "stat2",
				utils.RouteWeights:        "*string:~*req.RouteWeight:route3;30",
				utils.RouteBlockers:       "*string:~*req.RouteBlocker:route3;false",
				utils.RouteParameters:     "p3:v3",
			}),
		},
		utils.ChargersCsv: {
			chargerHeader,
			csvRow(t, chargerHeader, map[string]string{
				utils.Tenant:       utils.CGRateSorg,
				utils.ID:           "charger1",
				utils.FilterIDs:    "*string:~*req.Account:1001;*exists:~*opts.*usage:",
				utils.Weights:      "*string:~*req.Weight:charger1;10",
				utils.Blockers:     "*string:~*req.Blocker:charger1;false",
				utils.RunID:        utils.MetaDefault,
				utils.AttributeIDs: "attr1;attr2",
			}),
			csvRow(t, chargerHeader, map[string]string{
				utils.Tenant:       utils.CGRateSorg,
				utils.ID:           "charger2",
				utils.FilterIDs:    "*string:~*req.Account:2001",
				utils.Weights:      "*string:~*req.Weight:charger2;20",
				utils.Blockers:     "*string:~*req.Blocker:charger2;true",
				utils.RunID:        "run2",
				utils.AttributeIDs: "attr2",
			}),
		},
	}
	inputZip := zipCSV(t, files)
	cfg, dm, connMgr, filters := newTestEnv(t, files)

	loader := loaders.NewLoaderS(cfg, dm, filters, connMgr)
	var loaderReply string
	if err := loader.V1ImportZip(context.Background(), &loaders.ArgsProcessZip{
		Data: inputZip,
		APIOpts: map[string]any{
			utils.MetaCache:       utils.MetaNone,
			utils.MetaWithIndex:   true,
			utils.MetaStopOnError: true,
			utils.MetaForceLock:   true,
		},
	}, &loaderReply); err != nil {
		t.Fatal(err)
	} else if loaderReply != utils.OK {
		t.Fatalf("loader reply: want %q, got %q", utils.OK, loaderReply)
	}

	tp := NewTPeS(cfg, dm, connMgr)
	var outputZip []byte
	if err := tp.V1ExportTariffPlan(context.Background(), &ArgsExportTP{
		Tenant:      utils.CGRateSorg,
		ExportItems: exportItemsForFiles(t, files),
	}, &outputZip); err != nil {
		t.Fatal(err)
	}

	checkCSVZip(t, inputZip, outputZip)
}

func newTestEnv(t *testing.T, files map[string][][]string) (*config.CGRConfig, *engine.DataManager, *engine.ConnManager, *engine.FilterS) {
	t.Helper()
	cfg := config.NewDefaultCGRConfig()
	cfg.LoaderCfg()[0].Enabled = true
	cfg.LoaderCfg()[0].LockFilePath = utils.MetaMemory
	cfg.LoaderCfg()[0].TpInDir = ""
	cfg.LoaderCfg()[0].TpOutDir = ""
	cfg.LoaderCfg()[0].Data = filterLoaderData(t, cfg.LoaderCfg()[0].Data, files)

	cache := engine.NewCacheS(cfg, nil, nil, nil)
	connMgr := engine.NewConnManager(cfg)
	connMgr.SetCache(cache)
	internalDB, err := engine.NewInternalDB(nil, nil, nil, cfg.DbCfg().Items)
	if err != nil {
		t.Fatal(err)
	}
	dbConn := engine.NewDBConnManager(map[string]engine.DataDB{utils.MetaDefault: internalDB}, cfg.DbCfg())
	dm := engine.NewDataManager(dbConn, cfg, connMgr)
	dm.SetCache(cache)
	filters := engine.NewFilterS(cfg, connMgr, dm)
	return cfg, dm, connMgr, filters
}

func filterLoaderData(t *testing.T, data []*config.LoaderDataType, files map[string][][]string) []*config.LoaderDataType {
	t.Helper()
	wanted := make(map[string]struct{}, len(files))
	for filename := range files {
		wanted[filename] = struct{}{}
	}
	filtered := make([]*config.LoaderDataType, 0, len(wanted))
	for _, dataType := range data {
		if _, has := wanted[dataType.Filename]; !has {
			continue
		}
		filtered = append(filtered, dataType)
		delete(wanted, dataType.Filename)
	}
	if len(wanted) != 0 {
		t.Fatalf("default loader config is missing data for %v", slices.Sorted(maps.Keys(wanted)))
	}
	return filtered
}

func exportItemsForFiles(t *testing.T, files map[string][][]string) map[string][]string {
	t.Helper()
	if len(files) != len(exporters) {
		t.Fatalf("test data has %d CSV files, want %d", len(files), len(exporters))
	}
	if len(itemIDsByExport) != len(exporters) {
		t.Fatalf("export items have %d export types, want %d", len(itemIDsByExport), len(exporters))
	}
	items := make(map[string][]string, len(itemIDsByExport))
	for exportType, itemIDs := range itemIDsByExport {
		exporter, has := exporters[exportType]
		if !has {
			t.Fatalf("export items include unknown type %s", exportType)
		}
		if _, has := files[exporter.fileName]; !has {
			t.Fatalf("test data is missing %s", exporter.fileName)
		}
		items[exportType] = itemIDs
	}
	return items
}

func zipCSV(t *testing.T, files map[string][][]string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	for _, filename := range slices.Sorted(maps.Keys(files)) {
		w, err := zipWriter.Create(filename)
		if err != nil {
			t.Fatal(err)
		}
		csvWriter := csv.NewWriter(w)
		csvWriter.Comma = utils.CSVSep
		if err := csvWriter.WriteAll(files[filename]); err != nil {
			t.Fatal(err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func readCSVZip(t *testing.T, data []byte) map[string][][]string {
	t.Helper()
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string][][]string, len(zipReader.File))
	for _, file := range zipReader.File {
		if _, has := files[file.Name]; has {
			t.Fatalf("zip contains duplicate file %s", file.Name)
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		records, err := readCSVRecords(reader)
		if closeErr := reader.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		if err != nil {
			t.Fatalf("%s: %v", file.Name, err)
		}
		files[file.Name] = records
	}
	return files
}

func readCSVRecords(reader io.Reader) ([][]string, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = utils.CSVSep
	csvReader.FieldsPerRecord = -1
	return csvReader.ReadAll()
}

func checkCSVZip(t *testing.T, wantZip, gotZip []byte) {
	t.Helper()
	want := readCSVZip(t, wantZip)
	got := readCSVZip(t, gotZip)
	for _, filename := range slices.Sorted(maps.Keys(want)) {
		wantRecords := want[filename]
		gotRecords, has := got[filename]
		if !has {
			t.Errorf("exported zip is missing %s", filename)
			continue
		}
		if !reflect.DeepEqual(wantRecords, gotRecords) {
			t.Errorf("%s records differ\nwant: %v\ngot: %v", filename, wantRecords, gotRecords)
		}
		delete(got, filename)
	}
	if len(got) != 0 {
		t.Errorf("exported zip has unexpected files %v", slices.Sorted(maps.Keys(got)))
	}
}

func csvRow(t *testing.T, header []string, fields map[string]string) []string {
	t.Helper()
	indexes := make(map[string]int, len(header))
	for i, name := range header {
		name = strings.TrimPrefix(name, "#")
		if _, has := indexes[name]; has {
			t.Fatalf("duplicate header field %q", name)
		}
		indexes[name] = i
	}
	record := make([]string, len(header))
	for name, value := range fields {
		idx, has := indexes[name]
		if !has {
			t.Fatalf("header is missing field %q", name)
		}
		record[idx] = value
	}
	return record
}
