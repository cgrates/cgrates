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

package trends

import (
	"slices"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// V1ScheduleQueries manually schedules or reschedules trend queries.
func (tS *TrendS) V1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleTrendQueries, scheduled *int) (err error) {
	if sched, errSched := tS.scheduleTrendQueries(ctx, args.Tenant, args.TrendIDs); errSched != nil {
		return errSched
	} else {
		*scheduled = sched
	}
	return
}

// V1GetTrend retrieves trend metrics with optional time and index filtering.
func (tS *TrendS) V1GetTrend(ctx *context.Context, arg *utils.ArgGetTrend, retTrend *utils.Trend) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var trnd *utils.Trend
	if trnd, err = tS.dm.GetTrend(ctx, arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	trnd.RLock()
	defer trnd.RUnlock()
	retTrend.Tenant = trnd.Tenant // avoid vet complaining for mutex copying
	retTrend.ID = trnd.ID
	startIdx := arg.RunIndexStart
	if startIdx > len(trnd.RunTimes) {
		startIdx = len(trnd.RunTimes)
	}
	endIdx := arg.RunIndexEnd
	if endIdx > len(trnd.RunTimes) ||
		endIdx < startIdx ||
		endIdx == 0 {
		endIdx = len(trnd.RunTimes)
	}
	runTimes := trnd.RunTimes[startIdx:endIdx]
	if len(runTimes) == 0 {
		return utils.ErrNotFound
	}
	var tStart, tEnd time.Time
	if arg.RunTimeStart == utils.EmptyString {
		tStart = runTimes[0]
	} else if tStart, err = utils.ParseTimeDetectLayout(arg.RunTimeStart, tS.cfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	if arg.RunTimeEnd == utils.EmptyString {
		tEnd = runTimes[len(runTimes)-1].Add(time.Duration(1))
	} else if tEnd, err = utils.ParseTimeDetectLayout(arg.RunTimeEnd, tS.cfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	retTrend.RunTimes = make([]time.Time, 0, len(runTimes))
	for _, runTime := range runTimes {
		if !runTime.Before(tStart) && runTime.Before(tEnd) {
			retTrend.RunTimes = append(retTrend.RunTimes, runTime)
		}
	}
	if len(retTrend.RunTimes) == 0 { // filtered out all
		return utils.ErrNotFound
	}
	retTrend.Metrics = make(map[time.Time]map[string]*utils.MetricWithTrend)
	for _, runTime := range retTrend.RunTimes {
		retTrend.Metrics[runTime] = trnd.Metrics[runTime]
	}
	return
}

// V1GetScheduledTrends retrieves information about currently scheduled trends.
func (tS *TrendS) V1GetScheduledTrends(ctx *context.Context, args *utils.ArgScheduledTrends, schedTrends *[]utils.ScheduledTrend) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cfg.GeneralCfg().DefaultTenant
	}
	tS.crnTQsMux.RLock()
	defer tS.crnTQsMux.RUnlock()
	trendIDsMp, has := tS.crnTQs[tnt]
	if !has {
		return utils.ErrNotFound
	}
	var scheduledTrends []utils.ScheduledTrend
	var entryIds map[string]cron.EntryID
	if len(args.TrendIDPrefixes) == 0 {
		entryIds = trendIDsMp
	} else {
		entryIds = make(map[string]cron.EntryID)
		for _, tID := range args.TrendIDPrefixes {
			for key, entryID := range trendIDsMp {
				if strings.HasPrefix(key, tID) {
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
		entry = tS.crn.Entry(entryID)
		if entry.ID == 0 {
			continue
		}
		scheduledTrends = append(scheduledTrends, utils.ScheduledTrend{
			TrendID:  id,
			Next:     entry.Next,
			Previous: entry.Prev,
		})
	}
	slices.SortFunc(scheduledTrends, func(a, b utils.ScheduledTrend) int {
		return a.Next.Compare(b.Next)
	})
	*schedTrends = scheduledTrends
	return nil
}

// V1GetTrendSummary retrieves the most recent trend summary.
func (tS *TrendS) V1GetTrendSummary(ctx *context.Context, arg utils.TenantIDWithAPIOpts, reply *utils.TrendSummary) (err error) {
	var trnd *utils.Trend
	if trnd, err = tS.dm.GetTrend(ctx, arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	trnd.RLock()
	trndS := trnd.AsTrendSummary()
	trnd.RUnlock()
	*reply = *trndS
	return
}
