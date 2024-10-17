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
package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	expTimeAttributes = time.Now().Add(20 * time.Minute)
	attrS             *AttributeService
	dmAtr             *DataManager
	attrEvs           = []*utils.CGREvent{
		{ //matching AttributeProfile1
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Attribute":      "AttributeProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfile2
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Attribute": "AttributeProfile2",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfilePrefix
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Attribute": "AttributeProfilePrefix",
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfilePrefix
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"DistinctMatch": 20,
			},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	atrPs = AttributeProfiles{
		&AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeProfile1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_ATTR_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
		&AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeProfile2",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_ATTR_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
		&AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeProfilePrefix",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_ATTR_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
		&AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeIDMatch",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*gte:~*req.DistinctMatch:20"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     expTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
					Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
				},
			},
			Weight: 20,
		},
	}
)

func TestAttributePopulateAttrService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	cfg.AttributeSCfg().StringIndexedFields = nil
	cfg.AttributeSCfg().PrefixIndexedFields = nil
	data := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	Cache.Clear(nil)
}

func TestAttributeAddFilters(t *testing.T) {
	fltrAttr1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_1",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfile1"},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req.UsageInterval",
				Values:  []string{(time.Second).String()},
			},
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req." + utils.Weight,
				Values:  []string{"9.0"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr1, true)
	fltrAttr2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_2",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaString,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfile2"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr2, true)
	fltrAttrPrefix := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_3",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaPrefix,
				Element: "~*req.Attribute",
				Values:  []string{"AttributeProfilePrefix"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttrPrefix, true)
	fltrAttr4 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_4",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterOrEqual,
				Element: "~*req." + utils.Weight,
				Values:  []string{"200.00"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr4, true)
}

func TestAttributeCache(t *testing.T) {
	for _, atr := range atrPs {
		if err = dmAtr.SetAttributeProfile(atr, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each attribute from cache
	for _, atr := range atrPs {
		if tempAttr, err := dmAtr.GetAttributeProfile(atr.Tenant, atr.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(atr, tempAttr) {
			t.Errorf("Expecting: %+v, received: %+v", atr, tempAttr)
		}
	}
}

func TestAttributeProfileForEvent(t *testing.T) {
	context := utils.StringPointer(utils.IfaceAsString(attrEvs[0].APIOpts[utils.OptsContext]))
	var attrIDs []string
	if attrIDs, err = utils.IfaceAsSliceString(attrEvs[0].APIOpts[utils.OptsAttributesProfileIDs]); err != nil {
		t.Error(err)
	}
	atrp, err := attrS.attributeProfileForEvent(attrEvs[0].Tenant, context,
		attrIDs, attrEvs[0].Time, utils.MapStorage{
			utils.MetaReq:  attrEvs[0].Event,
			utils.MetaOpts: attrEvs[0].APIOpts,
			utils.MetaVars: utils.MapStorage{
				utils.OptsAttributesProcessRuns: 0,
			},
		}, utils.EmptyString, make(map[string]int), 0, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[0]), utils.ToJSON(atrp))
	}
	context = utils.StringPointer(utils.IfaceAsString(attrEvs[1].APIOpts[utils.OptsContext]))
	if attrIDs, err = utils.IfaceAsSliceString(attrEvs[1].APIOpts[utils.OptsAttributesProfileIDs]); err != nil {
		t.Error(err)
	}
	atrp, err = attrS.attributeProfileForEvent(attrEvs[1].Tenant, context,
		attrIDs, attrEvs[1].Time, utils.MapStorage{
			utils.MetaReq:  attrEvs[1].Event,
			utils.MetaOpts: attrEvs[1].APIOpts,
			utils.MetaVars: utils.MapStorage{
				utils.OptsAttributesProcessRuns: 0,
			},
		}, utils.EmptyString, make(map[string]int), 0, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[1], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[1]), utils.ToJSON(atrp))
	}
	context = utils.StringPointer(utils.IfaceAsString(attrEvs[2].APIOpts[utils.OptsContext]))
	if attrIDs, err = utils.IfaceAsSliceString(attrEvs[2].APIOpts[utils.OptsAttributesProfileIDs]); err != nil {
		t.Error(err)
	}
	atrp, err = attrS.attributeProfileForEvent(attrEvs[2].Tenant, context,
		attrIDs, attrEvs[2].Time, utils.MapStorage{
			utils.MetaReq:  attrEvs[2].Event,
			utils.MetaOpts: attrEvs[2].APIOpts,
			utils.MetaVars: utils.MapStorage{
				utils.OptsAttributesProcessRuns: 0,
			},
		}, utils.EmptyString, make(map[string]int), 0, false)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[2], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[2]), utils.ToJSON(atrp))
	}
}

func TestAttributeProcessEvent(t *testing.T) {
	attrEvs[0].Event["Account"] = "1010" //Field added in event after process
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:AttributeProfile1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Account"},
		CGREvent:        attrEvs[0],
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  attrEvs[0].Event,
		utils.MetaOpts: attrEvs[0].APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	atrp, err := attrS.processEvent(attrEvs[0].Tenant, attrEvs[0], eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply, atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(atrp))
	}
}

func TestAttributeProcessEventWithNotFound(t *testing.T) {
	attrEvs[3].Event["Account"] = "1010" //Field added in event after process
	eNM := utils.MapStorage{
		utils.MetaReq:  attrEvs[3].Event,
		utils.MetaOpts: attrEvs[3].APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	if _, err := attrS.processEvent(attrEvs[0].Tenant, attrEvs[3], eNM,
		newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err == nil || err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestAttributeProcessEventWithIDs(t *testing.T) {
	attrEvs[3].Event["Account"] = "1010" //Field added in event after process
	attrEvs[3].APIOpts[utils.OptsAttributesProfileIDs] = []string{"AttributeIDMatch"}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:AttributeIDMatch"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Account"},
		CGREvent:        attrEvs[3],
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  attrEvs[3].Event,
		utils.MetaOpts: attrEvs[3].APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	if atrp, err := attrS.processEvent(attrEvs[0].Tenant, attrEvs[3], eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err != nil {
	} else if !reflect.DeepEqual(eRply, atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(atrp))
	}
}

func TestAttributeEventReplyDigest(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.AccountField, utils.Subject},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destinations: "+491511231234",
			},
		},
	}
	expRpl := "Account:1001,Subject:1001"
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeEventReplyDigest2(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destinations: "+491511231234",
			},
		},
	}
	expRpl := ""
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeEventReplyDigest3(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{"*req.Subject"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destinations: "+491511231234",
			},
		},
	}
	expRpl := "Subject:1001"
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeEventReplyDigest4(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{"*req.Subject"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destinations: "+491511231234",
			},
		},
	}
	expRpl := ""
	val := eRpl.Digest()
	if !reflect.DeepEqual(val, expRpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expRpl), utils.ToJSON(val))
	}
}

func TestAttributeIndexer(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		Contexts:  []string{utils.MetaAny},
		FilterIDs: []string{"*string:~*req.Account:1007"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     expTimeAttributes,
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Account:1007": {
			"AttrPrf": struct{}{},
		},
	}
	if rcvIdx, err := dmAtr.GetIndexes(utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey(attrPrf.Tenant, utils.MetaAny), "", false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	//Set AttributeProfile with new context (*sessions)
	cpAttrPrf := new(AttributeProfile)
	*cpAttrPrf = *attrPrf
	cpAttrPrf.Contexts = []string{utils.MetaSessionS}
	if err := dmAtr.SetAttributeProfile(cpAttrPrf, true); err != nil {
		t.Error(err)
	}
	if rcvIdx, err := dmAtr.GetIndexes(utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey(attrPrf.Tenant, utils.MetaSessionS), "", false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	//verify if old index was deleted ( context *any)
	if _, err := dmAtr.GetIndexes(utils.CacheAttributeFilterIndexes,
		utils.ConcatenatedKey(attrPrf.Tenant, utils.MetaAny), "", false, false); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestAttributeProcessWithMultipleRuns1(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2", "cgrates.org:ATTR_3", "cgrates.org:ATTR_2"},
		AlteredFields: []string{
			utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2",
			utils.MetaReq + utils.NestingSep + "Field3",
		},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
				"Field3":       "Value3",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Fatalf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessWithMultipleRuns2(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.NotFound:NotFound"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2", "cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessWithMultipleRuns3(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessWithMultipleRuns4(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2", "cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMultipleProcessWithBlocker(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
		return
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMultipleProcessWithBlocker2(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field1"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessValue(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1": "Value1",
				"Field2": "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeAttributeFilterIDs(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       "ATTR_1",
		Contexts: []string{utils.MetaAny},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FilterIDs: []string{"*string:~*req.PassField:Test"},
				Path:      utils.MetaReq + utils.NestingSep + "PassField",
				Value:     config.NewRSRParsersMustCompile("Pass", utils.InfieldSep),
			},
			{
				FilterIDs: []string{"*string:~*req.PassField:RandomValue"},
				Path:      utils.MetaReq + utils.NestingSep + "NotPassField",
				Value:     config.NewRSRParsersMustCompile("NotPass", utils.InfieldSep),
			},
			{
				FilterIDs: []string{"*notexists:~*req.RandomField:"},
				Path:      utils.MetaReq + utils.NestingSep + "RandomField",
				Value:     config.NewRSRParsersMustCompile("RandomValue", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"PassField": "Test",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "PassField",
			utils.MetaReq + utils.NestingSep + "RandomField"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"PassField":   "Pass",
				"RandomField": "RandomValue",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventConstant(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("ConstVal", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1": "Value1",
				"Field2": "ConstVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventVariable(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("~*req.TheField", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":   "Value1",
			"TheField": "TheVal",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":   "Value1",
				"Field2":   "TheVal",
				"TheField": "TheVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventComposed(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("_", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.TheField", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":   "Value1",
			"TheField": "TheVal",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":   "Value1",
				"Field2":   "Value1_TheVal",
				"TheField": "TheVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Fatalf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventSum(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaSum,
				Value: config.NewRSRParsersMustCompile("10;~*req.NumField;20", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":   "Value1",
			"TheField": "TheVal",
			"NumField": "20",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":   "Value1",
				"TheField": "TheVal",
				"NumField": "20",
				"Field2":   "50",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventUsageDifference(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaUsageDifference,
				Value: config.NewRSRParsersMustCompile("~*req.UnixTimeStamp;~*req.UnixTimeStamp2", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":         "Value1",
			"TheField":       "TheVal",
			"UnixTimeStamp":  "1554364297",
			"UnixTimeStamp2": "1554364287",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":         "Value1",
				"TheField":       "TheVal",
				"UnixTimeStamp":  "1554364297",
				"UnixTimeStamp2": "1554364287",
				"Field2":         "10s",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeProcessEventValueExponent(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaValueExponent,
				Value: config.NewRSRParsersMustCompile("~*req.Multiplier;~*req.Pow", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1":     "Value1",
			"TheField":   "TheVal",
			"Multiplier": "2",
			"Pow":        "3",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"Field1":     "Value1",
				"TheField":   "TheVal",
				"Multiplier": "2",
				"Pow":        "3",
				"Field2":     "2000",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func BenchmarkAttributeProcessEventConstant(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		b.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		b.Error(err)
	} else if test != true {
		b.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("ConstVal", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		b.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
			b.Errorf("Error: %+v", err)
		}
	}
}

func BenchmarkAttributeProcessEventVariable(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)

	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		b.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		b.Error(err)
	} else if test != true {
		b.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("~*req.Field1", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		b.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
			b.Errorf("Error: %+v", err)
		}
	}
}

func TestGetAttributeProfileFromInline(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrID := "*sum:*req.Field2:10&~*req.NumField&20"
	expAttrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       attrID,
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Type:  utils.MetaSum,
			Value: config.NewRSRParsersMustCompile("10;~*req.NumField;20", utils.InfieldSep),
		}},
	}
	attr, err := dm.GetAttributeProfile(config.CgrConfig().GeneralCfg().DefaultTenant, attrID, false, false, "")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}

func TestProcessAttributeConstant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_CONSTANT",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("Val2", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_CONSTANT
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeConstant",
		Event: map[string]any{
			"Field1":     "Val1",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev.Event["Field2"] = "Val2"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_CONSTANT"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        ev,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeVariable(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_VARIABLE",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VARIABLE
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeVariable",
		Event: map[string]any{
			"Field1":      "Val1",
			"RandomField": "Val2",
			utils.Weight:  "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "Val2"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_VARIABLE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeComposed(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_COMPOSED",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaComposed,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField2", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_COMPOSED
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeComposed",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "Val2",
			"RandomField2": "Concatenated",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "Val2Concatenated"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_COMPOSED"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeUsageDifference(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_USAGE_DIFF",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaUsageDifference,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField;~*req.RandomField2", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_USAGE_DIFF
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUsageDifference",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1514808000",
			"RandomField2": "1514804400",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "1h0m0s"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_USAGE_DIFF"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeSum(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_SUM",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaSum,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField;~*req.RandomField2;10", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_SUM
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeSum",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "16"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_SUM"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeDiff(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_DIFF",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaDifference,
				Value: config.NewRSRParsersMustCompile("55;~*req.RandomField;~*req.RandomField2;10", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_DIFF
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeDiff",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "39"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_DIFF"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeMultiply(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_MULTIPLY",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaMultiply,
				Value: config.NewRSRParsersMustCompile("55;~*req.RandomField;~*req.RandomField2;10", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_MULTIPLY
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeMultiply",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "2750"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_MULTIPLY"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeDivide(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_DIVIDE",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaDivide,
				Value: config.NewRSRParsersMustCompile("55.0;~*req.RandomField;~*req.RandomField2;4", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_DIVIDE
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeDivide",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "2.75"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_DIVIDE"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeValueExponent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_VAL_EXP",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaValueExponent,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField2;4", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "50000"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_VAL_EXP"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeUnixTimeStamp(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_UNIX_TIMESTAMP",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaUnixTimestamp,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField2", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_UNIX_TIMESTAMP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUnixTimeStamp",
		Event: map[string]any{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "2013-12-30T15:00:01Z",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "1388415601"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_UNIX_TIMESTAMP"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributePrefix(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_PREFIX",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.ATTR:ATTR_PREFIX"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaPrefix,
				Value: config.NewRSRParsersMustCompile("abc_", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]any{
			"ATTR":       "ATTR_PREFIX",
			"Field2":     "Val2",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "abc_Val2"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_PREFIX"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeSuffix(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_SUFFIX",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.ATTR:ATTR_SUFFIX"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaSuffix,
				Value: config.NewRSRParsersMustCompile("_abc", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]any{
			"ATTR":       "ATTR_SUFFIX",
			"Field2":     "Val2",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	rcv, err := attrS.processEvent(ev.Tenant, ev, eNM, newDynamicDP(nil, nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "Val2_abc"
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_SUFFIX"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent:        clnEv,
	}
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestProcessAttributeHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		responses := map[string]struct {
			code  int
			reply string
		}{
			"/passwd":  {code: http.StatusOK, reply: "passwd1"},
			"/account": {code: http.StatusOK, reply: "1001"},
		}
		if val, has := responses[r.URL.EscapedPath()]; has {
			w.WriteHeader(val.code)
			fmt.Fprint(w, val.reply)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}))
	defer ts.Close()
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)

	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)

	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_HTTP",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Value1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:  utils.MetaHTTP + utils.HashtagSep + utils.IdxStart + ts.URL + "/account" + utils.IdxEnd,
				Value: config.NewRSRParsersMustCompile("*attributes", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Password",
				Type:  utils.MetaHTTP + utils.HashtagSep + utils.IdxStart + ts.URL + "/passwd" + utils.IdxEnd,
				Value: config.NewRSRParsersMustCompile("*attributes", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Field1": "Value1",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_HTTP"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + utils.AccountField, utils.MetaReq + utils.NestingSep + "Password"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     "TestProcessAttributeHTTP",
			Event: map[string]any{
				"Field1":   "Value1",
				"Account":  "1001",
				"Password": "passwd1",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !slices.Equal(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	if sort.Slice(reply.AlteredFields, func(i, j int) bool {
		return reply.AlteredFields[i] < reply.AlteredFields[j]
	}); !slices.Equal(reply.AlteredFields, eRply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}

}

func TestAttributeIndexSelectsFalse(t *testing.T) {
	// change the IndexedSelects to false
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	cfg.AttributeSCfg().StringIndexedFields = nil
	cfg.AttributeSCfg().PrefixIndexedFields = nil
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)

	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		Contexts:  []string{utils.MetaCDRs},
		FilterIDs: []string{"*string:~*req.Account:1007"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     expTimeAttributes,
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}

	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"Account": "1007",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 1,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}

	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected not found, reveiced: %+v", err)
	}

}

func TestProcessAttributeWithSameWeight(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Opts.ProcessRuns = 1
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~*req.Field1:Val1"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("~*req.RandomField", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(attrPrf2, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_UNIX_TIMESTAMP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUnixTimeStamp",
		Event: map[string]any{
			"Field1":      "Val1",
			"RandomField": "1",
			utils.Weight:  "20.0",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	var rcv AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), ev, &rcv); err != nil {
		t.Errorf("Error: %+v", err)
	}
	clnEv := ev.Clone()
	clnEv.Event["Field2"] = "1"
	clnEv.Event["Field3"] = "1"
	eRply := AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2", utils.MetaReq + utils.NestingSep + "Field3"},
		CGREvent:        clnEv,
	}
	sort.Strings(rcv.MatchedProfiles)
	sort.Strings(rcv.AlteredFields)
	if !reflect.DeepEqual(eRply, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(rcv))
	}
}

func TestAttributeMultipleProcessWithFiltersExists(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf1Exists := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1_EXISTS",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*exists:~*req.InitialField:"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2Exists := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2_EXISTS",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*exists:~*req.Field1:"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1Exists, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2Exists, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dmAtr.GetAttributeProfile(attrPrf1Exists.Tenant, attrPrf1Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dmAtr.GetAttributeProfile(attrPrf2Exists.Tenant, attrPrf2Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1_EXISTS", "cgrates.org:ATTR_2_EXISTS", "cgrates.org:ATTR_1_EXISTS", "cgrates.org:ATTR_2_EXISTS"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: 4,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMultipleProcessWithFiltersNotEmpty(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf1NotEmpty := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1_NOTEMPTY",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*notempty:~*req.InitialField:"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field1",
				Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
			},
		},
		Weight: 10,
	}
	attrPrf2NotEmpty := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2_NOTEMPTY",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*notempty:~*req.Field1:"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1NotEmpty, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(attrPrf2NotEmpty, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dmAtr.GetAttributeProfile(attrPrf1NotEmpty.Tenant, attrPrf1NotEmpty.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dmAtr.GetAttributeProfile(attrPrf2NotEmpty.Tenant, attrPrf2NotEmpty.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	attrArgs := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1_NOTEMPTY", "cgrates.org:ATTR_2_NOTEMPTY", "cgrates.org:ATTR_1_NOTEMPTY", "cgrates.org:ATTR_2_NOTEMPTY"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: 4,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), attrArgs, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfiles, reply.MatchedProfiles) {
		t.Errorf("Expecting %+v, received: %+v", eRply.MatchedProfiles, reply.MatchedProfiles)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply.AlteredFields, reply.AlteredFields) {
		t.Errorf("Expecting %+v, received: %+v", eRply.AlteredFields, reply.AlteredFields)
	}
	if !reflect.DeepEqual(eRply.CGREvent.Event, reply.CGREvent.Event) {
		t.Errorf("Expecting %+v, received: %+v", eRply.CGREvent.Event, reply.CGREvent.Event)
	}
}

func TestAttributeMetaTenant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	dm := NewDataManager(NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items), config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dm, &FilterS{dm: dm, cfg: cfg}, cfg)
	attr1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_TNT",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{},
		Attributes: []*Attribute{{
			Type:  utils.MetaPrefix,
			Path:  utils.MetaTenant,
			Value: config.NewRSRParsersMustCompile("prfx_", utils.InfieldSep),
		}},
		Weight: 10,
	}

	// Add attribute in DM
	if err := dm.SetAttributeProfile(attr1, true); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(attr1.Tenant, attr1.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		Event:  map[string]any{},
		APIOpts: map[string]any{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_TNT"},
		AlteredFields:   []string{utils.MetaTenant},
		CGREvent: &utils.CGREvent{
			Tenant: "prfx_" + config.CgrConfig().GeneralCfg().DefaultTenant,
			Event:  map[string]any{},
			APIOpts: map[string]any{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %s, received: %s", utils.ToJSON(eRply), utils.ToJSON(reply))
	}
}

func TestAttributesPorcessEventMatchingProcessRuns(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().Enabled = true
	cfg.AttributeSCfg().IndexedSelects = false
	db := NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltrS := NewFilterS(cfg, nil, dm)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "Process_Runs_Fltr",
		Rules: []*FilterRule{
			/*
				{
					Type: utils.MetaString,
					Element: "~*req.Account",
					Values: []string{"pc_test"},
				},

			*/
			{
				Type:    utils.MetaGreaterThan,
				Element: "~*vars.*processRuns",
				Values:  []string{"1"},
			},
		},
	}
	if err := dm.SetFilter(fltr, true); err != nil {
		t.Error(err)
	}

	attrPfr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ProcessRuns",
		Contexts:  []string{"*any"},
		FilterIDs: []string{"Process_Runs_Fltr"},
		Attributes: []*Attribute{
			{
				Path:  "*req.CompanyName",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("ITSYS COMMUNICATIONS SRL", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// this I'll match first, no fltr and processRuns will be 1
	attrPfr2 := &AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "ATTR_MatchSecond",
		Contexts: []string{"*any"},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep),
			},
		},
		Weight: 10,
	}

	attrPfr.Compile()
	fltr.Compile()
	attrPfr2.Compile()
	if err := dm.SetAttributeProfile(attrPfr, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetAttributeProfile(attrPfr2, true); err != nil {
		t.Error(err)
	}

	attr := NewAttributeService(dm, fltrS, cfg)

	args := &utils.CGREvent{
		Event: map[string]any{
			"Account":     "pc_test",
			"CompanyName": "MY_company_will_be_changed",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	reply := &AttrSProcessEventReply{}
	expReply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_MatchSecond", "cgrates.org:ATTR_ProcessRuns"},
		AlteredFields:   []string{"*req.CompanyName", "*req.Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				"Account":     "pc_test",
				"CompanyName": "ITSYS COMMUNICATIONS SRL",
				"Password":    "CGRateS.org",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProcessRuns: 2,
				utils.OptsContext:               utils.MetaSessionS,
			},
		},
	}
	if err := attr.V1ProcessEvent(context.Background(), args, reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply.AlteredFields); !reflect.DeepEqual(expReply, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expReply), utils.ToJSON(reply))
	}
}

func TestAttributeMultipleProfileRunns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	dm := NewDataManager(NewInternalDB(nil, nil, true, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dm, &FilterS{dm: dm, cfg: cfg}, cfg)
	attrPrf1Exists := &AttributeProfile{
		Tenant:    cfg.GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{},
		Attributes: []*Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field1",
			Value: config.NewRSRParsersMustCompile("Value1", utils.InfieldSep),
		}},
		Weight: 10,
	}
	attrPrf2Exists := &AttributeProfile{
		Tenant:    cfg.GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{},
		Attributes: []*Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
		}},
		Weight: 5,
	}
	// Add attribute in DM
	if err := dm.SetAttributeProfile(attrPrf1Exists, true); err != nil {
		t.Error(err)
	}
	if err = dm.SetAttributeProfile(attrPrf2Exists, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dm.GetAttributeProfile(attrPrf1Exists.Tenant, attrPrf1Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(attrPrf2Exists.Tenant, attrPrf2Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	args := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileRuns: 2,
			utils.OptsAttributesProcessRuns: 40,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply := AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2", "cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     args.ID,
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileRuns: 2,
				utils.OptsAttributesProcessRuns: 40,
				utils.OptsContext:               utils.MetaSessionS,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(reply))
	}

	args = &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]any{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]any{
			utils.OptsAttributesProfileRuns: 1,
			utils.OptsAttributesProcessRuns: 40,
			utils.OptsContext:               utils.MetaSessionS,
		},
	}
	eRply = AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     args.ID,
			Event: map[string]any{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileRuns: 1,
				utils.OptsAttributesProcessRuns: 40,
				utils.OptsContext:               utils.MetaSessionS,
			},
		},
	}
	reply = AttrSProcessEventReply{}
	if err := attrS.V1ProcessEvent(context.Background(), args, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(reply))
	}
}
