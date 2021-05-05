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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/sipingo"
)

func TestUpdateSIPMsgFromNavMap(t *testing.T) {
	m := sipingo.Message{}
	rplyFlds := []*config.FCTemplate{
		{Tag: "Request", Path: utils.MetaRep + utils.NestingSep + "Request",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.Request", utils.InfieldSep)},
		{Tag: "Contact", Path: utils.MetaRep + utils.NestingSep + "Contact",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.Account", utils.InfieldSep)},
	}
	for _, v := range rplyFlds {
		v.ComputePath()
	}
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", nil, nil)
	agReq.CGRReply.Set([]string{utils.CapMaxUsage}, utils.NewLeafNode(time.Hour))
	agReq.CGRReply.Set([]string{utils.CapAttributes, "Request"}, utils.NewLeafNode("SIP/2.0 302 Moved Temporarily"))
	agReq.CGRReply.Set([]string{utils.CapAttributes, utils.AccountField}, utils.NewLeafNode("1001"))

	if err := agReq.SetFields(rplyFlds); err != nil {
		t.Error(err)
	}
	if err := updateSIPMsgFromNavMap(m, agReq.Reply); err != nil {
		t.Error(err)
	}
	expected := sipingo.Message{
		"Request": "SIP/2.0 302 Moved Temporarily",
		"Contact": "1001",
	}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected: %s , received: %s", expected, m)
	}
}
