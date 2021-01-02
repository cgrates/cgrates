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
	ID          string // Balance identificator, unique within an Account
	FilterIDs   []string
	Weight      float64
	Blocker     bool
	Type        string
	Opts        map[string]interface{}
	UnitFactors []*UnitFactor
	Value       float64

	val *decimal.Big
}

// UnitFactor is a multiplicator for the usage received
type UnitFactor struct {
	FilterIDs []string
	Factor    float64

	fct *decimal.Big
}

// DecimalFactor exports the decimal value of the factor
func (uf *UnitFactor) DecimalFactor() *decimal.Big {
	return uf.fct
}

func (aP *AccountProfile) TenantID() string {
	return ConcatenatedKey(aP.Tenant, aP.ID)
}

// Clone returns a clone of the Account
func (aP *AccountProfile) Clone() (acnt *AccountProfile) {
	return
}

// Compile populates the internal data
func (aP *AccountProfile) Compile() (err error) {
	return
}

// Compile populates the internal data
func (b *Balance) Compile() (err error) {
	b.val = new(decimal.Big).SetFloat64(b.Value)
	for _, uf := range b.UnitFactors {
		uf.fct = new(decimal.Big).SetFloat64(uf.Factor)
	}
	return
}

// SetDecimalValue populates the internal decimal value
func (b *Balance) SetDecimalValue(dVal *decimal.Big) {
	b.val = dVal
}

// DecimalValue returns the internal decimal value
func (b *Balance) DecimalValue() *decimal.Big {
	return b.val
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

type Account struct {
	Tenant string
	ID     string
}

func (ac *Account) TenantID() string {
	return ConcatenatedKey(ac.Tenant, ac.ID)
}

type AccountWithOpts struct {
	*Account
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
