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
	"errors"
	"fmt"
	"strings"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

// radAttrVendorFromPath returns AttributenName and VendorName from path
// path should be the form attributeName or vendorName/attributeName
func attrVendorFromPath(path string) (attrName, vendorName string) {
	splt := strings.Split(path, "/")
	if len(splt) > 1 {
		vendorName, attrName = splt[0], splt[1]
	} else {
		attrName = splt[0]
	}
	return
}

// radPassesFieldFilter checks whether fieldFilter matches either in processorsVars or AVPs of packet
func radPassesFieldFilter(pkt *radigo.Packet, processorVars map[string]string, fieldFilter *utils.RSRField) (pass bool) {
	if fieldFilter == nil {
		return true
	}
	if val, hasIt := processorVars[fieldFilter.Id]; hasIt { // ProcessorVars have priority
		if fieldFilter.FilterPasses(val) {
			pass = true
		}
		return
	}
	avps := pkt.AttributesWithName(attrVendorFromPath(fieldFilter.Id))
	if len(avps) == 0 { // no attribute found, filter not passing
		return
	}
	for _, avp := range avps { // they all need to match the filter
		if !fieldFilter.FilterPasses(avp.GetStringValue()) {
			return
		}
	}
	return true
}

// radComposedFieldValue extracts the field value out of RADIUS packet
func radComposedFieldValue(pkt *radigo.Packet,
	processorVars map[string]string, outTpl utils.RSRFields) (outVal string) {
	for _, rsrTpl := range outTpl {
		if rsrTpl.IsStatic() {
			outVal += rsrTpl.ParseValue("")
			continue
		}
		if val, hasIt := processorVars[rsrTpl.Id]; hasIt { // ProcessorVars have priority
			outVal += rsrTpl.ParseValue(val)
			continue
		}
		for _, avp := range pkt.AttributesWithName(
			attrVendorFromPath(rsrTpl.Id)) {
			outVal += rsrTpl.ParseValue(avp.GetStringValue())
		}
	}
	return outVal
}

// radMetaHandler handles *handler type in configuration fields
func radMetaHandler(pkt *radigo.Packet, processorVars map[string]string,
	cfgFld *config.CfgCdrField) (outVal string, err error) {
	handlerArgs := strings.Split(
		radComposedFieldValue(pkt, processorVars, cfgFld.Value), utils.HandlerArgSep)
	switch cfgFld.HandlerId {
	case MetaUsageDifference: // expects tEnd|tStart in the composed val
		if len(handlerArgs) != 2 {
			return "", errors.New("unexpected number of arguments")
		}
		tEnd, err := utils.ParseTimeDetectLayout(handlerArgs[0], cfgFld.Timezone)
		if err != nil {
			return "", err
		}
		tStart, err := utils.ParseTimeDetectLayout(handlerArgs[1], cfgFld.Timezone)
		if err != nil {
			return "", err
		}
		return tEnd.Sub(tStart).String(), nil
	}
	return
}

// radFieldOutVal formats the field value retrieved from RADIUS packet
func radFieldOutVal(pkt *radigo.Packet, processorVars map[string]string,
	cfgFld *config.CfgCdrField) (outVal string, err error) {
	// different output based on cgrFld.Type
	switch cfgFld.Type {
	case utils.META_FILLER:
		outVal = cfgFld.Value.Id()
		cfgFld.Padding = "right"
	case utils.META_CONSTANT:
		outVal = cfgFld.Value.Id()
	case utils.META_COMPOSED:
		outVal = radComposedFieldValue(pkt, processorVars, cfgFld.Value)
	case utils.META_HANDLER:
		if outVal, err = radMetaHandler(pkt, processorVars, cfgFld); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported configuration field type: <%s>", cfgFld.Type)
	}
	if outVal, err = utils.FmtFieldWidth(cfgFld.Tag, outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
		return "", err
	}
	return
}

// radPktAsSMGEvent converts a RADIUS packet into SMGEvent
func radReqAsSMGEvent(radPkt *radigo.Packet, procVars map[string]string, procFlags utils.StringMap,
	cfgFlds []*config.CfgCdrField) (smgEv sessionmanager.SMGenericEvent, err error) {
	outMap := make(map[string]string) // work with it so we can append values to keys
	outMap[utils.EVENT_NAME] = EvRadiusReq
	for _, cfgFld := range cfgFlds {
		passedAllFilters := true
		for _, fldFilter := range cfgFld.FieldFilter {
			if !radPassesFieldFilter(radPkt, procVars, fldFilter) {
				passedAllFilters = false
				break
			}
		}
		if !passedAllFilters {
			continue
		}
		fmtOut, err := radFieldOutVal(radPkt, procVars, cfgFld)
		if err != nil {
			return nil, err
		}
		if _, hasKey := outMap[cfgFld.FieldId]; hasKey && cfgFld.Append {
			outMap[cfgFld.FieldId] += fmtOut
		} else {
			outMap[cfgFld.FieldId] = fmtOut
		}
		if cfgFld.BreakOnSuccess {
			break
		}
	}
	if len(procFlags) != 0 {
		outMap[utils.CGRFlags] = procFlags.String()
	}
	return sessionmanager.SMGenericEvent(utils.ConvertMapValStrIf(outMap)), nil
}

// radReplyAppendAttributes appends attributes to a RADIUS reply based on predefined template
func radReplyAppendAttributes(reply *radigo.Packet, procVars map[string]string,
	cfgFlds []*config.CfgCdrField) (err error) {
	for _, cfgFld := range cfgFlds {
		passedAllFilters := true
		for _, fldFilter := range cfgFld.FieldFilter {
			if !radPassesFieldFilter(reply, procVars, fldFilter) {
				passedAllFilters = false
				break
			}
		}
		if !passedAllFilters {
			continue
		}
		fmtOut, err := radFieldOutVal(reply, procVars, cfgFld)
		if err != nil {
			return err
		}
		if cfgFld.FieldId == MetaRadReplyCode { // Special case used to control the reply code of RADIUS reply
			if err = reply.SetCodeWithName(fmtOut); err != nil {
				return err
			}
			continue
		}
		attrName, vendorName := attrVendorFromPath(cfgFld.FieldId)
		if err = reply.AddAVPWithName(attrName, fmtOut, vendorName); err != nil {
			return err
		}
		if cfgFld.BreakOnSuccess {
			break
		}
	}
	return
}
