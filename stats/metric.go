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

package stats

import (
	"fmt"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewStatsMetrics instantiates the StatsMetrics
// cfg serves as general purpose container to pass config options to metric
func NewStatsMetric(metricID string) (sm StatsMetric, err error) {
	metrics := map[string]func() (StatsMetric, error){
		utils.MetaASR: NewASR,
		utils.MetaACD: NewACD,
	}
	if _, has := metrics[metricID]; !has {
		return nil, fmt.Errorf("unsupported metric: %s", metricID)
	}
	return metrics[metricID]()
}

// StatsMetric is the interface which a metric should implement
type StatsMetric interface {
	GetStringValue(fmtOpts string) (val string)
	AddEvent(ev engine.StatsEvent) error
	RemEvent(ev engine.StatsEvent) error
	GetMarshaled(ms engine.Marshaler) (vals []byte, err error)
	SetFromMarshaled(vals []byte, ms engine.Marshaler) (err error) // mostly used to load from DB
}
