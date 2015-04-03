/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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

package engine

import (
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/cgrates/cgrates/utils"
)

const (
	MINE_PREFIX           = "*mine_"
	STRATEGY_MINE_LOWEST  = "*mine_lowest"
	STRATEGY_MINE_HIGHEST = "*mine_highest"
	STRATEGY_MINE_RANDOM  = "*mine_random"
	STRATEGY_LOWEST       = "*lowest"
	STRATEGY_HIGHEST      = "*highest"
	STRATEGY_RANDOM       = "*random"
)

type SharedGroup struct {
	Id                string
	AccountParameters map[string]*SharingParameters
	MemberIds         []string
	//members           []*Account // accounts caching
}

type SharingParameters struct {
	Strategy      string
	RatingSubject string
}

func (sg *SharedGroup) SortBalancesByStrategy(myBalance *Balance, bc BalanceChain) BalanceChain {
	sharingParameters := sg.AccountParameters[utils.ANY]
	if sp, hasParamsForAccount := sg.AccountParameters[myBalance.account.Id]; hasParamsForAccount {
		sharingParameters = sp
	}

	strategy := STRATEGY_MINE_RANDOM
	if sharingParameters != nil && sharingParameters.Strategy != "" {
		strategy = sharingParameters.Strategy
	}
	switch strategy {
	case STRATEGY_LOWEST, STRATEGY_MINE_LOWEST:
		sort.Sort(LowestBalanceChainSorter(bc))
	case STRATEGY_HIGHEST, STRATEGY_MINE_HIGHEST:
		sort.Sort(HighestBalanceChainSorter(bc))
	case STRATEGY_RANDOM, STRATEGY_MINE_RANDOM:
		rbc := RandomBalanceChainSorter(bc)
		(&rbc).Sort()
		bc = BalanceChain(rbc)
	default: // use mine random for anything else
		strategy = STRATEGY_MINE_RANDOM
		rbc := RandomBalanceChainSorter(bc)
		(&rbc).Sort()
		bc = BalanceChain(rbc)
	}
	if strings.HasPrefix(strategy, MINE_PREFIX) {
		// find index of my balance
		index := 0
		for i, b := range bc {
			if b.Uuid == myBalance.Uuid {
				index = i
				break
			}
		}
		// move my balance first
		bc[0], bc[index] = bc[index], bc[0]
	}
	return bc
}

// Returns all shared group's balances collected from user accounts'
func (sg *SharedGroup) GetBalances(destination, category, balanceType string, ub *Account) (bc BalanceChain) {
	//	if len(sg.members) == 0 {
	for _, ubId := range sg.MemberIds {
		var nUb *Account
		if ubId == ub.Id { // skip the initiating user
			nUb = ub
		} else {
			nUb, _ = accountingStorage.GetAccount(ubId)
			if nUb == nil || nUb.Disabled {
				continue
			}
		}
		//sg.members = append(sg.members, nUb)
		sb := nUb.getBalancesForPrefix(destination, category, nUb.BalanceMap[balanceType], sg.Id)
		bc = append(bc, sb...)
	}
	/*	} else {
		for _, m := range sg.members {
			sb := m.getBalancesForPrefix(destination, m.BalanceMap[balanceType], sg.Id)
			bc = append(bc, sb...)
		}
	}*/
	return
}

type LowestBalanceChainSorter []*Balance

func (lbcs LowestBalanceChainSorter) Len() int {
	return len(lbcs)
}

func (lbcs LowestBalanceChainSorter) Swap(i, j int) {
	lbcs[i], lbcs[j] = lbcs[j], lbcs[i]
}

func (lbcs LowestBalanceChainSorter) Less(i, j int) bool {
	return lbcs[i].Value < lbcs[j].Value
}

type HighestBalanceChainSorter []*Balance

func (hbcs HighestBalanceChainSorter) Len() int {
	return len(hbcs)
}

func (hbcs HighestBalanceChainSorter) Swap(i, j int) {
	hbcs[i], hbcs[j] = hbcs[j], hbcs[i]
}

func (hbcs HighestBalanceChainSorter) Less(i, j int) bool {
	return hbcs[i].Value > hbcs[j].Value
}

type RandomBalanceChainSorter []*Balance

func (rbcs *RandomBalanceChainSorter) Sort() {
	src := *rbcs
	// randomize balance chain
	dest := make([]*Balance, len(src))
	rand.Seed(time.Now().UnixNano())
	perm := rand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	*rbcs = dest
}
