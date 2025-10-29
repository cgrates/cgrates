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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestConvertExternalToProfile(t *testing.T) {
	external := &ExternalAttributeProfile{
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
				Value: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
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
	external := &ExternalAttributeProfile{
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
	external := &ExternalAttributeProfile{
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

func TestNewAttributeFromInline(t *testing.T) {
	attrID := "*sum:*req.Field2:10;~*req.NumField;20"
	expAttrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       attrID,
		Contexts: []string{utils.META_ANY},
		Attributes: []*Attribute{{
			Path:  utils.MetaReq + utils.NestingSep + "Field2",
			Type:  utils.MetaSum,
			Value: config.NewRSRParsersMustCompile("10;~*req.NumField;20", true, utils.INFIELD_SEP),
		}},
	}
	attr, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}

func TestLibattributescompileSubstitutes(t *testing.T) {
	ap := AttributeProfile{
		Attributes: []*Attribute{
			{
				Value: config.RSRParsers{{Rules: "test)"}},
			},
		},
	}

	err := ap.compileSubstitutes()

	if err != nil {
		if err.Error() != "invalid RSRFilter start rule in string: <test)>" {
			t.Error(err)
		}
	}
}

func TestLibattributesSort(t *testing.T) {
	aps := AttributeProfiles{{Weight: 1.2}, {Weight: 1.5}}

	aps.Sort()

	if aps[0].Weight != 1.5 {
		t.Error("didn't sort")
	}
}

func TestLibattributesNewAttributeFromInline(t *testing.T) {
	tests := []struct {
		name string
		t    string
		in   string
		err  string
	}{
		{
			name: "split error check",
			t:    "",
			in:   "",
			err:  "inline parse error for string: <>",
		},
		{
			name: "NewRSRParsers error check",
			t:    "",
			in:   "test):test):test)",
			err:  "invalid RSRFilter start rule in string: <test)>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAttributeFromInline(tt.t, tt.in)

			if err != nil {
				if tt.err != err.Error() {
					t.Error(err)
				}
			}
		})
	}
}
