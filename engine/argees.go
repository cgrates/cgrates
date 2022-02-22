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

// CGREventWithEeIDs is the CGREventWithOpts with EventExporterIDs
type CGREventWithEeIDs struct {
	EeIDs []string
	*utils.CGREvent
}

func (cgr *CGREventWithEeIDs) UnmarshalJSON(data []byte) error {
	// firstly, we will unamrshall the entire data into raw bytes
	ids := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &ids); err != nil {
		return err
	}
	// populate eeids in case of it's existance
	eeIDs := make([]string, len(ids[utils.EeIDs]))
	if err := json.Unmarshal(ids[utils.EeIDs], &eeIDs); err != nil {
		return err
	}
	cgr.EeIDs = eeIDs
	// populate the entire CGRevent struct in case of it's existance
	var cgrEv *utils.CGREvent
	if err := json.Unmarshal(data, &cgrEv); err != nil {
		return err
	}
	cgr.CGREvent = cgrEv
	// check if we have EventCost and modify it's type (by default it was map[string]interface{} by unrmarshaling, now it will be EventCost)
	if ecEv, has := cgrEv.Event[utils.EventCost]; has {
		ec := new(EventCost)
		bts, err := json.Marshal(ecEv)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bts, &ec); err != nil {
			return err
		}
		cgr.Event[utils.EventCost] = ec
	}
	return nil
}
