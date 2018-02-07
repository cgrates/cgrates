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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

// processorVars will hold various variables using during request processing
// here so we can define methods on it
type processorVars map[string]interface{}

// hasSubsystems will return true on single subsystem being present in processorVars
func (pv processorVars) hasSubsystems() (has bool) {
	for _, k := range []string{utils.MetaAccounts, utils.MetaResources,
		utils.MetaSuppliers, utils.MetaAttributes} {
		if _, has = pv[k]; has {
			return
		}
	}
	return
}

func (pv processorVars) hasVar(k string) (has bool) {
	_, has = pv[k]
	return
}

// valAsInterface returns the string value for fldName
func (pv processorVars) valAsInterface(fldPath string) (val interface{}, err error) {
	fldName := fldPath
	if strings.HasPrefix(fldPath, utils.MetaCGRReply) {
		fldName = utils.MetaCGRReply
	}
	if !pv.hasVar(fldName) {
		err = errors.New("not found")
		return
	}
	if fldName == utils.MetaCGRReply {
		cgrRply := pv[utils.MetaCGRReply].(utils.CGRReply)
		return cgrRply.GetField(fldPath, utils.HIERARCHY_SEP)
	}
	return pv[fldName], nil
}

// valAsString returns the string value for fldName
// returns empty if fldName not found
func (pv processorVars) valAsString(fldPath string) (val string, err error) {
	fldName := fldPath
	if strings.HasPrefix(fldPath, utils.MetaCGRReply) {
		fldName = utils.MetaCGRReply
	}
	if !pv.hasVar(fldName) {
		return "", errors.New("not found")
	}
	if fldName == utils.MetaCGRReply {
		cgrRply := pv[utils.MetaCGRReply].(utils.CGRReply)
		return cgrRply.GetFieldAsString(fldPath, utils.HIERARCHY_SEP)
	}
	if valIface, hasIt := pv[fldName]; hasIt {
		var canCast bool
		if val, canCast = utils.CastFieldIfToString(valIface); !canCast {
			return "", fmt.Errorf("cannot cast field <%s> to string", fldPath)
		}
	}
	return
}

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
func radPassesFieldFilter(pkt *radigo.Packet, processorVars processorVars,
	fieldFilter *utils.RSRField) (pass bool) {
	if fieldFilter == nil {
		return true
	}
	if valIface, hasIt := processorVars[fieldFilter.Id]; hasIt { // ProcessorVars have priority
		if val, canCast := utils.CastFieldIfToString(valIface); !canCast {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> cannot cast field <%s> to string",
					utils.RadiusAgent, fieldFilter.Id))
		} else if fieldFilter.FilterPasses(val) {
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
// procVars have priority over packet variables
func radComposedFieldValue(pkt *radigo.Packet,
	procVars processorVars, outTpl utils.RSRFields) (outVal string) {
	for _, rsrTpl := range outTpl {
		if rsrTpl.IsStatic() {
			outVal += rsrTpl.ParseValue("")
			continue
		}
		if val, err := procVars.valAsString(rsrTpl.Id); err != nil {
			if err.Error() != "not found" {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> %s",
						utils.RadiusAgent, err.Error()))
				continue
			}
		} else {
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
func radMetaHandler(pkt *radigo.Packet, procVars processorVars,
	cfgFld *config.CfgCdrField, roundingDecimals int) (outVal string, err error) {
	handlerArgs := strings.Split(
		radComposedFieldValue(pkt, procVars, cfgFld.Value), utils.HandlerArgSep)
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
	case utils.MetaUsageSeconds:
		if len(handlerArgs) != 1 {
			return "", errors.New("unexpected number of arguments")
		}
		val, err := utils.ParseDurationWithNanosecs(handlerArgs[0])
		if err != nil {
			return "", err
		}
		return strconv.FormatInt(int64(utils.Round(val.Seconds(),
			roundingDecimals, utils.ROUNDING_MIDDLE)), 10), nil
	}
	return
}

// radFieldOutVal formats the field value retrieved from RADIUS packet
func radFieldOutVal(pkt *radigo.Packet, processorVars processorVars,
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
		if outVal, err = radMetaHandler(pkt, processorVars, cfgFld,
			config.CgrConfig().RoundingDecimals); err != nil {
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
func radReqAsCGREvent(radPkt *radigo.Packet, procVars map[string]interface{}, procFlags utils.StringMap,
	cfgFlds []*config.CfgCdrField) (cgrEv *utils.CGREvent, err error) {
	outMap := make(map[string]string) // work with it so we can append values to keys
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
	cgrEv = &utils.CGREvent{
		Tenant: utils.FirstNonEmpty(outMap[utils.Tenant],
			config.CgrConfig().DefaultTenant),
		ID:    utils.UUIDSha1Prefix(),
		Time:  utils.TimePointer(time.Now()),
		Event: utils.ConvertMapValStrIf(outMap),
	}
	return
}

// radReplyAppendAttributes appends attributes to a RADIUS reply based on predefined template
func radReplyAppendAttributes(reply *radigo.Packet, procVars map[string]interface{},
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

// radV1AuthorizeArgs returns the arguments needed by SessionSv1.AuthorizeEvent
func radV1AuthorizeArgs(cgrEv *utils.CGREvent, procVars processorVars) (args *sessions.V1AuthorizeArgs) {
	args = &sessions.V1AuthorizeArgs{ // defaults
		GetMaxUsage: true,
		CGREvent:    *cgrEv,
	}
	if !procVars.hasSubsystems() {
		return
	}
	if !procVars.hasVar(utils.MetaAccounts) {
		args.GetMaxUsage = false
	}
	if procVars.hasVar(utils.MetaResources) {
		args.AuthorizeResources = true
	}
	if procVars.hasVar(utils.MetaSuppliers) {
		args.GetSuppliers = true
	}
	if procVars.hasVar(utils.MetaAttributes) {
		args.GetAttributes = true
	}
	return
}

// radV1InitSessionArgs returns the arguments used in SessionSv1.InitSession
func radV1InitSessionArgs(cgrEv *utils.CGREvent, procVars processorVars) (args *sessions.V1InitSessionArgs) {
	args = &sessions.V1InitSessionArgs{ // defaults
		InitSession: true,
		CGREvent:    *cgrEv,
	}
	if !procVars.hasSubsystems() {
		return
	}
	if !procVars.hasVar(utils.MetaAccounts) {
		args.InitSession = false
	}
	if procVars.hasVar(utils.MetaResources) {
		args.AllocateResources = true
	}
	if procVars.hasVar(utils.MetaAttributes) {
		args.GetAttributes = true
	}
	return
}

// radV1InitSessionArgs returns the arguments used in SessionSv1.InitSession
func radV1UpdateSessionArgs(cgrEv *utils.CGREvent, procVars processorVars) (args *sessions.V1UpdateSessionArgs) {
	args = &sessions.V1UpdateSessionArgs{ // defaults
		UpdateSession: true,
		CGREvent:      *cgrEv,
	}
	if !procVars.hasSubsystems() {
		return
	}
	if !procVars.hasVar(utils.MetaAccounts) {
		args.UpdateSession = false
	}
	if procVars.hasVar(utils.MetaAttributes) {
		args.GetAttributes = true
	}
	return
}

// radV1TerminateSessionArgs returns the arguments used in SMGv1.TerminateSession
func radV1TerminateSessionArgs(cgrEv *utils.CGREvent, procVars processorVars) (args *sessions.V1TerminateSessionArgs) {
	args = &sessions.V1TerminateSessionArgs{ // defaults
		TerminateSession: true,
		CGREvent:         *cgrEv,
	}
	if !procVars.hasSubsystems() {
		return
	}
	if !procVars.hasVar(utils.MetaAccounts) {
		args.TerminateSession = false
	}
	if procVars.hasVar(utils.MetaResources) {
		args.ReleaseResources = true
	}
	return
}
