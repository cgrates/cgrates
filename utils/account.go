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
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
)

// Account represents one Account on a Tenant
type Account struct {
	Tenant       string
	ID           string // Account identificator, unique within the tenant
	FilterIDs    []string
	Weights      DynamicWeights
	Blockers     DynamicBlockers
	Opts         map[string]any
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
	Blockers       DynamicBlockers
	Type           string
	Units          *Decimal
	UnitFactors    []*UnitFactor
	Opts           map[string]any
	CostIncrements []*CostIncrement
	AttributeIDs   []string
	RateProfileIDs []string
}

// Equals returns the equality between two Balance
func (bL *Balance) Equals(bal *Balance) (eq bool) {
	if bL.ID != bal.ID || bL.Type != bal.Type ||
		(bL.FilterIDs == nil && bal.FilterIDs != nil ||
			bL.FilterIDs != nil && bal.FilterIDs == nil ||
			len(bL.FilterIDs) != len(bal.FilterIDs)) ||
		(bL.Weights == nil && bal.Weights != nil ||
			bL.Weights != nil && bal.Weights == nil ||
			len(bL.Weights) != len(bal.Weights)) ||
		(bL.Blockers == nil && bal.Blockers != nil ||
			bL.Blockers != nil && bal.Blockers == nil ||
			len(bL.Blockers) != len(bal.Blockers)) ||
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
	if cI.FilterIDs == nil && ctIn.FilterIDs != nil ||
		cI.FilterIDs != nil && ctIn.FilterIDs == nil ||
		len(cI.FilterIDs) != len(ctIn.FilterIDs) ||
		(cI.Increment == nil && ctIn.Increment != nil ||
			cI.Increment != nil && ctIn.Increment == nil ||
			cI.Increment != nil && ctIn.Increment != nil &&
				cI.Increment.Compare(ctIn.Increment) != 0) ||
		(cI.RecurrentFee == nil && ctIn.RecurrentFee != nil ||
			cI.RecurrentFee != nil && ctIn.RecurrentFee == nil ||
			cI.RecurrentFee != nil && ctIn.RecurrentFee != nil &&
				cI.RecurrentFee.Compare(ctIn.RecurrentFee) != 0) ||
		(cI.FixedFee == nil && ctIn.FixedFee != nil ||
			cI.FixedFee != nil && ctIn.FixedFee == nil ||
			cI.FixedFee != nil && ctIn.FixedFee != nil &&
				cI.FixedFee.Compare(ctIn.FixedFee) != 0) {
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
		copy(cIcln.FilterIDs, cI.FilterIDs)
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

// Clone returns a copy of uF
func (uF *UnitFactor) Clone() *UnitFactor {
	cln := new(UnitFactor)
	if uF.FilterIDs != nil {
		cln.FilterIDs = slices.Clone(uF.FilterIDs)
	}
	if uF.Factor != nil {
		cln.Factor = uF.Factor.Clone()
	}
	return cln
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
		len(uF.FilterIDs) != len(nUf.FilterIDs) ||
		(uF.Factor == nil && nUf.Factor != nil ||
			uF.Factor != nil && nUf.Factor == nil ||
			uF.Factor != nil && nUf.Factor != nil &&
				uF.Factor.Compare(nUf.Factor) != 0) {
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
	if aC.Tenant != acnt.Tenant ||
		aC.ID != acnt.ID ||
		(aC.FilterIDs == nil && acnt.FilterIDs != nil ||
			aC.FilterIDs != nil && acnt.FilterIDs == nil ||
			len(aC.FilterIDs) != len(acnt.FilterIDs)) ||
		(aC.Blockers == nil && acnt.Blockers != nil ||
			aC.Blockers != nil && acnt.Blockers == nil ||
			len(aC.Blockers) != len(acnt.Blockers)) ||
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
func (acc *Account) Clone() (cln *Account) {
	cln = &Account{
		Tenant:   acc.Tenant,
		ID:       acc.ID,
		Blockers: acc.Blockers.Clone(),
		Weights:  acc.Weights.Clone(),
	}
	if acc.FilterIDs != nil {
		cln.FilterIDs = make([]string, len(acc.FilterIDs))
		copy(cln.FilterIDs, acc.FilterIDs)
	}
	if acc.Opts != nil {
		cln.Opts = make(map[string]any)
		for key, value := range acc.Opts {
			cln.Opts[key] = value
		}
	}
	if acc.Balances != nil {
		cln.Balances = make(map[string]*Balance, len(acc.Balances))
		for i, value := range acc.Balances {
			cln.Balances[i] = value.Clone()
		}
	}
	if acc.ThresholdIDs != nil {
		cln.ThresholdIDs = make([]string, len(acc.ThresholdIDs))
		copy(cln.ThresholdIDs, acc.ThresholdIDs)
	}
	return
}

// CacheClone returns a clone of Account used by ltcache CacheCloner
func (acc *Account) CacheClone() any {
	return acc.Clone()
}

// Clone returns a clone of the ActivationInterval
func (aI *ActivationInterval) Clone() *ActivationInterval {
	if aI == nil {
		return nil
	}
	return &ActivationInterval{
		ActivationTime: aI.ActivationTime,
		ExpiryTime:     aI.ExpiryTime,
	}
}

// Clone return a clone of the Balance
func (blnc *Balance) Clone() (cln *Balance) {
	cln = &Balance{
		ID:       blnc.ID,
		Weights:  blnc.Weights.Clone(),
		Blockers: blnc.Blockers.Clone(),
		Type:     blnc.Type,
	}
	if blnc.FilterIDs != nil {
		cln.FilterIDs = make([]string, len(blnc.FilterIDs))
		copy(cln.FilterIDs, blnc.FilterIDs)
	}
	if blnc.Units != nil {
		cln.Units = blnc.Units.Clone()
	}
	if blnc.UnitFactors != nil {
		cln.UnitFactors = make([]*UnitFactor, len(blnc.UnitFactors))
		for i, value := range blnc.UnitFactors {
			cln.UnitFactors[i] = value.Clone()
		}
	}
	if blnc.Opts != nil {
		cln.Opts = make(map[string]any)
		for key, value := range blnc.Opts {
			cln.Opts[key] = value
		}
	}
	if blnc.CostIncrements != nil {
		cln.CostIncrements = make([]*CostIncrement, len(blnc.CostIncrements))
		for i, value := range blnc.CostIncrements {
			cln.CostIncrements[i] = value.Clone()
		}
	}
	if blnc.AttributeIDs != nil {
		cln.AttributeIDs = make([]string, len(blnc.AttributeIDs))
		copy(cln.AttributeIDs, blnc.AttributeIDs)
	}
	if blnc.RateProfileIDs != nil {
		cln.RateProfileIDs = make([]string, len(blnc.RateProfileIDs))
		copy(cln.RateProfileIDs, blnc.RateProfileIDs)
	}
	return
}

// AccountWithLock wraps Account with LockID.
type AccountWithLock struct {
	*Account
	LockID string
}

// Accounts is a collection of AccountWithLock objects
type Accounts []*AccountWithLock

// Accounts returns the list of Account
func (apWws Accounts) Accounts() (aps []*Account) {
	if apWws != nil {
		aps = make([]*Account, len(apWws))
		for i, apWw := range apWws {
			aps[i] = apWw.Account
		}
	}
	return
}

// LockIDs returns the list of LockIDs
func (apWws Accounts) LockIDs() (lkIDs []string) {
	if apWws != nil {
		lkIDs = make([]string, len(apWws))
		for i, apWw := range apWws {
			lkIDs[i] = apWw.LockID
		}
	}
	return
}

// Account returns the Account object with ID
func (apWws Accounts) Account(acntID string) (acnt *Account) {
	for _, aWw := range apWws {
		if aWw.Account.ID == acntID {
			acnt = aWw.Account
			break
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

type AccountWithAPIOpts struct {
	*Account
	APIOpts map[string]any
}

type ArgsActSetBalance struct {
	Tenant    string
	AccountID string
	Diktats   []*BalDiktat
	Reset     bool
	APIOpts   map[string]any
}

type BalDiktat struct {
	Path  string
	Value string
}

type ArgsActRemoveBalances struct {
	Tenant     string
	AccountID  string
	BalanceIDs []string
	APIOpts    map[string]any
}

func (ap *Account) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	case 0:
		return ErrWrongPath
	case 1:
		switch path[0] {
		default:
			if strings.HasPrefix(path[0], Opts) &&
				path[0][4] == '[' && path[0][len(path[0])-1] == ']' {
				ap.Opts[path[0][5:len(path[0])-1]] = val
				return
			}
			// if strings.HasPrefix(path[0], Balances) &&
			// 	path[0][8] == '[' && path[0][len(path[0])-1] == ']' {
			// 	id := path[0][9 : len(path[0])-1]
			// 	if _, has := ap.Balances[id]; !has {
			// 		ap.Balances[id] = &Balance{ID: id, Opts: make(map[string]any), Units: NewDecimal(0, 0)}
			// 	}
			// 	return ap.Balances[id].Set(path[1:], val, newBranch)
			// }
			return ErrWrongPath
		case Tenant:
			ap.Tenant = IfaceAsString(val)
		case ID:
			ap.ID = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			ap.FilterIDs = append(ap.FilterIDs, valA...)
		case ThresholdIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			ap.ThresholdIDs = append(ap.ThresholdIDs, valA...)
		case Weights:
			if val != EmptyString {
				ap.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Blockers:
			if val != EmptyString {
				ap.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Opts:
			ap.Opts, err = NewMapFromCSV(IfaceAsString(val))
		}
		return
	default:
	}
	if path[0] == Opts {
		return MapStorage(ap.Opts).Set(path[1:], val)
	}
	if strings.HasPrefix(path[0], Opts) &&
		path[0][4] == '[' && path[0][len(path[0])-1] == ']' {
		return MapStorage(ap.Opts).Set(append([]string{path[0][5 : len(path[0])-1]}, path[1:]...), val)
	}
	var id string
	if path[0] == Balances {
		id = path[1]
		path = path[1:]
	} else if strings.HasPrefix(path[0], Balances) &&
		path[0][8] == '[' && path[0][len(path[0])-1] == ']' {
		id = path[0][9 : len(path[0])-1]
	}
	if id != EmptyString {
		if _, has := ap.Balances[id]; !has {
			ap.Balances[id] = &Balance{ID: path[0], Opts: make(map[string]any), Units: NewDecimal(0, 0)}
		}
		return ap.Balances[id].Set(path[1:], val, newBranch)
	}
	return ErrWrongPath
}

func (bL *Balance) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
	case 0:
		return ErrWrongPath
	case 1:
		switch path[0] {
		default:
			if strings.HasPrefix(path[0], Opts) &&
				path[0][4] == '[' && path[0][len(path[0])-1] == ']' {
				bL.Opts[path[0][5:len(path[0])-1]] = val
				return
			}
			return ErrWrongPath
		case ID:
			bL.ID = IfaceAsString(val)
		case Type:
			bL.Type = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			bL.FilterIDs = append(bL.FilterIDs, valA...)
		case AttributeIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			bL.AttributeIDs = append(bL.AttributeIDs, valA...)
		case RateProfileIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			bL.RateProfileIDs = append(bL.RateProfileIDs, valA...)
		case Units:
			var valB *decimal.Big
			valB, err = IfaceAsBig(val)
			bL.Units = &Decimal{valB}
		case Weights:
			if val != EmptyString {
				bL.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Blockers:
			if val != EmptyString {
				bL.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case UnitFactors:
			if ufStr := IfaceAsString(val); len(ufStr) != 0 {
				sls := strings.Split(ufStr, InfieldSep)
				if len(sls)%2 != 0 {
					return fmt.Errorf("invalid key: <%s> for BalanceUnitFactors", IfaceAsString(val))
				}
				for j := 0; j < len(sls); j += 2 {
					uf := new(UnitFactor)
					if len(sls[j]) != 0 {
						uf.FilterIDs = strings.Split(sls[j], ANDSep)
					}

					var valB *decimal.Big
					if valB, err = IfaceAsBig(sls[j+1]); err != nil {
						return
					}
					uf.Factor = &Decimal{valB}
					bL.UnitFactors = append(bL.UnitFactors, uf)
				}
			}
		case CostIncrements:
			if ciStr := IfaceAsString(val); len(ciStr) != 0 {
				sls := strings.Split(ciStr, InfieldSep)
				if len(sls)%4 != 0 {
					return fmt.Errorf("invalid key: <%s> for BalanceCostIncrements", IfaceAsString(val))
				}
				for j := 0; j < len(sls); j += 4 {
					cI := new(CostIncrement)
					if len(sls[j]) != 0 {
						cI.FilterIDs = strings.Split(sls[j], ANDSep)
					}
					if len(sls[j+1]) != 0 {
						if cI.Increment, err = NewDecimalFromString(sls[j+1]); err != nil {
							return
						}
					}
					if len(sls[j+2]) != 0 {
						if cI.FixedFee, err = NewDecimalFromString(sls[j+2]); err != nil {
							return
						}
					}
					if len(sls[j+3]) != 0 {
						if cI.RecurrentFee, err = NewDecimalFromString(sls[j+3]); err != nil {
							return
						}
					}
					bL.CostIncrements = append(bL.CostIncrements, cI)
				}
			}
		case Opts:
			bL.Opts, err = NewMapFromCSV(IfaceAsString(val))
		}
		return
	case 2:
		switch path[0] {
		default:
		case UnitFactors:
			if len(bL.UnitFactors) == 0 || newBranch {
				bL.UnitFactors = append(bL.UnitFactors, &UnitFactor{Factor: NewDecimal(0, 0)})
			}
			uf := bL.UnitFactors[len(bL.UnitFactors)-1]
			switch path[1] {
			default:
				return ErrWrongPath
			case FilterIDs:
				var valA []string
				valA, err = IfaceAsStringSlice(val)
				uf.FilterIDs = append(uf.FilterIDs, valA...)
			case Factor:
				if val != EmptyString {
					var valB *decimal.Big
					valB, err = IfaceAsBig(val)
					uf.Factor = &Decimal{valB}
				}
			}
			return
		case CostIncrements:
			if len(bL.CostIncrements) == 0 || newBranch {
				bL.CostIncrements = append(bL.CostIncrements, &CostIncrement{FixedFee: NewDecimal(0, 0)})
			}
			cI := bL.CostIncrements[len(bL.CostIncrements)-1]
			switch path[1] {
			default:
				return ErrWrongPath
			case FilterIDs:
				var valA []string
				valA, err = IfaceAsStringSlice(val)
				cI.FilterIDs = append(cI.FilterIDs, valA...)
			case Increment:
				if val != EmptyString {
					var valB *decimal.Big
					valB, err = IfaceAsBig(val)
					cI.Increment = &Decimal{valB}
				}
			case FixedFee:
				if val != EmptyString {
					var valB *decimal.Big
					valB, err = IfaceAsBig(val)
					cI.FixedFee = &Decimal{valB}
				}
			case RecurrentFee:
				if val != EmptyString {
					var valB *decimal.Big
					valB, err = IfaceAsBig(val)
					cI.RecurrentFee = &Decimal{valB}
				}
			}
			return
		}
	}

	if path[0] == Opts {
		return MapStorage(bL.Opts).Set(path[1:], val)
	}
	if strings.HasPrefix(path[0], Opts) &&
		path[0][4] == '[' && path[0][len(path[0])-1] == ']' {
		return MapStorage(bL.Opts).Set(append([]string{path[0][5 : len(path[0])-1]}, path[1:]...), val)
	}
	return ErrWrongPath
}

func (ap *Account) Merge(v2 any) {
	vi := v2.(*Account)
	if len(vi.Tenant) != 0 {
		ap.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		ap.ID = vi.ID
	}
	ap.FilterIDs = append(ap.FilterIDs, vi.FilterIDs...)
	ap.Weights = append(ap.Weights, vi.Weights...)
	ap.Blockers = append(ap.Blockers, vi.Blockers...)
	ap.ThresholdIDs = append(ap.ThresholdIDs, vi.ThresholdIDs...)
	for k, v := range vi.Opts {
		ap.Opts[k] = v
	}
	for k, v := range vi.Balances {
		if bl, has := ap.Balances[k]; has {
			bl.Merge(v)
			continue
		}
		ap.Balances[k] = v
	}
}

func (ap *Account) String() string { return ToJSON(ap) }
func (ap *Account) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = ap.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (ap *Account) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idxStr := GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.FilterIDs) {
						return ap.FilterIDs[idx], nil
					}
				case Weights:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return nil, err
					}
					if idx < len(ap.Weights) {
						return ap.Weights[idx], nil
					}
				case Blockers:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return nil, err
					}
					if idx < len(ap.Blockers) {
						return ap.Blockers[idx], nil
					}
				case ThresholdIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.ThresholdIDs) {
						return ap.ThresholdIDs[idx], nil
					}
				case Opts:
					return MapStorage(ap.Opts).FieldAsInterface([]string{*idxStr})
				case Balances:
					if rt, has := ap.Balances[*idxStr]; has {
						return rt, nil
					}
				}
			}
			return nil, ErrNotFound
		case Tenant:
			return ap.Tenant, nil
		case ID:
			return ap.ID, nil
		case FilterIDs:
			return ap.FilterIDs, nil
		case Weights:
			return ap.Weights.String(InfieldSep, ANDSep), nil
		case Blockers:
			return ap.Blockers.String(InfieldSep, ANDSep), nil
		case ThresholdIDs:
			return ap.ThresholdIDs, nil
		case Opts:
			return ap.Opts, nil
		case Balances:
			return ap.Balances, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	fld, idxStr := GetPathIndexString(fldPath[0])
	switch fld {
	default:
		return nil, ErrNotFound
	case Opts:
		path := fldPath[1:]
		if idxStr != nil {
			path = append([]string{*idxStr}, path...)
		}
		return MapStorage(ap.Opts).FieldAsInterface(path)
	case Balances:
		if idxStr == nil {
			idxStr = &fldPath[1]
			fldPath = fldPath[1:]
		}
		bl, has := ap.Balances[*idxStr]
		if !has {
			return nil, ErrNotFound
		}
		if len(fldPath) == 1 {
			return bl, nil
		}
		return bl.FieldAsInterface(fldPath[1:])
	case Weights:
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(ap.Weights) {
			return nil, fmt.Errorf("invalid index for '%s' field", Weights)
		}
		return ap.Weights[idx].FieldAsInterface(fldPath[1:])
	case Blockers:
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(ap.Blockers) {
			return nil, fmt.Errorf("invalid index for '%s' field", Blockers)
		}
		return ap.Blockers[idx].FieldAsInterface(fldPath[1:])
	}
}

func (bL *Balance) String() string { return ToJSON(bL) }
func (bL *Balance) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = bL.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (bL *Balance) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idxStr := GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(bL.FilterIDs) {
						return bL.FilterIDs[idx], nil
					}
				case Weights:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return nil, err
					}
					if idx < len(bL.Weights) {
						return bL.Weights[idx], nil
					}
				case Blockers:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return nil, err
					}
					if idx < len(bL.Blockers) {
						return bL.Blockers[idx], nil
					}
				case AttributeIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(bL.AttributeIDs) {
						return bL.AttributeIDs[idx], nil
					}
				case RateProfileIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(bL.RateProfileIDs) {
						return bL.RateProfileIDs[idx], nil
					}
				case UnitFactors:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(bL.UnitFactors) {
						return bL.UnitFactors[idx], nil
					}
				case CostIncrements:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(bL.CostIncrements) {
						return bL.CostIncrements[idx], nil
					}
				case Opts:
					return MapStorage(bL.Opts).FieldAsInterface([]string{*idxStr})
				}
			}
			return nil, ErrNotFound
		case Type:
			return bL.Type, nil
		case ID:
			return bL.ID, nil
		case FilterIDs:
			return bL.FilterIDs, nil
		case Weights:
			return bL.Weights.String(InfieldSep, ANDSep), nil
		case Blockers:
			return bL.Blockers.String(InfieldSep, ANDSep), nil
		case AttributeIDs:
			return bL.AttributeIDs, nil
		case Units:
			return bL.Units, nil
		case RateProfileIDs:
			return bL.RateProfileIDs, nil
		case Opts:
			return bL.Opts, nil
		case UnitFactors:
			return bL.UnitFactors, nil
		case CostIncrements:
			return bL.CostIncrements, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	fld, idxStr := GetPathIndexString(fldPath[0])
	switch fld {
	default:
		return nil, ErrNotFound
	case Weights:
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(bL.Weights) {
			return nil, fmt.Errorf("invalid index for '%s' field", Weights)
		}
		return bL.Weights[idx].FieldAsInterface(fldPath[1:])
	case Blockers:
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(bL.Blockers) {
			return nil, fmt.Errorf("invalid index for '%s' field", Blockers)
		}
		return bL.Blockers[idx].FieldAsInterface(fldPath[1:])
	case Opts:
		path := fldPath[1:]
		if idxStr != nil {
			path = append([]string{*idxStr}, path...)
		}
		return MapStorage(bL.Opts).FieldAsInterface(path)
	case UnitFactors:
		if idxStr == nil {
			return nil, ErrNotFound
		}
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(bL.UnitFactors) {
			return nil, ErrNotFound
		}
		return bL.UnitFactors[idx].FieldAsInterface(fldPath[1:])
	case CostIncrements:
		if idxStr == nil {
			return nil, ErrNotFound
		}
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(bL.CostIncrements) {
			return nil, ErrNotFound
		}
		return bL.CostIncrements[idx].FieldAsInterface(fldPath[1:])
	}
}

func (uF *UnitFactor) String() string { return ToJSON(uF) }
func (uF *UnitFactor) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = uF.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (uF *UnitFactor) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil &&
			fld == FilterIDs {
			if *idx < len(uF.FilterIDs) {
				return uF.FilterIDs[*idx], nil
			}
		}
		return nil, ErrNotFound
	case FilterIDs:
		return uF.FilterIDs, nil
	case Factor:
		return uF.Factor, nil
	}
}

func (cI *CostIncrement) String() string { return ToJSON(cI) }
func (cI *CostIncrement) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = cI.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (cI *CostIncrement) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil &&
			fld == FilterIDs {
			if *idx < len(cI.FilterIDs) {
				return cI.FilterIDs[*idx], nil
			}
		}
		return nil, ErrNotFound
	case FilterIDs:
		return cI.FilterIDs, nil
	case Increment:
		return cI.Increment, nil
	case FixedFee:
		return cI.FixedFee, nil
	case RecurrentFee:
		return cI.RecurrentFee, nil
	}
}

func (bL *Balance) Merge(vi *Balance) {
	if len(vi.ID) != 0 {
		bL.ID = vi.ID
	}
	if len(vi.Type) != 0 {
		bL.Type = vi.Type
	}
	if vi.Units != nil {
		bL.Units = vi.Units
	}
	bL.FilterIDs = append(bL.FilterIDs, vi.FilterIDs...)
	bL.Weights = append(bL.Weights, vi.Weights...)
	bL.Blockers = append(bL.Blockers, vi.Blockers...)
	bL.UnitFactors = append(bL.UnitFactors, vi.UnitFactors...)
	bL.CostIncrements = append(bL.CostIncrements, vi.CostIncrements...)
	bL.AttributeIDs = append(bL.AttributeIDs, vi.AttributeIDs...)
	bL.RateProfileIDs = append(bL.RateProfileIDs, vi.RateProfileIDs...)
	for k, v := range vi.Opts {
		bL.Opts[k] = v
	}
}
