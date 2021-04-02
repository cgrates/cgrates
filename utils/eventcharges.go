/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
)

// NewEventChargers instantiates the EventChargers in a central place
func NewEventCharges() (ec *EventCharges) {
	ec = &EventCharges{
		Accounting:  make(map[string]*AccountCharge),
		UnitFactors: make(map[string]*UnitFactor),
		Rating:      make(map[string]*RateSInterval),
	}
	return
}

// EventCharges records the charges applied to an Event
type EventCharges struct {
	Abstracts *Decimal // total abstract units charged
	Concretes *Decimal // total concrete units charged

	ChargingIntervals []*ChargingInterval
	Accounts          []*Account

	Accounting  map[string]*AccountCharge
	UnitFactors map[string]*UnitFactor
	Rating      map[string]*RateSInterval
}

// Merge will merge the event charges into existing
func (ec *EventCharges) Merge(eCs ...*EventCharges) {
	ec.syncIDs(eCs...) // so we can compare properly
	for _, nEc := range eCs {
		if sumAbst := SumDecimalAsBig(ec.Abstracts, nEc.Abstracts); sumAbst != nil {
			ec.Abstracts = &Decimal{sumAbst}
		}
		if sumCrct := SumDecimalAsBig(ec.Concretes, nEc.Concretes); sumCrct != nil {
			ec.Concretes = &Decimal{sumCrct}
		}
		ec.appendChargingIntervals(ec.ChargingIntervals...)
	}
}

// appendChargingIntervals will add new charging intervals to the  existing.
// if possible, the existing last one in ec will be compressed
func (ec *EventCharges) appendChargingIntervals(cIls ...*ChargingInterval) {
	for i, cIl := range cIls {
		if i == 0 && len(ec.ChargingIntervals) == 0 {
			ec.ChargingIntervals = []*ChargingInterval{cIl}
			continue
		}
		if ec.ChargingIntervals[len(ec.ChargingIntervals)-1].CompressEquals(cIl) {
			ec.ChargingIntervals[len(ec.ChargingIntervals)-1].CompressFactor += 1
			continue
		}
		ec.ChargingIntervals = append(ec.ChargingIntervals, cIl)
	}
}

// syncIDs will repopulate Accounting, UnitFactors and  Rating IDs if they equal the references in ec
func (ec *EventCharges) syncIDs(eCs ...*EventCharges) {
	for _, nEc := range eCs {
		for _, cIl := range nEc.ChargingIntervals {
			for _, cIcrm := range cIl.Increments {
				nEcAcntChrg := nEc.Accounting[cIcrm.accountChargeID]

				// UnitFactors
				if nEcAcntChrg.unitFactorID != EmptyString {
					if uFctID := ec.unitFactorID(nEc.UnitFactors[nEcAcntChrg.unitFactorID]); uFctID != EmptyString &&
						uFctID != nEcAcntChrg.unitFactorID {
						nEc.UnitFactors[uFctID] = ec.UnitFactors[uFctID]
						delete(nEc.UnitFactors, nEcAcntChrg.unitFactorID)
						nEcAcntChrg.unitFactorID = uFctID
					}
				}

				// Rating
				if nEcAcntChrg.ratingID != EmptyString {
					if rtID := ec.ratingID(nEc.Rating[nEcAcntChrg.ratingID]); rtID != EmptyString &&
						rtID != nEcAcntChrg.ratingID {
						nEc.Rating[rtID] = ec.Rating[rtID]
						delete(nEc.Rating, nEcAcntChrg.ratingID)
						nEcAcntChrg.ratingID = rtID
					}
				}

				// AccountCharges
				for i, chrgID := range nEc.Accounting[cIcrm.accountChargeID].JoinedChargeIDs {
					if acntChrgID := ec.accountChargeID(nEc.Accounting[chrgID]); acntChrgID != chrgID {
						// matched the accountChargeID with an existing one in reference ec, replace in nEc
						nEc.Accounting[acntChrgID] = ec.Accounting[acntChrgID]
						delete(nEc.Accounting, chrgID)
						nEc.Accounting[cIcrm.accountChargeID].JoinedChargeIDs[i] = acntChrgID
					}
				}
				if acntChrgID := ec.accountChargeID(nEcAcntChrg); acntChrgID != EmptyString &&
					acntChrgID != cIcrm.accountChargeID {
					// matched the accountChargeID with an existing one in reference ec, replace in nEc
					nEc.Accounting[acntChrgID] = ec.Accounting[cIcrm.accountChargeID]
					delete(nEc.Accounting, cIcrm.accountChargeID)
					cIcrm.accountChargeID = acntChrgID
				}

			}
		}
	}
}

// AsExtEventCharges converts EventCharges to ExtEventCharges
func (ec *EventCharges) AsExtEventCharges() (eEc *ExtEventCharges, err error) {
	eEc = new(ExtEventCharges)
	if ec.Abstracts != nil {
		if flt, ok := ec.Abstracts.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Abstracts to float64")
		} else {
			eEc.Abstracts = &flt
		}
	}
	if ec.Concretes != nil {
		if flt, ok := ec.Concretes.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Concretes to float64")
		} else {
			eEc.Concretes = &flt
		}
	}
	// add here code for the rest of the fields
	return
}

// unitFactorID returns the ID of the matching UnitFactor within ec.UnitFactors
func (ec *EventCharges) unitFactorID(uF *UnitFactor) (ufID string) {
	for ecUfID, ecUf := range ec.UnitFactors {
		if ecUf.Equals(uF) {
			return ecUfID
		}
	}
	return
}

// ratingID returns the ID of the matching RateSInterval within ec.Rating
func (ec *EventCharges) ratingID(rIl *RateSInterval) (rID string) {
	for ecID, ecRtIl := range ec.Rating {
		if ecRtIl.Equals(rIl) {
			return ecID
		}
	}
	return
}

// accountChargeID returns the ID of the matching AccountCharge within ec.Accounting
func (ec *EventCharges) accountChargeID(ac *AccountCharge) (acID string) {
	for ecID, ecAc := range ec.Accounting {
		if ecAc.Equals(ac) {
			return ecID
		}
	}
	return
}

// ExtEventCharges is a generic EventCharges used in APIs
type ExtEventCharges struct {
	Abstracts *float64
	Concretes *float64
}

type ChargingInterval struct {
	Increments     []*ChargingIncrement // specific increments applied to this interval
	CompressFactor int
}

// CompressEquals compares two ChargingIntervals for aproximate equality, ignoring compress field
func (cIl *ChargingInterval) CompressEquals(nCil *ChargingInterval) (eq bool) {
	if len(cIl.Increments) != len(nCil.Increments) {
		return
	}
	for i, chIr := range cIl.Increments {
		if !chIr.CompressEquals(nCil.Increments[i]) {
			return
		}
	}
	return true
}

// ChargingIncrement represents one unit charged inside an interval
type ChargingIncrement struct {
	Units           *Decimal // Can differ from AccountCharge due to JoinedCharging
	accountChargeID string   // Account charging information
	CompressFactor  int
}

func (cI *ChargingIncrement) CompressEquals(chIh *ChargingIncrement) (eq bool) {
	if cI.Units == nil && chIh.Units != nil ||
		cI.Units != nil && chIh.Units == nil {
		return
	}
	return cI.Units.Compare(chIh.Units) == 0 &&
		cI.accountChargeID == chIh.accountChargeID
}

// AccountCharge represents one Account charge
type AccountCharge struct {
	AccountID       string
	BalanceID       string
	Units           *Decimal
	BalanceLimit    *Decimal // the minimum balance value accepted
	unitFactorID    string   // identificator in ChargingUnitFactors
	AttributeIDs    []string // list of attribute profiles matched
	ratingID        string   // identificator in cost increments
	JoinedChargeIDs []string // identificator of extra account charges
}

// Equals compares two AccountCharges
func (ac *AccountCharge) Equals(nAc *AccountCharge) (eq bool) {
	if ac.AttributeIDs == nil && nAc.AttributeIDs != nil ||
		ac.AttributeIDs != nil && nAc.AttributeIDs == nil ||
		len(ac.AttributeIDs) != len(nAc.AttributeIDs) {
		return
	}
	for i := range ac.AttributeIDs {
		if ac.AttributeIDs[i] != nAc.AttributeIDs[i] {
			return
		}
	}
	if ac.JoinedChargeIDs == nil && nAc.JoinedChargeIDs != nil ||
		ac.JoinedChargeIDs != nil && nAc.JoinedChargeIDs == nil ||
		len(ac.JoinedChargeIDs) != len(nAc.JoinedChargeIDs) {
		return
	}
	for i := range ac.JoinedChargeIDs {
		if ac.JoinedChargeIDs[i] != nAc.JoinedChargeIDs[i] {
			return
		}
	}
	if ac.AccountID != nAc.AccountID ||
		ac.BalanceID != nAc.BalanceID ||
		ac.unitFactorID != nAc.unitFactorID ||
		ac.ratingID != nAc.ratingID {
		return
	}
	if ac.Units == nil && nAc.Units != nil ||
		ac.Units != nil && nAc.Units == nil {
		return
	}
	if ac.BalanceLimit == nil && nAc.BalanceLimit != nil ||
		ac.BalanceLimit != nil && nAc.BalanceLimit == nil {
		return
	}
	if ac.Units == nil && nAc.Units == nil ||
		ac.BalanceLimit == nil && nAc.BalanceLimit == nil {
		return true
	}
	return ac.Units.Compare(nAc.Units) == 0 && ac.BalanceLimit.Compare(nAc.BalanceLimit) == 0
}
