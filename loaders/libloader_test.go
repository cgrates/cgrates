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
	"encoding/json"
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
	if err := newRecord(utils.MapStorage{}, nil, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), []*config.FCTemplate{
			{Filters: []string{"*exists:~*req.NoField:"}},
			{Filters: []string{"*string"}},
		}, fs, 0, "", utils.InfieldSep); err == nil || err.Error() != expErrMsg {
		t.Errorf("Expeceted: %q, received: %v", expErrMsg, err)
	}
	pt := profileTest{}
	if err := newRecord(utils.MapStorage{}, pt, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), []*config.FCTemplate{}, fs, 0, "", utils.InfieldSep); err != nil {
		t.Error(err)
	}
	if len(pt) != 0 {
		t.Fatal("Expected empty map")
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
	exp := profileTest{
		utils.Tenant: "cgrates.org",
		utils.ID:     "Attr1",
		utils.Value:  "0",
	}
	r := profileTest{}
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+v", exp, r)
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
	exp := profileTest{
		utils.Tenant: "cgrates.org",
		utils.ID:     "Attr1",
		utils.Value:  "0",
	}
	r := profileTest{}
	if err := newRecord(config.NewSliceDP([]string{"cgrates.org", "Attr1"}, nil), r, "cgrates.org", cfg, ltcache.NewCache(-1, 0, false, nil)).
		SetFields(context.Background(), fc, fs, 0, "", utils.InfieldSep); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(r, exp) {
		t.Errorf("Expected %+v, received %+v", exp, r)
	}
}

type profileTest utils.MapStorage

func (p profileTest) Set(path []string, val interface{}, _ bool, _ string) error {
	return utils.MapStorage(p).Set(path, val)
}
func (p profileTest) Merge(v2 interface{}) {
	var vi map[string]interface{}
	json.Unmarshal([]byte(utils.ToJSON(v2)), &vi)
	for k, v := range vi {
		(map[string]interface{}(p))[k] = v
	}
}
func (p profileTest) TenantID() string {
	return utils.ConcatenatedKey(utils.IfaceAsString(map[string]interface{}(p)[utils.Tenant]), utils.IfaceAsString(map[string]interface{}(p)[utils.ID]))
}
