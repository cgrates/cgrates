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
			t.Errorf("Expected: %s, recived: %s", m.exp, rply)
		}
	}

}

func TestFiltersMigrate(t *testing.T) {
	data := []struct{ in, exp *engine.Filter }{
		{
			in: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_1",
				Rules: []*engine.FilterRule{
					&engine.FilterRule{
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
					&engine.FilterRule{
						Type:      utils.MetaString,
						FieldName: "~*req.Account",
						Values:    []string{},
					},
				},
			},
		},
		{
			in: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_2",
				Rules: []*engine.FilterRule{
					&engine.FilterRule{
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
					&engine.FilterRule{
						Type:      utils.MetaPrefix,
						FieldName: "~Account",
						Values:    []string{},
					},
				},
			},
		},
	}
	for _, m := range data {
		if rply := migrateFilterV1(m.in); !reflect.DeepEqual(rply, m.exp) {
			t.Errorf("Expected: %s, recived: %s", utils.ToJSON(m.exp), utils.ToJSON(rply))
		}
	}
}

func TestFiltersMigrateV2(t *testing.T) {
	data := []struct{ in, exp *engine.Filter }{
		{
			in: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_1",
				Rules: []*engine.FilterRule{
					&engine.FilterRule{
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
					&engine.FilterRule{
						Type:      utils.MetaString,
						FieldName: "~*req.Account",
						Values:    []string{},
					},
				},
			},
		},
		{
			in: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_2",
				Rules: []*engine.FilterRule{
					&engine.FilterRule{
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
					&engine.FilterRule{
						Type:      utils.MetaPrefix,
						FieldName: "~*req.Account",
						Values:    []string{},
					},
				},
			},
		},
		{
			in: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_3",
				Rules: []*engine.FilterRule{
					&engine.FilterRule{
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
					&engine.FilterRule{
						Type:      utils.MetaPrefix,
						FieldName: "~*act.Account",
						Values:    []string{},
					},
				},
			},
		},
		{
			in: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_4",
				Rules: []*engine.FilterRule{
					&engine.FilterRule{
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
					&engine.FilterRule{
						Type:      utils.MetaPrefix,
						FieldName: "~*act.Account",
						Values:    []string{},
					},
				},
			},
		},
		{
			in: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FLTR_5",
				Rules: []*engine.FilterRule{
					&engine.FilterRule{
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
					&engine.FilterRule{
						Type:      utils.MetaPrefix,
						FieldName: "~*vars.Account",
						Values:    []string{},
					},
				},
			},
		},
	}
	for _, m := range data {
		if rply := migrateFilterV2(m.in); !reflect.DeepEqual(rply, m.exp) {
			t.Errorf("Expected: %s, recived: %s", utils.ToJSON(m.exp), utils.ToJSON(rply))
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
			t.Errorf("Expected: %s, recived: %s", m.exp, rply)
		}
	}

}
