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
	getStringValue() string
	addEvent(ev StatsEvent) error
	remEvent(ev StatsEvent) error
	getStoredValues() ([]byte, error) // used to generate the values which are stored into DB
	loadStoredValues([]byte) error    // load the values from DB data
}

func NewStatsASR() (StatsMetric, error) {
	return new(StatsASR), nil
}

// StatsASR implements AverageSuccessRatio metric
type StatsASR struct {
	answered int
	count    int
}

func (asr *StatsASR) getStringValue() (val string) {
	return
}

func (asr *StatsASR) addEvent(ev StatsEvent) (err error) {
	return
}

func (asr *StatsASR) remEvent(ev StatsEvent) (err error) {
	return
}

func (asr *StatsASR) getStoredValues() (vals []byte, err error) {
	return
}

func (asr *StatsASR) loadStoredValues(vals []byte) (err error) {
	return
}

func NewStatsACD() (StatsMetric, error) {
	return new(StatsACD), nil
}

// StatsACD implements AverageCallDuration metric
type StatsACD struct{}

func (acd *StatsACD) getStringValue() (val string) {
	return
}

func (acd *StatsACD) addEvent(ev StatsEvent) (err error) {
	return
}

func (acd *StatsACD) remEvent(ev StatsEvent) (err error) {
	return
}

func (asr *StatsACD) getStoredValues() (vals []byte, err error) {
	return
}

func (asr *StatsACD) loadStoredValues(vals []byte) (err error) {
	return
}
