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
		Accounts:    make(map[string]*Account),
	}
	return
}

// EventCharges records the charges applied to an Event
type EventCharges struct {
	Abstracts *Decimal // total abstract units charged
	Concretes *Decimal // total concrete units charged

	Charges []*ChargeEntry

	Accounting  map[string]*AccountCharge
	UnitFactors map[string]*UnitFactor
	Rating      map[string]*RateSInterval
	Accounts    map[string]*Account
}

// ChargeEntry is a reference towards Accounting or Rating ID (depending on request type)
type ChargeEntry struct {
	ChargingID     string
	CompressFactor int
}

// Merge will merge the event charges into existing
func (ec *EventCharges) Merge(eCs ...*EventCharges) {
	//ec.SyncIDs(eCs...) // so we can compare properly
	for _, nEc := range eCs {
		if sumAbst := SumDecimalAsBig(ec.Abstracts, nEc.Abstracts); sumAbst != nil {
			ec.Abstracts = &Decimal{sumAbst}
		}
		if sumCrct := SumDecimalAsBig(ec.Concretes, nEc.Concretes); sumCrct != nil {
			ec.Concretes = &Decimal{sumCrct}
		}
		ec.appendChargeEntry(nEc.Charges...)
		for acntID, acntChrg := range nEc.Accounting {
			ec.Accounting[acntID] = acntChrg
		}
		for ufID, uF := range nEc.UnitFactors {
			ec.UnitFactors[ufID] = uF
		}
		for riID, rI := range nEc.Rating {
			ec.Rating[riID] = rI
		}
		for acntID, acnt := range nEc.Accounts {
			ec.Accounts[acntID] = acnt
		}
	}
}

// appendChargeEntry will add new charge to the  existing.
// if possible, the existing last one in ec will be compressed
func (ec *EventCharges) appendChargeEntry(cIls ...*ChargeEntry) {
	for i, cIl := range cIls {
		if i == 0 && len(ec.Charges) == 0 {
			ec.Charges = []*ChargeEntry{cIl}
			continue
		}
		if ec.Charges[len(ec.Charges)-1].CompressEquals(cIl) {
			ec.Charges[len(ec.Charges)-1].CompressFactor += 1
			continue
		}
		ec.Charges = append(ec.Charges, cIl)
	}
}

func (cE *ChargeEntry) CompressEquals(chEn *ChargeEntry) bool {
	return cE.ChargingID == chEn.ChargingID
}

// SyncIDs will repopulate Accounting, UnitFactors and  Rating IDs if they equal the references in ec
func (ec *EventCharges) SyncIDs(eCs ...*EventCharges) {
	for _, nEc := range eCs {
		for _, cIl := range nEc.Charges {
			nEcAcntChrg := nEc.Accounting[cIl.ChargingID]

			// UnitFactors
			if nEcAcntChrg.UnitFactorID != EmptyString {
				if uFctID := ec.unitFactorID(nEc.UnitFactors[nEcAcntChrg.UnitFactorID]); uFctID != EmptyString &&
					uFctID != nEcAcntChrg.UnitFactorID {
					nEc.UnitFactors[uFctID] = ec.UnitFactors[uFctID]
					delete(nEc.UnitFactors, nEcAcntChrg.UnitFactorID)
					nEcAcntChrg.UnitFactorID = uFctID
				}
			}

			// Rating
			if nEcAcntChrg.RatingID != EmptyString {
				if rtID := ec.ratingID(nEc.Rating[nEcAcntChrg.RatingID]); rtID != EmptyString &&
					rtID != nEcAcntChrg.RatingID {
					nEc.Rating[rtID] = ec.Rating[rtID]
					delete(nEc.Rating, nEcAcntChrg.RatingID)
					nEcAcntChrg.RatingID = rtID
				}
			}

			// AccountCharges
			for i, chrgID := range nEc.Accounting[cIl.ChargingID].JoinedChargeIDs {
				if acntChrgID := ec.accountChargeID(nEc.Accounting[chrgID]); acntChrgID != chrgID {
					// matched the accountChargeID with an existing one in reference ec, replace in nEc
					nEc.Accounting[acntChrgID] = ec.Accounting[acntChrgID]
					delete(nEc.Accounting, chrgID)
					nEc.Accounting[cIl.ChargingID].JoinedChargeIDs[i] = acntChrgID
				}
			}
			if acntChrgID := ec.accountChargeID(nEcAcntChrg); acntChrgID != EmptyString &&
				acntChrgID != cIl.ChargingID {
				// matched the accountChargeID with an existing one in reference ec, replace in nEc
				nEc.Accounting[acntChrgID] = ec.Accounting[cIl.ChargingID]
				delete(nEc.Accounting, cIl.ChargingID)
				cIl.ChargingID = acntChrgID
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
	if ec.Charges != nil {
		eEc.Charges = make([]*ChargeEntry, len(ec.Charges))
		for idx, val := range ec.Charges {
			eEc.Charges[idx] = val
		}
	}
	if ec.Accounting != nil {
		eEc.Accounting = make(map[string]*ExtAccountCharge, len(eEc.Accounting))
		for key, val := range ec.Accounting {
			if extAcc, err := val.AsExtAccountCharge(); err != nil {
				return nil, err
			} else {
				eEc.Accounting[key] = extAcc
			}
		}
	}
	if ec.UnitFactors != nil {
		eEc.UnitFactors = make(map[string]*ExtUnitFactor, len(ec.UnitFactors))
		for key, val := range ec.UnitFactors {
			if extUnit, err := val.AsExtUnitFactor(); err != nil {
				return nil, err
			} else {
				eEc.UnitFactors[key] = extUnit
			}
		}
	}
	if ec.Rating != nil {
		eEc.Rating = make(map[string]*ExtRateSInterval, len(ec.Rating))
		for key, val := range ec.Rating {
			if extRate, err := val.AsExtRateSInterval(); err != nil {
				return nil, err
			} else {
				eEc.Rating[key] = extRate
			}
		}
	}
	if ec.Accounts != nil {
		eEc.Accounts = make(map[string]*ExtAccount, len(ec.Accounts))
		for acntID, acnt := range ec.Accounts {
			if extAccs, err := acnt.AsExtAccount(); err != nil {
				return nil, err
			} else {
				eEc.Accounts[acntID] = extAccs
			}
		}
	}
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

	Charges []*ChargeEntry

	Accounting  map[string]*ExtAccountCharge
	UnitFactors map[string]*ExtUnitFactor
	Rating      map[string]*ExtRateSInterval
	Accounts    map[string]*ExtAccount
}

type ExtChargingIncrement struct {
	Units           *float64
	AccountChargeID string
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

type ExtAccountCharge struct {
	AccountID       string
	BalanceID       string
	Units           *float64
	BalanceLimit    *float64 // the minimum balance value accepted(float64 type)
	UnitFactorID    string   // identificator in ChargingUnitFactors
	AttributeIDs    []string // list of attribute profiles matched
	RatingID        string   // identificator in cost increments
	JoinedChargeIDs []string // identificator of extra account charges
}

func (aC *AccountCharge) AsExtAccountCharge() (eAc *ExtAccountCharge, err error) {
	eAc = &ExtAccountCharge{
		AccountID:    aC.AccountID,
		BalanceID:    aC.BalanceID,
		UnitFactorID: aC.UnitFactorID,
		RatingID:     aC.RatingID,
	}
	if aC.Units != nil {
		if fltUnit, ok := aC.Units.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Units to float64 ")
		} else {
			eAc.Units = &fltUnit
		}
	}
	if aC.BalanceLimit != nil {
		if fltBlUnit, ok := aC.BalanceLimit.Big.Float64(); !ok {
			return nil, errors.New("cannot convert decimal BalanceLimit to float64 ")
		} else {
			eAc.BalanceLimit = &fltBlUnit
		}
	}
	if aC.AttributeIDs != nil {
		eAc.AttributeIDs = make([]string, len(aC.AttributeIDs))
		for idx, val := range aC.AttributeIDs {
			eAc.AttributeIDs[idx] = val
		}
	}
	if aC.JoinedChargeIDs != nil {
		eAc.JoinedChargeIDs = make([]string, len(aC.JoinedChargeIDs))
		for idx, val := range aC.JoinedChargeIDs {
			eAc.JoinedChargeIDs[idx] = val
		}
	}
	return
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
		ac.UnitFactorID != nAc.UnitFactorID ||
		ac.RatingID != nAc.RatingID {
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
