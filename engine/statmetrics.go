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
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// NewStatMetric instantiates the StatMetric
// cfg serves as general purpose container to pass config options to metric
func NewStatMetric(metricID string, minItems uint64, filterIDs []string) (sm *StatMetricWithFilters, err error) {
	metrics := map[string]func(uint64, string) StatMetric{
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
	return &StatMetricWithFilters{StatMetric: metrics[metricSplit[0]](minItems, extraParams), FilterIDs: filterIDs}, nil
}

// StatMetric is the interface which a metric should implement
type StatMetric interface {
	GetValue() *utils.Decimal
	GetStringValue() string
	AddEvent(evID string, ev utils.DataProvider) error
	RemEvent(evID string) error
	GetMinItems() (minIts uint64)
	Compress(queueLen uint64, defaultID string) (eventIDs []string)
	GetCompressFactor(events map[string]uint64) map[string]uint64
	Clone() StatMetric
}

func NewASR(minItems uint64, _ string) StatMetric {
	return &StatASR{Metric: NewMetric(minItems)}
}

// ASR implements AverageSuccessRatio metric
type StatASR struct {
	*Metric
}

func (asr *StatASR) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := asr.getAvgValue(); val != utils.DecimalNaN {
		valStr = utils.MultiplyDecimal(val, utils.NewDecimal(100, 0)).String() + "%"
	}
	return
}

func (asr *StatASR) GetValue() *utils.Decimal {
	return utils.MultiplyDecimal(asr.getAvgValue(), utils.NewDecimal(100, 0))
}

// AddEvent is part of StatMetric interface
func (asr *StatASR) AddEvent(evID string, ev utils.DataProvider) error {
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
	return asr.addEvent(evID, answered)
}

func (asr *StatASR) Clone() StatMetric {
	return &StatASR{
		Metric: asr.Metric.Clone(),
	}
}

func NewACD(minItems uint64, _ string) StatMetric {
	return &StatACD{Metric: NewMetric(minItems)}
}

// ACD implements AverageCallDuration metric
type StatACD struct {
	*Metric
}

func (acd *StatACD) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := acd.getAvgValue(); val != utils.DecimalNaN {
		dur, _ := val.Duration()
		valStr = dur.String()
	}
	return
}

func (acd *StatACD) GetValue() *utils.Decimal {
	return acd.getAvgValue()
}

func (acd *StatACD) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaReq, utils.Usage})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Usage)
		}
		return err
	}
	return acd.addEvent(evID, ival)
}

func (acd *StatACD) Clone() StatMetric {
	return &StatAverage{
		Metric: acd.Metric.Clone(),
	}
}

func NewTCD(minItems uint64, _ string) StatMetric {
	return &StatTCD{Metric: NewMetric(minItems)}
}

// TCD implements TotalCallDuration metric
type StatTCD struct {
	*Metric
}

func (sum *StatTCD) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := sum.getTotalValue(); val != utils.DecimalNaN {
		dur, _ := val.Duration()
		valStr = dur.String()
	}
	return
}

func (sum *StatTCD) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaReq, utils.Usage})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Usage)
		}
		return err
	}
	return sum.addEvent(evID, ival)
}

func (sum *StatTCD) Clone() StatMetric {
	return &StatTCD{
		Metric: sum.Metric.Clone(),
	}
}

func NewACC(minItems uint64, _ string) StatMetric {
	return &StatACC{Metric: NewMetric(minItems)}
}

// ACC implements AverageCallCost metric
type StatACC struct {
	*Metric
}

func (acc *StatACC) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := acc.getAvgValue(); val != utils.DecimalNaN {
		valStr = val.String()
	}
	return
}

func (acc *StatACC) GetValue() *utils.Decimal {
	return acc.getAvgValue()
}

func (acc *StatACC) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaReq, utils.Cost})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Cost)
		}
		return err
	}
	return acc.addEvent(evID, ival)
}

func (acc *StatACC) Clone() StatMetric {
	return &StatACC{
		Metric: acc.Metric.Clone(),
	}
}

func NewTCC(minItems uint64, _ string) StatMetric {
	return &StatTCC{Metric: NewMetric(minItems)}
}

// TCC implements TotalCallCost metric
type StatTCC struct {
	*Metric
}

func (tcc *StatTCC) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := tcc.getTotalValue(); val != utils.DecimalNaN {
		valStr = val.String()
	}
	return
}

func (tcc *StatTCC) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaReq, utils.Cost})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.Cost)
		}
		return err
	}
	return tcc.addEvent(evID, ival)
}

func (tcc *StatTCC) Clone() StatMetric {
	return &StatTCC{
		Metric: tcc.Metric.Clone(),
	}
}

func NewPDD(minItems uint64, _ string) StatMetric {
	return &StatPDD{Metric: NewMetric(minItems)}
}

// PDD implements Post Dial Delay (average) metric
type StatPDD struct {
	*Metric
}

func (pdd *StatPDD) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := pdd.getAvgValue(); val != utils.DecimalNaN {
		dur, _ := val.Duration()
		valStr = dur.String()
	}
	return
}

func (pdd *StatPDD) GetValue() *utils.Decimal {
	return pdd.getAvgValue()
}

func (pdd *StatPDD) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaReq, utils.PDD})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.PDD)
		}
		return err
	}
	return pdd.addEvent(evID, ival)
}

func (pdd *StatPDD) Clone() StatMetric {
	return &StatPDD{
		Metric: pdd.Metric.Clone(),
	}
}

func NewDDC(minItems uint64, _ string) StatMetric {
	return &StatDDC{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    minItems,
	}
}

type StatDDC struct {
	FieldValues map[string]utils.StringSet   // map[fieldValue]map[eventID]
	Events      map[string]map[string]uint64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    uint64
	Count       uint64
}

// getValue returns tcd.val
func (ddc *StatDDC) getValue(roundingDecimal int) float64 {
	if ddc.Count == 0 || ddc.Count < ddc.MinItems {
		return utils.StatsNA
	}
	return float64(len(ddc.FieldValues))
}

func (ddc *StatDDC) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := ddc.GetValue(); val != utils.DecimalNaN {
		valStr = val.String()
	}
	return
}

func (ddc *StatDDC) GetValue() *utils.Decimal {
	if ddc.Count == 0 || ddc.Count < ddc.MinItems {
		return utils.DecimalNaN
	}
	return utils.NewDecimal(int64(len(ddc.FieldValues)), 0)
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
	cln := &StatDDC{
		FieldValues: make(map[string]utils.StringSet),
		Count:       ddc.Count,
		Events:      make(map[string]map[string]uint64),
		MinItems:    ddc.MinItems,
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

////////////////////////////////////
type StatMetricWithFilters struct {
	StatMetric
	FilterIDs []string
}

func (sm *StatMetricWithFilters) Clone() *StatMetricWithFilters {
	return &StatMetricWithFilters{
		StatMetric: sm.StatMetric.Clone(),
		FilterIDs:  utils.CloneStringSlice(sm.FilterIDs),
	}
}

// ACDHelper structure
type DecimalWithCompress struct {
	Stat           *utils.Decimal
	CompressFactor uint64
}

func NewMetric(minItems uint64) *Metric {
	return &Metric{
		Value:    utils.NewDecimal(0, 0),
		Events:   make(map[string]*DecimalWithCompress),
		MinItems: minItems,
	}
}

type Metric struct {
	Value    *utils.Decimal
	Count    uint64
	Events   map[string]*DecimalWithCompress // map[EventTenantID]Cost
	MinItems uint64
}

func (sum *Metric) getTotalValue() *utils.Decimal {
	if len(sum.Events) == 0 || sum.Count < sum.MinItems {
		return utils.DecimalNaN
	}
	return sum.Value
}

func (sum *Metric) getAvgValue() *utils.Decimal {
	if len(sum.Events) == 0 || sum.Count < sum.MinItems {
		return utils.DecimalNaN
	}
	return utils.DivideDecimal(sum.Value, utils.NewDecimal(int64(sum.Count), 0))
}

func (sum *Metric) GetValue() (v *utils.Decimal) {
	return sum.getTotalValue()
}

func (sum *Metric) addEvent(evID string, ival interface{}) (err error) {
	var val *decimal.Big
	if val, err = utils.IfaceAsBig(ival); err != nil {
		return
	}
	dVal := &utils.Decimal{val}
	sum.Value = utils.SumDecimal(sum.Value, dVal)
	if v, has := sum.Events[evID]; !has {
		sum.Events[evID] = &DecimalWithCompress{Stat: dVal, CompressFactor: 1}
	} else {
		v.Stat = utils.DivideDecimal(
			utils.MultiplyDecimal(v.Stat, utils.NewDecimal(int64(v.CompressFactor), 0)),
			utils.NewDecimal(int64(v.CompressFactor)+1, 0))
		v.CompressFactor = v.CompressFactor + 1
	}
	sum.Count++
	return
}

func (sum *Metric) RemEvent(evID string) (err error) {
	val, has := sum.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	if val.Stat.Compare(utils.NewDecimal(0, 0)) != 0 {
		sum.Value = utils.SubstractDecimal(sum.Value, val.Stat)
	}
	sum.Count--
	if val.CompressFactor <= 1 {
		delete(sum.Events, evID)
	} else {
		val.CompressFactor = val.CompressFactor - 1
	}
	return
}

// GetMinItems returns the minim items for the metric
func (sum *Metric) GetMinItems() uint64 { return sum.MinItems }

// Compress is part of StatMetric interface
func (sum *Metric) Compress(queueLen uint64, defaultID string) (eventIDs []string) {
	if sum.Count < queueLen {
		eventIDs = make([]string, 0, len(sum.Events))
		for id := range sum.Events {
			eventIDs = append(eventIDs, id)
		}
		return
	}
	sum.Events = map[string]*DecimalWithCompress{defaultID: {
		Stat:           utils.DivideDecimal(sum.Value, utils.NewDecimalFromFloat64(float64(sum.Count))),
		CompressFactor: sum.Count,
	}}
	return []string{defaultID}
}

// Compress is part of StatMetric interface
func (sum *Metric) GetCompressFactor(events map[string]uint64) map[string]uint64 {
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

func (sum *Metric) Clone() (cln *Metric) {
	cln = &Metric{
		Value:    sum.Value.Clone(),
		Count:    sum.Count,
		Events:   make(map[string]*DecimalWithCompress),
		MinItems: sum.MinItems,
	}
	for k, v := range sum.Events {
		cln.Events[k] = &(*v)
	}
	return
}

func NewStatSum(minItems uint64, fieldName string) StatMetric {
	return &StatSum{Metric: NewMetric(minItems),
		FieldName: fieldName}
}

type StatSum struct {
	*Metric
	FieldName string
}

func (sum *StatSum) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := sum.getTotalValue(); val != utils.DecimalNaN {
		valStr = val.String()
	}
	return
}

func (sum *StatSum) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := utils.DPDynamicInterface(sum.FieldName, ev)
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, sum.FieldName)
		}
		return err
	}
	return sum.addEvent(evID, ival)
}

func (sum *StatSum) Clone() StatMetric {
	return &StatSum{
		Metric:    sum.Metric.Clone(),
		FieldName: sum.FieldName,
	}
}

func NewStatAverage(minItems uint64, fieldName string) StatMetric {
	return &StatAverage{Metric: NewMetric(minItems),
		FieldName: fieldName}
}

// StatAverage implements TotalCallCost metric
type StatAverage struct {
	*Metric
	FieldName string
}

func (avg *StatAverage) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := avg.getAvgValue(); val != utils.DecimalNaN {
		valStr = val.String()
	}
	return
}

func (avg *StatAverage) GetValue() *utils.Decimal {
	return avg.getAvgValue()
}

func (avg *StatAverage) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := utils.DPDynamicInterface(avg.FieldName, ev)
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, avg.FieldName)
		}
		return err
	}
	return avg.addEvent(evID, ival)
}

func (avg *StatAverage) Clone() StatMetric {
	return &StatAverage{
		Metric:    avg.Metric.Clone(),
		FieldName: avg.FieldName,
	}
}

func NewStatDistinct(minItems uint64, fieldName string) StatMetric {
	return &StatDistinct{
		Events:      make(map[string]map[string]uint64),
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    minItems,
		FieldName:   fieldName,
	}
}

type StatDistinct struct {
	FieldValues map[string]utils.StringSet   // map[fieldValue]map[eventID]
	Events      map[string]map[string]uint64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    uint64
	FieldName   string
	Count       uint64
}

// getValue returns tcd.val
func (dst *StatDistinct) getValue(roundingDecimal int) float64 {
	if dst.Count == 0 || dst.Count < dst.MinItems {
		return utils.StatsNA
	}
	return float64(len(dst.FieldValues))
}

func (dst *StatDistinct) GetStringValue() (valStr string) {
	valStr = utils.NotAvailable
	if val := dst.GetValue(); val != utils.DecimalNaN {
		valStr = val.String()
	}
	return
}

func (dst *StatDistinct) GetValue() *utils.Decimal {
	if dst.Count == 0 || dst.Count < dst.MinItems {
		return utils.DecimalNaN
	}
	return utils.NewDecimal(int64(len(dst.FieldValues)), 0)
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
	cln := &StatDistinct{
		Count:       dst.Count,
		Events:      make(map[string]map[string]uint64),
		MinItems:    dst.MinItems,
		FieldName:   dst.FieldName,
		FieldValues: make(map[string]utils.StringSet),
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
