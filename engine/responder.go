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
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type Responder struct {
	ShdChan               *utils.SyncedChan
	Timeout               time.Duration
	Timezone              string
	MaxComputedUsage      map[string]time.Duration
	maxComputedUsageMutex sync.RWMutex // used for MaxComputedUsage reload
}

// SetMaxComputedUsage sets MaxComputedUsage, used for config reload (is thread safe)
func (rs *Responder) SetMaxComputedUsage(mx map[string]time.Duration) {
	rs.maxComputedUsageMutex.Lock()
	rs.MaxComputedUsage = make(map[string]time.Duration)
	for k, v := range mx {
		rs.MaxComputedUsage[k] = v
	}
	rs.maxComputedUsageMutex.Unlock()
}

// usageAllowed checks requested usage against configured MaxComputedUsage
func (rs *Responder) usageAllowed(tor string, reqUsage time.Duration) (allowed bool) {
	rs.maxComputedUsageMutex.RLock()
	mcu, has := rs.MaxComputedUsage[tor]
	if !has {
		mcu = rs.MaxComputedUsage[utils.MetaAny]
	}
	rs.maxComputedUsageMutex.RUnlock()
	if reqUsage <= mcu {
		allowed = true
	}
	return
}

/*
RPC method that provides the external RPC interface for getting the rating information.
*/
func (rs *Responder) GetCost(arg *CallDescriptorWithAPIOpts, reply *CallCost) (err error) {
	// RPC caching
	if arg.CgrID != utils.EmptyString && config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResponderGetCost, arg.CgrID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*CallCost)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	if arg.Tenant == "" {
		arg.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if arg.Category == "" {
		arg.Category = config.CgrConfig().GeneralCfg().DefaultCategory
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	if !rs.usageAllowed(arg.ToR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	r, e := guardian.Guardian.Guard(func() (interface{}, error) {
		return arg.GetCost()
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+arg.GetAccountKey())
	if r != nil {
		*reply = *r.(*CallCost)
	}
	if e != nil {
		return e
	}
	return
}

// GetCostOnRatingPlans is used by RouteS to calculate the cost
// Receive a list of RatingPlans and pick the first without error
func (rs *Responder) GetCostOnRatingPlans(arg *utils.GetCostOnRatingPlansArgs, reply *map[string]interface{}) (err error) {
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	for _, rp := range arg.RatingPlanIDs { // loop through RatingPlans until we find one without errors
		rPrfl := &RatingProfile{
			Id: utils.ConcatenatedKey(utils.MetaOut,
				tnt, utils.MetaTmp, arg.Subject),
			RatingPlanActivations: RatingPlanActivations{
				&RatingPlanActivation{
					ActivationTime: arg.SetupTime,
					RatingPlanId:   rp,
				},
			},
		}
		var cc *CallCost
		if _, errGuard := guardian.Guardian.Guard(func() (_ interface{}, errGuard error) { // prevent cache data concurrency

			// force cache set so it can be picked by calldescriptor for cost calculation
			if errGuard := Cache.Set(utils.CacheRatingProfilesTmp, rPrfl.Id, rPrfl, nil,
				true, utils.NonTransactional); errGuard != nil {
				return nil, errGuard
			}
			cd := &CallDescriptor{
				Category:      utils.MetaTmp,
				Tenant:        tnt,
				Subject:       arg.Subject,
				Account:       arg.Account,
				Destination:   arg.Destination,
				TimeStart:     arg.SetupTime,
				TimeEnd:       arg.SetupTime.Add(arg.Usage),
				DurationIndex: arg.Usage,
			}
			cc, err = cd.GetCost()
			if errGuard := Cache.Remove(utils.CacheRatingProfilesTmp, rPrfl.Id,
				true, utils.NonTransactional); errGuard != nil { // Remove here so we don't overload memory
				return nil, errGuard
			}
			return

		}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ConcatenatedKey(utils.CacheRatingProfilesTmp, rPrfl.Id)); errGuard != nil {
			return errGuard
		}

		if err != nil {
			if err != utils.ErrNotFound {
				return err
			}
			continue
		}
		*reply = map[string]interface{}{
			utils.Cost:         cc.Cost,
			utils.RatingPlanID: rp,
		}
		return nil
	}
	return
}

func (rs *Responder) Debit(arg *CallDescriptorWithAPIOpts, reply *CallCost) (err error) {
	// RPC caching
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if arg.CgrID != utils.EmptyString && config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResponderDebit, arg.CgrID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*CallCost)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	if !rs.usageAllowed(arg.ToR, arg.GetDuration()) {
		err = utils.ErrMaxUsageExceeded
		return
	}
	var r *CallCost
	if r, err = arg.Debit(); err != nil {
		return
	}
	if r != nil {
		*reply = *r
	}
	return
}

func (rs *Responder) MaxDebit(arg *CallDescriptorWithAPIOpts, reply *CallCost) (err error) {
	// RPC caching
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if arg.CgrID != utils.EmptyString && config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResponderMaxDebit, arg.CgrID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*CallCost)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	if !rs.usageAllowed(arg.ToR, arg.GetDuration()) {
		err = utils.ErrMaxUsageExceeded
		return
	}
	var r *CallCost
	if r, err = arg.MaxDebit(); err != nil {
		return
	}
	if r != nil {
		*reply = *r
	}
	return
}

func (rs *Responder) RefundIncrements(arg *CallDescriptorWithAPIOpts, reply *Account) (err error) {
	// RPC caching
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if arg.CgrID != utils.EmptyString && config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResponderRefundIncrements, arg.CgrID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*Account)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	if !rs.usageAllowed(arg.ToR, arg.GetDuration()) {
		err = utils.ErrMaxUsageExceeded
		return
	}
	var acnt *Account
	if acnt, err = arg.RefundIncrements(); err != nil {
		return
	}
	if acnt != nil {
		*reply = *acnt
	}
	return
}

func (rs *Responder) RefundRounding(arg *CallDescriptorWithAPIOpts, reply *float64) (err error) {
	// RPC caching
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if arg.CgrID != utils.EmptyString && config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResponderRefundRounding, arg.CgrID)
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)

		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*float64)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}

	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	if !rs.usageAllowed(arg.ToR, arg.GetDuration()) {
		err = utils.ErrMaxUsageExceeded
		return
	}
	err = arg.RefundRounding()
	return
}

func (rs *Responder) GetMaxSessionTime(arg *CallDescriptorWithAPIOpts, reply *time.Duration) (err error) {
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if arg.Subject == "" {
		arg.Subject = arg.Account
	}
	if !rs.usageAllowed(arg.ToR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	*reply, err = arg.GetMaxSessionDuration()
	return
}

func (rs *Responder) GetMaxSessionTimeOnAccounts(arg *utils.GetMaxSessionTimeOnAccountsArgs,
	reply *map[string]interface{}) (err error) {
	var maxDur time.Duration
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	for _, anctID := range arg.AccountIDs {
		cd := &CallDescriptor{
			Category:      utils.MetaRoutes,
			Tenant:        tnt,
			Subject:       arg.Subject,
			Account:       anctID,
			Destination:   arg.Destination,
			TimeStart:     arg.SetupTime,
			TimeEnd:       arg.SetupTime.Add(arg.Usage),
			DurationIndex: arg.Usage,
		}
		if maxDur, err = cd.GetMaxSessionDuration(); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> ignoring cost for account: %s, err: %s",
					utils.Responder, anctID, err.Error()))
		} else {
			*reply = map[string]interface{}{
				utils.CapMaxUsage:  maxDur,
				utils.Cost:         0.0,
				utils.AccountField: anctID,
			}
			return nil
		}
	}
	return
}

func (rs *Responder) Shutdown(arg *utils.TenantWithOpts, reply *string) (err error) {
	dm.DataDB().Close()
	cdrStorage.Close()
	defer rs.ShdChan.CloseOnce()
	*reply = "Done!"
	return
}

// Ping used to detreminate if component is active
func (chSv1 *Responder) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

func (rs *Responder) Call(serviceMethod string, args interface{}, reply interface{}) error {
	parts := strings.Split(serviceMethod, ".")
	if len(parts) != 2 {
		return utils.ErrNotImplemented
	}
	// get method
	method := reflect.ValueOf(rs).MethodByName(parts[1])
	if !method.IsValid() {
		return utils.ErrNotImplemented
	}
	// construct the params
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
