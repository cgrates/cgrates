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

		testV1FIdxStopEngine,
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
	fltr := &engine.FilterWithAPIOpts{
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
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"fltr_for_attr"},
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
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS, ItemType: utils.MetaAttributes},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	//update the filter for checking the indexes
	fltr = &engine.FilterWithAPIOpts{
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
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	// check the updated indexes
	expectedIDx = []string{"*string:*req.CGRID:QWEASDZXC:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.CGRID:IOPJKLBNM:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS, ItemType: utils.MetaAttributes},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	// context changed, not gonna match any indexes
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers, ItemType: utils.MetaAttributes},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//back to our initial filter
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
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testV1FIdxSetAttributeSMoreFltrsMoreIndexing(t *testing.T) {
	// More filters for our AttributeProfile
	fltr1 := &engine.FilterWithAPIOpts{
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
	fltr2 := &engine.FilterWithAPIOpts{
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
	attrPrf := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"fltr_for_attr", "fltr_for_attr2", "fltr_for_attr3"},
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
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS, ItemType: utils.MetaAttributes},
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
	//indexes will not be removed because of this different context
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			ItemType: utils.MetaAttributes},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//these are not removed, so the indexes are not removed
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS, ItemType: utils.MetaAttributes},
		&replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyIdx)
		sort.Strings(expectedIDx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, replyIdx)
		}
	}

	//indexes will be removed for this specific context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS, ItemType: utils.MetaAttributes},
		&replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}
}

func testV1FIdxAttributeComputeIndexes(t *testing.T) {
	// compute our indexes
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			AttributeS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	// our indexes are sotred again, so we can get them
	// firstly, not gonna get for a different context
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			ItemType: utils.MetaAttributes}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %T, received %T", utils.ErrNotFound, err)
	}

	//matching for our context
	expectedIDx := []string{"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
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
	//we will more attributes with different context for matching filters
	attrPrf2 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_new_fltr",
			Contexts:  []string{utils.MetaChargers},
			FilterIDs: []string{"fltr_for_attr2", "fltr_for_attr3"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
			},
		},
	}
	attrPrf3 := &engine.AttributeWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTE3",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"fltr_for_attr3"},
			Attributes: []*engine.ExternalAttribute{
				{
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

	// now we will match indexes for *chargers
	var replyIdx []string
	expectedIDx := []string{"*string:*req.Usage:123s:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_new_fltr"}
	//"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_TEST_ATTRIBUTE3",
	//"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_TEST_ATTRIBUTE"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}

	// now we will match indexes for *sessions
	expectedIDx = []string{"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
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
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIDx, replyIdx)
		}
	}
}

func testV1FIdxAttributeSRemoveComputedIndexesIDs(t *testing.T) {
	//indexes will ne removed for both contexts
	var reply string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			ItemType: utils.MetaAttributes},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes, APIOpts: map[string]interface{}{
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
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			ItemType: utils.MetaAttributes}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// now we will ComputeFilterIndexes by IDs for *chargers context
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			AttributeIDs: []string{"TEST_ATTRIBUTES_new_fltr"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	//able to get indexes with context *chargers
	expIdx := []string{"*string:*req.Usage:123s:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_new_fltr"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expIdx, replyIdx)
		}
	}

	// now we will ComputeFilterIndexes by IDs for *sessions context(but just only 1 profile, not both)
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			AttributeIDs: []string{"TEST_ATTRIBUTE3"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	//able to get indexes with context *sessions
	expIdx = []string{"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE3"}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", expIdx, replyIdx)
		}
	}

	// compute for the last profile remain
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			AttributeIDs: []string{"TEST_ATTRIBUTES_IT_TEST"}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}
	expIdx = []string{"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
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
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", expIdx, replyIdx)
		}
	}
	time.Sleep(100)
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

	// Check indexes as we removed, not found for both indexes
	var replyIdx []string
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaSessionS,
			ItemType: utils.MetaAttributes}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(context.Background(), utils.AdminSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: utils.CGRateSorg, Context: utils.MetaChargers,
			ItemType: utils.MetaAttributes}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
