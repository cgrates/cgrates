/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
const ACC = "ACC"

func CreateMetric(metric string) Metric {
	switch metric {
	case ASR:
		return &ASRMetric{}
	case ACD:
		return &ACDMetric{}
	case ACC:
		return &ACCMetric{}
	}
	return nil
}

// ASR - Answer-Seizure Ratio
// successfully answered Calls divided by the total number of Calls attempted and multiplied by 100
type ASRMetric struct {
	answered float64
	total    float64
}

func (asr *ASRMetric) AddCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		asr.answered += 1
	}
	asr.total += 1
}

func (asr *ASRMetric) RemoveCdr(cdr *QCdr) {
	if !cdr.AnswerTime.IsZero() {
		asr.answered -= 1
	}
	asr.total -= 1
}

func (asr *ASRMetric) GetValue() float64 {
	if asr.total == 0 {
		return 0
	}
	val := asr.answered / asr.total * 100
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
		return 0
	}
	val := acd.sum.Seconds() / acd.count
	return utils.Round(val, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
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
		return 0
	}
	val := acc.sum / acc.count
	return utils.Round(val, globalRoundingDecimals, utils.ROUNDING_MIDDLE)
}
