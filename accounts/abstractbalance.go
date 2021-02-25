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

// debitAbstracts implements the balanceOperator interface
func (aB *abstractBalance) debitAbstracts(usage *decimal.Big,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {

	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}

	// pass the general balance filters
	var pass bool
	if pass, err = aB.fltrS.Pass(cgrEv.Tenant, aB.blnCfg.FilterIDs, evNm); err != nil {
		return
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
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
		usage = utils.MultiplyBig(usage, uF.Factor.Big)
		hasUF = true
	}

	// costIncrement
	var costIcrm *utils.CostIncrement
	if costIcrm, err = costIncrement(aB.blnCfg.CostIncrements, aB.fltrS,
		cgrEv.Tenant, evNm); err != nil {
		return
	}

	// balance smaller than usage, correct usage if the balance has limit
	if aB.blnCfg.Units.Big.Cmp(usage) == -1 && blncLmt != nil {
		// decrease the usage to match the maximum increments
		// will use special rounding to 0 since otherwise we go negative (ie: 0.05 as increment)
		usage = roundUnitsWithIncrements(aB.blnCfg.Units.Big, costIcrm.Increment.Big)
	}
	if costIcrm.RecurrentFee.Cmp(decimal.New(0, 0)) == 0 &&
		(costIcrm.FixedFee == nil ||
			costIcrm.FixedFee.Cmp(decimal.New(0, 0)) == 0) {
		// cost 0, no need of concrete
		ec = &utils.EventCharges{Usage: &utils.Decimal{usage}}
	} else {
		// attempt to debit usage with cost
		if ec, err = maxDebitAbstractsFromConcretes(aB.cncrtBlncs, usage,
			aB.connMgr, cgrEv,
			aB.attrSConns, aB.blnCfg.AttributeIDs,
			aB.rateSConns, aB.blnCfg.RateProfileIDs,
			costIcrm); err != nil {
			return
		}
	}

	if ec.Usage.Cmp(decimal.New(0, 0)) != 0 {
		aB.blnCfg.Units.Big = utils.SubstractBig(aB.blnCfg.Units.Big, ec.Usage.Big)
	}
	if hasLmt { // put back the limit
		aB.blnCfg.Units.Big = utils.SumBig(aB.blnCfg.Units.Big, blncLmt.Big)
	}
	if hasUF {
		usage = utils.DivideBig(usage, uF.Factor.Big)
	}
	return
}

// debitConcretes implements the balanceOperator interface
func (aB *abstractBalance) debitConcretes(usage *decimal.Big,
	cgrEv *utils.CGREvent) (ec *utils.EventCharges, err error) {
	return
}
