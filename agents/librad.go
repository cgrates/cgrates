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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

// radAttrVendorFromPath returns AttributenName and VendorName from path
// path should be the form attributeName or vendorName/attributeName
func attrVendorFromPath(path string) (attrName, vendorName string) {
	splt := strings.Split(path, utils.HIERARCHY_SEP)
	if len(splt) > 1 {
		vendorName, attrName = splt[0], splt[1]
	} else {
		attrName = splt[0]
	}
	return
}

// radComposedFieldValue extracts the field value out of RADIUS packet
// procVars have priority over packet variables
func radComposedFieldValue(pkt *radigo.Packet,
	procVars processorVars, outTpl utils.RSRFields) (outVal string) {
	for _, rsrTpl := range outTpl {
		if rsrTpl.IsStatic() {
			if parsed, err := rsrTpl.Parse(""); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> %s",
						utils.RadiusAgent, err.Error()))
			} else {
				outVal += parsed
			}
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
			if parsed, err := rsrTpl.Parse(val); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> %s",
						utils.RadiusAgent, err.Error()))
			} else {
				outVal += parsed
			}
			continue
		}
		for _, avp := range pkt.AttributesWithName(
			attrVendorFromPath(rsrTpl.Id)) {
			if parsed, err := rsrTpl.Parse(avp.GetStringValue()); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> %s",
						utils.RadiusAgent, err.Error()))
			} else {
				outVal += parsed
			}
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
	case utils.MetaDurationSeconds:
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

// NewCGRReply is specific to replies coming from CGRateS
func NewCGRReply(rply engine.NavigableMapper,
	errRply error) (mp *engine.NavigableMap, err error) {
	if errRply != nil {
		return engine.NewNavigableMap(map[string]interface{}{
			utils.Error: errRply.Error()}), nil
	}
	mp, err = rply.AsNavigableMap(nil)
	if err != nil {
		return nil, err
	}
	mp.Set([]string{utils.Error}, "", false) // enforce empty error
	return mp, nil
}

// newRADataProvider constructs a DataProvider
func newRADataProvider(req *radigo.Packet) (dP engine.DataProvider, err error) {
	dP = &radiusDP{req: req, cache: engine.NewNavigableMap(nil)}
	return
}

// radiusDP implements engine.DataProvider, serving as radigo.Packet data decoder
// decoded data is only searched once and cached
type radiusDP struct {
	req   *radigo.Packet
	cache *engine.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (pk *radiusDP) String() string {
	return ""
}

// FieldAsInterface is part of engine.DataProvider interface
func (pk *radiusDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	if data, err = pk.cache.FieldAsInterface(fldPath); err == nil ||
		err != utils.ErrNotFound { // item found in cache
		return
	}
	err = nil // cancel previous err
	pk.cache.Set(fldPath, data, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (pk *radiusDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = pk.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, _ = utils.CastFieldIfToString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (pk *radiusDP) AsNavigableMap([]*config.CfgCdrField) (
	nm *engine.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}
