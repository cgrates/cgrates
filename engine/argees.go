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

	"github.com/cgrates/cgrates/utils"
)

// CGREventWithEeIDs struct is moved in engine due to importing ciclying packages in order to unmarshalling properly for our EventCost type. This is the API struct argument

// CGREventWithEeIDs is CGREvent with EventExporterIDs
type CGREventWithEeIDs struct {
	EeIDs []string
	*utils.CGREvent
	clnb bool
}

func (attr *CGREventWithEeIDs) Clone() *CGREventWithEeIDs {
	return &CGREventWithEeIDs{
		EeIDs:    utils.CloneStringSlice(attr.EeIDs),
		CGREvent: attr.CGREvent.Clone(),
	}
}

// SetCloneable sets if the args should be cloned on internal connections
func (attr *CGREventWithEeIDs) SetCloneable(clnb bool) {
	attr.clnb = clnb
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *CGREventWithEeIDs) RPCClone() (interface{}, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

func (cgr *CGREventWithEeIDs) UnmarshalJSON(data []byte) (err error) {
	// firstly, we will unamrshall the entire data into raw bytes
	ids := make(map[string]json.RawMessage)
	if err = json.Unmarshal(data, &ids); err != nil {
		return
	}
	// populate eeids in case of it's existance
	eeIDs := make([]string, len(ids[utils.EeIDs]))
	if err = json.Unmarshal(ids[utils.EeIDs], &eeIDs); err != nil {
		return
	}
	cgr.EeIDs = eeIDs
	// populate the entire CGRevent struct in case of it's existance
	var cgrEv *utils.CGREvent
	if err = json.Unmarshal(data, &cgrEv); err != nil {
		return
	}
	cgr.CGREvent = cgrEv
	// check if we have CostDetails and modify it's type (by default it was map[string]interface{} by unrmarshaling, now it will be EventCost)
	if ecEv, has := cgrEv.Event[utils.CostDetails]; has {
		var bts []byte
		switch ecEv.(type) {
		case string:
			btsToStr, err := json.Marshal(ecEv)
			if err != nil {
				return err
			}
			var toString string
			if err = json.Unmarshal(btsToStr, &toString); err != nil {
				return err
			}
			bts = []byte(toString)
		default:
			bts, err = json.Marshal(ecEv)
			if err != nil {
				return err
			}
		}
		ec := new(EventCost)
		if err = json.Unmarshal(bts, &ec); err != nil {
			return err
		}
		cgr.Event[utils.CostDetails] = ec
	}
	return
}
