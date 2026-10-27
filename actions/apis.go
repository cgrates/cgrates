// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package actions

import (
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// V1ScheduleActions will be called to schedule actions matching the arguments
func (aS *ActionS) V1ScheduleActions(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
	if args.Tenant == utils.EmptyString {
		args.Tenant = aS.cfg.GeneralCfg().DefaultTenant
	}
	var actPrfIDs []string
	if actPrfIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.ActionSCfg().Opts.ProfileIDs,
		config.ActionsProfileIDsDftOpt, utils.OptsActionsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.ActionSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var overwrite bool
	if value, has := args.APIOpts[utils.MetaOverwrite]; has {
		if overwrite, err = utils.IfaceAsBool(value); err != nil {
			return err
		}
	}
	if err = aS.scheduleActions(ctx, []*utils.CGREvent{args},
		actPrfIDs, ignFilters, overwrite, false); err != nil {
		return
	}
	*rpl = utils.OK
	return
}

// V1GetScheduledActions returns the registered cron batches for a tenant.
func (aS *ActionS) V1GetScheduledActions(_ *context.Context, args *utils.GetScheduledActionsArgs, reply *[]utils.ScheduledActionBatch) error {
	tenant := args.Tenant
	if tenant == "" {
		tenant = aS.cfg.GeneralCfg().DefaultTenant
	}
	aS.crnLk.RLock()
	entries := aS.crn.Entries()
	aS.crnLk.RUnlock()

	var batches []utils.ScheduledActionBatch
	for _, entry := range entries {
		job := entry.Job.(*actionJob)
		first := job.batch[0]
		if first.tenant != tenant {
			continue
		}
		if len(args.ActionProfileIDs) != 0 && !slices.Contains(args.ActionProfileIDs, first.apID) {
			continue
		}
		targets := make(map[string][]string)
		for _, target := range job.batch {
			targets[target.trgTyp] = append(targets[target.trgTyp], target.trgID)
		}
		for _, ids := range targets {
			slices.Sort(ids)
		}
		batches = append(batches, utils.ScheduledActionBatch{
			Tenant:          first.tenant,
			ActionProfileID: first.apID,
			Schedule:        job.schedule,
			Targets:         targets,
			NextRun:         entry.Next,
		})
	}
	if len(batches) == 0 {
		return utils.ErrNotFound
	}
	slices.SortFunc(batches, func(a, b utils.ScheduledActionBatch) int {
		return a.NextRun.Compare(b.NextRun)
	})
	*reply = batches
	return nil
}

// V1ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aS *ActionS) V1ExecuteActions(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = aS.cfg.GeneralCfg().DefaultTenant
	}
	var actPrfIDs []string
	if actPrfIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.ActionSCfg().Opts.ProfileIDs,
		config.ActionsProfileIDsDftOpt, utils.OptsActionsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, aS.fltrS, aS.cfg.ActionSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var schedActSet []*scheduledActs
	if schedActSet, err = aS.scheduledActions(ctx, args.Tenant,
		args, actPrfIDs, ignFilters); err != nil {
		return
	}
	var partExec bool
	// execute the actions
	for _, sActs := range schedActSet {
		if err = aS.asapExecuteActions(ctx, sActs); err != nil {
			partExec = true
		}
	}
	if partExec {
		err = utils.ErrPartiallyExecuted
		return
	}
	*rpl = utils.OK
	return
}
