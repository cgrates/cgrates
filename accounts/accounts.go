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
func (aS *AccountS) matchingAccountForEvent(tnt string, cgrEv *utils.CGREvent, acntIDs []string) (acnt *utils.AccountProfile, err error) {
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
			utils.CacheActionProfilesFilterIndexes,
			tnt,
			aS.cfg.AccountSCfg().IndexedSelects,
			aS.cfg.AccountSCfg().NestedFields,
		); err != nil {
			return
		}
		acntIDs = actIDsMp.AsSlice()
	}
	for _, acntID := range acntIDs {
		var qAcnt *utils.AccountProfile
		if qAcnt, err = aS.dm.GetAccountProfile(tnt, acntID,
			true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if _, isDisabled := qAcnt.Opts[utils.Disabled]; isDisabled {
			continue
		}
		if qAcnt.ActivationInterval != nil && cgrEv.Time != nil &&
			!qAcnt.ActivationInterval.IsActiveAtTime(*cgrEv.Time) { // not active
			continue
		}
		var pass bool
		if pass, err = aS.fltrS.Pass(tnt, qAcnt.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		if acnt == nil || acnt.Weight < qAcnt.Weight {
			acnt = qAcnt
		}
	}
	if acnt == nil {
		return nil, utils.ErrNotFound
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

	for _, blncOper := range blncOpers {
		if usage.Big.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debit
		}
		var dbted *utils.Decimal
		if dbted, _, err = blncOper.debitUsage(usage, cgrEv); err != nil {
			if err == utils.ErrFilterNotPassingNoCaps {
				err = nil
				continue
			}
		}
		usage.Big = utils.SubstractBig(usage.Big, dbted.Big)
	}
	return
}

// V1MaxUsage returns the maximum usage for the event, based on matching Account
func (aS *AccountS) V1MaxUsage(args *utils.ArgsAccountForEvent, ec *utils.EventCharges) (err error) {
	var acnt *utils.AccountProfile
	if acnt, err = aS.matchingAccountForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}

	var procEC *utils.EventCharges
	if procEC, err = aS.accountProcessEvent(acnt, args.CGREvent); err != nil {
		return
	}

	*ec = *procEC
	return
}
