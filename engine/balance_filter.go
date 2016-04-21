package engine

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type BalanceFilter struct {
	Uuid           *string
	ID             *string
	Type           *string
	Value          *ValueFormula
	Directions     *utils.StringMap
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

func (bp *BalanceFilter) CreateBalance() *Balance {
	b := &Balance{
		Uuid:           bp.GetUuid(),
		ID:             bp.GetID(),
		Value:          bp.GetValue(),
		Directions:     bp.GetDirections(),
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
	result := &BalanceFilter{}
	if bf.Uuid != nil {
		result.Uuid = new(string)
		*result.Uuid = *bf.Uuid
	}
	if bf.ID != nil {
		result.ID = new(string)
		*result.ID = *bf.ID
	}
	if bf.Value != nil {
		result.Value = new(ValueFormula)
		*result.Value = *bf.Value
	}
	if bf.RatingSubject != nil {
		result.RatingSubject = new(string)
		*result.RatingSubject = *bf.RatingSubject
	}
	if bf.Type != nil {
		result.Type = new(string)
		*result.Type = *bf.Type
	}
	if bf.ExpirationDate != nil {
		result.ExpirationDate = new(time.Time)
		*result.ExpirationDate = *bf.ExpirationDate
	}
	if bf.Weight != nil {
		result.Weight = new(float64)
		*result.Weight = *bf.Weight
	}
	if bf.Disabled != nil {
		result.Disabled = new(bool)
		*result.Disabled = *bf.Disabled
	}
	if bf.Blocker != nil {
		result.Blocker = new(bool)
		*result.Blocker = *bf.Blocker
	}
	if bf.Factor != nil {
		result.Factor = new(ValueFactor)
		*result.Factor = *bf.Factor
	}
	if bf.Directions != nil {
		result.Directions = utils.StringMapPointer(bf.Directions.Clone())
	}
	if bf.DestinationIDs != nil {
		result.DestinationIDs = utils.StringMapPointer(bf.DestinationIDs.Clone())
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
	if !b.Directions.IsEmpty() {
		bf.Directions = &b.Directions
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
	formula, exists := valueFormulas[bp.Value.Method]
	if !exists {
		return 0.0
	}
	return formula(bp.Value.Params)
}

func (bp *BalanceFilter) SetValue(v float64) {
	if bp.Value == nil {
		bp.Value = new(ValueFormula)
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

func (bp *BalanceFilter) GetDirections() utils.StringMap {
	if bp == nil || bp.Directions == nil {
		return utils.StringMap{}
	}
	return *bp.Directions
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
	if bf.Directions != nil {
		b.Directions = *bf.Directions
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

//for computing a dynamic value for Value field
type ValueFormula struct {
	Method string
	Params map[string]interface{}
	Static float64
}

func ParseBalanceFilterValue(val string) (*ValueFormula, error) {
	u, err := strconv.ParseFloat(val, 64)
	if err == nil {
		return &ValueFormula{Static: u}, err
	}
	var vf ValueFormula
	err = json.Unmarshal([]byte(val), &vf)
	if err == nil {
		return &vf, err
	}
	return nil, errors.New("Invalid value: " + val)
}

type valueFormula func(map[string]interface{}) float64

const (
	PERIODIC = "*periodic"
)

var valueFormulas = map[string]valueFormula{
	PERIODIC: periodicFormula,
}

func periodicFormula(params map[string]interface{}) float64 {
	return 0.0
}
