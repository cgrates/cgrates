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
	atrPs AttributeProfiles
	srv   AttributeService
	dmAtr *DataManager
)

var sTestsAttributes = []func(t *testing.T){
	testPopulateAttrService,
	testAttributeMatchingAttributeProfilesForEvent,
	testAttributeProfileForEvent,
	testAttributeProcessEvent,
}

func TestAttributes(t *testing.T) {
	for _, stest := range sTestsAttributes {
		t.Run("Test Suppliers", stest)
	}
}

func testPopulateAttrService(t *testing.T) {
	var filters1 []*RequestFilter
	var filters2 []*RequestFilter
	second := 1 * time.Second
	data, _ := NewMapStorage()
	dmAtr = NewDataManager(data)
	context := utils.MetaRating
	attrMap := make(map[string]map[string]*Attribute)
	attrMap["FL1"] = make(map[string]*Attribute)
	attrMap["FL1"]["In1"] = &Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: "Al1",
		Append:     true,
	}
	atrPs = AttributeProfiles{
		&AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "attributeprofile1",
			Context:   context,
			FilterIDs: []string{"filter1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Now().Add(time.Duration(20 * time.Minute)).Local(),
			},
			Attributes: attrMap,
			Weight:     20,
		},
		&AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "attributeprofile2",
			Context:   context,
			FilterIDs: []string{"filter2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
				ExpiryTime:     time.Now().Add(time.Duration(20 * time.Minute)).Local(),
			},
			Attributes: attrMap,
			Weight:     20,
		},
	}
	x, err := NewRequestFilter(MetaString, "attributeprofile1", []string{"Attribute"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewRequestFilter(MetaGreaterOrEqual, "UsageInterval", []string{second.String()})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)
	x, err = NewRequestFilter(MetaGreaterOrEqual, "Weight", []string{"9.0"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters1 = append(filters1, x)

	x, err = NewRequestFilter(MetaString, "attributeprofile2", []string{"Attribute"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	filter1 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter1", RequestFilters: filters1}
	filter2 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter2", RequestFilters: filters2}
	dmAtr.SetFilter(filter1)
	dmAtr.SetFilter(filter2)
	srv = AttributeService{
		dm:            dmAtr,
		filterS:       &FilterS{dm: dmAtr},
		indexedFields: []string{"attributeprofile1", "attributeprofile2"},
	}
	ev := make(map[string]interface{})
	ev["attributeprofile1"] = "Attribute"
	ev["attributeprofile2"] = "Attribute"
	ev[utils.AnswerTime] = time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC).Local()
	ev["UsageInterval"] = "1s"
	ev["PddInterval"] = "1s"
	ev["Weight"] = "20.0"
	sev = &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "attribute_event",
		Context: &context,
		Event:   ev,
	}

	for _, atr := range atrPs {
		dmAtr.SetAttributeProfile(atr)
	}
	prefix := utils.ConcatenatedKey(sev.Tenant, *sev.Context)
	ref := NewReqFilterIndexer(dmAtr, utils.AttributeProfilePrefix, prefix)
	ref.IndexTPFilter(FilterToTPFilter(filter1), "attributeprofile1")
	ref.IndexTPFilter(FilterToTPFilter(filter2), "attributeprofile2")
	err = ref.StoreIndexes()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func testAttributeMatchingAttributeProfilesForEvent(t *testing.T) {
	atrpl, err := srv.matchingAttributeProfilesForEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrpl[0]) && !reflect.DeepEqual(atrPs[0], atrpl[1]) {
		t.Errorf("Expecting: %+v, received: %+v", atrPs[0], atrpl[0])
	} else if !reflect.DeepEqual(atrPs[1], atrpl[1]) && !reflect.DeepEqual(atrPs[1], atrpl[0]) {
		t.Errorf("Expecting: %+v, received: %+v", atrPs[1], atrpl[1])
	}
}

func testAttributeProfileForEvent(t *testing.T) {
	context := utils.MetaRating
	ev := make(map[string]interface{})
	ev["attributeprofile1"] = "Attribute"
	ev["UsageInterval"] = "1s"
	ev["Weight"] = "9.0"
	sev = &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "attribute_event",
		Context: &context,
		Event:   ev,
	}
	atrpl, err := srv.attributeProfileForEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrpl) && !reflect.DeepEqual(atrPs[1], atrpl) {
		t.Errorf("Expecting: %+v, received: %+v", atrPs[0], atrpl)
	}
}

func testAttributeProcessEvent(t *testing.T) {
	context := utils.MetaRating
	ev := make(map[string]interface{})
	ev["attributeprofile1"] = "Attribute"
	ev["UsageInterval"] = "1s"
	ev["Weight"] = "9.0"
	sev = &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "attribute_event",
		Context: &context,
		Event:   ev,
	}
	eRply := &AttrSProcessEventReply{
		MatchedProfile: "attributeprofile1",
		CGREvent:       sev,
	}
	atrpl, err := srv.processEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfile, atrpl.MatchedProfile) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.MatchedProfile, atrpl.MatchedProfile)
	} else if !reflect.DeepEqual(eRply.AlteredFields, atrpl.AlteredFields) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.AlteredFields, atrpl.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent, atrpl.CGREvent) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.CGREvent, atrpl.CGREvent)
	}
}
