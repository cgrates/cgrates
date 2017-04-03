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
			switch field.Kind() {
			case reflect.Float64:
				result[fieldIndex] = strconv.FormatFloat(field.Float(), 'f', -1, 64)
			case reflect.Int:
				result[fieldIndex] = strconv.FormatInt(field.Int(), 10)
			case reflect.Bool:
				result[fieldIndex] = strconv.FormatBool(field.Bool())
			case reflect.String:
				result[fieldIndex] = field.String()
			}
		}
	}
	return result, nil
}

func modelEqual(this interface{}, other interface{}) bool {
	var fieldNames []string
	st := reflect.TypeOf(this)
	stO := reflect.TypeOf(other)
	if st != stO {
		return false
	}
	numFields := st.NumField()
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		index := field.Tag.Get("index")
		if index != "" {
			fieldNames = append(fieldNames, field.Name)
		}
	}
	thisElem := reflect.ValueOf(this)
	otherElem := reflect.ValueOf(other)
	for _, fieldName := range fieldNames {
		thisField := thisElem.FieldByName(fieldName)
		otherField := otherElem.FieldByName(fieldName)
		switch thisField.Kind() {
		case reflect.Float64:
			if thisField.Float() != otherField.Float() {
				return false
			}
		case reflect.Int:
			if thisField.Int() != otherField.Int() {
				return false
			}
		case reflect.Bool:
			if thisField.Bool() != otherField.Bool() {
				return false
			}
		case reflect.String:
			if thisField.String() != otherField.String() {
				return false
			}
		}

	}
	return true
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

// AsTPDestination converts TpDestinations into *utils.TPDestination
func (tps TpDestinations) AsTPDestinations() (tpDsts []*utils.TPDestination) {
	uTPDestsMp := make(map[string]*utils.TPDestination) // Should save us some CPU if we index here for big number of destinations to search
	for _, tpDt := range tps {
		if uTPDst, hasIt := uTPDestsMp[tpDt.Tag]; !hasIt {
			uTPDestsMp[tpDt.Tag] = &utils.TPDestination{TPid: tpDt.Tpid, ID: tpDt.Tag, Prefixes: []string{tpDt.Prefix}}
		} else {
			uTPDst.Prefixes = append(uTPDst.Prefixes, tpDt.Prefix)
		}
	}
	tpDsts = make([]*utils.TPDestination, len(uTPDestsMp))
	i := 0
	for _, uTPDest := range uTPDestsMp {
		tpDsts[i] = uTPDest
		i++
	}
	return
}

type TpTimings []TpTiming

func (tps TpTimings) GetTimings() (map[string]*utils.TPTiming, error) {
	timings := make(map[string]*utils.TPTiming)
	for _, tp := range tps {
		rt := &utils.TPTiming{}
		rt.ID = tp.Tag
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

func (tps TpTimings) GetApierTimings() (map[string]*utils.ApierTPTiming, error) {
	timings := make(map[string]*utils.ApierTPTiming)
	for _, tp := range tps {
		rt := &utils.ApierTPTiming{
			TPid:      tp.Tpid,
			ID:        tp.Tag,
			Years:     tp.Years,
			Months:    tp.Months,
			MonthDays: tp.MonthDays,
			WeekDays:  tp.WeekDays,
			Time:      tp.Time,
		}
		timings[tp.Tag] = rt
	}
	return timings, nil
}

func (tps TpTimings) AsTPTimings() (result []*utils.ApierTPTiming) {
	ats, _ := tps.GetApierTimings()
	for _, tp := range ats {
		result = append(result, tp)
	}
	return result
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
			TPid:      tp.Tpid,
			ID:        tp.Tag,
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

func (tps TpRates) GetRatesA() (map[string]*utils.TPRate, error) {
	rpfs := make(map[string]*utils.TPRate)
	for _, tpRpf := range tps {
		rp := &utils.TPRate{
			TPid: tpRpf.Tpid,
			ID:   tpRpf.Tag,
		}
		ra := &utils.RateSlot{
			ConnectFee:         tpRpf.ConnectFee,
			Rate:               tpRpf.Rate,
			RateUnit:           tpRpf.RateUnit,
			RateIncrement:      tpRpf.RateIncrement,
			GroupIntervalStart: tpRpf.GroupIntervalStart,
		}
		if existingRpf, exists := rpfs[rp.ID]; !exists {
			rp.RateSlots = []*utils.RateSlot{ra}
			rpfs[rp.ID] = rp
		} else { // Exists, update
			existingRpf.RateSlots = append(existingRpf.RateSlots, ra)
		}

	}
	return rpfs, nil
}

func (tps TpRates) AsTPRates() (result []*utils.TPRate, err error) {
	if utps, err := tps.GetRatesA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpDestinationRates []TpDestinationRate

func (tps TpDestinationRates) GetDestinationRates() (map[string]*utils.TPDestinationRate, error) {
	rts := make(map[string]*utils.TPDestinationRate)
	for _, tpDr := range tps {
		dr := &utils.TPDestinationRate{
			TPid: tpDr.Tpid,
			ID:   tpDr.Tag,
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

func (tps TpDestinationRates) AsTPDestinationRates() (result []*utils.TPDestinationRate, err error) {
	if utps, err := tps.GetDestinationRates(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
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

func (tps TpRatingPlans) GetRatingPlansA() (map[string]*utils.TPRatingPlan, error) {
	rps := make(map[string]*utils.TPRatingPlan)
	for _, tp := range tps {
		rp := &utils.TPRatingPlan{
			TPid: tp.Tpid,
			ID:   tp.Tag,
		}
		rpb := &utils.TPRatingPlanBinding{
			DestinationRatesId: tp.DestratesTag,
			TimingId:           tp.TimingTag,
			Weight:             tp.Weight,
		}
		if existingRpf, exists := rps[rp.ID]; !exists {
			rp.RatingPlanBindings = []*utils.TPRatingPlanBinding{rpb}
			rps[rp.ID] = rp
		} else {
			existingRpf.RatingPlanBindings = append(existingRpf.RatingPlanBindings, rpb)
		}

	}
	return rps, nil
}

func (tps TpRatingPlans) AsTPRatingPlans() (result []*utils.TPRatingPlan, err error) {
	if utps, err := tps.GetRatingPlansA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func GetRateInterval(rpl *utils.TPRatingPlanBinding, dr *utils.DestinationRate) (i *RateInterval) {
	i = &RateInterval{
		Timing: &RITiming{
			Years:     rpl.Timing().Years,
			Months:    rpl.Timing().Months,
			MonthDays: rpl.Timing().MonthDays,
			WeekDays:  rpl.Timing().WeekDays,
			StartTime: rpl.Timing().StartTime,
			tag:       rpl.Timing().ID,
		},
		Weight: rpl.Weight,
		Rating: &RIRate{
			ConnectFee:       dr.Rate.RateSlots[0].ConnectFee,
			RoundingMethod:   dr.RoundingMethod,
			RoundingDecimals: dr.RoundingDecimals,
			MaxCost:          dr.MaxCost,
			MaxCostStrategy:  dr.MaxCostStrategy,
			tag:              dr.Rate.ID,
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
		if existingRpf, exists := rpfs[rp.KeyIdA()]; !exists {
			rp.RatingPlanActivations = []*utils.TPRatingActivation{ra}
			rpfs[rp.KeyIdA()] = rp
		} else { // Exists, update
			existingRpf.RatingPlanActivations = append(existingRpf.RatingPlanActivations, ra)
		}

	}
	return rpfs, nil
}

func (tps TpRatingProfiles) AsTPRatingProfiles() (result []*utils.TPRatingProfile, err error) {
	if utps, err := tps.GetRatingProfiles(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
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

func (tps TpSharedGroups) GetSharedGroupsA() (map[string]*utils.TPSharedGroups, error) {
	result := make(map[string]*utils.TPSharedGroups)
	for _, tp := range tps {
		sgs := &utils.TPSharedGroups{
			TPid: tp.Tpid,
			ID:   tp.Tag,
		}
		sg := &utils.TPSharedGroup{
			Account:       tp.Account,
			Strategy:      tp.Strategy,
			RatingSubject: tp.RatingSubject,
		}
		if existing, exists := result[sgs.ID]; !exists {
			sgs.SharedGroups = []*utils.TPSharedGroup{sg}
			result[sgs.ID] = sgs
		} else {
			existing.SharedGroups = append(existing.SharedGroups, sg)
		}
	}
	return result, nil
}

func (tps TpSharedGroups) AsTPSharedGroups() (result []*utils.TPSharedGroups, err error) {
	if utps, err := tps.GetSharedGroupsA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpActions []TpAction

func (tps TpActions) GetActions() (map[string][]*utils.TPAction, error) {
	as := make(map[string][]*utils.TPAction)
	for _, tpAc := range tps {
		a := &utils.TPAction{
			Identifier:      tpAc.Action,
			BalanceId:       tpAc.BalanceTag,
			BalanceType:     tpAc.BalanceType,
			Directions:      tpAc.Directions,
			Units:           tpAc.Units,
			ExpiryTime:      tpAc.ExpiryTime,
			Filter:          tpAc.Filter,
			TimingTags:      tpAc.TimingTags,
			DestinationIds:  tpAc.DestinationTags,
			RatingSubject:   tpAc.RatingSubject,
			Categories:      tpAc.Categories,
			SharedGroups:    tpAc.SharedGroups,
			BalanceWeight:   tpAc.BalanceWeight,
			BalanceBlocker:  tpAc.BalanceBlocker,
			BalanceDisabled: tpAc.BalanceDisabled,
			ExtraParameters: tpAc.ExtraParameters,
			Weight:          tpAc.Weight,
		}
		as[tpAc.Tag] = append(as[tpAc.Tag], a)
	}

	return as, nil
}

func (tps TpActions) GetActionsA() (map[string]*utils.TPActions, error) {
	result := make(map[string]*utils.TPActions)
	for _, tp := range tps {
		as := &utils.TPActions{
			TPid: tp.Tpid,
			ID:   tp.Tag,
		}
		a := &utils.TPAction{
			Identifier:      tp.Action,
			BalanceId:       tp.BalanceTag,
			BalanceType:     tp.BalanceType,
			Directions:      tp.Directions,
			Units:           tp.Units,
			ExpiryTime:      tp.ExpiryTime,
			Filter:          tp.Filter,
			TimingTags:      tp.TimingTags,
			DestinationIds:  tp.DestinationTags,
			RatingSubject:   tp.RatingSubject,
			Categories:      tp.Categories,
			SharedGroups:    tp.SharedGroups,
			BalanceWeight:   tp.BalanceWeight,
			BalanceBlocker:  tp.BalanceBlocker,
			BalanceDisabled: tp.BalanceDisabled,
			ExtraParameters: tp.ExtraParameters,
			Weight:          tp.Weight,
		}
		if existing, exists := result[as.ID]; !exists {
			as.Actions = []*utils.TPAction{a}
			result[as.ID] = as
		} else {
			existing.Actions = append(existing.Actions, a)
		}
	}
	return result, nil
}

func (tps TpActions) AsTPActions() (result []*utils.TPActions, err error) {
	if utps, err := tps.GetActionsA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpActionPlans []TpActionPlan

func (tps TpActionPlans) GetActionPlans() (map[string][]*utils.TPActionTiming, error) {
	ats := make(map[string][]*utils.TPActionTiming)
	for _, tpAp := range tps {
		ats[tpAp.Tag] = append(ats[tpAp.Tag], &utils.TPActionTiming{ActionsId: tpAp.ActionsTag, TimingId: tpAp.TimingTag, Weight: tpAp.Weight})
	}
	return ats, nil
}

func (tps TpActionPlans) GetActionPlansA() (map[string]*utils.TPActionPlan, error) {
	result := make(map[string]*utils.TPActionPlan)
	for _, tp := range tps {
		as := &utils.TPActionPlan{
			TPid: tp.Tpid,
			ID:   tp.Tag,
		}
		a := &utils.TPActionTiming{
			ActionsId: tp.ActionsTag,
			TimingId:  tp.TimingTag,
			Weight:    tp.Weight,
		}
		if existing, exists := result[as.ID]; !exists {
			as.ActionPlan = []*utils.TPActionTiming{a}
			result[as.ID] = as
		} else {
			existing.ActionPlan = append(existing.ActionPlan, a)
		}
	}
	return result, nil
}

func (tps TpActionPlans) AsTPActionPlans() (result []*utils.TPActionPlan, err error) {
	if utps, err := tps.GetActionPlansA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpActionTriggers []TpActionTrigger

func (tps TpActionTriggers) GetActionTriggers() (map[string][]*utils.TPActionTrigger, error) {
	ats := make(map[string][]*utils.TPActionTrigger)
	for _, tpAt := range tps {
		at := &utils.TPActionTrigger{
			Id:                    tpAt.Tag,
			UniqueID:              tpAt.UniqueId,
			ThresholdType:         tpAt.ThresholdType,
			ThresholdValue:        tpAt.ThresholdValue,
			Recurrent:             tpAt.Recurrent,
			MinSleep:              tpAt.MinSleep,
			ExpirationDate:        tpAt.ExpiryTime,
			ActivationDate:        tpAt.ActivationTime,
			BalanceId:             tpAt.BalanceTag,
			BalanceType:           tpAt.BalanceType,
			BalanceDirections:     tpAt.BalanceDirections,
			BalanceDestinationIds: tpAt.BalanceDestinationTags,
			BalanceWeight:         tpAt.BalanceWeight,
			BalanceExpirationDate: tpAt.BalanceExpiryTime,
			BalanceTimingTags:     tpAt.BalanceTimingTags,
			BalanceRatingSubject:  tpAt.BalanceRatingSubject,
			BalanceCategories:     tpAt.BalanceCategories,
			BalanceSharedGroups:   tpAt.BalanceSharedGroups,
			BalanceBlocker:        tpAt.BalanceBlocker,
			BalanceDisabled:       tpAt.BalanceDisabled,
			Weight:                tpAt.Weight,
			ActionsId:             tpAt.ActionsTag,
			MinQueuedItems:        tpAt.MinQueuedItems,
		}
		ats[tpAt.Tag] = append(ats[tpAt.Tag], at)
	}
	return ats, nil
}

func (tps TpActionTriggers) GetActionTriggersA() (map[string]*utils.TPActionTriggers, error) {
	result := make(map[string]*utils.TPActionTriggers)
	for _, tp := range tps {
		ats := &utils.TPActionTriggers{
			TPid: tp.Tpid,
			ID:   tp.Tag,
		}
		at := &utils.TPActionTrigger{
			Id:                    tp.Tag,
			UniqueID:              tp.UniqueId,
			ThresholdType:         tp.ThresholdType,
			ThresholdValue:        tp.ThresholdValue,
			Recurrent:             tp.Recurrent,
			MinSleep:              tp.MinSleep,
			ExpirationDate:        tp.ExpiryTime,
			ActivationDate:        tp.ActivationTime,
			BalanceId:             tp.BalanceTag,
			BalanceType:           tp.BalanceType,
			BalanceDirections:     tp.BalanceDirections,
			BalanceDestinationIds: tp.BalanceDestinationTags,
			BalanceWeight:         tp.BalanceWeight,
			BalanceExpirationDate: tp.BalanceExpiryTime,
			BalanceTimingTags:     tp.BalanceTimingTags,
			BalanceRatingSubject:  tp.BalanceRatingSubject,
			BalanceCategories:     tp.BalanceCategories,
			BalanceSharedGroups:   tp.BalanceSharedGroups,
			BalanceBlocker:        tp.BalanceBlocker,
			BalanceDisabled:       tp.BalanceDisabled,
			Weight:                tp.Weight,
			ActionsId:             tp.ActionsTag,
			MinQueuedItems:        tp.MinQueuedItems,
		}
		if existing, exists := result[ats.ID]; !exists {
			ats.ActionTriggers = []*utils.TPActionTrigger{at}
			result[ats.ID] = ats
		} else {
			existing.ActionTriggers = append(existing.ActionTriggers, at)
		}
	}
	return result, nil
}

func (tps TpActionTriggers) AsTPActionTriggers() (result []*utils.TPActionTriggers, err error) {
	if utps, err := tps.GetActionTriggersA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
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
			ActionPlanId:     tpAa.ActionPlanTag,
			ActionTriggersId: tpAa.ActionTriggersTag,
			AllowNegative:    tpAa.AllowNegative,
			Disabled:         tpAa.Disabled,
		}
		aas[aacts.KeyId()] = aacts
	}
	return aas, nil
}

func (tps TpAccountActions) AsTPAccountActions() (result []*utils.TPAccountActions, err error) {
	if utps, err := tps.GetAccountActions(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpDerivedChargers []TpDerivedCharger

func (tps TpDerivedChargers) GetDerivedChargers() (map[string]*utils.TPDerivedChargers, error) {
	dcs := make(map[string]*utils.TPDerivedChargers)
	for _, tpDcMdl := range tps {
		tpDc := &utils.TPDerivedChargers{TPid: tpDcMdl.Tpid, LoadId: tpDcMdl.Loadid, Direction: tpDcMdl.Direction, Tenant: tpDcMdl.Tenant, Category: tpDcMdl.Category,
			Account: tpDcMdl.Account, Subject: tpDcMdl.Subject, DestinationIds: tpDcMdl.DestinationIds}
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
			PddField:             ValueOrDefault(tpDcMdl.PddField, utils.META_DEFAULT),
			AnswerTimeField:      ValueOrDefault(tpDcMdl.AnswerTimeField, utils.META_DEFAULT),
			UsageField:           ValueOrDefault(tpDcMdl.UsageField, utils.META_DEFAULT),
			SupplierField:        ValueOrDefault(tpDcMdl.SupplierField, utils.META_DEFAULT),
			DisconnectCauseField: ValueOrDefault(tpDcMdl.DisconnectCauseField, utils.META_DEFAULT),
			CostField:            ValueOrDefault(tpDcMdl.CostField, utils.META_DEFAULT),
			RatedField:           ValueOrDefault(tpDcMdl.RatedField, utils.META_DEFAULT),
		}
		dcs[tag].DerivedChargers = append(dcs[tag].DerivedChargers, nDc)
	}
	return dcs, nil
}

func (tps TpDerivedChargers) AsTPDerivedChargers() (result []*utils.TPDerivedChargers, err error) {
	if utps, err := tps.GetDerivedChargers(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpCdrStats []TpCdrstat

func (tps TpCdrStats) GetCdrStats() (map[string][]*utils.TPCdrStat, error) {
	css := make(map[string][]*utils.TPCdrStat)
	for _, tpCs := range tps {
		css[tpCs.Tag] = append(css[tpCs.Tag], &utils.TPCdrStat{
			QueueLength:      strconv.Itoa(tpCs.QueueLength),
			TimeWindow:       tpCs.TimeWindow,
			Metrics:          tpCs.Metrics,
			SaveInterval:     tpCs.SaveInterval,
			SetupInterval:    tpCs.SetupInterval,
			TORs:             tpCs.Tors,
			CdrHosts:         tpCs.CdrHosts,
			CdrSources:       tpCs.CdrSources,
			ReqTypes:         tpCs.ReqTypes,
			Directions:       tpCs.Directions,
			Tenants:          tpCs.Tenants,
			Categories:       tpCs.Categories,
			Accounts:         tpCs.Accounts,
			Subjects:         tpCs.Subjects,
			DestinationIds:   tpCs.DestinationIds,
			PddInterval:      tpCs.PddInterval,
			UsageInterval:    tpCs.UsageInterval,
			Suppliers:        tpCs.Suppliers,
			DisconnectCauses: tpCs.DisconnectCauses,
			MediationRunIds:  tpCs.MediationRunids,
			RatedAccounts:    tpCs.RatedAccounts,
			RatedSubjects:    tpCs.RatedSubjects,
			CostInterval:     tpCs.CostInterval,
			ActionTriggers:   tpCs.ActionTriggers,
		})
	}
	return css, nil
}

func (tps TpCdrStats) GetCdrStatsA() (map[string]*utils.TPCdrStats, error) {
	result := make(map[string]*utils.TPCdrStats)
	for _, tp := range tps {
		css := &utils.TPCdrStats{
			TPid: tp.Tpid,
			ID:   tp.Tag,
		}
		cs := &utils.TPCdrStat{
			QueueLength:      strconv.Itoa(tp.QueueLength),
			TimeWindow:       tp.TimeWindow,
			Metrics:          tp.Metrics,
			SaveInterval:     tp.SaveInterval,
			SetupInterval:    tp.SetupInterval,
			TORs:             tp.Tors,
			CdrHosts:         tp.CdrHosts,
			CdrSources:       tp.CdrSources,
			ReqTypes:         tp.ReqTypes,
			Directions:       tp.Directions,
			Tenants:          tp.Tenants,
			Categories:       tp.Categories,
			Accounts:         tp.Accounts,
			Subjects:         tp.Subjects,
			DestinationIds:   tp.DestinationIds,
			PddInterval:      tp.PddInterval,
			UsageInterval:    tp.UsageInterval,
			Suppliers:        tp.Suppliers,
			DisconnectCauses: tp.DisconnectCauses,
			MediationRunIds:  tp.MediationRunids,
			RatedAccounts:    tp.RatedAccounts,
			RatedSubjects:    tp.RatedSubjects,
			CostInterval:     tp.CostInterval,
			ActionTriggers:   tp.ActionTriggers}
		if existing, exists := result[css.ID]; !exists {
			css.CdrStats = []*utils.TPCdrStat{cs}
			result[css.ID] = css
		} else {
			existing.CdrStats = append(existing.CdrStats, cs)
		}

	}
	return result, nil
}

func (tps TpCdrStats) AsTPCdrStats() (result []*utils.TPCdrStats, err error) {
	if utps, err := tps.GetCdrStatsA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func UpdateCdrStats(cs *CdrStats, triggers ActionTriggers, tpCs *utils.TPCdrStat, timezone string) {
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
	if tpCs.SaveInterval != "" {
		if si, err := time.ParseDuration(tpCs.SaveInterval); err == nil {
			cs.SaveInterval = si
		} else {
			log.Printf("Error parsing SaveInterval %v for cdr stats %v", tpCs.SaveInterval, cs.Id)
		}
	}
	if tpCs.Metrics != "" {
		cs.Metrics = append(cs.Metrics, tpCs.Metrics)
	}
	if tpCs.SetupInterval != "" {
		times := strings.Split(tpCs.SetupInterval, utils.INFIELD_SEP)
		if len(times) > 0 {
			if sTime, err := utils.ParseTimeDetectLayout(times[0], timezone); err == nil {
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
			if eTime, err := utils.ParseTimeDetectLayout(times[1], timezone); err == nil {
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
	if tpCs.DestinationIds != "" {
		cs.DestinationIds = append(cs.DestinationIds, tpCs.DestinationIds)
	}
	if tpCs.PddInterval != "" {
		pdds := strings.Split(tpCs.PddInterval, utils.INFIELD_SEP)
		if len(pdds) > 0 {
			if sPdd, err := time.ParseDuration(pdds[0]); err == nil {
				if len(cs.PddInterval) < 1 {
					cs.PddInterval = append(cs.PddInterval, sPdd)
				} else {
					cs.PddInterval[0] = sPdd
				}
			} else {
				log.Printf("Error parsing PddInterval %v for cdrs stats %v", tpCs.PddInterval, cs.Id)
			}
		}
		if len(pdds) > 1 {
			if ePdd, err := time.ParseDuration(pdds[1]); err == nil {
				if len(cs.PddInterval) < 2 {
					cs.PddInterval = append(cs.PddInterval, ePdd)
				} else {
					cs.PddInterval[1] = ePdd
				}
			} else {
				log.Printf("Error parsing UsageInterval %v for cdrs stats %v", tpCs.PddInterval, cs.Id)
			}
		}
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

type TpUsers []TpUser

func (tps TpUsers) GetUsers() (map[string]*utils.TPUsers, error) {
	users := make(map[string]*utils.TPUsers)
	for _, tp := range tps {
		var user *utils.TPUsers
		var found bool
		if user, found = users[tp.GetId()]; !found {
			user = &utils.TPUsers{
				Tenant:   tp.Tenant,
				UserName: tp.UserName,
				Weight:   tp.Weight,
			}
			users[tp.GetId()] = user
		}
		if tp.Masked == true {
			user.Masked = true
		}
		user.Profile = append(user.Profile,
			&utils.TPUserProfile{
				AttrName:  tp.AttributeName,
				AttrValue: tp.AttributeValue,
			})
	}
	return users, nil
}

func (tps TpUsers) GetUsersA() (map[string]*utils.TPUsers, error) {
	users := make(map[string]*utils.TPUsers)
	for _, tp := range tps {
		var user *utils.TPUsers
		var found bool
		if user, found = users[tp.GetId()]; !found {
			user = &utils.TPUsers{
				TPid:     tp.Tpid,
				Tenant:   tp.Tenant,
				UserName: tp.UserName,
				Masked:   tp.Masked,
				Weight:   tp.Weight,
			}
			users[tp.GetId()] = user
		}
		user.Profile = append(user.Profile,
			&utils.TPUserProfile{
				AttrName:  tp.AttributeName,
				AttrValue: tp.AttributeValue,
			})
	}
	return users, nil
}

func (tps TpUsers) AsTPUsers() (result []*utils.TPUsers, err error) {
	if utps, err := tps.GetUsersA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpAliases []TpAlias

func (tps TpAliases) GetAliases() (map[string]*utils.TPAliases, error) {
	als := make(map[string]*utils.TPAliases)
	for _, tp := range tps {
		var al *utils.TPAliases
		var found bool
		if al, found = als[tp.GetId()]; !found {
			al = &utils.TPAliases{
				Direction: tp.Direction,
				Tenant:    tp.Tenant,
				Category:  tp.Category,
				Account:   tp.Account,
				Subject:   tp.Subject,
				Context:   tp.Context,
			}
			als[tp.GetId()] = al
		}
		al.Values = append(al.Values, &utils.TPAliasValue{
			DestinationId: tp.DestinationId,
			Target:        tp.Target,
			Original:      tp.Original,
			Alias:         tp.Alias,
			Weight:        tp.Weight,
		})
	}
	return als, nil
}

func (tps TpAliases) AsTPAliases() (result []*utils.TPAliases, err error) {
	if utps, err := tps.GetAliases(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpLcrRules []TpLcrRule

func (tps TpLcrRules) GetLcrRules() (map[string]*utils.TPLcrRules, error) {
	lcrs := make(map[string]*utils.TPLcrRules)
	for _, tp := range tps {
		var lcr *utils.TPLcrRules
		var found bool
		if lcr, found = lcrs[tp.GetLcrRuleId()]; !found {
			lcr = &utils.TPLcrRules{
				Direction: tp.Direction,
				Tenant:    tp.Tenant,
				Category:  tp.Category,
				Account:   tp.Account,
				Subject:   tp.Subject,
			}
			lcrs[tp.GetLcrRuleId()] = lcr
		}
		lcr.Rules = append(lcr.Rules, &utils.TPLcrRule{
			DestinationId:  tp.DestinationTag,
			RpCategory:     tp.RpCategory,
			Strategy:       tp.Strategy,
			StrategyParams: tp.StrategyParams,
			ActivationTime: tp.ActivationTime,
			Weight:         tp.Weight,
		})
	}
	return lcrs, nil
}

func (tps TpLcrRules) GetLcrRulesA() (map[string]*utils.TPLcrRules, error) {
	lcrs := make(map[string]*utils.TPLcrRules)
	for _, tp := range tps {
		var lcr *utils.TPLcrRules
		var found bool
		if lcr, found = lcrs[tp.GetLcrRuleId()]; !found {
			lcr = &utils.TPLcrRules{
				TPid:      tp.Tpid,
				Direction: tp.Direction,
				Tenant:    tp.Tenant,
				Category:  tp.Category,
				Account:   tp.Account,
				Subject:   tp.Subject,
			}
			lcrs[tp.GetLcrRuleId()] = lcr
		}
		lcr.Rules = append(lcr.Rules, &utils.TPLcrRule{
			DestinationId:  tp.DestinationTag,
			RpCategory:     tp.RpCategory,
			Strategy:       tp.Strategy,
			StrategyParams: tp.StrategyParams,
			ActivationTime: tp.ActivationTime,
			Weight:         tp.Weight,
		})
	}
	return lcrs, nil
}

func (tps TpLcrRules) AsTPLcrRules() (result []*utils.TPLcrRules, err error) {
	if utps, err := tps.GetLcrRulesA(); err != nil {
		return nil, err
	} else {
		for _, tp := range utps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpResourceLimits []*TpResourceLimit

// AsTPResourceLimit converts TpResourceLimits into *utils.TPResourceLimit
func (trls TpResourceLimits) AsTPResourceLimits() (tpRLs []*utils.TPResourceLimit) {
	tpRLsMap := make(map[string]*utils.TPResourceLimit)
	for _, tpRL := range trls {
		resLimit, found := tpRLsMap[tpRL.Tag]
		if !found {
			resLimit = &utils.TPResourceLimit{
				TPid:           tpRL.Tpid,
				ID:             tpRL.Tag,
				ActivationTime: tpRL.ActivationTime,
				Weight:         tpRL.Weight,
				Limit:          tpRL.Limit,
			}
		}
		if tpRL.ActionTriggerIds != "" {
			resLimit.ActionTriggerIDs = append(resLimit.ActionTriggerIDs, strings.Split(tpRL.ActionTriggerIds, utils.INFIELD_SEP)...)
		}
		if tpRL.FilterType != "" {
			resLimit.Filters = append(resLimit.Filters, &utils.TPRequestFilter{
				Type:      tpRL.FilterType,
				FieldName: tpRL.FilterFieldName,
				Values:    strings.Split(tpRL.FilterFieldValues, utils.INFIELD_SEP)})
		}
		tpRLsMap[tpRL.Tag] = resLimit
	}

	tpRLs = make([]*utils.TPResourceLimit, len(tpRLsMap))
	i := 0
	for _, tpRL := range tpRLsMap {
		tpRLs[i] = tpRL
		i++
	}
	return
}
