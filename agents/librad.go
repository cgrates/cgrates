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

package agents

import (
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

func radPassesFieldFilter(pkt *radigo.Packet, fieldFilter *utils.RSRField, processorVars map[string]string) (pass bool) {
	if fieldFilter == nil {
		return true
	}
	if val, hasIt := processorVars[fieldFilter.Id]; hasIt { // ProcessorVars have priority
		if fieldFilter.FilterPasses(val) {
			pass = true
		}
		return
	}
	splt := strings.Split(fieldFilter.Id, "/")
	var attrName, vendorName string
	if len(splt) > 1 {
		vendorName, attrName = splt[0], splt[1]
	} else {
		attrName = splt[0]
	}
	avps := pkt.AttributesWithName(attrName, vendorName)
	if len(avps) == 0 { // no attribute found, filter not passing
		return
	}
	for _, avp := range avps { // they all need to match the filter
		if !fieldFilter.FilterPasses(avp.StringValue()) {
			return
		}
	}
	return true
}

// radPktAsSMGEvent converts a RADIUS packet into SMGEvent
func radReqAsSMGEvent(radPkt *radigo.Packet, procVars map[string]string,
	tplFlds []*config.CfgCdrField, procFlags utils.StringMap) (smgEv sessionmanager.SMGenericEvent, err error) {
	return
}

// radReplyAppendAttributes appends attributes to a RADIUS reply based on predefined template
func radReplyAppendAttributes(reply *radigo.Packet, procVars map[string]string,
	tplFlds []*config.CfgCdrField, procFlags utils.StringMap) (err error) {
	return
}
