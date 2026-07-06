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
	"encoding/csv"
	"io"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newCSVWriter(w io.Writer) *csv.Writer {
	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = utils.CSVSep
	return csvWriter
}

// row lays out named fields in header order; fields missing from the map are
// left empty, which is how continuation rows drop the profile-level columns.
func row(header []string, fields map[string]string) []string {
	r := make([]string, len(header))
	for i, name := range header {
		r[i] = fields[name]
	}
	return r
}

func joinValues(values []string) string {
	return strings.Join(values, utils.InfieldSep)
}

func durationRecordValue(value time.Duration) string {
	if value == 0 {
		return ""
	}
	return utils.IfaceAsString(value)
}

func decimalRecordValue(value *utils.Decimal) string {
	if value == nil || value.Big == nil {
		return ""
	}
	return value.String()
}

func mapRecordValue(values map[string]any) string {
	if len(values) == 0 {
		return ""
	}
	keys := slices.Sorted(maps.Keys(values))
	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = utils.ConcatenatedKey(key, utils.IfaceAsString(values[key]))
	}
	return strings.Join(parts, utils.InfieldSep)
}

func stringSetRecordValue(values utils.StringSet) string {
	return joinValues(values.AsOrderedSlice())
}

func unitFactorsRecordValue(factors []*utils.UnitFactor) string {
	if len(factors) == 0 {
		return ""
	}
	fields := make([]string, 0, len(factors)*2)
	for _, factor := range factors {
		fields = append(fields,
			strings.Join(factor.FilterIDs, utils.ANDSep),
			decimalRecordValue(factor.Factor),
		)
	}
	return strings.Join(fields, utils.InfieldSep)
}

func costIncrementsRecordValue(increments []*utils.CostIncrement) string {
	if len(increments) == 0 {
		return ""
	}
	fields := make([]string, 0, len(increments)*4)
	for _, increment := range increments {
		fields = append(fields,
			strings.Join(increment.FilterIDs, utils.ANDSep),
			decimalRecordValue(increment.Increment),
			decimalRecordValue(increment.FixedFee),
			decimalRecordValue(increment.RecurrentFee),
		)
	}
	return strings.Join(fields, utils.InfieldSep)
}

// identityFields returns the tenant/id key, repeated on every row of a multi-row
// profile so the loader can group rows back into one profile.
func identityFields(tenant, id string) map[string]string {
	return map[string]string{
		"#" + utils.Tenant: tenant,
		utils.ID:           id,
	}
}

func filterRecords(filter *engine.Filter) [][]string {
	if len(filter.Rules) == 0 {
		return [][]string{row(filterHeader, identityFields(filter.Tenant, filter.ID))}
	}
	records := make([][]string, 0, len(filter.Rules))
	for _, rule := range filter.Rules {
		fields := identityFields(filter.Tenant, filter.ID)
		fields[utils.Type] = rule.Type
		fields[utils.Path] = rule.Element
		fields[utils.Values] = joinValues(rule.Values)
		records = append(records, row(filterHeader, fields))
	}
	return records
}

func resourceProfileRecords(profile *utils.ResourceProfile) [][]string {
	fields := identityFields(profile.Tenant, profile.ID)
	fields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	fields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	fields[utils.TTL] = durationRecordValue(profile.UsageTTL)
	fields[utils.Limit] = utils.IfaceAsString(profile.Limit)
	fields[utils.AllocationMessage] = profile.AllocationMessage
	fields[utils.Blocker] = utils.IfaceAsString(profile.Blocker)
	fields[utils.Stored] = utils.IfaceAsString(profile.Stored)
	fields[utils.ThresholdIDs] = joinValues(profile.ThresholdIDs)
	return [][]string{row(resourceHeader, fields)}
}

func chargerProfileRecords(profile *utils.ChargerProfile) [][]string {
	fields := identityFields(profile.Tenant, profile.ID)
	fields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	fields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	fields[utils.Blockers] = profile.Blockers.String(utils.InfieldSep, utils.ANDSep)
	fields[utils.RunID] = profile.RunID
	fields[utils.AttributeIDs] = joinValues(profile.AttributeIDs)
	return [][]string{row(chargerHeader, fields)}
}

func rankingProfileRecords(profile *utils.RankingProfile) [][]string {
	fields := identityFields(profile.Tenant, profile.ID)
	fields[utils.Schedule] = profile.Schedule
	fields[utils.StatIDs] = joinValues(profile.StatIDs)
	fields[utils.MetricIDs] = joinValues(profile.MetricIDs)
	fields[utils.Sorting] = profile.Sorting
	fields[utils.SortingParameters] = joinValues(profile.SortingParameters)
	fields[utils.Stored] = utils.IfaceAsString(profile.Stored)
	fields[utils.ThresholdIDs] = joinValues(profile.ThresholdIDs)
	return [][]string{row(rankingHeader, fields)}
}

func thresholdProfileRecords(profile *utils.ThresholdProfile) [][]string {
	fields := identityFields(profile.Tenant, profile.ID)
	fields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	fields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	fields[utils.MaxHits] = utils.IfaceAsString(profile.MaxHits)
	fields[utils.MinHits] = utils.IfaceAsString(profile.MinHits)
	fields[utils.MinSleep] = durationRecordValue(profile.MinSleep)
	fields[utils.Blocker] = utils.IfaceAsString(profile.Blocker)
	fields[utils.AttributeIDs] = joinValues(profile.AttributeIDs)
	fields[utils.ActionProfileIDs] = joinValues(profile.ActionProfileIDs)
	fields[utils.Async] = utils.IfaceAsString(profile.Async)
	fields[utils.EeIDs] = joinValues(profile.EeIDs)
	return [][]string{row(thresholdHeader, fields)}
}

func trendProfileRecords(profile *utils.TrendProfile) [][]string {
	fields := identityFields(profile.Tenant, profile.ID)
	fields[utils.Schedule] = profile.Schedule
	fields[utils.StatID] = profile.StatID
	fields[utils.Metrics] = joinValues(profile.Metrics)
	fields[utils.TTL] = durationRecordValue(profile.TTL)
	fields[utils.QueueLength] = utils.IfaceAsString(profile.QueueLength)
	fields[utils.MinItems] = utils.IfaceAsString(profile.MinItems)
	fields[utils.CorrelationType] = profile.CorrelationType
	fields[utils.Tolerance] = utils.IfaceAsString(profile.Tolerance)
	fields[utils.Stored] = utils.IfaceAsString(profile.Stored)
	fields[utils.ThresholdIDs] = joinValues(profile.ThresholdIDs)
	return [][]string{row(trendHeader, fields)}
}

func attributeProfileRecords(profile *utils.AttributeProfile) [][]string {
	profileFields := identityFields(profile.Tenant, profile.ID)
	profileFields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	profileFields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Blockers] = profile.Blockers.String(utils.InfieldSep, utils.ANDSep)
	if len(profile.Attributes) == 0 {
		return [][]string{row(attributeHeader, profileFields)}
	}
	records := make([][]string, 0, len(profile.Attributes))
	for i, attr := range profile.Attributes {
		fields := identityFields(profile.Tenant, profile.ID)
		fields[utils.AttributeFilterIDs] = joinValues(attr.FilterIDs)
		fields[utils.AttributeBlockers] = attr.Blockers.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.Path] = attr.Path
		fields[utils.Type] = attr.Type
		fields[utils.Value] = attr.Value.GetRule()
		if i == 0 {
			maps.Copy(fields, profileFields)
		}
		records = append(records, row(attributeHeader, fields))
	}
	return records
}

func statQueueProfileRecords(profile *utils.StatQueueProfile) [][]string {
	profileFields := identityFields(profile.Tenant, profile.ID)
	profileFields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	profileFields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Blockers] = profile.Blockers.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.QueueLength] = utils.IfaceAsString(profile.QueueLength)
	profileFields[utils.TTL] = durationRecordValue(profile.TTL)
	profileFields[utils.MinItems] = utils.IfaceAsString(profile.MinItems)
	profileFields[utils.Stored] = utils.IfaceAsString(profile.Stored)
	profileFields[utils.ThresholdIDs] = joinValues(profile.ThresholdIDs)
	if len(profile.Metrics) == 0 {
		return [][]string{row(statHeader, profileFields)}
	}
	records := make([][]string, 0, len(profile.Metrics))
	for i, metric := range profile.Metrics {
		fields := identityFields(profile.Tenant, profile.ID)
		fields[utils.MetricIDs] = metric.MetricID
		fields[utils.MetricFilterIDs] = joinValues(metric.FilterIDs)
		fields[utils.MetricBlockers] = metric.Blockers.String(utils.InfieldSep, utils.ANDSep)
		if i == 0 {
			maps.Copy(fields, profileFields)
		}
		records = append(records, row(statHeader, fields))
	}
	return records
}

func routeProfileRecords(profile *utils.RouteProfile) [][]string {
	profileFields := identityFields(profile.Tenant, profile.ID)
	profileFields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	profileFields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Blockers] = profile.Blockers.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Sorting] = profile.Sorting
	profileFields[utils.SortingParameters] = joinValues(profile.SortingParameters)
	if len(profile.Routes) == 0 {
		return [][]string{row(routeHeader, profileFields)}
	}
	records := make([][]string, 0, len(profile.Routes))
	for i, route := range profile.Routes {
		fields := identityFields(profile.Tenant, profile.ID)
		fields[utils.RouteID] = route.ID
		fields[utils.RouteFilterIDs] = joinValues(route.FilterIDs)
		fields[utils.RouteAccountIDs] = joinValues(route.AccountIDs)
		fields[utils.RouteRateProfileIDs] = joinValues(route.RateProfileIDs)
		fields[utils.RouteResourceIDs] = joinValues(route.ResourceIDs)
		fields[utils.RouteStatIDs] = joinValues(route.StatIDs)
		fields[utils.RouteWeights] = route.Weights.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.RouteBlockers] = route.Blockers.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.RouteParameters] = route.RouteParameters
		if i == 0 {
			maps.Copy(fields, profileFields)
		}
		records = append(records, row(routeHeader, fields))
	}
	return records
}

func rateProfileRecords(profile *utils.RateProfile) [][]string {
	profileFields := identityFields(profile.Tenant, profile.ID)
	profileFields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	profileFields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.MinCost] = decimalRecordValue(profile.MinCost)
	profileFields[utils.MaxCost] = decimalRecordValue(profile.MaxCost)
	profileFields[utils.MaxCostStrategy] = profile.MaxCostStrategy
	rateIDs := slices.Sorted(maps.Keys(profile.Rates))
	if len(rateIDs) == 0 {
		return [][]string{row(rateHeader, profileFields)}
	}
	rateFields := func(rateID string, rate *utils.Rate) map[string]string {
		fields := identityFields(profile.Tenant, profile.ID)
		fields[utils.RateID] = rateID
		fields[utils.RateFilterIDs] = joinValues(rate.FilterIDs)
		fields[utils.RateActivationStart] = rate.ActivationTimes
		fields[utils.RateWeights] = rate.Weights.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.RateBlocker] = utils.IfaceAsString(rate.Blocker)
		return fields
	}
	intervalFields := func(tenant, id string, interval *utils.IntervalRate) map[string]string {
		fields := identityFields(tenant, id)
		fields[utils.RateIntervalStart] = decimalRecordValue(interval.IntervalStart)
		fields[utils.RateFixedFee] = decimalRecordValue(interval.FixedFee)
		fields[utils.RateRecurrentFee] = decimalRecordValue(interval.RecurrentFee)
		fields[utils.RateUnit] = decimalRecordValue(interval.Unit)
		fields[utils.RateIncrement] = decimalRecordValue(interval.Increment)
		return fields
	}
	var records [][]string
	first := true
	add := func(fields map[string]string) {
		if first {
			maps.Copy(fields, profileFields)
			first = false
		}
		records = append(records, row(rateHeader, fields))
	}
	for _, rateID := range rateIDs {
		rate := profile.Rates[rateID]
		if rate.ID != "" {
			rateID = rate.ID
		}
		if len(rate.IntervalRates) == 0 {
			add(rateFields(rateID, rate))
			continue
		}
		for i, interval := range rate.IntervalRates {
			fields := intervalFields(profile.Tenant, profile.ID, interval)
			if i == 0 {
				fields[utils.RateFilterIDs] = joinValues(rate.FilterIDs)
				fields[utils.RateActivationStart] = rate.ActivationTimes
				fields[utils.RateWeights] = rate.Weights.String(utils.InfieldSep, utils.ANDSep)
				fields[utils.RateBlocker] = utils.IfaceAsString(rate.Blocker)
			}
			fields[utils.RateID] = rateID
			add(fields)
		}
	}
	return records
}

func actionProfileRecords(profile *utils.ActionProfile) [][]string {
	profileFields := identityFields(profile.Tenant, profile.ID)
	profileFields[utils.FilterIDs] = joinValues(profile.FilterIDs)
	profileFields[utils.Weights] = profile.Weights.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Blockers] = profile.Blockers.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Schedule] = profile.Schedule
	actionFields := func(action *utils.APAction) map[string]string {
		fields := identityFields(profile.Tenant, profile.ID)
		fields[utils.ActionID] = action.ID
		fields[utils.ActionFilterIDs] = joinValues(action.FilterIDs)
		fields[utils.ActionTTL] = durationRecordValue(action.TTL)
		fields[utils.ActionType] = action.Type
		fields[utils.ActionOpts] = mapRecordValue(action.Opts)
		fields[utils.ActionWeights] = action.Weights.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.ActionBlockers] = action.Blockers.String(utils.InfieldSep, utils.ANDSep)
		return fields
	}
	diktatFields := func(diktat *utils.APDiktat) map[string]string {
		fields := identityFields(profile.Tenant, profile.ID)
		fields[utils.ActionDiktatsID] = diktat.ID
		fields[utils.ActionDiktatsFilterIDs] = joinValues(diktat.FilterIDs)
		fields[utils.ActionDiktatsOpts] = mapRecordValue(diktat.Opts)
		fields[utils.ActionDiktatsWeights] = diktat.Weights.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.ActionDiktatsBlockers] = diktat.Blockers.String(utils.InfieldSep, utils.ANDSep)
		return fields
	}
	var records [][]string
	first := true
	add := func(fields map[string]string) {
		if first {
			maps.Copy(fields, profileFields)
			first = false
		}
		records = append(records, row(actionHeader, fields))
	}
	for _, targetType := range slices.Sorted(maps.Keys(profile.Targets)) {
		if targetType == "" {
			continue
		}
		fields := identityFields(profile.Tenant, profile.ID)
		fields[utils.TargetType] = targetType
		fields[utils.TargetIDs] = stringSetRecordValue(profile.Targets[targetType])
		add(fields)
	}
	for _, action := range profile.Actions {
		if len(action.Diktats) == 0 {
			add(actionFields(action))
			continue
		}
		for i, diktat := range action.Diktats {
			fields := diktatFields(diktat)
			if i == 0 {
				fields[utils.ActionFilterIDs] = joinValues(action.FilterIDs)
				fields[utils.ActionTTL] = durationRecordValue(action.TTL)
				fields[utils.ActionType] = action.Type
				fields[utils.ActionOpts] = mapRecordValue(action.Opts)
				fields[utils.ActionWeights] = action.Weights.String(utils.InfieldSep, utils.ANDSep)
				fields[utils.ActionBlockers] = action.Blockers.String(utils.InfieldSep, utils.ANDSep)
			}
			fields[utils.ActionID] = action.ID
			add(fields)
		}
	}
	if len(records) == 0 {
		add(identityFields(profile.Tenant, profile.ID))
	}
	return records
}

func accountRecords(account *utils.Account) [][]string {
	profileFields := identityFields(account.Tenant, account.ID)
	profileFields[utils.FilterIDs] = joinValues(account.FilterIDs)
	profileFields[utils.Weights] = account.Weights.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Blockers] = account.Blockers.String(utils.InfieldSep, utils.ANDSep)
	profileFields[utils.Opts] = mapRecordValue(account.Opts)
	profileFields[utils.ThresholdIDs] = joinValues(account.ThresholdIDs)
	balanceIDs := slices.Sorted(maps.Keys(account.Balances))
	if len(balanceIDs) == 0 {
		return [][]string{row(accountHeader, profileFields)}
	}
	records := make([][]string, 0, len(balanceIDs))
	for i, balanceID := range balanceIDs {
		balance := account.Balances[balanceID]
		if balance.ID != "" {
			balanceID = balance.ID
		}
		fields := identityFields(account.Tenant, account.ID)
		fields[utils.BalanceID] = balanceID
		fields[utils.BalanceFilterIDs] = joinValues(balance.FilterIDs)
		fields[utils.BalanceWeights] = balance.Weights.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.BalanceBlockers] = balance.Blockers.String(utils.InfieldSep, utils.ANDSep)
		fields[utils.BalanceType] = balance.Type
		fields[utils.BalanceUnits] = decimalRecordValue(balance.Units)
		fields[utils.BalanceUnitFactors] = unitFactorsRecordValue(balance.UnitFactors)
		fields[utils.BalanceOpts] = mapRecordValue(balance.Opts)
		fields[utils.BalanceCostIncrements] = costIncrementsRecordValue(balance.CostIncrements)
		fields[utils.BalanceAttributeIDs] = joinValues(balance.AttributeIDs)
		fields[utils.BalanceRateProfileIDs] = joinValues(balance.RateProfileIDs)
		if i == 0 {
			maps.Copy(fields, profileFields)
		}
		records = append(records, row(accountHeader, fields))
	}
	return records
}
