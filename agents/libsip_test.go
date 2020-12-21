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
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.Request", utils.INFIELD_SEP)},
		{Tag: "Contact", Path: utils.MetaRep + utils.NestingSep + "Contact",
			Type:  utils.MetaVariable,
			Value: config.NewRSRParsersMustCompile("~*cgrep.Attributes.Account", utils.INFIELD_SEP)},
	}
	for _, v := range rplyFlds {
		v.ComputePath()
	}
	agReq := NewAgentRequest(nil, nil, nil, nil, nil, nil, "cgrates.org", "", nil, nil, nil)
	agReq.CGRReply.Set(utils.NewPathItems([]string{utils.CapMaxUsage}), utils.NewNMData(time.Hour))
	agReq.CGRReply.Set(utils.NewPathItems([]string{utils.CapAttributes, "Request"}), utils.NewNMData("SIP/2.0 302 Moved Temporarily"))
	agReq.CGRReply.Set(utils.NewPathItems([]string{utils.CapAttributes, utils.AccountField}), utils.NewNMData("1001"))

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

func TestUpdateSIPMsgFromNavMap2(t *testing.T) {
	m := sipingo.Message{}
	mv := utils.NewOrderedNavigableMap()
	var nm *config.NMItem
	mv.Set(&utils.FullPath{PathItems: utils.NewPathItems([]string{utils.CapAttributes, utils.AccountField}), Path: utils.CapAttributes + utils.NestingSep + utils.AccountField}, nm)
	mv.Set(&utils.FullPath{PathItems: utils.NewPathItems([]string{utils.CapMaxUsage}), Path: utils.CapMaxUsage}, utils.NewNMData(time.Hour))

	expectedErr := `cannot encode reply value: [{"Field":"MaxUsage","Index":null}], err: not NMItems`
	if err := updateSIPMsgFromNavMap(m, mv); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected error %s,received:%v", expectedErr, err)
	}
}
