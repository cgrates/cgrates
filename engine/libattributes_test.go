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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestConvertExternalToProfile(t *testing.T) {
	external := &APIAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ATTR_ID",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE", "*string:~*opts.*context:*sessions|*cdrs"},
		Attributes: []*ExternalAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: "1001",
			},
		},
		Weights: utils.DynamicWeights{
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
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weights: make(utils.DynamicWeights, 1),
	}
	expAttr.Weights[0] = &utils.DynamicWeight{
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
		Weights: utils.DynamicWeights{
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
				Path: utils.MetaReq + utils.NestingSep + "Account",
			},
		},
		Weights: utils.DynamicWeights{
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
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     attrID,
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
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     attrID,
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
		Tenant: config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:     attrID,
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

func TestLibAttributesTenantIDInLine(t *testing.T) {
	ap := &AttributeProfile{
		Tenant:  "cgrates.org",
		ID:      "AttrPrf",
		Weights: make(utils.DynamicWeights, 1),
	}
	ap.Weights[0] = &utils.DynamicWeight{
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
		Weights: make(utils.DynamicWeights, 1),
	}
	ap.Weights[0] = &utils.DynamicWeight{
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
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      utils.MetaConstant,
			Value:     config.NewRSRParsersMustCompile("10", utils.InfieldSep),
			FilterIDs: []string{"fltr1"},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
		}},
		Weights: make(utils.DynamicWeights, 1),
	}
	exp.Weights[0] = &utils.DynamicWeight{
		Weight: 10,
	}
	if err := dp.Set([]string{}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField"}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}
	if err := dp.Set([]string{"NotAField", "1"}, "", false); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if err := dp.Set([]string{utils.Tenant}, "cgrates.org", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.ID}, "ID", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.FilterIDs}, "fltr1;*string:~*req.Account:1001", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Weights}, ";10", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Attributes, utils.Path}, "*req.Account", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Attributes, utils.Type}, utils.MetaConstant, false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Attributes, utils.Value}, "10", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Attributes, utils.FilterIDs}, "fltr1", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Attributes, utils.Blockers}, ";true", false); err != nil {
		t.Error(err)
	}
	if err := dp.Set([]string{utils.Attributes, "Wrong"}, true, false); err != utils.ErrWrongPath {
		t.Error(err)
	}

	if !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestAttributeProfileAsInterface(t *testing.T) {
	ap := AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   make(utils.DynamicWeights, 1),
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      utils.MetaConstant,
			Value:     config.NewRSRParsersMustCompile("10", utils.InfieldSep),
			FilterIDs: []string{"fltr1"},
		}},
	}
	ap.Weights[0] = &utils.DynamicWeight{
		Weight: 10,
	}
	if _, err := ap.FieldAsInterface(nil); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{"field", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{utils.Tenant}); err != nil {
		t.Fatal(err)
	} else if exp := "cgrates.org"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := utils.ID; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.FilterIDs[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Weights}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Weights; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Blockers}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Blockers; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0]; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if _, err := ap.FieldAsInterface([]string{utils.Attributes + "[4]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]", ""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if _, err := ap.FieldAsInterface([]string{utils.Attributes + "0]"}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]", utils.FilterIDs}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].FilterIDs; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]", utils.FilterIDs + "[0]"}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].FilterIDs[0]; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]", utils.Path}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].Path; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]", utils.Type}); err != nil {
		t.Fatal(err)
	} else if exp := ap.Attributes[0].Type; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, err := ap.FieldAsInterface([]string{utils.Attributes + "[0]", utils.Value}); err != nil {
		t.Fatal(err)
	} else if exp := "10"; !reflect.DeepEqual(exp, val) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := ap.FieldAsString([]string{""}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.FieldAsString([]string{utils.ID}); err != nil {
		t.Fatal(err)
	} else if exp := "ID"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := ap.String(), utils.ToJSON(ap); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}

	if _, err := ap.Attributes[0].FieldAsString([]string{}); err != utils.ErrNotFound {
		t.Fatal(err)
	}
	if val, err := ap.Attributes[0].FieldAsString([]string{utils.Value}); err != nil {
		t.Fatal(err)
	} else if exp := "10"; exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
	if val, exp := ap.Attributes[0].String(), utils.ToJSON(ap.Attributes[0]); exp != val {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(val))
	}
}

func TestAttributeProfileMerge(t *testing.T) {
	dp := &AttributeProfile{}
	exp := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   make(utils.DynamicWeights, 1),
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      utils.MetaConstant,
			Value:     config.NewRSRParsersMustCompile("10", utils.InfieldSep),
			FilterIDs: []string{"fltr1"},
		}},
	}
	exp.Weights[0] = &utils.DynamicWeight{
		Weight: 10,
	}
	dp.Merge(&AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ID",
		FilterIDs: []string{"fltr1", "*string:~*req.Account:1001"},
		Weights:   make(utils.DynamicWeights, 1),
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
		Attributes: []*Attribute{{
			Path:      "*req.Account",
			Type:      utils.MetaConstant,
			Value:     config.NewRSRParsersMustCompile("10", utils.InfieldSep),
			FilterIDs: []string{"fltr1"},
		}},
	})
	dp.Weights[0] = &utils.DynamicWeight{
		Weight: 10,
	}
	if !reflect.DeepEqual(exp, dp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(dp))
	}
}

func TestAttributeProfilCompileSubstitutes(t *testing.T) {

	ap := &AttributeProfile{
		Attributes: []*Attribute{
			{Value: config.RSRParsers{&config.RSRParser{
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
		Type:      utils.MetaConstant,
		Value:     config.NewRSRParsersMustCompile("10", utils.InfieldSep),
		FilterIDs: []string{"fltr1"},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: true,
			},
		},
	}

	if rcv, err := at.FieldAsInterface([]string{utils.Blockers}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcv, at.Blockers) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(at), utils.ToJSON(rcv))
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
		Weights: utils.DynamicWeights{
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
				Path:  utils.MetaReq + utils.NestingSep + "Account",
				Value: "a{*",
			},
		},
		Weights: utils.DynamicWeights{
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
	_, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err == nil || err.Error() != expErr {
		t.Errorf("Expecting error <%v>, Reveived error <%v>", expErr, err)
	}
}
