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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	cloneExpTimeAttributes time.Time
	expTimeAttributes      = time.Now().Add(time.Duration(20 * time.Minute))
	attrService            *AttributeService
	dmAtr                  *DataManager
	mapSubstitutes         = map[string]map[interface{}]*Attribute{
		utils.Account: map[interface{}]*Attribute{
			utils.META_ANY: &Attribute{
				FieldName:  utils.Account,
				Initial:    utils.META_ANY,
				Substitute: "1010",
				Append:     true,
			},
		},
	}
	attrEvs = []*utils.CGREvent{
		&utils.CGREvent{ //matching AttributeProfile1
			Tenant:  config.CgrConfig().DefaultTenant,
			ID:      utils.GenUUID(),
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				"Attribute":      "AttributeProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "20.0",
			},
		},
		&utils.CGREvent{ //matching AttributeProfile2
			Tenant:  config.CgrConfig().DefaultTenant,
			ID:      utils.GenUUID(),
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				"Attribute": "AttributeProfile2",
			},
		},
		&utils.CGREvent{ //matching AttributeProfilePrefix
			Tenant:  config.CgrConfig().DefaultTenant,
			ID:      utils.GenUUID(),
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				"Attribute": "AttributeProfilePrefix",
			},
		},
	}

	atrPs = AttributeProfiles{
		&AttributeProfile{
			Tenant:    config.CgrConfig().DefaultTenant,
			ID:        "AttributeProfile1",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_ATTR_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				&Attribute{
					FieldName:  utils.Account,
					Initial:    utils.META_ANY,
					Substitute: "1010",
					Append:     true,
				},
			},
			Weight:     20,
			attributes: mapSubstitutes,
		},
		&AttributeProfile{
			Tenant:    config.CgrConfig().DefaultTenant,
			ID:        "AttributeProfile2",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_ATTR_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				&Attribute{
					FieldName:  utils.Account,
					Initial:    utils.META_ANY,
					Substitute: "1010",
					Append:     true,
				},
			},
			Weight:     20,
			attributes: mapSubstitutes,
		},
		&AttributeProfile{
			Tenant:    config.CgrConfig().DefaultTenant,
			ID:        "AttributeProfilePrefix",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_ATTR_3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				&Attribute{
					FieldName:  utils.Account,
					Initial:    utils.META_ANY,
					Substitute: "1010",
					Append:     true,
				},
			},
			attributes: mapSubstitutes,
			Weight:     20,
		},
	}
)

func TestAttributePopulateAttrService(t *testing.T) {
	//Need clone because time.Now adds extra information that DeepEqual doesn't like
	if err := utils.Clone(expTimeAttributes, &cloneExpTimeAttributes); err != nil {
		t.Error(err)
	}
	data, _ := NewMapStorage()
	dmAtr = NewDataManager(data)
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	attrService, err = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: defaultCfg}, nil, nil)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestAttributeAddFilters(t *testing.T) {
	fltrAttr1 := &Filter{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     "FLTR_ATTR_1",
		Rules: []*FilterRule{
			&FilterRule{
				Type:      MetaString,
				FieldName: "Attribute",
				Values:    []string{"AttributeProfile1"},
			},
			&FilterRule{
				Type:      MetaGreaterOrEqual,
				FieldName: "UsageInterval",
				Values:    []string{(1 * time.Second).String()},
			},
			&FilterRule{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.Weight,
				Values:    []string{"9.0"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr1)
	fltrAttr2 := &Filter{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     "FLTR_ATTR_2",
		Rules: []*FilterRule{
			&FilterRule{
				Type:      MetaString,
				FieldName: "Attribute",
				Values:    []string{"AttributeProfile2"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr2)
	fltrAttrPrefix := &Filter{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     "FLTR_ATTR_3",
		Rules: []*FilterRule{
			&FilterRule{
				Type:      MetaPrefix,
				FieldName: "Attribute",
				Values:    []string{"AttributeProfilePrefix"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttrPrefix)
	fltrAttr4 := &Filter{
		Tenant: config.CgrConfig().DefaultTenant,
		ID:     "FLTR_ATTR_4",
		Rules: []*FilterRule{
			&FilterRule{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.Weight,
				Values:    []string{"200.00"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr4)
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
			false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(atr, tempAttr) {
			t.Errorf("Expecting: %+v, received: %+v", atr, tempAttr)
		}
	}
}

func TestAttributeMatchingAttributeProfilesForEvent(t *testing.T) {
	atrp, err := attrService.matchingAttributeProfilesForEvent(attrEvs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrp[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", atrPs[0], atrp[0])
	}
	atrp, err = attrService.matchingAttributeProfilesForEvent(attrEvs[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[1], atrp[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs), utils.ToJSON(atrp))
	}
	atrp, err = attrService.matchingAttributeProfilesForEvent(attrEvs[2])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[2], atrp[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", atrPs[2], atrp[0])
	}
}

func TestAttributeProfileForEvent(t *testing.T) {
	atrp, err := attrService.attributeProfileForEvent(attrEvs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[0]), utils.ToJSON(atrp))
	}

	atrp, err = attrService.attributeProfileForEvent(attrEvs[1])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[1], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[1]), utils.ToJSON(atrp))
	}

	atrp, err = attrService.attributeProfileForEvent(attrEvs[2])
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
		MatchedProfile: "AttributeProfile1",
		AlteredFields:  []string{"Account"},
		CGREvent:       attrEvs[0],
	}
	atrp, err := attrService.processEvent(attrEvs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply, atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(atrp))
	}
}

func TestAttributeEventReplyDigest(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{utils.Account, utils.Subject},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:      "1001",
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
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:      "1001",
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
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{"Subject"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:      "1001",
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
		MatchedProfile: "ATTR_1",
		AlteredFields:  []string{"Subject"},
		CGREvent: &utils.CGREvent{
			Tenant:  "cgrates.org",
			ID:      "testAttributeSProcessEvent",
			Context: utils.StringPointer(utils.MetaSessionS),
			Event: map[string]interface{}{
				utils.Account:      "1001",
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
	data, _ := NewMapStorage()
	dmAtr = NewDataManager(data)
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		Contexts:  []string{utils.META_ANY},
		FilterIDs: []string{"*string:Account:1007"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTimeAttributes,
		},
		Attributes: []*Attribute{
			&Attribute{
				FieldName:  utils.Account,
				Initial:    utils.META_ANY,
				Substitute: "1001",
				Append:     true,
			},
		},
		Weight: 20,
	}
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*string:Account:1007": utils.StringMap{
			"AttrPrf": true,
		},
	}
	reverseIdxes := map[string]utils.StringMap{
		"AttrPrf": utils.StringMap{
			"*string:Account:1007": true,
		},
	}
	rfi1 := NewFilterIndexer(dmAtr, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(attrPrf.Tenant, utils.META_ANY))
	if rcvIdx, err := dmAtr.GetFilterIndexes(utils.PrefixToIndexCache[rfi1.itemType],
		rfi1.dbKeySuffix, MetaString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dmAtr.GetFilterReverseIndexes(
		utils.PrefixToRevIndexCache[rfi1.itemType], rfi1.dbKeySuffix, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
		}
	}
	//Set AttributeProfile with new context (*sessions)
	attrPrf.Contexts = []string{utils.MetaSessionS}
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	rfi2 := NewFilterIndexer(dmAtr, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(attrPrf.Tenant, utils.MetaSessionS))
	if rcvIdx, err := dmAtr.GetFilterIndexes(utils.PrefixToIndexCache[rfi2.itemType],
		rfi2.dbKeySuffix, MetaString, nil); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	if reverseRcvIdx, err := dmAtr.GetFilterReverseIndexes(
		utils.PrefixToRevIndexCache[rfi2.itemType], rfi2.dbKeySuffix, nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reverseIdxes, reverseRcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", reverseIdxes, reverseRcvIdx)
	}
	//verify if old index was deleted ( context *any)
	if _, err := dmAtr.GetFilterIndexes(utils.PrefixToIndexCache[rfi1.itemType],
		rfi1.dbKeySuffix, MetaString, nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err := dmAtr.GetFilterReverseIndexes(
		utils.PrefixToRevIndexCache[rfi1.itemType], rfi1.dbKeySuffix, nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}
