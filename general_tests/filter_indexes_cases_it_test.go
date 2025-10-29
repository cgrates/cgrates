//go:build integration
// +build integration

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

		// ATTRIBUTES
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

		// DISPATCHER
		testFilterIndexesCasesSetDispatcherWithFltr,
		testFilterIndexesCasesGetDispatchersIndexesAnyContext,
		testFilterIndexesCasesGetDispatchersIndexesDifferentContext,
		testFilterIndexesCasesOverwriteFilterForDispatchers,
		testFilterIndexesCasesGetDispatchersIndexesChangedAnyContext,
		testFilterIndexesCasesGetDispatchersIndexesChangedDifferentContext,

		testFilterIndexesCasesGetReverseFilterIndexes6,
		testFilterIndexesCasesRemoveDispatchersProfile,
		testFilterIndexesCasesGetIndexesAfterRemoveAnyContext,
		testFilterIndexesCasesGetIndexesAfterRemoveDifferentContext,
		testFilterIndexesCasesGetReverseIndexesAfterRemove6,
		testFilterIndexesCasesOverwriteDispatchersProfile,
		testFilterIndexesCasesOverwriteDispatchersGetIndexesEveryContext,
		testFilterIndexesCasesOverwriteDispatchersGetReverseIndexes,

		// RESOURCES
		testFilterIndexesCasesSetResourceWithFltr,
		testFilterIndexesCasesGetResourcesIndexes,
		testFilterIndexesCasesOverwriteFilterForResources,
		testFilterIndexesCasesGetResourcesIndexesChanged,

		testFilterIndexesCasesGetReverseFilterIndexes3,
		testFilterIndexesCasesRemoveResourcesProfile,
		testFilterIndexesCasesGetIndexesAfterRemove3,
		testFilterIndexesCasesGetReverseIndexesAfterRemove3,
		testFilterIndexesCasesOverwriteResourceProfiles,
		testFilterIndexesCasesResourcesGetIndexesAfterOverwrite,
		testFilterIndexesCasesResourcesGetReverseIndexesAfterOverwrite,

		// SUPPLIER
		testFilterIndexesCasesSetSupplierWithFltr,
		testFilterIndexesCasesGetSuppliersIndexes,
		testFilterIndexesCasesOverwriteFilterForSuppliers,
		testFilterIndexesCasesGetSuppliersIndexesChanged,

		testFilterIndexesCasesGetReverseFilterIndexes4,
		testFilterIndexesCasesRemoveSuppliersProfile,
		testFilterIndexesCasesGetIndexesAfterRemove4,
		testFilterIndexesCasesGetReverseIndexesAfterRemove4,
		testFilterIndexesCasesOverwriteSupplierProfiles,
		testFilterIndexesCasesSuppliersGetIndexesAfterOverwrite,
		testFilterIndexesCasesSuppliersGetReverseIndexesAfterOverwrite,

		// STATS

		//testFilterIndexesCasesOverwriteFilterForSuppliers1,

		testFilterIndexesCasesSetStatQueueWithFltr,
		testFilterIndexesCasesGetStatQueuesIndexes,
		testFilterIndexesCasesOverwriteFilterForStatQueues,
		testFilterIndexesCasesGetStatQueuesIndexesChanged,

		testFilterIndexesCasesGetReverseFilterIndexes5,
		testFilterIndexesCasesRemoveStatQueuesProfile,
		testFilterIndexesCasesGetIndexesAfterRemove5,
		testFilterIndexesCasesGetReverseIndexesAfterRemove5,
		testFilterIndexesCasesOverwriteStatQueueProfiles,
		testFilterIndexesCasesStatQueuesGetIndexesAfterOverwrite,
		testFilterIndexesCasesStatQueuesGetReverseIndexesAfterOverwrite,

		//testMongoFIdx,

		testFilterIndexesCasesStopEngine,
	}
)

// Test start here
func TestFilterIndexesCasesIT(t *testing.T) {
	switch *utils.DBType {
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
	fIdxCasesCfgPath = path.Join(*utils.DataDir, "conf", "samples", fIdxCasesCfgDir)
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
	if _, err := engine.StopStartEngine(fIdxCasesCfgPath, *utils.WaitRater); err != nil {
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

func testFilterIndexesCasesSetDispatcherWithFltr(t *testing.T) {
	dispatcherProfile1 := &v1.DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp1",
			Subsystems: []string{utils.META_ANY},
			FilterIDs: []string{
				"FLTR_Charger",
				"*string:~*req.Account:1001",
				"FLTR_Charger12312",
			},
			Strategy: utils.MetaFirst,
			Weight:   20,
		},
	}
	dispatcherProfile2 := &v1.DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp2",
			Subsystems: []string{utils.META_ANY, utils.MetaSessionS, utils.MetaChargers},
			FilterIDs: []string{
				"*suffix:~*req.Destiantion:01",
				"FLTR_Charger12312",
				"*prefix:~*req.AnswerTime:2022;2021",
			},
			Strategy: utils.MetaFirst,
			Weight:   20,
		},
	}
	dispatcherProfile3 := &v1.DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp3",
			Subsystems: []string{},
			FilterIDs: []string{
				"FLTR_Charger",
				"FLTR_Charger12312",
				"FLTR_Charger4564",
			},
			Strategy: utils.MetaFirst,
			Weight:   20,
		},
	}
	dispatcherProfile4 := &v1.DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp4",
			Subsystems: []string{utils.MetaSessionS},
			FilterIDs: []string{
				"*string:~*req.Account:1001",
				"FLTR_Charger12312",
				"FLTR_Charger4564",
			},
			Strategy: utils.MetaFirst,
			Weight:   20,
		},
	}
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile1,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile2,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile3,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile4,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
}

func testFilterIndexesCasesGetDispatchersIndexesAnyContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.META_ANY,
	}
	// *any context just for Dsp1, Dsp2 and Dsp3
	expectedIndexes := []string{
		// Dsp1
		"*string:~*req.Account:1001:Dsp1",
		"*prefix:~*req.CostRefunded:12345:Dsp1",
		"*string:~*req.ToR:*voice:Dsp1",
		"*string:~*req.Destination:1443:Dsp1",
		"*prefix:~*req.SetupTime:2022:Dsp1",

		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Destination:1443:Dsp2",
		"*prefix:~*req.SetupTime:2022:Dsp2",

		// Dsp3
		"*prefix:~*req.CostRefunded:12345:Dsp3",
		"*string:~*req.ToR:*voice:Dsp3",
		"*string:~*req.Destination:1443:Dsp3",
		"*prefix:~*req.SetupTime:2022:Dsp3",
		"*string:~*req.Increment:1s:Dsp3",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetDispatchersIndexesDifferentContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaSessionS,
	}
	// *sessions context just for Dsp2 and Dsp4
	expectedIndexes := []string{
		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Destination:1443:Dsp2",
		"*prefix:~*req.SetupTime:2022:Dsp2",

		// Dsp4
		"*string:~*req.Account:1001:Dsp4",
		"*string:~*req.Destination:1443:Dsp4",
		"*prefix:~*req.SetupTime:2022:Dsp4",
		"*string:~*req.Increment:1s:Dsp4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaChargers,
	}
	// *chargers context just for Dsp2
	expectedIndexes = []string{
		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Destination:1443:Dsp2",
		"*prefix:~*req.SetupTime:2022:Dsp2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteFilterForDispatchers(t *testing.T) {
	//  FLTR_Charger12312 and FLTR_Charger4564 will be changed,
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger12312",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Dynamically",
					Values:  []string{"true"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.CostUsage",
					Values:  []string{"10", "20", "30"},
				},
			},
		},
	}
	filter3 := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger4564",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Dimension",
					Values:  []string{"20", "25"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.DebitVal",
					Values:  []string{"166"},
				},
				{
					Type:    utils.MetaNotEmpty,
					Element: "~*req.CGRID",
					Values:  []string{},
				},
			},
		},
	}
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetDispatchersIndexesChangedAnyContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.META_ANY,
	}
	// *any context just for Dsp1, Dsp2 and Dsp3
	expectedIndexes := []string{
		// Dsp1
		"*string:~*req.Account:1001:Dsp1",
		"*prefix:~*req.CostRefunded:12345:Dsp1",
		"*string:~*req.ToR:*voice:Dsp1",
		"*string:~*req.Dynamically:true:Dsp1",
		"*prefix:~*req.CostUsage:10:Dsp1",
		"*prefix:~*req.CostUsage:20:Dsp1",
		"*prefix:~*req.CostUsage:30:Dsp1",

		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Dynamically:true:Dsp2",
		"*prefix:~*req.CostUsage:10:Dsp2",
		"*prefix:~*req.CostUsage:20:Dsp2",
		"*prefix:~*req.CostUsage:30:Dsp2",

		// Dsp3
		"*prefix:~*req.CostRefunded:12345:Dsp3",
		"*string:~*req.ToR:*voice:Dsp3",
		"*string:~*req.Dynamically:true:Dsp3",
		"*prefix:~*req.CostUsage:10:Dsp3",
		"*prefix:~*req.CostUsage:20:Dsp3",
		"*prefix:~*req.CostUsage:30:Dsp3",
		"*string:~*req.Dimension:20:Dsp3",
		"*string:~*req.Dimension:25:Dsp3",
		"*prefix:~*req.DebitVal:166:Dsp3",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetDispatchersIndexesChangedDifferentContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaSessionS,
	}
	// *sessions context just for Dsp2 and Dsp4
	expectedIndexes := []string{
		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Dynamically:true:Dsp2",
		"*prefix:~*req.CostUsage:10:Dsp2",
		"*prefix:~*req.CostUsage:20:Dsp2",
		"*prefix:~*req.CostUsage:30:Dsp2",

		// Dsp4
		"*string:~*req.Account:1001:Dsp4",
		"*string:~*req.Dynamically:true:Dsp4",
		"*prefix:~*req.CostUsage:10:Dsp4",
		"*prefix:~*req.CostUsage:20:Dsp4",
		"*prefix:~*req.CostUsage:30:Dsp4",
		"*string:~*req.Dimension:20:Dsp4",
		"*string:~*req.Dimension:25:Dsp4",
		"*prefix:~*req.DebitVal:166:Dsp4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaChargers,
	}
	// *sessions context just for Dsp2
	expectedIndexes = []string{
		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Dynamically:true:Dsp2",
		"*prefix:~*req.CostUsage:10:Dsp2",
		"*prefix:~*req.CostUsage:20:Dsp2",
		"*prefix:~*req.CostUsage:30:Dsp2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseFilterIndexes6(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp1",
		"*dispatcher_filter_indexes:Dsp3",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*dispatcher_filter_indexes:Dsp3",
		"*dispatcher_filter_indexes:Dsp4",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*dispatcher_filter_indexes:Dsp1",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp3",
		"*dispatcher_filter_indexes:Dsp4",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesRemoveDispatchersProfile(t *testing.T) {
	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Dsp1"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Dsp3"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}
}

func testFilterIndexesCasesGetIndexesAfterRemoveAnyContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.META_ANY,
	}
	// *any context just for Dsp2
	expectedIndexes := []string{
		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Dynamically:true:Dsp2",
		"*prefix:~*req.CostUsage:10:Dsp2",
		"*prefix:~*req.CostUsage:20:Dsp2",
		"*prefix:~*req.CostUsage:30:Dsp2",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetIndexesAfterRemoveDifferentContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaSessionS,
	}
	// *sessions context just for Dsp2 and Dsp4
	expectedIndexes := []string{
		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Dynamically:true:Dsp2",
		"*prefix:~*req.CostUsage:10:Dsp2",
		"*prefix:~*req.CostUsage:20:Dsp2",
		"*prefix:~*req.CostUsage:30:Dsp2",

		// Dsp4
		"*string:~*req.Account:1001:Dsp4",
		"*string:~*req.Dynamically:true:Dsp4",
		"*prefix:~*req.CostUsage:10:Dsp4",
		"*prefix:~*req.CostUsage:20:Dsp4",
		"*prefix:~*req.CostUsage:30:Dsp4",
		"*string:~*req.Dimension:20:Dsp4",
		"*string:~*req.Dimension:25:Dsp4",
		"*prefix:~*req.DebitVal:166:Dsp4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaChargers,
	}
	// *sessions context just for Dsp2
	expectedIndexes = []string{
		// Dsp2
		"*prefix:~*req.AnswerTime:2022:Dsp2",
		"*prefix:~*req.AnswerTime:2021:Dsp2",
		"*string:~*req.Dynamically:true:Dsp2",
		"*prefix:~*req.CostUsage:10:Dsp2",
		"*prefix:~*req.CostUsage:20:Dsp2",
		"*prefix:~*req.CostUsage:30:Dsp2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseIndexesAfterRemove6(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*dispatcher_filter_indexes:Dsp4",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteDispatchersProfile(t *testing.T) {
	dispatcherProfile1 := &v1.DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp2",
			Subsystems: []string{utils.MetaSessionS, utils.MetaAttributes},
			FilterIDs: []string{
				"*string:~*req.AccountToMatch:1005",
				"FLTR_Charger",
				"*prefix:~*req.Destination:990",
			},
			Strategy: utils.MetaFirst,
			Weight:   20,
		},
	}
	dispatcherProfile4 := &v1.DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "Dsp4",
			Subsystems: []string{utils.MetaChargers, utils.MetaSessionS},
			FilterIDs: []string{
				"FLTR_Charger",
				"*gt:~*req.ProcessRuns:2",
			},
			Strategy: utils.MetaFirst,
			Weight:   20,
		},
	}
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile1,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile4,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}
}

func testFilterIndexesCasesOverwriteDispatchersGetIndexesEveryContext(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaSessionS,
	}
	// *sessions context just for Dsp2 and Dsp4
	expectedIndexes := []string{
		// Dsp2
		"*prefix:~*req.Destination:990:Dsp2",
		"*string:~*req.AccountToMatch:1005:Dsp2",
		"*prefix:~*req.CostRefunded:12345:Dsp2",
		"*string:~*req.ToR:*voice:Dsp2",

		// Dsp4
		"*prefix:~*req.CostRefunded:12345:Dsp4",
		"*string:~*req.ToR:*voice:Dsp4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaAttributes,
	}
	// *sessions context just for Dsp2
	expectedIndexes = []string{
		// Dsp2
		"*prefix:~*req.Destination:990:Dsp2",
		"*string:~*req.AccountToMatch:1005:Dsp2",
		"*prefix:~*req.CostRefunded:12345:Dsp2",
		"*string:~*req.ToR:*voice:Dsp2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaDispatchers,
		Context:  utils.MetaChargers,
	}
	// *sessions context just for Dsp4
	expectedIndexes = []string{
		// Dsp4
		"*prefix:~*req.CostRefunded:12345:Dsp4",
		"*string:~*req.ToR:*voice:Dsp4",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteDispatchersGetReverseIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*prefix:~*req.CostUsage:10:RESOURCE_FLTR2",
		"*prefix:~*req.CostUsage:20:RESOURCE_FLTR2",
		"*prefix:~*req.CostUsage:30:RESOURCE_FLTR2",
		"*prefix:~*req.DebitVal:166:RESOURCE_FLTR2",
		"*string:~*req.Dimension:20:RESOURCE_FLTR2",
		"*string:~*req.Dimension:25:RESOURCE_FLTR2",
		"*string:~*req.Dynamically:true:RESOURCE_FLTR2",
		"*string:~*req.ToR:*voice:RESOURCE_FLTR2",

		// RESOURCE_FLTR3
		"*prefix:~*req.CostRefunded:12345:RESOURCE_FLTR3",
		"*prefix:~*req.CostUsage:10:RESOURCE_FLTR3",
		"*prefix:~*req.CostUsage:20:RESOURCE_FLTR3",
		"*prefix:~*req.CostUsage:30:RESOURCE_FLTR3",
		"*prefix:~*req.Usage:15s:RESOURCE_FLTR3",
		"*string:~*req.Dynamically:true:RESOURCE_FLTR3",
		"*string:~*req.ToR:*voice:RESOURCE_FLTR3",

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
	// FLTR_Charger, FLTR_Charger12312 and FLTR_Charger4564 will be changed
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Destination",
					Values:  []string{"1023"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Agent",
					Values:  []string{"Freeswitchv"},
				},
			},
		},
	}
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger12312",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.TimeToAnswer",
					Values:  []string{"2022"},
				},
			},
		},
	}
	filter3 := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger4564",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subsystem",
					Values:  []string{"Resources"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.CGRID",
					Values:  []string{"test_"},
				},
				{
					Type:    utils.MetaNotEmpty,
					Element: "~*req.Hash",
					Values:  []string{},
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
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetResourcesIndexesChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaResources,
	}
	expectedIndexes := []string{
		// RESOURCE_FLTR1
		"*string:~*req.Destination:1023:RESOURCE_FLTR1",
		"*prefix:~*req.Agent:Freeswitchv:RESOURCE_FLTR1",

		// RESOURCE_FLTR2
		"*string:~*req.Destination:1023:RESOURCE_FLTR2",
		"*prefix:~*req.Agent:Freeswitchv:RESOURCE_FLTR2",
		"*prefix:~*req.TimeToAnswer:2022:RESOURCE_FLTR2",
		"*string:~*req.Subsystem:Resources:RESOURCE_FLTR2",
		"*prefix:~*req.CGRID:test_:RESOURCE_FLTR2",

		// RESOURCE_FLTR3
		"*string:~*req.Destination:1023:RESOURCE_FLTR3",
		"*prefix:~*req.Agent:Freeswitchv:RESOURCE_FLTR3",
		"*prefix:~*req.TimeToAnswer:2022:RESOURCE_FLTR3",
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

func testFilterIndexesCasesGetReverseFilterIndexes3(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR2",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR2",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesRemoveResourcesProfile(t *testing.T) {
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RESOURCE_FLTR2"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "RESOURCE_FLTR4"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testFilterIndexesCasesGetIndexesAfterRemove3(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaResources,
	}
	expectedIndexes := []string{
		// RESOURCE_FLTR1
		"*string:~*req.Destination:1023:RESOURCE_FLTR1",
		"*prefix:~*req.Agent:Freeswitchv:RESOURCE_FLTR1",

		// RESOURCE_FLTR3
		"*string:~*req.Destination:1023:RESOURCE_FLTR3",
		"*prefix:~*req.Agent:Freeswitchv:RESOURCE_FLTR3",
		"*prefix:~*req.TimeToAnswer:2022:RESOURCE_FLTR3",
		"*prefix:~*req.Usage:15s:RESOURCE_FLTR3",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseIndexesAfterRemove3(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteResourceProfiles(t *testing.T) {
	rsPrf1 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "RESOURCE_FLTR1",
			FilterIDs: []string{
				"FLTR_Charger12312",
				"*string:~*req.UsageCost:15000",
				"FLTR_Charger4564",
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
	rsPrf3 := &v1.ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "RESOURCE_FLTR3",
			FilterIDs: []string{
				"*prefix:~*req.Time:9s",
				"FLTR_Charger4564",
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

	var result string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetResourceProfile, rsPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetResourceProfile, rsPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesResourcesGetIndexesAfterOverwrite(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaResources,
	}
	expectedIndexes := []string{
		// RESOURCE_FLTR1
		"*prefix:~*req.TimeToAnswer:2022:RESOURCE_FLTR1",
		"*string:~*req.UsageCost:15000:RESOURCE_FLTR1",
		"*string:~*req.Subsystem:Resources:RESOURCE_FLTR1",
		"*prefix:~*req.CGRID:test_:RESOURCE_FLTR1",

		// RESOURCE_FLTR3
		"*string:~*req.Subsystem:Resources:RESOURCE_FLTR3",
		"*prefix:~*req.CGRID:test_:RESOURCE_FLTR3",
		"*prefix:~*req.Time:9s:RESOURCE_FLTR3",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesResourcesGetReverseIndexesAfterOverwrite(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesSetSupplierWithFltr(t *testing.T) {
	// set another routes profile different than the one from tariffplan
	rPrf1 := &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant: "cgrates.org",
			ID:     "Supplier1",
			FilterIDs: []string{
				"FLTR_Charger",
			},
			Sorting:           "*weight",
			SortingParameters: []string{"Param1"},
			Suppliers: []*engine.Supplier{{
				ID:      "SPL1",
				Weight:  20,
				Blocker: false,
			}},
			Weight: 10,
		},
	}
	rPrf2 := &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant: "cgrates.org",
			ID:     "Supplier2",
			FilterIDs: []string{
				"FLTR_Charger4564",
				"FLTR_Charger",
				"FLTR_Charger12312",
			},
			Sorting:           "*weight",
			SortingParameters: []string{"Param1"},
			Suppliers: []*engine.Supplier{{
				ID:      "SPL1",
				Weight:  20,
				Blocker: false,
			}},
			Weight: 10,
		},
	}
	rPrf3 := &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant: "cgrates.org",
			ID:     "Supplier3",
			FilterIDs: []string{
				"FLTR_Charger12312",
				"*prefix:~*req.Cost:4",
				"FLTR_Charger",
			},
			Sorting:           "*weight",
			SortingParameters: []string{"Param1"},
			Suppliers: []*engine.Supplier{{
				ID:      "SPL1",
				Weight:  20,
				Blocker: false,
			}},
			Weight: 10,
		},
	}
	rPrf4 := &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:            "cgrates.org",
			ID:                "Supplier4",
			FilterIDs:         []string{},
			Sorting:           "*weight",
			SortingParameters: []string{"Param1"},
			Suppliers: []*engine.Supplier{{
				ID:      "SPL1",
				Weight:  20,
				Blocker: false,
			}},
			Weight: 10,
		},
	}
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetSupplierProfile, rPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetSupplierProfile, rPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetSupplierProfile, rPrf3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetSupplierProfile, rPrf4, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testFilterIndexesCasesGetSuppliersIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaSuppliers,
	}
	expectedIndexes := []string{
		// Supplier1
		"*string:~*req.Destination:1023:Supplier1",
		"*prefix:~*req.Agent:Freeswitchv:Supplier1",

		// Supplier2
		"*string:~*req.Destination:1023:Supplier2",
		"*prefix:~*req.Agent:Freeswitchv:Supplier2",
		"*prefix:~*req.TimeToAnswer:2022:Supplier2",
		"*string:~*req.Subsystem:Resources:Supplier2",
		"*prefix:~*req.CGRID:test_:Supplier2",

		// Supplier3
		"*string:~*req.Destination:1023:Supplier3",
		"*prefix:~*req.Agent:Freeswitchv:Supplier3",
		"*prefix:~*req.TimeToAnswer:2022:Supplier3",
		"*prefix:~*req.Cost:4:Supplier3",

		// Supplier4
		"*none:*any:*any:Supplier4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteFilterForSuppliers1(t *testing.T) {
	// FLTR_Charger, FLTR_Charger12312 and FLTR_Charger4564 will be changed
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Bank",
					Values:  []string{"BoA", "CEC"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Customer",
					Values:  []string{"11", "22"},
				},
			},
		},
	}
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger12312",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.AnswerTime",
					Values:  []string{"2010", "2011"},
				},
				{
					Type:    utils.MetaGreaterThan,
					Element: "~*req.ProcessRuns",
					Values:  []string{"2"},
				},
			},
		},
	}
	filter3 := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger4564",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Caching",
					Values:  []string{"true"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.NoCall",
					Values:  []string{"+4332225465"},
				},
				{
					Type:    utils.MetaNotEmpty,
					Element: "~*req.Hash",
					Values:  []string{},
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
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetSuppliersIndexesChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaSuppliers,
	}
	expectedIndexes := []string{
		// Supplier1
		"*string:~*req.Bank:BoA:Supplier1",
		"*string:~*req.Bank:CEC:Supplier1",
		"*prefix:~*req.Customer:11:Supplier1",
		"*prefix:~*req.Customer:22:Supplier1",

		// Supplier2
		"*string:~*req.Bank:BoA:Supplier2",
		"*string:~*req.Bank:CEC:Supplier2",
		"*prefix:~*req.Customer:11:Supplier2",
		"*prefix:~*req.Customer:22:Supplier2",
		"*prefix:~*req.NoCall:+4332225465:Supplier2",
		"*string:~*req.Caching:true:Supplier2",
		"*prefix:~*req.AnswerTime:2010:Supplier2",
		"*prefix:~*req.AnswerTime:2011:Supplier2",

		// Supplier3
		"*string:~*req.Bank:BoA:Supplier3",
		"*string:~*req.Bank:CEC:Supplier3",
		"*prefix:~*req.Customer:11:Supplier3",
		"*prefix:~*req.Customer:22:Supplier3",
		"*prefix:~*req.Cost:4:Supplier3",
		"*prefix:~*req.AnswerTime:2010:Supplier3",
		"*prefix:~*req.AnswerTime:2011:Supplier3",

		// Supplier4
		"*none:*any:*any:Supplier4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseFilterIndexes4(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*supplier_filter_indexes:Supplier1",
		"*supplier_filter_indexes:Supplier2",
		"*supplier_filter_indexes:Supplier3",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*supplier_filter_indexes:Supplier2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*supplier_filter_indexes:Supplier2",
		"*supplier_filter_indexes:Supplier3",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesRemoveSuppliersProfile(t *testing.T) {
	var resp string
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Supplier1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Supplier3"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testFilterIndexesCasesGetIndexesAfterRemove4(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaSuppliers,
	}
	expectedIndexes := []string{
		// Supplier2
		"*string:~*req.Bank:BoA:Supplier2",
		"*string:~*req.Bank:CEC:Supplier2",
		"*prefix:~*req.Customer:11:Supplier2",
		"*prefix:~*req.Customer:22:Supplier2",
		"*prefix:~*req.NoCall:+4332225465:Supplier2",
		"*string:~*req.Caching:true:Supplier2",
		"*prefix:~*req.AnswerTime:2010:Supplier2",
		"*prefix:~*req.AnswerTime:2011:Supplier2",

		// Supplier4
		"*none:*any:*any:Supplier4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseIndexesAfterRemove4(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*supplier_filter_indexes:Supplier2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*supplier_filter_indexes:Supplier2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*supplier_filter_indexes:Supplier2",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteSupplierProfiles(t *testing.T) {
	rPrf2 := &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant: "cgrates.org",
			ID:     "Supplier2",
			FilterIDs: []string{
				"FLTR_Charger12312",
				"*lt:~*req.Distance:2000",
			},
			Sorting:           "*weight",
			SortingParameters: []string{"Param1"},
			Suppliers: []*engine.Supplier{{
				ID:      "SPL1",
				Weight:  20,
				Blocker: false,
			}},
			Weight: 10,
		},
	}
	rPrf4 := &v1.SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant: "cgrates.org",
			ID:     "Supplier4",
			FilterIDs: []string{
				"FLTR_Charger",
				"FLTR_Charger12312",
				"*string:~*req.Account:itsyscom",
				"*gt:~*req.Usage:10s",
			},
			Sorting:           "*weight",
			SortingParameters: []string{"Param1"},
			Suppliers: []*engine.Supplier{{
				ID:      "SPL1",
				Weight:  20,
				Blocker: false,
			}},
			Weight: 10,
		},
	}
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetSupplierProfile, rPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetSupplierProfile, rPrf4, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testFilterIndexesCasesSuppliersGetIndexesAfterOverwrite(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaSuppliers,
	}
	expectedIndexes := []string{
		// Supplier2
		"*prefix:~*req.AnswerTime:2010:Supplier2",
		"*prefix:~*req.AnswerTime:2011:Supplier2",

		// Supplier4
		"*string:~*req.Bank:BoA:Supplier4",
		"*string:~*req.Bank:CEC:Supplier4",
		"*prefix:~*req.Customer:11:Supplier4",
		"*prefix:~*req.Customer:22:Supplier4",
		"*prefix:~*req.AnswerTime:2010:Supplier4",
		"*prefix:~*req.AnswerTime:2011:Supplier4",
		"*string:~*req.Account:itsyscom:Supplier4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesSuppliersGetReverseIndexesAfterOverwrite(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*supplier_filter_indexes:Supplier4",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*supplier_filter_indexes:Supplier2",
		"*supplier_filter_indexes:Supplier4",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteFilterForSuppliers(t *testing.T) {
	// FLTR_Charger, FLTR_Charger12312 and FLTR_Charger4564 will be changed
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Bank",
					Values:  []string{"BoA", "CEC"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Customer",
					Values:  []string{"11", "22"},
				},
			},
		},
	}
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger12312",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.AnswerTime",
					Values:  []string{"2010", "2011"},
				},
				{
					Type:    utils.MetaGreaterThan,
					Element: "~*req.ProcessRuns",
					Values:  []string{"2"},
				},
			},
		},
	}
	filter3 := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger4564",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Caching",
					Values:  []string{"true"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.NoCall",
					Values:  []string{"+4332225465"},
				},
				{
					Type:    utils.MetaNotEmpty,
					Element: "~*req.Hash",
					Values:  []string{},
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
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesSetStatQueueWithFltr(t *testing.T) {
	stat1 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats1",
			FilterIDs: []string{
				"FLTR_Charger",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	stat2 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats2",
			FilterIDs: []string{
				"FLTR_Charger4564",
				"FLTR_Charger",
				"FLTR_Charger12312",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	stat3 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats3",
			FilterIDs: []string{
				"FLTR_Charger12312",
				"*prefix:~*req.Cost:4",
				"FLTR_Charger",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	stat4 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "Stats4",
			FilterIDs:   []string{},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat4, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testFilterIndexesCasesGetStatQueuesIndexes(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes := []string{
		// Stats1
		"*string:~*req.Bank:BoA:Stats1",
		"*string:~*req.Bank:CEC:Stats1",
		"*prefix:~*req.Customer:11:Stats1",
		"*prefix:~*req.Customer:22:Stats1",

		// Stats2
		"*string:~*req.Bank:BoA:Stats2",
		"*string:~*req.Bank:CEC:Stats2",
		"*prefix:~*req.Customer:11:Stats2",
		"*prefix:~*req.Customer:22:Stats2",
		"*prefix:~*req.AnswerTime:2010:Stats2",
		"*prefix:~*req.AnswerTime:2011:Stats2",
		"*string:~*req.Caching:true:Stats2",
		"*prefix:~*req.NoCall:+4332225465:Stats2",

		// Stats3
		"*prefix:~*req.AnswerTime:2010:Stats3",
		"*prefix:~*req.AnswerTime:2011:Stats3",
		"*string:~*req.Bank:BoA:Stats3",
		"*string:~*req.Bank:CEC:Stats3",
		"*prefix:~*req.Customer:11:Stats3",
		"*prefix:~*req.Customer:22:Stats3",
		"*prefix:~*req.Cost:4:Stats3",

		// Stats4
		"*none:*any:*any:Stats4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteFilterForStatQueues(t *testing.T) {
	// FLTR_Charger, FLTR_Charger12312 and FLTR_Charger4564 will be changed
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.RatingPlan",
					Values:  []string{"RP1"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Subject",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger12312",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Category",
					Values:  []string{"call"},
				},
				{
					Type:    utils.MetaGreaterThan,
					Element: "~*req.ProcessRuns",
					Values:  []string{"2"},
				},
			},
		},
	}
	filter3 := &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger4564",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaNotEmpty,
					Element: "~*req.Destintion",
					Values:  []string{},
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
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterIndexesCasesGetStatQueuesIndexesChanged(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes := []string{
		// Stats1
		"*string:~*req.RatingPlan:RP1:Stats1",
		"*prefix:~*req.Subject:1001:Stats1",
		"*prefix:~*req.Subject:1002:Stats1",

		// Stats2
		"*string:~*req.RatingPlan:RP1:Stats2",
		"*prefix:~*req.Subject:1001:Stats2",
		"*prefix:~*req.Subject:1002:Stats2",
		"*string:~*req.Category:call:Stats2",

		// Stats3
		"*string:~*req.Category:call:Stats3",
		"*prefix:~*req.Cost:4:Stats3",
		"*string:~*req.RatingPlan:RP1:Stats3",
		"*prefix:~*req.Subject:1001:Stats3",
		"*prefix:~*req.Subject:1002:Stats3",

		// Stats4
		"*none:*any:*any:Stats4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseFilterIndexes5(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*supplier_filter_indexes:Supplier4",
		"*stat_filter_indexes:Stats1",
		"*stat_filter_indexes:Stats2",
		"*stat_filter_indexes:Stats3",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*stat_filter_indexes:Stats2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*supplier_filter_indexes:Supplier2",
		"*supplier_filter_indexes:Supplier4",
		"*stat_filter_indexes:Stats2",
		"*stat_filter_indexes:Stats3",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesRemoveStatQueuesProfile(t *testing.T) {
	var rply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "Stats1",
		}, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "Stats3",
		}, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
}

func testFilterIndexesCasesGetIndexesAfterRemove5(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes := []string{
		// Stats2
		"*string:~*req.RatingPlan:RP1:Stats2",
		"*prefix:~*req.Subject:1001:Stats2",
		"*prefix:~*req.Subject:1002:Stats2",
		"*string:~*req.Category:call:Stats2",

		// Stats4
		"*none:*any:*any:Stats4",
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesGetReverseIndexesAfterRemove5(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*supplier_filter_indexes:Supplier4",
		"*stat_filter_indexes:Stats2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*stat_filter_indexes:Stats2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*supplier_filter_indexes:Supplier2",
		"*supplier_filter_indexes:Supplier4",
		"*stat_filter_indexes:Stats2",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesOverwriteStatQueueProfiles(t *testing.T) {
	stat2 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats2",
			FilterIDs: []string{
				"FLTR_Charger",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	stat4 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats4",
			FilterIDs: []string{
				"FLTR_Charger4564",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat4, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testFilterIndexesCasesStatQueuesGetIndexesAfterOverwrite(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes := []string{
		// Stats2
		"*string:~*req.RatingPlan:RP1:Stats2",
		"*prefix:~*req.Subject:1001:Stats2",
		"*prefix:~*req.Subject:1002:Stats2",

		// Stats4 nothing
	}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testFilterIndexesCasesStatQueuesGetReverseIndexesAfterOverwrite(t *testing.T) {
	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org:FLTR_Charger",
		ItemType: utils.CacheReverseFilterIndexes,
	}
	expectedIndexes := []string{
		"*charger_filter_indexes:ChrgerIndexable",
		"*dispatcher_filter_indexes:Dsp2",
		"*dispatcher_filter_indexes:Dsp4",
		"*supplier_filter_indexes:Supplier4",
		"*stat_filter_indexes:Stats2",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*resource_filter_indexes:RESOURCE_FLTR3",
		"*stat_filter_indexes:Stats4",
		"*threshold_filter_indexes:TEST_PROFILE2",
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
		"*resource_filter_indexes:RESOURCE_FLTR1",
		"*supplier_filter_indexes:Supplier2",
		"*supplier_filter_indexes:Supplier4",
		"*threshold_filter_indexes:TEST_PROFILE2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testMongoFIdx(t *testing.T) {
	// FLTR_Charger, FLTR_Charger12312 and FLTR_Charger4564 will be changed
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Bank",
					Values:  []string{"BoA", "CEC"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Customer",
					Values:  []string{"11", "22"},
				},
			},
		},
	}
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
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

	stat1 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats1",
			FilterIDs: []string{
				"FLTR_Charger",
				"FLTR_Charger2",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	stat2 := &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats2",
			FilterIDs: []string{
				"FLTR_Charger",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	var reply string
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetStatQueueProfile, stat2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

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
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetThresholdProfile, tPrfl1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	arg := &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes := []string{
		// Stats1
		"*string:~*req.Bank:BoA:Stats1",
		"*string:~*req.Bank:CEC:Stats1",
		"*prefix:~*req.Customer:11:Stats1",
		"*prefix:~*req.Customer:22:Stats1",
		"*string:~*req.Account:1001:Stats1",

		// Stats2
		"*string:~*req.Bank:BoA:Stats2",
		"*string:~*req.Bank:CEC:Stats2",
		"*prefix:~*req.Customer:11:Stats2",
		"*prefix:~*req.Customer:22:Stats2",
	}
	sort.Strings(expectedIndexes)
	var replyIDx []string
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &replyIDx); err != nil {
		t.Error(err)
	} else if sort.Strings(replyIDx); !reflect.DeepEqual(expectedIndexes, replyIDx) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(replyIDx))
	}

	// FLTR_Charger, FLTR_Charger12312 and FLTR_Charger4564 will be changed
	filter1 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.RatingPlan",
					Values:  []string{"RP1"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Subject",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}
	filter2 = &v1.FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Destination",
					Values:  []string{"randomID"},
				},
			},
		},
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := fIdxCasesRPC.Call(utils.APIerSv1SetFilter, filter2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	arg = &v1.AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes = []string{
		// Stats1
		"*string:~*req.RatingPlan:RP1:Stats1",
		"*prefix:~*req.Subject:1001:Stats1",
		"*prefix:~*req.Subject:1002:Stats1",
		"*string:~*req.Destination:randomID:Stats1",

		// Stats2
		"*string:~*req.RatingPlan:RP1:Stats2",
		"*prefix:~*req.Subject:1001:Stats2",
		"*prefix:~*req.Subject:1002:Stats2",
	}
	sort.Strings(expectedIndexes)
	if err := fIdxCasesRPC.Call(utils.APIerSv1GetFilterIndexes, arg, &replyIDx); err != nil {
		t.Error(err)
	} else if sort.Strings(replyIDx); !reflect.DeepEqual(expectedIndexes, replyIDx) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(replyIDx))
	}
}

func testFilterIndexesCasesStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
