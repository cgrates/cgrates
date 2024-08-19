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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

type TrendProfile struct {
	Tenant       string
	ID           string
	Schedule     string // Cron expression scheduling gathering of the metrics
	StatID       string
	Metrics      []MetricWithSettings
	QueueLength  int
	TTL          time.Duration
	TrendType    string // *last, *average
	ThresholdIDs []string
}

// MetricWithSettings adds specific settings to the Metric
type MetricWithSettings struct {
	MetricID         string
	TrendSwingMargin float64 // allow this margin for *neutral trend
}

type TrendProfileWithAPIOpts struct {
	*TrendProfile
	APIOpts map[string]any
}

type TrendProfilesAPI struct {
	Tenant string
	TpIDs  []string
}

func (srp *TrendProfile) TenantID() string {
	return utils.ConcatenatedKey(srp.Tenant, srp.ID)
}

type TrendWithAPIOpts struct {
	*Trend
	APIOpts map[string]any
}

// Trend is the unit matched by filters
type Trend struct {
	Tenant   string
	ID       string
	RunTimes []time.Time
	Metrics  map[time.Time]map[string]MetricWithTrend
	totals   map[string]float64 // cached sum, used for average calculations
}

// MetricWithTrend represents one read from StatS
type MetricWithTrend struct {
	ID    string  // Metric ID
	Value float64 // Metric Value
	Trend string  // *positive, *negative, *neutral
}

func (tr *Trend) TenantID() string {
	return utils.ConcatenatedKey(tr.Tenant, tr.ID)
}

// NewTrendService the constructor for TrendS service
func NewTrendService(dm *DataManager, cgrcfg *config.CGRConfig, filterS *FilterS) *TrendService {
	return &TrendService{
		dm:      dm,
		cgrcfg:  cgrcfg,
		filterS: filterS,
	}
}

type TrendService struct {
	dm      *DataManager
	cgrcfg  *config.CGRConfig
	filterS *FilterS
}
