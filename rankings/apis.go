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

package rankings

import (
	"slices"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// V1ScheduleQueries is the query for manually re-/scheduling Ranking Queries
func (rkS *RankingS) V1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleRankingQueries, scheduled *int) (err error) {
	if sched, errSched := rkS.scheduleRankingQueries(ctx, args.Tenant, args.RankingIDs); errSched != nil {
		return errSched
	} else {
		*scheduled = sched
	}
	return
}

// V1GetRanking is the API to return the Ranking instance
func (rkS *RankingS) V1GetRanking(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, retRanking *utils.Ranking) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var rk *utils.Ranking
	if rk, err = rkS.dm.GetRanking(ctx, arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	rk.RLock()
	defer rk.RUnlock()
	retRanking.Tenant = rk.Tenant // avoid vet complaining for mutex copying
	retRanking.ID = rk.ID
	retRanking.Metrics = make(map[string]map[string]float64)
	for statID, metrics := range rk.Metrics {
		retRanking.Metrics[statID] = make(map[string]float64)
		for metricID, val := range metrics {
			retRanking.Metrics[statID][metricID] = val
		}
	}
	retRanking.LastUpdate = rk.LastUpdate
	retRanking.Sorting = rk.Sorting

	retRanking.SortingParameters = make([]string, len(rk.SortingParameters))
	copy(retRanking.SortingParameters, rk.SortingParameters)

	retRanking.SortedStatIDs = make([]string, len(rk.SortedStatIDs))
	copy(retRanking.SortedStatIDs, rk.SortedStatIDs)
	return
}

// V1GetSchedule returns the active schedule for Ranking queries
func (rkS *RankingS) V1GetSchedule(ctx *context.Context, args *utils.ArgScheduledRankings, schedRankings *[]utils.ScheduledRanking) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rkS.cgrcfg.GeneralCfg().DefaultTenant
	}
	rkS.crnRQsMux.RLock()
	defer rkS.crnRQsMux.RUnlock()
	rankingIDsMp, has := rkS.crnRQs[tnt]
	if !has {
		return utils.ErrNotFound
	}
	var scheduledRankings []utils.ScheduledRanking
	var entryIds map[string]cron.EntryID
	if len(args.RankingIDPrefixes) == 0 {
		entryIds = rankingIDsMp
	} else {
		entryIds = make(map[string]cron.EntryID)
		for _, rkID := range args.RankingIDPrefixes {
			for key, entryID := range rankingIDsMp {
				if strings.HasPrefix(key, rkID) {
					entryIds[key] = entryID
				}
			}
		}
	}
	if len(entryIds) == 0 {
		return utils.ErrNotFound
	}
	var entry cron.Entry
	for id, entryID := range entryIds {
		entry = rkS.crn.Entry(entryID)
		if entry.ID == 0 {
			continue
		}
		scheduledRankings = append(scheduledRankings,
			utils.ScheduledRanking{
				RankingID: id,
				Next:      entry.Next,
				Previous:  entry.Prev,
			})
	}
	slices.SortFunc(scheduledRankings, func(a, b utils.ScheduledRanking) int {
		return a.Next.Compare(b.Next)
	})
	*schedRankings = scheduledRankings
	return nil
}

// V1GetRankingSummary returns a summary of ascending/descending stat of the last updated ranking
func (rS *RankingS) V1GetRankingSummary(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.RankingSummary) (err error) {
	var rnk *utils.Ranking
	if rnk, err = rS.dm.GetRanking(ctx, arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	rnk.RLock()
	rnkS := rnk.AsRankingSummary()
	rnk.RUnlock()
	*reply = *rnkS
	return
}
