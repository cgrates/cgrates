/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"math"
	"math/rand"
	"time"
)

const (
	STRATEGY_LOWEST_FIRST  = "*lowest_first"
	STRATEGY_HIGHEST_FIRST = "*highest_first"
	STRATEGY_RANDOM        = "*random"
)

type SharedGroup struct {
	Id          string
	Strategy    string
	Account     string
	RateSubject string
	Weight      float64
	Members     []string
}

func (sg *SharedGroup) GetMembersExceptUser(ubId string) []string {
	for i, m := range sg.Members {
		if m == ubId {
			a := make([]string, len(sg.Members))
			copy(a, sg.Members)
			a[i], a = a[len(a)-1], a[:len(a)-1]
			return a
		}
	}
	return sg.Members
}

func (sg *SharedGroup) PopBalanceByStrategy(balanceChain *BalanceChain) (bal *Balance) {
	bc := *balanceChain
	if len(bc) == 0 {
		return
	}
	index := 0
	switch sg.Strategy {
	case STRATEGY_RANDOM:
		rand.Seed(time.Now().Unix())
		index = rand.Intn(len(bc))
	case STRATEGY_LOWEST_FIRST:
		minVal := math.MaxFloat64
		for i, b := range bc {
			if b.Value < minVal {
				minVal = b.Value
				index = i
			}
		}
	case STRATEGY_HIGHEST_FIRST:
		maxVal := math.SmallestNonzeroFloat64
		for i, b := range bc {
			if b.Value > maxVal {
				maxVal = b.Value
				index = i
			}
		}
	}
	bal, bc[index], *balanceChain = bc[index], bc[len(bc)-1], bc[:len(bc)-1]
	return
}
