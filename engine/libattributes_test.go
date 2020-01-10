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
			&ExternalAttribute{
				FieldName: "Account",
				Value:     "1001",
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
				FieldName: "Account",
				Value:     config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
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
			&ExternalAttribute{
				FieldName: "Account",
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
	attrID := "*sum:Field2:10;~NumField;20"
	expAttrPrf1 := &AttributeProfile{
		Tenant:   config.CgrConfig().GeneralCfg().DefaultTenant,
		ID:       attrID,
		Contexts: []string{utils.META_ANY},
		Attributes: []*Attribute{&Attribute{
			FieldName: "Field2",
			Type:      utils.MetaSum,
			Value:     config.NewRSRParsersMustCompile("10;~NumField;20", true, utils.INFIELD_SEP),
		}},
	}
	attr, err := NewAttributeFromInline(config.CgrConfig().GeneralCfg().DefaultTenant, attrID)
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expAttrPrf1, attr) {
		t.Errorf("Expecting %+v, received: %+v", utils.ToJSON(expAttrPrf1), utils.ToJSON(attr))
	}
}
