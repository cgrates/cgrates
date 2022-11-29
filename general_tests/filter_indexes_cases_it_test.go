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
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001", "1002"},
				},
				{
					Type:    utils.MetaExists,
					Element: "~*req.Destiantion",
					Values:  []string{},
				},
				{
					Type:    utils.MetaLessThan,
					Element: "~*req.ProcessedField",
					Values:  []string{"5"},
				},
				{
					Type:    utils.MetaSuffix,
					Element: "~*req.Contacts",
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
					Type:    utils.MetaEmpty,
					Element: "~*req.Subject",
					Values:  []string{},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Missed",
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
		testFilterIndexesCasesGetIndexesAnyContext,
		testFilterIndexesCasesGetIndexesSessionsContext,

		testFilterIndexesCasesSetDifferentFilters,
		testFilterIndexesCasesOverwriteAttributes,
		testFilterIndexesCasesGetIndexesAnyContextChanged,
		testFilterIndexesCasesGetIndexesSessionsContextChanged,

		testFilterIndexesCasesSetIndexedFilter,
		testFilterIndexesCasesSetChargerWithFltr,
		testFilterIndexesCasesGetChargerIndexes,
		testFilterIndexesCasesOverwriteFilterForCharger,
		testFilterIndexesCasesComputeChargersIndexes,
		testFilterIndexesCasesGetChargerIndexesChanged,

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
			Contexts:  []string{utils.MetaSessionS, utils.META_ANY},
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
				"*lt:~*req.Usage:10s",
				"*string:~*req.Category:call",
			},
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

func testFilterIndexesCasesGetIndexesAnyContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
		Context:  utils.META_ANY,
	}
	expectedIndexes := []string{
		"*none:*any:*any:ATTR_NO_FLTR",
		"*prefix:~*req.Missed:1:ATTR_NO_FLTR2",
		"*string:~*req.Account:1001:ATTR_NO_FLTR1",
		"*string:~*req.Account:1002:ATTR_NO_FLTR1",
		"*string:~*req.Category:call:ATTR_NO_FLTR2"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetIndexesSessionsContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
		Context:  utils.MetaSessionS,
	}
	expectedIndexes := []string{
		"*none:*any:*any:ATTR_NO_FLTR"}
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesSetDifferentFilters(t *testing.T) {
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TEST1",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Cost",
					Values:  []string{"10", "15", "210"},
				},
				{
					Type:    utils.MetaExists,
					Element: "~*req.Usage",
					Values:  []string{},
				},
				{
					Type:    utils.MetaSuffix,
					Element: "~*req.AnswerTime",
					Values:  []string{"202"},
				},
			},
		},
	}
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesComputeAttributesIndexes(t *testing.T) {
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:     "cgrates.org",
			Context:    utils.META_ANY,
			AttributeS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
}

func testFilterIndexesCasesOverwriteAttributes(t *testing.T) {
	eAttrPrf1 := &v1.AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant: "cgrates.org",
			ID:     "ATTR_NO_FLTR",
			FilterIDs: []string{
				"FLTR_TEST1",
				"*destinations:~*req.Destination:+443",
			},
			Contexts: []string{},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Destination,
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
				"*none"},
			Contexts: []string{utils.MetaSessionS, utils.META_ANY},
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
				"*rsr::~*req.Tenant(~^cgr.*\\.org$)",
			},
			Contexts: []string{utils.META_ANY},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.Account,
					Value: config.NewRSRParsersMustCompile("ChangedFIlter", true, utils.INFIELD_SEP),
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

func testFilterIndexesCasesGetIndexesAnyContextChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
		Context:  utils.META_ANY,
	}
	expectedIndexes := []string{
		"*prefix:~*req.Cost:10:ATTR_NO_FLTR",
		"*prefix:~*req.Cost:15:ATTR_NO_FLTR",
		"*prefix:~*req.Cost:210:ATTR_NO_FLTR",
		"*none:*any:*any:ATTR_NO_FLTR1",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetIndexesSessionsContextChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaAttributes,
		Context:  utils.MetaSessionS,
	}
	expectedIndexes := []string{
		"*none:*any:*any:ATTR_NO_FLTR1"}
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesSetIndexedFilter(t *testing.T) {
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.CGRID",
					Values:  []string{"tester_id"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.AnswerTime",
					Values:  []string{"2022"},
				},
				{
					Type:    utils.MetaSuffix,
					Element: "~*req.AnswerTime",
					Values:  []string{"202"},
				},
			},
		},
	}
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesSetChargerWithFltr(t *testing.T) {
	chargerProfile := &v1.ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "ChrgerIndexable",
			FilterIDs:    []string{"FLTR_Charger"},
			RunID:        utils.MetaRaw,
			AttributeIDs: []string{"ATTR_FLTR1"},
			Weight:       20,
		},
	}
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetChargerIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaChargers,
	}
	expectedIndexes := []string{
		"*string:~*req.CGRID:tester_id:ChrgerIndexable",
		"*prefix:~*req.AnswerTime:2022:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteFilterForCharger(t *testing.T) {
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"12345"},
				},
				{
					Type:    utils.MetaLessOrEqual,
					Element: "~*req.ProcessRuns",
					Values:  []string{"1"},
				},
			},
		},
	}
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesComputeChargersIndexes(t *testing.T) {
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:   "cgrates.org",
			ChargerS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
}

func testFilterIndexesCasesGetChargerIndexesChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaChargers,
	}
	expectedIndexes := []string{
		"*string:~*req.Account:12345:ChrgerIndexable",
	}
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
