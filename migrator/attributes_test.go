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

func Testv1AttributeProfileAsAttributeProfile(t *testing.T) {
	var cloneExpTime time.Time
	expTime := time.Now().Add(time.Duration(20 * time.Minute))
	if err := utils.Clone(expTime, &cloneExpTime); err != nil {
		t.Error(err)
	}
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
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	if ap, err := v1Attribute.AsAttributeProfile(); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(attrPrf, ap) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(attrPrf), utils.ToJSON(ap))
	}
}
