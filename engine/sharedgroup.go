/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
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
	MemberIds         utils.StringMap
	//members           []*Account // accounts caching
}

type SharingParameters struct {
	Strategy      string
	RatingSubject string
}

func (sg *SharedGroup) SortBalancesByStrategy(myBalance *Balance, bc Balances) Balances {
	sharingParameters := sg.AccountParameters[utils.ANY]
	if sp, hasParamsForAccount := sg.AccountParameters[myBalance.account.ID]; hasParamsForAccount {
		sharingParameters = sp
	}

	strategy := STRATEGY_MINE_RANDOM
	if sharingParameters != nil && sharingParameters.Strategy != "" {
		strategy = sharingParameters.Strategy
	}
	switch strategy {
	case STRATEGY_LOWEST, STRATEGY_MINE_LOWEST:
		sort.Sort(LowestBalancesSorter(bc))
	case STRATEGY_HIGHEST, STRATEGY_MINE_HIGHEST:
		sort.Sort(HighestBalancesSorter(bc))
	case STRATEGY_RANDOM, STRATEGY_MINE_RANDOM:
		rbc := RandomBalancesSorter(bc)
		(&rbc).Sort()
		bc = Balances(rbc)
	default: // use mine random for anything else
		strategy = STRATEGY_MINE_RANDOM
		rbc := RandomBalancesSorter(bc)
		(&rbc).Sort()
		bc = Balances(rbc)
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
func (sg *SharedGroup) GetBalances(destination, category, direction, balanceType string, ub *Account) (bc Balances) {
	//	if len(sg.members) == 0 {
	for ubId := range sg.MemberIds {
		var nUb *Account
		if ubId == ub.ID { // skip the initiating user
			nUb = ub
		} else {
			nUb, _ = dataStorage.GetAccount(ubId)
			if nUb == nil || nUb.Disabled {
				continue
			}
		}
		//sg.members = append(sg.members, nUb)
		sb := nUb.getBalancesForPrefix(destination, category, direction, balanceType, sg.Id)
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

type LowestBalancesSorter []*Balance

func (lbcs LowestBalancesSorter) Len() int {
	return len(lbcs)
}

func (lbcs LowestBalancesSorter) Swap(i, j int) {
	lbcs[i], lbcs[j] = lbcs[j], lbcs[i]
}

func (lbcs LowestBalancesSorter) Less(i, j int) bool {
	return lbcs[i].GetValue() < lbcs[j].GetValue()
}

type HighestBalancesSorter []*Balance

func (hbcs HighestBalancesSorter) Len() int {
	return len(hbcs)
}

func (hbcs HighestBalancesSorter) Swap(i, j int) {
	hbcs[i], hbcs[j] = hbcs[j], hbcs[i]
}

func (hbcs HighestBalancesSorter) Less(i, j int) bool {
	return hbcs[i].GetValue() > hbcs[j].GetValue()
}

type RandomBalancesSorter []*Balance

func (rbcs *RandomBalancesSorter) Sort() {
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
