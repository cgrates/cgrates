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
	"errors"
	"sort"
	"time"

	"github.com/ericlagergren/decimal"
)

// Account represents one Account on a Tenant
type Account struct {
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
func (ap *Account) BalancesAltered(abb AccountBalancesBackup) (altred bool) {
	if len(ap.Balances) != len(abb) {
		return true
	}
	for blncID, blnc := range ap.Balances {
		if bkpVal, has := abb[blncID]; !has ||
			blnc.Units.Big.Cmp(bkpVal) != 0 {
			return true
		}
	}
	return
}

func (ap *Account) RestoreFromBackup(abb AccountBalancesBackup) {
	for blncID, val := range abb {
		ap.Balances[blncID].Units.Big = val
	}
}

// AccountBalancesBackup returns a backup of all balance values
func (ap *Account) AccountBalancesBackup() (abb AccountBalancesBackup) {
	if ap.Balances != nil {
		abb = make(AccountBalancesBackup)
		for blncID, blnc := range ap.Balances {
			abb[blncID] = new(decimal.Big).Copy(blnc.Units.Big)
		}
	}
	return
}

// AccountBalanceBackups is used to create balance snapshots as backups
type AccountBalancesBackup map[string]*decimal.Big

// NewDefaultBalance returns a balance with default costIncrements
func NewDefaultBalance(id string) *Balance {
	const torFltr = "*string:~*req.ToR:"
	return &Balance{
		ID:    id,
		Type:  MetaConcrete,
		Units: NewDecimal(0, 0),
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

type ExtAccount struct {
	Tenant             string
	ID                 string // Account identificator, unique within the tenant
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weights            DynamicWeights
	Opts               map[string]interface{}
	Balances           map[string]*ExtBalance
	ThresholdIDs       []string
}

func (aC *Account) AsExtAccount() (eAc *ExtAccount, err error) {
	eAc = &ExtAccount{
		Tenant: aC.Tenant,
		ID:     aC.ID,
	}
	if aC.FilterIDs != nil {
		eAc.FilterIDs = make([]string, len(aC.FilterIDs))
		for idx, val := range aC.FilterIDs {
			eAc.FilterIDs[idx] = val
		}
	}
	if aC.ActivationInterval != nil {
		eAc.ActivationInterval = aC.ActivationInterval
	}
	if aC.Weights != nil {
		eAc.Weights = aC.Weights
	}
	if aC.Opts != nil {
		eAc.Opts = make(map[string]interface{}, len(aC.Opts))
		for key, val := range aC.Opts {
			eAc.Opts[key] = val
		}
	}
	if aC.Balances != nil {
		eAc.Balances = make(map[string]*ExtBalance, len(aC.Balances))
		for key, val := range aC.Balances {
			if bal, err := val.AsExtBalance(); err != nil {
				return nil, err
			} else {
				eAc.Balances[key] = bal
			}
		}
	}
	if aC.ThresholdIDs != nil {
		eAc.ThresholdIDs = make([]string, len(aC.ThresholdIDs))
		for idx, val := range aC.ThresholdIDs {
			eAc.ThresholdIDs[idx] = val
		}
	}
	return
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

type ExtBalance struct {
	ID             string // Balance identificator, unique within an Account
	FilterIDs      []string
	Weights        DynamicWeights
	Type           string
	Units          *float64
	UnitFactors    []*ExtUnitFactor
	Opts           map[string]interface{}
	CostIncrements []*ExtCostIncrement
	AttributeIDs   []string
	RateProfileIDs []string
}

func (bL *Balance) AsExtBalance() (eBl *ExtBalance, err error) {
	eBl = &ExtBalance{
		ID:   bL.ID,
		Type: bL.Type,
	}
	if bL.FilterIDs != nil {
		eBl.FilterIDs = make([]string, len(bL.FilterIDs))
		for idx, val := range bL.FilterIDs {
			eBl.FilterIDs[idx] = val
		}
	}
	if bL.Weights != nil {
		eBl.Weights = bL.Weights
	}
	if bL.Units != nil {
		if fltUnits, ok := bL.Units.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Units to float64 ")
		} else {
			eBl.Units = &fltUnits
		}
	}
	if bL.UnitFactors != nil {
		eBl.UnitFactors = make([]*ExtUnitFactor, len(bL.UnitFactors))
		for idx, val := range bL.UnitFactors {
			if uFctr, err := val.AsExtUnitFactor(); err != nil {
				return nil, err
			} else {
				eBl.UnitFactors[idx] = uFctr
			}
		}
	}
	if bL.Opts != nil {
		eBl.Opts = make(map[string]interface{}, len(bL.Opts))
		for key, val := range bL.Opts {
			eBl.Opts[key] = val
		}
	}
	if bL.CostIncrements != nil {
		eBl.CostIncrements = make([]*ExtCostIncrement, len(bL.CostIncrements))
		for idx, val := range bL.CostIncrements {
			if extCstIncr, err := val.AsExtCostIncrement(); err != nil {
				return nil, err
			} else {
				eBl.CostIncrements[idx] = extCstIncr
			}
		}
	}
	if bL.AttributeIDs != nil {
		eBl.AttributeIDs = make([]string, len(bL.AttributeIDs))
		for idx, val := range bL.AttributeIDs {
			eBl.AttributeIDs[idx] = val
		}
	}
	if bL.RateProfileIDs != nil {
		eBl.RateProfileIDs = make([]string, len(bL.RateProfileIDs))
		for idx, val := range bL.RateProfileIDs {
			eBl.RateProfileIDs[idx] = val
		}
	}
	return
}

func (bL *Balance) Equals(bal *Balance) (eq bool) {
	if bL.ID != bal.ID || bL.Type != bal.Type {
		return
	}
	if bL.FilterIDs == nil && bal.FilterIDs != nil ||
		bL.FilterIDs != nil && bal.FilterIDs == nil ||
		len(bL.FilterIDs) != len(bal.FilterIDs) {
		return
	}
	for i, val := range bL.FilterIDs {
		if val != bal.FilterIDs[i] {
			return
		}
	}
	if bL.Weights == nil && bal.Weights != nil ||
		bL.Weights != nil && bal.Weights == nil ||
		len(bL.Weights) != len(bal.Weights) {
		return
	}
	for idx, val := range bL.Weights {
		if ok := val.Equals(bal.Weights[idx]); !ok {
			return
		}
	}
	if bL.Units == nil && bal.Units != nil ||
		bL.Units != nil && bal.Units == nil ||
		bL.Units.Compare(bal.Units) != 0 {
		return
	}
	if bL.UnitFactors == nil && bal.UnitFactors != nil ||
		bL.UnitFactors != nil && bal.UnitFactors == nil ||
		len(bL.UnitFactors) != len(bal.UnitFactors) {
		return
	}
	for idx, val := range bL.UnitFactors {
		if ok := val.Equals(bal.UnitFactors[idx]); !ok {
			return
		}
	}
	if bL.Opts == nil && bal.Opts != nil ||
		bL.Opts != nil && bal.Opts == nil ||
		len(bL.Opts) != len(bal.Opts) {
		return
	}
	for key, val := range bL.Opts {
		if val != bal.Opts[key] {
			return
		}
	}
	if bL.CostIncrements == nil && bal.CostIncrements != nil ||
		bL.CostIncrements != nil && bal.CostIncrements == nil ||
		len(bL.CostIncrements) != len(bal.CostIncrements) {
		return
	}
	for idx, val := range bL.CostIncrements {
		if ok := val.Equals(bal.CostIncrements[idx]); !ok {
			return
		}
	}
	if bL.AttributeIDs == nil && bal.AttributeIDs != nil ||
		bL.AttributeIDs != nil && bal.AttributeIDs == nil ||
		len(bL.AttributeIDs) != len(bal.AttributeIDs) {
		return
	}
	for i, val := range bL.AttributeIDs {
		if val != bal.AttributeIDs[i] {
			return
		}
	}
	if bL.RateProfileIDs == nil && bal.RateProfileIDs != nil ||
		bL.RateProfileIDs != nil && bal.RateProfileIDs == nil ||
		len(bL.RateProfileIDs) != len(bal.RateProfileIDs) {
		return
	}
	for i, val := range bL.RateProfileIDs {
		if val != bal.RateProfileIDs[i] {
			return
		}
	}
	return true
}

// CostIncrement enforces cost calculation to specific balance increments
type CostIncrement struct {
	FilterIDs    []string
	Increment    *Decimal
	FixedFee     *Decimal
	RecurrentFee *Decimal
}

type ExtCostIncrement struct {
	FilterIDs    []string
	Increment    *float64
	FixedFee     *float64
	RecurrentFee *float64
}

func (cI *CostIncrement) AsExtCostIncrement() (eCi *ExtCostIncrement, err error) {
	eCi = new(ExtCostIncrement)
	if cI.FilterIDs != nil {
		eCi.FilterIDs = make([]string, len(cI.FilterIDs))
		for idx, val := range cI.FilterIDs {
			eCi.FilterIDs[idx] = val
		}
	}
	if cI.Increment != nil {
		if fltIncr, ok := cI.Increment.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Increment to float64 ")
		} else {
			eCi.Increment = &fltIncr
		}
	}
	if cI.FixedFee != nil {
		if fltFxdFee, ok := cI.FixedFee.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal FixedFee to float64 ")
		} else {
			eCi.FixedFee = &fltFxdFee
		}
	}
	if cI.RecurrentFee != nil {
		if fltRecFee, ok := cI.RecurrentFee.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal RecurrentFee to float64 ")
		} else {
			eCi.RecurrentFee = &fltRecFee
		}
	}
	return
}

// Equals returns the equality between two CostIncrement
func (cI *CostIncrement) Equals(ctIn *CostIncrement) (eq bool) {
	if cI.FilterIDs == nil && ctIn.FilterIDs != nil ||
		cI.FilterIDs != nil && ctIn.FilterIDs == nil ||
		len(cI.FilterIDs) != len(ctIn.FilterIDs) {
		return
	}
	for i, val := range cI.FilterIDs {
		if val != ctIn.FilterIDs[i] {
			return
		}
	}
	if cI.Increment == nil && ctIn.Increment != nil ||
		cI.Increment != nil && ctIn.Increment == nil ||
		cI.RecurrentFee == nil && ctIn.RecurrentFee != nil ||
		cI.RecurrentFee != nil && ctIn.RecurrentFee == nil ||
		cI.FixedFee == nil && ctIn.FixedFee != nil ||
		cI.FixedFee != nil && ctIn.FixedFee == nil {
		return
	}
	return cI.Increment.Compare(ctIn.Increment) == 0 &&
		cI.FixedFee.Compare(ctIn.FixedFee) == 0 &&
		cI.RecurrentFee.Compare(ctIn.RecurrentFee) == 0
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

type ExtUnitFactor struct {
	FilterIDs []string
	Factor    *float64
}

func (uF *UnitFactor) AsExtUnitFactor() (eUf *ExtUnitFactor, err error) {
	eUf = new(ExtUnitFactor)
	if uF.FilterIDs != nil {
		eUf.FilterIDs = make([]string, len(uF.FilterIDs))
		for idx, val := range uF.FilterIDs {
			eUf.FilterIDs[idx] = val
		}
	}
	if uF.Factor != nil {
		if fltFct, ok := uF.Factor.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Factor to float64 ")
		} else {
			eUf.Factor = &fltFct
		}
	}
	return
}

// Clone return a copy of the UnitFactor
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

// Equals compares two UnitFactors
func (uF *UnitFactor) Equals(nUf *UnitFactor) (eq bool) {
	if uF.FilterIDs == nil && nUf.FilterIDs != nil ||
		uF.FilterIDs != nil && nUf.FilterIDs == nil ||
		len(uF.FilterIDs) != len(nUf.FilterIDs) {
		return
	}
	for i := range uF.FilterIDs {
		if uF.FilterIDs[i] != nUf.FilterIDs[i] {
			return
		}
	}
	if uF.Factor == nil && nUf.Factor != nil ||
		uF.Factor != nil && nUf.Factor == nil {
		return
	}
	if uF.Factor == nil && nUf.Factor == nil {
		return true
	}
	return uF.Factor.Compare(nUf.Factor) == 0
}

// TenantID returns the combined Tenant:ID
func (aP *Account) TenantID() string {
	return ConcatenatedKey(aP.Tenant, aP.ID)
}

// Equals return the equality between two Accounts
func (aC *Account) Equals(acnt *Account) (eq bool) {
	if aC.Tenant != acnt.Tenant ||
		aC.ID != acnt.ID {
		return
	}
	if aC.FilterIDs == nil && acnt.FilterIDs != nil ||
		aC.FilterIDs != nil && acnt.FilterIDs == nil ||
		len(aC.FilterIDs) != len(acnt.FilterIDs) {
		return
	}
	for idx, val := range aC.FilterIDs {
		if val != acnt.FilterIDs[idx] {
			return
		}
	}
	if aC.ActivationInterval == nil && acnt.ActivationInterval != nil ||
		aC.ActivationInterval != nil && acnt.ActivationInterval != nil {
		return
	}
	if ok := aC.ActivationInterval.Equals(acnt.ActivationInterval); !ok {
		return
	}
	if aC.Weights == nil && acnt.Weights != nil ||
		aC.Weights != nil && acnt.Weights == nil ||
		len(aC.Weights) != len(acnt.Weights) {
		return
	}
	for idx, val := range aC.Weights {
		if ok := val.Equals(acnt.Weights[idx]); !ok {
			return
		}
	}
	if aC.Opts == nil && acnt.Opts != nil ||
		aC.Opts != nil && acnt.Opts == nil ||
		len(aC.Opts) != len(acnt.Opts) {
		return
	}
	for key := range aC.Opts {
		if aC.Opts[key] != acnt.Opts[key] {
			return
		}
	}
	if aC.Balances == nil && acnt.Balances != nil ||
		aC.Balances != nil && acnt.Balances == nil ||
		len(aC.Balances) != len(acnt.Balances) {
		return
	}
	for key, val := range aC.Balances {
		if ok := val.Equals(acnt.Balances[key]); !ok {
			return
		}
	}
	if aC.ThresholdIDs == nil && acnt.ThresholdIDs != nil ||
		aC.ThresholdIDs != nil && acnt.ThresholdIDs == nil ||
		len(aC.ThresholdIDs) != len(acnt.ThresholdIDs) {
		return
	}
	for idx, val := range aC.ThresholdIDs {
		if val != acnt.ThresholdIDs[idx] {
			return
		}
	}
	return true
}

// Clone returns a clone of the Account
func (aP *Account) Clone() (acnt *Account) {
	acnt = &Account{
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

// AccountWithWeight attaches static weight to Account
type AccountWithWeight struct {
	*Account
	Weight float64
	LockID string
}

// AccountsWithWeight is a sortable list of AccountWithWeight
type AccountsWithWeight []*AccountWithWeight

// Sort is part of sort interface, sort based on Weight
func (aps AccountsWithWeight) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

// Accounts returns the list of Account
func (apWws AccountsWithWeight) Accounts() (aps []*Account) {
	if apWws != nil {
		aps = make([]*Account, len(apWws))
		for i, apWw := range apWws {
			aps[i] = apWw.Account
		}
	}
	return
}

// LockIDs returns the list of LockIDs
func (apWws AccountsWithWeight) LockIDs() (lkIDs []string) {
	if apWws != nil {
		lkIDs = make([]string, len(apWws))
		for i, apWw := range apWws {
			lkIDs[i] = apWw.LockID
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

// APIAccountWithOpts is used in API calls
type APIAccountWithOpts struct {
	*APIAccount
	APIOpts map[string]interface{}
}

type AccountWithAPIOpts struct {
	*Account
	APIOpts map[string]interface{}
}

// ArgsAccountForEvent arguments used for process event
type ArgsAccountsForEvent struct {
	*CGREvent
	AccountIDs []string
}

// APIAccount represents one APIAccount on a Tenant
type APIAccount struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weights            string
	Opts               map[string]interface{}
	Balances           map[string]*APIBalance
	ThresholdIDs       []string
}

// AsAccount convert APIAccount struct to Account struct
func (ext *APIAccount) AsAccount() (profile *Account, err error) {
	profile = &Account{
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

type ArgsActSetBalance struct {
	Tenant    string
	AccountID string
	Diktats   []*BalDiktat
	Reset     bool
	APIOpts   map[string]interface{}
}

type BalDiktat struct {
	Path  string
	Value string
}

type ArgsActRemoveBalances struct {
	Tenant     string
	AccountID  string
	BalanceIDs []string
	APIOpts    map[string]interface{}
}
