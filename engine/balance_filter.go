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
	"math"
	"reflect"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type BalanceFilter struct {
	Uuid           *string
	ID             *string
	Type           *string
	Value          *utils.ValueFormula
	ExpirationDate *time.Time
	Weight         *float64
	DestinationIDs *utils.StringMap
	RatingSubject  *string
	Categories     *utils.StringMap
	SharedGroups   *utils.StringMap
	TimingIDs      *utils.StringMap
	Timings        []*RITiming
	Disabled       *bool
	Factor         *ValueFactor
	Blocker        *bool
}

// NewBalanceFilter creates a new BalanceFilter based on given filter
func NewBalanceFilter(filter map[string]interface{}, defaultTimezone string) (bf *BalanceFilter, err error) {
	bf = new(BalanceFilter)
	if id, has := filter[utils.ID]; has {
		bf.ID = utils.StringPointer(utils.IfaceAsString(id))
	}
	if uuid, has := filter[utils.UUID]; has {
		bf.Uuid = utils.StringPointer(utils.IfaceAsString(uuid))
	}
	// if ty, has := filter[utils.Type]; has {
	// 	bf.Type = utils.StringPointer(utils.IfaceAsString(ty))
	// }
	if val, has := filter[utils.Value]; has {
		var value float64
		if value, err = utils.IfaceAsFloat64(val); err != nil {
			return
		}
		bf.Value = &utils.ValueFormula{Static: math.Abs(value)}
	}
	if exp, has := filter[utils.ExpiryTime]; has {
		var expTime time.Time
		if expTime, err = utils.IfaceAsTime(exp, defaultTimezone); err != nil {
			return
		}
		bf.ExpirationDate = utils.TimePointer(expTime)
	}
	if weight, has := filter[utils.Weight]; has {
		var value float64
		if value, err = utils.IfaceAsFloat64(weight); err != nil {
			return
		}
		bf.Weight = utils.Float64Pointer(value)
	}
	if dst, has := filter[utils.DestinationIDs]; has {
		bf.DestinationIDs = utils.StringMapPointer(utils.ParseStringMap(utils.IfaceAsString(dst)))
	}
	if rs, has := filter[utils.RatingSubject]; has {
		bf.RatingSubject = utils.StringPointer(utils.IfaceAsString(rs))
	}
	if cat, has := filter[utils.Categories]; has {
		bf.Categories = utils.StringMapPointer(utils.ParseStringMap(utils.IfaceAsString(cat)))
	}
	if grps, has := filter[utils.SharedGroups]; has {
		bf.SharedGroups = utils.StringMapPointer(utils.ParseStringMap(utils.IfaceAsString(grps)))
	}
	if tim, has := filter[utils.TimingIDs]; has {
		bf.TimingIDs = utils.StringMapPointer(utils.ParseStringMap(utils.IfaceAsString(tim)))
	}
	if dis, has := filter[utils.Disabled]; has {
		var value bool
		if value, err = utils.IfaceAsBool(dis); err != nil {
			return
		}
		bf.Disabled = utils.BoolPointer(value)
	}
	if blk, has := filter[utils.Blocker]; has {
		var value bool
		if value, err = utils.IfaceAsBool(blk); err != nil {
			return
		}
		bf.Blocker = utils.BoolPointer(value)
	}
	return
}

func (bp *BalanceFilter) CreateBalance() *Balance {
	b := &Balance{
		Uuid:           bp.GetUuid(),
		ID:             bp.GetID(),
		Value:          bp.GetValue(),
		ExpirationDate: bp.GetExpirationDate(),
		Weight:         bp.GetWeight(),
		DestinationIDs: bp.GetDestinationIDs(),
		RatingSubject:  bp.GetRatingSubject(),
		Categories:     bp.GetCategories(),
		SharedGroups:   bp.GetSharedGroups(),
		Timings:        bp.Timings,
		TimingIDs:      bp.GetTimingIDs(),
		Disabled:       bp.GetDisabled(),
		Factor:         bp.GetFactor(),
		Blocker:        bp.GetBlocker(),
	}
	return b.Clone()
}

func (bf *BalanceFilter) Clone() *BalanceFilter {
	if bf == nil {
		return nil
	}
	result := &BalanceFilter{}
	if bf.Uuid != nil {
		result.Uuid = new(string)
		*result.Uuid = *bf.Uuid
	}
	if bf.ID != nil {
		result.ID = new(string)
		*result.ID = *bf.ID
	}
	if bf.Type != nil {
		result.Type = new(string)
		*result.Type = *bf.Type
	}
	if bf.Value != nil {
		result.Value = new(utils.ValueFormula)
		*result.Value = *bf.Value
	}
	if bf.ExpirationDate != nil {
		result.ExpirationDate = new(time.Time)
		*result.ExpirationDate = *bf.ExpirationDate
	}
	if bf.Weight != nil {
		result.Weight = new(float64)
		*result.Weight = *bf.Weight
	}
	if bf.DestinationIDs != nil {
		result.DestinationIDs = utils.StringMapPointer(bf.DestinationIDs.Clone())
	}
	if bf.RatingSubject != nil {
		result.RatingSubject = new(string)
		*result.RatingSubject = *bf.RatingSubject
	}
	if bf.Categories != nil {
		result.Categories = utils.StringMapPointer(bf.Categories.Clone())
	}
	if bf.SharedGroups != nil {
		result.SharedGroups = utils.StringMapPointer(bf.SharedGroups.Clone())
	}
	if bf.TimingIDs != nil {
		result.TimingIDs = utils.StringMapPointer(bf.TimingIDs.Clone())
	}
	if bf.Timings != nil {
		result.Timings = make([]*RITiming, len(bf.Timings))
		for i, rit := range bf.Timings {
			result.Timings[i] = rit.Clone()
		}
	}
	if bf.Disabled != nil {
		result.Disabled = new(bool)
		*result.Disabled = *bf.Disabled
	}
	if bf.Factor != nil {
		result.Factor = new(ValueFactor)
		*result.Factor = *bf.Factor
	}
	if bf.Blocker != nil {
		result.Blocker = new(bool)
		*result.Blocker = *bf.Blocker
	}
	return result
}

func (bf *BalanceFilter) LoadFromBalance(b *Balance) *BalanceFilter {
	if b.Uuid != "" {
		bf.Uuid = &b.Uuid
	}
	if b.ID != "" {
		bf.ID = &b.ID
	}
	if b.Value != 0 {
		bf.Value.Static = b.Value
	}
	if !b.ExpirationDate.IsZero() {
		bf.ExpirationDate = &b.ExpirationDate
	}
	if b.Weight != 0 {
		bf.Weight = &b.Weight
	}
	if !b.DestinationIDs.IsEmpty() {
		bf.DestinationIDs = &b.DestinationIDs
	}
	if b.RatingSubject != "" {
		bf.RatingSubject = &b.RatingSubject
	}
	if !b.Categories.IsEmpty() {
		bf.Categories = &b.Categories
	}
	if !b.SharedGroups.IsEmpty() {
		bf.SharedGroups = &b.SharedGroups
	}
	if !b.TimingIDs.IsEmpty() {
		bf.TimingIDs = &b.TimingIDs
	}
	if len(b.Timings) != 0 {
		bf.Timings = make([]*RITiming, len(b.Timings))
		for i, timing := range b.Timings {
			bf.Timings[i] = timing
		}
	}
	if len(b.Factor) != 0 {
		bf.Factor = &b.Factor
	}
	if b.Disabled {
		bf.Disabled = &b.Disabled
	}
	if b.Blocker {
		bf.Blocker = &b.Blocker
	}
	bf.Timings = b.Timings
	return bf
}

func (bp *BalanceFilter) Equal(o *BalanceFilter) bool {
	if bp.ID != nil && o.ID != nil {
		return *bp.ID == *o.ID
	}
	return reflect.DeepEqual(bp, o)
}

func (bp *BalanceFilter) GetType() string {
	if bp == nil || bp.Type == nil {
		return ""
	}
	return *bp.Type
}

func (bp *BalanceFilter) GetValue() float64 {
	if bp == nil || bp.Value == nil {
		return 0.0
	}
	if bp.Value.Method == "" {
		return bp.Value.Static
	}
	// calculate using formula
	formula, exists := utils.ValueFormulas[bp.Value.Method]
	if !exists {
		return 0.0
	}
	return formula(bp.Value.Params)
}

func (bp *BalanceFilter) SetValue(v float64) {
	if bp.Value == nil {
		bp.Value = new(utils.ValueFormula)
	}
	bp.Value.Static = v
}

func (bp *BalanceFilter) GetUuid() string {
	if bp == nil || bp.Uuid == nil {
		return ""
	}
	return *bp.Uuid
}

func (bp *BalanceFilter) GetID() string {
	if bp == nil || bp.ID == nil {
		return ""
	}
	return *bp.ID
}

func (bp *BalanceFilter) GetDestinationIDs() utils.StringMap {
	if bp == nil || bp.DestinationIDs == nil {
		return utils.StringMap{}
	}
	return *bp.DestinationIDs
}

func (bp *BalanceFilter) GetCategories() utils.StringMap {
	if bp == nil || bp.Categories == nil {
		return utils.StringMap{}
	}
	return *bp.Categories
}

func (bp *BalanceFilter) GetTimingIDs() utils.StringMap {
	if bp == nil || bp.TimingIDs == nil {
		return utils.StringMap{}
	}
	return *bp.TimingIDs
}

func (bp *BalanceFilter) GetSharedGroups() utils.StringMap {
	if bp == nil || bp.SharedGroups == nil {
		return utils.StringMap{}
	}
	return *bp.SharedGroups
}

func (bp *BalanceFilter) GetWeight() float64 {
	if bp == nil || bp.Weight == nil {
		return 0.0
	}
	return *bp.Weight
}

func (bp *BalanceFilter) GetRatingSubject() string {
	if bp == nil || bp.RatingSubject == nil {
		return ""
	}
	return *bp.RatingSubject
}

func (bp *BalanceFilter) GetDisabled() bool {
	if bp == nil || bp.Disabled == nil {
		return false
	}
	return *bp.Disabled
}

func (bp *BalanceFilter) GetBlocker() bool {
	if bp == nil || bp.Blocker == nil {
		return false
	}
	return *bp.Blocker
}

func (bp *BalanceFilter) GetExpirationDate() time.Time {
	if bp == nil || bp.ExpirationDate == nil {
		return time.Time{}
	}
	return *bp.ExpirationDate
}

func (bp *BalanceFilter) GetFactor() ValueFactor {
	if bp == nil || bp.Factor == nil {
		return ValueFactor{}
	}
	return *bp.Factor
}

func (bp *BalanceFilter) EmptyExpirationDate() bool {
	if bp.ExpirationDate == nil {
		return true
	}
	return (*bp.ExpirationDate).IsZero()
}

func (bf *BalanceFilter) ModifyBalance(b *Balance) {
	if b == nil {
		return
	}
	if bf.ID != nil {
		b.ID = *bf.ID
	}
	if bf.Value != nil {
		b.Value = bf.GetValue()
	}
	if bf.ExpirationDate != nil {
		b.ExpirationDate = *bf.ExpirationDate
	}
	if bf.RatingSubject != nil {
		b.RatingSubject = *bf.RatingSubject
	}
	if bf.Categories != nil {
		b.Categories = *bf.Categories
	}
	if bf.DestinationIDs != nil {
		b.DestinationIDs = *bf.DestinationIDs
	}
	if bf.SharedGroups != nil {
		b.SharedGroups = *bf.SharedGroups
	}
	if bf.TimingIDs != nil {
		b.TimingIDs = *bf.TimingIDs
	}
	if bf.Timings != nil && len(bf.Timings) != 0 {
		b.Timings = make([]*RITiming, len(bf.Timings))
		for i, timing := range bf.Timings {
			b.Timings[i] = timing
		}
	}
	if bf.Weight != nil {
		b.Weight = *bf.Weight
	}
	if bf.Blocker != nil {
		b.Blocker = *bf.Blocker
	}
	if bf.Disabled != nil {
		b.Disabled = *bf.Disabled
	}
	b.SetDirty() // Mark the balance as dirty since we have modified and it should be checked by action triggers
}
