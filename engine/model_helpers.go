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
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
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
	st := reflect.ValueOf(s)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		s = st.Interface()
	}
	numFields := st.NumField()
	stcopy := reflect.TypeOf(s)
	for i := 0; i < numFields; i++ {
		field := stcopy.Field(i)
		index := field.Tag.Get("index")
		if index != "" {
			if idx, err := strconv.Atoi(index); err != nil {
				return nil, fmt.Errorf("invalid %v.%v index %v", stcopy.Name(), field.Name, index)
			} else {
				fieldIndexMap[field.Name] = idx
			}
		}
	}
	result := make([]string, len(fieldIndexMap))
	for fieldName, fieldIndex := range fieldIndexMap {
		field := st.FieldByName(fieldName)
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

func (tps TpDestinations) AsMapDestinations() (map[string]*Destination, error) {
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

type TpTimings []TpTiming

func (tps TpTimings) AsMapTPTimings() (map[string]*utils.ApierTPTiming, error) {
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

func MapTPTimings(tps []*utils.ApierTPTiming) (map[string]*utils.TPTiming, error) {
	result := make(map[string]*utils.TPTiming)
	for _, tp := range tps {
		t := &utils.TPTiming{}
		t.ID = tp.ID
		t.Years.Parse(tp.Years, utils.INFIELD_SEP)
		t.Months.Parse(tp.Months, utils.INFIELD_SEP)
		t.MonthDays.Parse(tp.MonthDays, utils.INFIELD_SEP)
		t.WeekDays.Parse(tp.WeekDays, utils.INFIELD_SEP)
		times := strings.Split(tp.Time, utils.INFIELD_SEP)
		t.StartTime = times[0]
		if len(times) > 1 {
			t.EndTime = times[1]
		}
		if _, found := result[tp.ID]; found {
			return nil, fmt.Errorf("duplicate timing tag: %s", tp.ID)
		}
		result[tp.ID] = t
	}
	return result, nil
}

func (tps TpTimings) AsTPTimings() (result []*utils.ApierTPTiming) {
	ats, _ := tps.AsMapTPTimings()
	for _, tp := range ats {
		result = append(result, tp)
	}
	return result
}

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

type TpRates []TpRate

func (tps TpRates) AsMapRates() (map[string]*utils.TPRate, error) {
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
		if err := rs.SetDurations(); err != nil {
			return nil, err
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
	if atps, err := tps.AsMapRates(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPRates(s []*utils.TPRate) (map[string]*utils.TPRate, error) {
	result := make(map[string]*utils.TPRate)
	for _, e := range s {
		if _, found := result[e.ID]; !found {
			result[e.ID] = e
		} else {
			return nil, fmt.Errorf("Non unique ID %+v", e.ID)
		}
	}
	return result, nil
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

type TpDestinationRates []TpDestinationRate

func (tps TpDestinationRates) AsMapDestinationRates() (map[string]*utils.TPDestinationRate, error) {
	result := make(map[string]*utils.TPDestinationRate)
	for _, tp := range tps {
		dr := &utils.TPDestinationRate{
			TPid: tp.Tpid,
			ID:   tp.Tag,
			DestinationRates: []*utils.DestinationRate{
				{
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
	if atps, err := tps.AsMapDestinationRates(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPDestinationRates(s []*utils.TPDestinationRate) (map[string]*utils.TPDestinationRate, error) {
	result := make(map[string]*utils.TPDestinationRate)
	for _, e := range s {
		if _, found := result[e.ID]; !found {
			result[e.ID] = e
		} else {
			return nil, fmt.Errorf("Non unique ID %+v", e.ID)
		}
	}
	return result, nil
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

type TpRatingPlans []TpRatingPlan

func (tps TpRatingPlans) AsMapTPRatingPlans() (map[string]*utils.TPRatingPlan, error) {
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
	if atps, err := tps.AsMapTPRatingPlans(); err != nil {
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

func MapTPRatingPlanBindings(s []*utils.TPRatingPlan) map[string][]*utils.TPRatingPlanBinding {
	result := make(map[string][]*utils.TPRatingPlanBinding)
	for _, e := range s {
		for _, rpb := range e.RatingPlanBindings {
			if _, found := result[e.ID]; !found {
				result[e.ID] = []*utils.TPRatingPlanBinding{rpb}
			} else {
				result[e.ID] = append(result[e.ID], rpb)
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

type TpRatingProfiles []TpRatingProfile

func (tps TpRatingProfiles) AsMapTPRatingProfiles() (map[string]*utils.TPRatingProfile, error) {
	result := make(map[string]*utils.TPRatingProfile)
	for _, tp := range tps {
		rp := &utils.TPRatingProfile{
			TPid:     tp.Tpid,
			LoadId:   tp.Loadid,
			Tenant:   tp.Tenant,
			Category: tp.Category,
			Subject:  tp.Subject,
		}
		ra := &utils.TPRatingActivation{
			ActivationTime:   tp.ActivationTime,
			RatingPlanId:     tp.RatingPlanTag,
			FallbackSubjects: tp.FallbackSubjects,
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
	if atps, err := tps.AsMapTPRatingProfiles(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPRatingProfiles(s []*utils.TPRatingProfile) (map[string]*utils.TPRatingProfile, error) {
	result := make(map[string]*utils.TPRatingProfile)
	for _, e := range s {
		if _, found := result[e.KeyIdA()]; !found {
			result[e.KeyIdA()] = e
		} else {
			return nil, fmt.Errorf("Non unique id %+v", e.KeyIdA())
		}
	}
	return result, nil
}

func APItoModelRatingProfile(rp *utils.TPRatingProfile) (result TpRatingProfiles) {
	if rp != nil {
		for _, rpa := range rp.RatingPlanActivations {
			result = append(result, TpRatingProfile{
				Tpid:             rp.TPid,
				Loadid:           rp.LoadId,
				Tenant:           rp.Tenant,
				Category:         rp.Category,
				Subject:          rp.Subject,
				ActivationTime:   rpa.ActivationTime,
				RatingPlanTag:    rpa.RatingPlanId,
				FallbackSubjects: rpa.FallbackSubjects,
			})
		}
		if len(rp.RatingPlanActivations) == 0 {
			result = append(result, TpRatingProfile{
				Tpid:     rp.TPid,
				Loadid:   rp.LoadId,
				Tenant:   rp.Tenant,
				Category: rp.Category,
				Subject:  rp.Subject,
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

type TpSharedGroups []TpSharedGroup

func (tps TpSharedGroups) AsMapTPSharedGroups() (map[string]*utils.TPSharedGroups, error) {
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
	if atps, err := tps.AsMapTPSharedGroups(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPSharedGroup(s []*utils.TPSharedGroups) map[string][]*utils.TPSharedGroup {
	result := make(map[string][]*utils.TPSharedGroup)
	for _, e := range s {
		for _, sg := range e.SharedGroups {
			if _, found := result[e.ID]; !found {
				result[e.ID] = []*utils.TPSharedGroup{sg}
			} else {
				result[e.ID] = append(result[e.ID], sg)
			}
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

type TpActions []TpAction

func (tps TpActions) AsMapTPActions() (map[string]*utils.TPActions, error) {
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
	if atps, err := tps.AsMapTPActions(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPActions(s []*utils.TPActions) map[string][]*utils.TPAction {
	result := make(map[string][]*utils.TPAction)
	for _, e := range s {
		for _, a := range e.Actions {
			if _, found := result[e.ID]; !found {
				result[e.ID] = []*utils.TPAction{a}
			} else {
				result[e.ID] = append(result[e.ID], a)
			}
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

type TpActionPlans []TpActionPlan

func (tps TpActionPlans) AsMapTPActionPlans() (map[string]*utils.TPActionPlan, error) {
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
	if atps, err := tps.AsMapTPActionPlans(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPActionTimings(s []*utils.TPActionPlan) map[string][]*utils.TPActionTiming {
	result := make(map[string][]*utils.TPActionTiming)
	for _, e := range s {
		for _, at := range e.ActionPlan {
			if _, found := result[e.ID]; !found {
				result[e.ID] = []*utils.TPActionTiming{at}
			} else {
				result[e.ID] = append(result[e.ID], at)
			}
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

type TpActionTriggers []TpActionTrigger

func (tps TpActionTriggers) AsMapTPActionTriggers() (map[string]*utils.TPActionTriggers, error) {
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
	if atps, err := tps.AsMapTPActionTriggers(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPActionTriggers(s []*utils.TPActionTriggers) map[string][]*utils.TPActionTrigger {
	result := make(map[string][]*utils.TPActionTrigger)
	for _, e := range s {
		for _, at := range e.ActionTriggers {
			if _, found := result[e.ID]; !found {
				result[e.ID] = []*utils.TPActionTrigger{at}
			} else {
				result[e.ID] = append(result[e.ID], at)
			}
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
				BalanceDestinationTags: at.BalanceDestinationIds,
				BalanceWeight:          at.BalanceWeight,
				BalanceExpiryTime:      at.BalanceExpirationDate,
				BalanceTimingTags:      at.BalanceTimingTags,
				BalanceRatingSubject:   at.BalanceRatingSubject,
				BalanceCategories:      at.BalanceCategories,
				BalanceSharedGroups:    at.BalanceSharedGroups,
				BalanceBlocker:         at.BalanceBlocker,
				BalanceDisabled:        at.BalanceDisabled,
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

type TpAccountActions []TpAccountAction

func (tps TpAccountActions) AsMapTPAccountActions() (map[string]*utils.TPAccountActions, error) {
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
	if atps, err := tps.AsMapTPAccountActions(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPAccountActions(s []*utils.TPAccountActions) (map[string]*utils.TPAccountActions, error) {
	result := make(map[string]*utils.TPAccountActions)
	for _, e := range s {
		if _, found := result[e.KeyId()]; !found {
			result[e.KeyId()] = e
		} else {
			return nil, fmt.Errorf("Non unique ID %+v", e.KeyId())
		}
	}
	return result, nil
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

type TpResources []*TpResource

func (tps TpResources) AsTPResources() (result []*utils.TPResourceProfile) {
	mrl := make(map[string]*utils.TPResourceProfile)
	filterMap := make(map[string]utils.StringMap)
	thresholdMap := make(map[string]utils.StringMap)
	for _, tp := range tps {
		rl, found := mrl[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]
		if !found {
			rl = &utils.TPResourceProfile{
				TPid:    tp.Tpid,
				Tenant:  tp.Tenant,
				ID:      tp.ID,
				Blocker: tp.Blocker,
				Stored:  tp.Stored,
			}
		}
		if tp.UsageTTL != "" {
			rl.UsageTTL = tp.UsageTTL
		}
		if tp.Weight != 0 {
			rl.Weight = tp.Weight
		}
		if tp.Limit != "" {
			rl.Limit = tp.Limit
		}
		if tp.AllocationMessage != "" {
			rl.AllocationMessage = tp.AllocationMessage
		}
		rl.Blocker = tp.Blocker
		rl.Stored = tp.Stored
		if len(tp.ActivationInterval) != 0 {
			rl.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				rl.ActivationInterval.ActivationTime = aiSplt[0]
				rl.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				rl.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.ThresholdIDs != "" {
			if _, has := thresholdMap[(&utils.TenantID{Tenant: tp.Tenant,
				ID: tp.ID}).TenantID()]; !has {
				thresholdMap[(&utils.TenantID{Tenant: tp.Tenant,
					ID: tp.ID}).TenantID()] = make(utils.StringMap)
			}
			trshSplt := strings.Split(tp.ThresholdIDs, utils.INFIELD_SEP)
			for _, trsh := range trshSplt {
				thresholdMap[(&utils.TenantID{Tenant: tp.Tenant,
					ID: tp.ID}).TenantID()][trsh] = true
			}
		}
		if tp.FilterIDs != "" {
			if _, has := filterMap[(&utils.TenantID{Tenant: tp.Tenant,
				ID: tp.ID}).TenantID()]; !has {
				filterMap[(&utils.TenantID{Tenant: tp.Tenant,
					ID: tp.ID}).TenantID()] = make(utils.StringMap)
			}
			filterSplit := strings.Split(tp.FilterIDs, utils.INFIELD_SEP)
			for _, filter := range filterSplit {
				filterMap[(&utils.TenantID{Tenant: tp.Tenant,
					ID: tp.ID}).TenantID()][filter] = true
			}
		}
		mrl[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = rl
	}
	result = make([]*utils.TPResourceProfile, len(mrl))
	i := 0
	for tntID, rl := range mrl {
		result[i] = rl
		for filter := range filterMap[tntID] {
			result[i].FilterIDs = append(result[i].FilterIDs, filter)
		}
		for threshold := range thresholdMap[tntID] {
			result[i].ThresholdIDs = append(result[i].ThresholdIDs, threshold)
		}
		i++
	}
	return
}

func APItoModelResource(rl *utils.TPResourceProfile) (mdls TpResources) {
	if rl == nil {
		return
	}
	// In case that TPResourceProfile don't have filter
	if len(rl.FilterIDs) == 0 {
		mdl := &TpResource{
			Tpid:              rl.TPid,
			Tenant:            rl.Tenant,
			ID:                rl.ID,
			Blocker:           rl.Blocker,
			Stored:            rl.Stored,
			UsageTTL:          rl.UsageTTL,
			Weight:            rl.Weight,
			Limit:             rl.Limit,
			AllocationMessage: rl.AllocationMessage,
		}
		if rl.ActivationInterval != nil {
			if rl.ActivationInterval.ActivationTime != "" {
				mdl.ActivationInterval = rl.ActivationInterval.ActivationTime
			}
			if rl.ActivationInterval.ExpiryTime != "" {
				mdl.ActivationInterval += utils.INFIELD_SEP + rl.ActivationInterval.ExpiryTime
			}
		}
		for i, val := range rl.ThresholdIDs {
			if i != 0 {
				mdl.ThresholdIDs += utils.INFIELD_SEP
			}
			mdl.ThresholdIDs += val
		}
		mdls = append(mdls, mdl)
	}
	for i, fltr := range rl.FilterIDs {
		mdl := &TpResource{
			Tpid:    rl.TPid,
			Tenant:  rl.Tenant,
			ID:      rl.ID,
			Blocker: rl.Blocker,
			Stored:  rl.Stored,
		}
		if i == 0 {
			mdl.UsageTTL = rl.UsageTTL
			mdl.Weight = rl.Weight
			mdl.Limit = rl.Limit
			mdl.AllocationMessage = rl.AllocationMessage
			if rl.ActivationInterval != nil {
				if rl.ActivationInterval.ActivationTime != "" {
					mdl.ActivationInterval = rl.ActivationInterval.ActivationTime
				}
				if rl.ActivationInterval.ExpiryTime != "" {
					mdl.ActivationInterval += utils.INFIELD_SEP + rl.ActivationInterval.ExpiryTime
				}
			}
			for i, val := range rl.ThresholdIDs {
				if i != 0 {
					mdl.ThresholdIDs += utils.INFIELD_SEP
				}
				mdl.ThresholdIDs += val
			}
		}
		mdl.FilterIDs = fltr
		mdls = append(mdls, mdl)
	}
	return
}

func APItoResource(tpRL *utils.TPResourceProfile, timezone string) (rp *ResourceProfile, err error) {
	rp = &ResourceProfile{
		Tenant:            tpRL.Tenant,
		ID:                tpRL.ID,
		Weight:            tpRL.Weight,
		Blocker:           tpRL.Blocker,
		Stored:            tpRL.Stored,
		AllocationMessage: tpRL.AllocationMessage,
		ThresholdIDs:      make([]string, len(tpRL.ThresholdIDs)),
		FilterIDs:         make([]string, len(tpRL.FilterIDs)),
	}
	if tpRL.UsageTTL != "" {
		if rp.UsageTTL, err = utils.ParseDurationWithNanosecs(tpRL.UsageTTL); err != nil {
			return nil, err
		}
	}
	for i, fltr := range tpRL.FilterIDs {
		rp.FilterIDs[i] = fltr
	}
	for i, th := range tpRL.ThresholdIDs {
		rp.ThresholdIDs[i] = th
	}
	if tpRL.ActivationInterval != nil {
		if rp.ActivationInterval, err = tpRL.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	if tpRL.Limit != "" {
		if rp.Limit, err = strconv.ParseFloat(tpRL.Limit, 64); err != nil {
			return nil, err
		}
	}
	return rp, nil
}

type TpStats []*TpStat

func (models TpStats) AsTPStats() (result []*utils.TPStatProfile) {
	filterMap := make(map[string]utils.StringMap)
	thresholdMap := make(map[string]utils.StringMap)
	statMetricsMap := make(map[string]map[string]*utils.MetricWithFilters)
	mst := make(map[string]*utils.TPStatProfile)
	for _, model := range models {
		key := &utils.TenantID{Tenant: model.Tenant, ID: model.ID}
		st, found := mst[key.TenantID()]
		if !found {
			st = &utils.TPStatProfile{
				Tenant:      model.Tenant,
				TPid:        model.Tpid,
				ID:          model.ID,
				Blocker:     model.Blocker,
				Stored:      model.Stored,
				Weight:      model.Weight,
				MinItems:    model.MinItems,
				TTL:         model.TTL,
				QueueLength: model.QueueLength,
			}
		}
		if model.Blocker {
			st.Blocker = model.Blocker
		}
		if model.Stored {
			st.Stored = model.Stored
		}
		if model.Weight != 0 {
			st.Weight = model.Weight
		}
		if model.MinItems != 0 {
			st.MinItems = model.MinItems
		}
		if model.TTL != "" {
			st.TTL = model.TTL
		}
		if model.QueueLength != 0 {
			st.QueueLength = model.QueueLength
		}
		if model.ThresholdIDs != "" {
			if _, has := thresholdMap[key.TenantID()]; !has {
				thresholdMap[key.TenantID()] = make(utils.StringMap)
			}
			trshSplt := strings.Split(model.ThresholdIDs, utils.INFIELD_SEP)
			for _, trsh := range trshSplt {
				thresholdMap[key.TenantID()][trsh] = true
			}
		}
		if len(model.ActivationInterval) != 0 {
			st.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(model.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				st.ActivationInterval.ActivationTime = aiSplt[0]
				st.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				st.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if model.FilterIDs != "" {
			if _, has := filterMap[key.TenantID()]; !has {
				filterMap[key.TenantID()] = make(utils.StringMap)
			}
			filterSplit := strings.Split(model.FilterIDs, utils.INFIELD_SEP)
			for _, filter := range filterSplit {
				filterMap[key.TenantID()][filter] = true
			}
		}
		if model.MetricIDs != "" {
			if _, has := statMetricsMap[key.TenantID()]; !has {
				statMetricsMap[key.TenantID()] = make(map[string]*utils.MetricWithFilters)
			}
			metricIDsSplit := strings.Split(model.MetricIDs, utils.INFIELD_SEP)
			for _, metricID := range metricIDsSplit {
				stsMetric, found := statMetricsMap[key.TenantID()][metricID]
				if !found {
					stsMetric = &utils.MetricWithFilters{
						MetricID: metricID,
					}
				}
				if model.MetricFilterIDs != "" {
					filterIDs := strings.Split(model.MetricFilterIDs, utils.INFIELD_SEP)
					stsMetric.FilterIDs = append(stsMetric.FilterIDs, filterIDs...)
				}
				statMetricsMap[key.TenantID()][metricID] = stsMetric
			}
		}
		mst[key.TenantID()] = st
	}
	result = make([]*utils.TPStatProfile, len(mst))
	i := 0
	for tntID, st := range mst {
		result[i] = st
		for filter := range filterMap[tntID] {
			result[i].FilterIDs = append(result[i].FilterIDs, filter)
		}
		for threshold := range thresholdMap[tntID] {
			result[i].ThresholdIDs = append(result[i].ThresholdIDs, threshold)
		}
		for _, metric := range statMetricsMap[tntID] {
			result[i].Metrics = append(result[i].Metrics, metric)
		}
		i++
	}
	return
}

func APItoModelStats(st *utils.TPStatProfile) (mdls TpStats) {
	if st != nil && len(st.Metrics) != 0 {
		for i, metric := range st.Metrics {
			mdl := &TpStat{
				Tpid:   st.TPid,
				Tenant: st.Tenant,
				ID:     st.ID,
			}
			if i == 0 {
				for i, val := range st.FilterIDs {
					if i != 0 {
						mdl.FilterIDs += utils.INFIELD_SEP
					}
					mdl.FilterIDs += val
				}
				if st.ActivationInterval != nil {
					if st.ActivationInterval.ActivationTime != "" {
						mdl.ActivationInterval = st.ActivationInterval.ActivationTime
					}
					if st.ActivationInterval.ExpiryTime != "" {
						mdl.ActivationInterval += utils.INFIELD_SEP + st.ActivationInterval.ExpiryTime
					}
				}
				mdl.QueueLength = st.QueueLength
				mdl.TTL = st.TTL
				mdl.MinItems = st.MinItems
				mdl.Stored = st.Stored
				mdl.Blocker = st.Blocker
				mdl.Weight = st.Weight
				for i, val := range st.ThresholdIDs {
					if i != 0 {
						mdl.ThresholdIDs += utils.INFIELD_SEP
					}
					mdl.ThresholdIDs += val
				}
			}
			for i, val := range metric.FilterIDs {
				if i != 0 {
					mdl.MetricFilterIDs += utils.INFIELD_SEP
				}
				mdl.MetricFilterIDs += val
			}
			mdl.MetricIDs = metric.MetricID
			mdls = append(mdls, mdl)
		}
	}
	return
}

func APItoStats(tpST *utils.TPStatProfile, timezone string) (st *StatQueueProfile, err error) {
	st = &StatQueueProfile{
		Tenant:       tpST.Tenant,
		ID:           tpST.ID,
		FilterIDs:    make([]string, len(tpST.FilterIDs)),
		QueueLength:  tpST.QueueLength,
		MinItems:     tpST.MinItems,
		Metrics:      make([]*MetricWithFilters, len(tpST.Metrics)),
		Stored:       tpST.Stored,
		Blocker:      tpST.Blocker,
		Weight:       tpST.Weight,
		ThresholdIDs: make([]string, len(tpST.ThresholdIDs)),
	}
	if tpST.TTL != "" {
		if st.TTL, err = utils.ParseDurationWithNanosecs(tpST.TTL); err != nil {
			return nil, err
		}
	}
	for i, metric := range tpST.Metrics {
		st.Metrics[i] = &MetricWithFilters{
			MetricID:  metric.MetricID,
			FilterIDs: metric.FilterIDs,
		}
	}
	for i, trh := range tpST.ThresholdIDs {
		st.ThresholdIDs[i] = trh
	}
	for i, fltr := range tpST.FilterIDs {
		st.FilterIDs[i] = fltr
	}
	if tpST.ActivationInterval != nil {
		if st.ActivationInterval, err = tpST.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	return st, nil
}

type TpThresholds []*TpThreshold

func (tps TpThresholds) AsTPThreshold() (result []*utils.TPThresholdProfile) {
	mst := make(map[string]*utils.TPThresholdProfile)
	filterMap := make(map[string]utils.StringMap)
	actionMap := make(map[string]utils.StringMap)
	for _, tp := range tps {
		th, found := mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]
		if !found {
			th = &utils.TPThresholdProfile{
				TPid:     tp.Tpid,
				Tenant:   tp.Tenant,
				ID:       tp.ID,
				Blocker:  tp.Blocker,
				MaxHits:  tp.MaxHits,
				MinHits:  tp.MinHits,
				MinSleep: tp.MinSleep,
				Async:    tp.Async,
			}
		}
		if tp.ActionIDs != "" {
			if _, has := actionMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]; !has {
				actionMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = make(utils.StringMap)
			}
			actionSplit := strings.Split(tp.ActionIDs, utils.INFIELD_SEP)
			for _, action := range actionSplit {
				actionMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()][action] = true
			}
		}
		if tp.Weight != 0 {
			th.Weight = tp.Weight
		}
		if len(tp.ActivationInterval) != 0 {
			th.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != "" {
			if _, has := filterMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]; !has {
				filterMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = make(utils.StringMap)
			}
			filterSplit := strings.Split(tp.FilterIDs, utils.INFIELD_SEP)
			for _, filter := range filterSplit {
				filterMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()][filter] = true
			}
		}

		mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = th
	}
	result = make([]*utils.TPThresholdProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		for filter := range filterMap[tntID] {
			result[i].FilterIDs = append(result[i].FilterIDs, filter)
		}
		for action := range actionMap[tntID] {
			result[i].ActionIDs = append(result[i].ActionIDs, action)
		}
		i++
	}
	return
}

func APItoModelTPThreshold(th *utils.TPThresholdProfile) (mdls TpThresholds) {
	if th != nil {
		if len(th.ActionIDs) == 0 {
			return
		}
		min := len(th.FilterIDs)
		if min > len(th.ActionIDs) {
			min = len(th.ActionIDs)
		}
		for i := 0; i < min; i++ {
			mdl := &TpThreshold{
				Tpid:   th.TPid,
				Tenant: th.Tenant,
				ID:     th.ID,
			}
			if i == 0 {
				mdl.Blocker = th.Blocker
				mdl.Weight = th.Weight
				mdl.MaxHits = th.MaxHits
				mdl.MinHits = th.MinHits
				mdl.MinSleep = th.MinSleep
				mdl.Async = th.Async
				if th.ActivationInterval != nil {
					if th.ActivationInterval.ActivationTime != "" {
						mdl.ActivationInterval = th.ActivationInterval.ActivationTime
					}
					if th.ActivationInterval.ExpiryTime != "" {
						mdl.ActivationInterval += utils.INFIELD_SEP + th.ActivationInterval.ExpiryTime
					}
				}
			}
			mdl.FilterIDs = th.FilterIDs[i]
			mdl.ActionIDs = th.ActionIDs[i]
			mdls = append(mdls, mdl)
		}

		if len(th.FilterIDs)-min > 0 {
			for i := min; i < len(th.FilterIDs); i++ {
				mdl := &TpThreshold{
					Tpid:   th.TPid,
					Tenant: th.Tenant,
					ID:     th.ID,
				}
				mdl.FilterIDs = th.FilterIDs[i]
				mdls = append(mdls, mdl)
			}
		}
		if len(th.ActionIDs)-min > 0 {
			for i := min; i < len(th.ActionIDs); i++ {
				mdl := &TpThreshold{
					Tpid:   th.TPid,
					Tenant: th.Tenant,
					ID:     th.ID,
				}
				if min == 0 && i == 0 {
					mdl.Blocker = th.Blocker
					mdl.Weight = th.Weight
					mdl.MaxHits = th.MaxHits
					mdl.MinHits = th.MinHits
					mdl.MinSleep = th.MinSleep
					mdl.Async = th.Async
					if th.ActivationInterval != nil {
						if th.ActivationInterval.ActivationTime != "" {
							mdl.ActivationInterval = th.ActivationInterval.ActivationTime
						}
						if th.ActivationInterval.ExpiryTime != "" {
							mdl.ActivationInterval += utils.INFIELD_SEP + th.ActivationInterval.ExpiryTime
						}
					}
				}
				mdl.ActionIDs = th.ActionIDs[i]
				mdls = append(mdls, mdl)
			}
		}
	}
	return
}

func APItoThresholdProfile(tpTH *utils.TPThresholdProfile, timezone string) (th *ThresholdProfile, err error) {
	th = &ThresholdProfile{
		Tenant:    tpTH.Tenant,
		ID:        tpTH.ID,
		MaxHits:   tpTH.MaxHits,
		MinHits:   tpTH.MinHits,
		Weight:    tpTH.Weight,
		Blocker:   tpTH.Blocker,
		Async:     tpTH.Async,
		ActionIDs: make([]string, len(tpTH.ActionIDs)),
		FilterIDs: make([]string, len(tpTH.FilterIDs)),
	}
	if tpTH.MinSleep != "" {
		if th.MinSleep, err = utils.ParseDurationWithNanosecs(tpTH.MinSleep); err != nil {
			return nil, err
		}
	}
	for i, ati := range tpTH.ActionIDs {
		th.ActionIDs[i] = ati

	}
	for i, fli := range tpTH.FilterIDs {
		th.FilterIDs[i] = fli
	}
	if tpTH.ActivationInterval != nil {
		if th.ActivationInterval, err = tpTH.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	return th, nil
}

type TpFilterS []*TpFilter

func (tps TpFilterS) AsTPFilter() (result []*utils.TPFilterProfile) {
	mst := make(map[string]*utils.TPFilterProfile)
	for _, tp := range tps {
		th, found := mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]
		if !found {
			th = &utils.TPFilterProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if len(tp.ActivationInterval) != 0 {
			th.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterType != "" {
			th.Filters = append(th.Filters, &utils.TPFilter{
				Type:      tp.FilterType,
				FieldName: tp.FilterFieldName,
				Values:    strings.Split(tp.FilterFieldValues, utils.INFIELD_SEP)})
		}
		mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = th
	}
	result = make([]*utils.TPFilterProfile, len(mst))
	i := 0
	for _, th := range mst {
		result[i] = th
		i++
	}
	return
}

func APItoModelTPFilter(th *utils.TPFilterProfile) (mdls TpFilterS) {
	if len(th.Filters) == 0 {
		return
	}
	for _, fltr := range th.Filters {
		mdl := &TpFilter{
			Tpid:   th.TPid,
			Tenant: th.Tenant,
			ID:     th.ID,
		}
		mdl.FilterType = fltr.Type
		mdl.FilterFieldName = fltr.FieldName
		if th.ActivationInterval != nil {
			if th.ActivationInterval.ActivationTime != "" {
				mdl.ActivationInterval = th.ActivationInterval.ActivationTime
			}
			if th.ActivationInterval.ExpiryTime != "" {
				mdl.ActivationInterval += utils.INFIELD_SEP + th.ActivationInterval.ExpiryTime
			}
		}
		for i, val := range fltr.Values {
			if i != 0 {
				mdl.FilterFieldValues += utils.INFIELD_SEP
			}
			mdl.FilterFieldValues += val
		}
		mdls = append(mdls, mdl)
	}
	return
}

func APItoFilter(tpTH *utils.TPFilterProfile, timezone string) (th *Filter, err error) {
	th = &Filter{
		Tenant: tpTH.Tenant,
		ID:     tpTH.ID,
		Rules:  make([]*FilterRule, len(tpTH.Filters)),
	}
	for i, f := range tpTH.Filters {
		rf := &FilterRule{Type: f.Type, FieldName: f.FieldName, Values: f.Values}
		if err := rf.CompileValues(); err != nil {
			return nil, err
		}
		th.Rules[i] = rf
	}
	if tpTH.ActivationInterval != nil {
		if th.ActivationInterval, err = tpTH.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	return th, nil
}

func FilterToTPFilter(f *Filter) (tpFltr *utils.TPFilterProfile) {
	tpFltr = &utils.TPFilterProfile{
		Tenant:  f.Tenant,
		ID:      f.ID,
		Filters: make([]*utils.TPFilter, len(f.Rules)),
	}
	for i, reqFltr := range f.Rules {
		tpFltr.Filters[i] = &utils.TPFilter{
			Type:      reqFltr.Type,
			FieldName: reqFltr.FieldName,
			Values:    make([]string, len(reqFltr.Values)),
		}
		for j, val := range reqFltr.Values {
			tpFltr.Filters[i].Values[j] = val
		}
	}
	if f.ActivationInterval != nil {
		tpFltr.ActivationInterval = &utils.TPActivationInterval{
			ActivationTime: f.ActivationInterval.ActivationTime.Format(time.RFC3339),
			ExpiryTime:     f.ActivationInterval.ExpiryTime.Format(time.RFC3339),
		}
	}
	return
}

type TpSuppliers []*TpSupplier

func (tps TpSuppliers) AsTPSuppliers() (result []*utils.TPSupplierProfile) {
	filtermap := make(map[string]utils.StringMap)
	mst := make(map[string]*utils.TPSupplierProfile)
	suppliersMap := make(map[string]map[string]*utils.TPSupplier)
	sortingParameterMap := make(map[string]utils.StringMap)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
		if !found {
			th = &utils.TPSupplierProfile{
				TPid:              tp.Tpid,
				Tenant:            tp.Tenant,
				ID:                tp.ID,
				Sorting:           tp.Sorting,
				SortingParameters: []string{},
			}
		}
		if tp.SupplierID != "" {
			if _, has := suppliersMap[tenID]; !has {
				suppliersMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = make(map[string]*utils.TPSupplier)
			}
			sup, found := suppliersMap[tenID][tp.SupplierID]
			if !found {
				sup = &utils.TPSupplier{
					ID:      tp.SupplierID,
					Weight:  tp.SupplierWeight,
					Blocker: tp.SupplierBlocker,
				}
			}
			if tp.SupplierParameters != "" {
				sup.SupplierParameters = tp.SupplierParameters
			}
			if tp.SupplierFilterIDs != "" {
				supFilterSplit := strings.Split(tp.SupplierFilterIDs, utils.INFIELD_SEP)
				sup.FilterIDs = append(sup.FilterIDs, supFilterSplit...)
			}
			if tp.SupplierRatingplanIDs != "" {
				ratingPlanSplit := strings.Split(tp.SupplierRatingplanIDs, utils.INFIELD_SEP)
				sup.RatingPlanIDs = append(sup.RatingPlanIDs, ratingPlanSplit...)
			}
			if tp.SupplierResourceIDs != "" {
				resSplit := strings.Split(tp.SupplierResourceIDs, utils.INFIELD_SEP)
				sup.ResourceIDs = append(sup.ResourceIDs, resSplit...)
			}
			if tp.SupplierStatIDs != "" {
				statSplit := strings.Split(tp.SupplierStatIDs, utils.INFIELD_SEP)
				sup.StatIDs = append(sup.StatIDs, statSplit...)
			}
			if tp.SupplierAccountIDs != "" {
				accSplit := strings.Split(tp.SupplierAccountIDs, utils.INFIELD_SEP)
				sup.AccountIDs = append(sup.AccountIDs, accSplit...)
			}
			suppliersMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()][tp.SupplierID] = sup
		}
		if tp.SortingParameters != "" {
			if _, has := sortingParameterMap[tenID]; !has {
				sortingParameterMap[tenID] = make(utils.StringMap)
			}
			sortingParamSplit := strings.Split(tp.SortingParameters, utils.INFIELD_SEP)
			for _, sortingParam := range sortingParamSplit {
				sortingParameterMap[tenID][sortingParam] = true
			}
		}
		if tp.Weight != 0 {
			th.Weight = tp.Weight
		}
		if len(tp.ActivationInterval) != 0 {
			th.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != "" {
			if _, has := filtermap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]; !has {
				filtermap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = make(utils.StringMap)
			}
			filterSplit := strings.Split(tp.FilterIDs, utils.INFIELD_SEP)
			for _, filter := range filterSplit {
				filtermap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()][filter] = true
			}
		}
		mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = th
	}
	result = make([]*utils.TPSupplierProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		for _, supdata := range suppliersMap[tntID] {
			result[i].Suppliers = append(result[i].Suppliers, supdata)
		}
		for filterdata := range filtermap[tntID] {
			result[i].FilterIDs = append(result[i].FilterIDs, filterdata)
		}
		for sortingParam := range sortingParameterMap[tntID] {
			result[i].SortingParameters = append(result[i].SortingParameters, sortingParam)
		}
		i++
	}
	return
}

func APItoModelTPSuppliers(st *utils.TPSupplierProfile) (mdls TpSuppliers) {
	if len(st.Suppliers) == 0 {
		return
	}
	for i, supl := range st.Suppliers {
		mdl := &TpSupplier{
			Tenant: st.Tenant,
			Tpid:   st.TPid,
			ID:     st.ID,
		}
		if i == 0 {
			mdl.Sorting = st.Sorting
			mdl.Weight = st.Weight
			for i, val := range st.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.INFIELD_SEP
				}
				mdl.FilterIDs += val
			}
			for i, val := range st.SortingParameters {
				if i != 0 {
					mdl.SortingParameters += utils.INFIELD_SEP
				}
				mdl.SortingParameters += val
			}
			if st.ActivationInterval != nil {
				if st.ActivationInterval.ActivationTime != "" {
					mdl.ActivationInterval = st.ActivationInterval.ActivationTime
				}
				if st.ActivationInterval.ExpiryTime != "" {
					mdl.ActivationInterval += utils.INFIELD_SEP + st.ActivationInterval.ExpiryTime
				}
			}
		}
		mdl.SupplierID = supl.ID
		for i, val := range supl.AccountIDs {
			if i != 0 {
				mdl.SupplierAccountIDs += utils.INFIELD_SEP
			}
			mdl.SupplierAccountIDs += val
		}
		for i, val := range supl.RatingPlanIDs {
			if i != 0 {
				mdl.SupplierRatingplanIDs += utils.INFIELD_SEP
			}
			mdl.SupplierRatingplanIDs += val
		}
		for i, val := range supl.FilterIDs {
			if i != 0 {
				mdl.SupplierFilterIDs += utils.INFIELD_SEP
			}
			mdl.SupplierFilterIDs += val
		}
		for i, val := range supl.ResourceIDs {
			if i != 0 {
				mdl.SupplierResourceIDs += utils.INFIELD_SEP
			}
			mdl.SupplierResourceIDs += val
		}
		for i, val := range supl.StatIDs {
			if i != 0 {
				mdl.SupplierStatIDs += utils.INFIELD_SEP
			}
			mdl.SupplierStatIDs += val
		}
		mdl.SupplierWeight = supl.Weight
		mdl.SupplierParameters = supl.SupplierParameters
		mdl.SupplierBlocker = supl.Blocker
		mdls = append(mdls, mdl)
	}
	return
}

func APItoSupplierProfile(tpSPP *utils.TPSupplierProfile, timezone string) (spp *SupplierProfile, err error) {
	spp = &SupplierProfile{
		Tenant:            tpSPP.Tenant,
		ID:                tpSPP.ID,
		Sorting:           tpSPP.Sorting,
		Weight:            tpSPP.Weight,
		Suppliers:         make([]*Supplier, len(tpSPP.Suppliers)),
		SortingParameters: make([]string, len(tpSPP.SortingParameters)),
		FilterIDs:         make([]string, len(tpSPP.FilterIDs)),
	}
	for i, stp := range tpSPP.SortingParameters {
		spp.SortingParameters[i] = stp
	}
	for i, fli := range tpSPP.FilterIDs {
		spp.FilterIDs[i] = fli
	}
	if tpSPP.ActivationInterval != nil {
		if spp.ActivationInterval, err = tpSPP.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	for i, suplier := range tpSPP.Suppliers {
		spp.Suppliers[i] = &Supplier{
			ID:                 suplier.ID,
			Weight:             suplier.Weight,
			Blocker:            suplier.Blocker,
			RatingPlanIDs:      suplier.RatingPlanIDs,
			FilterIDs:          suplier.FilterIDs,
			ResourceIDs:        suplier.ResourceIDs,
			StatIDs:            suplier.StatIDs,
			SupplierParameters: suplier.SupplierParameters,
		}
	}
	return spp, nil
}

type TPAttributes []*TPAttribute

func (tps TPAttributes) AsTPAttributes() (result []*utils.TPAttributeProfile) {
	mst := make(map[string]*utils.TPAttributeProfile)
	filterMap := make(map[string]utils.StringMap)
	contextMap := make(map[string]utils.StringMap)
	for _, tp := range tps {
		key := &utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}
		th, found := mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]
		if !found {
			th = &utils.TPAttributeProfile{
				TPid:    tp.Tpid,
				Tenant:  tp.Tenant,
				ID:      tp.ID,
				Blocker: tp.Blocker,
			}
		}
		if tp.Weight != 0 {
			th.Weight = tp.Weight
		}
		if len(tp.ActivationInterval) != 0 {
			th.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != "" {
			if _, has := filterMap[key.TenantID()]; !has {
				filterMap[key.TenantID()] = make(utils.StringMap)
			}
			filterSplit := strings.Split(tp.FilterIDs, utils.INFIELD_SEP)
			for _, filter := range filterSplit {
				filterMap[key.TenantID()][filter] = true
			}
		}
		if tp.Contexts != "" {
			if _, has := contextMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]; !has {
				contextMap[key.TenantID()] = make(utils.StringMap)
			}
			contextSplit := strings.Split(tp.Contexts, utils.INFIELD_SEP)
			for _, context := range contextSplit {
				contextMap[key.TenantID()][context] = true
			}
		}
		if tp.FieldName != "" {
			filterIDs := make([]string, 0)
			if tp.AttributeFilterIDs != "" {
				filterAttrSplit := strings.Split(tp.AttributeFilterIDs, utils.INFIELD_SEP)
				for _, filterAttr := range filterAttrSplit {
					filterIDs = append(filterIDs, filterAttr)
				}
			}
			th.Attributes = append(th.Attributes, &utils.TPAttribute{
				FilterIDs: filterIDs,
				Type:      tp.Type,
				FieldName: tp.FieldName,
				Value:     tp.Value,
			})
		}
		mst[key.TenantID()] = th
	}
	result = make([]*utils.TPAttributeProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		for filterID := range filterMap[tntID] {
			result[i].FilterIDs = append(result[i].FilterIDs, filterID)
		}
		for context := range contextMap[tntID] {
			result[i].Contexts = append(result[i].Contexts, context)
		}
		i++
	}
	return
}

func APItoModelTPAttribute(th *utils.TPAttributeProfile) (mdls TPAttributes) {
	if len(th.Attributes) == 0 {
		return
	}
	for i, reqAttribute := range th.Attributes {
		mdl := &TPAttribute{
			Tpid:   th.TPid,
			Tenant: th.Tenant,
			ID:     th.ID,
		}
		if i == 0 {
			mdl.Blocker = th.Blocker
			if th.ActivationInterval != nil {
				if th.ActivationInterval.ActivationTime != "" {
					mdl.ActivationInterval = th.ActivationInterval.ActivationTime
				}
				if th.ActivationInterval.ExpiryTime != "" {
					mdl.ActivationInterval += utils.INFIELD_SEP + th.ActivationInterval.ExpiryTime
				}
			}
			for i, val := range th.Contexts {
				if i != 0 {
					mdl.Contexts += utils.INFIELD_SEP
				}
				mdl.Contexts += val
			}
			for i, val := range th.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.INFIELD_SEP
				}
				mdl.FilterIDs += val
			}
			if th.Weight != 0 {
				mdl.Weight = th.Weight
			}
			for i, val := range reqAttribute.FilterIDs {
				if i != 0 {
					mdl.AttributeFilterIDs += utils.INFIELD_SEP
				}
				mdl.AttributeFilterIDs += val
			}
		}
		mdl.FieldName = reqAttribute.FieldName
		mdl.Value = reqAttribute.Value
		mdl.Type = reqAttribute.Type
		mdls = append(mdls, mdl)
	}
	return
}

func APItoAttributeProfile(tpAttr *utils.TPAttributeProfile, timezone string) (attrPrf *AttributeProfile, err error) {
	attrPrf = &AttributeProfile{
		Tenant:     tpAttr.Tenant,
		ID:         tpAttr.ID,
		Weight:     tpAttr.Weight,
		Blocker:    tpAttr.Blocker,
		FilterIDs:  make([]string, len(tpAttr.FilterIDs)),
		Contexts:   make([]string, len(tpAttr.Contexts)),
		Attributes: make([]*Attribute, len(tpAttr.Attributes)),
	}
	for i, fli := range tpAttr.FilterIDs {
		attrPrf.FilterIDs[i] = fli
	}
	for i, context := range tpAttr.Contexts {
		attrPrf.Contexts[i] = context
	}
	for i, reqAttr := range tpAttr.Attributes {
		sbstPrsr, err := config.NewRSRParsers(reqAttr.Value, true, config.CgrConfig().GeneralCfg().RsrSepatarot)
		if err != nil {
			return nil, err
		}
		attrPrf.Attributes[i] = &Attribute{
			FilterIDs: reqAttr.FilterIDs,
			FieldName: reqAttr.FieldName,
			Type:      reqAttr.Type,
			Value:     sbstPrsr,
		}
	}
	if tpAttr.ActivationInterval != nil {
		if attrPrf.ActivationInterval, err = tpAttr.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	return attrPrf, nil
}

type TPChargers []*TPCharger

func (tps TPChargers) AsTPChargers() (result []*utils.TPChargerProfile) {
	mst := make(map[string]*utils.TPChargerProfile)
	filterMap := make(map[string]utils.StringMap)
	attributeMap := make(map[string]utils.StringMap)
	for _, tp := range tps {
		tpCPP, found := mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]
		if !found {
			tpCPP = &utils.TPChargerProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.Weight != 0 {
			tpCPP.Weight = tp.Weight
		}
		if len(tp.ActivationInterval) != 0 {
			tpCPP.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				tpCPP.ActivationInterval.ActivationTime = aiSplt[0]
				tpCPP.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				tpCPP.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != "" {
			if _, has := filterMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]; !has {
				filterMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = make(utils.StringMap)
			}
			filterSplit := strings.Split(tp.FilterIDs, utils.INFIELD_SEP)
			for _, filter := range filterSplit {
				filterMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()][filter] = true
			}
		}
		if tp.RunID != "" {
			tpCPP.RunID = tp.RunID
		}
		if tp.AttributeIDs != "" {
			if _, has := attributeMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()]; !has {
				attributeMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = make(utils.StringMap)
			}
			attributeSplit := strings.Split(tp.AttributeIDs, utils.INFIELD_SEP)
			for _, attribute := range attributeSplit {
				attributeMap[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()][attribute] = true
			}
		}
		mst[(&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()] = tpCPP
	}
	result = make([]*utils.TPChargerProfile, len(mst))
	i := 0
	for tntID, tp := range mst {
		result[i] = tp
		for filterID := range filterMap[tntID] {
			result[i].FilterIDs = append(result[i].FilterIDs, filterID)
		}
		for attributeID := range attributeMap[tntID] {
			result[i].AttributeIDs = append(result[i].AttributeIDs, attributeID)
		}
		i++
	}
	return
}

func APItoModelTPCharger(tpCPP *utils.TPChargerProfile) (mdls TPChargers) {
	if tpCPP != nil {
		min := len(tpCPP.FilterIDs)
		isFilter := true
		if min > len(tpCPP.AttributeIDs) {
			min = len(tpCPP.AttributeIDs)
			isFilter = false
		}
		if min == 0 {
			mdl := &TPCharger{
				Tenant: tpCPP.Tenant,
				Tpid:   tpCPP.TPid,
				ID:     tpCPP.ID,
				Weight: tpCPP.Weight,
				RunID:  tpCPP.RunID,
			}
			if tpCPP.ActivationInterval != nil {
				if tpCPP.ActivationInterval.ActivationTime != "" {
					mdl.ActivationInterval = tpCPP.ActivationInterval.ActivationTime
				}
				if tpCPP.ActivationInterval.ExpiryTime != "" {
					mdl.ActivationInterval += utils.INFIELD_SEP + tpCPP.ActivationInterval.ExpiryTime
				}
			}
			if isFilter && len(tpCPP.AttributeIDs) > 0 {
				mdl.AttributeIDs = tpCPP.AttributeIDs[0]
			} else if len(tpCPP.FilterIDs) > 0 {
				mdl.FilterIDs = tpCPP.FilterIDs[0]
			}
			min = 1
			mdls = append(mdls, mdl)
		} else {
			for i := 0; i < min; i++ {
				mdl := &TPCharger{
					Tenant: tpCPP.Tenant,
					Tpid:   tpCPP.TPid,
					ID:     tpCPP.ID,
				}
				if i == 0 {
					mdl.Weight = tpCPP.Weight
					mdl.RunID = tpCPP.RunID
					if tpCPP.ActivationInterval != nil {
						if tpCPP.ActivationInterval.ActivationTime != "" {
							mdl.ActivationInterval = tpCPP.ActivationInterval.ActivationTime
						}
						if tpCPP.ActivationInterval.ExpiryTime != "" {
							mdl.ActivationInterval += utils.INFIELD_SEP + tpCPP.ActivationInterval.ExpiryTime
						}
					}
				}
				mdl.AttributeIDs = tpCPP.AttributeIDs[i]
				mdl.FilterIDs = tpCPP.FilterIDs[i]
				mdls = append(mdls, mdl)
			}
		}
		if len(tpCPP.FilterIDs)-min > 0 {
			for i := min; i < len(tpCPP.FilterIDs); i++ {
				mdl := &TPCharger{
					Tenant: tpCPP.Tenant,
					Tpid:   tpCPP.TPid,
					ID:     tpCPP.ID,
				}
				mdl.FilterIDs = tpCPP.FilterIDs[i]
				mdls = append(mdls, mdl)
			}
		}
		if len(tpCPP.AttributeIDs)-min > 0 {
			for i := min; i < len(tpCPP.AttributeIDs); i++ {
				mdl := &TPCharger{
					Tenant: tpCPP.Tenant,
					Tpid:   tpCPP.TPid,
					ID:     tpCPP.ID,
				}
				mdl.AttributeIDs = tpCPP.AttributeIDs[i]
				mdls = append(mdls, mdl)
			}
		}

	}
	return
}

func APItoChargerProfile(tpCPP *utils.TPChargerProfile, timezone string) (cpp *ChargerProfile, err error) {
	cpp = &ChargerProfile{
		Tenant:       tpCPP.Tenant,
		ID:           tpCPP.ID,
		Weight:       tpCPP.Weight,
		RunID:        tpCPP.RunID,
		FilterIDs:    make([]string, len(tpCPP.FilterIDs)),
		AttributeIDs: make([]string, len(tpCPP.AttributeIDs)),
	}
	for i, fli := range tpCPP.FilterIDs {
		cpp.FilterIDs[i] = fli
	}
	for i, attribute := range tpCPP.AttributeIDs {
		cpp.AttributeIDs[i] = attribute
	}
	if tpCPP.ActivationInterval != nil {
		if cpp.ActivationInterval, err = tpCPP.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	return cpp, nil
}

type TPDispatcherProfiles []*TPDispatcherProfile

func (tps TPDispatcherProfiles) AsTPDispatcherProfiles() (result []*utils.TPDispatcherProfile) {
	mst := make(map[string]*utils.TPDispatcherProfile)
	filterMap := make(map[string]utils.StringMap)
	contextMap := make(map[string]utils.StringMap)
	connsMap := make(map[string]map[string]utils.TPDispatcherHostProfile)
	connsFilterMap := make(map[string]map[string]utils.StringMap)
	for _, tp := range tps {
		tenantID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		tpDPP, found := mst[tenantID]
		if !found {
			tpDPP = &utils.TPDispatcherProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.Subsystems != "" {
			if _, has := contextMap[tenantID]; !has {
				contextMap[tenantID] = make(utils.StringMap)
			}
			contextSplit := strings.Split(tp.Subsystems, utils.INFIELD_SEP)
			for _, context := range contextSplit {
				contextMap[tenantID][context] = true
			}
		}
		if tp.FilterIDs != "" {
			if _, has := filterMap[tenantID]; !has {
				filterMap[tenantID] = make(utils.StringMap)
			}
			filterSplit := strings.Split(tp.FilterIDs, utils.INFIELD_SEP)
			for _, filter := range filterSplit {
				filterMap[tenantID][filter] = true
			}
		}
		if len(tp.ActivationInterval) != 0 {
			tpDPP.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.INFIELD_SEP)
			if len(aiSplt) == 2 {
				tpDPP.ActivationInterval.ActivationTime = aiSplt[0]
				tpDPP.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				tpDPP.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.Strategy != "" {
			tpDPP.Strategy = tp.Strategy
		}
		if tp.StrategyParameters != "" {
			for _, param := range strings.Split(tp.FilterIDs, utils.INFIELD_SEP) {
				tpDPP.StrategyParams = append(tpDPP.StrategyParams, param)
			}
		}
		if tp.ConnID != "" {
			if _, has := connsMap[tenantID]; !has {
				connsMap[tenantID] = make(map[string]utils.TPDispatcherHostProfile)
			}
			conn, has := connsMap[tenantID][tp.ConnID]
			if !has {
				conn = utils.TPDispatcherHostProfile{
					ID:      tp.ConnID,
					Weight:  tp.ConnWeight,
					Blocker: tp.ConnBlocker,
				}
			}
			for _, param := range strings.Split(tp.ConnParameters, utils.INFIELD_SEP) {
				conn.Params = append(conn.Params, param)
			}
			connsMap[tenantID][tp.ConnID] = conn

			if dFilter, has := connsFilterMap[tenantID]; !has {
				connsFilterMap[tenantID] = make(map[string]utils.StringMap)
				connsFilterMap[tenantID][tp.ConnID] = make(utils.StringMap)
			} else if _, has := dFilter[tp.ConnID]; !has {
				connsFilterMap[tenantID][tp.ConnID] = make(utils.StringMap)
			}

			for _, filter := range strings.Split(tp.ConnFilterIDs, utils.INFIELD_SEP) {
				connsFilterMap[tenantID][tp.ConnID][filter] = true
			}

		}
		if tp.Weight != 0 {
			tpDPP.Weight = tp.Weight
		}
		mst[tenantID] = tpDPP
	}
	result = make([]*utils.TPDispatcherProfile, len(mst))
	i := 0
	for tntID, tp := range mst {
		result[i] = tp
		for filterID := range filterMap[tntID] {
			if filterID != "" {
				result[i].FilterIDs = append(result[i].FilterIDs, filterID)
			}
		}
		for context := range contextMap[tntID] {
			if context != "" {
				result[i].Subsystems = append(result[i].Subsystems, context)
			}
		}
		for conID, conn := range connsMap[tntID] {
			for filter := range connsFilterMap[tntID][conID] {
				if filter != "" {
					conn.FilterIDs = append(conn.FilterIDs, filter)
				}
			}
			result[i].Hosts = append(result[i].Hosts,
				&utils.TPDispatcherHostProfile{
					ID:        conn.ID,
					FilterIDs: conn.FilterIDs,
					Weight:    conn.Weight,
					Params:    conn.Params,
					Blocker:   conn.Blocker,
				})
		}
		i++
	}
	return
}

func paramsToString(sp []interface{}) (strategy string) {
	if len(sp) != 0 {
		strategy = sp[0].(string)
		for i := 1; i < len(sp); i++ {
			strategy += utils.INFIELD_SEP + sp[i].(string)
		}
	}
	return
}

func APItoModelTPDispatcherProfile(tpDPP *utils.TPDispatcherProfile) (mdls TPDispatcherProfiles) {
	if tpDPP == nil {
		return
	}

	filters := strings.Join(tpDPP.FilterIDs, utils.INFIELD_SEP)
	subsystems := strings.Join(tpDPP.Subsystems, utils.INFIELD_SEP)

	interval := ""
	if tpDPP.ActivationInterval != nil {
		if tpDPP.ActivationInterval.ActivationTime != "" {
			interval = tpDPP.ActivationInterval.ActivationTime
		}
		if tpDPP.ActivationInterval.ExpiryTime != "" {
			interval += utils.INFIELD_SEP + tpDPP.ActivationInterval.ExpiryTime
		}
	}

	strategy := paramsToString(tpDPP.StrategyParams)

	if len(tpDPP.Hosts) == 0 {
		return append(mdls, &TPDispatcherProfile{
			Tpid:               tpDPP.TPid,
			Tenant:             tpDPP.Tenant,
			ID:                 tpDPP.ID,
			Subsystems:         subsystems,
			FilterIDs:          filters,
			ActivationInterval: interval,
			Strategy:           tpDPP.Strategy,
			StrategyParameters: strategy,
			Weight:             tpDPP.Weight,
		})
	}

	confilter := strings.Join(tpDPP.Hosts[0].FilterIDs, utils.INFIELD_SEP)
	conparam := paramsToString(tpDPP.Hosts[0].Params)

	mdls = append(mdls, &TPDispatcherProfile{
		Tpid:               tpDPP.TPid,
		Tenant:             tpDPP.Tenant,
		ID:                 tpDPP.ID,
		Subsystems:         subsystems,
		FilterIDs:          filters,
		ActivationInterval: interval,
		Strategy:           tpDPP.Strategy,
		StrategyParameters: strategy,
		Weight:             tpDPP.Weight,

		ConnID:         tpDPP.Hosts[0].ID,
		ConnFilterIDs:  confilter,
		ConnWeight:     tpDPP.Hosts[0].Weight,
		ConnBlocker:    tpDPP.Hosts[0].Blocker,
		ConnParameters: conparam,
	})
	for i := 1; i < len(tpDPP.Hosts); i++ {
		confilter = strings.Join(tpDPP.Hosts[i].FilterIDs, utils.INFIELD_SEP)
		conparam = paramsToString(tpDPP.Hosts[i].Params)
		mdls = append(mdls, &TPDispatcherProfile{
			Tpid:   tpDPP.TPid,
			Tenant: tpDPP.Tenant,
			ID:     tpDPP.ID,

			ConnID:         tpDPP.Hosts[i].ID,
			ConnFilterIDs:  confilter,
			ConnWeight:     tpDPP.Hosts[i].Weight,
			ConnBlocker:    tpDPP.Hosts[i].Blocker,
			ConnParameters: conparam,
		})
	}
	return
}

func APItoDispatcherProfile(tpDPP *utils.TPDispatcherProfile, timezone string) (dpp *DispatcherProfile, err error) {
	dpp = &DispatcherProfile{
		Tenant:         tpDPP.Tenant,
		ID:             tpDPP.ID,
		Weight:         tpDPP.Weight,
		Strategy:       tpDPP.Strategy,
		FilterIDs:      make([]string, len(tpDPP.FilterIDs)),
		Subsystems:     make([]string, len(tpDPP.Subsystems)),
		StrategyParams: make(map[string]interface{}),
		Hosts:          make(DispatcherHostProfiles, len(tpDPP.Hosts)),
	}
	for i, fli := range tpDPP.FilterIDs {
		dpp.FilterIDs[i] = fli
	}
	for i, sub := range tpDPP.Subsystems {
		dpp.Subsystems[i] = sub
	}
	for i, param := range tpDPP.StrategyParams {
		if param != "" {
			dpp.StrategyParams[strconv.Itoa(i)] = param
		}
	}
	for i, conn := range tpDPP.Hosts {
		dpp.Hosts[i] = &DispatcherHostProfile{
			ID:        conn.ID,
			Weight:    conn.Weight,
			Blocker:   conn.Blocker,
			FilterIDs: make([]string, len(conn.FilterIDs)),
			Params:    make(map[string]interface{}),
		}
		for j, fltr := range conn.FilterIDs {
			dpp.Hosts[i].FilterIDs[j] = fltr
		}
		for j, param := range conn.Params {
			if param != "" {
				dpp.Hosts[i].Params[strconv.Itoa(j)] = param
			}
		}
	}
	if tpDPP.ActivationInterval != nil {
		if dpp.ActivationInterval, err = tpDPP.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	return dpp, nil
}

// TPHosts
type TPDispatcherHosts []*TPDispatcherHost

func (tps TPDispatcherHosts) AsTPDispatcherHosts() (result []*utils.TPDispatcherHost) {
	hostsMap := make(map[string]*utils.TPDispatcherHost)
	for _, tp := range tps {
		if len(tp.Address) == 0 { // empty addres do not populate conns
			continue
		}
		if len(tp.Transport) == 0 {
			tp.Transport = utils.MetaJSONrpc
		}
		tenantID := utils.ConcatenatedKey(tp.Tenant, tp.ID)
		if th, has := hostsMap[tenantID]; !has || th == nil {
			hostsMap[tenantID] = &utils.TPDispatcherHost{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
				Conns: []*utils.TPDispatcherHostConn{
					{
						Address:   tp.Address,
						Transport: tp.Transport,
						TLS:       tp.TLS,
					},
				},
			}
			continue
		}
		hostsMap[tenantID].Conns = append(hostsMap[tenantID].Conns, &utils.TPDispatcherHostConn{
			Address:   tp.Address,
			Transport: tp.Transport,
			TLS:       tp.TLS,
		})
	}
	for _, host := range hostsMap {
		result = append(result, host)
	}
	return
}

func APItoModelTPDispatcherHost(tpDPH *utils.TPDispatcherHost) (mdls TPDispatcherHosts) {
	if tpDPH == nil {
		return
	}
	mdls = make(TPDispatcherHosts, len(tpDPH.Conns))
	for i, conn := range tpDPH.Conns {
		mdls[i] = &TPDispatcherHost{
			Tpid:      tpDPH.TPid,
			Tenant:    tpDPH.Tenant,
			ID:        tpDPH.ID,
			Address:   conn.Address,
			Transport: conn.Transport,
			TLS:       conn.TLS,
		}
	}
	return
}

func APItoDispatcherHost(tpDPH *utils.TPDispatcherHost) (dpp *DispatcherHost) {
	if tpDPH == nil {
		return
	}
	dpp = &DispatcherHost{
		Tenant: tpDPH.Tenant,
		ID:     tpDPH.ID,
		Conns:  make([]*config.RemoteHost, len(tpDPH.Conns)),
	}
	for i, conn := range tpDPH.Conns {
		dpp.Conns[i] = &config.RemoteHost{
			Address:   conn.Address,
			Transport: conn.Transport,
			TLS:       conn.TLS,
		}
	}
	return
}
