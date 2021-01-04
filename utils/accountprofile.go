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

// AccountProfile represents one Account on a Tenant
type AccountProfile struct {
	Tenant             string
	ID                 string // Account identificator, unique within the tenant
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weight             float64
	Opts               map[string]interface{}
	Balances           []*Balance
	ThresholdIDs       []string
}

// Balance represents one Balance inside an Account
type Balance struct {
	ID             string // Balance identificator, unique within an Account
	FilterIDs      []string
	Weight         float64
	Blocker        bool
	Type           string
	Opts           map[string]interface{}
	CostIncrements []*CostIncrement
	CostAttributes []string
	UnitFactors    []*UnitFactor
	Value          float64
}

// CostIncrement enforces cost calculation to specific balance increments
type CostIncrement struct {
	FilterIDs    []string
	Increment    *decimal.Big
	FixedFee     *decimal.Big
	RecurrentFee *decimal.Big
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
		cIcln.Increment = new(decimal.Big).Copy(cI.Increment)
	}
	if cI.FixedFee != nil {
		cIcln.FixedFee = new(decimal.Big).Copy(cI.FixedFee)
	}
	if cI.RecurrentFee != nil {
		cIcln.RecurrentFee = new(decimal.Big).Copy(cI.RecurrentFee)
	}
	return
}

//Clone return a copy of the CostAttributes
func (cA *CostAttributes) Clone() (cstAtr *CostAttributes) {
	cstAtr = new(CostAttributes)
	if cA.FilterIDs != nil {
		cstAtr.FilterIDs = make([]string, len(cA.FilterIDs))
		for i, value := range cA.FilterIDs {
			cstAtr.FilterIDs[i] = value
		}
	}
	if cA.AttributeProfileIDs != nil {
		cstAtr.AttributeProfileIDs = make([]string, len(cA.AttributeProfileIDs))
		for i, value := range cA.AttributeProfileIDs {
			cstAtr.AttributeProfileIDs[i] = value
		}
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
		untFct.Factor = new(decimal.Big).Copy(uF.Factor)
	}
	return
}

// UnitFactor is a multiplicator for the usage received
type UnitFactor struct {
	FilterIDs []string
	Factor    *decimal.Big
}

func (aP *AccountProfile) TenantID() string {
	return ConcatenatedKey(aP.Tenant, aP.ID)
}

// Clone returns a clone of the Account
func (aP *AccountProfile) Clone() (acnt *AccountProfile) {
	acnt = &AccountProfile{
		Tenant:             aP.Tenant,
		ID:                 aP.ID,
		Weight:             aP.Weight,
		Opts:               make(map[string]interface{}),
		ActivationInterval: aP.ActivationInterval.Clone(),
	}
	if aP.FilterIDs != nil {
		acnt.FilterIDs = make([]string, len(aP.FilterIDs))
		for i, value := range aP.FilterIDs {
			acnt.FilterIDs[i] = value
		}
	}
	for key, value := range aP.Opts {
		acnt.Opts[key] = value
	}
	if aP.ThresholdIDs != nil {
		acnt.ThresholdIDs = make([]string, len(aP.ThresholdIDs))
		for i, value := range aP.ThresholdIDs {
			acnt.ThresholdIDs[i] = value
		}
	}
	if aP.Balances != nil {
		acnt.Balances = make([]*Balance, len(aP.Balances))
		for i, value := range aP.Balances {
			acnt.Balances[i] = value.Clone()
		}
	}
	return
}

//Clone returns a clone of the ActivationInterval
func (aI *ActivationInterval) Clone() *ActivationInterval {
	return &ActivationInterval{
		ActivationTime: aI.ActivationTime,
		ExpiryTime:     aI.ExpiryTime,
	}
}

//Clone return a clone of the Balance
func (bL *Balance) Clone() (blnc *Balance) {
	blnc = &Balance{
		ID:      bL.ID,
		Weight:  bL.Weight,
		Blocker: bL.Blocker,
		Type:    bL.Type,
		Opts:    make(map[string]interface{}),
		Value:   bL.Value,
	}
	if bL.FilterIDs != nil {
		blnc.FilterIDs = make([]string, len(bL.FilterIDs))
		for i, value := range bL.FilterIDs {
			blnc.FilterIDs[i] = value
		}
	}
	for key, value := range bL.Opts {
		blnc.Opts[key] = value
	}
	if bL.CostIncrements != nil {
		blnc.CostIncrements = make([]*CostIncrement, len(bL.CostIncrements))
		for i, value := range bL.CostIncrements {
			blnc.CostIncrements[i] = value.Clone()
		}
	}
	if bL.CostAttributes != nil {
		blnc.CostAttributes = make([]*CostAttributes, len(bL.CostAttributes))
		for i, value := range bL.CostAttributes {
			blnc.CostAttributes[i] = value.Clone()
		}
	}
	if bL.UnitFactors != nil {
		blnc.UnitFactors = make([]*UnitFactor, len(bL.UnitFactors))
		for i, value := range bL.UnitFactors {
			blnc.UnitFactors[i] = value.Clone()
		}
	}
	return
}

// ActionProfiles is a sortable list of ActionProfiles
type AccountProfiles []*AccountProfile

// Sort is part of sort interface, sort based on Weight
func (aps AccountProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

// Balances is a sortable list of Balances
type Balances []*Balance

// Sort is part of sort interface, sort based on Weight
func (blcs Balances) Sort() {
	sort.Slice(blcs, func(i, j int) bool { return blcs[i].Weight > blcs[j].Weight })
}

// AccountProfileWithOpts is used in API calls
type AccountProfileWithOpts struct {
	*AccountProfile
	Opts map[string]interface{}
}

// ArgsAccountForEvent arguments used for process event
type ArgsAccountForEvent struct {
	*CGREventWithOpts
	AccountIDs []string
}

type ReplyMaxUsage struct {
	AccountID string
	MaxUsage  time.Duration
	Cost      *EventCharges
}
