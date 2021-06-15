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

package apis

import (
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tFltrIdxCfgPath string
	tFIdxRpc        *birpc.Client
	tFltrIdxCfg     *config.CGRConfig
	tFltrIdxConfDIR string

	sTestsFilterIndexesSV1 = []func(t *testing.T){
		testV1FIdxLoadConfig,
		testV1FIdxdxInitDataDb,
		testV1FIdxResetStorDb,
		testV1FIdxStartEngine,
		testV1FIdxRpcConn,

		testV1FIdxSetAttributeSProfileWithFltr,
		testV1FIdxSetAttributeSMoreFltrsMoreIndexing,
		testV1FIdxAttributesRemoveIndexes,
		testV1FIdxAttributeComputeIndexes,
		testV1FIdxAttributeMoreProfilesForFilters,
		testV1FIdxAttributeSRemoveComputedIndexesIDs,
		testV1FIdxAttributesRemoveProfilesNoIndexes,
		testV1IndexClearCache,

		testV1FIdxSetAccountWithFltr,
		testVF1FIdxSetAccountMoreFltrsMoreIndexing,
		testVIFIdxAccountRemoveIndexes,
		testV1FIdxAccountComputeIndexes,
		testV1FIdxAccountsMoreProfilesForFilters,
		testV1FIdxAccountSRemoveComputedIndexesIDs,
		testV1FIdxAccountRemoveAccountNoIndexes,
		testV1IndexClearCache,

		testV1FIdxSetActionProfileWithFltr,
		testV1FIdxSetActionProfileMoreFltrsMoreIndexing,
		testV1FIdxActionProfileRemoveIndexes,
		testV1FIdxActionProfileComputeIndexes,
		testV1FIdxActionMoreProfileForFilters,
		testV1FIdxActionSRemoveComputedIndexesIDs,
		testV1FIdxActionSRemoveActionsNoIndexes,
		testV1IndexClearCache,

		testV1FIdxSetChargerSProfileWithFltr,
		testV1FIdxSetChargerProfileMoreFltrsMoreIndexing,
		testV1FIdxChargerProfileRemoveIndexes,
		testV1FIdxChargerProfileComputeIndexes,
		testV1FIdxChargerMoreProfileForFilters,
		testV1FIdxChargerSRemoveComputedIndexesIDs,
		testV1FIdxChargerSRemoveActionsNoIndexes,
		testV1IndexClearCache,

		testV1FIdxSetRateSProfileWithFltr,
		testV1FIdxSetRateProfileMoreFltrsMoreIndexing,
		testV1FIdxRateSProfileRemoveIndexes,
		testV1FIdxRateSProfileComputeIndexes,
		testV1FIdxRateSProfileMoreProfilesForFilters,
		testV1FIdxRateSRemoveComputedIndexesIDs,
		testV1FIdxRateSProfileRemoveRateProfileNoIndexes,
		testV1IndexClearCache,

		testV1FIdxRateSProfileRatesWithFltr,
		testV1FIdxSetRatePofileRatesMoreFltrsMoreIndexing,
		testV1FIdxRateProfileRatesRemoveIndexes,
		testV1FIdxRateProfileRatesComputeIndexes,
		testV1FIdxRateProfileRatesMoreRatesForFilters,
		testV1FIdxRateProfileRatesRemoveComputedIndexesIDs,
		testV1FIdxRateProfileRatesRemoveRateProfileRatesNoIndexes,
		testV1IndexClearCache,

		testV1FIdxSetResourceSProfileWithFltr,
		testV1FIdxSetResourceSMoreFltrsMoreIndexing,
		testV1FIdxResourceSProfileRemoveIndexes,
		testV1FIdxResourceSProfileComputeIndexes,

		testV1FIdxStopEngine,
	}
	fltr = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
			},
		},
	}
	fltrSameID = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.CGRID",
					Values:  []string{"QWEASDZXC", "IOPJKLBNM"},
				},
			},
		},
	}
	fltr1 = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Usage",
					Values:  []string{"123s"},
				},
			},
		},
	}
	fltr2 = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_attr3",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.AnswerTime",
					Values:  []string{"12", "33"},
				},
			},
		},
	}
)

// Test start here
func TestFltrIdxV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tFltrIdxConfDIR = "filter_indexes_internal"
	case utils.MetaMySQL:
		tFltrIdxConfDIR = "filter_indexes_mysql"
	case utils.MetaMongo:
		tFltrIdxConfDIR = "filter_indexes_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsFilterIndexesSV1 {
		t.Run(tFltrIdxConfDIR, stest)
	}
}

func testV1FIdxLoadConfig(t *testing.T) {
	tFltrIdxCfgPath = path.Join(*dataDir, "conf", "samples", tFltrIdxConfDIR)
	var err error
	if tFltrIdxCfg, err = config.NewCGRConfigFromPath(tFltrIdxCfgPath); err != nil {
		t.Error(err)
	}
}

func testV1FIdxdxInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(tFltrIdxCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1IndexClearCache(t *testing.T) {
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(tFltrIdxCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tFltrIdxCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxRpcConn(t *testing.T) {
	var err error
	tFIdxRpc, err = newRPCClient(tFltrIdxCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FIdxSetAttributeSProfileWithFltr(t *testing.T) {
	// First we will set a filter for usage
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// Get filter for checking it's existence
	var resultFltr *engine.Filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: utils.CGRateSorg, ID: "fltr_for_attr"}, &resultFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultFltr, fltr.Filter) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(fltr.Filter), utils.ToJSON(resultFltr))
	}

	//we will set an AttributeProfile with our filter and check the indexes
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			FilterIDs: []string{"fltr_for_attr", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	//update the filter for checking the indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrSameID, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*opts.*context:*sessions:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.CGRID:QWEASDZXC:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.CGRID:IOPJKLBNM:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	//back to our initial filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testV1FIdxSetAttributeSMoreFltrsMoreIndexing(t *testing.T) {
	// More filters for our AttributeProfile
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// update our Attribute with our filters
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant: utils.CGRateSorg,
			ID:     "TEST_ATTRIBUTES_IT_TEST",
			FilterIDs: []string{"fltr_for_attr", "fltr_for_attr2",
				"fltr_for_attr3", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxAttributesRemoveIndexes(t *testing.T) {
	var reply string
	var replyIdx []string
	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}

func testV1FIdxAttributeComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, AttributeS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	var replyIdx []string

	//matching for our context
	expectedIDx := []string{"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxAttributeMoreProfilesForFilters(t *testing.T) {
	//we will add more attributes with different context for matching filters
	attrPrf2 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_new_fltr",
			FilterIDs: []string{"fltr_for_attr2", "fltr_for_attr3", "*string:~*opts.*context:*chargers"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
			},
		},
	}
	attrPrf3 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTE3",
			FilterIDs: []string{"fltr_for_attr3", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*req.Destinations",
					Type:  utils.MetaConstant,
					Value: "1008",
				},
			},
		},
	}
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var replyIdx []string
	expectedIDx := []string{"*string:*req.Usage:123s:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*chargers:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTE3",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxAttributeSRemoveComputedIndexesIDs(t *testing.T) {
	//indexes will be removed for both contexts
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes, APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaClear,
		}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//not found for both cases
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// now we will ComputeFilterIndexes by IDs for *sessions context(but just only 1 profile, not both)
	var expIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			AttributeIDs: []string{"TEST_ATTRIBUTES_new_fltr", "TEST_ATTRIBUTE3"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	//able to get indexes with context *sessions
	expIdx = []string{"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*chargers:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTE3",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_new_fltr"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}

	// compute for the last profile remain
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			AttributeIDs: []string{"TEST_ATTRIBUTES_IT_TEST"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}
	expIdx = []string{"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.*context:*chargers:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTE3",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_new_fltr",
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxAttributesRemoveProfilesNoIndexes(t *testing.T) {
	//as we delete our profiles, indexes will  be deleted too
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "TEST_ATTRIBUTES_IT_TEST",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "TEST_ATTRIBUTE3",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "TEST_ATTRIBUTES_new_fltr",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: utils.CGRateSorg,
			ID: "fltr_for_attr"}}, &reply); err != nil {
		t.Error(err)
	}

	// Check indexes as we removed, not found for both indexes
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAttributes}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSetAccountWithFltr(t *testing.T) {
	// First we will set a filter for usage
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// Get filter for checking it's existence
	var resultFltr *engine.Filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: utils.CGRateSorg, ID: "fltr_for_attr"}, &resultFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultFltr, fltr.Filter) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(fltr.Filter), utils.ToJSON(resultFltr))
	}

	//we will set an Account with our filter and check the indexes
	accPrf := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "ACCOUNT_FILTER_INDEXES",
			Weights:   ";0",
			FilterIDs: []string{"fltr_for_attr", "*string:~*opts.*context:*sessions"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";15",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(1),
						},
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//here will check the indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	//update the filter for checking the indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrSameID, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES",
		"*string:*req.CGRID:QWEASDZXC:ACCOUNT_FILTER_INDEXES",
		"*string:*req.CGRID:IOPJKLBNM:ACCOUNT_FILTER_INDEXES"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
	//back to our initial filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Errorf("%q", err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testVF1FIdxSetAccountMoreFltrsMoreIndexing(t *testing.T) {
	// More filters for our AttributeProfile
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// update our Account with our filters
	accPrf := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:  "cgrates.org",
			ID:      "ACCOUNT_FILTER_INDEXES",
			Weights: ";0",
			FilterIDs: []string{"fltr_for_attr", "fltr_for_attr2",
				"fltr_for_attr3", "*string:~*opts.*context:*sessions"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";15",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(1),
						},
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testVIFIdxAccountRemoveIndexes(t *testing.T) {
	var reply string
	var replyIdx []string
	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}

func testV1FIdxAccountComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, AccountS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	var replyIdx []string

	//matching for our context
	expectedIDx := []string{"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxAccountsMoreProfilesForFilters(t *testing.T) {
	// more accounts with our filters
	accPrf2 := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "ACCOUNT_FILTER_INDEXES2",
			Weights:   ";0",
			FilterIDs: []string{"fltr_for_attr2", "fltr_for_attr3"},
			Balances: map[string]*utils.APIBalance{
				"ConcreteBalance1": {
					ID:      "ConcreteBalance1",
					Weights: ";15",
					Type:    utils.MetaConcrete,
					Units:   float64(40 * time.Second),
				},
			},
		},
	}
	accPrf3 := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "ACCOUNT_FILTER_INDEXES3",
			Weights:   ";0",
			FilterIDs: []string{"fltr_for_attr", "*string:~*opts.*context:*sessions"},
			Balances: map[string]*utils.APIBalance{
				"ConcreteBalance1": {
					ID:      "ConcreteBalance1",
					Weights: ";15",
					Type:    utils.MetaConcrete,
					Units:   float64(40 * time.Second),
				},
			},
		},
	}
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var replyIdx []string
	expectedIDx := []string{"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES3",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES2",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES2",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES2",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES3",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES3",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES3",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES3",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES3",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES3"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxAccountSRemoveComputedIndexesIDs(t *testing.T) {
	//indexes will ne removed again
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	//not found
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// now we will ComputeFilterIndexes by IDs(2 of the 3 profiles)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			AccountIDs: []string{"ACCOUNT_FILTER_INDEXES", "ACCOUNT_FILTER_INDEXES2"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	expIdx := []string{"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES2",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES2",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES2",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}

	// now we will ComputeFilterIndexes of the remain profile
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			AccountIDs: []string{"ACCOUNT_FILTER_INDEXES3"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	//compute for the remain Account
	expIdx = []string{"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.*context:*sessions:ACCOUNT_FILTER_INDEXES3",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES2",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES2",
		"*prefix:*req.AnswerTime:12:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.AnswerTime:33:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES2",
		"*string:*req.Usage:123s:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES",
		"*string:*req.Subject:1004:ACCOUNT_FILTER_INDEXES3",
		"*string:*req.Subject:6774:ACCOUNT_FILTER_INDEXES3",
		"*string:*req.Subject:22312:ACCOUNT_FILTER_INDEXES3",
		"*string:*opts.Subsystems:*attributes:ACCOUNT_FILTER_INDEXES3",
		"*prefix:*req.Destinations:+0775:ACCOUNT_FILTER_INDEXES3",
		"*prefix:*req.Destinations:+442:ACCOUNT_FILTER_INDEXES3"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxAccountRemoveAccountNoIndexes(t *testing.T) {
	//as we delete our accounts, indexes will  be deleted too
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "ACCOUNT_FILTER_INDEXES",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "ACCOUNT_FILTER_INDEXES2",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "ACCOUNT_FILTER_INDEXES3",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: utils.CGRateSorg,
			ID: "fltr_for_attr"}}, &reply); err != nil {
		t.Error(err)
	}

	// Check indexes as we removed, not found for both indexes
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaAccounts}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSetActionProfileWithFltr(t *testing.T) {
	// First we will set a filter for usage
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// Get filter for checking it's existence
	var resultFltr *engine.Filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: utils.CGRateSorg, ID: "fltr_for_attr"}, &resultFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultFltr, fltr.Filter) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(fltr.Filter), utils.ToJSON(resultFltr))
	}

	// we will set anActionProfile with our filter and check the indexes
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant:    "cgrates.org",
			ID:        "REM_ACC",
			FilterIDs: []string{"fltr_for_attr", "*string:~*req.Account:1001"},
			Weight:    0,
			Targets:   map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
			Schedule:  utils.MetaASAP,
			Actions: []*engine.APAction{
				{
					ID:   "REM_BAL",
					Type: utils.MetaRemBalance,
					Diktats: []*engine.APDiktat{
						{
							Path: "MONETARY",
						},
						{
							Path: "VOICE",
						},
					},
				},
			},
		},
		APIOpts: map[string]interface{}{},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:REM_ACC",
		"*string:*req.Subject:6774:REM_ACC",
		"*string:*req.Account:1001:REM_ACC",
		"*string:*req.Subject:22312:REM_ACC",
		"*string:*opts.Subsystems:*attributes:REM_ACC",
		"*prefix:*req.Destinations:+0775:REM_ACC",
		"*prefix:*req.Destinations:+442:REM_ACC"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	// update the filter for checking the indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrSameID, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*req.Account:1001:REM_ACC",
		"*string:*req.CGRID:QWEASDZXC:REM_ACC",
		"*string:*req.CGRID:IOPJKLBNM:REM_ACC"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	// back to our initial filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testV1FIdxSetActionProfileMoreFltrsMoreIndexing(t *testing.T) {
	// More filters for our ActionProfile
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	//update our ActionProfile with our filters
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "REM_ACC",
			FilterIDs: []string{"fltr_for_attr", "*string:~*req.Account:1001",
				"fltr_for_attr3", "fltr_for_attr2"},
			Weight:   0,
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
			Schedule: utils.MetaASAP,
			Actions: []*engine.APAction{
				{
					ID:   "REM_BAL",
					Type: utils.MetaRemBalance,
					Diktats: []*engine.APDiktat{
						{
							Path: "MONETARY",
						},
						{
							Path: "VOICE",
						},
					},
				},
			},
		},
		APIOpts: map[string]interface{}{},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:REM_ACC",
		"*string:*req.Subject:6774:REM_ACC",
		"*string:*req.Subject:22312:REM_ACC",
		"*string:*opts.Subsystems:*attributes:REM_ACC",
		"*prefix:*req.Destinations:+0775:REM_ACC",
		"*prefix:*req.Destinations:+442:REM_ACC",
		"*string:*req.Usage:123s:REM_ACC",
		"*string:*req.Account:1001:REM_ACC",
		"*prefix:*req.AnswerTime:12:REM_ACC",
		"*prefix:*req.AnswerTime:33:REM_ACC"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxActionProfileRemoveIndexes(t *testing.T) {
	var reply string
	var replyIdx []string
	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}
func testV1FIdxActionProfileComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, ActionS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	var replyIdx []string

	//matching for our context
	expectedIDx := []string{"*string:*req.Subject:1004:REM_ACC",
		"*string:*req.Subject:6774:REM_ACC",
		"*string:*req.Subject:22312:REM_ACC",
		"*string:*opts.Subsystems:*attributes:REM_ACC",
		"*prefix:*req.Destinations:+0775:REM_ACC",
		"*prefix:*req.Destinations:+442:REM_ACC",
		"*string:*req.Usage:123s:REM_ACC",
		"*string:*req.Account:1001:REM_ACC",
		"*prefix:*req.AnswerTime:12:REM_ACC",
		"*prefix:*req.AnswerTime:33:REM_ACC"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}
func testV1FIdxActionMoreProfileForFilters(t *testing.T) {
	//we will add more attributes with different context for matching filters
	actPrf2 := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "TOPUP_ACC",
			FilterIDs: []string{"fltr_for_attr3", "fltr_for_attr2",
				"*string:~*req.Account:1001"},
			Weight:   0,
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
			Schedule: utils.MetaASAP,
			Actions: []*engine.APAction{
				{
					ID:   "ADD_BAL",
					Type: utils.MetaAddBalance,
					Diktats: []*engine.APDiktat{
						{
							Path:  "MONETARY",
							Value: "10",
						}},
				},
			},
		},
		APIOpts: map[string]interface{}{},
	}
	actPrf3 := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "SET_BAL",
			FilterIDs: []string{"fltr_for_attr",
				"*string:~*req.Account:1001"},
			Weight:   0,
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
			Schedule: utils.MetaASAP,
			Actions: []*engine.APAction{
				{
					ID:   "SET_BAL",
					Type: utils.MetaSetBalance,
					Diktats: []*engine.APDiktat{
						{
							Path:  "MONETARY",
							Value: "10",
						}},
				},
			},
		},
		APIOpts: map[string]interface{}{},
	}
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Account:1001:SET_BAL",
		"*string:*req.Account:1001:TOPUP_ACC",
		"*string:*req.Account:1001:REM_ACC",
		"*prefix:*req.AnswerTime:12:TOPUP_ACC",
		"*prefix:*req.AnswerTime:33:TOPUP_ACC",
		"*prefix:*req.AnswerTime:12:REM_ACC",
		"*prefix:*req.AnswerTime:33:REM_ACC",
		"*string:*req.Usage:123s:TOPUP_ACC",
		"*string:*req.Usage:123s:REM_ACC",
		"*string:*req.Subject:1004:REM_ACC",
		"*string:*req.Subject:6774:REM_ACC",
		"*string:*req.Subject:22312:REM_ACC",
		"*string:*opts.Subsystems:*attributes:REM_ACC",
		"*prefix:*req.Destinations:+0775:REM_ACC",
		"*prefix:*req.Destinations:+442:REM_ACC",
		"*string:*req.Subject:1004:SET_BAL",
		"*string:*req.Subject:6774:SET_BAL",
		"*string:*req.Subject:22312:SET_BAL",
		"*string:*opts.Subsystems:*attributes:SET_BAL",
		"*prefix:*req.Destinations:+0775:SET_BAL",
		"*prefix:*req.Destinations:+442:SET_BAL"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxActionSRemoveComputedIndexesIDs(t *testing.T) {
	//indexes will be removed again
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	//not found
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// now we will ComputeFilterIndexes by IDs(2 of the 3 profiles)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			ActionProfileIDs: []string{"TOPUP_ACC", "REM_ACC"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	expectedIDx := []string{"*string:*req.Account:1001:TOPUP_ACC",
		"*string:*req.Account:1001:REM_ACC",
		"*prefix:*req.AnswerTime:12:TOPUP_ACC",
		"*prefix:*req.AnswerTime:33:TOPUP_ACC",
		"*prefix:*req.AnswerTime:12:REM_ACC",
		"*prefix:*req.AnswerTime:33:REM_ACC",
		"*string:*req.Usage:123s:TOPUP_ACC",
		"*string:*req.Usage:123s:REM_ACC",
		"*string:*req.Subject:1004:REM_ACC",
		"*string:*req.Subject:6774:REM_ACC",
		"*string:*req.Subject:22312:REM_ACC",
		"*string:*opts.Subsystems:*attributes:REM_ACC",
		"*prefix:*req.Destinations:+0775:REM_ACC",
		"*prefix:*req.Destinations:+442:REM_ACC"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	// now we will ComputeFilterIndexes of the remain profile
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			ActionProfileIDs: []string{"SET_BAL"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	expectedIDx = []string{"*string:*req.Account:1001:SET_BAL",
		"*string:*req.Account:1001:TOPUP_ACC",
		"*string:*req.Account:1001:REM_ACC",
		"*prefix:*req.AnswerTime:12:TOPUP_ACC",
		"*prefix:*req.AnswerTime:33:TOPUP_ACC",
		"*prefix:*req.AnswerTime:12:REM_ACC",
		"*prefix:*req.AnswerTime:33:REM_ACC",
		"*string:*req.Usage:123s:TOPUP_ACC",
		"*string:*req.Usage:123s:REM_ACC",
		"*string:*req.Subject:1004:REM_ACC",
		"*string:*req.Subject:6774:REM_ACC",
		"*string:*req.Subject:22312:REM_ACC",
		"*string:*opts.Subsystems:*attributes:REM_ACC",
		"*prefix:*req.Destinations:+0775:REM_ACC",
		"*prefix:*req.Destinations:+442:REM_ACC",
		"*string:*req.Subject:1004:SET_BAL",
		"*string:*req.Subject:6774:SET_BAL",
		"*string:*req.Subject:22312:SET_BAL",
		"*string:*opts.Subsystems:*attributes:SET_BAL",
		"*prefix:*req.Destinations:+0775:SET_BAL",
		"*prefix:*req.Destinations:+442:SET_BAL"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}
func testV1FIdxActionSRemoveActionsNoIndexes(t *testing.T) {
	//as we delete our actions, indexes will  be deleted too
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SET_BAL",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "REM_ACC",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "TOPUP_ACC",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: utils.CGRateSorg,
			ID: "fltr_for_attr"}}, &reply); err != nil {
		t.Error(err)
	}

	//not found as we removed profiles
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaActions}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSetChargerSProfileWithFltr(t *testing.T) {
	// First we will set a filter for usage
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// Get filter for checking it's existence
	var resultFltr *engine.Filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: utils.CGRateSorg, ID: "fltr_for_attr"}, &resultFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultFltr, fltr.Filter) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(fltr.Filter), utils.ToJSON(resultFltr))
	}

	//we will set a ChargerProfile with our filter and check the indexes
	chg := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:    "1001_Charger",
			RunID: utils.MetaDefault,
			FilterIDs: []string{"fltr_for_attr",
				"*string:~*req.Account:1001"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		APIOpts: nil,
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:1001_Charger",
		"*string:*req.Subject:6774:1001_Charger",
		"*string:*req.Account:1001:1001_Charger",
		"*string:*req.Subject:22312:1001_Charger",
		"*string:*opts.Subsystems:*attributes:1001_Charger",
		"*prefix:*req.Destinations:+0775:1001_Charger",
		"*prefix:*req.Destinations:+442:1001_Charger"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	// update the filter for checking the indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrSameID, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*req.Account:1001:1001_Charger",
		"*string:*req.CGRID:QWEASDZXC:1001_Charger",
		"*string:*req.CGRID:IOPJKLBNM:1001_Charger"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	// back to our initial filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testV1FIdxSetChargerProfileMoreFltrsMoreIndexing(t *testing.T) {
	// More filters for our ChargerProfile
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// update our ChargerProfile with our filter
	chg := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:    "1001_Charger",
			RunID: utils.MetaDefault,
			FilterIDs: []string{"fltr_for_attr",
				"*string:~*req.Account:1001",
				"fltr_for_attr3", "fltr_for_attr2"},
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
		APIOpts: nil,
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:1001_Charger",
		"*string:*req.Subject:6774:1001_Charger",
		"*string:*req.Subject:22312:1001_Charger",
		"*string:*opts.Subsystems:*attributes:1001_Charger",
		"*prefix:*req.Destinations:+0775:1001_Charger",
		"*prefix:*req.Destinations:+442:1001_Charger",
		"*string:*req.Usage:123s:1001_Charger",
		"*string:*req.Account:1001:1001_Charger",
		"*prefix:*req.AnswerTime:12:1001_Charger",
		"*prefix:*req.AnswerTime:33:1001_Charger"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxChargerProfileRemoveIndexes(t *testing.T) {
	var reply string
	var replyIdx []string
	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}

func testV1FIdxChargerProfileComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, ChargerS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	var replyIdx []string

	//matching
	expectedIDx := []string{"*string:*req.Subject:1004:1001_Charger",
		"*string:*req.Subject:6774:1001_Charger",
		"*string:*req.Subject:22312:1001_Charger",
		"*string:*opts.Subsystems:*attributes:1001_Charger",
		"*prefix:*req.Destinations:+0775:1001_Charger",
		"*prefix:*req.Destinations:+442:1001_Charger",
		"*string:*req.Usage:123s:1001_Charger",
		"*string:*req.Account:1001:1001_Charger",
		"*prefix:*req.AnswerTime:12:1001_Charger",
		"*prefix:*req.AnswerTime:33:1001_Charger"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxChargerMoreProfileForFilters(t *testing.T) {
	// we will add more charger profiles for indexing
	chg1 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:    "1001_Charger_2",
			RunID: "NEW_ID",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"fltr_for_attr3", "fltr_for_attr2"},
			Weight: 40,
		},
		APIOpts: nil,
	}
	chg2 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			ID:    "1001_Charger_3",
			RunID: "NEW_ID2",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"fltr_for_attr"},
			Weight: 40,
		},
		APIOpts: nil,
	}
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chg1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chg2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Account:1001:1001_Charger_2",
		"*string:*req.Account:1001:1001_Charger_3",
		"*string:*req.Account:1001:1001_Charger",
		"*prefix:*req.AnswerTime:12:1001_Charger_2",
		"*prefix:*req.AnswerTime:33:1001_Charger_2",
		"*prefix:*req.AnswerTime:12:1001_Charger",
		"*prefix:*req.AnswerTime:33:1001_Charger",
		"*string:*req.Usage:123s:1001_Charger_2",
		"*string:*req.Usage:123s:1001_Charger",
		"*string:*req.Subject:1004:1001_Charger",
		"*string:*req.Subject:6774:1001_Charger",
		"*string:*req.Subject:22312:1001_Charger",
		"*string:*opts.Subsystems:*attributes:1001_Charger",
		"*prefix:*req.Destinations:+0775:1001_Charger",
		"*prefix:*req.Destinations:+442:1001_Charger",
		"*string:*req.Subject:1004:1001_Charger_3",
		"*string:*req.Subject:6774:1001_Charger_3",
		"*string:*req.Subject:22312:1001_Charger_3",
		"*string:*opts.Subsystems:*attributes:1001_Charger_3",
		"*prefix:*req.Destinations:+0775:1001_Charger_3",
		"*prefix:*req.Destinations:+442:1001_Charger_3"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxChargerSRemoveComputedIndexesIDs(t *testing.T) {
	// indexes will be removed again
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	//not found
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// now we will ComputeFilterIndexes by IDs(2 of the 3 profiles)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			ChargerIDs: []string{"1001_Charger", "1001_Charger_2"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	expectedIDx := []string{"*string:*req.Account:1001:1001_Charger_2",
		"*string:*req.Account:1001:1001_Charger",
		"*prefix:*req.AnswerTime:12:1001_Charger_2",
		"*prefix:*req.AnswerTime:33:1001_Charger_2",
		"*prefix:*req.AnswerTime:12:1001_Charger",
		"*prefix:*req.AnswerTime:33:1001_Charger",
		"*string:*req.Usage:123s:1001_Charger_2",
		"*string:*req.Usage:123s:1001_Charger",
		"*string:*req.Subject:1004:1001_Charger",
		"*string:*req.Subject:6774:1001_Charger",
		"*string:*req.Subject:22312:1001_Charger",
		"*string:*opts.Subsystems:*attributes:1001_Charger",
		"*prefix:*req.Destinations:+0775:1001_Charger",
		"*prefix:*req.Destinations:+442:1001_Charger"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	// now we will ComputeFilterIndexes of the remain profile
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			ChargerIDs: []string{"1001_Charger_3"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	expectedIDx = []string{"*string:*req.Account:1001:1001_Charger_2",
		"*string:*req.Account:1001:1001_Charger_3",
		"*string:*req.Account:1001:1001_Charger",
		"*prefix:*req.AnswerTime:12:1001_Charger_2",
		"*prefix:*req.AnswerTime:33:1001_Charger_2",
		"*prefix:*req.AnswerTime:12:1001_Charger",
		"*prefix:*req.AnswerTime:33:1001_Charger",
		"*string:*req.Usage:123s:1001_Charger_2",
		"*string:*req.Usage:123s:1001_Charger",
		"*string:*req.Subject:1004:1001_Charger",
		"*string:*req.Subject:6774:1001_Charger",
		"*string:*req.Subject:22312:1001_Charger",
		"*string:*opts.Subsystems:*attributes:1001_Charger",
		"*prefix:*req.Destinations:+0775:1001_Charger",
		"*prefix:*req.Destinations:+442:1001_Charger",
		"*string:*req.Subject:1004:1001_Charger_3",
		"*string:*req.Subject:6774:1001_Charger_3",
		"*string:*req.Subject:22312:1001_Charger_3",
		"*string:*opts.Subsystems:*attributes:1001_Charger_3",
		"*prefix:*req.Destinations:+0775:1001_Charger_3",
		"*prefix:*req.Destinations:+442:1001_Charger_3"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxChargerSRemoveActionsNoIndexes(t *testing.T) {
	//as we delete our ChargerS, indexes will  be deleted too
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "1001_Charger",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "1001_Charger_2",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "1001_Charger_3",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: utils.CGRateSorg,
			ID: "fltr_for_attr"}}, &reply); err != nil {
		t.Error(err)
	}

	//not found as we removed profiles
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaChargers}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSetRateSProfileWithFltr(t *testing.T) {
	// First we will set a filter usage
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// Get filter for checking it's existence
	var resultFltr *engine.Filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: utils.CGRateSorg, ID: "fltr_for_attr"}, &resultFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultFltr, fltr.Filter) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(fltr.Filter), utils.ToJSON(resultFltr))
	}

	//we will set a RateProfile with our filter and check the indexes
	rtPrf := utils.APIRateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		FilterIDs: []string{"fltr_for_attr",
			"*string:~*req.Account:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				FilterIDs:       []string{"*string:~*req.RequestType:*prepaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.01),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:RATE_1",
		"*string:*req.Subject:6774:RATE_1",
		"*string:*req.Account:1001:RATE_1",
		"*string:*req.Subject:22312:RATE_1",
		"*string:*opts.Subsystems:*attributes:RATE_1",
		"*prefix:*req.Destinations:+0775:RATE_1",
		"*prefix:*req.Destinations:+442:RATE_1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	// update the filter for checking the indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrSameID, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*req.Account:1001:RATE_1",
		"*string:*req.CGRID:QWEASDZXC:RATE_1",
		"*string:*req.CGRID:IOPJKLBNM:RATE_1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	// back to our initial filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testV1FIdxSetRateProfileMoreFltrsMoreIndexing(t *testing.T) {
	// More filters for our ChargerProfile
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// update our RateProfile with our filter
	rtPrf := utils.APIRateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_1",
		FilterIDs: []string{"fltr_for_attr",
			"fltr_for_attr2", "fltr_for_attr3",
			"*string:~*req.Account:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				FilterIDs:       []string{"*string:~*req.RequestType:*prepaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.01),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:RATE_1",
		"*string:*req.Subject:6774:RATE_1",
		"*string:*req.Subject:22312:RATE_1",
		"*string:*opts.Subsystems:*attributes:RATE_1",
		"*prefix:*req.Destinations:+0775:RATE_1",
		"*prefix:*req.Destinations:+442:RATE_1",
		"*string:*req.Usage:123s:RATE_1",
		"*string:*req.Account:1001:RATE_1",
		"*prefix:*req.AnswerTime:12:RATE_1",
		"*prefix:*req.AnswerTime:33:RATE_1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxRateSProfileRemoveIndexes(t *testing.T) {
	var reply string
	var replyIdx []string
	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}

func testV1FIdxRateSProfileComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, RateS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	var replyIdx []string

	//matching
	expectedIDx := []string{"*string:*req.Subject:1004:RATE_1",
		"*string:*req.Subject:6774:RATE_1",
		"*string:*req.Subject:22312:RATE_1",
		"*string:*opts.Subsystems:*attributes:RATE_1",
		"*prefix:*req.Destinations:+0775:RATE_1",
		"*prefix:*req.Destinations:+442:RATE_1",
		"*string:*req.Usage:123s:RATE_1",
		"*string:*req.Account:1001:RATE_1",
		"*prefix:*req.AnswerTime:12:RATE_1",
		"*prefix:*req.AnswerTime:33:RATE_1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxRateSProfileMoreProfilesForFilters(t *testing.T) {
	// we will add more RateProfiles for indexing
	rtPrf1 := utils.APIRateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_2",
		FilterIDs: []string{
			"fltr_for_attr2", "fltr_for_attr3",
			"*string:~*req.Account:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				FilterIDs:       []string{"*string:~*req.RequestType:*pseudoprepaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.001),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	rtPrf2 := utils.APIRateProfile{
		Tenant: "cgrates.org",
		ID:     "RATE_3",
		FilterIDs: []string{
			"fltr_for_attr",
			"*string:~*req.Account:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				FilterIDs:       nil,
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.5),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	//get indexes for all profiles
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Account:1001:RATE_2",
		"*string:*req.Account:1001:RATE_3",
		"*string:*req.Account:1001:RATE_1",
		"*prefix:*req.AnswerTime:12:RATE_2",
		"*prefix:*req.AnswerTime:33:RATE_2",
		"*prefix:*req.AnswerTime:12:RATE_1",
		"*prefix:*req.AnswerTime:33:RATE_1",
		"*string:*req.Usage:123s:RATE_2",
		"*string:*req.Usage:123s:RATE_1",
		"*string:*req.Subject:1004:RATE_1",
		"*string:*req.Subject:6774:RATE_1",
		"*string:*req.Subject:22312:RATE_1",
		"*string:*opts.Subsystems:*attributes:RATE_1",
		"*prefix:*req.Destinations:+0775:RATE_1",
		"*prefix:*req.Destinations:+442:RATE_1",
		"*string:*req.Subject:1004:RATE_3",
		"*string:*req.Subject:6774:RATE_3",
		"*string:*req.Subject:22312:RATE_3",
		"*string:*opts.Subsystems:*attributes:RATE_3",
		"*prefix:*req.Destinations:+0775:RATE_3",
		"*prefix:*req.Destinations:+442:RATE_3"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxRateSRemoveComputedIndexesIDs(t *testing.T) {
	// indexes will be removed again
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	//not found
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// now we will ComputeFilterIndexes by IDs(2 of the 3 profiles)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			RateProfileIDs: []string{"RATE_1", "RATE_2"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	expectedIDx := []string{"*string:*req.Account:1001:RATE_2",
		"*string:*req.Account:1001:RATE_1",
		"*prefix:*req.AnswerTime:12:RATE_2",
		"*prefix:*req.AnswerTime:33:RATE_2",
		"*prefix:*req.AnswerTime:12:RATE_1",
		"*prefix:*req.AnswerTime:33:RATE_1",
		"*string:*req.Usage:123s:RATE_2",
		"*string:*req.Usage:123s:RATE_1",
		"*string:*req.Subject:1004:RATE_1",
		"*string:*req.Subject:6774:RATE_1",
		"*string:*req.Subject:22312:RATE_1",
		"*string:*opts.Subsystems:*attributes:RATE_1",
		"*prefix:*req.Destinations:+0775:RATE_1",
		"*prefix:*req.Destinations:+442:RATE_1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
	// now we will ComputeFilterIndexes of the remain profile
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			RateProfileIDs: []string{"RATE_3"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	expectedIDx = []string{"*string:*req.Account:1001:RATE_2",
		"*string:*req.Account:1001:RATE_3",
		"*string:*req.Account:1001:RATE_1",
		"*prefix:*req.AnswerTime:12:RATE_2",
		"*prefix:*req.AnswerTime:33:RATE_2",
		"*prefix:*req.AnswerTime:12:RATE_1",
		"*prefix:*req.AnswerTime:33:RATE_1",
		"*string:*req.Usage:123s:RATE_2",
		"*string:*req.Usage:123s:RATE_1",
		"*string:*req.Subject:1004:RATE_1",
		"*string:*req.Subject:6774:RATE_1",
		"*string:*req.Subject:22312:RATE_1",
		"*string:*opts.Subsystems:*attributes:RATE_1",
		"*prefix:*req.Destinations:+0775:RATE_1",
		"*prefix:*req.Destinations:+442:RATE_1",
		"*string:*req.Subject:1004:RATE_3",
		"*string:*req.Subject:6774:RATE_3",
		"*string:*req.Subject:22312:RATE_3",
		"*string:*opts.Subsystems:*attributes:RATE_3",
		"*prefix:*req.Destinations:+0775:RATE_3",
		"*prefix:*req.Destinations:+442:RATE_3"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxRateSProfileRemoveRateProfileNoIndexes(t *testing.T) {
	// as we delete our RateProfile, indexes will be deleted too
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RATE_1",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RATE_2",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RATE_3",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: utils.CGRateSorg,
			ID: "fltr_for_attr"}}, &reply); err != nil {
		t.Error(err)
	}

	//not found as we removed profiles
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfiles}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxRateSProfileRatesWithFltr(t *testing.T) {
	// First we will set a filter usage
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// Get filter for checking it's existence
	var resultFltr *engine.Filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: utils.CGRateSorg, ID: "fltr_for_attr"}, &resultFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultFltr, fltr.Filter) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(fltr.Filter), utils.ToJSON(resultFltr))
	}

	//we will set a RateProfile with our filter and check the indexes
	rtPrf := utils.APIRateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID: "RT_ALWAYS",
				FilterIDs: []string{"fltr_for_attr",
					"*string:~*req.RequestType:*prepaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.01),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				FilterIDs:       []string{"*string:~*req.RequestType:*postpaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.01),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var replyIdx []string
	expectedIDx := []string{"*string:*req.RequestType:*prepaid:RT_ALWAYS",
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS",
		"*string:*req.Subject:1004:RT_ALWAYS",
		"*string:*req.Subject:6774:RT_ALWAYS",
		"*string:*req.Subject:22312:RT_ALWAYS",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS",
		"*prefix:*req.Destinations:+442:RT_ALWAYS"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	//update the filter for checking the indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrSameID, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*req.CGRID:QWEASDZXC:RT_ALWAYS",
		"*string:*req.CGRID:IOPJKLBNM:RT_ALWAYS",
		"*string:*req.RequestType:*prepaid:RT_ALWAYS",
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	//back to our initial filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testV1FIdxSetRatePofileRatesMoreFltrsMoreIndexing(t *testing.T) {
	// More filters for our RateProfileRates
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// update our RateProfileRates with our filters
	rtPrf := utils.APIRateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001", "fltr_for_attr"},
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID: "RT_ALWAYS",
				FilterIDs: []string{"fltr_for_attr", "fltr_for_attr2",
					"*string:~*req.RequestType:*prepaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.01),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				FilterIDs: []string{"fltr_for_attr3",
					"*string:~*req.RequestType:*postpaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.01),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.RequestType:*prepaid:RT_ALWAYS",
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS",
		"*string:*req.Subject:1004:RT_ALWAYS",
		"*string:*req.Subject:6774:RT_ALWAYS",
		"*string:*req.Subject:22312:RT_ALWAYS",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS",
		"*prefix:*req.Destinations:+442:RT_ALWAYS",
		"*prefix:*req.AnswerTime:12:RT_CHRISTMAS",
		"*prefix:*req.AnswerTime:33:RT_CHRISTMAS",
		"*string:*req.Usage:123s:RT_ALWAYS"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxRateProfileRatesRemoveIndexes(t *testing.T) {
	var reply string
	var replyIdx []string
	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}

func testV1FIdxRateProfileRatesComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, RateS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}
	var replyIdx []string

	//matching for our context
	expectedIDx := []string{"*string:*req.RequestType:*prepaid:RT_ALWAYS",
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS",
		"*string:*req.Subject:1004:RT_ALWAYS",
		"*string:*req.Subject:6774:RT_ALWAYS",
		"*string:*req.Subject:22312:RT_ALWAYS",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS",
		"*prefix:*req.Destinations:+442:RT_ALWAYS",
		"*prefix:*req.AnswerTime:12:RT_CHRISTMAS",
		"*prefix:*req.AnswerTime:33:RT_CHRISTMAS",
		"*string:*req.Usage:123s:RT_ALWAYS"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxRateProfileRatesMoreRatesForFilters(t *testing.T) {
	//we will add more rates in RATE_1 for matching filters and also
	// another RateProfile with new rates
	rtPrf := utils.APIRateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_1",
		FilterIDs: []string{"*string:~*req.Account:1001", "fltr_for_attr"},
		Rates: map[string]*utils.APIRate{
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				FilterIDs: []string{"fltr_for_attr3", "fltr_for_attr2",
					"*string:~*req.RequestType:*postpaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.1),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
			"RT_GOLDEN_HOUR": {
				ID:              "RT_GOLDEN_HOUR",
				FilterIDs:       []string{"fltr_for_attr"},
				ActivationTimes: "* * 1 * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.001),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfileRates,
		rtPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returne")
	}
	rtPrf2 := utils.APIRateProfile{
		Tenant:    "cgrates.org",
		ID:        "RATE_2",
		FilterIDs: []string{"*string:~*req.Account:1007", "fltr_for_attr"},
		Rates: map[string]*utils.APIRate{
			"RT_CHRISTMAS_2": {
				ID: "RT_CHRISTMAS_2",
				FilterIDs: []string{"fltr_for_attr3", "fltr_for_attr2",
					"*string:~*req.RequestType:*postpaid"},
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.1),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
			"RT_ALWAYS_2": {
				ID:              "RT_ALWAYS_2",
				FilterIDs:       []string{"fltr_for_attr", "fltr_for_attr2"},
				ActivationTimes: "* * 1 * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.001),
						FixedFee:      utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var replyIdx []string
	// first expected for RATE_1 context
	expectedIDx := []string{"*string:*req.RequestType:*prepaid:RT_ALWAYS",
		"*string:*req.RequestType:*postpaid:RT_WEEKEND",
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS",
		"*string:*req.Subject:1004:RT_ALWAYS",
		"*string:*req.Subject:6774:RT_ALWAYS",
		"*string:*req.Subject:22312:RT_ALWAYS",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS",
		"*prefix:*req.Destinations:+442:RT_ALWAYS",
		"*prefix:*req.AnswerTime:12:RT_CHRISTMAS",
		"*prefix:*req.AnswerTime:33:RT_CHRISTMAS",
		"*string:*req.Usage:123s:RT_ALWAYS",
		"*string:*req.Usage:123s:RT_WEEKEND",
		"*prefix:*req.AnswerTime:12:RT_WEEKEND",
		"*prefix:*req.AnswerTime:33:RT_WEEKEND",
		"*string:*req.Subject:1004:RT_GOLDEN_HOUR",
		"*string:*req.Subject:6774:RT_GOLDEN_HOUR",
		"*string:*req.Subject:22312:RT_GOLDEN_HOUR",
		"*string:*opts.Subsystems:*attributes:RT_GOLDEN_HOUR",
		"*prefix:*req.Destinations:+0775:RT_GOLDEN_HOUR",
		"*prefix:*req.Destinations:+442:RT_GOLDEN_HOUR"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}

	// second expected for RATE_2 context
	expectedIDx = []string{
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS_2",
		"*prefix:*req.AnswerTime:12:RT_CHRISTMAS_2",
		"*prefix:*req.AnswerTime:33:RT_CHRISTMAS_2",
		"*string:*req.Usage:123s:RT_CHRISTMAS_2",
		"*string:*req.Usage:123s:RT_ALWAYS_2",
		"*string:*req.Subject:1004:RT_ALWAYS_2",
		"*string:*req.Subject:6774:RT_ALWAYS_2",
		"*string:*req.Subject:22312:RT_ALWAYS_2",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS_2",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS_2",
		"*prefix:*req.Destinations:+442:RT_ALWAYS_2"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_2"}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxRateProfileRatesRemoveComputedIndexesIDs(t *testing.T) {
	// indexes will be removed for both context, both RateProfile
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_2"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//not found for both cases
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_2"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// now we will ComputeFilterIndexes by IDs for *RATE_1 context(but just only 1 profile, not both)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			RateProfileIDs: []string{"RATE_1"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}
	// first expected for RATE_1 context
	expectedIDx := []string{"*string:*req.RequestType:*prepaid:RT_ALWAYS",
		"*string:*req.RequestType:*postpaid:RT_WEEKEND",
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS",
		"*string:*req.Subject:1004:RT_ALWAYS",
		"*string:*req.Subject:6774:RT_ALWAYS",
		"*string:*req.Subject:22312:RT_ALWAYS",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS",
		"*prefix:*req.Destinations:+442:RT_ALWAYS",
		"*prefix:*req.AnswerTime:12:RT_CHRISTMAS",
		"*prefix:*req.AnswerTime:33:RT_CHRISTMAS",
		"*string:*req.Usage:123s:RT_ALWAYS",
		"*string:*req.Usage:123s:RT_WEEKEND",
		"*prefix:*req.AnswerTime:12:RT_WEEKEND",
		"*prefix:*req.AnswerTime:33:RT_WEEKEND",
		"*string:*req.Subject:1004:RT_GOLDEN_HOUR",
		"*string:*req.Subject:6774:RT_GOLDEN_HOUR",
		"*string:*req.Subject:22312:RT_GOLDEN_HOUR",
		"*string:*opts.Subsystems:*attributes:RT_GOLDEN_HOUR",
		"*prefix:*req.Destinations:+0775:RT_GOLDEN_HOUR",
		"*prefix:*req.Destinations:+442:RT_GOLDEN_HOUR"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
	// now we will ComputeFilterIndexes by IDs for *RATE_2 context(but just only 1 profile, not both)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg,
			RateProfileIDs: []string{"RATE_2"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}
	// second expected for RATE_2 context
	expectedIDx = []string{
		"*string:*req.RequestType:*postpaid:RT_CHRISTMAS_2",
		"*prefix:*req.AnswerTime:12:RT_CHRISTMAS_2",
		"*prefix:*req.AnswerTime:33:RT_CHRISTMAS_2",
		"*string:*req.Usage:123s:RT_CHRISTMAS_2",
		"*string:*req.Usage:123s:RT_ALWAYS_2",
		"*string:*req.Subject:1004:RT_ALWAYS_2",
		"*string:*req.Subject:6774:RT_ALWAYS_2",
		"*string:*req.Subject:22312:RT_ALWAYS_2",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS_2",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS_2",
		"*prefix:*req.Destinations:+442:RT_ALWAYS_2"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_2"}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxRateProfileRatesRemoveRateProfileRatesNoIndexes(t *testing.T) {
	//as we delete our rates from profiles, indexes will  be deleted too
	args := &utils.RemoveRPrfRates{
		Tenant:  "cgrates.org",
		ID:      "RATE_1",
		RateIDs: []string{"RT_GOLDEN_HOUR", "RT_CHRISTMAS"},
	}
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveRateProfileRates,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	// we did not remove all rates from RATE_1, so we will check the
	// indexes for the remain rates
	expectedIDx := []string{"*string:*req.RequestType:*prepaid:RT_ALWAYS",
		"*string:*req.RequestType:*postpaid:RT_WEEKEND",
		"*string:*req.Subject:1004:RT_ALWAYS",
		"*string:*req.Subject:6774:RT_ALWAYS",
		"*string:*req.Subject:22312:RT_ALWAYS",
		"*string:*opts.Subsystems:*attributes:RT_ALWAYS",
		"*prefix:*req.Destinations:+0775:RT_ALWAYS",
		"*prefix:*req.Destinations:+442:RT_ALWAYS",
		"*string:*req.Usage:123s:RT_ALWAYS",
		"*string:*req.Usage:123s:RT_WEEKEND",
		"*prefix:*req.AnswerTime:12:RT_WEEKEND",
		"*prefix:*req.AnswerTime:33:RT_WEEKEND"}
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
	//remove the remain rates,so the rate profile will be out of rates
	args = &utils.RemoveRPrfRates{
		Tenant:  "cgrates.org",
		ID:      "RATE_1",
		RateIDs: []string{"RT_WEEKEND", "RT_ALWAYS"},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveRateProfileRates,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_1"}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//for the RATE_2, we will remove the whole profile, just the rates
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RATE_2",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaRateProfileRates,
			Context: "RATE_2"}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//remove the filter(we will remove RATE_1 not not broke the reference)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RATE_1",
			Tenant: utils.CGRateSorg}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilter,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: utils.CGRateSorg,
			ID: "fltr_for_attr"}}, &reply); err != nil {
		t.Error(err)
	}
}
func testV1FIdxSetResourceSProfileWithFltr(t *testing.T) {
	// First we will set a filter for usage
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// Get filter for checking it's existence
	var resultFltr *engine.Filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilter,
		&utils.TenantID{Tenant: utils.CGRateSorg, ID: "fltr_for_attr"}, &resultFltr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultFltr, fltr.Filter) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(fltr.Filter), utils.ToJSON(resultFltr))
	}

	//we will a ResourceProfile with our filter and check the indexes
	resPrfl := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "RESOURCE1",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"fltr_for_attr"},
			UsageTTL:     time.Second,
			Limit:        1,
			Weight:       10,
			ThresholdIDs: []string{"TH1"},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		resPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	// check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:RESOURCE1",
		"*string:*req.Subject:6774:RESOURCE1",
		"*string:*req.Account:1001:RESOURCE1",
		"*string:*req.Subject:22312:RESOURCE1",
		"*string:*opts.Subsystems:*attributes:RESOURCE1",
		"*prefix:*req.Destinations:+0775:RESOURCE1",
		"*prefix:*req.Destinations:+442:RESOURCE1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaResources},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	// update the filter for checking the indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltrSameID, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*req.Account:1001:RESOURCE1",
		"*string:*req.CGRID:QWEASDZXC:RESOURCE1",
		"*string:*req.CGRID:IOPJKLBNM:RESOURCE1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaResources},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	// back to our initial filter
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testV1FIdxSetResourceSMoreFltrsMoreIndexing(t *testing.T) {
	// more filters for our ResourceProfile
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	//update our resourceProfile with our filter
	resPrfl := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "RESOURCE1",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"fltr_for_attr",
				"fltr_for_attr3", "fltr_for_attr2"},
			UsageTTL:     time.Second,
			Limit:        1,
			Weight:       10,
			ThresholdIDs: []string{"TH1"},
		},
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		resPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	//check indexes
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:RESOURCE1",
		"*string:*req.Subject:6774:RESOURCE1",
		"*string:*req.Subject:22312:RESOURCE1",
		"*string:*opts.Subsystems:*attributes:RESOURCE1",
		"*prefix:*req.Destinations:+0775:RESOURCE1",
		"*prefix:*req.Destinations:+442:RESOURCE1",
		"*string:*req.Usage:123s:RESOURCE1",
		"*string:*req.Account:1001:RESOURCE1",
		"*prefix:*req.AnswerTime:12:RESOURCE1",
		"*prefix:*req.AnswerTime:33:RESOURCE1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaResources},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxResourceSProfileRemoveIndexes(t *testing.T) {
	var reply string
	var replyIdx []string
	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaResources},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaResources},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}

func testV1FIdxResourceSProfileComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, ResourceS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	var replyIdx []string

	//matching
	expectedIDx := []string{"*string:*req.Subject:1004:RESOURCE1",
		"*string:*req.Subject:6774:RESOURCE1",
		"*string:*req.Subject:22312:RESOURCE1",
		"*string:*opts.Subsystems:*attributes:RESOURCE1",
		"*prefix:*req.Destinations:+0775:RESOURCE1",
		"*prefix:*req.Destinations:+442:RESOURCE1",
		"*string:*req.Usage:123s:RESOURCE1",
		"*string:*req.Account:1001:RESOURCE1",
		"*prefix:*req.AnswerTime:12:RESOURCE1",
		"*prefix:*req.AnswerTime:33:RESOURCE1"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, ItemType: utils.MetaResources}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
