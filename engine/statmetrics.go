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
)

// NewStatMetric instantiates the StatMetric
// cfg serves as general purpose container to pass config options to metric
func NewStatMetric(metricID string) (sm StatMetric, err error) {
	metrics := map[string]func() (StatMetric, error){
		utils.MetaASR: NewASR,
		utils.MetaACD: NewACD,
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
	return &StatACD{Events: make(map[string]float64)}, nil
}

// ACD implements AverageCallDuration metric
type StatACD struct {
	Sum    float64
	Count  float64
	Events map[string]float64 // map[EventTenantID]Duration
	val    *float64           // cached ACD value
}

// getValue returns asr.val
func (acd *StatACD) getValue() float64 {
	if acd.val == nil {
		if acd.Count == 0 {
			acd.val = utils.Float64Pointer(float64(STATS_NA))
		} else {
			acd.val = utils.Float64Pointer(utils.Round(acd.Sum/acd.Count,
				config.CgrConfig().RoundingDecimals, utils.ROUNDING_MIDDLE))
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
	return acd.getValue()
}

func (acd *StatACD) AddEvent(ev *StatEvent) (err error) {
	var answered float64
	if at, err := ev.AnswerTime(config.CgrConfig().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		duration, _ := ev.Usage(config.CgrConfig().DefaultTimezone)
		answered = duration.Seconds()
		acd.Sum += duration.Seconds()
	}
	acd.Events[ev.TenantID()] = answered
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
