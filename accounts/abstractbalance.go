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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newAbstractBalance constructs an abstractBalanceOperator
func newAbstractBalanceOperator(blnCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) balanceOperator {
	return &abstractBalance{blnCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns}
}

// abstractBalance is the operator for *abstract balance type
type abstractBalance struct {
	blnCfg     *utils.Balance
	cncrtBlncs []*concreteBalance // paying balances
	fltrS      *engine.FilterS
	connMgr    *engine.ConnManager
	attrSConns,
	rateSConns []string
}

// costIncrement finds out the cost increment for the event
func (aB *abstractBalance) costIncrement(tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	for _, cIcrm := range aB.blnCfg.CostIncrements {
		var pass bool
		if pass, err = aB.fltrS.Pass(tnt, cIcrm.FilterIDs, ev); err != nil {
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

// unitFactor returns the unitFactor for the event
func (aB *abstractBalance) unitFactor(tnt string, ev utils.DataProvider) (uF *utils.UnitFactor, err error) {
	for _, uF = range aB.blnCfg.UnitFactors {
		var pass bool
		if pass, err = aB.fltrS.Pass(tnt, uF.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		return
	}
	return
}

// balanceLimit returns the balance's limit
func (aB *abstractBalance) balanceLimit() (bL *utils.Decimal) {
	if _, isUnlimited := aB.blnCfg.Opts[utils.MetaBalanceUnlimited]; isUnlimited {
		return
	}
	if lmtIface, has := aB.blnCfg.Opts[utils.MetaBalanceLimit]; has {
		bL = lmtIface.(*utils.Decimal)
	}
	// nothing matched, return default
	bL = utils.NewDecimal(0, 0)
	return
}

// rateSCostForEvent will process the event with RateS in order to get the cost
func (aB *abstractBalance) rateSCostForEvent(cgrEv *utils.CGREvent) (rplyCost *engine.RateProfileCost, err error) {
	if len(aB.rateSConns) == 0 {
		return nil, utils.NewErrNotConnected(utils.RateS)
	}
	err = aB.connMgr.Call(aB.rateSConns, nil, utils.RateSv1CostForEvent,
		&utils.ArgsCostForEvent{CGREvent: cgrEv}, &rplyCost)
	return
}

// debitUsageFromConcrete attempts to debit the usage out of concrete balances
// returns utils.ErrInsufficientCredit if complete usage cannot be debitted
func (aB *abstractBalance) debitUsageFromConcrete(usage *utils.Decimal,
	costIcrm *utils.CostIncrement, cgrEv *utils.CGREvent) (err error) {
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil {
		var rplyCost *engine.RateProfileCost
		if rplyCost, err = aB.rateSCostForEvent(cgrEv); err != nil {
			return
		}
		costIcrm.FixedFee = utils.NewDecimalFromFloat64(rplyCost.Cost)
	}
	var tCost *decimal.Big
	if costIcrm.FixedFee != nil {
		tCost = costIcrm.FixedFee.Big
	}
	// RecurrentFee is configured, used it with increments
	if costIcrm.RecurrentFee.Big.Cmp(decimal.New(-1, 0)) != 0 {
		rcrntCost := utils.MultiplyBig(
			utils.DivideBig(usage.Big, costIcrm.Increment.Big),
			costIcrm.RecurrentFee.Big)
		if tCost == nil {
			tCost = rcrntCost
		} else {
			tCost = utils.SumBig(tCost, rcrntCost)
		}
	}
	clnedUnts := cloneUnitsFromConcretes(aB.cncrtBlncs)
	for _, cB := range aB.cncrtBlncs {
		ev := utils.MapStorage{
			utils.MetaOpts: cgrEv.Opts,
			utils.MetaReq:  cgrEv.Event,
		}
		var dbted *utils.Decimal
		if dbted, _, err = cB.debitUnits(&utils.Decimal{tCost}, cgrEv.Tenant, ev); err != nil {
			restoreUnitsFromClones(aB.cncrtBlncs, clnedUnts)
			return
		}
		tCost = utils.SubstractBig(tCost, dbted.Big)
		if tCost.Cmp(decimal.New(0, 0)) <= 0 {
			return // have debited all, total is smaller or equal to 0
		}
	}
	// we could not debit all, put back what we have debited
	restoreUnitsFromClones(aB.cncrtBlncs, clnedUnts)
	return utils.ErrInsufficientCredit
}

// debitUsage implements the balanceOperator interface
func (aB *abstractBalance) debitUsage(usage *utils.Decimal,
	cgrEv *utils.CGREvent) (dbted *utils.Decimal, ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = aB.fltrS.Pass(cgrEv.Tenant, aB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, nil, utils.ErrFilterNotPassingNoCaps
	}

	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = aB.costIncrement(cgrEv.Tenant, evNm); err != nil {
		return
	}
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil &&
		len(aB.blnCfg.AttributeIDs) != 0 { // cost unknown, apply AttributeS to query from RateS
		var rplyAttrS *engine.AttrSProcessEventReply
		if rplyAttrS, err = processAttributeS(aB.connMgr, cgrEv, aB.attrSConns,
			aB.blnCfg.AttributeIDs); err != nil {
			return
		}
		if len(rplyAttrS.AlteredFields) != 0 { // event was altered
			cgrEv = rplyAttrS.CGREvent
		}
	}

	// balanceLimit
	var hasLmt bool
	blncLmt := aB.balanceLimit()
	if blncLmt != nil && blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		aB.blnCfg.Units.Big = utils.SubstractBig(aB.blnCfg.Units.Big, blncLmt.Big)
		hasLmt = true
	}

	// unitFactor
	var uF *utils.UnitFactor
	if uF, err = aB.unitFactor(cgrEv.Tenant, evNm); err != nil {
		return
	}
	var hasUF bool
	if uF != nil && uF.Factor.Cmp(decimal.New(1, 0)) != 0 {
		usage.Big = utils.MultiplyBig(usage.Big, uF.Factor.Big)
		hasUF = true
	}

	// balance smaller than usage, correct usage
	if aB.blnCfg.Units.Compare(usage) == -1 {
		// decrease the usage to match the maximum increments
		// will use special rounding to 0 since otherwise we go negative (ie: 0.05 as increment)
		usage.Big = roundedUsageWithIncrements(aB.blnCfg.Units.Big, costIcrm.Increment.Big)
	}

	// attempt to debit usage with cost
	// fix the maximum number of iterations
	origConcrtUnts := cloneUnitsFromConcretes(aB.cncrtBlncs) // so we can revert during usage checks
	paidConcrtUnts := origConcrtUnts
	var usagePaid, usageDenied *decimal.Big
	maxIter := 100
	for i := 0; i < maxIter; i++ {
		if i != 0 {
			restoreUnitsFromClones(aB.cncrtBlncs, origConcrtUnts)
		}
		if i == maxIter {
			return nil, nil, utils.ErrMaxIncrementsExceeded
		}
		qriedUsage := usage.Big // so we can detect loops
		if err = aB.debitUsageFromConcrete(usage, costIcrm, cgrEv); err != nil {
			if err != utils.ErrInsufficientCredit {
				return
			}
			err = nil
			// ErrInsufficientCredit
			usageDenied = new(decimal.Big).Copy(usage.Big)
			if usagePaid == nil { // going backwards
				usage.Big = utils.DivideBig( // divide by 2
					usage.Big, decimal.New(2, 0))
				usage.Big = roundedUsageWithIncrements(usage.Big, costIcrm.Increment.Big) // make sure usage is multiple of increments
				if usage.Big.Cmp(usageDenied) >= 0 ||
					usage.Big.Cmp(decimal.New(0, 0)) == 0 ||
					usage.Big.Cmp(qriedUsage) == 0 { // loop
					break
				}
				continue
			}

		} else {
			usagePaid = new(decimal.Big).Copy(usage.Big)
			paidConcrtUnts = cloneUnitsFromConcretes(aB.cncrtBlncs)
			if i == 0 { // no estimation done, covering full
				break
			}
		}
		// going upwards
		usage.Big = utils.SumBig(usagePaid,
			utils.DivideBig(usagePaid, decimal.New(2, 0)).RoundToInt())
		if usage.Big.Cmp(usageDenied) >= 0 {
			usage.Big = utils.SumBig(usagePaid, costIcrm.Increment.Big)
		}
		usage.Big = roundedUsageWithIncrements(usage.Big, costIcrm.Increment.Big)
		if usage.Big.Cmp(usagePaid) <= 0 ||
			usage.Big.Cmp(usageDenied) >= 0 ||
			usage.Big.Cmp(qriedUsage) == 0 { // loop
			break
		}
	}
	// Nothing paid
	if usagePaid == nil {
		// since we are erroring, we restore the concerete balances
		usagePaid = decimal.New(0, 0)
	}

	restoreUnitsFromClones(aB.cncrtBlncs, paidConcrtUnts)
	if usagePaid.Cmp(decimal.New(0, 0)) != 0 {
		aB.blnCfg.Units.Big = utils.SubstractBig(aB.blnCfg.Units.Big, usagePaid)
	}
	if hasLmt { // put back the limit
		aB.blnCfg.Units.Big = utils.SumBig(aB.blnCfg.Units.Big, blncLmt.Big)
	}
	if hasUF {
		usage.Big = utils.DivideBig(usage.Big, uF.Factor.Big)
	}
	dbted = &utils.Decimal{usagePaid}
	return
}
