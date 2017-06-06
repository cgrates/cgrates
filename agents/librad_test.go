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
	"strings"
	"testing"

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

# Vendors
VENDOR    Cisco    9
VENDOR    Microsoft 311

# Vendor AVPs
BEGIN-VENDOR    Cisco
ATTRIBUTE       Cisco-AVPair    1   string
ATTRIBUTE       Cisco-NAS-Port  2	string
END-VENDOR      Cisco
`

func init() {
	dictRad = radigo.RFC2865Dictionary()
	dictRad.ParseFromReader(strings.NewReader(freeRADIUSDocDictSample))
	coder = radigo.NewCoder()
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
	if !radPassesFieldFilter(pkt, map[string]string{MetaRadReqCode: "4"},
		utils.NewRSRFieldMustCompile(fmt.Sprintf("%s(4)", MetaRadReqCode))) {
		t.Error("not passing valid filter")
	}
	if radPassesFieldFilter(pkt, map[string]string{MetaRadReqCode: "5"},
		utils.NewRSRFieldMustCompile(fmt.Sprintf("%s(4)", MetaRadReqCode))) {
		t.Error("passing invalid filter")
	}
	if radPassesFieldFilter(pkt, nil,
		utils.NewRSRFieldMustCompile("UnknownField(notmatching)")) {
		t.Error("passing invalid filter value")
	}
}
