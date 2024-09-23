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

// do not modify this code because it's generated
package dispatchers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) AdminSv1ComputeFilterIndexIDs(ctx *context.Context, args *utils.ArgsComputeFilterIndexIDs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1ComputeFilterIndexIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1ComputeFilterIndexes(ctx *context.Context, args *utils.ArgsComputeFilterIndexes, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1ComputeFilterIndexes, args, reply)
}
func (dS *DispatcherService) AdminSv1FiltersMatch(ctx *context.Context, args *engine.ArgsFiltersMatch, reply *bool) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.CGREvent != nil && len(args.CGREvent.Tenant) != 0) {
		tnt = args.CGREvent.Tenant
	}
	ev := make(map[string]any)
	if args != nil && args.CGREvent != nil {
		ev = args.CGREvent.Event
	}
	opts := make(map[string]any)
	if args != nil && args.CGREvent != nil {
		opts = args.CGREvent.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1FiltersMatch, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAccount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.Account) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAccount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAccountIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAccountIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAccounts(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*utils.Account) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAccounts, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAccountsCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAccountsCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAccountsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAccountsIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetActionProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ActionProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetActionProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetActionProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetActionProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetActionProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.ActionProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetActionProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetActionProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetActionProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetActionsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetActionsIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.APIAttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAttributeProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAttributeProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAttributeProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAttributeProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.APIAttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAttributeProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAttributeProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAttributeProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetAttributesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetAttributesIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetCDRs(ctx *context.Context, args *utils.CDRFilters, reply *[]*utils.CDR) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetCDRs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetChargerProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetChargerProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetChargerProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetChargerProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.ChargerProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetChargerProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetChargerProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetChargerProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetChargersIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetChargersIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherHost, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherHostIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherHostIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherHosts(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.DispatcherHost) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherHosts, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherHostsCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherHostsCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.DispatcherProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatcherProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatcherProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetDispatchersIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetDispatchersIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Filter) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetFilter, args, reply)
}
func (dS *DispatcherService) AdminSv1GetFilterIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetFilterIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetFilterIndexes(ctx *context.Context, args *apis.AttrGetFilterIndexes, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetFilterIndexes, args, reply)
}
func (dS *DispatcherService) AdminSv1GetFilters(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.Filter) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetFilters, args, reply)
}
func (dS *DispatcherService) AdminSv1GetFiltersCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetFiltersCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRankingProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.RankingProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRankingProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRankingProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRankingProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRankingProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.RankingProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRankingProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRankingProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRankingProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *utils.RateProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfileRateIDs(ctx *context.Context, args *utils.ArgsSubItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfileRateIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfileRates(ctx *context.Context, args *utils.ArgsSubItemIDs, reply *[]*utils.Rate) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfileRates, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfileRatesCount(ctx *context.Context, args *utils.ArgsSubItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfileRatesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*utils.RateProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateProfilesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateProfilesIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRateRatesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRateRatesIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetResourceProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetResourceProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetResourceProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetResourceProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.ResourceProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetResourceProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetResourceProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetResourceProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetResourcesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetResourcesIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetReverseFilterHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *map[string]*engine.ReverseFilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetReverseFilterHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRouteProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRouteProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRouteProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRouteProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.RouteProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRouteProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRouteProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRouteProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetRoutesIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetRoutesIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetStatQueueProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetStatQueueProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetStatQueueProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetStatQueueProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.StatQueueProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetStatQueueProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetStatQueueProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetStatQueueProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetStatsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetStatsIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetThresholdProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetThresholdProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetThresholdProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetThresholdProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.ThresholdProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetThresholdProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetThresholdProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetThresholdProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1GetThresholdsIndexesHealth(ctx *context.Context, args *engine.IndexHealthArgs, reply *engine.FilterIHReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetThresholdsIndexesHealth, args, reply)
}
func (dS *DispatcherService) AdminSv1GetTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.TrendProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetTrendProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1GetTrendProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetTrendProfileIDs, args, reply)
}
func (dS *DispatcherService) AdminSv1GetTrendProfiles(ctx *context.Context, args *utils.ArgsItemIDs, reply *[]*engine.TrendProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetTrendProfiles, args, reply)
}
func (dS *DispatcherService) AdminSv1GetTrendProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1GetTrendProfilesCount, args, reply)
}
func (dS *DispatcherService) AdminSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1Ping, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveAccount(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveAccount, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveActionProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveActionProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveAttributeProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveCDRs(ctx *context.Context, args *utils.CDRFilters, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveCDRs, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveChargerProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveDispatcherHost, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveDispatcherProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveFilter, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveFilterIndexes(ctx *context.Context, args *apis.AttrRemFilterIndexes, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveFilterIndexes, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveRankingProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveRankingProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveRateProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveRateProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveRateProfileRates(ctx *context.Context, args *utils.RemoveRPrfRates, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveRateProfileRates, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveResourceProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveRouteProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveStatQueueProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveThresholdProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1RemoveTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TenantID != nil && len(args.TenantID.Tenant) != 0) {
		tnt = args.TenantID.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1RemoveTrendProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetAccount(ctx *context.Context, args *utils.AccountWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.Account != nil && len(args.Account.Tenant) != 0) {
		tnt = args.Account.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetAccount, args, reply)
}
func (dS *DispatcherService) AdminSv1SetActionProfile(ctx *context.Context, args *engine.ActionProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ActionProfile != nil && len(args.ActionProfile.Tenant) != 0) {
		tnt = args.ActionProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetActionProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetAttributeProfile(ctx *context.Context, args *engine.APIAttributeProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.APIAttributeProfile != nil && len(args.APIAttributeProfile.Tenant) != 0) {
		tnt = args.APIAttributeProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetAttributeProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetChargerProfile(ctx *context.Context, args *apis.ChargerWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ChargerProfile != nil && len(args.ChargerProfile.Tenant) != 0) {
		tnt = args.ChargerProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetChargerProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetDispatcherHost(ctx *context.Context, args *engine.DispatcherHostWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.DispatcherHost != nil && len(args.DispatcherHost.Tenant) != 0) {
		tnt = args.DispatcherHost.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetDispatcherHost, args, reply)
}
func (dS *DispatcherService) AdminSv1SetDispatcherProfile(ctx *context.Context, args *apis.DispatcherWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.DispatcherProfile != nil && len(args.DispatcherProfile.Tenant) != 0) {
		tnt = args.DispatcherProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetDispatcherProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetFilter(ctx *context.Context, args *engine.FilterWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.Filter != nil && len(args.Filter.Tenant) != 0) {
		tnt = args.Filter.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetFilter, args, reply)
}
func (dS *DispatcherService) AdminSv1SetRankingProfile(ctx *context.Context, args *engine.RankingProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.RankingProfile != nil && len(args.RankingProfile.Tenant) != 0) {
		tnt = args.RankingProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetRankingProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetRateProfile(ctx *context.Context, args *utils.APIRateProfile, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.RateProfile != nil && len(args.RateProfile.Tenant) != 0) {
		tnt = args.RateProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetRateProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetResourceProfile(ctx *context.Context, args *engine.ResourceProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ResourceProfile != nil && len(args.ResourceProfile.Tenant) != 0) {
		tnt = args.ResourceProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetResourceProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetRouteProfile(ctx *context.Context, args *engine.RouteProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.RouteProfile != nil && len(args.RouteProfile.Tenant) != 0) {
		tnt = args.RouteProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetRouteProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetStatQueueProfile(ctx *context.Context, args *engine.StatQueueProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.StatQueueProfile != nil && len(args.StatQueueProfile.Tenant) != 0) {
		tnt = args.StatQueueProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetStatQueueProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.ThresholdProfile != nil && len(args.ThresholdProfile.Tenant) != 0) {
		tnt = args.ThresholdProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetThresholdProfile, args, reply)
}
func (dS *DispatcherService) AdminSv1SetTrendProfile(ctx *context.Context, args *engine.TrendProfileWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && (args.TrendProfile != nil && len(args.TrendProfile.Tenant) != 0) {
		tnt = args.TrendProfile.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAdminS, utils.AdminSv1SetTrendProfile, args, reply)
}
