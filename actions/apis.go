/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package actions

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// V1ScheduleActions will be called to schedule actions matching the arguments
func (aS *ActionS) V1ScheduleActions(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
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
	if err = aS.scheduleActions(ctx, []*utils.CGREvent{args},
		actPrfIDs, ignFilters, false); err != nil {
		return
	}
	*rpl = utils.OK
	return
}

// V1ExecuteActions will be called to execute ASAP action profiles, ignoring their Schedule field
func (aS *ActionS) V1ExecuteActions(ctx *context.Context, args *utils.CGREvent, rpl *string) (err error) {
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
		args, actPrfIDs, ignFilters, true); err != nil {
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
