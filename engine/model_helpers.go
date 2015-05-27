package engine

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

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
	for fildName, fieldValue := range fieldValueMap {
		field := elem.FieldByName(fildName)
		if field.IsValid() && field.CanSet() {
			switch field.Kind() {
			case reflect.Float64:
				value, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return nil, fmt.Errorf(`invalid value "%s" for field %s.%s`, fieldValue, st.Name(), fildName)
				}
				field.SetFloat(value)
			case reflect.String:
				field.SetString(fieldValue)
			}
		}
	}
	return elem.Interface(), nil
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

type TpDestinations []*TpDestination

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

type TpTimings []*TpTiming

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

type TpRates []*TpRate

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

type TpDestinationRates []*TpDestinationRate

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

type TpRatingPlans []*TpRatingPlan

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

type TpRatingProfiles []*TpRatingProfile

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
