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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"strconv"
	"time"
)

// NewStatMetric instantiates the StatMetric
// cfg serves as general purpose container to pass config options to metric
func NewStatMetric(metricID string) (sm StatMetric, err error) {
	metrics := map[string]func() (StatMetric, error){
		utils.MetaASR: NewASR,
		utils.MetaACD: NewACD,
		utils.MetaTCD: NewTCD,
		utils.MetaACC: NewACC,
		utils.MetaTCC: NewTCC,
		utils.MetaPDD: NewPDD,
		utils.MetaDDC: NewDCC,
	}
	if _, has := metrics[metricID]; !has {
		return nil, fmt.Errorf("unsupported metric: %s", metricID)
	}
	return metrics[metricID]()
}

// StatMetric is the interface which a metric should implement
type StatMetric interface {
	GetValue() interface{}
	GetStringValue(fmtOpts string) (val string)
	GetFloat64Value() (val float64)
	AddEvent(ev *StatEvent) error
	RemEvent(evTenantID string) error
	Marshal(ms Marshaler) (marshaled []byte, err error)
	LoadMarshaled(ms Marshaler, marshaled []byte) (err error)
}

func NewASR() (StatMetric, error) {
	return &StatASR{Events: make(map[string]bool)}, nil
}

// ASR implements AverageSuccessRatio metric
type StatASR struct {
	Answered float64
	Count    float64
	Events   map[string]bool // map[EventTenantID]Answered
	val      *float64        // cached ASR value
}

// getValue returns asr.val
func (asr *StatASR) getValue() float64 {
	if asr.val == nil {
		if asr.Count == 0 {
			asr.val = utils.Float64Pointer(float64(STATS_NA))
		} else {
			asr.val = utils.Float64Pointer(utils.Round((asr.Answered / asr.Count * 100),
				config.CgrConfig().RoundingDecimals, utils.ROUNDING_MIDDLE))
		}
	}
	return *asr.val
}

// GetValue returns the ASR value as part of StatMetric interface
func (asr *StatASR) GetValue() (v interface{}) {
	return asr.getValue()
}

func (asr *StatASR) GetStringValue(fmtOpts string) (valStr string) {
	if asr.Count == 0 {
		return utils.NOT_AVAILABLE
	}
	return fmt.Sprintf("%v%%", asr.getValue()) // %v will automatically limit the number of decimals printed
}

// GetFloat64Value is part of StatMetric interface
func (asr *StatASR) GetFloat64Value() (val float64) {
	return asr.getValue()
}

// AddEvent is part of StatMetric interface
func (asr *StatASR) AddEvent(ev *StatEvent) (err error) {
	var answered bool
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		answered = true
	}
	asr.Events[ev.TenantID()] = answered
	asr.Count += 1
	if answered {
		asr.Answered += 1
	}
	asr.val = nil
	return
}

func (asr *StatASR) RemEvent(evTenantID string) (err error) {
	answered, has := asr.Events[evTenantID]
	if !has {
		return utils.ErrNotFound
	}
	if answered {
		asr.Answered -= 1
	}
	asr.Count -= 1
	delete(asr.Events, evTenantID)
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

func NewACD() (StatMetric, error) {
	return &StatACD{Events: make(map[string]time.Duration)}, nil
}

// ACD implements AverageCallDuration metric
type StatACD struct {
	Sum    time.Duration
	Count  int64
	Events map[string]time.Duration // map[EventTenantID]Duration
	val    *time.Duration           // cached ACD value
}

// getValue returns acr.val
func (acd *StatACD) getValue() time.Duration {
	if acd.val == nil {
		if acd.Count == 0 {
			acd.val = utils.DurationPointer(time.Duration((-1) * time.Nanosecond))
		} else {
			acd.val = utils.DurationPointer(time.Duration(acd.Sum.Nanoseconds() / acd.Count))
		}
	}
	return *acd.val
}

func (acd *StatACD) GetStringValue(fmtOpts string) (val string) {
	if acd.Count == 0 {
		return utils.NOT_AVAILABLE
	}
	return fmt.Sprintf("%+v", acd.getValue())
}

func (acd *StatACD) GetValue() (v interface{}) {
	return acd.getValue()
}

func (acd *StatACD) GetFloat64Value() (v float64) {
	if acd.Count == 0 {
		return -1.0
	}
	return acd.getValue().Seconds()
}

func (acd *StatACD) AddEvent(ev *StatEvent) (err error) {
	var value time.Duration
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if duration, err := ev.Usage(config.CgrConfig().DefaultTimezone); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else {
			value = duration
			acd.Sum += duration
		}
	}
	acd.Events[ev.TenantID()] = value
	acd.Count += 1
	acd.val = nil
	return
}

func (acd *StatACD) RemEvent(evTenantID string) (err error) {
	duration, has := acd.Events[evTenantID]
	if !has {
		return utils.ErrNotFound
	}
	if duration != 0 {
		acd.Sum -= duration
	}
	acd.Count -= 1
	delete(acd.Events, evTenantID)
	acd.val = nil
	return
}

func (acd *StatACD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(acd)
}
func (acd *StatACD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, acd)
}

func NewTCD() (StatMetric, error) {
	return &StatTCD{Events: make(map[string]time.Duration)}, nil
}

// TCD implements TotalCallDuration metric
type StatTCD struct {
	Sum    time.Duration
	Count  int64
	Events map[string]time.Duration // map[EventTenantID]Duration
	val    *time.Duration           // cached TCD value
}

// getValue returns tcd.val
func (tcd *StatTCD) getValue() time.Duration {
	if tcd.val == nil {
		if tcd.Count == 0 {
			tcd.val = utils.DurationPointer(time.Duration((-1) * time.Nanosecond))
		} else {
			tcd.val = utils.DurationPointer(time.Duration(tcd.Sum.Nanoseconds()))
		}
	}
	return *tcd.val
}

func (tcd *StatTCD) GetStringValue(fmtOpts string) (val string) {
	if tcd.Count == 0 {
		return utils.NOT_AVAILABLE
	}
	return fmt.Sprintf("%+v", tcd.getValue())
}

func (tcd *StatTCD) GetValue() (v interface{}) {
	return tcd.getValue()
}

func (tcd *StatTCD) GetFloat64Value() (v float64) {
	if tcd.Count == 0 {
		return -1.0
	}
	return tcd.getValue().Seconds()
}

func (tcd *StatTCD) AddEvent(ev *StatEvent) (err error) {
	var value time.Duration
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if duration, err := ev.Usage(config.CgrConfig().DefaultTimezone); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else {
			value = duration
			tcd.Sum += duration
		}

	}
	tcd.Events[ev.TenantID()] = value
	tcd.Count += 1
	tcd.val = nil
	return
}

func (tcd *StatTCD) RemEvent(evTenantID string) (err error) {
	duration, has := tcd.Events[evTenantID]
	if !has {
		return utils.ErrNotFound
	}
	if duration != 0 {
		tcd.Sum -= duration
	}
	tcd.Count -= 1
	delete(tcd.Events, evTenantID)
	tcd.val = nil
	return
}

func (tcd *StatTCD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(tcd)
}

func (tcd *StatTCD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, tcd)
}

func NewACC() (StatMetric, error) {
	return &StatACC{Events: make(map[string]float64)}, nil
}

// ACC implements AverageCallCost metric
type StatACC struct {
	Sum    float64
	Count  float64
	Events map[string]float64 // map[EventTenantID]Cost
	val    *float64           // cached ACC value
}

// getValue returns tcd.val
func (acc *StatACC) getValue() float64 {
	if acc.val == nil {
		if acc.Count == 0 {
			acc.val = utils.Float64Pointer(float64(STATS_NA))
		} else {
			acc.val = utils.Float64Pointer(utils.Round((acc.Sum / acc.Count),
				config.CgrConfig().RoundingDecimals, utils.ROUNDING_MIDDLE))
		}
	}
	return *acc.val
}

func (acc *StatACC) GetStringValue(fmtOpts string) (val string) {
	if acc.Count == 0 {
		return utils.NOT_AVAILABLE
	}
	return strconv.FormatFloat(acc.getValue(), 'f', -1, 64)

}

func (acc *StatACC) GetValue() (v interface{}) {
	return acc.getValue()
}

func (acc *StatACC) GetFloat64Value() (v float64) {
	return acc.getValue()
}

func (acc *StatACC) AddEvent(ev *StatEvent) (err error) {
	var value float64
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if cost, err := ev.Cost(config.CgrConfig().DefaultTimezone); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else if cost >= 0 {
			value = cost
			acc.Sum += cost
		}
	}
	acc.Events[ev.TenantID()] = value
	acc.Count += 1
	acc.val = nil
	return
}

func (acc *StatACC) RemEvent(evTenantID string) (err error) {
	cost, has := acc.Events[evTenantID]
	if !has {
		return utils.ErrNotFound
	}
	if cost >= 0 {
		acc.Sum -= cost
	}
	acc.Count -= 1
	delete(acc.Events, evTenantID)
	acc.val = nil
	return
}

func (acc *StatACC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(acc)
}

func (acc *StatACC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, acc)
}

func NewTCC() (StatMetric, error) {
	return &StatTCC{Events: make(map[string]float64)}, nil
}

// TCC implements TotalCallCost metric
type StatTCC struct {
	Sum    float64
	Count  float64
	Events map[string]float64 // map[EventTenantID]Cost
	val    *float64           // cached TCC value
}

// getValue returns tcd.val
func (tcc *StatTCC) getValue() float64 {
	if tcc.val == nil {
		if tcc.Count == 0 {
			tcc.val = utils.Float64Pointer(float64(STATS_NA))
		} else {
			tcc.val = utils.Float64Pointer(utils.Round(tcc.Sum,
				config.CgrConfig().RoundingDecimals, utils.ROUNDING_MIDDLE))
		}
	}
	return *tcc.val
}

func (tcc *StatTCC) GetStringValue(fmtOpts string) (val string) {
	if tcc.Count == 0 {
		return utils.NOT_AVAILABLE
	}
	return strconv.FormatFloat(tcc.getValue(), 'f', -1, 64)
}

func (tcc *StatTCC) GetValue() (v interface{}) {
	return tcc.getValue()
}

func (tcc *StatTCC) GetFloat64Value() (v float64) {
	return tcc.getValue()
}

func (tcc *StatTCC) AddEvent(ev *StatEvent) (err error) {
	var value float64
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		if cost, err := ev.Cost(config.CgrConfig().DefaultTimezone); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else if cost >= 0 {
			value = cost
			tcc.Sum += cost
		}
	}
	tcc.Events[ev.TenantID()] = value
	tcc.Count += 1
	tcc.val = nil
	return
}

func (tcc *StatTCC) RemEvent(evTenantID string) (err error) {
	cost, has := tcc.Events[evTenantID]
	if !has {
		return utils.ErrNotFound
	}
	if cost != 0 {
		tcc.Sum -= cost
	}
	tcc.Count -= 1
	delete(tcc.Events, evTenantID)
	tcc.val = nil
	return
}

func (tcc *StatTCC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(tcc)
}

func (tcc *StatTCC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, tcc)
}

func NewPDD() (StatMetric, error) {
	return &StatPDD{Events: make(map[string]time.Duration)}, nil
}

// PDD implements Post Dial Delay (average) metric
type StatPDD struct {
	Sum    time.Duration
	Count  int64
	Events map[string]time.Duration // map[EventTenantID]Duration
	val    *time.Duration           // cached PDD value
}

// getValue returns pdd.val
func (pdd *StatPDD) getValue() time.Duration {
	if pdd.val == nil {
		if pdd.Count == 0 {
			pdd.val = utils.DurationPointer(time.Duration((-1) * time.Nanosecond))
		} else {
			pdd.val = utils.DurationPointer(time.Duration(pdd.Sum.Nanoseconds() / pdd.Count))
		}
	}
	return *pdd.val
}

func (pdd *StatPDD) GetStringValue(fmtOpts string) (val string) {
	if pdd.Count == 0 {
		return utils.NOT_AVAILABLE
	}
	return fmt.Sprintf("%+v", pdd.getValue())
}

func (pdd *StatPDD) GetValue() (v interface{}) {
	return pdd.getValue()
}

func (pdd *StatPDD) GetFloat64Value() (v float64) {
	if pdd.Count == 0 {
		return -1.0
	}
	return pdd.getValue().Seconds()
}

func (pdd *StatPDD) AddEvent(ev *StatEvent) (err error) {
	var value time.Duration
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		if duration, err := ev.Pdd(config.CgrConfig().DefaultTimezone); err != nil &&
			err != utils.ErrNotFound {
			return err
		} else {
			value = duration
			pdd.Sum += duration
		}
	}
	pdd.Events[ev.TenantID()] = value
	pdd.Count += 1
	pdd.val = nil
	return
}

func (pdd *StatPDD) RemEvent(evTenantID string) (err error) {
	duration, has := pdd.Events[evTenantID]
	if !has {
		return utils.ErrNotFound
	}
	if duration != 0 {
		pdd.Sum -= duration
	}
	pdd.Count -= 1
	delete(pdd.Events, evTenantID)
	pdd.val = nil
	return
}

func (pdd *StatPDD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(pdd)
}
func (pdd *StatPDD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, pdd)
}

func NewDCC() (StatMetric, error) {
	return &StatDDC{Destinations: make(map[string]utils.StringMap), EventDestinations: make(map[string]string)}, nil
}

type StatDDC struct {
	Destinations      map[string]utils.StringMap
	EventDestinations map[string]string // map[EventTenantID]Destination
}

func (ddc *StatDDC) GetStringValue(fmtOpts string) (val string) {
	if len(ddc.Destinations) == 0 {
		return utils.NOT_AVAILABLE
	}
	return fmt.Sprintf("%+v", len(ddc.Destinations))
}

func (ddc *StatDDC) GetValue() (v interface{}) {
	return len(ddc.Destinations)
}

func (ddc *StatDDC) GetFloat64Value() (v float64) {
	if len(ddc.Destinations) == 0 {
		return -1.0
	}
	return float64(len(ddc.Destinations))
}

func (ddc *StatDDC) AddEvent(ev *StatEvent) (err error) {
	var dest string
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		if destination, err := ev.Destination(config.CgrConfig().DefaultTimezone); err != nil {
			return err
		} else {
			dest = destination
			if _, has := ddc.Destinations[dest]; !has {
				ddc.Destinations[dest] = make(map[string]bool)
			}
			ddc.Destinations[dest][ev.TenantID()] = true
		}
	}
	ddc.EventDestinations[ev.TenantID()] = dest
	return
}

func (ddc *StatDDC) RemEvent(evTenantID string) (err error) {
	destination, has := ddc.EventDestinations[evTenantID]
	if !has {
		return utils.ErrNotFound
	}
	if len(ddc.Destinations[destination]) == 1 {
		delete(ddc.Destinations, destination)
	} else {
		delete(ddc.Destinations[destination], evTenantID)
	}

	delete(ddc.EventDestinations, evTenantID)
	return
}

func (ddc *StatDDC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(DDC)
}
func (ddc *StatDDC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, ddc)
}
