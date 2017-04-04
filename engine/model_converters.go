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

func APItoModelTiming(t *utils.ApierTPTiming) (result TpTiming) {
	return TpTiming{
		Tpid:      t.TPid,
		Tag:       t.ID,
		Years:     t.Years,
		Months:    t.Months,
		MonthDays: t.MonthDays,
		WeekDays:  t.WeekDays,
		Time:      t.Time,
	}
}

func APItoModelTimings(ts []*utils.ApierTPTiming) (result TpTimings) {
	for _, t := range ts {
		if t != nil {
			at := APItoModelTiming(t)
			result = append(result, at)
		}
	}
	return result
}

func APItoModelDestination(d *utils.TPDestination) (result TpDestinations) {
	if d != nil {
		for _, p := range d.Prefixes {
			result = append(result, TpDestination{
				Tpid:   d.TPid,
				Tag:    d.ID,
				Prefix: p,
			})
		}
		if len(d.Prefixes) == 0 {
			result = append(result, TpDestination{
				Tpid: d.TPid,
				Tag:  d.ID,
			})
		}
	}
	return
}

func APItoModelRate(r *utils.TPRate) (result TpRates) {
	if r != nil {
		for _, rs := range r.RateSlots {
			result = append(result, TpRate{
				Tpid:               r.TPid,
				Tag:                r.ID,
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
				Tag:  r.ID,
			})
		}
	}
	return
}

func APItoModelRates(rs []*utils.TPRate) (result TpRates) {
	for _, r := range rs {
		for _, sr := range APItoModelRate(r) {
			result = append(result, sr)
		}
	}
	return result
}

func APItoModelDestinationRate(d *utils.TPDestinationRate) (result TpDestinationRates) {
	if d != nil {
		for _, dr := range d.DestinationRates {
			result = append(result, TpDestinationRate{
				Tpid:             d.TPid,
				Tag:              d.ID,
				DestinationsTag:  dr.DestinationId,
				RatesTag:         dr.RateId,
				RoundingMethod:   dr.RoundingMethod,
				RoundingDecimals: dr.RoundingDecimals,
				MaxCost:          dr.MaxCost,
				MaxCostStrategy:  dr.MaxCostStrategy,
			})
		}
		if len(d.DestinationRates) == 0 {
			result = append(result, TpDestinationRate{
				Tpid: d.TPid,
				Tag:  d.ID,
			})
		}
	}
	return
}

func APItoModelDestinationRates(drs []*utils.TPDestinationRate) (result TpDestinationRates) {
	if drs != nil {
		for _, dr := range drs {
			for _, sdr := range APItoModelDestinationRate(dr) {
				result = append(result, sdr)
			}
		}
	}
	return result
}

func APItoModelRatingPlan(rp *utils.TPRatingPlan) (result TpRatingPlans) {
	if rp != nil {
		for _, rpb := range rp.RatingPlanBindings {
			result = append(result, TpRatingPlan{
				Tpid:         rp.TPid,
				Tag:          rp.ID,
				DestratesTag: rpb.DestinationRatesId,
				TimingTag:    rpb.TimingId,
				Weight:       rpb.Weight,
			})
		}
		if len(rp.RatingPlanBindings) == 0 {
			result = append(result, TpRatingPlan{
				Tpid: rp.TPid,
				Tag:  rp.ID,
			})
		}
	}
	return
}

func APItoModelRatingPlans(rps []*utils.TPRatingPlan) (result TpRatingPlans) {
	for _, rp := range rps {
		for _, srp := range APItoModelRatingPlan(rp) {
			result = append(result, srp)
		}
	}
	return result
}

func APItoModelRatingProfile(rp *utils.TPRatingProfile) (result TpRatingProfiles) {
	if rp != nil {
		for _, rpa := range rp.RatingPlanActivations {
			result = append(result, TpRatingProfile{
				Tpid:             rp.TPid,
				Loadid:           rp.LoadId,
				Direction:        rp.Direction,
				Tenant:           rp.Tenant,
				Category:         rp.Category,
				Subject:          rp.Subject,
				ActivationTime:   rpa.ActivationTime,
				RatingPlanTag:    rpa.RatingPlanId,
				FallbackSubjects: rpa.FallbackSubjects,
				CdrStatQueueIds:  rpa.CdrStatQueueIds,
			})
		}
		if len(rp.RatingPlanActivations) == 0 {
			result = append(result, TpRatingProfile{
				Tpid:      rp.TPid,
				Loadid:    rp.LoadId,
				Direction: rp.Direction,
				Tenant:    rp.Tenant,
				Category:  rp.Category,
				Subject:   rp.Subject,
			})
		}
	}
	return
}

func APItoModelRatingProfiles(rps []*utils.TPRatingProfile) (result TpRatingProfiles) {
	for _, rp := range rps {
		for _, srp := range APItoModelRatingProfile(rp) {
			result = append(result, srp)
		}
	}
	return result
}

func APItoModelLcrRule(lrs *utils.TPLcrRules) (result TpLcrRules) {
	if lrs != nil {
		for _, r := range lrs.Rules {
			result = append(result, TpLcrRule{
				Tpid:           lrs.TPid,
				Direction:      lrs.Direction,
				Tenant:         lrs.Tenant,
				Category:       lrs.Category,
				Account:        lrs.Account,
				Subject:        lrs.Subject,
				DestinationTag: r.DestinationId,
				RpCategory:     r.RpCategory,
				Strategy:       r.Strategy,
				StrategyParams: r.StrategyParams,
				ActivationTime: r.ActivationTime,
				Weight:         r.Weight,
			})
		}
		if len(lrs.Rules) == 0 {
			result = append(result, TpLcrRule{
				Tpid:      lrs.TPid,
				Direction: lrs.Direction,
				Tenant:    lrs.Tenant,
				Category:  lrs.Category,
				Account:   lrs.Account,
				Subject:   lrs.Subject,
			})
		}
	}
	return
}

func APItoModelLcrRules(ts []*utils.TPLcrRules) (result TpLcrRules) {
	for _, t := range ts {
		for _, st := range APItoModelLcrRule(t) {
			result = append(result, st)
		}
	}
	return result
}

func APItoModelAction(as *utils.TPActions) (result TpActions) {
	if as != nil {
		for _, a := range as.Actions {
			result = append(result, TpAction{
				Tpid:            as.TPid,
				Tag:             as.ID,
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
				Tag:  as.ID,
			})
		}
	}
	return
}

func APItoModelActions(as []*utils.TPActions) (result TpActions) {
	for _, a := range as {
		for _, sa := range APItoModelAction(a) {
			result = append(result, sa)
		}
	}
	return result
}

func APItoModelActionPlan(a *utils.TPActionPlan) (result TpActionPlans) {
	if a != nil {
		for _, ap := range a.ActionPlan {
			result = append(result, TpActionPlan{
				Tpid:       a.TPid,
				Tag:        a.ID,
				ActionsTag: ap.ActionsId,
				TimingTag:  ap.TimingId,
				Weight:     ap.Weight,
			})
		}
		if len(a.ActionPlan) == 0 {
			result = append(result, TpActionPlan{
				Tpid: a.TPid,
				Tag:  a.ID,
			})
		}
	}
	return
}

func APItoModelActionPlans(aps []*utils.TPActionPlan) (result TpActionPlans) {
	for _, ap := range aps {
		for _, sap := range APItoModelActionPlan(ap) {
			result = append(result, sap)
		}
	}
	return result
}

func APItoModelActionTrigger(ats *utils.TPActionTriggers) (result TpActionTriggers) {
	if ats != nil {
		for _, at := range ats.ActionTriggers {
			result = append(result, TpActionTrigger{
				Tpid:                   ats.TPid,
				Tag:                    ats.ID,
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
				Tag:  ats.ID,
			})
		}
	}
	return
}

func APItoModelActionTriggers(ts []*utils.TPActionTriggers) (result TpActionTriggers) {
	for _, t := range ts {
		for _, st := range APItoModelActionTrigger(t) {
			result = append(result, st)
		}
	}
	return result
}

func APItoModelAccountAction(aa *utils.TPAccountActions) *TpAccountAction {
	return &TpAccountAction{
		Tpid:              aa.TPid,
		Loadid:            aa.LoadId,
		Tenant:            aa.Tenant,
		Account:           aa.Account,
		ActionPlanTag:     aa.ActionPlanId,
		ActionTriggersTag: aa.ActionTriggersId,
		AllowNegative:     aa.AllowNegative,
		Disabled:          aa.Disabled,
	}
}

func APItoModelAccountActions(aas []*utils.TPAccountActions) (result TpAccountActions) {
	for _, aa := range aas {
		if aa != nil {
			result = append(result, *APItoModelAccountAction(aa))
		}
	}
	return result
}

func APItoModelSharedGroup(sgs *utils.TPSharedGroups) (result TpSharedGroups) {
	if sgs != nil {
		for _, sg := range sgs.SharedGroups {
			result = append(result, TpSharedGroup{
				Tpid:          sgs.TPid,
				Tag:           sgs.ID,
				Account:       sg.Account,
				Strategy:      sg.Strategy,
				RatingSubject: sg.RatingSubject,
			})
		}
		if len(sgs.SharedGroups) == 0 {
			result = append(result, TpSharedGroup{
				Tpid: sgs.TPid,
				Tag:  sgs.ID,
			})
		}
	}
	return
}

func APItoModelSharedGroups(sgs []*utils.TPSharedGroups) (result TpSharedGroups) {
	for _, sg := range sgs {
		for _, ssg := range APItoModelSharedGroup(sg) {
			result = append(result, ssg)
		}
	}
	return result
}

func APItoModelDerivedCharger(dcs *utils.TPDerivedChargers) (result TpDerivedChargers) {
	if dcs != nil {
		for _, dc := range dcs.DerivedChargers {
			result = append(result, TpDerivedCharger{
				Tpid:                 dcs.TPid,
				Loadid:               dcs.LoadId,
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
				Loadid:    dcs.LoadId,
				Direction: dcs.Direction,
				Tenant:    dcs.Tenant,
				Category:  dcs.Category,
				Account:   dcs.Account,
				Subject:   dcs.Subject,
			})
		}
	}
	return
}

func APItoModelDerivedChargers(dcs []*utils.TPDerivedChargers) (result TpDerivedChargers) {
	for _, dc := range dcs {
		for _, sdc := range APItoModelDerivedCharger(dc) {
			result = append(result, sdc)
		}
	}
	return result
}

func APItoModelCdrStat(css *utils.TPCdrStats) (result TpCdrStats) {
	if css != nil {
		for _, cs := range css.CdrStats {
			ql, _ := strconv.Atoi(cs.QueueLength)
			result = append(result, TpCdrstat{
				Tpid:             css.TPid,
				Tag:              css.ID,
				QueueLength:      ql,
				TimeWindow:       cs.TimeWindow,
				SaveInterval:     cs.SaveInterval,
				Metrics:          cs.Metrics,
				SetupInterval:    cs.SetupInterval,
				Tors:             cs.TORs,
				CdrHosts:         cs.CdrHosts,
				CdrSources:       cs.CdrSources,
				ReqTypes:         cs.ReqTypes,
				Directions:       cs.Directions,
				Tenants:          cs.Tenants,
				Categories:       cs.Categories,
				Accounts:         cs.Accounts,
				Subjects:         cs.Subjects,
				DestinationIds:   cs.DestinationIds,
				PddInterval:      cs.PddInterval,
				UsageInterval:    cs.UsageInterval,
				Suppliers:        cs.Suppliers,
				DisconnectCauses: cs.DisconnectCauses,
				MediationRunids:  cs.MediationRunIds,
				RatedAccounts:    cs.RatedAccounts,
				RatedSubjects:    cs.RatedSubjects,
				CostInterval:     cs.CostInterval,
				ActionTriggers:   cs.ActionTriggers,
			})
		}
		if len(css.CdrStats) == 0 {
			result = append(result, TpCdrstat{
				Tpid: css.TPid,
				Tag:  css.ID,
			})
		}
	}
	return
}

func APItoModelCdrStats(css []*utils.TPCdrStats) (result TpCdrStats) {
	for _, cs := range css {
		for _, scs := range APItoModelCdrStat(cs) {
			result = append(result, scs)
		}
	}
	return result
}

func APItoModelAliases(as *utils.TPAliases) (result TpAliases) {
	if as != nil {
		for _, v := range as.Values {
			result = append(result, TpAlias{
				Tpid:          as.TPid,
				Direction:     as.Direction,
				Tenant:        as.Tenant,
				Category:      as.Category,
				Account:       as.Account,
				Subject:       as.Subject,
				Context:       as.Context,
				DestinationId: v.DestinationId,
				Target:        v.Target,
				Original:      v.Original,
				Alias:         v.Alias,
				Weight:        v.Weight,
			})
		}
		if len(as.Values) == 0 {
			result = append(result, TpAlias{
				Tpid:      as.TPid,
				Direction: as.Direction,
				Tenant:    as.Tenant,
				Category:  as.Category,
				Account:   as.Account,
				Subject:   as.Subject,
				Context:   as.Context,
			})
		}
	}
	return
}

func APItoModelAliasesA(as []*utils.TPAliases) (result TpAliases) {
	for _, a := range as {
		for _, sa := range APItoModelAliases(a) {
			result = append(result, sa)
		}
	}
	return result
}

func APItoModelUsers(us *utils.TPUsers) (result TpUsers) {
	if us != nil {
		for _, p := range us.Profile {
			result = append(result, TpUser{
				Tpid:           us.TPid,
				Tenant:         us.Tenant,
				UserName:       us.UserName,
				Masked:         us.Masked,
				Weight:         us.Weight,
				AttributeName:  p.AttrName,
				AttributeValue: p.AttrValue,
			})
		}
		if len(us.Profile) == 0 {
			result = append(result, TpUser{
				Tpid:     us.TPid,
				Tenant:   us.Tenant,
				UserName: us.UserName,
				Masked:   us.Masked,
				Weight:   us.Weight,
			})
		}
	}
	return
}

func APItoModelUsersA(ts []*utils.TPUsers) (result TpUsers) {
	for _, t := range ts {
		for _, st := range APItoModelUsers(t) {
			result = append(result, st)
		}
	}
	return result
}

func APItoModelResourceLimit(rl *utils.TPResourceLimit) TpResourceLimits {
	result := TpResourceLimits{}
	for _, f := range rl.Filters {
		tp := &TpResourceLimit{
			Tpid:           rl.TPid,
			Tag:            rl.ID,
			ActivationTime: rl.ActivationTime,
			Weight:         rl.Weight,
			Limit:          rl.Limit,
		}
		for i, atid := range rl.ActionTriggerIDs {
			if i != 0 {
				tp.ActionTriggerIds = tp.ActionTriggerIds + utils.INFIELD_SEP + atid
			} else {
				tp.ActionTriggerIds = atid
			}
		}
		tp.FilterType = f.Type
		tp.FilterFieldName = f.FieldName
		for i, val := range f.Values {
			if i != 0 {
				tp.FilterFieldValues = tp.FilterFieldValues + utils.INFIELD_SEP + val
			} else {
				tp.FilterFieldValues = val
			}
		}
		result = append(result, tp)
	}
	if len(rl.Filters) == 0 {
		tp := &TpResourceLimit{
			Tpid:           rl.TPid,
			Tag:            rl.ID,
			ActivationTime: rl.ActivationTime,
			Weight:         rl.Weight,
			Limit:          rl.Limit,
		}
		for i, atid := range rl.ActionTriggerIDs {
			if i != 0 {
				tp.ActionTriggerIds = tp.ActionTriggerIds + utils.INFIELD_SEP + atid
			} else {
				tp.ActionTriggerIds = atid
			}
		}
		result = append(result, tp)
	}
	return result
}

func APItoResourceLimit(tpRL *utils.TPResourceLimit, timezone string) (rl *ResourceLimit, err error) {
	rl = &ResourceLimit{ID: tpRL.ID, Weight: tpRL.Weight, Filters: make([]*RequestFilter, len(tpRL.Filters)), Usage: make(map[string]*ResourceUsage)}
	for i, f := range tpRL.Filters {
		rf := &RequestFilter{Type: f.Type, FieldName: f.FieldName, Values: f.Values}
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
