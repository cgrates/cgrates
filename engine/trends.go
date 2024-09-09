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
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// NewTrendS is the constructor for TrendS
func NewTrendS(dm *DataManager,
	connMgr *ConnManager,
	filterS *FilterS,
	cgrcfg *config.CGRConfig) *TrendS {
	return &TrendS{
		dm:          dm,
		connMgr:     connMgr,
		filterS:     filterS,
		cgrcfg:      cgrcfg,
		loopStopped: make(chan struct{}),
		crnTQsMux:   new(sync.RWMutex),
		crnTQs:      make(map[string]map[string]cron.EntryID),
	}
}

// TrendS is responsible of implementing the logic of TrendService
type TrendS struct {
	dm      *DataManager
	connMgr *ConnManager
	filterS *FilterS
	cgrcfg  *config.CGRConfig

	crn *cron.Cron // cron reference

	crnTQsMux *sync.RWMutex                      // protects the crnTQs
	crnTQs    map[string]map[string]cron.EntryID // save the EntryIDs for TrendQueries so we can reschedule them when needed

	loopStopped chan struct{}
}

// computeTrend will query a stat and build the Trend for it
//
//	it is be called by Cron service
func (tS *TrendS) computeTrend(tP *TrendProfile) (err error) {
	return
}

// scheduleTrendQueries will schedule/re-schedule specific trend queries
func (tS *TrendS) scheduleTrendQueries(ctx *context.Context, tnt string, tIDs []string) (complete bool) {
	complete = true
	for _, tID := range tIDs {
		tS.crnTQsMux.RLock()
		if entryID, has := tS.crnTQs[tnt][tID]; has {
			tS.crn.Remove(entryID) // deschedule the query
		}
		tS.crnTQsMux.RUnlock()
		if tP, err := tS.dm.GetTrendProfile(tnt, tID, true, true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed retrieving TrendProfile with id: <%s:%s> for scheduling, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			complete = false
		} else if entryID, err := tS.crn.AddFunc(tP.Schedule,
			func() { tS.computeTrend(tP) }); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduling TrendProfile <%s:%s>, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))

		} else {
			tS.crnTQsMux.Lock()
			tS.crnTQs[tP.Tenant][tP.ID] = entryID
		}

	}
	return
}
