// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package rankings

import (
	"slices"
	"strings"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// V1ScheduleQueries manually schedules or reschedules ranking queries.
func (rkS *RankingS) V1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleRankingQueries, scheduled *int) (err error) {
	if sched, errSched := rkS.scheduleRankingQueries(ctx, args.Tenant, args.RankingIDs); errSched != nil {
		return errSched
	} else {
		*scheduled = sched
	}
	return
}

// V1GetRanking retrieves ranking metrics with optional filtering.
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

// V1GetSchedule retrieves information about currently scheduled rankings.
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

// V1GetRankingSummary retrieves the most recent ranking summary.
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
