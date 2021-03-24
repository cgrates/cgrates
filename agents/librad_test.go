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

func TestRadReplyAppendAttributes(t *testing.T) {
	rply := radigo.NewPacket(radigo.AccessRequest, 2, dictRad, coder, "CGRateS.org").Reply()
	rplyFlds := []*config.FCTemplate{
		{Tag: "ReplyCode", Path: utils.MetaRep + utils.NestingSep + MetaRadReplyCode,
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.RadReply", utils.InfieldSep)},
		{Tag: "Acct-Session-Time", Path: utils.MetaRep + utils.NestingSep + "Acct-Session-Time",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.MaxUsage{*duration_seconds}", utils.InfieldSep)},
	}
	for _, v := range rplyFlds {
		v.ComputePath()
	}
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", nil, nil, nil)
	agReq.CGRReply.Set([]string{utils.CapMaxUsage}, utils.NewLeafNode(time.Hour))
	agReq.CGRReply.Set([]string{utils.CapAttributes, "RadReply"}, utils.NewLeafNode("AccessAccept"))
	agReq.CGRReply.Set([]string{utils.CapAttributes, utils.AccountField}, utils.NewLeafNode("1001"))

	if err := agReq.SetFields(rplyFlds); err != nil {
		t.Error(err)
	}
	if err := radReplyAppendAttributes(rply, agReq.Reply); err != nil {
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
