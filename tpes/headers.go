// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package tpes

import "github.com/cgrates/cgrates/utils"

var accountHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.Blockers,
	utils.Opts,
	utils.BalanceID,
	utils.BalanceFilterIDs,
	utils.BalanceWeights,
	utils.BalanceBlockers,
	utils.BalanceType,
	utils.BalanceUnits,
	utils.BalanceUnitFactors,
	utils.BalanceOpts,
	utils.BalanceCostIncrements,
	utils.BalanceAttributeIDs,
	utils.BalanceRateProfileIDs,
	utils.ThresholdIDs,
}

var actionHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.Blockers,
	utils.Schedule,
	utils.TargetType,
	utils.TargetIDs,
	utils.ActionID,
	utils.ActionFilterIDs,
	utils.ActionTTL,
	utils.ActionType,
	utils.ActionOpts,
	utils.ActionWeights,
	utils.ActionBlockers,
	utils.ActionDiktatsID,
	utils.ActionDiktatsFilterIDs,
	utils.ActionDiktatsOpts,
	utils.ActionDiktatsWeights,
	utils.ActionDiktatsBlockers,
}

var attributeHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.Blockers,
	utils.AttributeFilterIDs,
	utils.AttributeBlockers,
	utils.Path,
	utils.Type,
	utils.Value,
}

var chargerHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.Blockers,
	utils.RunID,
	utils.AttributeIDs,
}

var filterHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.Type,
	utils.Path,
	utils.Values,
}

var rankingHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.Schedule,
	utils.StatIDs,
	utils.MetricIDs,
	utils.Sorting,
	utils.SortingParameters,
	utils.Stored,
	utils.ThresholdIDs,
}

var rateHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.MinCost,
	utils.MaxCost,
	utils.MaxCostStrategy,
	utils.RateID,
	utils.RateFilterIDs,
	utils.RateActivationStart,
	utils.RateWeights,
	utils.RateBlocker,
	utils.RateIntervalStart,
	utils.RateFixedFee,
	utils.RateRecurrentFee,
	utils.RateUnit,
	utils.RateIncrement,
}

var resourceHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.TTL,
	utils.Limit,
	utils.AllocationMessage,
	utils.Blocker,
	utils.Stored,
	utils.ThresholdIDs,
}

var routeHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.Blockers,
	utils.Sorting,
	utils.SortingParameters,
	utils.RouteID,
	utils.RouteFilterIDs,
	utils.RouteAccountIDs,
	utils.RouteRateProfileIDs,
	utils.RouteResourceIDs,
	utils.RouteStatIDs,
	utils.RouteWeights,
	utils.RouteBlockers,
	utils.RouteParameters,
}

var statHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.Blockers,
	utils.QueueLength,
	utils.TTL,
	utils.MinItems,
	utils.Stored,
	utils.ThresholdIDs,
	utils.MetricIDs,
	utils.MetricFilterIDs,
	utils.MetricBlockers,
}

var thresholdHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.FilterIDs,
	utils.Weights,
	utils.MaxHits,
	utils.MinHits,
	utils.MinSleep,
	utils.Blocker,
	utils.AttributeIDs,
	utils.ActionProfileIDs,
	utils.Async,
	utils.EeIDs,
}

var trendHeader = []string{
	"#" + utils.Tenant,
	utils.ID,
	utils.Schedule,
	utils.StatID,
	utils.Metrics,
	utils.TTL,
	utils.QueueLength,
	utils.MinItems,
	utils.CorrelationType,
	utils.Tolerance,
	utils.Stored,
	utils.ThresholdIDs,
}
