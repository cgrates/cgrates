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

package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestConvertExternalToProfile(t *testing.T) {
	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*string:~*opts.*context:*sessions|*cdrs"},
		Attributes: []*ExternalAttribute{
			{
				Path:  MetaReq + NestingSep + "Account",
				Value: "1001",
			},
		},
		Weights: DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	expAttr := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*string:~*opts.*context:*sessions|*cdrs"},
		Attributes: []*Attribute{
			{
				Path:  MetaReq + NestingSep + "Account",
				Value: NewRSRParsersMustCompile("1001", InfieldSep),
			},
		},
		Weights: make(DynamicWeights, 1),
	}
	expAttr.Weights[0] = &DynamicWeight{
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
		Tenant:     "cgrates.org",
		ID:         "ATTR_ID",
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "*string:~*opts.*context:*sessions|*cdrs"},
		Attributes: []*ExternalAttribute{},
		Weights: DynamicWeights{
			{
				Weight: 20,
			},
		},
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
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "*string:~*opts.*context:*sessions|*cdrs"},
		Attributes: []*ExternalAttribute{
			{
				Path: MetaReq + NestingSep + "Account",
			},
		},
		Weights: DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	_, err := external.AsAttributeProfile()
	if err == nil || err.Error() != "MANDATORY_IE_MISSING: [Value]" {
		t.Error(err)
	}

}

func TestNewAttributeFromInline(t *testing.T) {
	attrID := "*sum:*req.Field2:10&~*req.NumField&20;*sum:*req.Field3:10&~*req.NumField4&20"
	expAttrPrf1 := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     attrID,
		Attributes: []*Attribute{
			{
				Path:  MetaReq + NestingSep + "Field2",
				Type:  MetaSum,
				Value: NewRSRParsersMustCompile("10;~*req.NumField;20", InfieldSep),
			},
			{
				Path:  MetaReq + NestingSep + "Field3",
				Type:  MetaSum,
				Value: NewRSRParsersMustCompile("10;~*req.NumField4;20", InfieldSep),
			},
		},
	}
	attr, err := NewAttributeFromInline("cgrates.org", attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", ToJSON(expAttrPrf1), ToJSON(attr))
	}
}

func TestNewAttributeFromInlineWithMultipleRuns(t *testing.T) {
	attrID := "*constant:*req.RequestType:*rated;*constant:*req.Category:call"
	expAttrPrf1 := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     attrID,
		Attributes: []*Attribute{
			{
				Path:  MetaReq + NestingSep + "RequestType",
				Type:  MetaConstant,
				Value: NewRSRParsersMustCompile("*rated", InfieldSep),
			},
			{
				Path:  MetaReq + NestingSep + "Category",
				Type:  MetaConstant,
				Value: NewRSRParsersMustCompile("call", InfieldSep),
			},
		},
	}
	attr, err := NewAttributeFromInline("cgrates.org", attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", ToJSON(expAttrPrf1), ToJSON(attr))
	}
}
func TestNewAttributeFromInlineWithMultipleRuns2(t *testing.T) {
	attrID := "*constant:*req.RequestType*rated;*constant:*req.Category:call"

	expErr := fmt.Sprintf("inline parse error for string: <%s>", "*constant:*req.RequestType*rated")
	if _, err := NewAttributeFromInline("cgrates.org", attrID); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %s received %v", expErr, err)
	}

	attrID = "*constant:*req.RequestType:`*rated;*constant:*req.Category:call"

	if _, err := NewAttributeFromInline("cgrates.org", attrID); err == nil {
		t.Error(err)
	}
}

func TestNewAttributeFromInlineWithMultipleVaslues(t *testing.T) {
	attrID := "*variable:*req.Category:call_&*req.OriginID;*constant:*req.RequestType:*rated"
	expAttrPrf1 := &AttributeProfile{
		Tenant: "cgrates.org",
		ID:     attrID,
		Attributes: []*Attribute{
			{
				Path:  MetaReq + NestingSep + "Category",
				Type:  MetaVariable,
				Value: NewRSRParsersMustCompile("call_;*req.OriginID", InfieldSep),
			},
			{
				Path:  MetaReq + NestingSep + "RequestType",
				Type:  MetaConstant,
				Value: NewRSRParsersMustCompile("*rated", InfieldSep),
			},
		},
	}
	attr, err := NewAttributeFromInline("cgrates.org", attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", ToJSON(expAttrPrf1), ToJSON(attr))
	}
}

func TestLibAttributesTenantIDInLine(t *testing.T) {
	ap := &AttributeProfile{
		Tenant:  "cgrates.org",
		ID:      "AttrPrf",
		Weights: make(DynamicWeights, 1),
	}
	ap.Weights[0] = &DynamicWeight{
		Weight: 0,
	}
	exp := "cgrates.org:AttrPrf"
	if rcv := ap.TenantIDInline(); rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestLibAttributesTenantIDMetaPrefix(t *testing.T) {
	ap := &AttributeProfile{
		Tenant:  "cgrates.org",
		ID:      "*default",
		Weights: make(DynamicWeights, 1),
	}
	ap.Weights[0] = &DynamicWeight{
		FilterIDs: []string{""},
		Weight:    0,
	}

	exp := "*default"
	if rcv := ap.TenantIDInline(); rcv != exp {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, rcv)
	}
}

func TestAttributeProfileSet(t *testing.T) {
	dp := AttributeProfile{}
	exp := AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Blockers: DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      MetaConstant,
			Value:     NewRSRParsersMustCompile("10", InfieldSep),
			FilterIDs: []string{"fltr1"},
			Blockers: DynamicBlockers{
				{
					Blocker: true,
				},
			},
		}},
		Weights: make(DynamicWeights, 1),
	}
	exp.Weights[0] = &DynamicWeight{
		Weight: 10,
	}
	if err := dp.Set([]string{}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField", "1"}, "", false); err != ErrWrongPath {
		t.Error(err)
	}

	if err := dp.Set([]string{Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Attributes, Path}, "*req.Account", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Attributes, Type}, MetaConstant, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Attributes, Value}, "10", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Attributes, FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Attributes, Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{Attributes, "Wrong"}, true, false); err != ErrWrongPath {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(dp))
	}
}

func TestAttributeProfileAsInterface(t *testing.T) {
	ap := AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   make(DynamicWeights, 1),
		Blockers: DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      MetaConstant,
			Value:     NewRSRParsersMustCompile("10", InfieldSep),
			FilterIDs: []string{"fltr1"},
		}},
	}
	ap.Weights[0] = &DynamicWeight{
		Weight: 10,
	}
	if _, err := ap.FieldAsInterface(nil); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Attributes}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Attributes + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Attributes + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if _, err := ap.FieldAsInterface([]string{Attributes + "[4]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Attributes + "[0]", ""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{Attributes + "0]"}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{Attributes + "[0]", FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Attributes + "[0]", FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].FilterIDs[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Attributes + "[0]", Path}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].Path; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Attributes + "[0]", Type}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].Type; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{Attributes + "[0]", Value}); err != nil {
		t.Fatal(err)
	} else if exp := "10"; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := ap.FieldAsString([]string{""}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsString([]string{ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := ap.String(), ToJSON(ap); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}

	if _, err := ap.Attributes[0].FieldAsString([]string{}); err != ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.Attributes[0].FieldAsString([]string{Value}); err != nil {
		t.Fatal(err)
	} else if exp := "10"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
	if val, exp := ap.Attributes[0].String(), ToJSON(ap.Attributes[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(val))
	}
}

func TestAttributeProfileMerge(t *testing.T) {
	dp := &AttributeProfile{}
	exp := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   make(DynamicWeights, 1),
		Blockers: DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      MetaConstant,
			Value:     NewRSRParsersMustCompile("10", InfieldSep),
			FilterIDs: []string{"fltr1"},
		}},
	}
	exp.Weights[0] = &DynamicWeight{
		Weight: 10,
	}
	dp.Merge(&AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   make(DynamicWeights, 1),
		Blockers: DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      MetaConstant,
			Value:     NewRSRParsersMustCompile("10", InfieldSep),
			FilterIDs: []string{"fltr1"},
		}},
	})
	dp.Weights[0] = &DynamicWeight{
		Weight: 10,
	}
	if !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(exp), ToJSON(dp))
	}
}

func TestAttributeProfilCompileSubstitutes(t *testing.T) {

	ap := &AttributeProfile{
		Attributes: []*Attribute{
			{Value: RSRParsers{&RSRParser{
				Rules: "~*req.Account{*unuportedConverter}",
			}}},
		},
	}
	expErr := "invalid converter value in string: <*unuportedConverter>, err: unsupported converter definition: <*unuportedConverter>"
	if err := ap.compileSubstitutes(); err == nil || err.Error() != expErr {
		t.Errorf("Expected error <%v> \n but received error \n <%v>", expErr, err)
	}

}

func TestAttributeFieldAsInterface(t *testing.T) {
	at := &Attribute{
		Path:      "*req.Account",
		Type:      MetaConstant,
		Value:     NewRSRParsersMustCompile("10", InfieldSep),
		FilterIDs: []string{"fltr1"},
		Blockers: DynamicBlockers{
			{
				Blocker: true,
			},
		},
	}

	if rcv, err := at.FieldAsInterface([]string{Blockers}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, at.Blockers) {
		t.Errorf("Expected %v \n but received \n %v", ToJSON(at), ToJSON(rcv))
	}
}

func TestAPIAPAsAttributeProfileNilPathErr(t *testing.T) {

	ext := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "*string:~*opts.*context:*sessions|*cdrs"},
		Attributes: []*ExternalAttribute{
			{
				Value: "1001",
			},
		},
		Weights: DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	expErr := "MANDATORY_IE_MISSING: [Path]"
	if _, err := ext.AsAttributeProfile(); err == nil || err.Error() != expErr {
		t.Errorf("Expecting error <%v>, Reveived error <%v>", expErr, err)
	}

}

func TestAPIAPAsAttributeProfileParseErr(t *testing.T) {

	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*string:~*opts.*context:*sessions|*cdrs"},
		Attributes: []*ExternalAttribute{
			{
				Path:  MetaReq + NestingSep + "Account",
				Value: "a{*",
			},
		},
		Weights: DynamicWeights{
			{
				Weight: 20,
			},
		},
	}

	expErr := "invalid converter terminator in rule: <a{*>"
	if _, err := external.AsAttributeProfile(); err == nil || err.Error() != expErr {
		t.Errorf("Expecting error <%v>, Reveived error <%v>", expErr, err)
	}

}

func TestNewAttributeFromInlineNilPathErr(t *testing.T) {
	attrID := "*variable:*req.Category:call_&*req.OriginID;*constant::"

	expErr := "empty path in inline AttributeProfile <*variable:*req.Category:call_&*req.OriginID;*constant::>"
	_, err := NewAttributeFromInline("cgrates.org", attrID)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expecting error <%v>, Reveived error <%v>", expErr, err)
	}
}
