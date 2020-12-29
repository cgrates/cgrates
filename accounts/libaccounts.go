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
	acntBlncs.cncrtBlncs = make([]balanceProcessor, len(acntBlncs.typIdx[utils.MetaConcrete]))
	for i, blncIdx := range acntBlncs.typIdx[utils.MetaConcrete] {
		acntBlncs.cncrtBlncs[i] = newConcreteBalanceProcessor(acntBlncs.blnCfgs[blncIdx])
		acntBlncs.procs[acntBlncs.blnCfgs[blncIdx].ID] = acntBlncs.cncrtBlncs[i]
	}
	// populate procs
	for _, blnCfg := range acntBlncs.blnCfgs {
		if blnCfg.Type == utils.MetaConcrete { // already computed above
			continue
		}
		if acntBlncs.procs[blnCfg.ID], err = newBalanceProcessor(blnCfg,
			acntBlncs.cncrtBlncs); err != nil {
			return
		}
	}
	return
}

// accountBalances implements processing of the events centralized
type accountBalances struct {
	blnCfgs    []*utils.Balance            // ordered list of balance configurations
	typIdx     map[string][]int            // index based on type
	cncrtBlncs []balanceProcessor          // concrete balances so we can pass them to the newBalanceProcessor
	procs      map[string]balanceProcessor // map[blncID]balanceProcessor

	fltrS     *engine.FilterS
	ralsConns []string
}

// newBalanceProcessor instantiates balanceProcessor interface
// cncrtBlncs are needed for abstract debits
func newBalanceProcessor(blncCfg *utils.Balance,
	cncrtBlncs []balanceProcessor) (bP balanceProcessor, err error) {
	switch blncCfg.Type {
	default:
		return nil, fmt.Errorf("unsupported balance type: <%s>", blncCfg.Type)
	case utils.MetaConcrete:
		return newConcreteBalanceProcessor(blncCfg), nil
	case utils.MetaAbstract:
		return newAbstractBalanceProcessor(blncCfg, cncrtBlncs), nil
	}
}

// balanceProcessor is the implementation of a balance type
type balanceProcessor interface {
	process(cgrEv *utils.CGREventWithOpts,
		startTime time.Time, usage time.Duration) (ec *utils.EventCharges, err error)
}

// newConcreteBalanceProcessor constructs a concreteBalanceProcessor
func newConcreteBalanceProcessor(blnCfg *utils.Balance) balanceProcessor {
	return &concreteBalanceProcessor{blnCfg}
}

// concreteBalanceProcessor is the processor for *concrete balance type
type concreteBalanceProcessor struct {
	blnCfg *utils.Balance
}

// process implements the balanceProcessor interface
func (cb *concreteBalanceProcessor) process(cgrEv *utils.CGREventWithOpts,
	startTime time.Time, usage time.Duration) (ec *utils.EventCharges, err error) {
	return
}

// newAbstractBalanceProcessor constructs an abstractBalanceProcessor
func newAbstractBalanceProcessor(blnCfg *utils.Balance, cncrtBlncs []balanceProcessor) balanceProcessor {
	return &abstractBalanceProcessor{blnCfg, cncrtBlncs}
}

// abstractBalanceProcessor is the processor for *abstract balance type
type abstractBalanceProcessor struct {
	blnCfg     *utils.Balance
	cncrtBlncs []balanceProcessor // paying balances
}

// process implements the balanceProcessor interface
func (ab *abstractBalanceProcessor) process(cgrEv *utils.CGREventWithOpts,
	startTime time.Time, usage time.Duration) (ec *utils.EventCharges, err error) {
	return
}
