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

package engine

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestAttributesShutdown(t *testing.T) {
	alS := &AttributeService{}

	utils.Logger.SetLogLevel(6)
	utils.Logger.SetSyslog(nil)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	exp := []string{
		"CGRateS <> [INFO] <AttributeS> shutdown initialized",
		"CGRateS <> [INFO] <AttributeS> shutdown complete",
	}
	alS.Shutdown()
	rcv := strings.Split(buf.String(), "\n")

	for i := 0; i < 2; i++ {
		rcv[i] = rcv[i][20:]
		if rcv[i] != exp[i] {
			t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp[i], rcv[i])
		}
	}

	utils.Logger.SetLogLevel(0)
}

func TestAttributesSetCloneable(t *testing.T) {
	attr := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		clnb:         true,
	}

	exp := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		clnb:         false,
	}
	attr.SetCloneable(false)

	if !reflect.DeepEqual(attr, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, attr)
	}
}

func TestAttributesRPCClone(t *testing.T) {
	attr := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
		},
		clnb: true,
	}

	rcv, err := attr.RPCClone()

	exp := &AttrArgsProcessEvent{
		AttributeIDs: []string{"ATTR_ID"},
		Context:      utils.StringPointer(utils.MetaAny),
		ProcessRuns:  utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Event:   make(map[string]interface{}),
			APIOpts: make(map[string]interface{}),
		},
		clnb: false,
	}

	if err != nil {
		t.Fatalf("\nexpected: <%+v>, \nreceived: <%+v>", nil, err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}

}

func TestAttributesV1GetAttributeForEventNilCGREvent(t *testing.T) {
	alS := &AttributeService{}
	args := &AttrArgsProcessEvent{}
	reply := &AttributeProfile{}

	experr := fmt.Sprintf("MANDATORY_IE_MISSING: [%s]", "CGREvent")
	err := alS.V1GetAttributeForEvent(args, reply)

	if err == nil || err.Error() != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1GetAttributeForEventProfileNotFound(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, nil)
	alS := &AttributeService{
		dm:      dm,
		filterS: &FilterS{},
		cgrcfg:  defaultCfg,
	}
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{},
	}
	reply := &AttributeProfile{}

	experr := utils.ErrNotFound
	err := alS.V1GetAttributeForEvent(args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1GetAttributeForEvent2(t *testing.T) {
	defaultCfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, nil)
	alS := &AttributeService{
		dm:      dm,
		filterS: &FilterS{},
		cgrcfg:  defaultCfg,
	}
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{},
	}
	reply := &AttributeProfile{}

	experr := utils.ErrNotFound
	err := alS.V1GetAttributeForEvent(args, reply)

	if err == nil || err != experr {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1ProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := &ConnManager{}
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:             "cgrates.org",
		ID:                 "ATTR_CHANGE_TENANT_FROM_USER",
		Contexts:           []string{utils.MetaAny},
		FilterIDs:          []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant:             "adrian.itsyscom.com.co.uk",
		ID:                 "ATTR_MATCH_TENANT",
		Contexts:           []string{utils.MetaAny},
		ActivationInterval: nil,
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	args := &AttrArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "123",
			Event: map[string]interface{}{
				utils.AccountField: "adrian@itsyscom.com",
			},
		},
		ProcessRuns: utils.IntPointer(2),
	}
	rply := &AttrSProcessEventReply{}
	expected := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_CHANGE_TENANT_FROM_USER", "ATTR_MATCH_TENANT"},
		AlteredFields:   []string{"*req.Account", "*req.Password", "*tenant"},
		CGREvent: &utils.CGREvent{
			Tenant: "adrian.itsyscom.com.co.uk",
			ID:     "123",
			Time:   nil,
			Event: map[string]interface{}{
				utils.AccountField: "andrei.itsyscom.com",
				"Password":         "CGRATES.ORG",
			},
			APIOpts: map[string]interface{}{},
		},
		blocker: false,
	}
	err = alS.V1ProcessEvent(args, rply)
	sort.Strings(rply.AlteredFields)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}
