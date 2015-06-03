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
func APItoModelDestination(dest *utils.TPDestination) (result []TpDestination) {
	for _, p := range dest.Prefixes {
		result = append(result, TpDestination{
			Tpid:   dest.TPid,
			Tag:    dest.DestinationId,
			Prefix: p,
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
	return
}

func APItoModelLcrRule(lcrs *utils.TPLcrRules) (result []TpLcrRule) {
	for _, lcr := range lcrs.LcrRules {
		result = append(result, TpLcrRule{
			Tpid:           lcrs.TPid,
			Direction:      lcr.Direction,
			Tenant:         lcr.Tenant,
			Category:       lcr.Category,
			Account:        lcr.Account,
			Subject:        lcr.Subject,
			DestinationTag: lcr.DestinationId,
			RpCategory:     lcr.RpCategory,
			Strategy:       lcr.Strategy,
			StrategyParams: lcr.StrategyParams,
			ActivationTime: lcr.ActivationTime,
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
			Direction:       a.Direction,
			Units:           a.Units,
			ExpiryTime:      a.ExpiryTime,
			TimingTags:      a.TimingTags,
			DestinationTags: a.DestinationIds,
			RatingSubject:   a.RatingSubject,
			Category:        a.Category,
			SharedGroup:     a.SharedGroup,
			BalanceWeight:   a.BalanceWeight,
			ExtraParameters: a.ExtraParameters,
			Weight:          a.Weight,
		})
	}
	return
}

func APItoModelActionPlan(aps *utils.TPActionPlan) (result []TpActionPlan) {
	for _, ap := range aps.ActionPlan {
		result = append(result, TpActionPlan{
			Tpid:       aps.TPid,
			Tag:        aps.Id,
			ActionsTag: ap.ActionsId,
			TimingTag:  ap.TimingId,
			Weight:     ap.Weight,
		})
	}
	return
}

func APItoModelActionTrigger(ats *utils.TPActionTriggers) (result []TpActionTrigger) {
	for _, at := range ats.ActionTriggers {
		result = append(result, TpActionTrigger{
			Tpid:                   ats.TPid,
			Tag:                    ats.ActionTriggersId,
			UniqueId:               at.Id,
			ThresholdType:          at.ThresholdType,
			ThresholdValue:         at.ThresholdValue,
			Recurrent:              at.Recurrent,
			MinSleep:               at.MinSleep,
			BalanceTag:             at.BalanceId,
			BalanceType:            at.BalanceType,
			BalanceDirection:       at.BalanceDirection,
			BalanceDestinationTags: at.BalanceDestinationIds,
			BalanceWeight:          at.BalanceWeight,
			BalanceExpiryTime:      at.BalanceExpirationDate,
			BalanceTimingTags:      at.BalanceTimingTags,
			BalanceRatingSubject:   at.BalanceRatingSubject,
			BalanceCategory:        at.BalanceCategory,
			BalanceSharedGroup:     at.BalanceSharedGroup,
			MinQueuedItems:         at.MinQueuedItems,
			ActionsTag:             at.ActionsId,
			Weight:                 at.Weight,
		})
	}
	return
}

func APItoModelAccountAction(aa *utils.TPAccountActions) *TpAccountAction {
	return &TpAccountAction{
		Tpid:              aa.TPid,
		Loadid:            aa.LoadId,
		Direction:         aa.Direction,
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
			DestinationField:     dc.DestinationField,
			SetupTimeField:       dc.SetupTimeField,
			AnswerTimeField:      dc.AnswerTimeField,
			UsageField:           dc.UsageField,
			SupplierField:        dc.SupplierField,
			DisconnectCauseField: dc.DisconnectCauseField,
		})
	}
	return
}

func APItoModelCdrStat(stats *utils.TPCdrStats) (result []TpCdrStat) {
	for _, st := range stats.CdrStats {
		ql, _ := strconv.Atoi(st.QueueLength)
		result = append(result, TpCdrStat{
			Tpid:                stats.TPid,
			Tag:                 stats.CdrStatsId,
			QueueLength:         ql,
			TimeWindow:          st.TimeWindow,
			Metrics:             st.Metrics,
			SetupInterval:       st.SetupInterval,
			Tors:                st.TORs,
			CdrHosts:            st.CdrHosts,
			CdrSources:          st.CdrSources,
			ReqTypes:            st.ReqTypes,
			Directions:          st.Directions,
			Tenants:             st.Tenants,
			Categories:          st.Categories,
			Accounts:            st.Accounts,
			Subjects:            st.Subjects,
			DestinationPrefixes: st.DestinationPrefixes,
			UsageInterval:       st.UsageInterval,
			Suppliers:           st.Suppliers,
			DisconnectCauses:    st.DisconnectCauses,
			MediationRunids:     st.MediationRunIds,
			RatedAccounts:       st.RatedAccounts,
			RatedSubjects:       st.RatedSubjects,
			CostInterval:        st.CostInterval,
			ActionTriggers:      st.ActionTriggers,
		})
	}
	return
}
