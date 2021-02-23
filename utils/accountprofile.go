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

package utils

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

// AccountProfile represents one Account on a Tenant
type AccountProfile struct {
	Tenant             string
	ID                 string // Account identificator, unique within the tenant
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weights            DynamicWeights
	Opts               map[string]interface{}
	Balances           map[string]*Balance
	ThresholdIDs       []string
}

// BalancesAltered detects altering of the Balances by comparing the Balance values with the ones from backup
func (ap *AccountProfile) BalancesAltered(abb AccountBalancesBackup) (altred bool) {
	if len(ap.Balances) != len(abb) {
		return true
	}
	for blncID, blnc := range ap.Balances {
		if bkpVal, has := abb[blncID]; !has {
			return true
		} else if blnc.Units.Big.Cmp(bkpVal) != 0 {
			return true
		}
	}
	return
}

func (ap *AccountProfile) RestoreFromBackup(abb AccountBalancesBackup) {
	for blncID, val := range abb {
		ap.Balances[blncID].Units.Big = val
	}
}

// AccountBalanceBackups is used to create balance snapshots as backups
type AccountBalancesBackup map[string]*decimal.Big

// AccountBalancesBackup returns a backup of all balance values
func (ap *AccountProfile) AccountBalancesBackup() (abb AccountBalancesBackup) {
	if ap.Balances != nil {
		abb = make(AccountBalancesBackup)
		for blncID, blnc := range ap.Balances {
			abb[blncID] = new(decimal.Big).Copy(blnc.Units.Big)
		}
	}
	return
}

func NewDefaultBalance(id string, val *Decimal) *Balance {
	const torFltr = "*string:~*req.ToR:"
	return &Balance{
		ID:    id,
		Type:  MetaConcrete,
		Units: val,
		CostIncrements: []*CostIncrement{
			{
				FilterIDs:    []string{torFltr + MetaVoice},
				Increment:    NewDecimal(int64(time.Second), 0),
				RecurrentFee: NewDecimal(0, 0),
			},
			{
				FilterIDs:    []string{torFltr + MetaData},
				Increment:    NewDecimal(1024*1024, 0),
				RecurrentFee: NewDecimal(0, 0),
			},
			{
				FilterIDs:    []string{torFltr + MetaSMS},
				Increment:    NewDecimal(1, 0),
				RecurrentFee: NewDecimal(0, 0),
			},
		},
	}
}

// Balance represents one Balance inside an Account
type Balance struct {
	ID             string // Balance identificator, unique within an Account
	FilterIDs      []string
	Weights        DynamicWeights
	Type           string
	Units          *Decimal
	UnitFactors    []*UnitFactor
	Opts           map[string]interface{}
	CostIncrements []*CostIncrement
	AttributeIDs   []string
	RateProfileIDs []string
}

// CostIncrement enforces cost calculation to specific balance increments
type CostIncrement struct {
	FilterIDs    []string
	Increment    *Decimal
	FixedFee     *Decimal
	RecurrentFee *Decimal
}

// Clone returns a copy of the CostIncrement
func (cI *CostIncrement) Clone() (cIcln *CostIncrement) {
	cIcln = new(CostIncrement)
	if cI.FilterIDs != nil {
		cIcln.FilterIDs = make([]string, len(cI.FilterIDs))
		for i, fID := range cI.FilterIDs {
			cIcln.FilterIDs[i] = fID
		}
	}
	if cI.Increment != nil {
		cIcln.Increment = cI.Increment.Clone()
	}
	if cI.FixedFee != nil {
		cIcln.FixedFee = cI.FixedFee.Clone()
	}
	if cI.RecurrentFee != nil {
		cIcln.RecurrentFee = cI.RecurrentFee.Clone()
	}
	return
}

//Clone return a copy of the UnitFactor
func (uF *UnitFactor) Clone() (untFct *UnitFactor) {
	untFct = new(UnitFactor)
	if uF.FilterIDs != nil {
		untFct.FilterIDs = make([]string, len(uF.FilterIDs))
		for i, value := range uF.FilterIDs {
			untFct.FilterIDs[i] = value
		}
	}
	if uF.Factor != nil {
		untFct.Factor = uF.Factor.Clone()
	}
	return
}

// UnitFactor is a multiplicator for the usage received
type UnitFactor struct {
	FilterIDs []string
	Factor    *Decimal
}

func (aP *AccountProfile) TenantID() string {
	return ConcatenatedKey(aP.Tenant, aP.ID)
}

// Clone returns a clone of the Account
func (aP *AccountProfile) Clone() (acnt *AccountProfile) {
	acnt = &AccountProfile{
		Tenant:             aP.Tenant,
		ID:                 aP.ID,
		ActivationInterval: aP.ActivationInterval.Clone(),
		Weights:            aP.Weights.Clone(),
	}
	if aP.FilterIDs != nil {
		acnt.FilterIDs = make([]string, len(aP.FilterIDs))
		for i, value := range aP.FilterIDs {
			acnt.FilterIDs[i] = value
		}
	}
	if aP.Opts != nil {
		acnt.Opts = make(map[string]interface{})
		for key, value := range aP.Opts {
			acnt.Opts[key] = value
		}
	}
	if aP.Balances != nil {
		acnt.Balances = make(map[string]*Balance, len(aP.Balances))
		for i, value := range aP.Balances {
			acnt.Balances[i] = value.Clone()
		}
	}
	if aP.ThresholdIDs != nil {
		acnt.ThresholdIDs = make([]string, len(aP.ThresholdIDs))
		for i, value := range aP.ThresholdIDs {
			acnt.ThresholdIDs[i] = value
		}
	}
	return
}

//Clone returns a clone of the ActivationInterval
func (aI *ActivationInterval) Clone() *ActivationInterval {
	if aI == nil {
		return nil
	}
	return &ActivationInterval{
		ActivationTime: aI.ActivationTime,
		ExpiryTime:     aI.ExpiryTime,
	}
}

//Clone return a clone of the Balance
func (bL *Balance) Clone() (blnc *Balance) {
	blnc = &Balance{
		ID:      bL.ID,
		Weights: bL.Weights.Clone(),
		Type:    bL.Type,
	}
	if bL.FilterIDs != nil {
		blnc.FilterIDs = make([]string, len(bL.FilterIDs))
		for i, value := range bL.FilterIDs {
			blnc.FilterIDs[i] = value
		}
	}
	if bL.Units != nil {
		blnc.Units = bL.Units.Clone()
	}
	if bL.UnitFactors != nil {
		blnc.UnitFactors = make([]*UnitFactor, len(bL.UnitFactors))
		for i, value := range bL.UnitFactors {
			blnc.UnitFactors[i] = value.Clone()
		}
	}
	if bL.Opts != nil {
		blnc.Opts = make(map[string]interface{})
		for key, value := range bL.Opts {
			blnc.Opts[key] = value
		}
	}
	if bL.CostIncrements != nil {
		blnc.CostIncrements = make([]*CostIncrement, len(bL.CostIncrements))
		for i, value := range bL.CostIncrements {
			blnc.CostIncrements[i] = value.Clone()
		}
	}
	if bL.AttributeIDs != nil {
		blnc.AttributeIDs = make([]string, len(bL.AttributeIDs))
		for i, value := range bL.AttributeIDs {
			blnc.AttributeIDs[i] = value
		}
	}
	if bL.RateProfileIDs != nil {
		blnc.RateProfileIDs = make([]string, len(bL.RateProfileIDs))
		for i, value := range bL.RateProfileIDs {
			blnc.RateProfileIDs[i] = value
		}
	}
	return
}

// AccountProfileWithWeight attaches static weight to AccountProfile
type AccountProfileWithWeight struct {
	*AccountProfile
	Weight float64
	LockID string
}

// AccountProfilesWithWeight is a sortable list of AccountProfileWithWeight
type AccountProfilesWithWeight []*AccountProfileWithWeight

// Sort is part of sort interface, sort based on Weight
func (aps AccountProfilesWithWeight) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

// AccountProfiles returns the list of AccountProfiles
func (apWws AccountProfilesWithWeight) AccountProfiles() (aps []*AccountProfile) {
	if apWws != nil {
		aps = make([]*AccountProfile, len(apWws))
		for i, apWw := range apWws {
			aps[i] = apWw.AccountProfile
		}
	}
	return
}

// LockIDs returns the list of LockIDs
func (apWws AccountProfilesWithWeight) LockIDs() (lkIDs []string) {
	if apWws != nil {
		lkIDs = make([]string, len(apWws))
		for i, apWw := range apWws {
			lkIDs[i] = apWw.LockID
		}
	}
	return
}

func (apWws AccountProfilesWithWeight) TenantIDs() (tntIDs []string) {
	if apWws != nil {
		tntIDs = make([]string, len(apWws))
		for i, apWw := range apWws {
			tntIDs[i] = apWw.AccountProfile.TenantID()
		}
	}
	return
}

// BalanceWithWeight attaches static Weight to Balance
type BalanceWithWeight struct {
	*Balance
	Weight float64
}

// BalancesWithWeight is a sortable list of BalanceWithWeight
type BalancesWithWeight []*BalanceWithWeight

// Sort is part of sort interface, sort based on Weight
func (blcs BalancesWithWeight) Sort() {
	sort.Slice(blcs, func(i, j int) bool { return blcs[i].Weight > blcs[j].Weight })
}

// Balances returns the list of Balances
func (bWws BalancesWithWeight) Balances() (blncs []*Balance) {
	if bWws != nil {
		blncs = make([]*Balance, len(bWws))
		for i, bWw := range bWws {
			blncs[i] = bWw.Balance
		}
	}
	return
}

// APIAccountProfileWithOpts is used in API calls
type APIAccountProfileWithOpts struct {
	*APIAccountProfile
	Opts map[string]interface{}
}

// AccountProfileWithOpts is used in API calls
type AccountProfileWithOpts struct {
	*AccountProfile
	Opts map[string]interface{}
}

// ArgsAccountForEvent arguments used for process event
type ArgsAccountsForEvent struct {
	*CGREvent
	AccountIDs []string
}

type ReplyMaxUsage struct {
	AccountID string
	MaxUsage  time.Duration
	Cost      *EventCharges
}

// APIAccountProfile represents one APIAccount on a Tenant
type APIAccountProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weights            string
	Opts               map[string]interface{}
	Balances           map[string]*APIBalance
	ThresholdIDs       []string
}

// AsAccountProfile convert APIAccountProfile struct to AccountProfile struct
func (ext *APIAccountProfile) AsAccountProfile() (profile *AccountProfile, err error) {
	profile = &AccountProfile{
		Tenant:             ext.Tenant,
		ID:                 ext.ID,
		FilterIDs:          ext.FilterIDs,
		ActivationInterval: ext.ActivationInterval,
		Opts:               ext.Opts,
		ThresholdIDs:       ext.ThresholdIDs,
	}
	if ext.Weights != EmptyString {
		if profile.Weights, err = NewDynamicWeightsFromString(ext.Weights, ";", "&"); err != nil {
			return nil, err
		}
	}
	if len(ext.Balances) != 0 {
		profile.Balances = make(map[string]*Balance, len(ext.Balances))
		for i, bal := range ext.Balances {
			if profile.Balances[i], err = bal.AsBalance(); err != nil {
				return nil, err
			}
		}
	}
	return
}

// APIBalance represents one APIBalance inside an APIAccount
type APIBalance struct {
	ID             string // Balance identificator, unique within an Account
	FilterIDs      []string
	Weights        string
	Type           string
	Units          float64
	UnitFactors    []*APIUnitFactor
	Opts           map[string]interface{}
	CostIncrements []*APICostIncrement
	AttributeIDs   []string
	RateProfileIDs []string
}

// AsBalance convert APIBalance struct to Balance struct
func (ext *APIBalance) AsBalance() (balance *Balance, err error) {
	balance = &Balance{
		ID:             ext.ID,
		FilterIDs:      ext.FilterIDs,
		Type:           ext.Type,
		Units:          NewDecimalFromFloat64(ext.Units),
		Opts:           ext.Opts,
		AttributeIDs:   ext.AttributeIDs,
		RateProfileIDs: ext.RateProfileIDs,
	}
	if ext.Weights != EmptyString {
		if balance.Weights, err = NewDynamicWeightsFromString(ext.Weights, ";", "&"); err != nil {
			return nil, err
		}
	}
	if len(ext.UnitFactors) != 0 {
		balance.UnitFactors = make([]*UnitFactor, len(ext.UnitFactors))
		for i, uFct := range ext.UnitFactors {
			balance.UnitFactors[i] = uFct.AsUnitFactor()
		}
	}
	if len(ext.CostIncrements) != 0 {
		balance.CostIncrements = make([]*CostIncrement, len(ext.CostIncrements))
		for i, cIncr := range ext.CostIncrements {
			balance.CostIncrements[i] = cIncr.AsCostIncrement()
		}
	}
	return

}

// APICostIncrement represent one CostIncrement inside an APIBalance
type APICostIncrement struct {
	FilterIDs    []string
	Increment    *float64
	FixedFee     *float64
	RecurrentFee *float64
}

// AsCostIncrement convert APICostIncrement struct to CostIncrement struct
func (ext *APICostIncrement) AsCostIncrement() (cIncr *CostIncrement) {
	cIncr = &CostIncrement{
		FilterIDs: ext.FilterIDs,
	}
	if ext.Increment != nil {
		cIncr.Increment = NewDecimalFromFloat64(*ext.Increment)
	}
	if ext.FixedFee != nil {
		cIncr.FixedFee = NewDecimalFromFloat64(*ext.FixedFee)
	}
	if ext.RecurrentFee != nil {
		cIncr.RecurrentFee = NewDecimalFromFloat64(*ext.RecurrentFee)
	}
	return
}

// APIUnitFactor represent one UnitFactor inside an APIBalance
type APIUnitFactor struct {
	FilterIDs []string
	Factor    float64
}

// AsUnitFactor convert APIUnitFactor struct to UnitFactor struct
func (ext *APIUnitFactor) AsUnitFactor() *UnitFactor {
	return &UnitFactor{
		FilterIDs: ext.FilterIDs,
		Factor:    NewDecimalFromFloat64(ext.Factor),
	}
}

type ArgsUpdateBalance struct {
	Tenant    string
	AccountID string
	Params    []*ArgsBalParams
	Reset     bool
}

type ArgsBalParams struct {
	Path  string
	Value string
}

func (ap *AccountProfile) Set(path []string, value interface{}) (err error) {
	if len(path) == 0 {
		return ErrWrongPath
	}
	// if len(path) == 1 {
	// 	return
	// }
	switch path[0] {
	case "*balance":
		if len(path) < 3 {
			return ErrWrongPath
		}
		return ap.Balances[path[1]].Set(path[2:], value)
	default:
		return ErrWrongPath
	}
}

func (bal *Balance) Set(path []string, value interface{}) (err error) {
	if len(path) == 0 {
		return ErrWrongPath
	}
	switch path[0] {
	case "ID":
		bal.ID = IfaceAsString(value)
	case "FilterIDs":
		var fltrsIDs []string
		if fltrsIDs, err = IfaceAsSliceString(value); err != nil {
			err = nil
			fltrsIDs = NewStringSet(strings.Split(IfaceAsString(value), InfieldSep)).AsSlice()
		}
		bal.FilterIDs = fltrsIDs
	case "Weights":
		var wg DynamicWeights
		if wg, err = NewDynamicWeightsFromString(IfaceAsString(value), InfieldSep, ANDSep); err != nil {
			return
		}
		bal.Weights = wg
	case "Type":
		bal.Type = IfaceAsString(value)
	case "Units":
		switch vl := value.(type) {
		case *Decimal:
			bal.Units = vl
		default:
			z, ok := new(decimal.Big).SetString(IfaceAsString(value))
			if !ok {
				return fmt.Errorf("can't convert <%+v> to decimal", value)
			}
			bal.Units.Big = z
		}
	case "UnitFactors":
	case "Opts":
	case "CostIncrements":
	case "AttributeIDs":
		var attrIDs []string
		if attrIDs, err = IfaceAsSliceString(value); err != nil {
			err = nil
			attrIDs = NewStringSet(strings.Split(IfaceAsString(value), InfieldSep)).AsSlice()
		}
		bal.AttributeIDs = attrIDs
	case "RateProfileIDs":
		var rateIDs []string
		if rateIDs, err = IfaceAsSliceString(value); err != nil {
			err = nil
			rateIDs = NewStringSet(strings.Split(IfaceAsString(value), InfieldSep)).AsSlice()
		}
		bal.RateProfileIDs = rateIDs
	default:
		return ErrWrongPath
	}
	return nil
}
