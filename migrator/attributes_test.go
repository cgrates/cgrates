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
package migrator

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestV1AttributeProfileAsAttributeProfile(t *testing.T) {
	cloneExpTime := time.Now().Add(20 * time.Minute)
	mapSubstitutes := make(map[string]map[string]*v1Attribute)
	mapSubstitutes["FL1"] = make(map[string]*v1Attribute)
	mapSubstitutes["FL1"]["In1"] = &v1Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: "Al1",
		Append:     true,
	}
	v1Attribute := &v1AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	eOut := &v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v2Attribute{
			&v2Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
				Append:     true,
			},
		},
		Weight: 20,
	}
	if ap, err := v1Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, ap) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(ap))
	}
}

func TestV2AttributeProfileAsAttributeProfile(t *testing.T) {
	cloneExpTime := time.Now().Add(20 * time.Minute)
	v2Attribute := v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v2Attribute{
			&v2Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
				Append:     true,
			},
		},
		Weight: 20,
	}
	eOut := &v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v3Attribute{
			&v3Attribute{
				FilterIDs:  []string{"*string:FL1:In1"},
				FieldName:  "FL1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if ap, err := v2Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, ap) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(ap))
	}
}

func TestV2AttributeProfileAsAttributeProfile2(t *testing.T) {
	cloneExpTime := time.Now().Add(20 * time.Minute)
	v2Attribute := v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v2Attribute{
			&v2Attribute{
				FieldName:  "FL1",
				Initial:    nil,
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
				Append:     true,
			},
		},
		Weight: 20,
	}
	eOut := &v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v3Attribute{
			&v3Attribute{
				FilterIDs:  []string{"*string:FL1:"},
				FieldName:  "FL1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if ap, err := v2Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, ap) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(ap))
	}
}

func TestV3AttributeProfileAsAttributeProfile(t *testing.T) {
	cloneExpTime := time.Now().Add(20 * time.Minute)
	v3Attribute := v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v3Attribute{
			&v3Attribute{
				FilterIDs:  []string{"*string:FL1:In1"},
				FieldName:  "FL1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	attrPrf := &v4AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v4Attribute{
			&v4Attribute{
				FilterIDs: []string{"*string:FL1:In1"},
				FieldName: "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if ap, err := v3Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attrPrf, ap) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(ap))
	}
}

func TestV4AttributeProfileAsAttributeProfile(t *testing.T) {
	cloneExpTime := time.Now().Add(20 * time.Minute)
	v4Attribute := v4AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*v4Attribute{
			&v4Attribute{
				FieldName: "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~Category:s/(.*)/${1}_UK_Mobile_Vodafone_GBRVF/", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	attrPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     cloneExpTime,
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("~*req.Category:s/(.*)/${1}_UK_Mobile_Vodafone_GBRVF/", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	if ap, err := v4Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attrPrf, ap) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(ap))
	}
}

func TestAsAttributeProfileV2(t *testing.T) {
	// contruct the v1 attribute with all fields filled up
	mapSubstitutes := make(map[string]map[string]*v1Attribute)
	mapSubstitutes["FL1"] = make(map[string]*v1Attribute)
	mapSubstitutes["FL1"]["In1"] = &v1Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: "Al1",
		Append:     true,
	}
	v1Attribute := &v1AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	sbstPrsr, err := config.NewRSRParsers("Al1", config.CgrConfig().GeneralCfg().RSRSep)
	if err != nil {
		t.Error("Error converting Substitute from string to RSRParser: ", err)
	}
	eOut := &v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v2Attribute{
			&v2Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: sbstPrsr,
				Append:     true,
			},
		},
		Weight: 20,
	}

	if v2Attribute, err := v1Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, v2Attribute) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(v2Attribute))
	}

}

func TestAsAttributeProfileV3(t *testing.T) {
	v2Attribute := v2AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v2Attribute{
			&v2Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
				Append:     true,
			},
		},
		Weight: 20,
	}
	eOut := &v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v3Attribute{
			&v3Attribute{
				FilterIDs:  []string{"*string:FL1:In1"}, //here
				FieldName:  "FL1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			}},
		Weight: 20,
	}
	if v3Attribute, err := v2Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, v3Attribute) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(v3Attribute))
	}
}

func TestAsAttributeProfileV4(t *testing.T) {
	v3Attribute := v3AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v3Attribute{
			&v3Attribute{
				FilterIDs:  []string{"*string:FL1:In1"},
				FieldName:  "FL1",
				Substitute: config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	eOut := &v4AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v4Attribute{
			&v4Attribute{
				FilterIDs: []string{"*string:FL1:In1"},
				FieldName: "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
			}},

		Blocker: false,
		Weight:  20,
	}

	if v4Attribute, err := v3Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, v4Attribute) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(v4Attribute))
	}
}

func TestAsAttributeProfileV5(t *testing.T) {
	v4Attribute := v4AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v4Attribute{
			&v4Attribute{
				FilterIDs: []string{"*string:FL1:In1"},
				FieldName: "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}

	eOut := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FilterIDs: []string{"*string:FL1:In1"},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("~*req.Al1", utils.InfieldSep),
			},
		},
		Weight: 20,
	}

	if v5Attribute, err := v4Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, v5Attribute) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(v5Attribute))
	}
}

func TestAsAttributeProfileV1To4(t *testing.T) {
	// contruct the v1 attribute with all fields filled up
	mapSubstitutes := make(map[string]map[string]*v1Attribute)
	mapSubstitutes["FL1"] = make(map[string]*v1Attribute)
	mapSubstitutes["FL1"]["In1"] = &v1Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: "Al1",
		Append:     true,
	}
	v1Attribute := &v1AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: mapSubstitutes,
		Weight:     20,
	}
	sbstPrsr, err := config.NewRSRParsers("Al1", config.CgrConfig().GeneralCfg().RSRSep)
	if err != nil {
		t.Error("Error converting Substitute from string to RSRParser: ", err)
	}
	eOut := &v4AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "attributeprofile1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"filter1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
		},
		Attributes: []*v4Attribute{
			&v4Attribute{
				FieldName: "FL1",
				Type:      utils.MetaVariable,
				Value:     sbstPrsr,
				FilterIDs: []string{"*string:FL1:In1"},
			}},
		Weight: 20,
	}
	if rcv, err := v1Attribute.AsAttributeProfileV1To4(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eOut, rcv) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(eOut), utils.ToJSON(rcv))
	}

}
