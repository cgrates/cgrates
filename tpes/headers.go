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
