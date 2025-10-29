/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package migrator

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestFiltersInlineMigrate(t *testing.T) {
	data := []struct{ in, exp string }{
		{
			in:  "*string:Account:1002",
			exp: "*string:~*req.Account:1002",
		},
		{
			in:  "*string:~Account:1002",
			exp: "*string:~Account:1002",
		},
		{
			in:  "FLTR_1",
			exp: "FLTR_1",
		},
		{
			in:  "",
			exp: "",
		},
		{
			in:  "*rsr::~Tenant(~^cgr.*\\.org$)",
			exp: "*rsr::~Tenant(~^cgr.*\\.org$)",
		},
	}
	for _, m := range data {
		if rply := migrateInlineFilter(m.in); rply != m.exp {
			t.Errorf("Expected: %s, received: %s", m.exp, rply)
		}
	}

}

func TestFiltersMigrate(t *testing.T) {
	data := []struct {
		in  *v1Filter
		exp *engine.Filter
	}{
		{
			in: &v1Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_1",
				Rules: []*v1FilterRule{
					{
						Type:      utils.MetaString,
						FieldName: "Account",
						Values:    []string{},
					},
				},
			},
			exp: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_1",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
						Values:  []string{},
					},
				},
			},
		},
		{
			in: &v1Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_2",
				Rules: []*v1FilterRule{
					{
						Type:      utils.MetaPrefix,
						FieldName: "~Account",
						Values:    []string{},
					},
				},
			},
			exp: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_2",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~Account",
						Values:  []string{},
					},
				},
			},
		},
	}
	for _, m := range data {
		if rply := migrateFilterV1(m.in); !reflect.DeepEqual(rply, m.exp) {
			t.Errorf("Expected: %s, received: %s", utils.ToJSON(m.exp), utils.ToJSON(rply))
		}
	}
}

func TestFiltersMigrateV2(t *testing.T) {
	data := []struct {
		in  *v1Filter
		exp *engine.Filter
	}{
		{
			in: &v1Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_1",
				Rules: []*v1FilterRule{
					{
						Type:      utils.MetaString,
						FieldName: "~Account",
						Values:    []string{},
					},
				},
			},
			exp: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_1",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
						Values:  []string{},
					},
				},
			},
		},
		{
			in: &v1Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_2",
				Rules: []*v1FilterRule{
					{
						Type:      utils.MetaPrefix,
						FieldName: "~*req.Account",
						Values:    []string{},
					},
				},
			},
			exp: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_2",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
						Values:  []string{},
					},
				},
			},
		},
		{
			in: &v1Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_3",
				Rules: []*v1FilterRule{
					{
						Type:      utils.MetaPrefix,
						FieldName: "~*act.Account",
						Values:    []string{},
					},
				},
			},
			exp: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_3",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~*act.Account",
						Values:  []string{},
					},
				},
			},
		},
		{
			in: &v1Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_4",
				Rules: []*v1FilterRule{
					{
						Type:      utils.MetaPrefix,
						FieldName: "~*act.Account",
						Values:    []string{},
					},
				},
			},
			exp: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_4",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~*act.Account",
						Values:  []string{},
					},
				},
			},
		},
		{
			in: &v1Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_5",
				Rules: []*v1FilterRule{
					{
						Type:      utils.MetaPrefix,
						FieldName: "~*vars.Account",
						Values:    []string{},
					},
				},
			},
			exp: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_5",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaPrefix,
						Element: "~*vars.Account",
						Values:  []string{},
					},
				},
			},
		},
	}
	for _, m := range data {
		if rply := migrateFilterV2(m.in); !reflect.DeepEqual(rply, m.exp) {
			t.Errorf("Expected: %s, received: %s", utils.ToJSON(m.exp), utils.ToJSON(rply))
		}
	}
}

func TestFiltersInlineV2Migrate(t *testing.T) {
	data := []struct{ in, exp string }{
		{
			in:  "*string:~Account:1002",
			exp: "*string:~*req.Account:1002",
		},
		{
			in:  "*string:~*req.Account:1002",
			exp: "*string:~*req.Account:1002",
		},
		{
			in:  "FLTR_1",
			exp: "FLTR_1",
		},
		{
			in:  "",
			exp: "",
		},
		{
			in:  "*rsr::~Tenant(~^cgr.*\\.org$)",
			exp: "*rsr::~*req.Tenant(~^cgr.*\\.org$)",
		},
	}
	for _, m := range data {
		if rply := migrateInlineFilterV2(m.in); rply != m.exp {
			t.Errorf("Expected: %s, received: %s", m.exp, rply)
		}
	}

}

func TestMigrateFilterV3(t *testing.T) {
	v1F := &v1Filter{
		Tenant: "cgrates.org",
		ID:     "filter1",
		Rules: []*v1FilterRule{
			{Type: "type1", FieldName: "field1", Values: []string{"value1", "value2"}},
			{Type: "type2", FieldName: "field2", Values: []string{"value3"}},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Now(),
			ExpiryTime:     time.Now().Add(24 * time.Hour),
		},
	}
	fltr := migrateFilterV3(v1F)
	if fltr.Tenant != v1F.Tenant {
		t.Errorf("expected Tenant %v, got %v", v1F.Tenant, fltr.Tenant)
	}
	if fltr.ID != v1F.ID {
		t.Errorf("expected ID %v, got %v", v1F.ID, fltr.ID)
	}
	if len(fltr.Rules) != len(v1F.Rules) {
		t.Errorf("expected %v rules, got %v", len(v1F.Rules), len(fltr.Rules))
	} else {
		for i, rule := range v1F.Rules {
			if fltr.Rules[i].Type != rule.Type {
				t.Errorf("for rule %d, expected Type %v, got %v", i, rule.Type, fltr.Rules[i].Type)
			}
			if fltr.Rules[i].Element != rule.FieldName {
				t.Errorf("for rule %d, expected FieldName %v, got %v", i, rule.FieldName, fltr.Rules[i].Element)
			}
			if len(fltr.Rules[i].Values) != len(rule.Values) {
				t.Errorf("for rule %d, expected %v values, got %v", i, len(rule.Values), len(fltr.Rules[i].Values))
			} else {
				for j, value := range rule.Values {
					if fltr.Rules[i].Values[j] != value {
						t.Errorf("for rule %d, expected value %v, got %v", i, value, fltr.Rules[i].Values[j])
					}
				}
			}
		}
	}
	if fltr.ActivationInterval != nil && (fltr.ActivationInterval.ActivationTime != v1F.ActivationInterval.ActivationTime || fltr.ActivationInterval.ExpiryTime != v1F.ActivationInterval.ExpiryTime) {
		t.Errorf("expected ActivationInterval %+v, got %+v", v1F.ActivationInterval, fltr.ActivationInterval)
	}
}
