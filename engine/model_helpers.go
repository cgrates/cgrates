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
	result := make(map[string]*Destination)
	for _, tp := range tps {
		var d *Destination
		var found bool
		if d, found = result[tp.Tag]; !found {
			d = &Destination{Id: tp.Tag}
			result[tp.Tag] = d
		}
		d.AddPrefix(tp.Prefix)
	}
	return result, nil
}

// AsTPDestination converts TpDestinations into *utils.TPDestination
func (tps TpDestinations) AsTPDestinations() (result []*utils.TPDestination) {
	md := make(map[string]*utils.TPDestination) // Should save us some CPU if we index here for big number of destinations to search
	for _, tp := range tps {
		if d, hasIt := md[tp.Tag]; !hasIt {
			md[tp.Tag] = &utils.TPDestination{TPid: tp.Tpid, ID: tp.Tag, Prefixes: []string{tp.Prefix}}
		} else {
			d.Prefixes = append(d.Prefixes, tp.Prefix)
		}
	}
	result = make([]*utils.TPDestination, len(md))
	i := 0
	for _, d := range md {
		result[i] = d
		i++
	}
	return
}

type TpTimings []TpTiming

func (tps TpTimings) GetTimings() (map[string]*utils.TPTiming, error) {
	result := make(map[string]*utils.TPTiming)
	for _, tp := range tps {
		t := &utils.TPTiming{}
		t.ID = tp.Tag
		t.Years.Parse(tp.Years, utils.INFIELD_SEP)
		t.Months.Parse(tp.Months, utils.INFIELD_SEP)
		t.MonthDays.Parse(tp.MonthDays, utils.INFIELD_SEP)
		t.WeekDays.Parse(tp.WeekDays, utils.INFIELD_SEP)
		times := strings.Split(tp.Time, utils.INFIELD_SEP)
		t.StartTime = times[0]
		if len(times) > 1 {
			t.EndTime = times[1]
		}
		if _, found := result[tp.Tag]; found {
			return nil, fmt.Errorf("duplicate timing tag: %s", tp.Tag)
		}
		result[tp.Tag] = t
	}
	return result, nil
}

func (tps TpTimings) GetApierTimings() (map[string]*utils.ApierTPTiming, error) {
	result := make(map[string]*utils.ApierTPTiming)
	for _, tp := range tps {
		t := &utils.ApierTPTiming{
			TPid:      tp.Tpid,
			ID:        tp.Tag,
			Years:     tp.Years,
			Months:    tp.Months,
			MonthDays: tp.MonthDays,
			WeekDays:  tp.WeekDays,
			Time:      tp.Time,
		}
		result[tp.Tag] = t
	}
	return result, nil
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
	result := make(map[string]*utils.TPRate)
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
		_, exists := result[tp.Tag]
		if exists {
			result[tp.Tag].RateSlots = append(result[tp.Tag].RateSlots, r.RateSlots[0])
		} else {
			result[tp.Tag] = r
		}
	}
	return result, nil
}

func (tps TpRates) GetRatesA() (map[string]*utils.TPRate, error) {
	result := make(map[string]*utils.TPRate)
	for _, tp := range tps {
		r := &utils.TPRate{
			TPid: tp.Tpid,
			ID:   tp.Tag,
		}
		rs := &utils.RateSlot{
			ConnectFee:         tp.ConnectFee,
			Rate:               tp.Rate,
			RateUnit:           tp.RateUnit,
			RateIncrement:      tp.RateIncrement,
			GroupIntervalStart: tp.GroupIntervalStart,
		}
		if existing, exists := result[r.ID]; !exists {
			r.RateSlots = []*utils.RateSlot{rs}
			result[r.ID] = r
		} else {
			existing.RateSlots = append(existing.RateSlots, rs)
		}

	}
	return result, nil
}

func (tps TpRates) AsTPRates() (result []*utils.TPRate, err error) {
	if atps, err := tps.GetRatesA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpDestinationRates []TpDestinationRate

func (tps TpDestinationRates) GetDestinationRates() (map[string]*utils.TPDestinationRate, error) {
	result := make(map[string]*utils.TPDestinationRate)
	for _, tp := range tps {
		dr := &utils.TPDestinationRate{
			TPid: tp.Tpid,
			ID:   tp.Tag,
			DestinationRates: []*utils.DestinationRate{
				&utils.DestinationRate{
					DestinationId:    tp.DestinationsTag,
					RateId:           tp.RatesTag,
					RoundingMethod:   tp.RoundingMethod,
					RoundingDecimals: tp.RoundingDecimals,
					MaxCost:          tp.MaxCost,
					MaxCostStrategy:  tp.MaxCostStrategy,
				},
			},
		}
		existing, exists := result[tp.Tag]
		if exists {
			existing.DestinationRates = append(existing.DestinationRates, dr.DestinationRates[0])
		} else {
			existing = dr
		}
		result[tp.Tag] = existing
	}
	return result, nil
}

func (tps TpDestinationRates) AsTPDestinationRates() (result []*utils.TPDestinationRate, err error) {
	if atps, err := tps.GetDestinationRates(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpRatingPlans []TpRatingPlan

func (tps TpRatingPlans) GetRatingPlanBindings() (map[string][]*utils.TPRatingPlanBinding, error) {
	result := make(map[string][]*utils.TPRatingPlanBinding)
	for _, tp := range tps {
		rpb := &utils.TPRatingPlanBinding{
			DestinationRatesId: tp.DestratesTag,
			TimingId:           tp.TimingTag,
			Weight:             tp.Weight,
		}
		if _, exists := result[tp.Tag]; exists {
			result[tp.Tag] = append(result[tp.Tag], rpb)
		} else {
			result[tp.Tag] = []*utils.TPRatingPlanBinding{rpb}
		}
	}
	return result, nil
}

func (tps TpRatingPlans) GetRatingPlans() (map[string]*utils.TPRatingPlan, error) {
	result := make(map[string]*utils.TPRatingPlan)
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
		if existing, exists := result[rp.ID]; !exists {
			rp.RatingPlanBindings = []*utils.TPRatingPlanBinding{rpb}
			result[rp.ID] = rp
		} else {
			existing.RatingPlanBindings = append(existing.RatingPlanBindings, rpb)
		}
	}
	return result, nil
}

func (tps TpRatingPlans) AsTPRatingPlans() (result []*utils.TPRatingPlan, err error) {
	if atps, err := tps.GetRatingPlans(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
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
	result := make(map[string]*utils.TPRatingProfile)
	for _, tp := range tps {
		rp := &utils.TPRatingProfile{
			TPid:      tp.Tpid,
			LoadId:    tp.Loadid,
			Direction: tp.Direction,
			Tenant:    tp.Tenant,
			Category:  tp.Category,
			Subject:   tp.Subject,
		}
		ra := &utils.TPRatingActivation{
			ActivationTime:   tp.ActivationTime,
			RatingPlanId:     tp.RatingPlanTag,
			FallbackSubjects: tp.FallbackSubjects,
			CdrStatQueueIds:  tp.CdrStatQueueIds,
		}
		if existing, exists := result[rp.KeyIdA()]; !exists {
			rp.RatingPlanActivations = []*utils.TPRatingActivation{ra}
			result[rp.KeyIdA()] = rp
		} else {
			existing.RatingPlanActivations = append(existing.RatingPlanActivations, ra)
		}
	}
	return result, nil
}

func (tps TpRatingProfiles) AsTPRatingProfiles() (result []*utils.TPRatingProfile, err error) {
	if atps, err := tps.GetRatingProfiles(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpSharedGroups []TpSharedGroup

func (tps TpSharedGroups) GetSharedGroups() (map[string][]*utils.TPSharedGroup, error) {
	result := make(map[string][]*utils.TPSharedGroup)
	for _, tp := range tps {
		result[tp.Tag] = append(result[tp.Tag], &utils.TPSharedGroup{
			Account:       tp.Account,
			Strategy:      tp.Strategy,
			RatingSubject: tp.RatingSubject,
		})
	}
	return result, nil
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
	if atps, err := tps.GetSharedGroupsA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpActions []TpAction

func (tps TpActions) GetActions() (map[string][]*utils.TPAction, error) {
	result := make(map[string][]*utils.TPAction)
	for _, tp := range tps {
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
		result[tp.Tag] = append(result[tp.Tag], a)
	}
	return result, nil
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
	if atps, err := tps.GetActionsA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpActionPlans []TpActionPlan

func (tps TpActionPlans) GetActionPlans() (map[string][]*utils.TPActionTiming, error) {
	result := make(map[string][]*utils.TPActionTiming)
	for _, tp := range tps {
		result[tp.Tag] = append(result[tp.Tag], &utils.TPActionTiming{ActionsId: tp.ActionsTag, TimingId: tp.TimingTag, Weight: tp.Weight})
	}
	return result, nil
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
	if atps, err := tps.GetActionPlansA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpActionTriggers []TpActionTrigger

func (tps TpActionTriggers) GetActionTriggers() (map[string][]*utils.TPActionTrigger, error) {
	result := make(map[string][]*utils.TPActionTrigger)
	for _, tp := range tps {
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
		result[tp.Tag] = append(result[tp.Tag], at)
	}
	return result, nil
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
	if atps, err := tps.GetActionTriggersA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpAccountActions []TpAccountAction

func (tps TpAccountActions) GetAccountActions() (map[string]*utils.TPAccountActions, error) {
	result := make(map[string]*utils.TPAccountActions)
	for _, tp := range tps {
		aas := &utils.TPAccountActions{
			TPid:             tp.Tpid,
			LoadId:           tp.Loadid,
			Tenant:           tp.Tenant,
			Account:          tp.Account,
			ActionPlanId:     tp.ActionPlanTag,
			ActionTriggersId: tp.ActionTriggersTag,
			AllowNegative:    tp.AllowNegative,
			Disabled:         tp.Disabled,
		}
		result[aas.KeyId()] = aas
	}
	return result, nil
}

func (tps TpAccountActions) AsTPAccountActions() (result []*utils.TPAccountActions, err error) {
	if atps, err := tps.GetAccountActions(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpDerivedChargers []TpDerivedCharger

func (tps TpDerivedChargers) GetDerivedChargers() (map[string]*utils.TPDerivedChargers, error) {
	result := make(map[string]*utils.TPDerivedChargers)
	for _, tp := range tps {
		dcs := &utils.TPDerivedChargers{
			TPid:           tp.Tpid,
			LoadId:         tp.Loadid,
			Direction:      tp.Direction,
			Tenant:         tp.Tenant,
			Category:       tp.Category,
			Account:        tp.Account,
			Subject:        tp.Subject,
			DestinationIds: tp.DestinationIds,
		}
		tag := dcs.GetDerivedChargesId()
		if _, hasIt := result[tag]; !hasIt {
			result[tag] = dcs
		}
		dc := &utils.TPDerivedCharger{
			RunId:                ValueOrDefault(tp.Runid, utils.META_DEFAULT),
			RunFilters:           tp.RunFilters,
			ReqTypeField:         ValueOrDefault(tp.ReqTypeField, utils.META_DEFAULT),
			DirectionField:       ValueOrDefault(tp.DirectionField, utils.META_DEFAULT),
			TenantField:          ValueOrDefault(tp.TenantField, utils.META_DEFAULT),
			CategoryField:        ValueOrDefault(tp.CategoryField, utils.META_DEFAULT),
			AccountField:         ValueOrDefault(tp.AccountField, utils.META_DEFAULT),
			SubjectField:         ValueOrDefault(tp.SubjectField, utils.META_DEFAULT),
			DestinationField:     ValueOrDefault(tp.DestinationField, utils.META_DEFAULT),
			SetupTimeField:       ValueOrDefault(tp.SetupTimeField, utils.META_DEFAULT),
			PddField:             ValueOrDefault(tp.PddField, utils.META_DEFAULT),
			AnswerTimeField:      ValueOrDefault(tp.AnswerTimeField, utils.META_DEFAULT),
			UsageField:           ValueOrDefault(tp.UsageField, utils.META_DEFAULT),
			SupplierField:        ValueOrDefault(tp.SupplierField, utils.META_DEFAULT),
			DisconnectCauseField: ValueOrDefault(tp.DisconnectCauseField, utils.META_DEFAULT),
			CostField:            ValueOrDefault(tp.CostField, utils.META_DEFAULT),
			RatedField:           ValueOrDefault(tp.RatedField, utils.META_DEFAULT),
		}
		result[tag].DerivedChargers = append(result[tag].DerivedChargers, dc)
	}
	return result, nil
}

func (tps TpDerivedChargers) AsTPDerivedChargers() (result []*utils.TPDerivedChargers, err error) {
	if atps, err := tps.GetDerivedChargers(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpCdrStats []TpCdrstat

func (tps TpCdrStats) GetCdrStats() (map[string][]*utils.TPCdrStat, error) {
	result := make(map[string][]*utils.TPCdrStat)
	for _, tp := range tps {
		result[tp.Tag] = append(result[tp.Tag], &utils.TPCdrStat{
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
			ActionTriggers:   tp.ActionTriggers,
		})
	}
	return result, nil
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
	if atps, err := tps.GetCdrStatsA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
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
	result := make(map[string]*utils.TPUsers)
	for _, tp := range tps {
		var u *utils.TPUsers
		var found bool
		if u, found = result[tp.GetId()]; !found {
			u = &utils.TPUsers{
				Tenant:   tp.Tenant,
				UserName: tp.UserName,
				Weight:   tp.Weight,
			}
			result[tp.GetId()] = u
		}
		if tp.Masked == true {
			u.Masked = true
		}
		u.Profile = append(u.Profile,
			&utils.TPUserProfile{
				AttrName:  tp.AttributeName,
				AttrValue: tp.AttributeValue,
			})
	}
	return result, nil
}

func (tps TpUsers) GetUsersA() (map[string]*utils.TPUsers, error) {
	result := make(map[string]*utils.TPUsers)
	for _, tp := range tps {
		var u *utils.TPUsers
		var found bool
		if u, found = result[tp.GetId()]; !found {
			u = &utils.TPUsers{
				TPid:     tp.Tpid,
				Tenant:   tp.Tenant,
				UserName: tp.UserName,
				Masked:   tp.Masked,
				Weight:   tp.Weight,
			}
			result[tp.GetId()] = u
		}
		u.Profile = append(u.Profile,
			&utils.TPUserProfile{
				AttrName:  tp.AttributeName,
				AttrValue: tp.AttributeValue,
			})
	}
	return result, nil
}

func (tps TpUsers) AsTPUsers() (result []*utils.TPUsers, err error) {
	if atps, err := tps.GetUsersA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpAliases []TpAlias

func (tps TpAliases) GetAliases() (map[string]*utils.TPAliases, error) {
	result := make(map[string]*utils.TPAliases)
	for _, tp := range tps {
		var as *utils.TPAliases
		var found bool
		if as, found = result[tp.GetId()]; !found {
			as = &utils.TPAliases{
				Direction: tp.Direction,
				Tenant:    tp.Tenant,
				Category:  tp.Category,
				Account:   tp.Account,
				Subject:   tp.Subject,
				Context:   tp.Context,
			}
			result[tp.GetId()] = as
		}
		as.Values = append(as.Values, &utils.TPAliasValue{
			DestinationId: tp.DestinationId,
			Target:        tp.Target,
			Original:      tp.Original,
			Alias:         tp.Alias,
			Weight:        tp.Weight,
		})
	}
	return result, nil
}

func (tps TpAliases) AsTPAliases() (result []*utils.TPAliases, err error) {
	if atps, err := tps.GetAliases(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpLcrRules []TpLcrRule

func (tps TpLcrRules) GetLcrRules() (map[string]*utils.TPLcrRules, error) {
	result := make(map[string]*utils.TPLcrRules)
	for _, tp := range tps {
		var lrs *utils.TPLcrRules
		var found bool
		if lrs, found = result[tp.GetLcrRuleId()]; !found {
			lrs = &utils.TPLcrRules{
				Direction: tp.Direction,
				Tenant:    tp.Tenant,
				Category:  tp.Category,
				Account:   tp.Account,
				Subject:   tp.Subject,
			}
			result[tp.GetLcrRuleId()] = lrs
		}
		lrs.Rules = append(lrs.Rules, &utils.TPLcrRule{
			DestinationId:  tp.DestinationTag,
			RpCategory:     tp.RpCategory,
			Strategy:       tp.Strategy,
			StrategyParams: tp.StrategyParams,
			ActivationTime: tp.ActivationTime,
			Weight:         tp.Weight,
		})
	}
	return result, nil
}

func (tps TpLcrRules) GetLcrRulesA() (map[string]*utils.TPLcrRules, error) {
	result := make(map[string]*utils.TPLcrRules)
	for _, tp := range tps {
		var lrs *utils.TPLcrRules
		var found bool
		if lrs, found = result[tp.GetLcrRuleId()]; !found {
			lrs = &utils.TPLcrRules{
				TPid:      tp.Tpid,
				Direction: tp.Direction,
				Tenant:    tp.Tenant,
				Category:  tp.Category,
				Account:   tp.Account,
				Subject:   tp.Subject,
			}
			result[tp.GetLcrRuleId()] = lrs
		}
		lrs.Rules = append(lrs.Rules, &utils.TPLcrRule{
			DestinationId:  tp.DestinationTag,
			RpCategory:     tp.RpCategory,
			Strategy:       tp.Strategy,
			StrategyParams: tp.StrategyParams,
			ActivationTime: tp.ActivationTime,
			Weight:         tp.Weight,
		})
	}
	return result, nil
}

func (tps TpLcrRules) AsTPLcrRules() (result []*utils.TPLcrRules, err error) {
	if atps, err := tps.GetLcrRulesA(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

type TpResourceLimits []*TpResourceLimit

func (tps TpResourceLimits) AsTPResourceLimits() (result []*utils.TPResourceLimit) {
	mrl := make(map[string]*utils.TPResourceLimit)
	for _, tp := range tps {
		rl, found := mrl[tp.Tag]
		if !found {
			rl = &utils.TPResourceLimit{
				TPid:           tp.Tpid,
				ID:             tp.Tag,
				ActivationTime: tp.ActivationTime,
				Weight:         tp.Weight,
				Limit:          tp.Limit,
			}
		}
		if tp.ActionTriggerIds != "" {
			rl.ActionTriggerIDs = append(rl.ActionTriggerIDs, strings.Split(tp.ActionTriggerIds, utils.INFIELD_SEP)...)
		}
		if tp.FilterType != "" {
			rl.Filters = append(rl.Filters, &utils.TPRequestFilter{
				Type:      tp.FilterType,
				FieldName: tp.FilterFieldName,
				Values:    strings.Split(tp.FilterFieldValues, utils.INFIELD_SEP)})
		}
		mrl[tp.Tag] = rl
	}
	result = make([]*utils.TPResourceLimit, len(mrl))
	i := 0
	for _, rl := range mrl {
		result[i] = rl
		i++
	}
	return
}
