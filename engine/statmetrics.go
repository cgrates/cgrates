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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

//to be moved in utils
const STATS_NA = -1.0

// NewStatMetric instantiates the StatMetric
// cfg serves as general purpose container to pass config options to metric
func NewStatMetric(metricID string, minItems int, extraParams string) (sm StatMetric, err error) {
	metrics := map[string]func(int, string) (StatMetric, error){
		utils.MetaASR:     NewASR,
		utils.MetaACD:     NewACD,
		utils.MetaTCD:     NewTCD,
		utils.MetaACC:     NewACC,
		utils.MetaTCC:     NewTCC,
		utils.MetaPDD:     NewPDD,
		utils.MetaDDC:     NewDCC,
		utils.MetaSum:     NewStatSum,
		utils.MetaAverage: NewStatAverage,
	}
	metricType := strings.Split(metricID, utils.InInFieldSep)[0]
	if _, has := metrics[metricType]; !has {
		return nil, fmt.Errorf("unsupported metric type <%s>", metricType)
	}
	return metrics[metricType](minItems, extraParams)
}

// StatMetric is the interface which a metric should implement
type StatMetric interface {
	GetValue() interface{}
	GetStringValue(fmtOpts string) (val string)
	GetFloat64Value() (val float64)
	AddEvent(ev *utils.CGREvent) error
	RemEvent(evTenantID string) error
	Marshal(ms Marshaler) (marshaled []byte, err error)
	LoadMarshaled(ms Marshaler, marshaled []byte) (err error)
}

func NewASR(minItems int, extraParams string) (StatMetric, error) {
	return &StatASR{Events: make(map[string]bool), MinItems: minItems}, nil
}

// ASR implements AverageSuccessRatio metric
type StatASR struct {
	Answered float64
	Count    float64
	Events   map[string]bool // map[EventTenantID]Answered
	MinItems int
	val      *float64 // cached ASR value
}

// getValue returns asr.val
func (asr *StatASR) getValue() float64 {
	if asr.val == nil {
		if (asr.MinItems > 0 && len(asr.Events) < asr.MinItems) || (asr.Count == 0) {
			asr.val = utils.Float64Pointer(STATS_NA)
		} else {
			asr.val = utils.Float64Pointer(utils.Round((asr.Answered / asr.Count * 100),
				config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE))
		}
	}
	return *asr.val
}

// GetValue returns the ASR value as part of StatMetric interface
func (asr *StatASR) GetValue() (v interface{}) {
	return asr.getValue()
}

func (asr *StatASR) GetStringValue(fmtOpts string) (valStr string) {
	if val := asr.getValue(); val == STATS_NA {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = fmt.Sprintf("%v%%", asr.getValue())
	}
	return
}

// GetFloat64Value is part of StatMetric interface
func (asr *StatASR) GetFloat64Value() (val float64) {
	return asr.getValue()
}

// AddEvent is part of StatMetric interface
func (asr *StatASR) AddEvent(ev *utils.CGREvent) (err error) {
	var answered bool
	if at, err := ev.FieldAsTime(utils.AnswerTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		answered = true
	}
	asr.Events[ev.ID] = answered
	asr.Count += 1
	if answered {
		asr.Answered += 1
	}
	asr.val = nil
	return
}

func (asr *StatASR) RemEvent(evID string) (err error) {
	answered, has := asr.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if answered {
		asr.Answered -= 1
	}
	asr.Count -= 1
	delete(asr.Events, evID)
	asr.val = nil
	return
}

// Marshal is part of StatMetric interface
func (asr *StatASR) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(asr)
}

// LoadMarshaled is part of StatMetric interface
func (asr *StatASR) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, asr)
}

func NewACD(minItems int, extraParams string) (StatMetric, error) {
	return &StatACD{Events: make(map[string]time.Duration), MinItems: minItems}, nil
}

// ACD implements AverageCallDuration metric
type StatACD struct {
	Sum      time.Duration
	Count    int64
	Events   map[string]time.Duration // map[EventTenantID]Duration
	MinItems int
	val      *time.Duration // cached ACD value
}

// getValue returns acr.val
func (acd *StatACD) getValue() time.Duration {
	if acd.val == nil {
		if (acd.MinItems > 0 && len(acd.Events) < acd.MinItems) || (acd.Count == 0) {
			acd.val = utils.DurationPointer(time.Duration((-1) * time.Nanosecond))
		} else {
			acd.val = utils.DurationPointer(time.Duration(acd.Sum.Nanoseconds() / acd.Count))
		}
	}
	return *acd.val
}

func (acd *StatACD) GetStringValue(fmtOpts string) (valStr string) {
	if val := acd.getValue(); val == time.Duration((-1)*time.Nanosecond) {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = fmt.Sprintf("%+v", acd.getValue())
	}
	return
}

func (acd *StatACD) GetValue() (v interface{}) {
	return acd.getValue()
}

func (acd *StatACD) GetFloat64Value() (v float64) {
	if val := acd.getValue(); val == time.Duration((-1)*time.Nanosecond) {
		v = -1.0
	} else {
		v = acd.getValue().Seconds()
	}
	return
}

func (acd *StatACD) AddEvent(ev *utils.CGREvent) (err error) {
	var value time.Duration
	if at, err := ev.FieldAsTime(utils.AnswerTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if duration, err := ev.FieldAsDuration(utils.Usage); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else {
			value = duration
			acd.Sum += duration
		}
	}
	acd.Events[ev.ID] = value
	acd.Count += 1
	acd.val = nil
	return
}

func (acd *StatACD) RemEvent(evID string) (err error) {
	duration, has := acd.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if duration != 0 {
		acd.Sum -= duration
	}
	acd.Count -= 1
	delete(acd.Events, evID)
	acd.val = nil
	return
}

func (acd *StatACD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(acd)
}
func (acd *StatACD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, acd)
}

func NewTCD(minItems int, extraParams string) (StatMetric, error) {
	return &StatTCD{Events: make(map[string]time.Duration), MinItems: minItems}, nil
}

// TCD implements TotalCallDuration metric
type StatTCD struct {
	Sum      time.Duration
	Count    int64
	Events   map[string]time.Duration // map[EventTenantID]Duration
	MinItems int
	val      *time.Duration // cached TCD value
}

// getValue returns tcd.val
func (tcd *StatTCD) getValue() time.Duration {
	if tcd.val == nil {
		if (tcd.MinItems > 0 && len(tcd.Events) < tcd.MinItems) || (tcd.Count == 0) {
			tcd.val = utils.DurationPointer(time.Duration((-1) * time.Nanosecond))
		} else {
			tcd.val = utils.DurationPointer(time.Duration(tcd.Sum.Nanoseconds()))
		}
	}
	return *tcd.val
}

func (tcd *StatTCD) GetStringValue(fmtOpts string) (valStr string) {
	if val := tcd.getValue(); val == time.Duration((-1)*time.Nanosecond) {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = fmt.Sprintf("%+v", tcd.getValue())
	}
	return
}

func (tcd *StatTCD) GetValue() (v interface{}) {
	return tcd.getValue()
}

func (tcd *StatTCD) GetFloat64Value() (v float64) {
	if val := tcd.getValue(); val == time.Duration((-1)*time.Nanosecond) {
		v = -1.0
	} else {
		v = tcd.getValue().Seconds()
	}
	return
}

func (tcd *StatTCD) AddEvent(ev *utils.CGREvent) (err error) {
	var value time.Duration
	if at, err := ev.FieldAsTime(utils.AnswerTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if duration, err := ev.FieldAsDuration(utils.Usage); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else {
			value = duration
			tcd.Sum += duration
		}

	}
	tcd.Events[ev.ID] = value
	tcd.Count += 1
	tcd.val = nil
	return
}

func (tcd *StatTCD) RemEvent(evID string) (err error) {
	duration, has := tcd.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if duration != 0 {
		tcd.Sum -= duration
	}
	tcd.Count -= 1
	delete(tcd.Events, evID)
	tcd.val = nil
	return
}

func (tcd *StatTCD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(tcd)
}

func (tcd *StatTCD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, tcd)
}

func NewACC(minItems int, extraParams string) (StatMetric, error) {
	return &StatACC{Events: make(map[string]float64), MinItems: minItems}, nil
}

// ACC implements AverageCallCost metric
type StatACC struct {
	Sum      float64
	Count    float64
	Events   map[string]float64 // map[EventTenantID]Cost
	MinItems int
	val      *float64 // cached ACC value
}

// getValue returns tcd.val
func (acc *StatACC) getValue() float64 {
	if acc.val == nil {
		if (acc.MinItems > 0 && len(acc.Events) < acc.MinItems) || (acc.Count == 0) {
			acc.val = utils.Float64Pointer(STATS_NA)
		} else {
			acc.val = utils.Float64Pointer(utils.Round((acc.Sum / acc.Count),
				config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE))
		}
	}
	return *acc.val
}

func (acc *StatACC) GetStringValue(fmtOpts string) (valStr string) {
	if val := acc.getValue(); val == STATS_NA {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = strconv.FormatFloat(acc.getValue(), 'f', -1, 64)
	}
	return

}

func (acc *StatACC) GetValue() (v interface{}) {
	return acc.getValue()
}

func (acc *StatACC) GetFloat64Value() (v float64) {
	return acc.getValue()
}

func (acc *StatACC) AddEvent(ev *utils.CGREvent) (err error) {
	var value float64
	if at, err := ev.FieldAsTime(utils.AnswerTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if cost, err := ev.FieldAsFloat64(utils.COST); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else if cost >= 0 {
			value = cost
			acc.Sum += cost
		}
	}
	acc.Events[ev.ID] = value
	acc.Count += 1
	acc.val = nil
	return
}

func (acc *StatACC) RemEvent(evID string) (err error) {
	cost, has := acc.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if cost >= 0 {
		acc.Sum -= cost
	}
	acc.Count -= 1
	delete(acc.Events, evID)
	acc.val = nil
	return
}

func (acc *StatACC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(acc)
}

func (acc *StatACC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, acc)
}

func NewTCC(minItems int, extraParams string) (StatMetric, error) {
	return &StatTCC{Events: make(map[string]float64), MinItems: minItems}, nil
}

// TCC implements TotalCallCost metric
type StatTCC struct {
	Sum      float64
	Count    float64
	Events   map[string]float64 // map[EventTenantID]Cost
	MinItems int
	val      *float64 // cached TCC value
}

// getValue returns tcd.val
func (tcc *StatTCC) getValue() float64 {
	if tcc.val == nil {
		if (tcc.MinItems > 0 && len(tcc.Events) < tcc.MinItems) || (tcc.Count == 0) {
			tcc.val = utils.Float64Pointer(STATS_NA)
		} else {
			tcc.val = utils.Float64Pointer(utils.Round(tcc.Sum,
				config.CgrConfig().GeneralCfg().RoundingDecimals,
				utils.ROUNDING_MIDDLE))
		}
	}
	return *tcc.val
}

func (tcc *StatTCC) GetStringValue(fmtOpts string) (valStr string) {
	if val := tcc.getValue(); val == STATS_NA {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = strconv.FormatFloat(tcc.getValue(), 'f', -1, 64)
	}
	return
}

func (tcc *StatTCC) GetValue() (v interface{}) {
	return tcc.getValue()
}

func (tcc *StatTCC) GetFloat64Value() (v float64) {
	return tcc.getValue()
}

func (tcc *StatTCC) AddEvent(ev *utils.CGREvent) (err error) {
	var value float64
	if at, err := ev.FieldAsTime(utils.AnswerTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if cost, err := ev.FieldAsFloat64(utils.COST); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else if cost >= 0 {
			value = cost
			tcc.Sum += cost
		}
	}
	tcc.Events[ev.ID] = value
	tcc.Count += 1
	tcc.val = nil
	return
}

func (tcc *StatTCC) RemEvent(evID string) (err error) {
	cost, has := tcc.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if cost != 0 {
		tcc.Sum -= cost
	}
	tcc.Count -= 1
	delete(tcc.Events, evID)
	tcc.val = nil
	return
}

func (tcc *StatTCC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(tcc)
}

func (tcc *StatTCC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, tcc)
}

func NewPDD(minItems int, extraParams string) (StatMetric, error) {
	return &StatPDD{Events: make(map[string]time.Duration), MinItems: minItems}, nil
}

// PDD implements Post Dial Delay (average) metric
type StatPDD struct {
	Sum      time.Duration
	Count    int64
	Events   map[string]time.Duration // map[EventTenantID]Duration
	MinItems int
	val      *time.Duration // cached PDD value
}

// getValue returns pdd.val
func (pdd *StatPDD) getValue() time.Duration {
	if pdd.val == nil {
		if (pdd.MinItems > 0 && len(pdd.Events) < pdd.MinItems) || (pdd.Count == 0) {
			pdd.val = utils.DurationPointer(time.Duration((-1) * time.Nanosecond))
		} else {
			pdd.val = utils.DurationPointer(time.Duration(pdd.Sum.Nanoseconds() / pdd.Count))
		}
	}
	return *pdd.val
}

func (pdd *StatPDD) GetStringValue(fmtOpts string) (valStr string) {
	if val := pdd.getValue(); val == time.Duration((-1)*time.Nanosecond) {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = fmt.Sprintf("%+v", pdd.getValue())
	}
	return
}

func (pdd *StatPDD) GetValue() (v interface{}) {
	return pdd.getValue()
}

func (pdd *StatPDD) GetFloat64Value() (v float64) {
	if val := pdd.getValue(); val == time.Duration((-1)*time.Nanosecond) {
		v = -1.0
	} else {
		v = pdd.getValue().Seconds()
	}
	return
}

func (pdd *StatPDD) AddEvent(ev *utils.CGREvent) (err error) {
	var value time.Duration
	if at, err := ev.FieldAsTime(utils.AnswerTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		if duration, err := ev.FieldAsDuration(utils.PDD); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else {
			value = duration
			pdd.Sum += duration
		}
	}
	pdd.Events[ev.ID] = value
	pdd.Count += 1
	pdd.val = nil
	return
}

func (pdd *StatPDD) RemEvent(evID string) (err error) {
	duration, has := pdd.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if duration != 0 {
		pdd.Sum -= duration
	}
	pdd.Count -= 1
	delete(pdd.Events, evID)
	pdd.val = nil
	return
}

func (pdd *StatPDD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(pdd)
}
func (pdd *StatPDD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, pdd)
}

func NewDCC(minItems int, extraParams string) (StatMetric, error) {
	return &StatDDC{Destinations: make(map[string]utils.StringMap), Events: make(map[string]string), MinItems: minItems}, nil
}

// DDC implements Destination Distinct Count metric
type StatDDC struct {
	Destinations map[string]utils.StringMap
	Events       map[string]string // map[EventTenantID]Destination
	MinItems     int
}

func (ddc *StatDDC) GetStringValue(fmtOpts string) (valStr string) {
	if val := len(ddc.Destinations); (val == 0) || (ddc.MinItems > 0 && len(ddc.Events) < ddc.MinItems) {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = fmt.Sprintf("%+v", len(ddc.Destinations))
	}
	return
}

func (ddc *StatDDC) GetValue() (v interface{}) {
	return len(ddc.Destinations)
}

func (ddc *StatDDC) GetFloat64Value() (v float64) {
	if val := len(ddc.Destinations); (val == 0) || (ddc.MinItems > 0 && len(ddc.Events) < ddc.MinItems) {
		v = -1.0
	} else {
		v = float64(len(ddc.Destinations))
	}
	return
}

func (ddc *StatDDC) AddEvent(ev *utils.CGREvent) (err error) {
	var dest string
	if dest, err = ev.FieldAsString(utils.Destination); err != nil {
		return err
	}
	if _, has := ddc.Destinations[dest]; !has {
		ddc.Destinations[dest] = make(map[string]bool)
	}
	ddc.Destinations[dest][ev.ID] = true
	ddc.Events[ev.ID] = dest
	return
}

func (ddc *StatDDC) RemEvent(evID string) (err error) {
	destination, has := ddc.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	delete(ddc.Events, evID)
	if len(ddc.Destinations[destination]) == 1 {
		delete(ddc.Destinations, destination)
		return
	}
	delete(ddc.Destinations[destination], evID)
	return
}

func (ddc *StatDDC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(ddc)
}
func (ddc *StatDDC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, ddc)
}
func NewStatSum(minItems int, extraParams string) (StatMetric, error) {
	return &StatSum{Events: make(map[string]float64), MinItems: minItems, FieldName: extraParams}, nil
}

type StatSum struct {
	Sum       float64
	Events    map[string]float64 // map[EventTenantID]Cost
	MinItems  int
	FieldName string
	val       *float64 // cached sum value
}

// getValue returns tcd.val
func (sum *StatSum) getValue() float64 {
	if sum.val == nil {
		if len(sum.Events) == 0 || len(sum.Events) < sum.MinItems {
			sum.val = utils.Float64Pointer(STATS_NA)
		} else {
			sum.val = utils.Float64Pointer(utils.Round(sum.Sum,
				config.CgrConfig().GeneralCfg().RoundingDecimals,
				utils.ROUNDING_MIDDLE))
		}
	}
	return *sum.val
}

func (sum *StatSum) GetStringValue(fmtOpts string) (valStr string) {
	if val := sum.getValue(); val == STATS_NA {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = strconv.FormatFloat(sum.getValue(), 'f', -1, 64)
	}
	return
}

func (sum *StatSum) GetValue() (v interface{}) {
	return sum.getValue()
}

func (sum *StatSum) GetFloat64Value() (v float64) {
	return sum.getValue()
}

func (sum *StatSum) AddEvent(ev *utils.CGREvent) (err error) {
	var value float64
	if val, err := ev.FieldAsFloat64(sum.FieldName); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if val >= 0 {
		value = val
		sum.Sum += val
	}
	sum.Events[ev.ID] = value
	sum.val = nil
	return
}

func (sum *StatSum) RemEvent(evID string) (err error) {
	val, has := sum.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if val != 0 {
		sum.Sum -= val
	}
	delete(sum.Events, evID)
	sum.val = nil
	return
}

func (sum *StatSum) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(sum)
}

func (sum *StatSum) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, sum)
}

func NewStatAverage(minItems int, extraParams string) (StatMetric, error) {
	return &StatAverage{Events: make(map[string]float64), MinItems: minItems, FieldName: extraParams}, nil
}

// StatAverage implements TotalCallCost metric
type StatAverage struct {
	Sum       float64
	Count     float64
	Events    map[string]float64 // map[EventTenantID]Cost
	MinItems  int
	FieldName string
	val       *float64 // cached avg value
}

// getValue returns tcd.val
func (avg *StatAverage) getValue() float64 {
	if avg.val == nil {
		if (avg.MinItems > 0 && len(avg.Events) < avg.MinItems) || (avg.Count == 0) {
			avg.val = utils.Float64Pointer(STATS_NA)
		} else {
			avg.val = utils.Float64Pointer(utils.Round((avg.Sum / avg.Count),
				config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE))
		}
	}
	return *avg.val
}

func (avg *StatAverage) GetStringValue(fmtOpts string) (valStr string) {
	if val := avg.getValue(); val == STATS_NA {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = strconv.FormatFloat(avg.getValue(), 'f', -1, 64)
	}
	return

}

func (avg *StatAverage) GetValue() (v interface{}) {
	return avg.getValue()
}

func (avg *StatAverage) GetFloat64Value() (v float64) {
	return avg.getValue()
}

func (avg *StatAverage) AddEvent(ev *utils.CGREvent) (err error) {
	var value float64
	if val, err := ev.FieldAsFloat64(avg.FieldName); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if val > 0 {
		value = val
		avg.Sum += val
		avg.Events[ev.ID] = value
		avg.Count += 1
		avg.val = nil
	}
	return
}

func (avg *StatAverage) RemEvent(evID string) (err error) {
	val, has := avg.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if avg.Events[avg.FieldName] >= 0 {
		avg.Sum -= val
	}
	avg.Count -= 1
	delete(avg.Events, evID)
	avg.val = nil
	return
}

func (avg *StatAverage) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(avg)
}

func (avg *StatAverage) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, avg)
}
