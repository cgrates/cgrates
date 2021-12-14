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
	"sort"
	"time"

	"github.com/ericlagergren/decimal"
)

// Account represents one Account on a Tenant
type Account struct {
	Tenant       string
	ID           string // Account identificator, unique within the tenant
	FilterIDs    []string
	Weights      DynamicWeights
	Opts         map[string]interface{}
	Balances     map[string]*Balance
	ThresholdIDs []string
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
			abb[blncID] = CloneDecimalBig(blnc.Units.Big)
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

// Equals returns the equality between two Balance
func (bL *Balance) Equals(bal *Balance) (eq bool) {
	if (bL.ID != bal.ID || bL.Type != bal.Type) ||
		(bL.FilterIDs == nil && bal.FilterIDs != nil ||
			bL.FilterIDs != nil && bal.FilterIDs == nil ||
			len(bL.FilterIDs) != len(bal.FilterIDs)) ||
		(bL.Weights == nil && bal.Weights != nil ||
			bL.Weights != nil && bal.Weights == nil ||
			len(bL.Weights) != len(bal.Weights)) ||
		(bL.Units == nil && bal.Units != nil ||
			bL.Units != nil && bal.Units == nil ||
			bL.Units.Compare(bal.Units) != 0) ||
		(bL.UnitFactors == nil && bal.UnitFactors != nil ||
			bL.UnitFactors != nil && bal.UnitFactors == nil ||
			len(bL.UnitFactors) != len(bal.UnitFactors)) ||
		(bL.Opts == nil && bal.Opts != nil ||
			bL.Opts != nil && bal.Opts == nil ||
			len(bL.Opts) != len(bal.Opts)) ||
		(bL.CostIncrements == nil && bal.CostIncrements != nil ||
			bL.CostIncrements != nil && bal.CostIncrements == nil ||
			len(bL.CostIncrements) != len(bal.CostIncrements)) ||
		(bL.AttributeIDs == nil && bal.AttributeIDs != nil ||
			bL.AttributeIDs != nil && bal.AttributeIDs == nil ||
			len(bL.AttributeIDs) != len(bal.AttributeIDs)) ||
		(bL.RateProfileIDs == nil && bal.RateProfileIDs != nil ||
			bL.RateProfileIDs != nil && bal.RateProfileIDs == nil ||
			len(bL.RateProfileIDs) != len(bal.RateProfileIDs)) {
		return
	}
	for i, val := range bL.FilterIDs {
		if val != bal.FilterIDs[i] {
			return
		}
	}
	for idx, val := range bL.Weights {
		if ok := val.Equals(bal.Weights[idx]); !ok {
			return
		}
	}
	for idx, val := range bL.UnitFactors {
		if ok := val.Equals(bal.UnitFactors[idx]); !ok {
			return
		}
	}
	for key, val := range bL.Opts {
		if val != bal.Opts[key] {
			return
		}
	}
	for idx, val := range bL.CostIncrements {
		if ok := val.Equals(bal.CostIncrements[idx]); !ok {
			return
		}
	}
	for i, val := range bL.AttributeIDs {
		if val != bal.AttributeIDs[i] {
			return
		}
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

// Equals returns the equality between two CostIncrement
func (cI *CostIncrement) Equals(ctIn *CostIncrement) (eq bool) {
	if (cI.FilterIDs == nil && ctIn.FilterIDs != nil ||
		cI.FilterIDs != nil && ctIn.FilterIDs == nil ||
		len(cI.FilterIDs) != len(ctIn.FilterIDs)) ||
		(cI.Increment == nil && ctIn.Increment != nil ||
			cI.Increment != nil && ctIn.Increment == nil ||
			(cI.Increment != nil && ctIn.Increment != nil &&
				cI.Increment.Compare(ctIn.Increment) != 0)) ||
		(cI.RecurrentFee == nil && ctIn.RecurrentFee != nil ||
			cI.RecurrentFee != nil && ctIn.RecurrentFee == nil ||
			(cI.RecurrentFee != nil && ctIn.RecurrentFee != nil &&
				cI.RecurrentFee.Compare(ctIn.RecurrentFee) != 0)) ||
		(cI.FixedFee == nil && ctIn.FixedFee != nil ||
			cI.FixedFee != nil && ctIn.FixedFee == nil ||
			(cI.FixedFee != nil && ctIn.FixedFee != nil &&
				cI.FixedFee.Compare(ctIn.FixedFee) != 0)) {
		return
	}
	for i, val := range cI.FilterIDs {
		if val != ctIn.FilterIDs[i] {
			return
		}
	}
	return true
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
	if (uF.FilterIDs == nil && nUf.FilterIDs != nil ||
		uF.FilterIDs != nil && nUf.FilterIDs == nil ||
		len(uF.FilterIDs) != len(nUf.FilterIDs)) ||
		(uF.Factor == nil && nUf.Factor != nil ||
			uF.Factor != nil && nUf.Factor == nil ||
			(uF.Factor != nil && nUf.Factor != nil &&
				uF.Factor.Compare(nUf.Factor) != 0)) {
		return
	}
	for idx, val := range uF.FilterIDs {
		if val != nUf.FilterIDs[idx] {
			return
		}
	}
	return true
}

// TenantID returns the combined Tenant:ID
func (aP *Account) TenantID() string {
	return ConcatenatedKey(aP.Tenant, aP.ID)
}

// Equals return the equality between two Accounts
func (aC *Account) Equals(acnt *Account) (eq bool) {
	if (aC.Tenant != acnt.Tenant ||
		aC.ID != acnt.ID) ||
		(aC.FilterIDs == nil && acnt.FilterIDs != nil ||
			aC.FilterIDs != nil && acnt.FilterIDs == nil ||
			len(aC.FilterIDs) != len(acnt.FilterIDs)) ||
		(aC.Weights == nil && acnt.Weights != nil ||
			aC.Weights != nil && acnt.Weights == nil ||
			len(aC.Weights) != len(acnt.Weights)) ||
		(aC.Opts == nil && acnt.Opts != nil ||
			aC.Opts != nil && acnt.Opts == nil ||
			len(aC.Opts) != len(acnt.Opts)) ||
		(aC.Balances == nil && acnt.Balances != nil ||
			aC.Balances != nil && acnt.Balances == nil ||
			len(aC.Balances) != len(acnt.Balances)) ||
		(aC.ThresholdIDs == nil && acnt.ThresholdIDs != nil ||
			aC.ThresholdIDs != nil && acnt.ThresholdIDs == nil ||
			len(aC.ThresholdIDs) != len(acnt.ThresholdIDs)) {
		return
	}
	for idx, val := range aC.FilterIDs {
		if val != acnt.FilterIDs[idx] {
			return
		}
	}
	for idx, val := range aC.Weights {
		if ok := val.Equals(acnt.Weights[idx]); !ok {
			return
		}
	}
	for key := range aC.Opts {
		if aC.Opts[key] != acnt.Opts[key] {
			return
		}
	}
	for key, val := range aC.Balances {
		if ok := val.Equals(acnt.Balances[key]); !ok {
			return
		}
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
		Tenant:  aP.Tenant,
		ID:      aP.ID,
		Weights: aP.Weights.Clone(),
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

// APIAccount represents one APIAccount on a Tenant
type APIAccount struct {
	Tenant       string
	ID           string
	FilterIDs    []string
	Weights      string
	Opts         map[string]interface{}
	Balances     map[string]*APIBalance
	ThresholdIDs []string
}

// AsAccount convert APIAccount struct to Account struct
func (ext *APIAccount) AsAccount() (profile *Account, err error) {
	profile = &Account{
		Tenant:       ext.Tenant,
		ID:           ext.ID,
		FilterIDs:    ext.FilterIDs,
		Opts:         ext.Opts,
		ThresholdIDs: ext.ThresholdIDs,
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
	Units          string
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
		Opts:           ext.Opts,
		AttributeIDs:   ext.AttributeIDs,
		RateProfileIDs: ext.RateProfileIDs,
	}
	if ext.Units != EmptyString {
		if balance.Units, err = NewDecimalFromUsage(ext.Units); err != nil {
			return nil, err
		}
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
			if balance.CostIncrements[i], err = cIncr.AsCostIncrement(); err != nil {
				return nil, err
			}
		}
	}
	return
}

// APICostIncrement represent one CostIncrement inside an APIBalance
type APICostIncrement struct {
	FilterIDs    []string
	Increment    string
	FixedFee     *float64
	RecurrentFee *float64
}

// AsCostIncrement convert APICostIncrement struct to CostIncrement struct
func (ext *APICostIncrement) AsCostIncrement() (cIncr *CostIncrement, err error) {
	cIncr = &CostIncrement{
		FilterIDs: ext.FilterIDs,
	}
	if ext.FixedFee != nil {
		cIncr.FixedFee = NewDecimalFromFloat64(*ext.FixedFee)
	}
	if ext.RecurrentFee != nil {
		cIncr.RecurrentFee = NewDecimalFromFloat64(*ext.RecurrentFee)
	}
	if ext.Increment != EmptyString {
		cIncr.Increment, err = NewDecimalFromUsage(ext.Increment)
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
