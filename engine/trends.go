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

type TrendProfile struct {
	Tenant         string
	ID             string
	QueryInterval  time.Duration
	StatID         string
	QueueLength    int
	TTL            time.Duration
	PurgeFilterIDs []string
	Trend          string
	ThresholdIDs   []string
}

type TrendProfileWithAPIOpts struct {
	*TrendProfile
	APIOpts map[string]any
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
	Tenant      string
	ID          string
	Trend       string
	QueueLength int
	trPrfl      *TrendProfile
}

func (tr *Trend) TenantID() string {
	return utils.ConcatenatedKey(tr.Tenant, tr.ID)
}
