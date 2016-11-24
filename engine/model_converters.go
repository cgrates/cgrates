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
package engine

import (
	"strconv"

	"github.com/cgrates/cgrates/utils"
)

func APItoModelTiming(t *utils.ApierTPTiming) (result *TpTiming) {
	return &TpTiming{
		Tpid:      t.TPid,
		Tag:       t.TimingId,
		Years:     t.Years,
		Months:    t.Months,
		MonthDays: t.MonthDays,
		WeekDays:  t.WeekDays,
		Time:      t.Time,
	}
}

func APItoModelApierTiming(t *utils.ApierTPTiming) (result *TpTiming) {
	return &TpTiming{
		Tpid:      t.TPid,
		Tag:       t.TimingId,
		Years:     t.Years,
		Months:    t.Months,
		MonthDays: t.MonthDays,
		WeekDays:  t.WeekDays,
		Time:      t.Time,
	}
}

func APItoModelDestination(dest *utils.TPDestination) (result []TpDestination) {
	for _, p := range dest.Prefixes {
		result = append(result, TpDestination{
			Tpid:   dest.TPid,
			Tag:    dest.Tag,
			Prefix: p,
		})
	}
	if len(dest.Prefixes) == 0 {
		result = append(result, TpDestination{
			Tpid: dest.TPid,
			Tag:  dest.Tag,
		})
	}
	return
}

func APItoModelRate(r *utils.TPRate) (result []TpRate) {
	for _, rs := range r.RateSlots {
		result = append(result, TpRate{
			Tpid:               r.TPid,
			Tag:                r.RateId,
			ConnectFee:         rs.ConnectFee,
			Rate:               rs.Rate,
			RateUnit:           rs.RateUnit,
			RateIncrement:      rs.RateIncrement,
			GroupIntervalStart: rs.GroupIntervalStart,
		})
	}
	if len(r.RateSlots) == 0 {
		result = append(result, TpRate{
			Tpid: r.TPid,
			Tag:  r.RateId,
		})
	}
	return
}

func APItoModelDestinationRate(drs *utils.TPDestinationRate) (result []TpDestinationRate) {
	for _, dr := range drs.DestinationRates {
		result = append(result, TpDestinationRate{
			Tpid:             drs.TPid,
			Tag:              drs.DestinationRateId,
			DestinationsTag:  dr.DestinationId,
			RatesTag:         dr.RateId,
			RoundingMethod:   dr.RoundingMethod,
			RoundingDecimals: dr.RoundingDecimals,
			MaxCost:          dr.MaxCost,
			MaxCostStrategy:  dr.MaxCostStrategy,
		})
	}
	if len(drs.DestinationRates) == 0 {
		result = append(result, TpDestinationRate{
			Tpid: drs.TPid,
			Tag:  drs.DestinationRateId,
		})
	}
	return
}

func APItoModelRatingPlan(rps *utils.TPRatingPlan) (result []TpRatingPlan) {
	for _, rp := range rps.RatingPlanBindings {
		result = append(result, TpRatingPlan{
			Tpid:         rps.TPid,
			Tag:          rps.RatingPlanId,
			DestratesTag: rp.DestinationRatesId,
			TimingTag:    rp.TimingId,
			Weight:       rp.Weight,
		})
	}
	if len(rps.RatingPlanBindings) == 0 {
		result = append(result, TpRatingPlan{
			Tpid: rps.TPid,
			Tag:  rps.RatingPlanId,
		})
	}
	return
}

func APItoModelRatingProfile(rpf *utils.TPRatingProfile) (result []TpRatingProfile) {
	for _, ra := range rpf.RatingPlanActivations {
		result = append(result, TpRatingProfile{
			Tpid:             rpf.TPid,
			Loadid:           rpf.LoadId,
			Direction:        rpf.Direction,
			Tenant:           rpf.Tenant,
			Category:         rpf.Category,
			Subject:          rpf.Subject,
			ActivationTime:   ra.ActivationTime,
			RatingPlanTag:    ra.RatingPlanId,
			FallbackSubjects: ra.FallbackSubjects,
			CdrStatQueueIds:  ra.CdrStatQueueIds,
		})
	}
	if len(rpf.RatingPlanActivations) == 0 {
		result = append(result, TpRatingProfile{
			Tpid:      rpf.TPid,
			Loadid:    rpf.LoadId,
			Direction: rpf.Direction,
			Tenant:    rpf.Tenant,
			Category:  rpf.Category,
			Subject:   rpf.Subject,
		})
	}
	return
}

func APItoModelLcrRule(lcrs *utils.TPLcrRules) (result []TpLcrRule) {
	for _, lcr := range lcrs.Rules {
		result = append(result, TpLcrRule{
			Tpid:           lcrs.TPid,
			Direction:      lcrs.Direction,
			Tenant:         lcrs.Tenant,
			Category:       lcrs.Category,
			Account:        lcrs.Account,
			Subject:        lcrs.Subject,
			DestinationTag: lcr.DestinationId,
			RpCategory:     lcr.RpCategory,
			Strategy:       lcr.Strategy,
			StrategyParams: lcr.StrategyParams,
			ActivationTime: lcr.ActivationTime,
			Weight:         lcr.Weight,
		})
	}
	if len(lcrs.Rules) == 0 {
		result = append(result, TpLcrRule{
			Tpid: lcrs.TPid,
		})
	}
	return
}

func APItoModelAction(as *utils.TPActions) (result []TpAction) {
	for _, a := range as.Actions {
		result = append(result, TpAction{
			Tpid:            as.TPid,
			Tag:             as.ActionsId,
			Action:          a.Identifier,
			BalanceTag:      a.BalanceId,
			BalanceType:     a.BalanceType,
			Directions:      a.Directions,
			Units:           a.Units,
			ExpiryTime:      a.ExpiryTime,
			Filter:          a.Filter,
			TimingTags:      a.TimingTags,
			DestinationTags: a.DestinationIds,
			RatingSubject:   a.RatingSubject,
			Categories:      a.Categories,
			SharedGroups:    a.SharedGroups,
			BalanceWeight:   a.BalanceWeight,
			BalanceBlocker:  a.BalanceBlocker,
			BalanceDisabled: a.BalanceDisabled,
			ExtraParameters: a.ExtraParameters,
			Weight:          a.Weight,
		})
	}
	if len(as.Actions) == 0 {
		result = append(result, TpAction{
			Tpid: as.TPid,
			Tag:  as.ActionsId,
		})
	}
	return
}

func APItoModelActionPlan(aps *utils.TPActionPlan) (result []TpActionPlan) {
	for _, ap := range aps.ActionPlan {
		result = append(result, TpActionPlan{
			Tpid:       aps.TPid,
			Tag:        aps.ActionPlanId,
			ActionsTag: ap.ActionsId,
			TimingTag:  ap.TimingId,
			Weight:     ap.Weight,
		})
	}
	if len(aps.ActionPlan) == 0 {
		result = append(result, TpActionPlan{
			Tpid: aps.TPid,
			Tag:  aps.ActionPlanId,
		})
	}
	return
}

func APItoModelActionTrigger(ats *utils.TPActionTriggers) (result []TpActionTrigger) {
	for _, at := range ats.ActionTriggers {
		result = append(result, TpActionTrigger{
			Tpid:                   ats.TPid,
			Tag:                    at.Id,
			UniqueId:               at.UniqueID,
			ThresholdType:          at.ThresholdType,
			ThresholdValue:         at.ThresholdValue,
			Recurrent:              at.Recurrent,
			MinSleep:               at.MinSleep,
			ExpiryTime:             at.ExpirationDate,
			ActivationTime:         at.ActivationDate,
			BalanceTag:             at.BalanceId,
			BalanceType:            at.BalanceType,
			BalanceDirections:      at.BalanceDirections,
			BalanceDestinationTags: at.BalanceDestinationIds,
			BalanceWeight:          at.BalanceWeight,
			BalanceExpiryTime:      at.BalanceExpirationDate,
			BalanceTimingTags:      at.BalanceTimingTags,
			BalanceRatingSubject:   at.BalanceRatingSubject,
			BalanceCategories:      at.BalanceCategories,
			BalanceSharedGroups:    at.BalanceSharedGroups,
			BalanceBlocker:         at.BalanceBlocker,
			BalanceDisabled:        at.BalanceDisabled,
			MinQueuedItems:         at.MinQueuedItems,
			ActionsTag:             at.ActionsId,
			Weight:                 at.Weight,
		})
	}
	if len(ats.ActionTriggers) == 0 {
		result = append(result, TpActionTrigger{
			Tpid: ats.TPid,
			Tag:  ats.ActionTriggersId,
		})
	}
	return
}

func APItoModelAccountAction(aa *utils.TPAccountActions) *TpAccountAction {
	return &TpAccountAction{
		Tpid:              aa.TPid,
		Loadid:            aa.LoadId,
		Tenant:            aa.Tenant,
		Account:           aa.Account,
		ActionPlanTag:     aa.ActionPlanId,
		ActionTriggersTag: aa.ActionTriggersId,
	}
}

func APItoModelSharedGroup(sgs *utils.TPSharedGroups) (result []TpSharedGroup) {
	for _, sg := range sgs.SharedGroups {
		result = append(result, TpSharedGroup{
			Tpid:          sgs.TPid,
			Tag:           sgs.SharedGroupsId,
			Account:       sg.Account,
			Strategy:      sg.Strategy,
			RatingSubject: sg.RatingSubject,
		})
	}
	if len(sgs.SharedGroups) == 0 {
		result = append(result, TpSharedGroup{
			Tpid: sgs.TPid,
			Tag:  sgs.SharedGroupsId,
		})
	}
	return
}

func APItoModelDerivedCharger(dcs *utils.TPDerivedChargers) (result []TpDerivedCharger) {
	for _, dc := range dcs.DerivedChargers {
		result = append(result, TpDerivedCharger{
			Tpid:                 dcs.TPid,
			Loadid:               dcs.Loadid,
			Direction:            dcs.Direction,
			Tenant:               dcs.Tenant,
			Category:             dcs.Category,
			Account:              dcs.Account,
			Subject:              dcs.Subject,
			Runid:                dc.RunId,
			RunFilters:           dc.RunFilters,
			ReqTypeField:         dc.ReqTypeField,
			DirectionField:       dc.DirectionField,
			TenantField:          dc.TenantField,
			CategoryField:        dc.CategoryField,
			AccountField:         dc.AccountField,
			SubjectField:         dc.SubjectField,
			PddField:             dc.PddField,
			DestinationField:     dc.DestinationField,
			SetupTimeField:       dc.SetupTimeField,
			AnswerTimeField:      dc.AnswerTimeField,
			UsageField:           dc.UsageField,
			SupplierField:        dc.SupplierField,
			DisconnectCauseField: dc.DisconnectCauseField,
			CostField:            dc.CostField,
			RatedField:           dc.RatedField,
		})
	}
	if len(dcs.DerivedChargers) == 0 {
		result = append(result, TpDerivedCharger{
			Tpid:      dcs.TPid,
			Loadid:    dcs.Loadid,
			Direction: dcs.Direction,
			Tenant:    dcs.Tenant,
			Category:  dcs.Category,
			Account:   dcs.Account,
			Subject:   dcs.Subject,
		})
	}
	return
}

func APItoModelCdrStat(stats *utils.TPCdrStats) (result []TpCdrstat) {
	for _, st := range stats.CdrStats {
		ql, _ := strconv.Atoi(st.QueueLength)
		result = append(result, TpCdrstat{
			Tpid:             stats.TPid,
			Tag:              stats.CdrStatsId,
			QueueLength:      ql,
			TimeWindow:       st.TimeWindow,
			SaveInterval:     st.SaveInterval,
			Metrics:          st.Metrics,
			SetupInterval:    st.SetupInterval,
			Tors:             st.TORs,
			CdrHosts:         st.CdrHosts,
			CdrSources:       st.CdrSources,
			ReqTypes:         st.ReqTypes,
			Directions:       st.Directions,
			Tenants:          st.Tenants,
			Categories:       st.Categories,
			Accounts:         st.Accounts,
			Subjects:         st.Subjects,
			DestinationIds:   st.DestinationIds,
			PddInterval:      st.PddInterval,
			UsageInterval:    st.UsageInterval,
			Suppliers:        st.Suppliers,
			DisconnectCauses: st.DisconnectCauses,
			MediationRunids:  st.MediationRunIds,
			RatedAccounts:    st.RatedAccounts,
			RatedSubjects:    st.RatedSubjects,
			CostInterval:     st.CostInterval,
			ActionTriggers:   st.ActionTriggers,
		})
	}
	if len(stats.CdrStats) == 0 {
		result = append(result, TpCdrstat{
			Tpid: stats.TPid,
			Tag:  stats.CdrStatsId,
		})
	}
	return
}

func APItoModelAliases(attr *utils.TPAliases) (result []TpAlias) {
	for _, v := range attr.Values {
		result = append(result, TpAlias{
			Tpid:          attr.TPid,
			Direction:     attr.Direction,
			Tenant:        attr.Tenant,
			Category:      attr.Category,
			Account:       attr.Account,
			Subject:       attr.Subject,
			Context:       attr.Context,
			DestinationId: v.DestinationId,
			Target:        v.Target,
			Original:      v.Original,
			Alias:         v.Alias,
			Weight:        v.Weight,
		})
	}
	if len(attr.Values) == 0 {
		result = append(result, TpAlias{
			Tpid: attr.TPid,
		})
	}
	return
}

func APItoModelUsers(attr *utils.TPUsers) (result []TpUser) {
	for _, p := range attr.Profile {
		result = append(result, TpUser{
			Tpid:           attr.TPid,
			Tenant:         attr.Tenant,
			UserName:       attr.UserName,
			AttributeName:  p.AttrName,
			AttributeValue: p.AttrValue,
			Weight:         attr.Weight,
		})
	}
	if len(attr.Profile) == 0 {
		result = append(result, TpUser{
			Tpid: attr.TPid,
		})
	}
	return
}

func APItoResourceLimit(tpRL *utils.TPResourceLimit, timezone string) (rl *ResourceLimit, err error) {
	rl = &ResourceLimit{ID: tpRL.ID, Weight: tpRL.Weight, Filters: make([]*RequestFilter, len(tpRL.Filters)), Usage: make(map[string]*ResourceUsage)}
	for i, tpFltr := range tpRL.Filters {
		rf := &RequestFilter{Type: tpFltr.Type, FieldName: tpFltr.FieldName, Values: tpFltr.Values}
		if err := rf.CompileValues(); err != nil {
			return nil, err
		}
		rl.Filters[i] = rf
	}
	if rl.ActivationTime, err = utils.ParseTimeDetectLayout(tpRL.ActivationTime, timezone); err != nil {
		return nil, err
	}
	if rl.Limit, err = strconv.ParseFloat(tpRL.Limit, 64); err != nil {
		return nil, err
	}
	return rl, nil
}
