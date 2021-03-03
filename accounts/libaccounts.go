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
	"errors"
	"fmt"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newAccountBalances constructs accountBalances
func newBalanceOperators(blnCfgs []*utils.Balance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (blncOpers []balanceOperator, err error) {

	blncOpers = make([]balanceOperator, len(blnCfgs))
	var cncrtBlncs []*concreteBalance
	for i, blnCfg := range blnCfgs { // build the concrete balances
		if blnCfg.Type != utils.MetaConcrete {
			continue
		}
		blncOpers[i] = newConcreteBalanceOperator(blnCfg,
			fltrS, connMgr, attrSConns, rateSConns)
		cncrtBlncs = append(cncrtBlncs, blncOpers[i].(*concreteBalance))
	}

	for i, blnCfg := range blnCfgs { // build the abstract balances
		if blnCfg.Type == utils.MetaConcrete {
			continue
		}
		if blncOpers[i], err = newBalanceOperator(blnCfg, cncrtBlncs, fltrS, connMgr,
			attrSConns, rateSConns); err != nil {
			return
		}
	}

	return
}

// newBalanceOperator instantiates balanceOperator interface
// cncrtBlncs are needed for abstract balance debits
func newBalanceOperator(blncCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) (bP balanceOperator, err error) {
	switch blncCfg.Type {
	default:
		return nil, fmt.Errorf("unsupported balance type: <%s>", blncCfg.Type)
	case utils.MetaConcrete:
		return newConcreteBalanceOperator(blncCfg, fltrS, connMgr, attrSConns, rateSConns), nil
	case utils.MetaAbstract:
		return newAbstractBalanceOperator(blncCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns), nil
	}
}

// balanceOperator is the implementation of a balance type
type balanceOperator interface {
	debitAbstracts(usage *decimal.Big, cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error)
	debitConcretes(usage *decimal.Big, cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error)
}

// roundUnitsWithIncrements rounds the usage based on increments
func roundUnitsWithIncrements(usage, incrm *decimal.Big) (rndedUsage *decimal.Big) {
	usgMaxIncrm := decimal.WithContext(
		decimal.Context{RoundingMode: decimal.ToZero}).Quo(usage,
		incrm).RoundToInt()
	rndedUsage = utils.MultiplyBig(usgMaxIncrm, incrm)
	return
}

// processAttributeS will process the event with AttributeS
func processAttributeS(connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	attrSConns, attrIDs []string) (rplyEv *engine.AttrSProcessEventReply, err error) {
	if len(attrSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.AttributeS)
	}
	var procRuns *int
	if val, has := cgrEv.Opts[utils.OptsAttributesProcessRuns]; has {
		if v, err := utils.IfaceAsTInt64(val); err == nil {
			procRuns = utils.IntPointer(int(v))
		}
	}
	attrArgs := &engine.AttrArgsProcessEvent{
		Context: utils.StringPointer(utils.FirstNonEmpty(
			engine.MapEvent(cgrEv.Opts).GetStringIgnoreErrors(utils.OptsContext),
			utils.MetaAccounts)),
		CGREvent:     cgrEv,
		AttributeIDs: attrIDs,
		ProcessRuns:  procRuns,
	}
	var tmpReply engine.AttrSProcessEventReply
	if err = connMgr.Call(attrSConns, nil, utils.AttributeSv1ProcessEvent,
		attrArgs, &tmpReply); err != nil {
		return
	}
	return &tmpReply, nil
}

// rateSCostForEvent will process the event with RateS in order to get the cost
func rateSCostForEvent(connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	rateSConns, rpIDs []string) (rplyCost *engine.RateProfileCost, err error) {
	if len(rateSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RateS)
	}
	var tmpReply engine.RateProfileCost
	if err = connMgr.Call(rateSConns, nil, utils.RateSv1CostForEvent,
		&utils.ArgsCostForEvent{CGREvent: cgrEv, RateProfileIDs: rpIDs}, &tmpReply); err != nil {
		return
	}
	return &tmpReply, nil
}

// costIncrement computes the costIncrement for the event
func costIncrement(cfgCostIncrmts []*utils.CostIncrement,
	fltrS *engine.FilterS, tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	for _, cIcrm := range cfgCostIncrmts {
		var pass bool
		if pass, err = fltrS.Pass(tnt, cIcrm.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		costIcrm = cIcrm
		break
	}
	if costIcrm == nil {
		costIcrm = new(utils.CostIncrement)
	}
	if costIcrm.Increment == nil {
		costIcrm.Increment = utils.NewDecimal(1, 0)
	}
	if costIcrm.RecurrentFee == nil {
		costIcrm.RecurrentFee = utils.NewDecimal(-1, 0)
	}
	return
}

// unitFactor detects the unitFactor for the event
func unitFactor(cfgUnitFactors []*utils.UnitFactor,
	fltrS *engine.FilterS, tnt string, ev utils.DataProvider) (uF *utils.UnitFactor, err error) {
	for _, uFcfg := range cfgUnitFactors {
		var pass bool
		if pass, err = fltrS.Pass(tnt, uFcfg.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		uF = uFcfg
		return
	}
	return
}

// balanceLimit returns the balance limit based on configuration
func balanceLimit(optsCfg map[string]interface{}) (bL *utils.Decimal, err error) {
	if _, isUnlimited := optsCfg[utils.MetaBalanceUnlimited]; isUnlimited {
		return // unlimited is nil pointer
	}
	if lmtIface, has := optsCfg[utils.MetaBalanceLimit]; has {
		flt64Lmt, canCast := lmtIface.(float64)
		if !canCast {
			return nil, errors.New("unsupported *balanceLimit format")
		}
		return utils.NewDecimalFromFloat64(flt64Lmt), nil
	}
	// nothing matched, return default
	bL = utils.NewDecimal(0, 0)
	return
}

// debitAbstractsFromConcretes attempts to debit the usage out of concrete balances
// returns utils.ErrInsufficientCredit if complete usage cannot be debited
func debitAbstractsFromConcretes(cncrtBlncs []*concreteBalance, usage *decimal.Big,
	costIcrm *utils.CostIncrement, cgrEv *utils.CGREvent,
	connMgr *engine.ConnManager, rateSConns, rpIDs []string) (ec *utils.EventCharges, err error) {
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil {
		var rplyCost *engine.RateProfileCost
		if rplyCost, err = rateSCostForEvent(connMgr, cgrEv, rateSConns, rpIDs); err != nil {
			err = utils.NewErrRateS(err)
			return
		}
		costIcrm = costIcrm.Clone() // so we don't modify the original
		costIcrm.FixedFee = utils.NewDecimalFromFloat64(rplyCost.Cost)
	}
	var tCost *decimal.Big
	if costIcrm.FixedFee != nil {
		tCost = costIcrm.FixedFee.Big
	}
	// RecurrentFee is configured, used it with increments
	if costIcrm.RecurrentFee.Big.Cmp(decimal.New(-1, 0)) != 0 {
		rcrntCost := utils.MultiplyBig(
			utils.DivideBig(usage, costIcrm.Increment.Big),
			costIcrm.RecurrentFee.Big)
		if tCost == nil {
			tCost = rcrntCost
		} else {
			tCost = utils.SumBig(tCost, rcrntCost)
		}
	}
	clnedUnts := cloneUnitsFromConcretes(cncrtBlncs)
	for i, cB := range cncrtBlncs {
		var ecCncrt *utils.EventCharges
		if ecCncrt, err = cB.debitConcretes(tCost, cgrEv); err != nil {
			restoreUnitsFromClones(cncrtBlncs, clnedUnts)
			return nil, err
		}
		if i == 0 {
			ec = utils.NewEventCharges()
		}
		ec.Merge(ecCncrt)
		tCost = utils.SubstractBig(tCost, ecCncrt.Concretes.Big)
		if tCost.Cmp(decimal.New(0, 0)) <= 0 {
			return // have debited all, total is smaller or equal to 0
		}
	}
	// we could not debit all, put back what we have debited
	restoreUnitsFromClones(cncrtBlncs, clnedUnts)
	return nil, utils.ErrInsufficientCredit
}

// maxDebitAbstractsFromConcretes will debit the maximum possible usage out of concretes
func maxDebitAbstractsFromConcretes(cncrtBlncs []*concreteBalance, usage *decimal.Big,
	connMgr *engine.ConnManager, cgrEv *utils.CGREvent,
	attrSConns, attributeIDs, rateSConns, rpIDs []string,
	costIcrm *utils.CostIncrement) (ec *utils.EventCharges, err error) {
	ec = utils.NewEventCharges()
	// process AttributeS if needed
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil &&
		len(attributeIDs) != 0 { // cost unknown, apply AttributeS to query from RateS
		var rplyAttrS *engine.AttrSProcessEventReply
		if rplyAttrS, err = processAttributeS(connMgr, cgrEv, attrSConns,
			attributeIDs); err != nil {
			return
		}
		if len(rplyAttrS.AlteredFields) != 0 { // event was altered
			cgrEv = rplyAttrS.CGREvent
		}
	}

	// fix the maximum number of iterations
	origConcrtUnts := cloneUnitsFromConcretes(cncrtBlncs) // so we can revert on errors
	paidConcrtUnts := origConcrtUnts                      // so we can revert when higher usages are not possible
	var usagePaid, usageDenied *decimal.Big
	maxItr := config.CgrConfig().AccountSCfg().MaxIterations
	for i := 0; i <= maxItr; i++ {
		if i != 0 {
			restoreUnitsFromClones(cncrtBlncs, origConcrtUnts)
		}
		if i == maxItr {
			return nil, utils.ErrMaxIncrementsExceeded
		}
		qriedUsage := usage // so we can detect loops
		var ecDbt *utils.EventCharges
		if ecDbt, err = debitAbstractsFromConcretes(cncrtBlncs, usage, costIcrm, cgrEv,
			connMgr, rateSConns, rpIDs); err != nil {
			if err != utils.ErrInsufficientCredit {
				return
			}
			err = nil
			// ErrInsufficientCredit
			usageDenied = new(decimal.Big).Copy(usage)
			if usagePaid == nil { // going backwards
				usage = utils.DivideBig( // divide by 2
					usage, decimal.New(2, 0))
				usage = roundUnitsWithIncrements(usage, costIcrm.Increment.Big) // make sure usage is multiple of increments
				if usage.Cmp(usageDenied) >= 0 ||
					usage.Cmp(decimal.New(0, 0)) == 0 ||
					usage.Cmp(qriedUsage) == 0 { // loop
					break
				}
				continue
			}
		} else {
			usagePaid = new(decimal.Big).Copy(usage)
			paidConcrtUnts = cloneUnitsFromConcretes(cncrtBlncs)
			ec.Merge(ecDbt)
			if i == 0 { // no estimation done, covering full
				break
			}
		}
		// going upwards
		usage = utils.SumBig(usagePaid,
			utils.DivideBig(usagePaid, decimal.New(2, 0)).RoundToInt())
		if usage.Cmp(usageDenied) >= 0 {
			usage = utils.SumBig(usagePaid, costIcrm.Increment.Big)
		}
		usage = roundUnitsWithIncrements(usage, costIcrm.Increment.Big)
		if usage.Cmp(usagePaid) <= 0 ||
			usage.Cmp(usageDenied) >= 0 ||
			usage.Cmp(qriedUsage) == 0 { // loop
			break
		}
	}
	// Nothing paid
	if usagePaid == nil {
		// since we are erroring, we restore the concerete balances
		usagePaid = decimal.New(0, 0)
	}
	restoreUnitsFromClones(cncrtBlncs, paidConcrtUnts)
	return &utils.EventCharges{Abstracts: &utils.Decimal{usagePaid}}, nil
}

// restoreAccounts will restore the accounts in DataDB out of their backups if present
func restoreAccounts(dm *engine.DataManager,
	acnts []*utils.AccountProfileWithWeight, bkps []utils.AccountBalancesBackup) {
	for i, bkp := range bkps {
		if bkp == nil ||
			!acnts[i].AccountProfile.BalancesAltered(bkp) {
			continue
		}
		acnts[i].AccountProfile.RestoreFromBackup(bkp)
		if err := dm.SetAccountProfile(acnts[i].AccountProfile, false); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> error <%s> restoring account <%s>",
				utils.AccountS, err, acnts[i].AccountProfile.TenantID()))
		}
	}
}

// unlockAccountProfiles is used to unlock the accounts based on their lock identifiers
func unlockAccountProfiles(acnts utils.AccountProfilesWithWeight) {
	for _, lkID := range acnts.LockIDs() {
		guardian.Guardian.UnguardIDs(lkID)
	}
}
