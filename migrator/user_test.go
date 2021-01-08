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

package migrator

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestUserProfile2attributeProfile(t *testing.T) {
	usrCfgIn := config.CgrConfig()
	usrCfgIn.MigratorCgrCfg().UsersFilters = []string{"Account"}
	config.SetCgrConfig(usrCfgIn)
	usrTenant := "cgrates.com"
	users := map[int]*v1UserProfile{
		0: {
			Tenant:   defaultTenant,
			UserName: "1001",
			Masked:   true,
			Profile:  map[string]string{},
			Weight:   10,
		},
		1: {
			Tenant:   usrTenant,
			UserName: "1001",
			Masked:   true,
			Profile: map[string]string{
				"Account": "1002",
				"Subject": "call_1001",
			},
			Weight: 10,
		},
		2: {
			Tenant:   defaultTenant,
			UserName: "1001",
			Masked:   false,
			Profile: map[string]string{
				"Account": "1002",
				"ReqType": "*prepaid",
				"msisdn":  "123423534646752",
			},
			Weight: 10,
		},
		3: {
			Tenant:   usrTenant,
			UserName: "1001",
			Masked:   false,
			Profile: map[string]string{
				"Account": "1002",
				"ReqType": "*prepaid",
			},
			Weight: 10,
		},
		4: {
			Tenant:   usrTenant,
			UserName: "acstmusername",
			Profile: map[string]string{
				"Account": "acnt63",
				"Subject": "acnt63",
				"ReqType": "*prepaid",
				"msisdn":  "12345",
				"imsi":    "12345",
			},
			Weight: 10,
		},
	}
	expected := map[int]*engine.AttributeProfile{
		0: {
			Tenant:             defaultTenant,
			ID:                 "1001",
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes:         []*engine.Attribute{},
			Blocker:            false,
			Weight:             10,
		},
		1: {
			Tenant:             defaultTenant,
			ID:                 "1001",
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          []string{"*string:~Account:1002"},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Subject",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("call_1001", utils.InfieldSep),
				},
				{
					Path:  utils.MetaTenant,
					Type:  utils.META_CONSTANT,
					Value: config.NewRSRParsersMustCompile(usrTenant, utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
		2: {
			Tenant:   defaultTenant,
			ID:       "1001",
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{
				"*string:~Account:1002",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.RequestType,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("*prepaid", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "msisdn",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("123423534646752", utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
		3: {
			Tenant:             defaultTenant,
			ID:                 "1001",
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          []string{"*string:~Account:1002"},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.RequestType,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("*prepaid", utils.InfieldSep),
				},
				{
					Path:  utils.MetaTenant,
					Type:  utils.META_CONSTANT,
					Value: config.NewRSRParsersMustCompile(usrTenant, utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
		4: {
			Tenant:   defaultTenant,
			ID:       "acstmusername",
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{
				"*string:~Account:acnt63",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.RequestType,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("*prepaid", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("acnt63", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "imsi",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("12345", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "msisdn",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("12345", utils.InfieldSep),
				},
				{
					Path:  utils.MetaTenant,
					Type:  utils.META_CONSTANT,
					Value: config.NewRSRParsersMustCompile(usrTenant, utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  10,
		},
	}
	for i := range expected {
		rply := userProfile2attributeProfile(users[i])
		sort.Slice(rply.Attributes, func(i, j int) bool {
			if rply.Attributes[i].Path == rply.Attributes[j].Path {
				return rply.Attributes[i].FilterIDs[0] < rply.Attributes[j].FilterIDs[0]
			}
			return rply.Attributes[i].Path < rply.Attributes[j].Path
		}) // only for test; map returns random keys
		if !reflect.DeepEqual(expected[i], rply) {
			t.Errorf("For %v expected: %s ,\nreceived: %s ", i, utils.ToJSON(expected[i]), utils.ToJSON(rply))
		}
	}
}
