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
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/sessions"
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
	if attrName, vendorName := attrVendorFromPath("Cisco/Cisco-NAS-Port"); attrName != "Cisco-NAS-Port" ||
		vendorName != "Cisco" {
		t.Error("failed")
	}
}

func TestRadPassesFieldFilter(t *testing.T) {
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	if err := pkt.AddAVPWithName("User-Name", "flopsy", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Cisco-NAS-Port", "CGR1", "Cisco"); err != nil {
		t.Error(err)
	}
	if !radPassesFieldFilter(pkt, nil, nil) {
		t.Error("not passing empty filter")
	}
	if !radPassesFieldFilter(pkt, nil,
		utils.NewRSRFieldMustCompile("User-Name(flopsy)")) {
		t.Error("not passing valid filter")
	}
	if radPassesFieldFilter(pkt, nil,
		utils.NewRSRFieldMustCompile("User-Name(notmatching)")) {
		t.Error("passing invalid filter value")
	}
	if !radPassesFieldFilter(pkt, nil,
		utils.NewRSRFieldMustCompile("Cisco/Cisco-NAS-Port(CGR1)")) {
		t.Error("not passing valid filter")
	}
	if radPassesFieldFilter(pkt, nil,
		utils.NewRSRFieldMustCompile("Cisco/Cisco-NAS-Port(notmatching)")) {
		t.Error("passing invalid filter value")
	}
	if !radPassesFieldFilter(pkt, processorVars{MetaRadReqType: MetaRadAuth},
		utils.NewRSRFieldMustCompile(fmt.Sprintf("%s(%s)", MetaRadReqType, MetaRadAuth))) {
		t.Error("not passing valid filter")
	}
	if radPassesFieldFilter(pkt, processorVars{MetaRadReqType: MetaRadAcctStart},
		utils.NewRSRFieldMustCompile(fmt.Sprintf("%s(%s)", MetaRadReqType, MetaRadAuth))) {
		t.Error("passing invalid filter")
	}
	if radPassesFieldFilter(pkt, nil,
		utils.NewRSRFieldMustCompile("UnknownField(notmatching)")) {
		t.Error("passing invalid filter value")
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
	eOut := fmt.Sprintf("%s|flopsy|CGR1", MetaRadAcctStart)
	if out := radComposedFieldValue(pkt, processorVars{MetaRadReqType: MetaRadAcctStart},
		utils.ParseRSRFieldsMustCompile(fmt.Sprintf("%s;^|;User-Name;^|;Cisco/Cisco-NAS-Port",
			MetaRadReqType), utils.INFIELD_SEP)); out != eOut {
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
	cfgFld := &config.CfgCdrField{Tag: "ComposedTest", Type: utils.META_COMPOSED, FieldId: utils.Destination,
		Value: utils.ParseRSRFieldsMustCompile(fmt.Sprintf("%s;^|;User-Name;^|;Cisco/Cisco-NAS-Port", MetaRadReqType), utils.INFIELD_SEP), Mandatory: true}
	if outVal, err := radFieldOutVal(pkt, processorVars{MetaRadReqType: MetaRadAcctStart}, cfgFld); err != nil {
		t.Error(err)
	} else if outVal != eOut {
		t.Errorf("Expecting: <%s>, received: <%s>", eOut, outVal)
	}
}

func TestRadReqAsCGREvent(t *testing.T) {
	pkt := radigo.NewPacket(radigo.AccountingRequest, 1, dictRad, coder, "CGRateS.org")
	// Sample minimal packet sent by Kamailio
	if err := pkt.AddAVPWithName("Acct-Status-Type", "2", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Service-Type", "15", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Sip-Response-Code", "200", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Sip-Method", "8", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Event-Timestamp", "1497106119", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Sip-From-Tag", "75c2f57b", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Sip-To-Tag", "51585361", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Acct-Session-Id", "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("User-Name", "1001", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Called-Station-Id", "1002", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Ascend-User-Acct-Time", "1497106115", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("NAS-Port-Id", "5060", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("Acct-Delay-Time", "0", ""); err != nil {
		t.Error(err)
	}
	if err := pkt.AddAVPWithName("NAS-IP-Address", "127.0.0.1", ""); err != nil {
		t.Error(err)
	}

	cfgFlds := []*config.CfgCdrField{
		&config.CfgCdrField{Tag: "TOR", FieldId: utils.TOR, Type: utils.META_CONSTANT,
			Value: utils.ParseRSRFieldsMustCompile(utils.VOICE, utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "OriginID", FieldId: utils.OriginID, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("Acct-Session-Id;^-;Sip-From-Tag;^-;Sip-To-Tag", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "OriginHost", FieldId: utils.OriginHost, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("NAS-IP-Address", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "RequestType", FieldId: utils.RequestType, Type: utils.META_CONSTANT,
			Value: utils.ParseRSRFieldsMustCompile(utils.META_PREPAID, utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Direction", FieldId: utils.Direction, Type: utils.META_CONSTANT,
			Value: utils.ParseRSRFieldsMustCompile(utils.OUT, utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Tenant", FieldId: utils.Tenant, Type: utils.META_CONSTANT,
			Value: utils.ParseRSRFieldsMustCompile("cgrates.org", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Category", FieldId: utils.Category, Type: utils.META_CONSTANT,
			Value: utils.ParseRSRFieldsMustCompile("call", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Account", FieldId: utils.Account, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("User-Name", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Destination", FieldId: utils.Destination, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("Called-Station-Id", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "SetupTime", FieldId: utils.SetupTime, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("Ascend-User-Acct-Time", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "AnswerTime", FieldId: utils.AnswerTime, Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("Ascend-User-Acct-Time", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Usage", FieldId: utils.Usage, Type: utils.META_HANDLER, HandlerId: MetaUsageDifference,
			Value: utils.ParseRSRFieldsMustCompile("Event-Timestamp;^|;Ascend-User-Acct-Time", utils.INFIELD_SEP)},
	}
	eOut := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.AnswerTime:  "1497106115",
			utils.Category:    "call",
			utils.Destination: "1002",
			utils.Direction:   utils.META_OUT,
			utils.OriginHost:  "127.0.0.1",
			utils.OriginID:    "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-75c2f57b-51585361",
			utils.RequestType: "*prepaid",
			utils.SetupTime:   "1497106115",
			utils.Tenant:      "cgrates.org",
			utils.TOR:         "*voice",
			utils.Usage:       "4s",
		},
	}
	if outVal, err := radReqAsCGREvent(pkt, processorVars{MetaRadReqType: MetaRadAcctStart}, nil, cfgFlds); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(outVal.Tenant, eOut.Tenant) {
		t.Errorf("Expecting: <%s>, received: <%s>", utils.ToJSON(eOut.Tenant), utils.ToJSON(outVal.Tenant))
	} else if !reflect.DeepEqual(outVal.Event, eOut.Event) {
		t.Errorf("Expecting: <%s>, received: <%s>", utils.ToJSON(eOut.Event), utils.ToJSON(outVal.Event))
	}
}

func TestRadV1AuthorizeArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.AnswerTime:  "1497106115",
			utils.Category:    "call",
			utils.Destination: "1002",
			utils.Direction:   utils.META_OUT,
			utils.OriginHost:  "127.0.0.1",
			utils.OriginID:    "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-75c2f57b-51585361",
			utils.RequestType: "*prepaid",
			utils.SetupTime:   "1497106115",
			utils.Tenant:      "cgrates.org",
			utils.TOR:         "*voice",
			utils.Usage:       "4s",
		},
	}
	expected := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent:    *cgrEv,
	}
	outVal := radV1AuthorizeArgs(cgrEv, processorVars{MetaRadReqType: MetaRadAcctStart})

	if !reflect.DeepEqual(expected, outVal) {
		t.Errorf("Expecting: <%s>, received: <%s>", utils.ToJSON(expected), utils.ToJSON(outVal))
	}
}

func TestRadV1InitSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.AnswerTime:  "1497106115",
			utils.Category:    "call",
			utils.Destination: "1002",
			utils.Direction:   utils.META_OUT,
			utils.OriginHost:  "127.0.0.1",
			utils.OriginID:    "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-75c2f57b-51585361",
			utils.RequestType: "*prepaid",
			utils.SetupTime:   "1497106115",
			utils.Tenant:      "cgrates.org",
			utils.TOR:         "*voice",
			utils.Usage:       "4s",
		},
	}
	expected := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent:    *cgrEv,
	}
	outVal := radV1InitSessionArgs(cgrEv, processorVars{MetaRadReqType: MetaRadAcctStart})

	if !reflect.DeepEqual(expected, outVal) {
		t.Errorf("Expecting: <%s>, received: <%s>", utils.ToJSON(expected), utils.ToJSON(outVal))
	}
}

func TestRadV1UpdateSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.AnswerTime:  "1497106115",
			utils.Category:    "call",
			utils.Destination: "1002",
			utils.Direction:   utils.META_OUT,
			utils.OriginHost:  "127.0.0.1",
			utils.OriginID:    "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-75c2f57b-51585361",
			utils.RequestType: "*prepaid",
			utils.SetupTime:   "1497106115",
			utils.Tenant:      "cgrates.org",
			utils.TOR:         "*voice",
			utils.Usage:       "4s",
		},
	}
	expected := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent:      *cgrEv,
	}
	outVal := radV1UpdateSessionArgs(cgrEv, processorVars{MetaRadReqType: MetaRadAcctStart})

	if !reflect.DeepEqual(expected, outVal) {
		t.Errorf("Expecting: <%s>, received: <%s>", utils.ToJSON(expected), utils.ToJSON(outVal))
	}
}

func TestRadV1TerminateSessionArgs(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]interface{}{
			utils.Account:     "1001",
			utils.AnswerTime:  "1497106115",
			utils.Category:    "call",
			utils.Destination: "1002",
			utils.Direction:   utils.META_OUT,
			utils.OriginHost:  "127.0.0.1",
			utils.OriginID:    "e4921177ab0e3586c37f6a185864b71a@0:0:0:0:0:0:0:0-75c2f57b-51585361",
			utils.RequestType: "*prepaid",
			utils.SetupTime:   "1497106115",
			utils.Tenant:      "cgrates.org",
			utils.TOR:         "*voice",
			utils.Usage:       "4s",
		},
	}
	expected := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent:         *cgrEv,
	}
	outVal := radV1TerminateSessionArgs(cgrEv, processorVars{MetaRadReqType: MetaRadAcctStart})

	if !reflect.DeepEqual(expected, outVal) {
		t.Errorf("Expecting: <%s>, received: <%s>", utils.ToJSON(expected), utils.ToJSON(outVal))
	}
}

/*
func TestRadReplyAppendAttributes(t *testing.T) {
	rply := radigo.NewPacket(radigo.AccessRequest, 2, dictRad, coder, "CGRateS.org").Reply()
	rplyFlds := []*config.CfgCdrField{
		&config.CfgCdrField{Tag: "ReplyCode", FieldId: MetaRadReplyCode, Type: utils.META_CONSTANT,
			Value: utils.ParseRSRFieldsMustCompile("AccessAccept", utils.INFIELD_SEP)},
		&config.CfgCdrField{Tag: "Acct-Session-Time", FieldId: "Acct-Session-Time", Type: utils.META_COMPOSED,
			Value: utils.ParseRSRFieldsMustCompile("~*cgrMaxUsage:s/(\\d*)\\d{9}$/$1/", utils.INFIELD_SEP)},
	}
	if err := radReplyAppendAttributes(rply, processorVars{MetaCGRMaxUsage: "30000000000"}, rplyFlds); err != nil {
		t.Error(err)
	}
	if rply.Code != radigo.AccessAccept {
		t.Errorf("Wrong reply code: %d", rply.Code)
	}
	if avps := rply.AttributesWithName("Acct-Session-Time", ""); len(avps) == 0 {
		t.Error("Cannot find Acct-Session-Time in reply")
	} else if avps[0].GetStringValue() != "30" {
		t.Errorf("Expecting: 30, received: %s", avps[0].GetStringValue())
	}
}
*/
