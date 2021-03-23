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
func (aS *AccountS) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.AccountS))
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
		utils.MetaOpts: cgrEv.APIOpts,
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
			unlockAccountProfiles(acnts) // in case of errors will not have unlocks in upper layers
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
			unlockAccountProfiles(acnts)
			return
		} else if !pass {
			guardian.Guardian.UnguardIDs(refID)
			continue
		}
		var weight float64
		if weight, err = engine.WeightFromDynamics(qAcnt.Weights,
			aS.fltrS, cgrEv.Tenant, evNm); err != nil {
			guardian.Guardian.UnguardIDs(refID)
			unlockAccountProfiles(acnts)
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

// accountsDebit will debit an usage out of multiple accounts
func (aS *AccountS) accountsDebit(acnts []*utils.AccountProfileWithWeight,
	cgrEv *utils.CGREvent, concretes, store bool) (ec *utils.EventCharges, err error) {
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
		if usage.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debits
		}
		acntBkps[i] = acnt.AccountProfile.AccountBalancesBackup()
		var ecDbt *utils.EventCharges
		if ecDbt, err = aS.accountDebit(acnt.AccountProfile,
			new(decimal.Big).Copy(usage), cgrEv, concretes); err != nil {
			if store {
				restoreAccounts(aS.dm, acnts, acntBkps)
			}
			return
		}
		if ecDbt == nil {
			continue
		}
		if ec == nil { // no debit performed yet
			ec = utils.NewEventCharges()
		}
		if store && acnt.AccountProfile.BalancesAltered(acntBkps[i]) {
			if err = aS.dm.SetAccountProfile(acnt.AccountProfile, false); err != nil {
				restoreAccounts(aS.dm, acnts, acntBkps)
				return
			}
		}
		var used *decimal.Big
		if concretes {
			used = ecDbt.Concretes.Big
		} else {
			used = ecDbt.Abstracts.Big
		}
		usage = utils.SubstractBig(usage, used)
		ec.Merge(ecDbt)
	}
	return
}

// accountDebit will debit the usage out of an Account
func (aS *AccountS) accountDebit(acnt *utils.AccountProfile, usage *decimal.Big,
	cgrEv *utils.CGREvent, concretes bool) (ec *utils.EventCharges, err error) {

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
	if blncOpers, err = newBalanceOperators(acnt.ID, blcsWithWeight.Balances(), aS.fltrS, aS.connMgr,
		aS.cfg.AccountSCfg().AttributeSConns, aS.cfg.AccountSCfg().RateSConns); err != nil {
		return
	}

	for _, blncOper := range blncOpers {
		debFunc := blncOper.debitAbstracts
		if concretes {
			debFunc = blncOper.debitConcretes
		}
		if usage.Cmp(decimal.New(0, 0)) == 0 {
			return // no more debit
		}
		var ecDbt *utils.EventCharges
		if ecDbt, err = debFunc(new(decimal.Big).Copy(usage), cgrEv); err != nil {
			if err == utils.ErrFilterNotPassingNoCaps ||
				err == utils.ErrNotImplemented {
				err = nil
				continue
			}
			return
		}
		if ecDbt == nil {
			continue // no debit performed
		}
		if ec == nil { // first debit
			ec = utils.NewEventCharges()
		}
		var used *decimal.Big
		if concretes {
			used = ecDbt.Concretes.Big
		} else {
			used = ecDbt.Abstracts.Big
		}
		usage = utils.SubstractBig(usage, used)
		ec.Merge(ecDbt)
	}
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

// V1MaxAbstracts returns the maximum abstract units for the event, based on matching Accounts
func (aS *AccountS) V1MaxAbstracts(args *utils.ArgsAccountsForEvent, eEc *utils.ExtEventCharges) (err error) {
	var acnts utils.AccountProfilesWithWeight
	if acnts, err = aS.matchingAccountsForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer unlockAccountProfiles(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(acnts, args.CGREvent, false, false); err != nil {
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
	defer unlockAccountProfiles(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(acnts, args.CGREvent, false, true); err != nil {
		return
	}

	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}

	*eEc = *rcvEec
	return
}

// V1MaxConcretes returns the maximum concrete units for the event, based on matching Accounts
func (aS *AccountS) V1MaxConcretes(args *utils.ArgsAccountsForEvent, eEc *utils.ExtEventCharges) (err error) {
	var acnts utils.AccountProfilesWithWeight
	if acnts, err = aS.matchingAccountsForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer unlockAccountProfiles(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(acnts, args.CGREvent, true, false); err != nil {
		return
	}
	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}
	*eEc = *rcvEec
	return
}

// V1DebitConcretes performs debit of concrete units for the provided event
func (aS *AccountS) V1DebitConcretes(args *utils.ArgsAccountsForEvent, eEc *utils.ExtEventCharges) (err error) {
	var acnts utils.AccountProfilesWithWeight
	if acnts, err = aS.matchingAccountsForEvent(args.CGREvent.Tenant,
		args.CGREvent, args.AccountIDs, true); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	defer unlockAccountProfiles(acnts)

	var procEC *utils.EventCharges
	if procEC, err = aS.accountsDebit(acnts, args.CGREvent, true, true); err != nil {
		return
	}

	var rcvEec *utils.ExtEventCharges
	if rcvEec, err = procEC.AsExtEventCharges(); err != nil {
		return
	}

	*eEc = *rcvEec
	return
}

// V1ActionSetBalance performs an update for a specific balance in account
func (aS *AccountS) V1ActionSetBalance(args *utils.ArgsActSetBalance, rply *string) (err error) {
	if args.AccountID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AccountID)
	}
	if len(args.Diktats) == 0 {
		return utils.NewErrMandatoryIeMissing(utils.Diktats)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = aS.cfg.GeneralCfg().DefaultTenant
	}
	if _, err = guardian.Guardian.Guard(func() (interface{}, error) {
		return nil, actSetAccount(aS.dm, tnt, args.AccountID, args.Diktats, args.Reset)
	}, aS.cfg.GeneralCfg().LockingTimeout,
		utils.ConcatenatedKey(utils.CacheAccountProfiles, tnt, args.AccountID)); err != nil {
		return
	}

	*rply = utils.OK
	return
}

// V1RemoveBalance removes a balance for a specific account
func (aS *AccountS) V1ActionRemoveBalance(args *utils.ArgsActRemoveBalances, rply *string) (err error) {
	if args.AccountID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AccountID)
	}
	if len(args.BalanceIDs) == 0 {
		return utils.NewErrMandatoryIeMissing(utils.BalanceIDs)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = aS.cfg.GeneralCfg().DefaultTenant
	}
	if _, err = guardian.Guardian.Guard(func() (interface{}, error) {
		qAcnt, err := aS.dm.GetAccountProfile(tnt, args.AccountID)
		if err != nil {
			return nil, err
		}
		for _, balID := range args.BalanceIDs {
			delete(qAcnt.Balances, balID)
		}
		return nil, aS.dm.SetAccountProfile(qAcnt, false)
	}, aS.cfg.GeneralCfg().LockingTimeout,
		utils.ConcatenatedKey(utils.CacheAccountProfiles, tnt, args.AccountID)); err != nil {
		return
	}
	*rply = utils.OK
	return
}
