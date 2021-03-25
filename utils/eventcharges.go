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
	Accounts          []*AccountProfile

	Accounting  map[string]*AccountCharge
	UnitFactors map[string]*UnitFactor
	Rating      map[string]*RateSInterval
}

// Merge will merge the event charges into existing
func (ec *EventCharges) Merge(eCs ...*EventCharges) {
	for _, nEc := range eCs {
		if ec.Abstracts != nil {
			ec.Abstracts = &Decimal{SumBig(ec.Abstracts.Big, nEc.Abstracts.Big)}
		} else { // initial
			ec.Abstracts = nEc.Abstracts
		}
		if ec.Concretes != nil {
			ec.Concretes = &Decimal{SumBig(ec.Concretes.Big, nEc.Concretes.Big)}
		} else { // initial
			ec.Concretes = nEc.Concretes
		}
		ec.AppendChargingIntervals(ec.ChargingIntervals...)
	}
}

// SyncIDs will repopulate Accounting, UnitFactors and  Rating IDs if they equal the references in ec
func (ec *EventCharges) SyncIDs(eCs ...*EventCharges) {
	for _, nEc := range eCs {
		for _, cIl := range nEc.ChargingIntervals {
			for _, cIcrm := range cIl.Increments {

				nEcAcntChrg := nEc.Accounting[cIcrm.AccountChargeID]

				// UnitFactors
				if nEcAcntChrg.UnitFactorID != EmptyString {
					if uFctID := ec.UnitFactorID(nEc.UnitFactors[nEcAcntChrg.UnitFactorID]); uFctID != EmptyString &&
						uFctID != nEcAcntChrg.UnitFactorID {
						nEc.UnitFactors[uFctID] = ec.UnitFactors[uFctID]
						delete(nEc.UnitFactors, nEcAcntChrg.UnitFactorID)
						nEcAcntChrg.UnitFactorID = uFctID
					}
				}

				// Rating
				if nEcAcntChrg.RatingID != EmptyString {
					if rtID := ec.RatingID(nEc.Rating[nEcAcntChrg.RatingID]); rtID != EmptyString &&
						rtID != nEcAcntChrg.RatingID {
						nEc.Rating[rtID] = ec.Rating[rtID]
						delete(nEc.Rating, nEcAcntChrg.RatingID)
						nEcAcntChrg.RatingID = rtID
					}
				}

				// AccountCharges
				for i, chrgID := range nEc.Accounting[cIcrm.AccountChargeID].JoinedChargeIDs {
					if acntChrgID := ec.AccountChargeID(nEc.Accounting[chrgID]); acntChrgID != chrgID {
						// matched the AccountChargeID with an existing one in reference ec, replace in nEc
						nEc.Accounting[acntChrgID] = ec.Accounting[acntChrgID]
						delete(nEc.Accounting, chrgID)
						nEc.Accounting[cIcrm.AccountChargeID].JoinedChargeIDs[i] = acntChrgID
					}
				}
				if acntChrgID := ec.AccountChargeID(nEcAcntChrg); acntChrgID != EmptyString &&
					acntChrgID != cIcrm.AccountChargeID {
					// matched the AccountChargeID with an existing one in reference ec, replace in nEc
					nEc.Accounting[acntChrgID] = ec.Accounting[cIcrm.AccountChargeID]
					delete(nEc.Accounting, cIcrm.AccountChargeID)
					cIcrm.AccountChargeID = acntChrgID
				}

			}
		}
	}
}

// AppendChargingIntervals will add new charging intervals to the  existing.
// if possible, the existing last one in ec will be compressed
func (ec *EventCharges) AppendChargingIntervals(cIls ...*ChargingInterval) {
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

// UnitFactorID returns the ID of the matching UnitFactor within ec.UnitFactors
func (ec *EventCharges) UnitFactorID(uF *UnitFactor) (ufID string) {
	for ecUfID, ecUf := range ec.UnitFactors {
		if ecUf.Equals(uF) {
			return ecUfID
		}
	}
	return
}

// RatingID returns the ID of the matching RateSInterval within ec.Rating
func (ec *EventCharges) RatingID(rIl *RateSInterval) (rID string) {
	for ecID, ecRtIl := range ec.Rating {
		if ecRtIl.Equals(rIl) {
			return ecID
		}
	}
	return
}

// AccountChargeID returns the ID of the matching AccountCharge within ec.Accounting
func (ec *EventCharges) AccountChargeID(ac *AccountCharge) (acID string) {
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
	return
}

// ChargingIncrement represents one unit charged inside an interval
type ChargingIncrement struct {
	Units           *Decimal // Can differ from AccountCharge due to JoinedCharging
	AccountChargeID string   // Account charging information
	CompressFactor  int
}

// AccountCharge represents one Account charge
type AccountCharge struct {
	AccountID       string
	BalanceID       string
	Units           *Decimal
	BalanceLimit    *Decimal // the minimum balance value accepted
	UnitFactorID    string   // identificator in ChargingUnitFactors
	AttributeIDs    []string // list of attribute profiles matched
	RatingID        string   // identificator in cost increments
	JoinedChargeIDs []string // identificator of extra account charges
}

// Equals compares two AccountCharges
func (ac *AccountCharge) Equals(nAc *AccountCharge) (eq bool) {
	return
}
