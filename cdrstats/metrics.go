/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

package cdrstats

type Metric interface {
	AddCDR(*QCDR)
	RemoveCDR(*QCDR)
}

func CreateMetric(metric string) *Metric {
	switch metric {
	case "ASR":
		return &ASRMetric{}
	case "ACD":
		return &ACDMetric{}
	case "ACC":
		return &ACCMetric{}
	}
	return nil
}

// ASR - Answer-Seizure Ratio
// successfully answered Calls divided by the total number of Calls attempted and multiplied by 100
type ASRMetric struct {
	sum   float64
	count int64
}

func (asr *ASRMetric) AddCdr(cdr *QCDR) {
}

// ACD – Average Call Duration
// the sum of billable seconds (billsec) of answered calls divided by the number of these answered calls.
type ACDMetric struct {
	sum   float64
	count int64
}

// ACC – Average Call Cost
// the sum of cost of answered calls divided by the number of these answered calls.
type ACCMetric struct {
	sum   float64
	count int64
}
