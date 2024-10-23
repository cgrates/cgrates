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
	"os"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/analyzers"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	anzDocCfgPath string
	anzDocCfg     *config.CGRConfig
	anzDocRPC     *birpc.Client
	timeVar       string // will be used to query results based on date

	sTestsAnzDoc = []func(t *testing.T){
		testAnzDocInitCfg,
		testAnzDocFlushDBs,
		testAnzDocStartEngine,
		testAnzDocRPCConn,

		/*
			Generate traffic. The following API methods will be stored in the blevesearch db:
			- 2 * CoreSv1.Status
			- 5 * AttributeSv1.SetAttributeProfile
			- 1 * AttributeSv1.GetAttributeProfiles
			- 3 * AttributeSv1.ProcessEvent
			- ? * CacheSv1.ReloadCache
		*/
		testAnzDocCoreSStatus,
		testPopulateTimeVariable,
		testAnzDocSetAttributeProfiles,
		testAnzDocCheckAttributeProfiles,
		testAnzDocAttributeSProcessEvent,
		testAnzDocCoreSStatus,
		// make queries to the AnalyzerS db using only HeaderFilters
		testAnzDocQueryWithHeaderFilters,
		// make queries to the AnalyzerS db using only ContentFilters
		testAnzDocQueryWithContentFiltersFilters,
		// make queries to the AnalyzerS db using a combination of both types of filters
		testAnzDocQuery,
		testAnzDocKillEngine,
	}
)

func TestAnzDocIT(t *testing.T) {
	for _, stest := range sTestsAnzDoc {
		t.Run("TestAnzDocIT", stest)
	}
}

func testAnzDocFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(anzDocCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(anzDocCfg); err != nil {
		t.Fatal(err)
	}
}

func testAnzDocStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(anzDocCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAnzDocInitCfg(t *testing.T) {
	var err error
	if err := os.RemoveAll("/tmp/analyzers/"); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll("/tmp/analyzers/", 0700); err != nil {
		t.Fatal(err)
	}
	anzDocCfgPath = path.Join(*utils.DataDir, "conf", "samples", "analyzers_doc")
	anzDocCfg, err = config.NewCGRConfigFromPath(context.Background(), anzDocCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAnzDocRPCConn(t *testing.T) {
	anzDocRPC = engine.NewRPCClient(t, anzDocCfg.ListenCfg(), *utils.Encoding)
}

func testAnzDocKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll(anzDocCfg.AnalyzerSCfg().DBPath); err != nil {
		t.Fatal(err)
	}
}

func testAnzDocCoreSStatus(t *testing.T) {
	var status map[string]any
	if err := anzDocRPC.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		return
	}
}

func testPopulateTimeVariable(t *testing.T) {
	time.Sleep(time.Second)
	timeVar = time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
}

func testAnzDocSetAttributeProfiles(t *testing.T) {
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
		if err := anzDocRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
			attributeProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testAnzDocCheckAttributeProfiles(t *testing.T) {
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
	if err := anzDocRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfiles,
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

func testAnzDocAttributeSProcessEvent(t *testing.T) {
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
	if err := anzDocRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		eventCall1001to1002, &rplyEv); err != nil {
		t.Error(err)
	} else if jsonReply := utils.ToJSON(rplyEv); jsonReply != expectedReply {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expectedReply, jsonReply)
	}
	expectedReply = `{"AlteredFields":[{"MatchedProfileID":"cgrates.org:ATTR_PRF_11","Fields":["*req.Cost"]},{"MatchedProfileID":"cgrates.org:ATTR_1101","Fields":["*req.RequestType"]}],"Tenant":"cgrates.org","ID":"call1011to1002","Event":{"Account":"1101","AnswerTime":"2013-11-07T08:42:35Z","Category":"call","Cost":"0","Destination":"1002","OriginHost":"192.168.1.1","OriginID":"ghijkl","RequestType":"*prepaid","RunID":"*default","SetupTime":"2013-11-07T08:42:28Z","Subject":"1101","Tenant":"cgrates.org","ToR":"*voice","Usage":15000000000},"APIOpts":{}}`
	if err := anzDocRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		eventCall1101to1002, &rplyEv); err != nil {
		t.Error(err)
	} else if jsonReply := utils.ToJSON(rplyEv); jsonReply != expectedReply {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expectedReply, jsonReply)
	}
	expectedReply = `{"AlteredFields":[{"MatchedProfileID":"cgrates.org:ATTR_PRF_11","Fields":["*req.Cost"]},{"MatchedProfileID":"cgrates.org:ATTR_1102","Fields":["*req.RequestType"]}],"Tenant":"cgrates.org","ID":"call1102to1001","Event":{"Account":"1102","AnswerTime":"2013-11-07T08:42:42Z","Category":"call","Cost":"0","Destination":"1001","OriginHost":"192.168.1.1","OriginID":"uvwxyz","RequestType":"*pseudoprepaid","RunID":"*default","SetupTime":"2013-11-07T08:42:35Z","Subject":"1102","Tenant":"cgrates.org","ToR":"*voice","Usage":20000000000},"APIOpts":{}}`
	if err := anzDocRPC.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		eventCall1102to1001, &rplyEv); err != nil {
		t.Error(err)
	} else if jsonReply := utils.ToJSON(rplyEv); jsonReply != expectedReply {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			expectedReply, jsonReply)
	}

}

func testAnzDocQueryWithHeaderFilters(t *testing.T) {
	time.Sleep(500 * time.Millisecond)
	var result []map[string]any

	// Query results for the AdminSv1.SetAttributeProfile request method using HeaderFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"AdminSv1.SetAttributeProfile"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 5 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the AdminSv1.GetAttributeProfiles request method using HeaderFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"AdminSv1.GetAttributeProfiles"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the AttributeSv1.ProcessEvent request method using HeaderFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"AttributeSv1.ProcessEvent"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 3 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the CoreSv1.Status request method using HeaderFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"CoreSv1.Status"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 2 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the CacheSv1.ReloadCache request method using HeaderFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"CacheSv1.ReloadCache"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 5 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for all the request methods except CacheSv1.ReloadCache using HeaderFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `-RequestMethod:"CacheSv1.ReloadCache"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 11 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query all results for the requests made to the AdminS service using regular expressions
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"/AdminSv1.*/"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 6 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results that happen before a given time
	headerFltr := `+RequestStartTime:<"%s"`
	headerFltr = fmt.Sprintf(headerFltr, timeVar)
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  headerFltr,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for API calls with RequestID smaller or equal to 2
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestID:<=2 -RequestMethod:"CacheSv1.ReloadCache"`,
		ContentFilters: []string{},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 2 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnzDocQueryWithContentFiltersFilters(t *testing.T) {
	var result []map[string]any
	// Query results for the AdminSv1.SetAttributeProfile request method using ContentFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*string:~*hdr.RequestMethod:AdminSv1.SetAttributeProfile"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 5 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the AdminSv1.GetAttributeProfile request method using ContentFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*string:~*hdr.RequestMethod:AdminSv1.GetAttributeProfiles"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the AttributeSv1.ProcessEvent request method using ContentFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*string:~*hdr.RequestMethod:AttributeSv1.ProcessEvent"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 3 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the CoreSv1.Status request method using ContentFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*string:~*hdr.RequestMethod:CoreSv1.Status"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 2 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for the CacheSv1.ReloadCache request method using ContentFilters
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*string:~*hdr.RequestMethod:CacheSv1.ReloadCache"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 5 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query all results for the requests made to the AdminS service using prefix filter
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*prefix:~*hdr.RequestMethod:AdminSv1."},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 6 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results that happen before a given time
	contentFltr := "*lt:~*hdr.RequestStartTime:%s"
	contentFltr = fmt.Sprintf(contentFltr, timeVar)
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{contentFltr},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for API calls with RequestID smaller or equal to 2
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*lte:~*hdr.RequestID:2"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 7 {
		fmt.Println(len(result))
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Query results for API calls with with an execution duration longer than 30ms
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  "",
		ContentFilters: []string{"*gt:~*hdr.RequestDuration:10ms"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) == 0 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}

func testAnzDocQuery(t *testing.T) {
	var result []map[string]any

	// Get results for AttributeSv1.ProcessEvent request replies whose events have a non-null cost
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"AttributeSv1.ProcessEvent"`,
		ContentFilters: []string{"*notstring:~*rep.Event.Cost:0"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}

	// Get results for CoreSv1.Status request replies that state a higher number of goroutines than 46
	if err := anzDocRPC.Call(context.Background(), utils.AnalyzerSv1StringQuery, &analyzers.QueryArgs{
		HeaderFilters:  `+RequestMethod:"CoreSv1.Status"`,
		ContentFilters: []string{"*gt:~*rep.ActiveGoroutines:46"},
	}, &result); err != nil {
		t.Error(err)
	} else if len(result) != 1 {
		t.Errorf("Unexpected result: %s", utils.ToJSON(result))
	}
}
