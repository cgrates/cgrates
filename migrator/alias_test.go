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

var defaultTenant = "cgrates.org"

func TestAlias2AtttributeProfile(t *testing.T) {
	aliases := map[int]*v1Alias{
		0: {
			Tenant:    utils.MetaAny,
			Direction: utils.MetaOut,
			Category:  utils.MetaAny,
			Account:   utils.MetaAny,
			Subject:   utils.MetaAny,
			Context:   "*rating",
			Values:    v1AliasValues{},
		},
		1: {
			Tenant:    utils.MetaAny,
			Direction: utils.MetaOut,
			Category:  utils.MetaAny,
			Account:   utils.MetaAny,
			Subject:   utils.MetaAny,
			Context:   "*rating",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: utils.MetaAny,
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
						},
					},
					Weight: 20,
				},
			},
		},
		2: {
			Tenant:    utils.MetaAny,
			Direction: utils.MetaOut,
			Category:  utils.MetaAny,
			Account:   utils.MetaAny,
			Subject:   utils.MetaAny,
			Context:   "*rating",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: utils.MetaAny,
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
							"1003": "1004",
						},
					},
					Weight: 10,
				},
			},
		},
		3: {
			Tenant:    "",
			Direction: "",
			Category:  "",
			Account:   "",
			Subject:   "",
			Context:   "",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: utils.MetaAny,
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
							"1003": "1004",
						},
					},
					Weight: 10,
				},
			},
		},
		4: {
			Tenant:    "notDefaultTenant",
			Direction: "*out",
			Category:  "*voice",
			Account:   "1001",
			Subject:   utils.MetaAny,
			Context:   "*rated",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: "DST_1003",
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"": "1002",
						},
						"Subject": map[string]string{
							"": "call_1001",
						},
					},
					Weight: 10,
				},
			},
		},
		5: {
			Tenant:    "notDefaultTenant",
			Direction: "*out",
			Category:  utils.MetaAny,
			Account:   "1001",
			Subject:   "call_1001",
			Context:   "*rated",
			Values: v1AliasValues{
				&v1AliasValue{
					DestinationId: "DST_1003",
					Pairs: map[string]map[string]string{
						"Account": map[string]string{
							"1001": "1002",
						},
						"Category": map[string]string{
							"call_1001": "call_1002",
						},
					},
					Weight: 10,
				},
			},
		},
		6: {
			Tenant:   utils.MetaAny,
			Category: "somecateg_5141",
			Account:  utils.MetaAny,
			Subject:  utils.MetaAny,
			Context:  "*rated",
			Values: v1AliasValues{
				&v1AliasValue{
					Pairs: map[string]map[string]string{
						utils.Category: map[string]string{
							"somecateg_5141": "somecateg_roam_fromz4",
						},
					},
				},
			},
		},
	}
	expected := map[int]*engine.AttributeProfile{
		0: {
			Tenant:             defaultTenant,
			ID:                 aliases[0].GetId(),
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes:         make([]*engine.Attribute, 0),
			Blocker:            false,
			Weight:             20,
		},
		1: {
			Tenant:             defaultTenant,
			ID:                 aliases[1].GetId(),
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Path:      utils.MetaReq + utils.NestingSep + "Account",
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1002", utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  20,
		},
		2: {
			Tenant:             defaultTenant,
			ID:                 aliases[2].GetId(),
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Path:      utils.MetaReq + utils.NestingSep + "Account",
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1002", utils.InfieldSep),
				},
				{
					FilterIDs: []string{"*string:~*req.Account:1003"},
					Path:      utils.MetaReq + utils.NestingSep + "Account",
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1004", utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  20,
		},
		3: {
			Tenant:             defaultTenant,
			ID:                 aliases[3].GetId(),
			Contexts:           []string{utils.MetaAny},
			FilterIDs:          make([]string, 0),
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Path:      utils.MetaReq + utils.NestingSep + "Account",
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1002", utils.InfieldSep),
				},
				{
					FilterIDs: []string{"*string:~*req.Account:1003"},
					Path:      utils.MetaReq + utils.NestingSep + "Account",
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1004", utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  20,
		},
		4: {
			Tenant:   "notDefaultTenant",
			ID:       aliases[4].GetId(),
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{
				"*string:~*req.Category:*voice",
				"*string:~*req.Account:1001",
				"*destinations:~*req.Destination:DST_1003",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("1002", utils.InfieldSep),
				},
				{
					Path:  utils.MetaReq + utils.NestingSep + "Subject",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("call_1001", utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  20,
		},
		5: {
			Tenant:   "notDefaultTenant",
			ID:       aliases[5].GetId(),
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{
				"*string:~*req.Account:1001",
				"*string:~*req.Subject:call_1001",
				"*destinations:~*req.Destination:DST_1003",
			},
			ActivationInterval: nil,
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Account",
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("1002", utils.InfieldSep),
				},
				{
					Path:      utils.MetaReq + utils.NestingSep + "Category",
					Type:      utils.MetaVariable,
					FilterIDs: []string{"*string:~*req.Category:call_1001"},
					Value:     config.NewRSRParsersMustCompile("call_1002", utils.InfieldSep),
				},
			},
			Blocker: false,
			Weight:  20,
		},
		6: {
			Tenant:   "cgrates.org",
			ID:       aliases[6].GetId(),
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{
				"*string:~*req.Category:somecateg_5141",
			},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Category,
					Type:  utils.MetaVariable,
					Value: config.NewRSRParsersMustCompile("somecateg_roam_fromz4", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
	for i := range expected {
		rply := alias2AtttributeProfile(aliases[i], defaultTenant)
		sort.Slice(rply.Attributes, func(i, j int) bool {
			if rply.Attributes[i].Path == rply.Attributes[j].Path {
				return rply.Attributes[i].FilterIDs[0] < rply.Attributes[j].FilterIDs[0]
			}
			return rply.Attributes[i].Path < rply.Attributes[j].Path
		}) // only for test; map returns random keys
		if !reflect.DeepEqual(expected[i], rply) {
			t.Errorf("For %v expected: %s ,received: %s ", i, utils.ToJSON(expected[i]), utils.ToJSON(rply))
		}
	}
}
