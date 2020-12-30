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

// newAccountBalances constructs accountBalances
func newAccountBalances(acnt *utils.AccountProfile,
	fltrS *engine.FilterS, ralsConns []string) (acntBlncs *accountBalances, err error) {
	blncs := utils.Balances(acnt.Balances) // will be changed with using map so OK to reference for now
	blncs.Sort()
	acntBlncs = &accountBalances{blnCfgs: blncs}
	// populate typIdx
	for i, blnCfg := range blncs {
		acntBlncs.typIdx[blnCfg.Type] = append(acntBlncs.typIdx[blnCfg.Type], i)
	}
	// populate cncrtBlncs
	acntBlncs.cncrtBlncs = make([]balanceOperator, len(acntBlncs.typIdx[utils.MetaConcrete]))
	for i, blncIdx := range acntBlncs.typIdx[utils.MetaConcrete] {
		acntBlncs.cncrtBlncs[i] = newConcreteBalanceOperator(acntBlncs.blnCfgs[blncIdx], fltrS, ralsConns)
		acntBlncs.procs[acntBlncs.blnCfgs[blncIdx].ID] = acntBlncs.cncrtBlncs[i]
	}
	// populate procs
	for _, blnCfg := range acntBlncs.blnCfgs {
		if blnCfg.Type == utils.MetaConcrete { // already computed above
			continue
		}
		if acntBlncs.procs[blnCfg.ID], err = newBalanceOperator(blnCfg,
			acntBlncs.cncrtBlncs, fltrS, ralsConns); err != nil {
			return
		}
	}
	return
}

// accountBalances implements processing of the events centralized
type accountBalances struct {
	blnCfgs    []*utils.Balance           // ordered list of balance configurations
	typIdx     map[string][]int           // index based on type
	cncrtBlncs []balanceOperator          // concrete balances so we can pass them to the newBalanceOperator
	procs      map[string]balanceOperator // map[blncID]balanceOperator

	fltrS     *engine.FilterS
	ralsConns []string
}

// newBalanceOperator instantiates balanceOperator interface
// cncrtBlncs are needed for abstract balance debits
func newBalanceOperator(blncCfg *utils.Balance, cncrtBlncs []balanceOperator,
	fltrS *engine.FilterS, ralsConns []string) (bP balanceOperator, err error) {
	switch blncCfg.Type {
	default:
		return nil, fmt.Errorf("unsupported balance type: <%s>", blncCfg.Type)
	case utils.MetaConcrete:
		return newConcreteBalanceOperator(blncCfg, fltrS, ralsConns), nil
	case utils.MetaAbstract:
		return newAbstractBalanceOperator(blncCfg, cncrtBlncs, fltrS, ralsConns), nil
	}
}

// balanceOperator is the implementation of a balance type
type balanceOperator interface {
	debit(cgrEv *utils.CGREventWithOpts,
		startTime time.Time, usage *decimal.Big) (ec *utils.EventCharges, err error)
}

// usagewithFactor returns the usage considering also factor for the debit
//	includes event filtering to avoid code duplication
func usageWithFactor(usage *decimal.Big, blnCfg *utils.Balance, fltrS *engine.FilterS,
	cgrEv *utils.CGREventWithOpts) (mtchUF *utils.UsageFactor, err error) {
	fctr := decimal.New(1, 0)
	if len(blnCfg.FilterIDs) == 0 &&
		len(blnCfg.UsageFactors) == 0 {
		return
	}
	evNm := utils.MapStorage{
		utils.MetaOpts: cgrEv.Opts,
		utils.MetaReq:  cgrEv.Event,
	}
	// match the general balance filters
	var pass bool
	if pass, err = fltrS.Pass(cgrEv.CGREvent.Tenant, blnCfg.FilterIDs, evNm); err != nil {
		return nil, err
	} else if !pass {
		return nil, utils.ErrFilterNotPassingNoCaps
	}
	// find out the factor
	for _, uF := range blnCfg.UsageFactors {
		if pass, err = fltrS.Pass(cgrEv.CGREvent.Tenant, uF.FilterIDs, evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		fctr = uF.DecimalFactor()
		mtchUF = uF
		break
	}
	if mtchUF == nil {
		return
	}
	if fctr.Cmp(decimal.New(1, 0)) == 0 {
		*usage = *utils.MultiplyBig(usage, fctr)
	}
	return
}
