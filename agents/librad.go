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
	"fmt"
	"net"
	"strings"

	"github.com/cgrates/cgrates/config"
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
	agReq *AgentRequest, outTpl config.RSRParsers) (outVal string) {
	for _, rsrTpl := range outTpl {
		if out, err := rsrTpl.ParseDataProvider(agReq, utils.NestingSep); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> %s",
					utils.RadiusAgent, err.Error()))
			continue
		} else {
			outVal += out
			continue
		}
		for _, avp := range pkt.AttributesWithName(
			attrVendorFromPath(rsrTpl.Rules)) {
			if parsed, err := rsrTpl.ParseValue(avp.GetStringValue()); err != nil {
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

// radFieldOutVal formats the field value retrieved from RADIUS packet
func radFieldOutVal(pkt *radigo.Packet, agReq *AgentRequest,
	cfgFld *config.FCTemplate) (outVal string, err error) {
	// different output based on cgrFld.Type
	switch cfgFld.Type {
	case utils.META_FILLER:
		outVal, err = cfgFld.Value.ParseValue(utils.EmptyString)
		cfgFld.Padding = "right"
	case utils.META_CONSTANT:
		outVal, err = cfgFld.Value.ParseValue(utils.EmptyString)
	case utils.META_COMPOSED:
		outVal = radComposedFieldValue(pkt, agReq, cfgFld.Value)
	default:
		return "", fmt.Errorf("unsupported configuration field type: <%s>", cfgFld.Type)
	}
	if err != nil {
		return
	}
	if outVal, err = utils.FmtFieldWidth(cfgFld.Tag, outVal, cfgFld.Width, cfgFld.Strip, cfgFld.Padding, cfgFld.Mandatory); err != nil {
		return "", err
	}
	return
}

// radReplyAppendAttributes appends attributes to a RADIUS reply based on predefined template
func radReplyAppendAttributes(reply *radigo.Packet, agReq *AgentRequest,
	cfgFlds []*config.FCTemplate) (err error) {
	for _, cfgFld := range cfgFlds {
		fmtOut, err := radFieldOutVal(reply, agReq, cfgFld)
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
func NewCGRReply(rply config.NavigableMapper,
	errRply error) (mp *config.NavigableMap, err error) {
	if errRply != nil {
		return config.NewNavigableMap(map[string]interface{}{
			utils.Error: errRply.Error()}), nil
	}
	mp = config.NewNavigableMap(nil)
	if rply != nil {
		mp, err = rply.AsNavigableMap(nil)
		if err != nil {
			return nil, err
		}
	}
	mp.Set([]string{utils.Error}, "", false, false) // enforce empty error
	return
}

// newRADataProvider constructs a DataProvider
func newRADataProvider(req *radigo.Packet) (dP config.DataProvider) {
	dP = &radiusDP{req: req, cache: config.NewNavigableMap(nil)}
	return
}

// radiusDP implements engine.DataProvider, serving as radigo.Packet data decoder
// decoded data is only searched once and cached
type radiusDP struct {
	req   *radigo.Packet
	cache *config.NavigableMap
}

// String is part of engine.DataProvider interface
// when called, it will display the already parsed values out of cache
func (pk *radiusDP) String() string {
	return utils.ToIJSON(pk.req) // return ToJSON because Packet don't have a string method
}

// FieldAsInterface is part of engine.DataProvider interface
func (pk *radiusDP) FieldAsInterface(fldPath []string) (data interface{}, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	if data, err = pk.cache.FieldAsInterface(fldPath); err != nil {
		if err != utils.ErrNotFound { // item found in cache
			return
		}
		err = nil // cancel previous err
	} else {
		return // data found in cache
	}
	if len(pk.req.AttributesWithName(fldPath[0], "")) != 0 {
		data = pk.req.AttributesWithName(fldPath[0], "")[0].GetStringValue()
	}
	pk.cache.Set(fldPath, data, false, false)
	return
}

// FieldAsString is part of engine.DataProvider interface
func (pk *radiusDP) FieldAsString(fldPath []string) (data string, err error) {
	var valIface interface{}
	valIface, err = pk.FieldAsInterface(fldPath)
	if err != nil {
		return
	}
	data, err = utils.IfaceAsString(valIface)
	return
}

// AsNavigableMap is part of engine.DataProvider interface
func (pk *radiusDP) AsNavigableMap([]*config.FCTemplate) (
	nm *config.NavigableMap, err error) {
	return nil, utils.ErrNotImplemented
}

// RemoteHost is part of engine.DataProvider interface
func (pk *radiusDP) RemoteHost() net.Addr {
	return utils.NewNetAddr(pk.req.RemoteAddr().Network(), pk.req.RemoteAddr().String())
}
