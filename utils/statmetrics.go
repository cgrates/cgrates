/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"maps"

	"github.com/ericlagergren/decimal"
)

type metricConstructor func(uint64, string, []string) StatMetric
type metricConstructorErr func(uint64, string, []string) (StatMetric, error)

// withErrReturn wraps a constructor to return an error (always nil).
func withErrReturn(fn metricConstructor) metricConstructorErr {
	return func(minItems uint64, extraParams string, filterIDs []string) (StatMetric, error) {
		return fn(minItems, extraParams, filterIDs), nil
	}
}

// NewStatMetric instantiates the StatMetric
// cfg serves as general purpose container to pass config options to metric
func NewStatMetric(metricID string, minItems uint64, filterIDs []string) (sm StatMetric, err error) {
	metrics := map[string]metricConstructorErr{
		MetaASR:      withErrReturn(NewASR),
		MetaACD:      withErrReturn(NewACD),
		MetaTCD:      withErrReturn(NewTCD),
		MetaACC:      withErrReturn(NewACC),
		MetaTCC:      withErrReturn(NewTCC),
		MetaPDD:      withErrReturn(NewPDD),
		MetaDDC:      withErrReturn(NewDDC),
		MetaSum:      NewStatSum, // Already returns (StatMetric, error)
		MetaAverage:  withErrReturn(NewStatAverage),
		MetaDistinct: withErrReturn(NewStatDistinct),
		MetaHighest:  withErrReturn(NewStatHighest),
		MetaLowest:   withErrReturn(NewStatLowest),
		MetaREPSC:    withErrReturn(NewStatREPSC),
		MetaREPFC:    withErrReturn(NewStatREPFC),
	}
	// split the metricID
	// in case of *sum we have *sum#~*req.FieldName
	metricSplit := strings.Split(metricID, HashtagSep)
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
	GetValue() *Decimal
	GetStringValue(rounding int) string
	AddOneEvent(ev DataProvider) error
	AddEvent(evID string, ev DataProvider) error
	RemEvent(evID string) error
	GetMinItems() (minIts uint64)
	Compress(queueLen uint64, defaultID string) (eventIDs []string)
	GetCompressFactor(events map[string]uint64) map[string]uint64
	Clone() StatMetric
	GetFilterIDs() []string
}

func NewASR(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatASR{Metric: NewMetric(minItems, filterIDs)}
}

// TimezoneOf returns the provider's timezone if it exposes one, else "".
func TimezoneOf(dp DataProvider) string {
	if p, ok := dp.(interface{ Timezone() string }); ok {
		return p.Timezone()
	}
	return ""
}

// ASR implements AverageSuccessRatio metric
type StatASR struct {
	*Metric
}

func (asr *StatASR) GetStringValue(rounding int) (valStr string) {
	valStr = NotAvailable
	if val := asr.getAvgValue(); val != DecimalNaN {
		v, _ := MultiplyDecimal(val, NewDecimal(100, 0)).Round(rounding).Float64()
		valStr = strconv.FormatFloat(v, 'f', -1, 64) + "%"
	}
	return
}

func (asr *StatASR) GetValue() (val *Decimal) {
	if val = asr.getAvgValue(); val != DecimalNaN {
		val = MultiplyDecimal(val, NewDecimal(100, 0))
	}
	return
}

func (asr *StatASR) AddOneEvent(ev DataProvider) (err error) {
	var (
		answered int64
		val      any
	)
	if val, err = ev.FieldAsInterface([]string{MetaOpts, MetaStartTime}); err != nil {
		if err != ErrNotFound {
			return
		}
	} else if at, err := IfaceAsTime(val, TimezoneOf(ev)); err != nil {
		return err
	} else if !at.IsZero() {
		answered = 1
	}

	return asr.addOneEvent(answered)
}

// AddEvent is part of StatMetric interface
func (asr *StatASR) AddEvent(evID string, ev DataProvider) (err error) {
	var answered int
	var val any
	if val, err = ev.FieldAsInterface([]string{MetaOpts, MetaStartTime}); err != nil {
		if err != ErrNotFound {
			return err
		}
	} else if at, err := IfaceAsTime(val, TimezoneOf(ev)); err != nil {
		return err
	} else if !at.IsZero() {
		answered = 1
	}
	return asr.addEvent(evID, answered)
}

func (asr *StatASR) RemEvent(evID string) (err error) {
	val, has := asr.Events[evID]
	if !has {
		return ErrNotFound
	}
	ans := NewDecimal(0, 0)
	if val.Stat.Compare(NewDecimalFromFloat64(0.5)) > 0 {
		ans := NewDecimal(1, 0)
		asr.Value = SubstractDecimal(asr.Value, ans)
	}
	asr.Count--
	if val.CompressFactor <= 1 {
		delete(asr.Events, evID)
	} else {
		val.Stat = DivideDecimal(
			SubstractDecimal(
				MultiplyDecimal(val.Stat, NewDecimal(int64(val.CompressFactor), 0)),
				ans),
			NewDecimal(int64(val.CompressFactor)-1, 0))
		val.CompressFactor = val.CompressFactor - 1
	}
	return
}

func (asr *StatASR) Clone() StatMetric {
	return &StatASR{
		Metric: asr.Metric.Clone(),
	}
}

func NewACD(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatACD{Metric: NewMetric(minItems, filterIDs)}
}

// ACD implements AverageCallDuration metric
type StatACD struct {
	*Metric
}

func (acd *StatACD) GetStringValue(rounding int) string {
	if acd.Count == 0 || acd.Count < acd.MinItems {
		return NotAvailable
	}
	v, _ := acd.getAvgValue().Round(rounding).Duration()
	return v.String()
}

func (acd *StatACD) GetValue() *Decimal {
	return acd.getAvgValue()
}

func (acd *StatACD) AddEvent(evID string, ev DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaUsage})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaUsage)
		}
		return err
	}
	return acd.addEvent(evID, ival)
}

func (acd *StatACD) AddOneEvent(ev DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaUsage})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaUsage)
		}
		return err
	}

	return acd.addOneEvent(ival)

}

func (acd *StatACD) Clone() StatMetric {
	return &StatACD{
		Metric: acd.Metric.Clone(),
	}
}

func NewTCD(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatTCD{Metric: NewMetric(minItems, filterIDs)}
}

// TCD implements TotalCallDuration metric
type StatTCD struct {
	*Metric
}

func (sum *StatTCD) GetStringValue(rounding int) string {
	if sum.Count == 0 || sum.Count < sum.MinItems {
		return NotAvailable
	}
	v, _ := sum.Value.Round(rounding).Duration()
	return v.String()
}

func (sum *StatTCD) AddEvent(evID string, ev DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaUsage})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaUsage)
		}
		return err
	}
	return sum.addEvent(evID, ival)
}

func (sum *StatTCD) AddOneEvent(ev DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaUsage})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaUsage)
		}
		return err
	}
	return sum.addOneEvent(ival)

}

func (sum *StatTCD) Clone() StatMetric {
	return &StatTCD{
		Metric: sum.Metric.Clone(),
	}
}

func NewACC(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatACC{Metric: NewMetric(minItems, filterIDs)}
}

// ACC implements AverageCallCost metric
type StatACC struct {
	*Metric
}

func (acc *StatACC) GetStringValue(rounding int) string {
	return acc.getAvgStringValue(rounding)
}

func (acc *StatACC) GetValue() *Decimal {
	return acc.getAvgValue()
}

func (acc *StatACC) AddEvent(evID string, ev DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaCost})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaCost)
		}
		return err
	}
	val, err := IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return ErrPrefix(ErrNegative, MetaCost)
	}
	return acc.addEvent(evID, val)
}

func (acc *StatACC) AddOneEvent(ev DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaCost})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaCost)
		}
		return err
	}
	val, err := IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return ErrPrefix(ErrNegative, MetaCost)
	}
	return acc.addOneEvent(val)
}

func (acc *StatACC) Clone() StatMetric {
	return &StatACC{
		Metric: acc.Metric.Clone(),
	}
}

func NewTCC(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatTCC{Metric: NewMetric(minItems, filterIDs)}
}

// TCC implements TotalCallCost metric
type StatTCC struct {
	*Metric
}

func (tcc *StatTCC) AddEvent(evID string, ev DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaCost})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaCost)
		}
		return err
	}
	val, err := IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return ErrPrefix(ErrNegative, MetaCost)
	}
	return tcc.addEvent(evID, val)
}

func (tcc *StatTCC) AddOneEvent(ev DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaCost})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaCost)
		}
		return err
	}
	val, err := IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return ErrPrefix(ErrNegative, MetaCost)
	}
	return tcc.addOneEvent(ival)
}

func (tcc *StatTCC) Clone() StatMetric {
	return &StatTCC{
		Metric: tcc.Metric.Clone(),
	}
}

func NewPDD(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatPDD{Metric: NewMetric(minItems, filterIDs)}
}

// PDD implements Post Dial Delay (average) metric
type StatPDD struct {
	*Metric
}

func (pdd *StatPDD) GetStringValue(rounding int) string {
	if pdd.Count == 0 || pdd.Count < pdd.MinItems {
		return NotAvailable
	}
	v, _ := pdd.getAvgValue().Round(rounding).Duration()
	return v.String()
}

func (pdd *StatPDD) GetValue() *Decimal {
	return pdd.getAvgValue()
}

func (pdd *StatPDD) AddEvent(evID string, ev DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaPDD})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaPDD)
		}
		return err
	}
	return pdd.addEvent(evID, ival)
}

func (pdd *StatPDD) AddOneEvent(ev DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{MetaOpts, MetaPDD})
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaPDD)
		}
		return err
	}
	return pdd.addOneEvent(ival)
}

func (pdd *StatPDD) Clone() StatMetric {
	return &StatPDD{
		Metric: pdd.Metric.Clone(),
	}
}

func NewDDC(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatDDC{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]StringSet),
		MinItems:    minItems,
		FilterIDs:   filterIDs,
	}
}

// StatDDC count  values occurring in destination field
type StatDDC struct {
	FieldValues map[string]StringSet         // map[fieldValue]map[eventID]
	Events      map[string]map[string]uint64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    uint64
	Count       uint64
	FilterIDs   []string
}

func (ddc *StatDDC) GetFilterIDs() []string { return ddc.FilterIDs }

func (ddc *StatDDC) GetStringValue(rounding int) (valStr string) {
	valStr = NotAvailable
	if val := ddc.GetValue(); val != DecimalNaN {
		v, _ := val.Round(rounding).Float64()
		valStr = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return
}

func (ddc *StatDDC) GetValue() *Decimal {
	if ddc.Count == 0 || ddc.Count < ddc.MinItems {
		return DecimalNaN
	}
	return NewDecimal(int64(len(ddc.FieldValues)), 0)
}

func (ddc *StatDDC) AddEvent(evID string, ev DataProvider) (err error) {
	var fieldValue string
	if fieldValue, err = ev.FieldAsString([]string{MetaOpts, MetaDestination}); err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaDestination)
		}
		return
	}

	// add to fieldValues
	if _, has := ddc.FieldValues[fieldValue]; !has {
		ddc.FieldValues[fieldValue] = make(StringSet)
	}
	ddc.FieldValues[fieldValue].Add(evID)

	// add to events
	if _, has := ddc.Events[evID]; !has {
		ddc.Events[evID] = make(map[string]uint64)
	}
	ddc.Count++
	if _, has := ddc.Events[evID][fieldValue]; !has {
		ddc.Events[evID][fieldValue] = 1
		return
	}
	ddc.Events[evID][fieldValue] = ddc.Events[evID][fieldValue] + 1
	return
}

func (ddc *StatDDC) AddOneEvent(ev DataProvider) (err error) {
	var fieldValue string
	if fieldValue, err = ev.FieldAsString([]string{MetaOpts, MetaDestination}); err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, MetaDestination)
		}
		return
	}
	if _, has := ddc.FieldValues[fieldValue]; !has {
		ddc.FieldValues[fieldValue] = make(StringSet)
	}
	ddc.Count++
	return
}

func (ddc *StatDDC) RemEvent(evID string) (err error) {
	fieldValues, has := ddc.Events[evID]
	if !has {
		return ErrNotFound
	}
	if len(fieldValues) == 0 {
		delete(ddc.Events, evID)
		return ErrNotFound
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

// GetMinItems returns the minim items for the metric
func (ddc *StatDDC) GetMinItems() (minIts uint64) { return ddc.MinItems }

func (ddc *StatDDC) Compress(queueLen uint64, defaultID string) (eventIDs []string) {
	eventIDs = make([]string, 0, len(ddc.Events))
	for id := range ddc.Events {
		eventIDs = append(eventIDs, id)
	}
	return
}

// Compress is part of StatMetric interface
func (ddc *StatDDC) GetCompressFactor(events map[string]uint64) map[string]uint64 {
	for id, ev := range ddc.Events {
		var compressFactor uint64
		for _, fields := range ev {
			compressFactor += fields
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

func (ddc *StatDDC) Clone() StatMetric {
	if ddc == nil {
		return nil
	}
	cln := &StatDDC{
		FieldValues: make(map[string]StringSet),
		Count:       ddc.Count,
		Events:      make(map[string]map[string]uint64),
		MinItems:    ddc.MinItems,
		FilterIDs:   slices.Clone(ddc.FilterIDs),
	}
	for k, v := range ddc.Events {
		cln.Events[k] = make(map[string]uint64)
		for d, n := range v {
			cln.Events[k][d] = n
		}
	}
	for k, v := range ddc.FieldValues {
		cln.FieldValues[k] = v.Clone()
	}
	return cln
}

// ACDHelper structure
type DecimalWithCompress struct {
	Stat           *Decimal
	CompressFactor uint64
}

func NewMetric(minItems uint64, filterIDs []string) *Metric {
	return &Metric{
		Value:     NewDecimal(0, 0),
		Events:    make(map[string]*DecimalWithCompress),
		MinItems:  minItems,
		FilterIDs: filterIDs,
	}
}

type Metric struct {
	Value     *Decimal
	Count     uint64
	Events    map[string]*DecimalWithCompress // map[EventTenantID]Cost
	MinItems  uint64
	FilterIDs []string
}

func (m *Metric) GetFilterIDs() []string { return m.FilterIDs }

func (m *Metric) getTotalValue() *Decimal {
	if m.Count == 0 || m.Count < m.MinItems {
		return DecimalNaN
	}
	return m.Value
}

func (m *Metric) getAvgValue() *Decimal {
	if m.Count == 0 || m.Count < m.MinItems {
		return DecimalNaN
	}
	return DivideDecimal(m.Value, NewDecimal(int64(m.Count), 0))
}

func (m *Metric) getAvgStringValue(rounding int) string {
	if m.Count == 0 || m.Count < m.MinItems {
		return NotAvailable
	}
	v, _ := DivideDecimal(m.Value, NewDecimal(int64(m.Count), 0)).Round(rounding).Float64()
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (m *Metric) GetStringValue(rounding int) string {
	if m.Count == 0 || m.Count < m.MinItems {
		return NotAvailable
	}
	v, _ := m.Value.Round(rounding).Float64()
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (m *Metric) GetValue() (v *Decimal) {
	return m.getTotalValue()
}

func (m *Metric) addEvent(evID string, ival any) (err error) {
	var val *decimal.Big
	if val, err = IfaceAsBig(ival); err != nil {
		return
	}
	dVal := &Decimal{Big: val}
	m.Value = SumDecimal(m.Value, dVal)
	if v, has := m.Events[evID]; !has {
		m.Events[evID] = &DecimalWithCompress{Stat: dVal, CompressFactor: 1}
	} else {
		v.Stat = DivideDecimal(
			SumDecimal(
				MultiplyDecimal(v.Stat, NewDecimal(int64(v.CompressFactor), 0)),
				dVal),
			NewDecimal(int64(v.CompressFactor)+1, 0))
		v.CompressFactor = v.CompressFactor + 1
	}
	m.Count++
	return
}

// Adding aggregated metrics without events
func (m *Metric) addOneEvent(ival any) (err error) {
	var val *decimal.Big
	if val, err = IfaceAsBig(ival); err != nil {
		return
	}
	dVal := &Decimal{Big: val}
	m.Value = SumDecimal(m.Value, dVal)
	m.Count++
	return
}

// Deleting a specific event and updating metrics
func (m *Metric) RemEvent(evID string) (err error) {
	val, has := m.Events[evID]
	if !has {
		return ErrNotFound
	}
	if val.Stat.Compare(NewDecimal(0, 0)) != 0 {
		m.Value = SubstractDecimal(m.Value, val.Stat)
	}
	m.Count--
	if val.CompressFactor <= 1 {
		delete(m.Events, evID)
	} else {
		val.CompressFactor = val.CompressFactor - 1
	}
	return
}

// GetMinItems returns the minim items for the metric
func (m *Metric) GetMinItems() uint64 { return m.MinItems }

// Compress is part of StatMetric interface
func (m *Metric) Compress(queueLen uint64, defaultID string) (eventIDs []string) {
	if m.Count < queueLen {
		eventIDs = make([]string, 0, len(m.Events))
		for id := range m.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	m.Events = map[string]*DecimalWithCompress{defaultID: {
		Stat:           DivideDecimal(m.Value, NewDecimalFromFloat64(float64(m.Count))),
		CompressFactor: m.Count,
	}}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (m *Metric) GetCompressFactor(events map[string]uint64) map[string]uint64 {
	for id, val := range m.Events {
		if _, has := events[id]; !has {
			events[id] = val.CompressFactor
		}
		if events[id] < val.CompressFactor {
			events[id] = val.CompressFactor
		}
	}
	return events
}

func (m *Metric) Clone() (cln *Metric) {
	if m == nil {
		return nil
	}
	cln = &Metric{
		Count:    m.Count,
		MinItems: m.MinItems,
	}
	if m.Value != nil {
		cln.Value = m.Value.Clone()
	}
	if m.Events != nil {
		cln.Events = make(map[string]*DecimalWithCompress, len(m.Events))
		maps.Copy(cln.Events, m.Events)
	}
	if m.FilterIDs != nil {
		cln.FilterIDs = make([]string, len(m.FilterIDs))
		cln.FilterIDs = slices.Clone(m.FilterIDs)
	}
	return
}

func (m *Metric) Equal(v *Metric) bool {
	if m.MinItems != v.MinItems ||
		m.Count != v.Count ||
		m.Value.Compare(v.Value) != 0 ||
		len(m.Events) != len(v.Events) {
		return false
	}
	for k, c1 := range m.Events {
		c2, has := v.Events[k]
		if !has ||
			c1.CompressFactor != c2.CompressFactor ||
			c1.Stat.Compare(c2.Stat) != 0 {
			return false
		}
	}
	return true
}

func NewStatSum(minItems uint64, fieldName string, filterIDs []string) (StatMetric, error) {
	flds, err := NewRSRParsers(fieldName, InfieldSep)
	if err != nil {
		return nil, err
	}
	return &StatSum{
		Metric: NewMetric(minItems, filterIDs),
		Fields: flds,
	}, nil
}

type StatSum struct {
	*Metric
	Fields RSRParsers
}

func (sum *StatSum) AddEvent(evID string, ev DataProvider) error {
	ival, err := sum.Fields.ParseDataProvider(ev)
	if err != nil {
		return err
	}
	return sum.addEvent(evID, ival)
}
func (sum *StatSum) AddOneEvent(ev DataProvider) error {
	ival, err := sum.Fields.ParseDataProvider(ev)
	if err != nil {
		return err
	}
	return sum.addOneEvent(ival)
}

func (sum *StatSum) Clone() StatMetric {
	return &StatSum{
		Metric: sum.Metric.Clone(),
		Fields: sum.Fields,
	}
}

func NewStatAverage(minItems uint64, fieldName string, filterIDs []string) StatMetric {
	return &StatAverage{Metric: NewMetric(minItems, filterIDs),
		FieldName: fieldName}
}

// StatAverage implements TotalCallCost metric
type StatAverage struct {
	*Metric
	FieldName string
}

func (avg *StatAverage) GetStringValue(rounding int) string {
	return avg.getAvgStringValue(rounding)
}

func (avg *StatAverage) GetValue() *Decimal {
	return avg.getAvgValue()
}

func (avg *StatAverage) AddEvent(evID string, ev DataProvider) error {
	ival, err := DPDynamicInterface(avg.FieldName, ev)
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, avg.FieldName)
		}
		return err
	}
	return avg.addEvent(evID, ival)
}

func (avg *StatAverage) AddOneEvent(ev DataProvider) error {
	ival, err := DPDynamicInterface(avg.FieldName, ev)
	if err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, avg.FieldName)
		}
		return err
	}
	return avg.addOneEvent(ival)
}

func (avg *StatAverage) Clone() StatMetric {
	return &StatAverage{
		Metric:    avg.Metric.Clone(),
		FieldName: avg.FieldName,
	}
}

// StatDistinct counts the different values occurring  in a specific event field
func NewStatDistinct(minItems uint64, fieldName string, filterIDs []string) StatMetric {
	return &StatDistinct{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]StringSet),
		MinItems:    minItems,
		FieldName:   fieldName,
		FilterIDs:   filterIDs,
	}
}

type StatDistinct struct {
	FieldValues map[string]StringSet         // map[fieldValue]map[eventID]
	Events      map[string]map[string]uint64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    uint64
	FieldName   string
	Count       uint64
	FilterIDs   []string
}

func (dst *StatDistinct) GetFilterIDs() []string { return dst.FilterIDs }

func (dst *StatDistinct) GetStringValue(rounding int) (valStr string) {
	valStr = NotAvailable
	if val := dst.GetValue(); val != DecimalNaN {
		v, _ := val.Round(rounding).Float64()
		valStr = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return
}

func (dst *StatDistinct) GetValue() *Decimal {
	if dst.Count == 0 || dst.Count < dst.MinItems {
		return DecimalNaN
	}
	return NewDecimal(int64(len(dst.FieldValues)), 0)
}

func (dst *StatDistinct) AddEvent(evID string, ev DataProvider) (err error) {
	var fieldValue string
	// simply remove the ~*req./~*opts. prefix and do normal process
	if !strings.HasPrefix(dst.FieldName, DynamicDataPrefix+MetaReq+NestingSep) && !strings.HasPrefix(dst.FieldName, DynamicDataPrefix+MetaOpts+NestingSep) {
		return fmt.Errorf("invalid format for field <%s>", dst.FieldName)
	}

	if fieldValue, err = DPDynamicString(dst.FieldName, ev); err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, dst.FieldName)
		}
		return
	}

	// add to fieldValues
	if _, has := dst.FieldValues[fieldValue]; !has {
		dst.FieldValues[fieldValue] = make(StringSet)
	}
	dst.FieldValues[fieldValue].Add(evID)

	// add to events
	if _, has := dst.Events[evID]; !has {
		dst.Events[evID] = make(map[string]uint64)
	}
	dst.Count++
	if _, has := dst.Events[evID][fieldValue]; !has {
		dst.Events[evID][fieldValue] = 1
		return
	}
	dst.Events[evID][fieldValue] = dst.Events[evID][fieldValue] + 1
	return
}

func (dst *StatDistinct) AddOneEvent(ev DataProvider) (err error) {
	var fieldValue string
	// simply remove the ~*req./~*opts. prefix and do normal process
	if !strings.HasPrefix(dst.FieldName, DynamicDataPrefix+MetaReq+NestingSep) && !strings.HasPrefix(dst.FieldName, DynamicDataPrefix+MetaOpts+NestingSep) {
		return fmt.Errorf("invalid format for field <%s>", dst.FieldName)
	}
	if fieldValue, err = DPDynamicString(dst.FieldName, ev); err != nil {
		if err == ErrNotFound {
			err = ErrPrefix(err, dst.FieldName)
		}
		return
	}
	// add to fieldValues
	if _, has := dst.FieldValues[fieldValue]; !has {
		dst.FieldValues[fieldValue] = make(StringSet)
	}

	dst.Count++
	return
}

func (dst *StatDistinct) RemEvent(evID string) (err error) {
	fieldValues, has := dst.Events[evID]
	if !has {
		return ErrNotFound
	}
	if len(fieldValues) == 0 {
		delete(dst.Events, evID)
		return ErrNotFound
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

// GetMinItems returns the minim items for the metric
func (dst *StatDistinct) GetMinItems() uint64 { return dst.MinItems }

func (dst *StatDistinct) Compress(uint64, string) (eventIDs []string) {
	eventIDs = make([]string, 0, len(dst.Events))
	for id := range dst.Events {
		eventIDs = append(eventIDs, id)
	}
	return
}

// Compress is part of StatMetric interface
func (dst *StatDistinct) GetCompressFactor(events map[string]uint64) map[string]uint64 {
	for id, ev := range dst.Events {
		var compressFactor uint64
		for _, fields := range ev {
			compressFactor += fields
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

func (dst *StatDistinct) Clone() StatMetric {
	if dst == nil {
		return nil
	}
	cln := &StatDistinct{
		Count:       dst.Count,
		Events:      make(map[string]map[string]uint64),
		MinItems:    dst.MinItems,
		FieldName:   dst.FieldName,
		FieldValues: make(map[string]StringSet),
		FilterIDs:   slices.Clone(dst.FilterIDs),
	}
	for k, v := range dst.Events {
		cln.Events[k] = make(map[string]uint64)
		for d, n := range v {
			cln.Events[k][d] = n
		}
	}
	for k, v := range dst.FieldValues {
		cln.FieldValues[k] = v.Clone()
	}
	return cln
}

// NewStatHighest creates a StatHighest metric for tracking maximum field values.
func NewStatHighest(minItems uint64, fieldName string, filterIDs []string) StatMetric {
	return &StatHighest{
		FilterIDs: filterIDs,
		MinItems:  minItems,
		FieldName: fieldName,
		Highest:   NewDecimal(0, 0),
		Events:    make(map[string]*Decimal),
	}
}

// StatHighest tracks the maximum value for a specific field across events.
type StatHighest struct {
	FilterIDs []string // event filters to apply before processing
	FieldName string   // field path to extract from events
	MinItems  uint64   // minimum events required for valid results

	Highest *Decimal            // current maximum value tracked
	Count   uint64              // number of events currently tracked
	Events  map[string]*Decimal // event values indexed by ID for deletion
}

// Clone creates a deep copy of StatHighest.
func (s *StatHighest) Clone() StatMetric {
	if s == nil {
		return nil
	}
	clone := &StatHighest{
		FilterIDs: slices.Clone(s.FilterIDs),
		Highest:   s.Highest,
		Count:     s.Count,
		MinItems:  s.MinItems,
		FieldName: s.FieldName,
		Events:    maps.Clone(s.Events),
	}
	return clone
}

func (s *StatHighest) GetStringValue(decimals int) string {
	if s.Count == 0 || s.Count < s.MinItems {
		return NotAvailable
	}
	v, _ := s.Highest.Round(decimals).Float64()
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (s *StatHighest) GetValue() *Decimal {
	if s.Count == 0 || s.Count < s.MinItems {
		return DecimalNaN
	}
	return s.Highest
}

// AddEvent processes a new event, updating highest value if necessary
func (s *StatHighest) AddEvent(evID string, ev DataProvider) error {
	val, err := fieldValueFromDP(s.FieldName, ev)
	if err != nil {
		return err
	}
	if val.Compare(s.Highest) == 1 {
		s.Highest = val
	}

	// Only increment count for new events.
	if _, exists := s.Events[evID]; !exists {
		s.Count++
	}

	s.Events[evID] = val
	return nil
}

// AddOneEvent processes event without storing for removal (used when events
// never expire).
func (s *StatHighest) AddOneEvent(ev DataProvider) error {
	val, err := fieldValueFromDP(s.FieldName, ev)
	if err != nil {
		return err
	}
	if val.Compare(s.Highest) == 1 {
		s.Highest = val
	}
	s.Count++
	return nil
}

func (s *StatHighest) RemEvent(evID string) error {
	v, exists := s.Events[evID]
	if !exists {
		return ErrNotFound
	}
	delete(s.Events, evID)
	s.Count--
	if v.Compare(s.Highest) == 0 {
		s.Highest = NewDecimal(0, 0) // reset highest

		// Find new highest among remaining events.
		for _, val := range s.Events {
			if val.Compare(s.Highest) == 1 {
				s.Highest = val
			}
		}
	}
	return nil
}

// GetFilterIDs is part of StatMetric interface.
func (s *StatHighest) GetFilterIDs() []string {
	return s.FilterIDs
}

// GetMinItems returns the minimum items for the metric.
func (s *StatHighest) GetMinItems() uint64 { return s.MinItems }

// Compress is part of StatMetric interface.
func (s *StatHighest) Compress(_ uint64, _ string) []string {
	eventIDs := make([]string, 0, len(s.Events))
	for id := range s.Events {
		eventIDs = append(eventIDs, id)
	}
	return eventIDs
}

func (s *StatHighest) GetCompressFactor(events map[string]uint64) map[string]uint64 {
	for id := range s.Events {
		if _, exists := events[id]; !exists {
			events[id] = 1
		}
	}
	return events
}

// NewStatLowest creates a StatLowest metric for tracking minimum field values.
func NewStatLowest(minItems uint64, fieldName string, filterIDs []string) StatMetric {
	return &StatLowest{
		FilterIDs: filterIDs,
		MinItems:  minItems,
		FieldName: fieldName,
		Lowest:    NewDecimalFromFloat64(math.MaxFloat64),
		Events:    make(map[string]*Decimal),
	}
}

// StatLowest tracks the minimum value for a specific field across events.
type StatLowest struct {
	FilterIDs []string // event filters to apply before processing
	FieldName string   // field path to extract from events
	MinItems  uint64   // minimum events required for valid results

	Lowest *Decimal            // current minimum value tracked
	Count  uint64              // number of events currently tracked
	Events map[string]*Decimal // event values indexed by ID for deletion
}

// Clone creates a deep copy of StatLowest.
func (s *StatLowest) Clone() StatMetric {
	if s == nil {
		return nil
	}
	clone := &StatLowest{
		FilterIDs: slices.Clone(s.FilterIDs),
		Lowest:    s.Lowest,
		Count:     s.Count,
		MinItems:  s.MinItems,
		FieldName: s.FieldName,
		Events:    maps.Clone(s.Events),
	}
	return clone
}

func (s *StatLowest) GetStringValue(decimals int) string {
	if s.Count == 0 || s.Count < s.MinItems {
		return NotAvailable
	}
	v, _ := s.Lowest.Round(decimals).Float64()
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (s *StatLowest) GetValue() *Decimal {
	if s.Count == 0 || s.Count < s.MinItems {
		return DecimalNaN
	}
	return s.Lowest
}

// AddEvent processes a new event, updating lowest value if necessary.
func (s *StatLowest) AddEvent(evID string, ev DataProvider) error {
	val, err := fieldValueFromDP(s.FieldName, ev)
	if err != nil {
		return err
	}
	if val.Compare(s.Lowest) == -1 {
		s.Lowest = val
	}

	// Only increment count for new events.
	if _, exists := s.Events[evID]; !exists {
		s.Count++
	}

	s.Events[evID] = val
	return nil
}

// AddOneEvent processes event without storing for removal (used when events
// never expire).
func (s *StatLowest) AddOneEvent(ev DataProvider) error {
	val, err := fieldValueFromDP(s.FieldName, ev)
	if err != nil {
		return err
	}
	if val.Compare(s.Lowest) == -1 {
		s.Lowest = val
	}
	s.Count++
	return nil
}

func (s *StatLowest) RemEvent(evID string) error {
	v, exists := s.Events[evID]
	if !exists {
		return ErrNotFound
	}
	delete(s.Events, evID)
	s.Count--
	if v.Compare(s.Lowest) == 0 {
		s.Lowest = NewDecimalFromFloat64(math.MaxFloat64) // reset lowest

		// Find new lowest among remaining events.
		for _, val := range s.Events {
			if val.Compare(s.Lowest) == -1 {
				s.Lowest = val
			}
		}
	}
	return nil
}

// GetFilterIDs is part of StatMetric interface.
func (s *StatLowest) GetFilterIDs() []string {
	return s.FilterIDs
}

// GetMinItems returns the minimum items for the metric.
func (s *StatLowest) GetMinItems() uint64 { return s.MinItems }

// Compress is part of StatMetric interface.
func (s *StatLowest) Compress(_ uint64, _ string) []string {
	eventIDs := make([]string, 0, len(s.Events))
	for id := range s.Events {
		eventIDs = append(eventIDs, id)
	}
	return eventIDs
}

func (s *StatLowest) GetCompressFactor(events map[string]uint64) map[string]uint64 {
	for id := range s.Events {
		if _, exists := events[id]; !exists {
			events[id] = 1
		}
	}
	return events
}

// NewStatREPSC creates a StatREPSC metric for counting successful requests.
func NewStatREPSC(minItems uint64, _ string, filterIDs []string) StatMetric {
	return &StatREPSC{
		FilterIDs: filterIDs,
		MinItems:  minItems,
		Events:    make(map[string]struct{}),
	}
}

// StatREPSC counts requests where ReplyState equals "OK"
type StatREPSC struct {
	FilterIDs []string            // event filters to apply before processing
	MinItems  uint64              // minimum events required for valid results
	Count     uint64              // number of successful events tracked
	Events    map[string]struct{} // event IDs indexed for deletion
}

// Clone creates a deep copy of StatREPSC.
func (s *StatREPSC) Clone() StatMetric {
	if s == nil {
		return nil
	}
	clone := &StatREPSC{
		FilterIDs: slices.Clone(s.FilterIDs),
		MinItems:  s.MinItems,
		Count:     s.Count,
		Events:    maps.Clone(s.Events),
	}
	return clone
}

func (s *StatREPSC) GetStringValue(_ int) string {
	if s.Count == 0 || s.Count < s.MinItems {
		return NotAvailable
	}
	return strconv.Itoa(int(s.Count))
}

func (s *StatREPSC) GetValue() *Decimal {
	if s.Count == 0 || s.Count < s.MinItems {
		return DecimalNaN
	}
	return NewDecimal(int64(s.Count), 0)
}

// AddEvent processes a new event, incrementing count if ReplyState is "OK".
func (s *StatREPSC) AddEvent(evID string, ev DataProvider) error {
	replyState, err := replyStateFromDP(ev)
	if err != nil {
		return err
	}
	if replyState != OK {
		return nil
	}

	// Only increment count for new events.
	if _, exists := s.Events[evID]; !exists {
		s.Events[evID] = struct{}{}
		s.Count++
	}

	return nil
}

// AddOneEvent processes event without storing for removal (used when events
// never expire).
func (s *StatREPSC) AddOneEvent(ev DataProvider) error {
	replyState, err := replyStateFromDP(ev)
	if err != nil {
		return err
	}
	if replyState != OK {
		return nil
	}
	s.Count++
	return nil
}

func (s *StatREPSC) RemEvent(evID string) error {
	if _, exists := s.Events[evID]; !exists {
		return ErrNotFound
	}
	delete(s.Events, evID)
	s.Count--
	return nil
}

// GetFilterIDs is part of StatMetric interface.
func (s *StatREPSC) GetFilterIDs() []string {
	return s.FilterIDs
}

// GetMinItems returns the minimum items for the metric.
func (s *StatREPSC) GetMinItems() uint64 {
	return s.MinItems
}

// Compress is part of StatMetric interface.
func (s *StatREPSC) Compress(_ uint64, _ string) []string {
	eventIDs := make([]string, 0, len(s.Events))
	for id := range s.Events {
		eventIDs = append(eventIDs, id)
	}
	return eventIDs
}

func (s *StatREPSC) GetCompressFactor(events map[string]uint64) map[string]uint64 {
	for id := range s.Events {
		if _, exists := events[id]; !exists {
			events[id] = 1
		}
	}
	return events
}

// NewStatREPFC creates a StatREPFC metric for counting failed requests.
func NewStatREPFC(minItems uint64, errorType string, filterIDs []string) StatMetric {
	return &StatREPFC{
		FilterIDs: filterIDs,
		MinItems:  minItems,
		ErrorType: errorType,
		Events:    make(map[string]struct{}),
	}
}

// StatREPFC counts requests where ReplyState is not "OK".
type StatREPFC struct {
	FilterIDs []string            // event filters to apply before processing
	MinItems  uint64              // minimum events required for valid results
	ErrorType string              // specific error type to filter for (empty = all errors)
	Count     uint64              // number of failed events tracked
	Events    map[string]struct{} // event IDs indexed for deletion
}

// Clone creates a deep copy of StatREPFC.
func (s *StatREPFC) Clone() StatMetric {
	if s == nil {
		return nil
	}
	clone := &StatREPFC{
		FilterIDs: slices.Clone(s.FilterIDs),
		MinItems:  s.MinItems,
		ErrorType: s.ErrorType,
		Count:     s.Count,
		Events:    maps.Clone(s.Events),
	}
	return clone
}

func (s *StatREPFC) GetStringValue(_ int) string {
	if s.Count == 0 || s.Count < s.MinItems {
		return NotAvailable
	}
	return strconv.Itoa(int(s.Count))
}

func (s *StatREPFC) GetValue() *Decimal {
	if s.Count == 0 || s.Count < s.MinItems {
		return DecimalNaN
	}
	return NewDecimal(int64(s.Count), 0)
}

// AddEvent processes a new event, incrementing count if ReplyState is not "OK".
func (s *StatREPFC) AddEvent(evID string, ev DataProvider) error {
	replyState, err := replyStateFromDP(ev)
	if err != nil {
		return err
	}

	// Skip if success when counting all failures, or if not matching specific
	// error type.
	if s.ErrorType == "" && replyState == OK {
		return nil
	}
	// Handle multiple errors separated by ";" (e.g., "ERR_TERMINATE;ERR_CDRS")
	// Use split + exact match instead of strings.Contains to avoid false positives.
	if s.ErrorType != "" {
		errors := strings.Split(replyState, InfieldSep)
		if !slices.Contains(errors, s.ErrorType) {
			return nil
		}
	}

	// Only increment count for new events.
	if _, exists := s.Events[evID]; !exists {
		s.Events[evID] = struct{}{}
		s.Count++
	}

	return nil
}

// AddOneEvent processes event without storing for removal (used when events
// never expire).
func (s *StatREPFC) AddOneEvent(ev DataProvider) error {
	replyState, err := replyStateFromDP(ev)
	if err != nil {
		return err
	}

	// Skip if success when counting all failures, or if not matching specific
	// error type.
	if s.ErrorType == "" && replyState == OK {
		return nil
	}
	// Handle multiple errors separated by ";" (e.g., "ERR_TERMINATE;ERR_CDRS")
	// Use split + exact match instead of strings.Contains to avoid false positives
	if s.ErrorType != "" {
		errors := strings.Split(replyState, InfieldSep)
		if !slices.Contains(errors, s.ErrorType) {
			return nil
		}
	}

	s.Count++
	return nil
}

func (s *StatREPFC) RemEvent(evID string) error {
	if _, exists := s.Events[evID]; !exists {
		return ErrNotFound
	}
	delete(s.Events, evID)
	s.Count--
	return nil
}

// GetFilterIDs is part of StatMetric interface.
func (s *StatREPFC) GetFilterIDs() []string {
	return s.FilterIDs
}

// GetMinItems returns the minimum items for the metric.
func (s *StatREPFC) GetMinItems() uint64 {
	return s.MinItems
}

// Compress is part of StatMetric interface.
func (s *StatREPFC) Compress(_ uint64, _ string) []string {
	eventIDs := make([]string, 0, len(s.Events))
	for id := range s.Events {
		eventIDs = append(eventIDs, id)
	}
	return eventIDs
}

func (s *StatREPFC) GetCompressFactor(events map[string]uint64) map[string]uint64 {
	for id := range s.Events {
		if _, exists := events[id]; !exists {
			events[id] = 1
		}
	}
	return events
}

// fieldValueFromDP gets the numeric value from the DataProvider.
func fieldValueFromDP(fldName string, dp DataProvider) (*Decimal, error) {
	ival, err := DPDynamicInterface(fldName, dp)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrPrefix(err, fldName)
			// NOTE: return below might be clearer
			// return nil, fmt.Errorf("field %s: %v", field, err)
		}
		return nil, err
	}
	v, err := IfaceAsBig(ival)
	if err != nil {
		return nil, err
	}
	return &Decimal{Big: v}, nil
}

// replyStateFromDP gets the numeric value from the DataProvider.
func replyStateFromDP(dp DataProvider) (string, error) {
	ival, err := dp.FieldAsInterface([]string{MetaReq, ReplyState})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrPrefix(err, ReplyState)
			// NOTE: return below might be clearer
			// return 0, fmt.Errorf("field %s: %v", ReplyState, err)
		}
		return "", err
	}
	return IfaceAsString(ival), nil
}
