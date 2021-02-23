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
	"strings"
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

// matchingAccountsForEvent returns the matched Accounts for the given event
// if lked option is passed, each AccountProfile will be also locked
//   so it becomes responsibility of upper layers to release the lock
func (aS *AccountS) matchingAccountsForEvent(tnt string, cgrEv *utils.CGREvent,
	acntIDs []string, lked bool) (acnts utils.AccountProfilesWithWeight, err error) {
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
			cacheKey := utils.ConcatenatedKey(utils.CacheAccountProfiles, tnt, acntID)
			refID = guardian.Guardian.GuardIDs("",
				aS.cfg.GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		}
		var qAcnt *utils.AccountProfile
		if qAcnt, err = aS.dm.GetAccountProfile(tnt, acntID); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if _, isDisabled := qAcnt.Opts[utils.Disabled]; isDisabled ||
			(qAcnt.ActivationInterval != nil && cgrEv.Time != nil &&
				!qAcnt.ActivationInterval.IsActiveAtTime(*cgrEv.Time)) { // not active
			guardian.Guardian.UnguardIDs(refID)
			continue
		}
		var pass bool
		if pass, err = aS.fltrS.Pass(tnt, qAcnt.FilterIDs, evNm); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			return
		} else if !pass {
			guardian.Guardian.UnguardIDs(refID)
			continue
		}
		var weight float64
		if weight, err = engine.WeightFromDynamics(qAcnt.Weights,
			aS.fltrS, cgrEv.Tenant, evNm); err != nil {
			return
		}
		acnts = append(acnts, &utils.AccountProfileWithWeight{qAcnt, weight, refID})
	}
	if len(acnts) == 0 {
		return nil, utils.ErrNotFound
	}
	acnts.Sort()
	return
}

// accountDebitAbstracts will debit the usage out of an Account
func (aS *AccountS) accountDebitAbstracts(acnt *utils.AccountProfile, usage *decimal.Big,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {
	// Find balances matching event
	blcsWithWeight := make(utils.BalancesWithWeight, 0, len(acnt.Balances))
	for _, blnCfg := range acnt.Balances {
		var weight float64
		if weight, err = engine.WeightFromDynamics(blnCfg.Weights,
			aS.fltrS, cgrEv.Tenant, cgrEv.AsDataProvider()); err != nil {
			return
		}
		blcsWithWeight = append(blcsWithWeight, &utils.BalanceWithWeight{blnCfg, weight})
	}
	blcsWithWeight.Sort()
	var blncOpers []balanceOperator
	if blncOpers, err = newBalanceOperators(blcsWithWeight.Balances(), aS.fltrS, aS.connMgr,
		aS.cfg.AccountSCfg().AttributeSConns, aS.cfg.AccountSCfg().RateSConns); err != nil {
		return
	}

	for i, blncOper := range blncOpers {
		if i == 0 {
			ec = utils.NewEventCharges()
		}
		if usage.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debit
		}
		var ecDbt *utils.EventCharges
		if ecDbt, err = blncOper.debitAbstracts(new(decimal.Big).Copy(usage), cgrEv); err != nil {
			if err == utils.ErrFilterNotPassingNoCaps {
				err = nil
				continue
			}
			return
		}
		usage = utils.SubstractBig(usage, ecDbt.Usage.Big)
		ec.Merge(ecDbt)
	}
	return
}

// accountsDebitAbstracts will debit an usage out of multiple accounts
func (aS *AccountS) accountsDebitAbstracts(acnts []*utils.AccountProfileWithWeight,
	cgrEv *utils.CGREvent, store bool) (ec *utils.EventCharges, err error) {
	usage := decimal.New(int64(72*time.Hour), 0)
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
			usage = decimal.New(int64(usgEv), 0)
		}
	} else {
		usage = decimal.New(int64(usgEv), 0)
	}
	acntBkps := make([]utils.AccountBalancesBackup, len(acnts))
	for i, acnt := range acnts {
		if i == 0 {
			ec = utils.NewEventCharges()
		}
		if usage.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debits
		}
		acntBkps[i] = acnt.AccountProfile.AccountBalancesBackup()
		var ecDbt *utils.EventCharges
		if ecDbt, err = aS.accountDebitAbstracts(acnt.AccountProfile,
			new(decimal.Big).Copy(usage), cgrEv); err != nil {
			if store {
				restoreAccounts(aS.dm, acnts, acntBkps)
			}
			return
		}
		if store && acnt.AccountProfile.BalancesAltered(acntBkps[i]) {
			if err = aS.dm.SetAccountProfile(acnt.AccountProfile, false); err != nil {
				restoreAccounts(aS.dm, acnts, acntBkps)
				return
			}
		}
		usage = utils.SubstractBig(usage, ecDbt.Usage.Big)
		ec.Merge(ecDbt)
	}
	return
}

func (aS *AccountS) accountDebitCost(acnt *utils.AccountProfile,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {
	return
}

// V1AccountProfilesForEvent returns the matching AccountProfiles for Event
func (aS *AccountS) V1AccountProfilesForEvent(args *utils.ArgsAccountsForEvent, aps *[]*utils.AccountProfile) (err error) {
	var acnts utils.AccountProfilesWithWeight
	if acnts, err = aS.matchingAccountsForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, false); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*aps = acnts.AccountProfiles()
	return
}

// V1MaxUsage returns the maximum usage for the event, based on matching Accounts
func (aS *AccountS) V1MaxAbstracts(args *utils.ArgsAccountsForEvent, eEc *utils.ExtEventCharges) (err error) {
	var acnts utils.AccountProfilesWithWeight
	if acnts, err = aS.matchingAccountsForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer func() {
		for _, lkID := range acnts.LockIDs() {
			guardian.Guardian.UnguardIDs(lkID)
		}
	}()
	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebitAbstracts(acnts, args.CGREvent, false); err != nil {
		return
	}
	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}
	*eEc = *rcvEec
	return
}

// V1DebitAbstracts performs debit for the provided event
func (aS *AccountS) V1DebitAbstracts(args *utils.ArgsAccountsForEvent, eEc *utils.ExtEventCharges) (err error) {
	var acnts utils.AccountProfilesWithWeight
	if acnts, err = aS.matchingAccountsForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer func() {
		for _, lkID := range acnts.LockIDs() {
			guardian.Guardian.UnguardIDs(lkID)
		}
	}()

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebitAbstracts(acnts, args.CGREvent, true); err != nil {
		return
	}

	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}

	*eEc = *rcvEec
	return
}

// V1TopupBalance performs a topup for a specific account
func (aS *AccountS) V1UpdateBalance(args *utils.ArgsUpdateBalance, rply *string) (err error) {
	if args.AccountID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AccountID)
	}
	if len(args.Params) == 0 {
		return utils.NewErrMandatoryIeMissing("Params")
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = aS.cfg.GeneralCfg().DefaultTenant
	}
	if _, err = guardian.Guardian.Guard(func() (interface{}, error) {
		return nil, aS.updateBalance(tnt, args.AccountID, args.Params, args.Reset)
	}, aS.cfg.GeneralCfg().LockingTimeout,
		utils.ConcatenatedKey(utils.CacheAccountProfiles, tnt, args.AccountID)); err != nil {
		return
	}

	*rply = utils.OK
	return
}

func (aS *AccountS) updateBalance(tnt, acntID string, params []*utils.ArgsBalParams, reset bool) (err error) {
	var qAcnt *utils.AccountProfile
	if qAcnt, err = aS.dm.GetAccountProfile(tnt, acntID); err != nil {
		return
	}
	for _, param := range params {
		path := strings.Split(param.Path, utils.NestingSep)
		if len(path) < 3 {
			return fmt.Errorf("unsupported path:%s ", param.Path)
		}
		if path[0] != "~*balance" {
			return fmt.Errorf("unsupported field prefix: <%s>", path[0])
		}
		switch
		if bal, has := qAcnt.Balances[balID]; !has {
			qAcnt.Balances[balID] = utils.NewDefaultBalance(balID, value)
		} else if reset {
			bal.Units = value
		} else {
			bal.Units.Add(bal.Units.Big, value.Big)
		}
	}
	return aS.dm.SetAccountProfile(qAcnt, false)
}
