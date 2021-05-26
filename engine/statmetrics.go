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
	// in case of *sum we have *sum#~*req.FieldName
	metricSplit := strings.Split(metricID, utils.HashtagSep)
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
	GetValue(roundingDecimal int) interface{}
	GetStringValue(roundingDecimal int) (val string)
	GetFloat64Value(roundingDecimal int) (val float64)
	AddEvent(evID string, ev utils.DataProvider) error
	RemEvent(evTenantID string) error
	Marshal(ms Marshaler) (marshaled []byte, err error)
	LoadMarshaled(ms Marshaler, marshaled []byte) (err error)
	GetFilterIDs() (filterIDs []string)
	GetMinItems() (minIts int)
	Compress(queueLen int64, defaultID string, roundingDec int) (eventIDs []string)
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
func (asr *StatASR) getValue(roundingDecimal int) float64 {
	if asr.val == nil {
		if (asr.MinItems > 0 && asr.Count < int64(asr.MinItems)) || (asr.Count == 0) {
			asr.val = utils.Float64Pointer(utils.StatsNA)
		} else {
			asr.val = utils.Float64Pointer(utils.Round((asr.Answered / float64(asr.Count) * 100.0),
				roundingDecimal, utils.MetaRoundingMiddle))
		}
	}
	return *asr.val
}

// GetValue returns the ASR value as part of StatMetric interface
func (asr *StatASR) GetValue(roundingDecimal int) (v interface{}) {
	return asr.getValue(roundingDecimal)
}

func (asr *StatASR) GetStringValue(roundingDecimal int) (valStr string) {
	if val := asr.getValue(roundingDecimal); val == utils.StatsNA {
		valStr = utils.NotAvailable
	} else {
		valStr = fmt.Sprintf("%v%%", asr.getValue(roundingDecimal))
	}
	return
}

// GetFloat64Value is part of StatMetric interface
func (asr *StatASR) GetFloat64Value(roundingDecimal int) (val float64) {
	return asr.getValue(roundingDecimal)
}

// AddEvent is part of StatMetric interface
func (asr *StatASR) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var answered int
	if val, err := ev.FieldAsInterface([]string{utils.MetaReq, utils.AnswerTime}); err != nil {
		if err != utils.ErrNotFound {
			return err
		}
	} else if at, err := utils.IfaceAsTime(val,
		config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		answered = 1
	}

	if val, has := asr.Events[evID]; !has {
		asr.Events[evID] = &StatWithCompress{Stat: float64(answered), CompressFactor: 1}
	} else {
		val.Stat = (val.Stat*float64(val.CompressFactor) + float64(answered)) / float64(val.CompressFactor+1)
		val.CompressFactor = val.CompressFactor + 1
	}
	asr.Count++
	if answered == 1 {
		asr.Answered++
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
		asr.Answered--
	}
	asr.Count--
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

// GetMinItems returns the minim items for the metric
func (asr *StatASR) GetMinItems() (minIts int) { return asr.MinItems }

// Compress is part of StatMetric interface
func (asr *StatASR) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if asr.Count < queueLen {
		for id := range asr.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round(asr.Answered/float64(asr.Count),
			roundingDecimal, utils.MetaRoundingMiddle),
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

// getValue returns acd.val
func (acd *StatACD) getValue(roundingDecimal int) time.Duration {
	if acd.val == nil {
		if (acd.MinItems > 0 && acd.Count < int64(acd.MinItems)) || (acd.Count == 0) {
			acd.val = utils.DurationPointer(-time.Nanosecond)
		} else {
			acd.val = utils.DurationPointer(utils.RoundStatDuration(
				time.Duration(acd.Sum.Nanoseconds()/acd.Count),
				roundingDecimal))
		}
	}
	return *acd.val
}

func (acd *StatACD) GetStringValue(roundingDecimal int) (valStr string) {
	if val := acd.getValue(roundingDecimal); val == -time.Nanosecond {
		valStr = utils.NotAvailable
	} else {
		valStr = fmt.Sprintf("%+v", acd.getValue(roundingDecimal))
	}
	return
}

func (acd *StatACD) GetValue(roundingDecimal int) (v interface{}) {
	return acd.getValue(roundingDecimal)
}

func (acd *StatACD) GetFloat64Value(roundingDecimal int) (v float64) {
	if val := acd.getValue(roundingDecimal); val == -time.Nanosecond {
		v = utils.StatsNA
	} else {
		v = float64(acd.getValue(roundingDecimal).Nanoseconds())
	}
	return
}

func (acd *StatACD) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var dur time.Duration
	var val interface{}
	if val, err = ev.FieldAsInterface([]string{utils.MetaReq, utils.Usage}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Usage)
		}
		return
	} else if dur, err = utils.IfaceAsDuration(val); err != nil {
		return
	}
	acd.Sum += dur
	if val, has := acd.Events[evID]; !has {
		acd.Events[evID] = &DurationWithCompress{Duration: dur, CompressFactor: 1}
	} else {
		val.Duration = time.Duration((float64(val.Duration.Nanoseconds())*float64(val.CompressFactor) + float64(dur.Nanoseconds())) / float64(val.CompressFactor+1))
		val.CompressFactor = val.CompressFactor + 1
	}
	acd.Count++
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
	acd.Count--
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

// GetMinItems returns the minim items for the metric
func (acd *StatACD) GetMinItems() (minIts int) { return acd.MinItems }

// Compress is part of StatMetric interface
func (acd *StatACD) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if acd.Count < queueLen {
		for id := range acd.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &DurationWithCompress{
		Duration: utils.RoundStatDuration(time.Duration(acd.Sum.Nanoseconds()/acd.Count),
			roundingDecimal),
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
func (tcd *StatTCD) getValue(roundingDecimal int) time.Duration {
	if tcd.val == nil {
		if (tcd.MinItems > 0 && tcd.Count < int64(tcd.MinItems)) || (tcd.Count == 0) {
			tcd.val = utils.DurationPointer(-time.Nanosecond)
		} else {
			tcd.val = utils.DurationPointer(utils.RoundStatDuration(
				time.Duration(tcd.Sum.Nanoseconds()),
				roundingDecimal))

		}
	}
	return *tcd.val
}

func (tcd *StatTCD) GetStringValue(roundingDecimal int) (valStr string) {
	if val := tcd.getValue(roundingDecimal); val == -time.Nanosecond {
		valStr = utils.NotAvailable
	} else {
		valStr = fmt.Sprintf("%+v", tcd.getValue(roundingDecimal))
	}
	return
}

func (tcd *StatTCD) GetValue(roundingDecimal int) (v interface{}) {
	return tcd.getValue(roundingDecimal)
}

func (tcd *StatTCD) GetFloat64Value(roundingDecimal int) (v float64) {
	if val := tcd.getValue(roundingDecimal); val == -time.Nanosecond {
		v = utils.StatsNA
	} else {
		v = float64(tcd.getValue(roundingDecimal).Nanoseconds())
	}
	return
}

func (tcd *StatTCD) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var dur time.Duration
	var val interface{}
	if val, err = ev.FieldAsInterface([]string{utils.MetaReq, utils.Usage}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Usage)
		}
		return
	} else if dur, err = utils.IfaceAsDuration(val); err != nil {
		return
	}
	tcd.Sum += dur
	if val, has := tcd.Events[evID]; !has {
		tcd.Events[evID] = &DurationWithCompress{Duration: dur, CompressFactor: 1}
	} else {
		val.Duration = time.Duration((float64(val.Duration.Nanoseconds())*float64(val.CompressFactor) + float64(dur.Nanoseconds())) / float64(val.CompressFactor+1))
		val.CompressFactor = val.CompressFactor + 1
	}
	tcd.Count++
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
	tcd.Count--
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

// GetMinItems returns the minim items for the metric
func (tcd *StatTCD) GetMinItems() (minIts int) { return tcd.MinItems }

// Compress is part of StatMetric interface
func (tcd *StatTCD) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if tcd.Count < queueLen {
		for id := range tcd.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &DurationWithCompress{
		Duration: utils.RoundStatDuration(time.Duration(tcd.Sum.Nanoseconds()/tcd.Count),
			roundingDecimal),
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
func (acc *StatACC) getValue(roundingDecimal int) float64 {
	if acc.val == nil {
		if (acc.MinItems > 0 && acc.Count < int64(acc.MinItems)) || (acc.Count == 0) {
			acc.val = utils.Float64Pointer(utils.StatsNA)
		} else {
			acc.val = utils.Float64Pointer(utils.Round(acc.Sum/float64(acc.Count),
				roundingDecimal, utils.MetaRoundingMiddle))
		}
	}
	return *acc.val
}

func (acc *StatACC) GetStringValue(roundingDecimal int) (valStr string) {
	if val := acc.getValue(roundingDecimal); val == utils.StatsNA {
		valStr = utils.NotAvailable
	} else {
		valStr = strconv.FormatFloat(acc.getValue(roundingDecimal), 'f', -1, 64)
	}
	return

}

func (acc *StatACC) GetValue(roundingDecimal int) (v interface{}) {
	return acc.getValue(roundingDecimal)
}

func (acc *StatACC) GetFloat64Value(roundingDecimal int) (v float64) {
	return acc.getValue(roundingDecimal)
}

func (acc *StatACC) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var cost float64
	var val interface{}
	if val, err = ev.FieldAsInterface([]string{utils.MetaReq, utils.Cost}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Cost)
		}
		return
	} else if cost, err = utils.IfaceAsFloat64(val); err != nil {
		return
	} else if cost < 0 {
		return utils.ErrPrefix(utils.ErrNegative, utils.Cost)
	}
	acc.Sum += cost
	if val, has := acc.Events[evID]; !has {
		acc.Events[evID] = &StatWithCompress{Stat: cost, CompressFactor: 1}
	} else {
		val.Stat = (val.Stat*float64(val.CompressFactor) + cost) / float64(val.CompressFactor+1)
		val.CompressFactor = val.CompressFactor + 1
	}
	acc.Count++
	acc.val = nil
	return
}

func (acc *StatACC) RemEvent(evID string) (err error) {
	cost, has := acc.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	acc.Sum -= cost.Stat
	acc.Count--
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

// GetMinItems returns the minim items for the metric
func (acc *StatACC) GetMinItems() (minIts int) { return acc.MinItems }

// Compress is part of StatMetric interface
func (acc *StatACC) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if acc.Count < queueLen {
		for id := range acc.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round(acc.Sum/float64(acc.Count),
			roundingDecimal, utils.MetaRoundingMiddle),
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
func (tcc *StatTCC) getValue(roundingDecimal int) float64 {
	if tcc.val == nil {
		if (tcc.MinItems > 0 && tcc.Count < int64(tcc.MinItems)) || (tcc.Count == 0) {
			tcc.val = utils.Float64Pointer(utils.StatsNA)
		} else {
			tcc.val = utils.Float64Pointer(utils.Round(tcc.Sum,
				roundingDecimal,
				utils.MetaRoundingMiddle))
		}
	}
	return *tcc.val
}

func (tcc *StatTCC) GetStringValue(roundingDecimal int) (valStr string) {
	if val := tcc.getValue(roundingDecimal); val == utils.StatsNA {
		valStr = utils.NotAvailable
	} else {
		valStr = strconv.FormatFloat(tcc.getValue(roundingDecimal), 'f', -1, 64)
	}
	return
}

func (tcc *StatTCC) GetValue(roundingDecimal int) (v interface{}) {
	return tcc.getValue(roundingDecimal)
}

func (tcc *StatTCC) GetFloat64Value(roundingDecimal int) (v float64) {
	return tcc.getValue(roundingDecimal)
}

func (tcc *StatTCC) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var cost float64
	var val interface{}
	if val, err = ev.FieldAsInterface([]string{utils.MetaReq, utils.Cost}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Cost)
		}
		return
	} else if cost, err = utils.IfaceAsFloat64(val); err != nil {
		return
	} else if cost < 0 {
		return utils.ErrPrefix(utils.ErrNegative, utils.Cost)
	}
	tcc.Sum += cost
	if val, has := tcc.Events[evID]; !has {
		tcc.Events[evID] = &StatWithCompress{Stat: cost, CompressFactor: 1}
	} else {
		val.Stat = (val.Stat*float64(val.CompressFactor) + cost) / float64(val.CompressFactor+1)
		val.CompressFactor = val.CompressFactor + 1
	}
	tcc.Count++
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
	tcc.Count--
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

// GetMinItems returns the minim items for the metric
func (tcc *StatTCC) GetMinItems() (minIts int) { return tcc.MinItems }

// Compress is part of StatMetric interface
func (tcc *StatTCC) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if tcc.Count < queueLen {
		for id := range tcc.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round((tcc.Sum / float64(tcc.Count)),
			roundingDecimal, utils.MetaRoundingMiddle),
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
func (pdd *StatPDD) getValue(roundingDecimal int) time.Duration {
	if pdd.val == nil {
		if (pdd.MinItems > 0 && pdd.Count < int64(pdd.MinItems)) || (pdd.Count == 0) {
			pdd.val = utils.DurationPointer(-time.Nanosecond)
		} else {
			pdd.val = utils.DurationPointer(utils.RoundStatDuration(
				time.Duration(pdd.Sum.Nanoseconds()/pdd.Count),
				roundingDecimal))
		}
	}
	return *pdd.val
}

func (pdd *StatPDD) GetStringValue(roundingDecimal int) (valStr string) {
	if val := pdd.getValue(roundingDecimal); val == -time.Nanosecond {
		valStr = utils.NotAvailable
	} else {
		valStr = fmt.Sprintf("%+v", pdd.getValue(roundingDecimal))
	}
	return
}

func (pdd *StatPDD) GetValue(roundingDecimal int) (v interface{}) {
	return pdd.getValue(roundingDecimal)
}

func (pdd *StatPDD) GetFloat64Value(roundingDecimal int) (v float64) {
	if val := pdd.getValue(roundingDecimal); val == -time.Nanosecond {
		v = utils.StatsNA
	} else {
		v = float64(pdd.getValue(roundingDecimal).Nanoseconds())
	}
	return
}

func (pdd *StatPDD) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var dur time.Duration
	var val interface{}
	if val, err = ev.FieldAsInterface([]string{utils.MetaReq, utils.PDD}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.PDD)
		}
		return
	} else if dur, err = utils.IfaceAsDuration(val); err != nil {
		return
	}
	pdd.Sum += dur
	if val, has := pdd.Events[evID]; !has {
		pdd.Events[evID] = &DurationWithCompress{Duration: dur, CompressFactor: 1}
	} else {
		val.Duration = time.Duration((float64(val.Duration.Nanoseconds())*float64(val.CompressFactor) + float64(dur.Nanoseconds())) / float64(val.CompressFactor+1))
		val.CompressFactor = val.CompressFactor + 1
	}
	pdd.Count++
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
	pdd.Count--
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

// GetMinItems returns the minim items for the metric
func (pdd *StatPDD) GetMinItems() (minIts int) { return pdd.MinItems }

// Compress is part of StatMetric interface
func (pdd *StatPDD) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if pdd.Count < queueLen {
		for id := range pdd.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &DurationWithCompress{
		Duration: utils.RoundStatDuration(time.Duration(pdd.Sum.Nanoseconds()/pdd.Count),
			roundingDecimal),
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
	return &StatDDC{Events: make(map[string]map[string]int64), FieldValues: make(map[string]utils.StringSet),
		MinItems: minItems, FilterIDs: filterIDs}, nil
}

type StatDDC struct {
	FilterIDs   []string
	FieldValues map[string]utils.StringSet  // map[fieldValue]map[eventID]
	Events      map[string]map[string]int64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    int
	Count       int64
}

// getValue returns tcd.val
func (ddc *StatDDC) getValue(roundingDecimal int) float64 {
	if ddc.Count == 0 || ddc.Count < int64(ddc.MinItems) {
		return utils.StatsNA
	}
	return float64(len(ddc.FieldValues))
}

func (ddc *StatDDC) GetStringValue(roundingDecimal int) (valStr string) {
	if val := ddc.getValue(roundingDecimal); val == utils.StatsNA {
		valStr = utils.NotAvailable
	} else {
		valStr = strconv.FormatFloat(ddc.getValue(roundingDecimal), 'f', -1, 64)
	}
	return
}

func (ddc *StatDDC) GetValue(roundingDecimal int) (v interface{}) {
	return ddc.getValue(roundingDecimal)
}

func (ddc *StatDDC) GetFloat64Value(roundingDecimal int) (v float64) {
	return ddc.getValue(roundingDecimal)
}

func (ddc *StatDDC) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var fieldValue string
	if fieldValue, err = ev.FieldAsString([]string{utils.MetaReq, utils.Destination}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Destination)
		}
		return
	}

	// add to fieldValues
	if _, has := ddc.FieldValues[fieldValue]; !has {
		ddc.FieldValues[fieldValue] = make(utils.StringSet)
	}
	ddc.FieldValues[fieldValue].Add(evID)

	// add to events
	if _, has := ddc.Events[evID]; !has {
		ddc.Events[evID] = make(map[string]int64)
	}
	ddc.Count++
	if _, has := ddc.Events[evID][fieldValue]; !has {
		ddc.Events[evID][fieldValue] = 1
		return
	}
	ddc.Events[evID][fieldValue] = ddc.Events[evID][fieldValue] + 1
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
	for k := range fieldValues {
		fieldValue = k
		break
	}
	ddc.Count--
	if fieldValues[fieldValue] > 1 {
		ddc.Events[evID][fieldValue] = ddc.Events[evID][fieldValue] - 1
		return // do not delete the reference until it reaches 0
	}
	delete(ddc.Events[evID], fieldValue)

	// remove from fieldValues
	if _, has := ddc.FieldValues[fieldValue]; !has {
		return
	}
	ddc.FieldValues[fieldValue].Remove(evID)
	if ddc.FieldValues[fieldValue].Size() <= 0 {
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

// GetMinItems returns the minim items for the metric
func (ddc *StatDDC) GetMinItems() (minIts int) { return ddc.MinItems }

func (ddc *StatDDC) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	for id := range ddc.Events {
		eventIDs = append(eventIDs, id)
	}
	return
}

////////////////////

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
func (sum *StatSum) getValue(roundingDecimal int) float64 {
	if sum.val == nil {
		if len(sum.Events) == 0 || sum.Count < int64(sum.MinItems) {
			sum.val = utils.Float64Pointer(utils.StatsNA)
		} else {
			sum.val = utils.Float64Pointer(utils.Round(sum.Sum,
				roundingDecimal,
				utils.MetaRoundingMiddle))
		}
	}
	return *sum.val
}

func (sum *StatSum) GetStringValue(roundingDecimal int) (valStr string) {
	if val := sum.getValue(roundingDecimal); val == utils.StatsNA {
		valStr = utils.NotAvailable
	} else {
		valStr = strconv.FormatFloat(sum.getValue(roundingDecimal), 'f', -1, 64)
	}
	return
}

func (sum *StatSum) GetValue(roundingDecimal int) (v interface{}) {
	return sum.getValue(roundingDecimal)
}

func (sum *StatSum) GetFloat64Value(roundingDecimal int) (v float64) {
	return sum.getValue(roundingDecimal)
}

func (sum *StatSum) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var val float64
	var ival interface{}
	if ival, err = utils.DPDynamicInterface(sum.FieldName, ev); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, sum.FieldName)
		}
		return
	} else if val, err = utils.IfaceAsFloat64(ival); err != nil {
		return
	}
	sum.Sum += val
	if v, has := sum.Events[evID]; !has {
		sum.Events[evID] = &StatWithCompress{Stat: val, CompressFactor: 1}
	} else {
		v.Stat = (v.Stat*float64(v.CompressFactor) + val) / float64(v.CompressFactor+1)
		v.CompressFactor = v.CompressFactor + 1
	}
	sum.Count++
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
	sum.Count--
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

// GetMinItems returns the minim items for the metric
func (sum *StatSum) GetMinItems() (minIts int) { return sum.MinItems }

// Compress is part of StatMetric interface
func (sum *StatSum) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if sum.Count < queueLen {
		for id := range sum.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round((sum.Sum / float64(sum.Count)),
			roundingDecimal, utils.MetaRoundingMiddle),
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
func (avg *StatAverage) getValue(roundingDecimal int) float64 {
	if avg.val == nil {
		if (avg.MinItems > 0 && avg.Count < int64(avg.MinItems)) || (avg.Count == 0) {
			avg.val = utils.Float64Pointer(utils.StatsNA)
		} else {
			avg.val = utils.Float64Pointer(utils.Round((avg.Sum / float64(avg.Count)),
				roundingDecimal, utils.MetaRoundingMiddle))
		}
	}
	return *avg.val
}

func (avg *StatAverage) GetStringValue(roundingDecimal int) (valStr string) {
	if val := avg.getValue(roundingDecimal); val == utils.StatsNA {
		valStr = utils.NotAvailable
	} else {
		valStr = strconv.FormatFloat(avg.getValue(roundingDecimal), 'f', -1, 64)
	}
	return

}

func (avg *StatAverage) GetValue(roundingDecimal int) (v interface{}) {
	return avg.getValue(roundingDecimal)
}

func (avg *StatAverage) GetFloat64Value(roundingDecimal int) (v float64) {
	return avg.getValue(roundingDecimal)
}

func (avg *StatAverage) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var val float64
	var ival interface{}
	if ival, err = utils.DPDynamicInterface(avg.FieldName, ev); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, avg.FieldName)
		}
		return
	} else if val, err = utils.IfaceAsFloat64(ival); err != nil {
		return
	}
	avg.Sum += val
	if v, has := avg.Events[evID]; !has {
		avg.Events[evID] = &StatWithCompress{Stat: val, CompressFactor: 1}
	} else {
		v.Stat = (v.Stat*float64(v.CompressFactor) + val) / float64(v.CompressFactor+1)
		v.CompressFactor = v.CompressFactor + 1
	}
	avg.Count++
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
	avg.Count--
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

// GetMinItems returns the minim items for the metric
func (avg *StatAverage) GetMinItems() (minIts int) { return avg.MinItems }

// Compress is part of StatMetric interface
func (avg *StatAverage) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	if avg.Count < queueLen {
		for id := range avg.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	stat := &StatWithCompress{
		Stat: utils.Round((avg.Sum / float64(avg.Count)),
			roundingDecimal, utils.MetaRoundingMiddle),
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
	return &StatDistinct{Events: make(map[string]map[string]int64), FieldValues: make(map[string]utils.StringSet),
		MinItems: minItems, FieldName: extraParams, FilterIDs: filterIDs}, nil
}

type StatDistinct struct {
	FilterIDs   []string
	FieldValues map[string]utils.StringSet  // map[fieldValue]map[eventID]
	Events      map[string]map[string]int64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    int
	FieldName   string
	Count       int64
}

// getValue returns tcd.val
func (dst *StatDistinct) getValue(roundingDecimal int) float64 {
	if dst.Count == 0 || dst.Count < int64(dst.MinItems) {
		return utils.StatsNA
	}
	return float64(len(dst.FieldValues))
}

func (dst *StatDistinct) GetStringValue(roundingDecimal int) (valStr string) {
	if val := dst.getValue(roundingDecimal); val == utils.StatsNA {
		valStr = utils.NotAvailable
	} else {
		valStr = strconv.FormatFloat(dst.getValue(roundingDecimal), 'f', -1, 64)
	}
	return
}

func (dst *StatDistinct) GetValue(roundingDecimal int) (v interface{}) {
	return dst.getValue(roundingDecimal)
}

func (dst *StatDistinct) GetFloat64Value(roundingDecimal int) (v float64) {
	return dst.getValue(roundingDecimal)
}

func (dst *StatDistinct) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var fieldValue string
	// simply remove the ~*req. prefix and do normal process
	if !strings.HasPrefix(dst.FieldName, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep) {
		return fmt.Errorf("Invalid format for field <%s>", dst.FieldName)
	}

	if fieldValue, err = utils.DPDynamicString(dst.FieldName, ev); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, dst.FieldName)
		}
		return
	}

	// add to fieldValues
	if _, has := dst.FieldValues[fieldValue]; !has {
		dst.FieldValues[fieldValue] = make(utils.StringSet)
	}
	dst.FieldValues[fieldValue].Add(evID)

	// add to events
	if _, has := dst.Events[evID]; !has {
		dst.Events[evID] = make(map[string]int64)
	}
	dst.Count++
	if _, has := dst.Events[evID][fieldValue]; !has {
		dst.Events[evID][fieldValue] = 1
		return
	}
	dst.Events[evID][fieldValue] = dst.Events[evID][fieldValue] + 1
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
	for k := range fieldValues {
		fieldValue = k
		break
	}
	dst.Count--
	if fieldValues[fieldValue] > 1 {
		dst.Events[evID][fieldValue] = dst.Events[evID][fieldValue] - 1
		return // do not delete the reference until it reaches 0
	}
	delete(dst.Events[evID], fieldValue)

	// remove from fieldValues
	if _, has := dst.FieldValues[fieldValue]; !has {
		return
	}
	dst.FieldValues[fieldValue].Remove(evID)
	if dst.FieldValues[fieldValue].Size() <= 0 {
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

// GetMinItems returns the minim items for the metric
func (dst *StatDistinct) GetMinItems() (minIts int) { return dst.MinItems }

func (dst *StatDistinct) Compress(queueLen int64, defaultID string, roundingDecimal int) (eventIDs []string) {
	for id := range dst.Events {
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
