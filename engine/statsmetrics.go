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

	"github.com/cgrates/cgrates/utils"
)

// NewStatsMetrics instantiates the StatsMetrics
func NewStatsMetric(metricID string) (sm StatsMetric, err error) {
	metrics := map[string]func() (StatsMetric, error){
		utils.MetaASR: NewStatsASR,
		utils.MetaACD: NewStatsACD,
	}
	if _, has := metrics[metricID]; !has {
		return nil, fmt.Errorf("unsupported metric: %s", metricID)
	}
	return metrics[metricID]()
}

// StatsMetric is the interface which a metric should implement
type StatsMetric interface {
	GetStringValue(fmtOpts string) (val string)
	AddEvent(ev StatsEvent) error
	RemEvent(ev StatsEvent) error
	GetMarshaled(ms Marshaler) (vals []byte, err error)
	SetFromMarshaled(vals []byte, ms Marshaler) (err error) // mostly used to load from DB
}

func NewStatsASR() (StatsMetric, error) {
	return new(StatsASR), nil
}

// StatsASR implements AverageSuccessRatio metric
type StatsASR struct {
	answered int
	count    int
}

func (asr *StatsASR) GetStringValue(fmtOpts string) (val string) {
	return
}

func (asr *StatsASR) AddEvent(ev StatsEvent) (err error) {
	return
}

func (asr *StatsASR) RemEvent(ev StatsEvent) (err error) {
	return
}

func (asr *StatsASR) GetMarshaled(ms Marshaler) (vals []byte, err error) {
	return
}

func (asr *StatsASR) SetFromMarshaled(vals []byte, ms Marshaler) (err error) {
	return
}

func NewStatsACD() (StatsMetric, error) {
	return new(StatsACD), nil
}

// StatsACD implements AverageCallDuration metric
type StatsACD struct{}

func (acd *StatsACD) GetStringValue(fmtOpts string) (val string) {
	return
}

func (acd *StatsACD) AddEvent(ev StatsEvent) (err error) {
	return
}

func (acd *StatsACD) RemEvent(ev StatsEvent) (err error) {
	return
}

func (acd *StatsACD) GetMarshaled(ms Marshaler) (vals []byte, err error) {
	return
}

func (acd *StatsACD) SetFromMarshaled(vals []byte, ms Marshaler) (err error) {
	return
}
