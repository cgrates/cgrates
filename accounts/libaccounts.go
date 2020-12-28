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
)

// newAccountBalances constructs accountBalances struct
func newAccountBalances(acnt *utils.AccountProfile,
	fltrS *engine.FilterS, ralsConns []string) (acntBlncs *accountBalances, err error) {
	blncs := utils.Balances(acnt.Balances) // will be changed with using map so OK to reference for now
	blncs.Sort()
	acntBlncs = &accountBalances{blnCfgs: blncs}
	// populate typIdx
	for i, blnCfg := range blncs {
		acntBlncs.typIdx[blnCfg.Type] = append(acntBlncs.typIdx[blnCfg.Type], i)
	}
	// populate mntryBlncs
	acntBlncs.mntryBlncs = make([]balanceProcessor, len(acntBlncs.typIdx[utils.MONETARY]))
	for i, blncIdx := range acntBlncs.typIdx[utils.MONETARY] {
		acntBlncs.mntryBlncs[i] = newMonetaryBalanceProcessor(acntBlncs.blnCfgs[blncIdx])
		acntBlncs.procs[acntBlncs.blnCfgs[blncIdx].ID] = acntBlncs.mntryBlncs[i]
	}
	// populate procs
	for _, blnCfg := range acntBlncs.blnCfgs {
		if blnCfg.Type == utils.MONETARY { // already computed above
			continue
		}
		if acntBlncs.procs[blnCfg.ID], err = newBalanceProcessor(blnCfg,
			acntBlncs.mntryBlncs); err != nil {
			return
		}
	}
	return
}

// accountBalances implements processing of the events centralized
type accountBalances struct {
	blnCfgs    []*utils.Balance            // ordered list of balance configurations
	typIdx     map[string][]int            // index based on type
	mntryBlncs []balanceProcessor          // separate references so we can pass them to the newBalanceProcessor
	procs      map[string]balanceProcessor // map[blncID]balanceProcessor

	fltrS     *engine.FilterS
	ralsConns []string
}

// newBalanceProcessor instantiates balanceProcessor interface
// mntBlncs are needed for combined debits
func newBalanceProcessor(blncCfg *utils.Balance,
	mntryBlncs []balanceProcessor) (bP balanceProcessor, err error) {
	switch blncCfg.Type {
	default:
		return nil, fmt.Errorf("unsupported balance type: <%s>", blncCfg.Type)
	case utils.MONETARY:
		return newMonetaryBalanceProcessor(blncCfg), nil
	}
}

// balanceProcessor is the implementation of a balance type
type balanceProcessor interface {
	process(cgrEv *utils.CGREventWithOpts,
		startTime time.Time, usage time.Duration) (ec *utils.EventCharges, err error)
}

// newMonetaryBalanceProcessor constructs a monetaryBalanceProcessor
func newMonetaryBalanceProcessor(blnCfg *utils.Balance) balanceProcessor {
	return &monetaryBalanceProcessor{blnCfg}
}

// monetaryBalanceProcessor is the processor for *monetary balance type
type monetaryBalanceProcessor struct {
	blnCfg *utils.Balance
}

// process implements the balanceProcessor interface
func (mb *monetaryBalanceProcessor) process(cgrEv *utils.CGREventWithOpts,
	startTime time.Time, usage time.Duration) (ec *utils.EventCharges, err error) {
	return
}
