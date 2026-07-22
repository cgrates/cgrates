// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

type Responder struct {
	FilterS               *FilterS
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
func (rs *Responder) GetCost(ctx *context.Context, arg *CallDescriptorWithAPIOpts, reply *CallCost) (err error) {
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
	var r *CallCost
	guardian.Guardian.Guard(func() (_ error) {
		r, err = arg.GetCost()
		return
	}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.AccountPrefix+arg.GetAccountKey())
	if err != nil {
		return
	}
	if r != nil {
		*reply = *r
	}
	return
}

// GetCostOnRatingPlans is used by RouteS to calculate the cost
// Receive a list of RatingPlans and pick the first without error
func (rs *Responder) GetCostOnRatingPlans(ctx *context.Context, arg *utils.GetCostOnRatingPlansArgs, reply *map[string]any) (err error) {
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
		if errGuard := guardian.Guardian.Guard(func() (errGuard error) { // prevent cache data concurrency
			// force cache set so it can be picked by calldescriptor for cost calculation
			if errGuard := Cache.Set(utils.CacheRatingProfilesTmp, rPrfl.Id, rPrfl, nil,
				true, utils.NonTransactional); errGuard != nil {
				return errGuard
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
			return Cache.Remove(utils.CacheRatingProfilesTmp, rPrfl.Id,
				true, utils.NonTransactional) // Remove here so we don't overload memory

		}, config.CgrConfig().GeneralCfg().LockingTimeout, utils.ConcatenatedKey(utils.CacheRatingProfilesTmp, rPrfl.Id)); errGuard != nil {
			return errGuard
		}

		if err != nil {
			if err != utils.ErrNotFound {
				return err
			}
			continue
		}
		*reply = map[string]any{
			utils.Cost:         cc.Cost,
			utils.RatingPlanID: rp,
		}
		return nil
	}
	return
}

func (rs *Responder) Debit(ctx *context.Context, arg *CallDescriptorWithAPIOpts, reply *CallCost) (err error) {
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
	if ralsDryRun, exists := arg.APIOpts[utils.MetaRALsDryRun]; exists {
		if arg.DryRun, err = utils.IfaceAsBool(ralsDryRun); err != nil {
			return err
		}
	}
	var r *CallCost
	if r, err = arg.Debit(rs.FilterS); err != nil {
		return
	}
	if r != nil {
		*reply = *r
	}
	return
}

func (rs *Responder) DebitMonetary(ctx *context.Context, cdr *CDRWithAPIOpts, reply *CallCost) (err error) {
	// RPC caching
	if cdr.Tenant == utils.EmptyString {
		cdr.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if cdr.CGRID != utils.EmptyString && config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.ResponderDebit, cdr.CGRID)
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

	if cdr.Subject == "" {
		cdr.Subject = cdr.Account
	}
	timeStart := cdr.AnswerTime
	if timeStart.IsZero() { // Fix for FreeSWITCH unanswered calls
		timeStart = cdr.SetupTime
	}
	cd := &CallDescriptor{
		ToR:             cdr.ToR,
		Tenant:          cdr.Tenant,
		Category:        cdr.Category,
		Subject:         cdr.Subject,
		Account:         cdr.Account,
		Destination:     cdr.Destination,
		ExtraFields:     cdr.ExtraFields,
		TimeStart:       timeStart,
		TimeEnd:         timeStart.Add(cdr.Usage),
		DurationIndex:   cdr.Usage,
		PerformRounding: true,
	}
	if ralsDryRun, exists := cdr.APIOpts[utils.MetaRALsDryRun]; exists {
		if cd.DryRun, err = utils.IfaceAsBool(ralsDryRun); err != nil {
			return err
		}
	}
	var r *CallCost
	if r, err = cd.DebitMonetary(rs.FilterS, cdr.Cost); err != nil {
		return
	}
	if r != nil {
		*reply = *r
	}
	return
}

func (rs *Responder) MaxDebit(ctx *context.Context, arg *CallDescriptorWithAPIOpts, reply *CallCost) (err error) {
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
	if r, err = arg.MaxDebit(rs.FilterS); err != nil {
		return
	}
	if r != nil {
		*reply = *r
	}
	return
}

func (rs *Responder) RefundIncrements(ctx *context.Context, arg *CallDescriptorWithAPIOpts, reply *Account) (err error) {
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
	if acnt, err = arg.RefundIncrements(rs.FilterS); err != nil {
		return
	}
	if acnt != nil {
		*reply = *acnt
	}
	return
}

func (rs *Responder) RefundRounding(ctx *context.Context, arg *CallDescriptorWithAPIOpts, reply *Account) (err error) {
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
				*reply = *cachedResp.Result.(*Account)
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
	var acc *Account
	if acc, err = arg.RefundRounding(rs.FilterS); err != nil || acc == nil {
		return
	}
	*reply = *acc
	return
}

func (rs *Responder) GetMaxSessionTime(ctx *context.Context, arg *CallDescriptorWithAPIOpts, reply *time.Duration) (err error) {
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = config.CgrConfig().GeneralCfg().DefaultTenant
	}
	if arg.Subject == utils.EmptyString {
		arg.Subject = arg.Account
	}
	if !rs.usageAllowed(arg.ToR, arg.GetDuration()) {
		return utils.ErrMaxUsageExceeded
	}
	*reply, err = arg.GetMaxSessionDuration(rs.FilterS)
	return
}

func (rs *Responder) GetMaxSessionTimeOnAccounts(ctx *context.Context, arg *utils.GetMaxSessionTimeOnAccountsArgs,
	reply *map[string]any) (err error) {
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
		if maxDur, err = cd.GetMaxSessionDuration(rs.FilterS); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> ignoring cost for account: %s, err: %s",
					utils.Responder, anctID, err.Error()))
		} else {
			*reply = map[string]any{
				utils.CapMaxUsage:  maxDur,
				utils.Cost:         0.0,
				utils.AccountField: anctID,
			}
			return nil
		}
	}
	return
}

func (rs *Responder) Shutdown(ctx *context.Context, arg *utils.TenantWithAPIOpts, reply *string) (err error) {
	dm.DataDB().Close()
	cdrStorage.Close()
	defer rs.ShdChan.CloseOnce()
	*reply = "Done!"
	return
}
