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
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fIdxCasesCfg     *config.CGRConfig
	fIdxCasesCfgPath string
	fIdxCasesCfgDir  string
	fIdxCasesRPC     *rpc.Client
	filter1          = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TEST1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.MetaString,
					Type:    "~Account",
					Values:  []string{"1001", "1002"},
				},
				{
					Element: utils.MetaExists,
					Type:    "~Destiantion",
					Values:  []string{},
				},
				{
					Element: utils.MetaLessThan,
					Type:    "~ProcessedField",
					Values:  []string{"5"},
				},
				{
					Element: utils.MetaSuffix,
					Type:    "~Contancts",
					Values:  []string{"455643", "99984", "123"},
				},
			},
		},
	}
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TEST2",
			Rules: []*engine.FilterRule{
				{
					Element: utils.MetaEmpty,
					Type:    "~Subject",
					Values:  []string{},
				},
				{
					Element: utils.MetaPrefix,
					Type:    "~Missed",
					Values:  []string{"1"},
				},
			},
		},
	}

	sTestsFilterIndexesCases = []func(t *testing.T){
		testFilterIndexesCasesITLoadConfig,
		testFilterIndexesCasesInitDataDb,
		testFilterIndexesCasesResetStorDb,
		testFilterIndexesCasesStartEngine,
		testFilterIndexesCasesRpcConn,

		testFilterIndexesCasesSetFilters,
		testFilterIndexesCasesSetAttributesWithFilters,
		testFilterIndexesCasesGetIndexes,

		testFilterIndexesCasesStopEngine,
	}
)

// Test start here
func TestFilterIndexesCasesIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		fIdxCasesCfgDir = "filter_indexes_cases_internal"
	case utils.MetaMySQL:
		fIdxCasesCfgDir = "filter_indexes_cases_mysql"
	case utils.MetaMongo:
		fIdxCasesCfgDir = "filter_indexes_cases_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFilterIndexesCases {
		t.Run(fIdxCasesCfgDir, stest)
	}
}

func testFilterIndexesCasesITLoadConfig(t *testing.T) {
	fIdxCasesCfgPath = path.Join(*dataDir, "conf", "samples", fIdxCasesCfgDir)
	var err error
	if fIdxCasesCfg, err = config.NewCGRConfigFromPath(fIdxCasesCfgPath); err != nil {
		t.Error(err)
	}
}

func testFilterIndexesCasesInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(fIdxCasesCfg); err != nil {
		t.Fatal(err)
	}
}

func testFilterIndexesCasesResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(fIdxCasesCfg); err != nil {
		t.Fatal(err)
	}
}

func testFilterIndexesCasesStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fIdxCasesCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testFilterIndexesCasesRpcConn(t *testing.T) {
	var err error
	fIdxCasesRPC, err = newRPCClient(fIdxCasesCfg.ListenCfg())
	if err != nil {
		t.Fatal("Could not connect: ", err.Error())
	}
}

func testFilterIndexesCasesSetFilters(t *testing.T) {
	//set two filters from above, those will be used in our profiles
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesSetAttributesWithFilters(t *testing.T) {
	eAttrPrf1 := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "ATTR_NO_FLTR",
			FilterIDs: []string{},
			Contexts:  []string{utils.MetaSessionS},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.CGRID,
					Value: config.NewRSRParsersMustCompile("test_generated_id", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20.0,
		},
	}
	eAttrPrf1.Compile()
	eAttrPrf2 := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "ATTR_NO_FLTR1",
			FilterIDs: []string{
				"FLTR_TEST1",
				"*rsr::~*req.Usage(>0s)"},
			Contexts: []string{},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Account,
					Value: config.NewRSRParsersMustCompile("SuccesfullID", true, utils.INFIELD_SEP),
				},
			},
			Weight: 30.0,
		},
	}
	eAttrPrf2.Compile()
	eAttrPrf3 := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "ATTR_NO_FLTR2",
			FilterIDs: []string{
				"FLTR_TEST2",
				"*string:~*req.Category:call"},
			Contexts: []string{utils.META_ANY},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Account,
					Value: config.NewRSRParsersMustCompile("SuccesfullID", true, utils.INFIELD_SEP),
				},
			},
			Weight: 5.0,
		},
	}
	eAttrPrf3.Compile()
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetAttributeProfile, eAttrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetAttributeProfile, eAttrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetAttributeProfile, eAttrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
		Context:  utils.META_ANY,
	}
	expectedIndexes := []string{
		"*string:~*req.Account:3001:ResProfile3",
		"*string:~*req.Destination:1001:ResProfile1",
		"*string:~*req.Destination:1001:ResProfile2",
		"*string:~*req.Destination:1001:ResProfile3",
		"*string:~*req.Account:1002:ResProfile1",
		"*string:~*req.Account:1002:ResProfile2",
		"*string:~*req.Account:1002:ResProfile3",
		"*string:~*req.Account:1003:ResProfile3",
		"*prefix:~*req.Destination:20:ResProfile1",
		"*prefix:~*req.Destination:20:ResProfile2",
		"*string:~*req.Account:1001:ResProfile1",
		"*string:~*req.Account:1001:ResProfile2",
		"*string:~*req.Account:2002:ResProfile2",
		"*prefix:~*req.Destination:1001:ResProfile3",
		"*prefix:~*req.Destination:200:ResProfile3",
		"*string:~*req.Destination:2001:ResProfile1",
		"*string:~*req.Destination:2001:ResProfile2",
		"*string:~*req.Destination:2001:ResProfile3",
		"*prefix:~*req.Account:10:ResProfile1",
		"*prefix:~*req.Account:10:ResProfile2",
		"*prefix:~*req.Account:10:ResProfile3"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
