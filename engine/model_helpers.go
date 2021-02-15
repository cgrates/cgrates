/*
/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
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
		if index != utils.EmptyString {
			idx, err := strconv.Atoi(index)
			if err != nil || len(values) <= idx {
				return nil, fmt.Errorf("invalid %v.%v index %v", st.Name(), field.Name, index)
			}
			if re != utils.EmptyString {
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
				if fieldValue == utils.EmptyString {
					fieldValue = "0"
				}
				value, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetFloat(value)
			case reflect.Int:
				if fieldValue == utils.EmptyString {
					fieldValue = "0"
				}
				value, err := strconv.Atoi(fieldValue)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fieldName)
				}
				field.SetInt(int64(value))
			case reflect.Bool:
				if fieldValue == utils.EmptyString {
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

//CsvDump receive and interface and convert it to a slice of string
func CsvDump(s interface{}) ([]string, error) {
	fieldIndexMap := make(map[string]int)
	st := reflect.ValueOf(s)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		s = st.Interface()
	}
	numFields := st.NumField()
	stCopy := reflect.TypeOf(s)
	for i := 0; i < numFields; i++ {
		field := stCopy.Field(i)
		index := field.Tag.Get("index")
		if index != utils.EmptyString {
			if idx, err := strconv.Atoi(index); err != nil {
				return nil, fmt.Errorf("invalid %v.%v index %v", stCopy.Name(), field.Name, index)
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

func getColumnCount(s interface{}) int {
	st := reflect.TypeOf(s)
	numFields := st.NumField()
	count := 0
	for i := 0; i < numFields; i++ {
		field := st.Field(i)
		index := field.Tag.Get("index")
		if index != utils.EmptyString {
			count++
		}
	}
	return count
}

type DestinationMdls []DestinationMdl

func (tps DestinationMdls) AsMapDestinations() (map[string]*Destination, error) {
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

// AsTPDestination converts DestinationMdls  into *utils.TPDestination
func (tps DestinationMdls) AsTPDestinations() (result []*utils.TPDestination) {
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

func APItoModelDestination(d *utils.TPDestination) (result DestinationMdls) {
	if d != nil {
		for _, p := range d.Prefixes {
			result = append(result, DestinationMdl{
				Tpid:   d.TPid,
				Tag:    d.ID,
				Prefix: p,
			})
		}
		if len(d.Prefixes) == 0 {
			result = append(result, DestinationMdl{
				Tpid: d.TPid,
				Tag:  d.ID,
			})
		}
	}
	return
}

type TimingMdls []TimingMdl

func (tps TimingMdls) AsMapTPTimings() (map[string]*utils.ApierTPTiming, error) {
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
		t := utils.NewTiming(tp.ID, tp.Years, tp.Months, tp.MonthDays, tp.WeekDays, tp.Time)
		if _, found := result[tp.ID]; found {
			return nil, fmt.Errorf("duplicate timing tag: %s", tp.ID)
		}
		result[tp.ID] = t
	}
	return result, nil
}

func (tps TimingMdls) AsTPTimings() (result []*utils.ApierTPTiming) {
	ats, _ := tps.AsMapTPTimings()
	for _, tp := range ats {
		result = append(result, tp)
	}
	return result
}

func APItoModelTiming(t *utils.ApierTPTiming) (result TimingMdl) {
	return TimingMdl{
		Tpid:      t.TPid,
		Tag:       t.ID,
		Years:     t.Years,
		Months:    t.Months,
		MonthDays: t.MonthDays,
		WeekDays:  t.WeekDays,
		Time:      t.Time,
	}
}

func APItoModelTimings(ts []*utils.ApierTPTiming) (result TimingMdls) {
	for _, t := range ts {
		if t != nil {
			at := APItoModelTiming(t)
			result = append(result, at)
		}
	}
	return result
}

type RateMdls []RateMdl

func (tps RateMdls) AsMapRates() (map[string]*utils.TPRateRALs, error) {
	result := make(map[string]*utils.TPRateRALs)
	for _, tp := range tps {
		r := &utils.TPRateRALs{
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

func (tps RateMdls) AsTPRates() (result []*utils.TPRateRALs, err error) {
	if atps, err := tps.AsMapRates(); err != nil {
		return nil, err
	} else {
		for _, tp := range atps {
			result = append(result, tp)
		}
		return result, nil
	}
}

func MapTPRates(s []*utils.TPRateRALs) (map[string]*utils.TPRateRALs, error) {
	result := make(map[string]*utils.TPRateRALs)
	for _, e := range s {
		if _, found := result[e.ID]; !found {
			result[e.ID] = e
		} else {
			return nil, fmt.Errorf("Non unique ID %+v", e.ID)
		}
	}
	return result, nil
}

func APItoModelRate(r *utils.TPRateRALs) (result RateMdls) {
	if r != nil {
		for _, rs := range r.RateSlots {
			result = append(result, RateMdl{
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
			result = append(result, RateMdl{
				Tpid: r.TPid,
				Tag:  r.ID,
			})
		}
	}
	return
}

func APItoModelRates(rs []*utils.TPRateRALs) (result RateMdls) {
	for _, r := range rs {
		for _, sr := range APItoModelRate(r) {
			result = append(result, sr)
		}
	}
	return result
}

type DestinationRateMdls []DestinationRateMdl

func (tps DestinationRateMdls) AsMapDestinationRates() (result map[string]*utils.TPDestinationRate) {
	result = make(map[string]*utils.TPDestinationRate)
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
	return
}

func (tps DestinationRateMdls) AsTPDestinationRates() (result []*utils.TPDestinationRate) {
	for _, tp := range tps.AsMapDestinationRates() {
		result = append(result, tp)
	}
	return
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

func APItoModelDestinationRate(d *utils.TPDestinationRate) (result DestinationRateMdls) {
	if d != nil {
		for _, dr := range d.DestinationRates {
			result = append(result, DestinationRateMdl{
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
			result = append(result, DestinationRateMdl{
				Tpid: d.TPid,
				Tag:  d.ID,
			})
		}
	}
	return
}

func APItoModelDestinationRates(drs []*utils.TPDestinationRate) (result DestinationRateMdls) {
	if drs != nil {
		for _, dr := range drs {
			for _, sdr := range APItoModelDestinationRate(dr) {
				result = append(result, sdr)
			}
		}
	}
	return result
}

type RatingPlanMdls []RatingPlanMdl

func (tps RatingPlanMdls) AsMapTPRatingPlans() (result map[string]*utils.TPRatingPlan) {
	result = make(map[string]*utils.TPRatingPlan)
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
	return
}

func (tps RatingPlanMdls) AsTPRatingPlans() (result []*utils.TPRatingPlan) {
	for _, tp := range tps.AsMapTPRatingPlans() {
		result = append(result, tp)
	}
	return
}

func GetRateInterval(rpl *utils.TPRatingPlanBinding, dr *utils.DestinationRate) (i *RateInterval) {
	i = &RateInterval{
		Timing: &RITiming{
			ID:        rpl.Timing().ID,
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
		i.Rating.Rates = append(i.Rating.Rates, &RGRate{
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

func APItoModelRatingPlan(rp *utils.TPRatingPlan) (result RatingPlanMdls) {
	if rp != nil {
		for _, rpb := range rp.RatingPlanBindings {
			result = append(result, RatingPlanMdl{
				Tpid:         rp.TPid,
				Tag:          rp.ID,
				DestratesTag: rpb.DestinationRatesId,
				TimingTag:    rpb.TimingId,
				Weight:       rpb.Weight,
			})
		}
		if len(rp.RatingPlanBindings) == 0 {
			result = append(result, RatingPlanMdl{
				Tpid: rp.TPid,
				Tag:  rp.ID,
			})
		}
	}
	return
}

func APItoModelRatingPlans(rps []*utils.TPRatingPlan) (result RatingPlanMdls) {
	for _, rp := range rps {
		for _, srp := range APItoModelRatingPlan(rp) {
			result = append(result, srp)
		}
	}
	return result
}

type RatingProfileMdls []RatingProfileMdl

func (tps RatingProfileMdls) AsMapTPRatingProfiles() (result map[string]*utils.TPRatingProfile) {
	result = make(map[string]*utils.TPRatingProfile)
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
		if existing, exists := result[rp.GetId()]; !exists {
			rp.RatingPlanActivations = []*utils.TPRatingActivation{ra}
			result[rp.GetId()] = rp
		} else {
			existing.RatingPlanActivations = append(existing.RatingPlanActivations, ra)
		}
	}
	return
}

func (tps RatingProfileMdls) AsTPRatingProfiles() (result []*utils.TPRatingProfile) {
	for _, tp := range tps.AsMapTPRatingProfiles() {
		result = append(result, tp)
	}
	return
}

func MapTPRatingProfiles(s []*utils.TPRatingProfile) (map[string]*utils.TPRatingProfile, error) {
	result := make(map[string]*utils.TPRatingProfile)
	for _, e := range s {
		if _, found := result[e.GetId()]; !found {
			result[e.GetId()] = e
		} else {
			return nil, fmt.Errorf("Non unique id %+v", e.GetId())
		}
	}
	return result, nil
}

func APItoModelRatingProfile(rp *utils.TPRatingProfile) (result RatingProfileMdls) {
	if rp != nil {
		for _, rpa := range rp.RatingPlanActivations {
			result = append(result, RatingProfileMdl{
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
			result = append(result, RatingProfileMdl{
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

func APItoModelRatingProfiles(rps []*utils.TPRatingProfile) (result RatingProfileMdls) {
	for _, rp := range rps {
		for _, srp := range APItoModelRatingProfile(rp) {
			result = append(result, srp)
		}
	}
	return result
}

type SharedGroupMdls []SharedGroupMdl

func (tps SharedGroupMdls) AsMapTPSharedGroups() (result map[string]*utils.TPSharedGroups) {
	result = make(map[string]*utils.TPSharedGroups)
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
	return
}

func (tps SharedGroupMdls) AsTPSharedGroups() (result []*utils.TPSharedGroups) {
	for _, tp := range tps.AsMapTPSharedGroups() {
		result = append(result, tp)
	}
	return
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

func APItoModelSharedGroup(sgs *utils.TPSharedGroups) (result SharedGroupMdls) {
	if sgs != nil {
		for _, sg := range sgs.SharedGroups {
			result = append(result, SharedGroupMdl{
				Tpid:          sgs.TPid,
				Tag:           sgs.ID,
				Account:       sg.Account,
				Strategy:      sg.Strategy,
				RatingSubject: sg.RatingSubject,
			})
		}
		if len(sgs.SharedGroups) == 0 {
			result = append(result, SharedGroupMdl{
				Tpid: sgs.TPid,
				Tag:  sgs.ID,
			})
		}
	}
	return
}

func APItoModelSharedGroups(sgs []*utils.TPSharedGroups) (result SharedGroupMdls) {
	for _, sg := range sgs {
		for _, ssg := range APItoModelSharedGroup(sg) {
			result = append(result, ssg)
		}
	}
	return result
}

type ActionMdls []ActionMdl

func (tps ActionMdls) AsMapTPActions() (result map[string]*utils.TPActions) {
	result = make(map[string]*utils.TPActions)
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
	return
}

func (tps ActionMdls) AsTPActions() (result []*utils.TPActions) {
	for _, tp := range tps.AsMapTPActions() {
		result = append(result, tp)
	}
	return result
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

func APItoModelAction(as *utils.TPActions) (result ActionMdls) {
	if as != nil {
		for _, a := range as.Actions {
			result = append(result, ActionMdl{
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
			result = append(result, ActionMdl{
				Tpid: as.TPid,
				Tag:  as.ID,
			})
		}
	}
	return
}

func APItoModelActions(as []*utils.TPActions) (result ActionMdls) {
	for _, a := range as {
		for _, sa := range APItoModelAction(a) {
			result = append(result, sa)
		}
	}
	return result
}

type ActionPlanMdls []ActionPlanMdl

func (tps ActionPlanMdls) AsMapTPActionPlans() (result map[string]*utils.TPActionPlan) {
	result = make(map[string]*utils.TPActionPlan)
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
	return
}

func (tps ActionPlanMdls) AsTPActionPlans() (result []*utils.TPActionPlan) {
	for _, tp := range tps.AsMapTPActionPlans() {
		result = append(result, tp)
	}
	return
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

func APItoModelActionPlan(a *utils.TPActionPlan) (result ActionPlanMdls) {
	if a != nil {
		for _, ap := range a.ActionPlan {
			result = append(result, ActionPlanMdl{
				Tpid:       a.TPid,
				Tag:        a.ID,
				ActionsTag: ap.ActionsId,
				TimingTag:  ap.TimingId,
				Weight:     ap.Weight,
			})
		}
		if len(a.ActionPlan) == 0 {
			result = append(result, ActionPlanMdl{
				Tpid: a.TPid,
				Tag:  a.ID,
			})
		}
	}
	return
}

func APItoModelActionPlans(aps []*utils.TPActionPlan) (result ActionPlanMdls) {
	for _, ap := range aps {
		for _, sap := range APItoModelActionPlan(ap) {
			result = append(result, sap)
		}
	}
	return result
}

type ActionTriggerMdls []ActionTriggerMdl

func (tps ActionTriggerMdls) AsMapTPActionTriggers() (result map[string]*utils.TPActionTriggers) {
	result = make(map[string]*utils.TPActionTriggers)
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
	return
}

func (tps ActionTriggerMdls) AsTPActionTriggers() (result []*utils.TPActionTriggers) {
	for _, tp := range tps.AsMapTPActionTriggers() {
		result = append(result, tp)
	}
	return
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

func APItoModelActionTrigger(ats *utils.TPActionTriggers) (result ActionTriggerMdls) {
	if ats != nil {
		for _, at := range ats.ActionTriggers {
			result = append(result, ActionTriggerMdl{
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
			result = append(result, ActionTriggerMdl{
				Tpid: ats.TPid,
				Tag:  ats.ID,
			})
		}
	}
	return
}

func APItoModelActionTriggers(ts []*utils.TPActionTriggers) (result ActionTriggerMdls) {
	for _, t := range ts {
		for _, st := range APItoModelActionTrigger(t) {
			result = append(result, st)
		}
	}
	return result
}

type AccountActionMdls []AccountActionMdl

func (tps AccountActionMdls) AsMapTPAccountActions() (map[string]*utils.TPAccountActions, error) {
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

func (tps AccountActionMdls) AsTPAccountActions() (result []*utils.TPAccountActions) {
	atps, _ := tps.AsMapTPAccountActions()
	for _, tp := range atps {
		result = append(result, tp)
	}
	return result
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

func APItoModelAccountAction(aa *utils.TPAccountActions) *AccountActionMdl {
	return &AccountActionMdl{
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

func APItoModelAccountActions(aas []*utils.TPAccountActions) (result AccountActionMdls) {
	for _, aa := range aas {
		if aa != nil {
			result = append(result, *APItoModelAccountAction(aa))
		}
	}
	return result
}

type ResourceMdls []*ResourceMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps ResourceMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.UsageTTL, utils.Limit, utils.AllocationMessage, utils.Blocker, utils.Stored,
		utils.Weight, utils.ThresholdIDs}
}

func (tps ResourceMdls) AsTPResources() (result []*utils.TPResourceProfile) {
	mrl := make(map[string]*utils.TPResourceProfile)
	filterMap := make(map[string]utils.StringSet)
	thresholdMap := make(map[string]utils.StringSet)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		rl, found := mrl[tenID]
		if !found {
			rl = &utils.TPResourceProfile{
				TPid:    tp.Tpid,
				Tenant:  tp.Tenant,
				ID:      tp.ID,
				Blocker: tp.Blocker,
				Stored:  tp.Stored,
			}
		}
		if tp.UsageTTL != utils.EmptyString {
			rl.UsageTTL = tp.UsageTTL
		}
		if tp.Weight != 0 {
			rl.Weight = tp.Weight
		}
		if tp.Limit != utils.EmptyString {
			rl.Limit = tp.Limit
		}
		if tp.AllocationMessage != utils.EmptyString {
			rl.AllocationMessage = tp.AllocationMessage
		}
		rl.Blocker = tp.Blocker
		rl.Stored = tp.Stored
		if len(tp.ActivationInterval) != 0 {
			rl.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				rl.ActivationInterval.ActivationTime = aiSplt[0]
				rl.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				rl.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdMap[tenID]; !has {
				thresholdMap[tenID] = make(utils.StringSet)
			}
			thresholdMap[tenID].AddSlice(strings.Split(tp.ThresholdIDs, utils.InfieldSep))
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		mrl[tenID] = rl
	}
	result = make([]*utils.TPResourceProfile, len(mrl))
	i := 0
	for tntID, rl := range mrl {
		result[i] = rl
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].ThresholdIDs = thresholdMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoModelResource(rl *utils.TPResourceProfile) (mdls ResourceMdls) {
	if rl == nil {
		return
	}
	// In case that TPResourceProfile don't have filter
	if len(rl.FilterIDs) == 0 {
		mdl := &ResourceMdl{
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
			if rl.ActivationInterval.ActivationTime != utils.EmptyString {
				mdl.ActivationInterval = rl.ActivationInterval.ActivationTime
			}
			if rl.ActivationInterval.ExpiryTime != utils.EmptyString {
				mdl.ActivationInterval += utils.InfieldSep + rl.ActivationInterval.ExpiryTime
			}
		}
		for i, val := range rl.ThresholdIDs {
			if i != 0 {
				mdl.ThresholdIDs += utils.InfieldSep
			}
			mdl.ThresholdIDs += val
		}
		mdls = append(mdls, mdl)
	}
	for i, fltr := range rl.FilterIDs {
		mdl := &ResourceMdl{
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
				if rl.ActivationInterval.ActivationTime != utils.EmptyString {
					mdl.ActivationInterval = rl.ActivationInterval.ActivationTime
				}
				if rl.ActivationInterval.ExpiryTime != utils.EmptyString {
					mdl.ActivationInterval += utils.InfieldSep + rl.ActivationInterval.ExpiryTime
				}
			}
			for i, val := range rl.ThresholdIDs {
				if i != 0 {
					mdl.ThresholdIDs += utils.InfieldSep
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
	if tpRL.UsageTTL != utils.EmptyString {
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
	if tpRL.Limit != utils.EmptyString {
		if rp.Limit, err = strconv.ParseFloat(tpRL.Limit, 64); err != nil {
			return nil, err
		}
	}
	return rp, nil
}

func ResourceProfileToAPI(rp *ResourceProfile) (tpRL *utils.TPResourceProfile) {
	tpRL = &utils.TPResourceProfile{
		Tenant:             rp.Tenant,
		ID:                 rp.ID,
		FilterIDs:          make([]string, len(rp.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		Limit:              strconv.FormatFloat(rp.Limit, 'f', -1, 64),
		AllocationMessage:  rp.AllocationMessage,
		Blocker:            rp.Blocker,
		Stored:             rp.Stored,
		Weight:             rp.Weight,
		ThresholdIDs:       make([]string, len(rp.ThresholdIDs)),
	}
	if rp.UsageTTL != time.Duration(0) {
		tpRL.UsageTTL = rp.UsageTTL.String()
	}
	for i, fli := range rp.FilterIDs {
		tpRL.FilterIDs[i] = fli
	}
	for i, fli := range rp.ThresholdIDs {
		tpRL.ThresholdIDs[i] = fli
	}

	if rp.ActivationInterval != nil {
		if !rp.ActivationInterval.ActivationTime.IsZero() {
			tpRL.ActivationInterval.ActivationTime = rp.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !rp.ActivationInterval.ExpiryTime.IsZero() {
			tpRL.ActivationInterval.ExpiryTime = rp.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

type StatMdls []*StatMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps StatMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.QueueLength, utils.TTL, utils.MinItems, utils.MetricIDs, utils.MetricFilterIDs,
		utils.Stored, utils.Blocker, utils.Weight, utils.ThresholdIDs}
}

func (models StatMdls) AsTPStats() (result []*utils.TPStatProfile) {
	filterMap := make(map[string]utils.StringSet)
	thresholdMap := make(map[string]utils.StringSet)
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
		if model.TTL != utils.EmptyString {
			st.TTL = model.TTL
		}
		if model.QueueLength != 0 {
			st.QueueLength = model.QueueLength
		}
		if model.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdMap[key.TenantID()]; !has {
				thresholdMap[key.TenantID()] = make(utils.StringSet)
			}
			thresholdMap[key.TenantID()].AddSlice(strings.Split(model.ThresholdIDs, utils.InfieldSep))
		}
		if len(model.ActivationInterval) != 0 {
			st.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(model.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				st.ActivationInterval.ActivationTime = aiSplt[0]
				st.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				st.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if model.FilterIDs != utils.EmptyString {
			if _, has := filterMap[key.TenantID()]; !has {
				filterMap[key.TenantID()] = make(utils.StringSet)
			}
			filterMap[key.TenantID()].AddSlice(strings.Split(model.FilterIDs, utils.InfieldSep))
		}
		if model.MetricIDs != utils.EmptyString {
			if _, has := statMetricsMap[key.TenantID()]; !has {
				statMetricsMap[key.TenantID()] = make(map[string]*utils.MetricWithFilters)
			}
			metricIDsSplit := strings.Split(model.MetricIDs, utils.InfieldSep)
			for _, metricID := range metricIDsSplit {
				stsMetric, found := statMetricsMap[key.TenantID()][metricID]
				if !found {
					stsMetric = &utils.MetricWithFilters{
						MetricID: metricID,
					}
				}
				if model.MetricFilterIDs != utils.EmptyString {
					filterIDs := strings.Split(model.MetricFilterIDs, utils.InfieldSep)
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
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].ThresholdIDs = thresholdMap[tntID].AsSlice()
		for _, metric := range statMetricsMap[tntID] {
			result[i].Metrics = append(result[i].Metrics, metric)
		}
		i++
	}
	return
}

func APItoModelStats(st *utils.TPStatProfile) (mdls StatMdls) {
	if st != nil && len(st.Metrics) != 0 {
		for i, metric := range st.Metrics {
			mdl := &StatMdl{
				Tpid:   st.TPid,
				Tenant: st.Tenant,
				ID:     st.ID,
			}
			if i == 0 {
				for i, val := range st.FilterIDs {
					if i != 0 {
						mdl.FilterIDs += utils.InfieldSep
					}
					mdl.FilterIDs += val
				}
				if st.ActivationInterval != nil {
					if st.ActivationInterval.ActivationTime != utils.EmptyString {
						mdl.ActivationInterval = st.ActivationInterval.ActivationTime
					}
					if st.ActivationInterval.ExpiryTime != utils.EmptyString {
						mdl.ActivationInterval += utils.InfieldSep + st.ActivationInterval.ExpiryTime
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
						mdl.ThresholdIDs += utils.InfieldSep
					}
					mdl.ThresholdIDs += val
				}
			}
			for i, val := range metric.FilterIDs {
				if i != 0 {
					mdl.MetricFilterIDs += utils.InfieldSep
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
	if tpST.TTL != utils.EmptyString {
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

func StatQueueProfileToAPI(st *StatQueueProfile) (tpST *utils.TPStatProfile) {
	tpST = &utils.TPStatProfile{
		Tenant:             st.Tenant,
		ID:                 st.ID,
		FilterIDs:          make([]string, len(st.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		QueueLength:        st.QueueLength,
		Metrics:            make([]*utils.MetricWithFilters, len(st.Metrics)),
		Blocker:            st.Blocker,
		Stored:             st.Stored,
		Weight:             st.Weight,
		MinItems:           st.MinItems,
		ThresholdIDs:       make([]string, len(st.ThresholdIDs)),
	}
	for i, metric := range st.Metrics {
		tpST.Metrics[i] = &utils.MetricWithFilters{
			MetricID: metric.MetricID,
		}
		if len(metric.FilterIDs) != 0 {
			tpST.Metrics[i].FilterIDs = make([]string, len(metric.FilterIDs))
			for j, fltr := range metric.FilterIDs {
				tpST.Metrics[i].FilterIDs[j] = fltr
			}
		}

	}
	if st.TTL != time.Duration(0) {
		tpST.TTL = st.TTL.String()
	}
	for i, fli := range st.FilterIDs {
		tpST.FilterIDs[i] = fli
	}
	for i, fli := range st.ThresholdIDs {
		tpST.ThresholdIDs[i] = fli
	}

	if st.ActivationInterval != nil {
		if !st.ActivationInterval.ActivationTime.IsZero() {
			tpST.ActivationInterval.ActivationTime = st.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !st.ActivationInterval.ExpiryTime.IsZero() {
			tpST.ActivationInterval.ExpiryTime = st.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

type ThresholdMdls []*ThresholdMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps ThresholdMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.MaxHits, utils.MinHits, utils.MinSleep,
		utils.Blocker, utils.Weight, utils.ActionIDs, utils.Async}
}

func (tps ThresholdMdls) AsTPThreshold() (result []*utils.TPThresholdProfile) {
	mst := make(map[string]*utils.TPThresholdProfile)
	filterMap := make(map[string]utils.StringSet)
	actionMap := make(map[string]utils.StringSet)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
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
		if tp.ActionIDs != utils.EmptyString {
			if _, has := actionMap[tenID]; !has {
				actionMap[tenID] = make(utils.StringSet)
			}
			actionMap[tenID].AddSlice(strings.Split(tp.ActionIDs, utils.InfieldSep))
		}
		if tp.Weight != 0 {
			th.Weight = tp.Weight
		}
		if len(tp.ActivationInterval) != 0 {
			th.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}

		mst[tenID] = th
	}
	result = make([]*utils.TPThresholdProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].ActionIDs = actionMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoModelTPThreshold(th *utils.TPThresholdProfile) (mdls ThresholdMdls) {
	if th != nil {
		if len(th.ActionIDs) == 0 {
			return
		}
		min := len(th.FilterIDs)
		if min > len(th.ActionIDs) {
			min = len(th.ActionIDs)
		}
		for i := 0; i < min; i++ {
			mdl := &ThresholdMdl{
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
					if th.ActivationInterval.ActivationTime != utils.EmptyString {
						mdl.ActivationInterval = th.ActivationInterval.ActivationTime
					}
					if th.ActivationInterval.ExpiryTime != utils.EmptyString {
						mdl.ActivationInterval += utils.InfieldSep + th.ActivationInterval.ExpiryTime
					}
				}
			}
			mdl.FilterIDs = th.FilterIDs[i]
			mdl.ActionIDs = th.ActionIDs[i]
			mdls = append(mdls, mdl)
		}

		if len(th.FilterIDs)-min > 0 {
			for i := min; i < len(th.FilterIDs); i++ {
				mdl := &ThresholdMdl{
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
				mdl := &ThresholdMdl{
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
						if th.ActivationInterval.ActivationTime != utils.EmptyString {
							mdl.ActivationInterval = th.ActivationInterval.ActivationTime
						}
						if th.ActivationInterval.ExpiryTime != utils.EmptyString {
							mdl.ActivationInterval += utils.InfieldSep + th.ActivationInterval.ExpiryTime
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
	if tpTH.MinSleep != utils.EmptyString {
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

func ThresholdProfileToAPI(th *ThresholdProfile) (tpTH *utils.TPThresholdProfile) {
	tpTH = &utils.TPThresholdProfile{
		Tenant:             th.Tenant,
		ID:                 th.ID,
		FilterIDs:          make([]string, len(th.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		MaxHits:            th.MaxHits,
		MinHits:            th.MinHits,
		Blocker:            th.Blocker,
		Weight:             th.Weight,
		ActionIDs:          make([]string, len(th.ActionIDs)),
		Async:              th.Async,
	}
	if th.MinSleep != time.Duration(0) {
		tpTH.MinSleep = th.MinSleep.String()
	}
	for i, fli := range th.FilterIDs {
		tpTH.FilterIDs[i] = fli
	}
	for i, fli := range th.ActionIDs {
		tpTH.ActionIDs[i] = fli
	}

	if th.ActivationInterval != nil {
		if !th.ActivationInterval.ActivationTime.IsZero() {
			tpTH.ActivationInterval.ActivationTime = th.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !th.ActivationInterval.ExpiryTime.IsZero() {
			tpTH.ActivationInterval.ExpiryTime = th.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

type FilterMdls []*FilterMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps FilterMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.Type, utils.Element,
		utils.Values, utils.ActivationIntervalString}
}

func (tps FilterMdls) AsTPFilter() (result []*utils.TPFilterProfile) {
	mst := make(map[string]*utils.TPFilterProfile)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
		if !found {
			th = &utils.TPFilterProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if len(tp.ActivationInterval) != 0 {
			th.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.Type != utils.EmptyString {
			var vals []string
			if tp.Values != utils.EmptyString {
				vals = splitDynFltrValues(tp.Values, utils.InfieldSep)
			}
			th.Filters = append(th.Filters, &utils.TPFilter{
				Type:    tp.Type,
				Element: tp.Element,
				Values:  vals,
			})
		}
		mst[tenID] = th
	}
	result = make([]*utils.TPFilterProfile, len(mst))
	i := 0
	for _, th := range mst {
		result[i] = th
		i++
	}
	return
}

func APItoModelTPFilter(th *utils.TPFilterProfile) (mdls FilterMdls) {
	if th == nil || len(th.Filters) == 0 {
		return
	}
	for _, fltr := range th.Filters {
		mdl := &FilterMdl{
			Tpid:   th.TPid,
			Tenant: th.Tenant,
			ID:     th.ID,
		}
		mdl.Type = fltr.Type
		mdl.Element = fltr.Element
		if th.ActivationInterval != nil {
			if th.ActivationInterval.ActivationTime != utils.EmptyString {
				mdl.ActivationInterval = th.ActivationInterval.ActivationTime
			}
			if th.ActivationInterval.ExpiryTime != utils.EmptyString {
				mdl.ActivationInterval += utils.InfieldSep + th.ActivationInterval.ExpiryTime
			}
		}
		for i, val := range fltr.Values {
			if i != 0 {
				mdl.Values += utils.InfieldSep
			}
			mdl.Values += val
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
		rf := &FilterRule{Type: f.Type, Element: f.Element, Values: f.Values}
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
		Tenant:             f.Tenant,
		ID:                 f.ID,
		Filters:            make([]*utils.TPFilter, len(f.Rules)),
		ActivationInterval: new(utils.TPActivationInterval),
	}
	for i, reqFltr := range f.Rules {
		tpFltr.Filters[i] = &utils.TPFilter{
			Type:    reqFltr.Type,
			Element: reqFltr.Element,
			Values:  make([]string, len(reqFltr.Values)),
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

type RouteMdls []*RouteMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps RouteMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.Sorting, utils.SortingParameters, utils.RouteID, utils.RouteFilterIDs,
		utils.RouteAccountIDs, utils.RouteRatingplanIDs, utils.RouteRateProfileIDs, utils.RouteResourceIDs,
		utils.RouteStatIDs, utils.RouteWeight, utils.RouteBlocker,
		utils.RouteParameters, utils.Weight,
	}
}

func (tps RouteMdls) AsTPRouteProfile() (result []*utils.TPRouteProfile) {
	filterMap := make(map[string]utils.StringSet)
	mst := make(map[string]*utils.TPRouteProfile)
	routeMap := make(map[string]map[string]*utils.TPRoute)
	sortingParameterMap := make(map[string]utils.StringSet)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
		if !found {
			th = &utils.TPRouteProfile{
				TPid:              tp.Tpid,
				Tenant:            tp.Tenant,
				ID:                tp.ID,
				SortingParameters: []string{},
			}
		}
		if tp.RouteID != utils.EmptyString {
			if _, has := routeMap[tenID]; !has {
				routeMap[tenID] = make(map[string]*utils.TPRoute)
			}
			routeID := tp.RouteID
			if tp.RouteFilterIDs != utils.EmptyString {
				routeID = utils.ConcatenatedKey(routeID,
					utils.NewStringSet(strings.Split(tp.RouteFilterIDs, utils.InfieldSep)).Sha1())
			}
			sup, found := routeMap[tenID][routeID]
			if !found {
				sup = &utils.TPRoute{
					ID:              tp.RouteID,
					Weight:          tp.RouteWeight,
					Blocker:         tp.RouteBlocker,
					RouteParameters: tp.RouteParameters,
				}
			}
			if tp.RouteFilterIDs != utils.EmptyString {
				supFilterSplit := strings.Split(tp.RouteFilterIDs, utils.InfieldSep)
				sup.FilterIDs = append(sup.FilterIDs, supFilterSplit...)
			}
			if tp.RouteRatingplanIDs != utils.EmptyString {
				ratingPlanSplit := strings.Split(tp.RouteRatingplanIDs, utils.InfieldSep)
				sup.RatingPlanIDs = append(sup.RatingPlanIDs, ratingPlanSplit...)
			}
			if tp.RouteResourceIDs != utils.EmptyString {
				resSplit := strings.Split(tp.RouteResourceIDs, utils.InfieldSep)
				sup.ResourceIDs = append(sup.ResourceIDs, resSplit...)
			}
			if tp.RouteStatIDs != utils.EmptyString {
				statSplit := strings.Split(tp.RouteStatIDs, utils.InfieldSep)
				sup.StatIDs = append(sup.StatIDs, statSplit...)
			}
			if tp.RouteAccountIDs != utils.EmptyString {
				accSplit := strings.Split(tp.RouteAccountIDs, utils.InfieldSep)
				sup.AccountIDs = append(sup.AccountIDs, accSplit...)
			}
			routeMap[tenID][routeID] = sup
		}
		if tp.Sorting != utils.EmptyString {
			th.Sorting = tp.Sorting
		}
		if tp.SortingParameters != utils.EmptyString {
			if _, has := sortingParameterMap[tenID]; !has {
				sortingParameterMap[tenID] = make(utils.StringSet)
			}
			sortingParameterMap[tenID].AddSlice(strings.Split(tp.SortingParameters, utils.InfieldSep))
		}
		if tp.Weight != 0 {
			th.Weight = tp.Weight
		}
		if tp.ActivationInterval != utils.EmptyString {
			th.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		mst[tenID] = th
	}
	result = make([]*utils.TPRouteProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		for _, supData := range routeMap[tntID] {
			result[i].Routes = append(result[i].Routes, supData)
		}
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].SortingParameters = sortingParameterMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoModelTPRoutes(st *utils.TPRouteProfile) (mdls RouteMdls) {
	if len(st.Routes) == 0 {
		return
	}
	for i, supl := range st.Routes {
		mdl := &RouteMdl{
			Tenant: st.Tenant,
			Tpid:   st.TPid,
			ID:     st.ID,
		}
		if i == 0 {
			mdl.Sorting = st.Sorting
			mdl.Weight = st.Weight
			for i, val := range st.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.InfieldSep
				}
				mdl.FilterIDs += val
			}
			for i, val := range st.SortingParameters {
				if i != 0 {
					mdl.SortingParameters += utils.InfieldSep
				}
				mdl.SortingParameters += val
			}
			if st.ActivationInterval != nil {
				if st.ActivationInterval.ActivationTime != utils.EmptyString {
					mdl.ActivationInterval = st.ActivationInterval.ActivationTime
				}
				if st.ActivationInterval.ExpiryTime != utils.EmptyString {
					mdl.ActivationInterval += utils.InfieldSep + st.ActivationInterval.ExpiryTime
				}
			}
		}
		mdl.RouteID = supl.ID
		for i, val := range supl.AccountIDs {
			if i != 0 {
				mdl.RouteAccountIDs += utils.InfieldSep
			}
			mdl.RouteAccountIDs += val
		}
		for i, val := range supl.RatingPlanIDs {
			if i != 0 {
				mdl.RouteRatingplanIDs += utils.InfieldSep
			}
			mdl.RouteRatingplanIDs += val
		}
		for i, val := range supl.FilterIDs {
			if i != 0 {
				mdl.RouteFilterIDs += utils.InfieldSep
			}
			mdl.RouteFilterIDs += val
		}
		for i, val := range supl.ResourceIDs {
			if i != 0 {
				mdl.RouteResourceIDs += utils.InfieldSep
			}
			mdl.RouteResourceIDs += val
		}
		for i, val := range supl.StatIDs {
			if i != 0 {
				mdl.RouteStatIDs += utils.InfieldSep
			}
			mdl.RouteStatIDs += val
		}
		mdl.RouteWeight = supl.Weight
		mdl.RouteParameters = supl.RouteParameters
		mdl.RouteBlocker = supl.Blocker
		mdls = append(mdls, mdl)
	}
	return
}

func APItoRouteProfile(tpRp *utils.TPRouteProfile, timezone string) (rp *RouteProfile, err error) {
	rp = &RouteProfile{
		Tenant:            tpRp.Tenant,
		ID:                tpRp.ID,
		Sorting:           tpRp.Sorting,
		Weight:            tpRp.Weight,
		Routes:            make([]*Route, len(tpRp.Routes)),
		SortingParameters: make([]string, len(tpRp.SortingParameters)),
		FilterIDs:         make([]string, len(tpRp.FilterIDs)),
	}
	for i, stp := range tpRp.SortingParameters {
		rp.SortingParameters[i] = stp
	}
	for i, fli := range tpRp.FilterIDs {
		rp.FilterIDs[i] = fli
	}
	if tpRp.ActivationInterval != nil {
		if rp.ActivationInterval, err = tpRp.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	for i, route := range tpRp.Routes {
		rp.Routes[i] = &Route{
			ID:              route.ID,
			Weight:          route.Weight,
			Blocker:         route.Blocker,
			RatingPlanIDs:   route.RatingPlanIDs,
			AccountIDs:      route.AccountIDs,
			FilterIDs:       route.FilterIDs,
			ResourceIDs:     route.ResourceIDs,
			StatIDs:         route.StatIDs,
			RouteParameters: route.RouteParameters,
		}
	}
	return rp, nil
}

func RouteProfileToAPI(rp *RouteProfile) (tpRp *utils.TPRouteProfile) {
	tpRp = &utils.TPRouteProfile{
		Tenant:             rp.Tenant,
		ID:                 rp.ID,
		FilterIDs:          make([]string, len(rp.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		Sorting:            rp.Sorting,
		SortingParameters:  make([]string, len(rp.SortingParameters)),
		Routes:             make([]*utils.TPRoute, len(rp.Routes)),
		Weight:             rp.Weight,
	}

	for i, route := range rp.Routes {
		tpRp.Routes[i] = &utils.TPRoute{
			ID:              route.ID,
			FilterIDs:       route.FilterIDs,
			AccountIDs:      route.AccountIDs,
			RatingPlanIDs:   route.RatingPlanIDs,
			ResourceIDs:     route.ResourceIDs,
			StatIDs:         route.StatIDs,
			Weight:          route.Weight,
			Blocker:         route.Blocker,
			RouteParameters: route.RouteParameters,
		}
	}
	for i, fli := range rp.FilterIDs {
		tpRp.FilterIDs[i] = fli
	}
	for i, fli := range rp.SortingParameters {
		tpRp.SortingParameters[i] = fli
	}
	if rp.ActivationInterval != nil {
		if !rp.ActivationInterval.ActivationTime.IsZero() {
			tpRp.ActivationInterval.ActivationTime = rp.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !rp.ActivationInterval.ExpiryTime.IsZero() {
			tpRp.ActivationInterval.ExpiryTime = rp.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

type AttributeMdls []*AttributeMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps AttributeMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.AttributeFilterIDs, utils.Path, utils.Type, utils.Value, utils.Blocker, utils.Weight}
}

func (tps AttributeMdls) AsTPAttributes() (result []*utils.TPAttributeProfile) {
	mst := make(map[string]*utils.TPAttributeProfile)
	filterMap := make(map[string]utils.StringSet)
	contextMap := make(map[string]utils.StringSet)
	for _, tp := range tps {
		key := &utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		th, found := mst[tenID]
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
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
				th.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				th.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[key.TenantID()]; !has {
				filterMap[key.TenantID()] = make(utils.StringSet)
			}
			filterMap[key.TenantID()].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.Contexts != utils.EmptyString {
			if _, has := contextMap[tenID]; !has {
				contextMap[key.TenantID()] = make(utils.StringSet)
			}
			contextMap[key.TenantID()].AddSlice(strings.Split(tp.Contexts, utils.InfieldSep))
		}
		if tp.Path != utils.EmptyString {
			filterIDs := make([]string, 0)
			if tp.AttributeFilterIDs != utils.EmptyString {
				filterAttrSplit := strings.Split(tp.AttributeFilterIDs, utils.InfieldSep)
				for _, filterAttr := range filterAttrSplit {
					filterIDs = append(filterIDs, filterAttr)
				}
			}
			th.Attributes = append(th.Attributes, &utils.TPAttribute{
				FilterIDs: filterIDs,
				Type:      tp.Type,
				Path:      tp.Path,
				Value:     tp.Value,
			})
		}
		mst[key.TenantID()] = th
	}
	result = make([]*utils.TPAttributeProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].Contexts = contextMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoModelTPAttribute(th *utils.TPAttributeProfile) (mdls AttributeMdls) {
	if len(th.Attributes) == 0 {
		return
	}
	for i, reqAttribute := range th.Attributes {
		mdl := &AttributeMdl{
			Tpid:   th.TPid,
			Tenant: th.Tenant,
			ID:     th.ID,
		}
		if i == 0 {
			mdl.Blocker = th.Blocker
			if th.ActivationInterval != nil {
				if th.ActivationInterval.ActivationTime != utils.EmptyString {
					mdl.ActivationInterval = th.ActivationInterval.ActivationTime
				}
				if th.ActivationInterval.ExpiryTime != utils.EmptyString {
					mdl.ActivationInterval += utils.InfieldSep + th.ActivationInterval.ExpiryTime
				}
			}
			for i, val := range th.Contexts {
				if i != 0 {
					mdl.Contexts += utils.InfieldSep
				}
				mdl.Contexts += val
			}
			for i, val := range th.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.InfieldSep
				}
				mdl.FilterIDs += val
			}
			if th.Weight != 0 {
				mdl.Weight = th.Weight
			}
		}
		mdl.Path = reqAttribute.Path
		mdl.Value = reqAttribute.Value
		mdl.Type = reqAttribute.Type
		mdl.AttributeFilterIDs = strings.Join(reqAttribute.FilterIDs, utils.InfieldSep)
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
		if reqAttr.Path == utils.EmptyString { // we do not suppot empty Path in Attributes
			err = fmt.Errorf("empty path in AttributeProfile <%s>", attrPrf.TenantID())
			return
		}
		sbstPrsr, err := config.NewRSRParsers(reqAttr.Value, config.CgrConfig().GeneralCfg().RSRSep)
		if err != nil {
			return nil, err
		}
		attrPrf.Attributes[i] = &Attribute{
			FilterIDs: reqAttr.FilterIDs,
			Path:      reqAttr.Path,
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

func AttributeProfileToAPI(attrPrf *AttributeProfile) (tpAttr *utils.TPAttributeProfile) {
	tpAttr = &utils.TPAttributeProfile{
		Tenant:             attrPrf.Tenant,
		ID:                 attrPrf.ID,
		FilterIDs:          make([]string, len(attrPrf.FilterIDs)),
		Contexts:           make([]string, len(attrPrf.Contexts)),
		Attributes:         make([]*utils.TPAttribute, len(attrPrf.Attributes)),
		ActivationInterval: new(utils.TPActivationInterval),
		Blocker:            attrPrf.Blocker,
		Weight:             attrPrf.Weight,
	}
	for i, fli := range attrPrf.FilterIDs {
		tpAttr.FilterIDs[i] = fli
	}
	for i, fli := range attrPrf.Contexts {
		tpAttr.Contexts[i] = fli
	}
	for i, attr := range attrPrf.Attributes {
		tpAttr.Attributes[i] = &utils.TPAttribute{
			FilterIDs: attr.FilterIDs,
			Path:      attr.Path,
			Type:      attr.Type,
			Value:     attr.Value.GetRule(utils.InfieldSep),
		}
	}
	if attrPrf.ActivationInterval != nil {
		if !attrPrf.ActivationInterval.ActivationTime.IsZero() {
			tpAttr.ActivationInterval.ActivationTime = attrPrf.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !attrPrf.ActivationInterval.ExpiryTime.IsZero() {
			tpAttr.ActivationInterval.ExpiryTime = attrPrf.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

type ChargerMdls []*ChargerMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps ChargerMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs, utils.ActivationIntervalString,
		utils.RunID, utils.AttributeIDs, utils.Weight}
}

func (tps ChargerMdls) AsTPChargers() (result []*utils.TPChargerProfile) {
	mst := make(map[string]*utils.TPChargerProfile)
	filterMap := make(map[string]utils.StringSet)
	attributeMap := make(map[string][]string)
	for _, tp := range tps {
		tntID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		tpCPP, found := mst[tntID]
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
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				tpCPP.ActivationInterval.ActivationTime = aiSplt[0]
				tpCPP.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				tpCPP.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tntID]; !has {
				filterMap[tntID] = make(utils.StringSet)
			}
			filterMap[tntID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.RunID != utils.EmptyString {
			tpCPP.RunID = tp.RunID
		}
		if tp.AttributeIDs != utils.EmptyString {
			attributeSplit := strings.Split(tp.AttributeIDs, utils.InfieldSep)
			var inlineAttribute string
			var dynam bool
			for _, attribute := range attributeSplit {
				if !dynam && !strings.HasPrefix(attribute, utils.Meta) {
					if inlineAttribute != utils.EmptyString {
						attributeMap[tntID] = append(attributeMap[tntID], inlineAttribute[1:])
						inlineAttribute = utils.EmptyString
					}
					attributeMap[tntID] = append(attributeMap[tntID], attribute)
					continue
				}
				if dynam {
					dynam = !strings.Contains(attribute, string(utils.RSRDynEndChar))
				} else {
					dynam = strings.Contains(attribute, string(utils.RSRDynStartChar))
				}
				inlineAttribute += utils.InfieldSep + attribute
			}
			if inlineAttribute != utils.EmptyString {
				attributeMap[tntID] = append(attributeMap[tntID], inlineAttribute[1:])
				inlineAttribute = utils.EmptyString
			}
		}
		mst[tntID] = tpCPP
	}
	result = make([]*utils.TPChargerProfile, len(mst))
	i := 0
	for tntID, tp := range mst {
		result[i] = tp
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].AttributeIDs = make([]string, 0, len(attributeMap[tntID]))
		for _, attributeID := range attributeMap[tntID] {
			result[i].AttributeIDs = append(result[i].AttributeIDs, attributeID)
		}
		i++
	}
	return
}

func APItoModelTPCharger(tpCPP *utils.TPChargerProfile) (mdls ChargerMdls) {
	if tpCPP != nil {
		min := len(tpCPP.FilterIDs)
		isFilter := true
		if min > len(tpCPP.AttributeIDs) {
			min = len(tpCPP.AttributeIDs)
			isFilter = false
		}
		if min == 0 {
			mdl := &ChargerMdl{
				Tenant: tpCPP.Tenant,
				Tpid:   tpCPP.TPid,
				ID:     tpCPP.ID,
				Weight: tpCPP.Weight,
				RunID:  tpCPP.RunID,
			}
			if tpCPP.ActivationInterval != nil {
				if tpCPP.ActivationInterval.ActivationTime != utils.EmptyString {
					mdl.ActivationInterval = tpCPP.ActivationInterval.ActivationTime
				}
				if tpCPP.ActivationInterval.ExpiryTime != utils.EmptyString {
					mdl.ActivationInterval += utils.InfieldSep + tpCPP.ActivationInterval.ExpiryTime
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
				mdl := &ChargerMdl{
					Tenant: tpCPP.Tenant,
					Tpid:   tpCPP.TPid,
					ID:     tpCPP.ID,
				}
				if i == 0 {
					mdl.Weight = tpCPP.Weight
					mdl.RunID = tpCPP.RunID
					if tpCPP.ActivationInterval != nil {
						if tpCPP.ActivationInterval.ActivationTime != utils.EmptyString {
							mdl.ActivationInterval = tpCPP.ActivationInterval.ActivationTime
						}
						if tpCPP.ActivationInterval.ExpiryTime != utils.EmptyString {
							mdl.ActivationInterval += utils.InfieldSep + tpCPP.ActivationInterval.ExpiryTime
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
				mdl := &ChargerMdl{
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
				mdl := &ChargerMdl{
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

func ChargerProfileToAPI(chargerPrf *ChargerProfile) (tpCharger *utils.TPChargerProfile) {
	tpCharger = &utils.TPChargerProfile{
		Tenant:             chargerPrf.Tenant,
		ID:                 chargerPrf.ID,
		FilterIDs:          make([]string, len(chargerPrf.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		RunID:              chargerPrf.RunID,
		AttributeIDs:       make([]string, len(chargerPrf.AttributeIDs)),
		Weight:             chargerPrf.Weight,
	}
	for i, fli := range chargerPrf.FilterIDs {
		tpCharger.FilterIDs[i] = fli
	}
	for i, fli := range chargerPrf.AttributeIDs {
		tpCharger.AttributeIDs[i] = fli
	}
	if chargerPrf.ActivationInterval != nil {
		if !chargerPrf.ActivationInterval.ActivationTime.IsZero() {
			tpCharger.ActivationInterval.ActivationTime = chargerPrf.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !chargerPrf.ActivationInterval.ExpiryTime.IsZero() {
			tpCharger.ActivationInterval.ExpiryTime = chargerPrf.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

type DispatcherProfileMdls []*DispatcherProfileMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps DispatcherProfileMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.Subsystems, utils.FilterIDs, utils.ActivationIntervalString,
		utils.Strategy, utils.StrategyParameters, utils.ConnID, utils.ConnFilterIDs,
		utils.ConnWeight, utils.ConnBlocker, utils.ConnParameters, utils.Weight}
}

func (tps DispatcherProfileMdls) AsTPDispatcherProfiles() (result []*utils.TPDispatcherProfile) {
	mst := make(map[string]*utils.TPDispatcherProfile)
	filterMap := make(map[string]utils.StringSet)
	contextMap := make(map[string]utils.StringSet)
	connsMap := make(map[string]map[string]utils.TPDispatcherHostProfile)
	connsFilterMap := make(map[string]map[string]utils.StringSet)
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
		if tp.Subsystems != utils.EmptyString {
			if _, has := contextMap[tenantID]; !has {
				contextMap[tenantID] = make(utils.StringSet)
			}
			contextMap[tenantID].AddSlice(strings.Split(tp.Subsystems, utils.InfieldSep))
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenantID]; !has {
				filterMap[tenantID] = make(utils.StringSet)
			}
			filterMap[tenantID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if len(tp.ActivationInterval) != 0 {
			tpDPP.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				tpDPP.ActivationInterval.ActivationTime = aiSplt[0]
				tpDPP.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				tpDPP.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.Strategy != utils.EmptyString {
			tpDPP.Strategy = tp.Strategy
		}
		if tp.StrategyParameters != utils.EmptyString {
			for _, param := range strings.Split(tp.StrategyParameters, utils.InfieldSep) {
				tpDPP.StrategyParams = append(tpDPP.StrategyParams, param)
			}
		}
		if tp.ConnID != utils.EmptyString {
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
			for _, param := range strings.Split(tp.ConnParameters, utils.InfieldSep) {
				conn.Params = append(conn.Params, param)
			}
			connsMap[tenantID][tp.ConnID] = conn

			if dFilter, has := connsFilterMap[tenantID]; !has {
				connsFilterMap[tenantID] = make(map[string]utils.StringSet)
				connsFilterMap[tenantID][tp.ConnID] = make(utils.StringSet)
			} else if _, has := dFilter[tp.ConnID]; !has {
				connsFilterMap[tenantID][tp.ConnID] = make(utils.StringSet)
			}
			if tp.ConnFilterIDs != utils.EmptyString {
				connsFilterMap[tenantID][tp.ConnID].AddSlice(strings.Split(tp.ConnFilterIDs, utils.InfieldSep))
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
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		result[i].Subsystems = contextMap[tntID].AsSlice()
		for conID, conn := range connsMap[tntID] {
			conn.FilterIDs = connsFilterMap[tntID][conID].AsSlice()
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
			strategy += utils.InfieldSep + sp[i].(string)
		}
	}
	return
}

func APItoModelTPDispatcherProfile(tpDPP *utils.TPDispatcherProfile) (mdls DispatcherProfileMdls) {
	if tpDPP == nil {
		return
	}

	filters := strings.Join(tpDPP.FilterIDs, utils.InfieldSep)
	subsystems := strings.Join(tpDPP.Subsystems, utils.InfieldSep)

	interval := utils.EmptyString
	if tpDPP.ActivationInterval != nil {
		if tpDPP.ActivationInterval.ActivationTime != utils.EmptyString {
			interval = tpDPP.ActivationInterval.ActivationTime
		}
		if tpDPP.ActivationInterval.ExpiryTime != utils.EmptyString {
			interval += utils.InfieldSep + tpDPP.ActivationInterval.ExpiryTime
		}
	}

	strategy := paramsToString(tpDPP.StrategyParams)

	if len(tpDPP.Hosts) == 0 {
		return append(mdls, &DispatcherProfileMdl{
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

	conFilter := strings.Join(tpDPP.Hosts[0].FilterIDs, utils.InfieldSep)
	conParam := paramsToString(tpDPP.Hosts[0].Params)

	mdls = append(mdls, &DispatcherProfileMdl{
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
		ConnFilterIDs:  conFilter,
		ConnWeight:     tpDPP.Hosts[0].Weight,
		ConnBlocker:    tpDPP.Hosts[0].Blocker,
		ConnParameters: conParam,
	})
	for i := 1; i < len(tpDPP.Hosts); i++ {
		conFilter = strings.Join(tpDPP.Hosts[i].FilterIDs, utils.InfieldSep)
		conParam = paramsToString(tpDPP.Hosts[i].Params)
		mdls = append(mdls, &DispatcherProfileMdl{
			Tpid:   tpDPP.TPid,
			Tenant: tpDPP.Tenant,
			ID:     tpDPP.ID,

			ConnID:         tpDPP.Hosts[i].ID,
			ConnFilterIDs:  conFilter,
			ConnWeight:     tpDPP.Hosts[i].Weight,
			ConnBlocker:    tpDPP.Hosts[i].Blocker,
			ConnParameters: conParam,
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
		if param != utils.EmptyString {
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
			if param == utils.EmptyString {
				continue
			}
			if p := strings.SplitN(utils.IfaceAsString(param), utils.ConcatenatedKeySep, 2); len(p) == 1 {
				dpp.Hosts[i].Params[strconv.Itoa(j)] = p[0]
			} else {
				dpp.Hosts[i].Params[p[0]] = p[1]
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

func DispatcherProfileToAPI(dpp *DispatcherProfile) (tpDPP *utils.TPDispatcherProfile) {
	tpDPP = &utils.TPDispatcherProfile{
		Tenant:             dpp.Tenant,
		ID:                 dpp.ID,
		Subsystems:         make([]string, len(dpp.Subsystems)),
		FilterIDs:          make([]string, len(dpp.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		Strategy:           dpp.Strategy,
		StrategyParams:     make([]interface{}, len(dpp.StrategyParams)),
		Weight:             dpp.Weight,
		Hosts:              make([]*utils.TPDispatcherHostProfile, len(dpp.Hosts)),
	}

	for i, fli := range dpp.FilterIDs {
		tpDPP.FilterIDs[i] = fli
	}
	for i, sub := range dpp.Subsystems {
		tpDPP.Subsystems[i] = sub
	}
	for key, val := range dpp.StrategyParams {
		// here we expect that the key to be an integer because
		// according to APItoDispatcherProfile when we convert from TP to obj we use index as key
		// so we can ignore error
		idx, _ := strconv.Atoi(key)
		tpDPP.StrategyParams[idx] = val
	}
	for i, host := range dpp.Hosts {
		tpDPP.Hosts[i] = &utils.TPDispatcherHostProfile{
			ID:        host.ID,
			FilterIDs: make([]string, len(host.FilterIDs)),
			Weight:    host.Weight,
			Params:    make([]interface{}, len(host.Params)),
			Blocker:   host.Blocker,
		}
		for j, fltr := range host.FilterIDs {
			tpDPP.Hosts[i].FilterIDs[j] = fltr
		}
		idx := 0
		for key, val := range host.Params {
			paramVal := val
			if _, err := strconv.Atoi(key); err != nil {
				paramVal = utils.ConcatenatedKey(key, utils.IfaceAsString(val))
			}
			tpDPP.Hosts[i].Params[idx] = paramVal
			idx++
		}
	}

	if dpp.ActivationInterval != nil {
		if !dpp.ActivationInterval.ActivationTime.IsZero() {
			tpDPP.ActivationInterval.ActivationTime = dpp.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !dpp.ActivationInterval.ExpiryTime.IsZero() {
			tpDPP.ActivationInterval.ExpiryTime = dpp.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

// TPHosts
type DispatcherHostMdls []*DispatcherHostMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps DispatcherHostMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.Address, utils.Transport, utils.TLS}
}

func (tps DispatcherHostMdls) AsTPDispatcherHosts() (result []*utils.TPDispatcherHost) {
	hostsMap := make(map[string]*utils.TPDispatcherHost)
	for _, tp := range tps {
		if len(tp.Address) == 0 { // empty addres do not populate conns
			continue
		}
		if len(tp.Transport) == 0 {
			tp.Transport = utils.MetaJSON
		}
		hostsMap[utils.ConcatenatedKey(tp.Tenant, tp.ID)] = &utils.TPDispatcherHost{
			TPid:   tp.Tpid,
			Tenant: tp.Tenant,
			ID:     tp.ID,
			Conn: &utils.TPDispatcherHostConn{
				Address:   tp.Address,
				Transport: tp.Transport,
				TLS:       tp.TLS,
			},
		}
		continue
	}
	for _, host := range hostsMap {
		result = append(result, host)
	}
	return
}

func APItoModelTPDispatcherHost(tpDPH *utils.TPDispatcherHost) (mdls *DispatcherHostMdl) {
	if tpDPH == nil {
		return
	}
	return &DispatcherHostMdl{
		Tpid:      tpDPH.TPid,
		Tenant:    tpDPH.Tenant,
		ID:        tpDPH.ID,
		Address:   tpDPH.Conn.Address,
		Transport: tpDPH.Conn.Transport,
		TLS:       tpDPH.Conn.TLS,
	}
}

func APItoDispatcherHost(tpDPH *utils.TPDispatcherHost) (dpp *DispatcherHost) {
	if tpDPH == nil {
		return
	}
	return &DispatcherHost{
		Tenant: tpDPH.Tenant,
		ID:     tpDPH.ID,
		Conn: &config.RemoteHost{
			Address:   tpDPH.Conn.Address,
			Transport: tpDPH.Conn.Transport,
			TLS:       tpDPH.Conn.TLS,
		},
	}
}

func DispatcherHostToAPI(dph *DispatcherHost) (tpDPH *utils.TPDispatcherHost) {
	return &utils.TPDispatcherHost{
		Tenant: dph.Tenant,
		ID:     dph.ID,
		Conn: &utils.TPDispatcherHostConn{
			Address:   dph.Conn.Address,
			Transport: dph.Conn.Transport,
			TLS:       dph.Conn.TLS,
		},
	}
}

// RateProfileMdls is used
type RateProfileMdls []*RateProfileMdl

// CSVHeader return the header for csv fields as a slice of string
func (tps RateProfileMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.ActivationIntervalString, utils.Weight, utils.ConnectFee, utils.MinCost,
		utils.MaxCost, utils.MaxCostStrategy, utils.RateID,
		utils.RateFilterIDs, utils.RateActivationStart, utils.RateWeight, utils.RateBlocker,
		utils.RateIntervalStart, utils.RateFixedFee, utils.RateRecurrentFee, utils.RateUnit, utils.RateIncrement,
	}
}

func (tps RateProfileMdls) AsTPRateProfile() (result []*utils.TPRateProfile) {
	filterMap := make(map[string]utils.StringSet)
	mst := make(map[string]*utils.TPRateProfile)
	rateMap := make(map[string]map[string]*utils.TPRate)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		rPrf, found := mst[tenID]
		if !found {
			rPrf = &utils.TPRateProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.RateID != utils.EmptyString {
			if _, has := rateMap[tenID]; !has {
				rateMap[tenID] = make(map[string]*utils.TPRate)
			}
			rate, found := rateMap[tenID][tp.RateID]
			if !found {
				rate = &utils.TPRate{
					ID:            tp.RateID,
					IntervalRates: make([]*utils.TPIntervalRate, 0),
					Blocker:       tp.RateBlocker,
				}
			}
			if tp.RateFilterIDs != utils.EmptyString {
				rateFilterSplit := strings.Split(tp.RateFilterIDs, utils.InfieldSep)
				rate.FilterIDs = append(rate.FilterIDs, rateFilterSplit...)
			}
			if tp.RateActivationTimes != utils.EmptyString {
				rate.ActivationTimes = tp.RateActivationTimes
			}
			if tp.RateWeight != 0 {
				rate.Weight = tp.RateWeight
			}
			// create new interval rate and append to the slice
			intervalRate := new(utils.TPIntervalRate)
			if tp.RateIntervalStart != utils.EmptyString {
				intervalRate.IntervalStart = tp.RateIntervalStart
			}
			if tp.RateFixedFee != 0 {
				intervalRate.FixedFee = tp.RateFixedFee
			}
			if tp.RateRecurrentFee != 0 {
				intervalRate.RecurrentFee = tp.RateRecurrentFee
			}
			if tp.RateIncrement != utils.EmptyString {
				intervalRate.Increment = tp.RateIncrement
			}
			if tp.RateUnit != utils.EmptyString {
				intervalRate.Unit = tp.RateUnit
			}
			rate.IntervalRates = append(rate.IntervalRates, intervalRate)
			rateMap[tenID][tp.RateID] = rate
		}

		if tp.Weight != 0 {
			rPrf.Weight = tp.Weight
		}
		if tp.MinCost != 0 {
			rPrf.MinCost = tp.MinCost
		}
		if tp.MaxCost != 0 {
			rPrf.MaxCost = tp.MaxCost
		}
		if tp.MaxCostStrategy != utils.EmptyString {
			rPrf.MaxCostStrategy = tp.MaxCostStrategy
		}
		if tp.ActivationInterval != utils.EmptyString {
			rPrf.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				rPrf.ActivationInterval.ActivationTime = aiSplt[0]
				rPrf.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				rPrf.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterMap[tenID]; !has {
				filterMap[tenID] = make(utils.StringSet)
			}
			filterMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		mst[tenID] = rPrf
	}
	result = make([]*utils.TPRateProfile, len(mst))
	i := 0
	for tntID, th := range mst {
		result[i] = th
		result[i].Rates = rateMap[tntID]
		result[i].FilterIDs = filterMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoModelTPRateProfile(tPrf *utils.TPRateProfile) (mdls RateProfileMdls) {
	if len(tPrf.Rates) == 0 {
		return
	}
	i := 0
	for _, rate := range tPrf.Rates {
		for j, intervalRate := range rate.IntervalRates {
			mdl := &RateProfileMdl{
				Tenant: tPrf.Tenant,
				Tpid:   tPrf.TPid,
				ID:     tPrf.ID,
			}
			if i == 0 {
				for i, val := range tPrf.FilterIDs {
					if i != 0 {
						mdl.FilterIDs += utils.InfieldSep
					}
					mdl.FilterIDs += val
				}

				if tPrf.ActivationInterval != nil {
					if tPrf.ActivationInterval.ActivationTime != utils.EmptyString {
						mdl.ActivationInterval = tPrf.ActivationInterval.ActivationTime
					}
					if tPrf.ActivationInterval.ExpiryTime != utils.EmptyString {
						mdl.ActivationInterval += utils.InfieldSep + tPrf.ActivationInterval.ExpiryTime
					}
				}
				mdl.Weight = tPrf.Weight
				mdl.MinCost = tPrf.MinCost
				mdl.MaxCost = tPrf.MaxCost
				mdl.MaxCostStrategy = tPrf.MaxCostStrategy
			}
			mdl.RateID = rate.ID
			if j == 0 {
				for i, val := range rate.FilterIDs {
					if i != 0 {
						mdl.RateFilterIDs += utils.InfieldSep
					}
					mdl.RateFilterIDs += val
				}
				mdl.RateWeight = rate.Weight
				mdl.RateActivationTimes = rate.ActivationTimes
				mdl.RateBlocker = rate.Blocker

			}
			mdl.RateRecurrentFee = intervalRate.RecurrentFee
			mdl.RateFixedFee = intervalRate.FixedFee
			mdl.RateUnit = intervalRate.Unit
			mdl.RateIncrement = intervalRate.Increment
			mdl.RateIntervalStart = intervalRate.IntervalStart
			mdls = append(mdls, mdl)
			i++
		}

	}
	return
}

func APItoRateProfile(tpRp *utils.TPRateProfile, timezone string) (rp *RateProfile, err error) {
	rp = &RateProfile{
		Tenant:          tpRp.Tenant,
		ID:              tpRp.ID,
		FilterIDs:       make([]string, len(tpRp.FilterIDs)),
		Weight:          tpRp.Weight,
		MaxCostStrategy: tpRp.MaxCostStrategy,
		Rates:           make(map[string]*Rate),
		MinCost:         utils.NewDecimalFromFloat64(tpRp.MinCost),
		MaxCost:         utils.NewDecimalFromFloat64(tpRp.MaxCost),
	}
	for i, stp := range tpRp.FilterIDs {
		rp.FilterIDs[i] = stp
	}
	if tpRp.ActivationInterval != nil {
		if rp.ActivationInterval, err = tpRp.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return nil, err
		}
	}
	for key, rate := range tpRp.Rates {
		rp.Rates[key] = &Rate{
			ID:              rate.ID,
			Weight:          rate.Weight,
			Blocker:         rate.Blocker,
			FilterIDs:       rate.FilterIDs,
			ActivationTimes: rate.ActivationTimes,
			IntervalRates:   make([]*IntervalRate, len(rate.IntervalRates)),
		}
		for i, iRate := range rate.IntervalRates {
			rp.Rates[key].IntervalRates[i] = new(IntervalRate)
			if rp.Rates[key].IntervalRates[i].IntervalStart, err = utils.ParseDurationWithNanosecs(iRate.IntervalStart); err != nil {
				return nil, err
			}
			rp.Rates[key].IntervalRates[i].FixedFee = utils.NewDecimalFromFloat64(iRate.FixedFee)
			rp.Rates[key].IntervalRates[i].RecurrentFee = utils.NewDecimalFromFloat64(iRate.RecurrentFee)
			if rp.Rates[key].IntervalRates[i].Unit, err = utils.NewDecimalFromUsage(iRate.Unit); err != nil {
				return nil, err
			}
			if rp.Rates[key].IntervalRates[i].Increment, err = utils.NewDecimalFromUsage(iRate.Increment); err != nil {
				return nil, err
			}
		}
	}
	return rp, nil
}

func RateProfileToAPI(rp *RateProfile) (tpRp *utils.TPRateProfile) {
	tpRp = &utils.TPRateProfile{
		Tenant:             rp.Tenant,
		ID:                 rp.ID,
		FilterIDs:          make([]string, len(rp.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		Weight:             rp.Weight,
		MaxCostStrategy:    rp.MaxCostStrategy,
		Rates:              make(map[string]*utils.TPRate),
	}
	if rp.MinCost != nil {
		//there should not be an invalid value of converting from Decimal into float64
		minCostF, _ := rp.MinCost.Float64()
		tpRp.MinCost = minCostF
	}
	if rp.MaxCost != nil {
		//there should not be an invalid value of converting from Decimal into float64
		maxCostF, _ := rp.MaxCost.Float64()
		tpRp.MaxCost = maxCostF
	}

	for key, rate := range rp.Rates {
		tpRp.Rates[key] = &utils.TPRate{
			ID:              rate.ID,
			Weight:          rate.Weight,
			Blocker:         rate.Blocker,
			FilterIDs:       rate.FilterIDs,
			ActivationTimes: rate.ActivationTimes,
			IntervalRates:   make([]*utils.TPIntervalRate, len(rate.IntervalRates)),
		}
		for i, iRate := range rate.IntervalRates {
			tpRp.Rates[key].IntervalRates[i] = &utils.TPIntervalRate{
				IntervalStart: iRate.IntervalStart.String(),
			}
			if iRate.FixedFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				fixedFeeF, _ := iRate.FixedFee.Float64()
				tpRp.Rates[key].IntervalRates[i].FixedFee = fixedFeeF
			}
			if iRate.Unit != nil {
				tpRp.Rates[key].IntervalRates[i].Unit = iRate.Unit.String()
			}
			if iRate.Increment != nil {
				tpRp.Rates[key].IntervalRates[i].Increment = iRate.Increment.String()
			}
			if iRate.RecurrentFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				recFeeF, _ := iRate.RecurrentFee.Float64()
				tpRp.Rates[key].IntervalRates[i].RecurrentFee = recFeeF
			}
		}
	}
	for i, fli := range rp.FilterIDs {
		tpRp.FilterIDs[i] = fli
	}
	if rp.ActivationInterval != nil {
		if !rp.ActivationInterval.ActivationTime.IsZero() {
			tpRp.ActivationInterval.ActivationTime = rp.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !rp.ActivationInterval.ExpiryTime.IsZero() {
			tpRp.ActivationInterval.ExpiryTime = rp.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	return
}

type ActionProfileMdls []*ActionProfileMdl

// CSVHeader return the header for csv fields as a slice of string
func (apm ActionProfileMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.ActivationIntervalString, utils.Weight, utils.Schedule, utils.TargetType,
		utils.TargetIDs, utils.ActionID, utils.ActionFilterIDs, utils.ActionBlocker, utils.ActionTTL,
		utils.ActionType, utils.ActionOpts, utils.ActionPath, utils.ActionValue,
	}
}

func (tps ActionProfileMdls) AsTPActionProfile() (result []*utils.TPActionProfile) {
	filterIDsMap := make(map[string]utils.StringSet)
	targetIDsMap := make(map[string]map[string]utils.StringSet)
	actPrfMap := make(map[string]*utils.TPActionProfile)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		aPrf, found := actPrfMap[tenID]
		if !found {
			aPrf = &utils.TPActionProfile{
				TPid:   tp.Tpid,
				Tenant: tp.Tenant,
				ID:     tp.ID,
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterIDsMap[tenID]; !has {
				filterIDsMap[tenID] = make(utils.StringSet)
			}
			filterIDsMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.ActivationInterval != utils.EmptyString {
			aPrf.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				aPrf.ActivationInterval.ActivationTime = aiSplt[0]
				aPrf.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				aPrf.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.Weight != 0 {
			aPrf.Weight = tp.Weight
		}
		if tp.Schedule != utils.EmptyString {
			aPrf.Schedule = tp.Schedule
		}
		if tp.TargetType != utils.EmptyString {
			if _, has := targetIDsMap[tenID]; !has {
				targetIDsMap[tenID] = make(map[string]utils.StringSet)
			}
			targetIDsMap[tenID][tp.TargetType] = utils.NewStringSet(strings.Split(tp.TargetIDs, utils.InfieldSep))
		}

		if tp.ActionID != utils.EmptyString {
			tpAAction := &utils.TPAPAction{
				ID:      tp.ActionID,
				Blocker: tp.ActionBlocker,
				TTL:     tp.ActionTTL,
				Type:    tp.ActionType,
				Opts:    tp.ActionOpts,
				Path:    tp.ActionPath,
				Value:   tp.ActionValue,
			}
			if tp.ActionFilterIDs != utils.EmptyString {
				filterIDs := make(utils.StringSet)
				filterIDs.AddSlice(strings.Split(tp.ActionFilterIDs, utils.InfieldSep))
				tpAAction.FilterIDs = filterIDs.AsSlice()
			}
			aPrf.Actions = append(aPrf.Actions, tpAAction)
		}
		actPrfMap[tenID] = aPrf
	}
	result = make([]*utils.TPActionProfile, len(actPrfMap))
	i := 0
	for tntID, th := range actPrfMap {
		result[i] = th
		result[i].FilterIDs = filterIDsMap[tntID].AsSlice()
		for targetType, targetIDs := range targetIDsMap[tntID] {
			result[i].Targets = append(result[i].Targets, &utils.TPActionTarget{TargetType: targetType, TargetIDs: targetIDs.AsSlice()})
		}
		i++
	}
	return
}

func APItoModelTPActionProfile(tPrf *utils.TPActionProfile) (mdls ActionProfileMdls) {
	if len(tPrf.Actions) == 0 {
		return
	}
	i := 0
	for _, action := range tPrf.Actions {
		mdl := &ActionProfileMdl{
			Tenant: tPrf.Tenant,
			Tpid:   tPrf.TPid,
			ID:     tPrf.ID,
		}
		if i == 0 {
			for i, val := range tPrf.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.InfieldSep
				}
				mdl.FilterIDs += val
			}

			if tPrf.ActivationInterval != nil {
				if tPrf.ActivationInterval.ActivationTime != utils.EmptyString {
					mdl.ActivationInterval = tPrf.ActivationInterval.ActivationTime
				}
				if tPrf.ActivationInterval.ExpiryTime != utils.EmptyString {
					mdl.ActivationInterval += utils.InfieldSep + tPrf.ActivationInterval.ExpiryTime
				}
			}
			mdl.Weight = tPrf.Weight
			mdl.Schedule = tPrf.Schedule
			for _, target := range tPrf.Targets {
				mdl.TargetType = target.TargetType
				mdl.TargetIDs = strings.Join(target.TargetIDs, utils.InfieldSep)
			}
		}
		mdl.ActionID = action.ID
		for i, val := range action.FilterIDs {
			if i != 0 {
				mdl.ActionFilterIDs += utils.InfieldSep
			}
			mdl.ActionFilterIDs += val
		}
		mdl.ActionBlocker = action.Blocker
		mdl.ActionTTL = action.TTL
		mdl.ActionType = action.Type
		mdl.ActionOpts = action.Opts
		mdl.ActionPath = action.Path
		mdl.ActionValue = action.Value
		mdls = append(mdls, mdl)
		i++
	}
	return
}

func APItoActionProfile(tpAp *utils.TPActionProfile, timezone string) (ap *ActionProfile, err error) {
	ap = &ActionProfile{
		Tenant:    tpAp.Tenant,
		ID:        tpAp.ID,
		FilterIDs: make([]string, len(tpAp.FilterIDs)),
		Weight:    tpAp.Weight,
		Schedule:  tpAp.Schedule,
		Targets:   make(map[string]utils.StringSet),
		Actions:   make([]*APAction, len(tpAp.Actions)),
	}
	for i, stp := range tpAp.FilterIDs {
		ap.FilterIDs[i] = stp
	}
	if tpAp.ActivationInterval != nil {
		if ap.ActivationInterval, err = tpAp.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return
		}
	}
	for _, target := range tpAp.Targets {
		ap.Targets[target.TargetType] = utils.NewStringSet(target.TargetIDs)
	}
	for i, act := range tpAp.Actions {
		if act.Path == utils.EmptyString {
			err = fmt.Errorf("empty path in ActionProfile <%s> for action <%s>", ap.TenantID(), act.ID)
			return
		}
		ap.Actions[i] = &APAction{
			ID:        act.ID,
			FilterIDs: act.FilterIDs,
			Blocker:   act.Blocker,
			Type:      act.Type,
			Path:      act.Path,
		}
		if ap.Actions[i].TTL, err = utils.ParseDurationWithNanosecs(act.TTL); err != nil {
			return
		}
		if act.Opts != utils.EmptyString {
			ap.Actions[i].Opts = make(map[string]interface{})
			for _, opt := range strings.Split(act.Opts, utils.InfieldSep) { // example of opts: key1:val1;key2:val2;key3:val3
				keyValSls := utils.SplitConcatenatedKey(opt)
				if len(keyValSls) != 2 {
					err = fmt.Errorf("malformed option for ActionProfile <%s> for action <%s>", ap.TenantID(), act.ID)
					return
				}
				ap.Actions[i].Opts[keyValSls[0]] = keyValSls[1]
			}
		}
		if ap.Actions[i].Value, err = config.NewRSRParsers(act.Value, config.CgrConfig().GeneralCfg().RSRSep); err != nil {
			return
		}
	}
	return
}

func ActionProfileToAPI(ap *ActionProfile) (tpAp *utils.TPActionProfile) {
	tpAp = &utils.TPActionProfile{
		Tenant:             ap.Tenant,
		ID:                 ap.ID,
		FilterIDs:          make([]string, len(ap.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		Weight:             ap.Weight,
		Schedule:           ap.Schedule,
		Targets:            make([]*utils.TPActionTarget, 0, len(ap.Targets)),
		Actions:            make([]*utils.TPAPAction, len(ap.Actions)),
	}
	for i, fli := range ap.FilterIDs {
		tpAp.FilterIDs[i] = fli
	}
	if ap.ActivationInterval != nil {
		if !ap.ActivationInterval.ActivationTime.IsZero() {
			tpAp.ActivationInterval.ActivationTime = ap.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !ap.ActivationInterval.ExpiryTime.IsZero() {
			tpAp.ActivationInterval.ExpiryTime = ap.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}
	for targetType, targetIDs := range ap.Targets {
		tpAp.Targets = append(tpAp.Targets, &utils.TPActionTarget{TargetType: targetType, TargetIDs: targetIDs.AsSlice()})
	}
	for i, act := range ap.Actions {
		tpAp.Actions[i] = &utils.TPAPAction{
			ID:        act.ID,
			FilterIDs: act.FilterIDs,
			Blocker:   act.Blocker,
			TTL:       act.TTL.String(),
			Type:      act.Type,
			Path:      act.Path,
			Value:     act.Value.GetRule(config.CgrConfig().GeneralCfg().RSRSep),
		}

		elems := make([]string, 0, len(act.Opts))
		for k, v := range act.Opts {
			elems = append(elems, utils.ConcatenatedKey(k, utils.IfaceAsString(v)))
		}
		tpAp.Actions[i].Opts = strings.Join(elems, utils.InfieldSep)
	}
	return
}

type AccountProfileMdls []*AccountProfileMdl

// CSVHeader return the header for csv fields as a slice of string
func (apm AccountProfileMdls) CSVHeader() (result []string) {
	return []string{"#" + utils.Tenant, utils.ID, utils.FilterIDs,
		utils.ActivationIntervalString, utils.Weight, utils.BalanceID,
		utils.BalanceFilterIDs, utils.BalanceWeight, utils.BalanceBlocker,
		utils.BalanceType, utils.BalanceOpts, utils.BalanceUnits, utils.ThresholdIDs,
	}
}

func (tps AccountProfileMdls) AsTPAccountProfile() (result []*utils.TPAccountProfile, err error) {
	filterIDsMap := make(map[string]utils.StringSet)
	thresholdIDsMap := make(map[string]utils.StringSet)
	actPrfMap := make(map[string]*utils.TPAccountProfile)
	for _, tp := range tps {
		tenID := (&utils.TenantID{Tenant: tp.Tenant, ID: tp.ID}).TenantID()
		aPrf, found := actPrfMap[tenID]
		if !found {
			aPrf = &utils.TPAccountProfile{
				TPid:     tp.Tpid,
				Tenant:   tp.Tenant,
				ID:       tp.ID,
				Weights:  tp.Weights,
				Balances: make(map[string]*utils.TPAccountBalance),
			}
		}
		if tp.FilterIDs != utils.EmptyString {
			if _, has := filterIDsMap[tenID]; !has {
				filterIDsMap[tenID] = make(utils.StringSet)
			}
			filterIDsMap[tenID].AddSlice(strings.Split(tp.FilterIDs, utils.InfieldSep))
		}
		if tp.ThresholdIDs != utils.EmptyString {
			if _, has := thresholdIDsMap[tenID]; !has {
				thresholdIDsMap[tenID] = make(utils.StringSet)
			}
			thresholdIDsMap[tenID].AddSlice(strings.Split(tp.ThresholdIDs, utils.InfieldSep))
		}
		if tp.ActivationInterval != utils.EmptyString {
			aPrf.ActivationInterval = new(utils.TPActivationInterval)
			aiSplt := strings.Split(tp.ActivationInterval, utils.InfieldSep)
			if len(aiSplt) == 2 {
				aPrf.ActivationInterval.ActivationTime = aiSplt[0]
				aPrf.ActivationInterval.ExpiryTime = aiSplt[1]
			} else if len(aiSplt) == 1 {
				aPrf.ActivationInterval.ActivationTime = aiSplt[0]
			}
		}
		if tp.BalanceID != utils.EmptyString {
			aPrf.Balances[tp.BalanceID] = &utils.TPAccountBalance{
				ID:      tp.BalanceID,
				Weights: tp.BalanceWeights,
				Type:    tp.BalanceType,
				Opts:    tp.BalanceOpts,
				Units:   tp.BalanceUnits,
			}

			if tp.BalanceFilterIDs != utils.EmptyString {
				filterIDs := make(utils.StringSet)
				filterIDs.AddSlice(strings.Split(tp.BalanceFilterIDs, utils.InfieldSep))
				aPrf.Balances[tp.BalanceID].FilterIDs = filterIDs.AsSlice()
			}
			if tp.BalanceCostIncrements != utils.EmptyString {
				costIncrements := make([]*utils.TPBalanceCostIncrement, 0)
				sls := strings.Split(tp.BalanceCostIncrements, utils.InfieldSep)
				if len(sls)%4 != 0 {
					return nil, fmt.Errorf("invlid key: <%s> for BalanceCostIncrements", tp.BalanceCostIncrements)
				}
				for j := 0; j < len(sls); j = j + 4 {
					costIncrement, err := utils.NewTPBalanceCostIncrement(sls[j], sls[j+1], sls[j+2], sls[j+3])
					if err != nil {
						return nil, err
					}
					costIncrements = append(costIncrements, costIncrement)
				}
				aPrf.Balances[tp.BalanceID].CostIncrement = costIncrements
			}
			if tp.BalanceAttributeIDs != utils.EmptyString {
				attributeIDs := make(utils.StringSet)
				attributeIDs.AddSlice(strings.Split(tp.BalanceAttributeIDs, utils.InfieldSep))
				aPrf.Balances[tp.BalanceID].AttributeIDs = attributeIDs.AsSlice()
			}
			if tp.BalanceRateProfileIDs != utils.EmptyString {
				rateProfileIDs := make(utils.StringSet)
				rateProfileIDs.AddSlice(strings.Split(tp.BalanceRateProfileIDs, utils.InfieldSep))
				aPrf.Balances[tp.BalanceID].RateProfileIDs = rateProfileIDs.AsSlice()
			}
			if tp.BalanceUnitFactors != utils.EmptyString {
				unitFactors := make([]*utils.TPBalanceUnitFactor, 0)
				sls := strings.Split(tp.BalanceUnitFactors, utils.InfieldSep)
				if len(sls)%2 != 0 {
					return nil, fmt.Errorf("invlid key: <%s> for BalanceUnitFactors", tp.BalanceUnitFactors)
				}

				for j := 0; j < len(sls); j = j + 2 {
					unitFactor, err := utils.NewTPBalanceUnitFactor(sls[j], sls[j+1])
					if err != nil {
						return nil, err
					}
					unitFactors = append(unitFactors, unitFactor)
				}
				aPrf.Balances[tp.BalanceID].UnitFactors = unitFactors
			}

		}
		actPrfMap[tenID] = aPrf
	}
	result = make([]*utils.TPAccountProfile, len(actPrfMap))
	i := 0
	for tntID, th := range actPrfMap {
		result[i] = th
		result[i].FilterIDs = filterIDsMap[tntID].AsSlice()
		result[i].ThresholdIDs = thresholdIDsMap[tntID].AsSlice()
		i++
	}
	return
}

func APItoModelTPAccountProfile(tPrf *utils.TPAccountProfile) (mdls AccountProfileMdls) {
	if len(tPrf.Balances) == 0 {
		return
	}
	i := 0
	for _, balance := range tPrf.Balances {
		mdl := &AccountProfileMdl{
			Tenant: tPrf.Tenant,
			Tpid:   tPrf.TPid,
			ID:     tPrf.ID,
		}
		if i == 0 {
			for i, val := range tPrf.FilterIDs {
				if i != 0 {
					mdl.FilterIDs += utils.InfieldSep
				}
				mdl.FilterIDs += val
			}
			for i, val := range tPrf.ThresholdIDs {
				if i != 0 {
					mdl.ThresholdIDs += utils.InfieldSep
				}
				mdl.ThresholdIDs += val
			}
			if tPrf.ActivationInterval != nil {
				if tPrf.ActivationInterval.ActivationTime != utils.EmptyString {
					mdl.ActivationInterval = tPrf.ActivationInterval.ActivationTime
				}
				if tPrf.ActivationInterval.ExpiryTime != utils.EmptyString {
					mdl.ActivationInterval += utils.InfieldSep + tPrf.ActivationInterval.ExpiryTime
				}
			}
			mdl.Weights = tPrf.Weights
		}
		mdl.BalanceID = balance.ID
		for i, val := range balance.FilterIDs {
			if i != 0 {
				mdl.BalanceFilterIDs += utils.InfieldSep
			}
			mdl.BalanceFilterIDs += val
		}
		mdl.BalanceWeights = balance.Weights
		mdl.BalanceType = balance.Type
		mdl.BalanceOpts = balance.Opts
		for i, costIncr := range balance.CostIncrement {
			if i != 0 {
				mdl.BalanceCostIncrements += utils.InfieldSep
			}
			mdl.BalanceCostIncrements += costIncr.AsString()
		}
		for i, attrID := range balance.AttributeIDs {
			if i != 0 {
				mdl.BalanceAttributeIDs += utils.InfieldSep
			}
			mdl.BalanceAttributeIDs += attrID
		}
		for i, ratePrfID := range balance.RateProfileIDs {
			if i != 0 {
				mdl.BalanceRateProfileIDs += utils.InfieldSep
			}
			mdl.BalanceRateProfileIDs += ratePrfID
		}
		for i, unitFactor := range balance.UnitFactors {
			if i != 0 {
				mdl.BalanceUnitFactors += utils.InfieldSep
			}
			mdl.BalanceUnitFactors += unitFactor.AsString()
		}
		mdl.BalanceUnits = balance.Units
		mdls = append(mdls, mdl)
		i++
	}
	return
}

func APItoAccountProfile(tpAp *utils.TPAccountProfile, timezone string) (ap *utils.AccountProfile, err error) {
	ap = &utils.AccountProfile{
		Tenant:       tpAp.Tenant,
		ID:           tpAp.ID,
		FilterIDs:    make([]string, len(tpAp.FilterIDs)),
		Balances:     make(map[string]*utils.Balance, len(tpAp.Balances)),
		ThresholdIDs: make([]string, len(tpAp.ThresholdIDs)),
	}
	if tpAp.Weights != utils.EmptyString {
		weight, err := utils.NewDynamicWeightsFromString(tpAp.Weights, ";", "&")
		if err != nil {
			return nil, err
		}
		ap.Weights = weight
	}
	for i, stp := range tpAp.FilterIDs {
		ap.FilterIDs[i] = stp
	}
	if tpAp.ActivationInterval != nil {
		if ap.ActivationInterval, err = tpAp.ActivationInterval.AsActivationInterval(timezone); err != nil {
			return
		}
	}

	for id, bal := range tpAp.Balances {
		ap.Balances[id] = &utils.Balance{
			ID:        bal.ID,
			FilterIDs: bal.FilterIDs,
			Type:      bal.Type,
			Units:     utils.NewDecimalFromFloat64(bal.Units),
		}
		if bal.Opts != utils.EmptyString {
			ap.Balances[id].Opts = make(map[string]interface{})
			for _, opt := range strings.Split(bal.Opts, utils.InfieldSep) { // example of opts: key1:val1;key2:val2;key3:val3
				keyValSls := utils.SplitConcatenatedKey(opt)
				if len(keyValSls) != 2 {
					err = fmt.Errorf("malformed option for ActionProfile <%s> for action <%s>", ap.TenantID(), bal.ID)
					return
				}
				ap.Balances[id].Opts[keyValSls[0]] = keyValSls[1]
			}
		}
		if bal.Weights != utils.EmptyString {
			weight, err := utils.NewDynamicWeightsFromString(bal.Weights, ";", "&")
			if err != nil {
				return nil, err
			}
			ap.Balances[id].Weights = weight
		}
		if bal.CostIncrement != nil {
			ap.Balances[id].CostIncrements = make([]*utils.CostIncrement, len(bal.CostIncrement))
			for j, costIncrement := range bal.CostIncrement {
				ap.Balances[id].CostIncrements[j] = &utils.CostIncrement{
					FilterIDs: costIncrement.FilterIDs,
				}
				if costIncrement.Increment != nil {
					ap.Balances[id].CostIncrements[j].Increment = utils.NewDecimalFromFloat64(*costIncrement.Increment)
				}
				if costIncrement.FixedFee != nil {
					ap.Balances[id].CostIncrements[j].FixedFee = utils.NewDecimalFromFloat64(*costIncrement.FixedFee)
				}
				if costIncrement.RecurrentFee != nil {
					ap.Balances[id].CostIncrements[j].RecurrentFee = utils.NewDecimalFromFloat64(*costIncrement.RecurrentFee)
				}
			}
		}
		if bal.AttributeIDs != nil {
			ap.Balances[id].AttributeIDs = make([]string, len(bal.AttributeIDs))
			for j, costAttribute := range bal.AttributeIDs {
				ap.Balances[id].AttributeIDs[j] = costAttribute
			}
		}
		if bal.RateProfileIDs != nil {
			ap.Balances[id].RateProfileIDs = make([]string, len(bal.RateProfileIDs))
			for j, costAttribute := range bal.RateProfileIDs {
				ap.Balances[id].RateProfileIDs[j] = costAttribute
			}
		}
		if bal.UnitFactors != nil {
			ap.Balances[id].UnitFactors = make([]*utils.UnitFactor, len(bal.UnitFactors))
			for j, unitFactor := range bal.UnitFactors {
				ap.Balances[id].UnitFactors[j] = &utils.UnitFactor{
					FilterIDs: unitFactor.FilterIDs,
					Factor:    utils.NewDecimalFromFloat64(unitFactor.Factor),
				}
			}
		}
	}
	for i, stp := range tpAp.ThresholdIDs {
		ap.ThresholdIDs[i] = stp
	}
	return
}

func AccountProfileToAPI(ap *utils.AccountProfile) (tpAp *utils.TPAccountProfile) {
	tpAp = &utils.TPAccountProfile{
		Tenant:             ap.Tenant,
		ID:                 ap.ID,
		Weights:            ap.Weights.String(";", "&"),
		FilterIDs:          make([]string, len(ap.FilterIDs)),
		ActivationInterval: new(utils.TPActivationInterval),
		Balances:           make(map[string]*utils.TPAccountBalance, len(ap.Balances)),
		ThresholdIDs:       make([]string, len(ap.ThresholdIDs)),
	}
	for i, fli := range ap.FilterIDs {
		tpAp.FilterIDs[i] = fli
	}
	if ap.ActivationInterval != nil {
		if !ap.ActivationInterval.ActivationTime.IsZero() {
			tpAp.ActivationInterval.ActivationTime = ap.ActivationInterval.ActivationTime.Format(time.RFC3339)
		}
		if !ap.ActivationInterval.ExpiryTime.IsZero() {
			tpAp.ActivationInterval.ExpiryTime = ap.ActivationInterval.ExpiryTime.Format(time.RFC3339)
		}
	}

	for i, bal := range ap.Balances {
		tpAp.Balances[i] = &utils.TPAccountBalance{
			ID:             bal.ID,
			FilterIDs:      make([]string, len(bal.FilterIDs)),
			Weights:        bal.Weights.String(";", "&"),
			Type:           bal.Type,
			CostIncrement:  make([]*utils.TPBalanceCostIncrement, len(bal.CostIncrements)),
			AttributeIDs:   make([]string, len(bal.AttributeIDs)),
			RateProfileIDs: make([]string, len(bal.RateProfileIDs)),
			UnitFactors:    make([]*utils.TPBalanceUnitFactor, len(bal.UnitFactors)),
		}
		for k, fli := range bal.FilterIDs {
			tpAp.Balances[i].FilterIDs[k] = fli
		}
		//there should not be an invalid value of converting into float64
		tpAp.Balances[i].Units, _ = bal.Units.Float64()
		elems := make([]string, 0, len(bal.Opts))
		for k, v := range bal.Opts {
			elems = append(elems, utils.ConcatenatedKey(k, utils.IfaceAsString(v)))
		}
		for k, cIncrement := range bal.CostIncrements {
			tpAp.Balances[i].CostIncrement[k] = &utils.TPBalanceCostIncrement{
				FilterIDs: make([]string, len(cIncrement.FilterIDs)),
			}
			for kk, fli := range cIncrement.FilterIDs {
				tpAp.Balances[i].CostIncrement[k].FilterIDs[kk] = fli
			}
			if cIncrement.Increment != nil {
				//there should not be an invalid value of converting from Decimal into float64
				incr, _ := cIncrement.Increment.Float64()
				tpAp.Balances[i].CostIncrement[k].Increment = &incr
			}
			if cIncrement.FixedFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				fxdFee, _ := cIncrement.FixedFee.Float64()
				tpAp.Balances[i].CostIncrement[k].FixedFee = &fxdFee
			}
			if cIncrement.RecurrentFee != nil {
				//there should not be an invalid value of converting from Decimal into float64
				rcrFee, _ := cIncrement.RecurrentFee.Float64()
				tpAp.Balances[i].CostIncrement[k].RecurrentFee = &rcrFee
			}
		}
		for k, attrID := range bal.AttributeIDs {
			tpAp.Balances[i].AttributeIDs[k] = attrID
		}
		for k, ratePrfID := range bal.RateProfileIDs {
			tpAp.Balances[i].RateProfileIDs[k] = ratePrfID
		}
		for k, uFactor := range bal.UnitFactors {
			tpAp.Balances[i].UnitFactors[k] = &utils.TPBalanceUnitFactor{
				FilterIDs: make([]string, len(uFactor.FilterIDs)),
			}
			for kk, fli := range uFactor.FilterIDs {
				tpAp.Balances[i].UnitFactors[k].FilterIDs[kk] = fli
			}
			if uFactor.Factor != nil {
				//there should not be an invalid value of converting from Decimal into float64
				untFctr, _ := uFactor.Factor.Float64()
				tpAp.Balances[i].UnitFactors[k].Factor = untFctr
			}
		}
		tpAp.Balances[i].Opts = strings.Join(elems, utils.InfieldSep)
	}
	for i, fli := range ap.ThresholdIDs {
		tpAp.ThresholdIDs[i] = fli
	}
	return
}
