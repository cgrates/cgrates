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

// newConcreteBalance constructs a concreteBalanceOperator
func newConcreteBalanceOperator(blnCfg *utils.Balance,
	fltrS *engine.FilterS, connMgr *engine.ConnManager,
	attrSConns, rateSConns []string) balanceOperator {
	return &concreteBalance{blnCfg, fltrS, connMgr, attrSConns, rateSConns}
}

// concreteBalance is the operator for *concrete balance type
type concreteBalance struct {
	blnCfg  *utils.Balance
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager
	attrSConns,
	rateSConns []string
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
	if uF, err = unitFactor(cB.blnCfg.UnitFactors, cB.fltrS, tnt, ev); err != nil {
		return
	}
	if uF != nil && uF.Factor.Cmp(decimal.New(1, 0)) != 0 {
		hasUF = true
		dUnts = &utils.Decimal{utils.MultiplyBig(dUnts.Big, uF.Factor.Big)}
	}

	// balanceLimit
	var hasLmt bool
	var blncLmt *utils.Decimal
	if blncLmt, err = balanceLimit(cB.blnCfg.Opts); err != nil {
		return
	}
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
func (cB *concreteBalance) debitAbstracts(usage *decimal.Big,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {
	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = cB.fltrS.Pass(cgrEv.Tenant, cB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}

	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = costIncrement(cB.blnCfg.CostIncrements,
		cB.fltrS, cgrEv.Tenant, evNm); err != nil {
		return
	}

	return maxDebitAbstractsFromConcretes(
		[]*concreteBalance{cB}, usage,
		cB.connMgr, cgrEv,
		cB.attrSConns, cB.blnCfg.AttributeIDs,
		cB.rateSConns, cB.blnCfg.RateProfileIDs,
		costIcrm)

}
