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

// ExtEventCharges is a generic EventCharges used in APIs
type ExtEventCharges struct {
	Abstracts *float64
	Concretes *float64
}

type ChargingInterval struct {
	Increments     []*ChargingIncrement // specific increments applied to this interval
	CompressFactor int
}

// ChargingIncrement represents one unit charged inside an interval
type ChargingIncrement struct {
	Units           *Decimal
	AccountChargeID string // Account charging information
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
