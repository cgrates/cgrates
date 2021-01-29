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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newConcreteBalance constructs a concreteBalanceOperator
func newConcreteBalanceOperator(blnCfg *utils.Balance, cncrtBlncs []*concreteBalance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) balanceOperator {
	return &concreteBalance{blnCfg, cncrtBlncs, fltrS, connMgr, attrSConns, rateSConns}
}

// concreteBalance is the operator for *concrete balance type
type concreteBalance struct {
	blnCfg     *utils.Balance
	cncrtBlncs []*concreteBalance // paying balances
	fltrS      *engine.FilterS
	connMgr    *engine.ConnManager
	attrSConns,
	rateSConns []string
}

// costIncrement finds out the cost increment for the event
func (cB *concreteBalance) costIncrement(tnt string, ev utils.DataProvider) (costIcrm *utils.CostIncrement, err error) {
	for _, cIcrm := range cB.blnCfg.CostIncrements {
		var pass bool
		if pass, err = cB.fltrS.Pass(tnt, cIcrm.FilterIDs, ev); err != nil {
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
func (cB *concreteBalance) unitFactor(tnt string, ev utils.DataProvider) (uF *utils.UnitFactor, err error) {
	for _, uF = range cB.blnCfg.UnitFactors {
		var pass bool
		if pass, err = cB.fltrS.Pass(tnt, uF.FilterIDs, ev); err != nil {
			return
		} else if !pass {
			continue
		}
		return
	}
	return
}

// balanceLimit returns the balance's limit
func (cB *concreteBalance) balanceLimit() (bL *utils.Decimal) {
	if _, isUnlimited := cB.blnCfg.Opts[utils.MetaBalanceUnlimited]; isUnlimited {
		return
	}
	if lmtIface, has := cB.blnCfg.Opts[utils.MetaBalanceLimit]; has {
		bL = lmtIface.(*utils.Decimal)
		return
	}
	// nothing matched, return default
	bL = utils.NewDecimal(0, 0)
	return
}

// debitUnits is a direct debit of balance units
func (cB *concreteBalance) debitUnits(dUnts *utils.Decimal, tnt string,
	ev utils.DataProvider) (dbted *utils.Decimal, uF *utils.UnitFactor, err error) {

	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(tnt, cB.blnCfg.FilterIDs, ev); err != nil {
		return
	} else if !pass {
		return nil, nil, utils.ErrFilterNotPassingNoCaps
	}

	// unitFactor
	var hasUF bool
	if uF, err = cB.unitFactor(tnt, ev); err != nil {
		return
	}
	if uF != nil && uF.Factor.Cmp(decimal.New(1, 0)) != 0 {
		hasUF = true
		dUnts = &utils.Decimal{utils.MultiplyBig(dUnts.Big, uF.Factor.Big)}
	}

	// balanceLimit
	var hasLmt bool
	blncLmt := cB.balanceLimit()
	if blncLmt != nil && blncLmt.Big.Cmp(decimal.New(0, 0)) != 0 {
		cB.blnCfg.Units.Big = utils.SubstractBig(cB.blnCfg.Units.Big, blncLmt.Big)
		hasLmt = true
	}

	if cB.blnCfg.Units.Compare(dUnts) <= 0 && blncLmt != nil { // balance smaller than debit and limited
		dbted = &utils.Decimal{cB.blnCfg.Units.Big}
		cB.blnCfg.Units.Big = blncLmt.Big
	} else {
		cB.blnCfg.Units.Big = utils.SubstractBig(cB.blnCfg.Units.Big, dUnts.Big)
		if hasLmt { // put back the limit
			cB.blnCfg.Units.Big = utils.SumBig(cB.blnCfg.Units.Big, blncLmt.Big)
		}
		dbted = dUnts
	}
	if hasUF {
		dbted.Big = utils.DivideBig(dbted.Big, uF.Factor.Big)
	}

	return
}

// debit implements the balanceOperator interface
func (cB *concreteBalance) debitUsage(usage *utils.Decimal,
	cgrEv *utils.CGREvent) (dbted *utils.Decimal, ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(cgrEv.Tenant, cB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, nil, utils.ErrFilterNotPassingNoCaps
	}

	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = cB.costIncrement(cgrEv.Tenant, evNm); err != nil {
		return
	}
	if costIcrm.RecurrentFee.Cmp(decimal.New(-1, 0)) == 0 &&
		costIcrm.FixedFee == nil &&
		len(cB.blnCfg.AttributeIDs) != 0 { // cost unknown, apply AttributeS to query from RateS
		var rplyAttrS *engine.AttrSProcessEventReply
		if rplyAttrS, err = processAttributeS(cB.connMgr, cgrEv, cB.attrSConns,
			cB.blnCfg.AttributeIDs); err != nil {
			return
		}
		if len(rplyAttrS.AlteredFields) != 0 { // event was altered
			cgrEv = rplyAttrS.CGREvent
		}
	}

	// balanceLimit
	var hasLmt bool
	blncLmt := cB.balanceLimit()
	if blncLmt != nil && blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		cB.blnCfg.Units.Big = utils.SubstractBig(cB.blnCfg.Units.Big, blncLmt.Big)
		hasLmt = true
	}

	// unitFactor
	var uF *utils.UnitFactor
	if uF, err = cB.unitFactor(cgrEv.Tenant, evNm); err != nil {
		return
	}
	var hasUF bool
	if uF != nil && uF.Factor.Cmp(decimal.New(1, 0)) != 0 {
		usage.Big = utils.MultiplyBig(usage.Big, uF.Factor.Big)
		hasUF = true
	}

	// balance smaller than usage, correct usage
	if cB.blnCfg.Units.Compare(usage) == -1 {
		// decrease the usage to match the maximum increments
		// will use special rounding to 0 since otherwise we go negative (ie: 0.05 as increment)
		usage.Big = roundedUsageWithIncrements(cB.blnCfg.Units.Big, costIcrm.Increment.Big)
	}

	fmt.Printf("hasLmt: %v, hasUF: %v\n", hasLmt, hasUF)

	return
}

// cloneUnitsFromConcretes returns cloned units from the concrete balances passed as parameters
func cloneUnitsFromConcretes(cBs []*concreteBalance) (clnedUnts []*utils.Decimal) {
	if cBs == nil {
		return
	}
	clnedUnts = make([]*utils.Decimal, len(cBs))
	for i := range cBs {
		clnedUnts[i] = cBs[i].blnCfg.Units.Clone()
	}
	return
}

// restoreUnitsFromClones will restore the units from the clones
func restoreUnitsFromClones(cBs []*concreteBalance, clnedUnts []*utils.Decimal) {
	for i, clnedUnt := range clnedUnts {
		cBs[i].blnCfg.Units.Big = clnedUnt.Big
	}
}
