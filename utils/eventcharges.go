/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package utils

import (
	"fmt"
	"slices"
	"strings"
)

// NewEventChargers instantiates the EventChargers in a central place
func NewEventCharges() (ec *EventCharges) {
	ec = &EventCharges{
		Accounting:  make(map[string]*AccountCharge),
		UnitFactors: make(map[string]*UnitFactor),
		Rating:      make(map[string]*RateSInterval),
		Rates:       make(map[string]*IntervalRate),
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
	Rates       map[string]*IntervalRate
	Accounts    map[string]*Account
}

// Clone returns a copy of ec
func (ec *EventCharges) Clone() *EventCharges {
	cln := new(EventCharges)
	if ec.Abstracts != nil {
		cln.Abstracts = ec.Abstracts.Clone()
	}
	if ec.Concretes != nil {
		cln.Concretes = ec.Concretes.Clone()
	}
	if ec.Charges != nil {
		cln.Charges = make([]*ChargeEntry, len(ec.Charges))
		for i, value := range ec.Charges {
			cln.Charges[i] = value.Clone()
		}
	}
	if ec.Accounting != nil {
		cln.Accounting = make(map[string]*AccountCharge)
		for key, value := range ec.Accounting {
			cln.Accounting[key] = value.Clone()
		}
	}
	if ec.UnitFactors != nil {
		cln.UnitFactors = make(map[string]*UnitFactor)
		for key, value := range ec.UnitFactors {
			cln.UnitFactors[key] = value.Clone()
		}
	}
	if ec.Rating != nil {
		cln.Rating = make(map[string]*RateSInterval)
		for key, value := range ec.Rating {
			cln.Rating[key] = value.Clone()
		}
	}
	if ec.Rates != nil {
		cln.Rates = make(map[string]*IntervalRate)
		for key, value := range ec.Rates {
			cln.Rates[key] = value.Clone()
		}
	}
	if ec.Accounts != nil {
		cln.Accounts = make(map[string]*Account)
		for key, value := range ec.Accounts {
			cln.Accounts[key] = value.Clone()
		}
	}

	return cln
}

// ChargeEntry is a reference towards Accounting or Rating ID (depending on request type)
type ChargeEntry struct {
	ChargingID     string
	CompressFactor int
}

// Clone returns a copy of ce
func (ce *ChargeEntry) Clone() *ChargeEntry {
	return &ChargeEntry{
		ChargingID:     ce.ChargingID,
		CompressFactor: ce.CompressFactor,
	}
}

func (ce *ChargeEntry) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	case CompressFactor:
		return ce.CompressFactor, nil
	case ChargingID:
		return ce.ChargingID, nil
	}
	return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
}

// Merge will merge the event charges into existing
func (ec *EventCharges) Merge(eCs ...*EventCharges) {
	//ec.SyncIDs(eCs...) // so we can compare properly
	for _, nEc := range eCs {
		ec.Abstracts = SumDecimal(ec.Abstracts, nEc.Abstracts)
		ec.Concretes = SumDecimal(ec.Concretes, nEc.Concretes)
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
		for rateID, rI := range nEc.Rates {
			ec.Rates[rateID] = rI
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

// Equals return the equality between two ChargeEntry ignoring CompressFactor
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
				if rtID := ec.ratingID(nEc.Rating[nEcAcntChrg.RatingID], nEc.Rates); rtID != EmptyString &&
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

// Equals returns the equality between two EventCharges
func (eC *EventCharges) Equals(evCh *EventCharges) (eq bool) {
	if eC == nil && evCh == nil {
		return true
	}
	if eC == nil && evCh != nil ||
		eC != nil && evCh == nil ||
		(eC.Abstracts == nil && evCh.Abstracts != nil ||
			eC.Abstracts != nil && evCh.Abstracts == nil ||
			eC.Abstracts != nil && evCh.Abstracts != nil &&
				eC.Abstracts.Compare(evCh.Abstracts) != 0) ||
		(eC.Concretes == nil && evCh.Concretes != nil ||
			eC.Concretes != nil && evCh.Concretes == nil ||
			eC.Concretes != nil && evCh.Concretes != nil &&
				eC.Concretes.Compare(evCh.Concretes) != 0) ||
		(eC.Charges == nil && evCh.Charges != nil ||
			eC.Charges != nil && evCh.Charges == nil ||
			len(eC.Charges) != len(evCh.Charges)) ||
		(eC.Accounting == nil && evCh.Accounting != nil ||
			eC.Accounting != nil && evCh.Accounting == nil ||
			len(eC.Accounting) != len(evCh.Accounting)) ||
		(eC.UnitFactors == nil && evCh.UnitFactors != nil ||
			eC.UnitFactors != nil && evCh.UnitFactors == nil ||
			len(eC.UnitFactors) != len(evCh.UnitFactors)) ||
		(eC.Rating == nil && evCh.Rating != nil ||
			eC.Rating != nil && evCh.Rating == nil ||
			len(eC.Rating) != len(evCh.Rating)) ||
		(eC.Rates == nil && evCh.Rates != nil ||
			eC.Rates != nil && evCh.Rates == nil ||
			len(eC.Rates) != len(evCh.Rates)) ||
		(eC.Accounts == nil && evCh.Accounts != nil ||
			eC.Accounts != nil && evCh.Accounts == nil ||
			len(eC.Accounts) != len(evCh.Accounts)) {
		return
	}
	for idx, ch1 := range eC.Charges {
		if ch2 := evCh.Charges[idx]; ch1.CompressFactor != ch2.CompressFactor ||
			!equalsAccounting(eC.Accounting[ch1.ChargingID], evCh.Accounting[ch2.ChargingID],
				eC.Accounting, evCh.Accounting, eC.UnitFactors, evCh.UnitFactors,
				eC.Rating, evCh.Rating, eC.Rates, evCh.Rates) {
			return
		}
	}
	for key, val := range eC.Accounts {
		if ok := val.Equals(evCh.Accounts[key]); !ok {
			return
		}
	}
	return true
}

func equalsAccounting(acc1, acc2 *AccountCharge,
	accM1, accM2 map[string]*AccountCharge,
	uf1, uf2 map[string]*UnitFactor,
	rat1, rat2 map[string]*RateSInterval,
	rts1, rts2 map[string]*IntervalRate) (_ bool) {
	if !acc1.equals(acc2) ||
		uf1 != nil && uf2 != nil &&
			acc1.UnitFactorID != EmptyString && acc2.UnitFactorID != EmptyString &&
			!uf1[acc1.UnitFactorID].Equals(uf2[acc2.UnitFactorID]) ||
		!rat1[acc1.RatingID].Equals(rat2[acc2.RatingID], rts1, rts2) {
		return
	}
	for idx, jc1 := range acc1.JoinedChargeIDs {
		jc2 := acc2.JoinedChargeIDs[idx]
		if !equalsAccounting(accM1[jc1], accM2[jc2], accM1, accM2,
			uf1, uf2, rat1, rat2, rts1, rts2) {
			return
		}
	}
	return true
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
func (ec *EventCharges) ratingID(rIl *RateSInterval, nIrRef map[string]*IntervalRate) (rID string) {
	for ecID, ecRtIl := range ec.Rating {
		if ecRtIl.Equals(rIl, ec.Rates, nIrRef) {
			return ecID
		}
	}
	return
}

// accountChargeID returns the ID of the matching AccountCharge within ec.Accounting
func (ec *EventCharges) accountChargeID(ac *AccountCharge) (acID string) {
	for ecID, ecAc := range ec.Accounting {
		if ecAc.equals(ac) {
			return ecID
		}
	}
	return
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

func (ac *AccountCharge) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	case AccountID:
		return ac.AccountID, nil
	case BalanceID:
		return ac.BalanceID, nil
	case Units:
		return ac.Units, nil
	case BalanceLimit:
		return ac.BalanceLimit, nil
	case UnitFactorID:
		return ac.UnitFactorID, nil
	case AttributeIDs:
		return ac.AttributeIDs, nil
	case RatingID:
		return ac.RatingID, nil
	case JoinedChargeIDs:
		return ac.JoinedChargeIDs, nil
	default:
		opath, indx := GetPathIndex(fldPath[0])
		if indx != nil {
			switch opath {
			case AttributeIDs:
				if *indx < len(ac.AttributeIDs) {
					return ac.AttributeIDs[*indx], nil
				}
			case JoinedChargeIDs:
				if *indx < len(ac.JoinedChargeIDs) {
					return ac.JoinedChargeIDs[*indx], nil
				}
			}
			return nil, ErrNotFound
		}
	}
	return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
}

// Clone returns a copy of ac
func (ac *AccountCharge) Clone() *AccountCharge {
	cln := &AccountCharge{
		AccountID:    ac.AccountID,
		BalanceID:    ac.BalanceID,
		UnitFactorID: ac.UnitFactorID,
		RatingID:     ac.RatingID,
	}
	if ac.Units != nil {
		cln.Units = ac.Units.Clone()
	}
	if ac.BalanceLimit != nil {
		cln.BalanceLimit = ac.BalanceLimit.Clone()
	}
	if ac.AttributeIDs != nil {
		cln.AttributeIDs = slices.Clone(ac.AttributeIDs)
	}
	if ac.JoinedChargeIDs != nil {
		cln.JoinedChargeIDs = slices.Clone(ac.JoinedChargeIDs)
	}
	return cln
}

// Equals compares two AccountCharges
func (ac *AccountCharge) equals(nAc *AccountCharge) (eq bool) {
	if ac == nil && nAc == nil {
		return true
	}
	if ac.AttributeIDs == nil && nAc.AttributeIDs != nil ||
		ac.AttributeIDs != nil && nAc.AttributeIDs == nil ||
		len(ac.AttributeIDs) != len(nAc.AttributeIDs) ||
		ac.JoinedChargeIDs == nil && nAc.JoinedChargeIDs != nil ||
		ac.JoinedChargeIDs != nil && nAc.JoinedChargeIDs == nil ||
		len(ac.JoinedChargeIDs) != len(nAc.JoinedChargeIDs) ||
		(ac.AccountID != nAc.AccountID ||
			ac.BalanceID != nAc.BalanceID) ||
		len(ac.UnitFactorID) == 0 != (len(nAc.UnitFactorID) == 0) ||
		len(ac.RatingID) == 0 != (len(nAc.RatingID) == 0) ||
		(ac.Units == nil && nAc.Units != nil ||
			ac.Units != nil && nAc.Units == nil ||
			ac.Units != nil && nAc.Units != nil &&
				ac.Units.Compare(nAc.Units) != 0) ||
		(ac.BalanceLimit == nil && nAc.BalanceLimit != nil ||
			ac.BalanceLimit != nil && nAc.BalanceLimit == nil ||
			ac.BalanceLimit != nil && nAc.BalanceLimit != nil &&
				ac.BalanceLimit.Compare(nAc.BalanceLimit) != 0) {
		return
	}
	for idx, val := range ac.AttributeIDs {
		if val != nAc.AttributeIDs[idx] {
			return
		}
	}
	return true
}

// APIEventCharges is used in APIs, ie: refundCharges
type APIEventCharges struct {
	Tenant  string
	APIOpts map[string]any
	*EventCharges
}

// String returns the string representation of an *EventCharges value. Needed
// in order to satisfy the DataProvider interface.
func (ec *EventCharges) String() string {
	return ToJSON(ec)
}

// FieldAsInterface retrieves a value of a field based on its path from within an *EventCharges
// value. Needed in order to satisfy the DataProvider interface.
func (ec *EventCharges) FieldAsInterface(fldPath []string) (any, error) {
	switch fldPath[0] {
	default: // "e.g. Charges[1]"
		opath, indx := GetPathIndex(fldPath[0])
		if opath != Charges {
			return nil, fmt.Errorf("unsupported field prefix: <%s>", opath)
		}
		if indx == nil {
			return nil, fmt.Errorf("invalid index for '%s' field", Charges)
		}
		if len(ec.Charges) <= *indx {
			return nil, ErrNotFound
		}
		return ec.getChargesForPath(fldPath[1:], ec.Charges[*indx])
	case Abstracts:
		if len(fldPath) != 1 {
			return nil, ErrNotFound
		}
		return ec.Abstracts, nil
	case Concretes:
		if len(fldPath) != 1 {
			return nil, ErrNotFound
		}
		return ec.Concretes, nil
	case Charges:
		if len(fldPath) != 1 { // slice has no members
			return nil, ErrNotFound
		}
		return ec.Charges, nil
	case Accounting:
		if len(fldPath) == 1 {
			return ec.Accounting, nil
		}
		if ec.Accounting == nil {
			return nil, ErrNotFound
		}
		accCharge, has := ec.Accounting[fldPath[1]]
		if !has || accCharge == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 2 {
			return accCharge, nil
		}
		return accCharge.FieldAsInterface(fldPath[2:])
	case UnitFactorField:
		if len(fldPath) == 1 {
			return ec.UnitFactors, nil
		}
		if ec.UnitFactors == nil {
			return nil, ErrNotFound
		}
		unitFactor, has := ec.UnitFactors[fldPath[1]]
		if !has || unitFactor == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 2 {
			return unitFactor, nil
		}
		return unitFactor.FieldAsInterface(fldPath[2:])
	case Rating:
		if len(fldPath) == 1 {
			return ec.Rating, nil
		}
		if ec.Rating == nil {
			return nil, ErrNotFound
		}
		rtInterval, has := ec.Rating[fldPath[1]]
		if !has || rtInterval == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 2 {
			return rtInterval, nil
		}
		return rtInterval.FieldAsInterface(fldPath[2:])
	case RateField:
		if len(fldPath) == 1 {
			return ec.Rates, nil
		}
		if ec.Rates == nil {
			return nil, ErrNotFound
		}
		rate, has := ec.Rates[fldPath[1]]
		if !has || rate == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 2 {
			return rate, nil
		}
		return rate.FieldAsInterface(fldPath[2:])
	case AccountField:
		if len(fldPath) == 1 {
			return ec.Accounts, nil
		}
		if ec.Accounts == nil {
			return nil, ErrNotFound
		}
		acc, has := ec.Accounts[fldPath[1]]
		if !has || acc == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 2 {
			return acc, nil
		}
		return acc.FieldAsInterface(fldPath[2:])
	}
}

func (ec *EventCharges) FieldAsString(fldPath []string) (string, error) {
	val, err := ec.FieldAsInterface(fldPath)
	if err != nil {
		return "", nil
	}
	return IfaceAsString(val), nil
}

func (ec *EventCharges) getChargesForPath(fldPath []string, chr *ChargeEntry) (val any, err error) {
	if chr == nil {
		return nil, ErrNotFound
	}
	if len(fldPath) == 0 {
		return chr, nil
	}
	if fldPath[0] != Charging {
		return chr.FieldAsInterface(fldPath)
	}
	chargingType, chargingID, sepFound := strings.Cut(chr.ChargingID, InInFieldSep)
	if !sepFound {
		return nil, fmt.Errorf("expected ChargingID format '*accounting:*' or '*rating:*', got '%s'", chr.ChargingID)
	}

	// Handle different charging types.
	switch chargingType {
	case MetaRating:
		return ec.getRatingForPath(fldPath[1:], ec.Rating[chargingID])
	case MetaAccounting:
		return ec.getAccountingForPath(fldPath[1:], ec.Accounting[chargingID])
	}
	return nil, fmt.Errorf("unsupported field prefix: <%s>", fldPath[0])
}

func (ec *EventCharges) getAccountingForPath(fldPath []string, accCharge *AccountCharge) (val any, err error) {
	if accCharge == nil {
		return nil, ErrNotFound
	}
	if len(fldPath) == 0 {
		return accCharge, nil
	}

	switch fldPath[0] {
	case AccountField:
		acc := ec.Accounts[accCharge.AccountID]
		if acc == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 1 {
			return acc, nil
		}
		return acc.FieldAsInterface(fldPath[1:])
	case BalanceField:
		acc := ec.Accounts[accCharge.AccountID]
		if acc == nil {
			return nil, ErrNotFound
		}
		balance := acc.Balances[accCharge.BalanceID]
		if balance == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 1 {
			return balance, nil
		}
		return balance.FieldAsInterface(fldPath[1:])
	case UnitFactorField:
		unitFactor := ec.UnitFactors[accCharge.UnitFactorID]
		if unitFactor == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 1 {
			return unitFactor, nil
		}
		return unitFactor.FieldAsInterface(fldPath[1:])
	case Rating:
		return ec.getRatingForPath(fldPath[1:], ec.Rating[accCharge.RatingID])
	default: // JoinedCharge[0]
		opath, indx := GetPathIndex(fldPath[0])
		if opath != JoinedCharge {
			break
		}
		if indx == nil {
			return nil, fmt.Errorf("invalid index for '%s' field", JoinedCharge)
		}
		if len(accCharge.JoinedChargeIDs) <= *indx {
			return nil, ErrNotFound
		}
		return ec.getAccountingForPath(fldPath[1:], ec.Accounting[accCharge.JoinedChargeIDs[*indx]])
	}

	return accCharge.FieldAsInterface(fldPath)
}

func (ec *EventCharges) getRatingForPath(fldPath []string, rtInterval *RateSInterval) (val any, err error) {
	if rtInterval == nil {
		return nil, ErrNotFound
	}
	if len(fldPath) == 0 {
		return rtInterval, nil
	}

	opath, indx := GetPathIndex(fldPath[0])
	if opath != Increments || indx == nil {
		return rtInterval.FieldAsInterface(fldPath)
	}
	if len(rtInterval.Increments) <= *indx {
		return nil, ErrNotFound
	}
	incr := rtInterval.Increments[*indx]
	if len(fldPath) == 1 {
		return incr, nil
	}
	if incr == nil {
		return nil, ErrNotFound
	}
	if fldPath[1] == RateField {
		rate := ec.Rates[incr.RateID]
		if rate == nil {
			return nil, ErrNotFound
		}
		if len(fldPath) == 2 {
			return rate, nil
		}
		return rate.FieldAsInterface(fldPath[2:])
	}
	return incr.FieldAsInterface(fldPath[1:])
}
