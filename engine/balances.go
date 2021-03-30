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
	"fmt"
	"sort"
	"time"

	"github.com/cgrates/cgrates/utils"
)

// Can hold different units as seconds or monetary
type Balance struct {
	Uuid           string //system wide unique
	ID             string // account wide unique
	Value          float64
	ExpirationDate time.Time
	Weight         float64
	DestinationIDs utils.StringMap
	RatingSubject  string
	Categories     utils.StringMap
	SharedGroups   utils.StringMap
	TimingIDs      utils.StringMap
	Disabled       bool
	Factor         ValueFactor
	Blocker        bool
	precision      int
	dirty          bool
}

func (b *Balance) Equal(o *Balance) bool {
	if len(b.DestinationIDs) == 0 {
		b.DestinationIDs = utils.StringMap{utils.MetaAny: true}
	}
	if len(o.DestinationIDs) == 0 {
		o.DestinationIDs = utils.StringMap{utils.MetaAny: true}
	}
	return b.Uuid == o.Uuid &&
		b.ID == o.ID &&
		b.ExpirationDate.Equal(o.ExpirationDate) &&
		b.Weight == o.Weight &&
		b.DestinationIDs.Equal(o.DestinationIDs) &&
		b.RatingSubject == o.RatingSubject &&
		b.Categories.Equal(o.Categories) &&
		b.SharedGroups.Equal(o.SharedGroups) &&
		b.Disabled == o.Disabled &&
		b.Blocker == o.Blocker
}

func (b *Balance) MatchFilter(o *BalanceFilter, skipIds, skipExpiry bool) bool {
	if o == nil {
		return true
	}
	if !skipIds && o.Uuid != nil && *o.Uuid != "" {
		return b.Uuid == *o.Uuid
	}
	if !skipIds && o.ID != nil && *o.ID != "" {
		return b.ID == *o.ID
	}
	if !skipExpiry {
		if o.ExpirationDate != nil && !b.ExpirationDate.Equal(*o.ExpirationDate) {
			return false
		}
	}
	return (o.Weight == nil || b.Weight == *o.Weight) &&
		(o.Blocker == nil || b.Blocker == *o.Blocker) &&
		(o.Disabled == nil || b.Disabled == *o.Disabled) &&
		(o.DestinationIDs == nil || b.DestinationIDs.Includes(*o.DestinationIDs)) &&
		(o.Categories == nil || b.Categories.Includes(*o.Categories)) &&
		(o.TimingIDs == nil || b.TimingIDs.Includes(*o.TimingIDs)) &&
		(o.SharedGroups == nil || b.SharedGroups.Includes(*o.SharedGroups)) &&
		(o.RatingSubject == nil || b.RatingSubject == *o.RatingSubject)
}

func (b *Balance) HardMatchFilter(o *BalanceFilter, skipIds bool) bool {
	if o == nil {
		return true
	}
	if !skipIds && o.Uuid != nil && *o.Uuid != "" {
		return b.Uuid == *o.Uuid
	}
	if !skipIds && o.ID != nil && *o.ID != "" {
		return b.ID == *o.ID
	}
	return (o.ExpirationDate == nil || b.ExpirationDate.Equal(*o.ExpirationDate)) &&
		(o.Weight == nil || b.Weight == *o.Weight) &&
		(o.Blocker == nil || b.Blocker == *o.Blocker) &&
		(o.Disabled == nil || b.Disabled == *o.Disabled) &&
		(o.DestinationIDs == nil || b.DestinationIDs.Equal(*o.DestinationIDs)) &&
		(o.Categories == nil || b.Categories.Equal(*o.Categories)) &&
		(o.TimingIDs == nil || b.TimingIDs.Equal(*o.TimingIDs)) &&
		(o.SharedGroups == nil || b.SharedGroups.Equal(*o.SharedGroups)) &&
		(o.RatingSubject == nil || b.RatingSubject == *o.RatingSubject)
}

// the default balance has standard Id
func (b *Balance) IsDefault() bool {
	return b.ID == utils.MetaDefault
}

// IsExpiredAt check if ExpirationDate is before time t
func (b *Balance) IsExpiredAt(t time.Time) bool {
	return !b.ExpirationDate.IsZero() && b.ExpirationDate.Before(t)
}

func (b *Balance) MatchCategory(category string) bool {
	return len(b.Categories) == 0 || b.Categories[category]
}

func (b *Balance) HasDestination() bool {
	return len(b.DestinationIDs) > 0 && !b.DestinationIDs[utils.MetaAny]
}

func (b *Balance) MatchDestination(destinationID string) bool {
	return !b.HasDestination() || b.DestinationIDs[destinationID]
}

func (b *Balance) Clone() *Balance {
	if b == nil {
		return nil
	}
	n := &Balance{
		Uuid:           b.Uuid,
		ID:             b.ID,
		Value:          b.Value, // this value is in seconds
		ExpirationDate: b.ExpirationDate,
		Weight:         b.Weight,
		RatingSubject:  b.RatingSubject,
		Categories:     b.Categories,
		SharedGroups:   b.SharedGroups,
		TimingIDs:      b.TimingIDs,
		Blocker:        b.Blocker,
		Disabled:       b.Disabled,
		dirty:          b.dirty,
	}
	if b.DestinationIDs != nil {
		n.DestinationIDs = b.DestinationIDs.Clone()
	}
	return n
}

func (b *Balance) GetValue() float64 {
	return b.Value
}

func (b *Balance) SetDirty() {
	b.dirty = true
}

// AsBalanceSummary converts the balance towards compressed information to be displayed
func (b *Balance) AsBalanceSummary(typ string) *BalanceSummary {
	bd := &BalanceSummary{UUID: b.Uuid, ID: b.ID, Type: typ, Value: b.Value, Disabled: b.Disabled}
	if bd.ID == "" {
		bd.ID = b.Uuid
	}
	return bd
}

/*
Structure to store minute buckets according to weight, precision or price.
*/
type Balances []*Balance

func (bc Balances) Len() int {
	return len(bc)
}

func (bc Balances) Swap(i, j int) {
	bc[i], bc[j] = bc[j], bc[i]
}

// we need the better ones at the beginning
func (bc Balances) Less(j, i int) bool {
	return bc[i].precision < bc[j].precision ||
		(bc[i].precision == bc[j].precision && bc[i].Weight < bc[j].Weight)
}

func (bc Balances) Sort() {
	sort.Sort(bc)
}

func (bc Balances) Equal(o Balances) bool {
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

func (bc Balances) Clone() Balances {
	var newChain Balances
	for _, b := range bc {
		newChain = append(newChain, b.Clone())
	}
	return newChain
}

func (bc Balances) GetBalance(uuid string) *Balance {
	for _, balance := range bc {
		if balance.Uuid == uuid {
			return balance
		}
	}
	return nil
}

func (bc Balances) HasBalance(balance *Balance) bool {
	for _, b := range bc {
		if b.Equal(balance) {
			return true
		}
	}
	return false
}

type ValueFactor map[string]float64

func (f ValueFactor) GetValue(tor string) float64 {
	if value, ok := f[tor]; ok {
		return value
	}
	return 1.0
}

// BalanceSummary represents compressed information about a balance
type BalanceSummary struct {
	UUID     string  // Balance UUID
	ID       string  // Balance ID  if not defined
	Type     string  // *voice, *data, etc
	Initial  float64 // initial value before the debit operation
	Value    float64
	Disabled bool
}

// BalanceSummaries is a list of BalanceSummaries
type BalanceSummaries []*BalanceSummary

// BalanceSummaryWithUUD returns a BalanceSummary based on an UUID
func (bs BalanceSummaries) BalanceSummaryWithUUD(bsUUID string) (b *BalanceSummary) {
	for _, blc := range bs {
		if blc.UUID == bsUUID {
			b = blc
			break
		}
	}
	return
}

// FieldAsInterface func to help EventCost FieldAsInterface
func (bl *BalanceSummary) FieldAsInterface(fldPath []string) (val interface{}, err error) {
	if bl == nil || len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
	case utils.UUID:
		return bl.UUID, nil
	case utils.ID:
		return bl.ID, nil
	case utils.Type:
		return bl.Type, nil
	case utils.Value:
		return bl.Value, nil
	case utils.Disabled:
		return bl.Disabled, nil
	}
}
