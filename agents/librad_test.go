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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/radigo"
)

var (
	dictRad *radigo.Dictionary
	coder   radigo.Coder
)

var freeRADIUSDocDictSample = `
# Most of the lines are copied from freeradius documentation here:
# http://networkradius.com/doc/3.0.10/concepts/dictionary/introduction.html

# Attributes
ATTRIBUTE    User-Name    1    string
ATTRIBUTE    Password     2    string

# Alias values
VALUE    Framed-Protocol    PPP    1
VALUE Service-Type       Sip-Session      15   # Schulzrinne, acc, auth_radius
VALUE Service-Type       SIP-Caller-AVPs  30   # Proprietary, avp_radius

VALUE Sip-Method         Invite         1
VALUE Sip-Method         Bye            8
VALUE Acct-Status-Type	Start			1
VALUE Acct-Status-Type	Stop			2

# Vendors
VENDOR    Cisco    9
VENDOR    Microsoft 311

# Vendor AVPs
BEGIN-VENDOR    Cisco
ATTRIBUTE       Cisco-AVPair    1   string
ATTRIBUTE       Cisco-NAS-Port  2	string
END-VENDOR      Cisco

ATTRIBUTE	Sip-Method		101	integer
ATTRIBUTE	Sip-Response-Code	102	integer
ATTRIBUTE	Sip-From-Tag		105	string
ATTRIBUTE	Sip-To-Tag		104	string
ATTRIBUTE	Ascend-User-Acct-Time		143	integer

`

func init() {
	dictRad = radigo.RFC2865Dictionary()
	dictRad.ParseFromReader(strings.NewReader(freeRADIUSDocDictSample))
	coder = radigo.NewCoder()
}

func TestAttrVendorFromPath(t *testing.T) {
	if attrName, vendorName := attrVendorFromPath("User-Name"); attrName != "User-Name" ||
		vendorName != "" {
		t.Error("failed")
	}
	if attrName, vendorName := attrVendorFromPath("Cisco>Cisco-NAS-Port"); attrName != "Cisco-NAS-Port" ||
		vendorName != "Cisco" {
		t.Error("failed")
	}
}

func TestRadComposedFieldValue(t *testing.T) {
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	if err := pkt.AddAVPWithName("User-Name", "flopsy", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Cisco-NAS-Port", "CGR1", "Cisco"); err != nil {
		t.Error(err)
	}
	agReq := newAgentRequest(nil, nil, nil, nil, "cgrates.org", "", nil)
	agReq.Vars.Set([]string{MetaRadReqType}, MetaRadAcctStart, false, false)
	agReq.Vars.Set([]string{"Cisco"}, "CGR1", false, false)
	agReq.Vars.Set([]string{"User-Name"}, "flopsy", false, false)
	eOut := "*radAcctStart|flopsy|CGR1"
	if out := radComposedFieldValue(pkt, agReq,
		config.NewRSRParsersMustCompile("~*vars.*radReqType;|;~*vars.User-Name;|;~*vars.Cisco", true, utils.INFIELD_SEP)); out != eOut {
		t.Errorf("Expecting: <%s>, received: <%s>", eOut, out)
	}
}

func TestRadFieldOutVal(t *testing.T) {
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	if err := pkt.AddAVPWithName("User-Name", "flopsy", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Cisco-NAS-Port", "CGR1", "Cisco"); err != nil {
		t.Error(err)
	}
	eOut := fmt.Sprintf("%s|flopsy|CGR1", MetaRadAcctStart)
	agReq := newAgentRequest(nil, nil, nil, nil, "cgrates.org", "", nil)
	agReq.Vars.Set([]string{MetaRadReqType}, MetaRadAcctStart, false, false)
	agReq.Vars.Set([]string{"Cisco"}, "CGR1", false, false)
	agReq.Vars.Set([]string{"User-Name"}, "flopsy", false, false)
	//processorVars{MetaRadReqType: MetaRadAcctStart}
	cfgFld := &config.FCTemplate{Tag: "ComposedTest", Type: utils.META_COMPOSED, FieldId: utils.Destination,
		Value: config.NewRSRParsersMustCompile("~*vars.*radReqType;|;~*vars.User-Name;|;~*vars.Cisco", true, utils.INFIELD_SEP), Mandatory: true}
	if outVal, err := radFieldOutVal(pkt, agReq, cfgFld); err != nil {
		t.Error(err)
	} else if outVal != eOut {
		t.Errorf("Expecting: <%s>, received: <%s>", eOut, outVal)
	}
}

func TestRadReplyAppendAttributes(t *testing.T) {
	rply := radigo.NewPacket(radigo.AccessRequest, 2, dictRad, coder, "CGRateS.org").Reply()
	rplyFlds := []*config.FCTemplate{
		&config.FCTemplate{Tag: "ReplyCode", FieldId: MetaRadReplyCode, Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.RadReply", true, utils.INFIELD_SEP)},
		&config.FCTemplate{Tag: "Acct-Session-Time", FieldId: "Acct-Session-Time", Type: utils.META_COMPOSED,
			Value: config.NewRSRParsersMustCompile("~*cgrep.MaxUsage{*duration_seconds}", true, utils.INFIELD_SEP)},
	}
	agReq := newAgentRequest(nil, nil, nil, nil, "cgrates.org", "", nil)
	agReq.CGRReply.Set([]string{utils.CapMaxUsage}, time.Duration(time.Hour), false, false)
	agReq.CGRReply.Set([]string{utils.CapAttributes, "RadReply"}, "AccessAccept", false, false)
	agReq.CGRReply.Set([]string{utils.CapAttributes, utils.Account}, "1001", false, false)
	if err := radReplyAppendAttributes(rply, agReq, rplyFlds); err != nil {
		t.Error(err)
	}
	if rply.Code != radigo.AccessAccept {
		t.Errorf("Wrong reply code: %d", rply.Code)
	}
	if avps := rply.AttributesWithName("Acct-Session-Time", ""); len(avps) == 0 {
		t.Error("Cannot find Acct-Session-Time in reply")
	} else if avps[0].GetStringValue() != "3600" {
		t.Errorf("Expecting: 3600, received: %s", avps[0].GetStringValue())
	}
}

type myEv map[string]interface{}

func (ev myEv) AsNavigableMap(tpl []*config.CfgCdrField) (*config.NavigableMap, error) {
	return config.NewNavigableMap(ev), nil
}

func TestNewCGRReply(t *testing.T) {
	eCgrRply := config.NewNavigableMap(map[string]interface{}{
		utils.Error: "some",
	})
	if rpl, err := NewCGRReply(nil, errors.New("some")); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrRply, rpl) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(eCgrRply), utils.ToJSON(rpl))
	}
	ev := myEv{
		"FirstLevel": map[string]interface{}{
			"SecondLevel": map[string]interface{}{
				"Fld1": "Val1",
			},
		},
	}
	eCgrRply = config.NewNavigableMap(ev)
	eCgrRply.Set([]string{utils.Error}, "", false, false)
	if rpl, err := NewCGRReply(config.NavigableMapper(ev), nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eCgrRply, rpl) {
		t.Errorf("Expecting: %+v, received: %+v", eCgrRply, rpl)
	}
}

func TestRadiusDPFieldAsInterface(t *testing.T) {
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	if err := pkt.AddAVPWithName("User-Name", "flopsy", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Cisco-NAS-Port", "CGR1", "Cisco"); err != nil {
		t.Error(err)
	}
	dp := newRADataProvider(pkt)
	if data, err := dp.FieldAsInterface([]string{"User-Name"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, utils.StringToInterface("flopsy")) {
		t.Errorf("Expecting: <%s>, received: <%s>", utils.StringToInterface("flopsy"), data)
	}
}

func TestRadiusDPFieldAsString(t *testing.T) {
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	if err := pkt.AddAVPWithName("User-Name", "flopsy", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Cisco-NAS-Port", "CGR1", "Cisco"); err != nil {
		t.Error(err)
	}
	dp := newRADataProvider(pkt)
	if data, err := dp.FieldAsString([]string{"User-Name"}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(data, "flopsy") {
		t.Errorf("Expecting: flopsy, received: <%s>", data)
	}
}
