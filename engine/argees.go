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

package engine

import (
	"encoding/json"
	"slices"

	"github.com/cgrates/cgrates/utils"
)

// CGREventWithEeIDs is CGREvent with EventExporterIDs. This
// struct is used as an API argument. It has been moved into
// the engine package to avoid import cycling issues that were
// encountered when trying to properly handle unmarshalling
// for our EventCost type.
type CGREventWithEeIDs struct {
	EeIDs []string
	*utils.CGREvent
	clnb bool
}

func (attr *CGREventWithEeIDs) Clone() *CGREventWithEeIDs {
	return &CGREventWithEeIDs{
		EeIDs:    slices.Clone(attr.EeIDs),
		CGREvent: attr.CGREvent.Clone(),
	}
}

// SetCloneable sets if the args should be cloned on internal connections
func (attr *CGREventWithEeIDs) SetCloneable(clnb bool) {
	attr.clnb = clnb
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *CGREventWithEeIDs) RPCClone() (any, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// UnmarshalJSON decodes the JSON data into a CGREventWithEeIDs, while
// ensuring that the CostDetails key of the embedded CGREvent is of
// type *engine.EventCost.
func (cgr *CGREventWithEeIDs) UnmarshalJSON(data []byte) error {

	// Define a temporary struct with the same
	// structure as CGREventWithEeIDs.
	var temp struct {
		EeIDs []string
		*utils.CGREvent
	}

	// Unmarshal JSON data into the temporary struct,
	// using the default unmarshaler.
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	if ecEv, has := temp.Event[utils.CostDetails]; has {
		var ecBytes []byte

		// CostDetails value can either be a JSON string (which is
		// the marshaled form of an EventCost) or a map representing
		// the EventCost directly.
		switch v := ecEv.(type) {
		case string:
			// If string, it's assumed to be the JSON
			// representation of EventCost.
			ecBytes = []byte(v)
		default:
			// Otherwise we assume it's a map and we marshal
			// it back to JSON to prepare for unmarshalling
			// into EventCost.
			ecBytes, err = json.Marshal(v)
			if err != nil {
				return err
			}
		}

		// Unmarshal the JSON (either directly from the string case
		// or from the marshaled map) into an EventCost struct.
		var ec EventCost
		if err := json.Unmarshal(ecBytes, &ec); err != nil {
			return err
		}

		// Update the Event map with the unmarshalled EventCost,
		// ensuring the type of CostDetails is *EventCost.
		temp.Event[utils.CostDetails] = &ec
	}
	isAccountUpdate := temp.Event[utils.EventType] == utils.AccountUpdate
	if accEv, has := temp.Event[utils.AccountField]; has && isAccountUpdate {
		accBytes, err := json.Marshal(accEv)
		if err != nil {
			return err
		}
		var as Account
		if err = json.Unmarshal(accBytes, &as); err != nil {
			return err
		}
		temp.Event[utils.AccountField] = &as
	}
	// Assign the extracted EeIDs and CGREvent
	// to the main struct fields.
	cgr.EeIDs = temp.EeIDs
	cgr.CGREvent = temp.CGREvent

	return nil
}
