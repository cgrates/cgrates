// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
						Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
						Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
						Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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

func TestMigrateInlineFilterV4(t *testing.T) {
	flts := []string{
		"*string:*~req.Account:1001",
		"*rsr::*~req.Destination",
		"*notrsr::*~req.MaxUsage(<0);*req.Account(^10)",
	}
	exp := []string{
		"*string:*~req.Account:1001",
		"*notrsr:*~req.MaxUsage:<0",
		"*notrsr:*req.Account:^10",
	}
	if rply, err := migrateInlineFilterV4(flts); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %s,received: %s", exp, rply)
	}

	flts = []string{
		"*string:*~req.Account:1001",
		"*rsr::*~req.Destination)",
	}
	if _, err := migrateInlineFilterV4(flts); err == nil {
		t.Error("Expected error received none")
	}
	flts = []string{
		"*rsr::*~req.Destination{*(1001)",
	}
	if _, err := migrateInlineFilterV4(flts); err == nil {
		t.Error("Expected error received none")
	}
}

func TestMigrateRequestFilterV4(t *testing.T) {
	flt := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLT_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
			{
				Type:    utils.MetaRSR,
				Element: utils.EmptyString,
				Values:  []string{"~*req.Account"},
			},
			{
				Type:    utils.MetaRSR,
				Element: utils.EmptyString,
				Values:  []string{"~*req.Destination(^1001&1$)"},
			},
		},
	}
	exp := &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLT_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Account",
				Values:  []string{"1001"},
			},
			{
				Type:    utils.MetaRSR,
				Element: "~*req.Destination",
				Values:  []string{"^1001", "1$"},
			},
		},
	}
	m := new(Migrator)
	if rply, err := m.migrateRequestFilterV4(flt); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected: %s,received: %s", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	flt = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLT_1",
		Rules: []*engine.FilterRule{
			{
				Type:    utils.MetaRSR,
				Element: utils.EmptyString,
				Values:  []string{"~*req.Destination^1001&1$)"},
			},
		},
	}

	if _, err := m.migrateRequestFilterV4(flt); err == nil {
		t.Error("Expected error received none")
	}
}
