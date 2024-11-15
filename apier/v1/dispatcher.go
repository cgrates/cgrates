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

package v1

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/ers"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// GetDispatcherProfile returns a Dispatcher Profile
func (apierSv1 *APIerSv1) GetDispatcherProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.DispatcherProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	dpp, err := apierSv1.DataManager.GetDispatcherProfile(tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *dpp
	return nil
}

// GetDispatcherProfileIDs returns list of dispatcherProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetDispatcherProfileIDs(ctx *context.Context, tenantArg *utils.PaginatorWithTenant, dPrfIDs *[]string) error {
	tenant := tenantArg.Tenant
	if tenant == utils.EmptyString {
		tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.DispatcherProfilePrefix + tenant + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*dPrfIDs = tenantArg.PaginateStringSlice(retIDs)
	return nil
}

type DispatcherWithAPIOpts struct {
	*engine.DispatcherProfile
	APIOpts map[string]any
}

// SetDispatcherProfile add/update a new Dispatcher Profile
func (apierSv1 *APIerSv1) SetDispatcherProfile(ctx *context.Context, args *DispatcherWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.DispatcherProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.SetDispatcherProfile(args.DispatcherProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheDispatcherProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<SetDispatcherProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatcherProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), args.Tenant, utils.CacheDispatcherProfiles,
		args.TenantID(), utils.EmptyString, &args.FilterIDs, args.Subsystems, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	cacheAct := utils.MetaRemove
	if err := apierSv1.CallCache(utils.FirstNonEmpty(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), cacheAct),
		args.Tenant, utils.CacheDispatchers, args.TenantID(), utils.EmptyString, &args.FilterIDs, args.Subsystems, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherRoutes
	if err := apierSv1.CallCache(utils.FirstNonEmpty(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), cacheAct),
		args.Tenant, utils.CacheDispatcherRoutes, args.TenantID(),
		utils.ConcatenatedKey(utils.CacheDispatcherProfiles, args.Tenant, args.ID),
		&args.FilterIDs, args.Subsystems, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveDispatcherProfile remove a specific Dispatcher Profile
func (apierSv1 *APIerSv1) RemoveDispatcherProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveDispatcherProfile(tnt,
		arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheDispatcherProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if apierSv1.Config.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<RemoveDispatcherProfile> Delaying cache call for %v", apierSv1.Config.GeneralCfg().CachingDelay))
		time.Sleep(apierSv1.Config.GeneralCfg().CachingDelay)
	}
	//handle caching for DispatcherProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheDispatcherProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetDispatcherHost returns a Dispatcher Host
func (apierSv1 *APIerSv1) GetDispatcherHost(ctx *context.Context, arg *utils.TenantID, reply *engine.DispatcherHost) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	dpp, err := apierSv1.DataManager.GetDispatcherHost(tnt, arg.ID, true, false, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *dpp
	return nil
}

// GetDispatcherHostIDs returns list of dispatcherHost IDs registered for a tenant
func (apierSv1 *APIerSv1) GetDispatcherHostIDs(ctx *context.Context, tenantArg *utils.PaginatorWithTenant, dPrfIDs *[]string) error {
	tenant := tenantArg.Tenant
	if tenant == utils.EmptyString {
		tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.DispatcherHostPrefix + tenant + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*dPrfIDs = tenantArg.PaginateStringSlice(retIDs)
	return nil
}

// SetDispatcherHost add/update a new Dispatcher Host
func (apierSv1 *APIerSv1) SetDispatcherHost(ctx *context.Context, args *engine.DispatcherHostWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.DispatcherHost, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.SetDispatcherHost(args.DispatcherHost); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherHosts and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheDispatcherHosts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), args.Tenant, utils.CacheDispatcherHosts,
		args.TenantID(), utils.EmptyString, nil, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveDispatcherHost remove a specific Dispatcher Host
func (apierSv1 *APIerSv1) RemoveDispatcherHost(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveDispatcherHost(tnt,
		arg.ID); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheDispatcherHosts and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheDispatcherHosts: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for DispatcherProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheDispatcherHosts,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewDispatcherThresholdSv1(dps *dispatchers.DispatcherService) *DispatcherThresholdSv1 {
	return &DispatcherThresholdSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherThresholdSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements ThresholdSv1Ping
func (dT *DispatcherThresholdSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dT.dS.ThresholdSv1Ping(ctx, args, reply)
}

// GetThresholdsForEvent implements ThresholdSv1GetThresholdsForEvent
func (dT *DispatcherThresholdSv1) GetThresholdsForEvent(ctx *context.Context, tntID *utils.CGREvent,
	t *engine.Thresholds) error {
	return dT.dS.ThresholdSv1GetThresholdsForEvent(ctx, tntID, t)
}

// ProcessEvent implements ThresholdSv1ProcessEvent
func (dT *DispatcherThresholdSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1ProcessEvent(ctx, args, tIDs)
}

func (dT *DispatcherThresholdSv1) GetThresholdIDs(ctx *context.Context, args *utils.TenantWithAPIOpts,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1GetThresholdIDs(ctx, args, tIDs)
}

func (dT *DispatcherThresholdSv1) GetThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts,
	th *engine.Threshold) error {
	return dT.dS.ThresholdSv1GetThreshold(ctx, args, th)
}

func NewDispatcherTrendSv1(dps *dispatchers.DispatcherService) *DispatcherTrendSv1 {
	return &DispatcherTrendSv1{dS: dps}
}

type DispatcherTrendSv1 struct {
	dS *dispatchers.DispatcherService
}

func (dT *DispatcherTrendSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	return dT.dS.TrendSv1Ping(ctx, args, reply)
}

func (dT *DispatcherTrendSv1) ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleTrendQueries,
	reply *int) (err error) {
	return dT.dS.TrendSv1ScheduleQueries(ctx, args, reply)
}

func (dT *DispatcherTrendSv1) GetTrend(ctx *context.Context, args *utils.ArgGetTrend, reply *engine.Trend) (err error) {
	return dT.dS.TrendSv1GetTrend(ctx, args, reply)
}

func (dT *DispatcherTrendSv1) GetScheduledTrends(ctx *context.Context, args *utils.ArgScheduledTrends, reply *[]utils.ScheduledTrend) (err error) {
	return dT.dS.TrendSv1GetScheduledTrends(ctx, args, reply)
}

func (dT *DispatcherTrendSv1) GetTrendSummary(ctx *context.Context, args utils.TenantIDWithAPIOpts, reply *engine.TrendSummary) error {

	return dT.dS.TrendSv1GetTrendSummary(ctx, args, reply)
}

func NewDispatcherStatSv1(dps *dispatchers.DispatcherService) *DispatcherStatSv1 {
	return &DispatcherStatSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherStatSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements StatSv1Ping
func (dSts *DispatcherStatSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dSts.dS.StatSv1Ping(ctx, args, reply)
}

// GetStatQueuesForEvent implements StatSv1GetStatQueuesForEvent
func (dSts *DispatcherStatSv1) GetStatQueuesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) error {
	return dSts.dS.StatSv1GetStatQueuesForEvent(ctx, args, reply)
}

// GetQueueStringMetrics implements StatSv1GetQueueStringMetrics
func (dSts *DispatcherStatSv1) GetQueueStringMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts,
	reply *map[string]string) error {
	return dSts.dS.StatSv1GetQueueStringMetrics(ctx, args, reply)
}

func (dSts *DispatcherStatSv1) GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts,
	reply *map[string]float64) error {
	return dSts.dS.StatSv1GetQueueFloatMetrics(ctx, args, reply)
}

func (dSts *DispatcherStatSv1) GetQueueIDs(ctx *context.Context, args *utils.TenantWithAPIOpts,
	reply *[]string) error {
	return dSts.dS.StatSv1GetQueueIDs(ctx, args, reply)
}

// GetQueueStringMetrics implements StatSv1ProcessEvent
func (dSts *DispatcherStatSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) error {
	return dSts.dS.StatSv1ProcessEvent(ctx, args, reply)
}

func NewDispatcherRankingSv1(dps *dispatchers.DispatcherService) *DispatcherRankingSv1 {
	return &DispatcherRankingSv1{ds: dps}
}

type DispatcherRankingSv1 struct {
	ds *dispatchers.DispatcherService
}

func (dRn *DispatcherRankingSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dRn.ds.RankingSv1Ping(ctx, args, reply)
}

func (dRn *DispatcherRankingSv1) GetRankingSummary(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.RankingSummary) error {
	return dRn.ds.RankingSv1GetRankingSummary(ctx, args, reply)
}

func (dRn *DispatcherRankingSv1) GetSchedule(ctx *context.Context, args *utils.ArgScheduledRankings, reply *[]utils.ScheduledRanking) error {
	return dRn.ds.RankingSv1GetSchedule(ctx, args, reply)
}

func (dRn *DispatcherRankingSv1) ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleRankingQueries, reply *int) (err error) {
	return dRn.ds.RankingSv1ScheduleQueries(ctx, args, reply)
}

func (dRn *DispatcherRankingSv1) GetRanking(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Ranking) (err error) {
	return dRn.ds.RankingSv1GetRanking(ctx, args, reply)
}

func NewDispatcherResourceSv1(dps *dispatchers.DispatcherService) *DispatcherResourceSv1 {
	return &DispatcherResourceSv1{dRs: dps}
}

// Exports RPC from RLs
type DispatcherResourceSv1 struct {
	dRs *dispatchers.DispatcherService
}

// Ping implements ResourceSv1Ping
func (dRs *DispatcherResourceSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dRs.dRs.ResourceSv1Ping(ctx, args, reply)
}

// GetResourcesForEvent implements ResourceSv1GetResourcesForEvent
func (dRs *DispatcherResourceSv1) GetResourcesForEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.Resources) error {
	return dRs.dRs.ResourceSv1GetResourcesForEvent(ctx, args, reply)
}

func (dRs *DispatcherResourceSv1) GetResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	return dRs.dRs.ResourceSv1GetResource(ctx, args, reply)
}

func (dRs *DispatcherResourceSv1) GetResourceWithConfig(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ResourceWithConfig) error {
	return dRs.dRs.ResourceSv1GetResourceWithConfig(ctx, args, reply)
}

func (dRs *DispatcherResourceSv1) AuthorizeResources(ctx *context.Context, args *utils.CGREvent,
	reply *string) error {
	return dRs.dRs.ResourceSv1AuthorizeResources(ctx, args, reply)
}

func (dRs *DispatcherResourceSv1) AllocateResources(ctx *context.Context, args *utils.CGREvent,
	reply *string) error {
	return dRs.dRs.ResourceSv1AllocateResources(ctx, args, reply)
}

func (dRs *DispatcherResourceSv1) ReleaseResources(ctx *context.Context, args *utils.CGREvent,
	reply *string) error {
	return dRs.dRs.ResourceSv1ReleaseResources(ctx, args, reply)
}

func NewDispatcherRouteSv1(dps *dispatchers.DispatcherService) *DispatcherRouteSv1 {
	return &DispatcherRouteSv1{dRoute: dps}
}

// Exports RPC from RouteS
type DispatcherRouteSv1 struct {
	dRoute *dispatchers.DispatcherService
}

// Ping implements RouteSv1Ping
func (dRoute *DispatcherRouteSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dRoute.dRoute.RouteSv1Ping(ctx, args, reply)
}

// GetRoutes implements RouteSv1GetRoutes
func (dRoute *DispatcherRouteSv1) GetRoutes(ctx *context.Context, args *utils.CGREvent, reply *engine.SortedRoutesList) error {
	return dRoute.dRoute.RouteSv1GetRoutes(ctx, args, reply)
}

// GetRouteProfilesForEvent returns a list of route profiles that match for Event
func (dRoute *DispatcherRouteSv1) GetRouteProfilesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*engine.RouteProfile) error {
	return dRoute.dRoute.RouteSv1GetRouteProfilesForEvent(ctx, args, reply)
}

// GetRoutesList returns sorted list of routes for Event as a string slice
func (dRoute *DispatcherRouteSv1) GetRoutesList(ctx *context.Context, args *utils.CGREvent, reply *[]string) error {
	return dRoute.dRoute.RouteSv1GetRoutesList(ctx, args, reply)
}

func NewDispatcherAttributeSv1(dps *dispatchers.DispatcherService) *DispatcherAttributeSv1 {
	return &DispatcherAttributeSv1{dA: dps}
}

// Exports RPC from RLs
type DispatcherAttributeSv1 struct {
	dA *dispatchers.DispatcherService
}

// Ping implements AttributeSv1Ping
func (dA *DispatcherAttributeSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dA.dA.AttributeSv1Ping(ctx, args, reply)
}

// GetAttributeForEvent implements AttributeSv1GetAttributeForEvent
func (dA *DispatcherAttributeSv1) GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AttributeProfile) error {
	return dA.dA.AttributeSv1GetAttributeForEvent(ctx, args, reply)
}

// ProcessEvent implements AttributeSv1ProcessEvent
func (dA *DispatcherAttributeSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) error {
	return dA.dA.AttributeSv1ProcessEvent(ctx, args, reply)
}

func NewDispatcherChargerSv1(dps *dispatchers.DispatcherService) *DispatcherChargerSv1 {
	return &DispatcherChargerSv1{dC: dps}
}

// Exports RPC from RLs
type DispatcherChargerSv1 struct {
	dC *dispatchers.DispatcherService
}

// Ping implements ChargerSv1Ping
func (dC *DispatcherChargerSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dC.dC.ChargerSv1Ping(ctx, args, reply)
}

// GetChargersForEvent implements ChargerSv1GetChargersForEvent
func (dC *DispatcherChargerSv1) GetChargersForEvent(ctx *context.Context, args *utils.CGREvent,
	reply *engine.ChargerProfiles) (err error) {
	return dC.dC.ChargerSv1GetChargersForEvent(ctx, args, reply)
}

// ProcessEvent implements ChargerSv1ProcessEvent
func (dC *DispatcherChargerSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *[]*engine.ChrgSProcessEventReply) (err error) {
	return dC.dC.ChargerSv1ProcessEvent(ctx, args, reply)
}

func NewDispatcherSessionSv1(dps *dispatchers.DispatcherService) *DispatcherSessionSv1 {
	return &DispatcherSessionSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherSessionSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements SessionSv1Ping
func (dS *DispatcherSessionSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.SessionSv1Ping(ctx, args, reply)
}

// AuthorizeEventWithDigest implements SessionSv1AuthorizeEventWithDigest
func (dS *DispatcherSessionSv1) AuthorizeEventWithDigest(ctx *context.Context, args *sessions.V1AuthorizeArgs,
	reply *sessions.V1AuthorizeReplyWithDigest) error {
	return dS.dS.SessionSv1AuthorizeEventWithDigest(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) AuthorizeEvent(ctx *context.Context, args *sessions.V1AuthorizeArgs,
	reply *sessions.V1AuthorizeReply) error {
	return dS.dS.SessionSv1AuthorizeEvent(ctx, args, reply)
}

// InitiateSessionWithDigest implements SessionSv1InitiateSessionWithDigest
func (dS *DispatcherSessionSv1) InitiateSessionWithDigest(ctx *context.Context, args *sessions.V1InitSessionArgs,
	reply *sessions.V1InitReplyWithDigest) (err error) {
	return dS.dS.SessionSv1InitiateSessionWithDigest(ctx, args, reply)
}

// InitiateSessionWithDigest implements SessionSv1InitiateSessionWithDigest
func (dS *DispatcherSessionSv1) InitiateSession(ctx *context.Context, args *sessions.V1InitSessionArgs,
	reply *sessions.V1InitSessionReply) (err error) {
	return dS.dS.SessionSv1InitiateSession(ctx, args, reply)
}

// ProcessCDR implements SessionSv1ProcessCDR
func (dS *DispatcherSessionSv1) ProcessCDR(ctx *context.Context, args *utils.CGREvent,
	reply *string) (err error) {
	return dS.dS.SessionSv1ProcessCDR(ctx, args, reply)
}

// ProcessMessage implements SessionSv1ProcessMessage
func (dS *DispatcherSessionSv1) ProcessMessage(ctx *context.Context, args *sessions.V1ProcessMessageArgs,
	reply *sessions.V1ProcessMessageReply) (err error) {
	return dS.dS.SessionSv1ProcessMessage(ctx, args, reply)
}

// ProcessMessage implements SessionSv1ProcessMessage
func (dS *DispatcherSessionSv1) ProcessEvent(ctx *context.Context, args *sessions.V1ProcessEventArgs,
	reply *sessions.V1ProcessEventReply) (err error) {
	return dS.dS.SessionSv1ProcessEvent(ctx, args, reply)
}

// GetCost implements SessionSv1GetCost
func (dS *DispatcherSessionSv1) GetCost(ctx *context.Context, args *sessions.V1ProcessEventArgs,
	reply *sessions.V1GetCostReply) (err error) {
	return dS.dS.SessionSv1GetCost(ctx, args, reply)
}

// TerminateSession implements SessionSv1TerminateSession
func (dS *DispatcherSessionSv1) TerminateSession(ctx *context.Context, args *sessions.V1TerminateSessionArgs,
	reply *string) (err error) {
	return dS.dS.SessionSv1TerminateSession(ctx, args, reply)
}

// UpdateSession implements SessionSv1UpdateSession
func (dS *DispatcherSessionSv1) UpdateSession(ctx *context.Context, args *sessions.V1UpdateSessionArgs,
	reply *sessions.V1UpdateSessionReply) (err error) {
	return dS.dS.SessionSv1UpdateSession(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) GetActiveSessions(ctx *context.Context, args *utils.SessionFilter,
	reply *[]*sessions.ExternalSession) (err error) {
	return dS.dS.SessionSv1GetActiveSessions(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) GetActiveSessionsCount(ctx *context.Context, args *utils.SessionFilter,
	reply *int) (err error) {
	return dS.dS.SessionSv1GetActiveSessionsCount(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) ForceDisconnect(ctx *context.Context, args utils.SessionFilterWithEvent,
	reply *string) (err error) {
	return dS.dS.SessionSv1ForceDisconnect(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) AlterSessions(ctx *context.Context, args utils.SessionFilterWithEvent,
	reply *string) (err error) {
	return dS.dS.SessionSv1AlterSessions(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) GetPassiveSessions(ctx *context.Context, args *utils.SessionFilter,
	reply *[]*sessions.ExternalSession) (err error) {
	return dS.dS.SessionSv1GetPassiveSessions(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) GetPassiveSessionsCount(ctx *context.Context, args *utils.SessionFilter,
	reply *int) (err error) {
	return dS.dS.SessionSv1GetPassiveSessionsCount(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) ReplicateSessions(ctx *context.Context, args *dispatchers.ArgsReplicateSessionsWithAPIOpts,
	reply *string) (err error) {
	return dS.dS.SessionSv1ReplicateSessions(ctx, *args, reply)
}

func (dS *DispatcherSessionSv1) SetPassiveSession(ctx *context.Context, args *sessions.Session,
	reply *string) (err error) {
	return dS.dS.SessionSv1SetPassiveSession(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) ActivateSessions(ctx *context.Context, args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return dS.dS.SessionSv1ActivateSessions(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) DeactivateSessions(ctx *context.Context, args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return dS.dS.SessionSv1DeactivateSessions(ctx, args, reply)
}

func (dS *DispatcherSessionSv1) SyncSessions(ctx *context.Context, args *utils.TenantWithAPIOpts, rply *string) error {
	return dS.dS.SessionSv1SyncSessions(ctx, args, rply)
}

func (dS *DispatcherSessionSv1) STIRAuthenticate(ctx *context.Context, args *sessions.V1STIRAuthenticateArgs, reply *string) error {
	return dS.dS.SessionSv1STIRAuthenticate(ctx, args, reply)
}
func (dS *DispatcherSessionSv1) STIRIdentity(ctx *context.Context, args *sessions.V1STIRIdentityArgs, reply *string) error {
	return dS.dS.SessionSv1STIRIdentity(ctx, args, reply)
}

func NewDispatcherResponder(dps *dispatchers.DispatcherService) *DispatcherResponder {
	return &DispatcherResponder{dS: dps}
}

// Exports RPC from RLs
type DispatcherResponder struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherResponder) GetCost(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) error {
	return dS.dS.ResponderGetCost(ctx, args, reply)
}

func (dS *DispatcherResponder) Debit(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) error {
	return dS.dS.ResponderDebit(ctx, args, reply)
}

func (dS *DispatcherResponder) MaxDebit(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) error {
	return dS.dS.ResponderMaxDebit(ctx, args, reply)
}

func (dS *DispatcherResponder) RefundIncrements(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts, reply *engine.Account) error {
	return dS.dS.ResponderRefundIncrements(ctx, args, reply)
}

func (dS *DispatcherResponder) RefundRounding(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts, reply *engine.Account) error {
	return dS.dS.ResponderRefundRounding(ctx, args, reply)
}

func (dS *DispatcherResponder) GetMaxSessionTime(ctx *context.Context, args *engine.CallDescriptorWithAPIOpts, reply *time.Duration) error {
	return dS.dS.ResponderGetMaxSessionTime(ctx, args, reply)
}

func (dS *DispatcherResponder) Shutdown(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) error {
	return dS.dS.ResponderShutdown(ctx, args, reply)
}

func (dS *DispatcherResponder) GetCostOnRatingPlans(ctx *context.Context, arg *utils.GetCostOnRatingPlansArgs, reply *map[string]any) (err error) {
	return dS.dS.ResponderGetCostOnRatingPlans(ctx, arg, reply)
}
func (dS *DispatcherResponder) GetMaxSessionTimeOnAccounts(ctx *context.Context, arg *utils.GetMaxSessionTimeOnAccountsArgs, reply *map[string]any) (err error) {
	return dS.dS.ResponderGetMaxSessionTimeOnAccounts(ctx, arg, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherResponder) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.ResponderPing(ctx, args, reply)
}

func NewDispatcherCacheSv1(dps *dispatchers.DispatcherService) *DispatcherCacheSv1 {
	return &DispatcherCacheSv1{dS: dps}
}

// Exports RPC from CacheSv1
type DispatcherCacheSv1 struct {
	dS *dispatchers.DispatcherService
}

// GetItemIDs returns the IDs for cacheID with given prefix
func (dS *DispatcherCacheSv1) GetItemIDs(ctx *context.Context, args *utils.ArgsGetCacheItemIDsWithAPIOpts,
	reply *[]string) error {
	return dS.dS.CacheSv1GetItemIDs(ctx, args, reply)
}

// HasItem verifies the existence of an Item in cache
func (dS *DispatcherCacheSv1) HasItem(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *bool) error {
	return dS.dS.CacheSv1HasItem(ctx, args, reply)
}

// GetItem returns an Item from the cache
func (dS *DispatcherCacheSv1) GetItem(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *any) error {
	return dS.dS.CacheSv1GetItem(ctx, args, reply)
}

// GetItemWithRemote returns an Item from local or remote cache
func (dS *DispatcherCacheSv1) GetItemWithRemote(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *any) error {
	return dS.dS.CacheSv1GetItemWithRemote(ctx, args, reply)
}

// GetItemExpiryTime returns the expiryTime for an item
func (dS *DispatcherCacheSv1) GetItemExpiryTime(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *time.Time) error {
	return dS.dS.CacheSv1GetItemExpiryTime(ctx, args, reply)
}

// RemoveItem removes the Item with ID from cache
func (dS *DispatcherCacheSv1) RemoveItem(ctx *context.Context, args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *string) error {
	return dS.dS.CacheSv1RemoveItem(ctx, args, reply)
}

// RemoveItems removes the Item with ID from cache
func (dS *DispatcherCacheSv1) RemoveItems(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts,
	reply *string) error {
	return dS.dS.CacheSv1RemoveItems(ctx, args, reply)
}

// Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (dS *DispatcherCacheSv1) Clear(ctx *context.Context, args *utils.AttrCacheIDsWithAPIOpts,
	reply *string) error {
	return dS.dS.CacheSv1Clear(ctx, args, reply)
}

// GetCacheStats returns CacheStats filtered by cacheIDs
func (dS *DispatcherCacheSv1) GetCacheStats(ctx *context.Context, args *utils.AttrCacheIDsWithAPIOpts,
	reply *map[string]*ltcache.CacheStats) error {
	return dS.dS.CacheSv1GetCacheStats(ctx, args, reply)
}

// PrecacheStatus checks status of active precache processes
func (dS *DispatcherCacheSv1) PrecacheStatus(ctx *context.Context, args *utils.AttrCacheIDsWithAPIOpts, reply *map[string]string) error {
	return dS.dS.CacheSv1PrecacheStatus(ctx, args, reply)
}

// HasGroup checks existence of a group in cache
func (dS *DispatcherCacheSv1) HasGroup(ctx *context.Context, args *utils.ArgsGetGroupWithAPIOpts,
	reply *bool) (err error) {
	return dS.dS.CacheSv1HasGroup(ctx, args, reply)
}

// GetGroupItemIDs returns a list of itemIDs in a cache group
func (dS *DispatcherCacheSv1) GetGroupItemIDs(ctx *context.Context, args *utils.ArgsGetGroupWithAPIOpts,
	reply *[]string) (err error) {
	return dS.dS.CacheSv1GetGroupItemIDs(ctx, args, reply)
}

// RemoveGroup will remove a group and all items belonging to it from cache
func (dS *DispatcherCacheSv1) RemoveGroup(ctx *context.Context, args *utils.ArgsGetGroupWithAPIOpts,
	reply *string) (err error) {
	return dS.dS.CacheSv1RemoveGroup(ctx, args, reply)
}

// ReloadCache reloads cache from DB for a prefix or completely
func (dS *DispatcherCacheSv1) ReloadCache(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return dS.dS.CacheSv1ReloadCache(ctx, args, reply)
}

// LoadCache loads cache from DB for a prefix or completely
func (dS *DispatcherCacheSv1) LoadCache(ctx *context.Context, args *utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return dS.dS.CacheSv1LoadCache(ctx, args, reply)
}

// ReplicateSet replicate an item
func (dS *DispatcherCacheSv1) ReplicateSet(ctx *context.Context, args *utils.ArgCacheReplicateSet, reply *string) (err error) {
	return dS.dS.CacheSv1ReplicateSet(ctx, args, reply)
}

// ReplicateRemove remove an item
func (dS *DispatcherCacheSv1) ReplicateRemove(ctx *context.Context, args *utils.ArgCacheReplicateRemove, reply *string) (err error) {
	return dS.dS.CacheSv1ReplicateRemove(ctx, args, reply)
}

// Ping used to determinate if component is active
func (dS *DispatcherCacheSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.CacheSv1Ping(ctx, args, reply)
}

func NewDispatcherGuardianSv1(dps *dispatchers.DispatcherService) *DispatcherGuardianSv1 {
	return &DispatcherGuardianSv1{dS: dps}
}

// Exports RPC from CacheSv1
type DispatcherGuardianSv1 struct {
	dS *dispatchers.DispatcherService
}

// RemoteLock will lock a key from remote
func (dS *DispatcherGuardianSv1) RemoteLock(ctx *context.Context, attr *dispatchers.AttrRemoteLockWithAPIOpts, reply *string) (err error) {
	return dS.dS.GuardianSv1RemoteLock(ctx, *attr, reply)
}

// RemoteUnlock will unlock a key from remote based on reference ID
func (dS *DispatcherGuardianSv1) RemoteUnlock(ctx *context.Context, attr *dispatchers.AttrRemoteUnlockWithAPIOpts, reply *[]string) (err error) {
	return dS.dS.GuardianSv1RemoteUnlock(ctx, *attr, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherGuardianSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.GuardianSv1Ping(ctx, args, reply)
}

func NewDispatcherSchedulerSv1(dps *dispatchers.DispatcherService) *DispatcherSchedulerSv1 {
	return &DispatcherSchedulerSv1{dS: dps}
}

// Exports RPC from SchedulerSv1
type DispatcherSchedulerSv1 struct {
	dS *dispatchers.DispatcherService
}

// Reload reloads scheduler instructions
func (dS *DispatcherSchedulerSv1) Reload(ctx *context.Context, attr *utils.CGREvent, reply *string) (err error) {
	return dS.dS.SchedulerSv1Reload(ctx, attr, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherSchedulerSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.SchedulerSv1Ping(ctx, args, reply)
}

// ExecuteActions execute an actionPlan or multiple actionsPlans between a time interval
func (dS *DispatcherSchedulerSv1) ExecuteActions(ctx *context.Context, args *utils.AttrsExecuteActions, reply *string) error {
	return dS.dS.SchedulerSv1ExecuteActions(ctx, args, reply)
}

// ExecuteActionPlans execute multiple actionPlans one by one
func (dS *DispatcherSchedulerSv1) ExecuteActionPlans(ctx *context.Context, args *utils.AttrsExecuteActionPlans, reply *string) (err error) {
	return dS.dS.SchedulerSv1ExecuteActionPlans(ctx, args, reply)
}

func NewDispatcherSv1(dS *dispatchers.DispatcherService) *DispatcherSv1 {
	return &DispatcherSv1{dS: dS}
}

type DispatcherSv1 struct {
	dS *dispatchers.DispatcherService
}

// GetProfileForEvent returns the matching dispatcher profile for the provided event
func (dSv1 DispatcherSv1) GetProfilesForEvent(ctx *context.Context, ev *utils.CGREvent,
	dPrfl *engine.DispatcherProfiles) error {
	return dSv1.dS.DispatcherSv1GetProfilesForEvent(ctx, ev, dPrfl)
}

func (dS *DispatcherSv1) RemoteStatus(ctx *context.Context, args *cores.V1StatusParams, reply *map[string]any) (err error) {
	return dS.dS.DispatcherSv1RemoteStatus(ctx, args, reply)
}

func (dS *DispatcherSv1) RemotePing(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	return dS.dS.DispatcherSv1RemotePing(ctx, args, reply)
}

func (dS *DispatcherSv1) RemoteSleep(ctx *context.Context, args *utils.DurationArgs, reply *string) (err error) {
	return dS.dS.DispatcherSv1RemoteSleep(ctx, args, reply)
}

/*
func (dSv1 DispatcherSv1) Apier(ctx *context.Context,args *utils.MethodParameters, reply *any) (err error) {
	return dSv1.dS.V1Apier(ctx,new(APIerSv1), args, reply)
}
*/

func (rS *DispatcherSv1) Ping(ctx *context.Context, ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

func NewDispatcherSCDRsV1(dps *dispatchers.DispatcherService) *DispatcherSCDRsV1 {
	return &DispatcherSCDRsV1{dS: dps}
}

// Exports RPC from CDRsV1
type DispatcherSCDRsV1 struct {
	dS *dispatchers.DispatcherService
}

// Ping used to detreminate if component is active
func (dS *DispatcherSCDRsV1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.CDRsV1Ping(ctx, args, reply)
}

func (dS *DispatcherSCDRsV1) GetCDRs(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *[]*engine.CDR) error {
	return dS.dS.CDRsV1GetCDRs(ctx, args, reply)
}

func (dS *DispatcherSCDRsV1) GetCDRsCount(ctx *context.Context, args *utils.RPCCDRsFilterWithAPIOpts, reply *int64) error {
	return dS.dS.CDRsV1GetCDRsCount(ctx, args, reply)
}

func (dS *DispatcherSCDRsV1) StoreSessionCost(ctx *context.Context, args *engine.AttrCDRSStoreSMCost, reply *string) error {
	return dS.dS.CDRsV1StoreSessionCost(ctx, args, reply)
}

func (dS *DispatcherSCDRsV1) RateCDRs(ctx *context.Context, args *engine.ArgRateCDRs, reply *string) error {
	return dS.dS.CDRsV1RateCDRs(ctx, args, reply)
}

func (dS *DispatcherSCDRsV1) ProcessExternalCDR(ctx *context.Context, args *engine.ExternalCDRWithAPIOpts, reply *string) error {
	return dS.dS.CDRsV1ProcessExternalCDR(ctx, args, reply)
}

func (dS *DispatcherSCDRsV1) ProcessEvent(ctx *context.Context, args *engine.ArgV1ProcessEvent, reply *string) error {
	return dS.dS.CDRsV1ProcessEvent(ctx, args, reply)
}

func (dS *DispatcherSCDRsV1) ProcessCDR(ctx *context.Context, args *engine.CDRWithAPIOpts, reply *string) error {
	return dS.dS.CDRsV1ProcessCDR(ctx, args, reply)
}

func NewDispatcherSServiceManagerV1(dps *dispatchers.DispatcherService) *DispatcherSServiceManagerV1 {
	return &DispatcherSServiceManagerV1{dS: dps}
}

// Exports RPC from ServiceManagerV1
type DispatcherSServiceManagerV1 struct {
	dS *dispatchers.DispatcherService
}

// Ping used to detreminate if component is active
func (dS *DispatcherSServiceManagerV1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.ServiceManagerV1Ping(ctx, args, reply)
}
func (dS *DispatcherSServiceManagerV1) StartService(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error {
	return dS.dS.ServiceManagerV1StartService(ctx, *args, reply)
}
func (dS *DispatcherSServiceManagerV1) StopService(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error {
	return dS.dS.ServiceManagerV1StopService(ctx, *args, reply)
}
func (dS *DispatcherSServiceManagerV1) ServiceStatus(ctx *context.Context, args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error {
	return dS.dS.ServiceManagerV1ServiceStatus(ctx, *args, reply)
}

func NewDispatcherConfigSv1(dps *dispatchers.DispatcherService) *DispatcherConfigSv1 {
	return &DispatcherConfigSv1{dS: dps}
}

// Exports RPC from CDRsV1
type DispatcherConfigSv1 struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherConfigSv1) GetConfig(ctx *context.Context, args *config.SectionWithAPIOpts, reply *map[string]any) (err error) {
	return dS.dS.ConfigSv1GetConfig(ctx, args, reply)
}

func (dS *DispatcherConfigSv1) ReloadConfig(ctx *context.Context, args *config.ReloadArgs, reply *string) (err error) {
	return dS.dS.ConfigSv1ReloadConfig(ctx, args, reply)
}

func (dS *DispatcherConfigSv1) SetConfig(ctx *context.Context, args *config.SetConfigArgs, reply *string) (err error) {
	return dS.dS.ConfigSv1SetConfig(ctx, args, reply)
}

func (dS *DispatcherConfigSv1) SetConfigFromJSON(ctx *context.Context, args *config.SetConfigFromJSONArgs, reply *string) (err error) {
	return dS.dS.ConfigSv1SetConfigFromJSON(ctx, args, reply)
}
func (dS *DispatcherConfigSv1) GetConfigAsJSON(ctx *context.Context, args *config.SectionWithAPIOpts, reply *string) (err error) {
	return dS.dS.ConfigSv1GetConfigAsJSON(ctx, args, reply)
}

func NewDispatcherRALsV1(dps *dispatchers.DispatcherService) *DispatcherRALsV1 {
	return &DispatcherRALsV1{dS: dps}
}

// Exports RPC from RLs
type DispatcherRALsV1 struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherRALsV1) GetRatingPlansCost(ctx *context.Context, args *utils.RatingPlanCostArg, reply *dispatchers.RatingPlanCost) error {
	return dS.dS.RALsV1GetRatingPlansCost(ctx, args, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherRALsV1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.RALsV1Ping(ctx, args, reply)
}

// DispatcherCoreSv1 exports RPC from CoreSv1
type DispatcherCoreSv1 struct {
	dS *dispatchers.DispatcherService
}

func NewDispatcherCoreSv1(dps *dispatchers.DispatcherService) *DispatcherCoreSv1 {
	return &DispatcherCoreSv1{dS: dps}
}

func (dS *DispatcherCoreSv1) Status(ctx *context.Context, params *cores.V1StatusParams, reply *map[string]any) error {
	return dS.dS.CoreSv1Status(ctx, params, reply)
}

func (dS *DispatcherCoreSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.CoreSv1Ping(ctx, args, reply)
}

func (dS *DispatcherCoreSv1) Sleep(ctx *context.Context, args *utils.DurationArgs, reply *string) error {
	return dS.dS.CoreSv1Sleep(ctx, args, reply)
}

func (dS *DispatcherCoreSv1) StartCPUProfiling(ctx *context.Context, args *utils.DirectoryArgs, reply *string) error {
	return dS.dS.CoreSv1StartCPUProfiling(ctx, args, reply)
}

func (dS *DispatcherCoreSv1) StopCPUProfiling(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) error {
	return dS.dS.CoreSv1StopCPUProfiling(ctx, args, reply)
}

func (dS *DispatcherCoreSv1) StartMemoryProfiling(ctx *context.Context, params cores.MemoryProfilingParams, reply *string) error {
	return dS.dS.CoreSv1StartMemoryProfiling(ctx, params, reply)
}

func (dS *DispatcherCoreSv1) StopMemoryProfiling(ctx *context.Context, params utils.TenantWithAPIOpts, reply *string) error {
	return dS.dS.CoreSv1StopMemoryProfiling(ctx, params, reply)
}

func (dS *DispatcherCoreSv1) Panic(ctx *context.Context, args *utils.PanicMessageArgs, reply *string) error {
	return dS.dS.CoreSv1Panic(ctx, args, reply)
}

// DispatcherEeSv1 exports RPC from EeSv1.
type DispatcherEeSv1 struct {
	dS *dispatchers.DispatcherService
}

func NewDispatcherEeSv1(dps *dispatchers.DispatcherService) *DispatcherEeSv1 {
	return &DispatcherEeSv1{dS: dps}
}

func (dS *DispatcherEeSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.EeSv1Ping(ctx, args, reply)
}

func (dS *DispatcherEeSv1) ProcessEvent(ctx *context.Context, args *engine.CGREventWithEeIDs, reply *map[string]map[string]any) error {
	return dS.dS.EeSv1ProcessEvent(ctx, args, reply)
}

// DispatcherErSv1 exports RPC from ErSv1.
type DispatcherErSv1 struct {
	dS *dispatchers.DispatcherService
}

func NewDispatcherErSv1(dps *dispatchers.DispatcherService) *DispatcherErSv1 {
	return &DispatcherErSv1{dS: dps}
}

func (dS *DispatcherErSv1) Ping(ctx *context.Context, cgrEv *utils.CGREvent, reply *string) error {
	return dS.dS.ErSv1Ping(ctx, cgrEv, reply)
}

func (dS *DispatcherErSv1) ProcessEvent(ctx *context.Context, params ers.V1RunReaderParams, reply *string) error {
	return dS.dS.ErSv1RunReader(ctx, params, reply)
}

type DispatcherReplicatorSv1 struct {
	dS *dispatchers.DispatcherService
}

func NewDispatcherReplicatorSv1(dps *dispatchers.DispatcherService) *DispatcherReplicatorSv1 {
	return &DispatcherReplicatorSv1{dS: dps}
}

// Ping used to detreminate if component is active
func (dS *DispatcherReplicatorSv1) Ping(ctx *context.Context, args *utils.CGREvent, reply *string) error {
	return dS.dS.ReplicatorSv1Ping(ctx, args, reply)
}

// GetAccount
func (dS *DispatcherReplicatorSv1) GetAccount(ctx *context.Context, args *utils.StringWithAPIOpts, reply *engine.Account) error {
	return dS.dS.ReplicatorSv1GetAccount(ctx, args, reply)
}

// GetDestination
func (dS *DispatcherReplicatorSv1) GetDestination(ctx *context.Context, key *utils.StringWithAPIOpts, reply *engine.Destination) error {
	return dS.dS.ReplicatorSv1GetDestination(ctx, key, reply)
}

// GetReverseDestination
func (dS *DispatcherReplicatorSv1) GetReverseDestination(ctx *context.Context, key *utils.StringWithAPIOpts, reply *[]string) error {
	return dS.dS.ReplicatorSv1GetReverseDestination(ctx, key, reply)
}

// GetStatQueue
func (dS *DispatcherReplicatorSv1) GetStatQueue(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) error {
	return dS.dS.ReplicatorSv1GetStatQueue(ctx, tntID, reply)
}

// GetFilter
func (dS *DispatcherReplicatorSv1) GetFilter(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Filter) error {
	return dS.dS.ReplicatorSv1GetFilter(ctx, tntID, reply)
}

// GetThreshold
func (dS *DispatcherReplicatorSv1) GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Threshold) error {
	return dS.dS.ReplicatorSv1GetThreshold(ctx, tntID, reply)
}

// GetThresholdProfile
func (dS *DispatcherReplicatorSv1) GetThresholdProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) error {
	return dS.dS.ReplicatorSv1GetThresholdProfile(ctx, tntID, reply)
}

// GetStatQueueProfile
func (dS *DispatcherReplicatorSv1) GetStatQueueProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) error {
	return dS.dS.ReplicatorSv1GetStatQueueProfile(ctx, tntID, reply)
}

// GetTiming
func (dS *DispatcherReplicatorSv1) GetTiming(ctx *context.Context, id *utils.StringWithAPIOpts, reply *utils.TPTiming) error {
	return dS.dS.ReplicatorSv1GetTiming(ctx, id, reply)
}

// GetResource
func (dS *DispatcherReplicatorSv1) GetResource(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	return dS.dS.ReplicatorSv1GetResource(ctx, tntID, reply)
}

// GetResourceProfile
func (dS *DispatcherReplicatorSv1) GetResourceProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) error {
	return dS.dS.ReplicatorSv1GetResourceProfile(ctx, tntID, reply)
}

// GetActionTriggers
func (dS *DispatcherReplicatorSv1) GetActionTriggers(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.ActionTriggers) error {
	return dS.dS.ReplicatorSv1GetActionTriggers(ctx, id, reply)
}

// GetSharedGroup
func (dS *DispatcherReplicatorSv1) GetSharedGroup(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.SharedGroup) error {
	return dS.dS.ReplicatorSv1GetSharedGroup(ctx, id, reply)
}

// GetActions
func (dS *DispatcherReplicatorSv1) GetActions(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.Actions) error {
	return dS.dS.ReplicatorSv1GetActions(ctx, id, reply)
}

// GetActionPlan
func (dS *DispatcherReplicatorSv1) GetActionPlan(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.ActionPlan) error {
	return dS.dS.ReplicatorSv1GetActionPlan(ctx, id, reply)
}

// GetAllActionPlans
func (dS *DispatcherReplicatorSv1) GetAllActionPlans(ctx *context.Context, args *utils.StringWithAPIOpts, reply *map[string]*engine.ActionPlan) error {
	return dS.dS.ReplicatorSv1GetAllActionPlans(ctx, args, reply)
}

// GetAccountActionPlans
func (dS *DispatcherReplicatorSv1) GetAccountActionPlans(ctx *context.Context, id *utils.StringWithAPIOpts, reply *[]string) error {
	return dS.dS.ReplicatorSv1GetAccountActionPlans(ctx, id, reply)
}

// GetRatingPlan
func (dS *DispatcherReplicatorSv1) GetRatingPlan(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.RatingPlan) error {
	return dS.dS.ReplicatorSv1GetRatingPlan(ctx, id, reply)
}

// GetRatingProfile
func (dS *DispatcherReplicatorSv1) GetRatingProfile(ctx *context.Context, id *utils.StringWithAPIOpts, reply *engine.RatingProfile) error {
	return dS.dS.ReplicatorSv1GetRatingProfile(ctx, id, reply)
}

// GetRouteProfile
func (dS *DispatcherReplicatorSv1) GetRouteProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) error {
	return dS.dS.ReplicatorSv1GetRouteProfile(ctx, tntID, reply)
}

// GetAttributeProfile
func (dS *DispatcherReplicatorSv1) GetAttributeProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) error {
	return dS.dS.ReplicatorSv1GetAttributeProfile(ctx, tntID, reply)
}

// GetChargerProfile
func (dS *DispatcherReplicatorSv1) GetChargerProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) error {
	return dS.dS.ReplicatorSv1GetChargerProfile(ctx, tntID, reply)
}

// GetDispatcherProfile
func (dS *DispatcherReplicatorSv1) GetDispatcherProfile(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) error {
	return dS.dS.ReplicatorSv1GetDispatcherProfile(ctx, tntID, reply)
}

// GetDispatcherHost
func (dS *DispatcherReplicatorSv1) GetDispatcherHost(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) error {
	return dS.dS.ReplicatorSv1GetDispatcherHost(ctx, tntID, reply)
}

// GetItemLoadIDs
func (dS *DispatcherReplicatorSv1) GetItemLoadIDs(ctx *context.Context, itemID *utils.StringWithAPIOpts, reply *map[string]int64) error {
	return dS.dS.ReplicatorSv1GetItemLoadIDs(ctx, itemID, reply)
}

//finished all the above

// SetThresholdProfile
func (dS *DispatcherReplicatorSv1) SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetThresholdProfile(ctx, args, reply)
}

// SetThreshold
func (dS *DispatcherReplicatorSv1) SetThreshold(ctx *context.Context, args *engine.ThresholdWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetThreshold(ctx, args, reply)
}

// SetDestination
func (dS *DispatcherReplicatorSv1) SetDestination(ctx *context.Context, args *engine.DestinationWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetDestination(ctx, args, reply)
}

// SetAccount
func (dS *DispatcherReplicatorSv1) SetAccount(ctx *context.Context, args *engine.AccountWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetAccount(ctx, args, reply)
}

// SetReverseDestination
func (dS *DispatcherReplicatorSv1) SetReverseDestination(ctx *context.Context, args *engine.DestinationWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetReverseDestination(ctx, args, reply)
}

// SetStatQueue
func (dS *DispatcherReplicatorSv1) SetStatQueue(ctx *context.Context, args *engine.StatQueueWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetStatQueue(ctx, args, reply)
}

// SetFilter
func (dS *DispatcherReplicatorSv1) SetFilter(ctx *context.Context, args *engine.FilterWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetFilter(ctx, args, reply)
}

// SetStatQueueProfile
func (dS *DispatcherReplicatorSv1) SetStatQueueProfile(ctx *context.Context, args *engine.StatQueueProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetStatQueueProfile(ctx, args, reply)
}

// SetTiming
func (dS *DispatcherReplicatorSv1) SetTiming(ctx *context.Context, args *utils.TPTimingWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetTiming(ctx, args, reply)
}

// SetResource
func (dS *DispatcherReplicatorSv1) SetResource(ctx *context.Context, args *engine.ResourceWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetResource(ctx, args, reply)
}

// SetResourceProfile
func (dS *DispatcherReplicatorSv1) SetResourceProfile(ctx *context.Context, args *engine.ResourceProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetResourceProfile(ctx, args, reply)
}

// SetActionTriggers
func (dS *DispatcherReplicatorSv1) SetActionTriggers(ctx *context.Context, args *engine.SetActionTriggersArgWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetActionTriggers(ctx, args, reply)
}

// SetSharedGroup
func (dS *DispatcherReplicatorSv1) SetSharedGroup(ctx *context.Context, args *engine.SharedGroupWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetSharedGroup(ctx, args, reply)
}

// SetActions
func (dS *DispatcherReplicatorSv1) SetActions(ctx *context.Context, args *engine.SetActionsArgsWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetActions(ctx, args, reply)
}

// SetRatingPlan
func (dS *DispatcherReplicatorSv1) SetRatingPlan(ctx *context.Context, args *engine.RatingPlanWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetRatingPlan(ctx, args, reply)
}

// SetRatingProfile
func (dS *DispatcherReplicatorSv1) SetRatingProfile(ctx *context.Context, args *engine.RatingProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetRatingProfile(ctx, args, reply)
}

// SetRouteProfile
func (dS *DispatcherReplicatorSv1) SetRouteProfile(ctx *context.Context, args *engine.RouteProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetRouteProfile(ctx, args, reply)
}

// SetAttributeProfile
func (dS *DispatcherReplicatorSv1) SetAttributeProfile(ctx *context.Context, args *engine.AttributeProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetAttributeProfile(ctx, args, reply)
}

// SetChargerProfile
func (dS *DispatcherReplicatorSv1) SetChargerProfile(ctx *context.Context, args *engine.ChargerProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetChargerProfile(ctx, args, reply)
}

// SetDispatcherProfile
func (dS *DispatcherReplicatorSv1) SetDispatcherProfile(ctx *context.Context, args *engine.DispatcherProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetDispatcherProfile(ctx, args, reply)
}

// SetActionPlan
func (dS *DispatcherReplicatorSv1) SetActionPlan(ctx *context.Context, args *engine.SetActionPlanArgWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetActionPlan(ctx, args, reply)
}

// SetAccountActionPlans
func (dS *DispatcherReplicatorSv1) SetAccountActionPlans(ctx *context.Context, args *engine.SetAccountActionPlansArgWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetAccountActionPlans(ctx, args, reply)
}

// SetDispatcherHost
func (dS *DispatcherReplicatorSv1) SetDispatcherHost(ctx *context.Context, args *engine.DispatcherHostWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetDispatcherHost(ctx, args, reply)
}

// RemoveThreshold
func (dS *DispatcherReplicatorSv1) RemoveThreshold(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveThreshold(ctx, args, reply)
}

// SetLoadIDs
func (dS *DispatcherReplicatorSv1) SetLoadIDs(ctx *context.Context, args *utils.LoadIDsWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetLoadIDs(ctx, args, reply)
}

// RemoveDestination
func (dS *DispatcherReplicatorSv1) RemoveDestination(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveDestination(ctx, args, reply)
}

// RemoveAccount
func (dS *DispatcherReplicatorSv1) RemoveAccount(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveAccount(ctx, args, reply)
}

// RemoveStatQueue
func (dS *DispatcherReplicatorSv1) RemoveStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveStatQueue(ctx, args, reply)
}

// RemoveFilter
func (dS *DispatcherReplicatorSv1) RemoveFilter(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveFilter(ctx, args, reply)
}

// RemoveThresholdProfile
func (dS *DispatcherReplicatorSv1) RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveThresholdProfile(ctx, args, reply)
}

// RemoveStatQueueProfile
func (dS *DispatcherReplicatorSv1) RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveStatQueueProfile(ctx, args, reply)
}

// RemoveTiming
func (dS *DispatcherReplicatorSv1) RemoveTiming(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveTiming(ctx, args, reply)
}

// RemoveResource
func (dS *DispatcherReplicatorSv1) RemoveResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveResource(ctx, args, reply)
}

// RemoveResourceProfile
func (dS *DispatcherReplicatorSv1) RemoveResourceProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveResourceProfile(ctx, args, reply)
}

// RemoveActionTriggers
func (dS *DispatcherReplicatorSv1) RemoveActionTriggers(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveActionTriggers(ctx, args, reply)
}

// RemoveSharedGroup
func (dS *DispatcherReplicatorSv1) RemoveSharedGroup(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveSharedGroup(ctx, args, reply)
}

// RemoveActions
func (dS *DispatcherReplicatorSv1) RemoveActions(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveActions(ctx, args, reply)
}

// RemoveActionPlan
func (dS *DispatcherReplicatorSv1) RemoveActionPlan(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveActionPlan(ctx, args, reply)
}

// RemAccountActionPlans
func (dS *DispatcherReplicatorSv1) RemAccountActionPlans(ctx *context.Context, args *engine.RemAccountActionPlansArgsWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemAccountActionPlans(ctx, args, reply)
}

// RemoveRatingPlan
func (dS *DispatcherReplicatorSv1) RemoveRatingPlan(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveRatingPlan(ctx, args, reply)
}

// RemoveRatingProfile
func (dS *DispatcherReplicatorSv1) RemoveRatingProfile(ctx *context.Context, args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveRatingProfile(ctx, args, reply)
}

// RemoveRouteProfile
func (dS *DispatcherReplicatorSv1) RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveRouteProfile(ctx, args, reply)
}

// RemoveAttributeProfile
func (dS *DispatcherReplicatorSv1) RemoveAttributeProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveAttributeProfile(ctx, args, reply)
}

// RemoveChargerProfile
func (dS *DispatcherReplicatorSv1) RemoveChargerProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveChargerProfile(ctx, args, reply)
}

// RemoveDispatcherProfile
func (dS *DispatcherReplicatorSv1) RemoveDispatcherProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveDispatcherProfile(ctx, args, reply)
}

// RemoveDispatcherHost
func (dS *DispatcherReplicatorSv1) RemoveDispatcherHost(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveDispatcherHost(ctx, args, reply)
}

// GetIndexes .
func (dS *DispatcherReplicatorSv1) GetIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error {
	return dS.dS.ReplicatorSv1GetIndexes(ctx, args, reply)
}

// SetIndexes .
func (dS *DispatcherReplicatorSv1) SetIndexes(ctx *context.Context, args *utils.SetIndexesArg, reply *string) error {
	return dS.dS.ReplicatorSv1SetIndexes(ctx, args, reply)
}

// RemoveIndexes .
func (dS *DispatcherReplicatorSv1) RemoveIndexes(ctx *context.Context, args *utils.GetIndexesArg, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveIndexes(ctx, args, reply)
}
