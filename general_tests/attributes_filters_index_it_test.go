//go:build integration
// +build integration

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

package general_tests

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAttributeFilterIndexing(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaPostgres:
		dbCfg = engine.PostgresDBCfg
	case utils.MetaMongo:
		dbCfg = engine.MongoDBCfg
	case utils.MetaMySQL:
	default:
		t.Fatal("Unknown Database type")
	}

	jsonCfg := `{
        "general": {
            "log_level": 7
        },
        "stor_db": {
            "db_password": "CGRateS.org"
        },
        "apiers": {
            "enabled": true
        },
        "attributes": {
            "enabled": true,
            "prefix_indexed_fields": ["*req.Subject","*req.Account"],
            "suffix_indexed_fields": ["*req.Subject","*req.Account"],
            "string_indexed_fields": ["*req.Subject","*req.Account"],
            "exists_indexed_fields": ["*req.Subject","*req.Account"]
        }
    }`

	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		DBCfg:      dbCfg,
	}
	client, _ := ng.Run(t)

	// Set filter with value 48
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					Element: "~*req.Subject",
					Type:    "*prefix",
					Values:  []string{"48"},
				},
				{
					Element: "~*req.Subject",
					Type:    "*suffix",
					Values:  []string{"48"},
				},
				{
					Element: "~*req.Subject",
					Type:    "*exists",
				},
				{
					Element: "~*req.Subject",
					Type:    "*string",
					Values:  []string{"48"},
				},
				{
					Element: "~*req.Account",
					Type:    "*niprefix",
					Values:  []string{"48"},
				},
				{
					Element: "~*req.Account",
					Type:    "*nisuffix",
					Values:  []string{"48"},
				},
				{
					Element: "~*req.Account",
					Type:    "*nistring",
					Values:  []string{"48"},
				},
				{
					Element: "~*req.Account",
					Type:    "*niexists",
				},
			},
		},
	}
	var result string
	if err := client.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var indexes []string
	if err := client.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix,
		Context: utils.MetaSessionS},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// Set attribute profile using this filter
	attrProf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ApierTest",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_1"},
			Attributes: []*engine.Attribute{{
				Path:  "*req.FL1",
				Value: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			}},
			Weight: 20,
		},
	}
	if err := client.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrProf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	// Event should not match (prefix 44 vs 48)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			"Subject": "44",
			"Account": "44",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	// Check filter indexes (should exist now with prefix 48)
	expIdx := []string{
		"*exists:*req.Subject:ApierTest",
		"*prefix:*req.Subject:48:ApierTest",
		"*string:*req.Subject:48:ApierTest",
		"*suffix:*req.Subject:48:ApierTest",
	}
	if err := client.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org",
		Context: utils.MetaSessionS},
		&indexes); err != nil {
		t.Error(err)
	} else {
		slices.Sort(indexes)
		if !reflect.DeepEqual(indexes, expIdx) {
			t.Errorf("Expecting: %+v, received: %+v",
				utils.ToJSON(expIdx), utils.ToJSON(indexes))
		}
	}

	// Update filter to match prefix 44

	for _, rule := range filter.Filter.Rules {
		if strings.HasSuffix(rule.Type, utils.MetaExists[1:]) {
			continue
		}
		rule.Values[0] = "44"

	}
	if err := client.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	// Now matching event should work
	exp := engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ApierTest"},
		AlteredFields:   []string{"*req.FL1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				"Subject": "44",
				"Account": "44",
				"FL1":     "Al1",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(exp, rplyEv) {
		t.Errorf("Expected %s, received %s", utils.ToJSON(exp), utils.ToJSON(rplyEv))
	}

	// Check updated filter indexes (should now show prefix 44)
	expIdx = []string{
		"*prefix:*req.Subject:44:ApierTest",
	}
	if err := client.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix,
		Context: utils.MetaSessionS},
		&indexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(indexes, expIdx) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(expIdx), utils.ToJSON(indexes))
	}

	// Remove attribute profile
	if err := client.Call(context.Background(), utils.APIerSv1RemoveAttributeProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := client.Call(context.Background(), utils.APIerSv1RemoveFilter, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "FLTR_1"}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := client.Call(context.Background(), utils.APIerSv1GetFilterIndexes, &v1.AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: "cgrates.org", FilterType: utils.MetaPrefix,
		Context: utils.MetaSessionS},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
