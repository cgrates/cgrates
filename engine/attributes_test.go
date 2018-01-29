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
	srv                    AttributeService
	dmAtr                  *DataManager

	context = utils.MetaRating

	mapSubstitutes = map[string]map[interface{}]*Attribute{
		"FL1": map[interface{}]*Attribute{
			"In1": &Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
			},
		},
	}

	sev = &utils.CGREvent{
		Tenant:  config.CgrConfig().DefaultTenant,
		ID:      "attribute_event",
		Context: &context,
		Event: map[string]interface{}{
			"Attribute":      "AttributeProfile1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 30, 0, 0, time.UTC),
			"UsageInterval":  "1s",
			"PddInterval":    "1s",
			"Weight":         "20.0",
		},
	}
	sev2 = &utils.CGREvent{
		Tenant:  config.CgrConfig().DefaultTenant,
		ID:      "attribute_event",
		Context: &context,
		Event: map[string]interface{}{
			"Attribute": "AttributeProfile2",
		},
	}
	sev3 = &utils.CGREvent{
		Tenant:  config.CgrConfig().DefaultTenant,
		ID:      "attribute_event",
		Context: &context,
		Event: map[string]interface{}{
			"Attribute": "AttributeProfilePrefix",
		},
	}
	sev4 = &utils.CGREvent{
		Tenant:  config.CgrConfig().DefaultTenant,
		ID:      "attribute_event",
		Context: &context,
		Event: map[string]interface{}{
			"Weight": "200.0",
		},
	}

	atrPs = AttributeProfiles{
		&AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "attributeprofile1",
			Contexts:  []string{context},
			FilterIDs: []string{"filter1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				&Attribute{
					FieldName:  "FL1",
					Initial:    "In1",
					Substitute: "Al1",
					Append:     true,
				},
			},
			Weight:     20,
			attributes: mapSubstitutes,
		},
		&AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "attributeprofile2",
			Contexts:  []string{context},
			FilterIDs: []string{"filter2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				&Attribute{
					FieldName:  "FL1",
					Initial:    "In1",
					Substitute: "Al1",
					Append:     true,
				},
			},
			Weight:     20,
			attributes: mapSubstitutes,
		},
		&AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "attributeprofile3",
			Contexts:  []string{context},
			FilterIDs: []string{"preffilter1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				&Attribute{
					FieldName:  "FL1",
					Initial:    "In1",
					Substitute: "Al1",
					Append:     true,
				},
			},
			attributes: mapSubstitutes,
			Weight:     20,
		},
		&AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "attributeprofile4",
			Contexts:  []string{context},
			FilterIDs: []string{"defaultf1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     cloneExpTimeAttributes,
			},
			Attributes: []*Attribute{
				&Attribute{
					FieldName:  "FL1",
					Initial:    "In1",
					Substitute: "Al1",
					Append:     true,
				},
			},
			attributes: mapSubstitutes,
			Weight:     20,
		},
	}
)

func TestAttributeCache(t *testing.T) {
	//Need clone because time.Now adds extra information that DeepEqual doesn't like
	if err := utils.Clone(expTimeAttributes, &cloneExpTimeAttributes); err != nil {
		t.Error(err)
	}
	data, _ := NewMapStorage()
	dmAtr = NewDataManager(data)
	for _, atr := range atrPs {
		if err = dmAtr.SetAttributeProfile(atr, false); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//Test each attribute from cache
	for _, atr := range atrPs {
		if tempAttr, err := dmAtr.GetAttributeProfile(atr.Tenant, atr.ID, false, utils.NonTransactional); err != nil {
			t.Errorf("Error: %+v", err)
		} else if !reflect.DeepEqual(atr, tempAttr) {
			t.Errorf("Expecting: %+v, received: %+v", atr, tempAttr)
		}
	}
}

func TestAttributePopulateAttrService(t *testing.T) {
	var filters1 []*RequestFilter
	var filters2 []*RequestFilter
	var preffilter []*RequestFilter
	var defaultf []*RequestFilter
	second := 1 * time.Second
	//refresh the DM
	data, _ := NewMapStorage()
	dmAtr = NewDataManager(data)
	srv = AttributeService{
		dm:      dmAtr,
		filterS: &FilterS{dm: dmAtr},
	}
	ref := NewReqFilterIndexer(dmAtr, utils.AttributeProfilePrefix,
		utils.ConcatenatedKey(config.CgrConfig().DefaultTenant, utils.MetaRating))
	for _, atr := range atrPs {
		if err = dmAtr.SetAttributeProfile(atr, false); err != nil {
			t.Errorf("Error: %+v", err)
		}
	}
	//filter1
	x, err := NewRequestFilter(MetaString, "Attribute", []string{"AttributeProfile1"})
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
	filter1 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter1", RequestFilters: filters1}
	dmAtr.SetFilter(filter1)
	ref.IndexTPFilter(FilterToTPFilter(filter1), "attributeprofile1")

	//filter2
	x, err = NewRequestFilter(MetaString, "Attribute", []string{"AttributeProfile2"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	filters2 = append(filters2, x)
	filter2 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "filter2", RequestFilters: filters2}
	dmAtr.SetFilter(filter2)
	ref.IndexTPFilter(FilterToTPFilter(filter2), "attributeprofile2")

	//prefix filter
	x, err = NewRequestFilter(MetaPrefix, "Attribute", []string{"AttributeProfilePrefix"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	preffilter = append(preffilter, x)
	preffilter1 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "preffilter1", RequestFilters: preffilter}
	dmAtr.SetFilter(preffilter1)
	ref.IndexTPFilter(FilterToTPFilter(preffilter1), "attributeprofile3")

	//default filter
	x, err = NewRequestFilter(MetaGreaterOrEqual, "Weight", []string{"200.00"})
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	defaultf = append(defaultf, x)
	defaultf1 := &Filter{Tenant: config.CgrConfig().DefaultTenant, ID: "defaultf1", RequestFilters: defaultf}
	dmAtr.SetFilter(defaultf1)
	ref.IndexTPFilter(FilterToTPFilter(defaultf1), "attributeprofile4")
	err = ref.StoreIndexes()
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
}

func TestAttributeMatchingAttributeProfilesForEvent(t *testing.T) {
	atrp, err := srv.matchingAttributeProfilesForEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrp[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", atrPs[0], atrp[0])
	}
	atrp, err = srv.matchingAttributeProfilesForEvent(sev2)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[1], atrp[0]) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs), utils.ToJSON(atrp))
	}
	atrp, err = srv.matchingAttributeProfilesForEvent(sev3)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[2], atrp[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", atrPs[2], atrp[0])
	}
	atrp, err = srv.matchingAttributeProfilesForEvent(sev4)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[3], atrp[0]) {
		t.Errorf("Expecting: %+v, received: %+v ", atrPs[3], atrp[0])
	}
}

func TestAttributeProfileForEvent(t *testing.T) {
	atrp, err := srv.attributeProfileForEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[0], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[0]), utils.ToJSON(atrp))
	}

	atrp, err = srv.attributeProfileForEvent(sev2)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[1], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[1]), utils.ToJSON(atrp))

	}
	atrp, err = srv.attributeProfileForEvent(sev3)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[2], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[2]), utils.ToJSON(atrp))
	}
	atrp, err = srv.attributeProfileForEvent(sev4)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(atrPs[3], atrp) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(atrPs[3]), utils.ToJSON(atrp))
	}
}

func TestAttributeProcessEvent(t *testing.T) {

	eRply := &AttrSProcessEventReply{
		MatchedProfile: "attributeprofile1",
		CGREvent:       sev,
	}
	atrp, err := srv.processEvent(sev)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfile, atrp.MatchedProfile) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.MatchedProfile, atrp.MatchedProfile)
	} else if !reflect.DeepEqual(eRply.AlteredFields, atrp.AlteredFields) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.AlteredFields, atrp.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent, atrp.CGREvent) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.CGREvent, atrp.CGREvent)
	}
	eRply = &AttrSProcessEventReply{
		MatchedProfile: "attributeprofile2",
		CGREvent:       sev2,
	}
	atrp, err = srv.processEvent(sev2)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfile, atrp.MatchedProfile) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.MatchedProfile, atrp.MatchedProfile)
	} else if !reflect.DeepEqual(eRply.AlteredFields, atrp.AlteredFields) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.AlteredFields, atrp.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent, atrp.CGREvent) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.CGREvent, atrp.CGREvent)
	}
	eRply = &AttrSProcessEventReply{
		MatchedProfile: "attributeprofile3",
		CGREvent:       sev3,
	}
	atrp, err = srv.processEvent(sev3)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfile, atrp.MatchedProfile) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.MatchedProfile, atrp.MatchedProfile)
	} else if !reflect.DeepEqual(eRply.AlteredFields, atrp.AlteredFields) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.AlteredFields, atrp.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent, atrp.CGREvent) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.CGREvent, atrp.CGREvent)
	}
	eRply = &AttrSProcessEventReply{
		MatchedProfile: "attributeprofile4",
		CGREvent:       sev4,
	}
	atrp, err = srv.processEvent(sev4)
	if err != nil {
		t.Errorf("Error: %+v", err)
	}
	if !reflect.DeepEqual(eRply.MatchedProfile, atrp.MatchedProfile) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.MatchedProfile, atrp.MatchedProfile)
	} else if !reflect.DeepEqual(eRply.AlteredFields, atrp.AlteredFields) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.AlteredFields, atrp.AlteredFields)
	} else if !reflect.DeepEqual(eRply.CGREvent, atrp.CGREvent) {
		t.Errorf("Expecting: %+v, received: %+v", eRply.CGREvent, atrp.CGREvent)
	}
}
