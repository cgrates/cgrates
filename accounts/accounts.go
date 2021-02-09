/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

package accounts

import (
	"fmt"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// NewAccountS instantiates the AccountS
func NewAccountS(cfg *config.CGRConfig, fltrS *engine.FilterS, connMgr *engine.ConnManager, dm *engine.DataManager) *AccountS {
	return &AccountS{cfg, fltrS, connMgr, dm}
}

// AccountS operates Accounts
type AccountS struct {
	cfg     *config.CGRConfig
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
	dm      *engine.DataManager
}

// ListenAndServe keeps the service alive
func (aS *AccountS) ListenAndServe(stopChan, cfgRld chan struct{}) {
	for {
		select {
		case <-stopChan:
			return
		case rld := <-cfgRld: // configuration was reloaded
			cfgRld <- rld
		}
	}
}

// Shutdown is called to shutdown the service
func (aS *AccountS) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.AccountS))
	return
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (aS *AccountS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(aS, serviceMethod, args, reply)
}

// matchingAccountForEvent returns the matched Account for the given event
// if lked option is passed, the AccountProfile will be also locked
//   so it becomes responsibility of upper layers to release the lock
func (aS *AccountS) matchingAccountForEvent(tnt string, cgrEv *utils.CGREvent,
	acntIDs []string, lked bool) (acnt *utils.AccountProfile, lkID string, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.Opts,
	}
	if len(acntIDs) == 0 {
		var actIDsMp utils.StringSet
		if actIDsMp, err = engine.MatchingItemIDsForEvent(
			evNm,
			aS.cfg.AccountSCfg().StringIndexedFields,
			aS.cfg.AccountSCfg().PrefixIndexedFields,
			aS.cfg.AccountSCfg().SuffixIndexedFields,
			aS.dm,
			utils.CacheAccountProfilesFilterIndexes,
			tnt,
			aS.cfg.AccountSCfg().IndexedSelects,
			aS.cfg.AccountSCfg().NestedFields,
		); err != nil {
			return
		}
		acntIDs = actIDsMp.AsSlice()
	}
	for _, acntID := range acntIDs {
		var refID string
		if lked {
			cacheKey := utils.ConcatenatedKey(utils.CacheAccountProfiles, acntID)
			refID = guardian.Guardian.GuardIDs("",
				aS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		}
		var qAcnt *utils.AccountProfile
		if qAcnt, err = aS.dm.GetAccountProfile(tnt, acntID,
			true, true, utils.NonTransactional); err != nil {
			if lked {
				guardian.Guardian.UnguardIDs(refID)
			}
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if _, isDisabled := qAcnt.Opts[utils.Disabled]; isDisabled ||
			(qAcnt.ActivationInterval != nil && cgrEv.Time != nil &&
				!qAcnt.ActivationInterval.IsActiveAtTime(*cgrEv.Time)) { // not active
			if lked {
				guardian.Guardian.UnguardIDs(refID)
			}
			continue
		}
		var pass bool
		if pass, err = aS.fltrS.Pass(tnt, qAcnt.FilterIDs, evNm); err != nil {
			if lked {
				guardian.Guardian.UnguardIDs(refID)
			}
			return
		} else if !pass {
			if lked {
				guardian.Guardian.UnguardIDs(refID)
			}
			continue
		}
		if acnt == nil || acnt.Weight < qAcnt.Weight {
			acnt = qAcnt
			if lked {
				if lkID != utils.EmptyString {
					guardian.Guardian.UnguardIDs(lkID)
				}
				lkID = refID
			}
		} else if lked {
			guardian.Guardian.UnguardIDs(refID)
		}
	}
	if acnt == nil {
		return nil, "", utils.ErrNotFound
	}
	return
}

// accountProcessEvent implements event processing by an Account
func (aS *AccountS) accountProcessEvent(acnt *utils.AccountProfile,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {
	var blncOpers []balanceOperator
	if blncOpers, err = newAccountBalanceOperators(acnt, aS.fltrS, aS.connMgr,
		aS.cfg.AccountSCfg().AttributeSConns, aS.cfg.AccountSCfg().RateSConns); err != nil {
		return
	}
	usage := utils.NewDecimal(int64(72*time.Hour), 0)
	var usgEv time.Duration
	if usgEv, err = cgrEv.FieldAsDuration(utils.Usage); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		// not found, try at opts level
		if usgEv, err = cgrEv.OptAsDuration(utils.MetaUsage); err != nil {
			if err != utils.ErrNotFound {
				return
			}
			err = nil
		} else { // found, overwrite usage
			usage.Big = decimal.New(int64(usgEv), 0)
		}
	} else {
		usage.Big = decimal.New(int64(usgEv), 0)
	}
	for i, blncOper := range blncOpers {
		if i == 0 {
			ec = utils.NewEventCharges()
		}
		if usage.Big.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debit
		}
		var ecDbt *utils.EventCharges
		if ecDbt, err = blncOper.debitUsage(usage.Clone(), cgrEv); err != nil {
			if err == utils.ErrFilterNotPassingNoCaps {
				err = nil
				continue
			}
			return
		}
		usage.Big = utils.SubstractBig(usage.Big, ecDbt.Usage)
		ec.Merge(ecDbt)
	}
	return
}

// V1AccountProfileForEvent returns the matching AccountProfile for Event
func (aS *AccountS) V1AccountProfileForEvent(args *utils.ArgsAccountForEvent, ap *utils.AccountProfile) (err error) {
	var acnt *utils.AccountProfile
	if acnt, _, err = aS.matchingAccountForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, false); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*ap = *acnt // ToDo: make sure we clone in RPC
	return
}

// V1MaxUsage returns the maximum usage for the event, based on matching Account
func (aS *AccountS) V1MaxUsage(args *utils.ArgsAccountForEvent, eEc *utils.ExtEventCharges) (err error) {
	var acnt *utils.AccountProfile
	var lkID string
	if acnt, lkID, err = aS.matchingAccountForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer guardian.Guardian.UnguardIDs(lkID)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountProcessEvent(acnt, args.CGREvent); err != nil {
		return
	}
	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}
	*eEc = *rcvEec
	return
}

// V1DebitUsage performs debit for the provided event
func (aS *AccountS) V1DebitUsage(args *utils.ArgsAccountForEvent, eEc *utils.ExtEventCharges) (err error) {
	var acnt *utils.AccountProfile
	var lkID string
	if acnt, lkID, err = aS.matchingAccountForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer guardian.Guardian.UnguardIDs(lkID)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountProcessEvent(acnt, args.CGREvent); err != nil {
		return
	}

	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}

	if err = aS.dm.SetAccountProfile(acnt, false); err != nil {
		return // no need of revert since we did not save
	}

	*eEc = *rcvEec
	return
}
