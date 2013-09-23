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
	"sort"
	"time"
)

// Can hold different units as seconds or monetary
type Balance struct {
	Id               string
	Value            float64
	ExpirationDate   time.Time
	Weight           float64
	GroupIds         []string
	SpecialPriceType string
	SpecialPrice     float64 // absolute for minutes and percent for monetary (can be positive or negative)
	DestinationId    string
	RateSubject      string
	precision        int
}

func (b *Balance) Equal(o *Balance) bool {
	return b.ExpirationDate.Equal(o.ExpirationDate) &&
		b.Weight == o.Weight &&
		b.DestinationId == o.DestinationId &&
		b.RateSubject == o.RateSubject
}

func (b *Balance) IsExpired() bool {
	return !b.ExpirationDate.IsZero() && b.ExpirationDate.Before(time.Now())
}

func (b *Balance) Clone() *Balance {
	return &Balance{
		Id:             b.Id,
		Value:          b.Value,
		DestinationId:  b.DestinationId,
		ExpirationDate: b.ExpirationDate,
		Weight:         b.Weight,
		RateSubject:    b.RateSubject,
	}
}

// Returns the available number of seconds for a specified credit
func (b *Balance) GetSecondsForCredit(credit float64) (seconds float64) {
	seconds = b.Value
	if b.SpecialPrice > 0 {
		seconds = math.Min(credit/b.SpecialPrice, b.Value)
	}
	return
}

/*
Structure to store minute buckets according to weight, precision or price.
*/
type BalanceChain []*Balance

func (bc BalanceChain) Len() int {
	return len(bc)
}

func (bc BalanceChain) Swap(i, j int) {
	bc[i], bc[j] = bc[j], bc[i]
}

func (bc BalanceChain) Less(j, i int) bool {
	return bc[i].Weight < bc[j].Weight ||
		bc[i].precision < bc[j].precision ||
		bc[i].SpecialPrice > bc[j].SpecialPrice
}

func (bc BalanceChain) Sort() {
	sort.Sort(bc)
}

func (bc BalanceChain) GetTotalValue() (total float64) {
	for _, b := range bc {
		if !b.IsExpired() {
			total += b.Value
		}
	}
	return
}

func (bc BalanceChain) Debit(amount float64) float64 {
	bc.Sort()
	for i, b := range bc {
		if b.IsExpired() {
			continue
		}
		if b.Value >= amount || i == len(bc)-1 { // if last one go negative
			b.Value -= amount
			break
		}
		b.Value = 0
		amount -= b.Value
	}
	return bc.GetTotalValue()
}

func (bc BalanceChain) Equal(o BalanceChain) bool {
	if len(bc) != len(o) {
		return false
	}
	bc.Sort()
	o.Sort()
	for i := 0; i < len(bc); i++ {
		if !bc[i].Equal(o[i]) {
			return false
		}
	}
	return true
}

func (bc BalanceChain) Clone() BalanceChain {
	var newChain BalanceChain
	for _, b := range bc {
		newChain = append(newChain, b.Clone())
	}
	return newChain
}
