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
)

// CGREventWithEeIDs is CGREvent with EventExporterIDs. This
// struct is used as an API argument. It has been moved into
// the engine package to avoid import cycling issues that were
// encountered when trying to properly handle unmarshalling
// for our EventCost type.
type CGREventWithEeIDs struct {
	EeIDs []string
	*CGREvent
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

// UnmarshalJSON ensures that JSON data is correctly decoded into a
// CGREventWithEeIDs while respecting the different unmarshalling logic
// required by the embedded CGREvent.
func (cgr *CGREventWithEeIDs) UnmarshalJSON(data []byte) error {

	// Define a temporary struct to capture only
	// the EeIDs field from the JSON data.
	var tempEeIDs struct {
		EeIDs []string
	}

	// Unmarshal JSON data into the temporary struct to extract EeIDs.
	// Will use the default unmarshaler.
	if err := json.Unmarshal(data, &tempEeIDs); err != nil {
		return err
	}

	// Assign the extracted EeIDs to the main struct's EeIDs field.
	cgr.EeIDs = tempEeIDs.EeIDs

	// Ensure the embedded CGREvent is initialized before attempting
	// to unmarshal into it. This is needed to avoid a nil pointer
	// dereference during the unmarshalling process.
	cgr.CGREvent = &CGREvent{}

	// Directly unmarshal the original JSON data into the embedded
	// CGREvent. Will be using CGREvent's UnmarshalJSON method.
	return json.Unmarshal(data, cgr.CGREvent)
}
