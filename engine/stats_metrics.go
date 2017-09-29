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
	"time"

	"github.com/cgrates/cgrates/utils"
)

type Metric interface {
	AddCdr(*QCdr)
	RemoveCdr(*QCdr)
	GetValue() float64
}

const ASR = "ASR"
const ACD = "ACD"
const TCD = "TCD"
const ACC = "ACC"
const TCC = "TCC"
const PDD = "PDD"
const DDC = "DDC"
const STATS_NA = -1.0

func CreateMetric(metric string) Metric {
	switch metric {
	case ASR:
		return &ASRMetric{}
	case PDD:
		return &PDDMetric{}
	case ACD:
		return &ACDMetric{}
	case TCD:
		return &TCDMetric{}
	case ACC:
		return &ACCMetric{}
	case TCC:
		return &TCCMetric{}
	case DDC:
		return NewDccMetric()
	}
	return nil
}

// ASR - Answer-Seizure Ratio
// successfully answered Calls divided by the total number of Calls attempted and multiplied by 100
type ASRMetric struct {
	answered float64
	count    float64
}

func (asr *ASRMetric) AddCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		asr.answered += 1
	}
	asr.count += 1
}

func (asr *ASRMetric) RemoveCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		asr.answered -= 1
	}
	asr.count -= 1
}

func (asr *ASRMetric) GetValue() float64 {
	if asr.count == 0 {
		return STATS_NA
	}
	val := asr.answered / asr.count * 100
	return utils.Round(val, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}

// PDD – Post Dial Delay (average)
// the sum of PDD seconds of total calls divided by the number of these calls.
type PDDMetric struct {
	sum   time.Duration
	count float64
}

func (PDD *PDDMetric) AddCdr(cdr *QCdr) {
	if cdr.Pdd == 0 { // Pdd not defined
		return
	}
	PDD.sum += cdr.Pdd
	PDD.count += 1
}

func (PDD *PDDMetric) RemoveCdr(cdr *QCdr) {
	if cdr.Pdd == 0 { // Pdd not defined
		return
	}
	PDD.sum -= cdr.Pdd
	PDD.count -= 1
}

func (PDD *PDDMetric) GetValue() float64 {
	if PDD.count == 0 {
		return STATS_NA
	}
	val := PDD.sum.Seconds() / PDD.count
	return utils.Round(val, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}

// ACD – Average Call Duration
// the sum of billable seconds (billsec) of answered calls divided by the number of these answered calls.
type ACDMetric struct {
	sum   time.Duration
	count float64
}

func (acd *ACDMetric) AddCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		acd.sum += cdr.Usage
		acd.count += 1
	}
}

func (acd *ACDMetric) RemoveCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		acd.sum -= cdr.Usage
		acd.count -= 1
	}
}

func (acd *ACDMetric) GetValue() float64 {
	if acd.count == 0 {
		return STATS_NA
	}
	val := acd.sum.Seconds() / acd.count
	return utils.Round(val, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}

// TCD – Total Call Duration
// the sum of billable seconds (billsec) of answered calls
type TCDMetric struct {
	sum   time.Duration
	count float64
}

func (tcd *TCDMetric) AddCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		tcd.sum += cdr.Usage
		tcd.count += 1
	}
}

func (tcd *TCDMetric) RemoveCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		tcd.sum -= cdr.Usage
		tcd.count -= 1
	}
}

func (tcd *TCDMetric) GetValue() float64 {
	if tcd.count == 0 {
		return STATS_NA
	}
	return utils.Round(tcd.sum.Seconds(), globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}

// ACC – Average Call Cost
// the sum of cost of answered calls divided by the number of these answered calls.
type ACCMetric struct {
	sum   float64
	count float64
}

func (acc *ACCMetric) AddCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost >= 0 {
		acc.sum += cdr.Cost
		acc.count += 1
	}
}

func (acc *ACCMetric) RemoveCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost >= 0 {
		acc.sum -= cdr.Cost
		acc.count -= 1
	}
}

func (acc *ACCMetric) GetValue() float64 {
	if acc.count == 0 {
		return STATS_NA
	}
	val := acc.sum / acc.count
	return utils.Round(val, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}

// TCC – Total Call Cost
// the sum of cost of answered calls
type TCCMetric struct {
	sum   float64
	count float64
}

func (tcc *TCCMetric) AddCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost >= 0 {
		tcc.sum += cdr.Cost
		tcc.count += 1
	}
}

func (tcc *TCCMetric) RemoveCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() && cdr.Cost >= 0 {
		tcc.sum -= cdr.Cost
		tcc.count -= 1
	}
}

func (tcc *TCCMetric) GetValue() float64 {
	if tcc.count == 0 {
		return STATS_NA
	}
	return utils.Round(tcc.sum, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}

// DDC - Destination Distinct Count
//
type DCCMetric struct {
	destinations map[string]int64
}

func NewDccMetric() *DCCMetric {
	return &DCCMetric{
		destinations: make(map[string]int64),
	}
}

func (dcc *DCCMetric) AddCdr(cdr *QCdr) {
	if count, exists := dcc.destinations[cdr.Dest]; exists {
		dcc.destinations[cdr.Dest] = count + 1
	} else {
		dcc.destinations[cdr.Dest] = 0
	}
}

func (dcc *DCCMetric) RemoveCdr(cdr *QCdr) {
	if count, exists := dcc.destinations[cdr.Dest]; exists && count > 1 {
		dcc.destinations[cdr.Dest] = count - 1
	} else {
		dcc.destinations[cdr.Dest] = 0
	}
}

func (dcc *DCCMetric) GetValue() float64 {
	if len(dcc.destinations) == 0 {
		return STATS_NA
	}
	return float64(len(dcc.destinations))
}
