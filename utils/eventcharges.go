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

	"github.com/ericlagergren/decimal"
)

// EventCharges records the charges applied to an Event
type EventCharges struct {
	Usage      *decimal.Big
	Cost       *decimal.Big
	Charges    []*ChargedInterval
	Account    *AccountProfile
	Accounting *ChargedAccounting
	Rating     *ChargedRating
}

// Merge will merge the event charges into existing
func (ec *EventCharges) Merge(eCs ...*EventCharges) {
	for _, nEc := range eCs {
		if ec.Usage == nil {
			ec.Usage = nEc.Usage
			continue
		}
		ec.Usage = SumBig(ec.Usage, nEc.Usage)
	}
}

// AsExtEventCharges converts EventCharges to ExtEventCharges
func (ec *EventCharges) AsExtEventCharges() (eEc *ExtEventCharges, err error) {
	eEc = new(ExtEventCharges)
	if ec.Usage != nil {
		if flt, ok := ec.Usage.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Usage to float64")
		} else {
			eEc.Usage = &flt
		}
	}
	if ec.Cost != nil {
		if flt, ok := ec.Cost.Float64(); !ok {
			return nil, errors.New("cannot convert decimal Cost to float64")
		} else {
			eEc.Cost = &flt
		}
	}
	// add here code for the rest of the fields
	return
}

// ExtEventCharges is a generic EventCharges used in APIs
type ExtEventCharges struct {
	Usage *float64
	Cost  *float64
}
