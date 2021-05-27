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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestFilterIndexesCheckingDynamicPathToNotIndex(t *testing.T) {
	Cache.Clear(nil)
	cfg := config.NewDefaultCGRConfig()
	db := NewInternalDB(nil, nil, true)
	dm := NewDataManager(db, cfg.CacheCfg(), nil)
	//set 4 attr profiles with different filters to index them

	attrPrf1 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf1",
		FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-29T15:00:00Z", "*string:~*opts.*context:con1|con2|con3"},
		Attributes: []*Attribute{
			{
				FilterIDs: []string{"*string:~*req.Field1:Initial"},
				Path:      utils.MetaReq + utils.NestingSep + "Field1",
				Type:      utils.MetaVariable,
				Value:     config.NewRSRParsersMustCompile("Sub1", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  20,
	}

	attrPrf2 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf2",
		FilterIDs: []string{"*string:~*resources.RES_GRP1.Available:4"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "Password",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("admin", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  20,
	}

	attrPrf3 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf3",
		FilterIDs: []string{"*prefix:~*req.Destination:1007", "*string:~*req.Account:1001", "*string:~*opts.TotalCost:~*stats.STS_PRF1.*tcc"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "RequestType",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("*rated", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  20,
	}

	attrPrf4 := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "AttrPrf4",
		FilterIDs: []string{"*prefix:~*req.Destination:1007", "*prefix:~*accounts.RES_GRP1.Available:10"},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "TCC",
				Type:  utils.MetaVariable,
				Value: config.NewRSRParsersMustCompile("203", utils.InfieldSep),
			},
		},
		Blocker: true,
		Weight:  20,
	}

	if err := dm.SetAttributeProfile(context.Background(), attrPrf1, true); err != nil {
		t.Error(err)
	} else if err := dm.SetAttributeProfile(context.Background(), attrPrf2, true); err != nil {
		t.Error(err)
	} else if err := dm.SetAttributeProfile(context.Background(), attrPrf3, true); err != nil {
		t.Error(err)
	} else if err := dm.SetAttributeProfile(context.Background(), attrPrf4, true); err != nil {
		t.Error(err)
	}

	expIDx := map[string]utils.StringSet{
		"*prefix:*req.Destination:1007": {
			"AttrPrf3": {},
			"AttrPrf4": {},
		},
		"*string:*req.Account:1001": {
			"AttrPrf3": {},
			"AttrPrf1": {},
		},
		"*string:*opts.*context:con1": {
			"AttrPrf1": {},
		},
		"*string:*opts.*context:con2": {
			"AttrPrf1": {},
		},
		"*string:*opts.*context:con3": {
			"AttrPrf1": {},
		},
	}
	if fltrIDx, err := dm.GetIndexes(context.Background(), utils.CacheAttributeFilterIndexes,
		"cgrates.org", utils.EmptyString, true, true); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expIDx, fltrIDx) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIDx), utils.ToJSON(fltrIDx))
	}
}
