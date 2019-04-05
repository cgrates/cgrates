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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestFieldinfo2Attribute(t *testing.T) {
	type testA struct {
		FieldName string
		FieldInfo string
		Initial   []*engine.Attribute
		Expected  []*engine.Attribute
	}
	tests := []testA{
		testA{
			FieldName: utils.Account,
			FieldInfo: utils.MetaDefault,
			Initial:   make([]*engine.Attribute, 0),
			Expected:  make([]*engine.Attribute, 0),
		},
		testA{
			FieldName: utils.Account,
			FieldInfo: "",
			Initial:   make([]*engine.Attribute, 0),
			Expected:  make([]*engine.Attribute, 0),
		},
		testA{
			FieldName: utils.Account,
			FieldInfo: "^1003",
			Initial:   make([]*engine.Attribute, 0),
			Expected: []*engine.Attribute{
				&engine.Attribute{
					FieldName: utils.Account,
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1003", true, utils.INFIELD_SEP),
				},
			},
		},
		testA{
			FieldName: utils.Subject,
			FieldInfo: `~effective_caller_id_number:s/(\d+)/+$1/`,
			Initial:   make([]*engine.Attribute, 0),
			Expected: []*engine.Attribute{
				&engine.Attribute{
					FieldName: utils.Subject,
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile(`~effective_caller_id_number:s/(\d+)/+$1/`, true, utils.INFIELD_SEP),
				},
			},
		},
		testA{
			FieldName: utils.Subject,
			FieldInfo: "^call_1003",
			Initial: []*engine.Attribute{
				&engine.Attribute{
					FieldName: utils.Account,
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1003", true, utils.INFIELD_SEP),
				},
			},
			Expected: []*engine.Attribute{
				&engine.Attribute{
					FieldName: utils.Account,
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("1003", true, utils.INFIELD_SEP),
				},
				&engine.Attribute{
					FieldName: utils.Subject,
					Type:      utils.MetaVariable,
					Value:     config.NewRSRParsersMustCompile("call_1003", true, utils.INFIELD_SEP),
				},
			},
		},
	}
	for i, v := range tests {
		if rply := fieldinfo2Attribute(v.Initial, v.FieldName, v.FieldInfo); !reflect.DeepEqual(v.Expected, rply) {
			t.Errorf("For %v expected: %s ,recieved: %s", i, utils.ToJSON(v.Expected), utils.ToJSON(rply))
		}
	}
}

func TestDerivedChargers2AttributeProfile(t *testing.T) {
	type testC struct {
		DC       *v1DerivedCharger
		Tenant   string
		Key      string
		Filters  []string
		Expected *engine.AttributeProfile
	}
	tests := []testC{
		testC{
			DC: &v1DerivedCharger{
				RequestTypeField: utils.MetaDefault,
				CategoryField:    "^*voice",
				AccountField:     "^1003",
			},
			Tenant:  defaultTenant,
			Key:     "key1",
			Filters: make([]string, 0),
			Expected: &engine.AttributeProfile{
				Tenant:             defaultTenant,
				ID:                 "key1",
				Contexts:           []string{utils.MetaChargers},
				FilterIDs:          make([]string, 0),
				ActivationInterval: nil,
				Attributes: []*engine.Attribute{
					&engine.Attribute{
						FieldName: utils.Category,
						Type:      utils.MetaVariable,
						Value:     config.NewRSRParsersMustCompile("*voice", true, utils.INFIELD_SEP),
					},
					&engine.Attribute{
						FieldName: utils.Account,
						Type:      utils.MetaVariable,
						Value:     config.NewRSRParsersMustCompile("1003", true, utils.INFIELD_SEP),
					},
				},
				Blocker: false,
				Weight:  10,
			},
		},
		testC{
			DC: &v1DerivedCharger{
				RequestTypeField: utils.MetaDefault,
				CategoryField:    "^*voice",
				AccountField:     "^1003",
				SubjectField:     "call_1003_to_1004",
				DestinationField: "^1004",
			},
			Tenant:  defaultTenant,
			Key:     "key1",
			Filters: []string{"*string:~Subject:1005"},
			Expected: &engine.AttributeProfile{
				Tenant:             defaultTenant,
				ID:                 "key1",
				Contexts:           []string{utils.MetaChargers},
				FilterIDs:          []string{"*string:~Subject:1005"},
				ActivationInterval: nil,
				Attributes: []*engine.Attribute{
					&engine.Attribute{
						FieldName: utils.Category,
						Type:      utils.MetaVariable,
						Value:     config.NewRSRParsersMustCompile("*voice", true, utils.INFIELD_SEP),
					},
					&engine.Attribute{
						FieldName: utils.Account,
						Type:      utils.MetaVariable,
						Value:     config.NewRSRParsersMustCompile("1003", true, utils.INFIELD_SEP),
					},
					&engine.Attribute{
						FieldName: utils.Subject,
						Type:      utils.MetaVariable,
						Value:     config.NewRSRParsersMustCompile("call_1003_to_1004", true, utils.INFIELD_SEP),
					},
					&engine.Attribute{
						FieldName: utils.Destination,
						Type:      utils.MetaVariable,
						Value:     config.NewRSRParsersMustCompile("1004", true, utils.INFIELD_SEP),
					},
				},
				Blocker: false,
				Weight:  10,
			},
		},
	}
	for i, v := range tests {
		if rply := derivedChargers2AttributeProfile(v.DC, v.Tenant, v.Key, v.Filters); !reflect.DeepEqual(v.Expected, rply) {
			t.Errorf("For %v expected: %s ,recieved: %s", i, utils.ToJSON(v.Expected), utils.ToJSON(rply))
		}
	}
}

func TestDerivedChargers2Charger(t *testing.T) {
	type testB struct {
		DC       *v1DerivedCharger
		Tenant   string
		Key      string
		Filters  []string
		Expected *engine.ChargerProfile
	}
	tests := []testB{
		testB{
			DC: &v1DerivedCharger{
				RunID:            "runID",
				RunFilters:       "~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)",
				RequestTypeField: utils.MetaDefault,
				CategoryField:    "^*voice",
				AccountField:     "^1003",
			},
			Tenant: defaultTenant,
			Key:    "key2",
			Filters: []string{
				"*string:~Category:*voice1",
				"*string:~Account:1001",
			},
			Expected: &engine.ChargerProfile{
				Tenant: defaultTenant,
				ID:     "key2",
				FilterIDs: []string{
					"*string:~Category:*voice1",
					"*string:~Account:1001",
					"*rsr::~Header4:s/a/${1}b/{*duration_seconds&*round:2}(b&c)",
				},
				ActivationInterval: nil,
				RunID:              "runID",
				AttributeIDs:       make([]string, 0),
				Weight:             10,
			},
		},
		testB{
			DC: &v1DerivedCharger{
				RunID:        "runID2",
				RunFilters:   "^1003",
				AccountField: "^1003",
			},
			Tenant:  defaultTenant,
			Key:     "key2",
			Filters: []string{},
			Expected: &engine.ChargerProfile{
				Tenant:             defaultTenant,
				ID:                 "key2",
				FilterIDs:          []string{"*rsr::1003"},
				ActivationInterval: nil,
				RunID:              "runID2",
				AttributeIDs:       make([]string, 0),
				Weight:             10,
			},
		},
	}
	for i, v := range tests {
		if rply := derivedChargers2Charger(v.DC, v.Tenant, v.Key, v.Filters); !reflect.DeepEqual(v.Expected, rply) {
			t.Errorf("For %v expected: %s ,recieved: %s", i, utils.ToJSON(v.Expected), utils.ToJSON(rply))
		}
	}
}
