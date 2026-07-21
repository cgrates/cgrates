// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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

func TestNewAPIAttributeProfile(t *testing.T) {
	tests := []struct {
		name string
		attr *AttributeProfile
		want *APIAttributeProfile
	}{
		{
			attr: &AttributeProfile{
				Tenant:    CGRateSorg,
				ID:        "attrTestID",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Attributes: []*Attribute{
					{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: nil,
					},
					{
						Path:  "*tenant",
						Type:  MetaConstant,
						Value: nil,
					},
				},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: DynamicBlockers{},
			},
			want: &APIAttributeProfile{
				Tenant:    CGRateSorg,
				ID:        "attrTestID",
				FilterIDs: []string{"*string:~*req.Account:1002"},
				Attributes: []*ExternalAttribute{
					{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: "",
					},
					{
						Path:  "*tenant",
						Type:  MetaConstant,
						Value: "",
					},
				},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: DynamicBlockers{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAPIAttributeProfile(tt.attr)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected %+v, received %+v", tt.want, got)
			}
		})
	}
}

func TestAttributeProfileAsMapStringInterface(t *testing.T) {
	tests := []struct {
		name string
		ap   *AttributeProfile
		want map[string]any
	}{
		{
			ap: &AttributeProfile{
				Tenant:    CGRateSorg,
				ID:        "attrTestID",
				FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
				Attributes: []*Attribute{
					{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: nil,
					},
					{
						Path:  "*tenant",
						Type:  MetaConstant,
						Value: nil,
					},
				},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: DynamicBlockers{},
			},
			want: map[string]any{
				Tenant:    CGRateSorg,
				ID:        "attrTestID",
				FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
				Attributes: []*Attribute{
					{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: nil,
					},
					{
						Path:  "*tenant",
						Type:  MetaConstant,
						Value: nil,
					},
				},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: DynamicBlockers{},
			},
		},
		{
			name: "Nil AttributeProfile",
			ap:   nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ap.AsMapStringInterface()
			if !reflect.DeepEqual(ToJSON(got), ToJSON(tt.want)) {
				t.Errorf("Expected %+v, received %+v", tt.want, got)
			}
		})
	}
}

func TestMapStringInterfaceToAttributeProfile(t *testing.T) {
	tests := []struct {
		name    string
		m       map[string]any
		want    *AttributeProfile
		wantErr string
	}{
		{
			m: map[string]any{
				Tenant:    CGRateSorg,
				ID:        "attrTestID",
				FilterIDs: []string{"*exists:~*opts.*usage:"},
				Attributes: []any{
					map[string]any{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: nil,
					},
					map[string]any{
						Path:  "*tenant",
						Type:  MetaConstant,
						Value: nil,
					},
				},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: DynamicBlockers{},
			},
			want: &AttributeProfile{
				Tenant:    CGRateSorg,
				ID:        "attrTestID",
				FilterIDs: []string{"*exists:~*opts.*usage:"},
				Attributes: []*Attribute{
					{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: nil,
					},
					{
						Path:  "*tenant",
						Type:  MetaConstant,
						Value: nil,
					},
				},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: DynamicBlockers{},
			},
		},
		{
			name: "Attributes as any",
			m: map[string]any{
				Attributes: []any{
					map[string]string{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: "",
					},
				},
			},
			want: &AttributeProfile{
				Attributes: []*Attribute{},
			},
		},
		{
			name: "Attributes as *Attribute",
			m: map[string]any{
				Attributes: []*Attribute{
					{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: nil,
					},
				},
			},
			want: &AttributeProfile{
				Attributes: nil,
			},
		},
		{
			name: "Nil Attributes",
			m: map[string]any{
				Attributes: nil,
			},
			want: &AttributeProfile{
				Attributes: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapStringInterfaceToAttributeProfile(tt.m)
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("Expected %+v, received %+v", tt.wantErr, err)
			}

			if !reflect.DeepEqual(ToJSON(got), ToJSON(tt.want)) {
				t.Errorf("Expected %+v, received %+v", tt.want, got)
			}
		})
	}
}
func TestAttributeProfileClone(t *testing.T) {
	tests := []struct {
		name string
		ap   *AttributeProfile
	}{
		{
			name: "Complete AttributeProfile",
			ap: &AttributeProfile{
				Tenant:    CGRateSorg,
				ID:        "attrTestID",
				FilterIDs: []string{"*string:~*req.Account:1003"},
				Attributes: []*Attribute{
					{
						Path:  AccountField,
						Type:  MetaConstant,
						Value: nil,
					},
					{
						Path:  "*tenant",
						Type:  MetaConstant,
						Value: nil,
					},
				},
				Weights: DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: DynamicBlockers{},
			},
		},
		{
			name: "nil case",
			ap:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := tt.ap.Clone()
			if !reflect.DeepEqual(rcv, tt.ap) {
				t.Errorf("Expected %+v, received %+v", tt.ap, rcv)
			}

			if tt.ap != nil && rcv == tt.ap {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			rcv := tt.ap.CacheClone()
			if !reflect.DeepEqual(rcv, tt.ap) {
				t.Errorf("Expected %+v, received %+v", tt.ap, rcv)
			}
		})
	}
}

func TestAttributeClone(t *testing.T) {
	tests := []struct {
		name string
		attr *Attribute
	}{
		{
			name: "Complete Attribute",
			attr: &Attribute{
				FilterIDs: []string{"*string:~*req.PassField:Test"},
				Path:      "*req.Password",
				Type:      MetaPassword,
				Value:     NewRSRParsersMustCompile("abcd123", RSRSep),
				Blockers:  DynamicBlockers{},
			},
		},
		{
			name: "nil case",
			attr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := tt.attr.Clone()
			if !reflect.DeepEqual(rcv, tt.attr) {
				t.Errorf("Expected %+v, received %+v", tt.attr, rcv)
			}

			if tt.attr != nil && rcv == tt.attr {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}
