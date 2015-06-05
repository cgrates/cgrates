package engine

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func csvLoad(s interface{}, values []string) (interface{}, error) {
	fieldValueMap := make(map[string]string)
	st := reflect.TypeOf(s)
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		re := field.Tag.Get("re")
		index := field.Tag.Get("index")
		if index != "" {
			idx, err := strconv.Atoi(index)
			if err != nil || len(values) <= idx {
				return nil, fmt.Errorf("invalid %v.%v index %v", st.Name(), field.Name, index)
			}
			if re != "" {
				if matched, err := regexp.MatchString(re, values[idx]); !matched || err != nil {
					return nil, fmt.Errorf("invalid %v.%v value %v", st.Name(), field.Name, values[idx])
				}
			}
			fieldValueMap[field.Name] = values[idx]
		}
	}
	elem := reflect.New(st).Elem()
	for fieldName, fieldValue := range fieldValueMap {
		field := elem.FieldByName(fieldName)
		if field.IsValid() {
			switch field.Kind() {
			case reflect.Float64:
				if fieldValue == "" {
					fieldValue = "0"
				}
				value, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetFloat(value)
			case reflect.Int:
				if fieldValue == "" {
					fieldValue = "0"
				}
				value, err := strconv.Atoi(fieldValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetInt(int64(value))
			case reflect.Bool:
				if fieldValue == "" {
					fieldValue = "false"
				}
				value, err := strconv.ParseBool(fieldValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetBool(value)
			case reflect.String:
				field.SetString(fieldValue)
			}
		}
	}
	return elem.Interface(), nil
}

func csvDump(s interface{}) ([]string, error) {
	fieldIndexMap := make(map[string]int)
	st := reflect.TypeOf(s)
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		index := field.Tag.Get("index")
		if index != "" {
			if idx, err := strconv.Atoi(index); err != nil {
				return nil, fmt.Errorf("invalid %v.%v index %v", st.Name(), field.Name, index)
			} else {
				fieldIndexMap[field.Name] = idx
			}
		}
	}
	elem := reflect.ValueOf(s)
	result := make([]string, len(fieldIndexMap))
	for fieldName, fieldIndex := range fieldIndexMap {
		field := elem.FieldByName(fieldName)
		if field.IsValid() && fieldIndex < len(result) {
			result[fieldIndex] = field.String()
		}
	}
	return result, nil
}

func getColumnCount(s interface{}) int {
	st := reflect.TypeOf(s)
	numFields := st.NumField()
	count := 0
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		index := field.Tag.Get("index")
		if index != "" {
			count++
		}
	}
	return count
}

type TpDestinations []TpDestination

func (tps TpDestinations) GetDestinations() (map[string]*Destination, error) {
	destinations := make(map[string]*Destination)
	for _, tp := range tps {
		var dest *Destination
		var found bool
		if dest, found = destinations[tp.Tag]; !found {
			dest = &Destination{Id: tp.Tag}
			destinations[tp.Tag] = dest
		}
		dest.AddPrefix(tp.Prefix)
	}
	return destinations, nil
}

type TpTimings []TpTiming

func (tps TpTimings) GetTimings() (map[string]*utils.TPTiming, error) {
	timings := make(map[string]*utils.TPTiming)
	for _, tp := range tps {
		rt := &utils.TPTiming{}
		rt.Id = tp.Tag
		rt.Years.Parse(tp.Years, utils.INFIELD_SEP)
		rt.Months.Parse(tp.Months, utils.INFIELD_SEP)
		rt.MonthDays.Parse(tp.MonthDays, utils.INFIELD_SEP)
		rt.WeekDays.Parse(tp.WeekDays, utils.INFIELD_SEP)
		times := strings.Split(tp.Time, utils.INFIELD_SEP)
		rt.StartTime = times[0]
		if len(times) > 1 {
			rt.EndTime = times[1]
		}

		if _, found := timings[tp.Tag]; found {
			return nil, fmt.Errorf("duplicate timing tag: %s", tp.Tag)
		}
		timings[tp.Tag] = rt
	}
	return timings, nil
}

type TpRates []TpRate

func (tps TpRates) GetRates() (map[string]*utils.TPRate, error) {
	rates := make(map[string]*utils.TPRate)
	for _, tp := range tps {

		rs, err := utils.NewRateSlot(tp.ConnectFee, tp.Rate, tp.RateUnit, tp.RateIncrement, tp.GroupIntervalStart)
		if err != nil {
			return nil, err
		}
		r := &utils.TPRate{
			RateId:    tp.Tag,
			RateSlots: []*utils.RateSlot{rs},
		}

		// same tag only to create rate groups
		_, exists := rates[tp.Tag]
		if exists {
			rates[tp.Tag].RateSlots = append(rates[tp.Tag].RateSlots, r.RateSlots[0])
		} else {
			rates[tp.Tag] = r
		}
	}
	return rates, nil
}

type TpDestinationRates []TpDestinationRate

func (tps TpDestinationRates) GetDestinationRates() (map[string]*utils.TPDestinationRate, error) {
	rts := make(map[string]*utils.TPDestinationRate)
	for _, tpDr := range tps {
		dr := &utils.TPDestinationRate{
			DestinationRateId: tpDr.Tag,
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId:    tpDr.DestinationsTag,
					RateId:           tpDr.RatesTag,
					RoundingMethod:   tpDr.RoundingMethod,
					RoundingDecimals: tpDr.RoundingDecimals,
					MaxCost:          tpDr.MaxCost,
					MaxCostStrategy:  tpDr.MaxCostStrategy,
				},
			},
		}
		existingDR, exists := rts[tpDr.Tag]
		if exists {
			existingDR.DestinationRates = append(existingDR.DestinationRates, dr.DestinationRates[0])
		} else {
			existingDR = dr
		}
		rts[tpDr.Tag] = existingDR

	}
	return rts, nil
}

type TpRatingPlans []TpRatingPlan

func (tps TpRatingPlans) GetRatingPlans() (map[string][]*utils.TPRatingPlanBinding, error) {
	rpbns := make(map[string][]*utils.TPRatingPlanBinding)

	for _, tpRp := range tps {
		rpb := &utils.TPRatingPlanBinding{
			DestinationRatesId: tpRp.DestratesTag,
			TimingId:           tpRp.TimingTag,
			Weight:             tpRp.Weight,
		}
		if _, exists := rpbns[tpRp.Tag]; exists {
			rpbns[tpRp.Tag] = append(rpbns[tpRp.Tag], rpb)
		} else { // New
			rpbns[tpRp.Tag] = []*utils.TPRatingPlanBinding{rpb}
		}
	}
	return rpbns, nil
}

func GetRateInterval(rpl *utils.TPRatingPlanBinding, dr *utils.DestinationRate) (i *RateInterval) {
	i = &RateInterval{
		Timing: &RITiming{
			Years:     rpl.Timing().Years,
			Months:    rpl.Timing().Months,
			MonthDays: rpl.Timing().MonthDays,
			WeekDays:  rpl.Timing().WeekDays,
			StartTime: rpl.Timing().StartTime,
		},
		Weight: rpl.Weight,
		Rating: &RIRate{
			ConnectFee:       dr.Rate.RateSlots[0].ConnectFee,
			RoundingMethod:   dr.RoundingMethod,
			RoundingDecimals: dr.RoundingDecimals,
			MaxCost:          dr.MaxCost,
			MaxCostStrategy:  dr.MaxCostStrategy,
		},
	}
	for _, rl := range dr.Rate.RateSlots {
		i.Rating.Rates = append(i.Rating.Rates, &Rate{
			GroupIntervalStart: rl.GroupIntervalStartDuration(),
			Value:              rl.Rate,
			RateIncrement:      rl.RateIncrementDuration(),
			RateUnit:           rl.RateUnitDuration(),
		})
	}
	return
}

type TpRatingProfiles []TpRatingProfile

func (tps TpRatingProfiles) GetRatingProfiles() (map[string]*utils.TPRatingProfile, error) {
	rpfs := make(map[string]*utils.TPRatingProfile)
	for _, tpRpf := range tps {

		rp := &utils.TPRatingProfile{
			TPid:      tpRpf.Tpid,
			LoadId:    tpRpf.Loadid,
			Direction: tpRpf.Direction,
			Tenant:    tpRpf.Tenant,
			Category:  tpRpf.Category,
			Subject:   tpRpf.Subject,
		}
		ra := &utils.TPRatingActivation{
			ActivationTime:   tpRpf.ActivationTime,
			RatingPlanId:     tpRpf.RatingPlanTag,
			FallbackSubjects: tpRpf.FallbackSubjects,
			CdrStatQueueIds:  tpRpf.CdrStatQueueIds,
		}
		if existingRpf, exists := rpfs[rp.KeyId()]; !exists {
			rp.RatingPlanActivations = []*utils.TPRatingActivation{ra}
			rpfs[rp.KeyId()] = rp
		} else { // Exists, update
			existingRpf.RatingPlanActivations = append(existingRpf.RatingPlanActivations, ra)
		}

	}
	return rpfs, nil
}

type TpSharedGroups []TpSharedGroup

func (tps TpSharedGroups) GetSharedGroups() (map[string][]*utils.TPSharedGroup, error) {
	sgs := make(map[string][]*utils.TPSharedGroup)
	for _, tpSg := range tps {
		sgs[tpSg.Tag] = append(sgs[tpSg.Tag], &utils.TPSharedGroup{
			Account:       tpSg.Account,
			Strategy:      tpSg.Strategy,
			RatingSubject: tpSg.RatingSubject,
		})
	}
	return sgs, nil
}

type TpActions []TpAction

func (tps TpActions) GetActions() (map[string][]*utils.TPAction, error) {
	as := make(map[string][]*utils.TPAction)
	for _, tpAc := range tps {
		a := &utils.TPAction{
			Identifier:      tpAc.Action,
			BalanceId:       tpAc.BalanceTag,
			BalanceType:     tpAc.BalanceType,
			Direction:       tpAc.Direction,
			Units:           tpAc.Units,
			ExpiryTime:      tpAc.ExpiryTime,
			TimingTags:      tpAc.TimingTags,
			DestinationIds:  tpAc.DestinationTags,
			RatingSubject:   tpAc.RatingSubject,
			Category:        tpAc.Category,
			SharedGroup:     tpAc.SharedGroup,
			BalanceWeight:   tpAc.BalanceWeight,
			ExtraParameters: tpAc.ExtraParameters,
			Weight:          tpAc.Weight,
		}
		as[tpAc.Tag] = append(as[tpAc.Tag], a)
	}

	return as, nil
}

type TpActionPlans []TpActionPlan

func (tps TpActionPlans) GetActionPlans() (map[string][]*utils.TPActionTiming, error) {
	ats := make(map[string][]*utils.TPActionTiming)
	for _, tpAp := range tps {
		ats[tpAp.Tag] = append(ats[tpAp.Tag], &utils.TPActionTiming{ActionsId: tpAp.ActionsTag, TimingId: tpAp.TimingTag, Weight: tpAp.Weight})
	}
	return ats, nil
}

type TpActionTriggers []TpActionTrigger

func (tps TpActionTriggers) GetActionTriggers() (map[string][]*utils.TPActionTrigger, error) {
	ats := make(map[string][]*utils.TPActionTrigger)
	for _, tpAt := range tps {
		at := &utils.TPActionTrigger{
			Id:                    tpAt.UniqueId,
			ThresholdType:         tpAt.ThresholdType,
			ThresholdValue:        tpAt.ThresholdValue,
			Recurrent:             tpAt.Recurrent,
			MinSleep:              tpAt.MinSleep,
			BalanceId:             tpAt.BalanceTag,
			BalanceType:           tpAt.BalanceType,
			BalanceDirection:      tpAt.BalanceDirection,
			BalanceDestinationIds: tpAt.BalanceDestinationTags,
			BalanceWeight:         tpAt.BalanceWeight,
			BalanceExpirationDate: tpAt.BalanceExpiryTime,
			BalanceTimingTags:     tpAt.BalanceTimingTags,
			BalanceRatingSubject:  tpAt.BalanceRatingSubject,
			BalanceCategory:       tpAt.BalanceCategory,
			BalanceSharedGroup:    tpAt.BalanceSharedGroup,
			Weight:                tpAt.Weight,
			ActionsId:             tpAt.ActionsTag,
			MinQueuedItems:        tpAt.MinQueuedItems,
		}
		ats[tpAt.Tag] = append(ats[tpAt.Tag], at)
	}
	return ats, nil
}

type TpAccountActions []TpAccountAction

func (tps TpAccountActions) GetAccountActions() (map[string]*utils.TPAccountActions, error) {
	aas := make(map[string]*utils.TPAccountActions)
	for _, tpAa := range tps {
		aacts := &utils.TPAccountActions{
			TPid:             tpAa.Tpid,
			LoadId:           tpAa.Loadid,
			Tenant:           tpAa.Tenant,
			Account:          tpAa.Account,
			Direction:        tpAa.Direction,
			ActionPlanId:     tpAa.ActionPlanTag,
			ActionTriggersId: tpAa.ActionTriggersTag,
		}
		aas[aacts.KeyId()] = aacts
	}
	return aas, nil
}

type TpDerivedChargers []TpDerivedCharger

func (tps TpDerivedChargers) GetDerivedChargers() (map[string]*utils.TPDerivedChargers, error) {
	dcs := make(map[string]*utils.TPDerivedChargers)
	for _, tpDcMdl := range tps {
		tpDc := &utils.TPDerivedChargers{TPid: tpDcMdl.Tpid, Loadid: tpDcMdl.Loadid, Direction: tpDcMdl.Direction, Tenant: tpDcMdl.Tenant, Category: tpDcMdl.Category,
			Account: tpDcMdl.Account, Subject: tpDcMdl.Subject}
		tag := tpDc.GetDerivedChargesId()
		if _, hasIt := dcs[tag]; !hasIt {
			dcs[tag] = tpDc
		}
		nDc := &utils.TPDerivedCharger{
			RunId:                ValueOrDefault(tpDcMdl.Runid, utils.META_DEFAULT),
			RunFilters:           tpDcMdl.RunFilters,
			ReqTypeField:         ValueOrDefault(tpDcMdl.ReqTypeField, utils.META_DEFAULT),
			DirectionField:       ValueOrDefault(tpDcMdl.DirectionField, utils.META_DEFAULT),
			TenantField:          ValueOrDefault(tpDcMdl.TenantField, utils.META_DEFAULT),
			CategoryField:        ValueOrDefault(tpDcMdl.CategoryField, utils.META_DEFAULT),
			AccountField:         ValueOrDefault(tpDcMdl.AccountField, utils.META_DEFAULT),
			SubjectField:         ValueOrDefault(tpDcMdl.SubjectField, utils.META_DEFAULT),
			DestinationField:     ValueOrDefault(tpDcMdl.DestinationField, utils.META_DEFAULT),
			SetupTimeField:       ValueOrDefault(tpDcMdl.SetupTimeField, utils.META_DEFAULT),
			AnswerTimeField:      ValueOrDefault(tpDcMdl.AnswerTimeField, utils.META_DEFAULT),
			UsageField:           ValueOrDefault(tpDcMdl.UsageField, utils.META_DEFAULT),
			SupplierField:        ValueOrDefault(tpDcMdl.SupplierField, utils.META_DEFAULT),
			DisconnectCauseField: ValueOrDefault(tpDcMdl.DisconnectCauseField, utils.META_DEFAULT),
		}
		dcs[tag].DerivedChargers = append(dcs[tag].DerivedChargers, nDc)
	}
	return dcs, nil
}

type TpCdrStats []TpCdrStat

func (tps TpCdrStats) GetCdrStats() (map[string][]*utils.TPCdrStat, error) {
	css := make(map[string][]*utils.TPCdrStat)
	for _, tpCs := range tps {
		css[tpCs.Tag] = append(css[tpCs.Tag], &utils.TPCdrStat{
			QueueLength:         strconv.Itoa(tpCs.QueueLength),
			TimeWindow:          tpCs.TimeWindow,
			Metrics:             tpCs.Metrics,
			SetupInterval:       tpCs.SetupInterval,
			TORs:                tpCs.Tors,
			CdrHosts:            tpCs.CdrHosts,
			CdrSources:          tpCs.CdrSources,
			ReqTypes:            tpCs.ReqTypes,
			Directions:          tpCs.Directions,
			Tenants:             tpCs.Tenants,
			Categories:          tpCs.Categories,
			Accounts:            tpCs.Accounts,
			Subjects:            tpCs.Subjects,
			DestinationPrefixes: tpCs.DestinationPrefixes,
			UsageInterval:       tpCs.UsageInterval,
			Suppliers:           tpCs.Suppliers,
			DisconnectCauses:    tpCs.DisconnectCauses,
			MediationRunIds:     tpCs.MediationRunids,
			RatedAccounts:       tpCs.RatedAccounts,
			RatedSubjects:       tpCs.RatedSubjects,
			CostInterval:        tpCs.CostInterval,
			ActionTriggers:      tpCs.ActionTriggers,
		})
	}
	return css, nil
}

func UpdateCdrStats(cs *CdrStats, triggers ActionTriggerPriotityList, tpCs *utils.TPCdrStat) {
	if tpCs.QueueLength != "" && tpCs.QueueLength != "0" {
		if qi, err := strconv.Atoi(tpCs.QueueLength); err == nil {
			cs.QueueLength = qi
		} else {
			log.Printf("Error parsing QueuedLength %v for cdrs stats %v", tpCs.QueueLength, cs.Id)
		}
	}
	if tpCs.TimeWindow != "" {
		if d, err := time.ParseDuration(tpCs.TimeWindow); err == nil {
			cs.TimeWindow = d
		} else {
			log.Printf("Error parsing TimeWindow %v for cdrs stats %v", tpCs.TimeWindow, cs.Id)
		}
	}
	if tpCs.Metrics != "" {
		cs.Metrics = append(cs.Metrics, tpCs.Metrics)
	}
	if tpCs.SetupInterval != "" {
		times := strings.Split(tpCs.SetupInterval, utils.INFIELD_SEP)
		if len(times) > 0 {
			if sTime, err := utils.ParseTimeDetectLayout(times[0]); err == nil {
				if len(cs.SetupInterval) < 1 {
					cs.SetupInterval = append(cs.SetupInterval, sTime)
				} else {
					cs.SetupInterval[0] = sTime
				}
			} else {
				log.Printf("Error parsing TimeWindow %v for cdrs stats %v", tpCs.SetupInterval, cs.Id)
			}
		}
		if len(times) > 1 {
			if eTime, err := utils.ParseTimeDetectLayout(times[1]); err == nil {
				if len(cs.SetupInterval) < 2 {
					cs.SetupInterval = append(cs.SetupInterval, eTime)
				} else {
					cs.SetupInterval[1] = eTime
				}
			} else {
				log.Printf("Error parsing TimeWindow %v for cdrs stats %v", tpCs.SetupInterval, cs.Id)
			}
		}
	}
	if tpCs.TORs != "" {
		cs.TOR = append(cs.TOR, tpCs.TORs)
	}
	if tpCs.CdrHosts != "" {
		cs.CdrHost = append(cs.CdrHost, tpCs.CdrHosts)
	}
	if tpCs.CdrSources != "" {
		cs.CdrSource = append(cs.CdrSource, tpCs.CdrSources)
	}
	if tpCs.ReqTypes != "" {
		cs.ReqType = append(cs.ReqType, tpCs.ReqTypes)
	}
	if tpCs.Directions != "" {
		cs.Direction = append(cs.Direction, tpCs.Directions)
	}
	if tpCs.Tenants != "" {
		cs.Tenant = append(cs.Tenant, tpCs.Tenants)
	}
	if tpCs.Categories != "" {
		cs.Category = append(cs.Category, tpCs.Categories)
	}
	if tpCs.Accounts != "" {
		cs.Account = append(cs.Account, tpCs.Accounts)
	}
	if tpCs.Subjects != "" {
		cs.Subject = append(cs.Subject, tpCs.Subjects)
	}
	if tpCs.DestinationPrefixes != "" {
		cs.DestinationPrefix = append(cs.DestinationPrefix, tpCs.DestinationPrefixes)
	}
	if tpCs.UsageInterval != "" {
		durations := strings.Split(tpCs.UsageInterval, utils.INFIELD_SEP)
		if len(durations) > 0 {
			if sDuration, err := time.ParseDuration(durations[0]); err == nil {
				if len(cs.UsageInterval) < 1 {
					cs.UsageInterval = append(cs.UsageInterval, sDuration)
				} else {
					cs.UsageInterval[0] = sDuration
				}
			} else {
				log.Printf("Error parsing UsageInterval %v for cdrs stats %v", tpCs.UsageInterval, cs.Id)
			}
		}
		if len(durations) > 1 {
			if eDuration, err := time.ParseDuration(durations[1]); err == nil {
				if len(cs.UsageInterval) < 2 {
					cs.UsageInterval = append(cs.UsageInterval, eDuration)
				} else {
					cs.UsageInterval[1] = eDuration
				}
			} else {
				log.Printf("Error parsing UsageInterval %v for cdrs stats %v", tpCs.UsageInterval, cs.Id)
			}
		}
	}
	if tpCs.Suppliers != "" {
		cs.Supplier = append(cs.Supplier, tpCs.Suppliers)
	}
	if tpCs.DisconnectCauses != "" {
		cs.DisconnectCause = append(cs.DisconnectCause, tpCs.DisconnectCauses)
	}
	if tpCs.MediationRunIds != "" {
		cs.MediationRunIds = append(cs.MediationRunIds, tpCs.MediationRunIds)
	}
	if tpCs.RatedAccounts != "" {
		cs.RatedAccount = append(cs.RatedAccount, tpCs.RatedAccounts)
	}
	if tpCs.RatedSubjects != "" {
		cs.RatedSubject = append(cs.RatedSubject, tpCs.RatedSubjects)
	}
	if tpCs.CostInterval != "" {
		costs := strings.Split(tpCs.CostInterval, utils.INFIELD_SEP)
		if len(costs) > 0 {
			if sCost, err := strconv.ParseFloat(costs[0], 64); err == nil {
				if len(cs.CostInterval) < 1 {
					cs.CostInterval = append(cs.CostInterval, sCost)
				} else {
					cs.CostInterval[0] = sCost
				}
			} else {
				log.Printf("Error parsing CostInterval %v for cdrs stats %v", tpCs.CostInterval, cs.Id)
			}
		}
		if len(costs) > 1 {
			if eCost, err := strconv.ParseFloat(costs[1], 64); err == nil {
				if len(cs.CostInterval) < 2 {
					cs.CostInterval = append(cs.CostInterval, eCost)
				} else {
					cs.CostInterval[1] = eCost
				}
			} else {
				log.Printf("Error parsing CostInterval %v for cdrs stats %v", tpCs.CostInterval, cs.Id)
			}
		}
	}
	if triggers != nil {
		cs.Triggers = append(cs.Triggers, triggers...)
	}
}

// ValueOrDefault is used to populate empty values with *any or *default if value missing
func ValueOrDefault(val string, deflt string) string {
	if len(val) == 0 {
		val = deflt
	}
	return val
}
