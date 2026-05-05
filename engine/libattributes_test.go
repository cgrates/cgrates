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

package engine

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
)

func TestConvertExternalToProfile(t *testing.T) {
	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*ExternalAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: "1001",
			},
		},
		Weight: 20,
	}

	expAttr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weight: 20,
	}

	rcv, err := external.AsAttributeProfile()
	if err != nil {
		t.Error(err)
	}
	rcv.Compile()

	if !reflect.DeepEqual(expAttr, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", expAttr, rcv)
	}
}

func TestConvertExternalToProfileMissing(t *testing.T) {
	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*ExternalAttribute{},
		Weight:     20,
	}

	_, err := external.AsAttributeProfile()
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [Attributes]" {
		t.Error(err)
	}

}

func TestConvertExternalToProfileMissing2(t *testing.T) {
	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*ExternalAttribute{
			{
				Path: utils.MetaReq + utils.NestingSep + "Account",
			},
		},
		Weight: 20,
	}

	_, err := external.AsAttributeProfile()
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [Value]" {
		t.Error(err)
	}

}

func TestConvertExternalToProfileInvalidValue(t *testing.T) {
	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"FLTR_ACNT", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2026, 4, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2026, 4, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*ExternalAttribute{
			{
				Path: utils.EmptyString,
			},
		},
		Weight: 20,
	}
	expectedErr := "MANDATORY_IE_MISSING: [Path]"
	_, err := external.AsAttributeProfile()
	if err == nil || err.Error() != expectedErr {
		t.Error(err)
	}
}
func TestConvertExternalToProfileEmptyPath(t *testing.T) {
	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"FLTR_ACNT", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2026, 4, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2026, 4, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*ExternalAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: "`error",
			},
		},
		Weight: 20,
	}
	expectedErr := "Unclosed unspilit syntax"
	_, err := external.AsAttributeProfile()
	if err == nil || err.Error() != expectedErr {
		t.Errorf("Expected: %#v, recieved: %#v", expectedErr, err)
	}
}
func TestNewAttributeFromInline(t *testing.T) {
	attrID := "*sum:*req.Field2:10&~*req.NumField&20;*sum:*req.Field3:10&~*req.NumField4&20"
	expAttrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       attrID,
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field2",
				Type:  utils.MetaSum,
				Value: config.NewRSRParsersMustCompile("10;~*req.NumField;20", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Field3",
				Type:  utils.MetaSum,
				Value: config.NewRSRParsersMustCompile("10;~*req.NumField4;20", utils.InfieldSep),
			},
		},
	}
	attr, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}

func TestNewAttributeFromInlineWithMultipleRuns(t *testing.T) {
	attrID := "*constant:*req.RequestType:*rated;*constant:*req.Category:call"
	expAttrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       attrID,
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "RequestType",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("*rated", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "Category",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("call", utils.InfieldSep),
			},
		},
	}
	attr, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}
func TestNewAttributeFromInlineWithMultipleRuns2(t *testing.T) {
	attrID := "*constant:*req.RequestType*rated;*constant:*req.Category:call"

	expErr := fmt.Sprintf("inline parse error for string: <%s>", "*constant:*req.RequestType*rated")
	if _, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s received %v", expErr, err)
	}

	attrID = "*constant:*req.RequestType:`*rated;*constant:*req.Category:call"

	if _, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID); err == nil {
		t.Error(err)
	}
}

func TestNewAttributeFromInlineWithMultipleVaslues(t *testing.T) {
	attrID := "*variable:*req.Category:call_&*req.OriginID;*constant:*req.RequestType:*rated"
	expAttrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       attrID,
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Category",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("call_;*req.OriginID", utils.InfieldSep),
			},
			{
				Path:  utils.MetaReq + utils.NestingSep + "RequestType",
				Type:  utils.MetaConstant,
				Value: config.NewRSRParsersMustCompile("*rated", utils.InfieldSep),
			},
		},
	}
	attr, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}

func TestNewAttributeFromInlineWithEmptyRule(t *testing.T) {
	attrID := "*variable:*req.Category:call_&*req.OriginID;*constant::*rated"
	expAttrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       attrID,
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Category",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("call_;*req.OriginID", utils.InfieldSep),
			},
		},
	}
	expErr := "empty path in inline AttributeProfile <*variable:*req.Category:call_&*req.OriginID;*constant::*rated>"
	attr, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err != nil && err.Error() != expErr {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}

func TestLibAttributesTenantIDInLine(t *testing.T) {
	ap := &AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "AttrPrf",
		Contexts: []string{utils.MetaAny},
		Weight:   10,
	}

	exp := "cgrates.org:AttrPrf"
	if rcv := ap.TenantIDInline(); rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLibAttributesTenantIDMetaPrefix(t *testing.T) {
	ap := &AttributeProfile{
		Tenant:   "cgrates.org",
		ID:       "*default",
		Contexts: []string{utils.MetaAny},
		Weight:   10,
	}

	exp := "*default"
	if rcv := ap.TenantIDInline(); rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestEngineAttributeProfilesSort(t *testing.T) {

	unsorted := AttributeProfiles{
		{Weight: 10},
		{Weight: 2},
		{Weight: 15},
	}
	expected := AttributeProfiles{
		{Weight: 15},
		{Weight: 10},
		{Weight: 2},
	}
	unsorted.Sort()
	if !cmp.Equal(unsorted, expected) {
		t.Errorf("Sort failed. Expected %v, got %v", expected, unsorted)
	}
}

func TestAttributeClone(t *testing.T) {
	tests := []struct {
		name      string
		attribute *Attribute
	}{
		{
			name: "Complete Attribute",
			attribute: &Attribute{
				FilterIDs: []string{"AttrFltr1", "AttrFltr2"},
				Path:      utils.MetaReq + utils.NestingSep + "Category",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("call_;*req.OriginID", utils.InfieldSep),
			},
		},
		{
			name: "Nil Fields",
			attribute: &Attribute{
				FilterIDs: nil,
				Path:      utils.MetaReq + utils.NestingSep + "Category",
				Type:      utils.MetaVariable,
				Value:     nil,
			},
		},
		{
			name:      "Nil case",
			attribute: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attribute.Clone()

			if !reflect.DeepEqual(result, tt.attribute) {
				t.Errorf("Clone() = %v, want %v", result, tt.attribute)
			}

			if result != nil && result == tt.attribute {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}

func TestAttributeProfileClone(t *testing.T) {
	tests := []struct {
		name        string
		attrProfile *AttributeProfile
	}{
		{
			name: "Complete AttributeProfile",
			attrProfile: &AttributeProfile{
				Tenant:    "cgrates.org",
				ID:        "ATTR_ID",
				Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
				FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
				ActivationInterval: &utils.ActivationInterval{
					ActivationTime: time.Date(2026, 4, 14, 14, 35, 0, 0, time.UTC),
					ExpiryTime:     time.Date(2026, 4, 14, 14, 35, 0, 0, time.UTC),
				},
				Attributes: []*Attribute{
					{
						Path:  utils.MetaReq + utils.NestingSep + "Account",
						Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
					},
				},
				Weight: 20,
			},
		},
		{
			name: "Nil fields",
			attrProfile: &AttributeProfile{
				Tenant:             "cgrates.org",
				ID:                 "ATTR_ID",
				Contexts:           nil,
				FilterIDs:          nil,
				ActivationInterval: nil,
				Attributes:         nil,
				Weight:             20,
			},
		},
		{
			name:        "Nil case",
			attrProfile: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.attrProfile.Clone()

			if !reflect.DeepEqual(result, tt.attrProfile) {
				t.Errorf("Clone() = %v, want %v", result, tt.attrProfile)
			}

			if result != nil && result == tt.attrProfile {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}

func TestAttributeProfileCompileError(t *testing.T) {
	attrProfile := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       "attrID",
		Contexts: []string{utils.MetaAny},
		Attributes: []*Attribute{
			{
				Path: utils.MetaReq + utils.NestingSep + "Field2",
				Type: utils.MetaSum,
				Value: config.RSRParsers{
					{Rules: "~*req.Field{*}"},
				},
			},
		},
	}
	expErr := errors.New("invalid converter value in string: <*>, err: unsupported converter definition: <*>")
	if err := attrProfile.Compile(); err != nil && err.Error() != expErr.Error() {
		t.Errorf("Expected: %#v, recieved: %#v", expErr, err)
	}
}

func TestExternalAttributeAPIError(t *testing.T) {
	dDP := utils.MapStorage{"key": "value"}
	_, err := externalAttributeAPI("invalid", dDP)
	expectedErr := "url is not specified"
	if err != nil && err.Error() != expectedErr {
		t.Errorf("externalAttributeAPI() failed: %v", err)
	}
}
