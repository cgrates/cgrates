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
			Event: map[string]interface{}{
				"Attribute":      "AttributeProfile1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
				"UsageInterval":  "1s",
				utils.Weight:     "20.0",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfile2
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Attribute": "AttributeProfile2",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfilePrefix
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"Attribute": "AttributeProfilePrefix",
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
		{ //matching AttributeProfilePrefix
			Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:     utils.GenUUID(),
			Event: map[string]interface{}{
				"DistinctMatch": 20,
			},
			APIOpts: map[string]interface{}{
				utils.OptsContext:               utils.MetaSessionS,
				utils.OptsAttributesProcessRuns: 0,
			},
		},
	}
	atrPs = AttributeProfiles{
		&AttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "AttributeProfile1",
			FilterIDs: []string{"FLTR_ATTR_1", "*string:~*opts.*context:*sessions"},
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
			FilterIDs: []string{"FLTR_ATTR_2", "*string:~*opts.*context:*sessions"},
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
			FilterIDs: []string{"FLTR_ATTR_3", "*string:~*opts.*context:*sessions"},
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
			FilterIDs: []string{"*gte:~*req.DistinctMatch:20"},
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

func TestAttributesV1GetAttributeForEventErr(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	var ev utils.CGREvent
	rply := &APIAttributeProfile{}
	alS.Shutdown()
	err = alS.V1GetAttributeForEvent(context.Background(), &ev, rply)
	if err == nil || err != utils.ErrNotFound {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ErrNotFound, err)
	}

}
func TestAttributePopulateAttrService(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().StringIndexedFields = nil
	cfg.AttributeSCfg().PrefixIndexedFields = nil
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
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
	dmAtr.SetFilter(context.Background(), fltrAttr1, true)
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
	dmAtr.SetFilter(context.Background(), fltrAttr2, true)
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
	dmAtr.SetFilter(context.Background(), fltrAttrPrefix, true)
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
	dmAtr.SetFilter(context.Background(), fltrAttr4, true)
}

func TestAttributeCache(t *testing.T) {
	for _, atr := range atrPs {
		if err = dmAtr.SetAttributeProfile(context.TODO(), atr, true); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//verify each attribute from cache
	for _, atr := range atrPs {
		if tempAttr, err := dmAtr.GetAttributeProfile(context.TODO(), atr.Tenant, atr.ID,
			true, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(atr, tempAttr) {
			t.Errorf("Expecting: %+v, received: %+v", atr, tempAttr)
		}
	}
}

func TestAttributeProfileForEvent(t *testing.T) {
	attrIDs, err := utils.OptAsStringSlice(attrEvs[0].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}
	atrp, err := attrS.attributeProfileForEvent(context.TODO(), attrEvs[0].Tenant,
		attrIDs, utils.MapStorage{
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

	attrIDs, err = utils.OptAsStringSlice(attrEvs[1].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}
	atrp, err = attrS.attributeProfileForEvent(context.TODO(), attrEvs[1].Tenant,
		attrIDs, utils.MapStorage{
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

	attrIDs, err = utils.OptAsStringSlice(attrEvs[2].APIOpts, utils.OptsAttributesProfileIDs)
	if err != nil {
		t.Fatal(err)
	}
	atrp, err = attrS.attributeProfileForEvent(context.TODO(), attrEvs[2].Tenant,
		attrIDs, utils.MapStorage{
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
	atrp, err := attrS.processEvent(context.TODO(), attrEvs[0].Tenant, attrEvs[0], eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	if _, err := attrS.processEvent(context.TODO(), attrEvs[0].Tenant, attrEvs[3], eNM,
		newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err == nil || err != utils.ErrNotFound {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
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
	if atrp, err := attrS.processEvent(context.TODO(), attrEvs[0].Tenant, attrEvs[3], eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err != nil {
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
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
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
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
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
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Subject:      "1001",
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
			Event: map[string]interface{}{
				utils.AccountField: "1001",
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
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	expTimeStr := expTimeAttributes.Format("2006-01-02T15-04-05Z")
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"*string:~*req.Account:1007", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z|" + expTimeStr},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	eIdxes := map[string]utils.StringSet{
		"*string:*req.Account:1007": {
			"AttrPrf": struct{}{},
		},
	}
	if rcvIdx, err := dmAtr.GetIndexes(context.TODO(), utils.CacheAttributeFilterIndexes,
		attrPrf.Tenant, "", utils.NonTransactional, false, false); err != nil {
		t.Error(err)
	} else {
		if !reflect.DeepEqual(eIdxes, rcvIdx) {
			t.Errorf("Expecting %+v, received: %+v", eIdxes, rcvIdx)
		}
	}
	//Set AttributeProfile with new context (*sessions)
	cpAttrPrf := new(AttributeProfile)
	*cpAttrPrf = *attrPrf
	cpAttrPrf.FilterIDs = append(attrPrf.FilterIDs, "*string:~*opts.*context:*sessions")
	eIdxes["*string:*opts.*context:*sessions"] = utils.StringSet{
		"AttrPrf": {},
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), cpAttrPrf, true); err != nil {
		t.Error(err)
	}
	if rcvIdx, err := dmAtr.GetIndexes(context.TODO(), utils.CacheAttributeFilterIndexes,
		attrPrf.Tenant, "", utils.NonTransactional, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eIdxes, rcvIdx) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eIdxes), utils.ToJSON(rcvIdx))
	}

	expected := map[string]utils.StringSet{
		"*string:*opts.*context:*sessions": {
			"AttrPrf": {},
		},
		"*string:*req.Account:1007": {
			"AttrPrf": {},
		},
	}
	//verify if old index was deleted ( context *any)
	if rcv, err := dmAtr.GetIndexes(context.TODO(), utils.CacheAttributeFilterIndexes,
		attrPrf.Tenant, "", utils.NonTransactional, false, false); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", expected, rcv)
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
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
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
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
				"Field3":       "Value3",
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.NotFound:NotFound", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2", "cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
	Cache.Clear(nil)
	if test, err := dmAtr.DataDB().IsDBEmpty(); err != nil {
		t.Error(err)
	} else if test != true {
		t.Errorf("Expecting: true got :%+v", test)
	}
	attrPrf1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2", "cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Value: config.NewRSRParsersMustCompile("Value3", utils.InfieldSep),
			},
		},
		Weight: 30,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf3, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field1"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1": "Value1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "ATTR_1",
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"PassField": "Test",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "PassField",
			utils.MetaReq + utils.NestingSep + "RandomField"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1": "Value1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1":   "Value1",
			"TheField": "TheVal",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1":   "Value1",
			"TheField": "TheVal",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1":   "Value1",
			"TheField": "TheVal",
			"NumField": "20",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1":         "Value1",
			"TheField":       "TheVal",
			"UnixTimeStamp":  "1554364297",
			"UnixTimeStamp2": "1554364287",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1":     "Value1",
			"TheField":   "TheVal",
			"Multiplier": "2",
			"Pow":        "3",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		b.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1": "Value1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
			b.Errorf("Error: %+v", err)
		}
	}
}

func BenchmarkAttributeProcessEventVariable(b *testing.B) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
		FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1, true); err != nil {
		b.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Field1": "Value1",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	var reply AttrSProcessEventReply
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     attrID,
		Attributes: []*Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Type:  utils.MetaSum,
			Value: config.NewRSRParsersMustCompile("10;~*req.NumField;20", utils.InfieldSep),
		}},
	}
	attr, err := NewDataManager(nil, nil, nil).GetAttributeProfile(context.TODO(), config.CgrConfig().GeneralCfg().DefaultTenant, attrID, false, false, "")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}

func TestProcessAttributeConstant(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_CONSTANT",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_CONSTANT
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeConstant",
		Event: map[string]interface{}{
			"Field1":     "Val1",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_VARIABLE",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VARIABLE
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeVariable",
		Event: map[string]interface{}{
			"Field1":      "Val1",
			"RandomField": "Val2",
			utils.Weight:  "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_COMPOSED",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_COMPOSED
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeComposed",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "Val2",
			"RandomField2": "Concatenated",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_USAGE_DIFF",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_USAGE_DIFF
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUsageDifference",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "1514808000",
			"RandomField2": "1514804400",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_SUM",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_SUM
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeSum",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_DIFF",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_DIFF
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeDiff",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_MULTIPLY",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_MULTIPLY
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeMultiply",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_DIVIDE",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_DIVIDE
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeDivide",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_VAL_EXP",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "5",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_UNIX_TIMESTAMP",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_UNIX_TIMESTAMP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUnixTimeStamp",
		Event: map[string]interface{}{
			"Field1":       "Val1",
			"RandomField":  "1",
			"RandomField2": "2013-12-30T15:00:01Z",
			utils.Weight:   "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_PREFIX",
		FilterIDs: []string{"*string:~*req.ATTR:ATTR_PREFIX", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]interface{}{
			"ATTR":       "ATTR_PREFIX",
			"Field2":     "Val2",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_SUFFIX",
		FilterIDs: []string{"*string:~*req.ATTR:ATTR_SUFFIX", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_VAL_EXP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeValueExponent",
		Event: map[string]interface{}{
			"ATTR":       "ATTR_SUFFIX",
			"Field2":     "Val2",
			utils.Weight: "20.0",
		},
		APIOpts: map[string]interface{}{
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
	rcv, err := attrS.processEvent(context.TODO(), ev.Tenant, ev, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
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

func TestAttributeIndexSelectsFalse(t *testing.T) {
	// change the IndexedSelects to false
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().StringIndexedFields = nil
	cfg.AttributeSCfg().PrefixIndexedFields = nil
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
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
	expTimeStr := expTimeAttributes.Format("2006-01-02T15:04:05Z")
	attrPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf",
		FilterIDs: []string{"*string:~*req.Account:1007", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|" + expTimeStr, "*string:~*opts.*context:*cdrs"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.AccountField,
				Value: config.NewRSRParsersMustCompile("1010", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"Account": "1007",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}

	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err == nil || err != utils.ErrNotFound {
		t.Errorf("Expected not found, reveiced: %+v", err)
	}

}

func TestProcessAttributeWithSameWeight(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*string:~*req.Field1:Val1", "*string:~*opts.*context:*sessions"},
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
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf, true); err != nil {
		t.Error(err)
	}
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf2, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{ //matching ATTR_UNIX_TIMESTAMP
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     "TestProcessAttributeUnixTimeStamp",
		Event: map[string]interface{}{
			"Field1":      "Val1",
			"RandomField": "1",
			utils.Weight:  "20.0",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	var rcv AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &rcv); err != nil {
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf1Exists := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1_EXISTS",
		FilterIDs: []string{"*exists:~*req.InitialField:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*exists:~*req.Field1:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1Exists, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2Exists, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf1Exists.Tenant, attrPrf1Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf2Exists.Tenant, attrPrf2Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1_EXISTS", "cgrates.org:ATTR_2_EXISTS", "cgrates.org:ATTR_1_EXISTS", "cgrates.org:ATTR_2_EXISTS"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dmAtr = NewDataManager(data, config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dmAtr, &FilterS{dm: dmAtr, cfg: cfg}, cfg)
	attrPrf1NotEmpty := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_1_NOTEMPTY",
		FilterIDs: []string{"*notempty:~*req.InitialField:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
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
		FilterIDs: []string{"*notempty:~*req.Field1:", "*ai:*now:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	// Add attribute in DM
	if err := dmAtr.SetAttributeProfile(context.TODO(), attrPrf1NotEmpty, true); err != nil {
		t.Error(err)
	}
	if err = dmAtr.SetAttributeProfile(context.TODO(), attrPrf2NotEmpty, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf1NotEmpty.Tenant, attrPrf1NotEmpty.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dmAtr.GetAttributeProfile(context.TODO(), attrPrf2NotEmpty.Tenant, attrPrf2NotEmpty.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsContext:               utils.MetaSessionS,
			utils.OptsAttributesProcessRuns: 4,
		},
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1_NOTEMPTY", "cgrates.org:ATTR_2_NOTEMPTY", "cgrates.org:ATTR_1_NOTEMPTY", "cgrates.org:ATTR_2_NOTEMPTY"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
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
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), config.CgrConfig().CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dm, &FilterS{dm: dm, cfg: cfg}, cfg)
	attr1 := &AttributeProfile{
		Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:        "ATTR_TNT",
		FilterIDs: []string{"*string:~*opts.*context:*sessions"},
		Attributes: []*Attribute{{
			Type:  utils.MetaPrefix,
			Path:  utils.MetaTenant,
			Value: config.NewRSRParsersMustCompile("prfx_", utils.InfieldSep),
		}},
		Weight: 10,
	}

	// Add attribute in DM
	if err := dm.SetAttributeProfile(context.TODO(), attr1, true); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.TODO(), attr1.Tenant, attr1.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			utils.OptsContext: utils.MetaSessionS,
		},
	}
	eRply := AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_TNT"},
		AlteredFields:   []string{utils.MetaTenant},
		CGREvent: &utils.CGREvent{
			Tenant: "prfx_" + config.CgrConfig().GeneralCfg().DefaultTenant,
			Event:  map[string]interface{}{},
			APIOpts: map[string]interface{}{
				utils.OptsContext: utils.MetaSessionS,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
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
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	fltrS := NewFilterS(cfg, nil, dm)
	fltr := &Filter{
		Tenant: "cgrates.org",
		ID:     "Process_Runs_Fltr",
		Rules: []*FilterRule{
			{
				Type:    utils.MetaGreaterThan,
				Element: "~*vars.*processRuns",
				Values:  []string{"1"},
			},
		},
	}
	if err := dm.SetFilter(context.Background(), fltr, true); err != nil {
		t.Error(err)
	}

	attrPfr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ProcessRuns",
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
		Tenant: "cgrates.org",
		ID:     "ATTR_MatchSecond",
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
	if err := dm.SetAttributeProfile(context.Background(), attrPfr, true); err != nil {
		t.Error(err)
	}
	if err := dm.SetAttributeProfile(context.Background(), attrPfr2, true); err != nil {
		t.Error(err)
	}

	attr := NewAttributeService(dm, fltrS, cfg)

	ev := &utils.CGREvent{
		Event: map[string]interface{}{
			"Account":     "pc_test",
			"CompanyName": "MY_company_will_be_changed",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	reply := &AttrSProcessEventReply{}
	expReply := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_MatchSecond", "cgrates.org:ATTR_ProcessRuns"},
		AlteredFields:   []string{"*req.CompanyName", "*req.Password"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				"Account":     "pc_test",
				"CompanyName": "ITSYS COMMUNICATIONS SRL",
				"Password":    "CGRateS.org",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
	}
	if err := attr.V1ProcessEvent(context.Background(), ev, reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply.AlteredFields); !reflect.DeepEqual(expReply, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expReply), utils.ToJSON(reply))
	}
}

func TestAttributeMultipleProfileRunns(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	dm := NewDataManager(NewInternalDB(nil, nil, cfg.DataDbCfg().Items), cfg.CacheCfg(), nil)
	Cache.Clear(nil)
	attrS = NewAttributeService(dm, &FilterS{dm: dm, cfg: cfg}, cfg)
	attrPrf1Exists := &AttributeProfile{
		Tenant:    cfg.GeneralCfg().DefaultTenant,
		ID:        "ATTR_1",
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
		FilterIDs: []string{},
		Attributes: []*Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Value: config.NewRSRParsersMustCompile("Value2", utils.InfieldSep),
		}},
		Weight: 5,
	}
	// Add attribute in DM
	if err := dm.SetAttributeProfile(context.TODO(), attrPrf1Exists, true); err != nil {
		t.Error(err)
	}
	if err = dm.SetAttributeProfile(context.TODO(), attrPrf2Exists, true); err != nil {
		t.Errorf("Error: %+v", err)
	}
	// Add attribute in DM
	if _, err := dm.GetAttributeProfile(context.TODO(), attrPrf1Exists.Tenant, attrPrf1Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if _, err := dm.GetAttributeProfile(context.TODO(), attrPrf2Exists.Tenant, attrPrf2Exists.ID, true, false, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProfileRuns: 2,
			utils.OptsAttributesProcessRuns: 40,
		},
	}
	eRply := AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2", "cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     ev.ID,
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProfileRuns: 2,
				utils.OptsAttributesProcessRuns: 40,
			},
		},
	}
	var reply AttrSProcessEventReply
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(reply))
	}

	ev = &utils.CGREvent{
		Tenant: cfg.GeneralCfg().DefaultTenant,
		ID:     utils.GenUUID(),
		Event: map[string]interface{}{
			"InitialField": "InitialValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProfileRuns: 1,
			utils.OptsAttributesProcessRuns: 40,
		},
	}
	eRply = AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_1", "cgrates.org:ATTR_2"},
		AlteredFields: []string{utils.MetaReq + utils.NestingSep + "Field1",
			utils.MetaReq + utils.NestingSep + "Field2"},
		CGREvent: &utils.CGREvent{
			Tenant: cfg.GeneralCfg().DefaultTenant,
			ID:     ev.ID,
			Event: map[string]interface{}{
				"InitialField": "InitialValue",
				"Field1":       "Value1",
				"Field2":       "Value2",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProfileRuns: 1,
				utils.OptsAttributesProcessRuns: 40,
			},
		},
	}
	reply = AttrSProcessEventReply{}
	if err := attrS.V1ProcessEvent(context.TODO(), ev, &reply); err != nil {
		t.Errorf("Error: %+v", err)
	}
	sort.Strings(reply.AlteredFields)
	if !reflect.DeepEqual(eRply, reply) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(eRply), utils.ToJSON(reply))
	}
}

func TestAttributesV1ProcessEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	expected := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR_CHANGE_TENANT_FROM_USER", "adrian.itsyscom.com.co.uk:ATTR_MATCH_TENANT"},
		AlteredFields:   []string{"*req.Account", "*req.Password", "*tenant"},
		CGREvent: &utils.CGREvent{
			Tenant: "adrian.itsyscom.com.co.uk",
			ID:     "123",
			Event: map[string]interface{}{
				utils.AccountField: "andrei.itsyscom.com",
				"Password":         "CGRATES.ORG",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 2,
			},
		},
		blocker: false,
	}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	sort.Strings(rply.AlteredFields)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestAttributesV1ProcessEventErrorMetaSum(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaSum,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaDifference(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaDifference,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: NotEnoughParameters"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
	}

}

func TestAttributesV1ProcessEventErrorMetaValueExponent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	Cache.Clear(nil)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaValueExponent,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &AttrSProcessEventReply{}
	err = alS.V1ProcessEvent(context.Background(), ev, rply)
	sort.Strings(rply.AlteredFields)
	expErr := "SERVER_ERROR: invalid arguments <[{\"Rules\":\"CGRATES.ORG\"}]> to *valueExponent"
	if err == nil || err.Error() != expErr {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expErr, err)
	}

}

func TestAttributesattributeProfileForEventNoDBConn(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cfg:   cfg,
		dm:    dm,
		fltrS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_2",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""
	alS.dm = nil

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_3"}, evNm, lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNoDatabaseConn {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNoDatabaseConn, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrNotFound(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cfg:   cfg,
		dm:    dm,
		fltrS: NewFilterS(cfg, nil, dm),
	}

	apNil := &AttributeProfile{}
	err = alS.dm.SetAttributeProfile(context.Background(), apNil, true)
	if err != nil {
		t.Error(err)
	}

	tnt := ""
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_3"}, evNm, lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrNotFound {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesattributeProfileForEventErrPass(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	dataDB := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(dataDB, cfg.CacheCfg(), nil)
	Cache = NewCacheS(cfg, dm, nil)
	alS := &AttributeService{
		cfg:   cfg,
		dm:    dm,
		fltrS: NewFilterS(cfg, nil, dm),
	}

	postpaid, err := config.NewRSRParsers(utils.MetaPostpaid, utils.InfieldSep)
	if err != nil {
		t.Error(err)
	}
	ap := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_1",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap, true)
	if err != nil {
		t.Error(err)
	}

	tnt := "cgrates.org"
	evNm := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			utils.AccountField: "1001",
		},
		utils.MetaVars: utils.MapStorage{},
	}
	lastID := ""

	evNm = utils.MapStorage{
		utils.MetaReq:  1,
		utils.MetaVars: utils.MapStorage{},
	}

	if rcv, err := alS.attributeProfileForEvent(context.Background(), tnt, []string{"ATTR_1"}, evNm, lastID, make(map[string]int), 0, false); err == nil || err != utils.ErrWrongPath {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", utils.ErrWrongPath, err)
	} else if rcv != nil {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", nil, rcv)
	}
}

func TestAttributesParseAttributeSIPCID(t *testing.T) {
	exp := "12345;1001;1002"
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1002",
			"from": "1001",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	exp = "12345;1001;1002;1003"
	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":   "12345",
			"to":    "1001",
			"from":  "1002",
			"extra": "1003",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.extra;~*req.from",
		utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":   "12345",
			"to":    "1002",
			"from":  "1001",
			"extra": "1003",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from",
		utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Fatal(err)
	} else if exp != out {
		t.Errorf("Expected %q, Received %q", exp, out)
	}

	dp = utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid": "12345",
		},
	}
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.extra;~*req.to;~*req.from", utils.
		InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != utils.ErrNotFound {
		t.Errorf("Expected <%+v>, received <%+v>", utils.ErrNotFound, err)
	}
}

func TestAttributesParseAttributeSIPCIDWrongPathErr(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
		utils.MetaOpts: 13,
	}
	value := config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from;~*opts.WrongPath", utils.InfieldSep)
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString, utils.InfieldSep); err == nil ||
		err.Error() != utils.ErrWrongPath.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrWrongPath, err)
	}
}

func TestAttributesParseAttributeSIPCIDNotFoundErr(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"to":   "1001",
			"from": "1002",
		},
	}
	value := config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep)
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString, utils.InfieldSep); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func TestAttributesParseAttributeSIPCIDInvalidArguments(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"to":   "1001",
			"from": "1002",
		},
	}
	value := config.RSRParsers{}
	experr := `invalid number of arguments <[]> to *sipcid`
	if _, err := ParseAttribute(dp, utils.MetaSIPCID, utils.EmptyString, value,
		0, time.UTC.String(), utils.EmptyString, utils.InfieldSep); err == nil ||
		err.Error() != experr {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", experr, err)
	}
}

func TestAttributesV1ProcessEventMultipleRuns1(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	alS := NewAttributeService(dm, filterS, cfg)

	postpaid := config.NewRSRParsersMustCompile(utils.MetaPostpaid, utils.InfieldSep)
	pw := config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)

	ap1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR1",
		FilterIDs: []string{"*notexists:~*vars.*processedProfileIDs[<~*vars.*apTenantID>]:"},
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR2",
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event: map[string]interface{}{
			"Password": "passwd",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 4,
			utils.OptsAttributesProfileIDs:  []string{"ATTR1", "ATTR2"},
		},
	}
	reply := &AttrSProcessEventReply{}
	exp := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR2", "cgrates.org:ATTR1", "cgrates.org:ATTR2"},
		AlteredFields:   []string{"*req.Password", "*req.RequestType"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]interface{}{
				"Password":        "CGRateS.org",
				utils.RequestType: utils.MetaPostpaid,
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProfileIDs:  []string{"ATTR1", "ATTR2"},
				utils.OptsAttributesProcessRuns: 4,
			},
		},
	}

	if err := alS.V1ProcessEvent(context.Background(), ev, reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply.AlteredFields)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(reply))
		}
	}
}

func TestAttributesV1ProcessEventMultipleRuns2(t *testing.T) {
	tmp := Cache
	defer func() {
		Cache = tmp
	}()

	cfg := config.NewDefaultCGRConfig()
	cfg.AttributeSCfg().IndexedSelects = false
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	Cache = NewCacheS(cfg, dm, nil)
	alS := NewAttributeService(dm, filterS, cfg)

	postpaid := config.NewRSRParsersMustCompile(utils.MetaPostpaid, utils.InfieldSep)
	pw := config.NewRSRParsersMustCompile("CGRateS.org", utils.InfieldSep)
	paypal := config.NewRSRParsersMustCompile("cgrates@paypal.com", utils.InfieldSep)

	ap1 := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     "ATTR1",
		Attributes: []*Attribute{
			{
				Path:  "*req.Password",
				Type:  utils.MetaConstant,
				Value: pw,
			},
		},
		Weight: 10,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap1, true)
	if err != nil {
		t.Error(err)
	}

	ap2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR2",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR1]:"},
		Attributes: []*Attribute{
			{
				Path:  "*req.RequestType",
				Type:  utils.MetaConstant,
				Value: postpaid,
			},
		},
		Weight: 20,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap2, true)
	if err != nil {
		t.Error(err)
	}

	ap3 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR3",
		FilterIDs: []string{"*exists:~*vars.*processedProfileIDs[cgrates.org:ATTR2]:"},
		Attributes: []*Attribute{
			{
				Path:  "*req.PaypalAccount",
				Type:  utils.MetaConstant,
				Value: paypal,
			},
		},
		Weight: 30,
	}
	err = alS.dm.SetAttributeProfile(context.Background(), ap3, true)
	if err != nil {
		t.Error(err)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AttrProcessEventMultipleRuns",
		Event:  map[string]interface{}{},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 3,
		},
	}

	reply := &AttrSProcessEventReply{}
	exp := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:ATTR1", "cgrates.org:ATTR2", "cgrates.org:ATTR3"},
		AlteredFields:   []string{"*req.Password", "*req.PaypalAccount", "*req.RequestType"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AttrProcessEventMultipleRuns",
			Event: map[string]interface{}{
				"Password":        "CGRateS.org",
				"PaypalAccount":   "cgrates@paypal.com",
				utils.RequestType: utils.MetaPostpaid,
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProcessRuns: 3,
			},
		},
	}
	if err := alS.V1ProcessEvent(context.Background(), ev, reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply.AlteredFields)
		if !reflect.DeepEqual(reply, exp) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(exp), utils.ToJSON(reply))
		}
	}
}

func TestAttributesV1GetAttributeForEvent(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
		},
	}
	rply := &APIAttributeProfile{}
	expected := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*ExternalAttribute{
			{
				Path:  "*tenant",
				Type:  "*variable",
				Value: "~*req.Account:s/(.*)@(.*)/${1}.${2}/",
			},
			{
				Path:  "*req.Account",
				Type:  "*variable",
				Value: "~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/",
			},
			{
				Path:  "*tenant",
				Type:  "*composed",
				Value: ".co.uk",
			},
		},
		Weight: 20,
	}

	err = alS.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func TestAttributesV1GetAttributeForEventErrorBoolOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
			utils.MetaProfileIgnoreFilters:  time.Second,
		},
	}
	rply := &APIAttributeProfile{}

	err = alS.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <cannot convert field: 1s to bool>, \nReceived <%+v>", err)
	}

}

func TestAttributesV1GetAttributeForEventErrorNil(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	rply := &APIAttributeProfile{}

	err = alS.V1GetAttributeForEvent(context.Background(), nil, rply)
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [CGREvent]" {
		t.Errorf("\nExpected <MANDATORY_IE_MISSING: [CGREvent]>, \nReceived <%+v>", err)
	}

}

func TestAttributesV1GetAttributeForEventErrOptsI(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	attr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_CHANGE_TENANT_FROM_USER",
		FilterIDs: []string{"*string:~*req.Account:dan@itsyscom.com|adrian@itsyscom.com"},
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(.*)@(.*)/${1}.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*req.Account",
				Type:      "*variable",
				Value:     config.NewRSRParsersMustCompile("~*req.Account:s/(dan)@(.*)/${1}.${2}/:s/(adrian)@(.*)/andrei.${2}/", utils.InfieldSep),
			},
			{
				FilterIDs: nil,
				Path:      "*tenant",
				Type:      "*composed",
				Value:     config.NewRSRParsersMustCompile(".co.uk", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}
	err := dm.SetAttributeProfile(context.Background(), attr, true)
	if err != nil {
		t.Error(err)
	}

	attr2 := &AttributeProfile{
		Tenant: "adrian.itsyscom.com.co.uk",
		ID:     "ATTR_MATCH_TENANT",
		Attributes: []*Attribute{
			{
				FilterIDs: nil,
				Path:      "*req.Password",
				Type:      utils.MetaConstant,
				Value:     config.NewRSRParsersMustCompile("CGRATES.ORG", utils.InfieldSep),
			},
		},
		Blocker: false,
		Weight:  20,
	}

	err = dm.SetAttributeProfile(context.Background(), attr2, true)
	if err != nil {
		t.Error(err)
	}

	alS := NewAttributeService(dm, filterS, cfg)
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "123",
		Event: map[string]interface{}{
			utils.AccountField: "adrian@itsyscom.com",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsAttributesProfileIDs:  time.Second,
		},
	}
	rply := &APIAttributeProfile{}

	err = alS.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err == nil || err.Error() != "cannot convert field: 1s to []string" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cannot convert field: 1s to []string", err)
	}

}
func TestAttributesProcessEventProfileIgnoreFilters(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, cfg)
	cfg.AttributeSCfg().Opts.ProfileIgnoreFilters = []*utils.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	acPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Attribute:testAttrValue"},
	}
	if err := dm.SetAttributeProfile(context.Background(), acPrf, true); err != nil {
		t.Error(err)
	}
	//should match the attr profile for event because the option is false but the filter matches
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		Event: map[string]interface{}{
			"Attribute": "testAttrValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProfileIDs: []string{"AC1"},
			utils.MetaProfileIgnoreFilters: false,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	exp2 := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:AC1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AcProcessEvent",
			Event: map[string]interface{}{
				"Attribute": "testAttrValue",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProfileIDs: []string{"AC1"},
				utils.MetaProfileIgnoreFilters: false,
			},
		},
	}
	if rcv2, err := aA.processEvent(context.Background(), args2.Tenant, args2, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv2, exp2) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp2), utils.ToJSON(rcv2))
	}
	//should match the attr profile for event because the option is true even if the filter doesn't match
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent2",
		Event: map[string]interface{}{
			"Attribute": "testAttrValue2",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProfileIDs: []string{"AC1"},
			utils.MetaProfileIgnoreFilters: true,
		},
	}
	eNM2 := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	exp := &AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:AC1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "AcProcessEvent2",
			Event: map[string]interface{}{
				"Attribute": "testAttrValue2",
			},
			APIOpts: map[string]interface{}{
				utils.OptsAttributesProfileIDs: []string{"AC1"},
				utils.MetaProfileIgnoreFilters: true,
			},
		},
	}
	if rcv, err := aA.processEvent(context.Background(), args.Tenant, args, eNM2, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM2), utils.EmptyString, make(map[string]int), 0); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestAttributesV1GetAttributeForEventProfileIgnoreOpts(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	cfg.FilterSCfg().ResourceSConns = []string{}
	conMng := NewConnManager(cfg)
	db := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(db, nil, conMng)
	filterS := NewFilterS(cfg, conMng, dm)
	aA := NewAttributeService(dm, filterS, cfg)
	cfg.AttributeSCfg().Opts.ProfileIgnoreFilters = []*utils.DynamicBoolOpt{
		{
			Value: true,
		},
	}
	acPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Attribute:testAttrValue"},
	}
	if err := dm.SetAttributeProfile(context.Background(), acPrf, true); err != nil {
		t.Error(err)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent1",
		Event: map[string]interface{}{
			"Attribute": "testAttrValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsAttributesProfileIDs:  []string{"AC1"},
			utils.MetaProfileIgnoreFilters:  false,
		},
	}
	rply := &APIAttributeProfile{}
	expected := &APIAttributeProfile{
		Tenant:     "cgrates.org",
		ID:         "AC1",
		FilterIDs:  []string{"*string:~*req.Attribute:testAttrValue"},
		Attributes: []*ExternalAttribute{},
		Weight:     0,
	}

	err = aA.V1GetAttributeForEvent(context.Background(), ev, rply)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected, rply) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	// correct filter but ignore filters opt on false
	ev2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent2",
		Event: map[string]interface{}{
			"Attribute": "testAttrValue2",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProcessRuns: 2,
			utils.OptsAttributesProfileIDs:  []string{"AC1"},
			utils.MetaProfileIgnoreFilters:  true,
		},
	}
	rply2 := &APIAttributeProfile{}
	expected2 := &APIAttributeProfile{
		Tenant:     "cgrates.org",
		ID:         "AC1",
		FilterIDs:  []string{"*string:~*req.Attribute:testAttrValue"},
		Attributes: []*ExternalAttribute{},
		Weight:     0,
	}
	// with ignore filters on true and with bad filter
	err = aA.V1GetAttributeForEvent(context.Background(), ev2, rply2)
	if err != nil {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", nil, err)
	}
	if !reflect.DeepEqual(expected2, rply2) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(expected2), utils.ToJSON(rply2))
	}
}

func TestAttributeServicesProcessEventGetStringSliceOptsError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, cfg)
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProfileIDs: time.Second,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	_, err := aA.processEvent(context.Background(), args2.Tenant, args2, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != "cannot convert field: 1s to []string" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cannot convert field: 1s to []string", err)
	}
}

func TestAttributeServicesProcessEventGetBoolOptsError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, cfg)
	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		APIOpts: map[string]interface{}{
			utils.MetaProfileIgnoreFilters: time.Second,
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}
	_, err := aA.processEvent(context.Background(), args2.Tenant, args2, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0)
	if err == nil || err.Error() != "cannot convert field: 1s to bool" {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", "cannot convert field: 1s to bool", err)
	}
}

func TestAttributesParseAttributeMetaNone(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	if out, err := ParseAttribute(dp, utils.MetaNone, utils.EmptyString, config.NewRSRParsersMustCompile("~*req.cid;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep); err != nil {
		t.Fatal(err)
	} else if out != nil {
		t.Errorf("Expected %+v, Received %+v", nil, out)
	}
}

func TestAttributesParseAttributeMetaUsageDifferenceBadValError(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaUsageDifference, utils.EmptyString, config.NewRSRParsersMustCompile("", utils.InfieldSep), 0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)

	if err == nil || err.Error() != "invalid arguments <null> to *usageDifference" {
		t.Fatal(err)
	}
}

func TestAttributesParseAttributeMetaCCUsageError(t *testing.T) {
	dp := utils.MapStorage{
		utils.MetaReq: utils.MapStorage{
			"cid":  "12345",
			"to":   "1001",
			"from": "1002",
		},
	}
	_, err := ParseAttribute(dp, utils.MetaCCUsage, utils.EmptyString, config.NewRSRParsersMustCompile("::;~*req.to;~*req.from", utils.InfieldSep),
		0, utils.EmptyString, utils.EmptyString, utils.InfieldSep)
	if err == nil || err.Error() != "invalid requestNumber <::> to *ccUsage" {
		t.Fatal(err)
	}
}

func TestAttributesProcessEventSetError(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, cfg.DataDbCfg().Items)
	dm := NewDataManager(data, cfg.CacheCfg(), nil)
	filterS := NewFilterS(cfg, nil, dm)
	aA := NewAttributeService(dm, filterS, cfg)
	acPrf := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AC1",
		FilterIDs: []string{"*string:~*req.Attribute:testAttrValue"},
		Attributes: []*Attribute{
			{
				Path: "",
			},
		},
	}
	if err := dm.SetAttributeProfile(context.Background(), acPrf, true); err != nil {
		t.Error(err)
	}

	args2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "AcProcessEvent",
		Event: map[string]interface{}{
			"Attribute": "testAttrValue",
		},
		APIOpts: map[string]interface{}{
			utils.OptsAttributesProfileIDs: []string{"AC1"},
		},
	}
	eNM := utils.MapStorage{
		utils.MetaReq:  args2.Event,
		utils.MetaOpts: args2.APIOpts,
		utils.MetaVars: utils.MapStorage{
			utils.OptsAttributesProcessRuns: 0,
		},
	}

	if _, err := aA.processEvent(context.Background(), args2.Tenant, args2, eNM, newDynamicDP(context.TODO(), nil, nil, nil, "cgrates.org", eNM), utils.EmptyString, make(map[string]int), 0); err != nil {
		t.Error(err)
	}
}
