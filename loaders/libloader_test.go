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

package loaders

import (
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

func TestTenantIDFromMap(t *testing.T) {
	exp := utils.NewTenantID("cgrates.org:ATTR1")
	r := TenantIDFromOrderedNavigableMap(
		newOrderNavMap(utils.MapStorage{
			utils.Tenant: exp.Tenant,
			utils.ID:     exp.ID,
		}))
	if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}

func TestRateIDsFromMap(t *testing.T) {
	expErrMsg := "cannot find RateIDs in map"
	if _, err := RateIDsFromOrderedNavigableMap(newOrderNavMap(utils.MapStorage{})); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %v, received: %v", expErrMsg, err)
	}
	exp := []string{"RT1", "RT2"}
	r, err := RateIDsFromOrderedNavigableMap(newOrderNavMap(utils.MapStorage{
		utils.RateIDs: "RT1;RT2",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}

func TestNewRecord(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	expErrMsg := "inline parse error for string: <*string>"
	if _, err := newRecord(context.Background(), utils.MapStorage{},
		[]*config.FCTemplate{
			{Filters: []string{"*exists:~*req.NoField:"}},
			{Filters: []string{"*string"}},
		},
		"cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil)); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %v", expErrMsg, err)
	}
	r, err := newRecord(context.Background(), utils.MapStorage{}, []*config.FCTemplate{}, "cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil))
	if err != nil {
		t.Fatal(err)
	}
	exp := utils.NewOrderedNavigableMap()
	if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}

func TestNewRecordWithCahe(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "Tenant", Value: config.NewRSRParsersMustCompile("~*req.0", ";")},
		{Type: utils.MetaVariable, Path: "ID", Value: config.NewRSRParsersMustCompile("~*req.1", ";")},
		{Type: utils.MetaComposed, Path: "*uch.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaVariable, Path: "Value", Value: config.NewRSRParsersMustCompile("~*uch.*tntID.Value", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	exp := utils.NewOrderedNavigableMap()
	exp.SetAsSlice(utils.NewFullPath(utils.Tenant), []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	exp.SetAsSlice(utils.NewFullPath(utils.ID), []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "Attr1"}}})
	exp.SetAsSlice(utils.NewFullPath(utils.Value), []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "0"}}})
	if r, err := newRecord(context.Background(), config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil),
		fc, "cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+v", exp, r)
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp.GetOrder()), utils.ToJSON(r.GetOrder()))
	}
}

func TestNewRecordWithTmp(t *testing.T) {
	cfg := config.NewDefaultCGRConfig()
	fs := engine.NewFilterS(cfg, nil, nil)
	fc := []*config.FCTemplate{
		{Type: utils.MetaVariable, Path: "Tenant", Value: config.NewRSRParsersMustCompile("~*req.0", ";")},
		{Type: utils.MetaVariable, Path: "ID", Value: config.NewRSRParsersMustCompile("~*req.1", ";")},
		{Type: utils.MetaComposed, Path: "*tmp.*tntID.Value", Value: config.NewRSRParsersMustCompile("0", ";")},
		{Type: utils.MetaVariable, Path: "Value", Value: config.NewRSRParsersMustCompile("~*tmp.*tntID.Value", ";")},
	}
	for _, f := range fc {
		f.ComputePath()
	}
	exp := utils.NewOrderedNavigableMap()
	exp.SetAsSlice(utils.NewFullPath(utils.Tenant), []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "cgrates.org"}}})
	exp.SetAsSlice(utils.NewFullPath(utils.ID), []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "Attr1"}}})
	exp.SetAsSlice(utils.NewFullPath(utils.Value), []*utils.DataNode{{Type: utils.NMDataType, Value: &utils.DataLeaf{Data: "0"}}})
	if r, err := newRecord(context.Background(), config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil),
		fc, "cgrates.org", fs, cfg, ltcache.NewCache(-1, 0, false, nil)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+q", exp, r)
	}
}
