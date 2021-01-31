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
	if costIcrm, err = costIncrement(aB.blnCfg.CostIncrements, aB.fltrS,
		cgrEv.Tenant, evNm); err != nil {
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
	var blncLmt *utils.Decimal
	if blncLmt, err = balanceLimit(aB.blnCfg.Opts); err != nil {
		return
	}
	if blncLmt != nil && blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		aB.blnCfg.Units.Big = utils.SubstractBig(aB.blnCfg.Units.Big, blncLmt.Big)
		hasLmt = true
	}

	// unitFactor
	var uF *utils.UnitFactor
	if uF, err = unitFactor(aB.blnCfg.UnitFactors, aB.fltrS, cgrEv.Tenant, evNm); err != nil {
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
		if err = debitUsageFromConcretes(aB.cncrtBlncs, usage, costIcrm, cgrEv,
			aB.connMgr, aB.rateSConns, aB.blnCfg.RateProfileIDs); err != nil {
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
