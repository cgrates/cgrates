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

func NewTrendService(dm *DataManager, cgrcfg *config.CGRConfig, filterS *FilterS, connMgr *ConnManager) (tS *TrendS) {
	return &TrendS{
		dm:      dm,
		cfg:     cgrcfg,
		fltrS:   filterS,
		connMgr: connMgr,
	}
}

// TrendS manages Trend execution
type TrendS struct {
	dm      *DataManager
	cfg     *config.CGRConfig
	fltrS   *FilterS
	connMgr *ConnManager
}
