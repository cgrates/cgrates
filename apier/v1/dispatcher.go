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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

// GetDispatcherProfile returns a Dispatcher Profile
func (apierSv1 *APIerSv1) GetDispatcherProfile(arg *utils.TenantID, reply *engine.DispatcherProfile) error {
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
func (apierSv1 *APIerSv1) GetDispatcherProfileIDs(tenantArg *utils.PaginatorWithTenant, dPrfIDs *[]string) error {
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
	APIOpts map[string]interface{}
}

// SetDispatcherProfile add/update a new Dispatcher Profile
func (apierSv1 *APIerSv1) SetDispatcherProfile(args *DispatcherWithAPIOpts, reply *string) error {
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
func (apierSv1 *APIerSv1) RemoveDispatcherProfile(arg *utils.TenantIDWithAPIOpts, reply *string) error {
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
	//handle caching for DispatcherProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheDispatcherProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetDispatcherHost returns a Dispatcher Host
func (apierSv1 *APIerSv1) GetDispatcherHost(arg *utils.TenantID, reply *engine.DispatcherHost) error {
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
func (apierSv1 *APIerSv1) GetDispatcherHostIDs(tenantArg *utils.PaginatorWithTenant, dPrfIDs *[]string) error {
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
func (apierSv1 *APIerSv1) SetDispatcherHost(args *engine.DispatcherHostWithAPIOpts, reply *string) error {
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
func (apierSv1 *APIerSv1) RemoveDispatcherHost(arg *utils.TenantIDWithAPIOpts, reply *string) error {
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
func (dT *DispatcherThresholdSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dT.dS.ThresholdSv1Ping(args, reply)
}

// GetThresholdsForEvent implements ThresholdSv1GetThresholdsForEvent
func (dT *DispatcherThresholdSv1) GetThresholdsForEvent(tntID *utils.CGREvent,
	t *engine.Thresholds) error {
	return dT.dS.ThresholdSv1GetThresholdsForEvent(tntID, t)
}

// ProcessEvent implements ThresholdSv1ProcessEvent
func (dT *DispatcherThresholdSv1) ProcessEvent(args *utils.CGREvent,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1ProcessEvent(args, tIDs)
}

func (dT *DispatcherThresholdSv1) GetThresholdIDs(args *utils.TenantWithAPIOpts,
	tIDs *[]string) error {
	return dT.dS.ThresholdSv1GetThresholdIDs(args, tIDs)
}

func (dT *DispatcherThresholdSv1) GetThreshold(args *utils.TenantIDWithAPIOpts,
	th *engine.Threshold) error {
	return dT.dS.ThresholdSv1GetThreshold(args, th)
}

func NewDispatcherStatSv1(dps *dispatchers.DispatcherService) *DispatcherStatSv1 {
	return &DispatcherStatSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherStatSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements StatSv1Ping
func (dSts *DispatcherStatSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dSts.dS.StatSv1Ping(args, reply)
}

// GetStatQueuesForEvent implements StatSv1GetStatQueuesForEvent
func (dSts *DispatcherStatSv1) GetStatQueuesForEvent(args *utils.CGREvent, reply *[]string) error {
	return dSts.dS.StatSv1GetStatQueuesForEvent(args, reply)
}

// GetQueueStringMetrics implements StatSv1GetQueueStringMetrics
func (dSts *DispatcherStatSv1) GetQueueStringMetrics(args *utils.TenantIDWithAPIOpts,
	reply *map[string]string) error {
	return dSts.dS.StatSv1GetQueueStringMetrics(args, reply)
}

func (dSts *DispatcherStatSv1) GetQueueFloatMetrics(args *utils.TenantIDWithAPIOpts,
	reply *map[string]float64) error {
	return dSts.dS.StatSv1GetQueueFloatMetrics(args, reply)
}

func (dSts *DispatcherStatSv1) GetQueueIDs(args *utils.TenantWithAPIOpts,
	reply *[]string) error {
	return dSts.dS.StatSv1GetQueueIDs(args, reply)
}

// GetQueueStringMetrics implements StatSv1ProcessEvent
func (dSts *DispatcherStatSv1) ProcessEvent(args *utils.CGREvent, reply *[]string) error {
	return dSts.dS.StatSv1ProcessEvent(args, reply)
}

func NewDispatcherResourceSv1(dps *dispatchers.DispatcherService) *DispatcherResourceSv1 {
	return &DispatcherResourceSv1{dRs: dps}
}

// Exports RPC from RLs
type DispatcherResourceSv1 struct {
	dRs *dispatchers.DispatcherService
}

// Ping implements ResourceSv1Ping
func (dRs *DispatcherResourceSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dRs.dRs.ResourceSv1Ping(args, reply)
}

// GetResourcesForEvent implements ResourceSv1GetResourcesForEvent
func (dRs *DispatcherResourceSv1) GetResourcesForEvent(args *utils.CGREvent,
	reply *engine.Resources) error {
	return dRs.dRs.ResourceSv1GetResourcesForEvent(args, reply)
}

func (dRs *DispatcherResourceSv1) GetResource(args *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	return dRs.dRs.ResourceSv1GetResource(args, reply)
}

func (dRs *DispatcherResourceSv1) GetResourceWithConfig(args *utils.TenantIDWithAPIOpts, reply *engine.ResourceWithConfig) error {
	return dRs.dRs.ResourceSv1GetResourceWithConfig(args, reply)
}

func (dRs *DispatcherResourceSv1) AuthorizeResources(args *utils.CGREvent,
	reply *string) error {
	return dRs.dRs.ResourceSv1AuthorizeResources(args, reply)
}

func (dRs *DispatcherResourceSv1) AllocateResources(args *utils.CGREvent,
	reply *string) error {
	return dRs.dRs.ResourceSv1AllocateResources(args, reply)
}

func (dRs *DispatcherResourceSv1) ReleaseResources(args *utils.CGREvent,
	reply *string) error {
	return dRs.dRs.ResourceSv1ReleaseResources(args, reply)
}

func NewDispatcherRouteSv1(dps *dispatchers.DispatcherService) *DispatcherRouteSv1 {
	return &DispatcherRouteSv1{dRoute: dps}
}

// Exports RPC from RouteS
type DispatcherRouteSv1 struct {
	dRoute *dispatchers.DispatcherService
}

// Ping implements RouteSv1Ping
func (dRoute *DispatcherRouteSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dRoute.dRoute.RouteSv1Ping(args, reply)
}

// GetRoutes implements RouteSv1GetRoutes
func (dRoute *DispatcherRouteSv1) GetRoutes(args *utils.CGREvent, reply *engine.SortedRoutesList) error {
	return dRoute.dRoute.RouteSv1GetRoutes(args, reply)
}

// GetRouteProfilesForEvent returns a list of route profiles that match for Event
func (dRoute *DispatcherRouteSv1) GetRouteProfilesForEvent(args *utils.CGREvent, reply *[]*engine.RouteProfile) error {
	return dRoute.dRoute.RouteSv1GetRouteProfilesForEvent(args, reply)
}

// GetRoutesList returns sorted list of routes for Event as a string slice
func (dRoute *DispatcherRouteSv1) GetRoutesList(args *utils.CGREvent, reply *[]string) error {
	return dRoute.dRoute.RouteSv1GetRoutesList(args, reply)
}

func NewDispatcherAttributeSv1(dps *dispatchers.DispatcherService) *DispatcherAttributeSv1 {
	return &DispatcherAttributeSv1{dA: dps}
}

// Exports RPC from RLs
type DispatcherAttributeSv1 struct {
	dA *dispatchers.DispatcherService
}

// Ping implements AttributeSv1Ping
func (dA *DispatcherAttributeSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dA.dA.AttributeSv1Ping(args, reply)
}

// GetAttributeForEvent implements AttributeSv1GetAttributeForEvent
func (dA *DispatcherAttributeSv1) GetAttributeForEvent(args *utils.CGREvent,
	reply *engine.AttributeProfile) error {
	return dA.dA.AttributeSv1GetAttributeForEvent(args, reply)
}

// ProcessEvent implements AttributeSv1ProcessEvent
func (dA *DispatcherAttributeSv1) ProcessEvent(args *utils.CGREvent,
	reply *engine.AttrSProcessEventReply) error {
	return dA.dA.AttributeSv1ProcessEvent(args, reply)
}

func NewDispatcherChargerSv1(dps *dispatchers.DispatcherService) *DispatcherChargerSv1 {
	return &DispatcherChargerSv1{dC: dps}
}

// Exports RPC from RLs
type DispatcherChargerSv1 struct {
	dC *dispatchers.DispatcherService
}

// Ping implements ChargerSv1Ping
func (dC *DispatcherChargerSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dC.dC.ChargerSv1Ping(args, reply)
}

// GetChargersForEvent implements ChargerSv1GetChargersForEvent
func (dC *DispatcherChargerSv1) GetChargersForEvent(args *utils.CGREvent,
	reply *engine.ChargerProfiles) (err error) {
	return dC.dC.ChargerSv1GetChargersForEvent(args, reply)
}

// ProcessEvent implements ChargerSv1ProcessEvent
func (dC *DispatcherChargerSv1) ProcessEvent(args *utils.CGREvent,
	reply *[]*engine.ChrgSProcessEventReply) (err error) {
	return dC.dC.ChargerSv1ProcessEvent(args, reply)
}

func NewDispatcherSessionSv1(dps *dispatchers.DispatcherService) *DispatcherSessionSv1 {
	return &DispatcherSessionSv1{dS: dps}
}

// Exports RPC from RLs
type DispatcherSessionSv1 struct {
	dS *dispatchers.DispatcherService
}

// Ping implements SessionSv1Ping
func (dS *DispatcherSessionSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.SessionSv1Ping(args, reply)
}

// AuthorizeEventWithDigest implements SessionSv1AuthorizeEventWithDigest
func (dS *DispatcherSessionSv1) AuthorizeEventWithDigest(args *sessions.V1AuthorizeArgs,
	reply *sessions.V1AuthorizeReplyWithDigest) error {
	return dS.dS.SessionSv1AuthorizeEventWithDigest(args, reply)
}

func (dS *DispatcherSessionSv1) AuthorizeEvent(args *sessions.V1AuthorizeArgs,
	reply *sessions.V1AuthorizeReply) error {
	return dS.dS.SessionSv1AuthorizeEvent(args, reply)
}

// InitiateSessionWithDigest implements SessionSv1InitiateSessionWithDigest
func (dS *DispatcherSessionSv1) InitiateSessionWithDigest(args *sessions.V1InitSessionArgs,
	reply *sessions.V1InitReplyWithDigest) (err error) {
	return dS.dS.SessionSv1InitiateSessionWithDigest(args, reply)
}

// InitiateSessionWithDigest implements SessionSv1InitiateSessionWithDigest
func (dS *DispatcherSessionSv1) InitiateSession(args *sessions.V1InitSessionArgs,
	reply *sessions.V1InitSessionReply) (err error) {
	return dS.dS.SessionSv1InitiateSession(args, reply)
}

// ProcessCDR implements SessionSv1ProcessCDR
func (dS *DispatcherSessionSv1) ProcessCDR(args *utils.CGREvent,
	reply *string) (err error) {
	return dS.dS.SessionSv1ProcessCDR(args, reply)
}

// ProcessMessage implements SessionSv1ProcessMessage
func (dS *DispatcherSessionSv1) ProcessMessage(args *sessions.V1ProcessMessageArgs,
	reply *sessions.V1ProcessMessageReply) (err error) {
	return dS.dS.SessionSv1ProcessMessage(args, reply)
}

// ProcessMessage implements SessionSv1ProcessMessage
func (dS *DispatcherSessionSv1) ProcessEvent(args *sessions.V1ProcessEventArgs,
	reply *sessions.V1ProcessEventReply) (err error) {
	return dS.dS.SessionSv1ProcessEvent(args, reply)
}

// GetCost implements SessionSv1GetCost
func (dS *DispatcherSessionSv1) GetCost(args *sessions.V1ProcessEventArgs,
	reply *sessions.V1GetCostReply) (err error) {
	return dS.dS.SessionSv1GetCost(args, reply)
}

// TerminateSession implements SessionSv1TerminateSession
func (dS *DispatcherSessionSv1) TerminateSession(args *sessions.V1TerminateSessionArgs,
	reply *string) (err error) {
	return dS.dS.SessionSv1TerminateSession(args, reply)
}

// UpdateSession implements SessionSv1UpdateSession
func (dS *DispatcherSessionSv1) UpdateSession(args *sessions.V1UpdateSessionArgs,
	reply *sessions.V1UpdateSessionReply) (err error) {
	return dS.dS.SessionSv1UpdateSession(args, reply)
}

func (dS *DispatcherSessionSv1) GetActiveSessions(args *utils.SessionFilter,
	reply *[]*sessions.ExternalSession) (err error) {
	return dS.dS.SessionSv1GetActiveSessions(args, reply)
}

func (dS *DispatcherSessionSv1) GetActiveSessionsCount(args *utils.SessionFilter,
	reply *int) (err error) {
	return dS.dS.SessionSv1GetActiveSessionsCount(args, reply)
}

func (dS *DispatcherSessionSv1) ForceDisconnect(args *utils.SessionFilter,
	reply *string) (err error) {
	return dS.dS.SessionSv1ForceDisconnect(args, reply)
}

func (dS *DispatcherSessionSv1) GetPassiveSessions(args *utils.SessionFilter,
	reply *[]*sessions.ExternalSession) (err error) {
	return dS.dS.SessionSv1GetPassiveSessions(args, reply)
}

func (dS *DispatcherSessionSv1) GetPassiveSessionsCount(args *utils.SessionFilter,
	reply *int) (err error) {
	return dS.dS.SessionSv1GetPassiveSessionsCount(args, reply)
}

func (dS *DispatcherSessionSv1) ReplicateSessions(args *dispatchers.ArgsReplicateSessionsWithAPIOpts,
	reply *string) (err error) {
	return dS.dS.SessionSv1ReplicateSessions(*args, reply)
}

func (dS *DispatcherSessionSv1) SetPassiveSession(args *sessions.Session,
	reply *string) (err error) {
	return dS.dS.SessionSv1SetPassiveSession(args, reply)
}

func (dS *DispatcherSessionSv1) ActivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return dS.dS.SessionSv1ActivateSessions(args, reply)
}

func (dS *DispatcherSessionSv1) DeactivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return dS.dS.SessionSv1DeactivateSessions(args, reply)
}

func (dS *DispatcherSessionSv1) SyncSessions(args *utils.TenantWithAPIOpts, rply *string) error {
	return dS.dS.SessionSv1SyncSessions(args, rply)
}

func (dS *DispatcherSessionSv1) STIRAuthenticate(args *sessions.V1STIRAuthenticateArgs, reply *string) error {
	return dS.dS.SessionSv1STIRAuthenticate(args, reply)
}
func (dS *DispatcherSessionSv1) STIRIdentity(args *sessions.V1STIRIdentityArgs, reply *string) error {
	return dS.dS.SessionSv1STIRIdentity(args, reply)
}

func NewDispatcherResponder(dps *dispatchers.DispatcherService) *DispatcherResponder {
	return &DispatcherResponder{dS: dps}
}

// Exports RPC from RLs
type DispatcherResponder struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherResponder) GetCost(args *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) error {
	return dS.dS.ResponderGetCost(args, reply)
}

func (dS *DispatcherResponder) Debit(args *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) error {
	return dS.dS.ResponderDebit(args, reply)
}

func (dS *DispatcherResponder) MaxDebit(args *engine.CallDescriptorWithAPIOpts, reply *engine.CallCost) error {
	return dS.dS.ResponderMaxDebit(args, reply)
}

func (dS *DispatcherResponder) RefundIncrements(args *engine.CallDescriptorWithAPIOpts, reply *engine.Account) error {
	return dS.dS.ResponderRefundIncrements(args, reply)
}

func (dS *DispatcherResponder) RefundRounding(args *engine.CallDescriptorWithAPIOpts, reply *engine.Account) error {
	return dS.dS.ResponderRefundRounding(args, reply)
}

func (dS *DispatcherResponder) GetMaxSessionTime(args *engine.CallDescriptorWithAPIOpts, reply *time.Duration) error {
	return dS.dS.ResponderGetMaxSessionTime(args, reply)
}

func (dS *DispatcherResponder) Shutdown(args *utils.TenantWithAPIOpts, reply *string) error {
	return dS.dS.ResponderShutdown(args, reply)
}

func (dS *DispatcherResponder) GetCostOnRatingPlans(arg *utils.GetCostOnRatingPlansArgs, reply *map[string]interface{}) (err error) {
	return dS.dS.ResponderGetCostOnRatingPlans(arg, reply)
}
func (dS *DispatcherResponder) GetMaxSessionTimeOnAccounts(arg *utils.GetMaxSessionTimeOnAccountsArgs, reply *map[string]interface{}) (err error) {
	return dS.dS.ResponderGetMaxSessionTimeOnAccounts(arg, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherResponder) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.ResponderPing(args, reply)
}

func NewDispatcherCacheSv1(dps *dispatchers.DispatcherService) *DispatcherCacheSv1 {
	return &DispatcherCacheSv1{dS: dps}
}

// Exports RPC from CacheSv1
type DispatcherCacheSv1 struct {
	dS *dispatchers.DispatcherService
}

// GetItemIDs returns the IDs for cacheID with given prefix
func (dS *DispatcherCacheSv1) GetItemIDs(args *utils.ArgsGetCacheItemIDsWithAPIOpts,
	reply *[]string) error {
	return dS.dS.CacheSv1GetItemIDs(args, reply)
}

// HasItem verifies the existence of an Item in cache
func (dS *DispatcherCacheSv1) HasItem(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *bool) error {
	return dS.dS.CacheSv1HasItem(args, reply)
}

// GetItem returns an Item from the cache
func (dS *DispatcherCacheSv1) GetItem(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *interface{}) error {
	return dS.dS.CacheSv1GetItem(args, reply)
}

// GetItemWithRemote returns an Item from local or remote cache
func (dS *DispatcherCacheSv1) GetItemWithRemote(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *interface{}) error {
	return dS.dS.CacheSv1GetItemWithRemote(args, reply)
}

// GetItemExpiryTime returns the expiryTime for an item
func (dS *DispatcherCacheSv1) GetItemExpiryTime(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *time.Time) error {
	return dS.dS.CacheSv1GetItemExpiryTime(args, reply)
}

// RemoveItem removes the Item with ID from cache
func (dS *DispatcherCacheSv1) RemoveItem(args *utils.ArgsGetCacheItemWithAPIOpts,
	reply *string) error {
	return dS.dS.CacheSv1RemoveItem(args, reply)
}

// RemoveItems removes the Item with ID from cache
func (dS *DispatcherCacheSv1) RemoveItems(args *utils.AttrReloadCacheWithAPIOpts,
	reply *string) error {
	return dS.dS.CacheSv1RemoveItems(args, reply)
}

// Clear will clear partitions in the cache (nil fol all, empty slice for none)
func (dS *DispatcherCacheSv1) Clear(args *utils.AttrCacheIDsWithAPIOpts,
	reply *string) error {
	return dS.dS.CacheSv1Clear(args, reply)
}

// GetCacheStats returns CacheStats filtered by cacheIDs
func (dS *DispatcherCacheSv1) GetCacheStats(args *utils.AttrCacheIDsWithAPIOpts,
	reply *map[string]*ltcache.CacheStats) error {
	return dS.dS.CacheSv1GetCacheStats(args, reply)
}

// PrecacheStatus checks status of active precache processes
func (dS *DispatcherCacheSv1) PrecacheStatus(args *utils.AttrCacheIDsWithAPIOpts, reply *map[string]string) error {
	return dS.dS.CacheSv1PrecacheStatus(args, reply)
}

// HasGroup checks existence of a group in cache
func (dS *DispatcherCacheSv1) HasGroup(args *utils.ArgsGetGroupWithAPIOpts,
	reply *bool) (err error) {
	return dS.dS.CacheSv1HasGroup(args, reply)
}

// GetGroupItemIDs returns a list of itemIDs in a cache group
func (dS *DispatcherCacheSv1) GetGroupItemIDs(args *utils.ArgsGetGroupWithAPIOpts,
	reply *[]string) (err error) {
	return dS.dS.CacheSv1GetGroupItemIDs(args, reply)
}

// RemoveGroup will remove a group and all items belonging to it from cache
func (dS *DispatcherCacheSv1) RemoveGroup(args *utils.ArgsGetGroupWithAPIOpts,
	reply *string) (err error) {
	return dS.dS.CacheSv1RemoveGroup(args, reply)
}

// ReloadCache reloads cache from DB for a prefix or completely
func (dS *DispatcherCacheSv1) ReloadCache(args *utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return dS.dS.CacheSv1ReloadCache(args, reply)
}

// LoadCache loads cache from DB for a prefix or completely
func (dS *DispatcherCacheSv1) LoadCache(args *utils.AttrReloadCacheWithAPIOpts, reply *string) (err error) {
	return dS.dS.CacheSv1LoadCache(args, reply)
}

// ReplicateSet replicate an item
func (dS *DispatcherCacheSv1) ReplicateSet(args *utils.ArgCacheReplicateSet, reply *string) (err error) {
	return dS.dS.CacheSv1ReplicateSet(args, reply)
}

// ReplicateRemove remove an item
func (dS *DispatcherCacheSv1) ReplicateRemove(args *utils.ArgCacheReplicateRemove, reply *string) (err error) {
	return dS.dS.CacheSv1ReplicateRemove(args, reply)
}

// Ping used to determinate if component is active
func (dS *DispatcherCacheSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.CacheSv1Ping(args, reply)
}

func NewDispatcherGuardianSv1(dps *dispatchers.DispatcherService) *DispatcherGuardianSv1 {
	return &DispatcherGuardianSv1{dS: dps}
}

// Exports RPC from CacheSv1
type DispatcherGuardianSv1 struct {
	dS *dispatchers.DispatcherService
}

// RemoteLock will lock a key from remote
func (dS *DispatcherGuardianSv1) RemoteLock(attr *dispatchers.AttrRemoteLockWithAPIOpts, reply *string) (err error) {
	return dS.dS.GuardianSv1RemoteLock(*attr, reply)
}

// RemoteUnlock will unlock a key from remote based on reference ID
func (dS *DispatcherGuardianSv1) RemoteUnlock(attr *dispatchers.AttrRemoteUnlockWithAPIOpts, reply *[]string) (err error) {
	return dS.dS.GuardianSv1RemoteUnlock(*attr, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherGuardianSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.GuardianSv1Ping(args, reply)
}

func NewDispatcherSchedulerSv1(dps *dispatchers.DispatcherService) *DispatcherSchedulerSv1 {
	return &DispatcherSchedulerSv1{dS: dps}
}

// Exports RPC from SchedulerSv1
type DispatcherSchedulerSv1 struct {
	dS *dispatchers.DispatcherService
}

// Reload reloads scheduler instructions
func (dS *DispatcherSchedulerSv1) Reload(attr *utils.CGREvent, reply *string) (err error) {
	return dS.dS.SchedulerSv1Reload(attr, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherSchedulerSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.SchedulerSv1Ping(args, reply)
}

// ExecuteActions execute an actionPlan or multiple actionsPlans between a time interval
func (dS *DispatcherSchedulerSv1) ExecuteActions(args *utils.AttrsExecuteActions, reply *string) error {
	return dS.dS.SchedulerSv1ExecuteActions(args, reply)
}

// ExecuteActionPlans execute multiple actionPlans one by one
func (dS *DispatcherSchedulerSv1) ExecuteActionPlans(args *utils.AttrsExecuteActionPlans, reply *string) (err error) {
	return dS.dS.SchedulerSv1ExecuteActionPlans(args, reply)
}

func NewDispatcherSv1(dS *dispatchers.DispatcherService) *DispatcherSv1 {
	return &DispatcherSv1{dS: dS}
}

type DispatcherSv1 struct {
	dS *dispatchers.DispatcherService
}

// GetProfileForEvent returns the matching dispatcher profile for the provided event
func (dSv1 DispatcherSv1) GetProfilesForEvent(ev *utils.CGREvent,
	dPrfl *engine.DispatcherProfiles) error {
	return dSv1.dS.V1GetProfilesForEvent(ev, dPrfl)
}

func (dS *DispatcherSv1) RemoteStatus(args *utils.TenantWithAPIOpts, reply *map[string]interface{}) (err error) {
	return dS.dS.DispatcherSv1RemoteStatus(args, reply)
}

func (dS *DispatcherSv1) RemotePing(args *utils.CGREvent, reply *string) (err error) {
	return dS.dS.DispatcherSv1RemotePing(args, reply)
}

func (dS *DispatcherSv1) RemoteSleep(args *utils.DurationArgs, reply *string) (err error) {
	return dS.dS.DispatcherSv1RemoteSleep(args, reply)
}

/*
func (dSv1 DispatcherSv1) Apier(args *utils.MethodParameters, reply *interface{}) (err error) {
	return dSv1.dS.V1Apier(new(APIerSv1), args, reply)
}
*/

func (rS *DispatcherSv1) Ping(ign *utils.CGREvent, reply *string) error {
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
func (dS *DispatcherSCDRsV1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.CDRsV1Ping(args, reply)
}

func (dS *DispatcherSCDRsV1) GetCDRs(args *utils.RPCCDRsFilterWithAPIOpts, reply *[]*engine.CDR) error {
	return dS.dS.CDRsV1GetCDRs(args, reply)
}

func (dS *DispatcherSCDRsV1) GetCDRsCount(args *utils.RPCCDRsFilterWithAPIOpts, reply *int64) error {
	return dS.dS.CDRsV1GetCDRsCount(args, reply)
}

func (dS *DispatcherSCDRsV1) StoreSessionCost(args *engine.AttrCDRSStoreSMCost, reply *string) error {
	return dS.dS.CDRsV1StoreSessionCost(args, reply)
}

func (dS *DispatcherSCDRsV1) RateCDRs(args *engine.ArgRateCDRs, reply *string) error {
	return dS.dS.CDRsV1RateCDRs(args, reply)
}

func (dS *DispatcherSCDRsV1) ProcessExternalCDR(args *engine.ExternalCDRWithAPIOpts, reply *string) error {
	return dS.dS.CDRsV1ProcessExternalCDR(args, reply)
}

func (dS *DispatcherSCDRsV1) ProcessEvent(args *engine.ArgV1ProcessEvent, reply *string) error {
	return dS.dS.CDRsV1ProcessEvent(args, reply)
}

func (dS *DispatcherSCDRsV1) ProcessCDR(args *engine.CDRWithAPIOpts, reply *string) error {
	return dS.dS.CDRsV1ProcessCDR(args, reply)
}

func NewDispatcherSServiceManagerV1(dps *dispatchers.DispatcherService) *DispatcherSServiceManagerV1 {
	return &DispatcherSServiceManagerV1{dS: dps}
}

// Exports RPC from ServiceManagerV1
type DispatcherSServiceManagerV1 struct {
	dS *dispatchers.DispatcherService
}

// Ping used to detreminate if component is active
func (dS *DispatcherSServiceManagerV1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.ServiceManagerV1Ping(args, reply)
}
func (dS *DispatcherSServiceManagerV1) StartService(args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error {
	return dS.dS.ServiceManagerV1StartService(*args, reply)
}
func (dS *DispatcherSServiceManagerV1) StopService(args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error {
	return dS.dS.ServiceManagerV1StopService(*args, reply)
}
func (dS *DispatcherSServiceManagerV1) ServiceStatus(args *dispatchers.ArgStartServiceWithAPIOpts, reply *string) error {
	return dS.dS.ServiceManagerV1ServiceStatus(*args, reply)
}

func NewDispatcherConfigSv1(dps *dispatchers.DispatcherService) *DispatcherConfigSv1 {
	return &DispatcherConfigSv1{dS: dps}
}

// Exports RPC from CDRsV1
type DispatcherConfigSv1 struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherConfigSv1) GetConfig(args *config.SectionWithAPIOpts, reply *map[string]interface{}) (err error) {
	return dS.dS.ConfigSv1GetConfig(args, reply)
}

func (dS *DispatcherConfigSv1) ReloadConfig(args *config.ReloadArgs, reply *string) (err error) {
	return dS.dS.ConfigSv1ReloadConfig(args, reply)
}

func (dS *DispatcherConfigSv1) SetConfig(args *config.SetConfigArgs, reply *string) (err error) {
	return dS.dS.ConfigSv1SetConfig(args, reply)
}

func (dS *DispatcherConfigSv1) SetConfigFromJSON(args *config.SetConfigFromJSONArgs, reply *string) (err error) {
	return dS.dS.ConfigSv1SetConfigFromJSON(args, reply)
}
func (dS *DispatcherConfigSv1) GetConfigAsJSON(args *config.SectionWithAPIOpts, reply *string) (err error) {
	return dS.dS.ConfigSv1GetConfigAsJSON(args, reply)
}

func NewDispatcherRALsV1(dps *dispatchers.DispatcherService) *DispatcherRALsV1 {
	return &DispatcherRALsV1{dS: dps}
}

// Exports RPC from RLs
type DispatcherRALsV1 struct {
	dS *dispatchers.DispatcherService
}

func (dS *DispatcherRALsV1) GetRatingPlansCost(args *utils.RatingPlanCostArg, reply *dispatchers.RatingPlanCost) error {
	return dS.dS.RALsV1GetRatingPlansCost(args, reply)
}

// Ping used to detreminate if component is active
func (dS *DispatcherRALsV1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.RALsV1Ping(args, reply)
}

// DispatcherCoreSv1 exports RPC from CoreSv1
type DispatcherCoreSv1 struct {
	dS *dispatchers.DispatcherService
}

func NewDispatcherCoreSv1(dps *dispatchers.DispatcherService) *DispatcherCoreSv1 {
	return &DispatcherCoreSv1{dS: dps}
}

func (dS *DispatcherCoreSv1) Status(args *utils.TenantWithAPIOpts, reply *map[string]interface{}) error {
	return dS.dS.CoreSv1Status(args, reply)
}

func (dS *DispatcherCoreSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.CoreSv1Ping(args, reply)
}

func (dS *DispatcherCoreSv1) Sleep(args *utils.DurationArgs, reply *string) error {
	return dS.dS.CoreSv1Sleep(args, reply)
}

func (dS *DispatcherCoreSv1) StartCPUProfiling(args *utils.DirectoryArgs, reply *string) error {
	return dS.dS.CoreSv1StartCPUProfiling(args, reply)
}

func (dS *DispatcherCoreSv1) StopCPUProfiling(args *utils.TenantWithAPIOpts, reply *string) error {
	return dS.dS.CoreSv1StopCPUProfiling(args, reply)
}

func (dS *DispatcherCoreSv1) StartMemoryProfiling(args *utils.MemoryPrf, reply *string) error {
	return dS.dS.CoreSv1StartMemoryProfiling(args, reply)
}

func (dS *DispatcherCoreSv1) StopMemoryProfiling(args *utils.TenantWithAPIOpts, reply *string) error {
	return dS.dS.CoreSv1StopMemoryProfiling(args, reply)
}

func (dS *DispatcherCoreSv1) Panic(args *utils.PanicMessageArgs, reply *string) error {
	return dS.dS.CoreSv1Panic(args, reply)
}

// DispatcherCoreSv1 exports RPC from CoreSv1
type DispatcherEeSv1 struct {
	dS *dispatchers.DispatcherService
}

func NewDispatcherEeSv1(dps *dispatchers.DispatcherService) *DispatcherEeSv1 {
	return &DispatcherEeSv1{dS: dps}
}

func (dS *DispatcherEeSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.EeSv1Ping(args, reply)
}

func (dS *DispatcherEeSv1) ProcessEvent(args *engine.CGREventWithEeIDs, reply *map[string]map[string]interface{}) error {
	return dS.dS.EeSv1ProcessEvent(args, reply)
}

type DispatcherReplicatorSv1 struct {
	dS *dispatchers.DispatcherService
}

func NewDispatcherReplicatorSv1(dps *dispatchers.DispatcherService) *DispatcherReplicatorSv1 {
	return &DispatcherReplicatorSv1{dS: dps}
}

// Ping used to detreminate if component is active
func (dS *DispatcherReplicatorSv1) Ping(args *utils.CGREvent, reply *string) error {
	return dS.dS.ReplicatorSv1Ping(args, reply)
}

// GetAccount
func (dS *DispatcherReplicatorSv1) GetAccount(args *utils.StringWithAPIOpts, reply *engine.Account) error {
	return dS.dS.ReplicatorSv1GetAccount(args, reply)
}

// GetDestination
func (dS *DispatcherReplicatorSv1) GetDestination(key *utils.StringWithAPIOpts, reply *engine.Destination) error {
	return dS.dS.ReplicatorSv1GetDestination(key, reply)
}

// GetReverseDestination
func (dS *DispatcherReplicatorSv1) GetReverseDestination(key *utils.StringWithAPIOpts, reply *[]string) error {
	return dS.dS.ReplicatorSv1GetReverseDestination(key, reply)
}

// GetStatQueue
func (dS *DispatcherReplicatorSv1) GetStatQueue(tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueue) error {
	return dS.dS.ReplicatorSv1GetStatQueue(tntID, reply)
}

// GetFilter
func (dS *DispatcherReplicatorSv1) GetFilter(tntID *utils.TenantIDWithAPIOpts, reply *engine.Filter) error {
	return dS.dS.ReplicatorSv1GetFilter(tntID, reply)
}

// GetThreshold
func (dS *DispatcherReplicatorSv1) GetThreshold(tntID *utils.TenantIDWithAPIOpts, reply *engine.Threshold) error {
	return dS.dS.ReplicatorSv1GetThreshold(tntID, reply)
}

// GetThresholdProfile
func (dS *DispatcherReplicatorSv1) GetThresholdProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) error {
	return dS.dS.ReplicatorSv1GetThresholdProfile(tntID, reply)
}

// GetStatQueueProfile
func (dS *DispatcherReplicatorSv1) GetStatQueueProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) error {
	return dS.dS.ReplicatorSv1GetStatQueueProfile(tntID, reply)
}

// GetTiming
func (dS *DispatcherReplicatorSv1) GetTiming(id *utils.StringWithAPIOpts, reply *utils.TPTiming) error {
	return dS.dS.ReplicatorSv1GetTiming(id, reply)
}

// GetResource
func (dS *DispatcherReplicatorSv1) GetResource(tntID *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	return dS.dS.ReplicatorSv1GetResource(tntID, reply)
}

// GetResourceProfile
func (dS *DispatcherReplicatorSv1) GetResourceProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.ResourceProfile) error {
	return dS.dS.ReplicatorSv1GetResourceProfile(tntID, reply)
}

// GetActionTriggers
func (dS *DispatcherReplicatorSv1) GetActionTriggers(id *utils.StringWithAPIOpts, reply *engine.ActionTriggers) error {
	return dS.dS.ReplicatorSv1GetActionTriggers(id, reply)
}

// GetSharedGroup
func (dS *DispatcherReplicatorSv1) GetSharedGroup(id *utils.StringWithAPIOpts, reply *engine.SharedGroup) error {
	return dS.dS.ReplicatorSv1GetSharedGroup(id, reply)
}

// GetActions
func (dS *DispatcherReplicatorSv1) GetActions(id *utils.StringWithAPIOpts, reply *engine.Actions) error {
	return dS.dS.ReplicatorSv1GetActions(id, reply)
}

// GetActionPlan
func (dS *DispatcherReplicatorSv1) GetActionPlan(id *utils.StringWithAPIOpts, reply *engine.ActionPlan) error {
	return dS.dS.ReplicatorSv1GetActionPlan(id, reply)
}

// GetAllActionPlans
func (dS *DispatcherReplicatorSv1) GetAllActionPlans(args *utils.StringWithAPIOpts, reply *map[string]*engine.ActionPlan) error {
	return dS.dS.ReplicatorSv1GetAllActionPlans(args, reply)
}

// GetAccountActionPlans
func (dS *DispatcherReplicatorSv1) GetAccountActionPlans(id *utils.StringWithAPIOpts, reply *[]string) error {
	return dS.dS.ReplicatorSv1GetAccountActionPlans(id, reply)
}

// GetRatingPlan
func (dS *DispatcherReplicatorSv1) GetRatingPlan(id *utils.StringWithAPIOpts, reply *engine.RatingPlan) error {
	return dS.dS.ReplicatorSv1GetRatingPlan(id, reply)
}

// GetRatingProfile
func (dS *DispatcherReplicatorSv1) GetRatingProfile(id *utils.StringWithAPIOpts, reply *engine.RatingProfile) error {
	return dS.dS.ReplicatorSv1GetRatingProfile(id, reply)
}

// GetRouteProfile
func (dS *DispatcherReplicatorSv1) GetRouteProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.RouteProfile) error {
	return dS.dS.ReplicatorSv1GetRouteProfile(tntID, reply)
}

// GetAttributeProfile
func (dS *DispatcherReplicatorSv1) GetAttributeProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.AttributeProfile) error {
	return dS.dS.ReplicatorSv1GetAttributeProfile(tntID, reply)
}

// GetChargerProfile
func (dS *DispatcherReplicatorSv1) GetChargerProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.ChargerProfile) error {
	return dS.dS.ReplicatorSv1GetChargerProfile(tntID, reply)
}

// GetDispatcherProfile
func (dS *DispatcherReplicatorSv1) GetDispatcherProfile(tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherProfile) error {
	return dS.dS.ReplicatorSv1GetDispatcherProfile(tntID, reply)
}

// GetDispatcherHost
func (dS *DispatcherReplicatorSv1) GetDispatcherHost(tntID *utils.TenantIDWithAPIOpts, reply *engine.DispatcherHost) error {
	return dS.dS.ReplicatorSv1GetDispatcherHost(tntID, reply)
}

// GetItemLoadIDs
func (dS *DispatcherReplicatorSv1) GetItemLoadIDs(itemID *utils.StringWithAPIOpts, reply *map[string]int64) error {
	return dS.dS.ReplicatorSv1GetItemLoadIDs(itemID, reply)
}

//finished all the above

// SetThresholdProfile
func (dS *DispatcherReplicatorSv1) SetThresholdProfile(args *engine.ThresholdProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetThresholdProfile(args, reply)
}

// SetThreshold
func (dS *DispatcherReplicatorSv1) SetThreshold(args *engine.ThresholdWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetThreshold(args, reply)
}

// SetDestination
func (dS *DispatcherReplicatorSv1) SetDestination(args *engine.DestinationWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetDestination(args, reply)
}

// SetAccount
func (dS *DispatcherReplicatorSv1) SetAccount(args *engine.AccountWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetAccount(args, reply)
}

// SetReverseDestination
func (dS *DispatcherReplicatorSv1) SetReverseDestination(args *engine.DestinationWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetReverseDestination(args, reply)
}

// SetStatQueue
func (dS *DispatcherReplicatorSv1) SetStatQueue(args *engine.StatQueueWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetStatQueue(args, reply)
}

// SetFilter
func (dS *DispatcherReplicatorSv1) SetFilter(args *engine.FilterWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetFilter(args, reply)
}

// SetStatQueueProfile
func (dS *DispatcherReplicatorSv1) SetStatQueueProfile(args *engine.StatQueueProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetStatQueueProfile(args, reply)
}

// SetTiming
func (dS *DispatcherReplicatorSv1) SetTiming(args *utils.TPTimingWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetTiming(args, reply)
}

// SetResource
func (dS *DispatcherReplicatorSv1) SetResource(args *engine.ResourceWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetResource(args, reply)
}

// SetResourceProfile
func (dS *DispatcherReplicatorSv1) SetResourceProfile(args *engine.ResourceProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetResourceProfile(args, reply)
}

// SetActionTriggers
func (dS *DispatcherReplicatorSv1) SetActionTriggers(args *engine.SetActionTriggersArgWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetActionTriggers(args, reply)
}

// SetSharedGroup
func (dS *DispatcherReplicatorSv1) SetSharedGroup(args *engine.SharedGroupWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetSharedGroup(args, reply)
}

// SetActions
func (dS *DispatcherReplicatorSv1) SetActions(args *engine.SetActionsArgsWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetActions(args, reply)
}

// SetRatingPlan
func (dS *DispatcherReplicatorSv1) SetRatingPlan(args *engine.RatingPlanWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetRatingPlan(args, reply)
}

// SetRatingProfile
func (dS *DispatcherReplicatorSv1) SetRatingProfile(args *engine.RatingProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetRatingProfile(args, reply)
}

// SetRouteProfile
func (dS *DispatcherReplicatorSv1) SetRouteProfile(args *engine.RouteProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetRouteProfile(args, reply)
}

// SetAttributeProfile
func (dS *DispatcherReplicatorSv1) SetAttributeProfile(args *engine.AttributeProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetAttributeProfile(args, reply)
}

// SetChargerProfile
func (dS *DispatcherReplicatorSv1) SetChargerProfile(args *engine.ChargerProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetChargerProfile(args, reply)
}

// SetDispatcherProfile
func (dS *DispatcherReplicatorSv1) SetDispatcherProfile(args *engine.DispatcherProfileWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetDispatcherProfile(args, reply)
}

// SetActionPlan
func (dS *DispatcherReplicatorSv1) SetActionPlan(args *engine.SetActionPlanArgWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetActionPlan(args, reply)
}

// SetAccountActionPlans
func (dS *DispatcherReplicatorSv1) SetAccountActionPlans(args *engine.SetAccountActionPlansArgWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetAccountActionPlans(args, reply)
}

// SetDispatcherHost
func (dS *DispatcherReplicatorSv1) SetDispatcherHost(args *engine.DispatcherHostWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetDispatcherHost(args, reply)
}

// RemoveThreshold
func (dS *DispatcherReplicatorSv1) RemoveThreshold(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveThreshold(args, reply)
}

// SetLoadIDs
func (dS *DispatcherReplicatorSv1) SetLoadIDs(args *utils.LoadIDsWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1SetLoadIDs(args, reply)
}

// RemoveDestination
func (dS *DispatcherReplicatorSv1) RemoveDestination(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveDestination(args, reply)
}

// RemoveAccount
func (dS *DispatcherReplicatorSv1) RemoveAccount(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveAccount(args, reply)
}

// RemoveStatQueue
func (dS *DispatcherReplicatorSv1) RemoveStatQueue(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveStatQueue(args, reply)
}

// RemoveFilter
func (dS *DispatcherReplicatorSv1) RemoveFilter(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveFilter(args, reply)
}

// RemoveThresholdProfile
func (dS *DispatcherReplicatorSv1) RemoveThresholdProfile(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveThresholdProfile(args, reply)
}

// RemoveStatQueueProfile
func (dS *DispatcherReplicatorSv1) RemoveStatQueueProfile(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveStatQueueProfile(args, reply)
}

// RemoveTiming
func (dS *DispatcherReplicatorSv1) RemoveTiming(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveTiming(args, reply)
}

// RemoveResource
func (dS *DispatcherReplicatorSv1) RemoveResource(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveResource(args, reply)
}

// RemoveResourceProfile
func (dS *DispatcherReplicatorSv1) RemoveResourceProfile(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveResourceProfile(args, reply)
}

// RemoveActionTriggers
func (dS *DispatcherReplicatorSv1) RemoveActionTriggers(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveActionTriggers(args, reply)
}

// RemoveSharedGroup
func (dS *DispatcherReplicatorSv1) RemoveSharedGroup(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveSharedGroup(args, reply)
}

// RemoveActions
func (dS *DispatcherReplicatorSv1) RemoveActions(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveActions(args, reply)
}

// RemoveActionPlan
func (dS *DispatcherReplicatorSv1) RemoveActionPlan(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveActionPlan(args, reply)
}

// RemAccountActionPlans
func (dS *DispatcherReplicatorSv1) RemAccountActionPlans(args *engine.RemAccountActionPlansArgsWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemAccountActionPlans(args, reply)
}

// RemoveRatingPlan
func (dS *DispatcherReplicatorSv1) RemoveRatingPlan(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveRatingPlan(args, reply)
}

// RemoveRatingProfile
func (dS *DispatcherReplicatorSv1) RemoveRatingProfile(args *utils.StringWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveRatingProfile(args, reply)
}

// RemoveRouteProfile
func (dS *DispatcherReplicatorSv1) RemoveRouteProfile(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveRouteProfile(args, reply)
}

// RemoveAttributeProfile
func (dS *DispatcherReplicatorSv1) RemoveAttributeProfile(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveAttributeProfile(args, reply)
}

// RemoveChargerProfile
func (dS *DispatcherReplicatorSv1) RemoveChargerProfile(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveChargerProfile(args, reply)
}

// RemoveDispatcherProfile
func (dS *DispatcherReplicatorSv1) RemoveDispatcherProfile(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveDispatcherProfile(args, reply)
}

// RemoveDispatcherHost
func (dS *DispatcherReplicatorSv1) RemoveDispatcherHost(args *utils.TenantIDWithAPIOpts, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveDispatcherHost(args, reply)
}

// GetIndexes .
func (dS *DispatcherReplicatorSv1) GetIndexes(args *utils.GetIndexesArg, reply *map[string]utils.StringSet) error {
	return dS.dS.ReplicatorSv1GetIndexes(args, reply)
}

// SetIndexes .
func (dS *DispatcherReplicatorSv1) SetIndexes(args *utils.SetIndexesArg, reply *string) error {
	return dS.dS.ReplicatorSv1SetIndexes(args, reply)
}

// RemoveIndexes .
func (dS *DispatcherReplicatorSv1) RemoveIndexes(args *utils.GetIndexesArg, reply *string) error {
	return dS.dS.ReplicatorSv1RemoveIndexes(args, reply)
}
