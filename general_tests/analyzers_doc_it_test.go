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
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestAnzDocIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	jsonCfg := fmt.Sprintf(`{
"admins": {
	"enabled": true
},
"attributes": {
	"enabled": true,
	"prefix_indexed_fields": ["*req.Destination", "*req.Account"],
	"opts": {
		"*processRuns": [{
			"Tenant": "cgrates.org",
			"FilterIDs": [],
			"Value": 2
		}]
	}
},
"analyzers": {
	"enabled": true,
	"db_path": "%s"
}
}`, t.TempDir())
	ng := engine.TestEngine{
		ConfigJSON: jsonCfg,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   *utils.Encoding,
	}
	client, _ := ng.Run(t)

	status := func(t *testing.T) {
		t.Helper()
		var status map[string]any
		if err := client.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
			t.Error(err)
		}
	}

	//Generate traffic.
	status(t) //1*CoreSv1.Status

	time.Sleep(time.Second)
	timeVar := time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")

	setAttrProfiles(t, client)   // 5*AdminSv1.SetAttributeProfiles
	getAttrProfiles(t, client)   // 1*AdminSv1.GetAttributeProfiles
	attrProcessEvents(t, client) // 3*AttributeSv1.ProcessEvent
	status(t)                    //1*CoreSv1.Status (2 total)

	time.Sleep(100 * time.Millisecond)

	// HeaderFilters only queries.
	anzStringQuery(t, client, 5, `+RequestMethod:"AdminSv1.SetAttributeProfile"`)
	anzStringQuery(t, client, 1, `+RequestMethod:"AdminSv1.GetAttributeProfiles"`)
	anzStringQuery(t, client, 3, `+RequestMethod:"AttributeSv1.ProcessEvent"`)
	anzStringQuery(t, client, 2, `+RequestMethod:"CoreSv1.Status"`)
	anzStringQuery(t, client, 5, `+RequestMethod:"CacheSv1.ReloadCache"`)
	anzStringQuery(t, client, 11, `-RequestMethod:"CacheSv1.ReloadCache"`)
	anzStringQuery(t, client, 6, `+RequestMethod:"/AdminSv1.*/"`) // regex

	// Query results that happen before a given time.
	headerFltr := `+RequestStartTime:<"%s"`
	headerFltr = fmt.Sprintf(headerFltr, timeVar)
	anzStringQuery(t, client, 1, headerFltr)

	// ContentFilters only queries.
	anzStringQuery(t, client, 2, `+RequestID:<=2 -RequestMethod:"CacheSv1.ReloadCache"`)
	anzStringQuery(t, client, 5, "", "*string:~*hdr.RequestMethod:AdminSv1.SetAttributeProfile")
	anzStringQuery(t, client, 1, "", "*string:~*hdr.RequestMethod:AdminSv1.GetAttributeProfiles")
	anzStringQuery(t, client, 3, "", "*string:~*hdr.RequestMethod:AttributeSv1.ProcessEvent")
	anzStringQuery(t, client, 2, "", "*string:~*hdr.RequestMethod:CoreSv1.Status")
	anzStringQuery(t, client, 5, "", "*string:~*hdr.RequestMethod:CacheSv1.ReloadCache")
	anzStringQuery(t, client, 6, "", "*prefix:~*hdr.RequestMethod:AdminSv1.")

	// Query results that happen before a given time
	contentFltr := "*lt:~*hdr.RequestStartTime:%s"
	contentFltr = fmt.Sprintf(contentFltr, timeVar)
	anzStringQuery(t, client, 1, "", contentFltr)

	anzStringQuery(t, client, 7, "", "*lte:~*hdr.RequestID:2")
	anzStringQuery(t, client, -1, "", "*gt:~*hdr.RequestDuration:1ms")
	anzStringQuery(t, client, 1, `+RequestMethod:"AttributeSv1.ProcessEvent"`, "*notstring:~*rep.Event.Cost:0")
	anzStringQuery(t, client, 1, `+RequestMethod:"CoreSv1.Status"`, "*gt:~*rep.goroutines:55")
}

// anzStringQuery sends an AnalyzerSv1.StringQuery request. First filter represents
// the HeaderFilters parameter, while the rest are the ContentFilters. Checks if the
// result contains the expected amount of matches (wantRC).
func anzStringQuery(t *testing.T, client *birpc.Client, wantRC int, filters ...string) {
	t.Helper()
	var headerFilters string
	var contentFilters []string
	if len(filters) > 0 {
		headerFilters = filters[0]
		contentFilters = filters[1:]
	}
	var result []map[string]any
	if err := client.Call(context.Background(), utils.AnalyzerSv1StringQuery,
		&analyzers.QueryArgs{
			HeaderFilters:  headerFilters,
			ContentFilters: contentFilters,
		}, &result); err != nil {
		t.Error(err)
	} else if len(result) != wantRC && wantRC != -1 {
		t.Errorf("AnalyzerSv1.StringQuery: len(result)=%d, want %d\n%s", len(result), wantRC, utils.ToJSON(result))
	} else if wantRC == -1 && len(result) < 1 {
		t.Errorf("AnalyzerSv1.StringQuery: len(result)=%d, want >0\n%s", len(result), utils.ToJSON(result))
	}

}

func setAttrProfiles(t *testing.T, client *birpc.Client) {
	attributeProfiles := []*engine.APIAttributeProfileWithAPIOpts{
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "ATTR_1001",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						FilterIDs: []string{"*notexists:~*req.RequestType:"},
						Path:      "*req.RequestType",
						Type:      utils.MetaConstant,
						Value:     "*rated",
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "ATTR_1002",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						FilterIDs: []string{"*notexists:~*req.RequestType:"},
						Path:      "*req.RequestType",
						Type:      utils.MetaConstant,
						Value:     "*postpaid",
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "ATTR_1101",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.Account:1101"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						FilterIDs: []string{"*notexists:~*req.RequestType:"},
						Path:      "*req.RequestType",
						Type:      utils.MetaConstant,
						Value:     "*prepaid",
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "ATTR_1102",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.Account:1102"},
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						FilterIDs: []string{"*notexists:~*req.RequestType:"},
						Path:      "*req.RequestType",
						Type:      utils.MetaConstant,
						Value:     "*pseudoprepaid",
					},
				},
			},
		},
		{
			APIAttributeProfile: &engine.APIAttributeProfile{
				ID:        "ATTR_PRF_11",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*prefix:~*req.Account:11"},
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Attributes: []*engine.ExternalAttribute{
					{
						FilterIDs: []string{"*prefix:~*req.Destination:10"},
						Path:      "*req.Cost",
						Type:      utils.MetaConstant,
						Value:     "0",
					},
				},
			},
		},
	}
	var reply string
	for _, attributeProfile := range attributeProfiles {
		if err := client.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
			attributeProfile, &reply); err != nil {
			t.Error(err)
		}
	}
}

func getAttrProfiles(t *testing.T, client *birpc.Client) {
	expectedAttributeProfiles := []*engine.APIAttributeProfile{
		{
			ID:        "ATTR_1001",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					FilterIDs: []string{"*notexists:~*req.RequestType:"},
					Path:      "*req.RequestType",
					Type:      utils.MetaConstant,
					Value:     "*rated",
				},
			},
		},
		{
			ID:        "ATTR_1002",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					FilterIDs: []string{"*notexists:~*req.RequestType:"},
					Path:      "*req.RequestType",
					Type:      utils.MetaConstant,
					Value:     "*postpaid",
				},
			},
		},
		{
			ID:        "ATTR_1101",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Account:1101"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					FilterIDs: []string{"*notexists:~*req.RequestType:"},
					Path:      "*req.RequestType",
					Type:      utils.MetaConstant,
					Value:     "*prepaid",
				},
			},
		},
		{
			ID:        "ATTR_1102",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Account:1102"},
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					FilterIDs: []string{"*notexists:~*req.RequestType:"},
					Path:      "*req.RequestType",
					Type:      utils.MetaConstant,
					Value:     "*pseudoprepaid",
				},
			},
		},
		{
			ID:        "ATTR_PRF_11",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*prefix:~*req.Account:11"},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
			Attributes: []*engine.ExternalAttribute{
				{
					FilterIDs: []string{"*prefix:~*req.Destination:10"},
					Path:      "*req.Cost",
					Type:      utils.MetaConstant,
					Value:     "0",
				},
			},
		},
	}
	var replyAttributeProfiles []*engine.APIAttributeProfile
	if err := client.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyAttributeProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyAttributeProfiles, func(i, j int) bool {
			return replyAttributeProfiles[i].ID < replyAttributeProfiles[j].ID
		})
		if !reflect.DeepEqual(replyAttributeProfiles, expectedAttributeProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedAttributeProfiles), utils.ToJSON(replyAttributeProfiles))
		}
	}
}

func attrProcessEvents(t *testing.T, client *birpc.Client) {
	eventCall1001to1002 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "call1001to1002",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "abcdef",
			utils.OriginHost:   "192.168.1.1",
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
			utils.AnswerTime:   time.Unix(1383813748, 0).UTC(),
			utils.Usage:        10 * time.Second,
			utils.RunID:        utils.MetaDefault,
			utils.Cost:         1.01,
		},
	}

	eventCall1101to1002 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "call1011to1002",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "ghijkl",
			utils.OriginHost:   "192.168.1.1",
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.AccountField: "1101",
			utils.Subject:      "1101",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Unix(1383813748, 0).UTC(),
			utils.AnswerTime:   time.Unix(1383813755, 0).UTC(),
			utils.Usage:        15 * time.Second,
			utils.RunID:        utils.MetaDefault,
			utils.Cost:         1.51,
		},
	}
	eventCall1102to1001 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "call1102to1001",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "uvwxyz",
			utils.OriginHost:   "192.168.1.1",
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.AccountField: "1102",
			utils.Subject:      "1102",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Unix(1383813755, 0).UTC(),
			utils.AnswerTime:   time.Unix(1383813762, 0).UTC(),
			utils.Usage:        20 * time.Second,
			utils.RunID:        utils.MetaDefault,
			utils.Cost:         2.01,
		},
	}
	expectedReply := `{"AlteredFields":[{"MatchedProfileID":"cgrates.org:ATTR_1001","Fields":["*req.RequestType"]}],"Tenant":"cgrates.org","ID":"call1001to1002","Event":{"Account":"1001","AnswerTime":"2013-11-07T08:42:28Z","Category":"call","Cost":1.01,"Destination":"1002","OriginHost":"192.168.1.1","OriginID":"abcdef","RequestType":"*rated","RunID":"*default","SetupTime":"2013-11-07T08:42:25Z","Subject":"1001","Tenant":"cgrates.org","ToR":"*voice","Usage":10000000000},"APIOpts":{}}`
	var rplyEv engine.AttrSProcessEventReply
	if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		eventCall1001to1002, &rplyEv); err != nil {
		t.Error(err)
	} else if jsonReply := utils.ToJSON(rplyEv); jsonReply != expectedReply {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expectedReply, jsonReply)
	}
	expectedReply = `{"AlteredFields":[{"MatchedProfileID":"cgrates.org:ATTR_PRF_11","Fields":["*req.Cost"]},{"MatchedProfileID":"cgrates.org:ATTR_1101","Fields":["*req.RequestType"]}],"Tenant":"cgrates.org","ID":"call1011to1002","Event":{"Account":"1101","AnswerTime":"2013-11-07T08:42:35Z","Category":"call","Cost":"0","Destination":"1002","OriginHost":"192.168.1.1","OriginID":"ghijkl","RequestType":"*prepaid","RunID":"*default","SetupTime":"2013-11-07T08:42:28Z","Subject":"1101","Tenant":"cgrates.org","ToR":"*voice","Usage":15000000000},"APIOpts":{}}`
	if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		eventCall1101to1002, &rplyEv); err != nil {
		t.Error(err)
	} else if jsonReply := utils.ToJSON(rplyEv); jsonReply != expectedReply {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expectedReply, jsonReply)
	}
	expectedReply = `{"AlteredFields":[{"MatchedProfileID":"cgrates.org:ATTR_PRF_11","Fields":["*req.Cost"]},{"MatchedProfileID":"cgrates.org:ATTR_1102","Fields":["*req.RequestType"]}],"Tenant":"cgrates.org","ID":"call1102to1001","Event":{"Account":"1102","AnswerTime":"2013-11-07T08:42:42Z","Category":"call","Cost":"0","Destination":"1001","OriginHost":"192.168.1.1","OriginID":"uvwxyz","RequestType":"*pseudoprepaid","RunID":"*default","SetupTime":"2013-11-07T08:42:35Z","Subject":"1102","Tenant":"cgrates.org","ToR":"*voice","Usage":20000000000},"APIOpts":{}}`
	if err := client.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		eventCall1102to1001, &rplyEv); err != nil {
		t.Error(err)
	} else if jsonReply := utils.ToJSON(rplyEv); jsonReply != expectedReply {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expectedReply, jsonReply)
	}
}
