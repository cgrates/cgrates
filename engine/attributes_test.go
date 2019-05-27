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
	attrEvs                = []*AttrArgsProcessEvent{
		{
			Context: utils.StringPointer(utils.MetaSessionS),
			CGREvent: &utils.CGREvent{ //matching AttributeProfile1
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					"Attribute":      "AttributeProfile1",
					utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
					"UsageInterval":  "1s",
					utils.Weight:     "20.0",
				},
			},
		},
		{
			Context: utils.StringPointer(utils.MetaSessionS),
			CGREvent: &utils.CGREvent{ //matching AttributeProfile2
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					"Attribute": "AttributeProfile2",
				},
			},
		},
		{
			Context: utils.StringPointer(utils.MetaSessionS),
			CGREvent: &utils.CGREvent{ //matching AttributeProfilePrefix
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					"Attribute": "AttributeProfilePrefix",
				},
			},
		},
		{
			Context: utils.StringPointer(utils.MetaSessionS),
			CGREvent: &utils.CGREvent{ //matching AttributeProfilePrefix
				Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					"DistinctMatch": 20,
				},
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
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					FieldName: utils.Account,
					Value:     config.NewRSRParsersMustCompile("1010", true, utils.INFIELD_SEP),
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
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					FieldName: utils.Account,
					Value:     config.NewRSRParsersMustCompile("1010", true, utils.INFIELD_SEP),
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
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					FieldName: utils.Account,
					Value:     config.NewRSRParsersMustCompile("1010", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
		&AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeIDMatch",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"*gte:~DistinctMatch:20"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				{
					FieldName: utils.Account,
					Value:     config.NewRSRParsersMustCompile("1010", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
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
	attrService, err = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: defaultCfg}, nil, nil, 1)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestAttributeAddFilters(t *testing.T) {
	fltrAttr1 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_1",
		Rules: []*FilterRule{
			{
				Type:      MetaString,
				FieldName: "~Attribute",
				Values:    []string{"AttributeProfile1"},
			},
			{
				Type:      MetaGreaterOrEqual,
				FieldName: "~UsageInterval",
				Values:    []string{(1 * time.Second).String()},
			},
			{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.DynamicDataPrefix + utils.Weight,
				Values:    []string{"9.0"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr1)
	fltrAttr2 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_2",
		Rules: []*FilterRule{
			{
				Type:      MetaString,
				FieldName: "~Attribute",
				Values:    []string{"AttributeProfile2"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttr2)
	fltrAttrPrefix := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_3",
		Rules: []*FilterRule{
			{
				Type:      MetaPrefix,
				FieldName: "~Attribute",
				Values:    []string{"AttributeProfilePrefix"},
			},
		},
	}
	dmAtr.SetFilter(fltrAttrPrefix)
	fltrAttr4 := &Filter{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "FLTR_ATTR_4",
		Rules: []*FilterRule{
			{
				Type:      MetaGreaterOrEqual,
				FieldName: utils.DynamicDataPrefix + utils.Weight,
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
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(atr, tempAttr) {
			t.Errorf("Expecting: %+v, received: %+v", atr, tempAttr)
		}
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
	attrEvs[0].CGREvent.Event["Account"] = "1010" //Field added in event after process
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"AttributeProfile1"},
		AlteredFields:   []string{"Account"},
		CGREvent:        attrEvs[0].CGREvent,
	}
	atrp, err := attrService.processEvent(attrEvs[0])
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply, atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(atrp))
	}
}

func TestAttributeProcessEventWithNotFound(t *testing.T) {
	attrEvs[3].CGREvent.Event["Account"] = "1010" //Field added in event after process
	if _, err := attrService.processEvent(attrEvs[3]); err == nil || err != utils.ErrNotFound {
		t.Errorf("Error: %+v", err)
	}
}

func TestAttributeProcessEventWithIDs(t *testing.T) {
	attrEvs[3].CGREvent.Event["Account"] = "1010" //Field added in event after process
	attrEvs[3].AttributeIDs = []string{"AttributeIDMatch"}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"AttributeIDMatch"},
		AlteredFields:   []string{"Account"},
		CGREvent:        attrEvs[3].CGREvent,
	}
	if atrp, err := attrService.processEvent(attrEvs[3]); err != nil {
	} else if !reflect.DeepEqual(eRply, atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(atrp))
	}
}

func TestAttributeEventReplyDigest(t *testing.T) {
	eRpl := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{utils.Account, utils.Subject},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
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
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
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
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Subject"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
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
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Subject"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAttributeSProcessEvent",
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
		FilterIDs: []string{"*string:~Account:1007"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTimeAttributes,
		},
		Attributes: []*Attribute{
			{
				FieldName: utils.Account,
				Value:     config.NewRSRParsersMustCompile("1010", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	if err := dmAtr.SetAttributeProfile(attrPrf, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringMap{
		"*string:~Account:1007": {
			"AttrPrf": true,
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
	//verify if old index was deleted ( context *any)
	if _, err := dmAtr.GetFilterIndexes(utils.PrefixToIndexCache[rfi1.itemType],
		rfi1.dbKeySuffix, MetaString, nil); err != utils.ErrNotFound {
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
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field1",
				Value:     config.NewRSRParsersMustCompile("Value1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Value:     config.NewRSRParsersMustCompile("Value2", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field3",
				Value:     config.NewRSRParsersMustCompile("Value3", true, utils.INFIELD_SEP),
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
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(4),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2", "ATTR_3"},
		AlteredFields:   []string{"Field1", "Field2", "Field3"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
				"Field3":       "Value3",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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

func TestAttributeProcessWithMultipleRuns2(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field1",
				Value:     config.NewRSRParsersMustCompile("Value1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Value:     config.NewRSRParsersMustCompile("Value2", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~NotFound:NotFound"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field3",
				Value:     config.NewRSRParsersMustCompile("Value3", true, utils.INFIELD_SEP),
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
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(4),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2"},
		AlteredFields:   []string{"Field1", "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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

func TestAttributeProcessWithMultipleRuns3(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field1",
				Value:     config.NewRSRParsersMustCompile("Value1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Value:     config.NewRSRParsersMustCompile("Value2", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field3",
				Value:     config.NewRSRParsersMustCompile("Value3", true, utils.INFIELD_SEP),
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
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(2),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2"},
		AlteredFields:   []string{"Field1", "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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

func TestAttributeProcessWithMultipleRuns4(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field1",
				Value:     config.NewRSRParsersMustCompile("Value1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Value:     config.NewRSRParsersMustCompile("Value2", true, utils.INFIELD_SEP),
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
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(4),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2"},
		AlteredFields:   []string{"Field1", "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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

func TestAttributeMultipleProcessWithBlocker(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
		return
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field1",
				Value:     config.NewRSRParsersMustCompile("Value1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Value:     config.NewRSRParsersMustCompile("Value2", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field3",
				Value:     config.NewRSRParsersMustCompile("Value3", true, utils.INFIELD_SEP),
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
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(4),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1", "ATTR_2"},
		AlteredFields:   []string{"Field1", "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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

func TestAttributeMultipleProcessWithBlocker2(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~InitialField:InitialValue"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field1",
				Value:     config.NewRSRParsersMustCompile("Value1", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	attrPrf2 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_2",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Value:     config.NewRSRParsersMustCompile("Value2", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	attrPrf3 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_3",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field2:Value2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field3",
				Value:     config.NewRSRParsersMustCompile("Value3", true, utils.INFIELD_SEP),
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
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(4),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field1"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Value:     config.NewRSRParsersMustCompile("~Field1", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1": "Value1",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1": "Value1",
				"Field2": "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       "ATTR_1",
		Contexts: []string{utils.META_ANY},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FilterIDs: []string{"*string:~PassField:Test"},
				FieldName: "PassField",
				Value:     config.NewRSRParsersMustCompile("Pass", true, utils.INFIELD_SEP),
			},
			{
				FilterIDs: []string{"*string:~PassField:RandomValue"},
				FieldName: "NotPassField",
				Value:     config.NewRSRParsersMustCompile("NotPass", true, utils.INFIELD_SEP),
			},
			{
				FilterIDs: []string{"*notexists:~RandomField:"},
				FieldName: "RandomField",
				Value:     config.NewRSRParsersMustCompile("RandomValue", true, utils.INFIELD_SEP),
			},
		},
		Weight: 10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"PassField": "Test",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"PassField", "RandomField"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"PassField":   "Pass",
				"RandomField": "RandomValue",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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

func TestAttributeProcessEventConstant(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("ConstVal", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1": "Value1",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1": "Value1",
				"Field2": "ConstVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~Field1", true, utils.INFIELD_SEP),
			},
			{
				FieldName: "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~TheField", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":   "Value1",
				"TheField": "TheVal",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":   "Value1",
				"Field2":   "TheVal",
				"TheField": "TheVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~Field1", true, utils.INFIELD_SEP),
			},
			{
				FieldName: "Field2",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("_", true, utils.INFIELD_SEP),
			},
			{
				FieldName: "Field2",
				Type:      utils.META_COMPOSED,
				Value:     config.NewRSRParsersMustCompile("~TheField", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":   "Value1",
				"TheField": "TheVal",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":   "Value1",
				"Field2":   "Value1_TheVal",
				"TheField": "TheVal",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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

func TestAttributeProcessEventSum(t *testing.T) {
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		t.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.MetaSum,
				Value:     config.NewRSRParsersMustCompile("10;~NumField;20", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":   "Value1",
				"TheField": "TheVal",
				"NumField": "20",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":   "Value1",
				"TheField": "TheVal",
				"NumField": "20",
				"Field2":   "50",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.META_USAGE_DIFFERENCE,
				Value:     config.NewRSRParsersMustCompile("~UnixTimeStamp;~UnixTimeStamp2", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":         "Value1",
				"TheField":       "TheVal",
				"UnixTimeStamp":  "1554364297",
				"UnixTimeStamp2": "1554364287",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":         "Value1",
				"TheField":       "TheVal",
				"UnixTimeStamp":  "1554364297",
				"UnixTimeStamp2": "1554364287",
				"Field2":         "10s",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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
		t.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.MetaValueExponent,
				Value:     config.NewRSRParsersMustCompile("~Multiplier;~Pow", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		t.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":     "Value1",
				"TheField":   "TheVal",
				"Multiplier": "2",
				"Pow":        "3",
			},
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"ATTR_1"},
		AlteredFields:   []string{"Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1":     "Value1",
				"TheField":   "TheVal",
				"Multiplier": "2",
				"Pow":        "3",
				"Field2":     "2000",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
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
	data, _ := NewMapStorage()
	dmAtr = NewDataManager(data)
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		b.Errorf("Error: %+v", err)
	}
	attrService, err = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: defaultCfg}, nil, nil, 1)
	if err != nil {
		b.Errorf("Error: %+v", err)
	}
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		b.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		b.Error(err)
	} else if test != true {
		b.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.META_CONSTANT,
				Value:     config.NewRSRParsersMustCompile("ConstVal", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		b.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1": "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
			b.Errorf("Error: %+v", err)
		}
	}
}

func BenchmarkAttributeProcessEventVariable(b *testing.B) {
	data, _ := NewMapStorage()
	dmAtr = NewDataManager(data)
	defaultCfg, err := config.NewDefaultCGRConfig()
	if err != nil {
		b.Errorf("Error: %+v", err)
	}
	attrService, err = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: defaultCfg}, nil, nil, 1)
	if err != nil {
		b.Errorf("Error: %+v", err)
	}
	//refresh the DM
	if err := dmAtr.DataDB().Flush(""); err != nil {
		b.Error(err)
	}
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		b.Error(err)
	} else if test != true {
		b.Errorf("\nExpecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"*string:~Field1:Value1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				FieldName: "Field2",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~Field1", true, utils.INFIELD_SEP),
			},
		},
		Blocker: true,
		Weight:  10,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(attrPrf1, true); err != nil {
		b.Error(err)
	}
	attrArgs := &AttrArgsProcessEvent{
		Context:     utils.StringPointer(utils.MetaSessionS),
		ProcessRuns: utils.IntPointer(1),
		CGREvent: &utils.CGREvent{
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Field1": "Value1",
			},
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrService.V1ProcessEvent(attrArgs, &reply); err != nil {
			b.Errorf("Error: %+v", err)
		}
	}
}
