/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOev.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestAttributesV1ProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := &ConnManager{}
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		Contexts:  []string{utils.MetaAny},
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
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
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant:   "adrian.itsyscom.com.co.uk",
		ID:       "ATTR_MATCH_TENANT",
		Contexts: []string{utils.MetaAny},
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

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
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
			Event: map[string]interface{}{
				utils.AccountField: "andrei.itsyscom.com",
				"Password":         "CGRATES.ORG",
			},
			APIOpts: map[string]interface{}{},
		},
		blocker: false,
	}
	err = alS.V1ProcessEvent(context.Background(), args, rply)
	sort.Strings(rply.AlteredFields)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}
