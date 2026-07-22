// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func TestLibsipBareSipErr(t *testing.T) {
	m := sipingo.Message{
		"requestHeader": "test-header",
	}
	errMsg := "test-header"
	updatedMsg := bareSipErr(m, errMsg)
	if updatedMsg["requestHeader"] != errMsg {
		t.Errorf("expected error message %s, got %s", errMsg, updatedMsg["requestHeader"])
	}
}
