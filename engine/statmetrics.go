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
	"slices"
	"strconv"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// NewStatMetric instantiates the StatMetric
// cfg serves as general purpose container to pass config options to metric
func NewStatMetric(metricID string, minItems uint64, filterIDs []string) (sm StatMetric, err error) {
	metrics := map[string]func(uint64, string, []string) StatMetric{
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
	return metrics[metricSplit[0]](minItems, extraParams, filterIDs), nil
}

// StatMetric is the interface which a metric should implement
type StatMetric interface {
	GetValue() *utils.Decimal
	GetStringValue(rounding int) string
	AddOneEvent(ev utils.DataProvider) error
	AddEvent(evID string, ev utils.DataProvider) error
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

// ASR implements AverageSuccessRatio metric
type StatASR struct {
	*Metric
}

func (asr *StatASR) GetStringValue(rounding int) (valStr string) {
	valStr = utils.NotAvailable
	if val := asr.getAvgValue(); val != utils.DecimalNaN {
		v, _ := utils.MultiplyDecimal(val, utils.NewDecimal(100, 0)).Round(rounding).Float64()
		valStr = strconv.FormatFloat(v, 'f', -1, 64) + "%"
	}
	return
}

func (asr *StatASR) GetValue() (val *utils.Decimal) {
	if val = asr.getAvgValue(); val != utils.DecimalNaN {
		val = utils.MultiplyDecimal(val, utils.NewDecimal(100, 0))
	}
	return
}

func (asr *StatASR) AddOneEvent(ev utils.DataProvider) (err error) {
	var (
		answered int64
		val      any
	)
	if val, err = ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaStartTime}); err != nil {
		if err != utils.ErrNotFound {
			return
		}
	} else if at, err := utils.IfaceAsTime(val, config.CgrConfig().GeneralCfg().DefaultTimezone); err != nil {
		return err
	} else if !at.IsZero() {
		answered = 1
	}

	return asr.addOneEvent(answered)
}

// AddEvent is part of StatMetric interface
func (asr *StatASR) AddEvent(evID string, ev utils.DataProvider) (err error) {
	var answered int
	var val any
	if val, err = ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaStartTime}); err != nil {
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

func (asr *StatASR) RemEvent(evID string) (err error) {
	val, has := asr.Events[evID]
	if !has {
		return utils.ErrNotFound
	}
	ans := utils.NewDecimal(0, 0)
	if val.Stat.Compare(utils.NewDecimalFromFloat64(0.5)) > 0 {
		ans := utils.NewDecimal(1, 0)
		asr.Value = utils.SubstractDecimal(asr.Value, ans)
	}
	asr.Count--
	if val.CompressFactor <= 1 {
		delete(asr.Events, evID)
	} else {
		val.Stat = utils.DivideDecimal(
			utils.SubstractDecimal(
				utils.MultiplyDecimal(val.Stat, utils.NewDecimal(int64(val.CompressFactor), 0)),
				ans),
			utils.NewDecimal(int64(val.CompressFactor)-1, 0))
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
		return utils.NotAvailable
	}
	v, _ := acd.getAvgValue().Round(rounding).Duration()
	return v.String()
}

func (acd *StatACD) GetValue() *utils.Decimal {
	return acd.getAvgValue()
}

func (acd *StatACD) AddEvent(evID string, ev utils.DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaUsage})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaUsage)
		}
		return err
	}
	return acd.addEvent(evID, ival)
}

func (acd *StatACD) AddOneEvent(ev utils.DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaUsage})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaUsage)
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
		return utils.NotAvailable
	}
	v, _ := sum.Value.Round(rounding).Duration()
	return v.String()
}

func (sum *StatTCD) AddEvent(evID string, ev utils.DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaUsage})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaUsage)
		}
		return err
	}
	return sum.addEvent(evID, ival)
}

func (sum *StatTCD) AddOneEvent(ev utils.DataProvider) (err error) {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaUsage})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaUsage)
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

func (acc *StatACC) GetValue() *utils.Decimal {
	return acc.getAvgValue()
}

func (acc *StatACC) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaCost})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaCost)
		}
		return err
	}
	val, err := utils.IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return utils.ErrPrefix(utils.ErrNegative, utils.MetaCost)
	}
	return acc.addEvent(evID, val)
}

func (acc *StatACC) AddOneEvent(ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaCost})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaCost)
		}
		return err
	}
	val, err := utils.IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return utils.ErrPrefix(utils.ErrNegative, utils.MetaCost)
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

func (tcc *StatTCC) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaCost})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaCost)
		}
		return err
	}
	val, err := utils.IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return utils.ErrPrefix(utils.ErrNegative, utils.MetaCost)
	}
	return tcc.addEvent(evID, val)
}

func (tcc *StatTCC) AddOneEvent(ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaCost})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaCost)
		}
		return err
	}
	val, err := utils.IfaceAsBig(ival)
	if err != nil {
		return err
	}
	if val.Cmp(decimal.New(0, 0)) < 0 {
		return utils.ErrPrefix(utils.ErrNegative, utils.MetaCost)
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
		return utils.NotAvailable
	}
	v, _ := pdd.getAvgValue().Round(rounding).Duration()
	return v.String()
}

func (pdd *StatPDD) GetValue() *utils.Decimal {
	return pdd.getAvgValue()
}

func (pdd *StatPDD) AddEvent(evID string, ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaPDD})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaPDD)
		}
		return err
	}
	return pdd.addEvent(evID, ival)
}

func (pdd *StatPDD) AddOneEvent(ev utils.DataProvider) error {
	ival, err := ev.FieldAsInterface([]string{utils.MetaOpts, utils.MetaPDD})
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaPDD)
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
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    minItems,
		FilterIDs:   filterIDs,
	}
}

// StatDDC count  values occurring in destination field
type StatDDC struct {
	FieldValues map[string]utils.StringSet   // map[fieldValue]map[eventID]
	Events      map[string]map[string]uint64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    uint64
	Count       uint64
	FilterIDs   []string
}

func (ddc *StatDDC) GetFilterIDs() []string { return ddc.FilterIDs }

func (ddc *StatDDC) GetStringValue(rounding int) (valStr string) {
	valStr = utils.NotAvailable
	if val := ddc.GetValue(); val != utils.DecimalNaN {
		v, _ := val.Round(rounding).Float64()
		valStr = strconv.FormatFloat(v, 'f', -1, 64)
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
	if fieldValue, err = ev.FieldAsString([]string{utils.MetaOpts, utils.MetaDestination}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaDestination)
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

func (ddc *StatDDC) AddOneEvent(ev utils.DataProvider) (err error) {
	var fieldValue string
	if fieldValue, err = ev.FieldAsString([]string{utils.MetaOpts, utils.MetaDestination}); err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, utils.MetaDestination)
		}
		return
	}
	if _, has := ddc.FieldValues[fieldValue]; !has {
		ddc.FieldValues[fieldValue] = make(utils.StringSet)
	}
	ddc.Count++
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
	Stat           *utils.Decimal
	CompressFactor uint64
}

func NewMetric(minItems uint64, filterIDs []string) *Metric {
	return &Metric{
		Value:     utils.NewDecimal(0, 0),
		Events:    make(map[string]*DecimalWithCompress),
		MinItems:  minItems,
		FilterIDs: filterIDs,
	}
}

type Metric struct {
	Value     *utils.Decimal
	Count     uint64
	Events    map[string]*DecimalWithCompress // map[EventTenantID]Cost
	MinItems  uint64
	FilterIDs []string
}

func (sum *Metric) GetFilterIDs() []string { return sum.FilterIDs }

func (sum *Metric) getTotalValue() *utils.Decimal {
	if sum.Count == 0 || sum.Count < sum.MinItems {
		return utils.DecimalNaN
	}
	return sum.Value
}

func (sum *Metric) getAvgValue() *utils.Decimal {
	if sum.Count == 0 || sum.Count < sum.MinItems {
		return utils.DecimalNaN
	}
	return utils.DivideDecimal(sum.Value, utils.NewDecimal(int64(sum.Count), 0))
}

func (sum *Metric) getAvgStringValue(rounding int) string {
	if sum.Count == 0 || sum.Count < sum.MinItems {
		return utils.NotAvailable
	}
	v, _ := utils.DivideDecimal(sum.Value, utils.NewDecimal(int64(sum.Count), 0)).Round(rounding).Float64()
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (sum *Metric) GetStringValue(rounding int) string {
	if sum.Count == 0 || sum.Count < sum.MinItems {
		return utils.NotAvailable
	}
	v, _ := sum.Value.Round(rounding).Float64()
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (sum *Metric) GetValue() (v *utils.Decimal) {
	return sum.getTotalValue()
}

func (sum *Metric) addEvent(evID string, ival any) (err error) {
	var val *decimal.Big
	if val, err = utils.IfaceAsBig(ival); err != nil {
		return
	}
	dVal := &utils.Decimal{Big: val}
	sum.Value = utils.SumDecimal(sum.Value, dVal)
	if v, has := sum.Events[evID]; !has {
		sum.Events[evID] = &DecimalWithCompress{Stat: dVal, CompressFactor: 1}
	} else {
		v.Stat = utils.DivideDecimal(
			utils.SumDecimal(
				utils.MultiplyDecimal(v.Stat, utils.NewDecimal(int64(v.CompressFactor), 0)),
				dVal),
			utils.NewDecimal(int64(v.CompressFactor)+1, 0))
		v.CompressFactor = v.CompressFactor + 1
	}
	sum.Count++
	return
}

// Adding aggregated metrics without events
func (sum *Metric) addOneEvent(ival any) (err error) {
	var val *decimal.Big
	if val, err = utils.IfaceAsBig(ival); err != nil {
		return
	}
	dVal := &utils.Decimal{Big: val}
	sum.Value = utils.SumDecimal(sum.Value, dVal)
	sum.Count++
	return
}

// Deleting a specific event and updating metrics
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
		Value:     sum.Value.Clone(),
		Count:     sum.Count,
		Events:    make(map[string]*DecimalWithCompress),
		MinItems:  sum.MinItems,
		FilterIDs: slices.Clone(sum.FilterIDs),
	}
	for k, v := range sum.Events {
		cln.Events[k] = v
	}
	return
}

func (sum *Metric) Equal(v *Metric) bool {
	if sum.MinItems != v.MinItems ||
		sum.Count != v.Count ||
		sum.Value.Compare(v.Value) != 0 ||
		len(sum.Events) != len(v.Events) {
		return false
	}
	for k, c1 := range sum.Events {
		c2, has := v.Events[k]
		if !has ||
			c1.CompressFactor != c2.CompressFactor ||
			c1.Stat.Compare(c2.Stat) != 0 {
			return false
		}
	}
	return true
}

func NewStatSum(minItems uint64, fieldName string, filterIDs []string) StatMetric {
	return &StatSum{Metric: NewMetric(minItems, filterIDs),
		FieldName: fieldName}
}

type StatSum struct {
	*Metric
	FieldName string
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
func (sum *StatSum) AddOneEvent(ev utils.DataProvider) error {
	ival, err := utils.DPDynamicInterface(sum.FieldName, ev)
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, sum.FieldName)
		}
		return err
	}

	return sum.addOneEvent(ival)
}

func (sum *StatSum) Clone() StatMetric {
	return &StatSum{
		Metric:    sum.Metric.Clone(),
		FieldName: sum.FieldName,
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

func (avg *StatAverage) AddOneEvent(ev utils.DataProvider) error {
	ival, err := utils.DPDynamicInterface(avg.FieldName, ev)
	if err != nil {
		if err == utils.ErrNotFound {
			err = utils.ErrPrefix(err, avg.FieldName)
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
		FieldValues: make(map[string]utils.StringSet),
		MinItems:    minItems,
		FieldName:   fieldName,
		FilterIDs:   filterIDs,
	}
}

type StatDistinct struct {
	FieldValues map[string]utils.StringSet   // map[fieldValue]map[eventID]
	Events      map[string]map[string]uint64 // map[EventTenantID]map[fieldValue]compressfactor
	MinItems    uint64
	FieldName   string
	Count       uint64
	FilterIDs   []string
}

func (dst *StatDistinct) GetFilterIDs() []string { return dst.FilterIDs }

func (dst *StatDistinct) GetStringValue(rounding int) (valStr string) {
	valStr = utils.NotAvailable
	if val := dst.GetValue(); val != utils.DecimalNaN {
		v, _ := val.Round(rounding).Float64()
		valStr = strconv.FormatFloat(v, 'f', -1, 64)
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
	// simply remove the ~*req./~*opts. prefix and do normal process
	if !strings.HasPrefix(dst.FieldName, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep) && !strings.HasPrefix(dst.FieldName, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep) {
		return fmt.Errorf("invalid format for field <%s>", dst.FieldName)
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

func (dst *StatDistinct) AddOneEvent(ev utils.DataProvider) (err error) {
	var fieldValue string
	// simply remove the ~*req./~*opts. prefix and do normal process
	if !strings.HasPrefix(dst.FieldName, utils.DynamicDataPrefix+utils.MetaReq+utils.NestingSep) && !strings.HasPrefix(dst.FieldName, utils.DynamicDataPrefix+utils.MetaOpts+utils.NestingSep) {
		return fmt.Errorf("invalid format for field <%s>", dst.FieldName)
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

	dst.Count++
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
