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
	"time"

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
		testFilterIndexesCasesComputeAttributesIndexes,
		testFilterIndexesCasesGetIndexesAnyContextChanged,
		testFilterIndexesCasesGetIndexesSessionsContextChanged,

		// CHARGERS
		testFilterIndexesCasesSetIndexedFilter,
		testFilterIndexesCasesSetChargerWithFltr,
		testFilterIndexesCasesGetChargerIndexes,
		testFilterIndexesCasesOverwriteFilterForCharger,
		testFilterIndexesCasesGetChargerIndexesChanged,

		testFilterIndexesCasesGetReverseFilterIndexes, // for chargers
		testFilterIndexesCasesRemoveChargerProfile,
		testFilterIndexesCasesGetIndexesAfterRemove,
		testFilterIndexesCasesGetReverseIndexesAfterRemove,

		// THRESHOLDS
		testFilterIndexesCasesSetThresholdWithFltr,
		testFilterIndexesCasesGetThresholdsIndexes,
		testFilterIndexesCasesOverwriteFilterForThresholds,
		testFilterIndexesCasesGetThresholdsIndexesChanged,

		testFilterIndexesCasesGetReverseFilterIndexes2,
		testFilterIndexesCasesRemoveThresholdsProfile,
		testFilterIndexesCasesGetIndexesAfterRemove2,
		testFilterIndexesCasesGetReverseIndexesAfterRemove2,

		// RESOURCES
		testFilterIndexesCasesSetResourceWithFltr,
		testFilterIndexesCasesGetResourcesIndexes,
		/*
			testFilterIndexesCasesOverwriteFilterForResources,
			testFilterIndexesCasesGetResourcesIndexesChanged,


			testFilterIndexesCasesGetReverseFilterIndexes3,
			testFilterIndexesCasesRemoveResourcesProfile,
			testFilterIndexesCasesGetIndexesAfterRemove3,
			testFilterIndexesCasesGetReverseIndexesAfterRemove3, */

		// SUPPLIER

		// DISPATCHER

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
		"*string:~*req.Category:call:ATTR_NO_FLTR2",
	}
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
					Element: "~*req.Account",
					Values:  []string{"02"},
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

	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger12312",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Destination",
					Values:  []string{"1443"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.SetupTime",
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
	filter2 := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger4564",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.RequestType",
					Values:  []string{"*none"},
				},
				{
					Type:    utils.MetaSuffix,
					Element: "~*req.AnswerTime",
					Values:  []string{"212"},
				},
			},
		},
	}
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

func testFilterIndexesCasesSetChargerWithFltr(t *testing.T) {
	chargerProfile := &v1.ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "ChrgerIndexable",
			FilterIDs:    []string{"FLTR_Charger", "FLTR_Charger12312", "FLTR_Charger4564"},
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
	chargerProfile =
		&v1.ChargerWithCache{
			ChargerProfile: &engine.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "ChrgerIndexable222",
				FilterIDs:    []string{"FLTR_Charger", "FLTR_Charger4564"},
				RunID:        utils.MetaRaw,
				AttributeIDs: []string{"ATTR_FLTR1"},
				Weight:       20,
			},
		}
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
		"*prefix:~*req.AnswerTime:2022:ChrgerIndexable",
		"*prefix:~*req.AnswerTime:2022:ChrgerIndexable222",
		"*prefix:~*req.SetupTime:2022:ChrgerIndexable",
		"*string:~*req.CGRID:tester_id:ChrgerIndexable",
		"*string:~*req.CGRID:tester_id:ChrgerIndexable222",
		"*string:~*req.Destination:1443:ChrgerIndexable",
		"*string:~*req.RequestType:*none:ChrgerIndexable",
		"*string:~*req.RequestType:*none:ChrgerIndexable222",
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

func testFilterIndexesCasesGetChargerIndexesChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaChargers,
	}
	expectedIndexes := []string{
		"*string:~*req.Account:12345:ChrgerIndexable",
		"*string:~*req.Account:12345:ChrgerIndexable222",
		"*prefix:~*req.SetupTime:2022:ChrgerIndexable",
		"*string:~*req.Destination:1443:ChrgerIndexable",
		"*string:~*req.RequestType:*none:ChrgerIndexable",
		"*string:~*req.RequestType:*none:ChrgerIndexable222",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseFilterIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*charger_filter_indexes:ChrgerIndexable222",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger4564",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*charger_filter_indexes:ChrgerIndexable222",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger12312",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesRemoveChargerProfile(t *testing.T) {
	var resp string
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ChrgerIndexable222"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testFilterIndexesCasesGetIndexesAfterRemove(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaChargers,
	}
	expectedIndexes := []string{
		"*string:~*req.Account:12345:ChrgerIndexable",
		"*prefix:~*req.SetupTime:2022:ChrgerIndexable",
		"*string:~*req.Destination:1443:ChrgerIndexable",
		"*string:~*req.RequestType:*none:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseIndexesAfterRemove(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger4564",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger12312",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesSetThresholdWithFltr(t *testing.T) {
	tPrfl1 := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"FLTR_Charger"},
			MaxHits:   1,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    10.0,
			Async:     true,
		},
	}
	tPrfl2 := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "TEST_PROFILE2",
			FilterIDs: []string{
				"FLTR_Charger4564",
				"FLTR_Charger",
				"FLTR_Charger12312",
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:  1,
			MinSleep: time.Duration(5 * time.Minute),
			Blocker:  false,
			Weight:   20.0,
			Async:    true,
		},
	}
	tPrfl3 := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "TEST_PROFILE3",
			FilterIDs: []string{
				"FLTR_Charger12312",
				"*prefix:~*req.Cost:4",
				"FLTR_Charger",
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:  1,
			MinSleep: time.Duration(5 * time.Minute),
			Blocker:  false,
			Weight:   40.0,
			Async:    true,
		},
	}
	tPrfl4 := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE4",
			FilterIDs: []string{},
			MaxHits:   1,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    5.0,
			Async:     true,
		},
	}
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl4, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetThresholdsIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaThresholds,
	}
	expectedIndexes := []string{
		// TEST_PROFILE1"
		"*string:~*req.Account:12345:TEST_PROFILE1",
		// TEST_PROFILE2"
		"*string:~*req.Account:12345:TEST_PROFILE2",
		"*string:~*req.RequestType:*none:TEST_PROFILE2",
		"*string:~*req.Destination:1443:TEST_PROFILE2",
		"*prefix:~*req.SetupTime:2022:TEST_PROFILE2",
		// TEST_PROFILE3"
		"*string:~*req.Account:12345:TEST_PROFILE3",
		"*prefix:~*req.Cost:4:TEST_PROFILE3",
		"*string:~*req.Destination:1443:TEST_PROFILE3",
		"*prefix:~*req.SetupTime:2022:TEST_PROFILE3",
		// TEST_PROFILE4"
		"*none:*any:*any:TEST_PROFILE4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteFilterForThresholds(t *testing.T) {
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.CostRefunded",
					Values:  []string{"12345"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*req.ToR",
					Values:  []string{"*voice"},
				},
				{
					Type:    utils.MetaLessOrEqual,
					Element: "~*req.ProcessRuns",
					Values:  []string{"1"},
				},
			},
		},
	}
	filter2 := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger4564",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Increment",
					Values:  []string{"1s"},
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
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetThresholdsIndexesChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaThresholds,
	}
	expectedIndexes := []string{
		// TEST_PROFILE1"
		"*string:~*req.ToR:*voice:TEST_PROFILE1",
		"*prefix:~*req.CostRefunded:12345:TEST_PROFILE1",
		// TEST_PROFILE2"
		"*string:~*req.ToR:*voice:TEST_PROFILE2",
		"*prefix:~*req.CostRefunded:12345:TEST_PROFILE2",
		"*string:~*req.Increment:1s:TEST_PROFILE2",
		"*string:~*req.Destination:1443:TEST_PROFILE2",
		"*prefix:~*req.SetupTime:2022:TEST_PROFILE2",
		// TEST_PROFILE3"
		"*string:~*req.ToR:*voice:TEST_PROFILE3",
		"*prefix:~*req.CostRefunded:12345:TEST_PROFILE3",
		"*prefix:~*req.Cost:4:TEST_PROFILE3",
		"*string:~*req.Destination:1443:TEST_PROFILE3",
		"*prefix:~*req.SetupTime:2022:TEST_PROFILE3",
		// TEST_PROFILE4"
		"*none:*any:*any:TEST_PROFILE4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseFilterIndexes2(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*threshold_filter_indexes:TEST_PROFILE1",
		"*threshold_filter_indexes:TEST_PROFILE2",
		"*threshold_filter_indexes:TEST_PROFILE3",
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger4564",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*threshold_filter_indexes:TEST_PROFILE2",
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger12312",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*threshold_filter_indexes:TEST_PROFILE2",
		"*threshold_filter_indexes:TEST_PROFILE3",
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesRemoveThresholdsProfile(t *testing.T) {
	var resp string
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "TEST_PROFILE3"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testFilterIndexesCasesGetIndexesAfterRemove2(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaThresholds,
	}
	expectedIndexes := []string{
		// TEST_PROFILE2"
		"*string:~*req.ToR:*voice:TEST_PROFILE2",
		"*prefix:~*req.CostRefunded:12345:TEST_PROFILE2",
		"*string:~*req.Increment:1s:TEST_PROFILE2",
		"*string:~*req.Destination:1443:TEST_PROFILE2",
		"*prefix:~*req.SetupTime:2022:TEST_PROFILE2",
		// TEST_PROFILE4"
		"*none:*any:*any:TEST_PROFILE4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseIndexesAfterRemove2(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*threshold_filter_indexes:TEST_PROFILE2",
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger4564",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*threshold_filter_indexes:TEST_PROFILE2",
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger12312",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes = []string{
		"*threshold_filter_indexes:TEST_PROFILE2",
		"*charger_filter_indexes:ChrgerIndexable",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesSetResourceWithFltr(t *testing.T) {
	rsPrf1 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "RESOURCE_FLTR1",
			FilterIDs: []string{
				"FLTR_Charger",
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(1) * time.Nanosecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            20,
		},
	}
	rsPrf2 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "RESOURCE_FLTR2",
			FilterIDs: []string{
				"FLTR_Charger4564",
				"FLTR_Charger",
				"FLTR_Charger12312"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(10) * time.Nanosecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            50,
		},
	}
	rsPrf3 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "RESOURCE_FLTR3",
			FilterIDs: []string{
				"FLTR_Charger12312",
				"*prefix:~*req.Usage:15s",
				"FLTR_Charger",
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(5) * time.Nanosecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            10,
		},
	}
	rsPrf4 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "RESOURCE_FLTR4",
			FilterIDs: []string{},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2015, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:          time.Duration(1) * time.Nanosecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            26,
		},
	}

	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetResourceProfile, rsPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetResourceProfile, rsPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetResourceProfile, rsPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetResourceProfile, rsPrf4, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetResourcesIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaResources,
	}
	expectedIndexes := []string{
		// RESOURCE_FLTR1
		"*prefix:~*req.CostRefunded:12345:RESOURCE_FLTR1",
		"*string:~*req.ToR:*voice:RESOURCE_FLTR1",

		// RESOURCE_FLTR2
		"*prefix:~*req.CostRefunded:12345:RESOURCE_FLTR2",
		"*string:~*req.ToR:*voice:RESOURCE_FLTR2",
		"*string:~*req.Increment:1s:RESOURCE_FLTR2",
		"*string:~*req.Destination:1443:RESOURCE_FLTR2",
		"*prefix:~*req.SetupTime:2022:RESOURCE_FLTR2",

		// RESOURCE_FLTR3
		"*prefix:~*req.CostRefunded:12345:RESOURCE_FLTR3",
		"*string:~*req.ToR:*voice:RESOURCE_FLTR3",
		"*string:~*req.Destination:1443:RESOURCE_FLTR3",
		"*prefix:~*req.SetupTime:2022:RESOURCE_FLTR3",
		"*prefix:~*req.Usage:15s:RESOURCE_FLTR3",

		// RESOURCE_FLTR4
		"*none:*any:*any:RESOURCE_FLTR4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteFilterForResources(t *testing.T) {

}

func testFilterIndexesCasesGetResourcesIndexesChanged(t *testing.T) {

}

func testFilterIndexesCasesGetReverseFilterIndexes3(t *testing.T) {

}

func testFilterIndexesCasesRemoveResourcesProfile(t *testing.T) {

}

func testFilterIndexesCasesGetIndexesAfterRemove3(t *testing.T) {

}

func testFilterIndexesCasesGetReverseIndexesAfterRemove3(t *testing.T) {

}

func testFilterIndexesCasesStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
