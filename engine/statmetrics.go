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

// ACDHelper structure
type DurationWithCompress struct {
	Duration       time.Duration
	CompressFactor int
}

// ACDHelper structure
type StatWithCompress struct {
	Stat           float64
	CompressFactor int
}

// NewStatMetric instantiates the StatMetric
// cfg serves as general purpose container to pass config options to metric
func NewStatMetric(metricID string, minItems int, filterIDs []string) (sm StatMetric, err error) {
	metrics := map[string]func(int, string, []string) (StatMetric, error){
		utils.MetaASR:      NewASR,
		utils.MetaACD:      NewACD,
		utils.MetaTCD:      NewTCD,
		utils.MetaACC:      NewACC,
		utils.MetaTCC:      NewTCC,
		utils.MetaPDD:      NewPDD,
		utils.MetaDDC:      NewDDC,
		utils.MetaSum:      NewStatSum,
		utils.MetaAverage:  NewStatAverage,
		utils.MetaDistinct: NewStatDistinct,
	}
	// split the metricID
	// in case of *sum we have *sum:~FieldName
	metricSplit := utils.SplitConcatenatedKey(metricID)
	if _, has := metrics[metricSplit[0]]; !has {
		return nil, fmt.Errorf("unsupported metric type <%s>", metricSplit[0])
	}
	var extraParams string
	if len(metricSplit[1:]) > 0 {
		extraParams = metricSplit[1]
	}
	return metrics[metricSplit[0]](minItems, extraParams, filterIDs)
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
	GetFilterIDs() (filterIDs []string)
	Compress(queueLen int64, defaultID string) (eventIDs []string)
	GetCompressFactor(events map[string]int) map[string]int
}

func NewASR(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatASR{Events: make(map[string]*StatWithCompress),
		MinItems: minItems, FilterIDs: filterIDs}, nil
}

// ASR implements AverageSuccessRatio metric
type StatASR struct {
	FilterIDs []string
	Answered  float64
	Count     int64
	Events    map[string]*StatWithCompress // map[EventTenantID]Answered
	MinItems  int
	val       *float64 // cached ASR value
}

// getValue returns asr.val
func (asr *StatASR) getValue() float64 {
	if asr.val == nil {
		if (asr.MinItems > 0 && asr.Count < int64(asr.MinItems)) || (asr.Count == 0) {
			asr.val = utils.Float64Pointer(STATS_NA)
		} else {
			asr.val = utils.Float64Pointer(utils.Round((asr.Answered / float64(asr.Count) * 100.0),
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
	var answered int
	if at, err := ev.FieldAsTime(utils.AnswerTime,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil &&
		err != utils.ErrNotFound {
		return err
	} else if !at.IsZero() {
		answered = 1
	}

	if val, has := asr.Events[ev.ID]; !has {
		asr.Events[ev.ID] = &StatWithCompress{Stat: float64(answered), CompressFactor: 1}
	} else {
		val.Stat = (val.Stat*float64(val.CompressFactor) + float64(answered)) / float64(val.CompressFactor+1)
		val.CompressFactor = val.CompressFactor + 1
	}
	asr.Count += 1
	if answered == 1 {
		asr.Answered += 1
	}
	asr.val = nil
	return
}

func (asr *StatASR) RemEvent(evID string) (err error) {
	val, has := asr.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	ans := 0
	if val.Stat > 0.5 {
		ans = 1
		asr.Answered -= 1
	}
	asr.Count -= 1
	if val.CompressFactor <= 1 {
		delete(asr.Events, evID)
	} else {
		val.Stat = (val.Stat*float64(val.CompressFactor) - float64(ans)) / (float64(val.CompressFactor - 1))
		val.CompressFactor = val.CompressFactor - 1
	}
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

// GetFilterIDs is part of StatMetric interface
func (asr *StatASR) GetFilterIDs() []string {
	return asr.FilterIDs
}

// Compress is part of StatMetric interface
func (asr *StatASR) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if asr.Count < queueLen {
		for id, _ := range asr.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round(asr.Answered/float64(asr.Count),
			config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE),
		CompressFactor: int(asr.Count),
	}
	asr.Events = map[string]*StatWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (asr *StatASR) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range asr.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewACD(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatACD{Events: make(map[string]*DurationWithCompress), MinItems: minItems, FilterIDs: filterIDs}, nil
}

// ACD implements AverageCallDuration metric
type StatACD struct {
	FilterIDs []string
	Sum       time.Duration
	Count     int64
	Events    map[string]*DurationWithCompress // map[EventTenantID]Duration
	MinItems  int
	val       *time.Duration // cached ACD value
}

// getValue returns acr.val
func (acd *StatACD) getValue() time.Duration {
	if acd.val == nil {
		if (acd.MinItems > 0 && acd.Count < int64(acd.MinItems)) || (acd.Count == 0) {
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
	var dur time.Duration
	if dur, err = ev.FieldAsDuration(utils.Usage); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Usage)
		}
		return
	}
	acd.Sum += dur
	if val, has := acd.Events[ev.ID]; !has {
		acd.Events[ev.ID] = &DurationWithCompress{Duration: dur, CompressFactor: 1}
	} else {
		val.Duration = time.Duration((float64(val.Duration.Nanoseconds())*float64(val.CompressFactor) + float64(dur.Nanoseconds())) / float64(val.CompressFactor+1))
		val.CompressFactor = val.CompressFactor + 1
	}
	acd.Count += 1
	acd.val = nil
	return
}

func (acd *StatACD) RemEvent(evID string) (err error) {
	val, has := acd.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if val.Duration != 0 {
		acd.Sum -= val.Duration
	}
	acd.Count -= 1
	if val.CompressFactor <= 1 {
		delete(acd.Events, evID)
	} else {
		val.CompressFactor = val.CompressFactor - 1
	}
	acd.val = nil
	return
}

func (acd *StatACD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(acd)
}
func (acd *StatACD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, acd)
}

// GetFilterIDs is part of StatMetric interface
func (acd *StatACD) GetFilterIDs() []string {
	return acd.FilterIDs
}

// Compress is part of StatMetric interface
func (acd *StatACD) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if acd.Count < queueLen {
		for id, _ := range acd.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &DurationWithCompress{
		Duration:       time.Duration(acd.Sum.Nanoseconds() / acd.Count),
		CompressFactor: int(acd.Count),
	}
	acd.Events = map[string]*DurationWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (acd *StatACD) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range acd.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewTCD(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatTCD{Events: make(map[string]*DurationWithCompress), MinItems: minItems, FilterIDs: filterIDs}, nil
}

// TCD implements TotalCallDuration metric
type StatTCD struct {
	FilterIDs []string
	Sum       time.Duration
	Count     int64
	Events    map[string]*DurationWithCompress // map[EventTenantID]Duration
	MinItems  int
	val       *time.Duration // cached TCD value
}

// getValue returns tcd.val
func (tcd *StatTCD) getValue() time.Duration {
	if tcd.val == nil {
		if (tcd.MinItems > 0 && tcd.Count < int64(tcd.MinItems)) || (tcd.Count == 0) {
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
	var dur time.Duration
	if dur, err = ev.FieldAsDuration(utils.Usage); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Usage)
		}
		return
	}
	tcd.Sum += dur
	if val, has := tcd.Events[ev.ID]; !has {
		tcd.Events[ev.ID] = &DurationWithCompress{Duration: dur, CompressFactor: 1}
	} else {
		val.Duration = time.Duration((float64(val.Duration.Nanoseconds())*float64(val.CompressFactor) + float64(dur.Nanoseconds())) / float64(val.CompressFactor+1))
		val.CompressFactor = val.CompressFactor + 1
	}
	tcd.Count += 1
	tcd.val = nil
	return
}

func (tcd *StatTCD) RemEvent(evID string) (err error) {
	val, has := tcd.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if val.Duration != 0 {
		tcd.Sum -= val.Duration
	}
	tcd.Count -= 1
	if val.CompressFactor <= 1 {
		delete(tcd.Events, evID)
	} else {
		val.CompressFactor = val.CompressFactor - 1
	}
	tcd.val = nil
	return
}

func (tcd *StatTCD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(tcd)
}

func (tcd *StatTCD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, tcd)
}

// GetFilterIDs is part of StatMetric interface
func (tcd *StatTCD) GetFilterIDs() []string {
	return tcd.FilterIDs
}

// Compress is part of StatMetric interface
func (tcd *StatTCD) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if tcd.Count < queueLen {
		for id, _ := range tcd.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &DurationWithCompress{
		Duration:       time.Duration(tcd.Sum.Nanoseconds() / tcd.Count),
		CompressFactor: int(tcd.Count),
	}
	tcd.Events = map[string]*DurationWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (tcd *StatTCD) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range tcd.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewACC(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatACC{Events: make(map[string]*StatWithCompress), MinItems: minItems, FilterIDs: filterIDs}, nil
}

// ACC implements AverageCallCost metric
type StatACC struct {
	FilterIDs []string
	Sum       float64
	Count     int64
	Events    map[string]*StatWithCompress // map[EventTenantID]Cost
	MinItems  int
	val       *float64 // cached ACC value
}

// getValue returns tcd.val
func (acc *StatACC) getValue() float64 {
	if acc.val == nil {
		if (acc.MinItems > 0 && acc.Count < int64(acc.MinItems)) || (acc.Count == 0) {
			acc.val = utils.Float64Pointer(STATS_NA)
		} else {
			acc.val = utils.Float64Pointer(utils.Round((acc.Sum / float64(acc.Count)),
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
	var cost float64
	if cost, err = ev.FieldAsFloat64(utils.COST); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.COST)
		}
		return
	}
	acc.Sum += cost
	if val, has := acc.Events[ev.ID]; !has {
		acc.Events[ev.ID] = &StatWithCompress{Stat: cost, CompressFactor: 1}
	} else {
		val.Stat = (val.Stat*float64(val.CompressFactor) + cost) / float64(val.CompressFactor+1)
		val.CompressFactor = val.CompressFactor + 1
	}
	acc.Count += 1
	acc.val = nil
	return
}

func (acc *StatACC) RemEvent(evID string) (err error) {
	cost, has := acc.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	acc.Sum -= cost.Stat
	acc.Count -= 1
	if cost.CompressFactor <= 1 {
		delete(acc.Events, evID)
	} else {
		cost.CompressFactor = cost.CompressFactor - 1
	}
	acc.val = nil
	return
}

func (acc *StatACC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(acc)
}

func (acc *StatACC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, acc)
}

// GetFilterIDs is part of StatMetric interface
func (acc *StatACC) GetFilterIDs() []string {
	return acc.FilterIDs
}

// Compress is part of StatMetric interface
func (acc *StatACC) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if acc.Count < queueLen {
		for id, _ := range acc.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round((acc.Sum / float64(acc.Count)),
			config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE),
		CompressFactor: int(acc.Count),
	}
	acc.Events = map[string]*StatWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (acc *StatACC) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range acc.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewTCC(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatTCC{Events: make(map[string]*StatWithCompress), MinItems: minItems, FilterIDs: filterIDs}, nil
}

// TCC implements TotalCallCost metric
type StatTCC struct {
	FilterIDs []string
	Sum       float64
	Count     int64
	Events    map[string]*StatWithCompress // map[EventTenantID]Cost
	MinItems  int
	val       *float64 // cached TCC value
}

// getValue returns tcd.val
func (tcc *StatTCC) getValue() float64 {
	if tcc.val == nil {
		if (tcc.MinItems > 0 && tcc.Count < int64(tcc.MinItems)) || (tcc.Count == 0) {
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
	var cost float64
	if cost, err = ev.FieldAsFloat64(utils.COST); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.COST)
		}
		return
	}
	tcc.Sum += cost
	if val, has := tcc.Events[ev.ID]; !has {
		tcc.Events[ev.ID] = &StatWithCompress{Stat: cost, CompressFactor: 1}
	} else {
		val.Stat = (val.Stat*float64(val.CompressFactor) + cost) / float64(val.CompressFactor+1)
		val.CompressFactor = val.CompressFactor + 1
	}
	tcc.Count += 1
	tcc.val = nil
	return
}

func (tcc *StatTCC) RemEvent(evID string) (err error) {
	cost, has := tcc.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if cost.Stat != 0 {
		tcc.Sum -= cost.Stat
	}
	tcc.Count -= 1
	if cost.CompressFactor <= 1 {
		delete(tcc.Events, evID)
	} else {
		cost.CompressFactor = cost.CompressFactor - 1
	}
	tcc.val = nil
	return
}

func (tcc *StatTCC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(tcc)
}

func (tcc *StatTCC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, tcc)
}

// GetFilterIDs is part of StatMetric interface
func (tcc *StatTCC) GetFilterIDs() []string {
	return tcc.FilterIDs
}

// Compress is part of StatMetric interface
func (tcc *StatTCC) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if tcc.Count < queueLen {
		for id, _ := range tcc.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round((tcc.Sum / float64(tcc.Count)),
			config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE),
		CompressFactor: int(tcc.Count),
	}
	tcc.Events = map[string]*StatWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (tcc *StatTCC) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range tcc.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewPDD(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatPDD{Events: make(map[string]*DurationWithCompress), MinItems: minItems, FilterIDs: filterIDs}, nil
}

// PDD implements Post Dial Delay (average) metric
type StatPDD struct {
	FilterIDs []string
	Sum       time.Duration
	Count     int64
	Events    map[string]*DurationWithCompress // map[EventTenantID]Duration
	MinItems  int
	val       *time.Duration // cached PDD value
}

// getValue returns pdd.val
func (pdd *StatPDD) getValue() time.Duration {
	if pdd.val == nil {
		if (pdd.MinItems > 0 && pdd.Count < int64(pdd.MinItems)) || (pdd.Count == 0) {
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
	var dur time.Duration
	if dur, err = ev.FieldAsDuration(utils.PDD); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.PDD)
		}
		return
	}
	pdd.Sum += dur
	if val, has := pdd.Events[ev.ID]; !has {
		pdd.Events[ev.ID] = &DurationWithCompress{Duration: dur, CompressFactor: 1}
	} else {
		val.Duration = time.Duration((float64(val.Duration.Nanoseconds())*float64(val.CompressFactor) + float64(dur.Nanoseconds())) / float64(val.CompressFactor+1))
		val.CompressFactor = val.CompressFactor + 1
	}
	pdd.Count += 1
	pdd.val = nil
	return
}

func (pdd *StatPDD) RemEvent(evID string) (err error) {
	val, has := pdd.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if val.Duration != 0 {
		pdd.Sum -= val.Duration
	}
	pdd.Count -= 1
	if val.CompressFactor <= 1 {
		delete(pdd.Events, evID)
	} else {
		val.CompressFactor = val.CompressFactor - 1
	}
	pdd.val = nil
	return
}

func (pdd *StatPDD) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(pdd)
}
func (pdd *StatPDD) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, pdd)
}

// GetFilterIDs is part of StatMetric interface
func (pdd *StatPDD) GetFilterIDs() []string {
	return pdd.FilterIDs
}

// Compress is part of StatMetric interface
func (pdd *StatPDD) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if pdd.Count < queueLen {
		for id, _ := range pdd.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &DurationWithCompress{
		Duration:       time.Duration(pdd.Sum.Nanoseconds() / pdd.Count),
		CompressFactor: int(pdd.Count),
	}
	pdd.Events = map[string]*DurationWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (pdd *StatPDD) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range pdd.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewDDC(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatDDC{Events: make(map[string]map[string]int64), FieldValues: make(map[string]map[string]struct{}),
		MinItems: minItems, FilterIDs: filterIDs}, nil
}

type StatDDC struct {
	FilterIDs   []string
	FieldValues map[string]map[string]struct{} // map[fieldValue]map[eventID]
	Events      map[string]map[string]int64    // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    int
	Count       int64
}

// getValue returns tcd.val
func (ddc *StatDDC) getValue() float64 {
	if ddc.Count == 0 || ddc.Count < int64(ddc.MinItems) {
		return STATS_NA
	}
	return float64(len(ddc.FieldValues))
}

func (ddc *StatDDC) GetStringValue(fmtOpts string) (valStr string) {
	if val := ddc.getValue(); val == STATS_NA {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = strconv.FormatFloat(ddc.getValue(), 'f', -1, 64)
	}
	return
}

func (ddc *StatDDC) GetValue() (v interface{}) {
	return ddc.getValue()
}

func (ddc *StatDDC) GetFloat64Value() (v float64) {
	return ddc.getValue()
}

func (ddc *StatDDC) AddEvent(ev *utils.CGREvent) (err error) {
	var fieldValue string
	if fieldValue, err = ev.FieldAsString(utils.Destination); err != nil {
		return err
	}

	// add to fieldValues
	if _, has := ddc.FieldValues[fieldValue]; !has {
		ddc.FieldValues[fieldValue] = make(map[string]struct{})
	}
	ddc.FieldValues[fieldValue][ev.ID] = struct{}{}

	// add to events
	if _, has := ddc.Events[ev.ID]; !has {
		ddc.Events[ev.ID] = make(map[string]int64)
	}
	ddc.Count += 1
	if _, has := ddc.Events[ev.ID][fieldValue]; !has {
		ddc.Events[ev.ID][fieldValue] = 1
		return
	}
	ddc.Events[ev.ID][fieldValue] = ddc.Events[ev.ID][fieldValue] + 1
	return
}

func (ddc *StatDDC) RemEvent(evID string) (err error) {
	fieldValues, has := ddc.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if len(fieldValues) == 0 {
		delete(ddc.Events, evID)
		return utils.ErrNotFound
	}

	// decrement events
	var fieldValue string
	for k, _ := range fieldValues {
		fieldValue = k
		break
	}
	ddc.Count -= 1
	if fieldValues[fieldValue] > 1 {
		ddc.Events[evID][fieldValue] = ddc.Events[evID][fieldValue] - 1
		return // do not delete the reference until it reaches 0
	}
	delete(ddc.Events[evID], fieldValue)

	// remove from fieldValues
	if _, has := ddc.FieldValues[fieldValue]; !has {
		return
	}
	delete(ddc.FieldValues[fieldValue], evID)
	if len(ddc.FieldValues[fieldValue]) <= 0 {
		delete(ddc.FieldValues, fieldValue)
	}
	return
}

func (ddc *StatDDC) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(ddc)
}

func (ddc *StatDDC) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, ddc)
}

// GetFilterIDs is part of StatMetric interface
func (ddc *StatDDC) GetFilterIDs() []string {
	return ddc.FilterIDs
}

func (ddc *StatDDC) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	for id, _ := range ddc.Events {
		eventIDs = append(eventIDs, id)
	}
	return
}

// Compress is part of StatMetric interface
func (ddc *StatDDC) GetCompressFactor(events map[string]int) map[string]int {
	for id, ev := range ddc.Events {
		compressFactor := 0
		for _, fields := range ev {
			compressFactor += int(fields)
		}
		if _, has := events[id]; !has {
			events[id] = compressFactor
		}
		if events[id] < compressFactor {
			events[id] = compressFactor
		}
	}
	return events
}

func NewStatSum(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatSum{Events: make(map[string]*StatWithCompress),
		MinItems: minItems, FieldName: extraParams, FilterIDs: filterIDs}, nil
}

type StatSum struct {
	FilterIDs []string
	Sum       float64
	Count     int64
	Events    map[string]*StatWithCompress // map[EventTenantID]Cost
	MinItems  int
	FieldName string
	val       *float64 // cached sum value
}

// getValue returns tcd.val
func (sum *StatSum) getValue() float64 {
	if sum.val == nil {
		if len(sum.Events) == 0 || sum.Count < int64(sum.MinItems) {
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
	var val float64
	switch {
	case strings.HasPrefix(sum.FieldName, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep): // ~*req.
		//Remove the dynamic prefix and check in event for field
		field := sum.FieldName[6:]
		if val, err = ev.FieldAsFloat64(field); err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefix(err, field)
			}
			return
		}
	default:
		val, err = utils.IfaceAsFloat64(sum.FieldName)
		if err != nil {
			return
		}
	}
	sum.Sum += val
	if v, has := sum.Events[ev.ID]; !has {
		sum.Events[ev.ID] = &StatWithCompress{Stat: val, CompressFactor: 1}
	} else {
		v.Stat = (v.Stat*float64(v.CompressFactor) + val) / float64(v.CompressFactor+1)
		v.CompressFactor = v.CompressFactor + 1
	}
	sum.Count += 1
	sum.val = nil
	return
}

func (sum *StatSum) RemEvent(evID string) (err error) {
	val, has := sum.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if val.Stat != 0 {
		sum.Sum -= val.Stat
	}
	sum.Count -= 1
	if val.CompressFactor <= 1 {
		delete(sum.Events, evID)
	} else {
		val.CompressFactor = val.CompressFactor - 1
	}
	sum.val = nil
	return
}

func (sum *StatSum) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(sum)
}

func (sum *StatSum) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, sum)
}

// GetFilterIDs is part of StatMetric interface
func (sum *StatSum) GetFilterIDs() []string {
	return sum.FilterIDs
}

// Compress is part of StatMetric interface
func (sum *StatSum) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if sum.Count < queueLen {
		for id, _ := range sum.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round((sum.Sum / float64(sum.Count)),
			config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE),
		CompressFactor: int(sum.Count),
	}
	sum.Events = map[string]*StatWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (sum *StatSum) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range sum.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewStatAverage(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatAverage{Events: make(map[string]*StatWithCompress),
		MinItems: minItems, FieldName: extraParams, FilterIDs: filterIDs}, nil
}

// StatAverage implements TotalCallCost metric
type StatAverage struct {
	FilterIDs []string
	Sum       float64
	Count     int64
	Events    map[string]*StatWithCompress // map[EventTenantID]Cost
	MinItems  int
	FieldName string
	val       *float64 // cached avg value
}

// getValue returns tcd.val
func (avg *StatAverage) getValue() float64 {
	if avg.val == nil {
		if (avg.MinItems > 0 && avg.Count < int64(avg.MinItems)) || (avg.Count == 0) {
			avg.val = utils.Float64Pointer(STATS_NA)
		} else {
			avg.val = utils.Float64Pointer(utils.Round((avg.Sum / float64(avg.Count)),
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
	var val float64
	switch {
	case strings.HasPrefix(avg.FieldName, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep): // ~*req.
		//Remove the dynamic prefix and check in event for field
		field := avg.FieldName[6:]
		if val, err = ev.FieldAsFloat64(field); err != nil {
			if err == utils.ErrNotFound {
				err = utils.ErrPrefix(err, field)
			}
			return
		}
	default:
		val, err = utils.IfaceAsFloat64(avg.FieldName)
		if err != nil {
			return
		}
	}
	avg.Sum += val
	if v, has := avg.Events[ev.ID]; !has {
		avg.Events[ev.ID] = &StatWithCompress{Stat: val, CompressFactor: 1}
	} else {
		v.Stat = (v.Stat*float64(v.CompressFactor) + val) / float64(v.CompressFactor+1)
		v.CompressFactor = v.CompressFactor + 1
	}
	avg.Count += 1
	avg.val = nil
	return
}

func (avg *StatAverage) RemEvent(evID string) (err error) {
	val, has := avg.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if val.Stat >= 0 {
		avg.Sum -= val.Stat
	}
	avg.Count -= 1
	if val.CompressFactor <= 1 {
		delete(avg.Events, evID)
	} else {
		val.CompressFactor = val.CompressFactor - 1
	}
	avg.val = nil
	return
}

func (avg *StatAverage) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(avg)
}

func (avg *StatAverage) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, avg)
}

// GetFilterIDs is part of StatMetric interface
func (avg *StatAverage) GetFilterIDs() []string {
	return avg.FilterIDs
}

// Compress is part of StatMetric interface
func (avg *StatAverage) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	if avg.Count < queueLen {
		for id, _ := range avg.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round((avg.Sum / float64(avg.Count)),
			config.CgrConfig().GeneralCfg().RoundingDecimals, utils.ROUNDING_MIDDLE),
		CompressFactor: int(avg.Count),
	}
	avg.Events = map[string]*StatWithCompress{defaultID: stat}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (avg *StatAverage) GetCompressFactor(events map[string]int) map[string]int {
	for id, val := range avg.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func NewStatDistinct(minItems int, extraParams string, filterIDs []string) (StatMetric, error) {
	return &StatDistinct{Events: make(map[string]map[string]int64), FieldValues: make(map[string]map[string]struct{}),
		MinItems: minItems, FieldName: extraParams, FilterIDs: filterIDs}, nil
}

type StatDistinct struct {
	FilterIDs   []string
	FieldValues map[string]map[string]struct{} // map[fieldValue]map[eventID]
	Events      map[string]map[string]int64    // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    int
	FieldName   string
	Count       int64
}

// getValue returns tcd.val
func (dst *StatDistinct) getValue() float64 {
	if dst.Count == 0 || dst.Count < int64(dst.MinItems) {
		return STATS_NA
	}
	return float64(len(dst.FieldValues))
}

func (dst *StatDistinct) GetStringValue(fmtOpts string) (valStr string) {
	if val := dst.getValue(); val == STATS_NA {
		valStr = utils.NOT_AVAILABLE
	} else {
		valStr = strconv.FormatFloat(dst.getValue(), 'f', -1, 64)
	}
	return
}

func (dst *StatDistinct) GetValue() (v interface{}) {
	return dst.getValue()
}

func (dst *StatDistinct) GetFloat64Value() (v float64) {
	return dst.getValue()
}

func (dst *StatDistinct) AddEvent(ev *utils.CGREvent) (err error) {
	var fieldValue string
	// simply remove the ~*req. prefix and do normal process
	if !strings.HasPrefix(dst.FieldName, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep) {
		return fmt.Errorf("Invalid format for field <%s>", dst.FieldName)
	}
	field := dst.FieldName[6:]
	if fieldValue, err = ev.FieldAsString(field); err != nil {
		return err
	}

	// add to fieldValues
	if _, has := dst.FieldValues[fieldValue]; !has {
		dst.FieldValues[fieldValue] = make(map[string]struct{})
	}
	dst.FieldValues[fieldValue][ev.ID] = struct{}{}

	// add to events
	if _, has := dst.Events[ev.ID]; !has {
		dst.Events[ev.ID] = make(map[string]int64)
	}
	dst.Count += 1
	if _, has := dst.Events[ev.ID][fieldValue]; !has {
		dst.Events[ev.ID][fieldValue] = 1
		return
	}
	dst.Events[ev.ID][fieldValue] = dst.Events[ev.ID][fieldValue] + 1
	return
}

func (dst *StatDistinct) RemEvent(evID string) (err error) {
	fieldValues, has := dst.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if len(fieldValues) == 0 {
		delete(dst.Events, evID)
		return utils.ErrNotFound
	}

	// decrement events
	var fieldValue string
	for k, _ := range fieldValues {
		fieldValue = k
		break
	}
	dst.Count -= 1
	if fieldValues[fieldValue] > 1 {
		dst.Events[evID][fieldValue] = dst.Events[evID][fieldValue] - 1
		return // do not delete the reference until it reaches 0
	}
	delete(dst.Events[evID], fieldValue)

	// remove from fieldValues
	if _, has := dst.FieldValues[fieldValue]; !has {
		return
	}
	delete(dst.FieldValues[fieldValue], evID)
	if len(dst.FieldValues[fieldValue]) <= 0 {
		delete(dst.FieldValues, fieldValue)
	}
	return
}

func (dst *StatDistinct) Marshal(ms Marshaler) (marshaled []byte, err error) {
	return ms.Marshal(dst)
}

func (dst *StatDistinct) LoadMarshaled(ms Marshaler, marshaled []byte) (err error) {
	return ms.Unmarshal(marshaled, dst)
}

// GetFilterIDs is part of StatMetric interface
func (dst *StatDistinct) GetFilterIDs() []string {
	return dst.FilterIDs
}

func (dst *StatDistinct) Compress(queueLen int64, defaultID string) (eventIDs []string) {
	for id, _ := range dst.Events {
		eventIDs = append(eventIDs, id)
	}
	return
}

// Compress is part of StatMetric interface
func (dst *StatDistinct) GetCompressFactor(events map[string]int) map[string]int {
	for id, ev := range dst.Events {
		compressFactor := 0
		for _, fields := range ev {
			compressFactor += int(fields)
		}
		if _, has := events[id]; !has {
			events[id] = compressFactor
		}
		if events[id] < compressFactor {
			events[id] = compressFactor
		}
	}
	return events
}
