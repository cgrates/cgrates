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

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

// newConcreteBalance constructs a concreteBalanceOperator
func newConcreteBalanceOperator(blnCfg *utils.Balance,
	fltrS *engine.FilterS, ralsConns []string) balanceOperator {
	return &concreteBalance{blnCfg, fltrS, ralsConns}
}

// concreteBalance is the operator for *concrete balance type
type concreteBalance struct {
	blnCfg    *utils.Balance
	fltrS     *engine.FilterS
	ralsConns []string
}

// debit implements the balanceOperator interface
func (cb *concreteBalance) debit(cgrEv *utils.CGREventWithOpts,
	startTime time.Time, usage *decimal.Big) (ec *utils.EventCharges, err error) {
	//var uF *utils.UsageFactor
	if _, _, err = usageWithFactor(cb.blnCfg, cb.fltrS, cgrEv, usage); err != nil {
		return
	}
	return
}

// debitUnits is a direct debit of balance units
func (cb *concreteBalance) debitUnits(dUnts *decimal.Big, incrm *decimal.Big,
	cgrEv *utils.CGREventWithOpts) (dbted *decimal.Big, mtchedUF *utils.UnitFactor, err error) {
	// *balanceLimit
	blncLmt := decimal.New(0, 0)
	if lmt, has := cb.blnCfg.Opts[utils.MetaBalanceLimit].(*decimal.Big); has {
		blncLmt = lmt
	}
	blcVal := new(decimal.Big).SetFloat64(cb.blnCfg.Value)
	var hasLmt bool
	if blncLmt.Cmp(decimal.New(0, 0)) != 0 {
		blcVal = utils.SubstractBig(blcVal, blncLmt)
		hasLmt = true
	}
	// dynamic unit factor
	fctr := decimal.New(1, 0)
	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}
	for _, uF := range cb.blnCfg.UnitFactors {
		var pass bool
		if pass, err = cb.fltrS.Pass(cgrEv.CGREvent.Tenant, uF.FilterIDs, evNm); err != nil {
			return nil, nil, err
		} else if !pass {
			continue
		}
		fctr = uF.Factor
		mtchedUF = uF
		break
	}
	var hasUF bool
	if fctr.Cmp(decimal.New(1, 0)) != 0 {
		dUnts = utils.MultiplyBig(dUnts, fctr)
		incrm = utils.MultiplyBig(incrm, fctr)
		hasUF = true
	}
	if blcVal.Cmp(dUnts) == -1 { // balance smaller than debit
		maxIncrm := utils.DivideBig(blcVal, incrm).RoundToInt()
		dUnts = utils.MultiplyBig(incrm, maxIncrm)
	}
	rmain := utils.SubstractBig(blcVal, dUnts)
	if hasLmt {
		rmain = utils.AddBig(rmain, blncLmt)
	}
	if hasUF {
		dbted = utils.DivideBig(dUnts, fctr)
	} else {
		dbted = dUnts
	}
	rmainFlt64, ok := rmain.Float64()
	if !ok {
		return nil, nil, fmt.Errorf("failed representing decimal <%s> as float64", rmain)
	}
	cb.blnCfg.Value = rmainFlt64
	return
}
